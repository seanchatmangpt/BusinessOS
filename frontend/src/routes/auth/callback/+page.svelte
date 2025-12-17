<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { browser } from '$app/environment';

	onMount(async () => {
		if (!browser) return;

		// The OAuth flow completed and set a session cookie
		// Just redirect to the main app
		// Small delay to ensure cookie is set
		setTimeout(() => {
			goto('/window');
		}, 100);
	});
</script>

<div class="min-h-screen flex items-center justify-center bg-white">
	<div class="text-center">
		<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-black mx-auto mb-4"></div>
		<p class="text-gray-500 font-mono text-sm">Completing sign in...</p>
	</div>
</div>
