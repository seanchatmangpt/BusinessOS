<!--
	ModeIndicator.svelte
	Badge showing the current active OSA mode and confidence score.
	Used in OsaPill header and on each OSA message in ResponseStream.
-->
<script lang="ts">
	import { osaStore, type OsaMode } from '$lib/stores/osa';

	interface Props {
		mode?: OsaMode;
		confidence?: number;
		compact?: boolean;
	}

	let { mode, confidence, compact = false }: Props = $props();

	let activeMode = $derived(mode ?? $osaStore.activeMode);
	let activeConfidence = $derived(confidence ?? $osaStore.modeConfidence);
	let confidencePercent = $derived(Math.round(activeConfidence * 100));

	const MODE_COLORS: Record<OsaMode, string> = {
		BUILD: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300',
		ASSIST: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300',
		ANALYZE: 'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-300',
		EXECUTE: 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300',
		MAINTAIN: 'bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300'
	};

	let colorClass = $derived(MODE_COLORS[activeMode] ?? MODE_COLORS.ASSIST);
	let label = $derived(
		compact || confidencePercent === 0
			? activeMode
			: `${activeMode} ${confidencePercent}%`
	);
	let ariaLabel = $derived(`${activeMode} mode, ${confidencePercent}% confidence`);
</script>

<span
	class="mode-badge inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-semibold leading-tight {colorClass}"
	role="status"
	aria-label={ariaLabel}
>
	{label}
</span>
