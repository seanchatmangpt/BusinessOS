<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { useSession } from '$lib/auth-client';
	import * as integrationsApi from '$lib/api/integrations';
	import type { UserIntegration, IntegrationSettings } from '$lib/api/integrations';

	const session = useSession();

	// State
	let isLoading = $state(true);
	let isSaving = $state(false);
	let integration = $state<UserIntegration | null>(null);
	let settings = $state<IntegrationSettings>({
		enabledSkills: [],
		notifications: true,
		syncSettings: {}
	});
	let error = $state<string | null>(null);
	let successMessage = $state<string | null>(null);

	// Get integration ID from route params
	let integrationId = $derived($page.params.id ?? '');

	onMount(async () => {
		if (!$session?.data?.user) {
			goto('/login');
			return;
		}

		if (integrationId) {
			await loadIntegration();
		}
	});

	async function loadIntegration() {
		if (!integrationId) return;
		isLoading = true;
		error = null;
		try {
			const response = await integrationsApi.getUserIntegration(integrationId);
			integration = response.integration;
			settings = { ...response.integration.settings };
		} catch (err) {
			console.error('Failed to load integration:', err);
			error = 'Failed to load integration details';
		} finally {
			isLoading = false;
		}
	}

	async function saveSettings() {
		if (!integrationId) return;
		isSaving = true;
		error = null;
		successMessage = null;
		try {
			await integrationsApi.updateIntegrationSettings(integrationId, settings);
			successMessage = 'Settings saved successfully';
			setTimeout(() => (successMessage = null), 3000);
		} catch (err) {
			console.error('Failed to save settings:', err);
			error = 'Failed to save settings';
		} finally {
			isSaving = false;
		}
	}

	async function handleDisconnect() {
		if (!integrationId) return;
		if (!confirm('Are you sure you want to disconnect this integration?')) {
			return;
		}

		try {
			await integrationsApi.disconnectUserIntegration(integrationId);
			goto('/integrations');
		} catch (err) {
			console.error('Failed to disconnect:', err);
			error = 'Failed to disconnect integration';
		}
	}

	async function triggerSync() {
		if (!integrationId) return;
		try {
			const response = await integrationsApi.triggerIntegrationSync(integrationId);
			successMessage = response.message || 'Sync started';
			setTimeout(() => (successMessage = null), 3000);
		} catch (err) {
			console.error('Failed to trigger sync:', err);
			error = 'Failed to trigger sync';
		}
	}

	function toggleSkill(skillId: string) {
		if (settings.enabledSkills.includes(skillId)) {
			settings.enabledSkills = settings.enabledSkills.filter((s) => s !== skillId);
		} else {
			settings.enabledSkills = [...settings.enabledSkills, skillId];
		}
	}

	function getStatusBadgeClass(status: string) {
		switch (status) {
			case 'connected':
				return 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400';
			case 'expired':
				return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400';
			case 'error':
				return 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400';
			default:
				return 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300';
		}
	}
</script>

<svelte:head>
	<title>{integration?.provider_name || 'Integration'} Settings | BusinessOS</title>
</svelte:head>

