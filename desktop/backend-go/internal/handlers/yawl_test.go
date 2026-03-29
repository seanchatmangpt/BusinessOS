package handlers

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newYawlTestRouter constructs a test gin.Engine with YAWL routes,
// pointed at the specified yawlURL via env var.
func newYawlTestRouter(t *testing.T, yawlURL string) *gin.Engine {
	t.Helper()
	t.Setenv("YAWLV6_URL", yawlURL)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewYawlHandler(slog.Default())
	api := r.Group("/api")
	api.GET("/yawl/health", h.GetHealth)
	api.POST("/yawl/conformance", h.CheckConformance)
	api.POST("/yawl/spec", h.BuildSpec)
	api.GET("/yawl/spec/load", h.LoadSpec)
	return r
}

// TestYawlHandler_GetHealth_Unreachable_Returns502 tests that health check
// returns 502 Bad Gateway when the YAWL engine is unreachable.
func TestYawlHandler_GetHealth_Unreachable_Returns502(t *testing.T) {
	// Port 19998 is not bound — connection will be refused immediately.
	r := newYawlTestRouter(t, "http://localhost:19998")

	req := httptest.NewRequest(http.MethodGet, "/api/yawl/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadGateway, w.Code)

	var body gin.H
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.NotEmpty(t, body["error"])
}

// TestYawlHandler_CheckConformance_MissingFields_Returns400 tests that
// conformance check returns 400 when required fields are missing.
func TestYawlHandler_CheckConformance_MissingFields_Returns400(t *testing.T) {
	r := newYawlTestRouter(t, "http://localhost:19998")

	req := httptest.NewRequest(http.MethodPost, "/api/yawl/conformance", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var body gin.H
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.NotEmpty(t, body["error"])
}

// TestYawlHandler_CheckConformance_ValidRequest_Proxies tests that a valid
// conformance request is proxied to the YAWL engine and the result is returned.
func TestYawlHandler_CheckConformance_ValidRequest_Proxies(t *testing.T) {
	// Stand up a mock YAWL conformance endpoint.
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/process-mining/conformance" && r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"fitness":    0.92,
				"violations": []string{},
				"is_sound":   true,
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer mockServer.Close()

	r := newYawlTestRouter(t, mockServer.URL)

	body := map[string]interface{}{
		"spec_xml":  "<specificationSet/>",
		"event_log": json.RawMessage(`[{"case_id":"1","activity":"A"}]`),
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/yawl/conformance", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
	assert.InDelta(t, 0.92, result["fitness"].(float64), 0.001, "fitness field must be present and proxied")
	assert.True(t, result["is_sound"].(bool))
}

// TestYawlHandler_BuildSpec_SequenceType tests that a sequence spec
// is built correctly with the required task names.
func TestYawlHandler_BuildSpec_SequenceType(t *testing.T) {
	r := newYawlTestRouter(t, "http://localhost:19998")

	body := map[string]interface{}{
		"type":  "sequence",
		"tasks": []string{"A", "B"},
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/yawl/spec", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/xml")

	xml := w.Body.String()
	assert.Contains(t, xml, `id="A"`)
	assert.Contains(t, xml, `id="B"`)
	assert.Contains(t, xml, "specificationSet")
}

// TestYawlHandler_BuildSpec_ParallelType tests that a parallel spec
// is built correctly with AND-split gateway.
func TestYawlHandler_BuildSpec_ParallelType(t *testing.T) {
	r := newYawlTestRouter(t, "http://localhost:19998")

	body := map[string]interface{}{
		"type":     "parallel",
		"trigger":  "Start",
		"branches": []string{"A", "B"},
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/yawl/spec", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	xml := w.Body.String()
	assert.Contains(t, xml, `split code="and"`, "parallel split must produce an AND-split gateway")
	assert.Contains(t, xml, `id="A"`)
	assert.Contains(t, xml, `id="B"`)
	assert.Contains(t, xml, "OSA_ParallelSplit")
}

// TestYawlHandler_LoadSpec_MissingPatternID_Returns400 tests that
// LoadSpec returns 400 when the pattern_id query parameter is missing.
func TestYawlHandler_LoadSpec_MissingPatternID_Returns400(t *testing.T) {
	r := newYawlTestRouter(t, "http://localhost:19998")

	req := httptest.NewRequest(http.MethodGet, "/api/yawl/spec/load", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var body gin.H
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.NotEmpty(t, body["error"])
}
