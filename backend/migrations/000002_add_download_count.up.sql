-- Add download_count column to isos table
ALTER TABLE isos ADD COLUMN download_count INTEGER DEFAULT 0;
