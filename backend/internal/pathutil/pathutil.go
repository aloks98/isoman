package pathutil

import (
	"path/filepath"
)

// ConstructISOPath constructs the full path to an ISO file
// Example: ConstructISOPath("/data/isos", "alpine/3.19.1/x86_64/alpine-3.19.1-x86_64.iso")
// Returns: "/data/isos/alpine/3.19.1/x86_64/alpine-3.19.1-x86_64.iso"
func ConstructISOPath(isoDir, filePath string) string {
	return filepath.Join(isoDir, filePath)
}

// ConstructTempPath constructs the path to a temporary file
// Example: ConstructTempPath("/data/isos", "alpine-3.19.1-x86_64.iso")
// Returns: "/data/isos/.tmp/alpine-3.19.1-x86_64.iso"
func ConstructTempPath(isoDir, filename string) string {
	return filepath.Join(isoDir, ".tmp", filename)
}

// ConstructChecksumPath constructs the path to a checksum file
// Example: ConstructChecksumPath("/data/isos/file.iso", "sha256")
// Returns: "/data/isos/file.iso.sha256"
func ConstructChecksumPath(isoPath, checksumType string) string {
	return isoPath + "." + checksumType
}

// GetTempDir returns the temp directory path
func GetTempDir(isoDir string) string {
	return filepath.Join(isoDir, ".tmp")
}

// GetDBDir returns the database directory path
func GetDBDir(dataDir string) string {
	return filepath.Join(dataDir, "db")
}

// GetISODir returns the ISO storage directory path
func GetISODir(dataDir string) string {
	return filepath.Join(dataDir, "isos")
}

// GetDBPath returns the full database file path
func GetDBPath(dataDir string) string {
	return filepath.Join(dataDir, "db", "isos.db")
}
