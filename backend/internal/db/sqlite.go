package db

import (
	"database/sql"
	"fmt"
	"linux-iso-manager/internal/config"
	"linux-iso-manager/internal/models"
	"log/slog"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "modernc.org/sqlite"
)

// SQL constants for ISO queries
const (
	isoSelectFields = `id, name, version, arch, edition, file_type, filename, file_path, download_link,
		size_bytes, checksum, checksum_type, download_url, checksum_url,
		status, progress, error_message, created_at, completed_at`

	isoInsertFields = `id, name, version, arch, edition, file_type, filename, file_path, download_link,
		size_bytes, checksum, checksum_type, download_url, checksum_url,
		status, progress, error_message, created_at, completed_at`

	isoUpdateFields = `name = ?, version = ?, arch = ?, edition = ?, file_type = ?,
		filename = ?, file_path = ?, download_link = ?,
		size_bytes = ?, checksum = ?, checksum_type = ?,
		download_url = ?, checksum_url = ?, status = ?, progress = ?,
		error_message = ?, completed_at = ?`
)

// DB wraps the SQLite database connection
type DB struct {
	conn *sql.DB
	cfg  *config.DatabaseConfig
}

// scanISO scans a single ISO from a sql.Row or sql.Rows
type scanner interface {
	Scan(dest ...interface{}) error
}

// scanISO scans an ISO from a database row
func scanISO(s scanner) (*models.ISO, error) {
	iso := &models.ISO{}
	err := s.Scan(
		&iso.ID,
		&iso.Name,
		&iso.Version,
		&iso.Arch,
		&iso.Edition,
		&iso.FileType,
		&iso.Filename,
		&iso.FilePath,
		&iso.DownloadLink,
		&iso.SizeBytes,
		&iso.Checksum,
		&iso.ChecksumType,
		&iso.DownloadURL,
		&iso.ChecksumURL,
		&iso.Status,
		&iso.Progress,
		&iso.ErrorMessage,
		&iso.CreatedAt,
		&iso.CompletedAt,
	)
	if err != nil {
		return nil, err
	}
	return iso, nil
}

