package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoad_DefaultConfig tests loading config with default values
func TestLoad_DefaultConfig(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `env: development
server:
  port: 8080
database:
  path: /tmp/test.db`
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err, "Failed to write test config file")

	// Load the config
	err = Load(configPath)
	require.NoError(t, err, "Failed to load config")

	// Verify values
	cfg := Get()
	assert.Equal(t, EnvDevelopment, cfg.Env)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "/tmp/test.db", cfg.Database.Path)
}

// TestLoad_ProductionConfig tests loading production environment
func TestLoad_ProductionConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `env: production
server:
  port: 443
database:
  path: /var/lib/simplify/prod.db`
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	err = Load(configPath)
	require.NoError(t, err)

	cfg := Get()
	assert.Equal(t, EnvProduction, cfg.Env)
	assert.Equal(t, 443, cfg.Server.Port)
	assert.Equal(t, "/var/lib/simplify/prod.db", cfg.Database.Path)
}

// TestLoad_InvalidEnv tests that invalid env values are rejected
func TestLoad_InvalidEnv(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `env: invalid`
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	err = Load(configPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid env value")
}

// TestLoad_InvalidPort tests that invalid port values are rejected
func TestLoad_InvalidPort(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `env: development
server:
  port: 70000`
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	err = Load(configPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid server port")
}

// TestLoad_InvalidPortZero tests that port 0 is rejected
func TestLoad_InvalidPortZero(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `env: development
server:
  port: 0`
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	err = Load(configPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid server port")
}

// TestLoad_EmptyDatabasePath tests that empty database path is rejected
func TestLoad_EmptyDatabasePath(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `env: development
database:
  path: ""`
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	err = Load(configPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database path cannot be empty")
}

// TestLoad_EnvVarOverride tests that environment variables override config file
func TestLoad_EnvVarOverride(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Config file says development
	configContent := `env: development
server:
  port: 8080
database:
  path: /tmp/test.db`
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	// But ENV var says production
	t.Setenv("SIMPLIFY_ENV", "production")

	err = Load(configPath)
	require.NoError(t, err)

	cfg := Get()
	assert.Equal(t, EnvProduction, cfg.Env)
}

// TestLoad_EnvVarPortOverride tests that port env var overrides config
func TestLoad_EnvVarPortOverride(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `env: development
server:
  port: 8080
database:
  path: /tmp/test.db`
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	t.Setenv("SIMPLIFY_SERVER_PORT", "9090")

	err = Load(configPath)
	require.NoError(t, err)

	cfg := Get()
	assert.Equal(t, 9090, cfg.Server.Port)
}

// TestLoad_EnvVarDatabasePathOverride tests that database path env var overrides config
func TestLoad_EnvVarDatabasePathOverride(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `env: development
server:
  port: 8080
database:
  path: /tmp/test.db`
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	t.Setenv("SIMPLIFY_DATABASE_PATH", "/var/data/override.db")

	err = Load(configPath)
	require.NoError(t, err)

	cfg := Get()
	assert.Equal(t, "/var/data/override.db", cfg.Database.Path)
}

// TestLoad_CreatesDefaultConfig tests that a default config is created if missing
func TestLoad_CreatesDefaultConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "subdir", "config.yaml")

	// File doesn't exist yet
	_, err := os.Stat(configPath)
	assert.True(t, os.IsNotExist(err))

	// Load should create it
	err = Load(configPath)
	require.NoError(t, err)

	// File should now exist
	_, err = os.Stat(configPath)
	assert.NoError(t, err)

	// And contain default values
	cfg := Get()
	assert.Equal(t, EnvDevelopment, cfg.Env)
	assert.Equal(t, DefaultServerPort, cfg.Server.Port)
	assert.Equal(t, DefaultDatabasePath, cfg.Database.Path)
}

// TestLoad_DefaultTimeouts tests that default timeout values are applied
func TestLoad_DefaultTimeouts(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Minimal config without timeouts
	configContent := `env: development`
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	err = Load(configPath)
	require.NoError(t, err)

	cfg := Get()
	assert.Equal(t, 30, cfg.Server.ReadTimeout)
	assert.Equal(t, 30, cfg.Server.WriteTimeout)
	assert.Equal(t, 120, cfg.Server.IdleTimeout)
	assert.Equal(t, 30, cfg.Server.ShutdownTimeout)
}

// TestLoad_CustomTimeouts tests that custom timeout values are loaded
func TestLoad_CustomTimeouts(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `env: development
server:
  port: 8080
  read_timeout: 60
  write_timeout: 90
  idle_timeout: 180
  shutdown_timeout: 45`
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	err = Load(configPath)
	require.NoError(t, err)

	cfg := Get()
	assert.Equal(t, 60, cfg.Server.ReadTimeout)
	assert.Equal(t, 90, cfg.Server.WriteTimeout)
	assert.Equal(t, 180, cfg.Server.IdleTimeout)
	assert.Equal(t, 45, cfg.Server.ShutdownTimeout)
}

// TestLoad_InvalidReadTimeout tests that non-positive read timeout is rejected
func TestLoad_InvalidReadTimeout(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `env: development
server:
  read_timeout: 0`
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	err = Load(configPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "read_timeout must be positive")
}

// TestIsDevelopment tests the IsDevelopment helper
func TestIsDevelopment(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `env: development`
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	err = Load(configPath)
	require.NoError(t, err)

	assert.True(t, IsDevelopment())
	assert.False(t, IsProduction())
}

// TestIsProduction tests the IsProduction helper
func TestIsProduction(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `env: production`
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	err = Load(configPath)
	require.NoError(t, err)

	assert.True(t, IsProduction())
	assert.False(t, IsDevelopment())
}

// TestGet_BeforeLoad tests Get returns defaults before Load is called
func TestGet_BeforeLoad(t *testing.T) {
	// Reset global config
	globalConfig = nil

	cfg := Get()
	assert.NotNil(t, cfg)
	assert.Equal(t, EnvDevelopment, cfg.Env)
	assert.Equal(t, DefaultServerPort, cfg.Server.Port)
	assert.Equal(t, DefaultDatabasePath, cfg.Database.Path)
	assert.Equal(t, 30, cfg.Server.ReadTimeout)
	assert.Equal(t, 30, cfg.Server.WriteTimeout)
	assert.Equal(t, 120, cfg.Server.IdleTimeout)
	assert.Equal(t, 30, cfg.Server.ShutdownTimeout)
}

// TestLoad_MalformedYAML tests handling of invalid YAML
func TestLoad_MalformedYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Invalid YAML
	configContent := `env: [invalid yaml`
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	err = Load(configPath)
	assert.Error(t, err)
}

// TestLoad_EmptyConfig tests loading an empty config file uses defaults
func TestLoad_EmptyConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Empty file
	err := os.WriteFile(configPath, []byte(""), 0o644)
	require.NoError(t, err)

	err = Load(configPath)
	require.NoError(t, err)

	cfg := Get()
	assert.Equal(t, EnvDevelopment, cfg.Env)
	assert.Equal(t, DefaultServerPort, cfg.Server.Port)
	assert.Equal(t, DefaultDatabasePath, cfg.Database.Path)
}

