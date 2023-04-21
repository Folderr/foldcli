# folderr-cli
The Folderr Management CLI

Useful for setting up and updating Folderr.

folderr-cli is a application written in `go` and expects `go` version `1.20.2` or later (if building).

Commands tested on Linux (Ubuntu):

- `folderr-cli`
- `folderr-cli init`
- `folderr-cli install`

## Installation

Not supported yet. Must build instead.

## Building source code into a binary

Prerequestities:
- [go 1.20.2 or later](https://go.dev)

install with

```sh
git clone https://github.com/Folderr/folderr-cli
```

Build with

```sh
# in install directory
go build .
```

Place into your PATH

Find path in your terminal

```sh
$PATH
# Usually /usr/bin, /usr/share/bin, or /usr/local/bin for linux
```

## Usage

On first run use `folderr-cli init` to initialize the cli

This can be done interactively or non-interactively. This is the only command that has interactivity currently

Non-interactive example:
```sh
folderr-cli init /home/folderr/folderr https://github.com/Folderr/Folderr
```

## Contributing

Please use `stacticcheck` for linting Go, and use `go vet` before comitting.
These are ran on pull request and push!

Dev environment is unchanged from other environments.
Please also use `go test ./cmd` before committing (this requires NodeJS 14 or later installed, and NPM).