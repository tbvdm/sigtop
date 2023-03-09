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

package signal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

func warn(format string, a ...any) {
	fmt.Fprintf(os.Stderr, format, a...)
	if len(format) > 0 && format[len(format)-1] != '\n' {
		fmt.Fprintln(os.Stderr)
	}
}

func dumpMessageJSON(msg *Message) {
	var buf bytes.Buffer
	if err := json.Indent(&buf, []byte(msg.JSON), "", "  "); err != nil {
		warn("cannot dump message JSON data: %v", err)
		return
	}
	fmt.Fprintln(os.Stderr, buf.String())
}
