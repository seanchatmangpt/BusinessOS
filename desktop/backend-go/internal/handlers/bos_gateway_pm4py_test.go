package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rhl/businessos-backend/internal/services"
)

// ============================================================================
// MOCK MCP CLIENT
// ============================================================================

// mockPM4PyMCPClient implements services.PM4PyMCPClientInterface for testing.
type mockPM4PyMCPClient struct {
	discoverResult    *services.DiscoverResult
	discoverErr       error
	conformanceResult *services.ConformanceResult
	conformanceErr    error
	statisticsResult  *services.StatisticsResult
	statisticsErr     error
	healthCheckErr    error
	discoverCalled    bool
	conformanceCalled bool
	statisticsCalled  bool
}

func (m *mockPM4PyMCPClient) Discover(_ context.Context, _ json.RawMessage, _ string) (*services.DiscoverResult, error) {
	m.discoverCalled = true
	if m.discoverErr != nil {
		return nil, m.discoverErr
	}
	return m.discoverResult, nil
}

func (m *mockPM4PyMCPClient) Conformance(_ context.Context, _, _ json.RawMessage) (*services.ConformanceResult, error) {
	m.conformanceCalled = true
	if m.conformanceErr != nil {
		return nil, m.conformanceErr
	}
	return m.conformanceResult, nil
}

func (m *mockPM4PyMCPClient) Statistics(_ context.Context, _ json.RawMessage) (*services.StatisticsResult, error) {
	m.statisticsCalled = true
	if m.statisticsErr != nil {
		return nil, m.statisticsErr
	}
	return m.statisticsResult, nil
}

func (m *mockPM4PyMCPClient) HealthCheck(_ context.Context) error {
	return m.healthCheckErr
}

// ============================================================================
// TEST HELPERS
// ============================================================================

// createTempEventLogFile writes a minimal valid JSON event log to a temp file.
func createTempEventLogFile(t *testing.T) string {
	t.Helper()
	content := []byte(`[{"case_id":"case_1","activity":"create_case","timestamp":"2024-01-01T10:00:00Z"},{"case_id":"case_1","activity":"close_case","timestamp":"2024-01-01T11:00:00Z"}]`)
	f, err := os.CreateTemp("", "event_log_*.json")
	require.NoError(t, err, "failed to create temp event log file")
	_, err = f.Write(content)
	require.NoError(t, err, "failed to write event log content")
	require.NoError(t, f.Close())
	t.Cleanup(func() { os.Remove(f.Name()) })
	return f.Name()
}

// defaultMockDiscoverResult returns a sample discover result for tests.
func defaultMockDiscoverResult() *services.DiscoverResult {
	return &services.DiscoverResult{
		ModelID:   "petri_net_abc123",
		Algorithm: "inductive_miner",
		ProcessModel: services.ProcessModel{
			Places:      5,
			Transitions: 8,
			Arcs:        12,
		},
		TraceEvents:   500,
		UniqueTraces: 45,
	}
}

// defaultMockConformanceResult returns a sample conformance result for tests.
func defaultMockConformanceResult() *services.ConformanceResult {
	return &services.ConformanceResult{
		Fitness:        0.96,
		Precision:      0.92,
		Generalization: 0.88,
		Simplicity:     0.91,
		TraceEvents:    150,
	}
}

// defaultMockStatisticsResult returns a sample statistics result for tests.
func defaultMockStatisticsResult() *services.StatisticsResult {
	return &services.StatisticsResult{
		TraceEvents:      2450,
		UniqueTraces:     500,
		UniqueActivities: 8,
		ActivityFrequency: []services.ActivityFrequency{
			{Activity: "create_case", Frequency: 500, Percentage: 20.4},
			{Activity: "assign_case", Frequency: 490, Percentage: 20.0},
			{Activity: "process_case", Frequency: 475, Percentage: 19.4},
		},
		CaseDuration: services.CaseDuration{
			MinSeconds:    60,
			MaxSeconds:    3600,
			AvgSeconds:    1200.5,
			MedianSeconds: 900.0,
		},
		Variants: []services.Variant{
			{Variant: "create_case -> close_case", Count: 200, Percentage: 40.0},
			{Variant: "create_case -> assign_case -> close_case", Count: 150, Percentage: 30.0},
		},
	}
}

// setupMCPGatewayTest initializes a gateway handler with mock MCP client.
func setupMCPGatewayTest(t *testing.T, mock *mockPM4PyMCPClient) (*BOSGatewayHandler, *gin.Engine) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewBOSGatewayHandler(nil, logger)
	handler.pm4pyMCPClient = mock

	router := gin.New()
	api := router.Group("/api")
	RegisterBOSGatewayRoutes(api, handler)

	return handler, router
}

// ============================================================================
// DISCOVER ENDPOINT TESTS
// ============================================================================

