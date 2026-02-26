<script lang="ts">
	import type { CustomModule } from '$lib/types/modules';
	import { categoryColors, categoryLabels } from '$lib/types/modules';
	import { Download, Star, Package } from 'lucide-svelte';

	interface Props {
		module: CustomModule;
		onClick?: () => void;
	}

	let { module, onClick }: Props = $props();
</script>

<button
	onclick={onClick}
	class="w-full text-left bg-white border border-gray-200 rounded-xl p-5 hover:shadow-lg hover:border-gray-300 transition-all duration-200 cursor-pointer group"
>
	<!-- Header with Icon and Category -->
	<div class="flex items-start justify-between mb-3">
		<div class="flex items-center gap-3">
			{#if module.icon}
				<div class="w-12 h-12 rounded-xl bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center text-white text-xl font-bold">
					{module.icon}
				</div>
			{:else}
				<div class="w-12 h-12 rounded-xl bg-gradient-to-br from-gray-400 to-gray-600 flex items-center justify-center text-white">
					<Package class="w-6 h-6" />
				</div>
			{/if}
		</div>
		<span class="text-xs px-2.5 py-1 rounded-full border {categoryColors[module.category]}">
			{categoryLabels[module.category]}
		</span>
	</div>

	<!-- Title and Description -->
	<h3 class="font-semibold text-base text-gray-900 mb-2 line-clamp-1 group-hover:text-blue-600 transition-colors">
		{module.name}
	</h3>
	<p class="text-sm text-gray-600 mb-4 line-clamp-2">
		{module.description}
	</p>

	<!-- Metadata Footer -->
	<div class="flex items-center justify-between pt-3 border-t border-gray-100">
		<!-- Left: Stats -->
		<div class="flex items-center gap-4 text-xs text-gray-500">
			<span class="flex items-center gap-1" title="Installs">
				<Download class="w-3.5 h-3.5" />
				{module.install_count}
			</span>
			<span class="flex items-center gap-1" title="Stars">
				<Star class="w-3.5 h-3.5" />
				{module.star_count}
			</span>
		</div>

		<!-- Right: Version -->
		<span class="text-xs text-gray-400">
			v{module.version}
		</span>
	</div>

	<!-- Author -->
	{#if module.creator_name}
		<div class="mt-3 pt-3 border-t border-gray-100">
			<p class="text-xs text-gray-500">
				by <span class="font-medium text-gray-700">{module.creator_name}</span>
			</p>
		</div>
	{/if}
</button>
