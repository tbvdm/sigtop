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
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/tbvdm/sigtop/signal"
)

func parseInterval(str string) (signal.Interval, error) {
	minStr, maxStr, found := strings.Cut(str, ",")
	if !found {
		maxStr = minStr
	}

	min, err := parseTime(minStr, false)
	if err != nil {
		return signal.Interval{}, err
	}

	max, err := parseTime(maxStr, true)
	if err != nil {
		return signal.Interval{}, err
	}

	return signal.Interval{min, max}, nil
}

func parseTime(str string, max bool) (time.Time, error) {
	if str == "" {
		return time.Time{}, nil
	}

	year, month, day := 0, 0, 0
	nsec := time.Duration(0)

	switch len(str) {
	case 4: // yyyy
		year = 1
	case 7: // yyyy-mm
		month = 1
	case 10: // yyyy-mm-dd
		day = 1
	case 13: // yyyy-mm-ddThh
		nsec = time.Hour
	case 16: // yyyy-mm-ddThh:mm
		nsec = time.Minute
	case 19: // yyyy-mm-ddThh:mm:ss
		nsec = time.Second
	default:
		return time.Time{}, invalidTimeError(str)
	}

	layout := "2006-01-02T15:04:05"
	t, err := time.ParseInLocation(layout[:len(str)], str, time.Local)
	if err != nil {
		var perr *time.ParseError
		if errors.As(err, &perr) {
			if perr.Message == "" {
				err = invalidTimeError(str)
			} else {
				err = fmt.Errorf("%s%s", str, perr.Message)
			}
		}
		return t, err
	}

	if max {
		t = t.AddDate(year, month, day)
		t = t.Add(nsec)
		t = t.Add(-1)
	}

	return t, err
}

func invalidTimeError(str string) error {
	return fmt.Errorf("%s: invalid time", str)
}
