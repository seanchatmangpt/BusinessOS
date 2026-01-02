package services

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ContextTier represents the level of context detail
type ContextTier int

const (
	TierFullContext ContextTier = 1 // Full content, embeddings searched
	TierAwareness   ContextTier = 2 // Titles/summaries only
	TierOnDemand    ContextTier = 3 // Available via tool call
)

// TieredContextRequest contains all context selection parameters
type TieredContextRequest struct {
	UserID      string
	ContextIDs  []uuid.UUID // Selected contexts (Level 1)
	ProjectID   *uuid.UUID  // Selected project (Level 1)
	NodeID      *uuid.UUID  // Business node context
	DocumentIDs []uuid.UUID // Attached document IDs for RAG
}

// TieredContext contains all context organized by tier
type TieredContext struct {
	Level1 *FullContext      `json:"level_1"` // Full context for selected items
	Level2 *AwarenessContext `json:"level_2"` // Awareness of related items
	Level3 *OnDemandRegistry `json:"level_3"` // Registry of fetchable items
}

// FullContext contains detailed information for selected items (Level 1)
type FullContext struct {
	Project      *ProjectFullContext  `json:"project,omitempty"`
	Contexts     []ContextFullContext `json:"contexts,omitempty"`
	Tasks        []TaskFullContext    `json:"tasks,omitempty"`
	LinkedClient *ClientFullContext   `json:"linked_client,omitempty"`
	TeamMembers  []TeamMemberContext  `json:"team_members,omitempty"`
	RelevantRAG  []RelevantBlock      `json:"relevant_rag,omitempty"`
	Documents    []DocumentContext    `json:"documents,omitempty"` // Attached documents for RAG
}

// DocumentContext contains document information for context injection
type DocumentContext struct {
	ID          uuid.UUID `json:"id"`
	Filename    string    `json:"filename"`
	DisplayName string    `json:"display_name,omitempty"`
	Content     string    `json:"content,omitempty"`
	ChunkCount  int       `json:"chunk_count"`
	MimeType    string    `json:"mime_type,omitempty"`
}

// ProjectFullContext contains complete project information
type ProjectFullContext struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Status      string    `json:"status"`
	Priority    string    `json:"priority"`
	ClientName  string    `json:"client_name,omitempty"`
	ProjectType string    `json:"project_type,omitempty"`
}

// ContextFullContext contains complete context/document information
type ContextFullContext struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Type          string    `json:"type"`
	Content       string    `json:"content,omitempty"`
	SystemPrompt  string    `json:"system_prompt,omitempty"`
	WordCount     int       `json:"word_count"`
	HasEmbeddings bool      `json:"has_embeddings"`
}

// TaskFullContext contains complete task information
type TaskFullContext struct {
	ID           uuid.UUID `json:"id"`
	Title        string    `json:"title"`
	Description  string    `json:"description,omitempty"`
	Status       string    `json:"status"`
	Priority     string    `json:"priority"`
	DueDate      string    `json:"due_date,omitempty"`
	AssigneeName string    `json:"assignee_name,omitempty"`
}

// ClientFullContext contains complete client information
type ClientFullContext struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email,omitempty"`
	Industry     string    `json:"industry,omitempty"`
	Status       string    `json:"status"`
	ContactCount int       `json:"contact_count"`
}

// TeamMemberContext contains team member summary
type TeamMemberContext struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Role   string    `json:"role"`
	Status string    `json:"status"`
}

// AwarenessContext contains summaries for related items (Level 2)
type AwarenessContext struct {
	OtherProjects   []EntitySummary `json:"other_projects,omitempty"`
	SiblingContexts []EntitySummary `json:"sibling_contexts,omitempty"`
	RelatedClients  []EntitySummary `json:"related_clients,omitempty"`
	NodeInfo        *NodeSummary    `json:"node_info,omitempty"`
}

// EntitySummary provides minimal awareness of an entity
type EntitySummary struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Type string    `json:"type,omitempty"` // For contexts: document, business, etc.
}

