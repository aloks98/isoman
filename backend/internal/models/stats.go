package models

import "time"

// Stats represents aggregated statistics for the dashboard.
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

// DownloadTrend represents download trends over time.
type DownloadTrend struct {
	Period string           `json:"period"`
	Data   []TrendDataPoint `json:"data"`
}

// TrendDataPoint represents a single data point in a trend.
type TrendDataPoint struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// DownloadEvent represents a single download event for tracking.
type DownloadEvent struct {
	ID           int64     `json:"id"`
	ISOID        string    `json:"iso_id"`
	DownloadedAt time.Time `json:"downloaded_at"`
}
