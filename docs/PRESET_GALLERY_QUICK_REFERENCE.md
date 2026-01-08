# Preset Gallery - Quick Reference Card

## Microtask Index (32 Total)

### Phase 1: Database (3 tasks, 4 hours)
| ID | Task | Effort | Blocker | Status |
|----|------|--------|---------|--------|
| MT-1.1 | Create Presets Migration | 1-2h | - | Pending |
| MT-1.2 | Create Analytics Migration | 1h | MT-1.1 | Pending |
| MT-1.3 | Generate SQLC Go Code | 1h | MT-1.2 | Pending |

### Phase 2A: Backend API - Core (5 tasks, 7 hours)
| ID | Task | Effort | Blocker | Status |
|----|------|--------|---------|--------|
| MT-2.1 | List Presets Endpoint | 2h | MT-1.3 | Pending |
| MT-2.2 | Presets Repository | 2h | MT-2.1 | Pending |
| MT-2.3 | Presets Service | 2h | MT-2.2 | Pending |
| MT-2.4 | Get Preset Detail | 1.5h | MT-2.3 | Pending |
| MT-2.5 | Advanced Search | 2h | MT-2.4 | Pending |

### Phase 2B: Backend API - Create (3 tasks, 5.5 hours)
| ID | Task | Effort | Blocker | Status |
|----|------|--------|---------|--------|
| MT-3.1 | Use Preset Endpoint | 2.5h | MT-2.5 | Pending |
| MT-3.2 | Config Validator | 1.5h | MT-3.1 | Pending |
| MT-3.3 | Customize Endpoint | 2h | MT-3.2 | Pending |

### Phase 3A: Frontend - Gallery (5 tasks, 7 hours)
| ID | Task | Effort | Blocker | Status |
|----|------|--------|---------|--------|
| MT-4.1 | Gallery Page | 2h | - | Pending |
| MT-4.2 | Gallery Store | 1.5h | MT-4.1 | Pending |
| MT-4.3 | Preset Card | 1.5h | MT-4.2 | Pending |
| MT-4.4 | Grid Component | 1.5h | MT-4.3 | Pending |
| MT-4.5 | Category Sidebar | 1.5h | MT-4.4 | Pending |

### Phase 3B: Frontend - Modals (4 tasks, 6.5 hours)
| ID | Task | Effort | Blocker | Status |
|----|------|--------|---------|--------|
| MT-5.1 | Detail Modal | 2.5h | MT-4.5 | Pending |
| MT-5.2 | Use Preset Flow | 2h | MT-5.1, MT-3.1 | Pending |
| MT-5.3 | Customize Modal | 2.5h | MT-5.2, MT-3.3 | Pending |
| MT-5.4 | Search Integration | 1.5h | MT-5.3 | Pending |

### Phase 3C: Frontend - Search (4 tasks, 5.5 hours)
| ID | Task | Effort | Blocker | Status |
|----|------|--------|---------|--------|
| MT-6.1 | Advanced Search Form | 2h | MT-5.4 | Pending |
| MT-6.2 | Search Results | 1.5h | MT-6.1 | Pending |
| MT-6.3 | Favorites/Bookmarks | 1.5h | MT-6.2 | Pending |
| MT-6.4 | Analytics Dashboard | 2h | MT-6.3 | Optional |

### Phase 4: Testing (4 tasks, 6.5 hours)
| ID | Task | Effort | Blocker | Status |
|----|------|--------|---------|--------|
| MT-7.1 | API Integration Tests | 2h | MT-3.3 | Pending |
| MT-7.2 | Frontend Tests | 2h | MT-6.3 | Pending |
| MT-7.3 | E2E Tests | 2.5h | MT-7.1, MT-7.2 | Pending |
| MT-7.4 | Audit & Security | 2h | MT-7.3 | Pending |

