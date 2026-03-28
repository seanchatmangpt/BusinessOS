---
title: "How To: Use the BOS Gateway Pattern"
type: how-to
signal: "S=(linguistic, how-to, direct, markdown, numbered-steps)"
relates_to: [api-endpoints, add-api-endpoint, circuit-breaker-configuration]
prerequisites: [Go 1.24, BusinessOS backend running on :8001, pm4py-rust on :8090]
time: 15 minutes
difficulty: Intermediate
version: "1.0.0"
created: "2026-03-27"
---

# How To: Use the BOS Gateway Pattern

> **Proxy a request from BusinessOS to a downstream service (pm4py-rust, YAWL, OSA).**
>
> Problem: You need to forward an incoming HTTP request to an external engine, add OTEL tracing, handle timeouts, and return a structured response — without duplicating boilerplate across every endpoint.

---

## What is the BOS Gateway?

The BOS Gateway is a thin HTTP proxy layer inside BusinessOS. It accepts a request on `/api/bos/*`, forwards it to a downstream service (e.g., pm4py-rust on `:8090`), records an OTEL span, and returns a normalized JSON response to the caller. The target service URL comes from an environment variable, and the 30-second client timeout prevents unbounded hangs.

---

## When to Use It

Use the BOS Gateway pattern when:

- You need to call **pm4py-rust** (`PM4PY_RUST_URL`) for process mining (discovery, conformance, statistics)
- You need to call **Canopy** (`CANOPY_WEBHOOK_URL`) for workflow notifications
- You are adding a new downstream engine (YAWL, OSA) that BusinessOS should proxy to
- You want distributed tracing (W3C `traceparent`) propagated end-to-end to the downstream service

Do not use it for business logic that lives entirely inside BusinessOS (use a service + repository instead — see [add-api-endpoint.md](add-api-endpoint.md)).

---

## Quick Start: Call the Existing Gateway

With BusinessOS on `:8001` and pm4py-rust on `:8090`:

```bash
# Check gateway health
curl http://localhost:8001/api/bos/status

# Run process discovery (requires a JSON event log file on the server)
curl -X POST http://localhost:8001/api/bos/discover \
  -H "Authorization: Bearer $BUSINESSOS_API_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"log_path": "/tmp/my_log.json", "algorithm": "inductive_miner"}'

# Check conformance of an existing model
curl -X POST http://localhost:8001/api/bos/conformance \
  -H "Authorization: Bearer $BUSINESSOS_API_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"log_path": "/tmp/my_log.json", "model_id": "model_test_001"}'
```

Expected 200 response from `/discover`:

```json
{
  "model_id": "a1b2c3...",
  "algorithm": "inductive_miner",
  "places": 4,
  "transitions": 3,
  "arcs": 7,
  "model_data": { ... },
  "latency_ms": 42
}
```

---

## Step-by-Step: Add a New Gateway Route

Follow these five steps to proxy a new downstream service through the BOS Gateway.

### Step 1: Add the Target URL as an Environment Variable

Add the downstream URL to `.env` and load it in `NewBOSGatewayHandler` (or create a new handler struct if the target is a separate service):

```bash
# .env
MY_SERVICE_URL=http://localhost:9100
```

In `internal/handlers/bos_gateway.go`, add the field and load it in the constructor:

```go
type BOSGatewayHandler struct {
    // ... existing fields ...
    myServiceURL string
}

func NewBOSGatewayHandler(pool *pgxpool.Pool, logger *slog.Logger) *BOSGatewayHandler {
    // ... existing init ...
    myServiceURL := os.Getenv("MY_SERVICE_URL")
    return &BOSGatewayHandler{
        // ... existing fields ...
        myServiceURL: myServiceURL,
        httpClient: &http.Client{
            Transport: otelhttp.NewTransport(http.DefaultTransport),
            Timeout:   30 * time.Second, // required: prevents unbounded hang
        },
    }
}
```

The `otelhttp.NewTransport` wrapper automatically injects W3C `traceparent` headers into every outbound request.

### Step 2: Define Request/Response Types

Add typed structs for your endpoint in `bos_gateway.go`. Match the JSON shape your downstream service expects:

```go
// MyServiceRequest is the inbound request body from the BOS CLI / frontend.
type MyServiceRequest struct {
    InputPath string `json:"input_path" binding:"required"`
    Mode      string `json:"mode,omitempty"`
}

// MyServiceResponse is what BusinessOS returns to the caller.
type MyServiceResponse struct {
    ResultID  string `json:"result_id"`
    Status    string `json:"status"`
    LatencyMs uint64 `json:"latency_ms"`
}
```

