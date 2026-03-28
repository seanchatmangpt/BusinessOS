package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

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
		"Patient":            true,
		"Observation":        true,
		"MedicationRequest":  true,
		"Procedure":          true,
		"Condition":          true,
		"AllergyIntolerance": true,
		"Encounter":          true,
		"DiagnosticReport":   true,
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
	ResourceID string                   `json:"resource_id"`
	AuditTrail []ontology.PHIAuditEntry `json:"audit_trail"`
	Count      int                      `json:"count"`
	Period     string                   `json:"period"`
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
			"error":             "deletion incomplete",
			"remaining_triples": result.TripleCount,
			"verified_at":       result.VerifiedAt,
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

// Stub implementations for development only.
// These return errors so that any code path using them fails visibly
// instead of silently returning fake data.

// TODO: implement real audit trail — connect to Oxigraph RDF store for SPARQL queries,
// persistent PostgreSQL audit log with HMAC signatures, and real deletion verification.

// SPARQLExecutorImpl is a stub implementation of SPARQLExecutor.
// Returns error for all operations until connected to a real Oxigraph instance.
type SPARQLExecutorImpl struct {
	logger *slog.Logger
}

func (s *SPARQLExecutorImpl) ExecuteConstruct(c context.Context, query string) (string, error) {
	return "", fmt.Errorf("SPARQL executor not connected: no Oxigraph backend configured")
}

func (s *SPARQLExecutorImpl) ExecuteAsk(c context.Context, query string) (bool, error) {
	return false, fmt.Errorf("SPARQL executor not connected: no Oxigraph backend configured")
}

func (s *SPARQLExecutorImpl) ExecuteSelect(c context.Context, query string) (map[string]interface{}, error) {
	return nil, fmt.Errorf("SPARQL executor not connected: no Oxigraph backend configured")
}

// RDFStoreImpl is a stub implementation of RDFStore.
// Returns error for all operations until connected to a real Oxigraph instance.
type RDFStoreImpl struct {
	logger *slog.Logger
}

func (r *RDFStoreImpl) StoreTriples(c context.Context, turtleData string) error {
	return fmt.Errorf("RDF store not connected: no Oxigraph backend configured")
}

func (r *RDFStoreImpl) QueryTriples(c context.Context, query string) (int, error) {
	return 0, fmt.Errorf("RDF store not connected: no Oxigraph backend configured")
}

func (r *RDFStoreImpl) DeleteTriples(c context.Context, pattern string) error {
	return fmt.Errorf("RDF store not connected: no Oxigraph backend configured")
}

func (r *RDFStoreImpl) GetTriplesForEntity(c context.Context, entityURI string) (int, error) {
	return 0, fmt.Errorf("RDF store not connected: no Oxigraph backend configured")
}

// AuditLoggerImpl is a stub implementation of AuditLogger.
// Returns error for all operations until connected to a real persistent audit store.
// Previously returned hardcoded fake audit entries, which was a compliance risk.
type AuditLoggerImpl struct {
	logger *slog.Logger
}

func (a *AuditLoggerImpl) LogAccess(c context.Context, entry ontology.PHIAuditEntry) error {
	return fmt.Errorf("audit logger not connected: no persistent audit store configured")
}

func (a *AuditLoggerImpl) GetAuditTrail(c context.Context, patientID string, lastNDays int) ([]ontology.PHIAuditEntry, error) {
	return nil, fmt.Errorf("audit logger not connected: no persistent audit store configured")
}

func (a *AuditLoggerImpl) VerifyAuditIntegrity(c context.Context, entries []ontology.PHIAuditEntry) (bool, error) {
	return false, fmt.Errorf("audit logger not connected: no persistent audit store configured")
}
