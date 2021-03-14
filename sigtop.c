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

#include <sys/types.h>

#include <err.h>
#include <libgen.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <unistd.h>

#include "sigtop.h"

__dead void
usage(void)
{
	fprintf(stderr, "usage: %s database keyfile\n", getprogname());
	exit(1);
}

static void
print_date_field(const char *field, int64_t date)
{
	const char	*days[] = {
	    "Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat" };

	const char	*months[] = {
	    "Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep",
	    "Oct", "Nov", "Dec" };

	struct tm	*tm;
	time_t		 tt;

	tt = date / 1000;

	if ((tm = localtime(&tt)) == NULL) {
		warnx("localtime() failed");
		return;
	}

	printf("%s: %s, %d %s %d %02d:%02d:%02d %c%02ld%02ld\n",
	    field,
	    days[tm->tm_wday],
	    tm->tm_mday,
	    months[tm->tm_mon],
	    tm->tm_year + 1900,
	    tm->tm_hour,
	    tm->tm_min,
	    tm->tm_sec,
	    (tm->tm_gmtoff < 0) ? '-' : '+',
	    labs(tm->tm_gmtoff) / 3600,
	    labs(tm->tm_gmtoff) % 3600 / 60);
}

int
print_messages(struct sbk_ctx *ctx)
{
	struct sbk_message_list	*lst;
	struct sbk_message	*msg;

	if ((lst = sbk_get_all_messages(ctx)) == NULL) {
		warnx("%s", sbk_error(ctx));
		return -1;
	}

	SIMPLEQ_FOREACH(msg, lst, entries) {
		printf("Conversation: %s\n",
		    sbk_get_recipient_display_name(msg->conversation));

		if (sbk_is_outgoing_message(msg))
			printf("To: %s\n",
			    sbk_get_recipient_display_name(msg->conversation));
		else
			printf("From: %s\n",
			    sbk_get_recipient_display_name(msg->source));

		print_date_field("Sent", msg->time_sent);

		if (!sbk_is_outgoing_message(msg))
			print_date_field("Received", msg->time_recv);

		if (msg->text != NULL)
			printf("\n%s\n", msg->text);

		putchar('\n');
	}

	sbk_free_message_list(lst);
	return 0;
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
		warn("unveil");
		free(tmp);
		return -1;
	}

	free(tmp);
	return 0;
}

int
main(int argc, char **argv)
{
	struct sbk_ctx	*ctx;
	char		*db, *keyfile;

	if (argc != 3)
		usage();

	db = argv[1];
	keyfile = argv[2];

	/* For the database and its temporary files */
	if (unveil_dirname(db, "r") == -1)
		return 1;

	if (unveil(keyfile, "r") == -1)
		err(1, "unveil: %s", keyfile);

	/* For SQLite/SQLCipher */
	if (unveil("/dev/urandom", "r") == -1)
		err(1, "unveil");

	/* For SQLite/SQLCipher */
	if (unveil("/tmp", "rwc") == -1)
		err(1, "unveil");

	if (unveil("/etc/localtime", "r") == -1)
		err(1, "unveil");

	if (unveil("/usr/share/zoneinfo", "r") == -1)
		err(1, "unveil");

	if (unveil(NULL, NULL) == -1)
		err(1, "unveil");

	if (sbk_open(&ctx, db, keyfile) == -1) {
		warnx("%s: %s", db, sbk_error(ctx));
		sbk_close(ctx);
		return 1;
	}

	if (print_messages(ctx) == -1) {
		sbk_close(ctx);
		return 1;
	}

	sbk_close(ctx);
	return 0;
}
