/*
 * Copyright (c) 2018 Tim van der Molen <tim@kariliq.nl>
 *
 * Permission to use, copy, modify, and distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

#ifndef SBK_H
#define SBK_H

#include <sys/queue.h>
#include <sys/types.h>

#include <stdint.h>

#ifndef nitems
#define nitems(a) (sizeof (a) / sizeof (a)[0])
#endif

struct sbk_ctx;

struct sbk_contact {
	char		*uuid;
	char		*name;
	char		*profile_name;
	char		*profile_family_name;
	char		*profile_joined_name;
	char		*phone;
};

struct sbk_group {
	char		*name;
};

struct sbk_recipient {
	enum {
		SBK_CONTACT,
		SBK_GROUP
	} type;
	struct sbk_contact	*contact;
	struct sbk_group	*group;
};

struct sbk_conversation {
	char		*id;
	struct sbk_recipient *recipient;
	SIMPLEQ_ENTRY(sbk_conversation) entries;
};

SIMPLEQ_HEAD(sbk_conversation_list, sbk_conversation);

struct sbk_attachment {
	char		*path;
	char		*filename;
	char		*content_type;
	uint64_t	 size;
	uint64_t	 time_sent;
	uint64_t	 time_recv;
	TAILQ_ENTRY(sbk_attachment) entries;
};

TAILQ_HEAD(sbk_attachment_list, sbk_attachment);

struct sbk_mention {
	struct sbk_recipient *recipient;
	uint64_t	 start;
	uint64_t	 length;
	TAILQ_ENTRY(sbk_mention) entries;
};

TAILQ_HEAD(sbk_mention_list, sbk_mention);

struct sbk_reaction {
	struct sbk_recipient *recipient;
	uint64_t	 time_sent;
	uint64_t	 time_recv;
	char		*emoji;
	SIMPLEQ_ENTRY(sbk_reaction) entries;
};

SIMPLEQ_HEAD(sbk_reaction_list, sbk_reaction);

struct sbk_quote {
	uint64_t	 id;
	struct sbk_recipient *recipient;
	char		*text;
	struct sbk_attachment_list *attachments;
	struct sbk_mention_list *mentions;
};

struct sbk_message {
	struct sbk_recipient *conversation;
	struct sbk_recipient *source;
	uint64_t	 time_sent;
	uint64_t	 time_recv;
	char		*type;
	char		*text;
	char		*json;
	struct sbk_attachment_list *attachments;
	struct sbk_mention_list *mentions;
	struct sbk_reaction_list *reactions;
	struct sbk_quote *quote;
	SIMPLEQ_ENTRY(sbk_message) entries;
};

SIMPLEQ_HEAD(sbk_message_list, sbk_message);

int		 sbk_open(struct sbk_ctx **, const char *);
void		 sbk_close(struct sbk_ctx *);
const char	*sbk_error(struct sbk_ctx *);

int		 sbk_check_database(struct sbk_ctx *, char ***);
int		 sbk_write_database(struct sbk_ctx *, const char *);

struct sbk_conversation_list *sbk_get_conversations(struct sbk_ctx *);
void		 sbk_free_conversation_list(struct sbk_conversation_list *);

struct sbk_attachment_list *sbk_get_attachments(struct sbk_ctx *,
		    struct sbk_conversation *);
struct sbk_attachment_list *sbk_get_attachments_sent_after(struct sbk_ctx *,
		    struct sbk_conversation *, time_t);
struct sbk_attachment_list *sbk_get_attachments_sent_before(struct sbk_ctx *,
		    struct sbk_conversation *, time_t);
struct sbk_attachment_list *sbk_get_attachments_sent_between(struct sbk_ctx *,
		    struct sbk_conversation *, time_t, time_t);
void		 sbk_free_attachment_list(struct sbk_attachment_list *);
int		 sbk_get_attachment_path(struct sbk_ctx *, char **,
		    struct sbk_attachment *);

struct sbk_message_list *sbk_get_messages(struct sbk_ctx *,
		    struct sbk_conversation *);
struct sbk_message_list *sbk_get_messages_sent_after(struct sbk_ctx *,
		    struct sbk_conversation *, time_t);
struct sbk_message_list *sbk_get_messages_sent_before(struct sbk_ctx *,
		    struct sbk_conversation *, time_t);
struct sbk_message_list *sbk_get_messages_sent_between(struct sbk_ctx *,
		    struct sbk_conversation *, time_t, time_t);
void		 sbk_free_message_list(struct sbk_message_list *);
int		 sbk_is_outgoing_message(const struct sbk_message *);

const char	*sbk_get_recipient_display_name(const struct sbk_recipient *);

#endif
