# hocon2

[![CI](https://github.com/o3co/hocon2/actions/workflows/ci.yml/badge.svg)](https://github.com/o3co/hocon2/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/o3co/hocon2.svg)](https://pkg.go.dev/github.com/o3co/hocon2)
[![Release](https://img.shields.io/github/v/release/o3co/hocon2)](https://github.com/o3co/hocon2/releases/latest)
[![License](https://img.shields.io/github/license/o3co/hocon2)](LICENSE)

Go 向け HOCON 変換ツール — [HOCON](https://github.com/lightbend/config/blob/main/HOCON.md) 設定ファイルを JSON、YAML、TOML、Java Properties に変換します。

HOCON (Human-Optimized Config Object Notation) は、人間にとって読みやすい JSON のスーパーセットです。コメント、変数参照（`${var}`）、インクルード、クォートやカンマの省略などをサポートしています。`hocon2` を使えば、HOCON をネイティブに扱えないツール向けに、広くサポートされたフォーマットへ変換できます。

パーサーには [go.hocon](https://github.com/o3co/go.hocon) を使用。[Lightbend のリファレンステストスイート](https://github.com/lightbend/config)で準拠性を検証済みです。

> **[Claude Code](https://claude.ai/claude-code)** (Anthropic) により設計・実装されました。

## 対応フォーマット

| コマンド | 出力フォーマット |
|---|---|
| `hocon2json` | JSON |
| `hocon2yaml` | YAML |
| `hocon2toml` | TOML |
| `hocon2properties` | Java Properties |

## インストール

### Go

```bash
go install github.com/o3co/hocon2/cmd/hocon2json@latest
go install github.com/o3co/hocon2/cmd/hocon2yaml@latest
go install github.com/o3co/hocon2/cmd/hocon2toml@latest
go install github.com/o3co/hocon2/cmd/hocon2properties@latest
```

### バイナリリリース

[リリースページ](https://github.com/o3co/hocon2/releases/latest)からビルド済みバイナリをダウンロードできます（Linux/macOS、amd64/arm64）。

## 使い方

### 基本的な変換

```bash
# ファイルを変換
hocon2json app.conf

# 標準入力から読み込み
cat app.conf | hocon2yaml

# ヘルプを表示
hocon2json --help
```

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
make all      # vet + test + build
make build    # ビルドのみ
make test     # テストのみ
make install  # 全バイナリをインストール
```

## 関連プロジェクト

- [go.hocon](https://github.com/o3co/go.hocon) — Go 向け HOCON パーサー（本プロジェクトで使用）

## ライセンス

Apache 2.0 — [LICENSE](LICENSE) を参照。
