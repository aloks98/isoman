# ISO Manager Implementation Plan - FINAL

## Data Model

### ISO Struct
```go
type ISO struct {
    ID           string     // UUID
    Name         string     // "alpine" (normalized: lowercase, hyphenated)
    Version      string     // "3.19.1" (any string: rolling, latest, nightly, etc.)
    Arch         string     // "x86_64", "aarch64", "arm64", etc.
    Edition      string     // "desktop", "server", "minimal", "" (optional)
    FileType     string     // "iso", "qcow2", "vmdk", etc. (auto-detected)

    Filename     string     // Computed: "alpine-3.19.1-minimal-x86_64.iso"
    FilePath     string     // Computed: "alpine/3.19.1/x86_64/alpine-3.19.1-minimal-x86_64.iso"
    DownloadLink string     // Computed: "/images/alpine/3.19.1/x86_64/alpine-3.19.1-minimal-x86_64.iso"

    SizeBytes    int64
    Checksum     string
    ChecksumType string
    DownloadURL  string
    ChecksumURL  string

    Status       ISOStatus
    Progress     int
    ErrorMessage string
    CreatedAt    time.Time
    CompletedAt  *time.Time
}
```

### Unique Constraint
`(name, version, arch, edition, file_type)`

### CreateISORequest
```go
type CreateISORequest struct {
    Name         string // "Alpine Linux" -> normalized to "alpine"
    Version      string // "3.19.1" (required)
    Arch         string // "x86_64" (required)
    Edition      string // "minimal" (optional, can be empty)
    DownloadURL  string // File type auto-detected from URL
    ChecksumURL  string // (optional)
    ChecksumType string // Default: "sha256"
}
```

## Storage Structure

```
data/isos/
├── alpine/
│   ├── 3.19.1/
│   │   ├── x86_64/
│   │   │   ├── alpine-3.19.1-x86_64.iso
│   │   │   ├── alpine-3.19.1-minimal-x86_64.iso
│   │   │   └── alpine-3.19.1-x86_64.qcow2
│   │   └── aarch64/
│   │       └── alpine-3.19.1-aarch64.iso
│   └── 3.18.0/
│       └── x86_64/
│           └── alpine-3.18.0-x86_64.iso
├── ubuntu/
│   └── 24.04/
│       └── x86_64/
│           ├── ubuntu-24.04-desktop-x86_64.iso
│           └── ubuntu-24.04-server-x86_64.iso
```

## Helper Functions

### 1. Name Normalization
```go
func NormalizeName(name string) string {
    // "Alpine Linux" -> "alpine"
    // "Ubuntu Server" -> "ubuntu-server"
    name = strings.ToLower(strings.TrimSpace(name))
    name = strings.ReplaceAll(name, " ", "-")
    name = regexp.MustCompile(`[^a-z0-9-]`).ReplaceAllString(name, "")
    return name
}
```

### 2. File Type Detection
```go
func DetectFileType(url string) (string, error) {
    ext := filepath.Ext(url)
    ext = strings.ToLower(strings.TrimPrefix(ext, "."))

    supported := []string{"iso", "qcow2", "vmdk", "vdi", "img", "raw", "vhd", "vhdx"}
    if contains(supported, ext) {
        return ext, nil
    }
    return "", fmt.Errorf("unsupported file type: %s", ext)
}
```

### 3. Filename Generation
```go
func GenerateFilename(name, version, edition, arch, fileType string) string {
    // alpine + 3.19.1 + minimal + x86_64 + iso
    // -> "alpine-3.19.1-minimal-x86_64.iso"

    // alpine + 3.19.1 + "" + x86_64 + iso
    // -> "alpine-3.19.1-x86_64.iso"

    parts := []string{name, version}
    if edition != "" {
        parts = append(parts, edition)
    }
    parts = append(parts, arch)

    filename := strings.Join(parts, "-")
    return fmt.Sprintf("%s.%s", filename, fileType)
}
```

### 4. Path Generation
```go
func GenerateFilePath(name, version, arch, filename string) string {
    // alpine + 3.19.1 + x86_64 + alpine-3.19.1-x86_64.iso
    // -> "alpine/3.19.1/x86_64/alpine-3.19.1-x86_64.iso"
    return filepath.Join(name, version, arch, filename)
}
```

### 5. Download Link Generation
```go
func GenerateDownloadLink(filePath string) string {
    // "alpine/3.19.1/x86_64/alpine-3.19.1-x86_64.iso"
    // -> "/images/alpine/3.19.1/x86_64/alpine-3.19.1-x86_64.iso"
    return "/images/" + filePath
}
```

## Database Schema

```sql
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
```

## API Examples

### Create Request
```json
{
  "name": "Alpine Linux",
  "version": "3.19.1",
  "arch": "x86_64",
  "edition": "minimal",
  "download_url": "https://example.com/alpine-minimal-3.19.1-x86_64.iso",
  "checksum_url": "https://example.com/alpine-minimal-3.19.1-x86_64.iso.sha256",
  "checksum_type": "sha256"
}
```

### Response
```json
{
  "id": "uuid-1",
  "name": "alpine",
  "version": "3.19.1",
  "arch": "x86_64",
  "edition": "minimal",
  "file_type": "iso",
  "filename": "alpine-3.19.1-minimal-x86_64.iso",
  "file_path": "alpine/3.19.1/x86_64/alpine-3.19.1-minimal-x86_64.iso",
  "download_link": "/images/alpine/3.19.1/x86_64/alpine-3.19.1-minimal-x86_64.iso",
  "size_bytes": 200000000,
  "checksum": "abc123...",
  "checksum_type": "sha256",
  "status": "complete",
  "progress": 100,
  "created_at": "2024-01-01T00:00:00Z",
  "completed_at": "2024-01-01T00:05:00Z"
}
```

## Implementation Steps

1. **Update models/iso.go**
   - Add new fields
   - Add helper functions
   - Update CreateISORequest

2. **Update db/sqlite.go**
   - New schema with compound unique key
   - Update all CRUD methods
   - Add ISOExistsByNameVersionArchEditionType()

3. **Update download/worker.go**
   - Create nested directories (name/version/arch/)
   - Use computed filename and file_path

4. **Update all tests**

5. **Update main.go for testing**

## Supported File Types (Initial)

- `iso` - Standard ISO image
- `qcow2` - QEMU/KVM format
- `vmdk` - VMware format
- `vdi` - VirtualBox format
- `img` - Raw disk image
- `raw` - Raw disk image
- `vhd` - Hyper-V format
- `vhdx` - Hyper-V format (newer)

Can add more later: ova, box (Vagrant), etc.
