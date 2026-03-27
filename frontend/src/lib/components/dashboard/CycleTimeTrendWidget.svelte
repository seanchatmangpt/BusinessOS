<script lang="ts">
	import type { ProcessMiningKPIData } from '$lib/stores/dashboard/types';

	let { data, loading = false }: { data: ProcessMiningKPIData | null; loading?: boolean } =
		$props();
</script>

{#if loading}
	<div class="animate-pulse h-24 bg-gray-200 rounded"></div>
{:else}
	<div class="p-4">
		<div class="text-sm font-medium text-gray-700 mb-2">Process Stats</div>
		{#if !data}
			<p class="text-xs text-gray-400 text-center py-4">No data — pm4py may be offline</p>
		{:else}
			<div class="grid grid-cols-3 gap-2 text-center">
				<div class="bg-gray-50 rounded p-2">
					<div class="text-lg font-bold text-blue-600">{data.traceCount}</div>
					<div class="text-xs text-gray-500">Cases</div>
				</div>
				<div class="bg-gray-50 rounded p-2">
					<div class="text-lg font-bold text-green-600">{data.eventCount}</div>
					<div class="text-xs text-gray-500">Events</div>
				</div>
				<div class="bg-gray-50 rounded p-2">
					<div class="text-lg font-bold text-purple-600">{data.variantCount}</div>
					<div class="text-xs text-gray-500">Variants</div>
				</div>
			</div>
			<div class="mt-3 text-xs text-gray-400 text-center">
				Last fetched: {data.fetchedAt ? new Date(data.fetchedAt).toLocaleTimeString() : '—'}
			</div>
		{/if}
	</div>
{/if}
