/*
 * Written by Tim van der Molen.
 * Public domain.
 */

#include "../config.h"

#if !defined(HAVE_EXPLICIT_BZERO) && !defined(LIBRESSL_VERSION_NUMBER)

#include <openssl/crypto.h>

void
explicit_bzero(void *buf, size_t len)
{
	OPENSSL_cleanse(buf, len);
}

#endif
