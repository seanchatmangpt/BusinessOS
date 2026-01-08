# Agent Testing UI - Atomic Microtasks Breakdown

## Feature Overview
Build a complete Agent Testing UI for BusinessOS that allows users to test custom agents with various configurations, monitor performance metrics, and visualize results with loading states and error handling.

**Total Estimated Time:** 24-28 hours (6-7 days of work)
**Complexity Level:** Moderate to Complex (Full-Stack)

---

## Architecture Overview

### Component Hierarchy
```
AgentTestPanel (Main Container)
├── AgentSelector (Dropdown + Quick Preview)
├── TestInputSection
│   ├── MessageInput (Textarea)
│   ├── ConfigPanel (Temperature, Tokens, Model)
│   └── AdvancedOptions (Thinking, Streaming)
├── ExecutionPanel (Running State)
│   ├── ProgressIndicator
│   ├── TokenCounter (Real-time)
│   └── TimelineDisplay
├── ResultsPanel (Success State)
│   ├── MetadataStrip (Model, Duration, Tokens, Cost)
│   ├── ResponseDisplay (Markdown)
│   ├── SourcesPanel (if RAG enabled)
│   └── CopyExportButtons
└── ErrorPanel (Error State)
    ├── ErrorMessage
    ├── ErrorStackTrace (Dev mode)
    └── RetryButton

```

### Backend Flow
```
HTTP POST /api/agents/{id}/test
    ↓
TestAgentHandler (validation)
    ↓
AgentTestService (orchestration)
    ├─→ GetAgentConfig (fetch agent)
    ├─→ BuildTestContext (setup)
    ├─→ InvokeAgent (run)
    └─→ RecordMetrics (logging)
    ↓
Response with streaming or sync result
```

### Database Dependencies
- `custom_agents` table (already exists)
- `agent_test_runs` table (NEW - track history)
- `agent_test_metrics` table (NEW - performance data)

---

## Microtasks (Atomic Units: 2-4 Hours Each)

### TIER 1: Database Foundation (1 Microtask)

#### MICROTASK 1.1: Create Database Migration for Agent Testing Tables
**Estimated Time:** 2-3 hours
**Priority:** Critical (blocks all backend)
**Dependencies:** None

**Scope:**
- Create migration `037_agent_testing.sql`
- Table `agent_test_runs`: id, agent_id, user_id, workspace_id, test_message, response, model_used, duration_ms, tokens_used, status, created_at
- Table `agent_test_metrics`: id, test_run_id, metric_name, metric_value, recorded_at
- Table `agent_test_history`: id, agent_id, user_id, test_count, last_test_at, avg_duration_ms, success_rate
- Add indexes on agent_id, user_id, created_at
- Add foreign keys to custom_agents

**Acceptance Criteria:**
- Migration applies cleanly to PostgreSQL
- Tables created with correct schema
- Indexes created for query performance
- Foreign key constraints enforced

**SQL Script Checklist:**
```sql
-- agent_test_runs table
CREATE TABLE agent_test_runs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  agent_id UUID NOT NULL REFERENCES custom_agents(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
  workspace_id UUID NOT NULL,
  test_message TEXT NOT NULL,
  response TEXT,
  model_used VARCHAR(255),
  duration_ms INTEGER,
  tokens_used INTEGER,
  status VARCHAR(50) CHECK (status IN ('pending', 'running', 'completed', 'failed')),
  error_message TEXT,
  created_at TIMESTAMPTZ DEFAULT now(),
  updated_at TIMESTAMPTZ DEFAULT now()
);

-- Indexes
CREATE INDEX idx_agent_test_runs_agent_id ON agent_test_runs(agent_id);
CREATE INDEX idx_agent_test_runs_user_id ON agent_test_runs(user_id);
CREATE INDEX idx_agent_test_runs_created_at ON agent_test_runs(created_at DESC);

-- Metrics table
CREATE TABLE agent_test_metrics (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  test_run_id UUID NOT NULL REFERENCES agent_test_runs(id) ON DELETE CASCADE,
  metric_name VARCHAR(255) NOT NULL,
  metric_value NUMERIC NOT NULL,
  recorded_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_agent_test_metrics_test_run_id ON agent_test_metrics(test_run_id);
```

