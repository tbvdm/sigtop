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

package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/tbvdm/go-openbsd"
	"github.com/tbvdm/sigtop/getopt"
	"github.com/tbvdm/sigtop/signal"
)

var cmdQueryDatabaseEntry = cmdEntry{
	name:  "query-database",
	alias: "query",
	usage: "[-d signal-directory] [-p passfile] query",
	exec:  cmdQueryDatabase,
}

func cmdQueryDatabase(args []string) cmdStatus {
	getopt.ParseArgs("d:p:", args)
	var dArg, pArg getopt.Arg
	for getopt.Next() {
		switch opt := getopt.Option(); opt {
		case 'd':
			dArg = getopt.OptionArg()
		case 'p':
			pArg = getopt.OptionArg()
		}
	}

	if err := getopt.Err(); err != nil {
		log.Fatal(err)
	}

	args = getopt.Args()
	if len(args) != 1 {
		return cmdUsage
	}

	query := args[0]

	password, err := passwordFromFile(pArg)
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
	if password == nil {
		ctx, err = signal.Open(signalDir)
	} else {
		ctx, err = signal.OpenWithPassword(signalDir, password)
	}
	if err != nil {
		log.Fatal(err)
	}
	defer ctx.Close()

	rows, err := ctx.QueryDatabase(query)
	if err != nil {
		log.Print(err)
		return cmdError
	}

	for _, cols := range rows {
		fmt.Println(strings.Join(cols, "|"))
	}

	return cmdOK
}
