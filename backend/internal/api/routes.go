package api

import (
	"linux-iso-manager/internal/db"
	"linux-iso-manager/internal/download"
	"linux-iso-manager/internal/ws"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all routes and middleware
func SetupRoutes(database *db.DB, manager *download.Manager, isoDir string, wsHub *ws.Hub) *gin.Engine {
	// Set Gin to release mode for production (can be overridden by GIN_MODE env var)
	// gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	// Configure CORS
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{
		"http://localhost:3000",  // React dev server (npm/yarn)
		"http://localhost:5173",  // Vite dev server (default)
		"http://localhost:8080",  // Same origin
	}
	config.AllowMethods = []string{"GET", "POST", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept"}
	router.Use(cors.New(config))

	// Create handlers
	handlers := NewHandlers(database, manager, isoDir)

	// API routes
	api := router.Group("/api")
	{
		// ISO management
		api.GET("/isos", handlers.ListISOs)
		api.GET("/isos/:id", handlers.GetISO)
		api.POST("/isos", handlers.CreateISO)
		api.DELETE("/isos/:id", handlers.DeleteISO)
		api.POST("/isos/:id/retry", handlers.RetryISO)
	}

	// WebSocket endpoint
	router.GET("/ws", func(c *gin.Context) {
		ws.ServeWS(wsHub, c)
	})

	// Health check
	router.GET("/health", handlers.HealthCheck)

	// Static file serving and directory listing
	// This handles both /images/ (directory listing) and /images/* (file downloads)
	router.GET("/images/*filepath", DirectoryHandler(isoDir))

	// Serve frontend (will be implemented in Phase 5/6)
	// For now, just a placeholder
	router.GET("/", func(c *gin.Context) {
		SuccessResponse(c, 200, gin.H{
			"message": "ISO Manager API",
			"version": "1.0.0",
			"endpoints": gin.H{
				"api":    "/api/isos",
				"images": "/images/",
				"ws":     "/ws",
				"health": "/health",
			},
		})
	})

	return router
}
