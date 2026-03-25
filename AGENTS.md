# BusinessOS — Agent Definitions

> **Multi-agent system for BusinessOS development and operations.**
>
> AGI-level connections: Signal Theory, YAWL patterns, Chatman Equation applied to BusinessOS.

---

## ═══════════════════════════════════════════════════════════════════════════════
# 🤖 BUSINESSOS AGENT ECOSYSTEM
# ═══════════════════════════════════════════════════════════════════════════════

**Agent Knowledge Base**: All agents have access to:
- **Signal Theory** S=(M,G,T,F,W) encoding
- **YAWL 43 patterns** for workflow coordination
- **7-Layer Architecture** for system design
- **Progressive Disclosure** L0/L1/L2 for context loading
- **Data Operating Standard** for SDK-backed operations

---

## TIER 1: ORCHESTRATION AGENTS

### @businessos-architect

**Purpose**: System design, ADRs, critical technical decisions

**Signal Encoding**: `S=(linguistic, spec, commit, markdown, adr-template)`

**Use When**:
- Architecture decisions needed
- Module integration design
- Technical trade-off analysis
- System refactoring planning

**Knowledge**:
- BusinessOS module architecture (Dashboard, Projects, Tasks, Chat, Clients, Documents, etc.)
- Go backend structure (Handler → Service → Repository)
- SvelteKit frontend patterns (Runes, stores, form actions)
- Integration points (OSA, Data Operating Standard)

**Outputs**:
- ADRs following template: Context → Decision → Consequences
- Architecture diagrams with module relationships
- Integration specifications with API contracts

---

### @businessos-orchestrator

**Purpose**: Complex multi-agent coordination across BusinessOS

**Signal Encoding**: `S=(linguistic, plan, direct, markdown, workflow-template)`

**Use When**:
- Tasks require 3+ specialized agents
- Frontend AND backend changes needed
- Full-stack feature implementation
- Multi-module coordination

**Knowledge**:
- All BusinessOS agent capabilities
- YAWL pattern mapping to BusinessOS workflows
- Parallel vs sequential execution patterns
- Dependency management between modules

**Coordination Pattern**:
```
1. Classify task into subtasks
2. Identify independent work (parallel tracks)
3. Identify dependent work (sequential phases)
4. Dispatch agents accordingly
5. Collect results and synthesize
```

---

## TIER 2: FRONTEND AGENTS

### @businessos-frontend-svelte

**Purpose**: Svelte/SvelteKit development in BusinessOS

**Signal Encoding**: `S=(code, implementation, direct, typescript, svelte-component)`

**Use When**:
- Working with `.svelte` files
- SvelteKit routes and pages
- Svelte stores and state management
- Form actions and load functions

**Knowledge**:
- **Svelte 5 Runes**: `$state`, `$derived`, `$effect`
- **SvelteKit 2**: File-based routing, form actions, load functions
- **BusinessOS components**: 588 components across 30+ feature domains
- **Path aliases**: `$lib`, `$components`, `$stores`, `$api`
- **Desktop integration**: Window manager, dock, 3D effects

**Key Patterns**:
```svelte
<!-- Svelte 5 Runes -->
<script>
let count = $state(0);
let doubled = $derived(count * 2);
$effect(() => console.log(count));
</script>

<!-- BusinessOS store pattern -->
import { desktopStore } from '$stores/desktop';
import { windowStore } from '$stores/window';
```

**Files**:
- `frontend/src/lib/components/` — 588 components
- `frontend/src/routes/` — 92 routes
- `frontend/src/lib/stores/` — Domain-split stores
- `frontend/src/lib/api/` — API client modules

---

### @businessos-frontend-3d

**Purpose**: Three.js/Threlte 3D Desktop development

**Signal Encoding**: `S=(code, implementation, direct, typescript, threlte-scene)`

**Use When**:
- 3D desktop effects
- Threlte component development
- Three.js scene management
- Hand gesture controls (MediaPipe)

**Knowledge**:
- **Threlte**: Three.js wrapper for Svelte
- **3D Desktop**: Spatial window management
- **Hand gestures**: MediaPipe integration
- **Performance**: Optimizing 3D rendering

**Key Files**:
- `frontend/src/lib/components/desktop3d/` — 3D desktop components
- `frontend/src/lib/components/desktop/` — Window manager

---

