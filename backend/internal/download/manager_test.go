package download

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"linux-iso-manager/internal/config"
	"linux-iso-manager/internal/db"
	"linux-iso-manager/internal/models"

	"github.com/google/uuid"
)

// setupTestManager creates a test manager with database.
func setupTestManager(t *testing.T, workerCount int) (*Manager, *db.DB, string, func()) {
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

	// Create manager
	manager := NewManager(database, isoDir, workerCount)

	cleanup := func() {
		manager.Stop()
		database.Close()
		os.RemoveAll(tmpDir)
	}

	return manager, database, isoDir, cleanup
}

// TestManagerStartStop tests starting and stopping the manager.
func TestManagerStartStop(t *testing.T) {
	manager, _, _, _ := setupTestManager(t, 2)
	// Don't defer cleanup here since we're testing Stop explicitly

	// Start manager
	manager.Start()

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	// Stop manager
	manager.Stop()

	// Manager should stop cleanly without panic
}

// TestManagerQueueDownload tests queuing downloads.
func TestManagerQueueDownload(t *testing.T) {
	manager, database, _, cleanup := setupTestManager(t, 1)
	defer cleanup()

	// Create test HTTP server
	testContent := []byte("test content")
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

	// Start manager
	manager.Start()

	// Queue download
	manager.QueueDownload(iso)

	// Wait for download to complete
	time.Sleep(2 * time.Second)

	// Verify download completed
	updatedISO, err := database.GetISO(iso.ID)
	if err != nil {
		t.Fatalf("Failed to get updated ISO: %v", err)
	}

	if updatedISO.Status != models.StatusComplete {
		t.Errorf("Expected status 'complete', got: %s", updatedISO.Status)
	}
}

// TestManagerMultipleWorkers tests multiple concurrent downloads.
func TestManagerMultipleWorkers(t *testing.T) {
	manager, database, _, cleanup := setupTestManager(t, 3)
	defer cleanup()

	// Create test HTTP server with slow response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		testContent := []byte("test content")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testContent)))
		w.WriteHeader(http.StatusOK)
		// Slow write to ensure concurrent downloads
		time.Sleep(200 * time.Millisecond)
		w.Write(testContent)
	}))
	defer server.Close()

	// Create multiple test ISOs
	isos := make([]*models.ISO, 5)
	for i := 0; i < 5; i++ {
		iso := &models.ISO{
			ID:          uuid.New().String(),
			Name:        fmt.Sprintf("test-%d", i),
			Version:     "1.0",
			Arch:        "x86_64",
			FileType:    "iso",
			DownloadURL: server.URL,
			Status:      models.StatusPending,
			CreatedAt:   time.Now(),
		}
		iso.ComputeFields()
		database.CreateISO(iso)
		isos[i] = iso
	}

	// Start manager
	manager.Start()

	// Queue all downloads
	startTime := time.Now()
	for _, iso := range isos {
		manager.QueueDownload(iso)
	}

	// Wait for all downloads to complete with polling
	timeout := time.After(10 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	allComplete := false
	for !allComplete {
		select {
		case <-timeout:
			t.Fatal("Timeout waiting for downloads to complete")
		case <-ticker.C:
			allComplete = true
			for _, iso := range isos {
				updatedISO, err := database.GetISO(iso.ID)
				if err != nil {
					t.Fatalf("Failed to get updated ISO %s: %v", iso.ID, err)
				}
				if updatedISO.Status != models.StatusComplete {
					t.Logf("ISO %s status: %s, progress: %d", iso.ID[:8], updatedISO.Status, updatedISO.Progress)
					allComplete = false
					break
				}
			}
		}
	}

	// Verify all downloads completed
	for _, iso := range isos {
		updatedISO, err := database.GetISO(iso.ID)
		if err != nil {
			t.Fatalf("Failed to get updated ISO %s: %v", iso.ID, err)
		}

		if updatedISO.Status != models.StatusComplete {
			t.Errorf("ISO %s: expected status 'complete', got: %s", iso.ID, updatedISO.Status)
		}
	}

	// With 3 workers and 5 downloads (200ms each), should take ~2 rounds
	// Sequential would take 1000ms, parallel with 3 workers should be faster
	elapsed := time.Since(startTime)
	if elapsed > 3*time.Second {
		t.Logf("Download took %v, might want to verify parallel execution", elapsed)
	}
}

