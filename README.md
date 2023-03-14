# sigtop

[sigtop][1] is a utility to export messages, attachments and other data from
[Signal Desktop][2].

For example, the following two commands export all messages to the `messages`
directory and all attachments to the `attachments` directory:

	sigtop export-messages messages
	sigtop export-attachments attachments

Documentation is available in the `sigtop.1` manual page. You can also [read it
online][3].

## Installing on macOS

First install [Homebrew][4]. Then, to install sigtop, run:

	brew install --HEAD tbvdm/tap/sigtop

Later, if you want to update sigtop, run:

	brew upgrade --fetch-HEAD sigtop

## Installing on other Unix-like systems

First install [Go][5] (version 1.18 or later) and a C compiler. On Ubuntu 22.04
or later, you can run:

	sudo apt-get install golang gcc

Then, to install sigtop, run:

	go install github.com/tbvdm/sigtop@master

This command installs a `sigtop` binary in `~/go/bin`. You can choose another
installation directory by setting the `GOBIN` environment variable. For
example, to install sigtop in `~/bin`, run:

	GOBIN=~/bin go install github.com/tbvdm/sigtop@master

## Installing on Windows

First install [Go][5]. Next, install the C compiler from [WinLibs][6]: download
[this Zip archive][7] and extract it to `C:\winlibs`.

Then, to install sigtop, open a PowerShell window and run:

	$env:cc = 'c:\winlibs\mingw64\bin\gcc'
	go install github.com/tbvdm/sigtop@master

This command installs `sigtop.exe` in `C:\Users\<username>\go\bin`. This
directory has been added to your `PATH`, so you can simply type `sigtop` in
PowerShell to run sigtop.

## Downloading pre-compiled binaries

You can also download a pre-compiled binary from the [latest release][8]:

- [macOS (Intel)][9]
- [Linux (x86-64)][10]
- [Windows (x86-64)][11]

## Reporting problems

Please report bugs and other problems with sigtop. You can [open an issue on
GitHub][12] or [send an email][13].

[1]: https://github.com/tbvdm/sigtop
[2]: https://github.com/signalapp/Signal-Desktop
[3]: https://www.kariliq.nl/man/sigtop.1.html
[4]: https://brew.sh/
[5]: https://go.dev/
[6]: https://winlibs.com/
[7]: https://github.com/brechtsanders/winlibs_mingw/releases/download/12.2.0-15.0.7-10.0.0-ucrt-r4/winlibs-x86_64-posix-seh-gcc-12.2.0-mingw-w64ucrt-10.0.0-r4.zip
[8]: https://github.com/tbvdm/sigtop/releases/latest
[9]: https://github.com/tbvdm/sigtop/releases/latest/download/sigtop-darwin-amd64
[10]: https://github.com/tbvdm/sigtop/releases/latest/download/sigtop-linux-amd64
[11]: https://github.com/tbvdm/sigtop/releases/latest/download/sigtop-windows-amd64.exe
[12]: https://github.com/tbvdm/sigtop/issues
[13]: https://www.kariliq.nl/contact.html
