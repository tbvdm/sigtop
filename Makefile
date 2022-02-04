PROG=		sigtop
SRCS=		cmd-check-database.c cmd-export-attachments.c \
		cmd-export-database.c cmd-export-messages.c mime.c sbk.c \
		sigtop.c utf.c

.if !(make(clean) || make(cleandir) || make(obj))
CFLAGS+!=	pkg-config --cflags sqlcipher
LDADD+!=	pkg-config --libs sqlcipher
.endif

.include <bsd.prog.mk>
