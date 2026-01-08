# 5 Features Dependency Tree - Visual Summary

**Generated:** 2026-01-08 | **Project:** BusinessOS | **Total Effort:** 158-212 hours

---

## Quick Reference: Feature Dependencies

```
                    SHARED INFRASTRUCTURE
        (Auth, API Client, Stores, PostgreSQL)
                           ▲
                    ┌──────┼──────┐
                    │      │      │
            ┌───────┴──────┐│    ││
            │              │├────┘│
            │              ││     │
            ▼              ▼▼     ▼

        FEATURE 1      FEATURE 2    FEATURE 3      FEATURE 4      FEATURE 5
      TESTING UI    PRESET GALLERY DELEGATION   CATEGORIES    DETAIL PAGE
       28-37h        50-65h         36-48h        24-32h        20-30h
        8 MT           32 MT          18 MT         8 MT          7 MT
                          △               △           △            △
                          │               │           │            │
                    ┌─────┴───────────────┴───────────┴────────────┘
                    │
                    └─ Uses components & data from Features 1,4
```

---

## Feature Dependency Chain

```
INDEPENDENCE LEVELS:
═══════════════════════════════════════════════════════════════

Level 0 (Standalone - can start immediately):
  ├─ Feature 1: Testing UI
  ├─ Feature 2: Preset Gallery
  └─ Feature 4: Categories

Level 1 (Depends on Level 0):
  ├─ Feature 3: Delegation (depends on Feature 1 components)
  └─ Feature 5: Detail Page (depends on Feature 1 components)

Level 2 (Optimal order):
  └─ All features can be done in parallel with proper API contracts
```

---

## Critical Path Analysis

```
LONGEST SEQUENTIAL PATH (Minimum Duration):
═══════════════════════════════════════════════════════════════

Feature 2 (Preset Gallery) - CRITICAL PATH
  MT-1.1 (DB) → MT-1.2 → MT-1.3 (SQLC)
    → MT-2.1 → MT-2.2 → MT-2.3 → MT-2.4 → MT-2.5
    → MT-3.1 → MT-3.2 → MT-3.3
    → MT-7.1 (Tests) → MT-7.3 → MT-7.4
    → MT-8.1 → MT-8.2 → MT-8.3 → MT-8.4
  Duration: 36-40 hours (longest single chain)

PARALLELIZATION OPPORTUNITY:
  While Feature 2 backend builds (9.5h), Feature 2 frontend
  can run parallel (7.5h for MT-4.x)
  Result: 18-20 hours total for Feature 2 instead of 27

WITH 2+ ENGINEERS:
  ├─ Engineer A: Feature 2 Backend (36h)
  └─ Engineer B: Feature 2 Frontend (7.5h) + Feature 4 (24h)
     Total: 3-4 weeks instead of 6 weeks
```

---

## Parallel Execution Opportunities

```
PHASES:
═══════════════════════════════════════════════════════════════

PHASE 1 (Week 1-2) - 3 Engineers
  Engineer A: Feature 1 (Testing UI) ████████ 28-37h
  Engineer B: Feature 4 (Categories) ████████ 24-32h
  Engineer C: Feature 5 (Detail Page database/API) ████████ 8-10h

  Why: These 3 provide infrastructure for Features 2,3

PHASE 2 (Week 3-4) - 4 Engineers
  Engineer A: Feature 2 (Preset Gallery) ████████ 50-65h
  Engineer B: Feature 3 (Delegation) ████████ 36-48h
  Engineer C: Feature 5 (Detail Page UI) ████████ 15-20h
  Engineer D: Testing & Documentation

  Why: Features 2,3 now have components from Phase 1

TOTAL TIMELINE:
  ├─ Sequential: 9 weeks (1 engineer)
  ├─ With parallelization: 4 weeks (3-4 engineers)
  └─ Aggressive: 2-3 weeks (5+ engineers)
```

---

## Component Reusability Map

