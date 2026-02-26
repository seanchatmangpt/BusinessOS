<script lang="ts">
	let visible = $state(false);

	if (typeof window !== 'undefined') {
		visible = !localStorage.getItem('cookie_consent');
	}

	function accept() {
		localStorage.setItem('cookie_consent', 'accepted');
		localStorage.setItem('cookie_consent_date', new Date().toISOString());
		visible = false;
	}

	function decline() {
		localStorage.setItem('cookie_consent', 'essential_only');
		localStorage.setItem('cookie_consent_date', new Date().toISOString());
		visible = false;
	}
</script>

{#if visible}
<div class="fixed bottom-0 left-0 right-0 z-50 bg-white dark:bg-gray-900 border-t border-gray-200 dark:border-gray-700 shadow-lg p-4">
	<div class="max-w-4xl mx-auto flex flex-col sm:flex-row items-center justify-between gap-4">
		<p class="text-sm text-gray-600 dark:text-gray-300">
			We use essential cookies for authentication and security. See our
			<a href="/privacy" class="text-blue-600 dark:text-blue-400 underline hover:no-underline">Privacy Policy</a>
			for details.
		</p>
		<div class="flex gap-3 shrink-0">
			<button
				onclick={decline}
				class="px-4 py-2 text-sm text-gray-600 dark:text-gray-400 border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors"
			>
				Essential Only
			</button>
			<button
				onclick={accept}
				class="px-4 py-2 text-sm text-white bg-blue-600 rounded-lg hover:bg-blue-700 transition-colors"
			>
				Accept All
			</button>
		</div>
	</div>
</div>
{/if}
