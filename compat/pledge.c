/*
 * Written by Tim van der Molen.
 * Public domain.
 */

#include "../config.h"

#ifndef HAVE_PLEDGE

#include "../compat.h"

int
pledge(__unused const char *promises, __unused const char *execpromises)
{
	return 0;
}

#endif
