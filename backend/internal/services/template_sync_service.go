package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gopkg.in/yaml.v3"
)

// TemplateSyncService handles synchronization of YAML templates to database
type TemplateSyncService struct {
	pool         *pgxpool.Pool
	logger       *slog.Logger
	templatesDir string
}

// DBTemplate represents an app template in the database
type DBTemplate struct {
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
	YAMLTemplateName    string                 `json:"yaml_template_name"`
	YAMLVersion         string                 `json:"yaml_version"`
	TemplateVariables   map[string]interface{} `json:"template_variables"`
}

// SyncResult contains the results of a sync operation
type SyncResult struct {
	Inserted int      `json:"inserted"`
	Updated  int      `json:"updated"`
	Skipped  int      `json:"skipped"`
	Errors   []string `json:"errors"`
}

// NewTemplateSyncService creates a new template sync service
func NewTemplateSyncService(pool *pgxpool.Pool, logger *slog.Logger, templatesDir string) *TemplateSyncService {
	return &TemplateSyncService{
		pool:         pool,
		logger:       logger,
		templatesDir: templatesDir,
	}
}

// SyncTemplates syncs all YAML templates to the database
// Strategy: YAML files are the source of truth
func (s *TemplateSyncService) SyncTemplates(ctx context.Context) (*SyncResult, error) {
	s.logger.Info("starting template sync", "templates_dir", s.templatesDir)

	result := &SyncResult{
		Errors: []string{},
	}

	// Find all YAML files in the templates directory
	yamlFiles, err := s.findYAMLFiles(s.templatesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to find YAML files: %w", err)
	}

	s.logger.Info("found YAML templates", "count", len(yamlFiles))

	// Process each YAML file
	for _, yamlFile := range yamlFiles {
		if err := s.syncTemplate(ctx, yamlFile, result); err != nil {
			errMsg := fmt.Sprintf("%s: %v", yamlFile, err)
			result.Errors = append(result.Errors, errMsg)
			s.logger.Error("failed to sync template", "file", yamlFile, "error", err)
		}
	}

	s.logger.Info("template sync completed",
		"inserted", result.Inserted,
		"updated", result.Updated,
		"skipped", result.Skipped,
		"errors", len(result.Errors),
	)

	return result, nil
}

// syncTemplate syncs a single YAML template to the database
func (s *TemplateSyncService) syncTemplate(ctx context.Context, yamlPath string, result *SyncResult) error {
	// Load YAML template
	tmpl, err := s.loadYAMLTemplate(yamlPath)
	if err != nil {
		return fmt.Errorf("load YAML: %w", err)
	}

	s.logger.Debug("loaded YAML template", "name", tmpl.Name, "category", tmpl.Category)

	// Map YAML to DB structure
	dbTemplate := s.MapYAMLToDB(tmpl)

	// Check if template exists in DB
	exists, err := s.templateExists(ctx, tmpl.Name)
	if err != nil {
		return fmt.Errorf("check existence: %w", err)
	}

	if exists {
		// Update existing template
		if err := s.updateTemplate(ctx, dbTemplate); err != nil {
			return fmt.Errorf("update: %w", err)
		}
		result.Updated++
		s.logger.Info("updated template", "name", tmpl.Name)
	} else {
		// Insert new template
		if err := s.insertTemplate(ctx, dbTemplate); err != nil {
			return fmt.Errorf("insert: %w", err)
		}
		result.Inserted++
		s.logger.Info("inserted template", "name", tmpl.Name)
	}

	return nil
}

