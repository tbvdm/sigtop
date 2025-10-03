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

First install [Go][4] (version 1.21 or later) and a C compiler. On systems
other than OpenBSD, you also need to install `libsecret` and `pkg-config`.

On Ubuntu 24.04 or later, you can run the following command to install the
required packages:

	sudo apt install gcc golang libsecret-1-dev pkg-config

Then, to install sigtop, run:

	go install github.com/tbvdm/sigtop@master

This command installs a `sigtop` binary in `~/go/bin`. You can choose another
installation directory by setting the `GOBIN` environment variable. For
example, to install sigtop in `~/bin`, run:

	GOBIN=~/bin go install github.com/tbvdm/sigtop@master

If you prefer, you can install sigtop without `libsecret` support by specifying
the `no_libsecret` build tag:

	go install -tags no_libsecret github.com/tbvdm/sigtop@master

If you are running Ubuntu 18.04, 20.04 or 22.04, you will have to install a
newer version of Go. See the [Go wiki][5] for more information.

### Build the binary with docker
If you already have docker installed you can simply build the binary using the Dockerfile which exports it in the current directory
```sh
docker build --target export --output . .
```

## Installing on macOS

First install [Homebrew][6]. Then, to install sigtop, run:

	brew install --HEAD tbvdm/tap/sigtop

Later, if you want to update sigtop, run:

	brew upgrade --fetch-HEAD sigtop

## Installing on Windows

There are several ways to get sigtop on Windows.

Note that sigtop is a command-line program; it should be run in a PowerShell or
Command Prompt window.

### Downloading a pre-compiled binary

You can download a [pre-compiled Windows binary][7] from the [latest
release][8].

### Building from source

First install [Go][4]. Next, install the C compiler from [WinLibs][9]: download
[this Zip archive][10] and unzip it to `C:\winlibs`.

Then, to install sigtop, open a PowerShell window and run:

	$env:cc = 'c:\winlibs\mingw64\bin\gcc'
	go install github.com/tbvdm/sigtop@master

This command installs `sigtop.exe` in `C:\Users\<username>\go\bin`. This
directory has been added to your `PATH`, so you can simply type `sigtop` in
PowerShell to run sigtop.

### Cross-compiling in WSL

If you have installed [WSL][11], you may find it simpler to cross-compile. For
example, if you are running Ubuntu (22.04 or later) in WSL:

	sudo apt install golang gcc-mingw-w64-x86-64
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go install github.com/tbvdm/sigtop@master

This command installs `sigtop.exe` in `~/go/bin/windows_amd64`. You can move
the binary to another location if you wish. For example:

	mv ~/go/bin/windows_amd64/sigtop.exe /mnt/c/Users/Alice

## Reporting problems

Please report bugs and other problems with sigtop. You can [open an issue on
GitHub][12] or [send an email][13].

[1]: https://github.com/tbvdm/sigtop
[2]: https://github.com/signalapp/Signal-Desktop
[3]: https://www.kariliq.nl/man/sigtop.1.html
[4]: https://go.dev/
[5]: https://go.dev/wiki/Ubuntu
[6]: https://brew.sh/
[7]: https://github.com/tbvdm/sigtop/releases/latest/download/sigtop.exe
[8]: https://github.com/tbvdm/sigtop/releases/latest
[9]: https://winlibs.com/
[10]: https://github.com/brechtsanders/winlibs_mingw/releases/download/14.2.0posix-18.1.8-12.0.0-ucrt-r1/winlibs-x86_64-posix-seh-gcc-14.2.0-mingw-w64ucrt-12.0.0-r1.zip
[11]: https://learn.microsoft.com/windows/wsl/
[12]: https://github.com/tbvdm/sigtop/issues
[13]: https://www.kariliq.nl/contact.html
