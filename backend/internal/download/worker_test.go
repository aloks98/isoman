package download

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"linux-iso-manager/internal/config"
	"linux-iso-manager/internal/db"
	"linux-iso-manager/internal/models"

	"github.com/google/uuid"
)

// setupTestWorker creates a test database and worker.
func setupTestWorker(t *testing.T) (*Worker, *db.DB, string, func()) {
	// Create temp directory for test files
	tmpDir := t.TempDir()
	isoDir := filepath.Join(tmpDir, "isos")
	os.MkdirAll(isoDir, 0o755)

	// Create test database
	dbPath := filepath.Join(tmpDir, "test.db")

	// Use default config for tests
	cfg := config.Load()

	database, err := db.New(dbPath, &cfg.Database)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Create worker
	worker := NewWorker(database, isoDir, nil)

	cleanup := func() {
		database.Close()
		os.RemoveAll(tmpDir)
	}

	return worker, database, isoDir, cleanup
}

// TestWorkerDownloadSuccess tests successful download.
func TestWorkerDownloadSuccess(t *testing.T) {
	worker, database, isoDir, cleanup := setupTestWorker(t)
	defer cleanup()

	// Create test HTTP server
	testContent := []byte("test file content")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testContent)))
		w.WriteHeader(http.StatusOK)
		w.Write(testContent)
	}))
	defer server.Close()

	// Create test ISO
	iso := &models.ISO{
		ID:          uuid.New().String(),
		Name:        "test",
		Version:     "1.0",
		Arch:        "x86_64",
		FileType:    "iso",
		DownloadURL: server.URL,
		Status:      models.StatusPending,
		Progress:    0,
		CreatedAt:   time.Now(),
	}
	iso.ComputeFields()
	database.CreateISO(iso)

	// Process download
	ctx := context.Background()
	err := worker.Process(ctx, iso)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Verify file was downloaded
	finalPath := filepath.Join(isoDir, iso.FilePath)
	if _, err := os.Stat(finalPath); os.IsNotExist(err) {
		t.Errorf("Downloaded file does not exist at %s", finalPath)
	}

	// Verify file content
	content, err := os.ReadFile(finalPath)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}
	if !bytes.Equal(content, testContent) {
		t.Errorf("File content mismatch: got %q, want %q", content, testContent)
	}

	// THIS IS THE CRITICAL TEST: Verify status is "complete" in database
	updatedISO, err := database.GetISO(iso.ID)
	if err != nil {
		t.Fatalf("Failed to get updated ISO: %v", err)
	}

	if updatedISO.Status != models.StatusComplete {
		t.Errorf("Status should be 'complete', got: %s", updatedISO.Status)
	}

	if updatedISO.Progress != 100 {
		t.Errorf("Progress should be 100, got: %d", updatedISO.Progress)
	}

	if updatedISO.CompletedAt == nil {
		t.Error("CompletedAt should be set")
	}

	if updatedISO.SizeBytes != int64(len(testContent)) {
		t.Errorf("Size should be %d, got: %d", len(testContent), updatedISO.SizeBytes)
	}
}

// TestWorkerDownloadWithChecksum tests download with checksum verification.
func TestWorkerDownloadWithChecksum(t *testing.T) {
	worker, database, isoDir, cleanup := setupTestWorker(t)
	defer cleanup()

	// Test content
	testContent := []byte("test iso content")
	expectedChecksum := "c7da1a887c6ae353996b75d2ce95833ee2723f62a70386182bb2db5e26904802" // SHA256 of "test iso content"

	// Create test HTTP server for ISO download
	isoServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testContent)))
		w.WriteHeader(http.StatusOK)
		w.Write(testContent)
	}))
	defer isoServer.Close()

	// Create test HTTP server for checksum file
	checksumContent := fmt.Sprintf("%s  test-1.0-x86_64.iso\n", expectedChecksum)
	checksumServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(checksumContent))
	}))
	defer checksumServer.Close()

	// Create test ISO
	iso := &models.ISO{
		ID:           uuid.New().String(),
		Name:         "test",
		Version:      "1.0",
		Arch:         "x86_64",
		FileType:     "iso",
		DownloadURL:  isoServer.URL + "/test-1.0-x86_64.iso",
		ChecksumURL:  checksumServer.URL,
		ChecksumType: "sha256",
		Status:       models.StatusPending,
		CreatedAt:    time.Now(),
	}
	iso.ComputeFields()
	database.CreateISO(iso)

	// Process download
	ctx := context.Background()
	err := worker.Process(ctx, iso)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Verify status is complete
	updatedISO, err := database.GetISO(iso.ID)
	if err != nil {
		t.Fatalf("Failed to get updated ISO: %v", err)
	}

	if updatedISO.Status != models.StatusComplete {
		t.Errorf("Status should be 'complete', got: %s", updatedISO.Status)
	}

	if updatedISO.Checksum != expectedChecksum {
		t.Errorf("Checksum mismatch: got %s, want %s", updatedISO.Checksum, expectedChecksum)
	}

	// Verify checksum file was saved
	checksumFilePath := filepath.Join(isoDir, iso.FilePath+".sha256")
	if _, err := os.Stat(checksumFilePath); os.IsNotExist(err) {
		t.Errorf("Checksum file does not exist at %s", checksumFilePath)
	}

	// Verify checksum file content
	savedChecksumContent, err := os.ReadFile(checksumFilePath)
	if err != nil {
		t.Fatalf("Failed to read checksum file: %v", err)
	}
	if string(savedChecksumContent) != checksumContent {
		t.Errorf("Checksum file content mismatch: got %q, want %q", savedChecksumContent, checksumContent)
	}
}

