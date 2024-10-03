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
	"fmt"
	"log"
	"os"

	"github.com/tbvdm/go-openbsd"
	"github.com/tbvdm/sigtop/getopt"
	"github.com/tbvdm/sigtop/signal"
)

var cmdExportKeyEntry = cmdEntry{
	name:  "export-key",
	alias: "key",
	usage: "[-d signal-directory] [file]",
	exec:  cmdExportKey,
}

func cmdExportKey(args []string) cmdStatus {
	getopt.ParseArgs("d:", args)
	var dArg getopt.Arg
	for getopt.Next() {
		switch opt := getopt.Option(); opt {
		case 'd':
			dArg = getopt.OptionArg()
		}
	}

	if err := getopt.Err(); err != nil {
		log.Fatal(err)
	}

	args = getopt.Args()
	var outfile *os.File
	switch len(args) {
	case 0:
		outfile = os.Stdout
	case 1:
		var err error
		if outfile, err = os.OpenFile(args[0], os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600); err != nil {
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

	if err := unveilSignalDir(signalDir); err != nil {
		log.Fatal(err)
	}

	if err := openbsd.Pledge("stdio rpath"); err != nil {
		log.Fatal(err)
	}

	key, err := signal.EncryptionKey(signalDir)
	if err != nil {
		log.Fatalf("cannot get encryption key: %v", err)
	}
	fmt.Fprintln(outfile, string(key))

	if outfile != os.Stdout {
		if err := outfile.Close(); err != nil {
			log.Print(err)
			return cmdError
		}
	}

	return cmdOK
}
