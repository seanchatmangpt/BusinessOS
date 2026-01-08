# Agent Detail Page - Visual & Technical Guide

---

## 1. Page Layout Structure

### Overall Layout (Desktop)
```
┌──────────────────────────────────────────────────────────────────┐
│ [Navbar with workspace switcher]                                 │
├──────────────────────────────────────────────────────────────────┤
│ Agents / [Agent Name] [← Back]                                   │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  HEADER SECTION                                                 │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │ [Avatar] Agent Name                [Edit] [Share] [More ⋮] │ │
│  │ Active Badge • Category • 2 hours ago                       │ │
│  └────────────────────────────────────────────────────────────┘ │
│                                                                  │
│  TAB NAVIGATION                                                 │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │ [Overview] [Usage Stats] [Settings] [Test]                │ │
│  └────────────────────────────────────────────────────────────┘ │
│                                                                  │
│  CONTENT AREA (Dynamic based on active tab)                    │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │                                                            │ │
│  │  [Tab content renders here]                              │ │
│  │                                                            │ │
│  │                                                            │ │
│  └────────────────────────────────────────────────────────────┘ │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

### Responsive Breakpoints
```
Mobile (< 640px):
  - Tab buttons stack or scroll horizontally
  - Single column layout
  - Reduced padding and margins

Tablet (640px - 1024px):
  - Two column layout where applicable
  - Tab buttons in horizontal scroll

Desktop (> 1024px):
  - Full three-column layout possible
  - All elements visible at once
```

---

## 2. Component Hierarchy

```
AgentDetailPage (+page.svelte)
│
├── Header
│   ├── Breadcrumb
│   ├── AgentCard
│   │   ├── Avatar
│   │   ├── AgentInfo (Name, Status, Category)
│   │   └── ActionButtons
│   │       ├── EditButton
│   │       ├── ShareButton
│   │       └── MoreMenu
│   │           ├── CloneOption
│   │           ├── CopyPromptOption
│   │           └── DeleteOption
│   │
│   └── TabBar
│       ├── TabButton (Overview)
│       ├── TabButton (Usage Stats)
│       ├── TabButton (Settings)
│       └── TabButton (Test)
│
├── ContentArea
│   ├── OverviewTab (conditional render)
│   │   ├── DescriptionSection
│   │   ├── ConfigurationSection
│   │   │   ├── ConfigItem
│   │   │   ├── ConfigItem
│   │   │   └── ConfigItem
│   │   ├── CapabilitiesSection
│   │   │   ├── TagBadge
│   │   │   └── TagBadge
│   │   ├── ToolsSection
│   │   │   └── ToolBadge
│   │   ├── ContextSourcesSection
│   │   │   └── SourceBadge
│   │   └── MetricsPanel (Sidebar)
│   │       ├── StatCard
│   │       └── StatCard
│   │
│   ├── UsageStatsTab (conditional render)
│   │   ├── PeriodSelector
│   │   ├── StatsCards (2x2 grid)
│   │   │   ├── StatCard
│   │   │   ├── StatCard
│   │   │   ├── StatCard
│   │   │   └── StatCard
│   │   ├── Charts (optional)
│   │   │   ├── LineChart
│   │   │   ├── BarChart
│   │   │   └── HistogramChart
│   │   ├── TestHistoryTable
│   │   │   ├── TableHeader
│   │   │   ├── TableRow
│   │   │   └── Pagination
│   │   └── TestDetailModal (on row click)
│   │
│   ├── SettingsTab (conditional render)
│   │   ├── FormSection (Basic Info)
│   │   │   ├── FormInput (DisplayName)
│   │   │   ├── FormTextarea (Description)
│   │   │   ├── FormSelect (Category)
│   │   │   └── FormInput (Avatar URL)
│   │   ├── FormSection (Model Config)
│   │   │   ├── FormTextarea (SystemPrompt)
│   │   │   ├── FormSelect (Model)
│   │   │   ├── FormSlider (Temperature)
│   │   │   ├── FormInput (MaxTokens)
│   │   │   ├── FormToggle (ThinkingEnabled)
│   │   │   └── FormToggle (StreamingEnabled)
│   │   ├── FormSection (Capabilities)
│   │   │   ├── CheckboxList
│   │   │   └── SearchableCheckbox
│   │   ├── FormSection (Status)
│   │   │   ├── RadioButton (Active)
│   │   │   ├── RadioButton (Inactive)
│   │   │   └── ActionButtons
│   │   │       ├── SaveButton
│   │   │       ├── RevertButton
│   │   │       └── DeleteButton
│   │   └── UnsavedChangesWarning (on exit)
│   │
│   └── TestTab (conditional render)
│       ├── TestInputForm
│       │   ├── FormTextarea (Message)
│       │   ├── FormToggle (Override Model)
│       │   ├── FormSlider (Override Temperature)
│       │   ├── SendButton
│       │   └── ClearButton
│       ├── StreamingResponseArea
│       │   ├── TypingIndicator (while streaming)
│       │   ├── ResponseText (displayed as received)
│       │   └── MetricsPanel
│       │       ├── TokenCount
│       │       ├── Duration
│       │       ├── Model
│       │       └── EstimatedCost
│       └── TestHistory
│           ├── HistoryItem
│           ├── RerunButton
│           └── CopyButton
│
└── LoadingState / ErrorState / NotFoundState

