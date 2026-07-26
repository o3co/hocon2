package convert

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/o3co/go.hocon"
)

// Decoder reads a foreign config format and produces a resolved HOCON config.
// It is the reverse of Encoder: where Encoder turns a value tree into JSON /
// YAML / TOML / Properties text, a Decoder turns that text back into a value
// tree, which RunReverse then renders as HOCON.
type Decoder interface {
	// Decode parses data (the whole input document). origin names the source in
	// error messages. The returned Config is resolved and holds only data.
	Decode(data []byte, origin string) (*hocon.Config, error)
	// Format is the human-readable source format name, for usage text.
	Format() string
}

// RunReverse reads a foreign-format document and renders it as HOCON.
//
// It mirrors Run: same -o / -overwrite / -validate flags and the same atomic
// output write. There is no -env-file — a foreign document carries no
// substitutions to resolve, so no environment is consulted (spec F0.2).
func RunReverse(name string, dec Decoder, args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(stderr)

	var outFile string
	var overwrite bool
	var validate bool
	fs.StringVar(&outFile, "o", "", "output file path")
	fs.BoolVar(&overwrite, "overwrite", false, "overwrite existing output file")
	fs.BoolVar(&validate, "validate", false, "validate input syntax only (no output)")

	fs.Usage = func() { printReverseUsage(fs, name, dec.Format(), stdout) }

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	data, origin, err := readReverseInput(fs.Args(), stdin)
	if err != nil {
		return err
	}

	cfg, err := dec.Decode(data, origin)
	if err != nil {
		return fmt.Errorf("parsing %s: %w", dec.Format(), err)
	}

	if validate {
		return nil
	}

	text, err := cfg.RenderHOCON()
	if err != nil {
		return fmt.Errorf("rendering HOCON: %w", err)
	}

	write := func(w io.Writer) error {
		_, werr := io.WriteString(w, text)
		return werr
	}
	if outFile == "" {
		if err := write(stdout); err != nil {
			return fmt.Errorf("writing output: %w", err)
		}
		return nil
	}
	return writeOutputFile(outFile, overwrite, write)
}

// readReverseInput reads the single input document. Unlike Run, reverse
// conversion does not merge multiple inputs: merging belongs to HOCON's own
// fallback semantics, and there is no meaningful merge of, say, two YAML files
// into one HOCON document that the caller could not express more clearly by
// converting each.
func readReverseInput(args []string, stdin io.Reader) (data []byte, origin string, err error) {
	switch len(args) {
	case 0:
		data, err = io.ReadAll(stdin)
		if err != nil {
			return nil, "", fmt.Errorf("reading stdin: %w", err)
		}
		return data, "(stdin)", nil
	case 1:
		data, err = os.ReadFile(args[0])
		if err != nil {
			return nil, "", fmt.Errorf("reading %s: %w", args[0], err)
		}
		return data, args[0], nil
	default:
		return nil, "", fmt.Errorf("reverse conversion takes at most one input file, got %d", len(args))
	}
}

func printReverseUsage(fs *flag.FlagSet, name, format string, w io.Writer) {
	fmt.Fprintf(w, "Usage: %s [OPTIONS] [FILE]\n\n", name)
	fmt.Fprintf(w, "Convert %s to HOCON.\n\n", format)
	fmt.Fprintln(w, "If no FILE is given, reads from stdin.")
	fmt.Fprintln(w, "\nOptions:")
	origOut := fs.Output()
	fs.SetOutput(w)
	fs.PrintDefaults()
	fs.SetOutput(origOut)
}