// NodeSummary provides business node context
type NodeSummary struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Type          string    `json:"type"`
	Health        string    `json:"health,omitempty"`
	Purpose       string    `json:"purpose,omitempty"`
	ThisWeekFocus []string  `json:"this_week_focus,omitempty"`
}

// OnDemandRegistry tracks what can be fetched on-demand (Level 3)
type OnDemandRegistry struct {
	AvailableEntities []OnDemandEntity `json:"available_entities"`
}

// OnDemandEntity represents an entity that can be fetched
type OnDemandEntity struct {
	Type string    `json:"type"` // project, context, task, client, team_member
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// TieredContextService builds tiered context for AI queries
type TieredContextService struct {
	pool             *pgxpool.Pool
	embeddingService *EmbeddingService
}

// NewTieredContextService creates a new tiered context service
func NewTieredContextService(pool *pgxpool.Pool, embeddingService *EmbeddingService) *TieredContextService {
	return &TieredContextService{
		pool:             pool,
		embeddingService: embeddingService,
	}
}

// BuildTieredContext creates the full tiered context for an AI request
func (s *TieredContextService) BuildTieredContext(
	ctx context.Context,
	req TieredContextRequest,
) (*TieredContext, error) {
	tc := &TieredContext{
		Level1: &FullContext{},
		Level2: &AwarenessContext{},
		Level3: &OnDemandRegistry{},
	}

	// Build Level 1: Full context for selected items
	if err := s.buildLevel1(ctx, req, tc.Level1); err != nil {
		return nil, fmt.Errorf("build level 1: %w", err)
	}

	// Build Level 2: Awareness of related items
	if err := s.buildLevel2(ctx, req, tc.Level1, tc.Level2); err != nil {
		return nil, fmt.Errorf("build level 2: %w", err)
	}

	// Build Level 3: Registry of on-demand entities
	if err := s.buildLevel3(ctx, req, tc.Level3); err != nil {
		return nil, fmt.Errorf("build level 3: %w", err)
	}

	return tc, nil
}

// buildLevel1 populates full context for selected items
func (s *TieredContextService) buildLevel1(
	ctx context.Context,
	req TieredContextRequest,
	level1 *FullContext,
) error {
	// 1. Get selected project with full details
	if req.ProjectID != nil {
		project, err := s.getProjectFull(ctx, *req.ProjectID, req.UserID)
		if err == nil {
			level1.Project = project

			// Get project tasks
			tasks, err := s.getProjectTasks(ctx, *req.ProjectID, req.UserID)
			if err == nil {
				level1.Tasks = tasks
			}

			// Get linked client if project has one
			client, err := s.getProjectClient(ctx, *req.ProjectID, req.UserID)
			if err == nil && client != nil {
				level1.LinkedClient = client
			}

			// Get assigned team members
			team, err := s.getProjectTeam(ctx, *req.ProjectID, req.UserID)
			if err == nil {
				level1.TeamMembers = team
			}
		}
	}

	// 2. Get selected contexts with full details
	for _, ctxID := range req.ContextIDs {
		doc, err := s.getContextFull(ctx, ctxID, req.UserID)
		if err == nil {
			level1.Contexts = append(level1.Contexts, *doc)
		}
	}

	// 3. Get attached documents with full content
	for _, docID := range req.DocumentIDs {
		doc, err := s.getDocumentFull(ctx, docID, req.UserID)
		if err == nil {
			level1.Documents = append(level1.Documents, *doc)
		}
	}

	return nil
}

// getDocumentFull retrieves full document content for context injection
// Waits for document processing to complete if needed
func (s *TieredContextService) getDocumentFull(ctx context.Context, docID uuid.UUID, userID string) (*DocumentContext, error) {
	// Query from uploaded_documents table with extracted_text column
	query := `
		SELECT d.id, d.original_filename, d.display_name, d.mime_type,
			   COALESCE(d.extracted_text, '') as content,
			   d.processing_status,
			   (SELECT COUNT(*) FROM document_chunks WHERE document_id = d.id) as chunk_count
		FROM uploaded_documents d
		WHERE d.id = $1 AND d.user_id = $2
	`

	// Poll for document processing completion (max 10 seconds)
	maxWait := 10
	for attempt := 0; attempt < maxWait; attempt++ {
		row := s.pool.QueryRow(ctx, query, docID, userID)

		var doc DocumentContext
		var content string
		var status string
		err := row.Scan(&doc.ID, &doc.Filename, &doc.DisplayName, &doc.MimeType, &content, &status, &doc.ChunkCount)
		if err != nil {
			if attempt < maxWait-1 {
				// Document might not be inserted yet, wait and retry
				slog.Info("Document not found, waiting...", "docID", docID, "attempt", attempt+1)
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(time.Second):
					continue
				}
			}
			slog.Error("Failed to get document for context", "docID", docID, "userID", userID, "error", err)
			return nil, err
		}

		// If still processing and content is empty, wait for extraction
		if status == "processing" && content == "" {
			if attempt < maxWait-1 {
				slog.Info("Document still processing, waiting...", "docID", docID, "status", status, "attempt", attempt+1)
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(time.Second):
					continue
				}
			}
		}

		slog.Info("Document retrieved for context injection",
			"docID", docID,
			"filename", doc.Filename,
			"status", status,
			"contentLength", len(content),
			"chunkCount", doc.ChunkCount)

		// Truncate content if too large (max 50KB for context injection)
		const maxContentLen = 50 * 1024
		if len(content) > maxContentLen {
			doc.Content = content[:maxContentLen] + "\n\n[Content truncated - document is too large]"
		} else {
			doc.Content = content
		}

		return &doc, nil
	}

	return nil, fmt.Errorf("timeout waiting for document processing")
}

