package service

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"linux-iso-manager/internal/download"
	"linux-iso-manager/internal/models"
	"linux-iso-manager/internal/testutil"
)

func TestNormalizeName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "basic name",
			input: "Alpine Linux",
			want:  "alpine-linux",
		},
		{
			name:  "multiple spaces",
			input: "Alpine  Linux  Server",
			want:  "alpine-linux-server",
		},
		{
			name:  "leading and trailing spaces",
			input: "  Alpine Linux  ",
			want:  "alpine-linux",
		},
		{
			name:  "uppercase",
			input: "ALPINE LINUX",
			want:  "alpine-linux",
		},
		{
			name:  "mixed case",
			input: "AlPiNe LiNuX",
			want:  "alpine-linux",
		},
		{
			name:  "special characters",
			input: "Alpine@Linux#2024",
			want:  "alpinelinux2024",
		},
		{
			name:  "unicode characters",
			input: "Alpine Linux 中文",
			want:  "alpine-linux",
		},
		{
			name:  "emoji",
			input: "Alpine Linux 🐧",
			want:  "alpine-linux",
		},
		{
			name:  "path traversal attempt",
			input: "../../../etc/passwd",
			want:  "etc-passwd",
		},
		{
			name:  "consecutive hyphens",
			input: "Alpine---Linux",
			want:  "alpine-linux",
		},
		{
			name:  "hyphens at edges",
			input: "-Alpine-Linux-",
			want:  "alpine-linux",
		},
		{
			name:  "only special characters",
			input: "@#$%^&*()",
			want:  "",
		},
		{
			name:  "numbers",
			input: "Rocky Linux 8",
			want:  "rocky-linux-8",
		},
		{
			name:  "dots and slashes",
			input: "Alpine.Linux/3.19",
			want:  "alpine-linux-3-19",
		},
		{
			name:  "null bytes",
			input: "Alpine\x00Linux",
			want:  "alpinelinux",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeName(tt.input)
			if got != tt.want {
				t.Errorf("NormalizeName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestDetectFileType(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    string
		wantErr bool
	}{
		{
			name:    "iso extension",
			url:     "https://example.com/alpine.iso",
			want:    "iso",
			wantErr: false,
		},
		{
			name:    "ISO uppercase",
			url:     "https://example.com/alpine.ISO",
			want:    "iso",
			wantErr: false,
		},
		{
			name:    "qcow2 extension",
			url:     "https://example.com/ubuntu.qcow2",
			want:    "qcow2",
			wantErr: false,
		},
		{
			name:    "img extension",
			url:     "https://example.com/disk.img",
			want:    "img",
			wantErr: false,
		},
		{
			name:    "no extension",
			url:     "https://example.com/alpine",
			want:    "",
			wantErr: true,
		},
		{
			name:    "unsupported extension",
			url:     "https://example.com/file.txt",
			want:    "",
			wantErr: true,
		},
		{
			name:    "multiple dots",
			url:     "https://example.com/alpine.3.19.iso",
			want:    "iso",
			wantErr: false,
		},
		{
			name:    "compressed file (not supported directly)",
			url:     "https://example.com/alpine.iso.gz",
			want:    "",
			wantErr: true,
		},
		{
			name:    "query parameters",
			url:     "https://example.com/alpine.iso?download=true",
			want:    "",
			wantErr: true,
		},
		{
			name:    "fragment",
			url:     "https://example.com/alpine.iso#section",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DetectFileType(tt.url)

			if tt.wantErr {
				if err == nil {
					t.Errorf("DetectFileType(%q) expected error, got nil", tt.url)
				}
				return
			}

			if err != nil {
				t.Errorf("DetectFileType(%q) unexpected error: %v", tt.url, err)
				return
			}

			if got != tt.want {
				t.Errorf("DetectFileType(%q) = %q, want %q", tt.url, got, tt.want)
			}
		})
	}
}

