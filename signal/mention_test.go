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

package signal

import "testing"

func TestUpdatedBody(t *testing.T) {
	part, foo, bar := "aàạ𝔞", "Fộo", "Bậr"
	body := MessageBody{
		Text: part + "\ufffc" + part + "\ufffc" + part,
		Mentions: []Mention{
			{4, 1, contact(foo)},
			{9, 1, contact(bar)},
		},
	}

	if err := body.insertMentions(); err != nil {
		t.Fatal(err)
	}

	testText(t, &body, part+"@"+foo+part+"@"+bar+part)
	testMention(t, &body, 0, 10, 6, foo)
	testMention(t, &body, 1, 26, 6, bar)
}

func TestUnsortedMentions(t *testing.T) {
	body := MessageBody{
		Text: "\ufffc\ufffc",
		Mentions: []Mention{
			{1, 1, contact("Bar")},
			{0, 1, contact("Foo")},
		},
	}

	if err := body.insertMentions(); err != nil {
		t.Fatal(err)
	}

	testText(t, &body, "@Foo@Bar")
	testMention(t, &body, 0, 0, 4, "Foo")
	testMention(t, &body, 1, 4, 4, "Bar")
}

func TestZeroMentionLength(t *testing.T) {
	part, foo := "aàạ𝔞", "Fộo"
	body := MessageBody{
		Text: part + part,
		Mentions: []Mention{
			{4, 0, contact(foo)},
		},
	}

	if err := body.insertMentions(); err != nil {
		t.Fatal(err)
	}

	testText(t, &body, part+"@"+foo+part)
	testMention(t, &body, 0, 10, 6, foo)
}

func TestLongMentionLength(t *testing.T) {
	part, foo := "aàạ𝔞", "Fộo"
	body := MessageBody{
		Text: part + part + part,
		Mentions: []Mention{
			{4, 4, contact(foo)},
		},
	}

	if err := body.insertMentions(); err != nil {
		t.Fatal(err)
	}

	testText(t, &body, part+"@"+foo+part)
	testMention(t, &body, 0, 10, 6, foo)
}

func TestOverlappingMentions1(t *testing.T) {
	body := MessageBody{
		Text: "\ufffc\ufffc",
		Mentions: []Mention{
			{0, 1, nil},
			{0, 1, nil},
		},
	}

	if body.insertMentions() == nil {
		t.Fatal("no error for overlapping mentions")
	}
}

func TestOverlappingMentions2(t *testing.T) {
	body := MessageBody{
		Text: "\ufffc\ufffc\ufffc",
		Mentions: []Mention{
			{0, 2, nil},
			{1, 1, nil},
		},
	}

	if body.insertMentions() == nil {
		t.Fatal("no error for overlapping mentions")
	}
}

func TestNegativeMentionStart(t *testing.T) {
	body := MessageBody{
		Text: "a",
		Mentions: []Mention{
			{-1, 1, nil},
		},
	}

	if body.insertMentions() == nil {
		t.Fatal("no error for negative mention start")
	}
}

func TestNegativeMentionLength(t *testing.T) {
	body := MessageBody{
		Text: "a",
		Mentions: []Mention{
			{0, -1, nil},
		},
	}

	if body.insertMentions() == nil {
		t.Fatal("no error for negative mention length")
	}
}

func TestOutOfBoundsMention1(t *testing.T) {
	body := MessageBody{
		Text: "a",
		Mentions: []Mention{
			{1, 1, nil},
		},
	}

	if body.insertMentions() == nil {
		t.Fatal("no error for out-of-bounds mention")
	}
}

func TestOutOfBoundsMention2(t *testing.T) {
	body := MessageBody{
		Text: "a",
		Mentions: []Mention{
			{0, 2, nil},
		},
	}

	if body.insertMentions() == nil {
		t.Fatal("no error for out-of-bounds mention")
	}
}

func contact(name string) *Recipient {
	return &Recipient{
		Type:    RecipientTypeContact,
		Contact: Contact{Name: name},
	}
}

func testText(t *testing.T, body *MessageBody, text string) {
	if body.Text != text {
		t.Fatalf("body text: want %q, have %q", text, body.Text)
	}
}

func testMention(t *testing.T, body *MessageBody, idx, start, length int, name string) {
	if body.Mentions[idx].Start != start {
		t.Fatalf("start of mention %d: want %d, have %d", idx, start, body.Mentions[idx].Start)
	}
	if body.Mentions[idx].Length != length {
		t.Fatalf("length of mention %d: want %d, have %d", idx, length, body.Mentions[idx].Length)
	}
	if body.Mentions[idx].Recipient.Contact.Name != name {
		t.Fatalf("contact name of mention %d: want %q, have %q", idx, name, body.Mentions[idx].Recipient.Contact.Name)
	}
}
