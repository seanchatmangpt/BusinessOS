# Delegation System - Microtasks Quick Reference

**Total Effort:** 52-60 hours | **24 microtasks** | **2-3 weeks execution time**

---

## All Microtasks at a Glance

### TRACK A: @Mention Autocomplete (8-12 hours)
1. **A1** (2-3h) - Basic dropdown that shows when typing "@"
2. **A2** (2-3h) - Fuzzy search and filtering by agent name
3. **A3** (2h) - Support multiple mentions, prevent duplicates
4. **A4** (2-3h) - Load agent list from API endpoint

### TRACK B: Mention Resolution Service (6-8 hours)
5. **B1** (2h) - Verify single mention resolution endpoint works
6. **B2** (2h) - Create batch resolution endpoint for multiple mentions
7. **B3** (2h) - Extract and record mentions with database storage
8. **B4** (2h) - Add Redis caching for agent lookups

### TRACK C: Delegation Panel UI (8-10 hours)
9. **C1** (2h) - Basic panel displaying agent info and buttons
10. **C2** (2h) - Add agent details: model, prompt preview, metrics
11. **C3** (2h) - Add reason input field and context form
12. **C4** (2h) - Optional: Side-by-side agent comparison view

### TRACK D: Delegation Confirmation Flow (6-8 hours)
13. **D1** (2h) - Basic confirmation modal with confirm/cancel
14. **D2** (2h) - Expanded preview with checklist and details
15. **D3** (2h) - Loading states and progress feedback
16. **D4** (2h) - Store delegation history and track outcomes

### TRACK E: Backend Integration (8-10 hours)
17. **E1** (2h) - Frontend calls POST /api/agents/delegate endpoint
18. **E2** (2h) - Store full conversation context with delegation
19. **E3** (2h) - Track delegation status (pending → processing → done)
20. **E4** (2h) - Add analytics: success rates, completion time, costs
21. **E5** (2h) - Implement webhooks for real-time delegation updates

### TRACK F: Error States & Resilience (6-8 hours)
22. **F1** (2h) - Handle network timeouts and connection failures
23. **F2** (2h) - Validate inputs, show field-level errors
24. **F3** (2h) - Implement rate limiting and timeout handling
25. **F4** (2h) - Graceful degradation when services unavailable

---

## Time Breakdown by Complexity

| Complexity | Count | Hours | Examples |
|-----------|-------|-------|----------|
| **Simple** | 9 | 16-18h | A3, A4, B1, B3, C1, C2, C3, D1, E1, F1, F2 |
| **Moderate** | 12 | 24-26h | A1, A2, B2, B4, C4, D2, D3, E2-E5, F3, F4 |
| **Complex** | 3 | 8-10h | C4, E5, F4 |

---

## Recommended Execution Order

### Week 1 (Sprint A - Parallel)
Execute in parallel: A, B, C, D early tasks

```
Day 1-2:  A1 + B1 + C1 + D1  (4 basic tasks)
Day 2-3:  A2 + B2 + C2 + D2  (4 moderate tasks)
Day 3-4:  A3 + B3 + C3       (3 tasks)
Day 4-5:  A4 + B4 + C4 + D3  (4 tasks)
```

### Week 2 (Sprint B - Integration Setup)
Start E track, finish D track

```
Day 1:    D4 + E1           (2 tasks - complete D, start E)
Day 2-3:  E2 + E3           (2 tasks - backend context & tracking)
Day 3-4:  E4 + E5           (2 tasks - analytics & webhooks)
Day 5:    Testing & fixes
```

### Week 3 (Sprint C - Error Handling & Polish)
Complete F track and integration

```
Day 1-2:  F1 + F2 + F3      (3 error handling tasks)
Day 2-3:  F4 + integration tests
Day 3-5:  Bug fixes, optimization, final testing
```

---

## Dependency Tree (Simplified)

```
START
├─ A1 (Basic dropdown)
│  ├─ A2 (Fuzzy search)
│  ├─ A3 (Multi-mention)
│  └─ A4 (API loading) ─┐
│                        │
├─ B1 (Single resolution)├─ B2 (Batch)
│                        │  ├─ B3 (Extraction)
│                        │  └─ B4 (Caching)
│                        │
├─ C1 (Panel) ─ C2 (Details) ─ C3 (Reason input) ─ C4 (Comparison)
│
├─ D1 (Modal) ─ D2 (Preview) ─ D3 (Loading) ─┐
│                                            │
└─ E1 (Frontend call) ◄─────────────────────┘
   ├─ E2 (Store context)
   ├─ E3 (Status tracking)
   ├─ E4 (Analytics)
   └─ E5 (Webhooks)

F1, F2, F3, F4 (Error handling - can happen anytime)
```

---

## Critical Path (Minimum)

**Fastest route to working delegation (40 hours):**

