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
	textWriteRecipientField(ew, "", "Conversation", msgs[0].Conversation)
	fmt.Fprintln(ew)
	for _, msg := range msgs {
		textWriteMessage(ew, &msg)
	}
	return ew.Err()
}

func textWriteMessage(ew *errio.Writer, msg *signal.Message) {
	if msg.IsOutgoing() {
		textWriteField(ew, "", "From", "You")
	} else if msg.Source != nil {
		textWriteRecipientField(ew, "", "From", msg.Source)
	}
	if msg.Type != "" {
		textWriteField(ew, "", "Type", msg.Type)
	} else {
		textWriteField(ew, "", "Type", "unknown")
	}
	if msg.TimeSent != 0 {
		textWriteTimeField(ew, "", "Sent", msg.TimeSent)
	}
	if !msg.IsOutgoing() {
		textWriteTimeField(ew, "", "Received", msg.TimeRecv)
	}
	textWriteAttachmentFields(ew, "", msg.Attachments)
	for _, rct := range msg.Reactions {
		textWriteFieldf(ew, "", "Reaction", "%s from %s", rct.Emoji, rct.Recipient.DetailedDisplayName())
	}
	if len(msg.Edits) == 0 {
		textWriteQuote(ew, "", msg.Quote)
		textWriteBody(ew, "", &msg.Body)
	} else {
		textWriteFieldf(ew, "", "Edited", "%d versions", len(msg.Edits))
		textWriteEditHistory(ew, msg.Edits)
	}
	fmt.Fprintln(ew)
}

func textWriteField(ew *errio.Writer, prefix, field, value string) {
	if prefix != "" {
		prefix += " "
	}
	fmt.Fprintf(ew, "%s%s: %s\n", prefix, field, value)
}

func textWriteFieldf(ew *errio.Writer, prefix, field, format string, a ...any) {
	textWriteField(ew, prefix, field, fmt.Sprintf(format, a...))
}

func textWriteRecipientField(ew *errio.Writer, prefix, field string, rpt *signal.Recipient) {
	textWriteField(ew, prefix, field, rpt.DetailedDisplayName())
}

func textWriteTimeField(ew *errio.Writer, prefix, field string, msec int64) {
	s := "unknown"
	if msec >= 0 {
		s = time.UnixMilli(msec).Format("Mon, 2 Jan 2006 15:04:05 -0700")
	}
	textWriteField(ew, prefix, field, s)
}

func textWriteAttachmentFields(ew *errio.Writer, prefix string, atts []signal.Attachment) {
	for _, att := range atts {
		fileName := "no filename"
		if att.FileName != "" {
			fileName = att.FileName
		}
		textWriteFieldf(ew, prefix, "Attachment", "%s (%s, %d bytes)", fileName, att.ContentType, att.Size)
	}
}

func textWriteBody(ew *errio.Writer, prefix string, body *signal.MessageBody) {
	if body.Text == "" {
		return
	}
	fmt.Fprintln(ew, prefix)
	if prefix != "" {
		prefix += " "
	}
	for _, line := range strings.Split(body.Text, "\n") {
		fmt.Fprintln(ew, prefix+line)
	}
}

func textWriteQuote(ew *errio.Writer, prefix string, qte *signal.Quote) {
	if qte == nil {
		return
	}
	fmt.Fprintln(ew, prefix)
	if prefix != "" {
		prefix += " "
	}
	prefix += ">"
	textWriteRecipientField(ew, prefix, "From", qte.Recipient)
	textWriteTimeField(ew, prefix, "Sent", qte.TimeSent)
	textWriteQuoteAttachmentFields(ew, prefix, qte.Attachments)
	textWriteBody(ew, prefix, &qte.Body)
}

func textWriteQuoteAttachmentFields(ew *errio.Writer, prefix string, atts []signal.QuoteAttachment) {
	for _, att := range atts {
		fileName := "no filename"
		if att.FileName != "" {
			fileName = att.FileName
		}
		textWriteFieldf(ew, prefix, "Attachment", "%s (%s)", fileName, att.ContentType)
	}
}

func textWriteEditHistory(ew *errio.Writer, edits []signal.Edit) {
	fmt.Fprintln(ew)
	prefix := "|"
	for i := range edits {
		textWriteFieldf(ew, prefix, "Version", "%d", len(edits)-i)
		textWriteAttachmentFields(ew, prefix, edits[i].Attachments)
		textWriteTimeField(ew, prefix, "Sent", edits[i].TimeEdit)
		textWriteQuote(ew, prefix, edits[i].Quote)
		textWriteBody(ew, prefix, &edits[i].Body)
		if i+1 < len(edits) {
			fmt.Fprintln(ew, prefix)
		}
	}
}
