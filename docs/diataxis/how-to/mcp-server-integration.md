---
title: "How To: Integrate an MCP Server"
type: how-to
signal: "S=(linguistic, how-to, direct, markdown, numbered-steps)"
relates_to: [api-endpoints, bos-gateway-pattern]
prerequisites: [Go 1.24, BusinessOS backend running, MCP server binary or URL available]
time: 20 minutes
difficulty: Intermediate
version: "1.0.0"
created: "2026-03-27"
---

# How To: Integrate an MCP Server

> **Problem:** You have an MCP-compatible server (a binary, a remote HTTP endpoint, or an SSE stream) and want BusinessOS agents to discover and call its tools.

---

## What is MCP in BusinessOS?

BusinessOS uses the [Model Context Protocol](https://modelcontextprotocol.io/) (MCP) — JSON-RPC 2.0 over stdio, HTTP, or SSE — to surface external tools to agents. The `MCPService` (`internal/services/mcp.go`) merges built-in tools (calendar, Slack, Notion) with dynamically registered external servers stored in the `mcp_servers` database table. Tools from external servers are namespaced as `<server-name>.<tool-name>` so they can be routed back to the correct server at call time.

---

## Quick Start — Connect an SSE MCP Server

```bash
# 1. Register the server via the API (requires a valid JWT)
curl -X POST http://localhost:8001/api/integrations/mcp/connectors \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-tools",
    "server_url": "https://mcp.example.com/tools",
    "transport": "sse",
    "auth_type": "bearer",
    "api_key": "sk-..."
  }'

# 2. Trigger tool discovery
curl -X POST http://localhost:8001/api/integrations/mcp/connectors/<id>/discover \
  -H "Authorization: Bearer $TOKEN"

# 3. List all tools (built-in + dynamic)
curl http://localhost:8001/api/mcp/tools \
  -H "Authorization: Bearer $TOKEN"
```

Tools from this server now appear as `my-tools.<tool-name>` in every `GetAllTools()` response.

---

## Three Connection Modes

| Mode | `transport` value | When to use |
|------|-------------------|-------------|
| **SSE** | `sse` | Remote server streams responses as Server-Sent Events |
| **Streamable HTTP** | `streamable_http` | Remote server returns plain JSON (no streaming) |
| **stdio (Weaver)** | n/a — internal only | Local binary launched by BusinessOS; used for `weaver registry mcp` |

### SSE — minimal example

```go
client := services.NewMCPClient(
    "https://mcp.example.com/tools",
    "bearer",
    "sk-mytoken",
    nil,
)
tools, err := client.DiscoverTools(ctx)
```

`MCPClient` detects `Content-Type: text/event-stream` automatically and switches to its SSE parser.

### Streamable HTTP — minimal example

```go
client := services.NewMCPClient(
    "https://mcp.example.com/jsonrpc",
    "api_key",  // sends X-API-Key header
    "my-api-key",
    map[string]string{"X-Workspace-ID": "ws_123"},
)
result, err := client.ExecuteTool(ctx, "search", map[string]interface{}{
    "query": "revenue Q1",
})
```

Both SSE and streamable HTTP use the same `MCPClient` struct — the content-type header drives the parse path.

### stdio (Weaver) — internal only

The `weaverMCPProcess` in `internal/services/weaver_mcp_stdio.go` is used exclusively to talk to the local `weaver registry mcp` subprocess. It is not user-facing. You do not register it via the API.

```go
// Internal use only — started during weaver semconv tool initialization
proc, err := startWeaverMCPProcess(ctx, "/usr/local/bin/weaver", "./semconv/model")
tools, err := proc.listTools(ctx)
```

---

## Step-by-Step: Register a stdio MCP Server (Weaver)

This path is for BusinessOS developers extending the built-in weaver semconv tools, not end users.

1. Ensure the `weaver` binary is on `$PATH`:
   ```bash
   weaver --version
   ```

2. Start the process from your service code using `startWeaverMCPProcess`:
   ```go
   proc, err := startWeaverMCPProcess(ctx, "weaver", "./semconv/model")
   if err != nil {
       return fmt.Errorf("weaver mcp start: %w", err)
   }
   defer proc.Close()
   ```
   `startWeaverMCPProcess` performs the MCP `initialize` handshake and sends `notifications/initialized` before returning. A 60-second deadline governs each `rpc` call.

3. List tools:
   ```go
   tools, err := proc.listTools(ctx)
   ```

4. Call a tool:
   ```go
   result, err := proc.callTool(ctx, "check_registry", map[string]interface{}{
       "registry_path": "./semconv/model",
   })
   ```

5. Always close the process when done — `proc.Close()` kills the subprocess and releases the stdin pipe.

---

## Step-by-Step: Register an HTTP/SSE MCP Server

1. **POST** to register the server. Names must be lowercase alphanumeric with hyphens or underscores, 1–100 characters:
   ```bash
   curl -X POST http://localhost:8001/api/integrations/mcp/connectors \
     -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "name":        "github",
       "description": "GitHub MCP server",
       "server_url":  "https://mcp.github.example.com",
       "transport":   "sse",
       "auth_type":   "bearer",
       "api_key":     "ghp_...",
       "headers":     {"X-GitHub-Api-Version": "2022-11-28"}
     }'
   ```
   The API key is encrypted with AES-256-GCM before storage (`security.GetGlobalEncryption()`). The response never returns the raw token — only `has_auth: true`.

2. **POST `/:id/discover`** to populate the tools cache:
   ```bash
   curl -X POST http://localhost:8001/api/integrations/mcp/connectors/<id>/discover \
     -H "Authorization: Bearer $TOKEN"
   ```
   BusinessOS calls `client.DiscoverTools(ctx)` which sends a `tools/list` JSON-RPC request. Discovered tools are written to `mcp_servers.tools_cache` (JSONB column). On failure the `status` column is set to `"error"` and `last_error` is populated.

3. **(Optional) Update the server** — enable/disable, rotate the API key, or change transport:
   ```bash
   curl -X PUT http://localhost:8001/api/integrations/mcp/connectors/<id> \
     -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"enabled": false}'
   ```

4. **(Optional) Delete** when no longer needed:
   ```bash
   curl -X DELETE http://localhost:8001/api/integrations/mcp/connectors/<id> \
     -H "Authorization: Bearer $TOKEN"
   ```

Limits: max 20 servers per user, max 10 custom headers per server.

---

## Tool Discovery

BusinessOS discovers tools lazily — tools are loaded from `mcp_servers.tools_cache` (written during `/discover`) rather than connecting live on every request.

```
GET /api/mcp/tools
→ MCPHandler.ListMCPTools
  → MCPService.GetAllTools()
    → GetBuiltinTools()      // calendar, Slack, Notion, search_conversations, …
    → appendWeaverSemconvTools()
    → getDynamicMCPTools()   // reads tools_cache for all enabled servers
```

`getDynamicMCPTools` in `mcp.go` reads `ListEnabledMCPServers` from the database and unmarshals the cached `[]MCPClientTool` JSON. Each tool is exposed as `<server-name>.<original-tool-name>` with `source: "mcp:<server-name>"`.

To force a refresh, call the `/discover` endpoint again.

---

## Tool Execution

When an agent calls `my-tools.search`, `MCPService.ExecuteTool` routes it:

```
ExecuteTool(ctx, "my-tools.search", args)
  → strings.SplitN("my-tools.search", ".", 2)  →  ["my-tools", "search"]
  → executeDynamicMCPTool(ctx, "my-tools", "search", args)
    → ListEnabledMCPServers  →  find server named "my-tools"
    → decrypt auth token
    → NewMCPClient(server_url, auth_type, token, headers)
    → client.ExecuteTool(ctx, "search", args)
```

If the server is not found, `executeDynamicMCPTool` returns `"MCP server not found: my-tools"` and `ExecuteTool` falls through to built-in tool handling.

---

## Error Handling

| Situation | What happens |
|-----------|--------------|
| Server URL is a private/loopback IP | `ValidateMCPServerURL` returns error; registration blocked (SSRF protection) |
| `/discover` cannot reach the server | `status` → `"error"`, `last_error` set; `DiscoverTools` returns the HTTP error |
| Tool call fails at runtime | `executeDynamicMCPTool` returns `fmt.Errorf("failed to call MCP tool: %w", err)` |
| Server returns JSON-RPC error | `parseToolsFromJSON` / `parseToolCallFromJSON` surfaces the RPC error code and message |
| stdio process crashes | `readUntilID` detects closed stdout and returns `"weaver mcp: stdout closed before response"` |
| Auth token decryption fails | Returns `"failed to decrypt MCP server credentials"` — tool call aborted |

There is no automatic circuit breaker on MCP calls. If you need retry/fallback logic, add it in the calling service layer.

---

## Testing

### Verify registration and tool cache

```bash
# Register and discover in one pass
curl -X POST http://localhost:8001/api/integrations/mcp/connectors/<id>/test \
  -H "Authorization: Bearer $TOKEN"
# Response: {"success": true, "tool_count": 12, "tools": [...]}

# Confirm tools appear in the unified list
curl http://localhost:8001/api/mcp/tools \
  -H "Authorization: Bearer $TOKEN" | jq '.tools[] | select(.source | startswith("mcp:"))'
```

### Unit test your handler/service interaction

```go
// Use the real MCPClient against a local test server
srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "jsonrpc": "2.0",
        "id":      1,
        "result": map[string]interface{}{
            "tools": []map[string]interface{}{
                {"name": "echo", "description": "Echo input", "inputSchema": map[string]interface{}{}},
            },
        },
    })
}))
defer srv.Close()

client := services.NewMCPClient(srv.URL, "none", "", nil)
tools, err := client.DiscoverTools(context.Background())
assert.NoError(t, err)
assert.Equal(t, "echo", tools[0].Name)
```

Existing tests: `internal/services/mcp_client_test.go`, `internal/handlers/mcp_servers_test.go`.

---

## See Also

- `desktop/backend-go/internal/services/mcp.go` — `MCPService`, `getDynamicMCPTools`, `executeDynamicMCPTool`
- `desktop/backend-go/internal/services/mcp_client.go` — `MCPClient`, `ValidateMCPServerURL`, `DiscoverTools`, `ExecuteTool`
- `desktop/backend-go/internal/services/weaver_mcp_stdio.go` — `weaverMCPProcess`, `startWeaverMCPProcess`
- `desktop/backend-go/internal/handlers/mcp_servers.go` — REST CRUD for `mcp_servers` table
- `desktop/backend-go/internal/database/migrations/101_mcp_servers.sql` — schema reference
- `BusinessOS/CLAUDE.md` — MCP / A2A section for architecture overview
