package db

import (
	"database/sql"
	"fmt"
	"linux-iso-manager/internal/models"

	_ "modernc.org/sqlite"
)

// DB wraps the SQLite database connection
type DB struct {
	conn *sql.DB
}

// New creates a new database connection and runs migrations
func New(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable WAL mode for better concurrent write performance
	if _, err := conn.Exec("PRAGMA journal_mode=WAL"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	// Set busy timeout to 5 seconds for handling concurrent writes
	if _, err := conn.Exec("PRAGMA busy_timeout=5000"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to set busy timeout: %w", err)
	}

	db := &DB{conn: conn}
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

// migrate creates the isos table if it doesn't exist
func (db *DB) migrate() error {
	query := `
	CREATE TABLE IF NOT EXISTS isos (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		version TEXT NOT NULL,
		arch TEXT NOT NULL,
		edition TEXT NOT NULL DEFAULT '',
		file_type TEXT NOT NULL,
		filename TEXT NOT NULL,
		file_path TEXT NOT NULL,
		download_link TEXT NOT NULL,
		size_bytes INTEGER DEFAULT 0,
		checksum TEXT DEFAULT '',
		checksum_type TEXT DEFAULT '',
		download_url TEXT NOT NULL,
		checksum_url TEXT DEFAULT '',
		status TEXT NOT NULL,
		progress INTEGER DEFAULT 0,
		error_message TEXT DEFAULT '',
		created_at TIMESTAMP NOT NULL,
		completed_at TIMESTAMP,
		UNIQUE(name, version, arch, edition, file_type)
	);
	`
	_, err := db.conn.Exec(query)
	return err
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
	return err
}

// GetISO retrieves a single ISO by ID
func (db *DB) GetISO(id string) (*models.ISO, error) {
	query := `
	SELECT id, name, version, arch, edition, file_type, filename, file_path, download_link,
		   size_bytes, checksum, checksum_type, download_url, checksum_url,
		   status, progress, error_message, created_at, completed_at
	FROM isos WHERE id = ?
	`
	iso := &models.ISO{}
	err := db.conn.QueryRow(query, id).Scan(
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
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("iso not found")
	}
	if err != nil {
		return nil, err
	}
	return iso, nil
}

// ListISOs retrieves all ISOs ordered by created_at DESC
func (db *DB) ListISOs() ([]models.ISO, error) {
	query := `
	SELECT id, name, version, arch, edition, file_type, filename, file_path, download_link,
		   size_bytes, checksum, checksum_type, download_url, checksum_url,
		   status, progress, error_message, created_at, completed_at
	FROM isos
	ORDER BY created_at DESC
	`
	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			fmt.Printf("failed to close rows: %v\n", err)
		}
	}(rows)

	isos := make([]models.ISO, 0)
	for rows.Next() {
		var iso models.ISO
		err := rows.Scan(
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
		isos = append(isos, iso)
	}
	return isos, rows.Err()
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
	return err
}

// UpdateISOStatus updates the status and error message of an ISO
func (db *DB) UpdateISOStatus(id string, status models.ISOStatus, errorMsg string) error {
	query := `UPDATE isos SET status = ?, error_message = ? WHERE id = ?`
	_, err := db.conn.Exec(query, status, errorMsg, id)
	return err
}

// UpdateISOProgress updates the progress of an ISO
func (db *DB) UpdateISOProgress(id string, progress int) error {
	query := `UPDATE isos SET progress = ? WHERE id = ?`
	_, err := db.conn.Exec(query, progress, id)
	return err
}

// UpdateISOSize updates the size of an ISO
func (db *DB) UpdateISOSize(id string, sizeBytes int64) error {
	query := `UPDATE isos SET size_bytes = ? WHERE id = ?`
	_, err := db.conn.Exec(query, sizeBytes, id)
	return err
}

// UpdateISOChecksum updates the checksum of an ISO
func (db *DB) UpdateISOChecksum(id string, checksum string) error {
	query := `UPDATE isos SET checksum = ? WHERE id = ?`
	_, err := db.conn.Exec(query, checksum, id)
	return err
}

// DeleteISO deletes an ISO record from the database
func (db *DB) DeleteISO(id string) error {
	query := `DELETE FROM isos WHERE id = ?`
	_, err := db.conn.Exec(query, id)
	return err
}

// ISOExists checks if an ISO with the given combination already exists
func (db *DB) ISOExists(name, version, arch, edition, fileType string) (bool, error) {
	query := `SELECT COUNT(*) FROM isos WHERE name = ? AND version = ? AND arch = ? AND edition = ? AND file_type = ?`
	var count int
	err := db.conn.QueryRow(query, name, version, arch, edition, fileType).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetISOByComposite retrieves an ISO by its composite key
func (db *DB) GetISOByComposite(name, version, arch, edition, fileType string) (*models.ISO, error) {
	query := `
	SELECT id, name, version, arch, edition, file_type, filename, file_path, download_link,
		   size_bytes, checksum, checksum_type, download_url, checksum_url,
		   status, progress, error_message, created_at, completed_at
	FROM isos WHERE name = ? AND version = ? AND arch = ? AND edition = ? AND file_type = ?
	`
	iso := &models.ISO{}
	err := db.conn.QueryRow(query, name, version, arch, edition, fileType).Scan(
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
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("iso not found")
	}
	if err != nil {
		return nil, err
	}
	return iso, nil
}
