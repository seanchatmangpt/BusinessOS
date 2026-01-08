# Category Filtering - Quick Start Card

## All 6 Microtasks at a Glance

| ID | Task | Time | What | Key Output |
|---|---|---|---|---|
| **MT-1** | Filter Store | 2-3h | Svelte store for filter state | `filterStore.ts` |
| **MT-2** | Filter Component | 2-3h | Dropdown with checkboxes | `CategoryFilter.svelte` |
| **MT-3** | API Integration | 3-4h | Connect filters to API | Query params + debouncing |
| **MT-4** | URL Persistence | 2-3h | URL `?categories=...` | History + bookmarkable |
| **MT-5** | Badges | 2-3h | Show active filters | `FilterBadge.svelte` |
| **MT-6** | Clear Button | 1-2h | Remove all filters | Modal + confirmation |
| **MT-7** | Animations | 2-3h | Smooth transitions | Polish + UX |
| **MT-8** | Tests | 3-4h | 80%+ coverage | Unit + e2e tests |

---

## Start Here: MT-1 (Filter Store)

```typescript
// frontend/src/lib/stores/filterStore.ts

import { writable, derived } from 'svelte/store';

export interface FilterState {
  selectedCategories: string[];
  isOpen: boolean;
  appliedFilters: Record<string, boolean>;
}

// Create writable stores
export const filterState = writable<FilterState>({
  selectedCategories: [],
  isOpen: false,
  appliedFilters: {}
});

// Helper functions
export function toggleCategory(category: string) {
  filterState.update(state => ({
    ...state,
    selectedCategories: state.selectedCategories.includes(category)
      ? state.selectedCategories.filter(c => c !== category)
      : [...state.selectedCategories, category]
  }));
}

export function clearFilters() {
  filterState.set({
    selectedCategories: [],
    isOpen: false,
    appliedFilters: {}
  });
}

// Derived: count of active filters
export const activeFilterCount = derived(
  filterState,
  $state => $state.selectedCategories.length
);
```

---

## Then: MT-2 (Filter Component)

```svelte
<!-- frontend/src/lib/components/filters/CategoryFilter.svelte -->
<script lang="ts">
  import { filterState, toggleCategory, clearFilters } from '$lib/stores/filterStore';
  import { ChevronDown, X } from 'lucide-svelte';

  export let categories: string[] = [];
  export let disabled = false;

  let isOpen = false;
  let searchTerm = '';

  $: filtered = categories.filter(c =>
    c.toLowerCase().includes(searchTerm.toLowerCase())
  );

  function toggle(cat: string) {
    toggleCategory(cat);
  }

  function clear() {
    clearFilters();
    isOpen = false;
  }
</script>

<div class="relative">
  <!-- Dropdown button -->
  <button
    class="flex items-center gap-2 px-4 py-2 border rounded-lg hover:bg-gray-50"
    on:click={() => (isOpen = !isOpen)}
    {disabled}
  >
    <span>Filter</span>
    <ChevronDown class="w-4 h-4" />
  </button>

  <!-- Dropdown panel -->
  {#if isOpen}
    <div class="absolute top-full left-0 mt-2 w-64 bg-white border rounded-lg shadow-lg">
      <!-- Search box -->
      <input
        type="text"
        placeholder="Search categories..."
        bind:value={searchTerm}
        class="w-full px-3 py-2 border-b"
      />

      <!-- Category list -->
      <div class="max-h-48 overflow-y-auto">
        {#each filtered as category (category)}
          <label class="flex items-center gap-2 px-3 py-2 hover:bg-gray-50">
            <input
              type="checkbox"
              checked={$filterState.selectedCategories.includes(category)}
              on:change={() => toggle(category)}
            />
            <span>{category}</span>
          </label>
        {/each}
      </div>

      <!-- Action buttons -->
      <div class="flex gap-2 p-3 border-t">
        <button
          class="flex-1 px-3 py-1 text-sm bg-gray-100 rounded hover:bg-gray-200"
          on:click={clear}
        >
          Clear All
        </button>
      </div>
    </div>
  {/if}
</div>
```

---

## Then: MT-3 (API Integration)

```typescript
// frontend/src/lib/api/filterUtils.ts

export function buildFilterQuery(filters: Record<string, boolean>): string {
  const selected = Object.entries(filters)
    .filter(([, value]) => value)
    .map(([key]) => key);

  return selected.length > 0 ? selected.join(',') : '';
}

export function parseFilterFromQuery(query: string): Record<string, boolean> {
  const filters: Record<string, boolean> = {};
  if (query) {
    query.split(',').forEach(cat => {
      filters[cat] = true;
    });
  }
  return filters;
}

// Debounced API call
let timeoutId: ReturnType<typeof setTimeout>;

export async function applyFilters(
  categories: string[],
  searchTerm: string,
  page: number = 1
) {
  clearTimeout(timeoutId);

  timeoutId = setTimeout(async () => {
    const query = new URLSearchParams();
    if (categories.length > 0) {
      query.append('categories', categories.join(','));
    }
    if (searchTerm) {
      query.append('search', searchTerm);
    }
    query.append('page', page.toString());

    const response = await fetch(`/api/artifacts?${query}`);
    return response.json();
  }, 300); // 300ms debounce
}
```

---

