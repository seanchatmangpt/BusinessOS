<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { useSession } from '$lib/auth-client';
	import { fade, fly, scale } from 'svelte/transition';
	import { flip } from 'svelte/animate';
	import { DropdownMenu } from 'bits-ui';
	import {
		DashboardHeader,
		TodaysFocusWidget,
		QuickActionsWidget,
		ActiveProjectsWidget,
		MyTasksWidget,
		RecentActivityWidget,
		AnalyticsSidepanel,
		MetricCardWidget,
		InsightsPanelWidget,
		ProductivityChartWidget,
		SmartNotificationsWidget,
		InfiniteCanvas,
		CanvasWidget
	} from '$lib/components/dashboard';
	import Tooltip from '$lib/components/ui/Tooltip.svelte';
	import { NotificationDropdown } from '$lib/components/notifications';
	import {
		api,
		type DashboardProject,
		type DashboardTask,
		type DashboardActivity
	} from '$lib/api';
	import {
		dashboardLayoutStore,
		activeLayout,
		activeViewport,
		activeGridConfig,
		type WidgetLayout
	} from '$lib/stores/dashboardLayoutStore';

	const session = useSession();

	// Initialize dashboard layout store
	onMount(() => {
		dashboardLayoutStore.initialize();
	});

	// ============================================================================
	// WIDGET PICKER STATE
	// ============================================================================

	type WidgetType = 'focus' | 'quick-actions' | 'projects' | 'tasks' | 'activity' | 'metric' | 'insights' | 'productivity-chart' | 'notifications';
	type WidgetSize = 'small' | 'medium' | 'large';

	// Edit mode state
	let isEditMode = $state(false);
	let showWidgetPicker = $state(false);
	let pickerSelectedSize = $state<WidgetSize>('medium');
	
	// Analytics state
	let showAnalyticsSidepanel = $state(false);
	let analyticsLoading = $state(false);
	let analyticsTimeRange = $state<'today' | 'week' | 'month' | '30days'>('week');
	
	// Seeded analytics data (mock data for demonstration)
	let seededAnalytics = $state<{
		focus: { completionRate: number; completedToday: number; totalToday: number; streak: number; avgCompletionTime: string; weeklyData: number[] };
		tasks: { completedThisWeek: number; dueToday: number; overdue: number; completionRate: number; byPriority: { critical: number; high: number; medium: number; low: number }; weeklyData: number[] };
		projects: { active: number; completed: number; atRisk: number; onTimeRate: number; avgProgress: number };
		activity: { totalActions: number; mostActiveDay: string; topActivityType: string; weeklyData: number[] };
	} | null>({
		focus: {
			completionRate: 78,
			completedToday: 3,
			totalToday: 4,
			streak: 7,
			avgCompletionTime: '2.3 hrs',
			weeklyData: [4, 5, 6, 4, 3, 2, 0] // Mon-Sun
		},
		tasks: {
			completedThisWeek: 23,
			dueToday: 5,
			overdue: 2,
			completionRate: 82,
			byPriority: { critical: 2, high: 5, medium: 8, low: 3 },
			weeklyData: [5, 6, 8, 4, 0, 0, 0]
		},
		projects: {
			active: 3,
			completed: 2,
			atRisk: 1,
			onTimeRate: 85,
			avgProgress: 67
		},
		activity: {
			totalActions: 47,
			mostActiveDay: 'Wednesday',
			topActivityType: 'task_completed',
			weeklyData: [12, 15, 18, 8, 0, 0, 0]
		}
	});
	
	// Handle time range change with loading simulation
	async function handleAnalyticsTimeRangeChange(range: 'today' | 'week' | 'month' | '30days') {
		analyticsTimeRange = range;
		analyticsLoading = true;
		
		// TODO: Replace with actual API call
		// const response = await fetch(`/api/analytics?range=${range}`);
		// seededAnalytics = await response.json();
		
		// Simulate API call delay
		await new Promise(resolve => setTimeout(resolve, 600));
		
		// Mock different data based on range
		const multiplier = range === 'today' ? 0.3 : range === 'week' ? 1 : range === 'month' ? 2.5 : 4;
		seededAnalytics = {
			focus: {
				completionRate: Math.min(100, Math.round(78 * (range === 'today' ? 0.8 : 1))),
				completedToday: range === 'today' ? 2 : 3,
				totalToday: 4,
				streak: 7,
				avgCompletionTime: '2.3 hrs',
				weeklyData: range === 'today' ? [0, 0, 0, 0, 0, 0, 2] : [4, 5, 6, 4, 3, 2, 0]
			},
			tasks: {
				completedThisWeek: Math.round(23 * multiplier),
				dueToday: 5,
				overdue: range === 'today' ? 1 : 2,
				completionRate: Math.min(100, Math.round(82 * (0.9 + Math.random() * 0.2))),
				byPriority: { critical: 2, high: 5, medium: 8, low: 3 },
				weeklyData: range === 'today' ? [0, 0, 0, 0, 0, 0, 3] : [5, 6, 8, 4, 0, 0, 0].map(v => Math.round(v * multiplier))
			},
			projects: {
				active: 3,
				completed: Math.round(2 * multiplier),
				atRisk: 1,
				onTimeRate: Math.min(100, Math.round(85 * (0.9 + Math.random() * 0.2))),
				avgProgress: Math.min(100, Math.round(67 + (range === 'month' ? 10 : range === '30days' ? 15 : 0)))
			},
			activity: {
				totalActions: Math.round(47 * multiplier),
				mostActiveDay: range === 'today' ? 'Today' : 'Wednesday',
				topActivityType: 'task_completed',
				weeklyData: range === 'today' ? [0, 0, 0, 0, 0, 0, 8] : [12, 15, 18, 8, 0, 0, 0].map(v => Math.round(v * multiplier))
			}
		};
		
		analyticsLoading = false;
	}
	
	// Available widget types for picker
	const availableWidgets: { type: WidgetType; title: string; description: string; icon: string }[] = [
		{ type: 'focus', title: "Today's Focus", description: 'Track your daily priorities', icon: '🎯' },
		{ type: 'quick-actions', title: 'Quick Actions', description: 'Common shortcuts', icon: '⚡' },
		{ type: 'notifications', title: 'Smart Alerts', description: 'Important notifications', icon: '🔔' },
		{ type: 'insights', title: 'Insights', description: 'AI-generated insights', icon: '⚡' },
		{ type: 'productivity-chart', title: 'Productivity Chart', description: 'Visual analytics', icon: '📊' },
		{ type: 'projects', title: 'Active Projects', description: 'Project progress overview', icon: '📁' },
		{ type: 'tasks', title: 'My Tasks', description: 'Tasks due soon', icon: '✓' },
		{ type: 'activity', title: 'Recent Activity', description: 'Latest workspace activity', icon: '📊' },
		{ type: 'metric', title: 'Metric Card', description: 'Single KPI display', icon: '📈' },
	];

	// Dashboard state
	let energyLevel = $state<number | null>(null);
	let isLoading = $state(true);
	let error = $state<string | null>(null);

	// Focus items from API
	let focusItems = $state<{ id: string; text: string; completed: boolean }[]>([]);

	// Projects, tasks, activities from API (using API types directly)
	let projects = $state<DashboardProject[]>([]);
	let tasks = $state<DashboardTask[]>([]);

	let activities = $state<
		{
			id: string;
			type:
				| 'task_completed'
				| 'task_started'
				| 'project_created'
				| 'project_updated'
				| 'conversation'
				| 'team'
				| 'artifact';
			description: string;
			actorName?: string;
			actorAvatar?: string;
			targetId?: string;
			targetType?: string;
			createdAt: string;
		}[]
	>([]);

	// Load dashboard data
	async function loadDashboard() {
		try {
			isLoading = true;
			error = null;

			const summary = await api.getDashboardSummary();

			// Transform focus items
			focusItems = summary.focus_items.map((item) => ({
				id: item.id,
				text: item.text,
				completed: item.completed
			}));

			// Keep API data as-is (widgets already updated to expect snake_case)
			projects = summary.projects;
			tasks = summary.tasks;

			// Transform activities
			activities = summary.activities.map((a) => ({
				id: a.id,
				type: a.type,
				description: a.description,
				actorName: a.actor_name ?? undefined,
				actorAvatar: a.actor_avatar ?? undefined,
				targetId: a.target_id ?? undefined,
				targetType: a.target_type ?? undefined,
				createdAt: a.created_at
			}));

			energyLevel = summary.energy_level;
		} catch (err) {
			console.error('Failed to load dashboard:', err);
			error = err instanceof Error ? err.message : 'Failed to load dashboard';
		} finally {
			isLoading = false;
		}
	}

	// Load on mount
	$effect(() => {
		if ($session.data) {
			loadDashboard();
		}
	});

	// ============================================================================
	// EDIT MODE FUNCTIONS
	// ============================================================================
	
	function toggleEditMode() {
		isEditMode = !isEditMode;
		if (!isEditMode) {
			showWidgetPicker = false;
		}
	}

	function addWidget(type: WidgetType) {
		const template = availableWidgets.find(w => w.type === type);
		if (!template) return;

		// Calculate position for new widget (center of current viewport)
		const viewport = $activeViewport;
		if (!viewport) return;

		// Get widget size based on picker selection
		const sizeMap = {
			small: { width: 400, height: 350 },
			medium: { width: 600, height: 500 },
			large: { width: 900, height: 550 }
		};
		const { width, height } = sizeMap[pickerSelectedSize];

		// Place widget in center of current view
		const x = -viewport.offsetX / viewport.zoom + (window.innerWidth / 2 / viewport.zoom) - width / 2;
		const y = -viewport.offsetY / viewport.zoom + (window.innerHeight / 2 / viewport.zoom) - height / 2;

		// Add widget using the store
		dashboardLayoutStore.addWidget({
			type,
			title: template.title,
			x: Math.max(50, x), // Ensure minimum offset
			y: Math.max(50, y),
			width,
			height
		});

		showWidgetPicker = false;
		pickerSelectedSize = 'medium';
	}

	// Keyboard shortcuts
	function handleKeydown(e: KeyboardEvent) {
		// Don't trigger if typing in an input
		if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) return;

		// Toggle edit mode with 'E' key
		if (e.key === 'e' || e.key === 'E') {
			e.preventDefault();
			toggleEditMode();
		}

		// Exit edit mode with Escape
		if (e.key === 'Escape' && isEditMode) {
			e.preventDefault();
			isEditMode = false;
			showWidgetPicker = false;
		}
	}
	
	// Zoom state and handlers
	let canvasZoom = $derived($activeViewport?.zoom ?? 1);
	const MIN_ZOOM = 0.25; // Allow zooming out to 25% to see more
	const MAX_ZOOM = 3.0;  // Allow zooming in to 300% for details

	function handleZoomIn() {
		if (!$activeViewport) return;
		const newZoom = Math.min($activeViewport.zoom + 0.15, MAX_ZOOM);
		dashboardLayoutStore.updateViewport({ ...$activeViewport, zoom: newZoom });
	}

	function handleZoomOut() {
		if (!$activeViewport) return;
		const newZoom = Math.max($activeViewport.zoom - 0.15, MIN_ZOOM);
		dashboardLayoutStore.updateViewport({ ...$activeViewport, zoom: newZoom });
	}

	function handleResetView() {
		dashboardLayoutStore.updateViewport({ offsetX: 0, offsetY: 0, zoom: 1.0 });
	}

	function handleFitAll() {
		// Fit all widgets in view
		if (!$activeLayout || !$activeViewport || $activeLayout.widgets.length === 0) {
			handleResetView();
			return;
		}

		const widgets = $activeLayout.widgets;

		// Calculate bounding box of all widgets
		const minX = Math.min(...widgets.map((w) => w.x));
		const minY = Math.min(...widgets.map((w) => w.y));
		const maxX = Math.max(...widgets.map((w) => w.x + w.width));
		const maxY = Math.max(...widgets.map((w) => w.y + w.height));

		const contentWidth = maxX - minX;
		const contentHeight = maxY - minY;

		// Use window size as approximation (InfiniteCanvas will be similar)
		const viewportWidth = window.innerWidth - 100; // Account for padding
		const viewportHeight = window.innerHeight - 200; // Account for header
		const padding = 100;

		// Calculate zoom to fit
		const zoomX = (viewportWidth - padding * 2) / contentWidth;
		const zoomY = (viewportHeight - padding * 2) / contentHeight;
		const newZoom = Math.min(Math.max(Math.min(zoomX, zoomY), MIN_ZOOM), MAX_ZOOM);

		// Center the content
		const centerX = minX + contentWidth / 2;
		const centerY = minY + contentHeight / 2;
		const newOffsetX = viewportWidth / 2 - centerX * newZoom;
		const newOffsetY = viewportHeight / 2 - centerY * newZoom;

		dashboardLayoutStore.updateViewport({
			offsetX: newOffsetX,
			offsetY: newOffsetY,
			zoom: newZoom
		});
	}

	// Quick action handlers
	function handleQuickAction(action: string) {
		switch (action) {
			case 'new-task':
				goto('/tasks?new=true');
				break;
			case 'new-project':
				goto('/projects?new=true');
				break;
			case 'new-chat':
				goto('/chat?new=true');
				break;
			case 'daily-log':
				goto('/daily');
				break;
		}
	}

	// Focus item handlers
	async function handleFocusToggle(id: string) {
		const item = focusItems.find((i) => i.id === id);
		if (!item) return;

		try {
			await api.updateFocusItem(id, { completed: !item.completed });
			focusItems = focusItems.map((i) =>
				i.id === id ? { ...i, completed: !i.completed } : i
			);
		} catch (err) {
			console.error('Failed to toggle focus item:', err);
		}
	}

	async function handleFocusAdd(text: string) {
		try {
			const newItem = await api.createFocusItem(text);
			focusItems = [
				...focusItems,
				{ id: newItem.id, text: newItem.text, completed: newItem.completed }
			];
		} catch (err) {
			console.error('Failed to add focus item:', err);
		}
	}

	async function handleFocusRemove(id: string) {
		try {
			await api.deleteFocusItem(id);
			focusItems = focusItems.filter((item) => item.id !== id);
		} catch (err) {
			console.error('Failed to remove focus item:', err);
		}
	}

	function handleFocusEdit() {
		// TODO: Implement focus edit mode
		console.log('Edit focus items');
	}

	// Task handlers
	async function handleTaskToggle(id: string) {
		try {
			await api.toggleTask(id);
			tasks = tasks.map((task) =>
				task.id === id ? { ...task, completed: !task.completed } : task
			);
		} catch (err) {
			console.error('Failed to toggle task:', err);
		}
	}

	// Energy check handler
	function handleEnergySet(level: number) {
		energyLevel = level;
		// TODO: Save to backend
	}
