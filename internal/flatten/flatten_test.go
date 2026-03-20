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
