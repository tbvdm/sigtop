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

package at

import (
	"errors"
	"io/fs"
	"os"
	"time"
)

var (
	UtimeOmit time.Time

	ErrNotDir          = errors.New("not a directory")
	ErrInvalidFlag     = errors.New("invalid flag")
	ErrUnsupportedFlag = errors.New("unsupported flag")
	ErrMtimeOmitted    = errors.New("omitted modification time")
)

type Error struct {
	Op  string
	Err error
}

func (e *Error) Error() string {
	return e.Op + ": " + e.Err.Error()
}

func (e *Error) Unwrap() error {
	return e.Err
}

func Open(path string) (Dir, error) {
	return open(path)
}

func (d Dir) Close() error {
	return d.close()
}

func (d Dir) OpenDir(path string) (Dir, error) {
	return d.openDir(path)
}

func (d Dir) OpenFile(path string, flag int, perm fs.FileMode) (*os.File, error) {
	return d.openFile(path, flag, perm)
}

func (d Dir) Mkdir(path string, perm fs.FileMode) error {
	return d.mkdir(path, perm)
}

func (d Dir) Chdir() error {
	return d.chdir()
}

func (d Dir) Link(srcDir Dir, src, dst string, flag int) error {
	if flag != 0 && flag != SymlinkFollow {
		return &Error{Op: "link", Err: ErrInvalidFlag}
	}
	return d.link(srcDir, src, dst, flag)
}

func (d Dir) Symlink(src, dst string) error {
	return d.symlink(src, dst)
}

func (d Dir) Unlink(path string, flag int) error {
	if flag != 0 && flag != RemoveDir {
		return &Error{Op: "unlink", Err: ErrInvalidFlag}
	}
	return d.unlink(path, flag)
}

func (d Dir) Stat(path string, flag int) (fs.FileInfo, error) {
	if flag != 0 && flag != SymlinkNoFollow {
		return nil, &Error{Op: "stat", Err: ErrInvalidFlag}
	}
	return d.stat(path, flag)
}

func (d Dir) Utimes(path string, atime, mtime time.Time, flag int) error {
	if flag != 0 && flag != SymlinkNoFollow {
		return &Error{Op: "utimes", Err: ErrInvalidFlag}
	}
	return d.utimes(path, atime, mtime, flag)
}

func Futimes(f *os.File, atime, mtime time.Time) error {
	return futimes(f, atime, mtime)
}
