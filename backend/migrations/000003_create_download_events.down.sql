-- Drop download_events table and indexes
DROP INDEX IF EXISTS idx_download_events_downloaded_at;
DROP INDEX IF EXISTS idx_download_events_iso_id;
DROP TABLE IF EXISTS download_events;
