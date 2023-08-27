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
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/tbvdm/sigtop/sqlcipher"
)

const (
	// For database versions [8, 19]
	messageSelect8 = "SELECT "              +
		"m.conversationId, "            +
		"m.source, "                    +
		"m.type, "                      +
		"m.body, "                      +
		"m.json, "                      +
		"m.sent_at "                    +
		"FROM messages AS m "

	// For database versions [20, 87]
	messageSelect20 = "SELECT "             +
		"m.conversationId, "            +
		"c.id, "                        +
		"m.type, "                      +
		"m.body, "                      +
		"m.json, "                      +
		"m.sent_at "                    +
		"FROM messages AS m "           +
		"LEFT JOIN conversations AS c " +
		"ON m.sourceUuid = c.uuid "

	// For database versions >= 88
	messageSelect88 = "SELECT "             +
		"m.conversationId, "            +
		"c.id, "                        +
		"m.type, "                      +
		"m.body, "                      +
		"m.json, "                      +
		"m.sent_at "                    +
		"FROM messages AS m "           +
		"LEFT JOIN conversations AS c " +
		"ON m.sourceServiceId = c.serviceId "

	messageWhereConversationID               = "WHERE m.conversationId = ? "
	messageWhereConversationIDAndSentBefore  = messageWhereConversationID + "AND (m.sent_at <= ? OR m.sent_at IS NULL) "
	messageWhereConversationIDAndSentAfter   = messageWhereConversationID + "AND m.sent_at >= ? "
	messageWhereConversationIDAndSentBetween = messageWhereConversationID + "AND m.sent_at BETWEEN ? AND ? "
	messageOrder                             = "ORDER BY m.received_at, m.sent_at"

	messageQuery8  = messageSelect8  + messageWhereConversationID + messageOrder
	messageQuery20 = messageSelect20 + messageWhereConversationID + messageOrder
	messageQuery88 = messageSelect88 + messageWhereConversationID + messageOrder

	messageQuerySentBefore8  = messageSelect8  + messageWhereConversationIDAndSentBefore + messageOrder
	messageQuerySentBefore20 = messageSelect20 + messageWhereConversationIDAndSentBefore + messageOrder
	messageQuerySentBefore88 = messageSelect88 + messageWhereConversationIDAndSentBefore + messageOrder

	messageQuerySentAfter8  = messageSelect8  + messageWhereConversationIDAndSentAfter + messageOrder
	messageQuerySentAfter20 = messageSelect20 + messageWhereConversationIDAndSentAfter + messageOrder
	messageQuerySentAfter88 = messageSelect88 + messageWhereConversationIDAndSentAfter + messageOrder

	messageQuerySentBetween8  = messageSelect8  + messageWhereConversationIDAndSentBetween + messageOrder
	messageQuerySentBetween20 = messageSelect20 + messageWhereConversationIDAndSentBetween + messageOrder
	messageQuerySentBetween88 = messageSelect88 + messageWhereConversationIDAndSentBetween + messageOrder
)

const (
	messageColumnConversationID = iota
	messageColumnID
	messageColumnType
	messageColumnBody
	messageColumnJSON
	messageColumnSentAt
)

type messageJSON struct {
	Attachments  []attachmentJSON `json:"attachments"`
	ReceivedAt   int64            `json:"received_at"`
	ReceivedAtMS int64            `json:"received_at_ms"`
	Mentions     []mentionJSON    `json:"bodyRanges"`
	Reactions    []reactionJSON   `json:"reactions"`
	Quote        *quoteJSON       `json:"quote"`
}

type Message struct {
	Conversation *Recipient
	Source       *Recipient
	TimeSent     int64
	TimeRecv     int64
	Type         string
	Body         MessageBody
	JSON         string
	Attachments  []Attachment
	Reactions    []Reaction
	Quote        *Quote
}

type MessageBody struct {
	Text     string
	Mentions []Mention
}

type Interval struct {
	Min time.Time
	Max time.Time
}

func (c *Context) ConversationMessages(conv *Conversation, ival Interval) ([]Message, error) {
	switch {
	case ival.Min.IsZero() && ival.Max.IsZero():
		return c.allConversationMessages(conv)
	case ival.Min.IsZero():
		return c.conversationMessagesSentBefore(conv, ival.Max)
	case ival.Max.IsZero():
		return c.conversationMessagesSentAfter(conv, ival.Min)
	default:
		return c.conversationMessagesSentBetween(conv, ival.Min, ival.Max)
	}
}

func (c *Context) allConversationMessages(conv *Conversation) ([]Message, error) {
	var query string
	switch {
	case c.dbVersion < 20:
		query = messageQuery8
	case c.dbVersion < 88:
		query = messageQuery20
	default:
		query = messageQuery88
	}

	stmt, err := c.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	if err := stmt.BindText(1, conv.ID); err != nil {
		stmt.Finalize()
		return nil, err
	}

	return c.messages(stmt)
}

func (c *Context) conversationMessagesSentBefore(conv *Conversation, max time.Time) ([]Message, error) {
	var query string
	switch {
	case c.dbVersion < 20:
		query = messageQuerySentBefore8
	case c.dbVersion < 88:
		query = messageQuerySentBefore20
	default:
		query = messageQuerySentBefore88
	}

	stmt, err := c.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	if err := stmt.BindText(1, conv.ID); err != nil {
		stmt.Finalize()
		return nil, err
	}
	if err := stmt.BindInt64(2, max.UnixMilli()); err != nil {
		stmt.Finalize()
		return nil, err
	}

	return c.messages(stmt)
}

