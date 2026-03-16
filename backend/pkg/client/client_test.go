package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// envelope builds a standard ISOMan API success response.
func envelope(data any) []byte {
	b, _ := json.Marshal(map[string]any{
		"success": true,
		"data":    data,
	})
	return b
}

// envelopeError builds a standard ISOMan API error response.
func envelopeError(code, message string) []byte {
	b, _ := json.Marshal(map[string]any{
		"success": false,
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
	return b
}

func sampleISO() map[string]any {
	return map[string]any{
		"id":             "test-id-123",
		"name":           "alpine",
		"version":        "3.19.1",
		"arch":           "x86_64",
		"edition":        "",
		"file_type":      "iso",
		"filename":       "alpine-3.19.1-x86_64.iso",
		"file_path":      "alpine/3.19.1/x86_64/alpine-3.19.1-x86_64.iso",
		"download_link":  "/images/alpine/3.19.1/x86_64/alpine-3.19.1-x86_64.iso",
		"size_bytes":     float64(204800),
		"checksum":       "abc123",
		"checksum_type":  "sha256",
		"download_url":   "https://example.com/alpine.iso",
		"checksum_url":   "https://example.com/alpine.iso.sha256",
		"status":         "complete",
		"progress":       float64(100),
		"error_message":  "",
		"created_at":     "2024-01-01T00:00:00Z",
		"completed_at":   "2024-01-01T00:05:00Z",
		"download_count": float64(5),
	}
}

func TestNewClient(t *testing.T) {
	c := NewClient("http://localhost:8080")
	if c.baseURL != "http://localhost:8080" {
		t.Errorf("baseURL = %q, want %q", c.baseURL, "http://localhost:8080")
	}
	if c.userAgent != "isoman-go-client" {
		t.Errorf("userAgent = %q, want %q", c.userAgent, "isoman-go-client")
	}
}

func TestNewClientTrailingSlash(t *testing.T) {
	c := NewClient("http://localhost:8080/")
	if c.baseURL != "http://localhost:8080" {
		t.Errorf("baseURL = %q, want trailing slash trimmed", c.baseURL)
	}
}

func TestNewClientOptions(t *testing.T) {
	hc := &http.Client{Timeout: 5 * time.Second}
	c := NewClient("http://localhost:8080",
		WithHTTPClient(hc),
		WithUserAgent("test-agent"),
	)
	if c.httpClient != hc {
		t.Error("WithHTTPClient did not set httpClient")
	}
	if c.userAgent != "test-agent" {
		t.Errorf("userAgent = %q, want %q", c.userAgent, "test-agent")
	}
}

func TestWithTimeout(t *testing.T) {
	c := NewClient("http://localhost:8080", WithTimeout(10*time.Second))
	if c.httpClient.Timeout != 10*time.Second {
		t.Errorf("Timeout = %v, want %v", c.httpClient.Timeout, 10*time.Second)
	}
}

func TestListISOs(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if !strings.HasPrefix(r.URL.Path, "/api/isos") {
			t.Errorf("path = %s, want /api/isos", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(envelope(map[string]any{
			"isos": []any{sampleISO()},
			"pagination": map[string]any{
				"page":        1,
				"page_size":   10,
				"total":       1,
				"total_pages": 1,
			},
		}))
	}))
	defer ts.Close()

	c := NewClient(ts.URL)
	result, err := c.ListISOs(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListISOs() error: %v", err)
	}
	if len(result.ISOs) != 1 {
		t.Fatalf("len(ISOs) = %d, want 1", len(result.ISOs))
	}
	if result.ISOs[0].ID != "test-id-123" {
		t.Errorf("ISO.ID = %q, want %q", result.ISOs[0].ID, "test-id-123")
	}
	if result.Pagination.Total != 1 {
		t.Errorf("Pagination.Total = %d, want 1", result.Pagination.Total)
	}
}

func TestListISOsWithOptions(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("page") != "2" {
			t.Errorf("page = %q, want %q", q.Get("page"), "2")
		}
		if q.Get("page_size") != "5" {
			t.Errorf("page_size = %q, want %q", q.Get("page_size"), "5")
		}
		if q.Get("sort_by") != "name" {
			t.Errorf("sort_by = %q, want %q", q.Get("sort_by"), "name")
		}
		if q.Get("sort_dir") != "asc" {
			t.Errorf("sort_dir = %q, want %q", q.Get("sort_dir"), "asc")
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(envelope(map[string]any{
			"isos":       []any{},
			"pagination": map[string]any{"page": 2, "page_size": 5, "total": 0, "total_pages": 0},
		}))
	}))
	defer ts.Close()

	c := NewClient(ts.URL)
	_, err := c.ListISOs(context.Background(), &ListISOsOptions{
		Page:     2,
		PageSize: 5,
		SortBy:   "name",
		SortDir:  "asc",
	})
	if err != nil {
		t.Fatalf("ListISOs() error: %v", err)
	}
}

