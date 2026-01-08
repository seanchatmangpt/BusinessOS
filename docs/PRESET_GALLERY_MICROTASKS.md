# Preset Gallery Feature - Microtask Breakdown

**Feature**: Add a gallery interface for browsing, searching, and creating agents from presets
**Complexity**: Full-Stack (Frontend + Backend + Database)
**Estimated Duration**: 5-8 days (1-4 hours per microtask)
**Target Sprint**: Q1 2026

---

## 1. DATABASE & SCHEMA (PHASE 1)

### MT-1.1: Create Database Migration for Presets Table
**Effort**: 1-2 hours | **Blocker**: None
**Description**: Create SQL migration for `agent_presets` table with all necessary fields
**Files**: `desktop/backend-go/internal/database/migrations/037_agent_presets.sql`
**Acceptance Criteria**:
- [ ] Table created with fields: id, name, description, category, config (JSONB), created_at, updated_at
- [ ] Proper indexes on category and created_at
- [ ] UUID primary key
- [ ] Migration applies cleanly to test DB

**Dependencies**: None
**Subtasks**:
- Define schema structure
- Add proper constraints and defaults
- Create migration file
- Test migration locally

---

### MT-1.2: Create Database Migration for Preset Analytics Table
**Effort**: 1 hour | **Blocker**: MT-1.1
**Description**: Create SQL migration for `preset_usage_analytics` table for tracking preset usage
**Files**: `desktop/backend-go/internal/database/migrations/038_preset_analytics.sql`
**Acceptance Criteria**:
- [ ] Table created with fields: id, preset_id, user_id, action_type (used/viewed), timestamp
- [ ] Foreign keys to presets and users
- [ ] Indexes for efficient querying
- [ ] Soft delete support (optional but recommended)

**Dependencies**: MT-1.1 (presets table must exist)
**Subtasks**:
- Create analytics table schema
- Add foreign key relationships
- Test migration

---

### MT-1.3: Generate SQLC Go Code from Migrations
**Effort**: 1 hour | **Blocker**: MT-1.2
**Description**: Run sqlc to generate Go query methods for presets and analytics
**Files**:
  - `desktop/backend-go/internal/database/sqlc/agent_presets.sql.go` (generated)
  - `desktop/backend-go/internal/database/sqlc/preset_analytics.sql.go` (generated)
**Acceptance Criteria**:
- [ ] SQLC generates without errors
- [ ] All CRUD operations available
- [ ] Proper error handling in generated code
- [ ] TypeScript types match Go structures (if applicable)

**Dependencies**: MT-1.2
**Subtasks**:
- Write .sql query files
- Run sqlc generator
- Verify generated code
- Add to git

---

## 2. BACKEND API - CORE ENDPOINTS (PHASE 2A)

### MT-2.1: Create Presets Handler with List Endpoint
**Effort**: 2 hours | **Blocker**: MT-1.3
**Description**: Implement HTTP handler for listing presets with filtering and pagination
**Files**: `desktop/backend-go/internal/handlers/presets.go`
**Acceptance Criteria**:
- [ ] GET /api/presets endpoint works
- [ ] Supports query params: category, search, limit, offset
- [ ] Returns paginated results
- [ ] Proper error handling with meaningful messages
- [ ] Logs requests with slog

**Dependencies**: MT-1.3
**Subtasks**:
- Create handler struct and methods
- Parse query parameters
- Validate inputs (search length, pagination bounds)
- Return proper JSON response with pagination metadata
- Add error handling

---

### MT-2.2: Create Presets Repository Layer
**Effort**: 2 hours | **Blocker**: MT-2.1
**Description**: Create repository methods for database queries on presets
**Files**: `desktop/backend-go/internal/repository/presets_repository.go`
**Acceptance Criteria**:
- [ ] ListPresets method with filtering
- [ ] GetPresetByID method
- [ ] Repository interface defined
- [ ] Proper error handling
- [ ] Context propagation in all methods

