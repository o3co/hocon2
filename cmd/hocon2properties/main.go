package main

import (
	"fmt"
	"os"

	"github.com/o3co/go.hocon2/internal/convert"
)

func main() {
	if err := convert.Run("hocon2properties", convert.PropertiesEncoder{}, os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "hocon2properties: %v\n", err)
		os.Exit(1)
	}
}
