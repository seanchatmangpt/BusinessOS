<script lang="ts">
	import { Plus, X, GripVertical } from 'lucide-svelte';
	import type { ModuleAction, ModuleActionType } from '$lib/types/modules';

	interface Props {
		actions: ModuleAction[];
		onActionsChange: (actions: ModuleAction[]) => void;
	}

	let { actions, onActionsChange }: Props = $props();

	function addAction() {
		const newAction: ModuleAction = {
			name: '',
			description: '',
			type: 'function',
			parameters: {},
			returns: {}
		};
		onActionsChange([...actions, newAction]);
	}

	function removeAction(index: number) {
		onActionsChange(actions.filter((_, i) => i !== index));
	}

	function updateAction(index: number, field: keyof ModuleAction, value: unknown) {
		const updated = [...actions];
		updated[index] = { ...updated[index], [field]: value };
		onActionsChange(updated);
	}

	const actionTypes: ModuleActionType[] = ['function', 'api', 'workflow'];
</script>

<div class="space-y-4">
	<!-- Actions List -->
	{#each actions as action, index}
		<div class="bg-gray-50 border border-gray-200 rounded-xl p-4">
			<!-- Header with drag handle and delete -->
			<div class="flex items-center gap-2 mb-3">
				<button
					type="button"
					class="cursor-move p-1 text-gray-400 hover:text-gray-600"
					title="Drag to reorder"
				>
					<GripVertical class="w-5 h-5" />
				</button>
				<span class="text-sm font-medium text-gray-700">Action {index + 1}</span>
				<button
					type="button"
					onclick={() => removeAction(index)}
					class="ml-auto p-1 text-red-500 hover:text-red-700 rounded-lg hover:bg-red-50 transition-colors"
					title="Remove action"
				>
					<X class="w-5 h-5" />
				</button>
			</div>

			<!-- Action Fields -->
			<div class="grid grid-cols-2 gap-3">
				<!-- Name -->
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">
						Name <span class="text-red-500">*</span>
					</label>
					<input
						type="text"
						value={action.name}
						oninput={(e) => updateAction(index, 'name', e.currentTarget.value)}
						placeholder="e.g., sendEmail"
						class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
						required
					/>
				</div>

				<!-- Type -->
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">
						Type <span class="text-red-500">*</span>
					</label>
					<select
						value={action.type}
						onchange={(e) => updateAction(index, 'type', e.currentTarget.value)}
						class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
					>
						{#each actionTypes as type}
							<option value={type}>{type.charAt(0).toUpperCase() + type.slice(1)}</option>
						{/each}
					</select>
				</div>

				<!-- Description -->
				<div class="col-span-2">
					<label class="block text-sm font-medium text-gray-700 mb-1">
						Description <span class="text-red-500">*</span>
					</label>
					<textarea
						value={action.description}
						oninput={(e) => updateAction(index, 'description', e.currentTarget.value)}
						placeholder="Describe what this action does..."
						rows="2"
						class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
						required
					></textarea>
				</div>

				<!-- Parameters (JSON) -->
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">
						Parameters (JSON)
					</label>
					<textarea
						value={JSON.stringify(action.parameters, null, 2)}
						oninput={(e) => {
							try {
								const parsed = JSON.parse(e.currentTarget.value);
								updateAction(index, 'parameters', parsed);
							} catch {
								// Invalid JSON, ignore
							}
						}}
						placeholder={'{"param1": "type", "param2": "type"}'}
						rows="3"
						class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 font-mono text-sm"
					></textarea>
				</div>

				<!-- Returns (JSON) -->
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">
						Returns (JSON)
					</label>
					<textarea
						value={JSON.stringify(action.returns, null, 2)}
						oninput={(e) => {
							try {
								const parsed = JSON.parse(e.currentTarget.value);
								updateAction(index, 'returns', parsed);
							} catch {
								// Invalid JSON, ignore
							}
						}}
						placeholder={'{"result": "type"}'}
						rows="3"
						class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 font-mono text-sm"
					></textarea>
				</div>
			</div>
		</div>
	{/each}

	<!-- Add Action Button -->
	<button
		type="button"
		onclick={addAction}
		class="w-full flex items-center justify-center gap-2 px-4 py-3 border-2 border-dashed border-gray-300 rounded-xl text-gray-600 hover:border-blue-500 hover:text-blue-600 transition-colors"
	>
		<Plus class="w-5 h-5" />
		<span class="font-medium">Add Action</span>
	</button>
</div>
