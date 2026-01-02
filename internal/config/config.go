// Package config handles application configuration loading and management.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	DefaultConfigPath = "/etc/simplify/config.yaml"
	EnvDevelopment    = "development"
	EnvProduction     = "production"
)

type Config struct {
	Env string `mapstructure:"env"`
}

var globalConfig *Config

// Load initializes configuration from file, env vars, with priority:
// config file -> environment variables -> CLI flags
func Load(configPath string) error {
	if configPath == "" {
		configPath = DefaultConfigPath
	}

	// Reset viper for testing
	viper.Reset()

	// Create default config if it doesn't exist
	if err := ensureConfigExists(configPath); err != nil {
		return fmt.Errorf("ensuring config exists: %w", err)
	}

	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// Bind environment variables
	if err := viper.BindEnv("env", "ENV"); err != nil {
		return fmt.Errorf("binding env variable: %w", err)
	}

	// Set defaults
	viper.SetDefault("env", EnvDevelopment)

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("reading config file: %w", err)
	}

	globalConfig = &Config{}
	if err := viper.Unmarshal(globalConfig); err != nil {
		return fmt.Errorf("unmarshaling config: %w", err)
	}

	// Validate env value
	if globalConfig.Env != EnvDevelopment && globalConfig.Env != EnvProduction {
		return fmt.Errorf("invalid env value %q: must be %q or %q",
			globalConfig.Env, EnvDevelopment, EnvProduction)
	}

	return nil
}

// Get returns the global configuration
func Get() *Config {
	if globalConfig == nil {
		// Return default config if not loaded
		return &Config{Env: EnvDevelopment}
	}
	return globalConfig
}

// IsDevelopment returns true if running in development mode
func IsDevelopment() bool {
	return Get().Env == EnvDevelopment
}

// IsProduction returns true if running in production mode
func IsProduction() bool {
	return Get().Env == EnvProduction
}

// ensureConfigExists creates the config file with defaults if it doesn't exist
func ensureConfigExists(configPath string) error {
	if _, err := os.Stat(configPath); err == nil {
		// File exists
		return nil
	}

	// Create directory if needed
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating config directory %s: %w", dir, err)
	}

	// Create default config (0o600 for security - may contain secrets later)
	defaultConfig := []byte(`# Simplify Configuration
# Environment: development | production
env: development
`)

	if err := os.WriteFile(configPath, defaultConfig, 0o600); err != nil {
		return fmt.Errorf("writing default config: %w", err)
	}

	return nil
}
