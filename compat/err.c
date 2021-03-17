/*
 * Copyright (c) 2011 Tim van der Molen <tim@kariliq.nl>
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

#include "../config.h"

#ifndef HAVE_ERR

#include <errno.h>
#include <stdarg.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "../compat.h"

void
err(int ret, const char *fmt, ...)
{
	va_list ap;

	va_start(ap, fmt);
	verr(ret, fmt, ap);
}

void
errc(int ret, int errnum, const char *fmt, ...)
{
	va_list ap;

	va_start(ap, fmt);
	verrc(ret, errnum, fmt, ap);
}

void
errx(int ret, const char *fmt, ...)
{
	va_list ap;

	va_start(ap, fmt);
	verrx(ret, fmt, ap);
}

void
verr(int ret, const char *fmt, va_list ap)
{
	vwarn(fmt, ap);
	exit(ret);
}

void
verrc(int ret, int errnum, const char *fmt, va_list ap)
{
	vwarnc(errnum, fmt, ap);
	exit(ret);
}

void
verrx(int ret, const char *fmt, va_list ap)
{
	vwarnx(fmt, ap);
	exit(ret);
}

void
warn(const char *fmt, ...)
{
	va_list ap;

	va_start(ap, fmt);
	vwarn(fmt, ap);
	va_end(ap);
}

void
warnc(int errnum, const char *fmt, ...)
{
	va_list ap;

	va_start(ap, fmt);
	vwarnc(errnum, fmt, ap);
	va_end(ap);
}

void
warnx(const char *fmt, ...)
{
	va_list ap;

	va_start(ap, fmt);
	vwarnx(fmt, ap);
	va_end(ap);
}

void
vwarn(const char *fmt, va_list ap)
{
	int saved_errno;

	saved_errno = errno;
	vwarnc(errno, fmt, ap);
	errno = saved_errno;
}

void
vwarnc(int errnum, const char *fmt, va_list ap)
{
	fputs(getprogname(), stderr);
	fputs(": ", stderr);

	if (fmt != NULL) {
		vfprintf(stderr, fmt, ap);
		fputs(": ", stderr);
	}

	fputs(strerror(errnum), stderr);
	putc('\n', stderr);
}

void
vwarnx(const char *fmt, va_list ap)
{
	fputs(getprogname(), stderr);
	fputs(": ", stderr);

	if (fmt != NULL)
		vfprintf(stderr, fmt, ap);

	putc('\n', stderr);
}

#endif
