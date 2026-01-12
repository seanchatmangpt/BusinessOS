<!--
  Onboarding Page
  Conversational AI onboarding flow powered by Grok
  See ONBOARDING_ARCHITECTURE.md for full specification
-->
<script lang="ts">
	import { goto } from '$app/navigation';
	import { ConversationalOnboarding } from '$lib/components/onboarding';

	interface ExtractedData {
		workspaceName?: string;
		businessType?: string;
		teamSize?: string;
		role?: string;
		goals?: string[];
		challenges?: string[];
		integrations?: string[];
	}

	function handleComplete(data: ExtractedData) {
		// Store onboarding data
		localStorage.setItem('onboarding_data', JSON.stringify(data));
		localStorage.setItem('onboarding_completed', 'true');

		// TODO: Save to backend when API is ready
		// await fetch('/api/onboarding/complete', { ... });

		// Redirect to windows
		goto('/windows');
	}
</script>

<ConversationalOnboarding onComplete={handleComplete} />
