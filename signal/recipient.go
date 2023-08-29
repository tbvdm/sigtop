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
	"strings"

	"github.com/tbvdm/sigtop/sqlcipher"
)

const (
	// For database version 19
	recipientQuery19 = "SELECT "                     +
		"id, "                                   +
		"type, "                                 +
		"name, "                                 +
		"profileName, "                          +
		"profileFamilyName, "                    +
		"profileFullName, "                      +
		"CASE type "                             +
			"WHEN 'private' THEN '+' || id " +
			"ELSE NULL "                     +
		"END, "                                  + // e164
		"NULL "                                  + // serviceId
		"FROM conversations"

	// For database versions [20, 87]
	recipientQuery20 = "SELECT "                     +
		"id, "                                   +
		"type, "                                 +
		"name, "                                 +
		"profileName, "                          +
		"profileFamilyName, "                    +
		"profileFullName, "                      +
		"e164, "                                 +
		"uuid "                                  + //serviceId
		"FROM conversations"

	// For database versions >= 88
	recipientQuery88 = "SELECT "                     +
		"id, "                                   +
		"type, "                                 +
		"name, "                                 +
		"profileName, "                          +
		"profileFamilyName, "                    +
		"profileFullName, "                      +
		"e164, "                                 +
		"serviceId "                             +
		"FROM conversations"
)

const (
	recipientColumnID = iota
	recipientColumnType
	recipientColumnName
	recipientColumnProfileName
	recipientColumnProfileFamilyName
	recipientColumnProfileFullName
	recipientColumnE164
	recipientColumnServiceID
)

type Recipient struct {
	Type    RecipientType
	Contact Contact
	Group   Group
}

type RecipientType int

const (
	RecipientTypeContact RecipientType = iota
	RecipientTypeGroup
)

type Contact struct {
	ACI               string // Account Identity
	Name              string
	ProfileName       string
	ProfileFamilyName string
	ProfileJoinedName string
	Phone             string
}

type Group struct {
	Name string
}

func (c *Context) makeRecipientMaps() error {
	if c.recipientsByConversationID != nil {
		// Nothing to do
		return nil
	}

	c.recipientsByConversationID = make(map[string]*Recipient)
	c.recipientsByPhone = make(map[string]*Recipient)
	c.recipientsByACI = make(map[string]*Recipient)

	var query string
	switch {
	case c.dbVersion < 20:
		query = recipientQuery19
	case c.dbVersion < 88:
		query = recipientQuery20
	default:
		query = recipientQuery88
	}

	stmt, err := c.db.Prepare(query)
	if err != nil {
		return err
	}

	for stmt.Step() {
		if err := c.addRecipient(stmt); err != nil {
			stmt.Finalize()
			return err
		}
	}

	return stmt.Finalize()
}

func (c *Context) addRecipient(stmt *sqlcipher.Stmt) error {
	var r *Recipient

	switch t := stmt.ColumnText(recipientColumnType); t {
	case "private":
		r = &Recipient{
			Type: RecipientTypeContact,
			Contact: Contact{
				Name:              trimBidiChars(stmt.ColumnText(recipientColumnName)),
				ProfileName:       stmt.ColumnText(recipientColumnProfileName),
				ProfileFamilyName: stmt.ColumnText(recipientColumnProfileFamilyName),
				ProfileJoinedName: stmt.ColumnText(recipientColumnProfileFullName),
				Phone:             stmt.ColumnText(recipientColumnE164),
				ACI:               stmt.ColumnText(recipientColumnServiceID),
			},
		}
	case "group":
		r = &Recipient{
			Type: RecipientTypeGroup,
			Group: Group{
				Name: stmt.ColumnText(recipientColumnName),
			},
		}
	default:
		return fmt.Errorf("unknown recipient type: %q", t)
	}

	id := stmt.ColumnText(recipientColumnID)
	c.recipientsByConversationID[id] = r

	if r.Type == RecipientTypeContact {
		if r.Contact.Phone != "" {
			c.recipientsByPhone[r.Contact.Phone] = r
		}
		if r.Contact.ACI != "" {
			c.recipientsByACI[strings.ToLower(r.Contact.ACI)] = r
		}
	}

	return nil
}

// trimBidiChars removes all leading and trailing FSI (U+2068) and PDI (U+2069)
// characters from the string s.
func trimBidiChars(s string) string {
	return strings.Trim(s, "\u2068\u2069")
}

func (c *Context) recipientFromConversationID(id string) (*Recipient, error) {
	if err := c.makeRecipientMaps(); err != nil {
		return nil, err
	}
	return c.recipientsByConversationID[id], nil
}

func (c *Context) recipientFromPhone(phone string) (*Recipient, error) {
	if err := c.makeRecipientMaps(); err != nil {
		return nil, err
	}
	return c.recipientsByPhone[phone], nil
}

func (c *Context) recipientFromACI(aci string) (*Recipient, error) {
	if err := c.makeRecipientMaps(); err != nil {
		return nil, err
	}
	return c.recipientsByACI[strings.ToLower(aci)], nil
}

func (r *Recipient) DisplayName() string {
	if r != nil {
		switch r.Type {
		case RecipientTypeContact:
			switch {
			case r.Contact.Name != "":
				return r.Contact.Name
			case r.Contact.ProfileJoinedName != "":
				return r.Contact.ProfileJoinedName
			case r.Contact.ProfileName != "":
				return r.Contact.ProfileName
			case r.Contact.Phone != "":
				return r.Contact.Phone
			case r.Contact.ACI != "":
				return r.Contact.ACI
			}
		case RecipientTypeGroup:
			switch {
			case r.Group.Name != "":
				return r.Group.Name
			}
		}
	}
	return "Unknown"
}
