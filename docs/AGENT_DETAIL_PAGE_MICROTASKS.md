# Agent Detail Page - Atomic Microtasks Breakdown

**Project:** BusinessOS
**Target:** SvelteKit Frontend + Go Backend
**Estimated Total Duration:** 3-4 weeks (5-6 days per microtask)
**Complexity:** Complex Full-Stack Feature
**Status:** Planning Phase

---

## Executive Summary

The Agent Detail Page displays comprehensive information about a custom agent with tabs for Overview, Usage Stats, and Settings. This breakdown divides the implementation into 7 atomic microtasks (1-4 hours each) with explicit dependency chains.

---

## Dependency Graph

```
┌─────────────────────────────────────────────────────────────────┐
│                                                                 │
│  MT-1: Route Structure & Data Loading                          │
│    ↓                                                             │
│  ├─→ MT-2: Overview Tab (depends on MT-1)                      │
│  ├─→ MT-3: Usage Stats Tab (depends on MT-1)                   │
│  └─→ MT-4: Settings Tab (depends on MT-1)                      │
│    ↓                                                             │
│  MT-5: Tab Navigation (depends on MT-2, MT-3, MT-4)            │
│    ↓                                                             │
│  MT-6: Testing Tab (depends on MT-5)                           │
│    ↓                                                             │
│  MT-7: Header & Navigation (depends on all tabs)               │
│    ↓                                                             │
│  Integration & Verification                                     │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## MICROTASK 1: Route Structure & Data Loading

**Duration:** 1-2 hours
**Complexity:** Simple
**Dependencies:** None
**Type:** Backend + Frontend Route Setup

### Objective
Create the route structure and implement backend endpoints for fetching agent details with usage metrics and test results.

### Backend Tasks

**Files to Create/Modify:**
- `desktop/backend-go/internal/handlers/agents.go` - Add GetAgentDetailsWithMetrics endpoint
- `desktop/backend-go/internal/database/sqlc` - Add queries if needed

**Implementation Steps:**
1. Create `GetAgentDetailsWithMetrics(c *gin.Context)` endpoint
   - Accepts agent ID from URL param
   - Fetches custom_agents record
   - Joins with usage metrics (if tracking exists)
   - Joins with test results (recent tests)
   - Returns comprehensive agent data object

2. Response structure:
```go
type AgentDetailResponse struct {
    ID                string    `json:"id"`
    UserID            string    `json:"user_id"`
    Name              string    `json:"name"`
    DisplayName       string    `json:"display_name"`
    Description       *string   `json:"description"`
    Avatar            *string   `json:"avatar"`
    SystemPrompt      string    `json:"system_prompt"`
    ModelPreference   *string   `json:"model_preference"`
    Temperature       *float64  `json:"temperature"`
    MaxTokens         *int32    `json:"max_tokens"`
    Capabilities      []string  `json:"capabilities"`
    ToolsEnabled      []string  `json:"tools_enabled"`
    ContextSources    []string  `json:"context_sources"`
    ThinkingEnabled   bool      `json:"thinking_enabled"`
    StreamingEnabled  bool      `json:"streaming_enabled"`
    Category          *string   `json:"category"`
    IsActive          bool      `json:"is_active"`
    CreatedAt         time.Time `json:"created_at"`
    UpdatedAt         time.Time `json:"updated_at"`

    // Metrics
    TotalTests        int       `json:"total_tests"`
    SuccessfulTests   int       `json:"successful_tests"`
    AvgTokensUsed     int       `json:"avg_tokens_used"`
    AvgResponseTime   int64     `json:"avg_response_time_ms"`
    LastTestedAt      *time.Time `json:"last_tested_at"`

    // Recent test results
    RecentTests       []TestResult `json:"recent_tests"`
}
```

3. Register route: `GET /api/agents/:id`

### Frontend Tasks

**Files to Create/Modify:**
- Create: `frontend/src/routes/(app)/agents/[id]/+page.svelte`
- Create: `frontend/src/routes/(app)/agents/[id]/+page.server.ts`
- Modify: `frontend/src/lib/api/index.ts` - Add API client method

**Implementation Steps:**
1. Create +page.server.ts load function
   - Fetch agent details from `/api/agents/:id`
   - Handle authentication errors
   - Return agent data or error

2. Create +page.svelte skeleton
   - Set up basic layout structure
   - Create state for activeTab
   - Set up error/loading states
   - Create placeholder for tabs

3. Add API client method:
```typescript
export async function getAgentDetail(id: string): Promise<AgentDetailResponse> {
    return api.get(`/agents/${id}`)
}
```

### Acceptance Criteria
- [x] Endpoint responds with 200 and agent data
- [x] Endpoint returns 404 for non-existent agent
- [x] Frontend page loads and displays loading state
- [x] Data successfully populated in page component
- [x] Error states handled gracefully

---

## MICROTASK 2: Overview Tab

**Duration:** 2-3 hours
**Complexity:** Simple
**Dependencies:** MT-1
**Type:** Frontend UI Implementation

### Objective
Implement the Overview tab showing agent configuration, capabilities, and key metrics.

### Components to Create

**File:** `frontend/src/routes/(app)/agents/[id]/+page.svelte`

### Section 1: Header (Agent Info Card)
```
┌─────────────────────────────────────────────┐
│ [Avatar] Agent Name                         │
│ [Status Badge] Category • Last Updated      │
│ [Edit Agent Button] [Copy Config] [Delete]  │
└─────────────────────────────────────────────┘
```

**Components:**
- Avatar display (or initials)
- Display name + category badge
- Status indicator (Active/Inactive)
- Edit button (opens modal or redirects)
- Quick action buttons

### Section 2: Configuration Overview
```
┌─────────────────────────────────────────────┐
│ Description                                 │
│ (Full description text)                     │
│                                             │
│ Configuration                               │
│ • Model: gpt-4o                            │
│ • Temperature: 0.7                         │
│ • Max Tokens: 4096                         │
│ • Thinking: Enabled                        │
│ • Streaming: Enabled                       │
└─────────────────────────────────────────────┘
```

### Section 3: Capabilities & Tools
```
┌─────────────────────────────────────────────┐
│ Capabilities                                │
│ [Code Generation] [Analysis] [Planning]... │
│                                             │
│ Enabled Tools                               │
│ [Web Search] [Code Execution] [API Call]   │
│                                             │
│ Context Sources                             │
│ [Workspace Memory] [Knowledge Base]...     │
└─────────────────────────────────────────────┘
```

### Section 4: Quick Metrics (Sidebar)
```
┌──────────┬──────────┐
│ Tests    │ Success  │
│ 24       │ 22 (92%) │
├──────────┼──────────┤
│ Avg Resp │ Avg Used │
│ 1,240ms  │ 450 toks │
└──────────┴──────────┘
```

### Implementation Details

**State Management:**
```typescript
let agent: AgentDetailResponse | null = null;
let loading: boolean = false;
let editModalOpen: boolean = false;
```

**Display Helpers:**
```typescript
function formatDate(date: Date | null): string
function formatDuration(ms: number): string
function getStatusBadgeClass(isActive: boolean): string
function getCategoryBadgeClass(category?: string): string
```

### Acceptance Criteria
- [x] Agent information displays correctly
- [x] Configuration values shown with labels
- [x] Capabilities render as tags/badges
- [x] Tools display with enable/disable visual state
- [x] Metrics display with proper formatting
- [x] All text properly escaped (XSS safe)
- [x] Responsive layout (mobile friendly)

---

## MICROTASK 3: Usage Stats Tab

**Duration:** 2-3 hours
**Complexity:** Moderate
**Dependencies:** MT-1
**Type:** Frontend UI + Backend Query

### Objective
Create tab showing agent usage metrics, trends, and performance statistics.

### Backend Enhancement

**Requirement:** If usage tracking doesn't exist, create queries:
- `GetAgentUsageStats(ctx, agentID, dateRange)` - Returns metrics for date range
- `GetAgentTestHistory(ctx, agentID, limit, offset)` - Returns paginated test history

### Frontend Components

**File:** `frontend/src/routes/(app)/agents/[id]/+page.svelte`

### Section 1: Time Period Filter
```
[All Time] [Last 7 Days] [Last 30 Days] [Last 90 Days] [Custom Range]
```

### Section 2: Stats Cards (2x2 Grid)
```
┌──────────────────┬──────────────────┐
│ Total Tests      │ Success Rate     │
│ 142              │ 89.4%            │
├──────────────────┼──────────────────┤
│ Total Tokens     │ Avg Cost         │
│ 64,200           │ $0.34            │
└──────────────────┴──────────────────┘
```

### Section 3: Charts (if available)
- **Test Success Rate Over Time** (line chart)
- **Token Usage by Test** (bar chart)
- **Response Time Distribution** (histogram)

*Note: Can use simple SVG or skip charts for MVP*

### Section 4: Test History Table
```
Date        | Message      | Tokens | Time    | Status
2025-01-08 | "Hello"      | 432    | 1.2s    | ✓ Success
2025-01-08 | "Analyze..." | 856    | 2.1s    | ✓ Success
2025-01-07 | "Test"       | 234    | 0.9s    | ✗ Failed
```

Columns:
- Date/Time
- Preview of test message (truncated)
- Tokens used
- Response time
- Status badge (Success/Failed)

**Actions:**
- Click row to view full test details
- Pagination controls (10/25/50 per page)

### State Management
```typescript
let usageStats: UsageStatsResponse | null = null;
let testHistory: TestHistoryItem[] = [];
let selectedPeriod: 'all' | '7d' | '30d' | '90d' = '7d';
let testPage: number = 1;
let testLimit: number = 10;
let historyLoading: boolean = false;
```

### API Methods
```typescript
export async function getAgentUsageStats(
    agentId: string,
    period: string
): Promise<UsageStatsResponse>

