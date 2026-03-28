// Package stubs_test verifies that stub implementations have been replaced
// with real behaviour (Chicago TDD — black-box assertions, interface injection
// for unavailable external dependencies such as DB and message broker).
package stubs_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/carrier"
	"github.com/rhl/businessos-backend/internal/db"
	"github.com/rhl/businessos-backend/internal/handlers"
	"github.com/rhl/businessos-backend/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// BOS-GO-C3: AdaptivePool — real connection factory, IsValid calls Ping
// ---------------------------------------------------------------------------

// TestAdaptivePool_IsValid_healthy verifies that a connection whose underlying
// Ping returns nil is considered valid (replaces the always-true mock stub).
func TestAdaptivePool_IsValid_healthy(t *testing.T) {
	factory := db.NewMockConnFactory(nil) // nil → Ping succeeds
	conn, err := db.NewPooledConnection("conn-ok", factory)
	require.NoError(t, err)
	assert.True(t, conn.IsValid(), "IsValid must return true when Ping succeeds")
}

// TestAdaptivePool_IsValid_dead verifies that a connection whose Ping returns
// an error is treated as invalid.
func TestAdaptivePool_IsValid_dead(t *testing.T) {
	pingErr := assert.AnError
	factory := db.NewMockConnFactory(pingErr) // non-nil → Ping fails
	conn, err := db.NewPooledConnection("conn-dead", factory)
	require.NoError(t, err)
	assert.False(t, conn.IsValid(), "IsValid must return false when Ping fails")
}

// TestAdaptivePool_Init_uses_factory ensures AdaptivePool.Init drives all
// connection creation through the injected factory (no createMockConnection).
func TestAdaptivePool_Init_uses_factory(t *testing.T) {
	factory := db.NewMockConnFactory(nil)
	cfg := &db.PoolConfig{MinSize: 3, MaxSize: 10}
	pool := db.NewAdaptivePoolWithFactory(cfg, factory)

	assert.Equal(t, 3, pool.Size(), "pool size must equal MinSize after init")
	assert.Equal(t, 3, factory.AcquireCalls(), "factory.Acquire must be called MinSize times")
}

// ---------------------------------------------------------------------------
// BOS-GO-H4: signal_health handler — real subsystem keys in response
// ---------------------------------------------------------------------------

// TestSignalHealth_response_has_subsystem_keys verifies that the JSON
// response from GetSignalHealth contains all expected top-level keys.
func TestSignalHealth_response_has_subsystem_keys(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/signal/health", nil)

	h := handlers.NewHandlersForTest(nil, nil)
	h.GetSignalHealth(c)

	require.Equal(t, http.StatusOK, w.Code)

	var raw map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &raw))

	for _, key := range []string{"status", "classification", "metrics", "feedback_loop"} {
		assert.Contains(t, raw, key, "response must contain key %q", key)
	}
}

// TestSignalHealth_status_non_empty verifies the overall status is non-empty.
func TestSignalHealth_status_non_empty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/signal/health", nil)

	h := handlers.NewHandlersForTest(nil, nil)
	h.GetSignalHealth(c)

	var resp handlers.SignalHealthResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.NotEmpty(t, resp.Status)
	assert.NotEmpty(t, resp.Classification.Type)
	assert.NotEmpty(t, resp.FeedbackLoop.Interval)
}

// ---------------------------------------------------------------------------
// BOS-GO-H7: proactive consumer broadcasts via SSEBroadcaster interface
// ---------------------------------------------------------------------------

// mockBroadcaster captures SendToAll calls for assertion.
type mockBroadcaster struct {
	events []services.SSEEvent
}

func (b *mockBroadcaster) SendToAll(event services.SSEEvent) {
	b.events = append(b.events, event)
}

// TestProactiveDispatcher_decision_broadcasts verifies that HandleRequestDecision
// sends an SSE event via the injected broadcaster.
func TestProactiveDispatcher_decision_broadcasts(t *testing.T) {
	bc := &mockBroadcaster{}
	disp := carrier.NewTestableProactiveDispatcher(bc)

	cmd := carrier.ActionCommand{
		Type:          "request_decision",
		CorrelationID: "corr-1",
		ExecutionID:   "exec-1",
		StepID:        "step-1",
		OSInstanceID:  "bos-test",
		Params: map[string]any{
			"question": "Approve the budget?",
			"options":  []interface{}{"yes", "no"},
		},
	}

	disp.HandleRequestDecision(cmd)

	require.Len(t, bc.events, 1, "exactly one SSE event per decision request")
	assert.Equal(t, "request_decision", bc.events[0].Type)
}

