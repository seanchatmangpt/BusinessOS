package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CustomModuleService handles custom module management
type CustomModuleService struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

// CustomModule represents a user-created module
type CustomModule struct {
	ID          uuid.UUID              `json:"id"`
	CreatedBy   uuid.UUID              `json:"created_by"`
	WorkspaceID uuid.UUID              `json:"workspace_id"`
	Name        string                 `json:"name"`
	Slug        string                 `json:"slug"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	Version     string                 `json:"version"`
	Manifest    map[string]interface{} `json:"manifest"`
	Config      map[string]interface{} `json:"config"`
	Icon        string                 `json:"icon"`
	Tags        []string               `json:"tags"`
	Keywords    []string               `json:"keywords"`
	IsPublic    bool                   `json:"is_public"`
	IsPublished bool                   `json:"is_published"`
	IsTemplate  bool                   `json:"is_template"`
	InstallCount int                   `json:"install_count"`
	StarCount    int                   `json:"star_count"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	PublishedAt  *time.Time             `json:"published_at,omitempty"`
}

// ModuleVersion represents a version snapshot
type ModuleVersion struct {
	ID               uuid.UUID              `json:"id"`
	ModuleID         uuid.UUID              `json:"module_id"`
	Version          string                 `json:"version"`
	Changelog        string                 `json:"changelog"`
	ManifestSnapshot map[string]interface{} `json:"manifest_snapshot"`
	ConfigSnapshot   map[string]interface{} `json:"config_snapshot"`
	CreatedBy        uuid.UUID              `json:"created_by"`
	CreatedAt        time.Time              `json:"created_at"`
	IsStable         bool                   `json:"is_stable"`
	IsBreaking       bool                   `json:"is_breaking"`
}

// ModuleInstallation represents an installed module
type ModuleInstallation struct {
	ID              uuid.UUID              `json:"id"`
	ModuleID        uuid.UUID              `json:"module_id"`
	WorkspaceID     uuid.UUID              `json:"workspace_id"`
	InstalledBy     uuid.UUID              `json:"installed_by"`
	InstalledVersion string                 `json:"installed_version"`
	ConfigOverride  map[string]interface{} `json:"config_override"`
	IsEnabled       bool                   `json:"is_enabled"`
	IsAutoUpdate    bool                   `json:"is_auto_update"`
	InstalledAt     time.Time              `json:"installed_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	LastUsedAt      *time.Time             `json:"last_used_at,omitempty"`
}

// ModuleShare represents sharing permissions
type ModuleShare struct {
	ID                    uuid.UUID  `json:"id"`
	ModuleID              uuid.UUID  `json:"module_id"`
	SharedWithUserID      *uuid.UUID `json:"shared_with_user_id,omitempty"`
	SharedWithWorkspaceID *uuid.UUID `json:"shared_with_workspace_id,omitempty"`
	SharedWithEmail       *string    `json:"shared_with_email,omitempty"`
	CanView               bool       `json:"can_view"`
	CanInstall            bool       `json:"can_install"`
	CanModify             bool       `json:"can_modify"`
	CanReshare            bool       `json:"can_reshare"`
	SharedBy              uuid.UUID  `json:"shared_by"`
	SharedAt              time.Time  `json:"shared_at"`
	ExpiresAt             *time.Time `json:"expires_at,omitempty"`
}

// CreateModuleRequest contains data for creating a module
type CreateModuleRequest struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	Manifest    map[string]interface{} `json:"manifest"`
	Config      map[string]interface{} `json:"config,omitempty"`
	Icon        string                 `json:"icon,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Keywords    []string               `json:"keywords,omitempty"`
}

