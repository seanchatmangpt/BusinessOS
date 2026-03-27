<script lang="ts">
	import { onMount } from 'svelte';
	import LineageViewer from '$lib/components/LineageViewer.svelte';
	import QualityScoreboard from '$lib/components/QualityScoreboard.svelte';
	import { meshClient, type Domain, type Dataset, type QualityMetrics, type Lineage, type DataContract } from '$lib/api/mesh';

	let domains: Domain[] = [];
	let selectedDomain: Domain | null = null;
	let datasets: Dataset[] = [];
	let selectedDataset: Dataset | null = null;
	let quality: QualityMetrics | null = null;
	let lineage: Lineage | null = null;
	let contracts: DataContract[] = [];
	let showContracts = false;

	let isLoading = false;
	let error: string | null = null;

	onMount(async () => {
		await loadDomains();
	});

	async function loadDomains() {
		isLoading = true;
		error = null;

		try {
			domains = await meshClient.listDomains();
			if (domains.length > 0) {
				selectedDomain = domains[0];
				await loadDatasets(domains[0].id);
			}
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load domains';
		} finally {
			isLoading = false;
		}
	}

	async function loadDatasets(domainId: string) {
		isLoading = true;
		error = null;

		try {
			datasets = await meshClient.getDatasets(domainId);
			selectedDataset = null;
			quality = null;
			lineage = null;
			contracts = [];
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load datasets';
		} finally {
			isLoading = false;
		}
	}

	async function selectDataset(dataset: Dataset) {
		selectedDataset = dataset;
		isLoading = true;
		error = null;

		try {
			[quality, lineage, contracts] = await Promise.all([
				meshClient.getQuality(dataset.id),
				meshClient.getLineage(dataset.id, 5),
				meshClient.getContracts(dataset.id)
			]);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load dataset details';
		} finally {
			isLoading = false;
		}
	}

	function handleDomainChange(event: Event) {
		const select = event.target as HTMLSelectElement;
		const domainId = select.value;
		const domain = domains.find(d => d.id === domainId);
		if (domain) {
			selectedDomain = domain;
			loadDatasets(domain.id);
		}
	}
</script>

