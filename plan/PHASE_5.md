# Phase 5: React Frontend

## Goal
React UI for managing ISO downloads with real-time progress.

## Tasks

### 5.1 Project Setup

```bash
cd isoman
npm create vite@latest frontend -- --template react-ts
cd frontend
npm install zustand @tanstack/react-query axios
npm install -D tailwindcss postcss autoprefixer
npx tailwindcss init -p
```

### 5.2 Types (`src/types/iso.ts`)

```typescript
type ISOStatus = 'pending' | 'downloading' | 'verifying' | 'complete' | 'failed'

interface ISO {
  id, name, filename, size_bytes, checksum, checksum_type,
  download_url, checksum_url, status, progress, error_message,
  created_at, completed_at
}

interface CreateISORequest {
  name, download_url, checksum_url?, checksum_type?
}

interface WSMessage {
  type: 'progress'
  payload: { id, progress, status }
}
```

### 5.3 API Client (`src/api/isos.ts`)

Base URL from env: `PUBLIC_API_URL` (default: http://localhost:8080)

**Functions:**
- `list()` → ISO[]
- `get(id)` → ISO
- `create(req)` → ISO
- `delete(id)` → void
- `retry(id)` → ISO

### 5.4 Zustand Store (`src/stores/isoStore.ts`)

**State:**
- isos: ISO[]

**Actions:**
- setISOs(isos)
- updateProgress(id, progress, status)
- addISO(iso)
- removeISO(id)

### 5.5 WebSocket Hook (`src/hooks/useWebSocket.ts`)

- Connect to `PUBLIC_WS_URL` (default: ws://localhost:8080/ws)
- Auto-reconnect on disconnect (3s delay)
- Parse incoming messages
- Call store.updateProgress on progress messages

### 5.6 Components

**AddIsoForm.tsx:**
- Form fields: name, download_url, checksum_url (optional), checksum_type (dropdown: sha256/sha512/md5)
- Submit calls API create, adds to store
- Clear form on success

**IsoCard.tsx:**
- Display: name, filename, size (formatted), status icon
- Progress bar (color by status)
- Error message if failed
- Buttons: Delete, Retry (if failed)
- Link to download: `/images/{filename}`

**IsoList.tsx:**
- Maps over store.isos
- Renders IsoCard for each
- Empty state message

**ProgressBar.tsx:**
- Width based on progress %
- Color: blue (downloading), yellow (verifying), green (complete), red (failed), gray (pending)

### 5.7 App.tsx

**Layout:**
```
┌─────────────────────────────────────────┐
│  Linux ISO Manager          [/images/]  │
├─────────────────────────────────────────┤
│  ┌─ Add New ISO ─────────────────────┐  │
│  │ Name: [___________]               │  │
│  │ URL:  [___________]               │  │
│  │ Checksum URL: [___________]       │  │
│  │ Type: [sha256 ▼]    [Download]    │  │
│  └───────────────────────────────────┘  │
│                                         │
│  ┌─ Downloads ───────────────────────┐  │
│  │ ✓ Ubuntu 24.04        5.4 GB      │  │
│  │ ↓ Fedora 40           45% ████░░  │  │
│  │ ✗ Arch (failed)       [Retry]     │  │
│  └───────────────────────────────────┘  │
└─────────────────────────────────────────┘
```

**Features:**
- Header with link to /images/ directory
- Initialize WebSocket connection
- Fetch ISOs on mount
- Sync store with TanStack Query

### 5.8 Vite Config

Proxy `/api` and `/ws` to backend in dev mode:
```typescript
server: {
  proxy: {
    '/api': 'http://localhost:8080',
    '/ws': { target: 'ws://localhost:8080', ws: true }
  }
}
```

### 5.9 Build & Serve

**Development:**
```bash
cd frontend && npm run dev
# Backend serves API on :8080
# Frontend dev server on :5173 with proxy
```

**Production:**
- Build frontend: `npm run build`
- Backend serves `frontend/dist` at `/`
- Or use separate nginx container

### 5.10 Styling

Use Tailwind with dark theme:
- bg-gray-900 body
- bg-gray-800 cards
- Minimal, clean design

## Next
Proceed to PHASE_6.md