</script>

<svelte:window onkeydown={handleKeydown} />

<div class="h-full flex flex-col bg-gray-50 dark:bg-gray-900">
	
	{#if isLoading}
		<div class="flex-1 flex items-center justify-center" in:fade>
			<div class="flex flex-col items-center gap-3">
				<div class="animate-spin h-8 w-8 border-2 border-gray-900 border-t-transparent rounded-full"></div>
				<p class="text-sm text-gray-500">Loading dashboard...</p>
			</div>
		</div>
	{:else if error}
		<div class="flex-1 flex items-center justify-center" in:fade>
			<div class="text-center bg-white p-8 rounded-2xl shadow-sm border border-gray-200 max-w-md">
				<div class="w-16 h-16 bg-red-50 rounded-xl flex items-center justify-center mx-auto mb-4">
					<svg class="w-8 h-8 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
					</svg>
				</div>
				<p class="text-gray-600 mb-4">{error}</p>
				<button onclick={() => loadDashboard()} class="btn-pill btn-pill-primary">
					Try Again
				</button>
			</div>
		</div>
	{:else}
		<div class="flex-1 flex flex-col relative" in:fade={{ duration: 300 }}>
			<!-- Compact Header -->
			<div class="px-6 pt-4 pb-3 flex-shrink-0 border-b border-gray-100">
				<div class="flex items-center justify-between">
					<!-- Left: Title -->
					<div>
						<h1 class="text-xl font-semibold text-gray-900">Dashboard</h1>
						<p class="text-xs text-gray-500 mt-0.5">{new Date().toLocaleDateString('en-US', { weekday: 'long', month: 'long', day: 'numeric', year: 'numeric' })}</p>
					</div>

					<!-- Right: Actions -->
					<div class="flex items-center gap-3">
						<!-- View/Edit Toggle -->
						<div class="flex items-center gap-2 px-1 py-1 bg-gray-50 rounded-lg border border-gray-200">
							<button
								onclick={() => isEditMode && toggleEditMode()}
								class="px-3 py-1.5 rounded-md text-sm font-medium transition-all
									{!isEditMode
										? 'bg-white text-gray-900 shadow-sm'
										: 'text-gray-600 hover:text-gray-900'}"
							>
								View
							</button>
							<button
								onclick={() => !isEditMode && toggleEditMode()}
								class="px-3 py-1.5 rounded-md text-sm font-medium transition-all
									{isEditMode
										? 'bg-white text-blue-600 shadow-sm'
										: 'text-gray-600 hover:text-gray-900'}"
							>
								Edit
							</button>
						</div>

						<!-- Zoom Controls -->
						<div class="flex items-center gap-1.5 px-2 py-1 bg-gray-50/50 rounded-xl border border-gray-200/50">
							<Tooltip text="Zoom In" position="bottom">
								<button
									onclick={handleZoomIn}
									disabled={canvasZoom >= MAX_ZOOM}
									class="btn-pill btn-pill-icon btn-pill-sm btn-pill-ghost disabled:opacity-30 disabled:cursor-not-allowed"
									aria-label="Zoom in"
								>
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2.5">
										<path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4" />
									</svg>
								</button>
							</Tooltip>

							<div class="px-2 py-1 text-xs font-bold text-gray-900 min-w-[45px] text-center" title="Current zoom level">
								{Math.round(canvasZoom * 100)}%
							</div>

							<Tooltip text="Zoom Out" position="bottom">
								<button
									onclick={handleZoomOut}
									disabled={canvasZoom <= MIN_ZOOM}
									class="btn-pill btn-pill-icon btn-pill-sm btn-pill-ghost disabled:opacity-30 disabled:cursor-not-allowed"
									aria-label="Zoom out"
								>
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2.5">
										<path stroke-linecap="round" stroke-linejoin="round" d="M20 12H4" />
									</svg>
								</button>
							</Tooltip>

							<div class="w-px h-5 bg-gray-200 mx-0.5"></div>

							<Tooltip text="Reset View" position="bottom">
								<button
									onclick={handleResetView}
									class="btn-pill btn-pill-icon btn-pill-sm btn-pill-ghost"
									aria-label="Reset view"
								>
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
										<path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
									</svg>
								</button>
							</Tooltip>

							<Tooltip text="Fit All Widgets" position="bottom">
								<button
									onclick={handleFitAll}
									class="btn-pill btn-pill-icon btn-pill-sm btn-pill-ghost"
									aria-label="Fit all widgets"
								>
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
										<path stroke-linecap="round" stroke-linejoin="round" d="M4 8V4m0 0h4M4 4l5 5m11-1V4m0 0h-4m4 0l-5 5M4 16v4m0 0h4m-4 0l5-5m11 5l-5-5m5 5v-4m0 4h-4" />
									</svg>
								</button>
							</Tooltip>
						</div>

						{#if isEditMode}
							<button
								onclick={() => showWidgetPicker = true}
								class="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium rounded-lg transition-colors flex items-center gap-2"
							>
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
								</svg>
								Add Widget
							</button>
						{/if}

						<NotificationDropdown />
					</div>
				</div>
			</div>

			<!-- Infinite Canvas -->
			<div class="flex-1 overflow-hidden min-h-0">
				{#if $activeLayout}
					<InfiniteCanvas
						widgets={$activeLayout.widgets}
						viewport={$activeViewport}
						gridConfig={$activeGridConfig}
						{isEditMode}
						onViewportChange={(v) => dashboardLayoutStore.updateViewport(v)}
						onZoomIn={handleZoomIn}
						onZoomOut={handleZoomOut}
						onResetView={handleResetView}
						onFitAll={handleFitAll}
					>
						{#snippet children(widget: WidgetLayout)}
							<CanvasWidget
								{widget}
								{isEditMode}
								zoom={$activeViewport.zoom}
								snapToGrid={$activeGridConfig.snapToGrid}
								gridSize={$activeGridConfig.cellSize}
								onMove={(x, y) => dashboardLayoutStore.updateWidgetPosition(widget.id, x, y)}
								onResize={(w, h) => dashboardLayoutStore.updateWidgetSize(widget.id, w, h)}
								onClick={() => dashboardLayoutStore.bringToFront(widget.id)}
								onRemove={() => dashboardLayoutStore.removeWidget(widget.id)}
							>
								<!-- Widget content based on type -->
								{#if widget.type === 'focus'}
									<TodaysFocusWidget
										items={focusItems}
										onToggle={handleFocusToggle}
										onAdd={handleFocusAdd}
										onRemove={handleFocusRemove}
										onEdit={handleFocusEdit}
									/>
								{:else if widget.type === 'quick-actions'}
									<QuickActionsWidget onAction={handleQuickAction} />
								{:else if widget.type === 'projects'}
									<ActiveProjectsWidget
										{projects}
										onViewAll={() => goto('/projects')}
									/>
								{:else if widget.type === 'tasks'}
									<MyTasksWidget
										{tasks}
										onToggle={handleTaskToggle}
										onViewAll={() => goto('/tasks')}
									/>
								{:else if widget.type === 'activity'}
									<RecentActivityWidget {activities} onViewAll={() => goto('/chat')} />
								{:else if widget.type === 'metric'}
									<MetricCardWidget
										title="Tasks Completed"
										value={tasks.filter(t => t.completed).length}
										previousValue={tasks.filter(t => t.completed).length - 2}
										color="blue"
										sparklineData={[8, 12, 10, 15, 18, 14, tasks.filter(t => t.completed).length]}
									/>
								{:else if widget.type === 'insights'}
									<InsightsPanelWidget
										{tasks}
										{projects}
										onAction={handleQuickAction}
									/>
								{:else if widget.type === 'productivity-chart'}
									<ProductivityChartWidget
										title="Weekly Productivity"
										type="bar"
										color="blue"
									/>
								{:else if widget.type === 'notifications'}
									<SmartNotificationsWidget
										{tasks}
										{projects}
										onAction={handleQuickAction}
										onViewAll={() => goto('/notifications')}
									/>
								{/if}
							</CanvasWidget>
						{/snippet}
					</InfiniteCanvas>
				{/if}
			</div>
		</div>
	{/if}
	
	<!-- Widget Picker Drawer -->
	{#if showWidgetPicker}
		<!-- Backdrop -->
		<button
			class="fixed inset-0 bg-black/20 backdrop-blur-sm z-40"
			onclick={() => showWidgetPicker = false}
			transition:fade={{ duration: 200 }}
			aria-label="Close widget picker"
		></button>

		<!-- Drawer -->
		<div
			class="fixed bottom-0 left-0 right-0 bg-white rounded-t-3xl shadow-2xl z-50 max-h-[75vh] overflow-hidden border-t border-gray-100"
			transition:fly={{ y: 300, duration: 300 }}
		>
			<!-- Handle -->
			<div class="flex justify-center pt-4 pb-3">
				<div class="w-12 h-1.5 bg-gray-300 rounded-full"></div>
			</div>

			<!-- Header -->
			<div class="px-8 pb-6 border-b border-gray-100">
				<div class="flex items-center justify-between mb-5">
					<div>
						<h2 class="text-xl font-semibold text-gray-900">Add Widget</h2>
						<p class="text-sm text-gray-500 mt-1">Choose a widget to add to your dashboard</p>
					</div>
					<button
						onclick={() => showWidgetPicker = false}
						class="w-9 h-9 flex items-center justify-center text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
						aria-label="Close widget picker"
					>
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
							<path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
						</svg>
					</button>
				</div>

				<!-- Size Picker -->
				<div class="flex items-center gap-3">
					<span class="text-sm font-medium text-gray-600">Widget Size:</span>
					<div class="flex items-center gap-2 px-1 py-1 bg-gray-50 rounded-lg border border-gray-200">
						<button
							onclick={() => pickerSelectedSize = 'small'}
							class="px-4 py-1.5 rounded-md text-sm font-medium transition-all {pickerSelectedSize === 'small' ? 'bg-white text-gray-900 shadow-sm' : 'text-gray-600 hover:text-gray-900'}"
						>
							Small
						</button>
						<button
							onclick={() => pickerSelectedSize = 'medium'}
							class="px-4 py-1.5 rounded-md text-sm font-medium transition-all {pickerSelectedSize === 'medium' ? 'bg-white text-gray-900 shadow-sm' : 'text-gray-600 hover:text-gray-900'}"
						>
							Medium
						</button>
						<button
							onclick={() => pickerSelectedSize = 'large'}
							class="px-4 py-1.5 rounded-md text-sm font-medium transition-all {pickerSelectedSize === 'large' ? 'bg-white text-gray-900 shadow-sm' : 'text-gray-600 hover:text-gray-900'}"
						>
							Large
						</button>
					</div>
				</div>
			</div>

			<!-- Widget Grid -->
			<div class="p-8 overflow-y-auto max-h-[calc(75vh-200px)]">
				<div class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
					{#each availableWidgets as widgetOption}
						{@const existingCount = $activeLayout ? $activeLayout.widgets.filter(w => w.type === widgetOption.type).length : 0}
						<button
							onclick={() => addWidget(widgetOption.type)}
							class="group relative flex flex-col items-start p-5 rounded-xl transition-all duration-200 text-left bg-white hover:bg-blue-50 border-2 border-gray-200 hover:border-blue-400 hover:shadow-md"
						>
							{#if existingCount > 0}
								<div class="absolute top-3 right-3 flex items-center justify-center w-6 h-6 text-xs font-bold text-blue-700 bg-blue-100 rounded-full">
									{existingCount}
								</div>
							{/if}

							<div class="text-3xl mb-3">{widgetOption.icon}</div>
							<h3 class="font-semibold text-sm mb-1 text-gray-900 group-hover:text-blue-900">
								{widgetOption.title}
							</h3>
							<p class="text-xs leading-relaxed text-gray-500 group-hover:text-blue-700">
								{widgetOption.description}
							</p>
						</button>
					{/each}
				</div>
			</div>
		</div>
	{/if}

	<!-- Floating Analytics Button (FAB) -->
	{#if !isEditMode && !showWidgetPicker}
		<button
			onclick={() => showAnalyticsSidepanel = true}
			class="fixed bottom-6 right-6 z-30 w-12 h-12 bg-gray-900 hover:bg-gray-800 text-white rounded-xl shadow-md hover:shadow-lg flex items-center justify-center transition-all duration-200"
			transition:scale={{ duration: 200 }}
			title="View Dashboard Analytics"
		>
			<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
			</svg>
		</button>
	{/if}
	
	<!-- Analytics Sidepanel -->
	<AnalyticsSidepanel 
		isOpen={showAnalyticsSidepanel} 
		onClose={() => showAnalyticsSidepanel = false}
		analytics={seededAnalytics}
		isLoading={analyticsLoading}
		onTimeRangeChange={handleAnalyticsTimeRangeChange}
	/>
</div>