**Dependencies**: MT-2.1
**Subtasks**:
- Define repository interface
- Implement list with filters
- Implement get by ID
- Add context to all methods
- Write repository methods

---

### MT-2.3: Create Presets Service Layer
**Effort**: 2 hours | **Blocker**: MT-2.2
**Description**: Create service methods for business logic on presets
**Files**: `desktop/backend-go/internal/services/presets_service.go`
**Acceptance Criteria**:
- [ ] Service wraps repository
- [ ] Business logic for filtering/searching implemented
- [ ] Proper error handling and validation
- [ ] Caching of popular presets (optional, can defer)
- [ ] Analytics tracking calls

**Dependencies**: MT-2.2
**Subtasks**:
- Define service interface
- Wrap repository methods
- Add filtering logic
- Add validation
- Track usage analytics

---

### MT-2.4: Create GetPreset Endpoint (Detail View)
**Effort**: 1.5 hours | **Blocker**: MT-2.3
**Description**: Implement GET /api/presets/{id} endpoint for fetching single preset details
**Files**: `desktop/backend-go/internal/handlers/presets.go`
**Acceptance Criteria**:
- [ ] GET /api/presets/{id} endpoint works
- [ ] Returns detailed preset config
- [ ] Includes related metadata
- [ ] Proper 404 handling
- [ ] Tracks view analytics

**Dependencies**: MT-2.3
**Subtasks**:
- Extract ID from URL path
- Validate UUID format
- Call service to get preset
- Track analytics event
- Return response

---

### MT-2.5: Create Search Endpoint with Full-Text Search
**Effort**: 2 hours | **Blocker**: MT-2.4
**Description**: Implement advanced search with full-text capabilities
**Files**: `desktop/backend-go/internal/handlers/presets.go`
**Acceptance Criteria**:
- [ ] POST /api/presets/search endpoint works
- [ ] Full-text search on name and description
- [ ] Filter by category, tags
- [ ] Sort by relevance, popularity, date
- [ ] Returns relevant results ranked

**Dependencies**: MT-2.4
**Subtasks**:
- Define search request structure
- Implement full-text search logic
- Add filtering and sorting
- Implement relevance scoring
- Return ranked results

---

## 3. BACKEND API - CREATE FROM PRESET (PHASE 2B)

### MT-3.1: Create "Use Preset" Endpoint (Agent Creation)
**Effort**: 2.5 hours | **Blocker**: MT-2.5
**Description**: Implement POST /api/presets/{id}/use endpoint to create agent from preset
**Files**: `desktop/backend-go/internal/handlers/presets.go`
**Acceptance Criteria**:
- [ ] POST /api/presets/{id}/use endpoint works
- [ ] Creates new agent with preset config
- [ ] Assigns to current user/workspace
- [ ] Returns created agent details
- [ ] Validates preset exists
- [ ] Tracks usage analytics

**Dependencies**: MT-2.5
**Subtasks**:
- Validate preset exists
- Create agent from preset config
- Assign permissions
- Track analytics
- Return created agent

---

### MT-3.2: Create Preset Config Validation Service
**Effort**: 1.5 hours | **Blocker**: MT-3.1
**Description**: Create service to validate and transform preset configs before use
**Files**: `desktop/backend-go/internal/services/preset_config_validator.go`
**Acceptance Criteria**:
- [ ] Validates preset configuration schema
- [ ] Transforms legacy configs if needed
- [ ] Provides helpful error messages
- [ ] Merges user overrides with preset
- [ ] Handles version compatibility

**Dependencies**: MT-3.1
**Subtasks**:
- Define validation rules
- Implement schema validation
- Handle config transformations
- Merge user overrides
- Return validation errors

---

