<script lang="ts">
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';

	// Check if running in Electron
	const isElectron = browser && typeof window !== 'undefined' && !!(window as any).electron;

	interface FileItem {
		id: string;
		name: string;
		type: 'folder' | 'file';
		size?: number;
		modified?: Date;
		extension?: string;
		children?: FileItem[];
	}

	// File source mode
	type FileSource = 'businessos' | 'local';
	let fileSource = $state<FileSource>('businessos');

	// Mock file system - in real app this would come from backend
	let fileSystem: FileItem[] = $state([
		{
			id: '1',
			name: 'Documents',
			type: 'folder',
			children: [
				{ id: '1-1', name: 'Business Plan.pdf', type: 'file', extension: 'pdf', size: 2400000, modified: new Date('2024-12-10') },
				{ id: '1-2', name: 'Q4 Report.xlsx', type: 'file', extension: 'xlsx', size: 156000, modified: new Date('2024-12-15') },
				{ id: '1-3', name: 'Meeting Notes.md', type: 'file', extension: 'md', size: 12000, modified: new Date('2024-12-16') },
				{
					id: '1-4',
					name: 'Contracts',
					type: 'folder',
					children: [
						{ id: '1-4-1', name: 'Client Agreement.docx', type: 'file', extension: 'docx', size: 45000 },
						{ id: '1-4-2', name: 'NDA Template.pdf', type: 'file', extension: 'pdf', size: 89000 },
					]
				}
			]
		},
		{
			id: '2',
			name: 'Projects',
			type: 'folder',
			children: [
				{ id: '2-1', name: 'Website Redesign', type: 'folder', children: [] },
				{ id: '2-2', name: 'Mobile App', type: 'folder', children: [] },
				{ id: '2-3', name: 'API Integration', type: 'folder', children: [] },
			]
		},
		{
			id: '3',
			name: 'Downloads',
			type: 'folder',
			children: [
				{ id: '3-1', name: 'installer.dmg', type: 'file', extension: 'dmg', size: 125000000 },
				{ id: '3-2', name: 'logo.png', type: 'file', extension: 'png', size: 456000 },
			]
		},
		{
			id: '4',
			name: 'Pictures',
			type: 'folder',
			children: []
		},
		{ id: '5', name: 'README.md', type: 'file', extension: 'md', size: 3200, modified: new Date('2024-12-01') },
		{ id: '6', name: 'config.json', type: 'file', extension: 'json', size: 1200, modified: new Date('2024-11-20') },
	]);

	// Navigation state
	let currentPath = $state<string[]>([]);
	let selectedItems = $state<Set<string>>(new Set());
	let viewMode = $state<'grid' | 'list'>('list');
	let searchQuery = $state('');

	// Get current directory items
	const currentItems = $derived(() => {
		let items = activeFileSystem;
		for (const segment of currentPath) {
			const folder = items.find(i => i.name === segment && i.type === 'folder');
			if (folder?.children) {
				items = folder.children;
			}
		}

		// Filter by search
		if (searchQuery.trim()) {
			const query = searchQuery.toLowerCase();
			items = items.filter(i => i.name.toLowerCase().includes(query));
		}

		// Sort: folders first, then alphabetically
		return [...items].sort((a, b) => {
			if (a.type === 'folder' && b.type !== 'folder') return -1;
			if (a.type !== 'folder' && b.type === 'folder') return 1;
			return a.name.localeCompare(b.name);
		});
	});

	// Switch file source mode
	function switchSource(source: FileSource) {
		fileSource = source;
		currentPath = [];
		selectedItems = new Set();
		searchQuery = '';
	}

	// Mock local file system (simulates macOS folders)
	const localFileSystem: FileItem[] = [
		{
			id: 'local-1',
			name: 'Desktop',
			type: 'folder',
			children: [
				{ id: 'local-1-1', name: 'Screenshot 2024-12-15.png', type: 'file', extension: 'png', size: 1240000 },
				{ id: 'local-1-2', name: 'Notes.txt', type: 'file', extension: 'txt', size: 2400 },
			]
		},
		{
			id: 'local-2',
			name: 'Documents',
			type: 'folder',
			children: [
				{ id: 'local-2-1', name: 'Resume.pdf', type: 'file', extension: 'pdf', size: 145000 },
				{ id: 'local-2-2', name: 'Tax Returns 2024', type: 'folder', children: [] },
			]
		},
		{
			id: 'local-3',
			name: 'Downloads',
			type: 'folder',
			children: [
				{ id: 'local-3-1', name: 'installer.dmg', type: 'file', extension: 'dmg', size: 89000000 },
				{ id: 'local-3-2', name: 'meeting-recording.mp4', type: 'file', extension: 'mp4', size: 245000000 },
			]
		},
		{
			id: 'local-4',
			name: 'Applications',
			type: 'folder',
			children: []
		},
		{
			id: 'local-5',
			name: 'Pictures',
			type: 'folder',
			children: []
		},
	];

	// Sidebar favorites based on source
	const businessOSFavorites = [
		{ name: 'Documents', icon: 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z' },
		{ name: 'Downloads', icon: 'M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4' },
		{ name: 'Projects', icon: 'M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z' },
		{ name: 'Pictures', icon: 'M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z' },
	];

	const localFavorites = [
		{ name: 'Desktop', icon: 'M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z' },
		{ name: 'Documents', icon: 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z' },
		{ name: 'Downloads', icon: 'M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4' },
		{ name: 'Applications', icon: 'M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z' },
		{ name: 'Pictures', icon: 'M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z' },
	];

	const favorites = $derived(fileSource === 'businessos' ? businessOSFavorites : localFavorites);
	const activeFileSystem = $derived(fileSource === 'businessos' ? fileSystem : localFileSystem);

	function navigateTo(path: string[]) {
		currentPath = path;
		selectedItems = new Set();
	}

	function openItem(item: FileItem) {
		if (item.type === 'folder') {
			currentPath = [...currentPath, item.name];
			selectedItems = new Set();
		} else {
			// Would open file preview or external app
			console.log('Opening file:', item.name);
		}
	}

	function goBack() {
		if (currentPath.length > 0) {
			currentPath = currentPath.slice(0, -1);
			selectedItems = new Set();
		}
	}

	function goHome() {
		currentPath = [];
		selectedItems = new Set();
	}

	function toggleSelect(itemId: string, event: MouseEvent) {
		if (event.metaKey || event.ctrlKey) {
			const newSelected = new Set(selectedItems);
			if (newSelected.has(itemId)) {
				newSelected.delete(itemId);
			} else {
				newSelected.add(itemId);
			}
			selectedItems = newSelected;
		} else {
			selectedItems = new Set([itemId]);
		}
	}

	function formatSize(bytes?: number): string {
		if (!bytes) return '-';
		if (bytes < 1024) return `${bytes} B`;
		if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
		if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
		return `${(bytes / (1024 * 1024 * 1024)).toFixed(1)} GB`;
	}

	function formatDate(date?: Date): string {
		if (!date) return '-';
		return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
	}

	function getFileIcon(extension?: string): { icon: string; color: string } {
		const icons: Record<string, { icon: string; color: string }> = {
			pdf: { icon: 'M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z', color: '#E53935' },
			docx: { icon: 'M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z', color: '#1565C0' },
			xlsx: { icon: 'M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z', color: '#2E7D32' },
			md: { icon: 'M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z', color: '#455A64' },
			json: { icon: 'M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z', color: '#FFA000' },
			png: { icon: 'M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z', color: '#7B1FA2' },
			jpg: { icon: 'M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z', color: '#7B1FA2' },
			dmg: { icon: 'M8 4H6a2 2 0 00-2 2v12a2 2 0 002 2h12a2 2 0 002-2V6a2 2 0 00-2-2h-2m-4-1v8m0 0l3-3m-3 3L9 8', color: '#546E7A' },
		};
		return icons[extension || ''] || { icon: 'M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z', color: '#78909C' };
	}
</script>

<div class="file-browser">
	<!-- Sidebar -->
	<aside class="file-sidebar">
		<!-- Source Switcher -->
		<div class="source-switcher">
			<button
				class="source-tab"
				class:active={fileSource === 'businessos'}
				onclick={() => switchSource('businessos')}
			>
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
					<path d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
				</svg>
				BusinessOS
			</button>
			<button
				class="source-tab"
				class:active={fileSource === 'local'}
				onclick={() => switchSource('local')}
			>
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
					<path d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
				</svg>
				This Mac
			</button>
		</div>

		{#if fileSource === 'local'}
			<div class="local-notice">
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
					<path d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
				</svg>
				<span>Read-only access to local files</span>
			</div>
		{/if}

		<div class="sidebar-section">
			<h3>{fileSource === 'businessos' ? 'Favorites' : 'Locations'}</h3>
			{#each favorites as fav}
				<button
					class="sidebar-item"
					class:active={currentPath.length === 1 && currentPath[0] === fav.name}
					onclick={() => navigateTo([fav.name])}
				>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
						<path d={fav.icon} />
					</svg>
					{fav.name}
				</button>
			{/each}
		</div>
	</aside>

	<!-- Main content -->
	<main class="file-main">
		<!-- Toolbar -->
		<div class="file-toolbar">
			<div class="toolbar-nav">
				<button class="toolbar-btn" onclick={goBack} disabled={currentPath.length === 0} title="Go Back">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M15 19l-7-7 7-7" />
					</svg>
				</button>
				<button class="toolbar-btn" onclick={goHome} title="Go Home">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
					</svg>
				</button>
			</div>

			<!-- Breadcrumb -->
			<div class="file-breadcrumb">
				<button onclick={goHome}>Home</button>
				{#each currentPath as segment, i}
					<span class="breadcrumb-sep">/</span>
					<button onclick={() => navigateTo(currentPath.slice(0, i + 1))}>{segment}</button>
				{/each}
			</div>

			<!-- Search & View toggle -->
			<div class="toolbar-right">
				<div class="search-box">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<circle cx="11" cy="11" r="8" />
						<path d="m21 21-4.35-4.35" />
					</svg>
					<input type="text" placeholder="Search..." bind:value={searchQuery} />
				</div>
				<div class="view-toggle">
					<button class:active={viewMode === 'list'} onclick={() => viewMode = 'list'} title="List View">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M4 6h16M4 12h16M4 18h16" />
						</svg>
					</button>
					<button class:active={viewMode === 'grid'} onclick={() => viewMode = 'grid'} title="Grid View">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<rect x="3" y="3" width="7" height="7" />
							<rect x="14" y="3" width="7" height="7" />
							<rect x="3" y="14" width="7" height="7" />
							<rect x="14" y="14" width="7" height="7" />
						</svg>
					</button>
				</div>
			</div>
		</div>

		<!-- File List -->
		{#if viewMode === 'list'}
			<div class="file-list">
				<div class="file-list-header">
					<span class="col-name">Name</span>
					<span class="col-date">Date Modified</span>
					<span class="col-size">Size</span>
				</div>
				{#each currentItems() as item (item.id)}
					{@const fileIcon = item.type === 'file' ? getFileIcon(item.extension) : null}
					<button
						class="file-list-item"
						class:selected={selectedItems.has(item.id)}
						onclick={(e) => toggleSelect(item.id, e)}
						ondblclick={() => openItem(item)}
					>
						<span class="col-name">
							{#if item.type === 'folder'}
								<svg class="item-icon folder" viewBox="0 0 24 24" fill="#3B82F6" stroke="none">
									<path d="M3 7V17C3 18.1046 3.89543 19 5 19H19C20.1046 19 21 18.1046 21 17V9C21 7.89543 20.1046 7 19 7H12L10 5H5C3.89543 5 3 5.89543 3 7Z"/>
								</svg>
							{:else}
								<svg class="item-icon" viewBox="0 0 24 24" fill="none" stroke={fileIcon?.color} stroke-width="1.5">
									<path d={fileIcon?.icon} />
								</svg>
							{/if}
							{item.name}
						</span>
						<span class="col-date">{formatDate(item.modified)}</span>
						<span class="col-size">{item.type === 'folder' ? '-' : formatSize(item.size)}</span>
					</button>
				{/each}

				{#if currentItems().length === 0}
					<div class="empty-folder">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
							<path d="M3 7V17C3 18.1046 3.89543 19 5 19H19C20.1046 19 21 18.1046 21 17V9C21 7.89543 20.1046 7 19 7H12L10 5H5C3.89543 5 3 5.89543 3 7Z"/>
						</svg>
						<p>This folder is empty</p>
					</div>
				{/if}
			</div>
		{:else}
			<!-- Grid View -->
			<div class="file-grid">
				{#each currentItems() as item (item.id)}
					{@const fileIcon = item.type === 'file' ? getFileIcon(item.extension) : null}
					<button
						class="file-grid-item"
						class:selected={selectedItems.has(item.id)}
						onclick={(e) => toggleSelect(item.id, e)}
						ondblclick={() => openItem(item)}
					>
						{#if item.type === 'folder'}
							<svg class="grid-icon folder" viewBox="0 0 24 24" fill="#3B82F6" stroke="none">
								<path d="M3 7V17C3 18.1046 3.89543 19 5 19H19C20.1046 19 21 18.1046 21 17V9C21 7.89543 20.1046 7 19 7H12L10 5H5C3.89543 5 3 5.89543 3 7Z"/>
							</svg>
						{:else}
							<svg class="grid-icon" viewBox="0 0 24 24" fill="none" stroke={fileIcon?.color} stroke-width="1.5">
								<path d={fileIcon?.icon} />
							</svg>
						{/if}
						<span class="grid-name">{item.name}</span>
					</button>
				{/each}
			</div>
		{/if}

		<!-- Status Bar -->
		<div class="file-status">
			<span>{currentItems().length} items</span>
			{#if selectedItems.size > 0}
				<span>{selectedItems.size} selected</span>
			{/if}
		</div>
	</main>
</div>

<style>
	.file-browser {
		display: flex;
		height: 100%;
		background: #fff;
		font-size: 13px;
	}

	/* Sidebar */
	.file-sidebar {
		width: 200px;
		background: #f5f5f7;
		border-right: 1px solid #e0e0e0;
		padding: 8px 0;
		flex-shrink: 0;
		display: flex;
		flex-direction: column;
	}

	.source-switcher {
		display: flex;
		gap: 2px;
		padding: 4px 8px;
		margin-bottom: 4px;
	}

	.source-tab {
		flex: 1;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 4px;
		padding: 6px 8px;
		border: none;
		background: transparent;
		border-radius: 6px;
		cursor: pointer;
		font-size: 11px;
		font-weight: 500;
		color: #666;
		transition: all 0.15s ease;
	}

	.source-tab:hover {
		background: rgba(0, 0, 0, 0.05);
	}

	.source-tab.active {
		background: #fff;
		color: #333;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
	}

	.source-tab svg {
		width: 14px;
		height: 14px;
	}

	.local-notice {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 6px 12px;
		margin: 0 8px 8px;
		background: #FEF3C7;
		border-radius: 6px;
		font-size: 10px;
		color: #92400E;
	}

	.local-notice svg {
		width: 12px;
		height: 12px;
		flex-shrink: 0;
	}

	.sidebar-section h3 {
		font-size: 11px;
		font-weight: 600;
		color: #666;
		text-transform: uppercase;
		padding: 8px 16px;
		margin: 0;
	}

	.sidebar-item {
		display: flex;
		align-items: center;
		gap: 8px;
		width: 100%;
		padding: 6px 16px;
		border: none;
		background: none;
		cursor: pointer;
		color: #333;
		text-align: left;
		font-size: 13px;
	}

	.sidebar-item:hover {
		background: rgba(0, 0, 0, 0.05);
	}

	.sidebar-item.active {
		background: rgba(0, 102, 255, 0.1);
		color: #0066FF;
	}

	.sidebar-item svg {
		width: 16px;
		height: 16px;
		flex-shrink: 0;
	}

	/* Main */
	.file-main {
		flex: 1;
		display: flex;
		flex-direction: column;
		min-width: 0;
	}

	/* Toolbar */
	.file-toolbar {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 8px 12px;
		border-bottom: 1px solid #e0e0e0;
		background: #fafafa;
	}

	.toolbar-nav {
		display: flex;
		gap: 4px;
	}

	.toolbar-btn {
		width: 28px;
		height: 28px;
		border: none;
		background: none;
		border-radius: 6px;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		color: #333;
	}

	.toolbar-btn:hover:not(:disabled) {
		background: rgba(0, 0, 0, 0.08);
	}

	.toolbar-btn:disabled {
		opacity: 0.3;
		cursor: not-allowed;
	}

	.toolbar-btn svg {
		width: 16px;
		height: 16px;
	}

	.file-breadcrumb {
		flex: 1;
		display: flex;
		align-items: center;
		gap: 4px;
		font-size: 13px;
		min-width: 0;
	}

	.file-breadcrumb button {
		border: none;
		background: none;
		cursor: pointer;
		padding: 4px 8px;
		border-radius: 4px;
		color: #333;
		white-space: nowrap;
	}

	.file-breadcrumb button:hover {
		background: rgba(0, 0, 0, 0.05);
	}

	.breadcrumb-sep {
		color: #999;
	}

	.toolbar-right {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.search-box {
		display: flex;
		align-items: center;
		gap: 6px;
		background: #fff;
		border: 1px solid #ddd;
		border-radius: 6px;
		padding: 4px 8px;
	}

	.search-box svg {
		width: 14px;
		height: 14px;
		color: #999;
	}

	.search-box input {
		border: none;
		background: none;
		outline: none;
		width: 120px;
		font-size: 12px;
	}

	.view-toggle {
		display: flex;
		border: 1px solid #ddd;
		border-radius: 6px;
		overflow: hidden;
	}

	.view-toggle button {
		width: 28px;
		height: 26px;
		border: none;
		background: #fff;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		color: #666;
	}

	.view-toggle button:not(:last-child) {
		border-right: 1px solid #ddd;
	}

	.view-toggle button.active {
		background: #e8e8e8;
		color: #333;
	}

	.view-toggle button svg {
		width: 14px;
		height: 14px;
	}

	/* File List */
	.file-list {
		flex: 1;
		overflow-y: auto;
	}

	.file-list-header {
		display: flex;
		padding: 8px 16px;
		border-bottom: 1px solid #e0e0e0;
		font-size: 11px;
		font-weight: 600;
		color: #666;
		text-transform: uppercase;
		background: #fafafa;
		position: sticky;
		top: 0;
	}

	.file-list-item {
		display: flex;
		align-items: center;
		padding: 8px 16px;
		border: none;
		background: none;
		width: 100%;
		cursor: pointer;
		text-align: left;
		border-bottom: 1px solid #f0f0f0;
	}

	.file-list-item:hover {
		background: rgba(0, 0, 0, 0.03);
	}

	.file-list-item.selected {
		background: rgba(0, 102, 255, 0.1);
	}

	.col-name {
		flex: 1;
		display: flex;
		align-items: center;
		gap: 10px;
		min-width: 0;
	}

	.col-date {
		width: 140px;
		color: #666;
		flex-shrink: 0;
	}

	.col-size {
		width: 80px;
		text-align: right;
		color: #666;
		flex-shrink: 0;
	}

	.item-icon {
		width: 20px;
		height: 20px;
		flex-shrink: 0;
	}

	.item-icon.folder {
		color: #3B82F6;
	}

	/* Grid View */
	.file-grid {
		flex: 1;
		overflow-y: auto;
		padding: 16px;
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(90px, 1fr));
		gap: 16px;
		align-content: start;
	}

	.file-grid-item {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 8px;
		padding: 12px 8px;
		border: none;
		background: none;
		border-radius: 8px;
		cursor: pointer;
	}

	.file-grid-item:hover {
		background: rgba(0, 0, 0, 0.05);
	}

	.file-grid-item.selected {
		background: rgba(0, 102, 255, 0.1);
	}

	.grid-icon {
		width: 48px;
		height: 48px;
	}

	.grid-name {
		font-size: 12px;
		text-align: center;
		max-width: 80px;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	/* Empty state */
	.empty-folder {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 60px 20px;
		color: #999;
	}

	.empty-folder svg {
		width: 64px;
		height: 64px;
		opacity: 0.3;
		margin-bottom: 16px;
	}

	/* Status Bar */
	.file-status {
		display: flex;
		gap: 16px;
		padding: 6px 16px;
		border-top: 1px solid #e0e0e0;
		background: #fafafa;
		font-size: 11px;
		color: #666;
	}
</style>