// New creates a new database connection and runs migrations
func New(dbPath string, cfg *config.DatabaseConfig) (*DB, error) {
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable journal mode (configurable, default: WAL)
	journalMode := cfg.JournalMode
	if _, err := conn.Exec(fmt.Sprintf("PRAGMA journal_mode=%s", journalMode)); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to set journal mode: %w", err)
	}

	// Set busy timeout (configurable, default: 5000ms)
	busyTimeoutMs := int(cfg.BusyTimeout.Milliseconds())
	if _, err := conn.Exec(fmt.Sprintf("PRAGMA busy_timeout=%d", busyTimeoutMs)); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to set busy timeout: %w", err)
	}

	// Configure connection pool
	conn.SetMaxOpenConns(cfg.MaxOpenConns)
	conn.SetMaxIdleConns(cfg.MaxIdleConns)
	conn.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	conn.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	db := &DB{
		conn: conn,
		cfg:  cfg,
	}
	if err := db.migrate(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// migrate runs database migrations using golang-migrate
func (db *DB) migrate() error {
	// Create a driver instance for golang-migrate
	driver, err := sqlite.WithInstance(db.conn, &sqlite.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	// Find migrations directory - try multiple paths for production and tests
	migrationsPath := findMigrationsPath()
	sourceURL := fmt.Sprintf("file://%s", migrationsPath)

	// Create migrate instance with file source
	m, err := migrate.NewWithDatabaseInstance(
		sourceURL,
		"sqlite", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	// Run migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	slog.Info("Database migrations completed successfully")
	return nil
}

// findMigrationsPath finds the migrations directory
// Checks multiple paths to work in both production and test environments
func findMigrationsPath() string {
	paths := []string{
		"./migrations",        // Production: binary in project root
		"../../migrations",    // Tests: internal/db/sqlite_test.go
		"../../../migrations", // Tests: internal/download/*_test.go
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Default fallback
	return "./migrations"
}

// CreateISO inserts a new ISO record into the database
func (db *DB) CreateISO(iso *models.ISO) error {
	query := `
	INSERT INTO isos (
		id, name, version, arch, edition, file_type, filename, file_path, download_link,
		size_bytes, checksum, checksum_type, download_url, checksum_url,
		status, progress, error_message, created_at, completed_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := db.conn.Exec(
		query,
		iso.ID,
		iso.Name,
		iso.Version,
		iso.Arch,
		iso.Edition,
		iso.FileType,
		iso.Filename,
		iso.FilePath,
		iso.DownloadLink,
		iso.SizeBytes,
		iso.Checksum,
		iso.ChecksumType,
		iso.DownloadURL,
		iso.ChecksumURL,
		iso.Status,
		iso.Progress,
		iso.ErrorMessage,
		iso.CreatedAt,
		iso.CompletedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert ISO record (id=%s): %w", iso.ID, err)
	}
	return nil
}

// GetISO retrieves a single ISO by ID
func (db *DB) GetISO(id string) (*models.ISO, error) {
	query := fmt.Sprintf("SELECT %s FROM isos WHERE id = ?", isoSelectFields)
	row := db.conn.QueryRow(query, id)

	iso, err := scanISO(row)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("ISO not found (id=%s)", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan ISO record (id=%s): %w", id, err)
	}
	return iso, nil
}

// ListISOs retrieves all ISOs ordered by created_at DESC
func (db *DB) ListISOs() ([]models.ISO, error) {
	query := fmt.Sprintf("SELECT %s FROM isos ORDER BY created_at DESC", isoSelectFields)
	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query ISO list: %w", err)
	}
	defer func(rows *sql.Rows) {
		if err := rows.Close(); err != nil {
			slog.Warn("failed to close rows", slog.Any("error", err))
		}
	}(rows)

	isos := make([]models.ISO, 0)
	for rows.Next() {
		iso, err := scanISO(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ISO record: %w", err)
		}
		isos = append(isos, *iso)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating ISO rows: %w", err)
	}
	return isos, nil
}

// UpdateISO updates an existing ISO record
func (db *DB) UpdateISO(iso *models.ISO) error {
	query := `
	UPDATE isos SET
		name = ?, version = ?, arch = ?, edition = ?, file_type = ?,
		filename = ?, file_path = ?, download_link = ?,
		size_bytes = ?, checksum = ?, checksum_type = ?,
		download_url = ?, checksum_url = ?, status = ?, progress = ?,
		error_message = ?, completed_at = ?
	WHERE id = ?
	`
	_, err := db.conn.Exec(
		query,
		iso.Name,
		iso.Version,
		iso.Arch,
		iso.Edition,
		iso.FileType,
		iso.Filename,
		iso.FilePath,
		iso.DownloadLink,
		iso.SizeBytes,
		iso.Checksum,
		iso.ChecksumType,
		iso.DownloadURL,
		iso.ChecksumURL,
		iso.Status,
		iso.Progress,
		iso.ErrorMessage,
		iso.CompletedAt,
		iso.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update ISO record (id=%s): %w", iso.ID, err)
	}
	return nil
}

// UpdateISOStatus updates the status and error message of an ISO
func (db *DB) UpdateISOStatus(id string, status models.ISOStatus, errorMsg string) error {
	query := `UPDATE isos SET status = ?, error_message = ? WHERE id = ?`
	if _, err := db.conn.Exec(query, status, errorMsg, id); err != nil {
		return fmt.Errorf("failed to update ISO status (id=%s, status=%s): %w", id, status, err)
	}
	return nil
}

// UpdateISOProgress updates the progress of an ISO
func (db *DB) UpdateISOProgress(id string, progress int) error {
	query := `UPDATE isos SET progress = ? WHERE id = ?`
	if _, err := db.conn.Exec(query, progress, id); err != nil {
		return fmt.Errorf("failed to update ISO progress (id=%s, progress=%d): %w", id, progress, err)
	}
	return nil
}

// UpdateISOSize updates the size of an ISO
func (db *DB) UpdateISOSize(id string, sizeBytes int64) error {
	query := `UPDATE isos SET size_bytes = ? WHERE id = ?`
	if _, err := db.conn.Exec(query, sizeBytes, id); err != nil {
		return fmt.Errorf("failed to update ISO size (id=%s): %w", id, err)
	}
	return nil
}

// UpdateISOChecksum updates the checksum of an ISO
func (db *DB) UpdateISOChecksum(id string, checksum string) error {
	query := `UPDATE isos SET checksum = ? WHERE id = ?`
	if _, err := db.conn.Exec(query, checksum, id); err != nil {
		return fmt.Errorf("failed to update ISO checksum (id=%s): %w", id, err)
	}
	return nil
}

// DeleteISO deletes an ISO record from the database
func (db *DB) DeleteISO(id string) error {
	query := `DELETE FROM isos WHERE id = ?`
	if _, err := db.conn.Exec(query, id); err != nil {
		return fmt.Errorf("failed to delete ISO record (id=%s): %w", id, err)
	}
	return nil
}

// ISOExists checks if an ISO with the given combination already exists
func (db *DB) ISOExists(name, version, arch, edition, fileType string) (bool, error) {
	query := `SELECT COUNT(*) FROM isos WHERE name = ? AND version = ? AND arch = ? AND edition = ? AND file_type = ?`
	var count int
	err := db.conn.QueryRow(query, name, version, arch, edition, fileType).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check ISO existence (name=%s, version=%s, arch=%s): %w", name, version, arch, err)
	}
	return count > 0, nil
}

// GetISOByComposite retrieves an ISO by its composite key
func (db *DB) GetISOByComposite(name, version, arch, edition, fileType string) (*models.ISO, error) {
	query := fmt.Sprintf("SELECT %s FROM isos WHERE name = ? AND version = ? AND arch = ? AND edition = ? AND file_type = ?", isoSelectFields)
	row := db.conn.QueryRow(query, name, version, arch, edition, fileType)

	iso, err := scanISO(row)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("ISO not found (name=%s, version=%s, arch=%s, edition=%s, fileType=%s)", name, version, arch, edition, fileType)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan ISO record (name=%s, version=%s, arch=%s): %w", name, version, arch, err)
	}
	return iso, nil
}
