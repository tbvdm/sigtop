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
	"os"
	"path/filepath"
	"strings"
)

type attachmentJSON struct {
	ContentType string `json:"contentType"`
	FileName    string `json:"fileName"`
	Size        int64  `json:"size"`
	Pending     bool   `json:"pending"`
	Path        string `json:"path"`
}

type Attachment struct {
	Path        string
	FileName    string
	ContentType string
	Size        int64
	TimeSent    int64
	TimeRecv    int64
	Pending     bool
}

func (c *Context) parseAttachmentJSON(msg *Message, jatts []attachmentJSON) []Attachment {
	atts := make([]Attachment, 0, len(jatts))
	for _, jatt := range jatts {
		att := Attachment{
			Path:        jatt.Path,
			FileName:    jatt.FileName,
			ContentType: jatt.ContentType,
			Size:        jatt.Size,
			TimeSent:    msg.TimeSent,
			TimeRecv:    msg.TimeRecv,
			Pending:     jatt.Pending,
		}
		atts = append(atts, att)
	}
	return atts
}

func (c *Context) ConversationAttachments(conv *Conversation, ival Interval) ([]Attachment, error) {
	msgs, err := c.ConversationMessages(conv, ival)
	if err != nil {
		return nil, err
	}

	var atts []Attachment
	for _, msg := range msgs {
		atts = append(atts, msg.Attachments...)
	}

	return atts, nil
}

func (c *Context) AttachmentPath(att *Attachment) string {
	return c.absoluteAttachmentPath(att.Path)
}

func (c *Context) absoluteAttachmentPath(path string) string {
	if path == "" {
		return ""
	}

	// Replace foreign path separators, if any
	var foreignSep string
	if os.PathSeparator == '/' {
		foreignSep = "\\"
	} else {
		foreignSep = "/"
	}
	path = strings.Replace(path, foreignSep, string(os.PathSeparator), -1)

	return filepath.Join(c.dir, AttachmentDir, path)
}
