<script lang="ts">
	/**
	 * Draggable and resizable widget wrapper for canvas
	 * Based on Window.svelte drag/resize patterns
	 */
	import type { WidgetLayout } from '$lib/stores/dashboardLayoutStore';
	import type { Snippet } from 'svelte';

	interface Props {
		widget: WidgetLayout;
		isEditMode: boolean;
		zoom: number;
		snapToGrid?: boolean;
		gridSize?: number;
		onMove?: (x: number, y: number) => void;
		onResize?: (width: number, height: number) => void;
		onClick?: () => void;
		onRemove?: () => void;
		children: Snippet;
	}

	let {
		widget,
		isEditMode,
		zoom,
		snapToGrid = false,
		gridSize = 50,
		onMove,
		onResize,
		onClick,
		onRemove,
		children
	}: Props = $props();

	// Drag state
	let isDragging = $state(false);
	let dragOffset = { x: 0, y: 0 };

	// Resize state
	let isResizing = $state(false);
	let resizeDirection = $state('');
	let startBounds = { x: 0, y: 0, width: 0, height: 0 };
	let startMouse = { x: 0, y: 0 };

	// Minimum dimensions
	const MIN_WIDTH = 250;
	const MIN_HEIGHT = 200;

	/**
	 * Snap coordinate to grid
	 */
	function snapToGridValue(value: number): number {
		if (!snapToGrid) return value;
		return Math.round(value / gridSize) * gridSize;
	}

	/**
	 * Handle drag start - title bar drag
	 */
	function handleDragStart(e: MouseEvent | PointerEvent) {
		if (!isEditMode) return;
		if ((e.target as HTMLElement).closest('.widget-controls')) return;
		if ((e.target as HTMLElement).closest('.resize-handle')) return;

		isDragging = true;
		dragOffset = {
			x: (e.clientX / zoom) - widget.x,
			y: (e.clientY / zoom) - widget.y
		};

		onClick?.();

		document.addEventListener('mousemove', handleDragMove);
		document.addEventListener('mouseup', handleDragEnd);
		e.preventDefault();
	}

	/**
	 * Handle drag move
	 */
	function handleDragMove(e: MouseEvent) {
		if (!isDragging) return;

		// Convert screen coordinates to canvas coordinates (account for zoom)
		let newX = (e.clientX / zoom) - dragOffset.x;
		let newY = (e.clientY / zoom) - dragOffset.y;

		// Apply snap to grid
		if (snapToGrid) {
			newX = snapToGridValue(newX);
			newY = snapToGridValue(newY);
		}

		// Prevent dragging off-canvas (optional bounds)
		newX = Math.max(0, newX);
		newY = Math.max(0, newY);

		onMove?.(newX, newY);
	}

	/**
	 * Handle drag end
	 */
	function handleDragEnd() {
		isDragging = false;
		document.removeEventListener('mousemove', handleDragMove);
		document.removeEventListener('mouseup', handleDragEnd);
	}

	/**
	 * Handle resize start
	 */
	function handleResizeStart(e: MouseEvent, direction: string) {
		if (!isEditMode) return;

		isResizing = true;
		resizeDirection = direction;
		startBounds = {
			x: widget.x,
			y: widget.y,
			width: widget.width,
			height: widget.height
		};
		startMouse = { x: e.clientX, y: e.clientY };

		onClick?.();

		document.addEventListener('mousemove', handleResizeMove);
		document.addEventListener('mouseup', handleResizeEnd);
		e.preventDefault();
		e.stopPropagation();
	}

	/**
	 * Handle resize move
	 */
	function handleResizeMove(e: MouseEvent) {
		if (!isResizing) return;

		// Convert screen deltas to canvas deltas (account for zoom)
		const deltaX = (e.clientX - startMouse.x) / zoom;
		const deltaY = (e.clientY - startMouse.y) / zoom;

		let newX = startBounds.x;
		let newY = startBounds.y;
		let newWidth = startBounds.width;
		let newHeight = startBounds.height;

		// East (right edge)
		if (resizeDirection.includes('e')) {
			newWidth = Math.max(MIN_WIDTH, startBounds.width + deltaX);
		}

		// West (left edge)
		if (resizeDirection.includes('w')) {
			const maxDelta = startBounds.width - MIN_WIDTH;
			const actualDelta = Math.min(deltaX, maxDelta);
			newX = startBounds.x + actualDelta;
			newWidth = startBounds.width - actualDelta;
		}

		// South (bottom edge)
		if (resizeDirection.includes('s')) {
			newHeight = Math.max(MIN_HEIGHT, startBounds.height + deltaY);
		}

		// North (top edge)
		if (resizeDirection.includes('n')) {
			const maxDelta = startBounds.height - MIN_HEIGHT;
			const actualDelta = Math.min(deltaY, maxDelta);
			newY = startBounds.y + actualDelta;
			newHeight = startBounds.height - actualDelta;
		}

		// Apply snap to grid
		if (snapToGrid) {
			newX = snapToGridValue(newX);
			newY = snapToGridValue(newY);
			newWidth = snapToGridValue(newWidth);
			newHeight = snapToGridValue(newHeight);
		}

		// Update position if resizing from top or left
		if (resizeDirection.includes('n') || resizeDirection.includes('w')) {
			onMove?.(newX, newY);
		}

		onResize?.(newWidth, newHeight);
	}

	/**
	 * Handle resize end
	 */
	function handleResizeEnd() {
		isResizing = false;
		resizeDirection = '';
		document.removeEventListener('mousemove', handleResizeMove);
		document.removeEventListener('mouseup', handleResizeEnd);
	}

	// Widget transform style (zoom is applied to parent canvas layer)
	const transformStyle = $derived(
		`translate(${widget.x}px, ${widget.y}px)`
	);

	// Cursor during drag/resize
	const cursor = $derived(
		isDragging ? 'grabbing' : isEditMode ? 'grab' : 'default'
	);
