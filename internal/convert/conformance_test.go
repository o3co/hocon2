package convert_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/o3co/go.hocon2/internal/convert"
)

// skipConformance lists tests to skip with reasons.
var skipConformance = map[string]string{
	// equiv03: go.hocon v0.2.0 does not support extensionless include probing
	// (all equiv03 tests are excluded by not listing equiv03 in equivDirs)

	// equiv05 conformance: triple-quoted string whitespace handling differs
	// between go.hocon and Lightbend reference implementation
	"equiv05/triple-quotes.conf/conformance": "go.hocon triple-quote whitespace handling differs from Lightbend reference",
}

func TestLightbendConformance(t *testing.T) {
	equivDirs := []string{"equiv01", "equiv02", "equiv04", "equiv05"}
	formats := []struct {
		name    string
		encoder convert.Encoder
	}{
		{"json", convert.JSONEncoder{}},
		{"yaml", convert.YAMLEncoder{}},
		{"toml", convert.TOMLEncoder{}},
		{"properties", convert.PropertiesEncoder{}},
	}

	lightbendDir := filepath.Join("..", "..", "testdata", "lightbend")

	for _, dir := range equivDirs {
		dirPath := filepath.Join(lightbendDir, dir)
		confFiles := findConfFiles(t, dirPath)

		// Load original.json for Phase 1
		originalPath := filepath.Join(dirPath, "original.json")
		originalData := parseJSONFile(t, originalPath)

		for _, confFile := range confFiles {
			confPath := filepath.Join(dirPath, confFile)

			// Phase 1: Conformance check (JSON only, semantic comparison)
			t.Run(dir+"/"+confFile+"/conformance", func(t *testing.T) {
				skipKey := dir + "/" + confFile + "/conformance"
				if reason, ok := skipConformance[skipKey]; ok {
					t.Skip(reason)
				}

				var stdout, stderr bytes.Buffer
				err := convert.Run("hocon2json", convert.JSONEncoder{}, []string{confPath}, strings.NewReader(""), &stdout, &stderr)
				if err != nil {
					t.Fatalf("Run() error: %v", err)
				}

				var actual map[string]any
				if err := json.Unmarshal(stdout.Bytes(), &actual); err != nil {
					t.Fatalf("parsing JSON output: %v", err)
				}

				if !reflect.DeepEqual(originalData, actual) {
					t.Errorf("conformance mismatch with original.json\ngot:  %v\nwant: %v", actual, originalData)
				}
			})

			// Phase 2: Regression check (all formats, string comparison)
			for _, f := range formats {
				format := f
				testName := dir + "/" + confFile + "/" + format.name
				t.Run(testName, func(t *testing.T) {
					if reason, ok := skipConformance[testName]; ok {
						t.Skip(reason)
					}

					expectedPath := filepath.Join(dirPath, "expected."+format.name)
					expectedBytes, err := os.ReadFile(expectedPath)
					if err != nil {
						t.Skipf("no expected.%s file: %v", format.name, err)
					}

					var stdout, stderr bytes.Buffer
					err = convert.Run("hocon2"+format.name, format.encoder, []string{confPath}, strings.NewReader(""), &stdout, &stderr)
					if err != nil {
						t.Fatalf("Run() error: %v", err)
					}

					if stdout.String() != string(expectedBytes) {
						t.Errorf("output mismatch:\ngot:\n%s\nwant:\n%s", stdout.String(), string(expectedBytes))
					}
				})
			}
		}
	}
}

func findConfFiles(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("reading directory %s: %v", dir, err)
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if name == "original.json" {
			continue
		}
		if strings.HasPrefix(name, "expected.") {
			continue
		}
		ext := filepath.Ext(name)
		if ext == ".conf" || ext == ".json" || ext == ".properties" {
			files = append(files, name)
		}
	}
	return files
}

func parseJSONFile(t *testing.T, path string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("parsing JSON %s: %v", path, err)
	}
	return m
}
