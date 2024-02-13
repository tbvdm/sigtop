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

package main

import (
	"os"
	"unicode"

	"golang.org/x/text/unicode/rangetable"
)

var unicode9 *unicode.RangeTable

func init() {
	unicode9 = rangetable.Assigned("9.0.0")
	if unicode9 == nil {
		panic("cannot get range table")
	}
}

func sanitiseFilename(name string) string {
	if name == "" || name == "." || name == ".." {
		return name + "_"
	}

	runes := []rune(name)
	for i, r := range runes {
		// Note that APFS allows only Unicode 9.0 characters
		if r == os.PathSeparator || unicode.IsControl(r) || !unicode.Is(unicode9, r) {
			runes[i] = '_'
		}
	}

	return string(runes)
}
