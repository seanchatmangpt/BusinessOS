# Complete Dependency Tree for 5 Features
## BusinessOS Q1 Implementation

**Generated:** 2026-01-08
**Project:** BusinessOS (Go Backend + SvelteKit Frontend)
**Status:** Detailed Analysis & Planning

---

## Executive Summary

This document provides a complete dependency tree for all 5 major features, showing:
1. **Which features depend on others**
2. **Shared components and services**
3. **Critical path analysis**
4. **Parallel execution opportunities**
5. **Integration points between features**

---

## Feature Overview

| # | Feature | Type | Est. Hours | Microtasks | Status |
|---|---------|------|-----------|-----------|--------|
| 1 | **Testing UI** | Full-Stack | 28-37 | 8 | Planning |
| 2 | **Preset Gallery** | Full-Stack | 50-65 | 32 | Planning |
| 3 | **Delegation System** | Frontend-Heavy | 36-48 | 18+ | Partial |
| 4 | **Categories** | Full-Stack | 24-32 | 8 | Planning |
| 5 | **Detail Page** | Full-Stack | 20-30 | 7 | Planning |
| | **TOTAL** | | **158-212 hours** | **73+** | |

---

## Master Dependency Graph

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        SHARED INFRASTRUCTURE                                │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐   │
│  │  Auth Layer  │  │  API Client  │  │  Svelte 5    │  │ PostgreSQL   │   │
│  │  (existing)  │  │  (existing)  │  │  Stores      │  │  (existing)  │   │
│  └──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘   │
└─────────────────────────────────────────────────────────────────────────────┘
                         ▲         ▲         ▲         ▲
                         │         │         │         │
        ┌────────────────┼─────────┼─────────┼─────────┤
        │                │         │         │         │
        ▼                ▼         ▼         ▼         ▼
   ┌─────────┐    ┌─────────┐   ┌──────────┐  ┌──────────┐  ┌──────────┐
   │ Feature │    │ Feature │   │ Feature  │  │ Feature  │  │ Feature  │
   │    1    │    │    2    │   │    3     │  │    4     │  │    5     │
   │ Testing │    │ Presets │   │Delegation│  │Categories│  │  Detail  │
   │   UI    │    │ Gallery │   │  System  │  │ Filtering│  │  Page    │
   └────┬────┘    └────┬────┘   └────┬─────┘  └────┬─────┘  └────┬─────┘
        │              │             │              │             │
        └──────────────┴─────────────┴──────────────┴─────────────┘
                 Dependencies flow downward →
```

---

## Feature 1: Testing UI

### Overview
- **Estimated Total:** 28-37 hours
- **Microtasks:** 8
- **Type:** Full-Stack (Backend + Frontend + Tests)
- **Critical Dependencies:** None (standalone feature)

### Detailed Structure

```
TESTING UI FEATURE
═══════════════════════════════════════════════════════════════

TIER 1: DATABASE FOUNDATION (2-3 hours)
┌─────────────────────────────────────────────────────┐
│ MT-1.1: Agent Testing Tables Migration (2-3h)       │
│  • agent_test_runs table                            │
│  • agent_test_metrics table                         │
│  • Indexes and foreign keys                         │
│  Dependencies: None                                 │
└────────────────────┬────────────────────────────────┘
                     ▼
TIER 2: BACKEND API LAYER (8-10 hours)
┌─────────────────────────────────────────────────────┐
│ MT-2.1: SQLC Queries (2-3h)                         │
│  • InsertAgentTestRun, UpdateAgentTestRun           │
│  • ListAgentTestRuns, GetAgentTestRun               │
│  Dependencies: MT-1.1                               │
└────────────────────┬────────────────────────────────┘
                     ▼
┌─────────────────────────────────────────────────────┐
│ MT-2.2: Agent Test Service (3-4h)                   │
│  • RunTest, RecordTestRun, GetTestHistory           │
│  • Metrics collection                               │
│  Dependencies: MT-2.1                               │
└────────────────────┬────────────────────────────────┘
                     ▼
┌─────────────────────────────────────────────────────┐
│ MT-2.3: HTTP Handler & Endpoints (2-3h)             │
│  • POST /api/agents/:id/test                        │
│  • GET /api/agents/:id/test-history                 │
│  Dependencies: MT-2.2                               │
└────────────────────┬────────────────────────────────┘
                     ▼
TIER 3: FRONTEND COMPONENTS (14-18 hours)
┌──────────────────────────────┬──────────────────────────────┐
│ MT-3.1: Main Container (2-3h)│ MT-4.1: API Client (2h)      │
│  • AgentTestPanel            │  • testAgent()               │
│  • Tab navigation            │  • getTestHistory()          │
│  • State management          │  Dependencies: MT-2.3        │
│  Dependencies: None          └──────────────────────────────┘
└────────────────┬─────────────┘
                 ▼
