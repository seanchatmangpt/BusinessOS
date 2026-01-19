<script lang="ts">
	import { onMount } from 'svelte';
	import { tweened } from 'svelte/motion';
	import { cubicOut } from 'svelte/easing';
	import { fade, fly } from 'svelte/transition';

	interface Props {
		title: string;
		value: number;
		previousValue?: number;
		unit?: string;
		icon?: string;
		color?: 'blue' | 'green' | 'purple' | 'orange' | 'red' | 'indigo';
		trend?: 'up' | 'down' | 'neutral';
		sparklineData?: number[];
		onClick?: () => void;
	}

	let {
		title,
		value,
		previousValue,
		unit = '',
		icon,
		color = 'blue',
		trend,
		sparklineData = [],
		onClick
	}: Props = $props();

	// Animated counter
	const animatedValue = tweened(0, { duration: 1000, easing: cubicOut });

	// Calculate percentage change
	const percentageChange = $derived(() => {
		if (!previousValue || previousValue === 0) return null;
		const change = ((value - previousValue) / previousValue) * 100;
		return Math.round(change);
	});

	// Auto-detect trend if not provided
	const effectiveTrend = $derived(() => {
		if (trend) return trend;
		if (!previousValue) return 'neutral';
		return value > previousValue ? 'up' : value < previousValue ? 'down' : 'neutral';
	});

	// Color mappings
	const colorClasses = {
		blue: 'from-blue-500 to-blue-600',
		green: 'from-green-500 to-green-600',
		purple: 'from-purple-500 to-purple-600',
		orange: 'from-orange-500 to-orange-600',
		red: 'from-red-500 to-red-600',
		indigo: 'from-indigo-500 to-indigo-600'
	};

	const trendColors = {
		up: 'text-green-600 bg-green-50 dark:text-green-400 dark:bg-green-500/10',
		down: 'text-red-600 bg-red-50 dark:text-red-400 dark:bg-red-500/10',
		neutral: 'text-gray-600 bg-gray-50 dark:text-gray-400 dark:bg-gray-500/10'
	};

	// Animate on mount
	onMount(() => {
		animatedValue.set(value);
	});

	// Update animation when value changes
	$effect(() => {
		animatedValue.set(value);
	});

	// Sparkline SVG path
	const sparklinePath = $derived(() => {
		if (sparklineData.length === 0) return '';

		const width = 100;
		const height = 30;
		const max = Math.max(...sparklineData);
		const min = Math.min(...sparklineData);
		const range = max - min || 1;

		const points = sparklineData.map((val, i) => {
			const x = (i / (sparklineData.length - 1)) * width;
			const y = height - ((val - min) / range) * height;
			return `${x},${y}`;
		});

		return `M ${points.join(' L ')}`;
	});
</script>

<button
	onclick={onClick}
	class="w-full bg-white dark:bg-[#1c1c1e] rounded-xl border border-gray-200 dark:border-white/10 p-5 shadow-sm hover:shadow-md transition-all duration-300 text-left {onClick
		? 'cursor-pointer hover:scale-[1.02]'
		: 'cursor-default'}"
	disabled={!onClick}
>
	<!-- Header -->
	<div class="flex items-center justify-between mb-3">
		<div class="flex items-center gap-2">
			{#if icon}
				<div
					class="w-8 h-8 rounded-lg bg-gradient-to-br {colorClasses[
						color
					]} flex items-center justify-center shadow-sm"
				>
					<span class="text-lg">{icon}</span>
				</div>
			{/if}
			<span class="text-sm font-medium text-gray-600 dark:text-gray-400">{title}</span>
		</div>

		{#if percentageChange() !== null}
			<div
				class="flex items-center gap-1 text-xs font-semibold px-2 py-1 rounded-md {trendColors[
					effectiveTrend()
				]}"
				transition:fly={{ y: -5, duration: 300 }}
			>
				{#if effectiveTrend() === 'up'}
					<svg class="w-3 h-3" fill="currentColor" viewBox="0 0 20 20">
						<path
							fill-rule="evenodd"
							d="M5.293 7.707a1 1 0 010-1.414l4-4a1 1 0 011.414 0l4 4a1 1 0 01-1.414 1.414L11 5.414V17a1 1 0 11-2 0V5.414L6.707 7.707a1 1 0 01-1.414 0z"
							clip-rule="evenodd"
						/>
					</svg>
				{:else if effectiveTrend() === 'down'}
					<svg class="w-3 h-3" fill="currentColor" viewBox="0 0 20 20">
						<path
							fill-rule="evenodd"
							d="M14.707 12.293a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 111.414-1.414L9 14.586V3a1 1 0 012 0v11.586l2.293-2.293a1 1 0 011.414 0z"
							clip-rule="evenodd"
						/>
					</svg>
				{:else}
					<svg class="w-3 h-3" fill="currentColor" viewBox="0 0 20 20">
						<path
							fill-rule="evenodd"
							d="M3 10a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1z"
							clip-rule="evenodd"
						/>
					</svg>
				{/if}
				{Math.abs(percentageChange() || 0)}%
			</div>
		{/if}
	</div>

	<!-- Value -->
	<div class="mb-3">
		<div class="text-3xl font-bold text-gray-900 dark:text-white">
			{Math.round($animatedValue)}{unit}
		</div>
		{#if previousValue !== undefined}
			<div class="text-xs text-gray-500 dark:text-gray-400 mt-1">
				vs {previousValue}{unit} {trend === 'up' ? 'yesterday' : 'before'}
			</div>
		{/if}
	</div>

	<!-- Sparkline -->
	{#if sparklineData.length > 0}
		<div class="w-full h-8" transition:fade={{ duration: 300 }}>
			<svg viewBox="0 0 100 30" class="w-full h-full" preserveAspectRatio="none">
				<path
					d={sparklinePath}
					fill="none"
					stroke="currentColor"
					stroke-width="2"
					class="text-{color}-500 dark:text-{color}-400"
					vector-effect="non-scaling-stroke"
				/>
				<path
					d={sparklinePath + ` L 100,30 L 0,30 Z`}
					fill="currentColor"
					class="text-{color}-500/10 dark:text-{color}-400/10"
				/>
			</svg>
		</div>
	{/if}
</button>