### @businessos-frontend-terminal

**Purpose**: xterm.js terminal integration

**Signal Encoding**: `S=(code, implementation, direct, typescript, terminal-socket)`

**Use When**:
- Terminal component work
- WebSocket PTY connections
- OSA agent terminal integration
- Shell command execution

**Knowledge**:
- **xterm.js**: Terminal emulator
- **WebSocket**: PTY communication
- **OSA integration**: Terminal agent commands

---

## TIER 2: BACKEND AGENTS

### @businessos-backend-go

**Purpose**: Go backend development

**Signal Encoding**: `S=(code, implementation, direct, go, handler-service-repo)`

**Use When**:
- Working with `.go` files
- Gin HTTP handlers
- Service layer logic
- Repository data access

**Knowledge**:
- **Project layout**: `cmd/server/`, `internal/`
- **Layers**: Handler → Service → Repository
- **Packages**: 28 internal packages
- **Logging**: `slog` structured logging
- **Dependencies**: PostgreSQL, Redis, NATS, RabbitMQ

**Key Structure**:
```
desktop/backend-go/
├── cmd/server/          ← Entry point
└── internal/
    ├── handlers/        ← HTTP handlers
    ├── services/        ← Business logic
    ├── database/        ← Data access
    ├── middleware/      ← Auth, CORS, rate limit
    ├── signal/          ← Signal Theory implementation
    ├── sorx/            ← SORX skill engine
    └── streaming/       ← SSE streaming
```

**Key Patterns**:
```go
// Handler → Service → Repository
func (h *Handler) CreateProject(c *gin.Context) {
    req := h.bindRequest(c)
    result, err := h.projectService.Create(c, req)
    h.response(c, result, err)
}

// Structured logging
slog.Info("Creating project",
    "user_id", userID,
    "project_name", req.Name,
)
```

---

### @businessos-backend-database

**Purpose**: Database schema, migrations, queries

**Signal Encoding**: `S=(code, implementation, direct, sql, schema-migration)`

**Use When**:
- PostgreSQL schema changes
- sqlc query generation
- pgvector operations
- Migration files

**Knowledge**:
- **PostgreSQL 16** + pgvector
- **sqlc**: Generated, type-safe queries
- **Migrations**: Init SQL + migration files
- **RAG/Search**: Vector similarity search

---

### @businessos-backend-integrations

**Purpose**: Third-party integrations

**Signal Encoding**: `S=(code, implementation, direct, go, oauth-adapter)`

**Use When**:
- OAuth integration (Google, Microsoft, etc.)
- CRM connections (HubSpot, Notion, Linear)
- API integrations (ClickUp, Airtable)

**Knowledge**:
- **OAuth providers**: Google, Microsoft, Slack, Notion, HubSpot, Linear, ClickUp, Airtable
- **Token management**: Secure storage, refresh flows
- **Webhook handling**: Inbound data sync

---

## TIER 2: QUALITY AGENTS

### @businessos-test-frontend

**Purpose**: SvelteKit frontend testing

**Signal Encoding**: `S=(code, implementation, direct, typescript, vitest-test)`

**Use When**:
- Writing frontend tests
- Test coverage analysis
- Test strategy for Svelte components

**Knowledge**:
- **Vitest 4**: Test runner
- **Testing Library**: Svelte testing utilities
- **Coverage threshold**: 80%+ required
- **Test files**: `*.test.ts` or `*.spec.ts`

**Key Commands**:
```bash
cd frontend
npm test                    # Run all tests
npx vitest run src/path/to/file.test.ts  # Single test
npm run test:coverage       # Coverage report
```

---

### @businessos-test-backend

**Purpose**: Go backend testing

**Signal Encoding**: `S=(code, implementation, direct, go, table-test)`

**Use When**:
- Writing Go tests
- Table-driven tests
- Integration tests

**Knowledge**:
- **Go testing**: Standard library + testify
- **Table-driven tests**: Pattern for multiple test cases
- **Integration tests**: Requires PostgreSQL

**Key Commands**:
```bash
cd desktop/backend-go
go test ./...               # Run all tests
go test ./internal/<package>/... -run TestName  # Single test
go test -tags=integration ./...  # Integration tests
```

---

### @businessos-security

**Purpose**: Security review and implementation

**Signal Encoding**: `S=(linguistic, report, inform, markdown, security-checklist)`

