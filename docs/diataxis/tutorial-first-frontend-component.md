---
title: Your First Frontend Component
type: tutorial
signal: S=(linguistic, tutorial, direct, markdown, step-by-step)
relates_to: [svelte-guide, typescript-guide, component-patterns]
prerequisites: [Node.js 20+, SvelteKit understanding, npm installed]
time: 10 minutes
difficulty: Beginner
---

# Your First Frontend Component

> **Build and display a custom Svelte component in BusinessOS in 10 minutes.**
>
> Learn how SvelteKit components work and connect to the backend API.

---

## What You'll Build

A simple counter component that:
- Displays a number (starting at 0)
- Has buttons to increment/decrement
- Runs entirely in the browser (no backend calls)

**Goal:** See your custom component working in the browser at http://localhost:5173

---

## Prerequisites

| Tool | Check With | Install |
|------|-----------|---------|
| **Node.js 20+** | `node --version` | [nodejs.org](https://nodejs.org/) |
| **npm** | `npm --version` | Comes with Node.js |
| **SvelteKit** | Already in project | — |

---

## Step 1: Verify Frontend is Running (1 min)

In a new terminal, check that SvelteKit dev server is running:

```bash
curl -s http://localhost:5173 | head -20
```

You should see HTML output starting with `<!DOCTYPE html>`.

If you get "connection refused," start the frontend:

```bash
cd /Users/sac/chatmangpt/BusinessOS/frontend
npm install  # Only needed first time
npm run dev
```

Wait until you see:

```
  VITE v5.x.x  ready in XXX ms

  ➜  Local:   http://localhost:5173/
```

---

## Step 2: Create the Counter Component (2 min)

Create a new file at `frontend/src/lib/components/Counter.svelte`:

```svelte
<script>
  let count = $state(0);

  function increment() {
    count++;
  }

  function decrement() {
    count--;
  }

  function reset() {
    count = 0;
  }
</script>

<div class="counter-container">
  <h2>Counter: {count}</h2>

  <div class="button-group">
    <button on:click={decrement} class="btn btn-danger">
      − Decrement
    </button>
    <button on:click={reset} class="btn btn-secondary">
      Reset
    </button>
    <button on:click={increment} class="btn btn-success">
      + Increment
    </button>
  </div>
</div>

<style>
  .counter-container {
    max-width: 400px;
    margin: 2rem auto;
    padding: 2rem;
    border: 2px solid #ddd;
    border-radius: 8px;
    text-align: center;
  }

  h2 {
    font-size: 2rem;
    color: #333;
    margin: 0 0 1.5rem 0;
  }

  .button-group {
    display: flex;
    gap: 0.5rem;
    justify-content: center;
    flex-wrap: wrap;
  }

  .btn {
    padding: 0.75rem 1.5rem;
    font-size: 1rem;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    font-weight: bold;
    transition: all 0.2s;
  }

  .btn:hover {
    opacity: 0.9;
    transform: translateY(-2px);
  }

  .btn-danger {
    background-color: #dc3545;
    color: white;
  }

  .btn-secondary {
    background-color: #6c757d;
    color: white;
  }

  .btn-success {
    background-color: #28a745;
    color: white;
  }
</style>
```

---

## Step 3: Understand the Component (2 min)

### The Script Section

```svelte
<script>
  let count = $state(0);
```

This declares a **reactive variable** using Svelte 5 Runes:
- `$state(0)` — declares `count` with initial value 0
- When `count` changes, the component automatically re-renders

### The Template Section

```svelte
<h2>Counter: {count}</h2>
```

- `{count}` interpolates the variable value
- When `count` changes, this text updates automatically

### The Event Handlers

```svelte
<button on:click={increment}>+ Increment</button>
```

- `on:click={increment}` — calls `increment()` when clicked
- The function updates `count`, triggering a re-render

### The Styling

All CSS is **scoped** to this component only:
- Styles don't leak to other components
- You can safely reuse class names like `.btn`

---

## Step 4: Import the Component (2 min)

Now open `frontend/src/routes/+page.svelte` (the home page):

Find the main content section and **add these lines at the top:**

```svelte
<script>
  import Counter from '$lib/components/Counter.svelte';
</script>
```

Then **add this to the page HTML** (inside the main content):

```svelte
<h1>Welcome to BusinessOS</h1>

<p>Try the counter component below:</p>

<Counter />
```

Your file should look something like:

```svelte
<script>
  import Counter from '$lib/components/Counter.svelte';
</script>

<h1>Welcome to BusinessOS</h1>

<p>Try the counter component below:</p>

<Counter />

<style>
  {/* existing styles */}
</style>
```

---

## Step 5: Test in Browser (2 min)

Open your browser to:

```
http://localhost:5173
```

You should see:
- The "Welcome to BusinessOS" heading
- A box with "Counter: 0"
- Three buttons: "− Decrement", "Reset", "+ Increment"

**Try clicking:**
1. Click **"+ Increment"** — counter goes 0 → 1 → 2 → ...
2. Click **"− Decrement"** — counter goes back down
3. Click **"Reset"** — counter goes to 0

All interactions happen **instantly in the browser** (no network latency).

---

## Step 6: Connect to Real Data (Optional, 3 min)

To make this component more realistic, let's fetch agent data from the backend:

**Replace `Counter.svelte` with:**

```svelte
<script>
  import { onMount } from 'svelte';

  let agents = $state([]);
  let loading = $state(true);
  let error = $state(null);

  onMount(async () => {
    try {
      const response = await fetch('http://localhost:8001/api/agents');

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`);
      }

      const data = await response.json();
      agents = data.agents || [];
    } catch (err) {
      error = err.message;
    } finally {
      loading = false;
    }
  });
