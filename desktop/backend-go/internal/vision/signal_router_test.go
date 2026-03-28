package vision

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// --- fake prober ---

// fakeProber implements HealthProber for deterministic tests.
type fakeProber struct {
	// results maps health URL to (healthy, latencyMs, error).
	results map[string]probeResult
}

type probeResult struct {
	healthy   bool
	latencyMs int64
	err       error
}

func (f *fakeProber) Probe(_ context.Context, url string) (bool, int64, error) {
	r, ok := f.results[url]
	if !ok {
		return false, 0, fmt.Errorf("no fake result for %s", url)
	}
	return r.healthy, r.latencyMs, r.err
}

// allUpProber returns a prober where every service is healthy.
func allUpProber() *fakeProber {
	return &fakeProber{
		results: map[string]probeResult{
			"http://localhost:8090/api/health": {healthy: true, latencyMs: 5},
			"http://localhost:8001/healthz":    {healthy: true, latencyMs: 2},
			"http://localhost:8089/health":     {healthy: true, latencyMs: 8},
			"http://localhost:9089/health":     {healthy: true, latencyMs: 3},
		},
	}
}

// oneDownProber returns a prober where OSA (port 8089) is down.
func oneDownProber() *fakeProber {
	return &fakeProber{
		results: map[string]probeResult{
			"http://localhost:8090/api/health": {healthy: true, latencyMs: 5},
			"http://localhost:8001/healthz":    {healthy: true, latencyMs: 2},
			"http://localhost:8089/health":     {healthy: false, latencyMs: 0, err: fmt.Errorf("connection refused")},
			"http://localhost:9089/health":     {healthy: true, latencyMs: 3},
		},
	}
}

// --- tests ---

// TestVisionStatusAllServicesUp verifies that when all 4 services are healthy,
// all_up is true and every service reports healthy.
func TestVisionStatusAllServicesUp(t *testing.T) {
	sr := NewSignalRouterWithProber(allUpProber(), nil)

	status := sr.ProbeAll(context.Background())

	assert.True(t, status.AllUp, "all_up must be true when all services are healthy")
	assert.Len(t, status.Services, 4, "must report exactly 4 services")
	for _, svc := range status.Services {
		assert.True(t, svc.Healthy, "service %s must be healthy", svc.Name)
		assert.Greater(t, svc.Latency, int64(0), "service %s must have positive latency", svc.Name)
	}
	assert.NotEmpty(t, status.Timestamp)
}

// TestVisionStatusServiceDown verifies that when one service is down,
// all_up is false and the downed service reports healthy=false.
func TestVisionStatusServiceDown(t *testing.T) {
	sr := NewSignalRouterWithProber(oneDownProber(), nil)

	status := sr.ProbeAll(context.Background())

	assert.False(t, status.AllUp, "all_up must be false when any service is down")
	assert.Len(t, status.Services, 4)

	// Verify OSA is reported as down.
	var osaFound bool
	for _, svc := range status.Services {
		if svc.Name == "OSA" {
			osaFound = true
			assert.False(t, svc.Healthy, "OSA must be reported as unhealthy")
		}
	}
	assert.True(t, osaFound, "OSA must be present in services list")

	// Other services should still be healthy.
	healthyCount := 0
	for _, svc := range status.Services {
		if svc.Healthy {
			healthyCount++
		}
	}
	assert.Equal(t, 3, healthyCount, "3 of 4 services should be healthy")
}

// TestSignalEnvelopeCorrect verifies the Signal Theory S=(M,G,T,F,W) envelope
// has the correct fixed values on the response.
func TestSignalEnvelopeCorrect(t *testing.T) {
	sr := NewSignalRouterWithProber(allUpProber(), nil)

	status := sr.ProbeAll(context.Background())

	assert.Equal(t, "data", status.Signal.Mode, "Mode must be 'data'")
	assert.Equal(t, "status", status.Signal.Genre, "Genre must be 'status'")
	assert.Equal(t, "inform", status.Signal.Type, "Type must be 'inform'")
	assert.Equal(t, "json", status.Signal.Format, "Format must be 'json'")
	assert.Equal(t, "vision-status", status.Signal.Structure, "Structure must be 'vision-status'")
}

// TestHealthProbeTimeout verifies that a slow service does not block the handler
// beyond the probe timeout, and that the slow service is reported as unhealthy.
func TestHealthProbeTimeout(t *testing.T) {
	// Start a mock HTTP server that delays responses by 500ms.
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	// Build a service list with only the slow mock.
	slowServices := []serviceSpec{
		{Name: "slow-service", Port: 9999, HealthURL: mockServer.URL},
	}

	sr := NewSignalRouterWithProber(NewHTTPProber(), slowServices)

	// Use a tight context timeout to prove the probe times out.
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	start := time.Now()
	status := sr.ProbeAll(ctx)
	elapsed := time.Since(start)

	// The handler must return well before the mock's 500ms delay.
	assert.Less(t, elapsed, 300*time.Millisecond,
		"ProbeAll blocked for %v; expected to return within 300ms", elapsed)

	assert.Len(t, status.Services, 1)
	assert.False(t, status.Services[0].Healthy, "slow service must be reported as unhealthy on timeout")
	assert.False(t, status.AllUp)
}

// TestVisionStatusHTTPEndpoint verifies the full Gin handler returns valid JSON
// with the correct response shape via GET /api/vision/status.
func TestVisionStatusHTTPEndpoint(t *testing.T) {
	sr := NewSignalRouterWithProber(allUpProber(), nil)

	router := gin.New()
	api := router.Group("/api")
	sr.RegisterRoutes(api)

	req := httptest.NewRequest(http.MethodGet, "/api/vision/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

	var resp VisionStatus
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	assert.True(t, resp.AllUp)
	assert.Len(t, resp.Services, 4)
	assert.Equal(t, "data", resp.Signal.Mode)
	assert.Equal(t, "vision-status", resp.Signal.Structure)
	assert.NotEmpty(t, resp.Timestamp)
}

// TestVisionStatusServiceOrder verifies services are returned in the canonical order:
// pm4py-rust, BusinessOS, OSA, Canopy.
func TestVisionStatusServiceOrder(t *testing.T) {
	sr := NewSignalRouterWithProber(allUpProber(), nil)

	status := sr.ProbeAll(context.Background())

	expected := []string{"pm4py-rust", "BusinessOS", "OSA", "Canopy"}
	actual := make([]string, len(status.Services))
	for i, svc := range status.Services {
		actual[i] = svc.Name
	}
	assert.Equal(t, expected, actual, "services must be in canonical order")
}

// TestVisionStatusPortMapping verifies each service reports the correct port.
func TestVisionStatusPortMapping(t *testing.T) {
	sr := NewSignalRouterWithProber(allUpProber(), nil)

	status := sr.ProbeAll(context.Background())

	expectedPorts := map[string]int{
		"pm4py-rust": 8090,
		"BusinessOS": 8001,
		"OSA":        8089,
		"Canopy":     9089,
	}

	for _, svc := range status.Services {
		assert.Equal(t, expectedPorts[svc.Name], svc.Port,
			"service %s must report port %d", svc.Name, expectedPorts[svc.Name])
	}
}
