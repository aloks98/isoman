package db

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"linux-iso-manager/internal/models"
)

// IncrementDownloadCount increments the download count for an ISO.
func (db *DB) IncrementDownloadCount(id string) error {
	query := `UPDATE isos SET download_count = download_count + 1 WHERE id = ?`
	if _, err := db.conn.Exec(query, id); err != nil {
		return fmt.Errorf("failed to increment download count (id=%s): %w", id, err)
	}
	return nil
}

// RecordDownloadEvent records a download event for time-based tracking.
func (db *DB) RecordDownloadEvent(isoID string, downloadedAt time.Time) error {
	query := `INSERT INTO download_events (iso_id, downloaded_at) VALUES (?, ?)`
	// Format as RFC3339 for consistent SQLite timestamp handling
	_, err := db.conn.Exec(query, isoID, downloadedAt.Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("failed to record download event: %w", err)
	}
	return nil
}

// GetStats retrieves aggregated statistics.
func (db *DB) GetStats() (*models.Stats, error) {
	stats := &models.Stats{
		ISOsByArch:    make(map[string]int64),
		ISOsByEdition: make(map[string]int64),
		ISOsByStatus:  make(map[string]int64),
		TopDownloaded: make([]models.ISODownloadStat, 0),
	}

	// Get total counts
	row := db.conn.QueryRow(`SELECT COUNT(*) FROM isos`)
	if err := row.Scan(&stats.TotalISOs); err != nil {
		return nil, fmt.Errorf("failed to get total ISOs count: %w", err)
	}

	// Get counts by status
	if err := db.getStatsByStatus(stats); err != nil {
		return nil, err
	}

	// Get total storage used (only complete ISOs)
	row = db.conn.QueryRow(`SELECT COALESCE(SUM(size_bytes), 0) FROM isos WHERE status = 'complete'`)
	if err := row.Scan(&stats.TotalSizeBytes); err != nil {
		return nil, fmt.Errorf("failed to get total storage: %w", err)
	}

	// Get total downloads
	row = db.conn.QueryRow(`SELECT COALESCE(SUM(download_count), 0) FROM isos`)
	if err := row.Scan(&stats.TotalDownloads); err != nil {
		return nil, fmt.Errorf("failed to get total downloads: %w", err)
	}

	// Calculate bandwidth saved: Σ (download_count - 1) × size_bytes for downloads > 1
	row = db.conn.QueryRow(`SELECT COALESCE(SUM((download_count - 1) * size_bytes), 0) FROM isos WHERE download_count > 1 AND status = 'complete'`)
	if err := row.Scan(&stats.BandwidthSaved); err != nil {
		return nil, fmt.Errorf("failed to calculate bandwidth saved: %w", err)
	}

	// Get ISOs by arch
	if err := db.getStatsByArch(stats); err != nil {
		return nil, err
	}

	// Get ISOs by edition (only non-empty editions)
	if err := db.getStatsByEdition(stats); err != nil {
		return nil, err
	}

	// Get top 10 downloaded ISOs
	if err := db.getTopDownloaded(stats); err != nil {
		return nil, err
	}

	return stats, nil
}

func (db *DB) getStatsByStatus(stats *models.Stats) error {
	rows, err := db.conn.Query(`SELECT status, COUNT(*) FROM isos GROUP BY status`) //nolint:sqlclosecheck
	if err != nil {
		return fmt.Errorf("failed to get ISOs by status: %w", err)
	}
	defer closeRows(rows)

	for rows.Next() {
		var status string
		var count int64
		if err := rows.Scan(&status, &count); err != nil {
			return err
		}
		stats.ISOsByStatus[status] = count
		switch status {
		case "complete":
			stats.CompletedISOs = count
		case "failed":
			stats.FailedISOs = count
		case "pending", "downloading", "verifying":
			stats.PendingISOs += count
		}
	}
	return rows.Err()
}