// TestWorkerDownloadFailure tests download failure handling.
func TestWorkerDownloadFailure(t *testing.T) {
	worker, database, _, cleanup := setupTestWorker(t)
	defer cleanup()

	// Create test HTTP server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not Found"))
	}))
	defer server.Close()

	// Create test ISO
	iso := &models.ISO{
		ID:          uuid.New().String(),
		Name:        "test",
		Version:     "1.0",
		Arch:        "x86_64",
		FileType:    "iso",
		DownloadURL: server.URL,
		Status:      models.StatusPending,
		CreatedAt:   time.Now(),
	}
	iso.ComputeFields()
	database.CreateISO(iso)

	// Process download (should fail)
	ctx := context.Background()
	err := worker.Process(ctx, iso)
	if err == nil {
		t.Fatal("Expected download to fail, but it succeeded")
	}

	// Verify status is failed
	updatedISO, err := database.GetISO(iso.ID)
	if err != nil {
		t.Fatalf("Failed to get updated ISO: %v", err)
	}

	if updatedISO.Status != models.StatusFailed {
		t.Errorf("Status should be 'failed', got: %s", updatedISO.Status)
	}

	if updatedISO.ErrorMessage == "" {
		t.Error("ErrorMessage should be set on failure")
	}
}

// TestWorkerDownloadCancellation tests context cancellation.
func TestWorkerDownloadCancellation(t *testing.T) {
	worker, database, _, cleanup := setupTestWorker(t)
	defer cleanup()

	// Create test HTTP server with slow response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "1000000")
		w.WriteHeader(http.StatusOK)
		// Write slowly to allow cancellation
		for i := 0; i < 100; i++ {
			w.Write(make([]byte, 1000))
			time.Sleep(10 * time.Millisecond)
		}
	}))
	defer server.Close()

	// Create test ISO
	iso := &models.ISO{
		ID:          uuid.New().String(),
		Name:        "test",
		Version:     "1.0",
		Arch:        "x86_64",
		FileType:    "iso",
		DownloadURL: server.URL,
		Status:      models.StatusPending,
		CreatedAt:   time.Now(),
	}
	iso.ComputeFields()
	database.CreateISO(iso)

	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	// Start download in goroutine
	done := make(chan error, 1)
	go func() {
		done <- worker.Process(ctx, iso)
	}()

	// Cancel after short delay
	time.Sleep(50 * time.Millisecond)
	cancel()

	// Wait for download to finish
	err := <-done
	if err == nil {
		t.Fatal("Expected download to be canceled, but it succeeded")
	}

	// Verify status is failed with cancellation message
	updatedISO, err := database.GetISO(iso.ID)
	if err != nil {
		t.Fatalf("Failed to get updated ISO: %v", err)
	}

	if updatedISO.Status != models.StatusFailed {
		t.Errorf("Status should be 'failed', got: %s", updatedISO.Status)
	}

	if updatedISO.ErrorMessage != "Download canceled" {
		t.Errorf("ErrorMessage should be 'Download cancelled', got: %s", updatedISO.ErrorMessage)
	}
}

// TestWorkerProgressCallback tests progress callback.
func TestWorkerProgressCallback(t *testing.T) {
	worker, database, _, cleanup := setupTestWorker(t)
	defer cleanup()

	// Track progress updates
	var progressUpdates []int
	var statusUpdates []models.ISOStatus
	worker.progressCallback = func(isoID string, progress int, status models.ISOStatus) {
		progressUpdates = append(progressUpdates, progress)
		statusUpdates = append(statusUpdates, status)
	}

	// Create test HTTP server
	testContent := make([]byte, 10000)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testContent)))
		w.WriteHeader(http.StatusOK)
		w.Write(testContent)
	}))
	defer server.Close()

	// Create test ISO
	iso := &models.ISO{
		ID:          uuid.New().String(),
		Name:        "test",
		Version:     "1.0",
		Arch:        "x86_64",
		FileType:    "iso",
		DownloadURL: server.URL,
		Status:      models.StatusPending,
		CreatedAt:   time.Now(),
	}
	iso.ComputeFields()
	database.CreateISO(iso)

	// Process download
	ctx := context.Background()
	err := worker.Process(ctx, iso)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Verify we got progress callbacks
	if len(progressUpdates) == 0 {
		t.Error("Expected progress updates, got none")
	}

	// Verify we got status updates including "complete"
	foundComplete := false
	for _, status := range statusUpdates {
		if status == models.StatusComplete {
			foundComplete = true
			break
		}
	}
	if !foundComplete {
		t.Errorf("Expected 'complete' status in updates, got: %v", statusUpdates)
	}
}

