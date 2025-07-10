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
		"m.id, "                        +
		"m.conversationId, "            +
		"m.source, "                    +
		"m.type, "                      +
		"m.body, "                      +
		"m.json, "                      +
		"m.sent_at, "                   +
		"m.json ->> '$.received_at' "   +
		"FROM messages AS m "

	// For database versions [20, 22]
	messageSelect20 = "SELECT "             +
		"m.id, "                        +
		"m.conversationId, "            +
		"c.id, "                        +
		"m.type, "                      +
		"m.body, "                      +
		"m.json, "                      +
		"m.sent_at, "                   +
		"m.json ->> '$.received_at' "   +
		"FROM messages AS m "           +
		"LEFT JOIN conversations AS c " +
		"ON m.sourceUuid = c.uuid "

	// For database versions [23, 87]
	messageSelect23 = "SELECT "             +
		"m.id, "                        +
		"m.conversationId, "            +
		"c.id, "                        +
		"m.type, "                      +
		"m.body, "                      +
		"m.json, "                      +
		"m.sent_at, "                   +
		"coalesce(m.json ->> '$.received_at_ms', m.json ->> '$.received_at') " +
		"FROM messages AS m "           +
		"LEFT JOIN conversations AS c " +
		"ON m.sourceUuid = c.uuid "

	// For database versions [88, 1270)
	messageSelect88 = "SELECT "             +
		"m.id, "                        +
		"m.conversationId, "            +
		"c.id, "                        +
		"m.type, "                      +
		"m.body, "                      +
		"m.json, "                      +
		"m.sent_at, "                   +
		"coalesce(m.json ->> '$.received_at_ms', m.json ->> '$.received_at') " +
		"FROM messages AS m "           +
		"LEFT JOIN conversations AS c " +
		"ON m.sourceServiceId = c.serviceId "

	// For database versions >= 1270
	messageSelect1270 = "SELECT "           +
		"m.id, "                        +
		"m.conversationId, "            +
		"c.id, "                        +
		"m.type, "                      +
		"m.body, "                      +
		"m.json, "                      +
		"m.sent_at, "                   +
		"coalesce(m.received_at_ms, m.json ->> '$.received_at_ms', m.json ->> '$.received_at') " +
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
	messageQuery23 = messageSelect23 + messageWhereConversationID + messageOrder
	messageQuery88 = messageSelect88 + messageWhereConversationID + messageOrder
	messageQuery1270 = messageSelect1270 + messageWhereConversationID + messageOrder

	messageQuerySentBefore8  = messageSelect8  + messageWhereConversationIDAndSentBefore + messageOrder
	messageQuerySentBefore20 = messageSelect20 + messageWhereConversationIDAndSentBefore + messageOrder
	messageQuerySentBefore23 = messageSelect23 + messageWhereConversationIDAndSentBefore + messageOrder
	messageQuerySentBefore88 = messageSelect88 + messageWhereConversationIDAndSentBefore + messageOrder
	messageQuerySentBefore1270 = messageSelect1270 + messageWhereConversationIDAndSentBefore + messageOrder

	messageQuerySentAfter8  = messageSelect8  + messageWhereConversationIDAndSentAfter + messageOrder
	messageQuerySentAfter20 = messageSelect20 + messageWhereConversationIDAndSentAfter + messageOrder
	messageQuerySentAfter23 = messageSelect23 + messageWhereConversationIDAndSentAfter + messageOrder
	messageQuerySentAfter88 = messageSelect88 + messageWhereConversationIDAndSentAfter + messageOrder
	messageQuerySentAfter1270 = messageSelect1270 + messageWhereConversationIDAndSentAfter + messageOrder

	messageQuerySentBetween8  = messageSelect8  + messageWhereConversationIDAndSentBetween + messageOrder
	messageQuerySentBetween20 = messageSelect20 + messageWhereConversationIDAndSentBetween + messageOrder
	messageQuerySentBetween23 = messageSelect23 + messageWhereConversationIDAndSentBetween + messageOrder
	messageQuerySentBetween88 = messageSelect88 + messageWhereConversationIDAndSentBetween + messageOrder
	messageQuerySentBetween1270 = messageSelect1270 + messageWhereConversationIDAndSentBetween + messageOrder
)

const (
	messageColumnID = iota
	messageColumnConversationID
	messageColumnSourceID
	messageColumnType
	messageColumnBody
	messageColumnJSON
	messageColumnSentAt
	messageColumnReceivedAtMS
)

type messageJSON struct {
	Attachments  []attachmentJSON `json:"attachments"`
	Mentions     []mentionJSON    `json:"bodyRanges"`
	Reactions    []reactionJSON   `json:"reactions"`
	Quote        *quoteJSON       `json:"quote"`
	Edits        []editJSON       `json:"editHistory"`
}