// UpdateModuleRequest contains data for updating a module
type UpdateModuleRequest struct {
	Name        *string                 `json:"name,omitempty"`
	Description *string                 `json:"description,omitempty"`
	Category    *string                 `json:"category,omitempty"`
	Version     *string                 `json:"version,omitempty"`
	Manifest    *map[string]interface{} `json:"manifest,omitempty"`
	Config      *map[string]interface{} `json:"config,omitempty"`
	Icon        *string                 `json:"icon,omitempty"`
	Tags        *[]string               `json:"tags,omitempty"`
	Keywords    *[]string               `json:"keywords,omitempty"`
	IsPublic    *bool                   `json:"is_public,omitempty"`
	IsTemplate  *bool                   `json:"is_template,omitempty"`
}

func NewCustomModuleService(pool *pgxpool.Pool, logger *slog.Logger) *CustomModuleService {
	return &CustomModuleService{
		pool:   pool,
		logger: logger,
	}
}

// CreateModule creates a new custom module
func (s *CustomModuleService) CreateModule(
	ctx context.Context,
	workspaceID uuid.UUID,
	userID uuid.UUID,
	req CreateModuleRequest,
) (*CustomModule, error) {
	// Generate slug from name
	slug := GenerateModuleSlug(req.Name)

	// Validate manifest
	if err := validateManifest(req.Manifest); err != nil {
		return nil, fmt.Errorf("invalid manifest: %w", err)
	}

	// Default config if not provided
	if req.Config == nil {
		req.Config = make(map[string]interface{})
	}

	// Marshal JSONB fields
	manifestJSON, err := json.Marshal(req.Manifest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal manifest: %w", err)
	}

	configJSON, err := json.Marshal(req.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	// Insert module
	query := `
		INSERT INTO custom_modules (
			created_by,
			workspace_id,
			name,
			slug,
			description,
			category,
			version,
			manifest,
			config,
			icon,
			tags,
			keywords
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		) RETURNING
			id, created_by, workspace_id, name, slug, description, category,
			version, manifest, config, icon, tags, keywords, is_public, is_published,
			is_template, install_count, star_count, created_at, updated_at, published_at
	`

	row := s.pool.QueryRow(ctx, query,
		userID,
		workspaceID,
		req.Name,
		slug,
		req.Description,
		req.Category,
		"0.0.1", // Initial version
		manifestJSON,
		configJSON,
		req.Icon,
		req.Tags,
		req.Keywords,
	)

	module := &CustomModule{}
	err = row.Scan(
		&module.ID,
		&module.CreatedBy,
		&module.WorkspaceID,
		&module.Name,
		&module.Slug,
		&module.Description,
		&module.Category,
		&module.Version,
		&module.Manifest,
		&module.Config,
		&module.Icon,
		&module.Tags,
		&module.Keywords,
		&module.IsPublic,
		&module.IsPublished,
		&module.IsTemplate,
		&module.InstallCount,
		&module.StarCount,
		&module.CreatedAt,
		&module.UpdatedAt,
		&module.PublishedAt,
	)

	if err != nil {
		s.logger.Error("Failed to create module", "error", err)
		return nil, fmt.Errorf("failed to create module: %w", err)
	}

	s.logger.Info("Module created", "module_id", module.ID, "name", module.Name)
	return module, nil
}

// GetModule retrieves a module by ID
func (s *CustomModuleService) GetModule(ctx context.Context, moduleID uuid.UUID) (*CustomModule, error) {
	query := `
		SELECT
			id, created_by, workspace_id, name, slug, description, category,
			version, manifest, config, icon, tags, keywords, is_public, is_published,
			is_template, install_count, star_count, created_at, updated_at, published_at
		FROM custom_modules
		WHERE id = $1
	`

	module := &CustomModule{}
	err := s.pool.QueryRow(ctx, query, moduleID).Scan(
		&module.ID,
		&module.CreatedBy,
		&module.WorkspaceID,
		&module.Name,
		&module.Slug,
		&module.Description,
		&module.Category,
		&module.Version,
		&module.Manifest,
		&module.Config,
		&module.Icon,
		&module.Tags,
		&module.Keywords,
		&module.IsPublic,
		&module.IsPublished,
		&module.IsTemplate,
		&module.InstallCount,
		&module.StarCount,
		&module.CreatedAt,
		&module.UpdatedAt,
		&module.PublishedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("module not found")
	}
	if err != nil {
		s.logger.Error("Failed to get module", "error", err, "module_id", moduleID)
		return nil, fmt.Errorf("failed to get module: %w", err)
	}

	return module, nil
}

// ListModules lists modules in a workspace
func (s *CustomModuleService) ListModules(
	ctx context.Context,
	workspaceID uuid.UUID,
	limit int,
	offset int,
) ([]CustomModule, error) {
	if limit <= 0 {
		limit = 20
	}

	query := `
		SELECT
			id, created_by, workspace_id, name, slug, description, category,
			version, manifest, config, icon, tags, keywords, is_public, is_published,
			is_template, install_count, star_count, created_at, updated_at, published_at
		FROM custom_modules
		WHERE workspace_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := s.pool.Query(ctx, query, workspaceID, limit, offset)
	if err != nil {
		s.logger.Error("Failed to list modules", "error", err, "workspace_id", workspaceID)
		return nil, fmt.Errorf("failed to list modules: %w", err)
	}
	defer rows.Close()

	var modules []CustomModule
	for rows.Next() {
		var module CustomModule
		err := rows.Scan(
			&module.ID,
			&module.CreatedBy,
			&module.WorkspaceID,
			&module.Name,
			&module.Slug,
			&module.Description,
			&module.Category,
			&module.Version,
			&module.Manifest,
			&module.Config,
			&module.Icon,
			&module.Tags,
			&module.Keywords,
			&module.IsPublic,
			&module.IsPublished,
			&module.IsTemplate,
			&module.InstallCount,
			&module.StarCount,
			&module.CreatedAt,
			&module.UpdatedAt,
			&module.PublishedAt,
		)
		if err != nil {
			s.logger.Error("Failed to scan module", "error", err)
			continue
		}
		modules = append(modules, module)
	}

	return modules, nil
}

// UpdateModule updates a module
func (s *CustomModuleService) UpdateModule(
	ctx context.Context,
	moduleID uuid.UUID,
	userID uuid.UUID,
	req UpdateModuleRequest,
) (*CustomModule, error) {
	// First check ownership
	var createdBy uuid.UUID
	err := s.pool.QueryRow(ctx, "SELECT created_by FROM custom_modules WHERE id = $1", moduleID).Scan(&createdBy)
	if err != nil {
		return nil, fmt.Errorf("module not found")
	}
	if createdBy != userID {
		return nil, fmt.Errorf("unauthorized: you don't own this module")
	}

	// Build dynamic update query
	updates := []string{}
	args := []interface{}{moduleID}
	argIdx := 2

	if req.Name != nil {
		updates = append(updates, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *req.Name)
		argIdx++

		// Update slug too
		slug := GenerateModuleSlug(*req.Name)
		updates = append(updates, fmt.Sprintf("slug = $%d", argIdx))
		args = append(args, slug)
		argIdx++
	}

	if req.Description != nil {
		updates = append(updates, fmt.Sprintf("description = $%d", argIdx))
		args = append(args, *req.Description)
		argIdx++
	}

	if req.Category != nil {
		updates = append(updates, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, *req.Category)
		argIdx++
	}

	if req.Version != nil {
		updates = append(updates, fmt.Sprintf("version = $%d", argIdx))
		args = append(args, *req.Version)
		argIdx++
	}

	if req.Manifest != nil {
		manifestJSON, err := json.Marshal(*req.Manifest)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal manifest: %w", err)
		}
		updates = append(updates, fmt.Sprintf("manifest = $%d", argIdx))
		args = append(args, manifestJSON)
		argIdx++
	}

	if req.Config != nil {
		configJSON, err := json.Marshal(*req.Config)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal config: %w", err)
		}
		updates = append(updates, fmt.Sprintf("config = $%d", argIdx))
		args = append(args, configJSON)
		argIdx++
	}

	if req.Icon != nil {
		updates = append(updates, fmt.Sprintf("icon = $%d", argIdx))
		args = append(args, *req.Icon)
		argIdx++
	}

	if req.Tags != nil {
		updates = append(updates, fmt.Sprintf("tags = $%d", argIdx))
		args = append(args, *req.Tags)
		argIdx++
	}

	if req.Keywords != nil {
		updates = append(updates, fmt.Sprintf("keywords = $%d", argIdx))
		args = append(args, *req.Keywords)
		argIdx++
	}

	if req.IsPublic != nil {
		updates = append(updates, fmt.Sprintf("is_public = $%d", argIdx))
		args = append(args, *req.IsPublic)
		argIdx++
	}

	if req.IsTemplate != nil {
		updates = append(updates, fmt.Sprintf("is_template = $%d", argIdx))
		args = append(args, *req.IsTemplate)
		argIdx++
	}

	if len(updates) == 0 {
		return s.GetModule(ctx, moduleID)
	}

	updates = append(updates, "updated_at = NOW()")

	query := fmt.Sprintf(`
		UPDATE custom_modules SET %s
		WHERE id = $1
		RETURNING
			id, created_by, workspace_id, name, slug, description, category,
			version, manifest, config, icon, tags, keywords, is_public, is_published,
			is_template, install_count, star_count, created_at, updated_at, published_at
	`, strings.Join(updates, ", "))

	module := &CustomModule{}
	err = s.pool.QueryRow(ctx, query, args...).Scan(
		&module.ID,
		&module.CreatedBy,
		&module.WorkspaceID,
		&module.Name,
		&module.Slug,
		&module.Description,
		&module.Category,
		&module.Version,
		&module.Manifest,
		&module.Config,
		&module.Icon,
		&module.Tags,
		&module.Keywords,
		&module.IsPublic,
		&module.IsPublished,
		&module.IsTemplate,
		&module.InstallCount,
		&module.StarCount,
		&module.CreatedAt,
		&module.UpdatedAt,
		&module.PublishedAt,
	)

	if err != nil {
		s.logger.Error("Failed to update module", "error", err, "module_id", moduleID)
		return nil, fmt.Errorf("failed to update module: %w", err)
	}

	s.logger.Info("Module updated", "module_id", module.ID)
	return module, nil
}

// DeleteModule deletes a module
func (s *CustomModuleService) DeleteModule(ctx context.Context, moduleID uuid.UUID, userID uuid.UUID) error {
	query := "DELETE FROM custom_modules WHERE id = $1 AND created_by = $2"
	result, err := s.pool.Exec(ctx, query, moduleID, userID)
	if err != nil {
		s.logger.Error("Failed to delete module", "error", err, "module_id", moduleID)
		return fmt.Errorf("failed to delete module: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("module not found or unauthorized")
	}

	s.logger.Info("Module deleted", "module_id", moduleID)
	return nil
}

// PublishModule publishes a module to the registry
func (s *CustomModuleService) PublishModule(ctx context.Context, moduleID uuid.UUID, userID uuid.UUID) error {
	query := `
		UPDATE custom_modules
		SET is_published = TRUE, is_public = TRUE, published_at = NOW()
		WHERE id = $1 AND created_by = $2
	`

	result, err := s.pool.Exec(ctx, query, moduleID, userID)
	if err != nil {
		s.logger.Error("Failed to publish module", "error", err, "module_id", moduleID)
		return fmt.Errorf("failed to publish module: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("module not found or unauthorized")
	}

	s.logger.Info("Module published", "module_id", moduleID)
	return nil
}

// SearchModules searches public modules
func (s *CustomModuleService) SearchModules(
	ctx context.Context,
	query string,
	limit int,
	offset int,
) ([]CustomModule, error) {
	if limit <= 0 {
		limit = 20
	}

	sqlQuery := `
		SELECT
			id, created_by, workspace_id, name, slug, description, category,
			version, manifest, config, icon, tags, keywords, is_public, is_published,
			is_template, install_count, star_count, created_at, updated_at, published_at
		FROM custom_modules
		WHERE is_public = TRUE AND is_published = TRUE
			AND (
				name ILIKE '%' || $1 || '%'
				OR description ILIKE '%' || $1 || '%'
				OR $1 = ANY(tags)
				OR $1 = ANY(keywords)
			)
		ORDER BY install_count DESC, created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := s.pool.Query(ctx, sqlQuery, query, limit, offset)
	if err != nil {
		s.logger.Error("Failed to search modules", "error", err, "query", query)
		return nil, fmt.Errorf("failed to search modules: %w", err)
	}
	defer rows.Close()

	var modules []CustomModule
	for rows.Next() {
		var module CustomModule
		err := rows.Scan(
			&module.ID,
			&module.CreatedBy,
			&module.WorkspaceID,
			&module.Name,
			&module.Slug,
			&module.Description,
			&module.Category,
			&module.Version,
			&module.Manifest,
			&module.Config,
			&module.Icon,
			&module.Tags,
			&module.Keywords,
			&module.IsPublic,
			&module.IsPublished,
			&module.IsTemplate,
			&module.InstallCount,
			&module.StarCount,
			&module.CreatedAt,
			&module.UpdatedAt,
			&module.PublishedAt,
		)
		if err != nil {
			s.logger.Error("Failed to scan module", "error", err)
			continue
		}
		modules = append(modules, module)
	}

	return modules, nil
}

// Helper functions

// GenerateModuleSlug converts a module name into a URL-friendly slug
func GenerateModuleSlug(name string) string {
	// Convert to lowercase
	slug := strings.ToLower(name)
	// Replace spaces with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")
	// Remove non-alphanumeric characters (except hyphens)
	reg := regexp.MustCompile("[^a-z0-9-]+")
	slug = reg.ReplaceAllString(slug, "")
	// Remove consecutive hyphens
	reg = regexp.MustCompile("-+")
	slug = reg.ReplaceAllString(slug, "-")
	// Trim hyphens from start and end
	slug = strings.Trim(slug, "-")
	return slug
}

// CreateInstallation creates a module installation
func (s *CustomModuleService) CreateInstallation(
	ctx context.Context,
	moduleID uuid.UUID,
	workspaceID uuid.UUID,
	userID uuid.UUID,
	version string,
) (*ModuleInstallation, error) {
	query := `
		INSERT INTO custom_module_installations (
			module_id, workspace_id, installed_by, installed_version
		) VALUES ($1, $2, $3, $4)
		ON CONFLICT (module_id, workspace_id) DO UPDATE
		SET installed_version = EXCLUDED.installed_version,
		    updated_at = NOW()
		RETURNING
			id, module_id, workspace_id, installed_by, installed_version,
			config_override, is_enabled, is_auto_update,
			installed_at, updated_at, last_used_at
	`

	installation := &ModuleInstallation{}
	err := s.pool.QueryRow(ctx, query, moduleID, workspaceID, userID, version).Scan(
		&installation.ID,
		&installation.ModuleID,
		&installation.WorkspaceID,
		&installation.InstalledBy,
		&installation.InstalledVersion,
		&installation.ConfigOverride,
		&installation.IsEnabled,
		&installation.IsAutoUpdate,
		&installation.InstalledAt,
		&installation.UpdatedAt,
		&installation.LastUsedAt,
	)

	if err != nil {
		s.logger.Error("Failed to create installation", "error", err)
		return nil, fmt.Errorf("failed to create installation: %w", err)
	}

	// Increment install count
	_, _ = s.pool.Exec(ctx, "UPDATE custom_modules SET install_count = install_count + 1 WHERE id = $1", moduleID)

	s.logger.Info("Module installed", "module_id", moduleID, "workspace_id", workspaceID)
	return installation, nil
}

// ListInstallations lists all module installations for a workspace
func (s *CustomModuleService) ListInstallations(
	ctx context.Context,
	workspaceID uuid.UUID,
) ([]ModuleInstallation, error) {
	query := `
		SELECT
			id, module_id, workspace_id, installed_by, installed_version,
			config_override, is_enabled, is_auto_update,
			installed_at, updated_at, last_used_at
		FROM custom_module_installations
		WHERE workspace_id = $1
		ORDER BY installed_at DESC
	`

	rows, err := s.pool.Query(ctx, query, workspaceID)
	if err != nil {
		s.logger.Error("Failed to list installations", "error", err)
		return nil, fmt.Errorf("failed to list installations: %w", err)
	}
	defer rows.Close()

	var installations []ModuleInstallation
	for rows.Next() {
		var inst ModuleInstallation
		err := rows.Scan(
			&inst.ID,
			&inst.ModuleID,
			&inst.WorkspaceID,
			&inst.InstalledBy,
			&inst.InstalledVersion,
			&inst.ConfigOverride,
			&inst.IsEnabled,
			&inst.IsAutoUpdate,
			&inst.InstalledAt,
			&inst.UpdatedAt,
			&inst.LastUsedAt,
		)
		if err != nil {
			s.logger.Error("Failed to scan installation", "error", err)
			continue
		}
		installations = append(installations, inst)
	}

	return installations, nil
}

// CreateShareRequest represents a request to share a module
type CreateShareRequest struct {
	ModuleID              uuid.UUID
	SharedWithUserID      *uuid.UUID
	SharedWithWorkspaceID *uuid.UUID
	SharedWithEmail       *string
	CanView               bool
	CanInstall            bool
	CanModify             bool
	CanReshare            bool
	SharedBy              uuid.UUID
}

// CreateShare creates a module share
func (s *CustomModuleService) CreateShare(
	ctx context.Context,
	req CreateShareRequest,
) (*ModuleShare, error) {
	query := `
		INSERT INTO custom_module_shares (
			module_id, shared_with_user_id, shared_with_workspace_id, shared_with_email,
			can_view, can_install, can_modify, can_reshare, shared_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING
			id, module_id, shared_with_user_id, shared_with_workspace_id, shared_with_email,
			can_view, can_install, can_modify, can_reshare, shared_by, shared_at, expires_at
	`

	share := &ModuleShare{}
	err := s.pool.QueryRow(ctx, query,
		req.ModuleID,
		req.SharedWithUserID,
		req.SharedWithWorkspaceID,
		req.SharedWithEmail,
		req.CanView,
		req.CanInstall,
		req.CanModify,
		req.CanReshare,
		req.SharedBy,
	).Scan(
		&share.ID,
		&share.ModuleID,
		&share.SharedWithUserID,
		&share.SharedWithWorkspaceID,
		&share.SharedWithEmail,
		&share.CanView,
		&share.CanInstall,
		&share.CanModify,
		&share.CanReshare,
		&share.SharedBy,
		&share.SharedAt,
		&share.ExpiresAt,
	)

	if err != nil {
		s.logger.Error("Failed to create share", "error", err)
		return nil, fmt.Errorf("failed to create share: %w", err)
	}

	s.logger.Info("Module shared", "module_id", req.ModuleID)
	return share, nil
}

func validateManifest(manifest map[string]interface{}) error {
	// Basic validation - check if actions exist
	if _, ok := manifest["actions"]; !ok {
		return fmt.Errorf("manifest must contain 'actions' field")
	}

	actions, ok := manifest["actions"].([]interface{})
	if !ok {
		return fmt.Errorf("manifest 'actions' must be an array")
	}

	if len(actions) == 0 {
		return fmt.Errorf("manifest must contain at least one action")
	}

	// Validate each action
	for i, action := range actions {
		actionMap, ok := action.(map[string]interface{})
		if !ok {
			return fmt.Errorf("action %d must be an object", i)
		}

		if _, ok := actionMap["name"]; !ok {
			return fmt.Errorf("action %d must have a 'name' field", i)
		}

		if _, ok := actionMap["type"]; !ok {
			return fmt.Errorf("action %d must have a 'type' field", i)
		}
	}

	return nil
}
