package convert_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/o3co/go.hocon"
	"github.com/o3co/hocon2/internal/convert"
)

// runReverse drives one reverse conversion and returns stdout.
func runReverse(t *testing.T, name string, dec convert.Decoder, args []string, stdin string) (string, error) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	err := convert.RunReverse(name, dec, args, strings.NewReader(stdin), &stdout, &stderr)
	return stdout.String(), err
}

// assertTree parses emitted HOCON and asserts it unmarshals to want. This
// checks the conversion by value tree, not by exact text.
func assertTree(t *testing.T, hoconText string, want map[string]any) {
	t.Helper()
	cfg, err := hocon.ParseString(hoconText)
	if err != nil {
		t.Fatalf("emitted HOCON did not re-parse: %v\n--- emitted ---\n%s", err, hoconText)
	}
	var got map[string]any
	if err := cfg.Unmarshal(&got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !treeEqual(got, want) {
		t.Errorf("tree mismatch\n  got:  %#v\n  want: %#v\n--- emitted ---\n%s", got, want, hoconText)
	}
}

func TestReverse_YAML(t *testing.T) {
	out, err := runReverse(t, "yaml2hocon", convert.YAMLDecoder{}, nil,
		"name: svc\nport: 8080\nnorway: no\ndb:\n  host: localhost\n  replicas:\n    - id: 1\n    - id: 2\n")
	if err != nil {
		t.Fatalf("RunReverse: %v", err)
	}
	assertTree(t, out, map[string]any{
		"name": "svc", "port": int64(8080), "norway": "no",
		"db": map[string]any{"host": "localhost",
			"replicas": []any{map[string]any{"id": int64(1)}, map[string]any{"id": int64(2)}}},
	})
}

func TestReverse_TOML(t *testing.T) {
	out, err := runReverse(t, "toml2hocon", convert.TOMLDecoder{}, nil,
		"name = \"svc\"\nport = 8080\nwhen = 1979-05-27T07:32:00Z\n[db]\nhost = \"localhost\"\n")
	if err != nil {
		t.Fatalf("RunReverse: %v", err)
	}
	// F4.2: the datetime becomes a string and survives the round trip.
	assertTree(t, out, map[string]any{
		"name": "svc", "port": int64(8080), "when": "1979-05-27T07:32:00Z",
		"db": map[string]any{"host": "localhost"},
	})
}

func TestReverse_JSON(t *testing.T) {
	out, err := runReverse(t, "json2hocon", convert.JSONDecoder{}, nil,
		`{"a": 1, "b": [2, 3], "c": {"d": "x"}, "s": "8080"}`)
	if err != nil {
		t.Fatalf("RunReverse: %v", err)
	}
	// "8080" is a JSON string and must stay a string, not become a number.
	assertTree(t, out, map[string]any{
		"a": int64(1), "b": []any{int64(2), int64(3)},
		"c": map[string]any{"d": "x"}, "s": "8080",
	})
}

func TestReverse_Properties(t *testing.T) {
	out, err := runReverse(t, "properties2hocon", convert.PropertiesDecoder{}, nil,
		"db.host = localhost\ndb.port = 5432\napp.name = svc\n")
	if err != nil {
		t.Fatalf("RunReverse: %v", err)
	}
	// Properties values are all strings (S23.3), so 5432 stays a string.
	assertTree(t, out, map[string]any{
		"db":  map[string]any{"host": "localhost", "port": "5432"},
		"app": map[string]any{"name": "svc"},
	})
}

func TestReverse_ValidateOnly(t *testing.T) {
	out, err := runReverse(t, "yaml2hocon", convert.YAMLDecoder{}, []string{"-validate"}, "a: 1\n")
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if out != "" {
		t.Errorf("-validate produced output: %q", out)
	}
	// Invalid input is reported even under -validate.
	if _, err := runReverse(t, "yaml2hocon", convert.YAMLDecoder{}, []string{"-validate"}, "a: 1\n---\nb: 2\n"); err == nil {
		t.Error("multi-document YAML accepted under -validate, want error")
	}
}

func TestReverse_OutputFile(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.conf")
	if _, err := runReverse(t, "json2hocon", convert.JSONDecoder{}, []string{"-o", out}, `{"a":1}`); err != nil {
		t.Fatalf("RunReverse: %v", err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	assertTree(t, string(data), map[string]any{"a": int64(1)})

	// Refuses to clobber without -overwrite.
	if _, err := runReverse(t, "json2hocon", convert.JSONDecoder{}, []string{"-o", out}, `{"b":2}`); err == nil {
		t.Error("overwrote existing file without -overwrite")
	}
}

func TestReverse_InputErrors(t *testing.T) {
	if _, err := runReverse(t, "yaml2hocon", convert.YAMLDecoder{}, nil, "a: [unclosed\n"); err == nil {
		t.Error("malformed YAML accepted, want error")
	}
	if _, err := runReverse(t, "toml2hocon", convert.TOMLDecoder{}, []string{"a.toml", "b.toml"}, ""); err == nil {
		t.Error("two input files accepted, want error")
	}
}

// treeEqual compares two decoded trees, treating any integer kind as equal by
// value so int/int64 from different paths match.
func treeEqual(a, b any) bool {
	switch av := a.(type) {
	case map[string]any:
		bv, ok := b.(map[string]any)
		if !ok || len(av) != len(bv) {
			return false
		}
		for k, x := range av {
			if !treeEqual(x, bv[k]) {
				return false
			}
		}
		return true
	case []any:
		bv, ok := b.([]any)
		if !ok || len(av) != len(bv) {
			return false
		}
		for i := range av {
			if !treeEqual(av[i], bv[i]) {
				return false
			}
		}
		return true
	default:
		return normNum(a) == normNum(b)
	}
}

func normNum(v any) any {
	switch n := v.(type) {
	case int:
		return int64(n)
	case int32:
		return int64(n)
	}
	return v
}
