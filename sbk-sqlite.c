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

#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "sbk-internal.h"

/* UTF-8 encoding of FSI (U+2068) and PDI (U+2069) */
#define SBK_FSI		"\xe2\x81\xa8"
#define SBK_FSI_LEN	(sizeof SBK_FSI - 1)
#define SBK_PDI		"\xe2\x81\xa9"
#define SBK_PDI_LEN	(sizeof SBK_PDI - 1)

int
sbk_sqlite_open(struct sbk_ctx *ctx, sqlite3 **db, const char *path, int flags)
{
	if (sqlite3_open_v2(path, db, flags, NULL) != SQLITE_OK) {
		sbk_error_sqlite_setd(ctx, *db, "Cannot open database");
		return -1;
	}

	return 0;
}

int
sbk_sqlite_prepare(struct sbk_ctx *ctx, sqlite3 *db, sqlite3_stmt **stm,
    const char *query)
{
	if (sqlite3_prepare_v2(db, query, -1, stm, NULL) != SQLITE_OK) {
		sbk_error_sqlite_setd(ctx, db, "Cannot prepare SQL statement");
		return -1;
	}

	return 0;
}

int
sbk_sqlite_bind_int64(struct sbk_ctx *ctx, sqlite3 *db, sqlite3_stmt *stm,
    int idx, int64_t val)
{
	if (sqlite3_bind_int64(stm, idx, val) != SQLITE_OK) {
		sbk_error_sqlite_setd(ctx, db, "Cannot bind SQL parameter");
		return -1;
	}

	return 0;
}

int
sbk_sqlite_bind_text(struct sbk_ctx *ctx, sqlite3 *db, sqlite3_stmt *stm,
    int idx, const char *val)
{
	if (sqlite3_bind_text(stm, idx, val, -1, SQLITE_STATIC) != SQLITE_OK) {
		sbk_error_sqlite_setd(ctx, db, "Cannot bind SQL parameter");
		return -1;
	}

	return 0;
}

int
sbk_sqlite_bind_time(struct sbk_ctx *ctx, sqlite3 *db, sqlite3_stmt *stm,
    int idx, time_t val)
{
	int64_t msec;

	msec = (int64_t)val * 1000;
	return sbk_sqlite_bind_int64(ctx, db, stm, idx, msec);
}

int
sbk_sqlite_step(struct sbk_ctx *ctx, sqlite3 *db, sqlite3_stmt *stm)
{
	int ret;

	ret = sqlite3_step(stm);
	if (ret != SQLITE_ROW && ret != SQLITE_DONE)
		sbk_error_sqlite_setd(ctx, db, "Cannot execute SQL statement");

	return ret;
}

int
sbk_sqlite_column_text_copy(struct sbk_ctx *ctx, char **buf, sqlite3_stmt *stm,
    int idx)
{
	const char	*sub, *txt;
	size_t		 len;

	*buf = NULL;

	if (sqlite3_column_type(stm, idx) == SQLITE_NULL)
		return 0;

	if ((txt = (const char *)sqlite3_column_text(stm, idx)) == NULL) {
		sbk_error_sqlite_set(ctx, "Cannot get column text");
		return -1;
	}

	/*
	 * If the FSI character (U+2068) appears at the beginning of the text
	 * and the PDI character (U+2069) at the end, then skip both
	 */
	sub = NULL;
	if (strncmp(txt, SBK_FSI, SBK_FSI_LEN) == 0) {
		len = strlen(txt + SBK_FSI_LEN);
		if (len >= SBK_PDI_LEN) {
			len -= SBK_PDI_LEN;
			if (strcmp(txt + SBK_FSI_LEN + len, SBK_PDI) == 0)
				sub = txt + SBK_FSI_LEN;
		}
	}

	if (sub == NULL)
		*buf = strdup(txt);
	else
		*buf = strndup(sub, len);

	if (*buf == NULL) {
		sbk_error_set(ctx, NULL);
		return -1;
	}

	return 0;
}

int
sbk_sqlite_exec(struct sbk_ctx *ctx, sqlite3 *db, const char *sql)
{
	char *errmsg;

	if (sqlite3_exec(db, sql, NULL, NULL, &errmsg) != SQLITE_OK) {
		sbk_error_setx(ctx, "Cannot execute SQL statement: %s",
		    errmsg);
		sqlite3_free(errmsg);
		return -1;
	}

	return 0;
}

int
sbk_sqlite_key(struct sbk_ctx *ctx, sqlite3 *db, const char *key)
{
	if (sqlite3_key(db, key, strlen(key)) == -1) {
		sbk_error_sqlite_setd(ctx, db, "Cannot set key");
		return -1;
	}

	return 0;
}

