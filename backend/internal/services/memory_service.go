package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)

// MemoryService handles memory persistence operations
type MemoryService struct {
	pool         *pgxpool.Pool
	embeddingSvc *EmbeddingService
}

// NewMemoryService creates a new memory service
func NewMemoryService(pool *pgxpool.Pool, embeddingSvc *EmbeddingService) *MemoryService {
	return &MemoryService{
		pool:         pool,
		embeddingSvc: embeddingSvc,
	}
}

// ========== NEW WORKSPACE MEMORY METHODS ==========

// WorkspaceMemoryRequest represents a request to create a workspace memory
type WorkspaceMemoryRequest struct {
	WorkspaceID      uuid.UUID
	UserID           string
	Title            string
	Summary          string
	Content          string
	MemoryType       string // 'general', 'decision', 'pattern', 'context', 'learning', 'preference'
	Category         string
	Visibility       string   // 'workspace', 'private', 'shared'
	Tags             []string
	Metadata         map[string]interface{}
	ImportanceScore  float64
	ScopeType        *string    // 'workspace', 'project', 'node'
	ScopeID          *uuid.UUID // project or node ID
}

// WorkspaceMemory represents a memory in the workspace_memories table
type WorkspaceMemory struct {
	ID              uuid.UUID              `json:"id"`
	WorkspaceID     uuid.UUID              `json:"workspace_id"`
	Title           string                 `json:"title"`
	Summary         string                 `json:"summary"`
	Content         string                 `json:"content"`
	MemoryType      string                 `json:"memory_type"`
	Category        string                 `json:"category"`
	Visibility      string                 `json:"visibility"`
	OwnerUserID     *string                `json:"owner_user_id,omitempty"`
	SharedWith      []string               `json:"shared_with,omitempty"`
	Tags            []string               `json:"tags"`
	Metadata        map[string]interface{} `json:"metadata"`
	ImportanceScore float64                `json:"importance_score"`
	AccessCount     int                    `json:"access_count"`
	ScopeType       string                 `json:"scope_type"`
	ScopeID         *uuid.UUID             `json:"scope_id,omitempty"`
	IsPinned        bool                   `json:"is_pinned"`
	IsActive        bool                   `json:"is_active"`
	CreatedBy       string                 `json:"created_by"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	LastAccessedAt  *time.Time             `json:"last_accessed_at,omitempty"`
}

// MemoryQueryOptions represents query options for retrieving memories
type MemoryQueryOptions struct {
	MemoryType *string
	Category   *string
	Tags       []string
	Limit      int
}

// CreateWorkspaceMemory creates a workspace-level memory (accessible to all workspace members)
func (m *MemoryService) CreateWorkspaceMemory(ctx context.Context, req WorkspaceMemoryRequest) (*WorkspaceMemory, error) {
	// Generate embedding
	var embeddingVec interface{}
	if m.embeddingSvc != nil {
		textToEmbed := req.Title + " " + req.Summary + " " + req.Content
		embedding, err := m.embeddingSvc.GenerateEmbedding(ctx, textToEmbed)
		if err == nil && len(embedding) > 0 {
			// Convert to pgvector format
			embeddingVec = pgvector.NewVector(embedding)
		}
	}

	// Prepare tags - pgx can handle []string directly
	tags := req.Tags
	if tags == nil {
		tags = []string{} // Empty slice instead of nil
	}

	// Serialize metadata to JSON
	metadataJSON := "{}"
	if req.Metadata != nil {
		metaBytes, _ := json.Marshal(req.Metadata)
		metadataJSON = string(metaBytes)
	}

	// Visibility must be 'workspace' for workspace memories
	visibility := "workspace"
	if req.Visibility != "" {
		visibility = req.Visibility
	}

	// Set owner_user_id only for private/shared memories
	var ownerUserID *string
	if visibility != "workspace" {
		ownerUserID = &req.UserID
	}

	scopeType := "workspace"
	if req.ScopeType != nil {
		scopeType = *req.ScopeType
	}

	importanceScore := req.ImportanceScore
	if importanceScore == 0 {
		importanceScore = 0.5
	}

	query := `
		INSERT INTO workspace_memories (
			workspace_id, title, summary, content, memory_type, category,
			visibility, owner_user_id, tags, metadata, importance_score,
			scope_type, scope_id, embedding, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10::jsonb, $11, $12, $13, $14, $15)
		RETURNING id, created_at, updated_at
	`

	var memory WorkspaceMemory
	err := m.pool.QueryRow(ctx, query,
		req.WorkspaceID,
		req.Title,
		req.Summary,
		req.Content,
		req.MemoryType,
		req.Category,
		visibility,
		ownerUserID,
		tags, // Pass slice directly
		metadataJSON,
		importanceScore,
		scopeType,
		req.ScopeID,
		embeddingVec,
		req.UserID,
	).Scan(&memory.ID, &memory.CreatedAt, &memory.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create workspace memory: %w", err)
	}

	// Populate return object
	memory.WorkspaceID = req.WorkspaceID
	memory.Title = req.Title
	memory.Summary = req.Summary
	memory.Content = req.Content
	memory.MemoryType = req.MemoryType
	memory.Category = req.Category
	memory.Visibility = visibility
	memory.OwnerUserID = ownerUserID
	memory.Tags = req.Tags
	memory.Metadata = req.Metadata
	memory.ImportanceScore = importanceScore
	memory.ScopeType = scopeType
	memory.ScopeID = req.ScopeID
	memory.CreatedBy = req.UserID
	memory.IsActive = true
	memory.IsPinned = false
	memory.AccessCount = 0

	return &memory, nil
}

// CreateUserMemory creates a private user memory (only owner can access)
func (m *MemoryService) CreateUserMemory(ctx context.Context, workspaceID uuid.UUID, userID string, title, summary, content, memoryType string, tags []string, metadata map[string]interface{}) (*WorkspaceMemory, error) {
	req := WorkspaceMemoryRequest{
		WorkspaceID:     workspaceID,
		UserID:          userID,
		Title:           title,
		Summary:         summary,
		Content:         content,
		MemoryType:      memoryType,
		Visibility:      "private",
		Tags:            tags,
		Metadata:        metadata,
		ImportanceScore: 0.5,
	}
	return m.CreateWorkspaceMemory(ctx, req)
}

// GetWorkspaceMemories retrieves workspace-level memories (accessible to all workspace members)
func (m *MemoryService) GetWorkspaceMemories(ctx context.Context, workspaceID uuid.UUID, userID string, opts MemoryQueryOptions) ([]WorkspaceMemory, error) {
	if opts.Limit == 0 {
		opts.Limit = 50
	}

	var memoryType *string
	if opts.MemoryType != nil {
		memoryType = opts.MemoryType
	}

	rows, err := m.pool.Query(ctx, "SELECT * FROM get_workspace_memories($1, $2, $3, $4)",
		workspaceID, userID, memoryType, opts.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace memories: %w", err)
	}
	defer rows.Close()

	var memories []WorkspaceMemory
	for rows.Next() {
		var mem WorkspaceMemory
		var tagsJSON []byte
		var metadataJSON []byte

		err := rows.Scan(
			&mem.ID,
			&mem.Content,
			&mem.MemoryType,
			&mem.ImportanceScore,
			&tagsJSON,
			&metadataJSON,
			&mem.AccessCount,
			&mem.CreatedAt,
		)
		if err != nil {
			continue
		}

		// Deserialize
		if len(tagsJSON) > 0 {
			if err := json.Unmarshal(tagsJSON, &mem.Tags); err != nil {
				// Log error but continue
				fmt.Printf("warning: failed to unmarshal tags for memory %s: %v\n", mem.ID, err)
			}
		}
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &mem.Metadata); err != nil {
				// Log error but continue
				fmt.Printf("warning: failed to unmarshal metadata for memory %s: %v\n", mem.ID, err)
			}
		}

		mem.WorkspaceID = workspaceID
		mem.Visibility = "workspace"
		memories = append(memories, mem)
	}

	return memories, nil
}

// GetUserMemories retrieves user's private and shared memories
func (m *MemoryService) GetUserMemories(ctx context.Context, workspaceID uuid.UUID, userID string, opts MemoryQueryOptions) ([]WorkspaceMemory, error) {
	if opts.Limit == 0 {
		opts.Limit = 50
	}

	var memoryType *string
	if opts.MemoryType != nil {
		memoryType = opts.MemoryType
	}

	rows, err := m.pool.Query(ctx, "SELECT * FROM get_user_memories($1, $2, $3, $4)",
		workspaceID, userID, memoryType, opts.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get user memories: %w", err)
	}
	defer rows.Close()

	var memories []WorkspaceMemory
	for rows.Next() {
		var mem WorkspaceMemory
		var tagsJSON []byte
		var metadataJSON []byte
		var sharedWithJSON []byte

		err := rows.Scan(
			&mem.ID,
			&mem.Content,
			&mem.MemoryType,
			&mem.ImportanceScore,
			&tagsJSON,
			&metadataJSON,
			&mem.Visibility,
			&sharedWithJSON,
			&mem.AccessCount,
			&mem.CreatedAt,
		)
		if err != nil {
			continue
		}

		// Deserialize
		if len(tagsJSON) > 0 {
			if err := json.Unmarshal(tagsJSON, &mem.Tags); err != nil {
				// Log error but continue
				fmt.Printf("warning: failed to unmarshal tags for memory %s: %v\n", mem.ID, err)
			}
		}
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &mem.Metadata); err != nil {
				// Log error but continue
				fmt.Printf("warning: failed to unmarshal metadata for memory %s: %v\n", mem.ID, err)
			}
		}
		if len(sharedWithJSON) > 0 {
			if err := json.Unmarshal(sharedWithJSON, &mem.SharedWith); err != nil {
				// Log error but continue
				fmt.Printf("warning: failed to unmarshal sharedWith for memory %s: %v\n", mem.ID, err)
			}
		}

		mem.WorkspaceID = workspaceID
		mem.OwnerUserID = &userID
		memories = append(memories, mem)
	}

	return memories, nil
}

// GetAccessibleMemories retrieves all memories accessible to the user (workspace + private + shared)
func (m *MemoryService) GetAccessibleMemories(ctx context.Context, workspaceID uuid.UUID, userID string, opts MemoryQueryOptions) ([]WorkspaceMemory, error) {
	if opts.Limit == 0 {
		opts.Limit = 100
	}

	var memoryType *string
	if opts.MemoryType != nil {
		memoryType = opts.MemoryType
	}

	rows, err := m.pool.Query(ctx, "SELECT * FROM get_accessible_memories($1, $2, $3, $4)",
		workspaceID, userID, memoryType, opts.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get accessible memories: %w", err)
	}
	defer rows.Close()

	var memories []WorkspaceMemory
	for rows.Next() {
		var mem WorkspaceMemory
		var tagsJSON []byte
		var metadataJSON []byte
		var isOwner bool

		err := rows.Scan(
			&mem.ID,
			&mem.Content,
			&mem.MemoryType,
			&mem.Visibility,
			&mem.ImportanceScore,
			&tagsJSON,
			&metadataJSON,
			&isOwner,
			&mem.AccessCount,
			&mem.CreatedAt,
		)
		if err != nil {
			continue
		}

		// Deserialize
		if len(tagsJSON) > 0 {
			if err := json.Unmarshal(tagsJSON, &mem.Tags); err != nil {
				// Log error but continue
				fmt.Printf("warning: failed to unmarshal tags for memory %s: %v\n", mem.ID, err)
			}
		}
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &mem.Metadata); err != nil {
				// Log error but continue
				fmt.Printf("warning: failed to unmarshal metadata for memory %s: %v\n", mem.ID, err)
			}
		}

		mem.WorkspaceID = workspaceID
		if isOwner && mem.Visibility != "workspace" {
			mem.OwnerUserID = &userID
		}
		memories = append(memories, mem)
	}

	return memories, nil
}

// ShareMemory shares a private memory with specific users
func (m *MemoryService) ShareMemory(ctx context.Context, memoryID uuid.UUID, ownerID string, shareWithUserIDs []string) error {
	_, err := m.pool.Exec(ctx, "SELECT share_memory($1, $2, $3)",
		memoryID, ownerID, shareWithUserIDs)
	if err != nil {
		return fmt.Errorf("failed to share memory: %w", err)
	}
	return nil
}

// UnshareMemory makes a shared memory private again
func (m *MemoryService) UnshareMemory(ctx context.Context, memoryID uuid.UUID, ownerID string) error {
	_, err := m.pool.Exec(ctx, "SELECT unshare_memory($1, $2)", memoryID, ownerID)
	if err != nil {
		return fmt.Errorf("failed to unshare memory: %w", err)
	}
	return nil
}

// TrackAccess increments access counter when memory is retrieved
func (m *MemoryService) TrackAccess(ctx context.Context, memoryID uuid.UUID) error {
	_, err := m.pool.Exec(ctx, "SELECT track_memory_access($1)", memoryID)
	if err != nil {
		return fmt.Errorf("failed to track memory access: %w", err)
	}
	return nil
}

// GetWorkspaceMemoryByID retrieves a specific workspace memory by ID
func (m *MemoryService) GetWorkspaceMemoryByID(ctx context.Context, workspaceID uuid.UUID, memoryID uuid.UUID, userID string) (*WorkspaceMemory, error) {
	// First check if user can access this memory
	var canAccess bool
	err := m.pool.QueryRow(ctx, "SELECT can_access_memory($1, $2)", userID, memoryID).Scan(&canAccess)
	if err != nil {
		return nil, fmt.Errorf("failed to check memory access: %w", err)
	}
	if !canAccess {
		return nil, fmt.Errorf("user does not have access to this memory")
	}

	query := `
		SELECT id, workspace_id, title, summary, content, memory_type, category,
		       visibility, owner_user_id, shared_with, tags, metadata,
		       importance_score, access_count, scope_type, scope_id,
		       is_pinned, is_active, created_by, created_at, updated_at, last_accessed_at
		FROM workspace_memories
		WHERE id = $1 AND workspace_id = $2
	`

	var mem WorkspaceMemory
	var tagsJSON []byte
	var metadataJSON []byte
	var sharedWithJSON []byte

	err = m.pool.QueryRow(ctx, query, memoryID, workspaceID).Scan(
		&mem.ID,
		&mem.WorkspaceID,
		&mem.Title,
		&mem.Summary,
		&mem.Content,
		&mem.MemoryType,
		&mem.Category,
		&mem.Visibility,
		&mem.OwnerUserID,
		&sharedWithJSON,
		&tagsJSON,
		&metadataJSON,
		&mem.ImportanceScore,
		&mem.AccessCount,
		&mem.ScopeType,
		&mem.ScopeID,
		&mem.IsPinned,
		&mem.IsActive,
		&mem.CreatedBy,
		&mem.CreatedAt,
		&mem.UpdatedAt,
		&mem.LastAccessedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("memory not found")
		}
		return nil, fmt.Errorf("failed to get workspace memory: %w", err)
	}

	// Deserialize JSON fields
	if len(tagsJSON) > 0 {
		if err := json.Unmarshal(tagsJSON, &mem.Tags); err != nil {
			// Log error but continue - tags are not critical
			fmt.Printf("warning: failed to unmarshal tags for memory %s: %v\n", memoryID, err)
		}
	}
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &mem.Metadata); err != nil {
			// Log error but continue - metadata is not critical
			fmt.Printf("warning: failed to unmarshal metadata for memory %s: %v\n", memoryID, err)
		}
	}
	if len(sharedWithJSON) > 0 {
		if err := json.Unmarshal(sharedWithJSON, &mem.SharedWith); err != nil {
			// Log error but continue - sharedWith is not critical
			fmt.Printf("warning: failed to unmarshal sharedWith for memory %s: %v\n", memoryID, err)
		}
	}

	// Track access synchronously - no goroutine leak
	// Use a short timeout context to avoid blocking
	trackCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := m.TrackAccess(trackCtx, memoryID); err != nil {
		// Log error but don't fail the request
		fmt.Printf("warning: failed to track access for memory %s: %v\n", memoryID, err)
	}

	return &mem, nil
}

// UpdateWorkspaceMemory updates an existing workspace memory
func (m *MemoryService) UpdateWorkspaceMemory(ctx context.Context, memoryID uuid.UUID, userID string, updates map[string]interface{}) error {
	// Check if user can access this memory
	var canAccess bool
	err := m.pool.QueryRow(ctx, "SELECT can_access_memory($1, $2)", userID, memoryID).Scan(&canAccess)
	if err != nil {
		return fmt.Errorf("failed to check memory access: %w", err)
	}
	if !canAccess {
		return fmt.Errorf("user does not have access to this memory")
	}

	// Build dynamic update query
	query := "UPDATE workspace_memories SET updated_at = NOW()"
	args := []interface{}{memoryID}
	argCount := 1

	for key, value := range updates {
		argCount++
		query += fmt.Sprintf(", %s = $%d", key, argCount)
		args = append(args, value)
	}

	query += " WHERE id = $1"

	_, err = m.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update workspace memory: %w", err)
	}

	return nil
}

// DeleteWorkspaceMemory soft-deletes a workspace memory
func (m *MemoryService) DeleteWorkspaceMemory(ctx context.Context, memoryID uuid.UUID, userID string) error {
	query := `
		UPDATE workspace_memories
		SET is_active = FALSE, updated_at = NOW()
		WHERE id = $1 AND (created_by = $2 OR owner_user_id = $2)
	`

	result, err := m.pool.Exec(ctx, query, memoryID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete workspace memory: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("memory not found or user not authorized to delete")
	}

	return nil
}

// ========== LEGACY METHODS (for backward compatibility) ==========

// CreateMemory creates a new memory in the database (legacy table)
func (m *MemoryService) CreateMemory(ctx context.Context, memory *Memory) error {
	// Generate embedding if service is available
	var embeddingVec interface{}
	var embeddingModel *string
	if m.embeddingSvc != nil {
		textToEmbed := memory.Title + " " + memory.Summary + " " + memory.Content
		embedding, err := m.embeddingSvc.GenerateEmbedding(ctx, textToEmbed)
		if err == nil && len(embedding) > 0 {
			// Convert to pgvector format
			embeddingVec = pgvector.NewVector(embedding)
			model := "nomic-embed-text"
			embeddingModel = &model
		}
	}

	// Prepare tags - pgx can handle []string directly
	tags := memory.Tags
	if tags == nil {
		tags = []string{} // Empty slice instead of nil
	}

	// Insert query
	query := `
		INSERT INTO memories (
			user_id, title, summary, content, memory_type, category,
			source_type, source_id, project_id, node_id,
			importance_score, tags, embedding, embedding_model
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, created_at
	`

	err := m.pool.QueryRow(ctx, query,
		memory.UserID,
		memory.Title,
		memory.Summary,
		memory.Content,
		memory.MemoryType,
		memory.Category,
		memory.SourceType,
		memory.SourceID,
		memory.ProjectID,
		memory.NodeID,
		memory.ImportanceScore,
		tags, // Pass slice directly instead of JSON
		embeddingVec,
		embeddingModel,
	).Scan(&memory.ID, &memory.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create memory: %w", err)
	}

	return nil
}

// GetMemory retrieves a memory by ID (legacy table)
func (m *MemoryService) GetMemory(ctx context.Context, userID string, memoryID uuid.UUID) (*Memory, error) {
	query := `
		SELECT id, user_id, title, summary, content, memory_type, category,
		       source_type, source_id, project_id, node_id, importance_score,
		       access_count, is_pinned, tags, created_at, updated_at
		FROM memories
		WHERE id = $1 AND user_id = $2 AND is_active = TRUE
	`

	var memory Memory
	var tagsJSON []byte

	err := m.pool.QueryRow(ctx, query, memoryID, userID).Scan(
		&memory.ID,
		&memory.UserID,
		&memory.Title,
		&memory.Summary,
		&memory.Content,
		&memory.MemoryType,
		&memory.Category,
		&memory.SourceType,
		&memory.SourceID,
		&memory.ProjectID,
		&memory.NodeID,
		&memory.ImportanceScore,
		&memory.AccessCount,
		&memory.IsPinned,
		&tagsJSON,
		&memory.CreatedAt,
		&memory.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get memory: %w", err)
	}

	// Deserialize tags
	if len(tagsJSON) > 0 {
		if err := json.Unmarshal(tagsJSON, &memory.Tags); err != nil {
			// Log error but continue - tags are not critical
			fmt.Printf("warning: failed to unmarshal tags for memory %s: %v\n", memoryID, err)
		}
	}

	return &memory, nil
}

// ListMemories retrieves memories for a user (legacy table)
func (m *MemoryService) ListMemories(ctx context.Context, userID string, memoryType *string, limit int) ([]Memory, error) {
	query := `
		SELECT id, user_id, title, summary, content, memory_type, category,
		       source_type, source_id, project_id, node_id, importance_score,
		       access_count, is_pinned, tags, created_at, updated_at
		FROM memories
		WHERE user_id = $1 AND is_active = TRUE
	`

	args := []interface{}{userID}
	if memoryType != nil && *memoryType != "" {
		query += " AND memory_type = $2"
		args = append(args, *memoryType)
	}

	query += " ORDER BY created_at DESC LIMIT $" + fmt.Sprint(len(args)+1)
	args = append(args, limit)

	rows, err := m.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list memories: %w", err)
	}
	defer rows.Close()

	var memories []Memory
	for rows.Next() {
		var memory Memory
		var tagsJSON []byte

		err := rows.Scan(
			&memory.ID,
			&memory.UserID,
			&memory.Title,
			&memory.Summary,
			&memory.Content,
			&memory.MemoryType,
			&memory.Category,
			&memory.SourceType,
			&memory.SourceID,
			&memory.ProjectID,
			&memory.NodeID,
			&memory.ImportanceScore,
			&memory.AccessCount,
			&memory.IsPinned,
			&tagsJSON,
			&memory.CreatedAt,
			&memory.UpdatedAt,
		)

		if err != nil {
			continue
		}

		// Deserialize tags
		if len(tagsJSON) > 0 {
			if err := json.Unmarshal(tagsJSON, &memory.Tags); err != nil {
				// Log error but continue
				fmt.Printf("warning: failed to unmarshal tags for memory %s: %v\n", memory.ID, err)
			}
		}

		memories = append(memories, memory)
	}

	return memories, nil
}

// SearchByEmbedding performs semantic search on memories using pgvector similarity
func (m *MemoryService) SearchByEmbedding(ctx context.Context, userID string, embedding []float32, limit int) ([]Memory, error) {
	if len(embedding) == 0 {
		return nil, fmt.Errorf("empty embedding provided")
	}

	if limit <= 0 {
		limit = 5
	}

	// Convert embedding to pgvector format
	vec := pgvector.NewVector(embedding)

	// Use pgvector cosine distance operator (<=>) for semantic search
	// Note: Lower distance = higher similarity; <=> uses cosine distance
	query := `
		SELECT id, user_id, title, summary, content, memory_type, category,
		       source_type, source_id, project_id, node_id, importance_score,
		       access_count, is_pinned, tags, created_at, updated_at
		FROM memories
		WHERE user_id = $1 AND is_active = TRUE AND embedding IS NOT NULL
		ORDER BY embedding <=> $2
		LIMIT $3
	`

	rows, err := m.pool.Query(ctx, query, userID, vec, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search memories by embedding: %w", err)
	}
	defer rows.Close()

	var memories []Memory
	for rows.Next() {
		var memory Memory
		var tagsJSON []byte

		err := rows.Scan(
			&memory.ID,
			&memory.UserID,
			&memory.Title,
			&memory.Summary,
			&memory.Content,
			&memory.MemoryType,
			&memory.Category,
			&memory.SourceType,
			&memory.SourceID,
			&memory.ProjectID,
			&memory.NodeID,
			&memory.ImportanceScore,
			&memory.AccessCount,
			&memory.IsPinned,
			&tagsJSON,
			&memory.CreatedAt,
			&memory.UpdatedAt,
		)

		if err != nil {
			continue
		}

		// Deserialize tags
		if len(tagsJSON) > 0 {
			if err := json.Unmarshal(tagsJSON, &memory.Tags); err != nil {
				// Log error but continue
				fmt.Printf("warning: failed to unmarshal tags for memory %s: %v\n", memory.ID, err)
			}
		}

		memories = append(memories, memory)
	}

	return memories, nil
}
