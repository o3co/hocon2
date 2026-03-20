# hocon2

HOCON conversion tools — convert [HOCON](https://github.com/lightbend/config/blob/main/HOCON.md) configuration files to other formats.

## Supported Formats

| Command | Output Format |
|---|---|
| `hocon2json` | JSON |
| `hocon2yaml` | YAML |
| `hocon2toml` | TOML |
| `hocon2properties` | Java Properties |

## Install

```bash
go install github.com/o3co/go.hocon2/cmd/hocon2json@latest
go install github.com/o3co/go.hocon2/cmd/hocon2yaml@latest
go install github.com/o3co/go.hocon2/cmd/hocon2toml@latest
go install github.com/o3co/go.hocon2/cmd/hocon2properties@latest
```

## Usage

```bash
# Convert a file
hocon2json app.conf

# Read from stdin
cat app.conf | hocon2yaml

# Merge multiple files (last file takes precedence)
hocon2toml base.conf env.conf local.conf
```

## Build

```bash
make all      # lint + test + build
make build    # build only
make test     # test only
make install  # install all binaries
```

## License

Apache 2.0 — see [LICENSE](LICENSE).
