<script lang="ts">
	import type { ContextListItem } from '$lib/api/client';

	interface Props {
		page: ContextListItem;
		depth?: number;
		isSelected?: boolean;
		isExpanded?: boolean;
		hasChildren?: boolean;
		onSelect: () => void;
		onExpand?: () => void;
		onAddChild?: () => void;
		onOpenPeek?: () => void;
		onDuplicate?: () => void;
		onRename?: () => void;
		onMove?: () => void;
		onDelete?: () => void;
		onToggleFavorite?: () => void;
		onCopyLink?: () => void;
		isFavorite?: boolean;
	}

	let {
		page,
		depth = 0,
		isSelected = false,
		isExpanded = false,
		hasChildren = false,
		onSelect,
		onExpand,
		onAddChild,
		onOpenPeek,
		onDuplicate,
		onRename,
		onMove,
		onDelete,
		onToggleFavorite,
		onCopyLink,
		isFavorite = false
	}: Props = $props();

	let menuOpen = $state(false);
	let isRenaming = $state(false);
	let renameValue = $state(page.name);

	// Check if icon is an emoji (not a text string like "chart-bar")
	function isEmoji(str: string | null): boolean {
		if (!str) return false;
		// Emoji regex - detects most common emoji patterns
		const emojiRegex = /[\u{1F300}-\u{1F9FF}]|[\u{2600}-\u{26FF}]|[\u{2700}-\u{27BF}]|[\u{1F600}-\u{1F64F}]|[\u{1F680}-\u{1F6FF}]|[\u{1F1E0}-\u{1F1FF}]/u;
		return emojiRegex.test(str);
	}

	function getTypeIcon(type: string): string {
		switch (type) {
			case 'person':
				return 'M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z';
			case 'business':
				return 'M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4';
			case 'project':
				return 'M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z';
			case 'document':
			default:
				return 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z';
		}
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'F2' && isSelected) {
			e.preventDefault();
			startRename();
		}
	}

	function startRename() {
		renameValue = page.name;
		isRenaming = true;
	}

	function finishRename() {
		if (renameValue.trim() && renameValue !== page.name) {
			onRename?.();
		}
		isRenaming = false;
	}

	function cancelRename() {
		renameValue = page.name;
		isRenaming = false;
	}

	function handleWindowClick(e: MouseEvent) {
		const target = e.target as HTMLElement;
		// Close menu when clicking outside
		if (menuOpen && !target.closest('.context-menu-container')) {
			menuOpen = false;
		}
	}
</script>

<svelte:window onkeydown={handleKeydown} onclick={handleWindowClick} />

<div
	class="group relative flex items-center gap-1 px-2 py-1.5 rounded-md cursor-pointer transition-colors
		{isSelected ? 'bg-blue-50 dark:bg-blue-900/30' : 'hover:bg-gray-100 dark:hover:bg-gray-800'}"
	style="padding-left: {8 + depth * 16}px;"
	role="treeitem"
	aria-selected={isSelected}
	aria-expanded={hasChildren ? isExpanded : undefined}
	onclick={() => { console.log('[SidebarPageItem] Row clicked:', page.name, page.id); onSelect(); }}
