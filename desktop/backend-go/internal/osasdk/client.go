// Package osasdk provides a local stub for the MIOSA OSA SDK.
// This replaces the external github.com/Miosa-osa/sdk-go dependency for ChatmanGPT's
// BusinessOS fork. The original SDK is a private MIOSA repository.
package osasdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Event represents a streaming event from OSA.
type Event struct {
	Type EventType
	Data map[string]any
}

// EventType represents the type of OSA event.
type EventType string

const (
	EventThinking        EventType = "thinking"
	EventResponse        EventType = "response"
	EventSkillStarted    EventType = "skill_started"
	EventSkillCompleted  EventType = "skill_completed"
	EventSkillFailed     EventType = "skill_failed"
	EventError           EventType = "error"
	EventConnected       EventType = "connected"
	EventSignal          EventType = "signal"
	EventStreamingToken  EventType = "streaming_token"
)

// Client is the OSA cloud client interface.
type Client interface {
	Health(ctx context.Context) (*HealthResponse, error)
	Orchestrate(ctx context.Context, req OrchestrateRequest) (*OrchestrateResponse, error)
	GenerateApp(ctx context.Context, req AppGenerationRequest) (*AppGenerationResponse, error)
	GetAppStatus(ctx context.Context, appID string) (*AppStatusResponse, error)
	GetWorkspaces(ctx context.Context) (*WorkspacesResponse, error)
	GenerateAppFromTemplate(ctx context.Context, req GenerateFromTemplateRequest) (*AppGenerationResponse, error)
	Stream(ctx context.Context, sessionID string) (<-chan Event, error)
	LaunchSwarm(ctx context.Context, req SwarmRequest) (*SwarmResponse, error)
	ListSwarms(ctx context.Context) ([]SwarmStatus, error)
	GetSwarm(ctx context.Context, swarmID string) (*SwarmStatus, error)
	CancelSwarm(ctx context.Context, swarmID string) error
	DispatchInstruction(ctx context.Context, agentID string, instruction Instruction) error
	ListTools(ctx context.Context) ([]ToolDefinition, error)
	ExecuteTool(ctx context.Context, toolName string, params map[string]interface{}) (*ToolResult, error)
	Close() error
}

// CloudConfig configures the cloud client connection.
type CloudConfig struct {
	APIKey  string
	BaseURL string
	Timeout time.Duration
}

// HealthResponse is returned by the health check endpoint.
type HealthResponse struct {
	Status    string `json:"status"`
	Version   string `json:"version"`
	Timestamp string `json:"timestamp"`
	Provider  string `json:"provider,omitempty"`
	Model     string `json:"model,omitempty"`
}

