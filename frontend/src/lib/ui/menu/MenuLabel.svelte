<script lang="ts">
	import { DropdownMenu as MenuPrimitive } from 'bits-ui';
	import { type Snippet } from 'svelte';
	import { cn } from '$lib/utils';

	interface Props {
		class?: string;
		children: Snippet;
	}

	let { class: className = '', children }: Props = $props();
</script>

<MenuPrimitive.Label class={cn('px-2 py-1.5 text-sm font-semibold', className)}>
	{@render children()}
</MenuPrimitive.Label>
