package observability

import (
	"context"
	"testing"
	"time"
)

func TestTelemetry_InitTracer(t *testing.T) {
	tel := New()
	err := tel.InitTracer()

	if err != nil {
		t.Fatalf("InitTracer failed: %v", err)
	}

	if tel.spans == nil {
		t.Error("spans map is nil")
	}
	if tel.metrics == nil {
		t.Error("metrics slice is nil")
	}
}

func TestTelemetry_StartSpan(t *testing.T) {
	tel := New()
	tel.InitTracer()

	ctx := context.Background()
	attributes := map[string]any{
		"request_id": "req-123",
		"endpoint":   "/api/foo",
	}

	span, newCtx := tel.StartSpan(ctx, "request.process", attributes)

	if span.SpanID == "" {
		t.Error("span ID is empty")
	}
	if span.TraceID == "" {
		t.Error("trace ID is empty")
	}
	if span.SpanName != "request.process" {
		t.Errorf("expected span name 'request.process', got %q", span.SpanName)
	}
	if span.Status != "active" {
		t.Errorf("expected status 'active', got %q", span.Status)
	}
	if span.StartTimeUs <= 0 {
		t.Error("start time is invalid")
	}

	// Verify attributes are enriched
	if _, ok := span.Attributes["timestamp"]; !ok {
		t.Error("timestamp not enriched in attributes")
	}

	// Verify new context has span ID
	if newCtx.Value("span_id") != span.SpanID {
		t.Error("context not updated with span ID")
	}
	if newCtx.Value("trace_id") != span.TraceID {
		t.Error("context not updated with trace ID")
	}

	// Verify context was returned (not modified in place)
	if newCtx == ctx {
		t.Error("expected new context, got same context")
	}
}

func TestTelemetry_EndSpan(t *testing.T) {
	tel := New()
	tel.InitTracer()

	ctx := context.Background()
	span, _ := tel.StartSpan(ctx, "request.process", map[string]any{})

	time.Sleep(10 * time.Millisecond)
	tel.EndSpan(span, "ok", "")

	if span.Status != "ok" {
		t.Errorf("expected status 'ok', got %q", span.Status)
	}
	if span.EndTimeUs <= span.StartTimeUs {
		t.Error("end time should be after start time")
	}
	if span.EndTimeUs == 0 {
		t.Error("end time is zero")
	}
}

func TestTelemetry_EndSpan_WithError(t *testing.T) {
	tel := New()
	tel.InitTracer()

	ctx := context.Background()
	span, _ := tel.StartSpan(ctx, "auth.check", map[string]any{})

	tel.EndSpan(span, "error", "unauthorized")

	if span.Status != "error" {
		t.Errorf("expected status 'error', got %q", span.Status)
	}
	if span.ErrorMessage != "unauthorized" {
		t.Errorf("expected error message 'unauthorized', got %q", span.ErrorMessage)
	}
}

func TestTelemetry_RecordMetric(t *testing.T) {
	tel := New()
	tel.InitTracer()

	tel.RecordMetric("request.latency_ms", 234.5, map[string]string{
		"endpoint": "/api/foo",
		"status":   "200",
	}, "histogram")

	metrics := tel.GetMetrics()
	if len(metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(metrics))
	}

	metric := metrics[0]
	if metric.Name != "request.latency_ms" {
		t.Errorf("expected name 'request.latency_ms', got %q", metric.Name)
	}
	if metric.Value != 234.5 {
		t.Errorf("expected value 234.5, got %f", metric.Value)
	}
	if metric.Type != "histogram" {
		t.Errorf("expected type 'histogram', got %q", metric.Type)
	}
	if metric.Dimensions["endpoint"] != "/api/foo" {
		t.Error("dimension 'endpoint' not recorded")
	}
}

func TestTelemetry_RecordMetric_Counter(t *testing.T) {
	tel := New()
	tel.InitTracer()

	tel.RecordMetric("auth.failures", 1, map[string]string{
		"user_id": "user-123",
	}, "counter")

	tel.RecordMetric("auth.failures", 1, map[string]string{
		"user_id": "user-123",
	}, "counter")

	metrics := tel.GetMetrics()
	if len(metrics) != 2 {
		t.Fatalf("expected 2 metrics, got %d", len(metrics))
	}

	// Verify second metric is independent (no aggregation in this implementation)
	if metrics[1].Value != 1 {
		t.Errorf("expected second metric value 1, got %f", metrics[1].Value)
	}
}

