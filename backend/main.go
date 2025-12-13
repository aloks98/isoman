package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"linux-iso-manager/internal/api"
	"linux-iso-manager/internal/config"
	"linux-iso-manager/internal/db"
	"linux-iso-manager/internal/download"
	"linux-iso-manager/internal/fileutil"
	"linux-iso-manager/internal/logger"
	"linux-iso-manager/internal/models"
	"linux-iso-manager/internal/pathutil"
	"linux-iso-manager/internal/service"
	"linux-iso-manager/internal/ws"
)

// Version is set at build time via ldflags
var Version = "dev"

func main() {
	// Load configuration from environment variables
	cfg := config.Load()

	// Initialize structured logger
	log := logger.New(cfg.Log.Level, cfg.Log.Format)
	slog.SetDefault(log)

	log.Info("starting ISO Manager server",
		slog.String("version", Version),
		slog.String("log_level", cfg.Log.Level),
		slog.String("log_format", cfg.Log.Format),
	)

	// Create directory structure
	isoDir := pathutil.GetISODir(cfg.Download.DataDir)
	dbDir := pathutil.GetDBDir(cfg.Download.DataDir)
	tmpDir := pathutil.GetTempDir(isoDir)

	if err := fileutil.EnsureDirectories(isoDir, dbDir, tmpDir); err != nil {
		log.Error("failed to create directories", slog.Any("error", err))
		os.Exit(1)
	}
	log.Info("directories initialized", slog.String("data_dir", cfg.Download.DataDir))

	// Initialize database
	dbPath := pathutil.GetDBPath(cfg.Download.DataDir)
	database, err := db.New(dbPath, &cfg.Database)
	if err != nil {
		log.Error("failed to initialize database", slog.Any("error", err))
		os.Exit(1)
	}
	defer database.Close()
	log.Info("database initialized", slog.String("db_path", dbPath))

	// Initialize WebSocket hub
	wsHub := ws.NewHub()
	go wsHub.Run()
	log.Info("websocket hub started")

	// Initialize download manager with progress callback
	manager := download.NewManager(database, isoDir, cfg.Download.WorkerCount)
	manager.SetProgressCallback(func(isoID string, progress int, status models.ISOStatus) {
		// Broadcast progress to WebSocket clients
		wsHub.BroadcastProgress(isoID, progress, status)

		// Also log progress
		log.Debug("download progress",
			slog.String("iso_id", isoID),
			slog.Int("progress", progress),
			slog.String("status", string(status)),
		)
	})
	manager.Start()
	log.Info("download manager started", slog.Int("worker_count", cfg.Download.WorkerCount))

	// Initialize ISO service
	isoService := service.NewISOService(database, manager)
	log.Info("iso service initialized")

	// Setup routes
	router := api.SetupRoutes(isoService, isoDir, wsHub, cfg)
	log.Info("api routes configured")

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		log.Info("server starting",
			slog.String("address", ":"+cfg.Server.Port),
			slog.String("url", "http://localhost:"+cfg.Server.Port),
		)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server failed to start", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutdown signal received, starting graceful shutdown")

	// Stop download manager (cancels active downloads)
	log.Info("stopping download manager")
	manager.Stop()

	// Shutdown HTTP server with timeout
	log.Info("stopping http server")
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Warn("server forced to shutdown", slog.Any("error", err))
	}

	log.Info("server stopped successfully")
}
