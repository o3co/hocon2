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
