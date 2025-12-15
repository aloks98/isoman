package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"linux-iso-manager/internal/models"
	"linux-iso-manager/internal/service"
	"linux-iso-manager/internal/testutil"

	"github.com/gin-gonic/gin"
)

// setupStatsHandlers creates test stats handlers with database.
func setupStatsHandlers(t *testing.T) (*StatsHandlers, *testutil.TestEnv) {
	t.Helper()

	env := testutil.SetupTestEnvironment(t)
	statsService := service.NewStatsService(env.DB)
	handlers := NewStatsHandlers(statsService)

	return handlers, env
}

func TestNewStatsHandlers(t *testing.T) {
	env := testutil.SetupTestEnvironment(t)
	defer env.Cleanup()

	statsService := service.NewStatsService(env.DB)
	handlers := NewStatsHandlers(statsService)

	if handlers == nil {
		t.Fatal("NewStatsHandlers() returned nil")
	}
}

func TestGetStats_EmptyDatabase(t *testing.T) {
	handlers, env := setupStatsHandlers(t)
	defer env.Cleanup()

	// Create test request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/stats", http.NoBody)

	// Call handler
	handlers.GetStats(c)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got: %d", w.Code)
	}

	var response APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !response.Success {
		t.Error("Expected success response")
	}

	// Verify stats structure
	data, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Response data should be a map")
	}

	// Check for expected fields
	if _, exists := data["total_isos"]; !exists {
		t.Error("Response should contain 'total_isos'")
	}
	if _, exists := data["completed_isos"]; !exists {
		t.Error("Response should contain 'completed_isos'")
	}
	if _, exists := data["failed_isos"]; !exists {
		t.Error("Response should contain 'failed_isos'")
	}
	if _, exists := data["pending_isos"]; !exists {
		t.Error("Response should contain 'pending_isos'")
	}
	if _, exists := data["total_size_bytes"]; !exists {
		t.Error("Response should contain 'total_size_bytes'")
	}
	if _, exists := data["total_downloads"]; !exists {
		t.Error("Response should contain 'total_downloads'")
	}
	if _, exists := data["bandwidth_saved"]; !exists {
		t.Error("Response should contain 'bandwidth_saved'")
	}
}

func TestGetStats_WithData(t *testing.T) {
	handlers, env := setupStatsHandlers(t)
	defer env.Cleanup()

	// Create test ISOs
	testutil.CreateAndInsertTestISO(t, env.DB, &testutil.TestISO{
		Name:    "alpine",
		Version: "3.19",
		Arch:    "x86_64",
		Status:  models.StatusComplete,
	})

	testutil.CreateAndInsertTestISO(t, env.DB, &testutil.TestISO{
		Name:    "ubuntu",
		Version: "24.04",
		Arch:    "x86_64",
		Status:  models.StatusFailed,
	})

	testutil.CreateAndInsertTestISO(t, env.DB, &testutil.TestISO{
		Name:    "debian",
		Version: "12",
		Arch:    "x86_64",
		Status:  models.StatusPending,
	})

	// Create test request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/stats", http.NoBody)

	// Call handler
	handlers.GetStats(c)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got: %d", w.Code)
	}

	var response APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !response.Success {
		t.Error("Expected success response")
	}

	data, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Response data should be a map")
	}

	// Verify counts
	totalISOs, _ := data["total_isos"].(float64)
	if totalISOs != 3 {
		t.Errorf("Expected total_isos 3, got %v", totalISOs)
	}

	completedISOs, _ := data["completed_isos"].(float64)
	if completedISOs != 1 {
		t.Errorf("Expected completed_isos 1, got %v", completedISOs)
	}

	failedISOs, _ := data["failed_isos"].(float64)
	if failedISOs != 1 {
		t.Errorf("Expected failed_isos 1, got %v", failedISOs)
	}

	pendingISOs, _ := data["pending_isos"].(float64)
	if pendingISOs != 1 {
		t.Errorf("Expected pending_isos 1, got %v", pendingISOs)
	}
}

func TestGetDownloadTrends_DefaultParams(t *testing.T) {
	handlers, env := setupStatsHandlers(t)
	defer env.Cleanup()

	// Create test request with no query params
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/stats/trends", http.NoBody)

	// Call handler
	handlers.GetDownloadTrends(c)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got: %d", w.Code)
	}

	var response APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !response.Success {
		t.Error("Expected success response")
	}

	data, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Response data should be a map")
	}

	// Default should be daily
	period, _ := data["period"].(string)
	if period != "daily" {
		t.Errorf("Expected default period 'daily', got '%s'", period)
	}
}

func TestGetDownloadTrends_DailyPeriod(t *testing.T) {
	handlers, env := setupStatsHandlers(t)
	defer env.Cleanup()

	// Create test request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/stats/trends?period=daily&days=7", http.NoBody)

	// Call handler
	handlers.GetDownloadTrends(c)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got: %d", w.Code)
	}

	var response APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !response.Success {
		t.Error("Expected success response")
	}

	data, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Response data should be a map")
	}

	period, _ := data["period"].(string)
	if period != "daily" {
		t.Errorf("Expected period 'daily', got '%s'", period)
	}
}