// buildLevel2 populates awareness context for related items
func (s *TieredContextService) buildLevel2(
	ctx context.Context,
	req TieredContextRequest,
	level1 *FullContext,
	level2 *AwarenessContext,
) error {
	// 1. Get node info if provided
	if req.NodeID != nil {
		node, err := s.getNodeSummary(ctx, *req.NodeID, req.UserID)
		if err == nil {
			level2.NodeInfo = node
		}

		// Get other projects (not the selected one)
		projects, err := s.getOtherProjects(ctx, req.ProjectID, req.UserID, 10)
		if err == nil {
			level2.OtherProjects = projects
		}
	}

	// 2. Get sibling contexts (same parent as selected contexts)
	if len(req.ContextIDs) > 0 {
		siblings, err := s.getSiblingContexts(ctx, req.ContextIDs, req.UserID, 10)
		if err == nil {
			level2.SiblingContexts = siblings
		}
	}

	// 3. Get related clients
	clients, err := s.getRelatedClients(ctx, req.UserID, 5)
	if err == nil {
		// Filter out the linked client if already in Level 1
		for _, c := range clients {
			if level1.LinkedClient == nil || c.ID != level1.LinkedClient.ID {
				level2.RelatedClients = append(level2.RelatedClients, c)
			}
		}
	}

	return nil
}

// buildLevel3 builds the on-demand entity registry
func (s *TieredContextService) buildLevel3(
	ctx context.Context,
	req TieredContextRequest,
	level3 *OnDemandRegistry,
) error {
	// Get a lightweight registry of all fetchable entities
	// This tells the AI what it CAN fetch if needed

	// Projects
	projects, _ := s.getAllProjectNames(ctx, req.UserID, 20)
	for _, p := range projects {
		level3.AvailableEntities = append(level3.AvailableEntities, OnDemandEntity{
			Type: "project",
			ID:   p.ID,
			Name: p.Name,
		})
	}

	// Contexts (documents)
	contexts, _ := s.getAllContextNames(ctx, req.UserID, 30)
	for _, c := range contexts {
		level3.AvailableEntities = append(level3.AvailableEntities, OnDemandEntity{
			Type: "context",
			ID:   c.ID,
			Name: c.Name,
		})
	}

	// Clients
	clients, _ := s.getAllClientNames(ctx, req.UserID, 10)
	for _, c := range clients {
		level3.AvailableEntities = append(level3.AvailableEntities, OnDemandEntity{
			Type: "client",
			ID:   c.ID,
			Name: c.Name,
		})
	}

	return nil
}

