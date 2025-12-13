package download

import (
	"context"
	"linux-iso-manager/internal/testutil"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestCancelDownload(t *testing.T) {
	env := testutil.SetupTestEnvironment(t)
	defer env.Cleanup()

	manager := NewManager(env.DB, env.ISODir, 2)
	defer manager.Stop()

	// Test cancelling a non-existent download
	t.Run("cancel non-existent download returns false", func(t *testing.T) {
		result := manager.CancelDownload("non-existent-id")
		if result {
			t.Error("Expected CancelDownload to return false for non-existent download")
		}
	})

	// Test cancelling an active download
	t.Run("cancel active download returns true", func(t *testing.T) {
		// Create a test ISO
		iso := testutil.CreateAndInsertTestISO(t, env.DB, nil)

		// Start the manager to enable workers
		manager.Start()

		// Create a context that we can verify was cancelled
		downloadCtx, downloadCancel := context.WithCancel(context.Background())
		defer downloadCancel()

		// Manually register an active download
		manager.mu.Lock()
		manager.activeDownloads[iso.ID] = downloadCancel
		manager.mu.Unlock()

		// Verify download is active
		if !manager.IsDownloading(iso.ID) {
			t.Fatal("Expected download to be active before cancellation")
		}

		// Cancel the download
		result := manager.CancelDownload(iso.ID)
		if !result {
			t.Error("Expected CancelDownload to return true for active download")
		}

		// Verify download is no longer active
		if manager.IsDownloading(iso.ID) {
			t.Error("Expected download to not be active after cancellation")
		}

		// Verify context was cancelled
		select {
		case <-downloadCtx.Done():
			// Context was cancelled as expected
		case <-time.After(100 * time.Millisecond):
			t.Error("Expected download context to be cancelled")
		}
	})

	// Test double cancellation
	t.Run("double cancel returns false on second attempt", func(t *testing.T) {
		iso := testutil.CreateAndInsertTestISO(t, env.DB, &testutil.TestISO{
			Name: "test-double-cancel",
		})

		_, downloadCancel := context.WithCancel(context.Background())
		defer downloadCancel()

		// Register download
		manager.mu.Lock()
		manager.activeDownloads[iso.ID] = downloadCancel
		manager.mu.Unlock()

		// First cancellation should succeed
		if !manager.CancelDownload(iso.ID) {
			t.Error("Expected first cancellation to return true")
		}

		// Second cancellation should fail
		if manager.CancelDownload(iso.ID) {
			t.Error("Expected second cancellation to return false")
		}
	})
}

func TestIsDownloading(t *testing.T) {
	env := testutil.SetupTestEnvironment(t)
	defer env.Cleanup()

	manager := NewManager(env.DB, env.ISODir, 2)
	defer manager.Stop()

	// Test when download is not active
	t.Run("returns false when download not active", func(t *testing.T) {
		if manager.IsDownloading("non-existent-id") {
			t.Error("Expected IsDownloading to return false for non-existent download")
		}
	})

	// Test when download is active
	t.Run("returns true when download is active", func(t *testing.T) {
		iso := testutil.CreateAndInsertTestISO(t, env.DB, nil)

		_, downloadCancel := context.WithCancel(context.Background())
		defer downloadCancel()

		// Register download
		manager.mu.Lock()
		manager.activeDownloads[iso.ID] = downloadCancel
		manager.mu.Unlock()

		if !manager.IsDownloading(iso.ID) {
			t.Error("Expected IsDownloading to return true for active download")
		}

		// Clean up
		manager.CancelDownload(iso.ID)
	})
}

func TestCancelDownloadConcurrency(t *testing.T) {
	env := testutil.SetupTestEnvironment(t)
	defer env.Cleanup()

	manager := NewManager(env.DB, env.ISODir, 2)
	defer manager.Stop()

	// Create multiple ISOs
	isos := make([]*testutil.TestISO, 10)
	for i := 0; i < 10; i++ {
		isos[i] = &testutil.TestISO{
			Name: "test-concurrent-" + string(rune('a'+i)),
		}
	}

	// Register all downloads
	for _, isoData := range isos {
		iso := testutil.CreateAndInsertTestISO(t, env.DB, isoData)
		_, cancel := context.WithCancel(context.Background())

		manager.mu.Lock()
		manager.activeDownloads[iso.ID] = cancel
		manager.mu.Unlock()
	}

	// Cancel all downloads concurrently
	var wg sync.WaitGroup
	for _, isoData := range isos {
		wg.Add(1)
		go func(data *testutil.TestISO) {
			defer wg.Done()

			// Get the ISO from DB
			isos, _ := env.DB.ListISOs()
			for _, iso := range isos {
				if iso.Name == data.Name {
					manager.CancelDownload(iso.ID)
					return
				}
			}
		}(isoData)
	}

	// Wait for all cancellations to complete
	wg.Wait()

	// Verify all downloads were cancelled
	manager.mu.RLock()
	activeCount := len(manager.activeDownloads)
	manager.mu.RUnlock()

	if activeCount != 0 {
		t.Errorf("Expected 0 active downloads after cancellation, got %d", activeCount)
	}
}

func TestIsDownloadingConcurrency(t *testing.T) {
	env := testutil.SetupTestEnvironment(t)
	defer env.Cleanup()

	manager := NewManager(env.DB, env.ISODir, 2)
	defer manager.Stop()

	iso := testutil.CreateAndInsertTestISO(t, env.DB, nil)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Register download
	manager.mu.Lock()
	manager.activeDownloads[iso.ID] = cancel
	manager.mu.Unlock()

	// Check IsDownloading concurrently from multiple goroutines
	var wg sync.WaitGroup
	var errorCount int32

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			if !manager.IsDownloading(iso.ID) {
				atomic.AddInt32(&errorCount, 1)
			}
		}()
	}

	wg.Wait()

	if errorCount > 0 {
		t.Errorf("IsDownloading returned false %d times when download was active", errorCount)
	}

	// Clean up
	manager.CancelDownload(iso.ID)
}

func TestCancelDownloadRemovesFromMap(t *testing.T) {
	env := testutil.SetupTestEnvironment(t)
	defer env.Cleanup()

	manager := NewManager(env.DB, env.ISODir, 2)
	defer manager.Stop()

	iso := testutil.CreateAndInsertTestISO(t, env.DB, nil)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Register download
	manager.mu.Lock()
	manager.activeDownloads[iso.ID] = cancel
	initialCount := len(manager.activeDownloads)
	manager.mu.Unlock()

	if initialCount != 1 {
		t.Fatalf("Expected 1 active download, got %d", initialCount)
	}

	// Cancel download
	manager.CancelDownload(iso.ID)

	// Verify it was removed from the map
	manager.mu.RLock()
	finalCount := len(manager.activeDownloads)
	manager.mu.RUnlock()

	if finalCount != 0 {
		t.Errorf("Expected 0 active downloads after cancellation, got %d", finalCount)
	}
}
