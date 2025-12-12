# Linux ISO Manager

## Overview
Self-hosted app to download, verify, and serve Linux ISOs via HTTP.

## Routes
- `/` - React UI for managing ISO downloads
- `/images/` - Apache-style directory listing with direct ISO downloads
- `/api/*` - REST API
- `/ws` - WebSocket for real-time progress

## Tech Stack
- **Backend:** Go + Gin
- **Database:** SQLite (modernc.org/sqlite - pure Go, no CGO)
- **Frontend:** React + TypeScript + Vite + Tailwind
- **State:** Zustand + TanStack Query

## Features
1. Add ISO download via URL + optional checksum URL
2. Background download with progress tracking
3. Automatic checksum verification (SHA256/SHA512/MD5)
4. Browsable `/images/` directory with direct download links
5. Real-time progress via WebSocket

## Implementation Phases
1. `PHASE_1.md` - Project setup, database, models
2. `PHASE_2.md` - Download manager with checksum verification
3. `PHASE_3.md` - REST API + static file serving
4. `PHASE_4.md` - WebSocket progress updates
5. `PHASE_5.md` - React frontend
6. `PHASE_6.md` - Docker deployment

## Directory Structure
```
isoman/
├── backend/
│   ├── main.go
│   ├── go.mod
│   └── internal/
│       ├── models/iso.go
│       ├── db/sqlite.go
│       ├── api/routes.go, handlers.go
│       ├── download/manager.go, worker.go, checksum.go
│       └── ws/hub.go, client.go
├── frontend/
│   ├── package.json
│   ├── vite.config.ts
│   └── src/
│       ├── App.tsx
│       ├── api/isos.ts
│       ├── stores/isoStore.ts
│       ├── hooks/useWebSocket.ts
│       └── components/AddIsoForm.tsx, IsoList.tsx, IsoCard.tsx
├── data/
│   ├── isos/        # Downloaded ISOs (served at /images/)
│   └── db/isos.db
├── docker-compose.yml
└── Dockerfile
```
