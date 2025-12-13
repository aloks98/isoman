package download

import (
	"context"
	"linux-iso-manager/internal/db"
	"linux-iso-manager/internal/models"
	"log"
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
	log.Printf("Download manager started with %d workers", m.workerCount)
}

// Stop gracefully shuts down the manager (safe to call multiple times)
func (m *Manager) Stop() {
	m.stopOnce.Do(func() {
		log.Println("Stopping download manager...")
		m.cancel() // Cancel all ongoing downloads
		close(m.shutdown)
		m.wg.Wait()
		log.Println("Download manager stopped")
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
			log.Printf("Worker %d shutting down", id)
			return

		case iso := <-m.queue:
			log.Printf("Worker %d: Starting download of %s (ID: %s)", id, iso.Name, iso.ID)

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
				log.Printf("Worker %d: Failed to download %s: %v", id, iso.Name, err)
			} else {
				log.Printf("Worker %d: Successfully completed %s", id, iso.Name)
			}
		}
	}
}
