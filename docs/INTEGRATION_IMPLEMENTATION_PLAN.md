# Integration Infrastructure Implementation Plan

## Executive Summary

This plan lays the groundwork for BusinessOS's integration infrastructure. It focuses on **foundations first** - building the reusable patterns, database schema, and core services that all future integrations will use.

**Goal:** Enable the team to add new integrations quickly by establishing solid infrastructure.

---

## Phase 0: Foundation Infrastructure (Priority: CRITICAL)

### 0.1 Generic Integration Framework

**Objective:** Create a reusable integration framework that all providers will use.

#### Backend Tasks

| Task | Owner | Files | Effort |
|------|-------|-------|--------|
| Create base integration interfaces | Backend | `internal/integrations/types.go` | 2h |
| Create generic OAuth service | Backend | `internal/integrations/oauth.go` | 4h |
| Create integration registry | Backend | `internal/integrations/registry.go` | 2h |
| Create sync worker base | Backend | `internal/workers/sync_base.go` | 4h |
| Database migrations | Backend | `migrations/` | 2h |

#### Database Schema

```sql
-- Generic integration status table
CREATE TABLE integration_connections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'disconnected',
    connected_at TIMESTAMPTZ,
    last_sync_at TIMESTAMPTZ,
    sync_status VARCHAR(20),
    sync_error TEXT,
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, provider)
);

CREATE INDEX idx_integration_connections_user ON integration_connections(user_id);
CREATE INDEX idx_integration_connections_provider ON integration_connections(provider);

-- Sync job tracking
CREATE TABLE sync_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,
    job_type VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    items_processed INT DEFAULT 0,
    items_total INT DEFAULT 0,
    error_message TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_sync_jobs_user_provider ON sync_jobs(user_id, provider);
CREATE INDEX idx_sync_jobs_status ON sync_jobs(status);

-- Integration audit log
CREATE TABLE integration_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    details JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_integration_audit_user ON integration_audit_log(user_id);
```

#### Go Interfaces

```go
// internal/integrations/types.go
package integrations

import (
    "context"
    "time"
)

type Provider interface {
    Name() string
    DisplayName() string
    Icon() string
    GetAuthURL(state string) string
    ExchangeCode(ctx context.Context, code string) (*TokenResponse, error)
    RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error)
    GetConnectionStatus(ctx context.Context, userID string) (*ConnectionStatus, error)
    Disconnect(ctx context.Context, userID string) error
    SupportsSync() bool
    Sync(ctx context.Context, userID string, options SyncOptions) (*SyncResult, error)
}

type TokenResponse struct {
    AccessToken  string
    RefreshToken string
    ExpiresAt    time.Time
    Scopes       []string
    Metadata     map[string]interface{}
}

type ConnectionStatus struct {
    Connected    bool
    ConnectedAt  *time.Time
    AccountName  string
    AccountEmail string
    Scopes       []string
    LastSyncAt   *time.Time
    SyncStatus   string
}

type SyncOptions struct {
    Type      string
    Since     *time.Time
    Resources []string
}

type SyncResult struct {
    Success       bool
    ItemsCreated  int
    ItemsUpdated  int
    ItemsDeleted  int
    Errors        []string
    NextSyncToken string
}
```

---

### 0.2 Integration Settings UI

**Objective:** Create the frontend infrastructure for managing integrations.

#### Frontend Tasks

| Task | Owner | Files | Effort |
|------|-------|-------|--------|
| Create integrations API client | Frontend | `src/lib/api/integrations/client.ts` | 2h |
| Create integrations store | Frontend | `src/lib/stores/integrationsStore.ts` | 3h |
| Create IntegrationCard component | Frontend | `src/lib/components/settings/IntegrationCard.svelte` | 3h |
| Create IntegrationsPage | Frontend | `src/routes/(app)/settings/integrations/+page.svelte` | 4h |
| Create connection modals | Frontend | `src/lib/components/settings/ConnectModal.svelte` | 3h |

#### Store Design

```typescript
// src/lib/stores/integrationsStore.ts
import { writable, derived } from 'svelte/store';

export interface Integration {
  provider: string;
  displayName: string;
  icon: string;
  category: 'calendar' | 'communication' | 'tasks' | 'crm' | 'storage' | 'ai';
  status: 'connected' | 'disconnected' | 'error';
  connectedAt?: string;
  accountName?: string;
  lastSyncAt?: string;
  syncStatus?: 'idle' | 'syncing' | 'error';
  features: string[];
}

function createIntegrationsStore() {
  const { subscribe, set, update } = writable<{
    integrations: Integration[];
    loading: boolean;
    error: string | null;
  }>({
    integrations: [],
    loading: false,
    error: null
  });

  return {
    subscribe,
    async load() { /* fetch from API */ },
    async connect(provider: string) { /* redirect to OAuth */ },
    async disconnect(provider: string) { /* call DELETE endpoint */ },
    async triggerSync(provider: string) { /* call sync endpoint */ }
  };
}

export const integrationsStore = createIntegrationsStore();
```