### Step 3: Implement the Handler

Copy the gateway handler pattern. The key parts are: OTEL span start, bind JSON, build downstream request, forward, parse response, return.

```go
// HandleMyService handles POST /api/bos/my-service.
func (h *BOSGatewayHandler) HandleMyService(c *gin.Context) {
    startTime := time.Now()

    // 1. Start OTEL span — name matches semconv constant
    tracer := otel.Tracer("businessos-gateway")
    ctx, span := tracer.Start(c.Request.Context(), "bos.gateway.my_service")
    defer span.End()
    c.Request = c.Request.WithContext(ctx)

    // 2. Propagate correlation ID for end-to-end tracing
    if cid := c.Request.Header.Get("X-Correlation-ID"); cid != "" {
        span.SetAttributes(attribute.String("chatmangpt.run.correlation_id", cid))
    }

    // 3. Bind and validate the incoming request
    var req MyServiceRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        h.logger.Warn("my-service: invalid request", "error", err.Error())
        span.SetStatus(codes.Error, "invalid request")
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
        return
    }

    // 4. Build downstream payload
    downstream := map[string]interface{}{
        "input": req.InputPath,
        "mode":  req.Mode,
    }
    body, err := json.Marshal(downstream)
    if err != nil {
        h.logger.Error("my-service: marshal failed", "error", err.Error())
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to build request"})
        return
    }

    // 5. Forward to downstream service
    httpReq, _ := http.NewRequestWithContext(ctx, "POST",
        h.myServiceURL+"/api/process", bytes.NewReader(body))
    httpReq.Header.Set("Content-Type", "application/json")
    if cid := c.Request.Header.Get("X-Correlation-ID"); cid != "" {
        httpReq.Header.Set("X-Correlation-ID", cid)
    }

    httpResp, err := h.httpClient.Do(httpReq)
    if err != nil {
        h.logger.Error("my-service: downstream request failed", "error", err.Error())
        span.RecordError(err)
        span.SetStatus(codes.Error, "downstream unavailable")
        c.JSON(http.StatusServiceUnavailable, gin.H{"error": "my-service unavailable"})
        return
    }
    defer httpResp.Body.Close()

    if httpResp.StatusCode != http.StatusOK {
        span.SetStatus(codes.Error, fmt.Sprintf("downstream returned %d", httpResp.StatusCode))
        c.JSON(http.StatusServiceUnavailable, gin.H{"error": "my-service error"})
        return
    }

    // 6. Parse and reshape response
    var raw map[string]interface{}
    if err := json.NewDecoder(httpResp.Body).Decode(&raw); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse response"})
        return
    }

    response := MyServiceResponse{
        ResultID:  fmt.Sprintf("%v", raw["result_id"]),
        Status:    "ok",
        LatencyMs: uint64(time.Since(startTime).Milliseconds()),
    }

    span.SetAttributes(
        attribute.String("bos.result_id", response.ResultID),
        attribute.Int64("bos.latency_ms", int64(response.LatencyMs)),
    )
    span.SetStatus(codes.Ok, "")
    c.JSON(http.StatusOK, response)
}
```

### Step 4: Register the Route

Routes are registered in `internal/handlers/routes_bos_gateway.go`. Add your new endpoint to the existing `bosGroup`:

```go
// internal/handlers/routes_bos_gateway.go
func (h *Handlers) registerBOSGatewayRoutes(api *gin.RouterGroup) {
    token := os.Getenv("BUSINESSOS_API_TOKEN")
    bosAuth := middleware.StaticBearerAuth(token)
    bosHandler := NewBOSGatewayHandler(h.pool, slog.Default())
    bosGroup := api.Group("/bos")
    bosGroup.Use(bosAuth)
    bosGroup.POST("/discover", bosHandler.Discover)
    bosGroup.POST("/conformance", bosHandler.CheckConformance)
    bosGroup.POST("/statistics", bosHandler.GetStatistics)
    bosGroup.GET("/status", bosHandler.GetStatus)
    bosGroup.POST("/my-service", bosHandler.HandleMyService) // add this line
}
```

The `StaticBearerAuth` middleware reads `BUSINESSOS_API_TOKEN` from the environment and requires callers to provide `Authorization: Bearer <token>`.

### Step 5: Write the Test

Mirror the test pattern from `bos_gateway_test.go`. Spin up a `httptest.Server` as the mock downstream, point the handler at it, and assert the response shape.

