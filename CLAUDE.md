# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Linux ISO Manager - A self-hosted application to download, verify, and serve Linux ISOs via HTTP.

**Tech Stack:**
- Backend: Go + Gin framework
- Database: SQLite (modernc.org/sqlite - pure Go, no CGO)
- Frontend: React + TypeScript + Vite + Tailwind CSS (using Bun)
- State Management: Zustand + TanStack Query
- Real-time: WebSocket for progress updates

## Project Structure

```
isoman/
├── backend/
│   ├── main.go                    # Entry point, server initialization
│   ├── internal/
│   │   ├── models/iso.go          # ISO data model and request structs
│   │   ├── db/sqlite.go           # Database layer (CRUD operations)
│   │   ├── api/
│   │   │   ├── routes.go          # Route configuration
│   │   │   ├── handlers.go        # REST API handlers
│   │   │   ├── response.go        # Uniform API response helpers
│   │   │   └── directory.go       # Apache-style directory listing
│   │   ├── download/
│   │   │   ├── manager.go         # Download queue manager with workers
│   │   │   ├── worker.go          # Download worker with progress tracking
│   │   │   └── checksum.go        # Hash computation and verification
│   │   └── ws/
│   │       ├── hub.go             # WebSocket hub for broadcasting
│   │       └── client.go          # WebSocket client connection handling
│   └── data/
│       ├── isos/                  # Downloaded ISOs (nested: name/version/arch/)
│       │   ├── alpine/
│       │   │   └── 3.19.1/
│       │   │       └── x86_64/
│       │   │           └── alpine-3.19.1-x86_64.iso
│       │   ├── ubuntu/
│       │   │   └── 24.04/
│       │   │       └── x86_64/
│       │   │           ├── ubuntu-24.04-desktop-x86_64.iso
│       │   │           └── ubuntu-24.04-server-x86_64.iso
│       │   └── .tmp/              # Temporary download directory
│       └── db/
│           └── isos.db            # SQLite database
└── ui/
    └── src/
        ├── types/iso.ts           # TypeScript type definitions
        ├── api/isos.ts            # API client for backend communication
        ├── stores/                # Zustand state management
        │   └── useAppStore.ts     # Global UI state (theme, view mode, WebSocket status)
        ├── hooks/
        │   ├── useWebSocket.ts    # WebSocket connection hook
        │   └── useCopyWithFeedback.ts  # Copy to clipboard with feedback
        ├── lib/                   # Utility functions
        │   ├── format.ts          # Formatting utilities (bytes, dates)
        │   ├── iso-utils.ts       # ISO-specific utilities
        │   └── status-config.ts   # Status badge configuration
        ├── layouts/               # Layout components
        │   ├── MainLayout.tsx     # Main layout with header/footer
        │   ├── Header.tsx         # Application header
        │   └── Footer.tsx         # Application footer
        ├── routes/                # Route components
        │   ├── Root.tsx           # Root route (redirects to /isos)
        │   ├── NotFound.tsx       # 404 page
        │   └── isos/IsosPage.tsx  # Main ISOs page (container)
        └── components/
            ├── AddIsoForm.tsx     # Form to add new ISO downloads
            ├── IsoList.tsx        # List of all ISOs (presentational)
            ├── IsoCard.tsx        # Individual ISO card view
            ├── IsoListView.tsx    # Table view for ISOs
            ├── ProgressBar.tsx    # Visual progress indicator
            ├── StatusBadge.tsx    # Status badge component
            └── DarkModeToggle.tsx # Theme toggle
```

## Development Commands

### Backend (Go)

```bash
# From backend/ directory
go run main.go                     # Run development server (port 8080)
go build -o server .               # Build production binary
go test ./...                      # Run all tests
go mod download                    # Download dependencies
go mod tidy                        # Clean up dependencies
```

**Environment Variables:**

See `backend/ENV.md` for comprehensive documentation of all 26 configurable environment variables including:
- Server configuration (PORT, timeouts, CORS)
- Database settings (connection pool, journal mode)
- Download configuration (workers, retries, buffer sizes)
- WebSocket settings
- Logging configuration

### Frontend (React with Bun)

```bash
# From ui/ directory
bun install                        # Install dependencies
bun run dev                        # Start Vite dev server (port 5173, proxies to backend)
bun run build                      # Production build to dist/
bun run preview                    # Preview production build
bun run lint                       # Run ESLint
```

**Vite Dev Proxy Configuration:**
- `/api/*` → `http://localhost:8080`
- `/ws` → `ws://localhost:8080` (WebSocket)

