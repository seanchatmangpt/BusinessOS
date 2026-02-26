package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WorkspaceVersionService handles version snapshots for workspaces
type WorkspaceVersionService struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

// NewWorkspaceVersionService creates a new version service
func NewWorkspaceVersionService(pool *pgxpool.Pool, logger *slog.Logger) *WorkspaceVersionService {
	return &WorkspaceVersionService{
		pool:   pool,
		logger: logger,
	}
}

// CreateSnapshot captures the current workspace state as a new version
func (s *WorkspaceVersionService) CreateSnapshot(
	ctx context.Context,
	workspaceID uuid.UUID,
	userID string,
) (string, error) {
	// Get next version number
	var lastVersion *string
	err := s.pool.QueryRow(ctx, `
		SELECT version_number
		FROM workspace_versions
		WHERE workspace_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`, workspaceID).Scan(&lastVersion)

	if err != nil && err != pgx.ErrNoRows {
		return "", fmt.Errorf("fetch last version: %w", err)
	}

	nextVersion := incrementVersion(lastVersion)

	// Capture snapshot data
	snapshotData := s.captureWorkspaceState(ctx, workspaceID)

	// Parse to extract metadata
	var snapshot WorkspaceSnapshot
	json.Unmarshal(snapshotData, &snapshot)

	// Create metadata JSON
	metadataJSON, _ := json.Marshal(snapshot.Metadata)

	// Save snapshot
	_, err = s.pool.Exec(ctx, `
		INSERT INTO workspace_versions (workspace_id, version_number, snapshot_data, snapshot_metadata, created_by)
		VALUES ($1, $2, $3, $4, $5)
	`, workspaceID, nextVersion, snapshotData, metadataJSON, userID)

	if err != nil {
		return "", fmt.Errorf("save snapshot: %w", err)
	}

	s.logger.Info("workspace snapshot created",
		"workspace_id", workspaceID,
		"version", nextVersion)
	return nextVersion, nil
}