**Git Commit Message:**
```
database: Add agent testing tables migration

- Create agent_test_runs table for tracking test execution
- Create agent_test_metrics table for performance metrics
- Add indexes for query optimization
- Add foreign key constraints
- Includes status tracking (pending/running/completed/failed)

Implements foundation for agent testing UI feature.
```

---

### TIER 2: Backend API Layer (3 Microtasks)

#### MICROTASK 2.1: Create SQLC Queries for Agent Testing
**Estimated Time:** 2-3 hours
**Priority:** Critical (blocks service layer)
**Dependencies:** MICROTASK 1.1

**Scope:**
- Create `queries/agent_testing.sql` with SQLC queries:
  - `InsertAgentTestRun` - insert new test execution
  - `UpdateAgentTestRun` - update test result
  - `GetAgentTestRun` - fetch single test
  - `ListAgentTestRuns` - list with pagination
  - `InsertTestMetric` - record metric
  - `ListTestMetrics` - fetch metrics for run
  - `GetAgentTestHistory` - fetch summary stats

**Acceptance Criteria:**
- All queries compile without errors
- Queries handle NULL values correctly
- Pagination implemented (limit/offset)
- Generated sqlc code matches patterns in codebase

**Queries to Create:**
```sql
-- name: InsertAgentTestRun :one
INSERT INTO agent_test_runs (
  agent_id, user_id, workspace_id, test_message, status, created_at
) VALUES ($1, $2, $3, $4, 'running', now())
RETURNING id, agent_id, test_message, status, created_at;

-- name: UpdateAgentTestRun :exec
UPDATE agent_test_runs
SET response = $2,
    model_used = $3,
    duration_ms = $4,
    tokens_used = $5,
    status = $6,
    updated_at = now()
WHERE id = $1 AND user_id = $7;

-- name: GetAgentTestRun :one
SELECT * FROM agent_test_runs
WHERE id = $1 AND user_id = $2;

-- name: ListAgentTestRuns :many
SELECT * FROM agent_test_runs
WHERE agent_id = $1 AND user_id = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: InsertTestMetric :exec
INSERT INTO agent_test_metrics (test_run_id, metric_name, metric_value)
VALUES ($1, $2, $3);

-- name: ListTestMetrics :many
SELECT * FROM agent_test_metrics
WHERE test_run_id = $1
ORDER BY recorded_at;
```

**Git Commit Message:**
```
database: Add SQLC queries for agent testing

- Create InsertAgentTestRun for recording test execution
- Create UpdateAgentTestRun for storing results
- Create ListAgentTestRuns for test history
- Create metric insertion and retrieval queries
- Add pagination support to list queries

Enables service layer to persist test data.
```

---

#### MICROTASK 2.2: Create Agent Test Service
**Estimated Time:** 3-4 hours
**Priority:** Critical (blocks handler)
**Dependencies:** MICROTASK 2.1

**Scope:**
- Create `internal/services/agent_testing_service.go`
- Implement `AgentTestService` struct with:
  - `pool *pgxpool.Pool`
  - `llm services.LLMService`
  - `agentBridge *sorx.AgentBridge`
  - `metricsCollector *MetricsCollector`

- Implement methods:
  - `RunTest(ctx, agentID, testMsg, config) (*TestResult, error)`
  - `RecordTestRun(ctx, run *TestRun) error`
  - `GetTestHistory(ctx, agentID, userID, limit) ([]*TestRun, error)`
  - `CalculateMetrics(ctx, testRunID) (*TestMetrics, error)`

**Key Functions:**
- Start timer, run agent, capture output
- Track tokens, duration, errors
- Save to database
- Return formatted result

**Acceptance Criteria:**
- Service compiles without errors
- All methods have error handling
- Metrics collection works correctly
- Database operations are transactional
- Proper context propagation

**Service Structure:**
```go
type AgentTestService struct {
	pool                *pgxpool.Pool
	llm                 services.LLMService
	agentBridge         *sorx.AgentBridge
	metricsCollector    *MetricsCollector
	queries             *sqlc.Queries
}

type TestResult struct {
	ID           uuid.UUID
	AgentID      uuid.UUID
	TestMessage  string
	Response     string
	ModelUsed    string
	DurationMS   int
	TokensUsed   int
	CostUSD      float64
	Status       string
	Metrics      map[string]interface{}
	Error        string
	CreatedAt    time.Time
}

type TestConfig struct {
	Temperature     float32
	MaxTokens       int32
	ThinkingEnabled bool
	StreamingMode   bool
	ModelOverride   string
}
```

