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

	configContent := `env: development`
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err, "Failed to write test config file")

	// Load the config
	err = Load(configPath)
	require.NoError(t, err, "Failed to load config")

	// Verify values
	cfg := Get()
	assert.Equal(t, EnvDevelopment, cfg.Env)
}

// TestLoad_ProductionConfig tests loading production environment
func TestLoad_ProductionConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `env: production`
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	err = Load(configPath)
	require.NoError(t, err)

	cfg := Get()
	assert.Equal(t, EnvProduction, cfg.Env)
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

// TestLoad_EnvVarOverride tests that ENV variable overrides config file
func TestLoad_EnvVarOverride(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Config file says development
	configContent := `env: development`
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	// But ENV var says production
	os.Setenv("ENV", "production")
	defer os.Unsetenv("ENV")

	err = Load(configPath)
	require.NoError(t, err)

	cfg := Get()
	assert.Equal(t, EnvProduction, cfg.Env)
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
	assert.Equal(t, EnvDevelopment, cfg.Env) // Should use default
}
