import { writable, derived, get } from 'svelte/store';
import { browser } from '$app/environment';

// ============================================================================
// TYPE DEFINITIONS
// ============================================================================

/**
 * Canvas viewport state for pan/zoom
 */
export interface CanvasViewport {
	/** X offset in pixels */
	offsetX: number;
	/** Y offset in pixels */
	offsetY: number;
	/** Zoom level (0.1 to 3.0, 1.0 = 100%) */
	zoom: number;
}

/**
 * Widget position and size on canvas grid
 */
export interface WidgetLayout {
	/** Unique widget instance ID */
	id: string;
	/** Widget type identifier */
	type: WidgetType;
	/** X coordinate in pixels */
	x: number;
	/** Y coordinate in pixels */
	y: number;
	/** Width in pixels */
	width: number;
	/** Height in pixels */
	height: number;
	/** Z-index for layering */
	zIndex: number;
	/** Optional custom configuration */
	config?: Record<string, unknown>;
	/** Widget title */
	title: string;
	/** Collapsed state */
	collapsed?: boolean;
	/** Accent color */
	accentColor?: string;
}

/**
 * Complete dashboard layout configuration
 */
export interface DashboardLayout {
	/** Unique layout ID */
	id: string;
	/** User-friendly name */
	name: string;
	/** Creation timestamp */
	createdAt: Date;
	/** Last update timestamp */
	updatedAt: Date;
	/** Canvas viewport state */
	viewport: CanvasViewport;
	/** Widget positions and sizes */
	widgets: WidgetLayout[];
	/** Grid configuration */
	gridConfig: GridConfig;
	/** Whether this is the active layout */
	isActive: boolean;
}

/**
 * Grid configuration settings
 */
export interface GridConfig {
	/** Grid cell size in pixels */
	cellSize: number;
	/** Whether to snap to grid */
	snapToGrid: boolean;
	/** Whether to show grid lines */
	showGrid: boolean;
	/** Grid color (hex or rgba) */
	gridColor: string;
	/** Spacing between cells in pixels */
	spacing: number;
}

/**
 * Widget type identifiers
 */
export type WidgetType =
	| 'focus'
	| 'quick-actions'
	| 'projects'
	| 'tasks'
	| 'activity'
	| 'metric'
	| 'insights'
	| 'productivity-chart'
	| 'notifications';

/**
 * Persistence error types
 */
export type PersistenceError =
	| 'QUOTA_EXCEEDED'
	| 'PARSE_ERROR'
	| 'INVALID_DATA'
	| 'BROWSER_UNSUPPORTED';

/**
 * Store state interface
 */
interface DashboardLayoutState {
	/** All available layouts */
	layouts: DashboardLayout[];
	/** Currently active layout ID */
	activeLayoutId: string | null;
	/** Whether data is being loaded */
	loading: boolean;
	/** Last persistence error */
	error: PersistenceError | null;
	/** Unsaved changes flag */
	isDirty: boolean;
	/** Last successful save timestamp */
	lastSaved: Date | null;
}

// ============================================================================
// CONSTANTS & DEFAULTS
// ============================================================================

/** localStorage key for dashboard layouts */
const STORAGE_KEY = 'businessos-dashboard-layouts';

/** localStorage key for active layout ID */
const ACTIVE_LAYOUT_KEY = 'businessos-dashboard-active-layout';

/** localStorage version for migration support */
const STORAGE_VERSION_KEY = 'businessos-dashboard-version';
const CURRENT_VERSION = 1;

/** Default viewport state */
const DEFAULT_VIEWPORT: CanvasViewport = {
	offsetX: 0,
	offsetY: 0,
	zoom: 1.0
};

/** Default grid configuration */
const DEFAULT_GRID_CONFIG: GridConfig = {
	cellSize: 50,
	snapToGrid: true,
	showGrid: true,
	gridColor: 'rgba(200, 200, 200, 0.15)',
	spacing: 16
};