```
FEATURE 1 (Testing UI) PROVIDES:
═══════════════════════════════════════════════════════════════

AgentTestPanel
  ├─ Reused in Feature 5 (Detail Page) ✓
  └─ Can be embedded in Feature 2 (Preset Gallery) detail modal ✓

Temperature Slider
  ├─ Reused in Feature 3 (Delegation config) ✓
  └─ Reused in Feature 5 (Detail Page settings) ✓

Max Tokens Slider
  ├─ Reused in Feature 3 (Delegation config) ✓
  └─ Reused in Feature 5 (Detail Page settings) ✓

Model Selector Dropdown
  ├─ Reused in Feature 3 (Delegation) ✓
  └─ Can be reused elsewhere

Metrics Visualization
  ├─ Reused in Feature 5 (Detail Page stats tab) ✓
  └─ Can be reused in Feature 2 analytics ✓

Loading Spinner & Error States
  ├─ Used by ALL features ✓✓✓✓✓
  └─ Consider extracting as shared component

REUSABILITY BENEFIT: ~15-20h saved across other features
```

---

## Data Dependencies & Integration Points

```
SHARED DATA STRUCTURES:
═══════════════════════════════════════════════════════════════

custom_agents (existing table)
  ├─ Read by:    Features 1,2,3,4,5 ✓✓✓✓✓
  └─ Write by:   Feature 5 (Detail Page settings)

agent_test_runs (Feature 1 creates)
  ├─ Created by: Feature 1 (Testing UI)
  ├─ Read by:    Feature 5 (Detail Page → Testing tab)
  └─ Read by:    Feature 2 (Preset Gallery → detail modal)

presets (Feature 2 creates)
  ├─ Created by: Feature 2 (Preset Gallery)
  ├─ Filtered by: Feature 4 (Categories)
  └─ Used by:    Feature 5 (Detail Page → agent creation)

delegation_records (Feature 3 creates)
  ├─ Created by: Feature 3 (Delegation System)
  └─ Read by:    Feature 5 (Detail Page → delegation history)

categories (Feature 4 manages)
  ├─ Used by:    Feature 2 (Preset Gallery filter)
  ├─ Used by:    Feature 5 (Detail Page display)
  └─ Used by:    Feature 1 (Test UI categorization - optional)
```

---

## API Endpoint Dependencies

```
FEATURE 1 (Testing UI):
  POST /api/agents/:id/test
    └─ Used by: Feature 5 (Detail Page)
  GET /api/agents/:id/test-history
    └─ Used by: Feature 5 (Detail Page)

FEATURE 2 (Preset Gallery):
  GET /api/presets
    ├─ Filtered by: Feature 4 (Categories)
    └─ Linked by: Feature 5 (Detail Page)
  POST /api/presets/:id/use
    └─ Callable from: Feature 5 (Detail Page)
  POST /api/presets/:id/customize
    └─ Used by: Feature 3 (Delegation agent config)

FEATURE 3 (Delegation):
  GET /api/agents/available
    └─ User for @mention autocomplete
  POST /api/agents/:id/delegate
    └─ Terminal endpoint

FEATURE 4 (Categories):
  GET /api/categories
    └─ Populated: Presets, Agents, Documents
  GET /api/search?categories=...
    ├─ Used by: Feature 2 (gallery search)
    └─ Used by: Feature 5 (detail page)

FEATURE 5 (Detail Page):
  GET /api/agents/:id/details
    └─ Aggregates data from all other endpoints
```

---

## Backend Requirements Summary

