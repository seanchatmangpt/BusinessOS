# Preset Gallery - Complete Dependency Graph

## Visual Dependency Map

### All 32 Microtasks with Dependencies

```
PHASE 1: DATABASE SETUP
═══════════════════════════════════════════════════════════════════════════════

                    ┌─────────────────┐
                    │  MT-1.1: Create │
                    │  Presets        │
                    │  Migration      │
                    │  (1-2h)         │
                    └────────┬────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │  MT-1.2: Create │
                    │  Analytics      │
                    │  Migration      │
                    │  (1h)           │
                    └────────┬────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │  MT-1.3: Generate│
                    │  SQLC Go Code   │
                    │  (1h)           │
                    └────────┬────────┘
                             │
        ┌────────────────────┴────────────────────┐
        │                                         │
        ▼                                         ▼

PHASE 2: BACKEND API
═══════════════════════════════════════════════════════════════════════════════

TRACK A - CORE API                  TRACK B - CREATE FROM PRESET
┌──────────────────────┐            ┌──────────────────────┐
│ MT-2.1: List API     │            │ MT-3.1: Use Preset   │
│ (2h)                 │            │ Endpoint (2.5h)      │
└──────────┬───────────┘            │                      │
           │                        │ (Depends on MT-2.5)  │
           ▼                        └──────────┬───────────┘
┌──────────────────────┐                       │
│ MT-2.2: Repository   │                       ▼
│ Layer (2h)           │            ┌──────────────────────┐
└──────────┬───────────┘            │ MT-3.2: Config       │
           │                        │ Validator (1.5h)     │
           ▼                        └──────────┬───────────┘
┌──────────────────────┐                       │
│ MT-2.3: Service      │                       ▼
│ Layer (2h)           │            ┌──────────────────────┐
└──────────┬───────────┘            │ MT-3.3: Customize    │
           │                        │ Endpoint (2h)        │
           ▼                        └──────────────────────┘
┌──────────────────────┐
│ MT-2.4: Get Detail   │
│ Endpoint (1.5h)      │
└──────────┬───────────┘
           │
           ▼
┌──────────────────────┐
│ MT-2.5: Search API   │
│ (2h)                 │
└──────────┬───────────┘
           │
           └──────────────────────────────┐
                                          │
        ┌─────────────────────────────────┘
        │
        ▼

PHASE 3A: FRONTEND GALLERY COMPONENTS
═══════════════════════════════════════════════════════════════════════════════

           ┌─────────────────┐
           │ MT-4.1: Gallery │
           │ Page Component  │
           │ (2h)            │
           └────────┬────────┘
                    │
                    ▼
           ┌─────────────────┐
           │ MT-4.2: Gallery │
           │ Store (Svelte)  │
           │ (1.5h)          │
           └────────┬────────┘
                    │
                    ▼
           ┌─────────────────┐
           │ MT-4.3: Preset  │
           │ Card Component  │
           │ (1.5h)          │
           └────────┬────────┘
                    │
                    ▼
           ┌─────────────────┐
           │ MT-4.4: Grid    │
           │ Component       │
           │ (1.5h)          │
           └────────┬────────┘
                    │
                    ▼
           ┌─────────────────┐
           │ MT-4.5: Category│
           │ Sidebar         │
           │ (1.5h)          │
           └────────┬────────┘
                    │
        ┌───────────┴───────────┐
        │                       │
        ▼                       ▼

PHASE 3B: FRONTEND MODALS & INTERACTIONS
═══════════════════════════════════════════════════════════════════════════════

┌──────────────────────┐    ┌──────────────────────────────┐
│ MT-5.1: Detail Modal │    │ API READY: MT-3.1, MT-3.2,  │
│ Component (2.5h)     │    │ MT-3.3 (Backend Complete)    │
└──────────┬───────────┘    └──────────────────────────────┘
           │                                    ▲
           │                                    │
           ▼                                    │
┌──────────────────────┐                       │
│ MT-5.2: Use Preset   │◄──────────────────────┘
│ Flow (2h)            │
└──────────┬───────────┘
           │
           ▼
┌──────────────────────┐     ┌──────────────────────────────┐
│ MT-5.3: Customize    │     │ API READY: MT-3.3 (Config    │
│ Modal (2.5h)         │◄────│ Customization Endpoint)      │
└──────────┬───────────┘     └──────────────────────────────┘
           │
           ▼
┌──────────────────────┐
│ MT-5.4: Search &     │
│ Filter Integration   │
│ (1.5h)               │
└──────────┬───────────┘
           │
        ┌──┴──┐
        │     │
        ▼     ▼

PHASE 3C: FRONTEND ADVANCED SEARCH
═══════════════════════════════════════════════════════════════════════════════

┌──────────────────────┐    ┌──────────────────────┐
│ MT-6.1: Advanced     │    │ MT-6.2: Search       │
│ Search Form (2h)     │───▶│ Results (1.5h)       │
└──────────────────────┘    └──────────┬───────────┘
                                       │
                                       ▼
                            ┌──────────────────────┐
                            │ MT-6.3: Favorites    │
                            │ (1.5h)               │
                            └──────────┬───────────┘
                                       │
                            (Optional) ▼
                            ┌──────────────────────┐
                            │ MT-6.4: Analytics    │
                            │ Dashboard (2h)       │
                            └──────────────────────┘
                                       │
        ┌──────────────────────────────┘
        │
        ▼

PHASE 4: TESTING & QUALITY ASSURANCE
═══════════════════════════════════════════════════════════════════════════════

┌────────────────────────┐     ┌────────────────────────┐
│ MT-7.1: API            │     │ MT-7.2: Frontend       │
│ Integration Tests      │     │ Component Tests        │
│ (2h)                   │     │ (2h)                   │
│                        │     │                        │
│ Depends on: MT-3.3     │     │ Depends on: MT-6.3     │
│ (All API complete)     │     │ (All frontend complete)│
└────────────┬───────────┘     └────────────┬───────────┘
             │                              │
             │ (Can run in parallel) ◄─────┤
             │                              │
             └──────────────┬───────────────┘
                            │
                            ▼
                 ┌────────────────────────┐
                 │ MT-7.3: E2E Tests      │
                 │ Complete Workflows     │
                 │ (2.5h)                 │
                 │                        │
                 │ Depends on: Both       │
                 │ MT-7.1 & MT-7.2       │
                 └────────────┬───────────┘
                              │
                              ▼
                 ┌────────────────────────┐
                 │ MT-7.4: Performance &  │
                 │ Security Audit (2h)    │
                 │                        │
                 │ Depends on: MT-7.3     │
                 │ (All tests passing)    │
                 └────────────┬───────────┘
                              │
        ┌─────────────────────┘
        │
        ▼

PHASE 5: DOCUMENTATION & DEPLOYMENT
═══════════════════════════════════════════════════════════════════════════════

┌──────────────────────┐     ┌──────────────────────┐
│ MT-8.1: API Docs     │     │ MT-8.2: User Guide   │
│ (1.5h)               │     │ (1.5h)               │
│                      │     │                      │
│ Depends on: MT-7.4   │     │ Depends on: MT-8.1   │
└──────────┬───────────┘     └──────────┬───────────┘
           │                            │
           └──────────────┬─────────────┘
                          │
                          ▼
                 ┌──────────────────────┐
                 │ MT-8.3: Deployment   │
                 │ Guide (1h)           │
                 │                      │
                 │ Depends on: MT-8.2   │
                 └──────────┬───────────┘
                            │
                            ▼
                 ┌──────────────────────┐
                 │ MT-8.4: Production   │
                 │ Deployment (1.5h)    │
                 │                      │
                 │ Depends on: MT-8.3   │
                 │ FINAL TASK           │
                 └──────────────────────┘
```