1. A1 (2h) - Basic autocomplete
2. B1 (2h) - Single resolution
3. C1 (2h) - Basic panel
4. D1 (2h) - Confirmation modal
5. E1 (2h) - Frontend-backend call
6. F1 (2h) - Network error handling
7. A2 (2h) - Fuzzy search
8. C3 (2h) - Reason input
9. D3 (2h) - Loading states
10. E3 (2h) - Status tracking
11. F2 (2h) - Validation
12. A4 (2h) - API integration
13. B3 (2h) - Mention extraction
14. D2 (2h) - Preview
15. E2 (2h) - Store context
16. F3 (2h) - Rate limits
17. Rest...

---

## Key Metrics

| Metric | Value |
|--------|-------|
| **Total Microtasks** | 24 |
| **Total Hours** | 52-60 |
| **Avg Hours/Task** | 2.2 |
| **Min Hours/Task** | 2 |
| **Max Hours/Task** | 3 |
| **Parallel Capacity** | 4-6 tasks simultaneously |
| **Timeline (sequential)** | 52-60 hours = ~7-8 days |
| **Timeline (parallel)** | ~15-18 days = 2-3 weeks |

---

## Prerequisite Knowledge

Before starting, ensure familiarity with:

**Frontend (Svelte):**
- Svelte stores
- Component props and state
- Event handling and callbacks
- Floating UI positioning
- Modal patterns

**Backend (Go):**
- HTTP handlers with Gin
- Database queries (PostgreSQL)
- Error handling patterns
- Logging with slog
- Redis integration

**Database:**
- SQL queries and indexing
- JSONB data type
- UUID primary keys
- Foreign key relationships

---

## Team Assignment Recommendations

### If 3 developers:
- **Dev 1:** Tracks A, B (Frontend + Mention Resolution)
- **Dev 2:** Tracks C, D (Delegation UI + Confirmation)
- **Dev 3:** Tracks E, F (Backend + Error Handling)

### If 2 developers:
- **Dev 1:** Tracks A, C, E (All Frontend + Basic Backend Integration)
- **Dev 2:** Tracks B, D, F (Mention Resolution + Confirmation + Error Handling)

### If 1 developer:
- Execute in order: A → B → C → D → E → F

---

## Definition of Done (per microtask)

Each microtask is complete when:

- [ ] **Code written** - All functional code complete
- [ ] **Tested** - Unit tests written (if applicable)
- [ ] **No TypeScript errors** - Strict mode clean
- [ ] **Integrated** - Works with existing codebase
- [ ] **Documented** - Comments on complex logic
- [ ] **Verified manually** - Tested in browser/app
- [ ] **No regressions** - Existing tests still pass
- [ ] **PR ready** - Code review approved

---

## Cost Estimate (if outsourcing)

| Phase | Hours | Cost (@ $50/hr) | Cost (@ $100/hr) |
|-------|-------|-----------------|-----------------|
| A: Autocomplete | 8-12 | $400-600 | $800-1200 |
| B: Resolution | 6-8 | $300-400 | $600-800 |
| C: Panel UI | 8-10 | $400-500 | $800-1000 |
| D: Confirmation | 6-8 | $300-400 | $600-800 |
| E: Integration | 8-10 | $400-500 | $800-1000 |
| F: Error Handling | 6-8 | $300-400 | $600-800 |
| **Total** | **52-60** | **$2,100-2,800** | **$4,200-5,600** |

---

## Quality Checklist

Before marking the entire system "complete":

- [ ] All 24 microtasks completed and tested
- [ ] No TypeScript errors or warnings
- [ ] All user flows working end-to-end
- [ ] Network error handling tested
- [ ] Rate limiting enforced
- [ ] Offline fallback working
- [ ] Mobile responsive
- [ ] Accessibility (keyboard nav, screen readers)
- [ ] Performance acceptable (< 2s delegation time)
- [ ] Security reviewed (no XSS, SQL injection)
- [ ] Documentation complete
- [ ] Code review approved

---

## Links & References

**Backend Code:**
- `/desktop/backend-go/internal/handlers/delegation.go` - Handler implementation
- `/desktop/backend-go/internal/services/delegation.go` - Service layer

**Database:**
- `agent_mentions` table - Stores @mentions
- `agent_delegations` table - Tracks delegations

**API Endpoints:**
- `GET /api/agents/available` - List all agents
- `GET /api/agents/resolve/:mention` - Resolve single mention
- `POST /api/agents/mentions` - Extract mentions from message
- `POST /api/agents/delegate` - Create delegation

**Frontend Structure:**
- `/frontend/src/lib/components/ai-elements/PromptInput.svelte` - Chat input
- `/frontend/src/lib/components/delegation/` - New components folder (to create)

---

## Notes

- ✅ Backend infrastructure mostly complete - focus is frontend
- ✅ API endpoints already exist - just need integration
- ⚠️ Some endpoints may need enhancement (batch resolution)
- 📝 Database schema may need migration 037 for new tables
- 🔄 Follow existing BusinessOS patterns for consistency
- 🎨 Use existing design system (no new design needed)
- ⚡ Performance critical - optimize mention resolution caching

---

**Document Version:** 1.0
**Status:** Ready for Implementation
**Confidence:** High (based on existing code analysis)
**Last Updated:** 2026-01-08