// RestoreSnapshot restores a workspace to a specific version
func (s *WorkspaceVersionService) RestoreSnapshot(
	ctx context.Context,
	workspaceID uuid.UUID,
	versionNumber string,
	userID string,
) error {
	// Fetch snapshot
	var snapshotData json.RawMessage
	err := s.pool.QueryRow(ctx, `
		SELECT snapshot_data
		FROM workspace_versions
		WHERE workspace_id = $1 AND version_number = $2
	`, workspaceID, versionNumber).Scan(&snapshotData)

	if err != nil {
		return fmt.Errorf("snapshot not found: %w", err)
	}

	// Parse snapshot
	var snapshot WorkspaceSnapshot
	if err := json.Unmarshal(snapshotData, &snapshot); err != nil {
		return fmt.Errorf("invalid snapshot data: %w", err)
	}

	// Create backup of current state BEFORE restoring
	backupVersion, err := s.CreateSnapshot(ctx, workspaceID, userID)
	if err != nil {
		s.logger.Error("failed to create backup before restore", "error", err)
		// Continue anyway - restore is more important than backup failure
	} else {
		s.logger.Info("backup created before restore", "backup_version", backupVersion)
	}

	// Start transaction for atomic restore
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Restore workspace settings
	if err := s.restoreSettings(ctx, tx, workspaceID, snapshot.Settings); err != nil {
		return fmt.Errorf("restore settings: %w", err)
	}

	// Restore roles (must be before members as members reference roles)
	if err := s.restoreRoles(ctx, tx, workspaceID, snapshot.Roles); err != nil {
		return fmt.Errorf("restore roles: %w", err)
	}

	// Restore members
	if err := s.restoreMembers(ctx, tx, workspaceID, snapshot.Members); err != nil {
		return fmt.Errorf("restore members: %w", err)
	}

	// Restore apps
	if err := s.restoreApps(ctx, tx, workspaceID, snapshot.Apps); err != nil {
		return fmt.Errorf("restore apps: %w", err)
	}

	// Restore memories
	if err := s.restoreMemories(ctx, tx, workspaceID, snapshot.Memories); err != nil {
		return fmt.Errorf("restore memories: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	s.logger.Info("workspace restored successfully",
		"workspace_id", workspaceID,
		"version", versionNumber,
		"apps_restored", len(snapshot.Apps),
		"members_restored", len(snapshot.Members),
		"backup_version", backupVersion)

	return nil
}

// PreviewRestore previews changes that would be made by restoring a version (dry run)
func (s *WorkspaceVersionService) PreviewRestore(
	ctx context.Context,
	workspaceID uuid.UUID,
	versionNumber string,
) (map[string]interface{}, error) {
	// Fetch snapshot
	var snapshotData json.RawMessage
	var createdAt time.Time
	err := s.pool.QueryRow(ctx, `
		SELECT snapshot_data, created_at
		FROM workspace_versions
		WHERE workspace_id = $1 AND version_number = $2
	`, workspaceID, versionNumber).Scan(&snapshotData, &createdAt)

	if err != nil {
		return nil, fmt.Errorf("snapshot not found: %w", err)
	}

	// Parse snapshot
	var snapshot WorkspaceSnapshot
	if err := json.Unmarshal(snapshotData, &snapshot); err != nil {
		return nil, fmt.Errorf("invalid snapshot data: %w", err)
	}

	// Get current workspace state for comparison
	var currentName, currentSlug string
	var currentSettings json.RawMessage
	err = s.pool.QueryRow(ctx, `
		SELECT name, slug, settings
		FROM workspaces
		WHERE id = $1
	`, workspaceID).Scan(&currentName, &currentSlug, &currentSettings)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch current workspace: %w", err)
	}

	// Count current entities
	var currentRoleCount, currentMemberCount, currentAppCount int
	s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM workspace_roles WHERE workspace_id = $1", workspaceID).Scan(&currentRoleCount)
	s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM workspace_members WHERE workspace_id = $1", workspaceID).Scan(&currentMemberCount)
	s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM osa_generated_apps WHERE workspace_id = $1", workspaceID).Scan(&currentAppCount)

	// Build preview response
	preview := map[string]interface{}{
		"snapshot_info": map[string]interface{}{
			"version":    versionNumber,
			"created_at": createdAt,
		},
		"current_state": map[string]interface{}{
			"name":          currentName,
			"slug":          currentSlug,
			"role_count":    currentRoleCount,
			"member_count":  currentMemberCount,
			"app_count":     currentAppCount,
		},
		"snapshot_state": map[string]interface{}{
			"role_count":    snapshot.Metadata.RoleCount,
			"member_count":  snapshot.Metadata.MemberCount,
			"app_count":     snapshot.Metadata.AppCount,
			"memory_count":  snapshot.Metadata.MemoryCount,
		},
		"changes": map[string]interface{}{
			"roles_diff":     snapshot.Metadata.RoleCount - currentRoleCount,
			"members_diff":   snapshot.Metadata.MemberCount - currentMemberCount,
			"apps_diff":      snapshot.Metadata.AppCount - currentAppCount,
		},
		"details": map[string]interface{}{
			"roles_to_restore":   len(snapshot.Roles),
			"members_to_restore": len(snapshot.Members),
			"apps_to_restore":    len(snapshot.Apps),
			"memories_to_restore": len(snapshot.Memories),
		},
		"warnings": []string{
			"This operation will create a backup of the current state",
			"All current workspace data will be replaced with snapshot data",
			"This action cannot be undone except by restoring another version",
		},
	}

	return preview, nil
}

// restoreSettings restores workspace settings
func (s *WorkspaceVersionService) restoreSettings(
	ctx context.Context,
	tx pgx.Tx,
	workspaceID uuid.UUID,
	settings map[string]interface{},
) error {
	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}

	_, err = tx.Exec(ctx, `
		UPDATE workspaces
		SET settings = $1, updated_at = NOW()
		WHERE id = $2
	`, settingsJSON, workspaceID)

	return err
}

// restoreRoles restores workspace roles
func (s *WorkspaceVersionService) restoreRoles(
	ctx context.Context,
	tx pgx.Tx,
	workspaceID uuid.UUID,
	roles []RoleSnapshot,
) error {
	// Delete existing non-system roles
	_, err := tx.Exec(ctx, `
		DELETE FROM workspace_roles
		WHERE workspace_id = $1 AND is_system = FALSE
	`, workspaceID)
	if err != nil {
		return fmt.Errorf("delete existing roles: %w", err)
	}

	// Restore roles
	for _, role := range roles {
		if role.IsSystem {
			// Skip system roles - don't restore them
			continue
		}

		permissions, _ := json.Marshal(role.Permissions)

		_, err := tx.Exec(ctx, `
			INSERT INTO workspace_roles (
				id, workspace_id, name, display_name, description, color, icon,
				hierarchy_level, is_system, is_default, permissions
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			ON CONFLICT (workspace_id, name) DO UPDATE SET
				display_name = EXCLUDED.display_name,
				description = EXCLUDED.description,
				color = EXCLUDED.color,
				icon = EXCLUDED.icon,
				hierarchy_level = EXCLUDED.hierarchy_level,
				is_default = EXCLUDED.is_default,
				permissions = EXCLUDED.permissions
		`, role.ID, workspaceID, role.Name, role.DisplayName, role.Description,
			role.Color, role.Icon, role.HierarchyLevel, role.IsSystem,
			role.IsDefault, permissions)

		if err != nil {
			return fmt.Errorf("insert role %s: %w", role.Name, err)
		}
	}

	return nil
}

