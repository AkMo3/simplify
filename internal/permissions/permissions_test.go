package permissions

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AkMo3/simplify/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnsureDirectoryExists(t *testing.T) {
	t.Run("creates directory if not exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		newDir := filepath.Join(tmpDir, "subdir", "nested")

		err := EnsureDirectoryExists(newDir)
		require.NoError(t, err)

		info, err := os.Stat(newDir)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
	})

	t.Run("succeeds if directory already exists", func(t *testing.T) {
		tmpDir := t.TempDir()

		err := EnsureDirectoryExists(tmpDir)
		require.NoError(t, err)
	})

	t.Run("fails if path is a file", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "file.txt")

		// Create a file
		f, err := os.Create(filePath)
		require.NoError(t, err)
		f.Close()

		err = EnsureDirectoryExists(filePath)
		require.Error(t, err)
		assert.True(t, errors.IsPermissionError(err))
		assert.Contains(t, err.Error(), "not a directory")
	})
}

func TestEnsureDirectoryWritable(t *testing.T) {
	t.Run("succeeds for writable directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		err := EnsureDirectoryWritable(tmpDir)
		require.NoError(t, err)
	})

	t.Run("creates and verifies new directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		newDir := filepath.Join(tmpDir, "new-writable-dir")

		err := EnsureDirectoryWritable(newDir)
		require.NoError(t, err)

		// Verify directory exists
		info, err := os.Stat(newDir)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
	})

	t.Run("cleans up test file", func(t *testing.T) {
		tmpDir := t.TempDir()

		err := EnsureDirectoryWritable(tmpDir)
		require.NoError(t, err)

		// Verify test file was cleaned up
		testFile := filepath.Join(tmpDir, ".simplify-write-test")
		_, err = os.Stat(testFile)
		assert.True(t, os.IsNotExist(err), "test file should be cleaned up")
	})
}

func TestEnsureFileWritable(t *testing.T) {
	t.Run("ensures parent directory is writable", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "subdir", "data.db")

		err := EnsureFileWritable(filePath)
		require.NoError(t, err)

		// Verify parent directory was created
		parentDir := filepath.Dir(filePath)
		info, err := os.Stat(parentDir)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
	})
}

func TestCheckReadPermission(t *testing.T) {
	t.Run("succeeds for readable path", func(t *testing.T) {
		tmpDir := t.TempDir()

		err := CheckReadPermission(tmpDir)
		require.NoError(t, err)
	})

	t.Run("fails for non-existent path", func(t *testing.T) {
		err := CheckReadPermission("/non/existent/path")
		require.Error(t, err)
		assert.True(t, errors.IsNotFound(err))
	})
}

func TestFormatPermissionHelp(t *testing.T) {
	msg := formatPermissionHelp("write to directory", "/var/lib/simplify")

	assert.Contains(t, msg, "cannot write to directory")
	assert.Contains(t, msg, "/var/lib/simplify")
	assert.Contains(t, msg, "chown")
	assert.Contains(t, msg, "sudo")
	assert.Contains(t, msg, "--config")
}