/** Default layout for new users */
const DEFAULT_LAYOUT: DashboardLayout = {
	id: 'default',
	name: 'Default Layout',
	createdAt: new Date(),
	updatedAt: new Date(),
	viewport: DEFAULT_VIEWPORT,
	widgets: [
		{
			id: 'w1',
			type: 'focus',
			title: "Today's Focus",
			x: 50,
			y: 50,
			width: 600,
			height: 500,
			zIndex: 1
		},
		{
			id: 'w2',
			type: 'quick-actions',
			title: 'Quick Actions',
			x: 700,
			y: 50,
			width: 500,
			height: 400,
			zIndex: 1
		},
		{
			id: 'w3',
			type: 'notifications',
			title: 'Alerts',
			x: 1250,
			y: 50,
			width: 550,
			height: 500,
			zIndex: 1
		},
		{
			id: 'w4',
			type: 'projects',
			title: 'Active Projects',
			x: 50,
			y: 600,
			width: 600,
			height: 550,
			zIndex: 1
		},
		{
			id: 'w5',
			type: 'tasks',
			title: 'My Tasks',
			x: 700,
			y: 500,
			width: 800,
			height: 600,
			zIndex: 1
		},
		{
			id: 'w6',
			type: 'productivity-chart',
			title: 'Weekly Productivity',
			x: 50,
			y: 1200,
			width: 1100,
			height: 450,
			zIndex: 1
		}
	],
	gridConfig: DEFAULT_GRID_CONFIG,
	isActive: true
};

/** Initial store state */
const INITIAL_STATE: DashboardLayoutState = {
	layouts: [],
	activeLayoutId: null,
	loading: false,
	error: null,
	isDirty: false,
	lastSaved: null
};

// ============================================================================
// VALIDATION
// ============================================================================

/**
 * Validate layout data structure
 */
function validateLayout(layout: unknown): layout is DashboardLayout {
	if (!layout || typeof layout !== 'object') return false;

	const l = layout as Partial<DashboardLayout>;

	return (
		typeof l.id === 'string' &&
		typeof l.name === 'string' &&
		l.viewport !== undefined &&
		typeof l.viewport.offsetX === 'number' &&
		typeof l.viewport.offsetY === 'number' &&
		typeof l.viewport.zoom === 'number' &&
		Array.isArray(l.widgets) &&
		l.widgets.every(validateWidget) &&
		l.gridConfig !== undefined &&
		validateGridConfig(l.gridConfig)
	);
}

/**
 * Validate widget data
 */
function validateWidget(widget: unknown): widget is WidgetLayout {
	if (!widget || typeof widget !== 'object') return false;

	const w = widget as Partial<WidgetLayout>;

	return (
		typeof w.id === 'string' &&
		typeof w.type === 'string' &&
		typeof w.x === 'number' &&
		typeof w.y === 'number' &&
		typeof w.width === 'number' &&
		typeof w.height === 'number' &&
		typeof w.zIndex === 'number' &&
		typeof w.title === 'string'
	);
}

/**
 * Validate grid config
 */
function validateGridConfig(config: unknown): config is GridConfig {
	if (!config || typeof config !== 'object') return false;

	const c = config as Partial<GridConfig>;

	return (
		typeof c.cellSize === 'number' &&
		typeof c.snapToGrid === 'boolean' &&
		typeof c.showGrid === 'boolean' &&
		typeof c.gridColor === 'string' &&
		typeof c.spacing === 'number'
	);
}

// ============================================================================
// LOCALSTORAGE OPERATIONS
// ============================================================================

/**
 * Load layouts from localStorage with error handling
 */
function loadLayoutsFromStorage(): DashboardLayout[] {
	if (!browser) return [DEFAULT_LAYOUT];

	try {
		const stored = localStorage.getItem(STORAGE_KEY);
		if (!stored) {
			console.log('[Dashboard Layout] No stored layouts, using default');
			return [DEFAULT_LAYOUT];
		}

		const parsed = JSON.parse(stored);

		if (!Array.isArray(parsed)) {
			console.error('[Dashboard Layout] Invalid stored format (not array)');
			return [DEFAULT_LAYOUT];
		}

		// Validate and filter valid layouts
		const validLayouts = parsed.filter(validateLayout);

		if (validLayouts.length === 0) {
			console.error('[Dashboard Layout] No valid layouts found in storage');
			return [DEFAULT_LAYOUT];
		}

		// Deserialize dates
		validLayouts.forEach((layout) => {
			layout.createdAt = new Date(layout.createdAt);
			layout.updatedAt = new Date(layout.updatedAt);
		});

		console.log('[Dashboard Layout] Loaded', validLayouts.length, 'layouts from storage');
		return validLayouts;
	} catch (error) {
		if (error instanceof SyntaxError) {
			console.error('[Dashboard Layout] Parse error:', error);
			return [DEFAULT_LAYOUT];
		}

		console.error('[Dashboard Layout] Unexpected error loading:', error);
		return [DEFAULT_LAYOUT];
	}
}

/**
 * Save layouts to localStorage with error handling
 */