```

---

## 3. Detailed Component Designs

### 3.1 Header Section (Full Width)

```
┌─────────────────────────────────────────────────────────────────┐
│ 16px padding                                                     │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐  │
│  │ [Avatar]                                                │  │
│  │ 56x56px                                                 │  │
│  │ Initials if no image                                    │  │
│  │ Background: Linear gradient based on category           │  │
│  └─────────────────────────────────────────────────────────┘  │
│
│  Agent Display Name
│  [Status Badge] • Category Badge • Updated 2 hours ago
│
│                                    [Edit] [Share] [More ⋮]    │
│                                                                 │
│ 16px padding                                                    │
└─────────────────────────────────────────────────────────────────┘

Color Scheme:
  Background: #ffffff
  Text Primary: #1f2937 (dark gray)
  Text Secondary: #6b7280 (medium gray)
  Border: #e5e7eb (light gray)
  Action buttons:
    - Default: text-gray-600 hover:text-gray-900
    - Active: text-blue-600 hover:text-blue-700
```

### 3.2 Overview Tab Layout

**Column Layout:**
```
Desktop (1440px):
┌─────────────────────────────────────────────────────────────────┐
│                                                                 │
│  ┌──────────────────────────────┐  ┌──────────────────────┐  │
│  │ Description (full width)     │  │ Quick Metrics        │  │
│  │                              │  │ (300px sidebar)      │  │
│  ├──────────────────────────────┤  │                      │  │
│  │ Configuration (full width)   │  │ 4 stat cards         │  │
│  │ 2 columns of config items    │  │ 2x2 grid             │  │
│  │                              │  │                      │  │
│  ├──────────────────────────────┤  │                      │  │
│  │ Capabilities & Tools         │  │                      │  │
│  │ (3 columns of tags)          │  │                      │  │
│  │                              │  │                      │  │
│  ├──────────────────────────────┤  │                      │  │
│  │ Context Sources              │  │                      │  │
│  │ (3 columns of tags)          │  │                      │  │
│  └──────────────────────────────┘  └──────────────────────┘  │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘

Tablet (1024px):
┌─────────────────────────────────────────────────────────────────┐
│                                                                 │
│  ┌──────────────────────────────┐  ┌──────────────────┐       │
│  │ Description & Config         │  │ Quick Metrics    │       │
│  │                              │  │ (collapsed view) │       │
│  │ 2x2 config grid              │  │ 4 row layout     │       │
│  │                              │  │                  │       │
│  ├──────────────────────────────┤  └──────────────────┘       │
│  │ Capabilities & Tools (wrap)  │                             │
│  │                              │                             │
│  └──────────────────────────────┘                             │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘

