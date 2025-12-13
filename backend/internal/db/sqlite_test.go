package db

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"linux-iso-manager/internal/config"
	"linux-iso-manager/internal/models"

	"github.com/google/uuid"
)

// setupTestDB creates a temporary test database.
func setupTestDB(t *testing.T) (*DB, func()) {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Use default config for tests
	cfg := config.Load()

	db, err := New(dbPath, &cfg.Database)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	cleanup := func() {
		db.Close()
		os.RemoveAll(tmpDir)
	}

	return db, cleanup
}

// createTestISO creates a test ISO for use in tests.
func createTestISO() *models.ISO {
	return &models.ISO{
		ID:           uuid.New().String(),
		Name:         "Test ISO",
		Filename:     "test.iso",
		SizeBytes:    1024,
		Checksum:     "abc123",
		ChecksumType: "sha256",
		DownloadURL:  "https://example.com/test.iso",
		ChecksumURL:  "https://example.com/test.iso.sha256",
		Status:       models.StatusPending,
		Progress:     0,
		ErrorMessage: "",
		CreatedAt:    time.Now(),
		CompletedAt:  nil,
	}
}

func TestNew(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Use default config for tests
	cfg := config.Load()

	db, err := New(dbPath, &cfg.Database)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer db.Close()

	// Verify database file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("Database file was not created")
	}
}

func TestCreateISO(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	iso := createTestISO()

	err := db.CreateISO(iso)
	if err != nil {
		t.Fatalf("CreateISO() failed: %v", err)
	}

	// Verify the ISO was created by retrieving it
	retrieved, err := db.GetISO(iso.ID)
	if err != nil {
		t.Fatalf("GetISO() failed after create: %v", err)
	}

	if retrieved.ID != iso.ID {
		t.Errorf("Expected ID %s, got %s", iso.ID, retrieved.ID)
	}
	if retrieved.Name != iso.Name {
		t.Errorf("Expected Name %s, got %s", iso.Name, retrieved.Name)
	}
}

func TestGetISO(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	iso := createTestISO()
	err := db.CreateISO(iso)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	t.Run("ExistingISO", func(t *testing.T) {
		retrieved, err := db.GetISO(iso.ID)
		if err != nil {
			t.Fatalf("GetISO() failed: %v", err)
		}

		if retrieved.ID != iso.ID {
			t.Errorf("Expected ID %s, got %s", iso.ID, retrieved.ID)
		}
		if retrieved.Name != iso.Name {
			t.Errorf("Expected Name %s, got %s", iso.Name, retrieved.Name)
		}
		if retrieved.Filename != iso.Filename {
			t.Errorf("Expected Filename %s, got %s", iso.Filename, retrieved.Filename)
		}
		if retrieved.Status != iso.Status {
			t.Errorf("Expected Status %s, got %s", iso.Status, retrieved.Status)
		}
	})

	t.Run("NonExistentISO", func(t *testing.T) {
		_, err := db.GetISO("nonexistent-id")
		if err == nil {
			t.Error("Expected error for non-existent ISO, got nil")
		}
	})
}

func TestListISOs(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	t.Run("EmptyDatabase", func(t *testing.T) {
		isos, err := db.ListISOs()
		if err != nil {
			t.Fatalf("ListISOs() failed: %v", err)
		}
		if len(isos) != 0 {
			t.Errorf("Expected 0 ISOs, got %d", len(isos))
		}
	})

	t.Run("MultipleISOs", func(t *testing.T) {
		// Create multiple ISOs with unique filenames
		iso1 := createTestISO()
		iso1.Name = "ISO 1"
		iso1.Filename = "iso1.iso"
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps

		iso2 := createTestISO()
		iso2.Name = "ISO 2"
		iso2.Filename = "iso2.iso"
		time.Sleep(10 * time.Millisecond)

		iso3 := createTestISO()
		iso3.Name = "ISO 3"
		iso3.Filename = "iso3.iso"

		db.CreateISO(iso1)
		db.CreateISO(iso2)
		db.CreateISO(iso3)

		isos, err := db.ListISOs()
		if err != nil {
			t.Fatalf("ListISOs() failed: %v", err)
		}

		if len(isos) != 3 {
			t.Fatalf("Expected 3 ISOs, got %d", len(isos))
		}

		// Verify ordering (DESC by created_at)
		// Most recent should be first
		if isos[0].Name != "ISO 3" {
			t.Errorf("Expected first ISO to be 'ISO 3', got '%s'", isos[0].Name)
		}
	})
}