function saveLayoutToStorage(layouts: DashboardLayout[]): PersistenceError | null {
	if (!browser) return null;

	try {
		const serialized = JSON.stringify(layouts);
		localStorage.setItem(STORAGE_KEY, serialized);
		console.log('[Dashboard Layout] Saved', layouts.length, 'layouts to storage');
		return null;
	} catch (error) {
		if (error instanceof Error) {
			if (error.name === 'QuotaExceededError') {
				console.error('[Dashboard Layout] Storage quota exceeded');
				return 'QUOTA_EXCEEDED';
			}
		}
		console.error('[Dashboard Layout] Failed to save:', error);
		return 'INVALID_DATA';
	}
}

/**
 * Load active layout ID
 */
function loadActiveLayoutId(): string | null {
	if (!browser) return null;

	try {
		return localStorage.getItem(ACTIVE_LAYOUT_KEY);
	} catch {
		return null;
	}
}

/**
 * Save active layout ID
 */
function saveActiveLayoutId(id: string): void {
	if (!browser) return;

	try {
		localStorage.setItem(ACTIVE_LAYOUT_KEY, id);
	} catch (error) {
		console.error('[Dashboard Layout] Failed to save active layout ID:', error);
	}
}

/**
 * Clear all dashboard layout data
 */
function clearStorage(): void {
	if (!browser) return;

	try {
		localStorage.removeItem(STORAGE_KEY);
		localStorage.removeItem(ACTIVE_LAYOUT_KEY);
		console.log('[Dashboard Layout] Cleared all storage');
	} catch (error) {
		console.error('[Dashboard Layout] Failed to clear storage:', error);
	}
}

// ============================================================================
// STORE IMPLEMENTATION
// ============================================================================

