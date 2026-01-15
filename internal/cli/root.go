// Package cli provides the command-line interface for Simplify.
package cli

import (
	"fmt"
	"os"

	"github.com/AkMo3/simplify/internal/config"
	"github.com/AkMo3/simplify/internal/logger"
	"github.com/spf13/cobra"
)

var (
	cfgFile string

	// Version information (set via ldflags during build)
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "simplify",
	Short: "A self-hosted PaaS for containerized applications",
	Long: `Simplify is a Platform-as-a-Service for deploying and managing 
containerized applications across distributed infrastructure.

Features:
  - Container orchestration via Podman
  - HTTP API for management
  - Declarative desired-state reconciliation
  - Health monitoring and automatic recovery`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Simplify %s\n", Version)
		fmt.Printf("  Git commit: %s\n", GitCommit)
		fmt.Printf("  Build date: %s\n", BuildDate)
	},
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", config.DefaultConfigPath, "config file path")

	// Initialize logger after config is loaded but before command execution
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		return initLogger()
	}

	rootCmd.AddCommand(versionCmd)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if err := config.Load(cfgFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}
}

// initLogger initializes the logger
func initLogger() error {
	return logger.Init()
}

// GetConfigPath returns the config file path from flag
func GetConfigPath() string {
	return cfgFile
}
