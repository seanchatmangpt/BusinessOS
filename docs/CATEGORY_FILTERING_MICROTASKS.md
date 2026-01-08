# Category Filtering Feature - Atomic Microtasks Breakdown

**Feature:** Add category-based filtering to artifact/content lists
**Estimated Total Time:** 24-32 hours (spread across 6-8 microtasks)
**Priority:** Medium
**Complexity:** Moderate
**Dependencies:** Existing artifact system, query params, stores

---

## 📊 Microtasks Overview

| Task | Est. Time | Dependencies | Status |
|------|-----------|--------------|--------|
| MT-1 | Create filter state store | 2-3h | None | Pending |
| MT-2 | Build category filter component | 2-3h | MT-1 | Pending |
| MT-3 | Integrate filters with API calls | 3-4h | MT-2 | Pending |
| MT-4 | Add URL query param persistence | 2-3h | MT-3 | Pending |
| MT-5 | Create UI badge components | 2-3h | MT-1 | Pending |
| MT-6 | Implement clear filters feature | 1-2h | MT-2, MT-5 | Pending |
| MT-7 | Add filter animation & UX polish | 2-3h | MT-2, MT-5 | Pending |
| MT-8 | Integration tests & verification | 3-4h | All others | Pending |

---

## 📝 Detailed Microtasks

### MT-1: Create Filter State Store (2-3 hours)
**Goal:** Set up Svelte store for managing filter state
**Type:** Frontend - State Management
**Files to Create/Modify:**
- Create: `frontend/src/lib/stores/filterStore.ts`

**Requirements:**
- [x] Define `FilterState` interface with:
  - `selectedCategories: string[]`
  - `isOpen: boolean`
  - `appliedFilters: Record<string, boolean>`
- [x] Create writable stores:
  - `filterState` - main filter state
  - `filteredCount` - derived store showing count
- [x] Implement helper functions:
  - `toggleCategory(cat: string)`
  - `clearFilters()`
  - `setCategories(cats: string[])`
  - `getAppliedFilters()` - returns query string format
- [x] Add TypeScript strict mode support
- [x] Document store API in JSDoc comments

**Acceptance Criteria:**
- Store compiles without TypeScript errors
- All functions have proper typing
- Store is reactive (derived stores update correctly)
- Can toggle categories on/off
- Can clear all filters at once

**Dependencies:** None
**Blockers:** None

---

### MT-2: Build Category Filter Component (2-3 hours)
**Goal:** Create reusable filter component with category selection UI
**Type:** Frontend - Component
**Files to Create/Modify:**
- Create: `frontend/src/lib/components/filters/CategoryFilter.svelte`
- Create: `frontend/src/lib/components/filters/index.ts`

**Requirements:**
- [x] Component props:
  - `categories: string[]` - available categories
  - `selected: string[]` - selected categories
  - `onSelect: (cat: string) => void`
  - `onClear: () => void`
  - `disabled: boolean = false`
  - `maxHeight: string = 'auto'`
- [x] UI elements:
  - Dropdown/collapsible filter panel
  - Checkbox list for each category
  - "Select All" / "Clear All" buttons
  - Search/filter box for category search
  - Selected count indicator
- [x] Styling:
  - Use Tailwind CSS (match existing BusinessOS theme)
  - Dark mode support
  - Responsive layout (mobile-friendly)
- [x] Accessibility:
  - Proper ARIA labels
  - Keyboard navigation
  - Screen reader support

**Acceptance Criteria:**
- Component renders without errors
- All categories display correctly
- Selection/deselection works
- "Clear All" clears selections
- Component is keyboard navigable
- Matches BusinessOS design system

**Dependencies:** MT-1 (filterStore)
**Blockers:** None

---

### MT-3: Integrate Filters with API Calls (3-4 hours)
**Goal:** Connect filter state to API queries and handle filtered results
**Type:** Frontend - API Integration
**Files to Create/Modify:**
- Modify: `frontend/src/lib/api/*.ts` (artifact/content API)
- Create: `frontend/src/lib/api/filterUtils.ts`

