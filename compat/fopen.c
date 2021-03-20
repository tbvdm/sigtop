/*
 * Copyright (c) 2019 Tim van der Molen <tim@kariliq.nl>
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

#ifndef HAVE_FOPEN_X_MODE

#include <errno.h>
#include <fcntl.h>
#include <stdio.h>
#include <unistd.h>

/*
 * An fopen(3) implementation that supports the "e" and "x" mode extensions
 */
FILE *
xfopen(const char *path, const char *extmode)
{
	FILE		*fp;
	char		 mode[64];
	unsigned int	 i, j;
	int		 fd, flags, saved_errno;

	flags = 0;

	for (i = j = 0; extmode[i] != '\0'; i++) {
		switch (extmode[i]) {
		case 'r':
			flags = O_RDONLY;
			break;
		case 'w':
			flags = O_WRONLY | O_CREAT | O_TRUNC;
			break;
		case 'a':
			flags = O_WRONLY | O_CREAT | O_APPEND;
			break;
		case '+':
			flags &= ~(O_RDONLY | O_WRONLY); /* Satisfy POSIX */
			flags |= O_RDWR;
			break;
		case 'b':
			break;
		case 'e':
			flags |= O_CLOEXEC;
			continue;
		case 'x':
			if (flags & O_CREAT)
				flags |= O_EXCL;
			continue;
		default:
			errno = EINVAL;
			return NULL;
		}

		if (j == sizeof mode - 1) {
			errno = EINVAL;
			return NULL;
		}

		mode[j++] = extmode[i];
	}

	mode[j] = '\0';

	if ((fd = open(path, flags, 0666)) == -1)
		return NULL;

	if ((fp = fdopen(fd, mode)) == NULL) {
		saved_errno = errno;
		close(fd);
		errno = saved_errno;
		return NULL;
	}

	return fp;
}

#endif