// ScopedRAGSearch performs embedding search ONLY within specified contexts
func (s *TieredContextService) ScopedRAGSearch(
	ctx context.Context,
	query string,
	contextIDs []uuid.UUID,
	userID string,
	limit int,
) ([]RelevantBlock, error) {
	if len(contextIDs) == 0 || s.embeddingService == nil {
		return nil, nil
	}

	// Use the embedding service's scoped search
	return s.embeddingService.ScopedSimilaritySearch(ctx, query, contextIDs, userID, limit)
}

// FormatForAI formats the tiered context as a system prompt string
func (tc *TieredContext) FormatForAI() string {
	var sb strings.Builder

	sb.WriteString("## Context Overview\n\n")

	// Level 1: Full Context
	if tc.Level1 != nil {
		sb.WriteString("### Primary Focus (Full Details)\n\n")

		if tc.Level1.Project != nil {
			sb.WriteString(fmt.Sprintf("**Active Project: %s**\n", tc.Level1.Project.Name))
			sb.WriteString(fmt.Sprintf("- Status: %s | Priority: %s\n",
				tc.Level1.Project.Status, tc.Level1.Project.Priority))
			if tc.Level1.Project.Description != "" {
				sb.WriteString(fmt.Sprintf("- Description: %s\n", tc.Level1.Project.Description))
			}
			if tc.Level1.Project.ClientName != "" {
				sb.WriteString(fmt.Sprintf("- Client: %s\n", tc.Level1.Project.ClientName))
			}
			sb.WriteString("\n")

			// Tasks
			if len(tc.Level1.Tasks) > 0 {
				sb.WriteString("**Project Tasks:**\n")
				for _, task := range tc.Level1.Tasks {
					sb.WriteString(fmt.Sprintf("- [%s] %s (%s)", task.Status, task.Title, task.Priority))
					if task.DueDate != "" {
						sb.WriteString(fmt.Sprintf(" - Due: %s", task.DueDate))
					}
					if task.AssigneeName != "" {
						sb.WriteString(fmt.Sprintf(" - Assignee: %s", task.AssigneeName))
					}
					sb.WriteString("\n")
				}
				sb.WriteString("\n")
			}
		}

		// Selected Documents
		if len(tc.Level1.Contexts) > 0 {
			sb.WriteString("**Selected Documents:**\n")
			for _, doc := range tc.Level1.Contexts {
				sb.WriteString(fmt.Sprintf("- **%s** (%s, %d words)\n", doc.Name, doc.Type, doc.WordCount))
				if doc.SystemPrompt != "" {
					sb.WriteString(fmt.Sprintf("  System context: %s\n", truncateText(doc.SystemPrompt, 200)))
				}
				if doc.Content != "" {
					content := truncateText(doc.Content, 1500)
					sb.WriteString(fmt.Sprintf("  Content:\n  > %s\n", content))
				}
			}
			sb.WriteString("\n")
		}

		// Linked Client
		if tc.Level1.LinkedClient != nil {
			sb.WriteString(fmt.Sprintf("**Linked Client: %s**\n", tc.Level1.LinkedClient.Name))
			sb.WriteString(fmt.Sprintf("- Status: %s | Industry: %s\n",
				tc.Level1.LinkedClient.Status, tc.Level1.LinkedClient.Industry))
			sb.WriteString("\n")
		}

		// Team Members
		if len(tc.Level1.TeamMembers) > 0 {
			sb.WriteString("**Team Members:**\n")
			for _, tm := range tc.Level1.TeamMembers {
				sb.WriteString(fmt.Sprintf("- %s (%s) - %s\n", tm.Name, tm.Role, tm.Status))
			}
			sb.WriteString("\n")
		}

		// RAG Results
		if len(tc.Level1.RelevantRAG) > 0 {
			sb.WriteString("**Relevant Knowledge (from selected documents):**\n")
			for i, block := range tc.Level1.RelevantRAG {
				sb.WriteString(fmt.Sprintf("%d. From \"%s\" (%.0f%% match):\n",
					i+1, block.DocumentName, block.Similarity*100))
				sb.WriteString(fmt.Sprintf("   > %s\n", truncateText(block.BlockContent, 300)))
			}
			sb.WriteString("\n")
		}

		// Attached Documents (uploaded by user)
		if len(tc.Level1.Documents) > 0 {
			sb.WriteString("**Attached Documents (uploaded by user):**\n")
			for _, doc := range tc.Level1.Documents {
				displayName := doc.DisplayName
				if displayName == "" {
					displayName = doc.Filename
				}
				sb.WriteString(fmt.Sprintf("\n--- Document: %s ---\n", displayName))
				if doc.Content != "" {
					sb.WriteString(fmt.Sprintf("%s\n", doc.Content))
				} else {
					sb.WriteString(fmt.Sprintf("[Document has %d chunks - use RAG search for content]\n", doc.ChunkCount))
				}
			}
			sb.WriteString("\n")
		}
	}

	// Level 2: Awareness
	if tc.Level2 != nil && tc.hasLevel2Content() {
		sb.WriteString("### Context Awareness (Summaries Only)\n\n")

		if tc.Level2.NodeInfo != nil {
			sb.WriteString(fmt.Sprintf("**Business Node: %s** (%s)\n",
				tc.Level2.NodeInfo.Name, tc.Level2.NodeInfo.Type))
			if tc.Level2.NodeInfo.Purpose != "" {
				sb.WriteString(fmt.Sprintf("- Purpose: %s\n", tc.Level2.NodeInfo.Purpose))
			}
			if tc.Level2.NodeInfo.Health != "" {
				sb.WriteString(fmt.Sprintf("- Health: %s\n", tc.Level2.NodeInfo.Health))
			}
			sb.WriteString("\n")
		}

		if len(tc.Level2.OtherProjects) > 0 {
			sb.WriteString("**Other Projects in Scope:** ")
			names := make([]string, len(tc.Level2.OtherProjects))
			for i, p := range tc.Level2.OtherProjects {
				names[i] = p.Name
			}
			sb.WriteString(strings.Join(names, ", "))
			sb.WriteString("\n\n")
		}

		if len(tc.Level2.SiblingContexts) > 0 {
			sb.WriteString("**Related Documents:** ")
			names := make([]string, len(tc.Level2.SiblingContexts))
			for i, c := range tc.Level2.SiblingContexts {
				names[i] = c.Name
			}
			sb.WriteString(strings.Join(names, ", "))
			sb.WriteString("\n\n")
		}

		if len(tc.Level2.RelatedClients) > 0 {
			sb.WriteString("**Other Clients:** ")
			names := make([]string, len(tc.Level2.RelatedClients))
			for i, c := range tc.Level2.RelatedClients {
				names[i] = c.Name
			}
			sb.WriteString(strings.Join(names, ", "))
			sb.WriteString("\n\n")
		}
	}

	// Level 3: On-Demand Notice
	if tc.Level3 != nil && len(tc.Level3.AvailableEntities) > 0 {
		sb.WriteString("### On-Demand Context\n")
		sb.WriteString("You can use the `get_entity_context` tool to retrieve full details for any entity mentioned above or from the user's workspace.\n\n")
	}

	return sb.String()
}

