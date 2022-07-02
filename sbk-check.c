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

#include <limits.h>
#include <stdlib.h>
#include <string.h>

#include "sbk-internal.h"

static int
sbk_run_pragma(struct sbk_ctx *ctx, char ***errorsp, const char *pragma)
{
	sqlite3_stmt	 *stm;
	char		**errors, **newerrors;
	int		  i, max, n, newmax, ret;

	errors = NULL;
	max = n = 0;

	if (sbk_sqlite_prepare(ctx, ctx->db, &stm, pragma) == -1)
		goto error;

	while ((ret = sbk_sqlite_step(ctx, ctx->db, stm)) == SQLITE_ROW) {
		if (n == max) {
			if (max > INT_MAX - 100) {
				sbk_error_setx(ctx, "Too many errors");
				goto error;
			}
			newmax = max + 100;
			newerrors = reallocarray(errors, newmax,
			    sizeof *newerrors);
			if (newerrors == NULL) {
				sbk_error_set(ctx, NULL);
				goto error;
			}
			max = newmax;
			errors = newerrors;
		}
		if (sbk_sqlite_column_text_copy(ctx, &errors[n], stm, 0) == -1)
			goto error;
		if (errors[n] != NULL)
			n++;
	}

	if (ret != SQLITE_DONE)
		goto error;

	sqlite3_finalize(stm);
	*errorsp = errors;
	return n;

error:
	for (i = 0; i < n; i++)
		free(errors[i]);
	free(errors);
	sqlite3_finalize(stm);
	*errorsp = NULL;
	return -1;
}

int
sbk_check(struct sbk_ctx *ctx, char ***errorsp)
{
	char	**errors;
	int	  n;

	/*
	 * From the SQLCipher documentation: "The [cipher_integrity_check]
	 * PRAGMA will return one row per error condition. If no results are
	 * returned then the database was found to be externally consistent."
	 */

	n = sbk_run_pragma(ctx, &errors, "PRAGMA cipher_integrity_check");
	if (n != 0)
		goto out;

	/*
	 * From the SQLite documentation: "If the integrity_check pragma finds
	 * problems, strings are returned (as multiple rows with a single
	 * column per row) which describe the problems. [...] If pragma
	 * integrity_check finds no errors, a single row with the value 'ok' is
	 * returned."
	 */

	n = sbk_run_pragma(ctx, &errors, "PRAGMA integrity_check");
	if (n <= 0)
		goto out;

	if (n == 1 && strcmp(errors[0], "ok") == 0) {
		free(errors[0]);
		free(errors);
		errors = NULL;
		n = 0;
	}

out:
	*errorsp = errors;
	return n;
}
