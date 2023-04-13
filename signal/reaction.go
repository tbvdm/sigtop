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
	"log"
	"strings"
)

type reactionJSON struct {
	Emoji           string `json:"emoji"`
	FromID          string `json:"fromId"`
	TargetTimestamp int64  `json:"targetTimestamp"`
	Timestamp       int64  `json:"timestamp"`
}

type Reaction struct {
	Recipient *Recipient
	TimeSent  int64
	TimeRecv  int64
	Emoji     string
}

func (c *Context) parseReactionJSON(msg *Message, jmsg *messageJSON) error {
	for _, jrct := range jmsg.Reactions {
		rpt, err := c.recipientFromReactionID(jrct.FromID)
		if err != nil {
			return err
		}
		if rpt == nil {
			log.Printf("cannot find reaction recipient for ID %q", jrct.FromID)
		}
		rct := Reaction{
			Recipient: rpt,
			TimeSent:  jrct.TargetTimestamp,
			TimeRecv:  jrct.Timestamp,
			Emoji:     jrct.Emoji,
		}
		msg.Reactions = append(msg.Reactions, rct)
	}
	return nil
}

func (c *Context) recipientFromReactionID(id string) (*Recipient, error) {
	if c.dbVersion < 20 {
		if strings.HasPrefix(id, "+") {
			id = id[1:]
		}
		return c.recipientFromConversationID(id)
	} else {
		// Newer databases may still have reactions with older
		// recipient IDs
		if strings.HasPrefix(id, "+") {
			return c.recipientFromPhone(id)
		} else {
			return c.recipientFromConversationID(id)
		}
	}
}
