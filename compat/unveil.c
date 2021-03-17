/*
 * Written by Tim van der Molen.
 * Public domain.
 */

#include "../config.h"

#ifndef HAVE_UNVEIL

#include "../compat.h"

int
unveil(__unused const char *path, __unused const char *permissions)
{
	return 0;
}

#endif
