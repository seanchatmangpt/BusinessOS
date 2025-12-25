<script lang="ts">
	import { onMount } from 'svelte';
	import type { ContextListItem, ArtifactListItem } from '$lib/api/client';
	import { api } from '$lib/api/client';

	interface Props {
		pages: ContextListItem[];
		recentPages: ContextListItem[];
		onSelectPage: (page: ContextListItem) => void;
		onCreatePage: () => void;
	}

	let { pages, recentPages, onSelectPage, onCreatePage }: Props = $props();

	let artifacts = $state<ArtifactListItem[]>([]);
	let loadingArtifacts = $state(true);
	let activeFilter = $state<string | null>(null); // null = all, 'project', 'person', 'business', 'document'
	let searchQuery = $state('');

	// Stats
	const totalPages = $derived(pages.length);
	const totalDocuments = $derived(pages.filter(p => p.type === 'document').length);
	const totalProjects = $derived(pages.filter(p => p.type === 'project').length);
	const totalPeople = $derived(pages.filter(p => p.type === 'person').length);
	const totalBusiness = $derived(pages.filter(p => p.type === 'business').length);

	// Recent pages (last 6)
	const displayRecentPages = $derived(recentPages.slice(0, 6));

	// Recent artifacts (last 6)
	const recentArtifacts = $derived(artifacts.slice(0, 6));

	// Filtered and sorted pages
	const filteredPages = $derived.by(() => {
		let filtered = [...pages];

		// Apply type filter
		if (activeFilter) {
			filtered = filtered.filter(p => p.type === activeFilter);
		}

		// Apply search filter
		if (searchQuery.trim()) {
			const query = searchQuery.toLowerCase();
			filtered = filtered.filter(p =>
				p.name?.toLowerCase().includes(query) ||
				p.type?.toLowerCase().includes(query)
			);
		}

		// Sort by updated_at
		return filtered.sort((a, b) => {
			const aDate = a.updated_at ? new Date(a.updated_at).getTime() : 0;
			const bDate = b.updated_at ? new Date(b.updated_at).getTime() : 0;
			return bDate - aDate;
		});
	});

	onMount(async () => {
		try {
			const result = await api.getArtifacts();
			artifacts = result.sort((a, b) => {
				const aDate = a.created_at ? new Date(a.created_at).getTime() : 0;
				const bDate = b.created_at ? new Date(b.created_at).getTime() : 0;
				return bDate - aDate;
			});
		} catch (error) {
			console.error('Failed to load artifacts:', error);
		} finally {
			loadingArtifacts = false;
		}
	});

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

	function getArtifactIcon(type: string): string {
		switch (type) {
			case 'proposal':
				return 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z';
			case 'sop':
				return 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01';
			case 'framework':
				return 'M4 5a1 1 0 011-1h14a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1V5zM4 13a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H5a1 1 0 01-1-1v-6zM16 13a1 1 0 011-1h2a1 1 0 011 1v6a1 1 0 01-1 1h-2a1 1 0 01-1-1v-6z';
			case 'agenda':
				return 'M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z';
			case 'report':
				return 'M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z';
			case 'email':
				return 'M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z';
			default:
				return 'M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z';
		}
	}

	function formatDate(dateStr: string | undefined): string {
		if (!dateStr) return '';
		const date = new Date(dateStr);
		const now = new Date();
		const diff = now.getTime() - date.getTime();
		const days = Math.floor(diff / (1000 * 60 * 60 * 24));

		if (days === 0) return 'Today';
		if (days === 1) return 'Yesterday';
		if (days < 7) return `${days} days ago`;
		return date.toLocaleDateString();
	}

	function isEmoji(str: string): boolean {
		const emojiRegex = /^(\p{Emoji_Presentation}|\p{Emoji}\uFE0F)$/u;
		return emojiRegex.test(str);
	}
</script>

