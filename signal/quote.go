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
	// Newer quotes have an "authorAci" (since database version 88) or
	// "authorUuid" field. Older quotes have an "author" field containing a
	// phone number.
	Author      string                `json:"author"`
	AuthorUUID  string                `json:"authorUuid"`
	AuthorACI   string                `json:"authorAci"`
	Mentions    []mentionJSON         `json:"bodyRanges"`
	// The "id" field is a JSON number now, but apparently it used to be a
	// number encoded as a JSON string. See sigtop GitHub issue 9 and
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

func (c *Context) parseQuoteJSON(jqte *quoteJSON) (*Quote, error) {
	if jqte == nil {
		return nil, nil
	}

	var qte Quote
	var err error

	if jqte.ID.String() == "" {
		return nil, fmt.Errorf("quote without ID")
	}
	if qte.ID, err = jqte.ID.Int64(); err != nil {
		return nil, fmt.Errorf("cannot parse quote ID: %w", err)
	}

	switch {
	case jqte.AuthorACI != "":
		if qte.Recipient, err = c.recipientFromACI(jqte.AuthorACI); err != nil {
			return nil, err
		}
	case jqte.AuthorUUID != "":
		if qte.Recipient, err = c.recipientFromACI(jqte.AuthorUUID); err != nil {
			return nil, err
		}
	case jqte.Author != "":
		if qte.Recipient, err = c.recipientFromPhone(jqte.Author); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("quote without author")
	}

	qte.Body.Text = jqte.Text

	if qte.Body.Mentions, err = c.parseMentionJSON(jqte.Mentions); err != nil {
		return nil, err
	}

	for _, jatt := range jqte.Attachments {
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

	return &qte, nil
}