**Environment Variables:**
- `PUBLIC_API_URL` (default: http://localhost:8080)
- `PUBLIC_WS_URL` (default: ws://localhost:8080/ws)

### Docker

```bash
docker build -t isoman .           # Build single container image
docker run -d -p 8080:8080 \       # Run container
  -v isoman-data:/data isoman
docker logs -f <container-id>      # View logs
docker stop <container-id>         # Stop container
```

## Architecture

### Download Flow

1. **Client Request**: User submits ISO download URL + optional checksum URL via React form
2. **API Handler** (`CreateISO`): Validates request, creates DB record with status "pending"
3. **Download Manager**: Queues ISO to worker pool (buffered channel, default 2 workers)
4. **Worker Process**:
   - Status → "downloading": HTTP GET with streaming to temp file
   - Progress updates every 1% or 1 second via callback
   - Status → "verifying": If checksum URL provided, fetch expected hash and verify
   - Move temp file to final location
   - **Download checksum file**: Saves checksum file alongside ISO (e.g., `alpine.iso.sha256`)
   - Status → "complete" or "failed"
5. **Progress Callback**: Broadcasts to WebSocket hub
6. **WebSocket Hub**: Pushes progress updates to all connected clients
7. **React Frontend**: Updates UI in real-time via WebSocket messages

### Database Schema

**isos table:**
- `id` (TEXT PRIMARY KEY) - UUID
- `name` (TEXT NOT NULL) - Normalized name (e.g., "alpine", "ubuntu")
- `version` (TEXT NOT NULL) - Version string (e.g., "3.19.1", "24.04", "rolling")
- `arch` (TEXT NOT NULL) - Architecture (e.g., "x86_64", "aarch64", "arm64")
- `edition` (TEXT DEFAULT '') - Edition variant (e.g., "minimal", "desktop", "server")
- `file_type` (TEXT NOT NULL) - File type (e.g., "iso", "qcow2", "vmdk")
- `filename` (TEXT NOT NULL) - Computed filename (e.g., "alpine-3.19.1-x86_64.iso")
- `file_path` (TEXT NOT NULL) - Relative path (e.g., "alpine/3.19.1/x86_64/alpine-3.19.1-x86_64.iso")
- `download_link` (TEXT NOT NULL) - Public URL (e.g., "/images/alpine/3.19.1/x86_64/alpine-3.19.1-x86_64.iso")
- `size_bytes` (INTEGER DEFAULT 0)
- `checksum` (TEXT DEFAULT '') - Verified hash value
- `checksum_type` (TEXT DEFAULT '') - sha256/sha512/md5
- `download_url` (TEXT NOT NULL) - Original download URL
- `checksum_url` (TEXT DEFAULT '') - Checksum file URL
- `status` (TEXT NOT NULL) - pending/downloading/verifying/complete/failed
- `progress` (INTEGER DEFAULT 0) - 0-100
- `error_message` (TEXT DEFAULT '')
- `created_at` (TIMESTAMP NOT NULL)
- `completed_at` (TIMESTAMP)
- **UNIQUE CONSTRAINT**: (name, version, arch, edition, file_type)

### API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/isos` | List all ISOs (ordered by created_at DESC) |
| GET | `/api/isos/:id` | Get single ISO by ID |
| POST | `/api/isos` | Create new ISO download (queues immediately) |
| DELETE | `/api/isos/:id` | Delete ISO file, checksum files, and DB record |
| POST | `/api/isos/:id/retry` | Retry failed download (resets status to pending) |
| GET | `/images/` | Modern Tailwind CSS directory listing with file-type icons |
| GET | `/images/*filepath` | Direct ISO/checksum file download or subdirectory listing |
| GET | `/ws` | WebSocket endpoint for progress updates |
| GET | `/health` | Health check |

### API Response Format

**All API endpoints (except WebSocket and static files) return a uniform JSON response structure:**

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

### API Request/Response Examples

**POST /api/isos - Create ISO Download**

Request body:
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

**Required fields:**
- `name` - Display name (will be normalized to "alpine-linux")
- `version` - Version string (any format: "3.19.1", "24.04", "rolling")
- `arch` - Architecture ("x86_64", "aarch64", "arm64")
- `download_url` - URL to download file from

**Optional fields:**
- `edition` - Edition variant ("minimal", "desktop", "server", etc.) - default: ""
- `checksum_url` - URL to checksum file - default: ""
- `checksum_type` - Hash type ("sha256", "sha512", "md5") - default: "sha256"

**Auto-detected:**
- `file_type` - Extracted from download_url extension (.iso, .qcow2, .vmdk, etc.)

Success Response (201 Created):
```json
{
  "success": true,
  "data": {
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
    "download_url": "https://...",
    "checksum_url": "https://...",
    "status": "pending",
    "progress": 0,
    "error_message": "",
    "created_at": "2024-01-01T00:00:00Z",
    "completed_at": null
  },
  "message": "ISO download queued successfully"
}
```

Conflict Response (409 Conflict - when ISO already exists):
```json
{
  "success": false,
  "error": {
    "code": "CONFLICT",
    "message": "ISO already exists"
  },
  "data": {
    "existing": {
      "id": "existing-uuid",
      "name": "alpine-linux",
      "version": "3.19.1",
      ...
    }
  }
}
```

**GET /api/isos - List All ISOs**

Success Response (200 OK):
```json
{
  "success": true,
  "data": {
    "isos": [
      {
        "id": "uuid-1",
        "name": "alpine-linux",
        "version": "3.19.1",
        ...
      },
      {
        "id": "uuid-2",
        "name": "ubuntu",
        "version": "24.04",
        ...
      }
    ]
  }
}
```

**GET /api/isos/:id - Get Single ISO**

Success Response (200 OK):
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "name": "alpine-linux",
    "version": "3.19.1",
    ...
  }
}
```

Error Response (404 Not Found):
```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "ISO not found"
  }
}
```

**DELETE /api/isos/:id - Delete ISO**

Success Response (200 OK):
```json
{
  "success": true,
  "message": "Resource deleted successfully"
}
```

**POST /api/isos/:id/retry - Retry Failed Download**

Success Response (200 OK):
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "name": "alpine-linux",
    "status": "pending",
    "progress": 0,
    ...
  },
  "message": "Download retry queued successfully"
}
```

