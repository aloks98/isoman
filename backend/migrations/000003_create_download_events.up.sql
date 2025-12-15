-- Create download_events table for time-based statistics
CREATE TABLE IF NOT EXISTS download_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    iso_id TEXT NOT NULL,
    downloaded_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (iso_id) REFERENCES isos(id) ON DELETE CASCADE
);

-- Index for efficient time-based queries
CREATE INDEX idx_download_events_downloaded_at ON download_events(downloaded_at);
CREATE INDEX idx_download_events_iso_id ON download_events(iso_id);