---

## Linear Dependency Chain (Critical Path)

```
MT-1.1 (1.5h)
    ↓
MT-1.2 (1h)
    ↓
MT-1.3 (1h)
    ↓ Branch A (Backend)        Branch B (Frontend)
    ├─→ MT-2.1 (2h)             └─→ MT-4.1 (2h)
    │   MT-2.2 (2h)                 MT-4.2 (1.5h)
    │   MT-2.3 (2h)                 MT-4.3 (1.5h)
    │   MT-2.4 (1.5h)                MT-4.4 (1.5h)
    │   MT-2.5 (2h)                 MT-4.5 (1.5h)
    │   MT-3.1 (2.5h)                MT-5.1 (2.5h)
    │   MT-3.2 (1.5h)                MT-5.2 (2h) ◄── Waits for MT-3.1
    │   MT-3.3 (2h)                  MT-5.3 (2.5h) ◄── Waits for MT-3.3
    │                                MT-5.4 (1.5h)
    │                                MT-6.1 (2h)
    │                                MT-6.2 (1.5h)
    │                                MT-6.3 (1.5h)
    │                                MT-6.4 (2h) [Optional]
    │
    ├─ Merge Point: Both branches complete
    │
    ├─→ MT-7.1 (2h) ◄─ Waits for MT-3.3 (API complete)
    │   MT-7.2 (2h) ◄─ Waits for MT-6.3 (Frontend complete)
    │   [Both can run in parallel]
    │
    ├─→ MT-7.3 (2.5h) ◄─ Waits for both MT-7.1 & MT-7.2
    │
    ├─→ MT-7.4 (2h) ◄─ Waits for MT-7.3
    │
    ├─→ MT-8.1 (1.5h) ◄─ Waits for MT-7.4
    │
    ├─→ MT-8.2 (1.5h) ◄─ Waits for MT-8.1
    │
    ├─→ MT-8.3 (1h) ◄─ Waits for MT-8.2
    │
    └─→ MT-8.4 (1.5h) ◄─ Waits for MT-8.3 [FINAL TASK]
```

