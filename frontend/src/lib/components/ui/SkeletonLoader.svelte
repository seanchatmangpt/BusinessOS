<script lang="ts">
	interface Props {
		variant?: 'text' | 'card' | 'avatar' | 'button';
		count?: number;
		class?: string;
	}

	let { variant = 'text', count = 1, class: className = '' }: Props = $props();

	const items = $derived(Array.from({ length: count }, (_, i) => i));
</script>

<div class="space-y-2 {className}">
	{#each items as item (item)}
		{#if variant === 'text'}
			<div class="h-4 bg-gray-200 dark:bg-gray-700 rounded animate-pulse"></div>
		{:else if variant === 'card'}
			<div class="border border-gray-200 dark:border-gray-700 rounded-lg p-4 space-y-3">
				<div class="h-4 bg-gray-200 dark:bg-gray-700 rounded animate-pulse w-3/4"></div>
				<div class="h-4 bg-gray-200 dark:bg-gray-700 rounded animate-pulse w-1/2"></div>
				<div class="h-4 bg-gray-200 dark:bg-gray-700 rounded animate-pulse w-5/6"></div>
			</div>
		{:else if variant === 'avatar'}
			<div class="h-10 w-10 bg-gray-200 dark:bg-gray-700 rounded-full animate-pulse"></div>
		{:else if variant === 'button'}
			<div class="h-10 bg-gray-200 dark:bg-gray-700 rounded-md animate-pulse w-24"></div>
		{/if}
	{/each}
</div>
