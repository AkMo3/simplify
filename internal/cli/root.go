package cli

import (
	"github.com/AkMo3/simplify/internal/config"
	"github.com/spf13/cobra"
)

var cfgFile string

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
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", config.DefaultConfigPath, "config file path")
}

// GetConfigPath returns the config file path from flag
func GetConfigPath() string {
	return cfgFile
}