func TestGetDownloadTrends_WeeklyPeriod(t *testing.T) {
	handlers, env := setupStatsHandlers(t)
	defer env.Cleanup()

	// Create test request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/stats/trends?period=weekly&days=30", http.NoBody)

	// Call handler
	handlers.GetDownloadTrends(c)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got: %d", w.Code)
	}

	var response APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !response.Success {
		t.Error("Expected success response")
	}

	data, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Response data should be a map")
	}

	period, _ := data["period"].(string)
	if period != "weekly" {
		t.Errorf("Expected period 'weekly', got '%s'", period)
	}
}

func TestGetDownloadTrends_InvalidPeriod(t *testing.T) {
	handlers, env := setupStatsHandlers(t)
	defer env.Cleanup()

	// Create test request with invalid period
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/stats/trends?period=monthly", http.NoBody)

	// Call handler
	handlers.GetDownloadTrends(c)

	// Should still succeed, defaulting to daily
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got: %d", w.Code)
	}

	var response APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !response.Success {
		t.Error("Expected success response")
	}

	data, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Response data should be a map")
	}

	// Invalid period should default to daily
	period, _ := data["period"].(string)
	if period != "daily" {
		t.Errorf("Expected period to default to 'daily', got '%s'", period)
	}
}

func TestGetDownloadTrends_InvalidDays(t *testing.T) {
	handlers, env := setupStatsHandlers(t)
	defer env.Cleanup()

	testCases := []struct {
		name      string
		daysParam string
	}{
		{"negative days", "days=-1"},
		{"zero days", "days=0"},
		{"too many days", "days=500"},
		{"non-numeric days", "days=abc"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest("GET", "/api/stats/trends?"+tc.daysParam, http.NoBody)

			handlers.GetDownloadTrends(c)

			// Should still succeed, defaulting to 30 days
			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got: %d", w.Code)
			}

			var response APIResponse
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}

			if !response.Success {
				t.Error("Expected success response")
			}
		})
	}
}

func TestGetDownloadTrends_WithData(t *testing.T) {
	handlers, env := setupStatsHandlers(t)
	defer env.Cleanup()

	// Create test ISO and record download
	iso := testutil.CreateAndInsertTestISO(t, env.DB, &testutil.TestISO{
		Name:   "alpine",
		Status: models.StatusComplete,
	})

	statsService := service.NewStatsService(env.DB)
	statsService.RecordDownload(iso.ID)

	// Create test request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/stats/trends?period=daily&days=7", http.NoBody)

	// Call handler
	handlers.GetDownloadTrends(c)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got: %d", w.Code)
	}

	var response APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !response.Success {
		t.Error("Expected success response")
	}

	data, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Response data should be a map")
	}

	// Should have data array
	if _, exists := data["data"]; !exists {
		t.Error("Response should contain 'data' array")
	}
}

func TestGetStats_ISOsByArch(t *testing.T) {
	handlers, env := setupStatsHandlers(t)
	defer env.Cleanup()

	// Create ISOs with different architectures
	testutil.CreateAndInsertTestISO(t, env.DB, &testutil.TestISO{
		Name:    "alpine",
		Version: "3.19",
		Arch:    "x86_64",
		Status:  models.StatusComplete,
	})

	testutil.CreateAndInsertTestISO(t, env.DB, &testutil.TestISO{
		Name:    "ubuntu",
		Version: "24.04",
		Arch:    "aarch64",
		Status:  models.StatusComplete,
	})

	// Create test request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/stats", http.NoBody)

	// Call handler
	handlers.GetStats(c)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got: %d", w.Code)
	}

	var response APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	data, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Response data should be a map")
	}

	isosByArch, ok := data["isos_by_arch"].(map[string]interface{})
	if !ok {
		t.Fatal("Response should contain 'isos_by_arch' map")
	}

	x86Count, _ := isosByArch["x86_64"].(float64)
	if x86Count != 1 {
		t.Errorf("Expected 1 x86_64 ISO, got %v", x86Count)
	}

	aarch64Count, _ := isosByArch["aarch64"].(float64)
	if aarch64Count != 1 {
		t.Errorf("Expected 1 aarch64 ISO, got %v", aarch64Count)
	}
}

func TestGetStats_TopDownloaded(t *testing.T) {
	handlers, env := setupStatsHandlers(t)
	defer env.Cleanup()

	// Create ISO with downloads
	iso := testutil.CreateAndInsertTestISO(t, env.DB, &testutil.TestISO{
		Name:   "alpine",
		Status: models.StatusComplete,
	})

	statsService := service.NewStatsService(env.DB)
	for i := 0; i < 5; i++ {
		statsService.RecordDownload(iso.ID)
	}

	// Create test request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/stats", http.NoBody)

	// Call handler
	handlers.GetStats(c)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got: %d", w.Code)
	}

	var response APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	data, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Response data should be a map")
	}

	topDownloaded, ok := data["top_downloaded"].([]interface{})
	if !ok {
		t.Fatal("Response should contain 'top_downloaded' array")
	}

	if len(topDownloaded) < 1 {
		t.Error("Expected at least one top downloaded ISO")
	}

	// Verify first item has 5 downloads
	if len(topDownloaded) > 0 {
		first, ok := topDownloaded[0].(map[string]interface{})
		if !ok {
			t.Fatal("Top downloaded item should be a map")
		}

		downloadCount, _ := first["download_count"].(float64)
		if downloadCount != 5 {
			t.Errorf("Expected download_count 5, got %v", downloadCount)
		}
	}
}
