package models

import (
	"fmt"
	"linux-iso-manager/internal/constants"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// ISOStatus represents the status of an ISO download
type ISOStatus string

const (
	StatusPending     ISOStatus = "pending"
	StatusDownloading ISOStatus = "downloading"
	StatusVerifying   ISOStatus = "verifying"
	StatusComplete    ISOStatus = "complete"
	StatusFailed      ISOStatus = "failed"
)

// ISO represents an ISO file record in the database
type ISO struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`          // Normalized: "alpine", "ubuntu"
	Version      string     `json:"version"`       // "3.19.1", "24.04", "rolling", etc.
	Arch         string     `json:"arch"`          // "x86_64", "aarch64", "arm64"
	Edition      string     `json:"edition"`       // "minimal", "desktop", "server", "" (optional)
	FileType     string     `json:"file_type"`     // "iso", "qcow2", "vmdk", etc.
	Filename     string     `json:"filename"`      // Computed: "alpine-3.19.1-minimal-x86_64.iso"
	FilePath     string     `json:"file_path"`     // Computed: "alpine/3.19.1/x86_64/alpine-3.19.1-minimal-x86_64.iso"
	DownloadLink string     `json:"download_link"` // Computed: "/images/alpine/3.19.1/x86_64/alpine-3.19.1-minimal-x86_64.iso"
	SizeBytes    int64      `json:"size_bytes"`
	Checksum     string     `json:"checksum"`
	ChecksumType string     `json:"checksum_type"`
	DownloadURL  string     `json:"download_url"`
	ChecksumURL  string     `json:"checksum_url"`
	Status       ISOStatus  `json:"status"`
	Progress     int        `json:"progress"`
	ErrorMessage string     `json:"error_message"`
	CreatedAt    time.Time  `json:"created_at"`
	CompletedAt  *time.Time `json:"completed_at"`
}

// CreateISORequest represents the request to create a new ISO download
type CreateISORequest struct {
	Name         string `json:"name" binding:"required"`
	Version      string `json:"version" binding:"required"`
	Arch         string `json:"arch" binding:"required"`
	Edition      string `json:"edition"`
	DownloadURL  string `json:"download_url" binding:"required,url"`
	ChecksumURL  string `json:"checksum_url" binding:"omitempty,url"`
	ChecksumType string `json:"checksum_type" binding:"omitempty,oneof=sha256 sha512 md5"`
}

// NormalizeName converts a display name to a normalized storage name
// "Alpine Linux" -> "alpine"
// "Ubuntu Server" -> "ubuntu-server"
func NormalizeName(name string) string {
	// Convert to lowercase and trim
	name = strings.ToLower(strings.TrimSpace(name))

	// Replace spaces with hyphens
	name = strings.ReplaceAll(name, " ", "-")

	// Remove any characters that are not alphanumeric or hyphens
	reg := regexp.MustCompile(`[^a-z0-9-]`)
	name = reg.ReplaceAllString(name, "")

	// Remove consecutive hyphens
	reg = regexp.MustCompile(`-+`)
	name = reg.ReplaceAllString(name, "-")

	// Trim hyphens from start and end
	name = strings.Trim(name, "-")

	return name
}

// DetectFileType extracts and validates the file type from a URL
func DetectFileType(url string) (string, error) {
	ext := filepath.Ext(url)
	ext = strings.ToLower(strings.TrimPrefix(ext, "."))

	// Check if supported
	if !constants.IsSupportedFileType(ext) {
		return "", fmt.Errorf("unsupported file type: %s (supported: %v)", ext, constants.SupportedFileTypes)
	}

	return ext, nil
}

// GenerateFilename creates a filename from components
// alpine + 3.19.1 + minimal + x86_64 + iso -> "alpine-3.19.1-minimal-x86_64.iso"
// alpine + 3.19.1 + "" + x86_64 + iso -> "alpine-3.19.1-x86_64.iso"
func GenerateFilename(name, version, edition, arch, fileType string) string {
	parts := []string{name, version}
	if edition != "" {
		parts = append(parts, edition)
	}
	parts = append(parts, arch)

	filename := strings.Join(parts, "-")
	return fmt.Sprintf("%s.%s", filename, fileType)
}

// GenerateFilePath creates the full relative path for storage
// alpine + 3.19.1 + x86_64 + alpine-3.19.1-x86_64.iso
// -> "alpine/3.19.1/x86_64/alpine-3.19.1-x86_64.iso"
func GenerateFilePath(name, version, arch, filename string) string {
	return filepath.Join(name, version, arch, filename)
}

// GenerateDownloadLink creates the public download URL
// "alpine/3.19.1/x86_64/alpine-3.19.1-x86_64.iso"
// -> "/images/alpine/3.19.1/x86_64/alpine-3.19.1-x86_64.iso"
func GenerateDownloadLink(filePath string) string {
	return "/images/" + filepath.ToSlash(filePath)
}

// ExtractFilenameFromURL extracts the original filename from a download URL
// This is used for checksum verification as checksum files reference the original filename
func ExtractFilenameFromURL(url string) string {
	// Get the last part of the URL path
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

// GetOriginalFilename returns the original filename from the download URL
// Used for checksum verification
func (iso *ISO) GetOriginalFilename() string {
	return ExtractFilenameFromURL(iso.DownloadURL)
}

// ComputeFields computes all derived fields for an ISO
func (iso *ISO) ComputeFields() {
	iso.Name = NormalizeName(iso.Name)
	iso.Filename = GenerateFilename(iso.Name, iso.Version, iso.Edition, iso.Arch, iso.FileType)
	iso.FilePath = GenerateFilePath(iso.Name, iso.Version, iso.Arch, iso.Filename)
	iso.DownloadLink = GenerateDownloadLink(iso.FilePath)
}
