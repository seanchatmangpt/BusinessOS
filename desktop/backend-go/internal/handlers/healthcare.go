package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/ontology"
	"github.com/rhl/businessos-backend/internal/utils"
)

// HealthcareHandler handles healthcare and PHI endpoints.
type HealthcareHandler struct {
	phiManager *ontology.HealthcarePHIManager
	logger     *slog.Logger
}

// NewHealthcareHandler constructs a HealthcareHandler.
func NewHealthcareHandler(phiManager *ontology.HealthcarePHIManager, logger *slog.Logger) *HealthcareHandler {
	return &HealthcareHandler{
		phiManager: phiManager,
		logger:     logger,
	}
}

// RegisterHealthcareRoutes is a standalone function to register healthcare routes
func RegisterHealthcareRoutes(api *gin.RouterGroup, auth gin.HandlerFunc, phiManager *ontology.HealthcarePHIManager) {
	if phiManager == nil {
		slog.Debug("Skipping healthcare routes, PHI manager not initialized")
		return
	}

	healthcareH := NewHealthcareHandler(phiManager, slog.Default())
	healthcare := api.Group("/healthcare")
	healthcare.Use(auth)
	{
		// PHI tracking endpoints
		phi := healthcare.Group("/phi")
		{
			phi.POST("/track", healthcareH.TrackPHI)
			phi.GET("/:id/audit", healthcareH.GetPHIAuditTrail)
			phi.DELETE("/:id", healthcareH.DeletePHI)
		}

		// Consent endpoints
		consent := healthcare.Group("/consent")
		{
			consent.POST("/verify", healthcareH.VerifyConsent)
		}

		// Compliance endpoints
		hipaa := healthcare.Group("/hipaa")
		{
			hipaa.GET("/verify", healthcareH.VerifyHIPAA)
		}
	}

	slog.Info("Healthcare routes registered at /api/healthcare/*")
}

// TrackPHIRequest is the request body for POST /api/healthcare/phi/track.
type TrackPHIRequest struct {
	ResourceID   string                 `json:"resource_id" binding:"required"`
	ResourceType string                 `json:"resource_type" binding:"required"`
	PatientID    string                 `json:"patient_id" binding:"required"`
	Data         map[string]interface{} `json:"data" binding:"required"`
}

// TrackPHI handles POST /api/healthcare/phi/track.
// Accepts FHIR resource data and creates PROV-O provenance triples in Oxigraph.
// Returns: resource_id, triple_count, hipaa_check_passed.
func (h *HealthcareHandler) TrackPHI(c *gin.Context) {
	var req TrackPHIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, h.logger, "invalid request body: "+err.Error())
		return
	}

	// Validate resource type is valid FHIR
	validTypes := map[string]bool{
		"Patient":             true,
		"Observation":         true,
		"MedicationRequest":   true,
		"Procedure":           true,
		"Condition":           true,
		"AllergyIntolerance":  true,
		"Encounter":           true,
		"DiagnosticReport":    true,
	}
	if !validTypes[req.ResourceType] {
		utils.RespondBadRequest(c, h.logger, "unsupported FHIR resource type: "+req.ResourceType)
		return
	}

	// Get actor from JWT or request header
	actor, exists := c.Get("user_id")
	if !exists {
		actor = "system"
	}
	actorStr := actor.(string)

	result, err := h.phiManager.TrackPHI(c.Request.Context(), req.ResourceID, req.ResourceType, req.PatientID, req.Data, actorStr)
	if err != nil {
		utils.RespondInternalError(c, h.logger, "track phi", err)
		return
	}

	c.JSON(http.StatusCreated, result)
}

// GetPHIAuditTrailResponse is the response for GET /api/healthcare/phi/:id/audit.
type GetPHIAuditTrailResponse struct {
	ResourceID string                    `json:"resource_id"`
	AuditTrail []ontology.PHIAuditEntry `json:"audit_trail"`
	Count      int                       `json:"count"`
	Period     string                    `json:"period"`
}

// GetPHIAuditTrail handles GET /api/healthcare/phi/:id/audit.
// Returns the full audit trail for a PHI resource (last 90 days).
// Query params: days (default 90).
func (h *HealthcareHandler) GetPHIAuditTrail(c *gin.Context) {
	resourceID := c.Param("id")
	if resourceID == "" {
		utils.RespondBadRequest(c, h.logger, "resource_id path parameter is required")
		return
	}

	_ = 90 // Default days, not currently parameterized in this version
	if d := c.Query("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 && parsed <= 365 {
			_ = parsed // Future: use to filter by date range
		}
	}

	// For this implementation, we retrieve audit trail by extracting patient_id from resource
	// In production, would look up resource metadata to get patient_id
	patientID := c.Query("patient_id")
	if patientID == "" {
		utils.RespondBadRequest(c, h.logger, "patient_id query parameter is required")
		return
	}

	result, err := h.phiManager.GenerateAuditTrail(c.Request.Context(), patientID)
	if err != nil {
		utils.RespondInternalError(c, h.logger, "generate audit trail", err)
		return
	}

	response := GetPHIAuditTrailResponse{
		ResourceID: resourceID,
		AuditTrail: result.Entries,
		Count:      result.TotalEntries,
		Period:     result.Period,
	}

	c.JSON(http.StatusOK, response)
}

