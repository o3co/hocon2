# hocon2 Multi-Format Conversion & OSS Bootstrap — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Expand go.hocon2 from a single hocon2json tool into a multi-format HOCON conversion hub (JSON, YAML, TOML, Properties) with full OSS infrastructure.

**Architecture:** Shared `internal/convert` package with `Encoder` interface and `Run()` function. Each format implements `Encoder`; each `cmd/` binary is a thin wrapper calling `convert.Run()`. `internal/flatten` handles map flattening for Properties. TDD with golden test files in `testdata/`.

**Tech Stack:** Go 1.25, `github.com/o3co/go.hocon` v0.2.0, `gopkg.in/yaml.v3`, `github.com/BurntSushi/toml`, `github.com/magiconair/properties`, GitHub Actions, GoReleaser, golangci-lint

**Spec:** `docs/superpowers/specs/2026-03-20-hocon2-multi-format-design.md`

---

## File Map

### New Files
- `internal/convert/convert.go` — `Encoder` interface + `Run()` function
- `internal/convert/convert_test.go` — Golden tests for all encoders
- `internal/convert/json.go` — `JSONEncoder`
- `internal/convert/yaml.go` — `YAMLEncoder`
- `internal/convert/toml.go` — `TOMLEncoder`
- `internal/convert/properties.go` — `PropertiesEncoder`
- `internal/convert/integration_test.go` — CLI integration tests
- `internal/flatten/flatten.go` — `Flatten()` function
- `internal/flatten/flatten_test.go` — Flatten unit tests
- `testdata/basic.hocon`, `.json`, `.yaml`, `.toml`, `.properties`
- `testdata/nested.hocon`, `.json`, `.yaml`, `.toml`, `.properties`
- `testdata/array.hocon`, `.json`, `.yaml`, `.toml`, `.properties`
- `testdata/substitution.hocon`, `.json`, `.yaml`, `.toml`, `.properties`
- `cmd/hocon2yaml/main.go`
- `cmd/hocon2toml/main.go`
- `cmd/hocon2properties/main.go`
- `.github/workflows/ci.yml`
- `.github/workflows/release.yml`
- `.goreleaser.yml`
- `Makefile`
- `README.md`
- `CONTRIBUTING.md`

### Modified Files
- `cmd/hocon2json/main.go` — Refactor to use `convert.Run()`
- `go.mod` — Add new dependencies
- `CLAUDE.md` — Update architecture section
- `.gitignore` — Add new binary names

---

## Task 1: Encoder Interface + Run() with JSON Encoder (TDD)

**Files:**
- Create: `internal/convert/convert.go`
- Create: `internal/convert/json.go`
- Create: `internal/convert/convert_test.go`
- Create: `testdata/basic.hocon`
- Create: `testdata/basic.json`

- [ ] **Step 1: Create testdata/basic.hocon**

```hocon
# testdata/basic.hocon
name = "hocon2"
version = 1
enabled = true
```

- [ ] **Step 2: Create testdata/basic.json (expected output)**

```json
{
  "enabled": true,
  "name": "hocon2",
  "version": 1
}
```

Note: `json.Encoder` sorts keys alphabetically in Go when using `map[string]any`.

- [ ] **Step 3: Write Encoder interface and Run() skeleton**

