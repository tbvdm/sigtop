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
	"crypto/sha1"
	"errors"
	"runtime"

	"golang.org/x/crypto/pbkdf2"
)

const (
	keySize = 16 // AES-128
	salt    = "saltysalt"

	macosPrefix     = "v10"
	macosIterations = 1003

	linuxPrefixV10  = "v10"
	linuxPrefixV11  = "v11"
	linuxIterations = 1
)

func DecryptWithPassword(ciphertext, password []byte) ([]byte, error) {
	switch runtime.GOOS {
	case "darwin":
		return decryptWithMacosPassword(ciphertext, password)
	case "linux", "openbsd":
		return decryptWithLinuxPassword(ciphertext, password)
	default:
		return nil, errors.New("not yet supported")
	}
}

func decryptWithMacosPassword(ciphertext, password []byte) ([]byte, error) {
	if !bytes.HasPrefix(ciphertext, []byte(macosPrefix)) {
		return ciphertext, nil
	}
	ciphertext = bytes.TrimPrefix(ciphertext, []byte(macosPrefix))
	return decryptWithPassword(ciphertext, password, macosIterations)
}

func decryptWithLinuxPassword(ciphertext, password []byte) ([]byte, error) {
	if bytes.HasPrefix(ciphertext, []byte(linuxPrefixV10)) {
		return nil, errors.New("unsupported encryption version prefix")
	}
	if !bytes.HasPrefix(ciphertext, []byte(linuxPrefixV11)) {
		return ciphertext, nil
	}
	ciphertext = bytes.TrimPrefix(ciphertext, []byte(linuxPrefixV11))
	return decryptWithPassword(ciphertext, password, linuxIterations)
}

func decryptWithPassword(ciphertext, password []byte, iters int) ([]byte, error) {
	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, errors.New("invalid ciphertext length")
	}

	key := pbkdf2.Key(password, []byte(salt), iters, keySize, sha1.New)
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	iv := bytes.Repeat([]byte(" "), aes.BlockSize)
	cbc := cipher.NewCBCDecrypter(c, iv)

	plaintext := make([]byte, len(ciphertext))
	cbc.CryptBlocks(plaintext, ciphertext)
	return unpad(plaintext, aes.BlockSize)
}

func unpad(data []byte, blocksize int) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}

	n := int(data[len(data)-1])
	if n == 0 || n > blocksize || n > len(data) {
		return nil, errors.New("invalid padding size")
	}

	for i := len(data) - n; i < len(data)-1; i++ {
		if data[i] != data[len(data)-1] {
			return nil, errors.New("invalid byte in padding string")
		}
	}

	return data[:len(data)-n], nil
}
