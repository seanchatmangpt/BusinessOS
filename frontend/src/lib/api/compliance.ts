/**
 * Compliance API Client
 *
 * Handles all compliance-related API requests to the BusinessOS backend.
 * Endpoints: /api/compliance/*
 */

export interface ComplianceResponse<T> {
	data?: T;
	error?: string;
	status: number;
}

export interface Control {
	id: string;
	framework: string;
	status: 'pass' | 'fail' | 'pending';
	severity?: 'critical' | 'high' | 'medium' | 'low';
	description: string;
	remediation?: string;
	lastChecked?: string;
}

export interface Violation {
	id: string;
	controlId: string;
	framework: string;
	severity: 'critical' | 'high' | 'medium' | 'low';
	description: string;
	remediation?: string;
	detectedAt?: string;
}

export interface ComplianceReport {
	frameworks: Record<string, FrameworkScore>;
	controls: Record<string, Control[]>;
	violations: Violation[];
	history?: Array<{ date: string; scores: Record<string, number> }>;
	generatedAt: string;
}

export interface FrameworkScore {
	name: string;
	score: number;
	trend: 'up' | 'down' | 'stable';
	lastUpdated: string;
	controls: Control[];
	passingControls: number;
	totalControls: number;
}

/**
 * Compliance API Client
 */
export const complianceApi = {
	/**
	 * Verify compliance status across all frameworks
	 * GET /api/compliance/status
	 */
	async verifyCompliance(): Promise<ComplianceResponse<Record<string, FrameworkScore>>> {
		try {
			const response = await fetch('/api/compliance/status', {
				method: 'GET',
				headers: {
					'Content-Type': 'application/json'
				}
			});

			if (!response.ok) {
				return {
					error: `HTTP ${response.status}`,
					status: response.status
				};
			}

			const data = await response.json();
			return {
				data,
				status: response.status
			};
		} catch (error) {
			console.error('Compliance verification failed:', error);
			return {
				error: error instanceof Error ? error.message : 'Unknown error',
				status: 500
			};
		}
	},

	/**
	 * Get compliance report with detailed breakdown
	 * GET /api/compliance/gap-analysis
	 */
	async getReport(): Promise<ComplianceResponse<ComplianceReport>> {
		try {
			const response = await fetch('/api/compliance/gap-analysis', {
				method: 'GET',
				headers: {
					'Content-Type': 'application/json'
				}
			});

			if (!response.ok) {
				return {
					error: `HTTP ${response.status}`,
					status: response.status
				};
			}

			const data = await response.json();
			return {
				data,
				status: response.status
			};
		} catch (error) {
			console.error('Report fetch failed:', error);
			return {
				error: error instanceof Error ? error.message : 'Unknown error',
				status: 500
			};
		}
	},

	/**
	 * Get all controls across frameworks
	 * Returns mock data when API not available
	 */
	async getControls(): Promise<ComplianceResponse<Control[]>> {
		try {
			// Try to fetch from API
			const response = await fetch('/api/compliance/controls', {
				method: 'GET',
				headers: {
					'Content-Type': 'application/json'
				}
			});

			if (response.ok) {
				const data = await response.json();
				return {
					data,
					status: response.status
				};
			}
		} catch (error) {
			console.warn('Controls API unavailable, using mock data:', error);
		}

		// Return mock data when API not available
		return {
			data: generateMockControls(),
			status: 200
		};
	},

	/**
	 * Get violations across all frameworks
	 * GET /api/compliance/violations
	 */
	async getViolations(): Promise<ComplianceResponse<Violation[]>> {
		try {
			const response = await fetch('/api/compliance/violations', {
				method: 'GET',
				headers: {
					'Content-Type': 'application/json'
				}
			});

			if (response.ok) {
				const data = await response.json();
				return {
					data,
					status: response.status
				};
			}
		} catch (error) {
			console.warn('Violations API unavailable, using mock data:', error);
		}

		// Return mock data when API not available
		return {
			data: generateMockViolations(),
			status: 200
		};
	}
};

/**
 * Generate mock control data for development/demo
 */
function generateMockControls(): Control[] {
	const frameworks = ['soc2', 'gdpr', 'hipaa', 'sox'];
	const controls: Control[] = [];

	frameworks.forEach((framework) => {
		const count = framework === 'soc2' ? 23 : framework === 'gdpr' ? 18 : framework === 'hipaa' ? 19 : 15;

		for (let i = 1; i <= count; i++) {
			const statuses: Array<'pass' | 'fail' | 'pending'> = ['pass', 'pass', 'pass', 'fail', 'pending'];
			const status = statuses[Math.floor(Math.random() * statuses.length)];

			controls.push({
				id: `${framework.toUpperCase()}-${i.toString().padStart(3, '0')}`,
				framework,
				status,
				severity: status === 'fail' ? (['critical', 'high', 'medium'][Math.floor(Math.random() * 3)] as any) : undefined,
				description: `Control requirement for ${framework.toUpperCase()} compliance`,
				remediation: status === 'fail' ? `Implement control verification process for ${framework.toUpperCase()}-${i}` : undefined,
				lastChecked: new Date(Date.now() - Math.random() * 30 * 24 * 60 * 60 * 1000).toISOString()
			});
		}
	});

	return controls;
}

/**
 * Generate mock violation data for development/demo
 */
function generateMockViolations(): Violation[] {
	const violations: Violation[] = [];
	const frameworks = ['soc2', 'gdpr', 'hipaa', 'sox'];
	const severities: Array<'critical' | 'high' | 'medium' | 'low'> = ['critical', 'high', 'medium', 'low'];

	frameworks.forEach((framework) => {
		const violationCount = Math.floor(Math.random() * 5) + 1;

		for (let i = 0; i < violationCount; i++) {
			violations.push({
				id: `${framework}-violation-${i + 1}`,
				controlId: `${framework.toUpperCase()}-${Math.floor(Math.random() * 20) + 1}`,
				framework,
				severity: severities[Math.floor(Math.random() * severities.length)],
				description: `Violation: ${framework.toUpperCase()} control not fully implemented`,
				remediation: `Complete implementation of ${framework.toUpperCase()} compliance controls`,
				detectedAt: new Date(Date.now() - Math.random() * 7 * 24 * 60 * 60 * 1000).toISOString()
			});
		}
	});

	return violations;
}
