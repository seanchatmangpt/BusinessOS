<script lang="ts">
	import { DropdownMenu } from 'bits-ui';
	import type { ViewMode } from '$lib/stores/clients';
	import type { ClientStatus, ClientType } from '$lib/api';

	interface Props {
		view: ViewMode;
		searchQuery: string;
		statusFilter: ClientStatus | null;
		typeFilter: ClientType | null;
		onViewChange: (view: ViewMode) => void;
		onSearchChange: (query: string) => void;
		onStatusChange: (status: ClientStatus | null) => void;
		onTypeChange: (type: ClientType | null) => void;
	}

	let {
		view = $bindable(),
		searchQuery = $bindable(),
		statusFilter,
		typeFilter,
		onViewChange,
		onSearchChange,
		onStatusChange,
		onTypeChange
	}: Props = $props();

	const views: { id: ViewMode; label: string; icon: string }[] = [
		{ id: 'table', label: 'Table', icon: 'M3 10h18M3 14h18M3 18h18M3 6h18' },
		{
			id: 'cards',
			label: 'Cards',
			icon: 'M4 5a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1H5a1 1 0 01-1-1V5zm10 0a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 01-1-1V5zM4 15a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1H5a1 1 0 01-1-1v-4zm10 0a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 01-1-1v-4z'
		},
		{
			id: 'kanban',
			label: 'Pipeline',
			icon: 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2'
		}
	];

	const statuses: { id: ClientStatus | ''; label: string }[] = [
		{ id: '', label: 'All Statuses' },
		{ id: 'lead', label: 'Leads' },
		{ id: 'prospect', label: 'Prospects' },
		{ id: 'active', label: 'Active' },
		{ id: 'inactive', label: 'Inactive' },
		{ id: 'churned', label: 'Churned' }
	];

	const types: { id: ClientType | ''; label: string }[] = [
		{ id: '', label: 'All Types' },
		{ id: 'company', label: 'Companies' },
		{ id: 'individual', label: 'Individuals' }
	];

	// Get display label for current filter
	const getStatusLabel = $derived(
		statuses.find((s) => s.id === (statusFilter || ''))?.label || 'All Statuses'
	);

	const getTypeLabel = $derived(
		types.find((t) => t.id === (typeFilter || ''))?.label || 'All Types'
	);
</script>

<div class="px-6 py-3 border-b border-gray-200 bg-gray-50/50">
	<div class="flex items-center justify-between gap-4">
		<!-- Left: Search & Filters -->
		<div class="flex items-center gap-3 flex-1">
			<!-- Search -->
			<div class="relative flex-1 max-w-xs">
				<svg
					class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400"
					fill="none"
					stroke="currentColor"
					viewBox="0 0 24 24"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
					/>
				</svg>
				<input
					type="text"
					placeholder="Search clients..."
					bind:value={searchQuery}
					oninput={() => onSearchChange(searchQuery)}
					class="w-full pl-9 pr-4 py-2 text-sm border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-gray-900 focus:border-transparent"
				/>
			</div>

			<!-- Status Filter -->
			<DropdownMenu.Root>
				<DropdownMenu.Trigger
					class="flex items-center gap-2 px-3 py-2 text-sm border border-gray-200 rounded-lg bg-white hover:bg-gray-50 transition-colors"
				>
					<span>{getStatusLabel}</span>
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M19 9l-7 7-7-7"
						/>
					</svg>
				</DropdownMenu.Trigger>
				<DropdownMenu.Portal>
					<DropdownMenu.Content
						class="z-50 min-w-[160px] bg-white border border-gray-200 rounded-lg shadow-lg p-1"
						sideOffset={4}
					>
						{#each statuses as status}
							<DropdownMenu.Item
								class="px-3 py-2 text-sm rounded hover:bg-gray-100 cursor-pointer transition-colors {status.id ===
								(statusFilter || '')
									? 'bg-gray-50 font-medium'
									: ''}"
								onclick={() => onStatusChange(status.id || null)}
							>
								{status.label}
							</DropdownMenu.Item>
						{/each}
					</DropdownMenu.Content>
				</DropdownMenu.Portal>
			</DropdownMenu.Root>

			<!-- Type Filter -->
			<DropdownMenu.Root>
				<DropdownMenu.Trigger
					class="flex items-center gap-2 px-3 py-2 text-sm border border-gray-200 rounded-lg bg-white hover:bg-gray-50 transition-colors"
				>
					<span>{getTypeLabel}</span>
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M19 9l-7 7-7-7"
						/>
					</svg>
				</DropdownMenu.Trigger>
				<DropdownMenu.Portal>
					<DropdownMenu.Content
						class="z-50 min-w-[160px] bg-white border border-gray-200 rounded-lg shadow-lg p-1"
						sideOffset={4}
					>
						{#each types as type}
							<DropdownMenu.Item
								class="px-3 py-2 text-sm rounded hover:bg-gray-100 cursor-pointer transition-colors {type.id ===
								(typeFilter || '')
									? 'bg-gray-50 font-medium'
									: ''}"
								onclick={() => onTypeChange(type.id || null)}
							>
								{type.label}
							</DropdownMenu.Item>
						{/each}
					</DropdownMenu.Content>
				</DropdownMenu.Portal>
			</DropdownMenu.Root>
		</div>

		<!-- Right: View Switcher -->
		<div class="flex items-center gap-1 p-1 bg-gray-100 rounded-lg">
			{#each views as v}
				<button
					onclick={() => {
						view = v.id;
						onViewChange(v.id);
					}}
					class="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-md transition-colors {view ===
					v.id
						? 'bg-white text-gray-900 shadow-sm'
						: 'text-gray-600 hover:text-gray-900'}"
				>
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={v.icon} />
					</svg>
					{v.label}
				</button>
			{/each}
		</div>
	</div>
</div>
