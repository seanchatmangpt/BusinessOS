<!--
	ModeSelector.svelte
	Compact dropdown to select the 5 OSA modes.
	Fetches available modes from API on mount, falls back to hardcoded list.
-->
<script lang="ts">
	import { onMount } from 'svelte';
	import { osaStore, type OsaMode } from '$lib/stores/osa';

	interface Props {
		compact?: boolean;
	}

	let { compact = false }: Props = $props();

	let isOpen = $state(false);
	let dropdownElement: HTMLDivElement | undefined = $state(undefined);

	let activeMode = $derived($osaStore.activeMode);
	let availableModes = $derived($osaStore.availableModes);

	const FALLBACK_MODES: { mode: OsaMode; label: string; description: string }[] = [
		{ mode: 'BUILD', label: 'BUILD', description: 'Create new modules and features' },
		{ mode: 'ASSIST', label: 'ASSIST', description: 'Help with existing tasks' },
		{ mode: 'ANALYZE', label: 'ANALYZE', description: 'Analyze and surface insights' },
		{ mode: 'EXECUTE', label: 'EXECUTE', description: 'Execute actions and workflows' },
		{ mode: 'MAINTAIN', label: 'MAINTAIN', description: 'Maintain and monitor systems' }
	];

	let modes = $derived(
		availableModes.length > 0
			? availableModes.map((m) => ({
					mode: m.mode as OsaMode,
					label: m.label || m.mode,
					description: m.description
				}))
			: FALLBACK_MODES
	);

	onMount(async () => {
		await osaStore.loadAvailableModes();
	});

	function selectMode(mode: OsaMode) {
		osaStore.setMode(mode);
		isOpen = false;
	}

	function handleKeyDown(e: KeyboardEvent) {
		if (e.key === 'Escape') {
			isOpen = false;
		}
	}

	function handleClickOutside(e: MouseEvent) {
		if (dropdownElement && !dropdownElement.contains(e.target as Node)) {
			isOpen = false;
		}
	}

	$effect(() => {
		if (isOpen) {
			document.addEventListener('click', handleClickOutside, true);
			return () => document.removeEventListener('click', handleClickOutside, true);
		}
	});
</script>

<div
	bind:this={dropdownElement}
	class="mode-selector relative"
	role="combobox"
	aria-expanded={isOpen}
	aria-label="Select OSA mode"
	onkeydown={handleKeyDown}
>
	<!-- Trigger -->
	<button
		class="flex items-center gap-1.5 rounded-full border border-gray-200 bg-white/80 px-3 py-1 text-xs font-medium text-gray-700 transition-colors hover:border-gray-300 hover:bg-white dark:border-gray-700 dark:bg-gray-800/80 dark:text-gray-300 dark:hover:border-gray-600"
		onclick={() => (isOpen = !isOpen)}
		aria-haspopup="listbox"
	>
		<span>{compact ? activeMode.slice(0, 3) : activeMode}</span>
		<svg
			class="h-3 w-3 transition-transform"
			class:rotate-180={isOpen}
			viewBox="0 0 24 24"
			fill="none"
			stroke="currentColor"
			stroke-width="2"
		>
			<polyline points="6 9 12 15 18 9" />
		</svg>
	</button>

	<!-- Dropdown -->
	{#if isOpen}
		<div
			class="absolute bottom-full left-0 z-50 mb-1 min-w-[180px] overflow-hidden rounded-lg border border-gray-200 bg-white/95 shadow-lg backdrop-blur-sm dark:border-gray-700 dark:bg-gray-800/95"
			role="listbox"
			aria-label="OSA modes"
		>
			{#each modes as m}
				<button
					class="flex w-full items-center gap-2 px-3 py-2 text-left text-xs transition-colors hover:bg-gray-100 dark:hover:bg-gray-700/50 {m.mode === activeMode ? 'bg-gray-100 dark:bg-gray-700/50' : ''}"
					role="option"
					aria-selected={m.mode === activeMode}
					onclick={() => selectMode(m.mode)}
				>
					<span class="font-semibold">{m.label}</span>
					{#if !compact}
						<span class="text-gray-500 dark:text-gray-400">{m.description}</span>
					{/if}
				</button>
			{/each}
		</div>
	{/if}
</div>