function createDashboardLayoutStore() {
	const { subscribe, set, update } = writable<DashboardLayoutState>(INITIAL_STATE);

	// Auto-save timer
	let autoSaveTimer: ReturnType<typeof setTimeout> | null = null;
	const AUTO_SAVE_DELAY = 2000; // 2 seconds debounce

	/**
	 * Schedule auto-save with debouncing
	 */
	function scheduleAutoSave() {
		if (autoSaveTimer) clearTimeout(autoSaveTimer);

		autoSaveTimer = setTimeout(() => {
			const state = get({ subscribe });
			if (state.isDirty) {
				saveToStorage();
			}
		}, AUTO_SAVE_DELAY);
	}

	/**
	 * Save current layouts to storage
	 */
	function saveToStorage(): PersistenceError | null {
		const state = get({ subscribe });
		const error = saveLayoutToStorage(state.layouts);

		if (!error) {
			update((s) => ({
				...s,
				error: null,
				isDirty: false,
				lastSaved: new Date()
			}));
		} else {
			update((s) => ({ ...s, error }));
		}

		return error;
	}

	return {
		subscribe,

		/**
		 * Initialize store - load from localStorage
		 */
		initialize(): void {
			console.log('[Dashboard Layout Store] Initializing...');

			update((s) => ({ ...s, loading: true }));

			const layouts = loadLayoutsFromStorage();
			const activeId = loadActiveLayoutId() || layouts[0]?.id || null;

			// Ensure active layout exists
			const activeLayout = layouts.find((l) => l.id === activeId);
			const finalActiveId = activeLayout ? activeId : layouts[0]?.id || null;

			if (finalActiveId) {
				saveActiveLayoutId(finalActiveId);
			}

			update((s) => ({
				...s,
				layouts,
				activeLayoutId: finalActiveId,
				loading: false,
				error: null,
				isDirty: false,
				lastSaved: new Date()
			}));

			console.log('[Dashboard Layout Store] Initialized with', layouts.length, 'layouts');
		},

		/**
		 * Get active layout
		 */
		getActiveLayout(): DashboardLayout | null {
			const state = get({ subscribe });
			return state.layouts.find((l) => l.id === state.activeLayoutId) || null;
		},

		/**
		 * Update widget position
		 */
		updateWidgetPosition(widgetId: string, x: number, y: number): void {
			update((s) => {
				const layout = s.layouts.find((l) => l.id === s.activeLayoutId);
				if (!layout) return s;

				const widgetIndex = layout.widgets.findIndex((w) => w.id === widgetId);
				if (widgetIndex === -1) return s;

				layout.widgets[widgetIndex] = {
					...layout.widgets[widgetIndex],
					x,
					y
				};

				layout.updatedAt = new Date();

				scheduleAutoSave();

				return { ...s, isDirty: true, layouts: [...s.layouts] };
			});
		},

		/**
		 * Update widget size
		 */
		updateWidgetSize(widgetId: string, width: number, height: number): void {
			update((s) => {
				const layout = s.layouts.find((l) => l.id === s.activeLayoutId);
				if (!layout) return s;

				const widgetIndex = layout.widgets.findIndex((w) => w.id === widgetId);
				if (widgetIndex === -1) return s;

				layout.widgets[widgetIndex] = {
					...layout.widgets[widgetIndex],
					width,
					height
				};

				layout.updatedAt = new Date();

				scheduleAutoSave();

				return { ...s, isDirty: true, layouts: [...s.layouts] };
			});
		},

		/**
		 * Update viewport (pan/zoom)
		 */
		updateViewport(viewport: Partial<CanvasViewport>): void {
			update((s) => {
				const layout = s.layouts.find((l) => l.id === s.activeLayoutId);
				if (!layout) return s;

				layout.viewport = {
					...layout.viewport,
					...viewport
				};

				layout.updatedAt = new Date();

				scheduleAutoSave();

				return { ...s, isDirty: true, layouts: [...s.layouts] };
			});
		},

		/**
		 * Update grid configuration
		 */
		updateGridConfig(config: Partial<GridConfig>): void {
			update((s) => {
				const layout = s.layouts.find((l) => l.id === s.activeLayoutId);
				if (!layout) return s;

				layout.gridConfig = {
					...layout.gridConfig,
					...config
				};

				layout.updatedAt = new Date();

				scheduleAutoSave();

				return { ...s, isDirty: true, layouts: [...s.layouts] };
			});
		},

		/**
		 * Add new widget to canvas
		 */
		addWidget(widget: Omit<WidgetLayout, 'id' | 'zIndex'>): string {
			const newId = `w${Date.now()}-${Math.random().toString(36).slice(2, 9)}`;

			update((s) => {
				const layout = s.layouts.find((l) => l.id === s.activeLayoutId);
				if (!layout) return s;

				const maxZ = Math.max(...layout.widgets.map((w) => w.zIndex), 0);

				layout.widgets.push({
					...widget,
					id: newId,
					zIndex: maxZ + 1
				});

				layout.updatedAt = new Date();

				scheduleAutoSave();

				return { ...s, isDirty: true, layouts: [...s.layouts] };
			});

			return newId;
		},

		/**
		 * Remove widget from canvas
		 */
		removeWidget(widgetId: string): void {
			update((s) => {
				const layout = s.layouts.find((l) => l.id === s.activeLayoutId);
				if (!layout) return s;

				layout.widgets = layout.widgets.filter((w) => w.id !== widgetId);
				layout.updatedAt = new Date();

				scheduleAutoSave();

				return { ...s, isDirty: true, layouts: [...s.layouts] };
			});
		},

		/**
		 * Bring widget to front
		 */
		bringToFront(widgetId: string): void {
			update((s) => {
				const layout = s.layouts.find((l) => l.id === s.activeLayoutId);
				if (!layout) return s;

				const maxZ = Math.max(...layout.widgets.map((w) => w.zIndex), 0);
				const widgetIndex = layout.widgets.findIndex((w) => w.id === widgetId);
				if (widgetIndex === -1) return s;

				layout.widgets[widgetIndex].zIndex = maxZ + 1;

				return { ...s, layouts: [...s.layouts] };
			});
		},

		/**
		 * Force save (bypass debounce)
		 */
		forceSave(): PersistenceError | null {
			if (autoSaveTimer) {
				clearTimeout(autoSaveTimer);
				autoSaveTimer = null;
			}
			return saveToStorage();
		},

		/**
		 * Reset to defaults
		 */
		reset(): void {
			clearStorage();
			update((s) => ({
				...s,
				layouts: [DEFAULT_LAYOUT],
				activeLayoutId: DEFAULT_LAYOUT.id,
				error: null,
				isDirty: false,
				lastSaved: new Date()
			}));
			console.log('[Dashboard Layout Store] Reset to defaults');
		},

		/**
		 * Clear error
		 */
		clearError(): void {
			update((s) => ({ ...s, error: null }));
		}
	};
}

// ============================================================================
// EXPORTS
// ============================================================================

export const dashboardLayoutStore = createDashboardLayoutStore();

// Derived stores for convenience
export const activeLayout = derived(
	dashboardLayoutStore,
	($store) => $store.layouts.find((l) => l.id === $store.activeLayoutId) || null
);

export const activeViewport = derived(
	activeLayout,
	($layout) => $layout?.viewport || DEFAULT_VIEWPORT
);

export const activeGridConfig = derived(
	activeLayout,
	($layout) => $layout?.gridConfig || DEFAULT_GRID_CONFIG
);

export const layouts = derived(dashboardLayoutStore, ($store) => $store.layouts);

export const isDirty = derived(dashboardLayoutStore, ($store) => $store.isDirty);

export const persistenceError = derived(dashboardLayoutStore, ($store) => $store.error);