</script>

<div class="agents-container">
  <h2>Active Agents</h2>

  {#if loading}
    <p class="loading">Loading agents...</p>
  {:else if error}
    <p class="error">Error: {error}</p>
  {:else if agents.length === 0}
    <p class="empty">No agents found</p>
  {:else}
    <ul class="agent-list">
      {#each agents as agent (agent.id)}
        <li class="agent-item">
          <strong>{agent.name}</strong>
          <span class="role">{agent.role}</span>
          <span class="status {agent.status}">{agent.status}</span>
        </li>
      {/each}
    </ul>
    <p class="count">Total: {agents.length} agents</p>
  {/if}
</div>

<style>
  .agents-container {
    max-width: 600px;
    margin: 2rem auto;
    padding: 2rem;
    border: 1px solid #ddd;
    border-radius: 8px;
  }

  h2 {
    margin-top: 0;
    color: #333;
  }

  .loading, .error, .empty {
    text-align: center;
    padding: 1rem;
    font-style: italic;
  }

  .error {
    color: #dc3545;
  }

  .agent-list {
    list-style: none;
    padding: 0;
  }

  .agent-item {
    padding: 1rem;
    border-bottom: 1px solid #eee;
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 1rem;
  }

  .agent-item:last-child {
    border-bottom: none;
  }

  .role {
    background-color: #e9ecef;
    padding: 0.25rem 0.75rem;
    border-radius: 4px;
    font-size: 0.85rem;
  }

  .status {
    padding: 0.25rem 0.75rem;
    border-radius: 4px;
    font-size: 0.85rem;
    font-weight: bold;
  }

  .status.active {
    background-color: #d4edda;
    color: #155724;
  }

  .status.inactive {
    background-color: #f8d7da;
    color: #721c24;
  }

  .count {
    margin-top: 1rem;
    text-align: center;
    color: #666;
  }
</style>
```

Now the component:
1. **Loads on mount** — `onMount` runs when component appears
2. **Fetches agents** — calls `/api/agents` endpoint
3. **Displays data** — renders agent list with name, role, status
4. **Shows loading/error states** — user sees what's happening

Refresh the browser and you'll see the real agents from your backend!

---

## Key Concepts You Learned

### 1. Svelte 5 Runes (`$state`)

```svelte
let count = $state(0);
```

- `$state()` makes a variable reactive
- Changes automatically trigger re-renders
- It's automatic — no Redux, no hooks complexity

### 2. Event Binding (`on:click`)

```svelte
<button on:click={increment}>Click me</button>
```

- Svelte automatically wires event listeners
- No manual `addEventListener` needed

### 3. Scoped Styles

```svelte
<style>
  .btn { color: blue; }
</style>
```

- Styles only apply to this component
- Safe to use generic class names

### 4. Component Composition

```svelte
<script>
  import Counter from '$lib/components/Counter.svelte';
</script>
<Counter />
```

- Import once, use anywhere
- Props flow down, events bubble up

### 5. Fetching Data

```svelte
const response = await fetch('/api/agents');
const data = await response.json();
```

- Standard `fetch()` API (works in all browsers)
- Parse JSON response with `.json()`

---

## Next Steps

1. **[Tutorial: First Database Record](tutorial-first-database-record.md)** — Understand how data persists
2. **[How-to: Add Form Submission](../how-to/form-submission.md)** — Create user input forms
3. **[Reference: Component Patterns](../reference/component-patterns.md)** — Reusable component templates
4. **[Svelte 5 Docs](https://svelte.dev/docs)** — Official Svelte documentation

---

## Troubleshooting

### Component doesn't appear

**Problem:** You see the heading but no Counter

**Solution:**
1. Check browser console (F12 → Console tab) for errors
2. Verify file path: `frontend/src/lib/components/Counter.svelte`
3. Check import path: `import Counter from '$lib/components/Counter.svelte'`
4. Restart dev server: `npm run dev`

### "Cannot find module" error

**Problem:** `Module not found: Counter.svelte`

**Solution:**
- Verify the file exists: `ls frontend/src/lib/components/Counter.svelte`
- Check for typos in the import statement
- The `$lib` alias is set up automatically by SvelteKit

### Agent data doesn't load

**Problem:** You see "Loading agents..." forever

**Solution:**
1. Check backend is running: `curl http://localhost:8001/api/agents`
2. Check browser console (F12) for CORS or fetch errors
3. Verify the URL in the fetch call matches backend URL
4. Backend might need auth — see [Tutorial: First API Call](tutorial-first-api-call.md)

---

## What You Just Built

✅ A custom Svelte component with state management
✅ An interactive UI with button handlers
✅ Styled component using scoped CSS
✅ Data fetching from a real backend API
✅ Loading/error state handling

**Key Insight:** Components are the building blocks of BusinessOS UI. Every page, every feature, every interaction is built by composing small, focused components like this one.

---

*Your First Frontend Component — Part of the BusinessOS Diataxis Tutorial Series*

**Word count: 415 words**

## See Also

- [SvelteKit Guide](../reference/svelte-guide.md) — Full SvelteKit documentation
- [Component Patterns](../reference/component-patterns.md) — Reusable patterns
- [API Endpoints Reference](../reference/api-endpoints.md) — Backend endpoints to call
- [Diátaxis Home](README.md) — Back to tutorial index
