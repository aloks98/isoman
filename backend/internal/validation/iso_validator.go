package validation

import (
	"fmt"
	"linux-iso-manager/internal/constants"
	"net/url"
	"strings"
)

// ISOCreateRequest validation
type ISOCreateRequest struct {
	Name         string `json:"name"`
	Version      string `json:"version"`
	Arch         string `json:"arch"`
	Edition      string `json:"edition"`
	DownloadURL  string `json:"download_url"`
	ChecksumURL  string `json:"checksum_url"`
	ChecksumType string `json:"checksum_type"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors struct {
	Errors []ValidationError
}

func (e *ValidationErrors) Error() string {
	if len(e.Errors) == 0 {
		return "validation failed"
	}
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}
	var messages []string
	for _, err := range e.Errors {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}

func (e *ValidationErrors) Add(field, message string) {
	e.Errors = append(e.Errors, ValidationError{Field: field, Message: message})
}

func (e *ValidationErrors) HasErrors() bool {
	return len(e.Errors) > 0
}

// ValidateISOCreateRequest validates an ISO create request
func ValidateISOCreateRequest(req *ISOCreateRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	errs := &ValidationErrors{}

	// Validate name
	if strings.TrimSpace(req.Name) == "" {
		errs.Add("name", "name is required")
	} else if len(req.Name) > 100 {
		errs.Add("name", "name must be 100 characters or less")
	}

	// Validate version
	if strings.TrimSpace(req.Version) == "" {
		errs.Add("version", "version is required")
	} else if len(req.Version) > 50 {
		errs.Add("version", "version must be 50 characters or less")
	}

	// Validate arch
	if strings.TrimSpace(req.Arch) == "" {
		errs.Add("arch", "arch is required")
	} else if len(req.Arch) > 20 {
		errs.Add("arch", "arch must be 20 characters or less")
	}

	// Validate edition (optional)
	if len(req.Edition) > 50 {
		errs.Add("edition", "edition must be 50 characters or less")
	}

	// Validate download URL
	if strings.TrimSpace(req.DownloadURL) == "" {
		errs.Add("download_url", "download_url is required")
	} else if len(req.DownloadURL) > 2048 {
		errs.Add("download_url", "download_url must be 2048 characters or less")
	} else if !isValidHTTPURL(req.DownloadURL) {
		errs.Add("download_url", "download_url must be a valid HTTP or HTTPS URL")
	}

	// Validate checksum URL (optional)
	if req.ChecksumURL != "" {
		if len(req.ChecksumURL) > 2048 {
			errs.Add("checksum_url", "checksum_url must be 2048 characters or less")
		} else if !isValidHTTPURL(req.ChecksumURL) {
			errs.Add("checksum_url", "checksum_url must be a valid HTTP or HTTPS URL")
		}
	}

	// Validate checksum type (optional)
	if req.ChecksumType != "" && !constants.IsValidChecksumType(req.ChecksumType) {
		errs.Add("checksum_type", fmt.Sprintf("checksum_type must be one of: %v", constants.ChecksumTypes))
	}

	if errs.HasErrors() {
		return errs
	}

	return nil
}

// isValidHTTPURL checks if a string is a valid HTTP or HTTPS URL
func isValidHTTPURL(urlStr string) bool {
	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}
	scheme := strings.ToLower(u.Scheme)
	return (scheme == "http" || scheme == "https") && u.Host != ""
}
