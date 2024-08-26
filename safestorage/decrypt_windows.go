// Copyright (c) 2024 Tim van der Molen <tim@kariliq.nl>
//
// Permission to use, copy, modify, and distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
// WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
// ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
// WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
// ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
// OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.

package safestorage

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"unsafe"

	"golang.org/x/sys/windows"
)

func (a *App) setEncryptionKeyFromSystem() error {
	file := filepath.Join(a.dir, localStateFile)
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	var localState struct {
		OSCrypt struct {
			EncryptedKey *string `json:"encrypted_key"`
		} `json:"os_crypt"`
	}
	if err := json.Unmarshal(data, &localState); err != nil {
		return fmt.Errorf("cannot parse %s: %w", file, err)
	}

	if localState.OSCrypt.EncryptedKey == nil {
		return fmt.Errorf("encryption key not found")
	}

	key, err := base64.StdEncoding.DecodeString(*localState.OSCrypt.EncryptedKey)
	if err != nil {
		return fmt.Errorf("cannot decode encryption key: %w", err)
	}

	if !hasPrefix(key, windowsDPAPIKeyPrefix) {
		return fmt.Errorf("unsupported encryption key format")
	}
	key = trimPrefix(key, windowsDPAPIKeyPrefix)

	key, err = decryptWithDPAPI(key)
	if err != nil {
		return err
	}

	a.rawKey = RawEncryptionKey{
		Key: []byte(base64.StdEncoding.EncodeToString(key)),
		OS:  "windows",
	}
	a.key = key

	return nil
}

func decryptWithDPAPI(ciphertext []byte) ([]byte, error) {
	in := windows.DataBlob{
		Data: &ciphertext[0],
		Size: uint32(len(ciphertext)),
	}
	var out windows.DataBlob
	if err := windows.CryptUnprotectData(&in, nil, nil, 0, nil, 0, &out); err != nil {
		return nil, err
	}
	defer windows.LocalFree(windows.Handle(unsafe.Pointer(out.Data)))

	plaintext := make([]byte, out.Size)
	copy(plaintext, unsafe.Slice(out.Data, out.Size))

	return plaintext, nil
}
