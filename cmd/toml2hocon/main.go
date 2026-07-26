package main

import (
	"fmt"
	"os"

	"github.com/o3co/hocon2/internal/convert"
)

func main() {
	if err := convert.RunReverse("toml2hocon", &convert.TOMLDecoder{}, os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "toml2hocon: %v\n", err)
		os.Exit(1)
	}
}
