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

#include <sys/stat.h>
#include <sys/types.h>

#include <ctype.h>
#include <err.h>
#include <errno.h>
#include <libgen.h>
#include <pwd.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <unistd.h>

#include "sigtop.h"

extern const struct cmd_entry cmd_check_database_entry;
extern const struct cmd_entry cmd_export_attachments_entry;
extern const struct cmd_entry cmd_export_database_entry;
extern const struct cmd_entry cmd_export_messages_entry;

static const struct cmd_entry *commands[] = {
	&cmd_check_database_entry,
	&cmd_export_attachments_entry,
	&cmd_export_database_entry,
	&cmd_export_messages_entry,
};

__dead static void
usage(const char *cmd, const char *args)
{
	fprintf(stderr, "usage: %s %s %s\n", getprogname(), cmd, args);
	exit(1);
}

static char *
get_home_dir(void)
{
	struct passwd	*pw;
	char		*home;

	home = getenv("HOME");
	if (home != NULL && home[0] != '\0')
		return home;

	errno = 0;
	if ((pw = getpwuid(getuid())) == NULL) {
		if (errno)
			warn("getpwuid");
		else
			warnx("Unknown user");
		return NULL;
	}

	return pw->pw_dir;
}

static int
try_signal_dir(const char *dir)
{
	struct stat st;

	if (lstat(dir, &st) == 0)
		return 1;
	else if (errno == ENOENT || errno == ENOTDIR)
		return 0;
	else {
		warn("%s", dir);
		return -1;
	}
}

static int
try_default_signal_dir(char **defdir, const char *homedir)
{
	char	*configdir;
	int	 ret;

	configdir = getenv("XDG_CONFIG_HOME");
	if (configdir != NULL && configdir[0] != '\0')
		ret = asprintf(defdir, "%s/Signal", configdir);
	else
		ret = asprintf(defdir, "%s/.config/Signal", homedir);

	if (ret == -1) {
		warnx("asprintf() failed");
		*defdir = NULL;
		return -1;
	}

	if ((ret = try_signal_dir(*defdir)) == -1) {
		free(*defdir);
		*defdir = NULL;
	}

	return ret;
}

static int
try_alternative_signal_dir(char **altdir, const char *homedir,
    const char *subdir)
{
	int ret;

	if (asprintf(altdir, "%s/%s", homedir, subdir) == -1) {
		warnx("asprintf() failed");
		*altdir = NULL;
		return -1;
	}

	if ((ret = try_signal_dir(*altdir)) != 1) {
		free(*altdir);
		*altdir = NULL;
	}

	return ret;
}

char *
get_signal_dir(void)
{
	char *altdir, *defdir, *homedir;

	if ((homedir = get_home_dir()) == NULL)
		return NULL;

	if (try_default_signal_dir(&defdir, homedir) != 0)
		return defdir;

	/* Snap */
	if (try_alternative_signal_dir(&altdir, homedir,
	    "snap/signal-desktop/current/.config/Signal") != 0) {
		free(defdir);
		return altdir;
	}

	/* Flatpak */
	if (try_alternative_signal_dir(&altdir, homedir,
	    ".var/app/org.signal.Signal/config/Signal") != 0) {
		free(defdir);
		return altdir;
	}

	return defdir;
}

int
unveil_dirname(const char *path, const char *perms)
{
	char *dir, *tmp;

	if ((tmp = strdup(path)) == NULL) {
		warn(NULL);
		return -1;
	}

	if ((dir = dirname(tmp)) == NULL) {
		warnx("dirname() failed");
		free(tmp);
		return -1;
	}

	if (unveil(dir, perms) == -1) {
		warn("unveil: %s", dir);
		free(tmp);
		return -1;
	}

	free(tmp);
	return 0;
}

