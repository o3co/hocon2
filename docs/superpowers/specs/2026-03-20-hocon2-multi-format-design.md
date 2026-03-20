# hocon2 Multi-Format Conversion & OSS Bootstrap

## Overview

go.hocon2 を HOCON → JSON 専用ツールから、HOCON → 複数フォーマット変換ハブへ拡張する。同時に OSS プロジェクトとしての体裁を整備する。

## Goals

1. 4フォーマット対応: JSON, YAML, TOML, Properties
2. 共通ロジックの抽出による重複排除
3. OSS としての基盤整備（CI, docs, release automation）

## Non-Goals

- 逆方向の変換（JSON → HOCON 等）
- GUI / Web インターフェース
- go.hocon パーサー自体の変更

---

## Architecture

### Directory Structure

```
go.hocon2/
├── cmd/
│   ├── hocon2json/main.go
│   ├── hocon2yaml/main.go
│   ├── hocon2toml/main.go
│   └── hocon2properties/main.go
├── internal/
│   ├── convert/
│   │   ├── convert.go           # Encoder interface + Run()
│   │   ├── convert_test.go      # Golden tests
│   │   └── integration_test.go  # CLI integration tests
│   └── flatten/
│       ├── flatten.go           # map[string]any → map[string]string
│       └── flatten_test.go
├── testdata/
│   ├── basic.hocon / .json / .yaml / .toml / .properties
│   ├── nested.hocon / .json / .yaml / .toml / .properties
│   └── array.hocon / .json / .yaml / .toml / .properties
├── .github/
│   └── workflows/
│       └── ci.yml
├── .goreleaser.yml
├── Makefile
├── README.md
├── CONTRIBUTING.md
├── LICENSE                      # Apache 2.0 (existing)
├── CLAUDE.md
├── go.mod
└── go.sum
```

### Core Interface

```go
// internal/convert/convert.go
package convert

import "io"

// Encoder encodes structured data to a specific output format.
type Encoder interface {
    Encode(w io.Writer, data map[string]any) error
}

// Run parses HOCON input (from stdin or file) and encodes it using the given Encoder.
func Run(enc Encoder, args []string, stdin io.Reader, stdout, stderr io.Writer) error
```

`Run()` の処理フロー:
1. 引数判定: 0個 → stdin、1個 → file（`-h`/`--help` はヘルプ表示）、2個以上 → エラー
2. `hocon.ParseString()` or `hocon.ParseFile()` で Config を取得
3. `cfg.Unmarshal(&map[string]any{})` で中間表現に変換
4. `enc.Encode(stdout, data)` で出力

### Encoder Implementations

| Command | Encoder | Dependencies |
|---|---|---|
| hocon2json | stdlib `encoding/json` — `NewEncoder` + `SetIndent("", "  ")` | なし |
| hocon2yaml | `gopkg.in/yaml.v3` — `NewEncoder` | yaml.v3 |
| hocon2toml | `github.com/BurntSushi/toml` — `NewEncoder` | toml |
| hocon2properties | `internal/flatten` + `github.com/magiconair/properties` | properties |

各コマンドの `main.go` は Encoder を `convert.Run()` に渡すだけの薄いエントリポイント:

```go
// cmd/hocon2yaml/main.go
package main

import (
    "fmt"
    "os"
    "github.com/o3co/go.hocon2/internal/convert"
)

type yamlEncoder struct{}

func (yamlEncoder) Encode(w io.Writer, data map[string]any) error {
    enc := yaml.NewEncoder(w)
    defer enc.Close()
    return enc.Encode(data)
}

func main() {
    if err := convert.Run(yamlEncoder{}, os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
        fmt.Fprintf(os.Stderr, "hocon2yaml: %v\n", err)
        os.Exit(1)
    }
}
```

### Properties Flattening

Properties は flat な `key=value` 形式のため、ネストした `map[string]any` をフラット化する前処理が必要。

```go
// internal/flatten/flatten.go
package flatten

// Flatten converts a nested map[string]any to a flat map[string]string
// with dot-separated keys.
func Flatten(m map[string]any) map[string]string
```

変換ルール:
- ネストした map → ドット区切り: `{"db": {"host": "localhost"}}` → `db.host=localhost`
- スライス → インデックスキー: `{"items": [1, 2]}` → `items.0=1`, `items.1=2`
- 値の文字列化: `fmt.Sprintf("%v")`
- 空の map/スライス → キーを出力しない

