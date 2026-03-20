package convert

import (
	"io"

	"github.com/BurntSushi/toml"
)

// TOMLEncoder encodes data as TOML.
type TOMLEncoder struct{}

func (TOMLEncoder) Encode(w io.Writer, data map[string]any) error {
	enc := toml.NewEncoder(w)
	return enc.Encode(data)
}
