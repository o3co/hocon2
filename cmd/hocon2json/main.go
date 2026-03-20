package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/o3co/go.hocon"
)

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "hocon2json: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	var cfg *hocon.Config

	switch len(args) {
	case 0:
		// Read from stdin
		data, err := io.ReadAll(stdin)
		if err != nil {
			return fmt.Errorf("reading stdin: %w", err)
		}
		cfg, err = hocon.ParseString(string(data))
		if err != nil {
			return fmt.Errorf("parsing HOCON: %w", err)
		}
	case 1:
		if args[0] == "-h" || args[0] == "--help" {
			printUsage(stdout)
			return nil
		}
		var err error
		cfg, err = hocon.ParseFile(args[0])
		if err != nil {
			return fmt.Errorf("parsing HOCON: %w", err)
		}
	default:
		printUsage(stderr)
		return fmt.Errorf("too many arguments")
	}

	var m map[string]any
	if err := cfg.Unmarshal(&m); err != nil {
		return fmt.Errorf("converting to JSON: %w", err)
	}

	enc := json.NewEncoder(stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(m); err != nil {
		return fmt.Errorf("encoding JSON: %w", err)
	}

	return nil
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage: hocon2json [FILE]")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Convert HOCON to JSON.")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "If FILE is omitted, reads from stdin.")
}