// restoreMembers restores workspace members
func (s *WorkspaceVersionService) restoreMembers(
	ctx context.Context,
	tx pgx.Tx,
	workspaceID uuid.UUID,
	members []MemberSnapshot,
) error {
	// Don't delete existing members - only update/add
	// This prevents accidental removal of current members

	for _, member := range members {
		_, err := tx.Exec(ctx, `
			INSERT INTO workspace_members (
				id, workspace_id, user_id, role_id, role_name, status,
				invited_at, joined_at, invited_by
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (workspace_id, user_id) DO UPDATE SET
				role_id = EXCLUDED.role_id,
				role_name = EXCLUDED.role_name,
				status = EXCLUDED.status
		`, member.ID, workspaceID, member.UserID, member.RoleID, member.RoleName,
			member.Status, member.InvitedAt, member.JoinedAt, member.InvitedBy)

		if err != nil {
			return fmt.Errorf("restore member %s: %w", member.UserID, err)
		}
	}

	return nil
}

// restoreApps restores user-generated apps
func (s *WorkspaceVersionService) restoreApps(
	ctx context.Context,
	tx pgx.Tx,
	workspaceID uuid.UUID,
	apps []AppSnapshot,
) error {
	// Delete existing apps
	_, err := tx.Exec(ctx, `
		DELETE FROM user_generated_apps WHERE workspace_id = $1
	`, workspaceID)
	if err != nil {
		return fmt.Errorf("delete existing apps: %w", err)
	}

	// Restore apps
	for _, app := range apps {
		customConfig, _ := json.Marshal(app.CustomConfig)

		_, err := tx.Exec(ctx, `
			INSERT INTO user_generated_apps (
				id, workspace_id, template_id, app_name, osa_app_id,
				is_visible, is_pinned, is_favorite, position_index,
				custom_config, custom_icon, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())
		`, app.ID, workspaceID, app.TemplateID, app.AppName, app.OsaAppID,
			app.IsVisible, app.IsPinned, app.IsFavorite, app.PositionIndex,
			customConfig, app.CustomIcon)

		if err != nil {
			return fmt.Errorf("restore app %s: %w", app.AppName, err)
		}
	}

	return nil
}

// restoreMemories restores workspace memories
func (s *WorkspaceVersionService) restoreMemories(
	ctx context.Context,
	tx pgx.Tx,
	workspaceID uuid.UUID,
	memories []MemorySnapshot,
) error {
	// Delete existing memories (except user-specific ones)
	_, err := tx.Exec(ctx, `
		DELETE FROM workspace_memories
		WHERE workspace_id = $1 AND visibility = 'workspace'
	`, workspaceID)
	if err != nil {
		return fmt.Errorf("delete existing memories: %w", err)
	}

	// Restore memories
	for _, memory := range memories {
		metadata, _ := json.Marshal(memory.Metadata)

		_, err := tx.Exec(ctx, `
			INSERT INTO workspace_memories (
				id, workspace_id, user_id, title, summary, content,
				memory_type, category, scope_type, scope_id, visibility,
				created_by, importance_score, tags, source, metadata,
				is_pinned, is_active, is_archived, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, NOW())
			ON CONFLICT (id) DO UPDATE SET
				title = EXCLUDED.title,
				summary = EXCLUDED.summary,
				content = EXCLUDED.content,
				importance_score = EXCLUDED.importance_score
		`, memory.ID, workspaceID, memory.UserID, memory.Title, memory.Summary,
			memory.Content, memory.MemoryType, memory.Category, memory.ScopeType,
			memory.ScopeID, memory.Visibility, memory.CreatedBy, memory.ImportanceScore,
			memory.Tags, memory.Source, metadata, memory.IsPinned,
			memory.IsActive, memory.IsArchived)

		if err != nil {
			return fmt.Errorf("restore memory %s: %w", memory.ID, err)
		}
	}

	return nil
}