### MT-3.3: Create Preset Customization Endpoint
**Effort**: 2 hours | **Blocker**: MT-3.2
**Description**: Implement POST /api/presets/{id}/customize to return customizable config
**Files**: `desktop/backend-go/internal/handlers/presets.go`
**Acceptance Criteria**:
- [ ] Returns preset with customizable fields marked
- [ ] Shows default values and constraints
- [ ] Includes validation rules
- [ ] Supports partial customization
- [ ] Returns form schema for frontend

**Dependencies**: MT-3.2
**Subtasks**:
- Identify customizable fields
- Generate form schema
- Include constraints and validation
- Return to frontend
- Document API response

---

## 4. FRONTEND - GALLERY GRID & LAYOUT (PHASE 3A)

### MT-4.1: Create PresetGallery Page Component
**Effort**: 2 hours | **Blocker**: None
**Description**: Create main gallery page layout with grid and sidebar
**Files**: `frontend/src/routes/(app)/gallery/+page.svelte`
**Acceptance Criteria**:
- [ ] Page renders without errors
- [ ] Grid layout responsive (mobile, tablet, desktop)
- [ ] Category sidebar displays
- [ ] Search bar present and functional
- [ ] Loading states shown
- [ ] Error states handled

**Dependencies**: None (can run in parallel with backend)
**Subtasks**:
- Create page structure
- Build responsive grid layout
- Add sidebar for categories
- Add search input
- Handle loading/error states

---

### MT-4.2: Create Preset Gallery Store (Svelte)
**Effort**: 1.5 hours | **Blocker**: MT-4.1
**Description**: Create reactive store for managing preset gallery state
**Files**: `frontend/src/lib/stores/gallery.ts`
**Acceptance Criteria**:
- [ ] Store manages presets list
- [ ] Store manages filters (category, search)
- [ ] Store manages pagination
- [ ] Store manages loading/error states
- [ ] Proper type definitions
- [ ] Follows project conventions

**Dependencies**: MT-4.1
**Subtasks**:
- Define store types
- Create store with initial state
- Add filter management
- Add pagination
- Add loading states

---

### MT-4.3: Create PresetCard Component
**Effort**: 1.5 hours | **Blocker**: MT-4.2
**Description**: Create individual preset card component for gallery grid
**Files**: `frontend/src/lib/components/gallery/PresetCard.svelte`
**Acceptance Criteria**:
- [ ] Card displays preset info (name, description, category)
- [ ] Shows preview image/icon
- [ ] Displays category badge
- [ ] Shows usage count (optional)
- [ ] Click handler for details modal
- [ ] Hover states and animations
- [ ] No emojis, use Lucide icons

**Dependencies**: MT-4.2
**Subtasks**:
- Define card props
- Build card layout
- Add styling (Tailwind)
- Add icons from Lucide
- Add click/interaction handlers

---

### MT-4.4: Create Gallery Grid Component
**Effort**: 1.5 hours | **Blocker**: MT-4.3
**Description**: Create responsive grid wrapper for displaying preset cards
**Files**: `frontend/src/lib/components/gallery/PresetGrid.svelte`
**Acceptance Criteria**:
- [ ] Renders cards in responsive grid
- [ ] Works on mobile (1 column), tablet (2-3), desktop (3-4)
- [ ] Smooth animations between states
- [ ] Loading skeleton state
- [ ] Empty state message
- [ ] Pagination controls

**Dependencies**: MT-4.3
**Subtasks**:
- Create grid layout
- Add responsive breakpoints
- Add skeleton loading
- Add empty states
- Add pagination controls

---

### MT-4.5: Create Category Sidebar Component
**Effort**: 1.5 hours | **Blocker**: MT-4.4
**Description**: Create sidebar for filtering by category
**Files**: `frontend/src/lib/components/gallery/CategorySidebar.svelte`
**Acceptance Criteria**:
- [ ] Lists all categories
- [ ] Shows preset count per category
- [ ] Category selection works
- [ ] Active category highlighted
- [ ] Smooth transitions
- [ ] Mobile collapse/expand

