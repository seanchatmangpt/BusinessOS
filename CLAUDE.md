# BusinessOS — Claude Code Configuration

AI company operating system. Go 1.24 backend + SvelteKit 2 frontend + Electron desktop.
MIOSA template: self-hosted, agent-native business platform.

**REQUIRED READING:**
- **Root instructions**: See `../CLAUDE.md` (Constitutional Law, Evidence Standards, Git Safety, Dependencies)
- **Agent constraints**: See `./AGENTS.md` (Non-negotiable rules for all agents in this project)

## Build & Dev Commands

### Root Makefile (from repo root — boots everything)

```bash
make dev            # Boot ALL 4 systems + OTEL + Jaeger + run E2E smoke test
make health         # Quick health probe of all services
make verify         # Full verification: weaver + OTEL pipeline + E2E chain
make dev-logs       # Follow all service logs
make dev-down       # Stop all services (keep volumes)
make test-real      # Run real API integration tests (Groq, pm4py-mcp)
```

### BusinessOS Makefile (from BusinessOS/ — project-local)

```bash
make setup          # First-time: copy .env, pull images, start stack
make dev            # Start all services (Docker Compose), follow logs
make build          # Build all Docker images
make test           # Run all tests (backend + frontend)
make test-backend   # Go tests: cd desktop/backend-go && go test ./... -count=1
make test-frontend  # SvelteKit tests: cd frontend && npm test
make down           # Stop containers (preserves volumes)
make clean          # Stop + destroy volumes
make db-migrate     # Re-apply init.sql against running postgres
make db-seed        # Load seed data
```

### Single Test

```bash
# Go — run a specific test
cd desktop/backend-go && go test ./internal/<package>/... -run TestName

# Go — verbose
cd desktop/backend-go && go test ./internal/<package>/... -run TestName -v

# Frontend — run a specific test file
cd frontend && npx vitest run src/path/to/file.test.ts
```

## Ports

| Service       | Port  |
|---------------|-------|
| Go backend    | 8001  |
| SvelteKit     | 5173  |
| PostgreSQL    | 5432  |
| Redis         | 6379  |
| Ollama        | 11434 |
| YAWL v6       | 8080  |

## Architecture

```
HTTP Request -> Middleware (auth, CORS, rate limit)
  -> Handler (validation, serialization)
    -> Service (business logic, AI orchestration)
      -> Repository (data access via sqlc/pgx)
        -> PostgreSQL / Redis / AI Provider
```

**Key decisions:**
- Docker-first development (docker-compose.yml orchestrates all services)
- Gin HTTP framework, pgx/v5 for PostgreSQL, Redis for cache/pub-sub
- sqlc generates type-safe query code from SQL
- SvelteKit proxies `/api/*` to the Go backend via Vite

## Code Standards — Go Backend

- `slog` for all structured logging — never `fmt.Println` or `log`
- Handler -> Service -> Repository layering (strict separation)
- No hardcoded credentials — all secrets via `.env` / environment variables
- Parameterized queries only (sqlc handles this, raw SQL must use `$1` params)
- Input validation at handler boundaries
- Follow existing patterns in `internal/`

## Code Standards — Frontend

- TypeScript strict mode — no `any` type
- Svelte 5 Runes: `$state`, `$derived`, `$effect`
- Callback props — no `createEventDispatcher`
- TailwindCSS v4 for styling

## Key Backend Packages

| Package                  | Purpose                              |
|--------------------------|--------------------------------------|
| `internal/handlers/`     | Gin HTTP handlers, WebSocket, SSE    |
| `internal/services/`     | Business logic, AI workflows         |
| `internal/integrations/` | Google, Microsoft, HubSpot, Notion   |
| `internal/middleware/`    | Auth (JWT + OAuth), CORS, rate limit |
| `internal/database/`     | PostgreSQL, migrations, sqlc queries |
| `internal/agents/`       | Multi-agent orchestration            |
| `internal/signal/`       | Signal Theory S=(M,G,T,F,W) impl     |
| `internal/sorx/`         | SORX skill execution engine          |
| `internal/terminal/`     | PTY terminal + WebSocket             |
| `internal/cache/`        | Redis caching layer                  |
| `internal/security/`     | Encryption, audit, governance        |

## Auth

JWT tokens + Google OAuth. CSRF protection via cookies. Middleware in `internal/middleware/`.

## MCP / A2A

- MCP client: JSON-RPC 2.0 over stdio/HTTP/SSE for tool integration
- A2A: agent-to-agent communication protocol
- BusinessOS exposes A2A endpoints at `/api/integrations/a2a/agents`

## OSA SDK Stub

`github.com/Miosa-osa/sdk-go` is replaced with local stub at `internal/osasdk/` (see go.mod).
Provides Health(), Orchestrate(), Close() methods for local development.

## Cross-System Integration (pm4py-mcp)