```
EXISTING ✅ (No work needed):
  ├─ custom_agents table
  ├─ agent metadata fields
  ├─ auth.users table
  ├─ PostgreSQL database
  └─ Gin framework

FEATURE 1 CREATES:
  ├─ agent_test_runs table
  ├─ agent_test_metrics table
  └─ /api/agents/:id/test endpoint

FEATURE 2 CREATES:
  ├─ presets table (if not exists)
  ├─ preset_versions table
  ├─ 5 API endpoints
  └─ Search service

FEATURE 3 NEEDS (likely exists):
  ├─ /api/agents/available endpoint
  ├─ Delegation execution handler
  └─ Mention tracking service

FEATURE 4 NEEDS (minimal):
  ├─ ?categories parameter support in API
  ├─ Category listing endpoint
  └─ Filtering logic in queries

FEATURE 5 NEEDS:
  ├─ GET /api/agents/:id/details endpoint
  └─ Metrics aggregation service

EFFORT ESTIMATE:
  If all backend exists:      0-5 hours (integration only)
  If backend ~70% complete:   10-15 hours (Feature 3)
  If backend missing:         40-60 hours (Delegation, Detail)
```

---

## Feature Interaction Matrix

```
                Feature 1  Feature 2  Feature 3  Feature 4  Feature 5
                (Testing)  (Presets)  (Delegat)  (Categor)  (Detail)
┌───────────────┬──────────┬──────────┬──────────┬──────────┬──────────┐
│ Feature 1     │    -     │ Reuse UI │ Reuse UI │    -     │ Embedded │
│               │          │  config  │  config  │          │ in tab   │
├───────────────┼──────────┼──────────┼──────────┼──────────┼──────────┤
│ Feature 2     │    -     │    -     │ Presets  │ Filter   │ Link to  │
│               │          │          │ by agent │ presets  │ agent    │
├───────────────┼──────────┼──────────┼──────────┼──────────┼──────────┤
│ Feature 3     │    -     │    -     │    -     │    -     │ Link in  │
│               │          │          │          │          │ history  │
├───────────────┼──────────┼──────────┼──────────┼──────────┼──────────┤
│ Feature 4     │    -     │ Needed   │ Optional │    -     │ Display  │
│               │          │ for      │ agent    │          │ category │
│               │          │ filter   │ category │          │          │
├───────────────┼──────────┼──────────┼──────────┼──────────┼──────────┤
│ Feature 5     │ Uses     │ Links to │ Links to │ Category │    -     │
│               │ testing  │ preset   │ delegat. │ display  │          │
│               │ results  │ creation │ history  │          │          │
└───────────────┴──────────┴──────────┴──────────┴──────────┴──────────┘

INTERACTION DENSITY:
  ├─ Feature 1: ████░░░░░░ 40% (provides components, limited direct use)
  ├─ Feature 2: ████████░░ 80% (central, uses 3-4 other features)
  ├─ Feature 3: ████░░░░░░ 40% (uses Feature 1 components)
  ├─ Feature 4: ██████░░░░ 60% (filters Features 2,5)
  └─ Feature 5: ██████████ 100% (aggregates all other features)
```

---

## Implementation Priority & Order

```
PRIORITY RANKING:
═══════════════════════════════════════════════════════════════

🔴 P0 (Start immediately):
   ├─ Feature 1: Testing UI (enables Features 3,5)
   ├─ Feature 4: Categories (enables Features 2,5)
   └─ Feature 5: Detail Page DB/API (foundation)

🟠 P1 (Start week 2-3):
   ├─ Feature 2: Preset Gallery (largest, depends on P0)
   └─ Feature 3: Delegation (depends on Feature 1)

🟡 P2 (Start week 3-4):
   └─ Feature 5: Detail Page UI (final polish after P0 foundation)

RECOMMENDED SCHEDULE:
  Week 1: Feature 1 + Feature 4 (parallel)
  Week 2: Feature 5 + start Feature 2
  Week 3-4: Feature 2 + Feature 3 (parallel)
  Week 5: Testing, polish, documentation
```

---

## Risks by Dependency