┌─────────────────────────────────────────────────────┐
│ MT-3.2: Input & Config Components (2-3h)            │
│  • AgentTestInput (textarea)                        │
│  • AgentTestConfig (sliders, toggles)               │
│  • Character counter, validation                    │
│  Dependencies: MT-3.1                               │
└────────────────┬────────────────────────────────────┘
                 ▼
┌─────────────────────────────────────────────────────┐
│ MT-3.3: Results & Execution (3-4h)                  │
│  • AgentTestExecution (loading state)               │
│  • AgentTestResults (display output)                │
│  • AgentTestMetrics (performance viz)               │
│  Dependencies: MT-3.2                               │
└────────────────┬────────────────────────────────────┘
                 ▼
┌──────────────────────────────┬──────────────────────────────┐
│ MT-4.2: Error Handler (2h)   │ MT-4.3: Loading States (1.5h)│
│  • AgentTestError component  │  • Spinner animation         │
│  • Error detection & display │  • Skeleton screens          │
│  Dependencies: MT-3.3        │  Dependencies: MT-4.2        │
└──────────────────────────────┴──────────────────────────────┘

TIER 4: TESTING (6-9 hours) [Can run in parallel]
├─ MT-5.1: Integration Tests (2-3h)
├─ MT-5.2: Unit Tests (2-3h)
└─ MT-5.3: E2E & Performance (2-3h)
   All dependencies: MT-4.3
```

### Shared Dependencies
- None (completely standalone)

### Integration Points
- Uses existing `custom_agents` table
- Uses existing `AgentV2` interface
- Uses existing LLM service
- Frontend uses existing API client patterns

### Critical Path
```
MT-1.1 → MT-2.1 → MT-2.2 → MT-2.3 → MT-4.1 → MT-3.1 → MT-3.2
→ MT-3.3 → MT-4.2 → MT-4.3 → MT-5.1 (parallel with 5.2)
```

**Critical Path Duration:** ~20-25 hours sequential

---

## Feature 2: Preset Gallery

### Overview
- **Estimated Total:** 50-65 hours
- **Microtasks:** 32 (largest feature)
- **Type:** Full-Stack (Backend + Frontend + Tests)
- **Critical Dependencies:** None (standalone feature)

### Detailed Structure

```
PRESET GALLERY FEATURE
═══════════════════════════════════════════════════════════════

PHASE 1: DATABASE (3 hours)
┌────────────────────────────────────────────────────┐
│ MT-1.1: Presets Migration (1.5h)                   │
│ MT-1.2: Analytics Migration (1h)                   │
│ MT-1.3: SQLC Code Generation (1h)                  │
│ Dependency Chain: 1.1 → 1.2 → 1.3                  │
└────────────────────┬───────────────────────────────┘
                     ▼ (blocks 8+ tasks)

PHASE 2A: BACKEND CORE API (9.5 hours sequential)
┌────────────────────────────────────────────────────┐
│ MT-2.1: List Presets (2h)              ◄── MT-1.3  │
│ MT-2.2: Repository Layer (2h)          ◄── MT-2.1  │
│ MT-2.3: Service Layer (2h)             ◄── MT-2.2  │
│ MT-2.4: Get Detail Endpoint (1.5h)     ◄── MT-2.3  │
│ MT-2.5: Search Endpoint (2h)           ◄── MT-2.4  │
└────────────────────┬───────────────────────────────┘
                     ▼ (blocks Phase 2B)

PHASE 2B: BACKEND CREATE (6 hours sequential)
┌────────────────────────────────────────────────────┐
│ MT-3.1: Use Preset Endpoint (2.5h)     ◄── MT-2.5  │
│ MT-3.2: Config Validator (1.5h)        ◄── MT-3.1  │
│ MT-3.3: Customize Endpoint (2h)        ◄── MT-3.2  │
└────────────────────┬───────────────────────────────┘
                     ▼

PHASE 3A: FRONTEND GALLERY UI (7.5 hours)
  (Can start after MT-1.3, parallel to Phase 2)
┌────────────────────────────────────────────────────┐
│ MT-4.1: Gallery Page (2h)              ◄── None    │
│ MT-4.2: Store Management (1.5h)        ◄── MT-4.1  │
│ MT-4.3: Preset Card Component (1.5h)   ◄── MT-4.2  │
│ MT-4.4: Grid Component (1.5h)          ◄── MT-4.3  │
│ MT-4.5: Category Sidebar (1.5h)        ◄── MT-4.4  │
└────────────────────┬───────────────────────────────┘
                     ▼

PHASE 3B: FRONTEND MODALS (8 hours)
┌────────────────────────────────────────────────────┐
│ MT-5.1: Detail Modal (2.5h)            ◄── MT-4.5  │
│ MT-5.2: Use Preset Flow (2h)           ◄── MT-3.1  │
│ MT-5.3: Customize Modal (2.5h)         ◄── MT-3.3  │
│ MT-5.4: Search Integration (1.5h)      ◄── MT-5.3  │
└────────────────────┬───────────────────────────────┘
                     ▼