---

## Dependency Matrix (Complete)

```
┌─────────┬──────────────────────────┬─────────────┬─────────────────┐
│ Task ID │ Task Name                │ Dependencies│ Blocked By      │
├─────────┼──────────────────────────┼─────────────┼─────────────────┤
│ MT-1.1  │ Presets Migration        │ None        │ None            │
│ MT-1.2  │ Analytics Migration      │ MT-1.1      │ MT-1.1          │
│ MT-1.3  │ SQLC Generation          │ MT-1.2      │ MT-1.2          │
├─────────┼──────────────────────────┼─────────────┼─────────────────┤
│ MT-2.1  │ List Presets             │ MT-1.3      │ MT-1.3          │
│ MT-2.2  │ Repository Layer         │ MT-2.1      │ MT-2.1          │
│ MT-2.3  │ Service Layer            │ MT-2.2      │ MT-2.2          │
│ MT-2.4  │ Get Detail Endpoint      │ MT-2.3      │ MT-2.3          │
│ MT-2.5  │ Search Endpoint          │ MT-2.4      │ MT-2.4          │
├─────────┼──────────────────────────┼─────────────┼─────────────────┤
│ MT-3.1  │ Use Preset Endpoint      │ MT-2.5      │ MT-2.5          │
│ MT-3.2  │ Config Validator         │ MT-3.1      │ MT-3.1          │
│ MT-3.3  │ Customize Endpoint       │ MT-3.2      │ MT-3.2          │
├─────────┼──────────────────────────┼─────────────┼─────────────────┤
│ MT-4.1  │ Gallery Page             │ None        │ None            │
│ MT-4.2  │ Gallery Store            │ MT-4.1      │ MT-4.1          │
│ MT-4.3  │ Preset Card              │ MT-4.2      │ MT-4.2          │
│ MT-4.4  │ Grid Component           │ MT-4.3      │ MT-4.3          │
│ MT-4.5  │ Category Sidebar         │ MT-4.4      │ MT-4.4          │
├─────────┼──────────────────────────┼─────────────┼─────────────────┤
│ MT-5.1  │ Detail Modal             │ MT-4.5      │ MT-4.5          │
│ MT-5.2  │ Use Preset Flow          │ MT-5.1,3.1  │ MT-3.1, MT-5.1  │
│ MT-5.3  │ Customize Modal          │ MT-5.2,3.3  │ MT-3.3, MT-5.2  │
│ MT-5.4  │ Search Integration       │ MT-5.3      │ MT-5.3          │
├─────────┼──────────────────────────┼─────────────┼─────────────────┤
│ MT-6.1  │ Advanced Search Form     │ MT-5.4      │ MT-5.4          │
│ MT-6.2  │ Search Results           │ MT-6.1      │ MT-6.1          │
│ MT-6.3  │ Favorites                │ MT-6.2      │ MT-6.2          │
│ MT-6.4  │ Analytics Dashboard      │ MT-6.3      │ MT-6.3 (Opt)    │
├─────────┼──────────────────────────┼─────────────┼─────────────────┤
│ MT-7.1  │ API Tests                │ MT-3.3      │ MT-3.3          │
│ MT-7.2  │ Frontend Tests           │ MT-6.3      │ MT-6.3          │
│ MT-7.3  │ E2E Tests                │ MT-7.1,7.2  │ MT-7.1, MT-7.2  │
│ MT-7.4  │ Audit & Security         │ MT-7.3      │ MT-7.3          │
├─────────┼──────────────────────────┼─────────────┼─────────────────┤
│ MT-8.1  │ API Documentation        │ MT-7.4      │ MT-7.4          │
│ MT-8.2  │ User Guide               │ MT-8.1      │ MT-8.1          │
│ MT-8.3  │ Deployment Guide         │ MT-8.2      │ MT-8.2          │
│ MT-8.4  │ Production Deploy        │ MT-8.3      │ MT-8.3          │
└─────────┴──────────────────────────┴─────────────┴─────────────────┘
```

