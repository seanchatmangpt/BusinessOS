<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { projects } from '$lib/stores/projects';
	import { api, type ClientListResponse } from '$lib/api';
	import { Dialog, DropdownMenu } from 'bits-ui';
	import type { Project } from '$lib/api';
	import { Building2, Users, GraduationCap, FolderOpen, X, ChevronRight } from 'lucide-svelte';

	// Preloaded data from +page.ts load function (prefetched on hover)
	let { data } = $props();

	const embedSuffix = $derived($page.url.searchParams.get('embed') === 'true' ? '?embed=true' : '');

	let showNewProject = $state(false);
	let newProject = $state({
		name: '',
		description: '',
		client_name: '',
		project_type: 'internal',
		priority: 'medium' as 'low' | 'medium' | 'high' | 'critical'
	});
	let statusFilter = $state('');
	let typeFilter = $state('');
	let priorityFilter = $state('');
	let searchQuery = $state('');
	let viewMode = $state<'grid' | 'list' | 'kanban'>('grid');
	let groupByType = $state(false);
	let createError = $state('');

	// Clients for dropdown — seed from preloaded data
	let clients = $state<ClientListResponse[]>(data?.clients ?? []);
	let showAdvancedOptions = $state(false);

	onMount(async () => {
		// Seed the store with preloaded data if available (avoids redundant fetch)
		if (data?.projects?.length) {
			projects.setProjects(data.projects);
		} else {
			try {
				await projects.loadProjects();
			} catch {
				// Backend unavailable — empty state will show
			}
		}

		// Clients already seeded from preloaded data; refresh if empty
		if (!clients.length) {
			await loadClients();
		}
	});

	async function loadClients() {
		try {
			clients = await api.getClients();
		} catch (err) {
			console.error('Error loading clients:', err);
		}
	}

	// Filtered projects
	let filteredProjects = $derived((() => {
		let result = $projects.projects;

		if (typeFilter) {
			result = result.filter(p => p.project_type === typeFilter);
		}

		if (priorityFilter) {
			result = result.filter(p => p.priority === priorityFilter);
		}

		if (searchQuery) {
			const query = searchQuery.toLowerCase();
			result = result.filter(p =>
				p.name.toLowerCase().includes(query) ||
				(p.description && p.description.toLowerCase().includes(query)) ||
				(p.client_name && p.client_name.toLowerCase().includes(query))
			);
		}

		return result;
	})());

	// Stats
	let stats = $derived({
		total: $projects.projects.length,
		active: $projects.projects.filter(p => p.status === 'active').length,
		paused: $projects.projects.filter(p => p.status === 'paused').length,
		completed: $projects.projects.filter(p => p.status === 'completed').length
	});

	// Grouped by type
	let groupedByType = $derived({
		internal: filteredProjects.filter(p => p.project_type === 'internal'),
		client_work: filteredProjects.filter(p => p.project_type === 'client_work'),
		learning: filteredProjects.filter(p => p.project_type === 'learning'),
		other: filteredProjects.filter(p => !['internal', 'client_work', 'learning'].includes(p.project_type))
	});

	// Grouped by status for kanban
	let groupedByStatus = $derived({
		active: filteredProjects.filter(p => p.status === 'active'),
		paused: filteredProjects.filter(p => p.status === 'paused'),
		completed: filteredProjects.filter(p => p.status === 'completed')
	});

	async function handleCreateProject(e: Event) {
		e.preventDefault();
		createError = '';
		try {
			await projects.createProject(newProject);
			showNewProject = false;
			newProject = { name: '', description: '', client_name: '', project_type: 'internal', priority: 'medium' };
			showAdvancedOptions = false;
		} catch (error) {
			createError = (error as Error).message || 'Failed to create project';
		}
	}

	function getStatusColor(status: string) {
		switch (status) {
			case 'active': return 'prm-status prm-status--active';
			case 'paused': return 'prm-status prm-status--paused';
			case 'completed': return 'prm-status prm-status--completed';
			case 'archived': return 'prm-ls-status-default';
			default: return 'prm-ls-status-default';
		}
	}

	function getPriorityColor(priority: string) {
		switch (priority) {
			case 'critical': return 'prm-priority prm-priority--critical';
			case 'high': return 'prm-priority prm-priority--high';
			case 'medium': return 'prm-priority prm-priority--medium';
			case 'low': return 'prm-priority prm-priority--low';
			default: return 'prm-ls-priority-default';
		}
	}

	function getPriorityIcon(priority: string) {
		const count = priority === 'critical' ? 3 : priority === 'high' ? 2 : priority === 'medium' ? 1 : 0;
		return count;
	}

	function getTypeColor(type: string) {
		switch (type) {
			case 'internal': return 'prm-type prm-type--internal';
			case 'client_work': return 'prm-type prm-type--client';
			case 'learning': return 'prm-type prm-type--learning';
			default: return 'prm-ls-type-default';
		}
	}

	function getTypeLabel(type: string) {
		switch (type) {
			case 'internal': return 'Internal';
			case 'client_work': return 'Client Work';
			case 'learning': return 'Learning';
			default: return type;
		}
	}

	function formatDate(dateStr: string) {
		return new Date(dateStr).toLocaleDateString(undefined, { month: 'short', day: 'numeric' });
	}

	function clearFilters() {
		statusFilter = '';
		typeFilter = '';
		priorityFilter = '';
		searchQuery = '';
		projects.loadProjects();
	}

	// Type and Priority filter options
	const typeOptions = [
		{ value: '', label: 'All Types' },
		{ value: 'internal', label: 'Internal' },
		{ value: 'client_work', label: 'Client Work' },
		{ value: 'learning', label: 'Learning' }
	];

	const priorityOptions = [
		{ value: '', label: 'All Priorities' },
		{ value: 'critical', label: 'Critical' },
		{ value: 'high', label: 'High' },
		{ value: 'medium', label: 'Medium' },
		{ value: 'low', label: 'Low' }
	];

	let hasActiveFilters = $derived(statusFilter || typeFilter || priorityFilter || searchQuery);
