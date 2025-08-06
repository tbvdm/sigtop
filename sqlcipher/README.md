# About sqlite3.c and sqlite3.h

The `sqlite3.c` and `sqlite3.h` files were generated from the SQLCipher source
using the following procedure.

Clone the SQLCipher repository and check out the `v4.10.0` tag:

	git clone -b v4.10.0 https://github.com/sqlcipher/sqlcipher.git
	cd sqlcipher

Apply `sqlcipher.diff`:

	patch < /path/to/sigtop/sqlcipher/sqlcipher.diff

Generate `sqlite3.c` and `sqlite3.h`:

	./configure --with-tempstore=yes CFLAGS="-DSQLITE_HAS_CODEC -DSQLITE_EXTRA_INIT=sqlcipher_extra_init -DSQLITE_EXTRA_SHUTDOWN=sqlcipher_extra_shutdown"
	make sqlite3.c

Move `sqlite3.c` and `sqlite3.h` into place:

	mv sqlite3.[ch] /path/to/sigtop/sqlcipher
