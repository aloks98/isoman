// Package client provides a Go HTTP client for the ISOMan API.
package client

import "time"

// ISOStatus represents the status of an ISO download.
type ISOStatus string

const (
	StatusPending     ISOStatus = "pending"
	StatusDownloading ISOStatus = "downloading"
	StatusVerifying   ISOStatus = "verifying"
	StatusComplete    ISOStatus = "complete"
	StatusFailed      ISOStatus = "failed"
)

// ISO represents an ISO file managed by ISOMan.
type ISO struct {
	CreatedAt     time.Time  `json:"created_at"`
	CompletedAt   *time.Time `json:"completed_at"`
	DownloadLink  string     `json:"download_link"`
	ChecksumType  string     `json:"checksum_type"`
	Edition       string     `json:"edition"`
	FileType      string     `json:"file_type"`
	Filename      string     `json:"filename"`
	FilePath      string     `json:"file_path"`
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	Checksum      string     `json:"checksum"`
	Arch          string     `json:"arch"`
	DownloadURL   string     `json:"download_url"`
	ChecksumURL   string     `json:"checksum_url"`
	Status        ISOStatus  `json:"status"`
	Version       string     `json:"version"`
	ErrorMessage  string     `json:"error_message"`
	Progress      int        `json:"progress"`
	SizeBytes     int64      `json:"size_bytes"`
	DownloadCount int64      `json:"download_count"`
}

// CreateISORequest is the request body for creating a new ISO download.
type CreateISORequest struct {
	// Name is the display name (will be normalized, e.g. "Alpine Linux" -> "alpine-linux").
	Name string `json:"name"`
	// Version is the version string (e.g. "3.19.1", "24.04").
	Version string `json:"version"`
	// Arch is the architecture (e.g. "x86_64", "aarch64").
	Arch string `json:"arch"`
	// Edition is an optional variant (e.g. "minimal", "desktop", "server").
	Edition string `json:"edition,omitempty"`
	// DownloadURL is the URL to download the file from.
	DownloadURL string `json:"download_url"`
	// ChecksumURL is an optional URL to a checksum file.
	ChecksumURL string `json:"checksum_url,omitempty"`
	// ChecksumType is the hash type: "sha256", "sha512", or "md5" (default "sha256").
	ChecksumType string `json:"checksum_type,omitempty"`
}

// UpdateISORequest is the request body for updating an ISO.
// All fields are optional — only non-nil fields are applied.
type UpdateISORequest struct {
	Name         *string `json:"name,omitempty"`
	Version      *string `json:"version,omitempty"`
	Arch         *string `json:"arch,omitempty"`
	Edition      *string `json:"edition,omitempty"`
	DownloadURL  *string `json:"download_url,omitempty"`
	ChecksumURL  *string `json:"checksum_url,omitempty"`
	ChecksumType *string `json:"checksum_type,omitempty"`
}

// Stats represents aggregated statistics from the ISOMan dashboard.
type Stats struct {
	TotalISOs      int64             `json:"total_isos"`
	CompletedISOs  int64             `json:"completed_isos"`
	FailedISOs     int64             `json:"failed_isos"`
	PendingISOs    int64             `json:"pending_isos"`
	TotalSizeBytes int64             `json:"total_size_bytes"`
	TotalDownloads int64             `json:"total_downloads"`
	BandwidthSaved int64             `json:"bandwidth_saved"`
	ISOsByArch     map[string]int64  `json:"isos_by_arch"`
	ISOsByEdition  map[string]int64  `json:"isos_by_edition"`
	ISOsByStatus   map[string]int64  `json:"isos_by_status"`
	TopDownloaded  []ISODownloadStat `json:"top_downloaded"`
}

// ISODownloadStat represents download statistics for a single ISO.
type ISODownloadStat struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Version       string `json:"version"`
	Arch          string `json:"arch"`
	DownloadCount int64  `json:"download_count"`
	SizeBytes     int64  `json:"size_bytes"`
}

// DownloadTrends represents download trend data over a time period.
type DownloadTrends struct {
	Period string           `json:"period"`
	Data   []TrendDataPoint `json:"data"`
}

// TrendDataPoint represents a single data point in a download trend.
type TrendDataPoint struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// Pagination contains pagination metadata from list responses.
type Pagination struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// ListISOsResponse is the unwrapped response from ListISOs,
// containing both the ISO list and pagination metadata.
type ListISOsResponse struct {
	ISOs       []ISO      `json:"isos"`
	Pagination Pagination `json:"pagination"`
}

// ListISOsOptions configures the ListISOs request.
type ListISOsOptions struct {
	// Page number (1-based). Default: 1.
	Page int
	// PageSize is the number of results per page. Default: 10.
	PageSize int
	// SortBy is the field to sort by (e.g. "created_at", "name"). Default: "created_at".
	SortBy string
	// SortDir is the sort direction: "asc" or "desc". Default: "desc".
	SortDir string
}

// DownloadTrendsOptions configures the GetDownloadTrends request.
type DownloadTrendsOptions struct {
	// Period is "daily" or "weekly". Default: "daily".
	Period string
	// Days is the number of days to look back (1-365). Default: 30.
	Days int
}
