<!--
	Onboarding Screen 5: AI Analysis
	Consolidated page showing all 3 insights with carousel transitions
	Streams AI-generated insights in real-time
-->
<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { onboardingStore } from '$lib/stores/onboardingStore';
	import { onboardingAnalysis, analyzingInsights } from '$lib/stores/onboardingAnalysis';
	import { getSession } from '$lib/auth-client';
	import { fade } from 'svelte/transition';

	let phase = $state<'loading' | 'insights' | 'generating'>('loading');
	let currentInsightIndex = $state(0);
	let showInsight = $state(false);
	let error = $state<string | null>(null);

	// Get insights from store
	const insights = $derived([
		$analyzingInsights.message1,
		$analyzingInsights.message2,
		$analyzingInsights.message3
	]);

	const currentInsight = $derived(insights[currentInsightIndex]);
	const hasRealData = $derived($analyzingInsights.hasRealData);
	const gmailConnected = $derived($onboardingStore.userData.gmailConnected);

	// Delay helper
	function delay(ms: number): Promise<void> {
		return new Promise(resolve => setTimeout(resolve, ms));
	}

	// Cycle through insights with fade transitions
	async function cycleInsights() {
		phase = 'insights';

		for (let i = 0; i < 3; i++) {
			currentInsightIndex = i;
			showInsight = true;
			await delay(2500); // Show each insight for 2.5s
			showInsight = false;
			await delay(400); // Fade transition
		}

		// Show generating message
		phase = 'generating';
		await delay(1500);

		// Navigate to starter apps
		onboardingStore.nextStep();
		goto('/onboarding/starter-apps');
	}

	// Quick flow for skipped Gmail (no insights to show)
	async function quickFlow() {
		phase = 'loading';
		await delay(1500);

		phase = 'generating';
		await delay(1500);

		onboardingStore.nextStep();
		goto('/onboarding/starter-apps');
	}

	onMount(async () => {
		// Check if Gmail was connected
		if (!gmailConnected) {
			quickFlow();
			return;
		}

		// Timeout fallback - if analysis takes too long, use defaults
		const ANALYSIS_TIMEOUT = 10000; // 10 seconds max wait
		let hasStartedInsights = false;
		let timeoutId: ReturnType<typeof setTimeout>;

		const startFallbackInsights = () => {
			if (hasStartedInsights) return;
			hasStartedInsights = true;

			onboardingStore.setAnalysis({
				message1: 'No-code builder energy',
				message2: 'Design tools are your playground',
				message3: 'AI-curious, testing new platforms'
			});
			cycleInsights();
		};

		// Set timeout fallback
		timeoutId = setTimeout(() => {
			startFallbackInsights();
		}, ANALYSIS_TIMEOUT);

		// Subscribe to analysis state
		const unsubscribeAnalysis = onboardingAnalysis.subscribe(($analysis) => {
			error = $analysis.error;

			// Store additional data
			if ($analysis.interests.length > 0) {
				onboardingStore.setUserData({
					interests: $analysis.interests
				});
			}

			// When analysis completes (or fails but has fallback data)
			if ($analysis.status === 'completed' || $analysis.status === 'failed') {
				clearTimeout(timeoutId);

				// Update legacy onboarding store for backward compatibility
				onboardingStore.setAnalysis({
					message1: $analyzingInsights.message1,
					message2: $analyzingInsights.message2,
					message3: $analyzingInsights.message3
				});

				// Start cycling through insights
				if (!hasStartedInsights) {
					hasStartedInsights = true;
					cycleInsights();
				}
			}
		});

		// Get user session and start polling for real data
		try {
			const session = await getSession();
			if (session.data?.user?.id) {
				const userId = session.data.user.id;

				// Start polling for analysis status by user_id
				onboardingAnalysis.pollByUserId(userId);
			} else {
				clearTimeout(timeoutId);
				startFallbackInsights();
			}
		} catch (err) {
			clearTimeout(timeoutId);
			startFallbackInsights();
		}

		return () => {
			clearTimeout(timeoutId);
			unsubscribeAnalysis();
		};
	});
</script>

<svelte:head>
	<title>Analyzing - OSA Build</title>
</svelte:head>

<div class="onboarding-background">
	<div class="analyzing-screen">
		<div class="content">
			{#if phase === 'loading'}
				<div in:fade={{ duration: 300 }} out:fade={{ duration: 300 }}>
					<h1 class="title">
						Analyzing your workspace...
					</h1>

					<div class="spinner-wrapper">
						<div class="spinner"></div>
					</div>

					{#if $onboardingAnalysis.isStreaming}
						<p class="streaming-text">Reading your emails with AI...</p>
					{/if}

					{#if error}
						<p class="error-text">{error}</p>
						<p class="fallback-text">Using default insights</p>
					{/if}
				</div>
			{:else if phase === 'insights'}
				{#if showInsight}
					<div
						class="insight-container"
						in:fade={{ duration: 400 }}
						out:fade={{ duration: 400 }}
					>
						<h1 class="title insight-title">
							{currentInsight}
						</h1>

						{#if hasRealData}
							<p class="ai-badge">Based on your emails</p>
						{/if}

						<div class="dots">
							{#each [0, 1, 2] as i}
								<span
									class="dot"
									class:active={i === currentInsightIndex}
								></span>
							{/each}
						</div>
					</div>
				{/if}
			{:else if phase === 'generating'}
				<div in:fade={{ duration: 300 }}>
					<h1 class="title">
						Generating your apps...
					</h1>

					<div class="spinner-wrapper">
						<div class="spinner"></div>
					</div>
				</div>
			{/if}
		</div>
	</div>
</div>

<style>
	.onboarding-background {
		min-height: 100vh;
		width: 100%;
		background-image: url('/logos/integrations/MIOSABRANDBackround.png');
		background-size: cover;
		background-position: center;
		background-repeat: no-repeat;
	}

	.analyzing-screen {
		min-height: 100vh;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 2rem;
	}

	.content {
		width: 100%;
		max-width: 600px;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 3rem;
		text-align: center;
		min-height: 300px;
	}

	.title {
		font-size: 2.75rem;
		font-weight: 700;
		color: #1A1A1A;
		line-height: 1.2;
		letter-spacing: -0.02em;
		margin: 0;
	}

	.insight-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 2rem;
	}

	.insight-title {
		font-size: 2.5rem;
		max-width: 500px;
	}

	.spinner-wrapper {
		margin-top: 2rem;
	}

	.spinner {
		width: 64px;
		height: 64px;
		border: 3px solid #E5E5E5;
		border-top-color: #1A1A1A;
		border-radius: 50%;
		animation: spin 1s linear infinite;
	}

	.streaming-text {
		font-size: 0.875rem;
		color: #666;
		margin: 0;
	}

	.error-text {
		font-size: 1rem;
		color: #DC2626;
		margin: 0;
	}

	.fallback-text {
		font-size: 0.875rem;
		color: #999;
		margin: 0;
	}

	.ai-badge {
		font-size: 0.75rem;
		color: #666;
		background: #F5F5F5;
		padding: 0.375rem 0.75rem;
		border-radius: 1rem;
		margin: 0;
	}

	.dots {
		display: flex;
		gap: 0.5rem;
		margin-top: 1rem;
	}

	.dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		background: #D1D5DB;
		transition: all 0.3s ease;
	}

	.dot.active {
		background: #1A1A1A;
		transform: scale(1.25);
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}

	@media (max-width: 768px) {
		.title {
			font-size: 2rem;
		}

		.insight-title {
			font-size: 1.75rem;
		}

		.content {
			gap: 2.5rem;
		}
	}
</style>
