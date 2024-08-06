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
	"errors"
	"os"
	"strings"
)

// Based on EditHistoryType in ts/model-types.d.ts in the Signal-Desktop
// repository
type editJSON struct {
	Attachments []attachmentJSON `json:"attachments"`
	Body        string           `json:"body"`
	Mentions    []mentionJSON    `json:"bodyRanges"`
	Quote       *quoteJSON       `json:"quote"`
	Timestamp   int64            `json:"timestamp"`
}

type Edit struct {
	Body        MessageBody
	Attachments []Attachment
	Quote       *Quote
	TimeEdit    int64
}

func (c *Context) parseEditJSON(msg *Message, jmsg *messageJSON) error {
	for _, jedit := range jmsg.Edits {
		edit := Edit{
			Body:        MessageBody{Text: jedit.Body},
			Attachments: c.parseAttachmentJSON(msg, jedit.Attachments),
			TimeEdit:    jedit.Timestamp,
		}
		var err error
		if edit.Body.Mentions, err = c.parseMentionJSON(jedit.Mentions); err != nil {
			return err
		}
		if edit.Quote, err = c.parseQuoteJSON(jedit.Quote); err != nil {
			return err
		}
		if err = c.fixEditedLongMessage(&edit); err != nil {
			// Fixing edited long messages is a best-effort
			// attempt. Just report the error and move on.
			msg.logError(err, "cannot fix edited long message")
		}
		msg.Edits = append(msg.Edits, edit)
	}
	return nil
}

// fixEditedLongMessage restores the complete message text from a long-text
// attachment, if there is one. This works around what appears to be a bug in
// Signal Desktop; see Signal Desktop issue 6641.
func (c *Context) fixEditedLongMessage(edit *Edit) error {
	for i, att := range edit.Attachments {
		if att.ContentType != LongTextType {
			continue
		}

		data, err := c.readAttachment(&att)
		if err != nil {
			// Signal Desktop considers long-message attachments of
			// edits to be orphaned, and eventually removes them
			// from disk
			if errors.Is(err, os.ErrNotExist) {
				return errors.New("long-message attachment not or no longer available")
			}
			return err
		}

		longText := string(data)

		if !strings.HasPrefix(longText, edit.Body.Text) {
			return errors.New("long-message attachment does not match body text")
		}

		edit.Body.Text = longText
		edit.Attachments = append(edit.Attachments[:i], edit.Attachments[i+1:]...)
		break
	}

	return nil
}
