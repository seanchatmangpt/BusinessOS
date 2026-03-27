<script lang="ts">
	import type { QualityMetrics } from '$lib/api/mesh';

	export let quality: QualityMetrics | null = null;

	function getColorClass(score: number): string {
		if (score >= 80) return 'status-good';
		if (score >= 60) return 'status-warning';
		return 'status-critical';
	}

	function getStatusLabel(score: number): string {
		if (score >= 80) return 'Good';
		if (score >= 60) return 'Fair';
		return 'Poor';
	}

	// Calculate polar coordinates for radar chart
	// 5 dimensions: completeness, accuracy, consistency, timeliness, overall
	function getPolarCoord(index: number, radius: number): { x: number; y: number } {
		const angle = (index * 72) * (Math.PI / 180); // 360 / 5 dimensions
		const centerX = 100;
		const centerY = 100;
		return {
			x: centerX + radius * Math.sin(angle),
			y: centerY - radius * Math.cos(angle)
		};
	}

	function getRadarPath(metrics: number[]): string {
		const radius = 60;
		const coords = metrics.map((_, i) => getPolarCoord(i, (metrics[i] / 100) * radius));
		return coords.map((c, i) => `${i === 0 ? 'M' : 'L'} ${c.x} ${c.y}`).join(' ') + ' Z';
	}

	$: radarPath = quality
		? getRadarPath([quality.completeness, quality.accuracy, quality.consistency, quality.timeliness])
		: '';
</script>

<div class="quality-scoreboard">
	{#if !quality}
		<div class="empty-state">
			<p>No quality metrics available</p>
		</div>
	{:else}
		<div class="scoreboard-content">
			<!-- Overall score card -->
			<div class="overall-score">
				<div class="score-circle" class:good={quality.overall >= 80} class:fair={quality.overall >= 60 && quality.overall < 80} class:poor={quality.overall < 60}>
					<div class="score-number">{Math.round(quality.overall)}</div>
					<div class="score-label">Overall Quality</div>
				</div>
				<div class="score-status" class={getColorClass(quality.overall)}>
					{getStatusLabel(quality.overall)}
				</div>
			</div>

			<!-- Radar chart (polar coordinates) -->
			<div class="radar-chart">
				<svg viewBox="0 0 200 200" width="200" height="200">
					<!-- Grid circles -->
					{#each [20, 40, 60, 80, 100] as radius}
						<circle cx="100" cy="100" r={radius} fill="none" stroke="#e5e7eb" stroke-width="0.5" opacity="0.5" />
					{/each}

					<!-- Axis lines -->
					{#each [0, 1, 2, 3] as i}
						{@const coord = getPolarCoord(i, 60)}
						<line x1="100" y1="100" x2={coord.x} y2={coord.y} stroke="#e5e7eb" stroke-width="0.5" />
					{/each}

					<!-- Data polygon -->
					<path
						d={radarPath}
						fill="#3b82f6"
						fill-opacity="0.2"
						stroke="#3b82f6"
						stroke-width="2"
					/>

					<!-- Data points -->
					{#each [quality.completeness, quality.accuracy, quality.consistency, quality.timeliness] as value, i}
						{@const coord = getPolarCoord(i, (value / 100) * 60)}
						<circle cx={coord.x} cy={coord.y} r="3" fill="#3b82f6" />
					{/each}

					<!-- Axis labels -->
					{#each ['Completeness', 'Accuracy', 'Consistency', 'Timeliness'] as label, i}
						{@const coord = getPolarCoord(i, 75)}
						<text
							x={coord.x}
							y={coord.y}
							text-anchor="middle"
							font-size="10"
							fill="#6b7280"
							dominant-baseline="middle"
						>
							{label}
						</text>
					{/each}
				</svg>
			</div>

			<!-- Individual metrics -->
			<div class="metrics-grid">
				{#each [
					{ label: 'Completeness', value: quality.completeness },
					{ label: 'Accuracy', value: quality.accuracy },
					{ label: 'Consistency', value: quality.consistency },
					{ label: 'Timeliness', value: quality.timeliness }
				] as metric}
					<div class="metric-card {getColorClass(metric.value)}">
						<div class="metric-label">{metric.label}</div>
						<div class="metric-bar">
							<div
								class="metric-fill"
								style="width: {metric.value}%"
							/>
						</div>
						<div class="metric-value">{Math.round(metric.value)}%</div>
					</div>
				{/each}
			</div>

			<!-- Timestamp -->
			<div class="timestamp">
				Last calculated: {new Date(quality.last_calculated).toLocaleString()}
			</div>
		</div>
	{/if}
</div>

<style>
	.quality-scoreboard {
		width: 100%;
		padding: 16px;
		background: white;
		border: 1px solid #e5e7eb;
		border-radius: 8px;
	}

	.empty-state {
		display: flex;
		align-items: center;
		justify-content: center;
		height: 200px;
		color: #9ca3af;
		font-size: 14px;
	}

	.scoreboard-content {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 24px;
	}

	.overall-score {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 12px;
	}

	.score-circle {
		width: 150px;
		height: 150px;
		border-radius: 50%;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		border: 4px solid #e5e7eb;
		background: #f9fafb;
	}

	.score-circle.good {
		border-color: #10b981;
		background: #ecfdf5;
	}

	.score-circle.fair {
		border-color: #f59e0b;
		background: #fffbeb;
	}

	.score-circle.poor {
		border-color: #ef4444;
		background: #fef2f2;
	}

	.score-number {
		font-size: 48px;
		font-weight: 700;
		color: #1f2937;
	}

	.score-label {
		font-size: 12px;
		color: #6b7280;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.score-status {
		padding: 8px 16px;
		border-radius: 6px;
		font-size: 14px;
		font-weight: 600;
	}

	.status-good {
		background: #ecfdf5;
		color: #065f46;
	}

	.status-warning {
		background: #fffbeb;
		color: #92400e;
	}

	.status-critical {
		background: #fef2f2;
		color: #7f1d1d;
	}

	.radar-chart {
		display: flex;
		justify-content: center;
		align-items: center;
	}

	.radar-chart svg {
		max-width: 100%;
		height: auto;
	}

	.metrics-grid {
		grid-column: 1 / -1;
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
		gap: 12px;
	}

	.metric-card {
		padding: 12px;
		background: #f9fafb;
		border: 1px solid #e5e7eb;
		border-radius: 6px;
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.metric-card.status-good {
		border-color: #10b981;
		background: #ecfdf5;
	}

	.metric-card.status-warning {
		border-color: #f59e0b;
		background: #fffbeb;
	}

	.metric-card.status-critical {
		border-color: #ef4444;
		background: #fef2f2;
	}

	.metric-label {
		font-size: 12px;
		font-weight: 600;
		color: #6b7280;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.metric-bar {
		height: 8px;
		background: #e5e7eb;
		border-radius: 4px;
		overflow: hidden;
	}

	.metric-fill {
		height: 100%;
		background: linear-gradient(90deg, #3b82f6, #0ea5e9);
		transition: width 0.3s ease;
	}

	.metric-value {
		font-size: 14px;
		font-weight: 700;
		color: #1f2937;
		text-align: right;
	}

	.timestamp {
		grid-column: 1 / -1;
		font-size: 11px;
		color: #9ca3af;
		text-align: center;
		padding-top: 8px;
		border-top: 1px solid #e5e7eb;
	}
</style>