**Git Commit Message:**
```
feat: Add AgentTestService for test orchestration

- Create AgentTestService to coordinate test execution
- Implement RunTest method for executing tests with config
- Add metrics collection (duration, tokens, cost)
- Implement database persistence via SQLC
- Add error handling and logging

Enables agents to be tested with various configurations.
```

---

#### MICROTASK 2.3: Create Agent Test Handler & API Endpoint
**Estimated Time:** 2-3 hours
**Priority:** Critical (enables frontend)
**Dependencies:** MICROTASK 2.2

**Scope:**
- Create handler in `internal/handlers/agent_testing.go`
- Implement `TestAgentRequest` struct
- Implement `TestAgent` handler
- Add route: `POST /api/agents/:id/test`
- Add route: `GET /api/agents/:id/test-history`

**Request/Response Types:**
```go
type TestAgentRequest struct {
	TestMessage     string                 `json:"test_message" binding:"required"`
	Temperature     *float32               `json:"temperature"`
	MaxTokens       *int32                 `json:"max_tokens"`
	ThinkingEnabled *bool                  `json:"thinking_enabled"`
	StreamingMode   *bool                  `json:"streaming_mode"`
	ModelOverride   *string                `json:"model_override"`
	IncludeMetrics  bool                   `json:"include_metrics"`
}

type TestAgentResponse struct {
	ID            uuid.UUID              `json:"id"`
	AgentID       uuid.UUID              `json:"agent_id"`
	Response      string                 `json:"response"`
	ModelUsed     string                 `json:"model_used"`
	DurationMS    int                    `json:"duration_ms"`
	TokensUsed    int                    `json:"tokens_used"`
	CostUSD       float64                `json:"cost_usd"`
	Status        string                 `json:"status"`
	Metrics       map[string]interface{} `json:"metrics,omitempty"`
	Error         string                 `json:"error,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
}
```

**Handler Implementation:**
- Validate user is authenticated
- Validate agent exists and is owned by user
- Validate test message not empty
- Call AgentTestService.RunTest
- Return formatted response
- Handle errors appropriately (400/500)

**Acceptance Criteria:**
- Handler compiles without errors
- Proper authentication/authorization
- Input validation works
- Returns correct response format
- Error handling comprehensive

**Git Commit Message:**
```
feat: Add TestAgent HTTP handler and endpoints

- Create POST /api/agents/:id/test endpoint
- Create GET /api/agents/:id/test-history endpoint
- Implement input validation and auth checks
- Return structured test results with metrics
- Add proper error handling

Enables frontend to test agents via API.
```

---

### TIER 3: Frontend Components (3 Microtasks)

#### MICROTASK 3.1: Create Base AgentTestPanel Component
**Estimated Time:** 2-3 hours
**Priority:** High (main container)
**Dependencies:** MICROTASK 2.3

**Scope:**
- Create `src/lib/components/settings/AgentTestPanel.svelte`
- Implement main container component with:
  - State management (currentAgent, testMessage, isLoading, error)
  - Agent selector dropdown
  - Tab navigation (Playground, History, Advanced)
  - Sub-component slots for:
    - InputSection
    - ExecutionPanel
    - ResultsPanel
    - ErrorPanel

**Component Structure:**
```svelte
<script lang="ts">
  import AgentTestInput from './AgentTestInput.svelte';
  import AgentTestExecution from './AgentTestExecution.svelte';
  import AgentTestResults from './AgentTestResults.svelte';
  import AgentTestError from './AgentTestError.svelte';

  let agents = $state<CustomAgent[]>([]);
  let selectedAgent = $state<CustomAgent | null>(null);
  let testMessage = $state('');
  let isLoading = $state(false);
  let currentTab = $state('playground');
  let error = $state<string | null>(null);
  let testResult = $state<TestResult | null>(null);
