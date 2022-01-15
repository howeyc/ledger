# Installation

There are multiple ways to install the Ledger CLI tool.
Choose any one of the methods below that best suit your needs.

## Pre-compiled binaries

Executable binaries are available for download on the
[GitHub Releases page][releases].
Download the binary for your platform (Windows, macOS, or Linux) and extract
the archive.
The archive contains the `ledger` executable.

To make it easier to run, put the path to the binary into your `PATH`.

[releases]: https://github.com/howeyc/ledger/releases

## Build from source using Go

To build the `ledger` executable from source, you will first need to install Go
Follow the instructions on the [Go installation page].
ledger currently requires at least Go version 1.17.

Once you have installed Go, the following command can be used to build and
install ledger:

```sh
go install github.com/howeyc/ledger/ledger@latest
```

[Go installation page]: https://go.dev/doc/install

This will automatically download ledger, build it, and install it in Go's global
binary directory (`~/go/bin/` by default).

## Archlinux AUR

If you happen to be using Archlinux, you can use the port in AUR.

```sh
yay -S ledger-go
```

