package cli

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "simplify",
	Short: "A self-hosted PaaS for containerized applications",
	Long: `Simplify is a Platform-as-a-Service for deploying and managing 
containerized applications across distributed infrastructure.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags will go here
}
