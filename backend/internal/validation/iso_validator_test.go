package validation

import (
	"strings"
	"testing"
)

func TestValidateISOCreateRequest(t *testing.T) {
	tests := []struct {
		req     *ISOCreateRequest
		name    string
		errMsg  string
		wantErr bool
	}{
		{
			name: "valid request",
			req: &ISOCreateRequest{
				Name:         "Alpine Linux",
				Version:      "3.19.1",
				Arch:         "x86_64",
				Edition:      "standard",
				DownloadURL:  "https://example.com/alpine.iso",
				ChecksumURL:  "https://example.com/alpine.iso.sha256",
				ChecksumType: "sha256",
			},
			wantErr: false,
		},
		{
			name: "valid request without edition",
			req: &ISOCreateRequest{
				Name:        "Ubuntu",
				Version:     "24.04",
				Arch:        "amd64",
				DownloadURL: "https://example.com/ubuntu.iso",
			},
			wantErr: false,
		},
		{
			name: "valid request without checksum",
			req: &ISOCreateRequest{
				Name:        "Debian",
				Version:     "12",
				Arch:        "x86_64",
				DownloadURL: "https://example.com/debian.iso",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			req: &ISOCreateRequest{
				Name:        "",
				Version:     "1.0",
				Arch:        "x86_64",
				DownloadURL: "https://example.com/test.iso",
			},
			wantErr: true,
			errMsg:  "name",
		},
		{
			name: "whitespace-only name",
			req: &ISOCreateRequest{
				Name:        "   ",
				Version:     "1.0",
				Arch:        "x86_64",
				DownloadURL: "https://example.com/test.iso",
			},
			wantErr: true,
			errMsg:  "name",
		},
		{
			name: "empty version",
			req: &ISOCreateRequest{
				Name:        "Test",
				Version:     "",
				Arch:        "x86_64",
				DownloadURL: "https://example.com/test.iso",
			},
			wantErr: true,
			errMsg:  "version",
		},
		{
			name: "empty arch",
			req: &ISOCreateRequest{
				Name:        "Test",
				Version:     "1.0",
				Arch:        "",
				DownloadURL: "https://example.com/test.iso",
			},
			wantErr: true,
			errMsg:  "arch",
		},
		{
			name: "empty download URL",
			req: &ISOCreateRequest{
				Name:        "Test",
				Version:     "1.0",
				Arch:        "x86_64",
				DownloadURL: "",
			},
			wantErr: true,
			errMsg:  "download_url",
		},
		{
			name: "invalid download URL",
			req: &ISOCreateRequest{
				Name:        "Test",
				Version:     "1.0",
				Arch:        "x86_64",
				DownloadURL: "not-a-url",
			},
			wantErr: true,
			errMsg:  "download_url",
		},
		{
			name: "invalid download URL scheme",
			req: &ISOCreateRequest{
				Name:        "Test",
				Version:     "1.0",
				Arch:        "x86_64",
				DownloadURL: "ftp://example.com/test.iso",
			},
			wantErr: true,
			errMsg:  "download_url",
		},
		{
			name: "invalid checksum URL",
			req: &ISOCreateRequest{
				Name:         "Test",
				Version:      "1.0",
				Arch:         "x86_64",
				DownloadURL:  "https://example.com/test.iso",
				ChecksumURL:  "not-a-url",
				ChecksumType: "sha256",
			},
			wantErr: true,
			errMsg:  "checksum_url",
		},
		{
			name: "invalid checksum type",
			req: &ISOCreateRequest{
				Name:         "Test",
				Version:      "1.0",
				Arch:         "x86_64",
				DownloadURL:  "https://example.com/test.iso",
				ChecksumURL:  "https://example.com/test.iso.sha1",
				ChecksumType: "sha1",
			},
			wantErr: true,
			errMsg:  "checksum_type",
		},
		{
			name: "very long name",
			req: &ISOCreateRequest{
				Name:        strings.Repeat("a", 300),
				Version:     "1.0",
				Arch:        "x86_64",
				DownloadURL: "https://example.com/test.iso",
			},
			wantErr: true,
			errMsg:  "name",
		},
		{
			name: "very long version",
			req: &ISOCreateRequest{
				Name:        "Test",
				Version:     strings.Repeat("1", 150),
				Arch:        "x86_64",
				DownloadURL: "https://example.com/test.iso",
			},
			wantErr: true,
			errMsg:  "version",
		},
		{
			name: "very long arch",
			req: &ISOCreateRequest{
				Name:        "Test",
				Version:     "1.0",
				Arch:        strings.Repeat("x", 150),
				DownloadURL: "https://example.com/test.iso",
			},
			wantErr: true,
			errMsg:  "arch",
		},
		{
			name: "very long edition",
			req: &ISOCreateRequest{
				Name:        "Test",
				Version:     "1.0",
				Arch:        "x86_64",
				Edition:     strings.Repeat("a", 150),
				DownloadURL: "https://example.com/test.iso",
			},
			wantErr: true,
			errMsg:  "edition",
		},
		{
			name: "unicode in name",
			req: &ISOCreateRequest{
				Name:        "Alpine Linux 中文版",
				Version:     "1.0",
				Arch:        "x86_64",
				DownloadURL: "https://example.com/test.iso",
			},
			wantErr: false, // Should be allowed
		},
		{
			name: "emoji in name",
			req: &ISOCreateRequest{
				Name:        "Alpine Linux 🐧",
				Version:     "1.0",
				Arch:        "x86_64",
				DownloadURL: "https://example.com/test.iso",
			},
			wantErr: false, // Should be allowed
		},
		{
			name: "path traversal in name",
			req: &ISOCreateRequest{
				Name:        "../etc/passwd",
				Version:     "1.0",
				Arch:        "x86_64",
				DownloadURL: "https://example.com/test.iso",
			},
			wantErr: false, // Will be normalized by service layer
		},
		{
			name: "sql injection in name",
			req: &ISOCreateRequest{
				Name:        "'; DROP TABLE isos; --",
				Version:     "1.0",
				Arch:        "x86_64",
				DownloadURL: "https://example.com/test.iso",
			},
			wantErr: false, // Will be escaped by database layer
		},
		{
			name: "null bytes in name",
			req: &ISOCreateRequest{
				Name:        "Test\x00Linux",
				Version:     "1.0",
				Arch:        "x86_64",
				DownloadURL: "https://example.com/test.iso",
			},
			wantErr: false, // Should be allowed (will be handled by service)
		},
		{
			name: "very long URL",
			req: &ISOCreateRequest{
				Name:        "Test",
				Version:     "1.0",
				Arch:        "x86_64",
				DownloadURL: "https://example.com/" + strings.Repeat("a", 3000),
			},
			wantErr: true,
			errMsg:  "download_url",
		},
		{
			name: "checksum without checksum URL",
			req: &ISOCreateRequest{
				Name:         "Test",
				Version:      "1.0",
				Arch:         "x86_64",
				DownloadURL:  "https://example.com/test.iso",
				ChecksumType: "sha256",
			},
			wantErr: false, // Checksum type without URL is OK (will be ignored)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateISOCreateRequest(tt.req)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error containing %q, got nil", tt.errMsg)
					return
				}

				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error containing %q, got: %v", tt.errMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			}
		})
	}
}

func TestValidationErrors(t *testing.T) {
	req := &ISOCreateRequest{
		Name:        "",
		Version:     "",
		Arch:        "",
		DownloadURL: "",
	}

	err := ValidateISOCreateRequest(req)
	if err == nil {
		t.Fatal("Expected validation error, got nil")
	}

	// Should contain multiple field errors
	errStr := err.Error()
	expectedFields := []string{"name", "version", "arch", "download_url"}

	for _, field := range expectedFields {
		if !strings.Contains(errStr, field) {
			t.Errorf("Expected error to contain field %q, got: %v", field, errStr)
		}
	}
}

func TestValidateISOCreateRequestNilRequest(t *testing.T) {
	err := ValidateISOCreateRequest(nil)
	if err == nil {
		t.Error("Expected error for nil request, got nil")
	}
}
