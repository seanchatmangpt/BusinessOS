<script lang="ts">
	import type { AccessEvent } from '$lib/api/healthcare';
	import { healthcareAPI } from '$lib/api/healthcare';

	interface Props {
		events: AccessEvent[];
		sortBy?: 'timestamp' | 'action' | 'user';
	}

	let { events = [], sortBy = 'timestamp' }: Props = $props();

	// Derived: sorted events
	const sortedEvents = $derived.by(() => {
		const copy = [...events];

		switch (sortBy) {
			case 'action':
				return copy.sort((a, b) => a.action.localeCompare(b.action));
			case 'user':
				return copy.sort((a, b) => a.userName.localeCompare(b.userName));
			case 'timestamp':
			default:
				return copy.sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime());
		}
	});

	// Format timestamp for display
	function formatTime(timestamp: string): string {
		const date = new Date(timestamp);
		return date.toLocaleString();
	}

	// Get action badge color
	function getActionColor(action: string): string {
		switch (action) {
			case 'read':
				return 'bg-blue-100 text-blue-800';
			case 'write':
				return 'bg-green-100 text-green-800';
			case 'delete':
				return 'bg-red-100 text-red-800';
			case 'export':
				return 'bg-purple-100 text-purple-800';
			case 'access_denied':
				return 'bg-gray-100 text-gray-800';
			default:
				return 'bg-gray-100 text-gray-800';
		}
	}

	// Get success indicator icon
	function getSuccessIcon(success: boolean): string {
		return success ? '✓' : '✗';
	}
</script>

<div class="space-y-4">
	<div class="flex items-center justify-between">
		<h3 class="text-lg font-semibold text-gray-900">Access Log</h3>
		<span class="text-sm text-gray-600">{sortedEvents.length} events</span>
	</div>

	<div class="overflow-x-auto">
		<table class="min-w-full divide-y divide-gray-200 border border-gray-200">
			<thead class="bg-gray-50">
				<tr>
					<th class="px-4 py-3 text-left text-sm font-semibold text-gray-700">User</th>
					<th class="px-4 py-3 text-left text-sm font-semibold text-gray-700">Action</th>
					<th class="px-4 py-3 text-left text-sm font-semibold text-gray-700">Resource</th>
					<th class="px-4 py-3 text-left text-sm font-semibold text-gray-700">Date/Time</th>
					<th class="px-4 py-3 text-left text-sm font-semibold text-gray-700">Status</th>
					<th class="px-4 py-3 text-left text-sm font-semibold text-gray-700">Reason</th>
				</tr>
			</thead>
			<tbody class="divide-y divide-gray-200 bg-white">
				{#each sortedEvents as event (event.id)}
					<tr class="hover:bg-gray-50">
						<td class="px-4 py-3 text-sm text-gray-900">
							<!-- Masked display: role shown, full name hidden -->
							<div class="flex items-center space-x-2">
								<span class="font-medium">{healthcareAPI.maskPII(event.userName, event.userRole)}</span>
							</div>
						</td>
						<td class="px-4 py-3 text-sm">
							<span class="inline-flex items-center rounded-full px-3 py-1 text-xs font-medium {getActionColor(event.action)}">
								{event.action}
							</span>
						</td>
						<td class="px-4 py-3 text-sm text-gray-600">
							{event.resourceType}
						</td>
						<td class="px-4 py-3 text-sm text-gray-600">
							{formatTime(event.timestamp)}
						</td>
						<td class="px-4 py-3 text-center text-sm">
							<span class={event.success ? 'text-green-600 font-bold' : 'text-red-600 font-bold'}>
								{getSuccessIcon(event.success)}
							</span>
						</td>
						<td class="px-4 py-3 text-sm text-gray-600">
							{event.reason || '—'}
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>

	{#if sortedEvents.length === 0}
		<div class="rounded-lg border border-gray-200 bg-gray-50 px-4 py-8 text-center">
			<p class="text-gray-600">No access events recorded</p>
		</div>
	{/if}
</div>

<style>
	/* Ensures audit log is accessible and readable */
	:global(.access-log-table) {
		@apply w-full table-fixed;
	}
</style>
