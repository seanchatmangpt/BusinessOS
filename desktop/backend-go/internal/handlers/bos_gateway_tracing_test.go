package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// setupTestTracer registers an in-process SpanRecorder so tests can inspect
// emitted spans without a real OTEL collector.
func setupTestTracer(t *testing.T) *tracetest.SpanRecorder {
	t.Helper()
	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	t.Cleanup(func() {
		_ = provider.Shutdown(context.Background())
	})
	return recorder
}

// setupGatewayWithTracing creates a gateway handler + gin router wired to a
// mock MCP client, with the test tracer installed.
func setupGatewayWithTracing(t *testing.T) (*BOSGatewayHandler, *gin.Engine, *tracetest.SpanRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	recorder := setupTestTracer(t)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewBOSGatewayHandler(nil, logger)
	handler.pm4pyMCPClient = &testMockPM4PyMCPClient{
		discoverResult:    defaultTestDiscoverResult(),
		conformanceResult: defaultTestConformanceResult(),
		statisticsResult:  defaultTestStatisticsResult(),
	}

	router := gin.New()
	api := router.Group("/api")
	RegisterBOSGatewayRoutes(api, handler)

	return handler, router, recorder
}

// ============================================================================
// TestGatewayDiscoverCreatesSpan verifies that POST /api/bos/discover emits
// a span named "bos.gateway.discover".
// ============================================================================

func TestGatewayDiscoverCreatesSpan(t *testing.T) {
	logPath := bosCreateTempEventLogFile(t)
	_, router, recorder := setupGatewayWithTracing(t)

	body, _ := json.Marshal(BOSDiscoverRequest{LogPath: logPath, Algorithm: "inductive_miner"})
	req := httptest.NewRequest("POST", "/api/bos/discover", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	// Verify the span was recorded
	spans := recorder.Ended()
	assert.Greater(t, len(spans), 0, "At least one span should be emitted")

	// Find the gateway discover span
	var found bool
	for _, s := range spans {
		if s.Name() == "bos.gateway.discover" {
			found = true
			break
		}
	}
	assert.True(t, found, "bos.gateway.discover span should be emitted")
}

// ============================================================================
// TestGatewayDiscoverSetsCorrelationId verifies that if X-Correlation-ID is
// present, the span gets the chatmangpt.run.correlation_id attribute.
// ============================================================================

func TestGatewayDiscoverSetsCorrelationId(t *testing.T) {
	logPath := bosCreateTempEventLogFile(t)
	_, router, recorder := setupGatewayWithTracing(t)

	correlationID := "test-correlation-abc-123"
	body, _ := json.Marshal(BOSDiscoverRequest{LogPath: logPath, Algorithm: "inductive_miner"})
	req := httptest.NewRequest("POST", "/api/bos/discover", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-ID", correlationID)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	spans := recorder.Ended()
	var foundAttr bool
	for _, s := range spans {
		if s.Name() == "bos.gateway.discover" {
			for _, attr := range s.Attributes() {
				if string(attr.Key) == "chatmangpt.run.correlation_id" && attr.Value.AsString() == correlationID {
					foundAttr = true
					break
				}
			}
			break
		}
	}
	assert.True(t, foundAttr, "bos.gateway.discover span should have chatmangpt.run.correlation_id attribute")
}

// ============================================================================
// TestGatewayConformanceCreatesSpan verifies that POST /api/bos/conformance
// emits a span named "bos.gateway.conformance".
// ============================================================================

func TestGatewayConformanceCreatesSpan(t *testing.T) {
	logPath := bosCreateTempEventLogFile(t)
	_, router, recorder := setupGatewayWithTracing(t)

	body, _ := json.Marshal(BOSConformanceRequest{LogPath: logPath, ModelID: "model_123"})
	req := httptest.NewRequest("POST", "/api/bos/conformance", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	spans := recorder.Ended()
	var found bool
	for _, s := range spans {
		if s.Name() == "bos.gateway.conformance" {
			found = true
			break
		}
	}
	assert.True(t, found, "bos.gateway.conformance span should be emitted")
}

// ============================================================================
// TestGatewayStatisticsCreatesSpan verifies that POST /api/bos/statistics
// emits a span named "bos.gateway.statistics".
// ============================================================================

func TestGatewayStatisticsCreatesSpan(t *testing.T) {
	logPath := bosCreateTempEventLogFile(t)
	_, router, recorder := setupGatewayWithTracing(t)

	body, _ := json.Marshal(BOSStatisticsRequest{LogPath: logPath})
	req := httptest.NewRequest("POST", "/api/bos/statistics", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	spans := recorder.Ended()
	var found bool
	for _, s := range spans {
		if s.Name() == "bos.gateway.statistics" {
			found = true
			break
		}
	}
	assert.True(t, found, "bos.gateway.statistics span should be emitted")
}

// ============================================================================
// TestGatewaySpanHasOkStatus verifies that a successful discover sets
// span status to OK (not unset/error).
// ============================================================================

func TestGatewaySpanHasOkStatus(t *testing.T) {
	logPath := bosCreateTempEventLogFile(t)
	_, router, recorder := setupGatewayWithTracing(t)

	body, _ := json.Marshal(BOSDiscoverRequest{LogPath: logPath, Algorithm: "inductive_miner"})
	req := httptest.NewRequest("POST", "/api/bos/discover", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	spans := recorder.Ended()
	for _, s := range spans {
		if s.Name() == "bos.gateway.discover" {
			assert.Equal(t, "Ok", s.Status().Code.String(),
				"bos.gateway.discover span should have status OK on success")
			return
		}
	}
	t.Fatal("bos.gateway.discover span not found")
}
