package api

import (
	"fmt"
	"linux-iso-manager/internal/db"
	"linux-iso-manager/internal/download"
	"linux-iso-manager/internal/models"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handlers holds references to database, download manager, and storage directory
type Handlers struct {
	db      *db.DB
	manager *download.Manager
	isoDir  string
}

// NewHandlers creates a new Handlers instance
func NewHandlers(database *db.DB, manager *download.Manager, isoDir string) *Handlers {
	return &Handlers{
		db:      database,
		manager: manager,
		isoDir:  isoDir,
	}
}

// ListISOs returns all ISOs ordered by created_at DESC
func (h *Handlers) ListISOs(c *gin.Context) {
	isos, err := h.db.ListISOs()
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, ErrCodeInternalError, "Failed to list ISOs")
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{
		"isos": isos,
	})
}

// GetISO returns a single ISO by ID
func (h *Handlers) GetISO(c *gin.Context) {
	id := c.Param("id")

	iso, err := h.db.GetISO(id)
	if err != nil {
		ErrorResponse(c, http.StatusNotFound, ErrCodeNotFound, "ISO not found")
		return
	}

	SuccessResponse(c, http.StatusOK, iso)
}

// CreateISO creates a new ISO download
func (h *Handlers) CreateISO(c *gin.Context) {
	var req models.CreateISORequest

	// Bind and validate JSON request
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponseWithDetails(c, http.StatusBadRequest, ErrCodeValidationFailed, "Invalid request body", err.Error())
		return
	}

	// Detect file type from download URL
	fileType, err := models.DetectFileType(req.DownloadURL)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, ErrCodeValidationFailed, err.Error())
		return
	}

	// Normalize name
	normalizedName := models.NormalizeName(req.Name)

	// Default checksum type to sha256 if checksum URL is provided
	checksumType := req.ChecksumType
	if req.ChecksumURL != "" && checksumType == "" {
		checksumType = "sha256"
	}

	// Check if ISO already exists (based on unique constraint)
	exists, err := h.db.ISOExists(normalizedName, req.Version, req.Arch, req.Edition, fileType)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, ErrCodeInternalError, "Failed to check for duplicate")
		return
	}

	if exists {
		existingISO, _ := h.db.GetISOByComposite(normalizedName, req.Version, req.Arch, req.Edition, fileType)
		c.JSON(http.StatusConflict, APIResponse{
			Success: false,
			Error: &APIError{
				Code:    ErrCodeConflict,
				Message: "ISO already exists",
			},
			Data: gin.H{
				"existing": existingISO,
			},
		})
		return
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
	iso.ComputeFields()

	// Save to database
	if err := h.db.CreateISO(iso); err != nil {
		ErrorResponse(c, http.StatusInternalServerError, ErrCodeInternalError, "Failed to create ISO")
		return
	}

	// Queue download
	h.manager.QueueDownload(iso)

	// Return created ISO with 201 status
	SuccessResponseWithMessage(c, http.StatusCreated, iso, "ISO download queued successfully")
}

// DeleteISO deletes an ISO file and database record
func (h *Handlers) DeleteISO(c *gin.Context) {
	id := c.Param("id")

	// Get ISO from database
	iso, err := h.db.GetISO(id)
	if err != nil {
		ErrorResponse(c, http.StatusNotFound, ErrCodeNotFound, "ISO not found")
		return
	}

	// Cancel ongoing download if the ISO is being downloaded
	if iso.Status == models.StatusDownloading || iso.Status == models.StatusVerifying {
		if h.manager.CancelDownload(id) {
			// Wait a moment for the download to be cancelled and cleaned up
			time.Sleep(100 * time.Millisecond)
		}
	}

	// Delete file if it exists
	filePath := filepath.Join(h.isoDir, iso.FilePath)
	if _, err := os.Stat(filePath); err == nil {
		if err := os.Remove(filePath); err != nil {
			ErrorResponse(c, http.StatusInternalServerError, ErrCodeInternalError, "Failed to delete file")
			return
		}
	}

	// Delete checksum file if it exists
	checksumExtensions := []string{".sha256", ".sha512", ".md5"}
	for _, ext := range checksumExtensions {
		checksumFile := filePath + ext
		if _, err := os.Stat(checksumFile); err == nil {
			os.Remove(checksumFile)
		}
	}

	// Delete temp file if it exists
	tmpFile := filepath.Join(h.isoDir, ".tmp", iso.Filename)
	if _, err := os.Stat(tmpFile); err == nil {
		os.Remove(tmpFile)
	}

	// Delete database record
	if err := h.db.DeleteISO(id); err != nil {
		ErrorResponse(c, http.StatusInternalServerError, ErrCodeInternalError, "Failed to delete ISO from database")
		return
	}

	// Return success response with message
	NoContentResponse(c)
}

// RetryISO retries a failed download
func (h *Handlers) RetryISO(c *gin.Context) {
	id := c.Param("id")

	// Get ISO from database
	iso, err := h.db.GetISO(id)
	if err != nil {
		ErrorResponse(c, http.StatusNotFound, ErrCodeNotFound, "ISO not found")
		return
	}

	// Verify status is failed
	if iso.Status != models.StatusFailed {
		ErrorResponse(c, http.StatusBadRequest, ErrCodeInvalidState,
			fmt.Sprintf("Cannot retry ISO with status: %s. Only failed downloads can be retried", iso.Status))
		return
	}

	// Reset status, progress, and error message
	iso.Status = models.StatusPending
	iso.Progress = 0
	iso.ErrorMessage = ""
	iso.CompletedAt = nil

	// Update database
	if err := h.db.UpdateISO(iso); err != nil {
		ErrorResponse(c, http.StatusInternalServerError, ErrCodeInternalError, "Failed to update ISO")
		return
	}

	// Re-queue download
	h.manager.QueueDownload(iso)

	// Return updated ISO
	SuccessResponseWithMessage(c, http.StatusOK, iso, "Download retry queued successfully")
}

// HealthCheck returns server health status
func (h *Handlers) HealthCheck(c *gin.Context) {
	SuccessResponse(c, http.StatusOK, gin.H{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}
