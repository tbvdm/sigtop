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

#include "config.h"

#include <sys/types.h>

#include <stdlib.h>
#include <string.h>

#include "sbk-internal.h"

/* For database versions 8 to 19 */
#define SBK_SELECT_8							\
	"SELECT "							\
	"m.conversationId, "						\
	"m.source, "							\
	"m.type, "							\
	"m.body, "							\
	"m.json, "							\
	"m.sent_at "							\
	"FROM messages AS m "

/* For database versions >= 20 */
#define SBK_SELECT_20							\
	"SELECT "							\
	"m.conversationId, "						\
	"c.id, "							\
	"m.type, "							\
	"m.body, "							\
	"m.json, "							\
	"m.sent_at "							\
	"FROM messages AS m "						\
	"LEFT JOIN conversations AS c "					\
	"ON m.sourceUuid = c.uuid "

#define SBK_WHERE_CONVERSATIONID					\
	"WHERE m.conversationId = ? "

#define SBK_WHERE_CONVERSATIONID_SENT_AFTER				\
	SBK_WHERE_CONVERSATIONID					\
	"AND m.sent_at >= ? "

#define SBK_WHERE_CONVERSATIONID_SENT_BEFORE				\
	SBK_WHERE_CONVERSATIONID					\
	"AND m.sent_at <= ? "

#define SBK_WHERE_CONVERSATIONID_SENT_BETWEEN				\
	SBK_WHERE_CONVERSATIONID					\
	"AND m.sent_at BETWEEN ? AND ? "

#define SBK_ORDER							\
	"ORDER BY m.received_at"

#define SBK_QUERY_8							\
	SBK_SELECT_8							\
	SBK_WHERE_CONVERSATIONID					\
	SBK_ORDER

#define SBK_QUERY_20							\
	SBK_SELECT_20							\
	SBK_WHERE_CONVERSATIONID					\
	SBK_ORDER

#define SBK_QUERY_SENT_AFTER_8						\
	SBK_SELECT_8							\
	SBK_WHERE_CONVERSATIONID_SENT_AFTER				\
	SBK_ORDER

#define SBK_QUERY_SENT_AFTER_20						\
	SBK_SELECT_20							\
	SBK_WHERE_CONVERSATIONID_SENT_AFTER				\
	SBK_ORDER

#define SBK_QUERY_SENT_BEFORE_8						\
	SBK_SELECT_8							\
	SBK_WHERE_CONVERSATIONID_SENT_BEFORE				\
	SBK_ORDER

#define SBK_QUERY_SENT_BEFORE_20					\
	SBK_SELECT_20							\
	SBK_WHERE_CONVERSATIONID_SENT_BEFORE				\
	SBK_ORDER

#define SBK_QUERY_SENT_BETWEEN_8					\
	SBK_SELECT_8							\
	SBK_WHERE_CONVERSATIONID_SENT_BETWEEN				\
	SBK_ORDER

#define SBK_QUERY_SENT_BETWEEN_20					\
	SBK_SELECT_20							\
	SBK_WHERE_CONVERSATIONID_SENT_BETWEEN				\
	SBK_ORDER

#define SBK_COLUMN_CONVERSATIONID	0
#define SBK_COLUMN_ID			1
#define SBK_COLUMN_TYPE			2
#define SBK_COLUMN_BODY			3
#define SBK_COLUMN_JSON			4
#define SBK_COLUMN_SENT_AT		5

static void
sbk_free_message(struct sbk_message *msg)
{
	if (msg != NULL) {
		free(msg->type);
		free(msg->text);
		free(msg->json);
		sbk_free_attachment_list(msg->attachments);
		sbk_free_mention_list(msg->mentions);
		sbk_free_reaction_list(msg->reactions);
		sbk_free_quote(msg->quote);
		free(msg);
	}
}

void
sbk_free_message_list(struct sbk_message_list *lst)
{
	struct sbk_message *msg;

	if (lst != NULL) {
		while ((msg = SIMPLEQ_FIRST(lst)) != NULL) {
			SIMPLEQ_REMOVE_HEAD(lst, entries);
			sbk_free_message(msg);
		}
		free(lst);
	}
}

