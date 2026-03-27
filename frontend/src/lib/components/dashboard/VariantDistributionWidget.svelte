<script lang="ts">
	import type { ProcessMiningKPIData } from '$lib/stores/dashboard/types';

	let { data, loading = false }: { data: ProcessMiningKPIData | null; loading?: boolean } =
		$props();

	const variants = $derived((data?.topVariants ?? []).slice(0, 5));
	const maxCount = $derived(Math.max(...variants.map((v) => v.count), 1));
</script>

{#if loading}
	<div class="animate-pulse space-y-2 p-4">
		{#each [1, 2, 3] as _}
			<div class="h-6 bg-gray-200 rounded"></div>
		{/each}
	</div>
{:else}
	<div class="p-4 space-y-2">
		<div class="text-sm font-medium text-gray-700 mb-2">
			Top Variants ({data?.variantCount ?? 0} total)
		</div>
		{#if variants.length === 0}
			<p class="text-xs text-gray-400 text-center py-4">No variant data — pm4py may be offline</p>
		{:else}
			{#each variants as v, i}
				<div class="flex items-center gap-2 text-xs">
					<span class="w-20 truncate text-gray-500">Variant {i + 1}</span>
					<div class="flex-1 bg-gray-100 rounded h-5 relative">
						<div
							class="h-5 rounded bg-blue-500 transition-all"
							style="width: {(v.count / maxCount) * 100}%"
						></div>
					</div>
					<span class="w-8 text-right text-gray-600">{v.count}</span>
				</div>
			{/each}
		{/if}
	</div>
{/if}
