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

void
sbk_free_conversation_list(struct sbk_conversation_list *lst)
{
	struct sbk_conversation *cnv;

	if (lst != NULL) {
		while ((cnv = SIMPLEQ_FIRST(lst)) != NULL) {
			SIMPLEQ_REMOVE_HEAD(lst, entries);
			free(cnv);
		}
		free(lst);
	}
}

struct sbk_conversation_list *
sbk_get_conversations(struct sbk_ctx *ctx)
{
	struct sbk_conversation_list	*lst;
	struct sbk_conversation		*cnv;
	struct sbk_recipient_entry	*ent;

	if (sbk_build_recipient_tree(ctx) == -1)
		return NULL;

	if ((lst = malloc(sizeof *lst)) == NULL) {
		sbk_error_set(ctx, NULL);
		goto error;
	}

	SIMPLEQ_INIT(lst);

	RB_FOREACH(ent, sbk_recipient_tree, &ctx->recipients) {
		if ((cnv = malloc(sizeof *cnv)) == NULL) {
			sbk_error_set(ctx, NULL);
			goto error;
		}
		cnv->id = ent->id;
		cnv->recipient = &ent->recipient;
		SIMPLEQ_INSERT_TAIL(lst, cnv, entries);
	}

	return lst;

error:
	sbk_free_conversation_list(lst);
	return NULL;
}