---

## Parallel Execution Opportunities

### Level 1: Can Start Immediately (No Dependencies)
```
┌──────────────┐    ┌──────────────┐
│ MT-1.1       │    │ MT-4.1       │
│ Database     │    │ Gallery Page │
└──────────────┘    └──────────────┘
```

### Level 2: After Phase 1 Complete (2 hours)
```
┌──────────────┐    ┌──────────────┐
│ MT-2.1       │    │ MT-4.2       │
│ API List     │    │ Store        │
└──────────────┘    └──────────────┘
```

### Level 3: After Earlier Tasks (4-5 hours)
```
┌──────────────┐    ┌──────────────┐
│ MT-2.2       │    │ MT-4.3       │
│ Repository   │    │ Card         │
└──────────────┘    └──────────────┘
```

And so on... Maximum parallelization depth: 3-4 task levels

---

## Blocking Analysis

### Tasks That Block Many Others
```
BLOCKERS (High Impact):
┌─────────────────────────────────────────────────────┐
│ MT-1.3: SQLC Generation                             │
│   ├─ Blocks all Phase 2 (MT-2.1 through MT-3.3)    │
│   └─ 1 hour to complete, blocks 8 tasks (40 hours)  │
│                                                      │
│ MT-3.3: Customize Endpoint                          │
│   ├─ Blocks MT-5.3, MT-7.1                          │
│   └─ 2 hours to complete, blocks 2 tasks (5 hours)  │
│                                                      │
│ MT-5.4: Search Integration                          │
│   ├─ Blocks MT-6.1 → MT-6.2 → MT-6.3              │
│   └─ 1.5 hours to complete, blocks 3 tasks (5h)    │
│                                                      │
│ MT-7.4: Audit & Security                            │
│   ├─ Blocks MT-8.1 → MT-8.2 → MT-8.3 → MT-8.4    │
│   └─ 2 hours to complete, blocks 4 tasks (5 hours)  │
└─────────────────────────────────────────────────────┘
```

### Critical Path Analysis
```
Longest sequential path:
MT-1.1 → MT-1.2 → MT-1.3 → MT-2.1 → MT-2.2 → MT-2.3
→ MT-2.4 → MT-2.5 → MT-3.1 → MT-3.2 → MT-3.3
→ MT-7.1 → MT-7.3 → MT-7.4 → MT-8.1 → MT-8.2
→ MT-8.3 → MT-8.4

Total Duration: 36 hours (critical path)
Can be reduced to ~20 hours with optimal parallelization
```

---

## Optimal Parallel Execution Plan

### Execution Timeline (2 Engineers)

**Day 1 Morning (Engineer A: Database, Engineer B: Frontend)**
```
Engineer A:
  [0.00h] Start MT-1.1: Create Presets Migration (1.5h)
  [1.50h] Start MT-1.2: Create Analytics (1h)
  [2.50h] Start MT-1.3: SQLC Generation (1h)
  [3.50h] → Ready for Phase 2

Engineer B (parallel):
  [0.00h] Start MT-4.1: Gallery Page (2h)
  [2.00h] Start MT-4.2: Store (1.5h)
  [3.50h] Start MT-4.3: Card (1.5h)
  [5.00h] → Ready for Phase 3B
```

