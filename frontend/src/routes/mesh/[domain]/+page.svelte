<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { meshClient, type Domain, type Dataset } from '$lib/api/mesh';

	let domain: Domain | null = null;
	let datasets: Dataset[] = [];
	let isLoading = false;
	let error: string | null = null;

	$: domainId = $page.params.domain;

	onMount(async () => {
		if (domainId) {
			await loadDomain();
		}
	});

	async function loadDomain() {
		isLoading = true;
		error = null;

		try {
			[domain, datasets] = await Promise.all([
				meshClient.getDomain(domainId),
				meshClient.getDatasets(domainId)
			]);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load domain';
		} finally {
			isLoading = false;
		}
	}

	function navigateToMesh() {
		window.location.href = '/mesh';
	}
</script>

<div class="domain-detail">
	<header class="page-header">
		<button class="back-button" on:click={navigateToMesh}>← Back to Mesh</button>
		<h1>{domain?.name || 'Loading...'}</h1>
	</header>

	{#if error}
		<div class="error-banner">
			<span>⚠️ {error}</span>
		</div>
	{/if}

	<div class="domain-content">
		{#if isLoading && !domain}
			<div class="loading-state">
				<p>Loading domain details...</p>
			</div>
		{:else if domain}
			<!-- Domain Overview -->
			<section class="section">
				<h2>Overview</h2>
				<div class="info-grid">
					<div class="info-card">
						<div class="info-label">Owner</div>
						<div class="info-value">{domain.owner}</div>
					</div>
					<div class="info-card">
						<div class="info-label">Governance Model</div>
						<div class="info-value">{domain.governance_model}</div>
					</div>
					<div class="info-card">
						<div class="info-label">Service Level Agreement</div>
						<div class="info-value">{domain.sla}</div>
					</div>
					<div class="info-card">
						<div class="info-label">Created</div>
						<div class="info-value">
							{new Date(domain.created_at).toLocaleDateString()}
						</div>
					</div>
					<div class="info-card">
						<div class="info-label">Last Updated</div>
						<div class="info-value">
							{new Date(domain.updated_at).toLocaleDateString()}
						</div>
					</div>
					<div class="info-card">
						<div class="info-label">Datasets</div>
						<div class="info-value">{datasets.length}</div>
					</div>
				</div>
			</section>

			<!-- Datasets -->
			<section class="section">
				<div class="section-header">
					<h2>Datasets in {domain.name}</h2>
					<button class="primary-button">+ New Contract</button>
				</div>

				{#if datasets.length === 0}
					<div class="empty-state">
						<p>No datasets in this domain yet</p>
						<button class="secondary-button">Create Dataset</button>
					</div>
				{:else}
					<div class="datasets-table">
						<table>
							<thead>
								<tr>
									<th>Name</th>
									<th>Owner</th>
									<th>Quality Score</th>
									<th>Last Modified</th>
									<th>Actions</th>
								</tr>
							</thead>
							<tbody>
								{#each datasets as dataset (dataset.id)}
									<tr>
										<td class="name-cell">
											<a href={`/mesh?dataset=${dataset.id}`} class="dataset-link">
												{dataset.name}
											</a>
										</td>
										<td>{dataset.owner}</td>
										<td>
											<span
												class="quality-badge"
												class:good={dataset.quality_score >= 80}
												class:fair={dataset.quality_score >= 60}
												class:poor={dataset.quality_score < 60}
											>
												{Math.round(dataset.quality_score)}
											</span>
										</td>
										<td>
											{new Date(dataset.last_modified).toLocaleDateString()}
										</td>
										<td class="actions-cell">
											<button class="icon-button" title="View Details">📊</button>
											<button class="icon-button" title="View Lineage">🔗</button>
											<button class="icon-button" title="Edit">✏️</button>
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				{/if}
			</section>
		{/if}
	</div>
</div>

<style>
	.domain-detail {
		width: 100%;
		height: 100%;
		display: flex;
		flex-direction: column;
		background: #f9fafb;
	}

	.page-header {
		padding: 24px;
		background: white;
		border-bottom: 1px solid #e5e7eb;
	}

	.back-button {
		padding: 8px 12px;
		background: white;
		border: 1px solid #e5e7eb;
		border-radius: 6px;
		cursor: pointer;
		font-size: 14px;
		color: #6b7280;
		margin-bottom: 12px;
		transition: all 0.2s;
	}

	.back-button:hover {
		background: #f9fafb;
		border-color: #d1d5db;
	}

	.page-header h1 {
		margin: 0;
		font-size: 28px;
		font-weight: 700;
		color: #1f2937;
	}

	.error-banner {
		padding: 12px 24px;
		background: #fef2f2;
		border-bottom: 1px solid #fecaca;
		color: #991b1b;
		font-size: 14px;
	}

	.domain-content {
		flex: 1;
		overflow-y: auto;
		padding: 24px;
	}

	.loading-state {
		display: flex;
		align-items: center;
		justify-content: center;
		height: 100%;
		color: #9ca3af;
		font-size: 16px;
	}

	.section {
		background: white;
		border: 1px solid #e5e7eb;
		border-radius: 8px;
		padding: 24px;
		margin-bottom: 24px;
	}

	.section h2 {
		margin: 0 0 16px;
		font-size: 18px;
		font-weight: 600;
		color: #1f2937;
	}

	.section-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 16px;
	}

	.section-header h2 {
		margin: 0;
	}

	.info-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 16px;
	}

	.info-card {
		padding: 12px;
		background: #f9fafb;
		border: 1px solid #e5e7eb;
		border-radius: 6px;
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	.info-label {
		font-size: 12px;
		font-weight: 600;
		color: #6b7280;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.info-value {
		font-size: 14px;
		font-weight: 500;
		color: #1f2937;
	}

	.empty-state {
		text-align: center;
		padding: 40px;
		color: #9ca3af;
	}

	.empty-state p {
		margin: 0 0 16px;
		font-size: 14px;
	}

	.datasets-table {
		overflow-x: auto;
	}

	table {
		width: 100%;
		border-collapse: collapse;
		font-size: 14px;
	}

	thead {
		background: #f9fafb;
		border-bottom: 1px solid #e5e7eb;
	}

	th {
		padding: 12px;
		text-align: left;
		font-weight: 600;
		color: #6b7280;
		text-transform: uppercase;
		letter-spacing: 0.5px;
		font-size: 12px;
	}

	td {
		padding: 12px;
		border-bottom: 1px solid #e5e7eb;
		color: #1f2937;
	}

	tbody tr:hover {
		background: #f9fafb;
	}

	.name-cell {
		font-weight: 500;
	}

	.dataset-link {
		color: #3b82f6;
		text-decoration: none;
		transition: color 0.2s;
	}

	.dataset-link:hover {
		color: #1e40af;
		text-decoration: underline;
	}

	.quality-badge {
		padding: 4px 8px;
		border-radius: 4px;
		font-size: 12px;
		font-weight: 600;
		display: inline-block;
	}

	.quality-badge.good {
		background: #ecfdf5;
		color: #065f46;
	}

	.quality-badge.fair {
		background: #fffbeb;
		color: #92400e;
	}

	.quality-badge.poor {
		background: #fef2f2;
		color: #7f1d1d;
	}

	.actions-cell {
		display: flex;
		gap: 8px;
	}

	.icon-button {
		padding: 6px 8px;
		background: white;
		border: 1px solid #e5e7eb;
		border-radius: 4px;
		cursor: pointer;
		font-size: 12px;
		transition: all 0.2s;
	}

	.icon-button:hover {
		background: #f9fafb;
		border-color: #d1d5db;
	}

	.primary-button,
	.secondary-button {
		padding: 10px 16px;
		border-radius: 6px;
		cursor: pointer;
		font-size: 14px;
		font-weight: 500;
		transition: all 0.2s;
		border: none;
	}

	.primary-button {
		background: #3b82f6;
		color: white;
	}

	.primary-button:hover {
		background: #2563eb;
	}

	.secondary-button {
		background: white;
		color: #1f2937;
		border: 1px solid #e5e7eb;
	}

	.secondary-button:hover {
		background: #f9fafb;
		border-color: #d1d5db;
	}
</style>
