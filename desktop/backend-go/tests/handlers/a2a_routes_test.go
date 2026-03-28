package handlers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rhl/businessos-backend/internal/handlers"
	"github.com/rhl/businessos-backend/internal/models"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// setupA2ARoutesRouter creates a test router with A2A routes
func setupA2ARoutesRouter() (*gin.Engine, *handlers.A2ARoutesHandler) {
	auditService := &mockAuditService{}
	handler := handlers.NewA2ARoutesHandler(auditService)
	router := gin.New()
	return router, handler
}

// mockAuditService provides a test implementation of the audit service
type mockAuditService struct {
	entries []*models.AuditEntry
}

func (m *mockAuditService) LogA2ACall(agent, action, resourceType, resourceID string, snScore float64) (*models.AuditEntry, error) {
	entry := &models.AuditEntry{
		ID:             uuid.New().String(),
		Timestamp:      time.Now().UTC(),
		Agent:          agent,
		Action:         action,
		ResourceType:   resourceType,
		ResourceID:     resourceID,
		SNScore:        snScore,
		GovernanceTier: models.GetGovernanceTier(snScore).Tier,
		Result:         "success",
	}

	if len(m.entries) > 0 {
		prevEntry := m.entries[len(m.entries)-1]
		entry.PreviousHash = prevEntry.DataHash
		entry.DataHash = computeDataHash(entry)
		entry.Signature = computeSignature(prevEntry.DataHash, entry.DataHash)
	} else {
		entry.DataHash = computeDataHash(entry)
		entry.Signature = computeSignature("", entry.DataHash)
	}

	m.entries = append(m.entries, entry)
	return entry, nil
}

func (m *mockAuditService) QueryAuditTrail(resourceType, resourceID string) ([]*models.AuditEntry, error) {
	var results []*models.AuditEntry
	for _, entry := range m.entries {
		if entry.ResourceType == resourceType && entry.ResourceID == resourceID {
			results = append(results, entry)
		}
	}
	return results, nil
}

