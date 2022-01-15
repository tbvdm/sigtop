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

On OpenBSD, sigtop additionally depends on [SQLCipher][6].

Building
--------

sigtop should build on most Unix systems. This section contains generic build
instructions. See the sections below for build instructions for specific
systems.

First install all required packages (see the "Dependencies" section above). For
example, on Debian or Ubuntu, run the following command:

	sudo apt-get install build-essential git libssl-dev pkg-config

After you have installed the required packages, run the following commands:

	git clone https://github.com/tbvdm/sigtop.git
	cd sigtop
	git checkout portable
	make

Building on OpenBSD
-------------------

To build sigtop on OpenBSD, run the following commands:

	doas pkg_add git sqlcipher
	git clone https://github.com/tbvdm/sigtop.git
	cd sigtop
	make

Building on macOS
-----------------

To build sigtop on macOS, first install [Homebrew][7]. Then run the following
command:

	brew install --HEAD tbvdm/tap/sigtop

This will build and install sigtop from [my Homebrew tap][8].

If you prefer to build sigtop manually, run the following commands instead:

	brew install libressl make pkg-config
	git clone https://github.com/tbvdm/sigtop.git
	cd sigtop
	git checkout portable
	PKG_CONFIG_PATH=$(brew --prefix)/opt/libressl/lib/pkgconfig gmake

Building on Windows
-------------------

To build sigtop on Windows, first install [Cygwin][9]. See the [Cygwin User's
Guide][10] if you need help.

You will be able to select additional packages for installation. Ensure the
`gcc-core`, `git`, `libssl-devel`, `make` and `pkg-config` packages are
installed.

After the installation has completed, start the Cygwin terminal. Then run the
following commands to build and install sigtop:

	git clone https://github.com/tbvdm/sigtop.git
	cd sigtop
	git checkout portable
	make install

You can access your Windows drives through the `/cygdrive` directory. For
example:

	sigtop export-messages /cygdrive/c/Users/Alice/Documents/messages.txt

Reporting problems
------------------

Please report bugs and other problems with sigtop. If sigtop shows errors or
warnings unexpectedly, please report them as well. You can [open an issue on
GitHub][11] or send an email. You can find my email address at the top of the
`sigtop.c` file.

[1]: https://www.kariliq.nl/sigbak/
[2]: https://github.com/signalapp/Signal-Desktop
[3]: https://www.kariliq.nl/man/sigtop.1.html
[4]: https://www.libressl.org/
[5]: https://www.openssl.org/
[6]: https://www.zetetic.net/sqlcipher/
[7]: https://brew.sh/
[8]: https://github.com/tbvdm/homebrew-tap
[9]: https://cygwin.com/
[10]: https://cygwin.com/cygwin-ug-net/setup-net.html#internet-setup
[11]: https://github.com/tbvdm/sigtop/issues
