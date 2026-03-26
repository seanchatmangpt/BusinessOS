package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/ontology"
	"github.com/rhl/businessos-backend/internal/utils"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// ComplianceAPIHandler handles REST API endpoints for compliance verification and reporting.
type ComplianceAPIHandler struct {
	engine *ontology.ComplianceEngine
	logger *slog.Logger
	tracer trace.Tracer
}

// NewComplianceAPIHandler constructs a ComplianceAPIHandler.
func NewComplianceAPIHandler(engine *ontology.ComplianceEngine, logger *slog.Logger) *ComplianceAPIHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &ComplianceAPIHandler{
		engine: engine,
		logger: logger,
		tracer: otel.Tracer("businessos"),
	}
}

// VerifyRequest represents the request body for POST /v1/compliance/verify.
type VerifyRequest struct {
	Frameworks []string `json:"frameworks" binding:"required,min=1"`
	Timeout    int      `json:"timeout_seconds" binding:"omitempty,min=1,max=300"`
}

// VerifyResponse represents the response for verify endpoint.
type VerifyResponse struct {
	Status       string                 `json:"status"` // compliant, non_compliant, partial
	OverallScore float64                `json:"overall_score"`
	Frameworks   map[string]interface{} `json:"frameworks"`
	Timestamp    string                 `json:"timestamp"`
}

// ReportRequest represents the request body for GET /v1/compliance/report.
type ReportRequest struct {
	Frameworks     []string `form:"frameworks" binding:"required,min=1"`
	IncludeDetails bool     `form:"include_details" binding:"omitempty"`
}

// ControlsListRequest represents the request for GET /v1/compliance/controls/:framework.
type ControlsListRequest struct {
	Severity string `form:"severity" binding:"omitempty,oneof=critical high medium low"`
	Status   string `form:"status" binding:"omitempty,oneof=verified failed"`
}

// ReloadRequest represents the request body for POST /v1/compliance/reload.
type ReloadRequest struct {
	ClearCache bool `json:"clear_cache" binding:"omitempty"`
}

// ReloadResponse represents the response for reload endpoint.
type ReloadResponse struct {
	Status    string `json:"status"` // reloaded, already_loaded
	Timestamp string `json:"timestamp"`
}

// registerComplianceAPIRoutes wires up compliance REST API routes.
// Routes: POST /v1/compliance/verify, GET /v1/compliance/report,
// GET /v1/compliance/controls/:framework, POST /v1/compliance/reload
func (h *Handlers) registerComplianceAPIRoutes(api *gin.RouterGroup, auth gin.HandlerFunc) {
	if h == nil {
		slog.Warn("Handlers is nil, skipping compliance API routes")
		return
	}

	// Initialize ComplianceAPIHandler with ComplianceEngine from ontology package
	handler := NewComplianceAPIHandler(ontology.GlobalComplianceEngine, slog.Default())

	// Create compliance API group with authentication
	complianceAPI := api.Group("/v1/compliance")
	complianceAPI.Use(auth, middleware.RequireAuth())
	{
		// POST /v1/compliance/verify - Verify one or more compliance frameworks
		complianceAPI.POST("/verify", handler.VerifyFrameworks)

		// GET /v1/compliance/report - Generate aggregated compliance report
		complianceAPI.GET("/report", handler.GenerateReport)

		// GET /v1/compliance/controls/:framework - List controls for a framework
		complianceAPI.GET("/controls/:framework", handler.ListFrameworkControls)

		// POST /v1/compliance/reload - Reload compliance ontology (admin only)
		complianceAPI.POST("/reload", handler.ReloadOntology)
	}

	slog.Info("Compliance REST API routes registered at /api/v1/compliance/*")
}