// TestProactiveDispatcher_signal_broadcasts verifies that HandleProactiveSignal
// sends an SSE event via the injected broadcaster.
func TestProactiveDispatcher_signal_broadcasts(t *testing.T) {
	bc := &mockBroadcaster{}
	disp := carrier.NewTestableProactiveDispatcher(bc)

	cmd := carrier.ActionCommand{
		Type:          "proactive_signal",
		CorrelationID: "corr-2",
		OSInstanceID:  "bos-test",
		Params: map[string]any{
			"signal_type": "metric_alert",
			"severity":    "high",
			"message":     "CPU spike detected",
		},
	}

	disp.HandleProactiveSignal(cmd)

	require.Len(t, bc.events, 1, "exactly one SSE event per proactive signal")
	assert.Equal(t, "proactive_signal", bc.events[0].Type)
}

// ---------------------------------------------------------------------------
// BOS-GO-M9: LinkedIn contacts list — page/page_size produce correct LIMIT/OFFSET
// ---------------------------------------------------------------------------

// TestPagination_page2_size5 verifies that page=2 and page_size=5 produce
// LIMIT 5 OFFSET 5 (i.e. (page-1)*page_size).
func TestPagination_page2_size5(t *testing.T) {
	limit, offset := handlers.ComputePagination(2, 5)
	assert.Equal(t, 5, limit)
	assert.Equal(t, 5, offset)
}

// TestPagination_page1_size20 verifies that page=1 and page_size=20 produce
// LIMIT 20 OFFSET 0.
func TestPagination_page1_size20(t *testing.T) {
	limit, offset := handlers.ComputePagination(1, 20)
	assert.Equal(t, 20, limit)
	assert.Equal(t, 0, offset)
}

// TestPagination_defaults verifies that zero/negative inputs fall back to
// page=1 and page_size=20.
func TestPagination_defaults(t *testing.T) {
	limit, offset := handlers.ComputePagination(0, 0)
	assert.Equal(t, 20, limit, "page_size=0 must default to 20")
	assert.Equal(t, 0, offset, "page=0 must default to page 1 → offset 0")
}

// TestPagination_large_page verifies that page=3 with page_size=10 produces
// LIMIT 10 OFFSET 20.
func TestPagination_large_page(t *testing.T) {
	limit, offset := handlers.ComputePagination(3, 10)
	assert.Equal(t, 10, limit)
	assert.Equal(t, 20, offset)
}

// ---------------------------------------------------------------------------
// BOS-GO-M10: focus templates loaded from YAML with hardcoded fallback
// ---------------------------------------------------------------------------

// TestFocusTemplates_fallback_non_empty verifies that even without a YAML file
// the handler returns at least one template (hardcoded fallback).
func TestFocusTemplates_fallback_non_empty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/focus/templates", nil)

	h := handlers.NewFocusHandler(nil) // nil pool — templates endpoint needs none
	h.GetFocusModeTemplates(c)

	require.Equal(t, http.StatusOK, w.Code)

	var templates []handlers.FocusTemplateResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &templates))

	assert.NotEmpty(t, templates, "fallback must return at least one template")
	for _, tmpl := range templates {
		assert.NotEmpty(t, tmpl.Name, "every template must have a Name")
	}
}

// TestFocusTemplates_yaml_overrides_hardcoded verifies that when a YAML config
// file is provided its templates are returned instead of the hardcoded list.
func TestFocusTemplates_yaml_overrides_hardcoded(t *testing.T) {
	yamlContent := `templates:
  - id: yaml_only
    name: "YAML Deep Work"
    duration_minutes: 90
    display_name: "YAML Deep Work"
    description: "From YAML"
    icon: "zap"
    temperature: 0.5
    max_tokens: 2048
    output_style: "concise"
    auto_search: false
    thinking_enabled: false
`
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "focus_templates.yaml")
	require.NoError(t, os.WriteFile(yamlPath, []byte(yamlContent), 0644))

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/focus/templates", nil)

	h := handlers.NewFocusHandlerWithConfig(nil, yamlPath)
	h.GetFocusModeTemplates(c)

	require.Equal(t, http.StatusOK, w.Code)

	var templates []handlers.FocusTemplateResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &templates))

	require.Len(t, templates, 1, "YAML must override hardcoded list")
	assert.Equal(t, "YAML Deep Work", templates[0].Name)
}