</script>

<div class="h-full flex flex-col prm-ls-page">
	<!-- Header -->
	<div class="px-6 py-4 prm-ls-bar">
		<div class="flex items-center justify-between mb-4">
			<div>
				<h1 class="text-xl font-semibold prm-ls-title">Projects</h1>
				<p class="text-sm prm-ls-muted mt-0.5">Manage your work and track progress</p>
			</div>
			<button onclick={() => showNewProject = true} class="btn-pill btn-pill-primary btn-pill-sm flex items-center gap-2">
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
				</svg>
				New Project
			</button>
		</div>

		<!-- Stats Row -->
		<div class="grid grid-cols-4 gap-3 mt-3">
			<div class="prm-ls-stat">
				<div class="prm-ls-stat__icon prm-ls-stat__icon--total">
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
					</svg>
				</div>
				<div class="prm-ls-stat__body">
					<span class="prm-ls-stat__value">{stats.total}</span>
					<span class="prm-ls-stat__label">Total</span>
				</div>
			</div>
			<div class="prm-ls-stat">
				<div class="prm-ls-stat__icon prm-ls-stat__icon--active">
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
					</svg>
				</div>
				<div class="prm-ls-stat__body">
					<span class="prm-ls-stat__value">{stats.active}</span>
					<span class="prm-ls-stat__label">Active</span>
				</div>
			</div>
			<div class="prm-ls-stat">
				<div class="prm-ls-stat__icon prm-ls-stat__icon--paused">
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 9v6m4-6v6m7-3a9 9 0 11-18 0 9 9 0 0118 0z" />
					</svg>
				</div>
				<div class="prm-ls-stat__body">
					<span class="prm-ls-stat__value">{stats.paused}</span>
					<span class="prm-ls-stat__label">Paused</span>
				</div>
			</div>
			<div class="prm-ls-stat">
				<div class="prm-ls-stat__icon prm-ls-stat__icon--completed">
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
					</svg>
				</div>
				<div class="prm-ls-stat__body">
					<span class="prm-ls-stat__value">{stats.completed}</span>
					<span class="prm-ls-stat__label">Completed</span>
				</div>
			</div>
		</div>
	</div>

	<!-- Filters & Controls Bar -->
	<div class="px-6 py-3 prm-ls-bar flex items-center gap-3 flex-wrap">
		<!-- Search -->
		<div class="relative flex-1 min-w-[200px] max-w-xs">
			<svg class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 prm-ls-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
			</svg>
			<input
				type="text"
				bind:value={searchQuery}
				placeholder="Search projects..."
				class="w-full pl-9 pr-3 py-1.5 text-sm prm-ls-search rounded-lg focus:outline-none focus:ring-2 focus:border-transparent"
			/>
		</div>

		<!-- Status Filter Pills -->
		<div class="flex items-center gap-1 prm-ls-divider-l pl-3">
			<button
				onclick={() => { statusFilter = ''; projects.loadProjects(); }}
				class="btn-pill btn-pill-xs {statusFilter === '' ? 'btn-pill-primary' : 'btn-pill-ghost'}"
			>
				All
			</button>
			<button
				onclick={() => { statusFilter = 'active'; projects.loadProjects('active'); }}
				class="btn-pill btn-pill-xs {statusFilter === 'active' ? 'btn-pill-soft' : 'btn-pill-ghost'}"
			>
				Active
			</button>
			<button
				onclick={() => { statusFilter = 'paused'; projects.loadProjects('paused'); }}
				class="btn-pill btn-pill-xs {statusFilter === 'paused' ? 'btn-pill-soft' : 'btn-pill-ghost'}"
			>
				Paused
			</button>
			<button
				onclick={() => { statusFilter = 'completed'; projects.loadProjects('completed'); }}
				class="btn-pill btn-pill-xs {statusFilter === 'completed' ? 'btn-pill-soft' : 'btn-pill-ghost'}"
			>
				Completed
			</button>
		</div>

		<!-- Type Filter -->
		<DropdownMenu.Root>
			<DropdownMenu.Trigger class="btn-pill btn-pill-secondary btn-pill-sm flex items-center gap-2 {typeFilter ? 'prm-filter--active' : ''}">
				<span>{typeOptions.find(opt => opt.value === typeFilter)?.label || 'All Types'}</span>
				<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
				</svg>
			</DropdownMenu.Trigger>
			<DropdownMenu.Portal>
				<DropdownMenu.Content class="z-50 min-w-[160px] prm-ls-dropdown rounded-lg shadow-lg p-1" sideOffset={4}>
					{#each typeOptions as option}
						<DropdownMenu.Item
							class="px-3 py-2 text-xs rounded prm-ls-dropdown-item cursor-pointer transition-colors {typeFilter === option.value ? 'prm-dropdown-item--active' : ''}"
							onclick={() => typeFilter = option.value}
						>
							{option.label}
						</DropdownMenu.Item>
					{/each}
				</DropdownMenu.Content>
			</DropdownMenu.Portal>
		</DropdownMenu.Root>

		<!-- Priority Filter -->
		<DropdownMenu.Root>
			<DropdownMenu.Trigger class="btn-pill btn-pill-secondary btn-pill-sm flex items-center gap-2 {priorityFilter ? 'prm-filter--active' : ''}">
				<span>{priorityOptions.find(opt => opt.value === priorityFilter)?.label || 'All Priorities'}</span>
				<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
				</svg>
			</DropdownMenu.Trigger>
			<DropdownMenu.Portal>
				<DropdownMenu.Content class="z-50 min-w-[160px] prm-ls-dropdown rounded-lg shadow-lg p-1" sideOffset={4}>
					{#each priorityOptions as option}
						<DropdownMenu.Item
							class="px-3 py-2 text-xs rounded prm-ls-dropdown-item cursor-pointer transition-colors {priorityFilter === option.value ? 'prm-dropdown-item--active' : ''}"
							onclick={() => priorityFilter = option.value}
						>
							{option.label}
						</DropdownMenu.Item>
					{/each}
				</DropdownMenu.Content>
			</DropdownMenu.Portal>
		</DropdownMenu.Root>

		<!-- Clear Filters -->
		{#if hasActiveFilters}
			<button onclick={clearFilters} class="btn-pill btn-pill-ghost btn-pill-xs">
				Clear filters
			</button>
		{/if}

		<!-- Spacer -->
		<div class="flex-1"></div>

		<!-- Group by Type Toggle -->
		<label class="flex items-center gap-2 text-xs prm-ls-label cursor-pointer">
			<input
				type="checkbox"
				bind:checked={groupByType}
				class="w-3.5 h-3.5 rounded prm-ls-checkbox"
			/>
			Group by type
		</label>

		<!-- View Mode Toggle -->
		<div class="flex items-center gap-0.5 prm-ls-toggle-border rounded-lg overflow-hidden p-0.5">
			<button
				onclick={() => viewMode = 'grid'}
				class="btn-pill btn-pill-icon btn-pill-xs {viewMode === 'grid' ? 'btn-pill-primary' : 'btn-pill-ghost'}"
				title="Grid view"
				aria-label="Grid view"
			>
				<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
				</svg>
			</button>
			<button
				onclick={() => viewMode = 'list'}
				class="btn-pill btn-pill-icon btn-pill-xs {viewMode === 'list' ? 'btn-pill-primary' : 'btn-pill-ghost'}"
				title="List view"
				aria-label="List view"
			>
				<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
				</svg>
			</button>
			<button
				onclick={() => viewMode = 'kanban'}
				class="btn-pill btn-pill-icon btn-pill-xs {viewMode === 'kanban' ? 'btn-pill-primary' : 'btn-pill-ghost'}"
				title="Kanban view"
				aria-label="Kanban view"
			>
				<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 17V7m0 10a2 2 0 01-2 2H5a2 2 0 01-2-2V7a2 2 0 012-2h2a2 2 0 012 2m0 10a2 2 0 002 2h2a2 2 0 002-2M9 7a2 2 0 012-2h2a2 2 0 012 2m0 10V7m0 10a2 2 0 002 2h2a2 2 0 002-2V7a2 2 0 00-2-2h-2a2 2 0 00-2 2" />
				</svg>
			</button>
		</div>
	</div>

	<!-- Content -->
	<div class="flex-1 overflow-y-auto p-6">
		{#if $projects.loading}
			<div class="flex items-center justify-center h-48">
				<div class="animate-spin h-8 w-8 border-2 prm-ls-spinner rounded-full"></div>
			</div>
		{:else if $projects.projects.length === 0}
			<div class="flex flex-col items-center justify-center h-64 text-center">
				<div class="w-16 h-16 rounded-2xl prm-ls-empty-bg flex items-center justify-center mb-4">
					<svg class="w-8 h-8 prm-ls-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
					</svg>
				</div>
				<h3 class="text-lg font-medium prm-ls-title mb-1">No projects yet</h3>
				<p class="text-sm prm-ls-muted mb-4">Get started by creating your first project</p>
				<button onclick={() => showNewProject = true} class="btn-pill btn-pill-primary btn-pill-sm">
					Create Project
				</button>
			</div>
		{:else if filteredProjects.length === 0}
			<div class="flex flex-col items-center justify-center h-48 text-center">
				<p class="prm-ls-muted mb-2">No projects match your filters</p>
				<button onclick={clearFilters} class="btn-pill btn-pill-ghost btn-pill-sm underline">
					Clear all filters
				</button>
			</div>
		{:else if viewMode === 'kanban'}
			<!-- Kanban View -->
			<div class="flex gap-4 h-full min-h-0">
				<!-- Active Column -->
				<div class="flex-1 min-w-[280px] max-w-[350px] flex flex-col prm-ls-column overflow-hidden">
					<div class="px-4 py-3 prm-ls-column__header flex items-center gap-2">
						<span class="prm-dot prm-dot--active"></span>
						<span class="font-medium prm-ls-title">Active</span>
						<span class="prm-ls-kanban-count ml-auto">{groupedByStatus.active.length}</span>
					</div>
					<div class="flex-1 overflow-y-auto p-2 space-y-2">
						{#each groupedByStatus.active as project}
							<a href="/projects/{project.id}{embedSuffix}" class="block p-3 prm-ls-kanban-card rounded-lg transition-colors">
								<div class="flex items-start justify-between mb-1">
									<span class="text-sm font-medium prm-ls-title line-clamp-1">{project.name}</span>
									<span class="text-xs px-1.5 py-0.5 rounded {getTypeColor(project.project_type)}">{@render typeIcon(project.project_type, 12)}</span>
								</div>
								{#if project.client_name}
									<p class="text-xs prm-ls-muted mb-1">{project.client_name}</p>
								{/if}
								<div class="flex items-center justify-between text-xs prm-ls-icon">
									<span class={getPriorityColor(project.priority)}>
										{'●'.repeat(getPriorityIcon(project.priority) + 1)}
									</span>
									<span>{formatDate(project.updated_at)}</span>
								</div>
							</a>
						{/each}
						{#if groupedByStatus.active.length === 0}
							<p class="text-xs prm-ls-icon text-center py-4">No active projects</p>
						{/if}
					</div>
				</div>

				<!-- Paused Column -->
				<div class="flex-1 min-w-[280px] max-w-[350px] flex flex-col prm-ls-column overflow-hidden">
					<div class="px-4 py-3 prm-ls-column__header flex items-center gap-2">
						<span class="prm-dot prm-dot--paused"></span>
						<span class="font-medium prm-ls-title">Paused</span>
						<span class="prm-ls-kanban-count ml-auto">{groupedByStatus.paused.length}</span>
					</div>
					<div class="flex-1 overflow-y-auto p-2 space-y-2">
						{#each groupedByStatus.paused as project}
							<a href="/projects/{project.id}{embedSuffix}" class="block p-3 prm-ls-kanban-card rounded-lg transition-colors">
								<div class="flex items-start justify-between mb-1">
									<span class="text-sm font-medium prm-ls-title line-clamp-1">{project.name}</span>
									<span class="text-xs px-1.5 py-0.5 rounded {getTypeColor(project.project_type)}">{@render typeIcon(project.project_type, 12)}</span>
								</div>
								{#if project.client_name}
									<p class="text-xs prm-ls-muted mb-1">{project.client_name}</p>
								{/if}
								<div class="flex items-center justify-between text-xs prm-ls-icon">
									<span class={getPriorityColor(project.priority)}>
										{'●'.repeat(getPriorityIcon(project.priority) + 1)}
									</span>
									<span>{formatDate(project.updated_at)}</span>
								</div>
							</a>
						{/each}
						{#if groupedByStatus.paused.length === 0}
							<p class="text-xs prm-ls-icon text-center py-4">No paused projects</p>
						{/if}
					</div>
				</div>

				<!-- Completed Column -->
				<div class="flex-1 min-w-[280px] max-w-[350px] flex flex-col prm-ls-column overflow-hidden">
					<div class="px-4 py-3 prm-ls-column__header flex items-center gap-2">
						<span class="prm-dot prm-dot--completed"></span>
						<span class="font-medium prm-ls-title">Completed</span>
						<span class="prm-ls-kanban-count ml-auto">{groupedByStatus.completed.length}</span>
					</div>
					<div class="flex-1 overflow-y-auto p-2 space-y-2">
						{#each groupedByStatus.completed as project}
							<a href="/projects/{project.id}{embedSuffix}" class="block p-3 prm-ls-kanban-card rounded-lg transition-colors">
								<div class="flex items-start justify-between mb-1">
									<span class="text-sm font-medium prm-ls-title line-clamp-1">{project.name}</span>
									<span class="text-xs px-1.5 py-0.5 rounded {getTypeColor(project.project_type)}">{@render typeIcon(project.project_type, 12)}</span>
								</div>
								{#if project.client_name}
									<p class="text-xs prm-ls-muted mb-1">{project.client_name}</p>
								{/if}
								<div class="flex items-center justify-between text-xs prm-ls-icon">
									<span class={getPriorityColor(project.priority)}>
										{'●'.repeat(getPriorityIcon(project.priority) + 1)}
									</span>
									<span>{formatDate(project.updated_at)}</span>
								</div>
							</a>
						{/each}
						{#if groupedByStatus.completed.length === 0}
							<p class="text-xs prm-ls-icon text-center py-4">No completed projects</p>
						{/if}
					</div>
				</div>
			</div>
		{:else if viewMode === 'list'}
			<!-- List View -->
			<div class="prm-ls-column overflow-hidden">
				<table class="w-full">
					<thead class="prm-ls-table-head">
						<tr>
							<th class="text-left text-xs font-medium prm-ls-muted uppercase tracking-wider px-4 py-3">Project</th>
							<th class="text-left text-xs font-medium prm-ls-muted uppercase tracking-wider px-4 py-3">Type</th>
							<th class="text-left text-xs font-medium prm-ls-muted uppercase tracking-wider px-4 py-3">Status</th>
							<th class="text-left text-xs font-medium prm-ls-muted uppercase tracking-wider px-4 py-3">Priority</th>
							<th class="text-left text-xs font-medium prm-ls-muted uppercase tracking-wider px-4 py-3">Updated</th>
						</tr>
					</thead>
					<tbody class="prm-ls-table-body">
						{#each filteredProjects as project}
							<tr class="prm-ls-table-row transition-colors cursor-pointer" onclick={() => goto(`/projects/${project.id}${embedSuffix}`)}>
								<td class="px-4 py-3">
									<div>
										<span class="font-medium prm-ls-title">{project.name}</span>
										{#if project.client_name}
											<span class="prm-ls-icon ml-2">· {project.client_name}</span>
										{/if}
									</div>
									{#if project.description}
										<p class="text-sm prm-ls-muted line-clamp-1 mt-0.5">{project.description}</p>
									{/if}
								</td>
								<td class="px-4 py-3">
									<span class="text-xs px-2 py-1 rounded-full {getTypeColor(project.project_type)} prm-ls-type-inline">
										{@render typeIcon(project.project_type, 12)} {getTypeLabel(project.project_type)}
									</span>
								</td>
								<td class="px-4 py-3">
									<span class="text-xs font-medium px-2.5 py-1 rounded-full {getStatusColor(project.status)}">
										{project.status}
									</span>
								</td>
								<td class="px-4 py-3">
									<span class="text-xs font-medium {getPriorityColor(project.priority)} capitalize">
										{'●'.repeat(getPriorityIcon(project.priority) + 1)} {project.priority}
									</span>
								</td>
								<td class="px-4 py-3 text-sm prm-ls-muted">
									{formatDate(project.updated_at)}
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{:else if groupByType}
			<!-- Grid View Grouped by Type -->
			<div class="space-y-8">
				{#if groupedByType.internal.length > 0}
					<div>
						<div class="flex items-center gap-2 mb-3">
							<span class="prm-ls-type-icon">{@render typeIcon('internal', 18)}</span>
						<h2 class="font-semibold prm-ls-title">Internal Projects</h2>
						<span class="text-xs prm-ls-icon">({groupedByType.internal.length})</span>
						</div>
						<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
							{#each groupedByType.internal as project}
								{@render projectCard(project)}
							{/each}
						</div>
					</div>
				{/if}

				{#if groupedByType.client_work.length > 0}
					<div>
						<div class="flex items-center gap-2 mb-3">
							<span class="prm-ls-type-icon">{@render typeIcon('client_work', 18)}</span>
						<h2 class="font-semibold prm-ls-title">Client Work</h2>
						<span class="text-xs prm-ls-icon">({groupedByType.client_work.length})</span>
						</div>
						<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
							{#each groupedByType.client_work as project}
								{@render projectCard(project)}
							{/each}
						</div>
					</div>
				{/if}

				{#if groupedByType.learning.length > 0}
					<div>
						<div class="flex items-center gap-2 mb-3">
							<span class="prm-ls-type-icon">{@render typeIcon('learning', 18)}</span>
						<h2 class="font-semibold prm-ls-title">Learning</h2>
						<span class="text-xs prm-ls-icon">({groupedByType.learning.length})</span>
						</div>
						<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
							{#each groupedByType.learning as project}
								{@render projectCard(project)}
							{/each}
						</div>
					</div>
				{/if}

				{#if groupedByType.other.length > 0}
					<div>
						<div class="flex items-center gap-2 mb-3">
							<span class="prm-ls-type-icon">{@render typeIcon('other', 18)}</span>
						<h2 class="font-semibold prm-ls-title">Other</h2>
						<span class="text-xs prm-ls-icon">({groupedByType.other.length})</span>
						</div>
						<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
							{#each groupedByType.other as project}
								{@render projectCard(project)}
							{/each}
						</div>
					</div>
				{/if}
			</div>
		{:else}
			<!-- Standard Grid View -->
			<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
				{#each filteredProjects as project}
					{@render projectCard(project)}
				{/each}
			</div>
		{/if}
	</div>
</div>

{#snippet typeIcon(type: string, size: number)}
	{#if type === 'internal'}
		<Building2 {size} />
	{:else if type === 'client_work'}
		<Users {size} />
	{:else if type === 'learning'}
		<GraduationCap {size} />
	{:else}
		<FolderOpen {size} />
	{/if}
{/snippet}

{#snippet projectCard(project: Project)}
	<a href="/projects/{project.id}{embedSuffix}" class="block prm-ls-card prm-ls-card--{project.project_type} transition-all duration-200 cursor-pointer group">
		<div class="p-4">
			<div class="flex items-start justify-between mb-2">
				<div class="flex items-center gap-2 min-w-0">
					<span class="prm-ls-type-icon flex-shrink-0">{@render typeIcon(project.project_type, 16)}</span>
					<h3 class="font-medium prm-ls-title line-clamp-1 text-sm">{project.name}</h3>
				</div>
				<span class="text-xs font-medium px-2 py-0.5 rounded-full flex-shrink-0 ml-2 {getStatusColor(project.status)}">
					{project.status}
				</span>
			</div>
			{#if project.client_name}
				<p class="text-xs prm-ls-muted mb-1 ml-7">{project.client_name}</p>
			{/if}
			{#if project.description}
				<p class="text-xs prm-ls-icon line-clamp-2 mb-3 ml-7">{project.description}</p>
			{:else}
				<div class="mb-3"></div>
			{/if}
			<div class="flex items-center justify-between text-xs pt-3 prm-ls-card__footer">
				<div class="flex items-center gap-2">
					<span class="font-medium {getPriorityColor(project.priority)} capitalize">
						{'●'.repeat(getPriorityIcon(project.priority) + 1)} {project.priority}
					</span>
					<span class="px-1.5 py-0.5 rounded {getTypeColor(project.project_type)}">
						{getTypeLabel(project.project_type)}
					</span>
				</div>
				<span class="prm-ls-icon">{formatDate(project.updated_at)}</span>
			</div>
		</div>
	</a>
{/snippet}

<!-- New Project Dialog -->
<Dialog.Root bind:open={showNewProject}>
	<Dialog.Portal>
		<Dialog.Overlay class="dlg-np__overlay" />
		<Dialog.Content class="dlg-np__panel">
			<!-- Header -->
			<div class="dlg-np__header">
				<div>
					<Dialog.Title class="dlg-np__title">New Project</Dialog.Title>
					<p class="dlg-np__subtitle">Configure and create a new project</p>
				</div>
				<button class="dlg-np__close" onclick={() => { showNewProject = false; showAdvancedOptions = false; }} aria-label="Close dialog">
					<X size={16} />
				</button>
			</div>

			<!-- Body -->
			<form onsubmit={handleCreateProject} class="dlg-np__body">
				<!-- Name -->
				<div class="dlg-np__field">
					<label for="np-name" class="dlg-np__label">Project Name</label>
					<input
						id="np-name"
						type="text"
						bind:value={newProject.name}
						class="dlg-np__input"
						placeholder="e.g. Website Redesign"
						required
					/>
				</div>

				<!-- Type -->
				<div class="dlg-np__field">
					<span class="dlg-np__label">Type</span>
					<div class="dlg-np__type-grid">
						{#each [
							{ value: 'internal', label: 'Internal', color: '#8b5cf6' },
							{ value: 'client_work', label: 'Client Work', color: '#3b82f6' },
							{ value: 'learning', label: 'Learning', color: '#14b8a6' }
						] as t}
							<button
								type="button"
								onclick={() => newProject.project_type = t.value}
								class="dlg-np__card {newProject.project_type === t.value ? 'dlg-np__card--active' : ''}"
								style={newProject.project_type === t.value ? `border-color: ${t.color}; background: color-mix(in srgb, ${t.color} 8%, var(--dbg))` : ''}
								aria-label="Select {t.label} type"
							>
								<span class="dlg-np__card-icon" style={newProject.project_type === t.value ? `color: ${t.color}` : ''}>
									{@render typeIcon(t.value, 20)}
								</span>
								<span class="dlg-np__card-label">{t.label}</span>
							</button>
						{/each}
					</div>
				</div>

				<!-- Priority -->
				<div class="dlg-np__field">
					<span class="dlg-np__label">Priority</span>
					<div class="dlg-np__prio-grid">
						{#each [
							{ value: 'low', label: 'Low', color: '#22c55e' },
							{ value: 'medium', label: 'Medium', color: '#eab308' },
							{ value: 'high', label: 'High', color: '#f97316' },
							{ value: 'critical', label: 'Critical', color: '#ef4444' }
						] as p}
							<button
								type="button"
								onclick={() => newProject.priority = p.value as 'low' | 'medium' | 'high' | 'critical'}
								class="dlg-np__prio {newProject.priority === p.value ? 'dlg-np__prio--active' : ''}"
								style={newProject.priority === p.value ? `border-color: ${p.color}; background: color-mix(in srgb, ${p.color} 10%, var(--dbg))` : ''}
								aria-label="Set priority to {p.label}"
							>
								<span class="dlg-np__prio-dot" style="background: {p.color}"></span>
								<span>{p.label}</span>
							</button>
						{/each}
					</div>
				</div>

				<!-- Client -->
				<div class="dlg-np__field">
					<label for="np-client" class="dlg-np__label">Client <span class="dlg-np__optional">(optional)</span></label>
					<select
						id="np-client"
						class="dlg-np__select"
						bind:value={newProject.client_name}
					>
						<option value="">No client</option>
						{#each clients as client}
							<option value={client.name}>{client.name}</option>
						{/each}
					</select>
					{#if clients.length === 0}
						<p class="dlg-np__hint">No clients yet. Add one from the Clients page first.</p>
					{/if}
				</div>

				<!-- Description -->
				<div class="dlg-np__field">
					<label for="np-desc" class="dlg-np__label">Description <span class="dlg-np__optional">(optional)</span></label>
					<textarea
						id="np-desc"
						bind:value={newProject.description}
						class="dlg-np__textarea"
						rows="3"
						placeholder="Brief overview of the project scope and goals"
					></textarea>
				</div>

				<!-- Advanced Options -->
				<button
					type="button"
					onclick={() => showAdvancedOptions = !showAdvancedOptions}
					class="dlg-np__adv-toggle"
				>
					<ChevronRight size={14} class={showAdvancedOptions ? 'dlg-np__chevron--open' : ''} />
					Advanced Options
				</button>

				{#if showAdvancedOptions}
					<div class="dlg-np__adv-content">
						<p class="dlg-np__hint">Team assignment and tags are available after project creation.</p>
					</div>
				{/if}

				{#if createError}
					<div class="dlg-np__error">{createError}</div>
				{/if}
			</form>

			<!-- Footer -->
			<div class="dlg-np__footer">
				<button type="button" onclick={() => { showNewProject = false; showAdvancedOptions = false; }} class="btn-pill btn-pill-ghost btn-pill-sm">
					Cancel
				</button>
				<button type="button" onclick={() => { const form = document.querySelector('.dlg-np__body') as HTMLFormElement; form?.requestSubmit(); }} class="btn-pill btn-pill-primary btn-pill-sm" disabled={!newProject.name.trim()}>
					Create Project
				</button>
			</div>
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>

<style>
	/* Page & Layout */
	.prm-ls-page { background: var(--dbg2, rgba(249,250,251,.5)); }
	.prm-ls-bar { background: var(--dbg, #fff); border-bottom: 1px solid var(--dbd, #e5e7eb); }
	.prm-ls-title { color: var(--dt, #111); }
	.prm-ls-muted { color: var(--dt3, #6b7280); }
	.prm-ls-icon { color: var(--dt4, #9ca3af); }
	.prm-ls-label { color: var(--dt2, #4b5563); }
	.prm-ls-spinner { border-color: var(--dt, #111); border-top-color: transparent; }

	/* Search */
	.prm-ls-search { border: 1px solid var(--dbd, #e5e7eb); background: var(--dbg, #fff); color: var(--dt, #111); }
	.prm-ls-search:focus { box-shadow: 0 0 0 2px var(--dt, #111); }
	.prm-ls-divider-l { border-left: 1px solid var(--dbd, #e5e7eb); }

	/* Dropdowns */
	.prm-ls-dropdown { background: var(--dbg, #fff); border: 1px solid var(--dbd, #e5e7eb); }
	.prm-ls-dropdown-item:hover { background: var(--dbg3, #f3f4f6); }

	/* Controls */
	.prm-ls-checkbox { border-color: var(--dbd, #d1d5db); color: var(--dt, #111); }
	.prm-ls-toggle-border { border: 1px solid var(--dbd, #e5e7eb); }
	.prm-ls-empty-bg { background: var(--dbg3, #f3f4f6); }

	/* Stat Cards */
	.prm-ls-stat { display: flex; align-items: center; gap: 0.75rem; padding: 0.625rem 0.75rem; background: var(--dbg2, #f9fafb); border-radius: 0.5rem; border: 1px solid var(--dbd2, #f3f4f6); }
	.prm-ls-stat__icon { width: 2rem; height: 2rem; border-radius: 0.375rem; display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
	.prm-ls-stat__icon--total { background: color-mix(in srgb, #8b5cf6 12%, var(--dbg, #fff)); color: #8b5cf6; }
	.prm-ls-stat__icon--active { background: color-mix(in srgb, #22c55e 12%, var(--dbg, #fff)); color: #22c55e; }
	.prm-ls-stat__icon--paused { background: color-mix(in srgb, #f59e0b 12%, var(--dbg, #fff)); color: #f59e0b; }
	.prm-ls-stat__icon--completed { background: color-mix(in srgb, #3b82f6 12%, var(--dbg, #fff)); color: #3b82f6; }
	.prm-ls-stat__body { display: flex; flex-direction: column; }
	.prm-ls-stat__value { font-size: 1.125rem; font-weight: 700; line-height: 1.2; color: var(--dt, #111); }
	.prm-ls-stat__label { font-size: 0.6875rem; color: var(--dt3, #6b7280); }

	/* Kanban & Cards */
	.prm-ls-column { background: var(--dbg, #fff); border-radius: 0.75rem; border: 1px solid var(--dbd, #e5e7eb); }
	.prm-ls-column__header { border-bottom: 1px solid var(--dbd2, #f3f4f6); }
	.prm-ls-kanban-card { background: var(--dbg2, #f9fafb); }
	.prm-ls-kanban-card:hover { background: var(--dbg3, #f3f4f6); }
	.prm-ls-kanban-count { display: inline-flex; align-items: center; justify-content: center; min-width: 1.25rem; height: 1.25rem; padding: 0 0.375rem; font-size: 0.6875rem; font-weight: 600; border-radius: 9999px; background: var(--dbg3, #f3f4f6); color: var(--dt3, #6b7280); }
	.prm-ls-card { background: var(--dbg, #fff); border-radius: 0.75rem; border: 1px solid var(--dbd, #e5e7eb); }
	.prm-ls-card:hover { box-shadow: 0 4px 6px rgba(0,0,0,.1); transform: scale(1.02); }
	.prm-ls-card--internal { border-top: 3px solid #8b5cf6; }
	.prm-ls-card--client_work { border-top: 3px solid #3b82f6; }
	.prm-ls-card--learning { border-top: 3px solid #14b8a6; }
	.prm-ls-card__footer { border-top: 1px solid var(--dbd2, #f3f4f6); }

	/* List view */
	.prm-ls-table-head { background: var(--dbg2, #f9fafb); border-bottom: 1px solid var(--dbd, #e5e7eb); }
	.prm-ls-table-body { }
	.prm-ls-table-body > :global(tr + tr) { border-top: 1px solid var(--dbd2, #f3f4f6); }
	.prm-ls-table-row:hover { background: var(--dbg2, #f9fafb); }

	/* Status/Priority/Type defaults */
	.prm-ls-status-default { background: var(--dbg3, #f3f4f6); color: var(--dt2, #4b5563); }
	.prm-ls-priority-default { color: var(--dt4, #9ca3af); }
	.prm-ls-type-default { color: var(--dt2, #4b5563); background: var(--dbg2, #f9fafb); }

	/* ── New Project Dialog (Foundation dlg- pattern) ── */
	/* Portal renders outside component DOM — all dialog styles must be :global */
	:global(.dlg-np__overlay) {
		position: fixed; inset: 0; z-index: 300;
		background: rgba(0, 0, 0, 0.72);
		backdrop-filter: blur(6px);
		-webkit-backdrop-filter: blur(6px);
	}
	:global(.dlg-np__panel) {
		position: fixed; z-index: 301;
		top: 50%; left: 50%; transform: translate(-50%, -50%);
		width: 100%; max-width: 480px; max-height: 90vh;
		display: flex; flex-direction: column;
		background: var(--dbg); border: 1px solid var(--dbd);
		border-radius: 14px;
		box-shadow: 0 24px 64px rgba(0, 0, 0, 0.55), 0 8px 24px rgba(0, 0, 0, 0.35);
	}
	:global(.dlg-np__header) {
		display: flex; justify-content: space-between; align-items: flex-start;
		padding: 20px 24px 16px;
		border-bottom: 1px solid var(--dbd2);
	}
	:global(.dlg-np__title) { font-size: 15px; font-weight: 600; color: var(--dt); margin: 0; }
	:global(.dlg-np__subtitle) { font-size: 12px; color: var(--dt3); margin-top: 2px; }
	:global(.dlg-np__close) {
		width: 28px; height: 28px; display: flex; align-items: center; justify-content: center;
		background: var(--dbg2); border: 1px solid var(--dbd); border-radius: 7px;
		color: var(--dt3); cursor: pointer; transition: background 0.15s;
	}
	:global(.dlg-np__close:hover) { background: var(--dbg3); color: var(--dt); }
	:global(.dlg-np__body) {
		flex: 1; overflow-y: auto;
		padding: 20px 24px;
		display: flex; flex-direction: column; gap: 16px;
	}
	:global(.dlg-np__footer) {
		display: flex; justify-content: flex-end; gap: 8px;
		padding: 14px 24px;
		border-top: 1px solid var(--dbd2);
	}
	:global(.dlg-np__field) { display: flex; flex-direction: column; gap: 6px; }
	:global(.dlg-np__label) { font-size: 12px; font-weight: 600; color: var(--dt2); text-transform: uppercase; letter-spacing: 0.04em; }
	:global(.dlg-np__optional) { font-weight: 400; text-transform: none; color: var(--dt4); letter-spacing: 0; }
	:global(.dlg-np__input) {
		width: 100%; padding: 9px 12px;
		background: var(--dbg2); border: 1px solid var(--dbd); border-radius: 8px;
		color: var(--dt); font-size: 13px; outline: none; transition: border-color 0.15s;
		box-sizing: border-box;
	}
	:global(.dlg-np__input:focus) { border-color: var(--dt3); }
	:global(.dlg-np__input::placeholder) { color: var(--dt4); }
	:global(.dlg-np__select) {
		width: 100%; padding: 9px 12px;
		background: var(--dbg2); border: 1px solid var(--dbd); border-radius: 8px;
		color: var(--dt); font-size: 13px; outline: none; transition: border-color 0.15s;
		appearance: none; box-sizing: border-box;
		background-image: url("data:image/svg+xml,%3Csvg width='12' height='12' fill='none' stroke='%23888' viewBox='0 0 24 24' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath stroke-linecap='round' stroke-linejoin='round' stroke-width='2.5' d='M19 9l-7 7-7-7'/%3E%3C/svg%3E");
		background-repeat: no-repeat;
		background-position: right 12px center;
		padding-right: 32px;
	}
	:global(.dlg-np__select:focus) { border-color: var(--dt3); }
	:global(.dlg-np__textarea) {
		width: 100%; padding: 9px 12px;
		background: var(--dbg2); border: 1px solid var(--dbd); border-radius: 8px;
		color: var(--dt); font-size: 13px; outline: none; transition: border-color 0.15s;
		resize: none; font-family: inherit; box-sizing: border-box;
	}
	:global(.dlg-np__textarea:focus) { border-color: var(--dt3); }
	:global(.dlg-np__textarea::placeholder) { color: var(--dt4); }
	:global(.dlg-np__hint) { font-size: 11px; color: var(--dt4); margin-top: 2px; }

	/* Type selection cards */
	:global(.dlg-np__type-grid) { display: grid; grid-template-columns: repeat(3, 1fr); gap: 8px; }
	:global(.dlg-np__card) {
		display: flex; flex-direction: column; align-items: center; gap: 6px;
		padding: 14px 8px 12px; border: 1.5px solid var(--dbd); border-radius: 10px;
		background: var(--dbg2); cursor: pointer; transition: all 0.15s;
	}
	:global(.dlg-np__card:hover) { background: var(--dbg3); }
	:global(.dlg-np__card--active) { font-weight: 600; }
	:global(.dlg-np__card-icon) { color: var(--dt3); display: flex; align-items: center; justify-content: center; transition: color 0.15s; }
	:global(.dlg-np__card-label) { font-size: 12px; color: var(--dt2); }

	/* Priority buttons */
	:global(.dlg-np__prio-grid) { display: flex; gap: 8px; }
	:global(.dlg-np__prio) {
		flex: 1; display: flex; align-items: center; justify-content: center; gap: 6px;
		padding: 8px 10px; border: 1.5px solid var(--dbd); border-radius: 8px;
		background: var(--dbg2); cursor: pointer; font-size: 12px; font-weight: 500;
		color: var(--dt2); transition: all 0.15s;
	}
	:global(.dlg-np__prio:hover) { background: var(--dbg3); }
	:global(.dlg-np__prio--active) { font-weight: 600; }
	:global(.dlg-np__prio-dot) { width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; }

	/* Advanced toggle */
	:global(.dlg-np__adv-toggle) {
		display: flex; align-items: center; gap: 6px;
		font-size: 12px; color: var(--dt3); cursor: pointer;
		background: none; border: none; padding: 0; transition: color 0.15s;
	}
	:global(.dlg-np__adv-toggle:hover) { color: var(--dt2); }
	:global(.dlg-np__chevron--open) { transform: rotate(90deg); }
	:global(.dlg-np__adv-content) {
		padding: 10px 12px; background: var(--dbg2); border-radius: 8px;
		border: 1px solid var(--dbd);
	}
	:global(.dlg-np__error) {
		font-size: 13px; color: #ef4444;
		background: color-mix(in srgb, #ef4444 8%, var(--dbg));
		padding: 8px 12px; border-radius: 8px;
		border: 1px solid color-mix(in srgb, #ef4444 20%, var(--dbd));
	}

	/* Type icon wrappers for list/card views */
	.prm-ls-type-icon { display: inline-flex; align-items: center; color: var(--dt3); }
	.prm-ls-type-inline { display: inline-flex; align-items: center; gap: 4px; }

	/* Status Pills (global — used by utility import) */
	:global(.prm-status) { display: inline-block; padding: 0.125rem 0.5rem; font-size: 0.6875rem; font-weight: 600; border-radius: 9999px; border: 1px solid transparent; }
	:global(.prm-status--active) { background: color-mix(in srgb, #22c55e 15%, var(--dbg)); color: #22c55e; border-color: color-mix(in srgb, #22c55e 25%, var(--dbd)); }
	:global(.prm-status--paused) { background: color-mix(in srgb, #f59e0b 15%, var(--dbg)); color: #f59e0b; border-color: color-mix(in srgb, #f59e0b 25%, var(--dbd)); }
	:global(.prm-status--completed) { background: color-mix(in srgb, #3b82f6 15%, var(--dbg)); color: #3b82f6; border-color: color-mix(in srgb, #3b82f6 25%, var(--dbd)); }
	:global(.prm-status--archived) { background: var(--dbg3); color: var(--dt3); border-color: var(--dbd); }

	/* Priority Pills (global) */
	:global(.prm-priority) { display: inline-block; padding: 0.125rem 0.5rem; font-size: 0.6875rem; font-weight: 600; border-radius: 9999px; }
	:global(.prm-priority--critical) { color: #ef4444; background: color-mix(in srgb, #ef4444 10%, var(--dbg)); }
	:global(.prm-priority--high) { color: #f97316; background: color-mix(in srgb, #f97316 10%, var(--dbg)); }
	:global(.prm-priority--medium) { color: #eab308; background: color-mix(in srgb, #eab308 10%, var(--dbg)); }
	:global(.prm-priority--low) { color: #22c55e; background: color-mix(in srgb, #22c55e 10%, var(--dbg)); }
	:global(.prm-priority--default) { color: var(--dt3); background: var(--dbg2); }

	/* Type Pills (global) */
	:global(.prm-type) { display: inline-block; padding: 0.125rem 0.5rem; font-size: 0.6875rem; font-weight: 600; border-radius: 6px; }
	:global(.prm-type--internal) { color: #8b5cf6; background: color-mix(in srgb, #8b5cf6 12%, var(--dbg)); }
	:global(.prm-type--client) { color: #3b82f6; background: color-mix(in srgb, #3b82f6 12%, var(--dbg)); }
	:global(.prm-type--learning) { color: #14b8a6; background: color-mix(in srgb, #14b8a6 12%, var(--dbg)); }

	/* Status Dots */
	.prm-dot { width: 0.625rem; height: 0.625rem; border-radius: 50%; flex-shrink: 0; }
	.prm-dot--active { background: #22c55e; }
	.prm-dot--paused { background: #f59e0b; }
	.prm-dot--completed { background: #3b82f6; }

	/* Filter active state */
	.prm-filter--active { box-shadow: 0 0 0 2px var(--dt3); }
	.prm-dropdown-item--active { background: var(--dbg3); font-weight: 600; }
</style>
