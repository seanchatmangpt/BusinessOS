package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/codes"
)

// yawlSimulateRequest is the body accepted by POST /api/yawl/simulate.
type yawlSimulateRequest struct {
	// SpecSet selects which specs to simulate: "basic_wcp" (default), "wcp_patterns",
	// "real_data", or "all".
	SpecSet string `json:"spec_set"`
	// UserCount is the number of concurrent simulated users (default 3).
	UserCount int `json:"user_count"`
	// TimeoutMs is the per-user budget in milliseconds (default 30000).
	TimeoutMs int `json:"timeout_ms"`
	// MaxSteps is the maximum drain-loop iterations per user (default 50).
	MaxSteps int `json:"max_steps"`
	// MaxConcurrency caps Task.async_stream concurrency in OSA (default 10).
	MaxConcurrency int `json:"max_concurrency"`
}

// yawlUserResult mirrors OSA's Simulator.UserResult in JSON form.
type yawlUserResult struct {
	UserID         int    `json:"user_id"`
	CaseID         string `json:"case_id"`
	SpecID         string `json:"spec_id"`
	Status         string `json:"status"`
	StepsCompleted *int   `json:"steps_completed"`
	DurationMs     *int   `json:"duration_ms"`
	Error          string `json:"error,omitempty"`
}

// yawlSimulateResponse is the JSON returned by POST /api/yawl/simulate.
type yawlSimulateResponse struct {
	SpecSet        string           `json:"spec_set"`
	UserCount      int              `json:"user_count"`
	TotalDurationMs int             `json:"total_duration_ms"`
	CompletedCount int              `json:"completed_count"`
	ErrorCount     int              `json:"error_count"`
	TimeoutCount   int              `json:"timeout_count"`
	Summary        string           `json:"summary"`
	Results        []yawlUserResult `json:"results"`
}

// SimulateWorkflows handles POST /api/yawl/simulate.
//
// Forwards the request to the OSA engine's /api/v1/yawl/simulate endpoint and
// returns the aggregated SimulationResult.  When OSA is unreachable the handler
// returns 502 Bad Gateway so the caller knows the engine is down, not the
// business logic.
//
// OSA URL is read from OSA_URL env var (default: http://localhost:8089).
func (h *YawlHandler) SimulateWorkflows(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "yawl.simulate")
	defer span.End()

	var req yawlSimulateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Bind failure means malformed JSON — supply defaults and continue.
		req = yawlSimulateRequest{}
	}

	// Apply defaults (mirroring OSA Simulator defaults).
	if req.SpecSet == "" {
		req.SpecSet = "basic_wcp"
	}
	if req.UserCount <= 0 {
		req.UserCount = 3
	}
	if req.TimeoutMs <= 0 {
		req.TimeoutMs = 30_000
	}
	if req.MaxSteps <= 0 {
		req.MaxSteps = 50
	}
	if req.MaxConcurrency <= 0 {
		req.MaxConcurrency = 10
	}

	osaURL := osaBaseURL() + "/api/v1/yawl/simulate"

	// Budget: outer HTTP timeout = per-user timeout × 2 (for queue drain + OSA overhead).
	outerTimeout := time.Duration(req.TimeoutMs)*2*time.Millisecond + 10*time.Second
	reqCtx, cancel := context.WithTimeout(ctx, outerTimeout)
	defer cancel()

	body, err := json.Marshal(map[string]any{
		"spec_set":        req.SpecSet,
		"user_count":      req.UserCount,
		"timeout_ms":      req.TimeoutMs,
		"max_steps":       req.MaxSteps,
		"max_concurrency": req.MaxConcurrency,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "marshal failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build request"})
		return
	}

	httpReq, err := http.NewRequestWithContext(reqCtx, "POST", osaURL, bytes.NewReader(body))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "request build failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build OSA request"})
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		h.logger.Warn("yawl simulate: OSA unreachable", "osa_url", osaURL, "error", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, "OSA unreachable")
		c.JSON(http.StatusBadGateway, gin.H{
			"error":   "OSA engine unreachable",
			"osa_url": osaURL,
		})
		return
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "read response failed")
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to read OSA response"})
		return
	}

	if resp.StatusCode != http.StatusOK {
		h.logger.Warn("yawl simulate: OSA returned non-200",
			"status", resp.StatusCode,
			"body", string(rawBody),
		)
		span.SetStatus(codes.Error, fmt.Sprintf("OSA returned %d", resp.StatusCode))
		c.JSON(http.StatusBadGateway, gin.H{
			"error":      "OSA simulate returned non-200",
			"osa_status": resp.StatusCode,
		})
		return
	}

	var result yawlSimulateResponse
	if err := json.Unmarshal(rawBody, &result); err != nil {
		h.logger.Error("yawl simulate: failed to decode OSA response", "error", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, "decode failed")
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to decode OSA response"})
		return
	}

	h.logger.Info("yawl simulate complete",
		slog.String("spec_set", result.SpecSet),
		slog.Int("user_count", result.UserCount),
		slog.Int("completed", result.CompletedCount),
		slog.Int("errors", result.ErrorCount),
		slog.Int("timeouts", result.TimeoutCount),
	)

	span.SetStatus(codes.Ok, "")
	c.JSON(http.StatusOK, result)
}

// osaBaseURL returns the OSA engine base URL from env (default: http://localhost:8089).
func osaBaseURL() string {
	if u := os.Getenv("OSA_URL"); u != "" {
		return u
	}
	return "http://localhost:8089"
}
