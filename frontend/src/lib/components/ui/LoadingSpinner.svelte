<script lang="ts">
	import { Loader2 } from 'lucide-svelte';

	interface Props {
		size?: 'sm' | 'md' | 'lg';
		message?: string;
		fullscreen?: boolean;
	}

	let { size = 'md', message, fullscreen = false }: Props = $props();

	const sizeClasses = {
		sm: 'h-4 w-4',
		md: 'h-8 w-8',
		lg: 'h-12 w-12'
	};
</script>

<div class={fullscreen ? 'flex items-center justify-center min-h-[400px]' : 'inline-flex items-center gap-2'}>
	<Loader2 class="{sizeClasses[size]} animate-spin text-primary" />
	{#if message}
		<span class="text-sm text-muted-foreground">{message}</span>
	{/if}
</div>
