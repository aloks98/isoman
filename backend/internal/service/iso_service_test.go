package service

import (
	"linux-iso-manager/internal/models"
	"strings"
	"testing"
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
		name            string
		iso             *models.ISO
		wantName        string
		wantFilename    string
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
			wantName:        "alpine-linux",
			wantFilename:    "alpine-linux-3.19.1-standard-x86_64.iso",
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
			wantName:        "ubuntu",
			wantFilename:    "ubuntu-24.04-amd64.iso",
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
			wantName:        "alpinelinux2024",
			wantFilename:    "alpinelinux2024-3.19-x86_64.iso",
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
