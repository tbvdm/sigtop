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

package signal

import (
	"encoding/json"
	"fmt"
)

type quoteJSON struct {
	Attachments []quoteAttachmentJSON `json:"attachments"`
	Author      string                `json:"author"`
	AuthorUUID  string                `json:"authorUuid"`
	Mentions    []mentionJSON         `json:"bodyRanges"`
	// The ID is a JSON number now, but apparently it used to be a number
	// encoded as a JSON string. See sigtop GitHub issue 9 and
	// Signal-Desktop commit ddbbe3a6b1b725007597536a39651ae845366920.
	// Using a json.Number allows us to handle both cases.
	ID          json.Number           `json:"id"`
	Text        string                `json:"text"`
}

type quoteAttachmentJSON struct {
	ContentType string `json:"contentType"`
	FileName    string `json:"fileName"`
}

type Quote struct {
	ID          int64
	Recipient   *Recipient
	Body        MessageBody
	Attachments []QuoteAttachment
}

type QuoteAttachment struct {
	FileName    string
	ContentType string
}

func (c *Context) parseQuoteJSON(msg *Message, jmsg *messageJSON) error {
	if jmsg.Quote == nil {
		return nil
	}

	var qte Quote
	var err error

	if jmsg.Quote.ID.String() == "" {
		return fmt.Errorf("quote without ID")
	}
	if qte.ID, err = jmsg.Quote.ID.Int64(); err != nil {
		return fmt.Errorf("cannot parse quote ID: %w", err)
	}

	// Newer quotes have an "authorUuid" attribute, older quotes have an
	// "author" attribute containing a phone number
	if jmsg.Quote.AuthorUUID != "" {
		if qte.Recipient, err = c.recipientFromUUID(jmsg.Quote.AuthorUUID); err != nil {
			return err
		}
	} else if jmsg.Quote.Author != "" {
		if qte.Recipient, err = c.recipientFromPhone(jmsg.Quote.Author); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("quote without authorUuid and author")
	}

	qte.Body.Text = jmsg.Quote.Text

	if err := c.parseMentionJSON(&qte.Body, jmsg.Quote.Mentions); err != nil {
		return err
	}

	for _, jatt := range jmsg.Quote.Attachments {
		// Skip long-message attachments
		if jatt.ContentType == LongTextType {
			continue
		}
		att := QuoteAttachment{
			FileName:    jatt.FileName,
			ContentType: jatt.ContentType,
		}
		qte.Attachments = append(qte.Attachments, att)
	}

	msg.Quote = &qte
	return nil
}
