<script lang="ts">
	import { T } from '@threlte/core';
	import { HTML } from '@threlte/extras';
	import { spring } from 'svelte/motion';
	import type { Window3DState, ViewMode } from '$lib/stores/desktop3dStore';

	interface Props {
		window: Window3DState;
		isFocused: boolean;
		isNextWindow: boolean;
		isPrevWindow: boolean;
		isHovered: boolean;
		viewMode: ViewMode;
		onClick?: () => void;
		onResize?: (widthDelta: number, heightDelta: number) => void;
		onHover?: (hovered: boolean) => void;
	}

	let {
		window,
		isFocused = false,
		isNextWindow = false,
		isPrevWindow = false,
		isHovered = false,
		viewMode = 'orb',
		onClick,
		onResize,
		onHover
	}: Props = $props();

	// Track pointer down position to distinguish click vs drag
	let pointerDownPos: { x: number; y: number } | null = null;
	const DRAG_THRESHOLD = 15; // pixels - increased for better click detection

	// Spring animation for smooth position transitions
	type Vec3 = [number, number, number];
	let animatedPosition = spring<Vec3>(window.position, {
		stiffness: 0.08,
		damping: 0.7
	});
	let animatedRotation = spring<Vec3>([0, 0, 0], {
		stiffness: 0.08,
		damping: 0.7
	});
	let animatedScale = spring(1, {
		stiffness: 0.1,
		damping: 0.6
	});

	// Handle resize button clicks - must stop ALL event propagation
	function handleSizeIncrease(e: MouseEvent | Event) {
		e.stopPropagation();
		if ('stopImmediatePropagation' in e) e.stopImmediatePropagation();
		e.preventDefault();
		if (onResize) {
			onResize(100, 75);
		}
	}
	function handleSizeDecrease(e: MouseEvent | Event) {
		e.stopPropagation();
		if ('stopImmediatePropagation' in e) e.stopImmediatePropagation();
		e.preventDefault();
		if (onResize) {
			onResize(-100, -75);
		}
	}

	// Current opacity (not animated)
	// All windows always visible, just dimmed in focus mode
	let currentOpacity = $derived.by(() => {
		if (isFocused) return 1;
		if (isNextWindow || isPrevWindow) return 0.9; // Side previews clearly visible
		if (viewMode === 'focused') return 0.3; // Others dimmed but visible
		return 1;
	});

	// Always render all windows
	let shouldRender = true;

	// Track if we're dragging to prevent click on drag end
	let isDragging = $state(false);
	let pointerDownTime = 0;

	// Handle pointer down - track start position
	function handlePointerDown(e: any) {
		isDragging = false;
		pointerDownTime = Date.now();
		// Store the pointer position when pressed
		const event = e.nativeEvent || e;
		if (event.clientX !== undefined) {
			pointerDownPos = { x: event.clientX, y: event.clientY };
		}
	}

	// Handle pointer move - detect if dragging
	function handlePointerMove(e: any) {
		if (!pointerDownPos) return;
		const event = e.nativeEvent || e;
		if (event.clientX !== undefined) {
			const dx = Math.abs(event.clientX - pointerDownPos.x);
			const dy = Math.abs(event.clientY - pointerDownPos.y);
			if (dx > DRAG_THRESHOLD || dy > DRAG_THRESHOLD) {
				isDragging = true;
			}
		}
	}

	// Handle click - only trigger if it wasn't a drag
	function handleClick(e: any) {
		const clickDuration = Date.now() - pointerDownTime;

		// Don't trigger if: was dragging, held too long, or no pointer down recorded
		if (isDragging || clickDuration > 500 || !pointerDownPos) {
			pointerDownPos = null;
			isDragging = false;
			return;
		}

		pointerDownPos = null;
		isDragging = false;
		e.stopPropagation();
		onClick?.();
	}

	// Calculate target position based on state
	function getTargetPosition(): Vec3 {
		if (isFocused) {
			// Focused window: bring forward but keep in 3D context
			// Not too far forward since camera is backed up
			return [0, 0, 80];
		}
		if (isPrevWindow) {
			// Previous window: show on LEFT side
			return [-100, 0, 50];
		}
		if (isNextWindow) {
			// Next window: show on RIGHT side
			return [100, 0, 50];
		}
		// Other windows: stay in orb positions
		return window.position;
	}

	// Calculate target rotation - windows wrap around sphere like stickers
	// Windows stay in LANDSCAPE orientation (flat like monitors)
	function getTargetRotation(): Vec3 {
		// When focused or adjacent, face camera for readability
		if (isFocused || isNextWindow || isPrevWindow) {
			return [0, 0, 0];
		}

		// Use ORIGINAL window position for rotation (not animated)
		const [x, y, z] = window.position;

		// Y rotation only - face outward from sphere center
		// Keep windows FLAT (no X tilt) for proper click detection
		const angleY = Math.atan2(x, z);

		return [0, angleY, 0];
	}

	// Calculate target scale
	function getTargetScale(): number {
		if (isFocused) return 2.5; // Larger since camera is backed up
		if (isNextWindow || isPrevWindow) return 1.0; // Side previews visible
		if (viewMode === 'focused') return 0.7; // Background windows smaller
		return 1;
	}

	// Update springs when state changes
	$effect(() => {
		const targetPos = getTargetPosition();
		animatedPosition.set(targetPos);
	});

	$effect(() => {
		const targetRot = getTargetRotation();
		animatedRotation.set(targetRot);
	});

	$effect(() => {
		const targetScale = getTargetScale();
		animatedScale.set(targetScale);
	});

	// HUGE scale for visible windows - BIGGER for better visibility
	const htmlScale = 2.0;
