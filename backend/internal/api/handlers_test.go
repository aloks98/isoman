package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"linux-iso-manager/internal/db"
	"linux-iso-manager/internal/download"
	"linux-iso-manager/internal/models"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// setupTestHandlers creates test handlers with database and manager
func setupTestHandlers(t *testing.T) (*Handlers, *db.DB, *download.Manager, string, func()) {
	// Create temp directory
	tmpDir := t.TempDir()
	isoDir := filepath.Join(tmpDir, "isos")
	os.MkdirAll(isoDir, 0755)

	// Create test database
	dbPath := filepath.Join(tmpDir, "test.db")
	database, err := db.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Create download manager (but don't start it for most tests)
	manager := download.NewManager(database, isoDir, 1)

	// Create handlers
	handlers := NewHandlers(database, manager, isoDir)

	cleanup := func() {
		manager.Stop()
		database.Close()
		os.RemoveAll(tmpDir)
	}

	return handlers, database, manager, isoDir, cleanup
}

// parseAPIResponse parses the uniform API response structure
func parseAPIResponse(t *testing.T, body []byte) *APIResponse {
	var response APIResponse
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("Failed to parse API response: %v", err)
	}
	return &response
}

// TestListISOsEmpty tests listing ISOs when database is empty
func TestListISOsEmpty(t *testing.T) {
	handlers, _, _, _, cleanup := setupTestHandlers(t)
	defer cleanup()

	// Create test request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/isos", nil)

	// Call handler
	handlers.ListISOs(c)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got: %d", w.Code)
	}

	response := parseAPIResponse(t, w.Body.Bytes())

	if !response.Success {
		t.Error("Expected success response")
	}

	data, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Response data should be a map")
	}

	isos, ok := data["isos"].([]interface{})
	if !ok {
		t.Fatal("Response data should contain 'isos' array")
	}

	if len(isos) != 0 {
		t.Errorf("Expected empty isos array, got: %d items", len(isos))
	}
}

// TestListISOsWithData tests listing ISOs with data
func TestListISOsWithData(t *testing.T) {
	handlers, database, _, _, cleanup := setupTestHandlers(t)
	defer cleanup()

	// Create test ISOs
	for i := 0; i < 3; i++ {
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
	}

	// Create test request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/isos", nil)

	// Call handler
	handlers.ListISOs(c)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got: %d", w.Code)
	}

	response := parseAPIResponse(t, w.Body.Bytes())

	if !response.Success {
		t.Error("Expected success response")
	}

	data, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Response data should be a map")
	}

	isos, ok := data["isos"].([]interface{})
	if !ok {
		t.Fatal("Response data should contain 'isos' array")
	}

	if len(isos) != 3 {
		t.Errorf("Expected 3 ISOs, got: %d", len(isos))
	}
}

// TestGetISOSuccess tests getting an ISO by ID
func TestGetISOSuccess(t *testing.T) {
	handlers, database, _, _, cleanup := setupTestHandlers(t)
	defer cleanup()

	// Create test ISO
	iso := &models.ISO{
		ID:          uuid.New().String(),
		Name:        "test",
		Version:     "1.0",
		Arch:        "x86_64",
		FileType:    "iso",
		DownloadURL: "http://example.com/test.iso",
		Status:      models.StatusPending,
		CreatedAt:   time.Now(),
	}
	iso.ComputeFields()
	database.CreateISO(iso)

	// Create test request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", fmt.Sprintf("/api/isos/%s", iso.ID), nil)
	c.Params = gin.Params{{Key: "id", Value: iso.ID}}

	// Call handler
	handlers.GetISO(c)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got: %d", w.Code)
	}

	apiResp := parseAPIResponse(t, w.Body.Bytes())

	if !apiResp.Success {
		t.Error("Expected success response")
	}

	// Extract ISO from response data
	dataBytes, _ := json.Marshal(apiResp.Data)
	var response models.ISO
	json.Unmarshal(dataBytes, &response)

	if response.ID != iso.ID {
		t.Errorf("Expected ID %s, got: %s", iso.ID, response.ID)
	}
}

