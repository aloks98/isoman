# ISO Manager Restructure Plan
## Support Multiple File Formats per OS Version

## Current State
- One ISO record = one file
- Flat storage: `data/isos/filename.iso`
- Unique constraint: `filename`
- Single download per record

## Proposed State
- Multiple file formats for same OS version (e.g., alpine 3.19 can have .iso, .qcow2, .vmdk)
- Organized storage: `data/isos/{name}/{version}/{name}-{version}.{ext}`
- Unique constraint: `(name, version, file_type)`
- Each file tracked independently

## Example Structure
```
data/isos/
├── alpine/
│   ├── 3.19.1/
│   │   ├── alpine-3.19.1.iso
│   │   └── alpine-3.19.1.qcow2
│   └── 3.18.0/
│       └── alpine-3.18.0.iso
├── ubuntu/
│   └── 24.04/
│       ├── ubuntu-24.04.iso
│       └── ubuntu-24.04.qcow2
```

## Data Model Changes

### Updated ISO Model
```go
type ISO struct {
    ID           string     // UUID
    Name         string     // "alpine" (normalized: lowercase, no spaces)
    Version      string     // "3.19.1"
    FileType     string     // "iso", "qcow2", "vmdk", "vdi", etc.
    Filename     string     // "alpine-3.19.1.iso" (computed)
    FilePath     string     // "alpine/3.19.1/alpine-3.19.1.iso" (computed)
    DownloadLink string     // "/images/alpine/3.19.1/alpine-3.19.1.iso" (computed)

    SizeBytes    int64
    Checksum     string     // Each file has its own checksum
    ChecksumType string
    DownloadURL  string
    ChecksumURL  string

    Status       ISOStatus
    Progress     int
    ErrorMessage string
    CreatedAt    time.Time
    CompletedAt  *time.Time
}

// Unique constraint: (name, version, file_type)
```

### CreateISORequest
```go
type CreateISORequest struct {
    Name         string // "Alpine Linux" (will be normalized)
    Version      string // "3.19.1"
    DownloadURL  string // Auto-detect file type from URL
    ChecksumURL  string
    ChecksumType string // Default: sha256
}
```

## Key Considerations

### 1. File Type Detection
```go
func DetectFileType(url string) string {
    // Extract from URL/filename
    // Support: .iso, .qcow2, .vmdk, .vdi, .img, .raw
    ext := filepath.Ext(url)
    return strings.TrimPrefix(ext, ".")
}
```

### 2. Name Normalization
```go
func NormalizeName(name string) string {
    // "Alpine Linux" -> "alpine"
    // "Ubuntu Server" -> "ubuntu-server"
    // Remove special chars, lowercase, replace spaces with hyphens
    name = strings.ToLower(name)
    name = strings.ReplaceAll(name, " ", "-")
    // Remove trailing version numbers if present
    return name
}
```

### 3. Filename Generation
```go
func GenerateFilename(name, version, fileType string) string {
    // alpine + 3.19.1 + iso = "alpine-3.19.1.iso"
    normalizedName := NormalizeName(name)
    return fmt.Sprintf("%s-%s.%s", normalizedName, version, fileType)
}
```

### 4. Directory Structure
```go
func GenerateFilePath(name, version, filename string) string {
    // alpine + 3.19.1 + alpine-3.19.1.iso
    // = "alpine/3.19.1/alpine-3.19.1.iso"
    normalizedName := NormalizeName(name)
    return filepath.Join(normalizedName, version, filename)
}
```

### 5. Download Link Generation
```go
func GenerateDownloadLink(filePath string) string {
    // "alpine/3.19.1/alpine-3.19.1.iso"
    // -> "/images/alpine/3.19.1/alpine-3.19.1.iso"
    return "/images/" + filePath
}
```

### 6. Duplicate Detection
```sql
-- Check if (name, version, file_type) already exists
SELECT COUNT(*) FROM isos
WHERE name = ? AND version = ? AND file_type = ?
```

### 7. Storage Organization
- Worker creates nested directories: `os.MkdirAll(dir, 0755)`
- Temp files: `data/isos/.tmp/{id}/filename`
- Final location: `data/isos/{name}/{version}/{filename}`

## Database Schema Changes

### Migration Strategy
```sql
-- Add new columns
ALTER TABLE isos ADD COLUMN version TEXT NOT NULL DEFAULT '';
ALTER TABLE isos ADD COLUMN file_type TEXT NOT NULL DEFAULT '';
ALTER TABLE isos ADD COLUMN file_path TEXT NOT NULL DEFAULT '';
ALTER TABLE isos ADD COLUMN download_link TEXT NOT NULL DEFAULT '';

-- Drop old unique constraint on filename
-- (SQLite requires recreating the table)

-- Create new table with updated schema
CREATE TABLE isos_new (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    version TEXT NOT NULL,
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
    UNIQUE(name, version, file_type)
);

-- Copy data (if any exists)
-- Drop old table
-- Rename new table
```

## API Response Changes

### List ISOs Response
```json
{
  "isos": [
    {
      "id": "uuid-1",
      "name": "alpine",
      "version": "3.19.1",
      "file_type": "iso",
      "filename": "alpine-3.19.1.iso",
      "file_path": "alpine/3.19.1/alpine-3.19.1.iso",
      "download_link": "/images/alpine/3.19.1/alpine-3.19.1.iso",
      "size_bytes": 200000000,
      "checksum": "abc123...",
      "checksum_type": "sha256",
      "status": "complete",
      "progress": 100,
      "created_at": "2024-01-01T00:00:00Z",
      "completed_at": "2024-01-01T00:05:00Z"
    },
    {
      "id": "uuid-2",
      "name": "alpine",
      "version": "3.19.1",
      "file_type": "qcow2",
      "filename": "alpine-3.19.1.qcow2",
      "file_path": "alpine/3.19.1/alpine-3.19.1.qcow2",
      "download_link": "/images/alpine/3.19.1/alpine-3.19.1.qcow2",
      "size_bytes": 150000000,
      "checksum": "def456...",
      "checksum_type": "sha256",
      "status": "complete",
      "progress": 100,
      "created_at": "2024-01-01T00:10:00Z",
      "completed_at": "2024-01-01T00:14:00Z"
    }
  ]
}
```

