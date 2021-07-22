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

#include <sys/stat.h>
#include <sys/types.h>

#include <errno.h>
#include <fcntl.h>
#include <limits.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <unistd.h>

#include "sigtop.h"

enum mode {
	MODE_COPY,
	MODE_LINK,
	MODE_SYMLINK
};

/*
 * Check if a file with the specified name exists. If so, replace the name with
 * a new, unique name. Given a name of the form "base[.ext]", the new name is
 * of the form "base-n[.ext]" where 1 < n < 1000.
 */
static int
get_unique_filename(int dfd, char **name)
{
	struct stat	 st;
	char		*newname;
	const char	*ext;
	size_t		 baselen, namelen, size;
	int		 i;

	if (fstatat(dfd, *name, &st, AT_SYMLINK_NOFOLLOW) == -1) {
		if (errno != ENOENT) {
			warn("fstatat: %s", *name);
			return -1;
		}
		return 0;
	}

	namelen = strlen(*name);

	ext = strrchr(*name, '.');
	if (ext != NULL && ext != *name && ext[1] != '\0')
		baselen = ext - *name;
	else {
		baselen = namelen;
		ext = "";
	}

	if (namelen > SIZE_MAX - 5 || baselen > INT_MAX) {
		warnx("Attachment filename too long");
		return -1;
	}

	/* 4 for the "-n" affix and 1 for the NUL */
	size = namelen + 5;
	if ((newname = malloc(size)) == NULL) {
		warn(NULL);
		return -1;
	}

	for (i = 2; i < 1000; i++) {
		snprintf(newname, size, "%.*s-%d%s", (int)baselen, *name, i,
		    ext);
		if (fstatat(dfd, newname, &st, AT_SYMLINK_NOFOLLOW) == -1) {
			if (errno != ENOENT) {
				warn("fstatat: %s", newname);
				free(newname);
				return -1;
			}
			free(*name);
			*name = newname;
			return 0;
		}
	}

	warnx("%s: Cannot generate unique filename", *name);
	free(newname);
	return -1;
}

static char *
get_filename(int dfd, struct sbk_attachment *att)
{
	char		*c, *name;
	const char	*ext;
	struct tm	*tm;
	time_t		 tt;
	char		 base[32];

	if (att->filename != NULL && *att->filename != '\0') {
		if (strcmp(att->filename, ".") == 0)
			name = strdup("_");
		else if (strcmp(att->filename, "..") == 0)
			name = strdup("__");
		else {
			if ((name = strdup(att->filename)) != NULL) {
				c = name;
				while ((c = strchr(c, '/')) != NULL)
					*c++ = '_';
			}
		}
		if (name == NULL) {
			warn(NULL);
			return NULL;
		}
	} else {
		tt = att->time_sent / 1000;
		if ((tm = localtime(&tt)) == NULL) {
			warnx("localtime() failed");
			return NULL;
		}
		snprintf(base, sizeof base,
		    "attachment-%d-%02d-%02d-%02d-%02d-%02d",
		    tm->tm_year + 1900,
		    tm->tm_mon + 1,
		    tm->tm_mday,
		    tm->tm_hour,
		    tm->tm_min,
		    tm->tm_sec);
		if (att->content_type == NULL)
			ext = NULL;
		else
			ext = mime_get_extension(att->content_type);
		if (ext == NULL) {
			if ((name = strdup(base)) == NULL) {
				warn(NULL);
				return NULL;
			}
		} else {
			if (asprintf(&name, "%s.%s", base, ext) == -1) {
				warnx("asprintf() failed");
				return NULL;
			}
		}
	}

	if (get_unique_filename(dfd, &name) == -1) {
		free(name);
		return NULL;
	}

	return name;
}

#define COPY_BUFSIZE (1024 * 1024)

