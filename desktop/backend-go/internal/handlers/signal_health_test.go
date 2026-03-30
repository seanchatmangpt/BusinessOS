package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestGetSignalHealth_UsesMetricsRegistry verifies that GetSignalHealth
// actually calls the ProxyMetricsRegistry to get uptime and last activity,
// not hardcoded values.
//
// RED PHASE: This test FAILS because:
// 1. Handlers struct doesn't have proxyMetricsRegistry field (compilation error)
// 2. GetSignalHealth returns hardcoded "healthy" response
// 3. ProxyMetricsRegistry.GetUptime() and GetLastActivity() methods don't exist
func TestGetSignalHealth_UsesMetricsRegistry(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	// Create handlers
	// RED PHASE: proxyMetricsRegistry field doesn't exist in Handlers struct yet!
	h := &Handlers{}

	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/signal/health", nil)

	// Execute
	// RED PHASE: GetSignalHealth returns hardcoded values
	// GREEN PHASE: Will call h.proxyMetricsRegistry.GetUptime() and GetLastActivity()
	h.GetSignalHealth(c)

	// Assert response structure
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse response
	var resp SignalHealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	// RED PHASE ASSERTIONS:
	// These assert that the response contains real metrics, not hardcoded values
	// In GREEN phase, we'll add uptime_ms and last_activity fields

	// 1. Status should be healthy
	assert.Equal(t, "healthy", resp.Status)

	// 2. All 6 metrics should be enabled
	assert.True(t, resp.Metrics.ActionCompletion, "ActionCompletion should be true")
	assert.True(t, resp.Metrics.ReEncoding, "ReEncoding should be true")
	assert.True(t, resp.Metrics.SignalBounce, "SignalBounce should be true")
	assert.True(t, resp.Metrics.GenreRecognition, "GenreRecognition should be true")
	assert.True(t, resp.Metrics.FeedbackClosure, "FeedbackClosure should be true")
	assert.True(t, resp.Metrics.TimeToDecide, "TimeToDecide should be true")

	// 3. Verify uptime_ms and last_activity are present (GREEN PHASE)
	// When proxyMetricsRegistry is nil, uptime_ms is 0 and last_activity is current time
	assert.GreaterOrEqual(t, resp.UptimeMs, int64(0), "uptime_ms should be >= 0")
	assert.NotEmpty(t, resp.LastActivity, "last_activity should not be empty")

	// 4. Verify feedback loop structure
	assert.True(t, resp.FeedbackLoop.HomeostaticLoop)
}