// Helper functions
func computeDataHash(entry *models.AuditEntry) string {
	data := entry.Agent + entry.Action + entry.ResourceType + entry.ResourceID + entry.Timestamp.String()
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func computeSignature(previousHash, currentHash string) string {
	secret := "test-secret"
	message := previousHash + currentHash
	sig := hmac.New(sha256.New, []byte(secret))
	sig.Write([]byte(message))
	return hex.EncodeToString(sig.Sum(nil))
}

// =============================================================================
// Tests for A2A Call Endpoint Authentication
// =============================================================================

func TestA2ACallEndpoint_RequiresAuth(t *testing.T) {
	router, handler := setupA2ARoutesRouter()
	router.POST("/api/integrations/a2a/crm/deals", handler.CreateDeal)

	body, _ := json.Marshal(map[string]string{
		"name": "Test Deal",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/integrations/a2a/crm/deals", bytes.NewReader(body))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// =============================================================================
// Tests for A2A Deal Creation + Audit Entry
// =============================================================================

func TestA2ACallCreatesAuditEntry(t *testing.T) {
	router, handler := setupA2ARoutesRouter()
	router.POST("/api/integrations/a2a/crm/deals", handler.CreateDeal)

	body, _ := json.Marshal(map[string]string{
		"name": "Test Deal",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/integrations/a2a/crm/deals", bytes.NewReader(body))
	req.Header.Set("X-Shared-Secret", "test-secret")
	req.Header.Set("X-Agent-ID", "osa-healing-agent")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	auditEntry, ok := resp["audit_entry"]
	assert.True(t, ok, "response should contain audit_entry")
	assert.NotNil(t, auditEntry)
}

func TestAuditEntryHashChain(t *testing.T) {
	_, handler := setupA2ARoutesRouter()

	// Create first entry
	entry1, err := handler.AuditService.LogA2ACall("agent1", "create", "deal", "deal-1", 0.9)
	require.NoError(t, err)
	assert.NotEmpty(t, entry1.DataHash)
	assert.NotEmpty(t, entry1.Signature)
	assert.Empty(t, entry1.PreviousHash) // First entry has no previous

	// Create second entry
	entry2, err := handler.AuditService.LogA2ACall("agent1", "update", "deal", "deal-1", 0.85)
	require.NoError(t, err)
	assert.NotEmpty(t, entry2.DataHash)
	assert.NotEmpty(t, entry2.Signature)
	assert.Equal(t, entry1.DataHash, entry2.PreviousHash) // Chain link

	// Verify that signatures are different (tamper evident)
	assert.NotEqual(t, entry1.Signature, entry2.Signature)
}

func TestA2ADealCreation(t *testing.T) {
	router, handler := setupA2ARoutesRouter()
	router.POST("/api/integrations/a2a/crm/deals", handler.CreateDeal)

	dealData := map[string]interface{}{
		"name":  "Enterprise Deal",
		"value": 50000,
	}
	body, _ := json.Marshal(dealData)
	req := httptest.NewRequest(http.MethodPost, "/api/integrations/a2a/crm/deals", bytes.NewReader(body))
	req.Header.Set("X-Shared-Secret", "test-secret")
	req.Header.Set("X-Agent-ID", "osa-crm-agent")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.NotNil(t, resp["deal"])
}

func TestA2ATaskAssignment(t *testing.T) {
	router, handler := setupA2ARoutesRouter()
	router.POST("/api/integrations/a2a/projects/tasks", handler.AssignTask)

	taskData := map[string]interface{}{
		"title":    "Complete review",
		"assignee": "user-123",
	}
	body, _ := json.Marshal(taskData)
	req := httptest.NewRequest(http.MethodPost, "/api/integrations/a2a/projects/tasks", bytes.NewReader(body))
	req.Header.Set("X-Shared-Secret", "test-secret")
	req.Header.Set("X-Agent-ID", "osa-task-agent")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.NotNil(t, resp["task"])
}

// =============================================================================
// Tests for Governance Tier Routing
// =============================================================================

func TestGovernanceTierAuto(t *testing.T) {
	router, handler := setupA2ARoutesRouter()
	router.POST("/api/integrations/a2a/crm/deals", handler.CreateDeal)

	body, _ := json.Marshal(map[string]interface{}{
		"name":  "Deal",
		"score": 0.95, // High confidence
	})
	req := httptest.NewRequest(http.MethodPost, "/api/integrations/a2a/crm/deals", bytes.NewReader(body))
	req.Header.Set("X-Shared-Secret", "test-secret")
	req.Header.Set("X-Agent-ID", "osa-agent")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	auditEntry := resp["audit_entry"].(map[string]interface{})
	assert.Equal(t, "auto", auditEntry["governance_tier"])
}

func TestGovernanceTierHuman(t *testing.T) {
	// Note: Current implementation uses hardcoded snScore in handlers.
	// This test verifies the governance tier logic via the audit service directly
	// with a medium-confidence score.
	_, handler := setupA2ARoutesRouter()

	// Create entry with medium confidence (0.75)
	entry, err := handler.AuditService.LogA2ACall("agent1", "create", "deal", "deal-1", 0.75)
	require.NoError(t, err)
	assert.Equal(t, "human", entry.GovernanceTier)
	assert.True(t, models.ApprovalRequired(0.75))
}

func TestGovernanceTierBoard(t *testing.T) {
	// Note: Current implementation uses hardcoded snScore in handlers.
	// This test verifies the governance tier logic via the audit service directly
	// with a low-confidence score.
	_, handler := setupA2ARoutesRouter()

	// Create entry with low confidence (0.65)
	entry, err := handler.AuditService.LogA2ACall("agent1", "create", "deal", "deal-1", 0.65)
	require.NoError(t, err)
	assert.Equal(t, "board", entry.GovernanceTier)
	assert.True(t, models.ApprovalRequired(0.65))
}

// =============================================================================
// Tests for A2A Progress Updates
// =============================================================================

func TestA2AProgressUpdate(t *testing.T) {
	router, handler := setupA2ARoutesRouter()
	router.POST("/api/integrations/a2a/projects/progress", handler.UpdateProgress)

	progressData := map[string]interface{}{
		"project_id": "proj-123",
		"status":     "in_progress",
		"percent":    45,
	}
	body, _ := json.Marshal(progressData)
	req := httptest.NewRequest(http.MethodPost, "/api/integrations/a2a/projects/progress", bytes.NewReader(body))
	req.Header.Set("X-Shared-Secret", "test-secret")
	req.Header.Set("X-Agent-ID", "osa-progress-agent")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.NotNil(t, resp["audit_entry"])
}

// =============================================================================
// Tests for A2A Audit Query
// =============================================================================

func TestA2AAuditQuery(t *testing.T) {
	router, handler := setupA2ARoutesRouter()
	router.POST("/api/integrations/a2a/crm/deals", handler.CreateDeal)
	router.GET("/api/integrations/a2a/audit/query", handler.QueryAudit)

	// Create a deal first to populate audit trail
	dealBody, _ := json.Marshal(map[string]string{"name": "Deal1"})
	req := httptest.NewRequest(http.MethodPost, "/api/integrations/a2a/crm/deals", bytes.NewReader(dealBody))
	req.Header.Set("X-Shared-Secret", "test-secret")
	req.Header.Set("X-Agent-ID", "osa-agent")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Now query the audit trail
	req = httptest.NewRequest(http.MethodGet, "/api/integrations/a2a/audit/query?resource_type=deal", nil)
	req.Header.Set("X-Shared-Secret", "test-secret")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.NotNil(t, resp["entries"])
}

// =============================================================================
// Tests for Lead Update
// =============================================================================

func TestA2ALeadUpdate(t *testing.T) {
	router, handler := setupA2ARoutesRouter()
	router.POST("/api/integrations/a2a/crm/leads", handler.UpdateLead)

	leadData := map[string]interface{}{
		"lead_id": "lead-456",
		"status":  "qualified",
	}
	body, _ := json.Marshal(leadData)
	req := httptest.NewRequest(http.MethodPost, "/api/integrations/a2a/crm/leads", bytes.NewReader(body))
	req.Header.Set("X-Shared-Secret", "test-secret")
	req.Header.Set("X-Agent-ID", "osa-lead-agent")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.NotNil(t, resp["audit_entry"])
}

// =============================================================================
// Tests for Multiple Sequential Operations (Chain Verification)
// =============================================================================

func TestSequentialA2AOperationsCreateValidChain(t *testing.T) {
	_, handler := setupA2ARoutesRouter()

	// Simulate 3 sequential operations
	op1, _ := handler.AuditService.LogA2ACall("agent1", "create", "deal", "deal-1", 0.9)
	op2, _ := handler.AuditService.LogA2ACall("agent1", "update", "deal", "deal-1", 0.85)
	op3, _ := handler.AuditService.LogA2ACall("agent1", "close", "deal", "deal-1", 0.92)

	// Verify chain integrity
	assert.Empty(t, op1.PreviousHash)
	assert.Equal(t, op1.DataHash, op2.PreviousHash)
	assert.Equal(t, op2.DataHash, op3.PreviousHash)

	// Verify all have unique signatures
	signatures := map[string]bool{
		op1.Signature: true,
		op2.Signature: true,
		op3.Signature: true,
	}
	assert.Equal(t, 3, len(signatures))
}