Properties Encoder のデータフロー:
```
map[string]any → flatten.Flatten() → map[string]string → properties.LoadMap() → Write(stdout)
```

---

## Testing Strategy

### 3 Layers

| Layer | Location | Description |
|---|---|---|
| Unit tests | `internal/convert/convert_test.go` | Table-driven golden tests: 各 Encoder × 各 testdata |
| Flatten tests | `internal/flatten/flatten_test.go` | ネスト、配列、空値、型変換のエッジケース |
| CLI integration | `internal/convert/integration_test.go` | `exec.Command` でバイナリを実行し、stdin/file/help/error を検証 |

### Golden Test Pattern

```go
func TestEncoders(t *testing.T) {
    encoders := map[string]convert.Encoder{ /* ... */ }
    testcases := []string{"basic", "nested", "array"}

    for name, enc := range encoders {
        for _, tc := range testcases {
            // input:    testdata/{tc}.hocon
            // expected: testdata/{tc}.{name}
            // actual:   Run(enc, ...) の stdout 出力
            // assert:   actual == expected
        }
    }
}
```

### Test Data Sets

| Name | Coverage |
|---|---|
| basic | 単純な key-value ペア |
| nested | ネストしたオブジェクト |
| array | 配列・リスト |

---

## OSS Infrastructure

### README.md

- プロジェクト概要と対応フォーマット
- インストール方法（`go install ./cmd/hocon2json@latest` 等、4コマンド分）
- 使い方（stdin + file の例）
- ビルド方法（`make all`）
- ライセンス表記

### CI — GitHub Actions (`.github/workflows/ci.yml`)

- トリガー: push (`master`, `develop`) + PR
- Matrix: Go 1.25 × ubuntu-latest, macos-latest
- Steps: `go vet` → `golangci-lint` → `go test` → `go build`
- ステータスチェック名: `test`（branch protection に連動）

### CONTRIBUTING.md

- 開発環境セットアップ手順
- ブランチ戦略: `master`（リリース）/ `develop`（作業）
- コミットスタイル: conventional commits
- PR の作成・レビュープロセス
- テスト実行方法

### Makefile

```makefile
.PHONY: build test lint all install

build:
	go build ./cmd/...

test:
	go test ./...

lint:
	golangci-lint run

all: lint test build

install:
	go install ./cmd/...
```

### GoReleaser (`.goreleaser.yml`)

- 4バイナリそれぞれのビルド定義（`hocon2json`, `hocon2yaml`, `hocon2toml`, `hocon2properties`）
- タグ push で GitHub Actions から自動リリース
- クロスコンパイル: linux/darwin × amd64/arm64

### GoDoc

- `internal/convert` パッケージコメント
- `Encoder` インターフェース、`Run()` 関数のドキュメントコメント
- `internal/flatten` パッケージコメント

---

## Branch & Release Strategy

- **master**: リリースブランチ。branch protection 有効（PR 必須 + CI パス必須）
- **develop**: デフォルト作業ブランチ
- **リリースフロー**: develop → master への PR → マージ → タグ → GoReleaser

---

## Dependencies

| Package | Version | Purpose |
|---|---|---|
| `github.com/o3co/go.hocon` | v0.2.0 | HOCON parser (existing) |
| `gopkg.in/yaml.v3` | latest | YAML encoding |
| `github.com/BurntSushi/toml` | latest | TOML encoding |
| `github.com/magiconair/properties` | latest | Properties encoding |

---

## Design Decisions

1. **独立バイナリ方式**: 各フォーマットごとに独立した `cmd/` エントリポイント。Unix 哲学に沿い、`go install` で必要なものだけインストール可能。
2. **internal パッケージ**: 共通ロジックは `internal/` に配置し、外部公開しない。API 安定化の責務を避ける。
3. **Encoder インターフェース**: 変換ロジックの唯一の拡張ポイント。新フォーマット追加は Encoder 実装 + `cmd/` エントリポイントの追加のみ。
4. **`magiconair/properties` + 自前フラット化**: 実績あるライブラリでエスケープ・エンコーディングを処理し、フラット化ロジックのみ自前実装。
5. **ゴールデンテスト**: testdata で入出力ペアを管理。新フォーマット追加時は期待出力ファイルを追加するだけ。
