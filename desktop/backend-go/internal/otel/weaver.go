// Package otel provides Weaver live-check integration for BusinessOS.
//
// Weaver live-check validates that OTEL spans emitted during tests conform to
// the semconv schema. Enable by setting WEAVER_LIVE_CHECK=true in the test
// environment, which points spans at the Weaver OTLP receiver.
//
// Usage in TestMain:
//
//	func TestMain(m *testing.M) {
//	    if os.Getenv("WEAVER_LIVE_CHECK") == "true" {
//	        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	        defer cancel()
//	        shutdown, err := otel.SetupWeaverLiveCheck(ctx)
//	        if err != nil {
//	            log.Printf("weaver live-check setup failed: %v", err)
//	            os.Exit(1)
//	        }
//	        defer shutdown(context.Background())
//	    }
//	    os.Exit(m.Run())
//	}
package otel

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.25.0"
)

const (
	// DefaultOTLPEndpoint is host:port for OTLP gRPC (Weaver --otlp-grpc-port).
	DefaultOTLPEndpoint = "localhost:4317"

	// EnvWeaverLiveCheck enables Weaver live-check when set to "true".
	EnvWeaverLiveCheck = "WEAVER_LIVE_CHECK"

	// EnvWeaverOTLPEndpoint overrides the default OTLP endpoint (http://host:port or host:port).
	EnvWeaverOTLPEndpoint = "WEAVER_OTLP_ENDPOINT"

	// defaultShutdownTimeout is the maximum time to wait for span flush on shutdown.
	defaultShutdownTimeout = 10 * time.Second
)

// normalizeGRPCEndpoint strips schemes and paths so otlptracegrpc gets "host:port".
func normalizeGRPCEndpoint(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return DefaultOTLPEndpoint
	}
	raw = strings.TrimPrefix(raw, "http://")
	raw = strings.TrimPrefix(raw, "https://")
	// Drop any trailing path (e.g. rare misconfig)
	if i := strings.IndexByte(raw, '/'); i >= 0 {
		raw = raw[:i]
	}
	return raw
}

// SetupWeaverLiveCheck configures OTEL to export spans to the Weaver live-check
// OTLP gRPC receiver. It creates a TracerProvider with the "businessos" service
// resource and sets it as the global provider.
//
// Call the returned shutdown function in TestMain's deferred cleanup to flush
// any pending spans before the test binary exits.
func SetupWeaverLiveCheck(ctx context.Context) (func(context.Context) error, error) {
	endpoint := normalizeGRPCEndpoint(os.Getenv(EnvWeaverOTLPEndpoint))

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("weaver: create OTLP gRPC exporter: %w", err)
	}

	attrs := []attribute.KeyValue{
		semconv.ServiceName("businessos"),
		semconv.ServiceVersion("1.0.0"),
		semconv.ServiceNamespace("chatmangpt"),
		semconv.DeploymentEnvironment("weaver-live-check"),
	}
	if cid := strings.TrimSpace(os.Getenv("CHATMANGPT_CORRELATION_ID")); cid != "" {
		attrs = append(attrs, attribute.String("chatmangpt.run.correlation_id", cid))
	}
	res := resource.NewWithAttributes(semconv.SchemaURL, attrs...)

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exporter),
		tracesdk.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	shutdown := func(ctx context.Context) error {
		shutdownCtx, cancel := context.WithTimeout(ctx, defaultShutdownTimeout)
		defer cancel()
		if err := tp.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("weaver: tracer shutdown: %w", err)
		}
		return nil
	}

	return shutdown, nil
}

// IsLiveCheckEnabled returns true if WEAVER_LIVE_CHECK=true is set in the
// environment. Use this in TestMain to conditionally set up the tracer provider.
func IsLiveCheckEnabled() bool {
	return os.Getenv(EnvWeaverLiveCheck) == "true"
}
