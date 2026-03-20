# CLI Options & Project Polish — Iteration 2 Design

## Overview

hocon2 CLI ツール群に出力オプション、出力ファイル指定、エラーメッセージ確認、CHANGELOG、Windows 対応を追加する。

## Scope

1. **出力オプション** — JSON のみ。`-compact` と `-indent N`
2. **エラーメッセージ確認** — go.hocon のパースエラーに行番号が含まれるか確認。不足なら go.hocon 側 Issue
3. **`-o` フラグ** — 出力ファイル指定。`-overwrite` で既存ファイル上書き許可
4. **CHANGELOG** — Keep a Changelog 形式
5. **Windows 対応** — GoReleaser に `windows` 追加

## Design

### 1. フラグ体系と `Run()` の変更

**フラグパーサー:** Go 標準 `flag` パッケージ（`flag.FlagSet`）を使用。外部依存なし。

**Run() の変更:**

`Run()` 内で `flag.FlagSet` を作成し、共通フラグを登録する。エンコーダー固有フラグは `FlagRegistrar` インターフェース経由で登録。

```go
func Run(name string, enc Encoder, args []string, stdin io.Reader, stdout, stderr io.Writer) error {
    fs := flag.NewFlagSet(name, flag.ContinueOnError)
    fs.SetOutput(stderr) // フラグパースエラーは stderr へ

    var outFile string
    var overwrite bool
    fs.StringVar(&outFile, "o", "", "output file path")
    fs.BoolVar(&overwrite, "overwrite", false, "overwrite existing output file")

    // エンコーダー固有フラグの登録
    if fr, ok := enc.(FlagRegistrar); ok {
        fr.RegisterFlags(fs)
    }

    // カスタム Usage — ヘルプは stdout に出す（Unix 慣習: help → stdout, error → stderr）
    fs.Usage = func() { printUsage(fs, name, stdout) }

    if err := fs.Parse(args); err != nil {
        if errors.Is(err, flag.ErrHelp) {
            return nil // -h/--help は Usage 出力済み、正常終了
        }
        return err
    }
    // fs.Args() が位置引数（HOCON ファイルパス）
}
```

**`flag.ErrHelp` の処理:** `flag.ContinueOnError` モードでは `-h`/`--help` 時に `flag.ErrHelp` が返る。これを `nil` に変換して正常終了（exit 0）とする。Usage 関数は `fs.Parse` 内で自動呼び出し済み。

**ヘルプ出力先:** `fs.SetOutput(stderr)` はフラグパースエラーの出力先。`fs.Usage` クロージャ内では `stdout` を使い、Unix 慣習（help → stdout, exit 0）に従う。`printUsage(fs, name, stdout)` は `fs`（フラグ一覧表示用）、`name`（コマンド名）、`stdout`（出力先）を受け取るクロージャ。

**FlagRegistrar インターフェース:**

```go
type FlagRegistrar interface {
    RegisterFlags(fs *flag.FlagSet)
}
```

オプショナルインターフェース。エンコーダーが自身のフラグを登録する手段。JSON エンコーダーのみ実装。

**出力先の決定ロジック:**

| 条件 | 挙動 |
|------|------|
| `-o` 未指定 | stdout（現行どおり） |
| `-o` 指定 + ファイル存在しない | 新規作成 |
| `-o` 指定 + ファイル存在 + `-overwrite` あり | 上書き |
| `-o` 指定 + ファイル存在 + `-overwrite` なし | エラー |

`-overwrite` は `-o` なしで指定された場合は無視（エラーにしない）。

**`-o` の追加仕様:**

- ファイルは `0644` パーミッションで作成
- 親ディレクトリが存在しない場合はエラー（自動作成しない）
- `-o -` のような特殊値は扱わない（リテラルなファイル名として処理）

### 2. JSON エンコーダーの出力オプション

**JSONEncoder の変更:**

```go
type JSONEncoder struct {
    Compact bool
    Indent  int // デフォルト 2
}
```

値型からポインタ型に変更（`RegisterFlags` でフィールドを書き換えるため）。

```go
// cmd/hocon2json/main.go
convert.Run("hocon2json", &convert.JSONEncoder{}, os.Args[1:], ...)
```