func (tc *TieredContext) hasLevel2Content() bool {
	if tc.Level2 == nil {
		return false
	}
	return tc.Level2.NodeInfo != nil ||
		len(tc.Level2.OtherProjects) > 0 ||
		len(tc.Level2.SiblingContexts) > 0 ||
		len(tc.Level2.RelatedClients) > 0
}

// Helper functions for database queries

func (s *TieredContextService) getProjectFull(ctx context.Context, projectID uuid.UUID, userID string) (*ProjectFullContext, error) {
	query := `
		SELECT p.id, p.name, COALESCE(p.description, ''), p.status, p.priority,
			   COALESCE(p.client_name, ''), COALESCE(p.project_type, '')
		FROM projects p
		WHERE p.id = $1 AND p.user_id = $2`

	var project ProjectFullContext
	err := s.pool.QueryRow(ctx, query, projectID, userID).Scan(
		&project.ID, &project.Name, &project.Description, &project.Status,
		&project.Priority, &project.ClientName, &project.ProjectType,
	)
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (s *TieredContextService) getProjectTasks(ctx context.Context, projectID uuid.UUID, userID string) ([]TaskFullContext, error) {
	query := `
		SELECT t.id, t.title, COALESCE(t.description, ''), t.status, t.priority,
			   COALESCE(to_char(t.due_date, 'YYYY-MM-DD'), ''),
			   COALESCE(tm.name, '')
		FROM tasks t
		LEFT JOIN team_members tm ON tm.id = t.assignee_id
		WHERE t.project_id = $1 AND t.user_id = $2
		ORDER BY
			CASE t.priority WHEN 'critical' THEN 1 WHEN 'high' THEN 2 WHEN 'medium' THEN 3 ELSE 4 END,
			t.due_date ASC NULLS LAST
		LIMIT 20`

	rows, err := s.pool.Query(ctx, query, projectID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []TaskFullContext
	for rows.Next() {
		var task TaskFullContext
		if err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.Status,
			&task.Priority, &task.DueDate, &task.AssigneeName); err != nil {
			continue
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (s *TieredContextService) getProjectClient(ctx context.Context, projectID uuid.UUID, userID string) (*ClientFullContext, error) {
	query := `
		SELECT c.id, c.name, COALESCE(c.email, ''), COALESCE(c.industry, ''), c.status,
			   (SELECT COUNT(*) FROM client_contacts cc WHERE cc.client_id = c.id)
		FROM clients c
		JOIN projects p ON p.client_name = c.name
		WHERE p.id = $1 AND c.user_id = $2
		LIMIT 1`

	var client ClientFullContext
	err := s.pool.QueryRow(ctx, query, projectID, userID).Scan(
		&client.ID, &client.Name, &client.Email, &client.Industry, &client.Status, &client.ContactCount,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &client, nil
}

func (s *TieredContextService) getProjectTeam(ctx context.Context, projectID uuid.UUID, userID string) ([]TeamMemberContext, error) {
	query := `
		SELECT DISTINCT tm.id, tm.name, tm.role, tm.status
		FROM team_members tm
		JOIN tasks t ON t.assignee_id = tm.id
		WHERE t.project_id = $1 AND tm.user_id = $2
		LIMIT 10`

	rows, err := s.pool.Query(ctx, query, projectID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var team []TeamMemberContext
	for rows.Next() {
		var tm TeamMemberContext
		if err := rows.Scan(&tm.ID, &tm.Name, &tm.Role, &tm.Status); err != nil {
			continue
		}
		team = append(team, tm)
	}
	return team, nil
}

func (s *TieredContextService) getContextFull(ctx context.Context, contextID uuid.UUID, userID string) (*ContextFullContext, error) {
	query := `
		SELECT c.id, c.name, c.type::text, COALESCE(c.content, ''),
			   COALESCE(c.system_prompt_template, ''), COALESCE(c.word_count, 0),
			   EXISTS(SELECT 1 FROM context_embeddings ce WHERE ce.context_id = c.id)
		FROM contexts c
		WHERE c.id = $1 AND c.user_id = $2 AND c.is_archived = false`

	var doc ContextFullContext
	err := s.pool.QueryRow(ctx, query, contextID, userID).Scan(
		&doc.ID, &doc.Name, &doc.Type, &doc.Content, &doc.SystemPrompt,
		&doc.WordCount, &doc.HasEmbeddings,
	)
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func (s *TieredContextService) getNodeSummary(ctx context.Context, nodeID uuid.UUID, userID string) (*NodeSummary, error) {
	query := `
		SELECT n.id, n.name, n.type::text, COALESCE(n.health::text, ''),
			   COALESCE(n.purpose, '')
		FROM nodes n
		WHERE n.id = $1 AND n.user_id = $2 AND n.is_archived = false`

	var node NodeSummary
	err := s.pool.QueryRow(ctx, query, nodeID, userID).Scan(
		&node.ID, &node.Name, &node.Type, &node.Health, &node.Purpose,
	)
	if err != nil {
		return nil, err
	}
	return &node, nil
}

func (s *TieredContextService) getOtherProjects(ctx context.Context, excludeID *uuid.UUID, userID string, limit int) ([]EntitySummary, error) {
	query := `
		SELECT p.id, p.name, p.status::text
		FROM projects p
		WHERE p.user_id = $1 AND ($2::uuid IS NULL OR p.id != $2)
		ORDER BY p.updated_at DESC
		LIMIT $3`

	rows, err := s.pool.Query(ctx, query, userID, excludeID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []EntitySummary
	for rows.Next() {
		var p EntitySummary
		if err := rows.Scan(&p.ID, &p.Name, &p.Type); err != nil {
			continue
		}
		projects = append(projects, p)
	}
	return projects, nil
}

func (s *TieredContextService) getSiblingContexts(ctx context.Context, selectedIDs []uuid.UUID, userID string, limit int) ([]EntitySummary, error) {
	// Get parent IDs of selected contexts, then get their siblings
	query := `
		SELECT DISTINCT c2.id, c2.name, c2.type::text
		FROM contexts c1
		JOIN contexts c2 ON c2.parent_id = c1.parent_id
		WHERE c1.id = ANY($1)
		  AND c2.id != ALL($1)
		  AND c2.user_id = $2
		  AND c2.is_archived = false
		ORDER BY c2.updated_at DESC
		LIMIT $3`

	rows, err := s.pool.Query(ctx, query, selectedIDs, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var siblings []EntitySummary
	for rows.Next() {
		var s EntitySummary
		if err := rows.Scan(&s.ID, &s.Name, &s.Type); err != nil {
			continue
		}
		siblings = append(siblings, s)
	}
	return siblings, nil
}

func (s *TieredContextService) getRelatedClients(ctx context.Context, userID string, limit int) ([]EntitySummary, error) {
	query := `
		SELECT c.id, c.name, c.status::text
		FROM clients c
		WHERE c.user_id = $1
		ORDER BY c.updated_at DESC
		LIMIT $2`

	rows, err := s.pool.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []EntitySummary
	for rows.Next() {
		var c EntitySummary
		if err := rows.Scan(&c.ID, &c.Name, &c.Type); err != nil {
			continue
		}
		clients = append(clients, c)
	}
	return clients, nil
}

func (s *TieredContextService) getAllProjectNames(ctx context.Context, userID string, limit int) ([]EntitySummary, error) {
	return s.getOtherProjects(ctx, nil, userID, limit)
}

func (s *TieredContextService) getAllContextNames(ctx context.Context, userID string, limit int) ([]EntitySummary, error) {
	query := `
		SELECT c.id, c.name, c.type::text
		FROM contexts c
		WHERE c.user_id = $1 AND c.is_archived = false
		ORDER BY c.updated_at DESC
		LIMIT $2`

	rows, err := s.pool.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contexts []EntitySummary
	for rows.Next() {
		var c EntitySummary
		if err := rows.Scan(&c.ID, &c.Name, &c.Type); err != nil {
			continue
		}
		contexts = append(contexts, c)
	}
	return contexts, nil
}

func (s *TieredContextService) getAllClientNames(ctx context.Context, userID string, limit int) ([]EntitySummary, error) {
	return s.getRelatedClients(ctx, userID, limit)
}

// Note: truncateText is defined in context_builder.go and is used by both services
