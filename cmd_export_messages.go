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
	"errors"
	"io/fs"
	"log"
	"os"

	"github.com/tbvdm/sigtop/at"
	"github.com/tbvdm/sigtop/errio"
	"github.com/tbvdm/sigtop/getopt"
	"github.com/tbvdm/sigtop/signal"
	"github.com/tbvdm/sigtop/util"
)

type formatMode int

const (
	formatJSON formatMode = iota
	formatText
)

type msgMode struct {
	format      formatMode
	incremental bool
}

var cmdExportMessagesEntry = cmdEntry{
	name:  "export-messages",
	alias: "msg",
	usage: "[-i] [-d signal-directory] [-f format] [-s interval] [directory]",
	exec:  cmdExportMessages,
}

func cmdExportMessages(args []string) cmdStatus {
	mode := msgMode{
		format:      formatText,
		incremental: false,
	}

	getopt.ParseArgs("d:f:is:", args)
	var dArg, sArg getopt.Arg
	for getopt.Next() {
		switch opt := getopt.Option(); opt {
		case 'd':
			dArg = getopt.OptionArg()
		case 'f':
			switch arg := getopt.OptionArg().String(); arg {
			case "json":
				mode.format = formatJSON
			case "text":
				mode.format = formatText
			default:
				log.Fatalf("invalid format: %s", arg)
			}
		case 'i':
			mode.incremental = true
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

	if err := util.Unveil(exportDir, "rwc"); err != nil {
		log.Fatal(err)
	}

	// For SQLite/SQLCipher
	if err := util.Unveil("/dev/urandom", "r"); err != nil {
		log.Fatal(err)
	}

	if err := util.Pledge("stdio rpath wpath cpath flock", ""); err != nil {
		log.Fatal(err)
	}

	ctx, err := signal.Open(signalDir)
	if err != nil {
		log.Fatal(err)
	}
	defer ctx.Close()

	if !exportMessages(ctx, exportDir, mode, ival) {
		return cmdError
	}

	return cmdOK
}

func exportMessages(ctx *signal.Context, dir string, mode msgMode, ival signal.Interval) bool {
	d, err := at.Open(dir)
	if err != nil {
		log.Print(err)
		return false
	}
	defer d.Close()

	convs, err := ctx.Conversations()
	if err != nil {
		log.Print(err)
		return false
	}

	ret := true
	for _, conv := range convs {
		if err = exportConversationMessages(ctx, d, &conv, mode, ival); err != nil {
			log.Print(err)
			ret = false
		}
	}

	return ret
}

func exportConversationMessages(ctx *signal.Context, d at.Dir, conv *signal.Conversation, mode msgMode, ival signal.Interval) error {
	msgs, err := ctx.ConversationMessages(conv, ival)
	if err != nil {
		return err
	}

	if len(msgs) == 0 {
		return nil
	}

	f, err := conversationFile(d, conv, mode)
	if err != nil {
		return err
	}
	ew := errio.NewWriter(f)

	switch mode.format {
	case formatJSON:
		err = jsonWriteMessages(ew, msgs)
	case formatText:
		err = textWriteMessages(ew, msgs)
	}

	if err != nil {
		f.Close()
		return err
	}

	return f.Close()
}

func conversationFile(d at.Dir, conv *signal.Conversation, mode msgMode) (*os.File, error) {
	var ext string
	switch mode.format {
	case formatJSON:
		ext = ".json"
	case formatText:
		ext = ".txt"
	}

	flags := os.O_WRONLY | os.O_CREATE
	if !mode.incremental {
		flags |= os.O_EXCL
	}

	name := recipientFilename(conv.Recipient, ext)
	f, err := d.OpenFile(name, flags, 0666)
	if err != nil {
		return nil, err
	}

	return f, nil
}
