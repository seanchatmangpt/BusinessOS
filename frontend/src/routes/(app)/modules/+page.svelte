<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { Plus, Loader2 } from 'lucide-svelte';
	import { customModulesStore } from '$lib/stores/customModulesStore';
	import ModuleCard from '$lib/components/modules/ModuleCard.svelte';
	import ModuleFilters from '$lib/components/modules/ModuleFilters.svelte';

	let store = $state(customModulesStore);
	let storeState = $state($store);

	$effect(() => {
		storeState = $store;
	});

	onMount(() => {
		store.loadModules();
	});

	function handleFiltersChange(filters: Parameters<typeof store.setFilters>[0]) {
		store.setFilters(filters);
		store.loadModules();
	}

	function handleModuleClick(moduleId: string) {
		goto(`/modules/${moduleId}`);
	}

	function handleCreateModule() {
		goto('/modules/create');
	}
</script>

<div class="h-full flex flex-col bg-white">
	<!-- Header -->
	<div class="flex-shrink-0 border-b border-gray-200 bg-white px-8 py-6">
		<div class="flex items-center justify-between mb-6">
			<div>
				<h1 class="text-2xl font-bold text-gray-900">Custom Modules</h1>
				<p class="text-sm text-gray-600 mt-1">Browse and manage custom modules for your workspace</p>
			</div>
			<button
				onclick={handleCreateModule}
				class="flex items-center gap-2 px-4 py-2.5 bg-blue-600 text-white rounded-xl hover:bg-blue-700 transition-colors font-medium"
			>
				<Plus class="w-5 h-5" />
				<span>Create Module</span>
			</button>
		</div>

		<!-- Filters -->
		<ModuleFilters
			filters={storeState.filters}
			onFiltersChange={handleFiltersChange}
		/>
	</div>

	<!-- Content -->
	<div class="flex-1 overflow-y-auto px-8 py-6">
		{#if storeState.loading}
			<!-- Loading State -->
			<div class="flex items-center justify-center h-64">
				<div class="text-center">
					<Loader2 class="w-8 h-8 text-blue-600 animate-spin mx-auto mb-3" />
					<p class="text-sm text-gray-600">Loading modules...</p>
				</div>
			</div>
		{:else if storeState.error}
			<!-- Error State -->
			<div class="flex items-center justify-center h-64">
				<div class="text-center max-w-md">
					<div class="w-16 h-16 bg-red-100 rounded-full flex items-center justify-center mx-auto mb-4">
						<svg class="w-8 h-8 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
					</div>
					<p class="text-sm font-medium text-gray-900 mb-2">Failed to load modules</p>
					<p class="text-sm text-gray-600 mb-4">{storeState.error}</p>
					<button
						onclick={() => store.loadModules()}
						class="px-4 py-2 bg-gray-900 text-white rounded-lg hover:bg-gray-800 transition-colors text-sm"
					>
						Try Again
					</button>
				</div>
			</div>
		{:else if storeState.modules.length === 0}
			<!-- Empty State -->
			<div class="flex items-center justify-center h-64">
				<div class="text-center max-w-md">
					<div class="w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-4">
						<svg class="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
						</svg>
					</div>
					<p class="text-sm font-medium text-gray-900 mb-2">No modules found</p>
					<p class="text-sm text-gray-600 mb-4">
						{#if storeState.filters.search || storeState.filters.category}
							Try adjusting your filters or create a new module.
						{:else}
							Get started by creating your first custom module.
						{/if}
					</p>
					<button
						onclick={handleCreateModule}
						class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm"
					>
						Create Module
					</button>
				</div>
			</div>
		{:else}
			<!-- Modules Grid -->
			<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-5">
				{#each storeState.modules as module}
					<ModuleCard
						{module}
						onClick={() => handleModuleClick(module.id)}
					/>
				{/each}
			</div>

			<!-- Results Summary -->
			<div class="mt-8 text-center">
				<p class="text-sm text-gray-600">
					Showing {storeState.modules.length} of {storeState.total} modules
				</p>
			</div>
		{/if}
	</div>
</div>