func (db *DB) getStatsByArch(stats *models.Stats) error {
	rows, err := db.conn.Query(`SELECT arch, COUNT(*) FROM isos GROUP BY arch`) //nolint:sqlclosecheck
	if err != nil {
		return fmt.Errorf("failed to get ISOs by arch: %w", err)
	}
	defer closeRows(rows)

	for rows.Next() {
		var arch string
		var count int64
		if err := rows.Scan(&arch, &count); err != nil {
			return err
		}
		stats.ISOsByArch[arch] = count
	}
	return rows.Err()
}

func (db *DB) getStatsByEdition(stats *models.Stats) error {
	rows, err := db.conn.Query(`SELECT edition, COUNT(*) FROM isos WHERE edition != '' GROUP BY edition`) //nolint:sqlclosecheck
	if err != nil {
		return fmt.Errorf("failed to get ISOs by edition: %w", err)
	}
	defer closeRows(rows)

	for rows.Next() {
		var edition string
		var count int64
		if err := rows.Scan(&edition, &count); err != nil {
			return err
		}
		stats.ISOsByEdition[edition] = count
	}
	return rows.Err()
}

func (db *DB) getTopDownloaded(stats *models.Stats) error {
	//nolint:sqlclosecheck
	rows, err := db.conn.Query(`
		SELECT id, name, version, arch, download_count, size_bytes
		FROM isos
		WHERE status = 'complete' AND download_count > 0
		ORDER BY download_count DESC
		LIMIT 10
	`)
	if err != nil {
		return fmt.Errorf("failed to get top downloaded ISOs: %w", err)
	}
	defer closeRows(rows)

	for rows.Next() {
		var stat models.ISODownloadStat
		if err := rows.Scan(&stat.ID, &stat.Name, &stat.Version, &stat.Arch, &stat.DownloadCount, &stat.SizeBytes); err != nil {
			return err
		}
		stats.TopDownloaded = append(stats.TopDownloaded, stat)
	}
	return rows.Err()
}

// GetDownloadTrends retrieves download trends for a period.
func (db *DB) GetDownloadTrends(period string, days int) (*models.DownloadTrend, error) {
	trend := &models.DownloadTrend{
		Period: period,
		Data:   make([]models.TrendDataPoint, 0),
	}

	var dateFormat string
	if period == "weekly" {
		dateFormat = "%Y-W%W" // ISO week format
	} else {
		dateFormat = "%Y-%m-%d" // Daily format
	}

	startDate := time.Now().AddDate(0, 0, -days).Format(time.RFC3339)

	query := fmt.Sprintf(`
		SELECT strftime('%s', downloaded_at) as period, COUNT(*) as count
		FROM download_events
		WHERE downloaded_at >= ? AND downloaded_at IS NOT NULL
		GROUP BY period
		HAVING period IS NOT NULL
		ORDER BY period ASC
	`, dateFormat)

	rows, err := db.conn.Query(query, startDate) //nolint:sqlclosecheck
	if err != nil {
		return nil, fmt.Errorf("failed to get download trends: %w", err)
	}
	defer closeRows(rows)

	for rows.Next() {
		var point models.TrendDataPoint
		if err := rows.Scan(&point.Date, &point.Count); err != nil {
			return nil, err
		}
		trend.Data = append(trend.Data, point)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating trend rows: %w", err)
	}

	return trend, nil
}

// GetISOByFilePath retrieves an ISO by its file path (for download tracking).
func (db *DB) GetISOByFilePath(filePath string) (*models.ISO, error) {
	query := fmt.Sprintf("SELECT %s FROM isos WHERE file_path = ?", isoSelectFields)
	row := db.conn.QueryRow(query, filePath)

	iso, err := scanISO(row)
	if err == sql.ErrNoRows {
		return nil, nil // Not found, but not an error
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get ISO by file path: %w", err)
	}
	return iso, nil
}

// closeRows is a helper to safely close rows with error logging.
func closeRows(rows *sql.Rows) {
	if err := rows.Close(); err != nil {
		slog.Warn("failed to close rows", slog.Any("error", err))
	}
}
