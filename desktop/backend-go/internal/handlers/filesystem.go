package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// FileItem represents a file or directory
type FileItem struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Type      string     `json:"type"` // "file" or "folder"
	Path      string     `json:"path"`
	Size      int64      `json:"size,omitempty"`
	Modified  *time.Time `json:"modified,omitempty"`
	Extension string     `json:"extension,omitempty"`
	IsHidden  bool       `json:"isHidden"`
}

// ListDirectoryRequest represents the request for listing a directory
type ListDirectoryRequest struct {
	Path       string `json:"path" form:"path"`
	ShowHidden bool   `json:"showHidden" form:"showHidden"`
}

// ListDirectoryResponse represents the response for listing a directory
type ListDirectoryResponse struct {
	Path      string     `json:"path"`
	Items     []FileItem `json:"items"`
	ParentDir string     `json:"parentDir,omitempty"`
}

// FileContentResponse represents the response for reading a file
type FileContentResponse struct {
	Path     string `json:"path"`
	Name     string `json:"name"`
	Content  string `json:"content"`
	Size     int64  `json:"size"`
	MimeType string `json:"mimeType"`
}

// ListDirectory lists the contents of a directory
func (h *Handlers) ListDirectory(c *gin.Context) {
	var req ListDirectoryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
		return
	}

	// Default to home directory if no path specified
	dirPath := req.Path
	if dirPath == "" || dirPath == "~" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get home directory"})
			return
		}
		dirPath = homeDir
	}

	// Expand ~ to home directory
	if strings.HasPrefix(dirPath, "~/") {
		homeDir, _ := os.UserHomeDir()
		dirPath = filepath.Join(homeDir, dirPath[2:])
	}

	// Clean the path to prevent path traversal attacks
	dirPath = filepath.Clean(dirPath)

	// Check if path exists and is a directory
	info, err := os.Stat(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Directory not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to access directory"})
		return
	}

	if !info.IsDir() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Path is not a directory"})
		return
	}

	// Read directory contents
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read directory"})
		return
	}

	items := make([]FileItem, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		isHidden := strings.HasPrefix(name, ".")

		// Skip hidden files unless requested
		if isHidden && !req.ShowHidden {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue // Skip files we can't access
		}

		item := FileItem{
			ID:       generateFileID(filepath.Join(dirPath, name)),
			Name:     name,
			Path:     filepath.Join(dirPath, name),
			IsHidden: isHidden,
		}

		if entry.IsDir() {
			item.Type = "folder"
		} else {
			item.Type = "file"
			item.Size = info.Size()
			item.Extension = strings.TrimPrefix(filepath.Ext(name), ".")
		}

		modTime := info.ModTime()
		item.Modified = &modTime

		items = append(items, item)
	}

	// Sort: folders first, then alphabetically (case-insensitive)
	sort.Slice(items, func(i, j int) bool {
		if items[i].Type != items[j].Type {
			return items[i].Type == "folder"
		}
		return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
	})

	// Get parent directory
	parentDir := filepath.Dir(dirPath)
	if parentDir == dirPath {
		parentDir = "" // At root
	}

	c.JSON(http.StatusOK, ListDirectoryResponse{
		Path:      dirPath,
		Items:     items,
		ParentDir: parentDir,
	})
}

// ReadFile reads the content of a file
func (h *Handlers) ReadFile(c *gin.Context) {
	filePath := c.Query("path")
	if filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Path is required"})
		return
	}

	// Expand ~ to home directory
	if strings.HasPrefix(filePath, "~/") {
		homeDir, _ := os.UserHomeDir()
		filePath = filepath.Join(homeDir, filePath[2:])
	}

	// Clean the path
	filePath = filepath.Clean(filePath)

	// Check if file exists
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to access file"})
		return
	}

	if info.IsDir() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Path is a directory, not a file"})
		return
	}

	// Limit file size for reading (10MB max for text preview)
	const maxReadSize = 10 * 1024 * 1024
	if info.Size() > maxReadSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File too large for preview"})
		return
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	// Detect MIME type
	mimeType := http.DetectContentType(content)

	c.JSON(http.StatusOK, FileContentResponse{
		Path:     filePath,
		Name:     filepath.Base(filePath),
		Content:  string(content),
		Size:     info.Size(),
		MimeType: mimeType,
	})
}

// DownloadFile downloads a file
func (h *Handlers) DownloadFile(c *gin.Context) {
	filePath := c.Query("path")
	if filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Path is required"})
		return
	}

	// Expand ~ to home directory
	if strings.HasPrefix(filePath, "~/") {
		homeDir, _ := os.UserHomeDir()
		filePath = filepath.Join(homeDir, filePath[2:])
	}

	// Clean the path
	filePath = filepath.Clean(filePath)

	// Check if file exists
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to access file"})
		return
	}

	if info.IsDir() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot download a directory"})
		return
	}

	// Serve the file
	c.File(filePath)
}

// GetFileInfo returns information about a file or directory
func (h *Handlers) GetFileInfo(c *gin.Context) {
	filePath := c.Query("path")
	if filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Path is required"})
		return
	}

	// Expand ~ to home directory
	if strings.HasPrefix(filePath, "~/") {
		homeDir, _ := os.UserHomeDir()
		filePath = filepath.Join(homeDir, filePath[2:])
	}

	// Clean the path
	filePath = filepath.Clean(filePath)

	// Get file info
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Path not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to access path"})
		return
	}

	name := filepath.Base(filePath)
	modTime := info.ModTime()

	item := FileItem{
		ID:       generateFileID(filePath),
		Name:     name,
		Path:     filePath,
		IsHidden: strings.HasPrefix(name, "."),
		Modified: &modTime,
	}

	if info.IsDir() {
		item.Type = "folder"
		// Count items in directory
		entries, err := os.ReadDir(filePath)
		if err == nil {
			item.Size = int64(len(entries))
		}
	} else {
		item.Type = "file"
		item.Size = info.Size()
		item.Extension = strings.TrimPrefix(filepath.Ext(name), ".")
	}

	c.JSON(http.StatusOK, item)
}

