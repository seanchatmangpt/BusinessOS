<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { browser } from '$app/environment';
	import { getSession } from '$lib/auth-client';

	onMount(async () => {
		if (!browser) return;

		// The OAuth flow completed and set a session cookie
		// Check if this is a new user or if onboarding is incomplete

		// First check for new_user cookie (set by backend during Google OAuth signup)
		const newUserCookie = document.cookie.split('; ').find(row => row.startsWith('new_user='));
		const isNewUser = newUserCookie?.split('=')[1] === 'true';

		// Clear the cookie
		if (newUserCookie) {
			document.cookie = 'new_user=; path=/; max-age=0';
		}

		// If new user, go to onboarding
		if (isNewUser) {
			goto('/onboarding');
			return;
		}

		// For existing users, check onboarding status from session
		const { data } = await getSession();
		if (data?.user?.onboardingCompleted === false) {
			// Existing user who hasn't completed onboarding
			goto('/onboarding');
			return;
		}

		// User has completed onboarding, go to main app
		goto('/window');
	});
</script>

<div class="min-h-screen flex items-center justify-center bg-white">
	<div class="text-center">
		<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-black mx-auto mb-4"></div>
		<p class="text-gray-500 font-mono text-sm">Completing sign in...</p>
	</div>
</div>
