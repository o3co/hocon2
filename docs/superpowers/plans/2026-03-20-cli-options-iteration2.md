# CLI Options & Project Polish Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add output formatting options, file output, CHANGELOG, and Windows support to hocon2 CLI tools.

**Architecture:** Introduce `flag.FlagSet` into `Run()` for flag parsing. Encoder-specific flags use optional `FlagRegistrar` interface. Output file handling (`-o`/`-overwrite`) is managed by `Run()`. Only `JSONEncoder` gets format options (`-compact`/`-indent`).

**Tech Stack:** Go standard library (`flag`, `errors`, `os`), existing dependencies unchanged.

**Spec:** `docs/superpowers/specs/2026-03-20-cli-options-iteration2-design.md`

---

## File Structure

| File | Action | Responsibility |
|------|--------|---------------|
| `internal/convert/convert.go` | Modify | `FlagRegistrar` interface, `Run()` with `flag.FlagSet`, `-o`/`-overwrite` logic, updated `printUsage` |
| `internal/convert/json.go` | Modify | `JSONEncoder` fields + `FlagRegistrar` + indent/compact logic |
| `cmd/hocon2json/main.go` | Modify | `&JSONEncoder{}` pointer |
| `internal/convert/run_test.go` | Modify | Tests for new flags |
| `internal/convert/convert_test.go` | Modify | `JSONEncoder{}` → `&JSONEncoder{}` pointer update |
| `internal/convert/conformance_test.go` | Modify | `JSONEncoder{}` → `&JSONEncoder{}` pointer update |
| `internal/convert/integration_test.go` | Modify | CLI E2E tests for flags |
| `.goreleaser.yml` | Modify | Add `windows`, `zip` format override |
| `CHANGELOG.md` | Create | Keep a Changelog format |

---

### Task 1: FlagRegistrar interface and flag.FlagSet in Run()

Replace manual `-h`/`--help` parsing with `flag.FlagSet`. Add `FlagRegistrar` optional interface.

**Files:**
- Modify: `internal/convert/convert.go`
- Test: `internal/convert/run_test.go`

- [ ] **Step 1: Write failing tests for flag parsing**

Add tests to `internal/convert/run_test.go`:

```go
func TestRun_UnknownFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := convert.Run("hocon2json", &convert.JSONEncoder{}, []string{"-unknown"}, strings.NewReader(""), &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for unknown flag")
	}
	// After flag.FlagSet is introduced, this should be a flag parse error, not a file-not-found error.
	if !strings.Contains(stderr.String(), "flag") {
		t.Errorf("expected flag-related error on stderr, got stderr: %q, err: %v", stderr.String(), err)
	}
}
```

Note: On current code, `-unknown` is treated as a file path argument, producing a file-not-found error. After introducing `flag.FlagSet`, it becomes a flag parse error with "flag provided but not defined" on stderr.

- [ ] **Step 2: Run tests to verify the new assertion fails**

