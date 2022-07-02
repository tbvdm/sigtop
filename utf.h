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

#ifndef UTF_H
#define UTF_H

#include <stddef.h>
#include <stdint.h>

#define utf8_is_single(b)	(((uint8_t)(b) & 0x80) == 0)
#define utf8_is_start2(b)	(((uint8_t)(b) & 0xe0) == 0xc0)
#define utf8_is_start3(b)	(((uint8_t)(b) & 0xf0) == 0xe0)
#define utf8_is_start4(b)	(((uint8_t)(b) & 0xf8) == 0xf0)
#define utf8_is_cont(b)		(((uint8_t)(b) & 0xc0) == 0x80)

size_t		utf8_encode(uint8_t [4], uint32_t);
size_t		utf8_get_sequence_length(const uint8_t *);
size_t		utf8_get_substring_length(const uint8_t *, size_t);
int		utf16_is_high_surrogate(uint16_t);
int		utf16_is_low_surrogate(uint16_t);
uint32_t	utf16_decode_surrogate_pair(uint16_t, uint16_t);

#endif
