<script lang="ts">
	import { goto } from '$app/navigation';
	import { ArrowLeft, ArrowRight, Save, Eye } from 'lucide-svelte';
	import { customModulesStore } from '$lib/stores/customModulesStore';
	import ModuleEditor from '$lib/components/modules/ModuleEditor.svelte';
	import ActionBuilder from '$lib/components/modules/ActionBuilder.svelte';
	import ManifestViewer from '$lib/components/modules/ManifestViewer.svelte';
	import type { ModuleCategory, ModuleAction, ModuleManifest } from '$lib/types/modules';

	let store = $state(customModulesStore);

	let currentStep = $state(1);
	let isCreating = $state(false);

	// Step 1: Basic Info
	let name = $state('');
	let description = $state('');
	let category = $state<ModuleCategory>('custom');
	let icon = $state('📦');

	// Step 2: Actions
	let actions = $state<ModuleAction[]>([]);

	// Step 3: Configuration
	let configSchema = $state<Record<string, unknown>>({});
	let configSchemaText = $state('{}');
	let visibility = $state<'private' | 'workspace' | 'public'>('private');

	const manifest = $derived<ModuleManifest>({
		name,
		version: '1.0.0',
		description,
		author: 'Current User',
		category,
		icon,
		actions,
		config_schema: configSchema,
		dependencies: [],
		permissions: []
	});

	function handleNext() {
		if (currentStep === 1 && !name.trim()) {
			alert('Please enter a module name');
			return;
		}
		if (currentStep === 1 && !description.trim()) {
			alert('Please enter a description');
			return;
		}
		if (currentStep < 3) {
			currentStep++;
		}
	}

	function handlePrevious() {
		if (currentStep > 1) {
			currentStep--;
		}
	}

	function handleConfigSchemaChange(value: string) {
		configSchemaText = value;
		try {
			configSchema = JSON.parse(value);
		} catch {
			// Invalid JSON, keep old value
		}
	}

	async function handleCreate(isDraft: boolean) {
		if (!name.trim() || !description.trim()) {
			alert('Please fill in all required fields');
			return;
		}

		isCreating = true;
		const module = await store.createModule({
			name,
			description,
			category,
			icon,
			manifest,
			config_schema: configSchema,
			visibility
		});

		if (module) {
			goto(`/modules/${module.id}`);
		}
		isCreating = false;
	}
</script>