**Dependencies**: MT-4.4
**Subtasks**:
- Create category list
- Add count display
- Add selection logic
- Add highlighting
- Add mobile responsiveness

---

## 5. FRONTEND - DETAIL MODAL & INTERACTIONS (PHASE 3B)

### MT-5.1: Create PresetDetailModal Component
**Effort**: 2.5 hours | **Blocker**: MT-4.5
**Description**: Create modal for viewing detailed preset information
**Files**: `frontend/src/lib/components/gallery/PresetDetailModal.svelte`
**Acceptance Criteria**:
- [ ] Modal displays full preset details
- [ ] Shows description, config overview, usage stats
- [ ] Displays category, created date, author info
- [ ] Shows usage tips/documentation
- [ ] "Use Preset" and "Customize" buttons present
- [ ] Close button works
- [ ] Accessible (keyboard navigation)

**Dependencies**: MT-4.5
**Subtasks**:
- Create modal structure
- Add detail sections
- Add buttons
- Add accessibility features
- Style with Tailwind

---

### MT-5.2: Create Use Preset Flow
**Effort**: 2 hours | **Blocker**: MT-5.1, MT-3.1
**Description**: Implement flow for creating agent from preset (modal -> API -> success)
**Files**: `frontend/src/lib/components/gallery/UsePresetFlow.svelte`
**Acceptance Criteria**:
- [ ] "Use Preset" button triggers flow
- [ ] Shows confirmation modal
- [ ] Submits API request
- [ ] Shows loading state
- [ ] Shows success message
- [ ] Redirects to new agent
- [ ] Error handling with retry

**Dependencies**: MT-5.1, MT-3.1
**Subtasks**:
- Create confirmation modal
- Add API integration
- Add loading states
- Add success handling
- Add error handling
- Add redirect logic

---

### MT-5.3: Create Customize Preset Modal
**Effort**: 2.5 hours | **Blocker**: MT-5.2, MT-3.3
**Description**: Implement modal for customizing preset before creating agent
**Files**: `frontend/src/lib/components/gallery/CustomizePresetModal.svelte`
**Acceptance Criteria**:
- [ ] Modal displays form with customizable fields
- [ ] Form fields generated from preset schema
- [ ] Shows field descriptions and constraints
- [ ] Form validation on client
- [ ] Submit button creates agent with custom config
- [ ] Cancel button closes modal
- [ ] Proper error messages

**Dependencies**: MT-5.2, MT-3.3
**Subtasks**:
- Create form from schema
- Add field validation
- Add submit handler
- Handle API response
- Show errors

---

### MT-5.4: Create Search & Filter Integration
**Effort**: 1.5 hours | **Blocker**: MT-5.3
**Description**: Integrate search and filter functionality with gallery
**Files**: `frontend/src/routes/(app)/gallery/+page.svelte`
**Acceptance Criteria**:
- [ ] Search input filters presets in real-time
- [ ] Category filter works
- [ ] Filters combined correctly (AND logic)
- [ ] Results update smoothly
- [ ] Pagination resets on filter change
- [ ] URL params updated with filters

**Dependencies**: MT-5.3
**Subtasks**:
- Add search input handler
- Add category filter handler
- Combine filter logic
- Update gallery with filters
- Update URL params

---

## 6. FRONTEND - SEARCH & ADVANCED FEATURES (PHASE 3C)

### MT-6.1: Create Advanced Search Form
**Effort**: 2 hours | **Blocker**: MT-5.4
**Description**: Create advanced search interface with multiple filter options
**Files**: `frontend/src/lib/components/gallery/AdvancedSearch.svelte`
**Acceptance Criteria**:
- [ ] Search by name/description
- [ ] Filter by category (multi-select)
- [ ] Filter by tags (if applicable)
- [ ] Sort options (relevance, popularity, date)
- [ ] Clear filters button
- [ ] Search results count shown
- [ ] Keyboard shortcuts for search

