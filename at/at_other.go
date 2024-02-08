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

//go:build !(unix || aix || android || darwin || dragonfly || freebsd || hurd || illumos || ios || linux || netbsd || openbsd || solaris)

package at

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

const (
	SymlinkNoFollow = 1 << iota
	SymlinkFollow
	RemoveDir
)

type Dir struct {
	path string
	file *os.File
}

var (
	InvalidDir = Dir{}
	CurrentDir = Dir{path: "."}
)

func open(path string) (Dir, error) {
	f, err := os.Open(path)
	if err != nil {
		return InvalidDir, err
	}
	fi, err := f.Stat()
	if err != nil {
		f.Close()
		return InvalidDir, err
	}
	if !fi.IsDir() {
		f.Close()
		return InvalidDir, &os.PathError{Op: "open", Path: path, Err: ErrNotDir}
	}
	return Dir{path: path, file: f}, nil
}

func (d Dir) close() error {
	return d.file.Close()
}

func (d Dir) openDir(path string) (Dir, error) {
	return open(d.join(path))
}

func (d Dir) openFile(path string, flag int, perm fs.FileMode) (*os.File, error) {
	return os.OpenFile(d.join(path), flag, perm)
}

func (d Dir) mkdir(path string, perm fs.FileMode) error {
	return os.Mkdir(d.join(path), perm)
}

func (d Dir) chdir() error {
	if d == CurrentDir {
		return nil
	}
	// os.File.Chdir is not supported on Windows
	if runtime.GOOS == "windows" {
		return os.Chdir(d.path)
	}
	return d.file.Chdir()
}

func (d Dir) link(srcDir Dir, src, dst string, flag int) error {
	if flag == SymlinkFollow {
		return &Error{Op: "link", Err: ErrUnsupportedFlag}
	}
	return os.Link(srcDir.join(src), d.join(dst))
}

func (d Dir) symlink(src, dst string) error {
	return os.Symlink(src, d.join(dst))
}

func (d Dir) unlink(path string, flag int) error {
	return os.Remove(d.join(path))
}

func (d Dir) stat(path string, flag int) (fs.FileInfo, error) {
	if flag == SymlinkNoFollow {
		return os.Lstat(d.join(path))
	}
	return os.Stat(d.join(path))
}

func futimes(f *os.File, atime, mtime time.Time) error {
	return os.Chtimes(f.Name(), atime, mtime)
}

func (d Dir) join(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(d.path, path)
}
