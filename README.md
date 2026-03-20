# hocon2

[![CI](https://github.com/o3co/hocon2/actions/workflows/ci.yml/badge.svg)](https://github.com/o3co/hocon2/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/o3co/hocon2.svg)](https://pkg.go.dev/github.com/o3co/hocon2)
[![Release](https://img.shields.io/github/v/release/o3co/hocon2)](https://github.com/o3co/hocon2/releases/latest)
[![License](https://img.shields.io/github/license/o3co/hocon2)](LICENSE)

HOCON conversion tools for Go — convert [HOCON](https://github.com/lightbend/config/blob/main/HOCON.md) configuration files to JSON, YAML, TOML, and Java Properties.

HOCON (Human-Optimized Config Object Notation) is a superset of JSON designed for human readability. It supports comments, substitutions (`${var}`), includes, omitted quotes/commas, and more. `hocon2` lets you convert HOCON files to widely-supported formats for use with tools that don't natively understand HOCON.

Powered by [go.hocon](https://github.com/o3co/go.hocon) parser. Conformance tested against [Lightbend's reference test suite](https://github.com/lightbend/config).

## Supported Formats

| Command | Output Format |
|---|---|
| `hocon2json` | JSON |
| `hocon2yaml` | YAML |
| `hocon2toml` | TOML |
| `hocon2properties` | Java Properties |

## Install

### Go

```bash
go install github.com/o3co/hocon2/cmd/hocon2json@latest
go install github.com/o3co/hocon2/cmd/hocon2yaml@latest
go install github.com/o3co/hocon2/cmd/hocon2toml@latest
go install github.com/o3co/hocon2/cmd/hocon2properties@latest
```

### Binary releases

Download pre-built binaries from the [releases page](https://github.com/o3co/hocon2/releases/latest) (Linux/macOS, amd64/arm64).

## Usage

### Basic conversion

```bash
# Convert a file
hocon2json app.conf

# Read from stdin
cat app.conf | hocon2yaml

# Show help
hocon2json --help
```

### Multiple file merge

Multiple files can be passed as arguments. They are merged with **right-precedence** — the last file wins for conflicting keys:

```bash
hocon2toml base.conf env.conf local.conf
```

This is equivalent to `local.conf` overriding `env.conf`, which overrides `base.conf`. Useful for layered configuration (base → environment → local overrides).

### Environment variables

HOCON substitutions (`${VAR}`) resolve against environment variables:

```bash
# Pass environment variables inline
DB_HOST=prod-db.example.com hocon2json app.conf

# Or export them
export DB_HOST=prod-db.example.com
hocon2json app.conf
```

Given `app.conf`:

```hocon
database {
  host = ${DB_HOST}
  host = ${?DB_HOST}  # optional: use only if DB_HOST is set
}
```

### Example

Given `app.conf`:

```hocon
database {
  host = "localhost"
  port = 5432
  pool_size = 10
}

// Substitution
api_url = "https://"${database.host}":8080"
```

```bash
$ hocon2json app.conf
{
  "api_url": "https://localhost:8080",
  "database": {
    "host": "localhost",
    "pool_size": 10,
    "port": 5432
  }
}
```

## Build

```bash
make all      # vet + test + build
make build    # build only
make test     # test only
make install  # install all binaries
```

## Related Projects

- [go.hocon](https://github.com/o3co/go.hocon) — HOCON parser for Go (used by this project)

## License

Apache 2.0 — see [LICENSE](LICENSE).
