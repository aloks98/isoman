package constants

import "strings"

// File types supported for virtual machine images.
var SupportedFileTypes = []string{
	"iso", "qcow2", "vmdk", "vdi",
	"img", "raw", "vhd", "vhdx",
}

// Checksum types supported for file verification.
var ChecksumTypes = []string{"sha256", "sha512", "md5"}

// Checksum file extensions.
var ChecksumExtensions = []string{".sha256", ".sha512", ".md5"}

// Default configuration values.
const (
	// Download settings.
	DefaultWorkerCount              = 2
	DefaultQueueBuffer              = 100
	DefaultDownloadBufferSize       = 32 * 1024 // 32KB
	DefaultMaxRetries               = 5
	DefaultRetryDelayMs             = 100
	DefaultProgressPercentThreshold = 1

	// HTTP server settings.
	DefaultPort               = "8080"
	DefaultReadTimeoutSec     = 15
	DefaultWriteTimeoutSec    = 15
	DefaultIdleTimeoutSec     = 60
	DefaultShutdownTimeoutSec = 5

	// Database settings.
	DefaultBusyTimeoutMs      = 5000
	DefaultJournalMode        = "WAL"
	DefaultMaxOpenConns       = 25
	DefaultMaxIdleConns       = 5
	DefaultConnMaxLifetimeMin = 5
	DefaultConnMaxIdleTimeMin = 5

	// WebSocket settings.
	DefaultBroadcastChannelSize = 256

	// Cancellation settings.
	DefaultCancellationWaitMs = 100
)

// IsSupportedFileType checks if a file type is supported.
func IsSupportedFileType(fileType string) bool {
	fileType = strings.ToLower(fileType)
	for _, supported := range SupportedFileTypes {
		if fileType == supported {
			return true
		}
	}
	return false
}

// IsValidChecksumType checks if a checksum type is valid.
func IsValidChecksumType(checksumType string) bool {
	checksumType = strings.ToLower(checksumType)
	for _, valid := range ChecksumTypes {
		if checksumType == valid {
			return true
		}
	}
	return false
}