// OrchestrateRequest is the payload for OSA orchestration.
type OrchestrateRequest struct {
	UserID      string                 `json:"user_id"`
	Input       string                 `json:"input"`
	SessionID   string                 `json:"session_id,omitempty"`
	Phase       string                 `json:"phase,omitempty"`
	WorkspaceID string                 `json:"workspace_id,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
}

// OrchestrateResponse is the response from OSA orchestration.
type OrchestrateResponse struct {
	SessionID   string                 `json:"session_id"`
	Status      string                 `json:"status"`
	Success     bool                   `json:"success"`
	Output      string                 `json:"output"`
	AgentsUsed  []string               `json:"agents_used"`
	ExecutionMS int64                  `json:"execution_ms"`
	Metadata    map[string]interface{} `json:"metadata"`
	NextStep    string                 `json:"next_step"`
}

// AppGenerationRequest triggers application generation
type AppGenerationRequest struct {
	UserID      string                 `json:"user_id"`
	WorkspaceID string                 `json:"workspace_id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// AppGenerationResponse is returned from app generation
type AppGenerationResponse struct {
	AppID       string                 `json:"app_id"`
	Status      string                 `json:"status"`
	WorkspaceID string                 `json:"workspace_id"`
	Message     string                 `json:"message"`
	Data        map[string]interface{} `json:"data"`
	CreatedAt   string                 `json:"created_at"`
}

// AppStatusResponse contains app generation status
type AppStatusResponse struct {
	AppID       string                 `json:"app_id"`
	Status      string                 `json:"status"`
	Progress    float64                `json:"progress"`
	CurrentStep string                 `json:"current_step"`
	Output      string                 `json:"output"`
	Error       string                 `json:"error"`
	Metadata    map[string]interface{} `json:"metadata"`
	UpdatedAt   string                 `json:"updated_at"`
}

// Workspace represents an OSA workspace
type Workspace struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	OwnerID     string `json:"owner_id"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// WorkspacesResponse contains workspace list
type WorkspacesResponse struct {
	Workspaces []Workspace `json:"workspaces"`
	Total      int         `json:"total"`
}

// GenerateFromTemplateRequest generates app from template
type GenerateFromTemplateRequest struct {
	TemplateName string                 `json:"template_name"`
	Variables    map[string]interface{} `json:"variables"`
	UserID       string                 `json:"user_id"`
	WorkspaceID  string                 `json:"workspace_id,omitempty"`
}

// SwarmRequest initiates a multi-agent swarm
type SwarmRequest struct {
	Pattern   string                 `json:"pattern"`
	Task      string                 `json:"task"`
	Agents    []string               `json:"agents"`
	Config    map[string]interface{} `json:"config,omitempty"`
	MaxAgents int                    `json:"max_agents,omitempty"`
	SessionID string                 `json:"session_id,omitempty"`
}

// SwarmResponse is returned from swarm launch
type SwarmResponse struct {
	SwarmID string    `json:"swarm_id"`
	Status  string    `json:"status"`
	Agents  []string  `json:"agents"`
	Started time.Time `json:"started"`
}

// SwarmStatus contains the status of a swarm
type SwarmStatus struct {
	SwarmID     string                 `json:"swarm_id"`
	Status      string                 `json:"status"`
	Progress    int                    `json:"progress"`
	Agents      []string               `json:"agents"`
	Results     map[string]interface{} `json:"results,omitempty"`
	Error       string                 `json:"error,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
}

// Instruction is sent to a fleet agent
type Instruction struct {
	ID          string                 `json:"id"`
	Action      string                 `json:"action"`
	Params      map[string]interface{} `json:"params"`
	SpecVersion string                 `json:"spec_version,omitempty"`
	Type        string                 `json:"type,omitempty"`
	Source      string                 `json:"source,omitempty"`
	Data        map[string]interface{} `json:"data,omitempty"`
}

// ToolDefinition describes an available tool
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Params      map[string]interface{} `json:"params"`
	Version     string                 `json:"version"`
}

// ToolResult is returned from tool execution
type ToolResult struct {
	Success bool                   `json:"success"`
	Output  map[string]interface{} `json:"output"`
	Error   string                 `json:"error,omitempty"`
}

// LocalConfig configures local SDK connection
type LocalConfig struct {
	BaseURL      string
	SharedSecret string
	Timeout      time.Duration
	Resilience   *ResilienceConfig
}

// ResilienceConfig controls SDK resilience behavior
type ResilienceConfig struct {
	Enabled bool
}

// APIError represents an API error from OSA
type APIError struct {
	StatusCode int
	ErrorCode  string
	Details    string
}

// Error implements the error interface for APIError
func (e *APIError) Error() string {
	if e.Details != "" {
		return e.Details
	}
	return e.ErrorCode
}

// NewLocalClient creates a new local OSA client backed by real HTTP calls.
func NewLocalClient(cfg LocalConfig) (Client, error) {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "http://localhost:8089"
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}
	return &cloudClient{
		baseURL:      cfg.BaseURL,
		timeout:      cfg.Timeout,
		sharedSecret: cfg.SharedSecret,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}, nil
}

// cloudClient is an HTTP-based implementation of the OSA client interface.
type cloudClient struct {
	apiKey       string
	sharedSecret string
	baseURL      string
	timeout      time.Duration
	httpClient   *http.Client
}

// NewCloudClient creates a new OSA cloud client.
func NewCloudClient(cfg CloudConfig) (Client, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.miosa.ai"
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	return &cloudClient{
		apiKey:  cfg.APIKey,
		baseURL: cfg.BaseURL,
		timeout: cfg.Timeout,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}, nil
}


