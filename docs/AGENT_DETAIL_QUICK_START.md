# Agent Detail Page - Quick Start Guide

**For:** Development Team
**Purpose:** Fast reference to begin implementation
**Time to Implementation:** Start within 30 minutes

---

## 30-Second Overview

You're building a detail page for custom agents with:
- Overview tab (agent config + metrics)
- Usage stats tab (performance history)
- Settings tab (edit agent)
- Testing tab (try agent with message)

**7 atomic tasks**, each 1-4 hours, with explicit dependencies.

---

## Your 5-Minute Checklist

Before you start coding:

- [ ] Read `/AGENT_DETAIL_PAGE_MICROTASKS.md` (15 min)
- [ ] Skim `/AGENT_DETAIL_PAGE_VISUAL_GUIDE.md` (5 min)
- [ ] Check existing detail page: `/frontend/src/routes/(app)/clients/[id]/+page.svelte`
- [ ] Check agent API: `/desktop/backend-go/internal/handlers/agents.go`
- [ ] Create branch: `git checkout -b feature/agent-detail-page`

---

## Task Assignments

### Backend Developer
**Start with:** Microtask 1 - Backend Tasks

```
1. Open: /desktop/backend-go/internal/handlers/agents.go
2. Create: GetAgentDetailsWithMetrics() function
3. Create: GetAgentUsageStats() function
4. Create: GetAgentTestHistory() function
5. Register routes in main handlers file
6. Test with curl/Postman
```

**Time:** 2-3 hours
**Dependencies:** None

### Frontend Developer
**Start with:** Microtask 1 - Frontend Tasks

```
1. Create: /frontend/src/routes/(app)/agents/[id]/+page.svelte
2. Create: /frontend/src/routes/(app)/agents/[id]/+page.server.ts
3. Create: AgentDetailResponse type in /lib/api/
4. Test loading page with mock data
5. Build each tab (MT-2, 3, 4 can be parallel)
```

**Time:** 2-3 hours
**Dependencies:** Backend routes ready

---

## File Creation Checklist

### Backend Files to Create/Modify

**Modify:**
```
✓ desktop/backend-go/internal/handlers/agents.go
  - Add GetAgentDetailsWithMetrics()
  - Add GetAgentUsageStats()
  - Add GetAgentTestHistory()

✓ desktop/backend-go/cmd/server/main.go
  - Register new routes

? desktop/backend-go/internal/database/sqlc/
  - Check if usage tracking queries exist
  - Create if needed
```