```go
// internal/convert/convert.go
package convert

import (
	"fmt"
	"io"

	"github.com/o3co/go.hocon"
)

// Encoder encodes structured data to a specific output format.
type Encoder interface {
	Encode(w io.Writer, data map[string]any) error
}

// Run parses HOCON input and encodes it using the given Encoder.
func Run(name string, enc Encoder, args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	cfg, err := parseInput(name, args, stdin, stdout, stderr)
	if cfg == nil || err != nil {
		return err
	}

	var m map[string]any
	if err := cfg.Unmarshal(&m); err != nil {
		return fmt.Errorf("unmarshaling config: %w", err)
	}

	if err := enc.Encode(stdout, m); err != nil {
		return fmt.Errorf("encoding output: %w", err)
	}

	return nil
}

func parseInput(name string, args []string, stdin io.Reader, stdout, stderr io.Writer) (*hocon.Config, error) {
	// Check for help flag
	for _, a := range args {
		if a == "-h" || a == "--help" {
			printUsage(name, stdout)
			return nil, nil
		}
	}

	switch len(args) {
	case 0:
		data, err := io.ReadAll(stdin)
		if err != nil {
			return nil, fmt.Errorf("reading stdin: %w", err)
		}
		cfg, err := hocon.ParseString(string(data))
		if err != nil {
			return nil, fmt.Errorf("parsing HOCON: %w", err)
		}
		return cfg, nil

	case 1:
		cfg, err := hocon.ParseFile(args[0])
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", args[0], err)
		}
		return cfg, nil

	default:
		// Multiple files: parse each, merge with right-precedence (WithFallback)
		// hocon2json base.conf override.conf → override.WithFallback(base)
		configs := make([]*hocon.Config, len(args))
		for i, path := range args {
			cfg, err := hocon.ParseFile(path)
			if err != nil {
				return nil, fmt.Errorf("parsing %s: %w", path, err)
			}
			configs[i] = cfg
		}
		// Right-precedence: last file wins
		// Start from rightmost, fold left with WithFallback
		merged := configs[len(configs)-1]
		for i := len(configs) - 2; i >= 0; i-- {
			merged = merged.WithFallback(configs[i])
		}
		return merged, nil
	}
}

func printUsage(name string, w io.Writer) {
	fmt.Fprintf(w, "Usage: %s [FILE...]\n", name)
	fmt.Fprintln(w, "")
	fmt.Fprintf(w, "Convert HOCON to %s.\n", formatName(name))
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "If no FILE is given, reads from stdin.")
	fmt.Fprintln(w, "If multiple FILEs are given, they are merged (last file takes precedence).")
}

func formatName(name string) string {
	// "hocon2json" → "JSON", "hocon2yaml" → "YAML", etc.
	if len(name) > 6 {
		format := name[6:]
		switch format {
		case "json":
			return "JSON"
		case "yaml":
			return "YAML"
		case "toml":
			return "TOML"
		case "properties":
			return "Properties"
		}
	}
	return "the target format"
}
```

- [ ] **Step 4: Write JSONEncoder**

```go
// internal/convert/json.go
package convert

import (
	"encoding/json"
	"io"
)

// JSONEncoder encodes data as pretty-printed JSON.
type JSONEncoder struct{}

func (JSONEncoder) Encode(w io.Writer, data map[string]any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}
```

- [ ] **Step 5: Write golden test for JSON**

```go
// internal/convert/convert_test.go
package convert_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/o3co/hocon2/internal/convert"
)

func TestEncoders(t *testing.T) {
	encoders := map[string]convert.Encoder{
		"json": convert.JSONEncoder{},
	}
	testcases := []string{"basic"}

	for format, enc := range encoders {
		for _, tc := range testcases {
			t.Run(format+"/"+tc, func(t *testing.T) {
				inputPath := filepath.Join("..", "..", "testdata", tc+".hocon")

				expectedBytes, err := os.ReadFile(filepath.Join("..", "..", "testdata", tc+"."+format))
				if err != nil {
					t.Fatalf("reading expected output: %v", err)
				}
				expected := string(expectedBytes)

				var stdout bytes.Buffer
				var stderr bytes.Buffer
				err = convert.Run("hocon2"+format, enc, []string{inputPath}, strings.NewReader(""), &stdout, &stderr)
				if err != nil {
					t.Fatalf("Run() error: %v", err)
				}

				if stdout.String() != expected {
					t.Errorf("output mismatch:\ngot:\n%s\nwant:\n%s", stdout.String(), expected)
				}
			})
		}
	}
}
```

- [ ] **Step 6: Run test to verify it passes**

Run: `go test ./internal/convert/ -run TestEncoders/json/basic -v`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add internal/convert/ testdata/basic.hocon testdata/basic.json
git commit -m "feat: add Encoder interface, Run(), and JSONEncoder with golden test"
```

---

## Task 2: Refactor hocon2json to use convert.Run()

**Files:**
- Modify: `cmd/hocon2json/main.go`

- [ ] **Step 1: Rewrite cmd/hocon2json/main.go**

```go
// cmd/hocon2json/main.go
package main

import (
	"fmt"
	"os"

	"github.com/o3co/hocon2/internal/convert"
)

