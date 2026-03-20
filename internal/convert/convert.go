package convert

import (
	"fmt"
	"io"

	"github.com/o3co/go.hocon"
)

// Encoder encodes structured data to a specific output format.
type Encoder interface {
	Encode(w io.Writer, data map[string]any) error
}

// Run parses HOCON input and encodes it using the given Encoder.
func Run(name string, enc Encoder, args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	cfg, err := parseInput(name, args, stdin, stdout, stderr)
	if cfg == nil || err != nil {
		return err
	}

	var m map[string]any
	if err := cfg.Unmarshal(&m); err != nil {
		return fmt.Errorf("unmarshaling config: %w", err)
	}

	if err := enc.Encode(stdout, m); err != nil {
		return fmt.Errorf("encoding output: %w", err)
	}

	return nil
}

func parseInput(name string, args []string, stdin io.Reader, stdout, stderr io.Writer) (*hocon.Config, error) {
	for _, a := range args {
		if a == "-h" || a == "--help" {
			printUsage(name, stdout)
			return nil, nil
		}
	}

	switch len(args) {
	case 0:
		data, err := io.ReadAll(stdin)
		if err != nil {
			return nil, fmt.Errorf("reading stdin: %w", err)
		}
		cfg, err := hocon.ParseString(string(data))
		if err != nil {
			return nil, fmt.Errorf("parsing HOCON: %w", err)
		}
		return cfg, nil

	case 1:
		cfg, err := hocon.ParseFile(args[0])
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", args[0], err)
		}
		return cfg, nil

	default:
		configs := make([]*hocon.Config, len(args))
		for i, path := range args {
			cfg, err := hocon.ParseFile(path)
			if err != nil {
				return nil, fmt.Errorf("parsing %s: %w", path, err)
			}
			configs[i] = cfg
		}
		merged := configs[len(configs)-1]
		for i := len(configs) - 2; i >= 0; i-- {
			merged = merged.WithFallback(configs[i])
		}
		return merged, nil
	}
}

func printUsage(name string, w io.Writer) {
	fmt.Fprintf(w, "Usage: %s [FILE...]\n", name)
	fmt.Fprintln(w, "")
	fmt.Fprintf(w, "Convert HOCON to %s.\n", formatName(name))
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "If no FILE is given, reads from stdin.")
	fmt.Fprintln(w, "If multiple FILEs are given, they are merged (last file takes precedence).")
}

func formatName(name string) string {
	if len(name) > 6 {
		format := name[6:]
		switch format {
		case "json":
			return "JSON"
		case "yaml":
			return "YAML"
		case "toml":
			return "TOML"
		case "properties":
			return "Properties"
		}
	}
	return "the target format"
}