func TestUpdateISO(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	iso := createTestISO()
	err := db.CreateISO(iso)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Update fields
	iso.Name = "Updated Name"
	iso.Progress = 75
	iso.Status = models.StatusDownloading
	completedAt := time.Now()
	iso.CompletedAt = &completedAt

	err = db.UpdateISO(iso)
	if err != nil {
		t.Fatalf("UpdateISO() failed: %v", err)
	}

	// Verify updates
	retrieved, err := db.GetISO(iso.ID)
	if err != nil {
		t.Fatalf("GetISO() failed: %v", err)
	}

	if retrieved.Name != "Updated Name" {
		t.Errorf("Expected Name 'Updated Name', got '%s'", retrieved.Name)
	}
	if retrieved.Progress != 75 {
		t.Errorf("Expected Progress 75, got %d", retrieved.Progress)
	}
	if retrieved.Status != models.StatusDownloading {
		t.Errorf("Expected Status %s, got %s", models.StatusDownloading, retrieved.Status)
	}
	if retrieved.CompletedAt == nil {
		t.Error("Expected CompletedAt to be set, got nil")
	}
}

func TestUpdateISOStatus(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	iso := createTestISO()
	err := db.CreateISO(iso)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	err = db.UpdateISOStatus(iso.ID, models.StatusComplete, "")
	if err != nil {
		t.Fatalf("UpdateISOStatus() failed: %v", err)
	}

	retrieved, err := db.GetISO(iso.ID)
	if err != nil {
		t.Fatalf("GetISO() failed: %v", err)
	}

	if retrieved.Status != models.StatusComplete {
		t.Errorf("Expected Status %s, got %s", models.StatusComplete, retrieved.Status)
	}
	if retrieved.ErrorMessage != "" {
		t.Errorf("Expected empty ErrorMessage, got '%s'", retrieved.ErrorMessage)
	}
}

func TestUpdateISOStatusWithError(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	iso := createTestISO()
	err := db.CreateISO(iso)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	errorMsg := "Download failed: timeout"
	err = db.UpdateISOStatus(iso.ID, models.StatusFailed, errorMsg)
	if err != nil {
		t.Fatalf("UpdateISOStatus() failed: %v", err)
	}

	retrieved, err := db.GetISO(iso.ID)
	if err != nil {
		t.Fatalf("GetISO() failed: %v", err)
	}

	if retrieved.Status != models.StatusFailed {
		t.Errorf("Expected Status %s, got %s", models.StatusFailed, retrieved.Status)
	}
	if retrieved.ErrorMessage != errorMsg {
		t.Errorf("Expected ErrorMessage '%s', got '%s'", errorMsg, retrieved.ErrorMessage)
	}
}

func TestUpdateISOProgress(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	iso := createTestISO()
	err := db.CreateISO(iso)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	testCases := []int{0, 25, 50, 75, 100}
	for _, progress := range testCases {
		err = db.UpdateISOProgress(iso.ID, progress)
		if err != nil {
			t.Fatalf("UpdateISOProgress(%d) failed: %v", progress, err)
		}

		retrieved, err := db.GetISO(iso.ID)
		if err != nil {
			t.Fatalf("GetISO() failed: %v", err)
		}

		if retrieved.Progress != progress {
			t.Errorf("Expected Progress %d, got %d", progress, retrieved.Progress)
		}
	}
}

