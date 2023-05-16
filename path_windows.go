// Copyright (c) 2023 Tim van der Molen <tim@kariliq.nl>
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

package main

import (
	"strings"
	"unicode"
)

// sanitiseFilename sanitises a filename for Windows. The sanitation is based
// on https://learn.microsoft.com/en-us/windows/win32/fileio/naming-a-file.
func sanitiseFilename(name string) string {
	if name == "" {
		return "_"
	}

	// If X is a reserved name, then transform "X" and "X.ext" into "X_"
	// and "X_.ext", respectively
	rnames := [...]string{
		"AUX", "CON", "NUL", "PRN",
		"COM0", "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
		"LPT0", "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
	}
	i := strings.IndexByte(name, '.')
	if i <= 0 {
		i = len(name)
	}
	base, ext := name[:i], name[i:]
	for _, rname := range rnames {
		if strings.EqualFold(base, rname) {
			name = base + "_" + ext
			break
		}
	}

	// Handle filenames ending with a space or a dot (this also takes care
	// of "." and "..")
	if name[len(name)-1] == ' ' || name[len(name)-1] == '.' {
		name += "_"
	}

	// Replace reserved characters
	rchars := `"*/:<>?\|`
	runes := []rune(name)
	for i, r := range runes {
		if strings.ContainsRune(rchars, r) || unicode.IsControl(r) {
			runes[i] = '_'
		}
	}

	return string(runes)
}