// TestWorkerChecksumMismatch tests checksum verification failure.
func TestWorkerChecksumMismatch(t *testing.T) {
	worker, database, _, cleanup := setupTestWorker(t)
	defer cleanup()

	// Test content
	testContent := []byte("test iso content")
	wrongChecksum := "0000000000000000000000000000000000000000000000000000000000000000"

	// Create test HTTP server for ISO download
	isoServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testContent)))
		w.WriteHeader(http.StatusOK)
		w.Write(testContent)
	}))
	defer isoServer.Close()

	// Create test HTTP server for checksum file with wrong checksum
	checksumContent := fmt.Sprintf("%s  test-1.0-x86_64.iso\n", wrongChecksum)
	checksumServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(checksumContent))
	}))
	defer checksumServer.Close()

	// Create test ISO
	iso := &models.ISO{
		ID:           uuid.New().String(),
		Name:         "test",
		Version:      "1.0",
		Arch:         "x86_64",
		FileType:     "iso",
		DownloadURL:  isoServer.URL + "/test-1.0-x86_64.iso",
		ChecksumURL:  checksumServer.URL,
		ChecksumType: "sha256",
		Status:       models.StatusPending,
		CreatedAt:    time.Now(),
	}
	iso.ComputeFields()
	database.CreateISO(iso)

	// Process download (should fail on checksum)
	ctx := context.Background()
	err := worker.Process(ctx, iso)
	if err == nil {
		t.Fatal("Expected checksum verification to fail, but it succeeded")
	}

	// Verify status is failed
	updatedISO, err := database.GetISO(iso.ID)
	if err != nil {
		t.Fatalf("Failed to get updated ISO: %v", err)
	}

	if updatedISO.Status != models.StatusFailed {
		t.Errorf("Status should be 'failed', got: %s", updatedISO.Status)
	}

	if updatedISO.ErrorMessage == "" {
		t.Error("ErrorMessage should be set on checksum failure")
	}
}

// TestWorkerNestedDirectoryCreation tests that nested directories are created.
func TestWorkerNestedDirectoryCreation(t *testing.T) {
	worker, database, isoDir, cleanup := setupTestWorker(t)
	defer cleanup()

	// Create test HTTP server
	testContent := []byte("test content")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testContent)))
		w.WriteHeader(http.StatusOK)
		w.Write(testContent)
	}))
	defer server.Close()

	// Create test ISO with nested path
	iso := &models.ISO{
		ID:          uuid.New().String(),
		Name:        "alpine",
		Version:     "3.19.1",
		Arch:        "x86_64",
		FileType:    "iso",
		DownloadURL: server.URL,
		Status:      models.StatusPending,
		CreatedAt:   time.Now(),
	}
	iso.ComputeFields() // This creates: alpine/3.19.1/x86_64/alpine-3.19.1-x86_64.iso
	database.CreateISO(iso)

	// Process download
	ctx := context.Background()
	err := worker.Process(ctx, iso)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Verify nested directory structure was created
	expectedPath := filepath.Join(isoDir, "alpine", "3.19.1", "x86_64", "alpine-3.19.1-x86_64.iso")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Nested directory structure not created: %s", expectedPath)
	}
}

// TestWorkerTempFileCleanup tests that temp files are cleaned up on error.
func TestWorkerTempFileCleanup(t *testing.T) {
	worker, database, isoDir, cleanup := setupTestWorker(t)
	defer cleanup()

	// Create test HTTP server that fails mid-download
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "10000")
		w.WriteHeader(http.StatusOK)
		w.Write(make([]byte, 1000))
		// Abort connection
	}))
	defer server.Close()

	// Create test ISO
	iso := &models.ISO{
		ID:          uuid.New().String(),
		Name:        "test",
		Version:     "1.0",
		Arch:        "x86_64",
		FileType:    "iso",
		DownloadURL: server.URL,
		Status:      models.StatusPending,
		CreatedAt:   time.Now(),
	}
	iso.ComputeFields()
	database.CreateISO(iso)

	// Process download (should fail)
	ctx := context.Background()
	worker.Process(ctx, iso)

	// Verify temp file was cleaned up
	tmpFile := filepath.Join(isoDir, ".tmp", iso.Filename)
	if _, err := os.Stat(tmpFile); !os.IsNotExist(err) {
		t.Errorf("Temp file should be cleaned up: %s", tmpFile)
	}
}
