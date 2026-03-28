package observability

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// SpanContext represents a distributed trace span.
// It includes span ID, trace ID, parent span ID, and attributes.
type SpanContext struct {
	SpanID       string         `json:"span_id"`
	TraceID      string         `json:"trace_id"`
	ParentSpanID string         `json:"parent_span_id,omitempty"`
	SpanName     string         `json:"span_name"`
	Attributes   map[string]any `json:"attributes"`
	StartTimeUs  int64          `json:"start_time_us"`
	EndTimeUs    int64          `json:"end_time_us,omitempty"`
	Status       string         `json:"status"`
	ErrorMessage string         `json:"error_message,omitempty"`
}

// MetricPoint represents a single metric observation.
type MetricPoint struct {
	Name       string            `json:"name"`
	Value      float64           `json:"value"`
	Timestamp  int64             `json:"timestamp_us"`
	Dimensions map[string]string `json:"dimensions"`
	Type       string            `json:"type"` // "counter", "histogram", "gauge"
}

// Telemetry manages tracing and metrics for BusinessOS.
type Telemetry struct {
	mu       sync.RWMutex
	spans    map[string]*SpanContext
	metrics  []MetricPoint
	traceIDM sync.Map // Process-local trace ID storage (string -> string)
}

// New creates a new Telemetry instance.
func New() *Telemetry {
	return &Telemetry{
		spans:   make(map[string]*SpanContext),
		metrics: make([]MetricPoint, 0, 1000),
	}
}

// InitTracer initializes the tracer.
func (t *Telemetry) InitTracer() error {
	// Initialize trace storage
	t.mu.Lock()
	defer t.mu.Unlock()

	t.spans = make(map[string]*SpanContext)
	t.metrics = make([]MetricPoint, 0, 1000)

	return nil
}

// StartSpan creates a new span with optional attributes.
//
// Parameters:
//   - ctx: context.Context for trace context propagation
//   - spanName: name of the span, e.g. "request.process", "auth.check"
//   - attributes: optional map of span metadata
//
// Returns:
//   - SpanContext with auto-generated span ID and trace ID
//   - new context with trace context embedded
func (t *Telemetry) StartSpan(ctx context.Context, spanName string, attributes map[string]any) (*SpanContext, context.Context) {
	spanID := uuid.New().String()
	traceID := t.getOrCreateTraceID(ctx)

	var parentSpanID string
	if val := ctx.Value("span_id"); val != nil {
		if id, ok := val.(string); ok {
			parentSpanID = id
		}
	}

	startTimeUs := timeToMicroseconds(time.Now())

	// Enrich attributes with system context
	if attributes == nil {
		attributes = make(map[string]any)
	}
	attributes["timestamp"] = time.Now().UTC().Format(time.RFC3339Nano)
	attributes["version"] = "1.0.0" // TODO: inject version

	span := &SpanContext{
		SpanID:       spanID,
		TraceID:      traceID,
		ParentSpanID: parentSpanID,
		SpanName:     spanName,
		Attributes:   attributes,
		StartTimeUs:  startTimeUs,
		Status:       "active",
	}

	// Store span in map
	t.mu.Lock()
	t.spans[spanID] = span
	t.mu.Unlock()

	// Return span and context with embedded span ID for child spans
	newCtx := context.WithValue(ctx, "span_id", spanID)
	newCtx = context.WithValue(newCtx, "trace_id", traceID)

	return span, newCtx
}

// EndSpan marks a span as complete and records its duration.
//
// Parameters:
//   - span: SpanContext to end
//   - status: "ok" or "error"
//   - errorMessage: optional error message if status is "error"
func (t *Telemetry) EndSpan(span *SpanContext, status string, errorMessage string) {
	endTimeUs := timeToMicroseconds(time.Now())
	durationUs := endTimeUs - span.StartTimeUs

	span.EndTimeUs = endTimeUs
	span.Status = status
	if errorMessage != "" {
		span.ErrorMessage = errorMessage
	}

	// Update span in map
	t.mu.Lock()
	t.spans[span.SpanID] = span
	t.mu.Unlock()

	// Record latency metric
	t.RecordMetric(
		fmt.Sprintf("span.duration_us"),
		float64(durationUs),
		map[string]string{
			"span_name": span.SpanName,
			"status":    status,
		},
		"histogram",
	)
}

