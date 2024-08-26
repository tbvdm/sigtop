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
	"crypto/aes"
	"crypto/cipher"
	"fmt"
)

func (a *App) Decrypt(ciphertext []byte) ([]byte, error) {
	if !a.keySet {
		if err := a.setEncryptionKeyFromSystem(); err != nil {
			return nil, err
		}
	}
	switch a.rawKey.OS {
	case "linux":
		return a.decrypt(ciphertext, linuxCiphertextPrefix)
	case "macos":
		return a.decrypt(ciphertext, macosCiphertextPrefix)
	case "windows":
		return a.decryptWindows(ciphertext)
	default:
		// Should not happen
		return nil, fmt.Errorf("invalid operating system")
	}
}

func (a *App) decrypt(ciphertext []byte, prefix string) ([]byte, error) {
	if !hasPrefix(ciphertext, prefix) {
		return nil, fmt.Errorf("unsupported ciphertext format")
	}
	ciphertext = trimPrefix(ciphertext, prefix)

	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("invalid ciphertext length")
	}

	c, err := aes.NewCipher(a.key)
	if err != nil {
		return nil, err
	}
	iv := bytes.Repeat([]byte(" "), aes.BlockSize)
	cbc := cipher.NewCBCDecrypter(c, iv)

	plaintext := make([]byte, len(ciphertext))
	cbc.CryptBlocks(plaintext, ciphertext)
	return unpad(plaintext, aes.BlockSize)
}

func (a *App) decryptWindows(ciphertext []byte) ([]byte, error) {
	if !hasPrefix(ciphertext, windowsCiphertextPrefix) {
		return nil, fmt.Errorf("unsupported ciphertext format")
	}
	ciphertext = trimPrefix(ciphertext, windowsCiphertextPrefix)

	if len(ciphertext) < windowsNonceSize {
		return nil, fmt.Errorf("invalid ciphertext length")
	}
	nonce := ciphertext[:windowsNonceSize]
	ciphertext = ciphertext[windowsNonceSize:]

	c, err := aes.NewCipher(a.key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCMWithNonceSize(c, windowsNonceSize)
	if err != nil {
		return nil, err
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

func hasPrefix(b []byte, prefix string) bool {
	return bytes.HasPrefix(b, []byte(prefix))
}

func trimPrefix(b []byte, prefix string) []byte {
	return bytes.TrimPrefix(b, []byte(prefix))
}

func unpad(data []byte, blocksize int) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}

	n := int(data[len(data)-1])
	if n == 0 || n > blocksize || n > len(data) {
		return nil, fmt.Errorf("invalid padding size")
	}

	for i := len(data) - n; i < len(data)-1; i++ {
		if data[i] != data[len(data)-1] {
			return nil, fmt.Errorf("invalid byte in padding string")
		}
	}

	return data[:len(data)-n], nil
}