### Phase 5: Docs & Deploy (4 tasks, 5.5 hours)
| ID | Task | Effort | Blocker | Status |
|----|------|--------|---------|--------|
| MT-8.1 | API Documentation | 1.5h | MT-7.4 | Pending |
| MT-8.2 | User Documentation | 1.5h | MT-8.1 | Pending |
| MT-8.3 | Deployment Guide | 1h | MT-8.2 | Pending |
| MT-8.4 | Production Deploy | 1.5h | MT-8.3 | Pending |

---

## Execution Strategies

### Strategy 1: Solo Development (8 days)
**Best for**: Single engineer, lower priority feature
```
Week 1: MT-1 (1d) → MT-2 (2d) → MT-3 (1d) → MT-4 (2d)
Week 2: MT-5 (1.5d) → MT-6 (1.5d) → MT-7 (2d) → MT-8 (1d)
```

### Strategy 2: Parallel Backend + Frontend (5-6 days)
**Best for**: 2 engineers, aggressive timeline ⭐ RECOMMENDED
```
Backend Engineer:     Frontend Engineer:
Phase 1: DB (0.5d)   Phase 3: UI (1.5d)
Phase 2: API (2d)    Phase 3B: Modals (1d)
Phase 2B: Create (1d) Phase 3C: Search (1d)

Both: Phase 4: Testing (1.5d)
Both: Phase 5: Docs & Deploy (1d)
```

### Strategy 3: Incremental Delivery
**Best for**: Feature flags, iterative release
```
Release 1 (Day 3): MT-4.1-4.5 + Mock API
Release 2 (Day 6): MT-2.1-2.5 + MT-3.1
Release 3 (Day 8): MT-6.1-6.3 + Full Integration
```

---

## Parallel Execution Matrix

### Can Run Simultaneously
```
✅ MT-1.x (Database) + MT-4.x (Frontend) - Independent layers
✅ MT-2.1 + MT-4.1 - Different systems
✅ MT-7.1 (API Tests) + MT-7.2 (Frontend Tests) - Independent
✅ MT-8.1 + MT-8.2 - Documentation in parallel
```

### Must Be Sequential
```
❌ MT-1.1 → MT-1.2 → MT-1.3 - Database migrations ordered
❌ MT-2.1 → MT-2.2 → MT-2.3 - Layer dependencies
❌ MT-4.1 → MT-4.2 → MT-4.3 - Component composition
❌ MT-7.1/MT-7.2 → MT-7.3 → MT-7.4 - Testing pipeline
```

---

## Daily Checklists

### Backend Engineer (MT-1, MT-2, MT-3, MT-7.1)

**Day 1**
- [ ] MT-1.1: Create migrations/037_agent_presets.sql
- [ ] MT-1.2: Create migrations/038_preset_analytics.sql
- [ ] MT-1.3: Run sqlc, verify generated code

**Day 2-3**
- [ ] MT-2.1: Create handlers/presets.go (ListPresets)
- [ ] MT-2.2: Create repository/presets_repository.go
- [ ] MT-2.3: Create services/presets_service.go

**Day 4**
- [ ] MT-2.4: Add GetPreset endpoint
- [ ] MT-2.5: Add search endpoint with full-text

**Day 5**
- [ ] MT-3.1: Add use-preset endpoint
- [ ] MT-3.2: Create preset_config_validator.go

**Day 6**
- [ ] MT-3.3: Add customize endpoint
- [ ] MT-7.1: Write API integration tests (90%+ coverage)

### Frontend Engineer (MT-4, MT-5, MT-6, MT-7.2)

**Day 1-2**
- [ ] MT-4.1: Create routes/(app)/gallery/+page.svelte
- [ ] MT-4.2: Create lib/stores/gallery.ts
- [ ] MT-4.3: Create components/gallery/PresetCard.svelte
- [ ] MT-4.4: Create components/gallery/PresetGrid.svelte

**Day 3**
- [ ] MT-4.5: Create components/gallery/CategorySidebar.svelte
- [ ] MT-5.1: Create components/gallery/PresetDetailModal.svelte

