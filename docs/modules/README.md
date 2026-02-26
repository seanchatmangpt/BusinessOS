# Module Overview

BusinessOS is organized into 15 modules, each covering a distinct area of business operations. This page gives contributors a map of the codebase: which modules exist, what they do, and where their backend handlers and frontend routes live.

All modules share the same backend architecture (`Handler -> Service -> Repository`) and the same frontend conventions (`+page.svelte`, `+page.server.ts`, Svelte stores).

---

## Module Index

### 1. Dashboard

**Purpose:** Entry point after login. Displays a configurable widget-based overview of the user's workspace — recent activity, task summary, upcoming calendar events, and AI-generated insights.

| Layer | Location |
|-------|---------|
| Backend handlers | `backend/internal/handlers/dashboard.go`, `dashboard_handlers.go` |
| Frontend route | `frontend/src/routes/(app)/dashboard/+page.svelte` |

---

### 2. Chat / OSA Interface

**Purpose:** The primary interface for interacting with the OSA AI agent. Supports streaming responses, multiple focus modes (research, analyze, write, code, plan), slash commands, agent mentions, document attachments, and Chain of Thought reasoning display.

| Layer | Location |
|-------|---------|
| Backend handlers | `backend/internal/handlers/chat.go`, `chat_v2.go` |
| Agent routing | `backend/internal/agents/orchestration.go`, `intent_router_v2.go` |
| Frontend route | `frontend/src/routes/(app)/chat/+page.svelte` |
| API client | `frontend/src/lib/api/conversations.ts` |

---

### 3. Projects

**Purpose:** Project management with tasks, members, statuses, and timeline tracking. Projects can be linked to clients and can have AI-generated status summaries.

| Layer | Location |
|-------|---------|
| Backend handlers | `backend/internal/handlers/projects.go`, `project_access_handlers.go` |
| Specialist agent | `backend/internal/agents/project/agent.go` |
| Frontend routes | `frontend/src/routes/(app)/projects/` |

---

### 4. Tasks

**Purpose:** Task tracking within and outside of projects. Supports assignments, due dates, priorities, status workflows, and AI-assisted task breakdown from natural language descriptions.

| Layer | Location |
|-------|---------|
| Backend handlers | `backend/internal/handlers/` (task-related) |
| Specialist agent | `backend/internal/agents/task/agent.go` |
| Frontend route | `frontend/src/routes/(app)/tasks/+page.svelte` |

---

### 5. Clients

**Purpose:** Client relationship management for agencies and service businesses. Tracks client contact details, associated projects, and activity history.

| Layer | Location |
|-------|---------|
| Backend handlers | `backend/internal/handlers/clients.go` |
| Specialist agent | `backend/internal/agents/client/agent.go` |
| Frontend routes | `frontend/src/routes/(app)/clients/` |

---

### 6. CRM

**Purpose:** Contact relationship management including companies, deals pipeline, and contact activity tracking. Integrates with HubSpot to sync CRM data bidirectionally.

| Layer | Location |
|-------|---------|
| Backend handlers | `backend/internal/handlers/crm.go`, `deals.go` |
| Database queries | `backend/internal/database/queries/crm.sql` |
| Frontend routes | `frontend/src/routes/(app)/crm/` |

---

### 7. Calendar

**Purpose:** Calendar view integrating events from connected Google Calendar and Microsoft 365 accounts. Supports AI-assisted scheduling and meeting notes.

| Layer | Location |
|-------|---------|
| Backend handlers | `backend/internal/handlers/calendar.go`, `calendar_scheduling_handler.go` |
| Database queries | `backend/internal/database/queries/calendar.sql` |
| Frontend route | `frontend/src/routes/(app)/communication/calendar/+page.svelte` |

---

### 8. Communication Hub

**Purpose:** Unified inbox for email (Gmail, Microsoft 365), Slack channels, and notifications. The communication layout groups these views under a shared shell.

| Layer | Location |
|-------|---------|
| Backend handlers | `backend/internal/handlers/email_handler.go` |
| Frontend routes | `frontend/src/routes/(app)/communication/` |

---

### 9. Knowledge Base

**Purpose:** A searchable document library backed by pgvector semantic search. Documents are chunked, embedded, and stored for use in RAG (retrieval-augmented generation) when chatting with OSA.

| Layer | Location |
|-------|---------|
| Backend handlers | `backend/internal/handlers/rag.go`, `document_handler.go`, `embedding.go` |
| Specialist agent | `backend/internal/agents/document/agent.go` |
| Frontend routes | `frontend/src/routes/(app)/knowledge/`, `knowledge-v2/` |

