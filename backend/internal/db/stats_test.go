package db

import (
	"testing"
	"time"

	"linux-iso-manager/internal/models"
)

func TestIncrementDownloadCount(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create test ISO
	iso := createTestISO()
	iso.Status = models.StatusComplete
	err := db.CreateISO(iso)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Verify initial download count is 0
	retrieved, err := db.GetISO(iso.ID)
	if err != nil {
		t.Fatalf("GetISO() failed: %v", err)
	}
	if retrieved.DownloadCount != 0 {
		t.Errorf("Expected initial DownloadCount 0, got %d", retrieved.DownloadCount)
	}

	// Increment download count
	err = db.IncrementDownloadCount(iso.ID)
	if err != nil {
		t.Fatalf("IncrementDownloadCount() failed: %v", err)
	}

	// Verify download count incremented
	retrieved, err = db.GetISO(iso.ID)
	if err != nil {
		t.Fatalf("GetISO() failed: %v", err)
	}
	if retrieved.DownloadCount != 1 {
		t.Errorf("Expected DownloadCount 1, got %d", retrieved.DownloadCount)
	}

	// Increment again
	err = db.IncrementDownloadCount(iso.ID)
	if err != nil {
		t.Fatalf("IncrementDownloadCount() second call failed: %v", err)
	}

	retrieved, err = db.GetISO(iso.ID)
	if err != nil {
		t.Fatalf("GetISO() failed: %v", err)
	}
	if retrieved.DownloadCount != 2 {
		t.Errorf("Expected DownloadCount 2, got %d", retrieved.DownloadCount)
	}
}

func TestRecordDownloadEvent(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create test ISO
	iso := createTestISO()
	iso.Status = models.StatusComplete
	err := db.CreateISO(iso)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Record download event
	downloadTime := time.Now()
	err = db.RecordDownloadEvent(iso.ID, downloadTime)
	if err != nil {
		t.Fatalf("RecordDownloadEvent() failed: %v", err)
	}

	// Record another download event
	err = db.RecordDownloadEvent(iso.ID, downloadTime.Add(time.Hour))
	if err != nil {
		t.Fatalf("RecordDownloadEvent() second call failed: %v", err)
	}
}

func TestGetStats_EmptyDatabase(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	stats, err := db.GetStats()
	if err != nil {
		t.Fatalf("GetStats() failed: %v", err)
	}

	if stats.TotalISOs != 0 {
		t.Errorf("Expected TotalISOs 0, got %d", stats.TotalISOs)
	}
	if stats.CompletedISOs != 0 {
		t.Errorf("Expected CompletedISOs 0, got %d", stats.CompletedISOs)
	}
	if stats.FailedISOs != 0 {
		t.Errorf("Expected FailedISOs 0, got %d", stats.FailedISOs)
	}
	if stats.PendingISOs != 0 {
		t.Errorf("Expected PendingISOs 0, got %d", stats.PendingISOs)
	}
	if stats.TotalSizeBytes != 0 {
		t.Errorf("Expected TotalSizeBytes 0, got %d", stats.TotalSizeBytes)
	}
	if stats.TotalDownloads != 0 {
		t.Errorf("Expected TotalDownloads 0, got %d", stats.TotalDownloads)
	}
	if stats.BandwidthSaved != 0 {
		t.Errorf("Expected BandwidthSaved 0, got %d", stats.BandwidthSaved)
	}
}

