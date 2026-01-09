// Package cli provides the command-line interface for Simplify.
package cli

import (
	"fmt"

	"github.com/AkMo3/simplify/internal/config"
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
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", config.DefaultConfigPath, "config file path")
	rootCmd.AddCommand(versionCmd)
}

// GetConfigPath returns the config file path from flag
func GetConfigPath() string {
	return cfgFile
}