func (c *Context) conversationMessagesSentAfter(conv *Conversation, min time.Time) ([]Message, error) {
	var query string
	switch {
	case c.dbVersion < 20:
		query = messageQuerySentAfter8
	case c.dbVersion < 88:
		query = messageQuerySentAfter20
	default:
		query = messageQuerySentAfter88
	}

	stmt, err := c.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	if err := stmt.BindText(1, conv.ID); err != nil {
		stmt.Finalize()
		return nil, err
	}
	if err := stmt.BindInt64(2, min.UnixMilli()); err != nil {
		stmt.Finalize()
		return nil, err
	}

	return c.messages(stmt)
}

func (c *Context) conversationMessagesSentBetween(conv *Conversation, min, max time.Time) ([]Message, error) {
	var query string
	switch {
	case c.dbVersion < 20:
		query = messageQuerySentBetween8
	case c.dbVersion < 88:
		query = messageQuerySentBetween20
	default:
		query = messageQuerySentBetween88
	}

	stmt, err := c.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	if err := stmt.BindText(1, conv.ID); err != nil {
		stmt.Finalize()
		return nil, err
	}
	if err := stmt.BindInt64(2, min.UnixMilli()); err != nil {
		stmt.Finalize()
		return nil, err
	}
	if err := stmt.BindInt64(3, max.UnixMilli()); err != nil {
		stmt.Finalize()
		return nil, err
	}

	return c.messages(stmt)
}

func (c *Context) messages(stmt *sqlcipher.Stmt) ([]Message, error) {
	var msgs []Message
	for stmt.Step() {
		var msg Message

		if stmt.ColumnType(messageColumnConversationID) == sqlcipher.ColumnTypeNull {
			// Likely message with error
			log.Printf("conversation recipient has null ID")
		} else {
			id := stmt.ColumnText(messageColumnConversationID)
			rpt, err := c.recipientFromConversationID(id)
			if err != nil {
				stmt.Finalize()
				return nil, err
			}
			if rpt == nil {
				log.Printf("cannot find conversation recipient for ID %q", id)
			}
			msg.Conversation = rpt
		}

		if stmt.ColumnType(messageColumnID) != sqlcipher.ColumnTypeNull {
			id := stmt.ColumnText(messageColumnID)
			rpt, err := c.recipientFromConversationID(id)
			if err != nil {
				stmt.Finalize()
				return nil, err
			}
			if rpt == nil {
				log.Printf("cannot find source recipient for ID %q", id)
			}
			msg.Source = rpt
		}

		msg.Type = stmt.ColumnText(messageColumnType)
		msg.Body.Text = stmt.ColumnText(messageColumnBody)
		msg.JSON = stmt.ColumnText(messageColumnJSON)
		msg.TimeSent = stmt.ColumnInt64(messageColumnSentAt)

		if err := c.parseMessageJSON(&msg); err != nil {
			stmt.Finalize()
			return nil, err
		}

		if err := msg.Body.insertMentions(); err != nil {
			log.Print(err)
			log.Printf("message with invalid mention (conversation: %q, sent: %s)", msg.Conversation.DisplayName(), time.UnixMilli(msg.TimeSent).Format("2006-01-02 15:04:05"))
			msg.Body.Mentions = nil
		}

		if msg.Quote != nil {
			if err := msg.Quote.Body.insertMentions(); err != nil {
				log.Print(err)
				log.Printf("message with invalid mention in quote (conversation: %q, sent: %s)", msg.Conversation.DisplayName(), time.UnixMilli(msg.TimeSent).Format("2006-01-02 15:04:05"))
				msg.Quote.Body.Mentions = nil
			}
		}

		msgs = append(msgs, msg)
	}

	return msgs, stmt.Finalize()
}

func (c *Context) parseMessageJSON(msg *Message) error {
	var jmsg messageJSON
	if err := json.Unmarshal([]byte(msg.JSON), &jmsg); err != nil {
		return fmt.Errorf("cannot parse JSON data: %w", err)
	}
	// For older messages, the received time is stored in the "received_at"
	// attribute. For newer messages, it is in the new "received_at_ms"
	// attribute (and the "received_at" attribute was changed to store a
	// counter). See Signal-Desktop commit
	// d82ce079421c3fa08a0920a90b7abc19b1bb0e59.
	if jmsg.ReceivedAtMS != 0 {
		msg.TimeRecv = jmsg.ReceivedAtMS
	} else {
		msg.TimeRecv = jmsg.ReceivedAt
	}
	if err := c.parseAttachmentJSON(msg, &jmsg); err != nil {
		return err
	}
	if err := c.parseMentionJSON(&msg.Body, jmsg.Mentions); err != nil {
		return err
	}
	if err := c.parseReactionJSON(msg, &jmsg); err != nil {
		return err
	}
	if err := c.parseQuoteJSON(msg, &jmsg); err != nil {
		return err
	}
	return nil
}

func (m *Message) dumpJSON() {
	var buf bytes.Buffer
	if err := json.Indent(&buf, []byte(m.JSON), "", "  "); err != nil {
		log.Printf("cannot dump message JSON data: %v", err)
		return
	}
	fmt.Fprintln(log.Writer(), buf.String())
}

func (m *Message) IsOutgoing() bool {
	return m.Type == "outgoing"
}
