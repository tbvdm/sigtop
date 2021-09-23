/*
 * Copyright (c) 2021 Tim van der Molen <tim@kariliq.nl>
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

#include <err.h>
#include <stdlib.h>
#include <unistd.h>

#include "sigtop.h"

int
cmd_check(int argc, char **argv)
{
	struct sbk_ctx	*ctx;
	char		*dir, **errors;
	int		 i, n;

	if (argc != 2)
		goto usage;

	dir = argv[1];

	if (unveil_signal_dir(dir) == -1)
		return 1;

	/* For SQLite/SQLCipher */
	if (unveil("/dev/urandom", "r") == -1)
		err(1, "unveil: /dev/urandom");

	if (pledge("stdio rpath wpath cpath flock", NULL) == -1)
		err(1, "pledge");

	if (sbk_open(&ctx, dir) == -1) {
		warnx("%s", sbk_error(ctx));
		sbk_close(ctx);
		return 1;
	}

	if ((n = sbk_check(ctx, &errors)) == -1) {
		warnx("%s", sbk_error(ctx));
		sbk_close(ctx);
		return 1;
	}

	if (n > 0) {
		for (i = 0; i < n; i++) {
			warnx("%s", errors[i]);
			free(errors[i]);
		}
		free(errors);
	}

	sbk_close(ctx);
	return (n == 0) ? 0 : 1;

usage:
	usage("check", "signal-directory");
}
