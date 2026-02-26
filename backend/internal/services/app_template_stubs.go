package services

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AppTemplate represents an app template record.
// NOTE: This is a stub. The full implementation was in app_template_service.go (deleted).
// Restore that file to re-enable template service functionality.
type AppTemplate struct {
	ID                  uuid.UUID              `json:"id"`
	TemplateName        string                 `json:"template_name"`
	Category            string                 `json:"category"`
	DisplayName         string                 `json:"display_name"`
	Description         string                 `json:"description"`
	IconType            string                 `json:"icon_type"`
	TargetBusinessTypes []string               `json:"target_business_types"`
	TargetChallenges    []string               `json:"target_challenges"`
	TargetTeamSizes     []string               `json:"target_team_sizes"`
	PriorityScore       int                    `json:"priority_score"`
	TemplateConfig      map[string]interface{} `json:"template_config"`
	RequiredModules     []string               `json:"required_modules"`
	OptionalFeatures    []string               `json:"optional_features"`
	GenerationPrompt    string                 `json:"generation_prompt"`
	ScaffoldType        string                 `json:"scaffold_type"`
}

// AppTemplateFilters holds optional filter parameters for listing templates.
// NOTE: Stub — filters are unused while app_template_service.go is absent.
type AppTemplateFilters struct {
	Category     *string
	BusinessType *string
	Challenge    *string
	TeamSize     *string
}

// MatchedTemplate represents a template matched with a score for recommendations.
// NOTE: Stub — unused while app_template_service.go is absent.
type MatchedTemplate struct {
	Template   AppTemplate `json:"template"`
	MatchScore int         `json:"match_score"`
	Reasons    []string    `json:"reasons"`
}

// AppTemplateService is a stub for the deleted app_template_service.go.
// All methods return "template service not available".
// TODO: Restore app_template_service.go to re-enable this functionality.
type AppTemplateService struct {
	pool *pgxpool.Pool
}

// NewAppTemplateService creates a stub AppTemplateService.
// TODO: Restore app_template_service.go to re-enable this functionality.
func NewAppTemplateService(pool *pgxpool.Pool, _ interface{}) *AppTemplateService {
	return &AppTemplateService{pool: pool}
}
