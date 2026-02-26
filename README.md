# BusinessOS

> Open-source business operating system. Desktop-class productivity — CRM, Projects, Tasks, Calendar, Knowledge Base, and 9+ integrations.

## What is BusinessOS?

BusinessOS is a **complete desktop operating system** for running your business. It runs in the browser with a full desktop environment (windows, terminal, file browser, 3D mode), and comes with everything you need to manage your business out of the box.

**All modules included.** CRM, project management, task tracking, calendar, knowledge base, team management, communication, dashboards, and more.

**Bring your own AI.** Works with Anthropic Claude, OpenAI, Ollama (local), Groq, or xAI. Plug in your API key and go.

**Run it your way.** Self-host, run locally, or deploy to cloud. Docker + Cloud Run ready.

## Modules

| Module | What it does |
|--------|-------------|
| **Desktop Shell** | Window manager, taskbar, dock, file browser, spotlight search, 3D mode |
| **Terminal** | Full terminal emulator with Docker sandbox isolation |
| **CRM** | Clients, contacts, interactions, deals, sales pipelines |
| **Projects** | Project management with members, statuses, notes, analytics |
| **Tasks** | Subtasks, dependencies, assignments, burndown charts |
| **Team** | Members, capacity tracking, skills, org chart |
| **Calendar** | Google Calendar sync, scheduling, meeting types |
| **Communication** | Email (Gmail), channels |
| **Knowledge Base** | Documents, contexts, templates, sharing, graph visualization |
| **Dashboard** | Widgets, analytics, quick actions, activity feed |
| **Daily Log** | Reflections, energy levels, focus tracking |
| **Apps** | App marketplace, custom module builder |
| **Tables** | Spreadsheet-like data with views (table, kanban, gallery, timeline) |
| **Nodes** | Knowledge networking and linking |
| **Templates** | Template library with multiple view types |

## Integrations

Connect your existing tools:

| Service | What syncs |
|---------|-----------|
| **Google Workspace** | Calendar, Gmail, Drive, Contacts, Sheets, Docs |
| **Microsoft 365** | Calendar, Mail, OneDrive, Teams, To-Do |
| **Slack** | Channels, messages, users |
| **Notion** | Databases, pages |
| **Linear** | Issues, projects, teams |
| **HubSpot** | Contacts, companies, deals |
| **ClickUp** | Tasks, spaces, folders |
| **Airtable** | Bases, tables, records |
| **Fathom** | Website analytics |

## AI Agent — OSA

BusinessOS works with **[OSA](https://github.com/Miosa-osa/OSA)** (Optimal System Agent) — the open-source AI agent. OSA connects to your BusinessOS instance and gives you an intelligent assistant with full context of your business data.

### Setting up OSA

1. Clone the [OSA repo](https://github.com/Miosa-osa/OSA)
2. Point it at your BusinessOS backend URL
3. Add your LLM API key (Claude, GPT, Ollama, etc.)
4. OSA will have context of your projects, tasks, clients, calendar, and everything in BusinessOS

OSA runs locally on your machine. Your data stays yours.

### Want more?

The [MIOSA platform](https://miosa.ai) extends OSA with:
- Automated skill execution across all your integrations
- Cross-workspace context sharing
- Advanced orchestration and proactive agent behaviors
- Managed cloud infrastructure

## Tech Stack

```
Frontend:  SvelteKit 2 + Svelte 5 + TypeScript + Tailwind CSS
Backend:   Go 1.24 + Gin + PostgreSQL + Redis + pgvector
Infra:     Docker + Cloud Run ready
```

## Quick Start

### Prerequisites

- Go 1.24+
- Node.js 20+
- PostgreSQL 15+ with pgvector extension
- Redis 7+
- Docker (for terminal sandbox)

### Backend

```bash
cd backend
cp .env.example .env
# Edit .env with your database, Redis, and LLM API keys

go build -o bin/server ./cmd/server
./bin/server
# Server starts on :8001
```

### Frontend

```bash
cd frontend
npm install
npm run dev
# App starts on :5173
```

### Docker (Full Stack)

```bash
docker-compose up
```

## Architecture

```
HTTP Request
  → Middleware (Auth, CSRF, Rate Limit, Security Headers)
    → Handler (Route dispatch)
      → Service (Business logic)
        → Repository (PostgreSQL via sqlc)

Chat Message
  → LLM Provider (Claude/GPT/Ollama/Groq/xAI)
    → Context (workspace + RAG + conversation history)
      → SSE Stream → Frontend
```

## Project Structure

```
BusinessOS/
├── backend/                  # Go backend
│   ├── cmd/server/           # API server entry point
│   ├── internal/
│   │   ├── handlers/         # HTTP route handlers
│   │   ├── services/         # Business logic layer
│   │   ├── integrations/     # External API connectors (9 providers)
│   │   ├── middleware/       # Auth, CSRF, rate limiting
│   │   ├── database/         # PostgreSQL + migrations
│   │   ├── streaming/        # SSE real-time events
│   │   ├── container/        # Docker sandbox
│   │   └── terminal/         # Terminal emulator
│   ├── docker/               # Container images
│   └── Dockerfile            # Production build
│
├── frontend/                 # SvelteKit frontend
│   ├── src/
│   │   ├── routes/(app)/     # Protected app routes
│   │   ├── lib/components/   # UI components (200+)
│   │   ├── lib/stores/       # Svelte stores
│   │   ├── lib/api/          # API client modules
│   │   └── lib/modules/      # Feature modules
│   └── package.json
│
├── docs/                     # Documentation
└── docker-compose.yml        # Local dev environment
```

## Contributing

Contributions welcome. See `docs/` for architecture details.

## License

Apache 2.0

---

Built by [MIOSA](https://miosa.ai)
