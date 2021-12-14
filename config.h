/*
 * This file must be suited to your system. Edit it if necessary.
 *
 * Several systems are already supported; see further below. On those systems,
 * no editing should be necessary.
 */

/* Define if you have asprintf() and vasprintf(). */
/* #define HAVE_ASPRINTF */

/* Define if you have the err() family of functions. */
/* #define HAVE_ERR */

/* Define if you have explicit_bzero(). */
/* #define HAVE_EXPLICIT_BZERO */

/* Define if your fopen() supports the "x" mode extension. */
/* #define HAVE_FOPEN_X_MODE */

/* Define if you have getprogname() and setprogname(). */
/* #define HAVE_GETPROGNAME */

/* Define if you have pledge(). */
/* #define HAVE_PLEDGE */

/* Define if you have reallocarray(). */
/* #define HAVE_REALLOCARRAY */

/* Define if your struct tm has a tm_gmtoff member. */
/* #define HAVE_TM_GMTOFF */

/* Define if you have unveil(). */
/* #define HAVE_UNVEIL */

#ifdef __APPLE__

#define HAVE_ASPRINTF
#define HAVE_ERR
#define HAVE_FOPEN_X_MODE
#define HAVE_GETPROGNAME
#define HAVE_TM_GMTOFF

#elif defined(__DragonFly__)

#define HAVE_ASPRINTF
#define HAVE_ERR
#define HAVE_EXPLICIT_BZERO
#define HAVE_FOPEN_X_MODE
#define HAVE_GETPROGNAME
#define HAVE_REALLOCARRAY
#define HAVE_TM_GMTOFF

#elif defined(__FreeBSD__)

#define HAVE_ASPRINTF
#define HAVE_ERR
#define HAVE_EXPLICIT_BZERO
#define HAVE_FOPEN_X_MODE
#define HAVE_GETPROGNAME
#define HAVE_REALLOCARRAY
#define HAVE_TM_GMTOFF

#elif defined(__NetBSD__)

#define _OPENBSD_SOURCE

#define HAVE_ASPRINTF
#define HAVE_ERR
#define HAVE_FOPEN_X_MODE
#define HAVE_GETPROGNAME
#define HAVE_REALLOCARRAY
#define HAVE_TM_GMTOFF

#elif defined(__OpenBSD__)

#define HAVE_ASPRINTF
#define HAVE_ERR
#define HAVE_EXPLICIT_BZERO
#define HAVE_FOPEN_X_MODE
#define HAVE_GETPROGNAME
#define HAVE_PLEDGE
#define HAVE_REALLOCARRAY
#define HAVE_TM_GMTOFF
#define HAVE_UNVEIL

#elif defined(__linux__)

#define _GNU_SOURCE

/* All modern versions of glibc, musl and bionic have these. */
#define HAVE_ASPRINTF
#define HAVE_ERR
#define HAVE_FOPEN_X_MODE
#define HAVE_TM_GMTOFF

#include <features.h>

/* glibc */

#ifdef __GLIBC_PREREQ
#  if __GLIBC_PREREQ(2, 25)
#    define HAVE_EXPLICIT_BZERO
#  endif
#  if __GLIBC_PREREQ(2, 26)
#    define HAVE_REALLOCARRAY
#  endif
#endif

/* bionic */

#ifdef __ANDROID_API__
#  if __ANDROID_API__ >= 21
#    define HAVE_GETPROGNAME
#  endif
#  if __ANDROID_API__ >= 29
#    define HAVE_REALLOCARRAY
#  endif
#endif

/* musl */

/* Define if you have musl >= 1.1.20. */
/* #define HAVE_EXPLICIT_BZERO */

/* Define if you have musl >= 1.2.2. */
/* #define HAVE_REALLOCARRAY */

#elif defined(__sun)

#define HAVE_ASPRINTF
#define HAVE_ERR
#define HAVE_FOPEN_X_MODE
#define HAVE_GETPROGNAME

#endif
