-- SQLite doesn't support DROP COLUMN directly, need to recreate the table
-- Create backup without download_count
CREATE TABLE isos_backup AS SELECT
    id, name, version, arch, edition, file_type, filename, file_path, download_link,
    size_bytes, checksum, checksum_type, download_url, checksum_url,
    status, progress, error_message, created_at, completed_at
FROM isos;

DROP TABLE isos;

CREATE TABLE isos (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    version TEXT NOT NULL,
    arch TEXT NOT NULL,
    edition TEXT NOT NULL DEFAULT '',
    file_type TEXT NOT NULL,
    filename TEXT NOT NULL,
    file_path TEXT NOT NULL,
    download_link TEXT NOT NULL,
    size_bytes INTEGER DEFAULT 0,
    checksum TEXT DEFAULT '',
    checksum_type TEXT DEFAULT '',
    download_url TEXT NOT NULL,
    checksum_url TEXT DEFAULT '',
    status TEXT NOT NULL,
    progress INTEGER DEFAULT 0,
    error_message TEXT DEFAULT '',
    created_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP,
    UNIQUE(name, version, arch, edition, file_type)
);

INSERT INTO isos SELECT * FROM isos_backup;
DROP TABLE isos_backup;
