package download

import (
	"bufio"
	"bytes"
	"context"
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
	"io"
	"linux-iso-manager/internal/httputil"
	"os"
	"strings"
	"time"
)

// ComputeHash computes the hash of a file using the specified hash type
// Supports: sha256, sha512, md5
// Streams the file to avoid memory issues with large ISOs
func ComputeHash(filepath string, hashType string) (string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var hasher hash.Hash
	switch strings.ToLower(hashType) {
	case "sha256":
		hasher = sha256.New()
	case "sha512":
		hasher = sha512.New()
	case "md5":
		hasher = md5.New()
	default:
		return "", fmt.Errorf("unsupported hash type: %s", hashType)
	}

	// Stream file to hasher
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("failed to compute hash: %w", err)
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

// FetchExpectedChecksum downloads a checksum file and parses it to find the expected hash
// for the given filename
func FetchExpectedChecksum(checksumURL, filename string) (string, error) {
	// Use context with timeout for checksum download
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	data, err := httputil.FetchBytes(ctx, checksumURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch checksum file: %w", err)
	}

	return ParseChecksumFile(bytes.NewReader(data), filename)
}

// ParseChecksumFile parses a checksum file to find the hash for the given filename
// Supports standard format: "hash  filename" or "hash *filename"
// Handles comments (lines starting with #)
func ParseChecksumFile(reader io.Reader, filename string) (string, error) {
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split on whitespace
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		hash := parts[0]
		// The filename might have * prefix (binary mode indicator)
		fileInLine := strings.TrimPrefix(parts[1], "*")

		// Match the filename
		if fileInLine == filename {
			return strings.ToLower(hash), nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading checksum file: %w", err)
	}

	return "", fmt.Errorf("checksum not found for file: %s", filename)
}
