package convert_test

import (
	"bytes"
	"errors"
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
	output := stdout.String()
	for _, want := range []string{"Usage: hocon2json", "[OPTIONS]", "-compact", "-indent", "-o", "-overwrite"} {
		if !strings.Contains(output, want) {
			t.Errorf("expected %q in help output, got: %s", want, output)
		}
	}
}

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

func TestRun_JSONCompact(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := convert.Run("hocon2json", &convert.JSONEncoder{}, []string{"-compact"}, strings.NewReader(`name = "test"`), &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := strings.TrimSuffix(stdout.String(), "\n")
	if strings.Contains(output, "\n") {
		t.Errorf("expected single-line compact output, got: %q", output)
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
	output := strings.TrimSuffix(stdout.String(), "\n")
	if strings.Contains(output, "\n") {
		t.Errorf("expected single-line compact output, got: %q", output)
	}
}

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

type failingEncoder struct{}

func (failingEncoder) Encode(w io.Writer, data map[string]any) error {
	io.WriteString(w, "partial garbage")
	return errors.New("encode failed")
}

func TestRun_OutputFilePreservedOnEncodeError(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "existing.json")
	os.WriteFile(outPath, []byte("old content"), 0644)

	var stdout, stderr bytes.Buffer
	err := convert.Run("hocon2json", failingEncoder{}, []string{"-o", outPath, "-overwrite"}, strings.NewReader(`name = "test"`), &stdout, &stderr)
	if err == nil {
		t.Fatal("expected encode error")
	}
	data, _ := os.ReadFile(outPath)
	if string(data) != "old content" {
		t.Errorf("existing output file was corrupted, got: %s", data)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Errorf("expected temp file to be cleaned up, dir has: %v", entries)
	}
}

func TestRun_NoOutputFileOnEncodeError(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "new.json")

	var stdout, stderr bytes.Buffer
	err := convert.Run("hocon2json", failingEncoder{}, []string{"-o", outPath}, strings.NewReader(`name = "test"`), &stdout, &stderr)
	if err == nil {
		t.Fatal("expected encode error")
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 0 {
		t.Errorf("expected no files after failed encode, dir has: %v", entries)
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

func TestRun_JSONIndentBoundary(t *testing.T) {
	tests := []struct {
		name   string
		indent int
		prefix string
	}{
		{"min", 1, " \"name\""},
		{"max", 16, "                \"name\""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			err := convert.Run("hocon2json", &convert.JSONEncoder{}, []string{"-indent", fmt.Sprintf("%d", tt.indent)}, strings.NewReader(`name = "test"`), &stdout, &stderr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.Contains(stdout.String(), tt.prefix) {
				t.Errorf("expected %d-space indent, got: %q", tt.indent, stdout.String())
			}
		})
	}
}

func TestRun_OutputFileCompact(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "compact.json")

	var stdout, stderr bytes.Buffer
	err := convert.Run("hocon2json", &convert.JSONEncoder{}, []string{"-compact", "-o", outPath}, strings.NewReader(`name = "test"`), &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(outPath)
	output := strings.TrimSuffix(string(data), "\n")
	if strings.Contains(output, "\n") {
		t.Errorf("expected compact JSON in file, got: %q", data)
	}
}

func TestRun_OutputFileYAML(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "output.yaml")

	var stdout, stderr bytes.Buffer
	err := convert.Run("hocon2yaml", &convert.YAMLEncoder{}, []string{"-o", outPath}, strings.NewReader(`name = "test"`), &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(outPath)
	if !strings.Contains(string(data), "name:") {
		t.Errorf("expected YAML in output file, got: %s", data)
	}
	if stdout.Len() != 0 {
		t.Errorf("expected no stdout output with -o, got: %s", stdout.String())
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

func TestRun_Validate(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := convert.Run("hocon2json", &convert.JSONEncoder{}, []string{"-validate"}, strings.NewReader(`name = "test"`), &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stdout.Len() != 0 {
		t.Errorf("expected no output in validate mode, got: %s", stdout.String())
	}
}

func TestRun_ValidateInvalid(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := convert.Run("hocon2json", &convert.JSONEncoder{}, []string{"-validate"}, strings.NewReader("{{{{invalid"), &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for invalid HOCON in validate mode")
	}
}

func TestRun_ValidateFile(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "valid.conf")
	os.WriteFile(f, []byte(`name = "test"`), 0644)

	var stdout, stderr bytes.Buffer
	err := convert.Run("hocon2json", &convert.JSONEncoder{}, []string{"-validate", f}, strings.NewReader(""), &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stdout.Len() != 0 {
		t.Errorf("expected no output in validate mode, got: %s", stdout.String())
	}
}

func TestRun_ValidateIgnoresOutputFlag(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "output.json")

	var stdout, stderr bytes.Buffer
	err := convert.Run("hocon2json", &convert.JSONEncoder{}, []string{"-validate", "-o", outPath}, strings.NewReader(`name = "test"`), &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Output file should NOT be created in validate mode
	if _, err := os.Stat(outPath); err == nil {
		t.Error("expected no output file in validate mode")
	}
}

func TestRun_ValidateShowsInHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	convert.Run("hocon2json", &convert.JSONEncoder{}, []string{"--help"}, strings.NewReader(""), &stdout, &stderr)
	if !strings.Contains(stdout.String(), "-validate") {
		t.Errorf("expected -validate in help output, got: %s", stdout.String())
	}
}

func TestRun_EnvFile(t *testing.T) {
	for _, k := range []string{"MY_HOST", "MY_PORT"} {
		orig, existed := os.LookupEnv(k)
		t.Cleanup(func() {
			if existed {
				os.Setenv(k, orig)
			} else {
				os.Unsetenv(k)
			}
		})
		os.Unsetenv(k) // ensure loadEnvFile can set these (it won't overwrite existing vars)
	}

	dir := t.TempDir()

	envFile := filepath.Join(dir, ".env")
	os.WriteFile(envFile, []byte("MY_HOST=example.com\nMY_PORT=3000\n"), 0644)

	confFile := filepath.Join(dir, "app.conf")
	os.WriteFile(confFile, []byte("host = ${MY_HOST}\nport = ${MY_PORT}"), 0644)

	var stdout, stderr bytes.Buffer
	err := convert.Run("hocon2json", &convert.JSONEncoder{}, []string{"-env-file", envFile, confFile}, strings.NewReader(""), &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := stdout.String()
	if !strings.Contains(output, `"example.com"`) {
		t.Errorf("expected MY_HOST value in output, got: %s", output)
	}
	if !strings.Contains(output, `3000`) {
		t.Errorf("expected MY_PORT value in output, got: %s", output)
	}
}

func TestRun_EnvFileMissing(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := convert.Run("hocon2json", &convert.JSONEncoder{}, []string{"-env-file", "/nonexistent/.env"}, strings.NewReader(`name = "test"`), &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for nonexistent env file")
	}
}

func TestRun_EnvFileComments(t *testing.T) {
	for _, k := range []string{"HOCON2_TEST_KEY"} {
		orig, existed := os.LookupEnv(k)
		t.Cleanup(func() {
			if existed {
				os.Setenv(k, orig)
			} else {
				os.Unsetenv(k)
			}
		})
		os.Unsetenv(k)
	}

	dir := t.TempDir()

	envFile := filepath.Join(dir, ".env")
	os.WriteFile(envFile, []byte("# comment\nHOCON2_TEST_KEY=value\n\n# another comment\n"), 0644)

	var stdout, stderr bytes.Buffer
	err := convert.Run("hocon2json", &convert.JSONEncoder{}, []string{"-env-file", envFile}, strings.NewReader("k = ${HOCON2_TEST_KEY}"), &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), `"value"`) {
		t.Errorf("expected HOCON2_TEST_KEY value in output, got: %s", stdout.String())
	}
}

func TestRun_EnvFileQuotedValues(t *testing.T) {
	for _, k := range []string{"HOCON2_TEST_KEY"} {
		orig, existed := os.LookupEnv(k)
		t.Cleanup(func() {
			if existed {
				os.Setenv(k, orig)
			} else {
				os.Unsetenv(k)
			}
		})
		os.Unsetenv(k)
	}

	dir := t.TempDir()

	envFile := filepath.Join(dir, ".env")
	os.WriteFile(envFile, []byte(`HOCON2_TEST_KEY="quoted value"`+"\n"), 0644)

	var stdout, stderr bytes.Buffer
	err := convert.Run("hocon2json", &convert.JSONEncoder{}, []string{"-env-file", envFile}, strings.NewReader("k = ${HOCON2_TEST_KEY}"), &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), `"quoted value"`) {
		t.Errorf("expected quoted value in output, got: %s", stdout.String())
	}
}

func TestRun_EnvFileExportPrefix(t *testing.T) {
	for _, k := range []string{"HOCON2_TEST_KEY"} {
		orig, existed := os.LookupEnv(k)
		t.Cleanup(func() {
			if existed {
				os.Setenv(k, orig)
			} else {
				os.Unsetenv(k)
			}
		})
		os.Unsetenv(k)
	}

	dir := t.TempDir()

	envFile := filepath.Join(dir, ".env")
	os.WriteFile(envFile, []byte("export HOCON2_TEST_KEY=exported\n"), 0644)

	var stdout, stderr bytes.Buffer
	err := convert.Run("hocon2json", &convert.JSONEncoder{}, []string{"-env-file", envFile}, strings.NewReader("k = ${HOCON2_TEST_KEY}"), &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), `"exported"`) {
		t.Errorf("expected exported value in output, got: %s", stdout.String())
	}
	if _, exists := os.LookupEnv("export HOCON2_TEST_KEY"); exists {
		t.Error("malformed 'export HOCON2_TEST_KEY' variable was set")
	}
}

func TestRun_EnvFileShowsInHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	convert.Run("hocon2json", &convert.JSONEncoder{}, []string{"--help"}, strings.NewReader(""), &stdout, &stderr)
	if !strings.Contains(stdout.String(), "-env-file") {
		t.Errorf("expected -env-file in help output, got: %s", stdout.String())
	}
}