static int
copy_attachment(const char *src, int dfd, const char *dst)
{
	char	*buf;
	int	 n, ret, rfd, wfd;

	ret = rfd = wfd = -1;

	if ((buf = malloc(COPY_BUFSIZE)) == NULL) {
		warn(NULL);
		goto out;
	}
	if ((rfd = open(src, O_RDONLY)) == -1) {
		warn("open: %s", src);
		goto out;
	}
	if ((wfd = openat(dfd, dst, O_WRONLY | O_CREAT | O_EXCL, 0666)) ==
	    -1) {
		warn("openat: %s", dst);
		goto out;
	}
	while ((n = read(rfd, buf, COPY_BUFSIZE)) > 0)
		if (write(wfd, buf, n) != n) {
			warn("write: %s", dst);
			goto out;
		}
	if (n < 0) {
		warn("read: %s", src);
		goto out;
	}

	ret = 0;

out:
	if (rfd != -1)
		close(rfd);
	if (wfd != -1)
		close(wfd);
	free(buf);
	return ret;
}

static int
process_attachments(struct sbk_ctx *ctx, const char *dir,
    struct sbk_attachment_list *lst, enum mode mode)
{
	struct sbk_attachment	*att;
	char			*dst, *src;
	int			 dfd, ret;

	if ((dfd = open(dir, O_RDONLY | O_DIRECTORY)) == -1) {
		warn("open: %s", dir);
		return -1;
	}

	ret = 0;

	TAILQ_FOREACH(att, lst, entries) {
		if (att->path == NULL)
			continue;
		if ((src = sbk_get_attachment_path(ctx, att)) == NULL) {
			warnx("Cannot get attachment path: %s",
			    sbk_error(ctx));
			ret = -1;
			continue;
		}
		if (access(src, F_OK) == -1) {
			warn("access: %s", src);
			free(src);
			ret = -1;
			continue;
		}
		if ((dst = get_filename(dfd, att)) == NULL) {
			free(src);
			ret = -1;
			continue;
		}
		switch (mode) {
		case MODE_COPY:
			if (copy_attachment(src, dfd, dst) == -1)
				ret = -1;
			break;
		case MODE_LINK:
			if (linkat(AT_FDCWD, src, dfd, dst, 0) == -1) {
				warn("linkat: %s", dst);
				ret = -1;
			}
			break;
		case MODE_SYMLINK:
			if (symlinkat(src, dfd, dst) == -1) {
				warn("symlinkat: %s", dst);
				ret = -1;
			}
			break;
		}
		free(src);
		free(dst);
	}

	close(dfd);
	return ret;
}

int
cmd_attachments(int argc, char **argv)
{
	struct sbk_ctx			*ctx;
	struct sbk_attachment_list	*lst;
	char				*signaldir;
	const char			*outdir;
	int				 c, ret;
	enum mode			 mode;

	mode = MODE_COPY;

	while ((c = getopt(argc, argv, "Ll")) != -1)
		switch (c) {
		case 'L':
			mode = MODE_LINK;
			break;
		case 'l':
			mode = MODE_SYMLINK;
			break;
		default:
			goto usage;
		}

	argc -= optind;
	argv += optind;

	switch (argc) {
	case 1:
		signaldir = argv[0];
		outdir = ".";
		break;
	case 2:
		signaldir = argv[0];
		outdir = argv[1];
		if (mkdir(outdir, 0777) == -1 && errno != EEXIST)
			err(1, "mkdir: %s", outdir);
		break;
	default:
		goto usage;
	}

	if (unveil(signaldir, "r") == -1)
		err(1, "unveil");

	if (unveil(outdir, "rwc") == -1)
		err(1, "unveil");

	/* For SQLite/SQLCipher */
	if (unveil("/dev/urandom", "r") == -1)
		err(1, "unveil");

	if (pledge("stdio rpath wpath cpath flock", NULL) == -1)
		err(1, "pledge");

	lst = NULL;
	ret = 1;

	if (sbk_open(&ctx, signaldir) == -1) {
		warnx("%s", sbk_error(ctx));
		goto out;
	}

	if ((lst = sbk_get_all_attachments(ctx)) == NULL) {
		warnx("%s", sbk_error(ctx));
		goto out;
	}

	if (process_attachments(ctx, outdir, lst, mode) == -1)
		goto out;

	ret = 0;

out:
	sbk_free_attachment_list(lst);
	sbk_close(ctx);
	return ret;

usage:
	usage("attachments", "[-Ll] signal-directory [directory]");
}
