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
#include <inttypes.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <unistd.h>

#include "sigtop.h"

enum {
	FORMAT_JSON,
	FORMAT_TEXT
};

static int
json_write_messages(FILE *fp, struct sbk_message_list *lst)
{
	struct sbk_message *msg;

	fputs("[\n", fp);
	SIMPLEQ_FOREACH(msg, lst, entries)
		fprintf(fp, "%s%s\n", msg->json,
		    (SIMPLEQ_NEXT(msg, entries) != NULL) ? "," : "");
	fputs("]\n", fp);

	return 0;
}

static void
text_write_date_field(FILE *fp, const char *field, int64_t date)
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

	fprintf(fp, "%s: %s, %d %s %d %02d:%02d:%02d %c%02ld%02ld\n",
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

static void
text_write_attachment_fields(FILE *fp, struct sbk_attachment_list *lst)
{
	struct sbk_attachment	*att;
	char			*content_type, *filename;

	TAILQ_FOREACH(att, lst, entries) {
		filename = (att->filename != NULL) ?
		    att->filename : "no filename";

		content_type = (att->content_type != NULL) ?
		    att->content_type : "unknown type";

		fprintf(fp, "Attachment: %s (%s, %" PRIu64 " bytes)\n",
		    filename, content_type, att->size);
	}
}

static int
text_write_messages(FILE *fp, struct sbk_message_list *lst)
{
	struct sbk_message *msg;

	SIMPLEQ_FOREACH(msg, lst, entries) {
		fprintf(fp, "Conversation: %s\n",
		    sbk_get_recipient_display_name(msg->conversation));

		if (sbk_is_outgoing_message(msg))
			fprintf(fp, "To: %s\n",
			    sbk_get_recipient_display_name(msg->conversation));
		else if (msg->source != NULL)
			fprintf(fp, "From: %s\n",
			    sbk_get_recipient_display_name(msg->source));

		text_write_date_field(fp, "Sent", msg->time_sent);

		if (!sbk_is_outgoing_message(msg))
			text_write_date_field(fp, "Received", msg->time_recv);

		if (msg->attachments != NULL)
			text_write_attachment_fields(fp, msg->attachments);

		if (msg->text != NULL)
			fprintf(fp, "\n%s\n", msg->text);

		putc('\n', fp);
	}

	return 0;
}

int
cmd_messages(int argc, char **argv)
{
	struct sbk_ctx		*ctx;
	struct sbk_message_list	*lst;
	FILE			*fp;
	char			*dir, *file;
	time_t			 max, min;
	int			 c, format, ret;

	format = FORMAT_TEXT;
	min = max = (time_t)-1;

	while ((c = getopt(argc, argv, "f:s:")) != -1)
		switch (c) {
		case 'f':
			if (strcmp(optarg, "json") == 0)
				format = FORMAT_JSON;
			else if (strcmp(optarg, "text") == 0)
				format = FORMAT_TEXT;
			else
				errx(1, "%s: invalid format", optarg);
			break;
		case 's':
			if (parse_time_interval(optarg, &min, &max) == -1)
				return -1;
			break;
		default:
			goto usage;
		}

	argc -= optind;
	argv += optind;

	switch (argc) {
	case 1:
		file = NULL;
		break;
	case 2:
		file = argv[1];
		if (unveil(file, "wc") == -1)
			err(1, "unveil");
		break;
	default:
		goto usage;
	}

	dir = argv[0];

	if (unveil_signal_dir(dir) == -1)
		return 1;

	/* For SQLite/SQLCipher */
	if (unveil("/dev/urandom", "r") == -1)
		err(1, "unveil");

	if (pledge("stdio rpath wpath cpath flock", NULL) == -1)
		err(1, "pledge");

	if (sbk_open(&ctx, dir) == -1) {
		warnx("%s", sbk_error(ctx));
		sbk_close(ctx);
		return 1;
	}

	if (min == (time_t)-1 && max == (time_t)-1)
		lst = sbk_get_all_messages(ctx);
	else if (min == (time_t)-1)
		lst = sbk_get_messages_sent_before(ctx, max);
	else if (max == (time_t)-1)
		lst = sbk_get_messages_sent_after(ctx, min);
	else
		lst = sbk_get_messages_sent_between(ctx, min, max);

	if (lst == NULL) {
		warnx("%s", sbk_error(ctx));
		sbk_close(ctx);
		return -1;
	}

	if (file == NULL)
		fp = stdout;
	else if ((fp = fopen(file, "wx")) == NULL) {
		warn("fopen: %s", file);
		sbk_free_message_list(lst);
		sbk_close(ctx);
		return 1;
	}

	switch (format) {
	case FORMAT_JSON:
		ret = json_write_messages(fp, lst);
		break;
	case FORMAT_TEXT:
		ret = text_write_messages(fp, lst);
		break;
	}

	if (fp != stdout)
		fclose(fp);

	sbk_free_message_list(lst);
	sbk_close(ctx);
	return (ret == 0) ? 0 : 1;

usage:
	usage("messages", "[-f format] [-s interval] signal-directory [file]");
}
