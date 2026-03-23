/*
 * Copyright (c) 2023 Tim van der Molen <tim@kariliq.nl>
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

#if defined(SQLITE_HAS_CODEC) && defined(SQLCIPHER_CRYPTO_GO)

#include "_cgo_export.h"

static int
sqlcipher_go_ctx_init(void **ctx)
{
	return sqlcipherGoInit();
}

static int
sqlcipher_go_ctx_free(void **ctx)
{
	return sqlcipherGoFree();
}

static const char *
sqlcipher_go_get_provider_name(void *ctx)
{
	return sqlcipherGoGetProviderName();
}

static const char *
sqlcipher_go_get_provider_version(void *ctx)
{
	return sqlcipherGoGetProviderVersion();
}

static int
sqlcipher_go_add_random(void *ctx, const void *const_buf, int buf_sz)
{
	void *buf;

	/* Unconstify for Go */
	buf = (void *)const_buf;

	return sqlcipherGoAddRandom(buf, buf_sz);
}

static int
sqlcipher_go_random(void *ctx, void *buf, int buf_sz)
{
	return sqlcipherGoRandom(buf, buf_sz);
}

static int
sqlcipher_go_hmac(void *ctx, int alg, const unsigned char *const_key,
    int key_sz, const unsigned char *const_in, int in_sz,
    const unsigned char *const_in2, int in2_sz, unsigned char *out)
{
	unsigned char *key, *in, *in2;

	/* Unconstify for Go */
	key = (unsigned char *)const_key;
	in = (unsigned char *)const_in;
	in2 = (unsigned char *)const_in2;

	switch (alg) {
	case SQLCIPHER_HMAC_SHA1:
		return sqlcipherGoHMACSHA1(key, key_sz, in, in_sz, in2,
		    in2_sz, out);
	case SQLCIPHER_HMAC_SHA256:
		return sqlcipherGoHMACSHA256(key, key_sz, in, in_sz, in2,
		    in2_sz, out);
	case SQLCIPHER_HMAC_SHA512:
		return sqlcipherGoHMACSHA512(key, key_sz, in, in_sz, in2,
		    in2_sz, out);
	default:
		return SQLITE_ERROR;
	}
}

static int
sqlcipher_go_kdf(void *ctx, int alg, const unsigned char *const_pass,
    int pass_sz, const unsigned char *const_salt, int salt_sz, int workfactor,
    int key_sz, unsigned char *key)
{
	unsigned char *pass, *salt;

	/* Unconstify for Go */
	pass = (unsigned char *)const_pass;
	salt = (unsigned char *)const_salt;

	switch (alg) {
	case SQLCIPHER_HMAC_SHA1:
		return sqlcipherGoKDFSHA1(pass, pass_sz, salt, salt_sz,
		    workfactor, key, key_sz);
	case SQLCIPHER_HMAC_SHA256:
		return sqlcipherGoKDFSHA256(pass, pass_sz, salt, salt_sz,
		    workfactor, key, key_sz);
	case SQLCIPHER_HMAC_SHA512:
		return sqlcipherGoKDFSHA512(pass, pass_sz, salt, salt_sz,
		    workfactor, key, key_sz);
	default:
		return SQLITE_ERROR;
	}
}

static int
sqlcipher_go_cipher(void *ctx, int mode, const unsigned char *const_key,
    int key_sz, const unsigned char *const_iv, const unsigned char *const_in,
    int in_sz, unsigned char *out)
{
	unsigned char *key, *iv, *in;

	/* Unconstify for Go */
	key = (unsigned char *)const_key;
	iv = (unsigned char *)const_iv;
	in = (unsigned char *)const_in;

	return sqlcipherGoCipher(key, key_sz, iv, in, in_sz, out,
	    mode == SQLCIPHER_ENCRYPT);
}

static const char *
sqlcipher_go_get_cipher(void *ctx)
{
	return sqlcipherGoGetCipher();
}

static int
sqlcipher_go_get_key_sz(void *ctx)
{
	return sqlcipherGoGetKeySize();
}

static int
sqlcipher_go_get_iv_sz(void *ctx)
{
	return sqlcipherGoGetIVSize();
}

static int
sqlcipher_go_get_block_sz(void *ctx)
{
	return sqlcipherGoGetBlockSize();
}

static int
sqlcipher_go_get_hmac_sz(void *ctx, int alg)
{
	switch (alg) {
	case SQLCIPHER_HMAC_SHA1:
		return sqlcipherGoGetHMACSizeSHA1();
	case SQLCIPHER_HMAC_SHA256:
		return sqlcipherGoGetHMACSizeSHA256();
	case SQLCIPHER_HMAC_SHA512:
		return sqlcipherGoGetHMACSizeSHA512();
	default:
		return 0;
	}
}

static int
sqlcipher_go_fips_status(void *ctx)
{
	return sqlcipherGoFIPSStatus();
}

int
sqlcipher_go_setup(sqlcipher_provider *p)
{
	p->init = NULL;
	p->shutdown = NULL;
	p->get_provider_name = sqlcipher_go_get_provider_name;
	p->add_random = sqlcipher_go_add_random;
	p->random = sqlcipher_go_random;
	p->hmac = sqlcipher_go_hmac;
	p->kdf = sqlcipher_go_kdf;
	p->cipher = sqlcipher_go_cipher;
	p->get_cipher = sqlcipher_go_get_cipher;
	p->get_key_sz = sqlcipher_go_get_key_sz;
	p->get_iv_sz = sqlcipher_go_get_iv_sz;
	p->get_block_sz = sqlcipher_go_get_block_sz;
	p->get_hmac_sz = sqlcipher_go_get_hmac_sz;
	p->ctx_init = sqlcipher_go_ctx_init;
	p->ctx_free = sqlcipher_go_ctx_free;
	p->fips_status = sqlcipher_go_fips_status;
	p->get_provider_version = sqlcipher_go_get_provider_version;
	return SQLITE_OK;
}

#endif
