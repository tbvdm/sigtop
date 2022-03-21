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

static char *
get_xdg_config_dir(const char *home)
{
	char *config, *dir;

	config = getenv("XDG_CONFIG_HOME");
	if (config != NULL && config[0] != '\0') {
		if ((dir = strdup(config)) == NULL)
			warn(NULL);
		return dir;
	}

	if (asprintf(&dir, "%s/.config", home) == -1) {
		warnx("asprintf() failed");
		return NULL;
	}

	return dir;
}

char *
get_signal_dir(void)
{
	struct stat	 st;
	char		*config, *dir, *home, *snap;

	if ((home = get_home_dir()) == NULL)
		return NULL;

	if ((config = get_xdg_config_dir(home)) == NULL)
		return NULL;

	if (asprintf(&dir, "%s/Signal", config) == -1) {
		warnx("asprintf() failed");
		free(config);
		return NULL;
	}

	free(config);

	/*
	 * If the default Signal Desktop directory doesn't exist, try the one
	 * from the unofficial snap
	 */

	if (lstat(dir, &st) == 0)
		return dir;
	else if (errno != ENOENT && errno != ENOTDIR) {
		warn("%s", dir);
		free(dir);
		return NULL;
	}

	if (asprintf(&snap, "%s/snap/signal-desktop/current/.config/Signal",
	    home) == -1) {
		warnx("asprintf() failed");
		free(dir);
		return NULL;
	}

	if (lstat(snap, &st) == 0) {
		free(dir);
		dir = snap;
	} else if (errno != ENOENT && errno != ENOTDIR) {
		warn("%s", snap);
		free(dir);
		free(snap);
		return NULL;
	} else {
		free(snap);
	}

	return dir;
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
	char *dbdir;

	if (unveil(dir, "r") == -1) {
		warn("unveil: %s", dir);
		return -1;
	}

	/*
	 * SQLCipher needs to create the sql/db.sqlite-{shm,wal} files if they
	 * don't exist already
	 */

	if (asprintf(&dbdir, "%s/sql", dir) == -1) {
		warnx("asprintf() failed");
		return -1;
	}

	if (unveil(dbdir, "rwc") == -1) {
		warn("unveil: %s", dbdir);
		free(dbdir);
		return -1;
	}

	free(dbdir);
	return 0;
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

	for (i = 0; i < nitems(commands); i++) {
		if (strcmp(argv[0], commands[i]->name) == 0 ||
		    strcmp(argv[0], commands[i]->alias) == 0) {
			cmd = commands[i];
			break;
		}
		if (commands[i]->oldname != NULL &&
		    strcmp(argv[0], commands[i]->oldname) == 0)
			errx(1, "Command names and options have changed; see "
			    "the manual page");
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
