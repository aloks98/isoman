package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"linux-iso-manager/internal/api"
	"linux-iso-manager/internal/db"
	"linux-iso-manager/internal/download"
	"linux-iso-manager/internal/models"
	"linux-iso-manager/internal/ws"
)

func main() {
	fmt.Println("=== ISO Manager - Starting Server ===")
	fmt.Println()

	// Load configuration from environment variables
	port := getEnv("PORT", "8080")
	dataDir := getEnv("DATA_DIR", "./data")
	workerCount := getEnvInt("WORKER_COUNT", 2)

	// Create directory structure
	isoDir := filepath.Join(dataDir, "isos")
	dbDir := filepath.Join(dataDir, "db")
	tmpDir := filepath.Join(isoDir, ".tmp")

	for _, dir := range []string{isoDir, dbDir, tmpDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}
	fmt.Printf("✓ Directories initialized (data: %s)\n", dataDir)

	// Initialize database
	dbPath := filepath.Join(dbDir, "isos.db")
	database, err := db.New(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()
	fmt.Printf("✓ Database initialized (%s)\n", dbPath)

	// Initialize WebSocket hub
	wsHub := ws.NewHub()
	go wsHub.Run()
	fmt.Println("✓ WebSocket hub started")

	// Initialize download manager with progress callback
	manager := download.NewManager(database, isoDir, workerCount)
	manager.SetProgressCallback(func(isoID string, progress int, status models.ISOStatus) {
		// Broadcast progress to WebSocket clients
		wsHub.BroadcastProgress(isoID, progress, status)

		// Also log to console
		log.Printf("[%s] Progress: %d%%, Status: %s", isoID[:8], progress, status)
	})
	manager.Start()
	fmt.Printf("✓ Download manager started (%d workers)\n", workerCount)

	// Setup routes
	router := api.SetupRoutes(database, manager, isoDir, wsHub)
	fmt.Println("✓ API routes configured")

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		fmt.Println()
		fmt.Printf("🚀 Server listening on http://localhost:%s\n", port)
		fmt.Println()
		fmt.Println("API Endpoints:")
		fmt.Printf("  - GET    /api/isos           List all ISOs\n")
		fmt.Printf("  - POST   /api/isos           Create new ISO download\n")
		fmt.Printf("  - GET    /api/isos/:id       Get ISO by ID\n")
		fmt.Printf("  - DELETE /api/isos/:id       Delete ISO\n")
		fmt.Printf("  - POST   /api/isos/:id/retry Retry failed download\n")
		fmt.Printf("  - GET    /images/            Directory listing\n")
		fmt.Printf("  - GET    /ws                 WebSocket connection\n")
		fmt.Printf("  - GET    /health             Health check\n")
		fmt.Println()
		fmt.Println("Press Ctrl+C to shutdown")
		fmt.Println()

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\n\n⚠️  Shutdown signal received!")
	fmt.Println("Shutting down gracefully...")

	// Stop download manager (cancels active downloads)
	fmt.Println("  - Stopping download manager...")
	manager.Stop()

	// Shutdown HTTP server with timeout
	fmt.Println("  - Stopping HTTP server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	fmt.Println("✓ Server stopped successfully")
}

// getEnv returns environment variable value or default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt returns environment variable as int or default
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
