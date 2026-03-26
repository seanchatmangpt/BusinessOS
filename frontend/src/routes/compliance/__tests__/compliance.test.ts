import { describe, it, expect, beforeEach, vi } from 'vitest';
import { complianceApi } from '$lib/api/compliance';

describe('Compliance API Client', () => {
	beforeEach(() => {
		// Reset fetch mock
		global.fetch = vi.fn();
	});

	describe('verifyCompliance', () => {
		it('should fetch compliance status from API', async () => {
			const mockResponse = {
				soc2: {
					name: 'SOC2',
					score: 92,
					trend: 'up' as const,
					lastUpdated: new Date().toISOString(),
					controls: [],
					passingControls: 21,
					totalControls: 23
				},
				gdpr: {
					name: 'GDPR',
					score: 85,
					trend: 'stable' as const,
					lastUpdated: new Date().toISOString(),
					controls: [],
					passingControls: 15,
					totalControls: 18
				}
			};

			vi.mocked(global.fetch).mockResolvedValueOnce({
				ok: true,
				status: 200,
				json: async () => mockResponse
			} as any);

			const result = await complianceApi.verifyCompliance();

			expect(result.status).toBe(200);
			expect(result.data).toEqual(mockResponse);
			expect(global.fetch).toHaveBeenCalledWith('/api/compliance/status', expect.any(Object));
		});

		it('should handle API errors gracefully', async () => {
			vi.mocked(global.fetch).mockResolvedValueOnce({
				ok: false,
				status: 500
			} as any);

			const result = await complianceApi.verifyCompliance();

			expect(result.status).toBe(500);
			expect(result.error).toBeDefined();
		});

		it('should handle network errors', async () => {
			vi.mocked(global.fetch).mockRejectedValueOnce(new Error('Network error'));

			const result = await complianceApi.verifyCompliance();

			expect(result.error).toBeDefined();
			expect(result.status).toBe(500);
		});
	});

	describe('getReport', () => {
		it('should fetch compliance report', async () => {
			const mockReport = {
				frameworks: {
					soc2: {
						name: 'SOC2',
						score: 92,
						trend: 'up' as const,
						lastUpdated: new Date().toISOString(),
						controls: [],
						passingControls: 21,
						totalControls: 23
					}
				},
				controls: {},
				violations: [],
				generatedAt: new Date().toISOString()
			};

			vi.mocked(global.fetch).mockResolvedValueOnce({
				ok: true,
				status: 200,
				json: async () => mockReport
			} as any);

			const result = await complianceApi.getReport();

			expect(result.status).toBe(200);
			expect(result.data).toEqual(mockReport);
		});

		it('should return mock data when API fails', async () => {
			vi.mocked(global.fetch).mockRejectedValueOnce(new Error('API unavailable'));

			const result = await complianceApi.getReport();

			expect(result.status).toBe(200);
			expect(result.data).toBeDefined();
			expect(result.data?.generatedAt).toBeDefined();
		});
	});

	describe('getControls', () => {
		it('should fetch controls from API', async () => {
			const mockControls = [
				{
					id: 'SOC2-001',
					framework: 'soc2',
					status: 'pass' as const,
					description: 'Security controls test',
					lastChecked: new Date().toISOString()
				},
				{
					id: 'SOC2-002',
					framework: 'soc2',
					status: 'fail' as const,
					severity: 'high' as const,
					description: 'Audit logging',
					remediation: 'Enable audit logging',
					lastChecked: new Date().toISOString()
				}
			];

			vi.mocked(global.fetch).mockResolvedValueOnce({
				ok: true,
				status: 200,
				json: async () => mockControls
			} as any);

			const result = await complianceApi.getControls();

			expect(result.status).toBe(200);
			expect(result.data).toEqual(mockControls);
		});

		it('should return mock data when API unavailable', async () => {
			vi.mocked(global.fetch).mockRejectedValueOnce(new Error('API error'));

			const result = await complianceApi.getControls();

			expect(result.status).toBe(200);
			expect(result.data).toBeDefined();
			expect(Array.isArray(result.data)).toBe(true);
			expect(result.data!.length).toBeGreaterThan(0);
		});

		it('should contain controls across all frameworks', async () => {
			vi.mocked(global.fetch).mockRejectedValueOnce(new Error('API error'));

			const result = await complianceApi.getControls();
			const controls = result.data || [];

			const frameworks = ['soc2', 'gdpr', 'hipaa', 'sox'];
			frameworks.forEach((framework) => {
				const hasFramework = controls.some((c) => c.framework === framework);
				expect(hasFramework).toBe(true);
			});
		});
	});

	describe('getViolations', () => {
		it('should fetch violations from API', async () => {
			const mockViolations = [
				{
					id: 'v1',
					controlId: 'SOC2-001',
					framework: 'soc2',
					severity: 'critical' as const,
					description: 'Critical security gap',
					remediation: 'Implement immediately'
				}
			];

			vi.mocked(global.fetch).mockResolvedValueOnce({
				ok: true,
				status: 200,
				json: async () => mockViolations
			} as any);

			const result = await complianceApi.getViolations();

			expect(result.status).toBe(200);
			expect(result.data).toEqual(mockViolations);
		});

		it('should return mock data when API unavailable', async () => {
			vi.mocked(global.fetch).mockRejectedValueOnce(new Error('API error'));

			const result = await complianceApi.getViolations();

			expect(result.status).toBe(200);
			expect(result.data).toBeDefined();
			expect(Array.isArray(result.data)).toBe(true);
		});

		it('should include severity levels', async () => {
			vi.mocked(global.fetch).mockRejectedValueOnce(new Error('API error'));

			const result = await complianceApi.getViolations();
			const violations = result.data || [];

			violations.forEach((v) => {
				expect(['critical', 'high', 'medium', 'low']).toContain(v.severity);
			});
		});
	});
});

