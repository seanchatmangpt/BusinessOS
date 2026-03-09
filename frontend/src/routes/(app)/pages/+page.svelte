<script lang="ts">
	/**
	 * Pages - BusinessOS Document System
	 * Document listing with sidebar, editor, graph views.
	 * Composes knowledge-base module components with Foundation kb- patterns.
	 */
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import {
		// Stores
		documentsStore,
		activeDocumentStore,
		sidebarStore,
		documentMetas,
		// Services
		fetchDocuments,
		openAndFetchDocument,
		createDocument,
		deleteDocument,
		// Components
		KBSidebar,
		QuickSearch,
		DocumentEditor,
		GraphView
	} from '$lib/modules/knowledge-base';
	import type { DocumentMeta } from '$lib/modules/knowledge-base';

	// State
	let isLoading = $state(true);
	let error = $state<string | null>(null);
	let showQuickSearch = $state(false);
	let showNewDocForm = $state(false);
	let newDocTitle = $state('');

	// Derived
	let currentDocumentId = $derived($activeDocumentStore.id);
	let currentView = $derived($sidebarStore.view);
	let documents = $derived($documentMetas);

	// LocalStorage for last opened doc
	const LAST_DOC_KEY = 'bos-pages-last-document';

	$effect(() => {
		if (currentDocumentId) {
			localStorage.setItem(LAST_DOC_KEY, currentDocumentId);
		}
	});

	// Initialize
	onMount(async () => {
		try {
			await fetchDocuments();

			if (!$activeDocumentStore.id) {
				const lastDocId = localStorage.getItem(LAST_DOC_KEY);
				if (lastDocId) {
					const docExists = $documentMetas.some(d => d.id === lastDocId);
					if (docExists) {
						await openAndFetchDocument(lastDocId);
					} else {
						const mostRecent = $documentMetas[0];
						if (mostRecent) {
							await openAndFetchDocument(mostRecent.id);
						}
					}
				} else if ($documentMetas.length > 0) {
					await openAndFetchDocument($documentMetas[0].id);
				}
			}

			isLoading = false;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load documents';
			isLoading = false;
		}
	});

	// Handlers
	async function handleNewDocument() {
		try {
			const doc = await createDocument({
				title: '',
				type: 'document'
			});
			await openAndFetchDocument(doc.id);
		} catch (e) {
			console.error('Failed to create document:', e);
			error = 'Failed to create document';
		}
	}

	async function handleOpenDocument(id: string) {
		try {
			await openAndFetchDocument(id);
		} catch (e) {
			console.error('Failed to open document:', e);
			error = 'Failed to open document';
		}
	}

	function handleCloseDocument() {
		activeDocumentStore.setActiveDocument(null);
	}

	function handleOpenSearch() {
		showQuickSearch = true;
	}

	async function handleDeleteDocument(id: string) {
		try {
			await deleteDocument(id);
			if (currentDocumentId === id) {
				activeDocumentStore.setActiveDocument(null);
			}
		} catch (e) {
			console.error('Failed to delete document:', e);
			error = 'Failed to delete document';
		}
	}

	function formatDate(dateStr: string): string {
		const date = new Date(dateStr);
		const now = new Date();
		const diff = now.getTime() - date.getTime();
		const days = Math.floor(diff / (1000 * 60 * 60 * 24));
		if (days === 0) return 'Today';
		if (days === 1) return 'Yesterday';
		if (days < 7) return `${days}d ago`;
		return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
	}

	function getDocIcon(doc: DocumentMeta): string {
		if (doc.icon && typeof doc.icon === 'string') return doc.icon;
		if (doc.icon && typeof doc.icon === 'object' && 'value' in doc.icon) return doc.icon.value;
		return '';
	}
</script>

<svelte:head>
	<title>Pages | BusinessOS</title>
</svelte:head>

