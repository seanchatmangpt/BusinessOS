package handlers

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/services"
)

// CustomModulesHandler handles custom module operations
type CustomModulesHandler struct {
	service *services.CustomModuleService
	logger  *slog.Logger
}

// NewCustomModulesHandler creates a new custom modules handler
func NewCustomModulesHandler(pool *pgxpool.Pool, logger *slog.Logger) *CustomModulesHandler {
	return &CustomModulesHandler{
		service: services.NewCustomModuleService(pool, logger),
		logger:  logger,
	}
}

// CreateModule creates a new custom module
// POST /api/modules
func (h *CustomModulesHandler) CreateModule(c *gin.Context) {
	var req services.CreateModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	userID, workspaceID, err := h.getAuthContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	module, err := h.service.CreateModule(c.Request.Context(), workspaceID, userID, req)
	if err != nil {
		h.logger.Error("Failed to create module", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create module", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, module)
}

// GetModule retrieves a single module
// GET /api/modules/:id
func (h *CustomModulesHandler) GetModule(c *gin.Context) {
	moduleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid module ID"})
		return
	}

	module, err := h.service.GetModule(c.Request.Context(), moduleID)
	if err != nil {
		h.logger.Error("Failed to get module", "error", err, "module_id", moduleID)
		c.JSON(http.StatusNotFound, gin.H{"error": "Module not found"})
		return
	}

	c.JSON(http.StatusOK, module)
}

// ListModules lists modules in a workspace with pagination and filters
// GET /api/modules?limit=20&offset=0&category=utility
func (h *CustomModulesHandler) ListModules(c *gin.Context) {
	_, workspaceID, err := h.getAuthContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	modules, err := h.service.ListModules(c.Request.Context(), workspaceID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to list modules", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list modules"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"modules": modules,
		"count":   len(modules),
		"limit":   limit,
		"offset":  offset,
	})
}

// UpdateModule updates an existing module
// PUT /api/modules/:id
func (h *CustomModulesHandler) UpdateModule(c *gin.Context) {
	moduleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid module ID"})
		return
	}

	var req services.UpdateModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	userID, _, err := h.getAuthContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	module, err := h.service.UpdateModule(c.Request.Context(), moduleID, userID, req)
	if err != nil {
		h.logger.Error("Failed to update module", "error", err, "module_id", moduleID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update module", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, module)
}

// DeleteModule deletes a module
// DELETE /api/modules/:id
func (h *CustomModulesHandler) DeleteModule(c *gin.Context) {
	moduleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid module ID"})
		return
	}

	userID, _, err := h.getAuthContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	err = h.service.DeleteModule(c.Request.Context(), moduleID, userID)
	if err != nil {
		h.logger.Error("Failed to delete module", "error", err, "module_id", moduleID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete module", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Module deleted successfully"})
}

// PublishModule publishes a module to the marketplace
// POST /api/modules/:id/publish
func (h *CustomModulesHandler) PublishModule(c *gin.Context) {
	moduleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid module ID"})
		return
	}

	userID, _, err := h.getAuthContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	err = h.service.PublishModule(c.Request.Context(), moduleID, userID)
	if err != nil {
		h.logger.Error("Failed to publish module", "error", err, "module_id", moduleID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish module", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Module published successfully"})
}

// InstallModule installs a module to the current workspace
// POST /api/modules/:id/install
func (h *CustomModulesHandler) InstallModule(c *gin.Context) {
	moduleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid module ID"})
		return
	}

	userID, workspaceID, err := h.getAuthContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Get module to install
	module, err := h.service.GetModule(c.Request.Context(), moduleID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Module not found"})
		return
	}

	// Create installation record
	_, err = h.service.CreateInstallation(c.Request.Context(), moduleID, workspaceID, userID, module.Version)
	if err != nil {
		h.logger.Error("Failed to install module", "error", err, "module_id", moduleID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to install module", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Module installed successfully"})
}

