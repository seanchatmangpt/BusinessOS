<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { fade, slide } from 'svelte/transition';
	import { useSession, clearSession } from '$lib/auth-client';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { browser } from '$app/environment';
	import * as integrationsApi from '$lib/api/integrations';
	import type {
		IntegrationProviderInfo,
		UserIntegration,
		AIModelPreferences,
		PendingDecision,
		IntegrationCategory
	} from '$lib/api/integrations';

	const session = useSession();

	// State
	let isLoading = $state(true);
	let activeTab = $state<'connected' | 'available' | 'ai' | 'mcp' | 'decisions'>('available');
	let hoveredId = $state<string | null>(null);
	let connectingId = $state<string | null>(null);
	let selectedProvider = $state<IntegrationProviderInfo | null>(null);
	let showDetailModal = $state(false);

	// Data
	let connectedIntegrations = $state<UserIntegration[]>([]);
	let availableProviders = $state<IntegrationProviderInfo[]>([]);
	let aiPreferences = $state<AIModelPreferences | null>(null);
	let pendingDecisions = $state<PendingDecision[]>([]);
	let selectedCategory = $state<IntegrationCategory | 'all'>('all');
	let isAuthenticated = $state(false);

	// Guards to prevent duplicate API calls
	let authDataLoading = $state(false);
	let authDataLoaded = $state(false);

	// File import state
	let showFileImportModal = $state(false);
	let fileImportProvider = $state<IntegrationProviderInfo | null>(null);
	let fileImportFile = $state<File | null>(null);
	let fileImporting = $state(false);
	let fileImportError = $state<string | null>(null);
	let fileImportSuccess = $state<string | null>(null);
	let fileInputRef = $state<HTMLInputElement | null>(null);

	// AI prefs save state
	let savingAiPrefs = $state(false);
	let aiPrefsMessage = $state<string | null>(null);
	let aiPrefsError = $state<string | null>(null);

	// Sync state for connected cards
	let syncingId = $state<string | null>(null);

	// Search filter for providers
	let searchQuery = $state('');

	// Decisions error state
	let decisionsError = $state<string | null>(null);

	// MCP state
	let mcpServers = $state<integrationsApi.MCPConnector[]>([]);
	let mcpLoading = $state(false);
	let showMCPModal = $state(false);
	let editingMCPServer = $state<integrationsApi.MCPConnector | null>(null);
	let testingMCPId = $state<string | null>(null);
	let deletingMCPId = $state<string | null>(null);
	let mcpError = $state<string | null>(null);
	let mcpSuccess = $state<string | null>(null);
	let expandedMCPId = $state<string | null>(null);
	let mcpForm = $state({ name: '', server_url: '', description: '', auth_type: 'none' as 'none' | 'api_key' | 'bearer', auth_token: '', transport: 'sse' });
	let mcpFormError = $state<string | null>(null);
	let mcpFormSaving = $state(false);

	// AI providers that use file import instead of OAuth
	const fileImportProviders = ['chatgpt', 'claude', 'perplexity', 'gemini', 'granola'];

	// Cleanup for OAuth message listener
	let oauthMessageCleanup: (() => void) | null = null;
	onDestroy(() => oauthMessageCleanup?.());

	// Category icons and labels
	const categories: { id: IntegrationCategory | 'all'; label: string; icon: string }[] = [
		{ id: 'all', label: 'All', icon: 'M4 6h16M4 12h16M4 18h16' },
		{ id: 'communication', label: 'Communication', icon: 'M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z' },
		{ id: 'crm', label: 'CRM', icon: 'M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z' },
		{ id: 'tasks', label: 'Tasks', icon: 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4' },
		{ id: 'calendar', label: 'Calendar', icon: 'M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z' },
		{ id: 'storage', label: 'Storage', icon: 'M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4' },
		{ id: 'meetings', label: 'Meetings', icon: 'M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z' },
		{ id: 'ai', label: 'AI Assistants', icon: 'M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z' },
		{ id: 'custom', label: 'Custom', icon: 'M11 4a2 2 0 114 0v1a1 1 0 001 1h3a1 1 0 011 1v3a1 1 0 01-1 1h-1a2 2 0 100 4h1a1 1 0 011 1v3a1 1 0 01-1 1h-3a1 1 0 01-1-1v-1a2 2 0 10-4 0v1a1 1 0 01-1 1H7a1 1 0 01-1-1v-3a1 1 0 00-1-1H4a2 2 0 110-4h1a1 1 0 001-1V7a1 1 0 011-1h3a1 1 0 001-1V4z' },
		{ id: 'productivity', label: 'Productivity', icon: 'M13 10V3L4 14h7v7l9-11h-7z' }
	];

	// Category descriptions for integration info
	const categoryInfo: Record<string, { desc: string; features: string[] }> = {
		communication: { desc: 'Email and messaging integrations', features: ['Import conversations', 'Track threads', 'Send messages'] },
		crm: { desc: 'Customer relationship management', features: ['Sync contacts', 'Track deals', 'Manage pipelines'] },
		tasks: { desc: 'Task and project management', features: ['Sync tasks', 'Track progress', 'Bi-directional updates'] },
		calendar: { desc: 'Calendar and scheduling', features: ['Sync events', 'Track meetings', 'Auto-scheduling'] },
		storage: { desc: 'File storage and documents', features: ['Index files', 'Full-text search', 'Knowledge extraction'] },
		meetings: { desc: 'Video calls and recordings', features: ['Meeting summaries', 'Transcripts', 'Action items'] },
		ai: { desc: 'AI assistant exports', features: ['Import conversations', 'Knowledge extraction', 'Pattern learning'] },
		finance: { desc: 'Financial tools', features: ['Invoice tracking', 'Payment sync', 'Reports'] },
		code: { desc: 'Code repositories', features: ['PR tracking', 'Issue sync', 'Commit history'] }
	};

	// Sort and filter providers (ones with local logos first)
	let sortedProviders = $derived(
		[...availableProviders].sort((a, b) => {
			const aHasLocalLogo = a.icon_url?.startsWith('/logos/') ? 0 : 1;
			const bHasLocalLogo = b.icon_url?.startsWith('/logos/') ? 0 : 1;
			return aHasLocalLogo - bHasLocalLogo;
		})
	);

	// Filter providers by category and search
	let filteredProviders = $derived.by(() => {
		let result = selectedCategory === 'all'
			? sortedProviders
			: sortedProviders.filter((p) => p.category === selectedCategory);
		if (searchQuery.trim()) {
			const q = searchQuery.trim().toLowerCase();
			result = result.filter((p) => p.name.toLowerCase().includes(q));
		}
		return result;
	});

	// Check if provider is connected
	function isProviderConnected(providerId: string) {
		return connectedIntegrations.some(
			(i) => i.provider_id === providerId && i.status === 'connected'
		);
	}

	// Get connected integration for a provider
	function getConnectedIntegration(providerId: string) {
		return connectedIntegrations.find((i) => i.provider_id === providerId);
	}

	// Reactive auth check - updates when session changes
	$effect(() => {
		const sessionData = $session;
		// Only update if session is no longer pending
		if (!sessionData?.isPending) {
			isAuthenticated = !!sessionData?.data?.user;
		}
	});

	onMount(async () => {
		// Handle OAuth callback: if this page was opened in a popup after OAuth redirect
		if (browser) {
			const urlParams = new URLSearchParams(window.location.search);
			const connectedProvider = urlParams.get('connected');
			if (connectedProvider && window.opener) {
				// We're in the OAuth popup — notify parent and close
				window.opener.postMessage({ type: 'integration-connected', provider: connectedProvider }, window.location.origin);
				window.close();
				return;
			}
			if (connectedProvider) {
				// Direct navigation with ?connected= (popup blocked or manual redirect)
				activeTab = 'connected';
				// Clean up the URL
				const url = new URL(window.location.href);
				url.searchParams.delete('connected');
				window.history.replaceState({}, '', url.toString());
			}
		}

		// Listen for OAuth popup completion messages
		function handleOAuthMessage(event: MessageEvent) {
			if (event.origin !== window.location.origin) return;
			if (event.data?.type === 'integration-connected') {
				// Popup completed OAuth — refresh data
				loadData();
				activeTab = 'connected';
			}
		}
		window.addEventListener('message', handleOAuthMessage);
		oauthMessageCleanup = () => window.removeEventListener('message', handleOAuthMessage);

		// Always load providers immediately (public endpoint)
		loadProviders();

		// Wait for session to resolve (give it up to 2 seconds)
		let attempts = 0;
		while ($session?.isPending && attempts < 20) {
			await new Promise((r) => setTimeout(r, 100));
			attempts++;
		}

		// Session resolved - check auth and load data
		const sessionData = $session;
		isAuthenticated = !sessionData?.isPending && !!sessionData?.data?.user;

		// Only load authenticated data if user is logged in AND we haven't already
		if (isAuthenticated && !authDataLoaded && !authDataLoading) {
			await loadAuthenticatedData();
		}
		isLoading = false;
	});

	async function loadProviders() {
		try {
			const providers = await integrationsApi.getProviders();
			availableProviders = providers.providers || [];
		} catch {
			availableProviders = [];
		}
	}

	async function loadAuthenticatedData() {
		// Guard against duplicate calls
		if (authDataLoading || authDataLoaded) {
			return;
		}
		authDataLoading = true;

		let authFailed = false;

		// Fetch connected integrations
		try {
			const connected = await integrationsApi.getConnectedIntegrations();
			connectedIntegrations = connected.integrations || [];
		} catch (e: unknown) {
			connectedIntegrations = [];
			// Check if this is a 401 error (session expired/invalid)
			if (e instanceof Error && e.message.includes('401')) {
				authFailed = true;
			}
		}

		// If first call got 401, clear session and don't make more authenticated calls
		if (authFailed) {
			clearSession();
			isAuthenticated = false;
			authDataLoading = false;
			isLoading = false;
			return;
		}

		// Fetch AI preferences
		try {
			const prefs = await integrationsApi.getAIModelPreferences();
			aiPreferences = prefs.preferences;
		} catch {
			aiPreferences = null;
		}

		// Fetch pending decisions
		try {
			const decisions = await integrationsApi.getPendingDecisions();
			pendingDecisions = decisions.decisions || [];
			decisionsError = null;
		} catch {
			pendingDecisions = [];
			decisionsError = 'Failed to load pending decisions';
		}

		// Fetch MCP servers
		try {
			const mcpResult = await integrationsApi.getMCPConnectors();
			mcpServers = mcpResult.servers || [];
		} catch {
			mcpServers = [];
		}

		authDataLoading = false;
		authDataLoaded = true;
		isLoading = false;
	}

	async function loadData() {
		isLoading = true;
		// Reset guard to allow fresh load
		authDataLoaded = false;
		await loadProviders();
		if (isAuthenticated && !authDataLoading) {
			await loadAuthenticatedData();
		}
		isLoading = false;
	}

	function openProviderDetail(provider: IntegrationProviderInfo) {
		selectedProvider = provider;
		showDetailModal = true;
	}

	function closeDetailModal() {
		showDetailModal = false;
		selectedProvider = null;
	}

	async function saveAiPreferences(updates: Partial<AIModelPreferences>) {
		if (!aiPreferences) return;
		savingAiPrefs = true;
		aiPrefsError = null;
		aiPrefsMessage = null;
		try {
			await integrationsApi.updateAIModelPreferences({ ...aiPreferences, ...updates });
			Object.assign(aiPreferences, updates);
			aiPrefsMessage = 'Preferences saved';
			setTimeout(() => (aiPrefsMessage = null), 3000);
		} catch (err) {
			console.error('Failed to save AI preferences:', err);
			aiPrefsError = 'Failed to save preferences';
			setTimeout(() => (aiPrefsError = null), 5000);
		} finally {
			savingAiPrefs = false;
		}
	}

	async function handleSyncCard(integrationId: string) {
		syncingId = integrationId;
		try {
			await integrationsApi.triggerIntegrationSync(integrationId);
			await loadData();
		} catch (err) {
			console.error('Failed to sync:', err);
		} finally {
			syncingId = null;
		}
	}

	async function handleConnect(provider: IntegrationProviderInfo) {
		if (!isAuthenticated) {
			goto('/login');
			return;
		}

		// Check if this is a file import provider (AI assistants)
		if (fileImportProviders.includes(provider.id)) {
			fileImportProvider = provider;
			fileImportFile = null;
			fileImportError = null;
			fileImportSuccess = null;
			showFileImportModal = true;
			return;
		}

		// Use oauth_provider if available (maps provider to OAuth endpoint, e.g., google_calendar -> google)
		// Fall back to provider.id if not mapped
		const oauthProvider = provider.oauth_provider || provider.id;
		if (!oauthProvider) {
			alert(`OAuth not configured for ${provider.name}. Please try again later.`);
			return;
		}

		if (import.meta.env.DEV) console.log(`[handleConnect] Provider: ${provider.id}, OAuth Provider: ${oauthProvider}`);

		connectingId = provider.id;

		try {
			const response = await integrationsApi.initiateAuth(oauthProvider as integrationsApi.IntegrationProvider);
			if (response.auth_url) {
				window.open(response.auth_url, '_blank', 'width=600,height=700');
			}
		} catch (err) {
			console.error('Failed to initiate auth:', err);
			alert(`Failed to connect to ${provider.name}. Please try again.`);
		} finally {
			connectingId = null;
		}
	}

	async function handleDisconnect(integrationId: string) {
		try {
			await integrationsApi.disconnectUserIntegration(integrationId);
			await loadData();
		} catch (err) {
			console.error('Failed to disconnect:', err);
		}
	}

	async function handleDecision(decisionId: string, decision: string) {
		try {
			await integrationsApi.respondToDecision(decisionId, { decision });
			await loadData();
		} catch (err) {
			console.error('Failed to respond to decision:', err);
		}
	}

	function handleFileSelect(e: Event) {
		const target = e.target as HTMLInputElement;
		if (target.files && target.files.length > 0) {
			fileImportFile = target.files[0];
			fileImportError = null;
		}
	}

	async function handleFileImport() {
		if (!fileImportFile || !fileImportProvider) return;
		fileImporting = true;
		fileImportError = null;
		fileImportSuccess = null;
		try {
			const source = (fileImportProvider.id === 'granola' ? 'other' : fileImportProvider.id) as 'chatgpt' | 'claude' | 'perplexity' | 'gemini' | 'other';
			const result = await integrationsApi.importFile(fileImportFile, source);
			fileImportSuccess = result.message || `Successfully imported ${result.imported_count} items.`;
			fileImportFile = null;
			await loadData();
		} catch (err) {
			fileImportError = err instanceof Error ? err.message : 'Import failed. Please try again.';
		} finally {
			fileImporting = false;
		}
	}

	function closeFileImportModal() {
		showFileImportModal = false;
		fileImportProvider = null;
		fileImportFile = null;
		fileImportError = null;
		fileImportSuccess = null;
	}

	function getStatusBadgeClass(status: string) {
		switch (status) {
			case 'connected': return 'ih-badge--connected';
			case 'available': return 'ih-badge--available';
			case 'coming_soon': return 'ih-badge--coming-soon';
			case 'error': return 'ih-badge--error';
			default: return 'ih-badge--default';
		}
	}

	function getPriorityBadgeClass(priority: string) {
		switch (priority) {
			case 'urgent': return 'ih-priority--urgent';
			case 'high': return 'ih-priority--high';
			case 'medium': return 'ih-priority--medium';
			default: return 'ih-priority--default';
		}
	}

	// ── MCP Handlers ──
	function openMCPModal(server?: integrationsApi.MCPConnector) {
		if (server) {
			editingMCPServer = server;
			mcpForm = { name: server.name, server_url: server.server_url, description: server.description || '', auth_type: server.auth_type, auth_token: '', transport: server.transport || 'sse' };
		} else {
			editingMCPServer = null;
			mcpForm = { name: '', server_url: '', description: '', auth_type: 'none', auth_token: '', transport: 'sse' };
		}
		mcpFormError = null;
		showMCPModal = true;
	}

	function closeMCPModal() {
		showMCPModal = false;
		editingMCPServer = null;
		mcpFormError = null;
	}

	function validateMCPForm(): string | null {
		if (!mcpForm.name.trim()) return 'Name is required';
		if (!/^[a-z0-9][a-z0-9-]*$/.test(mcpForm.name)) return 'Name must be lowercase alphanumeric with hyphens';
		if (!mcpForm.server_url.trim()) return 'Server URL is required';
		try { new URL(mcpForm.server_url); } catch { return 'Invalid URL format'; }
		if (mcpForm.auth_type !== 'none' && !editingMCPServer && !mcpForm.auth_token.trim()) return 'Auth token is required for this auth type';
		return null;
	}

	async function handleMCPSubmit() {
		const err = validateMCPForm();
		if (err) { mcpFormError = err; return; }
		mcpFormSaving = true;
		mcpFormError = null;
		try {
			const data: integrationsApi.CreateMCPConnectorData = {
				name: mcpForm.name.trim(),
				server_url: mcpForm.server_url.trim(),
				description: mcpForm.description.trim() || undefined,
				auth_type: mcpForm.auth_type,
				transport: mcpForm.transport,
				auth_token: mcpForm.auth_token.trim() || undefined,
			};
			if (editingMCPServer) {
				await integrationsApi.updateMCPConnector(editingMCPServer.id, data);
				mcpSuccess = 'Server updated';
			} else {
				await integrationsApi.createMCPConnector(data);
				mcpSuccess = 'Server added';
			}
			closeMCPModal();
			await loadMCPServers();
			setTimeout(() => (mcpSuccess = null), 3000);
		} catch (e) {
			mcpFormError = e instanceof Error ? e.message : 'Failed to save server';
		} finally {
			mcpFormSaving = false;
		}
	}

	async function loadMCPServers() {
		mcpLoading = true;
		try {
			const result = await integrationsApi.getMCPConnectors();
			mcpServers = result.servers || [];
		} catch {
			mcpServers = [];
		} finally {
			mcpLoading = false;
		}
	}

	async function handleTestMCP(id: string) {
		testingMCPId = id;
		mcpError = null;
		try {
			const result = await integrationsApi.testMCPConnector(id);
			if (result.success) {
				mcpSuccess = `Connection successful — ${result.tools_count ?? 0} tools discovered`;
				await loadMCPServers();
			} else {
				mcpError = result.message || 'Connection test failed';
			}
			setTimeout(() => { mcpSuccess = null; mcpError = null; }, 5000);
		} catch (e) {
			mcpError = e instanceof Error ? e.message : 'Test failed';
			setTimeout(() => (mcpError = null), 5000);
		} finally {
			testingMCPId = null;
		}
	}

	async function handleDeleteMCP(id: string) {
		deletingMCPId = id;
		try {
			await integrationsApi.deleteMCPConnector(id);
			mcpSuccess = 'Server deleted';
			await loadMCPServers();
			setTimeout(() => (mcpSuccess = null), 3000);
		} catch (e) {
			mcpError = e instanceof Error ? e.message : 'Failed to delete';
			setTimeout(() => (mcpError = null), 5000);
		} finally {
			deletingMCPId = null;
		}
	}

	async function handleToggleMCP(server: integrationsApi.MCPConnector) {
		try {
			await integrationsApi.updateMCPConnector(server.id, { name: server.name, server_url: server.server_url, auth_type: server.auth_type });
			await loadMCPServers();
		} catch (e) {
			mcpError = e instanceof Error ? e.message : 'Failed to toggle';
			setTimeout(() => (mcpError = null), 5000);
		}
	}

	function getMCPStatusClass(status: string) {
		switch (status) {
			case 'connected': return 'mcp-status--connected';
			case 'error': return 'mcp-status--error';
			default: return 'mcp-status--disconnected';
		}
	}
</script>

<svelte:head>
	<title>Integrations | BusinessOS</title>
</svelte:head>

<div class="ih-page">
	<!-- Header -->
	<div class="ih-header">
		<div class="ih-header__inner">
			<div class="ih-header__top">
				<div>
					<h1 class="ih-header__title">Integrations</h1>
					<p class="ih-header__subtitle">
						Connect your favorite tools and configure AI models
					</p>
				</div>
				{#if pendingDecisions.length > 0}
					<button
						onclick={() => (activeTab = 'decisions')}
						class="ih-decisions-alert"
					>
						<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
						</svg>
						<span>{pendingDecisions.length} pending decisions</span>
					</button>
				{/if}
			</div>

			<!-- Tabs -->
			<div class="ih-tabs">
				<button
					onclick={() => (activeTab = 'connected')}
					class="ih-tab {activeTab === 'connected' ? 'ih-tab--active' : ''}"
				>
					Connected ({connectedIntegrations.length})
				</button>
				<button
					onclick={() => (activeTab = 'available')}
					class="ih-tab {activeTab === 'available' ? 'ih-tab--active' : ''}"
				>
					Available ({availableProviders.length})
				</button>
				<button
					onclick={() => (activeTab = 'ai')}
					class="ih-tab {activeTab === 'ai' ? 'ih-tab--active' : ''}"
				>
					AI Models
				</button>
				<button
					onclick={() => (activeTab = 'mcp')}
					class="ih-tab {activeTab === 'mcp' ? 'ih-tab--active' : ''}"
				>
					MCP Servers ({mcpServers.length})
				</button>
				<button
					onclick={() => (activeTab = 'decisions')}
					class="ih-tab {activeTab === 'decisions' ? 'ih-tab--active' : ''}"
				>
					Decisions
					{#if pendingDecisions.length > 0}
						<span class="ih-tab__count">{pendingDecisions.length}</span>
					{/if}
				</button>
			</div>
		</div>
	</div>

	<!-- Content -->
	<div class="ih-content">
		{#key activeTab}
		<div in:fade={{ duration: 150 }}>
		{#if isLoading}
			<div class="ih-spinner-wrap">
				<div class="ih-spinner"></div>
			</div>
		{:else if activeTab === 'connected'}
			<!-- Connected Integrations -->
			{#if connectedIntegrations.length === 0}
				<div class="ih-empty">
					<svg
						class="ih-empty__icon"
						fill="none"
						viewBox="0 0 24 24"
						stroke="currentColor"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"
						/>
					</svg>
					<h3 class="ih-empty__title">No integrations connected</h3>
					<p class="ih-empty__text">Connect your favorite tools to get started.</p>
					<button
						onclick={() => (activeTab = 'available')}
						class="btn-pill btn-pill-primary btn-pill-sm mt-4"
					>
						Browse Available Integrations
					</button>
				</div>
			{:else}
				<div class="ih-grid">
					{#each connectedIntegrations as integration}
						<div class="ih-card">
							<div class="ih-card__header">
								<div class="ih-card__icon-wrap">
									{#if integration.icon_url}
										<img
											src={integration.icon_url}
											alt={integration.provider_name}
											class="w-6 h-6"
										/>
									{:else}
										<span class="ih-card__icon-letter">
											{integration.provider_name.charAt(0)}
										</span>
									{/if}
								</div>
								<div class="ih-card__info">
									<div class="ih-card__name-row">
										<h3 class="ih-card__name">{integration.provider_name}</h3>
										<span class="ih-badge {getStatusBadgeClass(integration.status)}">
											{integration.status}
										</span>
									</div>
									<p class="ih-card__meta">
										{integration.external_account_name ||
											integration.external_workspace_name ||
											'Connected'}
									</p>
									{#if integration.last_used_at}
										<p class="ih-card__sub-meta">
											Last used {new Date(integration.last_used_at).toLocaleDateString()}
										</p>
									{/if}
								</div>
							</div>
							<div class="ih-card__actions">
								<button
									onclick={() => handleSyncCard(integration.id)}
									disabled={syncingId === integration.id}
									class="ih-card__sync-btn"
									title="Sync now"
								>
									{#if syncingId === integration.id}
										<svg class="w-4 h-4 ih-spinner--inline" viewBox="0 0 24 24">
											<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" fill="none" />
											<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
										</svg>
									{:else}
										<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
										</svg>
									{/if}
								</button>
								<a
									href="/integrations/{integration.id}"
									class="ih-card__settings-link"
									title="Configure"
								>
									<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
									</svg>
								</a>
								<button
									onclick={() => handleDisconnect(integration.id)}
									class="btn-pill btn-pill-danger btn-pill-sm ih-card__actions-btn"
								>
									Disconnect
								</button>
							</div>
						</div>
					{/each}
				</div>
			{/if}
		{:else if activeTab === 'available'}
			<!-- Header Text -->
			<div class="ih-section-intro">
				<h2 class="ih-section-intro__title">
					Let's bring all your data into a single place.
				</h2>
				<p class="ih-section-intro__text">
					When you connect your apps, we will process raw data and extract essential information and turn it into nodes.
				</p>
			</div>

			<!-- Category Filter -->
			<div class="ih-category-filter">
				<div class="ih-search-wrap">
					<svg class="w-4 h-4 ih-search-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
					</svg>
					<input
						type="text"
						placeholder="Search integrations..."
						bind:value={searchQuery}
						class="ih-search-input"
					/>
				</div>
				{#each categories as category}
					<button
						onclick={() => (selectedCategory = category.id)}
						class="ih-category-btn {selectedCategory === category.id ? 'ih-category-btn--active' : ''}"
					>
						<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={category.icon} />
						</svg>
						{category.label}
					</button>
				{/each}
			</div>

			<!-- Integrations Grid -->
			<div class="ih-grid ih-grid--pb">
				{#each filteredProviders as provider}
					{@const isConnected = isProviderConnected(provider.id)}
					{@const isConnecting = connectingId === provider.id}
					{@const isComingSoon = provider.status === 'coming_soon'}
					<div
						class="ih-provider-card
							{isComingSoon ? 'ih-provider-card--soon' : ''}
							{isConnecting ? 'ih-provider-card--connecting' : ''}"
						onmouseenter={() => hoveredId = provider.id}
						onmouseleave={() => hoveredId = null}
					>
						<!-- Tooltip -->
						{#if hoveredId === provider.id && provider.tooltip && !isComingSoon}
							<div
								class="ih-tooltip"
								transition:fade={{ duration: 150 }}
							>
								{provider.tooltip}
								<div class="ih-tooltip__arrow"></div>
							</div>
						{/if}

						<!-- Card Header -->
						<div class="ih-provider-card__header">
							<div class="ih-provider-card__left">
								<!-- Icon -->
								<div class="ih-provider-card__icon">
									{#if provider.icon_url}
										<img
											src={provider.icon_url}
											alt={provider.name}
											class="w-5 h-5 object-contain"
											onerror={(e) => { const target = e.currentTarget as HTMLImageElement; target.style.display = 'none'; target.nextElementSibling?.classList.remove('hidden'); }}
										/>
										<span class="hidden ih-card__icon-letter--sm">
											{provider.name.charAt(0)}
										</span>
									{:else}
										<span class="ih-card__icon-letter--sm">
											{provider.name.charAt(0)}
										</span>
									{/if}
								</div>
								<!-- Name -->
								<span class="ih-provider-card__name">{provider.name}</span>
								<!-- Auto Live-sync Badge -->
								{#if provider.auto_live_sync}
									<span class="ih-autosync-badge">
										Auto Live-sync
										<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
										</svg>
									</span>
								{/if}
							</div>

							<!-- Status / Connect Button -->
							{#if isConnected}
								<span class="ih-status-pill ih-status-pill--connected">
									<span class="ih-status-dot ih-status-dot--green"></span>
									Live-Synced
								</span>
							{:else if isComingSoon}
								<span class="ih-status-pill ih-status-pill--soon">Soon</span>
							{:else if isConnecting}
								<span class="ih-status-pill ih-status-pill--connecting">
									<span class="ih-spinner ih-spinner--sm"></span>
									Connecting...
								</span>
							{:else}
								<button
									onclick={() => handleConnect(provider)}
									class="btn-pill btn-pill-primary btn-pill-sm"
								>
									{fileImportProviders.includes(provider.id) ? 'Import' : 'Connect'}
								</button>
							{/if}
						</div>

						<!-- Description -->
						<p class="ih-provider-card__desc">
							{provider.description || `Connect your ${provider.name} account`}
						</p>

						<!-- Stats Footer -->
						{#if provider.est_nodes || provider.initial_sync}
							<div class="ih-provider-card__stats">
								{#if provider.est_nodes}
									<div class="ih-stat-row">
										<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
										</svg>
										<span class="ih-stat-row__label">{isConnected ? 'Tot. nodes' : 'Est. nodes'}</span>
										<span class="ih-stat-row__value">{provider.est_nodes}</span>
									</div>
								{/if}
								{#if provider.initial_sync}
									<div class="ih-stat-row">
										<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
										</svg>
										<span class="ih-stat-row__label">Initial sync</span>
										<span class="ih-stat-row__value">{provider.initial_sync}</span>
									</div>
								{/if}
							</div>
						{/if}

						<!-- Learn More Link -->
						<button
							onclick={() => openProviderDetail(provider)}
							class="ih-learn-more"
						>
							Learn more
							<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
							</svg>
						</button>
					</div>
				{/each}
			</div>
		{:else if activeTab === 'ai'}
			<!-- AI Model Preferences -->
			<div class="ih-section">
				<h2 class="ih-section__title">AI Model Configuration</h2>
				<p class="ih-section__desc">
					Configure which AI models to use for different task tiers. The system automatically selects
					the appropriate tier based on task complexity.
				</p>

				{#if aiPreferences}
					<div class="ih-tier-list">
						<!-- Tier 2 -->
						<div class="ih-tier">
							<h3 class="ih-tier__name">Tier 2: Fast Tasks</h3>
							<p class="ih-tier__desc">
								Quick, low-complexity operations like formatting and simple lookups.
							</p>
							<div class="ih-tier__model">
								<span>
									{aiPreferences.tier_2_model.provider}: {aiPreferences.tier_2_model.model_id}
								</span>
							</div>
						</div>

						<!-- Tier 3 -->
						<div class="ih-tier">
							<h3 class="ih-tier__name">Tier 3: Standard Tasks</h3>
							<p class="ih-tier__desc">
								Medium-complexity tasks requiring analysis and synthesis.
							</p>
							<div class="ih-tier__model">
								<span>
									{aiPreferences.tier_3_model.provider}: {aiPreferences.tier_3_model.model_id}
								</span>
							</div>
						</div>

						<!-- Tier 4 -->
						<div class="ih-tier">
							<h3 class="ih-tier__name">Tier 4: Complex Tasks</h3>
							<p class="ih-tier__desc">
								High-complexity tasks requiring deep reasoning and multi-step analysis.
							</p>
							<div class="ih-tier__model">
								<span>
									{aiPreferences.tier_4_model.provider}: {aiPreferences.tier_4_model.model_id}
								</span>
							</div>
						</div>

						<!-- Settings -->
						<div class="ih-ai-settings">
							<h3 class="ih-ai-settings__title">Settings</h3>
							{#if aiPrefsMessage}
								<div class="ih-alert ih-alert--success ih-alert--sm">
									<p>{aiPrefsMessage}</p>
								</div>
							{/if}
							{#if aiPrefsError}
								<div class="ih-alert ih-alert--error ih-alert--sm">
									<p>{aiPrefsError}</p>
								</div>
							{/if}
							<div class="ih-ai-settings__list">
								<label class="ih-checkbox-label">
									<input
										type="checkbox"
										checked={aiPreferences.allow_model_upgrade_on_failure}
										onchange={(e) => {
											const target = e.currentTarget as HTMLInputElement;
											saveAiPreferences({ allow_model_upgrade_on_failure: target.checked });
										}}
										disabled={savingAiPrefs}
										class="ih-checkbox"
									/>
									<span>Allow automatic model upgrade on failure</span>
								</label>
								<label class="ih-checkbox-label">
									<input
										type="checkbox"
										checked={aiPreferences.prefer_local}
										onchange={(e) => {
											const target = e.currentTarget as HTMLInputElement;
											saveAiPreferences({ prefer_local: target.checked });
										}}
										disabled={savingAiPrefs}
										class="ih-checkbox"
									/>
									<span>Prefer local models when available</span>
								</label>
								<div class="ih-latency-row">
									<span>Max latency:</span>
									<input
										type="number"
										value={aiPreferences.max_latency_ms}
										onchange={(e) => {
											const target = e.currentTarget as HTMLInputElement;
											const val = parseInt(target.value, 10);
											if (!isNaN(val) && val > 0) {
												saveAiPreferences({ max_latency_ms: val });
											}
										}}
										disabled={savingAiPrefs}
										class="ih-latency-input"
										min="100"
										step="100"
									/>
									<span class="ih-latency-unit">ms</span>
								</div>
							</div>
						</div>
					</div>
				{:else}
					<p class="ih-empty__text">
						AI preferences not configured. Default settings will be used.
					</p>
				{/if}
			</div>
		{:else if activeTab === 'mcp'}
			<!-- MCP Servers -->
			<div class="mcp-section">
				<!-- Header -->
				<div class="mcp-header">
					<div>
						<h2 class="mcp-header__title">MCP Servers</h2>
						<p class="mcp-header__subtitle">Connect external tool servers so your AI agents can use them</p>
					</div>
					<button onclick={() => openMCPModal()} class="btn-pill btn-pill-primary btn-pill-sm">
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
						Add MCP Server
					</button>
				</div>

				{#if mcpError}
					<div class="ih-alert ih-alert--error ih-alert--banner">
						<p>{mcpError}</p>
					</div>
				{/if}
				{#if mcpSuccess}
					<div class="ih-alert ih-alert--success ih-alert--banner">
						<p>{mcpSuccess}</p>
					</div>
				{/if}

				{#if mcpLoading}
					<div class="ih-empty">
						<svg class="w-8 h-8 ih-spinner" fill="none" viewBox="0 0 24 24">
							<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
							<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
						</svg>
						<p class="ih-empty__text">Loading MCP servers...</p>
					</div>
				{:else if mcpServers.length === 0}
					<div class="ih-empty">
						<svg class="ih-empty__icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2" />
						</svg>
						<h3 class="ih-empty__title">No MCP servers connected</h3>
						<p class="ih-empty__text">Connect external MCP servers to give your AI agents access to additional tools and capabilities.</p>
						<button onclick={() => openMCPModal()} class="btn-pill btn-pill-primary btn-pill-sm" style="margin-top: 1rem;">
							Add Your First Server
						</button>
					</div>
				{:else}
					<div class="mcp-grid">
						{#each mcpServers as server (server.id)}
							<div class="ih-card mcp-card">
								<div class="mcp-card__top">
									<div class="mcp-card__info">
										<div class="mcp-card__name-row">
											<h3 class="ih-card__name">{server.name}</h3>
											<span class="mcp-status {getMCPStatusClass(server.status)}">{server.status}</span>
										</div>
										<p class="mcp-card__url" title={server.server_url}>{server.server_url}</p>
										{#if server.description}
											<p class="mcp-card__desc">{server.description}</p>
										{/if}
									</div>
									<div class="mcp-card__meta">
										<span class="mcp-card__tools-count">
											<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" /></svg>
											{server.tool_count} tools
										</span>
										{#if server.has_auth}
											<span class="mcp-card__auth-badge">
												<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" /></svg>
												Auth
											</span>
										{/if}
										<span class="mcp-card__transport">{server.transport}</span>
									</div>
								</div>

								<div class="mcp-card__actions">
									<button
										onclick={() => handleTestMCP(server.id)}
										disabled={testingMCPId === server.id}
										class="btn-pill btn-pill-ghost btn-pill-sm"
									>
										{#if testingMCPId === server.id}
											<svg class="w-4 h-4 ih-spinner--inline" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path></svg>
											Testing...
										{:else}
											Test
										{/if}
									</button>
									<button onclick={() => openMCPModal(server)} class="btn-pill btn-pill-ghost btn-pill-sm">Edit</button>
									<button
										onclick={() => handleDeleteMCP(server.id)}
										disabled={deletingMCPId === server.id}
										class="btn-pill btn-pill-ghost btn-pill-sm mcp-btn--danger"
									>
										{deletingMCPId === server.id ? 'Deleting...' : 'Delete'}
									</button>
								</div>

								<!-- Expandable tools list -->
								{#if server.tools && server.tools.length > 0}
									<button class="mcp-card__expand-btn" onclick={() => expandedMCPId = expandedMCPId === server.id ? null : server.id}>
										<svg class="w-4 h-4 mcp-chevron {expandedMCPId === server.id ? 'mcp-chevron--open' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" /></svg>
										{server.tools.length} available tools
									</button>
									{#if expandedMCPId === server.id}
										<div class="mcp-tools-list" transition:slide={{ duration: 150 }}>
											{#each server.tools as tool}
												<div class="mcp-tool-item">
													<span class="mcp-tool-item__name">{tool.name}</span>
													{#if tool.description}
														<span class="mcp-tool-item__desc">{tool.description}</span>
													{/if}
												</div>
											{/each}
										</div>
									{/if}
								{/if}
							</div>
						{/each}
					</div>
				{/if}
			</div>

		{:else if activeTab === 'decisions'}
			<!-- Pending Decisions -->
			{#if decisionsError}
				<div class="ih-alert ih-alert--error ih-alert--banner">
					<p>{decisionsError}</p>
					<button onclick={loadData} class="btn-pill btn-pill-ghost btn-pill-sm">Retry</button>
				</div>
			{/if}
			{#if pendingDecisions.length === 0 && !decisionsError}
				<div class="ih-empty">
					<svg
						class="ih-empty__icon"
						fill="none"
						viewBox="0 0 24 24"
						stroke="currentColor"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
						/>
					</svg>
					<h3 class="ih-empty__title">No pending decisions</h3>
					<p class="ih-empty__text">
						When AI agents need your input, decisions will appear here.
					</p>
				</div>
			{:else}
				<div class="ih-decision-list">
					{#each pendingDecisions as decision}
						<div class="ih-card">
							<div class="ih-decision__top">
								<div>
									<div class="ih-decision__header">
										<h3 class="ih-card__name">{decision.question}</h3>
										<span class="ih-priority {getPriorityBadgeClass(decision.priority)}">
											{decision.priority}
										</span>
									</div>
									{#if decision.description}
										<p class="ih-decision__desc">{decision.description}</p>
									{/if}
									<p class="ih-decision__meta">
										Skill: {decision.skill_id} | Created: {new Date(decision.created_at).toLocaleString()}
									</p>
								</div>
							</div>
							{#if decision.options && decision.options.length > 0}
								<div class="ih-decision__options">
									{#each decision.options as option}
										<button
											onclick={() => handleDecision(decision.id, option)}
											class="btn-pill btn-pill-primary btn-pill-sm"
										>
											{option}
										</button>
									{/each}
								</div>
							{/if}
						</div>
					{/each}
				</div>
			{/if}
		{/if}
		</div>
		{/key}
	</div>
</div>

<!-- Integration Detail Modal -->
{#if showDetailModal && selectedProvider}
	<div
		class="ih-modal-backdrop"
		onclick={closeDetailModal}
		transition:fade={{ duration: 200 }}
	>
		<div
			class="ih-modal"
			onclick={(e) => e.stopPropagation()}
			transition:slide={{ duration: 200 }}
		>
			<!-- Header -->
			<div class="ih-modal__header">
				<div class="ih-modal__header-inner">
					<div class="ih-modal__provider">
						<div class="ih-modal__provider-icon">
							{#if selectedProvider.icon_url}
								<img src={selectedProvider.icon_url} alt={selectedProvider.name} class="ih-modal__provider-img" />
							{:else}
								<span class="ih-modal__provider-letter">{selectedProvider.name.charAt(0)}</span>
							{/if}
						</div>
						<div>
							<h2 class="ih-modal__title">{selectedProvider.name}</h2>
							<span class="ih-modal__category-badge">
								{selectedProvider.category}
							</span>
						</div>
					</div>
					<button
						onclick={closeDetailModal}
						class="btn-pill btn-pill-ghost btn-pill-icon"
					>
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
						</svg>
					</button>
				</div>
			</div>

			<!-- Body -->
			<div class="ih-modal__body">
				<!-- Description -->
				<p class="ih-modal__desc">
					{selectedProvider.description || `Connect your ${selectedProvider.name} account to sync data and enable powerful automations.`}
				</p>

				<!-- Category Features -->
				{#if categoryInfo[selectedProvider.category]}
					<div class="ih-modal__section">
						<h3 class="ih-modal__section-title">What you can do</h3>
						<ul class="ih-feature-list">
							{#each categoryInfo[selectedProvider.category].features as feature}
								<li class="ih-feature-item">
									<svg class="w-4 h-4 ih-feature-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
									</svg>
									{feature}
								</li>
							{/each}
						</ul>
					</div>
				{/if}

				<!-- Sync Info -->
				<div class="ih-sync-panel">
					<h3 class="ih-modal__section-title">Sync details</h3>
					<div class="ih-sync-grid">
						{#if selectedProvider.auto_live_sync}
							<div class="ih-sync-item">
								<div class="ih-sync-icon ih-sync-icon--green">
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
									</svg>
								</div>
								<div>
									<div class="ih-sync-label">Sync type</div>
									<div class="ih-sync-value">Live sync</div>
								</div>
							</div>
						{:else}
							<div class="ih-sync-item">
								<div class="ih-sync-icon ih-sync-icon--blue">
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
									</svg>
								</div>
								<div>
									<div class="ih-sync-label">Sync type</div>
									<div class="ih-sync-value">Manual/Scheduled</div>
								</div>
							</div>
						{/if}

						{#if selectedProvider.est_nodes}
							<div class="ih-sync-item">
								<div class="ih-sync-icon ih-sync-icon--purple">
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
									</svg>
								</div>
								<div>
									<div class="ih-sync-label">Est. nodes</div>
									<div class="ih-sync-value">{selectedProvider.est_nodes}</div>
								</div>
							</div>
						{/if}

						{#if selectedProvider.initial_sync}
							<div class="ih-sync-item">
								<div class="ih-sync-icon ih-sync-icon--amber">
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
									</svg>
								</div>
								<div>
									<div class="ih-sync-label">Initial sync</div>
									<div class="ih-sync-value">{selectedProvider.initial_sync}</div>
								</div>
							</div>
						{/if}
					</div>
				</div>

				<!-- Skills -->
				{#if selectedProvider.skills && selectedProvider.skills.length > 0}
					<div class="ih-modal__section">
						<h3 class="ih-modal__section-title">Available skills</h3>
						<div class="ih-tag-list">
							{#each selectedProvider.skills as skill}
								<span class="ih-skill-tag">{skill}</span>
							{/each}
						</div>
					</div>
				{/if}

				<!-- Modules -->
				{#if selectedProvider.modules && selectedProvider.modules.length > 0}
					<div class="ih-modal__section">
						<h3 class="ih-modal__section-title">Works with</h3>
						<div class="ih-tag-list">
							{#each selectedProvider.modules as module}
								<span class="ih-module-tag">
									<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h7" />
									</svg>
									{module.replace('_', ' ')}
								</span>
							{/each}
						</div>
					</div>
				{/if}
			</div>

			<!-- Footer -->
			<div class="ih-modal__footer">
				{#if isProviderConnected(selectedProvider.id)}
					<div class="ih-modal__connected-row">
						<div class="ih-modal__connected-status">
							<span class="ih-status-dot--green"></span>
							<span class="ih-modal__connected-label">Connected</span>
							{#if getConnectedIntegration(selectedProvider.id)?.external_account_name}
								<span class="ih-modal__connected-account">as {getConnectedIntegration(selectedProvider.id)?.external_account_name}</span>
							{/if}
						</div>
						<div class="ih-modal__connected-actions">
							{#if selectedProvider.auto_live_sync}
								<label class="ih-toggle-label">
									<span class="ih-toggle-text">Auto-sync</span>
									<div class="ih-toggle">
										<input type="checkbox" class="sr-only peer" checked />
										<div class="ih-toggle__track"></div>
										<div class="ih-toggle__thumb"></div>
									</div>
								</label>
							{/if}
							<a
								href="/integrations/{getConnectedIntegration(selectedProvider.id)?.id}"
								class="btn-pill btn-pill-ghost btn-pill-sm"
							>
								Settings
							</a>
						</div>
					</div>
				{:else if selectedProvider.status === 'coming_soon'}
					<button
						disabled
						class="btn-pill btn-pill-soft btn-pill-sm ih-modal__full-btn"
					>
						Coming Soon
					</button>
				{:else}
					<button
						onclick={() => { if (selectedProvider) { closeDetailModal(); handleConnect(selectedProvider); } }}
						class="btn-pill btn-pill-primary ih-modal__full-btn"
					>
						{selectedProvider && fileImportProviders.includes(selectedProvider.id) ? 'Import Data' : 'Connect'}
					</button>
				{/if}
			</div>
		</div>
	</div>
{/if}

<!-- File Import Modal -->
{#if showFileImportModal && fileImportProvider}
	<div class="ih-modal-backdrop" transition:fade={{ duration: 150 }}>
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div
			class="ih-modal-backdrop__overlay"
			onclick={closeFileImportModal}
			onkeydown={(e) => { if (e.key === 'Escape') closeFileImportModal(); }}
		></div>
		<div
			class="ih-modal ih-modal--sm"
			transition:fade={{ duration: 150 }}
			role="dialog"
			aria-label="Import data from {fileImportProvider.name}"
		>
			<!-- Header -->
			<div class="ih-modal__header">
				<div class="ih-modal__header-inner">
					<div class="ih-modal__provider">
						{#if fileImportProvider.logo_url}
							<img src={fileImportProvider.logo_url} alt="" class="ih-import-icon" />
						{/if}
						<div>
							<h3 class="ih-modal__title ih-modal__title--sm">Import from {fileImportProvider.name}</h3>
							<p class="ih-modal__subtitle">Upload your exported data file</p>
						</div>
					</div>
					<button
						onclick={closeFileImportModal}
						class="btn-pill btn-pill-ghost btn-pill-icon"
						aria-label="Close import dialog"
					>
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
						</svg>
					</button>
				</div>
			</div>

			<!-- Body -->
			<div class="ih-modal__body">
				<p class="ih-modal__help-text">
					Export your data from {fileImportProvider.name} and upload the file here. Supported formats: JSON, ZIP.
				</p>

				<!-- File Drop Zone -->
				<label class="ih-dropzone {fileImportFile ? 'ih-dropzone--active' : ''}">
					<input
						bind:this={fileInputRef}
						type="file"
						accept=".json,.zip,.csv,.txt"
						class="hidden"
						onchange={handleFileSelect}
					/>
					{#if fileImportFile}
						<svg class="w-8 h-8 ih-dropzone__icon--success" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
						</svg>
						<span class="ih-dropzone__filename">{fileImportFile.name}</span>
						<span class="ih-dropzone__filesize">
							{(fileImportFile.size / 1024).toFixed(1)} KB
						</span>
					{:else}
						<svg class="w-8 h-8 ih-dropzone__icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
						</svg>
						<span class="ih-dropzone__label">Click to select a file</span>
						<span class="ih-dropzone__formats">JSON, ZIP, CSV, or TXT</span>
					{/if}
				</label>

				<!-- Error / Success Messages -->
				{#if fileImportError}
					<div class="ih-alert ih-alert--error">
						<p>{fileImportError}</p>
					</div>
				{/if}
				{#if fileImportSuccess}
					<div class="ih-alert ih-alert--success">
						<p>{fileImportSuccess}</p>
					</div>
				{/if}
			</div>

			<!-- Footer -->
			<div class="ih-modal__footer">
				<button
					onclick={closeFileImportModal}
					class="btn-pill btn-pill-ghost btn-pill-sm"
				>
					Cancel
				</button>
				<button
					onclick={handleFileImport}
					disabled={!fileImportFile || fileImporting}
					class="btn-pill btn-pill-primary btn-pill-sm"
				>
					{#if fileImporting}
						<span class="ih-import-loading">
							<svg class="w-4 h-4 ih-spinner--inline" fill="none" viewBox="0 0 24 24">
								<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
								<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
							</svg>
							Importing...
						</span>
					{:else}
						Import Data
					{/if}
				</button>
			</div>
		</div>
	</div>
{/if}

<!-- MCP Server Modal -->
{#if showMCPModal}
	<div class="ih-modal-backdrop" transition:fade={{ duration: 150 }}>
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div class="ih-modal-backdrop__overlay" onclick={closeMCPModal} onkeydown={(e) => { if (e.key === 'Escape') closeMCPModal(); }}></div>
		<div class="ih-modal ih-modal--sm" transition:fade={{ duration: 150 }} role="dialog" aria-label="{editingMCPServer ? 'Edit' : 'Add'} MCP Server">
			<div class="ih-modal__header">
				<div class="ih-modal__header-inner">
					<h3 class="ih-modal__title ih-modal__title--sm">{editingMCPServer ? 'Edit' : 'Add'} MCP Server</h3>
					<button onclick={closeMCPModal} class="btn-pill btn-pill-ghost btn-pill-icon" aria-label="Close">
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
					</button>
				</div>
			</div>
			<div class="ih-modal__body">
				{#if mcpFormError}
					<div class="ih-alert ih-alert--error ih-alert--sm"><p>{mcpFormError}</p></div>
				{/if}
				<div class="mcp-form">
					<label class="mcp-form__label">
						<span>Name <span class="mcp-form__required">*</span></span>
						<input type="text" bind:value={mcpForm.name} placeholder="my-mcp-server" class="mcp-form__input" disabled={!!editingMCPServer} />
						<span class="mcp-form__hint">Lowercase letters, numbers, hyphens</span>
					</label>
					<label class="mcp-form__label">
						<span>Server URL <span class="mcp-form__required">*</span></span>
						<input type="url" bind:value={mcpForm.server_url} placeholder="https://mcp.example.com/sse" class="mcp-form__input" />
					</label>
					<label class="mcp-form__label">
						<span>Description</span>
						<textarea bind:value={mcpForm.description} placeholder="What does this server provide?" class="mcp-form__textarea" rows="2"></textarea>
					</label>
					<div class="mcp-form__row">
						<label class="mcp-form__label">
							<span>Transport</span>
							<select bind:value={mcpForm.transport} class="mcp-form__select">
								<option value="sse">SSE (Server-Sent Events)</option>
							</select>
						</label>
						<label class="mcp-form__label">
							<span>Auth Type</span>
							<select bind:value={mcpForm.auth_type} class="mcp-form__select">
								<option value="none">None</option>
								<option value="api_key">API Key</option>
								<option value="bearer">Bearer Token</option>
							</select>
						</label>
					</div>
					{#if mcpForm.auth_type !== 'none'}
						<label class="mcp-form__label" transition:slide={{ duration: 150 }}>
							<span>Auth Token <span class="mcp-form__required">*</span></span>
							<input type="password" bind:value={mcpForm.auth_token} placeholder={editingMCPServer ? '(unchanged)' : 'Enter token'} class="mcp-form__input" />
						</label>
					{/if}
				</div>
			</div>
			<div class="ih-modal__footer">
				<button onclick={closeMCPModal} class="btn-pill btn-pill-ghost btn-pill-sm">Cancel</button>
				<button onclick={handleMCPSubmit} disabled={mcpFormSaving} class="btn-pill btn-pill-primary btn-pill-sm">
					{#if mcpFormSaving}
						<svg class="w-4 h-4 ih-spinner--inline" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path></svg>
						Saving...
					{:else}
						{editingMCPServer ? 'Update' : 'Add Server'}
					{/if}
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	/* ═══════════════════════════════════════════════════════════
	   INTEGRATIONS HUB — Foundation ih- Prefix System
	   ═══════════════════════════════════════════════════════════ */

	/* Page Layout */
	.ih-page {
		min-height: 100vh;
		overflow-y: auto;
		background: var(--dbg);
	}
	.ih-header {
		background: var(--dbg2);
		border-bottom: 1px solid var(--dbd);
		position: sticky;
		top: 0;
		z-index: 10;
	}
	.ih-header__inner {
		max-width: 76rem;
		margin: 0 auto;
		padding: var(--space-6) var(--space-8) 0;
	}
	.ih-header__top {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		margin-bottom: var(--space-4);
	}
	.ih-header__title {
		font-size: var(--text-2xl);
		font-weight: var(--font-bold);
		color: var(--dt);
		letter-spacing: -0.01em;
	}
	.ih-header__subtitle {
		font-size: var(--text-sm);
		color: var(--dt3);
		margin-top: var(--space-1);
	}
	.ih-decisions-alert {
		display: inline-flex;
		align-items: center;
		gap: var(--space-2);
		padding: var(--space-1) var(--space-3);
		border-radius: var(--radius-full);
		background: rgba(245, 158, 11, 0.1);
		color: var(--accent-orange);
		font-size: var(--text-xs);
		font-weight: var(--font-medium);
		cursor: pointer;
		border: none;
		transition: background 200ms ease;
	}
	.ih-decisions-alert:hover {
		background: rgba(245, 158, 11, 0.18);
	}

	/* Tabs */
	.ih-tabs {
		display: flex;
		gap: var(--space-1);
		margin: 0 calc(-1 * var(--space-8));
		padding: 0 var(--space-8);
	}
	.ih-tab {
		padding: var(--space-3) var(--space-4);
		font-size: var(--text-sm);
		font-weight: var(--font-medium);
		color: var(--dt3);
		border-bottom: 2px solid transparent;
		cursor: pointer;
		background: none;
		border-top: none;
		border-left: none;
		border-right: none;
		transition: color 200ms ease, border-color 200ms ease;
		display: flex;
		align-items: center;
		gap: var(--space-2);
		white-space: nowrap;
	}
	.ih-tab:hover {
		color: var(--dt2);
	}
	.ih-tab--active {
		color: var(--accent-blue);
		border-bottom-color: var(--accent-blue);
	}
	.ih-tab__count {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		min-width: 1.25rem;
		height: 1.25rem;
		padding: 0 var(--space-1);
		border-radius: var(--radius-full);
		background: rgba(74, 144, 226, 0.1);
		color: var(--accent-blue);
		font-size: var(--text-xs);
		font-weight: var(--font-semibold);
	}

	/* Content Area */
	.ih-content {
		max-width: 76rem;
		margin: 0 auto;
		padding: var(--space-6) var(--space-8);
	}
	.ih-spinner-wrap {
		display: flex;
		align-items: center;
		justify-content: center;
		padding: var(--space-12) 0;
	}
	.ih-spinner {
		width: 2rem;
		height: 2rem;
		border: 2px solid var(--dbd);
		border-top-color: var(--accent-blue);
		border-radius: 50%;
		animation: ih-spin 0.7s linear infinite;
	}
	.ih-spinner--sm {
		width: 1rem;
		height: 1rem;
		border-width: 2px;
		border-color: var(--dbd);
		border-top-color: var(--accent-blue);
		border-radius: 50%;
		animation: ih-spin 0.7s linear infinite;
	}
	.ih-spinner--inline {
		animation: ih-spin 0.7s linear infinite;
	}
	@keyframes ih-spin {
		to { transform: rotate(360deg); }
	}

	/* Empty States */
	.ih-empty {
		text-align: center;
		padding: var(--space-16) var(--space-4);
	}
	.ih-empty__icon {
		width: 3rem;
		height: 3rem;
		margin: 0 auto;
		color: var(--dt4);
	}
	.ih-empty__title {
		margin-top: var(--space-4);
		font-size: var(--text-lg);
		font-weight: var(--font-medium);
		color: var(--dt);
	}
	.ih-empty__text {
		margin-top: var(--space-2);
		color: var(--dt3);
		font-size: var(--text-sm);
		max-width: 28rem;
		margin-left: auto;
		margin-right: auto;
		line-height: 1.5;
	}

	/* Card Grid */
	.ih-grid {
		display: grid;
		grid-template-columns: 1fr;
		gap: var(--space-4);
	}
	@media (min-width: 768px) {
		.ih-grid { grid-template-columns: repeat(2, 1fr); }
	}
	@media (min-width: 1024px) {
		.ih-grid { grid-template-columns: repeat(3, 1fr); }
	}
	.ih-grid--pb {
		display: grid;
		grid-template-columns: 1fr;
		gap: var(--space-3);
		padding-bottom: var(--space-8);
	}
	@media (min-width: 768px) {
		.ih-grid--pb { grid-template-columns: repeat(2, 1fr); }
	}
	@media (min-width: 1024px) {
		.ih-grid--pb { grid-template-columns: repeat(3, 1fr); }
	}

	/* Connected Cards */
	.ih-card {
		background: var(--dbg2);
		border: 1px solid var(--dbd);
		border-radius: var(--radius-md);
		padding: var(--space-4);
		transition: border-color 200ms ease, box-shadow 200ms ease;
	}
	.ih-card:hover {
		border-color: var(--dbd2);
		box-shadow: var(--shadow-sm);
	}
	.ih-card__header {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
	}
	.ih-card__icon-wrap {
		width: 2.5rem;
		height: 2.5rem;
		border-radius: var(--radius-sm);
		background: var(--dbg3);
		display: flex;
		align-items: center;
		justify-content: center;
		overflow: hidden;
		flex-shrink: 0;
	}
	.ih-card__icon-wrap img {
		width: 1.5rem;
		height: 1.5rem;
		object-fit: contain;
	}
	.ih-card__icon-letter {
		font-size: var(--text-sm);
		font-weight: var(--font-bold);
		color: var(--dt3);
	}
	.ih-card__icon-letter--sm {
		font-size: var(--text-xs);
		font-weight: var(--font-bold);
		color: var(--dt3);
	}
	.ih-card__info {
		margin-left: var(--space-3);
		flex: 1;
		min-width: 0;
	}
	.ih-card__name-row {
		display: flex;
		align-items: center;
		gap: var(--space-2);
	}
	.ih-card__name {
		font-weight: var(--font-medium);
		color: var(--dt);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		font-size: var(--text-sm);
	}
	.ih-card__meta {
		font-size: var(--text-xs);
		color: var(--dt4);
		margin-top: var(--space-1);
	}
	.ih-card__actions {
		display: flex;
		align-items: center;
		gap: var(--space-2);
		margin-left: var(--space-2);
	}
	.ih-card__actions-btn {
		padding: var(--space-1);
		border-radius: var(--radius-xs);
		color: var(--dt4);
		background: none;
		border: none;
		cursor: pointer;
		transition: color 200ms ease, background 200ms ease;
	}
	.ih-card__actions-btn:hover {
		color: var(--dt2);
		background: var(--dbg3);
	}
	.ih-card__settings-link {
		font-size: var(--text-xs);
		color: var(--accent-blue);
		text-decoration: none;
		display: inline-flex;
		align-items: center;
		gap: var(--space-1);
		margin-top: var(--space-2);
	}
	.ih-card__settings-link:hover {
		text-decoration: underline;
	}

	/* Badges */
	.ih-badge {
		display: inline-flex;
		padding: 2px var(--space-2);
		border-radius: var(--radius-full);
		font-size: 0.6875rem;
		font-weight: var(--font-medium);
		white-space: nowrap;
		letter-spacing: 0.01em;
	}
	.ih-badge--connected {
		background: rgba(16, 185, 129, 0.1);
		color: var(--accent-green);
	}
	.ih-badge--available {
		background: rgba(74, 144, 226, 0.1);
		color: var(--accent-blue);
	}
	.ih-badge--coming-soon {
		background: rgba(156, 163, 175, 0.1);
		color: var(--dt4);
	}
	.ih-badge--error {
		background: rgba(220, 38, 38, 0.1);
		color: var(--color-error);
	}
	.ih-badge--default {
		background: rgba(156, 163, 175, 0.1);
		color: var(--dt3);
	}

	/* Priority Badges */
	.ih-priority {
		display: inline-flex;
		padding: 2px var(--space-2);
		border-radius: var(--radius-full);
		font-size: 0.6875rem;
		font-weight: var(--font-medium);
	}
	.ih-priority--urgent {
		background: rgba(220, 38, 38, 0.1);
		color: var(--color-error);
	}
	.ih-priority--high {
		background: rgba(249, 115, 22, 0.1);
		color: var(--accent-orange);
	}
	.ih-priority--medium {
		background: rgba(245, 158, 11, 0.1);
		color: var(--accent-orange);
	}
	.ih-priority--default {
		background: rgba(156, 163, 175, 0.1);
		color: var(--dt3);
	}

	/* Section Intro (Available tab) */
	.ih-section-intro {
		margin-bottom: var(--space-5);
	}
	.ih-section-intro__title {
		font-size: var(--text-lg);
		font-weight: var(--font-semibold);
		color: var(--dt);
		letter-spacing: -0.01em;
	}
	.ih-section-intro__text {
		font-size: var(--text-sm);
		color: var(--dt3);
		margin-top: var(--space-1);
		line-height: 1.5;
	}

	/* Category Filter */
	.ih-category-filter {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: var(--space-2);
		margin-bottom: var(--space-5);
		padding-bottom: var(--space-4);
		border-bottom: 1px solid var(--dbd2);
	}
	.ih-category-btn {
		padding: var(--space-1) var(--space-3);
		border-radius: var(--radius-full);
		font-size: var(--text-xs);
		font-weight: var(--font-medium);
		border: 1px solid var(--dbd);
		background: var(--dbg);
		color: var(--dt3);
		cursor: pointer;
		transition: all 200ms ease;
		text-transform: capitalize;
		display: inline-flex;
		align-items: center;
		gap: var(--space-1);
	}
	.ih-category-btn:hover {
		border-color: var(--dbd);
		color: var(--dt2);
		background: var(--dbg2);
	}
	.ih-category-btn--active {
		background: rgba(74, 144, 226, 0.1);
		color: var(--accent-blue);
		border-color: rgba(74, 144, 226, 0.3);
	}

	/* Provider Cards */
	.ih-provider-card {
		background: var(--dbg2);
		border: 1px solid var(--dbd);
		border-radius: var(--radius-md);
		padding: var(--space-4);
		cursor: pointer;
		transition: border-color 200ms ease, box-shadow 200ms ease;
		position: relative;
	}
	.ih-provider-card:hover {
		border-color: var(--dbd2);
		box-shadow: var(--shadow-sm);
	}
	.ih-provider-card--soon {
		opacity: 0.6;
	}
	.ih-provider-card--connecting {
		border-color: rgba(74, 144, 226, 0.3);
	}
	.ih-provider-card__header {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		margin-bottom: var(--space-3);
	}
	.ih-provider-card__left {
		display: flex;
		align-items: center;
		gap: var(--space-3);
		flex: 1;
		min-width: 0;
	}
	.ih-provider-card__icon {
		width: 2.25rem;
		height: 2.25rem;
		border-radius: var(--radius-sm);
		background: var(--dbg3);
		display: flex;
		align-items: center;
		justify-content: center;
		overflow: hidden;
		flex-shrink: 0;
	}
	.ih-provider-card__icon img {
		width: 1.25rem;
		height: 1.25rem;
		object-fit: contain;
	}
	.ih-provider-card__name {
		font-size: var(--text-sm);
		font-weight: var(--font-semibold);
		color: var(--dt);
	}
	.ih-provider-card__desc {
		font-size: var(--text-xs);
		color: var(--dt3);
		margin-bottom: var(--space-3);
		display: -webkit-box;
		-webkit-line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
		line-height: 1.5;
	}
	.ih-provider-card__stats {
		display: flex;
		flex-direction: column;
		gap: var(--space-2);
		padding-top: var(--space-3);
		border-top: 1px solid var(--dbd2);
	}

	/* Status Pills */
	.ih-status-pill {
		display: inline-flex;
		align-items: center;
		gap: var(--space-1);
		font-size: 0.6875rem;
		font-weight: var(--font-medium);
		padding: 2px var(--space-2);
		border-radius: var(--radius-full);
	}
	.ih-status-pill--connected {
		background: rgba(16, 185, 129, 0.1);
		color: var(--accent-green);
	}
	.ih-status-pill--soon {
		background: rgba(156, 163, 175, 0.1);
		color: var(--dt4);
	}
	.ih-status-pill--connecting {
		background: rgba(74, 144, 226, 0.1);
		color: var(--accent-blue);
	}
	.ih-status-dot--green {
		width: 0.5rem;
		height: 0.5rem;
		border-radius: 50%;
		background: var(--accent-green);
		display: inline-block;
	}

	/* Auto-sync Badge */
	.ih-autosync-badge {
		display: inline-flex;
		align-items: center;
		gap: var(--space-1);
		font-size: 0.625rem;
		font-weight: var(--font-medium);
		padding: 2px var(--space-1);
		border-radius: var(--radius-xs);
		background: rgba(168, 85, 247, 0.1);
		color: var(--accent-purple);
	}

	/* Stat Row */
	.ih-stat-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		font-size: var(--text-xs);
	}
	.ih-stat-row__label {
		color: var(--dt4);
	}
	.ih-stat-row__value {
		color: var(--dt2);
		font-weight: var(--font-medium);
	}

	/* Learn More Link */
	.ih-learn-more {
		display: inline-flex;
		align-items: center;
		gap: var(--space-1);
		font-size: var(--text-xs);
		color: var(--accent-blue);
		text-decoration: none;
		margin-top: var(--space-2);
		font-weight: var(--font-medium);
		background: none;
		border: none;
		cursor: pointer;
		padding: 0;
	}
	.ih-learn-more:hover {
		text-decoration: underline;
	}

	/* Tooltip */
	.ih-tooltip {
		position: absolute;
		bottom: 100%;
		left: 50%;
		transform: translateX(-50%);
		margin-bottom: var(--space-2);
		padding: var(--space-1) var(--space-3);
		background: var(--dbg3);
		color: var(--dt2);
		font-size: var(--text-xs);
		border-radius: var(--radius-xs);
		white-space: nowrap;
		pointer-events: none;
		z-index: 10;
		box-shadow: var(--shadow-sm);
	}
	.ih-tooltip__arrow {
		position: absolute;
		top: 100%;
		left: 50%;
		transform: translateX(-50%);
		width: 0;
		height: 0;
		border-left: 4px solid transparent;
		border-right: 4px solid transparent;
		border-top: 4px solid var(--dbg3);
	}

	/* AI Models Section */
	.ih-section {
		background: var(--dbg2);
		border: 1px solid var(--dbd);
		border-radius: var(--radius-md);
		padding: var(--space-6);
		margin-bottom: var(--space-6);
	}
	.ih-section__title {
		font-size: var(--text-lg);
		font-weight: var(--font-semibold);
		color: var(--dt);
		margin-bottom: var(--space-1);
	}
	.ih-section__desc {
		font-size: var(--text-sm);
		color: var(--dt3);
		margin-bottom: var(--space-4);
		line-height: 1.5;
	}
	.ih-tier-list {
		display: flex;
		flex-direction: column;
		gap: var(--space-3);
	}
	.ih-tier {
		background: var(--dbg);
		border: 1px solid var(--dbd);
		border-radius: var(--radius-sm);
		padding: var(--space-4);
	}
	.ih-tier__name {
		font-weight: var(--font-semibold);
		color: var(--dt);
		text-transform: capitalize;
		margin-bottom: var(--space-1);
		font-size: var(--text-sm);
	}
	.ih-tier__desc {
		font-size: var(--text-sm);
		color: var(--dt3);
		margin-bottom: var(--space-2);
		line-height: 1.5;
	}
	.ih-tier__model {
		font-size: var(--text-xs);
		font-family: 'SF Mono', 'Fira Code', monospace;
		color: var(--dt4);
		background: var(--dbg2);
		padding: var(--space-2) var(--space-3);
		border-radius: var(--radius-xs);
		border: 1px solid var(--dbd2);
	}
	.ih-ai-settings {
		margin-top: var(--space-4);
	}
	.ih-ai-settings__title {
		font-size: var(--text-sm);
		font-weight: var(--font-semibold);
		color: var(--dt);
		margin-bottom: var(--space-3);
	}
	.ih-ai-settings__list {
		display: flex;
		flex-direction: column;
		gap: var(--space-3);
	}
	.ih-checkbox-label {
		display: flex;
		align-items: center;
		gap: var(--space-3);
		font-size: var(--text-sm);
		color: var(--dt2);
		cursor: pointer;
	}
	.ih-checkbox {
		width: 1rem;
		height: 1rem;
		border-radius: var(--radius-xs);
		accent-color: var(--accent-blue);
	}
	.ih-latency-row {
		display: flex;
		align-items: center;
		gap: var(--space-3);
		padding: var(--space-3) var(--space-4);
		background: var(--dbg);
		border: 1px solid var(--dbd);
		border-radius: var(--radius-sm);
		font-size: var(--text-sm);
		color: var(--dt2);
	}
	.ih-latency-row__value {
		font-family: 'SF Mono', 'Fira Code', monospace;
		font-size: var(--text-sm);
		color: var(--accent-green);
	}

	/* Decisions Tab */
	.ih-decision-list {
		display: flex;
		flex-direction: column;
		gap: var(--space-4);
	}
	.ih-decision__top {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
	}
	.ih-decision__header {
		display: flex;
		align-items: center;
		gap: var(--space-2);
	}
	.ih-decision__desc {
		font-size: var(--text-sm);
		color: var(--dt3);
		margin-top: var(--space-1);
		line-height: 1.5;
	}
	.ih-decision__meta {
		font-size: var(--text-xs);
		color: var(--dt4);
		margin-top: var(--space-2);
	}
	.ih-decision__options {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-2);
		margin-top: var(--space-4);
	}

	/* ═══════════════════════════════════════════
	   Detail Modal
	   ═══════════════════════════════════════════ */
	.ih-modal-backdrop {
		position: fixed;
		inset: 0;
		z-index: 50;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: var(--space-4);
		background: rgba(0, 0, 0, 0.5);
	}
	.ih-modal-backdrop__overlay {
		position: fixed;
		inset: 0;
	}
	.ih-modal {
		position: relative;
		z-index: 10;
		background: var(--dbg2);
		border: 1px solid var(--dbd);
		border-radius: var(--radius-lg);
		box-shadow: var(--shadow-xl);
		max-width: 32rem;
		width: 100%;
		max-height: 85vh;
		overflow: hidden;
		display: flex;
		flex-direction: column;
	}
	.ih-modal--sm {
		max-width: 28rem;
	}
	.ih-modal__header {
		padding: var(--space-6);
		border-bottom: 1px solid var(--dbd);
	}
	.ih-modal__header-inner {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
	}
	.ih-modal__provider {
		display: flex;
		align-items: center;
		gap: var(--space-4);
	}
	.ih-modal__provider-icon {
		width: 3rem;
		height: 3rem;
		border-radius: var(--radius-md);
		background: var(--dbg3);
		display: flex;
		align-items: center;
		justify-content: center;
		overflow: hidden;
		flex-shrink: 0;
	}
	.ih-modal__provider-img {
		width: 2rem;
		height: 2rem;
		object-fit: contain;
	}
	.ih-modal__provider-letter {
		font-size: var(--text-xl);
		font-weight: var(--font-bold);
		color: var(--dt3);
	}
	.ih-modal__title {
		font-size: var(--text-xl);
		font-weight: var(--font-bold);
		color: var(--dt);
	}
	.ih-modal__title--sm {
		font-size: var(--text-base);
	}
	.ih-modal__subtitle {
		font-size: var(--text-xs);
		color: var(--dt4);
	}
	.ih-modal__category-badge {
		display: inline-flex;
		align-items: center;
		gap: var(--space-1);
		padding: 2px var(--space-2);
		margin-top: var(--space-1);
		font-size: var(--text-xs);
		font-weight: var(--font-medium);
		border-radius: var(--radius-full);
		background: var(--dbg3);
		color: var(--dt3);
		text-transform: capitalize;
	}
	.ih-modal__body {
		padding: var(--space-6);
		overflow-y: auto;
		flex: 1;
	}
	.ih-modal__desc {
		color: var(--dt3);
		margin-bottom: var(--space-6);
		line-height: 1.5;
		font-size: var(--text-sm);
	}
	.ih-modal__help-text {
		font-size: var(--text-sm);
		color: var(--dt3);
		margin-bottom: var(--space-4);
		line-height: 1.5;
	}
	.ih-modal__section {
		margin-bottom: var(--space-6);
	}
	.ih-modal__section-title {
		font-size: var(--text-sm);
		font-weight: var(--font-semibold);
		color: var(--dt);
		margin-bottom: var(--space-3);
	}
	.ih-modal__footer {
		padding: var(--space-4) var(--space-6);
		border-top: 1px solid var(--dbd);
		background: var(--dbg);
		display: flex;
		align-items: center;
		justify-content: flex-end;
		gap: var(--space-3);
	}
	.ih-modal__full-btn {
		width: 100%;
	}
	.ih-modal__connected-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		width: 100%;
	}
	.ih-modal__connected-status {
		display: flex;
		align-items: center;
		gap: var(--space-2);
		color: var(--accent-green);
	}
	.ih-modal__connected-label {
		font-size: var(--text-sm);
		font-weight: var(--font-medium);
	}
	.ih-modal__connected-account {
		font-size: var(--text-sm);
		color: var(--dt3);
	}
	.ih-modal__connected-actions {
		display: flex;
		align-items: center;
		gap: var(--space-2);
	}

	/* Feature List */
	.ih-feature-list {
		display: flex;
		flex-direction: column;
		gap: var(--space-2);
	}
	.ih-feature-item {
		display: flex;
		align-items: center;
		gap: var(--space-2);
		font-size: var(--text-sm);
		color: var(--dt3);
		line-height: 1.5;
	}
	.ih-feature-icon {
		color: var(--accent-green);
		flex-shrink: 0;
	}

	/* Sync Panel */
	.ih-sync-panel {
		background: var(--dbg);
		border-radius: var(--radius-md);
		padding: var(--space-4);
		margin-bottom: var(--space-6);
	}
	.ih-sync-grid {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: var(--space-4);
	}
	.ih-sync-item {
		display: flex;
		align-items: center;
		gap: var(--space-2);
	}
	.ih-sync-icon {
		width: 2rem;
		height: 2rem;
		border-radius: var(--radius-sm);
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}
	.ih-sync-icon--green {
		background: rgba(16, 185, 129, 0.1);
		color: var(--accent-green);
	}
	.ih-sync-icon--blue {
		background: rgba(74, 144, 226, 0.1);
		color: var(--accent-blue);
	}
	.ih-sync-icon--purple {
		background: rgba(168, 85, 247, 0.1);
		color: var(--accent-purple);
	}
	.ih-sync-icon--amber {
		background: rgba(245, 158, 11, 0.1);
		color: var(--accent-orange);
	}
	.ih-sync-label {
		font-size: var(--text-xs);
		color: var(--dt4);
	}
	.ih-sync-value {
		font-size: var(--text-sm);
		font-weight: var(--font-medium);
		color: var(--dt);
	}

	/* Tags */
	.ih-tag-list {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-2);
	}
	.ih-skill-tag {
		padding: var(--space-1) var(--space-2);
		font-size: var(--text-xs);
		font-family: 'SF Mono', 'Fira Code', monospace;
		background: var(--dbg3);
		color: var(--dt3);
		border-radius: var(--radius-xs);
	}
	.ih-module-tag {
		display: inline-flex;
		align-items: center;
		gap: var(--space-1);
		padding: var(--space-1) var(--space-2);
		font-size: var(--text-xs);
		background: rgba(74, 144, 226, 0.08);
		color: var(--accent-blue);
		border-radius: var(--radius-xs);
		text-transform: capitalize;
	}

	/* Toggle */
	.ih-toggle-label {
		display: flex;
		align-items: center;
		gap: var(--space-2);
		cursor: pointer;
	}
	.ih-toggle-text {
		font-size: var(--text-sm);
		color: var(--dt3);
	}
	.ih-toggle {
		position: relative;
		width: 2.5rem;
		height: 1.25rem;
	}
	.ih-toggle__track {
		width: 100%;
		height: 100%;
		border-radius: var(--radius-full);
		background: var(--dbg3);
		transition: background 200ms ease;
	}
	.ih-toggle :checked ~ .ih-toggle__track {
		background: var(--accent-green);
	}
	.ih-toggle__thumb {
		position: absolute;
		left: 2px;
		top: 2px;
		width: 1rem;
		height: 1rem;
		border-radius: 50%;
		background: white;
		transition: transform 200ms ease;
	}
	.ih-toggle :checked ~ .ih-toggle__thumb {
		transform: translateX(1.25rem);
	}

	/* ═══════════════════════════════════════════
	   File Import
	   ═══════════════════════════════════════════ */
	.ih-import-icon {
		width: 2rem;
		height: 2rem;
		border-radius: var(--radius-sm);
	}
	.ih-dropzone {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		width: 100%;
		height: 8rem;
		border: 2px dashed var(--dbd);
		border-radius: var(--radius-md);
		cursor: pointer;
		transition: border-color 200ms ease, background 200ms ease;
		background: var(--dbg);
	}
	.ih-dropzone:hover {
		border-color: var(--dbd2);
	}
	.ih-dropzone--active {
		border-color: var(--accent-green);
		background: rgba(16, 185, 129, 0.05);
	}
	.ih-dropzone__icon {
		color: var(--dt4);
		margin-bottom: var(--space-2);
	}
	.ih-dropzone__icon--success {
		color: var(--accent-green);
		margin-bottom: var(--space-2);
	}
	.ih-dropzone__filename {
		font-size: var(--text-sm);
		font-weight: var(--font-medium);
		color: var(--accent-green);
	}
	.ih-dropzone__filesize {
		font-size: var(--text-xs);
		color: var(--dt4);
		margin-top: var(--space-1);
	}
	.ih-dropzone__label {
		font-size: var(--text-sm);
		color: var(--dt3);
	}
	.ih-dropzone__formats {
		font-size: var(--text-xs);
		color: var(--dt4);
		margin-top: var(--space-1);
	}

	/* Alerts */
	.ih-alert {
		margin-top: var(--space-3);
		padding: var(--space-3);
		border-radius: var(--radius-sm);
		font-size: var(--text-sm);
	}
	.ih-alert--error {
		background: rgba(220, 38, 38, 0.08);
		border: 1px solid rgba(220, 38, 38, 0.2);
		color: var(--color-error);
	}
	.ih-alert--success {
		background: rgba(16, 185, 129, 0.08);
		border: 1px solid rgba(16, 185, 129, 0.2);
		color: var(--accent-green);
	}
	.ih-import-loading {
		display: flex;
		align-items: center;
		gap: var(--space-2);
	}

	/* Search Input */
	.ih-search-wrap {
		position: relative;
		flex-shrink: 0;
	}
	.ih-search-icon {
		position: absolute;
		left: var(--space-3);
		top: 50%;
		transform: translateY(-50%);
		color: var(--dt4);
		pointer-events: none;
	}
	.ih-search-input {
		padding: var(--space-1) var(--space-3) var(--space-1) 2rem;
		font-size: 0.8125rem;
		border-radius: var(--radius-sm);
		border: 1px solid var(--dbd);
		background: var(--dbg);
		color: var(--dt);
		outline: none;
		width: 14rem;
		transition: border-color 200ms ease;
	}
	.ih-search-input:focus {
		border-color: var(--accent-blue);
	}
	.ih-search-input::placeholder {
		color: var(--dt4);
	}

	/* Card Sync Button */
	.ih-card__sync-btn {
		padding: var(--space-1);
		border-radius: var(--radius-xs);
		color: var(--dt4);
		background: none;
		border: 1px solid var(--dbd);
		cursor: pointer;
		transition: color 200ms ease, background 200ms ease;
		display: inline-flex;
		align-items: center;
	}
	.ih-card__sync-btn:hover:not(:disabled) {
		color: var(--accent-blue);
		background: rgba(74, 144, 226, 0.08);
		border-color: rgba(74, 144, 226, 0.3);
	}
	.ih-card__sync-btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	/* Card sub-meta (last used) */
	.ih-card__sub-meta {
		font-size: 0.6875rem;
		color: var(--dt4);
		margin-top: 2px;
	}

	/* Latency Input */
	.ih-latency-input {
		width: 5rem;
		padding: var(--space-1) var(--space-2);
		font-family: 'SF Mono', 'Fira Code', monospace;
		font-size: var(--text-sm);
		color: var(--accent-green);
		background: var(--dbg2);
		border: 1px solid var(--dbd);
		border-radius: var(--radius-xs);
		outline: none;
		text-align: right;
	}
	.ih-latency-input:focus {
		border-color: var(--accent-blue);
	}
	.ih-latency-unit {
		font-size: var(--text-xs);
		color: var(--dt4);
	}

	/* Alert small variant */
	.ih-alert--sm {
		margin-top: var(--space-2);
		margin-bottom: var(--space-2);
		padding: var(--space-2) var(--space-3);
		font-size: 0.8125rem;
	}

	/* ═══════════════════════════════════════════════════════════
	   MCP SERVERS — mcp- Prefix System
	   ═══════════════════════════════════════════════════════════ */

	.mcp-section {
		display: flex;
		flex-direction: column;
		gap: var(--space-4);
	}
	.mcp-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: var(--space-4);
		margin-bottom: var(--space-2);
	}
	.mcp-header__title {
		font-size: var(--text-lg);
		font-weight: var(--font-semibold);
		color: var(--dt);
		margin: 0;
	}
	.mcp-header__subtitle {
		font-size: 0.8125rem;
		color: var(--dt3);
		margin: 2px 0 0;
	}
	.mcp-grid {
		display: grid;
		grid-template-columns: 1fr;
		gap: var(--space-3);
	}
	@media (min-width: 768px) {
		.mcp-grid { grid-template-columns: repeat(2, 1fr); }
	}
	@media (min-width: 1200px) {
		.mcp-grid { grid-template-columns: repeat(3, 1fr); }
	}
	.mcp-card {
		display: flex;
		flex-direction: column;
		gap: var(--space-3);
	}
	.mcp-card__top {
		display: flex;
		flex-direction: column;
		gap: var(--space-2);
	}
	.mcp-card__info {
		display: flex;
		flex-direction: column;
		gap: var(--space-1);
	}
	.mcp-card__name-row {
		display: flex;
		align-items: center;
		gap: var(--space-2);
	}
	.mcp-card__url {
		font-size: var(--text-xs);
		color: var(--dt4);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		max-width: 100%;
		margin: 0;
	}
	.mcp-card__desc {
		font-size: 0.8125rem;
		color: var(--dt3);
		margin: 0;
		line-height: 1.5;
	}
	.mcp-card__meta {
		display: flex;
		align-items: center;
		gap: var(--space-2);
		flex-wrap: wrap;
	}
	.mcp-card__tools-count {
		display: flex;
		align-items: center;
		gap: var(--space-1);
		font-size: var(--text-xs);
		color: var(--dt3);
	}
	.mcp-card__auth-badge {
		display: inline-flex;
		align-items: center;
		gap: 2px;
		font-size: 0.6875rem;
		color: #facc15;
		background: rgba(250, 204, 21, 0.1);
		padding: 2px var(--space-1);
		border-radius: var(--radius-xs);
	}
	.mcp-card__transport {
		font-size: 0.6875rem;
		color: var(--dt4);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}
	.mcp-card__actions {
		display: flex;
		gap: var(--space-1);
		padding-top: var(--space-2);
		border-top: 1px solid var(--dbd);
	}
	.mcp-status {
		font-size: 0.6875rem;
		font-weight: var(--font-medium);
		padding: 2px var(--space-2);
		border-radius: var(--radius-full);
		text-transform: capitalize;
	}
	.mcp-status--connected {
		color: var(--accent-green);
		background: rgba(16, 185, 129, 0.1);
	}
	.mcp-status--disconnected {
		color: var(--dt4);
		background: rgba(255, 255, 255, 0.05);
	}
	.mcp-status--error {
		color: var(--color-error);
		background: rgba(220, 38, 38, 0.1);
	}
	.mcp-btn--danger {
		color: var(--color-error) !important;
	}
	.mcp-btn--danger:hover {
		background: rgba(220, 38, 38, 0.1) !important;
	}
	.mcp-card__expand-btn {
		display: flex;
		align-items: center;
		gap: var(--space-1);
		font-size: var(--text-xs);
		color: var(--dt3);
		background: none;
		border: none;
		cursor: pointer;
		padding: var(--space-1) 0;
	}
	.mcp-card__expand-btn:hover {
		color: var(--dt);
	}
	.mcp-chevron {
		transition: transform 150ms ease;
	}
	.mcp-chevron--open {
		transform: rotate(180deg);
	}
	.mcp-tools-list {
		display: flex;
		flex-direction: column;
		gap: var(--space-1);
		padding: var(--space-2) 0;
	}
	.mcp-tool-item {
		display: flex;
		flex-direction: column;
		gap: 2px;
		padding: var(--space-2) var(--space-2);
		background: var(--dbg);
		border-radius: var(--radius-xs);
	}
	.mcp-tool-item__name {
		font-size: 0.8125rem;
		font-weight: var(--font-medium);
		color: var(--dt);
		font-family: 'SF Mono', 'Fira Code', monospace;
	}
	.mcp-tool-item__desc {
		font-size: var(--text-xs);
		color: var(--dt4);
		line-height: 1.4;
	}

	/* MCP Form */
	.mcp-form {
		display: flex;
		flex-direction: column;
		gap: var(--space-3);
	}
	.mcp-form__label {
		display: flex;
		flex-direction: column;
		gap: var(--space-1);
		font-size: 0.8125rem;
		color: var(--dt2);
	}
	.mcp-form__required {
		color: var(--color-error);
	}
	.mcp-form__input,
	.mcp-form__select,
	.mcp-form__textarea {
		padding: var(--space-2) var(--space-3);
		font-size: var(--text-sm);
		color: var(--dt);
		background: var(--dbg);
		border: 1px solid var(--dbd);
		border-radius: var(--radius-xs);
		outline: none;
		transition: border-color 200ms ease;
	}
	.mcp-form__input:focus,
	.mcp-form__select:focus,
	.mcp-form__textarea:focus {
		border-color: var(--accent-blue);
	}
	.mcp-form__input:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}
	.mcp-form__textarea {
		resize: vertical;
		min-height: 3rem;
	}
	.mcp-form__hint {
		font-size: 0.6875rem;
		color: var(--dt4);
	}
	.mcp-form__row {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: var(--space-3);
	}
</style>
