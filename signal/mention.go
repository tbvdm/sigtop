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
	"errors"
	"strings"
)

type mentionJSON struct {
	Start  int    `json:"Start"`
	Length int    `json:"Length"`
	UUID   string `json:"mentionUuid"`
}

type Mention struct {
	Start     int
	Length    int
	Recipient *Recipient
}

func (c *Context) parseMentionJSON(body *MessageBody, jmnts []mentionJSON) error {
	for _, jmnt := range jmnts {
		rpt, err := c.recipientFromUUID(jmnt.UUID)
		if err != nil {
			return err
		}
		if rpt == nil {
			warn("cannot find mention recipient for UUID %q", jmnt.UUID)
		}

		mnt := Mention{
			Start:     jmnt.Start,
			Length:    jmnt.Length,
			Recipient: rpt,
		}

		// Insert the mention in order. It seems that in most cases the
		// mentions in the JSON data are already ordered. So, as an
		// optimisation, traverse the slice in reverse direction.
		var i int
		for i = len(body.Mentions); i > 0; i-- {
			if body.Mentions[i-1].Start < mnt.Start {
				break
			}
		}
		if i == len(body.Mentions) {
			body.Mentions = append(body.Mentions, mnt)
		} else {
			body.Mentions = append(body.Mentions[:i+1], body.Mentions[i:]...)
			body.Mentions[i] = mnt
		}
	}

	return nil
}

func (b *MessageBody) insertMentions() error {
	// Ensure the mentions are ordered and don't overlap
	for i := 1; i < len(b.Mentions); i++ {
		end := b.Mentions[i-1].Start + b.Mentions[i-1].Length
		if b.Mentions[i].Start < end {
			return errors.New("unordered or overlapping mentions")
		}
	}

	var runes = []rune(b.Text)
	var text strings.Builder
	var off int

	for _, mnt := range b.Mentions {
		// Copy text preceding mention
		text.WriteString(string(runes[off:mnt.Start]))
		off = mnt.Start + mnt.Length

		repl := "@" + mnt.Recipient.DisplayName()

		// Update mention. Note: the original start and length values
		// were character counts, but the updated values are byte
		// counts.
		mnt.Start = text.Len()
		mnt.Length = len(repl)

		text.WriteString(repl)
	}

	// Copy text succeeding last mention
	text.WriteString(string(runes[off:]))
	b.Text = text.String()

	return nil
}
