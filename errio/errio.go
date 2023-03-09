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

// Package errio provides a wrapper around io.Writer to simplify error
// handling. A wrapper around io.Reader could be added if needed. The package
// was inspired by Rob Pike's errWriter, described at
// https://go.dev/blog/errors-are-values.
package errio

import "io"

type Writer struct {
	w   io.Writer
	err error
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{w: w}
}

func (ew *Writer) Write(p []byte) (int, error) {
	if ew.err != nil {
		return 0, ew.err
	}
	var n int
	n, ew.err = ew.w.Write(p)
	return n, ew.err
}

func (ew *Writer) Err() error {
	return ew.err
}
