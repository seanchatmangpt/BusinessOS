package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
)

// PM4PyMCPClientInterface defines the operations for pm4py-mcp interaction.
type PM4PyMCPClientInterface interface {
	Discover(ctx context.Context, eventLog json.RawMessage, variant string) (*DiscoverResult, error)
	Conformance(ctx context.Context, eventLog, petriNet json.RawMessage) (*ConformanceResult, error)
	Statistics(ctx context.Context, eventLog json.RawMessage) (*StatisticsResult, error)
	HealthCheck(ctx context.Context) error
}

// PM4PyMCPClient wraps the MCP client for pm4py-mcp specific operations
type PM4PyMCPClient struct {
	mcpClient *MCPClient
	serverURL string
}

// NewPM4PyMCPClient creates a new pm4py-mcp client
// serverURL is the base URL of the pm4py-mcp server (e.g., http://localhost:7015)
func NewPM4PyMCPClient(serverURL string) *PM4PyMCPClient {
	return &PM4PyMCPClient{
		mcpClient: NewMCPClient(serverURL, "none", "", nil),
		serverURL: serverURL,
	}
}

// DiscoverResult represents the response from pm4py_discover
type DiscoverResult struct {
	ModelID     string       `json:"model_id"`
	Algorithm   string       `json:"algorithm"`
	ProcessModel ProcessModel `json:"process_model"`
	TraceEvents int          `json:"trace_events"`
	UniqueTraces int          `json:"unique_traces"`
}

// ProcessModel represents the discovered Petri net
type ProcessModel struct {
	Places      int `json:"places"`
	Transitions int `json:"transitions"`
	Arcs        int `json:"arcs"`
}

// Discover calls pm4py_discover MCP tool
func (p *PM4PyMCPClient) Discover(ctx context.Context, eventLog json.RawMessage, variant string) (*DiscoverResult, error) {
	slog.Debug("pm4py-mcp: calling discover", "variant", variant)

	result, err := p.mcpClient.ExecuteTool(ctx, "pm4py_discover", map[string]interface{}{
		"hours": 24,
		"limit": 1000,
	})
	if err != nil {
		return nil, fmt.Errorf("pm4py_discover: %w", err)
	}

	// Parse result - pm4py-mcp returns JSON string
	resultJSON, ok := result.(string)
	if !ok {
		return nil, fmt.Errorf("expected string result, got %T", result)
	}

	var resp DiscoverResult
	if err := json.Unmarshal([]byte(resultJSON), &resp); err != nil {
		return nil, fmt.Errorf("parse discover response: %w", err)
	}

	return &resp, nil
}

// ConformanceResult represents the response from pm4py_conformance
type ConformanceResult struct {
	Fitness        float64 `json:"fitness"`
	Precision      float64 `json:"precision"`
	Generalization float64 `json:"generalization"`
	Simplicity     float64 `json:"simplicity"`
	TraceEvents    int     `json:"trace_events"`
}

// Conformance calls pm4py_conformance MCP tool
func (p *PM4PyMCPClient) Conformance(ctx context.Context, eventLog, petriNet json.RawMessage) (*ConformanceResult, error) {
	slog.Debug("pm4py-mcp: calling conformance")

	// First discover the model if petriNet is not provided
	if petriNet == nil || len(petriNet) == 0 {
		return nil, fmt.Errorf("petri_net required for conformance check")
	}

	result, err := p.mcpClient.ExecuteTool(ctx, "pm4py_conformance", map[string]interface{}{
		"hours": 24,
		"limit": 1000,
	})
	if err != nil {
		return nil, fmt.Errorf("pm4py_conformance: %w", err)
	}

	resultJSON, ok := result.(string)
	if !ok {
		return nil, fmt.Errorf("expected string result, got %T", result)
	}

	var resp ConformanceResult
	if err := json.Unmarshal([]byte(resultJSON), &resp); err != nil {
		return nil, fmt.Errorf("parse conformance response: %w", err)
	}

	return &resp, nil
}

// StatisticsResult represents the response from pm4py_full
type StatisticsResult struct {
	TraceEvents       int                   `json:"trace_events"`
	UniqueTraces      int                   `json:"unique_traces"`
	UniqueActivities  int                   `json:"unique_activities"`
	ActivityFrequency []ActivityFrequency  `json:"activity_frequency"`
	CaseDuration      CaseDuration          `json:"case_duration"`
	Bottleneck        []BottleneckActivity `json:"bottleneck"`
	Variants          []Variant             `json:"variants"`
}

// ActivityFrequency represents per-activity statistics
type ActivityFrequency struct {
	Activity   string  `json:"activity"`
	Frequency  int     `json:"frequency"`
	Percentage float64 `json:"percentage"`
}

// CaseDuration represents case duration statistics
type CaseDuration struct {
	MinSeconds    float64 `json:"min_seconds"`
	MaxSeconds    float64 `json:"max_seconds"`
	AvgSeconds    float64 `json:"avg_seconds"`
	MedianSeconds float64 `json:"median_seconds"`
}

// BottleneckActivity represents a bottleneck activity
type BottleneckActivity struct {
	Activity string  `json:"activity"`
	MeanMs   float64 `json:"mean_ms"`
	MedianMs float64 `json:"median_ms"`
}

// Variant represents a trace variant
type Variant struct {
	Variant    string  `json:"variant"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

// Statistics calls pm4py_full MCP tool for comprehensive statistics
func (p *PM4PyMCPClient) Statistics(ctx context.Context, eventLog json.RawMessage) (*StatisticsResult, error) {
	slog.Debug("pm4py-mcp: calling statistics (pm4py_full)")

	result, err := p.mcpClient.ExecuteTool(ctx, "pm4py_full", map[string]interface{}{
		"hours": 24,
		"limit": 1000,
	})
	if err != nil {
		return nil, fmt.Errorf("pm4py_full: %w", err)
	}

	resultJSON, ok := result.(string)
	if !ok {
		return nil, fmt.Errorf("expected string result, got %T", result)
	}

	var resp StatisticsResult
	if err := json.Unmarshal([]byte(resultJSON), &resp); err != nil {
		return nil, fmt.Errorf("parse statistics response: %w", err)
	}

	return &resp, nil
}

// HealthCheck performs a health check on the pm4py-mcp server
func (p *PM4PyMCPClient) HealthCheck(ctx context.Context) error {
	return p.mcpClient.HealthCheck(ctx)
}