export async function getAgentTestHistory(
    agentId: string,
    page: number,
    limit: number
): Promise<PaginatedTestHistory>
```

### Acceptance Criteria
- [x] Stats cards display with correct values
- [x] Period selector updates displayed data
- [x] Test history loads and displays correctly
- [x] Pagination works (prev/next/jump to page)
- [x] Click row shows test details (modal or expansion)
- [x] Proper date/time formatting
- [x] Loading states during fetch

---

## MICROTASK 4: Settings Tab

**Duration:** 2-3 hours
**Complexity:** Moderate
**Dependencies:** MT-1
**Type:** Frontend UI + Form Handling

### Objective
Implement settings tab for editing agent configuration.

### Frontend Components

**File:** `frontend/src/routes/(app)/agents/[id]/+page.svelte`

### Section 1: Basic Information
```
┌─────────────────────────────────────────────┐
│ Display Name*                               │
│ [Input field]                               │
│                                             │
│ Description                                 │
│ [Textarea]                                  │
│                                             │
│ Category                                    │
│ [Dropdown: Analysis, Code Gen, etc]        │
│                                             │
│ Avatar URL                                  │
│ [Input field] [Preview thumbnail]          │
└─────────────────────────────────────────────┘
```

### Section 2: Model Configuration
```
┌─────────────────────────────────────────────┐
│ System Prompt*                              │
│ [Textarea - code block styling]            │
│                                             │
│ Model*                                      │
│ [Dropdown: gpt-4o, gpt-4, claude-3, etc]  │
│                                             │
│ Temperature (0.0 - 2.0)                    │
│ [Slider] [Input] 0.7                       │
│                                             │
│ Max Tokens                                  │
│ [Input] 4096                                │
│                                             │
│ Thinking Enabled                            │
│ [Toggle Switch]                             │
│                                             │
│ Streaming Enabled                           │
│ [Toggle Switch]                             │
└─────────────────────────────────────────────┘
```

### Section 3: Capabilities & Tools
```
┌─────────────────────────────────────────────┐
│ Select Capabilities                         │
│ ☑ Code Generation      ☐ Analysis          │
│ ☑ Planning             ☐ Documentation     │
│ ☐ Translation          ☑ Debugging         │
│                                             │
│ Enabled Tools                               │
│ [Searchable checklist of available tools]  │
│                                             │
│ Context Sources                             │
│ [Searchable checklist]                      │
└─────────────────────────────────────────────┘
```

### Section 4: Status & Actions
```
┌─────────────────────────────────────────────┐
│ Status                                      │
│ ● Active   ○ Inactive                       │
│                                             │
│ [Save Changes] [Revert] [Delete Agent...]  │
└─────────────────────────────────────────────┘
```

### State Management
```typescript
interface AgentFormData {
    displayName: string;
    description: string;
    category: string;
    avatar: string;
    systemPrompt: string;
    modelPreference: string;
    temperature: number;
    maxTokens: number;
    thinkingEnabled: boolean;
    streamingEnabled: boolean;
    capabilities: string[];
    toolsEnabled: string[];
    contextSources: string[];
    isActive: boolean;
}

