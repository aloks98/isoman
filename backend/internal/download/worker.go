package download

import (
	"context"
	"fmt"
	"io"
	"linux-iso-manager/internal/db"
	"linux-iso-manager/internal/models"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// ProgressCallback is called when download progress updates
type ProgressCallback func(isoID string, progress int, status models.ISOStatus)

// Worker handles the download and verification of a single ISO
type Worker struct {
	db               *db.DB
	isoDir           string
	tmpDir           string
	progressCallback ProgressCallback
}

// NewWorker creates a new download worker
func NewWorker(database *db.DB, isoDir string, callback ProgressCallback) *Worker {
	tmpDir := filepath.Join(isoDir, ".tmp")
	return &Worker{
		db:               database,
		isoDir:           isoDir,
		tmpDir:           tmpDir,
		progressCallback: callback,
	}
}

// Process downloads and verifies an ISO
func (w *Worker) Process(ctx context.Context, iso *models.ISO) error {
	// Ensure tmp directory exists
	if err := os.MkdirAll(w.tmpDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Use the computed FilePath for nested directory structure
	// FilePath example: "alpine/3.19.1/x86_64/alpine-3.19.1-x86_64.iso"
	tmpFile := filepath.Join(w.tmpDir, iso.Filename)
	finalFile := filepath.Join(w.isoDir, iso.FilePath)

	// Create the nested directory structure for the final file
	finalDir := filepath.Dir(finalFile)
	if err := os.MkdirAll(finalDir, 0755); err != nil {
		return fmt.Errorf("failed to create final directory: %w", err)
	}

	// Clean up temp file on error
	defer func() {
		if _, err := os.Stat(tmpFile); err == nil {
			os.Remove(tmpFile)
		}
	}()

	// Update status to downloading
	w.updateStatus(iso.ID, models.StatusDownloading, 0, "")

	// Download the file
	if err := w.download(ctx, iso, tmpFile); err != nil {
		// Check if it was cancelled
		if ctx.Err() == context.Canceled {
			w.updateStatus(iso.ID, models.StatusFailed, 0, "Download cancelled")
			return fmt.Errorf("download cancelled: %w", ctx.Err())
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
		checksumFile := finalFile + "." + iso.ChecksumType
		if err := w.downloadChecksumFile(iso.ChecksumURL, checksumFile); err != nil {
			log.Printf("WARNING: Failed to save checksum file for %s: %v", iso.ID, err)
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

	// Retry UpdateISO if database is busy
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		if err := w.db.UpdateISO(iso); err != nil {
			if i < maxRetries-1 {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			log.Printf("ERROR: Failed to update ISO %s to complete after %d retries: %v", iso.ID, maxRetries, err)
		}
		break
	}

	return nil
}

// download downloads the ISO file with progress tracking
func (w *Worker) download(ctx context.Context, iso *models.ISO, destPath string) error {
	// Create HTTP request with context
	req, err := http.NewRequestWithContext(ctx, "GET", iso.DownloadURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to start download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}

	// Get content length and update database
	contentLength := resp.ContentLength
	if contentLength > 0 {
		w.db.UpdateISOSize(iso.ID, contentLength)
		iso.SizeBytes = contentLength
	}

	// Create destination file
	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	// Download with progress tracking
	var downloaded int64
	lastProgress := -1
	lastUpdate := time.Now()
	buf := make([]byte, 32*1024) // 32KB buffer

	for {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := out.Write(buf[:n]); writeErr != nil {
				return fmt.Errorf("failed to write to file: %w", writeErr)
			}
			downloaded += int64(n)

			// Calculate progress
			var progress int
			if contentLength > 0 {
				progress = int((downloaded * 100) / contentLength)
			}

			// Update progress every 1% or every second
			now := time.Now()
			if progress != lastProgress && (progress-lastProgress >= 1 || now.Sub(lastUpdate) >= time.Second) {
				w.updateStatus(iso.ID, models.StatusDownloading, progress, "")
				lastProgress = progress
				lastUpdate = now
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("download error: %w", err)
		}
	}

	return nil
}

// verifyChecksum verifies the downloaded file's checksum
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

// updateStatus updates the ISO status and triggers progress callback
func (w *Worker) updateStatus(isoID string, status models.ISOStatus, progress int, errorMsg string) {
	w.db.UpdateISOStatus(isoID, status, errorMsg)
	if progress >= 0 {
		w.db.UpdateISOProgress(isoID, progress)
	}

	if w.progressCallback != nil {
		w.progressCallback(isoID, progress, status)
	}
}

// downloadChecksumFile downloads the checksum file and saves it
func (w *Worker) downloadChecksumFile(checksumURL, destPath string) error {
	resp, err := http.Get(checksumURL)
	if err != nil {
		return fmt.Errorf("failed to fetch checksum file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("checksum file download failed with status: %s", resp.Status)
	}

	// Create destination file
	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create checksum file: %w", err)
	}
	defer out.Close()

	// Copy content to file
	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("failed to write checksum file: %w", err)
	}

	return nil
}
