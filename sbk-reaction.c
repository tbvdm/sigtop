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

#include "sbk-internal.h"

static void
sbk_free_reaction(struct sbk_reaction *rct)
{
	if (rct != NULL) {
		free(rct->emoji);
		free(rct);
	}
}

void
sbk_free_reaction_list(struct sbk_reaction_list *lst)
{
	struct sbk_reaction *rct;

	if (lst != NULL) {
		while ((rct = SIMPLEQ_FIRST(lst)) != NULL) {
			SIMPLEQ_REMOVE_HEAD(lst, entries);
			sbk_free_reaction(rct);
		}
		free(lst);
	}
}

static int
sbk_get_recipient_from_reaction_id(struct sbk_ctx *ctx,
    struct sbk_recipient **rcp, const char *id)
{
	/* XXX */
	if (ctx->db_version < 20) {
		if (id[0] == '+')
			id++;
		return sbk_get_recipient_from_conversation_id(ctx, rcp, id);
	} else {
		if (id[0] == '+')
			return sbk_get_recipient_from_phone(ctx, rcp, id);
		else
			return sbk_get_recipient_from_conversation_id(ctx, rcp,
			    id);
	}
}

static int
sbk_add_reaction(struct sbk_ctx *ctx, struct sbk_message *msg,
    jsmntok_t *tokens)
{
	struct sbk_reaction	*rct;
	char			*id;
	int			 idx, ret;

	if ((rct = calloc(1, sizeof *rct)) == NULL) {
		sbk_error_set(ctx, NULL);
		goto error;
	}

	if (tokens[0].type != JSMN_OBJECT) {
		sbk_error_setx(ctx, "Unexpected reaction JSON type");
		goto error;
	}

	/*
	 * Get recipient
	 */

	idx = sbk_jsmn_get_string(msg->json, tokens, "fromId");
	if (idx == -1) {
		sbk_error_setx(ctx, "Missing reaction fromId");
		goto error;
	}

	id = sbk_jsmn_strdup(msg->json, &tokens[idx]);
	if (id == NULL) {
		sbk_error_set(ctx, NULL);
		goto error;
	}

	ret = sbk_get_recipient_from_reaction_id(ctx, &rct->recipient, id);
	if (ret == -1) {
		free(id);
		goto error;
	}

	if (rct->recipient == NULL)
		sbk_warnx(ctx, "Cannot find reaction recipient for id %s", id);

	free(id);

	/*
	 * Get emoji
	 */

	idx = sbk_jsmn_get_string(msg->json, tokens, "emoji");
	if (idx == -1) {
		sbk_error_setx(ctx, "Missing reaction emoji");
		goto error;
	}

	rct->emoji = sbk_jsmn_strdup(msg->json, &tokens[idx]);
	if (rct->emoji == NULL) {
		sbk_error_set(ctx, NULL);
		goto error;
	}

	/*
	 * Get sent time
	 */

	idx = sbk_jsmn_get_number(msg->json, tokens, "targetTimestamp");
	if (idx == -1) {
		sbk_error_setx(ctx, "Missing reaction targetTimestamp");
		goto error;
	}

	if (sbk_jsmn_parse_uint64(&rct->time_sent, msg->json, &tokens[idx]) ==
	    -1) {
		sbk_error_setx(ctx, "Cannot parse reaction targetTimestamp");
		goto error;
	}

	/*
	 * Get received time
	 */

	idx = sbk_jsmn_get_number(msg->json, tokens, "timestamp");
	if (idx == -1) {
		sbk_error_setx(ctx, "Missing reaction timestamp");
		goto error;
	}

	if (sbk_jsmn_parse_uint64(&rct->time_recv, msg->json, &tokens[idx]) ==
	    -1) {
		sbk_error_setx(ctx, "Cannot parse reaction timestamp");
		goto error;
	}

	SIMPLEQ_INSERT_TAIL(msg->reactions, rct, entries);
	return 0;

error:
	sbk_free_reaction(rct);
	return -1;
}

int
sbk_parse_reaction_json(struct sbk_ctx *ctx, struct sbk_message *msg,
    jsmntok_t *tokens)
{
	int i, idx, size;

	if (tokens[0].size == 0)
		return 0;

	msg->reactions = malloc(sizeof *msg->reactions);
	if (msg->reactions == NULL) {
		sbk_error_set(ctx, NULL);
		goto error;
	}

	SIMPLEQ_INIT(msg->reactions);

	idx = 1;
	for (i = 0; i < tokens[0].size; i++) {
		if (sbk_add_reaction(ctx, msg, &tokens[idx]) == -1)
			goto error;
		/* Skip to next element in array */
		size = sbk_jsmn_get_total_token_size(&tokens[idx]);
		if (size == -1) {
			sbk_error_setx(ctx, "Cannot parse reaction JSON data");
			goto error;
		}
		idx += size;
	}

	return 0;

error:
	sbk_free_reaction_list(msg->reactions);
	msg->reactions = NULL;
	return -1;
}
