# About sqlite3.c and sqlite3.h

The `sqlite3.c` and `sqlite3.h` files were generated from the SQLCipher source
using the following procedure.

Clone the SQLCipher repository and check out the `v4.17.0` tag:

	git clone -b v4.17.0 https://github.com/sqlcipher/sqlcipher.git
	cd sqlcipher

Copy (and rename) `_crypto_go.c`:

	cp /path/to/sigtop/sqlcipher/_crypto_go.c crypto_go.c

Generate `sqlite3.c` and `sqlite3.h`:

	./configure
	make EXTRA_SRC=crypto_go.c sqlite3.c

Move `sqlite3.c` and `sqlite3.h` into place:

	mv sqlite3.[ch] /path/to/sigtop/sqlcipher
