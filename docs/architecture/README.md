# Architecture Overview

BusinessOS is a full-stack web application composed of a Go backend, a SvelteKit frontend, and a PostgreSQL + Redis data layer. The system is built around an AI agent called OSA (Operating System Agent) that classifies user intent and dispatches work to the appropriate handler.

---

## System Overview

```
                        ┌───────────────────────────────────────────┐
                        │              Browser Client               │
                        │     SvelteKit 2 + Svelte 5 + Tailwind     │
                        └──────────────────┬────────────────────────┘
                                           │ HTTPS / SSE
                        ┌──────────────────▼────────────────────────┐
                        │             Go Backend (Gin)               │
                        │          http://localhost:8001             │
                        │                                            │
                        │  ┌──────────────────────────────────────┐ │
                        │  │          Chat / AI Layer              │ │
                        │  │   LLM Provider + RAG Context          │ │
                        │  │   (Claude, GPT, Ollama, Groq, xAI)   │ │
                        │  └──────────┬───────────────────────────┘ │
                        │             │                              │
                        │  ┌──────────────────────────────────────┐ │
                        │  │         Handler Layer                 │ │
                        │  │  (chat, projects, tasks, crm, ...)    │ │
                        │  └──────────┬───────────────────────────┘ │
                        │             │                              │
                        │  ┌──────────▼──────────┐ ┌─────────────┐ │
                        │  │  Service Layer       │ │  Container  │ │
                        │  │  (business logic)    │ │  Manager    │ │
                        │  └──────────┬───────────┘ │  (Docker)   │ │
                        │             │              └─────────────┘ │
                        │  ┌──────────▼───────────────────────────┐ │
                        │  │       Repository / SQLC Layer         │ │
                        │  └──────────┬───────────────────────────┘ │
                        └─────────────┼──────────────────────────────┘
                                      │
               ┌──────────────────────┼────────────────────────┐
               │                      │                         │
    ┌──────────▼──────────┐ ┌─────────▼───────────┐ ┌─────────▼──────────┐
    │  PostgreSQL 15+      │ │  Redis 7+           │ │  Docker Daemon     │
    │  + pgvector          │ │  Sessions / Cache   │ │  Terminal Sandbox  │
    │  Primary datastore   │ │  Pub/Sub (SSE)      │ │  (Alpine Linux)    │
    └──────────────────────┘ └─────────────────────┘ └────────────────────┘
```

---

## Backend