**Day 1 Afternoon + Day 2**
```
Engineer A:
  [3.50h] Start MT-2.1: List API (2h)
  [5.50h] Start MT-2.2: Repository (2h)
  [7.50h] Start MT-2.3: Service (2h)
  [9.50h] Start MT-2.4: Detail (1.5h)
  [11.00h] Start MT-2.5: Search (2h)
  [13.00h] Start MT-3.1: Use (2.5h)
  [15.50h] Start MT-3.2: Validator (1.5h)
  [17.00h] Start MT-3.3: Customize (2h)
  [19.00h] → Ready for Testing

Engineer B (parallel):
  [5.00h] Start MT-4.4: Grid (1.5h)
  [6.50h] Start MT-4.5: Sidebar (1.5h)
  [8.00h] Start MT-5.1: Detail Modal (2.5h)
  [10.50h] Start MT-5.2: Use Flow (2h)
  [12.50h] Start MT-5.3: Customize (2.5h)
  [15.00h] Start MT-5.4: Search (1.5h)
  [16.50h] Start MT-6.1: Adv Search (2h)
  [18.50h] Start MT-6.2: Results (1.5h)
  [20.00h] Start MT-6.3: Favorites (1.5h)
  [21.50h] → Ready for Testing
```

**Day 3 (Both Engineers)**
```
Engineer A:
  [19.00h] Start MT-7.1: API Tests (2h) [parallel with Engineer B]

Engineer B (parallel):
  [21.50h] Start MT-7.2: Frontend Tests (2h) [parallel with Engineer A]

After both complete (both have tested independently):
  Both: Start MT-7.3: E2E Tests (2.5h)
  Both: Start MT-7.4: Audit (2h)
```

**Day 4 (Both Engineers)**
```
Both:
  MT-8.1: API Docs (1.5h) [parallel]
  MT-8.2: User Docs (1.5h) [parallel]
  MT-8.3: Deploy Guide (1h)
  MT-8.4: Production Deploy (1.5h)
```

**Total Timeline: 4-5 days with 2 engineers**

---

## Alternative: Aggressive Parallelization (3+ Engineers)

```
Team 1 (Backend): MT-1.x → MT-2.x → MT-3.x → MT-7.1
Team 2 (Frontend): MT-4.x → MT-5.x → MT-6.x → MT-7.2
Team 3 (QA): MT-7.3, MT-7.4
Team 4 (DevOps): MT-8.x

Minimum Timeline: 3-4 days
```

---

## Blocking Hazards & Mitigation

### Hazard 1: Backend Dependency Hell (MT-1.3)
```
Risk: SQLC generation fails, blocks all backend
Mitigation:
  - Test sqlc config early
  - Have fallback manual code generation
  - Run in parallel to catch issues
```

### Hazard 2: API-Frontend Integration (MT-5.2, MT-5.3)
```
Risk: API contract mismatches between engineers
Mitigation:
  - Define API contract in OpenAPI/Postman first
  - Frontend uses mock API until real API ready
  - Daily sync on API changes
```

### Hazard 3: Testing Takes Longer Than Code
```
Risk: MT-7.x tasks exceed estimates
Mitigation:
  - Write tests incrementally, not at end
  - Aim for 90%+ coverage from start
  - Use TDD where possible
```

### Hazard 4: Performance Issues Discovered Late
```
Risk: MT-7.4 audit finds major performance problems
Mitigation:
  - Profile early (after MT-2.5)
  - Load test with 1000+ presets during MT-2.5
  - Add proper database indexes in MT-1.1
```

---

## Success Criteria by Dependency Level

**Level 1 (Database)**
- [ ] Migrations apply to clean database
- [ ] SQLC generates without warnings
- [ ] Can query sample data successfully

**Level 2A (Backend - Core API)**
- [ ] All 5 endpoints respond 200 OK
- [ ] Pagination works correctly
- [ ] Error handling returns proper status codes

**Level 2B (Backend - Create)**
- [ ] Can create agent from preset
- [ ] Config validation works
- [ ] Customization preserves valid configs

**Level 3A (Frontend - Gallery)**
- [ ] Page renders without console errors
- [ ] Grid responsive on all screen sizes
- [ ] Store state manages correctly

**Level 3B (Frontend - Modals)**
- [ ] Modals open/close cleanly
- [ ] Flow completes without errors
- [ ] API integration working

**Level 3C (Frontend - Search)**
- [ ] Search filters work
- [ ] Results display correctly
- [ ] Favorites persist

**Level 4 (Testing)**
- [ ] 90%+ API coverage, 80%+ Frontend coverage
- [ ] All E2E flows passing
- [ ] No security issues found
- [ ] Performance < 200ms for API calls

**Level 5 (Deployment)**
- [ ] All docs complete
- [ ] Staging deployment successful
- [ ] Production deployment successful
- [ ] Rollback procedure verified

---

**Last Updated**: 2026-01-08
**Version**: 1.0

