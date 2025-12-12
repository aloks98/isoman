# Phase 1: Project Setup & Database

## Goal
Initialize Go project with SQLite database and ISO model.

## Tasks

### 1.1 Initialize Go Module
```bash
mkdir -p isoman/backend
cd isoman/backend
go mod init isoman
```

### 1.2 Create Directory Structure
```
backend/
├── main.go
├── internal/
│   ├── models/
│   ├── db/
│   ├── api/
│   ├── download/
│   └── ws/
└── data/
    ├── isos/
    └── db/
```

### 1.3 Dependencies
```bash
go get modernc.org/sqlite
go get github.com/google/uuid
go get github.com/gin-gonic/gin
go get github.com/gin-contrib/cors
go get github.com/gorilla/websocket
```

### 1.4 ISO Model (`internal/models/iso.go`)

**Struct fields:**
- `ID` (string, UUID)
- `Name` (string) - display name
- `Filename` (string) - extracted from URL
- `SizeBytes` (int64)
- `Checksum` (string) - verified hash
- `ChecksumType` (string) - sha256/sha512/md5
- `DownloadURL` (string)
- `ChecksumURL` (string, optional)
- `Status` (string) - pending/downloading/verifying/complete/failed
- `Progress` (int) - 0-100
- `ErrorMessage` (string)
- `CreatedAt` (time.Time)
- `CompletedAt` (*time.Time)

**Status constants:** `StatusPending`, `StatusDownloading`, `StatusVerifying`, `StatusComplete`, `StatusFailed`

**Request struct** `CreateISORequest`: name, download_url, checksum_url (optional), checksum_type (optional)

### 1.5 Database Layer (`internal/db/sqlite.go`)

**Functions:**
- `New(dbPath string)` - opens DB, runs migrations
- `Close()`
- `CreateISO(iso *ISO)` 
- `GetISO(id string) (*ISO, error)`
- `ListISOs() ([]ISO, error)` - ordered by created_at DESC
- `UpdateISO(iso *ISO)`
- `UpdateISOStatus(id, status, errorMsg)`
- `UpdateISOProgress(id, progress)`
- `UpdateISOSize(id, sizeBytes)`
- `UpdateISOChecksum(id, checksum)`
- `DeleteISO(id)`

**Migration:** Create `isos` table with all fields

### 1.6 Verification
Create simple main.go that:
1. Initializes DB
2. Creates a test ISO record
3. Reads it back
4. Deletes it
5. Prints success

## Next
Proceed to PHASE_2.md
