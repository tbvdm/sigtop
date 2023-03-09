// Copyright (c) 2021, 2023 Tim van der Molen <tim@kariliq.nl>
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

package signal

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

func DesktopDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	// Try the default directory
	defDir := filepath.Join(configDir, "Signal")
	if ok, err := tryDesktopDir(defDir); ok {
		return defDir, err
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	// Try the Flatpak directory
	flatpakDir := filepath.Join(homeDir, ".var", "app", "org.signal.Signal", "config", "Signal")
	if ok, err := tryDesktopDir(flatpakDir); ok {
		return flatpakDir, err
	}

	// Try the Snap directory
	snapDir := filepath.Join(homeDir, "snap", "signal-desktop", "current", ".config", "Signal")
	if ok, err := tryDesktopDir(snapDir); ok {
		return snapDir, err
	}

	// Fall back to the default directory
	return defDir, nil
}

func tryDesktopDir(dir string) (bool, error) {
	_, err := os.Lstat(dir)
	if err != nil {
		return !errors.Is(err, fs.ErrNotExist), err
	}
	return true, nil
}
