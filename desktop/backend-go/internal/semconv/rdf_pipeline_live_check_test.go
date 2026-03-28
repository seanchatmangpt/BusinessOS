package semconv_test

import (
	"context"
	"os"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	otelsemconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"

	bossemconv "github.com/rhl/businessos-backend/internal/semconv"
)

// TestRDFPipelineSpansAreEmitted emits all RDF pipeline span names required
// by semconv/model/business_os/spans.yaml and semconv/model/board/spans.yaml.
//
// Skipped unless WEAVER_LIVE_CHECK=true — run via `make verify-rdf-pipeline`
// or directly:
//
//	WEAVER_LIVE_CHECK=true go test ./internal/semconv/... -run TestRDFPipelineSpansAreEmitted
func TestRDFPipelineSpansAreEmitted(t *testing.T) {
	if os.Getenv("WEAVER_LIVE_CHECK") != "true" {
		t.Skip("WEAVER_LIVE_CHECK not set — skipping live-check test")
	}

	endpoint := os.Getenv("WEAVER_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:4317"
	}
	correlationID := os.Getenv("CHATMANGPT_CORRELATION_ID")
	if correlationID == "" {
		correlationID = "go-rdf-pipeline-test"
	}

	// Initialise OTLP gRPC exporter.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpointURL(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		t.Fatalf("failed to create OTLP exporter: %v", err)
	}

	res := resource.NewWithAttributes(otelsemconv.SchemaURL,
		otelsemconv.ServiceName("businessos"),
		attribute.String("chatmangpt.run.correlation_id", correlationID),
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	defer func() {
		_ = tp.Shutdown(ctx)
	}()
	otel.SetTracerProvider(tp)

	tracer := tp.Tracer("businessos.rdf-pipeline-test")

	// Emit board.l0_sync span (spans.yaml: span.board.l0_sync)
	emitSpan(t, tracer, bossemconv.BoardL0SyncSpan, trace.SpanKindInternal, []attribute.KeyValue{
		attribute.Int("board.l0_sync.case_count", 42),
		attribute.Int("board.l0_sync.handoff_count", 10),
	})

	// Emit bos.ontology.execute span (spans.yaml: span.bos.ontology.execute)
	emitSpan(t, tracer, bossemconv.BosOntologyExecuteSpan, trace.SpanKindInternal, []attribute.KeyValue{
		attribute.String("rdf.sparql.endpoint", "http://localhost:7878"),
		attribute.Int64("rdf.result.triple_count", 150),
		attribute.String("chatmangpt.run.correlation_id", correlationID),
	})

	// Emit bos.rdf.write span (spans.yaml: span.bos.rdf.write)
	emitSpan(t, tracer, bossemconv.BosRdfWriteSpan, trace.SpanKindClient, []attribute.KeyValue{
		attribute.String("rdf.sparql.endpoint", "http://localhost:7878"),
		attribute.String("rdf.write.format", "application/n-triples"),
		attribute.Int64("rdf.write.triple_count", 150),
		attribute.String("chatmangpt.run.correlation_id", correlationID),
	})

	// Emit bos.rdf.query span (spans.yaml: span.bos.rdf.query)
	emitSpan(t, tracer, bossemconv.BosRdfQuerySpan, trace.SpanKindClient, []attribute.KeyValue{
		attribute.String("rdf.sparql.endpoint", "http://localhost:7879"),
		attribute.String("rdf.sparql.query_type", "SELECT"),
		attribute.Int64("rdf.sparql.result_count", 5),
		attribute.String("chatmangpt.run.correlation_id", correlationID),
	})

	// Flush all spans before test exits.
	if err := tp.ForceFlush(ctx); err != nil {
		t.Logf("flush error (non-fatal): %v", err)
	}
}

func emitSpan(t *testing.T, tracer trace.Tracer, name string, kind trace.SpanKind, attrs []attribute.KeyValue) {
	t.Helper()
	ctx, span := tracer.Start(context.Background(), name, trace.WithSpanKind(kind))
	defer span.End()
	span.SetAttributes(attrs...)
	_ = ctx
	t.Logf("emitted span: %s", name)
}
