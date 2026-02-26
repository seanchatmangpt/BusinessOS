<script lang="ts">
	import { Dialog } from 'bits-ui';
	import { Download, X, CheckCircle, XCircle, Loader2, Sparkles, Zap, Rocket, Building2, FileCode, FolderOpen, ChevronRight, ChevronDown, Copy, Check, Settings2, Code2 } from 'lucide-svelte';
	import { goto } from '$app/navigation';
	import AgentProgressPanel from './AgentProgressPanel.svelte';
	import type { AppGenerationRequest } from '$lib/types/agent';
	import { getCSRFToken } from '$lib/api/base';

	interface Props {
		workspaceId: string;
		open?: boolean;
	}

	let { workspaceId, open = $bindable(false) }: Props = $props();

	type ModalState = 'form' | 'progress' | 'completed' | 'error';
	type CreationMode = 'manual' | 'ai';
	type Complexity = 'simple' | 'standard' | 'complex';

	interface GeneratedFile {
		path: string;
		content: string;
		size: number;
	}

	let modalState = $state<ModalState>('form');
	let creationMode = $state<CreationMode>('ai');
	let complexity = $state<Complexity>('standard');
	let queueItemId = $state('');
	let appName = $state('');
	let description = $state('');
	let errorMessage = $state('');
	let isSubmitting = $state(false);

	// Advanced options state
	let showAdvanced = $state(false);
	type AppType = 'web_app' | 'api' | 'dashboard' | 'landing_page';
	let appType = $state<AppType>('web_app');
	let selectedFeatures = $state<string[]>([]);
	let customPrompt = $state('');

	const APP_TYPES = [
		{ value: 'web_app' as AppType, label: 'Web Application' },
		{ value: 'api' as AppType, label: 'API Service' },
		{ value: 'dashboard' as AppType, label: 'Dashboard' },
		{ value: 'landing_page' as AppType, label: 'Landing Page' }
	] as const;

	const FEATURES = [
		{ id: 'auth', label: 'Authentication' },
		{ id: 'database', label: 'Database' },
		{ id: 'api', label: 'REST API' },
		{ id: 'realtime', label: 'Real-time Updates' },
		{ id: 'file_upload', label: 'File Upload' },
		{ id: 'email', label: 'Email Integration' }
	];

	function toggleFeature(featureId: string) {
		if (selectedFeatures.includes(featureId)) {
			selectedFeatures = selectedFeatures.filter(f => f !== featureId);
		} else {
			selectedFeatures = [...selectedFeatures, featureId];
		}
	}

	// Generated files viewer state
	let generatedFiles = $state<GeneratedFile[]>([]);
	let selectedFile = $state<GeneratedFile | null>(null);
	let isLoadingFiles = $state(false);
	let copiedFile = $state<string | null>(null);
	let expandedFolders = $state<Set<string>>(new Set());
	let showCloseConfirm = $state(false);
	let formValidationErrors = $state<{ appName?: string; description?: string }>({});

	// Validate workspaceId on mount
	const isValidUUID = (str: string) => {
		const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
		return uuidRegex.test(str);
	};

	const hasValidWorkspace = $derived(workspaceId && isValidUUID(workspaceId));

	const COMPLEXITY_OPTIONS = [
		{
			value: 'simple' as Complexity,
			label: 'Simple',
			icon: Zap,
			description: 'Basic CRUD, 2-3 pages',
			agents: 2
		},
		{
			value: 'standard' as Complexity,
			label: 'Standard',
			icon: Rocket,
			description: 'Full features, auth, API',
			agents: 4
		},
		{
			value: 'complex' as Complexity,
			label: 'Complex',
			icon: Building2,
			description: 'Enterprise, multi-module',
			agents: 6
		}
	] as const;

	async function handleSubmit() {
		formValidationErrors = {};
		if (!appName.trim()) {
			formValidationErrors = { ...formValidationErrors, appName: 'App name is required' };
		}
		if (!description.trim()) {
			formValidationErrors = { ...formValidationErrors, description: 'Description is required' };
		}
		if (formValidationErrors.appName || formValidationErrors.description) return;

		// Validate workspace BEFORE making request
		if (!hasValidWorkspace) {
			errorMessage = 'Invalid workspace. Please select a valid workspace before creating an app.';
			modalState = 'error';
			return;
		}

		isSubmitting = true;
		errorMessage = '';

		try {
			// Build generation_context from advanced options if any are set
			const hasAdvancedOptions = appType !== 'web_app' || selectedFeatures.length > 0 || customPrompt.trim();
			const generationContext = hasAdvancedOptions ? {
				app_type: appType,
				...(selectedFeatures.length > 0 && { features: selectedFeatures }),
				...(customPrompt.trim() && { custom_prompt: customPrompt.trim() })
			} : undefined;

			const request: AppGenerationRequest & { complexity?: Complexity; generation_context?: Record<string, unknown> } = {
				app_name: appName,
				description,
				...(creationMode === 'ai' && { complexity }),
				...(generationContext && { generation_context: generationContext })
			};

			const endpoint = creationMode === 'ai'
				? `/api/workspaces/${workspaceId}/apps/generate-osa`
				: `/api/v1/workspaces/${workspaceId}/apps`;

			const csrfToken = getCSRFToken();
			const headers: Record<string, string> = { 'Content-Type': 'application/json' };
			if (csrfToken) {
				headers['X-CSRF-Token'] = csrfToken;
			}

			console.log('[CreateAppModal] Sending request:', {
				endpoint,
				workspaceId,
				workspaceIdValid: hasValidWorkspace,
				request,
				headers: Object.keys(headers)
			});

			const response = await fetch(endpoint, {
				method: 'POST',
				headers,
				credentials: 'include',
				body: JSON.stringify(request)
			});

			console.log('[CreateAppModal] Response status:', response.status);

			if (!response.ok) {
				const errorText = await response.text();
				console.error('[CreateAppModal] Error response body:', errorText);
				let error;
				try {
					error = JSON.parse(errorText);
				} catch {
					error = { message: errorText || 'Unknown error' };
				}
				throw new Error(error.details || error.message || `Server error: ${response.status}`);
			}

			const data = await response.json();

			if (creationMode === 'ai') {
				queueItemId = data.queue_item_id;
				modalState = 'progress';
			} else {
				modalState = 'completed';
			}
		} catch (err) {
			errorMessage = err instanceof Error ? err.message : 'Failed to create app';
			modalState = 'error';
		} finally {
			isSubmitting = false;
		}
	}

	async function handleComplete() {
		modalState = 'completed';
		await fetchGeneratedFiles();
	}

	function handleError(error: string) {
		errorMessage = error;
		modalState = 'error';
	}

	function handleClose() {
		if (modalState === 'progress') {
			showCloseConfirm = true;
			return;
		}
		showCloseConfirm = false;
		open = false;
		setTimeout(resetModal, 300);
	}

	function handleForceClose() {
		showCloseConfirm = false;
		open = false;
		setTimeout(resetModal, 300);
	}

	function resetModal() {
		modalState = 'form';
		creationMode = 'ai';
		complexity = 'standard';
		appName = '';
		description = '';
		errorMessage = '';
		queueItemId = '';
		generatedFiles = [];
		selectedFile = null;
		expandedFolders = new Set();
		showAdvanced = false;
		appType = 'web_app';
		selectedFeatures = [];
		customPrompt = '';
		showCloseConfirm = false;
		formValidationErrors = {};
	}

	function handleRetry() {
		modalState = 'form';
		errorMessage = '';
	}

	async function fetchGeneratedFiles() {
		isLoadingFiles = true;
		try {
			const response = await fetch(`/api/osa/apps/${queueItemId}/generated-files`, {
				credentials: 'include'
			});

			if (!response.ok) {
				console.warn('[CreateAppModal] Failed to fetch generated files:', response.status);
				return;
			}

			const data = await response.json();
			generatedFiles = data.files || [];

			// Auto-expand all top-level folders and select first file
			const folders = new Set<string>();
			for (const file of generatedFiles) {
				const topFolder = file.path.split('/')[0];
				if (topFolder) folders.add(topFolder);
			}
			expandedFolders = folders;
			if (generatedFiles.length > 0) {
				selectedFile = generatedFiles[0];
			}

			console.log(`[CreateAppModal] Loaded ${generatedFiles.length} generated files`);
		} catch (err) {
			console.warn('[CreateAppModal] Error fetching files:', err);
		} finally {
			isLoadingFiles = false;
		}
	}

	async function handleDownload() {
		try {
			const response = await fetch(`/api/osa/apps/${queueItemId}/download`, {
				credentials: 'include'
			});

			if (!response.ok) {
				throw new Error('Download failed');
			}

			const blob = await response.blob();
			const url = window.URL.createObjectURL(blob);
			const a = document.createElement('a');
			a.href = url;
			a.download = `${appName.replace(/\s+/g, '-').toLowerCase()}.zip`;
			document.body.appendChild(a);
			a.click();
			window.URL.revokeObjectURL(url);
			document.body.removeChild(a);
		} catch (err) {
			errorMessage = 'Failed to download app. The workspace files are available on the server.';
			modalState = 'error';
		}
	}

	function copyFileContent(content: string, path: string) {
		navigator.clipboard.writeText(content);
		copiedFile = path;
		setTimeout(() => { copiedFile = null; }, 2000);
	}

	function toggleFolder(folder: string) {
		const next = new Set(expandedFolders);
		if (next.has(folder)) {
			next.delete(folder);
		} else {
			next.add(folder);
		}
		expandedFolders = next;
	}

	function getLanguageFromPath(path: string): string {
		const ext = path.split('.').pop()?.toLowerCase() || '';
		const langMap: Record<string, string> = {
			go: 'go', ts: 'typescript', tsx: 'tsx', js: 'javascript', jsx: 'jsx',
			svelte: 'svelte', sql: 'sql', json: 'json', yaml: 'yaml', yml: 'yaml',
			md: 'markdown', html: 'html', css: 'css', sh: 'bash', mod: 'go', sum: 'text'
		};
		return langMap[ext] || 'text';
	}

	// Group files by top-level directory
	let fileTree = $derived.by(() => {
		const tree = new Map<string, GeneratedFile[]>();
		for (const file of generatedFiles) {
			const parts = file.path.split('/');
			const folder = parts.length > 1 ? parts[0] : '(root)';
			if (!tree.has(folder)) tree.set(folder, []);
			tree.get(folder)!.push(file);
		}
		return tree;
	});