<div class="h-full flex flex-col bg-white">
	<!-- Header -->
	<div class="flex-shrink-0 border-b border-gray-200 bg-white px-8 py-6">
		<!-- Back Button -->
		<button
			onclick={() => goto('/modules')}
			class="flex items-center gap-2 text-sm text-gray-600 hover:text-gray-900 mb-4"
		>
			<ArrowLeft class="w-4 h-4" />
			<span>Back to Modules</span>
		</button>

		<!-- Title -->
		<h1 class="text-2xl font-bold text-gray-900 mb-2">Create Custom Module</h1>
		<p class="text-sm text-gray-600">Build a custom module for your workspace</p>

		<!-- Step Indicator -->
		<div class="flex items-center gap-4 mt-6">
			{#each [1, 2, 3] as step}
				<div class="flex items-center gap-2">
					<div class="flex items-center gap-2">
						<div class="w-8 h-8 rounded-full flex items-center justify-center text-sm font-medium {currentStep === step ? 'bg-blue-600 text-white' : currentStep > step ? 'bg-green-600 text-white' : 'bg-gray-200 text-gray-600'}">
							{step}
						</div>
						<span class="text-sm font-medium {currentStep === step ? 'text-gray-900' : 'text-gray-600'}">
							{step === 1 ? 'Basic Info' : step === 2 ? 'Actions' : 'Configuration'}
						</span>
					</div>
					{#if step < 3}
						<div class="w-12 h-0.5 {currentStep > step ? 'bg-green-600' : 'bg-gray-200'}"></div>
					{/if}
				</div>
			{/each}
		</div>
	</div>

	<!-- Content -->
	<div class="flex-1 overflow-y-auto">
		<div class="max-w-4xl mx-auto px-8 py-6">
			{#if currentStep === 1}
				<!-- Step 1: Basic Info -->
				<div class="space-y-6">
					<ModuleEditor
						{name}
						{description}
						{category}
						{icon}
						onNameChange={(v) => name = v}
						onDescriptionChange={(v) => description = v}
						onCategoryChange={(v) => category = v}
						onIconChange={(v) => icon = v}
					/>
				</div>
			{:else if currentStep === 2}
				<!-- Step 2: Actions -->
				<div class="space-y-6">
					<div>
						<h2 class="text-lg font-semibold text-gray-900 mb-2">Module Actions</h2>
						<p class="text-sm text-gray-600 mb-4">Define the actions your module provides</p>
					</div>
					<ActionBuilder
						{actions}
						onActionsChange={(a) => actions = a}
					/>
				</div>
			{:else if currentStep === 3}
				<!-- Step 3: Configuration -->
				<div class="space-y-6">
					<div>
						<h2 class="text-lg font-semibold text-gray-900 mb-2">Configuration</h2>
						<p class="text-sm text-gray-600 mb-4">Set up module configuration and visibility</p>
					</div>

					<!-- Config Schema -->
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-2">
							Configuration Schema (JSON)
						</label>
						<textarea
							value={configSchemaText}
							oninput={(e) => handleConfigSchemaChange(e.currentTarget.value)}
							placeholder={'{"setting1": "string", "setting2": "number"}'}
							rows="6"
							class="w-full px-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 font-mono text-sm"
						></textarea>
					</div>

					<!-- Visibility -->
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-2">
							Visibility
						</label>
						<div class="space-y-2">
							<label class="flex items-center p-4 border rounded-lg cursor-pointer {visibility === 'private' ? 'border-blue-500 bg-blue-50' : 'border-gray-200'}">
								<input
									type="radio"
									name="visibility"
									value="private"
									checked={visibility === 'private'}
									onchange={() => visibility = 'private'}
									class="mr-3"
								/>
								<div>
									<p class="font-medium text-gray-900">Private</p>
									<p class="text-sm text-gray-600">Only you can see and use this module</p>
								</div>
							</label>
							<label class="flex items-center p-4 border rounded-lg cursor-pointer {visibility === 'workspace' ? 'border-blue-500 bg-blue-50' : 'border-gray-200'}">
								<input
									type="radio"
									name="visibility"
									value="workspace"
									checked={visibility === 'workspace'}
									onchange={() => visibility = 'workspace'}
									class="mr-3"
								/>
								<div>
									<p class="font-medium text-gray-900">Workspace</p>
									<p class="text-sm text-gray-600">Available to your entire workspace</p>
								</div>
							</label>
							<label class="flex items-center p-4 border rounded-lg cursor-pointer {visibility === 'public' ? 'border-blue-500 bg-blue-50' : 'border-gray-200'}">
								<input
									type="radio"
									name="visibility"
									value="public"
									checked={visibility === 'public'}
									onchange={() => visibility = 'public'}
									class="mr-3"
								/>
								<div>
									<p class="font-medium text-gray-900">Public</p>
									<p class="text-sm text-gray-600">Available to everyone</p>
								</div>
							</label>
						</div>
					</div>

					<!-- Manifest Preview -->
					<div>
						<div class="flex items-center gap-2 mb-2">
							<Eye class="w-4 h-4 text-gray-600" />
							<label class="text-sm font-medium text-gray-700">
								Manifest Preview
							</label>
						</div>
						<ManifestViewer {manifest} />
					</div>
				</div>
			{/if}
		</div>
	</div>

	<!-- Footer Actions -->
	<div class="flex-shrink-0 border-t border-gray-200 bg-gray-50 px-8 py-4">
		<div class="max-w-4xl mx-auto flex items-center justify-between">
			<div>
				{#if currentStep > 1}
					<button
						onclick={handlePrevious}
						class="flex items-center gap-2 px-4 py-2 text-gray-700 hover:bg-gray-100 rounded-lg transition-colors"
					>
						<ArrowLeft class="w-4 h-4" />
						<span>Previous</span>
					</button>
				{/if}
			</div>
			<div class="flex items-center gap-3">
				{#if currentStep === 3}
					<button
						onclick={() => handleCreate(true)}
						disabled={isCreating}
						class="flex items-center gap-2 px-4 py-2 border border-gray-300 text-gray-700 hover:bg-gray-100 rounded-lg transition-colors disabled:opacity-50"
					>
						<Save class="w-4 h-4" />
						<span>Save Draft</span>
					</button>
					<button
						onclick={() => handleCreate(false)}
						disabled={isCreating}
						class="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50"
					>
						<span>{isCreating ? 'Creating...' : 'Create Module'}</span>
					</button>
				{:else}
					<button
						onclick={handleNext}
						class="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
					>
						<span>Next</span>
						<ArrowRight class="w-4 h-4" />
					</button>
				{/if}
			</div>
		</div>
	</div>
</div>