// ListVersions returns all versions for a workspace
func (s *WorkspaceVersionService) ListVersions(
	ctx context.Context,
	workspaceID uuid.UUID,
) ([]map[string]interface{}, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, version_number, created_by, created_at, snapshot_metadata
		FROM workspace_versions
		WHERE workspace_id = $1
		ORDER BY created_at DESC
	`, workspaceID)

	if err != nil {
		return nil, fmt.Errorf("list versions: %w", err)
	}
	defer rows.Close()

	var versions []map[string]interface{}
	for rows.Next() {
		var id uuid.UUID
		var versionNumber string
		var createdBy *string
		var createdAt time.Time
		var metadata *json.RawMessage

		err := rows.Scan(&id, &versionNumber, &createdBy, &createdAt, &metadata)
		if err != nil {
			return nil, fmt.Errorf("scan version row: %w", err)
		}

		version := map[string]interface{}{
			"id":              id,
			"version_number":  versionNumber,
			"created_by":      createdBy,
			"created_at":      createdAt,
			"snapshot_metadata": metadata,
		}
		versions = append(versions, version)
	}

	return versions, nil
}

// WorkspaceSnapshot represents a complete workspace state
type WorkspaceSnapshot struct {
	Timestamp time.Time                `json:"timestamp"`
	Apps      []AppSnapshot            `json:"apps"`
	Members   []MemberSnapshot         `json:"members"`
	Roles     []RoleSnapshot           `json:"roles"`
	Settings  map[string]interface{}   `json:"settings"`
	Memories  []MemorySnapshot         `json:"memories"`
	Metadata  SnapshotMetadata         `json:"metadata"`
}

type AppSnapshot struct {
	ID           uuid.UUID              `json:"id"`
	AppName      string                 `json:"app_name"`
	TemplateID   *uuid.UUID             `json:"template_id"`
	OsaAppID     *uuid.UUID             `json:"osa_app_id"`
	IsVisible    bool                   `json:"is_visible"`
	IsPinned     bool                   `json:"is_pinned"`
	IsFavorite   bool                   `json:"is_favorite"`
	PositionIndex *int                  `json:"position_index"`
	CustomConfig map[string]interface{} `json:"custom_config"`
	CustomIcon   *string                `json:"custom_icon"`
}

type MemberSnapshot struct {
	ID        uuid.UUID  `json:"id"`
	UserID    string     `json:"user_id"`
	RoleID    *uuid.UUID `json:"role_id"`
	RoleName  string     `json:"role_name"`
	Status    string     `json:"status"`
	InvitedAt *time.Time `json:"invited_at"`
	JoinedAt  time.Time  `json:"joined_at"`
	InvitedBy *string    `json:"invited_by"`
}

type RoleSnapshot struct {
	ID             uuid.UUID              `json:"id"`
	Name           string                 `json:"name"`
	DisplayName    *string                `json:"display_name"`
	Description    *string                `json:"description"`
	Color          *string                `json:"color"`
	Icon           *string                `json:"icon"`
	HierarchyLevel int                    `json:"hierarchy_level"`
	IsSystem       bool                   `json:"is_system"`
	IsDefault      bool                   `json:"is_default"`
	Permissions    map[string]interface{} `json:"permissions"`
}

type MemorySnapshot struct {
	ID              uuid.UUID              `json:"id"`
	UserID          *string                `json:"user_id"`
	Title           *string                `json:"title"`
	Summary         *string                `json:"summary"`
	Content         string                 `json:"content"`
	MemoryType      string                 `json:"memory_type"`
	Category        string                 `json:"category"`
	ScopeType       *string                `json:"scope_type"`
	ScopeID         *uuid.UUID             `json:"scope_id"`
	Visibility      string                 `json:"visibility"`
	CreatedBy       *string                `json:"created_by"`
	ImportanceScore float64                `json:"importance_score"`
	Tags            []string               `json:"tags"`
	Source          *string                `json:"source"`
	Metadata        map[string]interface{} `json:"metadata"`
	IsPinned        bool                   `json:"is_pinned"`
	IsActive        bool                   `json:"is_active"`
	IsArchived      bool                   `json:"is_archived"`
}

type SnapshotMetadata struct {
	AppCount     int     `json:"app_count"`
	MemberCount  int     `json:"member_count"`
	RoleCount    int     `json:"role_count"`
	MemoryCount  int     `json:"memory_count"`
}

// captureWorkspaceState captures current workspace state
func (s *WorkspaceVersionService) captureWorkspaceState(
	ctx context.Context,
	workspaceID uuid.UUID,
) json.RawMessage {
	snapshot := WorkspaceSnapshot{
		Timestamp: time.Now(),
		Apps:      s.captureApps(ctx, workspaceID),
		Members:   s.captureMembers(ctx, workspaceID),
		Roles:     s.captureRoles(ctx, workspaceID),
		Settings:  s.captureSettings(ctx, workspaceID),
		Memories:  s.captureMemories(ctx, workspaceID),
	}

	// Calculate metadata
	snapshot.Metadata = SnapshotMetadata{
		AppCount:    len(snapshot.Apps),
		MemberCount: len(snapshot.Members),
		RoleCount:   len(snapshot.Roles),
		MemoryCount: len(snapshot.Memories),
	}

	data, err := json.Marshal(snapshot)
	if err != nil {
		s.logger.Error("failed to marshal snapshot", "error", err)
		return json.RawMessage("{}")
	}

	return data
}

// captureApps captures all user-generated apps
func (s *WorkspaceVersionService) captureApps(ctx context.Context, workspaceID uuid.UUID) []AppSnapshot {
	rows, err := s.pool.Query(ctx, `
		SELECT id, app_name, template_id, osa_app_id, is_visible, is_pinned, is_favorite,
		       position_index, custom_config, custom_icon
		FROM user_generated_apps
		WHERE workspace_id = $1
		ORDER BY position_index
	`, workspaceID)

	if err != nil {
		s.logger.Error("failed to capture apps", "error", err)
		return []AppSnapshot{}
	}
	defer rows.Close()

	var apps []AppSnapshot
	for rows.Next() {
		var app AppSnapshot
		var customConfig json.RawMessage

		err := rows.Scan(
			&app.ID, &app.AppName, &app.TemplateID, &app.OsaAppID,
			&app.IsVisible, &app.IsPinned, &app.IsFavorite,
			&app.PositionIndex, &customConfig, &app.CustomIcon,
		)
		if err != nil {
			s.logger.Error("failed to scan app", "error", err)
			continue
		}

		// Unmarshal custom config
		if len(customConfig) > 0 {
			json.Unmarshal(customConfig, &app.CustomConfig)
		}

		apps = append(apps, app)
	}

	return apps
}

// captureMembers captures all workspace members
func (s *WorkspaceVersionService) captureMembers(ctx context.Context, workspaceID uuid.UUID) []MemberSnapshot {
	rows, err := s.pool.Query(ctx, `
		SELECT id, user_id, role_id, role_name, status, invited_at, joined_at, invited_by
		FROM workspace_members
		WHERE workspace_id = $1
		ORDER BY joined_at
	`, workspaceID)

	if err != nil {
		s.logger.Error("failed to capture members", "error", err)
		return []MemberSnapshot{}
	}
	defer rows.Close()

	var members []MemberSnapshot
	for rows.Next() {
		var member MemberSnapshot
		err := rows.Scan(
			&member.ID, &member.UserID, &member.RoleID, &member.RoleName,
			&member.Status, &member.InvitedAt, &member.JoinedAt, &member.InvitedBy,
		)
		if err != nil {
			s.logger.Error("failed to scan member", "error", err)
			continue
		}
		members = append(members, member)
	}

	return members
}

// captureRoles captures all workspace roles
func (s *WorkspaceVersionService) captureRoles(ctx context.Context, workspaceID uuid.UUID) []RoleSnapshot {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, display_name, description, color, icon,
		       hierarchy_level, is_system, is_default, permissions
		FROM workspace_roles
		WHERE workspace_id = $1
		ORDER BY hierarchy_level DESC
	`, workspaceID)

	if err != nil {
		s.logger.Error("failed to capture roles", "error", err)
		return []RoleSnapshot{}
	}
	defer rows.Close()

	var roles []RoleSnapshot
	for rows.Next() {
		var role RoleSnapshot
		var permissions json.RawMessage

		err := rows.Scan(
			&role.ID, &role.Name, &role.DisplayName, &role.Description,
			&role.Color, &role.Icon, &role.HierarchyLevel,
			&role.IsSystem, &role.IsDefault, &permissions,
		)
		if err != nil {
			s.logger.Error("failed to scan role", "error", err)
			continue
		}

		// Unmarshal permissions
		if len(permissions) > 0 {
			json.Unmarshal(permissions, &role.Permissions)
		}

		roles = append(roles, role)
	}

	return roles
}

