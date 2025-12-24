<script lang="ts">
	import { slide, fade } from 'svelte/transition';
	import type { ContextListItem } from '$lib/api/client';

	interface Props {
		pages: ContextListItem[];
		isOpen: boolean;
		onClose: () => void;
		onSelect: (page: ContextListItem) => void;
	}

	let { pages, isOpen, onClose, onSelect }: Props = $props();

	let query = $state('');
	let selectedIndex = $state(0);
	let inputElement: HTMLInputElement | null = $state(null);

	// Filter results based on query
	const results = $derived.by(() => {
		if (!query.trim()) {
			// Show recent pages when no query
			return pages.slice(0, 10);
		}
		const q = query.toLowerCase();
		return pages.filter(p =>
			p.name?.toLowerCase().includes(q) ||
			p.type?.toLowerCase().includes(q)
		).slice(0, 20);
	});

	// Group results by date
	const groupedResults = $derived.by(() => {
		const items = results;
		const groups: { label: string; pages: ContextListItem[] }[] = [];

		const today = new Date();
		today.setHours(0, 0, 0, 0);
		const yesterday = new Date(today);
		yesterday.setDate(yesterday.getDate() - 1);
		const lastWeek = new Date(today);
		lastWeek.setDate(lastWeek.getDate() - 7);

		const todayItems: ContextListItem[] = [];
		const yesterdayItems: ContextListItem[] = [];
		const lastWeekItems: ContextListItem[] = [];
		const olderItems: ContextListItem[] = [];

		for (const page of items) {
			const updated = page.updated_at ? new Date(page.updated_at) : null;
			if (!updated) {
				olderItems.push(page);
			} else if (updated >= today) {
				todayItems.push(page);
			} else if (updated >= yesterday) {
				yesterdayItems.push(page);
			} else if (updated >= lastWeek) {
				lastWeekItems.push(page);
			} else {
				olderItems.push(page);
			}
		}

		if (todayItems.length > 0) groups.push({ label: 'Today', pages: todayItems });
		if (yesterdayItems.length > 0) groups.push({ label: 'Yesterday', pages: yesterdayItems });
		if (lastWeekItems.length > 0) groups.push({ label: 'Last 7 days', pages: lastWeekItems });
		if (olderItems.length > 0) groups.push({ label: 'Older', pages: olderItems });

		return groups;
	});

	// Flat list of all results for keyboard navigation
	const flatResults = $derived(groupedResults.flatMap(g => g.pages));

	// Focus input when opened
	$effect(() => {
		if (isOpen && inputElement) {
			setTimeout(() => inputElement?.focus(), 50);
		}
	});

	// Reset state when opened/closed
	$effect(() => {
		if (isOpen) {
			query = '';
			selectedIndex = 0;
		}
	});

	function handleKeydown(e: KeyboardEvent) {
		const items = flatResults;

		switch (e.key) {
			case 'Escape':
				e.preventDefault();
				onClose();
				break;
			case 'ArrowDown':
				e.preventDefault();
				selectedIndex = (selectedIndex + 1) % items.length;
				break;
			case 'ArrowUp':
				e.preventDefault();
				selectedIndex = (selectedIndex - 1 + items.length) % items.length;
				break;
			case 'Enter':
				e.preventDefault();
				if (items[selectedIndex]) {
					onSelect(items[selectedIndex]);
					onClose();
				}
				break;
		}
	}

	function handleBackdropClick(e: MouseEvent) {
		if (e.target === e.currentTarget) {
			onClose();
		}
	}

	function isEmoji(str: string): boolean {
		const emojiRegex = /^(\p{Emoji_Presentation}|\p{Emoji}\uFE0F)$/u;
		return emojiRegex.test(str);
	}

	function getTypeIcon(type: string): string {
		switch (type) {
			case 'project':
				return 'M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z';
			case 'person':
				return 'M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z';
			case 'business':
				return 'M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4';
			default:
				return 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z';
		}
	}
</script>

<svelte:window onkeydown={(e) => {
	// Global Cmd+K handler
	if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
		e.preventDefault();
		if (!isOpen) {
			// This won't work here, but the parent should handle opening
		}
	}
}} />

