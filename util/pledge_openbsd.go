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

package util

import (
	"fmt"

	"golang.org/x/sys/unix"
)

func Pledge(promises, execPromises string) error {
	if err := unix.Pledge(promises, execPromises); err != nil {
		return fmt.Errorf("pledge: %w", err)
	}
	return nil
}

func Unveil(path, permissions string) error {
	if path == "" && permissions == "" {
		if err := unix.UnveilBlock(); err != nil {
			return fmt.Errorf("unveil: %w", err)
		}
	} else {
		if err := unix.Unveil(path, permissions); err != nil {
			return fmt.Errorf("unveil: %s: %w", path, err)
		}
	}
	return nil
}
