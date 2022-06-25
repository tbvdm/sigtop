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

#include <stdlib.h>
#include <string.h>

#include "sbk-internal.h"

/* Content type of the long-text attachment of a long message */
#define SBK_LONG_TEXT_TYPE "text/x-signal-plain"

void
sbk_free_quote(struct sbk_quote *qte)
{
	if (qte != NULL) {
		free(qte->text);
		sbk_free_attachment_list(qte->attachments);
		free(qte);
	}
}

static int
sbk_add_quote_attachment(struct sbk_ctx *ctx, struct sbk_message *msg,
    struct sbk_quote *qte, jsmntok_t *tokens)
{
	struct sbk_attachment	*att;
	int			 idx;

	if ((att = calloc(1, sizeof *att)) == NULL) {
		sbk_error_set(ctx, NULL);
		goto error;
	}

	if (tokens[0].type != JSMN_OBJECT) {
		sbk_error_setx(ctx, "Unexpected quote attachment JSON type");
		goto error;
	}

	idx = sbk_jsmn_get_string(msg->json, tokens, "fileName");
	if (idx != -1) {
		att->filename = sbk_jsmn_parse_string(msg->json, &tokens[idx]);
		if (att->filename == NULL) {
			sbk_error_setx(ctx, "Cannot parse quote attachment "
			    "fileName");
			goto error;
		}
	}

	idx = sbk_jsmn_get_string(msg->json, tokens, "contentType");
	if (idx != -1) {
		att->content_type = sbk_jsmn_parse_string(msg->json,
		    &tokens[idx]);
		if (att->content_type == NULL) {
			sbk_error_setx(ctx, "Cannot parse quote attachment "
			    "contentType");
			goto error;
		}
	}

	/* Do not expose long-message attachments */
	if (att->content_type != NULL &&
	    strcmp(att->content_type, SBK_LONG_TEXT_TYPE) == 0) {
		sbk_free_attachment(att);
		return 0;
	}

	att->time_sent = qte->id;
	TAILQ_INSERT_TAIL(qte->attachments, att, entries);
	return 0;

error:
	sbk_free_attachment(att);
	return -1;
}

static int
sbk_parse_quote_attachment_json(struct sbk_ctx *ctx, struct sbk_message *msg,
    struct sbk_quote *qte, jsmntok_t *tokens)
{
	int i, idx, size;

	if (tokens[0].size == 0)
		return 0;

	qte->attachments = malloc(sizeof *qte->attachments);
	if (qte->attachments == NULL) {
		sbk_error_set(ctx, NULL);
		goto error;
	}

	TAILQ_INIT(qte->attachments);

	idx = 1;
	for (i = 0; i < tokens[0].size; i++) {
		if (sbk_add_quote_attachment(ctx, msg, qte, &tokens[idx]) ==
		    -1)
			goto error;
		/* Skip to next element in array */
		size = sbk_jsmn_get_total_token_size(&tokens[idx]);
		if (size == -1) {
			sbk_error_setx(ctx, "Cannot parse quote attachment "
			    "JSON data");
			goto error;
		}
		idx += size;
	}

	return 0;

error:
	sbk_free_attachment_list(qte->attachments);
	qte->attachments = NULL;
	return -1;
}

int
sbk_parse_quote_json(struct sbk_ctx *ctx, struct sbk_message *msg,
    jsmntok_t *tokens)
{
	struct sbk_quote	*qte;
	char			*author;
	int			 idx;

	if ((qte = calloc(1, sizeof *qte)) == NULL) {
		sbk_error_set(ctx, NULL);
		goto error;
	}

	if (tokens[0].type != JSMN_OBJECT) {
		sbk_error_setx(ctx, "Unexpected quote JSON type");
		goto error;
	}

	/*
	 * Get id
	 *
	 * The id usually is a JSON number, but in at least one case it was a
	 * JSON string.
	 */

	idx = sbk_jsmn_get_number_or_string(msg->json, tokens, "id");
	if (idx == -1) {
		sbk_error_setx(ctx, "Missing quote id");
		goto error;
	}

	if (sbk_jsmn_parse_uint64(&qte->id, msg->json, &tokens[idx]) == -1) {
		sbk_error_setx(ctx, "Cannot parse quote id");
		goto error;
	}

	/*
	 * Get recipient
	 *
	 * Newer quotes have an "authorUuid" attribute. Older quotes have an
	 * "author" attribute containing a phone number.
	 */

	idx = sbk_jsmn_get_string(msg->json, tokens, "authorUuid");
	if (idx != -1) {
		author = sbk_jsmn_parse_string(msg->json, &tokens[idx]);
		if (author == NULL) {
			sbk_error_setx(ctx, "Cannot parse quote authorUuid");
			goto error;
		}

		if (sbk_get_recipient_from_uuid(ctx, &qte->recipient, author)
		    == -1) {
			free(author);
			goto error;
		}
	} else {
		idx = sbk_jsmn_get_string(msg->json, tokens, "author");
		if (idx == -1) {
			sbk_error_setx(ctx, "Missing author and authorUuid in "
			    "quote");
			goto error;
		}

		author = sbk_jsmn_parse_string(msg->json, &tokens[idx]);
		if (author == NULL) {
			sbk_error_setx(ctx, "Cannot parse quote author");
			goto error;
		}

		if (sbk_get_recipient_from_phone(ctx, &qte->recipient, author)
		    == -1) {
			free(author);
			goto error;
		}
	}

	free(author);

	/*
	 * Get text
	 */

	idx = sbk_jsmn_get_string(msg->json, tokens, "text");
	if (idx != -1) {
		qte->text = sbk_jsmn_parse_string(msg->json, &tokens[idx]);
		if (qte->text == NULL) {
			sbk_error_setx(ctx, "Cannot parse quote text");
			goto error;
		}
	}

	/*
	 * Get attachments
	 */

	idx = sbk_jsmn_get_array(msg->json, tokens, "attachments");
	if (idx != -1 &&
	    sbk_parse_quote_attachment_json(ctx, msg, qte, &tokens[idx]) == -1)
		goto error;

	msg->quote = qte;
	return 0;

error:
	sbk_free_quote(qte);
	return -1;
}