<div class="min-h-screen bg-gray-50 dark:bg-gray-900">
	<!-- Header -->
	<div class="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">
		<div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
			<div class="flex items-center gap-4">
				<a
					href="/integrations"
					aria-label="Back to integrations"
					class="p-2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
				>
					<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M15 19l-7-7 7-7"
						/>
					</svg>
				</a>
				{#if integration}
					<div class="flex items-center gap-3">
						<div
							class="w-12 h-12 rounded-lg bg-gray-100 dark:bg-gray-700 flex items-center justify-center"
						>
							{#if integration.icon_url}
								<img src={integration.icon_url} alt={integration.provider_name} class="w-8 h-8" />
							{:else}
								<span class="text-xl font-medium text-gray-600 dark:text-gray-300">
									{integration.provider_name.charAt(0)}
								</span>
							{/if}
						</div>
						<div>
							<div class="flex items-center gap-2">
								<h1 class="text-xl font-bold text-gray-900 dark:text-white">
									{integration.provider_name}
								</h1>
								<span
									class="px-2 py-0.5 text-xs rounded-full {getStatusBadgeClass(integration.status)}"
								>
									{integration.status}
								</span>
							</div>
							<p class="text-sm text-gray-500 dark:text-gray-400">
								{integration.external_account_name ||
									integration.external_workspace_name ||
									integration.category}
							</p>
						</div>
					</div>
				{/if}
			</div>
		</div>
	</div>

	<!-- Content -->
	<div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
		{#if isLoading}
			<div class="flex items-center justify-center h-64">
				<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
			</div>
		{:else if !integration}
			<div class="text-center py-12">
				<svg
					class="mx-auto h-12 w-12 text-gray-400"
					fill="none"
					viewBox="0 0 24 24"
					stroke="currentColor"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M12 12h.01M12 14h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
					/>
				</svg>
				<h3 class="mt-4 text-lg font-medium text-gray-900 dark:text-white">
					Integration not found
				</h3>
				<p class="mt-2 text-gray-500 dark:text-gray-400">
					This integration doesn't exist or you don't have access to it.
				</p>
				<a
					href="/integrations"
					class="mt-4 inline-block px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
				>
					Back to Integrations
				</a>
			</div>
		{:else}
			<!-- Messages -->
			{#if error}
				<div class="mb-6 p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg">
					<p class="text-red-800 dark:text-red-400">{error}</p>
				</div>
			{/if}
			{#if successMessage}
				<div class="mb-6 p-4 bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-lg">
					<p class="text-green-800 dark:text-green-400">{successMessage}</p>
				</div>
			{/if}

			<div class="space-y-6">
				<!-- Connection Info -->
				<div class="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-6">
					<h2 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
						Connection Details
					</h2>
					<dl class="grid grid-cols-1 sm:grid-cols-2 gap-4">
						<div>
							<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Account</dt>
							<dd class="mt-1 text-sm text-gray-900 dark:text-white">
								{integration.external_account_name || 'N/A'}
							</dd>
						</div>
						{#if integration.external_workspace_name}
							<div>
								<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Workspace</dt>
								<dd class="mt-1 text-sm text-gray-900 dark:text-white">
									{integration.external_workspace_name}
								</dd>
							</div>
						{/if}
						<div>
							<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Connected</dt>
							<dd class="mt-1 text-sm text-gray-900 dark:text-white">
								{new Date(integration.connected_at).toLocaleDateString()}
							</dd>
						</div>
						{#if integration.last_used_at}
							<div>
								<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Last Used</dt>
								<dd class="mt-1 text-sm text-gray-900 dark:text-white">
									{new Date(integration.last_used_at).toLocaleDateString()}
								</dd>
							</div>
						{/if}
						<div>
							<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Category</dt>
							<dd class="mt-1 text-sm text-gray-900 dark:text-white capitalize">
								{integration.category}
							</dd>
						</div>
						{#if integration.scopes?.length > 0}
							<div class="sm:col-span-2">
								<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Permissions</dt>
								<dd class="mt-1 flex flex-wrap gap-1">
									{#each integration.scopes as scope}
										<span class="px-2 py-0.5 text-xs bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-300 rounded">
											{scope}
										</span>
									{/each}
								</dd>
							</div>
						{/if}
					</dl>
					<div class="mt-4 pt-4 border-t border-gray-200 dark:border-gray-700">
						<button
							onclick={triggerSync}
							class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
						>
							Sync Now
						</button>
					</div>
				</div>

				<!-- Skills -->
				{#if integration.skills?.length > 0}
					<div class="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-6">
						<h2 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
							Available Skills
						</h2>
						<p class="text-sm text-gray-500 dark:text-gray-400 mb-4">
							Enable or disable specific AI skills for this integration.
						</p>
						<div class="space-y-2">
							{#each integration.skills as skill}
								<label class="flex items-center gap-3 p-3 bg-gray-50 dark:bg-gray-700/50 rounded-lg cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors">
									<input
										type="checkbox"
										checked={settings.enabledSkills.includes(skill)}
										onchange={() => toggleSkill(skill)}
										class="rounded border-gray-300 dark:border-gray-600 text-blue-600 focus:ring-blue-500"
									/>
									<span class="text-sm text-gray-900 dark:text-white">{skill}</span>
								</label>
							{/each}
						</div>
					</div>
				{/if}

				<!-- Modules -->
				{#if integration.modules?.length > 0}
					<div class="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-6">
						<h2 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
							Available In Modules
						</h2>
						<div class="flex flex-wrap gap-2">
							{#each integration.modules as mod}
								<a
									href="/{mod.toLowerCase()}"
									class="px-3 py-1.5 bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors text-sm"
								>
									{mod}
								</a>
							{/each}
						</div>
					</div>
				{/if}

				<!-- Notifications -->
				<div class="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-6">
					<h2 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
						Notifications
					</h2>
					<label class="flex items-center gap-3">
						<input
							type="checkbox"
							bind:checked={settings.notifications}
							class="rounded border-gray-300 dark:border-gray-600 text-blue-600 focus:ring-blue-500"
						/>
						<span class="text-sm text-gray-900 dark:text-white">
							Enable notifications from this integration
						</span>
					</label>
				</div>

				<!-- Actions -->
				<div class="flex items-center justify-between pt-4">
					<button
						onclick={handleDisconnect}
						class="px-4 py-2 text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg transition-colors"
					>
						Disconnect Integration
					</button>
					<button
						onclick={saveSettings}
						disabled={isSaving}
						class="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
					>
						{isSaving ? 'Saving...' : 'Save Settings'}
					</button>
				</div>
			</div>
		{/if}
	</div>
</div>
