<script lang="ts">
	import { page } from '$app/stores';
	import { currentWorkspaceId, loadSavedWorkspace } from '$lib/stores/workspaces';
	import { userAppsStore } from '$lib/stores/userAppsStore';
	import { onMount } from 'svelte';
	import AppRegistryModal from '$lib/components/desktop/AppRegistryModal.svelte';

	// Check if we're in embed mode (inside iframe)
	const isEmbed = $derived($page.url.searchParams.get('embed') === 'true');

	// Workspace ID from store
	let workspaceId = $derived($currentWorkspaceId || '');

	onMount(() => {
		// Ensure workspace is loaded
		loadSavedWorkspace();
	});

	// Fetch user apps when workspace is available
	$effect(() => {
		if (workspaceId) {
			userAppsStore.fetch(workspaceId);
		}
	});
</script>

<svelte:head>
	<title>App Store - BusinessOS</title>
</svelte:head>

<div class="app-store-page" class:embed={isEmbed}>
	{#if workspaceId}
		<AppRegistryModal
			{workspaceId}
			isPage={true}
		/>
	{:else}
		<div class="loading-state">
			<p>Loading workspace...</p>
		</div>
	{/if}
</div>

<style>
	.app-store-page {
		height: 100%;
		width: 100%;
		display: flex;
		flex-direction: column;
		background: #f8fafc;
	}

	.app-store-page.embed {
		/* In embed mode, fill the iframe completely */
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
	}

	.loading-state {
		display: flex;
		align-items: center;
		justify-content: center;
		height: 100%;
		color: #6b7280;
	}
</style>