---

### 0.3 Unified API Endpoints

```
GET    /api/integrations/status              # All integration statuses
GET    /api/integrations/:provider/status    # Single provider status
GET    /api/integrations/:provider/auth      # Get OAuth URL
GET    /api/integrations/:provider/callback  # OAuth callback
DELETE /api/integrations/:provider           # Disconnect
POST   /api/integrations/:provider/sync      # Trigger sync
GET    /api/integrations/:provider/sync      # Get sync status
GET    /api/integrations/:provider/logs      # Audit logs
```

---

## Phase 1: File Import System (Priority: HIGH)

### 1.1 Database & Core Service

| Task | Owner | Files | Effort |
|------|-------|-------|--------|
| Create file_imports migration | Backend | `migrations/` | 1h |
| Create imported_conversations migration | Backend | `migrations/` | 1h |
| Create imported_messages migration | Backend | `migrations/` | 1h |
| Create ImportService | Backend | `internal/services/imports.go` | 4h |
| Create file storage service | Backend | `internal/services/file_storage.go` | 2h |

### 1.2 Parsers

| Task | Owner | Files | Effort |
|------|-------|-------|--------|
| Create parser interface | Backend | `internal/imports/parser.go` | 1h |
| Implement ChatGPT parser | Backend | `internal/imports/chatgpt.go` | 3h |
| Implement Claude parser | Backend | `internal/imports/claude.go` | 2h |
| Implement Perplexity parser | Backend | `internal/imports/perplexity.go` | 2h |
| Add parser tests | Backend | `internal/imports/*_test.go` | 3h |

### 1.3 API Endpoints

| Task | Owner | Files | Effort |
|------|-------|-------|--------|
| POST /api/imports/upload | Backend | `internal/handlers/imports.go` | 2h |
| GET /api/imports | Backend | handlers | 1h |
| GET /api/imports/:id | Backend | handlers | 1h |
| GET /api/imports/:id/progress | Backend | handlers | 1h |
| DELETE /api/imports/:id | Backend | handlers | 1h |

### 1.4 Frontend

| Task | Owner | Files | Effort |
|------|-------|-------|--------|
| Create imports API client | Frontend | `src/lib/api/imports.ts` | 2h |
| Create FileDropzone component | Frontend | `src/lib/components/imports/FileDropzone.svelte` | 2h |
| Create ImportProgress component | Frontend | `src/lib/components/imports/ImportProgress.svelte` | 2h |
| Create ImportsPage | Frontend | `src/routes/(app)/settings/imports/+page.svelte` | 4h |

---

## Phase 2: Refactor Existing Integrations (Priority: MEDIUM)

### 2.1 Migrate to New Framework

| Task | Owner | Files | Effort |
|------|-------|-------|--------|
| Refactor Google Calendar | Backend | `internal/integrations/google/calendar.go` | 3h |
| Refactor Slack | Backend | `internal/integrations/slack/slack.go` | 3h |
| Refactor Notion | Backend | `internal/integrations/notion/notion.go` | 3h |
| Update handlers to use registry | Backend | `internal/handlers/integrations.go` | 2h |

### 2.2 Display in New UI

| Task | Owner | Files | Effort |
|------|-------|-------|--------|
| Add integration cards to UI | Frontend | Update IntegrationsPage | 3h |
| Add sync status displays | Frontend | Update SyncStatus | 2h |

---

## Phase 3: New OAuth Integrations (Priority: MEDIUM)

### 3.1 Gmail & Google Drive

| Task | Owner | Files | Effort |
|------|-------|-------|--------|
| Add Gmail scopes to Google OAuth | Backend | `internal/integrations/google/oauth.go` | 1h |
| Implement Gmail service | Backend | `internal/integrations/google/gmail.go` | 4h |
| Implement Drive service | Backend | `internal/integrations/google/drive.go` | 4h |
| Create Gmail MCP tools | Backend | `internal/services/mcp_gmail.go` | 3h |
| Create Drive MCP tools | Backend | `internal/services/mcp_drive.go` | 3h |

### 3.2 ClickUp Integration

