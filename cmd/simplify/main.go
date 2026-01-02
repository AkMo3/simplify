package main

import (
	"fmt"
	"os"

	"github.com/AkMo3/simplify/internal/cli"
	"github.com/AkMo3/simplify/internal/config"
	"github.com/AkMo3/simplify/internal/logger"
)

func main() {

	// Load configuration
	if err := config.Load(cli.GetConfigPath()); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	if err := logger.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to Initialize logger :%v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting simplify", "env", config.Get().Env)

	// Execute CLI
	if err := cli.Execute(); err != nil {
		logger.Error("Command failed", "error", err)
		os.Exit(1)
	}
}
