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

package getopt

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"
)

var (
	opts   string
	args   []string
	argInd int
	optInd int
	opt    rune
	optArg string
	err    error
)

type Arg struct {
	arg string
	set bool
}

func (a Arg) Set() bool {
	return a.set
}

func (a Arg) String() string {
	return a.arg
}

func (a Arg) Int() (int, error) {
	i, err := strconv.ParseInt(a.arg, 0, 0)
	return int(i), err
}

func (a Arg) Int64() (int64, error) {
	return strconv.ParseInt(a.arg, 0, 64)
}

func (a Arg) Float() (float64, error) {
	return strconv.ParseFloat(a.arg, 64)
}

func Parse(opts string) {
	ParseArgs(opts, os.Args[1:])
}

func ParseArgs(newOpts string, newArgs []string) {
	opts = newOpts
	args = newArgs
	argInd = 0
	optInd = 0
	err = nil
}

func Next() bool {
	if err != nil || argInd == len(args) {
		return false
	}

	if optInd == 0 {
		if args[argInd] == "" || args[argInd][0] != '-' || args[argInd] == "-" {
			return false
		}
		if args[optInd] == "--" {
			argInd++
			return false
		}
		optInd = 1
	}

	var optLen int
	opt, optLen = utf8.DecodeRuneInString(args[argInd][optInd:])
	if opt == utf8.RuneError && optLen < 2 {
		err = fmt.Errorf("invalid UTF-8 in argument")
		return false
	}

	ind := strings.IndexRune(opts, opt)
	if ind == -1 {
		err = fmt.Errorf("-%c: invalid option", opt)
		return false
	}

	if ind+optLen == len(opts) || opts[ind+optLen] != ':' {
		if optInd+optLen < len(args[argInd]) {
			optInd += optLen
		} else {
			argInd++
			optInd = 0
		}
		optArg = ""
	} else {
		if optInd+optLen < len(args[argInd]) {
			optArg = args[argInd][optInd+optLen:]
		} else {
			if argInd+1 == len(args) {
				err = fmt.Errorf("-%c: missing argument", opt)
				return false
			}
			argInd++
			optArg = args[argInd]
		}
		argInd++
		optInd = 0
	}

	return true
}

func Option() rune {
	if err != nil {
		return 0
	}
	return opt
}

func OptionArg() Arg {
	if err != nil {
		return Arg{}
	}
	return Arg{optArg, true}
}

func Args() []string {
	return args[argInd:]
}

func Err() error {
	return err
}
