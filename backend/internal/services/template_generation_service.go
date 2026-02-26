package services

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TemplateGenerationService handles generating apps from templates
type TemplateGenerationService struct {
	pool            *pgxpool.Pool
	logger          *slog.Logger
	userAppsService *UserAppsService
}

// GenerateFromTemplateRequest represents a request to generate an app from a template
type GenerateFromTemplateRequest struct {
	TemplateID  uuid.UUID              `json:"template_id"`
	WorkspaceID uuid.UUID              `json:"workspace_id"`
	AppName     string                 `json:"app_name"`
	Config      map[string]interface{} `json:"config"`
}

// GeneratedFile represents a single generated file
type GeneratedFile struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Size    int    `json:"size"`
}

// GenerationResult represents the result of generating an app from a template
type GenerationResult struct {
	AppID         uuid.UUID       `json:"app_id"`
	AppName       string          `json:"app_name"`
	TemplateID    uuid.UUID       `json:"template_id"`
	TemplateName  string          `json:"template_name"`
	WorkspaceID   uuid.UUID       `json:"workspace_id"`
	Files         []GeneratedFile `json:"files"`
	TotalFiles    int             `json:"total_files"`
	Status        string          `json:"status"`
	VersionNumber string          `json:"version_number"`
	GeneratedAt   time.Time       `json:"generated_at"`
}

// NewTemplateGenerationService creates a new template generation service
func NewTemplateGenerationService(pool *pgxpool.Pool, logger *slog.Logger) *TemplateGenerationService {
	return &TemplateGenerationService{
		pool:            pool,
		logger:          logger,
		userAppsService: NewUserAppsService(pool, logger),
	}
}

// GenerateFromTemplate generates an app from a template with user configuration.
// TODO: template service not available — restore app_template_service.go to enable full generation.
func (s *TemplateGenerationService) GenerateFromTemplate(ctx context.Context, userID uuid.UUID, req GenerateFromTemplateRequest) (*GenerationResult, error) {
	return nil, fmt.Errorf("template service not available")
}

// generateFiles generates files from a built-in template with config substitution
func generateFiles(template *BuiltInTemplate, config map[string]interface{}) []GeneratedFile {
	var files []GeneratedFile
	for path, content := range template.FilesTemplate {
		rendered := substituteConfig(content, config)
		files = append(files, GeneratedFile{
			Path:    path,
			Content: rendered,
			Size:    len(rendered),
		})
	}
	return files
}

// substituteConfig replaces {{key}} placeholders with config values
func substituteConfig(content string, config map[string]interface{}) string {
	result := content
	for key, value := range config {
		placeholder := fmt.Sprintf("{{%s}}", key)
		strValue := fmt.Sprintf("%v", value)
		result = strings.ReplaceAll(result, placeholder, strValue)
	}
	return result
}

// mergeConfig merges user config with template defaults (user config takes precedence)
func mergeConfig(userConfig, templateConfig map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})
	for k, v := range templateConfig {
		merged[k] = v
	}
	for k, v := range userConfig {
		merged[k] = v
	}
	return merged
}
