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

#include <fcntl.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

#include "sbk-internal.h"

/* Read the database encryption key from a JSON file */
static int
sbk_get_key(char *buf, size_t bufsize, const char *path)
{
	jsmntok_t	tokens[64];
	ssize_t		jsonlen;
	int		fd, idx, keylen, len;
	char		json[2048], *key;

	if ((fd = open(path, O_RDONLY)) == -1) {
		warn("%s", path);
		return -1;
	}

	if ((jsonlen = read(fd, json, sizeof json - 1)) == -1) {
		warn("%s", path);
		close(fd);
		goto error;
	}

	json[jsonlen] = '\0';
	close(fd);

	if (sbk_jsmn_parse(json, jsonlen, tokens, nitems(tokens)) == -1) {
		warnx("%s: Cannot parse JSON data", path);
		goto error;
	}

	idx = sbk_jsmn_get_string(json, tokens, "key");
	if (idx == -1) {
		warnx("%s: Cannot find key", path);
		goto error;
	}

	key = json + tokens[idx].start;
	keylen = tokens[idx].end - tokens[idx].start;

	/* Write the key as an SQLite blob literal */
	len = snprintf(buf, bufsize, "x'%.*s'", keylen, key);
	if (len < 0 || (unsigned int)len >= bufsize) {
		warnx("%s: Cannot get key", path);
		goto error;
	}

	explicit_bzero(json, sizeof json);
	return 0;

error:
	explicit_bzero(json, sizeof json);
	explicit_bzero(buf, bufsize);
	return -1;
}

int
sbk_open(struct sbk_ctx **ctx, const char *dir)
{
	char	*dbfile, *errmsg, *keyfile;
	int	 ret;
	char	 key[128];

	dbfile = NULL;
	keyfile = NULL;
	ret = -1;

	if ((*ctx = calloc(1, sizeof **ctx)) == NULL) {
		warn(NULL);
		goto out;
	}

	RB_INIT(&(*ctx)->recipients);

	if (((*ctx)->dir = strdup(dir)) == NULL) {
		warn(NULL);
		goto out;
	}

	if (asprintf(&dbfile, "%s/sql/db.sqlite", dir) == -1) {
		warnx("asprintf() failed");
		dbfile = NULL;
		goto out;
	}

	if (asprintf(&keyfile, "%s/config.json", dir) == -1) {
		warnx("asprintf() failed");
		keyfile = NULL;
		goto out;
	}

	/*
	 * SQLite doesn't provide a useful error message if the database
	 * doesn't exist or can't be read
	 */
	if (access(dbfile, R_OK) == -1) {
		warn("%s", dbfile);
		goto out;
	}

	if (sbk_sqlite_open(&(*ctx)->db, dbfile, SQLITE_OPEN_READONLY) == -1)
		goto out;

	if (sbk_get_key(key, sizeof key, keyfile) == -1)
		goto out;

	if (sbk_sqlite_key((*ctx)->db, key) == -1) {
		explicit_bzero(key, sizeof key);
		goto out;
	}

	explicit_bzero(key, sizeof key);

	/* Verify key */
	if (sqlite3_exec((*ctx)->db, "SELECT count(*) FROM sqlite_master",
	    NULL, NULL, &errmsg) != SQLITE_OK) {
		warnx("Cannot verify key: %s", errmsg);
		sqlite3_free(errmsg);
		goto out;
	}

	if (((*ctx)->db_version = sbk_get_database_version(*ctx)) == -1)
		goto out;

	if ((*ctx)->db_version < 19) {
		warnx("Database version not supported (yet)");
		goto out;
	}

	ret = 0;

out:
	free(dbfile);
	free(keyfile);
	return ret;
}

void
sbk_close(struct sbk_ctx *ctx)
{
	if (ctx != NULL) {
		free(ctx->dir);
		sqlite3_close(ctx->db);
		sbk_free_recipient_tree(ctx);
		free(ctx);
	}
}