<div class="kb-page">
	<!-- Sidebar -->
	<KBSidebar
		onNewDocument={handleNewDocument}
		onOpenDocument={handleOpenDocument}
		onOpenSearch={handleOpenSearch}
	/>

	<!-- Main Content -->
	<main class="kb-page__main">
		{#if isLoading}
			<div class="kb-page__center">
				<div class="kb-page__spinner"></div>
				<p class="kb-page__center-text">Loading documents...</p>
			</div>
		{:else if error}
			<div class="kb-page__center">
				<p class="kb-page__center-text">{error}</p>
				<button
					class="btn-compact btn-compact-secondary"
					aria-label="Retry loading documents"
					onclick={() => { error = null; isLoading = true; fetchDocuments().finally(() => { isLoading = false; }) }}
				>Retry</button>
			</div>
		{:else if currentView === 'graph'}
			<GraphView
				documents={$documentMetas}
				selectedId={null}
				onSelect={(doc) => handleOpenDocument(doc.id)}
				onNavigate={(doc) => handleOpenDocument(doc.id)}
			/>
		{:else if currentDocumentId}
			<DocumentEditor
				documentId={currentDocumentId}
				onClose={handleCloseDocument}
			/>
		{:else}
			<!-- Document listing / empty state -->
			<div class="kb-page__listing">
				<!-- Header -->
				<div class="kb-page__header">
					<h1 class="kb-page__title">Pages</h1>
					<div class="kb-page__actions">
						<button
						class="btn-compact btn-compact-ghost"
							aria-label="Search documents"
							onclick={handleOpenSearch}
						>
							<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.3-4.3"/></svg>
							Search
						</button>
						<button
						class="btn-compact btn-compact-primary"
							aria-label="Create new page"
							onclick={handleNewDocument}
						>
							<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 5v14"/><path d="M5 12h14"/></svg>
							New Page
						</button>
					</div>
				</div>

				{#if documents.length === 0}
					<!-- Empty state -->
					<div class="kb-page__empty">
						<div class="kb-page__empty-icon">
							<svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="url(#emptyGrad)" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
								<defs>
									<linearGradient id="emptyGrad" x1="0%" y1="0%" x2="100%" y2="100%">
										<stop offset="0%" stop-color="#1e96eb" />
										<stop offset="100%" stop-color="#7c3aed" />
									</linearGradient>
								</defs>
								<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
								<polyline points="14,2 14,8 20,8"/>
								<line x1="16" y1="13" x2="8" y2="13"/>
								<line x1="16" y1="17" x2="8" y2="17"/>
							</svg>
						</div>
						<h2 class="kb-page__empty-title">Start your knowledge base</h2>
						<p class="kb-page__empty-desc">Create pages to organize notes, docs, and ideas — all in one place.</p>
						<div class="kb-page__empty-actions">
							<button
								class="btn-compact btn-compact-primary"
								aria-label="Create first page"
								onclick={handleNewDocument}
							>
								<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M12 5v14"/><path d="M5 12h14"/></svg>
								New Page
							</button>
						</div>
						<div class="kb-page__empty-hints">
							<span class="kb-page__empty-hint">
								<kbd>Ctrl</kbd>+<kbd>K</kbd> to search
							</span>
							<span class="kb-page__empty-hint">
								<kbd>/</kbd> for commands inside pages
							</span>
						</div>
					</div>
				{:else}
					<!-- Document grid -->
					<div class="kb-page__grid">
						{#each documents as doc (doc.id)}
							<button
								class="kb-page__card"
								class:kb-page__card--active={currentDocumentId === doc.id}
								aria-label="Open {doc.title || 'Untitled'}"
								onclick={() => handleOpenDocument(doc.id)}
							>
								<div class="kb-page__card-icon">
									{#if getDocIcon(doc)}
										<span class="kb-page__card-emoji">{getDocIcon(doc)}</span>
									{:else}
										<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
											<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
											<polyline points="14,2 14,8 20,8"/>
										</svg>
									{/if}
								</div>
								<div class="kb-page__card-body">
									<span class="kb-page__card-title">{doc.title || 'Untitled'}</span>
									<span class="kb-page__card-meta">
										{formatDate(doc.updated_at)}
										{#if doc.is_favorite}
											<svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor" stroke="none"><path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/></svg>
										{/if}
									</span>
								</div>
								{#if doc.type}
									<span class="kb-page__card-badge">{doc.type}</span>
								{/if}
							</button>
						{/each}
					</div>
				{/if}
			</div>
		{/if}
	</main>

	<!-- Quick Search Modal -->
	<QuickSearch
		bind:open={showQuickSearch}
		onSelectDocument={handleOpenDocument}
	/>
</div>

<style>
	/* Foundation kb- page patterns with --dt/--dbg/--dbd tokens */
	.kb-page {
		display: flex;
		height: 100vh;
		width: 100%;
		background-color: var(--dbg);
		font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
	}

	.kb-page__main {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		background-color: var(--dbg);
		overflow-y: auto;
		-ms-overflow-style: none;
		scrollbar-width: none;
	}

	.kb-page__main::-webkit-scrollbar {
		display: none;
	}

	/* Center states (loading, error) */
	.kb-page__center {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		height: 100%;
		gap: 16px;
		color: var(--dt2);
	}

	.kb-page__center-text {
		font-size: 14px;
		margin: 0;
	}

	.kb-page__spinner {
		width: 28px;
		height: 28px;
		border: 3px solid var(--dbd);
		border-top-color: #1e96eb;
		border-radius: 50%;
		animation: kb-spin 0.8s linear infinite;
	}

	@keyframes kb-spin {
		to { transform: rotate(360deg); }
	}

	/* Listing layout */
	.kb-page__listing {
		flex: 1;
		display: flex;
		flex-direction: column;
		max-width: 900px;
		width: 100%;
		margin: 0 auto;
		padding: 40px 32px;
	}

	.kb-page__header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 32px;
	}

	.kb-page__title {
		font-size: 24px;
		font-weight: 600;
		color: var(--dt);
		margin: 0;
	}

	.kb-page__actions {
		display: flex;
		gap: 8px;
	}

	/* Empty state */
	.kb-page__empty {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		text-align: center;
		flex: 1;
		padding: 64px 32px;
	}

	.kb-page__empty-icon {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 80px;
		height: 80px;
		margin-bottom: 20px;
		background: var(--dbg2);
		border-radius: 50%;
		color: var(--dt4);
	}

	.kb-page__empty-title {
		font-size: 20px;
		font-weight: 600;
		color: var(--dt);
		margin: 0 0 8px;
	}

	.kb-page__empty-desc {
		font-size: 14px;
		color: var(--dt2);
		margin: 0 0 24px;
		max-width: 320px;
		line-height: 1.5;
	}

	/* Document grid */
	.kb-page__grid {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	.kb-page__card {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 10px 14px;
		border-radius: 8px;
		border: none;
		background: transparent;
		cursor: pointer;
		transition: background 0.12s;
		width: 100%;
		text-align: left;
	}

	.kb-page__card:hover {
		background: var(--dbg2);
	}

	.kb-page__card--active {
		background: var(--dbg2);
	}

	.kb-page__card-icon {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 28px;
		height: 28px;
		flex-shrink: 0;
		color: var(--dt4);
	}

	.kb-page__card-emoji {
		font-size: 18px;
		line-height: 1;
	}

	.kb-page__card-body {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	.kb-page__card-title {
		font-size: 14px;
		font-weight: 500;
		color: var(--dt);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.kb-page__card-meta {
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 12px;
		color: var(--dt3);
	}

	.kb-page__card-badge {
		font-size: 11px;
		padding: 2px 8px;
		border-radius: 4px;
		background: var(--dbg2);
		color: var(--dt2);
		text-transform: capitalize;
	}
</style>