</script>

<!-- Window using HTML component for DOM content in 3D space -->
<!-- Only render if should be visible -->
{#if shouldRender}
	<!-- Invisible click mesh - ONLY for non-focused, non-adjacent windows in orb mode -->
	{#if !isFocused && !isNextWindow && !isPrevWindow}
		<!-- Click mesh - positioned at window center, faces camera (no rotation) -->
		<T.Mesh
			position={$animatedPosition}
			onpointerdown={(e) => { e.stopPropagation(); handlePointerDown(e); }}
			onpointermove={(e) => { handlePointerMove(e); }}
			onclick={(e) => { e.stopPropagation(); handleClick(e); }}
			onpointerenter={(e) => { e.stopPropagation(); onHover?.(true); }}
			onpointerleave={(e) => { e.stopPropagation(); onHover?.(false); }}
		>
			<T.SphereGeometry args={[50, 12, 12]} />
			<T.MeshBasicMaterial visible={false} transparent opacity={0} />
		</T.Mesh>
	{:else if isFocused}
		<!-- Blocker mesh when focused - prevents background clicks -->
		<T.Mesh
			position={$animatedPosition}
			onclick={(e) => e.stopPropagation()}
		>
			<T.PlaneGeometry args={[window.width * 0.6, window.height * 0.6]} />
			<T.MeshBasicMaterial visible={false} transparent opacity={0} />
		</T.Mesh>
	{/if}
	<!-- NOTE: No click mesh for prev/next windows - they use HTML pointer events -->

	<T.Group position={$animatedPosition} rotation={$animatedRotation}>
	<HTML
		transform
		scale={htmlScale * $animatedScale}
		pointerEvents={isFocused || isNextWindow || isPrevWindow ? 'auto' : 'none'}
		zIndexRange={[100, 0]}
	>
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div
			class="window-wrapper"
			class:focused={isFocused}
			class:dimmed={!isFocused && viewMode === 'focused'}
			class:clickable={isNextWindow || isPrevWindow}
			class:hovered={isHovered}
			style="opacity: {currentOpacity};"
			onclick={(isNextWindow || isPrevWindow) ? onClick : undefined}
		>
			<div
				class="window-3d"
				style="--window-width: {window.width}px; --window-height: {window.height}px;"
			>

				<!-- Title bar -->
				<div class="window-titlebar" style="border-left: 4px solid {window.color};">
					<span class="module-title">{window.title}</span>
					<div class="titlebar-right">
						{#if isFocused}
							<!-- Size controls when focused -->
							<!-- svelte-ignore a11y_no_static_element_interactions -->
							<!-- svelte-ignore a11y_click_events_have_key_events -->
							<div
								class="size-controls"
								onmousedown={(e) => e.stopPropagation()}
								onclick={(e) => e.stopPropagation()}
							>
								<button
									type="button"
									class="size-btn"
									onclick={(e) => handleSizeDecrease(e)}
									onpointerdown={(e) => e.stopPropagation()}
									onmousedown={(e) => e.stopPropagation()}
									title="Smaller (-)"
								>
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<path d="M5 12h14" />
									</svg>
								</button>
								<span class="size-label">{window.width} x {window.height}</span>
								<button
									type="button"
									class="size-btn"
									onclick={(e) => handleSizeIncrease(e)}
									onpointerdown={(e) => e.stopPropagation()}
									onmousedown={(e) => e.stopPropagation()}
									title="Larger (+)"
								>
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<path d="M12 5v14M5 12h14" />
									</svg>
								</button>
							</div>
						{/if}
						<div class="titlebar-controls">
							<span class="control-dot" style="background: #febc2e;"></span>
							<span class="control-dot" style="background: #28c840;"></span>
							<span class="control-dot" style="background: #ff5f57;"></span>
						</div>
					</div>
				</div>

				<!-- LIVE Content - Always show iframe -->
				<div class="window-content">
					<iframe
						src="/{window.module}?embed=true"
						title={window.title}
						class="window-iframe"
					></iframe>
				</div>
			</div>

			<!-- Label under window (hidden when focused) -->
			{#if !isFocused}
				<div class="window-label" style="background-color: {window.color};">
					{window.title}
				</div>
			{/if}
		</div>
	</HTML>
</T.Group>
{/if}

<style>
	.window-wrapper {
		display: flex;
		flex-direction: column;
		align-items: center;
		cursor: pointer;
	}

	.window-wrapper.dimmed {
		opacity: 0.5;
	}

	.window-wrapper.clickable {
		cursor: pointer;
	}

	.window-wrapper.clickable:hover {
		transform: scale(1.02);
	}

	/* Hover effect - shows which window you're about to click */
	.window-wrapper.hovered {
		transform: scale(1.08);
	}

	.window-wrapper.hovered .window-3d {
		box-shadow:
			0 0 0 4px rgba(74, 158, 255, 0.6),
			0 12px 48px rgba(74, 158, 255, 0.3),
			0 4px 12px rgba(0, 0, 0, 0.15);
	}

	.window-wrapper.hovered .window-label {
		transform: scale(1.1);
		box-shadow: 0 6px 20px rgba(0, 0, 0, 0.3);
	}

	.window-3d {
		position: relative;
		/* Width/height set dynamically via style attribute */
		width: var(--window-width, 700px);
		height: var(--window-height, 450px);
		background: white;
		border-radius: 10px;
		overflow: hidden;
		box-shadow:
			0 8px 32px rgba(0, 0, 0, 0.15),
			0 2px 8px rgba(0, 0, 0, 0.1);
		display: flex;
		flex-direction: column;
		transition: box-shadow 0.3s ease, transform 0.3s ease;
		border: 1px solid rgba(0, 0, 0, 0.1);
	}

	.window-label {
		margin-top: 12px;
		padding: 8px 20px;
		border-radius: 20px;
		color: white;
		font-size: 16px;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 1px;
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
	}

	.window-3d:hover {
		box-shadow:
			0 12px 48px rgba(0, 0, 0, 0.2),
			0 4px 12px rgba(0, 0, 0, 0.15);
	}

	.window-wrapper.focused .window-3d {
		cursor: default;
		box-shadow:
			0 20px 60px rgba(0, 0, 0, 0.25),
			0 8px 24px rgba(0, 0, 0, 0.2);
	}

	.window-titlebar {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 8px 12px;
		background: #f5f5f5;
		border-bottom: 1px solid #e0e0e0;
		user-select: none;
		flex-shrink: 0;
	}

	.module-title {
		font-size: 13px;
		font-weight: 600;
		color: #333;
	}

	.titlebar-right {
		display: flex;
		align-items: center;
		gap: 12px;
	}

	.titlebar-controls {
		display: flex;
		gap: 6px;
	}

	.control-dot {
		width: 10px;
		height: 10px;
		border-radius: 50%;
	}

	/* Size controls in titlebar */
	.size-controls {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 2px 8px;
		background: rgba(0, 0, 0, 0.05);
		border-radius: 6px;
		position: relative;
		z-index: 1000;
		pointer-events: auto !important;
	}

	.size-btn {
		width: 24px;
		height: 24px;
		padding: 0;
		background: rgba(0, 0, 0, 0.1);
		border: none;
		border-radius: 4px;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		transition: background 0.2s;
		pointer-events: auto !important;
		position: relative;
		z-index: 1001;
	}

	.size-btn:hover {
		background: rgba(0, 0, 0, 0.15);
	}

	.size-btn svg {
		width: 12px;
		height: 12px;
		stroke: #333;
	}

	.size-label {
		font-size: 10px;
		color: #666;
		min-width: 60px;
		text-align: center;
	}

	.window-content {
		flex: 1;
		overflow: hidden;
		background: white;
	}

	.window-iframe {
		width: 100%;
		height: 100%;
		border: none;
		pointer-events: none;
	}

	.window-wrapper.focused .window-iframe {
		pointer-events: auto;
	}
</style>