PHASE 3C: FRONTEND SEARCH (7 hours)
┌────────────────────────────────────────────────────┐
│ MT-6.1: Advanced Search Form (2h)      ◄── MT-5.4  │
│ MT-6.2: Search Results (1.5h)          ◄── MT-6.1  │
│ MT-6.3: Favorites (1.5h)               ◄── MT-6.2  │
│ MT-6.4: Analytics Dashboard (2h)       ◄── MT-6.3  │
└────────────────────┬───────────────────────────────┘
                     ▼

PHASE 4: TESTING (9 hours, can run parallel to frontend work)
┌────────────────────────────────────────────────────┐
│ MT-7.1: API Integration Tests (2h)     ◄── MT-3.3  │
│ MT-7.2: Frontend Component Tests (2h)  ◄── MT-6.3  │
│ MT-7.3: E2E Tests (2.5h)               ◄── 7.1+7.2 │
│ MT-7.4: Performance & Security (2h)    ◄── MT-7.3  │
└────────────────────┬───────────────────────────────┘
                     ▼

PHASE 5: DOCUMENTATION & DEPLOYMENT (5 hours)
┌────────────────────────────────────────────────────┐
│ MT-8.1: API Documentation (1.5h)       ◄── MT-7.4  │
│ MT-8.2: User Guide (1.5h)              ◄── MT-8.1  │
│ MT-8.3: Deployment Guide (1h)          ◄── MT-8.2  │
│ MT-8.4: Production Deploy (1.5h)       ◄── MT-8.3  │
└────────────────────┬───────────────────────────────┘
                     ▼ COMPLETE
```

### Dependency Summary
```
Linear Chain (Critical Path):
MT-1.1 → MT-1.2 → MT-1.3 → MT-2.1 → MT-2.2 → MT-2.3 → MT-2.4
→ MT-2.5 → MT-3.1 → MT-3.2 → MT-3.3 → MT-7.1 → MT-7.3 → MT-7.4
→ MT-8.1 → MT-8.2 → MT-8.3 → MT-8.4

Parallel Tracks:
A: MT-4.1 → MT-4.2 → MT-4.3 → MT-4.4 → MT-4.5 (starts day 1)
B: MT-5.x (depends on MT-3.1, MT-3.3)
C: MT-6.x (depends on MT-5.4)
```

### Critical Path Duration
```
With 1 engineer:   ~50-65 hours
With 2 engineers:  ~25-30 hours
With 3 engineers:  ~18-20 hours
```

### Shared Dependencies with Other Features
- **Testing UI:** Can use same API client patterns
- **Detail Page:** Shares agent detail data structure
- **Categories:** Presets can be filtered by category (Feature 4)

---

## Feature 3: Delegation System

### Overview
- **Estimated Total:** 36-48 hours
- **Microtasks:** 18+ (distributed across tracks)
- **Type:** Frontend-Heavy + Backend Integration
- **Backend Status:** ~70% complete (handlers exist)
- **Frontend Status:** 0% (needs full UI)

### Detailed Structure

```
DELEGATION SYSTEM
═══════════════════════════════════════════════════════════════

Backend Status: PARTIAL (handlers exist)
├── ✅ API endpoints for listing agents
├── ✅ Mention resolution service
├── ✅ Delegation tracking database
└── ⚠️ Frontend integration incomplete

TRACK A: @MENTION AUTOCOMPLETE (8-12 hours)
┌────────────────────────────────────────────────────┐
│ MT-A1: Basic Dropdown (2-3h)                       │
│  • Detect "@" in textarea                          │
│  • Show agent list below cursor                    │
│  • Keyboard navigation (arrows, Enter)             │
│  Dependencies: None                                │
└────────────────┬───────────────────────────────────┘
                 ▼
┌────────────────────────────────────────────────────┐
│ MT-A2: Fuzzy Search (2-3h)                         │
│  • Filter by agent name/description                │
│  • Rank by relevance                               │
│  • Highlight matches                              │
│  Dependencies: MT-A1                               │
└────────────────┬───────────────────────────────────┘
                 ▼
┌────────────────────────────────────────────────────┐
│ MT-A3: Multi-Mention Support (2h)                  │
│  • Allow multiple @mentions per message            │
│  • Detect duplicates                               │
│  • Show preview badges                             │
│  Dependencies: MT-A1, MT-A2                        │
└────────────────┬───────────────────────────────────┘
                 ▼
┌────────────────────────────────────────────────────┐
│ MT-A4: Dynamic Agent List (2-3h)                   │
│  • Call /api/agents/available                      │
│  • Cache agent list                                │
│  • Handle API errors                               │
│  Dependencies: MT-A3, (Backend ready)              │
└────────────────┬───────────────────────────────────┘
                 ▼

TRACK B: DELEGATION PANEL & WORKFLOW (12-16 hours)
┌────────────────────────────────────────────────────┐
│ MT-B1: Delegation Panel UI (2-3h)                  │
│  • Side panel showing mentioned agents             │
│  • Capability preview cards                        │
│  • Accept/Configure buttons                        │
│  Dependencies: MT-A4                               │
└────────────────┬───────────────────────────────────┘
                 ▼
