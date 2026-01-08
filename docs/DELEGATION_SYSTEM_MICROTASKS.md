# Delegation System - Atomic Microtasks Breakdown

**Total Effort:** 48-60 hours (6-8 days full-time development)
**Target Timeline:** 2 weeks (with parallel execution)
**Base Architecture:** Existing handlers + services are in place

---

## Overview

The Delegation System allows users to @mention agents in the chat interface and delegate tasks. This document breaks down the implementation into 1-4 hour microtasks, each completable independently.

**Existing Infrastructure:**
- ✅ Backend handlers: `delegation.go`, `delegation service` implemented
- ✅ API endpoints for listing agents, resolving mentions, extracting mentions
- ✅ Database schema for agent mentions and delegation tracking
- ❌ Frontend @mention autocomplete component
- ❌ Delegation panel UI
- ❌ Confirmation flow
- ⚠️ Frontend-backend integration incomplete

---

## TRACK A: @Mention Autocomplete (8-12 hours total)

### Microtask A1: Basic @Mention Dropdown (2-3 hours)
**Time Estimate:** 2-3 hours
**Complexity:** Moderate
**Dependencies:** None
**Agents:** @frontend-svelte, @api-designer

#### Deliverables
1. Create `MentionAutocomplete.svelte` component
2. Detect "@" character in textarea
3. Extract current mention text being typed
4. Show dropdown list below cursor position
5. Support arrow keys (up/down) to navigate
6. Support Enter to select
7. Support Escape to close

#### Technical Details
```
File: frontend/src/lib/components/ai-elements/MentionAutocomplete.svelte
Dependencies: @floating-ui/svelte (position), existing PromptInput component

Features:
- Listen for @ in textarea
- Extract mention prefix: "@co" → "co"
- Cursor position tracking
- Dropdown positioning relative to textarea
- Basic keyboard navigation
- Click to select
- ESC to close
```

#### Acceptance Criteria
- Dropdown appears when typing "@"
- Shows sample agent list (hardcoded for now)
- Keyboard navigation works
- Selection updates textarea correctly
- Dropdown closes on ESC or click outside

---

### Microtask A2: Advanced Filtering & Fuzzy Search (2-3 hours)
**Time Estimate:** 2-3 hours
**Complexity:** Moderate
**Dependencies:** A1 (basic dropdown complete)
**Agents:** @frontend-svelte, @frontend-react (for patterns)

#### Deliverables
1. Integrate fuzzy search library (fuse.js or similar)
2. Filter agents based on typed mention
3. Support partial matching (@code → @code-reviewer, @debugger, etc.)
4. Rank results by relevance
5. Show agent descriptions in dropdown
6. Highlight matched text portions

#### Technical Details
```
Libraries:
- fuse.js for fuzzy matching
- Optional: micromark for agent descriptions

Filtering Logic:
- "@code" matches: @code-reviewer, @codebase-analyzer
- Rank by: exact prefix > substring > fuzzy match
- Order by: usage frequency (from analytics)
```

#### Acceptance Criteria
- Fuzzy matching works for agent names
- Results ranked by relevance
- Agent descriptions visible in dropdown
- Typing "@" narrows results
- Matched portions highlighted in list

---

### Microtask A3: Multi-Mention Support & Conflict Detection (2 hours)
**Time Estimate:** 2 hours
**Complexity:** Simple
**Dependencies:** A1, A2 (basic autocomplete complete)
**Agents:** @frontend-svelte

#### Deliverables
1. Support multiple @mentions in single message
2. Track all mentions separately
3. Warn if same agent mentioned twice
4. Highlight all resolved mentions in message
5. Show preview of all mentioned agents before send

#### Technical Details
```
State Management:
- Array of MentionedAgent[] in component
- Update on each @mention selection
- Show badge below textarea with all mentions

Validation:
- No duplicate mentions
- Valid agent names only
- Max 5 mentions per message (configurable)
```

#### Acceptance Criteria
- Multiple @mentions in one message work
- Each mention can be selected independently
- Preview shows all @mentions before send
- Duplicate detection works
- Invalid mentions marked with warning

---

### Microtask A4: Dynamic Agent List From API (2-3 hours)
**Time Estimate:** 2-3 hours
**Complexity:** Simple
**Dependencies:** A1-A3 (autocomplete complete), existing API endpoints
**Agents:** @frontend-svelte, @api-designer

