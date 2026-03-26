<script lang="ts">
	import type { Deal } from '$lib/api/deals';

	interface Props {
		deals: Deal[];
		isLoading?: boolean;
		selectedRows?: Set<string>;
		onRowClick?: (id: string) => void;
		onStatusChange?: (id: string, status: string) => void;
		sortBy?: 'name' | 'amount' | 'status' | 'created';
		sortDirection?: 'asc' | 'desc';
	}

	let {
		deals = [],
		isLoading = false,
		selectedRows = new Set(),
		onRowClick = undefined,
		onStatusChange = undefined,
		sortBy = 'created',
		sortDirection = 'desc'
	}: Props = $props();

	const getComplianceBadgeClass = (status: string): string => {
		return {
			pass: 'badge-success',
			fail: 'badge-error',
			pending: 'badge-warning'
		}[status] || 'badge-default';
	};

	const getStatusBadgeClass = (status: string): string => {
		return {
			draft: 'badge-default',
			pending: 'badge-warning',
			active: 'badge-primary',
			closed: 'badge-success'
		}[status] || 'badge-default';
	};

	const formatCurrency = (amount: number, currency: string = 'USD'): string => {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency
		}).format(amount);
	};

	const formatDate = (dateStr: string): string => {
		try {
			return new Date(dateStr).toLocaleDateString('en-US', {
				year: 'numeric',
				month: 'short',
				day: 'numeric'
			});
		} catch {
			return dateStr;
		}
	};

	const sortedDeals = $derived.by(() => {
		const sorted = [...deals];

		if (sortBy === 'name') {
			sorted.sort((a, b) => a.name.localeCompare(b.name));
		} else if (sortBy === 'amount') {
			sorted.sort((a, b) => a.amount - b.amount);
		} else if (sortBy === 'status') {
			sorted.sort((a, b) => a.status.localeCompare(b.status));
		} else {
			sorted.sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime());
		}

		return sortDirection === 'desc' ? sorted.reverse() : sorted;
	});
</script>

