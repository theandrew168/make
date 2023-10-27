# make

Subset of [POSIX make](https://pubs.opengroup.org/onlinepubs/9699919799/utilities/make.html) implemented in [Go](https://golang.org/dl/).

## Setup

This project depends on the [Go programming language](https://golang.org/dl/).

On macOS, this dependency can be easily installed via [Homebrew](https://brew.sh/):

```
brew install go
```

## Usage

With a Makefile in the current directory (or you can specify a different file with the `-f` flag), simply run:

```
go run github.com/theandrew168/make@latest
```

## Features

While not a _complete_ subset of POSIX make, this tool is still useful for utilizing simple Makefiles across all platforms:

- Depends only on the [Go programming language](https://golang.org/dl/)
- Automatic resolution of targets and dependencies
- Executes with maximum concurrency while respecting dependency order
- Implemented in a small and readable [~150 lines of code](https://github.com/theandrew168/make/blob/main/make.go)
- Can bootstrap itself using this project's own [Makefile](https://github.com/theandrew168/make/blob/main/Makefile)
- Built with love from a remote cabin in central Iowa