func TestGetStats_WithData(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create ISOs with different statuses
	completeISO1 := createTestISO()
	completeISO1.Name = "complete1"
	completeISO1.Filename = "complete1.iso"
	completeISO1.Status = models.StatusComplete
	completeISO1.SizeBytes = 1000
	completeISO1.Arch = "x86_64"
	completeISO1.Edition = "standard"
	db.CreateISO(completeISO1)

	completeISO2 := createTestISO()
	completeISO2.Name = "complete2"
	completeISO2.Filename = "complete2.iso"
	completeISO2.Status = models.StatusComplete
	completeISO2.SizeBytes = 2000
	completeISO2.Arch = "aarch64"
	completeISO2.Edition = "minimal"
	db.CreateISO(completeISO2)

	failedISO := createTestISO()
	failedISO.Name = "failed"
	failedISO.Filename = "failed.iso"
	failedISO.Status = models.StatusFailed
	failedISO.SizeBytes = 500
	failedISO.Arch = "x86_64"
	db.CreateISO(failedISO)

	pendingISO := createTestISO()
	pendingISO.Name = "pending"
	pendingISO.Filename = "pending.iso"
	pendingISO.Status = models.StatusPending
	pendingISO.Arch = "x86_64"
	db.CreateISO(pendingISO)

	// Get stats
	stats, err := db.GetStats()
	if err != nil {
		t.Fatalf("GetStats() failed: %v", err)
	}

	// Verify totals
	if stats.TotalISOs != 4 {
		t.Errorf("Expected TotalISOs 4, got %d", stats.TotalISOs)
	}
	if stats.CompletedISOs != 2 {
		t.Errorf("Expected CompletedISOs 2, got %d", stats.CompletedISOs)
	}
	if stats.FailedISOs != 1 {
		t.Errorf("Expected FailedISOs 1, got %d", stats.FailedISOs)
	}
	if stats.PendingISOs != 1 {
		t.Errorf("Expected PendingISOs 1, got %d", stats.PendingISOs)
	}

	// Total size should only count complete ISOs
	if stats.TotalSizeBytes != 3000 {
		t.Errorf("Expected TotalSizeBytes 3000, got %d", stats.TotalSizeBytes)
	}

	// Verify ISOs by arch
	if stats.ISOsByArch["x86_64"] != 3 {
		t.Errorf("Expected 3 x86_64 ISOs, got %d", stats.ISOsByArch["x86_64"])
	}
	if stats.ISOsByArch["aarch64"] != 1 {
		t.Errorf("Expected 1 aarch64 ISO, got %d", stats.ISOsByArch["aarch64"])
	}

	// Verify ISOs by edition (only non-empty)
	if stats.ISOsByEdition["standard"] != 1 {
		t.Errorf("Expected 1 standard edition, got %d", stats.ISOsByEdition["standard"])
	}
	if stats.ISOsByEdition["minimal"] != 1 {
		t.Errorf("Expected 1 minimal edition, got %d", stats.ISOsByEdition["minimal"])
	}

	// Verify ISOs by status
	if stats.ISOsByStatus["complete"] != 2 {
		t.Errorf("Expected 2 complete ISOs, got %d", stats.ISOsByStatus["complete"])
	}
	if stats.ISOsByStatus["failed"] != 1 {
		t.Errorf("Expected 1 failed ISO, got %d", stats.ISOsByStatus["failed"])
	}
	if stats.ISOsByStatus["pending"] != 1 {
		t.Errorf("Expected 1 pending ISO, got %d", stats.ISOsByStatus["pending"])
	}
}

func TestGetStats_BandwidthSaved(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create complete ISO with downloads
	iso := createTestISO()
	iso.Status = models.StatusComplete
	iso.SizeBytes = 1000
	db.CreateISO(iso)

	// Increment download count multiple times (simulating 5 downloads)
	for i := 0; i < 5; i++ {
		db.IncrementDownloadCount(iso.ID)
	}

	stats, err := db.GetStats()
	if err != nil {
		t.Fatalf("GetStats() failed: %v", err)
	}

	// Bandwidth saved = (download_count - 1) * size_bytes = (5 - 1) * 1000 = 4000
	if stats.BandwidthSaved != 4000 {
		t.Errorf("Expected BandwidthSaved 4000, got %d", stats.BandwidthSaved)
	}

	if stats.TotalDownloads != 5 {
		t.Errorf("Expected TotalDownloads 5, got %d", stats.TotalDownloads)
	}
}

func TestGetStats_TopDownloaded(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create ISOs with different download counts
	for i := 1; i <= 5; i++ {
		iso := createTestISO()
		iso.Name = "iso" + string(rune('A'+i-1))
		iso.Filename = iso.Name + ".iso"
		iso.Status = models.StatusComplete
		iso.SizeBytes = int64(i * 100)
		db.CreateISO(iso)

		// Set download count
		for j := 0; j < i*2; j++ {
			db.IncrementDownloadCount(iso.ID)
		}
	}

	stats, err := db.GetStats()
	if err != nil {
		t.Fatalf("GetStats() failed: %v", err)
	}

	// Should have 5 top downloaded
	if len(stats.TopDownloaded) != 5 {
		t.Errorf("Expected 5 top downloaded, got %d", len(stats.TopDownloaded))
	}

	// First should have highest download count (10)
	if len(stats.TopDownloaded) > 0 && stats.TopDownloaded[0].DownloadCount != 10 {
		t.Errorf("Expected top download count 10, got %d", stats.TopDownloaded[0].DownloadCount)
	}
}

