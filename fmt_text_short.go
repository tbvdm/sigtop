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
	"strings"
	"time"

	"github.com/tbvdm/sigtop/errio"
	"github.com/tbvdm/sigtop/signal"
)

func textShortWriteMessages(ew *errio.Writer, msgs []signal.Message) error {
	for _, msg := range msgs {
		textShortWriteMessage(ew, &msg)
	}
	return ew.Err()
}

func textShortWriteMessage(ew *errio.Writer, msg *signal.Message) {
	name := "You"
	if !msg.IsOutgoing() {
		name = msg.Source.DisplayName()
	}
	fmt.Fprintf(ew, "%s %s:", textShortFormatTime(msg.TimeSent), name)
	if msg.Type != "incoming" && msg.Type != "outgoing" {
		fmt.Fprintf(ew, " [%s message]", msg.Type)
	} else {
		var details []string
		if msg.Quote != nil {
			details = append(details, fmt.Sprintf("reply to %s on %s", msg.Quote.Recipient.DisplayName(), textShortFormatTime(msg.Quote.ID)))
		}
		if len(msg.Edits) > 0 {
			details = append(details, "edited")
		}
		if len(msg.Attachments) > 0 {
			plural := ""
			if len(msg.Attachments) > 1 {
				plural = "s"
			}
			details = append(details, fmt.Sprintf("%d attachment%s", len(msg.Attachments), plural))
		}
		if len(details) > 0 {
			fmt.Fprintf(ew, " [%s]", strings.Join(details, ", "))
		}
		if msg.Body.Text != "" {
			fmt.Fprint(ew, " "+msg.Body.Text)
		}
	}
	fmt.Fprintln(ew)
}

func textShortFormatTime(msec int64) string {
	return time.UnixMilli(msec).Format("2006-01-02 15:04")
}
