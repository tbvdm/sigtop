# sigtop

[sigtop][1] is a utility to export messages, attachments and other data from
[Signal Desktop][2].

For example, the following two commands export all messages to the `messages`
directory and all attachments to the `attachments` directory:

	sigtop export-messages messages
	sigtop export-attachments attachments

Documentation is available in the `sigtop.1` manual page. You can also [read it
online][3].

## Installing on Unix

First install [Go][4] (version 1.18 or later) and a C compiler. On Ubuntu 22.04
or later, you can run:

	sudo apt install golang gcc

Then, to install sigtop, run:

	go install github.com/tbvdm/sigtop@master

This command installs a `sigtop` binary in `~/go/bin`. You can choose another
installation directory by setting the `GOBIN` environment variable. For
example, to install sigtop in `~/bin`, run:

	GOBIN=~/bin go install github.com/tbvdm/sigtop@master

## Installing on macOS

First install [Homebrew][5]. Then, to install sigtop, run:

	brew install --HEAD tbvdm/tap/sigtop

Later, if you want to update sigtop, run:

	brew upgrade --fetch-HEAD sigtop

## Installing on Windows

There are several ways to get sigtop on Windows.

Note that sigtop is a command-line program; it should be run in a PowerShell or
Command Prompt window.

### Downloading a pre-compiled binary

You can download a [pre-compiled Windows binary][6] from the [latest
release][7].

### Building from source

First install [Go][4]. Next, install the C compiler from [WinLibs][8]: download
[this Zip archive][9] and unzip it to `C:\winlibs`.

Then, to install sigtop, open a PowerShell window and run:

	$env:cc = 'c:\winlibs\mingw64\bin\gcc'
	go install github.com/tbvdm/sigtop@master

This command installs `sigtop.exe` in `C:\Users\<username>\go\bin`. This
directory has been added to your `PATH`, so you can simply type `sigtop` in
PowerShell to run sigtop.

## Reporting problems

Please report bugs and other problems with sigtop. You can [open an issue on
GitHub][10] or [send an email][11].

[1]: https://github.com/tbvdm/sigtop
[2]: https://github.com/signalapp/Signal-Desktop
[3]: https://www.kariliq.nl/man/sigtop.1.html
[4]: https://go.dev/
[5]: https://brew.sh/
[6]: https://github.com/tbvdm/sigtop/releases/latest/download/sigtop.exe
[7]: https://github.com/tbvdm/sigtop/releases/latest
[8]: https://winlibs.com/
[9]: https://github.com/brechtsanders/winlibs_mingw/releases/download/13.1.0-16.0.5-11.0.0-ucrt-r5/winlibs-x86_64-posix-seh-gcc-13.1.0-mingw-w64ucrt-11.0.0-r5.zip
[10]: https://github.com/tbvdm/sigtop/issues
[11]: https://www.kariliq.nl/contact.html