```go
// internal/handlers/bos_my_service_test.go
package handlers

import (
    "bytes"
    "encoding/json"
    "io"
    "log/slog"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestHandleMyService_Success(t *testing.T) {
    gin.SetMode(gin.TestMode)

    // 1. Spin up mock downstream
    mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]interface{}{
            "result_id": "res_001",
            "status":    "ok",
        })
    }))
    t.Cleanup(mock.Close)

    // 2. Wire handler to mock
    logger := slog.New(slog.NewTextHandler(io.Discard, nil))
    handler := NewBOSGatewayHandler(nil, logger)
    handler.myServiceURL = mock.URL

    router := gin.New()
    api := router.Group("/api")
    RegisterBOSGatewayRoutes(api, handler)

    // 3. Make request
    body := bytes.NewBufferString(`{"input_path": "/tmp/data.json"}`)
    req := httptest.NewRequest("POST", "/api/bos/my-service", body)
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    // 4. Assert
    require.Equal(t, http.StatusOK, w.Code)
    var resp MyServiceResponse
    require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
    assert.Equal(t, "res_001", resp.ResultID)
    assert.Equal(t, "ok", resp.Status)
}

func TestHandleMyService_DownstreamUnavailable_Returns503(t *testing.T) {
    gin.SetMode(gin.TestMode)

    logger := slog.New(slog.NewTextHandler(io.Discard, nil))
    handler := NewBOSGatewayHandler(nil, logger)
    handler.myServiceURL = "http://127.0.0.1:1" // unreachable port

    router := gin.New()
    api := router.Group("/api")
    RegisterBOSGatewayRoutes(api, handler)

    body := bytes.NewBufferString(`{"input_path": "/tmp/data.json"}`)
    req := httptest.NewRequest("POST", "/api/bos/my-service", body)
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}
```

Run:

```bash
cd desktop/backend-go
go test ./internal/handlers/... -run TestHandleMyService -v
```

---

## Troubleshooting

### 503 Service Unavailable — `"my-service unavailable"`

**Cause:** The downstream service is not reachable at `MY_SERVICE_URL`.

**Fix:**
```bash
# Check the env var is set
echo $MY_SERVICE_URL

# Verify the downstream is reachable
curl $MY_SERVICE_URL/api/health

# Check the backend logs
make dev-logs | grep "my-service"
```

### 400 Bad Request — `"Invalid request format"`

**Cause:** Missing a `binding:"required"` field in the request body (e.g., `input_path` was omitted).

**Fix:** Check the request JSON matches the struct tags:
```bash
# Wrong (missing required field):
curl -X POST .../bos/my-service -d '{}'

# Correct:
curl -X POST .../bos/my-service -d '{"input_path": "/tmp/data.json"}'
```

### 502 / connection refused from downstream

**Cause:** `MY_SERVICE_URL` points to a port that is not listening, or the Docker network name is wrong inside the container.

**Fix:** In `docker-compose.yml`, use the service container name, not `localhost`:
```yaml
environment:
  MY_SERVICE_URL: "http://my-service:9100"  # container-to-container
```

### OTEL span not appearing in Jaeger

**Cause:** The global OTEL provider is not initialized before requests arrive, or the span name is not registered in the semconv schema.

**Fix:**
1. Confirm `observability.InitTracer` is called in `bootstrap.go` before the HTTP server starts.
2. Add the span name to `internal/semconv/bos_span_names.go`:

```go
const (
    BosGatewayMyServiceSpan = "bos.gateway.my_service"
)
```

3. Use the constant in your handler (`tracer.Start(ctx, semconv.BosGatewayMyServiceSpan)`).
4. Open Jaeger at `http://localhost:16686`, select service `businessos`, search for `bos.gateway.my_service`.

### Timeout after 30 seconds — request hangs

**Cause:** Downstream service is alive but slow. The 30-second client timeout in `NewBOSGatewayHandler` will fire and return 503.

**Fix:** If the downstream operation is expected to be slow, add a context deadline before calling the downstream:

```go
ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
defer cancel()
httpReq, _ = http.NewRequestWithContext(ctx, "POST", h.myServiceURL+"/api/process", body)
```

---

## See Also

- [How To: Add a New API Endpoint](add-api-endpoint.md) — general endpoint pattern without proxy
- [`internal/handlers/bos_gateway.go`](../../../desktop/backend-go/internal/handlers/bos_gateway.go) — full source of the gateway
- [`internal/handlers/routes_bos_gateway.go`](../../../desktop/backend-go/internal/handlers/routes_bos_gateway.go) — route registration
- [`internal/semconv/bos_span_names.go`](../../../desktop/backend-go/internal/semconv/bos_span_names.go) — OTEL span name constants
- [BusinessOS CLAUDE.md — BOS Gateway section](../../CLAUDE.md#cross-system-integration-pm4py-rust) — environment variable reference
