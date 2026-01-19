<script lang="ts">
	import '../app.css';
	import { onMount, onDestroy } from 'svelte';
	import { browser } from '$app/environment';
	import { page } from '$app/stores';
	import { themeStore } from '$lib/stores/themeStore';
	import { useSession } from '$lib/auth-client';
	import { streamingVoice, type VoiceState } from '$lib/services/streamingVoice';
	import VoiceOrbPanel from '$lib/components/desktop3d/VoiceOrbPanel.svelte';
	import LiveCaptions from '$lib/components/desktop3d/LiveCaptions.svelte';
	import { isOnboardingComplete } from '$lib/stores/onboardingStore';

	let { children } = $props();

	const session = useSession();

	// Derived: Show voice UI if user completed onboarding OR is on main app pages
	let showVoiceUI = $derived(
		$session.data && (
			$isOnboardingComplete ||
			$page.url.pathname === '/window' ||
			$page.url.pathname.startsWith('/(app)')
		)
	);

	// Voice state (only for authenticated users)
	let voiceState = $state<VoiceState>('disconnected');
	let isListening = $state(false);
	let isSpeaking = $state(false);
	let userMessage = $state('');
	let osaMessage = $state('');

	// Track if voice system has been initialized
	let voiceInitialized = false;

	// Initialize theme on mount
	onMount(() => {
		// Theme is already applied by the store on creation,
		// but we ensure it's set on the document
		if (browser) {
			const storedTheme = localStorage.getItem('theme');
			if (storedTheme === 'dark' || storedTheme === 'light' || storedTheme === 'system') {
				themeStore.setTheme(storedTheme);
			}
		}

		// Setup voice callbacks only once
		if (!voiceInitialized) {
			streamingVoice.setStateCallback((state: VoiceState) => {
				voiceState = state;
				isListening = state === 'connected' || state === 'transcribing' || state === 'speaking';
				isSpeaking = state === 'speaking';
			});

			streamingVoice.setUserCallback((text: string) => {
				userMessage = text;
				setTimeout(() => {
					if (userMessage === text) userMessage = '';
				}, 5000);
			});

			streamingVoice.setAgentCallback((text: string) => {
				osaMessage = text;
				setTimeout(() => {
					if (osaMessage === text) osaMessage = '';
				}, 8000);
			});

			streamingVoice.setErrorCallback((error: string) => {
				console.error('[Root Layout] Streaming voice error:', error);
				// Show error to user (optional)
			});

			voiceInitialized = true;
			console.log('[Root Layout] Streaming voice system initialized');
		}
	});

	// Cleanup
	onDestroy(() => {
		if (voiceState !== 'disconnected') {
			streamingVoice.disconnect();
		}
	});

	// Toggle voice
	async function toggleVoice() {
		if (voiceState === 'disconnected') {
			await streamingVoice.connect();
		} else {
			await streamingVoice.disconnect();
		}
	}
</script>

<svelte:head>
	<title>Business OS</title>
	<meta name="description" content="Your internal command center" />
</svelte:head>

<!-- Page content -->
{@render children()}

<!-- Voice Orb (for authenticated users on main app pages) -->
{#if showVoiceUI}
	<VoiceOrbPanel {isListening} {isSpeaking} onToggleListening={toggleVoice} />
	<LiveCaptions {userMessage} {osaMessage} {isListening} {isSpeaking} />
{/if}
