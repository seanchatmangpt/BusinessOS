<!--
	OsaPill.svelte
	Outer container for the OSA Interface.
	Manages expand/collapse and orchestrates sub-components.
	Protected module — no close/dismiss button.
-->
<script lang="ts">
	import { osaStore } from '$lib/stores/osa';
	import ModeIndicator from './ModeIndicator.svelte';
	import ModeSelector from './ModeSelector.svelte';
	import ChatInput from './ChatInput.svelte';
	import ResponseStream from './ResponseStream.svelte';

	interface Props {
		class?: string;
	}

	let { class: className = '' }: Props = $props();

	let isExpanded = $derived($osaStore.isExpanded);
	let isStreaming = $derived($osaStore.isStreaming);
	let error = $derived($osaStore.error);
	let chatInputRef: ChatInput | undefined = $state(undefined);

	export function focusInput() {
		osaStore.setExpanded(true);
		chatInputRef?.focus();
	}

	function handleCollapsedClick() {
		osaStore.setExpanded(true);
		// Delay focus to allow DOM to update after expansion
		requestAnimationFrame(() => chatInputRef?.focus());
	}

	function handleKeyDown(e: KeyboardEvent) {
		if (e.key === 'Escape' && isExpanded) {
			osaStore.setExpanded(false);
		}
	}

	function handleInputFocus() {
		if (!isExpanded) {
			osaStore.setExpanded(true);
		}
	}
</script>

<section
	class="osa-pill {className}"
	role="region"
	aria-label="OSA Interface"
	aria-expanded={isExpanded}
	onkeydown={handleKeyDown}
>
	{#if isExpanded}
		<!-- Expanded view: mode selector + conversation + input -->
		<div class="osa-pill-expanded flex flex-col gap-2 rounded-2xl border border-gray-200/60 bg-white/90 p-3 shadow-lg backdrop-blur-md dark:border-gray-700/60 dark:bg-gray-900/90">
			<!-- Header: mode selector -->
			<div class="flex items-center justify-between">
				<ModeSelector />
				<ModeIndicator />
			</div>

			<!-- Error banner -->
			{#if error}
				<div class="rounded-lg bg-red-50 px-3 py-1.5 text-xs text-red-600 dark:bg-red-900/20 dark:text-red-400" role="alert">
					{error}
				</div>
			{/if}

			<!-- Conversation -->
			<ResponseStream maxHeight="260px" />

			<!-- Input -->
			<ChatInput bind:this={chatInputRef} placeholder="Ask OSA... (Ctrl+Enter to send)" onfocus={handleInputFocus} />
		</div>
	{:else}
		<!-- Collapsed view: mode badge + compact input -->
		<!-- svelte-ignore a11y_click_events_have_key_events -->
		<div
			class="osa-pill-collapsed flex items-center gap-2 rounded-full border border-gray-200/60 bg-white/90 px-3 py-1.5 shadow-md backdrop-blur-md transition-all hover:shadow-lg dark:border-gray-700/60 dark:bg-gray-900/90"
			role="button"
			tabindex="0"
			onclick={handleCollapsedClick}
			aria-label="Expand OSA Interface"
		>
			<ModeIndicator compact />
			<span class="text-xs text-gray-400 dark:text-gray-500">Ask OSA...</span>
		</div>
	{/if}
</section>
