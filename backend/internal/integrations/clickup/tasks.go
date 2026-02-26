package clickup

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
)

// GetTasks retrieves all tasks from a specific list.
func (p *Provider) GetTasks(ctx context.Context, userID, listID string) ([]Task, error) {
	token, err := p.GetToken(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	url := fmt.Sprintf("%s/list/%s/task", APIURL, listID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", token.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	// Add query parameters for filtering
	query := req.URL.Query()
	query.Add("archived", "false")
	query.Add("include_closed", "true")
	query.Add("subtasks", "true")
	req.URL.RawQuery = query.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api error: status %d", resp.StatusCode)
	}

	var result struct {
		Tasks []Task `json:"tasks"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Tasks, nil
}

// GetTask retrieves a specific task by ID.
func (p *Provider) GetTask(ctx context.Context, userID, taskID string) (*Task, error) {
	token, err := p.GetToken(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	url := fmt.Sprintf("%s/task/%s", APIURL, taskID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", token.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api error: status %d", resp.StatusCode)
	}

	var task Task
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &task, nil
}

// CreateTaskRequest represents the request body for creating a task.
type CreateTaskRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Status      string   `json:"status,omitempty"`
	Priority    int      `json:"priority,omitempty"`
	DueDate     int64    `json:"due_date,omitempty"`
	StartDate   int64    `json:"start_date,omitempty"`
	Assignees   []int    `json:"assignees,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Parent      string   `json:"parent,omitempty"`
}

// CreateTask creates a new task in a specific list.
func (p *Provider) CreateTask(ctx context.Context, userID, listID string, req CreateTaskRequest) (*Task, error) {
	token, err := p.GetToken(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/list/%s/task", APIURL, listID)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Authorization", token.AccessToken)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api error: status %d", resp.StatusCode)
	}

	var task Task
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &task, nil
}

// UpdateTaskRequest represents the request body for updating a task.
type UpdateTaskRequest struct {
	Name        *string  `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	Status      *string  `json:"status,omitempty"`
	Priority    *int     `json:"priority,omitempty"`
	DueDate     *int64   `json:"due_date,omitempty"`
	StartDate   *int64   `json:"start_date,omitempty"`
	Assignees   *struct {
		Add []int `json:"add,omitempty"`
		Rem []int `json:"rem,omitempty"`
	} `json:"assignees,omitempty"`
}

// UpdateTask updates an existing task.
func (p *Provider) UpdateTask(ctx context.Context, userID, taskID string, req UpdateTaskRequest) (*Task, error) {
	token, err := p.GetToken(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/task/%s", APIURL, taskID)
	httpReq, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Authorization", token.AccessToken)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api error: status %d", resp.StatusCode)
	}

	var task Task
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &task, nil
}

// SyncTasks fetches tasks from a ClickUp list and persists them to the database.
func (p *Provider) SyncTasks(ctx context.Context, userID, listID string) (int, error) {
	tasks, err := p.GetTasks(ctx, userID, listID)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch tasks: %w", err)
	}

	queries := sqlc.New(p.pool)
	synced := 0

	for _, task := range tasks {
		// Marshal complex fields to JSON
		assigneesJSON, _ := json.Marshal(task.Assignees)
		creatorJSON, _ := json.Marshal(task.Creator)
		tagsJSON, _ := json.Marshal(task.Tags)

		// Parse timestamps (ClickUp uses milliseconds since epoch as strings)
		parseDateMillis := func(dateStr string) pgtype.Timestamptz {
			var ts pgtype.Timestamptz
			if dateStr == "" {
				return ts
			}
			if millis, err := strconv.ParseInt(dateStr, 10, 64); err == nil {
				ts.Time = time.Unix(0, millis*1000000)
				ts.Valid = true
			}
			return ts
		}

		dueDate := parseDateMillis(task.DueDate)
		startDate := parseDateMillis(task.StartDate)
		dateCreated := parseDateMillis(task.DateCreated)
		dateUpdated := parseDateMillis(task.DateUpdated)
		dateClosed := parseDateMillis(task.DateClosed)

		// Helper to create string pointers
		strPtr := func(s string) *string {
			if s == "" {
				return nil
			}
			return &s
		}

		int64Ptr := func(i int64) *int64 {
			if i == 0 {
				return nil
			}
			return &i
		}

		_, err := queries.UpsertClickUpTask(ctx, sqlc.UpsertClickUpTaskParams{
			UserID:        userID,
			TaskID:        task.ID,
			CustomID:      strPtr(task.CustomID),
			ListID:        task.List.ID,
			FolderID:      strPtr(task.Folder.ID),
			SpaceID:       task.Space.ID,
			Name:          task.Name,
			Description:   strPtr(task.Description),
			Status:        strPtr(task.Status.Status),
			StatusColor:   strPtr(task.Status.Color),
			Priority:      strPtr(task.Priority.Priority),
			PriorityColor: strPtr(task.Priority.Color),
			DueDate:       dueDate,
			StartDate:     startDate,
			DateCreated:   dateCreated,
			DateUpdated:   dateUpdated,
			DateClosed:    dateClosed,
			TimeSpent:     int64Ptr(task.TimeSpent),
			TimeEstimate:  nil,
			ParentTaskID:  strPtr(task.Parent),
			Assignees:     assigneesJSON,
			Creator:       creatorJSON,
			Tags:          tagsJSON,
			Url:           strPtr(task.URL),
		})
		if err != nil {
			fmt.Printf("Failed to upsert task %s: %v\n", task.ID, err)
			continue
		}
		synced++
	}

	return synced, nil
}
