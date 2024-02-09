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

package at

import (
	"os"
	"time"

	"golang.org/x/sys/windows"
)

func (d Dir) utimes(path string, atime, mtime time.Time, flag int) error {
	upath, err := windows.UTF16PtrFromString(d.join(path))
	if err != nil {
		return &os.PathError{Op: "utimes", Path: path, Err: err}
	}

	var fileFlags uint32 = windows.FILE_FLAG_BACKUP_SEMANTICS
	if flag == SymlinkNoFollow {
		fileFlags |= windows.FILE_FLAG_OPEN_REPARSE_POINT
	}

	h, err := windows.CreateFile(upath, windows.FILE_WRITE_ATTRIBUTES, windows.FILE_SHARE_WRITE, nil, windows.OPEN_EXISTING, fileFlags, 0)
	if err != nil {
		return &os.PathError{Op: "utimes", Path: path, Err: err}
	}

	aft := timeToFiletime(atime)
	mft := timeToFiletime(mtime)
	if err := windows.SetFileTime(h, nil, &aft, &mft); err != nil {
		windows.CloseHandle(h)
		return &os.PathError{Op: "utimes", Path: path, Err: err}
	}

	if err := windows.CloseHandle(h); err != nil {
		return &os.PathError{Op: "utimes", Path: path, Err: err}
	}

	return nil
}

func timeToFiletime(t time.Time) windows.Filetime {
	if t == UtimeOmit {
		return windows.Filetime{LowDateTime: 0, HighDateTime: 0}
	}
	return windows.NsecToFiletime(t.UnixNano())
}
