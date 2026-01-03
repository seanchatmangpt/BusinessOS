<script lang="ts">
	import { fly, fade } from 'svelte/transition';
	import { onMount, onDestroy } from 'svelte';
	import type { Memory, MemoryType } from '$lib/api/memory/types';
	import type { ContextListItem } from '$lib/api/client';
	import { updateMemory, deleteMemory } from '$lib/api/memory';
	import { updateContext, archiveContext } from '$lib/api/contexts';
	import { editor, wordCount, type EditorBlock, createEmptyBlock } from '$lib/stores/editor';
	import BlockComponent from '$lib/components/editor/Block.svelte';
	import BlockMenu from '$lib/components/editor/BlockMenu.svelte';
	import { markdownToBlocks, blocksToMarkdown } from '$lib/utils/markdown-blocks';

	interface Props {
		selectedMemory?: Memory | null;
		selectedDocument?: ContextListItem | null;
		relatedMemories?: Memory[];
		onClose?: () => void;
		onEdit?: () => void;
		onAddToChat?: () => void;
		onSave?: (memory: Memory) => void;
		onDelete?: (id: string) => void;
		onSelectRelated?: (memory: Memory) => void;
		onExpand?: () => void;
	}

	let {
		selectedMemory = null,
		selectedDocument = null,
		relatedMemories = [],
		onClose,
		onEdit,
		onAddToChat,
		onSave,
		onDelete,
		onSelectRelated,
		onExpand
	}: Props = $props();

	// Editing states
	let isEditingTitle = $state(false);
	let editedTitle = $state('');
	let editedCoverImage = $state<string | null>(null);

	// File input reference
	let coverImageInput = $state<HTMLInputElement | null>(null);

	// Saving state
	let isSaving = $state(false);
	let hasChanges = $state(false);
	let saveTimeout: ReturnType<typeof setTimeout> | null = null;
	let lastSaved = $state<Date | null>(null);
	let currentMemoryId = $state<string | null>(null);

	// Memory type display
	const typeLabels: Record<string, string> = {
		'fact': 'Fact',
		'preference': 'Preference',
		'decision': 'Decision',
		'event': 'Event',
		'learning': 'Learning',
		'context': 'Context',
		'relationship': 'Relationship',
		'episode': 'Episode'
	};

	// Source icons (for showing where the memory came from)
	const sourceIcons: Record<string, string> = {
		'notion': '/logos/integrations/notion.svg',
		'gmail': '/logos/integrations/gmail.svg',
		'slack': '/logos/integrations/slack.svg',
		'calendar': '/logos/integrations/calendar.svg'
	};

	// Initialize edited values and blocks when memory changes
	$effect(() => {
		const newId = selectedMemory?.id || null;
		if (newId && newId !== currentMemoryId) {
			currentMemoryId = newId;
			editedTitle = selectedMemory?.title || '';
			editedCoverImage = selectedMemory?.cover_image || null;
			hasChanges = false;

			// Initialize block editor with content
			const content = selectedMemory?.content || '';
			if (content) {
				const blocks = markdownToBlocks(content);
				editor.initialize(blocks);
			} else {
				editor.initialize([createEmptyBlock()]);
			}
		}
	});

	// Auto-save when editor becomes dirty
	$effect(() => {
		if ($editor.isDirty && currentMemoryId) {
			hasChanges = true;
			// Debounce save
			if (saveTimeout) clearTimeout(saveTimeout);
			saveTimeout = setTimeout(() => {
				saveChanges();
			}, 1500);
		}
	});

	// Cleanup on destroy
	onDestroy(() => {
		if (saveTimeout) clearTimeout(saveTimeout);
		editor.reset();
	});

	// Handle cover image upload
	function handleCoverImageUpload(event: Event) {
		const input = event.target as HTMLInputElement;
		const file = input.files?.[0];
		if (!file) return;

		// Validate file type
		if (!file.type.startsWith('image/')) {
			alert('Please select an image file');
			return;
		}

		// Validate file size (max 5MB)
		if (file.size > 5 * 1024 * 1024) {
			alert('Image must be less than 5MB');
			return;
		}

		// Convert to data URL for preview and storage
		const reader = new FileReader();
		reader.onload = (e) => {
			editedCoverImage = e.target?.result as string;
			hasChanges = true;
		};
		reader.readAsDataURL(file);
	}

	// Remove cover image
	function removeCoverImage() {
		editedCoverImage = null;
		hasChanges = true;
	}

	// Format date nicely - pickledOS style
	function formatDate(dateString: string | null): string {
		if (!dateString) return '';
		const date = new Date(dateString);
		return date.toLocaleDateString('en-US', {
			month: 'short',
			day: 'numeric',
			year: 'numeric'
		}) + ' · ' + date.toLocaleTimeString('en-US', {
			hour: 'numeric',
			minute: '2-digit',
			hour12: true
		});
	}

	// Format date short for related clouds
	function formatDateShort(dateString: string | null): string {
		if (!dateString) return '';
		const date = new Date(dateString);
		return date.toLocaleDateString('en-US', {
			month: 'short',
			day: 'numeric',
			year: 'numeric'
		}) + ' · ' + date.toLocaleTimeString('en-US', {
			hour: 'numeric',
			minute: '2-digit',
			hour12: true
		});
	}

	// Get type label
	function getTypeLabel(type: string | null | undefined): string {
		return typeLabels[type || 'context'] || 'Memory';
	}

	// Parse TL;DR into bullet points
	function parseTLDR(summary: string | null | undefined): string[] {
		if (!summary) return [];
		// Split by newlines, bullet points, or numbered items
		const lines = summary.split(/\n|(?<=\. )(?=[A-Z])/);
		return lines
			.map(line => line.trim())
			.filter(line => line.length > 0)
			.map(line => line.replace(/^[-•*]\s*/, '').replace(/^\d+\.\s*/, ''));
	}

	// Save title
	function saveTitle() {
		isEditingTitle = false;
		if (editedTitle !== selectedMemory?.title) {
			hasChanges = true;
		}
	}


	// Save all changes
	async function saveChanges() {
		if (!selectedMemory || !hasChanges) return;

		isSaving = true;
		try {
			// Convert blocks to markdown for storage
			const editedContent = blocksToMarkdown($editor.blocks);

			// Check if this is a context (source_type === 'context') or actual memory
			if (selectedMemory.source_type === 'context') {
				// Update via context API
				const updatedContext = await updateContext(selectedMemory.id, {
					name: editedTitle,
					content: editedContent,
					cover_image: editedCoverImage || undefined
				});

				// Convert back to memory format for callback
				const updatedMemory: Memory = {
					...selectedMemory,
					title: updatedContext.name || editedTitle,
					content: updatedContext.content || editedContent,
					cover_image: editedCoverImage || undefined
				};

				editor.markSaved();
				hasChanges = false;
				lastSaved = new Date();
				onSave?.(updatedMemory);
			} else {
				// Update via memory API
				const updatedMemory = await updateMemory(selectedMemory.id, {
					title: editedTitle,
					content: editedContent,
					cover_image: editedCoverImage || undefined
				});

				editor.markSaved();
				hasChanges = false;
				lastSaved = new Date();
				onSave?.(updatedMemory);
			}
		} catch (err) {
			console.error('Failed to save changes:', err);
		} finally {
			isSaving = false;
		}
	}

	// Add new block at end
	function addNewBlockAtEnd() {
		const lastBlock = $editor.blocks[$editor.blocks.length - 1];
		if (lastBlock) {
			editor.addBlockAfter(lastBlock.id, 'paragraph');
		}
	}

	// Format time for saved indicator
	function formatTime(date: Date) {
		return date.toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' });
	}

	// Delete memory
	async function handleDelete() {
		if (!selectedMemory) return;
		if (!confirm('Are you sure you want to delete this?')) return;

		try {
			// Check if this is a context or actual memory
			if (selectedMemory.source_type === 'context') {
				// Archive context (soft delete)
				await archiveContext(selectedMemory.id);
			} else {
				// Delete memory
				await deleteMemory(selectedMemory.id);
			}
			onDelete?.(selectedMemory.id);
			onClose?.();
		} catch (err) {
			console.error('Failed to delete:', err);
		}
	}

	// Get TLDR bullet points
	let tldrPoints = $derived(parseTLDR(selectedMemory?.summary || selectedMemory?.learning_summary));