**Dependencies**: MT-5.4
**Subtasks**:
- Create search form layout
- Add filter inputs
- Add sort options
- Add clear button
- Add submit logic

---

### MT-6.2: Create Search Results Component
**Effort**: 1.5 hours | **Blocker**: MT-6.1
**Description**: Create component for displaying search results with relevance highlighting
**Files**: `frontend/src/lib/components/gallery/SearchResults.svelte`
**Acceptance Criteria**:
- [ ] Displays search results
- [ ] Highlights matching terms
- [ ] Shows relevance score/badge
- [ ] Sorts by relevance
- [ ] Pagination works
- [ ] Clear results option
- [ ] No results message

**Dependencies**: MT-6.1
**Subtasks**:
- Create results layout
- Add highlighting logic
- Add relevance display
- Add pagination
- Add empty state

---

### MT-6.3: Create Preset Favorites/Bookmarks
**Effort**: 1.5 hours | **Blocker**: MT-6.2
**Description**: Add ability to save presets to favorites for quick access
**Files**:
  - `frontend/src/lib/components/gallery/FavoritesButton.svelte`
  - `frontend/src/lib/stores/favorites.ts`
**Acceptance Criteria**:
- [ ] Star/heart button to favorite preset
- [ ] Favorites list accessible
- [ ] Persists across sessions (localStorage or backend)
- [ ] Favorites count displayed
- [ ] Filter by favorites option
- [ ] Visual feedback on favorite

**Dependencies**: MT-6.2
**Subtasks**:
- Create favorite button component
- Create favorites store
- Add persistence
- Add filter option
- Add visual feedback

---

### MT-6.4: Create Usage Analytics Dashboard (Optional)
**Effort**: 2 hours | **Blocker**: MT-6.3
**Description**: Create dashboard showing preset usage statistics
**Files**: `frontend/src/routes/(app)/gallery/analytics/+page.svelte`
**Acceptance Criteria**:
- [ ] Shows popular presets
- [ ] Shows recent usage trends
- [ ] Shows category popularity
- [ ] Shows user favorites
- [ ] Responsive charts/graphs
- [ ] Export data option (optional)

**Dependencies**: MT-6.3 (optional, can defer)
**Subtasks**:
- Create analytics page
- Add chart components
- Fetch analytics data
- Display statistics
- Add export (optional)

---

## 7. INTEGRATION & TESTING (PHASE 4)

### MT-7.1: Integration Test - API Endpoints
**Effort**: 2 hours | **Blocker**: MT-3.3 (all API endpoints complete)
**Description**: Write integration tests for all preset API endpoints
**Files**: `desktop/backend-go/internal/handlers/presets_test.go`
**Acceptance Criteria**:
- [ ] Tests for GET /api/presets (list, filters, pagination)
- [ ] Tests for GET /api/presets/{id}
- [ ] Tests for POST /api/presets/search
- [ ] Tests for POST /api/presets/{id}/use
- [ ] Tests for POST /api/presets/{id}/customize
- [ ] Error handling tests (404, invalid input)
- [ ] 90%+ coverage

**Dependencies**: MT-3.3
**Subtasks**:
- Create test file
- Write list endpoint tests
- Write detail endpoint tests
- Write search tests
- Write create from preset tests
- Add edge cases

---

### MT-7.2: Frontend Integration Tests
**Effort**: 2 hours | **Blocker**: MT-6.3
**Description**: Write Vitest/SvelteKit tests for frontend components
**Files**:
  - `frontend/src/routes/(app)/gallery/__tests__/gallery.test.ts`
  - `frontend/src/lib/components/gallery/__tests__/*.test.ts`
**Acceptance Criteria**:
- [ ] Gallery page renders
- [ ] PresetCard renders with props
- [ ] Search filters work
- [ ] Detail modal opens/closes
- [ ] Use preset flow works
- [ ] Error states handled
- [ ] 80%+ coverage

