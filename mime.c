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

#include <string.h>

#include "sigtop.h"

static struct {
	const char *type;
	const char *extension;
} mime_extensions[] = {
	{ "application/gzip",					"gz" },
	{ "application/msword",					"doc" },
	{ "application/pdf",					"pdf" },
	{ "application/rtf",					"rtf" },
	{ "application/vnd.oasis.opendocument.presentation",	"odp" },
	{ "application/vnd.oasis.opendocument.spreadsheet",	"ods" },
	{ "application/vnd.oasis.opendocument.text",		"odt" },
	{ "application/vnd.openxmlformats-officedocument.presentationml.presentation", "pptx" },
	{ "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", "xlsx" },
	{ "application/vnd.openxmlformats-officedocument.wordprocessingml.document", "docx" },
	{ "application/vnd.rar",				"rar" },
	{ "application/x-7z-compressed",			"7z" },
	{ "application/x-bzip2",				"bz2" },
	{ "application/x-tar",					"tar" },
	{ "application/zip",					"zip" },

	{ "audio/aac",						"aac" },
	{ "audio/flac",						"flac" },
	{ "audio/ogg",						"ogg" },
	{ "audio/mp4",						"mp4" },
	{ "audio/mpeg",						"mp3" },

	{ "image/gif",						"gif" },
	{ "image/jpeg",						"jpg" },
	{ "image/png",						"png" },
	{ "image/svg+xml",					"svg" },
	{ "image/tiff",						"tiff" },
	{ "image/webp",						"webp" },

	{ "text/html",						"html" },
	{ "text/plain",						"txt" },
	{ "text/x-signal-plain",				"txt" },

	{ "video/mp4",						"mp4" },
	{ "video/mpeg",						"mpg" },
};

const char *
mime_get_extension(const char *type)
{
	size_t i;

	for (i = 0; i < nitems(mime_extensions); i++)
		if (strcmp(mime_extensions[i].type, type) == 0)
			return mime_extensions[i].extension;

	return NULL;
}
