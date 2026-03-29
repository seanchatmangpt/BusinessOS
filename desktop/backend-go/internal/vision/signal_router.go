package vision

// signal_router.go aggregates health + metrics from all 4 projects and returns
// unified Vision 2030 status, wrapping responses in a Signal Theory S=(M,G,T,F,W)
// envelope.
//
// GET /api/vision/status
//
// WvdA: 3s per-probe timeout, 30s total context deadline (deadlock freedom).
// Armstrong: errors surface visibly per service; no silent swallowing.

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	// probeTimeout is the per-service health probe deadline.
	// WvdA: bounded operation prevents deadlock on a single slow service.
	probeTimeout = 3 * time.Second

	// totalTimeout is the overall deadline for the entire vision status request.
	// WvdA: bounded operation ensures the handler always returns.
	totalTimeout = 30 * time.Second
)

// SignalEnvelope encodes Signal Theory S=(M,G,T,F,W) metadata on every response.
type SignalEnvelope struct {
	Mode      string `json:"mode"`      // "data"
	Genre     string `json:"genre"`     // "status"
	Type      string `json:"type"`      // "inform"
	Format    string `json:"format"`    // "json"
	Structure string `json:"structure"` // "vision-status"
}

// ServiceStatus reports the health of a single service in the integration chain.
type ServiceStatus struct {
	Name    string `json:"name"`
	Port    int    `json:"port"`
	Healthy bool   `json:"healthy"`
	Latency int64  `json:"latency_ms"`
}

// VisionStatus is the top-level response for GET /api/vision/status.
type VisionStatus struct {
	Signal    SignalEnvelope  `json:"signal"`
	Services  []ServiceStatus `json:"services"`
	AllUp     bool            `json:"all_up"`
	Timestamp string          `json:"timestamp"`
}

// serviceSpec defines the name, port, and health endpoint for one service.
type serviceSpec struct {
	Name     string
	Port     int
	HealthURL string
}

// defaultServices returns the 4-project integration chain service specs.
func defaultServices() []serviceSpec {
	return []serviceSpec{
		{Name: "pm4py-rust", Port: 8090, HealthURL: "http://localhost:8090/api/health"},
		{Name: "BusinessOS", Port: 8001, HealthURL: "http://localhost:8001/healthz"},
		{Name: "OSA", Port: 8089, HealthURL: "http://localhost:8089/health"},
		{Name: "Canopy", Port: 9089, HealthURL: "http://localhost:9089/health"},
	}
}

// HealthProber is the interface for probing a single service's health endpoint.
// Tests can substitute a fake implementation to avoid live HTTP calls.
type HealthProber interface {
	Probe(ctx context.Context, url string) (healthy bool, latencyMs int64, err error)
}

// HTTPProber probes service health via real HTTP GET requests.
type HTTPProber struct {
	client *http.Client
}

// NewHTTPProber constructs a prober with per-request timeout enforced via context.
func NewHTTPProber() *HTTPProber {
	return &HTTPProber{
		client: &http.Client{
			// Transport-level timeout as a safety net; primary timeout is via context.
			Timeout: probeTimeout + 1*time.Second,
		},
	}
}

// Probe sends GET to the given URL and reports healthy (2xx) plus latency.
func (p *HTTPProber) Probe(ctx context.Context, url string) (bool, int64, error) {
	// WvdA: per-probe context deadline prevents deadlock on one slow service.
	probeCtx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(probeCtx, http.MethodGet, url, nil)
	if err != nil {
		return false, 0, fmt.Errorf("create request: %w", err)
	}

	start := time.Now()
	resp, err := p.client.Do(req)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return false, latency, err
	}
	defer resp.Body.Close()

	healthy := resp.StatusCode >= 200 && resp.StatusCode < 300
	return healthy, latency, nil
}

// SignalRouter aggregates health from all 4 projects into a Vision 2030 status.
type SignalRouter struct {
	prober   HealthProber
	services []serviceSpec
	logger   *slog.Logger
}

// NewSignalRouter constructs a router with the default HTTP prober and service list.
func NewSignalRouter() *SignalRouter {
	return &SignalRouter{
		prober:   NewHTTPProber(),
		services: defaultServices(),
		logger:   slog.Default(),
	}
}

// NewSignalRouterWithProber constructs a router with a caller-supplied prober.
// Use this in tests to inject a fake prober that does not make real HTTP calls.
func NewSignalRouterWithProber(prober HealthProber, services []serviceSpec) *SignalRouter {
	if services == nil {
		services = defaultServices()
	}
	return &SignalRouter{
		prober:   prober,
		services: services,
		logger:   slog.Default(),
	}
}

// HandleVisionStatus handles GET /api/vision/status.
// It concurrently probes all services and returns a VisionStatus with Signal envelope.
func (sr *SignalRouter) HandleVisionStatus(c *gin.Context) {
	// WvdA: total deadline covers all concurrent probes.
	ctx, cancel := context.WithTimeout(c.Request.Context(), totalTimeout)
	defer cancel()

	status := sr.ProbeAll(ctx)
	c.JSON(http.StatusOK, status)
}

// ProbeAll concurrently probes all services and returns the aggregated VisionStatus.
// Exported for direct use in tests without going through Gin.
func (sr *SignalRouter) ProbeAll(ctx context.Context) VisionStatus {
	results := make([]ServiceStatus, len(sr.services))
	var wg sync.WaitGroup
	wg.Add(len(sr.services))

	for i, svc := range sr.services {
		go func(idx int, spec serviceSpec) {
			defer wg.Done()

			healthy, latency, err := sr.prober.Probe(ctx, spec.HealthURL)
			if err != nil {
				sr.logger.WarnContext(ctx, "vision probe failed",
					"service", spec.Name,
					"port", spec.Port,
					"error", err,
				)
			}

			results[idx] = ServiceStatus{
				Name:    spec.Name,
				Port:    spec.Port,
				Healthy: healthy,
				Latency: latency,
			}
		}(i, svc)
	}

	wg.Wait()

	allUp := true
	for _, r := range results {
		if !r.Healthy {
			allUp = false
			break
		}
	}

	return VisionStatus{
		Signal: SignalEnvelope{
			Mode:      "data",
			Genre:     "status",
			Type:      "inform",
			Format:    "json",
			Structure: "vision-status",
		},
		Services:  results,
		AllUp:     allUp,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

// RegisterRoutes attaches the vision status endpoint to the given router group.
// Call with the /api group: sr.RegisterRoutes(api)
func (sr *SignalRouter) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/vision/status", sr.HandleVisionStatus)
}
