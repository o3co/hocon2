package convert_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/o3co/go.hocon2/internal/convert"
)

func TestEncoders(t *testing.T) {
	encoders := map[string]convert.Encoder{
		"json": convert.JSONEncoder{},
		"yaml": convert.YAMLEncoder{},
		"toml": convert.TOMLEncoder{},
	}
	testcases := []string{"basic", "nested", "array", "substitution"}

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