// captureSettings captures workspace settings
func (s *WorkspaceVersionService) captureSettings(ctx context.Context, workspaceID uuid.UUID) map[string]interface{} {
	var settings json.RawMessage
	err := s.pool.QueryRow(ctx, `
		SELECT settings FROM workspaces WHERE id = $1
	`, workspaceID).Scan(&settings)

	if err != nil {
		s.logger.Error("failed to capture settings", "error", err)
		return map[string]interface{}{}
	}

	var result map[string]interface{}
	if len(settings) > 0 {
		json.Unmarshal(settings, &result)
	}

	return result
}

// captureMemories captures workspace memories
func (s *WorkspaceVersionService) captureMemories(ctx context.Context, workspaceID uuid.UUID) []MemorySnapshot {
	rows, err := s.pool.Query(ctx, `
		SELECT id, user_id, title, summary, content, memory_type, category,
		       scope_type, scope_id, visibility, created_by, importance_score,
		       tags, source, metadata, is_pinned, is_active, is_archived
		FROM workspace_memories
		WHERE workspace_id = $1 AND is_active = TRUE AND is_archived = FALSE
		ORDER BY importance_score DESC
	`, workspaceID)

	if err != nil {
		s.logger.Error("failed to capture memories", "error", err)
		return []MemorySnapshot{}
	}
	defer rows.Close()

	var memories []MemorySnapshot
	for rows.Next() {
		var memory MemorySnapshot
		var metadata json.RawMessage

		err := rows.Scan(
			&memory.ID, &memory.UserID, &memory.Title, &memory.Summary,
			&memory.Content, &memory.MemoryType, &memory.Category,
			&memory.ScopeType, &memory.ScopeID, &memory.Visibility,
			&memory.CreatedBy, &memory.ImportanceScore, &memory.Tags,
			&memory.Source, &metadata, &memory.IsPinned,
			&memory.IsActive, &memory.IsArchived,
		)
		if err != nil {
			s.logger.Error("failed to scan memory", "error", err)
			continue
		}

		// Unmarshal metadata
		if len(metadata) > 0 {
			json.Unmarshal(metadata, &memory.Metadata)
		}

		memories = append(memories, memory)
	}

	return memories
}

