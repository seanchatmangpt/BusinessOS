package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/services"
	"github.com/stretchr/testify/assert"
)

// mockDashboardMCPClient implements services.PM4PyMCPClientInterface for tests.
type mockDashboardMCPClient struct {
	statsResult *services.StatisticsResult
	statsErr    error
	confResult  *services.ConformanceResult
	confErr     error
}

func (m *mockDashboardMCPClient) Discover(_ context.Context, _ json.RawMessage, _ string) (*services.DiscoverResult, error) {
	return nil, nil
}

func (m *mockDashboardMCPClient) Conformance(_ context.Context, _, _ json.RawMessage) (*services.ConformanceResult, error) {
	return m.confResult, m.confErr
}

func (m *mockDashboardMCPClient) Statistics(_ context.Context, _ json.RawMessage) (*services.StatisticsResult, error) {
	return m.statsResult, m.statsErr
}

func (m *mockDashboardMCPClient) HealthCheck(_ context.Context) error {
	return nil
}

// TestPM4PyDashboardKPI_MissingEventLog verifies 400 is returned when event_log is absent.
func TestPM4PyDashboardKPI_MissingEventLog(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewPM4PyDashboardHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/api/pm4py/dashboard-kpi", bytes.NewBufferString("{}"))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.GetDashboardKPI(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var body map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.Contains(t, body, "error")
}

// TestPM4PyDashboardKPI_ReturnsKPIShape verifies that a well-formed request returns
// 200 with a ProcessMiningKPIResponse shape using the MCP client.
func TestPM4PyDashboardKPI_ReturnsKPIShape(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mock := &mockDashboardMCPClient{
		statsResult: &services.StatisticsResult{
			TraceEvents:      100,
			UniqueTraces:     10,
			UniqueActivities: 5,
			ActivityFrequency: []services.ActivityFrequency{
				{Activity: "A", Frequency: 30, Percentage: 30.0},
			},
			Variants: []services.Variant{
				{Variant: "A->B->C", Count: 6, Percentage: 60.0},
			},
			Bottleneck: []services.BottleneckActivity{
				{Activity: "A", MeanMs: 500, MedianMs: 400},
			},
		},
		confResult: &services.ConformanceResult{
			Fitness:   0.95,
			Precision: 0.88,
		},
	}
	handler := NewPM4PyDashboardHandler(mock)

	body, err := json.Marshal(map[string]interface{}{
		"event_log": map[string]interface{}{"traces": []interface{}{}},
		"petri_net": map[string]interface{}{
			"places": []interface{}{}, "transitions": []interface{}{}, "arcs": []interface{}{},
		},
	})
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/api/pm4py/dashboard-kpi", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.GetDashboardKPI(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp ProcessMiningKPIResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.FetchedAt)
	assert.Equal(t, 100, resp.EventCount)
	assert.Equal(t, 10, resp.TraceCount)
	assert.Len(t, resp.TopVariants, 1)
	assert.Len(t, resp.BottleneckActivities, 1)
	assert.InDelta(t, 0.95, resp.ConformanceFitness, 0.001)
}

// TestPM4PyDashboardKPI_WithPetriNet verifies that supplying a petri_net does not crash.
func TestPM4PyDashboardKPI_WithPetriNet(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mock := &mockDashboardMCPClient{
		statsResult: &services.StatisticsResult{
			TraceEvents:  50,
			UniqueTraces: 5,
		},
		confResult: &services.ConformanceResult{Fitness: 1.0, Precision: 0.9},
	}
	handler := NewPM4PyDashboardHandler(mock)

	body, err := json.Marshal(map[string]interface{}{
		"event_log": map[string]interface{}{"traces": []interface{}{}},
		"petri_net": map[string]interface{}{
			"places": []interface{}{}, "transitions": []interface{}{}, "arcs": []interface{}{},
		},
	})
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/api/pm4py/dashboard-kpi", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.GetDashboardKPI(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp ProcessMiningKPIResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.FetchedAt)
}

