<script lang="ts">
	import type { ProcessMiningKPIData } from '$lib/stores/dashboard/types';

	let { data, loading = false }: { data: ProcessMiningKPIData | null; loading?: boolean } =
		$props();

	const fitness = $derived(data?.conformanceFitness ?? 0);
	const fitnessPercent = $derived(Math.round(fitness * 100));
	const color = $derived(
		fitness >= 0.8 ? '#22c55e' : fitness >= 0.6 ? '#f59e0b' : '#ef4444'
	);

	const RADIUS = 40;
	const CIRCUMFERENCE = $derived(2 * Math.PI * RADIUS);
	const dashOffset = $derived(CIRCUMFERENCE - (fitnessPercent / 100) * CIRCUMFERENCE);
</script>

{#if loading}
	<div class="animate-pulse h-32 bg-gray-200 rounded"></div>
{:else}
	<div class="flex flex-col items-center gap-2 p-4">
		<svg width="100" height="100" viewBox="0 0 100 100" aria-label="Conformance fitness ring">
			<circle cx="50" cy="50" r={RADIUS} fill="none" stroke="#e5e7eb" stroke-width="8" />
			<circle
				cx="50"
				cy="50"
				r={RADIUS}
				fill="none"
				stroke={color}
				stroke-width="8"
				stroke-dasharray={CIRCUMFERENCE}
				stroke-dashoffset={dashOffset}
				stroke-linecap="round"
				transform="rotate(-90 50 50)"
				style="transition: stroke-dashoffset 0.5s ease"
			/>
			<text x="50" y="55" text-anchor="middle" font-size="18" font-weight="bold" fill={color}>
				{fitnessPercent}%
			</text>
		</svg>
		<div class="text-sm text-gray-500">Conformance Fitness</div>
		<div class="text-xs text-gray-400">
			Precision: {Math.round((data?.conformancePrecision ?? 0) * 100)}%
		</div>
		{#if data}
			<div
				class="text-xs font-medium px-2 py-0.5 rounded-full {data.isConformant
					? 'text-green-700 bg-green-50'
					: 'text-red-700 bg-red-50'}"
			>
				{data.isConformant ? 'Conformant' : 'Non-Conformant'}
			</div>
		{:else}
			<div class="text-xs text-gray-400">No data — pm4py may be offline</div>
		{/if}
	</div>
{/if}
