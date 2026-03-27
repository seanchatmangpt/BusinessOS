<script lang="ts">
	import type { Lineage, LineageNode } from '$lib/api/mesh';

	export let lineage: Lineage | null = null;
	export let selectedNodeId: string | null = null;

	let expandedNodes = new Set<string>();

	function toggleNode(nodeId: string) {
		if (expandedNodes.has(nodeId)) {
			expandedNodes.delete(nodeId);
		} else {
			expandedNodes.add(nodeId);
		}
		expandedNodes = expandedNodes;
	}

	function getQualityColor(score: number): string {
		if (score >= 80) return '#10b981'; // green
		if (score >= 60) return '#f59e0b'; // yellow
		return '#ef4444'; // red
	}

	function getNodesByLevel(level: number): LineageNode[] {
		return (lineage?.nodes || []).filter(n => n.level === level);
	}

	function getOutgoingEdges(nodeId: string): string[] {
		return (lineage?.edges || [])
			.filter(e => e.source_id === nodeId)
			.map(e => e.target_id);
	}
</script>

<div class="lineage-viewer">
	{#if !lineage}
		<div class="empty-state">
			<p>Select a dataset to view lineage</p>
		</div>
	{:else}
		<div class="lineage-container">
			<!-- Level-based rendering: max 5 levels -->
			{#each Array.from({ length: Math.min(lineage.max_depth, 5) }) as _, levelIdx}
				{@const level = levelIdx}
				{@const nodesAtLevel = getNodesByLevel(level)}

				{#if nodesAtLevel.length > 0}
					<div class="lineage-level" data-level={level}>
						<div class="level-label">Level {level}</div>
						<div class="nodes-container">
							{#each nodesAtLevel as node (node.id)}
								<div
									class="lineage-node"
									class:selected={selectedNodeId === node.id}
									on:click={() => {
										selectedNodeId = node.id;
										toggleNode(node.id);
									}}
									role="button"
									tabindex="0"
								>
									<!-- Node circle with quality score -->
									<svg width="120" height="120" viewBox="0 0 120 120">
										<!-- Background circle -->
										<circle cx="60" cy="60" r="55" fill="white" stroke="#e5e7eb" stroke-width="2" />

										<!-- Quality indicator ring -->
										<circle
											cx="60"
											cy="60"
											r="55"
											fill="none"
											stroke={getQualityColor(node.quality_score)}
											stroke-width="8"
											opacity="0.3"
										/>

										<!-- Quality text -->
										<text
											x="60"
											y="65"
											text-anchor="middle"
											font-size="24"
											font-weight="bold"
											fill="#1f2937"
										>
											{node.quality_score}
										</text>
										<text
											x="60"
											y="80"
											text-anchor="middle"
											font-size="10"
											fill="#6b7280"
										>
											score
										</text>
									</svg>

									<!-- Node name -->
									<div class="node-info">
										<div class="node-name">{node.dataset_name}</div>
										<div class="node-id">ID: {node.dataset_id.slice(0, 8)}</div>
									</div>

									<!-- Expanded detail -->
									{#if expandedNodes.has(node.id)}
										<div class="node-detail">
											<div class="detail-row">
												<span>Quality:</span>
												<span class="quality-badge">{node.quality_score}</span>
											</div>
											<div class="detail-row">
												<span>Level:</span>
												<span>{node.level}</span>
											</div>
										</div>
									{/if}
								</div>

								<!-- Draw arrows to children -->
								{#each getOutgoingEdges(node.id) as targetId}
									<svg class="edge-arrow" viewBox="0 0 20 40" preserveAspectRatio="none">
										<path
											d="M 10 0 L 10 30 M 7 27 L 10 30 L 13 27"
											stroke="#9ca3af"
											stroke-width="1.5"
											fill="none"
										/>
									</svg>
								{/each}
							{/each}
						</div>
					</div>
				{/if}
			{/each}

			<!-- Max depth indicator -->
			{#if lineage.max_depth > 5}
				<div class="max-depth-notice">
					<p>Lineage depth limited to 5 levels. Full lineage has {lineage.max_depth} levels.</p>
				</div>
			{/if}
		</div>
	{/if}
</div>

<style>
	.lineage-viewer {
		width: 100%;
		height: 100%;
		overflow: auto;
		background: #f9fafb;
		border: 1px solid #e5e7eb;
		border-radius: 8px;
		padding: 16px;
	}

	.empty-state {
		display: flex;
		align-items: center;
		justify-content: center;
		height: 300px;
		color: #9ca3af;
		font-size: 14px;
	}

	.lineage-container {
		display: flex;
		flex-direction: column;
		gap: 32px;
	}

	.lineage-level {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}

	.level-label {
		font-size: 12px;
		font-weight: 600;
		color: #6b7280;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.nodes-container {
		display: flex;
		flex-wrap: wrap;
		gap: 16px;
		align-items: flex-start;
	}

	.lineage-node {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 8px;
		padding: 8px;
		border-radius: 8px;
		cursor: pointer;
		transition: all 0.2s ease;
		border: 2px solid transparent;
	}

	.lineage-node:hover {
		background: #f3f4f6;
		border-color: #e5e7eb;
	}

	.lineage-node.selected {
		background: #dbeafe;
		border-color: #3b82f6;
	}

	.lineage-node svg {
		width: 100px;
		height: 100px;
	}

	.node-info {
		text-align: center;
		max-width: 120px;
	}

	.node-name {
		font-size: 12px;
		font-weight: 600;
		color: #1f2937;
		word-break: break-word;
	}

	.node-id {
		font-size: 10px;
		color: #9ca3af;
	}

	.node-detail {
		width: 100%;
		margin-top: 8px;
		padding: 8px;
		background: white;
		border-radius: 4px;
		border: 1px solid #e5e7eb;
		font-size: 11px;
	}

	.detail-row {
		display: flex;
		justify-content: space-between;
		padding: 4px 0;
	}

	.quality-badge {
		font-weight: 600;
		padding: 2px 6px;
		border-radius: 3px;
		background: #f0fdf4;
		color: #15803d;
	}

	.edge-arrow {
		width: 100%;
		height: 40px;
		flex: 0 0 100%;
	}

	.max-depth-notice {
		padding: 12px;
		background: #fef3c7;
		border: 1px solid #fcd34d;
		border-radius: 6px;
		color: #78350f;
		font-size: 12px;
		text-align: center;
	}

	.max-depth-notice p {
		margin: 0;
	}
</style>
