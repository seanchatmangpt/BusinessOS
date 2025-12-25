<script lang="ts">
	import { editor, blockTypes, type BlockType } from '$lib/stores/editor';

	let selectedIndex = $state(0);

	// Filter block types based on query
	let filteredTypes = $derived(
		$editor.slashMenuQuery
			? blockTypes.filter(bt =>
				bt.label.toLowerCase().includes($editor.slashMenuQuery.toLowerCase()) ||
				bt.type.toLowerCase().includes($editor.slashMenuQuery.toLowerCase())
			)
			: blockTypes
	);

	function selectBlockType(type: BlockType) {
		if ($editor.focusedBlockId) {
			editor.changeBlockType($editor.focusedBlockId, type);
			editor.updateBlock($editor.focusedBlockId, '');
		}
		editor.hideSlashMenu();
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'ArrowDown') {
			e.preventDefault();
			selectedIndex = Math.min(selectedIndex + 1, filteredTypes.length - 1);
		} else if (e.key === 'ArrowUp') {
			e.preventDefault();
			selectedIndex = Math.max(selectedIndex - 1, 0);
		} else if (e.key === 'Enter' || e.key === 'Tab') {
			e.preventDefault();
			if (filteredTypes[selectedIndex]) {
				selectBlockType(filteredTypes[selectedIndex].type);
			}
		} else if (e.key === 'Escape') {
			e.preventDefault();
			editor.hideSlashMenu();
		}
	}
</script>

<svelte:window onkeydown={handleKeydown} />

{#if $editor.slashMenuPosition}
	<div
		class="block-menu fixed rounded-xl shadow-2xl border z-50 overflow-hidden w-72"
		style="left: {$editor.slashMenuPosition.x}px; top: {$editor.slashMenuPosition.y}px;"
	>
		<div class="menu-header px-3 py-2 border-b">
			<p class="text-xs uppercase tracking-wider font-medium">Basic blocks</p>
		</div>
		<div class="py-1 max-h-80 overflow-auto">
			{#each filteredTypes as blockType, idx}
				<button
					onclick={() => selectBlockType(blockType.type)}
					onmouseenter={() => selectedIndex = idx}
					class="menu-item w-full px-3 py-2.5 flex items-center gap-3 text-left transition-colors
						{idx === selectedIndex ? 'selected' : ''}"
				>
					<span class="icon-box w-10 h-10 rounded-lg border flex items-center justify-center text-base font-medium">
						{blockType.icon}
					</span>
					<div class="flex-1 min-w-0">
						<div class="item-label text-sm font-medium">{blockType.label}</div>
						<div class="item-desc text-xs">{blockType.description}</div>
					</div>
					{#if idx === selectedIndex}
						<kbd class="kbd-hint px-1.5 py-0.5 text-xs rounded">Enter</kbd>
					{/if}
				</button>
			{/each}
			{#if filteredTypes.length === 0}
				<div class="px-3 py-6 text-sm text-gray-400 text-center">
					No matching blocks found
				</div>
			{/if}
		</div>
	</div>
{/if}

<style>
	/* Light mode (default) */
	.block-menu {
		background-color: white;
		border-color: #e5e7eb;
	}

	.menu-header {
		border-color: #f3f4f6;
	}

	.menu-header p {
		color: #9ca3af;
	}

	.menu-item {
		color: #374151;
	}

	.menu-item:hover {
		background-color: #f9fafb;
	}

	.menu-item.selected {
		background-color: #f3f4f6;
	}

	.icon-box {
		background-color: #f9fafb;
		border-color: #e5e7eb;
		color: #4b5563;
	}

	.item-label {
		color: #111827;
	}

	.item-desc {
		color: #6b7280;
	}

	.kbd-hint {
		background-color: #e5e7eb;
		color: #6b7280;
	}

	/* Dark mode */
	:global(.dark) .block-menu {
		background-color: #2c2c2e;
		border-color: #374151;
	}

	:global(.dark) .menu-header {
		border-color: rgba(55, 65, 81, 0.5);
	}

	:global(.dark) .menu-header p {
		color: #9ca3af;
	}

	:global(.dark) .menu-item {
		color: #e5e7eb;
	}

	:global(.dark) .menu-item:hover {
		background-color: rgba(55, 65, 81, 0.5);
	}

	:global(.dark) .menu-item.selected {
		background-color: #374151;
	}

	:global(.dark) .icon-box {
		background-color: #1f2937;
		border-color: #4b5563;
		color: #d1d5db;
	}

	:global(.dark) .item-label {
		color: #f3f4f6;
	}

	:global(.dark) .item-desc {
		color: #9ca3af;
	}

	:global(.dark) .kbd-hint {
		background-color: #4b5563;
		color: #d1d5db;
	}
</style>
