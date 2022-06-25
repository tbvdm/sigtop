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
#include <strings.h>

#include "sbk-internal.h"

/* For database version 19 */
#define SBK_RECIPIENTS_QUERY_19						\
	"SELECT "							\
	"id, "								\
	"type, "							\
	"name, "							\
	"profileName, "							\
	"profileFamilyName, "						\
	"profileFullName, "						\
	"CASE type "							\
		"WHEN 'private' THEN '+' || id "			\
		"ELSE NULL "						\
	"END, "				/* e164 */			\
	"NULL "				/* uuid */			\
	"FROM conversations"

/* For database versions >= 20 */
#define SBK_RECIPIENTS_QUERY_20						\
	"SELECT "							\
	"id, "								\
	"type, "							\
	"name, "							\
	"profileName, "							\
	"profileFamilyName, "						\
	"profileFullName, "						\
	"e164, "							\
	"uuid "								\
	"FROM conversations"

#define SBK_RECIPIENTS_COLUMN_ID		0
#define SBK_RECIPIENTS_COLUMN_TYPE		1
#define SBK_RECIPIENTS_COLUMN_NAME		2
#define SBK_RECIPIENTS_COLUMN_PROFILENAME	3
#define SBK_RECIPIENTS_COLUMN_PROFILEFAMILYNAME	4
#define SBK_RECIPIENTS_COLUMN_PROFILEFULLNAME	5
#define SBK_RECIPIENTS_COLUMN_E164		6
#define SBK_RECIPIENTS_COLUMN_UUID		7

static int sbk_cmp_recipient_entries(struct sbk_recipient_entry *,
    struct sbk_recipient_entry *);

RB_GENERATE(sbk_recipient_tree, sbk_recipient_entry, entries,
    sbk_cmp_recipient_entries)

static int
sbk_cmp_recipient_entries(struct sbk_recipient_entry *e,
    struct sbk_recipient_entry *f)
{
	return strcmp(e->id, f->id);
}

static void
sbk_free_recipient_entry(struct sbk_recipient_entry *ent)
{
	if (ent == NULL)
		return;

	switch (ent->recipient.type) {
	case SBK_CONTACT:
		if (ent->recipient.contact != NULL) {
			free(ent->recipient.contact->uuid);
			free(ent->recipient.contact->name);
			free(ent->recipient.contact->profile_name);
			free(ent->recipient.contact->profile_family_name);
			free(ent->recipient.contact->profile_joined_name);
			free(ent->recipient.contact->phone);
			free(ent->recipient.contact);
		}
		break;
	case SBK_GROUP:
		if (ent->recipient.group != NULL) {
			free(ent->recipient.group->name);
			free(ent->recipient.group);
		}
		break;
	}

	free(ent->id);
	free(ent);
}

void
sbk_free_recipient_tree(struct sbk_ctx *ctx)
{
	struct sbk_recipient_entry *ent;

	while ((ent = RB_ROOT(&ctx->recipients)) != NULL) {
		RB_REMOVE(sbk_recipient_tree, &ctx->recipients, ent);
		sbk_free_recipient_entry(ent);
	}
}

static struct sbk_recipient_entry *
sbk_get_recipient_entry(struct sbk_ctx *ctx, sqlite3_stmt *stm)
{
	struct sbk_recipient_entry	*ent;
	struct sbk_contact		*con;
	struct sbk_group		*grp;
	const unsigned char		*type;

	if ((ent = calloc(1, sizeof *ent)) == NULL) {
		sbk_error_set(ctx, NULL);
		return NULL;
	}

	if (sbk_sqlite_column_text_copy(ctx, &ent->id, stm,
	    SBK_RECIPIENTS_COLUMN_ID) == -1)
		goto error;

	if ((type = sqlite3_column_text(stm, SBK_RECIPIENTS_COLUMN_TYPE)) ==
	    NULL) {
		sbk_error_sqlite_set(ctx, "Cannot get column text");
		goto error;
	}

	if (strcmp((const char *)type, "private") == 0)
		ent->recipient.type = SBK_CONTACT;
	else if (strcmp((const char *)type, "group") == 0)
		ent->recipient.type = SBK_GROUP;
	else {
		sbk_error_setx(ctx, "Unknown recipient type");
		goto error;
	}

	switch (ent->recipient.type) {
	case SBK_CONTACT:
		con = ent->recipient.contact = calloc(1, sizeof *con);
		if (con == NULL) {
			sbk_error_set(ctx, NULL);
			goto error;
		}

		if (sbk_sqlite_column_text_copy(ctx, &con->name,
		    stm, SBK_RECIPIENTS_COLUMN_NAME) == -1)
			goto error;

		if (sbk_sqlite_column_text_copy(ctx, &con->profile_name,
		    stm, SBK_RECIPIENTS_COLUMN_PROFILENAME) == -1)
			goto error;

		if (sbk_sqlite_column_text_copy(ctx, &con->profile_family_name,
		    stm, SBK_RECIPIENTS_COLUMN_PROFILEFAMILYNAME) == -1)
			goto error;

		if (sbk_sqlite_column_text_copy(ctx, &con->profile_joined_name,
		    stm, SBK_RECIPIENTS_COLUMN_PROFILEFULLNAME) == -1)
			goto error;

		if (sbk_sqlite_column_text_copy(ctx, &con->phone,
		    stm, SBK_RECIPIENTS_COLUMN_E164) == -1)
			goto error;

		if (sbk_sqlite_column_text_copy(ctx, &con->uuid,
		    stm, SBK_RECIPIENTS_COLUMN_UUID) == -1)
			goto error;

		break;

	case SBK_GROUP:
		grp = ent->recipient.group = calloc(1, sizeof *grp);
		if (grp == NULL) {
			sbk_error_set(ctx, NULL);
			goto error;
		}

		if (sbk_sqlite_column_text_copy(ctx, &grp->name,
		    stm, SBK_RECIPIENTS_COLUMN_NAME) == -1)
			goto error;
	}

	return ent;

error:
	sbk_free_recipient_entry(ent);
	return NULL;
}

