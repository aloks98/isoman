package api

import (
	_ "embed"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"linux-iso-manager/internal/db"
	"linux-iso-manager/internal/service"

	"github.com/gin-gonic/gin"
)

//go:embed templates/directory.html
var directoryTemplateContent string

// FileInfo represents a file in the directory listing.
type FileInfo struct {
	ModifiedTime time.Time
	Name         string
	Size         string
	Modified     string
	Path         string
	SizeBytes    int64
	IsDir        bool
}

// DirectoryHandlerConfig holds dependencies for the directory handler.
type DirectoryHandlerConfig struct {
	ISODir       string
	StatsService *service.StatsService
	DB           *db.DB
}

// isTrackableFile checks if the file should be tracked for download statistics.
// Only tracks actual ISO/image files, not checksum files.
func isTrackableFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	trackableExtensions := []string{".iso", ".qcow2", ".vmdk", ".img"}
	for _, trackable := range trackableExtensions {
		if ext == trackable {
			return true
		}
	}
	return false
}

// trackDownload records the download asynchronously.
func trackDownload(cfg *DirectoryHandlerConfig, filePath string) {
	// Look up the ISO by file path
	iso, err := cfg.DB.GetISOByFilePath(filePath)
	if err != nil {
		slog.Warn("failed to lookup ISO for download tracking", slog.String("path", filePath), slog.Any("error", err))
		return
	}
	if iso == nil {
		// ISO not found in database - might be a manually added file
		return
	}

	// Record the download
	if err := cfg.StatsService.RecordDownload(iso.ID); err != nil {
		slog.Warn("failed to record download", slog.String("iso_id", iso.ID), slog.Any("error", err))
	}
}

// DirectoryHandler serves Apache-style directory listing for /images/.
func DirectoryHandler(cfg *DirectoryHandlerConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the requested path (Gin includes leading slash in wildcard)
		requestPath := c.Param("filepath")

		// Remove leading slash if present
		if len(requestPath) > 0 && requestPath[0] == '/' {
			requestPath = requestPath[1:]
		}

		if requestPath == "" {
			requestPath = "."
		}

		// Construct full filesystem path
		fullPath := filepath.Join(cfg.ISODir, requestPath)

		// Check if path exists
		info, err := os.Stat(fullPath)
		if err != nil {
			c.String(http.StatusNotFound, "404 Not Found")
			return
		}

		// If it's a file, serve it directly
		if !info.IsDir() {
			// Track download if it's a trackable ISO file
			if isTrackableFile(requestPath) && cfg.StatsService != nil && cfg.DB != nil {
				go trackDownload(cfg, requestPath)
			}
			c.File(fullPath)
			return
		}

		// If it's a directory, show listing
		files, err := os.ReadDir(fullPath)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error reading directory")
			return
		}

		// Convert to FileInfo structs
		var fileInfos []FileInfo
		for _, file := range files {
			// Skip hidden files and temp directory
			if file.Name()[0] == '.' {
				continue
			}

			fileInfo, err := file.Info()
			if err != nil {
				continue
			}

			// Construct relative path for links
			relativePath := filepath.Join(requestPath, file.Name())

			// For directories, show "-" instead of directory entry size
			size := formatSize(fileInfo.Size())
			sizeBytes := fileInfo.Size()
			if file.IsDir() {
				size = "-"
				sizeBytes = 0
			}

			fileInfos = append(fileInfos, FileInfo{
				Name:         file.Name(),
				Size:         size,
				SizeBytes:    sizeBytes,
				Modified:     fileInfo.ModTime().Format("2006-01-02 15:04:05"),
				ModifiedTime: fileInfo.ModTime(),
				IsDir:        file.IsDir(),
				Path:         "/images/" + filepath.ToSlash(relativePath),
			})
		}

		// Sort by name (directories first, then files)
		sort.Slice(fileInfos, func(i, j int) bool {
			if fileInfos[i].IsDir != fileInfos[j].IsDir {
				return fileInfos[i].IsDir
			}
			return fileInfos[i].Name < fileInfos[j].Name
		})

		// Render HTML template
		renderDirectoryListing(c, requestPath, fileInfos)
	}
}

// formatSize converts bytes to human-readable format.
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// renderDirectoryListing renders the HTML directory listing.
func renderDirectoryListing(c *gin.Context, path string, files []FileInfo) {
	// Create template with custom functions
	tmpl := template.Must(template.New("directory").Funcs(template.FuncMap{
		"hasSuffix": func(s, suffix string) bool {
			return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
		},
	}).Parse(directoryTemplateContent))

	// Calculate parent path for "Parent Directory" link
	var parentPath string
	if path != "" {
		parentPath = "/images/" + filepath.ToSlash(filepath.Dir(path))
		if parentPath == "/images/." {
			parentPath = "/images/"
		}
	}

	data := gin.H{
		"Path":       path,
		"ParentPath": parentPath,
		"Files":      files,
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(c.Writer, data); err != nil {
		slog.Error("failed to execute template", slog.Any("error", err))
		ErrorResponse(c, http.StatusInternalServerError, ErrCodeInternalError, "Failed to generate directory listing")
	}
}

// WalkDirectory recursively walks a directory and returns all files.
func WalkDirectory(root string) ([]FileInfo, error) {
	var files []FileInfo

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden files and temp directory
		if d.Name()[0] == '.' {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		// Skip root directory itself
		if relPath == "." {
			return nil
		}

		fileInfo, err := d.Info()
		if err != nil {
			return err
		}

		// For directories, show "-" instead of directory entry size
		size := formatSize(fileInfo.Size())
		sizeBytes := fileInfo.Size()
		if d.IsDir() {
			size = "-"
			sizeBytes = 0
		}

		files = append(files, FileInfo{
			Name:         d.Name(),
			Size:         size,
			SizeBytes:    sizeBytes,
			Modified:     fileInfo.ModTime().Format("2006-01-02 15:04:05"),
			ModifiedTime: fileInfo.ModTime(),
			IsDir:        d.IsDir(),
			Path:         "/images/" + filepath.ToSlash(relPath),
		})

		return nil
	})

	return files, err
}