// TestValidateConfig tests the validation function directly
func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name        string
		cfg         *Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			cfg: &Config{
				Env: EnvDevelopment,
				Server: ServerConfig{
					Port:            8080,
					ReadTimeout:     30,
					WriteTimeout:    30,
					IdleTimeout:     120,
					ShutdownTimeout: 30,
				},
				Database: DatabaseConfig{
					Path: "/var/lib/simplify/data.db",
				},
			},
			expectError: false,
		},
		{
			name: "invalid env",
			cfg: &Config{
				Env: "staging",
				Server: ServerConfig{
					Port:            8080,
					ReadTimeout:     30,
					WriteTimeout:    30,
					IdleTimeout:     120,
					ShutdownTimeout: 30,
				},
				Database: DatabaseConfig{
					Path: "/var/lib/simplify/data.db",
				},
			},
			expectError: true,
			errorMsg:    "invalid env value",
		},
		{
			name: "negative port",
			cfg: &Config{
				Env: EnvDevelopment,
				Server: ServerConfig{
					Port:            -1,
					ReadTimeout:     30,
					WriteTimeout:    30,
					IdleTimeout:     120,
					ShutdownTimeout: 30,
				},
				Database: DatabaseConfig{
					Path: "/var/lib/simplify/data.db",
				},
			},
			expectError: true,
			errorMsg:    "invalid server port",
		},
		{
			name: "negative write timeout",
			cfg: &Config{
				Env: EnvDevelopment,
				Server: ServerConfig{
					Port:            8080,
					ReadTimeout:     30,
					WriteTimeout:    -1,
					IdleTimeout:     120,
					ShutdownTimeout: 30,
				},
				Database: DatabaseConfig{
					Path: "/var/lib/simplify/data.db",
				},
			},
			expectError: true,
			errorMsg:    "write_timeout must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.cfg)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