Mobile (640px):
┌─────────────────────────────────────────────────────────────────┐
│ Description (full width)                                        │
│                                                                 │
│ Configuration                                                   │
│ • Model: gpt-4o                                                │
│ • Temperature: 0.7                                             │
│ • Max Tokens: 4096                                             │
│                                                                 │
│ Quick Metrics (stacked)                                         │
│ ┌─────────────────────────────────────────────────────────┐   │
│ │ Tests: 24      Success: 22 (92%)                       │   │
│ │ Avg Resp: 1,240ms    Avg Used: 450 tokens             │   │
│ └─────────────────────────────────────────────────────────┘   │
│                                                                 │
│ Capabilities (wrap)                                             │
│ [Code Gen] [Analysis] [Planning] [Documentation]              │
│                                                                 │
│ Tools (wrap)                                                    │
│ [Web Search] [Code Exec] [API Call]                           │
└─────────────────────────────────────────────────────────────────┘
```

### 3.3 Configuration Items

```
┌─────────────────────────────┐
│ Label (small caps)          │
│ Value or description        │
│ (secondary color)           │
└─────────────────────────────┘

Example:
┌─────────────────────────────┐
│ MODEL                       │
│ gpt-4o                      │
└─────────────────────────────┘

┌─────────────────────────────┐
│ TEMPERATURE                 │
│ 0.7 (creative)             │
└─────────────────────────────┘

Grid Layout (CSS):
grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
gap: 1.5rem;
```

### 3.4 Tag/Badge Styling

```
Tag Badge:
┌────────────────────┐
│ [Icon] Tag Text   [✕] │
└────────────────────┘

Styling:
  Background: #f3f4f6 (light gray)
  Text: #374151 (dark gray)
  Border: none
  Padding: 0.375rem 0.75rem
  Border-radius: 0.375rem
  Font-size: 0.875rem
  Display: inline-block
  Margin: 0.25rem

Color Variants:
  - Primary (capabilities): blue-100 / blue-700
  - Success (enabled tools): green-100 / green-700
  - Info (context sources): purple-100 / purple-700
```

### 3.5 Settings Tab Form

```
┌─────────────────────────────────────────────────────────────┐
│ FORM SECTION                                                │
│                                                             │
│ Field Label*                                                │
│ [Input / Textarea / Select]                                │
│ Helper text (optional)                                      │
│                                                             │
│ Field Label                                                 │
│ [Input / Textarea / Select]                                │
│                                                             │
│ Field Label                                                 │
│ ┌─────────────────────────────────────────────────────┐    │
│ │ [Checkbox] Enabled                                  │    │
│ │ [Checkbox] Disabled                                 │    │
│ │ [Checkbox] Advanced                                 │    │
│ └─────────────────────────────────────────────────────┘    │
│                                                             │
│ [Save Changes] [Revert] [Delete...]                       │
└─────────────────────────────────────────────────────────────┘

Form Validation:
  - Red border on error
  - Red text message below field
  - Disabled submit if invalid
  - Green border on success (optional)
```

### 3.6 Test Tab Layout

```
┌───────────────────────────────────────────────────────────┐
│ TEST INPUT                                                │
│ ┌─────────────────────────────────────────────────────┐  │
│ │ Your test message...                                │  │
│ │                                                     │  │
│ │                                                     │  │
│ └─────────────────────────────────────────────────────┘  │
│                                                           │
│ ☑ Override Model: [gpt-4o]                              │
│ ☐ Override Temp: [Slider 0.7]                          │
│                                                           │
│ [Send Test] [Clear]                                     │
├───────────────────────────────────────────────────────────┤
│ RESPONSE                                                  │
│ ┌─────────────────────────────────────────────────────┐  │
│ │ [Typing indicator ●●●] or [Response text]          │  │
│ │                                                     │  │
│ │ Agent response appears here as it streams in...    │  │
│ │ Each chunk updates in real-time.                   │  │
│ │                                                     │  │
│ └─────────────────────────────────────────────────────┘  │
│                                                           │
│ Status: ✓ Complete                                       │
│ Tokens: 342 | Time: 1,245ms | Model: gpt-4o | Cost: $0.01 │
│                                                           │
│ [Copy Response]  [Run Again]                            │
├───────────────────────────────────────────────────────────┤
│ TEST HISTORY (Last 5)                                     │
│ • "Tell me about..." → 432 tokens (0.9s)                │
│ • "Analyze this..." → 856 tokens (2.1s)                 │
│ • "Generate code" → 1,234 tokens (3.4s)                 │
│ • "Translate..." → 234 tokens (0.6s)                    │
│ • "Summarize..." → 567 tokens (1.2s)                    │
└───────────────────────────────────────────────────────────┘
```

---

## 4. Color & Styling Guide

### Color Palette
```
Primary Colors:
  Dark Text: #1f2937 (gray-900)
  Medium Text: #374151 (gray-700)
  Light Text: #6b7280 (gray-600)

