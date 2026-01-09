// Package main is the entry point for the Simplify CLI.
package main

import (
	"fmt"
	"os"

	"github.com/AkMo3/simplify/internal/cli"
	"github.com/AkMo3/simplify/internal/config"
	"github.com/AkMo3/simplify/internal/logger"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Load configuration first
	if err := config.Load(cli.GetConfigPath()); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize logger based on environment
	if err := logger.Init(); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer logger.Sync()

	// Execute CLI
	if err := cli.Execute(); err != nil {
		return err
	}

	return nil
}
