PREFIX?=	/usr/local
BINDIR?=	${PREFIX}/bin
MANDIR?=	${PREFIX}/man

CC?=		cc
INSTALL?=	install
PKG_CONFIG?=	pkg-config

PKGS?=		libcrypto sqlcipher

PKGS_CFLAGS!=	${PKG_CONFIG} --cflags ${PKGS}
PKGS_LDFLAGS!=	${PKG_CONFIG} --libs ${PKGS}

CFLAGS+=	${PKGS_CFLAGS}
LDFLAGS+=	${PKGS_LDFLAGS}

COMPAT_OBJS=	compat/asprintf.o compat/err.o compat/explicit_bzero.o \
		compat/fopen.o compat/getprogname.o compat/unveil.o

OBJS=		cmd-messages.o cmd-sqlite.o sbk.o sigtop.o ${COMPAT_OBJS}

.PHONY: all clean install

.SUFFIXES: .c .o

.c.o:
	${CC} ${CFLAGS} ${CPPFLAGS} -c -o $@ $<

all: sigtop

sigtop: ${OBJS}
	${CC} -o $@ ${OBJS} ${LDFLAGS}

${OBJS}: config.h

clean:
	rm -f sigtop sigtop.core core ${OBJS}

install: all
	${INSTALL} -dm 755 ${DESTDIR}${BINDIR}
	${INSTALL} -dm 755 ${DESTDIR}${MANDIR}/man1
	${INSTALL} -m 555 sigtop ${DESTDIR}${BINDIR}
	${INSTALL} -m 444 sigtop.1 ${DESTDIR}${MANDIR}/man1