func main() {
	if err := convert.Run("hocon2json", convert.JSONEncoder{}, os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "hocon2json: %v\n", err)
		os.Exit(1)
	}
}
```

- [ ] **Step 2: Build and smoke test**

Run: `go build ./cmd/hocon2json/ && echo '{ name = "test" }' | ./hocon2json`
Expected: `{"name": "test"}` (pretty-printed)

- [ ] **Step 3: Commit**

```bash
git add cmd/hocon2json/main.go
git commit -m "refactor: hocon2json uses convert.Run()"
```

---

## Task 3: Additional Test Data (nested, array, substitution)

**Files:**
- Create: `testdata/nested.hocon`, `testdata/nested.json`
- Create: `testdata/array.hocon`, `testdata/array.json`
- Create: `testdata/substitution.hocon`, `testdata/substitution.json`
- Modify: `internal/convert/convert_test.go` — add test cases to list

- [ ] **Step 1: Create testdata/nested.hocon**

```hocon
database {
  host = "localhost"
  port = 5432
  connection {
    timeout = 30
    pool_size = 10
  }
}
```

- [ ] **Step 2: Create testdata/nested.json**

```json
{
  "database": {
    "connection": {
      "pool_size": 10,
      "timeout": 30
    },
    "host": "localhost",
    "port": 5432
  }
}
```

- [ ] **Step 3: Create testdata/array.hocon**

```hocon
tags = ["web", "api", "v2"]
ports = [8080, 8443]
```

- [ ] **Step 4: Create testdata/array.json**

```json
{
  "ports": [
    8080,
    8443
  ],
  "tags": [
    "web",
    "api",
    "v2"
  ]
}
```

- [ ] **Step 5: Create testdata/substitution.hocon**

```hocon
base_url = "https://api.example.com"
endpoints {
  users = ${base_url}"/users"
  posts = ${base_url}"/posts"
}
```

- [ ] **Step 6: Create testdata/substitution.json**

```json
{
  "base_url": "https://api.example.com",
  "endpoints": {
    "posts": "https://api.example.com/posts",
    "users": "https://api.example.com/users"
  }
}
```

- [ ] **Step 7: Update test to include new cases**

In `internal/convert/convert_test.go`, update:
```go
testcases := []string{"basic", "nested", "array", "substitution"}
```

- [ ] **Step 8: Run tests**

Run: `go test ./internal/convert/ -v`
Expected: PASS (4 test cases for json)

Note: The exact JSON output may need adjustment based on `go.hocon`'s actual Unmarshal behavior. Run the test, check the actual output, and fix the expected files if needed.

- [ ] **Step 9: Commit**

```bash
git add testdata/ internal/convert/convert_test.go
git commit -m "test: add nested, array, and substitution test data"
```

---

## Task 4: YAML Encoder (TDD)

**Files:**
- Create: `internal/convert/yaml.go`
- Create: `testdata/basic.yaml`, `testdata/nested.yaml`, `testdata/array.yaml`, `testdata/substitution.yaml`
- Modify: `internal/convert/convert_test.go` — add yaml encoder
- Modify: `go.mod` — add `gopkg.in/yaml.v3`

- [ ] **Step 1: Add yaml dependency**

Run: `go get gopkg.in/yaml.v3`

- [ ] **Step 2: Create expected YAML test data files**

`testdata/basic.yaml`:
```yaml
enabled: true
name: hocon2
version: 1
```

`testdata/nested.yaml`:
```yaml
database:
    connection:
        pool_size: 10
        timeout: 30
    host: localhost
    port: 5432
```

`testdata/array.yaml`:
```yaml
ports:
    - 8080
    - 8443
tags:
    - web
    - api
    - v2
```

`testdata/substitution.yaml`:
```yaml
base_url: https://api.example.com
endpoints:
    posts: https://api.example.com/posts
    users: https://api.example.com/users
```

Note: Exact indentation/formatting depends on `yaml.v3` defaults. Adjust after first test run.

- [ ] **Step 3: Add YAMLEncoder to test**

In `internal/convert/convert_test.go`, update encoders map:
```go
encoders := map[string]convert.Encoder{
    "json": convert.JSONEncoder{},
    "yaml": convert.YAMLEncoder{},
}
```

- [ ] **Step 4: Run test to verify it fails**

Run: `go test ./internal/convert/ -run TestEncoders/yaml -v`
Expected: FAIL (YAMLEncoder not defined)

- [ ] **Step 5: Write YAMLEncoder**

```go
// internal/convert/yaml.go
package convert