┌────────────────────────────────────────────────────┐
│ MT-B2: Agent Capability Config (2-3h)              │
│  • Configure agent parameters before delegation    │
│  • Temperature, tokens, model selection            │
│  • Custom system prompt (optional)                 │
│  Dependencies: MT-B1                               │
└────────────────┬───────────────────────────────────┘
                 ▼
┌────────────────────────────────────────────────────┐
│ MT-B3: Confirmation Modal (2-3h)                   │
│  • Confirm delegation action                       │
│  • Show agent + config summary                     │
│  • Confirm/Cancel buttons                          │
│  Dependencies: MT-B2                               │
└────────────────┬───────────────────────────────────┘
                 ▼
┌────────────────────────────────────────────────────┐
│ MT-B4: Delegation Execution (3-4h)                 │
│  • Send delegation request to API                  │
│  • Stream agent response                           │
│  • Show agent thinking/response inline             │
│  Dependencies: MT-B3, (Backend endpoints ready)    │
└────────────────┬───────────────────────────────────┘
                 ▼

TRACK C: DELEGATION HISTORY & MANAGEMENT (8-12 hours)
┌────────────────────────────────────────────────────┐
│ MT-C1: History Panel (2-3h)                        │
│  • List recent delegations                         │
│  • Filter by agent, date, status                   │
│  • Search by message                               │
│  Dependencies: MT-B4                               │
└────────────────┬───────────────────────────────────┘
                 ▼
┌────────────────────────────────────────────────────┐
│ MT-C2: Delegation Details (2h)                     │
│  • Modal showing full delegation info              │
│  • Agent response, duration, tokens                │
│  • Retry/Edit options                              │
│  Dependencies: MT-C1                               │
└────────────────┬───────────────────────────────────┘
                 ▼
┌────────────────────────────────────────────────────┐
│ MT-C3: Conflict Detection (2-3h)                   │
│  • Warn if same agent mentioned twice              │
│  • Warn if agent unavailable                       │
│  • Rate limiting feedback                          │
│  Dependencies: MT-C2                               │
└────────────────┬───────────────────────────────────┘
                 ▼
┌────────────────────────────────────────────────────┐
│ MT-C4: Analytics & Stats (2-3h)                    │
│  • Delegation frequency per agent                  │
│  • Success rate, avg response time                 │
│  • Cost tracking                                   │
│  Dependencies: MT-C3                               │
└────────────────┬───────────────────────────────────┘
                 ▼

TRACK D: TESTING & INTEGRATION (8-12 hours)
└── MT-D1: Integration Tests (2-3h) ◄── MT-B4
└── MT-D2: Component Tests (2-3h) ◄── MT-C4
└── MT-D3: E2E Delegation Flow (2-3h) ◄── D1+D2
└── MT-D4: Error Handling & Edge Cases (2-3h) ◄── D3
```

### Backend Dependencies
```
❌ BLOCKING: Backend needs existing handlers to be:
   • Connected to frontend API client
   • Verified for /api/agents/available endpoint
   • Verified for delegation execution endpoint
   • Error handling tested
```

### Frontend Dependencies
```
✅ Requires:
   • Feature 1 (Testing UI) patterns (optional, for config UI)
   • AgentV2 type definitions from backend
   • API client integration
```

### Shared Components Opportunity
```
✓ Can reuse from Testing UI (Feature 1):
  • Temperature slider component
  • Max tokens slider component
  • Model selector dropdown
  • Loading spinner and states
```

---

## Feature 4: Categories / Filtering

### Overview
- **Estimated Total:** 24-32 hours
- **Microtasks:** 8
- **Type:** Full-Stack (Frontend-Heavy)
- **Backend Requirements:** Minimal (query param support)

### Detailed Structure

```
CATEGORIES / FILTERING FEATURE
═══════════════════════════════════════════════════════════════

STATE MANAGEMENT & STORES (2-3 hours)
┌────────────────────────────────────────────────────┐
│ MT-1: Filter State Store (2-3h)                    │
│  • Svelte store for filter state                   │
│  • selectedCategories: string[]                    │
│  • Toggle, clear, set functions                    │
│  Dependencies: None                                │
└────────────────┬───────────────────────────────────┘
                 ▼

FRONTEND COMPONENTS (8-10 hours parallel)
┌──────────────────────────────┬──────────────────────────────┐
│ MT-2: Filter Component (2-3h)│ MT-5: Badge Components (2-3h)│
│  • Dropdown with checkboxes  │  • FilterBadge               │
│  • Category list             │  • FilterBadgeGroup          │
│  • Select All/Clear          │  • Remove button             │
│  Dependencies: MT-1          │  Dependencies: MT-1          │
└──────────────┬───────────────┴──────────────┬───────────────┘
               ▼ (parallel)                   ▼

┌────────────────────────────────────────────────────┐
│ MT-3: API Integration (3-4h)                       │
│  • Build filter query params                       │
│  • Debounce API calls (300ms)                      │
│  • Handle filtered results                         │
│  Dependencies: MT-2                                │
│  Backend Requirement: Support ?categories=cat1,cat2│
└────────────────┬───────────────────────────────────┘
                 ▼