<div class="mesh-container">
	<header class="page-header">
		<h1>Data Mesh Discovery</h1>
		<p>Explore domains, datasets, lineage, and quality metrics</p>
	</header>

	{#if error}
		<div class="error-banner">
			<span>⚠️ {error}</span>
		</div>
	{/if}

	<div class="mesh-layout">
		<!-- Left panel: Domain and Dataset selectors -->
		<aside class="left-panel">
			<!-- Domain selector -->
			<div class="panel-section">
				<h2>Domains</h2>
				{#if isLoading && domains.length === 0}
					<div class="spinner">Loading domains...</div>
				{:else if domains.length === 0}
					<p class="empty-text">No domains available</p>
				{:else}
					<select class="domain-select" value={selectedDomain?.id || ''} on:change={handleDomainChange}>
						{#each domains as domain}
							<option value={domain.id}>{domain.name}</option>
						{/each}
					</select>
				{/if}
			</div>

			<!-- Domain details -->
			{#if selectedDomain}
				<div class="panel-section">
					<h3>Domain Details</h3>
					<div class="detail-item">
						<span class="detail-label">Owner:</span>
						<span class="detail-value">{selectedDomain.owner}</span>
					</div>
					<div class="detail-item">
						<span class="detail-label">Governance:</span>
						<span class="detail-value">{selectedDomain.governance_model}</span>
					</div>
					<div class="detail-item">
						<span class="detail-label">SLA:</span>
						<span class="detail-value">{selectedDomain.sla}</span>
					</div>
				</div>
			{/if}

			<!-- Dataset list -->
			<div class="panel-section">
				<h2>Datasets</h2>
				{#if isLoading && datasets.length === 0}
					<div class="spinner">Loading datasets...</div>
				{:else if datasets.length === 0}
					<p class="empty-text">No datasets in this domain</p>
				{:else}
					<div class="dataset-list">
						{#each datasets as dataset (dataset.id)}
							<button
								class="dataset-item"
								class:active={selectedDataset?.id === dataset.id}
								on:click={() => selectDataset(dataset)}
								disabled={isLoading}
							>
								<div class="dataset-name">{dataset.name}</div>
								<div class="dataset-quality">
									<span class="quality-badge" class:good={dataset.quality_score >= 80} class:fair={dataset.quality_score >= 60} class:poor={dataset.quality_score < 60}>
										{Math.round(dataset.quality_score)}
									</span>
								</div>
							</button>
						{/each}
					</div>
				{/if}
			</div>
		</aside>

		<!-- Right panel: Details -->
		<main class="right-panel">
			{#if !selectedDataset}
				<div class="empty-state">
					<p>Select a dataset to view details</p>
				</div>
			{:else}
				<div class="details-tabs">
					<!-- Quality Score Tab -->
					<div class="tab-section">
						<h2>Quality Metrics</h2>
						<QualityScoreboard {quality} />
					</div>

					<!-- Lineage Tab -->
					<div class="tab-section">
						<h2>Data Lineage</h2>
						<div class="lineage-wrapper">
							<LineageViewer {lineage} selectedNodeId={selectedDataset.id} />
						</div>
					</div>

					<!-- Contracts Tab -->
					<div class="tab-section">
						<h2>Data Contracts</h2>
						<button
							class="toggle-contracts"
							on:click={() => (showContracts = !showContracts)}
						>
							{showContracts ? '▼' : '▶'} Show Constraints ({contracts.length})
						</button>

						{#if showContracts}
							<div class="contracts-list">
								{#each contracts as contract (contract.id)}
									<div class="contract-card">
										<h4>{contract.name}</h4>
										<div class="constraints">
											{#each contract.constraints as constraint}
												<div
													class="constraint"
													class:severity-error={constraint.severity === 'error'}
													class:severity-warn={constraint.severity === 'warn'}
												>
													<span class="constraint-field">{constraint.field}</span>
													<span class="constraint-rule">{constraint.rule}</span>
													<span class="constraint-severity">{constraint.severity}</span>
												</div>
											{/each}
										</div>
									</div>
								{/each}

								{#if contracts.length === 0}
									<p class="empty-text">No data contracts defined</p>
								{/if}
							</div>
						{/if}
					</div>
				</div>
			{/if}
		</main>
	</div>
</div>

<style>
	.mesh-container {
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

	.page-header h1 {
		margin: 0 0 8px;
		font-size: 28px;
		font-weight: 700;
		color: #1f2937;
	}

	.page-header p {
		margin: 0;
		font-size: 14px;
		color: #6b7280;
	}

	.error-banner {
		padding: 12px 24px;
		background: #fef2f2;
		border-bottom: 1px solid #fecaca;
		color: #991b1b;
		font-size: 14px;
	}

	.mesh-layout {
		display: flex;
		flex: 1;
		overflow: hidden;
	}

	.left-panel {
		width: 280px;
		background: white;
		border-right: 1px solid #e5e7eb;
		overflow-y: auto;
		padding: 16px;
		display: flex;
		flex-direction: column;
		gap: 16px;
	}

	.right-panel {
		flex: 1;
		overflow-y: auto;
		padding: 24px;
	}

	.panel-section {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}

	.panel-section h2 {
		margin: 0;
		font-size: 14px;
		font-weight: 600;
		color: #1f2937;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.panel-section h3 {
		margin: 0;
		font-size: 13px;
		font-weight: 600;
		color: #374151;
	}

	.domain-select {
		padding: 8px 12px;
		border: 1px solid #e5e7eb;
		border-radius: 6px;
		font-size: 14px;
		background: white;
		cursor: pointer;
	}

	.domain-select:focus {
		outline: none;
		border-color: #3b82f6;
		box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
	}

	.detail-item {
		display: flex;
		flex-direction: column;
		gap: 4px;
		padding: 8px;
		background: #f9fafb;
		border-radius: 4px;
		font-size: 12px;
	}

	.detail-label {
		color: #6b7280;
		text-transform: uppercase;
		letter-spacing: 0.5px;
		font-weight: 600;
	}

	.detail-value {
		color: #1f2937;
		font-weight: 500;
	}

	.dataset-list {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.dataset-item {
		padding: 12px;
		background: white;
		border: 1px solid #e5e7eb;
		border-radius: 6px;
		cursor: pointer;
		display: flex;
		justify-content: space-between;
		align-items: center;
		transition: all 0.2s;
		font-size: 14px;
	}

	.dataset-item:hover:not(:disabled) {
		background: #f9fafb;
		border-color: #d1d5db;
	}

	.dataset-item.active {
		background: #dbeafe;
		border-color: #3b82f6;
		color: #1e40af;
	}

	.dataset-item:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.dataset-name {
		font-weight: 500;
		flex: 1;
	}

	.dataset-quality {
		margin-left: 8px;
	}

	.quality-badge {
		padding: 4px 8px;
		border-radius: 4px;
		font-size: 12px;
		font-weight: 600;
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

	.empty-state {
		display: flex;
		align-items: center;
		justify-content: center;
		height: 100%;
		color: #9ca3af;
		font-size: 16px;
	}

	.empty-text {
		color: #9ca3af;
		font-size: 13px;
		text-align: center;
		padding: 16px;
	}

	.spinner {
		text-align: center;
		color: #9ca3af;
		font-size: 13px;
	}

	.details-tabs {
		display: flex;
		flex-direction: column;
		gap: 32px;
	}

	.tab-section {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}

	.tab-section h2 {
		margin: 0;
		font-size: 18px;
		font-weight: 600;
		color: #1f2937;
	}

	.lineage-wrapper {
		height: 400px;
		border: 1px solid #e5e7eb;
		border-radius: 8px;
		overflow: hidden;
	}

	.toggle-contracts {
		padding: 10px 12px;
		background: white;
		border: 1px solid #e5e7eb;
		border-radius: 6px;
		cursor: pointer;
		font-size: 14px;
		font-weight: 500;
		color: #1f2937;
		transition: all 0.2s;
		text-align: left;
	}

	.toggle-contracts:hover {
		background: #f9fafb;
		border-color: #d1d5db;
	}

	.contracts-list {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}

	.contract-card {
		padding: 12px;
		background: white;
		border: 1px solid #e5e7eb;
		border-radius: 6px;
	}

	.contract-card h4 {
		margin: 0 0 8px;
		font-size: 13px;
		font-weight: 600;
		color: #1f2937;
	}

	.constraints {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}

	.constraint {
		padding: 6px 8px;
		background: #f9fafb;
		border-radius: 4px;
		display: flex;
		gap: 8px;
		font-size: 12px;
	}

	.constraint.severity-error {
		background: #fef2f2;
		border-left: 3px solid #ef4444;
	}

	.constraint.severity-warn {
		background: #fffbeb;
		border-left: 3px solid #f59e0b;
	}

	.constraint-field {
		font-weight: 600;
		color: #1f2937;
		min-width: 80px;
	}

	.constraint-rule {
		color: #6b7280;
		flex: 1;
	}

	.constraint-severity {
		text-transform: uppercase;
		font-weight: 600;
		font-size: 10px;
		color: #6b7280;
	}
</style>
