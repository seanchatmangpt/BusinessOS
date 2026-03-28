---
title: "How To: Add OpenTelemetry Spans"
type: how-to
signal: "S=(linguistic, how-to, direct, markdown, numbered-steps)"
relates_to: [api-endpoints, bos-gateway-pattern, circuit-breaker-configuration]
prerequisites: [Go 1.24, OTEL Collector running (port 4317), Jaeger UI (port 16686)]
time: 20 minutes
difficulty: Intermediate
version: "1.0.0"
created: "2026-03-27"
---

# How To: Add OpenTelemetry Spans

> **Instrument your Go handler or service method so the operation appears in Jaeger.**
>
> Problem: You added a new handler or service call and need to see it in the trace — with attributes, status, and error recording.

---

## When to Add a Span

Use this decision guide before adding a span:

| Scenario | Add span? | Why |
|----------|-----------|-----|
| HTTP handler (Gin route) | Yes | Every handler is a trace entry point |
| Service method called by handler | Yes, if >5ms or calls external service | Provides latency breakdown in Jaeger |
| Internal utility / pure function | No | Adds noise; span overhead exceeds benefit |
| Background job / periodic sync | Yes | Lets you track job timing and failures |
| Database query helper | No, unless it wraps a slow query | `db.duration_ms` attribute on the parent span is enough |

**Rule of thumb:** Span every crossing of a layer boundary (Handler → Service, Service → pm4py-rust, Service → DB).

---

## Quick Start

Minimum viable span in 10 lines:

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/codes"
)

tracer := otel.Tracer("businessos")
ctx, span := tracer.Start(c.Request.Context(), "myhandler.operation")
defer span.End()

span.SetAttributes(attribute.String("my.key", "value"))
span.SetStatus(codes.Ok, "")
```

Open Jaeger at `http://localhost:16686`, select service **businessos**, and search for the span name.

---

## Step-by-Step: Add a Span to a Handler

### Step 1 — Import the OTEL packages

Add these imports to your handler file in `internal/handlers/`:

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/codes"

    semconv "github.com/rhl/businessos-backend/internal/semconv"
)
```

Use `semconv` for span name constants. Do not hard-code span name strings inline — the schema enforces them.

### Step 2 — Get a tracer from the global provider

At the top of your handler function (not at package level), acquire a tracer:

```go
gatewayTracer := otel.Tracer("businessos-gateway")
```

The string is the **instrumentation scope** — use `"businessos"` for general handlers and `"businessos-gateway"` for BOS gateway operations. See existing usage in `internal/handlers/bos_gateway.go` and `internal/observability/middleware.go`.

### Step 3 — Start the span and propagate context

Start the span immediately after getting the tracer. Always propagate the updated context back to the request:

```go
func (h *MyHandler) HandleSomething(c *gin.Context) {
    tracer := otel.Tracer("businessos")
    ctx, span := tracer.Start(c.Request.Context(), semconv.BosWorkspaceOperationSpan)
    defer span.End()
    c.Request = c.Request.WithContext(ctx)  // propagate so child spans link correctly
```

`defer span.End()` guarantees the span closes even if the handler returns early.

### Step 4 — Add attributes

Attach attributes that answer "what was this call about?":

```go
span.SetAttributes(
    attribute.String("bos.model_id", response.ModelID),
    attribute.String("bos.algorithm", response.Algorithm),
    attribute.Int64("bos.latency_ms", int64(latencyMs)),
)
```

For correlation IDs, follow the existing pattern from `bos_gateway.go`:

```go
if correlationID := c.Request.Header.Get("X-Correlation-ID"); correlationID != "" {
    span.SetAttributes(attribute.String(
        string(semconv.ChatmangptRunCorrelationIdKey), correlationID,
    ))
}
```

### Step 5 — Set status and record errors

Mark success explicitly at the end of the happy path:

```go
span.SetStatus(codes.Ok, "")
```

On error paths, record the error and set error status before returning:

```go
if err != nil {
    span.RecordError(err)
    span.SetStatus(codes.Error, "pm4py-rust unavailable")
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "upstream error"})
    return
}
```

Do not omit `span.SetStatus` — Jaeger shows spans without an explicit status as unset, which makes error detection unreliable.

---

## Step-by-Step: Add a Span to a Service Method

Service methods receive a `context.Context` rather than a `*gin.Context`. The pattern is the same except you use `ctx` directly.

### Step 1 — Store a tracer on the service struct

Initialize once in the constructor, not on every call:

```go
type MyService struct {
    tracer trace.Tracer
    // ...
}