┌────────────────────────────────────────────────────┐
│ MT-4: URL Query Params (2-3h)                      │
│  • Read from $page.url.searchParams                │
│  • Sync to URL on filter change                    │
│  • History API support (back/forward)              │
│  Dependencies: MT-3                                │
└────────────────┬───────────────────────────────────┘
                 ▼

POLISH & UX (3-4 hours)
┌────────────────────────────────────────────────────┐
│ MT-6: Clear Filters Feature (1-2h)                 │
│  • Clear All button with confirmation              │
│  • Reset pagination to page 1                      │
│  Dependencies: MT-2, MT-5                          │
└────────────────┬───────────────────────────────────┘
                 ▼
┌────────────────────────────────────────────────────┐
│ MT-7: Animation & UX Polish (2-3h)                 │
│  • Badge entrance/exit animations                  │
│  • Filter panel open/close                         │
│  • Result count animation                          │
│  Dependencies: MT-2, MT-5, MT-6                    │
└────────────────┬───────────────────────────────────┘
                 ▼

TESTING (3-4 hours)
┌────────────────────────────────────────────────────┐
│ MT-8: Integration & Unit Tests (3-4h)              │
│  • Filter store operations                         │
│  • Component rendering                             │
│  • E2E filtering flow                              │
│  • URL persistence                                 │
│  Dependencies: MT-7                                │
└────────────────┬───────────────────────────────────┘
                 ▼ COMPLETE
```

### Dependencies on Other Features
```
DEPENDS ON:
├── Feature 2 (Preset Gallery): Categories could filter presets
├── Feature 5 (Detail Page): Could filter agents by category

USED BY:
├── Feature 2 (Preset Gallery): Needs category sidebar filter
└── Feature 5 (Detail Page): Could categorize agent details
```

### Backend Requirements
```
Minimal changes needed:
├── Modify GET /api/artifacts?categories=cat1,cat2
├── Modify GET /api/search?categories=cat1&query=...
└── Add GET /api/categories (optional, for category list)

If already implemented: 0 hours
If needs implementation: ~2-4 hours (can be done in parallel)
```

### Critical Path
```
MT-1 → MT-2 → MT-3 → MT-4 → MT-6 → MT-7 → MT-8
Duration: ~18-22 hours sequential (can parallelize MT-2+5)
```

---

## Feature 5: Agent Detail Page

### Overview
- **Estimated Total:** 20-30 hours
- **Microtasks:** 7
- **Type:** Full-Stack (Backend + Frontend)
- **Complexity:** Moderate

### Detailed Structure

```
AGENT DETAIL PAGE FEATURE
═══════════════════════════════════════════════════════════════

FOUNDATION (1-2 hours)
┌────────────────────────────────────────────────────┐
│ MT-1: Route & Data Loading (1-2h)                  │
│  Backend:                                          │
│  • GetAgentDetailsWithMetrics endpoint             │
│  • Fetch agent + usage metrics + recent tests      │
│  Frontend:                                         │
│  • Create +page.svelte, +page.server.ts            │
│  • Load agent data via API                         │
│  Dependencies: None                                │
└────────────────┬───────────────────────────────────┘
                 ▼

TAB PANELS (12-16 hours parallel)
┌──────────────────────────────┬──────────────────────────────┐
│ MT-2: Overview Tab (2-3h)    │ MT-3: Usage Stats (2-3h)     │
│  • Agent config display      │  • Test metrics              │
│  • Capabilities list         │  • Response time avg         │
│  • Tools enabled             │  • Token usage history       │
│  • Category & metadata       │  • Success rate              │
│  Dependencies: MT-1          │  Dependencies: MT-1          │
└──────────────┬───────────────┴──────────────┬───────────────┘
               ▼ (can run parallel)           ▼

┌────────────────────────────────────────────────────┐
│ MT-4: Settings Tab (2-3h)                          │
│  • Edit agent configuration                        │
│  • System prompt editor                            │
│  • Temperature/tokens sliders                      │
│  • Tool toggles                                    │
│  Dependencies: MT-1                                │
└────────────────┬───────────────────────────────────┘
                 ▼

┌────────────────────────────────────────────────────┐
│ MT-5: Tab Navigation (1-2h)                        │
│  • Tab switcher component                          │
│  • Lazy load tab content                           │
│  • Persist active tab in state                     │
│  Dependencies: MT-2, MT-3, MT-4                    │
└────────────────┬───────────────────────────────────┘
                 ▼

┌────────────────────────────────────────────────────┐
│ MT-6: Testing Tab (2-3h)                           │
│  • Embed AgentTestPanel (from Feature 1)           │
│  • Run tests within detail page                    │
│  • Show test history                               │
│  Dependencies: MT-5, (Feature 1: Testing UI)       │
└────────────────┬───────────────────────────────────┘
                 ▼