// findYAMLFiles finds all YAML files in a directory recursively
func (s *TemplateSyncService) findYAMLFiles(dir string) ([]string, error) {
	var yamlFiles []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".yaml" || ext == ".yml" {
				yamlFiles = append(yamlFiles, path)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return yamlFiles, nil
}

// loadYAMLTemplate loads a template from a YAML file
func (s *TemplateSyncService) loadYAMLTemplate(yamlPath string) (*TemplateDefinition, error) {
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var tmpl TemplateDefinition
	if err := yaml.Unmarshal(data, &tmpl); err != nil {
		return nil, fmt.Errorf("unmarshal YAML: %w", err)
	}

	return &tmpl, nil
}

// MapYAMLToDB maps a YAML template to a database template structure
func (s *TemplateSyncService) MapYAMLToDB(yaml *TemplateDefinition) *DBTemplate {
	dbTemplate := &DBTemplate{
		TemplateName:     yaml.Name,
		DisplayName:      yaml.DisplayName,
		Description:      yaml.Description,
		Category:         yaml.Category,
		YAMLTemplateName: yaml.Name,
		YAMLVersion:      yaml.Version,
		GenerationPrompt: yaml.Template,
	}

	// Map category to icon type
	dbTemplate.IconType = s.categoryToIcon(yaml.Category)

	// Extract target business types, challenges, and team sizes from tags
	dbTemplate.TargetBusinessTypes = s.extractBusinessTypes(yaml.Tags, yaml.Category)
	dbTemplate.TargetChallenges = s.extractChallenges(yaml.Tags, yaml.Category)
	dbTemplate.TargetTeamSizes = s.extractTeamSizes(yaml.Tags)

	// Calculate priority score based on category and tags
	dbTemplate.PriorityScore = s.calculatePriorityScore(yaml.Category, yaml.Tags)

	// Determine scaffold type
	dbTemplate.ScaffoldType = s.determineScaffoldType(yaml.Tags, yaml.Category)

	// Extract required modules and optional features from variables and tags
	dbTemplate.RequiredModules = s.extractRequiredModules(yaml.Category)
	dbTemplate.OptionalFeatures = s.extractOptionalFeatures(yaml.Variables)

	// Build template config
	dbTemplate.TemplateConfig = s.buildTemplateConfig(yaml)

	// Store template variables as JSONB
	dbTemplate.TemplateVariables = s.mapVariables(yaml.Variables)

	return dbTemplate
}

// categoryToIcon maps template categories to icon types
func (s *TemplateSyncService) categoryToIcon(category string) string {
	iconMap := map[string]string{
		"app-generation":     "users",
		"data-visualization": "chart",
		"maintenance":        "wrench",
		"feature":            "plus",
		"operations":         "server",
		"marketing":          "globe",
		"crm":                "users",
		"project_management": "kanban",
	}

	if icon, ok := iconMap[category]; ok {
		return icon
	}

	return "file" // default icon
}

// extractBusinessTypes extracts target business types from tags and category
func (s *TemplateSyncService) extractBusinessTypes(tags []string, category string) []string {
	businessTypes := []string{}

	// Category-based defaults
	if category == "app-generation" || category == "crm" {
		businessTypes = append(businessTypes, "saas", "startup", "enterprise", "small_business")
	} else if category == "data-visualization" {
		businessTypes = append(businessTypes, "saas", "enterprise", "agency")
	} else if category == "maintenance" {
		businessTypes = append(businessTypes, "saas", "startup", "enterprise", "agency", "small_business")
	}

	// Tag-based additions
	for _, tag := range tags {
		switch tag {
		case "crm", "business":
			if !stringSliceContains(businessTypes, "small_business") {
				businessTypes = append(businessTypes, "small_business")
			}
		case "full-stack":
			if !stringSliceContains(businessTypes, "startup") {
				businessTypes = append(businessTypes, "startup")
			}
		}
	}

	if len(businessTypes) == 0 {
		// Default fallback
		businessTypes = []string{"saas", "startup"}
	}

	return businessTypes
}

// extractChallenges extracts target challenges from tags and category
func (s *TemplateSyncService) extractChallenges(tags []string, category string) []string {
	challenges := []string{}

	// Category-based challenges
	switch category {
	case "app-generation":
		challenges = []string{"rapid_prototyping", "scalability", "time_to_market"}
	case "data-visualization":
		challenges = []string{"analytics", "reporting", "data_insights"}
	case "maintenance":
		challenges = []string{"bug_fixing", "code_quality", "stability"}
	case "feature":
		challenges = []string{"feature_development", "user_experience", "innovation"}
	}

	// Tag-based additions
	for _, tag := range tags {
		switch tag {
		case "crm":
			if !stringSliceContains(challenges, "client_relationships") {
				challenges = append(challenges, "client_relationships")
			}
		case "analytics", "dashboard", "charts":
			if !stringSliceContains(challenges, "analytics") {
				challenges = append(challenges, "analytics")
			}
		case "bug", "debugging":
			if !stringSliceContains(challenges, "bug_fixing") {
				challenges = append(challenges, "bug_fixing")
			}
		}
	}

	if len(challenges) == 0 {
		challenges = []string{"development"}
	}

	return challenges
}

// extractTeamSizes extracts target team sizes
func (s *TemplateSyncService) extractTeamSizes(tags []string) []string {
	// Default: suitable for all team sizes
	return []string{"solo", "small", "medium", "large"}
}

// calculatePriorityScore calculates priority score based on category and tags
func (s *TemplateSyncService) calculatePriorityScore(category string, tags []string) int {
	baseScore := 70

	// Category-based scoring
	switch category {
	case "app-generation":
		baseScore = 85
	case "data-visualization":
		baseScore = 90
	case "maintenance":
		baseScore = 75
	case "feature":
		baseScore = 80
	}

	// Tag-based adjustments
	for _, tag := range tags {
		switch tag {
		case "full-stack":
			baseScore += 5
		case "crm", "dashboard":
			baseScore += 5
		case "bug":
			baseScore -= 5 // Maintenance has lower priority
		}
	}

	// Ensure score is within valid range
	if baseScore > 100 {
		baseScore = 100
	}
	if baseScore < 1 {
		baseScore = 1
	}

	return baseScore
}

// determineScaffoldType determines the scaffold type based on tags and category
func (s *TemplateSyncService) determineScaffoldType(tags []string, category string) string {
	// Check tags for explicit scaffold types
	for _, tag := range tags {
		if tag == "full-stack" {
			return "full-stack"
		}
	}

	// Category-based defaults
	switch category {
	case "app-generation":
		return "full-stack"
	case "data-visualization":
		return "svelte"
	case "maintenance":
		return "go"
	default:
		return "svelte"
	}
}

// extractRequiredModules extracts required modules based on category
func (s *TemplateSyncService) extractRequiredModules(category string) []string {
	moduleMap := map[string][]string{
		"app-generation":     {"database", "api", "auth"},
		"data-visualization": {"dashboard", "analytics"},
		"maintenance":        {"logging", "testing"},
		"feature":            {"api"},
	}

	if modules, ok := moduleMap[category]; ok {
		return modules
	}

	return []string{}
}

// extractOptionalFeatures extracts optional features from template variables
func (s *TemplateSyncService) extractOptionalFeatures(variables []TemplateVariable) []string {
	features := []string{}

	for _, v := range variables {
		if !v.Required && v.Default != nil {
			// Optional variables suggest optional features
			switch v.Name {
			case "AvailableIntegrations":
				features = append(features, "third_party_integrations")
			case "RefreshInterval":
				features = append(features, "real_time_updates")
			case "UserRoles":
				features = append(features, "role_based_access")
			case "ChartTypes":
				features = append(features, "custom_charts")
			}
		}
	}

	return features
}

// buildTemplateConfig builds the template configuration map
func (s *TemplateSyncService) buildTemplateConfig(yaml *TemplateDefinition) map[string]interface{} {
	config := map[string]interface{}{
		"category": yaml.Category,
		"version":  yaml.Version,
		"tags":     yaml.Tags,
	}

	// Add scaffold type hint
	if strings.Contains(yaml.Template, "Go + Gin") || strings.Contains(yaml.Template, "Backend API") {
		config["has_backend"] = true
	}
	if strings.Contains(yaml.Template, "SvelteKit") || strings.Contains(yaml.Template, "Frontend UI") {
		config["has_frontend"] = true
	}

	return config
}

// mapVariables maps YAML variables to JSONB structure
func (s *TemplateSyncService) mapVariables(variables []TemplateVariable) map[string]interface{} {
	varMap := make(map[string]interface{})

	for _, v := range variables {
		varMap[v.Name] = map[string]interface{}{
			"type":        v.Type,
			"required":    v.Required,
			"default":     v.Default,
			"description": v.Description,
		}
	}

	return varMap
}

// templateExists checks if a template with the given name exists in the database
func (s *TemplateSyncService) templateExists(ctx context.Context, name string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM app_templates WHERE template_name = $1
		)
	`, name).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("check existence query: %w", err)
	}

	return exists, nil
}

// insertTemplate inserts a new template into the database
func (s *TemplateSyncService) insertTemplate(ctx context.Context, tmpl *DBTemplate) error {
	configJSON, err := json.Marshal(tmpl.TemplateConfig)
	if err != nil {
		return fmt.Errorf("marshal template_config: %w", err)
	}

	variablesJSON, err := json.Marshal(tmpl.TemplateVariables)
	if err != nil {
		return fmt.Errorf("marshal template_variables: %w", err)
	}

	_, err = s.pool.Exec(ctx, `
		INSERT INTO app_templates (
			template_name, category, display_name, description, icon_type,
			target_business_types, target_challenges, target_team_sizes,
			priority_score, template_config, required_modules, optional_features,
			generation_prompt, scaffold_type,
			yaml_template_name, yaml_version, template_variables
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
		)
	`,
		tmpl.TemplateName,
		tmpl.Category,
		tmpl.DisplayName,
		tmpl.Description,
		tmpl.IconType,
		tmpl.TargetBusinessTypes,
		tmpl.TargetChallenges,
		tmpl.TargetTeamSizes,
		tmpl.PriorityScore,
		configJSON,
		tmpl.RequiredModules,
		tmpl.OptionalFeatures,
		tmpl.GenerationPrompt,
		tmpl.ScaffoldType,
		tmpl.YAMLTemplateName,
		tmpl.YAMLVersion,
		variablesJSON,
	)

	if err != nil {
		return fmt.Errorf("insert query: %w", err)
	}

	return nil
}

// updateTemplate updates an existing template in the database
func (s *TemplateSyncService) updateTemplate(ctx context.Context, tmpl *DBTemplate) error {
	configJSON, err := json.Marshal(tmpl.TemplateConfig)
	if err != nil {
		return fmt.Errorf("marshal template_config: %w", err)
	}

	variablesJSON, err := json.Marshal(tmpl.TemplateVariables)
	if err != nil {
		return fmt.Errorf("marshal template_variables: %w", err)
	}

	_, err = s.pool.Exec(ctx, `
		UPDATE app_templates SET
			category = $2,
			display_name = $3,
			description = $4,
			icon_type = $5,
			target_business_types = $6,
			target_challenges = $7,
			target_team_sizes = $8,
			priority_score = $9,
			template_config = $10,
			required_modules = $11,
			optional_features = $12,
			generation_prompt = $13,
			scaffold_type = $14,
			yaml_template_name = $15,
			yaml_version = $16,
			template_variables = $17,
			updated_at = NOW()
		WHERE template_name = $1
	`,
		tmpl.TemplateName,
		tmpl.Category,
		tmpl.DisplayName,
		tmpl.Description,
		tmpl.IconType,
		tmpl.TargetBusinessTypes,
		tmpl.TargetChallenges,
		tmpl.TargetTeamSizes,
		tmpl.PriorityScore,
		configJSON,
		tmpl.RequiredModules,
		tmpl.OptionalFeatures,
		tmpl.GenerationPrompt,
		tmpl.ScaffoldType,
		tmpl.YAMLTemplateName,
		tmpl.YAMLVersion,
		variablesJSON,
	)

	if err != nil {
		return fmt.Errorf("update query: %w", err)
	}

	return nil
}

// GetTemplateByName retrieves a template from the database by name
func (s *TemplateSyncService) GetTemplateByName(ctx context.Context, name string) (*DBTemplate, error) {
	var tmpl DBTemplate
	var configJSON, variablesJSON []byte

	err := s.pool.QueryRow(ctx, `
		SELECT
			id, template_name, category, display_name, description, icon_type,
			target_business_types, target_challenges, target_team_sizes,
			priority_score, COALESCE(template_config, '{}'::jsonb),
			required_modules, optional_features, generation_prompt,
			COALESCE(scaffold_type, ''), COALESCE(yaml_template_name, ''),
			COALESCE(yaml_version, ''), COALESCE(template_variables, '{}'::jsonb)
		FROM app_templates
		WHERE template_name = $1
	`, name).Scan(
		&tmpl.ID,
		&tmpl.TemplateName,
		&tmpl.Category,
		&tmpl.DisplayName,
		&tmpl.Description,
		&tmpl.IconType,
		&tmpl.TargetBusinessTypes,
		&tmpl.TargetChallenges,
		&tmpl.TargetTeamSizes,
		&tmpl.PriorityScore,
		&configJSON,
		&tmpl.RequiredModules,
		&tmpl.OptionalFeatures,
		&tmpl.GenerationPrompt,
		&tmpl.ScaffoldType,
		&tmpl.YAMLTemplateName,
		&tmpl.YAMLVersion,
		&variablesJSON,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("template not found: %s", name)
		}
		return nil, fmt.Errorf("query template: %w", err)
	}

	// Unmarshal JSON fields
	if len(configJSON) > 0 {
		if err := json.Unmarshal(configJSON, &tmpl.TemplateConfig); err != nil {
			s.logger.Warn("failed to unmarshal template_config", "name", name, "error", err)
		}
	}

	if len(variablesJSON) > 0 {
		if err := json.Unmarshal(variablesJSON, &tmpl.TemplateVariables); err != nil {
			s.logger.Warn("failed to unmarshal template_variables", "name", name, "error", err)
		}
	}

	return &tmpl, nil
}

// Helper function to check if a string slice contains a value
func stringSliceContains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}
