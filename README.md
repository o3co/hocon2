# hocon2 — HOCON Conversion CLI for Go

[![CI](https://github.com/o3co/hocon2/actions/workflows/ci.yml/badge.svg)](https://github.com/o3co/hocon2/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/o3co/hocon2.svg)](https://pkg.go.dev/github.com/o3co/hocon2)
[![Release](https://img.shields.io/github/v/release/o3co/hocon2)](https://github.com/o3co/hocon2/releases/latest)
[![License](https://img.shields.io/github/license/o3co/hocon2)](LICENSE)

HOCON conversion tools for Go — convert [HOCON](https://github.com/lightbend/config/blob/main/HOCON.md) configuration files to JSON, YAML, TOML, and Java Properties.

HOCON (Human-Optimized Config Object Notation) is a superset of JSON designed for human readability. It supports comments, substitutions (`${var}`), includes, omitted quotes/commas, and more. `hocon2` lets you convert HOCON files to widely-supported formats for use with tools that don't natively understand HOCON.

Powered by [go.hocon](https://github.com/o3co/go.hocon) parser. Conformance tested against [Lightbend's reference test suite](https://github.com/lightbend/config).

> **Implemented by [Claude Code](https://claude.ai/claude-code)** (Anthropic) — designed and built end-to-end with Claude Code.
> Reviewed by [GitHub Copilot](https://github.com/features/copilot) and [OpenAI Codex](https://openai.com/index/openai-codex/).

## Quick Start

### 1. Install

#### Go

```bash
go install github.com/o3co/hocon2/cmd/hocon2json@latest
go install github.com/o3co/hocon2/cmd/hocon2yaml@latest
go install github.com/o3co/hocon2/cmd/hocon2toml@latest
go install github.com/o3co/hocon2/cmd/hocon2properties@latest

# Reverse: foreign format -> HOCON (v0.7.0)
go install github.com/o3co/hocon2/cmd/json2hocon@latest
go install github.com/o3co/hocon2/cmd/yaml2hocon@latest
go install github.com/o3co/hocon2/cmd/toml2hocon@latest
go install github.com/o3co/hocon2/cmd/properties2hocon@latest
```

#### Binary releases

Download pre-built binaries from the [releases page](https://github.com/o3co/hocon2/releases/latest) (Linux/macOS/Windows, amd64/arm64).

### 2. Use

```bash
# Convert a file
hocon2json app.conf

# Read from stdin
cat app.conf | hocon2yaml

# Show help
hocon2json --help
```

## Why hocon2?

HOCON is great for authoring config, but many tools only understand JSON, YAML, or TOML. `hocon2` bridges this gap:

- **Write** config in HOCON (readable, composable, DRY)
- **Deploy** in whatever format your tools need (JSON for Kubernetes, YAML for Helm, TOML for Rust tools, Properties for Java)
- **Validate** syntax in CI before deployment (`-validate` flag)
- **Import** an existing JSON/YAML/TOML/Properties file into HOCON with the reverse `*2hocon` commands

## Supported Formats

HOCON to another format:

| Command | Output Format |
|---|---|
| `hocon2json` | JSON |
| `hocon2yaml` | YAML |
| `hocon2toml` | TOML |
| `hocon2properties` | Java Properties |

Another format to HOCON (v0.7.0):

| Command | Input Format |
|---|---|
| `json2hocon` | JSON |
| `yaml2hocon` | YAML |
| `toml2hocon` | TOML |
| `properties2hocon` | Java Properties |

The reverse commands read the foreign file, build a value tree, and render idiomatic
HOCON. Foreign data stays data: a `${...}` in an input value is emitted literally, not
resolved. Values keep their type across the round trip — a Properties value or a JSON
string `"8080"` is quoted so it re-parses as a string, not a number. Reading uses the
same format-mapping rules as the [go.hocon adapters](https://github.com/o3co/go.hocon/tree/main/adapters)
(TOML dates become strings, YAML follows the 1.2 core schema, a UTF-8 BOM is stripped),
so the two never drift.

YAML scalar resolution is the YAML library's answer, not a guarantee here: whether a
bare `010` is `8` or `10` depends on the parser. Quote values that must survive
verbatim.

## Usage

### Options

```bash
# Compact JSON output (no whitespace)
hocon2json -compact app.conf

# Custom indentation width (default: 2)
hocon2json -indent 4 app.conf

# Write output to a file
hocon2json -o output.json app.conf

# Overwrite an existing output file
hocon2json -o output.json -overwrite app.conf
```

`-compact` and `-indent` are available for `hocon2json` only. `-o` and `-overwrite` work with all commands.

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

| Project | Language | Registry | Description |
|---------|----------|----------|-------------|
| [go.hocon](https://github.com/o3co/go.hocon) | Go | [pkg.go.dev](https://pkg.go.dev/github.com/o3co/go.hocon) | HOCON parser for Go (used by this project) |
| [ts.hocon](https://github.com/o3co/ts.hocon) | TypeScript | [npm](https://www.npmjs.com/package/@o3co/ts.hocon) | HOCON parser for TypeScript/Node.js |
| [rs.hocon](https://github.com/o3co/rs.hocon) | Rust | [crates.io](https://crates.io/crates/o3co-hocon) | HOCON parser for Rust |

All implementations are full Lightbend HOCON spec compliant.

## Best Practices

### CI/CD Integration

- Use `hocon2json -validate` to check HOCON syntax in CI pipelines before deployment
- Use `-env-file` to inject environment-specific variables without polluting the shell environment

### Config Validation in CI

```yaml
# Example: GitHub Actions
- name: Validate config
  run: hocon2json -validate config/prod.conf
```

### Multi-File Merging

- Merge order matters: later files override earlier ones
- Use `base.conf` + `env.conf` pattern for environment-specific overrides:

  ```bash
  hocon2json base.conf prod.conf > config.json
  ```

## License

Apache 2.0 — see [LICENSE](LICENSE).
