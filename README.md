# dockhop

A small terminal UI for picking a running Docker container and dropping into a shell inside it.

## Install

```sh
# via install script
curl -fsSL https://raw.githubusercontent.com/Tvist1988/dockhop/master/install.sh | sh

# or with Go
go install github.com/Tvist1988/dockhop@latest
```

Prebuilt binaries for Linux and macOS (amd64/arm64) are attached to each [release](https://github.com/Tvist1988/dockhop/releases).

## Usage

```sh
dockhop
```

Pick a container from the list and press **Enter** to open a `/bin/sh` shell inside it.

| Key       | Action               |
| --------- | -------------------- |
| `↑` / `↓` | move selection       |
| `/`       | filter the list      |
| `r`       | refresh containers   |
| `Enter`   | shell into container |
| `q`       | quit                 |

Requires a running Docker daemon and the `docker` CLI on your `PATH`. Only running containers are listed.

## Build from source

```sh
go build -o dockhop .
```

## License

Released under the [MIT License](LICENSE).
