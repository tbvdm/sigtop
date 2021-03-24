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

#include "config.h"

#include <fcntl.h>
#include <unistd.h>

#include "sigtop.h"

int
cmd_sqlite(int argc, char **argv)
{
	struct sbk_ctx	*ctx;
	char		*db, *dir;
	int		 fd, ret;

	if (argc != 3)
		goto usage;

	dir = argv[1];
	db = argv[2];

	if (unveil(dir, "r") == -1)
		err(1, "unveil");

	/* For SQLite/SQLCipher */
	if (unveil("/dev/urandom", "r") == -1)
		err(1, "unveil");

	/* For SQLite/SQLCipher */
	if (unveil("/tmp", "rwc") == -1)
		err(1, "unveil");

	/* For the export database and its temporary files */
	if (unveil_dirname(db, "rwc") == -1)
		return 1;

	if (unveil(NULL, NULL) == -1)
		err(1, "unveil");

	/* Ensure the export database does not already exist */
	if ((fd = open(db, O_RDONLY | O_CREAT | O_EXCL, 0666)) == -1)
		err(1, "%s", db);

	close(fd);

	if (sbk_open(&ctx, dir) == -1) {
		warnx("%s", sbk_error(ctx));
		sbk_close(ctx);
		return 1;
	}

	if ((ret = sbk_write_database(ctx, db)) == -1)
		warnx("%s", sbk_error(ctx));

	sbk_close(ctx);
	return (ret == 0) ? 0 : 1;

usage:
	usage("sqlite", "signal-directory file");
}