**Dependencies**: MT-6.3
**Subtasks**:
- Create test suite
- Write component tests
- Write interaction tests
- Write store tests
- Add assertions

---

### MT-7.3: E2E Tests - User Flows
**Effort**: 2.5 hours | **Blocker**: MT-7.2
**Description**: Write end-to-end tests for complete user workflows
**Files**: `e2e/tests/preset-gallery.spec.ts` (or Playwright/Cypress equivalent)
**Acceptance Criteria**:
- [ ] Test: Browse gallery and view details
- [ ] Test: Search presets
- [ ] Test: Create agent from preset
- [ ] Test: Customize and create agent
- [ ] Test: Favorite presets
- [ ] All happy paths covered
- [ ] Cross-browser compatibility

**Dependencies**: MT-7.2
**Subtasks**:
- Create E2E test file
- Write browsing flow test
- Write search flow test
- Write creation flow test
- Write customization flow test

---

### MT-7.4: Performance & Security Audit
**Effort**: 2 hours | **Blocker**: MT-7.3
**Description**: Audit performance and security of gallery feature
**Files**: Documentation or code comments
**Acceptance Criteria**:
- [ ] Query performance tested (response times < 200ms)
- [ ] Large dataset performance tested (1000+ presets)
- [ ] Input validation on all endpoints
- [ ] SQL injection prevention verified
- [ ] XSS prevention verified
- [ ] CORS headers correct
- [ ] Rate limiting applied (if needed)
- [ ] Authorization checks in place

**Dependencies**: MT-7.3
**Subtasks**:
- Run performance tests
- Check query performance
- Verify security measures
- Test with large datasets
- Audit access controls

---

## 8. DEPLOYMENT & DOCUMENTATION (PHASE 5)

### MT-8.1: Create API Documentation
**Effort**: 1.5 hours | **Blocker**: MT-7.4
**Description**: Document all preset API endpoints in OpenAPI/Swagger format
**Files**: `desktop/backend-go/docs/preset_api.md` or OpenAPI spec
**Acceptance Criteria**:
- [ ] All endpoints documented
- [ ] Request/response examples included
- [ ] Error codes documented
- [ ] Authentication requirements clear
- [ ] Rate limits documented
- [ ] Published to /docs endpoint

**Dependencies**: MT-7.4
**Subtasks**:
- Write endpoint descriptions
- Add request/response examples
- Document error codes
- Create OpenAPI spec (optional)
- Publish documentation

---

### MT-8.2: Create User Documentation
**Effort**: 1.5 hours | **Blocker**: MT-8.1
**Description**: Create user-facing documentation for Preset Gallery feature
**Files**: `docs/PRESET_GALLERY_GUIDE.md`
**Acceptance Criteria**:
- [ ] Feature overview
- [ ] How to browse gallery
- [ ] How to search presets
- [ ] How to create agent from preset
- [ ] How to customize presets
- [ ] Screenshots/GIFs included
- [ ] Troubleshooting section

**Dependencies**: MT-8.1
**Subtasks**:
- Write feature overview
- Document browsing process
- Document search process
- Document creation process
- Add troubleshooting
- Add screenshots

---

### MT-8.3: Create Migration & Deployment Guide
**Effort**: 1 hour | **Blocker**: MT-8.2
**Description**: Create guide for deploying preset gallery to production
**Files**: `docs/PRESET_GALLERY_DEPLOYMENT.md`
**Acceptance Criteria**:
- [ ] Database migration steps
- [ ] Seed data import process (if applicable)
- [ ] Environment variable setup
- [ ] Feature flag activation (if needed)
- [ ] Rollback procedure
- [ ] Monitoring setup

**Dependencies**: MT-8.2
**Subtasks**:
- Document migration steps
- Document seed data
- Document environment setup
- Document deployment steps
- Document rollback process

---