</script>

<div class="panel-container">
	{#if selectedMemory}
		<!-- Zoom Controls (left side) -->
		<div class="zoom-controls">
			<button class="zoom-btn" title="Zoom in">
				<svg width="16" height="16" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0zM10 7v6m3-3H7" />
				</svg>
			</button>
			<button class="zoom-btn" title="Zoom out">
				<svg width="16" height="16" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0zM13 10H7" />
				</svg>
			</button>
			<button class="zoom-btn" title="Reset view">
				<svg width="16" height="16" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 8V4m0 0h4M4 4l5 5m11-1V4m0 0h-4m4 0l-5 5M4 16v4m0 0h4m-4 0l5-5m11 5l-5-5m5 5v-4m0 4h-4" />
				</svg>
			</button>
		</div>

		<!-- Main Panel -->
		<div class="panel-content">
			<!-- Header -->
			<div class="panel-header">
				<div class="nav-buttons">
					<button class="nav-btn" title="Previous">
						<svg width="16" height="16" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
						</svg>
					</button>
					<button class="nav-btn" title="Next">
						<svg width="16" height="16" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
						</svg>
					</button>
					<button class="nav-btn" title="Refresh">
						<svg width="16" height="16" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
						</svg>
					</button>
				</div>
				<div class="header-actions">
					{#if hasChanges}
						<button onclick={saveChanges} disabled={isSaving} class="save-btn">
							{isSaving ? 'Saving...' : 'Save'}
						</button>
					{/if}
					<button onclick={onAddToChat} class="add-to-chat-btn">
						Add to Chat
					</button>
					<button onclick={handleDelete} class="action-btn" title="Delete">
						<svg width="18" height="18" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
						</svg>
					</button>
					<button onclick={onExpand} class="action-btn" title="Expand">
						<svg width="18" height="18" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 8V4m0 0h4M4 4l5 5m11-1V4m0 0h-4m4 0l-5 5M4 16v4m0 0h4m-4 0l5-5m11 5l-5-5m5 5v-4m0 4h-4" />
						</svg>
					</button>
					<button onclick={onClose} class="action-btn" title="Close">
						<svg width="18" height="18" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
						</svg>
					</button>
				</div>
			</div>

			<!-- Scrollable Content -->
			<div class="panel-scroll">
				<!-- Hidden file input for cover image -->
				<input
					type="file"
					accept="image/*"
					bind:this={coverImageInput}
					onchange={handleCoverImageUpload}
					style="display: none;"
				/>

				<!-- Cover Image (clickable to upload) -->
				{#if editedCoverImage}
					<div class="cover-image-wrapper">
						<div class="cover-image" onclick={() => coverImageInput?.click()} role="button" tabindex="0" onkeydown={(e) => e.key === 'Enter' && coverImageInput?.click()}>
							<img src={editedCoverImage} alt={selectedMemory.title || 'Memory'} />
							<div class="cover-overlay">
								<svg width="24" height="24" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 9a2 2 0 012-2h.93a2 2 0 001.664-.89l.812-1.22A2 2 0 0110.07 4h3.86a2 2 0 011.664.89l.812 1.22A2 2 0 0018.07 7H19a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V9z" />
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 13a3 3 0 11-6 0 3 3 0 016 0z" />
								</svg>
								<span>Change</span>
							</div>
						</div>
						<button class="remove-cover-btn" onclick={removeCoverImage} title="Remove cover image">
							<svg width="16" height="16" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
							</svg>
						</button>
					</div>
				{:else}
					<!-- Upload placeholder -->
					<div class="cover-placeholder" onclick={() => coverImageInput?.click()} role="button" tabindex="0" onkeydown={(e) => e.key === 'Enter' && coverImageInput?.click()}>
						<div class="icon-bubble">
							<svg width="32" height="32" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
							</svg>
						</div>
						<span class="upload-hint">Click to add cover image</span>
					</div>
				{/if}

				<!-- Title -->
				<div class="title-section">
					{#if isEditingTitle}
						<input
							type="text"
							bind:value={editedTitle}
							onblur={saveTitle}
							onkeydown={(e) => e.key === 'Enter' && saveTitle()}
							class="title-input"
							autofocus
						/>
					{:else}
						<h1
							class="title"
							ondblclick={() => isEditingTitle = true}
						>
							{editedTitle || selectedMemory.title || 'Untitled Memory'}
						</h1>
					{/if}
				</div>

				<!-- Meta Line -->
				<div class="meta-line">
					<span class="meta-date">{formatDate(selectedMemory.created_at)}</span>
					<span class="meta-type">{getTypeLabel(selectedMemory.memory_type)}</span>
					{#if selectedMemory.source_type}
						<img
							src={sourceIcons[selectedMemory.source_type] || '/logos/integrations/notion.svg'}
							alt={selectedMemory.source_type}
							class="source-icon"
						/>
					{/if}
				</div>

				<!-- TL;DR Section -->
				{#if tldrPoints.length > 0}
					<div class="tldr-section">
						<h3 class="section-label">TL;DR:</h3>
						<ul class="tldr-list">
							{#each tldrPoints as point}
								<li>{point}</li>
							{/each}
						</ul>
					</div>
				{/if}

				<!-- Content - Block Editor -->
				<div class="content-section">
					<div class="blocks-container" role="textbox" tabindex="-1">
						{#each $editor.blocks as block, index (block.id)}
							<BlockComponent {block} {index} />
						{/each}
					</div>

					<!-- Click area to add new blocks -->
					<button
						onclick={addNewBlockAtEnd}
						class="add-block-area"
					>
						<span class="add-block-hint">
							Click to add a block, or press / for commands
						</span>
					</button>
				</div>

				<!-- Slash Command Menu -->
				{#if $editor.showSlashMenu && $editor.slashMenuPosition}
					<div class="block-menu-wrapper">
						<BlockMenu />
					</div>
				{/if}

				<!-- Related Clouds Section -->
				{#if relatedMemories.length > 0}
					<div class="related-section">
						<h3 class="related-label">Related Clouds <span class="related-count">{relatedMemories.length}</span></h3>
						<div class="related-list">
							{#each relatedMemories as related}
								<button
									class="related-item"
									onclick={() => onSelectRelated?.(related)}
								>
									<div class="related-icon">
										{#if related.cover_image}
											<img src={related.cover_image} alt="" />
										{:else}
											<svg width="16" height="16" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
											</svg>
										{/if}
									</div>
									<div class="related-info">
										<span class="related-title">{related.title || 'Untitled'}</span>
										<span class="related-date">{formatDateShort(related.created_at)}</span>
									</div>
								</button>
							{/each}
						</div>
					</div>
				{/if}
			</div>

			<!-- Status Bar -->
			<div class="status-bar">
				<div class="status-left">
					<span>{$wordCount} words</span>
					<span>{$editor.blocks.length} blocks</span>
				</div>
				<div class="status-right">
					{#if isSaving}
						<span class="status-saving">
							<svg class="spin-icon" width="12" height="12" viewBox="0 0 24 24" fill="none">
								<circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="2" stroke-opacity="0.3"></circle>
								<path d="M12 2a10 10 0 0 1 10 10" stroke="currentColor" stroke-width="2" stroke-linecap="round"></path>
							</svg>
							Saving...
						</span>
					{:else if lastSaved}
						<span>Saved at {formatTime(lastSaved)}</span>
					{:else if hasChanges}
						<span class="status-unsaved">Unsaved changes</span>
					{:else}
						<span>No changes</span>
					{/if}
				</div>
			</div>
		</div>

	{:else if selectedDocument}
		<!-- Document View -->
		<div class="panel-content">
			<div class="panel-header">
				<div class="nav-buttons">
					<button class="nav-btn" title="Previous">
						<svg width="16" height="16" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
						</svg>
					</button>
					<button class="nav-btn" title="Next">
						<svg width="16" height="16" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
						</svg>
					</button>
				</div>
				<div class="header-actions">
					<button onclick={onClose} class="action-btn" title="Close">
						<svg width="18" height="18" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
						</svg>
					</button>
				</div>
			</div>
			<div class="panel-scroll">
				<div class="cover-placeholder">
					<div class="icon-bubble document">
						<svg width="32" height="32" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
						</svg>
					</div>
				</div>
				<div class="title-section">
					<h1 class="title">{selectedDocument.name}</h1>
				</div>
				<div class="meta-line">
					<span class="meta-date">{formatDate(selectedDocument.updated_at)}</span>
					<span class="meta-type">Document</span>
				</div>
				<div class="document-actions">
					<button onclick={onEdit} class="open-document-btn">
						Open Document
					</button>
				</div>
			</div>
		</div>

	{:else}
		<!-- Empty State -->
		<div class="empty-state">
			<div class="empty-icon">
				<svg width="32" height="32" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
				</svg>
			</div>
			<h4 class="empty-title">Select a Bubble</h4>
			<p class="empty-text">Click on a bubble in the knowledge graph to view its details.</p>
		</div>
	{/if}
</div>

<style>
	.panel-container {
		display: flex;
		height: 100%;
		background: #fafafa;
		position: relative;
	}

	/* Zoom Controls - pickledOS style */
	.zoom-controls {
		position: absolute;
		left: 12px;
		top: 50%;
		transform: translateY(-50%);
		display: flex;
		flex-direction: column;
		gap: 4px;
		z-index: 10;
	}

	.zoom-btn {
		width: 32px;
		height: 32px;
		display: flex;
		align-items: center;
		justify-content: center;
		background: white;
		border: 1px solid #e5e5e5;
		border-radius: 8px;
		color: #666;
		cursor: pointer;
		transition: all 0.15s;
	}

	.zoom-btn:hover {
		background: #f5f5f5;
		color: #333;
	}

	/* Main Panel */
	.panel-content {
		flex: 1;
		display: flex;
		flex-direction: column;
		background: white;
		border-radius: 16px;
		margin: 12px;
		margin-left: 56px;
		overflow: hidden;
		box-shadow: 0 4px 24px rgba(0, 0, 0, 0.08);
		position: relative;
	}

	/* Header */
	.panel-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 16px 20px;
		border-bottom: 1px solid #f0f0f0;
	}

	.nav-buttons {
		display: flex;
		gap: 4px;
	}

	.nav-btn {
		width: 28px;
		height: 28px;
		display: flex;
		align-items: center;
		justify-content: center;
		background: transparent;
		border: none;
		border-radius: 6px;
		color: #999;
		cursor: pointer;
		transition: all 0.15s;
	}

	.nav-btn:hover {
		background: #f5f5f5;
		color: #333;
	}

	.header-actions {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.save-btn {
		padding: 6px 14px;
		font-size: 13px;
		font-weight: 500;
		color: white;
		background: #22c55e;
		border: none;
		border-radius: 8px;
		cursor: pointer;
		transition: all 0.15s;
	}

	.save-btn:hover {
		background: #16a34a;
	}

	.save-btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.add-to-chat-btn {
		padding: 6px 14px;
		font-size: 13px;
		font-weight: 500;
		color: #555;
		background: transparent;
		border: none;
		cursor: pointer;
		transition: all 0.15s;
	}

	.add-to-chat-btn:hover {
		color: #111;
	}

	.action-btn {
		width: 32px;
		height: 32px;
		display: flex;
		align-items: center;
		justify-content: center;
		background: transparent;
		border: none;
		border-radius: 6px;
		color: #999;
		cursor: pointer;
		transition: all 0.15s;
	}

	.action-btn:hover {
		background: #f5f5f5;
		color: #333;
	}

	/* Scrollable Content */
	.panel-scroll {
		flex: 1;
		overflow-y: auto;
		padding: 24px;
	}

	/* Cover Image */
	.cover-image-wrapper {
		position: relative;
		margin-bottom: 20px;
	}

	.cover-placeholder {
		margin-bottom: 20px;
		cursor: pointer;
		display: flex;
		flex-direction: column;
		align-items: flex-start;
		gap: 8px;
	}

	.cover-placeholder:hover .icon-bubble {
		transform: scale(1.02);
		box-shadow: 0 4px 16px rgba(139, 115, 85, 0.25);
	}

	.upload-hint {
		font-size: 11px;
		color: #888;
		font-weight: 500;
	}

	.cover-image {
		width: 80px;
		height: 80px;
		border-radius: 12px;
		overflow: hidden;
		box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
		cursor: pointer;
		position: relative;
		transition: transform 0.2s ease;
	}

	.cover-image:hover {
		transform: scale(1.02);
	}

	.cover-image:hover .cover-overlay {
		opacity: 1;
	}

	.cover-overlay {
		position: absolute;
		inset: 0;
		background: rgba(0, 0, 0, 0.5);
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 4px;
		color: white;
		font-size: 11px;
		font-weight: 500;
		opacity: 0;
		transition: opacity 0.2s ease;
	}

	.remove-cover-btn {
		position: absolute;
		top: -8px;
		right: -8px;
		width: 24px;
		height: 24px;
		border-radius: 50%;
		background: #ff4444;
		border: 2px solid white;
		color: white;
		display: flex;
		align-items: center;
		justify-content: center;
		cursor: pointer;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
		transition: transform 0.2s ease, background 0.2s ease;
	}

	.remove-cover-btn:hover {
		transform: scale(1.1);
		background: #ff2222;
	}

	.icon-bubble {
		width: 80px;
		height: 80px;
		border-radius: 12px;
		background: linear-gradient(135deg, #f5f0eb 0%, #e8e0d8 100%);
		display: flex;
		align-items: center;
		justify-content: center;
		color: #8B7355;
		box-shadow: 0 2px 12px rgba(139, 115, 85, 0.15);
		transition: transform 0.2s ease, box-shadow 0.2s ease;
	}

	.icon-bubble.document {
		background: linear-gradient(135deg, #e8f4ff 0%, #d0e8ff 100%);
		color: #4285F4;
		box-shadow: 0 2px 12px rgba(66, 133, 244, 0.15);
	}

	/* Title */
	.title-section {
		margin-bottom: 8px;
	}

	.title {
		font-size: 22px;
		font-weight: 600;
		color: #111;
		line-height: 1.3;
		margin: 0;
		cursor: text;
	}

	.title-input {
		width: 100%;
		font-size: 22px;
		font-weight: 600;
		color: #111;
		line-height: 1.3;
		border: none;
		background: #f5f5f5;
		border-radius: 8px;
		padding: 8px 12px;
		margin: -8px -12px;
	}

	.title-input:focus {
		outline: none;
		box-shadow: 0 0 0 2px #8B735533;
	}

	/* Meta Line */
	.meta-line {
		display: flex;
		align-items: center;
		gap: 8px;
		margin-bottom: 24px;
		font-size: 13px;
	}

	.meta-date {
		color: #888;
	}

	.meta-type {
		color: #666;
		padding: 2px 8px;
		background: #f5f5f5;
		border-radius: 4px;
	}

	.source-icon {
		width: 16px;
		height: 16px;
		opacity: 0.7;
	}

	/* TL;DR Section */
	.tldr-section {
		margin-bottom: 24px;
	}

	.section-label {
		font-size: 14px;
		font-weight: 600;
		color: #111;
		margin: 0 0 12px;
	}

	.tldr-list {
		margin: 0;
		padding-left: 20px;
		list-style-type: disc;
	}

	.tldr-list li {
		font-size: 14px;
		color: #444;
		line-height: 1.6;
		margin-bottom: 6px;
	}

	.tldr-list li:last-child {
		margin-bottom: 0;
	}

	/* Content Section - Block Editor */
	.content-section {
		margin-bottom: 24px;
		flex: 1;
	}

	.blocks-container {
		min-height: 100px;
	}

	.blocks-container :global([contenteditable]:focus) {
		outline: none;
	}

	.blocks-container :global(.block-wrapper) {
		color: #1f2937;
	}

	.blocks-container :global(.block-wrapper [contenteditable]) {
		color: #1f2937;
		caret-color: #1f2937;
	}

	.blocks-container :global(.block-wrapper [contenteditable]:empty::before) {
		color: #9ca3af;
	}

	.blocks-container :global(.block-wrapper h1),
	.blocks-container :global(.block-wrapper h2),
	.blocks-container :global(.block-wrapper h3) {
		color: #111827;
	}

	.blocks-container :global(.block-wrapper blockquote) {
		border-left-color: #d1d5db;
		color: #6b7280;
	}

	.blocks-container :global(.block-wrapper pre) {
		background-color: #f3f4f6;
		border-color: #e5e7eb;
	}

	.blocks-container :global(.block-wrapper code) {
		color: #374151;
	}

	.add-block-area {
		width: 100%;
		min-height: 48px;
		margin-top: 8px;
		text-align: left;
		background: transparent;
		border: none;
		cursor: text;
		display: block;
	}

	.add-block-hint {
		font-size: 13px;
		color: #9ca3af;
		opacity: 0;
		transition: opacity 0.15s;
	}

	.add-block-area:hover .add-block-hint {
		opacity: 1;
	}

	.block-menu-wrapper {
		position: absolute;
		z-index: 100;
	}

	/* Related Clouds Section */
	.related-section {
		border-top: 1px solid #f0f0f0;
		padding-top: 20px;
	}

	.related-label {
		font-size: 14px;
		font-weight: 500;
		color: #888;
		margin: 0 0 12px;
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.related-count {
		font-weight: 400;
		color: #bbb;
	}

	.related-list {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.related-item {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 12px;
		background: transparent;
		border: 1px solid #f0f0f0;
		border-radius: 10px;
		cursor: pointer;
		transition: all 0.15s;
		text-align: left;
	}

	.related-item:hover {
		background: #fafafa;
		border-color: #e5e5e5;
	}

	.related-icon {
		width: 36px;
		height: 36px;
		border-radius: 8px;
		background: #f5f5f5;
		display: flex;
		align-items: center;
		justify-content: center;
		color: #999;
		overflow: hidden;
		flex-shrink: 0;
	}

	.related-icon img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}

	.related-info {
		flex: 1;
		min-width: 0;
	}

	.related-title {
		display: block;
		font-size: 13px;
		font-weight: 500;
		color: #333;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.related-date {
		display: block;
		font-size: 11px;
		color: #999;
		margin-top: 2px;
	}

	/* Document Actions */
	.document-actions {
		margin-top: 24px;
	}

	.open-document-btn {
		width: 100%;
		padding: 14px 20px;
		font-size: 14px;
		font-weight: 500;
		color: white;
		background: #111;
		border: none;
		border-radius: 12px;
		cursor: pointer;
		transition: all 0.15s;
	}

	.open-document-btn:hover {
		background: #333;
	}

	/* Empty State */
	.empty-state {
		flex: 1;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 40px;
		text-align: center;
	}

	.empty-icon {
		width: 64px;
		height: 64px;
		border-radius: 50%;
		background: #f5f5f5;
		display: flex;
		align-items: center;
		justify-content: center;
		color: #ccc;
		margin-bottom: 16px;
	}

	.empty-title {
		font-size: 16px;
		font-weight: 600;
		color: #333;
		margin: 0 0 8px;
	}

	.empty-text {
		font-size: 13px;
		color: #888;
		margin: 0;
		max-width: 200px;
	}

	/* Status Bar */
	.status-bar {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 8px 16px;
		border-top: 1px solid #f0f0f0;
		background: #fafafa;
		font-size: 11px;
		color: #888;
	}

	.status-left {
		display: flex;
		gap: 12px;
	}

	.status-right {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.status-saving {
		display: flex;
		align-items: center;
		gap: 6px;
	}

	.spin-icon {
		animation: spin 1s linear infinite;
	}

	@keyframes spin {
		from { transform: rotate(0deg); }
		to { transform: rotate(360deg); }
	}

	.status-unsaved {
		color: #f59e0b;
	}
</style>
