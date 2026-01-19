<script lang="ts">
	import { onMount } from 'svelte';
	import { apiClient } from '$lib/api';
	import { fly } from 'svelte/transition';

	// Types
	export interface AIModel {
		id: string;
		name: string;
		provider: string;
		type?: 'local' | 'cloud';
		size?: string;
		capabilities?: string[];
	}

	export interface ModelSelectorProps {
		selectedModelId: string;
		onModelSelect: (modelId: string) => void;
		variant?: 'compact' | 'full' | 'icon-only';
	}

	// Props with defaults
	let {
		selectedModelId = $bindable(''),
		onModelSelect,
		variant = 'compact'
	}: ModelSelectorProps = $props();

	// State
	let isOpen = $state(false);
	let allModels = $state<AIModel[]>([]);
	let loadingModels = $state(false);
	let dropdownRef = $state<HTMLDivElement | null>(null);

	// Derived state
	let selectedModel = $derived(() => {
		return allModels.find((m) => m.id === selectedModelId);
	});

	let displayName = $derived(() => {
		if (!selectedModel()) return 'Select Model';
		const model = selectedModel()!;
		// For compact variant, show shortened name (first part before colon)
		if (variant === 'compact') {
			const parts = model.name.split(':');
			return parts[0];
		}
		return model.name;
	});

	let localModels = $derived(() => {
		return allModels.filter((m) => m.type === 'local' || m.provider === 'ollama_local');
	});

	let cloudModels = $derived(() => {
		return allModels.filter((m) => m.type === 'cloud' || (m.provider !== 'ollama_local' && m.provider !== 'ollama'));
	});

	// Capability badge info - clean minimal design
	const capabilityInfo: Record<string, { label: string }> = {
		'vision': { label: 'Vision' },
		'tools': { label: 'Tools' },
		'qycz': { label: 'QYCZ' },
		'multi-lang': { label: 'Multi' },
		'sV-8': { label: 'SV-8' },
		'4k-Fast': { label: 'Fast' }
	};

	// Functions
	async function loadModels() {
		loadingModels = true;
		try {
			// Get provider info for default model
			let defaultModelId = '';
			const providersRes = await apiClient.get('/ai/providers');
			if (providersRes.ok) {
				const data = await providersRes.json();
				defaultModelId = data.default_model || '';
			}

			// Get all available models
			const response = await apiClient.get('/ai/models');
			if (response.ok) {
				const data = await response.json();
				const models: AIModel[] = data.models || [];

				// Categorize models
				allModels = models.map((m: any) => ({
					...m,
					type: (m.provider === 'ollama_local' || m.provider === 'ollama') ? 'local' : 'cloud'
				}));

				// Set default if none selected
				if (!selectedModelId && allModels.length > 0) {
					const defaultModel = allModels.find((m) => m.id === defaultModelId) || allModels[0];
					selectedModelId = defaultModel.id;
					onModelSelect(defaultModel.id);
				}
			}
		} catch (e) {
			console.error('Failed to load models:', e);
		} finally {
			loadingModels = false;
		}
	}

	function toggleDropdown() {
		isOpen = !isOpen;
	}

	function selectModel(modelId: string) {
		selectedModelId = modelId;
		onModelSelect(modelId);
		isOpen = false;
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') {
			isOpen = false;
		}
	}

	function handleClickOutside(e: MouseEvent) {
		if (dropdownRef && !dropdownRef.contains(e.target as Node)) {
			isOpen = false;
		}
	}

	onMount(() => {
		loadModels();
		document.addEventListener('click', handleClickOutside);
		return () => {
			document.removeEventListener('click', handleClickOutside);
		};
	});
</script>