func NewMyService() *MyService {
    return &MyService{
        tracer: otel.Tracer("businessos"),
    }
}
```

Import `"go.opentelemetry.io/otel/trace"` for the `trace.Tracer` type. See `internal/ontology/boardchair_l0_sync.go` for the canonical example.

### Step 2 — Start the span from the incoming context

```go
func (s *MyService) ProcessData(ctx context.Context, input string) error {
    ctx, span := s.tracer.Start(ctx, "myservice.process_data")
    defer span.End()
```

Pass the updated `ctx` to every downstream call so child spans are nested correctly in Jaeger.

### Step 3 — Add attributes and record errors

```go
    span.SetAttributes(
        attribute.String("input.length", strconv.Itoa(len(input))),
    )

    result, err := s.callDownstream(ctx, input)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return fmt.Errorf("myservice.process_data: %w", err)
    }

    span.SetAttributes(attribute.String("result.id", result.ID))
    span.SetStatus(codes.Ok, "")
    return nil
}
```

---

## Required Attributes

All spans in BusinessOS must include these attributes when the data is available:

| Attribute key | Type | When required |
|---------------|------|---------------|
| `chatmangpt.run.correlation_id` | string | When `X-Correlation-ID` header is present |
| `http.method` | string | HTTP handler spans (set by `TracingMiddleware` automatically) |
| `http.status_code` | int | HTTP handler spans (set by `TracingMiddleware` automatically) |
| `bos.latency_ms` | int64 | BOS gateway calls to pm4py-rust |

Use the typed constants from `internal/semconv/` for attribute keys (e.g. `semconv.ChatmangptRunCorrelationIdKey`). This provides compile-time safety: if the schema renames an attribute, the build breaks and forces you to update.

---

## Use the Observability Helpers

`internal/observability/` provides ready-made span starters for common domains. Prefer these over raw `otel.Tracer` calls:

```go
// A2A handler
ctx, span := observability.StartA2ASpan(c, "a2a.discover_agent")
defer span.End()

// CRM handler
ctx, span := observability.StartCRMSpan(c, "crm.list_deals", "pipelineID", pipelineID)
defer span.End()

// Generic handler
ctx, span := observability.StartSpan(c, "my.operation")
defer span.End()

// Record an error anywhere
observability.RecordError(span, err)
```

`TracingMiddleware` in `internal/observability/middleware.go` automatically creates a root span for every HTTP request, so you only need child spans inside handlers.

---

## Verify in Jaeger

1. Start all services: `make dev` from `BusinessOS/`.
2. Issue a request to the endpoint you instrumented (e.g. `curl -X POST http://localhost:8001/api/bos/discover ...`).
3. Open Jaeger at `http://localhost:16686`.
4. In the **Service** dropdown, select **businessos** (or **businessos-gateway** for gateway spans).
5. Click **Find Traces**.
6. Click your trace. Confirm:
   - The span name matches the constant you used (e.g. `bos.gateway.discover`).
   - Attributes appear in the **Tags** panel.
   - Status is **OK** or **ERROR** (not unset).

---

## Testing Spans

Use `otel.TestMainFunc` to flush spans during `go test`. Add to your test file's package:

```go
// In your_handler_test.go or testmain_test.go for the package
package handlers_test

import (
    "os"
    "testing"

    bosotel "github.com/rhl/businessos-backend/internal/otel"
)

func TestMain(m *testing.M) {
    bosotel.TestMainFunc(m)
}
```

This is a no-op unless `WEAVER_LIVE_CHECK=true` is set. When set, spans are exported to the Weaver OTLP receiver for schema validation during the test run. See `internal/semconv/weaver_live_check_test.go` for a complete example.

To emit a span inside a unit test and validate attributes:

```go
func TestMyHandlerEmitsSpan(t *testing.T) {
    ctx := context.Background()
    tr := otel.Tracer("businessos")
    ctx, span := tr.Start(ctx, semconv.BosWorkspaceOperationSpan)

    span.SetAttributes(attribute.String(
        string(semconv.ChatmangptRunCorrelationIdKey), "test-cid-123",
    ))
    span.SetStatus(codes.Ok, "")
    span.End()
    // If WEAVER_LIVE_CHECK=true, Weaver validates the span against schema.
    // If not, the test simply exercises the instrumentation path.
}
```

---

## Troubleshooting

**Spans not appearing in Jaeger**

- Confirm the OTEL Collector is running: `curl http://localhost:13133` should return `{"status":"Server available"}`.
- Confirm the backend started with the collector endpoint configured. Check the log line: `OpenTelemetry tracer initialized endpoint=...`.
- Spans are batched (10s flush). Wait 10–15 seconds after the request.

**Wrong service name in Jaeger**

- The service name comes from the resource set in `observability.InitTracer`. It is always `"businessos"`. The instrumentation scope string you pass to `otel.Tracer(...)` (e.g. `"businessos-gateway"`) is visible in Jaeger's **Library** column, not the service dropdown.

**Span status shows "UNSET"**

- You forgot `span.SetStatus(codes.Ok, "")` on the happy path. OTEL does not infer status from the absence of errors.

**Attributes missing**

- `span.SetAttributes(...)` must be called before `span.End()`. If you call it after `defer span.End()` resolves, attributes are lost.
- Verify imports include `"go.opentelemetry.io/otel/attribute"` (not just `"go.opentelemetry.io/otel"`).

**Child span not nested under parent**

- You must pass the context returned by `tracer.Start` to the downstream call. If you pass the original `c.Request.Context()` instead of the updated `ctx`, the child span will appear as a separate root trace.

---

## See Also

- `internal/observability/tracer.go` — `InitTracer` and `ShutdownTracer` (how the global provider is wired at startup)
- `internal/observability/middleware.go` — `TracingMiddleware`, `StartSpan`, `RecordError` helpers
- `internal/observability/handlers.go` — domain-specific span starters (`StartA2ASpan`, `StartCRMSpan`, etc.)
- `internal/handlers/bos_gateway.go` — production example: handler-level spans with correlation ID, attributes, and error recording
- `internal/ontology/boardchair_l0_sync.go` — production example: service-level tracer stored on struct
- `internal/otel/weaver.go` — `SetupWeaverLiveCheck` for Weaver live-check in tests
- `otel-collector-config.yaml` — collector pipeline (receivers, processors, exporters to Jaeger and Weaver)
