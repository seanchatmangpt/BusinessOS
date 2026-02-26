<!--
	ResponseStream.svelte
	Displays the OSA conversation history with streaming support.
	User messages as plain text, OSA responses with mode badges and simple markdown.
-->
<script lang="ts">
	import { osaStore } from '$lib/stores/osa';
	import ModeIndicator from './ModeIndicator.svelte';

	interface Props {
		maxHeight?: string;
	}

	let { maxHeight = '300px' }: Props = $props();

	let conversation = $derived($osaStore.conversation);
	let isStreaming = $derived($osaStore.isStreaming);
	let streamingContent = $derived($osaStore.streamingContent);
	let scrollContainer: HTMLDivElement | undefined = $state(undefined);

	// Auto-scroll to bottom on new content
	$effect(() => {
		// Track dependencies
		conversation.length;
		streamingContent;
		if (scrollContainer) {
			scrollContainer.scrollTop = scrollContainer.scrollHeight;
		}
	});

	/** Simple inline markdown: bold, inline code, line breaks */
	function renderSimpleMarkdown(text: string): string {
		return text
			.replace(/&/g, '&amp;')
			.replace(/</g, '&lt;')
			.replace(/>/g, '&gt;')
			.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
			.replace(/`([^`]+)`/g, '<code class="rounded bg-gray-100 px-1 py-0.5 text-xs dark:bg-gray-700">$1</code>')
			.replace(/\n/g, '<br />');
	}
</script>

<div
	bind:this={scrollContainer}
	class="osa-response-stream overflow-y-auto"
	style:max-height={maxHeight}
	role="log"
	aria-label="OSA conversation"
	aria-live="polite"
	aria-relevant="additions"
>
	{#each conversation as message (message.id)}
		{#if message.role === 'user'}
			<div class="user-message mb-3 flex justify-end" role="listitem">
				<div class="max-w-[80%] rounded-lg bg-gray-800 px-3 py-2 text-sm text-white dark:bg-gray-200 dark:text-gray-900">
					{message.content}
				</div>
			</div>
		{:else}
			<div class="osa-message mb-3" role="listitem">
				<div class="mb-1">
					<ModeIndicator mode={message.mode} confidence={message.confidence} compact />
				</div>
				<div class="rounded-lg bg-gray-50 px-3 py-2 text-sm text-gray-800 dark:bg-gray-800/50 dark:text-gray-200">
					{@html renderSimpleMarkdown(message.content)}
				</div>
			</div>
		{/if}
	{/each}

	{#if isStreaming}
		<div class="osa-message streaming mb-3" aria-live="polite" aria-atomic="false">
			<div class="rounded-lg bg-gray-50 px-3 py-2 text-sm text-gray-800 dark:bg-gray-800/50 dark:text-gray-200">
				{#if streamingContent}
					{@html renderSimpleMarkdown(streamingContent)}
				{/if}
				<span class="streaming-cursor inline-block animate-pulse text-gray-400" aria-hidden="true">|</span>
			</div>
		</div>
	{/if}

	{#if conversation.length === 0 && !isStreaming}
		<div class="flex items-center justify-center py-6 text-xs text-gray-400 dark:text-gray-500">
			Ask OSA anything...
		</div>
	{/if}
</div>