// VersionDiffResult represents the diff between two workspace versions
type VersionDiffResult struct {
	FromVersion string            `json:"from_version"`
	ToVersion   string            `json:"to_version"`
	Summary     VersionDiffSummary `json:"summary"`
	Files       []FileDiff        `json:"files"`
}

// VersionDiffSummary provides a summary of changes between versions
type VersionDiffSummary struct {
	FilesAdded    int `json:"files_added"`
	FilesRemoved  int `json:"files_removed"`
	FilesModified int `json:"files_modified"`
	FilesUnchanged int `json:"files_unchanged"`
	TotalLinesAdded   int `json:"total_lines_added"`
	TotalLinesRemoved int `json:"total_lines_removed"`
	AppsAdded   int `json:"apps_added"`
	AppsRemoved int `json:"apps_removed"`
}

// FileDiff represents the diff for a single file
type FileDiff struct {
	FilePath   string `json:"file_path"`
	ChangeType string `json:"change_type"` // "added", "removed", "modified", "unchanged"
	Language   string `json:"language,omitempty"`
	FileType   string `json:"file_type,omitempty"`
	OldContent string `json:"old_content,omitempty"`
	NewContent string `json:"new_content,omitempty"`
	UnifiedDiff string `json:"unified_diff,omitempty"`
	LinesAdded   int  `json:"lines_added"`
	LinesRemoved int  `json:"lines_removed"`
}

