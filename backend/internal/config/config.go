package config

import (
	"linux-iso-manager/internal/constants"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Download DownloadConfig
	WebSocket WebSocketConfig
	Log      LogConfig
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	CORSOrigins     []string
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Path            string
	BusyTimeout     time.Duration
	JournalMode     string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// DownloadConfig holds download manager configuration
type DownloadConfig struct {
	DataDir                string
	WorkerCount            int
	QueueBuffer            int
	MaxRetries             int
	RetryDelay             time.Duration
	BufferSize             int
	ProgressUpdateInterval time.Duration
	ProgressPercentThreshold int
	CancellationWait       time.Duration
}

// WebSocketConfig holds WebSocket configuration
type WebSocketConfig struct {
	BroadcastChannelSize int
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level  string // debug, info, warn, error
	Format string // json, text
}

// Load loads configuration from environment variables with defaults
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:            getEnv("PORT", constants.DefaultPort),
			ReadTimeout:     getDuration("READ_TIMEOUT_SEC", constants.DefaultReadTimeoutSec) * time.Second,
			WriteTimeout:    getDuration("WRITE_TIMEOUT_SEC", constants.DefaultWriteTimeoutSec) * time.Second,
			IdleTimeout:     getDuration("IDLE_TIMEOUT_SEC", constants.DefaultIdleTimeoutSec) * time.Second,
			ShutdownTimeout: getDuration("SHUTDOWN_TIMEOUT_SEC", constants.DefaultShutdownTimeoutSec) * time.Second,
			CORSOrigins:     getCORSOrigins(),
		},
		Database: DatabaseConfig{
			Path:            getEnv("DB_PATH", ""),
			BusyTimeout:     getDuration("DB_BUSY_TIMEOUT_MS", constants.DefaultBusyTimeoutMs) * time.Millisecond,
			JournalMode:     getEnv("DB_JOURNAL_MODE", constants.DefaultJournalMode),
			MaxOpenConns:    getInt("DB_MAX_OPEN_CONNS", constants.DefaultMaxOpenConns),
			MaxIdleConns:    getInt("DB_MAX_IDLE_CONNS", constants.DefaultMaxIdleConns),
			ConnMaxLifetime: getDuration("DB_CONN_MAX_LIFETIME_MIN", constants.DefaultConnMaxLifetimeMin) * time.Minute,
			ConnMaxIdleTime: getDuration("DB_CONN_MAX_IDLE_TIME_MIN", constants.DefaultConnMaxIdleTimeMin) * time.Minute,
		},
		Download: DownloadConfig{
			DataDir:                  getEnv("DATA_DIR", "./data"),
			WorkerCount:              getInt("WORKER_COUNT", constants.DefaultWorkerCount),
			QueueBuffer:              getInt("QUEUE_BUFFER", constants.DefaultQueueBuffer),
			MaxRetries:               getInt("MAX_RETRIES", constants.DefaultMaxRetries),
			RetryDelay:               getDuration("RETRY_DELAY_MS", constants.DefaultRetryDelayMs) * time.Millisecond,
			BufferSize:               getInt("BUFFER_SIZE", constants.DefaultDownloadBufferSize),
			ProgressUpdateInterval:   getDuration("PROGRESS_UPDATE_INTERVAL_SEC", 1) * time.Second,
			ProgressPercentThreshold: getInt("PROGRESS_PERCENT_THRESHOLD", constants.DefaultProgressPercentThreshold),
			CancellationWait:         getDuration("CANCELLATION_WAIT_MS", constants.DefaultCancellationWaitMs) * time.Millisecond,
		},
		WebSocket: WebSocketConfig{
			BroadcastChannelSize: getInt("WS_BROADCAST_SIZE", constants.DefaultBroadcastChannelSize),
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "text"),
		},
	}
}

// getEnv returns environment variable value or default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getInt returns environment variable as int or default
func getInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getDuration returns environment variable as int (for duration calculation) or default
func getDuration(key string, defaultValue int) time.Duration {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return time.Duration(intValue)
		}
	}
	return time.Duration(defaultValue)
}

// getCORSOrigins returns CORS origins from environment or defaults
func getCORSOrigins() []string {
	if origins := os.Getenv("CORS_ORIGINS"); origins != "" {
		return strings.Split(origins, ",")
	}
	// Default CORS origins for development
	return []string{
		"http://localhost:3000",  // React dev server (npm/yarn)
		"http://localhost:5173",  // Vite dev server (default)
		"http://localhost:8080",  // Same origin
	}
}