Backgrounds:
  Primary BG: #ffffff
  Secondary BG: #f9fafb (gray-50)
  Hover BG: #f3f4f6 (gray-100)

Borders:
  Primary: #e5e7eb (gray-200)
  Secondary: #d1d5db (gray-300)

Semantic Colors:
  Success: #10b981 (green-500)
  Warning: #f59e0b (amber-500)
  Error: #ef4444 (red-500)
  Info: #3b82f6 (blue-500)

Category Badges:
  Analysis: #6366f1 (indigo)
  Code Generation: #8b5cf6 (violet)
  Planning: #ec4899 (pink)
  Documentation: #06b6d4 (cyan)
  Translation: #14b8a6 (teal)
  Debugging: #f97316 (orange)
```

### Typography
```
Font Family: System fonts (fallback to sans-serif)
  -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu...

Font Sizes:
  h1: 2rem (32px) - Page title
  h2: 1.5rem (24px) - Section title
  h3: 1.25rem (20px) - Subsection title
  h4: 1.125rem (18px) - Card title
  base: 1rem (16px) - Body text
  sm: 0.875rem (14px) - Secondary text, labels
  xs: 0.75rem (12px) - Helper text, badges

Font Weight:
  Normal: 400
  Medium: 500
  Semibold: 600
  Bold: 700

Line Height:
  Tight: 1.25
  Normal: 1.5
  Relaxed: 1.75
```

### Spacing Scale
```
xs: 0.25rem (4px)
sm: 0.5rem (8px)
md: 1rem (16px)
lg: 1.5rem (24px)
xl: 2rem (32px)
2xl: 3rem (48px)

Common Patterns:
  Section padding: lg (1.5rem)
  Card padding: md (1rem)
  Gap between elements: sm-md (8-16px)
  Gap between sections: lg-xl (24-32px)
```

### Shadow & Depth
```
None: 0 0 #0000
sm: 0 1px 2px 0 rgba(0, 0, 0, 0.05)
base: 0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px 0 rgba(0, 0, 0, 0.06)
md: 0 4px 6px -1px rgba(0, 0, 0, 0.1)
lg: 0 10px 15px -3px rgba(0, 0, 0, 0.1)

Cards: shadow-sm with border 1px gray-200
Hover States: shadow-md
Modals/Overlays: shadow-lg
```

### Border Radius
```
xs: 0.125rem (2px)
sm: 0.25rem (4px)
base: 0.375rem (6px)
md: 0.5rem (8px)
lg: 0.75rem (12px)
xl: 1rem (16px)

Usage:
  Buttons: md (0.5rem)
  Cards: lg (0.75rem)
  Inputs: md (0.5rem)
  Badges: sm (0.25rem)
  Avatars: full (circular)
```

---

## 5. State & Animation Guide

### Loading States
```
Global Loading:
  - Full page spinner centered
  - Blur background
  - Message: "Loading agent details..."

Partial Loading (Tab):
  - Skeleton loaders for cards
  - Pulse animation
  - Fade in when content loads

Streaming:
  - Typing indicator: animated dots ●●●
  - Fade in for each chunk
  - Smooth scrolling to bottom as content appears
```

### Error States
```
Global Error:
┌─────────────────────────────────────────┐
│ ⚠️ Error loading agent                  │
│ The agent you're looking for doesn't   │
│ exist or you don't have access.        │
│                                         │
│ [Back to Agents] [Retry]               │
└─────────────────────────────────────────┘

