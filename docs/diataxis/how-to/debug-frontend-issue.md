# How To: Debug a Frontend Issue

> **Find and fix a component that's not rendering or causing errors.**
>
> Problem: A SvelteKit component isn't showing up, or you see blank screen with no error message visible in the UI.

---

## Quick Start

Debug frontend issues in 4 steps:

```bash
# Step 1: Open DevTools (F12)
# Step 2: Check Console tab for errors
# Step 3: Check Application tab for state/storage
# Step 4: Use Svelte DevTools to inspect component props
```

---

## Step 1: Open Browser Developer Tools

Press **F12** or **Cmd+Option+I** (Mac) to open Chrome DevTools.

You'll see five tabs at the top:
- **Elements**: DOM tree (what HTML is rendered)
- **Console**: JavaScript errors and logs
- **Sources**: Debugger (set breakpoints)
- **Network**: HTTP requests
- **Application**: Storage, cookies, service workers

---

## Step 2: Check Console Tab for Errors

Click the **Console** tab. This is where all JavaScript errors appear.

**Red X = Error:**

```
Uncaught TypeError: Cannot read property 'name' of undefined
  at UserProfile.svelte:15
```

This means something is `undefined` when you tried to access `.name` on it.

**Yellow Warning = Non-fatal issue:**

```
[vite] hmr update failed: Missing chunk 1234.js
```

This usually goes away on refresh.

**Common Errors:**

| Error Message | Likely Cause | Fix |
|---------------|-------------|-----|
| `Cannot read property 'X' of undefined` | Component received no props | Check parent passes props |
| `fetch failed: 404 Not Found` | API endpoint doesn't exist | Check API route in backend |
| `"X is not a function"` | Tried to call something that isn't a function | Check variable type |
| `SvelteKit: Page not found` | Route doesn't exist in `src/routes/` | Add route file |

---

## Step 3: Check if Component is in DOM

In DevTools **Elements** tab, search for your component:

**Cmd+F** (or **Ctrl+F**) to open Find box, type the component name or text you expect to see:

```
UserProfile
```

If the element **is not in the DOM**, the component didn't render (check console for errors).

If the element **is in the DOM but hidden**, check CSS:
- Is it hidden by `display: none`?
- Is it off-screen (`position: absolute; top: -9999px`)?
- Is it white text on white background?

---

## Step 4: Use Svelte DevTools Extension

