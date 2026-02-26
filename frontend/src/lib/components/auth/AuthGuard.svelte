<script lang="ts">
	/**
	 * AuthGuard wraps protected routes and enforces authentication.
	 *
	 * Behaviour by auth mode (fetched from /api/auth/mode):
	 *   single     — always renders children (owner session is auto-injected)
	 *   local      — redirects to /login when no session is present
	 *   oauth      — redirects to /login when no session is present
	 *   local+oauth — redirects to /login when no session is present
	 *
	 * The component is intentionally lightweight and relies on the existing
	 * useSession() store from auth-client.ts rather than re-implementing
	 * session logic.
	 */
	import { goto } from '$app/navigation';
	import { browser } from '$app/environment';
	import { useSession } from '$lib/auth-client';
	import { getBackendUrl } from '$lib/api/base';
	import { onMount } from 'svelte';

	let { children } = $props();

	const session = useSession();

	let authMode = $state<string | null>(null);
	let modeChecked = $state(false);

	onMount(async () => {
		if (!browser) return;

		try {
			const resp = await fetch(`${getBackendUrl()}/api/auth/mode`, { credentials: 'include' });
			if (resp.ok) {
				const data = await resp.json();
				authMode = data.mode ?? 'single';

				// First-boot redirect
				if (data.needs_setup) {
					goto('/setup');
					return;
				}
			}
		} catch {
			// On network error assume single mode (safest for desktop installs).
			authMode = 'single';
		} finally {
			modeChecked = true;
		}
	});

	// Redirect to /login when session is gone in a mode that requires it.
	$effect(() => {
		if (!modeChecked) return;
		if (authMode === 'single') return; // Never redirect in single mode.

		if (!$session.isPending && !$session.data) {
			goto('/login');
		}
	});
</script>

<!--
	Render children immediately in single mode (no login needed).
	In other modes render only after the session is confirmed to avoid flash.
-->
{#if authMode === 'single'}
	{@render children()}
{:else if modeChecked && $session.data}
	{@render children()}
{:else if !modeChecked || $session.isPending}
	<!-- Loading state — transparent, no flash -->
	<div class="h-screen w-screen flex items-center justify-center bg-white">
		<svg class="animate-spin h-6 w-6 text-gray-300" viewBox="0 0 24 24">
			<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" fill="none" />
			<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
		</svg>
	</div>
{/if}
