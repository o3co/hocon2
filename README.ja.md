# hocon2 — Go 向け HOCON 変換 CLI

[![CI](https://github.com/o3co/hocon2/actions/workflows/ci.yml/badge.svg)](https://github.com/o3co/hocon2/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/o3co/hocon2.svg)](https://pkg.go.dev/github.com/o3co/hocon2)
[![Release](https://img.shields.io/github/v/release/o3co/hocon2)](https://github.com/o3co/hocon2/releases/latest)
[![License](https://img.shields.io/github/license/o3co/hocon2)](LICENSE)

Go 向け HOCON 変換ツール — [HOCON](https://github.com/lightbend/config/blob/main/HOCON.md) 設定ファイルを JSON、YAML、TOML、Java Properties に変換します。

HOCON (Human-Optimized Config Object Notation) は、人間にとって読みやすい JSON のスーパーセットです。コメント、変数参照（`${var}`）、インクルード、クォートやカンマの省略などをサポートしています。`hocon2` を使えば、HOCON をネイティブに扱えないツール向けに、広くサポートされたフォーマットへ変換できます。

パーサーには [go.hocon](https://github.com/o3co/go.hocon) を使用。[Lightbend のリファレンステストスイート](https://github.com/lightbend/config)で準拠性を検証済みです。

**ドキュメント:** [o3co.github.io/hocon](https://o3co.github.io/hocon/) — 4 実装・hocon2 CLI・共通の準拠コーパスをまとめた横断ハブです。

> **[Claude Code](https://claude.ai/claude-code)** (Anthropic) により設計・実装されました。
> [GitHub Copilot](https://github.com/features/copilot) および [OpenAI Codex](https://openai.com/index/openai-codex/) によるレビュー。

[English](README.md)

---

## クイックスタート

### 1. インストール

#### Go

```bash
go install github.com/o3co/hocon2/cmd/hocon2json@latest
go install github.com/o3co/hocon2/cmd/hocon2yaml@latest
go install github.com/o3co/hocon2/cmd/hocon2toml@latest
go install github.com/o3co/hocon2/cmd/hocon2properties@latest

# 逆方向: 他フォーマット -> HOCON (v0.7.0)
go install github.com/o3co/hocon2/cmd/json2hocon@latest
go install github.com/o3co/hocon2/cmd/yaml2hocon@latest
go install github.com/o3co/hocon2/cmd/toml2hocon@latest
go install github.com/o3co/hocon2/cmd/properties2hocon@latest
```

#### バイナリリリース

[リリースページ](https://github.com/o3co/hocon2/releases/latest)からビルド済みバイナリをダウンロードできます（Linux/macOS/Windows、amd64/arm64）。

### 2. 使い方

```bash
# ファイルを変換
hocon2json app.conf

# 標準入力から読み込み
cat app.conf | hocon2yaml

# ヘルプを表示
hocon2json --help
```

## なぜ hocon2？

HOCON は設定の記述に最適ですが、多くのツールは JSON、YAML、TOML しか理解しません。`hocon2` がそのギャップを埋めます：

- **記述** は HOCON で（可読性が高く、組み合わせ可能で DRY）
- **デプロイ** はツールが必要とするフォーマットで（Kubernetes は JSON、Helm は YAML、Rust ツールは TOML、Java は Properties）
- **検証** はデプロイ前に CI で構文チェック（`-validate` フラグ）
- **取り込み** 既存の JSON/YAML/TOML/Properties ファイルを、逆方向の `*2hocon` コマンドで HOCON に変換

## 対応フォーマット

HOCON から他のフォーマットへ:

| コマンド | 出力フォーマット |
|---|---|
| `hocon2json` | JSON |
| `hocon2yaml` | YAML |
| `hocon2toml` | TOML |
| `hocon2properties` | Java Properties |

他のフォーマットから HOCON へ (v0.7.0):

| コマンド | 入力フォーマット |
|---|---|
| `json2hocon` | JSON |
| `yaml2hocon` | YAML |
| `toml2hocon` | TOML |
| `properties2hocon` | Java Properties |

逆方向のコマンドは、入力ファイルを読んで値のツリーを作り、慣用的な HOCON として出力します。外部のデータはあくまでデータで、入力値の中の `${...}` は解決されずにそのまま出力されます。値の型は往復しても保たれます — Properties の値や JSON の文字列 `"8080"` は、数値ではなく文字列として読み直されるようにクォートされます。読み込みは [go.hocon のアダプタ](https://github.com/o3co/go.hocon/tree/develop/adapters)と同じフォーマット変換規則を使うので（TOML の日時は文字列に、YAML は 1.2 core schema に従い、UTF-8 の BOM は除去）、両者がずれることはありません。

YAML のスカラーの解決は YAML ライブラリの答えであって、ここでの保証ではありません。クォートなしの `010` が `8` になるか `10` になるかはパーサー次第です。そのまま残したい値はクォートしてください。

## 使い方

### オプション

```bash
# コンパクト JSON 出力（空白なし）
hocon2json -compact app.conf

# インデント幅を指定（デフォルト: 2）
hocon2json -indent 4 app.conf

# ファイルに出力
hocon2json -o output.json app.conf

# 既存ファイルを上書き
hocon2json -o output.json -overwrite app.conf
```

`-compact` と `-indent` は `hocon2json` のみで使用可能。`-o` と `-overwrite` は全コマンド共通です。

### 複数ファイルのマージ

引数に複数ファイルを指定できます。**右優先**でマージされます — 後に指定したファイルのキーが優先されます：

```bash
hocon2toml base.conf env.conf local.conf
```

これは `local.conf` が `env.conf` を、`env.conf` が `base.conf` を上書きする動作です。レイヤード構成（ベース → 環境別 → ローカル上書き）に便利です。

### 環境変数

HOCON の変数参照（`${VAR}`）は環境変数を解決します：

```bash
# 環境変数をインラインで渡す
DB_HOST=prod-db.example.com hocon2json app.conf

# または export する
export DB_HOST=prod-db.example.com
hocon2json app.conf
```

`app.conf`:

```hocon
database {
  host = ${DB_HOST}
  host = ${?DB_HOST}  # オプショナル: DB_HOST が設定されている場合のみ使用
}
```

### 変換例

`app.conf`:

```hocon
database {
  host = "localhost"
  port = 5432
  pool_size = 10
}

// 変数参照
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

## ビルド

```bash
make all      # lint + test + build
make build    # ビルドのみ
make test     # テストのみ
make install  # 全バイナリをインストール
```

## 関連プロジェクト

| プロジェクト | 言語 | レジストリ | 説明 |
|---------|----------|----------|-------------|
| [go.hocon](https://github.com/o3co/go.hocon) | Go | [pkg.go.dev](https://pkg.go.dev/github.com/o3co/go.hocon) | Go 向け HOCON パーサー（本プロジェクトで使用） |
| [ts.hocon](https://github.com/o3co/ts.hocon) | TypeScript | [npm](https://www.npmjs.com/package/@o3co/ts.hocon) | TypeScript/Node.js 向け HOCON パーサー |
| [rs.hocon](https://github.com/o3co/rs.hocon) | Rust | [crates.io](https://crates.io/crates/hocon-parser) | Rust 向け HOCON パーサー |
| [py.hocon](https://github.com/o3co/py.hocon) | Python | [PyPI](https://pypi.org/project/hocon-parser/) | Python 向け HOCON パーサー |

4 つのパーサー実装は同じ Lightbend HOCON 仕様で追跡されています — 実装ごとの準拠率は[実装横断のロールアップ](https://github.com/o3co/xx.hocon/blob/main/docs/compliance-matrix.md)を参照してください。

## ベストプラクティス

### CI/CD 統合

- CI パイプラインでデプロイ前に `hocon2json -validate` で HOCON の構文をチェックしましょう
- `-env-file` を使えばシェル環境を汚さずに環境別変数を注入できます

### CI での設定バリデーション

```yaml
# 例: GitHub Actions
- name: Validate config
  run: hocon2json -validate config/prod.conf
```

### 複数ファイルのマージ

- マージ順序が重要: 後のファイルが前のファイルを上書きします
- `base.conf` + `env.conf` パターンで環境別のオーバーライドを行いましょう:

  ```bash
  hocon2json base.conf prod.conf > config.json
  ```

## ライセンス

Apache 2.0 — [LICENSE](LICENSE) を参照。