**Day 4**
- [ ] MT-5.2: Create components/gallery/UsePresetFlow.svelte
- [ ] MT-5.3: Create components/gallery/CustomizePresetModal.svelte

**Day 5**
- [ ] MT-5.4: Integrate search & filters
- [ ] MT-6.1: Create AdvancedSearch.svelte
- [ ] MT-6.2: Create SearchResults.svelte
- [ ] MT-6.3: Create FavoritesButton.svelte

**Day 6**
- [ ] MT-7.2: Write component tests (80%+ coverage)

### Both Engineers

**Day 7**
- [ ] MT-7.3: Write E2E tests covering all flows
- [ ] MT-7.4: Performance audit & security review

**Day 8**
- [ ] MT-8.1: Create API documentation
- [ ] MT-8.2: Create user guide
- [ ] MT-8.3: Create deployment guide
- [ ] MT-8.4: Deploy to staging then production

---

## API Endpoints Summary

**List & Search**
```
GET /api/presets?category=X&search=Y&limit=20&offset=0
POST /api/presets/search {query, category, sort}
GET /api/presets/{id}
```

**Create from Preset**
```
POST /api/presets/{id}/use {workspace_id, agent_name}
POST /api/presets/{id}/customize {config_overrides}
```

**Response Format**
```json
{
  "id": "uuid",
  "name": "string",
  "description": "string",
  "category": "string",
  "config": {object},
  "usage_count": integer,
  "created_at": "timestamp",
  "tags": ["string"]
}
```

---

## File Count & LOC Estimates

| Category | Files | Estimated LOC |
|----------|-------|----------------|
| Database | 2 SQL | 150 |
| Backend Code | 5 .go | 800 |
| Backend Tests | 1 .go | 400 |
| Frontend Components | 9 .svelte | 1200 |
| Frontend Stores | 2 .ts | 200 |
| Frontend Tests | 2 .ts | 500 |
| E2E Tests | 1 .ts | 300 |
| Documentation | 3 .md | 500 |
| **TOTAL** | **25** | **~4,050** |

---

## Common Pitfalls & Solutions

| Pitfall | Solution |
|---------|----------|
| Forgetting pagination in list endpoint | Check MT-2.1 acceptance criteria |
| Not validating preset config | Implement MT-3.2 validator service |
| Hardcoding category list | Fetch dynamically from database |
| Missing error handling | See error handling pattern in artifacts.go |
| Not tracking analytics | Include in MT-2.1, MT-2.4 implementations |
| Frontend API mocking problems | Use mock-server or stub endpoints |
| Search performance issues | Add proper indexes in MT-1.1 |
| Missing accessibility | Use semantic HTML, ARIA labels |
| Inconsistent styling | Follow Tailwind conventions, use components |
| Skipping tests | Won't pass Phase 4 verification |

---

## Key File References

### Backend Patterns
- **Handler Pattern**: `internal/handlers/artifacts.go`
- **Service Pattern**: `internal/services/` (existing services)
- **Repository Pattern**: `internal/repository/` (existing repos)
- **Error Handling**: Use `fmt.Errorf` with context
- **Logging**: Use `slog` with structured logging

### Frontend Patterns
- **Component Pattern**: `src/lib/components/` (existing components)
- **Store Pattern**: `src/lib/stores/` (existing stores)
- **API Integration**: `src/lib/api/` (existing client)
- **Styling**: Tailwind CSS + project classes
- **Icons**: Lucide Svelte only (no emoji)

---

## Dependency Critical Path

```
                   MT-1.3 (SQLC)
                       ↓
              ┌────────┴────────┐
              ↓                 ↓
         MT-2.x (API)    MT-4.x (Components)
              ↓                 ↓
         MT-3.x (Create)  MT-5.x (Modals)
              ↓                 ↓
         MT-3.3          MT-6.x (Search)
              ↓                 ↓
              └────────┬────────┘
                       ↓
                 MT-7.x (Tests)
                       ↓
                 MT-8.x (Docs & Deploy)
```