**Language:** Go 1.24+
**Framework:** [Gin](https://github.com/gin-gonic/gin)
**Location:** `backend/`

The backend follows a strict three-layer architecture:

```
HTTP Request
    │
    ▼
Handler       — validates input, authenticates, writes HTTP response
    │
    ▼
Service       — business logic, orchestration, AI calls
    │
    ▼
Repository    — SQL queries via SQLC, typed database access
    │
    ▼
PostgreSQL
```

Each layer has a single responsibility. Handlers never contain business logic; services never write raw SQL. This makes each layer independently testable.

**Key packages:**

| Package | Responsibility |
|---------|---------------|
| `internal/handlers/` | HTTP handler registration and request parsing |
| `internal/services/` | Business logic and AI orchestration |
| `internal/integrations/` | External API connectors (9 providers) |
| `internal/database/sqlc/` | Auto-generated type-safe SQL query functions |
| `internal/middleware/` | Auth, CSRF, CORS, rate limiting |
| `internal/container/` | Docker sandbox lifecycle management |
| `internal/cache/` | Redis caching layer |

**Logging:** All logging uses the standard `log/slog` package. `fmt.Printf` is never used for application logs.

---

## Frontend

**Framework:** [SvelteKit 2](https://kit.svelte.dev/)
**UI Library:** [Svelte 5](https://svelte.dev/)
**Styling:** [Tailwind CSS](https://tailwindcss.com/)
**Language:** TypeScript (strict mode)
**Location:** `frontend/`

The frontend uses SvelteKit's filesystem-based routing. All application routes live under `src/routes/(app)/`. Authentication-gated routes share a layout that redirects unauthenticated users to the login page.

**Key directories:**

| Directory | Contents |
|-----------|---------|
| `src/routes/(app)/` | All authenticated application views |
| `src/routes/onboarding/` | Multi-step onboarding flow |
| `src/routes/auth/` | Login, register, password reset |
| `src/lib/api/` | API client functions (typed fetch wrappers) |
| `src/lib/stores/` | Svelte stores for shared state |
| `src/lib/components/` | Reusable UI components |

**Data loading:** `+page.server.ts` files handle server-side data loading. Form actions are used for mutations. Client-side state is managed via Svelte stores.

---

## Database

**Primary:** PostgreSQL 15+ with the [pgvector](https://github.com/pgvector/pgvector) extension
**Cache / Sessions:** Redis 7+

### PostgreSQL

PostgreSQL stores all application data. The `pgvector` extension adds a `vector` column type used to store semantic embeddings for:

- AI conversation memory
- Document knowledge base (RAG)
- Semantic search across business entities

Schema changes are managed as numbered SQL migration files in `backend/internal/database/migrations/`. Migrations must be applied in order and are never modified after being committed.

Type-safe query functions are generated from SQL using [sqlc](https://sqlc.dev/). To regenerate after modifying a query:

```bash
cd backend && sqlc generate
```

### Redis

Redis provides:

- **Session storage** — JWT session data with TTL
- **Query cache** — frequently accessed data (invalidated on write)
- **SSE channel** — real-time event broadcasting to connected clients

---

## Real-Time Streaming (SSE)

Chat responses and long-running operations stream results to the browser using [Server-Sent Events (SSE)](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events). The browser opens a persistent HTTP connection; the server pushes events as they are produced by the AI model.

SSE is preferred over WebSockets for this use case because it is unidirectional (server to client), works through standard HTTP/2, and requires no special infrastructure.

**Chat SSE endpoint:** `POST /api/v1/chat/send` (returns `text/event-stream`)

---

## AI Chat

BusinessOS includes a built-in chat interface that connects to any LLM provider. The chat has full context of your business data — projects, tasks, clients, calendar, knowledge base — so the AI can answer questions and help you work.

**Supported AI providers:**

BusinessOS is provider-agnostic. The `AI_PROVIDER` environment variable selects the active provider:

- `ollama_local` — local Ollama instance (default, no API key required)
- `anthropic` — Claude models via the Anthropic API
- `openai` — GPT models via the OpenAI API
- `groq` — Llama models via the Groq API
- `xai` — Grok models via the xAI API

### OSA Integration

BusinessOS works with **[OSA](https://github.com/Miosa-osa/OSA)** (Optimal System Agent) — the open-source AI agent. OSA connects to your BusinessOS instance and provides intelligent agent capabilities with full context of your business data. See the [OSA repo](https://github.com/Miosa-osa/OSA) for setup.

The [MIOSA platform](https://miosa.ai) extends OSA with automated skill execution, cross-workspace context, and managed infrastructure.

---

## Authentication

BusinessOS uses JWT-based session authentication with the following security layers:

- **JWT tokens** — signed with `SECRET_KEY`, stored in `httpOnly` cookies
- **CSRF protection** — double-submit cookie pattern on all state-changing requests
- **OAuth 2.0** — Google OAuth for social login and integration connections
- **Redis sessions** — token metadata stored in Redis for fast validation and revocation
- **Role-based access** — workspace-level permissions enforced at the handler layer

---

## Container Sandbox

The terminal feature uses Docker containers for isolated code execution. Each sandbox is an isolated Alpine Linux container with:

- A controlled filesystem with bind-mounted workspace volume
- Network access disabled by default
- Seccomp profile restricting dangerous syscalls
- Resource limits (CPU, memory, process count)
- Automatic cleanup after session ends

The workspace Docker image is built separately from the main application. See [docs/deployment/README.md](../deployment/README.md) for instructions on building it.

---

## See Also

- [Module Overview](../modules/README.md) — all 15 built-in modules
- [API Reference](../api/README.md) — REST API documentation
- [Integrations](../integrations/README.md) — connecting external services
- [Deployment Guide](../deployment/README.md) — running in production
