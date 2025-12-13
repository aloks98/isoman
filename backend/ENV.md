# Environment Variables Reference

This document lists all configurable environment variables for the Linux ISO Manager application.

## Quick Reference

| Category | Variables |
|----------|-----------|
| [Server](#server-configuration) | PORT, READ_TIMEOUT_SEC, WRITE_TIMEOUT_SEC, IDLE_TIMEOUT_SEC, SHUTDOWN_TIMEOUT_SEC, CORS_ORIGINS |
| [Database](#database-configuration) | DB_PATH, DB_BUSY_TIMEOUT_MS, DB_JOURNAL_MODE, DB_MAX_OPEN_CONNS, DB_MAX_IDLE_CONNS, DB_CONN_MAX_LIFETIME_MIN, DB_CONN_MAX_IDLE_TIME_MIN |
| [Download](#download-configuration) | DATA_DIR, WORKER_COUNT, QUEUE_BUFFER, MAX_RETRIES, RETRY_DELAY_MS, BUFFER_SIZE, PROGRESS_UPDATE_INTERVAL_SEC, PROGRESS_PERCENT_THRESHOLD, CANCELLATION_WAIT_MS |
| [WebSocket](#websocket-configuration) | WS_BROADCAST_SIZE |
| [Logging](#logging-configuration) | LOG_LEVEL, LOG_FORMAT |

---

## Server Configuration

HTTP server and network settings.

| Variable | Type | Default | Description | Possible Values |
|----------|------|---------|-------------|-----------------|
| `PORT` | String | `8080` | HTTP server port | Any valid port (1-65535) |
| `READ_TIMEOUT_SEC` | Integer | `15` | Maximum duration for reading request (including body) | Any positive integer |
| `WRITE_TIMEOUT_SEC` | Integer | `15` | Maximum duration before timing out response writes | Any positive integer |
| `IDLE_TIMEOUT_SEC` | Integer | `60` | Max wait time for next request with keep-alives | Any positive integer |
| `SHUTDOWN_TIMEOUT_SEC` | Integer | `30` | Maximum duration to wait for graceful shutdown | Any positive integer |
| `CORS_ORIGINS` | String | `http://localhost:3000,`<br/>`http://localhost:5173,`<br/>`http://localhost:8080` | Comma-separated list of allowed CORS origins | Any valid HTTP/HTTPS URLs |

**Examples:**
```bash
PORT=3000
READ_TIMEOUT_SEC=30
CORS_ORIGINS=https://example.com,https://app.example.com
```

---

## Database Configuration

SQLite database settings.

| Variable | Type | Default | Description | Possible Values |
|----------|------|---------|-------------|-----------------|
| `DB_PATH` | String | _(empty)_ | Path to SQLite database file | Any valid file path<br/>_(auto-resolves to `${DATA_DIR}/db/isos.db`)_ |
| `DB_BUSY_TIMEOUT_MS` | Integer | `5000` | Max time to wait when database is locked (ms) | Any positive integer |
| `DB_JOURNAL_MODE` | String | `WAL` | SQLite journal mode for transaction logging | `WAL` _(recommended)_<br/>`DELETE`<br/>`TRUNCATE`<br/>`PERSIST`<br/>`MEMORY` _(testing only)_ |
| `DB_MAX_OPEN_CONNS` | Integer | `10` | Maximum number of open database connections | 1 to 100 |
| `DB_MAX_IDLE_CONNS` | Integer | `5` | Maximum number of idle connections in pool | 0 to `DB_MAX_OPEN_CONNS` |
| `DB_CONN_MAX_LIFETIME_MIN` | Integer | `60` | Max lifetime of a connection before closing (minutes) | Any positive integer |
| `DB_CONN_MAX_IDLE_TIME_MIN` | Integer | `10` | Max time a connection can be idle before closing (minutes) | Any positive integer |

**Examples:**
```bash
DB_PATH=/var/lib/isoman/db/isos.db
DB_JOURNAL_MODE=WAL
DB_MAX_OPEN_CONNS=25
```

**Notes:**
- WAL mode is recommended for better concurrency
- SQLite handles concurrent reads well but serializes writes

---

## Download Configuration

Settings for ISO download manager and workers.

| Variable | Type | Default | Description | Possible Values |
|----------|------|---------|-------------|-----------------|
| `DATA_DIR` | String | `./data` | Base directory for all data (ISOs, database) | Any valid directory path |
| `WORKER_COUNT` | Integer | `2` | Number of concurrent download workers | 1 to 10 |
| `QUEUE_BUFFER` | Integer | `100` | Size of the download queue buffer | 1 to 1000 |
| `MAX_RETRIES` | Integer | `3` | Max retry attempts for failed downloads | 0 to 10<br/>_(0 = no retries)_ |
| `RETRY_DELAY_MS` | Integer | `5000` | Delay between retry attempts (ms) | Any positive integer |
| `BUFFER_SIZE` | Integer | `65536` | Buffer size for downloading files (bytes) | 1024 to 1048576<br/>_(1 KB to 1 MB)_ |
| `PROGRESS_UPDATE_INTERVAL_SEC` | Integer | `1` | Min time interval between progress updates (seconds) | 1 to 60 |
| `PROGRESS_PERCENT_THRESHOLD` | Integer | `1` | Min percentage change to trigger progress update | 1 to 100 |
| `CANCELLATION_WAIT_MS` | Integer | `100` | Time to wait for download cancellation (ms) | 0 to 5000 |

**Examples:**
```bash
DATA_DIR=/var/lib/isoman/data
WORKER_COUNT=4
BUFFER_SIZE=131072  # 128 KB
MAX_RETRIES=5
```

**Notes:**
- More workers = more concurrent downloads but higher resource usage
- Larger buffers may improve performance for large files
- Progress updates sent when time interval OR percentage threshold is met

---

## WebSocket Configuration

Real-time communication settings.

| Variable | Type | Default | Description | Possible Values |
|----------|------|---------|-------------|-----------------|
| `WS_BROADCAST_SIZE` | Integer | `100` | Size of WebSocket broadcast channel buffer | 1 to 1000 |

**Examples:**
```bash
WS_BROADCAST_SIZE=200
```

**Notes:**
- Larger buffer prevents dropped messages under high load

---

## Logging Configuration

Application logging settings.

| Variable | Type | Default | Description | Possible Values |
|----------|------|---------|-------------|-----------------|
| `LOG_LEVEL` | String | `info` | Minimum log level to output | `debug` - Detailed debug info<br/>`info` - Informational messages<br/>`warn` - Warning messages only<br/>`error` - Error messages only |
| `LOG_FORMAT` | String | `text` | Log output format | `text` - Human-readable<br/>`json` - Structured JSON _(recommended for production)_ |

**Examples:**
```bash
LOG_LEVEL=debug
LOG_FORMAT=json
```

---

## Example Configurations

### Development (Default)

```bash
# Server
PORT=8080

# Database
DATA_DIR=./data
DB_JOURNAL_MODE=WAL

# Download
WORKER_COUNT=2
BUFFER_SIZE=65536

# Logging
LOG_LEVEL=info
LOG_FORMAT=text
```

### Production

```bash
# Server
PORT=8080
READ_TIMEOUT_SEC=30
WRITE_TIMEOUT_SEC=30
IDLE_TIMEOUT_SEC=120
SHUTDOWN_TIMEOUT_SEC=60
CORS_ORIGINS=https://isoman.example.com

# Database
DATA_DIR=/var/lib/isoman/data
DB_BUSY_TIMEOUT_MS=10000
DB_JOURNAL_MODE=WAL
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=10

# Download
WORKER_COUNT=4
MAX_RETRIES=5
RETRY_DELAY_MS=10000
BUFFER_SIZE=131072  # 128 KB

# WebSocket
WS_BROADCAST_SIZE=200

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
```

### High Performance

```bash
# Server
PORT=8080
READ_TIMEOUT_SEC=60
WRITE_TIMEOUT_SEC=60

# Database
DB_MAX_OPEN_CONNS=50
DB_MAX_IDLE_CONNS=25

# Download
WORKER_COUNT=8
QUEUE_BUFFER=500
BUFFER_SIZE=262144  # 256 KB
PROGRESS_PERCENT_THRESHOLD=5

# WebSocket
WS_BROADCAST_SIZE=500
```

### Testing/Development (Minimal Resources)

```bash
# Server
PORT=8080

# Database
DATA_DIR=./test-data
DB_MAX_OPEN_CONNS=5
DB_MAX_IDLE_CONNS=2

# Download
WORKER_COUNT=1
BUFFER_SIZE=32768  # 32 KB

# Logging
LOG_LEVEL=debug
LOG_FORMAT=text
```

---

## Auto-Resolved Paths

Some paths are auto-resolved if not explicitly set:

| Path | Default Resolution |
|------|-------------------|
| `DB_PATH` | `${DATA_DIR}/db/isos.db` (if empty) |
| ISO Storage | `${DATA_DIR}/isos/` (always) |
| Migrations | `./migrations` (internal, not configurable) |

---

## Performance Tuning Guide

| Scenario | Recommended Settings |
|----------|---------------------|
| **Fast bulk downloads** | `WORKER_COUNT=4-8` |
| **Large files (>1GB)** | `BUFFER_SIZE=131072-262144` (128-256 KB) |
| **High concurrent access** | `DB_MAX_OPEN_CONNS=20-50` |
| **Many WebSocket clients** | `WS_BROADCAST_SIZE=200-500` |
| **Low resource environment** | `WORKER_COUNT=1`, `BUFFER_SIZE=32768` (32 KB) |

---

## Resource Considerations

| Component | RAM Usage |
|-----------|-----------|
| Each worker | ~10 MB + network bandwidth |
| Buffer size | Directly affects RAM per active download |
| Database connection | ~1 MB each |
| WebSocket connection | ~50 KB each |

---

## Security Best Practices

| Setting | Recommendation |
|---------|---------------|
| `CORS_ORIGINS` | Set to specific domains in production (never use `*`) |
| Timeouts | Set appropriate values to prevent resource exhaustion |
| `WORKER_COUNT` | Limit to prevent bandwidth saturation |
| `LOG_FORMAT` | Use `json` in production for better monitoring |
| `LOG_LEVEL` | Use `info` or `warn` in production (not `debug`) |