#### Deliverables
1. Call `/api/agents/available` to get list
2. Cache agent list in store/context
3. Refresh cache on component mount
4. Handle API errors gracefully
5. Show loading state while fetching

#### Technical Details
```
API Endpoint: GET /api/agents/available
Response:
{
  "agents": [
    {
      "id": "uuid",
      "name": "code-reviewer",
      "display_name": "Code Reviewer",
      "description": "Code quality review",
      "capabilities": ["review", "quality"],
      "category": "quality"
    }
  ],
  "count": 15
}

Frontend Store:
- Create `agentStore` (Svelte store)
- Cache with 5-minute TTL
- Background refresh
```

#### Acceptance Criteria
- Agents loaded from API on mount
- Dropdown shows all available agents
- Agent filtering works with API data
- Loading state shown while fetching
- Error state handled gracefully

---

## TRACK B: Mention Resolution Service (6-8 hours total)

### Microtask B1: Basic Mention Resolution Endpoint (2 hours)
**Time Estimate:** 2 hours
**Complexity:** Simple
**Dependencies:** Existing delegation handler
**Agents:** @backend-go

#### Deliverables
1. Verify `/api/agents/resolve/:mention` endpoint works
2. Add comprehensive error handling
3. Return full agent object with all metadata
4. Log resolution attempts for analytics
5. Handle case-insensitivity properly

#### Technical Details
```
Endpoint: GET /api/agents/resolve/:mention
Input: mention = "code-reviewer" or "@code-reviewer"
Output:
{
  "agent": DelegationTarget,
  "mention": "code-reviewer",
  "resolved": true,
  "timestamp": "2026-01-08T10:00:00Z"
}

Error Cases:
- Agent not found (404)
- Invalid mention format (400)
- Database error (500)
```

#### Acceptance Criteria
- Endpoint returns agent data correctly
- Handles both "@mention" and "mention" formats
- Returns 404 for unknown agents
- Proper error messages in responses
- All fields populated in response

---

### Microtask B2: Batch Mention Resolution (2 hours)
**Time Estimate:** 2 hours
**Complexity:** Moderate
**Dependencies:** B1 (single resolution works)
**Agents:** @backend-go, @database-specialist

#### Deliverables
1. Create new endpoint for batch resolution: `POST /api/agents/resolve-batch`
2. Resolve multiple mentions in single request
3. Return array of resolved agents
4. Handle partial failures gracefully
5. Optimize with single database query

#### Technical Details
```
Endpoint: POST /api/agents/resolve-batch
Request:
{
  "mentions": ["@code-reviewer", "@debugger", "@frontend-svelte"],
  "user_id": "user-123"
}

Response:
{
  "resolved": [
    { agent object },
    { agent object }
  ],
  "unresolved": ["@unknown-agent"],
  "count": 2,
  "errors": []
}

Optimization:
- Single query with IN clause
- Return all agents in 1 query, not N
```

