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
	"fmt"
	"log"
	"sort"
	"strings"
	"unicode/utf8"
)

type mentionJSON struct {
	Start  int    `json:"start"`
	Length int    `json:"length"`
	// The "mentionUuid" field was renamed to "mentionAci" in database
	// version 88
	UUID   string `json:"mentionUuid"`
	ACI    string `json:"mentionAci"`
}

type Mention struct {
	Start     int
	Length    int
	Recipient *Recipient
}

func (c *Context) parseMentionJSON(jmnts []mentionJSON) ([]Mention, error) {
	var mnts []Mention

	for _, jmnt := range jmnts {
		var mnt Mention
		var err error

		switch {
		case jmnt.ACI != "":
			mnt.Recipient, err = c.recipientFromACI(jmnt.ACI)
			if err != nil {
				return nil, err
			}
			if mnt.Recipient == nil {
				log.Printf("cannot find mention recipient for ACI %q", jmnt.ACI)
			}
		case jmnt.UUID != "":
			mnt.Recipient, err = c.recipientFromACI(jmnt.UUID)
			if err != nil {
				return nil, err
			}
			if mnt.Recipient == nil {
				log.Printf("cannot find mention recipient for UUID %q", jmnt.UUID)
			}
		default:
			// XXX Ignore non-mentions for now
			continue
		}

		mnt.Start = jmnt.Start
		mnt.Length = jmnt.Length
		mnts = append(mnts, mnt)
	}

	return mnts, nil
}

type ErrMention struct {
	Msg   string
	Index int
	Body  *MessageBody
}

func (e *ErrMention) Error() string {
	var buf strings.Builder
	var runes = []rune(e.Body.Text)

	buf.WriteString(fmt.Sprintf("%s (index: %d, body: %d %d", e.Msg, e.Index, len(runes), len(e.Body.Text)))

	if !utf8.ValidString(e.Body.Text) {
		buf.WriteString(" invalid")
	}

	buf.WriteString(", placeholders:")
	for i, r := range runes {
		if r == '\ufffc' {
			buf.WriteString(fmt.Sprintf(" %d", i))
		}
	}

	buf.WriteString(", mentions:")
	for i, mnt := range e.Body.Mentions {
		buf.WriteString(fmt.Sprintf(" %d:%d,%d", i, mnt.Start, mnt.Length))
	}

	buf.WriteString(")")
	return buf.String()
}

func (b *MessageBody) insertMentions() error {
	if len(b.Mentions) == 0 {
		return nil
	}

	sort.Slice(b.Mentions, func(i, j int) bool { return b.Mentions[i].Start < b.Mentions[j].Start })

	var runes = []rune(b.Text)
	var text strings.Builder
	var off int

	for i := range b.Mentions {
		mnt := &b.Mentions[i]

		if mnt.Start < off || mnt.Length < 0 || mnt.Start+mnt.Length > len(runes) {
			return &ErrMention{Msg: "invalid mention", Index: i, Body: b}
		}

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
