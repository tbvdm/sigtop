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
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	cipherKeySize = 32
	ivSize        = aes.BlockSize
	macKeySize    = 32
	macSize       = sha256.Size
)

type attachmentFile struct {
	Version int    `json:"version"`
	Path    string `json:"path"`
	Keys    string `json:"localKey"`
	Size    int64  `json:"size"`
}

type attachmentJSON struct {
	ContentType string `json:"contentType"`
	FileName    string `json:"fileName"`
	Pending     bool   `json:"pending"`
	attachmentFile
}

type Attachment struct {
	FileName    string
	ContentType string
	TimeSent    int64
	TimeRecv    int64
	Pending     bool
	attachmentFile
}

func (c *Context) parseAttachmentJSON(msg *Message, jatts []attachmentJSON) []Attachment {
	atts := make([]Attachment, 0, len(jatts))
	for _, jatt := range jatts {
		att := Attachment{
			FileName:       jatt.FileName,
			ContentType:    jatt.ContentType,
			TimeSent:       msg.TimeSent,
			TimeRecv:       msg.TimeRecv,
			Pending:        jatt.Pending,
			attachmentFile: jatt.attachmentFile,
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

func (c *Context) WriteAttachment(att *Attachment, w io.Writer) error {
	// XXX Don't read whole file at once
	data, err := c.readAttachment(att)
	if err != nil {
		return err
	}
	if _, err := w.Write(data); err != nil {
		return err
	}
	return nil
}

func (c *Context) readAttachment(att *Attachment) ([]byte, error) {
	if att.Pending {
		return nil, fmt.Errorf("attachment is pending")
	}
	return c.readAttachmentFile(&att.attachmentFile)
}

func (c *Context) readAttachmentFile(attf *attachmentFile) ([]byte, error) {
	if attf.Path == "" {
		return nil, fmt.Errorf("attachment without path")
	}
	path := c.attachmentFilePath(attf.Path)
	if attf.Version < 2 {
		return os.ReadFile(path)
	}
	return attf.decrypt(path)
}

func (a *attachmentFile) decrypt(path string) ([]byte, error) {
	keys, err := base64.StdEncoding.DecodeString(a.Keys)
	if err != nil {
		return nil, fmt.Errorf("cannot decode keys: %w", err)
	}
	if len(keys) != cipherKeySize+macKeySize {
		return nil, fmt.Errorf("invalid keys length")
	}
	cipherKey := keys[:cipherKeySize]
	macKey := keys[cipherKeySize:]

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(data) < ivSize+macSize {
		return nil, fmt.Errorf("attachment data too short")
	}

	iv := data[:ivSize]
	theirMAC := data[len(data)-macSize:]
	data = data[ivSize : len(data)-macSize]
	if len(data)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("invalid attachment data length")
	}

	m := hmac.New(sha256.New, macKey)
	m.Write(iv)
	m.Write(data)
	ourMAC := m.Sum(nil)
	if !hmac.Equal(ourMAC, theirMAC) {
		return nil, fmt.Errorf("MAC mismatch")
	}

	c, err := aes.NewCipher(cipherKey)
	if err != nil {
		return nil, err
	}
	cipher.NewCBCDecrypter(c, iv).CryptBlocks(data, data)

	if int64(len(data)) < a.Size {
		return nil, fmt.Errorf("invalid attachment data length")
	}
	data = data[:a.Size]

	return data, nil
}

func (c *Context) attachmentFilePath(path string) string {
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
