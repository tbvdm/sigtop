PROG=		sigtop
SRCS=		cmd-messages.c sbk.c sigtop.c
NOMAN=

.if !(make(clean) || make(cleandir) || make(obj))
CFLAGS+!=	pkg-config --cflags sqlcipher
LDADD+!=	pkg-config --libs sqlcipher
.endif

.include <bsd.prog.mk>