**Requirements:**
- [x] Create filter utility functions:
  - `buildFilterQuery(filters: Record<string, boolean>): string`
  - `parseFilterFromQuery(query: string): Record<string, boolean>`
  - `buildQueryParams(filters, searchTerm, pagination)`
- [x] Modify existing API calls:
  - `listArtifacts()` - add `categories` parameter
  - `searchArtifacts()` - add `categories` parameter
  - Any other list endpoints
- [x] Implement request debouncing:
  - Debounce API calls when filters change (300ms)
  - Show loading state during filter application
- [x] Handle API responses:
  - Cache filtered results
  - Show empty state when no results
  - Display result count

**API Endpoint Changes (Backend needed):**
```
GET /api/artifacts?categories=cat1,cat2&search=...&page=1
GET /api/search?categories=cat1&query=...
```

**Acceptance Criteria:**
- API calls include category filters
- Filters debounce properly
- Loading state shows during requests
- Results update when filters change
- Empty states display correctly
- Error handling for failed requests

**Dependencies:** MT-1, MT-2
**Blockers:** Backend API must support `categories` parameter

---

### MT-4: Add URL Query Parameter Persistence (2-3 hours)
**Goal:** Make filters shareable and persistent via URL query params
**Type:** Frontend - URL Management
**Files to Create/Modify:**
- Modify: `frontend/src/routes/+page.ts` or `+page.server.ts`
- Modify: Component that uses filters

**Requirements:**
- [x] SvelteKit `page` store integration:
  - Read `$page.url.searchParams` on page load
  - Initialize filter state from URL
- [x] Implement bidirectional sync:
  - When filters change → update URL (without page reload)
  - When URL changes → update filter state
- [x] Use SvelteKit navigation:
  - `goto()` with URL updates
  - History API integration (browser back/forward)
- [x] Query param format:
  ```
  ?categories=category1,category2&search=term&page=1
  ```

**Acceptance Criteria:**
- Filters persist when user shares URL
- Back/forward buttons work correctly
- Page load restores previous filters
- URL updates as filters change
- No page refresh on filter change
- Handles multiple filters in URL

**Dependencies:** MT-2, MT-3
**Blockers:** None

---

### MT-5: Create UI Badge Components (2-3 hours)
**Goal:** Build visual badges for displaying active filters
**Type:** Frontend - UI Components
**Files to Create/Modify:**
- Create: `frontend/src/lib/components/filters/FilterBadge.svelte`
- Create: `frontend/src/lib/components/filters/FilterBadgeGroup.svelte`

**Requirements:**
- [x] FilterBadge component:
  - Display category name
  - Show remove (X) button
  - Callback on remove click
  - Color coding for different categories
  - Icon support
- [x] FilterBadgeGroup component:
  - Display multiple badges
  - Horizontal scrollable layout
  - "Clear All" button
  - Count indicator
- [x] Styling:
  - Use Tailwind badge classes
  - Color palette for categories
  - Hover effects
  - Animation on add/remove
- [x] Responsive:
  - Stack on mobile
  - Scrollable on desktop
  - Truncate long names

**Acceptance Criteria:**
- Badges render correctly
- Remove buttons work
- All styling looks professional
- Responsive on mobile/tablet/desktop
- Colors match design system
- Smooth animations

**Dependencies:** MT-1
**Blockers:** None

---

### MT-6: Implement Clear Filters Feature (1-2 hours)
**Goal:** Add "Clear All Filters" functionality with confirmation
**Type:** Frontend - Feature
**Files to Create/Modify:**
- Modify: `frontend/src/lib/components/filters/CategoryFilter.svelte`
- Modify: Any component using FilterBadgeGroup

**Requirements:**
- [x] Clear button with icon
- [x] Optional confirmation modal:
  - Show count of filters being cleared
  - Confirm/Cancel buttons
  - Animation
