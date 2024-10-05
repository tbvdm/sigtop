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

//go:build !(darwin || windows)

package safestorage

import "fmt"

func (a *App) setEncryptionKeyFromSystem() error {
	key, err := a.rawEncryptionKeyFromBackend()
	if err != nil {
		return err
	}
	a.rawKey = RawEncryptionKey{
		Key: key,
		OS:  "linux",
	}
	a.key = deriveEncryptionKey(a.rawKey.Key, linuxIterations)
	return nil
}

func (a *App) rawEncryptionKeyFromBackend() ([]byte, error) {
	switch a.backend {
	case backendGnome:
		return a.rawEncryptionKeyFromLibsecret()
	case backendKwallet4:
		return a.rawEncryptionKeyFromKwallet(kwallet4Service, kwallet4Path)
	case backendKwallet5:
		return a.rawEncryptionKeyFromKwallet(kwallet5Service, kwallet5Path)
	case backendKwallet6:
		return a.rawEncryptionKeyFromKwallet(kwallet6Service, kwallet6Path)
	case backendNone:
		return nil, fmt.Errorf("safeStorage backend not set")
	default:
		// Should not happen
		return nil, fmt.Errorf("invalid safeStorage backend")
	}
}
