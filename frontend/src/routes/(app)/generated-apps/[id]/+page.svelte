<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { Monitor, Maximize2, Minimize2, Code2, Eye, Search, ExternalLink, RefreshCw, AlertTriangle, History, Terminal } from 'lucide-svelte';
	import { generatedAppsStore, type GeneratedApp } from '$lib/stores/generatedAppsStore';
	import BuildProgress from '$lib/components/osa/BuildProgress.svelte';
	import AppActions from '$lib/components/osa/AppActions.svelte';
	import SandboxStatusBadge from '$lib/components/osa/SandboxStatusBadge.svelte';
	import OpenSandboxButton from '$lib/components/osa/OpenSandboxButton.svelte';
	import FileTree from '$lib/components/osa/FileTree.svelte';
	import type { OSAFile, FileTreeNode } from '$lib/components/osa/types';
	import { getWorkflowFiles, getFileContent } from '$lib/api/osa/files';
	import MonacoEditor from '$lib/editor/MonacoEditor.svelte';
	import EditorToolbar from '$lib/editor/EditorToolbar.svelte';
	import EditorStatusBar from '$lib/editor/EditorStatusBar.svelte';
	import { detectLanguage } from '$lib/editor/utils/language-detection';
	import { deploySandbox, stopSandbox as stopSandboxAPI, getSandboxInfo } from '$lib/api/sandbox';
	import {
		VersionBadge,
		VersionDropdown,
		VersionTimelinePanel,
		SaveVersionModal,
		VersionPreviewModal,
		RestoreConfirmDialog,
		VersionDiffModal,
		toVersionSummary
	} from '$lib/components/versioning';
	import type { Version, VersionSummary } from '$lib/types/versions';
	import { listAppVersions, createAppSnapshot, restoreAppVersion } from '$lib/api/versions';
	import { mapBackendVersionsList, extractDisplayNumber } from '$lib/api/versions/mappers';
	import type { BackendVersionInfo } from '$lib/api/versions/types';

	let appId = $derived($page.params.id as string);
	let app = $state<GeneratedApp | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let previewExpanded = $state(false);

	// ── Code workspace state ──
	let activeTab = $state<'preview' | 'code' | 'terminal'>('preview');
	let isSaving = $state(false);
	let saveError = $state<string | null>(null);
	let files = $state<OSAFile[]>([]);
	let fileTreeNodes = $state<FileTreeNode[]>([]);
	let selectedFile = $state<OSAFile | null>(null);
	let editorValue = $state('');
	let originalValue = $state('');
	let isEditing = $state(false);
	let isDirty = $derived(editorValue !== originalValue);
	let fileLoading = $state(false);
	let cursorLine = $state(1);
	let cursorColumn = $state(1);
	let editorRef = $state<MonacoEditor>();
	let isTransitioning = $state(false);
	let fileSearchQuery = $state('');
	let languageId = $derived(selectedFile ? detectLanguage(selectedFile.path || selectedFile.name) : 'plaintext');

	// ── Sandbox execution state ──
	let sandboxError = $state<string | null>(null);
	let sandboxPolling = $state<ReturnType<typeof setInterval> | null>(null);
	let showStopConfirm = $state(false);
	let sandboxDeployProgress = $state(0);

	// ── Version history state ──
	let versions = $state<Version[]>([]);
	let versionsLoading = $state(false);
	let currentVersion = $derived(versions.find((v) => v.isCurrent)?.versionNumber ?? 1);
	let versionSummaries = $derived(versions.map(toVersionSummary));
	let timelinePanelOpen = $state(false);
	let saveModalOpen = $state(false);
	let isSavingVersion = $state(false);
	let previewModalOpen = $state(false);
	let previewVersion = $state<Version | null>(null);
	let restoreDialogOpen = $state(false);
	let restoreVersion = $state<Version | null>(null);
	let isRestoring = $state(false);
	let rawBackendVersions = $state<BackendVersionInfo[]>([]);
	let diffModalOpen = $state(false);
	let diffFromVersion = $state<BackendVersionInfo | null>(null);
	let diffToVersion = $state<BackendVersionInfo | null>(null);

	// Build file tree from flat file list
	function buildFileTree(files: OSAFile[]): FileTreeNode[] {
		const root: FileTreeNode[] = [];
		const folderMap = new Map<string, FileTreeNode>();

		for (const file of files) {
			const parts = (file.path || file.name).split('/');
			let currentPath = '';

			for (let i = 0; i < parts.length; i++) {
				const part = parts[i];
				const parentPath = currentPath;
				currentPath = currentPath ? `${currentPath}/${part}` : part;

				if (i === parts.length - 1) {
					// File node
					const node: FileTreeNode = {
						id: file.id,
						name: part,
						path: currentPath,
						type: 'file',
						file,
					};
					const parent = folderMap.get(parentPath);
					if (parent) {
						parent.children = parent.children || [];
						parent.children.push(node);
					} else {
						root.push(node);
					}
				} else {
					// Folder node
					if (!folderMap.has(currentPath)) {
						const folder: FileTreeNode = {
							id: `folder-${currentPath}`,
							name: part,
							path: currentPath,
							type: 'folder',
							children: [],
							expanded: true,
						};
						folderMap.set(currentPath, folder);
						const parent = folderMap.get(parentPath);
						if (parent) {
							parent.children = parent.children || [];
							parent.children.push(folder);
						} else {
							root.push(folder);
						}
					}
				}
			}
		}
		return root;
	}

	// Filter tree nodes by search query
	function filterTree(nodes: FileTreeNode[], query: string): FileTreeNode[] {
		if (!query.trim()) return nodes;
		const q = query.toLowerCase();
		return nodes
			.map((node) => {
				if (node.type === 'folder') {
					const children = filterTree(node.children || [], query);
					if (children.length > 0) return { ...node, children, expanded: true };
					return null;
				}
				return node.name.toLowerCase().includes(q) || node.path.toLowerCase().includes(q)
					? node
					: null;
			})
			.filter(Boolean) as FileTreeNode[];
	}

	let filteredTreeNodes = $derived(filterTree(fileTreeNodes, fileSearchQuery));

	async function loadFiles() {
		if (!app) return;
		try {
			// Try fetching files for the app's workflow
			const appFiles = await getWorkflowFiles(app.id);
			files = appFiles;
			fileTreeNodes = buildFileTree(appFiles);
		} catch {
			// Files may not be available yet
			files = [];
			fileTreeNodes = [];
		}
	}

	async function handleFileSelect(file: OSAFile) {
		if (selectedFile?.id === file.id) return;
		isTransitioning = true;
		fileLoading = true;

		setTimeout(async () => {
			selectedFile = file;
			isEditing = false;

			try {
				if (file.content) {
					editorValue = file.content;
					originalValue = file.content;
				} else {
					const result = await getFileContent(file.id);
					editorValue = result.content;
					originalValue = result.content;
				}
			} catch {
				editorValue = '// Failed to load file content';
				originalValue = editorValue;
			} finally {
				fileLoading = false;
				requestAnimationFrame(() => {
					isTransitioning = false;
				});
			}
		}, 150);
	}

	function handleToggleEdit() {
		isEditing = !isEditing;
		if (isEditing) {
			setTimeout(() => editorRef?.focus(), 50);
		}
	}

	async function handleSave(value: string) {
		if (!selectedFile || isSaving) return;
		isSaving = true;
		saveError = null;

		try {
			const response = await fetch(`/api/osa/apps/${appId}/files`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				credentials: 'include',
				body: JSON.stringify({
					file_id: selectedFile.id,
					content: value
				})
			});

			if (!response.ok) {
				if (response.status === 404) {
					console.warn('[Editor] Save endpoint not available yet (404)');
				}
				const errorData = await response.json().catch(() => ({}));
				throw new Error(errorData.message || `Save failed (${response.status})`);
			}

			originalValue = value;
			isEditing = false;
		} catch (err) {
			saveError = err instanceof Error ? err.message : 'Failed to save file';
			console.error('[Editor] Save failed:', err);
		} finally {
			isSaving = false;
		}
	}

	function handleCopy() {
		navigator.clipboard.writeText(editorValue);
	}

	function handleEditorChange(_value: string) {
		const editor = editorRef?.getEditor?.();
		if (editor) {
			const pos = editor.getPosition();
			if (pos) {
				cursorLine = pos.lineNumber;
				cursorColumn = pos.column;
			}
		}
	}

	onMount(async () => {
		loading = true;
		try {
			app = await generatedAppsStore.getAppById(appId);
			if (!app) {
				error = 'App not found';
				return;
			}

			// Subscribe to SSE if app is generating
			if (app.status === 'generating') {
				generatedAppsStore.subscribeToAppProgress(appId);
			}

			// Load files and versions if app is generated or deployed
			if (app.status === 'generated' || app.status === 'deployed') {
				await Promise.all([loadFiles(), loadVersions()]);
			}

			// Start polling if sandbox is deploying
			if (app.sandbox?.status === 'deploying' || app.sandbox?.status === 'pending') {
				startSandboxPolling();
			}
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load app';
		} finally {
			loading = false;
		}
	});

	onDestroy(() => {
		if (app?.status === 'generating') {
			generatedAppsStore.unsubscribeFromAppProgress(appId);
		}
		stopSandboxPolling();
	});

	async function handleDeploy() {
		if (!app) return;
		try {
			await generatedAppsStore.deployApp(app.id);
			// Reload app data
			app = await generatedAppsStore.getAppById(appId);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to deploy app';
		}
	}

	async function handleDelete() {
		if (!app) return;
		try {
			await generatedAppsStore.deleteApp(app.id);
			goto('/generated-apps');
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to delete app';
		}
	}

	function handleBack() {
		goto('/generated-apps');
	}

	async function handleStartSandbox() {
		if (!app) return;
		try {
			sandboxError = null;
			sandboxDeployProgress = 0;
			await deploySandbox(app.id, app.app_name);
			startSandboxPolling();
			app = await generatedAppsStore.getAppById(appId);
		} catch (err) {
			sandboxError = err instanceof Error ? err.message : 'Failed to start sandbox';
		}
	}

	async function handleStopSandbox() {
		if (!app) return;
		showStopConfirm = false;
		try {
			sandboxError = null;
			await stopSandboxAPI(app.id);
			stopSandboxPolling();
			sandboxDeployProgress = 0;
			app = await generatedAppsStore.getAppById(appId);
		} catch (err) {
			sandboxError = err instanceof Error ? err.message : 'Failed to stop sandbox';
		}
	}

	function startSandboxPolling() {
		stopSandboxPolling();
		sandboxPolling = setInterval(async () => {
			if (!app) return;
			try {
				const info = await getSandboxInfo(app.id);
				if (info.status === 'deploying' || info.status === 'pending') {
					sandboxDeployProgress = Math.min(sandboxDeployProgress + 12, 90);
				}
				if (info.status === 'running' || info.status === 'stopped' || info.status === 'failed') {
					sandboxDeployProgress = info.status === 'running' ? 100 : 0;
					stopSandboxPolling();
					if (info.status === 'failed') {
						sandboxError = 'Sandbox deployment failed. Try again.';
					}
				}
				app = await generatedAppsStore.getAppById(appId);
			} catch {
				// Silently retry on next interval
			}
		}, 3000);
	}

	function stopSandboxPolling() {
		if (sandboxPolling) {
			clearInterval(sandboxPolling);
			sandboxPolling = null;
		}
	}

	function handleRetrySandbox() {
		sandboxError = null;
		handleStartSandbox();
	}

	// ── Version handlers ──
	async function loadVersions() {
		if (!app) return;
		versionsLoading = true;
		try {
			rawBackendVersions = await listAppVersions(app.workspace_id, app.id);
			versions = mapBackendVersionsList(rawBackendVersions);
		} catch (err) {
			console.error('Failed to load versions:', err);
			versions = [];
			rawBackendVersions = [];
		} finally {
			versionsLoading = false;
		}
	}

	function handleVersionSelect(summary: VersionSummary) {
		const version = versions.find((v) => v.id === summary.id);
		if (version) {
			previewVersion = version;
			previewModalOpen = true;
		}
	}

	function handleVersionPreview(version: Version) {
		previewVersion = version;
		previewModalOpen = true;
	}

	function handleVersionRestore(version: Version) {
		restoreVersion = version;
		restoreDialogOpen = true;
	}

	async function handleConfirmRestore() {
		if (!restoreVersion || !app) return;
		isRestoring = true;
		try {
			const backendVer = restoreVersion.backendVersion
				?? rawBackendVersions.find((v) => v.id === restoreVersion!.id)?.version_number;
			if (!backendVer) throw new Error('Version not found');

			await restoreAppVersion(app.workspace_id, app.id, backendVer);
			restoreDialogOpen = false;
			previewModalOpen = false;
			await Promise.all([loadVersions(), loadFiles()]);
		} catch (err) {
			console.error('Failed to restore version:', err);
		} finally {
			isRestoring = false;
		}
	}

	async function handleSaveVersion(label?: string) {
		if (!app) return;
		isSavingVersion = true;
		try {
			await createAppSnapshot(app.workspace_id, app.id, label);
			saveModalOpen = false;
			await loadVersions();
		} catch (err) {
			console.error('Failed to save version:', err);
		} finally {
			isSavingVersion = false;
		}
	}

	function handleCompareVersions() {
		if (rawBackendVersions.length < 2) return;
		diffFromVersion = rawBackendVersions[rawBackendVersions.length - 1]; // oldest
		diffToVersion = rawBackendVersions[0]; // newest
		diffModalOpen = true;
	}
