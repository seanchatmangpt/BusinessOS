package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
)

// ToolExecutor handles execution of function calls from LLM
type ToolExecutor struct {
	queries *sqlc.Queries
	logger  *slog.Logger
}

// NewToolExecutor creates a new tool executor
func NewToolExecutor(queries *sqlc.Queries, logger *slog.Logger) *ToolExecutor {
	if logger == nil {
		logger = slog.Default()
	}
	return &ToolExecutor{
		queries: queries,
		logger:  logger,
	}
}

// ToolCall represents a function call from the LLM
type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

// ExecuteToolCall executes a single tool call and returns the result
func (e *ToolExecutor) ExecuteToolCall(ctx context.Context, toolCall ToolCall, userID string) (string, error) {
	e.logger.Info("🔧 Executing tool call",
		"tool", toolCall.Function.Name,
		"args", toolCall.Function.Arguments,
		"user_id", userID,
	)

	switch toolCall.Function.Name {
	case "navigate_to_module":
		return e.executeNavigation(ctx, toolCall.Function.Arguments, userID)

	case "create_task":
		return e.executeCreateTask(ctx, toolCall.Function.Arguments, userID)

	case "list_tasks":
		return e.executeListTasks(ctx, toolCall.Function.Arguments, userID)

	case "create_project":
		return e.executeCreateProject(ctx, toolCall.Function.Arguments, userID)

	case "search_context":
		return e.executeSearchContext(ctx, toolCall.Function.Arguments, userID)

	default:
		return "", fmt.Errorf("unknown tool: %s", toolCall.Function.Name)
	}
}

// executeNavigation handles module navigation
func (e *ToolExecutor) executeNavigation(ctx context.Context, argsJSON string, userID string) (string, error) {
	var args struct {
		Module string `json:"module"`
	}

	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	e.logger.Info("🧭 Navigating to module",
		"module", args.Module,
		"user_id", userID,
	)

	// Store navigation command in database for frontend to poll
	// Alternative: Use WebSocket or SSE to push to frontend
	// For now, return success - frontend will handle via transcript parsing or separate polling

	return fmt.Sprintf("Opened %s module", args.Module), nil
}

// executeCreateTask creates a new task
func (e *ToolExecutor) executeCreateTask(ctx context.Context, argsJSON string, userID string) (string, error) {
	var args struct {
		Title    string `json:"title"`
		DueDate  string `json:"due_date"`
		Priority string `json:"priority"`
	}

	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	e.logger.Info("✅ Creating task",
		"title", args.Title,
		"due_date", args.DueDate,
		"priority", args.Priority,
		"user_id", userID,
	)

	// Parse due date if provided
	var dueDate pgtype.Timestamp
	if args.DueDate != "" {
		t, err := time.Parse("2006-01-02", args.DueDate)
		if err == nil {
			dueDate = pgtype.Timestamp{Time: t, Valid: true}
		}
	}

	// Set status to "todo" (default)
	var status sqlc.NullTaskstatus
	status.Taskstatus = "todo"
	status.Valid = true

	// Create task
	task, err := e.queries.CreateTask(ctx, sqlc.CreateTaskParams{
		UserID:      userID,
		Title:       args.Title,
		Description: nil, // No description for voice-created tasks
		Status:      status,
		Priority:    sqlc.NullTaskpriority{}, // No priority
		DueDate:     dueDate,
		ProjectID:   pgtype.UUID{}, // No project
		AssigneeID:  pgtype.UUID{}, // No assignee
	})
	if err != nil {
		e.logger.Error("Failed to create task", "error", err)
		return "", fmt.Errorf("failed to create task: %w", err)
	}

	return fmt.Sprintf("Created task: %s (ID: %s)", task.Title, task.ID), nil
}

// executeListTasks lists user's tasks
func (e *ToolExecutor) executeListTasks(ctx context.Context, argsJSON string, userID string) (string, error) {
	var args struct {
		Status string  `json:"status"`
		Limit  float64 `json:"limit"`
	}

	// Set defaults
	args.Status = "all"
	args.Limit = 10

	if argsJSON != "" {
		json.Unmarshal([]byte(argsJSON), &args)
	}

	e.logger.Info("📋 Listing tasks",
		"status", args.Status,
		"limit", args.Limit,
		"user_id", userID,
	)

	// Get tasks
	tasks, err := e.queries.ListTasks(ctx, sqlc.ListTasksParams{
		UserID: userID,
		Status: sqlc.NullTaskstatus{}, // Get all tasks
	})
	if err != nil {
		return "", fmt.Errorf("failed to list tasks: %w", err)
	}

	if len(tasks) == 0 {
		return "You don't have any tasks yet.", nil
	}

	// Format tasks as text
	result := fmt.Sprintf("You have %d tasks:\n", len(tasks))
	for i, task := range tasks {
		status := "unknown"
		if task.Status.Valid {
			status = string(task.Status.Taskstatus)
		}
		result += fmt.Sprintf("%d. %s (status: %s)\n", i+1, task.Title, status)
	}

	return result, nil
}

// executeCreateProject creates a new project
func (e *ToolExecutor) executeCreateProject(ctx context.Context, argsJSON string, userID string) (string, error) {
	var args struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	e.logger.Info("🚀 Creating project",
		"name", args.Name,
		"user_id", userID,
	)

	// Create project (simplified - adjust based on your schema)
	// This is a placeholder - implement based on your actual project creation logic
	return fmt.Sprintf("Created project: %s", args.Name), nil
}

// executeSearchContext searches user's knowledge base
func (e *ToolExecutor) executeSearchContext(ctx context.Context, argsJSON string, userID string) (string, error) {
	var args struct {
		Query string `json:"query"`
	}

	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	e.logger.Info("🔍 Searching context",
		"query", args.Query,
		"user_id", userID,
	)

	// Implement context search (placeholder)
	return fmt.Sprintf("Search results for: %s (feature coming soon)", args.Query), nil
}