Error Response (400 Bad Request - invalid state):
```json
{
  "success": false,
  "error": {
    "code": "INVALID_STATE",
    "message": "Cannot retry ISO with status: complete. Only failed downloads can be retried"
  }
}
```

### WebSocket Message Format

**Server → Client (Progress Update):**
```json
{
  "type": "progress",
  "payload": {
    "id": "uuid-here",
    "progress": 45,
    "status": "downloading"
  }
}
```

### Checksum Verification

- Supports SHA256, SHA512, MD5
- Parses multiple checksum file formats:
  - **Standard format**: `hash  filename` or `hash *filename`
  - **BSD format**: `SHA256 (filename) = hash` (used by Rocky Linux, FreeBSD, macOS, etc.)
- Handles comments (lines starting with #)
- Streams files during hashing to avoid memory issues with large ISOs
- Comparison is case-insensitive
- **Checksum files are saved alongside ISOs** (e.g., `alpine.iso.sha256`) for user verification
- **Checksum files are cleaned up on deletion** (all .sha256, .sha512, .md5 extensions)

### Directory Listing Features

The `/images/` endpoint serves a modern, responsive directory listing built with Tailwind CSS:

**File Type Icons:**
- **Disc icon (purple)**: ISO and IMG files (`.iso`, `.img`)
- **Shield icon (green)**: Checksum files (`.sha256`, `.sha512`, `.md5`) with type labels
- **Document icon (gray)**: Generic files
- **Folder icon (blue)**: Directories

**Features:**
- File sizes shown in human-readable format (B, KB, MB, GB, TB)
- Directories show "-" for size instead of directory entry size
- Files sorted alphabetically with directories first
- Parent directory navigation
- Responsive design with gradient backgrounds and hover effects
- Custom `hasSuffix` template function for file type detection

## Implementation Status

The project is **fully implemented** with the following features:

1. ✅ **Backend**: Go service with SQLite database, golang-migrate for migrations
2. ✅ **Service Layer**: ISO download management with concurrent workers
3. ✅ **REST API**: Full CRUD operations with uniform response format
4. ✅ **WebSocket**: Real-time progress updates to connected clients
5. ✅ **Frontend**: React + TypeScript with Zustand state management and React Router
6. ✅ **Docker**: Single container deployment with multi-stage build
7. ✅ **Configuration**: Viper-based environment variable management (see `backend/ENV.md`)
8. ✅ **Migrations**: Automated database migrations (see `backend/MIGRATIONS.md`)

## Key Design Decisions

- **CGO-free**: Using modernc.org/sqlite (pure Go) to avoid CGO dependencies for easier cross-compilation
- **Concurrent Downloads**: Worker pool pattern with configurable worker count (default 2)
- **Streaming**: Downloads and checksum verification stream data to handle large ISO files without loading into memory
- **Real-time Updates**: WebSocket broadcast pattern pushes progress to all connected clients
- **Graceful Shutdown**: Main.go handles SIGINT/SIGTERM for clean download cancellation
- **Single Container Deployment**: Frontend is built with Bun and embedded in backend binary, served by Gin at `/` for simplified deployment

## Production Deployment

**Single container deployment:**
- Frontend built with Bun and copied into backend container
- Gin serves static files from `ui/dist` at `/`
- Single port exposure (8080)
- Alpine-based image for minimal size

**Dockerfile structure:**
```dockerfile
# Stage 1: Build frontend with Bun
FROM oven/bun:1 AS ui
WORKDIR /app
COPY ui/package.json ui/bun.lockb ./
RUN bun install
COPY ui/ .
RUN bun run build

# Stage 2: Build backend
FROM golang:1.22-alpine AS backend
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
RUN CGO_ENABLED=0 go build -o server .

# Stage 3: Runtime
FROM alpine:3.19
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=backend /app/server .
COPY --from=ui /app/dist ./ui/dist
EXPOSE 8080
CMD ["./server"]
```

**Gin routing** in `routes.go`:
```go
r.Static("/assets", "./ui/dist/assets")
r.StaticFile("/", "./ui/dist/index.html")
r.NoRoute(func(c *gin.Context) {
    c.File("./ui/dist/index.html")
})
```

**Volume mount:** Use `-v isoman-data:/data` or `-v ./isos:/data/isos` for persistence and external ISO access.
