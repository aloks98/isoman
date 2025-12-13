package download

import "log/slog"

// CancelDownload cancels an ongoing download by ISO ID
// Returns true if a download was cancelled, false if no download was active
func (m *Manager) CancelDownload(isoID string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if cancel, exists := m.activeDownloads[isoID]; exists {
		slog.Info("cancelling download", slog.String("iso_id", isoID))
		cancel()
		delete(m.activeDownloads, isoID)
		return true
	}

	return false
}

// IsDownloading checks if an ISO is currently being downloaded
func (m *Manager) IsDownloading(isoID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, exists := m.activeDownloads[isoID]
	return exists
}
