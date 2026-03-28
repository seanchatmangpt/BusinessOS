package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newYawlSimulateRouter builds a test router with only the simulate endpoint.
// osaSrvURL is injected via OSA_URL env var.
func newYawlSimulateRouter(t *testing.T, osaSrvURL string) *gin.Engine {
	t.Helper()
	t.Setenv("OSA_URL", osaSrvURL)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewYawlHandler(nil)
	api := r.Group("/api")
	api.POST("/yawl/simulate", h.SimulateWorkflows)
	return r
}

// TestSimulateWorkflows_OSAReturnsResult_Returns200 tests that when OSA returns
// a valid SimulationResult, SimulateWorkflows returns 200 with that payload.
func TestSimulateWorkflows_OSAReturnsResult_Returns200(t *testing.T) {
	// Fake OSA server that returns a canned simulation result.
	osaSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/yawl/simulate", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"spec_set": "basic_wcp",
			"user_count": 3,
			"total_duration_ms": 500,
			"completed_count": 3,
			"error_count": 0,
			"timeout_count": 0,
			"summary": "spec_set=basic_wcp users=3 completed=3 errors=0 timeouts=0",
			"results": []
		}`))
	}))
	defer osaSrv.Close()

	r := newYawlSimulateRouter(t, osaSrv.URL)

	body := bytes.NewBufferString(`{"spec_set":"basic_wcp","user_count":3}`)
	req := httptest.NewRequest(http.MethodPost, "/api/yawl/simulate", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp yawlSimulateResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "basic_wcp", resp.SpecSet)
	assert.Equal(t, 3, resp.UserCount)
	assert.Equal(t, 3, resp.CompletedCount)
	assert.Equal(t, 0, resp.ErrorCount)
	assert.NotEmpty(t, resp.Summary)
}

// TestSimulateWorkflows_EmptyBody_UsesDefaults tests that missing fields
// are filled with defaults before forwarding to OSA.
func TestSimulateWorkflows_EmptyBody_UsesDefaults(t *testing.T) {
	var capturedBody map[string]any

	osaSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&capturedBody)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"spec_set":"basic_wcp","user_count":3,"total_duration_ms":0,
			"completed_count":0,"error_count":0,"timeout_count":0,
			"summary":"empty","results":[]
		}`))
	}))
	defer osaSrv.Close()

	r := newYawlSimulateRouter(t, osaSrv.URL)

	// Send completely empty body — handler must apply defaults.
	req := httptest.NewRequest(http.MethodPost, "/api/yawl/simulate",
		bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify defaults were forwarded to OSA.
	assert.Equal(t, "basic_wcp", capturedBody["spec_set"])
	assert.EqualValues(t, 3, capturedBody["user_count"])
	assert.EqualValues(t, 30_000, capturedBody["timeout_ms"])
	assert.EqualValues(t, 50, capturedBody["max_steps"])
	assert.EqualValues(t, 10, capturedBody["max_concurrency"])
}

// TestSimulateWorkflows_OSAUnreachable_Returns502 tests that when OSA is
// unreachable, SimulateWorkflows returns 502 Bad Gateway.
func TestSimulateWorkflows_OSAUnreachable_Returns502(t *testing.T) {
	// Port 19997 — nothing listening.
	r := newYawlSimulateRouter(t, "http://localhost:19997")

	body := bytes.NewBufferString(`{"spec_set":"basic_wcp","user_count":1}`)
	req := httptest.NewRequest(http.MethodPost, "/api/yawl/simulate", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadGateway, w.Code)

	var resp gin.H
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.NotEmpty(t, resp["error"])
}

// TestSimulateWorkflows_OSAReturns500_Returns502 tests that a non-200 from
// OSA is surfaced as 502 (not 500) so the caller knows it's a gateway error.
func TestSimulateWorkflows_OSAReturns500_Returns502(t *testing.T) {
	osaSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal OSA error"}`))
	}))
	defer osaSrv.Close()

	r := newYawlSimulateRouter(t, osaSrv.URL)

	body := bytes.NewBufferString(`{"spec_set":"wcp_patterns","user_count":2}`)
	req := httptest.NewRequest(http.MethodPost, "/api/yawl/simulate", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadGateway, w.Code)
}

// TestSimulateWorkflows_AllSpecSets_ForwardedCorrectly tests that spec_set
// values are forwarded verbatim to OSA.
func TestSimulateWorkflows_AllSpecSets_ForwardedCorrectly(t *testing.T) {
	specSets := []string{"basic_wcp", "wcp_patterns", "real_data", "all"}

	for _, specSet := range specSets {
		t.Run(specSet, func(t *testing.T) {
			var got string

			osaSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var b map[string]any
				json.NewDecoder(r.Body).Decode(&b)
				got = b["spec_set"].(string)

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"spec_set":"` + specSet + `","user_count":1,"total_duration_ms":0,"completed_count":0,"error_count":0,"timeout_count":0,"summary":"ok","results":[]}`))
			}))
			defer osaSrv.Close()

			r := newYawlSimulateRouter(t, osaSrv.URL)
			body := bytes.NewBufferString(`{"spec_set":"` + specSet + `","user_count":1}`)
			req := httptest.NewRequest(http.MethodPost, "/api/yawl/simulate", body)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, specSet, got)
		})
	}
}