func TestGetISO(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/isos/test-id-123" {
			t.Errorf("path = %s, want /api/isos/test-id-123", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(envelope(sampleISO()))
	}))
	defer ts.Close()

	c := NewClient(ts.URL)
	iso, err := c.GetISO(context.Background(), "test-id-123")
	if err != nil {
		t.Fatalf("GetISO() error: %v", err)
	}
	if iso.Name != "alpine" {
		t.Errorf("Name = %q, want %q", iso.Name, "alpine")
	}
	if iso.Status != StatusComplete {
		t.Errorf("Status = %q, want %q", iso.Status, StatusComplete)
	}
}

func TestGetISONotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write(envelopeError("NOT_FOUND", "ISO not found"))
	}))
	defer ts.Close()

	c := NewClient(ts.URL)
	_, err := c.GetISO(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("IsNotFound() = false, want true; err = %v", err)
	}
}

func TestCreateISO(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", r.Header.Get("Content-Type"))
		}

		var req CreateISORequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Name != "Alpine Linux" {
			t.Errorf("req.Name = %q, want %q", req.Name, "Alpine Linux")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		resp, _ := json.Marshal(map[string]any{
			"success": true,
			"data":    sampleISO(),
			"message": "ISO download queued successfully",
		})
		w.Write(resp)
	}))
	defer ts.Close()

	c := NewClient(ts.URL)
	iso, err := c.CreateISO(context.Background(), CreateISORequest{
		Name:         "Alpine Linux",
		Version:      "3.19.1",
		Arch:         "x86_64",
		DownloadURL:  "https://example.com/alpine.iso",
		ChecksumURL:  "https://example.com/alpine.iso.sha256",
		ChecksumType: "sha256",
	})
	if err != nil {
		t.Fatalf("CreateISO() error: %v", err)
	}
	if iso.ID != "test-id-123" {
		t.Errorf("ISO.ID = %q, want %q", iso.ID, "test-id-123")
	}
}

func TestCreateISOConflict(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		w.Write(envelopeError("CONFLICT", "ISO already exists"))
	}))
	defer ts.Close()

	c := NewClient(ts.URL)
	_, err := c.CreateISO(context.Background(), CreateISORequest{
		Name:        "Alpine Linux",
		Version:     "3.19.1",
		Arch:        "x86_64",
		DownloadURL: "https://example.com/alpine.iso",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsConflict(err) {
		t.Errorf("IsConflict() = false, want true; err = %v", err)
	}
}

func TestUpdateISO(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("method = %s, want PUT", r.Method)
		}

		var req UpdateISORequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Name == nil || *req.Name != "Alpine Linux Updated" {
			t.Errorf("req.Name = %v, want %q", req.Name, "Alpine Linux Updated")
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(envelope(sampleISO()))
	}))
	defer ts.Close()

	c := NewClient(ts.URL)
	name := "Alpine Linux Updated"
	_, err := c.UpdateISO(context.Background(), "test-id-123", UpdateISORequest{
		Name: &name,
	})
	if err != nil {
		t.Fatalf("UpdateISO() error: %v", err)
	}
}

func TestDeleteISO(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s, want DELETE", r.Method)
		}
		if r.URL.Path != "/api/isos/test-id-123" {
			t.Errorf("path = %s, want /api/isos/test-id-123", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		resp, _ := json.Marshal(map[string]any{
			"success": true,
			"message": "Resource deleted successfully",
		})
		w.Write(resp)
	}))
	defer ts.Close()

	c := NewClient(ts.URL)
	err := c.DeleteISO(context.Background(), "test-id-123")
	if err != nil {
		t.Fatalf("DeleteISO() error: %v", err)
	}
}

func TestRetryISO(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/api/isos/test-id-123/retry" {
			t.Errorf("path = %s, want /api/isos/test-id-123/retry", r.URL.Path)
		}

		iso := sampleISO()
		iso["status"] = "pending"
		iso["progress"] = float64(0)
		w.Header().Set("Content-Type", "application/json")
		w.Write(envelope(iso))
	}))
	defer ts.Close()

	c := NewClient(ts.URL)
	iso, err := c.RetryISO(context.Background(), "test-id-123")
	if err != nil {
		t.Fatalf("RetryISO() error: %v", err)
	}
	if iso.Status != StatusPending {
		t.Errorf("Status = %q, want %q", iso.Status, StatusPending)
	}
}

func TestRetryISOInvalidState(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(envelopeError("INVALID_STATE", "Cannot retry ISO with status: complete"))
	}))
	defer ts.Close()

	c := NewClient(ts.URL)
	_, err := c.RetryISO(context.Background(), "test-id-123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.Code != "INVALID_STATE" {
		t.Errorf("Code = %q, want %q", apiErr.Code, "INVALID_STATE")
	}
}

