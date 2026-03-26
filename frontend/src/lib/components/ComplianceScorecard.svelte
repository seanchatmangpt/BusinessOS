<script lang="ts">
	import { TrendingUp, TrendingDown, Minus } from 'lucide-svelte';
	import Card from '$lib/ui/card/Card.svelte';

	interface Props {
		framework: string;
		score: number;
		trend: 'up' | 'down' | 'stable';
		lastUpdated: string;
		isSelected?: boolean;
	}

	let { framework = 'SOC2', score = 0, trend = 'stable', lastUpdated = '', isSelected = false }: Props = $props();

	const getScoreColor = (value: number): string => {
		if (value >= 90) return 'text-green-600';
		if (value >= 70) return 'text-yellow-600';
		return 'text-red-600';
	};

	const getBackgroundColor = (value: number): string => {
		if (value >= 90) return 'bg-green-50 border-green-300';
		if (value >= 70) return 'bg-yellow-50 border-yellow-300';
		return 'bg-red-50 border-red-300';
	};

	const getProgressColor = (value: number): string => {
		if (value >= 90) return 'bg-green-500';
		if (value >= 70) return 'bg-yellow-500';
		return 'bg-red-500';
	};

	const getTrendIcon = (trendValue: string) => {
		if (trendValue === 'up') return TrendingUp;
		if (trendValue === 'down') return TrendingDown;
		return Minus;
	};

	const getTrendColor = (trendValue: string): string => {
		if (trendValue === 'up') return 'text-green-600';
		if (trendValue === 'down') return 'text-red-600';
		return 'text-gray-600';
	};

	const formatDate = (dateString: string): string => {
		try {
			const date = new Date(dateString);
			return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
		} catch {
			return 'N/A';
		}
	};

	const TrendIcon = getTrendIcon(trend);
</script>

<Card class={`p-6 border-2 transition-all duration-300 cursor-pointer ${isSelected ? 'ring-2 ring-blue-500 border-blue-400 shadow-lg' : 'hover:shadow-md'} ${getBackgroundColor(score)}`}>
	<div class="flex items-start justify-between mb-4">
		<h3 class="text-lg font-bold text-gray-900">{framework}</h3>
		<div class={getTrendColor(trend)}>
			<TrendIcon class="w-5 h-5" />
		</div>
	</div>

	<div class="mb-4">
		<div class={`text-5xl font-bold ${getScoreColor(score)} mb-1`}>
			{score.toFixed(0)}%
		</div>
		<p class="text-sm text-gray-600">Compliance Score</p>
	</div>

	<div class="mb-4">
		<div class="flex items-center justify-between mb-2">
			<span class="text-xs font-medium text-gray-700">Progress</span>
			<span class="text-xs text-gray-500">
				{#if trend === 'up'}
					<span class="text-green-600">↑ Improving</span>
				{:else if trend === 'down'}
					<span class="text-red-600">↓ Declining</span>
				{:else}
					<span class="text-gray-600">→ Stable</span>
				{/if}
			</span>
		</div>
		<div class="w-full bg-gray-300 rounded-full h-2">
			<div class={`h-2 rounded-full transition-all duration-500 ${getProgressColor(score)}`} style={`width: ${score}%`}></div>
		</div>
	</div>

	<div class="pt-4 border-t border-gray-200">
		<p class="text-xs text-gray-500">Updated {formatDate(lastUpdated)}</p>
	</div>
</Card>