- [x] Reset behavior:
  - Clear selected categories
  - Clear search term (optional)
  - Reset pagination to page 1
  - Update URL to remove query params
- [x] Feedback:
  - Toast notification (optional)
  - Visual feedback that filters cleared
  - Results update immediately

**Acceptance Criteria:**
- Clear button removes all filters
- Confirmation modal works (if enabled)
- URL updates correctly
- Results refresh
- State resets completely
- Feedback shown to user

**Dependencies:** MT-2, MT-5
**Blockers:** None

---

### MT-7: Add Filter Animation & UX Polish (2-3 hours)
**Goal:** Enhance UX with animations and polish interactions
**Type:** Frontend - UX/Animation
**Files to Create/Modify:**
- Modify: All filter components
- Create: `frontend/src/lib/animations/filterAnimations.ts`

**Requirements:**
- [x] Animations:
  - Badge entrance/exit (fade + slide)
  - Filter panel open/close
  - Result count change (number animation)
  - Loading skeleton
- [x] Transitions:
  - Smooth filter application
  - Progressive result loading
  - Category selection feedback
- [x] Polish:
  - Disable filter when max selection reached (optional)
  - Show "No categories match search" state
  - Truncate long category names with tooltips
  - Keyboard shortcuts (ESC to close)
- [x] Loading states:
  - Show spinner while filtering
  - Skeleton loading for results
  - Disabled state during request

**Acceptance Criteria:**
- Animations are smooth (60fps)
- No jarring transitions
- Loading states clear
- Interactions feel responsive
- Polish matches design system
- Performance maintained

**Dependencies:** MT-2, MT-5, MT-6
**Blockers:** None

---

### MT-8: Integration Tests & Verification (3-4 hours)
**Goal:** Test all filter functionality end-to-end
**Type:** Testing & QA
**Files to Create/Modify:**
- Create: `frontend/src/lib/components/filters/__tests__/CategoryFilter.test.ts`
- Create: `frontend/src/lib/__tests__/stores/filterStore.test.ts`
- Create: `frontend/src/__tests__/integration/filtering.test.ts`

**Requirements:**
- [x] Unit tests:
  - Filter store operations
  - Query param parsing/building
  - Filter state transitions
- [x] Component tests:
  - CategoryFilter renders correctly
  - Selection/deselection works
  - Clear All works
  - Keyboard navigation
- [x] Integration tests:
  - E2E flow: select filter → API call → results update
  - URL persistence and restoration
  - Back/forward button functionality
  - Multiple filters together
- [x] Verification tests:
  - Filter count accuracy
  - API payload validation
  - Empty state handling
  - Error state handling

**Test Coverage:**
- Store: 100% of functions
- Components: 80%+ of UI paths
- Integration: Main user flows

**Acceptance Criteria:**
- All tests pass
- Coverage > 80%
- E2E flows work
- No console errors
- Performance acceptable
- Ready for production

**Dependencies:** MT-1 through MT-7
**Blockers:** None

---

## 📋 Backend Requirements Summary

For filters to work end-to-end, backend needs to:

### API Changes Required
1. **Modify existing list endpoints** to accept `categories` parameter:
   ```
   GET /api/artifacts?categories=cat1,cat2&search=...
   GET /api/search?categories=cat1&query=...
   ```

2. **Add category filtering in SQL queries:**
   - Filter artifacts by `artifact.category` or `artifact.categories` (JSON)
   - Or create `artifact_categories` junction table if not exists

3. **Return category info with results:**
   ```json
   {
     "id": "...",
     "title": "...",
     "category": "Research",
     "categories": ["Research", "AI"]
   }
   ```

4. **Add category listing endpoint:**
   ```
   GET /api/categories
   Returns: ["Research", "Document", "Code", ...]
   ```

---

## 🔗 Cross-Cutting Concerns

