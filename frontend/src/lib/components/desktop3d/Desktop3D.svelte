<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { Canvas } from '@threlte/core';
	import { desktop3dStore, openWindows, focusedWindow, type ModuleId, ALL_MODULES, MODULE_INFO } from '$lib/stores/desktop3dStore';
	import Desktop3DScene from './Desktop3DScene.svelte';
	import Desktop3DControls from './Desktop3DControls.svelte';
	import Desktop3DDock from './Desktop3DDock.svelte';
	import MenuBar from '$lib/components/desktop/MenuBar.svelte';

	interface Props {
		onExit?: () => void;
	}

	let { onExit }: Props = $props();

	// Initialize store on mount
	onMount(() => {
		desktop3dStore.initialize();
	});

	// Keyboard shortcuts
	function handleKeydown(e: KeyboardEvent) {
		// Escape - unfocus or exit
		if (e.key === 'Escape') {
			e.preventDefault();
			if ($desktop3dStore.focusedWindowId) {
				desktop3dStore.unfocusWindow();
			} else {
				onExit?.();
			}
		}

		// Space - toggle view mode (only when not focused)
		if (e.key === ' ' && !$desktop3dStore.focusedWindowId) {
			e.preventDefault();
			desktop3dStore.toggleViewMode();
		}

		// Arrow keys - navigate between windows when focused
		if ($desktop3dStore.focusedWindowId) {
			if (e.key === 'ArrowRight') {
				e.preventDefault();
				desktop3dStore.focusNext();
			} else if (e.key === 'ArrowLeft') {
				e.preventDefault();
				desktop3dStore.focusPrevious();
			}
			// +/- keys for resize
			if (e.key === '+' || e.key === '=') {
				e.preventDefault();
				desktop3dStore.resizeFocusedWindow(100, 75);
			} else if (e.key === '-') {
				e.preventDefault();
				desktop3dStore.resizeFocusedWindow(-100, -75);
			}
		}

		// Number keys 1-9 - focus window by index
		if (e.key >= '1' && e.key <= '9' && !$desktop3dStore.focusedWindowId) {
			const index = parseInt(e.key) - 1;
			const windows = $openWindows;
			if (windows[index]) {
				desktop3dStore.focusWindow(windows[index].id);
			}
		}
	}

	// Handle window focus from dock
	function handleDockSelect(module: ModuleId) {
		const window = $openWindows.find(w => w.module === module);
		if (window) {
			desktop3dStore.focusWindow(window.id);
		} else {
			desktop3dStore.openWindow(module);
		}
	}

	// Handle view mode toggle
	function handleToggleView() {
		desktop3dStore.toggleViewMode();
	}

	// Handle exit
	function handleExit() {
		onExit?.();
	}
</script>

<svelte:window onkeydown={handleKeydown} />

<div class="desktop-3d">
	<!-- Top Navigation (same as normal desktop) -->
	<MenuBar />

	<!-- 3D Canvas -->
	<div class="canvas-container">
		<Canvas>
			<Desktop3DScene
				windows={$openWindows}
				viewMode={$desktop3dStore.viewMode}
				focusedWindowId={$desktop3dStore.focusedWindowId}
				autoRotate={$desktop3dStore.autoRotate}
				sphereRadius={$desktop3dStore.sphereRadius}
				onWindowClick={(id) => {
					// Always focus the clicked window (smooth transition via springs)
					// If clicking the same window, nothing happens (iframe handles those clicks)
					// If clicking a different window, smoothly transition to it
					desktop3dStore.focusWindow(id);
				}}
				onBackgroundClick={() => {
					if ($desktop3dStore.focusedWindowId) {
						desktop3dStore.unfocusWindow();
					}
				}}
				onResize={(w, h) => desktop3dStore.resizeFocusedWindow(w, h)}
				onZoomOut={() => {
					// User zoomed out while in focus mode - exit focus
					if ($desktop3dStore.focusedWindowId) {
						desktop3dStore.unfocusWindow();
					}
				}}
			/>
		</Canvas>
	</div>

	<!-- UI Controls Overlay -->
	<Desktop3DControls
		viewMode={$desktop3dStore.viewMode}
		autoRotate={$desktop3dStore.autoRotate}
		hasFocusedWindow={!!$desktop3dStore.focusedWindowId}
		onToggleView={handleToggleView}
		onToggleAutoRotate={() => desktop3dStore.toggleAutoRotate()}
		onExit={handleExit}
	/>

	<!-- Bottom Dock -->
	<Desktop3DDock
		windows={$openWindows}
		focusedWindowId={$desktop3dStore.focusedWindowId}
		onSelect={handleDockSelect}
	/>

	<!-- Focused Window Title -->
	{#if $focusedWindow}
		<div class="focused-title">
			<span class="focused-dot" style="background-color: {$focusedWindow.color}"></span>
			{$focusedWindow.title}
		</div>

		<!-- Navigation Arrows -->
		<button class="nav-arrow nav-arrow-left" onclick={() => desktop3dStore.focusPrevious()}>
			<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<path d="M15 18l-6-6 6-6" />
			</svg>
		</button>
		<button class="nav-arrow nav-arrow-right" onclick={() => desktop3dStore.focusNext()}>
			<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<path d="M9 18l6-6-6-6" />
			</svg>
		</button>
	{/if}
</div>

<style>
	.desktop-3d {
		position: fixed;
		inset: 0;
		/* Knowledge Graph style: white top, gray bottom - floating room effect */
		background: linear-gradient(180deg,
			#ffffff 0%,
			#fafafa 30%,
			#e8e8e8 70%,
			#c8c8c8 100%
		);
		overflow: hidden;
	}

	.canvas-container {
		position: absolute;
		top: 40px; /* Below MenuBar */
		left: 0;
		right: 0;
		bottom: 0;
	}

	.focused-title {
		position: fixed;
		top: 80px;
		left: 50%;
		transform: translateX(-50%);
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 10px 20px;
		background: rgba(255, 255, 255, 0.85);
		backdrop-filter: blur(12px);
		border: 1px solid rgba(0, 0, 0, 0.08);
		border-radius: 12px;
		color: #333333;
		font-size: 16px;
		font-weight: 500;
		pointer-events: none;
		z-index: 100;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
	}

	.focused-dot {
		width: 10px;
		height: 10px;
		border-radius: 50%;
	}

	/* Navigation Arrows */
	.nav-arrow {
		position: fixed;
		top: 50%;
		transform: translateY(-50%);
		width: 60px;
		height: 60px;
		background: rgba(255, 255, 255, 0.9);
		backdrop-filter: blur(12px);
		border: 1px solid rgba(0, 0, 0, 0.1);
		border-radius: 50%;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 200;
		transition: all 0.2s ease;
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
	}

	.nav-arrow:hover {
		background: rgba(255, 255, 255, 1);
		transform: translateY(-50%) scale(1.1);
		box-shadow: 0 6px 20px rgba(0, 0, 0, 0.15);
	}

	.nav-arrow svg {
		width: 28px;
		height: 28px;
		stroke: #333;
	}

	.nav-arrow-left {
		left: 30px;
	}

	.nav-arrow-right {
		right: 30px;
	}

</style>
