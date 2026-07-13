package convert

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/o3co/go.hocon"
)

// Encoder encodes structured data to a specific output format.
type Encoder interface {
	Encode(w io.Writer, data map[string]any) error
}

// FlagRegistrar allows an Encoder to register custom flags on the FlagSet.
type FlagRegistrar interface {
	RegisterFlags(fs *flag.FlagSet)
}

// Run parses HOCON input and encodes it using the given Encoder.
func Run(name string, enc Encoder, args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(stderr)

	var outFile string
	var overwrite bool
	var validate bool
	var envFile string
	fs.StringVar(&outFile, "o", "", "output file path")
	fs.BoolVar(&overwrite, "overwrite", false, "overwrite existing output file")
	fs.BoolVar(&validate, "validate", false, "validate HOCON syntax only (no output)")
	fs.StringVar(&envFile, "env-file", "", "load environment variables from file")

	if fr, ok := enc.(FlagRegistrar); ok {
		fr.RegisterFlags(fs)
	}

	fs.Usage = func() { printUsage(fs, name, stdout) }

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	if envFile != "" {
		if err := loadEnvFile(envFile); err != nil {
			return err
		}
	}

	cfg, err := parseInput(name, fs.Args(), stdin)
	if err != nil {
		return err
	}

	if validate {
		return nil
	}

	var m map[string]any
	if err := cfg.Unmarshal(&m); err != nil {
		return fmt.Errorf("unmarshaling config: %w", err)
	}

	if outFile == "" {
		if err := enc.Encode(stdout, m); err != nil {
			return fmt.Errorf("encoding output: %w", err)
		}
		return nil
	}

	return writeOutputFile(outFile, overwrite, enc, m)
}

// writeOutputFile encodes into a temp file in the destination directory and
// renames it into place, so a failed encode never corrupts an existing file.
func writeOutputFile(path string, overwrite bool, enc Encoder, data map[string]any) (err error) {
	if !overwrite {
		if _, statErr := os.Stat(path); statErr == nil {
			return fmt.Errorf("output file %s already exists (use -overwrite to replace)", path)
		}
	}

	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".tmp*")
	if err != nil {
		return fmt.Errorf("creating temp output file: %w", err)
	}
	defer func() {
		if err != nil {
			tmp.Close()
			os.Remove(tmp.Name())
		}
	}()

	if err = enc.Encode(tmp, data); err != nil {
		return fmt.Errorf("encoding output: %w", err)
	}
	if err = tmp.Close(); err != nil {
		return fmt.Errorf("writing output file: %w", err)
	}
	if err = os.Chmod(tmp.Name(), 0644); err != nil {
		return fmt.Errorf("setting output file permissions: %w", err)
	}
	if err = os.Rename(tmp.Name(), path); err != nil {
		return fmt.Errorf("renaming output file into place: %w", err)
	}
	return nil
}

func loadEnvFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading env file %s: %w", path, err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// direnv / dotenv files commonly write `export KEY=value`
		if rest, found := strings.CutPrefix(line, "export "); found {
			line = strings.TrimSpace(rest)
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			return fmt.Errorf("malformed line in env file %s: %q", path, line)
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		// Strip surrounding quotes
		if len(val) >= 2 && ((val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'')) {
			val = val[1 : len(val)-1]
		}
		if _, exists := os.LookupEnv(key); !exists {
			if err := os.Setenv(key, val); err != nil {
				return fmt.Errorf("setting env var %s: %w", key, err)
			}
		}
	}
	return nil
}

func parseInput(name string, args []string, stdin io.Reader) (*hocon.Config, error) {
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

func printUsage(fs *flag.FlagSet, name string, w io.Writer) {
	fmt.Fprintf(w, "Usage: %s [OPTIONS] [FILE...]\n", name)
	fmt.Fprintln(w)
	fmt.Fprintf(w, "Convert HOCON to %s.\n", formatName(name))
	fmt.Fprintln(w)
	fmt.Fprintln(w, "If no FILE is given, reads from stdin.")
	fmt.Fprintln(w, "If multiple FILEs are given, they are merged (last file takes precedence).")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Options:")
	origOut := fs.Output()
	fs.SetOutput(w)
	fs.PrintDefaults()
	fs.SetOutput(origOut)
}

func formatName(name string) string {
	format, found := strings.CutPrefix(name, "hocon2")
	if !found {
		return "the target format"
	}
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
	return "the target format"
}