Install **Svelte DevTools** Chrome extension:
1. Go to [Chrome Web Store](https://chrome.google.com/webstore)
2. Search "Svelte DevTools"
3. Click "Add to Chrome"

**Using Svelte DevTools:**

1. Open DevTools (F12)
2. Find the **Svelte** tab
3. Hover over components in the tree
4. See their props, state, and bindings in real time

**Example:** If `<UserProfile>` component shows:
```
name: undefined
email: undefined
```

Then the parent component isn't passing props. Check parent:

```svelte
<!-- WRONG: No props passed -->
<UserProfile />

<!-- RIGHT: Pass required props -->
<UserProfile name="John" email="john@example.com" />
```

---

## Step 5: Check Network Requests

If a component is rendering but the data is wrong or missing:

1. Open DevTools **Network** tab
2. Reload page
3. Look for failed (red) or slow requests

**Example:** If you expect `/api/users/123` to load but it's not there:
- Maybe the URL is wrong (check in Chrome DevTools Network)
- Maybe the endpoint doesn't exist on the backend (check backend logs)
- Maybe auth failed (401 Unauthorized response)

Click on a request to see details:
- **Status**: 200 = OK, 404 = not found, 500 = server error, 401 = unauthorized
- **Response**: What the server returned
- **Headers**: What was sent (including auth token)

---

## Common Frontend Debug Scenarios

### Scenario 1: Component Shows, But No Data

**Symptom:** Component renders but all text is blank.

**Debug Steps:**

1. Open Console (F12 → Console)
2. Check for errors about undefined variables
3. Open Svelte DevTools → find component → check props
4. Check Network tab → did API call return data?

**Example Fix:**

```svelte
<!-- WRONG: Will show blank if user is undefined -->
<h1>{user.name}</h1>

<!-- RIGHT: Show loading state while data fetches -->
{#if loading}
  <p>Loading...</p>
{:else if error}
  <p class="error">{error}</p>
{:else}
  <h1>{user.name}</h1>
{/if}
```

### Scenario 2: "Page Not Found" Error

**Symptom:** Get white screen with "Not Found" message.

**Debug Steps:**

1. Check URL in address bar
2. Check if route file exists in `src/routes/`
3. Check route naming convention (SvelteKit uses dynamic `[param]` syntax)

**Example:**

URL is `/users/123`, so create file:
```
src/routes/users/[id]/+page.svelte
```

Not:
```
src/routes/users.svelte  # Won't work for /users/123
```

### Scenario 3: CSS Not Applying

**Symptom:** Component renders but styles are missing.

**Debug Steps:**

1. Open DevTools **Elements** tab
2. Right-click the element, select "Inspect Element"
3. Look at the **Styles** panel (right side)
4. Check if your CSS rule is there
5. Check if it's crossed out (meaning another rule overrides it)

**Common CSS Issues:**

```svelte
<!-- WRONG: TailwindCSS class not recognized -->
<button class="bg-{color}">Click me</button>

<!-- RIGHT: Tailwind class names must be static -->
<button class={color === 'blue' ? 'bg-blue-500' : 'bg-gray-500'}>Click me</button>

<!-- Or use CSS variables -->
<style>
  button {
    background-color: var(--button-color);
  }
</style>
```

### Scenario 4: State Not Updating

**Symptom:** Click button, nothing happens. Variable doesn't change.

**Debug Steps:**

1. Open Console (F12 → Console)
2. Type the variable name, hit Enter
3. See current value
4. Click button
5. Type variable name again
6. Did it change?

**Example Fix (Using Svelte Runes):**

```svelte
<script>
  // WRONG: Won't update when you change it
  let count = 0;

  // RIGHT: Use $state rune to make reactive
  let count = $state(0);

  function increment() {
    count++;  // Now UI updates automatically
  }
</script>

<button on:click={increment}>Count: {count}</button>
```

---

## Using VS Code Debugger

For deeper debugging, use VS Code's built-in debugger:

1. Create `.vscode/launch.json`:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "type": "chrome",
      "request": "launch",
      "name": "Launch Chrome",
      "url": "http://localhost:5173",
      "webRoot": "${workspaceFolder}/frontend",
      "sourceMaps": true,
      "preLaunchTask": "npm: dev"
    }
  ]
}
```

2. Press **Ctrl+Shift+D** (or **Cmd+Shift+D** on Mac)
3. Click "Run" (green play button)
4. Chrome opens with DevTools attached
5. Set breakpoints in VS Code by clicking line numbers
6. Code will pause when it hits your breakpoint

---

## Checking Component Props

Use Svelte DevTools to verify props are being passed correctly:

```svelte
<!-- Parent component -->
<UserProfile user={currentUser} />

<!-- In Svelte DevTools, you'll see:
     UserProfile
       ├─ user: {id: 1, name: "John", ...}
     If user is undefined, check that currentUser exists
-->
```

---

## Environment Variables Not Loaded

If you get "Cannot read property 'X' of undefined" but the code looks right:

Check that your `.env` file is in the root:
```bash
.env
PUBLIC_API_URL=http://localhost:8001
```

SvelteKit only loads vars that start with `PUBLIC_`:

```svelte
<!-- RIGHT: PUBLIC_ variables are available -->
<script>
  const API_URL = import.meta.env.PUBLIC_API_URL;
</script>

<!-- WRONG: Non-PUBLIC vars are undefined in browser -->
<script>
  const API_URL = import.meta.env.API_URL;  // undefined!
</script>
```

---

## Server-Side Code Issues

If error is in `+page.server.ts` (backend), console won't show it. Check server logs instead:

1. Terminal where you ran `npm run dev`
2. Look for error stack trace
3. Restart dev server after fixing

---

## Full Debugging Checklist

- [ ] Open Console (F12), check for red errors
- [ ] Search DOM for your component (Cmd+F)
- [ ] Install Svelte DevTools, inspect props
- [ ] Check Network tab for failed API calls (404, 500, 401)
- [ ] Verify route file exists in `src/routes/`
- [ ] Check CSS by inspecting element (Elements tab)
- [ ] Verify Svelte Runes (`$state`, `$derived`) for reactivity
- [ ] Confirm `.env` variables start with `PUBLIC_`
- [ ] Restart dev server after code changes
- [ ] Check server logs for `+page.server.ts` errors

---

## When It's a Backend Issue

If you see 500 errors or 404s in Network tab, the problem is in BusinessOS backend, not frontend.

Check backend logs:

```bash
# If running docker-compose
docker-compose logs businessos-backend

# Look for error stack trace with slog format
```

See [API Endpoint Guide](./add-api-endpoint.md) to fix backend issues.

---

*See also: [Code Standards](../../CLAUDE.md#code-standards-typescript--svelte), [Svelte DevTools](https://github.com/sveltejs/extensions), [SvelteKit Docs](https://kit.svelte.dev/)*