>
	<!-- Expand Arrow (if has children) -->
	{#if hasChildren}
		<button
			onclick={(e) => { e.stopPropagation(); onExpand?.(); }}
			class="w-5 h-5 flex items-center justify-center rounded hover:bg-gray-200 dark:hover:bg-gray-700 text-gray-400 transition-transform
				{isExpanded ? '' : '-rotate-90'}"
		>
			<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
			</svg>
		</button>
	{:else}
		<div class="w-5 h-5"></div>
	{/if}

	<!-- Page Icon and Name -->
	<div class="flex items-center gap-2 flex-1 min-w-0 text-left">
		<span class="w-5 h-5 flex items-center justify-center flex-shrink-0">
			{#if page.icon && isEmoji(page.icon)}
				<span class="text-base">{page.icon}</span>
			{:else}
				<svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d={getTypeIcon(page.type)} />
				</svg>
			{/if}
		</span>

		<!-- Page Name -->
		{#if isRenaming}
			<input
				type="text"
				bind:value={renameValue}
				onblur={finishRename}
				onkeydown={(e) => {
					if (e.key === 'Enter') finishRename();
					if (e.key === 'Escape') cancelRename();
				}}
				class="flex-1 min-w-0 px-1 py-0.5 text-sm bg-white dark:bg-gray-800 border border-blue-500 rounded outline-none"
				autofocus
			/>
		{:else}
			<span class="text-sm text-gray-700 dark:text-gray-200 truncate {isSelected ? 'font-medium' : ''}">
				{page.name || 'Untitled'}
			</span>
		{/if}
	</div>

	<!-- Hover Actions -->
	<div class="flex items-center gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity {menuOpen ? 'opacity-100' : ''}">
		<!-- Add Child Page -->
		<button
			onclick={(e) => { e.stopPropagation(); onAddChild?.(); }}
			class="w-6 h-6 flex items-center justify-center rounded hover:bg-gray-200 dark:hover:bg-gray-700 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
			title="Add page inside"
		>
			<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
			</svg>
		</button>

		<!-- Context Menu (Native Implementation) -->
		<div class="relative context-menu-container">
			<button
				class="w-6 h-6 flex items-center justify-center rounded hover:bg-gray-200 dark:hover:bg-gray-700 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
				onclick={(e) => { e.stopPropagation(); menuOpen = !menuOpen; }}
			>
				<svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
					<path d="M10 6a2 2 0 110-4 2 2 0 010 4zM10 12a2 2 0 110-4 2 2 0 010 4zM10 18a2 2 0 110-4 2 2 0 010 4z" />
				</svg>
			</button>

			{#if menuOpen}
				<div class="absolute right-0 top-full mt-1 z-50 min-w-[200px] bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-xl shadow-lg p-1">
					<!-- Add to Favorites -->
					<button
						class="w-full flex items-center gap-3 px-3 py-2 text-sm text-gray-700 dark:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg cursor-pointer transition-colors text-left"
						onclick={() => { console.log('[SidebarPageItem] Toggle Favorite clicked!'); menuOpen = false; onToggleFavorite?.(); }}
					>
						<svg class="w-4 h-4 {isFavorite ? 'text-yellow-500 fill-yellow-500' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z" />
						</svg>
						{isFavorite ? 'Remove from Favorites' : 'Add to Favorites'}
					</button>

					<div class="h-px bg-gray-200 dark:bg-gray-700 my-1"></div>

					<!-- Duplicate -->
					<button
						class="w-full flex items-center gap-3 px-3 py-2 text-sm text-gray-700 dark:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg cursor-pointer transition-colors text-left"
						onclick={() => { console.log('[SidebarPageItem] Duplicate clicked!'); menuOpen = false; onDuplicate?.(); }}
					>
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
						</svg>
						Duplicate
					</button>

					<!-- Rename -->
					<button
						class="w-full flex items-center gap-3 px-3 py-2 text-sm text-gray-700 dark:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg cursor-pointer transition-colors text-left"
						onclick={() => { console.log('[SidebarPageItem] Rename clicked!'); menuOpen = false; startRename(); }}
					>
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
						</svg>
						Rename
					</button>

					<!-- Move to -->
					<button
						class="w-full flex items-center gap-3 px-3 py-2 text-sm text-gray-700 dark:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg cursor-pointer transition-colors text-left"
						onclick={() => { console.log('[SidebarPageItem] Move clicked!'); menuOpen = false; onMove?.(); }}
					>
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
						</svg>
						Move to...
					</button>

					<div class="h-px bg-gray-200 dark:bg-gray-700 my-1"></div>

					<!-- Open in Side Peek -->
					<button
						class="w-full flex items-center gap-3 px-3 py-2 text-sm text-gray-700 dark:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg cursor-pointer transition-colors text-left"
						onclick={() => { console.log('[SidebarPageItem] Side Peek clicked!'); menuOpen = false; onOpenPeek?.(); }}
					>
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h7" />
						</svg>
						Open in side peek
					</button>

					<!-- Copy Link -->
					<button
						class="w-full flex items-center gap-3 px-3 py-2 text-sm text-gray-700 dark:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg cursor-pointer transition-colors text-left"
						onclick={() => { console.log('[SidebarPageItem] Copy Link clicked!'); menuOpen = false; onCopyLink?.(); }}
					>
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
						</svg>
						Copy link
					</button>

					<div class="h-px bg-gray-200 dark:bg-gray-700 my-1"></div>

					<!-- Delete -->
					<button
						class="w-full flex items-center gap-3 px-3 py-2 text-sm text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/30 rounded-lg cursor-pointer transition-colors text-left"
						onclick={() => { console.log('[SidebarPageItem] Delete clicked!'); menuOpen = false; onDelete?.(); }}
					>
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
						</svg>
						Delete
					</button>
				</div>
			{/if}
		</div>
	</div>

	<!-- Selected Indicator -->
	{#if isSelected}
		<div class="absolute left-0 top-1/2 -translate-y-1/2 w-0.5 h-4 bg-blue-600 rounded-r"></div>
	{/if}
</div>
