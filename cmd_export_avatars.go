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
	"bytes"
	"errors"
	"io/fs"
	"log"
	"os"

	"github.com/tbvdm/go-openbsd"
	"github.com/tbvdm/sigtop/at"
	"github.com/tbvdm/sigtop/getopt"
	"github.com/tbvdm/sigtop/signal"
)

var cmdExportAvatarsEntry = cmdEntry{
	name:  "export-avatars",
	alias: "avt",
	usage: "[-B] [-c conversation] [-d signal-directory] [-k [system:]keyfile] [directory]",
	exec:  cmdExportAvatars,
}

func cmdExportAvatars(args []string) cmdStatus {
	getopt.ParseArgs("Bc:d:k:p:", args)
	var dArg, kArg getopt.Arg
	var selectors []string
	Bflag := false
	for getopt.Next() {
		switch getopt.Option() {
		case 'B':
			Bflag = true
		case 'c':
			selectors = append(selectors, getopt.OptionArg().String())
		case 'd':
			dArg = getopt.OptionArg()
		case 'p':
			log.Print("-p is deprecated; use -k instead")
			fallthrough
		case 'k':
			kArg = getopt.OptionArg()
		}
	}

	if err := getopt.Err(); err != nil {
		log.Fatal(err)
	}

	args = getopt.Args()
	var exportDir string
	switch len(args) {
	case 0:
		exportDir = "."
	case 1:
		exportDir = args[0]
		if err := os.Mkdir(exportDir, 0777); err != nil && !errors.Is(err, fs.ErrExist) {
			log.Fatal(err)
		}
	default:
		return cmdUsage
	}

	key, err := encryptionKeyFromArgument(kArg)
	if err != nil {
		log.Fatal(err)
	}

	signalDir, err := signalDirFromArgument(dArg, Bflag)
	if err != nil {
		log.Fatal(err)
	}

	if err := unveilSignalDir(signalDir); err != nil {
		log.Fatal(err)
	}

	if err := openbsd.Unveil(exportDir, "rwc"); err != nil {
		log.Fatal(err)
	}

	// For SQLite/SQLCipher
	if err := openbsd.Unveil("/dev/urandom", "r"); err != nil {
		log.Fatal(err)
	}

	if err := openbsd.Pledge("stdio rpath wpath cpath flock"); err != nil {
		log.Fatal(err)
	}

	ctx, err := signal.Open(Bflag, signalDir, key)
	if err != nil {
		log.Fatal(err)
	}
	defer ctx.Close()

	if !exportAvatars(ctx, exportDir, selectors) {
		return cmdError
	}

	return cmdOK
}

func exportAvatars(ctx *signal.Context, dir string, selectors []string) bool {
	d, err := at.Open(dir)
	if err != nil {
		log.Print(err)
		return false
	}
	defer d.Close()

	convs, err := selectConversations(ctx, selectors)
	if err != nil {
		log.Print(err)
		return false
	}

	ret := true
	for _, conv := range convs {
		if err := exportAvatar(ctx, d, conv.Recipient); err != nil {
			log.Print(err)
			ret = false
		}
	}

	return ret
}

func exportAvatar(ctx *signal.Context, d at.Dir, rpt *signal.Recipient) error {
	if rpt.Avatar.Path == "" {
		return nil
	}

	data, err := ctx.ReadAvatar(&rpt.Avatar)
	if err != nil {
		return err
	}

	f, err := d.OpenFile(avatarFilename(rpt, data), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		return err
	}
	if _, err = f.Write(data); err != nil {
		f.Close()
		return err
	}

	return f.Close()
}

func avatarFilename(rpt *signal.Recipient, data []byte) string {
	equals := func(b []byte, s string) bool { return bytes.Equal(b, []byte(s)) }

	var ext string
	switch {
	case len(data) >= 3 && equals(data[:3], "\xff\xd8\xff"):
		ext = ".jpg"
	case len(data) >= 8 && equals(data[:8], "\x89PNG\r\n\x1a\n"):
		ext = ".png"
	case len(data) >= 12 && equals(data[:4], "RIFF") && equals(data[8:12], "WEBP"):
		ext = ".webp"
	}

	return recipientFilename(rpt, ext)
}