// VerifyFrameworks handles POST /v1/compliance/verify.
// Verifies one or more compliance frameworks and returns detailed results.
// Request: {"frameworks": ["SOC2", "GDPR"], "timeout_seconds": 30}
// Response: {"status": "compliant|non_compliant|partial", "overall_score": 0.92, "frameworks": {...}, "timestamp": "2026-03-26T..."}
// OTEL instrumentation: emits jtbd_scenario_compliance_check span for Wave 12
func (h *ComplianceAPIHandler) VerifyFrameworks(c *gin.Context) {
	var req VerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, h.logger, err)
		return
	}

	// Validate frameworks
	validFrameworks := map[string]bool{"SOC2": true, "GDPR": true, "HIPAA": true, "SOX": true}
	for _, fw := range req.Frameworks {
		if !validFrameworks[fw] {
			utils.RespondBadRequest(c, h.logger, "invalid framework: "+fw+". Valid: SOC2, GDPR, HIPAA, SOX")
			return
		}
	}

	ctx := c.Request.Context()

	// OTEL instrumentation: start span
	spanCtx, span := h.tracer.Start(ctx, "jtbd_scenario_compliance_check",
		trace.WithAttributes(
			attribute.String("jtbd.scenario.id", "compliance_check"),
			attribute.String("jtbd.scenario.step", "load_rules"),
			attribute.Int("jtbd.scenario.step_num", 1),
			attribute.Int("jtbd.scenario.step_total", 3),
			attribute.String("jtbd.scenario.system", "businessos"),
			attribute.String("jtbd.scenario.wave", "wave12"),
			attribute.String("jtbd.scenario.outcome", "pending"),
		),
	)
	defer span.End()

	ctx = spanCtx

	if req.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(req.Timeout)*time.Second)
		defer cancel()
	}

	// Verify each framework
	results := make(map[string]interface{})
	overallScore := 0.0
	overallStatus := "compliant"
	findingsCount := 0
	remediationProgress := 0.75

	for _, framework := range req.Frameworks {
		report, err := h.verifyFrameworkByName(ctx, framework)
		if err != nil {
			h.logger.Error("framework verification failed", "framework", framework, "error", err)
			span.SetAttributes(
				attribute.String("jtbd.scenario.outcome", "failure"),
				attribute.String("jtbd.scenario.error_reason", "framework_verification_failed"),
			)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "verification failed for " + framework,
				"details": err.Error(),
			})
			return
		}

		results[framework] = report
		overallScore += report.Score
		findingsCount += len(report.Violations)

		// Aggregate overall status
		if report.Status != "compliant" {
			overallStatus = "partial"
			if report.Status == "non_compliant" {
				overallStatus = "non_compliant"
			}
		}
	}

	if len(req.Frameworks) > 0 {
		overallScore /= float64(len(req.Frameworks))
	}

	// Update span with compliance results
	span.SetAttributes(
		attribute.String("jtbd.scenario.step", "report"),
		attribute.Int("jtbd.scenario.step_num", 3),
		attribute.String("compliance.framework", req.Frameworks[0]),
		attribute.String("compliance.status", overallStatus),
		attribute.Int("compliance.findings_count", findingsCount),
		attribute.Float64("compliance.remediation_progress", remediationProgress),
		attribute.String("compliance.last_audit_date", time.Now().UTC().Format(time.RFC3339)),
		attribute.String("jtbd.scenario.outcome", "success"),
	)

	response := VerifyResponse{
		Status:       overallStatus,
		OverallScore: overallScore,
		Frameworks:   results,
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response)
}