// RecordMetric records a metric observation with optional dimensions.
//
// Parameters:
//   - name: metric name, e.g. "request.latency_ms", "auth.failures"
//   - value: numeric value
//   - dimensions: optional tags, e.g. map[string]string{"endpoint": "/api/foo", "status": "200"}
//   - metricType: "counter", "histogram", or "gauge"
func (t *Telemetry) RecordMetric(name string, value float64, dimensions map[string]string, metricType string) {
	timestamp := timeToMicroseconds(time.Now())

	if dimensions == nil {
		dimensions = make(map[string]string)
	}

	point := MetricPoint{
		Name:       name,
		Value:      value,
		Timestamp:  timestamp,
		Dimensions: dimensions,
		Type:       metricType,
	}

	t.mu.Lock()
	t.metrics = append(t.metrics, point)
	t.mu.Unlock()
}

// PropagateTraceContext extracts trace context from headers.
// Used for incoming HTTP requests to propagate trace ID from upstream systems.
//
// Parameters:
//   - ctx: base context
//   - traceID: trace ID from incoming header (e.g. "traceparent" header)
//   - spanID: span ID from incoming header (optional)
//
// Returns:
//   - new context with propagated trace IDs
func (t *Telemetry) PropagateTraceContext(ctx context.Context, traceID string, spanID string) context.Context {
	if traceID != "" {
		t.traceIDM.Store(ctxKey(ctx), traceID)
		ctx = context.WithValue(ctx, "trace_id", traceID)
	}
	if spanID != "" {
		ctx = context.WithValue(ctx, "parent_span_id", spanID)
	}
	return ctx
}

// GetSpans returns all recorded spans.
func (t *Telemetry) GetSpans() []*SpanContext {
	t.mu.RLock()
	defer t.mu.RUnlock()

	spans := make([]*SpanContext, 0, len(t.spans))
	for _, span := range t.spans {
		spans = append(spans, span)
	}
	return spans
}

// GetMetrics returns all recorded metrics.
func (t *Telemetry) GetMetrics() []MetricPoint {
	t.mu.RLock()
	defer t.mu.RUnlock()

	metrics := make([]MetricPoint, len(t.metrics))
	copy(metrics, t.metrics)
	return metrics
}

// GetSpan retrieves a span by ID.
func (t *Telemetry) GetSpan(spanID string) *SpanContext {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.spans[spanID]
}

// Clear resets all spans and metrics (for testing).
func (t *Telemetry) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.spans = make(map[string]*SpanContext)
	t.metrics = make([]MetricPoint, 0, 1000)
	t.traceIDM.Range(func(key, value any) bool {
		t.traceIDM.Delete(key)
		return true
	})
}

// ============================================================================
// Private Functions
// ============================================================================

// getOrCreateTraceID retrieves or creates a trace ID for the context.
func (t *Telemetry) getOrCreateTraceID(ctx context.Context) string {
	// Check context for existing trace ID
	if val := ctx.Value("trace_id"); val != nil {
		if traceID, ok := val.(string); ok {
			return traceID
		}
	}

	// Check process-local storage
	key := ctxKey(ctx)
	if val, ok := t.traceIDM.Load(key); ok {
		if traceID, ok := val.(string); ok {
			return traceID
		}
	}

	// Create new trace ID
	traceID := uuid.New().String()
	t.traceIDM.Store(key, traceID)

	return traceID
}

// ctxKey generates a unique key for context (goroutine-local).
// This is a simplified approach; in production, use context values only.
func ctxKey(ctx context.Context) string {
	return fmt.Sprintf("ctx_%p", ctx)
}

// timeToMicroseconds converts time.Time to microseconds since epoch.
func timeToMicroseconds(t time.Time) int64 {
	return t.UnixNano() / 1000
}
