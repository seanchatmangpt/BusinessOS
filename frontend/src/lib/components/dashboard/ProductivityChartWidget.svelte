<script lang="ts">
	import { fade, fly } from 'svelte/transition';

	interface DataPoint {
		label: string;
		value: number;
	}

	interface Props {
		title?: string;
		data?: DataPoint[];
		type?: 'bar' | 'line';
		color?: 'blue' | 'green' | 'purple' | 'orange';
	}

	let { title = 'Productivity', data = [], type = 'bar', color = 'blue' }: Props = $props();

	// Default mock data if none provided
	const defaultData: DataPoint[] = [
		{ label: 'Mon', value: 12 },
		{ label: 'Tue', value: 15 },
		{ label: 'Wed', value: 18 },
		{ label: 'Thu', value: 14 },
		{ label: 'Fri', value: 10 },
		{ label: 'Sat', value: 5 },
		{ label: 'Sun', value: 3 }
	];

	const chartData = $derived(data.length > 0 ? data : defaultData);
	const maxValue = $derived(Math.max(...chartData.map((d) => d.value)));

	// Color mappings
	const colorClasses = {
		blue: {
			bar: 'bg-blue-500 dark:bg-blue-400',
			gradient: 'from-blue-500 to-blue-600',
			line: 'text-blue-500 dark:text-blue-400'
		},
		green: {
			bar: 'bg-green-500 dark:bg-green-400',
			gradient: 'from-green-500 to-green-600',
			line: 'text-green-500 dark:text-green-400'
		},
		purple: {
			bar: 'bg-purple-500 dark:bg-purple-400',
			gradient: 'from-purple-500 to-purple-600',
			line: 'text-purple-500 dark:text-purple-400'
		},
		orange: {
			bar: 'bg-orange-500 dark:bg-orange-400',
			gradient: 'from-orange-500 to-orange-600',
			line: 'text-orange-500 dark:text-orange-400'
		}
	};

	// Line chart path
	const linePath = $derived(() => {
		if (type !== 'line' || chartData.length === 0) return '';

		const width = 100;
		const height = 100;
		const padding = 10;
		const effectiveWidth = width - padding * 2;
		const effectiveHeight = height - padding * 2;

		const points = chartData.map((d, i) => {
			const x = padding + (i / (chartData.length - 1)) * effectiveWidth;
			const y = padding + effectiveHeight - (d.value / maxValue) * effectiveHeight;
			return `${x},${y}`;
		});

		return `M ${points.join(' L ')}`;
	});

	// Calculate total and average
	const total = $derived(chartData.reduce((sum, d) => sum + d.value, 0));
	const average = $derived(Math.round(total / chartData.length));
</script>

<div
	class="bg-white dark:bg-[#1c1c1e] rounded-xl border border-gray-200 dark:border-white/10 p-5 shadow-sm hover:shadow-md transition-shadow duration-300"
>
	<!-- Header -->
	<div class="flex items-center justify-between mb-4">
		<div class="flex items-center gap-2">
			<div
				class="w-8 h-8 rounded-lg bg-gradient-to-br {colorClasses[color]
					.gradient} flex items-center justify-center shadow-sm"
			>
				<svg class="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
					/>
				</svg>
			</div>
			<h2 class="text-base font-semibold text-gray-900 dark:text-white/90">{title}</h2>
		</div>
		<div class="text-xs text-gray-500 dark:text-gray-400">Last 7 days</div>
	</div>

	<!-- Stats -->
	<div class="grid grid-cols-2 gap-3 mb-4">
		<div class="bg-gray-50 dark:bg-white/5 rounded-lg p-3">
			<div class="text-xs text-gray-500 dark:text-gray-400 mb-1">Total</div>
			<div class="text-lg font-bold text-gray-900 dark:text-white">{total}</div>
		</div>
		<div class="bg-gray-50 dark:bg-white/5 rounded-lg p-3">
			<div class="text-xs text-gray-500 dark:text-gray-400 mb-1">Daily Avg</div>
			<div class="text-lg font-bold text-gray-900 dark:text-white">{average}</div>
		</div>
	</div>

	<!-- Chart -->
	{#if type === 'bar'}
		<div class="space-y-2" transition:fade={{ duration: 300 }}>
			{#each chartData as point, index (point.label)}
				<div class="flex items-center gap-2" in:fly={{ x: -10, duration: 300, delay: index * 50 }}>
					<div class="text-xs text-gray-500 dark:text-gray-400 w-10">{point.label}</div>
					<div class="flex-1 bg-gray-100 dark:bg-white/5 rounded-full h-6 overflow-hidden">
						<div
							class="{colorClasses[color].bar} h-full rounded-full transition-all duration-500"
							style="width: {(point.value / maxValue) * 100}%"
						></div>
					</div>
					<div class="text-xs font-medium text-gray-700 dark:text-gray-300 w-8 text-right">
						{point.value}
					</div>
				</div>
			{/each}
		</div>
	{:else}
		<div class="w-full h-32" transition:fade={{ duration: 300 }}>
			<svg viewBox="0 0 100 100" class="w-full h-full">
				<!-- Grid lines -->
				{#each [0.25, 0.5, 0.75] as line}
					<line
						x1="10"
						y1={10 + (100 - 20) * line}
						x2="90"
						y2={10 + (100 - 20) * line}
						stroke="currentColor"
						stroke-width="0.2"
						class="text-gray-200 dark:text-gray-700"
					/>
				{/each}

				<!-- Line -->
				<path
					d={linePath}
					fill="none"
					stroke="currentColor"
					stroke-width="2"
					class={colorClasses[color].line}
					vector-effect="non-scaling-stroke"
				/>

				<!-- Fill area -->
				<path
					d={linePath + ` L 90,90 L 10,90 Z`}
					fill="currentColor"
					class="{colorClasses[color].line}/10"
				/>

				<!-- Data points -->
				{#each chartData as point, i}
					{@const x = 10 + (i / (chartData.length - 1)) * 80}
					{@const y = 10 + (100 - 20) - (point.value / maxValue) * (100 - 20)}
					<circle
						cx={x}
						cy={y}
						r="2"
						fill="currentColor"
						class={colorClasses[color].line}
					/>
				{/each}
			</svg>

			<!-- Labels -->
			<div class="flex justify-between mt-2">
				{#each chartData as point}
					<div class="text-xs text-gray-500 dark:text-gray-400">{point.label}</div>
				{/each}
			</div>
		</div>
	{/if}
</div>
