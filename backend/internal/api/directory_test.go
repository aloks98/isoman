package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// setupTestDirectory creates a test directory structure.
func setupTestDirectory(t *testing.T) (string, func()) {
	tmpDir := t.TempDir()

	// Create test directory structure
	os.MkdirAll(filepath.Join(tmpDir, "alpine", "3.19.1", "x86_64"), 0o755)
	os.MkdirAll(filepath.Join(tmpDir, "ubuntu", "24.04", "x86_64"), 0o755)

	// Create test files
	os.WriteFile(filepath.Join(tmpDir, "alpine", "3.19.1", "x86_64", "alpine.iso"), []byte("test alpine content"), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "ubuntu", "24.04", "x86_64", "ubuntu.iso"), []byte("test ubuntu content"), 0o644)

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

// TestDirectoryHandlerListRoot tests listing root directory.
func TestDirectoryHandlerListRoot(t *testing.T) {
	isoDir, cleanup := setupTestDirectory(t)
	defer cleanup()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/images/", http.NoBody)
	c.Params = gin.Params{{Key: "filepath", Value: "/"}}

	handler := DirectoryHandler(isoDir)
	handler(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got: %d", w.Code)
	}

	body := w.Body.String()

	// Check for HTML structure
	if !strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("Response should be HTML")
	}

	if !strings.Contains(body, "Index of /images/") {
		t.Error("Response should contain directory title")
	}

	// Check for directory listings
	if !strings.Contains(body, "alpine") {
		t.Error("Response should contain 'alpine' directory")
	}

	if !strings.Contains(body, "ubuntu") {
		t.Error("Response should contain 'ubuntu' directory")
	}
}

// TestDirectoryHandlerListSubdirectory tests listing a subdirectory.
func TestDirectoryHandlerListSubdirectory(t *testing.T) {
	isoDir, cleanup := setupTestDirectory(t)
	defer cleanup()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/images/alpine", http.NoBody)
	c.Params = gin.Params{{Key: "filepath", Value: "/alpine"}}

	handler := DirectoryHandler(isoDir)
	handler(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got: %d", w.Code)
	}

	body := w.Body.String()

	if !strings.Contains(body, "3.19.1") {
		t.Error("Response should contain '3.19.1' subdirectory")
	}

	// Should have parent directory link
	if !strings.Contains(body, "Parent Directory") {
		t.Error("Response should contain parent directory link")
	}
}

// TestDirectoryHandlerServeFile tests serving a file.
func TestDirectoryHandlerServeFile(t *testing.T) {
	isoDir, cleanup := setupTestDirectory(t)
	defer cleanup()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/images/alpine/3.19.1/x86_64/alpine.iso", http.NoBody)
	c.Params = gin.Params{{Key: "filepath", Value: "/alpine/3.19.1/x86_64/alpine.iso"}}

	handler := DirectoryHandler(isoDir)
	handler(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got: %d", w.Code)
	}

	body := w.Body.String()
	if body != "test alpine content" {
		t.Errorf("Expected file content, got: %s", body)
	}
}

// TestDirectoryHandlerFileNotFound tests 404 for non-existent file.
func TestDirectoryHandlerFileNotFound(t *testing.T) {
	isoDir, cleanup := setupTestDirectory(t)
	defer cleanup()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/images/nonexistent.iso", http.NoBody)
	c.Params = gin.Params{{Key: "filepath", Value: "/nonexistent.iso"}}

	handler := DirectoryHandler(isoDir)
	handler(c)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got: %d", w.Code)
	}
}

// TestDirectoryHandlerDirectoryNotFound tests 404 for non-existent directory.
func TestDirectoryHandlerDirectoryNotFound(t *testing.T) {
	isoDir, cleanup := setupTestDirectory(t)
	defer cleanup()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/images/nonexistent/", http.NoBody)
	c.Params = gin.Params{{Key: "filepath", Value: "/nonexistent/"}}

	handler := DirectoryHandler(isoDir)
	handler(c)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got: %d", w.Code)
	}
}

// TestDirectoryHandlerHiddenFilesSkipped tests that hidden files are not shown.
func TestDirectoryHandlerHiddenFilesSkipped(t *testing.T) {
	isoDir, cleanup := setupTestDirectory(t)
	defer cleanup()

	// Create hidden file/directory
	os.WriteFile(filepath.Join(isoDir, ".hidden"), []byte("hidden content"), 0o644)
	os.MkdirAll(filepath.Join(isoDir, ".tmp"), 0o755)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/images/", http.NoBody)
	c.Params = gin.Params{{Key: "filepath", Value: "/"}}

	handler := DirectoryHandler(isoDir)
	handler(c)

	body := w.Body.String()

	if strings.Contains(body, ".hidden") {
		t.Error("Hidden files should not be listed")
	}

	if strings.Contains(body, ".tmp") {
		t.Error("Hidden directories should not be listed")
	}
}

// TestFormatSize tests the formatSize function.
func TestFormatSize(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{100, "100 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1572864, "1.5 MB"},
		{1073741824, "1.0 GB"},
		{1099511627776, "1.0 TB"},
		{217055232, "207.0 MB"}, // Alpine ISO size
	}

	for _, tt := range tests {
		result := formatSize(tt.bytes)
		if result != tt.expected {
			t.Errorf("formatSize(%d) = %s, want %s", tt.bytes, result, tt.expected)
		}
	}
}

