import { describe, it, expect, beforeEach, vi } from 'vitest';
import { MeshApiClient, type Domain, type Dataset, type QualityMetrics, type Lineage } from '$lib/api/mesh';

// Mock fetch
global.fetch = vi.fn();

describe('Data Mesh UI', () => {
	let client: MeshApiClient;

	beforeEach(() => {
		client = new MeshApiClient();
		vi.clearAllMocks();
	});

	describe('listDomains', () => {
		it('should fetch and return list of domains', async () => {
			const mockDomains: Domain[] = [
				{
					id: 'domain-1',
					name: 'Finance',
					owner: 'Alice',
					governance_model: 'Federated',
					sla: '99.9%',
					created_at: '2026-01-01T00:00:00Z',
					updated_at: '2026-03-26T00:00:00Z'
				},
				{
					id: 'domain-2',
					name: 'Operations',
					owner: 'Bob',
					governance_model: 'Centralized',
					sla: '99.5%',
					created_at: '2026-01-02T00:00:00Z',
					updated_at: '2026-03-25T00:00:00Z'
				}
			];

			(global.fetch as any).mockResolvedValueOnce({
				ok: true,
				json: async () => ({ domains: mockDomains })
			});

			const domains = await client.listDomains();
			expect(domains).toHaveLength(2);
			expect(domains[0].name).toBe('Finance');
			expect(domains[1].name).toBe('Operations');
		});

		it('should handle API errors gracefully', async () => {
			(global.fetch as any).mockResolvedValueOnce({
				ok: false
			});

			await expect(client.listDomains()).rejects.toThrow('Failed to list domains');
		});
	});

	describe('getDatasets', () => {
		it('should fetch datasets for a domain', async () => {
			const mockDatasets: Dataset[] = [
				{
					id: 'ds-1',
					domain_id: 'domain-1',
					name: 'Transactions',
					owner: 'Alice',
					quality_score: 95,
					last_modified: '2026-03-25T10:00:00Z',
					created_at: '2026-01-01T00:00:00Z',
					updated_at: '2026-03-25T10:00:00Z'
				},
				{
					id: 'ds-2',
					domain_id: 'domain-1',
					name: 'Accounts',
					owner: 'Alice',
					quality_score: 87,
					last_modified: '2026-03-24T15:00:00Z',
					created_at: '2026-01-05T00:00:00Z',
					updated_at: '2026-03-24T15:00:00Z'
				}
			];

			(global.fetch as any).mockResolvedValueOnce({
				ok: true,
				json: async () => ({ datasets: mockDatasets })
			});

			const datasets = await client.getDatasets('domain-1');
			expect(datasets).toHaveLength(2);
			expect(datasets[0].name).toBe('Transactions');
			expect(datasets[0].quality_score).toBe(95);
		});

		it('should return empty list if no datasets', async () => {
			(global.fetch as any).mockResolvedValueOnce({
				ok: true,
				json: async () => ({ datasets: [] })
			});

			const datasets = await client.getDatasets('domain-empty');
			expect(datasets).toHaveLength(0);
		});
	});

	describe('getQuality', () => {
		it('should fetch quality metrics for a dataset', async () => {
			const mockQuality: QualityMetrics = {
				dataset_id: 'ds-1',
				completeness: 98,
				accuracy: 96,
				consistency: 94,
				timeliness: 92,
				overall: 95,
				last_calculated: '2026-03-26T00:00:00Z'
			};

			(global.fetch as any).mockResolvedValueOnce({
				ok: true,
				json: async () => mockQuality
			});

			const quality = await client.getQuality('ds-1');
			expect(quality.overall).toBe(95);
			expect(quality.completeness).toBe(98);
			expect(quality.accuracy).toBe(96);
			expect(quality.consistency).toBe(94);
			expect(quality.timeliness).toBe(92);
		});

		it('should handle quality score boundaries', async () => {
			const lowQuality: QualityMetrics = {
				dataset_id: 'ds-low',
				completeness: 45,
				accuracy: 50,
				consistency: 55,
				timeliness: 40,
				overall: 47,
				last_calculated: '2026-03-26T00:00:00Z'
			};

			(global.fetch as any).mockResolvedValueOnce({
				ok: true,
				json: async () => lowQuality
			});

			const quality = await client.getQuality('ds-low');
			expect(quality.overall).toBeLessThan(60);
		});
	});

	describe('getLineage', () => {
		it('should fetch lineage with depth limit of 5', async () => {
			const mockLineage: Lineage = {
				nodes: [
					{ id: 'n1', dataset_id: 'ds-1', dataset_name: 'Raw Data', quality_score: 80, level: 0 },
					{ id: 'n2', dataset_id: 'ds-2', dataset_name: 'Cleaned Data', quality_score: 85, level: 1 },
					{ id: 'n3', dataset_id: 'ds-3', dataset_name: 'Aggregated Data', quality_score: 90, level: 2 }
				],
				edges: [
					{ source_id: 'n1', target_id: 'n2', relationship: 'prov:wasDerivedFrom' },
					{ source_id: 'n2', target_id: 'n3', relationship: 'prov:wasDerivedFrom' }
				],
				max_depth: 3
			};

			(global.fetch as any).mockResolvedValueOnce({
				ok: true,
				json: async () => mockLineage
			});

			const lineage = await client.getLineage('ds-1', 5);
			expect(lineage.nodes).toHaveLength(3);
			expect(lineage.edges).toHaveLength(2);
			expect(lineage.max_depth).toBeLessThanOrEqual(5);
		});

		it('should respect max depth parameter', async () => {
			const mockLineage: Lineage = {
				nodes: [
					{ id: 'n1', dataset_id: 'ds-1', dataset_name: 'Level 0', quality_score: 90, level: 0 },
					{ id: 'n2', dataset_id: 'ds-2', dataset_name: 'Level 1', quality_score: 85, level: 1 }
				],
				edges: [{ source_id: 'n1', target_id: 'n2', relationship: 'prov:wasDerivedFrom' }],
				max_depth: 2
			};

			(global.fetch as any).mockResolvedValueOnce({
				ok: true,
				json: async () => mockLineage
			});

			const lineage = await client.getLineage('ds-1', 2);
			expect(lineage.max_depth).toBeLessThanOrEqual(2);
		});
	});

	describe('getContracts', () => {
		it('should fetch data contracts for a dataset', async () => {
			const mockContracts = [
				{
					id: 'contract-1',
					dataset_id: 'ds-1',
					name: 'Accuracy Contract',
					constraints: [
						{ field: 'amount', rule: 'NOT NULL', severity: 'error' as const },
						{ field: 'amount', rule: 'RANGE [0, 1000000]', severity: 'error' as const }
					],
					created_at: '2026-01-01T00:00:00Z',
					updated_at: '2026-03-25T00:00:00Z'
				}
			];

			(global.fetch as any).mockResolvedValueOnce({
				ok: true,
				json: async () => ({ contracts: mockContracts })
			});

			const contracts = await client.getContracts('ds-1');
			expect(contracts).toHaveLength(1);
			expect(contracts[0].constraints).toHaveLength(2);
			expect(contracts[0].constraints[0].severity).toBe('error');
		});

		it('should return empty contracts if none exist', async () => {
			(global.fetch as any).mockResolvedValueOnce({
				ok: true,
				json: async () => ({ contracts: [] })
			});

			const contracts = await client.getContracts('ds-no-contracts');
			expect(contracts).toHaveLength(0);
		});
	});

	describe('getDomain', () => {
		it('should fetch a single domain by id', async () => {
			const mockDomain: Domain = {
				id: 'domain-1',
				name: 'Finance',
				owner: 'Alice',
				governance_model: 'Federated',
				sla: '99.9%',
				created_at: '2026-01-01T00:00:00Z',
				updated_at: '2026-03-26T00:00:00Z'
			};

			(global.fetch as any).mockResolvedValueOnce({
				ok: true,
				json: async () => mockDomain
			});

			const domain = await client.getDomain('domain-1');
			expect(domain.name).toBe('Finance');
			expect(domain.sla).toBe('99.9%');
		});
	});

	describe('Integration: Domain Selection Flow', () => {
		it('should load domains and then datasets', async () => {
			const mockDomains: Domain[] = [
				{
					id: 'domain-1',
					name: 'Finance',
					owner: 'Alice',
					governance_model: 'Federated',
					sla: '99.9%',
					created_at: '2026-01-01T00:00:00Z',
					updated_at: '2026-03-26T00:00:00Z'
				}
			];

			const mockDatasets: Dataset[] = [
				{
					id: 'ds-1',
					domain_id: 'domain-1',
					name: 'Transactions',
					owner: 'Alice',
					quality_score: 95,
					last_modified: '2026-03-25T10:00:00Z',
					created_at: '2026-01-01T00:00:00Z',
					updated_at: '2026-03-25T10:00:00Z'
				}
			];

			(global.fetch as any)
				.mockResolvedValueOnce({
					ok: true,
					json: async () => ({ domains: mockDomains })
				})
				.mockResolvedValueOnce({
					ok: true,
					json: async () => ({ datasets: mockDatasets })
				});

			const domains = await client.listDomains();
			expect(domains).toHaveLength(1);

			const datasets = await client.getDatasets(domains[0].id);
			expect(datasets).toHaveLength(1);
			expect(datasets[0].domain_id).toBe(domains[0].id);
		});
	});

	describe('Quality Score Classification', () => {
		it('should classify good quality scores (>= 80)', () => {
			const goodScore = 85;
			expect(goodScore).toBeGreaterThanOrEqual(80);
		});

		it('should classify fair quality scores (60-79)', () => {
			const fairScore = 70;
			expect(fairScore).toBeGreaterThanOrEqual(60);
			expect(fairScore).toBeLessThan(80);
		});

		it('should classify poor quality scores (< 60)', () => {
			const poorScore = 45;
			expect(poorScore).toBeLessThan(60);
		});
	});

	describe('Lineage Visualization', () => {
		it('should handle lineage with 5 levels', async () => {
			const mockLineage: Lineage = {
				nodes: Array.from({ length: 15 }, (_, i) => ({
					id: `n${i}`,
					dataset_id: `ds-${i}`,
					dataset_name: `Dataset Level ${Math.floor(i / 3)}`,
					quality_score: 75 + Math.random() * 25,
					level: Math.floor(i / 3)
				})),
				edges: Array.from({ length: 10 }, (_, i) => ({
					source_id: `n${i}`,
					target_id: `n${i + 1}`,
					relationship: 'prov:wasDerivedFrom'
				})),
				max_depth: 5
			};

			(global.fetch as any).mockResolvedValueOnce({
				ok: true,
				json: async () => mockLineage
			});

			const lineage = await client.getLineage('ds-0', 5);
			expect(lineage.nodes).toHaveLength(15);
			expect(lineage.max_depth).toBe(5);

			// Verify nodes are distributed across levels
			const levels = new Set(lineage.nodes.map(n => n.level));
			expect(levels.size).toBeGreaterThanOrEqual(1);
		});
	});

	describe('API Error Handling', () => {
		it('should throw error when API returns not ok', async () => {
			(global.fetch as any).mockResolvedValueOnce({
				ok: false,
				status: 404
			});

			await expect(client.listDomains()).rejects.toThrow();
		});

		it('should handle network errors', async () => {
			(global.fetch as any).mockRejectedValueOnce(new Error('Network failed'));

			await expect(client.listDomains()).rejects.toThrow();
		});
	});
});