```
🔴 CRITICAL RISKS:
═══════════════════════════════════════════════════════════════

Risk 1: Feature 2 Database (Preset Gallery)
  │ Issue: If schema wrong, blocks 40+ hours of work
  │ Impact: All Feature 2 backend delayed
  │ Mitigation: Review schema on day 1, test with 1000+ records
  │ Timeline: End of day 1

Risk 2: Feature 3 Backend Incomplete
  │ Issue: If /api/agents/available not ready, Feature 3 blocked
  │ Impact: Blocks Features 3 @mention autocomplete (8-12h)
  │ Mitigation: Finalize API contract before Feature 3 starts
  │ Timeline: End of week 1

Risk 3: Feature 1 Component API Changes
  │ Issue: If Testing UI components props change, Features 3,5 break
  │ Impact: Rework in Features 3,5 (10-15h each)
  │ Mitigation: Lock component API by end of week 1
  │ Timeline: Complete Feature 1 before week 2

🟠 MEDIUM RISKS:

Risk 4: Feature 4 Backend Not Ready
  │ Issue: If category filtering not implemented in backend
  │ Impact: Feature 2 uses mock categories (2-3h workaround)
  │ Mitigation: Can implement in parallel, delay integration
  │ Timeline: Feature 4 backend by week 2

Risk 5: Performance Issues Late
  │ Issue: Feature 2 loading 1000 presets slowly
  │ Impact: Major refactoring needed in week 4
  │ Mitigation: Load test early (day 3), add pagination
  │ Timeline: Feature 2 database load test

Risk 6: Component Reuse Mismatch
  │ Issue: Features 3,5 want different slider behavior
  │ Impact: Can't reuse from Feature 1, rework (5-10h)
  │ Mitigation: Design shared component API early
  │ Timeline: End of Feature 1
```

---

## Resource Allocation Recommendations

```
OPTION A: 3-Person Team (14 weeks)
═══════════════════════════════════════════════════════════════

Engineer 1 (Backend - 280 total hours):
  Week 1-2: Feature 1 DB + API (10h)
  Week 3-4: Feature 4 backend (optional, 4h)
  Week 5-8: Feature 2 backend (56h) - CRITICAL PATH
  Week 9-11: Feature 3 backend (12h)
  Week 12-14: Feature 5 backend (8h)

Engineer 2 (Frontend - 280 total hours):
  Week 1-2: Feature 1 components (28h)
  Week 3-4: Feature 4 components (24h)
  Week 5-8: Feature 2 components (40h)
  Week 9-11: Feature 3 components (32h)
  Week 12-14: Feature 5 components (20h)

Engineer 3 (QA/DevOps - 140 total hours):
  Week 1-14: Testing, performance monitoring, documentation
  Week 8+: Feature 2 load testing (critical)
  Week 14: Production deployment

Total: 700 hours / 3 people = 233h per person ≈ 14 weeks

═══════════════════════════════════════════════════════════════

OPTION B: 5-Person Team (6 weeks) - RECOMMENDED
═══════════════════════════════════════════════════════════════

Team Backend (2 engineers):
  ├─ Engineer A: Features 1,4,5 backend (24h)
  └─ Engineer B: Feature 2 backend (56h)
             ↑
             └─ Both can start week 1 (parallel)

Team Frontend (2 engineers):
  ├─ Engineer C: Features 1,4,5 frontend (72h)
  └─ Engineer D: Features 2,3 frontend (72h)
             ↑
             └─ Start staggered when backend ready

Team QA/DevOps (1 engineer):
  └─ Engineer E: Testing, deployment, feature integration (60h)

Total: 284 hours / 5 people = 57h per person ≈ 6 weeks

═══════════════════════════════════════════════════════════════

OPTION C: 7-Person Team (4 weeks) - AGGRESSIVE
═══════════════════════════════════════════════════════════════

Parallel Teams (1 per feature + shared):
  ├─ Team 1: Features 1,4 (4 people)
  ├─ Team 2: Feature 2 (2 people)
  ├─ Team 3: Feature 3 (1 person, blocked by Team 1)
  ├─ Team 4: Feature 5 (1 person, blocked by Team 1)
  └─ Team 5: QA/DevOps (1 person, parallel)

Total: ~4 weeks with 7 engineers
Note: Requires weekly sync on API contracts, tight coordination
```