func TestUpdateISOSize(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	iso := createTestISO()
	err := db.CreateISO(iso)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	newSize := int64(5368709120) // 5 GB
	err = db.UpdateISOSize(iso.ID, newSize)
	if err != nil {
		t.Fatalf("UpdateISOSize() failed: %v", err)
	}

	retrieved, err := db.GetISO(iso.ID)
	if err != nil {
		t.Fatalf("GetISO() failed: %v", err)
	}

	if retrieved.SizeBytes != newSize {
		t.Errorf("Expected SizeBytes %d, got %d", newSize, retrieved.SizeBytes)
	}
}

func TestUpdateISOChecksum(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	iso := createTestISO()
	err := db.CreateISO(iso)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	newChecksum := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	err = db.UpdateISOChecksum(iso.ID, newChecksum)
	if err != nil {
		t.Fatalf("UpdateISOChecksum() failed: %v", err)
	}

	retrieved, err := db.GetISO(iso.ID)
	if err != nil {
		t.Fatalf("GetISO() failed: %v", err)
	}

	if retrieved.Checksum != newChecksum {
		t.Errorf("Expected Checksum '%s', got '%s'", newChecksum, retrieved.Checksum)
	}
}

func TestDeleteISO(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	iso := createTestISO()
	err := db.CreateISO(iso)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Verify ISO exists
	_, err = db.GetISO(iso.ID)
	if err != nil {
		t.Fatalf("ISO should exist before delete: %v", err)
	}

	// Delete the ISO
	err = db.DeleteISO(iso.ID)
	if err != nil {
		t.Fatalf("DeleteISO() failed: %v", err)
	}

	// Verify ISO no longer exists
	_, err = db.GetISO(iso.ID)
	if err == nil {
		t.Error("Expected error when getting deleted ISO, got nil")
	}
}

func TestDeleteNonExistentISO(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Should not error when deleting non-existent ISO
	err := db.DeleteISO("nonexistent-id")
	if err != nil {
		t.Errorf("DeleteISO() should not error on non-existent ID: %v", err)
	}
}

func TestDuplicateCompositeKeyRejected(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	iso1 := createTestISO()
	iso1.Name = "alpine"
	iso1.Version = "3.19.1"
	iso1.Arch = "x86_64"
	iso1.Edition = ""
	iso1.FileType = "iso"

	// Create first ISO
	err := db.CreateISO(iso1)
	if err != nil {
		t.Fatalf("CreateISO() failed for first ISO: %v", err)
	}

	// Try to create second ISO with same composite key
	iso2 := createTestISO()
	iso2.Name = "alpine"                                   // Same
	iso2.Version = "3.19.1"                                // Same
	iso2.Arch = "x86_64"                                   // Same
	iso2.Edition = ""                                      // Same
	iso2.FileType = "iso"                                  // Same
	iso2.DownloadURL = "http://different-url.com/file.iso" // Different URL

	err = db.CreateISO(iso2)
	if err == nil {
		t.Error("Expected error when creating ISO with duplicate composite key, got nil")
	}
}

func TestConcurrentOperations(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create base ISO
	iso := createTestISO()
	err := db.CreateISO(iso)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Run concurrent updates
	done := make(chan bool, 3)

	go func() {
		for i := 0; i <= 100; i += 10 {
			db.UpdateISOProgress(iso.ID, i)
			time.Sleep(5 * time.Millisecond)
		}
		done <- true
	}()

	go func() {
		for i := int64(0); i < 1000000; i += 100000 {
			db.UpdateISOSize(iso.ID, i)
			time.Sleep(5 * time.Millisecond)
		}
		done <- true
	}()

	go func() {
		statuses := []models.ISOStatus{
			models.StatusPending,
			models.StatusDownloading,
			models.StatusVerifying,
			models.StatusComplete,
		}
		for _, status := range statuses {
			db.UpdateISOStatus(iso.ID, status, "")
			time.Sleep(10 * time.Millisecond)
		}
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}

	// Verify ISO still exists and is valid
	retrieved, err := db.GetISO(iso.ID)
	if err != nil {
		t.Fatalf("GetISO() failed after concurrent operations: %v", err)
	}

	if retrieved.ID != iso.ID {
		t.Error("ISO corrupted after concurrent operations")
	}
}