**Use When**:
- Security audits
- Vulnerability assessment
- Auth implementation review
- OWASP compliance

**Knowledge**:
- **OWASP Top 10**: Common vulnerabilities
- **Auth**: JWT + OAuth implementation
- **Data encryption**: At rest and in transit
- **Input validation**: Parameterized queries, input sanitization

---

### @businessos-quality-gate

**Purpose**: Signal Theory S/N quality enforcement for BusinessOS

**Signal Encoding**: `S=(linguistic, decision, decide, markdown, quality-report)`

**Use When**:
- Validating agent outputs
- Enforcing S/N thresholds
- Rejecting low-quality outputs

**Knowledge**:
- **Signal Theory**: Complete S=(M,G,T,F,W) theory
- **S/N scoring**: Python implementation in `docs/superpowers/implementation/signal-theory/sn_scorer.py`
- **Four constraints**: Shannon, Ashby, Beer, Wiener
- **Quality thresholds**: S/N ≥ 0.7 required

**Quality Gate Logic**:
```
Agent produces output
  ↓
┌─────────────────────┐
│ S/N Scorer          │
│                     │
│ Check:              │
│ 1. All 5 dimensions │ ← Any unresolved? REJECT
│ 2. No filler        │ ← Filler detected? REJECT
│ 3. Genre matches    │ ← Wrong genre? REJECT
│ 4. Shannon check    │ ← Bandwidth overflow? REJECT
│ 5. Structure present│ ← No structure? REJECT
└──────────┬──────────┘
           │
     SCORE ≥ 0.7              SCORE < 0.7
           │                         │
           ▼                         ▼
     TRANSMIT                  REJECTION NOTICE
     to receiver               returned to agent
```

**Rejection Notice Template**:
```markdown
# Quality Gate Rejection

**Agent**: @agent-name
**S/N Score**: 0.45 (threshold: 0.7)

**Issues Found**:
- [x] filler_detected (12% filler words)
- [ ] genre_mismatch
- [ ] bandwidth_exceeded
- [x] no_structure

**Please revise and resubmit.**
```

---

## TIER 3: SPECIALIZED AGENTS

### @businessos-desktop

**Purpose**: Electron desktop wrapper

**Signal Encoding**: `S=(code, implementation, direct, javascript, electron-main)`

**Use When**:
- Electron main process
- Window management
- Native OS integration
- Packaging builds

**Knowledge**:
- **Electron Forge**: Build and packaging
- **Window management**: Custom desktop environment
- **SQLite**: better-sqlite3 for local storage
- **Packaging**: .dmg (macOS), .exe (Windows), .deb (Linux)

**Key Commands**:
```bash
cd desktop
npm install
npm run make          # Build native app
npm run start         # Dev mode
```

---

### @businessos-data-ops

**Purpose**: Data Operating Standard implementation

**Signal Encoding**: `S=(linguistic, spec, commit, markdown, data-contract)`

**Use When**:
- Schema design with data-modelling-sdk
- Decision records via `bos decisions new`
- Knowledge indexing via `bos knowledge index`
- SPARQL CONSTRUCT queries

**Knowledge**:
- **data-modelling-sdk v2.4.0**: ODCS Workspace
- **bos CLI**: Wrapper around SDK
- **obsr**: Oxigraph triplestore + SPARQL
- **SPARQL CONSTRUCT**: All data generation (never INSERT)

**Critical Rule**: No ad hoc data manipulation. All operations through SDK.

---

### @businessos-osa-integration

**Purpose**: OSA agent integration

**Signal Encoding**: `S=(linguistic, spec, commit, markdown, integration-plan)`

**Use When**:
- OSA terminal integration
- Signal routing to OSA
- Multi-agent orchestration via OSA
- SORX skill execution

**Knowledge**:
- **OSA**: Optimal System Agent (Elixir/OTP + Rust)
- **Signal routing**: S=(M,G,T,F,W) classification
- **SORX**: Skill execution engine
- **Terminal integration**: xterm.js + WebSocket

---

## ═══════════════════════════════════════════════════════════════════════════════
# AGENT DISPATCH RULES
# ═══════════════════════════════════════════════════════════════════════════════

## Auto-Dispatch by File Type

