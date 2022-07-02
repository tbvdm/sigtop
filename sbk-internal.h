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

#ifndef SBK_INTERNAL_H
#define SBK_INTERNAL_H

#define JSMN_HEADER

#include <stdarg.h>

#include "compat.h"
#include "jsmn.h"
#include "sbk.h"
#include "sqlite3.h"

#define sbk_warnx(ctx, ...) warnx(__VA_ARGS__)

struct sbk_recipient_entry {
	char		*id;
	struct sbk_recipient recipient;
	RB_ENTRY(sbk_recipient_entry) entries;
};

RB_HEAD(sbk_recipient_tree, sbk_recipient_entry);

struct sbk_ctx {
	char		*dir;
	sqlite3		*db;
	int		 db_version;
	char		*error;
	struct sbk_recipient_tree recipients;
};

void	 sbk_error_clear(struct sbk_ctx *);
void	 sbk_error_set(struct sbk_ctx *, const char *, ...);
void	 sbk_error_setx(struct sbk_ctx *, const char *, ...);
void	 sbk_error_sqlite_vsetd(struct sbk_ctx *, sqlite3 *, const char *,
	    va_list);
void	 sbk_error_sqlite_setd(struct sbk_ctx *, sqlite3 *, const char *,
	    ...);
void	 sbk_error_sqlite_set(struct sbk_ctx *, const char *, ...);

int	 sbk_sqlite_open(struct sbk_ctx *, sqlite3 **, const char *, int);
int	 sbk_sqlite_prepare(struct sbk_ctx *, sqlite3 *, sqlite3_stmt **,
	    const char *);
int	 sbk_sqlite_bind_int64(struct sbk_ctx *, sqlite3 *, sqlite3_stmt *,
	    int, int64_t);
int	 sbk_sqlite_bind_text(struct sbk_ctx *, sqlite3 *, sqlite3_stmt *, int,
	    const char *);
int	 sbk_sqlite_bind_time(struct sbk_ctx *, sqlite3 *, sqlite3_stmt *,
	    int, time_t);
int	 sbk_sqlite_step(struct sbk_ctx *, sqlite3 *, sqlite3_stmt *);
int	 sbk_sqlite_column_text_copy(struct sbk_ctx *, char **, sqlite3_stmt *,
	    int);
int	 sbk_sqlite_exec(struct sbk_ctx *, sqlite3 *, const char *);
int	 sbk_sqlite_key(struct sbk_ctx *, sqlite3 *, const char *);
int	 sbk_get_database_version(struct sbk_ctx *);
int	 sbk_set_database_version(struct sbk_ctx *, sqlite3 *, const char *,
	    int);

int	 sbk_jsmn_parse(const char *, size_t, jsmntok_t *, size_t);
int	 sbk_jsmn_get_total_token_size(const jsmntok_t *);
int	 sbk_jsmn_get_array(const char *, const jsmntok_t *, const char *);
int	 sbk_jsmn_get_object(const char *, const jsmntok_t *, const char *);
int	 sbk_jsmn_get_string(const char *, const jsmntok_t *, const char *);
int	 sbk_jsmn_get_number(const char *, const jsmntok_t *, const char *);
int	 sbk_jsmn_get_number_or_string(const char *, const jsmntok_t *,
	    const char *);
char	*sbk_jsmn_strdup(const char *, const jsmntok_t *);
char	*sbk_jsmn_parse_string(const char *, const jsmntok_t *);
int	 sbk_jsmn_parse_uint64(uint64_t *, const char *, const jsmntok_t *);

RB_PROTOTYPE(sbk_recipient_tree, sbk_recipient_entry, entries,
    sbk_cmp_recipient_entries)

int	 sbk_build_recipient_tree(struct sbk_ctx *);
void	 sbk_free_recipient_tree(struct sbk_ctx *);
int	 sbk_get_recipient_from_conversation_id(struct sbk_ctx *,
	    struct sbk_recipient **, const char *);
int	 sbk_get_recipient_from_phone(struct sbk_ctx *,
	    struct sbk_recipient **, const char *);
int	 sbk_get_recipient_from_uuid(struct sbk_ctx *, struct sbk_recipient **,
	    const char *);
const char *sbk_get_recipient_display_name(const struct sbk_recipient *);

int	 sbk_parse_attachment_json(struct sbk_ctx *, struct sbk_message *,
	    jsmntok_t *tokens);
void	 sbk_free_attachment(struct sbk_attachment *);

int	 sbk_parse_mention_json(struct sbk_ctx *, struct sbk_message *,
	    struct sbk_mention_list **, jsmntok_t *);
void	 sbk_free_mention_list(struct sbk_mention_list *);
int	 sbk_insert_mentions(struct sbk_ctx *, char **,
	    struct sbk_mention_list *);

int	 sbk_parse_reaction_json(struct sbk_ctx *, struct sbk_message *,
	    jsmntok_t *);
void	 sbk_free_reaction_list(struct sbk_reaction_list *);

int	 sbk_parse_quote_json(struct sbk_ctx *, struct sbk_message *,
	    jsmntok_t *);
void	 sbk_free_quote(struct sbk_quote *);

#endif