#### Acceptance Criteria
- Batch endpoint created and working
- Multiple mentions resolved in one request
- Partial failures handled (some resolve, some don't)
- Faster than calling single endpoint multiple times
- Error details returned for unresolved

---

### Microtask B3: Mention Extraction & Recording (2 hours)
**Time Estimate:** 2 hours
**Complexity:** Simple
**Dependencies:** B1, B2 (resolution working)
**Agents:** @backend-go

#### Deliverables
1. Enhance `/api/agents/mentions` POST endpoint
2. Parse message for @mentions using regex
3. Resolve all mentions found
4. Store mention records in database
5. Return structured mention data with metadata

#### Technical Details
```
Endpoint: POST /api/agents/mentions
Request:
{
  "message": "Hey @code-reviewer and @debugger can you check this?"
}

Response:
{
  "mentions": [
    {
      "mention": "@code-reviewer",
      "agent": { ...agent object },
      "position": 4,
      "resolved": true
    },
    {
      "mention": "@debugger",
      "agent": { ...agent object },
      "position": 23,
      "resolved": true
    }
  ],
  "count": 2
}

Storage:
- Insert into agent_mentions table
- user_id, conversation_id, message_id, agent_id, mention_text, position
```

#### Acceptance Criteria
- Mentions extracted from message correctly
- All @mentions found (including in punctuation)
- Mention positions accurate
- Records stored in database
- Return format matches frontend expectations

---

### Microtask B4: Mention Resolution Caching (2 hours)
**Time Estimate:** 2 hours
**Complexity:** Moderate
**Dependencies:** All B tasks (resolution working)
**Agents:** @backend-go, @performance-optimizer

#### Deliverables
1. Add Redis caching for agent list
2. Cache with 1-hour TTL
3. Implement cache invalidation
4. Monitor cache hit rates
5. Add cache warming on startup

#### Technical Details
```
Redis Keys:
- agents:all:list (agent list, 1h TTL)
- agents:by-name:{name} (specific agent, 1h TTL)
- agents:user:{user-id}:custom (user agents, 1h TTL)

Cache Strategy:
- Write-through: update cache after DB changes
- Invalidate on: agent creation, update, deletion
- Warm on startup with core agents

Metrics:
- Cache hit rate
- Cache miss count
- Average resolution time (cached vs uncached)
```

#### Acceptance Criteria
- Agent list cached in Redis
- Cache hit rate > 90% for repeated queries
- Cache properly invalidated on changes
- Resolution times faster with cache
- Metrics tracked and logged

---

## TRACK C: Delegation Panel UI (8-10 hours total)

### Microtask C1: Basic Delegation Panel Component (2 hours)
**Time Estimate:** 2 hours
**Complexity:** Simple
**Dependencies:** None (new component)
**Agents:** @frontend-svelte, @ui-ux-designer

#### Deliverables
1. Create `DelegationPanel.svelte` component
2. Accept agent object as prop
3. Display agent name, title, description
4. Show agent capabilities as badges
5. Show category/type of agent
6. Clean, professional styling using existing design system

#### Technical Details
```
File: frontend/src/lib/components/delegation/DelegationPanel.svelte

Props:
- agent: DelegationTarget
- onDelegate?: (agent) => void
- onCancel?: () => void

Display:
- Agent avatar/icon
- Display name (prominent)
- Description (2-3 lines max)
- Capabilities (as badges)
- Category tag
- Status indicator (online/offline)
- "Delegate" and "Cancel" buttons
```

#### Acceptance Criteria
- Component displays agent info correctly
- Styling consistent with BusinessOS design
- Responsive on mobile/tablet
- Buttons functional (callback props)
- All agent data fields displayed

---

### Microtask C2: Agent Details Section with Model/Prompt Info (2 hours)
**Time Estimate:** 2 hours
**Complexity:** Simple
**Dependencies:** C1 (panel component exists)
**Agents:** @frontend-svelte

#### Deliverables
1. Expand delegation panel with details section
2. Show agent model preference/override
3. Display system prompt preview (collapsed)
4. Show last activity timestamp
5. Display success rate / reliability metrics
6. Add expandable/collapsible sections

#### Technical Details
```
Additional Sections:
- System Prompt Preview (collapsible)
- Model Info: GPT-4, Claude 3, etc.
- Performance Metrics:
  - Task success rate
  - Avg response time
  - User rating
  - Last used: 2 hours ago

Styling:
- Use disclosure triangle for collapse/expand
- Monospace font for prompts
- Metrics in small cards
```

#### Acceptance Criteria
- System prompt viewable in collapsed state
- Model info displayed
- Metrics shown (if available)
- Expand/collapse works smoothly
- All details read from agent object

---

### Microtask C3: Delegation Reason/Context Input (2 hours)
**Time Estimate:** 2 hours
**Complexity:** Simple
**Dependencies:** C1 (panel component exists)
**Agents:** @frontend-svelte

#### Deliverables
1. Add text area for delegation reason
2. Add task context input (optional)
3. Show character count and limits
4. Suggest delegation reasons (dropdown hints)
5. Validate input before allowing delegation

#### Technical Details
```
Form Fields:
1. Reason (required, max 500 chars)
   - Placeholder: "Why are you delegating this task?"
   - Hints: "Code review needed", "Debugging required", etc.

2. Context (optional, max 1000 chars)
   - Placeholder: "Any additional context for the agent?"
   - Auto-fill from current conversation if available

3. Priority (dropdown)
   - Low, Normal, High, Critical
   - Default: Normal

Validation:
- Reason must be provided
- Reason > 10 characters
- Max character limits enforced
- No special characters in reason
```

#### Acceptance Criteria
- Text inputs accept user input
- Character limits enforced
- Reason validation works
- Hints/suggestions appear
- Data bound to component state

---

### Microtask C4: Agent Comparison View (Optional) (2 hours)
**Time Estimate:** 2 hours
**Complexity:** Moderate
**Dependencies:** C1-C3 (panel exists)
**Agents:** @frontend-svelte

#### Deliverables
1. Allow selection of multiple agents for comparison
2. Side-by-side agent comparison view
3. Highlight differences in capabilities/model
4. Show pros/cons for each agent
5. Recommend best agent for task

#### Technical Details
```
Comparison Features:
- Select up to 3 agents
- Side-by-side cards
- Color-coded differences
- Highlight best match for task type
- "Delegate to Best" button

Comparison Fields:
- Model version
- Capabilities match
- Success rate
- Response time
- Cost estimate
```

#### Acceptance Criteria
- Multiple agent selection works
- Comparison view displays correctly
- Visual comparison clear
- Recommendation logic works
- Can delegate from comparison view

---

## TRACK D: Delegation Confirmation Flow (6-8 hours total)

### Microtask D1: Basic Confirmation Modal (2 hours)
**Time Estimate:** 2 hours
**Complexity:** Simple
**Dependencies:** C1-C3 (delegation panel complete)
**Agents:** @frontend-svelte

#### Deliverables
1. Create `DelegationConfirmationModal.svelte`
2. Show agent being delegated to
3. Show reason and context
4. Display "Confirm" and "Cancel" buttons
5. Handle modal lifecycle (open/close/submit)

#### Technical Details
```
File: frontend/src/lib/components/delegation/DelegationConfirmationModal.svelte

Content:
- Header: "Delegate Task to [Agent Name]?"
- Summary: agent name, reason (truncated)
- Full details: clickable to expand
- Buttons: Cancel | Delegate

Behavior:
- Modal shows when delegation initiated
- Focus trap on modal
- ESC closes modal
- Enter confirms delegation
- onConfirm() callback on submit
```

#### Acceptance Criteria
- Modal appears on command
- Content displayed clearly
- Buttons functional
- Modal dismisses correctly
- Callbacks executed properly

---

### Microtask D2: Detailed Preview & Confirmation Checklist (2 hours)
**Time Estimate:** 2 hours
**Complexity:** Moderate
**Dependencies:** D1 (modal exists)
**Agents:** @frontend-svelte

#### Deliverables
1. Expand modal with full delegation details
2. Show checklist of what will happen
3. Display estimated cost/tokens
4. Show delegation history for this agent
5. Add "Don't show again" option for this agent

#### Technical Details
```
Preview Sections:
1. Delegation Summary
   - From: Current agent/user
   - To: Selected agent
   - Task: Reason/context

2. What Happens Next
   - Checklist of steps:
     - Agent will receive task in queue
     - You'll get notification when started
     - Results will appear here
     - May take X minutes

3. Agent Stats for Context
   - Success rate: 94%
   - Avg completion: 3 min
   - Estimated cost: $0.05
   - Previous delegations: 12

4. Task Breakdown (if available)
   - Subtasks agent will handle
   - Expected outputs
```

#### Acceptance Criteria
- All preview details shown
- Checklist helpful and clear
- Estimated cost displayed
- History relevant
- "Don't show again" works

---

### Microtask D3: Loading State & Progress Feedback (2 hours)
**Time Estimate:** 2 hours
**Complexity:** Moderate
**Dependencies:** D1-D2 (confirmation complete)
**Agents:** @frontend-svelte

#### Deliverables
1. Show loading state during delegation submission
2. Disable buttons while submitting
3. Show progress with spinner/animation
4. Display status messages
5. Handle timeout scenarios gracefully

#### Technical Details
```
States:
1. Confirming State
   - Button text: "Confirming..."
   - Spinner animation
   - Buttons disabled

2. Sending State
   - "Sending to [Agent]..."
   - Progress: uploading context
   - Cancel option available

3. Success State
   - "Task delegated successfully"
   - Show delegation ID
   - "View Details" link
   - Auto-close after 3 seconds

4. Error State
   - Show error message
   - "Retry" and "Cancel" buttons
   - Error details (optional expand)
```

#### Acceptance Criteria
- Loading states clear and visible
- Buttons disabled during submission
- Progress feedback shown
- Error handling works
- Success state clear

---

### Microtask D4: Audit Trail & Delegation History (2 hours)
**Time Estimate:** 2 hours
**Complexity:** Moderate
**Dependencies:** D1-D3 (confirmation complete)
**Agents:** @frontend-svelte, @database-specialist

#### Deliverables
1. Store delegation records in database
2. Show delegation history in panel
3. Display when task was delegated
4. Show results/outcome
5. Allow re-delegation of similar tasks
6. Track success/failure outcomes

#### Technical Details
```
Database: agent_delegations table
- id (UUID)
- user_id
- from_agent_id (if applicable)
- to_agent_id
- conversation_id
- message_id
- reason
- context
- status (pending, completed, failed)
- created_at
- completed_at
- result (text)
- cost (tokens used)

UI:
- "Recent Delegations" in panel
- Timeline of delegations
- Status badges
- Results preview
- "Re-delegate" button
```

#### Acceptance Criteria
- Delegations stored in database
- History shows in UI
- Outcomes visible
- Status tracking works
- Re-delegation functional

---

## TRACK E: Backend Integration (8-10 hours total)

### Microtask E1: Frontend-to-Backend Delegation Endpoint (2 hours)
**Time Estimate:** 2 hours
**Complexity:** Simple
**Dependencies:** D3 (confirmation flow works)
**Agents:** @frontend-svelte, @api-designer

#### Deliverables
1. Call `POST /api/agents/delegate` from frontend
2. Send delegation request with all details
3. Handle response and show result
4. Extract delegation ID from response
5. Error handling for all scenarios

#### Technical Details
```
Endpoint: POST /api/agents/delegate
Request:
{
  "from_agent": "current_agent_or_user",
  "to_agent": "code-reviewer",
  "reason": "Need code quality review",
  "context": "Full conversation context",
  "original_query": "Original user message",
  "metadata": {
    "task_id": "task-123",
    "project_id": "proj-456"
  }
}

Response (Success):
{
  "success": true,
  "target_agent": { DelegationTarget },
  "delegation_id": "deleg-789",
  "trace_id": "trace-xxx",
  "created_at": "2026-01-08T10:00:00Z"
}

Response (Error):
{
  "success": false,
  "error": "Agent not found",
  "trace_id": "trace-xxx"
}

Frontend:
- Show delegation_id to user
- Link to delegation tracking
- Handle errors gracefully
```

#### Acceptance Criteria
- Frontend successfully calls endpoint
- Request includes all required fields
- Response parsed correctly
- Error cases handled
- User feedback provided

---

### Microtask E2: Store Delegation Context & Metadata (2 hours)
**Time Estimate:** 2 hours
**Complexity:** Moderate
**Dependencies:** E1 (delegation endpoint works)
**Agents:** @backend-go, @database-specialist

#### Deliverables
1. Store full conversation context with delegation
2. Save user metadata and task info
3. Index for fast retrieval
4. Create delegation_context table
5. Handle large context payloads efficiently

#### Technical Details
```
New Table: delegation_context
- id (UUID)
- delegation_id (FK)
- conversation_id (UUID)
- message_id (UUID)
- user_id
- context (JSONB - conversation history)
- metadata (JSONB - tags, task info)
- created_at

Index:
- delegation_id (primary)
- user_id, created_at (for queries)
- conversation_id (for retrieval)

Context Storage:
- Include last 10 messages
- Include task details
- Include workspace/project info
- Compress if > 10KB
```

#### Acceptance Criteria
- Context stored on delegation
- Metadata indexed properly
- Retrieval fast (< 100ms)
- No data loss
- Privacy respected (no sensitive data)

---

### Microtask E3: Delegation Status Tracking & Updates (2 hours)
**Time Estimate:** 2 hours
**Complexity:** Moderate
**Dependencies:** E1-E2 (delegation storage works)
**Agents:** @backend-go

#### Deliverables
1. Implement delegation status lifecycle (pending → processing → completed/failed)
2. Create endpoint to get delegation status
3. Support status webhooks/callbacks
4. Store completion time and results
5. Track agent response metadata

#### Technical Details
```
Status Workflow:
pending → processing → completed
       → failed

Endpoint: GET /api/delegations/{delegation_id}
Response:
{
  "id": "deleg-123",
  "status": "processing",
  "created_at": "2026-01-08T10:00:00Z",
  "started_at": "2026-01-08T10:01:00Z",
  "completed_at": null,
  "result": null,
  "error": null,
  "agent_id": "uuid",
  "tokens_used": 0,
  "progress": {
    "percent": 45,
    "message": "Analyzing code..."
  }
}

Updates:
- Record when agent starts
- Update progress periodically
- Store result when complete
- Record any errors
- Calculate tokens/cost
```

#### Acceptance Criteria
- Status lifecycle implemented
- Status endpoint working
- Progress tracking functional
- Results stored properly
- Error tracking complete

---

### Microtask E4: Delegation Analytics & Metrics (2 hours)
**Time Estimate:** 2 hours
**Complexity:** Moderate
**Dependencies:** E1-E3 (delegation tracking works)
**Agents:** @backend-go, @database-specialist

#### Deliverables
1. Track delegation frequency by agent
2. Calculate success rates per agent
3. Measure avg completion time
4. Track cost/tokens per delegation
5. Create analytics dashboard endpoints

#### Technical Details
```
Metrics to Track:
- Total delegations per agent
- Success rate (%)
- Avg completion time
- Avg tokens used
- Cost per delegation
- User satisfaction rating
- Most used agents (by user, workspace)

Endpoints:
GET /api/analytics/delegations
GET /api/analytics/agents/{agent_id}/stats
GET /api/analytics/delegations/trends

Cache:
- Aggregate metrics (1h cache)
- User-specific metrics (10m cache)
- Real-time current metrics

Logging:
- Every delegation logged
- Metrics updated after completion
- No PII in metrics
```

#### Acceptance Criteria
- Metrics accurately tracked
- Endpoints return expected data
- Analytics queryable
- Performance efficient
- Data privacy maintained

---

### Microtask E5: Delegation Webhooks & Real-time Updates (2 hours)
**Time Estimate:** 2 hours
**Complexity:** Moderate
**Dependencies:** E1-E3 (delegation system working)
**Agents:** @backend-go, @api-designer

#### Deliverables
1. Implement webhook system for delegation events
2. Support multiple event types (started, completed, failed, etc.)
3. Allow users to register webhooks
4. Implement retry logic for failed webhooks
5. Add webhook testing/validation

#### Technical Details
```
Events:
- delegation.created
- delegation.started
- delegation.processing
- delegation.completed
- delegation.failed
- delegation.cancelled

Webhook Registration:
POST /api/webhooks
{
  "url": "https://example.com/webhooks",
  "events": ["delegation.completed", "delegation.failed"],
  "user_id": "user-123"
}

Webhook Payload:
{
  "event": "delegation.completed",
  "delegation_id": "deleg-123",
  "agent_id": "agent-456",
  "status": "completed",
  "result": {...},
  "timestamp": "2026-01-08T10:30:00Z"
}

Retry Logic:
- Exponential backoff (1s, 2s, 4s, 8s, 16s)
- Max 5 retries
- Store webhook delivery history
```

#### Acceptance Criteria
- Webhooks fire on correct events
- Payloads correctly formatted
- Retry logic working
- Webhook history stored
- Testing endpoint functional

---

## TRACK F: Error States & Resilience (6-8 hours total)

### Microtask F1: Network Error Handling - Delegation (2 hours)
**Time Estimate:** 2 hours
**Complexity:** Simple
**Dependencies:** E1 (delegation endpoint exists)
**Agents:** @frontend-svelte, @debugger

#### Deliverables
1. Handle network timeouts (> 10 seconds)
2. Handle connection failures
3. Show user-friendly error messages
4. Implement retry logic with backoff
5. Store queued delegations locally (optional)

#### Technical Details
```
Error Scenarios:
1. Network Timeout
   - Show: "Request timed out, retrying..."
   - Retry after 2s, 5s, 10s
   - Max 3 retries
   - Allow manual retry

2. Connection Failed
   - Show: "No internet connection"
   - Detect when connection restored
   - Auto-retry delegation
   - Save to local queue

3. Server Error (5xx)
   - Show: "Server error, please try again"
   - Retry with backoff
   - Contact support link
   - Preserve delegation data

Error UI:
- Clear error message
- Suggested action (retry/cancel)
- Error details (optional expand)
- Contact support link
```

#### Acceptance Criteria
- All network errors caught
- User messages clear
- Retry logic functional
- Backoff implemented
- Local queuing works

---

### Microtask F2: Validation Error Handling (2 hours)
**Time Estimate:** 2 hours
**Complexity:** Simple
**Dependencies:** C3, D2 (input validation exists)
**Agents:** @frontend-svelte

#### Deliverables
1. Validate all inputs before sending
2. Show validation errors in form
3. Prevent submission of invalid data
4. Provide helpful error messages
5. Support field-level error display

#### Technical Details
```
Validation Rules:
- Agent must be selected
- Reason must be 10+ characters
- Reason must not exceed 500 chars
- Priority must be valid enum
- Context must not exceed 1000 chars

Error Display:
- Inline validation messages
- Form submit disabled if invalid
- Field highlight (red border)
- Helper text with rules
- Clear what's wrong

Example:
[Text input] "Add async to React component"
X Error: Reason too short (min 10 characters)
  Remaining: 2 more characters needed

Validation Types:
- Required fields
- Min/max length
- Format validation (email, URL, etc.)
- Enum validation (priority)
```

#### Acceptance Criteria
- All inputs validated
- Error messages helpful
- Form submission prevented on error
- Field-level errors clear
- Validation messages update live

---

### Microtask F3: Timeout & Rate Limit Handling (2 hours)
**Time Estimate:** 2 hours
**Complexity:** Moderate
**Dependencies:** E1-E3 (delegation system working)
**Agents:** @backend-go, @frontend-svelte

#### Deliverables
1. Implement rate limiting on delegation endpoint
2. Return 429 status when exceeded
3. Show user rate limit info
4. Implement client-side rate limit detection
5. Provide queue position feedback

#### Technical Details
```
Rate Limiting:
- 10 delegations per minute per user
- 100 per hour per user
- 1000 per day per user
- Return: X-RateLimit-Remaining header

Client Handling:
- Detect 429 response
- Show: "Rate limit reached"
- "You have 3 delegations left this minute"
- Show queue position if available
- Suggest queuing for later

Timeout Handling:
- Request timeout: 30 seconds
- Show progress after 5 seconds
- Cancel button available
- Suggest contact support after 60s

Headers to Return:
- X-RateLimit-Limit: 10
- X-RateLimit-Remaining: 5
- X-RateLimit-Reset: 2026-01-08T10:05:00Z
```

#### Acceptance Criteria
- Rate limiting enforced on backend
- 429 errors handled gracefully
- Client detects rate limits
- User messaging clear
- Timeouts handled properly

---

### Microtask F4: Graceful Degradation & Fallbacks (2 hours)
**Time Estimate:** 2 hours
**Complexity:** Moderate
**Dependencies:** All error handling tasks
**Agents:** @frontend-svelte, @backend-go

#### Deliverables
1. Allow delegation without full agent data
2. Fallback to agent ID if name not available
3. Support offline delegation queueing
4. Gracefully handle partial API failures
5. Provide manual delegation option

#### Technical Details
```
Graceful Degradation:
1. Agent List Unavailable
   - Allow typing agent name manually
   - Show: "Enter agent name"
   - Warning: "Unable to verify agent"
   - Allow user to proceed anyway

2. Agent Details Unavailable
   - Show available info only
   - Hide missing sections gracefully
   - Use placeholder for images
   - Show: "Some details unavailable"

3. Metrics Unavailable
   - Don't show metrics section
   - Don't show "Success Rate"
   - Show only basic info

4. Offline Delegation
   - Queue delegation locally (IndexedDB)
   - Show: "Queued (offline)"
   - Send when connection restored
   - Auto-retry with notifications

Fallback Order:
- Use cached data if available
- Show simplified UI
- Allow manual entry
- Queue locally
- Notify when restored
```

#### Acceptance Criteria
- App works without agent list
- Manual entry possible
- Offline queuing functional
- Graceful UI degradation
- All fallbacks tested

---

## Summary Table

| Track | Microtask | Hours | Complexity | Status |
|-------|-----------|-------|-----------|--------|
| A | A1: Basic dropdown | 2-3 | Moderate | Pending |
| A | A2: Fuzzy search | 2-3 | Moderate | Pending |
| A | A3: Multi-mention | 2 | Simple | Pending |
| A | A4: API integration | 2-3 | Simple | Pending |
| B | B1: Single resolution | 2 | Simple | Pending |
| B | B2: Batch resolution | 2 | Moderate | Pending |
| B | B3: Mention extraction | 2 | Simple | Pending |
| B | B4: Resolution caching | 2 | Moderate | Pending |
| C | C1: Basic panel | 2 | Simple | Pending |
| C | C2: Agent details | 2 | Simple | Pending |
| C | C3: Reason input | 2 | Simple | Pending |
| C | C4: Agent comparison | 2 | Moderate | Pending |
| D | D1: Confirmation modal | 2 | Simple | Pending |
| D | D2: Preview & checklist | 2 | Moderate | Pending |
| D | D3: Loading state | 2 | Moderate | Pending |
| D | D4: History & audit | 2 | Moderate | Pending |
| E | E1: Frontend-backend call | 2 | Simple | Pending |
| E | E2: Store context | 2 | Moderate | Pending |
| E | E3: Status tracking | 2 | Moderate | Pending |
| E | E4: Analytics | 2 | Moderate | Pending |
| E | E5: Webhooks | 2 | Moderate | Pending |
| F | F1: Network errors | 2 | Simple | Pending |
| F | F2: Validation errors | 2 | Simple | Pending |
| F | F3: Rate limits/timeouts | 2 | Moderate | Pending |
| F | F4: Graceful degradation | 2 | Moderate | Pending |

**Total: 52-60 hours | 24 microtasks**

---

## Execution Strategy

### Phase 1: Parallel Tracks (Weeks 1-2)
**Execute in parallel:** A (autocomplete), B (resolution), C (panel), D (confirmation)
- Week 1: A1-A2, B1-B2, C1-C2, D1-D2
- Week 2: A3-A4, B3-B4, C3-C4, D3-D4

### Phase 2: Integration (Week 2-3)
**Execute sequentially:** E (backend integration), F (error handling)
- E1-E2: Frontend-backend wiring
- E3-E5: Status, analytics, webhooks
- F1-F4: Error handling in parallel

### Phase 3: Testing & Polish (Week 3)
- Integration testing
- End-to-end testing
- Performance optimization
- Bug fixes and refinement

---

## Dependencies Graph

```
A1 (Basic Dropdown)
├─ A2 (Fuzzy Search)
├─ A3 (Multi-Mention)
└─ A4 (API Integration) → B1

B1 (Single Resolution)
├─ B2 (Batch Resolution)
├─ B3 (Mention Extraction)
└─ B4 (Caching)

C1 (Basic Panel)
├─ C2 (Agent Details)
├─ C3 (Reason Input)
└─ C4 (Comparison)

D1 (Confirmation Modal)
├─ D2 (Preview)
├─ D3 (Loading State)
└─ D4 (History)

E1 (Frontend Call) → requires D3
├─ E2 (Store Context)
├─ E3 (Status Tracking)
├─ E4 (Analytics)
└─ E5 (Webhooks)

F1, F2, F3, F4 (Error Handling) → requires all above
```

---

## Success Criteria

For each microtask to be "complete":
1. Functional: Works as designed
2. Tested: Unit tests written (if applicable)
3. Documented: Code comments, type definitions
4. Integrated: Works with existing codebase
5. Verified: Manual testing completed
6. No regressions: Existing tests still pass

---

## Notes

- Each microtask is **completely independent** and can be worked on in parallel
- Time estimates include development + testing
- Complexity levels help prioritize: Finish "Simple" tasks first for quick wins
- Dependencies marked clearly - respect them to avoid rework
- Backend services already mostly built - frontend integration is the gap
- Use existing design patterns from BusinessOS codebase
- Test thoroughly - delegation is a critical system

---

**Document Version:** 1.0
**Created:** 2026-01-08
**Last Updated:** 2026-01-08
**Owner:** Architecture Team
**Status:** Ready for Implementation