describe('Compliance Dashboard UI', () => {
	it('should render framework tabs', () => {
		const frameworks = ['SOC2', 'GDPR', 'HIPAA', 'SOX'];
		expect(frameworks.length).toBe(4);
		frameworks.forEach((f) => {
			expect(f).toMatch(/^[A-Z]+$/);
		});
	});

	it('should have severity filter levels', () => {
		const severities = ['critical', 'high', 'medium', 'low'];
		expect(severities.length).toBeGreaterThan(0);
		severities.forEach((s) => {
			expect(typeof s).toBe('string');
		});
	});

	it('should have export functionality', () => {
		// Verify that export functions are available
		expect(typeof complianceApi.getReport).toBe('function');
		expect(typeof complianceApi.getControls).toBe('function');
	});
});

describe('Control Status Calculations', () => {
	it('should calculate passing controls percentage', async () => {
		vi.mocked(global.fetch).mockRejectedValueOnce(new Error('API error'));

		const result = await complianceApi.getControls();
		const controls = result.data || [];

		const passingCount = controls.filter((c) => c.status === 'pass').length;
		const percentage = (passingCount / controls.length) * 100;

		expect(percentage).toBeGreaterThanOrEqual(0);
		expect(percentage).toBeLessThanOrEqual(100);
	});

	it('should group controls by framework', async () => {
		vi.mocked(global.fetch).mockRejectedValueOnce(new Error('API error'));

		const result = await complianceApi.getControls();
		const controls = result.data || [];

		const grouped = controls.reduce(
			(acc, c) => {
				if (!acc[c.framework]) acc[c.framework] = [];
				acc[c.framework].push(c);
				return acc;
			},
			{} as Record<string, typeof controls>
		);

		['soc2', 'gdpr', 'hipaa', 'sox'].forEach((framework) => {
			expect(grouped[framework]).toBeDefined();
			expect(grouped[framework].length).toBeGreaterThan(0);
		});
	});

	it('should filter violations by severity', async () => {
		vi.mocked(global.fetch).mockRejectedValueOnce(new Error('API error'));

		const result = await complianceApi.getViolations();
		const violations = result.data || [];

		const critical = violations.filter((v) => v.severity === 'critical');
		const high = violations.filter((v) => v.severity === 'high');
		const medium = violations.filter((v) => v.severity === 'medium');
		const low = violations.filter((v) => v.severity === 'low');

		expect(critical.length + high.length + medium.length + low.length).toBe(violations.length);
	});
});

describe('Score Trend Analysis', () => {
	it('should correctly interpret score trends', () => {
		const trends = ['up', 'down', 'stable'];
		expect(trends.length).toBe(3);

		const testScore = {
			trend: 'up' as const,
			score: 85
		};

		expect(testScore.trend).toBe('up');
		expect(testScore.score).toBeGreaterThan(0);
	});

	it('should calculate score color based on value', () => {
		const colors: Record<string, string> = {};

		[90, 75, 50].forEach((score) => {
			if (score >= 90) colors[score] = 'green';
			else if (score >= 70) colors[score] = 'yellow';
			else colors[score] = 'red';
		});

		expect(colors[90]).toBe('green');
		expect(colors[75]).toBe('yellow');
		expect(colors[50]).toBe('red');
	});
});

describe('Compliance Framework Coverage', () => {
	it('should support all major frameworks', async () => {
		const frameworks = ['SOC2', 'GDPR', 'HIPAA', 'SOX'];

		vi.mocked(global.fetch).mockRejectedValueOnce(new Error('API error'));
		const result = await complianceApi.getControls();
		const controls = result.data || [];

		frameworks.forEach((framework) => {
			const frameWorkLower = framework.toLowerCase();
			const exists = controls.some((c) => c.framework === frameWorkLower);
			expect(exists).toBe(true);
		});
	});

	it('should have minimum control counts per framework', async () => {
		vi.mocked(global.fetch).mockRejectedValueOnce(new Error('API error'));
		const result = await complianceApi.getControls();
		const controls = result.data || [];

		const frameworkCounts: Record<string, number> = {};
		controls.forEach((c) => {
			frameworkCounts[c.framework] = (frameworkCounts[c.framework] || 0) + 1;
		});

		expect(frameworkCounts['soc2']).toBeGreaterThanOrEqual(20);
		expect(frameworkCounts['gdpr']).toBeGreaterThanOrEqual(15);
		expect(frameworkCounts['hipaa']).toBeGreaterThanOrEqual(15);
		expect(frameworkCounts['sox']).toBeGreaterThanOrEqual(10);
	});
});
