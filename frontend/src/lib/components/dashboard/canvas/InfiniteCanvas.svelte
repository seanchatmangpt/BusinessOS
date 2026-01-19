<script lang="ts">
	/**
	 * Infinite canvas container with pan/zoom
	 * Similar to Miro's infinite whiteboard
	 */
	import CanvasGrid from './CanvasGrid.svelte';
	import type { WidgetLayout } from '$lib/stores/dashboardLayoutStore';
	import type { Snippet } from 'svelte';

	interface Props {
		widgets: WidgetLayout[];
		viewport: { offsetX: number; offsetY: number; zoom: number };
		gridConfig: {
			showGrid: boolean;
			cellSize: number;
			gridColor: string;
			spacing: number;
		};
		isEditMode: boolean;
		onViewportChange?: (viewport: { offsetX: number; offsetY: number; zoom: number }) => void;
		onZoomIn?: () => void;
		onZoomOut?: () => void;
		onResetView?: () => void;
		onFitAll?: () => void;
		children: Snippet<[WidgetLayout]>;
	}

	let {
		widgets,
		viewport,
		gridConfig,
		isEditMode,
		onViewportChange,
		onZoomIn = undefined,
		onZoomOut = undefined,
		onResetView = undefined,
		onFitAll = undefined,
		children
	}: Props = $props();

	// Pan state
	let isPanning = $state(false);
	let panStartX = 0;
	let panStartY = 0;
	let startOffsetX = 0;
	let startOffsetY = 0;

	// Zoom constraints
	const MIN_ZOOM = 0.25; // Allow zooming out to 25% to see more
	const MAX_ZOOM = 3.0;  // Allow zooming in to 300% for details
	const ZOOM_STEP = 0.15; // Smoother zoom increments

	// Canvas container ref
	let canvasContainer: HTMLDivElement;

	/**
	 * Handle pan start - middle mouse or space + left mouse
	 */
	function handlePanStart(e: MouseEvent) {
		// Allow pan with middle mouse button or space + left mouse
		const isPanTrigger = e.button === 1 || (e.button === 0 && e.shiftKey);

		if (!isPanTrigger && !(!isEditMode && e.button === 0)) return;

		isPanning = true;
		panStartX = e.clientX;
		panStartY = e.clientY;
		startOffsetX = viewport.offsetX;
		startOffsetY = viewport.offsetY;

		document.addEventListener('mousemove', handlePanMove);
		document.addEventListener('mouseup', handlePanEnd);

		e.preventDefault();
	}

	/**
	 * Handle pan move
	 */
	function handlePanMove(e: MouseEvent) {
		if (!isPanning) return;

		const deltaX = e.clientX - panStartX;
		const deltaY = e.clientY - panStartY;

		onViewportChange?.({
			offsetX: startOffsetX + deltaX,
			offsetY: startOffsetY + deltaY,
			zoom: viewport.zoom
		});
	}

	/**
	 * Handle pan end
	 */
	function handlePanEnd() {
		isPanning = false;
		document.removeEventListener('mousemove', handlePanMove);
		document.removeEventListener('mouseup', handlePanEnd);
	}

	/**
	 * Handle zoom via mouse wheel
	 * Zooms from the center of the viewport for better UX
	 */
	function handleWheel(e: WheelEvent) {
		e.preventDefault();

		// Calculate zoom delta
		const delta = e.deltaY > 0 ? -ZOOM_STEP : ZOOM_STEP;
		const newZoom = Math.min(Math.max(viewport.zoom + delta, MIN_ZOOM), MAX_ZOOM);

		// Zoom from center of viewport
		const rect = canvasContainer.getBoundingClientRect();
		const centerX = rect.width / 2;
		const centerY = rect.height / 2;

		// Calculate new offset to keep center point stable
		const zoomRatio = newZoom / viewport.zoom;
		const newOffsetX = centerX - (centerX - viewport.offsetX) * zoomRatio;
		const newOffsetY = centerY - (centerY - viewport.offsetY) * zoomRatio;

		onViewportChange?.({
			offsetX: newOffsetX,
			offsetY: newOffsetY,
			zoom: newZoom
		});
	}

	/**
	 * Zoom in
	 */
	function zoomIn() {
		const newZoom = Math.min(viewport.zoom + ZOOM_STEP, MAX_ZOOM);
		onViewportChange?.({
			offsetX: viewport.offsetX,
			offsetY: viewport.offsetY,
			zoom: newZoom
		});
		onZoomIn?.(); // Notify parent
	}

	/**
	 * Zoom out
	 */
	function zoomOut() {
		const newZoom = Math.max(viewport.zoom - ZOOM_STEP, MIN_ZOOM);
		onViewportChange?.({
			offsetX: viewport.offsetX,
			offsetY: viewport.offsetY,
			zoom: newZoom
		});
		onZoomOut?.(); // Notify parent
	}

	/**
	 * Reset viewport to default
	 */
	function resetView() {
		onViewportChange?.({
			offsetX: 0,
			offsetY: 0,
			zoom: 1.0
		});
		onResetView?.(); // Notify parent
	}

	/**
	 * Fit all widgets in view
	 */
	function fitAll() {
		if (widgets.length === 0) {
			resetView();
			return;
		}

		// Calculate bounding box of all widgets
		const minX = Math.min(...widgets.map((w) => w.x));
		const minY = Math.min(...widgets.map((w) => w.y));
		const maxX = Math.max(...widgets.map((w) => w.x + w.width));
		const maxY = Math.max(...widgets.map((w) => w.y + w.height));

		const contentWidth = maxX - minX;
		const contentHeight = maxY - minY;

		const rect = canvasContainer.getBoundingClientRect();
		const padding = 100;

		// Calculate zoom to fit
		const zoomX = (rect.width - padding * 2) / contentWidth;
		const zoomY = (rect.height - padding * 2) / contentHeight;
		const newZoom = Math.min(Math.max(Math.min(zoomX, zoomY), MIN_ZOOM), MAX_ZOOM);

		// Center the content
		const centerX = minX + contentWidth / 2;
		const centerY = minY + contentHeight / 2;
		const newOffsetX = rect.width / 2 - centerX * newZoom;
		const newOffsetY = rect.height / 2 - centerY * newZoom;

		onViewportChange?.({
			offsetX: newOffsetX,
			offsetY: newOffsetY,
			zoom: newZoom
		});
		onFitAll?.(); // Notify parent
	}

	// Cursor during pan
	const canvasCursor = $derived(
		isPanning ? 'grabbing' : isEditMode ? 'default' : 'grab'
	);
</script>

<div
	bind:this={canvasContainer}
	class="infinite-canvas relative w-full h-full overflow-hidden bg-white dark:bg-[#0a0a0a]"
	style:cursor={canvasCursor}
	onmousedown={handlePanStart}
	onwheel={handleWheel}
	role="application"
	aria-label="Infinite canvas dashboard"
>
	<!-- Grid background -->
	<CanvasGrid
		viewportX={viewport.offsetX}
		viewportY={viewport.offsetY}
		zoom={viewport.zoom}
		dotSpacing={gridConfig.cellSize}
		dotColor={gridConfig.gridColor}
		showGrid={gridConfig.showGrid}
	/>

	<!-- Canvas layer (widgets) -->
	<div
		class="canvas-layer absolute inset-0"
		style:transform="translate({viewport.offsetX}px, {viewport.offsetY}px) scale({viewport.zoom})"
		style:transform-origin="top left"
	>
		{#each widgets as widget (widget.id)}
			{@render children(widget)}
		{/each}
	</div>
</div>

<style>
	.infinite-canvas {
		position: relative;
		touch-action: none;
	}

	.canvas-layer {
		pointer-events: none;
	}

	.canvas-layer > :global(*) {
		pointer-events: auto;
	}
</style>
