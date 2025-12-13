# ISOMan API Reference

Complete REST API documentation for ISOMan.

## Base URL

```
http://localhost:8080
```

## Response Format

All API endpoints return a uniform JSON response structure:

**Success Response:**
```json
{
  "success": true,
  "data": { /* endpoint-specific data */ },
  "message": "Optional success message"
}
```

**Error Response:**
```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message",
    "details": "Optional additional details"
  }
}
```

**Error Codes:**
- `BAD_REQUEST` - Invalid request (400)
- `NOT_FOUND` - Resource not found (404)
- `CONFLICT` - Resource already exists (409)
- `INTERNAL_ERROR` - Server error (500)
- `VALIDATION_FAILED` - Request validation failed (400)
- `INVALID_STATE` - Operation not allowed in current state (400)

---

## Endpoints

### 1. List All ISOs

Get a list of all ISO downloads.

**Endpoint:** `GET /api/isos`

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "isos": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "name": "alpine-linux",
        "version": "3.19.1",
        "arch": "x86_64",
        "edition": "",
        "file_type": "iso",
        "filename": "alpine-linux-3.19.1-x86_64.iso",
        "file_path": "alpine-linux/3.19.1/x86_64/alpine-linux-3.19.1-x86_64.iso",
        "download_link": "/images/alpine-linux/3.19.1/x86_64/alpine-linux-3.19.1-x86_64.iso",
        "size_bytes": 200000000,
        "checksum": "abc123...",
        "checksum_type": "sha256",
        "download_url": "https://...",
        "checksum_url": "https://...",
        "status": "complete",
        "progress": 100,
        "error_message": "",
        "created_at": "2024-01-01T00:00:00Z",
        "completed_at": "2024-01-01T00:05:00Z"
      }
    ]
  }
}
```

**Example:**
```bash
curl http://localhost:8080/api/isos
```

---

### 2. Get Single ISO

Get details of a specific ISO by ID.

**Endpoint:** `GET /api/isos/:id`

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "alpine-linux",
    "version": "3.19.1",
    ...
  }
}
```

**Error Response (404 Not Found):**
```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "ISO not found"
  }
}
```

**Example:**
```bash
curl http://localhost:8080/api/isos/550e8400-e29b-41d4-a716-446655440000
```

---

### 3. Create ISO Download

**Endpoint:** `POST /api/isos`

### Request Format

```json
{
  "name": "Alpine Linux",
  "version": "3.19.1",
  "arch": "x86_64",
  "edition": "minimal",
  "download_url": "https://dl-cdn.alpinelinux.org/alpine/v3.19/releases/x86_64/alpine-standard-3.19.1-x86_64.iso",
  "checksum_url": "https://dl-cdn.alpinelinux.org/alpine/v3.19/releases/x86_64/alpine-standard-3.19.1-x86_64.iso.sha256",
  "checksum_type": "sha256"
}
```

### Field Descriptions

| Field | Type | Required | Description | Example |
|-------|------|----------|-------------|---------|
| `name` | string | ✅ Yes | Display name (will be normalized) | "Alpine Linux" → "alpine-linux" |
| `version` | string | ✅ Yes | Version (any format) | "3.19.1", "24.04", "rolling" |
| `arch` | string | ✅ Yes | Architecture | "x86_64", "aarch64", "arm64" |
| `edition` | string | ❌ No | Edition variant | "minimal", "desktop", "server" |
| `download_url` | string | ✅ Yes | URL to download file | "https://..." |
| `checksum_url` | string | ❌ No | URL to checksum file | "https://...sha256" |
| `checksum_type` | string | ❌ No | Hash algorithm (default: sha256) | "sha256", "sha512", "md5" |

### Auto-Detected Fields

- **`file_type`** - Extracted from download_url extension
  - Supported: iso, qcow2, vmdk, vdi, img, raw, vhd, vhdx

