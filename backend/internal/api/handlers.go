package api

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"linux-iso-manager/internal/constants"
	"linux-iso-manager/internal/fileutil"
	"linux-iso-manager/internal/models"
	"linux-iso-manager/internal/pathutil"
	"linux-iso-manager/internal/service"
	"linux-iso-manager/internal/validation"

	"github.com/gin-gonic/gin"
)

// Handlers holds references to service layer and storage directory.
type Handlers struct {
	isoService *service.ISOService
	isoDir     string
}

// NewHandlers creates a new Handlers instance.
func NewHandlers(isoService *service.ISOService, isoDir string) *Handlers {
	return &Handlers{
		isoService: isoService,
		isoDir:     isoDir,
	}
}

// ListISOs returns all ISOs ordered by created_at DESC.
func (h *Handlers) ListISOs(c *gin.Context) {
	isos, err := h.isoService.ListISOs()
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, ErrCodeInternalError, "Failed to list ISOs")
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{
		"isos": isos,
	})
}

// GetISO returns a single ISO by ID.
func (h *Handlers) GetISO(c *gin.Context) {
	id := c.Param("id")

	iso, err := h.isoService.GetISO(id)
	if err != nil {
		ErrorResponse(c, http.StatusNotFound, ErrCodeNotFound, "ISO not found")
		return
	}

	SuccessResponse(c, http.StatusOK, iso)
}

// CreateISO creates a new ISO download.
func (h *Handlers) CreateISO(c *gin.Context) {
	var req validation.ISOCreateRequest

	// Parse JSON request
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponseWithDetails(c, http.StatusBadRequest, ErrCodeValidationFailed, "Invalid request body", err.Error())
		return
	}

	// Validate request
	if err := validation.ValidateISOCreateRequest(&req); err != nil {
		ErrorResponseWithDetails(c, http.StatusBadRequest, ErrCodeValidationFailed, "Validation failed", err.Error())
		return
	}

	// Call service layer
	iso, err := h.isoService.CreateISO(service.CreateISORequest{
		Name:         req.Name,
		Version:      req.Version,
		Arch:         req.Arch,
		Edition:      req.Edition,
		DownloadURL:  req.DownloadURL,
		ChecksumURL:  req.ChecksumURL,
		ChecksumType: req.ChecksumType,
	})
	if err != nil {
		// Check for specific error types
		var existsErr *service.ISOAlreadyExistsError
		if errors.As(err, &existsErr) {
			c.JSON(http.StatusConflict, APIResponse{
				Success: false,
				Error: &APIError{
					Code:    ErrCodeConflict,
					Message: "ISO already exists",
				},
				Data: gin.H{
					"existing": existsErr.ExistingISO,
				},
			})
			return
		}

		// Check if it's a validation error (invalid file type, etc.)
		errMsg := err.Error()
		if strings.Contains(errMsg, "unsupported file type") || strings.Contains(errMsg, "invalid file type") {
			ErrorResponse(c, http.StatusBadRequest, ErrCodeValidationFailed, err.Error())
			return
		}

		ErrorResponse(c, http.StatusInternalServerError, ErrCodeInternalError, "Failed to create ISO")
		return
	}

	// Return created ISO with 201 status
	SuccessResponseWithMessage(c, http.StatusCreated, iso, "ISO download queued successfully")
}

// DeleteISO deletes an ISO file and database record.
func (h *Handlers) DeleteISO(c *gin.Context) {
	id := c.Param("id")

	// Get ISO from database before deleting (for file cleanup)
	iso, err := h.isoService.GetISO(id)
	if err != nil {
		ErrorResponse(c, http.StatusNotFound, ErrCodeNotFound, "ISO not found")
		return
	}

	// Call service layer to delete ISO
	if err := h.isoService.DeleteISO(id); err != nil {
		ErrorResponse(c, http.StatusInternalServerError, ErrCodeInternalError, "Failed to delete ISO")
		return
	}

	// Clean up files (best effort - files can be manually cleaned up later if needed)
	filePath := pathutil.ConstructISOPath(h.isoDir, iso.FilePath)
	tmpFile := pathutil.ConstructTempPath(h.isoDir, iso.Filename)

	// Delete main ISO file and checksum files
	fileutil.DeleteFileSilently(filePath)
	for _, ext := range constants.ChecksumExtensions {
		fileutil.DeleteFileSilently(filePath + ext)
	}

	// Delete temp file if it exists
	fileutil.DeleteFileSilently(tmpFile)

	// Return success response
	NoContentResponse(c)
}

// RetryISO retries a failed download.
func (h *Handlers) RetryISO(c *gin.Context) {
	id := c.Param("id")

	// Call service layer to retry ISO
	iso, err := h.isoService.RetryISO(id)
	if err != nil {
		// Check for specific error types
		var invalidStateErr *service.InvalidStateError
		if errors.As(err, &invalidStateErr) {
			ErrorResponse(c, http.StatusBadRequest, ErrCodeInvalidState, invalidStateErr.Error())
			return
		}

		ErrorResponse(c, http.StatusNotFound, ErrCodeNotFound, "ISO not found")
		return
	}

	// Return updated ISO
	SuccessResponseWithMessage(c, http.StatusOK, iso, "Download retry queued successfully")
}

// UpdateISO updates an existing ISO.
func (h *Handlers) UpdateISO(c *gin.Context) {
	id := c.Param("id")

	// Parse request body
	var req models.UpdateISORequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, ErrCodeBadRequest, "Invalid request body: "+err.Error())
		return
	}

	// Call service layer to update ISO
	iso, err := h.isoService.UpdateISO(id, req)
	if err != nil {
		// Check for specific error types
		var invalidStateErr *service.InvalidStateError
		if errors.As(err, &invalidStateErr) {
			ErrorResponse(c, http.StatusBadRequest, ErrCodeInvalidState, invalidStateErr.Error())
			return
		}

		var existsErr *service.ISOAlreadyExistsError
		if errors.As(err, &existsErr) {
			c.JSON(http.StatusConflict, APIResponse{
				Success: false,
				Error: &APIError{
					Code:    ErrCodeConflict,
					Message: "ISO already exists",
				},
				Data: gin.H{
					"existing": existsErr.ExistingISO,
				},
			})
			return
		}

		// Check if it's a not found error
		if strings.Contains(err.Error(), "not found") {
			ErrorResponse(c, http.StatusNotFound, ErrCodeNotFound, "ISO not found")
			return
		}

		ErrorResponse(c, http.StatusInternalServerError, ErrCodeInternalError, "Failed to update ISO")
		return
	}

	// Return updated ISO
	SuccessResponseWithMessage(c, http.StatusOK, iso, "ISO updated successfully")
}

// HealthCheck returns server health status.
func (h *Handlers) HealthCheck(c *gin.Context) {
	SuccessResponse(c, http.StatusOK, gin.H{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}
