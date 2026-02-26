<script lang="ts">
	import type { ModuleCategory } from '$lib/types/modules';
	import { categoryLabels } from '$lib/types/modules';

	interface Props {
		name: string;
		description: string;
		category: ModuleCategory;
		icon: string;
		onNameChange: (value: string) => void;
		onDescriptionChange: (value: string) => void;
		onCategoryChange: (value: ModuleCategory) => void;
		onIconChange: (value: string) => void;
	}

	let { name, description, category, icon, onNameChange, onDescriptionChange, onCategoryChange, onIconChange }: Props = $props();

	const categories: ModuleCategory[] = [
		'productivity',
		'communication',
		'finance',
		'analytics',
		'automation',
		'integration',
		'utilities',
		'custom'
	];
</script>

<div class="space-y-4">
	<!-- Name -->
	<div>
		<label class="block text-sm font-medium text-gray-700 mb-2">
			Module Name <span class="text-red-500">*</span>
		</label>
		<input
			type="text"
			value={name}
			oninput={(e) => onNameChange(e.currentTarget.value)}
			placeholder="e.g., Email Automation"
			class="w-full px-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
			required
		/>
	</div>

	<!-- Description -->
	<div>
		<label class="block text-sm font-medium text-gray-700 mb-2">
			Description <span class="text-red-500">*</span>
		</label>
		<textarea
			value={description}
			oninput={(e) => onDescriptionChange(e.currentTarget.value)}
			placeholder="Describe what your module does..."
			rows="4"
			class="w-full px-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
			required
		></textarea>
	</div>

	<!-- Category -->
	<div>
		<label class="block text-sm font-medium text-gray-700 mb-2">
			Category <span class="text-red-500">*</span>
		</label>
		<select
			value={category}
			onchange={(e) => onCategoryChange(e.currentTarget.value as ModuleCategory)}
			class="w-full px-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
		>
			{#each categories as cat}
				<option value={cat}>{categoryLabels[cat]}</option>
			{/each}
		</select>
	</div>

	<!-- Icon (Emoji picker) -->
	<div>
		<label class="block text-sm font-medium text-gray-700 mb-2">
			Icon (Emoji)
		</label>
		<input
			type="text"
			value={icon}
			oninput={(e) => onIconChange(e.currentTarget.value)}
			placeholder="📦"
			maxlength="2"
			class="w-20 px-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-2xl text-center"
		/>
		<p class="mt-1 text-xs text-gray-500">Enter an emoji to represent your module</p>
	</div>
</div>