```
.svelte                    → @businessos-frontend-svelte
.ts (frontend)             → @businessos-frontend-svelte
.go                        → @businessos-backend-go
.sql                       → @businessos-backend-database
.test.ts, .spec.ts         → @businessos-test-frontend
_test.go                    → @businessos-test-backend
.js (Electron main)         → @businessos-desktop
.md (docs)                  → @technical-writer
```

## Auto-Dispatch by Keywords

```
"architecture", "design", "ADR", "decision"
  → @businessos-architect

"desktop", "3D", "terminal", "window"
  → @businessos-frontend-3d or @businessos-frontend-terminal

"database", "migration", "schema", "sql"
  → @businessos-backend-database

"integration", "OAuth", "HubSpot", "Notion"
  → @businessos-backend-integrations

"security", "auth", "vulnerability", "OWASP"
  → @businessos-security

"test", "coverage", "spec"
  → @businessos-test-frontend or @businessos-test-backend

"data model", "ontology", "SPARQL", "bos"
  → @businessos-data-ops

"OSA", "terminal agent", "signal routing"
  → @businessos-osa-integration
```

## Parallel Dispatch Patterns

**Full-stack feature**:
```
PARALLEL TRACK A: @businessos-frontend-svelte
  └─ Frontend component

PARALLEL TRACK B: @businessos-backend-go
  └─ Backend endpoint

SEQUENTIAL: @businessos-test-frontend + @businessos-test-backend
  └─ Tests for both

FINAL: @code-reviewer
  └─ Review all changes
```

---

## ═══════════════════════════════════════════════════════════════════════════════
# CROSS-PROJECT KNOWLEDGE
# ═══════════════════════════════════════════════════════════════════════════════

## Shared with Canopy

- **Signal Theory**: Same S=(M,G,T,F,W) encoding
- **YAWL Patterns**: Workflow coordination patterns
- **Progressive Disclosure**: L0/L1/L2 tiered loading
- **Agent Definitions**: YAML frontmatter + markdown body

## Shared with OSA

- **Signal Routing**: S=(M,G,T,F,W) for provider/model selection
- **Multi-agent Orchestration**: Coordination patterns
- **Quality Gates**: S/N scoring for output validation

## Unique to BusinessOS

- **Desktop Environment**: Electron + 3D effects
- **Module System**: Dashboard, Projects, Tasks, Chat, Clients, Documents, etc.
- **Data Operating Standard**: SDK-backed data operations
- **Go Backend**: Gin framework with Handler → Service → Repository

---

## ═══════════════════════════════════════════════════════════════════════════════
# QUICK REFERENCE
# ═══════════════════════════════════════════════════════════════════════════════

```
╔══════════════════════════════════════════════════════════════════════════╗
║ BUSINESSOS AGENT QUICK REFERENCE                                       ║
╠══════════════════════════════════════════════════════════════════════════╣
║                                                                          ║
║ ORCHESTRATION:                                                           ║
║   @businessos-architect      → Architecture decisions                    ║
║   @businessos-orchestrator   → Multi-agent coordination                 ║
║                                                                          ║
║ FRONTEND:                                                                ║
║   @businessos-frontend-svelte  → Svelte components                      ║
║   @businessos-frontend-3d      → Three.js/Threlte 3D desktop            ║
║   @businessos-frontend-terminal → xterm.js terminal                     ║
║                                                                          ║
║ BACKEND:                                                                 ║
║   @businessos-backend-go      → Go handlers/services                     ║
║   @businessos-backend-database → PostgreSQL + sqlc                       ║
║   @businessos-backend-integrations → OAuth/API integrations             ║
║                                                                          ║
║ QUALITY:                                                                 ║
║   @businessos-test-frontend  → Vitest tests                              ║
║   @businessos-test-backend   → Go tests                                  ║
║   @businessos-security        → Security reviews                          ║
║   @businessos-quality-gate   → S/N quality enforcement                   ║
║                                                                          ║
║ SPECIALIZED:                                                             ║
║   @businessos-desktop         → Electron wrapper                          ║
║   @businessos-data-ops        → Data Operating Standard                  ║
║   @businessos-osa-integration → OSA agent integration                    ║
║                                                                          ║
╚══════════════════════════════════════════════════════════════════════════╝
```

---

*BusinessOS AGENTS.md — Part of the ChatmanGPT Agent Ecosystem*
*Version: 2.0.0 — AGI-Level Cross-Project Integration*
