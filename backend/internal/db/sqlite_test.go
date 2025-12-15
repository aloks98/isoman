package db

import (
	"fmt"
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

func TestISOExists(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	iso := createTestISO()
	iso.Name = "alpine"
	iso.Version = "3.19.1"
	iso.Arch = "x86_64"
	iso.Edition = "standard"
	iso.FileType = "iso"
	err := db.CreateISO(iso)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	t.Run("ExistingISO", func(t *testing.T) {
		exists, err := db.ISOExists("alpine", "3.19.1", "x86_64", "standard", "iso")
		if err != nil {
			t.Fatalf("ISOExists() failed: %v", err)
		}
		if !exists {
			t.Error("Expected ISOExists to return true for existing ISO")
		}
	})

	t.Run("NonExistentISO_DifferentName", func(t *testing.T) {
		exists, err := db.ISOExists("ubuntu", "3.19.1", "x86_64", "standard", "iso")
		if err != nil {
			t.Fatalf("ISOExists() failed: %v", err)
		}
		if exists {
			t.Error("Expected ISOExists to return false for non-existent ISO")
		}
	})

	t.Run("NonExistentISO_DifferentVersion", func(t *testing.T) {
		exists, err := db.ISOExists("alpine", "3.20.0", "x86_64", "standard", "iso")
		if err != nil {
			t.Fatalf("ISOExists() failed: %v", err)
		}
		if exists {
			t.Error("Expected ISOExists to return false for different version")
		}
	})

	t.Run("NonExistentISO_DifferentArch", func(t *testing.T) {
		exists, err := db.ISOExists("alpine", "3.19.1", "aarch64", "standard", "iso")
		if err != nil {
			t.Fatalf("ISOExists() failed: %v", err)
		}
		if exists {
			t.Error("Expected ISOExists to return false for different arch")
		}
	})

	t.Run("NonExistentISO_DifferentEdition", func(t *testing.T) {
		exists, err := db.ISOExists("alpine", "3.19.1", "x86_64", "minimal", "iso")
		if err != nil {
			t.Fatalf("ISOExists() failed: %v", err)
		}
		if exists {
			t.Error("Expected ISOExists to return false for different edition")
		}
	})

	t.Run("NonExistentISO_DifferentFileType", func(t *testing.T) {
		exists, err := db.ISOExists("alpine", "3.19.1", "x86_64", "standard", "qcow2")
		if err != nil {
			t.Fatalf("ISOExists() failed: %v", err)
		}
		if exists {
			t.Error("Expected ISOExists to return false for different file type")
		}
	})
}

func TestGetISOByComposite(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	iso := createTestISO()
	iso.Name = "ubuntu"
	iso.Version = "24.04"
	iso.Arch = "x86_64"
	iso.Edition = "desktop"
	iso.FileType = "iso"
	iso.Filename = "ubuntu-24.04-desktop-x86_64.iso"
	err := db.CreateISO(iso)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	t.Run("ExistingISO", func(t *testing.T) {
		retrieved, err := db.GetISOByComposite("ubuntu", "24.04", "x86_64", "desktop", "iso")
		if err != nil {
			t.Fatalf("GetISOByComposite() failed: %v", err)
		}

		if retrieved.ID != iso.ID {
			t.Errorf("Expected ID %s, got %s", iso.ID, retrieved.ID)
		}
		if retrieved.Name != "ubuntu" {
			t.Errorf("Expected Name 'ubuntu', got '%s'", retrieved.Name)
		}
		if retrieved.Version != "24.04" {
			t.Errorf("Expected Version '24.04', got '%s'", retrieved.Version)
		}
		if retrieved.Arch != "x86_64" {
			t.Errorf("Expected Arch 'x86_64', got '%s'", retrieved.Arch)
		}
		if retrieved.Edition != "desktop" {
			t.Errorf("Expected Edition 'desktop', got '%s'", retrieved.Edition)
		}
		if retrieved.FileType != "iso" {
			t.Errorf("Expected FileType 'iso', got '%s'", retrieved.FileType)
		}
	})

	t.Run("NonExistentISO", func(t *testing.T) {
		_, err := db.GetISOByComposite("nonexistent", "1.0", "x86_64", "", "iso")
		if err == nil {
			t.Error("Expected error for non-existent ISO, got nil")
		}
	})

	t.Run("PartialMatch_DifferentEdition", func(t *testing.T) {
		_, err := db.GetISOByComposite("ubuntu", "24.04", "x86_64", "server", "iso")
		if err == nil {
			t.Error("Expected error when edition doesn't match, got nil")
		}
	})
}

func TestListISOsPaginated(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create 25 ISOs for pagination testing with unique composite keys
	for i := 0; i < 25; i++ {
		iso := createTestISO()
		iso.Name = "test-iso"
		iso.Version = fmt.Sprintf("1.0.%d", i) // Unique version for each
		iso.Arch = "x86_64"
		iso.Edition = ""
		iso.FileType = "iso"
		iso.Filename = fmt.Sprintf("test-iso-1.0.%d-x86_64.iso", i)
		db.CreateISO(iso)
		time.Sleep(2 * time.Millisecond) // Ensure different timestamps
	}

	t.Run("DefaultPagination", func(t *testing.T) {
		result, err := db.ListISOsPaginated(ListISOsParams{})
		if err != nil {
			t.Fatalf("ListISOsPaginated() failed: %v", err)
		}

		// Default page size should be 10
		if len(result.ISOs) != 10 {
			t.Errorf("Expected 10 ISOs, got %d", len(result.ISOs))
		}
		if result.Total != 25 {
			t.Errorf("Expected total 25, got %d", result.Total)
		}
		if result.Page != 1 {
			t.Errorf("Expected page 1, got %d", result.Page)
		}
		if result.TotalPages != 3 {
			t.Errorf("Expected 3 total pages, got %d", result.TotalPages)
		}
	})

	t.Run("CustomPageSize", func(t *testing.T) {
		result, err := db.ListISOsPaginated(ListISOsParams{
			Page:     1,
			PageSize: 5,
		})
		if err != nil {
			t.Fatalf("ListISOsPaginated() failed: %v", err)
		}

		if len(result.ISOs) != 5 {
			t.Errorf("Expected 5 ISOs, got %d", len(result.ISOs))
		}
		if result.TotalPages != 5 {
			t.Errorf("Expected 5 total pages, got %d", result.TotalPages)
		}
	})

	t.Run("SecondPage", func(t *testing.T) {
		result, err := db.ListISOsPaginated(ListISOsParams{
			Page:     2,
			PageSize: 10,
		})
		if err != nil {
			t.Fatalf("ListISOsPaginated() failed: %v", err)
		}

		if len(result.ISOs) != 10 {
			t.Errorf("Expected 10 ISOs on page 2, got %d", len(result.ISOs))
		}
		if result.Page != 2 {
			t.Errorf("Expected page 2, got %d", result.Page)
		}
	})

	t.Run("LastPage", func(t *testing.T) {
		result, err := db.ListISOsPaginated(ListISOsParams{
			Page:     3,
			PageSize: 10,
		})
		if err != nil {
			t.Fatalf("ListISOsPaginated() failed: %v", err)
		}

		// Last page should have 5 items (25 total, 10 per page)
		if len(result.ISOs) != 5 {
			t.Errorf("Expected 5 ISOs on last page, got %d", len(result.ISOs))
		}
	})

	t.Run("MaxPageSize", func(t *testing.T) {
		result, err := db.ListISOsPaginated(ListISOsParams{
			PageSize: 200, // Above max of 100
		})
		if err != nil {
			t.Fatalf("ListISOsPaginated() failed: %v", err)
		}

		// Should be capped at 100
		if result.PageSize != 100 {
			t.Errorf("Expected page size to be capped at 100, got %d", result.PageSize)
		}
	})

	t.Run("SortByName", func(t *testing.T) {
		result, err := db.ListISOsPaginated(ListISOsParams{
			SortBy:  "name",
			SortDir: "asc",
		})
		if err != nil {
			t.Fatalf("ListISOsPaginated() failed: %v", err)
		}

		if len(result.ISOs) == 0 {
			t.Fatal("Expected ISOs to be returned")
		}
		// All ISOs have same name, so just verify no error
	})

	t.Run("InvalidSortColumn", func(t *testing.T) {
		result, err := db.ListISOsPaginated(ListISOsParams{
			SortBy: "invalid_column",
		})
		if err != nil {
			t.Fatalf("ListISOsPaginated() failed: %v", err)
		}

		// Should default to created_at
		if len(result.ISOs) == 0 {
			t.Error("Expected ISOs to be returned")
		}
	})
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

func TestListISOsWithMissingSize(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create ISOs with various states
	// ISO 1: complete with size_bytes = 0 (should be returned)
	iso1 := createTestISO()
	iso1.ID = "iso-missing-size-1"
	iso1.Name = "missing-size-1"
	iso1.Status = models.StatusComplete
	iso1.SizeBytes = 0
	if err := db.CreateISO(iso1); err != nil {
		t.Fatalf("Failed to create iso1: %v", err)
	}

	// ISO 2: complete with size_bytes > 0 (should NOT be returned)
	iso2 := createTestISO()
	iso2.ID = "iso-has-size"
	iso2.Name = "has-size"
	iso2.Status = models.StatusComplete
	iso2.SizeBytes = 1000000
	if err := db.CreateISO(iso2); err != nil {
		t.Fatalf("Failed to create iso2: %v", err)
	}

	// ISO 3: pending with size_bytes = 0 (should NOT be returned - not complete)
	iso3 := createTestISO()
	iso3.ID = "iso-pending"
	iso3.Name = "pending"
	iso3.Status = models.StatusPending
	iso3.SizeBytes = 0
	if err := db.CreateISO(iso3); err != nil {
		t.Fatalf("Failed to create iso3: %v", err)
	}

	// ISO 4: failed with size_bytes = 0 (should NOT be returned - not complete)
	iso4 := createTestISO()
	iso4.ID = "iso-failed"
	iso4.Name = "failed"
	iso4.Status = models.StatusFailed
	iso4.SizeBytes = 0
	if err := db.CreateISO(iso4); err != nil {
		t.Fatalf("Failed to create iso4: %v", err)
	}

	// ISO 5: complete with size_bytes = 0 (should be returned)
	iso5 := createTestISO()
	iso5.ID = "iso-missing-size-2"
	iso5.Name = "missing-size-2"
	iso5.Status = models.StatusComplete
	iso5.SizeBytes = 0
	if err := db.CreateISO(iso5); err != nil {
		t.Fatalf("Failed to create iso5: %v", err)
	}

	// Test ListISOsWithMissingSize
	isos, err := db.ListISOsWithMissingSize()
	if err != nil {
		t.Fatalf("ListISOsWithMissingSize() failed: %v", err)
	}

	// Should return exactly 2 ISOs (iso1 and iso5)
	if len(isos) != 2 {
		t.Errorf("Expected 2 ISOs with missing size, got %d", len(isos))
	}

	// Verify the returned ISOs are the correct ones
	foundIDs := make(map[string]bool)
	for _, iso := range isos {
		foundIDs[iso.ID] = true
		// All returned ISOs should be complete with size_bytes = 0
		if iso.Status != models.StatusComplete {
			t.Errorf("Returned ISO %s has status %s, expected complete", iso.ID, iso.Status)
		}
		if iso.SizeBytes != 0 {
			t.Errorf("Returned ISO %s has size_bytes %d, expected 0", iso.ID, iso.SizeBytes)
		}
	}

	if !foundIDs["iso-missing-size-1"] {
		t.Error("Expected iso-missing-size-1 to be returned")
	}
	if !foundIDs["iso-missing-size-2"] {
		t.Error("Expected iso-missing-size-2 to be returned")
	}
}

func TestListISOsWithMissingSize_Empty(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create only ISOs that should NOT be returned
	iso := createTestISO()
	iso.Status = models.StatusComplete
	iso.SizeBytes = 500000
	if err := db.CreateISO(iso); err != nil {
		t.Fatalf("Failed to create ISO: %v", err)
	}

	isos, err := db.ListISOsWithMissingSize()
	if err != nil {
		t.Fatalf("ListISOsWithMissingSize() failed: %v", err)
	}

	if len(isos) != 0 {
		t.Errorf("Expected 0 ISOs with missing size, got %d", len(isos))
	}
}