static int
sbk_parse_message_json(struct sbk_ctx *ctx, struct sbk_message *msg)
{
	jsmntok_t	tokens[2048];
	int		idx;

	if (msg->json == NULL)
		return 0;

	if (sbk_jsmn_parse(msg->json, strlen(msg->json), tokens,
	    nitems(tokens)) == -1) {
		warnx("Cannot parse message JSON data");
		return -1;
	}

	if (tokens[0].type != JSMN_OBJECT) {
		warnx("Unexpected message JSON type");
		return -1;
	}

	/*
	 * Get received time
	 *
	 * For older messages, the received time is stored in the "received_at"
	 * attribute. For newer messages, it is in the "received_at_ms"
	 * attribute (and the "received_at" attribute was changed to store a
	 * counter). See Signal-Desktop commit
	 * d82ce079421c3fa08a0920a90b7abc19b1bb0e59.
	 */

	idx = sbk_jsmn_get_number(msg->json, tokens, "received_at_ms");
	if (idx == -1)
		idx = sbk_jsmn_get_number(msg->json, tokens, "received_at");
	if (idx != -1) {
		if (sbk_jsmn_parse_uint64(&msg->time_recv, msg->json,
		    &tokens[idx]) == -1) {
			warnx("Cannot parse message received time");
			return -1;
		}
	}

	/*
	 * Get attachments
	 */

	idx = sbk_jsmn_get_array(msg->json, tokens, "attachments");
	if (idx != -1 &&
	    sbk_parse_attachment_json(msg, &tokens[idx]) == -1)
		return -1;

	/*
	 * Get mentions
	 */

	idx = sbk_jsmn_get_array(msg->json, tokens, "bodyRanges");
	if (idx != -1 &&
	    sbk_parse_mention_json(ctx, msg, &msg->mentions, &tokens[idx]) ==
	    -1)
		return -1;

	/*
	 * Get reactions
	 */

	idx = sbk_jsmn_get_array(msg->json, tokens, "reactions");
	if (idx != -1 &&
	    sbk_parse_reaction_json(ctx, msg, &tokens[idx]) == -1)
		return -1;

	/*
	 * Get quote
	 */

	idx = sbk_jsmn_get_object(msg->json, tokens, "quote");
	if (idx != -1 &&
	    sbk_parse_quote_json(ctx, msg, &tokens[idx]) == -1)
		return -1;

	return 0;
}

static struct sbk_message *
sbk_get_message(struct sbk_ctx *ctx, sqlite3_stmt *stm)
{
	struct sbk_message	*msg;
	const unsigned char	*id;

	if ((msg = calloc(1, sizeof *msg)) == NULL) {
		warn(NULL);
		return NULL;
	}

	if ((id = sqlite3_column_text(stm, SBK_COLUMN_CONVERSATIONID)) ==
	    NULL) {
		/* Likely message with error */
		warnx("Conversation recipient has null id");
		msg->conversation = NULL;
	} else {
		if (sbk_get_recipient_from_conversation_id(ctx,
		    &msg->conversation, (const char *)id) == -1)
			goto error;
		if (msg->conversation == NULL)
			warnx("Cannot find conversation recipient for id %s",
			    id);
	}

	if ((id = sqlite3_column_text(stm, SBK_COLUMN_ID)) == NULL) {
		msg->source = NULL;
	} else {
		if (sbk_get_recipient_from_conversation_id(ctx, &msg->source,
		    (const char *)id) == -1)
			goto error;
		if (msg->source == NULL)
			warnx("Cannot find source recipient for id %s", id);
	}

	if (sbk_sqlite_column_text_copy(ctx, &msg->type, stm, SBK_COLUMN_TYPE)
	    == -1)
		goto error;

	if (sbk_sqlite_column_text_copy(ctx, &msg->text, stm, SBK_COLUMN_BODY)
	    == -1)
		goto error;

	if (sbk_sqlite_column_text_copy(ctx, &msg->json, stm, SBK_COLUMN_JSON)
	    == -1)
		goto error;

	msg->time_sent = sqlite3_column_int64(stm, SBK_COLUMN_SENT_AT);

	if (sbk_parse_message_json(ctx, msg) == -1)
		goto error;

	if (sbk_insert_mentions(&msg->text, msg->mentions) == -1)
		goto error;

	return msg;

error:
	sbk_free_message(msg);
	return NULL;
}