┌────────────────────────────────────────────────────┐
│ MT-7: Header & Navigation (1-2h)                   │
│  • Agent avatar + name + description               │
│  • Edit/Delete/Share buttons                       │
│  • Back navigation                                 │
│  • Loading/error states                            │
│  Dependencies: All tabs complete (MT-6)            │
└────────────────┬───────────────────────────────────┘
                 ▼

TESTING (3-4 hours)
└── MT-8: Integration tests (3-4h) ◄── MT-7
```

### Backend Dependencies
```
✅ EXISTING:
  • custom_agents table (has all needed fields)
  • Agent test runs data

⚠️ NEEDED:
  • GetAgentDetailsWithMetrics endpoint (~1-2h)
  • Join with test metrics if not already available
  • Performance optimization (indexes)
```

### Frontend Dependencies
```
✅ Can reuse from Feature 1 (Testing UI):
  • AgentTestPanel component
  • Test result display components
  • Loading spinners and states

NEW:
  • Overview tab component
  • Usage statistics visualization
  • Settings tab with edit functionality
```

### Shared Components Opportunity
```
REUSE FROM FEATURE 1 (Testing UI):
├── AgentTestPanel (test results)
├── AgentTestResults component
├── AgentTestMetrics component
├── Loading spinner
└── Error handling patterns

REUSE FROM FEATURE 4 (Categories):
├── Category filtering (optional, for agent categorization)
└── Badge components
```

### Critical Path
```
MT-1 → MT-2, MT-3, MT-4 (parallel) → MT-5 → MT-6 → MT-7 → MT-8
Duration: ~18-22 hours sequential
```

---

## Cross-Feature Dependencies Summary

### Dependency Matrix

```
┌────────┬──────────┬────────────┬────────────┬──────────┬──────────┐
│        │ Test UI  │ Presets    │Delegation  │Categories│  Detail  │
├────────┼──────────┼────────────┼────────────┼──────────┼──────────┤
│Test UI │    -     │ Could reuse│ Can reuse  │    -     │ Provides │
│        │          │ config UI  │ config UI  │          │ component│
├────────┼──────────┼────────────┼────────────┼──────────┼──────────┤
│Presets │    -     │     -      │ Presets by │ Presets  │ Link to  │
│        │          │            │ agent      │ filtered │ agent    │
├────────┼──────────┼────────────┼────────────┼──────────┼──────────┤
│Delegat.│    -     │     -      │     -      │    -     │    -     │
│        │          │            │            │          │          │
├────────┼──────────┼────────────┼────────────┼──────────┼──────────┤
│Categor.│    -     │ Needed for │ Agents by  │    -     │ Category │
│        │          │ filter     │ category   │          │ display  │
├────────┼──────────┼────────────┼────────────┼──────────┼──────────┤
│ Detail │ Embedded │ Link from  │ Link from  │ Category │    -     │
│        │ in tab   │ preset list│ delegation │ display  │          │
└────────┴──────────┴────────────┴────────────┴──────────┴──────────┘
```

### Shared Components & Services

#### Component Reuse

```
TESTING UI (Feature 1) provides to other features:
├── Temperature slider (used by Delegation, Detail Page)
├── Max tokens slider (used by Delegation, Detail Page)
├── Model selector dropdown (used by Delegation)
├── Loading spinner (used by all)
├── Error handling patterns (used by all)
└── Metrics visualization (used by Detail Page, Presets)

CATEGORIES (Feature 4) provides:
├── Filter badge component (used by Preset Gallery, Detail Page)
├── Category filter UI (used by Preset Gallery)
└── Filter state management patterns (reusable)

DETAIL PAGE (Feature 5) provides:
└── None (terminal feature)
```

#### Service/API Reuse

```
Agent Details Service (Feature 5):
├── GetAgentDetails() - used by Feature 1 (Test UI)
├── GetAgentMetrics() - used by Preset Gallery detail modal
└── GetAgentUsageStats() - used by Detail Page

Test History Service (Feature 1):
└── GetTestHistory() - displayed in Detail Page testing tab

Preset Service (Feature 2):
├── UsePreset() - called from Detail Page or Preset Gallery
├── CustomizePreset() - used by Delegation system config
└── ListPresets() - displayed in agent context
```

---

## Implementation Order & Critical Path

### Option A: Minimal Dependencies (Parallel All)

```
Day 1-2 (Parallel start):
├─ Engineer A: Feature 1 (Testing UI) - 28-37h
├─ Engineer B: Feature 2 (Preset Gallery) - 50-65h
├─ Engineer C: Feature 3 (Delegation) - 36-48h
├─ Engineer D: Feature 4 (Categories) - 24-32h
└─ Engineer E: Feature 5 (Detail Page) - 20-30h

Team size: 5 engineers
Total duration: ~5-6 weeks
```

### Option B: Dependency-Aware (Recommended)

```
PHASE 1 (Weeks 1-2): Infrastructure Features
├─ Feature 1: Testing UI (28-37h)
├─ Feature 4: Categories (24-32h)
└─ Feature 5: Detail Page (20-30h) ← Uses components from Feature 1
   Total: 72-99 hours (2 weeks with 3 engineers)