Field Error:
  Red border: 2px solid #ef4444
  Error message below: color: #dc2626
  Error icon: ●

Toast Notifications:
  Position: bottom-right
  Background:
    - Success: #10b981 bg-green-500
    - Error: #ef4444 bg-red-500
    - Info: #3b82f6 bg-blue-500
  Auto-dismiss: 3 seconds
```

### Transitions & Animations
```
Duration Presets:
  Fast: 150ms - hover states, quick feedback
  Normal: 250ms - tab switches, modal opens
  Slow: 350ms - page transitions, important content

Easing:
  Default: cubic-bezier(0.4, 0, 0.2, 1)
  Ease-in: cubic-bezier(0.4, 0, 1, 1)
  Ease-out: cubic-bezier(0, 0, 0.2, 1)

Common Animations:
  Tab switches: opacity 250ms, slide 250ms
  Button hover: background 150ms, shadow 150ms
  Modal open: opacity 250ms, scale 250ms
  Form input: border-color 150ms, shadow 150ms
```

### Hover & Focus States
```
Button Hover:
  - Slightly darker background
  - Increased shadow
  - Underline for text buttons

Input Focus:
  - Blue border (2px #3b82f6)
  - Blue shadow (inset)
  - Outline: none (browser default removed)

Card Hover:
  - Slight lift (translate-y -2px)
  - Shadow increase
  - Smooth transition

Interactive Elements:
  - Cursor: pointer
  - Transition: all 150ms
```

---

## 6. Accessibility Checklist

### WCAG 2.1 Level AA Compliance

**Color Contrast**
- [x] Text to background: 4.5:1 for normal text
- [x] Text to background: 3:1 for large text
- [x] UI components: 3:1 contrast
- [x] Avoid color-only info (always use text + icon)

**Keyboard Navigation**
- [x] Tab order logical and visible (focus indicator)
- [x] All buttons/links keyboard accessible
- [x] Tab key moves focus
- [x] Enter/Space activate buttons
- [x] Escape closes modals/dropdowns
- [x] Form fields labeled with `<label for="id">`

**Screen Readers**
- [x] Semantic HTML (`<button>`, `<nav>`, `<main>`)
- [x] ARIA labels where needed
- [x] ARIA describedby for error messages
- [x] ARIA live regions for dynamic content
- [x] Form error announcement: role="alert"

**Mobile Accessibility**
- [x] Touch targets: 44x44px minimum
- [x] No content lost in zoomed views
- [x] Readable font size (16px base minimum)
- [x] No horizontal scrolling required

**Testing Tools**
- Automated: axe DevTools, Lighthouse
- Manual: NVDA (Windows), JAWS (Windows), VoiceOver (Mac)
- Keyboard only navigation test
```

---

## 7. Example Code Snippets

### Header Component (Svelte)
```svelte
<script lang="ts">
    import type { AgentDetailResponse } from '$lib/api';

    export let agent: AgentDetailResponse;
    export let onEdit: () => void;

    let showMoreMenu = false;
    let copiedToast = false;

    function handleShare() {
        navigator.clipboard.writeText(window.location.href);
        copiedToast = true;
        setTimeout(() => copiedToast = false, 2000);
    }

    function handleClone() {
        // Implementation
    }
</script>

<header class="bg-white border-b border-gray-200 px-6 py-4">
    <div class="flex items-center justify-between">
        <div class="flex items-center gap-4">
            <div class="w-14 h-14 rounded-xl bg-gradient-to-br from-blue-100 to-blue-200
                        flex items-center justify-center text-lg font-semibold text-blue-600">
                {agent.display_name[0]}
            </div>
            <div>
                <h1 class="text-2xl font-semibold text-gray-900">{agent.display_name}</h1>
                <div class="flex items-center gap-3 mt-1">
                    <span class={`px-2 py-0.5 text-xs rounded-md font-medium
                                 ${agent.is_active ? 'bg-green-50 text-green-700' : 'bg-gray-50 text-gray-700'}`}>
                        {agent.is_active ? 'Active' : 'Inactive'}
                    </span>
                    {#if agent.category}
                        <span class="text-xs text-gray-500">{agent.category}</span>
                    {/if}
                    <span class="text-xs text-gray-400">
                        Updated {new Date(agent.updated_at).toLocaleDateString()}
                    </span>
                </div>
            </div>
        </div>

        <div class="flex items-center gap-2">
            <button onclick={onEdit}
                    class="px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-100 rounded-lg transition-colors">
                Edit
            </button>
            <button onclick={handleShare}
                    class="px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-100 rounded-lg transition-colors">
                Share
            </button>
            <div class="relative">
                <button onclick={() => showMoreMenu = !showMoreMenu}
                        class="p-2 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg">
                    ⋮
                </button>
                {#if showMoreMenu}
                    <div class="absolute right-0 mt-2 w-48 bg-white rounded-lg shadow-lg border border-gray-200 z-50">
                        <button onclick={handleClone}
                                class="w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-100">
                            Clone Agent
                        </button>
                        <button onclick={() => { /* Copy prompt */ }}
                                class="w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-100">
                            Copy Prompt
                        </button>
                        <button onclick={() => { /* Delete */ }}
                                class="w-full text-left px-4 py-2 text-sm text-red-600 hover:bg-red-50">
                            Delete
                        </button>
                    </div>
                {/if}
            </div>
        </div>
    </div>
</header>

{#if copiedToast}
    <div class="fixed bottom-4 right-4 bg-green-500 text-white px-4 py-2 rounded-lg shadow-lg">
        Link copied to clipboard!
    </div>
{/if}
```

### Tab Navigation (Svelte)
```svelte
<script lang="ts">
    import { page } from '$app/stores';
    import { goto } from '$app/navigation';

    type TabType = 'overview' | 'usage-stats' | 'settings' | 'test';

    let activeTab: TabType = ('overview' as TabType);

    $effect(() => {
        const tabParam = $page.url.searchParams.get('tab');
        if (tabParam && ['overview', 'usage-stats', 'settings', 'test'].includes(tabParam)) {
            activeTab = tabParam as TabType;
        }
    });

    function selectTab(tab: TabType) {
        activeTab = tab;
        const url = new URL($page.url);
        url.searchParams.set('tab', tab);
        goto(url.toString());
    }
</script>

<nav class="border-b border-gray-200 px-6">
    <div class="flex gap-6">
        {#each [
            { id: 'overview', label: 'Overview' },
            { id: 'usage-stats', label: 'Usage Stats' },
            { id: 'settings', label: 'Settings' },
            { id: 'test', label: 'Test' }
        ] as tab}
            <button
                onclick={() => selectTab(tab.id as TabType)}
                class="py-3 text-sm font-medium border-b-2 transition-colors
                       {activeTab === tab.id
                           ? 'border-gray-900 text-gray-900'
                           : 'border-transparent text-gray-500 hover:text-gray-700'}"
            >
                {tab.label}
            </button>
        {/each}
    </div>
</nav>
```

---

## 8. Performance Optimization Tips

### Rendering Optimization
```svelte
<!-- Lazy load heavy components -->
{#if activeTab === 'usage-stats'}
    <UsageStatsTab agent={agent} />
{/if}

<!-- Use key blocks for list rendering -->
{#each tests as test (test.id)}
    <TestRow {test} />
{/each}

<!-- Debounce form inputs -->
let formTimeout;
function handleFormChange(field, value) {
    clearTimeout(formTimeout);
    formTimeout = setTimeout(() => {
        // Save changes
    }, 500);
}
```

### Data Caching
```typescript
const CACHE_TTL = 5 * 60 * 1000; // 5 minutes

class AgentCache {
    private cache = new Map<string, { data: any, timestamp: number }>();

    get(id: string) {
        const cached = this.cache.get(id);
        if (cached && Date.now() - cached.timestamp < CACHE_TTL) {
            return cached.data;
        }
        return null;
    }

    set(id: string, data: any) {
        this.cache.set(id, { data, timestamp: Date.now() });
    }
}
```

---

**End of Visual & Technical Guide**
**Version:** 1.0
**Last Updated:** January 8, 2026
