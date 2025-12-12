# Phase 3: REST API & Static File Serving

## Goal
HTTP API for managing ISOs + Apache-style directory listing at `/images/`.

## Tasks

### 3.1 API Handlers (`internal/api/handlers.go`)

**Handlers struct:** holds db, downloadManager, isoDir references

**Endpoints:**

| Method | Path | Handler | Description |
|--------|------|---------|-------------|
| GET | /api/isos | ListISOs | Returns `{"isos": [...]}` |
| GET | /api/isos/:id | GetISO | Returns single ISO or 404 |
| POST | /api/isos | CreateISO | Queues download, returns created ISO |
| DELETE | /api/isos/:id | DeleteISO | Deletes file + DB record, returns 204 |
| POST | /api/isos/:id/retry | RetryISO | Resets failed ISO, re-queues |

**CreateISO logic:**
1. Bind JSON to CreateISORequest (name, version, arch, edition, download_url, checksum_url, checksum_type)
2. Detect file_type from download_url extension
3. Normalize name (e.g., "Alpine Linux" → "alpine-linux")
4. Default checksum_type to "sha256" if checksum_url provided
5. Create ISO record with status "pending"
6. Call iso.ComputeFields() to generate filename, file_path, download_link
7. Save to database
8. Queue to download manager
9. Return 201 with ISO (includes all computed fields)

**DeleteISO logic:**
1. Get ISO from DB
2. Delete file from isoDir if exists
3. Delete temp file if exists
4. Delete DB record
5. Return 204

**RetryISO logic:**
1. Get ISO, verify status is "failed"
2. Reset: status=pending, progress=0, error_message=""
3. Re-queue download
4. Return updated ISO

### 3.2 Directory Listing (`internal/api/directory.go`)

**DirectoryHandler:** Serves `/images/` with Apache-style listing

**Features:**
- HTML page listing all files in isoDir
- Shows: filename, size (human readable), modified date
- Direct download links
- Sort by name
- Simple clean styling (no external CSS needed)

**Template structure:**
```html
<!DOCTYPE html>
<html>
<head><title>Index of /images/</title></head>
<body>
<h1>Index of /images/</h1>
<table>
  <tr><th>Name</th><th>Size</th><th>Modified</th></tr>
  <!-- file rows -->
</table>
</body>
</html>
```

### 3.3 Routes Setup (`internal/api/routes.go`)

**SetupRoutes(db, manager, isoDir, wsHub) *gin.Engine**

```
GET  /                    → Serve frontend (static files or proxy in dev)
GET  /images/             → Directory listing handler
GET  /images/*filepath    → Static file server (direct downloads)
GET  /api/isos            → ListISOs
GET  /api/isos/:id        → GetISO
POST /api/isos            → CreateISO
DELETE /api/isos/:id      → DeleteISO
POST /api/isos/:id/retry  → RetryISO
GET  /ws                  → WebSocket handler
GET  /health              → Health check
```

**CORS config:** Allow localhost:3000, localhost:5173

### 3.4 Main Entry Point (`main.go`)

**Config (env vars with defaults):**
- `PORT` (8080)
- `DATA_DIR` (./data)
- `WORKER_COUNT` (2)

**Startup sequence:**
1. Create directories (isoDir, tmpDir, dbDir)
2. Initialize database
3. Initialize WebSocket hub, start goroutine
4. Initialize download manager with progress callback
5. Start download manager
6. Setup routes
7. Handle graceful shutdown (SIGINT, SIGTERM)
8. Start HTTP server

### 3.5 Verification

```bash
# Start server
go run main.go

# Test API
curl http://localhost:8080/api/isos
curl http://localhost:8080/health

# Add ISO
curl -X POST http://localhost:8080/api/isos \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Alpine Linux",
    "version": "3.19.1",
    "arch": "x86_64",
    "edition": "",
    "download_url": "https://dl-cdn.alpinelinux.org/alpine/v3.19/releases/x86_64/alpine-standard-3.19.1-x86_64.iso",
    "checksum_url": "https://dl-cdn.alpinelinux.org/alpine/v3.19/releases/x86_64/alpine-standard-3.19.1-x86_64.iso.sha256",
    "checksum_type": "sha256"
  }'

# Check directory listing
open http://localhost:8080/images/

# Direct download (note the new path structure)
curl -O http://localhost:8080/images/alpine-linux/3.19.1/x86_64/alpine-linux-3.19.1-x86_64.iso
```

## Next
Proceed to PHASE_4.md