### Response (201 Created)

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "alpine-linux",
  "version": "3.19.1",
  "arch": "x86_64",
  "edition": "minimal",
  "file_type": "iso",
  "filename": "alpine-linux-3.19.1-minimal-x86_64.iso",
  "file_path": "alpine-linux/3.19.1/x86_64/alpine-linux-3.19.1-minimal-x86_64.iso",
  "download_link": "/images/alpine-linux/3.19.1/x86_64/alpine-linux-3.19.1-minimal-x86_64.iso",
  "size_bytes": 0,
  "checksum": "",
  "checksum_type": "sha256",
  "download_url": "https://dl-cdn.alpinelinux.org/alpine/v3.19/releases/x86_64/alpine-standard-3.19.1-x86_64.iso",
  "checksum_url": "https://dl-cdn.alpinelinux.org/alpine/v3.19/releases/x86_64/alpine-standard-3.19.1-x86_64.iso.sha256",
  "status": "pending",
  "progress": 0,
  "error_message": "",
  "created_at": "2024-01-01T00:00:00Z",
  "completed_at": null
}
```

### Computed Fields

The following fields are **automatically computed** from your input:

1. **`filename`** - Generated from: `{name}-{version}-{edition}-{arch}.{file_type}`
   - With edition: `alpine-linux-3.19.1-minimal-x86_64.iso`
   - Without edition: `alpine-linux-3.19.1-x86_64.iso`

2. **`file_path`** - Storage path: `{name}/{version}/{arch}/{filename}`
   - Example: `alpine-linux/3.19.1/x86_64/alpine-linux-3.19.1-x86_64.iso`

3. **`download_link`** - Public URL: `/images/{file_path}`
   - Example: `/images/alpine-linux/3.19.1/x86_64/alpine-linux-3.19.1-x86_64.iso`

## Examples

### Example 1: Basic ISO without Edition

```bash
curl -X POST http://localhost:8080/api/isos \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Alpine Linux",
    "version": "3.19.1",
    "arch": "x86_64",
    "download_url": "https://dl-cdn.alpinelinux.org/alpine/v3.19/releases/x86_64/alpine-standard-3.19.1-x86_64.iso",
    "checksum_url": "https://dl-cdn.alpinelinux.org/alpine/v3.19/releases/x86_64/alpine-standard-3.19.1-x86_64.iso.sha256"
  }'
```

**Result:**
- Filename: `alpine-linux-3.19.1-x86_64.iso`
- Path: `alpine-linux/3.19.1/x86_64/alpine-linux-3.19.1-x86_64.iso`
- Link: `/images/alpine-linux/3.19.1/x86_64/alpine-linux-3.19.1-x86_64.iso`

### Example 2: Ubuntu with Edition

```bash
curl -X POST http://localhost:8080/api/isos \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Ubuntu",
    "version": "24.04",
    "arch": "x86_64",
    "edition": "desktop",
    "download_url": "https://releases.ubuntu.com/24.04/ubuntu-24.04-desktop-amd64.iso"
  }'
```

**Result:**
- Filename: `ubuntu-24.04-desktop-x86_64.iso`
- Path: `ubuntu/24.04/x86_64/ubuntu-24.04-desktop-x86_64.iso`
- Link: `/images/ubuntu/24.04/x86_64/ubuntu-24.04-desktop-x86_64.iso`

### Example 3: QCOW2 File

```bash
curl -X POST http://localhost:8080/api/isos \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Fedora",
    "version": "39",
    "arch": "x86_64",
    "download_url": "https://download.fedoraproject.org/pub/fedora/linux/releases/39/Cloud/x86_64/images/Fedora-Cloud-Base-39-1.5.x86_64.qcow2"
  }'
```

**Result:**
- Filename: `fedora-39-x86_64.qcow2`
- File Type: Auto-detected as "qcow2"
- Path: `fedora/39/x86_64/fedora-39-x86_64.qcow2`
- Link: `/images/fedora/39/x86_64/fedora-39-x86_64.qcow2`

### Example 4: ARM Architecture

```bash
curl -X POST http://localhost:8080/api/isos \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Alpine Linux",
    "version": "3.19.1",
    "arch": "aarch64",
    "download_url": "https://dl-cdn.alpinelinux.org/alpine/v3.19/releases/aarch64/alpine-standard-3.19.1-aarch64.iso"
  }'
