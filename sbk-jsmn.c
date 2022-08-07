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

#define JSMN_STRICT

#include "config.h"

#include <errno.h>
#include <stdint.h>
#include <stdlib.h>
#include <string.h>

#include "jsmn.h"
#include "sbk-internal.h"
#include "utf.h"

int
sbk_jsmn_parse(const char *json, size_t jsonlen, jsmntok_t *tokens,
    size_t ntokens)
{
	jsmn_parser	parser;
	int		ret;

	jsmn_init(&parser);
	ret = jsmn_parse(&parser, json, jsonlen, tokens, ntokens);
	if (ret < 0) {
		switch (ret) {
		case JSMN_ERROR_NOMEM:
			warnx("Not enough tokens to parse JSON");
			break;
		case JSMN_ERROR_INVAL:
		case JSMN_ERROR_PART:
			warnx("Invalid JSON data");
			break;
		default:
			warnx("Unexpected JSON parse error");
			break;
		}
		return -1;
	}

	return 0;
}

static int
sbk_jsmn_is_valid_key(const jsmntok_t *token)
{
	return token->type == JSMN_STRING && token->size == 1;
}

static int
sbk_jsmn_token_equals(const char *json, const jsmntok_t *token,
    const char *str)
{
	size_t len;

	len = strlen(str);
	if (len != (unsigned int)(token->end - token->start))
		return 0;
	else
		return memcmp(json + token->start, str, len) == 0;
}

int
sbk_jsmn_get_total_token_size(const jsmntok_t *tokens)
{
	int i, idx, size;

	idx = 1;
	switch (tokens[0].type) {
	case JSMN_OBJECT:
		for (i = 0; i < tokens[0].size; i++) {
			if (!sbk_jsmn_is_valid_key(&tokens[idx])) {
				warnx("Invalid JSON key");
				return -1;
			}
			size = sbk_jsmn_get_total_token_size(&tokens[++idx]);
			if (size == -1)
				return -1;
			idx += size;
		}
		break;
	case JSMN_ARRAY:
		for (i = 0; i < tokens[0].size; i++) {
			size = sbk_jsmn_get_total_token_size(&tokens[idx]);
			if (size == -1)
				return -1;
			idx += size;
		}
		break;
	case JSMN_STRING:
	case JSMN_PRIMITIVE:
		if (tokens[0].size != 0) {
			warnx("Invalid JSON data");
			return -1;
		}
		break;
	case JSMN_UNDEFINED:
		warnx("Invalid JSON data");
		return -1;
	}

	return idx;
}

static int
sbk_jsmn_find_key(const char *json, const jsmntok_t *tokens, const char *key)
{
	int i, idx, size;

	if (tokens[0].type != JSMN_OBJECT)
		return -1;

	idx = 1;
	for (i = 0; i < tokens[0].size; i++) {
		if (!sbk_jsmn_is_valid_key(&tokens[idx]))
			return -1;
		if (sbk_jsmn_token_equals(json, &tokens[idx], key))
			return idx;
		/* Skip value */
		size = sbk_jsmn_get_total_token_size(&tokens[++idx]);
		if (size == -1)
			return -1;
		idx += size;
	}

	/* Not found */
	return -1;
}

/* Check that a JSMN_PRIMITIVE token is a number (and not a boolean or null) */
static int
sbk_jsmn_primitive_is_number(const char *json, const jsmntok_t *token)
{
	char c;

	c = json[token->start];
	return c == '-' || (c >= '0' && c <= '9');
}

static int
sbk_jsmn_get_value(const char *json, const jsmntok_t *tokens, const char *key,
    jsmntype_t type)
{
	int idx;

	idx = sbk_jsmn_find_key(json, tokens, key);
	if (idx == -1)
		return -1;
	if (tokens[++idx].type != type)
		return -1;
	return idx;
}

int
sbk_jsmn_get_array(const char *json, const jsmntok_t *tokens, const char *key)
{
	return sbk_jsmn_get_value(json, tokens, key, JSMN_ARRAY);
}

int
sbk_jsmn_get_object(const char *json, const jsmntok_t *tokens, const char *key)
{
	return sbk_jsmn_get_value(json, tokens, key, JSMN_OBJECT);
}

int
sbk_jsmn_get_string(const char *json, const jsmntok_t *tokens, const char *key)
{
	return sbk_jsmn_get_value(json, tokens, key, JSMN_STRING);
}

int
sbk_jsmn_get_number(const char *json, const jsmntok_t *tokens, const char *key)
{
	int idx;

	idx = sbk_jsmn_get_value(json, tokens, key, JSMN_PRIMITIVE);
	if (idx == -1)
		return -1;

	if (!sbk_jsmn_primitive_is_number(json, &tokens[idx]))
		return -1;

	return idx;
}

int
sbk_jsmn_get_number_or_string(const char *json, const jsmntok_t *tokens,
    const char *key)
{
	int idx;

	idx = sbk_jsmn_find_key(json, tokens, key);
	if (idx == -1)
		return -1;

	idx++;

	if (tokens[idx].type == JSMN_STRING)
		return idx;

	if (tokens[idx].type == JSMN_PRIMITIVE &&
	    sbk_jsmn_primitive_is_number(json, &tokens[idx]))
		return idx;

	return -1;
}

