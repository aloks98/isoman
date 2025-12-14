package models

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"linux-iso-manager/internal/constants"
)

// ISOStatus represents the status of an ISO download.
type ISOStatus string

const (
	StatusPending     ISOStatus = "pending"
	StatusDownloading ISOStatus = "downloading"
	StatusVerifying   ISOStatus = "verifying"
	StatusComplete    ISOStatus = "complete"
	StatusFailed      ISOStatus = "failed"
)

// ISO represents an ISO file record in the database.
type ISO struct {
	CreatedAt    time.Time  `json:"created_at"`
	CompletedAt  *time.Time `json:"completed_at"`
	DownloadLink string     `json:"download_link"`
	ChecksumType string     `json:"checksum_type"`
	Edition      string     `json:"edition"`
	FileType     string     `json:"file_type"`
	Filename     string     `json:"filename"`
	FilePath     string     `json:"file_path"`
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	Checksum     string     `json:"checksum"`
	Arch         string     `json:"arch"`
	DownloadURL  string     `json:"download_url"`
	ChecksumURL  string     `json:"checksum_url"`
	Status       ISOStatus  `json:"status"`
	Version      string     `json:"version"`
	ErrorMessage string     `json:"error_message"`
	Progress     int        `json:"progress"`
	SizeBytes    int64      `json:"size_bytes"`
}

// CreateISORequest represents the request to create a new ISO download.
type CreateISORequest struct {
	Name         string `json:"name" binding:"required"`
	Version      string `json:"version" binding:"required"`
	Arch         string `json:"arch" binding:"required"`
	Edition      string `json:"edition"`
	DownloadURL  string `json:"download_url" binding:"required,url"`
	ChecksumURL  string `json:"checksum_url" binding:"omitempty,url"`
	ChecksumType string `json:"checksum_type" binding:"omitempty,oneof=sha256 sha512 md5"`
}

// UpdateISORequest represents the allowed fields for updating an ISO.
// Which fields are actually editable depends on the ISO's current status.
type UpdateISORequest struct {
	Name         *string `json:"name"`
	Version      *string `json:"version"`
	Arch         *string `json:"arch"`
	Edition      *string `json:"edition"`
	DownloadURL  *string `json:"download_url" binding:"omitempty,url"`
	ChecksumURL  *string `json:"checksum_url" binding:"omitempty,url"`
	ChecksumType *string `json:"checksum_type" binding:"omitempty,oneof=sha256 sha512 md5"`
}

// "Ubuntu Server" -> "ubuntu-server".
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

// DetectFileType extracts and validates the file type from a URL.
func DetectFileType(url string) (string, error) {
	ext := filepath.Ext(url)
	ext = strings.ToLower(strings.TrimPrefix(ext, "."))

	// Check if supported
	if !constants.IsSupportedFileType(ext) {
		return "", fmt.Errorf("unsupported file type: %s (supported: %v)", ext, constants.SupportedFileTypes)
	}

	return ext, nil
}

// alpine + 3.19.1 + "" + x86_64 + iso -> "alpine-3.19.1-x86_64.iso".
func GenerateFilename(name, version, edition, arch, fileType string) string {
	parts := []string{name, version}
	if edition != "" {
		parts = append(parts, edition)
	}
	parts = append(parts, arch)

	filename := strings.Join(parts, "-")
	return fmt.Sprintf("%s.%s", filename, fileType)
}

// -> "alpine/3.19.1/x86_64/alpine-3.19.1-x86_64.iso".
func GenerateFilePath(name, version, arch, filename string) string {
	return filepath.Join(name, version, arch, filename)
}

// -> "/images/alpine/3.19.1/x86_64/alpine-3.19.1-x86_64.iso".
func GenerateDownloadLink(filePath string) string {
	return "/images/" + filepath.ToSlash(filePath)
}

// This is used for checksum verification as checksum files reference the original filename.
func ExtractFilenameFromURL(url string) string {
	// Get the last part of the URL path
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

// Used for checksum verification.
func (iso *ISO) GetOriginalFilename() string {
	return ExtractFilenameFromURL(iso.DownloadURL)
}

// ComputeFields computes all derived fields for an ISO.
func (iso *ISO) ComputeFields() {
	iso.Name = NormalizeName(iso.Name)
	iso.Filename = GenerateFilename(iso.Name, iso.Version, iso.Edition, iso.Arch, iso.FileType)
	iso.FilePath = GenerateFilePath(iso.Name, iso.Version, iso.Arch, iso.Filename)
	iso.DownloadLink = GenerateDownloadLink(iso.FilePath)
}
