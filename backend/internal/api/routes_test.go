package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"linux-iso-manager/internal/download"
	"linux-iso-manager/internal/service"
	"linux-iso-manager/internal/testutil"
	"linux-iso-manager/internal/ws"
)

func TestSetupRoutes(t *testing.T) {
	env := testutil.SetupTestEnvironment(t)
	defer env.Cleanup()

	manager := download.NewManager(env.DB, env.ISODir, 1)
	defer manager.Stop()

	isoService := service.NewISOService(env.DB, manager, env.ISODir)

	wsHub := ws.NewHub()
	router := SetupRoutes(isoService, env.ISODir, wsHub, env.Config)

	if router == nil {
		t.Fatal("Expected router to be created, got nil")
	}
}

func TestAPIRoutes(t *testing.T) {
	env := testutil.SetupTestEnvironment(t)
	defer env.Cleanup()

	manager := download.NewManager(env.DB, env.ISODir, 1)
	defer manager.Stop()
	isoService := service.NewISOService(env.DB, manager, env.ISODir)

	wsHub := ws.NewHub()
	router := SetupRoutes(isoService, env.ISODir, wsHub, env.Config)

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
	}{
		{
			name:       "GET /api/isos - should be registered",
			method:     http.MethodGet,
			path:       "/api/isos",
			wantStatus: http.StatusOK,
		},
		{
			name:       "GET /api/isos/:id - should be registered",
			method:     http.MethodGet,
			path:       "/api/isos/test-id",
			wantStatus: http.StatusNotFound, // ID doesn't exist, but route exists
		},
		{
			name:       "POST /api/isos - should be registered",
			method:     http.MethodPost,
			path:       "/api/isos",
			wantStatus: http.StatusBadRequest, // No body, but route exists
		},
		{
			name:       "DELETE /api/isos/:id - should be registered",
			method:     http.MethodDelete,
			path:       "/api/isos/test-id",
			wantStatus: http.StatusNotFound, // ID doesn't exist, but route exists
		},
		{
			name:       "POST /api/isos/:id/retry - should be registered",
			method:     http.MethodPost,
			path:       "/api/isos/test-id/retry",
			wantStatus: http.StatusNotFound, // ID doesn't exist, but route exists
		},
		{
			name:       "GET /health - should be registered",
			method:     http.MethodGet,
			path:       "/health",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, http.NoBody)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestAPIRouteNotFound(t *testing.T) {
	env := testutil.SetupTestEnvironment(t)
	defer env.Cleanup()

	manager := download.NewManager(env.DB, env.ISODir, 1)
	defer manager.Stop()
	isoService := service.NewISOService(env.DB, manager, env.ISODir)

	wsHub := ws.NewHub()
	router := SetupRoutes(isoService, env.ISODir, wsHub, env.Config)

	req := httptest.NewRequest(http.MethodGet, "/api/nonexistent", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d for non-existent API route, got %d", http.StatusNotFound, w.Code)
	}

	// Verify error response format
	if w.Header().Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("Expected JSON content type, got %s", w.Header().Get("Content-Type"))
	}
}

func TestCORSConfiguration(t *testing.T) {
	env := testutil.SetupTestEnvironment(t)
	defer env.Cleanup()

	manager := download.NewManager(env.DB, env.ISODir, 1)
	defer manager.Stop()
	isoService := service.NewISOService(env.DB, manager, env.ISODir)

	wsHub := ws.NewHub()
	router := SetupRoutes(isoService, env.ISODir, wsHub, env.Config)

	// Test CORS preflight request
	req := httptest.NewRequest(http.MethodOptions, "/api/isos", http.NoBody)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check for CORS headers
	if w.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Error("Expected Access-Control-Allow-Origin header to be set")
	}

	if w.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("Expected Access-Control-Allow-Methods header to be set")
	}
}

func TestNoRouteHandler(t *testing.T) {
	env := testutil.SetupTestEnvironment(t)
	defer env.Cleanup()

	// Create frontend directory and index.html
	frontendPath := "./ui/dist"
	os.MkdirAll(frontendPath, 0o755)
	defer os.RemoveAll("./ui")

	indexContent := "<html><body>Test SPA</body></html>"
	err := os.WriteFile(filepath.Join(frontendPath, "index.html"), []byte(indexContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to create test index.html: %v", err)
	}

	manager := download.NewManager(env.DB, env.ISODir, 1)
	defer manager.Stop()
	isoService := service.NewISOService(env.DB, manager, env.ISODir)

	wsHub := ws.NewHub()
	router := SetupRoutes(isoService, env.ISODir, wsHub, env.Config)

	tests := []struct {
		name           string
		path           string
		wantStatus     int
		wantServeIndex bool
	}{
		{
			name:           "Root path should serve index.html",
			path:           "/",
			wantStatus:     http.StatusOK,
			wantServeIndex: true,
		},
		{
			name:           "SPA route should serve index.html",
			path:           "/dashboard",
			wantStatus:     http.StatusOK,
			wantServeIndex: true,
		},
		{
			name:           "API route should return 404",
			path:           "/api/unknown",
			wantStatus:     http.StatusNotFound,
			wantServeIndex: false,
		},
		{
			name:           "WebSocket route should return 400 without upgrade",
			path:           "/ws",
			wantStatus:     http.StatusBadRequest, // No websocket upgrade headers
			wantServeIndex: false,
		},
		{
			name:           "Images route should return 404",
			path:           "/images/test",
			wantStatus:     http.StatusNotFound,
			wantServeIndex: false,
		},
		{
			name:           "Health route should not serve index.html",
			path:           "/health",
			wantStatus:     http.StatusOK,
			wantServeIndex: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, http.NoBody)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}

			if tt.wantServeIndex {
				if w.Body.String() != indexContent {
					t.Error("Expected index.html content to be served")
				}
			}
		})
	}
}

func TestHealthEndpoint(t *testing.T) {
	env := testutil.SetupTestEnvironment(t)
	defer env.Cleanup()

	manager := download.NewManager(env.DB, env.ISODir, 1)
	defer manager.Stop()
	isoService := service.NewISOService(env.DB, manager, env.ISODir)

	wsHub := ws.NewHub()
	router := SetupRoutes(isoService, env.ISODir, wsHub, env.Config)

	req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Check content type
	if w.Header().Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("Expected JSON content type, got %s", w.Header().Get("Content-Type"))
	}

	// Check body contains expected fields
	body := w.Body.String()
	if !testutil.StringContains(body, "status") {
		t.Error("Expected health response to contain 'status' field")
	}
}

func TestImagesRoute(t *testing.T) {
	env := testutil.SetupTestEnvironment(t)
	defer env.Cleanup()

	manager := download.NewManager(env.DB, env.ISODir, 1)
	defer manager.Stop()
	isoService := service.NewISOService(env.DB, manager, env.ISODir)

	wsHub := ws.NewHub()
	router := SetupRoutes(isoService, env.ISODir, wsHub, env.Config)

	// Test directory listing
	req := httptest.NewRequest(http.MethodGet, "/images/", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return 200 with directory listing HTML
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d for /images/, got %d", http.StatusOK, w.Code)
	}

	// Should have HTML content type
	contentType := w.Header().Get("Content-Type")
	if !testutil.StringContains(contentType, "text/html") {
		t.Errorf("Expected HTML content type, got %s", contentType)
	}
}
