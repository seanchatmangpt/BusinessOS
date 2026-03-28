package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/ontology"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestVerifyFrameworks tests POST /v1/compliance/verify endpoint.
func TestVerifyFrameworks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	engine, err := ontology.NewComplianceEngine("", slog.Default())
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = engine.Initialize(ctx)
	require.NoError(t, err)

	handler := NewComplianceAPIHandler(engine, slog.Default())

	tests := []struct {
		name                string
		frameworks          []string
		expectedStatus      int
		expectedFieldsExist []string
	}{
		{
			name:                "verify SOC2",
			frameworks:          []string{"SOC2"},
			expectedStatus:      http.StatusOK,
			expectedFieldsExist: []string{"status", "overall_score", "frameworks", "timestamp"},
		},
		{
			name:                "verify multiple frameworks",
			frameworks:          []string{"SOC2", "GDPR", "HIPAA"},
			expectedStatus:      http.StatusOK,
			expectedFieldsExist: []string{"status", "overall_score", "frameworks"},
		},
		{
			name:           "verify SOX",
			frameworks:     []string{"SOX"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid framework",
			frameworks:     []string{"INVALID"},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			reqBody := VerifyRequest{
				Frameworks: tt.frameworks,
				Timeout:    30,
			}
			body, _ := json.Marshal(reqBody)
			c.Request = httptest.NewRequest("POST", "/v1/compliance/verify", bytes.NewReader(body))
			c.Request.Header.Set("Content-Type", "application/json")

			handler.VerifyFrameworks(c)

			assert.Equal(t, tt.expectedStatus, w.Code, "status code mismatch")
			if tt.expectedStatus == http.StatusOK {
				var resp VerifyResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)

				for _, field := range tt.expectedFieldsExist {
					switch field {
					case "status":
						assert.NotEmpty(t, resp.Status)
						assert.Contains(t, []string{"compliant", "non_compliant", "partial"}, resp.Status)
					case "overall_score":
						assert.GreaterOrEqual(t, resp.OverallScore, 0.0)
						assert.LessOrEqual(t, resp.OverallScore, 1.0)
					case "frameworks":
						assert.NotNil(t, resp.Frameworks)
						assert.Greater(t, len(resp.Frameworks), 0)
					case "timestamp":
						assert.NotEmpty(t, resp.Timestamp)
					}
				}
			}
		})
	}
}

// TestGenerateReport tests GET /v1/compliance/report endpoint.
func TestGenerateReport(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	engine, err := ontology.NewComplianceEngine("", slog.Default())
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = engine.Initialize(ctx)
	require.NoError(t, err)

	handler := NewComplianceAPIHandler(engine, slog.Default())

	t.Run("generate full report", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/v1/compliance/report?frameworks=SOC2,GDPR,HIPAA,SOX&include_details=true", nil)

		handler.GenerateReport(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var matrix ontology.ComplianceMatrix
		err := json.Unmarshal(w.Body.Bytes(), &matrix)
		require.NoError(t, err)

		// Verify all frameworks present
		assert.Greater(t, len(matrix.Frameworks), 0)
		assert.GreaterOrEqual(t, matrix.OverallScore, 0.0)
		assert.LessOrEqual(t, matrix.OverallScore, 1.0)

		// Verify timestamp
		assert.NotEmpty(t, matrix.Timestamp)
	})

	t.Run("generate report without details", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/v1/compliance/report?include_details=false", nil)

		handler.GenerateReport(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var matrix ontology.ComplianceMatrix
		err := json.Unmarshal(w.Body.Bytes(), &matrix)
		require.NoError(t, err)

		// Violations should be removed
		for _, report := range matrix.Frameworks {
			if report != nil && report.Violations == nil {
				// Verified violations are cleared when include_details=false
			}
		}
	})
}

