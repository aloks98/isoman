package service

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"linux-iso-manager/internal/constants"
	"linux-iso-manager/internal/db"
	"linux-iso-manager/internal/download"
	"linux-iso-manager/internal/fileutil"
	"linux-iso-manager/internal/models"
	"linux-iso-manager/internal/pathutil"

	"github.com/google/uuid"
)

// ISOService handles ISO-related business logic.
type ISOService struct {
	db      *db.DB
	manager *download.Manager
	isoDir  string
}

// NewISOService creates a new ISO service.
func NewISOService(database *db.DB, manager *download.Manager, isoDir string) *ISOService {
	return &ISOService{
		db:      database,
		manager: manager,
		isoDir:  isoDir,
	}
}

// CreateISORequest represents the request to create a new ISO download.
type CreateISORequest struct {
	Name         string
	Version      string
	Arch         string
	Edition      string
	DownloadURL  string
	ChecksumURL  string
	ChecksumType string
}

// CreateISO creates a new ISO download.
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
		existingISO, err := s.db.GetISOByComposite(normalizedName, req.Version, req.Arch, req.Edition, fileType)
		if err != nil {
			return nil, fmt.Errorf("failed to get existing ISO: %w", err)
		}
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

// GetISO retrieves a single ISO by ID.
func (s *ISOService) GetISO(id string) (*models.ISO, error) {
	return s.db.GetISO(id)
}

// ListISOs retrieves all ISOs.
func (s *ISOService) ListISOs() ([]models.ISO, error) {
	return s.db.ListISOs()
}

// DeleteISO deletes an ISO and its files.
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

// RetryISO retries a failed download.
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

// UpdateISO updates an existing ISO.
// For failed ISOs: can edit all fields, triggers re-download.
// For complete ISOs: can only edit metadata (name, version, arch, edition), moves files.
func (s *ISOService) UpdateISO(id string, req models.UpdateISORequest) (*models.ISO, error) {
	// Get existing ISO from database
	iso, err := s.db.GetISO(id)
	if err != nil {
		return nil, err
	}

	// Validate edit is allowed
	if err := s.validateISOUpdate(iso, req); err != nil {
		return nil, err
	}

	// Store old path for file operations
	oldFilePath := iso.FilePath

	// Apply changes and check if metadata changed
	metadataChanged := s.applyISOUpdates(iso, req)

	// Recompute derived fields
	ComputeFields(iso)

	// Check for conflicts if metadata changed
	if metadataChanged {
		if err := s.checkUpdateConflict(iso); err != nil {
			return nil, err
		}
	}

	// Perform file operations and update database
	return iso, s.finalizeISOUpdate(iso, oldFilePath, metadataChanged)
}

// validateISOUpdate checks if the update is allowed based on ISO status.
func (s *ISOService) validateISOUpdate(iso *models.ISO, req models.UpdateISORequest) error {
	// Can't edit downloads in progress
	if iso.Status == models.StatusPending || iso.Status == models.StatusDownloading || iso.Status == models.StatusVerifying {
		return &InvalidStateError{
			CurrentStatus: string(iso.Status),
			Message:       "Cannot edit ISO while download is in progress",
		}
	}

	// For complete ISOs, only allow editing metadata
	if iso.Status == models.StatusComplete {
		if req.DownloadURL != nil || req.ChecksumURL != nil || req.ChecksumType != nil {
			return &InvalidStateError{
				CurrentStatus: string(iso.Status),
				Message:       "Cannot edit URLs for complete ISOs. Only metadata (name, version, arch, edition) can be changed",
			}
		}
	}

	return nil
}

// applyISOUpdates applies the requested changes to the ISO and returns whether metadata changed.
func (s *ISOService) applyISOUpdates(iso *models.ISO, req models.UpdateISORequest) bool {
	metadataChanged := false

	// Apply metadata changes
	if req.Name != nil {
		iso.Name = *req.Name
		metadataChanged = true
	}
	if req.Version != nil {
		iso.Version = *req.Version
		metadataChanged = true
	}
	if req.Arch != nil {
		iso.Arch = *req.Arch
		metadataChanged = true
	}
	if req.Edition != nil {
		iso.Edition = *req.Edition
		metadataChanged = true
	}

	// For failed ISOs, allow URL changes
	if iso.Status == models.StatusFailed {
		if req.DownloadURL != nil {
			if newFileType, err := DetectFileType(*req.DownloadURL); err == nil {
				iso.DownloadURL = *req.DownloadURL
				iso.FileType = newFileType
				metadataChanged = true
			}
		}
		if req.ChecksumURL != nil {
			iso.ChecksumURL = *req.ChecksumURL
		}
		if req.ChecksumType != nil {
			iso.ChecksumType = *req.ChecksumType
		} else if req.ChecksumURL != nil && iso.ChecksumType == "" {
			iso.ChecksumType = "sha256"
		}
	}

	return metadataChanged
}

