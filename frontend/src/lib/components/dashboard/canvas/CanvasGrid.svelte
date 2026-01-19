<script lang="ts">
	/**
	 * Infinite dot grid background for Miro-style canvas
	 * Renders an SVG pattern that moves with pan/zoom
	 */
	interface Props {
		viewportX: number;
		viewportY: number;
		zoom: number;
		dotSize?: number;
		dotSpacing?: number;
		dotColor?: string;
		showGrid?: boolean;
	}

	let {
		viewportX,
		viewportY,
		zoom,
		dotSize = 2,
		dotSpacing = 50,
		dotColor = 'currentColor',
		showGrid = true
	}: Props = $props();

	// Calculate pattern offset based on viewport position
	const patternOffsetX = $derived((viewportX * zoom) % (dotSpacing * zoom));
	const patternOffsetY = $derived((viewportY * zoom) % (dotSpacing * zoom));

	// Fade dots at low zoom levels for cleaner appearance
	const dotOpacity = $derived(Math.max(0.15, Math.min(0.4, zoom * 0.4)));
</script>

{#if showGrid}
	<svg
		class="canvas-grid absolute inset-0 w-full h-full pointer-events-none select-none"
		aria-hidden="true"
	>
		<defs>
			<pattern
				id="dot-grid-pattern"
				width={dotSpacing * zoom}
				height={dotSpacing * zoom}
				patternUnits="userSpaceOnUse"
				x={patternOffsetX}
				y={patternOffsetY}
			>
				<circle
					cx={dotSpacing * zoom / 2}
					cy={dotSpacing * zoom / 2}
					r={dotSize * Math.min(zoom, 1)}
					fill={dotColor}
					opacity={dotOpacity}
					class="text-gray-400 dark:text-gray-600"
				/>
			</pattern>
		</defs>
		<rect width="100%" height="100%" fill="url(#dot-grid-pattern)" />
	</svg>
{/if}

<style>
	.canvas-grid {
		z-index: 0;
	}
</style>