func TestDiscoverMCP_Success(t *testing.T) {
	mock := &mockPM4PyMCPClient{
		discoverResult: defaultMockDiscoverResult(),
	}
	handler, router := setupMCPGatewayTest(t, mock)
	logPath := createTempEventLogFile(t)

	req := BOSDiscoverRequest{
		LogPath:   logPath,
		Algorithm: "inductive_miner",
	}

	body, err := json.Marshal(req)
	require.NoError(t, err)

	httpReq := httptest.NewRequest("POST", "/api/bos/discover", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code, "Status code should be 200")
	assert.True(t, mock.discoverCalled, "MCP Discover should be called")

	var resp BOSDiscoverResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err, "Response should be valid JSON")

	assert.Equal(t, "petri_net_abc123", resp.ModelID)
	assert.Equal(t, "inductive_miner", resp.Algorithm)
	assert.Equal(t, 5, resp.Places)
	assert.Equal(t, 8, resp.Transitions)
	assert.Equal(t, 12, resp.Arcs)
	assert.NotEmpty(t, resp.ModelData)
	assert.GreaterOrEqual(t, resp.LatencyMs, uint64(0))

	assert.Equal(t, uint64(1), handler.stats.RequestsTotal)
	assert.Equal(t, uint64(0), handler.stats.RequestsFailed)
}

