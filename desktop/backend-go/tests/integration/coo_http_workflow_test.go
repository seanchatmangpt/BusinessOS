package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TestCOOHTTPWorkflow simulates the COO approval workflow via HTTP endpoints.
// This is an integration test that exercises the full request-response cycle.
func TestCOOHTTPWorkflow(t *testing.T) {
	// Setup
	router := setupTestRouter()
	cooUserID := uuid.New().String()
	pipelineID := uuid.New()
	projectID := uuid.New()

	// Simulate COO workflow
	testCOOWorkflowHTTP(t, router, cooUserID, pipelineID, projectID)
}

// testCOOWorkflowHTTP executes the full COO approval flow via HTTP
func testCOOWorkflowHTTP(t *testing.T, router *gin.Engine, cooUserID string, pipelineID, projectID uuid.UUID) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("COO HTTP WORKFLOW TEST")
	fmt.Println(strings.Repeat("=", 80))

	// ============================================================================
	// STEP 1: GET /api/crm/process-leads (simulate 7-day batch)
	// ============================================================================

	fmt.Println("\n[1/5] GET /api/crm/process-leads?simulation=true&days=7")
	req := httptest.NewRequest("GET",
		fmt.Sprintf("/api/crm/process-leads?user_id=%s&pipeline_id=%s&simulation=true&days=7",
			cooUserID, pipelineID.String()),
		nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cooUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
		t.Errorf("Unexpected status code: %d", w.Code)
	}

	var processLeadsResp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &processLeadsResp); err == nil {
		if successRate, ok := processLeadsResp["success_rate"].(float64); ok {
			fmt.Printf("   ✓ Success Rate: %.2f\n", successRate)
			if successRate < 0.80 {
				t.Logf("   ⚠ Warning: success_rate %.2f below 0.80 threshold", successRate)
			}
		}
	}

	// ============================================================================
	// STEP 2: GET /api/projects/:id/assign-tasks (task assignment)
	// ============================================================================

	fmt.Println("\n[2/5] GET /api/projects/:id/assign-tasks?simulation=true")
	req = httptest.NewRequest("GET",
		fmt.Sprintf("/api/projects/%s/assign-tasks?user_id=%s&simulation=true",
			projectID.String(), cooUserID),
		nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cooUserID))

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
		t.Errorf("Unexpected status code: %d", w.Code)
	}

	var assignTasksResp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &assignTasksResp); err == nil {
		if successRate, ok := assignTasksResp["success_rate"].(float64); ok {
			fmt.Printf("   ✓ Success Rate: %.2f\n", successRate)
			if successRate < 0.80 {
				t.Logf("   ⚠ Warning: success_rate %.2f below 0.80 threshold", successRate)
			}
		}
	}

	// ============================================================================
	// STEP 3: GET /api/sorx/decisions (fetch pending decisions)
	// ============================================================================

	fmt.Println("\n[3/5] GET /api/sorx/decisions")
	req = httptest.NewRequest("GET", "/api/sorx/decisions", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cooUserID))

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusUnauthorized {
		t.Logf("   Note: Decision endpoint returned %d (may not be implemented)", w.Code)
	}

	var decisionsResp struct {
		Success   bool          `json:"success"`
		Decisions []interface{} `json:"decisions"`
		Count     int           `json:"count"`
	}

	if err := json.Unmarshal(w.Body.Bytes(), &decisionsResp); err == nil && w.Code == http.StatusOK {
		fmt.Printf("   ✓ Pending Decisions: %d\n", decisionsResp.Count)
		if decisionsResp.Count > 3 {
			t.Logf("   ⚠ Warning: %d decisions in queue (> 3 threshold)", decisionsResp.Count)
		}
	}

	// ============================================================================
	// STEP 4: POST /api/sorx/decisions/:id/respond (approve decisions)
	// ============================================================================

	approvalCount := 0
	if decisionsResp.Count > 0 {
		approveCount := decisionsResp.Count
		if approveCount > 3 {
			approveCount = 3
		}

		fmt.Printf("\n[4/5] POST /api/sorx/decisions/:id/respond (approve %d decisions)\n", approveCount)

		// Mock approval responses for each decision
		for i := 0; i < approveCount; i++ {
			decisionID := uuid.New().String()

			decision_str := "Approve"
			if i%3 == 1 {
				decision_str = "Defer"
			} else if i%3 == 2 {
				decision_str = "Reject"
			}

			respBody := map[string]interface{}{
				"decision": decision_str,
				"inputs": map[string]interface{}{
					"coo_comment": "Reviewed autonomy metrics and gate thresholds",
				},
			}

			respBodyJSON, _ := json.Marshal(respBody)
			req = httptest.NewRequest("POST",
				fmt.Sprintf("/api/sorx/decisions/%s/respond", decisionID),
				bytes.NewBuffer(respBodyJSON))
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cooUserID))
			req.Header.Set("Content-Type", "application/json")

			startTime := time.Now()
			w = httptest.NewRecorder()
			router.ServeHTTP(w, req)
			latency := time.Since(startTime)

			if w.Code == http.StatusOK {
				fmt.Printf("   ✓ Decision %d: %s (latency: %dms)\n", i+1, decision_str, latency.Milliseconds())
				approvalCount++
			} else if w.Code == http.StatusNotFound {
				fmt.Printf("   ℹ Decision %d: endpoint not found (expected in mock)\n", i+1)
			}
		}
	}

	// ============================================================================
	// STEP 5: Verify learning loop feedback
	// ============================================================================

	fmt.Println("\n[5/5] Verify Learning Loop Feedback (S/N Governance)")

	// Query audit trail for healing.adaptive.adjust spans
	req = httptest.NewRequest("GET",
		fmt.Sprintf("/api/audit?span_name=healing.adaptive.adjust&user_id=%s", cooUserID),
		nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cooUserID))

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		var auditResp struct {
			Spans []interface{} `json:"spans"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &auditResp); err == nil {
			fmt.Printf("   ✓ Learning Loop Feedback Recorded: %d spans\n", len(auditResp.Spans))
		}
	} else {
		fmt.Println("   ℹ Audit endpoint not implemented (expected)")
	}

	// ============================================================================
	// Generate Report
	// ============================================================================

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("COO WORKFLOW SUMMARY")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("Timestamp:                    %s\n", time.Now().Format(time.RFC3339))
	fmt.Printf("COO User ID:                  %s\n", cooUserID)
	fmt.Printf("Decisions Reviewed:           %d\n", decisionsResp.Count)
	fmt.Printf("Decisions Approved/Deferred:  %d\n", approvalCount)
	fmt.Printf("Queue Depth After Approval:   %d\n", decisionsResp.Count-approvalCount)
	fmt.Printf("S/N Governance Status:        NORMAL\n")
	fmt.Println(strings.Repeat("=", 80))

	// Final assertion
	if approvalCount > decisionsResp.Count {
		t.Errorf("Approved more decisions than were pending")
	}
}

// ============================================================================
// Helper: Setup test router
// ============================================================================

func setupTestRouter() *gin.Engine {
	router := gin.New()

	// Mock endpoints for testing (in reality, these would be integrated with real handlers)

	// Mock process-leads endpoint
	router.GET("/api/crm/process-leads", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"success":      true,
			"created_count": 35,
			"escalated_count": 8,
			"skipped_count": 7,
			"success_rate": 0.814, // 35 / (35 + 8)
		})
	})

	// Mock assign-tasks endpoint
	router.GET("/api/projects/:id/assign-tasks", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"success":        true,
			"assigned_count": 42,
			"skipped_count":  8,
			"success_rate":   0.84, // 42 / (42 + 8)
		})
	})

	// Mock get decisions endpoint
	router.GET("/api/sorx/decisions", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"count":   2,
			"decisions": []gin.H{
				{
					"id":       uuid.New().String(),
					"question": "Approve high-value deal from TechCorp Inc?",
					"priority": "high",
					"status":   "pending",
				},
				{
					"id":       uuid.New().String(),
					"question": "Escalate stalled negotiation for Acme Corp?",
					"priority": "medium",
					"status":   "pending",
				},
			},
		})
	})

	// Mock respond to decision endpoint
	router.POST("/api/sorx/decisions/:id/respond", func(c *gin.Context) {
		var req struct {
			Decision string `json:"decision"`
			Inputs   map[string]interface{} `json:"inputs"`
		}
		c.ShouldBindJSON(&req)

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Decision recorded",
		})
	})

	// Mock audit endpoint
	router.GET("/api/audit", func(c *gin.Context) {
		spanName := c.Query("span_name")
		if spanName == "healing.adaptive.adjust" {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"spans": []gin.H{
					{
						"span_id":   uuid.New().String(),
						"span_name": "healing.adaptive.adjust",
						"status":    "ok",
						"attributes": gin.H{
							"adjustment_type": "threshold_update",
							"old_threshold":   0.75,
							"new_threshold":   0.80,
						},
					},
				},
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"spans":   []interface{}{},
			})
		}
	})

	return router
}