### State Management
- Filter state in Svelte store
- Synchronized with URL query params
- Debounced API calls
- Loading state management

### Performance
- Debounce filter changes (300ms)
- Virtual scrolling for long category lists (if needed)
- Result caching
- Lazy load categories

### Accessibility
- ARIA labels on all interactive elements
- Keyboard navigation (Tab, Enter, Escape)
- Screen reader support
- Focus management

### Error Handling
- Network errors show retry button
- Invalid filters gracefully degrade
- Empty states with helpful messages
- API errors logged and displayed

---

## 🎯 Success Criteria (Overall)

- [x] User can filter artifacts by multiple categories
- [x] Selected filters persist in URL
- [x] Filters update results in real-time
- [x] "Clear All" removes all filters at once
- [x] Badges display active filters
- [x] Smooth animations and transitions
- [x] Mobile-friendly UI
- [x] Keyboard accessible
- [x] All tests pass
- [x] No console errors
- [x] TypeScript strict mode
- [x] Matches BusinessOS design

---

## 📊 Task Dependencies Graph

```
MT-1 (Filter Store)
  ├── MT-2 (Filter Component)
  │   ├── MT-3 (API Integration)
  │   │   ├── MT-4 (URL Params)
  │   │   │   └── MT-7 (Polish)
  │   │   │       └── MT-8 (Tests)
  │   └── MT-6 (Clear Filters)
  │       └── MT-7 (Polish)
  │           └── MT-8 (Tests)
  │
  └── MT-5 (Badge Components)
      ├── MT-6 (Clear Filters)
      │   └── MT-7 (Polish)
      │       └── MT-8 (Tests)
      └── MT-7 (Polish)
          └── MT-8 (Tests)
```

**Parallel Opportunities:**
- MT-2, MT-3, MT-5 can be done in parallel (all depend only on MT-1)
- MT-4 can start once MT-3 is done
- MT-6 can start once MT-2 and MT-5 are done
- MT-7 can start once MT-2, MT-5, MT-6 are done
- MT-8 must be last (depends on all others)

---

## 📈 Execution Strategy

**Recommended Order:**
1. Start MT-1 first (foundation)
2. Parallel: MT-2, MT-3, MT-5 (all start when MT-1 done)
3. Start MT-4 when MT-3 complete
4. Start MT-6 when MT-2 and MT-5 complete
5. Start MT-7 when MT-2, MT-5, MT-6 complete
6. Start MT-8 when all others complete

**Total Timeline with Parallelization:**
- Sequential approach: 24-32 hours
- With parallelization: ~16-20 hours of actual work time

---

## 🚀 Quick Reference

### Microtask Cards

**MT-1** | Filter State Store
- Duration: 2-3h | Store setup | Svelte stores, TypeScript
- Output: Fully typed, reactive filter store

**MT-2** | Category Filter Component
- Duration: 2-3h | Component | Svelte, Tailwind, a11y
- Output: Reusable filter dropdown with checkboxes

**MT-3** | API Integration
- Duration: 3-4h | Integration | API calls, debouncing
- Output: Connected filters that call API and show results

**MT-4** | URL Query Params
- Duration: 2-3h | URL Management | SvelteKit navigation
- Output: Shareable filter URLs, history integration

**MT-5** | Badge Components
- Duration: 2-3h | UI Components | Svelte, Tailwind
- Output: Active filter badges with remove buttons

**MT-6** | Clear Filters
- Duration: 1-2h | Feature | State reset, modal
- Output: Clear All button with optional confirmation

**MT-7** | Animation & Polish
- Duration: 2-3h | UX | Animations, transitions
- Output: Smooth, professional interactions

**MT-8** | Tests & Verification
- Duration: 3-4h | Testing | Unit, component, e2e tests
- Output: 80%+ test coverage, ready for production

---

**Last Updated:** 2026-01-08
**Status:** Ready for Sprint Planning
**Owner:** Frontend Team
**Reviewer:** Code Review Team
