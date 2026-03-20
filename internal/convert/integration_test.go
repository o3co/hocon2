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
	inputPath := filepath.Join(projectRoot(t), "testdata", "basic.hocon")
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
	dir := t.TempDir()

	base := filepath.Join(dir, "base.conf")
	os.WriteFile(base, []byte("name = \"base\"\nport = 8080"), 0644)

	override := filepath.Join(dir, "override.conf")
	os.WriteFile(override, []byte("name = \"override\""), 0644)

	cmd := exec.Command(bin, base, override)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("unexpected error: %v\nstderr: %s", err, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, `"override"`) {
		t.Errorf("expected override value, got: %s", output)
	}
	if !strings.Contains(output, `"port"`) {
		t.Errorf("expected port from base, got: %s", output)
	}
}
