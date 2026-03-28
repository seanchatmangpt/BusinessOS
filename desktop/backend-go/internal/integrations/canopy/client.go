// Package canopy provides a fire-and-forget client for pushing board
// intelligence from BusinessOS to the Canopy backend.
//
// Architecture constraint: This client is invoked after each successful L0
// sync as a detached goroutine.  Failures are logged at Warn level and never
// propagate to the caller (Armstrong: degraded state, not crash).
package canopy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	canopyAdapterCallSpan = "canopy.adapter_call"
	intelligencePath      = "/api/v1/bos/intelligence"
)

// Client pushes board intelligence payloads to the Canopy backend.
type Client struct {
	cfg        *Config
	httpClient *http.Client
	tracer     trace.Tracer
}

// NewClient creates a Canopy intelligence client.
// Returns nil when cfg is nil (i.e., env vars not set).
func NewClient(cfg *Config) *Client {
	return NewClientWithConfig(cfg)
}

// NewClientWithConfig creates a client with a pre-built Config (used in tests).
func NewClientWithConfig(cfg *Config) *Client {
	if cfg == nil {
		return nil
	}
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
			// otelhttp transport auto-injects W3C traceparent header.
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		},
		tracer: otel.Tracer("businessos.canopy"),
	}
}

// PushIntelligence implements CanopyIntelligencePusher.
// Posts caseCount and handoffCount as a board intelligence payload to Canopy.
// WvdA: caller must pass a context with deadline (10s recommended).
func (c *Client) PushIntelligence(ctx context.Context, caseCount, handoffCount int) error {
	if c == nil {
		return nil
	}

	correlationID := os.Getenv("CHATMANGPT_CORRELATION_ID")

	payload := IntelligencePayload{
		HealthSummary:    1.0, // BusinessOS always reports healthy when syncing
		ConformanceScore: 1.0,
		TopRisk:          "none",
		ConwayViolations: 0,
		CaseCount:        caseCount,
		HandoffCount:     handoffCount,
		Source:           "business_os",
	}

	ctx, span := c.tracer.Start(ctx, canopyAdapterCallSpan,
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	span.SetAttributes(
		attribute.String("canopy.adapter.name", "bos_intelligence_push"),
		attribute.String("canopy.adapter.type", "business_os"),
		attribute.String("chatmangpt.run.correlation_id", correlationID),
	)

	body, err := json.Marshal(payload)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("canopy: marshal payload: %w", err)
	}

	var lastErr error
	backoffs := []time.Duration{500 * time.Millisecond, time.Second}
	for attempt := 0; attempt <= c.cfg.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-time.After(backoffs[min(attempt-1, len(backoffs)-1)]):
			case <-ctx.Done():
				span.RecordError(ctx.Err())
				span.SetStatus(codes.Error, "context cancelled during retry")
				return ctx.Err()
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost,
			c.cfg.BaseURL+intelligencePath,
			bytes.NewReader(body),
		)
		if err != nil {
			lastErr = err
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-BOS-Secret", c.cfg.SharedSecret)
		if correlationID != "" {
			req.Header.Set("X-Correlation-ID", correlationID)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			slog.Warn("canopy intelligence push failed", "attempt", attempt+1, "error", err)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			span.SetStatus(codes.Ok, "")
			return nil
		}
		lastErr = fmt.Errorf("canopy returned HTTP %d", resp.StatusCode)
		slog.Warn("canopy intelligence push non-2xx", "attempt", attempt+1, "status", resp.StatusCode)
	}

	span.RecordError(lastErr)
	span.SetStatus(codes.Error, lastErr.Error())
	return fmt.Errorf("canopy push failed after %d attempts: %w", c.cfg.MaxRetries+1, lastErr)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