// CompareVersions compares two workspace snapshots and returns file-level diffs
func (s *WorkspaceVersionService) CompareVersions(
	ctx context.Context,
	workspaceID uuid.UUID,
	fromVersion string,
	toVersion string,
	filterFile string,
) (*VersionDiffResult, error) {
	// Fetch both snapshots
	fromSnapshot, err := s.getSnapshotData(ctx, workspaceID, fromVersion)
	if err != nil {
		return nil, fmt.Errorf("fetch version %s: %w", fromVersion, err)
	}

	toSnapshot, err := s.getSnapshotData(ctx, workspaceID, toVersion)
	if err != nil {
		return nil, fmt.Errorf("fetch version %s: %w", toVersion, err)
	}

	// Build maps of osa_app_ids from each snapshot
	fromAppIDs := extractOsaAppIDs(fromSnapshot)
	toAppIDs := extractOsaAppIDs(toSnapshot)

	// Fetch generated files for each set of app IDs
	fromFiles, err := s.fetchGeneratedFiles(ctx, fromAppIDs)
	if err != nil {
		return nil, fmt.Errorf("fetch files for version %s: %w", fromVersion, err)
	}

	toFiles, err := s.fetchGeneratedFiles(ctx, toAppIDs)
	if err != nil {
		return nil, fmt.Errorf("fetch files for version %s: %w", toVersion, err)
	}

	// Index files by path
	fromFileMap := indexFilesByPath(fromFiles)
	toFileMap := indexFilesByPath(toFiles)

	// Compute diffs
	var fileDiffs []FileDiff

	// Files in toVersion (added or modified)
	for path, toFile := range toFileMap {
		if filterFile != "" && path != filterFile {
			continue
		}
		if fromFile, exists := fromFileMap[path]; exists {
			// File exists in both versions
			if fromFile.ContentHash == toFile.ContentHash {
				fileDiffs = append(fileDiffs, FileDiff{
					FilePath:   path,
					ChangeType: "unchanged",
					Language:   toFile.Language,
					FileType:   toFile.FileType,
				})
			} else {
				diff := computeUnifiedDiff(path, fromFile.Content, toFile.Content)
				fileDiffs = append(fileDiffs, FileDiff{
					FilePath:    path,
					ChangeType:  "modified",
					Language:    toFile.Language,
					FileType:    toFile.FileType,
					OldContent:  fromFile.Content,
					NewContent:  toFile.Content,
					UnifiedDiff: diff.Text,
					LinesAdded:  diff.Added,
					LinesRemoved: diff.Removed,
				})
			}
		} else {
			// File only in toVersion → added
			lineCount := strings.Count(toFile.Content, "\n") + 1
			fileDiffs = append(fileDiffs, FileDiff{
				FilePath:   path,
				ChangeType: "added",
				Language:   toFile.Language,
				FileType:   toFile.FileType,
				NewContent: toFile.Content,
				LinesAdded: lineCount,
			})
		}
	}

	// Files only in fromVersion → removed
	for path, fromFile := range fromFileMap {
		if filterFile != "" && path != filterFile {
			continue
		}
		if _, exists := toFileMap[path]; !exists {
			lineCount := strings.Count(fromFile.Content, "\n") + 1
			fileDiffs = append(fileDiffs, FileDiff{
				FilePath:     path,
				ChangeType:   "removed",
				Language:     fromFile.Language,
				FileType:     fromFile.FileType,
				OldContent:   fromFile.Content,
				LinesRemoved: lineCount,
			})
		}
	}

	// Build summary
	summary := VersionDiffSummary{
		AppsAdded:   countNewApps(fromSnapshot, toSnapshot),
		AppsRemoved: countNewApps(toSnapshot, fromSnapshot),
	}
	for _, fd := range fileDiffs {
		switch fd.ChangeType {
		case "added":
			summary.FilesAdded++
		case "removed":
			summary.FilesRemoved++
		case "modified":
			summary.FilesModified++
		case "unchanged":
			summary.FilesUnchanged++
		}
		summary.TotalLinesAdded += fd.LinesAdded
		summary.TotalLinesRemoved += fd.LinesRemoved
	}

	return &VersionDiffResult{
		FromVersion: fromVersion,
		ToVersion:   toVersion,
		Summary:     summary,
		Files:       fileDiffs,
	}, nil
}

// getSnapshotData fetches and parses a specific workspace version
func (s *WorkspaceVersionService) getSnapshotData(
	ctx context.Context,
	workspaceID uuid.UUID,
	versionNumber string,
) (*WorkspaceSnapshot, error) {
	var snapshotData json.RawMessage
	err := s.pool.QueryRow(ctx, `
		SELECT snapshot_data
		FROM workspace_versions
		WHERE workspace_id = $1 AND version_number = $2
	`, workspaceID, versionNumber).Scan(&snapshotData)

	if err != nil {
		return nil, err
	}

	var snapshot WorkspaceSnapshot
	if err := json.Unmarshal(snapshotData, &snapshot); err != nil {
		return nil, fmt.Errorf("invalid snapshot data: %w", err)
	}

	return &snapshot, nil
}

// generatedFileInfo holds file info for diff comparison
type generatedFileInfo struct {
	FilePath    string
	Content     string
	ContentHash string
	Language    string
	FileType    string
}

