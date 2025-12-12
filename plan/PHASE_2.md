# Phase 2: Download Manager & Checksum Verification

## Goal
Background download worker with progress tracking and checksum verification.

## Tasks

### 2.1 Checksum Utilities (`internal/download/checksum.go`)

**Functions:**
- `ComputeHash(filepath, hashType string) (string, error)`
  - Supports: sha256, sha512, md5
  - Streams file to avoid memory issues
  
- `FetchExpectedChecksum(checksumURL, filename string) (string, error)`
  - Downloads checksum file
  - Parses standard format: `hash  filename` or `hash *filename`
  - Returns lowercase hash

- `ParseChecksumFile(reader io.Reader, filename string) (string, error)`
  - Parses checksum file content
  - Finds line matching filename
  - Handles comments (lines starting with #)

### 2.2 Download Worker (`internal/download/worker.go`)

**ProgressCallback type:** `func(isoID string, progress int, status ISOStatus)`

**Worker struct:**
- db reference
- isoDir path
- tmpDir path (isoDir/.tmp/)
- progressCallback

**Methods:**
- `NewWorker(db, isoDir, callback) *Worker`
- `Process(iso *ISO) error`

**Process flow:**
1. Update status → `downloading`
2. HTTP GET download URL
3. Get Content-Length, update ISO size
4. Stream to temp file with progress updates (every 1% or every second)
5. If checksum URL provided:
   - Update status → `verifying`
   - Fetch expected checksum
   - Compute actual checksum
   - Compare (fail if mismatch)
6. Move temp file to final location
7. Update status → `complete`
8. On any error: delete temp file, status → `failed` with error message

### 2.3 Download Manager (`internal/download/manager.go`)

**Manager struct:**
- db reference
- isoDir path
- queue (chan *ISO, buffered 100)
- workerCount
- progressCallback
- shutdown channel
- WaitGroup

**Methods:**
- `NewManager(db, isoDir, workerCount) *Manager`
- `SetProgressCallback(cb ProgressCallback)`
- `Start()` - launches worker goroutines
- `Stop()` - graceful shutdown
- `QueueDownload(iso *ISO)` - adds to queue

**Worker goroutine:**
- Loop: select on queue channel or shutdown
- Process each job
- Log start/complete/error

### 2.4 Verification
Test with real ISO download:
- Alpine Linux (~200MB): `https://dl-cdn.alpinelinux.org/alpine/v3.19/releases/x86_64/alpine-standard-3.19.1-x86_64.iso`
- Checksum: `https://dl-cdn.alpinelinux.org/alpine/v3.19/releases/x86_64/alpine-standard-3.19.1-x86_64.iso.sha256`

## Next
Proceed to PHASE_3.md