int
unveil_signal_dir(const char *dir)
{
	char	*shm, *wal;
	int	 ret;

	shm = wal = NULL;
	ret = -1;

	if (unveil(dir, "r") == -1) {
		warn("unveil: %s", dir);
		goto out;
	}

	/*
	 * SQLCipher needs to create the sql/db.sqlite-{shm,wal} files if they
	 * don't exist already
	 */

	if (asprintf(&shm, "%s/sql/db.sqlite-shm", dir) == -1) {
		warnx("asprintf() failed");
		shm = NULL;
		goto out;
	}

	if (asprintf(&wal, "%s/sql/db.sqlite-wal", dir) == -1) {
		warnx("asprintf() failed");
		wal = NULL;
		goto out;
	}

	if (unveil(shm, "rwc") == -1) {
		warn("unveil: %s", shm);
		goto out;
	}

	if (unveil(wal, "rwc") == -1) {
		warn("unveil: %s", wal);
		goto out;
	}

	ret = 0;

out:
	free(shm);
	free(wal);
	return ret;
}

static int
parse_time(const char *str, time_t *tt)
{
	struct tm	 tm;
	char		*c;

	if (*str == '\0') {
		*tt = (time_t)-1;
		return 0;
	}

	memset(&tm, 0, sizeof tm);
	c = strptime(str, "%Y-%m-%dT%H:%M:%S", &tm);

	if (c == NULL || *c != '\0') {
		warnx("%s: Invalid time specification", str);
		return -1;
	}

	tm.tm_isdst = -1;

	if ((*tt = mktime(&tm)) < 0) {
		warnx("mktime() failed");
		return -1;
	}

	return 0;
}

int
parse_time_interval(char *str, time_t *min, time_t *max)
{
	char *maxstr, *minstr, *sep;

	if ((sep = strchr(str, ',')) == NULL) {
		warnx("%s: Missing separator in time interval", str);
		return -1;
	}

	*sep = '\0';
	minstr = str;
	maxstr = sep + 1;

	if (parse_time(minstr, min) == -1 || parse_time(maxstr, max) == -1)
		return -1;

	if (*max != (time_t)-1 && *min > *max) {
		warnx("%s is later than %s", minstr, maxstr);
		return -1;
	}

	return 0;
}

void
sanitise_filename(char *name)
{
	char *c;

	if (strcmp(name, ".") == 0) {
		name[0] = '_';
		return;
	}

	if (strcmp(name, "..") == 0) {
		name[0] = name[1] = '_';
		return;
	}

	for (c = name; *c != '\0'; c++)
		if (*c == '/' || iscntrl((unsigned char)*c))
			*c = '_';
}

char *
get_recipient_filename(struct sbk_recipient *rcp, const char *ext)
{
	char		*fname;
	const char	*detail, *name;
	int		 ret;

	name = sbk_get_recipient_display_name(rcp);

	if (rcp->type == SBK_GROUP)
		detail = "group";
	else
		detail = rcp->contact->phone;

	if (ext == NULL)
		ext = "";

	if (detail != NULL)
		ret = asprintf(&fname, "%s (%s)%s", name, detail, ext);
	else
		ret = asprintf(&fname, "%s%s", name, ext);

	if (ret == -1) {
		warnx("asprintf() failed");
		return NULL;
	}

	sanitise_filename(fname);
	return fname;
}

int
main(int argc, char **argv)
{
	const struct cmd_entry	*cmd;
	size_t			 i;

	if (argc < 2)
		usage("command", "[argument ...]");

	argc--;
	argv++;
	cmd = NULL;

	for (i = 0; i < nitems(commands); i++)
		if (strcmp(argv[0], commands[i]->name) == 0 ||
		    strcmp(argv[0], commands[i]->alias) == 0) {
			cmd = commands[i];
			break;
		}

	if (cmd == NULL)
		errx(1, "%s: Invalid command", argv[0]);

	switch (cmd->exec(argc, argv)) {
	case CMD_OK:
		return 0;
	case CMD_ERROR:
		return 1;
	case CMD_USAGE:
		usage(cmd->name, cmd->usage);
	}
}
