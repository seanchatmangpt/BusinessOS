package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/services"
)

// AppProfilerHandler handles application profiling operations
type AppProfilerHandler struct {
	profiler *services.AppProfilerService
}

// NewAppProfilerHandler creates a new app profiler handler
func NewAppProfilerHandler(profiler *services.AppProfilerService) *AppProfilerHandler {
	return &AppProfilerHandler{
		profiler: profiler,
	}
}

// ProfileApplication profiles a codebase
func (h *AppProfilerHandler) ProfileApplication(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		userID = "default-user"
	}

	var req struct {
		Name     string                   `json:"name"`
		RootPath string                   `json:"root_path"`
		Options  *services.ProfileOptions `json:"options,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}

	if req.RootPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Root path is required"})
		return
	}

	profile, err := h.profiler.ProfileApplication(ctx, userID, req.RootPath, req.Name, req.Options)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to profile application: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, profile)
}

// GetProfile retrieves an application profile
func (h *AppProfilerHandler) GetProfile(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		userID = "default-user"
	}

	name := c.Param("name")

	profile, err := h.profiler.GetProfile(ctx, userID, name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, profile)
}

// ListProfiles lists all profiles for a user
func (h *AppProfilerHandler) ListProfiles(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		userID = "default-user"
	}

	profiles, err := h.profiler.ListProfiles(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list profiles: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, profiles)
}

// RefreshProfile re-profiles an application
func (h *AppProfilerHandler) RefreshProfile(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		userID = "default-user"
	}

	name := c.Param("name")

	// Get existing profile to get the root path
	existing, err := h.profiler.GetProfile(ctx, userID, name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found: " + err.Error()})
		return
	}

	// Re-profile
	profile, err := h.profiler.ProfileApplication(ctx, userID, existing.RootPath, name, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to refresh profile: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, profile)
}

// GetProfileComponents retrieves components for a profile
func (h *AppProfilerHandler) GetProfileComponents(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		userID = "default-user"
	}

	name := c.Param("name")

	profile, err := h.profiler.GetProfile(ctx, userID, name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, profile.Components)
}

// GetProfileEndpoints retrieves API endpoints for a profile
func (h *AppProfilerHandler) GetProfileEndpoints(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		userID = "default-user"
	}

	name := c.Param("name")

	profile, err := h.profiler.GetProfile(ctx, userID, name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, profile.APIEndpoints)
}

// GetProfileStructure retrieves directory structure for a profile
func (h *AppProfilerHandler) GetProfileStructure(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		userID = "default-user"
	}

	name := c.Param("name")

	profile, err := h.profiler.GetProfile(ctx, userID, name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, profile.StructureTree)
}

// GetProfileModules retrieves modules for a profile
func (h *AppProfilerHandler) GetProfileModules(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		userID = "default-user"
	}

	name := c.Param("name")

	profile, err := h.profiler.GetProfile(ctx, userID, name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, profile.Modules)
}

// GetProfileTechStack retrieves tech stack for a profile
func (h *AppProfilerHandler) GetProfileTechStack(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		userID = "default-user"
	}

	name := c.Param("name")

	profile, err := h.profiler.GetProfile(ctx, userID, name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tech_stack":  profile.TechStack,
		"languages":   profile.Languages,
		"frameworks":  profile.Frameworks,
		"conventions": profile.Conventions,
	})
}

// RegisterAppProfilerRoutes registers app profiler routes on a Gin router group
func RegisterAppProfilerRoutes(r *gin.RouterGroup, handler *AppProfilerHandler) {
	profiles := r.Group("/app-profiles")
	{
		profiles.POST("", handler.ProfileApplication)
		profiles.GET("", handler.ListProfiles)
		profiles.GET("/:name", handler.GetProfile)
		profiles.POST("/:name/refresh", handler.RefreshProfile)
		profiles.GET("/:name/components", handler.GetProfileComponents)
		profiles.GET("/:name/endpoints", handler.GetProfileEndpoints)
		profiles.GET("/:name/structure", handler.GetProfileStructure)
		profiles.GET("/:name/modules", handler.GetProfileModules)
		profiles.GET("/:name/tech-stack", handler.GetProfileTechStack)
	}
}
