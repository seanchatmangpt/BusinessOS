<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { slide } from 'svelte/transition';
	import { CheckCircle2, AlertCircle, Clock, ChevronDown } from 'lucide-svelte';
	import Badge from '$lib/ui/badge/Badge.svelte';

	interface Control {
		id: string;
		status: 'pass' | 'fail' | 'pending';
		severity?: string;
		description: string;
		remediation?: string;
		framework?: string;
		lastChecked?: string;
	}

	interface Props {
		controls: Control[];
		expandedControls: Set<string>;
	}

	let { controls = [], expandedControls = new Set() }: Props = $props();
	const dispatch = createEventDispatcher<{ toggle: string }>();

	const getStatusIcon = (status: string) => {
		if (status === 'pass') return CheckCircle2;
		if (status === 'pending') return Clock;
		return AlertCircle;
	};

	const getStatusColor = (status: string): string => {
		if (status === 'pass') return 'text-green-600';
		if (status === 'pending') return 'text-yellow-600';
		return 'text-red-600';
	};

	const getBadgeColor = (status: string): string => {
		if (status === 'pass') return 'bg-green-100 text-green-800 border-green-300';
		if (status === 'pending') return 'bg-yellow-100 text-yellow-800 border-yellow-300';
		return 'bg-red-100 text-red-800 border-red-300';
	};

	const sortedControls = $derived.by(() => {
		const copy = [...controls];
		// Sort by status: fail, pending, pass
		return copy.sort((a, b) => {
			const statusOrder = { fail: 0, pending: 1, pass: 2 };
			return (statusOrder[a.status] || 3) - (statusOrder[b.status] || 3);
		});
	});

	function handleToggle(controlId: string) {
		dispatch('toggle', controlId);
	}
</script>

<div class="space-y-3">
	{#each sortedControls as control (control.id)}
		{@const Icon = getStatusIcon(control.status)}
		{@const isExpanded = expandedControls.has(control.id)}

		<div class="border border-gray-200 rounded-lg hover:border-gray-300 transition-colors">
			<button
				on:click={() => handleToggle(control.id)}
				class="w-full flex items-start gap-3 p-4 text-left hover:bg-gray-50 transition-colors"
			>
				<!-- Status Icon -->
				<div class={`flex-shrink-0 mt-0.5 ${getStatusColor(control.status)}`}>
					<Icon class="w-5 h-5" />
				</div>

				<!-- Content -->
				<div class="flex-1 min-w-0">
					<div class="flex items-center gap-2 mb-1">
						<span class="font-medium text-gray-900">{control.id}</span>
						<Badge class={getBadgeColor(control.status)}>
							{control.status?.toUpperCase() || 'PENDING'}
						</Badge>
						{#if control.severity}
							<Badge
								class={control.severity === 'critical'
									? 'bg-red-100 text-red-800'
									: control.severity === 'high'
										? 'bg-orange-100 text-orange-800'
										: 'bg-yellow-100 text-yellow-800'}
							>
								{control.severity.toUpperCase()}
							</Badge>
						{/if}
					</div>
					<p class="text-sm text-gray-600 line-clamp-2">{control.description}</p>
				</div>

				<!-- Chevron -->
				<div class={`flex-shrink-0 text-gray-400 transition-transform ${isExpanded ? 'rotate-180' : ''}`}>
					<ChevronDown class="w-5 h-5" />
				</div>
			</button>

			<!-- Expanded Content -->
			{#if isExpanded}
				<div class="border-t border-gray-200 bg-gray-50 px-4 py-3" transition:slide={{ duration: 200 }}>
					<div class="space-y-3">
						{#if control.remediation}
							<div>
								<p class="text-sm font-semibold text-gray-700 mb-2">Remediation Steps</p>
								<div class="bg-white border border-gray-200 rounded p-3 text-sm text-gray-600">
									{control.remediation}
								</div>
							</div>
						{/if}

						{#if control.lastChecked}
							<div class="flex justify-between text-xs text-gray-500">
								<span>Last Checked:</span>
								<span>{new Date(control.lastChecked).toLocaleDateString()}</span>
							</div>
						{/if}

						<div class="flex gap-2 pt-2">
							<button class="text-xs px-3 py-1.5 rounded bg-blue-100 text-blue-700 hover:bg-blue-200 transition-colors">
								View Details
							</button>
							{#if control.status !== 'pass'}
								<button class="text-xs px-3 py-1.5 rounded bg-orange-100 text-orange-700 hover:bg-orange-200 transition-colors">
									Remediate
								</button>
							{/if}
						</div>
					</div>
				</div>
			{/if}
		</div>
	{/each}

	{#if controls.length === 0}
		<div class="text-center py-8 text-gray-500">No controls found</div>
	{/if}
</div>
