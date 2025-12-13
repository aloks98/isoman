package fileutil

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

// DeleteFile deletes a file and returns an error if it fails
// Returns nil if the file doesn't exist
func DeleteFile(path string) error {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil // File doesn't exist, nothing to do
	}

	// Attempt to delete
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to delete file %s: %w", path, err)
	}

	return nil
}

// DeleteFileSilently deletes a file and logs errors instead of returning them
// Useful for cleanup operations where errors shouldn't fail the entire operation
func DeleteFileSilently(path string) {
	if err := DeleteFile(path); err != nil {
		slog.Warn("failed to delete file", slog.Any("error", err))
	}
}

// DeleteMultipleFilesSilently deletes multiple files and logs any errors
func DeleteMultipleFilesSilently(paths ...string) {
	for _, path := range paths {
		DeleteFileSilently(path)
	}
}

// EnsureDirectory creates a directory if it doesn't exist
// Creates parent directories as needed
func EnsureDirectory(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", path, err)
	}
	return nil
}

// EnsureDirectories creates multiple directories
func EnsureDirectories(paths ...string) error {
	for _, path := range paths {
		if err := EnsureDirectory(path); err != nil {
			return err
		}
	}
	return nil
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// GetFileSize returns the size of a file in bytes
// Returns 0 if the file doesn't exist
func GetFileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}

// CleanupWithExtensions deletes a file and all files with the given extensions
// Example: CleanupWithExtensions("/path/file.iso", ".sha256", ".sha512", ".md5")
// Deletes: file.iso, file.iso.sha256, file.iso.sha512, file.iso.md5
func CleanupWithExtensions(basePath string, extensions ...string) error {
	var errs []error

	// Delete base file
	if err := DeleteFile(basePath); err != nil {
		errs = append(errs, err)
	}

	// Delete files with extensions (silently - these are optional)
	for _, ext := range extensions {
		DeleteFileSilently(basePath + ext)
	}

	if len(errs) > 0 {
		return errs[0] // Return first error
	}
	return nil
}

// EnsureParentDirectory ensures the parent directory of a file exists
func EnsureParentDirectory(filePath string) error {
	dir := filepath.Dir(filePath)
	return EnsureDirectory(dir)
}
