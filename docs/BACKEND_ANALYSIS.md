# ISOMan Backend - Code Analysis & Refactoring Recommendations

**Version:** 1.0
**Date:** December 2025
**Status:** Analysis Complete

---

## Executive Summary

This document provides a comprehensive analysis of the ISOMan backend codebase, identifying pain points, code quality issues, and opportunities for improvement. The backend is functionally complete and working, but suffers from technical debt that could impact maintainability, testability, and scalability.

### Key Findings

- **Code Duplication**: 15+ instances of duplicated logic across database, handlers, and workers
- **Configuration Management**: 10+ hardcoded values without centralized configuration
- **Error Handling**: Inconsistent patterns leading to potential bugs
- **Architectural Concerns**: Business logic mixed with infrastructure code
- **Testing Gaps**: Critical files without test coverage
- **No Structured Logging**: Mixed logging approaches, no audit trail

### Recommended Approach

**4-Phase Refactoring Plan** with estimated impact:
- **Phase 1 (Quick Wins)**: Configuration + Utilities - **High Impact, 2-3 days**
- **Phase 2 (Database)**: SQL refactoring + Error handling - **Medium Impact, 2 days**
- **Phase 3 (Architecture)**: Separation of concerns - **High Impact, 3-4 days**
- **Phase 4 (Quality)**: Testing + Logging - **Medium Impact, 2-3 days**

**Total Estimated Effort**: 9-12 days for complete refactoring

---

## Table of Contents

