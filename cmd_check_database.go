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

	"github.com/tbvdm/go-openbsd"
	"github.com/tbvdm/sigtop/getopt"
	"github.com/tbvdm/sigtop/signal"
)

var cmdCheckDatabaseEntry = cmdEntry{
	name:  "check-database",
	alias: "check",
	usage: "[-B] [-d signal-directory] [-k [system:]keyfile]",
	exec:  cmdCheckDatabase,
}

func cmdCheckDatabase(args []string) cmdStatus {
	getopt.ParseArgs("Bd:k:p:", args)
	var dArg, kArg getopt.Arg
	Bflag := false
	for getopt.Next() {
		switch opt := getopt.Option(); opt {
		case 'B':
			Bflag = true
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

	if len(getopt.Args()) != 0 {
		return cmdUsage
	}

	key, err := encryptionKeyFromFile(kArg)
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

	results, err := ctx.CheckDatabase()
	if err != nil {
		log.Print(err)
		return cmdError
	}

	if len(results) > 0 {
		for _, s := range results {
			fmt.Println(s)
		}
		return cmdError
	}

	return cmdOK
}