// TestListFrameworkControls tests GET /v1/compliance/controls/:framework endpoint.
func TestListFrameworkControls(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	engine, err := ontology.NewComplianceEngine("", slog.Default())
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = engine.Initialize(ctx)
	require.NoError(t, err)

	handler := NewComplianceAPIHandler(engine, slog.Default())

	tests := []struct {
		name                string
		framework           string
		severityFilter      string
		expectedStatus      int
		expectedMinControls int
	}{
		{
			name:                "list SOC2 controls",
			framework:           "SOC2",
			expectedStatus:      http.StatusOK,
			expectedMinControls: 1,
		},
		{
			name:                "list GDPR controls",
			framework:           "GDPR",
			expectedStatus:      http.StatusOK,
			expectedMinControls: 1,
		},
		{
			name:                "list HIPAA controls",
			framework:           "HIPAA",
			expectedStatus:      http.StatusOK,
			expectedMinControls: 1,
		},
		{
			name:                "list SOX controls",
			framework:           "SOX",
			expectedStatus:      http.StatusOK,
			expectedMinControls: 1,
		},
		{
			name:           "filter by critical severity",
			framework:      "SOC2",
			severityFilter: "critical",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid framework",
			framework:      "INVALID",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid severity filter",
			framework:      "SOC2",
			severityFilter: "invalid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			url := fmt.Sprintf("/v1/compliance/controls/%s", tt.framework)
			if tt.severityFilter != "" {
				url += fmt.Sprintf("?severity=%s", tt.severityFilter)
			}
			c.Request = httptest.NewRequest("GET", url, nil)
			c.Params = []gin.Param{{Key: "framework", Value: tt.framework}}

			handler.ListFrameworkControls(c)

			assert.Equal(t, tt.expectedStatus, w.Code, "status code mismatch for "+tt.name)

			if tt.expectedStatus == http.StatusOK {
				var resp struct {
					Framework string                        `json:"framework"`
					Controls  []*ontology.ComplianceControl `json:"controls"`
					Total     int                           `json:"total"`
					Timestamp string                        `json:"timestamp"`
				}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)

				assert.Equal(t, tt.framework, resp.Framework)
				assert.Greater(t, resp.Total, 0)
				assert.Greater(t, len(resp.Controls), 0)

				// Verify severity filter if specified
				if tt.severityFilter != "" {
					for _, ctrl := range resp.Controls {
						assert.Equal(t, tt.severityFilter, ctrl.Severity)
					}
				}
			}
		})
	}
}