// TestPM4PyDashboardKPI_BuildResponse_MapsStatisticsFields verifies buildResponse correctly
// maps StatisticsResult fields into ProcessMiningKPIResponse.
func TestPM4PyDashboardKPI_BuildResponse_MapsStatisticsFields(t *testing.T) {
	handler := NewPM4PyDashboardHandler(nil)

	stats := &services.StatisticsResult{
		TraceEvents: 100,
		UniqueTraces: 10,
		ActivityFrequency: []services.ActivityFrequency{
			{Activity: "A", Frequency: 5, Percentage: 50.0},
			{Activity: "B", Frequency: 3, Percentage: 30.0},
		},
		Variants: []services.Variant{
			{Variant: "v1", Count: 6, Percentage: 60.0},
			{Variant: "v2", Count: 4, Percentage: 40.0},
		},
		Bottleneck: []services.BottleneckActivity{
			{Activity: "A", MeanMs: 500, MedianMs: 400},
		},
	}

	resp := handler.buildResponse(stats, nil)

	assert.Equal(t, 100, resp.EventCount)
	assert.Equal(t, 10, resp.TraceCount)
	assert.Equal(t, 2, resp.VariantCount)
	assert.Equal(t, 5, resp.ActivityFrequencies["A"])
	assert.Len(t, resp.BottleneckActivities, 1)
	assert.Equal(t, "A", resp.BottleneckActivities[0].Activity)
	assert.Len(t, resp.TopVariants, 2)
	assert.NotEmpty(t, resp.FetchedAt)
}

// TestPM4PyDashboardKPI_BuildResponse_MapsConformanceFields verifies buildResponse
// correctly maps ConformanceResult into ProcessMiningKPIResponse.
func TestPM4PyDashboardKPI_BuildResponse_MapsConformanceFields(t *testing.T) {
	handler := NewPM4PyDashboardHandler(nil)

	conf := &services.ConformanceResult{
		Fitness:   0.95,
		Precision: 0.88,
	}

	resp := handler.buildResponse(nil, conf)

	assert.InDelta(t, 0.95, resp.ConformanceFitness, 0.001)
	assert.InDelta(t, 0.88, resp.ConformancePrecision, 0.001)
	assert.False(t, resp.IsConformant) // Fitness < 1.0
}

// TestPM4PyDashboardKPI_BuildResponse_PerfectFitnessIsConformant verifies fitness >= 1.0 maps to true.
func TestPM4PyDashboardKPI_BuildResponse_PerfectFitnessIsConformant(t *testing.T) {
	handler := NewPM4PyDashboardHandler(nil)

	conf := &services.ConformanceResult{Fitness: 1.0, Precision: 0.90}
	resp := handler.buildResponse(nil, conf)
	assert.True(t, resp.IsConformant)
}

// TestPM4PyDashboardKPI_BuildResponse_NilInputsReturnEmptySlices verifies that nil
// stats and conf produce a valid (non-nil collections) response.
func TestPM4PyDashboardKPI_BuildResponse_NilInputsReturnEmptySlices(t *testing.T) {
	handler := NewPM4PyDashboardHandler(nil)

	resp := handler.buildResponse(nil, nil)

	assert.NotNil(t, resp.TopVariants)
	assert.NotNil(t, resp.BottleneckActivities)
	assert.NotNil(t, resp.ActivityFrequencies)
	assert.NotEmpty(t, resp.FetchedAt)
	assert.False(t, resp.IsConformant)
	assert.Equal(t, 0.0, resp.ConformanceFitness)
}

// TestPM4PyDashboardKPI_StatsError_Returns502 verifies statistics failure surfaces as 502.
func TestPM4PyDashboardKPI_StatsError_Returns502(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mock := &mockDashboardMCPClient{
		statsErr: assert.AnError,
	}
	handler := NewPM4PyDashboardHandler(mock)

	body, _ := json.Marshal(map[string]interface{}{
		"event_log": map[string]interface{}{"traces": []interface{}{}},
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/api/pm4py/dashboard-kpi", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.GetDashboardKPI(c)

	assert.Equal(t, http.StatusBadGateway, w.Code)
	assert.Contains(t, w.Body.String(), "error")
}