// TestGetISONotFound tests getting non-existent ISO
func TestGetISONotFound(t *testing.T) {
	handlers, _, _, _, cleanup := setupTestHandlers(t)
	defer cleanup()

	// Create test request with non-existent ID
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	fakeID := uuid.New().String()
	c.Request, _ = http.NewRequest("GET", fmt.Sprintf("/api/isos/%s", fakeID), nil)
	c.Params = gin.Params{{Key: "id", Value: fakeID}}

	// Call handler
	handlers.GetISO(c)

	// Verify response
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got: %d", w.Code)
	}
}

// TestCreateISOSuccess tests creating a new ISO
func TestCreateISOSuccess(t *testing.T) {
	handlers, database, _, _, cleanup := setupTestHandlers(t)
	defer cleanup()

	// Create test request
	requestBody := models.CreateISORequest{
		Name:         "Alpine Linux",
		Version:      "3.19.1",
		Arch:         "x86_64",
		DownloadURL:  "http://example.com/alpine.iso",
		ChecksumURL:  "http://example.com/alpine.iso.sha256",
		ChecksumType: "sha256",
	}
	bodyJSON, _ := json.Marshal(requestBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/api/isos", bytes.NewBuffer(bodyJSON))
	c.Request.Header.Set("Content-Type", "application/json")

	// Call handler
	handlers.CreateISO(c)

	// Verify response
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got: %d, body: %s", w.Code, w.Body.String())
	}

	apiResp := parseAPIResponse(t, w.Body.Bytes())

	if !apiResp.Success {
		t.Error("Expected success response")
	}

	// Extract ISO from response data
	dataBytes, _ := json.Marshal(apiResp.Data)
	var response models.ISO
	json.Unmarshal(dataBytes, &response)

	// Verify computed fields
	if response.Name != "alpine-linux" {
		t.Errorf("Name should be normalized to 'alpine-linux', got: %s", response.Name)
	}

	if response.FileType != "iso" {
		t.Errorf("FileType should be auto-detected as 'iso', got: %s", response.FileType)
	}

	if response.Filename != "alpine-linux-3.19.1-x86_64.iso" {
		t.Errorf("Filename mismatch, got: %s", response.Filename)
	}

	if response.FilePath != "alpine-linux/3.19.1/x86_64/alpine-linux-3.19.1-x86_64.iso" {
		t.Errorf("FilePath mismatch, got: %s", response.FilePath)
	}

	if response.DownloadLink != "/images/alpine-linux/3.19.1/x86_64/alpine-linux-3.19.1-x86_64.iso" {
		t.Errorf("DownloadLink mismatch, got: %s", response.DownloadLink)
	}

	if response.Status != models.StatusPending {
		t.Errorf("Status should be 'pending', got: %s", response.Status)
	}

	// Verify ISO was created in database
	dbISO, err := database.GetISO(response.ID)
	if err != nil {
		t.Fatalf("ISO should be in database: %v", err)
	}

	if dbISO.Name != "alpine-linux" {
		t.Errorf("Database ISO name mismatch")
	}
}

