<!--
	ChatInput.svelte
	Textarea with send button for OSA messages.
	Handles submission lifecycle and auto-resize.
-->
<script lang="ts">
	import { osaStore } from '$lib/stores/osa';

	interface Props {
		compact?: boolean;
		placeholder?: string;
		onfocus?: () => void;
	}

	let { compact = false, placeholder, onfocus }: Props = $props();

	let inputValue = $state('');
	let inputElement: HTMLTextAreaElement | undefined = $state(undefined);

	let isStreaming = $derived($osaStore.isStreaming);

	export function focus() {
		inputElement?.focus();
	}

	async function handleSend() {
		const trimmed = inputValue.trim();
		if (!trimmed || isStreaming) return;

		inputValue = '';
		resetHeight();
		await osaStore.sendMessage(trimmed);
		inputElement?.focus();
	}

	function handleKeyDown(e: KeyboardEvent) {
		if (compact) {
			if (e.key === 'Enter' && !e.shiftKey) {
				e.preventDefault();
				handleSend();
			}
		} else {
			if (e.key === 'Enter' && e.ctrlKey) {
				e.preventDefault();
				handleSend();
			}
		}
	}

	function handleInput() {
		if (!inputElement) return;
		inputElement.style.height = 'auto';
		const max = compact ? 38 : 120;
		inputElement.style.height = `${Math.min(inputElement.scrollHeight, max)}px`;
	}

	function resetHeight() {
		if (!inputElement) return;
		inputElement.style.height = 'auto';
	}
</script>

<div class="osa-chat-input flex items-end gap-2">
	<textarea
		bind:this={inputElement}
		bind:value={inputValue}
		{placeholder}
		aria-label="Message OSA"
		aria-multiline={!compact}
		aria-busy={isStreaming}
		disabled={isStreaming}
		rows={1}
		class="flex-1 resize-none rounded-lg border border-gray-200 bg-white/80 px-3 py-2 text-sm outline-none transition-colors placeholder:text-gray-400 focus:border-gray-400 disabled:opacity-50 dark:border-gray-700 dark:bg-gray-800/80 dark:placeholder:text-gray-500 dark:focus:border-gray-500"
		class:max-h-[38px]={compact}
		class:max-h-[120px]={!compact}
		onkeydown={handleKeyDown}
		oninput={handleInput}
		onfocus={onfocus}
	></textarea>

	<button
		class="send-btn flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-gray-800 text-white transition-opacity hover:bg-gray-700 disabled:opacity-30 dark:bg-gray-200 dark:text-gray-900 dark:hover:bg-gray-300"
		disabled={!inputValue.trim() || isStreaming}
		onclick={handleSend}
		aria-label={isStreaming ? 'Sending...' : 'Send message'}
	>
		{#if isStreaming}
			<svg class="h-4 w-4 animate-spin" viewBox="0 0 24 24" fill="none">
				<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
				<path
					class="opacity-75"
					fill="currentColor"
					d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
				/>
			</svg>
		{:else}
			<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="h-4 w-4">
				<line x1="22" y1="2" x2="11" y2="13" />
				<polygon points="22 2 15 22 11 13 2 9 22 2" />
			</svg>
		{/if}
	</button>
</div>
