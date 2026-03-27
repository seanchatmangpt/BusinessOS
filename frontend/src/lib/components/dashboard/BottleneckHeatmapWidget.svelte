<script lang="ts">
	import type { ProcessMiningKPIData } from '$lib/stores/dashboard/types';

	let { data, loading = false }: { data: ProcessMiningKPIData | null; loading?: boolean } =
		$props();

	const activities = $derived(Object.entries(data?.activityFrequencies ?? {}));
	const maxFreq = $derived(Math.max(...activities.map(([, f]) => f), 1));

	function isBottleneck(act: string): boolean {
		return data?.bottleneckActivities?.some((b) => b.activity === act) ?? false;
	}

	function heatColor(freq: number): string {
		const ratio = freq / maxFreq;
		const hue = 240 - Math.round(ratio * 240); // blue (cold) to red (hot)
		return `hsl(${hue}, 70%, 50%)`;
	}
</script>

{#if loading}
	<div class="animate-pulse h-32 bg-gray-200 rounded"></div>
{:else}
	<div class="p-4">
		<div class="text-sm font-medium text-gray-700 mb-2">Activity Heatmap</div>
		{#if activities.length === 0}
			<p class="text-xs text-gray-400 text-center py-4">No activity data — pm4py may be offline</p>
		{:else}
			<div class="grid grid-cols-4 gap-1">
				{#each activities as [act, freq]}
					<div
						class="rounded p-1 text-xs text-white text-center truncate"
						style="background-color: {heatColor(freq)}; {isBottleneck(act)
							? 'outline: 2px solid #f59e0b;'
							: ''}"
						title="{act}: {freq}"
					>
						{act.length > 8 ? act.slice(0, 8) + '…' : act}
					</div>
				{/each}
			</div>
			{#if data && data.bottleneckActivities.length > 0}
				<div class="mt-2 text-xs text-amber-600 flex items-center gap-1">
					<span class="inline-block w-3 h-3 rounded border-2 border-amber-500"></span>
					{data.bottleneckActivities.length} bottleneck{data.bottleneckActivities.length > 1
						? 's'
						: ''} detected
				</div>
			{/if}
		{/if}
	</div>
{/if}
