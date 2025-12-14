package api

import (
	"linux-iso-manager/internal/config"
	"linux-iso-manager/internal/service"
	"linux-iso-manager/internal/ws"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all routes and middleware.
func SetupRoutes(isoService *service.ISOService, isoDir string, wsHub *ws.Hub, cfg *config.Config) *gin.Engine {
	// Set Gin to release mode for production (can be overridden by GIN_MODE env var)
	// gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	// Configure CORS
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = cfg.Server.CORSOrigins
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept"}
	router.Use(cors.New(corsConfig))

	// Create handlers
	handlers := NewHandlers(isoService, isoDir)

	// API routes
	api := router.Group("/api")
	{
		// ISO management
		api.GET("/isos", handlers.ListISOs)
		api.GET("/isos/:id", handlers.GetISO)
		api.POST("/isos", handlers.CreateISO)
		api.PUT("/isos/:id", handlers.UpdateISO)
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

	// Serve frontend static files
	// In production, frontend is built into ui/dist
	// In development, frontend runs on separate port (3000 or 5173)
	frontendPath := "./ui/dist"

	// Serve static assets (JS, CSS, images, etc.)
	router.Static("/static", frontendPath+"/static")

	// Serve favicon
	router.StaticFile("/favicon.svg", frontendPath+"/favicon.svg")

	// Serve index.html for root and all other routes (for React Router)
	router.NoRoute(func(c *gin.Context) {
		// Don't serve index.html for API routes, WS, images, or health check
		path := c.Request.URL.Path
		if len(path) >= 4 && path[:4] == "/api" {
			ErrorResponse(c, 404, "NOT_FOUND", "API endpoint not found")
			return
		}
		if path == "/ws" || (len(path) >= 7 && path[:7] == "/images") || path == "/health" {
			ErrorResponse(c, 404, "NOT_FOUND", "Resource not found")
			return
		}

		// Serve index.html for all other routes (SPA routing)
		c.File(frontendPath + "/index.html")
	})

	return router
}