</script>
```

**Features:**
- Dropdown to select agent from list
- Show agent preview (name, description, avatar)
- Tabs for: Playground, History, Advanced settings
- Responsive layout

**Acceptance Criteria:**
- Component renders without errors
- Dropdown populated from API
- Tabs switch correctly
- All sub-components mount properly
- Proper state initialization

**Git Commit Message:**
```
feat: Add AgentTestPanel main container component

- Create main test panel component with tabs
- Implement agent selection dropdown
- Add state management for test workflow
- Structure for input, execution, results sections
- Add tab navigation (Playground, History, Advanced)

Foundation for agent testing UI.
```

---

#### MICROTASK 3.2: Create Test Input & Configuration Components
**Estimated Time:** 2-3 hours
**Priority:** High (core UX)
**Dependencies:** MICROTASK 3.1

**Scope:**
- Create `src/lib/components/settings/AgentTestInput.svelte`
- Create `src/lib/components/settings/AgentTestConfig.svelte`

**AgentTestInput Component:**
- Textarea for test message
- Character counter
- Placeholder text
- Clear button
- Disabled state when loading

**AgentTestConfig Component:**
- Temperature slider (0.0 - 2.0)
- Max tokens slider (100 - 4000)
- Model selector dropdown
- Thinking toggle (on/off)
- Streaming toggle (on/off)
- Advanced options collapsible

**Features:**
- Real-time character count
- Input validation
- Config state management
- Reset to defaults button
- Save config as preset (optional for phase 2)

**Acceptance Criteria:**
- Textareas render with proper styling
- Sliders work correctly with bounds
- Toggles function properly
- Character count updates in real-time
- Config state properly managed
- Responsive design (mobile-friendly)

**Git Commit Message:**
```
feat: Add test input and configuration components

- Create AgentTestInput for message entry
- Create AgentTestConfig for temperature/tokens/model
- Implement sliders for numeric config
- Add character counter and validation
- Include toggles for thinking/streaming modes

Enables users to configure test parameters.
```

---

#### MICROTASK 3.3: Create Execution & Results Display Components
**Estimated Time:** 3-4 hours
**Priority:** High (critical UX)
**Dependencies:** MICROTASK 3.2

**Scope:**
- Create `src/lib/components/settings/AgentTestExecution.svelte`
- Create `src/lib/components/settings/AgentTestResults.svelte`
- Create `src/lib/components/settings/AgentTestMetrics.svelte`

**AgentTestExecution Component (Loading State):**
- Animated spinner
- Real-time token counter (if streaming)
- Elapsed time display
- Estimated time remaining
- Cancel button
- Status message ("Thinking...", "Processing...", etc)

**AgentTestResults Component (Success State):**
- Metadata strip with:
  - Model name
  - Duration (ms)
  - Tokens used
  - Cost (USD)
  - Timestamp
- Response display (markdown rendering)
- Copy button
- Export buttons (JSON, text)
- Retry button
- Toggle source details (if RAG used)

**AgentTestMetrics Component:**
- Token breakdown chart (input/output)
- Timeline of execution phases
- Cost breakdown
- Performance comparison (vs previous tests)

**Features:**
- Markdown rendering for response
- Code block syntax highlighting
- Smooth transitions between states
- Copy-to-clipboard functionality
- Export options (JSON, PDF later)
- Metrics visualization with Chart.js

**Acceptance Criteria:**
- All components render correctly
- Animations smooth and performant
- Text copy functionality works
- Export buttons functional
- Responsive on mobile
- Error states display properly

**Git Commit Message:**
```
feat: Add execution and results display components

- Create AgentTestExecution for loading state
- Create AgentTestResults for displaying test output
- Create AgentTestMetrics for performance visualization
- Implement markdown rendering with code highlighting
- Add copy/export functionality
- Include animated spinners and transitions