type Message struct {
	ID           string
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
	Edits        []Edit
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
	case c.dbVersion >= 1270:
		query = messageQuery1270
	case c.dbVersion >= 88:
		query = messageQuery88
	case c.dbVersion >= 23:
		query = messageQuery23
	case c.dbVersion >= 20:
		query = messageQuery20
	default:
		query = messageQuery8
	}

	stmt, _, err := c.db.Prepare(query)
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
	case c.dbVersion >= 1270:
		query = messageQuerySentBefore1270
	case c.dbVersion >= 88:
		query = messageQuerySentBefore88
	case c.dbVersion >= 23:
		query = messageQuerySentBefore23
	case c.dbVersion >= 20:
		query = messageQuerySentBefore20
	default:
		query = messageQuerySentBefore8
	}

	stmt, _, err := c.db.Prepare(query)
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
	case c.dbVersion >= 1270:
		query = messageQuerySentAfter1270
	case c.dbVersion >= 88:
		query = messageQuerySentAfter88
	case c.dbVersion >= 23:
		query = messageQuerySentAfter23
	case c.dbVersion >= 20:
		query = messageQuerySentAfter20
	default:
		query = messageQuerySentAfter8
	}

	stmt, _, err := c.db.Prepare(query)
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
	case c.dbVersion >= 1270:
		query = messageQuerySentBetween1270
	case c.dbVersion >= 88:
		query = messageQuerySentBetween88
	case c.dbVersion >= 23:
		query = messageQuerySentBetween23
	case c.dbVersion >= 20:
		query = messageQuerySentBetween20
	default:
		query = messageQuerySentBetween8
	}

	stmt, _, err := c.db.Prepare(query)
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

		if stmt.ColumnType(messageColumnSourceID) != sqlcipher.ColumnTypeNull {
			id := stmt.ColumnText(messageColumnSourceID)
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

		msg.ID = stmt.ColumnText(messageColumnID)
		msg.Type = stmt.ColumnText(messageColumnType)
		msg.Body.Text = stmt.ColumnText(messageColumnBody)
		msg.JSON = stmt.ColumnText(messageColumnJSON)
		msg.TimeSent = stmt.ColumnInt64(messageColumnSentAt)
		msg.TimeRecv = stmt.ColumnInt64(messageColumnReceivedAtMS)

		jmsg, err := c.parseMessageJSON(&msg)
		if err != nil {
			stmt.Finalize()
			return nil, err
		}

		msg.Attachments, err = c.attachmentsForMessage(&msg, jmsg.Attachments)
		if err != nil {
			stmt.Finalize()
			return nil, err
		}

		if err := msg.Body.insertMentions(); err != nil {
			msg.logError(err, "message with invalid mention")
			msg.Body.Mentions = nil
		}

		if msg.Quote != nil {
			if err := msg.Quote.Body.insertMentions(); err != nil {
				msg.logError(err, "message with invalid mention in quote")
				msg.Quote.Body.Mentions = nil
			}
		}

		for i := range msg.Edits {
			if err := msg.Edits[i].Body.insertMentions(); err != nil {
				msg.logError(err, "message with invalid mention in edit %d", i)
				msg.Edits[i].Body.Mentions = nil
			}
			if msg.Edits[i].Quote != nil {
				if err := msg.Edits[i].Quote.Body.insertMentions(); err != nil {
					msg.logError(err, "message with invalid mention in quote in edit %d", i)
					msg.Edits[i].Quote.Body.Mentions = nil
				}
			}
		}

		msgs = append(msgs, msg)
	}

	return msgs, stmt.Finalize()
}

func (c *Context) parseMessageJSON(msg *Message) (messageJSON, error) {
	var jmsg messageJSON
	var err error
	if err = json.Unmarshal([]byte(msg.JSON), &jmsg); err != nil {
		return jmsg, fmt.Errorf("cannot parse message JSON data: %w", err)
	}
	if msg.Body.Mentions, err = c.parseMentionJSON(jmsg.Mentions); err != nil {
		return jmsg, err
	}
	if msg.Quote, err = c.parseQuoteJSON(jmsg.Quote); err != nil {
		return jmsg, err
	}
	if err = c.parseReactionJSON(msg, &jmsg); err != nil {
		return jmsg, err
	}
	if err = c.parseEditJSON(msg, &jmsg); err != nil {
		return jmsg, err
	}
	return jmsg, nil
}

func (m *Message) logError(err error, format string, a ...any) {
	if err != nil {
		log.Print(err)
	}
	s := fmt.Sprintf(format, a...)
	s += fmt.Sprintf(" (conversation: %q, sent: %s (%d))", m.Conversation.DisplayName(), time.UnixMilli(m.TimeSent).Format("2006-01-02 15:04:05"), m.TimeSent)
	log.Print(s)
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
