<script lang="ts">
	import { onMount } from 'svelte';
	import * as integrationsApi from '$lib/api/integrations';
	import type { IntegrationProviderInfo, UserIntegration } from '$lib/api/integrations';

	// Props
	interface Props {
		moduleId: string;
		title?: string;
		compact?: boolean;
	}

	let { moduleId, title = 'Integrations', compact = false }: Props = $props();

	// State
	let isLoading = $state(true);
	let availableProviders = $state<IntegrationProviderInfo[]>([]);
	let connectedIntegrations = $state<UserIntegration[]>([]);
	let isExpanded = $state(!compact);

	onMount(async () => {
		await loadIntegrations();
	});

	async function loadIntegrations() {
		isLoading = true;
		try {
			const response = await integrationsApi.getModuleIntegrations(moduleId);
			availableProviders = response.available_providers || [];
			connectedIntegrations = response.connected_integrations || [];
		} catch (err) {
			console.error('Failed to load module integrations:', err);
		} finally {
			isLoading = false;
		}
	}

	async function handleConnect(providerId: string) {
		try {
			const response = await integrationsApi.initiateAuth(
				providerId as integrationsApi.IntegrationProvider
			);
			if (response.auth_url) {
				window.location.href = response.auth_url;
			}
		} catch (err) {
			console.error('Failed to initiate auth:', err);
		}
	}

	async function triggerSync(integrationId: string) {
		try {
			await integrationsApi.triggerIntegrationSync(integrationId, moduleId);
		} catch (err) {
			console.error('Failed to trigger sync:', err);
		}
	}

	function getStatusIndicator(status: string) {
		switch (status) {
			case 'connected':
				return 'bg-green-500';
			case 'expired':
				return 'bg-yellow-500';
			case 'error':
				return 'bg-red-500';
			default:
				return 'bg-gray-400';
		}
	}
</script>

<div class="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
	<!-- Header -->
	<button
		onclick={() => (isExpanded = !isExpanded)}
		class="w-full flex items-center justify-between p-4 text-left hover:bg-gray-50 dark:hover:bg-gray-750 transition-colors"
	>
		<div class="flex items-center gap-3">
			<svg class="w-5 h-5 text-gray-500 dark:text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
			</svg>
			<span class="font-medium text-gray-900 dark:text-white">{title}</span>
			{#if connectedIntegrations.length > 0}
				<span class="px-2 py-0.5 text-xs bg-green-100 dark:bg-green-900/30 text-green-800 dark:text-green-400 rounded-full">
					{connectedIntegrations.length} connected
				</span>
			{/if}
		</div>
		<svg
			class="w-5 h-5 text-gray-400 transition-transform {isExpanded ? 'rotate-180' : ''}"
			fill="none"
			viewBox="0 0 24 24"
			stroke="currentColor"
		>
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
		</svg>
	</button>

	{#if isExpanded}
		<div class="border-t border-gray-200 dark:border-gray-700 p-4">
			{#if isLoading}
				<div class="flex items-center justify-center py-8">
					<div class="animate-spin rounded-full h-6 w-6 border-b-2 border-blue-600"></div>
				</div>
			{:else if connectedIntegrations.length === 0 && availableProviders.length === 0}
				<p class="text-sm text-gray-500 dark:text-gray-400 text-center py-4">
					No integrations available for this module.
				</p>
			{:else}
				<!-- Connected Integrations -->
				{#if connectedIntegrations.length > 0}
					<div class="mb-4">
						<h4 class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-2">
							Connected
						</h4>
						<div class="space-y-2">
							{#each connectedIntegrations as integration}
								<div class="flex items-center justify-between p-2 bg-gray-50 dark:bg-gray-700/50 rounded-lg">
									<div class="flex items-center gap-2">
										<div class="relative">
											<div class="w-8 h-8 rounded bg-gray-100 dark:bg-gray-600 flex items-center justify-center">
												{#if integration.icon_url}
													<img src={integration.icon_url} alt={integration.provider_name} class="w-5 h-5" />
												{:else}
													<span class="text-sm font-medium text-gray-600 dark:text-gray-300">
														{integration.provider_name.charAt(0)}
													</span>
												{/if}
											</div>
											<div class="absolute -bottom-0.5 -right-0.5 w-2.5 h-2.5 rounded-full border-2 border-white dark:border-gray-800 {getStatusIndicator(integration.status)}"></div>
										</div>
										<div>
											<p class="text-sm font-medium text-gray-900 dark:text-white">
												{integration.provider_name}
											</p>
											<p class="text-xs text-gray-500 dark:text-gray-400">
												{integration.external_account_name || integration.external_workspace_name || 'Connected'}
											</p>
										</div>
									</div>
									<div class="flex items-center gap-1">
										<button
											onclick={() => triggerSync(integration.id)}
											class="p-1.5 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-600 rounded transition-colors"
											title="Sync now"
										>
											<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
											</svg>
										</button>
										<a
											href="/integrations/{integration.id}"
											class="p-1.5 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-600 rounded transition-colors"
											title="Settings"
										>
											<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
											</svg>
										</a>
									</div>
								</div>
							{/each}
						</div>
					</div>
				{/if}

				<!-- Available Integrations -->
				{#if availableProviders.length > 0}
					<div>
						<h4 class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-2">
							Available
						</h4>
						<div class="grid grid-cols-2 sm:grid-cols-3 gap-2">
							{#each availableProviders as provider}
								{@const isConnected = connectedIntegrations.some(i => i.provider_id === provider.id)}
								{#if !isConnected}
									<button
										onclick={() => handleConnect(provider.id)}
										disabled={provider.status !== 'available'}
										class="flex items-center gap-2 p-2 bg-gray-50 dark:bg-gray-700/50 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed text-left"
									>
										<div class="w-8 h-8 rounded bg-gray-100 dark:bg-gray-600 flex items-center justify-center flex-shrink-0">
											{#if provider.icon_url}
												<img src={provider.icon_url} alt={provider.name} class="w-5 h-5" />
											{:else}
												<span class="text-sm font-medium text-gray-600 dark:text-gray-300">
													{provider.name.charAt(0)}
												</span>
											{/if}
										</div>
										<div class="min-w-0">
											<p class="text-sm font-medium text-gray-900 dark:text-white truncate">
												{provider.name}
											</p>
											{#if provider.status !== 'available'}
												<p class="text-xs text-gray-500 dark:text-gray-400">
													{provider.status.replace('_', ' ')}
												</p>
											{/if}
										</div>
									</button>
								{/if}
							{/each}
						</div>
					</div>
				{/if}

				<!-- Link to full integrations page -->
				<div class="mt-4 pt-4 border-t border-gray-200 dark:border-gray-700">
					<a
						href="/integrations"
						class="text-sm text-blue-600 dark:text-blue-400 hover:underline"
					>
						Manage all integrations
					</a>
				</div>
			{/if}
		</div>
	{/if}
</div>
