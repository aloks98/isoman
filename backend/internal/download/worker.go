package download

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"linux-iso-manager/internal/db"
	"linux-iso-manager/internal/fileutil"
	"linux-iso-manager/internal/httputil"
	"linux-iso-manager/internal/models"
	"linux-iso-manager/internal/pathutil"
)

// ProgressCallback is called when download progress updates.
type ProgressCallback func(isoID string, progress int, status models.ISOStatus)

// Worker handles the download and verification of a single ISO.
type Worker struct {
	db               *db.DB
	progressCallback ProgressCallback
	isoDir           string
	tmpDir           string
}

// NewWorker creates a new download worker.
func NewWorker(database *db.DB, isoDir string, callback ProgressCallback) *Worker {
	tmpDir := filepath.Join(isoDir, ".tmp")
	return &Worker{
		db:               database,
		isoDir:           isoDir,
		tmpDir:           tmpDir,
		progressCallback: callback,
	}
}

// Process downloads and verifies an ISO.
func (w *Worker) Process(ctx context.Context, iso *models.ISO) error {
	// Ensure tmp directory exists
	if err := fileutil.EnsureDirectory(w.tmpDir); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Use the computed FilePath for nested directory structure
	// FilePath example: "alpine/3.19.1/x86_64/alpine-3.19.1-x86_64.iso"
	tmpFile := pathutil.ConstructTempPath(w.isoDir, iso.Filename)
	finalFile := pathutil.ConstructISOPath(w.isoDir, iso.FilePath)

	// Create the nested directory structure for the final file
	if err := fileutil.EnsureParentDirectory(finalFile); err != nil {
		return fmt.Errorf("failed to create final directory: %w", err)
	}

	// Clean up temp file on error
	defer func() {
		fileutil.DeleteFileSilently(tmpFile)
	}()

	// Update status to downloading
	w.updateStatus(iso.ID, models.StatusDownloading, 0, "")

	// Download the file
	if err := w.download(ctx, iso, tmpFile); err != nil {
		// Check if it was canceled
		if ctx.Err() == context.Canceled {
			w.updateStatus(iso.ID, models.StatusFailed, 0, "Download canceled")
			return fmt.Errorf("download canceled: %w", ctx.Err())
		}
		w.updateStatus(iso.ID, models.StatusFailed, 0, err.Error())
		return err
	}

	// Verify checksum if provided
	if iso.ChecksumURL != "" {
		w.updateStatus(iso.ID, models.StatusVerifying, 100, "")

		if err := w.verifyChecksum(iso, tmpFile); err != nil {
			w.updateStatus(iso.ID, models.StatusFailed, 100, err.Error())
			return err
		}
	}

	// Move temp file to final location
	if err := os.Rename(tmpFile, finalFile); err != nil {
		errMsg := fmt.Sprintf("failed to move file to final location: %v", err)
		w.updateStatus(iso.ID, models.StatusFailed, 100, errMsg)
		return fmt.Errorf("failed to move file to final location: %w", err)
	}

	// Download and save checksum file alongside ISO (after file is moved)
	if iso.ChecksumURL != "" {
		checksumFile := pathutil.ConstructChecksumPath(finalFile, iso.ChecksumType)
		if err := w.downloadChecksumFile(iso.ChecksumURL, checksumFile); err != nil {
			slog.Warn("failed to save checksum file",
				slog.String("iso_id", iso.ID),
				slog.Any("error", err),
			)
			// Don't fail the download if checksum file save fails
		}
	}

	// Mark as complete
	w.updateStatus(iso.ID, models.StatusComplete, 100, "")
	now := time.Now()
	iso.CompletedAt = &now
	iso.Status = models.StatusComplete
	iso.Progress = 100
	iso.ErrorMessage = ""

	// Update database to mark as complete (with retry for database busy errors)
	maxRetries := 5
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		lastErr = w.db.UpdateISO(iso)
		if lastErr == nil {
			break // Success
		}

		if i < maxRetries-1 {
			time.Sleep(100 * time.Millisecond)
		}
	}

	if lastErr != nil {
		slog.Error("failed to update ISO to complete status",
			slog.String("iso_id", iso.ID),
			slog.Int("retries", maxRetries),
			slog.Any("error", lastErr),
		)
		// Don't return error since download itself succeeded
	}

	return nil
}

// download downloads the ISO file with progress tracking.
func (w *Worker) download(ctx context.Context, iso *models.ISO, destPath string) error {
	// Use httputil to download with progress tracking
	lastProgress := -1
	lastUpdate := time.Now()

	err := httputil.DownloadFileWithProgress(ctx, iso.DownloadURL, destPath, 32*1024, func(downloaded, total int64) {
		// Update database with total size on first callback
		if iso.SizeBytes == 0 && total > 0 {
			w.db.UpdateISOSize(iso.ID, total)
			iso.SizeBytes = total
		}

		// Calculate progress
		var progress int
		if total > 0 {
			progress = int((downloaded * 100) / total)
		}

		// Update progress every 1% or every second
		now := time.Now()
		if progress != lastProgress && (progress-lastProgress >= 1 || now.Sub(lastUpdate) >= time.Second) {
			w.updateStatus(iso.ID, models.StatusDownloading, progress, "")
			lastProgress = progress
			lastUpdate = now
		}
	})
	if err != nil {
		return err
	}

	return nil
}

// verifyChecksum verifies the downloaded file's checksum.
func (w *Worker) verifyChecksum(iso *models.ISO, filepath string) error {
	// Fetch expected checksum using the original filename from the download URL
	// Checksum files reference the original filename, not our computed filename
	originalFilename := iso.GetOriginalFilename()
	expectedChecksum, err := FetchExpectedChecksum(iso.ChecksumURL, originalFilename)
	if err != nil {
		return fmt.Errorf("failed to fetch checksum: %w", err)
	}

	// Compute actual checksum
	actualChecksum, err := ComputeHash(filepath, iso.ChecksumType)
	if err != nil {
		return fmt.Errorf("failed to compute checksum: %w", err)
	}

	// Compare checksums (case-insensitive)
	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	// Update database with verified checksum
	w.db.UpdateISOChecksum(iso.ID, actualChecksum)
	iso.Checksum = actualChecksum

	return nil
}

// updateStatus updates the ISO status and triggers progress callback.
func (w *Worker) updateStatus(isoID string, status models.ISOStatus, progress int, errorMsg string) {
	w.db.UpdateISOStatus(isoID, status, errorMsg)
	if progress >= 0 {
		w.db.UpdateISOProgress(isoID, progress)
	}

	if w.progressCallback != nil {
		w.progressCallback(isoID, progress, status)
	}
}

// downloadChecksumFile downloads the checksum file and saves it.
func (w *Worker) downloadChecksumFile(checksumURL, destPath string) error {
	// Use context with timeout for checksum download
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return httputil.DownloadFile(ctx, checksumURL, destPath)
}
