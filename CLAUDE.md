# go.hocon2 — HOCON Conversion Tools

## Project Overview

HOCON を他のフォーマットに変換する CLI ツール群。プロジェクト名 `hocon2` は汎用的な変換ハブとしての役割を示し、個々のバイナリ（`hocon2json` 等）が具体的な変換を担う。

## Architecture

```
go.hocon2/
├── cmd/
│   └── hocon2json/    # HOCON → JSON 変換 CLI
│       └── main.go
├── go.mod             # module github.com/o3co/go.hocon2
└── LICENSE            # Apache 2.0
```

### Core Dependency

- `github.com/o3co/go.hocon` — HOCON パーサー（姉妹プロジェクト）
  - `hocon.ParseString(s)` — 文字列から Config を生成
  - `hocon.ParseFile(path)` — ファイルから Config を生成
  - `cfg.Unmarshal(&map[string]any{})` — Config を map に変換（JSON シリアライズに使用）

## Building & Running

```bash
go build ./cmd/hocon2json/

# File input
./hocon2json app.conf

# Stdin input
cat app.conf | ./hocon2json

# Install globally
go install github.com/o3co/go.hocon2/cmd/hocon2json@latest
```

## Design Decisions

- **stdin / file 引数の二通り** — Unix パイプラインと直接ファイル指定の両方をサポート
- **Unmarshal → json.Encode** — go.hocon の既存 API を活用し、Val→map[string]any→JSON の変換パス
- **将来の拡張**: `cmd/hocon2yaml/`, `cmd/hocon2toml/` を追加可能な構成

## Conventions

- License: Apache 2.0
- Go module path: `github.com/o3co/go.hocon2`
- Branch strategy: `master` (release), `develop` (work branch)
- Commit style: conventional commits (`feat:`, `fix:`, `chore:`, etc.)
