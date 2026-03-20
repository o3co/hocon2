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
		out = append(out, '\n')
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
