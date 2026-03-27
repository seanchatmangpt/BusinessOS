<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import DealTable from '$lib/components/DealTable.svelte';
	import { loadDeals, dealsStore, type Deal, type DealStatus, type DealDomain } from '$lib/api/deals';

	let deals = $derived($dealsStore.deals);
	let loading = $derived($dealsStore.loading);
	let error = $derived($dealsStore.error);

	let statusFilter: DealStatus | undefined = $state();
	let domainFilter: DealDomain | undefined = $state();
	let searchQuery = $state('');
	let selectedRows = $state<Set<string>>(new Set());

	const filteredDeals = $derived.by(() => {
		return deals.filter((deal) => {
			if (searchQuery) {
				const query = searchQuery.toLowerCase();
				const matches =
					deal.name.toLowerCase().includes(query) ||
					deal.id.toLowerCase().includes(query) ||
					deal.buyerId.toLowerCase().includes(query);
				if (!matches) return false;
			}

			if (statusFilter && deal.status !== statusFilter) {
				return false;
			}

			if (domainFilter && deal.domain !== domainFilter) {
				return false;
			}

			return true;
		});
	});

	onMount(async () => {
		try {
			await loadDeals(20, 0, statusFilter, domainFilter);
		} catch (err) {
			console.error('Failed to load deals:', err);
		}
	});

	async function handleCreateClick() {
		goto('/deals/create');
	}

	async function handleRowClick(dealId: string) {
		goto(`/deals/${dealId}`);
	}

	async function handleRefresh() {
		try {
			await loadDeals(20, 0, statusFilter, domainFilter);
		} catch (err) {
			console.error('Failed to refresh deals:', err);
		}
	}

	function handleStatusFilterChange(status: DealStatus | null) {
		statusFilter = status ?? undefined;
	}

	function handleDomainFilterChange(domain: DealDomain | null) {
		domainFilter = domain ?? undefined;
	}
</script>

<div class="deals-page">
	<!-- Header -->
	<div class="deals-header">
		<div>
			<h1 class="deals-title">FIBO Deals</h1>
			<p class="deals-subtitle">Manage financial deals and track compliance</p>
		</div>
		<div class="deals-actions">
			<button class="btn-secondary" onclick={handleRefresh} title="Refresh deals">
				<svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor">
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
					/>
				</svg>
			</button>
			<button class="btn-primary" onclick={handleCreateClick}>
				<svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
				</svg>
				Create Deal
			</button>
		</div>
	</div>

	<!-- Error State -->
	{#if error}
		<div class="error-banner">
			<p class="error-text">{error}</p>
			<button class="error-dismiss" onclick={() => (error = null)}>Dismiss</button>
		</div>
	{/if}

	<!-- Filters -->
	<div class="deals-filters">
		<div class="filter-group">
			<input
				type="text"
				class="search-input"
				placeholder="Search by deal name, ID, or buyer..."
				value={searchQuery}
				onchange={(e) => (searchQuery = e.currentTarget.value)}
			/>
		</div>

		<div class="filter-group">
			<select class="filter-select" onchange={(e) => handleStatusFilterChange(e.currentTarget.value as DealStatus | null)}>
				<option value="">All Statuses</option>
				<option value="draft">Draft</option>
				<option value="pending">Pending</option>
				<option value="active">Active</option>
				<option value="closed">Closed</option>
			</select>
		</div>

		<div class="filter-group">
			<select class="filter-select" onchange={(e) => handleDomainFilterChange(e.currentTarget.value as DealDomain | null)}>
				<option value="">All Domains</option>
				<option value="Finance">Finance</option>
				<option value="Other">Other</option>
			</select>
		</div>
	</div>

	<!-- Deals Table -->
	<DealTable
		{filteredDeals}
		isLoading={loading}
		selectedRows={selectedRows}
		onRowClick={handleRowClick}
	/>

	<!-- Empty State -->
	{#if !loading && filteredDeals.length === 0}
		<div class="deals-empty">
			<svg class="empty-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor">
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="1.5"
					d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4"
				/>
			</svg>
			<p class="empty-title">No deals found</p>
			<p class="empty-text">Create your first deal to get started</p>
			<button class="btn-primary" onclick={handleCreateClick}>
				<svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
				</svg>
				Create Deal
			</button>
		</div>
	{/if}
</div>

<style>
	.deals-page {
		display: flex;
		flex-direction: column;
		height: 100%;
		background: var(--dbg, #fff);
	}

	.deals-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 20px 24px;
		border-bottom: 1px solid var(--dbd, #e0e0e0);
	}

	.deals-title {
		font-size: 24px;
		font-weight: 700;
		color: var(--dt, #111);
		margin: 0;
		letter-spacing: -0.01em;
	}

	.deals-subtitle {
		font-size: 13px;
		color: var(--dt3, #888);
		margin: 4px 0 0 0;
	}

	.deals-actions {
		display: flex;
		gap: 8px;
		align-items: center;
	}

	.icon {
		width: 16px;
		height: 16px;
	}

	.btn-primary,
	.btn-secondary {
		padding: 8px 16px;
		border: none;
		border-radius: 6px;
		font-size: 13px;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.15s ease;
		display: flex;
		align-items: center;
		gap: 6px;
	}

	.btn-primary {
		background: #6366f1;
		color: white;
	}

	.btn-primary:hover {
		background: #4f46e5;
		transform: translateY(-1px);
	}

	.btn-secondary {
		background: var(--dbg-secondary, #f5f5f5);
		color: var(--dt, #111);
		border: 1px solid var(--dbd, #e0e0e0);
	}

	.btn-secondary:hover {
		background: var(--dbg, #fff);
	}

	.error-banner {
		margin: 16px 24px 0;
		padding: 12px 16px;
		background: rgba(239, 68, 68, 0.06);
		border: 1px solid rgba(239, 68, 68, 0.2);
		border-radius: 8px;
		display: flex;
		align-items: center;
		justify-content: space-between;
	}

	.error-text {
		font-size: 13px;
		color: #ef4444;
		font-weight: 500;
		margin: 0;
	}

	.error-dismiss {
		background: none;
		border: none;
		color: #ef4444;
		font-size: 12px;
		cursor: pointer;
		text-decoration: underline;
		padding: 0;
	}

	.deals-filters {
		display: flex;
		gap: 12px;
		padding: 16px 24px;
		border-bottom: 1px solid var(--dbd, #e0e0e0);
		flex-wrap: wrap;
		align-items: center;
	}

	.filter-group {
		flex: 1;
		min-width: 200px;
	}

	.search-input,
	.filter-select {
		width: 100%;
		padding: 8px 12px;
		border: 1px solid var(--dbd, #e0e0e0);
		border-radius: 6px;
		font-size: 13px;
		color: var(--dt, #111);
		background: var(--dbg, #fff);
		transition: border-color 0.15s ease;
	}

	.search-input:focus,
	.filter-select:focus {
		outline: none;
		border-color: #6366f1;
	}

	.deals-empty {
		flex: 1;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 48px 24px;
		gap: 12px;
		color: var(--dt3, #888);
	}

	.empty-icon {
		width: 48px;
		height: 48px;
		color: var(--dt3, #888);
		margin-bottom: 8px;
	}

	.empty-title {
		font-size: 16px;
		font-weight: 600;
		color: var(--dt, #111);
		margin: 0;
	}

	.empty-text {
		font-size: 13px;
		color: var(--dt3, #888);
		margin: 0 0 16px 0;
	}
</style>
