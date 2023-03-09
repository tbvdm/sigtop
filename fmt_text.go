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
	"strings"
	"time"

	"github.com/tbvdm/sigtop/errio"
	"github.com/tbvdm/sigtop/signal"
)

func textWriteMessages(ew *errio.Writer, msgs []signal.Message) error {
	textWriteRecipientField(ew, "Conversation", msgs[0].Conversation)
	fmt.Fprintln(ew)
	for _, msg := range msgs {
		textWriteMessage(ew, &msg)
	}
	return ew.Err()
}

func textWriteMessage(ew *errio.Writer, msg *signal.Message) {
	if msg.IsOutgoing() {
		textWriteField(ew, "From", "You")
	} else if msg.Source != nil {
		textWriteRecipientField(ew, "From", msg.Source)
	}
	if msg.Type != "" {
		textWriteField(ew, "Type", msg.Type)
	} else {
		textWriteField(ew, "Type", "unknown")
	}
	if msg.TimeSent != 0 {
		textWriteTimeField(ew, "Sent", msg.TimeSent)
	}
	if !msg.IsOutgoing() {
		textWriteTimeField(ew, "Received", msg.TimeRecv)
	}
	for _, att := range msg.Attachments {
		textWriteAttachmentField(ew, &att)
	}
	for _, rct := range msg.Reactions {
		textWriteFieldf(ew, "Reaction", "%s from %s", rct.Emoji, rct.Recipient.DisplayName())
	}
	if msg.Quote != nil {
		textWriteQuote(ew, msg.Quote)
	}
	if msg.Body.Text == "" {
		fmt.Fprintln(ew)
	} else {
		fmt.Fprintf(ew, "\n%s\n\n", msg.Body.Text)
	}
}

func textWriteField(ew *errio.Writer, field, value string) {
	fmt.Fprintf(ew, "%s: %s\n", field, value)
}

func textWriteFieldf(ew *errio.Writer, field, format string, a ...any) {
	textWriteField(ew, field, fmt.Sprintf(format, a...))
}

func textWriteRecipientField(ew *errio.Writer, field string, rpt *signal.Recipient) {
	value := rpt.DisplayName()
	if rpt != nil {
		if rpt.Type == signal.RecipientTypeGroup {
			value += " (group)"
		} else if rpt.Contact.Phone != "" {
			value += " (" + rpt.Contact.Phone + ")"
		}
	}
	textWriteField(ew, field, value)
}

func textWriteTimeField(ew *errio.Writer, field string, msec int64) {
	textWriteField(ew, field, time.UnixMilli(msec).Format("Mon, 2 Jan 2006 15:04:05 -0700"))
}

func textWriteAttachmentField(ew *errio.Writer, att *signal.Attachment) {
	fileName := "no filename"
	if att.FileName != "" {
		fileName = att.FileName
	}
	textWriteFieldf(ew, "Attachment", "%s (%s, %d bytes)", fileName, att.ContentType, att.Size)
}

func textWriteQuote(ew *errio.Writer, qte *signal.Quote) {
	fmt.Fprint(ew, "\n> ")
	textWriteRecipientField(ew, "From", qte.Recipient)
	fmt.Fprint(ew, "> ")
	textWriteTimeField(ew, "Sent", qte.ID)
	for _, att := range qte.Attachments {
		fmt.Fprint(ew, "> ")
		textWriteQuoteAttachmentField(ew, &att)
	}
	if qte.Body.Text != "" {
		fmt.Fprint(ew, ">\n> ")
		fmt.Fprintln(ew, strings.ReplaceAll(qte.Body.Text, "\n", "\n> "))
	}
}

func textWriteQuoteAttachmentField(ew *errio.Writer, att *signal.QuoteAttachment) {
	fileName := "no filename"
	if att.FileName != "" {
		fileName = att.FileName
	}
	textWriteFieldf(ew, "Attachment", "%s (%s)", fileName, att.ContentType)
}