func TestGetStats(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/stats" {
			t.Errorf("path = %s, want /api/stats", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(envelope(map[string]any{
			"total_isos":       float64(10),
			"completed_isos":   float64(8),
			"failed_isos":      float64(1),
			"pending_isos":     float64(1),
			"total_size_bytes": float64(1073741824),
			"total_downloads":  float64(42),
			"bandwidth_saved":  float64(536870912),
			"isos_by_arch":     map[string]any{"x86_64": float64(7), "aarch64": float64(3)},
			"isos_by_edition":  map[string]any{"": float64(5), "server": float64(5)},
			"isos_by_status":   map[string]any{"complete": float64(8), "failed": float64(1), "pending": float64(1)},
			"top_downloaded":   []any{},
		}))
	}))
	defer ts.Close()

	c := NewClient(ts.URL)
	stats, err := c.GetStats(context.Background())
	if err != nil {
		t.Fatalf("GetStats() error: %v", err)
	}
	if stats.TotalISOs != 10 {
		t.Errorf("TotalISOs = %d, want 10", stats.TotalISOs)
	}
	if stats.CompletedISOs != 8 {
		t.Errorf("CompletedISOs = %d, want 8", stats.CompletedISOs)
	}
	if stats.ISOsByArch["x86_64"] != 7 {
		t.Errorf("ISOsByArch[x86_64] = %d, want 7", stats.ISOsByArch["x86_64"])
	}
}

func TestGetDownloadTrends(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/stats/trends" {
			t.Errorf("path = %s, want /api/stats/trends", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("period") != "weekly" {
			t.Errorf("period = %q, want %q", q.Get("period"), "weekly")
		}
		if q.Get("days") != "7" {
			t.Errorf("days = %q, want %q", q.Get("days"), "7")
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(envelope(map[string]any{
			"period": "weekly",
			"data": []any{
				map[string]any{"date": "2024-01-01", "count": float64(5)},
			},
		}))
	}))
	defer ts.Close()

	c := NewClient(ts.URL)
	trends, err := c.GetDownloadTrends(context.Background(), &DownloadTrendsOptions{
		Period: "weekly",
		Days:   7,
	})
	if err != nil {
		t.Fatalf("GetDownloadTrends() error: %v", err)
	}
	if trends.Period != "weekly" {
		t.Errorf("Period = %q, want %q", trends.Period, "weekly")
	}
	if len(trends.Data) != 1 {
		t.Fatalf("len(Data) = %d, want 1", len(trends.Data))
	}
	if trends.Data[0].Count != 5 {
		t.Errorf("Data[0].Count = %d, want 5", trends.Data[0].Count)
	}
}

func TestHealth(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			t.Errorf("path = %s, want /health", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(envelope(map[string]any{
			"status": "ok",
			"time":   "2024-01-01T00:00:00Z",
		}))
	}))
	defer ts.Close()

	c := NewClient(ts.URL)
	err := c.Health(context.Background())
	if err != nil {
		t.Fatalf("Health() error: %v", err)
	}
}

func TestHealthUnhealthy(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write(envelopeError("INTERNAL_ERROR", "database unreachable"))
	}))
	defer ts.Close()

	c := NewClient(ts.URL)
	err := c.Health(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDownloadFile(t *testing.T) {
	content := "fake-iso-content"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/images/alpine/3.19.1/x86_64/alpine.iso" {
			t.Errorf("path = %s, want /images/alpine/3.19.1/x86_64/alpine.iso", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write([]byte(content))
	}))
	defer ts.Close()

	c := NewClient(ts.URL)
	rc, err := c.DownloadFile(context.Background(), "alpine/3.19.1/x86_64/alpine.iso")
	if err != nil {
		t.Fatalf("DownloadFile() error: %v", err)
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("ReadAll() error: %v", err)
	}
	if string(data) != content {
		t.Errorf("body = %q, want %q", string(data), content)
	}
}

func TestDownloadFileNotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	c := NewClient(ts.URL)
	_, err := c.DownloadFile(context.Background(), "nonexistent.iso")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("IsNotFound() = false, want true")
	}
}

func TestAPIErrorFormat(t *testing.T) {
	err := &APIError{
		StatusCode: 404,
		Code:       "NOT_FOUND",
		Message:    "ISO not found",
	}
	expected := "isoman: NOT_FOUND: ISO not found"
	if err.Error() != expected {
		t.Errorf("Error() = %q, want %q", err.Error(), expected)
	}

	errWithDetails := &APIError{
		StatusCode: 400,
		Code:       "VALIDATION_FAILED",
		Message:    "Validation failed",
		Details:    "name is required",
	}
	expected = "isoman: VALIDATION_FAILED: Validation failed (name is required)"
	if errWithDetails.Error() != expected {
		t.Errorf("Error() = %q, want %q", errWithDetails.Error(), expected)
	}
}

func TestUserAgentHeader(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != "my-app/1.0" {
			t.Errorf("User-Agent = %q, want %q", r.Header.Get("User-Agent"), "my-app/1.0")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(envelope(map[string]any{"status": "ok"}))
	}))
	defer ts.Close()

	c := NewClient(ts.URL, WithUserAgent("my-app/1.0"))
	_ = c.Health(context.Background())
}