**Create:** (If don't exist)
```
? desktop/backend-go/internal/services/agent_metrics.go
  - Helper functions for metrics calculation
```

### Frontend Files to Create

**Create:**
```
✓ frontend/src/routes/(app)/agents/+page.svelte
  - Agent list page (if not exists)

✓ frontend/src/routes/(app)/agents/[id]/+page.svelte
  - Main detail page with all 4 tabs

✓ frontend/src/routes/(app)/agents/[id]/+page.server.ts
  - Load function for agent data

? frontend/src/routes/(app)/agents/[id]/components/OverviewTab.svelte
  - Optional: modularize if page gets large

? frontend/src/routes/(app)/agents/[id]/components/UsageStatsTab.svelte
? frontend/src/routes/(app)/agents/[id]/components/SettingsTab.svelte
? frontend/src/routes/(app)/agents/[id]/components/TestTab.svelte
```

**Modify:**
```
✓ frontend/src/lib/api/index.ts
  - Add agent detail methods

✓ frontend/src/lib/api/agents/types.ts (create if needed)
  - Add AgentDetailResponse, UsageStatsResponse, etc.
```

---

## Implementation Sequence

### Week 1: Foundation & Basics

**Day 1 - Microtask 1 (2-3 hours)**
- Backend: Create API endpoints
- Frontend: Create routes and skeleton
- Together: Test end-to-end connectivity

**Days 2-3 - Microtasks 2, 3, 4 (6-9 hours)**
- Backend: Can work on other features
- Frontend (in parallel):
  - Overview Tab (straightforward)
  - Usage Stats Tab (more complex)
  - Settings Tab (form heavy)

### Week 2: Integration & Polish

**Days 4-5 - Microtasks 5, 6, 7 (4-6 hours)**
- Frontend: Tab navigation
- Frontend: Testing tab
- Frontend: Header & polish

**Days 5-6 - Verification (4-6 hours)**
- Testing
- Bug fixes
- Performance optimization
- Mobile responsiveness
- Accessibility

---

## Code Examples to Use

### Frontend: Basic Page Structure

```svelte
<script lang="ts">
    import { page } from '$app/stores';
    import type { AgentDetailResponse } from '$lib/api';

    export let data; // From +page.server.ts

    let agent: AgentDetailResponse = data.agent;
    let activeTab = $page.url.searchParams.get('tab') || 'overview';

    type TabType = 'overview' | 'usage-stats' | 'settings' | 'test';

    function selectTab(tab: TabType) {
        activeTab = tab;
        const url = new URL($page.url);
        url.searchParams.set('tab', tab);
        window.history.pushState({}, '', url);
    }
</script>

<!-- Header -->
<div class="bg-white border-b">
    <div class="px-6 py-4">
        <h1 class="text-2xl font-bold">{agent.display_name}</h1>
    </div>
</div>

<!-- Tabs -->
<div class="border-b">
    <div class="px-6 flex gap-6">
        <button onclick={() => selectTab('overview')}
                class={activeTab === 'overview' ? 'active' : ''}>
            Overview
        </button>
        <button onclick={() => selectTab('usage-stats')}
                class={activeTab === 'usage-stats' ? 'active' : ''}>
            Usage Stats
        </button>
        <button onclick={() => selectTab('settings')}
                class={activeTab === 'settings' ? 'active' : ''}>
            Settings
        </button>
        <button onclick={() => selectTab('test')}
                class={activeTab === 'test' ? 'active' : ''}>
            Test
        </button>
    </div>
</div>

<!-- Content -->
<div class="p-6">
    {#if activeTab === 'overview'}
        <!-- Overview content -->
    {:else if activeTab === 'usage-stats'}
        <!-- Usage stats content -->
    {:else if activeTab === 'settings'}
        <!-- Settings content -->
    {:else if activeTab === 'test'}
        <!-- Test content -->
    {/if}
</div>
```

### Backend: Simple Endpoint

```go
func (h *Handlers) GetAgentDetailsWithMetrics(c *gin.Context) {
    user := middleware.GetCurrentUser(c)
    if user == nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
        return
    }

    idStr := c.Param("id")
    id, err := uuid.Parse(idStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID"})
        return
    }

    ctx := context.Background()
    queries := sqlc.New(h.pool)

    agent, err := queries.GetCustomAgent(ctx, sqlc.GetCustomAgentParams{
        ID:     pgtype.UUID{Bytes: id, Valid: true},
        UserID: user.ID,
    })
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
        return
    }

    // TODO: Add metrics calculation

    c.JSON(http.StatusOK, gin.H{"agent": agent})
}
```

---

## Common Pitfalls to Avoid

### ❌ Don't:
- Start coding without reading microtasks document
- Build all UI before testing endpoints
- Forget authentication checks on backend
- Use hardcoded IDs for testing
- Skip mobile testing
- Ignore form validation

### ✅ Do:
- Test each endpoint individually first
- Use proper error handling everywhere
- Validate all inputs (frontend + backend)
- Check user authorization on backend
- Build mobile-first
- Follow BusinessOS patterns

---

## Testing Commands

### Backend Testing

```bash
# Start server
go run cmd/server/main.go

# Test endpoint
curl -X GET http://localhost:8080/api/agents/[ID] \
  -H "Authorization: Bearer [TOKEN]"

# View response
curl -X GET http://localhost:8080/api/agents/[ID] | jq

# Test update
curl -X PUT http://localhost:8080/api/agents/[ID] \
  -H "Content-Type: application/json" \
  -d '{"display_name":"New Name"}'

# Test delete
curl -X DELETE http://localhost:8080/api/agents/[ID]
```

### Frontend Testing

```bash
# Development server
npm run dev

# Navigate to
http://localhost:5173/agents/[ID]

# TypeScript check
npm run build

# Type checking
npm run check
```

---

## Quick Reference: Dependencies

```
Nothing needed    →  MT-1 (Route Setup)
                 ↓
              MT-1 Ready
              ↙ ↓ ↘
            /  |  \
          /    |    \
    MT-2 (Overview)
    MT-3 (Usage)
    MT-4 (Settings)
              ↘ ↓ ↙
                |
         MT-1, 2, 3, 4 Done
                |
              MT-5 (Tabs)
                |
              MT-6 (Test)
                |
              MT-7 (Header)
                |
         Integration & Verification
```

---

## Daily Stand-up Talking Points

### Day 1
- "Backend: Created API endpoints for agent detail"
- "Frontend: Created route structure and skeleton page"
- "Next: Build individual tabs"

### Day 2-3
- "Frontend: Built Overview tab"
- "Frontend: Built Usage Stats tab"
- "Frontend: Built Settings tab"
- "Backend: Ready for optimization"
- "Next: Tab navigation and testing tab"

### Day 4-5
- "Frontend: Tab navigation complete"
- "Frontend: Testing tab with streaming"
- "Frontend: Header and polish"
- "Next: Integration testing and mobile check"

### Day 6
- "Testing: E2E tests passing"
- "Mobile: Responsive on all devices"
- "Accessibility: WCAG 2.1 AA compliance"
- "Ready for merge"

---

## Success = Done When

Each microtask is done when:
- [x] Code written and tested
- [x] No TypeScript errors
- [x] No console errors
- [x] Acceptance criteria met
- [x] Unit tests passing (if applicable)
- [x] Code reviewed and approved

Full page is done when:
- [x] All 7 microtasks complete
- [x] Integration tests pass
- [x] Mobile responsive
- [x] Accessibility checked
- [x] Performance OK
- [x] Security review passed

---

## Need Help?

### Quick Questions?
→ Check the specific microtask in `AGENT_DETAIL_PAGE_MICROTASKS.md`

### Visual/Design Questions?
→ Check `AGENT_DETAIL_PAGE_VISUAL_GUIDE.md`

### Architecture Questions?
→ Check `AGENT_DETAIL_IMPLEMENTATION_SUMMARY.md`

### Code Examples?
→ Look at `/frontend/src/routes/(app)/clients/[id]/+page.svelte` or `/nodes/[id]/+page.svelte`

### Confused About Flow?
→ Read dependency graph in this file or in MICROTASKS document

---

## Git Workflow

```bash
# 1. Create branch
git checkout -b feature/agent-detail-page

# 2. After each microtask, commit
git add .
git commit -m "feat(agent): Implement [MT-#] [description]"

# 3. Push for review
git push origin feature/agent-detail-page

# 4. Create PR when ready
# PR description should list completed microtasks

# 5. After review + approval
git merge

# 6. Deploy to staging/production
```

---

## Timeline Summary

| Phase | Duration | What |
|-------|----------|------|
| Setup | 30 min | Branch, read docs, setup |
| MT-1 | 1-2 hrs | Backend APIs |
| MT-2,3,4 | 6-9 hrs | Tabs (parallel) |
| MT-5,6,7 | 4-6 hrs | Navigation & Polish |
| Testing | 4-6 hrs | Verification & Fixes |
| **TOTAL** | **3-4 days** | **Feature Complete** |

*Note: This assumes 1 backend dev + 1 frontend dev working somewhat in parallel*

---

## One More Thing

**Before you start coding:**
1. Read the full AGENT_DETAIL_PAGE_MICROTASKS.md document
2. Check the pattern in existing detail pages
3. Make sure you have access to all files
4. Get API keys/endpoints needed
5. Clarify any questions with tech lead

**Then:**
1. Pick your microtask
2. Follow the implementation steps
3. Test thoroughly
4. Get code reviewed
5. Merge when approved

---

**Now go build! 🚀**

Questions? → Check the docs above or ask in standup.

Created: January 8, 2026
For: Agent Detail Page Implementation
Status: Ready to start
