sigtop
======

[sigtop][1] is a utility to export messages and other data from [Signal
Desktop][2].

Documentation is available in the `sigtop.1` manual page. It can also be [read
online][3].

Dependencies
------------

sigtop depends on libcrypto (from either [LibreSSL][4] or [OpenSSL][5]). You
will also need a C compiler, `make` and `pkg-config`.

Building on OpenBSD
-------------------

On OpenBSD, sigtop also depends on [SQLCipher][6]. So install the `sqlcipher`
package first. Then run `make` and optionally `make install`.

Building on other systems
-------------------------

To build sigtop on other systems, first check out the `portable` branch:

	$ git checkout portable

Then check if `config.h` is suited to your system. Edit it if necessary.
`config.h` already has support for several systems. On those systems, no
editing should be necessary.

Finally, run `make` and optionally `make install`.

If you are unsure what to do with `config.h`, then leave it as is and just run
`make`. It is likely to work fine.

[1]: https://github.com/tbvdm/sigtop
[2]: https://github.com/signalapp/Signal-Desktop
[3]: https://www.kariliq.nl/sigtop/manual.html
[4]: https://www.libressl.org/
[5]: https://www.openssl.org/
[6]: https://www.zetetic.net/sqlcipher/