// TestCreateISODuplicate tests creating duplicate ISO
func TestCreateISODuplicate(t *testing.T) {
	handlers, database, _, _, cleanup := setupTestHandlers(t)
	defer cleanup()

	// Create initial ISO
	iso := &models.ISO{
		ID:          uuid.New().String(),
		Name:        "alpine-linux",
		Version:     "3.19.1",
		Arch:        "x86_64",
		FileType:    "iso",
		DownloadURL: "http://example.com/alpine.iso",
		Status:      models.StatusPending,
		CreatedAt:   time.Now(),
	}
	iso.ComputeFields()
	database.CreateISO(iso)

	// Try to create duplicate
	requestBody := models.CreateISORequest{
		Name:        "Alpine Linux", // Will normalize to alpine-linux
		Version:     "3.19.1",
		Arch:        "x86_64",
		DownloadURL: "http://example.com/alpine.iso",
	}
	bodyJSON, _ := json.Marshal(requestBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/api/isos", bytes.NewBuffer(bodyJSON))
	c.Request.Header.Set("Content-Type", "application/json")

	// Call handler
	handlers.CreateISO(c)

	// Verify response
	if w.Code != http.StatusConflict {
		t.Errorf("Expected status 409 (Conflict), got: %d", w.Code)
	}

	apiResp := parseAPIResponse(t, w.Body.Bytes())

	if apiResp.Success {
		t.Error("Expected error response for duplicate")
	}

	if apiResp.Error == nil {
		t.Error("Expected error details in response")
	}

	data, ok := apiResp.Data.(map[string]interface{})
	if !ok || data["existing"] == nil {
		t.Error("Expected existing ISO in response data")
	}
}

// TestCreateISOInvalidRequest tests creating ISO with invalid request
func TestCreateISOInvalidRequest(t *testing.T) {
	handlers, _, _, _, cleanup := setupTestHandlers(t)
	defer cleanup()

	// Missing required fields
	requestBody := map[string]string{
		"name": "Test",
		// Missing version, arch, download_url
	}
	bodyJSON, _ := json.Marshal(requestBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/api/isos", bytes.NewBuffer(bodyJSON))
	c.Request.Header.Set("Content-Type", "application/json")

	// Call handler
	handlers.CreateISO(c)

	// Verify response
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got: %d", w.Code)
	}
}

// TestCreateISOUnsupportedFileType tests creating ISO with unsupported file type
func TestCreateISOUnsupportedFileType(t *testing.T) {
	handlers, _, _, _, cleanup := setupTestHandlers(t)
	defer cleanup()

	requestBody := models.CreateISORequest{
		Name:        "Test",
		Version:     "1.0",
		Arch:        "x86_64",
		DownloadURL: "http://example.com/file.txt", // .txt is not supported
	}
	bodyJSON, _ := json.Marshal(requestBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/api/isos", bytes.NewBuffer(bodyJSON))
	c.Request.Header.Set("Content-Type", "application/json")

	// Call handler
	handlers.CreateISO(c)

	// Verify response
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got: %d", w.Code)
	}
}

// TestDeleteISOSuccess tests deleting an ISO
func TestDeleteISOSuccess(t *testing.T) {
	handlers, database, _, isoDir, cleanup := setupTestHandlers(t)
	defer cleanup()

	// Create test ISO
	iso := &models.ISO{
		ID:          uuid.New().String(),
		Name:        "test",
		Version:     "1.0",
		Arch:        "x86_64",
		FileType:    "iso",
		DownloadURL: "http://example.com/test.iso",
		Status:      models.StatusComplete,
		CreatedAt:   time.Now(),
	}
	iso.ComputeFields()
	database.CreateISO(iso)

	// Create test file
	filePath := filepath.Join(isoDir, iso.FilePath)
	os.MkdirAll(filepath.Dir(filePath), 0755)
	os.WriteFile(filePath, []byte("test content"), 0644)

	// Create test request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("DELETE", fmt.Sprintf("/api/isos/%s", iso.ID), nil)
	c.Params = gin.Params{{Key: "id", Value: iso.ID}}

	// Call handler
	handlers.DeleteISO(c)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got: %d", w.Code)
	}

	apiResp := parseAPIResponse(t, w.Body.Bytes())

	if !apiResp.Success {
		t.Error("Expected success response")
	}

	// Verify ISO was deleted from database
	_, err := database.GetISO(iso.ID)
	if err == nil {
		t.Error("ISO should be deleted from database")
	}

	// Verify file was deleted
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Error("File should be deleted")
	}
}

// TestDeleteISONotFound tests deleting non-existent ISO
func TestDeleteISONotFound(t *testing.T) {
	handlers, _, _, _, cleanup := setupTestHandlers(t)
	defer cleanup()

	// Create test request with non-existent ID
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	fakeID := uuid.New().String()
	c.Request, _ = http.NewRequest("DELETE", fmt.Sprintf("/api/isos/%s", fakeID), nil)
	c.Params = gin.Params{{Key: "id", Value: fakeID}}

	// Call handler
	handlers.DeleteISO(c)

	// Verify response
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got: %d", w.Code)
	}
}