func TestGetDownloadTrends_EmptyDatabase(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	trends, err := db.GetDownloadTrends("daily", 30)
	if err != nil {
		t.Fatalf("GetDownloadTrends() failed: %v", err)
	}

	if trends.Period != "daily" {
		t.Errorf("Expected period 'daily', got '%s'", trends.Period)
	}

	if len(trends.Data) != 0 {
		t.Errorf("Expected empty data, got %d items", len(trends.Data))
	}
}

func TestGetDownloadTrends_WithData(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create test ISO
	iso := createTestISO()
	iso.Status = models.StatusComplete
	db.CreateISO(iso)

	// Record download events at different times
	now := time.Now()
	db.RecordDownloadEvent(iso.ID, now)
	db.RecordDownloadEvent(iso.ID, now.Add(-time.Hour))
	db.RecordDownloadEvent(iso.ID, now.Add(-24*time.Hour))

	trends, err := db.GetDownloadTrends("daily", 7)
	if err != nil {
		t.Fatalf("GetDownloadTrends() failed: %v", err)
	}

	if trends.Period != "daily" {
		t.Errorf("Expected period 'daily', got '%s'", trends.Period)
	}

	// Should have at least one data point
	if len(trends.Data) == 0 {
		t.Error("Expected at least one data point")
	}
}

func TestGetDownloadTrends_Weekly(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create test ISO
	iso := createTestISO()
	iso.Status = models.StatusComplete
	db.CreateISO(iso)

	// Record download event
	db.RecordDownloadEvent(iso.ID, time.Now())

	trends, err := db.GetDownloadTrends("weekly", 30)
	if err != nil {
		t.Fatalf("GetDownloadTrends() failed: %v", err)
	}

	if trends.Period != "weekly" {
		t.Errorf("Expected period 'weekly', got '%s'", trends.Period)
	}
}

func TestGetISOByFilePath(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create test ISO with known file path
	iso := createTestISO()
	iso.FilePath = "alpine/3.19.1/x86_64/alpine-3.19.1-x86_64.iso"
	iso.Status = models.StatusComplete
	err := db.CreateISO(iso)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	t.Run("ExistingFilePath", func(t *testing.T) {
		retrieved, err := db.GetISOByFilePath("alpine/3.19.1/x86_64/alpine-3.19.1-x86_64.iso")
		if err != nil {
			t.Fatalf("GetISOByFilePath() failed: %v", err)
		}

		if retrieved == nil {
			t.Fatal("Expected ISO to be found, got nil")
		}

		if retrieved.ID != iso.ID {
			t.Errorf("Expected ID %s, got %s", iso.ID, retrieved.ID)
		}
	})

	t.Run("NonExistentFilePath", func(t *testing.T) {
		retrieved, err := db.GetISOByFilePath("nonexistent/path/file.iso")
		if err != nil {
			t.Fatalf("GetISOByFilePath() should not error: %v", err)
		}

		if retrieved != nil {
			t.Error("Expected nil for non-existent file path, got ISO")
		}
	})
}

func TestGetStats_DownloadingAndVerifyingStatus(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create ISOs with downloading and verifying statuses
	downloadingISO := createTestISO()
	downloadingISO.Name = "downloading"
	downloadingISO.Filename = "downloading.iso"
	downloadingISO.Status = models.StatusDownloading
	db.CreateISO(downloadingISO)

	verifyingISO := createTestISO()
	verifyingISO.Name = "verifying"
	verifyingISO.Filename = "verifying.iso"
	verifyingISO.Status = models.StatusVerifying
	db.CreateISO(verifyingISO)

	stats, err := db.GetStats()
	if err != nil {
		t.Fatalf("GetStats() failed: %v", err)
	}

	// Both should count as pending
	if stats.PendingISOs != 2 {
		t.Errorf("Expected PendingISOs 2, got %d", stats.PendingISOs)
	}

	if stats.ISOsByStatus["downloading"] != 1 {
		t.Errorf("Expected 1 downloading ISO, got %d", stats.ISOsByStatus["downloading"])
	}

	if stats.ISOsByStatus["verifying"] != 1 {
		t.Errorf("Expected 1 verifying ISO, got %d", stats.ISOsByStatus["verifying"])
	}
}
