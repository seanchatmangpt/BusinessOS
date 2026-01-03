<script lang="ts">
	import { fly, fade } from 'svelte/transition';
	import { onMount } from 'svelte';
	import {
		initiateAuth,
		getAllIntegrationsStatus,
		importFile,
		type IntegrationProvider,
		type GenericConnectionStatus
	} from '$lib/api/integrations';

	interface Integration {
		id: string;
		name: string;
		icon: string;
		iconBg?: string;
		iconType?: 'text' | 'svg';
		logoPath?: string; // Path to logo image in /static/logos/integrations/
		description: string;
		autoLiveSync: boolean;
		status: 'connected' | 'available' | 'coming_soon';
		totalNodes?: number;
		estNodes?: string;
		dataSince?: string;
		initialSync?: string;
		tooltip?: string;
		category: 'productivity' | 'communication' | 'ai' | 'meetings' | 'crm' | 'storage' | 'project' | 'custom';
		apiProvider?: IntegrationProvider; // Maps to backend provider ID
	}

	interface Props {
		open?: boolean;
		onClose?: () => void;
		onConnect?: (integration: Integration) => void;
		onCreateCustom?: () => void;
		inline?: boolean; // If true, render without fixed positioning (for embedding in containers)
	}

	let {
		open = false,
		onClose,
		onConnect,
		onCreateCustom,
		inline = false
	}: Props = $props();

	// Track hovered integration for tooltip
	let hoveredId = $state<string | null>(null);

	// Loading and connection states
	let loadingStatuses = $state(false);
	let connectingId = $state<string | null>(null);
	let connectionStatuses = $state<Record<string, GenericConnectionStatus>>({});
	let importFileInput: HTMLInputElement | null = null;
	let importingProvider = $state<string | null>(null);

	// Map frontend integration IDs to backend provider IDs
	const providerMap: Record<string, IntegrationProvider> = {
		'gmail': 'google',
		'google-calendar': 'google',
		'google-drive': 'google',
		'notion': 'notion',
		'slack': 'slack',
		'microsoft-teams': 'teams',
		'dropbox': 'dropbox',
		'hubspot': 'hubspot',
		'gohighlevel': 'gohighlevel',
		'salesforce': 'salesforce',
		'pipedrive': 'pipedrive',
		'linear': 'linear',
		'asana': 'asana',
		'monday': 'monday',
		'trello': 'trello',
		'jira': 'jira',
		'clickup': 'clickup',
		'zoom': 'zoom',
		'loom': 'loom',
		'fireflies': 'fireflies',
		'fathom': 'fathom',
		'tldv': 'tldv',
		'discord': 'discord',
		'evernote': 'evernote',
		'obsidian': 'obsidian',
		'roam': 'roam'
	};

	// AI providers that use file import instead of OAuth
	const fileImportProviders = ['chatgpt', 'claude', 'perplexity', 'gemini', 'granola'];

	// Load integration statuses when modal opens
	$effect(() => {
		if (open && Object.keys(connectionStatuses).length === 0) {
			loadIntegrationStatuses();
		}
	});

	async function loadIntegrationStatuses() {
		loadingStatuses = true;
		try {
			const response = await getAllIntegrationsStatus();
			connectionStatuses = response.integrations;
		} catch (error) {
			console.warn('Could not load integration statuses:', error);
		} finally {
			loadingStatuses = false;
		}
	}

	function getIntegrationStatus(integrationId: string): 'connected' | 'available' | 'coming_soon' {
		const provider = providerMap[integrationId];
		if (!provider) return 'available';

		const status = connectionStatuses[provider];
		if (status?.connected) return 'connected';
		return 'available';
	}

	// All available integrations - organized like pickledOS
	const integrations: Integration[] = [
		// Productivity - Email & Calendar (Auto Live-sync capable)
		{
			id: 'gmail',
			name: 'Gmail',
			icon: 'M',
			iconBg: 'linear-gradient(135deg, #EA4335 25%, #FBBC05 25%, #FBBC05 50%, #34A853 50%, #34A853 75%, #4285F4 75%)',
			logoPath: '/logos/integrations/gmail.svg',
			description: 'Import project details and track the context of important conversations.',
			autoLiveSync: true,
			status: 'available',
			estNodes: '50-200',
			initialSync: '15-30m',
			tooltip: 'Your new emails are processed into nodes every day.',
			category: 'productivity'
		},
		{
			id: 'google-calendar',
			name: 'Google Calendar',
			icon: '📅',
			iconBg: '#4285F4',
			logoPath: '/logos/integrations/calendar.svg',
			description: 'Sync your events so BusinessOS stays on top of meetings, plans, and deadlines.',
			autoLiveSync: true,
			status: 'available',
			estNodes: '20-100',
			initialSync: '5-10m',
			tooltip: 'Your calendar events are automatically synced to keep your schedule updated.',
			category: 'productivity'
		},
		{
			id: 'notion',
			name: 'Notion',
			icon: 'N',
			iconBg: '#000000',
			logoPath: '/logos/integrations/notion.svg',
			description: 'Sync your workspace pages, project roadmaps, and structured knowledge.',
			autoLiveSync: true,
			status: 'available',
			estNodes: '30-150',
			initialSync: '10-20m',
			tooltip: 'Your Notion updates are processed into nodes every day.',
			category: 'productivity'
		},
		{
			id: 'google-drive',
			name: 'Google Drive',
			icon: '▲',
			iconBg: 'linear-gradient(135deg, #4285F4, #34A853, #FBBC05)',
			description: 'Sync your documents, spreadsheets, and presentations into your knowledge base.',
			autoLiveSync: true,
			status: 'available',
			estNodes: '50-300',
			initialSync: '20-40m',
			tooltip: 'Your Drive files are indexed and searchable within your knowledge base.',
			category: 'storage'
		},
		{
			id: 'dropbox',
			name: 'Dropbox',
			icon: '📦',
			iconBg: '#0061FF',
			description: 'Import your files and folders to make them searchable and connected.',
			autoLiveSync: true,
			status: 'available',
			estNodes: '30-200',
			initialSync: '15-30m',
			tooltip: 'Your Dropbox files are continuously synced.',
			category: 'storage'
		},
		// Communication (Auto Live-sync capable)
		{
			id: 'slack',
			name: 'Slack',
			icon: '#',
			iconBg: '#4A154B',
			logoPath: '/logos/integrations/slack.svg',
			description: 'Extract key insights and memories from your team channels and DMs.',
			autoLiveSync: true,
			status: 'available',
			estNodes: '150-300',
			initialSync: '30-45m',
			tooltip: 'Your Slack messages are analyzed for important insights and decisions.',
			category: 'communication'
		},
		{
			id: 'microsoft-teams',
			name: 'Microsoft Teams',
			icon: 'T',
			iconBg: '#6264A7',
			description: 'Sync your Teams conversations, channels, and shared files.',
			autoLiveSync: true,
			status: 'available',
			estNodes: '100-250',
			initialSync: '25-40m',
			tooltip: 'Your Teams messages and files are synced automatically.',
			category: 'communication'
		},
		{
			id: 'discord',
			name: 'Discord',
			icon: 'D',
			iconBg: '#5865F2',
			description: 'Import conversations from your Discord servers and DMs.',
			autoLiveSync: true,
			status: 'coming_soon',
			estNodes: '100-300',
			initialSync: '20-35m',
			category: 'communication'
		},
		// AI Assistants (Manual sync - no auto live-sync)
		{
			id: 'chatgpt',
			name: 'ChatGPT',
			icon: '◯',
			iconBg: '#10A37F',
			logoPath: '/logos/integrations/openai.svg',
			description: 'Capture your brainstorming sessions, creative ideas, and problem-solving history.',
			autoLiveSync: false,
			status: 'available',
			estNodes: '80-120',
			initialSync: '30m',
			category: 'ai'
		},
		{
			id: 'claude',
			name: 'Claude',
			icon: '✦',
			iconBg: '#CC785C',
			logoPath: '/logos/integrations/claude.svg',
			description: 'Preserve your Claude in-depth discussions, research analysis, and writing drafts.',
			autoLiveSync: false,
			status: 'available',
			estNodes: '80-120',
			initialSync: '10-15m',
			category: 'ai'
		},
		{
			id: 'perplexity',
			name: 'Perplexity',
			icon: 'P',
			iconBg: '#20808D',
			description: 'Import your research queries, sources, and discovered insights.',
			autoLiveSync: false,
			status: 'available',
			estNodes: '40-80',
			initialSync: '10-15m',
			category: 'ai'
		},
		{
			id: 'gemini',
			name: 'Google Gemini',
			icon: '✧',
			iconBg: 'linear-gradient(135deg, #4285F4, #EA4335)',
			description: 'Sync your Gemini conversations and generated content.',
			autoLiveSync: false,
			status: 'coming_soon',
			estNodes: '60-100',
			initialSync: '15-20m',
			category: 'ai'
		},
		// Meetings (Auto Live-sync capable for most)
		{
			id: 'fireflies',
			name: 'Fireflies',
			icon: '🔥',
			iconBg: '#7C3AED',
			logoPath: '/logos/integrations/fireflies.svg',
			description: 'Turn meeting transcripts, summaries, and action items into memories.',
			autoLiveSync: true,
			status: 'available',
			estNodes: '20-50',
			initialSync: '10-15m',
			tooltip: 'Your meeting transcripts are processed into memories automatically.',
			category: 'meetings'
		},
		{
			id: 'fathom',
			name: 'Fathom',
			icon: '▶',
			iconBg: '#2563EB',
			logoPath: '/logos/integrations/fathom.svg',
			description: 'Turn meeting transcripts, summaries, and action items into memories.',
			autoLiveSync: true,
			status: 'available',
			estNodes: '20-50',
			initialSync: '10-15m',
			tooltip: 'Your meeting transcripts and summaries are processed automatically.',
			category: 'meetings'
		},
		{
			id: 'tldv',
			name: 'tl;dv',
			icon: '▷',
			iconBg: '#6366F1',
			description: 'Turn meeting transcripts, summaries, and action items into memories.',
			autoLiveSync: true,
			status: 'available',
			estNodes: '20-50',
			initialSync: '10-15m',
			tooltip: 'Your meeting recordings are transcribed and processed automatically.',
			category: 'meetings'
		},
		{
			id: 'granola',
			name: 'Granola',
			icon: 'G',
			iconBg: '#059669',
			description: 'Upload meeting notes to turn transcripts into memories.',
			autoLiveSync: false,
			status: 'available',
			estNodes: '20-50',
			initialSync: '10-15m',
			category: 'meetings'
		},
		{
			id: 'zoom',
			name: 'Zoom',
			icon: 'Z',
			iconBg: '#2D8CFF',
			description: 'Import meeting recordings, transcripts, and chat history.',
			autoLiveSync: true,
			status: 'available',
			estNodes: '30-80',
			initialSync: '15-25m',
			tooltip: 'Your Zoom recordings are automatically transcribed and imported.',
			category: 'meetings'
		},
		{
			id: 'loom',
			name: 'Loom',
			icon: 'L',
			iconBg: '#625DF5',
			description: 'Import your video messages and their transcripts.',
			autoLiveSync: true,
			status: 'available',
			estNodes: '15-40',
			initialSync: '10-15m',
			tooltip: 'Your Loom videos are transcribed and added automatically.',
			category: 'meetings'
		},
		// Project Management (Auto Live-sync capable)
		{
			id: 'linear',
			name: 'Linear',
			icon: '◇',
			iconBg: '#5E6AD2',
			description: 'Sync your issues, projects, and roadmaps for full context.',
			autoLiveSync: true,
			status: 'available',
			estNodes: '50-150',
			initialSync: '10-20m',
			tooltip: 'Your Linear issues and updates are synced in real-time.',
			category: 'project'
		},
		{
			id: 'asana',
			name: 'Asana',
			icon: '◉',
			iconBg: '#F06A6A',
			description: 'Import your tasks, projects, and team workflows.',
			autoLiveSync: true,
			status: 'available',
			estNodes: '40-120',
			initialSync: '15-25m',
			tooltip: 'Your Asana tasks and projects are synced automatically.',
			category: 'project'
		},
		{
			id: 'monday',
			name: 'Monday.com',
			icon: 'M',
			iconBg: '#FF3D57',
			description: 'Sync your boards, items, and updates into your knowledge base.',
			autoLiveSync: true,
			status: 'available',
			estNodes: '40-100',
			initialSync: '15-20m',
			tooltip: 'Your Monday boards are synced and updated automatically.',
			category: 'project'
		},
		{
			id: 'trello',
			name: 'Trello',
			icon: 'T',
			iconBg: '#0079BF',
			description: 'Import your boards, cards, and checklists.',
			autoLiveSync: true,
			status: 'available',
			estNodes: '30-80',
			initialSync: '10-15m',
			tooltip: 'Your Trello boards are synced in real-time.',
			category: 'project'
		},
		{
			id: 'jira',
			name: 'Jira',
			icon: 'J',
			iconBg: '#0052CC',
			description: 'Sync your issues, sprints, and project documentation.',
			autoLiveSync: true,
			status: 'available',
			estNodes: '60-200',
			initialSync: '20-35m',
			tooltip: 'Your Jira issues and sprints are synced automatically.',
			category: 'project'
		},
		{
			id: 'clickup',
			name: 'ClickUp',
			icon: 'C',
			iconBg: 'linear-gradient(135deg, #7B68EE, #49CCF9)',
			description: 'Import your tasks, docs, and workspace data.',
			autoLiveSync: true,
			status: 'available',
			estNodes: '50-150',
			initialSync: '15-25m',
			tooltip: 'Your ClickUp workspace is synced automatically.',
			category: 'project'
		},
		// CRM (Auto Live-sync capable)
		{
			id: 'hubspot',
			name: 'HubSpot',
			icon: '⬡',
			iconBg: '#FF7A59',
			logoPath: '/logos/integrations/hubspot.svg',
			description: 'Sync your CRM contacts, deals, and customer interactions into your knowledge base.',
			autoLiveSync: true,
			status: 'available',
			estNodes: '100-500',
			initialSync: '20-40m',
			tooltip: 'Your HubSpot contacts and deals are synced and analyzed for insights.',
			category: 'crm'
		},
		{
			id: 'gohighlevel',
			name: 'GoHighLevel',
			icon: '▲',
			iconBg: '#4CAF50',
			description: 'Import your marketing funnels, contacts, and automation data.',
			autoLiveSync: true,
			status: 'available',
			estNodes: '150-400',
			initialSync: '25-45m',
			tooltip: 'Your GHL contacts, funnels, and campaigns are synced automatically.',
			category: 'crm'
		},
		{
			id: 'salesforce',
			name: 'Salesforce',
			icon: 'S',
			iconBg: '#00A1E0',
			description: 'Sync your accounts, opportunities, and customer data.',
			autoLiveSync: true,
			status: 'available',
			estNodes: '200-600',
			initialSync: '30-60m',
			tooltip: 'Your Salesforce data is synced and enriched automatically.',
			category: 'crm'
		},
		{
			id: 'pipedrive',
			name: 'Pipedrive',
			icon: 'P',
			iconBg: '#25292C',
			description: 'Import your deals, contacts, and sales pipeline.',
			autoLiveSync: true,
			status: 'available',
			estNodes: '80-250',
			initialSync: '15-30m',
			tooltip: 'Your Pipedrive pipeline is synced in real-time.',
			category: 'crm'
		},
		// Notes (Manual sync mostly)
		{
			id: 'apple-notes',
			name: 'Apple Notes',
			icon: '📝',
			iconBg: '#FFD60A',
			description: 'Gather your spontaneous thoughts, quick checklists, and personal memos.',
			autoLiveSync: false,
			status: 'coming_soon',
			estNodes: '80-120',
			initialSync: '10-15m',
			category: 'productivity'
		},
		{
			id: 'evernote',
			name: 'Evernote',
			icon: 'E',
			iconBg: '#00A82D',
			description: 'Import your notes, notebooks, and web clips.',
			autoLiveSync: false,
			status: 'available',
			estNodes: '100-300',
			initialSync: '15-30m',
			category: 'productivity'
		},
		{
			id: 'obsidian',
			name: 'Obsidian',
			icon: '◈',
			iconBg: '#7C3AED',
			description: 'Sync your vault, notes, and knowledge graph connections.',
			autoLiveSync: false,
			status: 'available',
			estNodes: '50-200',
			initialSync: '10-20m',
			category: 'productivity'
		},
		{
			id: 'roam',
			name: 'Roam Research',
			icon: '◎',
			iconBg: '#343A40',
			description: 'Import your daily notes, linked references, and graph structure.',
			autoLiveSync: false,
			status: 'available',
			estNodes: '60-180',
			initialSync: '15-25m',
			category: 'productivity'
		}
	];

	// Sort integrations: ones with logos first, then ones without
	const sortedIntegrations = $derived(
		[...integrations].sort((a, b) => {
			const aHasLogo = a.logoPath ? 0 : 1;
			const bHasLogo = b.logoPath ? 0 : 1;
			return aHasLogo - bHasLogo;
		})
	);

	async function handleConnect(integration: Integration) {
		// Check if this is a file import provider (AI assistants)
		if (fileImportProviders.includes(integration.id)) {
			importingProvider = integration.id;
			// Trigger file input click
			importFileInput?.click();
			return;
		}

		const provider = providerMap[integration.id];
		if (!provider) {
			console.warn(`No provider mapping for integration: ${integration.id}`);
			onConnect?.(integration);
			return;
		}

		connectingId = integration.id;

		try {
			// Initiate OAuth flow
			const response = await initiateAuth(provider);
			if (response.auth_url) {
				// Open OAuth in new window/tab
				window.open(response.auth_url, '_blank', 'width=600,height=700');

				// Start polling for connection status
				pollForConnection(provider);
			}
		} catch (error) {
			console.error(`Failed to initiate ${provider} auth:`, error);
			alert(`Failed to connect to ${integration.name}. Please try again.`);
		} finally {
			connectingId = null;
		}

		onConnect?.(integration);
	}

	async function pollForConnection(provider: IntegrationProvider) {
		// Poll every 2 seconds for 2 minutes max
		const maxAttempts = 60;
		let attempts = 0;

		const poll = async () => {
			attempts++;
			try {
				const response = await getAllIntegrationsStatus();
				const status = response.integrations[provider];
				if (status?.connected) {
					connectionStatuses = response.integrations;
					return; // Success!
				}
			} catch (error) {
				console.warn('Poll error:', error);
			}

			if (attempts < maxAttempts) {
				setTimeout(poll, 2000);
			}
		};

		setTimeout(poll, 2000);
	}

	async function handleFileImport(event: Event) {
		const input = event.target as HTMLInputElement;
		const file = input.files?.[0];

		if (!file || !importingProvider) {
			importingProvider = null;
			return;
		}

		try {
			const source = importingProvider as 'chatgpt' | 'claude' | 'perplexity' | 'gemini' | 'other';
			const response = await importFile(file, source);

			if (response.success) {
				alert(`Successfully imported ${response.imported_count} items from ${importingProvider}.`);
				// Refresh statuses
				loadIntegrationStatuses();
			}
		} catch (error) {
			console.error(`Failed to import from ${importingProvider}:`, error);
			alert(`Failed to import data. Please make sure the file format is correct.`);
		} finally {
			importingProvider = null;
			input.value = ''; // Reset input
		}
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') {
			onClose?.();
		}
	}
