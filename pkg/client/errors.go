package client

import "fmt"

// APIError represents an error response from the ISOMan API.
type APIError struct {
	// StatusCode is the HTTP status code.
	StatusCode int
	// Code is the application error code (e.g. "NOT_FOUND", "CONFLICT").
	Code string
	// Message is a human-readable error message.
	Message string
	// Details contains optional additional error details.
	Details string
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("isoman: %s: %s (%s)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("isoman: %s: %s", e.Code, e.Message)
}

// IsNotFound reports whether err is an ISOMan API 404 Not Found error.
func IsNotFound(err error) bool {
	if e, ok := err.(*APIError); ok {
		return e.StatusCode == 404
	}
	return false
}

// IsConflict reports whether err is an ISOMan API 409 Conflict error.
func IsConflict(err error) bool {
	if e, ok := err.(*APIError); ok {
		return e.StatusCode == 409
	}
	return false
}
