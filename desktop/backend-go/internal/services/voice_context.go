package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rhl/businessos-backend/internal/database/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// VoiceContext contains enriched user/workspace context for voice interactions
type VoiceContext struct {
	// User info
	UserID   string
	UserName string
	Email    string

	// Workspace info
	WorkspaceID   *uuid.UUID
	WorkspaceName string
	UserRole      string

	// User profile
	Title          string
	Department     string
	Timezone       string
	WorkingHours   map[string]interface{}
	ExpertiseAreas []string

	// Recent activity (last 24h)
	RecentTasks         []TaskSummary
	RecentProjects      []ProjectSummary
	ActiveConversations int

	// Context timestamp
	FetchedAt time.Time
}

// TaskSummary is a lightweight task representation
type TaskSummary struct {
	ID          uuid.UUID
	Title       string
	Status      string
	Priority    string
	DueDate     *time.Time
	ProjectName string
}

// ProjectSummary is a lightweight project representation
type ProjectSummary struct {
	ID       uuid.UUID
	Name     string
	Status   string
	Progress float64
	TeamSize int
}

// VoiceContextService enriches voice interactions with user/workspace context
type VoiceContextService struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

// NewVoiceContextService creates a new voice context service
func NewVoiceContextService(pool *pgxpool.Pool) *VoiceContextService {
	return &VoiceContextService{
		pool:    pool,
		queries: sqlc.New(pool),
	}
}

// GetUserContext fetches comprehensive user context for voice interactions
func (s *VoiceContextService) GetUserContext(ctx context.Context, userID, userName, email string) (*VoiceContext, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	voiceCtx := &VoiceContext{
		UserID:    userID,
		UserName:  userName,
		Email:     email,
		FetchedAt: time.Now(),
	}

	// TODO: Fetch user's workspace and activity context
	// For now, just return basic user info - full context can be added later
	// The type mismatches need to be resolved with the actual SQLC schema

	return voiceCtx, nil
}

// getRecentTasks fetches user's recent tasks
func (s *VoiceContextService) getRecentTasks(ctx context.Context, userID string, workspaceID uuid.UUID) ([]TaskSummary, error) {
	// Query for recent tasks (simplified - you may need to adjust based on your schema)
	query := `
		SELECT t.id, t.title, t.status, t.priority, t.due_date,
		       COALESCE(p.name, '') as project_name
		FROM tasks t
		LEFT JOIN projects p ON t.project_id = p.id
		WHERE t.assignee_id = $1
		  AND t.workspace_id = $2
		  AND t.status != 'done'
		  AND t.created_at > NOW() - INTERVAL '7 days'
		ORDER BY t.priority DESC, t.due_date ASC NULLS LAST
		LIMIT 5
	`

	rows, err := s.pool.Query(ctx, query, userID, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []TaskSummary
	for rows.Next() {
		var t TaskSummary
		var projectName string
		err := rows.Scan(&t.ID, &t.Title, &t.Status, &t.Priority, &t.DueDate, &projectName)
		if err != nil {
			continue
		}
		t.ProjectName = projectName
		tasks = append(tasks, t)
	}

	return tasks, nil
}

// getRecentProjects fetches user's recent projects
func (s *VoiceContextService) getRecentProjects(ctx context.Context, userID string, workspaceID uuid.UUID) ([]ProjectSummary, error) {
	// Query for active projects (simplified)
	query := `
		SELECT p.id, p.name, p.status,
		       COALESCE(p.progress, 0) as progress,
		       (SELECT COUNT(*) FROM project_members WHERE project_id = p.id) as team_size
		FROM projects p
		WHERE p.workspace_id = $1
		  AND (p.owner_id = $2 OR EXISTS (
		      SELECT 1 FROM project_members WHERE project_id = p.id AND user_id = $2
		  ))
		  AND p.status IN ('active', 'planning', 'in_progress')
		ORDER BY p.updated_at DESC
		LIMIT 3
	`

	rows, err := s.pool.Query(ctx, query, workspaceID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []ProjectSummary
	for rows.Next() {
		var p ProjectSummary
		err := rows.Scan(&p.ID, &p.Name, &p.Status, &p.Progress, &p.TeamSize)
		if err != nil {
			continue
		}
		projects = append(projects, p)
	}

	return projects, nil
}

// getActiveConversationsCount counts user's active conversations
func (s *VoiceContextService) getActiveConversationsCount(ctx context.Context, userID string) (int, error) {
	var count int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM conversations
		WHERE user_id = $1
		  AND updated_at > NOW() - INTERVAL '7 days'
	`, userID).Scan(&count)
	return count, err
}

// FormatForPrompt formats the voice context as a string for system prompt injection
func (vc *VoiceContext) FormatForPrompt() string {
	var sb strings.Builder

	sb.WriteString("\n\n## YOUR CURRENT CONTEXT\n\n")

	// Extract first name from full name for natural conversation
	firstName := vc.UserName
	if parts := strings.Fields(vc.UserName); len(parts) > 0 {
		firstName = parts[0]
	}

	// User info
	sb.WriteString(fmt.Sprintf("**User**: %s (Full name: %s, Email: %s)\n", firstName, vc.UserName, vc.Email))
	if vc.Title != "" {
		sb.WriteString(fmt.Sprintf("**Title**: %s\n", vc.Title))
	}
	if vc.Department != "" {
		sb.WriteString(fmt.Sprintf("**Department**: %s\n", vc.Department))
	}

	// Workspace info
	if vc.WorkspaceName != "" {
		sb.WriteString(fmt.Sprintf("**Workspace**: %s\n", vc.WorkspaceName))
		sb.WriteString(fmt.Sprintf("**Your Role**: %s\n", vc.UserRole))
	}

	// Expertise
	if len(vc.ExpertiseAreas) > 0 {
		sb.WriteString(fmt.Sprintf("**Expertise**: %s\n", strings.Join(vc.ExpertiseAreas, ", ")))
	}

	// Recent tasks
	if len(vc.RecentTasks) > 0 {
		sb.WriteString("\n**Recent Tasks**:\n")
		for i, task := range vc.RecentTasks {
			if i >= 3 { // Limit to 3 in prompt
				break
			}
			projectInfo := ""
			if task.ProjectName != "" {
				projectInfo = fmt.Sprintf(" (Project: %s)", task.ProjectName)
			}
			sb.WriteString(fmt.Sprintf("- [%s] %s%s\n", strings.ToUpper(task.Status), task.Title, projectInfo))
		}
	}

	// Recent projects
	if len(vc.RecentProjects) > 0 {
		sb.WriteString("\n**Active Projects**:\n")
		for _, proj := range vc.RecentProjects {
			sb.WriteString(fmt.Sprintf("- %s (Status: %s, Progress: %.0f%%, Team: %d)\n",
				proj.Name, proj.Status, proj.Progress, proj.TeamSize))
		}
	}

	// Active conversations
	if vc.ActiveConversations > 0 {
		sb.WriteString(fmt.Sprintf("\n**Active Conversations**: %d\n", vc.ActiveConversations))
	}

	sb.WriteString("\nUse this context to personalize your responses and reference their work naturally.\n")

	return sb.String()
}