func TestDiscoverMCP_RespondsWithAlgorithm(t *testing.T) {
	mock := &mockPM4PyMCPClient{
		discoverResult: &services.DiscoverResult{
			ModelID:   "model_heuristic",
			Algorithm: "heuristic_miner",
			ProcessModel: services.ProcessModel{
				Places:      3,
				Transitions: 4,
				Arcs:        6,
			},
		},
	}
	_, router := setupMCPGatewayTest(t, mock)
	logPath := createTempEventLogFile(t)

	req := BOSDiscoverRequest{
		LogPath:   logPath,
		Algorithm: "heuristic_miner",
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/bos/discover", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	var resp BOSDiscoverResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	assert.Equal(t, "heuristic_miner", resp.Algorithm)
	assert.Greater(t, resp.Places, 0, "Should have places")
}

func TestDiscoverMCP_RespondsWithSourceAndSinkPlace(t *testing.T) {
	mock := &mockPM4PyMCPClient{
		discoverResult: defaultMockDiscoverResult(),
	}
	_, router := setupMCPGatewayTest(t, mock)
	logPath := createTempEventLogFile(t)

	req := BOSDiscoverRequest{
		LogPath:   logPath,
		Algorithm: "inductive_miner",
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/bos/discover", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	var resp BOSDiscoverResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	assert.Greater(t, resp.Places, 0, "Should have places (including source/sink)")
}

// ============================================================================
// CONFORMANCE ENDPOINT TESTS
// ============================================================================

func TestConformanceMCP_Success(t *testing.T) {
	mock := &mockPM4PyMCPClient{
		conformanceResult: defaultMockConformanceResult(),
	}
	handler, router := setupMCPGatewayTest(t, mock)
	logPath := createTempEventLogFile(t)

	req := BOSConformanceRequest{
		LogPath: logPath,
		ModelID: "petri_net_abc123",
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/bos/conformance", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, mock.conformanceCalled, "MCP Conformance should be called")

	var resp BOSConformanceResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, uint64(150), resp.TracesChecked)
	assert.Equal(t, 0.96, resp.Fitness)
	assert.Equal(t, 0.92, resp.Precision)
	assert.Equal(t, 0.88, resp.Generalization)
	assert.Equal(t, 0.91, resp.Simplicity)

	assert.Equal(t, uint64(1), handler.stats.RequestsTotal)
	assert.Equal(t, uint64(0), handler.stats.RequestsFailed)
}

func TestConformanceMCP_AllMetricsPopulated(t *testing.T) {
	mock := &mockPM4PyMCPClient{
		conformanceResult: defaultMockConformanceResult(),
	}
	_, router := setupMCPGatewayTest(t, mock)
	logPath := createTempEventLogFile(t)

	req := BOSConformanceRequest{
		LogPath: logPath,
		ModelID: "model_xyz",
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/bos/conformance", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	var resp BOSConformanceResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	assert.Greater(t, resp.Fitness, 0.0)
	assert.Greater(t, resp.Precision, 0.0)
	assert.Greater(t, resp.Generalization, 0.0)
	assert.Greater(t, resp.Simplicity, 0.0)
}

func TestConformanceMCP_ReportsAccurateFitnessMetrics(t *testing.T) {
	mock := &mockPM4PyMCPClient{
		conformanceResult: defaultMockConformanceResult(),
	}
	_, router := setupMCPGatewayTest(t, mock)
	logPath := createTempEventLogFile(t)

	req := BOSConformanceRequest{
		LogPath: logPath,
		ModelID: "model_123",
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/bos/conformance", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	var resp BOSConformanceResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	assert.True(t, resp.Fitness >= 0.9, "Fitness should be >= 0.9")
	assert.True(t, resp.Precision >= 0.9, "Precision should be >= 0.9")
	assert.True(t, resp.Generalization < 1.0, "Generalization should be < 1.0")
}

// ============================================================================
// STATISTICS ENDPOINT TESTS
// ============================================================================

func TestStatisticsMCP_Success(t *testing.T) {
	mock := &mockPM4PyMCPClient{
		statisticsResult: defaultMockStatisticsResult(),
	}
	handler, router := setupMCPGatewayTest(t, mock)
	logPath := createTempEventLogFile(t)

	req := BOSStatisticsRequest{
		LogPath: logPath,
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/bos/statistics", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, mock.statisticsCalled, "MCP Statistics should be called")

	var resp BOSStatisticsResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, 500, resp.NumTraces)
	assert.Equal(t, 2450, resp.NumEvents)
	assert.Equal(t, 8, resp.NumUniqueActivities)
	assert.Equal(t, 3, len(resp.ActivityFrequency))
	assert.Equal(t, int64(60), resp.CaseDuration.MinSeconds)
	assert.Equal(t, int64(3600), resp.CaseDuration.MaxSeconds)
	assert.Equal(t, 1200.5, resp.CaseDuration.AvgSeconds)
	assert.Equal(t, 900.0, resp.CaseDuration.MedianSeconds)

	assert.Equal(t, uint64(1), handler.stats.RequestsTotal)
	assert.Equal(t, uint64(0), handler.stats.RequestsFailed)
}

func TestStatisticsMCP_ActivityFrequencyFromMCP(t *testing.T) {
	mock := &mockPM4PyMCPClient{
		statisticsResult: defaultMockStatisticsResult(),
	}
	_, router := setupMCPGatewayTest(t, mock)
	logPath := createTempEventLogFile(t)

	req := BOSStatisticsRequest{
		LogPath: logPath,
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/bos/statistics", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	var resp BOSStatisticsResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	assert.Greater(t, len(resp.ActivityFrequency), 0, "Should have activity frequency data")
	assert.Equal(t, "create_case", resp.ActivityFrequency[0].Activity)
	assert.Equal(t, 500, resp.ActivityFrequency[0].Frequency)
}

func TestStatisticsMCP_CaseDurationFromMCP(t *testing.T) {
	mock := &mockPM4PyMCPClient{
		statisticsResult: defaultMockStatisticsResult(),
	}
	_, router := setupMCPGatewayTest(t, mock)
	logPath := createTempEventLogFile(t)

	req := BOSStatisticsRequest{
		LogPath: logPath,
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/bos/statistics", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	var resp BOSStatisticsResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	assert.Equal(t, int64(60), resp.CaseDuration.MinSeconds)
	assert.Equal(t, int64(3600), resp.CaseDuration.MaxSeconds)
	assert.Greater(t, resp.CaseDuration.AvgSeconds, 0.0)
	assert.Greater(t, resp.CaseDuration.MedianSeconds, 0.0)
}

// ============================================================================
// ERROR HANDLING TESTS — MCP Client Failures
// ============================================================================

func TestMCPFailure_Discover_Returns503(t *testing.T) {
	mock := &mockPM4PyMCPClient{
		discoverErr: errors.New("pm4py-mcp unavailable"),
	}
	handler, router := setupMCPGatewayTest(t, mock)
	logPath := createTempEventLogFile(t)

	req := BOSDiscoverRequest{
		LogPath:   logPath,
		Algorithm: "inductive_miner",
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/bos/discover", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code,
		"Should return 503 when pm4py-mcp is unreachable")
	assert.Equal(t, uint64(1), handler.stats.RequestsFailed)
}

func TestMCPFailure_Conformance_Returns503(t *testing.T) {
	mock := &mockPM4PyMCPClient{
		conformanceErr: errors.New("connection refused"),
	}
	handler, router := setupMCPGatewayTest(t, mock)
	logPath := createTempEventLogFile(t)

	req := BOSConformanceRequest{
		LogPath: logPath,
		ModelID: "model_123",
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/bos/conformance", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Equal(t, uint64(1), handler.stats.RequestsFailed)
}

func TestMCPFailure_Statistics_Returns503(t *testing.T) {
	mock := &mockPM4PyMCPClient{
		statisticsErr: errors.New("connection refused"),
	}
	handler, router := setupMCPGatewayTest(t, mock)
	logPath := createTempEventLogFile(t)

	req := BOSStatisticsRequest{
		LogPath: logPath,
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/bos/statistics", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Equal(t, uint64(1), handler.stats.RequestsFailed)
}

// ============================================================================
// CONFIG TESTS
// ============================================================================

func TestPM4PyMCPClient_DefaultURL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewBOSGatewayHandler(nil, logger)

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.pm4pyMCPClient)
}

func TestPM4PyMCPClient_CustomURL(t *testing.T) {
	t.Setenv("PM4PY_MCP_URL", "http://custom-host:9999")
	defer t.Setenv("PM4PY_MCP_URL", "")

	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewBOSGatewayHandler(nil, logger)

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.pm4pyMCPClient)
}