</script>

<div
	class="canvas-widget absolute"
	style:transform={transformStyle}
	style:transform-origin="top left"
	style:z-index={widget.zIndex}
	style:cursor={cursor}
	style:width="{widget.width}px"
	style:height="{widget.height}px"
	onpointerdown={handleDragStart}
	role="button"
	tabindex={isEditMode ? 0 : -1}
	aria-label={widget.title}
>
	<!-- Widget content wrapper -->
	<div
		class="widget-content bg-white dark:bg-[#1c1c1e] rounded-xl border border-gray-200 dark:border-white/10 shadow-lg overflow-hidden h-full transition-shadow duration-200"
		class:ring-2={isEditMode && (isDragging || isResizing)}
		class:ring-blue-500={isEditMode && (isDragging || isResizing)}
		class:hover:shadow-xl={isEditMode}
	>
		<!-- Edit mode controls -->
		{#if isEditMode}
			<div
				class="widget-controls absolute top-2 right-2 z-10 flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity"
			>
				{#if onRemove}
					<button
						onclick={(e) => {
							e.stopPropagation();
							onRemove();
						}}
						class="w-6 h-6 flex items-center justify-center bg-red-500 hover:bg-red-600 text-white rounded-md transition-colors"
						aria-label="Remove widget"
						title="Remove widget"
					>
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M6 18L18 6M6 6l12 12"
							/>
						</svg>
					</button>
				{/if}
			</div>
		{/if}

		<!-- Widget content -->
		<div class="widget-inner h-full overflow-auto">
			{@render children()}
		</div>

		<!-- Resize handles (8-direction) -->
		{#if isEditMode}
			<!-- Edges -->
			<div
				class="resize-handle resize-n"
				role="separator"
				aria-orientation="horizontal"
				onmousedown={(e) => handleResizeStart(e, 'n')}
			></div>
			<div
				class="resize-handle resize-s"
				role="separator"
				aria-orientation="horizontal"
				onmousedown={(e) => handleResizeStart(e, 's')}
			></div>
			<div
				class="resize-handle resize-e"
				role="separator"
				aria-orientation="vertical"
				onmousedown={(e) => handleResizeStart(e, 'e')}
			></div>
			<div
				class="resize-handle resize-w"
				role="separator"
				aria-orientation="vertical"
				onmousedown={(e) => handleResizeStart(e, 'w')}
			></div>

			<!-- Corners -->
			<div
				class="resize-handle resize-ne"
				role="separator"
				onmousedown={(e) => handleResizeStart(e, 'ne')}
			></div>
			<div
				class="resize-handle resize-nw"
				role="separator"
				onmousedown={(e) => handleResizeStart(e, 'nw')}
			></div>
			<div
				class="resize-handle resize-se"
				role="separator"
				onmousedown={(e) => handleResizeStart(e, 'se')}
			></div>
			<div
				class="resize-handle resize-sw"
				role="separator"
				onmousedown={(e) => handleResizeStart(e, 'sw')}
			></div>
		{/if}
	</div>
</div>

<style>
	.canvas-widget {
		user-select: none;
		will-change: transform;
		transition: box-shadow 0.2s ease;
	}

	.canvas-widget:hover .widget-controls {
		opacity: 1;
	}

	/* Resize handles */
	.resize-handle {
		position: absolute;
		z-index: 20;
		background: transparent;
	}

	.resize-handle:hover {
		background: rgba(59, 130, 246, 0.2);
	}

	/* Edge handles */
	.resize-n {
		top: 0;
		left: 10px;
		right: 10px;
		height: 6px;
		cursor: ns-resize;
	}

	.resize-s {
		bottom: 0;
		left: 10px;
		right: 10px;
		height: 6px;
		cursor: ns-resize;
	}

	.resize-e {
		right: 0;
		top: 10px;
		bottom: 10px;
		width: 6px;
		cursor: ew-resize;
	}

	.resize-w {
		left: 0;
		top: 10px;
		bottom: 10px;
		width: 6px;
		cursor: ew-resize;
	}

	/* Corner handles */
	.resize-ne {
		top: 0;
		right: 0;
		width: 12px;
		height: 12px;
		cursor: nesw-resize;
	}

	.resize-nw {
		top: 0;
		left: 0;
		width: 12px;
		height: 12px;
		cursor: nwse-resize;
	}

	.resize-se {
		bottom: 0;
		right: 0;
		width: 12px;
		height: 12px;
		cursor: nwse-resize;
	}

	.resize-sw {
		bottom: 0;
		left: 0;
		width: 12px;
		height: 12px;
		cursor: nesw-resize;
	}
</style>
