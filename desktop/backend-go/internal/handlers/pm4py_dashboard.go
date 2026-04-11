package handlers

// PM4PyDashboardHandler aggregates process mining KPIs from pm4py-mcp in a single call.
//
// POST /api/pm4py/dashboard-kpi
//
// Calls pm4py-mcp statistics and conformance tools concurrently via MCP protocol,
// then merges results into a single KPI response for the 4 new KPI dashboard widgets.
//
// WvdA: both concurrent calls have a shared 30s context timeout (deadlock freedom).
// Armstrong: stats error surfaces as 502 — no silent swallow.

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/services"
)

const pm4pyDashboardTimeout = 30 * time.Second

// ProcessMiningKPIRequest accepts raw event log JSON and an optional petri net.
type ProcessMiningKPIRequest struct {
	EventLog json.RawMessage `json:"event_log" binding:"required"`
	PetriNet json.RawMessage `json:"petri_net,omitempty"`
}

// ProcessMiningKPIResponse aggregates all process mining KPIs for dashboard widgets.
type ProcessMiningKPIResponse struct {
	// Conformance KPIs (widget 1 + 2)
	ConformanceFitness   float64 `json:"conformance_fitness"`
	ConformancePrecision float64 `json:"conformance_precision"`
	IsConformant         bool    `json:"is_conformant"`

	// Variant KPIs (widget 3)
	VariantCount int            `json:"variant_count"`
	TopVariants  []VariantEntry `json:"top_variants"`

	// Activity / bottleneck KPIs (widget 4)
	BottleneckActivities []BottleneckEntry `json:"bottleneck_activities"`
	ActivityFrequencies  map[string]int    `json:"activity_frequencies"`

	// Event log metadata
	EventCount int `json:"event_count"`
	TraceCount int `json:"trace_count"`

	// Metadata
	FetchedAt string `json:"fetched_at"`
}

// VariantEntry describes one process variant and its relative frequency.
type VariantEntry struct {
	Label      string  `json:"label"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

// BottleneckEntry describes one activity identified as a bottleneck.
type BottleneckEntry struct {
	Activity  string `json:"activity"`
	Frequency int    `json:"frequency"`
}

// PM4PyDashboardHandler handles POST /api/pm4py/dashboard-kpi.
type PM4PyDashboardHandler struct {
	client  services.PM4PyMCPClientInterface
	logger  *slog.Logger
	timeout time.Duration
}

// NewPM4PyDashboardHandler constructs a handler using a pm4py-mcp client.
func NewPM4PyDashboardHandler(client services.PM4PyMCPClientInterface) *PM4PyDashboardHandler {
	return &PM4PyDashboardHandler{
		client:  client,
		logger:  slog.Default(),
		timeout: pm4pyDashboardTimeout,
	}
}

// NewPM4PyDashboardHandlerWithTimeout constructs a handler with a caller-supplied timeout.
// Use this in tests or when the deployment environment requires a non-default deadline.
func NewPM4PyDashboardHandlerWithTimeout(client services.PM4PyMCPClientInterface, timeout time.Duration) *PM4PyDashboardHandler {
	return &PM4PyDashboardHandler{
		client:  client,
		logger:  slog.Default(),
		timeout: timeout,
	}
}

// GetDashboardKPI handles POST /api/pm4py/dashboard-kpi.
// It fans out to pm4py-mcp statistics + conformance concurrently, then merges results.
func (h *PM4PyDashboardHandler) GetDashboardKPI(c *gin.Context) {
	var req ProcessMiningKPIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "event_log is required"})
		return
	}

	// WvdA: shared deadline covers both goroutines.
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.timeout)
	defer cancel()

	var (
		statsResult *services.StatisticsResult
		statsErr    error
		confResult  *services.ConformanceResult
		confErr     error
		wg          sync.WaitGroup
	)

	wg.Add(2)

	go func() {
		defer wg.Done()
		statsResult, statsErr = h.callStatistics(ctx, req.EventLog)
	}()

	go func() {
		defer wg.Done()
		confResult, confErr = h.callConformance(ctx, req.EventLog, req.PetriNet)
	}()

	wg.Wait()

	// Armstrong: surface stats failure visibly rather than returning partial/corrupt data.
	if statsErr != nil {
		h.logger.ErrorContext(ctx, "pm4py dashboard: statistics call failed", "error", statsErr)
		c.JSON(http.StatusBadGateway, gin.H{"error": "pm4py-mcp statistics unavailable"})
		return
	}

	// Conformance errors are non-fatal (petri net may be absent or pm4py unavailable).
	if confErr != nil {
		h.logger.WarnContext(ctx, "pm4py dashboard: conformance call failed (non-fatal)", "error", confErr)
	}

	resp := h.buildResponse(statsResult, confResult)
	c.JSON(http.StatusOK, resp)
}

// callStatistics requests statistics from pm4py-mcp.
// Returns nil+error on failure; caller decides whether to surface as 502.
func (h *PM4PyDashboardHandler) callStatistics(ctx context.Context, eventLog json.RawMessage) (*services.StatisticsResult, error) {
	if h.client == nil {
		return nil, nil
	}
	return h.client.Statistics(ctx, eventLog)
}

// callConformance requests conformance from pm4py-mcp.
// Returns nil+nil when no petri net is supplied (graceful degradation).
func (h *PM4PyDashboardHandler) callConformance(ctx context.Context, eventLog, petriNet json.RawMessage) (*services.ConformanceResult, error) {
	if h.client == nil {
		return nil, nil
	}
	if len(petriNet) == 0 || string(petriNet) == "null" || string(petriNet) == "" {
		return nil, nil // No petri net — skip conformance gracefully.
	}

	return h.client.Conformance(ctx, eventLog, petriNet)
}

// buildResponse merges statistics and (optional) conformance into one KPI payload.
func (h *PM4PyDashboardHandler) buildResponse(
	stats *services.StatisticsResult,
	conf *services.ConformanceResult,
) ProcessMiningKPIResponse {

	resp := ProcessMiningKPIResponse{
		TopVariants:          []VariantEntry{},
		BottleneckActivities: []BottleneckEntry{},
		ActivityFrequencies:  map[string]int{},
		FetchedAt:            time.Now().UTC().Format(time.RFC3339),
	}

	if stats != nil {
		resp.EventCount = stats.TraceEvents
		resp.TraceCount = stats.UniqueTraces
		resp.VariantCount = len(stats.Variants)

		// Build activity frequencies map from structured data.
		for _, af := range stats.ActivityFrequency {
			resp.ActivityFrequencies[af.Activity] = af.Frequency
		}

		// Map variant data.
		for _, v := range stats.Variants {
			resp.TopVariants = append(resp.TopVariants, VariantEntry{
				Label:      v.Variant,
				Count:      v.Count,
				Percentage: v.Percentage,
			})
		}

		// Map bottleneck activities.
		for _, bn := range stats.Bottleneck {
			freq := 0
			if resp.ActivityFrequencies != nil {
				freq = resp.ActivityFrequencies[bn.Activity]
			}
			resp.BottleneckActivities = append(resp.BottleneckActivities, BottleneckEntry{
				Activity:  bn.Activity,
				Frequency: freq,
			})
		}
	}

	if conf != nil {
		resp.ConformanceFitness = conf.Fitness
		resp.ConformancePrecision = conf.Precision
		resp.IsConformant = conf.Fitness >= 1.0
	}

	return resp
}
