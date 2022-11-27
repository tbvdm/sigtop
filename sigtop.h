/*
 * Copyright (c) 2018 Tim van der Molen <tim@kariliq.nl>
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

#ifndef SIGTOP_H
#define SIGTOP_H

#include <sys/types.h>

#include "sbk.h"

enum cmd_status {
	CMD_OK,
	CMD_ERROR,
	CMD_USAGE
};

struct cmd_entry {
	const char	*name;
	const char	*alias;
	const char	*usage;
	enum cmd_status	 (*exec)(int, char **);
};

const char	*mime_get_extension(const char *);

char		*get_signal_dir(void);
int		 unveil_dirname(const char *, const char *);
int		 unveil_signal_dir(const char *);
int		 parse_time_interval(char *, time_t *, time_t *);
void		 sanitise_filename(char *);
char		*get_recipient_filename(struct sbk_recipient *, const char *);

#endif
