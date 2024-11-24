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
	usage: "[-BD] [-d signal-directory] [-k [system:]keyfile] [file]",
	exec:  cmdExportKey,
}

func cmdExportKey(args []string) cmdStatus {
	getopt.ParseArgs("BDd:k:", args)
	var dArg, kArg getopt.Arg
	exportDBKey := false
	Bflag := false
	for getopt.Next() {
		switch opt := getopt.Option(); opt {
		case 'B':
			Bflag = true
		case 'D':
			exportDBKey = true
		case 'd':
			dArg = getopt.OptionArg()
		case 'k':
			kArg = getopt.OptionArg()
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

	key, err := encryptionKeyFromFile(kArg)
	if err != nil {
		log.Fatal(err)
	}

	var signalDir string
	if dArg.Set() {
		signalDir = dArg.String()
	} else {
		var err error
		signalDir, err = signal.DesktopDir(Bflag)
		if err != nil {
			log.Fatal(err)
		}
	}

	if err := unveilSignalDir(signalDir); err != nil {
		log.Fatal(err)
	}

	// For SQLite/SQLCipher
	if err := openbsd.Unveil("/dev/urandom", "r"); err != nil {
		log.Fatal(err)
	}

	if err := openbsd.Pledge("stdio rpath wpath cpath flock"); err != nil {
		log.Fatal(err)
	}

	var ctx *signal.Context
	if key == nil {
		ctx, err = signal.Open(Bflag, signalDir)
	} else {
		ctx, err = signal.OpenWithEncryptionKey(Bflag, signalDir, key)
	}
	if err != nil {
		log.Fatal(err)
	}
	defer ctx.Close()

	var data []byte
	if exportDBKey {
		if data, err = ctx.DatabaseKey(); err != nil {
			log.Printf("cannot get database key: %v", err)
			return cmdError
		}
	} else {
		if data, err = ctx.EncryptionKey(); err != nil {
			log.Printf("cannot get encryption key: %v", err)
			return cmdError
		}
	}
	fmt.Fprintln(outfile, string(data))

	if outfile != os.Stdout {
		if err := outfile.Close(); err != nil {
			log.Print(err)
			return cmdError
		}
	}

	return cmdOK
}