Enables visualization of test execution and results.
```

---

### TIER 4: API Client Layer (1 Microtask)

#### MICROTASK 4.1: Create Agent Testing API Client Functions
**Estimated Time:** 2 hours
**Priority:** High (required for frontend)
**Dependencies:** MICROTASK 2.3

**Scope:**
- Create `src/lib/api/agent-testing.ts`
- Implement API functions:
  - `testAgent(agentId, message, config)`
  - `getAgentTestHistory(agentId, limit)`
  - `getTestResult(testRunId)`
  - `cancelAgentTest(testRunId)`

**Client Implementation:**
```typescript
export async function testAgent(
  agentId: string,
  message: string,
  config: TestConfig
): Promise<TestResult> {
  const response = await fetch(
    `/api/agents/${agentId}/test`,
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        test_message: message,
        temperature: config.temperature,
        max_tokens: config.maxTokens,
        thinking_enabled: config.thinkingEnabled,
        streaming_mode: config.streamingMode,
        include_metrics: true
      })
    }
  );
  return response.json();
}
```

**Features:**
- Proper error handling
- Type-safe requests/responses
- Request timeout handling
- Retry logic for failed requests
- Streaming support (for future)

**Acceptance Criteria:**
- Functions compile without type errors
- Error handling works
- Request/response types match backend
- API calls properly authenticated
- Timeout handling implemented

**Git Commit Message:**
```
feat: Add agent testing API client

- Create testAgent function for executing tests
- Add getAgentTestHistory for fetching history
- Implement proper error handling
- Type-safe request/response types
- Add timeout and retry logic

Enables frontend components to call testing endpoints.
```

---

### TIER 5: Error Handling & Loading States (2 Microtasks)

#### MICROTASK 5.1: Create Error Panel Component & Error States
**Estimated Time:** 2 hours
**Priority:** Medium (enhances UX)
**Dependencies:** MICROTASK 3.3

**Scope:**
- Create `src/lib/components/settings/AgentTestError.svelte`
- Implement error handling states:
  - Network error
  - Agent not found
  - Invalid message
  - Timeout
  - Server error (5xx)
  - LLM API error

**ErrorPanel Component:**
- Error icon
- Error title and message
- Error code (if applicable)
- Stack trace (dev mode only)
- Retry button
- Report bug link
- Copy error details button

**Error Type Detection:**
- Map HTTP status codes to user-friendly messages
- Distinguish network vs application errors
- Provide actionable suggestions

**Acceptance Criteria:**
- Error states display clearly
- User-friendly messages shown
- Retry functionality works
- Dev mode stack trace works
- No console errors from error handling

**Git Commit Message:**
```
feat: Add error handling and error display component

- Create AgentTestError component for error states
- Implement error type detection and classification
- Add user-friendly error messages
- Include retry functionality
- Add dev mode stack trace display

Improves error UX and debugging.
```

---

#### MICROTASK 5.2: Create Loading States & Spinners Component
**Estimated Time:** 1.5-2 hours
**Priority:** Medium (polish)
**Dependencies:** MICROTASK 5.1

**Scope:**
- Create `src/lib/components/settings/AgentTestSpinner.svelte`
- Create loading state UI:
  - Animated spinner
  - Progress indicators
  - Skeleton screens for results

**Features:**
- Smooth animations
- Multiple spinner styles
- Pulse/skeleton loading
- Progress bar (if duration available)
- Status messages ("Initializing...", "Processing...", "Finalizing...")

**Acceptance Criteria:**
- Animations smooth at 60fps
- No jank or stuttering
- Accessible (alt text)
- Works on low-end devices
- CSS animations (not JS timers)

**Git Commit Message:**
```
feat: Add loading states and spinner components

- Create AgentTestSpinner with smooth animations
- Implement skeleton screens for results
- Add progress indicators for long operations
- Create status messages for different phases
- Optimize animations for performance

Improves perceived performance and UX.
```

---

### TIER 6: Integration & Testing (3 Microtasks)

#### MICROTASK 6.1: Create End-to-End Integration Test
**Estimated Time:** 2-3 hours
**Priority:** High (validates flow)
**Dependencies:** MICROTASK 5.2

**Scope:**
- Create `frontend/src/lib/components/settings/__tests__/AgentTestPanel.integration.test.ts`
- Create `desktop/backend-go/internal/handlers/agent_testing_test.go`

**Frontend Tests:**
- Mount component with mock agent
- Type message
- Click test button
- Wait for result
- Verify result displays
- Test error state
- Test retry

**Backend Tests:**
- Test handler with valid request
- Test handler with invalid agent ID
- Test service database persistence
- Test metrics calculation
- Test error responses

**Acceptance Criteria:**
- All tests pass
- >80% code coverage
- No flaky tests
- Tests run in <5 seconds
- CI/CD integration

**Git Commit Message:**
```
test: Add integration tests for agent testing

