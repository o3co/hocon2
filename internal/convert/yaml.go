package convert

import (
	"io"

	"gopkg.in/yaml.v3"
)

// YAMLEncoder encodes data as YAML.
type YAMLEncoder struct{}

func (YAMLEncoder) Encode(w io.Writer, data map[string]any) error {
	enc := yaml.NewEncoder(w)
	defer enc.Close()
	return enc.Encode(data)
}
