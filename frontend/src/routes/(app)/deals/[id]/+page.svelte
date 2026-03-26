<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { getDeal, verifyCompliance, type Deal } from '$lib/api/deals';

	let deal = $state<Deal | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let verifying = $state(false);
	let activeTab = $state<'overview' | 'kyc' | 'sox' | 'audit'>('overview');

	const dealId = $derived($page.params.id);

	onMount(async () => {
		try {
			const fetchedDeal = await getDeal(dealId);
			deal = fetchedDeal;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load deal';
		} finally {
			loading = false;
		}
	});

	async function handleVerifyCompliance() {
		if (!deal) return;

		verifying = true;
		try {
			const updated = await verifyCompliance(dealId);
			deal = updated;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Verification failed';
		} finally {
			verifying = false;
		}
	}

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
				day: 'numeric',
				hour: '2-digit',
				minute: '2-digit'
			});
		} catch {
			return dateStr;
		}
	};
</script>

<div class="deal-detail-page">
	<!-- Header -->
	<div class="deal-header">
		<div class="header-back">
			<button class="btn-back" onclick={() => goto('/deals')} title="Back to deals">
				<svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
				</svg>
			</button>
			<div>
				<h1 class="deal-title">{deal?.name || 'Loading...'}</h1>
				<p class="deal-subtitle">Deal ID: {dealId}</p>
			</div>
		</div>
		<div class="header-actions">
			<button class="btn-primary" onclick={handleVerifyCompliance} disabled={verifying || !deal}>
				{#if verifying}
					<svg class="icon spinner" viewBox="0 0 24 24" fill="none" stroke="currentColor">
						<circle cx="12" cy="12" r="10" stroke-width="2" />
						<path d="M12 2a10 10 0 0110 10" stroke-width="2" stroke-linecap="round" />
					</svg>
					Verifying...
				{:else}
					<svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor">
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
						/>
					</svg>
					Verify Compliance
				{/if}
			</button>
			<button class="btn-secondary" onclick={() => goto(`/deals/${dealId}/edit`)} disabled={!deal} title="Edit deal">
				<svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor">
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
					/>
				</svg>
				Edit
			</button>
		</div>
	</div>

	<!-- Error Banner -->
	{#if error}
		<div class="error-banner">
			<p class="error-text">{error}</p>
			<button class="error-dismiss" onclick={() => (error = null)}>×</button>
		</div>
	{/if}

	<!-- Loading State -->
	{#if loading}
		<div class="deal-loading">
			<svg class="spinner" viewBox="0 0 24 24" fill="none" stroke="currentColor">
				<circle cx="12" cy="12" r="10" stroke-width="2" />
				<path d="M12 2a10 10 0 0110 10" stroke-width="2" stroke-linecap="round" />
			</svg>
			<p>Loading deal...</p>
		</div>
	{:else if deal}
		<!-- Content -->
		<div class="deal-content">
			<!-- Quick Stats -->
			<div class="quick-stats">
				<div class="stat-card">
					<p class="stat-label">Amount</p>
					<p class="stat-value">{formatCurrency(deal.amount, deal.currency)}</p>
				</div>
				<div class="stat-card">
					<p class="stat-label">Status</p>
					<p class="stat-value">
						<span class="badge {getStatusBadgeClass(deal.status)}">
							{deal.status}
						</span>
					</p>
				</div>
				<div class="stat-card">
					<p class="stat-label">Compliance</p>
					<p class="stat-value">
						<span class="badge {getComplianceBadgeClass(deal.complianceStatus)}">
							{deal.complianceStatus}
						</span>
					</p>
				</div>
				<div class="stat-card">
					<p class="stat-label">RDF Triples</p>
					<p class="stat-value">{deal.rdfTripleCount}</p>
				</div>
			</div>

			<!-- Tab Navigation -->
			<div class="tab-nav">
				<button
					class="tab-button"
					class:active={activeTab === 'overview'}
					onclick={() => (activeTab = 'overview')}
				>
					Overview
				</button>
				<button
					class="tab-button"
					class:active={activeTab === 'kyc'}
					onclick={() => (activeTab = 'kyc')}
				>
					KYC/AML
				</button>
				<button
					class="tab-button"
					class:active={activeTab === 'sox'}
					onclick={() => (activeTab = 'sox')}
				>
					SOX Verification
				</button>
				<button
					class="tab-button"
					class:active={activeTab === 'audit'}
					onclick={() => (activeTab = 'audit')}
				>
					Audit Trail
				</button>
			</div>

			<!-- Tab Contents -->
			<div class="tab-content">
				{#if activeTab === 'overview'}
					<div class="overview-section">
						<h3 class="section-title">Deal Details</h3>
						<div class="details-grid">
							<div class="detail-item">
								<span class="detail-label">Deal Name</span>
								<span class="detail-value">{deal.name}</span>
							</div>
							<div class="detail-item">
								<span class="detail-label">Amount</span>
								<span class="detail-value">{formatCurrency(deal.amount, deal.currency)}</span>
							</div>
							<div class="detail-item">
								<span class="detail-label">Currency</span>
								<span class="detail-value">{deal.currency}</span>
							</div>
							<div class="detail-item">
								<span class="detail-label">Stage</span>
								<span class="detail-value">{deal.stage}</span>
							</div>
							<div class="detail-item">
								<span class="detail-label">Probability</span>
								<span class="detail-value">{deal.probability}%</span>
							</div>
							<div class="detail-item">
								<span class="detail-label">Expected Close Date</span>
								<span class="detail-value">{deal.expectedCloseDate || 'N/A'}</span>
							</div>
							<div class="detail-item">
								<span class="detail-label">Buyer ID</span>
								<span class="detail-value">{deal.buyerId}</span>
							</div>
							<div class="detail-item">
								<span class="detail-label">Seller ID</span>
								<span class="detail-value">{deal.sellerId}</span>
							</div>
							<div class="detail-item">
								<span class="detail-label">Created</span>
								<span class="detail-value">{formatDate(deal.createdAt)}</span>
							</div>
							<div class="detail-item">
								<span class="detail-label">Last Updated</span>
								<span class="detail-value">{formatDate(deal.updatedAt)}</span>
							</div>
						</div>
					</div>
				{:else if activeTab === 'kyc'}
					<div class="kyc-section">
						<h3 class="section-title">KYC/AML Verification</h3>
						<div class="kyc-status">
							<div class="status-row">
								<span class="status-label">KYC Status</span>
								<span class="status-badge {deal.kycVerified ? 'badge-success' : 'badge-warning'}">
									{deal.kycVerified ? 'Verified' : 'Pending'}
								</span>
							</div>
							<div class="status-row">
								<span class="status-label">AML Screening</span>
								<span class="status-badge badge-{deal.amlScreening.toLowerCase()}">
									{deal.amlScreening}
								</span>
							</div>
						</div>
					</div>
				{:else if activeTab === 'sox'}
					<div class="sox-section">
						<h3 class="section-title">SOX Verification</h3>
						<div class="verification-status">
							<p class="verification-text">
								SOX compliance status: <strong>{deal.complianceStatus.toUpperCase()}</strong>
							</p>
							<p class="verification-text">
								RDF Triple Count: <strong>{deal.rdfTripleCount}</strong>
							</p>
						</div>
					</div>
				{:else if activeTab === 'audit'}
					<div class="audit-section">
						<h3 class="section-title">Audit Trail</h3>
						<div class="audit-entries">
							<div class="audit-entry">
								<span class="audit-label">Created:</span>
								<span class="audit-time">{formatDate(deal.createdAt)}</span>
							</div>
							<div class="audit-entry">
								<span class="audit-label">Last Modified:</span>
								<span class="audit-time">{formatDate(deal.updatedAt)}</span>
							</div>
						</div>
					</div>
				{/if}
			</div>
		</div>
	{:else}
		<div class="deal-not-found">
			<p>Deal not found</p>
		</div>
	{/if}
</div>

<style>
	.deal-detail-page {
		display: flex;
		flex-direction: column;
		height: 100%;
		background: var(--dbg, #fff);
	}

	.deal-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 20px 24px;
		border-bottom: 1px solid var(--dbd, #e0e0e0);
	}

	.header-back {
		display: flex;
		align-items: center;
		gap: 12px;
		flex: 1;
	}

	.btn-back {
		background: none;
		border: none;
		color: var(--dt, #111);
		cursor: pointer;
		padding: 4px;
		transition: opacity 0.15s ease;
	}

	.btn-back:hover {
		opacity: 0.6;
	}

	.icon {
		width: 16px;
		height: 16px;
	}

	.spinner {
		animation: spin 1s linear infinite;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}

	.deal-title {
		font-size: 24px;
		font-weight: 700;
		color: var(--dt, #111);
		margin: 0;
	}

	.deal-subtitle {
		font-size: 12px;
		color: var(--dt3, #888);
		margin: 4px 0 0 0;
	}

	.header-actions {
		display: flex;
		gap: 8px;
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
		white-space: nowrap;
	}

	.btn-primary {
		background: #6366f1;
		color: white;
	}

	.btn-primary:hover:not(:disabled) {
		background: #4f46e5;
	}

	.btn-primary:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.btn-secondary {
		background: var(--dbg-secondary, #f5f5f5);
		color: var(--dt, #111);
		border: 1px solid var(--dbd, #e0e0e0);
	}

	.btn-secondary:hover:not(:disabled) {
		background: var(--dbg, #fff);
	}

	.btn-secondary:disabled {
		opacity: 0.6;
		cursor: not-allowed;
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
		cursor: pointer;
		font-size: 18px;
		padding: 0;
	}

	.deal-loading,
	.deal-not-found {
		flex: 1;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 12px;
		color: var(--dt3, #888);
	}

	.deal-content {
		flex: 1;
		overflow-y: auto;
		padding: 24px;
	}

	.quick-stats {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 12px;
		margin-bottom: 32px;
	}

	.stat-card {
		padding: 16px;
		background: var(--dbg-secondary, #f5f5f5);
		border-radius: 8px;
		border: 1px solid var(--dbd, #e0e0e0);
	}

	.stat-label {
		font-size: 12px;
		color: var(--dt3, #888);
		margin: 0;
		font-weight: 500;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.stat-value {
		font-size: 18px;
		font-weight: 700;
		color: var(--dt, #111);
		margin: 8px 0 0 0;
	}

	.badge {
		display: inline-block;
		padding: 4px 8px;
		border-radius: 4px;
		font-size: 11px;
		font-weight: 600;
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

	.tab-nav {
		display: flex;
		gap: 0;
		border-bottom: 1px solid var(--dbd, #e0e0e0);
		margin-bottom: 24px;
	}

	.tab-button {
		background: none;
		border: none;
		padding: 12px 16px;
		border-bottom: 2px solid transparent;
		cursor: pointer;
		font-size: 13px;
		font-weight: 600;
		color: var(--dt3, #888);
		transition: all 0.15s ease;
	}

	.tab-button:hover {
		color: var(--dt, #111);
	}

	.tab-button.active {
		color: #6366f1;
		border-bottom-color: #6366f1;
	}

	.section-title {
		font-size: 16px;
		font-weight: 700;
		color: var(--dt, #111);
		margin: 0 0 16px 0;
	}

	.details-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 16px;
	}

	.detail-item {
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	.detail-label {
		font-size: 11px;
		color: var(--dt3, #888);
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.detail-value {
		font-size: 14px;
		color: var(--dt, #111);
		font-weight: 500;
	}

	.kyc-status,
	.verification-status,
	.audit-entries {
		background: var(--dbg-secondary, #f5f5f5);
		padding: 16px;
		border-radius: 8px;
	}

	.status-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 12px 0;
		border-bottom: 1px solid var(--dbd, #e0e0e0);
	}

	.status-row:last-child {
		border-bottom: none;
	}

	.status-label {
		font-size: 13px;
		font-weight: 600;
		color: var(--dt, #111);
	}

	.status-badge {
		padding: 4px 8px;
		border-radius: 4px;
		font-size: 11px;
		font-weight: 600;
	}

	.verification-text {
		font-size: 13px;
		color: var(--dt, #111);
		margin: 0 0 12px 0;
	}

	.verification-text strong {
		font-weight: 700;
	}

	.audit-entries {
		display: flex;
		flex-direction: column;
		gap: 0;
	}

	.audit-entry {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 12px 0;
		border-bottom: 1px solid var(--dbd, #e0e0e0);
	}

	.audit-entry:last-child {
		border-bottom: none;
	}

	.audit-label {
		font-size: 13px;
		font-weight: 600;
		color: var(--dt, #111);
	}

	.audit-time {
		font-size: 13px;
		color: var(--dt3, #888);
	}
</style>
