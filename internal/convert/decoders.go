package convert

import (
	"github.com/o3co/go.hocon"
	"github.com/o3co/go.hocon/adapters/properties"
	"github.com/o3co/go.hocon/adapters/toml"
	"github.com/o3co/go.hocon/adapters/yaml"
)

// The decoders reuse go.hocon/adapters so the foreign-format mapping rules
// (F-items: TOML dates to strings, YAML merge keys and Norway handling,
// Properties syntax, BOM stripping) have a single implementation shared with
// the parser libraries, rather than a second copy that could drift.

// JSONDecoder reads JSON. JSON is a subset of HOCON, so the parser accepts it
// directly — no adapter is needed, and none exists. Routing it through the
// emitter still normalizes it to idiomatic HOCON (unquoted keys, `=`, no
// commas).
type JSONDecoder struct{}

func (JSONDecoder) Decode(data []byte, origin string) (*hocon.Config, error) {
	return hocon.ParseStringWithOptions(string(data),
		hocon.DefaultParseOptions().WithOriginDescription(origin))
}

func (JSONDecoder) Format() string { return "JSON" }

// YAMLDecoder reads YAML via adapters/yaml (goccy/go-yaml, YAML 1.2 core).
type YAMLDecoder struct{}

func (YAMLDecoder) Decode(data []byte, origin string) (*hocon.Config, error) {
	return yaml.Parse(data, origin)
}

func (YAMLDecoder) Format() string { return "YAML" }

// TOMLDecoder reads TOML via adapters/toml (pelletier/go-toml/v2).
type TOMLDecoder struct{}

func (TOMLDecoder) Decode(data []byte, origin string) (*hocon.Config, error) {
	return toml.Parse(data, origin)
}

func (TOMLDecoder) Format() string { return "TOML" }

// PropertiesDecoder reads java.util.Properties via adapters/properties.
type PropertiesDecoder struct{}

func (PropertiesDecoder) Decode(data []byte, origin string) (*hocon.Config, error) {
	return properties.Parse(data, origin)
}

func (PropertiesDecoder) Format() string { return "Properties" }
