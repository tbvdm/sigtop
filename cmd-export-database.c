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
#include <fcntl.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

#include "sigtop.h"

static enum cmd_status cmd_export_database(int, char **);

const struct cmd_entry cmd_export_database_entry = {
	.name = "export-database",
	.alias = "db",
	.usage = "[-d signal-directory] file",
	.exec = cmd_export_database
};

static enum cmd_status
cmd_export_database(int argc, char **argv)
{
	struct sbk_ctx	*ctx;
	char		*db, *signaldir;
	int		 c, fd, ret;

	ctx = NULL;
	signaldir = NULL;

	while ((c = getopt(argc, argv, "d:")) != -1)
		switch (c) {
		case 'd':
			free(signaldir);
			if ((signaldir = strdup(optarg)) == NULL) {
				warn(NULL);
				goto error;
			}
			break;
		default:
			goto usage;
		}

	argc -= optind;
	argv += optind;

	if (argc != 1)
		goto usage;

	db = argv[0];

	if (signaldir == NULL)
		if ((signaldir = get_signal_dir()) == NULL)
			goto error;

	if (unveil_signal_dir(signaldir) == -1)
		goto error;

	/* For the export database and its temporary files */
	if (unveil_dirname(db, "rwc") == -1)
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

	/* Ensure the export database does not already exist */
	if ((fd = open(db, O_RDONLY | O_CREAT | O_EXCL, 0666)) == -1) {
		warn("%s", db);
		goto error;
	}
	close(fd);

	if (sbk_open(&ctx, signaldir) == -1) {
		warnx("%s", sbk_error(ctx));
		goto error;
	}

	if (sbk_write_database(ctx, db) == -1) {
		warnx("%s", sbk_error(ctx));
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
	free(signaldir);
	return ret;
}
