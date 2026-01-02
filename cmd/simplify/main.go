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
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Load configuration
	if err := config.Load(cli.GetConfigPath()); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize logger
	if err := logger.Init(); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer logger.Sync()

	logger.Info("Starting simplify", "env", config.Get().Env)

	// Execute CLI
	if err := cli.Execute(); err != nil {
		logger.Error("Command failed", "error", err)
		return err
	}

	return nil
}
