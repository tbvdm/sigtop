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

#include <sys/types.h>

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

static enum cmd_status cmd_export_messages(int, char **);

const struct cmd_entry cmd_export_messages_entry = {
	.name = "export-messages",
	.alias = "msg",
	.usage = "[-d signal-directory] [-f format] [-s interval] [file]",
	.oldname = "messages",
	.exec = cmd_export_messages
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

#ifdef HAVE_TM_GMTOFF
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
#else
	fprintf(fp, "%s: %s, %d %s %d %02d:%02d:%02d\n",
	    field,
	    days[tm->tm_wday],
	    tm->tm_mday,
	    months[tm->tm_mon],
	    tm->tm_year + 1900,
	    tm->tm_hour,
	    tm->tm_min,
	    tm->tm_sec);
#endif
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

static void
text_write_reaction_fields(FILE *fp, struct sbk_reaction_list *lst)
{
	struct sbk_reaction *rct;

	SIMPLEQ_FOREACH(rct, lst, entries)
		fprintf(fp, "Reaction: %s from %s\n", rct->emoji,
		    sbk_get_recipient_display_name(rct->recipient));
}

static int
text_write_messages(FILE *fp, struct sbk_message_list *lst)
{
	struct sbk_message *msg;

	SIMPLEQ_FOREACH(msg, lst, entries) {
		fprintf(fp, "Conversation: %s\n",
		    sbk_get_recipient_display_name(msg->conversation));

		fprintf(fp, "Type: %s\n",
		    (msg->type != NULL) ? msg->type : "Unknown");

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

		if (msg->reactions != NULL)
			text_write_reaction_fields(fp, msg->reactions);

		if (msg->text != NULL)
			fprintf(fp, "\n%s\n", msg->text);

		putc('\n', fp);
	}

	return 0;
}

static enum cmd_status
cmd_export_messages(int argc, char **argv)
{
	struct sbk_ctx		*ctx;
	struct sbk_message_list	*lst;
	FILE			*fp;
	char			*file, *signaldir;
	time_t			 max, min;
	int			 c, format, ret;

	ctx = NULL;
	lst = NULL;
	signaldir = NULL;
	format = FORMAT_TEXT;
	min = max = (time_t)-1;

	while ((c = getopt(argc, argv, "d:f:s:")) != -1)
		switch (c) {
		case 'd':
			free(signaldir);
			if ((signaldir = strdup(optarg)) == NULL) {
				warn(NULL);
				goto error;
			}
			break;
		case 'f':
			if (strcmp(optarg, "json") == 0)
				format = FORMAT_JSON;
			else if (strcmp(optarg, "text") == 0)
				format = FORMAT_TEXT;
			else {
				warnx("%s: Invalid format", optarg);
				goto error;
			}
			break;
		case 's':
			if (parse_time_interval(optarg, &min, &max) == -1)
				goto error;
			break;
		default:
			goto usage;
		}

	argc -= optind;
	argv += optind;

	switch (argc) {
	case 0:
		file = NULL;
		break;
	case 1:
		file = argv[0];
		break;
	default:
		goto usage;
	}

	if (signaldir == NULL)
		if ((signaldir = get_signal_dir()) == NULL)
			goto error;

	if (unveil_signal_dir(signaldir) == -1)
		goto error;

	if (file != NULL)
		if (unveil(file, "wc") == -1) {
			warn("unveil: %s", file);
			goto error;
		}

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
		goto error;
	}

	if (file == NULL)
		fp = stdout;
	else if ((fp = fopen(file, "wx")) == NULL) {
		warn("fopen: %s", file);
		goto error;
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

	if (ret == -1)
		goto error;

	ret = CMD_OK;
	goto out;

error:
	ret = CMD_ERROR;
	goto out;

usage:
	ret = CMD_USAGE;

out:
	sbk_free_message_list(lst);
	sbk_close(ctx);
	free(signaldir);
	return ret;
}