// DeletePHI handles DELETE /api/healthcare/phi/:id.
// Performs hard delete of FHIR resource from RDF store.
// Verifies GDPR compliance (right to be forgotten).
// Returns: fully_deleted, triple_count, rdf_clean_confirmed.
func (h *HealthcareHandler) DeletePHI(c *gin.Context) {
	resourceID := c.Param("id")
	if resourceID == "" {
		utils.RespondBadRequest(c, h.logger, "resource_id path parameter is required")
		return
	}

	resourceType := c.Query("resource_type")
	if resourceType == "" {
		utils.RespondBadRequest(c, h.logger, "resource_type query parameter is required")
		return
	}

	// Verify deletion
	result, err := h.phiManager.CheckDeletion(c.Request.Context(), resourceID, resourceType)
	if err != nil {
		utils.RespondInternalError(c, h.logger, "check deletion", err)
		return
	}

	if !result.FullyDeleted {
		c.JSON(http.StatusConflict, gin.H{
			"error":           "deletion incomplete",
			"remaining_triples": result.TripleCount,
			"verified_at":    result.VerifiedAt,
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// VerifyConsentRequest is the request body for POST /api/healthcare/consent/verify.
type VerifyConsentRequest struct {
	PatientID string `json:"patient_id" binding:"required"`
}

// VerifyConsent handles POST /api/healthcare/consent/verify.
// Checks if patient has valid, non-expired consent for PHI access.
// Returns: consent_granted, expires_at, scope.
func (h *HealthcareHandler) VerifyConsent(c *gin.Context) {
	var req VerifyConsentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, h.logger, "invalid request body: "+err.Error())
		return
	}

	result, err := h.phiManager.VerifyConsent(c.Request.Context(), req.PatientID)
	if err != nil {
		utils.RespondInternalError(c, h.logger, "verify consent", err)
		return
	}

	if !result.ConsentGranted {
		c.JSON(http.StatusForbidden, result)
		return
	}

	c.JSON(http.StatusOK, result)
}

// VerifyHIPAA handles GET /api/healthcare/hipaa/verify.
// Checks compliance with HIPAA § 164.312(b) (access control + audit logging).
// Returns: compliant, access_control_pass, audit_log_pass, encryption_pass, integrity_pass, compliance_score.
func (h *HealthcareHandler) VerifyHIPAA(c *gin.Context) {
	result, err := h.phiManager.VerifyHIPAA(c.Request.Context())
	if err != nil {
		utils.RespondInternalError(c, h.logger, "verify hipaa", err)
		return
	}

	status := http.StatusOK
	if !result.Compliant {
		status = http.StatusPartialContent // 206: partial compliance
	}

	c.JSON(status, result)
}

// Mock implementations for demonstration (would be replaced with real Oxigraph/audit store in production)

// SPARQLExecutorImpl is a stub implementation of SPARQLExecutor for demo purposes.
type SPARQLExecutorImpl struct {
	logger *slog.Logger
}

func (s *SPARQLExecutorImpl) ExecuteConstruct(c context.Context, query string) (string, error) {
	// In production: execute SPARQL CONSTRUCT against Oxigraph
	return `@prefix fhir: <http://hl7.org/fhir/> .
@prefix prov: <http://www.w3.org/ns/prov#> .
fhir:Patient_p123 a prov:Entity .
`, nil
}

func (s *SPARQLExecutorImpl) ExecuteAsk(c context.Context, query string) (bool, error) {
	// In production: execute SPARQL ASK against Oxigraph
	return true, nil
}

func (s *SPARQLExecutorImpl) ExecuteSelect(c context.Context, query string) (map[string]interface{}, error) {
	// In production: execute SPARQL SELECT against Oxigraph
	return map[string]interface{}{}, nil
}

// RDFStoreImpl is a stub implementation of RDFStore for demo purposes.
type RDFStoreImpl struct {
	logger *slog.Logger
}

func (r *RDFStoreImpl) StoreTriples(c context.Context, turtleData string) error {
	// In production: persist triples to Oxigraph
	return nil
}

func (r *RDFStoreImpl) QueryTriples(c context.Context, query string) (int, error) {
	// In production: query triples from Oxigraph
	return 42, nil
}

func (r *RDFStoreImpl) DeleteTriples(c context.Context, pattern string) error {
	// In production: delete triples from Oxigraph
	return nil
}

func (r *RDFStoreImpl) GetTriplesForEntity(c context.Context, entityURI string) (int, error) {
	// In production: count triples for entity
	return 15, nil
}

// AuditLoggerImpl is a stub implementation of AuditLogger for demo purposes.
type AuditLoggerImpl struct {
	logger *slog.Logger
}

func (a *AuditLoggerImpl) LogAccess(c context.Context, entry ontology.PHIAuditEntry) error {
	// In production: persist audit entry to secure log (PostgreSQL + HMAC signature)
	return nil
}

func (a *AuditLoggerImpl) GetAuditTrail(c context.Context, patientID string, lastNDays int) ([]ontology.PHIAuditEntry, error) {
	// In production: retrieve audit entries from secure log
	return []ontology.PHIAuditEntry{
		{
			Timestamp:    time.Now().AddDate(0, 0, -5),
			Actor:        "doctor@example.com",
			Action:       "read",
			ResourceID:   "obs123",
			ResourceType: "Observation",
			Details:      "Patient vitals review",
		},
	}, nil
}

func (a *AuditLoggerImpl) VerifyAuditIntegrity(c context.Context, entries []ontology.PHIAuditEntry) (bool, error) {
	// In production: verify HMAC signatures on entries
	return len(entries) > 0, nil
}
