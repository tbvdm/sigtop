PROG=		sigtop
SRCS=		cmd-check-database.c cmd-export-attachments.c \
		cmd-export-database.c cmd-export-messages.c mime.c \
		sbk-attachment.c sbk-conversation.c sbk-jsmn.c sbk-mention.c \
		sbk-message.c sbk-open.c sbk-quote.c sbk-reaction.c \
		sbk-recipient.c sbk-sqlite.c sigtop.c utf.c

.if !(make(clean) || make(cleandir) || make(obj))
CFLAGS+!=	pkg-config --cflags sqlcipher
LDADD+!=	pkg-config --libs sqlcipher
.endif

.include <bsd.prog.mk>
