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
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/tbvdm/go-openbsd"
	"github.com/tbvdm/sigtop/signal"
	"github.com/tbvdm/sigtop/util"
)

type cmdStatus int

const (
	cmdOK cmdStatus = iota
	cmdError
	cmdUsage
)

type cmdEntry struct {
	name  string
	alias string
	usage string
	exec  func([]string) cmdStatus
}

var cmdEntries = []cmdEntry{
	cmdCheckDatabaseEntry,
	cmdExportAttachmentsEntry,
	cmdExportDatabaseEntry,
	cmdExportMessagesEntry,
	cmdQueryDatabaseEntry,
}

func main() {
	util.SetLog()

	if len(os.Args) < 2 {
		usage("command", "[argument ...]")
	}

	cmd := command(os.Args[1])
	if cmd == nil {
		log.Fatalln("invalid command:", os.Args[1])
	}

	switch cmd.exec(os.Args[2:]) {
	case cmdError:
		os.Exit(1)
	case cmdUsage:
		usage(cmd.name, cmd.usage)
	}
}

func usage(cmd, args string) {
	fmt.Fprintln(os.Stderr, "usage:", util.Progname(), cmd, args)
	os.Exit(1)
}

func command(name string) *cmdEntry {
	for _, cmd := range cmdEntries {
		if name == cmd.name || name == cmd.alias {
			return &cmd
		}
	}
	return nil
}

func unveilSignalDir(dir string) error {
	if err := openbsd.Unveil(dir, "r"); err != nil {
		return err
	}

	// SQLite/SQLCipher needs to create the WAL and shared-memory files if
	// they don't exist already. See https://www.sqlite.org/tempfiles.html.

	walFile := filepath.Join(dir, signal.DatabaseFile+"-wal")
	shmFile := filepath.Join(dir, signal.DatabaseFile+"-shm")

	if err := openbsd.Unveil(walFile, "rwc"); err != nil {
		return err
	}

	if err := openbsd.Unveil(shmFile, "rwc"); err != nil {
		return err
	}

	return nil
}

func recipientFilename(rpt *signal.Recipient, ext string) string {
	var detail string

	switch rpt.Type {
	case signal.RecipientTypeContact:
		detail = rpt.Contact.Phone
	case signal.RecipientTypeGroup:
		detail = "group"
	}

	if detail != "" {
		detail = " (" + detail + ")"
	}

	return sanitiseFilename(rpt.DisplayName() + detail + ext)
}