</script>

<!-- Hidden file input for AI exports -->
<input
	bind:this={importFileInput}
	type="file"
	accept=".json,.zip"
	style="display: none;"
	onchange={handleFileImport}
/>

<svelte:window onkeydown={handleKeydown} />

{#if open}
	{#if !inline}
		<!-- Backdrop (only for non-inline mode) -->
		<div
			class="modal-backdrop"
			onclick={onClose}
			transition:fade={{ duration: 200 }}
		></div>

		<!-- Modal Container (only for non-inline mode) -->
		<div
			class="modal-container"
			transition:fly={{ y: 20, duration: 300 }}
		>
			<div class="modal-content">
			<!-- Header -->
			<div class="modal-header">
				<div class="header-text">
					<h2>Let's bring all your data into a single place.</h2>
					<p>When you connect your apps, we will process raw data and extract essential information and turn it into nodes.</p>
				</div>
				<button class="close-btn" onclick={onClose}>
					<svg width="20" height="20" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			</div>

			<!-- Scrollable Content Container -->
			<div class="modal-scroll">

			<!-- Section Header with Custom Connector Button -->
			<div class="section-header">
				<h3>Data Sources</h3>
				<button class="custom-connector-btn" onclick={onCreateCustom}>
					<svg width="16" height="16" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
					</svg>
					<span>Create Custom Connector</span>
					<span class="mcp-badge">MCP</span>
				</button>
			</div>

			<!-- Integrations Grid -->
			<div class="integrations-grid">
				{#each sortedIntegrations as integration (integration.id)}
					{@const dynamicStatus = integration.status === 'coming_soon' ? 'coming_soon' : getIntegrationStatus(integration.id)}
					{@const isConnecting = connectingId === integration.id}
					{@const isImporting = importingProvider === integration.id}
					<div
						class="integration-card"
						class:connected={dynamicStatus === 'connected'}
						class:coming-soon={dynamicStatus === 'coming_soon'}
						class:connecting={isConnecting || isImporting}
						onmouseenter={() => hoveredId = integration.id}
						onmouseleave={() => hoveredId = null}
					>
						<!-- Tooltip -->
						{#if hoveredId === integration.id && integration.tooltip && dynamicStatus !== 'coming_soon'}
							<div class="card-tooltip" transition:fade={{ duration: 150 }}>
								{integration.tooltip}
							</div>
						{/if}

						<div class="card-header">
							<div class="app-info">
								{#if integration.logoPath}
									<div class="app-icon logo">
										<img src={integration.logoPath} alt={integration.name} />
									</div>
								{:else}
									<div
										class="app-icon"
										style="background: {integration.iconBg}"
									>
										{integration.icon}
									</div>
								{/if}
								<span class="app-name">{integration.name}</span>
								{#if integration.autoLiveSync}
									<span class="auto-sync-badge">
										Auto Live-sync
										<svg width="12" height="12" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
										</svg>
									</span>
								{/if}
							</div>
							{#if dynamicStatus === 'connected'}
								<span class="status-badge connected">
									<span class="status-dot"></span>
									Live-Synced
								</span>
							{:else if dynamicStatus === 'coming_soon'}
								<span class="status-badge coming-soon">Soon</span>
							{:else if isConnecting || isImporting}
								<span class="status-badge connecting">
									<span class="spinner"></span>
									{isImporting ? 'Importing...' : 'Connecting...'}
								</span>
							{:else}
								<button
									class="connect-btn"
									onclick={() => handleConnect(integration)}
									disabled={loadingStatuses}
								>
									{fileImportProviders.includes(integration.id) ? 'Import' : 'Connect'}
								</button>
							{/if}
						</div>

						<p class="card-description">{integration.description}</p>

						<div class="card-stats">
							<div class="stat">
								<svg width="14" height="14" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
								</svg>
								<span class="stat-label">
									{dynamicStatus === 'connected' ? 'Tot. nodes' : 'Est. nodes'}
								</span>
								<span class="stat-value">
									{integration.totalNodes ?? integration.estNodes ?? '--'}
								</span>
							</div>
							<div class="stat">
								<svg width="14" height="14" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
								</svg>
								<span class="stat-label">
									{dynamicStatus === 'connected' ? 'Data since' : 'Initial sync'}
								</span>
								<span class="stat-value">
									{integration.dataSince ?? integration.initialSync ?? '--'}
								</span>
							</div>
						</div>
					</div>
				{/each}
			</div>
			</div>
		</div>
		</div>
	{:else}
		<!-- Inline mode - no backdrop or container, same layout as modal -->
		<div class="modal-content inline-mode">
			<!-- Header -->
			<div class="modal-header">
				<div class="header-text">
					<h2>Let's bring all your data into a single place.</h2>
					<p>When you connect your apps, we will process raw data and extract essential information and turn it into nodes.</p>
				</div>
				<button class="close-btn" onclick={onClose}>
					<svg width="20" height="20" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			</div>

			<!-- Scrollable Content -->
			<div class="modal-scroll">

			<!-- Section Header -->
			<div class="section-header">
				<h3>Data Sources</h3>
			</div>

			<!-- Integrations Grid - same as modal mode -->
			<div class="integrations-grid">
				{#each sortedIntegrations as integration (integration.id)}
					{@const dynamicStatus = integration.status === 'coming_soon' ? 'coming_soon' : getIntegrationStatus(integration.id)}
					{@const isConnecting = connectingId === integration.id}
					{@const isImporting = importingProvider === integration.id}
					<div
						class="integration-card"
						class:connected={dynamicStatus === 'connected'}
						class:coming-soon={dynamicStatus === 'coming_soon'}
						class:connecting={isConnecting || isImporting}
						onmouseenter={() => hoveredId = integration.id}
						onmouseleave={() => hoveredId = null}
					>
						<!-- Tooltip -->
						{#if hoveredId === integration.id && integration.tooltip && dynamicStatus !== 'coming_soon'}
							<div class="card-tooltip" transition:fade={{ duration: 150 }}>
								{integration.tooltip}
							</div>
						{/if}

						<div class="card-header">
							<div class="app-info">
								{#if integration.logoPath}
									<div class="app-icon logo">
										<img src={integration.logoPath} alt={integration.name} />
									</div>
								{:else}
									<div
										class="app-icon"
										style="background: {integration.iconBg}"
									>
										{integration.icon}
									</div>
								{/if}
								<span class="app-name">{integration.name}</span>
								{#if integration.autoLiveSync}
									<span class="auto-sync-badge">
										Auto Live-sync
										<svg width="12" height="12" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
										</svg>
									</span>
								{/if}
							</div>
							{#if dynamicStatus === 'connected'}
								<span class="status-badge connected">
									<span class="status-dot"></span>
									Live-Synced
								</span>
							{:else if dynamicStatus === 'coming_soon'}
								<span class="status-badge coming-soon">Soon</span>
							{:else if isConnecting || isImporting}
								<span class="status-badge connecting">
									<span class="spinner"></span>
									{isImporting ? 'Importing...' : 'Connecting...'}
								</span>
							{:else}
								<button
									class="connect-btn"
									onclick={() => handleConnect(integration)}
									disabled={loadingStatuses}
								>
									{fileImportProviders.includes(integration.id) ? 'Import' : 'Connect'}
								</button>
							{/if}
						</div>

						<p class="card-description">{integration.description}</p>

						<div class="card-stats">
							<div class="stat">
								<svg width="14" height="14" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
								</svg>
								<span class="stat-label">
									{dynamicStatus === 'connected' ? 'Tot. nodes' : 'Est. nodes'}
								</span>
								<span class="stat-value">
									{integration.totalNodes ?? integration.estNodes ?? '--'}
								</span>
							</div>
							<div class="stat">
								<svg width="14" height="14" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
								</svg>
								<span class="stat-label">
									{dynamicStatus === 'connected' ? 'Data since' : 'Initial sync'}
								</span>
								<span class="stat-value">
									{integration.dataSince ?? integration.initialSync ?? '--'}
								</span>
							</div>
						</div>
					</div>
				{/each}
			</div>
			</div>
		</div>
	{/if}
{/if}

<style>
	.modal-backdrop {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.4);
		backdrop-filter: blur(4px);
		z-index: 100;
	}

	.modal-container {
		position: fixed;
		inset: 0;
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 101;
		padding: 24px;
		pointer-events: none;
	}

	.modal-content {
		width: 100%;
		max-width: 1100px;
		max-height: 90vh;
		background: #fafafa;
		border-radius: 16px;
		overflow: hidden;
		pointer-events: auto;
		display: flex;
		flex-direction: column;
		box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
	}

	.modal-scroll {
		flex: 1;
		min-height: 0;
		overflow-y: auto;
		padding-bottom: 24px;
	}

	/* Inline mode - for embedding inside containers */
	.modal-content.inline-mode {
		max-width: 100%;
		max-height: 100%;
		height: 100%;
		border-radius: 32px;
		overflow: hidden;
		display: flex;
		flex-direction: column;
	}

	.modal-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		padding: 28px 32px 20px;
		background: white;
		flex-shrink: 0;
	}

	.header-text h2 {
		font-size: 24px;
		font-weight: 600;
		color: #1a1a1a;
		margin: 0 0 8px;
		letter-spacing: -0.3px;
	}

	.header-text p {
		font-size: 14px;
		color: #666;
		margin: 0;
		max-width: 650px;
		line-height: 1.5;
	}

	.close-btn {
		padding: 8px;
		background: transparent;
		border: none;
		color: #666;
		cursor: pointer;
		border-radius: 8px;
		transition: all 0.15s;
	}

	.close-btn:hover {
		background: #f0f0f0;
		color: #333;
	}

	.privacy-banner {
		display: flex;
		align-items: center;
		gap: 16px;
		padding: 16px 32px;
		background: #f5f5f5;
		border-top: 1px solid #eee;
		border-bottom: 1px solid #eee;
	}

	.privacy-icons {
		display: flex;
		gap: 8px;
	}

	.privacy-icon {
		width: 40px;
		height: 40px;
		display: flex;
		align-items: center;
		justify-content: center;
		background: white;
		border-radius: 50%;
		color: #666;
		border: 1px solid #e5e5e5;
	}

	.privacy-icon.soc2 {
		background: #f5f5f5;
	}

	.soc2-text {
		font-size: 8px;
		font-weight: 700;
		color: #666;
		text-align: center;
		line-height: 1.1;
	}

	.privacy-text {
		flex: 1;
		font-size: 13px;
		color: #555;
	}

	.privacy-text strong {
		display: block;
		color: #333;
		margin-bottom: 2px;
	}

	.learn-more-btn {
		padding: 8px 16px;
		background: white;
		border: 1px solid #ddd;
		border-radius: 8px;
		font-size: 13px;
		color: #555;
		cursor: pointer;
		transition: all 0.15s;
		white-space: nowrap;
	}

	.learn-more-btn:hover {
		background: #f5f5f5;
		border-color: #ccc;
	}

	.section-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 20px 32px 12px;
		background: white;
		position: relative;
		z-index: 1;
	}

	.section-header h3 {
		font-size: 15px;
		font-weight: 600;
		color: #333;
		margin: 0;
	}

	.custom-connector-btn {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 8px 14px;
		background: white;
		border: 1px solid #ddd;
		border-radius: 8px;
		font-size: 13px;
		color: #555;
		cursor: pointer;
		transition: all 0.15s;
	}

	.custom-connector-btn:hover {
		background: #f5f5f5;
		border-color: #bbb;
	}

	.mcp-badge {
		padding: 2px 6px;
		background: linear-gradient(135deg, #8B7355 0%, #A0826D 100%);
		color: white;
		border-radius: 4px;
		font-size: 10px;
		font-weight: 600;
		letter-spacing: 0.5px;
	}

	.integrations-grid {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 16px;
		padding: 12px 32px 32px;
		overflow-y: auto;
		flex: 1;
		background: white;
	}

	.integration-card {
		position: relative;
		background: white;
		border: 1px solid #e8e8e8;
		border-radius: 12px;
		padding: 18px;
		transition: all 0.2s;
	}

	.integration-card:hover:not(.coming-soon) {
		border-color: #d0d0d0;
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.06);
		z-index: 20;
	}

	.integration-card.coming-soon {
		opacity: 0.6;
	}

	.card-tooltip {
		position: absolute;
		bottom: -8px;
		left: 50%;
		transform: translateX(-50%) translateY(100%);
		background: #333;
		color: white;
		padding: 10px 14px;
		border-radius: 8px;
		font-size: 12px;
		max-width: 250px;
		text-align: center;
		z-index: 100;
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
		pointer-events: none;
	}

	.card-tooltip::after {
		content: '';
		position: absolute;
		top: -6px;
		left: 50%;
		transform: translateX(-50%);
		border-left: 6px solid transparent;
		border-right: 6px solid transparent;
		border-bottom: 6px solid #333;
	}

	.card-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 10px;
	}

	.app-info {
		display: flex;
		align-items: center;
		gap: 8px;
		flex-wrap: wrap;
	}

	.app-icon {
		width: 28px;
		height: 28px;
		border-radius: 6px;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 14px;
		font-weight: 600;
		color: white;
	}

	.app-icon.logo {
		background: transparent;
		padding: 2px;
	}

	.app-icon.logo img {
		width: 100%;
		height: 100%;
		object-fit: contain;
	}

	.app-name {
		font-size: 14px;
		font-weight: 600;
		color: #333;
	}

	.auto-sync-badge {
		display: flex;
		align-items: center;
		gap: 4px;
		padding: 3px 8px;
		background: #f5f5f5;
		border-radius: 4px;
		font-size: 11px;
		color: #666;
	}

	.status-badge {
		padding: 5px 12px;
		border-radius: 20px;
		font-size: 12px;
		font-weight: 500;
	}

	.status-badge.connected {
		display: flex;
		align-items: center;
		gap: 6px;
		background: white;
		color: #166534;
		border: 1px solid #e5e5e5;
	}

	.status-dot {
		width: 6px;
		height: 6px;
		background: #22c55e;
		border-radius: 50%;
	}

	.status-badge.coming-soon {
		background: #f0f0f0;
		color: #999;
	}

	.status-badge.connecting {
		display: flex;
		align-items: center;
		gap: 6px;
		background: #eef2ff;
		color: #4f46e5;
		border: 1px solid #e0e7ff;
	}

	.spinner {
		width: 12px;
		height: 12px;
		border: 2px solid #c7d2fe;
		border-top-color: #4f46e5;
		border-radius: 50%;
		animation: spin 0.8s linear infinite;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}

	.integration-card.connecting {
		border-color: #c7d2fe;
		background: linear-gradient(135deg, #fefefe 0%, #f5f8ff 100%);
	}

	.connect-btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.connect-btn {
		padding: 6px 16px;
		background: #1a1a1a;
		color: white;
		border: none;
		border-radius: 20px;
		font-size: 13px;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.15s;
	}

	.connect-btn:hover {
		background: #333;
	}

	.card-description {
		font-size: 13px;
		color: #666;
		line-height: 1.5;
		margin: 0 0 14px;
	}

	.card-stats {
		display: flex;
		flex-direction: column;
		gap: 6px;
		padding-top: 12px;
		border-top: 1px solid #f0f0f0;
	}

	.stat {
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 12px;
		color: #888;
	}

	.stat-label {
		flex: 1;
	}

	.stat-value {
		font-weight: 500;
		color: #555;
	}

	@media (max-width: 900px) {
		.integrations-grid {
			grid-template-columns: repeat(2, 1fr);
		}

		.privacy-banner {
			flex-wrap: wrap;
		}

		.privacy-text {
			order: 3;
			width: 100%;
			margin-top: 8px;
		}
	}

	@media (max-width: 600px) {
		.integrations-grid {
			grid-template-columns: 1fr;
		}

		.section-header {
			flex-direction: column;
			align-items: flex-start;
			gap: 12px;
		}
	}
</style>
