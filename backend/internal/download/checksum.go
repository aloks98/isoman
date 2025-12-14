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
	"os"
	"strings"
	"time"

	"linux-iso-manager/internal/httputil"
)

// Streams the file to avoid memory issues with large ISOs.
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

// for the given filename.
func FetchExpectedChecksum(checksumURL, filename string) (string, error) {
	// Use context with timeout for checksum download
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	data, err := httputil.FetchBytes(ctx, checksumURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch checksum file: %w", err)
	}

	checksum, err := ParseChecksumFile(bytes.NewReader(data), filename)
	if err != nil {
		return "", fmt.Errorf("failed to parse checksum file: %w", err)
	}

	return checksum, nil
}

// Handles comments (lines starting with #).
// Supports two formats:
// 1. Standard: "hash  filename" or "hash *filename"
// 2. BSD: "SHA256 (filename) = hash" or "MD5 (filename) = hash"
func ParseChecksumFile(reader io.Reader, filename string) (string, error) {
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Try BSD format first: SHA256 (filename) = hash
		if strings.Contains(line, "(") && strings.Contains(line, ")") && strings.Contains(line, "=") {
			// Extract filename from parentheses
			startParen := strings.Index(line, "(")
			endParen := strings.Index(line, ")")
			if startParen != -1 && endParen != -1 && endParen > startParen {
				fileInLine := strings.TrimSpace(line[startParen+1 : endParen])

				// Check if this is the file we're looking for
				if fileInLine == filename {
					// Extract hash after the = sign
					parts := strings.Split(line[endParen+1:], "=")
					if len(parts) >= 2 {
						hash := strings.TrimSpace(parts[1])
						return strings.ToLower(hash), nil
					}
				}
			}
			continue
		}

		// Try standard format: hash  filename
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