<div class="flex-1 overflow-auto p-8 bg-white dark:bg-[#1e1e20]">
	<div class="max-w-5xl mx-auto">
		<!-- Header -->
		<div class="mb-8">
			<h1 class="text-3xl font-bold text-gray-900 dark:text-gray-100 mb-2">Knowledge Base</h1>
			<p class="text-gray-500 dark:text-gray-400">Your central hub for all pages, documents, and artifacts</p>
		</div>

		<!-- Quick Stats -->
		<div class="grid grid-cols-2 md:grid-cols-5 gap-4 mb-8">
			<div class="bg-gray-50 dark:bg-gray-800/50 rounded-xl p-4 border border-gray-200 dark:border-gray-700">
				<div class="text-2xl font-bold text-gray-900 dark:text-gray-100">{totalPages}</div>
				<div class="text-sm text-gray-500 dark:text-gray-400">Total Pages</div>
			</div>
			<div class="bg-blue-50 dark:bg-blue-900/20 rounded-xl p-4 border border-blue-200 dark:border-blue-800">
				<div class="text-2xl font-bold text-blue-600 dark:text-blue-400">{totalProjects}</div>
				<div class="text-sm text-blue-600 dark:text-blue-400">Projects</div>
			</div>
			<div class="bg-green-50 dark:bg-green-900/20 rounded-xl p-4 border border-green-200 dark:border-green-800">
				<div class="text-2xl font-bold text-green-600 dark:text-green-400">{totalPeople}</div>
				<div class="text-sm text-green-600 dark:text-green-400">People</div>
			</div>
			<div class="bg-purple-50 dark:bg-purple-900/20 rounded-xl p-4 border border-purple-200 dark:border-purple-800">
				<div class="text-2xl font-bold text-purple-600 dark:text-purple-400">{totalBusiness}</div>
				<div class="text-sm text-purple-600 dark:text-purple-400">Business</div>
			</div>
			<div class="bg-orange-50 dark:bg-orange-900/20 rounded-xl p-4 border border-orange-200 dark:border-orange-800">
				<div class="text-2xl font-bold text-orange-600 dark:text-orange-400">{artifacts.length}</div>
				<div class="text-sm text-orange-600 dark:text-orange-400">Artifacts</div>
			</div>
		</div>

		<!-- Quick Actions -->
		<div class="flex gap-3 mb-8">
			<button
				onclick={() => onCreatePage()}
				class="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors"
			>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
				</svg>
				<span>New Page</span>
			</button>
		</div>

		<!-- Recent Pages -->
		{#if displayRecentPages.length > 0}
			<div class="mb-8">
				<h2 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">Recently Viewed</h2>
				<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
					{#each displayRecentPages as page (page.id)}
						<button
							onclick={() => onSelectPage(page)}
							class="flex items-start gap-3 p-4 bg-gray-50 dark:bg-gray-800/50 rounded-xl border border-gray-200 dark:border-gray-700 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors text-left"
						>
							<span class="w-8 h-8 flex items-center justify-center rounded-lg bg-white dark:bg-gray-700 border border-gray-200 dark:border-gray-600 flex-shrink-0">
								{#if page.icon && isEmoji(page.icon)}
									<span class="text-lg">{page.icon}</span>
								{:else}
									<svg class="w-4 h-4 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d={getTypeIcon(page.type || 'document')} />
									</svg>
								{/if}
							</span>
							<div class="flex-1 min-w-0">
								<div class="font-medium text-gray-900 dark:text-gray-100 truncate">
									{page.name || 'Untitled'}
								</div>
								<div class="text-xs text-gray-500 dark:text-gray-400 mt-0.5 capitalize">
									{page.type || 'document'}
								</div>
							</div>
						</button>
					{/each}
				</div>
			</div>
		{/if}

		<!-- Recent Artifacts -->
		{#if !loadingArtifacts && recentArtifacts.length > 0}
			<div class="mb-8">
				<h2 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">Recent Artifacts</h2>
				<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
					{#each recentArtifacts as artifact (artifact.id)}
						<div class="flex items-start gap-3 p-4 bg-orange-50 dark:bg-orange-900/10 rounded-xl border border-orange-200 dark:border-orange-800/50">
							<span class="w-8 h-8 flex items-center justify-center rounded-lg bg-orange-100 dark:bg-orange-900/30 flex-shrink-0">
								<svg class="w-4 h-4 text-orange-600 dark:text-orange-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d={getArtifactIcon(artifact.type)} />
								</svg>
							</span>
							<div class="flex-1 min-w-0">
								<div class="font-medium text-gray-900 dark:text-gray-100 truncate">
									{artifact.title || 'Untitled Artifact'}
								</div>
								<div class="text-xs text-gray-500 dark:text-gray-400 mt-0.5 capitalize">
									{artifact.type} &middot; {formatDate(artifact.created_at)}
								</div>
							</div>
						</div>
					{/each}
				</div>
			</div>
		{/if}

		<!-- All Pages -->
		<div>
			<div class="flex items-center justify-between mb-4">
				<h2 class="text-lg font-semibold text-gray-900 dark:text-gray-100">All Pages</h2>
				<!-- Search -->
				<div class="relative">
					<svg class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
					</svg>
					<input
						type="text"
						bind:value={searchQuery}
						placeholder="Search pages..."
						class="pl-10 pr-4 py-2 w-64 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm text-gray-900 dark:text-gray-100 placeholder:text-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500"
					/>
				</div>
			</div>

			<!-- Type Filter Tabs -->
			<div class="flex gap-2 mb-4 overflow-x-auto pb-2">
				<button
					onclick={() => activeFilter = null}
					class="px-4 py-2 rounded-lg text-sm font-medium transition-colors whitespace-nowrap {activeFilter === null ? 'bg-gray-900 text-white dark:bg-white dark:text-gray-900' : 'bg-gray-100 text-gray-600 hover:bg-gray-200 dark:bg-gray-800 dark:text-gray-400 dark:hover:bg-gray-700'}"
				>
					All ({totalPages})
				</button>
				<button
					onclick={() => activeFilter = 'project'}
					class="px-4 py-2 rounded-lg text-sm font-medium transition-colors whitespace-nowrap {activeFilter === 'project' ? 'bg-blue-600 text-white' : 'bg-blue-50 text-blue-600 hover:bg-blue-100 dark:bg-blue-900/20 dark:text-blue-400 dark:hover:bg-blue-900/30'}"
				>
					Projects ({totalProjects})
				</button>
				<button
					onclick={() => activeFilter = 'person'}
					class="px-4 py-2 rounded-lg text-sm font-medium transition-colors whitespace-nowrap {activeFilter === 'person' ? 'bg-green-600 text-white' : 'bg-green-50 text-green-600 hover:bg-green-100 dark:bg-green-900/20 dark:text-green-400 dark:hover:bg-green-900/30'}"
				>
					People ({totalPeople})
				</button>
				<button
					onclick={() => activeFilter = 'business'}
					class="px-4 py-2 rounded-lg text-sm font-medium transition-colors whitespace-nowrap {activeFilter === 'business' ? 'bg-purple-600 text-white' : 'bg-purple-50 text-purple-600 hover:bg-purple-100 dark:bg-purple-900/20 dark:text-purple-400 dark:hover:bg-purple-900/30'}"
				>
					Business ({totalBusiness})
				</button>
				<button
					onclick={() => activeFilter = 'document'}
					class="px-4 py-2 rounded-lg text-sm font-medium transition-colors whitespace-nowrap {activeFilter === 'document' ? 'bg-gray-600 text-white' : 'bg-gray-100 text-gray-600 hover:bg-gray-200 dark:bg-gray-800 dark:text-gray-400 dark:hover:bg-gray-700'}"
				>
					Documents ({totalDocuments})
				</button>
			</div>

			{#if filteredPages.length === 0}
				<div class="text-center py-12 text-gray-500 dark:text-gray-400">
					<svg class="w-12 h-12 mx-auto mb-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
					</svg>
					{#if activeFilter || searchQuery}
						<p class="mb-4">No pages found matching your filters.</p>
						<button
							onclick={() => { activeFilter = null; searchQuery = ''; }}
							class="inline-flex items-center gap-2 px-4 py-2 bg-gray-600 hover:bg-gray-700 text-white rounded-lg transition-colors"
						>
							Clear Filters
						</button>
					{:else}
						<p class="mb-4">No pages yet. Create your first page to get started.</p>
						<button
							onclick={() => onCreatePage()}
							class="inline-flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors"
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
							</svg>
							<span>Create Page</span>
						</button>
					{/if}
				</div>
			{:else}
				<div class="bg-white dark:bg-gray-800/50 rounded-xl border border-gray-200 dark:border-gray-700 overflow-hidden">
					<table class="w-full">
						<thead>
							<tr class="border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800">
								<th class="text-left px-4 py-3 text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Name</th>
								<th class="text-left px-4 py-3 text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Type</th>
								<th class="text-left px-4 py-3 text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Updated</th>
							</tr>
						</thead>
						<tbody class="divide-y divide-gray-200 dark:divide-gray-700">
							{#each filteredPages as page (page.id)}
								<tr
									class="hover:bg-gray-50 dark:hover:bg-gray-800 cursor-pointer transition-colors"
									onclick={() => onSelectPage(page)}
								>
									<td class="px-4 py-3">
										<div class="flex items-center gap-3">
											<span class="w-6 h-6 flex items-center justify-center flex-shrink-0">
												{#if page.icon && isEmoji(page.icon)}
													<span class="text-base">{page.icon}</span>
												{:else}
													<svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d={getTypeIcon(page.type || 'document')} />
													</svg>
												{/if}
											</span>
											<span class="font-medium text-gray-900 dark:text-gray-100 truncate">
												{page.name || 'Untitled'}
											</span>
										</div>
									</td>
									<td class="px-4 py-3">
										<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium capitalize
											{page.type === 'project' ? 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400' : ''}
											{page.type === 'person' ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400' : ''}
											{page.type === 'business' ? 'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-400' : ''}
											{page.type === 'document' || !page.type ? 'bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300' : ''}
										">
											{page.type || 'document'}
										</span>
									</td>
									<td class="px-4 py-3 text-sm text-gray-500 dark:text-gray-400">
										{formatDate(page.updated_at)}
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			{/if}
		</div>
	</div>
</div>