int
sbk_build_recipient_tree(struct sbk_ctx *ctx)
{
	struct sbk_recipient_entry	*ent;
	sqlite3_stmt			*stm;
	const char			*query;
	int				 ret;

	if (!RB_EMPTY(&ctx->recipients))
		return 0;

	if (ctx->db_version < 20)
		query = SBK_RECIPIENTS_QUERY_19;
	else
		query = SBK_RECIPIENTS_QUERY_20;

	if (sbk_sqlite_prepare(ctx, ctx->db, &stm, query) == -1)
		return -1;

	while ((ret = sbk_sqlite_step(ctx, ctx->db, stm)) == SQLITE_ROW) {
		if ((ent = sbk_get_recipient_entry(ctx, stm)) == NULL)
			goto error;
		RB_INSERT(sbk_recipient_tree, &ctx->recipients, ent);
	}

	if (ret != SQLITE_DONE)
		goto error;

	sqlite3_finalize(stm);
	return 0;

error:
	sbk_free_recipient_tree(ctx);
	sqlite3_finalize(stm);
	return -1;
}

int
sbk_get_recipient_from_conversation_id(struct sbk_ctx *ctx,
    struct sbk_recipient **rcp, const char *id)
{
	struct sbk_recipient_entry find, *result;

	if (sbk_build_recipient_tree(ctx) == -1)
		return -1;

	find.id = (char *)id;
	result = RB_FIND(sbk_recipient_tree, &ctx->recipients, &find);
	*rcp = (result == NULL) ? NULL : &result->recipient;
	return 0;
}

int
sbk_get_recipient_from_phone(struct sbk_ctx *ctx, struct sbk_recipient **rcp,
    const char *phone)
{
	struct sbk_recipient_entry *ent;

	if (sbk_build_recipient_tree(ctx) == -1)
		return -1;

	*rcp = NULL;
	RB_FOREACH(ent, sbk_recipient_tree, &ctx->recipients)
		if (ent->recipient.type == SBK_CONTACT &&
		    ent->recipient.contact->phone != NULL &&
		    strcmp(phone, ent->recipient.contact->phone) == 0) {
			*rcp = &ent->recipient;
			break;
		}

	return 0;
}

int
sbk_get_recipient_from_uuid(struct sbk_ctx *ctx, struct sbk_recipient **rcp,
    const char *uuid)
{
	struct sbk_recipient_entry *ent;

	if (sbk_build_recipient_tree(ctx) == -1)
		return -1;

	*rcp = NULL;
	RB_FOREACH(ent, sbk_recipient_tree, &ctx->recipients)
		if (ent->recipient.type == SBK_CONTACT &&
		    ent->recipient.contact->uuid != NULL &&
		    strcasecmp(uuid, ent->recipient.contact->uuid) == 0) {
			*rcp = &ent->recipient;
			break;
		}

	return 0;
}

const char *
sbk_get_recipient_display_name(const struct sbk_recipient *rcp)
{
	if (rcp != NULL)
		switch (rcp->type) {
		case SBK_CONTACT:
			if (rcp->contact->name != NULL)
				return rcp->contact->name;
			if (rcp->contact->profile_joined_name != NULL)
				return rcp->contact->profile_joined_name;
			if (rcp->contact->profile_name != NULL)
				return rcp->contact->profile_name;
			if (rcp->contact->phone != NULL)
				return rcp->contact->phone;
			if (rcp->contact->uuid != NULL)
				return rcp->contact->uuid;
			break;
		case SBK_GROUP:
			if (rcp->group->name != NULL)
				return rcp->group->name;
			break;
		}

	return "Unknown";
}
