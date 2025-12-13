package ws

import (
	"encoding/json"
	"log/slog"
	"sync"

	"linux-iso-manager/internal/models"
)

// Message types for WebSocket communication.
const (
	MessageTypeProgress = "progress"
	MessageTypeStatus   = "status"
)

// Message represents a WebSocket message.
type Message struct {
	Payload interface{} `json:"payload"`
	Type    string      `json:"type"`
}

// ProgressPayload represents a progress update message.
type ProgressPayload struct {
	ID       string           `json:"id"`
	Status   models.ISOStatus `json:"status"`
	Progress int              `json:"progress"`
}

// Hub maintains the set of active clients and broadcasts messages to them.
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Protects clients map for concurrent read access
	mu sync.RWMutex

	// Inbound messages from clients
	broadcast chan []byte

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client
}

// NewHub creates a new Hub instance.
func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

// Run starts the hub's main loop.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			count := len(h.clients)
			h.mu.Unlock()
			slog.Debug("websocket client connected", slog.Int("total_clients", count))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				count := len(h.clients)
				h.mu.Unlock()
				slog.Debug("websocket client disconnected", slog.Int("total_clients", count))
			} else {
				h.mu.Unlock()
			}

		case message := <-h.broadcast:
			// Broadcast to all connected clients
			h.mu.Lock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// Client's send buffer is full, close the connection
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.Unlock()
		}
	}
}

// BroadcastProgress sends a progress update to all connected clients.
func (h *Hub) BroadcastProgress(isoID string, progress int, status models.ISOStatus) {
	payload := ProgressPayload{
		ID:       isoID,
		Progress: progress,
		Status:   status,
	}

	message := Message{
		Type:    MessageTypeProgress,
		Payload: payload,
	}

	// Marshal to JSON
	data, err := json.Marshal(message)
	if err != nil {
		slog.Error("failed to marshal progress message", slog.Any("error", err))
		return
	}

	// Send to broadcast channel
	select {
	case h.broadcast <- data:
		// Message sent successfully
	default:
		// Broadcast channel is full, skip this update
		slog.Warn("broadcast channel full, skipping update", slog.String("iso_id", isoID))
	}
}

// ClientCount returns the number of connected clients.
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