PHASE 2 (Weeks 3-4): Gallery Features
├─ Feature 2: Preset Gallery (50-65h) ← Uses categories from Feature 4
└─ Feature 3: Delegation (36-48h) ← Uses components from Feature 1
   Total: 86-113 hours (2 weeks with 2-3 engineers)

Overall: 4 weeks with 3 engineers
```

### Option C: Sequential (Minimum team)

```
Week 1: Feature 1 (Testing UI) - 28-37h
Week 2: Feature 4 (Categories) - 24-32h
Week 3-4: Feature 5 (Detail Page) - 20-30h (uses Feature 1 components)
Week 5-7: Feature 2 (Preset Gallery) - 50-65h (uses Features 1, 4, 5)
Week 8-9: Feature 3 (Delegation) - 36-48h (uses Features 1, 5)

Total: 9 weeks with 1 engineer
```

---

## Shared Infrastructure & Dependencies

### Database Tables Used

```
EXISTING:
├── custom_agents (all features read/write)
├── agent_test_runs (Feature 1 creates, Features 2,5 read)
├── auth.users (all features for authorization)
└── workspaces (optional, for scoping)

NEW TO CREATE:
├── agent_testing tables (Feature 1)
├── presets tables (Feature 2)
├── delegation records (Feature 3)
├── category mappings (Feature 4)
└── usage_metrics (Feature 5)
```

### API Endpoints

```
FEATURE 1 (Testing UI):
├── POST /api/agents/:id/test
└── GET /api/agents/:id/test-history

FEATURE 2 (Preset Gallery):
├── GET /api/presets
├── GET /api/presets/:id
├── POST /api/presets/:id/use
├── POST /api/presets/:id/customize
└── GET /api/presets/search?q=...&categories=...

FEATURE 3 (Delegation):
├── GET /api/agents/available
├── POST /api/agents/:id/delegate
└── GET /api/delegation-history

FEATURE 4 (Categories):
├── GET /api/categories
├── GET /api/artifacts?categories=...
└── GET /api/search?categories=...&q=...

FEATURE 5 (Detail Page):
└── GET /api/agents/:id/details
```

### Svelte Stores & Context

```
SHARED:
├── filterStore (Feature 4) - used by Features 2, 5
├── agentStore (existing) - used by all features
└── authStore (existing) - used by all features

FEATURE-SPECIFIC:
├── testingStore (Feature 1)
├── presetGalleryStore (Feature 2)
├── delegationStore (Feature 3)
└── detailPageStore (Feature 5)
```

---

## Risk Analysis & Mitigation

### Critical Blockers

```
1. FEATURE 2 BLOCKER: Database schema (MT-1.1)
   Risk: If schema wrong, 40+ hours blocked
   Mitigation: Review schema day 1, test locally
   Impact: High - blocks all Phase 2

2. FEATURE 3 BLOCKER: Backend endpoints not ready
   Risk: Frontend teams waiting on backend
   Mitigation: Finalize API contract early, mock API
   Impact: High - blocks B4 through D4

3. FEATURE 4 BLOCKER: Backend category support
   Risk: If backend doesn't support ?categories param
   Mitigation: Implement backend changes in parallel
   Impact: Medium - can use mock data temporarily

4. FEATURE 5 BLOCKER: Feature 1 components
   Risk: Detail page blocked waiting for Test UI components
   Mitigation: Extract reusable components early
   Impact: Medium - causes ~1-2 week delay
```

### Dependency Hazards

```
HAZARD 1: Component API Changes
Risk: If Testing UI components change, Delegation + Detail break
Mitigation: Finalize component props by end of Feature 1
Timeline: End of Day 1

HAZARD 2: Category System Mismatch
Risk: Features 2,4,5 disagree on category storage/naming
Mitigation: Define category schema before Feature 2 starts
Timeline: Before Week 3

HAZARD 3: Test Data Conflicts
Risk: Multiple features creating test agents/presets interfere
Mitigation: Use separate test workspaces, clear data between runs
Timeline: End of each feature phase

HAZARD 4: Performance Issues Late
Risk: Only discovered during Feature 2 testing
Mitigation: Load test database early (MT-1.1)
Timeline: Week 1
```

---

## Team Allocation Recommendations

### Option A: 3-Person Team (14 weeks)

```
Engineer 1 (Backend):
├─ Week 1-2: Feature 1 DB + API
├─ Week 3: Feature 4 backend (if needed)
├─ Week 4-7: Feature 2 backend (largest)
└─ Week 8-9: Feature 3 backend

Engineer 2 (Frontend):
├─ Week 1-2: Feature 1 components (parallel)
├─ Week 3-4: Feature 4 components
├─ Week 5-7: Feature 2 components
└─ Week 8-9: Feature 3 components