### Create ISO Request
```json
{
  "name": "Alpine Linux",
  "version": "3.19.1",
  "download_url": "https://example.com/alpine-3.19.1-x86_64.iso",
  "checksum_url": "https://example.com/alpine-3.19.1-x86_64.iso.sha256",
  "checksum_type": "sha256"
}
```

## Frontend Grouping (Future Phase 5)

Frontend can group ISOs by (name, version):
```typescript
interface OSVersion {
  name: string;
  version: string;
  files: ISO[];
}

// Group by name + version
const grouped = groupBy(isos, iso => `${iso.name}-${iso.version}`);
```

Display:
```
Alpine Linux 3.19.1
  ├─ ISO (200 MB) [Download] [Delete]
  └─ QCOW2 (150 MB) [Download] [Delete]

Alpine Linux 3.18.0
  └─ ISO (195 MB) [Download] [Delete]

Ubuntu 24.04
  ├─ ISO (5.4 GB) [Download] [Delete]
  └─ QCOW2 (2.1 GB) [Download] [Delete]
```

## Implementation Steps

### Step 1: Update Models
- [ ] Add `Version`, `FileType`, `FilePath`, `DownloadLink` to ISO struct
- [ ] Add helper functions: `NormalizeName()`, `GenerateFilename()`, etc.
- [ ] Update `CreateISORequest` to include version
- [ ] Add file type detection logic

### Step 2: Update Database Layer
- [ ] Create new migration with updated schema
- [ ] Update all CRUD operations
- [ ] Change unique constraint from `filename` to `(name, version, file_type)`
- [ ] Add `ISOExistsByNameVersionType()` method
- [ ] Update `GetISOBy*` methods

### Step 3: Update Worker
- [ ] Update `Process()` to use nested directory structure
- [ ] Create directories: `name/version/`
- [ ] Generate filename based on name + version + file type
- [ ] Update temp file paths

### Step 4: Update Manager
- [ ] No changes needed (works per-file)

### Step 5: Update Tests
- [ ] Fix all tests to use new schema
- [ ] Add tests for name normalization
- [ ] Add tests for file type detection
- [ ] Add tests for path generation
- [ ] Test duplicate detection with (name, version, file_type)

### Step 6: Update main.go
- [ ] Test with multiple versions
- [ ] Test with multiple file types
- [ ] Verify directory structure
- [ ] Verify download links

## Edge Cases & Validation

### 1. Invalid Characters in Name/Version
- Sanitize: Remove `/, \, :, *, ?, ", <, >, |`
- Validate version format (optional: semantic versioning)

### 2. Very Long Names
- Limit name to 100 chars
- Limit version to 50 chars

### 3. Filename Conflicts
- With (name, version, file_type) uniqueness, conflicts are prevented
- If file exists on disk but not in DB, log warning and overwrite

### 4. Directory Traversal Prevention
```go
// Prevent ../../../etc/passwd
func sanitizePath(path string) string {
    path = filepath.Clean(path)
    if strings.Contains(path, "..") {
        return ""
    }
    return path
}
```

### 5. Supported File Types
```go
var SupportedFileTypes = []string{
    "iso", "qcow2", "vmdk", "vdi",
    "img", "raw", "vhd", "vhdx"
}
```

### 6. Missing Version
- Make version required in CreateISORequest
- If not provided, try to extract from URL/filename
- Default to "latest" if cannot determine

## Backward Compatibility

### Option A: Fresh Start (Recommended for Phase 2)
- Drop existing database
- Start with new schema
- Simpler, cleaner

### Option B: Migration
- Migrate existing records
- Set version = "unknown" or extract from filename
- Detect file_type from existing filename
- Move files to new directory structure

## Benefits of This Approach

1. **Organization**: Easy to browse by OS and version
2. **Multiple Formats**: Support different virtualization platforms
3. **Independent Tracking**: Each file has own progress/status
4. **Scalability**: Can add more metadata (arch, edition) later
5. **Clean URLs**: `/images/alpine/3.19.1/alpine-3.19.1.iso`
6. **API Clarity**: Each file is a distinct resource

## Questions to Consider

1. **Should we add architecture field?** (x86_64, aarch64)
   - `alpine/3.19.1/x86_64/alpine-3.19.1-x86_64.iso`
   - Unique: (name, version, arch, file_type)

2. **Should we add edition field?** (desktop, server, minimal)
   - `ubuntu/24.04/desktop/ubuntu-24.04-desktop.iso`
   - Unique: (name, version, edition, file_type)

3. **Should version be hierarchical?**
   - `alpine/3/3.19/3.19.1/` (too deep?)
   - Keep flat: `alpine/3.19.1/`

4. **Should we validate version format?**
   - Enforce semantic versioning (3.19.1)?
   - Allow any string (rolling, latest, nightly)?

## Recommended: Start Simple

For Phase 2, implement:
- `name` (normalized)
- `version` (string, no format enforcement)
- `file_type` (auto-detected)
- Unique: `(name, version, file_type)`
- Structure: `{name}/{version}/{name}-{version}.{ext}`

Can add `arch` and `edition` later if needed.