// TestRetryISOSuccess tests retrying a failed download
func TestRetryISOSuccess(t *testing.T) {
	handlers, database, _, _, cleanup := setupTestHandlers(t)
	defer cleanup()

	// Create failed ISO
	iso := &models.ISO{
		ID:           uuid.New().String(),
		Name:         "test",
		Version:      "1.0",
		Arch:         "x86_64",
		FileType:     "iso",
		DownloadURL:  "http://example.com/test.iso",
		Status:       models.StatusFailed,
		ErrorMessage: "Download failed",
		Progress:     50,
		CreatedAt:    time.Now(),
	}
	iso.ComputeFields()
	database.CreateISO(iso)

	// Create test request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", fmt.Sprintf("/api/isos/%s/retry", iso.ID), nil)
	c.Params = gin.Params{{Key: "id", Value: iso.ID}}

	// Call handler
	handlers.RetryISO(c)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got: %d", w.Code)
	}

	apiResp := parseAPIResponse(t, w.Body.Bytes())

	if !apiResp.Success {
		t.Error("Expected success response")
	}

	// Extract ISO from response data
	dataBytes, _ := json.Marshal(apiResp.Data)
	var response models.ISO
	json.Unmarshal(dataBytes, &response)

	// CRITICAL TEST: Verify status was reset to pending
	if response.Status != models.StatusPending {
		t.Errorf("Status should be reset to 'pending', got: %s", response.Status)
	}

	// Verify progress was reset
	if response.Progress != 0 {
		t.Errorf("Progress should be reset to 0, got: %d", response.Progress)
	}

	// Verify error message was cleared
	if response.ErrorMessage != "" {
		t.Errorf("ErrorMessage should be cleared, got: %s", response.ErrorMessage)
	}

	// Verify database was updated
	dbISO, _ := database.GetISO(iso.ID)
	if dbISO.Status != models.StatusPending {
		t.Errorf("Database status should be 'pending', got: %s", dbISO.Status)
	}
}

// TestRetryISONotFailed tests retrying a non-failed ISO
func TestRetryISONotFailed(t *testing.T) {
	handlers, database, _, _, cleanup := setupTestHandlers(t)
	defer cleanup()

	// Create completed ISO
	iso := &models.ISO{
		ID:          uuid.New().String(),
		Name:        "test",
		Version:     "1.0",
		Arch:        "x86_64",
		FileType:    "iso",
		DownloadURL: "http://example.com/test.iso",
		Status:      models.StatusComplete, // Not failed
		CreatedAt:   time.Now(),
	}
	iso.ComputeFields()
	database.CreateISO(iso)

	// Create test request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", fmt.Sprintf("/api/isos/%s/retry", iso.ID), nil)
	c.Params = gin.Params{{Key: "id", Value: iso.ID}}

	// Call handler
	handlers.RetryISO(c)

	// Verify response
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got: %d", w.Code)
	}
}

// TestRetryISONotFound tests retrying non-existent ISO
func TestRetryISONotFound(t *testing.T) {
	handlers, _, _, _, cleanup := setupTestHandlers(t)
	defer cleanup()

	// Create test request with non-existent ID
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	fakeID := uuid.New().String()
	c.Request, _ = http.NewRequest("POST", fmt.Sprintf("/api/isos/%s/retry", fakeID), nil)
	c.Params = gin.Params{{Key: "id", Value: fakeID}}

	// Call handler
	handlers.RetryISO(c)

	// Verify response
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got: %d", w.Code)
	}
}

// TestHealthCheck tests health check endpoint
func TestHealthCheck(t *testing.T) {
	handlers, _, _, _, cleanup := setupTestHandlers(t)
	defer cleanup()

	// Create test request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/health", nil)

	// Call handler
	handlers.HealthCheck(c)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got: %d", w.Code)
	}

	apiResp := parseAPIResponse(t, w.Body.Bytes())

	if !apiResp.Success {
		t.Error("Expected success response")
	}

	data, ok := apiResp.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Response data should be a map")
	}

	if data["status"] != "ok" {
		t.Errorf("Status should be 'ok', got: %v", data["status"])
	}

	if data["time"] == nil {
		t.Error("Response should include time")
	}
}

// TestCreateISOWithEdition tests creating ISO with edition field
func TestCreateISOWithEdition(t *testing.T) {
	handlers, _, _, _, cleanup := setupTestHandlers(t)
	defer cleanup()

	requestBody := models.CreateISORequest{
		Name:        "Ubuntu",
		Version:     "24.04",
		Arch:        "x86_64",
		Edition:     "desktop",
		DownloadURL: "http://example.com/ubuntu.iso",
	}
	bodyJSON, _ := json.Marshal(requestBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/api/isos", bytes.NewBuffer(bodyJSON))
	c.Request.Header.Set("Content-Type", "application/json")

	handlers.CreateISO(c)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got: %d", w.Code)
	}

	apiResp := parseAPIResponse(t, w.Body.Bytes())

	if !apiResp.Success {
		t.Error("Expected success response")
	}

	// Extract ISO from response data
	dataBytes, _ := json.Marshal(apiResp.Data)
	var response models.ISO
	json.Unmarshal(dataBytes, &response)

	if response.Edition != "desktop" {
		t.Errorf("Edition should be 'desktop', got: %s", response.Edition)
	}

	// Verify edition is in filename
	if response.Filename != "ubuntu-24.04-desktop-x86_64.iso" {
		t.Errorf("Filename should include edition, got: %s", response.Filename)
	}
}