static struct sbk_message_list *
sbk_get_message_list(struct sbk_ctx *ctx, sqlite3_stmt *stm)
{
	struct sbk_message_list	*lst;
	struct sbk_message	*msg;
	int			 ret;

	if ((lst = malloc(sizeof *lst)) == NULL) {
		warn(NULL);
		goto error;
	}

	SIMPLEQ_INIT(lst);

	while ((ret = sbk_sqlite_step(ctx->db, stm)) == SQLITE_ROW) {
		if ((msg = sbk_get_message(ctx, stm)) == NULL)
			goto error;
		SIMPLEQ_INSERT_TAIL(lst, msg, entries);
	}

	if (ret != SQLITE_DONE)
		goto error;

	sqlite3_finalize(stm);
	return lst;

error:
	sbk_free_message_list(lst);
	sqlite3_finalize(stm);
	return NULL;
}

struct sbk_message_list *
sbk_get_messages(struct sbk_ctx *ctx, struct sbk_conversation *cnv)
{
	sqlite3_stmt	*stm;
	const char	*query;

	if (ctx->db_version < 20)
		query = SBK_QUERY_8;
	else
		query = SBK_QUERY_20;

	if (sbk_sqlite_prepare(ctx->db, &stm, query) == -1)
		return NULL;

	if (sbk_sqlite_bind_text(ctx->db, stm, 1, cnv->id) == -1) {
		sqlite3_finalize(stm);
		return NULL;
	}

	return sbk_get_message_list(ctx, stm);
}

struct sbk_message_list *
sbk_get_messages_sent_after(struct sbk_ctx *ctx, struct sbk_conversation *cnv,
    time_t min)
{
	sqlite3_stmt	*stm;
	const char	*query;

	if (ctx->db_version < 20)
		query = SBK_QUERY_SENT_AFTER_8;
	else
		query = SBK_QUERY_SENT_AFTER_20;

	if (sbk_sqlite_prepare(ctx->db, &stm, query) == -1)
		return NULL;

	if (sbk_sqlite_bind_text(ctx->db, stm, 1, cnv->id) == -1) {
		sqlite3_finalize(stm);
		return NULL;
	}

	if (sbk_sqlite_bind_time(ctx->db, stm, 2, min) == -1) {
		sqlite3_finalize(stm);
		return NULL;
	}

	return sbk_get_message_list(ctx, stm);
}

struct sbk_message_list *
sbk_get_messages_sent_before(struct sbk_ctx *ctx, struct sbk_conversation *cnv,
    time_t max)
{
	sqlite3_stmt	*stm;
	const char	*query;

	if (ctx->db_version < 20)
		query = SBK_QUERY_SENT_BEFORE_8;
	else
		query = SBK_QUERY_SENT_BEFORE_20;

	if (sbk_sqlite_prepare(ctx->db, &stm, query) == -1)
		return NULL;

	if (sbk_sqlite_bind_text(ctx->db, stm, 1, cnv->id) == -1) {
		sqlite3_finalize(stm);
		return NULL;
	}

	if (sbk_sqlite_bind_time(ctx->db, stm, 2, max) == -1) {
		sqlite3_finalize(stm);
		return NULL;
	}

	return sbk_get_message_list(ctx, stm);
}

struct sbk_message_list *
sbk_get_messages_sent_between(struct sbk_ctx *ctx,
    struct sbk_conversation *cnv, time_t min, time_t max)
{
	sqlite3_stmt	*stm;
	const char	*query;

	if (ctx->db_version < 20)
		query = SBK_QUERY_SENT_BETWEEN_8;
	else
		query = SBK_QUERY_SENT_BETWEEN_20;

	if (sbk_sqlite_prepare(ctx->db, &stm, query) == -1)
		return NULL;

	if (sbk_sqlite_bind_text(ctx->db, stm, 1, cnv->id) == -1) {
		sqlite3_finalize(stm);
		return NULL;
	}

	if (sbk_sqlite_bind_time(ctx->db, stm, 2, min) == -1) {
		sqlite3_finalize(stm);
		return NULL;
	}

	if (sbk_sqlite_bind_time(ctx->db, stm, 3, max) == -1) {
		sqlite3_finalize(stm);
		return NULL;
	}

	return sbk_get_message_list(ctx, stm);
}

int
sbk_is_outgoing_message(const struct sbk_message *msg)
{
	return msg->type != NULL && strcmp(msg->type, "outgoing") == 0;
}