```

**Result:**
- Filename: `alpine-linux-3.19.1-aarch64.iso`
- Path: `alpine-linux/3.19.1/aarch64/alpine-linux-3.19.1-aarch64.iso`
- Link: `/images/alpine-linux/3.19.1/aarch64/alpine-linux-3.19.1-aarch64.iso`

## Unique Constraint

The system prevents duplicates based on: `(name, version, arch, edition, file_type)`

This means you **can** have:
- ✅ Alpine 3.19.1 x86_64 .iso
- ✅ Alpine 3.19.1 x86_64 .qcow2 (different file_type)
- ✅ Alpine 3.19.1 aarch64 .iso (different arch)
- ✅ Alpine 3.18.0 x86_64 .iso (different version)
- ✅ Ubuntu 24.04 x86_64 desktop .iso
- ✅ Ubuntu 24.04 x86_64 server .iso (different edition)

You **cannot** have:
- ❌ Two identical Alpine 3.19.1 x86_64 .iso files

## Cancellation & Retry

### On Cancellation (Ctrl+C or interrupt):
1. Download stops immediately (context cancelled)
2. Partial temp file is **deleted**
3. Status set to "failed" with message "Download cancelled"
4. Database record kept for retry

---

### 4. Delete ISO

Delete an ISO file and its database record.

**Endpoint:** `DELETE /api/isos/:id`

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Resource deleted successfully"
}
```

**Error Response (404 Not Found):**
```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "ISO not found"
  }
}
```

**Example:**
```bash
curl -X DELETE http://localhost:8080/api/isos/550e8400-e29b-41d4-a716-446655440000
```

**Note:** This will delete:
- The ISO file from disk
- Any checksum files (.sha256, .sha512, .md5)
- The database record

---

### 5. Retry Failed Download

Retry a failed ISO download.

**Endpoint:** `POST /api/isos/:id/retry`

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "alpine-linux",
    "version": "3.19.1",
    "status": "pending",
    "progress": 0,
    ...
  },
  "message": "Download retry queued successfully"
}
```

**Error Response (400 Bad Request - Invalid State):**
```json
{
  "success": false,
  "error": {
    "code": "INVALID_STATE",
    "message": "Cannot retry ISO with status: complete. Only failed downloads can be retried"
  }
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/api/isos/550e8400-e29b-41d4-a716-446655440000/retry
```

**What happens on retry:**
1. Reset status to "pending"
2. Reset progress to 0
3. Clear error message
4. Re-queue the download
5. **Start fresh from 0%** (no resume support)

---

### 6. Health Check

Check if the server is running.

**Endpoint:** `GET /health`

**Response (200 OK):**
```json
{
  "status": "ok"
}
```

**Example:**
```bash
curl http://localhost:8080/health
```

---

## File Serving

### Browse Directory

Browse ISO files with an Apache-style directory listing.

**Endpoint:** `GET /images/`

**Response:** HTML page with file listing

Features:
- File type icons (ISO, checksum files, directories)
- Human-readable file sizes
- Directories sorted first, then files alphabetically
- Parent directory navigation

**Example:**
```bash
curl http://localhost:8080/images/
# Or open in browser for styled HTML view
```

### Download File

Download an ISO or checksum file directly.

**Endpoint:** `GET /images/*filepath`

**Examples:**
```bash
# Download an ISO
curl -O http://localhost:8080/images/alpine-linux/3.19.1/x86_64/alpine-linux-3.19.1-x86_64.iso

# Download checksum file
curl http://localhost:8080/images/alpine-linux/3.19.1/x86_64/alpine-linux-3.19.1-x86_64.iso.sha256
```

---

## WebSocket

### Real-time Progress Updates

Connect to WebSocket for real-time download progress updates.

**Endpoint:** `GET /ws`

**Protocol:** `ws://` (or `wss://` for HTTPS)

**Message Format (Server → Client):**
```json
{
  "type": "progress",
  "payload": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "progress": 45,
    "status": "downloading"
  }
}
```

**Status Values:**
- `pending` - Queued, waiting to start
- `downloading` - Currently downloading
- `verifying` - Verifying checksum
- `complete` - Download and verification successful
- `failed` - Download or verification failed

**Example (JavaScript):**
```javascript
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  if (message.type === 'progress') {
    const { id, progress, status } = message.payload;
    console.log(`ISO ${id}: ${status} ${progress}%`);
  }
};
```

---

## Cancellation & Error Handling

### On Server Shutdown (Ctrl+C or interrupt):
1. All active downloads stop immediately (context cancelled)
2. Partial temp files are **deleted**
3. Status set to "failed" with message "Download cancelled"
4. Database records kept for retry

### Failed Downloads
Downloads can fail due to:
- Network errors
- Invalid URLs
- Checksum verification failure
- Disk space issues
- Server interruption

Failed downloads can be retried using the retry endpoint.