Run: `cd /Volumes/Workspace/o3co/repos/go.hocon2 && go test ./internal/convert/ -run TestRun_UnknownFlag -v`
Expected: FAIL on the `stderr` assertion (current code doesn't write flag errors to stderr)

- [ ] **Step 3: Add FlagRegistrar interface and refactor Run()**

In `internal/convert/convert.go`:
- Add `"errors"`, `"flag"` to imports
- Add `FlagRegistrar` interface
- Replace `parseInput()` with flag-based parsing in `Run()`
- Update `printUsage` signature to `printUsage(fs *flag.FlagSet, name string, w io.Writer)`
- Handle `flag.ErrHelp` → return nil
- `fs.SetOutput(stderr)` for parse errors
- `fs.Usage` closure writes to `stdout` via `printUsage`

```go
// FlagRegistrar allows encoders to register format-specific flags.
type FlagRegistrar interface {
	RegisterFlags(fs *flag.FlagSet)
}

func Run(name string, enc Encoder, args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(stderr)

	if fr, ok := enc.(FlagRegistrar); ok {
		fr.RegisterFlags(fs)
	}

	fs.Usage = func() { printUsage(fs, name, stdout) }

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	cfg, err := parseInput(name, fs.Args(), stdin)
	if err != nil {
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
```

Simplify `parseInput` — remove help handling, remove `stdout`/`stderr` params:

```go
func parseInput(name string, args []string, stdin io.Reader) (*hocon.Config, error) {
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
		configs := make([]*hocon.Config, len(args))
		for i, path := range args {
			cfg, err := hocon.ParseFile(path)
			if err != nil {
				return nil, fmt.Errorf("parsing %s: %w", path, err)
			}
			configs[i] = cfg
		}
		merged := configs[len(configs)-1]
		for i := len(configs) - 2; i >= 0; i-- {
			merged = merged.WithFallback(configs[i])
		}
		return merged, nil
	}
}
```

Update `printUsage`:

```go
func printUsage(fs *flag.FlagSet, name string, w io.Writer) {
	fmt.Fprintf(w, "Usage: %s [OPTIONS] [FILE...]\n", name)
	fmt.Fprintln(w)
	fmt.Fprintf(w, "Convert HOCON to %s.\n", formatName(name))
	fmt.Fprintln(w)
	fmt.Fprintln(w, "If no FILE is given, reads from stdin.")
	fmt.Fprintln(w, "If multiple FILEs are given, they are merged (last file takes precedence).")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Options:")
	origOut := fs.Output()
	fs.SetOutput(w)
	fs.PrintDefaults()
	fs.SetOutput(origOut)
}
```

- [ ] **Step 4: Update existing tests for pointer receiver**

Change all `convert.JSONEncoder{}` to `&convert.JSONEncoder{}` in:
- `internal/convert/run_test.go` — all test functions
- `internal/convert/convert_test.go` — golden test encoder map (line 15)
- `internal/convert/conformance_test.go` — encoder maps (lines 26, 53)

- [ ] **Step 5: Run all tests**

Run: `cd /Volumes/Workspace/o3co/repos/go.hocon2 && go test ./internal/convert/ -v`
Expected: ALL PASS (including existing help, stdin, multi-file tests)

- [ ] **Step 6: Commit**

```bash
git add internal/convert/convert.go internal/convert/run_test.go internal/convert/convert_test.go internal/convert/conformance_test.go
git commit -m "feat: introduce flag.FlagSet and FlagRegistrar interface in Run()"
```

---

### Task 2: JSONEncoder output options (-compact, -indent)

Add `-compact` and `-indent` flags to JSONEncoder via FlagRegistrar.

**Files:**
- Modify: `internal/convert/json.go`
- Modify: `cmd/hocon2json/main.go`
- Test: `internal/convert/run_test.go`

- [ ] **Step 1: Write failing tests**

Add to `internal/convert/run_test.go`:

```go
func TestRun_JSONCompact(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := convert.Run("hocon2json", &convert.JSONEncoder{}, []string{"-compact"}, strings.NewReader(`name = "test"`), &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := stdout.String()
	if strings.Contains(output, "\n") {
		t.Errorf("expected compact output without newlines, got: %q", output)
	}
	if !strings.Contains(output, `"name":"test"`) {
		t.Errorf("expected compact JSON, got: %q", output)
	}
}

func TestRun_JSONIndent(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := convert.Run("hocon2json", &convert.JSONEncoder{}, []string{"-indent", "4"}, strings.NewReader(`name = "test"`), &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), "    \"name\"") {
		t.Errorf("expected 4-space indent, got: %q", stdout.String())
	}
}

func TestRun_JSONIndentInvalid(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"zero", []string{"-indent", "0"}},
		{"negative", []string{"-indent", "-1"}},
		{"too_large", []string{"-indent", "17"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			err := convert.Run("hocon2json", &convert.JSONEncoder{}, tt.args, strings.NewReader(`name = "test"`), &stdout, &stderr)
			if err == nil {
				t.Fatal("expected error for invalid indent value")
			}
		})
	}
}

func TestRun_JSONCompactOverridesIndent(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := convert.Run("hocon2json", &convert.JSONEncoder{}, []string{"-compact", "-indent", "4"}, strings.NewReader(`name = "test"`), &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := stdout.String()
	if strings.Contains(output, "\n") {
		t.Errorf("expected compact to override indent, got: %q", output)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Volumes/Workspace/o3co/repos/go.hocon2 && go test ./internal/convert/ -run "TestRun_JSON(Compact|Indent)" -v`
Expected: FAIL

- [ ] **Step 3: Implement JSONEncoder changes**

Update `internal/convert/json.go`:

```go
package convert

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"strings"
)

// JSONEncoder encodes data as JSON with configurable formatting.
type JSONEncoder struct {
	Compact bool
	Indent  int
}

func (e *JSONEncoder) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(&e.Compact, "compact", false, "output compact JSON")
	fs.IntVar(&e.Indent, "indent", 2, "indentation width")
}

func (e *JSONEncoder) Encode(w io.Writer, data map[string]any) error {
	if e.Compact {
		out, err := json.Marshal(data)
		if err != nil {
			return err
		}
		_, err = w.Write(out)
		return err
	}

	if e.Indent < 1 || e.Indent > 16 {
		return fmt.Errorf("invalid indent value %d: must be between 1 and 16", e.Indent)
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", strings.Repeat(" ", e.Indent))
	return enc.Encode(data)
}
```

- [ ] **Step 4: Update cmd/hocon2json/main.go**

Change `convert.JSONEncoder{}` to `&convert.JSONEncoder{}`:

```go
if err := convert.Run("hocon2json", &convert.JSONEncoder{}, os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
```

- [ ] **Step 5: Run all tests**

Run: `cd /Volumes/Workspace/o3co/repos/go.hocon2 && go test ./internal/convert/ -v`
Expected: ALL PASS

- [ ] **Step 6: Verify existing golden tests still pass**

Run: `cd /Volumes/Workspace/o3co/repos/go.hocon2 && go test ./internal/convert/ -run TestGolden -v`
Expected: ALL PASS (default indent=2 matches existing golden files)

- [ ] **Step 7: Commit**

```bash
git add internal/convert/json.go cmd/hocon2json/main.go internal/convert/run_test.go
git commit -m "feat: add -compact and -indent flags for JSON output"
```

---

### Task 3: Output file flag (-o, -overwrite)

Add `-o` and `-overwrite` flags to `Run()` for all commands.

**Files:**
- Modify: `internal/convert/convert.go`
- Test: `internal/convert/run_test.go`

- [ ] **Step 1: Write failing tests**

Add to `internal/convert/run_test.go`:

```go
func TestRun_OutputFile(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "output.json")

	var stdout, stderr bytes.Buffer
	err := convert.Run("hocon2json", &convert.JSONEncoder{}, []string{"-o", outPath}, strings.NewReader(`name = "test"`), &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("reading output file: %v", err)
	}
	if !strings.Contains(string(data), `"name"`) {
		t.Errorf("expected JSON in output file, got: %s", data)
	}
	if stdout.Len() != 0 {
		t.Errorf("expected no stdout output with -o, got: %s", stdout.String())
	}
}

func TestRun_OutputFileNoOverwrite(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "existing.json")
	os.WriteFile(outPath, []byte("old content"), 0644)

	var stdout, stderr bytes.Buffer
	err := convert.Run("hocon2json", &convert.JSONEncoder{}, []string{"-o", outPath}, strings.NewReader(`name = "test"`), &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error when output file exists without -overwrite")
	}
}

func TestRun_OutputFileOverwrite(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "existing.json")
	os.WriteFile(outPath, []byte("old content"), 0644)

	var stdout, stderr bytes.Buffer
	err := convert.Run("hocon2json", &convert.JSONEncoder{}, []string{"-o", outPath, "-overwrite"}, strings.NewReader(`name = "test"`), &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(outPath)
	if strings.Contains(string(data), "old content") {
		t.Error("expected file to be overwritten")
	}
}

func TestRun_OutputFileNoDir(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := convert.Run("hocon2json", &convert.JSONEncoder{}, []string{"-o", "/nonexistent/dir/out.json"}, strings.NewReader(`name = "test"`), &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for nonexistent directory")
	}
}

func TestRun_OverwriteWithoutO(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := convert.Run("hocon2json", &convert.JSONEncoder{}, []string{"-overwrite"}, strings.NewReader(`name = "test"`), &stdout, &stderr)
	if err != nil {
		t.Fatalf("expected -overwrite without -o to succeed silently, got: %v", err)
	}
	if !strings.Contains(stdout.String(), `"name"`) {
		t.Errorf("expected normal JSON output, got: %s", stdout.String())
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Volumes/Workspace/o3co/repos/go.hocon2 && go test ./internal/convert/ -run "TestRun_OutputFile" -v`
Expected: FAIL

- [ ] **Step 3: Implement -o and -overwrite in Run()**

Update `Run()` in `internal/convert/convert.go` to add flag variables and output file logic:

```go
func Run(name string, enc Encoder, args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(stderr)

	var outFile string
	var overwrite bool
	fs.StringVar(&outFile, "o", "", "output file path")
	fs.BoolVar(&overwrite, "overwrite", false, "overwrite existing output file")

	if fr, ok := enc.(FlagRegistrar); ok {
		fr.RegisterFlags(fs)
	}

	fs.Usage = func() { printUsage(fs, name, stdout) }

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	cfg, err := parseInput(name, fs.Args(), stdin)
	if err != nil {
		return err
	}

	var m map[string]any
	if err := cfg.Unmarshal(&m); err != nil {
		return fmt.Errorf("unmarshaling config: %w", err)
	}

	w, closer, err := openOutput(outFile, overwrite, stdout)
	if err != nil {
		return err
	}
	if closer != nil {
		defer closer.Close()
	}

	if err := enc.Encode(w, m); err != nil {
		return fmt.Errorf("encoding output: %w", err)
	}

	return nil
}
```

Add `openOutput` helper (also add `"os"` to imports):

```go
func openOutput(path string, overwrite bool, stdout io.Writer) (io.Writer, io.Closer, error) {
	if path == "" {
		return stdout, nil, nil
	}

	if !overwrite {
		if _, err := os.Stat(path); err == nil {
			return nil, nil, fmt.Errorf("output file %s already exists (use -overwrite to replace)", path)
		}
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return nil, nil, fmt.Errorf("opening output file: %w", err)
	}
	return f, f, nil
}
```

- [ ] **Step 4: Run all tests**

Run: `cd /Volumes/Workspace/o3co/repos/go.hocon2 && go test ./internal/convert/ -v`
Expected: ALL PASS

- [ ] **Step 5: Commit**

```bash
git add internal/convert/convert.go internal/convert/run_test.go
git commit -m "feat: add -o and -overwrite flags for file output"
```

---

### Task 4: Update help text verification tests

Verify help output includes new options and updated format.

**Files:**
- Modify: `internal/convert/run_test.go`

- [ ] **Step 1: Update TestRun_Help to check for OPTIONS and flag names**

```go
func TestRun_Help(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := convert.Run("hocon2json", &convert.JSONEncoder{}, []string{"--help"}, strings.NewReader(""), &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := stdout.String()
	for _, want := range []string{"Usage: hocon2json", "[OPTIONS]", "-compact", "-indent", "-o", "-overwrite"} {
		if !strings.Contains(output, want) {
			t.Errorf("expected %q in help output, got: %s", want, output)
		}
	}
}
```

- [ ] **Step 2: Add test that non-JSON commands don't show -compact/-indent**

```go
func TestRun_HelpNoFormatFlags(t *testing.T) {
	var stdout, stderr bytes.Buffer
	convert.Run("hocon2yaml", convert.YAMLEncoder{}, []string{"--help"}, strings.NewReader(""), &stdout, &stderr)
	output := stdout.String()
	if strings.Contains(output, "-compact") {
		t.Error("YAML help should not show -compact flag")
	}
	if strings.Contains(output, "-indent") {
		t.Error("YAML help should not show -indent flag")
	}
	if !strings.Contains(output, "-o") {
		t.Error("YAML help should show -o flag")
	}
}
```

- [ ] **Step 3: Run tests**

Run: `cd /Volumes/Workspace/o3co/repos/go.hocon2 && go test ./internal/convert/ -run "TestRun_Help" -v`
Expected: ALL PASS

- [ ] **Step 4: Commit**

```bash
git add internal/convert/run_test.go
git commit -m "test: update help text tests for new flags"
```

---

### Task 5: CLI integration tests for flags

Add E2E tests that build the binary and test flag behavior.

**Files:**
- Modify: `internal/convert/integration_test.go`

- [ ] **Step 1: Add integration tests**

```go
func TestCLI_Compact(t *testing.T) {
	bin := buildBinary(t, "hocon2json")
	cmd := exec.Command(bin, "-compact")
	cmd.Stdin = strings.NewReader(`name = "test"`)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(string(out), "\n") {
		t.Errorf("expected compact output, got: %q", out)
	}
}

func TestCLI_OutputFile(t *testing.T) {
	bin := buildBinary(t, "hocon2json")
	dir := t.TempDir()
	outPath := filepath.Join(dir, "out.json")

	cmd := exec.Command(bin, "-o", outPath)
	cmd.Stdin = strings.NewReader(`name = "test"`)
	if err := cmd.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("reading output: %v", err)
	}
	if !strings.Contains(string(data), `"name"`) {
		t.Errorf("expected JSON in file, got: %s", data)
	}
}

func TestCLI_OutputFileNoOverwrite(t *testing.T) {
	bin := buildBinary(t, "hocon2json")
	dir := t.TempDir()
	outPath := filepath.Join(dir, "existing.json")
	os.WriteFile(outPath, []byte("old"), 0644)

	cmd := exec.Command(bin, "-o", outPath)
	cmd.Stdin = strings.NewReader(`name = "test"`)
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected error when file exists without -overwrite")
	}
}
```

- [ ] **Step 2: Run integration tests**

Run: `cd /Volumes/Workspace/o3co/repos/go.hocon2 && go test ./internal/convert/ -run "TestCLI_" -v`
Expected: ALL PASS

- [ ] **Step 3: Commit**

```bash
git add internal/convert/integration_test.go
git commit -m "test: add CLI integration tests for -compact, -o, -overwrite"
```

---

### Task 6: Error message check (go.hocon)

Check if go.hocon parse errors include line numbers. No code changes to hocon2.

**Files:**
- None modified (investigation only)

- [ ] **Step 1: Test go.hocon error output**

Run: `cd /Volumes/Workspace/o3co/repos/go.hocon2 && go test -run "TestRun_InvalidHOCN" -v ./internal/convert/ 2>&1`

Also write a quick throwaway test or use `go run` to check the actual error message:

```bash
cd /Volumes/Workspace/o3co/repos/go.hocon2 && echo '{{{{invalid' | go run ./cmd/hocon2json/ 2>&1
```

Examine the error output. If it includes line/column info → document as "sufficient." If not → create a GitHub Issue on `o3co/go.hocon`.

- [ ] **Step 2: Commit a note (if Issue created)**

If an issue was created, note it in the CHANGELOG under a comment. No code changes.

---

### Task 7: CHANGELOG

Create `CHANGELOG.md` in Keep a Changelog format.

**Files:**
- Create: `CHANGELOG.md`

- [ ] **Step 1: Create CHANGELOG.md**

```markdown
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/).

## [Unreleased]

### Added

- `-compact` and `-indent` options for JSON output formatting
- `-o` output file option with `-overwrite` safety flag
- Windows binary releases
- This CHANGELOG

## [0.2.0] - 2026-03-20

### Changed

- Module path changed from `go.hocon2` to `hocon2`

### Added

- Japanese README

## [0.1.0] - 2026-03-20

### Added

- HOCON to JSON, YAML, TOML, and Properties conversion
- Multi-file merge with right-precedence
- Stdin and file input support
- Lightbend conformance tests (equiv01–equiv05)
- GoReleaser configuration for Linux and macOS
```

- [ ] **Step 2: Commit**

```bash
git add CHANGELOG.md
git commit -m "docs: add CHANGELOG in Keep a Changelog format"
```

---

### Task 8: Windows support (GoReleaser)

Add `windows` to GoReleaser build targets with `zip` format.

**Files:**
- Modify: `.goreleaser.yml`

- [ ] **Step 1: Update .goreleaser.yml**

Add `windows` to each build's `goos` and add `format_overrides` for zip:

```yaml
version: 2
builds:
  - id: hocon2json
    main: ./cmd/hocon2json
    binary: hocon2json
    goos: [linux, darwin, windows]
    goarch: [amd64, arm64]

  - id: hocon2yaml
    main: ./cmd/hocon2yaml
    binary: hocon2yaml
    goos: [linux, darwin, windows]
    goarch: [amd64, arm64]

  - id: hocon2toml
    main: ./cmd/hocon2toml
    binary: hocon2toml
    goos: [linux, darwin, windows]
    goarch: [amd64, arm64]

  - id: hocon2properties
    main: ./cmd/hocon2properties
    binary: hocon2properties
    goos: [linux, darwin, windows]
    goarch: [amd64, arm64]

archives:
  - format: tar.gz
    format_overrides:
      - goos: windows
        format: zip
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"

changelog:
  sort: asc
  filters:
    exclude:
      - "^docs:"
      - "^test:"
      - "^chore:"
```

- [ ] **Step 2: Verify GoReleaser config**

Run: `cd /Volumes/Workspace/o3co/repos/go.hocon2 && goreleaser check` (if goreleaser is installed, otherwise skip)

- [ ] **Step 3: Cross-compile verify**

Run: `cd /Volumes/Workspace/o3co/repos/go.hocon2 && GOOS=windows GOARCH=amd64 go build ./cmd/hocon2json/`
Expected: Builds successfully (produces `hocon2json.exe` or `hocon2json` depending on output)

- [ ] **Step 4: Clean up and commit**

```bash
rm -f hocon2json hocon2json.exe
git add .goreleaser.yml
git commit -m "feat: add Windows support to GoReleaser builds"
```

---

### Task 9: Final verification

Run full test suite and build.

- [ ] **Step 1: Run full test suite**

Run: `cd /Volumes/Workspace/o3co/repos/go.hocon2 && make test`
Expected: ALL PASS

- [ ] **Step 2: Build all binaries**

Run: `cd /Volumes/Workspace/o3co/repos/go.hocon2 && make build`
Expected: All 4 binaries build successfully

- [ ] **Step 3: Manual smoke test**

```bash
cd /Volumes/Workspace/o3co/repos/go.hocon2
echo 'name = "test"' | ./hocon2json -compact
echo 'name = "test"' | ./hocon2json -indent 4
echo 'name = "test"' | ./hocon2json -o /tmp/test.json && cat /tmp/test.json
./hocon2json --help
./hocon2yaml --help
```

- [ ] **Step 4: Clean up build artifacts**

Run: `cd /Volumes/Workspace/o3co/repos/go.hocon2 && make clean`