- Create frontend integration tests for full workflow
- Add backend handler unit tests
- Test database persistence
- Verify error handling
- Add metrics calculation tests

Ensures feature quality and reliability.
```

---

#### MICROTASK 6.2: Create Component Tests (Unit)
**Estimated Time:** 2-3 hours
**Priority:** Medium (coverage)
**Dependencies:** MICROTASK 6.1

**Scope:**
- Create unit tests for each component:
  - AgentTestInput (character count, validation)
  - AgentTestConfig (slider bounds, config state)
  - AgentTestResults (markdown rendering, copy)
  - AgentTestError (error message display)
  - AgentTestExecution (timer, spinner animation)

**Test Coverage:**
- Component rendering
- State management
- Event handlers
- Edge cases (empty input, max tokens)
- Accessibility

**Acceptance Criteria:**
- 80%+ component coverage
- All edge cases tested
- No console errors
- Tests maintainable
- Snapshot tests for UI components

**Git Commit Message:**
```
test: Add unit tests for test panel components

- Test AgentTestInput validation and character count
- Test AgentTestConfig slider bounds and state
- Test AgentTestResults markdown rendering
- Test AgentTestError message display
- Test AgentTestExecution animations

Improves code quality and maintainability.
```

---

#### MICROTASK 6.3: Create Performance & E2E Tests
**Estimated Time:** 2-3 hours
**Priority:** Medium (ensures quality)
**Dependencies:** MICROTASK 6.2

**Scope:**
- Create E2E tests using Playwright/Cypress
- Create performance benchmarks
- Test workflow:
  - Login → Settings → Agent Testing
  - Select agent → Enter message
  - Configure options → Submit test
  - View results → Export

**Performance Tests:**
- Component render time <100ms
- API response time <5s
- Page load <2s
- Memory usage stable

**Accessibility Tests:**
- Keyboard navigation
- Screen reader compatibility
- Color contrast
- ARIA labels

**Acceptance Criteria:**
- E2E tests pass
- Performance meets targets
- Accessibility score >90
- Load time acceptable
- No memory leaks

**Git Commit Message:**
```
test: Add E2E and performance tests

- Create Playwright E2E test suite for full workflow
- Add performance benchmarks for components
- Test accessibility compliance
- Verify load time and memory usage
- Add lighthouse CI checks

Ensures production readiness.
```

---

### TIER 7: History & Advanced Features (2 Microtasks)

#### MICROTASK 7.1: Implement Test History Display Component
**Estimated Time:** 2-3 hours
**Priority:** Medium (nice-to-have)
**Dependencies:** MICROTASK 6.3

**Scope:**
- Create `src/lib/components/settings/AgentTestHistory.svelte`
- Display table/list of previous tests:
  - Test message (truncated)
  - Response (truncated)
  - Model used
  - Duration
  - Tokens
  - Cost
  - Date/time
  - Actions (View, Retry, Delete)

**Features:**
- Pagination (10/25/50 per page)
- Sorting by date, duration, tokens
- Filtering by model, status
- Search by message text
- Delete test runs
- View full result modal
- Export history (CSV)

**Acceptance Criteria:**
- Table displays correctly
- Pagination works
- Sorting/filtering functional
- Responsive design
- API calls performant
- No N+1 queries

**Git Commit Message:**
```
feat: Add test history component

- Create table view of previous test runs
- Implement pagination and sorting
- Add filtering by model and status
- Include export to CSV
- Add delete and view actions

Enables review of past test executions.
```

---

#### MICROTASK 7.2: Create Advanced Settings & Presets
**Estimated Time:** 2 hours
**Priority:** Low (enhancement)
**Dependencies:** MICROTASK 7.1

**Scope:**
- Create `src/lib/components/settings/AgentTestAdvanced.svelte`
- Implement:
  - Custom presets (Save/Load config)
  - Compare test results
  - Batch testing (multiple messages)
  - Test templates
  - Settings persistence (localStorage)

**Advanced Features:**
- Save current config as preset
- List saved presets
- Load preset (auto-fill config)
- Delete preset
- Compare side-by-side results
- Run multiple tests in sequence

**Acceptance Criteria:**
- Presets save/load correctly
- Comparison view displays both results
- Batch testing works
- localStorage usage correct
- UI organized and intuitive

**Git Commit Message:**
```
feat: Add advanced settings and test presets

