<script lang="ts">
	import { Search } from 'lucide-svelte';
	import type { ModuleCategory, ModuleFilters } from '$lib/types/modules';
	import { categoryLabels } from '$lib/types/modules';

	interface Props {
		filters: ModuleFilters;
		onFiltersChange: (filters: Partial<ModuleFilters>) => void;
	}

	let { filters, onFiltersChange }: Props = $props();

	let searchInput = $state(filters.search);

	// Debounce search
	let debounceTimer: number;
	function handleSearchInput(value: string) {
		searchInput = value;
		clearTimeout(debounceTimer);
		debounceTimer = setTimeout(() => {
			onFiltersChange({ search: value });
		}, 300) as unknown as number;
	}

	const categories: (ModuleCategory | null)[] = [
		null,
		'productivity',
		'communication',
		'finance',
		'analytics',
		'automation',
		'integration',
		'utilities',
		'custom'
	];

	const sortOptions: Array<{ value: 'popular' | 'newest' | 'name' | 'installs'; label: string }> = [
		{ value: 'popular', label: 'Popular' },
		{ value: 'newest', label: 'Newest' },
		{ value: 'name', label: 'Name' },
		{ value: 'installs', label: 'Most Installed' }
	];
</script>

<div class="flex flex-col sm:flex-row gap-3">
	<!-- Search Input -->
	<div class="flex-1 relative">
		<Search class="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
		<input
			type="text"
			placeholder="Search modules..."
			value={searchInput}
			oninput={(e) => handleSearchInput(e.currentTarget.value)}
			class="w-full pl-10 pr-4 py-2.5 border border-gray-200 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-all"
		/>
	</div>

	<!-- Category Filter -->
	<select
		value={filters.category || ''}
		onchange={(e) => onFiltersChange({ category: e.currentTarget.value || null })}
		class="px-4 py-2.5 border border-gray-200 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-all bg-white"
	>
		<option value="">All Categories</option>
		{#each categories.filter(c => c !== null) as category}
			<option value={category}>{categoryLabels[category]}</option>
		{/each}
	</select>

	<!-- Sort Dropdown -->
	<select
		value={filters.sort}
		onchange={(e) => onFiltersChange({ sort: e.currentTarget.value })}
		class="px-4 py-2.5 border border-gray-200 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-all bg-white"
	>
		{#each sortOptions as option}
			<option value={option.value}>{option.label}</option>
		{/each}
	</select>
</div>