Engineer 3 (QA/DevOps):
├─ Week 1-9: Continuous testing, performance monitoring
└─ Week 10+: Documentation, deployment
```

### Option B: 5-Person Team (4-5 weeks - Aggressive)

```
Team 1 (Backend): Feature 1 + 4 + 5 (parallel)
Team 2 (Frontend): Feature 1 + 4 + 5 (parallel)
Team 3 (Backend): Feature 2 (starts week 2)
Team 4 (Frontend): Feature 2 (starts week 2)
Team 5 (QA/Devops): Continuous, then Feature 3 support

NOTE: Feature 3 typically starts week 3-4 after Feature 1 components ready
```

---

## Success Criteria & Verification

### Per-Feature Completion

```
FEATURE 1 (Testing UI): ✅ Complete
├─ [ ] All 8 microtasks done
├─ [ ] 80%+ test coverage
├─ [ ] No console errors
├─ [ ] Performance < 100ms component render
└─ [ ] Components reusable by Features 3,5

FEATURE 2 (Preset Gallery): ✅ Complete
├─ [ ] All 32 microtasks done
├─ [ ] 90%+ API coverage, 80%+ frontend
├─ [ ] Pagination handles 1000+ presets
├─ [ ] Search < 200ms response
├─ [ ] Categories filter working

FEATURE 3 (Delegation): ✅ Complete
├─ [ ] All tracks complete (A-D)
├─ [ ] @mention autocomplete working
├─ [ ] Delegation execution successful
├─ [ ] Error handling for rate limits
└─ [ ] E2E flow tested

FEATURE 4 (Categories): ✅ Complete
├─ [ ] All 8 microtasks done
├─ [ ] Filter persistence in URL
├─ [ ] 80%+ test coverage
├─ [ ] Performance: filtering < 300ms
└─ [ ] Animations smooth (60fps)

FEATURE 5 (Detail Page): ✅ Complete
├─ [ ] All 8 microtasks done
├─ [ ] All tabs rendering
├─ [ ] Testing tab shows Feature 1 data
├─ [ ] Categories displaying correctly
└─ [ ] Performance: page load < 2s
```

### Cross-Feature Integration

```
WHEN ALL 5 FEATURES COMPLETE:
├─ [ ] User can create agent from preset (2→5)
├─ [ ] User can delegate to agent (3→5)
├─ [ ] User can filter presets by category (4→2)
├─ [ ] User can test agent from detail page (1→5)
├─ [ ] All workflows end-to-end tested (1,2,3,4,5)
└─ [ ] Total test coverage > 85%
```

---

## Appendix: File Checklist

### Feature 1: Testing UI Files
```
NEW:
├─ desktop/backend-go/internal/database/migrations/037_agent_testing.sql
├─ desktop/backend-go/internal/handlers/agent_testing.go
├─ desktop/backend-go/internal/services/agent_testing_service.go
├─ frontend/src/lib/api/agent-testing.ts
├─ frontend/src/lib/components/settings/AgentTestPanel.svelte
├─ frontend/src/lib/components/settings/AgentTestInput.svelte
├─ frontend/src/lib/components/settings/AgentTestConfig.svelte
├─ frontend/src/lib/components/settings/AgentTestExecution.svelte
├─ frontend/src/lib/components/settings/AgentTestResults.svelte
├─ frontend/src/lib/components/settings/AgentTestMetrics.svelte
├─ frontend/src/lib/components/settings/AgentTestError.svelte
├─ frontend/src/lib/components/settings/AgentTestSpinner.svelte
└─ docs/AGENT_TESTING_UI.md

MODIFY:
├─ desktop/backend-go/cmd/server/main.go (routes)
└─ frontend/src/lib/components/settings/+page.svelte
```

### Feature 2: Preset Gallery Files
```
NEW (32 total): Database migrations, handlers, services, 18+ Svelte components, tests, docs

MODIFY: Server routes, API exports
```

### Feature 3: Delegation System Files
```
NEW (15+): @mention component, delegation UI, modals, tests

MODIFY: Chat input component, server routes
```

### Feature 4: Categories Files
```
NEW: Filter store, filter components, badge components, tests, API client

MODIFY: Artifact list component, API calls, database queries
```

### Feature 5: Detail Page Files
```
NEW: Detail page route, overview/stats/settings tabs, tests

MODIFY: Agent list navigation, server routes
```

---

## Final Notes

1. **Feature 1 (Testing UI)** is a good starting point - no dependencies, enables other features
2. **Feature 4 (Categories)** should be done early - needed by Features 2,5
3. **Feature 2 (Preset Gallery)** is largest - start after Features 1,4 infrastructure ready
4. **Feature 3 (Delegation)** depends on Feature 1 components - can start week 2-3
5. **Feature 5 (Detail Page)** benefits from Features 1,4 but can start independently

**Recommended execution:** Features 1 + 4 in parallel (week 1-2), then 2,3,5 in subsequent phases

---

**Last Updated:** 2026-01-08
**Total Estimated Effort:** 158-212 hours (~4-6 weeks with 3 engineers)
**Status:** Ready for Sprint Planning & Parallel Execution
