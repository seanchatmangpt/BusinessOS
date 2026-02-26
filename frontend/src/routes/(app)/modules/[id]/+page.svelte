<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { ArrowLeft, Download, Star, Share2, Upload, Loader2, Trash2, Settings } from 'lucide-svelte';
	import { customModulesStore } from '$lib/stores/customModulesStore';
	import ManifestViewer from '$lib/components/modules/ManifestViewer.svelte';
	import ShareDialog from '$lib/components/modules/ShareDialog.svelte';
	import { categoryColors, categoryLabels } from '$lib/types/modules';

	let store = $state(customModulesStore);
	let storeState = $state($store);

	$effect(() => {
		storeState = $store;
	});

	let moduleId = $derived($page.params.id);
	let activeTab = $state<'overview' | 'manifest' | 'versions' | 'settings'>('overview');
	let isShareDialogOpen = $state(false);
	let isInstalled = $state(false);
	let isProcessing = $state(false);

	onMount(async () => {
		await store.loadModule(moduleId);
		await store.loadVersions(moduleId);
	});

	async function handleInstall() {
		isProcessing = true;
		const success = await store.installModule(moduleId);
		if (success) {
			isInstalled = true;
		}
		isProcessing = false;
	}

	async function handleUninstall() {
		if (!confirm('Are you sure you want to uninstall this module?')) return;
		isProcessing = true;
		const success = await store.uninstallModule(moduleId);
		if (success) {
			isInstalled = false;
		}
		isProcessing = false;
	}

	async function handleExport() {
		const blob = await store.exportModule(moduleId);
		if (blob) {
			const url = URL.createObjectURL(blob);
			const a = document.createElement('a');
			a.href = url;
			a.download = `${storeState.currentModule?.name || 'module'}.json`;
			a.click();
			URL.revokeObjectURL(url);
		}
	}

	async function handleShare(data: Parameters<typeof store.shareModule>[1]) {
		await store.shareModule(moduleId, data);
	}

	async function handleDelete() {
		if (!confirm('Are you sure you want to delete this module? This action cannot be undone.')) return;
		const success = await store.deleteModule(moduleId);
		if (success) {
			goto('/modules');
		}
	}
</script>