// fetchGeneratedFiles fetches all generated files for a set of app IDs
func (s *WorkspaceVersionService) fetchGeneratedFiles(
	ctx context.Context,
	appIDs []uuid.UUID,
) ([]generatedFileInfo, error) {
	if len(appIDs) == 0 {
		return nil, nil
	}

	// Build query with app IDs
	placeholders := make([]string, len(appIDs))
	args := make([]interface{}, len(appIDs))
	for i, id := range appIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT file_path, content, content_hash, COALESCE(language, ''), COALESCE(file_type, '')
		FROM osa_generated_files
		WHERE app_id IN (%s) AND is_latest = true
		ORDER BY file_path
	`, strings.Join(placeholders, ","))

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query generated files: %w", err)
	}
	defer rows.Close()

	var files []generatedFileInfo
	for rows.Next() {
		var f generatedFileInfo
		if err := rows.Scan(&f.FilePath, &f.Content, &f.ContentHash, &f.Language, &f.FileType); err != nil {
			s.logger.Error("failed to scan generated file", "error", err)
			continue
		}
		files = append(files, f)
	}

	return files, nil
}

// extractOsaAppIDs extracts osa_app_ids from a workspace snapshot
func extractOsaAppIDs(snapshot *WorkspaceSnapshot) []uuid.UUID {
	var ids []uuid.UUID
	for _, app := range snapshot.Apps {
		if app.OsaAppID != nil {
			ids = append(ids, *app.OsaAppID)
		}
	}
	return ids
}

// indexFilesByPath creates a map of file path → file info
func indexFilesByPath(files []generatedFileInfo) map[string]generatedFileInfo {
	m := make(map[string]generatedFileInfo, len(files))
	for _, f := range files {
		m[f.FilePath] = f
	}
	return m
}

// countNewApps counts apps in 'to' that are not in 'from' (by app name)
func countNewApps(from, to *WorkspaceSnapshot) int {
	fromNames := make(map[string]bool, len(from.Apps))
	for _, app := range from.Apps {
		fromNames[app.AppName] = true
	}
	count := 0
	for _, app := range to.Apps {
		if !fromNames[app.AppName] {
			count++
		}
	}
	return count
}

// diffResult holds the computed diff text and line counts
type diffResult struct {
	Text    string
	Added   int
	Removed int
}

// computeUnifiedDiff computes a unified diff between two strings
func computeUnifiedDiff(filePath, oldContent, newContent string) diffResult {
	oldLines := strings.Split(oldContent, "\n")
	newLines := strings.Split(newContent, "\n")

	// Simple line-by-line diff using LCS
	lcs := computeLCS(oldLines, newLines)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("--- a/%s\n", filePath))
	sb.WriteString(fmt.Sprintf("+++ b/%s\n", filePath))

	added, removed := 0, 0
	oldIdx, newIdx, lcsIdx := 0, 0, 0

	for lcsIdx < len(lcs) {
		// Output removed lines (in old but not matching LCS)
		for oldIdx < len(oldLines) && oldLines[oldIdx] != lcs[lcsIdx] {
			sb.WriteString(fmt.Sprintf("-%s\n", oldLines[oldIdx]))
			removed++
			oldIdx++
		}
		// Output added lines (in new but not matching LCS)
		for newIdx < len(newLines) && newLines[newIdx] != lcs[lcsIdx] {
			sb.WriteString(fmt.Sprintf("+%s\n", newLines[newIdx]))
			added++
			newIdx++
		}
		// Output context line
		sb.WriteString(fmt.Sprintf(" %s\n", lcs[lcsIdx]))
		oldIdx++
		newIdx++
		lcsIdx++
	}

	// Remaining old lines (removed)
	for oldIdx < len(oldLines) {
		sb.WriteString(fmt.Sprintf("-%s\n", oldLines[oldIdx]))
		removed++
		oldIdx++
	}
	// Remaining new lines (added)
	for newIdx < len(newLines) {
		sb.WriteString(fmt.Sprintf("+%s\n", newLines[newIdx]))
		added++
		newIdx++
	}

	return diffResult{Text: sb.String(), Added: added, Removed: removed}
}

// computeLCS computes the Longest Common Subsequence of two string slices
func computeLCS(a, b []string) []string {
	m, n := len(a), len(b)

	// For large files, limit LCS computation to avoid O(m*n) memory/CPU exhaustion
	// 100K cells ≈ ~800KB memory, ~10-50ms compute — safe for concurrent requests
	if m*n > 100_000 {
		// Fallback: treat entire content as changed
		return nil
	}

	// Standard DP LCS
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else if dp[i-1][j] >= dp[i][j-1] {
				dp[i][j] = dp[i-1][j]
			} else {
				dp[i][j] = dp[i][j-1]
			}
		}
	}

	// Backtrack to find LCS
	result := make([]string, 0, dp[m][n])
	i, j := m, n
	for i > 0 && j > 0 {
		if a[i-1] == b[j-1] {
			result = append(result, a[i-1])
			i--
			j--
		} else if dp[i-1][j] >= dp[i][j-1] {
			i--
		} else {
			j--
		}
	}

	// Reverse
	for left, right := 0, len(result)-1; left < right; left, right = left+1, right-1 {
		result[left], result[right] = result[right], result[left]
	}

	return result
}

// incrementVersion increments semantic version
func incrementVersion(current *string) string {
	if current == nil || *current == "" {
		return "0.0.1"
	}

	// Parse semantic version: "0.0.1" -> [0, 0, 1]
	parts := strings.Split(*current, ".")
	if len(parts) != 3 {
		return "0.0.1"
	}

	// Convert patch version
	var patch int
	_, err := fmt.Sscanf(parts[2], "%d", &patch)
	if err != nil {
		return "0.0.1"
	}

	patch++
	return fmt.Sprintf("%s.%s.%d", parts[0], parts[1], patch)
}