// doRequest performs an HTTP request and decodes the JSON response.
// On non-2xx status it returns an *APIError.
func (c *cloudClient) doRequest(ctx context.Context, method, path string, body interface{}, out interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	} else if c.sharedSecret != "" {
		// Local mode: use shared secret as bearer token so downstream tests/
		// services can verify requests are from an authorised BOS instance.
		req.Header.Set("Authorization", "Bearer "+c.sharedSecret)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Try to extract error details from response body
		var errResp struct {
			Error   string `json:"error"`
			Details string `json:"details"`
			Code    string `json:"code"`
		}
		details := ""
		if jsonErr := json.Unmarshal(respBody, &errResp); jsonErr == nil {
			if errResp.Details != "" {
				details = errResp.Details
			} else if errResp.Error != "" {
				details = errResp.Error
			}
		}
		return &APIError{
			StatusCode: resp.StatusCode,
			ErrorCode:  fmt.Sprintf("HTTP %d", resp.StatusCode),
			Details:    details,
		}
	}

	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}

func (c *cloudClient) Health(ctx context.Context) (*HealthResponse, error) {
	var result HealthResponse
	if err := c.doRequest(ctx, http.MethodGet, "/health", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *cloudClient) Orchestrate(ctx context.Context, req OrchestrateRequest) (*OrchestrateResponse, error) {
	var result OrchestrateResponse
	if err := c.doRequest(ctx, http.MethodPost, "/api/v1/orchestrate", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *cloudClient) GenerateApp(ctx context.Context, req AppGenerationRequest) (*AppGenerationResponse, error) {
	var result AppGenerationResponse
	if err := c.doRequest(ctx, http.MethodPost, "/api/v1/generate", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *cloudClient) GetAppStatus(ctx context.Context, appID string) (*AppStatusResponse, error) {
	var result AppStatusResponse
	if err := c.doRequest(ctx, http.MethodGet, "/api/v1/apps/"+appID+"/status", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *cloudClient) GetWorkspaces(ctx context.Context) (*WorkspacesResponse, error) {
	var result WorkspacesResponse
	if err := c.doRequest(ctx, http.MethodGet, "/api/v1/workspaces", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *cloudClient) GenerateAppFromTemplate(ctx context.Context, req GenerateFromTemplateRequest) (*AppGenerationResponse, error) {
	var result AppGenerationResponse
	if err := c.doRequest(ctx, http.MethodPost, "/api/v1/generate/template", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *cloudClient) Stream(ctx context.Context, sessionID string) (<-chan Event, error) {
	ch := make(chan Event)
	close(ch)
	return ch, nil
}

func (c *cloudClient) LaunchSwarm(ctx context.Context, req SwarmRequest) (*SwarmResponse, error) {
	var result SwarmResponse
	if err := c.doRequest(ctx, http.MethodPost, "/api/v1/swarms", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *cloudClient) ListSwarms(ctx context.Context) ([]SwarmStatus, error) {
	var result []SwarmStatus
	if err := c.doRequest(ctx, http.MethodGet, "/api/v1/swarms", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *cloudClient) GetSwarm(ctx context.Context, swarmID string) (*SwarmStatus, error) {
	var result SwarmStatus
	if err := c.doRequest(ctx, http.MethodGet, "/api/v1/swarms/"+swarmID, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *cloudClient) CancelSwarm(ctx context.Context, swarmID string) error {
	return c.doRequest(ctx, http.MethodDelete, "/api/v1/swarms/"+swarmID, nil, nil)
}

func (c *cloudClient) DispatchInstruction(ctx context.Context, agentID string, instruction Instruction) error {
	return c.doRequest(ctx, http.MethodPost, "/api/v1/agents/"+agentID+"/instructions", instruction, nil)
}

func (c *cloudClient) ListTools(ctx context.Context) ([]ToolDefinition, error) {
	var result []ToolDefinition
	if err := c.doRequest(ctx, http.MethodGet, "/api/v1/tools", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *cloudClient) ExecuteTool(ctx context.Context, toolName string, params map[string]interface{}) (*ToolResult, error) {
	var result ToolResult
	body := map[string]interface{}{
		"tool":   toolName,
		"params": params,
	}
	if err := c.doRequest(ctx, http.MethodPost, "/api/v1/tools/execute", body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *cloudClient) Close() error {
	return nil
}
