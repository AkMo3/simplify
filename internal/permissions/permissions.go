// Package permissions provides utilities for checking and ensuring file system permissions.
package permissions

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AkMo3/simplify/internal/errors"
)

// EnsureDirectoryExists checks if a directory exists and creates it if not.
// Returns a PermissionError with helpful instructions if permission is denied.
func EnsureDirectoryExists(path string) error {
	// Check if directory already exists
	info, err := os.Stat(path)
	if err == nil {
		if !info.IsDir() {
			return errors.NewPermissionError(path, fmt.Sprintf("path exists but is not a directory: %s", path))
		}

		return nil
	}

	// If error is not "not exists", it's likely a permission issue
	if !os.IsNotExist(err) {
		return errors.NewPermissionErrorWithCause(path,
			formatPermissionHelp("access path", path), err)
	}

	// Try to create the directory
	if err := os.MkdirAll(path, 0o755); err != nil {
		if os.IsPermission(err) {
			return errors.NewPermissionErrorWithCause(path,
				formatPermissionHelp("create directory", path), err)
		}
		return errors.NewInternalErrorWithCause(
			fmt.Sprintf("failed to create directory %s", path), err)
	}

	return nil
}

// EnsureDirectoryWritable ensures a directory exists and is writable.
// Creates the directory if it doesn't exist.
// Returns a PermissionError with helpful instructions if permission is denied.
func EnsureDirectoryWritable(path string) error {
	// First ensure the directory exists
	if err := EnsureDirectoryExists(path); err != nil {
		return err
	}

	// Check write permission by attempting to create a temp file
	testFile := filepath.Join(path, ".simplify-write-test")
	f, err := os.Create(testFile)
	if err != nil {
		if os.IsPermission(err) {
			return errors.NewPermissionErrorWithCause(path,
				formatPermissionHelp("write to directory", path), err)
		}
		return errors.NewInternalErrorWithCause(
			fmt.Sprintf("failed to verify write permission for %s", path), err)
	}

	// Clean up test file
	f.Close()
	os.Remove(testFile)

	return nil
}

// EnsureFileWritable ensures a file's parent directory exists and is writable.
// This is useful for checking database file paths.
func EnsureFileWritable(filePath string) error {
	dir := filepath.Dir(filePath)
	return EnsureDirectoryWritable(dir)
}

// CheckReadPermission checks if a path is readable.
func CheckReadPermission(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.NewNotFoundErrorWithCause("path", path, err)
		}
		if os.IsPermission(err) {
			return errors.NewPermissionErrorWithCause(path,
				formatPermissionHelp("read path", path), err)
		}
		return errors.NewInternalErrorWithCause(
			fmt.Sprintf("failed to access path %s", path), err)
	}
	return nil
}

// formatPermissionHelp creates a helpful error message with remediation steps.
func formatPermissionHelp(action, path string) string {
	return fmt.Sprintf(
		"cannot %s: %s. "+
			"Please ensure you have the necessary permissions. "+
			"You can either:\n"+
			"  1. Change ownership: sudo chown -R $(whoami) %s\n"+
			"  2. Run with elevated privileges: sudo simplify ...\n"+
			"  3. Use a different path with --config flag",
		action, path, filepath.Dir(path))
}
