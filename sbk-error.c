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

#include <errno.h>
#include <stdarg.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "sbk-internal.h"

void
sbk_error_clear(struct sbk_ctx *ctx)
{
	free(ctx->error);
	ctx->error = NULL;
}

void
sbk_error_set(struct sbk_ctx *ctx, const char *fmt, ...)
{
	va_list	 ap;
	char	*errmsg, *msg;
	int	 saved_errno;

	va_start(ap, fmt);
	saved_errno = errno;
	sbk_error_clear(ctx);
	errmsg = strerror(saved_errno);

	if (fmt == NULL || vasprintf(&msg, fmt, ap) == -1)
		ctx->error = strdup(errmsg);
	else if (asprintf(&ctx->error, "%s: %s", msg, errmsg) == -1)
		ctx->error = msg;
	else
		free(msg);

	errno = saved_errno;
	va_end(ap);
}

void
sbk_error_setx(struct sbk_ctx *ctx, const char *fmt, ...)
{
	va_list ap;

	va_start(ap, fmt);
	sbk_error_clear(ctx);

	if (fmt == NULL || vasprintf(&ctx->error, fmt, ap) == -1)
		ctx->error = NULL;

	va_end(ap);
}

void
sbk_error_sqlite_vsetd(struct sbk_ctx *ctx, sqlite3 *db, const char *fmt,
    va_list ap)
{
	const char	*errmsg;
	char		*msg;

	sbk_error_clear(ctx);
	errmsg = sqlite3_errmsg(db);

	if (fmt == NULL || vasprintf(&msg, fmt, ap) == -1)
		ctx->error = strdup(errmsg);
	else if (asprintf(&ctx->error, "%s: %s", msg, errmsg) == -1)
		ctx->error = msg;
	else
		free(msg);
}

void
sbk_error_sqlite_setd(struct sbk_ctx *ctx, sqlite3 *db, const char *fmt, ...)
{
	va_list ap;

	va_start(ap, fmt);
	sbk_error_sqlite_vsetd(ctx, db, fmt, ap);
	va_end(ap);
}

void
sbk_error_sqlite_set(struct sbk_ctx *ctx, const char *fmt, ...)
{
	va_list ap;

	va_start(ap, fmt);
	sbk_error_sqlite_vsetd(ctx, ctx->db, fmt, ap);
	va_end(ap);
}

const char *
sbk_error(struct sbk_ctx *ctx)
{
	if (ctx == NULL)
		return strerror(ENOMEM);
	else
		return (ctx->error != NULL) ? ctx->error : "Unknown error";
}