func TestGenerateFilename(t *testing.T) {
	tests := []struct {
		name     string
		isoName  string
		version  string
		edition  string
		arch     string
		fileType string
		want     string
	}{
		{
			name:     "with edition",
			isoName:  "alpine",
			version:  "3.19.1",
			edition:  "standard",
			arch:     "x86_64",
			fileType: "iso",
			want:     "alpine-3.19.1-standard-x86_64.iso",
		},
		{
			name:     "without edition",
			isoName:  "alpine",
			version:  "3.19.1",
			edition:  "",
			arch:     "x86_64",
			fileType: "iso",
			want:     "alpine-3.19.1-x86_64.iso",
		},
		{
			name:     "qcow2 file",
			isoName:  "ubuntu",
			version:  "24.04",
			edition:  "server",
			arch:     "amd64",
			fileType: "qcow2",
			want:     "ubuntu-24.04-server-amd64.qcow2",
		},
		{
			name:     "edge case - empty edition",
			isoName:  "test",
			version:  "1.0",
			edition:  "",
			arch:     "arm64",
			fileType: "img",
			want:     "test-1.0-arm64.img",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateFilename(tt.isoName, tt.version, tt.edition, tt.arch, tt.fileType)
			if got != tt.want {
				t.Errorf("GenerateFilename() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGenerateFilePath(t *testing.T) {
	tests := []struct {
		name     string
		isoName  string
		version  string
		arch     string
		filename string
		want     string
	}{
		{
			name:     "basic path",
			isoName:  "alpine",
			version:  "3.19.1",
			arch:     "x86_64",
			filename: "alpine-3.19.1-x86_64.iso",
			want:     "alpine/3.19.1/x86_64/alpine-3.19.1-x86_64.iso",
		},
		{
			name:     "nested version",
			isoName:  "ubuntu",
			version:  "24.04.1",
			arch:     "amd64",
			filename: "ubuntu-24.04.1-amd64.iso",
			want:     "ubuntu/24.04.1/amd64/ubuntu-24.04.1-amd64.iso",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateFilePath(tt.isoName, tt.version, tt.arch, tt.filename)
			// Normalize path separators for cross-platform testing
			got = strings.ReplaceAll(got, "\\", "/")
			if got != tt.want {
				t.Errorf("GenerateFilePath() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGenerateDownloadLink(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		want     string
	}{
		{
			name:     "basic path",
			filePath: "alpine/3.19.1/x86_64/alpine-3.19.1-x86_64.iso",
			want:     "/images/alpine/3.19.1/x86_64/alpine-3.19.1-x86_64.iso",
		},
		{
			name:     "windows path separators",
			filePath: "alpine\\3.19.1\\x86_64\\alpine-3.19.1-x86_64.iso",
			want:     "/images/alpine/3.19.1/x86_64/alpine-3.19.1-x86_64.iso",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateDownloadLink(tt.filePath)
			if got != tt.want {
				t.Errorf("GenerateDownloadLink() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestComputeFields(t *testing.T) {
	tests := []struct {
		name             string
		iso              *models.ISO
		wantName         string
		wantFilename     string
		wantDownloadLink string
	}{
		{
			name: "basic ISO",
			iso: &models.ISO{
				Name:     "Alpine Linux",
				Version:  "3.19.1",
				Arch:     "x86_64",
				Edition:  "standard",
				FileType: "iso",
			},
			wantName:         "alpine-linux",
			wantFilename:     "alpine-linux-3.19.1-standard-x86_64.iso",
			wantDownloadLink: "/images/alpine-linux/3.19.1/x86_64/alpine-linux-3.19.1-standard-x86_64.iso",
		},
		{
			name: "ISO without edition",
			iso: &models.ISO{
				Name:     "Ubuntu",
				Version:  "24.04",
				Arch:     "amd64",
				Edition:  "",
				FileType: "iso",
			},
			wantName:         "ubuntu",
			wantFilename:     "ubuntu-24.04-amd64.iso",
			wantDownloadLink: "/images/ubuntu/24.04/amd64/ubuntu-24.04-amd64.iso",
		},
		{
			name: "ISO with special characters in name",
			iso: &models.ISO{
				Name:     "Alpine@Linux#2024",
				Version:  "3.19",
				Arch:     "x86_64",
				Edition:  "",
				FileType: "iso",
			},
			wantName:         "alpinelinux2024",
			wantFilename:     "alpinelinux2024-3.19-x86_64.iso",
			wantDownloadLink: "/images/alpinelinux2024/3.19/x86_64/alpinelinux2024-3.19-x86_64.iso",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ComputeFields(tt.iso)

			if tt.iso.Name != tt.wantName {
				t.Errorf("ComputeFields() Name = %q, want %q", tt.iso.Name, tt.wantName)
			}
			if tt.iso.Filename != tt.wantFilename {
				t.Errorf("ComputeFields() Filename = %q, want %q", tt.iso.Filename, tt.wantFilename)
			}

			// Normalize path separators
			gotLink := strings.ReplaceAll(tt.iso.DownloadLink, "\\", "/")
			if gotLink != tt.wantDownloadLink {
				t.Errorf("ComputeFields() DownloadLink = %q, want %q", gotLink, tt.wantDownloadLink)
			}
		})
	}
}

// setupTestISOService creates a test ISO service with database and manager.
func setupTestISOService(t *testing.T) (*ISOService, *testutil.TestEnv) {
	t.Helper()

	env := testutil.SetupTestEnvironment(t)
	manager := download.NewManager(env.DB, env.ISODir, 1)
	service := NewISOService(env.DB, manager, env.ISODir)

	return service, env
}

func TestNewISOService(t *testing.T) {
	env := testutil.SetupTestEnvironment(t)
	defer env.Cleanup()

	manager := download.NewManager(env.DB, env.ISODir, 1)
	service := NewISOService(env.DB, manager, env.ISODir)

	if service == nil {
		t.Fatal("NewISOService() returned nil")
	}
}

func TestISOService_CreateISO(t *testing.T) {
	service, env := setupTestISOService(t)
	defer env.Cleanup()

	t.Run("Success", func(t *testing.T) {
		req := CreateISORequest{
			Name:         "Alpine Linux",
			Version:      "3.19.1",
			Arch:         "x86_64",
			Edition:      "",
			DownloadURL:  "https://example.com/alpine.iso",
			ChecksumURL:  "https://example.com/alpine.iso.sha256",
			ChecksumType: "sha256",
		}

		iso, err := service.CreateISO(req)
		if err != nil {
			t.Fatalf("CreateISO() failed: %v", err)
		}

		if iso.Name != "alpine-linux" {
			t.Errorf("Name should be normalized to 'alpine-linux', got: %s", iso.Name)
		}
		if iso.Status != models.StatusPending {
			t.Errorf("Status should be 'pending', got: %s", iso.Status)
		}
		if iso.FileType != "iso" {
			t.Errorf("FileType should be 'iso', got: %s", iso.FileType)
		}
	})

	t.Run("DuplicateISO", func(t *testing.T) {
		// Create first ISO
		req := CreateISORequest{
			Name:        "Ubuntu",
			Version:     "24.04",
			Arch:        "x86_64",
			DownloadURL: "https://example.com/ubuntu.iso",
		}

		_, err := service.CreateISO(req)
		if err != nil {
			t.Fatalf("First CreateISO() failed: %v", err)
		}

		// Try to create duplicate
		_, err = service.CreateISO(req)
		if err == nil {
			t.Fatal("Expected error for duplicate ISO")
		}

		var existsErr *ISOAlreadyExistsError
		if !errors.As(err, &existsErr) {
			t.Errorf("Expected ISOAlreadyExistsError, got: %T", err)
		}
	})

	t.Run("UnsupportedFileType", func(t *testing.T) {
		req := CreateISORequest{
			Name:        "Test",
			Version:     "1.0",
			Arch:        "x86_64",
			DownloadURL: "https://example.com/file.txt",
		}

		_, err := service.CreateISO(req)
		if err == nil {
			t.Fatal("Expected error for unsupported file type")
		}
	})

	t.Run("DefaultChecksumType", func(t *testing.T) {
		req := CreateISORequest{
			Name:        "Debian",
			Version:     "12",
			Arch:        "x86_64",
			DownloadURL: "https://example.com/debian.iso",
			ChecksumURL: "https://example.com/debian.iso.sha256",
			// ChecksumType not specified
		}

		iso, err := service.CreateISO(req)
		if err != nil {
			t.Fatalf("CreateISO() failed: %v", err)
		}

		if iso.ChecksumType != "sha256" {
			t.Errorf("ChecksumType should default to 'sha256', got: %s", iso.ChecksumType)
		}
	})
}

func TestISOService_GetISO(t *testing.T) {
	service, env := setupTestISOService(t)
	defer env.Cleanup()

	t.Run("ExistingISO", func(t *testing.T) {
		// Create ISO directly in DB
		iso := testutil.CreateAndInsertTestISO(t, env.DB, &testutil.TestISO{
			Name:    "alpine",
			Version: "3.20",
			Status:  models.StatusComplete,
		})

		retrieved, err := service.GetISO(iso.ID)
		if err != nil {
			t.Fatalf("GetISO() failed: %v", err)
		}

		if retrieved.ID != iso.ID {
			t.Errorf("ID mismatch: expected %s, got %s", iso.ID, retrieved.ID)
		}
	})

	t.Run("NonExistentISO", func(t *testing.T) {
		_, err := service.GetISO("nonexistent-id")
		if err == nil {
			t.Fatal("Expected error for non-existent ISO")
		}
	})
}

func TestISOService_ListISOs(t *testing.T) {
	service, env := setupTestISOService(t)
	defer env.Cleanup()

	t.Run("EmptyDatabase", func(t *testing.T) {
		isos, err := service.ListISOs()
		if err != nil {
			t.Fatalf("ListISOs() failed: %v", err)
		}

		if len(isos) != 0 {
			t.Errorf("Expected 0 ISOs, got %d", len(isos))
		}
	})

	t.Run("WithData", func(t *testing.T) {
		// Create test ISOs
		for i := 0; i < 3; i++ {
			testutil.CreateAndInsertTestISO(t, env.DB, &testutil.TestISO{
				Name:    "test",
				Version: strings.Repeat("v", i+1), // Unique versions
				Status:  models.StatusComplete,
			})
		}

		isos, err := service.ListISOs()
		if err != nil {
			t.Fatalf("ListISOs() failed: %v", err)
		}

		if len(isos) != 3 {
			t.Errorf("Expected 3 ISOs, got %d", len(isos))
		}
	})
}

func TestISOService_DeleteISO(t *testing.T) {
	service, env := setupTestISOService(t)
	defer env.Cleanup()

	t.Run("ExistingISO", func(t *testing.T) {
		iso := testutil.CreateAndInsertTestISO(t, env.DB, &testutil.TestISO{
			Name:   "deleteme",
			Status: models.StatusComplete,
		})

		err := service.DeleteISO(iso.ID)
		if err != nil {
			t.Fatalf("DeleteISO() failed: %v", err)
		}

		// Verify ISO was deleted
		_, err = service.GetISO(iso.ID)
		if err == nil {
			t.Error("Expected error when getting deleted ISO")
		}
	})

	t.Run("NonExistentISO", func(t *testing.T) {
		err := service.DeleteISO("nonexistent-id")
		if err == nil {
			t.Fatal("Expected error for non-existent ISO")
		}
	})
}

func TestISOService_RetryISO(t *testing.T) {
	service, env := setupTestISOService(t)
	defer env.Cleanup()

	t.Run("FailedISO", func(t *testing.T) {
		iso := testutil.CreateAndInsertTestISO(t, env.DB, &testutil.TestISO{
			Name:   "retry-test",
			Status: models.StatusFailed,
		})

		retried, err := service.RetryISO(iso.ID)
		if err != nil {
			t.Fatalf("RetryISO() failed: %v", err)
		}

		if retried.Status != models.StatusPending {
			t.Errorf("Status should be 'pending', got: %s", retried.Status)
		}
		if retried.Progress != 0 {
			t.Errorf("Progress should be 0, got: %d", retried.Progress)
		}
		if retried.ErrorMessage != "" {
			t.Errorf("ErrorMessage should be empty, got: %s", retried.ErrorMessage)
		}
	})

	t.Run("CompleteISO_ShouldFail", func(t *testing.T) {
		iso := testutil.CreateAndInsertTestISO(t, env.DB, &testutil.TestISO{
			Name:   "complete-iso",
			Status: models.StatusComplete,
		})

		_, err := service.RetryISO(iso.ID)
		if err == nil {
			t.Fatal("Expected error when retrying complete ISO")
		}

		var invalidStateErr *InvalidStateError
		if !errors.As(err, &invalidStateErr) {
			t.Errorf("Expected InvalidStateError, got: %T", err)
		}
	})

	t.Run("NonExistentISO", func(t *testing.T) {
		_, err := service.RetryISO("nonexistent-id")
		if err == nil {
			t.Fatal("Expected error for non-existent ISO")
		}
	})
}

func TestISOService_UpdateISO(t *testing.T) {
	service, env := setupTestISOService(t)
	defer env.Cleanup()

	t.Run("UpdateFailedISO", func(t *testing.T) {
		iso := testutil.CreateAndInsertTestISO(t, env.DB, &testutil.TestISO{
			Name:   "failed-iso",
			Status: models.StatusFailed,
		})

		newName := "Updated Name"
		newVersion := "2.0"
		req := models.UpdateISORequest{
			Name:    &newName,
			Version: &newVersion,
		}

		updated, err := service.UpdateISO(iso.ID, req)
		if err != nil {
			t.Fatalf("UpdateISO() failed: %v", err)
		}

		if updated.Name != "updated-name" {
			t.Errorf("Name should be normalized to 'updated-name', got: %s", updated.Name)
		}
		if updated.Version != "2.0" {
			t.Errorf("Version should be '2.0', got: %s", updated.Version)
		}
		// Should reset to pending
		if updated.Status != models.StatusPending {
			t.Errorf("Status should be 'pending', got: %s", updated.Status)
		}
	})

	t.Run("UpdateCompleteISO_MetadataOnly", func(t *testing.T) {
		iso := testutil.CreateAndInsertTestISO(t, env.DB, &testutil.TestISO{
			Name:    "complete-iso",
			Version: "1.0",
			Status:  models.StatusComplete,
		})

		// Create file for move operation
		filePath := filepath.Join(env.ISODir, iso.FilePath)
		os.MkdirAll(filepath.Dir(filePath), 0o755)
		os.WriteFile(filePath, []byte("test"), 0o644)

		newEdition := "server"
		req := models.UpdateISORequest{
			Edition: &newEdition,
		}

		updated, err := service.UpdateISO(iso.ID, req)
		if err != nil {
			t.Fatalf("UpdateISO() failed: %v", err)
		}

		if updated.Edition != "server" {
			t.Errorf("Edition should be 'server', got: %s", updated.Edition)
		}
		// Status should remain complete
		if updated.Status != models.StatusComplete {
			t.Errorf("Status should remain 'complete', got: %s", updated.Status)
		}
	})

	t.Run("UpdateDownloadingISO_ShouldFail", func(t *testing.T) {
		iso := testutil.CreateAndInsertTestISO(t, env.DB, &testutil.TestISO{
			Name:   "downloading-iso",
			Status: models.StatusDownloading,
		})

		newName := "test"
		req := models.UpdateISORequest{
			Name: &newName,
		}

		_, err := service.UpdateISO(iso.ID, req)
		if err == nil {
			t.Fatal("Expected error when updating downloading ISO")
		}

		var invalidStateErr *InvalidStateError
		if !errors.As(err, &invalidStateErr) {
			t.Errorf("Expected InvalidStateError, got: %T", err)
		}
	})

	t.Run("UpdateCompleteISO_URLChange_ShouldFail", func(t *testing.T) {
		iso := testutil.CreateAndInsertTestISO(t, env.DB, &testutil.TestISO{
			Name:   "complete-iso-url",
			Status: models.StatusComplete,
		})

		newURL := "https://example.com/new.iso"
		req := models.UpdateISORequest{
			DownloadURL: &newURL,
		}

		_, err := service.UpdateISO(iso.ID, req)
		if err == nil {
			t.Fatal("Expected error when changing URL of complete ISO")
		}

		var invalidStateErr *InvalidStateError
		if !errors.As(err, &invalidStateErr) {
			t.Errorf("Expected InvalidStateError, got: %T", err)
		}
	})

	t.Run("NonExistentISO", func(t *testing.T) {
		newName := "test"
		req := models.UpdateISORequest{
			Name: &newName,
		}

		_, err := service.UpdateISO("nonexistent-id", req)
		if err == nil {
			t.Fatal("Expected error for non-existent ISO")
		}
	})
}

func TestISOAlreadyExistsError(t *testing.T) {
	err := &ISOAlreadyExistsError{
		ExistingISO: &models.ISO{ID: "test-id"},
	}

	if err.Error() != "ISO already exists" {
		t.Errorf("Error() = %q, want 'ISO already exists'", err.Error())
	}
}

func TestInvalidStateError(t *testing.T) {
	err := &InvalidStateError{
		CurrentStatus: "downloading",
		Message:       "Cannot edit while downloading",
	}

	errStr := err.Error()
	if !strings.Contains(errStr, "downloading") {
		t.Errorf("Error() should contain status, got: %s", errStr)
	}
	if !strings.Contains(errStr, "Cannot edit while downloading") {
		t.Errorf("Error() should contain message, got: %s", errStr)
	}
}