// TestDeleteISOWithChecksumFile tests deleting ISO with checksum file cleanup
func TestDeleteISOWithChecksumFile(t *testing.T) {
	handlers, database, _, isoDir, cleanup := setupTestHandlers(t)
	defer cleanup()

	// Create test ISO with checksum
	iso := &models.ISO{
		ID:           uuid.New().String(),
		Name:         "test",
		Version:      "1.0",
		Arch:         "x86_64",
		FileType:     "iso",
		DownloadURL:  "http://example.com/test.iso",
		ChecksumURL:  "http://example.com/test.iso.sha256",
		ChecksumType: "sha256",
		Checksum:     "abc123",
		Status:       models.StatusComplete,
		CreatedAt:    time.Now(),
	}
	iso.ComputeFields()
	database.CreateISO(iso)

	// Create test file and checksum file
	filePath := filepath.Join(isoDir, iso.FilePath)
	os.MkdirAll(filepath.Dir(filePath), 0755)
	os.WriteFile(filePath, []byte("test content"), 0644)

	checksumPath := filePath + ".sha256"
	os.WriteFile(checksumPath, []byte("abc123  test-1.0-x86_64.iso\n"), 0644)

	// Verify checksum file exists before deletion
	if _, err := os.Stat(checksumPath); os.IsNotExist(err) {
		t.Fatal("Checksum file should exist before deletion")
	}

	// Create test request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("DELETE", fmt.Sprintf("/api/isos/%s", iso.ID), nil)
	c.Params = gin.Params{{Key: "id", Value: iso.ID}}

	// Call handler
	handlers.DeleteISO(c)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got: %d", w.Code)
	}

	// Parse response
	apiResp := parseAPIResponse(t, w.Body.Bytes())
	if !apiResp.Success {
		t.Error("Expected success to be true")
	}

	// Verify ISO file was deleted
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Error("ISO file should be deleted")
	}

	// Verify checksum file was deleted
	if _, err := os.Stat(checksumPath); !os.IsNotExist(err) {
		t.Error("Checksum file (.sha256) should be deleted")
	}
}

// TestDeleteISOWithMultipleChecksumTypes tests cleanup of different checksum types
func TestDeleteISOWithMultipleChecksumTypes(t *testing.T) {
	handlers, database, _, isoDir, cleanup := setupTestHandlers(t)
	defer cleanup()

	// Create test ISO
	iso := &models.ISO{
		ID:          uuid.New().String(),
		Name:        "test",
		Version:     "1.0",
		Arch:        "x86_64",
		FileType:    "iso",
		DownloadURL: "http://example.com/test.iso",
		Status:      models.StatusComplete,
		CreatedAt:   time.Now(),
	}
	iso.ComputeFields()
	database.CreateISO(iso)

	// Create test file
	filePath := filepath.Join(isoDir, iso.FilePath)
	os.MkdirAll(filepath.Dir(filePath), 0755)
	os.WriteFile(filePath, []byte("test content"), 0644)

	// Create multiple checksum files
	sha256Path := filePath + ".sha256"
	sha512Path := filePath + ".sha512"
	md5Path := filePath + ".md5"
	os.WriteFile(sha256Path, []byte("sha256sum"), 0644)
	os.WriteFile(sha512Path, []byte("sha512sum"), 0644)
	os.WriteFile(md5Path, []byte("md5sum"), 0644)

	// Delete ISO
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("DELETE", fmt.Sprintf("/api/isos/%s", iso.ID), nil)
	c.Params = gin.Params{{Key: "id", Value: iso.ID}}
	handlers.DeleteISO(c)

	// Verify all checksum files were deleted
	for _, path := range []string{sha256Path, sha512Path, md5Path} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("Checksum file should be deleted: %s", path)
		}
	}
}
