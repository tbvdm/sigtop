/*
 * Copyright (c) 2022 Tim van der Molen <tim@kariliq.nl>
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

#include <stdint.h>
#include <stdlib.h>
#include <string.h>

#include "sbk-internal.h"
#include "utf.h"

void
sbk_free_mention_list(struct sbk_mention_list *lst)
{
	struct sbk_mention *mnt;

	if (lst != NULL) {
		while ((mnt = TAILQ_FIRST(lst)) != NULL) {
			TAILQ_REMOVE(lst, mnt, entries);
			free(mnt);
		}
		free(lst);
	}
}

static int
sbk_add_mention(struct sbk_ctx *ctx, struct sbk_message *msg,
    struct sbk_mention_list *lst, jsmntok_t *tokens)
{
	struct sbk_mention	*lst_mnt, *mnt;
	char			*uuid;
	int			 idx;

	if ((mnt = calloc(1, sizeof *mnt)) == NULL) {
		warn(NULL);
		goto error;
	}

	/*
	 * Get recipient
	 */

	idx = sbk_jsmn_get_string(msg->json, tokens, "mentionUuid");
	if (idx != -1) {
		uuid = sbk_jsmn_parse_string(msg->json, &tokens[idx]);
		if (uuid == NULL)
			goto error;

		if (sbk_get_recipient_from_uuid(ctx, &mnt->recipient, uuid) ==
		    -1)
			warnx("Cannot find mention recipient for uuid %s",
			    uuid);

		free(uuid);
	}

	/*
	 * Get start
	 */

	idx = sbk_jsmn_get_number(msg->json, tokens, "start");
	if (idx == -1) {
		warnx("Missing mention start");
		goto error;
	}

	if (sbk_jsmn_parse_uint64(&mnt->start, msg->json, &tokens[idx]) == -1)
		goto error;

	/*
	 * Get length
	 */

	idx = sbk_jsmn_get_number(msg->json, tokens, "length");
	if (idx == -1) {
		warnx("Missing mention length");
		goto error;
	}

	if (sbk_jsmn_parse_uint64(&mnt->length, msg->json, &tokens[idx]) == -1)
		goto error;

	/*
	 * Insert the mention in order. It seems the mentions usually are
	 * already properly ordered, so, as an optimisation, traverse the list
	 * in reverse direction.
	 */

	TAILQ_FOREACH_REVERSE(lst_mnt, lst, sbk_mention_list, entries)
		if (lst_mnt->start < mnt->start) {
			TAILQ_INSERT_AFTER(lst, lst_mnt, mnt, entries);
			break;
		}
	if (lst_mnt == NULL)
		TAILQ_INSERT_HEAD(lst, mnt, entries);

	return 0;

error:
	free(mnt);
	return -1;
}

int
sbk_parse_mention_json(struct sbk_ctx *ctx, struct sbk_message *msg,
    struct sbk_mention_list **lst, jsmntok_t *tokens)
{
	int i, idx, size;

	if (tokens[0].size == 0)
		return 0;

	*lst = malloc(sizeof **lst);
	if (*lst == NULL) {
		warn(NULL);
		goto error;
	}

	TAILQ_INIT(*lst);

	idx = 1;
	for (i = 0; i < tokens[0].size; i++) {
		if (tokens[idx].type != JSMN_OBJECT) {
			warnx("Unexpected mention JSON type");
			goto error;
		}
		if (sbk_add_mention(ctx, msg, *lst, &tokens[idx]) == -1)
			goto error;
		/* Skip to next element in array */
		size = sbk_jsmn_get_total_token_size(&tokens[idx]);
		if (size == -1)
			goto error;
		idx += size;
	}

	return 0;

error:
	sbk_free_mention_list(*lst);
	*lst = NULL;
	return -1;
}

int
sbk_insert_mentions(char **text, struct sbk_mention_list *lst)
{
	struct sbk_mention *mnt, *next_mnt;
	char		*new_text, *old_text;
	const char	*name;
	size_t		 copy_len, mention_len, name_len, new_off,
			 new_text_len, old_off, old_text_len;

	new_text = NULL;

	if (*text == NULL || lst == NULL || TAILQ_EMPTY(lst))
		return 0;

	/*
	 * Ensure the mentions are properly ordered and don't overlap. We
	 * depend on this when we write the new text.
	 */
	mnt = TAILQ_FIRST(lst);
	while ((next_mnt = TAILQ_NEXT(mnt, entries)) != NULL) {
		if (next_mnt->start < mnt->start + mnt->length)
			goto invalid;
		mnt = next_mnt;
	}

	/*
	 * Compute length of new text
	 */

	old_text = *text;
	old_text_len = strlen(old_text);
	new_text_len = old_text_len;

	TAILQ_FOREACH(mnt, lst, entries) {
		/* Convert character counts to byte counts */
		mnt->start = utf8_get_substring_length(
		    (unsigned char *)old_text, mnt->start);
		mnt->length = utf8_get_substring_length(
		    (unsigned char *)old_text + mnt->start, mnt->length);

		/* Subtract placeholder length */
		if (mnt->length > new_text_len)
			goto invalid;
		new_text_len -= mnt->length;

		/* Compute mention length (add 1 for the "@" prefix) */
		name = sbk_get_recipient_display_name(mnt->recipient);
		mention_len = strlen(name) + 1;

		/* Add mention length */
		if (mention_len > SIZE_MAX - new_text_len)
			goto invalid;
		new_text_len += mention_len;
	}

	if (new_text_len == SIZE_MAX)
		goto invalid;

	if ((new_text = malloc(new_text_len + 1)) == NULL) {
		warn(NULL);
		return -1;
	}

	/*
	 * Write new text, replacing placeholders with mentions
	 */

	old_off = new_off = 0;
	TAILQ_FOREACH(mnt, lst, entries) {
		/* Copy text preceding mention */
		copy_len = mnt->start - old_off;
		memcpy(new_text + new_off, old_text + old_off, copy_len);
		old_off = mnt->start + mnt->length;
		new_off += copy_len;

		/* Compute mention length (add 1 for the "@" prefix) */
		name = sbk_get_recipient_display_name(mnt->recipient);
		name_len = strlen(name);
		mention_len = name_len + 1;

		/* Update mention */
		mnt->start = new_off;
		mnt->length = mention_len;

		/* Write mention */
		new_text[new_off++] = '@';
		memcpy(new_text + new_off, name, name_len);
		new_off += name_len;
	}

	/* Copy text succeeding last mention */
	copy_len = strlen(old_text + old_off);
	memcpy(new_text + new_off, old_text + old_off, copy_len);
	new_off += copy_len;

	new_text[new_off] = '\0';

	free(*text);
	*text = new_text;
	return 0;

invalid:
	warnx("Invalid mention");
	free(new_text);
	return -1;
}