- Create preset save/load functionality
- Implement test result comparison
- Add batch testing capability
- Include config templates
- Persist settings to localStorage

Enhances power-user capabilities.
```

---

### TIER 8: Documentation & Cleanup (1 Microtask)

#### MICROTASK 8.1: Create Documentation & User Guide
**Estimated Time:** 2 hours
**Priority:** Medium (helps adoption)
**Dependencies:** MICROTASK 7.2

**Scope:**
- Create `docs/AGENT_TESTING_UI.md`
- Document:
  - Feature overview
  - User guide (how to test agents)
  - Configuration options explained
  - API reference for developers
  - Troubleshooting guide
  - Performance optimization tips

**Content:**
- Screenshots/GIFs of workflow
- Example test cases
- Config recommendations (temp, tokens)
- Cost estimation
- Metrics explanation
- API integration guide

**Acceptance Criteria:**
- Documentation complete and clear
- Examples work as described
- Screenshots current
- Links functional
- Accessible (no jargon without explanation)

**Git Commit Message:**
```
docs: Add comprehensive agent testing documentation

- Create user guide for agent testing UI
- Document configuration options
- Add API reference for developers
- Include troubleshooting guide
- Provide example test cases

Improves feature discoverability and adoption.
```

---

## Execution Order & Dependencies

### Phase 1: Foundation (Days 1-2)
```
MICROTASK 1.1 (Database Migration)
    ↓
MICROTASK 2.1 (SQLC Queries)
    ↓
MICROTASK 2.2 (Service Layer)
    ↓
MICROTASK 2.3 (HTTP Handler)
```
**Parallel:** All database/backend tasks can overlap
**Duration:** 8-10 hours

### Phase 2: Frontend Components (Days 2-4)
```
MICROTASK 4.1 (API Client) - Can start after 2.3
    ↓
MICROTASK 3.1 (Main Container)
    ↓
MICROTASK 3.2 (Input & Config)
    ↓
MICROTASK 3.3 (Results & Execution)
    ↓
MICROTASK 5.1 (Error Handling)
    ↓
MICROTASK 5.2 (Loading States)
```
**Parallel:** Input and Config (3.2) can be done side-by-side
**Duration:** 14-18 hours

### Phase 3: Testing & Polish (Days 5-6)
```
MICROTASK 6.1 (Integration Tests)
    ↓
MICROTASK 6.2 (Unit Tests)
    ↓
