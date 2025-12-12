# Phase 4: WebSocket Progress Updates

## Goal
Real-time download progress via WebSocket.

## Tasks

### 4.1 WebSocket Hub (`internal/ws/hub.go`)

**Message types:**
```go
type Message struct {
    Type    string      `json:"type"`    // "progress"
    Payload interface{} `json:"payload"`
}

type ProgressPayload struct {
    ID       string `json:"id"`
    Progress int    `json:"progress"`
    Status   string `json:"status"`
}
```

**Hub struct:**
- clients map[*Client]bool
- broadcast chan []byte
- register chan *Client
- unregister chan *Client
- mutex for thread safety

**Methods:**
- `NewHub() *Hub`
- `Run()` - main loop handling register/unregister/broadcast
- `BroadcastProgress(id, progress, status)` - sends progress update to all clients
- `ClientCount() int`

### 4.2 WebSocket Client (`internal/ws/client.go`)

**Client struct:**
- hub reference
- conn *websocket.Conn
- send chan []byte

**Constants:**
- writeWait: 10s
- pongWait: 60s
- pingPeriod: 54s (90% of pongWait)
- maxMessageSize: 512

**Methods:**
- `readPump()` - handles incoming messages (we ignore them, just keep connection alive)
- `writePump()` - sends outgoing messages + ping/pong

**ServeWs(hub, w, r):**
- Upgrade HTTP to WebSocket
- Create client
- Register with hub
- Start read/write pumps

### 4.3 Integration

**Update routes.go:**
- Add GET /ws endpoint calling ws.ServeWs

**Update main.go:**
- Create hub, start Run() goroutine
- Pass hub to SetupRoutes
- Progress callback broadcasts via hub:
  ```go
  progressCallback := func(id string, progress int, status ISOStatus) {
      hub.BroadcastProgress(id, progress, string(status))
  }
  ```

### 4.4 WebSocket Message Format

**Progress update (server → client):**
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

### 4.5 Verification

```bash
# Install websocat for testing
# brew install websocat (mac) or cargo install websocat

# Connect to WebSocket
websocat ws://localhost:8080/ws

# In another terminal, start a download
curl -X POST http://localhost:8080/api/isos \
  -H "Content-Type: application/json" \
  -d '{"name":"Test","download_url":"https://dl-cdn.alpinelinux.org/alpine/v3.19/releases/x86_64/alpine-standard-3.19.1-x86_64.iso"}'

# Watch progress messages appear in websocat terminal
```

## Next
Proceed to PHASE_5.md
