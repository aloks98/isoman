package service

import (
	"testing"
	"time"

	"linux-iso-manager/internal/models"
	"linux-iso-manager/internal/testutil"
)

func TestNewStatsService(t *testing.T) {
	env := testutil.SetupTestEnvironment(t)
	defer env.Cleanup()

	service := NewStatsService(env.DB)
	if service == nil {
		t.Fatal("NewStatsService() returned nil")
	}
}

func TestStatsService_GetStats_EmptyDatabase(t *testing.T) {
	env := testutil.SetupTestEnvironment(t)
	defer env.Cleanup()

	service := NewStatsService(env.DB)

	stats, err := service.GetStats()
	if err != nil {
		t.Fatalf("GetStats() failed: %v", err)
	}

	if stats == nil {
		t.Fatal("GetStats() returned nil stats")
	}

	if stats.TotalISOs != 0 {
		t.Errorf("Expected TotalISOs 0, got %d", stats.TotalISOs)
	}
}

func TestStatsService_GetStats_WithData(t *testing.T) {
	env := testutil.SetupTestEnvironment(t)
	defer env.Cleanup()

	// Create test ISOs
	iso1 := testutil.CreateAndInsertTestISO(t, env.DB, &testutil.TestISO{
		Name:    "alpine",
		Version: "3.19",
		Arch:    "x86_64",
		Status:  models.StatusComplete,
	})

	iso2 := testutil.CreateAndInsertTestISO(t, env.DB, &testutil.TestISO{
		Name:    "ubuntu",
		Version: "24.04",
		Arch:    "x86_64",
		Status:  models.StatusFailed,
	})

	_ = iso1
	_ = iso2

	service := NewStatsService(env.DB)

	stats, err := service.GetStats()
	if err != nil {
		t.Fatalf("GetStats() failed: %v", err)
	}

	if stats.TotalISOs != 2 {
		t.Errorf("Expected TotalISOs 2, got %d", stats.TotalISOs)
	}

	if stats.CompletedISOs != 1 {
		t.Errorf("Expected CompletedISOs 1, got %d", stats.CompletedISOs)
	}

	if stats.FailedISOs != 1 {
		t.Errorf("Expected FailedISOs 1, got %d", stats.FailedISOs)
	}
}

func TestStatsService_GetDownloadTrends_Daily(t *testing.T) {
	env := testutil.SetupTestEnvironment(t)
	defer env.Cleanup()

	service := NewStatsService(env.DB)

	trends, err := service.GetDownloadTrends("daily", 30)
	if err != nil {
		t.Fatalf("GetDownloadTrends() failed: %v", err)
	}

	if trends == nil {
		t.Fatal("GetDownloadTrends() returned nil")
	}

	if trends.Period != "daily" {
		t.Errorf("Expected period 'daily', got '%s'", trends.Period)
	}
}

func TestStatsService_GetDownloadTrends_Weekly(t *testing.T) {
	env := testutil.SetupTestEnvironment(t)
	defer env.Cleanup()

	service := NewStatsService(env.DB)

	trends, err := service.GetDownloadTrends("weekly", 30)
	if err != nil {
		t.Fatalf("GetDownloadTrends() failed: %v", err)
	}

	if trends.Period != "weekly" {
		t.Errorf("Expected period 'weekly', got '%s'", trends.Period)
	}
}

func TestStatsService_GetDownloadTrends_DefaultDays(t *testing.T) {
	env := testutil.SetupTestEnvironment(t)
	defer env.Cleanup()

	service := NewStatsService(env.DB)

	// Test default days for daily (should be 30)
	trends, err := service.GetDownloadTrends("daily", 0)
	if err != nil {
		t.Fatalf("GetDownloadTrends() failed: %v", err)
	}
	if trends == nil {
		t.Fatal("GetDownloadTrends() returned nil")
	}

	// Test default days for weekly (should be 84)
	trends, err = service.GetDownloadTrends("weekly", 0)
	if err != nil {
		t.Fatalf("GetDownloadTrends() weekly failed: %v", err)
	}
	if trends == nil {
		t.Fatal("GetDownloadTrends() returned nil for weekly")
	}
}

func TestStatsService_RecordDownload(t *testing.T) {
	env := testutil.SetupTestEnvironment(t)
	defer env.Cleanup()

	// Create test ISO
	iso := testutil.CreateAndInsertTestISO(t, env.DB, &testutil.TestISO{
		Name:   "alpine",
		Status: models.StatusComplete,
	})

	service := NewStatsService(env.DB)

	// Record download
	err := service.RecordDownload(iso.ID)
	if err != nil {
		t.Fatalf("RecordDownload() failed: %v", err)
	}

	// Verify download count was incremented
	stats, err := service.GetStats()
	if err != nil {
		t.Fatalf("GetStats() failed: %v", err)
	}

	if stats.TotalDownloads != 1 {
		t.Errorf("Expected TotalDownloads 1, got %d", stats.TotalDownloads)
	}
}

