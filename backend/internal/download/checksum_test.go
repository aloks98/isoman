package download

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestComputeHash(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := "Hello, World!"

	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name     string
		hashType string
		want     string
	}{
		{
			name:     "SHA256",
			hashType: "sha256",
			want:     "dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f",
		},
		{
			name:     "SHA512",
			hashType: "sha512",
			want:     "374d794a95cdcfd8b35993185fef9ba368f160d8daf432d08ba9f1ed1e5abe6cc69291e0fa2fe0006a52570ef18c19def4e617c33ce52ef0a6e5fbe318cb0387",
		},
		{
			name:     "MD5",
			hashType: "md5",
			want:     "65a8e27d8879283831b664bd8b7f0ad4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ComputeHash(testFile, tt.hashType)
			if err != nil {
				t.Fatalf("ComputeHash() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("ComputeHash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestComputeHashUnsupportedType(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err = ComputeHash(testFile, "unsupported")
	if err == nil {
		t.Error("Expected error for unsupported hash type, got nil")
	}
}

func TestComputeHashNonExistentFile(t *testing.T) {
	_, err := ComputeHash("/nonexistent/file.txt", "sha256")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestParseChecksumFile(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		filename string
		want     string
		wantErr  bool
	}{
		{
			name: "Standard format with spaces",
			content: `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855  test.iso
abc123def456  other.iso`,
			filename: "test.iso",
			want:     "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			wantErr:  false,
		},
		{
			name: "Binary mode indicator with asterisk",
			content: `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 *test.iso`,
			filename: "test.iso",
			want:     "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			wantErr:  false,
		},
		{
			name: "With comments",
			content: `# This is a comment
# SHA256 checksums
e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855  test.iso`,
			filename: "test.iso",
			want:     "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			wantErr:  false,
		},
		{
			name: "Multiple files, pick correct one",
			content: `abc123  file1.iso
def456  file2.iso
e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855  test.iso
789ghi  file3.iso`,
			filename: "test.iso",
			want:     "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			wantErr:  false,
		},
		{
			name:     "File not found",
			content:  `abc123  other.iso`,
			filename: "test.iso",
			want:     "",
			wantErr:  true,
		},
		{
			name:     "Empty file",
			content:  "",
			filename: "test.iso",
			want:     "",
			wantErr:  true,
		},
		{
			name: "Empty lines and whitespace",
			content: `

e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855  test.iso

`,
			filename: "test.iso",
			want:     "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			wantErr:  false,
		},
		{
			name: "Uppercase hash normalized to lowercase",
			content: `E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855  test.iso`,
			filename: "test.iso",
			want:     "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.content)
			got, err := ParseChecksumFile(reader, tt.filename)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseChecksumFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("ParseChecksumFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestComputeHashLargeFile(t *testing.T) {
	// Test streaming with a larger file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "large.txt")

	// Create a 10MB file
	file, err := os.Create(testFile)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	data := make([]byte, 1024*1024) // 1MB chunk
	for i := 0; i < 10; i++ {
		_, err := file.Write(data)
		if err != nil {
			t.Fatalf("Failed to write test data: %v", err)
		}
	}
	file.Close()

	// Just verify it doesn't crash and returns a hash
	hash, err := ComputeHash(testFile, "sha256")
	if err != nil {
		t.Fatalf("ComputeHash() failed on large file: %v", err)
	}

	if len(hash) != 64 { // SHA256 is 64 hex characters
		t.Errorf("Expected hash length 64, got %d", len(hash))
	}
}