// TestManagerProgressCallback tests progress callbacks.
func TestManagerProgressCallback(t *testing.T) {
	manager, database, _, cleanup := setupTestManager(t, 1)
	defer cleanup()

	// Track progress callbacks with mutex protection
	var mu sync.Mutex
	progressCalls := 0
	var lastStatus models.ISOStatus
	manager.SetProgressCallback(func(isoID string, progress int, status models.ISOStatus) {
		mu.Lock()
		defer mu.Unlock()
		progressCalls++
		lastStatus = status
	})

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

	// Start manager
	manager.Start()

	// Queue download
	manager.QueueDownload(iso)

	// Wait for download to complete
	time.Sleep(2 * time.Second)

	// Verify we got progress callbacks
	mu.Lock()
	calls := progressCalls
	status := lastStatus
	mu.Unlock()

	if calls == 0 {
		t.Error("Expected progress callbacks, got none")
	}

	if status != models.StatusComplete {
		t.Errorf("Last status should be 'complete', got: %s", status)
	}
}

// TestManagerGracefulShutdown tests that manager stops gracefully.
func TestManagerGracefulShutdown(t *testing.T) {
	manager, database, _, cleanup := setupTestManager(t, 2)
	defer cleanup()

	// Create test HTTP server with slow response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		testContent := make([]byte, 1000000)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testContent)))
		w.WriteHeader(http.StatusOK)
		// Write slowly
		for i := 0; i < 100; i++ {
			w.Write(testContent[i*10000 : (i+1)*10000])
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

	// Start manager
	manager.Start()

	// Queue download
	manager.QueueDownload(iso)

	// Give download time to start
	time.Sleep(100 * time.Millisecond)

	// Stop manager (should cancel active downloads)
	manager.Stop()

	// Verify download was canceled
	updatedISO, err := database.GetISO(iso.ID)
	if err != nil {
		t.Fatalf("Failed to get updated ISO: %v", err)
	}

	if updatedISO.Status != models.StatusFailed {
		t.Errorf("Expected status 'failed' after cancellation, got: %s", updatedISO.Status)
	}

	if updatedISO.ErrorMessage != "Download canceled" {
		t.Errorf("Expected error message 'Download cancelled', got: %s", updatedISO.ErrorMessage)
	}
}

// TestManagerQueueCapacity tests queue buffer capacity.
func TestManagerQueueCapacity(t *testing.T) {
	manager, database, _, cleanup := setupTestManager(t, 1)
	defer cleanup()

	// Don't start the manager yet (so downloads queue up)

	// Create test ISOs (more than queue size of 100)
	for i := 0; i < 110; i++ {
		iso := &models.ISO{
			ID:          uuid.New().String(),
			Name:        fmt.Sprintf("test-%d", i),
			Version:     "1.0",
			Arch:        "x86_64",
			FileType:    "iso",
			DownloadURL: "http://example.com/test.iso",
			Status:      models.StatusPending,
			CreatedAt:   time.Now(),
		}
		iso.ComputeFields()
		database.CreateISO(iso)

		// Try to queue (should not block for first 100)
		done := make(chan bool, 1)
		go func() {
			manager.QueueDownload(iso)
			done <- true
		}()

		// Should complete quickly for first 100
		if i < 100 {
			select {
			case <-done:
				// OK
			case <-time.After(100 * time.Millisecond):
				t.Fatalf("Queueing item %d blocked unexpectedly", i)
			}
		}
	}
}