1. [Current Architecture](#current-architecture)
2. [Pain Points Analysis](#pain-points-analysis)
3. [Improvement Opportunities](#improvement-opportunities)
4. [Refactoring Recommendations](#refactoring-recommendations)
5. [Risk Assessment](#risk-assessment)
6. [Migration Strategy](#migration-strategy)

---

## Current Architecture

### Project Structure

```
backend/
├── main.go                          # Entry point, server setup
├── internal/
│   ├── models/
│   │   └── iso.go                   # Data models + business logic (MIXED)
│   ├── db/
│   │   ├── sqlite.go                # Database operations
│   │   └── sqlite_test.go
│   ├── download/
│   │   ├── manager.go               # Download queue manager
│   │   ├── worker.go                # Download worker
│   │   ├── checksum.go              # Checksum verification
│   │   ├── cancel.go                # Download cancellation
│   │   └── *_test.go
│   ├── api/
│   │   ├── routes.go                # Route configuration
│   │   ├── handlers.go              # HTTP handlers (DOING TOO MUCH)
│   │   ├── response.go              # Response helpers
│   │   ├── directory.go             # Directory listing
│   │   ├── templates/               # HTML templates
│   │   └── *_test.go
│   └── ws/
│       ├── hub.go                   # WebSocket hub
│       ├── client.go                # WebSocket client
│       └── hub_test.go
└── data/                            # Runtime data (gitignored)
```

### Current Dependencies

```go
// Core
- Go 1.24.4
- github.com/gin-gonic/gin          // HTTP framework
- modernc.org/sqlite                // Pure Go SQLite (CGO-free)

// Supporting
- github.com/google/uuid            // UUID generation
- github.com/gorilla/websocket      // WebSocket support
- github.com/gin-contrib/cors       // CORS middleware
```

### Data Flow

```
HTTP Request → Gin Router → Handlers → Database/Manager
                                      ↓
                              Download Worker → File System
                                      ↓
                              Progress Callback → WebSocket Hub → Clients
```

---

## Pain Points Analysis

### 1. Code Duplication (CRITICAL)

#### 1.1 SQL Field List Repetition

**Impact:** Schema changes require updates in 3+ locations, high risk of bugs.

**Location:** `internal/db/sqlite.go`

**Problem:**
```go
// Lines 115-118: GetISO()
SELECT id, name, version, arch, edition, file_type, filename, file_path,
       download_link, size_bytes, checksum, checksum_type, download_url,
       checksum_url, status, progress, error_message, created_at, completed_at
FROM isos WHERE id = ?

// Lines 155-157: ListISOs() - EXACT SAME FIELDS
SELECT id, name, version, arch, edition, file_type, filename, file_path,
       download_link, size_bytes, checksum, checksum_type, download_url,
       checksum_url, status, progress, error_message, created_at, completed_at
FROM isos ORDER BY created_at DESC

// Lines 287-290: GetISOByComposite() - EXACT SAME FIELDS AGAIN
SELECT id, name, version, arch, edition, file_type, filename, file_path,
       download_link, size_bytes, checksum, checksum_type, download_url,
       checksum_url, status, progress, error_message, created_at, completed_at
FROM isos WHERE ...
```

**Impact:**
- Adding/removing a column requires changing 3+ places
- Easy to introduce bugs if one location is missed
- Difficult to maintain

**Solution:**
```go
const isoSelectFields = `
    id, name, version, arch, edition, file_type, filename, file_path,
    download_link, size_bytes, checksum, checksum_type, download_url,
    checksum_url, status, progress, error_message, created_at, completed_at
`

// Then use:
query := "SELECT " + isoSelectFields + " FROM isos WHERE id = ?"
```

---

#### 1.2 ISO Scanning Logic Duplication

**Impact:** 19-field scan operation repeated 3 times.

**Location:** `internal/db/sqlite.go`

**Problem:**
```go
// Lines 122-142: GetISO()
var createdAt, completedAtStr string
err = row.Scan(
    &iso.ID, &iso.Name, &iso.Version, &iso.Arch, &iso.Edition,
    &iso.FileType, &iso.Filename, &iso.FilePath, &iso.DownloadLink,
    &iso.SizeBytes, &iso.Checksum, &iso.ChecksumType,
    &iso.DownloadURL, &iso.ChecksumURL,
    &iso.Status, &iso.Progress, &iso.ErrorMessage,
    &createdAt, &completedAtStr,
)
// ... time parsing logic ...

// Lines 175-195: ListISOs() - EXACT SAME SCAN
err := rows.Scan(
    &iso.ID, &iso.Name, &iso.Version, &iso.Arch, &iso.Edition,
    &iso.FileType, &iso.Filename, &iso.FilePath, &iso.DownloadLink,
    &iso.SizeBytes, &iso.Checksum, &iso.ChecksumType,
    &iso.DownloadURL, &iso.ChecksumURL,
    &iso.Status, &iso.Progress, &iso.ErrorMessage,
    &createdAt, &completedAtStr,
)
// ... time parsing logic ...

// Lines 294-314: GetISOByComposite() - EXACT SAME SCAN AGAIN
```

**Solution:**
```go
func scanISO(scanner interface{ Scan(...interface{}) error }) (*models.ISO, error) {
    iso := &models.ISO{}
    var createdAt, completedAtStr string

    err := scanner.Scan(
        &iso.ID, &iso.Name, &iso.Version, /* ... */
    )
    if err != nil {
        return nil, err
    }

    // Parse timestamps
    iso.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
    if completedAtStr != "" {
        t, _ := time.Parse(time.RFC3339, completedAtStr)
        iso.CompletedAt = &t
    }

    return iso, nil
}
```

---

#### 1.3 File Deletion Patterns

**Impact:** Inconsistent error handling, potential for orphaned files.

**Location:** `internal/api/handlers.go` (Lines 158-180)

**Problem:**
```go
// Pattern 1: Main file deletion - ERRORS ARE FATAL
filePath := filepath.Join(h.isoDir, iso.FilePath)
if _, err := os.Stat(filePath); err == nil {
    if err := os.Remove(filePath); err != nil {
        ErrorResponse(c, http.StatusInternalServerError, ...)
        return  // OPERATION FAILS
    }
}

// Pattern 2: Checksum files - ERRORS ARE IGNORED
checksumExtensions := []string{".sha256", ".sha512", ".md5"}
for _, ext := range checksumExtensions {
    checksumFile := filePath + ext
    if _, err := os.Stat(checksumFile); err == nil {
        os.Remove(checksumFile)  // NO ERROR HANDLING!
    }
}

// Pattern 3: Temp file - ERRORS ARE IGNORED
tmpFile := filepath.Join(h.isoDir, ".tmp", iso.Filename)
if _, err := os.Stat(tmpFile); err == nil {
    os.Remove(tmpFile)  // NO ERROR HANDLING!
}
```

**Why This is Bad:**
- Inconsistent: Main file errors fail the operation, but checksum/temp file errors don't
- User Experience: User gets an error if main file deletion fails, but orphaned checksums are silently left behind
- Resource Leak: Failed deletions can accumulate over time

**Solution:**
```go
// Create utility function with consistent behavior
func (h *Handlers) cleanupISOFiles(iso *models.ISO) error {
    var errs []error

    // Delete main file
    if err := fileutil.DeleteFile(filepath.Join(h.isoDir, iso.FilePath)); err != nil {
        errs = append(errs, fmt.Errorf("main file: %w", err))
    }

    // Delete checksum files (log but don't fail)
    for _, ext := range constants.ChecksumExtensions {
        fileutil.DeleteFileSilently(filepath.Join(h.isoDir, iso.FilePath) + ext)
    }

    // Delete temp file (log but don't fail)
    fileutil.DeleteFileSilently(filepath.Join(h.isoDir, ".tmp", iso.Filename))

    if len(errs) > 0 {
        return errors.Join(errs...)
    }
    return nil
}
```

---

#### 1.4 HTTP Download Patterns

**Impact:** HTTP request logic repeated 3 times.

**Locations:**
- `internal/download/worker.go` (Lines 130-143)
- `internal/download/worker.go` (Lines 248-256)
- `internal/download/checksum.go` (Lines 49-56)

**Problem:**
```go
// worker.go - Lines 130-143: Download ISO
req, err := http.NewRequestWithContext(ctx, "GET", iso.DownloadURL, nil)
if err != nil {
    return fmt.Errorf("failed to create request: %w", err)
}

resp, err := http.DefaultClient.Do(req)
if err != nil {
    return fmt.Errorf("download failed: %w", err)
}
defer resp.Body.Close()

if resp.StatusCode != http.StatusOK {
    return fmt.Errorf("download failed with status: %s", resp.Status)
}

// worker.go - Lines 248-256: Download checksum file - ALMOST IDENTICAL
req, err := http.NewRequest("GET", checksumURL, nil)
// ... exact same pattern ...

// checksum.go - Lines 49-56: Fetch checksum file - ALMOST IDENTICAL AGAIN
resp, err := http.Get(checksumURL)
// ... exact same pattern ...
```

**Solution:**
```go
// internal/httputil/httputil.go
func DownloadFile(ctx context.Context, url, destPath string) error {
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return fmt.Errorf("failed to create request: %w", err)
    }

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("server returned %s", resp.Status)
    }

    file, err := os.Create(destPath)
    if err != nil {
        return fmt.Errorf("failed to create file: %w", err)
    }
    defer file.Close()

    _, err = io.Copy(file, resp.Body)
    return err
}
```

---

### 2. Configuration Management (HIGH PRIORITY)

**Impact:** Hardcoded values make it difficult to adjust settings, test, or deploy in different environments.

#### 2.1 Configuration Scattered Across Files

**Locations & Values:**

| File | Line | Value | Impact |
|------|------|-------|--------|
| `main.go` | 77-79 | HTTP timeouts (15s/15s/60s) | Can't tune for large ISOs |
| `main.go` | 119 | Shutdown timeout (5s) | May interrupt downloads |
| `sqlite.go` | 24 | WAL mode | Can't switch to TRUNCATE for testing |
| `sqlite.go` | 30 | Busy timeout (5000ms) | Can't tune for concurrency |
| `routes.go` | 21-25 | CORS origins (3 hardcoded URLs) | Can't add new frontend URLs |
| `manager.go` | 34 | Download queue buffer (100) | Can't tune for load |
| `hub.go` | 46 | WebSocket broadcast channel (256) | Can't tune for # of clients |
| `worker.go` | 112 | Max retries (5) | Can't adjust retry policy |
| `worker.go` | 116 | Retry sleep (100ms) | Can't tune backoff |
| `worker.go` | 163 | Download buffer (32KB) | Can't optimize for network |
| `worker.go` | 192 | Progress update threshold (1% or 1s) | Can't reduce WebSocket traffic |

**Example - HTTP Server Configuration:**
```go
// main.go - Lines 74-80
server := &http.Server{
    Addr:         ":" + port,
    Handler:      router,
    ReadTimeout:  15 * time.Second,  // HARDCODED
    WriteTimeout: 15 * time.Second,  // HARDCODED
    IdleTimeout:  60 * time.Second,  // HARDCODED
}
```

**Problem:**
- Can't increase timeouts for multi-GB ISO downloads
- Can't decrease timeouts for faster shutdown in tests
- Can't tune based on production metrics

---

#### 2.2 Recommended Configuration Structure

**Create:** `internal/config/config.go`

```go
package config

import (
    "os"
    "strconv"
    "time"
)

type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Download DownloadConfig
    WebSocket WebSocketConfig
}

type ServerConfig struct {
    Port         string
    ReadTimeout  time.Duration
    WriteTimeout time.Duration
    IdleTimeout  time.Duration
    ShutdownTimeout time.Duration
    CORSOrigins  []string
}

type DatabaseConfig struct {
    Path         string
    BusyTimeout  time.Duration
    JournalMode  string
}

type DownloadConfig struct {
    DataDir      string
    WorkerCount  int
    QueueBuffer  int
    MaxRetries   int
    RetryDelay   time.Duration
    BufferSize   int
    ProgressUpdateInterval time.Duration
}

type WebSocketConfig struct {
    BroadcastChannelSize int
}

func Load() (*Config, error) {
    return &Config{
        Server: ServerConfig{
            Port:         getEnv("PORT", "8080"),
            ReadTimeout:  getDuration("READ_TIMEOUT", 15*time.Second),
            WriteTimeout: getDuration("WRITE_TIMEOUT", 15*time.Second),
            IdleTimeout:  getDuration("IDLE_TIMEOUT", 60*time.Second),
            ShutdownTimeout: getDuration("SHUTDOWN_TIMEOUT", 5*time.Second),
            CORSOrigins:  getCORSOrigins(),
        },
        Database: DatabaseConfig{
            Path:        getEnv("DB_PATH", "./data/db/isos.db"),
            BusyTimeout: getDuration("DB_BUSY_TIMEOUT", 5*time.Second),
            JournalMode: getEnv("DB_JOURNAL_MODE", "WAL"),
        },
        Download: DownloadConfig{
            DataDir:     getEnv("DATA_DIR", "./data"),
            WorkerCount: getInt("WORKER_COUNT", 2),
            QueueBuffer: getInt("QUEUE_BUFFER", 100),
            MaxRetries:  getInt("MAX_RETRIES", 5),
            RetryDelay:  getDuration("RETRY_DELAY", 100*time.Millisecond),
            BufferSize:  getInt("BUFFER_SIZE", 32*1024),
            ProgressUpdateInterval: getDuration("PROGRESS_UPDATE_INTERVAL", 1*time.Second),
        },
        WebSocket: WebSocketConfig{
            BroadcastChannelSize: getInt("WS_BROADCAST_SIZE", 256),
        },
    }, nil
}

// Helper functions...
func getEnv(key, defaultValue string) string { /* ... */ }
func getInt(key string, defaultValue int) int { /* ... */ }
func getDuration(key string, defaultValue time.Duration) time.Duration { /* ... */ }
```

**Benefits:**
- **Single source of truth** for all configuration
- **Environment variable support** for Docker/production
- **Type safety** - no string parsing errors
- **Easy testing** - can create test configs
- **Documentation** - all config in one place

---

### 3. Error Handling Inconsistencies (MEDIUM PRIORITY)

#### 3.1 Retry Logic Bug

**Impact:** Database update errors are not properly handled.

**Location:** `internal/download/worker.go` (Lines 111-122)

**Problem:**
```go
// Retry logic with database update
maxRetries := 5
for i := 0; i < maxRetries; i++ {
    if err := w.db.UpdateISO(iso); err != nil {
        if i < maxRetries-1 {
            time.Sleep(100 * time.Millisecond)
            continue
        }
        log.Printf("ERROR: Failed to update ISO %s after %d retries: %v",
            iso.ID[:8], maxRetries, err)
    }
    break  // THIS EXECUTES EVEN ON ERROR!
}
```

**Bug:**
1. If the last retry fails, the error is logged but `break` still executes
2. Function continues as if update succeeded
3. No error returned to caller
4. Download manager thinks update worked

**Correct Implementation:**
```go
var lastErr error
for i := 0; i < maxRetries; i++ {
    if err := w.db.UpdateISO(iso); err != nil {
        lastErr = err
        if i < maxRetries-1 {
            time.Sleep(100 * time.Millisecond)
            continue
        }
        return fmt.Errorf("failed to update ISO after %d retries: %w", maxRetries, err)
    }
    return nil  // Success
}
return fmt.Errorf("update failed after %d retries: %w", maxRetries, lastErr)
```

---

#### 3.2 Inconsistent File Error Handling

**Impact:** Operations can partially complete, leaving system in inconsistent state.

**Location:** `internal/api/handlers.go` (DeleteISO handler)

**Current Behavior:**

| Operation | Error Handling | Impact if Fails |
|-----------|----------------|-----------------|
| Delete main ISO file | Returns 500 error | User sees error, DB record kept |
| Delete checksum files | Silently ignored | Orphaned files accumulate |
| Delete temp file | Silently ignored | Orphaned tmp files accumulate |
| Cancel download | Best effort | May continue briefly |
| Delete DB record | Returns 500 error | File deleted but DB record remains |

**Inconsistency Example:**
```go
// Scenario: Main file deleted, DB delete fails
// Result: File is gone but DB still shows it exists
// User Experience: Shows as "complete" but download returns 404

if err := os.Remove(filePath); err != nil {
    ErrorResponse(...)  // Too late - file already deleted!
    return
}

// Later...
if err := h.db.DeleteISO(id); err != nil {
    ErrorResponse(...)  // File is gone but DB record remains
    return
}
```

**Better Approach:**
```go
// 1. Validate everything first
iso, err := h.db.GetISO(id)
if err != nil {
    ErrorResponse(...)
    return
}

// 2. Cancel download if active
if isActiveDownload(iso.Status) {
    h.manager.CancelDownload(id)
    waitForCancellation()
}

// 3. Delete DB record first
if err := h.db.DeleteISO(id); err != nil {
    ErrorResponse(...)
    return  // Nothing modified yet
}

// 4. Clean up files (best effort)
cleanupISOFiles(iso)  // Log errors but don't fail
```

---

### 4. Architecture & Separation of Concerns (HIGH PRIORITY)

#### 4.1 Models Package Mixing Concerns

**Impact:** Hard to test, violates Single Responsibility Principle.

**Location:** `internal/models/iso.go`

**Current Structure:**
```
iso.go (153 lines)
├── Data Structures (30%)
│   ├── ISOStatus type + constants
│   ├── ISO struct
│   └── CreateISORequest struct
│
├── Business Logic (50%)
│   ├── NormalizeName() - String manipulation
│   ├── DetectFileType() - File extension validation
│   ├── GenerateFilename() - Path construction
│   ├── GenerateFilePath() - Path construction
│   ├── GenerateDownloadLink() - URL construction
│   └── ExtractFilenameFromURL() - URL parsing
│
└── Helper Methods (20%)
    ├── GetOriginalFilename()
    └── ComputeFields()
```

**Problem:**
- Models should be **data structures only**
- Business logic should be in **service layer**
- Makes testing harder (can't mock business logic separately)
- Violates Single Responsibility Principle

**Better Organization:**
```
internal/models/iso.go          - Data structures only
internal/service/iso_service.go - Business logic
internal/util/naming.go         - String normalization
internal/util/pathutil.go       - Path construction
```

**Example Refactor:**
```go
// models/iso.go - DATA ONLY
type ISO struct {
    ID           string
    Name         string
    Version      string
    // ... fields only, no methods
}

// service/iso_service.go - BUSINESS LOGIC
type ISOService struct {
    naming   *NamingService
    pathUtil *PathUtil
}

func (s *ISOService) CreateISO(req CreateISORequest) (*models.ISO, error) {
    iso := &models.ISO{
        ID:      uuid.New().String(),
        Name:    s.naming.Normalize(req.Name),
        Version: req.Version,
        // ...
    }

    // Compute derived fields using utilities
    iso.Filename = s.pathUtil.GenerateFilename(...)
    iso.FilePath = s.pathUtil.GenerateFilePath(...)

    return iso, nil
}
```

---

#### 4.2 Handlers Doing Too Much

**Impact:** Hard to test handlers, business logic coupled to HTTP layer.

**Location:** `internal/api/handlers.go`

**Example - DeleteISO Handler (Lines 139-190):**

Current responsibilities:
1. HTTP request validation ✓
2. Database lookup ✓
3. **Download cancellation** ❌ (business logic)
4. **File system operations** ❌ (infrastructure)
5. **Checksum file cleanup** ❌ (infrastructure)
6. **Temp file cleanup** ❌ (infrastructure)
7. Database deletion ✓
8. HTTP response ✓

**Problem:**
```go
func (h *Handlers) DeleteISO(c *gin.Context) {
    id := c.Param("id")

    // HTTP layer concerns - CORRECT
    iso, err := h.db.GetISO(id)
    if err != nil {
        ErrorResponse(...)
        return
    }

    // Business logic - WRONG LAYER
    if iso.Status == models.StatusDownloading || iso.Status == models.StatusVerifying {
        if h.manager.CancelDownload(id) {
            time.Sleep(100 * time.Millisecond)  // Magic number!
        }
    }

    // Infrastructure - WRONG LAYER
    filePath := filepath.Join(h.isoDir, iso.FilePath)
    if _, err := os.Stat(filePath); err == nil {
        if err := os.Remove(filePath); err != nil {
            ErrorResponse(...)
            return
        }
    }

    // More infrastructure - WRONG LAYER
    checksumExtensions := []string{".sha256", ".sha512", ".md5"}
    for _, ext := range checksumExtensions {
        // ... file deletion ...
    }

    // Database - CORRECT
    if err := h.db.DeleteISO(id); err != nil {
        ErrorResponse(...)
        return
    }

    NoContentResponse(c)
}
```

**Better Approach:**
```go
// Handler - HTTP layer only
func (h *Handlers) DeleteISO(c *gin.Context) {
    id := c.Param("id")

    // Validate input
    if id == "" {
        ErrorResponse(c, 400, "BAD_REQUEST", "ID required")
        return
    }

    // Call service layer
    if err := h.isoService.Delete(id); err != nil {
        if errors.Is(err, service.ErrNotFound) {
            ErrorResponse(c, 404, "NOT_FOUND", "ISO not found")
            return
        }
        ErrorResponse(c, 500, "INTERNAL_ERROR", "Failed to delete ISO")
        return
    }

    SuccessResponse(c, 200, gin.H{"message": "ISO deleted successfully"})
}

// service/iso_service.go - Business logic
func (s *ISOService) Delete(id string) error {
    // 1. Get ISO
    iso, err := s.db.GetISO(id)
    if err != nil {
        return ErrNotFound
    }

    // 2. Cancel download if active
    if s.isActiveDownload(iso) {
        s.downloadManager.Cancel(id)
    }

    // 3. Delete from database
    if err := s.db.DeleteISO(id); err != nil {
        return fmt.Errorf("database delete failed: %w", err)
    }

    // 4. Clean up files (best effort)
    s.fileCleanup.CleanupISO(iso)

    return nil
}
```

**Benefits:**
- Handlers are thin and testable
- Business logic is reusable (can call from CLI, API, etc.)
- Easy to mock services in tests
- Clear separation of concerns

---

### 5. Testing Gaps (MEDIUM PRIORITY)

#### 5.1 Missing Test Coverage

**Files Without Tests:**

| File | Lines of Code | Risk Level | Why Critical |
|------|---------------|------------|--------------|
| `routes.go` | 83 | HIGH | Route configuration errors break entire API |
| `cancel.go` | 54 | MEDIUM | Download cancellation is error-prone |
| `response.go` | 83 | LOW | Response helpers (but used everywhere) |

**Partial Test Coverage:**

| File | Test Coverage | Missing Coverage |
|------|---------------|------------------|
| `handlers.go` | ~60% | Error paths, edge cases |
| `worker.go` | ~70% | Cancellation scenarios, retries |
| `manager.go` | ~50% | Concurrent operations, queue full |

---

#### 5.2 Test Setup Duplication

**Problem:** Test setup code repeated across test files.

**Example:**
```go
// handlers_test.go
func setupTestEnvironment(t *testing.T) (*Handlers, func()) {
    tmpDir := t.TempDir()
    isoDir := filepath.Join(tmpDir, "isos")
    os.MkdirAll(isoDir, 0755)
    dbPath := filepath.Join(tmpDir, "test.db")
    database, err := db.New(dbPath)
    // ... 20 more lines ...
}

// worker_test.go
func setupWorkerTest(t *testing.T) (*Worker, *db.DB, string, func()) {
    tmpDir := t.TempDir()
    isoDir := filepath.Join(tmpDir, "isos")
    os.MkdirAll(isoDir, 0755)
    dbPath := filepath.Join(tmpDir, "test.db")
    database, err := db.New(dbPath)
    // ... 18 more lines (almost identical) ...
}

// sqlite_test.go
func setupTestDB(t *testing.T) (*DB, string) {
    tmpDir := t.TempDir()
    dbPath := filepath.Join(tmpDir, "test.db")
    database, err := New(dbPath)
    // ... 15 more lines (almost identical) ...
}
```

**Solution:**
```go
// internal/testutil/testutil.go
type TestEnv struct {
    DB      *db.DB
    ISODir  string
    DBPath  string
    Cleanup func()
}

func SetupTestEnvironment(t *testing.T) *TestEnv {
    tmpDir := t.TempDir()
    isoDir := filepath.Join(tmpDir, "isos")
    if err := os.MkdirAll(isoDir, 0755); err != nil {
        t.Fatal(err)
    }

    dbPath := filepath.Join(tmpDir, "test.db")
    database, err := db.New(dbPath)
    if err != nil {
        t.Fatal(err)
    }

    return &TestEnv{
        DB:     database,
        ISODir: isoDir,
        DBPath: dbPath,
        Cleanup: func() {
            database.Close()
        },
    }
}
```

---

### 6. Logging & Observability (LOW PRIORITY)

#### 6.1 Mixed Logging Approaches

**Problem:** Inconsistent logging makes debugging difficult.

**Examples:**

```go
// main.go - Uses fmt.Println for startup
fmt.Println("=== ISO Manager - Starting Server ===")
fmt.Printf("✓ Database initialized (%s)\n", dbPath)

// main.go - Uses log.Printf for errors
log.Fatalf("Failed to initialize database: %v", err)

// worker.go - Uses log.Printf for progress
log.Printf("[%s] Progress: %d%%, Status: %s", isoID[:8], progress, status)

// handlers.go - NO LOGGING AT ALL
func (h *Handlers) DeleteISO(c *gin.Context) {
    // Deletes files but doesn't log anything!
}
```

**Issues:**
- No log levels (can't control verbosity)
- No structured logging (hard to parse/search)
- Inconsistent formats
- Missing audit trail for critical operations

---

#### 6.2 No Audit Trail

**Missing Logs:**
- ISO creation requests (who created what, when)
- ISO deletion (who deleted what)
- Download failures (why did it fail)
- Authentication/Authorization (when added)
- Configuration changes
- Performance metrics (download speeds, queue sizes)

**Example - What Should Be Logged:**
```go
// When ISO is created
log.Info("ISO created",
    "iso_id", iso.ID,
    "name", iso.Name,
    "version", iso.Version,
    "size", iso.SizeBytes,
    "user", user.ID,  // When auth is added
)

// When download completes
log.Info("Download completed",
    "iso_id", iso.ID,
    "duration", time.Since(startTime),
    "speed", calculateSpeed(size, duration),
)

// When errors occur
log.Error("Download failed",
    "iso_id", iso.ID,
    "error", err,
    "retries", retryCount,
)
```

---

### 7. Validation & Constants (LOW PRIORITY)

#### 7.1 Validation Logic Scattered

**Problem:** Checksum types defined in multiple places.

**Locations:**
1. `models/iso.go` Line 59: Gin binding validation
   ```go
   ChecksumType string `json:"checksum_type" binding:"omitempty,oneof=sha256 sha512 md5"`
   ```

2. `download/checksum.go` Lines 27-35: Switch statement
   ```go
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
   ```

**Impact:**
- Adding SHA384 requires changing 2 files
- Easy to forget one location
- Binding validation and logic can get out of sync

**Solution:**
```go
// internal/constants/constants.go
package constants

var (
    ChecksumTypes      = []string{"sha256", "sha512", "md5"}
    ChecksumExtensions = []string{".sha256", ".sha512", ".md5"}
    SupportedFileTypes = []string{"iso", "qcow2", "vmdk", "vdi", "img", "raw", "vhd", "vhdx"}
)

func ValidChecksumType(t string) bool {
    for _, valid := range ChecksumTypes {
        if strings.EqualFold(t, valid) {
            return true
        }
    }
    return false
}

// Usage in models:
ChecksumType string `json:"checksum_type" binding:"omitempty,oneof=sha256 sha512 md5"`

// Usage in checksum verification:
if !constants.ValidChecksumType(hashType) {
    return "", fmt.Errorf("unsupported hash type: %s (supported: %v)",
        hashType, constants.ChecksumTypes)
}
```

---

#### 7.2 Magic Numbers

**Hardcoded Values:**

```go
// Download buffer size
buf := make([]byte, 32*1024)  // worker.go:163

// Progress update threshold
if progress-lastProgress >= 1 || now.Sub(lastUpdate) >= time.Second  // worker.go:192

// Retry count
maxRetries := 5  // worker.go:112

// Retry delay
time.Sleep(100 * time.Millisecond)  // worker.go:116

// Queue buffer
downloadQueue: make(chan *models.ISO, 100),  // manager.go:34

// Broadcast channel
broadcast: make(chan *Message, 256),  // hub.go:46

// Cancellation wait
time.Sleep(100 * time.Millisecond)  // handlers.go:154
```

**Should Be:**
```go
// internal/config/defaults.go
const (
    DefaultDownloadBufferSize = 32 * 1024
    DefaultProgressUpdateInterval = 1 * time.Second
    DefaultProgressPercentThreshold = 1
    DefaultMaxRetries = 5
    DefaultRetryDelay = 100 * time.Millisecond
    DefaultQueueBuffer = 100
    DefaultBroadcastBuffer = 256
    DefaultCancellationWait = 100 * time.Millisecond
)
```

---

## Improvement Opportunities

### Quick Wins (High Impact, Low Effort)

1. **Extract SQL Constants** (1 hour)
   - Move SQL SELECT fields to constant
   - Create `scanISO()` helper function
   - **Impact**: Eliminates risk of schema change bugs

2. **Create File Utility Package** (2 hours)
   - `DeleteFile()`, `DeleteFileSafely()`, `EnsureDirectory()`
   - Centralize error handling
   - **Impact**: Consistent file operations, better error handling

3. **Extract Configuration** (3 hours)
   - Create `internal/config` package
   - Move all hardcoded values
   - **Impact**: Easier testing, tuning, and deployment

4. **Add Constants Package** (1 hour)
   - Checksum types, file extensions
   - Magic numbers as named constants
   - **Impact**: Easier to maintain and modify

### Medium Effort Improvements

5. **Standardize Error Handling** (4 hours)
   - Fix retry logic bug
   - Consistent file deletion error handling
   - Add error context and wrapping
   - **Impact**: Fewer bugs, better debugging

6. **Create HTTP Utility** (2 hours)
   - Centralize HTTP download logic
   - Consistent timeout and error handling
   - **Impact**: Less duplication, easier to test

7. **Add Structured Logging** (3 hours)
   - Replace `fmt` and `log` with structured logger
   - Add log levels
   - **Impact**: Better observability and debugging

8. **Extract Path Utilities** (2 hours)
   - Centralize path construction
   - Validate paths for security
   - **Impact**: Consistent path handling, security

### Architectural Improvements (High Effort)

9. **Service Layer Separation** (6 hours)
   - Create `internal/service` package
   - Move business logic out of handlers
   - Move business logic out of models
   - **Impact**: Better testability, reusable logic

10. **Improve Test Coverage** (4 hours)
    - Test `routes.go` and `cancel.go`
    - Add edge case tests
    - Create test utilities
    - **Impact**: Catch bugs before production

11. **Add Validation Layer** (3 hours)
    - Centralize input validation
    - Consistent error messages
    - **Impact**: Better API usability

---

## Refactoring Recommendations

### Phase 1: Foundation (2-3 days)

**Goal:** Extract utilities and configuration without changing architecture.

**Tasks:**
1. ✓ Create `internal/config` package
2. ✓ Create `internal/constants` package
3. ✓ Create `internal/fileutil` package
4. ✓ Create `internal/pathutil` package
5. ✓ Create `internal/httputil` package
6. ✓ Extract SQL constants in database layer
7. ✓ Fix retry logic bug
8. ✓ Standardize file deletion error handling

**Testing:**
- Run existing tests (should all pass)
- Manual smoke test of core functionality

**Risk:** Low - No architectural changes

---

### Phase 2: Database & Error Handling (2 days)

**Goal:** Improve database layer and error handling consistency.

**Tasks:**
1. ✓ Refactor SQL field selection
2. ✓ Create `scanISO()` helper
3. ✓ Standardize error messages
4. ✓ Add error context/wrapping
5. ✓ Improve database transaction handling
6. ✓ Add database connection pooling checks

**Testing:**
- Database tests must pass
- Integration tests for error scenarios

**Risk:** Low-Medium - Changes database layer but no API changes

---

### Phase 3: Architecture (3-4 days)

**Goal:** Separate concerns properly.

**Tasks:**
1. ✓ Create `internal/service` package
2. ✓ Move business logic from `models` to `service`
3. ✓ Move business logic from `handlers` to `service`
4. ✓ Create `internal/validation` package
5. ✓ Thin handlers to HTTP layer only
6. ✓ Add audit logging

**Testing:**
- Full integration test suite
- Handler tests should be simpler
- Service layer tests should be added

**Risk:** Medium - Architectural changes, requires careful migration

---

### Phase 4: Quality & Observability (2-3 days)

**Goal:** Improve testing and observability.

**Tasks:**
1. ✓ Add structured logging package
2. ✓ Replace all logging with structured logger
3. ✓ Create `internal/testutil` package
4. ✓ Add tests for `routes.go`
5. ✓ Add tests for `cancel.go`
6. ✓ Add edge case tests
7. ✓ Add integration tests
8. ✓ Add benchmarks for critical paths

**Testing:**
- Achieve >80% test coverage
- All benchmarks should pass performance targets

**Risk:** Low - Quality improvements, no functional changes

---

## Risk Assessment

### Low Risk Refactorings

- Creating utility packages
- Extracting constants
- Adding configuration management
- Improving logging

**Why Low Risk:**
- No changes to existing logic
- Easy to rollback
- Can be done incrementally

### Medium Risk Refactorings

- Database layer changes
- Error handling standardization
- Service layer extraction

**Why Medium Risk:**
- Changes core business logic flow
- Requires careful testing
- May impact existing behavior

**Mitigation:**
- Comprehensive test suite first
- Incremental migration
- Feature flags for new code paths

### High Risk Refactorings

- Major architectural changes
- API contract changes
- Database schema changes

**Why High Risk:**
- Breaks existing integrations
- Requires data migration
- Potential downtime

**Mitigation:**
- Not recommended in this analysis
- Current architecture is sufficient
- Focus on code quality, not complete rewrite

---

## Migration Strategy

### Incremental Approach (Recommended)

1. **No Breaking Changes**
   - All refactorings maintain current API
   - Existing tests continue to pass
   - No data migration required

2. **Feature Branch Strategy**
   ```bash
   main
   ├── refactor/phase-1-utilities
   ├── refactor/phase-2-database
   ├── refactor/phase-3-architecture
   └── refactor/phase-4-quality
   ```

3. **Testing Strategy**
   - Run existing tests after each change
   - Add new tests for refactored code
   - Manual QA for critical paths

4. **Rollout Plan**
   - Phase 1: Utilities (low risk, high value)
   - Phase 2: Database (medium risk, medium value)
   - Phase 3: Architecture (medium risk, high value)
   - Phase 4: Quality (low risk, long-term value)

---

## Metrics & Success Criteria

### Code Quality Metrics

| Metric | Current | Target (Phase 1) | Target (All Phases) |
|--------|---------|------------------|---------------------|
| Code Duplication | ~15 instances | <5 instances | 0 instances |
| Test Coverage | ~60% | ~70% | >85% |
| Cyclomatic Complexity | High in handlers | Medium | Low |
| Lines per Function | >50 in some handlers | <30 | <20 |
| Files without tests | 3 | 1 | 0 |

### Maintainability Metrics

| Metric | Current | Target |
|--------|---------|--------|
| Time to add new checksum type | ~30 min (2 files) | 5 min (1 constant) |
| Time to change config value | ~10 min (rebuild) | 1 min (env var) |
| Time to debug file deletion bug | ~30 min (scattered code) | 5 min (one function) |

### Performance Metrics

| Metric | Current | Target |
|--------|---------|--------|
| API Response Time | <10ms | <5ms |
| Download Throughput | Network limited | Network limited |
| Memory Usage | <100MB | <50MB |
| Test Suite Runtime | <2s | <3s |

---

## Conclusion

The ISOMan backend is **functionally complete and working**, but suffers from technical debt that will make it harder to maintain and extend over time. The recommended 4-phase refactoring plan addresses the most critical issues first while minimizing risk.

### Recommended Action

**Start with Phase 1** (Foundation) - it provides the most value with the least risk:
- Immediate improvement in code quality
- Makes future phases easier
- Low risk of breaking existing functionality
- Can be completed in 2-3 days

### Not Recommended

- Complete rewrite
- Database schema changes
- API contract changes
- Breaking changes to existing functionality

The current architecture is sound; we just need to clean up the implementation.

---

## Appendix: File-by-File Analysis

### main.go (146 lines)

**Strengths:**
- Clear startup sequence
- Graceful shutdown handling
- Good error messages

**Weaknesses:**
- Configuration hardcoded (lines 27-29, 77-79)
- Progress callback defined inline (lines 59-65)
- Mixed logging approaches (fmt vs log)

**Recommendations:**
- Extract configuration to config package
- Extract progress callback to separate function
- Use structured logging

---

### internal/models/iso.go (153 lines)

**Strengths:**
- Clear data structures
- Good validation tags

**Weaknesses:**
- Business logic mixed with models (70% of file)
- String manipulation in models package
- Path construction in models package

**Recommendations:**
- Keep only data structures
- Move business logic to service layer
- Move utilities to util packages

---

### internal/db/sqlite.go (325 lines)

**Strengths:**
- Good use of prepared statements
- Transaction support

**Weaknesses:**
- SQL field list repeated 3 times
- Scan logic repeated 3 times
- No connection pooling configuration

**Recommendations:**
- Extract SQL constants
- Create scanISO helper
- Add configuration for connection pool

---

### internal/api/handlers.go (236 lines)

**Strengths:**
- Clear handler signatures
- Good use of response helpers

**Weaknesses:**
- Handlers doing too much (file operations, cancellation logic)
- No logging for audit trail
- Inconsistent error handling

**Recommendations:**
- Move business logic to service layer
- Add structured logging
- Standardize error handling

---

### internal/download/worker.go (271 lines)

**Strengths:**
- Good progress tracking
- Checksum verification

**Weaknesses:**
- Too many responsibilities
- Retry logic bug (line 112-122)
- Magic numbers throughout

**Recommendations:**
- Extract HTTP download to httputil
- Fix retry logic
- Extract configuration
- Break into smaller components

---

**End of Analysis Document**
