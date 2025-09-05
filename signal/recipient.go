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
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tbvdm/sigtop/sqlcipher"
)

const (
	// For database version 19
	recipientQuery19 = "SELECT " +
		"id, " +
		"json, " +
		"type, " +
		"name, " +
		"profileName, " +
		"profileFamilyName, " +
		"profileFullName, " +
		"iif(type = 'private', '+' || id, NULL), " + // e164
		"NULL, " + // serviceId
		"iif(type = 'group', id, NULL) " + // groupId
		"FROM conversations"

	// For database versions [20, 87]
	recipientQuery20 = "SELECT " +
		"id, " +
		"json, " +
		"type, " +
		"name, " +
		"profileName, " +
		"profileFamilyName, " +
		"profileFullName, " +
		"e164, " +
		"uuid, " + // serviceId
		"groupId " +
		"FROM conversations"

	// For database versions >= 88
	recipientQuery88 = "SELECT " +
		"id, " +
		"json, " +
		"type, " +
		"name, " +
		"profileName, " +
		"profileFamilyName, " +
		"profileFullName, " +
		"e164, " +
		"serviceId, " +
		"groupId " +
		"FROM conversations"
)

const (
	recipientColumnID = iota
	recipientColumnJSON
	recipientColumnType
	recipientColumnName
	recipientColumnProfileName
	recipientColumnProfileFamilyName
	recipientColumnProfileFullName
	recipientColumnE164
	recipientColumnServiceID
	recipientColumnGroupID
)

// Based on ContactAvatarType in ts/types/Avatar.ts in the Signal-Desktop
// repository
type Avatar struct {
	attachmentFile
}

// Based on ConversationAttributesType in ts/model-types.d.ts in the
// Signal-Desktop repository
type recipientJSON struct {
	Username      string `json:"username"`
	ProfileAvatar Avatar `json:"profileAvatar"` // For contacts
	Avatar        Avatar `json:"avatar"`        // For groups
}

type Recipient struct {
	Type    RecipientType
	Contact Contact
	Group   Group
	Avatar  Avatar
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
	Username          string
}

type Group struct {
	ID   string
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
	case c.dbVersion >= 88:
		query = recipientQuery88
	case c.dbVersion >= 20:
		query = recipientQuery20
	default:
		query = recipientQuery19
	}

	stmt, _, err := c.db.Prepare(query)
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

	var jrpt recipientJSON
	if err := json.Unmarshal([]byte(stmt.ColumnText(recipientColumnJSON)), &jrpt); err != nil {
		return fmt.Errorf("cannot parse recipient JSON data: %w", err)
	}

	switch t := stmt.ColumnText(recipientColumnType); t {
	case "private":
		r = &Recipient{
			Type: RecipientTypeContact,
			Contact: Contact{
				ACI:               stmt.ColumnText(recipientColumnServiceID),
				Name:              trimBidiChars(stmt.ColumnText(recipientColumnName)),
				ProfileName:       stmt.ColumnText(recipientColumnProfileName),
				ProfileFamilyName: stmt.ColumnText(recipientColumnProfileFamilyName),
				ProfileJoinedName: stmt.ColumnText(recipientColumnProfileFullName),
				Phone:             stmt.ColumnText(recipientColumnE164),
				Username:          jrpt.Username,
			},
			Avatar: jrpt.ProfileAvatar,
		}
	case "group":
		r = &Recipient{
			Type: RecipientTypeGroup,
			Group: Group{
				ID:   stmt.ColumnText(recipientColumnGroupID),
				Name: stmt.ColumnText(recipientColumnName),
			},
			Avatar: jrpt.Avatar,
		}
	default:
		return fmt.Errorf("unknown recipient type: %q", t)
	}

	if r.Avatar.Path == SignalAvatarPath {
		// Ignore the avatar for the Signal release chat. It does not
		// exist in the Signal Desktop directory.
		r.Avatar.Path = ""
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

// trimBidiChars removes one surrounding pair of FSI (U+2068) and PDI (U+2069)
// characters from the string s.
func trimBidiChars(s string) string {
	const fsi, pdi = "\u2068", "\u2069"
	if strings.HasPrefix(s, fsi) && strings.HasSuffix(s[len(fsi):], pdi) {
		return s[len(fsi) : len(s)-len(pdi)]
	}
	return s
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

func (r *Recipient) displayNameAndDetail() (string, string) {
	name, detail := "Unknown", ""
	if r != nil {
		switch r.Type {
		case RecipientTypeContact:
			switch {
			case r.Contact.Name != "":
				name = r.Contact.Name
			case r.Contact.ProfileJoinedName != "":
				name = r.Contact.ProfileJoinedName
			case r.Contact.ProfileName != "":
				name = r.Contact.ProfileName
			case r.Contact.Phone != "":
				name = r.Contact.Phone
			case r.Contact.Username != "":
				name = r.Contact.Username
			case r.Contact.ACI != "":
				name = r.Contact.ACI
			}
			switch {
			case r.Contact.Phone != "":
				detail = r.Contact.Phone
			case r.Contact.Username != "":
				detail = r.Contact.Username
			case r.Contact.ACI != "":
				detail = r.Contact.ACI
			}
		case RecipientTypeGroup:
			switch {
			case r.Group.Name != "":
				name = r.Group.Name
			}
			// Newer group IDs are 32 bytes long and
			// base64-encoded, older ones are raw byte strings
			id, err := base64.StdEncoding.DecodeString(r.Group.ID)
			if err == nil && len(id) == 32 {
				// Convert to base64url without padding
				detail = strings.Map(func(r rune) rune {
					switch r {
					case '+':
						return '-'
					case '/':
						return '_'
					case '=':
						return -1
					default:
						return r
					}
				}, r.Group.ID)
			} else {
				detail = hex.EncodeToString([]byte(r.Group.ID))
			}
		}
	}
	return name, detail
}

func (r *Recipient) DisplayName() string {
	name, _ := r.displayNameAndDetail()
	return name
}

func (r *Recipient) DetailedDisplayName() string {
	name, detail := r.displayNameAndDetail()
	if detail == "" {
		return name
	}
	return name + " (" + detail + ")"
}

func (c *Context) ReadAvatar(avt *Avatar) ([]byte, error) {
	return c.readAttachmentFile(&avt.attachmentFile)
}
