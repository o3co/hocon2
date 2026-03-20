# Lightbend Conformance Tests

## Overview

lightbend/config の等価テスト（equiv01-05）を利用し、go.hocon2 の全4フォーマット変換が HOCON 仕様に準拠していることを検証する。

## Goals

1. HOCON の主要機能（コメント、省略構文、置換、include、トリプルクォート等）が正しく変換されることを検証
2. JSON / YAML / TOML / Properties の全フォーマットでテスト
3. go.hocon パーサーの非対応機能をスキップリストで管理

## Non-Goals

- go.hocon パーサー自体のバグ修正（見つけたらフィードバックするが、このプロジェクトのスコープ外）
- mockersf/hocon-test-suite のフル活用（equiv01-05 で十分）

---

## Test Data

### Source

lightbend/config リポジトリ `config/src/test/resources/equiv01/` 〜 `equiv05/` から `.conf` ファイルと `original.json` を `testdata/lightbend/` にコピー。

### Directory Structure

```
testdata/lightbend/
├── equiv01/
│   ├── original.json               # lightbend 参照 JSON
│   ├── comments.conf               # HOCON バリエーション
│   ├── equals.conf
│   ├── no-commas.conf
│   ├── no-root-braces.conf
│   ├── no-whitespace.json
│   ├── omit-colons.conf
│   ├── path-keys.conf
│   ├── properties-style.conf
│   ├── substitutions.conf
│   ├── unquoted.conf
│   ├── expected.json               # go.hocon2 の JSONEncoder 出力形式
│   ├── expected.yaml               # go.hocon2 の YAMLEncoder 出力形式
│   ├── expected.toml               # go.hocon2 の TOMLEncoder 出力形式
│   └── expected.properties         # go.hocon2 の PropertiesEncoder 出力形式
├── equiv02/
│   ├── original.json
│   ├── path-keys.conf
│   ├── path-keys-weird-whitespace.conf
│   ├── expected.json
│   ├── expected.yaml
│   ├── expected.toml
│   └── expected.properties
├── equiv03/
│   ├── original.json
│   ├── includes.conf
│   ├── letters/
│   │   ├── a.conf
│   │   ├── b.json
│   │   ├── c.conf
│   │   ├── c.properties
│   │   └── numbers/
│   │       ├── 1.conf
│   │       └── 2.properties
│   ├── root/
│   │   └── foo.conf
│   ├── expected.json
│   ├── expected.yaml
│   ├── expected.toml
│   └── expected.properties
├── equiv04/
│   ├── original.json
│   ├── missing-substitutions.conf
│   ├── expected.json
│   ├── expected.yaml
│   ├── expected.toml
│   └── expected.properties
└── equiv05/
    ├── original.json
    ├── triple-quotes.conf
    ├── expected.json
    ├── expected.yaml
    ├── expected.toml
    └── expected.properties
```

### Equivalence Test Coverage

| Dir | HOCON Features |
|---|---|
| equiv01 | コメント、`=`/`:`省略、カンマ省略、ルートブレース省略、クォート省略、パスキー、properties形式、置換、JSON互換 |
| equiv02 | パスキー、空白バリエーション |
| equiv03 | `include` ディレクティブ（.conf, .json, .properties からのインクルード、ネストしたサブディレクトリ） |
| equiv04 | オプショナル置換 `${?var}`（未定義変数の省略） |
| equiv05 | トリプルクォート文字列 `"""..."""` |

---

## Test Implementation

### File

```
internal/convert/conformance_test.go
```

### Test Structure

```go
func TestLightbendConformance(t *testing.T) {
    equivDirs := []string{"equiv01", "equiv02", "equiv03", "equiv04", "equiv05"}
    formats := []struct {
        name    string
        encoder convert.Encoder
    }{
        {"json", convert.JSONEncoder{}},
        {"yaml", convert.YAMLEncoder{}},
        {"toml", convert.TOMLEncoder{}},
        {"properties", convert.PropertiesEncoder{}},
    }

    for _, dir := range equivDirs {
        confFiles := findConfFiles(dir) // .conf と .json と .properties を走査（original.json, expected.* を除外）
        for _, confFile := range confFiles {
            for _, f := range formats {
                testName := dir + "/" + confFile + "/" + f.name
                t.Run(testName, func(t *testing.T) {
                    if reason, ok := skipConformance[testName]; ok {
                        t.Skip(reason)
                    }

                    // Run encoder against .conf file
                    var stdout bytes.Buffer
                    err := convert.Run("hocon2"+f.name, f.encoder, []string{confFilePath}, ...)
                    // ...

                    // Compare with expected.{format}
                    expected, _ := os.ReadFile(expectedPath)
                    if stdout.String() != string(expected) {
                        t.Errorf(...)
                    }
                })
            }
        }
    }
}
```

### Comparison Method

全フォーマットで文字列完全一致。期待出力は go.hocon2 の各 Encoder が実際に出す形式で事前生成するため、セマンティック比較は不要。

### Skip List

```go
var skipConformance = map[string]string{
    // 例: go.hocon が未対応の機能、または特定フォーマットで表現不可能なケース
    // "equiv04/missing-substitutions.conf/toml": "empty object not representable in TOML",
}
```

スキップ理由を明記。go.hocon が対応したら外す。

### Expected Output Generation

期待出力ファイル（`expected.json`, `expected.yaml`, `expected.toml`, `expected.properties`）は実装時に以下の手順で生成:

1. `original.json` を `encoding/json` でパース → `map[string]any`
2. 各 Encoder で書き出し → `expected.{format}` として保存
3. コミットして管理

TOML で表現できないデータ（null 値、混合型配列）がある場合はそのフォーマットの期待出力を作らず、スキップリストに追加。

---

## Helper Function

```go
func findConfFiles(dir string) []string
```

指定ディレクトリ内の `.conf`, `.json`, `.properties` ファイルを走査。ただし以下を除外:
- `original.json`
- `expected.*`
- サブディレクトリ内のファイル（equiv03 の `letters/`, `root/` 等は include 用であり、テスト入力ではない）

equiv03 のテスト入力は `includes.conf` のみ（他のファイルは include で参照される）。

---

## Design Decisions

1. **submodule 不使用**: lightbend の equiv ファイルを直接コピー。5ディレクトリと少量のファイルなので、外部依存を避ける方がシンプル。
2. **文字列完全一致**: go.hocon2 のエンコーダーで期待出力を生成するため、フォーマットの差異が発生しない。
3. **スキップリストは理由付き**: なぜスキップするか明記し、将来の対応時に判断できるようにする。
4. **equiv03 の include テスト**: `includes.conf` のみがテスト入力。サブディレクトリ構造ごとコピーして include が動作することを検証。
