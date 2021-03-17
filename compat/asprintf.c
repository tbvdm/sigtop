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

#ifndef HAVE_ASPRINTF

#include <limits.h>
#include <stdarg.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>

int
vasprintf(char **buf, const char *fmt, va_list ap)
{
	va_list	ap2;
	int	len;

	va_copy(ap2, ap);
	len = vsnprintf(NULL, 0, fmt, ap2);
	va_end(ap2);

	if (len < 0)
		return -1;

#if SIZE_MAX <= INT_MAX
	if (len >= SIZE_MAX)
		return -1;
#endif

	if ((*buf = malloc((size_t)len + 1)) == NULL)
		return -1;

	if (vsnprintf(*buf, (size_t)len + 1, fmt, ap) != len) {
		free(*buf);
		return -1;
	}

	return len;
}

int
asprintf(char **buf, const char *fmt, ...)
{
	va_list	ap;
	int	len;

	va_start(ap, fmt);
	len = vasprintf(buf, fmt, ap);
	va_end(ap);
	return len;
}

#endif