// checkUpdateConflict checks if the updated ISO conflicts with an existing ISO.
func (s *ISOService) checkUpdateConflict(iso *models.ISO) error {
	exists, err := s.db.ISOExists(iso.Name, iso.Version, iso.Arch, iso.Edition, iso.FileType)
	if err != nil {
		return fmt.Errorf("failed to check for duplicate: %w", err)
	}

	if exists {
		existingISO, err := s.db.GetISOByComposite(iso.Name, iso.Version, iso.Arch, iso.Edition, iso.FileType)
		if err != nil {
			return fmt.Errorf("failed to get existing ISO: %w", err)
		}
		// Only error if it's a different ISO
		if existingISO.ID != iso.ID {
			return &ISOAlreadyExistsError{ExistingISO: existingISO}
		}
	}

	return nil
}

// finalizeISOUpdate performs file operations and database update based on ISO status.
func (s *ISOService) finalizeISOUpdate(iso *models.ISO, oldFilePath string, metadataChanged bool) error {
	if iso.Status == models.StatusFailed {
		// Reset and re-queue download
		iso.Status = models.StatusPending
		iso.Progress = 0
		iso.ErrorMessage = ""
		iso.CompletedAt = nil

		if err := s.db.UpdateISO(iso); err != nil {
			return fmt.Errorf("failed to update ISO: %w", err)
		}

		s.manager.QueueDownload(iso)
		return nil
	}
	if iso.Status == models.StatusComplete && metadataChanged {
		// Move files for complete ISOs with metadata changes
		if err := s.moveISOFiles(oldFilePath, iso.FilePath); err != nil {
			return fmt.Errorf("failed to move ISO files: %w", err)
		}
	}

	// Update database
	if err := s.db.UpdateISO(iso); err != nil {
		return fmt.Errorf("failed to update ISO: %w", err)
	}

	return nil
}

// moveISOFiles moves an ISO file and its checksum files from old path to new path.
func (s *ISOService) moveISOFiles(oldRelPath, newRelPath string) error {
	// Convert relative paths to absolute paths
	oldAbsPath := pathutil.ConstructISOPath(s.isoDir, oldRelPath)
	newAbsPath := pathutil.ConstructISOPath(s.isoDir, newRelPath)

	// Move the main ISO file and checksum files
	if err := fileutil.MoveFileWithExtensions(oldAbsPath, newAbsPath, constants.ChecksumExtensions...); err != nil {
		return err
	}

	// Clean up empty parent directories from the old location
	fileutil.CleanupEmptyParentDirs(oldAbsPath, s.isoDir)

	return nil
}

// Business logic functions (moved from models package)

// "Ubuntu Server" -> "ubuntu-server".
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

// DetectFileType extracts and validates the file type from a URL.
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
	// Normalize path separators to forward slashes
	normalizedPath := strings.ReplaceAll(filePath, "\\", "/")
	return "/images/" + normalizedPath
}

// ComputeFields computes all derived fields for an ISO.
func ComputeFields(iso *models.ISO) {
	iso.Name = NormalizeName(iso.Name)
	iso.Filename = GenerateFilename(iso.Name, iso.Version, iso.Edition, iso.Arch, iso.FileType)
	iso.FilePath = GenerateFilePath(iso.Name, iso.Version, iso.Arch, iso.Filename)
	iso.DownloadLink = GenerateDownloadLink(iso.FilePath)
}

// Custom errors

// ISOAlreadyExistsError indicates that an ISO already exists.
type ISOAlreadyExistsError struct {
	ExistingISO *models.ISO
}

func (e *ISOAlreadyExistsError) Error() string {
	return "ISO already exists"
}

// InvalidStateError indicates an invalid state transition.
type InvalidStateError struct {
	CurrentStatus string
	Message       string
}

func (e *InvalidStateError) Error() string {
	return fmt.Sprintf("invalid state: %s (status: %s)", e.Message, e.CurrentStatus)
}