---

## Success Metrics by Feature

```
FEATURE 1 (Testing UI):
  ✓ Component renders <100ms
  ✓ 80%+ test coverage
  ✓ No console errors
  ✓ Components reusable (verified in Features 3,5)

FEATURE 2 (Preset Gallery):
  ✓ 90%+ backend test coverage
  ✓ 80%+ frontend test coverage
  ✓ Loads 1000+ presets <2s
  ✓ Search response <200ms
  ✓ Categories filter working

FEATURE 3 (Delegation):
  ✓ @mention autocomplete works
  ✓ Delegation executes successfully
  ✓ Rate limiting handled
  ✓ Error messages clear

FEATURE 4 (Categories):
  ✓ Filter persistence in URL
  ✓ 80%+ test coverage
  ✓ Animations 60fps smooth
  ✓ Mobile responsive

FEATURE 5 (Detail Page):
  ✓ All tabs render correctly
  ✓ Testing tab shows Feature 1 data
  ✓ Page loads <2s
  ✓ Responsive on all devices

OVERALL:
  ✓ All workflows end-to-end tested
  ✓ 85%+ total test coverage
  ✓ Production deployment successful
```

---

## Timeline Visualization

```
SEQUENTIAL (1 engineer):
Feature 1 ████████ 28-37h
Feature 4 ████████ 24-32h
Feature 5 ██████ 20-30h
Feature 2 █████████████ 50-65h
Feature 3 ███████████ 36-48h
                                      Week 1  2   3   4   5   6   7   8   9
                                      └─────────────────────────────────────┘
                                                    9 weeks

WITH 3 ENGINEERS (RECOMMENDED):
Eng A (Backend)  ████ ██████████████ ███████
Eng B (Frontend) ████ ████████ ██████████████ ███████
Eng C (QA)       ███████████████████████████████████
                                      Week 1  2   3   4   5   6   7
                                      └─────────────────────────────┘
                                                7 weeks

WITH 5 ENGINEERS (AGGRESSIVE):
Backend Team ████ ████████████████ █████
Frontend     ████ ████████ ██████████████ █████
Feature 3    ░░░░ ███████████ ███████
Feature 5    ░░░░ ███████ █████████
QA/DevOps    ███████████████████████████████
                                      Week 1  2   3   4   5   6
                                      └─────────────────────────┘
                                                6 weeks

LEGEND:  ████ Complete   ░░░░ Blocked   ─── Starting
```

---

## Final Recommendations

### ✅ DO:
1. **Start Features 1 + 4 together** (no dependencies)
2. **Lock Feature 1 component APIs** by end of week 1
3. **Test database schemas early** (Feature 2 critical path)
4. **Use 3+ engineers** for parallelization benefits
5. **Define API contracts** before frontend starts
6. **Write tests incrementally** (not at the end)
7. **Load test Feature 2** early (1000+ presets)
8. **Weekly sync** on cross-feature dependencies

### ❌ DON'T:
1. **Don't wait on all Features 1,4** to start Feature 2
2. **Don't change component APIs** mid-feature
3. **Don't start Feature 3** without Feature 1 components ready
4. **Don't delay performance testing** until week 5
5. **Don't ignore category system design** (blocks Features 2,4,5)
6. **Don't start everything in parallel** (insufficient coordination)

### 🎯 OPTIMAL STRATEGY:
```
Week 1-2:  Features 1 + 4 (foundation infrastructure)
Week 3-4:  Feature 5 API + Feature 2 backend (parallel)
Week 5-6:  All components (Feature 2,3,5 frontend in parallel)
Week 7:    Testing, performance, polish
Week 8:    Documentation, deployment
```

---

**Status:** Ready for Implementation
**Last Updated:** 2026-01-08
**Next Step:** Assign teams, create detailed sprint planning