<div class="model-selector" class:icon-only={variant === 'icon-only'} bind:this={dropdownRef}>
	<!-- Trigger Button -->
	<button
		class="selector-trigger"
		class:icon-only-trigger={variant === 'icon-only'}
		onclick={toggleDropdown}
		type="button"
		title={variant === 'icon-only' ? displayName() : undefined}
	>
		{#if variant === 'icon-only'}
			<!-- Icon Only Mode -->
			{#if selectedModel()}
				{#if selectedModel()!.type === 'local'}
					<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<rect x="2" y="3" width="20" height="14" rx="2"/>
						<path d="M8 21h8M12 17v4"/>
					</svg>
				{:else}
					<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M18 10h-1.26A8 8 0 1 0 9 20h9a5 5 0 0 0 0-10z"/>
					</svg>
				{/if}
			{:else}
				<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<circle cx="12" cy="12" r="3"/>
					<path d="M12 1v6m0 6v6"/>
				</svg>
			{/if}
		{:else}
			<!-- Full/Compact Mode -->
			<div class="trigger-content">
				{#if selectedModel()}
					<span class="model-icon">
						{#if selectedModel()!.type === 'local'}
							<!-- Computer icon for local models -->
							<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<rect x="2" y="3" width="20" height="14" rx="2"/>
								<path d="M8 21h8M12 17v4"/>
							</svg>
						{:else}
							<!-- Cloud icon for cloud models -->
							<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M18 10h-1.26A8 8 0 1 0 9 20h9a5 5 0 0 0 0-10z"/>
							</svg>
						{/if}
					</span>
				{/if}
				<span class="model-name">{displayName()}</span>
				<svg
					class="chevron"
					class:rotate={isOpen}
					width="12"
					height="12"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
				>
					<polyline points="6 9 12 15 18 9"></polyline>
				</svg>
			</div>
		{/if}
	</button>

	<!-- Dropdown Menu -->
	{#if isOpen}
		<div class="dropdown-menu" role="menu" onkeydown={handleKeydown} transition:fly={{ y: -10, duration: 200 }}>
			{#if loadingModels}
				<div class="loading-state">Loading models...</div>
			{:else if allModels.length === 0}
				<div class="empty-state">
					<p>No models available</p>
					<a href="/settings/ai" class="settings-link-empty">Configure in AI Settings</a>
				</div>
			{:else}
				<!-- SELECTED Section -->
				{#if selectedModel()}
					{@const model = selectedModel()!}
					{@const caps = model.capabilities || []}
					{@const isCloud = model.type === 'cloud'}
					<div class="section-header selected-header">
						<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M5 13l4 4L19 7" />
						</svg>
						SELECTED
					</div>
					<button
						class="model-item selected-item"
						onclick={() => { isOpen = false; }}
						type="button"
					>
						<div class="model-content">
							<div class="model-icon-wrapper">
								{#if isCloud}
									<svg class="model-svg cloud" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<path d="M18 10h-1.26A8 8 0 1 0 9 20h9a5 5 0 0 0 0-10z"/>
									</svg>
								{:else}
									<svg class="model-svg local" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<rect x="2" y="3" width="20" height="14" rx="2"/>
										<path d="M8 21h8M12 17v4"/>
									</svg>
								{/if}
							</div>
							<div class="model-info-wrapper">
								<div class="model-name-row">
									<span class="model-title">{model.name}</span>
									<span class="model-badge {isCloud ? 'cloud-badge' : 'local-badge'}">
										{isCloud ? 'Cloud' : 'Local'}
									</span>
								</div>
								{#if model.size}
									<div class="model-size">{model.size}</div>
								{/if}
								{#if caps.length > 0}
									<div class="capabilities-row">
										{#each caps.slice(0, 3) as cap}
											{@const info = capabilityInfo[cap]}
											{#if info}
												<span class="capability-badge">
													{info.label}
												</span>
											{/if}
										{/each}
										{#if caps.length > 3}
											<span class="capability-badge capability-more">
												+{caps.length - 3}
											</span>
										{/if}
									</div>
								{/if}
							</div>
							<svg class="checkmark" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M5 13l4 4L19 7" />
							</svg>
						</div>
					</button>
				{/if}

				<!-- LOCAL MODELS Section -->
				{#if localModels().length > 0}
					<div class="section-divider"></div>
					<div class="section-header">
						<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<rect x="2" y="3" width="20" height="14" rx="2"/>
							<path d="M8 21h8M12 17v4"/>
						</svg>
						LOCAL (OLLAMA)
					</div>
					{#each localModels().filter(m => m.id !== selectedModelId) as model}
						{@const caps = model.capabilities || []}
						<button
							class="model-item"
							onclick={() => selectModel(model.id)}
							type="button"
						>
							<div class="model-content">
								<div class="model-icon-wrapper">
									<svg class="model-svg local" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<rect x="2" y="3" width="20" height="14" rx="2"/>
										<path d="M8 21h8M12 17v4"/>
									</svg>
								</div>
								<div class="model-info-wrapper">
									<div class="model-name-row">
										<span class="model-title">{model.name}</span>
										<span class="model-badge local-badge">Local</span>
									</div>
									{#if model.size}
										<div class="model-size">{model.size}</div>
									{/if}
									{#if caps.length > 0}
										<div class="capabilities-row">
											{#each caps.slice(0, 3) as cap}
												{@const info = capabilityInfo[cap]}
												{#if info}
													<span class="capability-badge">
														{info.label}
													</span>
												{/if}
											{/each}
											{#if caps.length > 3}
												<span class="capability-badge capability-more">
													+{caps.length - 3}
												</span>
											{/if}
										</div>
									{/if}
								</div>
							</div>
						</button>
					{/each}
				{/if}

				<!-- CLOUD MODELS Section -->
				{#if cloudModels().length > 0}
					<div class="section-divider"></div>
					<div class="section-header">
						<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M18 10h-1.26A8 8 0 1 0 9 20h9a5 5 0 0 0 0-10z"/>
						</svg>
						CLOUD MODELS
					</div>
					{#if cloudModels()[0]?.provider}
						<div class="provider-name">{cloudModels()[0].provider}</div>
					{/if}
					{#each cloudModels().filter(m => m.id !== selectedModelId) as model}
						{@const caps = model.capabilities || []}
						<button
							class="model-item"
							onclick={() => selectModel(model.id)}
							type="button"
						>
							<div class="model-content">
								<div class="model-icon-wrapper">
									<svg class="model-svg cloud" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<path d="M18 10h-1.26A8 8 0 1 0 9 20h9a5 5 0 0 0 0-10z"/>
									</svg>
								</div>
								<div class="model-info-wrapper">
									<div class="model-name-row">
										<span class="model-title">{model.name}</span>
									</div>
									{#if model.size}
										<div class="model-size">{model.size}</div>
									{/if}
									{#if caps.length > 0}
										<div class="capabilities-row">
											{#each caps.slice(0, 3) as cap}
												{@const info = capabilityInfo[cap]}
												{#if info}
													<span class="capability-badge">
														{info.label}
													</span>
												{/if}
											{/each}
											{#if caps.length > 3}
												<span class="capability-badge capability-more">
													+{caps.length - 3}
												</span>
											{/if}
										</div>
									{/if}
								</div>
							</div>
						</button>
					{/each}
				{/if}

				<!-- Settings Footer -->
				<div class="section-divider"></div>
				<a href="/settings/ai" class="settings-link">
					<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<circle cx="12" cy="12" r="3"/>
						<path d="M12 1v6m0 6v6m8.66-14.66l-4.24 4.24m-5.66 5.66l-4.24 4.24M23 12h-6m-6 0H1m18.66 8.66l-4.24-4.24m-5.66-5.66L4.34 4.34"></path>
					</svg>
					<span>Model Settings</span>
				</a>
			{/if}
		</div>
	{/if}
</div>

<style>
	.model-selector {
		position: relative;
		width: 100%;
	}

	.model-selector.icon-only {
		width: auto;
	}

	/* Trigger Button - Glassy Design */
	.selector-trigger {
		display: flex;
		align-items: center;
		width: 100%;
		padding: 8px 14px;
		background: rgba(255, 255, 255, 0.05);
		backdrop-filter: blur(10px);
		border: 1px solid rgba(255, 255, 255, 0.1);
		border-radius: 10px;
		cursor: pointer;
		transition: all 0.2s ease;
		color: rgba(255, 255, 255, 0.9);
	}

	.selector-trigger:hover {
		background: rgba(255, 255, 255, 0.08);
		border-color: rgba(255, 255, 255, 0.15);
	}

	/* Icon Only Variant */
	.selector-trigger.icon-only-trigger {
		width: 36px;
		height: 36px;
		padding: 0;
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: 8px;
	}

	.selector-trigger.icon-only-trigger:hover {
		background: rgba(255, 255, 255, 0.1);
	}

	.trigger-content {
		display: flex;
		align-items: center;
		gap: 8px;
		width: 100%;
	}

	.model-icon {
		display: flex;
		align-items: center;
		color: rgba(255, 255, 255, 0.7);
	}

	.model-name {
		flex: 1;
		font-size: 13px;
		font-weight: 500;
		text-align: left;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.chevron {
		color: rgba(255, 255, 255, 0.5);
		transition: transform 0.2s ease;
	}

	.chevron.rotate {
		transform: rotate(180deg);
	}

	/* Dropdown Menu - Dark Glassmorphic */
	.dropdown-menu {
		position: absolute;
		bottom: calc(100% + 8px);
		left: 50%;
		transform: translateX(-50%);
		width: max-content;
		min-width: 320px;
		max-width: 400px;
		background: rgba(28, 28, 30, 0.95);
		backdrop-filter: blur(20px);
		border: 1px solid rgba(255, 255, 255, 0.1);
		border-radius: 14px;
		box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5), 0 0 0 1px rgba(255, 255, 255, 0.05);
		z-index: 10000;
		overflow: hidden;
		max-height: 480px;
		display: flex;
		flex-direction: column;
	}

	/* Loading & Empty States */
	.loading-state,
	.empty-state {
		padding: 32px 24px;
		text-align: center;
		color: rgba(255, 255, 255, 0.5);
		font-size: 14px;
	}

	.empty-state p {
		margin: 0 0 12px 0;
	}

	.settings-link-empty {
		font-size: 12px;
		color: #3b82f6;
		text-decoration: none;
	}

	.settings-link-empty:hover {
		text-decoration: underline;
	}

	/* Section Headers */
	.section-header {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 12px 16px 8px;
		font-size: 10px;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 0.8px;
		color: rgba(255, 255, 255, 0.4);
	}

	.section-header.selected-header {
		color: #3b82f6;
	}

	.section-header svg {
		width: 12px;
		height: 12px;
	}

	.section-divider {
		height: 1px;
		background: rgba(255, 255, 255, 0.08);
		margin: 4px 0;
	}

	.provider-name {
		padding: 4px 16px 8px;
		font-size: 11px;
		color: rgba(255, 255, 255, 0.5);
	}

	/* Model Items */
	.model-item {
		display: block;
		width: 100%;
		padding: 0;
		background: transparent;
		border: none;
		text-align: left;
		cursor: pointer;
		transition: background 0.15s ease;
	}

	.model-item:hover {
		background: rgba(255, 255, 255, 0.05);
	}

	.model-item.selected-item {
		background: rgba(59, 130, 246, 0.12);
	}

	.model-item.selected-item:hover {
		background: rgba(59, 130, 246, 0.15);
	}

	.model-content {
		display: flex;
		align-items: flex-start;
		gap: 12px;
		padding: 12px 16px;
	}

	.model-icon-wrapper {
		flex-shrink: 0;
		margin-top: 2px;
	}

	.model-svg {
		width: 16px;
		height: 16px;
	}

	.model-svg.local {
		color: #22c55e;
	}

	.model-svg.cloud {
		color: #3b82f6;
	}

	.model-info-wrapper {
		flex: 1;
		min-width: 0;
	}

	.model-name-row {
		display: flex;
		align-items: center;
		gap: 8px;
		flex-wrap: wrap;
		margin-bottom: 4px;
	}

	.model-title {
		font-size: 13px;
		font-weight: 500;
		color: rgba(255, 255, 255, 0.95);
	}

	.model-badge {
		display: inline-flex;
		align-items: center;
		padding: 2px 7px;
		font-size: 9px;
		font-weight: 600;
		border-radius: 4px;
		text-transform: uppercase;
		letter-spacing: 0.3px;
	}

	.model-badge.local-badge {
		background: rgba(34, 197, 94, 0.15);
		color: #22c55e;
	}

	.model-badge.cloud-badge {
		background: rgba(59, 130, 246, 0.15);
		color: #3b82f6;
	}

	.model-size {
		font-size: 11px;
		color: rgba(255, 255, 255, 0.4);
		margin-bottom: 6px;
	}

	.capabilities-row {
		display: flex;
		flex-wrap: wrap;
		gap: 4px;
	}

	.capability-badge {
		display: inline-flex;
		align-items: center;
		padding: 2px 6px;
		font-size: 9px;
		font-weight: 500;
		border-radius: 3px;
		background: rgba(255, 255, 255, 0.05);
		border: 1px solid rgba(255, 255, 255, 0.1);
		color: rgba(255, 255, 255, 0.6);
		text-transform: uppercase;
		letter-spacing: 0.3px;
	}

	.capability-badge.capability-more {
		background: rgba(255, 255, 255, 0.03);
		color: rgba(255, 255, 255, 0.4);
		font-size: 8px;
	}

	.checkmark {
		width: 16px;
		height: 16px;
		color: #3b82f6;
		flex-shrink: 0;
		margin-top: 2px;
	}

	/* Settings Link */
	.settings-link {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 12px 16px;
		font-size: 12px;
		font-weight: 500;
		color: rgba(255, 255, 255, 0.6);
		text-decoration: none;
		transition: all 0.15s ease;
	}

	.settings-link:hover {
		background: rgba(255, 255, 255, 0.05);
		color: rgba(255, 255, 255, 0.9);
	}

	.settings-link svg {
		width: 14px;
		height: 14px;
	}

	/* Scrollbar */
	.dropdown-menu {
		overflow-y: auto;
	}

	.dropdown-menu::-webkit-scrollbar {
		width: 6px;
	}

	.dropdown-menu::-webkit-scrollbar-track {
		background: rgba(255, 255, 255, 0.03);
	}

	.dropdown-menu::-webkit-scrollbar-thumb {
		background: rgba(255, 255, 255, 0.15);
		border-radius: 3px;
	}

	.dropdown-menu::-webkit-scrollbar-thumb:hover {
		background: rgba(255, 255, 255, 0.25);
	}
</style>