## Then: MT-4 (URL Persistence)

```typescript
// In your page component or +page.ts
import { page } from '$app/stores';
import { parseFilterFromQuery } from '$lib/api/filterUtils';
import { filterState } from '$lib/stores/filterStore';

// On page load, restore filters from URL
export async function load({ url }) {
  const categoriesParam = url.searchParams.get('categories');
  if (categoriesParam) {
    filterState.update(state => ({
      ...state,
      selectedCategories: categoriesParam.split(',')
    }));
  }
  return {};
}

// When filters change, update URL
import { goto } from '$app/navigation';

filterState.subscribe(state => {
  const params = new URLSearchParams();
  if (state.selectedCategories.length > 0) {
    params.append('categories', state.selectedCategories.join(','));
  }
  // Update URL without page reload
  goto(`?${params.toString()}`, { replaceState: true });
});
```

---

## Then: MT-5 (Badge Components)

```svelte
<!-- frontend/src/lib/components/filters/FilterBadge.svelte -->
<script lang="ts">
  import { X } from 'lucide-svelte';

  export let category: string;
  export let onRemove: () => void;
</script>

<div class="inline-flex items-center gap-1 px-3 py-1 bg-blue-100 text-blue-700 rounded-full text-sm">
  <span>{category}</span>
  <button on:click={onRemove} class="hover:text-blue-900">
    <X class="w-4 h-4" />
  </button>
</div>
```

```svelte
<!-- frontend/src/lib/components/filters/FilterBadgeGroup.svelte -->
<script lang="ts">
  import FilterBadge from './FilterBadge.svelte';
  import { filterState, toggleCategory } from '$lib/stores/filterStore';
</script>

<div class="flex flex-wrap gap-2">
  {#each $filterState.selectedCategories as category (category)}
    <FilterBadge
      {category}
      onRemove={() => toggleCategory(category)}
    />
  {/each}
</div>
```

---

## Then: MT-6 (Clear Filters)

```svelte
<!-- Add to CategoryFilter.svelte or standalone -->
<script lang="ts">
  import { clearFilters } from '$lib/stores/filterStore';
  import { XCircle } from 'lucide-svelte';

  let showConfirm = false;

  function handleClear() {
    showConfirm = true;
  }

  function confirm() {
    clearFilters();
    showConfirm = false;
  }
</script>

<button
  on:click={handleClear}
  class="flex items-center gap-2 px-3 py-1 text-sm bg-red-100 text-red-700 rounded hover:bg-red-200"
>
  <XCircle class="w-4 h-4" />
  Clear All
</button>

{#if showConfirm}
  <div class="fixed inset-0 flex items-center justify-center bg-black/50">
    <div class="bg-white rounded-lg p-6 max-w-sm">
      <p class="mb-4">Clear all filters?</p>
      <div class="flex gap-2">
        <button on:click={() => (showConfirm = false)} class="px-4 py-2 bg-gray-200 rounded">
          Cancel
        </button>
        <button on:click={confirm} class="px-4 py-2 bg-red-500 text-white rounded">
          Clear
        </button>
      </div>
    </div>
  </div>
{/if}
```

---

## Then: MT-7 (Animations)

```svelte
<!-- Add transitions to FilterBadge -->
<script lang="ts">
  import { scale, fade } from 'svelte/transition';
</script>

<div
  class="inline-flex items-center gap-1 px-3 py-1 bg-blue-100 text-blue-700 rounded-full text-sm"
  transition:scale={{ duration: 200 }}
  transition:fade={{ duration: 200 }}
>
  <!-- ... -->
</div>
```

---

## Finally: MT-8 (Tests)

```typescript
// frontend/src/lib/__tests__/stores/filterStore.test.ts
import { describe, it, expect } from 'vitest';
import { filterState, toggleCategory, clearFilters } from '$lib/stores/filterStore';

describe('filterStore', () => {
  it('should toggle category selection', () => {
    let state;
    filterState.subscribe(s => (state = s));

    toggleCategory('Research');
    expect(state.selectedCategories).toContain('Research');

    toggleCategory('Research');
    expect(state.selectedCategories).not.toContain('Research');
  });

  it('should clear all filters', () => {
    let state;
    filterState.subscribe(s => (state = s));

    toggleCategory('Research');
    toggleCategory('AI');
    clearFilters();

    expect(state.selectedCategories).toHaveLength(0);
  });
});
```

---

## 🎯 Key Points

- **MT-1:** Everything starts here (2-3h)
- **MT-2, MT-3, MT-5:** Can run in parallel after MT-1
- **MT-4:** Start when MT-3 is done
- **MT-6:** Start when MT-2 + MT-5 are done
- **MT-7:** Final polish, start when others are done
- **MT-8:** Tests last, depends on all others

---

## 📋 Checklist Before Declaring Done

- [ ] All TypeScript compiles
- [ ] All tests pass
- [ ] Filters work in real-time
- [ ] URL shows selected filters
- [ ] Back/forward buttons work
- [ ] Mobile responsive
- [ ] Keyboard navigable
- [ ] Animations smooth
- [ ] No console errors
- [ ] Code reviewed

---

**Total Time:** 24-32 hours (sequential) or 16-20 hours (parallel)
**Status:** Ready to start
**Next Step:** Begin MT-1
