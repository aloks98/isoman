package service

import (
	"fmt"
	"linux-iso-manager/internal/constants"
	"linux-iso-manager/internal/db"
	"linux-iso-manager/internal/download"
	"linux-iso-manager/internal/models"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ISOService handles ISO-related business logic
type ISOService struct {
	db      *db.DB
	manager *download.Manager
}

// NewISOService creates a new ISO service
func NewISOService(database *db.DB, manager *download.Manager) *ISOService {
	return &ISOService{
		db:      database,
		manager: manager,
	}
}

// CreateISORequest represents the request to create a new ISO download
type CreateISORequest struct {
	Name         string
	Version      string
	Arch         string
	Edition      string
	DownloadURL  string
	ChecksumURL  string
	ChecksumType string
}

// CreateISO creates a new ISO download
func (s *ISOService) CreateISO(req CreateISORequest) (*models.ISO, error) {
	// Detect file type from download URL
	fileType, err := DetectFileType(req.DownloadURL)
	if err != nil {
		return nil, fmt.Errorf("invalid file type: %w", err)
	}

	// Normalize name
	normalizedName := NormalizeName(req.Name)

	// Default checksum type to sha256 if checksum URL is provided
	checksumType := req.ChecksumType
	if req.ChecksumURL != "" && checksumType == "" {
		checksumType = "sha256"
	}

	// Check if ISO already exists (based on unique constraint)
	exists, err := s.db.ISOExists(normalizedName, req.Version, req.Arch, req.Edition, fileType)
	if err != nil {
		return nil, fmt.Errorf("failed to check for duplicate: %w", err)
	}

	if exists {
		existingISO, _ := s.db.GetISOByComposite(normalizedName, req.Version, req.Arch, req.Edition, fileType)
		return nil, &ISOAlreadyExistsError{ExistingISO: existingISO}
	}

	// Create ISO record
	iso := &models.ISO{
		ID:           uuid.New().String(),
		Name:         req.Name,
		Version:      req.Version,
		Arch:         req.Arch,
		Edition:      req.Edition,
		FileType:     fileType,
		DownloadURL:  req.DownloadURL,
		ChecksumURL:  req.ChecksumURL,
		ChecksumType: checksumType,
		Status:       models.StatusPending,
		Progress:     0,
		CreatedAt:    time.Now(),
	}

	// Compute derived fields (filename, file_path, download_link)
	ComputeFields(iso)

	// Save to database
	if err := s.db.CreateISO(iso); err != nil {
		return nil, fmt.Errorf("failed to create ISO: %w", err)
	}

	// Queue download
	s.manager.QueueDownload(iso)

	return iso, nil
}

// GetISO retrieves a single ISO by ID
func (s *ISOService) GetISO(id string) (*models.ISO, error) {
	return s.db.GetISO(id)
}

// ListISOs retrieves all ISOs
func (s *ISOService) ListISOs() ([]models.ISO, error) {
	return s.db.ListISOs()
}

// DeleteISO deletes an ISO and its files
func (s *ISOService) DeleteISO(id string) error {
	// Get ISO from database to validate it exists
	iso, err := s.db.GetISO(id)
	if err != nil {
		return err
	}

	// Cancel ongoing download if the ISO is being downloaded
	if iso.Status == models.StatusDownloading || iso.Status == models.StatusVerifying {
		s.manager.CancelDownload(id)
		time.Sleep(100 * time.Millisecond) // Wait for cancellation
	}

	// Delete database record
	return s.db.DeleteISO(id)
}

// RetryISO retries a failed download
func (s *ISOService) RetryISO(id string) (*models.ISO, error) {
	// Get ISO from database
	iso, err := s.db.GetISO(id)
	if err != nil {
		return nil, err
	}

	// Verify status is failed
	if iso.Status != models.StatusFailed {
		return nil, &InvalidStateError{
			CurrentStatus: string(iso.Status),
			Message:       "Only failed downloads can be retried",
		}
	}

	// Reset status, progress, and error message
	iso.Status = models.StatusPending
	iso.Progress = 0
	iso.ErrorMessage = ""
	iso.CompletedAt = nil

	// Update database
	if err := s.db.UpdateISO(iso); err != nil {
		return nil, fmt.Errorf("failed to update ISO: %w", err)
	}

	// Re-queue download
	s.manager.QueueDownload(iso)

	return iso, nil
}

// Business logic functions (moved from models package)

// NormalizeName converts a display name to a normalized storage name
// "Alpine Linux" -> "alpine-linux"
// "Ubuntu Server" -> "ubuntu-server"
func NormalizeName(name string) string {
	// Convert to lowercase and trim
	name = strings.ToLower(strings.TrimSpace(name))

	// Replace spaces, dots, and slashes with hyphens
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, ".", "-")
	name = strings.ReplaceAll(name, "/", "-")

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

	if ext == "" {
		return "", fmt.Errorf("could not detect file type from URL")
	}

	// Validate against supported file types
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
	// Normalize path separators to forward slashes
	normalizedPath := strings.ReplaceAll(filePath, "\\", "/")
	return "/images/" + normalizedPath
}

// ComputeFields computes all derived fields for an ISO
func ComputeFields(iso *models.ISO) {
	iso.Name = NormalizeName(iso.Name)
	iso.Filename = GenerateFilename(iso.Name, iso.Version, iso.Edition, iso.Arch, iso.FileType)
	iso.FilePath = GenerateFilePath(iso.Name, iso.Version, iso.Arch, iso.Filename)
	iso.DownloadLink = GenerateDownloadLink(iso.FilePath)
}

// Custom errors

// ISOAlreadyExistsError indicates that an ISO already exists
type ISOAlreadyExistsError struct {
	ExistingISO *models.ISO
}

func (e *ISOAlreadyExistsError) Error() string {
	return "ISO already exists"
}

// InvalidStateError indicates an invalid state transition
type InvalidStateError struct {
	CurrentStatus string
	Message       string
}

func (e *InvalidStateError) Error() string {
	return fmt.Sprintf("invalid state: %s (status: %s)", e.Message, e.CurrentStatus)
}
