<script lang="ts">
	import { Star, MoreHorizontal, Share, Clock, X, Check } from 'lucide-svelte';
	import { Tooltip, Menu, MenuItem, MenuSeparator } from '$lib/ui';
	import { formatRelativeTime } from '$lib/utils';

	interface Props {
		isSaving?: boolean;
		lastSaved?: string | null;
		hasChanges?: boolean;
		readOnly?: boolean;
		isFavorite?: boolean;
		onClose?: () => void;
		onToggleFavorite?: () => void;
		onShare?: () => void;
		onExport?: () => void;
		onDelete?: () => void;
	}

	let {
		isSaving = false,
		lastSaved = null,
		hasChanges = false,
		readOnly = false,
		isFavorite = false,
		onClose,
		onToggleFavorite,
		onShare,
		onExport,
		onDelete
	}: Props = $props();

	let showMenu = $state(false);

	const saveStatus = $derived.by(() => {
		if (isSaving) return 'Saving...';
		if (hasChanges) return 'Unsaved changes';
		if (lastSaved) return `Saved ${formatRelativeTime(lastSaved)}`;
		return 'All changes saved';
	});
</script>

<div class="editor-toolbar">
	<div class="editor-toolbar__left">
		{#if onClose}
			<Tooltip content="Close" side="bottom">
				<button class="toolbar-icon-btn" onclick={onClose} aria-label="Close">
					<X class="h-4 w-4" />
				</button>
			</Tooltip>
		{/if}
	</div>

	<div class="editor-toolbar__center">
		<span class="editor-toolbar__status" class:editor-toolbar__status--saving={isSaving}>
			{#if isSaving}
				<div class="editor-toolbar__spinner"></div>
			{:else if !hasChanges}
				<Check class="h-3 w-3" />
			{/if}
			{saveStatus}
		</span>
	</div>

	<div class="editor-toolbar__right">
		{#if !readOnly}
			<Tooltip content="Share" side="bottom">
				<button class="toolbar-icon-btn" onclick={onShare} aria-label="Share">
					<Share class="h-4 w-4" />
				</button>
			</Tooltip>

			<Tooltip content={isFavorite ? 'Remove from favorites' : 'Add to favorites'} side="bottom">
				<button class="toolbar-icon-btn" class:toolbar-icon-btn--favorite={isFavorite} onclick={onToggleFavorite} aria-label="Toggle favorite">
					<Star class="h-4 w-4" />
				</button>
			</Tooltip>
		{/if}

		<Menu bind:open={showMenu}>
			{#snippet trigger()}
				<button class="toolbar-icon-btn" aria-label="More options">
					<MoreHorizontal class="h-4 w-4" />
				</button>
			{/snippet}

			<MenuItem onSelect={onShare}>
				{#snippet prefix()}
					<Share class="h-4 w-4" />
				{/snippet}
				Share
			</MenuItem>
			<MenuItem onSelect={onExport}>
				Export
			</MenuItem>
			<MenuItem onSelect={onToggleFavorite}>
				{#snippet prefix()}
					<Star class="h-4 w-4" />
				{/snippet}
				{isFavorite ? 'Remove from favorites' : 'Add to favorites'}
			</MenuItem>
			<MenuSeparator />
			<MenuItem destructive onSelect={onDelete}>
				Delete
			</MenuItem>
		</Menu>
	</div>
</div>

<style>
	.editor-toolbar {
		display: flex;
		align-items: center;
		justify-content: space-between;
		height: 44px;
		padding: 0 0.75rem;
		border-bottom: 1px solid var(--dbd);
		background-color: var(--dbg);
	}

	.editor-toolbar__left,
	.editor-toolbar__right {
		display: flex;
		align-items: center;
		gap: 0.25rem;
	}

	.editor-toolbar__center {
		display: flex;
		align-items: center;
	}

	.editor-toolbar__status {
		display: flex;
		align-items: center;
		gap: 0.375rem;
		font-size: 0.75rem;
		color: var(--dt3);
	}

	.editor-toolbar__status--saving {
		color: #1e96eb;
	}

	.editor-toolbar__spinner {
		width: 12px;
		height: 12px;
		border: 2px solid rgba(30, 150, 235, 0.3);
		border-top-color: #1e96eb;
		border-radius: 50%;
		animation: spin 0.6s linear infinite;
	}

	@keyframes spin {
		to { transform: rotate(360deg); }
	}

	/* Plain icon buttons - no circles */
	.toolbar-icon-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 28px;
		height: 28px;
		padding: 0;
		background: transparent;
		border: none;
		border-radius: 4px;
		color: var(--dt3);
		cursor: pointer;
		transition: color 0.15s, background-color 0.15s;
	}

	.toolbar-icon-btn:hover {
		color: var(--dt);
		background-color: var(--dbg2);
	}

	/* Favorite star when active - yellow works in both light/dark modes */
	.toolbar-icon-btn--favorite {
		color: hsl(48 96% 53%); /* amber-400 */
	}

	.toolbar-icon-btn--favorite :global(svg) {
		fill: hsl(48 96% 53%); /* amber-400 */
	}
</style>
