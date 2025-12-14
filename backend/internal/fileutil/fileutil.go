package fileutil

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

// Returns nil if the file doesn't exist.
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

// Useful for cleanup operations where errors shouldn't fail the entire operation.
func DeleteFileSilently(path string) {
	if err := DeleteFile(path); err != nil {
		slog.Warn("failed to delete file", slog.Any("error", err))
	}
}

// DeleteMultipleFilesSilently deletes multiple files and logs any errors.
func DeleteMultipleFilesSilently(paths ...string) {
	for _, path := range paths {
		DeleteFileSilently(path)
	}
}

// Creates parent directories as needed.
func EnsureDirectory(path string) error {
	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", path, err)
	}
	return nil
}

// EnsureDirectories creates multiple directories.
func EnsureDirectories(paths ...string) error {
	for _, path := range paths {
		if err := EnsureDirectory(path); err != nil {
			return err
		}
	}
	return nil
}

// FileExists checks if a file exists.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Returns 0 if the file doesn't exist.
func GetFileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}

// Deletes: file.iso, file.iso.sha256, file.iso.sha512, file.iso.md5.
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

// EnsureParentDirectory ensures the parent directory of a file exists.
func EnsureParentDirectory(filePath string) error {
	dir := filepath.Dir(filePath)
	return EnsureDirectory(dir)
}

// MoveFile moves a file from oldPath to newPath.
// Creates parent directories for newPath if needed.
// Returns nil if oldPath doesn't exist.
func MoveFile(oldPath, newPath string) error {
	// Check if source file exists
	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		return nil // Source doesn't exist, nothing to move
	}

	// Ensure destination directory exists
	if err := EnsureParentDirectory(newPath); err != nil {
		return err
	}

	// Move the file
	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("failed to move file from %s to %s: %w", oldPath, newPath, err)
	}

	return nil
}

// MoveFileWithExtensions moves a file and its associated extension files.
// For example, moves file.iso, file.iso.sha256, file.iso.sha512, file.iso.md5.
func MoveFileWithExtensions(oldPath, newPath string, extensions ...string) error {
	// Move base file
	if err := MoveFile(oldPath, newPath); err != nil {
		return err
	}

	// Move extension files (silently - these are optional)
	for _, ext := range extensions {
		oldExtPath := oldPath + ext
		newExtPath := newPath + ext
		if err := MoveFile(oldExtPath, newExtPath); err != nil {
			slog.Warn("failed to move extension file", slog.String("ext", ext), slog.Any("error", err))
		}
	}

	return nil
}

// CleanupEmptyParentDirs removes empty parent directories up to the given root.
// Useful for cleaning up directory structure after moving/deleting files.
func CleanupEmptyParentDirs(filePath, root string) {
	dir := filepath.Dir(filePath)

	// Keep removing empty directories until we hit root or a non-empty dir
	for dir != root && dir != "." && dir != "/" {
		if err := os.Remove(dir); err != nil {
			// Directory not empty or other error - stop here
			break
		}
		slog.Debug("removed empty directory", slog.String("dir", dir))
		dir = filepath.Dir(dir)
	}
}