int
sbk_get_database_version(struct sbk_ctx *ctx)
{
	sqlite3_stmt	*stm;
	int		 version;

	if (sbk_sqlite_prepare(ctx, ctx->db, &stm, "PRAGMA user_version") ==
	    -1)
		return -1;

	if (sbk_sqlite_step(ctx, ctx->db, stm) != SQLITE_ROW) {
		sqlite3_finalize(stm);
		return -1;
	}

	if ((version = sqlite3_column_int(stm, 0)) < 0) {
		sbk_error_setx(ctx, "Negative database version");
		sqlite3_finalize(stm);
		return -1;
	}

	sqlite3_finalize(stm);
	return version;
}

int
sbk_set_database_version(struct sbk_ctx *ctx, sqlite3 *db, const char *schema,
    int version)
{
	char	*sql;
	int	 ret;

	if (asprintf(&sql, "PRAGMA %s.user_version = %d", schema, version) ==
	    -1) {
		sbk_error_setx(ctx, "asprintf() failed");
		return -1;
	}

	ret = sbk_sqlite_exec(ctx, db, sql);
	free(sql);
	return ret;
}

/*
 * To decrypt an encrypted database to a plaintext database, the SQLCipher
 * documentation recommends to do the following:
 *
 * 1. Open the encrypted database.
 * 2. Attach the plaintext database.
 * 3. Use the sqlcipher_export() SQL function to decrypt.
 *
 * This doesn't work in our case, because we insist on opening the Signal
 * Desktop database in read-only mode.
 *
 * The SQLite backup API doesn't work either, because it does not support
 * encrypted-to-plaintext backups.
 *
 * However, since SQLCipher 4.3.0, the backup API does support
 * encrypted-to-encrypted backups. This allows us to do the following:
 *
 * 1. Open the Signal Desktop database in read-only mode.
 * 2. Create a temporary encrypted database in memory.
 * 3. Back up the Signal Desktop database to the temporary database.
 * 4. Attach a new plaintext database to the temporary database.
 * 5. Use sqlcipher_export() to decrypt the temporary database to the plaintext
 *    database.
 */
int
sbk_write_database(struct sbk_ctx *ctx, const char *path)
{
	sqlite3		*db;
	sqlite3_backup	*bak;
	sqlite3_stmt	*stm;

	if (sbk_sqlite_open(ctx, &db, ":memory:",
	    SQLITE_OPEN_READWRITE | SQLITE_OPEN_CREATE) == -1)
		goto error;

	/* Set a dummy key to enable encryption */
	if (sbk_sqlite_key(ctx, db, "x") == -1)
		goto error;

	if ((bak = sqlite3_backup_init(db, "main", ctx->db, "main")) == NULL) {
		sbk_error_sqlite_setd(ctx, db, "Cannot write database");
		goto error;
	}

	if (sqlite3_backup_step(bak, -1) != SQLITE_DONE) {
		sbk_error_sqlite_setd(ctx, db, "Cannot write database");
		sqlite3_backup_finish(bak);
		goto error;
	}

	sqlite3_backup_finish(bak);

	/* Attaching with an empty key disables encryption */
	if (sbk_sqlite_prepare(ctx, db, &stm,
	    "ATTACH DATABASE ? AS plaintext KEY ''") == -1)
		goto error;

	if (sbk_sqlite_bind_text(ctx, db, stm, 1, path) == -1) {
		sqlite3_finalize(stm);
		goto error;
	}

	if (sbk_sqlite_step(ctx, db, stm) != SQLITE_DONE) {
		sqlite3_finalize(stm);
		goto error;
	}

	sqlite3_finalize(stm);

	if (sbk_sqlite_exec(ctx, db, "BEGIN TRANSACTION") == -1)
		goto error;

	if (sbk_sqlite_exec(ctx, db, "SELECT sqlcipher_export('plaintext')") ==
	    -1)
		goto error;

	if (sbk_set_database_version(ctx, db, "plaintext", ctx->db_version) ==
	    -1)
		goto error;

	if (sbk_sqlite_exec(ctx, db, "END TRANSACTION") == -1)
		goto error;

	if (sbk_sqlite_exec(ctx, db, "DETACH DATABASE plaintext") == -1)
		goto error;

	if (sqlite3_close(db) != SQLITE_OK) {
		sbk_error_sqlite_setd(ctx, db, "Cannot close database");
		return -1;
	}

	return 0;

error:
	sqlite3_close(db);
	return -1;
}
