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

//go:build unix || aix || android || darwin || dragonfly || freebsd || hurd || illumos || ios || linux || netbsd || openbsd || solaris

package at

import (
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/sys/unix"
)

const (
	SymlinkNoFollow = unix.AT_SYMLINK_NOFOLLOW
	SymlinkFollow   = unix.AT_SYMLINK_FOLLOW
	RemoveDir       = unix.AT_REMOVEDIR
)

type Dir int

const (
	InvalidDir = Dir(-1)
	CurrentDir = Dir(unix.AT_FDCWD)
)

type fileInfo struct {
	name    string
	mode    fs.FileMode
	modTime time.Time
	stat    unix.Stat_t
}

func (fi fileInfo) Name() string {
	return fi.name
}

func (fi fileInfo) Size() int64 {
	return fi.stat.Size
}

func (fi fileInfo) Mode() fs.FileMode {
	return fi.mode
}

func (fi fileInfo) ModTime() time.Time {
	return fi.modTime
}

func (fi fileInfo) IsDir() bool {
	return fi.Mode().IsDir()
}

func (fi fileInfo) Sys() any {
	return fi.stat
}

func open(path string) (Dir, error) {
	fd, err := unix.Open(path, unix.O_RDONLY|unix.O_DIRECTORY, 0)
	if err != nil {
		return InvalidDir, &os.PathError{Op: "open", Path: path, Err: err}
	}
	return Dir(fd), nil
}

func (d Dir) close() error {
	if err := unix.Close(int(d)); err != nil {
		return &Error{Op: "close", Err: err}
	}
	return nil
}

func (d Dir) openDir(path string) (Dir, error) {
	fd, err := unix.Openat(int(d), path, unix.O_RDONLY|unix.O_DIRECTORY, 0)
	if err != nil {
		return InvalidDir, &os.PathError{Op: "open", Path: path, Err: err}
	}
	return Dir(fd), nil
}

func (d Dir) openFile(path string, flag int, perm fs.FileMode) (*os.File, error) {
	fd, err := unix.Openat(int(d), path, flag, uint32(perm&fs.ModePerm))
	if err != nil {
		return nil, &os.PathError{Op: "open", Path: path, Err: err}
	}
	f := os.NewFile(uintptr(fd), path)
	if f == nil {
		panic("os.NewFile returned nil")
	}
	return f, nil
}

func (d Dir) mkdir(path string, perm fs.FileMode) error {
	if err := unix.Mkdirat(int(d), path, uint32(perm&fs.ModePerm)); err != nil {
		return &os.PathError{Op: "mkdir", Path: path, Err: err}
	}
	return nil
}

func (d Dir) chdir() error {
	if d == CurrentDir {
		return nil
	}
	if err := unix.Fchdir(int(d)); err != nil {
		return &Error{Op: "chdir", Err: err}
	}
	return nil
}

func (d Dir) link(srcDir Dir, src, dst string, flag int) error {
	if err := unix.Linkat(int(srcDir), src, int(d), dst, flag); err != nil {
		return &os.LinkError{Op: "link", Old: src, New: dst, Err: err}
	}
	return nil
}

func (d Dir) symlink(src, dst string) error {
	if err := unix.Symlinkat(src, int(d), dst); err != nil {
		return &os.LinkError{Op: "symlink", Old: src, New: dst, Err: err}
	}
	return nil
}

func (d Dir) unlink(path string, flag int) error {
	if err := unix.Unlinkat(int(d), path, flag); err != nil {
		return &os.PathError{Op: "unlink", Path: path, Err: err}
	}
	return nil
}

func (d Dir) stat(path string, flag int) (fileInfo, error) {
	var stat unix.Stat_t
	if err := unix.Fstatat(int(d), path, &stat, flag); err != nil {
		return fileInfo{}, &os.PathError{Op: "stat", Path: path, Err: err}
	}
	return statToFileInfo(path, stat), nil
}

func (d Dir) utimes(path string, atime, mtime time.Time, flag int) error {
	ats, err := timeToTimespec(atime)
	if err != nil {
		return &os.PathError{Op: "utimes", Path: path, Err: err}
	}
	mts, err := timeToTimespec(mtime)
	if err != nil {
		return &os.PathError{Op: "utimes", Path: path, Err: err}
	}
	if err := unix.UtimesNanoAt(int(d), path, []unix.Timespec{ats, mts}, flag); err != nil {
		return &os.PathError{Op: "utimes", Path: path, Err: err}
	}
	return nil
}

func futimes(f *os.File, atime, mtime time.Time) error {
	atv := unix.NsecToTimeval(atime.UnixNano())
	mtv := unix.NsecToTimeval(mtime.UnixNano())
	if err := unix.Futimes(int(f.Fd()), []unix.Timeval{atv, mtv}); err != nil {
		return &os.PathError{Op: "futimes", Path: f.Name(), Err: err}
	}
	return nil
}

func timeToTimespec(t time.Time) (unix.Timespec, error) {
	if t == UtimeOmit {
		return unix.Timespec{0, unixUtimeOmit}, nil
	}
	return unix.TimeToTimespec(t)
}

func statToFileInfo(path string, stat unix.Stat_t) fileInfo {
	var mode fs.FileMode

	switch stat.Mode & unix.S_IFMT {
	case unix.S_IFDIR:
		mode |= fs.ModeDir
	case unix.S_IFLNK:
		mode |= fs.ModeSymlink
	case unix.S_IFBLK:
		mode |= fs.ModeDevice
	case unix.S_IFCHR:
		mode |= fs.ModeDevice | fs.ModeCharDevice
	case unix.S_IFIFO:
		mode |= fs.ModeNamedPipe
	case unix.S_IFSOCK:
		mode |= fs.ModeSocket
	}

	if stat.Mode&unix.S_ISUID != 0 {
		mode |= fs.ModeSetuid
	}
	if stat.Mode&unix.S_ISGID != 0 {
		mode |= fs.ModeSetgid
	}
	if stat.Mode&unix.S_ISVTX != 0 {
		mode |= fs.ModeSticky
	}

	mode |= fs.FileMode(stat.Mode & (unix.S_IRWXU | unix.S_IRWXG | unix.S_IRWXO))

	return fileInfo{
		name:    filepath.Base(path),
		mode:    mode,
		modTime: time.Unix(stat.Mtim.Unix()),
		stat:    stat,
	}
}