// CreateDirectory creates a new directory
func (h *Handlers) CreateDirectory(c *gin.Context) {
	var req struct {
		Path string `json:"path" binding:"required"`
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Sanitize directory name
	if strings.Contains(req.Name, "/") || strings.Contains(req.Name, "\\") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid directory name"})
		return
	}

	// Expand ~ to home directory
	basePath := req.Path
	if strings.HasPrefix(basePath, "~/") {
		homeDir, _ := os.UserHomeDir()
		basePath = filepath.Join(homeDir, basePath[2:])
	}

	newDirPath := filepath.Join(basePath, req.Name)

	// Check if already exists
	if _, err := os.Stat(newDirPath); err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Directory already exists"})
		return
	}

	// Create directory
	if err := os.MkdirAll(newDirPath, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create directory"})
		return
	}

	// Return info about new directory
	info, _ := os.Stat(newDirPath)
	modTime := info.ModTime()

	c.JSON(http.StatusCreated, FileItem{
		ID:       generateFileID(newDirPath),
		Name:     req.Name,
		Type:     "folder",
		Path:     newDirPath,
		IsHidden: strings.HasPrefix(req.Name, "."),
		Modified: &modTime,
	})
}

// DeleteFile deletes a file or empty directory
func (h *Handlers) DeleteFileOrDir(c *gin.Context) {
	filePath := c.Query("path")
	if filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Path is required"})
		return
	}

	// Expand ~ to home directory
	if strings.HasPrefix(filePath, "~/") {
		homeDir, _ := os.UserHomeDir()
		filePath = filepath.Join(homeDir, filePath[2:])
	}

	// Clean the path
	filePath = filepath.Clean(filePath)

	// Safety: prevent deleting system directories
	homeDir, _ := os.UserHomeDir()
	dangerousPaths := []string{"/", "/bin", "/usr", "/etc", "/System", homeDir}
	for _, dangerous := range dangerousPaths {
		if filePath == dangerous {
			c.JSON(http.StatusForbidden, gin.H{"error": "Cannot delete this path"})
			return
		}
	}

	// Check if path exists
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Path not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to access path"})
		return
	}

	// For directories, only allow deleting if empty (use recursive flag for non-empty)
	if info.IsDir() {
		entries, err := os.ReadDir(filePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read directory"})
			return
		}
		if len(entries) > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Directory is not empty. Use recursive delete."})
			return
		}
	}

	// Delete the file/directory
	if err := os.Remove(filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Deleted successfully", "path": filePath})
}

// UploadFile handles file uploads
func (h *Handlers) UploadFile(c *gin.Context) {
	destPath := c.PostForm("path")
	if destPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Destination path is required"})
		return
	}

	// Expand ~ to home directory
	if strings.HasPrefix(destPath, "~/") {
		homeDir, _ := os.UserHomeDir()
		destPath = filepath.Join(homeDir, destPath[2:])
	}

	// Get uploaded file
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer file.Close()

	// Create full file path
	fullPath := filepath.Join(destPath, header.Filename)

	// Create destination file
	dst, err := os.Create(fullPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create file"})
		return
	}
	defer dst.Close()

	// Copy content
	if _, err := io.Copy(dst, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// Get file info
	info, _ := os.Stat(fullPath)
	modTime := info.ModTime()

	c.JSON(http.StatusCreated, FileItem{
		ID:        generateFileID(fullPath),
		Name:      header.Filename,
		Type:      "file",
		Path:      fullPath,
		Size:      info.Size(),
		Modified:  &modTime,
		Extension: strings.TrimPrefix(filepath.Ext(header.Filename), "."),
		IsHidden:  strings.HasPrefix(header.Filename, "."),
	})
}

// GetQuickAccessPaths returns commonly used paths
func (h *Handlers) GetQuickAccessPaths(c *gin.Context) {
	homeDir, _ := os.UserHomeDir()

	paths := []struct {
		Name string `json:"name"`
		Path string `json:"path"`
		Icon string `json:"icon"`
	}{
		{Name: "Home", Path: homeDir, Icon: "home"},
		{Name: "Desktop", Path: filepath.Join(homeDir, "Desktop"), Icon: "desktop"},
		{Name: "Documents", Path: filepath.Join(homeDir, "Documents"), Icon: "document"},
		{Name: "Downloads", Path: filepath.Join(homeDir, "Downloads"), Icon: "download"},
		{Name: "Pictures", Path: filepath.Join(homeDir, "Pictures"), Icon: "image"},
		{Name: "Music", Path: filepath.Join(homeDir, "Music"), Icon: "music"},
		{Name: "Videos", Path: filepath.Join(homeDir, "Movies"), Icon: "video"},
	}

	// Filter to only paths that exist
	validPaths := make([]struct {
		Name string `json:"name"`
		Path string `json:"path"`
		Icon string `json:"icon"`
	}, 0)

	for _, p := range paths {
		if _, err := os.Stat(p.Path); err == nil {
			validPaths = append(validPaths, p)
		}
	}

	c.JSON(http.StatusOK, gin.H{"paths": validPaths})
}

// generateFileID creates a unique ID for a file path
func generateFileID(path string) string {
	// Use base64 of path for a stable ID
	return fmt.Sprintf("file_%x", path)
}