// TestDirectoryHandlerSorting tests that directories come before files.
func TestDirectoryHandlerSorting(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files and directories with names that would sort differently
	os.MkdirAll(filepath.Join(tmpDir, "z-directory"), 0o755)
	os.WriteFile(filepath.Join(tmpDir, "a-file.iso"), []byte("content"), 0o644)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/images/", http.NoBody)
	c.Params = gin.Params{{Key: "filepath", Value: "/"}}

	handler := DirectoryHandler(tmpDir)
	handler(c)

	body := w.Body.String()

	// Directory should appear before file in HTML
	dirIndex := strings.Index(body, "z-directory")
	fileIndex := strings.Index(body, "a-file.iso")

	if dirIndex == -1 {
		t.Error("Directory should be in listing")
	}
	if fileIndex == -1 {
		t.Error("File should be in listing")
	}
	if dirIndex > fileIndex {
		t.Error("Directory should appear before file in listing")
	}
}

// TestDirectoryHandlerEmptyDirectory tests listing an empty directory.
func TestDirectoryHandlerEmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/images/", http.NoBody)
	c.Params = gin.Params{{Key: "filepath", Value: "/"}}

	handler := DirectoryHandler(tmpDir)
	handler(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got: %d", w.Code)
	}

	body := w.Body.String()

	// Should still have HTML structure
	if !strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("Response should be HTML")
	}

	if !strings.Contains(body, "Index of /images/") {
		t.Error("Response should contain directory title")
	}
}

// TestDirectoryHandlerContentType tests that HTML is served with correct content type.
func TestDirectoryHandlerContentType(t *testing.T) {
	tmpDir := t.TempDir()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/images/", http.NoBody)
	c.Params = gin.Params{{Key: "filepath", Value: "/"}}

	handler := DirectoryHandler(tmpDir)
	handler(c)

	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Errorf("Expected Content-Type to contain 'text/html', got: %s", contentType)
	}
}

// TestDirectoryHandlerNestedPath tests deeply nested directory paths.
func TestDirectoryHandlerNestedPath(t *testing.T) {
	isoDir, cleanup := setupTestDirectory(t)
	defer cleanup()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/images/alpine/3.19.1/x86_64", http.NoBody)
	c.Params = gin.Params{{Key: "filepath", Value: "/alpine/3.19.1/x86_64"}}

	handler := DirectoryHandler(isoDir)
	handler(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got: %d", w.Code)
	}

	body := w.Body.String()

	if !strings.Contains(body, "alpine.iso") {
		t.Error("Response should contain alpine.iso file")
	}

	// Should show correct path in title
	if !strings.Contains(body, "alpine/3.19.1/x86_64") {
		t.Error("Response should show nested path in title")
	}
}

// TestDirectoryHandlerFileTypeIcons tests that different file types show correct icons.
func TestDirectoryHandlerFileTypeIcons(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files of different types
	os.WriteFile(filepath.Join(tmpDir, "alpine.iso"), []byte("iso content"), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "debian.img"), []byte("img content"), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "ubuntu.iso.sha256"), []byte("sha256 checksum"), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "fedora.iso.sha512"), []byte("sha512 checksum"), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "arch.iso.md5"), []byte("md5 checksum"), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("text file"), 0o644)
	os.MkdirAll(filepath.Join(tmpDir, "folder"), 0o755)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/images/", http.NoBody)
	c.Params = gin.Params{{Key: "filepath", Value: "/"}}

	handler := DirectoryHandler(tmpDir)
	handler(c)

	body := w.Body.String()

	// Test that response contains files
	if !strings.Contains(body, "alpine.iso") {
		t.Error("Response should contain alpine.iso")
	}

	// Test for purple disc icon (ISO/IMG files)
	if !strings.Contains(body, "bg-purple-100") {
		t.Error("HTML should contain purple disc icon class (bg-purple-100)")
	}

	// Test for green shield icon (checksum files)
	if !strings.Contains(body, "bg-green-100") {
		t.Error("HTML should contain green shield icon class (bg-green-100)")
	}

	// Test for SHA-256/SHA-512/MD5 checksum labels
	if !strings.Contains(body, "SHA-256 Checksum") {
		t.Error("HTML should contain 'SHA-256 Checksum' label")
	}
	if !strings.Contains(body, "SHA-512 Checksum") {
		t.Error("HTML should contain 'SHA-512 Checksum' label")
	}
	if !strings.Contains(body, "MD5 Checksum") {
		t.Error("HTML should contain 'MD5 Checksum' label")
	}

	// Test for gray generic file icon
	if !strings.Contains(body, "bg-slate-100") {
		t.Error("HTML should contain gray icon class (bg-slate-100) for generic files")
	}

	// Test for blue folder icon
	if !strings.Contains(body, "bg-blue-100") {
		t.Error("HTML should contain blue folder icon class (bg-blue-100)")
	}
}

// TestDirectoryHandlerDirectorySizeDisplay tests that directories show "-" for size.
func TestDirectoryHandlerDirectorySizeDisplay(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directory and file
	os.MkdirAll(filepath.Join(tmpDir, "testdir"), 0o755)
	os.WriteFile(filepath.Join(tmpDir, "testfile.iso"), []byte("content"), 0o644)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/images/", http.NoBody)
	c.Params = gin.Params{{Key: "filepath", Value: "/"}}

	handler := DirectoryHandler(tmpDir)
	handler(c)

	body := w.Body.String()

	// Directory should be in listing
	if !strings.Contains(body, "testdir") {
		t.Fatal("Directory should be in listing")
	}

	// File should be in listing
	if !strings.Contains(body, "testfile.iso") {
		t.Fatal("File should be in listing")
	}

	// Extract the FileInfo structures by parsing lines
	// Directory entries should show "-" for size, files should show actual bytes
	// We can verify this by checking that "-" appears in a size span (for directory)
	// and that "7 B" appears for the file
	if !strings.Contains(body, ">-<") {
		t.Error("HTML should contain '-' for directory size")
	}

	// File should show "7 B" for 7 bytes
	if !strings.Contains(body, "7 B") {
		t.Error("HTML should show '7 B' for file size")
	}
}