</script>

<Dialog.Root bind:open>
	<Dialog.Portal>
		<Dialog.Overlay class="fixed inset-0 z-50 bg-black/50 backdrop-blur-sm" />
		<Dialog.Content
			class="fixed left-[50%] top-[50%] z-50 max-h-[90vh] w-[90vw] max-w-4xl translate-x-[-50%] translate-y-[-50%] rounded-lg border border-gray-200 bg-white p-6 shadow-lg dark:border-gray-700 dark:bg-gray-800 overflow-y-auto"
			onInteractOutside={(e) => {
				if (modalState === 'progress') {
					e.preventDefault();
					showCloseConfirm = true;
				}
			}}
		>
			<div class="flex items-center justify-between mb-6">
				<div class="flex items-center gap-3">
					<div class="w-10 h-10 rounded-full bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center">
						<Sparkles class="w-5 h-5 text-white" />
					</div>
					<div>
						<Dialog.Title class="text-2xl font-bold text-gray-900 dark:text-white">
							{#if modalState === 'form'}
								Create New App
							{:else if modalState === 'progress'}
								Generating App...
							{:else if modalState === 'completed'}
								Generation Complete!
							{:else}
								Generation Failed
							{/if}
						</Dialog.Title>
						<Dialog.Description class="text-sm text-gray-600 dark:text-gray-400">
							{#if modalState === 'form'}
								Describe your app and let AI agents build it for you
							{:else if modalState === 'progress'}
								4 AI agents working in parallel
							{:else if modalState === 'completed'}
								Your app is ready to download
							{:else}
								Something went wrong during generation
							{/if}
						</Dialog.Description>
					</div>
				</div>

				<Dialog.Close
					class="rounded-lg p-2 text-gray-500 hover:bg-gray-100 hover:text-gray-900 dark:hover:bg-gray-700 dark:hover:text-white transition-colors"
					onclick={handleClose}
				>
					<X class="w-5 h-5" />
				</Dialog.Close>
			</div>

			{#if modalState === 'form'}
				{#if !hasValidWorkspace}
					<div class="mb-6 p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg">
						<div class="flex items-start gap-3">
							<div class="w-5 h-5 text-red-500 flex-shrink-0 mt-0.5">
								<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
								</svg>
							</div>
							<div>
								<p class="text-sm font-medium text-red-800 dark:text-red-200">No Workspace Selected</p>
								<p class="text-sm text-red-600 dark:text-red-300 mt-1">
									Please select a valid workspace before creating an app.
									{#if workspaceId}
										<span class="block mt-1 text-xs font-mono">Current: {workspaceId}</span>
									{/if}
								</p>
							</div>
						</div>
					</div>
				{/if}

				<div class="flex gap-1 p-1 bg-gray-100 dark:bg-gray-700 rounded-lg mb-6">
					<button
						type="button"
						onclick={() => creationMode = 'ai'}
						class="flex-1 px-4 py-2 text-sm font-medium rounded-md transition-colors flex items-center justify-center gap-2"
						class:bg-white={creationMode === 'ai'}
						class:dark:bg-gray-800={creationMode === 'ai'}
						class:text-gray-900={creationMode === 'ai'}
						class:dark:text-white={creationMode === 'ai'}
						class:shadow-sm={creationMode === 'ai'}
						class:text-gray-600={creationMode !== 'ai'}
						class:dark:text-gray-400={creationMode !== 'ai'}
					>
						<Sparkles class="w-4 h-4" />
						Generate with AI
					</button>
					<button
						type="button"
						onclick={() => creationMode = 'manual'}
						class="flex-1 px-4 py-2 text-sm font-medium rounded-md transition-colors"
						class:bg-white={creationMode === 'manual'}
						class:dark:bg-gray-800={creationMode === 'manual'}
						class:text-gray-900={creationMode === 'manual'}
						class:dark:text-white={creationMode === 'manual'}
						class:shadow-sm={creationMode === 'manual'}
						class:text-gray-600={creationMode !== 'manual'}
						class:dark:text-gray-400={creationMode !== 'manual'}
					>
						Create Manually
					</button>
				</div>

				<form onsubmit={(e) => { e.preventDefault(); handleSubmit(); }} class="space-y-4">
					<div>
						<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
							App Name <span class="text-red-500">*</span>
						</label>
						<input
							type="text"
							bind:value={appName}
							placeholder="My Awesome App"
							class="w-full px-4 py-2.5 border rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white placeholder-gray-400 focus:outline-none focus:ring-2 transition-shadow {formValidationErrors.appName ? 'border-red-500 focus:ring-red-500' : 'border-gray-300 dark:border-gray-600 focus:ring-blue-500 dark:focus:ring-blue-400'}"
							required
							disabled={isSubmitting}
						/>
						{#if formValidationErrors.appName}
							<p class="text-xs text-red-500 mt-1">{formValidationErrors.appName}</p>
						{/if}
					</div>

					<div>
						<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
							Description <span class="text-red-500">*</span>
						</label>
						<textarea
							bind:value={description}
							placeholder="A task management app with kanban boards, due dates, and team collaboration features..."
							rows="4"
							class="w-full px-4 py-2.5 border rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white placeholder-gray-400 focus:outline-none focus:ring-2 resize-none transition-shadow {formValidationErrors.description ? 'border-red-500 focus:ring-red-500' : 'border-gray-300 dark:border-gray-600 focus:ring-blue-500 dark:focus:ring-blue-400'}"
							required
							disabled={isSubmitting}
						/>
						{#if formValidationErrors.description}
							<p class="text-xs text-red-500 mt-1">{formValidationErrors.description}</p>
						{:else}
							<p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
								Be specific about features, functionality, and user flows for best results
							</p>
						{/if}
					</div>

					{#if creationMode === 'ai'}
						<div>
							<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-3">
								App Complexity
							</label>
							<div class="grid grid-cols-3 gap-3">
								{#each COMPLEXITY_OPTIONS as option}
									{@const Icon = option.icon}
									{@const isSelected = complexity === option.value}
									<button
										type="button"
										onclick={() => complexity = option.value}
										class="p-3 border-2 rounded-lg text-left transition-all {isSelected ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20' : 'border-gray-200 dark:border-gray-600 hover:border-gray-300'}"
									>
										<Icon class="w-5 h-5 mb-2 {isSelected ? 'text-blue-600' : 'text-gray-400'}" />
										<div class="font-medium text-sm text-gray-900 dark:text-white">
											{option.label}
										</div>
										<div class="text-xs text-gray-500 dark:text-gray-400 mt-1">
											{option.description}
										</div>
										<div class="text-xs text-blue-600 dark:text-blue-400 mt-2">
											{option.agents} agents
										</div>
									</button>
								{/each}
							</div>
						</div>
					{/if}

					{#if creationMode === 'ai'}
						<div>
							<button
								type="button"
								onclick={() => showAdvanced = !showAdvanced}
								disabled={isSubmitting}
								class="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200 transition-colors"
							>
								<Settings2 class="w-4 h-4" />
								{showAdvanced ? 'Hide' : 'Show'} Advanced Options
								<ChevronDown class="w-3.5 h-3.5 transition-transform {showAdvanced ? 'rotate-180' : ''}" />
							</button>

							{#if showAdvanced}
								<div class="mt-3 space-y-4 p-4 border border-gray-200 dark:border-gray-600 rounded-lg bg-gray-50 dark:bg-gray-700/50">
									<!-- App Type -->
									<div>
										<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
											App Type
										</label>
										<select
											bind:value={appType}
											disabled={isSubmitting}
											class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500 transition-shadow"
										>
											{#each APP_TYPES as type}
												<option value={type.value}>{type.label}</option>
											{/each}
										</select>
									</div>

									<!-- Features -->
									<div>
										<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
											Features
										</label>
										<div class="flex flex-wrap gap-2">
											{#each FEATURES as feature}
												<button
													type="button"
													onclick={() => toggleFeature(feature.id)}
													disabled={isSubmitting}
													class="px-3 py-1.5 text-sm rounded-full border transition-colors {selectedFeatures.includes(feature.id)
														? 'bg-blue-100 border-blue-500 text-blue-700 dark:bg-blue-900/40 dark:border-blue-400 dark:text-blue-300'
														: 'bg-white border-gray-300 text-gray-600 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-600'}"
												>
													{feature.label}
												</button>
											{/each}
										</div>
									</div>

									<!-- Custom Prompt -->
									<div>
										<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
											Custom AI Instructions
										</label>
										<textarea
											bind:value={customPrompt}
											placeholder="Add specific instructions for the AI agents..."
											rows="2"
											disabled={isSubmitting}
											class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none transition-shadow text-sm"
										/>
									</div>
								</div>
							{/if}
						</div>
					{/if}

					<div class="flex justify-end gap-3 pt-4">
						<button
							type="button"
							onclick={handleClose}
							class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-600 transition-colors"
							disabled={isSubmitting}
						>
							Cancel
						</button>
						<button
							type="submit"
							class="px-4 py-2 text-sm font-medium text-white bg-gradient-to-r from-blue-600 to-purple-600 rounded-lg hover:from-blue-700 hover:to-purple-700 disabled:opacity-50 disabled:cursor-not-allowed transition-all flex items-center gap-2"
							disabled={isSubmitting || !appName.trim() || !description.trim() || !hasValidWorkspace}
						>
							{#if isSubmitting}
								<Loader2 class="w-4 h-4 animate-spin" />
								{creationMode === 'ai' ? 'Starting...' : 'Creating...'}
							{:else if creationMode === 'ai'}
								<Sparkles class="w-4 h-4" />
								Generate App
							{:else}
								Create App
							{/if}
						</button>
					</div>
				</form>

			{:else if modalState === 'progress'}
				<AgentProgressPanel
					{queueItemId}
					onComplete={handleComplete}
					onError={handleError}
				/>

			{:else if modalState === 'completed'}
				<div class="space-y-4">
					<!-- Success header -->
					<div class="flex items-center justify-between">
						<div class="flex items-center gap-3">
							<div class="w-10 h-10 bg-green-100 dark:bg-green-900/30 rounded-full flex items-center justify-center">
								<CheckCircle class="w-5 h-5 text-green-600 dark:text-green-400" />
							</div>
							<div>
								<h3 class="text-lg font-semibold text-gray-900 dark:text-white">
									{appName} Generated!
								</h3>
								<p class="text-sm text-gray-500 dark:text-gray-400">
									{generatedFiles.length} files created
								</p>
							</div>
						</div>
						<div class="flex gap-2">
							<button
								onclick={() => { open = false; goto(`/generated-apps/${queueItemId}`); }}
								class="px-3 py-1.5 text-sm font-medium text-white bg-gradient-to-r from-blue-600 to-blue-700 rounded-lg hover:from-blue-700 hover:to-blue-800 transition-all flex items-center gap-1.5"
							>
								<Code2 class="w-3.5 h-3.5" />
								Open in Editor
							</button>
							<button
								onclick={handleDownload}
								class="px-3 py-1.5 text-sm font-medium text-white bg-gradient-to-r from-green-600 to-green-700 rounded-lg hover:from-green-700 hover:to-green-800 transition-all flex items-center gap-1.5"
							>
								<Download class="w-3.5 h-3.5" />
								Download ZIP
							</button>
							<button
								onclick={handleClose}
								class="px-3 py-1.5 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-600 transition-colors"
							>
								Close
							</button>
						</div>
					</div>

					<!-- Code viewer -->
					{#if isLoadingFiles}
						<div class="flex items-center justify-center py-12">
							<Loader2 class="w-6 h-6 animate-spin text-gray-400 mr-2" />
							<span class="text-gray-500">Loading generated files...</span>
						</div>
					{:else if generatedFiles.length > 0}
						<div class="flex border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden" style="height: 420px;">
							<!-- File tree sidebar -->
							<div class="w-60 border-r border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-900 overflow-y-auto flex-shrink-0">
								<div class="p-2 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider border-b border-gray-200 dark:border-gray-700">
									Files
								</div>
								{#each [...fileTree.entries()] as [folder, files]}
									<div>
										<button
											onclick={() => toggleFolder(folder)}
											class="w-full flex items-center gap-1.5 px-2 py-1.5 text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
										>
											{#if expandedFolders.has(folder)}
												<ChevronDown class="w-3.5 h-3.5 text-gray-400 flex-shrink-0" />
											{:else}
												<ChevronRight class="w-3.5 h-3.5 text-gray-400 flex-shrink-0" />
											{/if}
											<FolderOpen class="w-3.5 h-3.5 text-yellow-500 flex-shrink-0" />
											<span class="truncate">{folder}</span>
											<span class="ml-auto text-xs text-gray-400">{files.length}</span>
										</button>
										{#if expandedFolders.has(folder)}
											{#each files as file}
												{@const fileName = file.path.split('/').slice(1).join('/') || file.path}
												<button
													onclick={() => selectedFile = file}
													class="w-full flex items-center gap-1.5 pl-7 pr-2 py-1 text-xs transition-colors {selectedFile?.path === file.path ? 'bg-blue-50 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300' : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800'}"
													title={file.path}
												>
													<FileCode class="w-3 h-3 flex-shrink-0 {selectedFile?.path === file.path ? 'text-blue-500' : 'text-gray-400'}" />
													<span class="truncate">{fileName}</span>
												</button>
											{/each}
										{/if}
									</div>
								{/each}
							</div>

							<!-- Code content -->
							<div class="flex-1 flex flex-col min-w-0 bg-gray-900">
								{#if selectedFile}
									<!-- File header -->
									<div class="flex items-center justify-between px-3 py-2 bg-gray-800 border-b border-gray-700 flex-shrink-0">
										<span class="text-xs text-gray-300 font-mono truncate">{selectedFile.path}</span>
										<div class="flex items-center gap-2 flex-shrink-0">
											<span class="text-xs text-gray-500">{getLanguageFromPath(selectedFile.path)}</span>
											<button
												onclick={() => selectedFile && copyFileContent(selectedFile.content, selectedFile.path)}
												class="p-1 text-gray-400 hover:text-white transition-colors rounded"
												title="Copy to clipboard"
											>
												{#if copiedFile === selectedFile.path}
													<Check class="w-3.5 h-3.5 text-green-400" />
												{:else}
													<Copy class="w-3.5 h-3.5" />
												{/if}
											</button>
										</div>
									</div>
									<!-- Code display -->
									<div class="flex-1 overflow-auto">
										<pre class="p-3 text-xs leading-relaxed"><code class="text-gray-100 font-mono whitespace-pre">{selectedFile.content}</code></pre>
									</div>
								{:else}
									<div class="flex-1 flex items-center justify-center text-gray-500 text-sm">
										Select a file to view its contents
									</div>
								{/if}
							</div>
						</div>
					{:else}
						<div class="text-center py-8 text-gray-500 dark:text-gray-400">
							<p>No generated files available for preview.</p>
							<p class="text-sm mt-1">You can still download the app using the button above.</p>
						</div>
					{/if}
				</div>

			{:else if modalState === 'error'}
				<div class="text-center py-8">
					<div class="w-16 h-16 bg-red-100 dark:bg-red-900/30 rounded-full flex items-center justify-center mx-auto mb-4">
						<XCircle class="w-8 h-8 text-red-600 dark:text-red-400" />
					</div>
					<h3 class="text-xl font-semibold text-gray-900 dark:text-white mb-2">
						Generation Failed
					</h3>
					<p class="text-gray-600 dark:text-gray-400 mb-6">
						{errorMessage}
					</p>

					<div class="flex justify-center gap-3">
						<button
							onclick={handleClose}
							class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-600 transition-colors"
						>
							Close
						</button>
						<button
							onclick={handleRetry}
							class="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 transition-colors"
						>
							Try Again
						</button>
					</div>
				</div>
			{/if}
			<!-- Close confirmation dialog during generation -->
			{#if showCloseConfirm}
				<div class="fixed inset-0 z-[100] flex items-center justify-center bg-black/40">
					<div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-6 max-w-sm mx-4 shadow-xl">
						<h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-2">Close during generation?</h3>
						<p class="text-sm text-gray-600 dark:text-gray-400 mb-4">
							App generation is in progress. You can safely close this dialog — generation will continue in the background.
						</p>
						<div class="flex justify-end gap-3">
							<button
								onclick={() => showCloseConfirm = false}
								class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 rounded-lg transition-colors"
							>
								Keep Open
							</button>
							<button
								onclick={handleForceClose}
								class="px-4 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded-lg transition-colors"
							>
								Close Anyway
							</button>
						</div>
					</div>
				</div>
			{/if}
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>
