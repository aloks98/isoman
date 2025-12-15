package service

import (
	"time"

	"linux-iso-manager/internal/db"
	"linux-iso-manager/internal/models"
)

// StatsService handles statistics-related business logic.
type StatsService struct {
	db *db.DB
}

// NewStatsService creates a new statistics service.
func NewStatsService(database *db.DB) *StatsService {
	return &StatsService{db: database}
}

// GetStats retrieves aggregated statistics.
func (s *StatsService) GetStats() (*models.Stats, error) {
	return s.db.GetStats()
}

// GetDownloadTrends retrieves download trends.
func (s *StatsService) GetDownloadTrends(period string, days int) (*models.DownloadTrend, error) {
	// Default to 30 days for daily, 12 weeks for weekly
	if days == 0 {
		if period == "weekly" {
			days = 84 // 12 weeks
		} else {
			days = 30
		}
	}
	return s.db.GetDownloadTrends(period, days)
}

// RecordDownload records a download event and increments the counter.
func (s *StatsService) RecordDownload(isoID string) error {
	// Increment the counter
	if err := s.db.IncrementDownloadCount(isoID); err != nil {
		return err
	}

	// Record the event for time-based tracking
	return s.db.RecordDownloadEvent(isoID, time.Now())
}