**FlagRegistrar 実装:**

```go
func (e *JSONEncoder) RegisterFlags(fs *flag.FlagSet) {
    fs.BoolVar(&e.Compact, "compact", false, "output compact JSON")
    fs.IntVar(&e.Indent, "indent", 2, "indentation width")
}
```

**Encode の挙動:**

- `-compact` → インデントなし、改行なし
- `-indent N` → N スペースインデント（有効範囲: 1〜16。範囲外はエラー）
- `-indent 0` は許可しない（compact と紛らわしいため `-compact` を使うこと）
- `-compact` + `-indent` 両方指定 → `-compact` が優先

**他のエンコーダー:** 変更なし。`FlagRegistrar` を実装しない。

### 3. ヘルプ表示

`flag.FlagSet` の `Usage` をカスタマイズ。`-h`/`--help` の手動チェックは削除し、`flag.FlagSet` に任せる。

**hocon2json の場合:**

```
Usage: hocon2json [OPTIONS] [FILE...]

Convert HOCON to JSON.

If no FILE is given, reads from stdin.
If multiple FILEs are given, they are merged (last file takes precedence).

Options:
  -compact      output compact JSON
  -indent int   indentation width (default 2)
  -o string     output file path
  -overwrite    overwrite existing output file
```

**hocon2yaml の場合（FlagRegistrar なし）:**

```
Usage: hocon2yaml [OPTIONS] [FILE...]

Convert HOCON to YAML.

If no FILE is given, reads from stdin.
If multiple FILEs are given, they are merged (last file takes precedence).

Options:
  -o string     output file path
  -overwrite    overwrite existing output file
```

フォーマット固有オプションは `FlagRegistrar` 経由で登録されたものだけ表示される。

### 4. エラーメッセージ

- go.hocon の `ParseString` / `ParseFile` が返すエラーに行番号情報が含まれるか確認
- 含まれていれば現状で OK（hocon2 側の変更なし）
- 含まれていなければ go.hocon リポジトリに Issue を作成

### 5. CHANGELOG

`CHANGELOG.md` を Keep a Changelog 形式で作成。

```markdown
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/).

## [Unreleased]

### Added
- `-compact` and `-indent` options for JSON output formatting
- `-o` output file option
- `-overwrite` flag for allowing overwrite of existing output files
- Windows binary releases
- CHANGELOG

## [0.1.0] - 2026-03-20
(過去リリース分は git log から記載)
```

### 6. Windows 対応

`.goreleaser.yml` の全ビルドの `goos` に `windows` を追加。

```yaml
goos: [linux, darwin, windows]
```

Windows では `tar.gz` が不便なため、`archives` セクションで Windows 向けに `zip` を生成する。

```yaml
archives:
  - format: tar.gz
    format_overrides:
      - goos: windows
        format: zip
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
```

## 影響範囲

| ファイル | 変更内容 |
|---------|---------|
| `internal/convert/convert.go` | `Run()` に flag.FlagSet 導入、`-o`/`-overwrite` 処理、`FlagRegistrar` インターフェース追加 |
| `internal/convert/json.go` | `JSONEncoder` にフィールド追加、`FlagRegistrar` 実装、`Encode` のインデント制御 |
| `cmd/hocon2json/main.go` | `JSONEncoder` をポインタ渡しに変更 |
| `internal/convert/run_test.go` | 新フラグのテスト追加 |
| `internal/convert/convert_test.go` | 既存ゴールデンテストが引き続きパスすることを確認 |
| `internal/convert/integration_test.go` | CLI フラグの E2E テスト追加 |
| `.goreleaser.yml` | `windows` 追加、`zip` フォーマットオーバーライド |
| `CHANGELOG.md` | 新規作成 |

## テスト戦略

- **既存テスト:** すべてそのままパスすること（デフォルト挙動が変わらないため）
- **新規ユニットテスト:** `-compact`, `-indent`, `-o`, `-overwrite` の各フラグ
- **新規統合テスト:** CLI バイナリ経由でのフラグ動作確認
- **エッジケース:** `-compact` + `-indent` 同時指定、`-o` なし + `-overwrite`、存在しないディレクトリへの `-o`