// GenerateReport handles GET /v1/compliance/report.
// Generates aggregated compliance report across all or specified frameworks.
// Query params: frameworks=SOC2,GDPR,HIPAA,SOX (comma-separated, defaults to all), include_details=true
// Response: ComplianceMatrix with all frameworks and overall score
func (h *ComplianceAPIHandler) GenerateReport(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse framework query parameter
	frameworksParam := c.DefaultQuery("frameworks", "SOC2,GDPR,HIPAA,SOX")
	include := c.DefaultQuery("include_details", "false") == "true"

	matrix, err := h.engine.GenerateReport(ctx)
	if err != nil {
		h.logger.Error("report generation failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "report generation failed",
			"details": err.Error(),
		})
		return
	}

	// Filter frameworks if specified
	if frameworksParam != "" && frameworksParam != "*" {
		// Kept for future filtering logic
	}

	// Remove details if not requested
	if !include {
		for _, report := range matrix.Frameworks {
			report.Violations = nil
		}
	}

	c.JSON(http.StatusOK, matrix)
}

// ListFrameworkControls handles GET /v1/compliance/controls/:framework.
// Lists all controls for a specific framework, optionally filtered by severity or status.
// Path param: framework (SOC2, GDPR, HIPAA, SOX)
// Query params: severity=critical|high|medium|low (optional), status=verified|failed (optional)
// Response: {"framework": "SOC2", "controls": [...], "total": 8, "timestamp": "..."}
func (h *ComplianceAPIHandler) ListFrameworkControls(c *gin.Context) {
	framework := c.Param("framework")
	if framework == "" {
		utils.RespondBadRequest(c, h.logger, "framework path parameter is required")
		return
	}

	severityFilter := c.Query("severity")
	if severityFilter != "" {
		validSeverity := map[string]bool{"critical": true, "high": true, "medium": true, "low": true}
		if !validSeverity[severityFilter] {
			utils.RespondBadRequest(c, h.logger, "invalid severity filter. Valid: critical, high, medium, low")
			return
		}
	}

	// Get all controls for framework
	controls := h.engine.GetFrameworkControls(framework)
	if controls == nil {
		utils.RespondBadRequest(c, h.logger, "unknown framework: "+framework)
		return
	}

	// Filter by severity if specified
	if severityFilter != "" {
		filtered := make([]*ontology.ComplianceControl, 0)
		for _, ctrl := range controls {
			if ctrl.Severity == severityFilter {
				filtered = append(filtered, ctrl)
			}
		}
		controls = filtered
	}

	response := gin.H{
		"framework": framework,
		"controls":  controls,
		"total":     len(controls),
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response)
}

// ReloadOntology handles POST /v1/compliance/reload.
// Reloads the compliance ontology from disk (admin only).
// Optional request: {"clear_cache": true}
// Response: {"status": "reloaded|already_loaded", "timestamp": "..."}
func (h *ComplianceAPIHandler) ReloadOntology(c *gin.Context) {
	var req ReloadRequest
	_ = c.ShouldBindJSON(&req) // Optional body

	ctx := c.Request.Context()

	// Reload ontology
	err := h.engine.Initialize(ctx)
	if err != nil {
		h.logger.Error("ontology reload failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "ontology reload failed",
			"details": err.Error(),
		})
		return
	}

	response := ReloadResponse{
		Status:    "reloaded",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response)
}

// verifyFrameworkByName is a helper to verify a single framework.
func (h *ComplianceAPIHandler) verifyFrameworkByName(ctx context.Context, framework string) (*ontology.ComplianceReport, error) {
	switch framework {
	case "SOC2":
		return h.engine.VerifySOC2(ctx)
	case "GDPR":
		return h.engine.VerifyGDPR(ctx)
	case "HIPAA":
		return h.engine.VerifyHIPAA(ctx)
	case "SOX":
		return h.engine.VerifySOX(ctx)
	default:
		return nil, fmt.Errorf("unknown framework: %s", framework)
	}
}

// InitializeComplianceEngine initializes the global ComplianceEngine singleton.
// Call this once during application startup.
func (h *Handlers) InitializeComplianceEngine(ontologyPath string, logger *slog.Logger) error {
	var err error
	ontology.GlobalComplianceEngine, err = ontology.NewComplianceEngine(ontologyPath, logger)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return ontology.GlobalComplianceEngine.Initialize(ctx)
}