func TestStatsService_RecordDownload_Multiple(t *testing.T) {
	env := testutil.SetupTestEnvironment(t)
	defer env.Cleanup()

	// Create test ISO
	iso := testutil.CreateAndInsertTestISO(t, env.DB, &testutil.TestISO{
		Name:   "alpine",
		Status: models.StatusComplete,
	})

	service := NewStatsService(env.DB)

	// Record multiple downloads
	for i := 0; i < 5; i++ {
		err := service.RecordDownload(iso.ID)
		if err != nil {
			t.Fatalf("RecordDownload() failed on iteration %d: %v", i, err)
		}
	}

	// Verify download count
	stats, err := service.GetStats()
	if err != nil {
		t.Fatalf("GetStats() failed: %v", err)
	}

	if stats.TotalDownloads != 5 {
		t.Errorf("Expected TotalDownloads 5, got %d", stats.TotalDownloads)
	}
}

func TestStatsService_GetDownloadTrends_WithData(t *testing.T) {
	env := testutil.SetupTestEnvironment(t)
	defer env.Cleanup()

	// Create test ISO
	iso := testutil.CreateAndInsertTestISO(t, env.DB, &testutil.TestISO{
		Name:   "alpine",
		Status: models.StatusComplete,
	})

	service := NewStatsService(env.DB)

	// Record download
	err := service.RecordDownload(iso.ID)
	if err != nil {
		t.Fatalf("RecordDownload() failed: %v", err)
	}

	// Small delay to ensure event is recorded
	time.Sleep(10 * time.Millisecond)

	// Get trends
	trends, err := service.GetDownloadTrends("daily", 7)
	if err != nil {
		t.Fatalf("GetDownloadTrends() failed: %v", err)
	}

	// Should have at least one data point for today
	if len(trends.Data) == 0 {
		t.Error("Expected at least one data point")
	}
}

func TestStatsService_GetStats_BandwidthCalculation(t *testing.T) {
	env := testutil.SetupTestEnvironment(t)
	defer env.Cleanup()

	// Create test ISO with specific size
	iso := testutil.CreateTestISO(&testutil.TestISO{
		Name:   "alpine",
		Status: models.StatusComplete,
	})
	iso.SizeBytes = 1000000 // 1MB
	env.DB.CreateISO(iso)

	service := NewStatsService(env.DB)

	// Record 5 downloads
	for i := 0; i < 5; i++ {
		service.RecordDownload(iso.ID)
	}

	stats, err := service.GetStats()
	if err != nil {
		t.Fatalf("GetStats() failed: %v", err)
	}

	// Bandwidth saved = (5 - 1) * 1000000 = 4000000
	expectedBandwidthSaved := int64(4000000)
	if stats.BandwidthSaved != expectedBandwidthSaved {
		t.Errorf("Expected BandwidthSaved %d, got %d", expectedBandwidthSaved, stats.BandwidthSaved)
	}
}

func TestStatsService_GetStats_TopDownloaded(t *testing.T) {
	env := testutil.SetupTestEnvironment(t)
	defer env.Cleanup()

	service := NewStatsService(env.DB)

	// Create multiple ISOs with different download counts
	iso1 := testutil.CreateTestISO(&testutil.TestISO{
		Name:    "alpine",
		Version: "3.19",
		Status:  models.StatusComplete,
	})
	env.DB.CreateISO(iso1)

	iso2 := testutil.CreateTestISO(&testutil.TestISO{
		Name:    "ubuntu",
		Version: "24.04",
		Status:  models.StatusComplete,
	})
	env.DB.CreateISO(iso2)

	// ISO1 gets 3 downloads, ISO2 gets 5 downloads
	for i := 0; i < 3; i++ {
		service.RecordDownload(iso1.ID)
	}
	for i := 0; i < 5; i++ {
		service.RecordDownload(iso2.ID)
	}

	stats, err := service.GetStats()
	if err != nil {
		t.Fatalf("GetStats() failed: %v", err)
	}

	if len(stats.TopDownloaded) < 2 {
		t.Fatalf("Expected at least 2 top downloaded, got %d", len(stats.TopDownloaded))
	}

	// Ubuntu should be first (5 downloads)
	if stats.TopDownloaded[0].DownloadCount != 5 {
		t.Errorf("Expected top download count 5, got %d", stats.TopDownloaded[0].DownloadCount)
	}
}
