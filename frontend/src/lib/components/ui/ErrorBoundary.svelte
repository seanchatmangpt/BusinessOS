<script lang="ts">
	import { AlertCircle, RefreshCw } from 'lucide-svelte';

	interface Props {
		error?: Error | string | null;
		onRetry?: () => void;
		children?: import('svelte').Snippet;
	}

	let { error = null, onRetry, children }: Props = $props();

	const errorMessage = $derived(() => {
		if (!error) return '';
		if (typeof error === 'string') return error;
		return error.message || 'An unexpected error occurred';
	});
</script>

{#if error}
	<div class="flex items-center justify-center min-h-[400px] p-8">
		<div class="max-w-md w-full bg-red-50 dark:bg-red-950 border border-red-200 dark:border-red-800 rounded-lg p-6">
			<div class="flex items-start gap-4">
				<div class="flex-shrink-0">
					<AlertCircle class="h-6 w-6 text-red-600 dark:text-red-400" />
				</div>
				<div class="flex-1">
					<h3 class="text-lg font-semibold text-red-900 dark:text-red-100 mb-2">
						Something went wrong
					</h3>
					<p class="text-sm text-red-700 dark:text-red-300 mb-4">
						{errorMessage()}
					</p>
					{#if onRetry}
						<button
							onclick={onRetry}
							class="inline-flex items-center gap-2 px-4 py-2 bg-red-600 hover:bg-red-700 text-white text-sm font-medium rounded-md transition-colors"
						>
							<RefreshCw class="h-4 w-4" />
							Try Again
						</button>
					{/if}
				</div>
			</div>
		</div>
	</div>
{:else}
	{@render children?.()}
{/if}
