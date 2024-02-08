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

//go:build !(unix || aix || android || darwin || dragonfly || freebsd || hurd || illumos || ios || linux || netbsd || openbsd || solaris || windows)

package at

import (
	"os"
	"time"
)

func (d Dir) utimes(path string, atime, mtime time.Time, flag int) error {
	if flag == SymlinkNoFollow {
		return &Error{Op: "utimes", Err: ErrUnsupportedFlag}
	}
	if mtime == UtimeOmit {
		return &Error{Op: "utimes", Err: ErrMtimeOmitted}
	}
	if atime == UtimeOmit {
		atime = time.Now()
	}
	return os.Chtimes(d.join(path), atime, mtime)
}