<div class="deal-table-container">
	{#if isLoading}
		<div class="table-loading">
			<svg class="spinner" viewBox="0 0 24 24" fill="none" stroke="currentColor">
				<circle cx="12" cy="12" r="10" stroke-width="2" />
				<path d="M12 2a10 10 0 0110 10" stroke-width="2" stroke-linecap="round" />
			</svg>
			<p>Loading deals...</p>
		</div>
	{:else if sortedDeals.length === 0}
		<div class="table-empty">
			<p>No deals found</p>
		</div>
	{:else}
		<table class="deal-table">
			<thead>
				<tr>
					<th class="col-select">
						<input
							type="checkbox"
							class="checkbox"
							checked={selectedRows.size === sortedDeals.length && sortedDeals.length > 0}
							onchange={(e) => {
								if (e.currentTarget.checked) {
									selectedRows = new Set(sortedDeals.map((d) => d.id));
								} else {
									selectedRows.clear();
									selectedRows = selectedRows;
								}
							}}
						/>
					</th>
					<th class="col-id">Deal ID</th>
					<th class="col-name">
						<button
							class="sort-header"
							onclick={() => {
								if (sortBy === 'name') {
									sortDirection = sortDirection === 'asc' ? 'desc' : 'asc';
								} else {
									sortBy = 'name';
									sortDirection = 'asc';
								}
							}}
						>
							Name {sortBy === 'name' ? (sortDirection === 'asc' ? '↑' : '↓') : ''}
						</button>
					</th>
					<th class="col-amount">
						<button
							class="sort-header"
							onclick={() => {
								if (sortBy === 'amount') {
									sortDirection = sortDirection === 'asc' ? 'desc' : 'asc';
								} else {
									sortBy = 'amount';
									sortDirection = 'asc';
								}
							}}
						>
							Amount {sortBy === 'amount' ? (sortDirection === 'asc' ? '↑' : '↓') : ''}
						</button>
					</th>
					<th class="col-status">
						<button
							class="sort-header"
							onclick={() => {
								if (sortBy === 'status') {
									sortDirection = sortDirection === 'asc' ? 'desc' : 'asc';
								} else {
									sortBy = 'status';
									sortDirection = 'asc';
								}
							}}
						>
							Status {sortBy === 'status' ? (sortDirection === 'asc' ? '↑' : '↓') : ''}
						</button>
					</th>
					<th class="col-compliance">Compliance</th>
					<th class="col-kyc">KYC</th>
					<th class="col-actions">Actions</th>
				</tr>
			</thead>
			<tbody>
				{#each sortedDeals as deal (deal.id)}
					<tr
						class="deal-row"
						class:selected={selectedRows.has(deal.id)}
						onclick={() => onRowClick?.(deal.id)}
					>
						<td class="col-select">
							<input
								type="checkbox"
								class="checkbox"
								checked={selectedRows.has(deal.id)}
								onchange={(e) => {
									if (e.currentTarget.checked) {
										selectedRows.add(deal.id);
									} else {
										selectedRows.delete(deal.id);
									}
									selectedRows = selectedRows;
								}}
								onclick={(e) => e.stopPropagation()}
							/>
						</td>
						<td class="col-id">{deal.id.slice(0, 8)}</td>
						<td class="col-name">{deal.name}</td>
						<td class="col-amount">{formatCurrency(deal.amount, deal.currency)}</td>
						<td class="col-status">
							<span class="badge {getStatusBadgeClass(deal.status)}">
								{deal.status}
							</span>
						</td>
						<td class="col-compliance">
							<span class="badge {getComplianceBadgeClass(deal.complianceStatus)}">
								{deal.complianceStatus}
							</span>
						</td>
						<td class="col-kyc">
							{#if deal.kycVerified}
								<span class="badge badge-success">Verified</span>
							{:else}
								<span class="badge badge-warning">Pending</span>
							{/if}
						</td>
						<td class="col-actions">
							<button
								class="btn-action btn-view"
								title="View deal"
								onclick={(e) => {
									e.stopPropagation();
									onRowClick?.(deal.id);
								}}
							>
								View
							</button>
							<button
								class="btn-action btn-edit"
								title="Edit deal"
								onclick={(e) => e.stopPropagation()}
							>
								Edit
							</button>
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	{/if}
</div>

<style>
	.deal-table-container {
		width: 100%;
		overflow-x: auto;
		background: var(--dbg, #fff);
		border-radius: 8px;
		border: 1px solid var(--dbd, #e0e0e0);
	}

	.table-loading,
	.table-empty {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 48px 24px;
		color: var(--dt3, #888);
		gap: 12px;
	}

	.spinner {
		width: 24px;
		height: 24px;
		animation: spin 1s linear infinite;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}

	.deal-table {
		width: 100%;
		border-collapse: collapse;
		font-size: 13px;
	}

	thead {
		background: var(--dbg-secondary, #f5f5f5);
		border-bottom: 2px solid var(--dbd, #e0e0e0);
	}

	th {
		padding: 12px 16px;
		text-align: left;
		font-weight: 600;
		color: var(--dt, #111);
		user-select: none;
	}

	.sort-header {
		background: none;
		border: none;
		color: var(--dt, #111);
		font-weight: 600;
		cursor: pointer;
		font-size: 13px;
		padding: 0;
		white-space: nowrap;
	}

	.sort-header:hover {
		opacity: 0.7;
	}

	tbody tr {
		border-bottom: 1px solid var(--dbd, #e0e0e0);
		transition: background-color 0.15s ease;
	}

	tbody tr:hover {
		background-color: var(--dbg-secondary, #f9f9f9);
	}

	.deal-row.selected {
		background-color: rgba(99, 102, 241, 0.1);
	}

	td {
		padding: 12px 16px;
		color: var(--dt, #111);
	}

	.col-select,
	.col-id {
		width: 60px;
	}

	.col-name {
		min-width: 200px;
		font-weight: 500;
		color: var(--dt, #111);
	}

	.col-amount {
		width: 120px;
		text-align: right;
		font-weight: 500;
	}

	.col-status,
	.col-compliance,
	.col-kyc {
		width: 100px;
		text-align: center;
	}

	.col-actions {
		width: 140px;
		text-align: center;
	}

	.badge {
		display: inline-block;
		padding: 4px 8px;
		border-radius: 4px;
		font-size: 11px;
		font-weight: 600;
		white-space: nowrap;
	}

	.badge-success {
		background: rgba(34, 197, 94, 0.1);
		color: #22c55e;
	}

	.badge-error {
		background: rgba(239, 68, 68, 0.1);
		color: #ef4444;
	}

	.badge-warning {
		background: rgba(245, 158, 11, 0.1);
		color: #f59e0b;
	}

	.badge-primary {
		background: rgba(99, 102, 241, 0.1);
		color: #6366f1;
	}

	.badge-default {
		background: rgba(107, 114, 128, 0.1);
		color: #6b7280;
	}

	.checkbox {
		width: 16px;
		height: 16px;
		cursor: pointer;
		accent-color: #6366f1;
	}

	.btn-action {
		padding: 4px 8px;
		margin: 0 2px;
		font-size: 11px;
		border: 1px solid var(--dbd, #e0e0e0);
		border-radius: 4px;
		cursor: pointer;
		transition: all 0.15s ease;
		background: white;
		color: var(--dt, #111);
	}

	.btn-view {
		color: #6366f1;
		border-color: #6366f1;
	}

	.btn-view:hover {
		background: rgba(99, 102, 241, 0.1);
	}

	.btn-edit {
		color: #f59e0b;
		border-color: #f59e0b;
	}

	.btn-edit:hover {
		background: rgba(245, 158, 11, 0.1);
	}
</style>
