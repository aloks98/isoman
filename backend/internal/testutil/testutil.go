package testutil

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"linux-iso-manager/internal/config"
	"linux-iso-manager/internal/db"
	"linux-iso-manager/internal/models"

	"github.com/google/uuid"
)

// TestEnv represents a complete test environment.
type TestEnv struct {
	DB      *db.DB
	Config  *config.Config
	Cleanup func()
	ISODir  string
	DBPath  string
	TmpDir  string
}

// SetupTestEnvironment creates a complete test environment with database, directories, and services.
func SetupTestEnvironment(t *testing.T) *TestEnv {
	t.Helper()

	// Create temp directory
	tmpDir := t.TempDir()
	isoDir := filepath.Join(tmpDir, "isos")
	dbDir := filepath.Join(tmpDir, "db")

	// Create directories
	if err := os.MkdirAll(isoDir, 0o755); err != nil {
		t.Fatalf("Failed to create iso directory: %v", err)
	}
	if err := os.MkdirAll(dbDir, 0o755); err != nil {
		t.Fatalf("Failed to create db directory: %v", err)
	}

	// Create test database
	dbPath := filepath.Join(dbDir, "test.db")
	cfg := config.Load()

	database, err := db.New(dbPath, &cfg.Database)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	cleanup := func() {
		database.Close()
		os.RemoveAll(tmpDir)
	}

	return &TestEnv{
		DB:      database,
		ISODir:  isoDir,
		DBPath:  dbPath,
		TmpDir:  tmpDir,
		Config:  cfg,
		Cleanup: cleanup,
	}
}

// SetupTestDB creates just a test database without other services.
func SetupTestDB(t *testing.T) (*db.DB, func()) {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := config.Load()

	database, err := db.New(dbPath, &cfg.Database)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	cleanup := func() {
		database.Close()
		os.RemoveAll(tmpDir)
	}

	return database, cleanup
}

// TestISO represents test ISO fixture data.
type TestISO struct {
	Name         string
	Version      string
	Arch         string
	Edition      string
	FileType     string
	DownloadURL  string
	ChecksumURL  string
	ChecksumType string
	Status       models.ISOStatus
}

// DefaultTestISO returns a default test ISO configuration.
func DefaultTestISO() *TestISO {
	return &TestISO{
		Name:         "alpine-linux",
		Version:      "3.19.1",
		Arch:         "x86_64",
		Edition:      "standard",
		FileType:     "iso",
		DownloadURL:  "https://example.com/alpine-3.19.1-x86_64.iso",
		ChecksumURL:  "https://example.com/alpine-3.19.1-x86_64.iso.sha256",
		ChecksumType: "sha256",
		Status:       models.StatusPending,
	}
}

// CreateTestISO creates a test ISO with default or custom values.
func CreateTestISO(overrides *TestISO) *models.ISO {
	base := DefaultTestISO()

	// Apply overrides
	if overrides != nil {
		if overrides.Name != "" {
			base.Name = overrides.Name
		}
		if overrides.Version != "" {
			base.Version = overrides.Version
		}
		if overrides.Arch != "" {
			base.Arch = overrides.Arch
		}
		if overrides.Edition != "" {
			base.Edition = overrides.Edition
		}
		if overrides.FileType != "" {
			base.FileType = overrides.FileType
		}
		if overrides.DownloadURL != "" {
			base.DownloadURL = overrides.DownloadURL
		}
		if overrides.ChecksumURL != "" {
			base.ChecksumURL = overrides.ChecksumURL
		}
		if overrides.ChecksumType != "" {
			base.ChecksumType = overrides.ChecksumType
		}
		if overrides.Status != "" {
			base.Status = overrides.Status
		}
	}

	iso := &models.ISO{
		ID:           uuid.New().String(),
		Name:         base.Name,
		Version:      base.Version,
		Arch:         base.Arch,
		Edition:      base.Edition,
		FileType:     base.FileType,
		Filename:     base.Name + "-" + base.Version + "-" + base.Arch + "." + base.FileType,
		FilePath:     filepath.Join(base.Name, base.Version, base.Arch, base.Name+"-"+base.Version+"-"+base.Arch+"."+base.FileType),
		DownloadLink: "/images/" + base.Name + "/" + base.Version + "/" + base.Arch + "/" + base.Name + "-" + base.Version + "-" + base.Arch + "." + base.FileType,
		SizeBytes:    1024 * 1024, // 1MB default
		Checksum:     "",
		ChecksumType: base.ChecksumType,
		DownloadURL:  base.DownloadURL,
		ChecksumURL:  base.ChecksumURL,
		Status:       base.Status,
		Progress:     0,
		ErrorMessage: "",
		CreatedAt:    time.Now(),
		CompletedAt:  nil,
	}

	return iso
}

// CreateAndInsertTestISO creates a test ISO and inserts it into the database.
func CreateAndInsertTestISO(t *testing.T, database *db.DB, overrides *TestISO) *models.ISO {
	t.Helper()

	iso := CreateTestISO(overrides)

	if err := database.CreateISO(iso); err != nil {
		t.Fatalf("Failed to insert test ISO: %v", err)
	}

	return iso
}

// AssertISOEqual asserts that two ISOs are equal (ignoring timestamps and IDs).
func AssertISOEqual(t *testing.T, expected, actual *models.ISO) {
	t.Helper()

	if expected.Name != actual.Name {
		t.Errorf("Name mismatch: expected %s, got %s", expected.Name, actual.Name)
	}
	if expected.Version != actual.Version {
		t.Errorf("Version mismatch: expected %s, got %s", expected.Version, actual.Version)
	}
	if expected.Arch != actual.Arch {
		t.Errorf("Arch mismatch: expected %s, got %s", expected.Arch, actual.Arch)
	}
	if expected.Edition != actual.Edition {
		t.Errorf("Edition mismatch: expected %s, got %s", expected.Edition, actual.Edition)
	}
	if expected.FileType != actual.FileType {
		t.Errorf("FileType mismatch: expected %s, got %s", expected.FileType, actual.FileType)
	}
	if expected.Status != actual.Status {
		t.Errorf("Status mismatch: expected %s, got %s", expected.Status, actual.Status)
	}
}

// AssertErrorContains asserts that an error contains a specific substring.
func AssertErrorContains(t *testing.T, err error, substring string) {
	t.Helper()

	if err == nil {
		t.Fatalf("Expected error containing %q, got nil", substring)
	}

	if !contains(err.Error(), substring) {
		t.Errorf("Expected error to contain %q, got: %v", substring, err)
	}
}

// AssertNoError asserts that an error is nil.
func AssertNoError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

// contains checks if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// CreateTestFile creates a test file with the given content.
func CreateTestFile(t *testing.T, dir, filename, content string) string {
	t.Helper()

	filePath := filepath.Join(dir, filename)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	return filePath
}

// FileExists checks if a file exists at the given path.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// StringContains checks if a string contains a substring (exported for use in tests).
func StringContains(s, substr string) bool {
	return contains(s, substr)
}

// AssertFileExists asserts that a file exists.
func AssertFileExists(t *testing.T, path string) {
	t.Helper()

	if !FileExists(path) {
		t.Errorf("Expected file to exist: %s", path)
	}
}

// AssertFileNotExists asserts that a file does not exist.
func AssertFileNotExists(t *testing.T, path string) {
	t.Helper()

	if FileExists(path) {
		t.Errorf("Expected file to not exist: %s", path)
	}
}
