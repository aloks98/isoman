package constants

import "testing"

func TestIsSupportedFileType(t *testing.T) {
	testCases := []struct {
		name     string
		fileType string
		expected bool
	}{
		// Valid file types
		{"iso lowercase", "iso", true},
		{"iso uppercase", "ISO", true},
		{"iso mixed case", "Iso", true},
		{"qcow2", "qcow2", true},
		{"vmdk", "vmdk", true},
		{"vdi", "vdi", true},
		{"img", "img", true},
		{"raw", "raw", true},
		{"vhd", "vhd", true},
		{"vhdx", "vhdx", true},

		// Invalid file types
		{"empty string", "", false},
		{"exe", "exe", false},
		{"zip", "zip", false},
		{"tar", "tar", false},
		{"pdf", "pdf", false},
		{"random string", "notafiletype", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsSupportedFileType(tc.fileType)
			if result != tc.expected {
				t.Errorf("IsSupportedFileType(%q) = %v, expected %v", tc.fileType, result, tc.expected)
			}
		})
	}
}

func TestIsValidChecksumType(t *testing.T) {
	testCases := []struct {
		name         string
		checksumType string
		expected     bool
	}{
		// Valid checksum types
		{"sha256 lowercase", "sha256", true},
		{"sha256 uppercase", "SHA256", true},
		{"sha256 mixed case", "Sha256", true},
		{"sha512", "sha512", true},
		{"SHA512 uppercase", "SHA512", true},
		{"md5", "md5", true},
		{"MD5 uppercase", "MD5", true},

		// Invalid checksum types
		{"empty string", "", false},
		{"sha1", "sha1", false},
		{"sha384", "sha384", false},
		{"crc32", "crc32", false},
		{"random string", "notachecksumtype", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsValidChecksumType(tc.checksumType)
			if result != tc.expected {
				t.Errorf("IsValidChecksumType(%q) = %v, expected %v", tc.checksumType, result, tc.expected)
			}
		})
	}
}

func TestSupportedFileTypes(t *testing.T) {
	// Ensure all supported file types are recognized
	for _, fileType := range SupportedFileTypes {
		if !IsSupportedFileType(fileType) {
			t.Errorf("SupportedFileTypes contains %q but IsSupportedFileType returns false", fileType)
		}
	}
}

func TestChecksumTypes(t *testing.T) {
	// Ensure all checksum types are recognized
	for _, checksumType := range ChecksumTypes {
		if !IsValidChecksumType(checksumType) {
			t.Errorf("ChecksumTypes contains %q but IsValidChecksumType returns false", checksumType)
		}
	}
}

func TestChecksumExtensions(t *testing.T) {
	// Verify checksum extensions correspond to checksum types
	expectedExtensions := map[string]bool{
		".sha256": true,
		".sha512": true,
		".md5":    true,
	}

	if len(ChecksumExtensions) != len(expectedExtensions) {
		t.Errorf("ChecksumExtensions has %d items, expected %d", len(ChecksumExtensions), len(expectedExtensions))
	}

	for _, ext := range ChecksumExtensions {
		if !expectedExtensions[ext] {
			t.Errorf("Unexpected checksum extension: %q", ext)
		}
	}
}
