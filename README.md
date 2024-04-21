# foldcli
The Folderr Management CLI

## More up-to-date documentation

More up to date documentation can be found https://folderr.net/guides/cli/getting-started

Useful for setting up and updating Folderr.

foldcli is a application written in `go` and expects `go` version `1.20` or later (if building).

Commands tested on Linux (Ubuntu):

- `foldcli`
- `foldcli init folderr`
- `foldcli install folderr`
- `foldcli setup db`

## Installation

Most up to date version will be from [building from source](#building-source-code-into-a-binary)

1. Grab from latest tag
2. Add to path
3. Reload any terminal you wish to use on.

## Building source code into a binary

Prerequestities:
- [go 1.20 or later](https://go.dev)

install with

```sh
git clone https://github.com/Folderr/foldcli
```

Build with

```sh
# in install directory
go build .
```

Place into your PATH

(Find path in your terminal)

```sh
$PATH
# Usually /usr/bin, /usr/share/bin, or /usr/local/bin for linux
# Alternatively for linux: $HOME/.local/bin
```

## Usage

On first run use `foldcli init` to initialize the cli

This can be done interactively or non-interactively. This is the only command that has interactivity currently

Non-interactive example:
```sh
foldcli init /home/folderr/folderr https://github.com/Folderr/Folderr
```

## Contributing

Please use `staticcheck` for linting Go, and use `go vet` before comitting.
These are ran on pull request and push!
- Uses staticcheck 2023.1.3 in CI

Dev environment is unchanged from other environments.
Please also use `go test ./cmd` before committing (this requires NodeJS 14 or later installed, and NPM).