### MT-8.4: Deploy to Staging & Production
**Effort**: 1.5 hours | **Blocker**: MT-8.3
**Description**: Deploy preset gallery feature to staging and production
**Files**: None (deployment process)
**Acceptance Criteria**:
- [ ] Deployed to staging environment
- [ ] Smoke tests pass
- [ ] User acceptance testing approved
- [ ] Deployed to production
- [ ] Monitoring alerts configured
- [ ] Rollback verified ready

**Dependencies**: MT-8.3
**Subtasks**:
- Deploy to staging
- Run smoke tests
- Get UAT approval
- Deploy to production
- Configure monitoring
- Verify rollback ready

---

## DEPENDENCY GRAPH

```
MT-1.1 (Presets Migration)
  ↓
MT-1.2 (Analytics Migration) → MT-1.3 (SQLC)
  ↓                              ↓
  └──────────────────────────────┘
  ↓
MT-2.1 → MT-2.2 → MT-2.3 → MT-2.4 → MT-2.5
  ↓
MT-3.1 → MT-3.2 → MT-3.3
  ↓
[All API complete]
  ↓
MT-4.1 → MT-4.2 → MT-4.3 → MT-4.4 → MT-4.5
  ↓
MT-5.1 → MT-5.2 → MT-5.3 → MT-5.4
  ↓
MT-6.1 → MT-6.2 → MT-6.3 → MT-6.4 (optional)
  ↓
MT-7.1 (API tests) + MT-7.2 (Frontend tests)
  ↓
MT-7.3 (E2E tests)
  ↓
MT-7.4 (Audit)
  ↓
MT-8.1 → MT-8.2 → MT-8.3 → MT-8.4
```

---

## PARALLEL EXECUTION TRACKS

### Track A: Backend Database & API (2-3 days)
- MT-1.1 → MT-1.2 → MT-1.3 (Database layer - sequential)
- MT-2.1 → MT-2.2 → MT-2.3 → MT-2.4 → MT-2.5 (API endpoints - sequential)
- MT-3.1 → MT-3.2 → MT-3.3 (Create from preset - sequential)
- Can run **parallel** with Track B

### Track B: Frontend Gallery & Interactions (2-3 days)
- MT-4.1 → MT-4.2 → MT-4.3 → MT-4.4 → MT-4.5 (Gallery layout - mostly sequential)
- MT-5.1 → MT-5.2 → MT-5.3 → MT-5.4 (Modals & interactions - sequential)
- MT-6.1 → MT-6.2 → MT-6.3 → MT-6.4 (Search & advanced - sequential)
- Can run **parallel** with Track A

### Track C: Testing & Deployment (1-2 days)
- MT-7.1 (API tests) - **parallel** with MT-7.2
- MT-7.2 (Frontend tests) - **parallel** with MT-7.1
- MT-7.3 (E2E tests) - sequential after MT-7.1 & MT-7.2
- MT-7.4 (Audit) - sequential after MT-7.3
- **BLOCKING** all testing tracks

### Track D: Documentation & Deploy (1 day)
- MT-8.1 (API Docs) - parallel with MT-8.2
- MT-8.2 (User Docs) - parallel with MT-8.1
- MT-8.3 (Deployment Guide) - sequential after docs
- MT-8.4 (Deploy) - sequential after guide, final step

---

## ESTIMATED TIMELINE

```
Week 1 (Days 1-3):
  Mon: MT-1.1, MT-1.2, MT-1.3, MT-4.1, MT-4.2 (parallel)
  Tue: MT-2.1, MT-2.2, MT-4.3, MT-4.4 (parallel)
  Wed: MT-2.3, MT-2.4, MT-4.5, MT-5.1 (parallel)

Week 1 (Days 4-5):
  Thu: MT-2.5, MT-3.1, MT-5.2 (parallel)
  Fri: MT-3.2, MT-3.3, MT-5.3, MT-6.1 (parallel)

Week 2 (Days 6-7):
  Mon: MT-5.4, MT-6.2 (parallel)
  Tue: MT-6.3, MT-6.4 (optional, parallel)

Week 2 (Days 8-9):
  Wed: MT-7.1, MT-7.2 (parallel)
  Thu: MT-7.3, MT-7.4 (sequential)

Week 2 (Day 10):
  Fri: MT-8.1, MT-8.2, MT-8.3, MT-8.4 (parallel then sequential)
```

