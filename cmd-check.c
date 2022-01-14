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

static enum cmd_status cmd_check(int, char **);

const struct cmd_entry cmd_check_entry = {
	.name = "check",
	.alias = "chk",
	.usage = "signal-directory",
	.exec = cmd_check
};

static enum cmd_status
cmd_check(int argc, char **argv)
{
	struct sbk_ctx	 *ctx;
	char		**errors, *signaldir;
	int		  i, n, ret;

	ctx = NULL;

	if (argc != 2)
		goto usage;

	signaldir = argv[1];

	if (unveil_signal_dir(signaldir) == -1)
		goto error;

	/* For SQLite/SQLCipher */
	if (unveil("/dev/urandom", "r") == -1) {
		warn("unveil: /dev/urandom");
		goto error;
	}

	if (pledge("stdio rpath wpath cpath flock", NULL) == -1) {
		warn("pledge");
		goto error;
	}

	if (sbk_open(&ctx, signaldir) == -1) {
		warnx("%s", sbk_error(ctx));
		goto error;
	}

	if ((n = sbk_check(ctx, &errors)) == -1) {
		warnx("%s", sbk_error(ctx));
		goto error;
	}

	if (n > 0) {
		for (i = 0; i < n; i++) {
			warnx("%s", errors[i]);
			free(errors[i]);
		}
		free(errors);
		goto error;
	}

	ret = CMD_OK;
	goto out;

error:
	ret = CMD_ERROR;
	goto out;

usage:
	ret = CMD_USAGE;

out:
	sbk_close(ctx);
	return ret;
}