<div class="h-full flex flex-col bg-white">
	{#if storeState.loading}
		<!-- Loading State -->
		<div class="flex items-center justify-center h-full">
			<div class="text-center">
				<Loader2 class="w-8 h-8 text-blue-600 animate-spin mx-auto mb-3" />
				<p class="text-sm text-gray-600">Loading module...</p>
			</div>
		</div>
	{:else if storeState.error || !storeState.currentModule}
		<!-- Error State -->
		<div class="flex items-center justify-center h-full">
			<div class="text-center max-w-md">
				<div class="w-16 h-16 bg-red-100 rounded-full flex items-center justify-center mx-auto mb-4">
					<svg class="w-8 h-8 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
					</svg>
				</div>
				<p class="text-sm font-medium text-gray-900 mb-2">Failed to load module</p>
				<p class="text-sm text-gray-600 mb-4">{storeState.error || 'Module not found'}</p>
				<button
					onclick={() => goto('/modules')}
					class="px-4 py-2 bg-gray-900 text-white rounded-lg hover:bg-gray-800 transition-colors text-sm"
				>
					Back to Modules
				</button>
			</div>
		</div>
	{:else}
		<!-- Module Header -->
		<div class="flex-shrink-0 border-b border-gray-200 bg-white px-8 py-6">
			<!-- Back Button -->
			<button
				onclick={() => goto('/modules')}
				class="flex items-center gap-2 text-sm text-gray-600 hover:text-gray-900 mb-4"
			>
				<ArrowLeft class="w-4 h-4" />
				<span>Back to Modules</span>
			</button>

			<!-- Module Info -->
			<div class="flex items-start justify-between">
				<div class="flex items-start gap-4">
					<!-- Icon -->
					{#if storeState.currentModule.icon}
						<div class="w-16 h-16 rounded-xl bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center text-white text-2xl font-bold">
							{storeState.currentModule.icon}
						</div>
					{:else}
						<div class="w-16 h-16 rounded-xl bg-gradient-to-br from-gray-400 to-gray-600 flex items-center justify-center text-white">
							<svg class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
							</svg>
						</div>
					{/if}

					<!-- Title and Meta -->
					<div>
						<div class="flex items-center gap-3 mb-2">
							<h1 class="text-2xl font-bold text-gray-900">{storeState.currentModule.name}</h1>
							<span class="text-xs px-2.5 py-1 rounded-full border {categoryColors[storeState.currentModule.category]}">
								{categoryLabels[storeState.currentModule.category]}
							</span>
						</div>
						<p class="text-sm text-gray-600 mb-3">{storeState.currentModule.description}</p>
						<div class="flex items-center gap-4 text-sm text-gray-500">
							<span class="flex items-center gap-1">
								<Download class="w-4 h-4" />
								{storeState.currentModule.install_count} installs
							</span>
							<span class="flex items-center gap-1">
								<Star class="w-4 h-4" />
								{storeState.currentModule.star_count} stars
							</span>
							<span>v{storeState.currentModule.version}</span>
						</div>
					</div>
				</div>

				<!-- Action Buttons -->
				<div class="flex items-center gap-2">
					{#if isInstalled}
						<button
							onclick={handleUninstall}
							disabled={isProcessing}
							class="px-4 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors text-sm font-medium disabled:opacity-50"
						>
							{isProcessing ? 'Uninstalling...' : 'Uninstall'}
						</button>
					{:else}
						<button
							onclick={handleInstall}
							disabled={isProcessing}
							class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium disabled:opacity-50"
						>
							{isProcessing ? 'Installing...' : 'Install'}
						</button>
					{/if}
					<button
						onclick={() => isShareDialogOpen = true}
						class="p-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
						title="Share"
					>
						<Share2 class="w-5 h-5 text-gray-700" />
					</button>
					<button
						onclick={handleExport}
						class="p-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
						title="Export"
					>
						<Upload class="w-5 h-5 text-gray-700" />
					</button>
					<button
						onclick={handleDelete}
						class="p-2 border border-red-300 text-red-600 rounded-lg hover:bg-red-50 transition-colors"
						title="Delete"
					>
						<Trash2 class="w-5 h-5" />
					</button>
				</div>
			</div>

			<!-- Tabs -->
			<div class="flex gap-1 mt-6 border-b border-gray-200">
				<button
					onclick={() => activeTab = 'overview'}
					class="px-4 py-2 text-sm font-medium transition-colors {activeTab === 'overview' ? 'text-blue-600 border-b-2 border-blue-600' : 'text-gray-600 hover:text-gray-900'}"
				>
					Overview
				</button>
				<button
					onclick={() => activeTab = 'manifest'}
					class="px-4 py-2 text-sm font-medium transition-colors {activeTab === 'manifest' ? 'text-blue-600 border-b-2 border-blue-600' : 'text-gray-600 hover:text-gray-900'}"
				>
					Manifest
				</button>
				<button
					onclick={() => activeTab = 'versions'}
					class="px-4 py-2 text-sm font-medium transition-colors {activeTab === 'versions' ? 'text-blue-600 border-b-2 border-blue-600' : 'text-gray-600 hover:text-gray-900'}"
				>
					Versions
				</button>
				<button
					onclick={() => activeTab = 'settings'}
					class="px-4 py-2 text-sm font-medium transition-colors {activeTab === 'settings' ? 'text-blue-600 border-b-2 border-blue-600' : 'text-gray-600 hover:text-gray-900'}"
				>
					Settings
				</button>
			</div>
		</div>

		<!-- Tab Content -->
		<div class="flex-1 overflow-y-auto px-8 py-6">
			{#if activeTab === 'overview'}
				<!-- Overview Tab -->
				<div class="max-w-4xl space-y-6">
					<!-- Actions -->
					<div>
						<h2 class="text-lg font-semibold text-gray-900 mb-4">Actions</h2>
						{#if storeState.currentModule.manifest.actions.length === 0}
							<p class="text-sm text-gray-600">No actions defined for this module.</p>
						{:else}
							<div class="space-y-3">
								{#each storeState.currentModule.manifest.actions as action}
									<div class="p-4 border border-gray-200 rounded-lg">
										<div class="flex items-center justify-between mb-2">
											<h3 class="font-medium text-gray-900">{action.name}</h3>
											<span class="text-xs px-2 py-1 bg-gray-100 text-gray-700 rounded">
												{action.type}
											</span>
										</div>
										<p class="text-sm text-gray-600">{action.description}</p>
									</div>
								{/each}
							</div>
						{/if}
					</div>

					<!-- Author -->
					{#if storeState.currentModule.creator_name}
						<div>
							<h2 class="text-lg font-semibold text-gray-900 mb-4">Author</h2>
							<p class="text-sm text-gray-600">{storeState.currentModule.creator_name}</p>
						</div>
					{/if}
				</div>
			{:else if activeTab === 'manifest'}
				<!-- Manifest Tab -->
				<div class="max-w-4xl">
					<h2 class="text-lg font-semibold text-gray-900 mb-4">Module Manifest</h2>
					<ManifestViewer manifest={storeState.currentModule.manifest} />
				</div>
			{:else if activeTab === 'versions'}
				<!-- Versions Tab -->
				<div class="max-w-4xl">
					<h2 class="text-lg font-semibold text-gray-900 mb-4">Version History</h2>
					{#if storeState.versions.length === 0}
						<p class="text-sm text-gray-600">No version history available.</p>
					{:else}
						<div class="space-y-3">
							{#each storeState.versions as version}
								<div class="p-4 border border-gray-200 rounded-lg">
									<div class="flex items-center justify-between mb-2">
										<h3 class="font-medium text-gray-900">v{version.version}</h3>
										<span class="text-xs text-gray-500">
											{new Date(version.created_at).toLocaleDateString()}
										</span>
									</div>
									{#if version.changelog}
										<p class="text-sm text-gray-600">{version.changelog}</p>
									{/if}
								</div>
							{/each}
						</div>
					{/if}
				</div>
			{:else if activeTab === 'settings'}
				<!-- Settings Tab -->
				<div class="max-w-4xl">
					<h2 class="text-lg font-semibold text-gray-900 mb-4">Module Settings</h2>
					<div class="space-y-4">
						<!-- Visibility -->
						<div class="p-4 border border-gray-200 rounded-lg">
							<label class="flex items-center justify-between">
								<div>
									<p class="font-medium text-gray-900">Active</p>
									<p class="text-sm text-gray-600">Enable or disable this module</p>
								</div>
								<input
									type="checkbox"
									checked={storeState.currentModule.is_active}
									class="w-5 h-5"
								/>
							</label>
						</div>

						<!-- More settings can be added here -->
					</div>
				</div>
			{/if}
		</div>

		<!-- Share Dialog -->
		<ShareDialog
			{moduleId}
			moduleName={storeState.currentModule.name}
			isOpen={isShareDialogOpen}
			onClose={() => isShareDialogOpen = false}
			onShare={handleShare}
		/>
	{/if}
</div>