MICROTASK 6.3 (E2E & Performance)
```
**Duration:** 6-9 hours

### Phase 4: Advanced Features & Docs (Days 6-7)
```
MICROTASK 7.1 (History) - Optional
MICROTASK 7.2 (Advanced) - Optional
MICROTASK 8.1 (Documentation)
```
**Duration:** 4-6 hours (optional)

---

## Success Criteria

### Functional Requirements
- Users can select any custom agent from dropdown
- Users can enter test message (any length)
- Users can configure temperature, max tokens, model
- Users can toggle thinking and streaming modes
- Agent executes test and returns response within 10 seconds
- Results display response, duration, tokens, cost
- Error states display clear error messages
- Loading state shows progress during execution

### Performance Requirements
- Component render time <100ms
- API response time <10 seconds
- Page load time <2 seconds
- No memory leaks during test execution
- Support 100+ test history entries
- Metrics calculation instant

### Quality Requirements
- 80%+ test coverage
- All edge cases handled
- Accessibility score >90
- No console errors
- Responsive design (mobile/tablet/desktop)
- Cross-browser compatible

### User Experience
- Intuitive interface
- Clear feedback during execution
- Helpful error messages
- Fast response times
- Visual polish (animations, transitions)
- Copy/export functionality

---

## Risk Assessment

### Technical Risks
1. **Token counting accuracy** - Mitigation: Use LLM provider's count
2. **Long-running tests** - Mitigation: Add timeout, cancellation
3. **Large response sizes** - Mitigation: Truncate in UI, pagination in history
4. **Concurrent test execution** - Mitigation: Queue tests per user
5. **Memory usage** - Mitigation: Cleanup old test runs, pagination

### Time Risks
1. **Backend complexity underestimated** - Mitigation: Start early, incremental
2. **Frontend animation performance** - Mitigation: Use CSS transforms, profile early
3. **Database schema changes** - Mitigation: Test migration thoroughly
4. **Testing coverage gap** - Mitigation: Write tests as you go

---

## Resource Requirements

### Developer Time
- **Backend Engineer:** 8-10 hours (database, service, handler)
- **Frontend Engineer:** 14-18 hours (components, API client, integration)
- **QA Engineer:** 6-9 hours (testing, performance, accessibility)
- **Total:** 28-37 hours (~4-5 developer days)

### Infrastructure
- Database: PostgreSQL (migrations)
- Cache: Redis (optional, for metrics)
- Logging: Structured slog (backend)

### Dependencies
- Existing: CustomAgent model, LLM service, Gin framework
- New: None required (uses existing systems)

---

## Optional Phase 2 Enhancements

### Performance Optimization
- Cache frequently tested agents
- Optimize database queries with materialized views
- Add Redis caching for metrics

### Advanced Features
- Compare multiple test runs side-by-side
- Batch testing (send 10 messages, test all)
- Test templates for common use cases
- Metrics export (CSV, JSON)
- Webhook integration for CI/CD

### Integrations
- GitHub Actions CI integration
- Slack notifications for test results
- DataDog monitoring integration
- Custom metric collection

---

## Monitoring & Observability

### Metrics to Track
- Test execution count (daily/weekly)
- Average duration per test
- Success rate (% passed)
- Cost per test (USD)
- Most tested agents
- Common error types

### Logging
- Test request/response (sanitized)
- Service execution time
- Database query performance
- LLM API calls
- Error stack traces (dev only)

### Alerts
- Test failures spike
- Duration exceeds threshold
- Cost per test exceeds budget
- Database connection pool exhaustion

---

## File Checklist

### Files to Create
```
Desktop/backend-go/
├── internal/database/migrations/037_agent_testing.sql [1.1]
├── internal/handlers/agent_testing.go [2.3]
├── internal/services/agent_testing_service.go [2.2]
├── internal/handlers/agent_testing_test.go [6.1]

frontend/src/
├── lib/api/agent-testing.ts [4.1]
├── lib/components/settings/
│   ├── AgentTestPanel.svelte [3.1]
│   ├── AgentTestInput.svelte [3.2]
│   ├── AgentTestConfig.svelte [3.2]
│   ├── AgentTestExecution.svelte [3.3]
│   ├── AgentTestResults.svelte [3.3]
│   ├── AgentTestMetrics.svelte [3.3]
│   ├── AgentTestError.svelte [5.1]
│   ├── AgentTestSpinner.svelte [5.2]
│   ├── AgentTestHistory.svelte [7.1]
│   ├── AgentTestAdvanced.svelte [7.2]
│   ├── __tests__/
│   │   ├── AgentTestPanel.integration.test.ts [6.1]
│   │   ├── AgentTestInput.test.ts [6.2]
│   │   ├── AgentTestConfig.test.ts [6.2]
│   │   └── AgentTestResults.test.ts [6.2]

docs/
└── AGENT_TESTING_UI.md [8.1]
```

### Files to Modify
```
Desktop/backend-go/
├── cmd/server/main.go (add route)
├── internal/handlers/handlers.go (add handler registration)
├── internal/database/queries/ (add queries SQL file)

frontend/src/
├── lib/components/settings/+page.svelte (include AgentTestPanel)
├── lib/api/ (export testing functions)
```

---

## Next Steps

1. **Start with MICROTASK 1.1** - Database migration
2. **Follow dependency chain** - Don't skip prerequisites
3. **Commit frequently** - One commit per microtask
4. **Test as you go** - Write tests alongside code
5. **Review before merging** - Peer review each microtask
6. **Document as you code** - Add comments and doc strings

---

## Questions & Clarifications

- Should test history be per-user or per-workspace? **Recommend: per-workspace**
- Should tests be concurrent or queued? **Recommend: queue per user to prevent rate limit**
- Should we track cost? **Recommend: yes, for transparency**
- Should we export results? **Recommend: yes, JSON + CSV**
- Should presets be saved? **Recommend: optional phase 2**

