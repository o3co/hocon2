package main

import (
	"fmt"
	"os"

	"github.com/o3co/hocon2/internal/convert"
)

func main() {
	if err := convert.Run("hocon2yaml", convert.YAMLEncoder{}, os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "hocon2yaml: %v\n", err)
		os.Exit(1)
	}
}
