package handlers

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/services"
	"github.com/rhl/businessos-backend/internal/utils"
	"gopkg.in/yaml.v3"
)

// defaultFocusTemplatesPath is the canonical location of the YAML file that
// overrides the hardcoded template list. The file is optional; if absent the
// hardcoded slice is returned instead.
const defaultFocusTemplatesPath = "configs/focus_templates.yaml"

// FocusHandler handles focus mode templates and settings.
type FocusHandler struct {
	pool          *pgxpool.Pool
	templatesPath string // YAML file path; empty → defaultFocusTemplatesPath
}

// NewFocusHandler creates a new FocusHandler using the default YAML path.
func NewFocusHandler(pool *pgxpool.Pool) *FocusHandler {
	return &FocusHandler{pool: pool, templatesPath: defaultFocusTemplatesPath}
}

// NewFocusHandlerWithConfig creates a FocusHandler with a custom YAML path.
// Use this in tests to inject a temp-dir path without touching the working dir.
func NewFocusHandlerWithConfig(pool *pgxpool.Pool, templatesPath string) *FocusHandler {
	return &FocusHandler{pool: pool, templatesPath: templatesPath}
}

// RegisterFocusRoutes registers focus mode routes under /api/focus.
func RegisterFocusRoutes(api *gin.RouterGroup, h *FocusHandler, auth gin.HandlerFunc) {
	focus := api.Group("/focus")
	focus.Use(auth, middleware.RequireAuth())
	{
		focus.GET("/templates", h.GetFocusModeTemplates)
		focus.GET("/settings", h.GetEffectiveFocusSettings)
		focus.POST("/preflight", h.BuildPreflightContext)
	}
}

// FocusTemplateResponse represents a focus mode template.
type FocusTemplateResponse struct {
	Name            string  `json:"name"            yaml:"name"`
	DisplayName     string  `json:"display_name"    yaml:"display_name"`
	Description     string  `json:"description"     yaml:"description"`
	Icon            string  `json:"icon"            yaml:"icon"`
	Temperature     float64 `json:"temperature"     yaml:"temperature"`
	MaxTokens       int     `json:"max_tokens"      yaml:"max_tokens"`
	OutputStyle     string  `json:"output_style"    yaml:"output_style"`
	AutoSearch      bool    `json:"auto_search"     yaml:"auto_search"`
	ThinkingEnabled bool    `json:"thinking_enabled" yaml:"thinking_enabled"`
}

// focusTemplatesFile is the top-level YAML document structure.
type focusTemplatesFile struct {
	Templates []FocusTemplateResponse `yaml:"templates"`
}

// hardcodedFocusTemplates is the built-in fallback list returned when the
// YAML file does not exist or cannot be parsed.
var hardcodedFocusTemplates = []FocusTemplateResponse{
	{Name: "quick", DisplayName: "Quick", Description: "Fast, concise responses", Icon: "zap", Temperature: 0.5, MaxTokens: 2048, OutputStyle: "concise", AutoSearch: false, ThinkingEnabled: false},
	{Name: "deep", DisplayName: "Deep Research", Description: "Thorough research with sources", Icon: "search", Temperature: 0.7, MaxTokens: 8192, OutputStyle: "detailed", AutoSearch: true, ThinkingEnabled: true},
	{Name: "creative", DisplayName: "Creative", Description: "Imaginative responses", Icon: "sparkles", Temperature: 0.9, MaxTokens: 4096, OutputStyle: "balanced", AutoSearch: false, ThinkingEnabled: true},
	{Name: "analyze", DisplayName: "Analysis", Description: "Data-driven analysis", Icon: "chart-bar", Temperature: 0.6, MaxTokens: 6144, OutputStyle: "structured", AutoSearch: false, ThinkingEnabled: true},
	{Name: "write", DisplayName: "Writing", Description: "Document creation", Icon: "file-text", Temperature: 0.7, MaxTokens: 8192, OutputStyle: "detailed", AutoSearch: false, ThinkingEnabled: false},
	{Name: "plan", DisplayName: "Planning", Description: "Strategic planning", Icon: "clipboard-list", Temperature: 0.6, MaxTokens: 6144, OutputStyle: "structured", AutoSearch: false, ThinkingEnabled: true},
	{Name: "code", DisplayName: "Coding", Description: "Software development", Icon: "code", Temperature: 0.4, MaxTokens: 8192, OutputStyle: "structured", AutoSearch: false, ThinkingEnabled: true},
}

// loadFocusTemplates attempts to read templates from the YAML file at path.
// Returns the hardcoded list if the file does not exist. Returns an error only
// for malformed YAML (not for a missing file, which is a normal case).
func loadFocusTemplates(path string) ([]FocusTemplateResponse, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return hardcodedFocusTemplates, nil
	}
	if err != nil {
		slog.Default().Warn("focus templates: failed to read YAML file, using defaults",
			"path", path, "error", err)
		return hardcodedFocusTemplates, nil
	}

	var doc focusTemplatesFile
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}

	if len(doc.Templates) == 0 {
		return hardcodedFocusTemplates, nil
	}
	return doc.Templates, nil
}

// GetFocusModeTemplates returns all available focus mode templates.
// Templates are loaded from the YAML file at h.templatesPath on every request
// (file is small; no caching needed for this endpoint's call frequency).
// If the file is absent the hardcoded list is returned as a fallback.
func (h *FocusHandler) GetFocusModeTemplates(c *gin.Context) {
	path := h.templatesPath
	if path == "" {
		path = defaultFocusTemplatesPath
	}

	templates, err := loadFocusTemplates(path)
	if err != nil {
		slog.Default().Error("focus templates: YAML parse error, falling back to defaults",
			"path", path, "error", err)
		templates = hardcodedFocusTemplates
	}

	c.JSON(http.StatusOK, templates)
}

// GetEffectiveFocusSettings returns merged settings for a focus mode
func (h *FocusHandler) GetEffectiveFocusSettings(c *gin.Context) {
	user := c.MustGet("user").(*middleware.BetterAuthUser)
	focusMode := c.Query("mode")

	if focusMode == "" {
		utils.RespondInvalidRequest(c, slog.Default(), nil)
		return
	}

	focusService := services.NewFocusService(h.pool)
	settings, err := focusService.GetEffectiveSettings(c.Request.Context(), user.ID, focusMode)
	if err != nil {
		utils.RespondInternalError(c, slog.Default(), "get settings", err)
		return
	}

	c.JSON(http.StatusOK, settings)
}

// BuildPreflightContext builds the preflight context for a focus mode
func (h *FocusHandler) BuildPreflightContext(c *gin.Context) {
	user := c.MustGet("user").(*middleware.BetterAuthUser)

	var req struct {
		FocusMode   string `json:"focus_mode" binding:"required"`
		UserMessage string `json:"user_message"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	focusService := services.NewFocusService(h.pool)
	focusCtx, err := focusService.BuildPreflightContext(
		c.Request.Context(),
		user.ID,
		req.FocusMode,
		req.UserMessage,
		nil,
		nil,
	)
	if err != nil {
		utils.RespondInternalError(c, slog.Default(), "build context", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"system_prompt": focusCtx.SystemPrompt,
		"llm_options":   focusCtx.LLMOptions,
		"constraints":   focusCtx.OutputConstraints,
	})
}
