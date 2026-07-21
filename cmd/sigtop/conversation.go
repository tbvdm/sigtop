// Copyright (c) 2023 Tim van der Molen <tim@kariliq.nl>
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

package main

import (
	"errors"
	"regexp"
	"strings"

	"github.com/tbvdm/sigtop/signal"
)

func selectConversations(ctx *signal.Context, selectors []string) ([]signal.Conversation, error) {
	allConvs, err := ctx.Conversations()
	if err != nil {
		return nil, err
	}

	if selectors == nil {
		return allConvs, nil
	}

	var selConvs []signal.Conversation
	for _, s := range selectors {
		if len(s) == 0 || (len(s) == 1 && strings.IndexByte("+/=:", s[0]) >= 0) {
			return nil, errors.New("empty conversation selector")
		}
		var match func(*signal.Recipient) bool
		switch s[0] {
		case '+':
			match = func(r *signal.Recipient) bool {
				return r.Type == signal.RecipientTypeContact && s == r.Contact.Phone
			}
		case '/':
			re, err := regexp.Compile("(?i)" + s[1:])
			if err != nil {
				return nil, err
			}
			match = func(r *signal.Recipient) bool {
				return re.MatchString(r.DisplayName())
			}
		case ':':
			id := s[1:]
			match = func(r *signal.Recipient) bool {
				switch r.Type {
				case signal.RecipientTypeContact:
					return strings.EqualFold(id, r.Contact.ACI)
				case signal.RecipientTypeGroup:
					return strings.EqualFold(id, r.Group.ID)
				default:
					return false
				}
			}
		case '=':
			s = s[1:]
			fallthrough
		default:
			match = func(r *signal.Recipient) bool {
				return strings.EqualFold(s, r.DisplayName())
			}
		}

		tmp := allConvs[:0]
		for _, c := range allConvs {
			if match(c.Recipient) {
				selConvs = append(selConvs, c)
			} else {
				tmp = append(tmp, c)
			}
		}
		allConvs = tmp
	}

	return selConvs, nil
}
