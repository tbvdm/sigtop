PROG=		sigtop
SRCS=		cmd-attachments.c cmd-check.c cmd-messages.c cmd-sqlite.c \
		mime.c sbk.c sigtop.c utf.c

.if !(make(clean) || make(cleandir) || make(obj))
CFLAGS+!=	pkg-config --cflags sqlcipher
LDADD+!=	pkg-config --libs sqlcipher
.endif

.include <bsd.prog.mk>