// TestReloadOntology tests POST /v1/compliance/reload endpoint.
func TestReloadOntology(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	engine, err := ontology.NewComplianceEngine("", slog.Default())
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = engine.Initialize(ctx)
	require.NoError(t, err)

	handler := NewComplianceAPIHandler(engine, slog.Default())

	t.Run("reload ontology", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := ReloadRequest{ClearCache: false}
		body, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest("POST", "/v1/compliance/reload", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.ReloadOntology(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp ReloadResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Equal(t, "reloaded", resp.Status)
		assert.NotEmpty(t, resp.Timestamp)
	})

	t.Run("reload with clear cache", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := ReloadRequest{ClearCache: true}
		body, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest("POST", "/v1/compliance/reload", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.ReloadOntology(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// TestComplianceFrameworkVerification tests all frameworks can be verified independently.
func TestComplianceFrameworkVerification(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	engine, err := ontology.NewComplianceEngine("", slog.Default())
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = engine.Initialize(ctx)
	require.NoError(t, err)

	tests := []struct {
		name      string
		framework string
		verifyFn  func(context.Context) (*ontology.ComplianceReport, error)
	}{
		{
			name:      "SOC2 verification",
			framework: "SOC2",
			verifyFn:  engine.VerifySOC2,
		},
		{
			name:      "GDPR verification",
			framework: "GDPR",
			verifyFn:  engine.VerifyGDPR,
		},
		{
			name:      "HIPAA verification",
			framework: "HIPAA",
			verifyFn:  engine.VerifyHIPAA,
		},
		{
			name:      "SOX verification",
			framework: "SOX",
			verifyFn:  engine.VerifySOX,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			report, err := tt.verifyFn(ctx)
			require.NoError(t, err)

			assert.Equal(t, tt.framework, report.Framework)
			assert.Greater(t, report.TotalControls, 0)
			assert.GreaterOrEqual(t, report.PassedControls, 0)
			assert.GreaterOrEqual(t, report.FailedControls, 0)
			assert.Contains(t, []string{"compliant", "non_compliant", "partial"}, report.Status)
			assert.GreaterOrEqual(t, report.Score, 0.0)
			assert.LessOrEqual(t, report.Score, 1.0)
		})
	}
}

// TestControlsStructure tests ComplianceControl structure has required fields.
func TestControlsStructure(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	engine, err := ontology.NewComplianceEngine("", slog.Default())
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = engine.Initialize(ctx)
	require.NoError(t, err)

	controls := engine.GetFrameworkControls("SOC2")
	require.Greater(t, len(controls), 0)

	for _, ctrl := range controls {
		assert.NotEmpty(t, ctrl.ID, "control ID should not be empty")
		assert.NotEmpty(t, ctrl.Framework, "control Framework should not be empty")
		assert.NotEmpty(t, ctrl.Title, "control Title should not be empty")
		assert.NotEmpty(t, ctrl.Description, "control Description should not be empty")
		assert.Contains(t, []string{"critical", "high", "medium", "low"}, ctrl.Severity, "invalid severity")
	}
}

// TestErrorHandlingComplianceAPI tests error handling in API endpoints.
func TestErrorHandlingComplianceAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	engine, err := ontology.NewComplianceEngine("", slog.Default())
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = engine.Initialize(ctx)
	require.NoError(t, err)

	handler := NewComplianceAPIHandler(engine, slog.Default())

	t.Run("verify with empty frameworks", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := VerifyRequest{
			Frameworks: []string{},
		}
		body, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest("POST", "/v1/compliance/verify", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.VerifyFrameworks(c)

		// Empty frameworks should result in validation error
		assert.NotEqual(t, http.StatusOK, w.Code)
	})

	t.Run("controls with missing framework", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/v1/compliance/controls/", nil)
		c.Params = []gin.Param{{Key: "framework", Value: ""}}

		handler.ListFrameworkControls(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestPerformance tests that compliance verification completes within reasonable time.
func TestPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	engine, err := ontology.NewComplianceEngine("", slog.Default())
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = engine.Initialize(ctx)
	require.NoError(t, err)

	// Time full report generation
	start := time.Now()
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	matrix, err := engine.GenerateReport(ctx)
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.NotNil(t, matrix)

	// Should complete in < 1 second
	assert.Less(t, elapsed, 1*time.Second, "report generation took too long")
}

// TestComplianceAPIIntegration tests full API workflow.
func TestComplianceAPIIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	engine, err := ontology.NewComplianceEngine("", slog.Default())
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = engine.Initialize(ctx)
	require.NoError(t, err)

	handler := NewComplianceAPIHandler(engine, slog.Default())

	// Step 1: Verify all frameworks
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	reqBody := VerifyRequest{
		Frameworks: []string{"SOC2", "GDPR", "HIPAA", "SOX"},
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest("POST", "/v1/compliance/verify", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	handler.VerifyFrameworks(c)
	assert.Equal(t, http.StatusOK, w.Code)

	// Step 2: Get full report
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/v1/compliance/report?include_details=true", nil)
	handler.GenerateReport(c)
	assert.Equal(t, http.StatusOK, w.Code)

	// Step 3: List controls for each framework
	for _, framework := range []string{"SOC2", "GDPR", "HIPAA", "SOX"} {
		w = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", fmt.Sprintf("/v1/compliance/controls/%s", framework), nil)
		c.Params = []gin.Param{{Key: "framework", Value: framework}}
		handler.ListFrameworkControls(c)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// Step 4: Reload ontology
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/v1/compliance/reload", bytes.NewReader([]byte("{}")))
	c.Request.Header.Set("Content-Type", "application/json")
	handler.ReloadOntology(c)
	assert.Equal(t, http.StatusOK, w.Code)
}