BusinessOS acts as the gateway between the frontend and the pm4py-mcp process mining engine (MCP protocol on port 7015).

### BOS Gateway Endpoints

| Method | Endpoint | Purpose |
|--------|----------|---------|
| POST | `/api/bos/discover` | Invoke pm4py-mcp discovery tool |
| POST | `/api/bos/conformance` | Invoke pm4py-mcp conformance checking tool |
| POST | `/api/bos/statistics` | Invoke pm4py-mcp statistics tool |

All gateway endpoints call pm4py-mcp via MCP protocol on port 7015 (see `PM4PY_MCP_URL` env var).

### Two-Phase Commit (2PC) Endpoints

| Method | Endpoint | Purpose |
|--------|----------|---------|
| POST | `/api/bos/tx/prepare` | Prepare phase of distributed transaction |
| POST | `/api/bos/tx/commit` | Commit phase after all participants prepare |
| POST | `/api/bos/tx/abort` | Abort and roll back all participants |

### SSE Streaming

- `POST /api/bos/progress` -- receives progress events from pm4py-mcp and fans out to SSE clients
- `GET /api/bos/stream/discover/:session_id` -- client-side SSE stream for discovery results

### Write-Ahead Logging (WAL)

Discovery results are persisted via write-ahead logging before acknowledgment, ensuring durability across crashes.

### Backpressure and Limits

| Resource | Limit | Purpose |
|----------|-------|---------|
| Semaphore | 10 concurrent | Max concurrent requests to pm4py-mcp |
| Connection pool | 100 connections | MCP connection reuse to pm4py-mcp |
| SSE limit | 1000 connections | Max concurrent SSE stream subscribers |

### Environment Variables

| Variable | Default | Purpose |
|----------|---------|---------|
| `PM4PY_MCP_URL` | `http://localhost:7015` | pm4py-mcp service URL |
| `PM4PY_RUST_URL` | `http://localhost:8090` | **DEPRECATED** — fallback for backward compatibility |

## YAWL v6 Integration

BusinessOS proxies conformance checking and WCP pattern spec loading to the local YAWL v6 engine.

| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/yawl/health` | Check YAWL engine status |
| POST | `/api/yawl/conformance` | Run conformance check via YAWL engine |
| POST | `/api/yawl/spec` | Build YAWL specification XML |
| GET | `/api/yawl/spec/load` | Load WCP pattern spec from disk |

### Environment Variables

| Variable | Default | Purpose |
|----------|---------|---------|
| `YAWLV6_URL` | `http://localhost:8080` | YAWL v6 engine base URL |
| `YAWLV6_SPECS_PATH` | `~/yawlv6/exampleSpecs` | Path to YAWL WCP pattern spec files |

## Code Generation (unrdf)

BusinessOS uses RDF-based code generation for business agents, MCP tools, and decision records via `unrdf` and the `BusinessOS/unrdf.toml` configuration.

**When to regenerate**: Only when EXPLICITLY requested for one of these reasons:
- New agent needs scaffolding (use `--rule <rule-name>` not full sync)
- Ontology changed and existing generated file is stale
- New rule added to unrdf.toml

**NEVER run `unrdf sync` without specifying `--rule` unless a full refresh is explicitly requested.**

### Selective generation (JIT - Just-In-Time)

```bash
# Generate only business-agent-composition rule
unrdf sync --config BusinessOS/unrdf.toml --rule business-agent-composition

# Generate only business-mcp-tools rule
unrdf sync --config BusinessOS/unrdf.toml --rule business-mcp-tools

# Generate only business-decision-records rule
unrdf sync --config BusinessOS/unrdf.toml --rule business-decision-records

# List available rules (dry-run)
unrdf sync --config BusinessOS/unrdf.toml --dry-run
```

### Configuration

File: `BusinessOS/unrdf.toml`

| Rule | Purpose | Output |
|------|---------|--------|
| `business-agent-composition` | Generate agent specifications (finance, operations, sales, hr, legal) | `docs/generated/agents/*_agent_spec.md` |
| `business-mcp-tools` | Generate MCP tool registrations (create-deal, schedule-meeting, etc.) | `desktop/backend-go/internal/mcp-tools/*_tool.mjs` |
| `business-decision-records` | Generate Architecture Decision Records | `docs/adr/adr-*.md` |
| `business-quality-metrics` | Generate quality metrics (DPMO, Sigma, SPC) | `docs/generated/quality-metrics/*_metrics_spec.md` |

### Principles

- **JIT (Just-In-Time)**: Generate only when needed, not speculatively
- **Selective**: Use `--rule` flag to generate specific artifacts, not entire ontology
- **Deterministic**: Same ontology + query = same output (idempotent)
- **Go-friendly**: Generated artifacts are JSON/markdown (not Elixir code) for Go projects

## Diátaxis Docs

`docs/diataxis/` — tutorials, explanation, how-to, reference for this project.