{#if isOpen}
	<!-- Backdrop -->
	<div
		class="fixed inset-0 z-50 bg-black/50 backdrop-blur-sm"
		transition:fade={{ duration: 150 }}
		onclick={handleBackdropClick}
		onkeydown={handleKeydown}
		role="dialog"
		aria-modal="true"
		tabindex="-1"
	>
		<!-- Modal -->
		<div
			class="fixed top-[15%] left-1/2 -translate-x-1/2 w-full max-w-2xl bg-white dark:bg-[#2c2c2e] rounded-xl shadow-2xl overflow-hidden"
			transition:slide={{ duration: 200 }}
		>
			<!-- Search Input -->
			<div class="flex items-center gap-3 px-4 py-3 border-b border-gray-200 dark:border-gray-700">
				<svg class="w-5 h-5 text-gray-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
				</svg>
				<input
					bind:this={inputElement}
					type="text"
					bind:value={query}
					placeholder="Search pages..."
					class="flex-1 bg-transparent border-0 outline-none text-gray-900 dark:text-gray-100 placeholder:text-gray-400 dark:placeholder:text-gray-500 text-base"
					onkeydown={handleKeydown}
				/>
				<kbd class="px-2 py-0.5 text-xs text-gray-400 bg-gray-100 dark:bg-gray-700 rounded">esc</kbd>
			</div>

			<!-- Results -->
			<div class="max-h-[400px] overflow-y-auto">
				{#if flatResults.length === 0}
					<div class="px-4 py-8 text-center text-gray-500 dark:text-gray-400">
						{#if query}
							<p>No results found for "{query}"</p>
						{:else}
							<p>No pages yet</p>
						{/if}
					</div>
				{:else}
					{#each groupedResults as group}
						<div class="py-1">
							<div class="px-4 py-1.5 text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
								{group.label}
							</div>
							{#each group.pages as page, i}
								{@const globalIndex = flatResults.indexOf(page)}
								<button
									onclick={() => { onSelect(page); onClose(); }}
									class="w-full flex items-center gap-3 px-4 py-2 text-left hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors
										{globalIndex === selectedIndex ? 'bg-blue-50 dark:bg-blue-900/30' : ''}"
								>
									<!-- Icon -->
									<span class="w-6 h-6 flex items-center justify-center flex-shrink-0">
										{#if page.icon && isEmoji(page.icon)}
											<span class="text-lg">{page.icon}</span>
										{:else}
											<svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d={getTypeIcon(page.type || 'document')} />
											</svg>
										{/if}
									</span>

									<!-- Title & Type -->
									<div class="flex-1 min-w-0">
										<div class="text-sm font-medium text-gray-900 dark:text-gray-100 truncate">
											{page.name || 'Untitled'}
										</div>
										{#if page.type && page.type !== 'document'}
											<div class="text-xs text-gray-500 dark:text-gray-400 capitalize">
												{page.type}
											</div>
										{/if}
									</div>

									<!-- Selection indicator -->
									{#if globalIndex === selectedIndex}
										<svg class="w-4 h-4 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
										</svg>
									{/if}
								</button>
							{/each}
						</div>
					{/each}
				{/if}
			</div>

			<!-- Footer with keyboard shortcuts -->
			<div class="px-4 py-2 border-t border-gray-200 dark:border-gray-700 flex items-center gap-4 text-xs text-gray-500 dark:text-gray-400">
				<span class="flex items-center gap-1">
					<kbd class="px-1.5 py-0.5 bg-gray-100 dark:bg-gray-700 rounded">↑</kbd>
					<kbd class="px-1.5 py-0.5 bg-gray-100 dark:bg-gray-700 rounded">↓</kbd>
					<span>to navigate</span>
				</span>
				<span class="flex items-center gap-1">
					<kbd class="px-1.5 py-0.5 bg-gray-100 dark:bg-gray-700 rounded">↵</kbd>
					<span>to open</span>
				</span>
				<span class="flex items-center gap-1">
					<kbd class="px-1.5 py-0.5 bg-gray-100 dark:bg-gray-700 rounded">esc</kbd>
					<span>to close</span>
				</span>
			</div>
		</div>
	</div>
{/if}