func TestTelemetry_PropagateTraceContext(t *testing.T) {
	tel := New()
	tel.InitTracer()

	ctx := context.Background()

	// Propagate trace context from upstream header
	newCtx := tel.PropagateTraceContext(ctx, "trace-123", "span-456")

	if newCtx.Value("trace_id") != "trace-123" {
		t.Errorf("expected trace ID 'trace-123', got %v", newCtx.Value("trace_id"))
	}
	if newCtx.Value("parent_span_id") != "span-456" {
		t.Errorf("expected parent span ID 'span-456', got %v", newCtx.Value("parent_span_id"))
	}
}

func TestTelemetry_TraceIDPropagation(t *testing.T) {
	tel := New()
	tel.InitTracer()

	ctx := context.Background()

	// Create parent span
	span1, ctx1 := tel.StartSpan(ctx, "request.start", map[string]any{})
	traceID1 := span1.TraceID

	// Create child span - should inherit trace ID
	span2, _ := tel.StartSpan(ctx1, "request.process", map[string]any{})
	traceID2 := span2.TraceID

	if traceID1 != traceID2 {
		t.Errorf("expected trace IDs to match: %s vs %s", traceID1, traceID2)
	}
	if span2.ParentSpanID != span1.SpanID {
		t.Errorf("expected parent span ID to be %s, got %s", span1.SpanID, span2.ParentSpanID)
	}
}

func TestTelemetry_GetSpans(t *testing.T) {
	tel := New()
	tel.InitTracer()

	ctx := context.Background()
	span1, ctx1 := tel.StartSpan(ctx, "request.start", map[string]any{})
	span2, _ := tel.StartSpan(ctx1, "request.process", map[string]any{})

	spans := tel.GetSpans()
	if len(spans) != 2 {
		t.Fatalf("expected 2 spans, got %d", len(spans))
	}

	spanIDs := make(map[string]bool)
	for _, span := range spans {
		spanIDs[span.SpanID] = true
	}

	if !spanIDs[span1.SpanID] {
		t.Error("span1 not found in GetSpans")
	}
	if !spanIDs[span2.SpanID] {
		t.Error("span2 not found in GetSpans")
	}
}

func TestTelemetry_GetMetrics(t *testing.T) {
	tel := New()
	tel.InitTracer()

	tel.RecordMetric("request.latency_ms", 100, map[string]string{}, "histogram")
	tel.RecordMetric("request.count", 1, map[string]string{}, "counter")

	metrics := tel.GetMetrics()
	if len(metrics) != 2 {
		t.Fatalf("expected 2 metrics, got %d", len(metrics))
	}

	if metrics[0].Name != "request.latency_ms" {
		t.Error("first metric not found")
	}
	if metrics[1].Name != "request.count" {
		t.Error("second metric not found")
	}
}

func TestTelemetry_GetSpan(t *testing.T) {
	tel := New()
	tel.InitTracer()

	ctx := context.Background()
	span, _ := tel.StartSpan(ctx, "request.process", map[string]any{})

	retrieved := tel.GetSpan(span.SpanID)
	if retrieved == nil {
		t.Fatal("GetSpan returned nil")
	}
	if retrieved.SpanID != span.SpanID {
		t.Errorf("expected span ID %s, got %s", span.SpanID, retrieved.SpanID)
	}
}

func TestTelemetry_Clear(t *testing.T) {
	tel := New()
	tel.InitTracer()

	ctx := context.Background()
	tel.StartSpan(ctx, "request.start", map[string]any{})
	tel.RecordMetric("request.latency_ms", 100, map[string]string{}, "histogram")

	tel.Clear()

	if len(tel.GetSpans()) != 0 {
		t.Error("spans not cleared")
	}
	if len(tel.GetMetrics()) != 0 {
		t.Error("metrics not cleared")
	}
}
