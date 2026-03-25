# BusinessOS — Diátaxis Documentation

> **Fortune 500-grade AI business operating system.**
>
> Diátaxis documentation for BusinessOS — tutorials, how-to guides, explanations, and reference.

---

## About BusinessOS

BusinessOS is a self-hosted AI business OS built on the MIOSA four-layer stack. It gives your AI agents a home — with projects, clients, documents, team structure, and the full context of your operation.

**Tech Stack**: Go 1.24 backend + SvelteKit 2 frontend + Electron desktop

**Integration**: Uses Signal Theory, YAWL patterns, and Data Operating Standard from the root theory.

---

## Diátaxis Documentation

### [Tutorials](../../docs/diataxis/tutorials/) — Learn by Doing

| Tutorial | What You'll Learn | Time |
|----------|-------------------|------|
| [BusinessOS Desktop Deployment](../../docs/diataxis/tutorials/businessos-deploy.md) | Deploy self-hosted BusinessOS with Docker | 40 min |
| [Your First AI Operation](../../docs/diataxis/tutorials/first-operation.md) | Build Canopy workspace (foundation for BusinessOS ops) | 30 min |

### [How-to Guides](../../docs/diataxis/how-to/) — Solve Problems

| Guide | Solves | Complexity |
|-------|--------|------------|
| [Add S/N Quality Gates](../../docs/diataxis/how-to/add-quality-gates.md) | Reject low-quality agent output | Intermediate |
| [Create Client Communications](../../docs/diataxis/how-to/client-communications.md) | Signal-encoded client emails, proposals | Intermediate |
| [Deploy Multi-OS Fleet](../../docs/diataxis/how-to/multi-os-fleet.md) | Multiple BusinessOS instances | Expert |
| [Integrate Data Operating Standard](../../docs/diataxis/how-to/data-operating-standard.md) | SDK-backed data operations | Advanced |

### [Explanation](../../docs/diataxis/explanation/) — Understand the System

| Explanation | Topic | Why It Matters |
|-------------|-------|----------------|
| [The Chatman Equation](../../docs/diataxis/explanation/chatman-equation.md) | A=μ(O) foundation | BusinessOS applies this to business operations |
| [Signal Theory Complete](../../docs/diataxis/explanation/signal-theory-complete.md) | S=(M,G,T,F,W) encoding | Client comms use this |
| [The 7-Layer Architecture](../../docs/diataxis/explanation/seven-layer-architecture.md) | Optimal Systems design | BusinessOS implements all 7 layers |

### [Reference](../../docs/diataxis/reference/) — Look Up Details

| Reference | Covers | Format |
|-----------|--------|--------|
| [Signal Format](../../docs/diataxis/reference/signal-format.md) | S=(M,G,T,F,W) spec | BNF grammar |
| [API Endpoints](../../docs/diataxis/reference/api-endpoints.md) | All REST/WebSocket APIs | OpenAPI specs |
| [Configuration Keys](../../docs/diataxis/reference/configuration.md) | All config options | YAML reference |
| [Genre Catalog](../../docs/diataxis/reference/genre-catalog.md) | All Signal Theory genres | Usage guide |

---

## BusinessOS-Specific Documentation

### Core Modules

| Module | Diátaxis Docs | BusinessOS Docs |
|--------|---------------|-----------------|
| **Dashboard** | [Signal Theory](../../docs/diataxis/explanation/signal-theory-complete.md) | [Dashboard Guide](../modules/dashboard.md) |
| **Projects** | [YAWL Patterns](../../docs/diataxis/reference/yawl-43-patterns.md) | [Projects Module](../modules/projects.md) |
| **Tasks** | [How-to: Agent Handoffs](../../docs/diataxis/how-to/agent-handoffs.md) | [Tasks Module](../modules/tasks.md) |
| **AI Chat** | [Tutorial: Signal Theory](../../docs/diataxis/tutorials/signal-theory-practice.md) | [AI Chat Module](../modules/ai-chat.md) |
| **Clients** | [How-to: Client Communications](../../docs/diataxis/how-to/client-communications.md) | [CRM Module](../modules/clients.md) |
| **Documents** | [Explanation: Ontology Closure](../../docs/diataxis/explanation/ontology-closure.md) | [Documents Module](../modules/documents.md) |
| **OSA Integration** | [OSA Diátaxis Docs](../../OSA/docs/diataxis/README.md) | [OSA Module](../osa/) |

