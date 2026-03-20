# go.hocon2 — HOCON Conversion Tools

## Project Overview

HOCON を他のフォーマットに変換する CLI ツール群。プロジェクト名 `hocon2` は汎用的な変換ハブとしての役割を示し、個々のバイナリ（`hocon2json` 等）が具体的な変換を担う。

## Architecture

```
go.hocon2/
├── cmd/
│   ├── hocon2json/main.go
│   ├── hocon2yaml/main.go
│   ├── hocon2toml/main.go
│   └── hocon2properties/main.go
├── internal/
│   ├── convert/         # Encoder interface + Run() + format encoders
│   └── flatten/         # map[string]any → map[string]string
├── testdata/            # Golden test data (.hocon + expected outputs)
├── go.mod               # module github.com/o3co/go.hocon2
└── LICENSE              # Apache 2.0
```

### Core Dependency

- `github.com/o3co/go.hocon` — HOCON パーサー（姉妹プロジェクト）
  - `hocon.ParseString(s)` — 文字列から Config を生成
  - `hocon.ParseFile(path)` — ファイルから Config を生成
  - `cfg.Unmarshal(&map[string]any{})` — Config を map に変換（JSON シリアライズに使用）
- `gopkg.in/yaml.v3` — YAML encoding
- `github.com/BurntSushi/toml` — TOML encoding
- `github.com/magiconair/properties` — Properties encoding

## Building & Running

```bash
# Build all
make build

# Or build individually
go build ./cmd/hocon2json/
go build ./cmd/hocon2yaml/
go build ./cmd/hocon2toml/
go build ./cmd/hocon2properties/

# File input
./hocon2json app.conf

# Stdin input
cat app.conf | ./hocon2yaml

# Merge multiple files (last takes precedence)
./hocon2toml base.conf env.conf local.conf

# Install all globally
make install
```

## Design Decisions

- **Encoder interface** — `internal/convert.Encoder` defines `Encode(w io.Writer, data map[string]any) error`. Each format implements this.
- **Shared Run()** — Input parsing, HOCON parsing, merging, and encoding in one function. Each command is a thin wrapper: `convert.Run(name, encoder, ...)`
- **Multi-file merge** — Multiple positional args supported. Right-precedence (last file wins) via `WithFallback`.
- **internal packages** — Not exported. No API stability commitment.
- **stdin / file args** — Unix pipeline and direct file input both supported.

## Conventions

- License: Apache 2.0
- Go module path: `github.com/o3co/go.hocon2`
- Branch strategy: `master` (release), `develop` (work branch)
- Commit style: conventional commits (`feat:`, `fix:`, `chore:`, etc.)
