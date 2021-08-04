sigtop
======

[sigtop][1] is a utility to export messages, attachments and other data from
[Signal Desktop][2].

Documentation is available in the `sigtop.1` manual page. You can also [read it
online][3].

Dependencies
------------

sigtop depends on libcrypto (from either [LibreSSL][4] or [OpenSSL][5]). You
will also need a C compiler, `make` and `pkg-config`.

On OpenBSD, sigtop additionally depends on the `sqlcipher` package.

Building on OpenBSD
-------------------

To build sigtop on OpenBSD, clone the repository and run `make`.

Building on other systems
-------------------------

To build sigtop on other systems, clone the repository, switch to the
`portable` branch and run `make`:

	git clone https://github.com/tbvdm/sigtop.git
	cd sigtop
	git checkout portable
	make

sigtop should build without problems on Linux and the BSDs.

If the build does fail, check if `config.h` is suited to your system. You may
have to edit it. After editing `config.h`, run `make` to retry the build.

Building on macOS
-----------------

To build sigtop on macOS, first install [Homebrew][6].

Then install the required packages, clone the repository, switch to the
`portable` branch and run `gmake`:

	brew install libressl make pkg-config
	git clone https://github.com/tbvdm/sigtop.git
	cd sigtop
	git checkout portable
	PKG_CONFIG_PATH=$(brew --prefix)/opt/libressl/lib/pkgconfig gmake

Reporting problems
------------------

Please report bugs and other problems with sigtop. If sigtop shows errors or
warnings unexpectedly, please report them as well. You can [open an issue on
GitHub][7] or send an email. You can find my email address at the top of the
`sigtop.c` file.

[1]: https://www.kariliq.nl/sigbak/
[2]: https://github.com/signalapp/Signal-Desktop
[3]: https://www.kariliq.nl/man/sigtop.1.html
[4]: https://www.libressl.org/
[5]: https://www.openssl.org/
[6]: https://brew.sh/
[7]: https://github.com/tbvdm/sigtop/issues
