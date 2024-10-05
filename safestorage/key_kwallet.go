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

// Don't build on OpenBSD because D-Bus requires extra pledge(2) promises
//
//go:build !(darwin || openbsd || windows)

package safestorage

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/godbus/dbus/v5"
)

func (a *App) rawEncryptionKeyFromKwallet(service, objectPath string) ([]byte, error) {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return nil, fmt.Errorf("cannot connect to D-Bus session bus: %w", err)
	}
	defer conn.Close()

	obj := conn.Object(service, dbus.ObjectPath(objectPath))
	call := func(method string, args ...interface{}) *dbus.Call {
		return obj.Call(kwalletInterface+"."+method, 0, args...)
	}

	// Get wallet name
	var wallet string
	err = call("networkWallet").Store(&wallet)
	if err != nil {
		return nil, fmt.Errorf("cannot get wallet name: %w", err)
	}

	// Create application ID
	appid, err := os.Executable()
	if err != nil {
		appid = os.Args[0]
		if len(appid) == 0 {
			appid = "unknown"
		}
	}
	appid = filepath.Base(appid)

	// Open wallet
	wid := int64(0) // We have no window ID
	var handle int
	err = call("open", wallet, wid, appid).Store(&handle)
	if err != nil {
		return nil, fmt.Errorf("cannot open wallet %q: %w", wallet, err)
	}
	if handle == kwalletInvalidHandle {
		return nil, fmt.Errorf("cannot open wallet %q: invalid handle", wallet)
	}
	defer call("close", handle, false, appid)

	// Check if entry exists
	var hasEntry bool
	err = call("hasEntry", handle, kwalletFolder, kwalletEntry, appid).Store(&hasEntry)
	if err != nil {
		return nil, fmt.Errorf("cannot check if wallet entry exists: %w", err)
	}
	if !hasEntry {
		return nil, fmt.Errorf("cannot find encryption key")
	}

	// Check entry type
	var entryType int
	err = call("entryType", handle, kwalletFolder, kwalletEntry, appid).Store(&entryType)
	if err != nil {
		return nil, fmt.Errorf("cannot get wallet entry type: %w", err)
	}
	if entryType != kwalletEntryTypePassword {
		return nil, fmt.Errorf("unexpected wallet entry type: %d", entryType)
	}

	// Get encryption key
	var key string
	err = call("readPassword", handle, kwalletFolder, kwalletEntry, appid).Store(&key)
	if err != nil {
		return nil, fmt.Errorf("cannot get encryption key from wallet: %w", err)
	}

	return []byte(key), nil
}
