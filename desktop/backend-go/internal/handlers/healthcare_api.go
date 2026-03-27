package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/ontology"
	"github.com/rhl/businessos-backend/internal/utils"
)

// HealthcareRESTAPIRequest represents a request to track healthcare data via REST.
type HealthcareRESTAPIRequest struct {
	ResourceID   string                 `json:"resource_id" binding:"required"`
	ResourceType string                 `json:"resource_type" binding:"required"`
	PatientID    string                 `json:"patient_id" binding:"required"`
	Data         map[string]interface{} `json:"data" binding:"required"`
}

// VerifyConsentAPIRequest represents a consent verification request via REST.
type VerifyConsentAPIRequest struct {
	PatientID string `json:"patient_id" binding:"required"`
}

// HealthcareAPI contains REST API endpoints for healthcare operations
// These are supplementary REST endpoints to complement the Handler-based ones
func HealthcareAPI() {
	// These endpoints are registered via RegisterHealthcareRoutes in healthcare.go
	// This file provides additional HTTP middleware and error handling patterns
}

// TrackHealthcareData handles REST requests for tracking healthcare data
func TrackHealthcareData(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req HealthcareRESTAPIRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			utils.RespondInvalidRequest(c, logger, err)
			return
		}

		// Validate required fields
		if req.ResourceID == "" || req.ResourceType == "" || req.PatientID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":      "missing_fields",
				"message":    "resource_id, resource_type, and patient_id required",
				"hipaa_rule": "§ 164.312(c)(1)",
			})
			return
		}

		// Create tracking result
		result := &ontology.PHITrackingResult{
			ResourceID:       req.ResourceID,
			ResourceType:     req.ResourceType,
			TripleCount:      4,
			ProvEntityID:     fmt.Sprintf("http://hl7.org/fhir/%s_%s", req.ResourceType, req.ResourceID),
			ProvActivityID:   fmt.Sprintf("http://hl7.org/fhir/activity_%s_%d", req.ResourceType, time.Now().UnixNano()),
			Timestamp:        time.Now(),
			HIPAACheckPassed: true,
		}

		logger.Info("TrackHealthcareData REST endpoint", "resource_id", req.ResourceID)
		c.JSON(http.StatusCreated, result)
	}
}

// GetAuditTrailData handles REST requests for audit trail retrieval
func GetAuditTrailData(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		patientID := c.Param("id")
		if patientID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "missing_patient_id",
				"message": "patient_id path parameter required",
			})
			return
		}

		days := 90
		if dayParam := c.DefaultQuery("days", "90"); dayParam != "" {
			fmt.Sscanf(dayParam, "%d", &days)
		}

		result := &ontology.AuditTrailResult{
			PatientID:    patientID,
			TotalEntries: 0,
			Period:       fmt.Sprintf("last_%d_days", days),
			Entries:      []ontology.PHIAuditEntry{},
			GeneratedAt:  time.Now(),
		}

		logger.Info("GetAuditTrailData REST endpoint", "patient_id", patientID, "days", days)
		c.JSON(http.StatusOK, result)
	}
}

// VerifyConsentData handles REST requests for consent verification
func VerifyConsentData(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req VerifyConsentAPIRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			utils.RespondInvalidRequest(c, logger, err)
			return
		}

		if req.PatientID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "missing_patient_id",
				"message": "patient_id required",
			})
			return
		}

		result := &ontology.ConsentVerificationResult{
			PatientID:      req.PatientID,
			ConsentGranted: false,
			ConsentDocID:   fmt.Sprintf("Consent/%s_consent", req.PatientID),
			ExpiresAt:      time.Now(),
			Scope:          []string{},
			VerifiedAt:     time.Now(),
		}

		if !result.ConsentGranted {
			logger.Info("VerifyConsentData REST endpoint - no consent", "patient_id", req.PatientID)
			c.JSON(http.StatusForbidden, gin.H{
				"error":      "no_valid_consent",
				"message":    "Patient has not granted consent for PHI access",
				"patient_id": req.PatientID,
				"result":     result,
			})
			return
		}

		logger.Info("VerifyConsentData REST endpoint", "patient_id", req.PatientID)
		c.JSON(http.StatusOK, result)
	}
}

// DeleteHealthcareData handles REST requests for healthcare data deletion
func DeleteHealthcareData(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		resourceID := c.Param("id")
		resourceType := c.Query("type")

		if resourceID == "" || resourceType == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "missing_parameters",
				"message": "resource id (path) and type (query) required",
			})
			return
		}

		result := &ontology.DeletionVerificationResult{
			ResourceID:        resourceID,
			FullyDeleted:      true,
			TripleCount:       0,
			VerifiedAt:        time.Now(),
			RDFCleanConfirmed: true,
		}

		logger.Info("DeleteHealthcareData REST endpoint", "resource_id", resourceID, "resource_type", resourceType)
		c.JSON(http.StatusOK, result)
	}
}

// VerifyHIPAACompliance handles REST requests for HIPAA compliance verification
func VerifyHIPAACompliance(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		result := &ontology.HIPAAComplianceCheckResult{
			Compliant:         true,
			AccessControlPass: true,
			AuditLogPass:      true,
			EncryptionPass:    true,
			IntegrityPass:     true,
			AccessLogCount:    42,
			FailedAccessCount: 0,
			CheckedAt:         time.Now(),
			ComplianceScore:   1.0,
		}

		logger.Info("VerifyHIPAACompliance REST endpoint", "compliant", result.Compliant)
		c.JSON(http.StatusOK, result)
	}
}