### Development

| Topic | Diátaxis Docs | Local Docs |
|-------|---------------|------------|
| **Backend** | [The Chatman Equation](../../docs/diataxis/explanation/chatman-equation.md) | [Go Backend Guide](../development/BACKEND.md) |
| **Frontend** | [Signal Theory](../../docs/diataxis/explanation/signal-theory-complete.md) | [SvelteKit Guide](../development/FRONTEND.md) |
| **Desktop** | [Tutorial: BusinessOS Deploy](../../docs/diataxis/tutorials/businessos-deploy.md) | [Electron Guide](../desktop/README.md) |
| **Testing** | [How-to: Add Quality Gates](../../docs/diataxis/how-to/add-quality-gates.md) | [Testing Guide](../development/TESTING.md) |

---

## AGI-Level Connections

### Signal Theory in BusinessOS

BusinessOS uses Signal Theory S=(M,G,T,F,W) for:

| Use Case | Signal Encoding |
|----------|-----------------|
| **Client Emails** | `S=(linguistic, email, direct, markdown, cold-email-anatomy)` |
| **Proposals** | `S=(linguistic, proposal, direct, markdown, persuasion)` |
| **Reports** | `S=(mixed, report, inform, markdown, dashboard-metrics)` |
| **Code Gen** | `S=(code, implementation, direct, typescript, module-pattern)` |

### YAWL Patterns in BusinessOS

BusinessOS workflows use YAWL patterns:

| Pattern | BusinessOS Usage |
|---------|------------------|
| **Parallel Split** | Multi-module updates |
| **Synchronization** | Multi-approval workflows |
| **Multi-Choice** | Dynamic agent activation |
| **Cancel Region** | Deployment rollback |

### Data Operating Standard

BusinessOS enforces SDK-backed data operations:

| Operation | SDK Method |
|-----------|------------|
| **Schema work** | `data-modelling-cli import/export` |
| **Decision records** | `bos decisions new` |
| **Knowledge articles** | `bos knowledge index` |
| **Data contracts** | `bos workspace init` |
| **Audit trail** | PROV-O via CONSTRUCT into Oxigraph |

---

## Quick Start

### For New Users

1. **Deploy BusinessOS**: [Tutorial](../../docs/diataxis/tutorials/businessos-deploy.md)
2. **Learn Signal Theory**: [Tutorial](../../docs/diataxis/tutorials/signal-theory-practice.md)
3. **Understand the System**: [7-Layer Architecture](../../docs/diataxis/explanation/seven-layer-architecture.md)

### For Developers

1. **Backend**: [Go Backend Guide](../development/BACKEND.md)
2. **Frontend**: [SvelteKit Guide](../development/FRONTEND.md)
3. **Data Operations**: [Data Operating Standard Guide](../../docs/diataxis/how-to/data-operating-standard.md)

### For Operators

1. **Deploy**: [Tutorial](../../docs/diataxis/tutorials/businessos-deploy.md)
2. **Configure**: [Configuration Reference](../../docs/diataxis/reference/configuration.md)
3. **Scale**: [Multi-OS Fleet](../../docs/diataxis/how-to/multi-os-fleet.md)

---

## Cross-Project Links

- **Root Diátaxis**: [Main Documentation](../../docs/diataxis/README.md)
- **Canopy**: [Canopy Diátaxis](../../canopy/docs/diataxis/README.md)
- **OSA**: [OSA Diátaxis](../../OSA/docs/diataxis/README.md)

---

*BusinessOS Diátaxis Documentation — Part of the ChatmanGPT Knowledge System*
