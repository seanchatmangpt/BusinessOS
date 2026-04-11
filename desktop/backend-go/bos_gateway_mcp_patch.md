# bos_gateway.go MCP Integration Patch

**File:** `BusinessOS/desktop/backend-go/internal/handlers/bos_gateway.go`

## Changes Required

### 1. Update Imports (lines 4-26)
**Remove:**
```go
"net/http"
"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
```

**Add:**
```go
"github.com/rhl/businessos-backend/internal/services"
```

### 2. Update Struct (lines 36-45)
**Remove:**
```go
	pm4pyURL         string
	httpClient       *http.Client
```

**Add:**
```go
	pm4pyMCPClient   *services.PMCPClient
```

### 3. Update Constructor (lines 54-89)
**Old:**
```go
pm4pyURL := "http://localhost:8090"
if envURL := os.Getenv("PM4PY_RUST_URL"); envURL != "" {
    pm4pyURL = envURL
}
// ... httpClient setup
```

**New:**
```go
pm4pyURL := "http://localhost:7015"
if envURL := os.Getenv("PM4PY_MCP_URL"); envURL != "" {
    pm4pyURL = envURL
} else if envURL := os.Getenv("PM4PY_RUST_URL"); envURL != "" {
    pm4pyURL = envURL
    logger.Warn("PM4PY_RUST_URL is deprecated, use PM4PY_MCP_URL instead")
}
// ... pm4pyMCPClient setup
pm4pyMCPClient: services.NewPM4PyMCPClient(pm4pyURL),
```

### 4. Update Discover Handler (lines 196-389)
**Remove all HTTP request code (lines 240-287)**

**Replace with:**
```go
// Call pm4py-mcp directly
result, err := h.pm4pyMCPClient.Discover(c.Request.Context(), eventLog, req.Algorithm)
if err != nil {
    h.logger.Error("discover: pm4py-mcp call failed", "error", err.Error())
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "pm4py-mcp unavailable"})
    return
}

// Map result to response format
response := BOSDiscoverResponse{
    ModelID:     result.ModelID,
    Algorithm:   result.Algorithm,
    Places:      result.ProcessModel.Places,
    Transitions: result.ProcessModel.Transitions,
    Arcs:        result.ProcessModel.Arcs,
    LatencyMs:   uint64(time.Since(startTime).Milliseconds()),
}
```

### 5. Update CheckConformance Handler (lines 393-579)
**Remove all HTTP request code**

**Replace with:**
```go
result, err := h.pm4pyMCPClient.Conformance(c.Request.Context(), eventLog, petriNetRaw)
if err != nil {
    // ... error handling ...
}

response := BOSConformanceResponse{
    Fitness:        result.Fitness,
    Precision:      result.Precision,
    Generalization: result.Generalization,
    Simplicity:     result.Simplicity,
    LatencyMs:      uint64(time.Since(startTime).Milliseconds()),
}
```

### 6. Update GetStatistics Handler (lines 583-803)
**Remove all HTTP request code**

**Replace with:**
```go
result, err := h.pm4pyMCPClient.Statistics(c.Request.Context(), eventLog)
if err != nil {
    // ... error handling ...
}

// Map to response format
activityFreq := make([]BOSActivityStatistic, len(result.ActivityFrequency))
for i, af := range result.ActivityFrequency {
    activityFreq[i] = BOSActivityStatistic{
        Activity:   af.Activity,
        Frequency:  af.Frequency,
        Percentage: af.Percentage,
    }
}
// ... etc
```

### 7. Update loadEventLogForGateway (lines 901-935)
**Remove XES parsing via HTTP** - pm4py-mcp doesn't support this

**Add comment:**
```go
// XES parsing via pm4py-mcp not supported - client must convert to JSON first
```

---

## Implementation Notes

1. **Timeout Handling**: MCP client has its own timeout (30s in mcp_client.go)
2. **Error Codes**: Map MCP errors to HTTP status codes appropriately
3. **Response Format**: pm4py-mcp returns JSON string, need to unmarshal twice
4. **Backward Compatibility**: Support PM4PY_RUST_URL during transition
5. **Tracing**: MCP client doesn't auto-propagate trace context - need to add manually

---

**Estimated Changes:** ~400 lines removed, ~150 lines added
