package convert_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/o3co/hocon2/internal/convert"
)

func TestRun_Stdin(t *testing.T) {
	var stdout, stderr bytes.Buffer
	input := `name = "from_stdin"`
	err := convert.Run("hocon2json", &convert.JSONEncoder{}, []string{}, strings.NewReader(input), &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), `"from_stdin"`) {
		t.Errorf("expected stdin value in output, got: %s", stdout.String())
	}
}

func TestRun_Help(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := convert.Run("hocon2json", &convert.JSONEncoder{}, []string{"--help"}, strings.NewReader(""), &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), "Usage: hocon2json") {
		t.Errorf("expected usage text, got: %s", stdout.String())
	}
}

func TestRun_HelpShort(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := convert.Run("hocon2yaml", &convert.JSONEncoder{}, []string{"-h"}, strings.NewReader(""), &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), "Usage: hocon2yaml") {
		t.Errorf("expected usage text, got: %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "YAML") {
		t.Errorf("expected YAML in help text, got: %s", stdout.String())
	}
}

func TestRun_MultipleFiles(t *testing.T) {
	dir := t.TempDir()

	base := filepath.Join(dir, "base.conf")
	os.WriteFile(base, []byte("name = \"base\"\nport = 8080"), 0644)

	override := filepath.Join(dir, "override.conf")
	os.WriteFile(override, []byte("name = \"override\""), 0644)

	var stdout, stderr bytes.Buffer
	err := convert.Run("hocon2json", &convert.JSONEncoder{}, []string{base, override}, strings.NewReader(""), &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, `"override"`) {
		t.Errorf("expected override value, got: %s", output)
	}
	if !strings.Contains(output, `"port"`) {
		t.Errorf("expected port from base, got: %s", output)
	}
}

func TestRun_InvalidFile(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := convert.Run("hocon2json", &convert.JSONEncoder{}, []string{"/nonexistent/file.conf"}, strings.NewReader(""), &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestRun_InvalidHOCN(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := convert.Run("hocon2json", &convert.JSONEncoder{}, []string{}, strings.NewReader("{{{{invalid"), &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for invalid HOCON")
	}
}

func TestRun_InvalidFileInMulti(t *testing.T) {
	dir := t.TempDir()
	valid := filepath.Join(dir, "valid.conf")
	os.WriteFile(valid, []byte(`name = "ok"`), 0644)

	var stdout, stderr bytes.Buffer
	err := convert.Run("hocon2json", &convert.JSONEncoder{}, []string{valid, "/nonexistent/file.conf"}, strings.NewReader(""), &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for nonexistent file in multi-file merge")
	}
}

func TestRun_HelpFormats(t *testing.T) {
	names := []struct {
		cmd    string
		expect string
	}{
		{"hocon2json", "JSON"},
		{"hocon2yaml", "YAML"},
		{"hocon2toml", "TOML"},
		{"hocon2properties", "Properties"},
		{"unknown", "the target format"},
	}
	for _, tt := range names {
		t.Run(tt.cmd, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			convert.Run(tt.cmd, &convert.JSONEncoder{}, []string{"--help"}, strings.NewReader(""), &stdout, &stderr)
			if !strings.Contains(stdout.String(), tt.expect) {
				t.Errorf("expected %q in help output for %s, got: %s", tt.expect, tt.cmd, stdout.String())
			}
		})
	}
}

func TestRun_UnknownFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := convert.Run("hocon2json", &convert.JSONEncoder{}, []string{"-unknown"}, strings.NewReader(""), &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for unknown flag")
	}
	if !strings.Contains(stderr.String(), "flag") {
		t.Errorf("expected flag-related error on stderr, got stderr: %q, err: %v", stderr.String(), err)
	}
}

type failEncoder struct{}

func (failEncoder) Encode(w io.Writer, data map[string]any) error {
	return fmt.Errorf("encode failed")
}

func TestRun_EncodeError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := convert.Run("hocon2json", failEncoder{}, []string{}, strings.NewReader(`name = "test"`), &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error from failing encoder")
	}
	if !strings.Contains(err.Error(), "encoding output") {
		t.Errorf("expected 'encoding output' in error, got: %v", err)
	}
}