**Critical Path Duration**: 5-6 days (parallel) or 8 days (sequential)

---

## Go Back-of-Envelope Math

**Total Effort**: 41.5 hours
- Phase 1: 4h
- Phase 2: 12.5h
- Phase 3: 19h
- Phase 4: 6.5h
- Phase 5: 5.5h (excluding MT-6.4 optional)

**Team Sizing**:
- 1 Engineer: 8 days (41.5 hours ÷ 5 hours/day)
- 2 Engineers: 5-6 days (parallel tracks)
- 3 Engineers: 3-4 days (aggressive parallelization)

**Risk Buffer**: Add 20-30% for unknowns = 6-7 days (2 engineers), 9-10 days (1 engineer)

---

## Quality Gates

### Per Microtask
- [ ] Acceptance criteria met
- [ ] Code review passed
- [ ] No console errors/warnings
- [ ] Tests written and passing

### Phase Gates
- [ ] Phase 1: Database applies cleanly
- [ ] Phase 2: API endpoints tested
- [ ] Phase 3: Frontend renders without errors
- [ ] Phase 4: 90%+ coverage, no critical issues
- [ ] Phase 5: Documented and deployed

### Final Gates (Before Production)
- [ ] All 32 microtasks complete
- [ ] E2E tests passing
- [ ] Performance benchmarks met
- [ ] Security audit passed
- [ ] Staging deployment successful
- [ ] Documentation complete
- [ ] Team sign-off obtained

---

## Progress Tracking Template

```markdown
# Preset Gallery Progress

## Phase 1: Database
- [x] MT-1.1 - Database migration
- [ ] MT-1.2 - Analytics migration
- [ ] MT-1.3 - SQLC generation

## Phase 2A: API Core
- [ ] MT-2.1 - List endpoint
- [ ] MT-2.2 - Repository
- [ ] MT-2.3 - Service
- [ ] MT-2.4 - Detail endpoint
- [ ] MT-2.5 - Search endpoint

## Phase 2B: API Create
- [ ] MT-3.1 - Use preset
- [ ] MT-3.2 - Config validator
- [ ] MT-3.3 - Customize endpoint

## Phase 3A: Gallery
- [ ] MT-4.1 - Page
- [ ] MT-4.2 - Store
- [ ] MT-4.3 - Card
- [ ] MT-4.4 - Grid
- [ ] MT-4.5 - Sidebar

## Phase 3B: Modals
- [ ] MT-5.1 - Detail modal
- [ ] MT-5.2 - Use flow
- [ ] MT-5.3 - Customize
- [ ] MT-5.4 - Search integration

## Phase 3C: Search
- [ ] MT-6.1 - Search form
- [ ] MT-6.2 - Results
- [ ] MT-6.3 - Favorites
- [ ] MT-6.4 - Analytics

## Phase 4: Testing
- [ ] MT-7.1 - API tests
- [ ] MT-7.2 - Frontend tests
- [ ] MT-7.3 - E2E tests
- [ ] MT-7.4 - Audit

## Phase 5: Deploy
- [ ] MT-8.1 - API docs
- [ ] MT-8.2 - User docs
- [ ] MT-8.3 - Deploy guide
- [ ] MT-8.4 - Production

**Completed**: 0/32 (0%)
**In Progress**: 0/32
**Blocked**: 0/32
**Pending**: 32/32
```

---

## Quick Commands

```bash
# Database
psql -U user -d businessos -f migrations/037_agent_presets.sql
sqlc generate

# Backend Testing
go test ./internal/handlers -v -cover
go test ./internal/services -v -cover

# Frontend Development
npm run dev
npm run test
npm run test:e2e

# Deployment
docker build -t presets-api .
gcloud run deploy presets-api

# Coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

**Last Updated**: 2026-01-08
**Version**: 1.0
**Owner**: Development Team

