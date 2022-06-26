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

#include <sys/types.h>

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "sbk-internal.h"

#define SBK_ATTACHMENT_DIR "attachments.noindex"

void
sbk_free_attachment(struct sbk_attachment *att)
{
	if (att != NULL) {
		free(att->path);
		free(att->filename);
		free(att->content_type);
		free(att);
	}
}

void
sbk_free_attachment_list(struct sbk_attachment_list *lst)
{
	struct sbk_attachment *att;

	if (lst != NULL) {
		while ((att = TAILQ_FIRST(lst)) != NULL) {
			TAILQ_REMOVE(lst, att, entries);
			sbk_free_attachment(att);
		}
		free(lst);
	}
}

static int
sbk_add_attachment(struct sbk_ctx *ctx, struct sbk_message *msg,
    jsmntok_t *tokens)
{
	struct sbk_attachment	*att;
	char			*c;
	int			 idx;

	if ((att = calloc(1, sizeof *att)) == NULL) {
		sbk_error_set(ctx, NULL);
		goto error;
	}

	idx = sbk_jsmn_get_string(msg->json, tokens, "path");
	if (idx != -1) {
		att->path = sbk_jsmn_parse_string(msg->json, &tokens[idx]);
		if (att->path == NULL) {
			sbk_error_setx(ctx, "Cannot parse attachment path");
			goto error;
		}
	}

	/* Replace Windows directory separators, if any */
	if (att->path != NULL) {
		c = att->path;
		while ((c = strchr(c, '\\')) != NULL)
			*c++ = '/';
	}

	idx = sbk_jsmn_get_string(msg->json, tokens, "fileName");
	if (idx != -1) {
		att->filename = sbk_jsmn_parse_string(msg->json, &tokens[idx]);
		if (att->filename == NULL) {
			sbk_error_setx(ctx, "Cannot parse attachment "
			    "fileName");
			goto error;
		}
	}

	idx = sbk_jsmn_get_string(msg->json, tokens, "contentType");
	if (idx != -1) {
		att->content_type = sbk_jsmn_parse_string(msg->json,
		    &tokens[idx]);
		if (att->content_type == NULL) {
			sbk_error_setx(ctx, "Cannot parse attachment "
			    "contentType");
			goto error;
		}
	}

	idx = sbk_jsmn_get_number(msg->json, tokens, "size");
	if (idx != -1) {
		if (sbk_jsmn_parse_uint64(&att->size, msg->json, &tokens[idx])
		    == -1) {
			sbk_error_setx(ctx, "Cannot parse attachment size");
			goto error;
		}
	}

	att->time_sent = msg->time_sent;
	att->time_recv = msg->time_recv;
	TAILQ_INSERT_TAIL(msg->attachments, att, entries);
	return 0;

error:
	sbk_free_attachment(att);
	return -1;
}

int
sbk_parse_attachment_json(struct sbk_ctx *ctx, struct sbk_message *msg,
    jsmntok_t *tokens)
{
	int i, idx, size;

	if (tokens[0].size == 0)
		return 0;

	msg->attachments = malloc(sizeof *msg->attachments);
	if (msg->attachments == NULL) {
		sbk_error_set(ctx, NULL);
		goto error;
	}

	TAILQ_INIT(msg->attachments);

	idx = 1;
	for (i = 0; i < tokens[0].size; i++) {
		if (tokens[idx].type != JSMN_OBJECT) {
			sbk_error_setx(ctx, "Unexpected attachment JSON type");
			goto error;
		}
		if (sbk_add_attachment(ctx, msg, &tokens[idx]) == -1)
			goto error;
		/* Skip to next element in array */
		size = sbk_jsmn_get_total_token_size(&tokens[idx]);
		if (size == -1) {
			sbk_error_setx(ctx, "Cannot parse attachment JSON "
			    "data");
			goto error;
		}
		idx += size;
	}

	return 0;

error:
	sbk_free_attachment_list(msg->attachments);
	msg->attachments = NULL;
	return -1;
}

static struct sbk_attachment_list *
sbk_get_attachment_list(struct sbk_ctx *ctx, struct sbk_message_list *msg_lst)
{
	struct sbk_attachment_list	*att_lst;
	struct sbk_message		*msg;

	if ((att_lst = malloc(sizeof *att_lst)) == NULL) {
		sbk_error_set(ctx, NULL);
		sbk_free_message_list(msg_lst);
		return NULL;
	}

	TAILQ_INIT(att_lst);

	SIMPLEQ_FOREACH(msg, msg_lst, entries)
		if (msg->attachments != NULL)
			TAILQ_CONCAT(att_lst, msg->attachments, entries);

	sbk_free_message_list(msg_lst);
	return att_lst;
}

struct sbk_attachment_list *
sbk_get_attachments(struct sbk_ctx *ctx, struct sbk_conversation *cnv)
{
	struct sbk_message_list *lst;

	if ((lst = sbk_get_messages(ctx, cnv)) == NULL)
		return NULL;

	return sbk_get_attachment_list(ctx, lst);
}

struct sbk_attachment_list *
sbk_get_attachments_sent_after(struct sbk_ctx *ctx,
    struct sbk_conversation *cnv, time_t min)
{
	struct sbk_message_list *lst;

	if ((lst = sbk_get_messages_sent_after(ctx, cnv, min)) == NULL)
		return NULL;

	return sbk_get_attachment_list(ctx, lst);
}

struct sbk_attachment_list *
sbk_get_attachments_sent_before(struct sbk_ctx *ctx,
    struct sbk_conversation *cnv, time_t max)
{
	struct sbk_message_list *lst;

	if ((lst = sbk_get_messages_sent_before(ctx, cnv, max)) == NULL)
		return NULL;

	return sbk_get_attachment_list(ctx, lst);
}

struct sbk_attachment_list *
sbk_get_attachments_sent_between(struct sbk_ctx *ctx,
    struct sbk_conversation *cnv, time_t min, time_t max)
{
	struct sbk_message_list *lst;

	if ((lst = sbk_get_messages_sent_between(ctx, cnv, min, max)) == NULL)
		return NULL;

	return sbk_get_attachment_list(ctx, lst);
}

char *
sbk_get_attachment_path(struct sbk_ctx *ctx, struct sbk_attachment *att)
{
	char *path;

	if (att->path == NULL) {
		sbk_error_setx(ctx, "Missing attachment path");
		return NULL;
	}

	if (asprintf(&path, "%s/%s/%s", ctx->dir, SBK_ATTACHMENT_DIR,
	    att->path) == -1) {
		sbk_error_setx(ctx, "asprintf() failed");
		return NULL;
	}

	return path;
}
