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
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"runtime"

	"golang.org/x/crypto/pbkdf2"
)

type backend int

const (
	backendNone backend = iota
	backendGnome
	backendKwallet4
	backendKwallet5
	backendKwallet6
)

type App struct {
	name    string
	dir     string
	backend backend
	rawKey  RawEncryptionKey
	key     []byte
	keySet  bool
}

type RawEncryptionKey struct {
	Key []byte
	OS  string
}

func NewApp(name, dir string) App {
	return App{name: name, dir: dir}
}

func (a *App) SetBackend(backend string) error {
	switch backend {
	case "gnome_libsecret":
		a.backend = backendGnome
	case "kwallet":
		a.backend = backendKwallet4
	case "kwallet5":
		a.backend = backendKwallet5
	case "kwallet6":
		a.backend = backendKwallet6
	default:
		return fmt.Errorf("invalid or unsupported safeStorage backend: %q", backend)
	}
	return nil
}

func (a *App) SetEncryptionKey(rawKey RawEncryptionKey) error {
	if rawKey.OS == "" {
		switch runtime.GOOS {
		case "darwin":
			rawKey.OS = "macos"
		case "windows":
			rawKey.OS = "windows"
		default:
			rawKey.OS = "linux"
		}
	}

	var key []byte
	switch rawKey.OS {
	case "linux":
		key = deriveEncryptionKey(rawKey.Key, linuxIterations)
	case "macos":
		key = deriveEncryptionKey(rawKey.Key, macosIterations)
	case "windows":
		var err error
		if key, err = base64.StdEncoding.DecodeString(string(rawKey.Key)); err != nil {
			return fmt.Errorf("invalid encryption key: %w", err)
		}
		if len(key) != windowsKeySize {
			return fmt.Errorf("invalid encryption key length")
		}
	default:
		return fmt.Errorf("invalid system: %s", rawKey.OS)
	}

	a.rawKey = rawKey
	a.key = key
	a.keySet = true
	return nil
}

func deriveEncryptionKey(rawKey []byte, iters int) []byte {
	return pbkdf2.Key(rawKey, []byte(salt), iters, keySize, sha1.New)
}

func (a *App) EncryptionKey() (*RawEncryptionKey, error) {
	if !a.keySet {
		if err := a.setEncryptionKeyFromSystem(); err != nil {
			return nil, err
		}
	}
	rawKey := a.rawKey
	return &rawKey, nil
}
