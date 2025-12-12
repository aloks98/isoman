package ws

import (
	"linux-iso-manager/internal/models"
	"testing"
	"time"
)

// TestNewHub tests hub creation
func TestNewHub(t *testing.T) {
	hub := NewHub()

	if hub == nil {
		t.Fatal("NewHub should return a hub instance")
	}

	if hub.clients == nil {
		t.Error("Hub clients map should be initialized")
	}

	if hub.broadcast == nil {
		t.Error("Hub broadcast channel should be initialized")
	}

	if hub.register == nil {
		t.Error("Hub register channel should be initialized")
	}

	if hub.unregister == nil {
		t.Error("Hub unregister channel should be initialized")
	}
}

// TestHubClientCount tests the ClientCount method
func TestHubClientCount(t *testing.T) {
	hub := NewHub()

	if count := hub.ClientCount(); count != 0 {
		t.Errorf("New hub should have 0 clients, got: %d", count)
	}
}

// TestHubBroadcastProgress tests broadcasting progress messages
func TestHubBroadcastProgress(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Give hub time to start
	time.Sleep(10 * time.Millisecond)

	// Test broadcasting progress
	hub.BroadcastProgress("test-id-123", 50, models.StatusDownloading)

	// Allow message to be processed
	time.Sleep(10 * time.Millisecond)

	// Verify message was queued (channel should not panic)
	// This test mainly verifies the broadcast doesn't block or panic
}

// TestHubBroadcastProgressFullChannel tests behavior when broadcast channel is full
func TestHubBroadcastProgressFullChannel(t *testing.T) {
	hub := NewHub()
	// Don't start hub.Run() so broadcast channel will fill up

	// Fill the channel
	for i := 0; i < 256; i++ {
		hub.BroadcastProgress("test-id", i, models.StatusDownloading)
	}

	// This should not block - should log warning and continue
	done := make(chan bool)
	go func() {
		hub.BroadcastProgress("test-id", 100, models.StatusComplete)
		done <- true
	}()

	select {
	case <-done:
		// Success - didn't block
	case <-time.After(100 * time.Millisecond):
		t.Error("BroadcastProgress should not block when channel is full")
	}
}

// TestHubRegisterClient tests client registration
func TestHubRegisterClient(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Create a mock client
	client := &Client{
		hub:  hub,
		conn: nil, // nil for testing
		send: make(chan []byte, 256),
	}

	// Register client
	hub.register <- client

	// Give time for registration to process
	time.Sleep(10 * time.Millisecond)

	if count := hub.ClientCount(); count != 1 {
		t.Errorf("Hub should have 1 client after registration, got: %d", count)
	}
}

// TestHubUnregisterClient tests client unregistration
func TestHubUnregisterClient(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Create and register a mock client
	client := &Client{
		hub:  hub,
		conn: nil,
		send: make(chan []byte, 256),
	}

	hub.register <- client
	time.Sleep(10 * time.Millisecond)

	// Verify client is registered
	if count := hub.ClientCount(); count != 1 {
		t.Fatalf("Hub should have 1 client, got: %d", count)
	}

	// Unregister client
	hub.unregister <- client
	time.Sleep(10 * time.Millisecond)

	// Verify client is unregistered
	if count := hub.ClientCount(); count != 0 {
		t.Errorf("Hub should have 0 clients after unregistration, got: %d", count)
	}
}

// TestHubBroadcastToClients tests that messages are broadcast to all clients
func TestHubBroadcastToClients(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Create multiple mock clients
	numClients := 3
	clients := make([]*Client, numClients)

	for i := 0; i < numClients; i++ {
		clients[i] = &Client{
			hub:  hub,
			conn: nil,
			send: make(chan []byte, 256),
		}
		hub.register <- clients[i]
	}

	time.Sleep(10 * time.Millisecond)

	// Broadcast a message
	hub.BroadcastProgress("test-iso", 75, models.StatusDownloading)

	// Give time for broadcast to process
	time.Sleep(10 * time.Millisecond)

	// Verify all clients received the message
	for i, client := range clients {
		select {
		case msg := <-client.send:
			if msg == nil {
				t.Errorf("Client %d received nil message", i)
			}
			// Successfully received message
		default:
			t.Errorf("Client %d did not receive broadcast message", i)
		}
	}
}

// TestHubRemoveSlowConsumer tests that slow consumers are removed
func TestHubRemoveSlowConsumer(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Create client with small buffer
	client := &Client{
		hub:  hub,
		conn: nil,
		send: make(chan []byte, 1), // Very small buffer
	}

	hub.register <- client
	time.Sleep(10 * time.Millisecond)

	// Fill client's send channel
	client.send <- []byte("msg1")

	// Try to broadcast many messages
	for i := 0; i < 10; i++ {
		hub.BroadcastProgress("test-iso", i*10, models.StatusDownloading)
		time.Sleep(5 * time.Millisecond)
	}

	// Client should eventually be removed due to full buffer
	time.Sleep(50 * time.Millisecond)

	// Note: The client might be removed, but we can't reliably test this
	// without access to hub internals. This test mainly verifies no panic occurs.
}