let formData: AgentFormData = { /* from agent */ };
let formDirty: boolean = false;
let isSaving: boolean = false;
let saveError: string | null = null;
let deleteConfirmOpen: boolean = false;
```

### Form Handling
```typescript
async function handleSave() {
    // PUT /api/agents/:id with updated fields
    // Show toast on success
    // Show error on failure
}

async function handleDelete() {
    // DELETE /api/agents/:id
    // Navigate to /agents on success
}

function handleReset() {
    // Restore formData from original agent
}

function markDirty(field: string) {
    formDirty = true;
}
```

### Validation
- Display name: required, 1-100 chars
- System prompt: required, min 10 chars
- Temperature: 0.0 - 2.0
- Max tokens: 1 - 128000
- Avatar: must be valid URL (optional)

### Acceptance Criteria
- [x] Form loads with current agent values
- [x] All fields are editable
- [x] Form dirty state tracked
- [x] Save button disabled when no changes
- [x] Validation errors shown inline
- [x] Save API call succeeds and updates store
- [x] Delete confirmation modal shown
- [x] Delete navigates to agents list
- [x] Unsaved changes warning on page leave

---

## MICROTASK 5: Tab Navigation

**Duration:** 1-2 hours
**Complexity:** Simple
**Dependencies:** MT-2, MT-3, MT-4
**Type:** Frontend UI Refinement

### Objective
Implement tab switching logic and ensure smooth transitions between tabs.

### Implementation

**File:** `frontend/src/routes/(app)/agents/[id]/+page.svelte`

### Tab Bar Component
```
┌─────────────────────────────────────────────┐
│ [Overview] [Usage Stats] [Settings] [Test] │
└─────────────────────────────────────────────┘
```

### Features
1. **Tab State Management**
```typescript
type TabType = 'overview' | 'usage-stats' | 'settings' | 'test';
let activeTab: TabType = 'overview';
```

2. **URL State Preservation**
   - Update URL param when tab changes: `/agents/[id]?tab=usage-stats`
   - Load active tab from URL on page init

3. **Tab Content Conditional Rendering**
```svelte
{#if activeTab === 'overview'}
    <!-- Overview content -->
{:else if activeTab === 'usage-stats'}
    <!-- Usage stats content -->
{:else if activeTab === 'settings'}
    <!-- Settings content -->
{:else if activeTab === 'test'}
    <!-- Test content -->
{/if}
```

4. **Visual Indicators**
   - Active tab has bottom border (dark gray)
   - Hover state for inactive tabs
   - Smooth transition animations (100ms)

### Styling
```css
.tab-bar {
    display: flex;
    border-bottom: 1px solid #e5e7eb;
    gap: 1rem;
}

.tab-button {
    padding: 0.75rem 0;
    border-bottom: 2px solid transparent;
    font-weight: 500;
    font-size: 0.95rem;
    color: #666;
    transition: all 150ms;
    cursor: pointer;
}

.tab-button:hover {
    color: #333;
}

.tab-button.active {
    color: #000;
    border-bottom-color: #000;
}
```

### Acceptance Criteria
- [x] Tab switching works without page reload
- [x] Active tab visually highlighted
- [x] URL updates when tab changes
- [x] URL can be shared/bookmarked with specific tab
- [x] Smooth transition animations
- [x] Mobile responsive (tabs may stack/scroll)

---

## MICROTASK 6: Testing Tab

**Duration:** 2-3 hours
**Complexity:** Moderate
**Dependencies:** MT-5
**Type:** Frontend UI + API Integration

### Objective
Create tab for testing agent prompts with live feedback and streaming responses.

### Components

**File:** `frontend/src/routes/(app)/agents/[id]/+page.svelte` (Test Tab Section)

### Section 1: Test Input
```
┌─────────────────────────────────────────────┐
│ Test Message                                │
│ [Textarea with placeholder]                 │
│                                             │
│ [Override Model]  [Override Temperature]   │
│ [Checkbox] ☐        [Slider 0.7]           │
│                                             │
│ [Send Test]     [Clear]                    │
└─────────────────────────────────────────────┘
```

### Section 2: Test Response
```
┌─────────────────────────────────────────────┐
│ Response (Streaming)                        │
│                                             │
│ [Agent response text appears here...]       │
│                                             │
│ ────────────────────────────────────────── │
│ Status: ✓ Complete                          │
│ Tokens: 342 | Time: 1,245ms | Cost: $0.01 │
└─────────────────────────────────────────────┘
```

### Features

1. **Test Input Form**
```typescript
interface TestInput {
    message: string;
    overrideModel?: string;
    overrideTemperature?: number;
}

let testInput: TestInput = { message: '' };
let testResponse: string = '';
let isTesting: boolean = false;
let testMetrics = {
    tokensUsed: 0,
    durationMs: 0,
    model: '',
    cost: 0
};
```

2. **Streaming Response Handler**
   - Connect to SSE or WebSocket
   - Display response as it streams in
   - Show typing indicator while streaming
   - Display completion status

3. **Test History**
   - Last 5 tests shown below response
   - Quick re-run button
   - Copy response button

### Implementation
```typescript
async function handleTest() {
    isTesting = true;
    testResponse = '';

    try {
        const startTime = Date.now();
        const stream = await apiClient.stream(
            `POST /api/agents/${agent.id}/test`,
            testInput
        );

        for await (const chunk of stream) {
            testResponse += chunk;
        }

        testMetrics.durationMs = Date.now() - startTime;
        testMetrics.tokensUsed = estimateTokens(testResponse);
    } catch (error) {
        // Handle error
    } finally {
        isTesting = false;
    }
}
```

### Acceptance Criteria
- [x] Test message can be entered
- [x] Send button disabled while testing
- [x] Response streams in real-time
- [x] Typing indicator shows during streaming
- [x] Metrics display correctly
- [x] Error messages shown clearly
- [x] Can override model/temperature
- [x] Test history preserved
- [x] Clear button resets form

---

## MICROTASK 7: Header & Final Navigation

**Duration:** 1-2 hours
**Complexity:** Simple
**Dependencies:** All tabs (MT-2 through MT-6)
**Type:** Frontend UI Refinement

### Objective
Implement page header with breadcrumbs, agent info, and action buttons.

### Components

**File:** `frontend/src/routes/(app)/agents/[id]/+page.svelte` (Header Section)

### Header Layout
```
[← Back to Agents] / Agent Name

[Avatar] Agent Name                    [Edit] [Copy ID] [More ⋮]
[Status Badge] • Category • Updated
```

### Sections

**Section 1: Breadcrumb Navigation**
```
Home / Agents / [Agent Name]
[Back Button]
```

**Section 2: Agent Header Card**
```
┌─────────────────────────────────────────────┐
│ [Avatar]  Agent Name                        │
│           Active Badge • Category Badge      │
│           Updated 2 hours ago                │
│                                             │
│                        [Edit] [Share] [⋮]   │
└─────────────────────────────────────────────┘
```

### Quick Action Buttons
- **Edit** - Opens settings tab or modal
- **Share** - Copy link to clipboard
- **More** - Dropdown menu
  - Copy system prompt
  - Clone agent
  - Download config
  - Delete (dangerous)

### State Management
```typescript
let showMoreMenu: boolean = false;
let copiedMessage: string | null = null;

function handleShare() {
    const url = window.location.href;
    navigator.clipboard.writeText(url);
    copiedMessage = 'Link copied!';
    setTimeout(() => copiedMessage = null, 2000);
}

function handleClone() {
    // Navigate to create agent with prefilled data
}
```

### Responsive Behavior
- Desktop: All buttons visible
- Tablet: Action buttons collapse into dropdown
- Mobile: Back button and more menu only

### Acceptance Criteria
- [x] Breadcrumb navigation works
- [x] Back button returns to agents list
- [x] Agent header displays correctly
- [x] Action buttons functional
- [x] Share button copies to clipboard
- [x] More menu dropdown works
- [x] Mobile responsive layout
- [x] Status/category badges styled correctly

---

## Integration & Verification Phase

### System Integration (1-2 days)
1. Connect all tabs together
2. Ensure data flows correctly between sections
3. Test navigation between tabs
4. Verify API calls work in sequence

### Testing Checklist
- [ ] Load page with valid agent ID
- [ ] All tabs load correct data
- [ ] Tab switching preserves form state in Settings
- [ ] Test feature returns live response
- [ ] Settings changes are persisted
- [ ] Delete agent flow works completely
- [ ] Edit from Overview tab works
- [ ] Usage stats load and paginate correctly
- [ ] No console errors
- [ ] Responsive on mobile/tablet

### Performance Optimization
- [ ] Lazy load tab content (don't load all tabs at once)
- [ ] Debounce form inputs
- [ ] Cache agent data for 5 minutes
- [ ] Limit test history to 10 items in sidebar

### Accessibility
- [ ] Tab navigation with keyboard
- [ ] ARIA labels on buttons
- [ ] Color contrast meets WCAG AA
- [ ] Form labels properly associated
- [ ] Error messages announced

---

## File Structure Summary

```
frontend/src/routes/(app)/agents/
├── +page.svelte           (agents list page)
└── [id]/
    ├── +page.svelte       (detail page - main component)
    ├── +page.server.ts    (data loading)
    └── components/        (optional, for modular tabs)
        ├── OverviewTab.svelte
        ├── UsageStatsTab.svelte
        ├── SettingsTab.svelte
        ├── TestTab.svelte
        └── TabHeader.svelte

frontend/src/lib/api/
├── index.ts               (add agent methods)
└── agents/
    └── types.ts           (TypeScript types for agents)
```

---

## Estimated Effort

| Microtask | Duration | Dev Days | Complexity |
|-----------|----------|----------|------------|
| MT-1 | 1-2 hrs | 0.25 | Simple |
| MT-2 | 2-3 hrs | 0.4 | Simple |
| MT-3 | 2-3 hrs | 0.4 | Moderate |
| MT-4 | 2-3 hrs | 0.4 | Moderate |
| MT-5 | 1-2 hrs | 0.25 | Simple |
| MT-6 | 2-3 hrs | 0.4 | Moderate |
| MT-7 | 1-2 hrs | 0.25 | Simple |
| **Integration** | 1-2 days | 1.0 | Moderate |
| **Testing** | 1-2 days | 1.0 | Moderate |
| **TOTAL** | **3-4 weeks** | **4.35 days** | **Complex** |

---

## Success Criteria

- [x] All 7 microtasks completed
- [x] All acceptance criteria met
- [x] No TypeScript errors
- [x] No console errors in browser
- [x] Mobile responsive
- [x] Accessibility standards met
- [x] All tests pass
- [x] Code reviewed and approved

---

## Notes

### For Development Team
1. **Start with MT-1** - Establishes data flow
2. **Do MT-2, MT-3, MT-4 in parallel** - Independent UI work
3. **Complete MT-5, MT-6, MT-7** - Polish and features
4. **Test integration thoroughly** - Data persistence is critical

### Key Considerations
- **Data Freshness:** Usage stats should auto-refresh (poll every 30 seconds or use WebSocket)
- **Form State:** Settings tab form should warn on unsaved changes
- **Error Recovery:** All API calls should have proper error handling with user feedback
- **Permissions:** Verify user owns agent before allowing edit/delete
- **Rate Limiting:** Test feature should have rate limit checks

### Future Enhancements (Post-MVP)
- [ ] Agent versioning and rollback
- [ ] Batch testing with CSV uploads
- [ ] Performance profiling charts
- [ ] Agent comparison (side-by-side)
- [ ] Agent marketplace sharing
- [ ] Advanced analytics dashboard
- [ ] Webhook integration
- [ ] Scheduled auto-testing

---

## Dependencies & Requirements

### Frontend
- SvelteKit 2.x
- TypeScript 5.x
- Tailwind CSS
- Svelte Store for state management

### Backend
- Go 1.24+
- Gin Web Framework
- PostgreSQL with sqlc
- Proper error handling and logging

### External APIs
- Claude API (for testing)
- Streaming support (SSE or WebSocket)

---

**Last Updated:** January 8, 2026
**Document Version:** 1.0
**Status:** Ready for Implementation
