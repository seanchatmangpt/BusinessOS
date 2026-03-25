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
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.25.0"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
)

const (
	// DefaultOTLPEndpoint is the default OTLP HTTP endpoint for the Weaver receiver.
	DefaultOTLPEndpoint = "localhost:4318"

	// EnvWeaverLiveCheck enables Weaver live-check when set to "true".
	EnvWeaverLiveCheck = "WEAVER_LIVE_CHECK"

	// EnvWeaverOTLPEndpoint overrides the default OTLP endpoint.
	EnvWeaverOTLPEndpoint = "WEAVER_OTLP_ENDPOINT"

	// defaultShutdownTimeout is the maximum time to wait for span flush on shutdown.
	defaultShutdownTimeout = 10 * time.Second
)

// SetupWeaverLiveCheck configures OTEL to export spans to the Weaver live-check
// OTLP receiver. It creates a TracerProvider with the "businessos" service
// resource and sets it as the global provider.
//
// Call the returned shutdown function in TestMain's deferred cleanup to flush
// any pending spans before the test binary exits.
//
// The endpoint defaults to localhost:4318 (OTLP HTTP) but can be overridden
// via the WEAVER_OTLP_ENDPOINT environment variable. Note: unlike gRPC OTLP
// on port 4317, the HTTP exporter uses port 4318 by default.
func SetupWeaverLiveCheck(ctx context.Context) (func(context.Context) error, error) {
	endpoint := os.Getenv(EnvWeaverOTLPEndpoint)
	if endpoint == "" {
		endpoint = DefaultOTLPEndpoint
	}

	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("weaver: create OTLP HTTP exporter: %w", err)
	}

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("businessos"),
		semconv.ServiceVersion("1.0.0"),
		semconv.DeploymentEnvironment("weaver-live-check"),
	)

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
