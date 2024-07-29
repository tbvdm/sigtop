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
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

const dpapiKeyPrefix = "DPAPI"

func Decrypt(ciphertext []byte) ([]byte, error) {
	return nil, fmt.Errorf("not supported")
}

func DecryptWithLocalState(ciphertext []byte, localStateFile string) ([]byte, error) {
	key, err := encryptionKey(localStateFile)
	if err != nil {
		return nil, err
	}
	return decryptWithWindowsPassword(ciphertext, key)
}

func encryptionKey(localStateFile string) ([]byte, error) {
	data, err := os.ReadFile(localStateFile)
	if err != nil {
		return nil, err
	}

	var localState struct {
		OSCrypt struct {
			EncryptedKey *string `json:"encrypted_key"`
		} `json:"os_crypt"`
	}
	if err := json.Unmarshal(data, &localState); err != nil {
		return nil, fmt.Errorf("cannot parse %s: %w", localStateFile, err)
	}

	if localState.OSCrypt.EncryptedKey == nil {
		return nil, fmt.Errorf("encryption key not found")
	}

	key, err := base64.StdEncoding.DecodeString(*localState.OSCrypt.EncryptedKey)
	if err != nil {
		return nil, fmt.Errorf("cannot decode encryption key: %w", err)
	}

	if !bytes.HasPrefix(key, []byte(dpapiKeyPrefix)) {
		return nil, fmt.Errorf("invalid encryption key format")
	}
	key = bytes.TrimPrefix(key, []byte(dpapiKeyPrefix))

	return decryptWithDPAPI(key)
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

	plaintext := make([]byte, out.Size)
	copy(plaintext, unsafe.Slice(out.Data, out.Size))
	windows.LocalFree(windows.Handle(unsafe.Pointer(out.Data)))

	return plaintext, nil
}