// ShareModule shares a module with another user/workspace
// POST /api/modules/:id/share
func (h *CustomModulesHandler) ShareModule(c *gin.Context) {
	moduleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid module ID"})
		return
	}

	var req struct {
		ShareWithUserID      *string `json:"share_with_user_id"`
		ShareWithWorkspaceID *string `json:"share_with_workspace_id"`
		ShareWithEmail       *string `json:"share_with_email"`
		CanView              bool    `json:"can_view"`
		CanInstall           bool    `json:"can_install"`
		CanModify            bool    `json:"can_modify"`
		CanReshare           bool    `json:"can_reshare"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	userID, _, err := h.getAuthContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Parse UUIDs if provided
	var shareWithUserID, shareWithWorkspaceID *uuid.UUID
	if req.ShareWithUserID != nil {
		uid, err := uuid.Parse(*req.ShareWithUserID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}
		shareWithUserID = &uid
	}
	if req.ShareWithWorkspaceID != nil {
		wid, err := uuid.Parse(*req.ShareWithWorkspaceID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
			return
		}
		shareWithWorkspaceID = &wid
	}

	share, err := h.service.CreateShare(c.Request.Context(), services.CreateShareRequest{
		ModuleID:              moduleID,
		SharedWithUserID:      shareWithUserID,
		SharedWithWorkspaceID: shareWithWorkspaceID,
		SharedWithEmail:       req.ShareWithEmail,
		CanView:               req.CanView,
		CanInstall:            req.CanInstall,
		CanModify:             req.CanModify,
		CanReshare:            req.CanReshare,
		SharedBy:              userID,
	})

	if err != nil {
		h.logger.Error("Failed to share module", "error", err, "module_id", moduleID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to share module", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, share)
}

// ListInstalledModules lists all installed modules in the workspace
// GET /api/modules/installed
func (h *CustomModulesHandler) ListInstalledModules(c *gin.Context) {
	_, workspaceID, err := h.getAuthContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	installations, err := h.service.ListInstallations(c.Request.Context(), workspaceID)
	if err != nil {
		h.logger.Error("Failed to list installations", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list installed modules"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"installations": installations,
		"count":         len(installations),
	})
}

// GetModuleStats returns stats about a module (install count, etc.)
// GET /api/modules/stats?id=uuid
func (h *CustomModulesHandler) GetModuleStats(c *gin.Context) {
	moduleIDStr := c.Query("id")
	if moduleIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Module ID required"})
		return
	}

	moduleID, err := uuid.Parse(moduleIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid module ID"})
		return
	}

	module, err := h.service.GetModule(c.Request.Context(), moduleID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Module not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"module_id":     module.ID,
		"install_count": module.InstallCount,
		"star_count":    module.StarCount,
		"version":       module.Version,
		"is_published":  module.IsPublished,
	})
}

// GetPopularModules returns popular public modules
// GET /api/modules/popular?limit=10
func (h *CustomModulesHandler) GetPopularModules(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	modules, err := h.service.SearchModules(c.Request.Context(), "", limit, 0)
	if err != nil {
		h.logger.Error("Failed to get popular modules", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get popular modules"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"modules": modules,
		"count":   len(modules),
	})
}

// ExportModule exports a module as a ZIP file
// GET /api/modules/export/:id
func (h *CustomModulesHandler) ExportModule(c *gin.Context) {
	moduleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid module ID"})
		return
	}

	module, err := h.service.GetModule(c.Request.Context(), moduleID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Module not found"})
		return
	}

	// Create ZIP file in memory
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	// Add manifest.json
	manifestFile, err := zipWriter.Create("manifest.json")
	if err != nil {
		h.logger.Error("Failed to create manifest file", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to export module"})
		return
	}
	manifestJSON, _ := json.MarshalIndent(module.Manifest, "", "  ")
	manifestFile.Write(manifestJSON)

	// Add config.json
	configFile, err := zipWriter.Create("config.json")
	if err != nil {
		h.logger.Error("Failed to create config file", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to export module"})
		return
	}
	configJSON, _ := json.MarshalIndent(module.Config, "", "  ")
	configFile.Write(configJSON)

	// Add metadata.json
	metadataFile, err := zipWriter.Create("metadata.json")
	if err != nil {
		h.logger.Error("Failed to create metadata file", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to export module"})
		return
	}
	metadata := map[string]interface{}{
		"name":        module.Name,
		"slug":        module.Slug,
		"description": module.Description,
		"category":    module.Category,
		"version":     module.Version,
		"icon":        module.Icon,
		"tags":        module.Tags,
		"keywords":    module.Keywords,
	}
	metadataJSON, _ := json.MarshalIndent(metadata, "", "  ")
	metadataFile.Write(metadataJSON)

	zipWriter.Close()

	// Set headers and return ZIP
	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.zip", module.Slug))
	c.Data(http.StatusOK, "application/zip", buf.Bytes())
}

// ImportModule imports a module from a ZIP file
// POST /api/modules/import
func (h *CustomModulesHandler) ImportModule(c *gin.Context) {
	userID, workspaceID, err := h.getAuthContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Parse multipart form
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer file.Close()

	// Read file into memory
	fileData, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	// Open ZIP
	zipReader, err := zip.NewReader(bytes.NewReader(fileData), int64(len(fileData)))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ZIP file"})
		return
	}

	// Extract files
	var manifestData, configData, metadataData []byte

	for _, file := range zipReader.File {
		rc, err := file.Open()
		if err != nil {
			continue
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			continue
		}

		switch file.Name {
		case "manifest.json":
			manifestData = data
		case "config.json":
			configData = data
		case "metadata.json":
			metadataData = data
		}
	}

	if manifestData == nil || metadataData == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid module ZIP: missing required files"})
		return
	}

	// Parse JSON
	var manifest, config, metadata map[string]interface{}
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid manifest JSON"})
		return
	}
	if err := json.Unmarshal(metadataData, &metadata); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid metadata JSON"})
		return
	}
	if configData != nil {
		json.Unmarshal(configData, &config)
	}

	// Create module
	req := services.CreateModuleRequest{
		Name:        metadata["name"].(string),
		Description: metadata["description"].(string),
		Category:    metadata["category"].(string),
		Manifest:    manifest,
		Config:      config,
		Icon:        metadata["icon"].(string),
		Tags:        []string{},
		Keywords:    []string{},
	}

	// Parse tags and keywords if present
	if tags, ok := metadata["tags"].([]interface{}); ok {
		for _, tag := range tags {
			if tagStr, ok := tag.(string); ok {
				req.Tags = append(req.Tags, tagStr)
			}
		}
	}
	if keywords, ok := metadata["keywords"].([]interface{}); ok {
		for _, keyword := range keywords {
			if kwStr, ok := keyword.(string); ok {
				req.Keywords = append(req.Keywords, kwStr)
			}
		}
	}

	module, err := h.service.CreateModule(c.Request.Context(), workspaceID, userID, req)
	if err != nil {
		h.logger.Error("Failed to import module", "error", err, "file", header.Filename)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to import module", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Module imported successfully",
		"module":  module,
	})
}

// Helper function to get auth context
func (h *CustomModulesHandler) getAuthContext(c *gin.Context) (userID uuid.UUID, workspaceID uuid.UUID, err error) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		return uuid.Nil, uuid.Nil, fmt.Errorf("unauthorized: user ID not found")
	}

	// Handle both string and uuid.UUID types
	switch v := userIDVal.(type) {
	case string:
		userID, err = uuid.Parse(v)
		if err != nil {
			return uuid.Nil, uuid.Nil, fmt.Errorf("invalid user ID")
		}
	case uuid.UUID:
		userID = v
	default:
		return uuid.Nil, uuid.Nil, fmt.Errorf("invalid user ID type")
	}

	// For now, use a default workspace ID or get from context
	// In production, this should come from the request context
	workspaceIDVal, exists := c.Get("workspaceID")
	if !exists {
		// Try to get from query param as fallback
		workspaceIDStr := c.Query("workspace_id")
		if workspaceIDStr != "" {
			workspaceID, err = uuid.Parse(workspaceIDStr)
			if err != nil {
				return uuid.Nil, uuid.Nil, fmt.Errorf("invalid workspace ID")
			}
		} else {
			return uuid.Nil, uuid.Nil, fmt.Errorf("workspace ID not found")
		}
	} else {
		switch v := workspaceIDVal.(type) {
		case string:
			workspaceID, err = uuid.Parse(v)
			if err != nil {
				return uuid.Nil, uuid.Nil, fmt.Errorf("invalid workspace ID")
			}
		case uuid.UUID:
			workspaceID = v
		default:
			return uuid.Nil, uuid.Nil, fmt.Errorf("invalid workspace ID type")
		}
	}

	return userID, workspaceID, nil
}