</script>

<svelte:head>
	<link rel="preconnect" href="https://fonts.googleapis.com" />
	<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin="anonymous" />
	<link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500;600&display=swap" rel="stylesheet" />
</svelte:head>

<div class="h-full flex flex-col bg-gray-50 dark:bg-gray-900">
	<!-- Header -->
	<div class="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">
		<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
			<div class="flex items-center gap-4 mb-4">
				<button
					onclick={handleBack}
					class="p-2 text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors"
					aria-label="Back to apps"
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M15 19l-7-7 7-7"
						/>
					</svg>
				</button>
				<div class="flex-1">
					<div class="flex items-center gap-3">
						<h1 class="text-2xl font-bold text-gray-900 dark:text-white">
							{app?.app_name || 'Loading...'}
						</h1>
						{#if versions.length > 0}
							<VersionBadge
								{currentVersion}
								onclick={() => timelinePanelOpen = true}
							/>
						{/if}
					</div>
					{#if app}
						<p class="text-sm text-gray-600 dark:text-gray-400 mt-1">{app.description}</p>
					{/if}
				</div>
				{#if versions.length > 0}
					<VersionDropdown
						{currentVersion}
						versions={versionSummaries}
						onVersionSelect={handleVersionSelect}
						onViewAll={() => timelinePanelOpen = true}
						onSaveVersion={() => saveModalOpen = true}
					/>
				{/if}
			</div>
		</div>
	</div>

	<!-- Content -->
	<div class="flex-1 overflow-auto">
		<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
			{#if error}
				<!-- Error State -->
				<div class="bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-800 rounded-lg p-6">
					<div class="flex items-start gap-3">
						<svg
							class="w-6 h-6 text-red-600 dark:text-red-400 flex-shrink-0"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
							/>
						</svg>
						<div>
							<h3 class="text-lg font-semibold text-red-900 dark:text-red-200">Error Loading App</h3>
							<p class="text-sm text-red-700 dark:text-red-300 mt-1">{error}</p>
							<button
								onclick={handleBack}
								class="mt-3 px-4 py-2 text-sm font-medium text-white bg-red-600 hover:bg-red-700 rounded-lg transition-colors"
							>
								Back to Apps
							</button>
						</div>
					</div>
				</div>
			{:else if loading}
				<!-- Loading State -->
				<div class="flex items-center justify-center py-12">
					<div class="flex flex-col items-center gap-4">
						<svg
							class="w-12 h-12 text-blue-500 animate-spin"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
							/>
						</svg>
						<p class="text-gray-600 dark:text-gray-400">Loading app details...</p>
					</div>
				</div>
			{:else if app}
				<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
					<!-- Main Content -->
					<div class="lg:col-span-2 space-y-6">
						<!-- Build Progress (if generating) -->
						{#if app.status === 'generating'}
							<BuildProgress
								buildId={app.id}
								onComplete={(result) => {
									console.log('Build complete:', result);
									generatedAppsStore.fetchApps();
								}}
								onError={(err) => {
									console.error('Build error:', err);
									error = err.message;
								}}
							/>
						{/if}

						<!-- App Info Card -->
						<div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-6">
							<h2 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">App Information</h2>
							<dl class="space-y-3">
								<div>
									<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Status</dt>
									<dd class="mt-1 text-sm text-gray-900 dark:text-white capitalize">{app.status}</dd>
								</div>
								<div>
									<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Created</dt>
									<dd class="mt-1 text-sm text-gray-900 dark:text-white">
										{new Date(app.generated_at).toLocaleString()}
									</dd>
								</div>
								{#if app.deployed_at}
									<div>
										<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Deployed</dt>
										<dd class="mt-1 text-sm text-gray-900 dark:text-white">
											{new Date(app.deployed_at).toLocaleString()}
										</dd>
									</div>
								{/if}
								{#if app.custom_config?.category}
									<div>
										<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Category</dt>
										<dd class="mt-1 text-sm text-gray-900 dark:text-white">
											{app.custom_config.category}
										</dd>
									</div>
								{/if}
								{#if app.custom_config?.keywords && app.custom_config.keywords.length > 0}
									<div>
										<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Keywords</dt>
										<dd class="mt-1 flex flex-wrap gap-2">
											{#each app.custom_config.keywords as keyword}
												<span
													class="px-2 py-1 text-xs rounded-full bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300"
												>
													{keyword}
												</span>
											{/each}
										</dd>
									</div>
								{/if}
							</dl>
						</div>

						<!-- Tab Switcher (Preview / Code) -->
						{#if app.status === 'generated' || app.status === 'deployed'}
							<div class="flex items-center gap-1 bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-1">
								<button
									class="flex items-center gap-2 px-4 py-2 text-sm font-medium rounded-lg transition-colors {activeTab === 'preview' ? 'bg-blue-50 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300' : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white hover:bg-gray-50 dark:hover:bg-gray-700'}"
									onclick={() => activeTab = 'preview'}
								>
									<Eye class="w-4 h-4" />
									Preview
								</button>
								<button
									class="flex items-center gap-2 px-4 py-2 text-sm font-medium rounded-lg transition-colors {activeTab === 'code' ? 'bg-blue-50 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300' : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white hover:bg-gray-50 dark:hover:bg-gray-700'}"
									onclick={() => activeTab = 'code'}
								>
									<Code2 class="w-4 h-4" />
									Code
									{#if files.length > 0}
										<span class="text-xs bg-gray-200 dark:bg-gray-600 px-1.5 py-0.5 rounded-full">{files.length}</span>
									{/if}
								</button>
								<button
									class="flex items-center gap-2 px-4 py-2 text-sm font-medium rounded-lg transition-colors {activeTab === 'terminal' ? 'bg-blue-50 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300' : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white hover:bg-gray-50 dark:hover:bg-gray-700'}"
									onclick={() => activeTab = 'terminal'}
								>
									<Terminal class="w-4 h-4" />
									Terminal
								</button>
							</div>

							<!-- Sandbox Preview Tab -->
							{#if activeTab === 'preview'}
								<div
									class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 overflow-hidden"
									class:fixed={previewExpanded}
									class:inset-4={previewExpanded}
									class:z-50={previewExpanded}
								>
									<div class="flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-700">
										<div class="flex items-center gap-3">
											<Monitor class="w-5 h-5 text-gray-500" />
											<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Live Preview</h2>
											{#if app.sandbox}
												<SandboxStatusBadge status={app.sandbox.status} size="sm" />
											{/if}
										</div>
										<div class="flex items-center gap-2">
											{#if app.sandbox?.status === 'running' && app.sandbox.url}
												<button
													onclick={() => {
														if (app?.sandbox?.url) navigator.clipboard.writeText(app.sandbox.url);
													}}
													class="px-3 py-1.5 text-xs font-mono text-blue-600 dark:text-blue-400 bg-blue-50 dark:bg-blue-900/30 border border-blue-200 dark:border-blue-800 rounded-lg hover:bg-blue-100 dark:hover:bg-blue-900/50 transition-colors flex items-center gap-1.5 max-w-[240px]"
													title="Copy sandbox URL"
												>
													<span class="health-dot running"></span>
													<span class="truncate">{app.sandbox.url}</span>
													<ExternalLink class="w-3 h-3 flex-shrink-0" />
												</button>
											{/if}
											<OpenSandboxButton
												sandbox={app.sandbox}
												appId={app.id}
												variant="secondary"
												onStart={handleStartSandbox}
												onStop={async () => { showStopConfirm = true; }}
											/>
											{#if app.sandbox?.status === 'running'}
												<button
													onclick={() => previewExpanded = !previewExpanded}
													class="p-2 text-gray-500 hover:text-gray-700 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors"
													title={previewExpanded ? 'Exit fullscreen' : 'Fullscreen'}
												>
													{#if previewExpanded}
														<Minimize2 class="w-5 h-5" />
													{:else}
														<Maximize2 class="w-5 h-5" />
													{/if}
												</button>
											{/if}
										</div>
									</div>

									<!-- Deployment Progress Bar -->
								{#if (app.sandbox?.status === 'deploying' || app.sandbox?.status === 'pending') && sandboxDeployProgress > 0}
									<div class="h-1 bg-gray-200 dark:bg-gray-700">
										<div
											class="h-full bg-blue-500 transition-all duration-500 ease-out"
											style="width: {sandboxDeployProgress}%"
										></div>
									</div>
								{/if}

								<!-- Sandbox Error Banner -->
								{#if sandboxError}
									<div class="flex items-center gap-3 px-4 py-3 bg-red-50 dark:bg-red-900/20 border-b border-red-200 dark:border-red-800">
										<AlertTriangle class="w-4 h-4 text-red-500 flex-shrink-0" />
										<span class="text-sm text-red-700 dark:text-red-300 flex-1">{sandboxError}</span>
										<button
											onclick={handleRetrySandbox}
											class="flex items-center gap-1.5 px-3 py-1 text-xs font-medium text-red-700 dark:text-red-300 bg-red-100 dark:bg-red-900/40 rounded-md hover:bg-red-200 dark:hover:bg-red-900/60 transition-colors"
										>
											<RefreshCw class="w-3 h-3" />
											Retry
										</button>
									</div>
								{/if}

								<!-- Preview Content -->
								<div class="bg-gray-100 dark:bg-gray-900" class:h-96={!previewExpanded} class:flex-1={previewExpanded} style={previewExpanded ? 'height: calc(100% - 65px)' : ''}>
									{#if app.sandbox?.status === 'running' && app.sandbox.url}
										<iframe
											src={app.sandbox.url}
											title="{app.app_name} Preview"
											class="w-full h-full border-0"
											sandbox="allow-scripts allow-same-origin allow-forms allow-popups"
										></iframe>
									{:else}
										<div class="flex flex-col items-center justify-center h-full text-gray-500 dark:text-gray-400">
											{#if app.sandbox?.status === 'pending' || app.sandbox?.status === 'deploying'}
												<div class="w-10 h-10 border-[3px] border-blue-500 border-t-transparent rounded-full animate-spin mb-4"></div>
												<p class="text-sm font-medium">Deploying sandbox...</p>
												{#if sandboxDeployProgress > 0}
													<p class="text-xs text-gray-400 mt-1">{sandboxDeployProgress}% complete</p>
												{/if}
											{:else if app.sandbox?.status === 'failed'}
												<AlertTriangle class="w-12 h-12 mb-3 text-red-400 opacity-70" />
												<p class="text-sm font-medium text-red-400">Sandbox failed to deploy</p>
												<button
													onclick={handleRetrySandbox}
													class="mt-4 px-4 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded-lg transition-colors flex items-center gap-2"
												>
													<RefreshCw class="w-4 h-4" />
													Retry Deployment
												</button>
											{:else}
												<Monitor class="w-12 h-12 mb-3 opacity-50" />
												<p class="text-sm font-medium">Start the sandbox to see a live preview</p>
												<button
													onclick={handleStartSandbox}
													class="mt-4 px-4 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded-lg transition-colors"
												>
													Start Sandbox
												</button>
											{/if}
										</div>
									{/if}
								</div>
							</div>

							<!-- Stop Confirmation Dialog -->
							{#if showStopConfirm}
								<div class="fixed inset-0 z-[200] flex items-center justify-center bg-black/40">
									<div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-6 max-w-sm mx-4 shadow-xl">
										<h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-2">Stop Sandbox?</h3>
										<p class="text-sm text-gray-600 dark:text-gray-400 mb-4">
											This will stop the running sandbox. The live preview will become unavailable until you restart it.
										</p>
										<div class="flex justify-end gap-3">
											<button
												onclick={() => showStopConfirm = false}
												class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 rounded-lg transition-colors"
											>
												Cancel
											</button>
											<button
												onclick={handleStopSandbox}
												class="px-4 py-2 text-sm font-medium text-white bg-red-600 hover:bg-red-700 rounded-lg transition-colors"
											>
												Stop Sandbox
											</button>
										</div>
									</div>
								</div>
							{/if}
						{/if}

							<!-- Terminal Tab -->
							{#if activeTab === 'terminal'}
								<div class="bg-gray-900 rounded-xl border border-gray-700 overflow-hidden" style="height: 600px;">
									<div class="flex items-center gap-2 px-4 py-3 border-b border-gray-700 bg-gray-800/50">
										<div class="flex gap-1.5">
											<span class="w-3 h-3 rounded-full bg-red-500"></span>
											<span class="w-3 h-3 rounded-full bg-yellow-500"></span>
											<span class="w-3 h-3 rounded-full bg-green-500"></span>
										</div>
										<span class="text-sm text-gray-400 ml-2">Terminal</span>
									</div>
									<div class="flex flex-col items-center justify-center h-[calc(100%-48px)] text-gray-500">
										<Terminal class="w-12 h-12 mb-3 opacity-30" />
										<p class="text-sm font-medium">Terminal integration coming soon</p>
										<p class="text-xs text-gray-600 mt-1">
											{#if app?.sandbox?.status === 'running'}
												Sandbox is running — terminal access will connect to your sandbox
											{:else}
												Start a sandbox to enable terminal access
											{/if}
										</p>
									</div>
								</div>
							{/if}

							<!-- Code Workspace Tab -->
							{#if activeTab === 'code'}
								<div class="bg-gray-900 rounded-xl border border-gray-700 overflow-hidden" style="height: 600px;">
									<div class="flex h-full">
										<!-- File Tree Sidebar -->
										<div class="w-64 flex-shrink-0 border-r border-gray-700 flex flex-col bg-gray-800/50">
											<div class="flex items-center justify-between px-3 py-2 border-b border-gray-700">
												<span class="text-xs font-semibold text-gray-400 uppercase tracking-wider">Explorer</span>
												<span class="text-xs text-gray-500 bg-gray-700/50 px-1.5 py-0.5 rounded-full">{files.length}</span>
											</div>
											{#if files.length > 6}
												<div class="px-2 py-2 border-b border-gray-700">
													<div class="relative">
														<Search class="absolute left-2 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-gray-500" />
														<input
															type="text"
															placeholder="Search files..."
															bind:value={fileSearchQuery}
															class="w-full pl-7 pr-2 py-1.5 text-xs bg-gray-700/50 border border-gray-600 rounded text-gray-200 placeholder-gray-500 focus:border-blue-500 focus:outline-none"
														/>
													</div>
												</div>
											{/if}
											<div class="flex-1 overflow-y-auto p-1">
												{#if filteredTreeNodes.length > 0}
													<FileTree
														nodes={filteredTreeNodes}
														{selectedFile}
														onFileSelect={handleFileSelect}
													/>
												{:else if files.length === 0}
													<div class="flex flex-col items-center justify-center h-full text-gray-500 text-sm">
														<Code2 class="w-8 h-8 mb-2 opacity-50" />
														<p>No files generated yet</p>
													</div>
												{:else}
													<div class="p-3 text-center text-xs text-gray-500">
														No files match "{fileSearchQuery}"
													</div>
												{/if}
											</div>
										</div>

										<!-- Editor Panel -->
										<div class="flex-1 flex flex-col min-w-0">
											{#if saveError}
												<div class="flex items-center gap-2 px-3 py-1.5 bg-red-900/30 border-b border-red-800 text-xs text-red-300">
													<AlertTriangle class="w-3.5 h-3.5 flex-shrink-0" />
													<span class="flex-1">{saveError}</span>
													<button onclick={() => saveError = null} class="text-red-400 hover:text-red-200">Dismiss</button>
												</div>
											{/if}
											{#if selectedFile}
												<EditorToolbar
													filename={selectedFile.path || selectedFile.name}
													{isEditing}
													{isDirty}
													readonly={!isEditing}
													{cursorLine}
													{cursorColumn}
													onToggleEdit={handleToggleEdit}
													onSave={() => handleSave(editorValue)}
													onCopy={handleCopy}
												/>

												<div
													class="flex-1 overflow-hidden transition-opacity duration-150"
													class:opacity-30={isTransitioning}
													class:editor-editing={isEditing}
												>
													{#if fileLoading}
														<div class="flex items-center justify-center h-full text-gray-400">
															<div class="flex flex-col items-center gap-3">
																<div class="w-6 h-6 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
																<span class="text-sm">Loading file...</span>
															</div>
														</div>
													{:else}
														<MonacoEditor
															bind:this={editorRef}
															bind:value={editorValue}
															filename={selectedFile.path || selectedFile.name}
															readonly={!isEditing}
															onSave={handleSave}
															onChange={handleEditorChange}
														/>
													{/if}
												</div>

												<EditorStatusBar
													{languageId}
													isReadonly={!isEditing}
													{isEditing}
												/>
											{:else}
												<div class="flex flex-col items-center justify-center h-full text-gray-500">
													<Code2 class="w-12 h-12 mb-3 opacity-30" />
													<p class="text-sm font-medium">Select a file to view</p>
													<p class="text-xs text-gray-600 mt-1">Choose a file from the explorer</p>
												</div>
											{/if}
										</div>
									</div>
								</div>
							{/if}
						{/if}

						<!-- Error Message (if failed) -->
						{#if app.status === 'failed' && app.error_message}
							<div class="bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-800 rounded-xl p-6">
								<h2 class="text-lg font-semibold text-red-900 dark:text-red-200 mb-2 flex items-center gap-2">
									<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											stroke-width="2"
											d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
										/>
									</svg>
									Build Failed
								</h2>
								<p class="text-sm text-red-700 dark:text-red-300">{app.error_message}</p>
							</div>
						{/if}
					</div>

					<!-- Sidebar -->
					<div class="space-y-6">
						<!-- Actions Card -->
						<div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-6">
							<h2 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">Actions</h2>
							<AppActions {app} onDeploy={handleDeploy} onDelete={handleDelete} />
						</div>

						<!-- Version History Card -->
						{#if app.status === 'generated' || app.status === 'deployed'}
							<div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-6">
								<div class="flex items-center justify-between mb-4">
									<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Versions</h2>
									{#if versions.length > 0}
										<span class="text-xs text-gray-500 bg-gray-100 dark:bg-gray-700 px-2 py-0.5 rounded-full">
											v{currentVersion}
										</span>
									{/if}
								</div>
								{#if versions.length > 0}
									<div class="space-y-3">
										{#each versions.slice(0, 3) as version (version.id)}
											<button
												onclick={() => handleVersionPreview(version)}
												class="w-full text-left px-3 py-2 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors group"
											>
												<div class="flex items-center justify-between">
													<span class="text-sm font-medium text-gray-900 dark:text-white">
														v{version.versionNumber}
														{#if version.isCurrent}
															<span class="text-xs text-green-600 dark:text-green-400 ml-1">current</span>
														{/if}
													</span>
												</div>
												{#if version.label}
													<p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5 truncate">{version.label}</p>
												{/if}
											</button>
										{/each}
										<button
											onclick={() => timelinePanelOpen = true}
											class="w-full flex items-center justify-center gap-2 px-3 py-2 text-sm font-medium text-blue-600 dark:text-blue-400 hover:bg-blue-50 dark:hover:bg-blue-900/20 rounded-lg transition-colors"
										>
											<History class="w-4 h-4" />
											View all versions
										</button>
									</div>
								{:else}
									<p class="text-sm text-gray-500 dark:text-gray-400">No version history yet</p>
									<button
										onclick={() => saveModalOpen = true}
										class="mt-3 w-full px-3 py-2 text-sm font-medium text-blue-600 dark:text-blue-400 border border-blue-200 dark:border-blue-800 rounded-lg hover:bg-blue-50 dark:hover:bg-blue-900/20 transition-colors"
									>
										Save first version
									</button>
								{/if}
							</div>
						{/if}
					</div>
				</div>
			{/if}
		</div>
	</div>
</div>

<!-- Version History Panels & Modals -->
{#if app}
	<VersionTimelinePanel
		appId={appId}
		{versions}
		isOpen={timelinePanelOpen}
		isLoading={versionsLoading}
		onClose={() => timelinePanelOpen = false}
		onPreview={handleVersionPreview}
		onRestore={handleVersionRestore}
		onCompare={handleCompareVersions}
	/>

	<SaveVersionModal
		appId={appId}
		{currentVersion}
		isOpen={saveModalOpen}
		isSaving={isSavingVersion}
		onClose={() => saveModalOpen = false}
		onSave={handleSaveVersion}
	/>

	{#if previewVersion}
		<VersionPreviewModal
			version={previewVersion}
			isOpen={previewModalOpen}
			onClose={() => previewModalOpen = false}
			onRestore={handleVersionRestore}
		/>
	{/if}

	{#if restoreVersion}
		<RestoreConfirmDialog
			version={restoreVersion}
			{currentVersion}
			isOpen={restoreDialogOpen}
			{isRestoring}
			onClose={() => restoreDialogOpen = false}
			onConfirm={handleConfirmRestore}
		/>
	{/if}

	{#if diffModalOpen && diffFromVersion && diffToVersion}
		<VersionDiffModal
			isOpen={diffModalOpen}
			workspaceId={app.workspace_id}
			fromVersion={diffFromVersion.version_number}
			toVersion={diffToVersion.version_number}
			fromDisplayNum={extractDisplayNumber(diffFromVersion.version_number)}
			toDisplayNum={extractDisplayNumber(diffToVersion.version_number)}
			onClose={() => diffModalOpen = false}
		/>
	{/if}
{/if}

<style>
	.editor-editing {
		box-shadow:
			inset 0 0 0 1px rgba(99, 102, 241, 0.3),
			inset 0 0 20px rgba(99, 102, 241, 0.05);
	}

	.health-dot {
		width: 6px;
		height: 6px;
		border-radius: 50%;
		flex-shrink: 0;
	}

	.health-dot.running {
		background: #22c55e;
		box-shadow: 0 0 4px rgba(34, 197, 94, 0.5);
		animation: health-pulse 2s ease-in-out infinite;
	}

	@keyframes health-pulse {
		0%, 100% { opacity: 1; }
		50% { opacity: 0.5; }
	}
</style>
