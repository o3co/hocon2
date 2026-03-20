package convert

import (
	"io"

	"github.com/magiconair/properties"
	"github.com/o3co/hocon2/internal/flatten"
)

// PropertiesEncoder encodes data as Java .properties format.
type PropertiesEncoder struct{}

func (PropertiesEncoder) Encode(w io.Writer, data map[string]any) error {
	flat := flatten.Flatten(data)
	p := properties.LoadMap(flat)
	p.Sort()
	_, err := p.Write(w, properties.UTF8)
	return err
}
