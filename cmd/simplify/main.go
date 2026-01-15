// Package main is the entry point for the Simplify CLI.
package main

import (
	"fmt"
	"os"

	"github.com/AkMo3/simplify/internal/cli"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Execute CLI
	if err := cli.Execute(); err != nil {
		return err
	}

	return nil
}
