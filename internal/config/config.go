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

	// Default values
	DefaultServerPort   = 8080
	DefaultDatabasePath = "/var/lib/simplify/data.db"
)

// Config is the root configuration structure
type Config struct {
	Database DatabaseConfig `mapstructure:"database"`
	Env      string         `mapstructure:"env"`
	Server   ServerConfig   `mapstructure:"server"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port            int `mapstructure:"port"`
	ReadTimeout     int `mapstructure:"read_timeout"`     // seconds
	WriteTimeout    int `mapstructure:"write_timeout"`    // seconds
	IdleTimeout     int `mapstructure:"idle_timeout"`     // seconds
	ShutdownTimeout int `mapstructure:"shutdown_timeout"` // seconds
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Path string `mapstructure:"path"`
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
	if err := bindEnvVariables(); err != nil {
		return err
	}

	// Set defaults
	setDefaults()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("reading config file: %w", err)
	}

	globalConfig = &Config{}
	if err := viper.Unmarshal(globalConfig); err != nil {
		return fmt.Errorf("unmarshaling config: %w", err)
	}

	// Validate configuration
	if err := validateConfig(globalConfig); err != nil {
		return err
	}

	return nil
}

// bindEnvVariables binds environment variables to config keys
func bindEnvVariables() error {
	bindings := map[string]string{
		"env":           "SIMPLIFY_ENV",
		"server.port":   "SIMPLIFY_SERVER_PORT",
		"database.path": "SIMPLIFY_DATABASE_PATH",
	}

	for key, envVar := range bindings {
		if err := viper.BindEnv(key, envVar); err != nil {
			return fmt.Errorf("binding env variable %s: %w", envVar, err)
		}
	}

	return nil
}

// setDefaults sets default values for all configuration options
func setDefaults() {
	// Environment
	viper.SetDefault("env", EnvDevelopment)

	// Server defaults
	viper.SetDefault("server.port", DefaultServerPort)
	viper.SetDefault("server.read_timeout", 30)
	viper.SetDefault("server.write_timeout", 30)
	viper.SetDefault("server.idle_timeout", 120)
	viper.SetDefault("server.shutdown_timeout", 30)

	// Database defaults
	viper.SetDefault("database.path", DefaultDatabasePath)
}

// validateConfig validates the loaded configuration
func validateConfig(cfg *Config) error {
	// Validate env value
	if cfg.Env != EnvDevelopment && cfg.Env != EnvProduction {
		return fmt.Errorf("invalid env value %q: must be %q or %q",
			cfg.Env, EnvDevelopment, EnvProduction)
	}

	// Validate server port
	if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
		return fmt.Errorf("invalid server port %d: must be between 1 and 65535",
			cfg.Server.Port)
	}

	// Validate database path is not empty
	if cfg.Database.Path == "" {
		return fmt.Errorf("database path cannot be empty")
	}

	// Validate timeouts are positive
	if cfg.Server.ReadTimeout <= 0 {
		return fmt.Errorf("server read_timeout must be positive")
	}
	if cfg.Server.WriteTimeout <= 0 {
		return fmt.Errorf("server write_timeout must be positive")
	}
	if cfg.Server.IdleTimeout <= 0 {
		return fmt.Errorf("server idle_timeout must be positive")
	}
	if cfg.Server.ShutdownTimeout <= 0 {
		return fmt.Errorf("server shutdown_timeout must be positive")
	}

	return nil
}

// Get returns the global configuration
func Get() *Config {
	if globalConfig == nil {
		// Return default config if not loaded
		return &Config{
			Env: EnvDevelopment,
			Server: ServerConfig{
				Port:            DefaultServerPort,
				ReadTimeout:     30,
				WriteTimeout:    30,
				IdleTimeout:     120,
				ShutdownTimeout: 30,
			},
			Database: DatabaseConfig{
				Path: DefaultDatabasePath,
			},
		}
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

# HTTP Server configuration
server:
  port: 8080
  read_timeout: 30      # seconds
  write_timeout: 30     # seconds
  idle_timeout: 120     # seconds
  shutdown_timeout: 30  # seconds

# Database configuration
database:
  path: /var/lib/simplify/data.db
`)

	if err := os.WriteFile(configPath, defaultConfig, 0o600); err != nil {
		return fmt.Errorf("writing default config: %w", err)
	}

	return nil
}