/* Auxiliary function for sbk_jsmn_parse_unicode_escape() */
static int
sbk_jsmn_parse_hex(uint16_t *u, const char *s)
{
	int		i;
	uint16_t	v;
	char		c;

	*u = 0;
	for (i = 0; i < 4; i++) {
		c = s[i];
		if (c >= '0' && c <= '9')
			v = c - '0';
		else if (c >= 'a' && c <= 'f')
			v = c - 'a' + 10;
		else if (c >= 'A' && c <= 'F')
			v = c - 'A' + 10;
		else
			return -1;
		*u = *u * 16 + v;
	}
	return 0;
}

static int
sbk_jsmn_parse_unicode_escape(char **r, char **w)
{
	size_t		len;
	uint32_t	cp;		/* Unicode code point */
	uint16_t	utf16[2];

	/* Skip the leading "\u". */
	*r += 2;

	/* Parse the four hexadecimal digits that should follow. */
	if (sbk_jsmn_parse_hex(&utf16[0], *r) == -1)
		goto error;
	*r += 4;

	if (!utf16_is_high_surrogate(utf16[0])) {
		/*
		 * The \u escape does not contain a high surrogate, so either
		 * it represents a character or it contains an unpaired low
		 * surrogate, which we'll also allow.
		 */
		cp = utf16[0];
		goto finish;
	}

	/*
	 * The \u escape contains a high surrogate, so it should be followed
	 * by a second \u escape containing the low surrogate.
	 */
	if ((*r)[0] != '\\' || (*r)[1] != 'u') {
		/*
		 * There's no \u escape following, so we end up with an
		 * unpaired high surrogate. Allow it.
		 */
		cp = utf16[0];
		goto finish;
	}

	/* Parse the four hexadecimal digits of the second \u escape. */
	if (sbk_jsmn_parse_hex(&utf16[1], *r + 2) == -1)
		goto error;

	if (!utf16_is_low_surrogate(utf16[1])) {
		/*
		 * The second \u escape does not contain a low surrogate, so we
		 * end up with an unpaired high surrogate. Allow it. (We will
		 * not parse the second \u escape further; it will be revisited
		 * in the next call.)
		 */
		cp = utf16[0];
		goto finish;
	}

	/*
	 * The second \u escape contains a low surrogate, so we now have a
	 * complete surrogate pair. First decode the code point in the
	 * surrogate pair. Then update the read pointer to point after the
	 * second \u escape.
	 */
	cp = utf16_decode_surrogate_pair(utf16[0], utf16[1]);
	*r += 6;

finish:
	/* Write the UTF-8 encoding of the code point. */
	if ((len = utf8_encode((uint8_t *)*w, cp)) == 0)
		goto error;
	*w += len;

	return 0;

error:
	warnx("Invalid \\u escape in JSON string");
	return -1;
}

static int
sbk_jsmn_parse_escape(char **r, char **w)
{
	switch ((*r)[1]) {
	case '"':
	case '\\':
	case '/':
		**w = (*r)[1];
		break;
	case 'b':
		**w = '\b';
		break;
	case 'f':
		**w = '\f';
		break;
	case 'n':
		**w = '\n';
		break;
	case 'r':
		**w = '\r';
		break;
	case 't':
		**w = '\t';
		break;
	case 'u':
		/* Handle \u escapes separately */
		return sbk_jsmn_parse_unicode_escape(r, w);
	default:
		return -1;
	}

	*r += 2;	/* We read a 2-char escape sequence... */
	*w += 1;	/* ... and wrote one char */
	return 0;
}

/*
 * Perform in-place substitution of escape sequences in a JSON string. In-place
 * substitution is possible because each escape sequence is longer than its
 * substitute.
 */
static char *
sbk_jsmn_unescape(char *s)
{
	char	*r, *w;
	size_t	 len;

	r = w = s + strcspn(s, "\\");
	while (*r == '\\') {
		if (sbk_jsmn_parse_escape(&r, &w) == -1) {
			*s = '\0';
			return NULL;
		}
		len = strcspn(r, "\\");
		memmove(w, r, len);
		r += len;
		w += len;
	}
	*w = '\0';
	return s;
}

char *
sbk_jsmn_parse_string(const char *json, const jsmntok_t *token)
{
	char *s;

	s = strndup(json + token->start, token->end - token->start);
	if (s == NULL) {
		warn(NULL);
		return NULL;
	}

	if (sbk_jsmn_unescape(s) == NULL) {
		free(s);
		return NULL;
	}

	return s;
}

int
sbk_jsmn_parse_uint64(uint64_t *val, const char *json, const jsmntok_t *token)
{
	char			*end;
	unsigned long long	 num;

	errno = 0;
	num = strtoull(json + token->start, &end, 10);

	if (errno != 0 || end != json + token->end)
		goto invalid;

#if ULLONG_MAX > UINT64_MAX
	if (num > UINT64_MAX)
		goto invalid;
#endif

	*val = num;
	return 0;

invalid:
	warnx("Invalid JSON number");
	return -1;
}
