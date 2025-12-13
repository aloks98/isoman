package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIResponse represents a standard API response structure.
type APIResponse struct {
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
	Success bool        `json:"success"`
}

// APIError represents error details in the response.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// SuccessResponse sends a successful response with data.
func SuccessResponse(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, APIResponse{
		Success: true,
		Data:    data,
	})
}

// SuccessResponseWithMessage sends a successful response with data and a message.
func SuccessResponseWithMessage(c *gin.Context, statusCode int, data interface{}, message string) {
	c.JSON(statusCode, APIResponse{
		Success: true,
		Data:    data,
		Message: message,
	})
}

// ErrorResponse sends an error response.
func ErrorResponse(c *gin.Context, statusCode int, code string, message string) {
	c.JSON(statusCode, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
		},
	})
}

// ErrorResponseWithDetails sends an error response with additional details.
func ErrorResponseWithDetails(c *gin.Context, statusCode int, code string, message string, details string) {
	c.JSON(statusCode, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

// Common error codes.
const (
	ErrCodeBadRequest       = "BAD_REQUEST"
	ErrCodeNotFound         = "NOT_FOUND"
	ErrCodeConflict         = "CONFLICT"
	ErrCodeInternalError    = "INTERNAL_ERROR"
	ErrCodeValidationFailed = "VALIDATION_FAILED"
	ErrCodeInvalidState     = "INVALID_STATE"
)

// NoContentResponse sends a 204 No Content response (for DELETE operations).
func NoContentResponse(c *gin.Context) {
	// For DELETE operations, we use 200 OK with success response instead of 204
	// This maintains uniform response structure
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Resource deleted successfully",
	})
}