---

### 10. Nodes

**Purpose:** A graph-based knowledge structure for linking business entities (people, companies, documents, projects) together. Nodes provide a visual relationship map for complex information.

| Layer | Location |
|-------|---------|
| Backend handlers | `backend/internal/handlers/nodes.go` |
| Database queries | `backend/internal/database/queries/nodes.sql` |
| Frontend routes | `frontend/src/routes/(app)/nodes/` |

---

### 11. Generated Apps

**Purpose:** Manages apps that OSA generated via the BUILD mode. Users can view, preview, and deploy AI-generated mini-applications directly from this module.

| Layer | Location |
|-------|---------|
| Backend handlers | `backend/internal/handlers/osa_api.go`, `osa_deployment.go` |
| Frontend routes | `frontend/src/routes/(app)/generated-apps/` |

---

### 12. Terminal

**Purpose:** An in-browser terminal connected to an isolated Docker sandbox. Users can run shell commands, install packages, and test code in a secure container environment.

| Layer | Location |
|-------|---------|
| Backend handlers | `backend/internal/handlers/sandbox.go` |
| Container management | `backend/internal/container/` |
| Frontend route | `frontend/src/routes/(app)/terminal/+page.svelte` |

---

### 13. Integrations

**Purpose:** Manages OAuth connections to external services (Google, Microsoft, Slack, Notion, Linear, HubSpot, ClickUp, Airtable, Fathom). Displays connection status and allows users to connect, disconnect, and re-authorize services.

| Layer | Location |
|-------|---------|
| Backend handlers | `backend/internal/handlers/integrations.go`, `oauth_integrations.go`, `integration_router.go` |
| Database queries | `backend/internal/database/queries/integrations.sql` |
| Frontend routes | `frontend/src/routes/(app)/integrations/` |

---

### 14. Settings

**Purpose:** Workspace and account settings including profile, AI provider configuration, reasoning templates, and workspace management. Also provides team member invitation and management.

| Layer | Location |
|-------|---------|
| Backend handlers | `backend/internal/handlers/settings.go`, `team.go`, `profile.go`, `ai_config.go` |
| Frontend routes | `frontend/src/routes/(app)/settings/` |

---

### 15. Analytics

**Purpose:** Usage analytics dashboard showing token consumption, AI request volume, integration activity, and workspace health metrics.

| Layer | Location |
|-------|---------|
| Backend handlers | `backend/internal/handlers/analytics_handlers.go`, `usage.go` |
| Database queries | `backend/internal/database/queries/analytics_enhancements.sql` |
| Frontend routes | `frontend/src/routes/(app)/usage/+page.svelte` |

---

## Additional Features

The following features span multiple modules and are not isolated to a single route:

| Feature | Description | Key Files |
|---------|-------------|-----------|
| **Daily Logs** | Daily work journal with AI summary | `handlers/daily_logs.go`, `routes/(app)/daily/` |
| **Voice Notes** | Audio recording with AI transcription | `handlers/voice_notes.go`, `routes/(app)/voice-notes/` |
| **Notifications** | In-app and push (Web Push) notifications | `handlers/notification_handlers.go`, `routes/(app)/notifications/` |
| **Flexible Tables** | User-defined data tables (like Airtable views) | `handlers/tables.go`, `routes/(app)/tables/` |
| **Artifacts** | Versioned output artifacts from AI sessions | `handlers/artifacts.go` |
| **Workflows** | Saved automation sequences | `handlers/osa_workflows.go`, `handlers/workflows.go` |

---

## Adding a New Module

To add a new module to BusinessOS:

1. **Database:** Add a new SQL migration file in `backend/internal/database/migrations/` following the numeric naming convention.
2. **Queries:** Write SQL queries in `backend/internal/database/queries/yourmodule.sql` and run `sqlc generate`.
3. **Repository:** The generated SQLC code in `backend/internal/database/sqlc/` provides typed query functions.
4. **Service:** Add `backend/internal/services/yourmodule.go` with business logic.
5. **Handler:** Add `backend/internal/handlers/yourmodule.go` with HTTP endpoints and register them in the main router.
6. **Frontend:** Create `frontend/src/routes/(app)/yourmodule/+page.svelte` and `+page.server.ts`.

Follow the existing patterns in any established module (e.g., `projects.go` + `/projects/+page.svelte`) as the template.