| Task | Owner | Files | Effort |
|------|-------|-------|--------|
| Create ClickUp provider | Backend | `internal/integrations/clickup/clickup.go` | 4h |
| Create task sync service | Backend | `internal/integrations/clickup/sync.go` | 4h |
| Create field mapping | Backend | `internal/integrations/clickup/mapping.go` | 2h |
| Create MCP tools | Backend | `internal/services/mcp_clickup.go` | 3h |

### 3.3 HubSpot Integration

| Task | Owner | Files | Effort |
|------|-------|-------|--------|
| Create HubSpot provider | Backend | `internal/integrations/hubspot/hubspot.go` | 4h |
| Create contact sync | Backend | `internal/integrations/hubspot/contacts.go` | 3h |
| Create deal sync | Backend | `internal/integrations/hubspot/deals.go` | 3h |
| Create MCP tools | Backend | `internal/services/mcp_hubspot.go` | 3h |

---

## Phase 4: Knowledge Enrichment (Priority: MEDIUM)

### 4.1 AI Processing Pipeline

| Task | Owner | Files | Effort |
|------|-------|-------|--------|
| Create enrichment service | Backend | `internal/services/enrichment.go` | 4h |
| Implement summary generation | Backend | enrichment.go | 2h |
| Implement topic extraction | Backend | enrichment.go | 2h |
| Implement entity extraction | Backend | enrichment.go | 2h |
| Implement knowledge extraction | Backend | enrichment.go | 3h |

### 4.2 Vector Embeddings

| Task | Owner | Files | Effort |
|------|-------|-------|--------|
| Add pgvector extension | Backend | Migration | 1h |
| Create embedding service | Backend | `internal/services/embeddings.go` | 3h |
| Implement semantic search | Backend | `internal/services/search.go` | 4h |

---

## Team Assignment

### Backend (Pedro) - Sprint 1
- Phase 0.1: Generic Integration Framework
- Phase 0.3: Unified API Endpoints
- Phase 1.1: File Import Database

### Backend - Sprint 2
- Phase 1.2: Parsers (ChatGPT, Claude)
- Phase 1.3: Import API Endpoints

### Backend - Sprint 3
- Phase 2.1: Refactor existing integrations
- Phase 3.1: Gmail & Drive

### Frontend - Sprint 1
- Phase 0.2: Integration Settings UI
- Integration cards and modals

### Frontend - Sprint 2
- Phase 1.4: File Import UI
- FileDropzone, Progress components

---

## File Structure

```
backend-go/
├── internal/
│   ├── integrations/
│   │   ├── types.go
│   │   ├── registry.go
│   │   ├── oauth.go
│   │   ├── google/
│   │   ├── slack/
│   │   ├── notion/
│   │   ├── clickup/
│   │   └── hubspot/
│   ├── imports/
│   │   ├── service.go
│   │   ├── parser.go
│   │   ├── chatgpt.go
│   │   └── claude.go
│   └── handlers/
│       ├── integrations.go
│       └── imports.go

frontend/
├── src/lib/
│   ├── api/
│   │   ├── integrations.ts
│   │   └── imports.ts
│   ├── stores/
│   │   ├── integrationsStore.ts
│   │   └── importsStore.ts
│   └── components/
│       └── settings/integrations/
└── src/routes/(app)/settings/
    ├── integrations/+page.svelte
    └── imports/+page.svelte
```

---

## Success Criteria

### Phase 0 Complete:
- [ ] Integration registry works with existing providers
- [ ] Settings UI shows all integrations with status
- [ ] Connect/disconnect works from UI

### Phase 1 Complete:
- [ ] Can upload ChatGPT/Claude exports
- [ ] Import progress shows real-time
- [ ] Can view/delete imports

### Phase 2 Complete:
- [ ] All 3 existing integrations use new registry
- [ ] UI shows sync status

### Phase 3 Complete:
- [ ] Gmail/Drive working
- [ ] ClickUp task sync working
- [ ] HubSpot contact sync working

---

## Dependencies

```
Phase 0 (Foundation) → Phase 1 (File Imports)
                    → Phase 2 (Refactor)
                              → Phase 3 (New OAuth)
Phase 1 + Phase 3 → Phase 4 (Enrichment)
```

**Critical Path:** Phase 0 → Phase 1 → Phase 4

---

## Next Actions

1. **This Week:** Start Phase 0.1 (Backend framework interfaces)
2. **This Week:** Start Phase 0.2 (Frontend UI skeleton)
3. **Next Week:** Complete Phase 0, begin Phase 1
4. **Sprint 2:** File imports functional
5. **Sprint 3:** First new OAuth integration