import (
	"io"

	"gopkg.in/yaml.v3"
)

// YAMLEncoder encodes data as YAML.
type YAMLEncoder struct{}

func (YAMLEncoder) Encode(w io.Writer, data map[string]any) error {
	enc := yaml.NewEncoder(w)
	defer enc.Close()
	return enc.Encode(data)
}
```

- [ ] **Step 6: Run test to verify it passes**

Run: `go test ./internal/convert/ -run TestEncoders/yaml -v`
Expected: PASS (adjust expected files if formatting differs)

- [ ] **Step 7: Commit**

```bash
git add internal/convert/yaml.go testdata/*.yaml internal/convert/convert_test.go go.mod go.sum
git commit -m "feat: add YAMLEncoder with golden tests"
```

---

## Task 5: TOML Encoder (TDD)

**Files:**
- Create: `internal/convert/toml.go`
- Create: `testdata/basic.toml`, `testdata/nested.toml`, `testdata/array.toml`, `testdata/substitution.toml`
- Modify: `internal/convert/convert_test.go` — add toml encoder
- Modify: `go.mod` — add `github.com/BurntSushi/toml`

- [ ] **Step 1: Add toml dependency**

Run: `go get github.com/BurntSushi/toml`

- [ ] **Step 2: Create expected TOML test data files**

`testdata/basic.toml`:
```toml
enabled = true
name = "hocon2"
version = 1
```

`testdata/nested.toml`:
```toml
[database]
  host = "localhost"
  port = 5432

  [database.connection]
    pool_size = 10
    timeout = 30
```

`testdata/array.toml`:
```toml
ports = [8080, 8443]
tags = ["web", "api", "v2"]
```

`testdata/substitution.toml`:
```toml
base_url = "https://api.example.com"

[endpoints]
  posts = "https://api.example.com/posts"
  users = "https://api.example.com/users"
```

Note: Exact formatting depends on `BurntSushi/toml` encoder output. Adjust after first test run.

- [ ] **Step 3: Add TOMLEncoder to test**

In `internal/convert/convert_test.go`, update encoders map:
```go
encoders := map[string]convert.Encoder{
    "json": convert.JSONEncoder{},
    "yaml": convert.YAMLEncoder{},
    "toml": convert.TOMLEncoder{},
}
```

- [ ] **Step 4: Run test to verify it fails**

Run: `go test ./internal/convert/ -run TestEncoders/toml -v`
Expected: FAIL (TOMLEncoder not defined)

- [ ] **Step 5: Write TOMLEncoder**

```go
// internal/convert/toml.go
package convert

import (
	"io"

	"github.com/BurntSushi/toml"
)

// TOMLEncoder encodes data as TOML.
type TOMLEncoder struct{}

func (TOMLEncoder) Encode(w io.Writer, data map[string]any) error {
	enc := toml.NewEncoder(w)
	return enc.Encode(data)
}
```

- [ ] **Step 6: Run test to verify it passes**

Run: `go test ./internal/convert/ -run TestEncoders/toml -v`
Expected: PASS (adjust expected files if formatting differs)

- [ ] **Step 7: Commit**

```bash
git add internal/convert/toml.go testdata/*.toml internal/convert/convert_test.go go.mod go.sum
git commit -m "feat: add TOMLEncoder with golden tests"
```

---

## Task 6: Flatten Package (TDD)

**Files:**
- Create: `internal/flatten/flatten.go`
- Create: `internal/flatten/flatten_test.go`

- [ ] **Step 1: Write flatten tests**

```go
// internal/flatten/flatten_test.go
package flatten_test

import (
	"testing"

	"github.com/o3co/hocon2/internal/flatten"
)

func TestFlatten(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected map[string]string
	}{
		{
			name:     "flat map",
			input:    map[string]any{"key": "value", "num": 42},
			expected: map[string]string{"key": "value", "num": "42"},
		},
		{
			name: "nested map",
			input: map[string]any{
				"db": map[string]any{
					"host": "localhost",
					"port": 5432,
				},
			},
			expected: map[string]string{
				"db.host": "localhost",
				"db.port": "5432",
			},
		},
		{
			name: "slice",
			input: map[string]any{
				"items": []any{1, 2, 3},
			},
			expected: map[string]string{
				"items.0": "1",
				"items.1": "2",
				"items.2": "3",
			},
		},
		{
			name:     "null value",
			input:    map[string]any{"key": nil},
			expected: map[string]string{"key": ""},
		},
		{
			name:     "empty map",
			input:    map[string]any{"obj": map[string]any{}},
			expected: map[string]string{},
		},
		{
			name:     "empty slice",
			input:    map[string]any{"arr": []any{}},
			expected: map[string]string{},
		},
		{
			name:     "bool value",
			input:    map[string]any{"flag": true},
			expected: map[string]string{"flag": "true"},
		},
		{
			name: "deeply nested",
			input: map[string]any{
				"a": map[string]any{
					"b": map[string]any{
						"c": "deep",
					},
				},
			},
			expected: map[string]string{"a.b.c": "deep"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flatten.Flatten(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("length mismatch: got %d, want %d\ngot:  %v\nwant: %v", len(result), len(tt.expected), result, tt.expected)
				return
			}
			for k, v := range tt.expected {
				if result[k] != v {
					t.Errorf("key %q: got %q, want %q", k, result[k], v)
				}
			}
		})
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/flatten/ -v`
Expected: FAIL (package doesn't exist)

- [ ] **Step 3: Implement Flatten**

```go
// internal/flatten/flatten.go
package flatten

import "fmt"

// Flatten converts a nested map[string]any to a flat map[string]string
// with dot-separated keys.
func Flatten(m map[string]any) map[string]string {
	result := make(map[string]string)
	flattenRecursive(m, "", result)
	return result
}

func flattenRecursive(m map[string]any, prefix string, result map[string]string) {
	for k, v := range m {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}

		switch val := v.(type) {
		case map[string]any:
			if len(val) == 0 {
				continue // skip empty maps
			}
			flattenRecursive(val, key, result)
		case []any:
			if len(val) == 0 {
				continue // skip empty slices
			}
			for i, item := range val {
				indexKey := fmt.Sprintf("%s.%d", key, i)
				switch nested := item.(type) {
				case map[string]any:
					flattenRecursive(nested, indexKey, result)
				default:
					result[indexKey] = fmt.Sprintf("%v", item)
				}
			}
		case nil:
			result[key] = ""
		default:
			result[key] = fmt.Sprintf("%v", val)
		}
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/flatten/ -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/flatten/
git commit -m "feat: add flatten package for properties encoding"
```

---

## Task 7: Properties Encoder (TDD)

**Files:**
- Create: `internal/convert/properties.go`
- Create: `testdata/basic.properties`, `testdata/nested.properties`, `testdata/array.properties`, `testdata/substitution.properties`
- Modify: `internal/convert/convert_test.go` — add properties encoder
- Modify: `go.mod` — add `github.com/magiconair/properties`

- [ ] **Step 1: Add properties dependency**

Run: `go get github.com/magiconair/properties`

- [ ] **Step 2: Create expected Properties test data files**

`testdata/basic.properties`:
```properties
enabled = true
name = hocon2
version = 1
```

`testdata/nested.properties`:
```properties
database.connection.pool_size = 10
database.connection.timeout = 30
database.host = localhost
database.port = 5432
```

`testdata/array.properties`:
```properties
ports.0 = 8080
ports.1 = 8443
tags.0 = web
tags.1 = api
tags.2 = v2
```

`testdata/substitution.properties`:
```properties
base_url = https\://api.example.com
endpoints.posts = https\://api.example.com/posts
endpoints.users = https\://api.example.com/users
```

Note: `magiconair/properties` may escape colons and other special chars. Exact output format needs to be adjusted after first test run.

- [ ] **Step 3: Add PropertiesEncoder to test**

In `internal/convert/convert_test.go`, update encoders map:
```go
encoders := map[string]convert.Encoder{
    "json":       convert.JSONEncoder{},
    "yaml":       convert.YAMLEncoder{},
    "toml":       convert.TOMLEncoder{},
    "properties": convert.PropertiesEncoder{},
}
```

- [ ] **Step 4: Run test to verify it fails**

Run: `go test ./internal/convert/ -run TestEncoders/properties -v`
Expected: FAIL (PropertiesEncoder not defined)

- [ ] **Step 5: Write PropertiesEncoder**

```go
// internal/convert/properties.go
package convert

import (
	"io"

	"github.com/magiconair/properties"
	"github.com/o3co/hocon2/internal/flatten"
)

// PropertiesEncoder encodes data as Java .properties format.
type PropertiesEncoder struct{}

func (PropertiesEncoder) Encode(w io.Writer, data map[string]any) error {
	flat := flatten.Flatten(data)
	p := properties.LoadMap(flat)
	_, err := p.Write(w, properties.UTF8)
	return err
}
```

- [ ] **Step 6: Run test to verify it passes**

Run: `go test ./internal/convert/ -run TestEncoders/properties -v`
Expected: PASS (adjust expected files based on actual output)

- [ ] **Step 7: Commit**

```bash
git add internal/convert/properties.go testdata/*.properties internal/convert/convert_test.go go.mod go.sum
git commit -m "feat: add PropertiesEncoder with golden tests"
```

---

## Task 8: New CLI Binaries (hocon2yaml, hocon2toml, hocon2properties)

**Files:**
- Create: `cmd/hocon2yaml/main.go`
- Create: `cmd/hocon2toml/main.go`
- Create: `cmd/hocon2properties/main.go`
- Modify: `.gitignore` — add binary names

- [ ] **Step 1: Create cmd/hocon2yaml/main.go**

```go
package main

import (
	"fmt"
	"os"

	"github.com/o3co/hocon2/internal/convert"
)

func main() {
	if err := convert.Run("hocon2yaml", convert.YAMLEncoder{}, os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "hocon2yaml: %v\n", err)
		os.Exit(1)
	}
}
```

- [ ] **Step 2: Create cmd/hocon2toml/main.go**

```go
package main

import (
	"fmt"
	"os"

	"github.com/o3co/hocon2/internal/convert"
)

func main() {
	if err := convert.Run("hocon2toml", convert.TOMLEncoder{}, os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "hocon2toml: %v\n", err)
		os.Exit(1)
	}
}
```

- [ ] **Step 3: Create cmd/hocon2properties/main.go**

```go
package main

import (
	"fmt"
	"os"

	"github.com/o3co/hocon2/internal/convert"
)

func main() {
	if err := convert.Run("hocon2properties", convert.PropertiesEncoder{}, os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "hocon2properties: %v\n", err)
		os.Exit(1)
	}
}
```

- [ ] **Step 4: Update .gitignore**

Add to `.gitignore`:
```
/hocon2yaml
/hocon2toml
/hocon2properties
```

- [ ] **Step 5: Build all and smoke test**

Run: `go build ./cmd/... && echo 'name = "test"' | ./hocon2yaml && echo 'name = "test"' | ./hocon2toml && echo 'name = "test"' | ./hocon2properties`
Expected: Each binary outputs the same data in its format

- [ ] **Step 6: Commit**

```bash
git add cmd/hocon2yaml/ cmd/hocon2toml/ cmd/hocon2properties/ .gitignore
git commit -m "feat: add hocon2yaml, hocon2toml, hocon2properties CLIs"
```

---

## Task 9: CLI Integration Tests

**Files:**
- Create: `internal/convert/integration_test.go`

- [ ] **Step 1: Write integration tests**

```go
// internal/convert/integration_test.go
package convert_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func projectRoot(t *testing.T) string {
	t.Helper()
	// internal/convert/ → project root
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolving project root: %v", err)
	}
	return root
}

func buildBinary(t *testing.T, name string) string {
	t.Helper()
	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, name)
	root := projectRoot(t)
	cmd := exec.Command("go", "build", "-o", binPath, "./cmd/"+name+"/")
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("building %s: %v\n%s", name, err, out)
	}
	return binPath
}

func TestCLI_Stdin(t *testing.T) {
	bin := buildBinary(t, "hocon2json")
	cmd := exec.Command(bin)
	cmd.Stdin = strings.NewReader(`name = "test"`)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(out), `"name"`) {
		t.Errorf("expected JSON output containing name, got: %s", out)
	}
}

func TestCLI_File(t *testing.T) {
	bin := buildBinary(t, "hocon2json")
	inputPath, _ := filepath.Abs(filepath.Join("..", "..", "testdata", "basic.hocon"))
	cmd := exec.Command(bin, inputPath)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(out), `"name"`) {
		t.Errorf("expected JSON output containing name, got: %s", out)
	}
}

func TestCLI_Help(t *testing.T) {
	bin := buildBinary(t, "hocon2json")
	cmd := exec.Command(bin, "--help")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(out), "Usage:") {
		t.Errorf("expected usage text, got: %s", out)
	}
}

func TestCLI_MultipleFiles(t *testing.T) {
	bin := buildBinary(t, "hocon2json")

	// Create temp files
	dir := t.TempDir()

	base := filepath.Join(dir, "base.conf")
	os.WriteFile(base, []byte(`name = "base", port = 8080`), 0644)

	override := filepath.Join(dir, "override.conf")
	os.WriteFile(override, []byte(`name = "override"`), 0644)

	cmd := exec.Command(bin, base, override)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("unexpected error: %v\nstderr: %s", err, stderr.String())
	}

	output := stdout.String()
	// override should win for "name"
	if !strings.Contains(output, `"override"`) {
		t.Errorf("expected override value, got: %s", output)
	}
	// base should provide "port"
	if !strings.Contains(output, `"port"`) {
		t.Errorf("expected port from base, got: %s", output)
	}
}
```

- [ ] **Step 2: Run integration tests**

Run: `go test ./internal/convert/ -run TestCLI -v`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add internal/convert/integration_test.go
git commit -m "test: add CLI integration tests (stdin, file, help, multi-file merge)"
```

---

## Task 10: Makefile

**Files:**
- Create: `Makefile`

- [ ] **Step 1: Create Makefile**

```makefile
.PHONY: build test lint all install clean

build:
	go build ./cmd/...

test:
	go test ./...

lint:
	golangci-lint run

all: lint test build

install:
	go install ./cmd/...

clean:
	rm -f hocon2json hocon2yaml hocon2toml hocon2properties
```

- [ ] **Step 2: Run make test**

Run: `make test`
Expected: All tests pass

- [ ] **Step 3: Run make build**

Run: `make build`
Expected: 4 binaries built without errors

- [ ] **Step 4: Commit**

```bash
git add Makefile
git commit -m "chore: add Makefile"
```

---

## Task 11: CI — GitHub Actions

**Files:**
- Create: `.github/workflows/ci.yml`

- [ ] **Step 1: Create CI workflow**

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [master, develop]
  pull_request:

jobs:
  test:
    strategy:
      matrix:
        go-version: ['1.25']
        os: [ubuntu-latest, macos-latest]
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go-version }}

      - name: Vet
        run: go vet ./...

      - name: Lint
        uses: golangci/golangci-lint-action@v6
        with:
          version: latest

      - name: Test
        run: go test ./...

      - name: Build
        run: go build ./cmd/...
```

- [ ] **Step 2: Commit**

```bash
git add .github/workflows/ci.yml
git commit -m "ci: add GitHub Actions workflow (vet, lint, test, build)"
```

---

## Task 12: GoReleaser + Release Workflow

**Files:**
- Create: `.goreleaser.yml`
- Create: `.github/workflows/release.yml`

- [ ] **Step 1: Create .goreleaser.yml**

```yaml
version: 2
builds:
  - id: hocon2json
    main: ./cmd/hocon2json
    binary: hocon2json
    goos: [linux, darwin]
    goarch: [amd64, arm64]

  - id: hocon2yaml
    main: ./cmd/hocon2yaml
    binary: hocon2yaml
    goos: [linux, darwin]
    goarch: [amd64, arm64]

  - id: hocon2toml
    main: ./cmd/hocon2toml
    binary: hocon2toml
    goos: [linux, darwin]
    goarch: [amd64, arm64]

  - id: hocon2properties
    main: ./cmd/hocon2properties
    binary: hocon2properties
    goos: [linux, darwin]
    goarch: [amd64, arm64]

archives:
  - format: tar.gz
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"

changelog:
  sort: asc
  filters:
    exclude:
      - "^docs:"
      - "^test:"
      - "^chore:"
```

- [ ] **Step 2: Create release workflow**

```yaml
# .github/workflows/release.yml
name: Release

on:
  push:
    tags:
      - 'v*'

permissions:
  contents: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v5
        with:
          go-version: '1.25'

      - uses: goreleaser/goreleaser-action@v6
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

- [ ] **Step 3: Commit**

```bash
git add .goreleaser.yml .github/workflows/release.yml
git commit -m "ci: add GoReleaser config and release workflow"
```

---

## Task 13: README.md

**Files:**
- Create: `README.md`

- [ ] **Step 1: Write README.md**

```markdown
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
go install github.com/o3co/hocon2/cmd/hocon2json@latest
go install github.com/o3co/hocon2/cmd/hocon2yaml@latest
go install github.com/o3co/hocon2/cmd/hocon2toml@latest
go install github.com/o3co/hocon2/cmd/hocon2properties@latest
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
```

- [ ] **Step 2: Commit**

```bash
git add README.md
git commit -m "docs: add README"
```

---

## Task 14: CONTRIBUTING.md

**Files:**
- Create: `CONTRIBUTING.md`

- [ ] **Step 1: Write CONTRIBUTING.md**

```markdown
# Contributing to hocon2

## Development Setup

```bash
git clone https://github.com/o3co/hocon2.git
cd hocon2
make all
```

### Requirements

- Go 1.25+
- [golangci-lint](https://golangci-lint.run/welcome/install/)

## Branch Strategy

- `master` — release branch (protected: PR + CI required)
- `develop` — default work branch

## Workflow

1. Create a feature branch from `develop`
2. Make changes with tests
3. Run `make all` to verify
4. Open a PR to `develop`

## Commit Style

Use [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` new feature
- `fix:` bug fix
- `test:` test changes
- `docs:` documentation
- `chore:` maintenance
- `refactor:` code restructuring

## Testing

```bash
make test           # run all tests
go test ./... -v    # verbose output
```

### Golden Tests

Test data lives in `testdata/`. Each `.hocon` file has corresponding output files (`.json`, `.yaml`, `.toml`, `.properties`). To add a test case, create the input and all expected outputs.
```

- [ ] **Step 2: Commit**

```bash
git add CONTRIBUTING.md
git commit -m "docs: add CONTRIBUTING guide"
```

---

## Task 15: Update CLAUDE.md

**Files:**
- Modify: `CLAUDE.md`

- [ ] **Step 1: Update CLAUDE.md to reflect new architecture**

Replace the Architecture and Design Decisions sections with content reflecting the current state. Key updates:

Architecture tree:
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
├── go.mod               # module github.com/o3co/hocon2
└── LICENSE              # Apache 2.0
```

Core Dependency section — add:
- `gopkg.in/yaml.v3` — YAML encoding
- `github.com/BurntSushi/toml` — TOML encoding
- `github.com/magiconair/properties` — Properties encoding

Design Decisions — replace with:
- **Encoder interface** — `internal/convert.Encoder` defines `Encode(w io.Writer, data map[string]any) error`。各フォーマットはこれを実装するだけ
- **共通 Run()** — 入力処理・パース・マージ・エンコードを一括で行う。各コマンドは `convert.Run(name, encoder, ...)` を呼ぶだけの薄いラッパー
- **複数ファイルマージ** — 位置引数で複数ファイル指定可能。右優先（後のファイルが上書き）で `WithFallback` マージ
- **internal パッケージ** — 外部に公開しない。API 安定化の責務を避ける

Building & Running section — add all 4 binaries and `make all` の記載

- [ ] **Step 2: Commit**

```bash
git add CLAUDE.md
git commit -m "docs: update CLAUDE.md for multi-format architecture"
```

---

## Task 16: Final Verification

- [ ] **Step 1: Run full test suite**

Run: `make all`
Expected: lint passes, all tests pass, all 4 binaries build

- [ ] **Step 2: Smoke test each binary**

```bash
echo 'name = "test", port = 8080' | ./hocon2json
echo 'name = "test", port = 8080' | ./hocon2yaml
echo 'name = "test", port = 8080' | ./hocon2toml
echo 'name = "test", port = 8080' | ./hocon2properties
```

- [ ] **Step 3: Push develop branch**

Run: `git push origin develop`
