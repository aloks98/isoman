package download

import (
	"context"
	"linux-iso-manager/internal/db"
	"linux-iso-manager/internal/models"
	"log/slog"
	"sync"
)

// Manager manages a pool of download workers
type Manager struct {
	db               *db.DB
	isoDir           string
	queue            chan *models.ISO
	workerCount      int
	progressCallback ProgressCallback
	shutdown         chan struct{}
	ctx              context.Context
	cancel           context.CancelFunc
	wg               sync.WaitGroup
	stopOnce         sync.Once
	// Track active downloads with their cancel functions
	activeDownloads map[string]context.CancelFunc
	mu              sync.RWMutex
}

// NewManager creates a new download manager
func NewManager(database *db.DB, isoDir string, workerCount int) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		db:              database,
		isoDir:          isoDir,
		queue:           make(chan *models.ISO, 100),
		workerCount:     workerCount,
		shutdown:        make(chan struct{}),
		ctx:             ctx,
		cancel:          cancel,
		activeDownloads: make(map[string]context.CancelFunc),
	}
}

// SetProgressCallback sets the callback function for progress updates
func (m *Manager) SetProgressCallback(callback ProgressCallback) {
	m.progressCallback = callback
}

// Start launches the worker goroutines
func (m *Manager) Start() {
	for i := 0; i < m.workerCount; i++ {
		m.wg.Add(1)
		go m.worker(i)
	}
	slog.Debug("download manager workers started", slog.Int("worker_count", m.workerCount))
}

// Stop gracefully shuts down the manager (safe to call multiple times)
func (m *Manager) Stop() {
	m.stopOnce.Do(func() {
		slog.Debug("stopping download manager")
		m.cancel() // Cancel all ongoing downloads
		close(m.shutdown)
		m.wg.Wait()
		slog.Debug("download manager stopped")
	})
}

// QueueDownload adds an ISO to the download queue
func (m *Manager) QueueDownload(iso *models.ISO) {
	m.queue <- iso
}

// worker is the main worker goroutine
func (m *Manager) worker(id int) {
	defer m.wg.Done()

	worker := NewWorker(m.db, m.isoDir, m.progressCallback)

	for {
		select {
		case <-m.shutdown:
			slog.Debug("worker shutting down", slog.Int("worker_id", id))
			return

		case iso := <-m.queue:
			slog.Info("worker starting download",
				slog.Int("worker_id", id),
				slog.String("name", iso.Name),
				slog.String("iso_id", iso.ID),
			)

			// Create a child context that can be cancelled independently
			downloadCtx, cancelDownload := context.WithCancel(m.ctx)

			// Register the cancel function
			m.mu.Lock()
			m.activeDownloads[iso.ID] = cancelDownload
			m.mu.Unlock()

			// Process the download
			err := worker.Process(downloadCtx, iso)

			// Clean up the cancel function
			m.mu.Lock()
			delete(m.activeDownloads, iso.ID)
			m.mu.Unlock()
			cancelDownload() // Clean up context resources

			if err != nil {
				slog.Error("worker download failed",
					slog.Int("worker_id", id),
					slog.String("name", iso.Name),
					slog.Any("error", err),
				)
			} else {
				slog.Info("worker download completed",
					slog.Int("worker_id", id),
					slog.String("name", iso.Name),
				)
			}
		}
	}
}