---

## DEPENDENCY MATRIX

| Microtask | Depends On | Blocked By | Can Run In Parallel With |
|-----------|-----------|-----------|-------------------------|
| MT-1.1 | None | None | All |
| MT-1.2 | MT-1.1 | MT-1.1 | All except MT-1.1 |
| MT-1.3 | MT-1.2 | MT-1.2 | B track |
| MT-2.1 | MT-1.3 | MT-1.3 | B track |
| MT-2.2 | MT-2.1 | MT-2.1 | B track |
| MT-2.3 | MT-2.2 | MT-2.2 | B track |
| MT-2.4 | MT-2.3 | MT-2.3 | B track |
| MT-2.5 | MT-2.4 | MT-2.4 | B track |
| MT-3.1 | MT-2.5 | MT-2.5 | B track (MT-1.3 complete) |
| MT-3.2 | MT-3.1 | MT-3.1 | B track |
| MT-3.3 | MT-3.2 | MT-3.2 | B track |
| MT-4.1 | None | None | A track |
| MT-4.2 | MT-4.1 | MT-4.1 | A track |
| MT-4.3 | MT-4.2 | MT-4.2 | A track |
| MT-4.4 | MT-4.3 | MT-4.3 | A track |
| MT-4.5 | MT-4.4 | MT-4.4 | A track |
| MT-5.1 | MT-4.5 | MT-4.5 | A track |
| MT-5.2 | MT-5.1, MT-3.1 | MT-3.1 | A track |
| MT-5.3 | MT-5.2, MT-3.3 | MT-3.3 | A track |
| MT-5.4 | MT-5.3 | MT-5.3 | A track |
| MT-6.1 | MT-5.4 | MT-5.4 | A track |
| MT-6.2 | MT-6.1 | MT-6.1 | A track |
| MT-6.3 | MT-6.2 | MT-6.2 | A track |
| MT-6.4 | MT-6.3 | MT-6.3 (optional) | A track |
| MT-7.1 | MT-3.3 | MT-3.3 | MT-7.2 |
| MT-7.2 | MT-6.3 | MT-6.3 | MT-7.1 |
| MT-7.3 | MT-7.1, MT-7.2 | Both | None |
| MT-7.4 | MT-7.3 | MT-7.3 | None |
| MT-8.1 | MT-7.4 | MT-7.4 | MT-8.2 |
| MT-8.2 | MT-7.4 | MT-7.4 | MT-8.1 |
| MT-8.3 | MT-8.1, MT-8.2 | Both | None |
| MT-8.4 | MT-8.3 | MT-8.3 | None |

---

## SUCCESS CRITERIA

- [ ] All 32 microtasks completed
- [ ] All tests passing (90%+ API coverage, 80%+ Frontend coverage)
- [ ] E2E tests cover all user flows
- [ ] No performance regressions
- [ ] Security audit passed
- [ ] Documentation complete and published
- [ ] Deployed to production successfully
- [ ] No critical bugs in first week of production
- [ ] User acceptance testing passed
- [ ] Feature monitoring configured

---

## NOTES

1. **Parallelization Strategy**: Tracks A and B can run fully in parallel (backend and frontend teams)
2. **Testing Critical Path**: MT-7.1, 7.2, 7.3, 7.4 are critical and cannot be skipped
3. **Optional Features**: MT-6.4 (Analytics) can be deferred to post-launch if needed
4. **Frontend First Possible**: Frontend can start immediately without waiting for backend (mock data)
5. **Incremental Deployment**: Can deploy frontend before backend APIs are 100% complete with feature flags

