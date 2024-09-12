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

package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tbvdm/go-openbsd"
	"github.com/tbvdm/sigtop/at"
	"github.com/tbvdm/sigtop/getopt"
	"github.com/tbvdm/sigtop/signal"
)

const incrementalFile = ".incremental"

type exportMode int

type mtimeMode int

const (
	mtimeNone mtimeMode = iota
	mtimeSent
	mtimeRecv
)

type attMode struct {
	mtime       mtimeMode
	incremental bool
}

var cmdExportAttachmentsEntry = cmdEntry{
	name:  "export-attachments",
	alias: "att",
	usage: "[-iMm] [-c conversation] [-d signal-directory] [-k [system:]keyfile] [-s interval] [directory]",
	exec:  cmdExportAttachments,
}

func cmdExportAttachments(args []string) cmdStatus {
	mode := attMode{
		mtime:       mtimeNone,
		incremental: false,
	}

	getopt.ParseArgs("c:d:ik:Mmp:s:", args)
	var dArg, kArg, sArg getopt.Arg
	var selectors []string
	for getopt.Next() {
		switch getopt.Option() {
		case 'c':
			selectors = append(selectors, getopt.OptionArg().String())
		case 'd':
			dArg = getopt.OptionArg()
		case 'i':
			mode.incremental = true
		case 'M':
			mode.mtime = mtimeSent
		case 'm':
			mode.mtime = mtimeRecv
		case 'p':
			log.Print("-p is deprecated; use -k instead")
			fallthrough
		case 'k':
			kArg = getopt.OptionArg()
		case 's':
			sArg = getopt.OptionArg()
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

	key, err := encryptionKeyFromFile(kArg)
	if err != nil {
		log.Fatal(err)
	}

	var signalDir string
	if dArg.Set() {
		signalDir = dArg.String()
	} else {
		var err error
		signalDir, err = signal.DesktopDir()
		if err != nil {
			log.Fatal(err)
		}
	}

	var ival signal.Interval
	if sArg.Set() {
		var err error
		ival, err = parseInterval(sArg.String())
		if err != nil {
			log.Fatal(err)
		}
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

	if err := unveilMimeFiles(); err != nil {
		log.Fatal(err)
	}

	if mode.mtime == mtimeNone {
		if err := openbsd.Pledge("stdio rpath wpath cpath flock"); err != nil {
			log.Fatal(err)
		}
	} else {
		if err := openbsd.Pledge("stdio rpath wpath cpath flock fattr"); err != nil {
			log.Fatal(err)
		}
	}

	if err := addContentTypes(); err != nil {
		log.Print(err)
	}

	var ctx *signal.Context
	if key == nil {
		ctx, err = signal.Open(signalDir)
	} else {
		ctx, err = signal.OpenWithEncryptionKey(signalDir, key)
	}
	if err != nil {
		log.Fatal(err)
	}
	defer ctx.Close()

	if !exportAttachments(ctx, exportDir, mode, selectors, ival) {
		return cmdError
	}

	return cmdOK
}

func exportAttachments(ctx *signal.Context, dir string, mode attMode, selectors []string, ival signal.Interval) bool {
	d, err := at.Open(dir)
	if err != nil {
		log.Print(err)
		return false
	}
	defer d.Close()

	var exported map[string]bool
	if mode.incremental {
		var err error
		if exported, err = readIncrementalFile(d); err != nil {
			log.Print(err)
			return false
		}
	}

	convs, err := selectConversations(ctx, selectors)
	if err != nil {
		log.Print(err)
		return false
	}

	ret := true
	usedFilenames := make(map[string]bool)
	for _, conv := range convs {
		var ok bool
		if ok, exported = exportConversationAttachments(ctx, d, &conv, mode, exported, ival, usedFilenames); !ok {
			ret = false
		}
	}

	if mode.incremental {
		if err := writeIncrementalFile(d, exported); err != nil {
			log.Print(err)
			return false
		}
	}

	return ret
}

func exportConversationAttachments(ctx *signal.Context, d at.Dir, conv *signal.Conversation, mode attMode, exported map[string]bool, ival signal.Interval, usedFilenames map[string]bool) (bool, map[string]bool) {
	atts, err := ctx.ConversationAttachments(conv, ival)
	if err != nil {
		log.Print(err)
		return false, exported
	}

	if len(atts) == 0 {
		return true, exported
	}

	cd, err := conversationDir(d, conv, usedFilenames)
	if err != nil {
		log.Print(err)
		return false, exported
	}

	ret := true
	for _, att := range atts {
		id := filepath.Base(att.Path)
		if mode.incremental && exported[id] {
			continue
		}
		if att.Path == "" {
			var msg string
			if att.Pending {
				msg = "skipping pending attachment"
			} else {
				msg = "skipping attachment without path"
				ret = false
			}
			log.Printf("%s (conversation: %q, sent: %s)", msg, conv.Recipient.DisplayName(), time.UnixMilli(att.TimeSent).Format("2006-01-02 15:04:05"))
			continue
		}
		path, err := attachmentFilename(cd, &att)
		if err != nil {
			log.Print(err)
			ret = false
			continue
		}
		if err := copyAttachment(ctx, cd, path, &att); err != nil {
			log.Print(err)
			ret = false
			continue
		}
		if err := setAttachmentModTime(cd, path, &att, mode.mtime); err != nil {
			log.Print(err)
			ret = false
		}
		if mode.incremental {
			exported[id] = true
		}
	}

	return ret, exported
}

func conversationDir(d at.Dir, conv *signal.Conversation, usedFilenames map[string]bool) (at.Dir, error) {
	name := recipientFilename(conv.Recipient, "", usedFilenames)
	if err := d.Mkdir(name, 0777); err != nil && !errors.Is(err, fs.ErrExist) {
		return at.InvalidDir, err
	}
	return d.OpenDir(name)
}

func attachmentFilename(d at.Dir, att *signal.Attachment) (string, error) {
	var name string
	if att.FileName != "" {
		name = sanitiseFilename(att.FileName)
	} else {
		var ext string
		if att.ContentType == "" {
			log.Printf("attachment without content type (sent: %d)", att.TimeSent)
		} else {
			var err error
			ext, err = extensionFromContentType(att.ContentType)
			if err != nil {
				return "", err
			}
			if ext == "" {
				log.Printf("no filename extension for content type %q (sent: %d)", att.ContentType, att.TimeSent)
			}
		}
		name = "attachment-" + time.UnixMilli(att.TimeSent).Format("2006-01-02-15-04-05") + ext
	}

	return uniqueFilename(d, name)
}

func uniqueFilename(d at.Dir, path string) (string, error) {
	if ok, err := fileExists(d, path); !ok {
		return path, err
	}

	suffix := filepath.Ext(path)
	prefix := strings.TrimSuffix(path, suffix)

	for i := 2; i > 0; i++ {
		newPath := fmt.Sprintf("%s-%d%s", prefix, i, suffix)
		if ok, err := fileExists(d, newPath); !ok {
			return newPath, err
		}
	}

	return "", fmt.Errorf("%s: cannot generate unique name", path)
}

func fileExists(d at.Dir, path string) (bool, error) {
	if _, err := d.Stat(path, at.SymlinkNoFollow); err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return false, err
		}
		return false, nil
	}
	return true, nil
}

func copyAttachment(ctx *signal.Context, d at.Dir, path string, att *signal.Attachment) error {
	f, err := d.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		return err
	}
	if err := ctx.WriteAttachment(att, f); err != nil {
		f.Close()
		d.Unlink(path, 0)
		return fmt.Errorf("cannot export %s: %w", path, err)
	}

	return f.Close()
}

func setAttachmentModTime(d at.Dir, path string, att *signal.Attachment, mode mtimeMode) error {
	var mtime int64
	switch mode {
	case mtimeSent:
		mtime = att.TimeSent
	case mtimeRecv:
		mtime = att.TimeRecv
	default:
		return nil
	}
	return d.Utimes(path, at.UtimeOmit, time.UnixMilli(mtime), at.SymlinkNoFollow)
}

func readIncrementalFile(d at.Dir) (map[string]bool, error) {
	exported := make(map[string]bool)

	f, err := d.OpenFile(incrementalFile, os.O_RDONLY, 0666)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return exported, nil
		}
		return nil, err
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		exported[s.Text()] = true
	}

	return exported, s.Err()
}

func writeIncrementalFile(d at.Dir, exported map[string]bool) error {
	f, err := d.OpenFile(incrementalFile, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}

	for id := range exported {
		if _, err := fmt.Fprintln(f, id); err != nil {
			f.Close()
			return err
		}
	}

	return f.Close()
}
