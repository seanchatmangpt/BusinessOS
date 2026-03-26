import { describe, it, expect, beforeEach, vi } from 'vitest';
import { healthcareAPI } from '$lib/api/healthcare';
import type { Patient, AccessEvent, ConsentStatus, HIPAACompliance, PatientListResponse } from '$lib/api/healthcare';

// Mock fetch globally
global.fetch = vi.fn();

describe('Healthcare API Client', () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	describe('listPatients', () => {
		it('should list patients with pagination', async () => {
			const mockResponse: PatientListResponse = {
				patients: [
					{
						id: 'p1',
						firstName: 'John',
						lastName: 'Doe',
						dateOfBirth: '1990-01-15',
						mrn: 'MRN-001',
						status: 'active',
						createdAt: '2024-01-01T00:00:00Z',
						updatedAt: '2024-01-15T00:00:00Z',
						consentStatus: 'granted'
					}
				],
				total: 1,
				page: 1,
				limit: 20
			};

			(global.fetch as any).mockResolvedValueOnce({
				ok: true,
				json: async () => mockResponse
			});

			const result = await healthcareAPI.listPatients(1, 20);
			expect(result.patients).toHaveLength(1);
			expect(result.patients[0].firstName).toBe('John');
			expect(result.total).toBe(1);
		});

		it('should handle search query in listPatients', async () => {
			const mockResponse: PatientListResponse = {
				patients: [],
				total: 0,
				page: 1,
				limit: 20
			};

			(global.fetch as any).mockResolvedValueOnce({
				ok: true,
				json: async () => mockResponse
			});

			await healthcareAPI.listPatients(1, 20, 'John');

			const fetchCall = (global.fetch as any).mock.calls[0][0];
			expect(fetchCall).toContain('search=John');
		});

		it('should throw error on failed patient list fetch', async () => {
			(global.fetch as any).mockResolvedValueOnce({
				ok: false,
				statusText: 'Unauthorized'
			});

			await expect(healthcareAPI.listPatients()).rejects.toThrow('Failed to list patients');
		});
	});

	describe('getPatient', () => {
		it('should retrieve patient by ID', async () => {
			const mockPatient: Patient = {
				id: 'p1',
				firstName: 'Jane',
				lastName: 'Smith',
				dateOfBirth: '1985-06-20',
				mrn: 'MRN-002',
				status: 'active',
				createdAt: '2024-01-01T00:00:00Z',
				updatedAt: '2024-01-15T00:00:00Z',
				consentStatus: 'granted'
			};

			(global.fetch as any).mockResolvedValueOnce({
				ok: true,
				json: async () => mockPatient
			});

			const result = await healthcareAPI.getPatient('p1');
			expect(result.id).toBe('p1');
			expect(result.firstName).toBe('Jane');
		});

		it('should mask PII in display', () => {
			const masked = healthcareAPI.maskPII('Dr. John Doe', 'Cardiologist');
			expect(masked).toContain('J.');
			expect(masked).toContain('Doe');
			expect(masked).toContain('Cardiologist');
		});

		it('should handle PII masking without role', () => {
			const masked = healthcareAPI.maskPII('Alice Johnson');
			expect(masked).toContain('A.');
			expect(masked).toContain('Johnson');
		});
	});

	describe('trackPHI', () => {
		it('should track PHI access', async () => {
			(global.fetch as any).mockResolvedValueOnce({
				ok: true
			});

			await healthcareAPI.trackPHI('p1', 'Observation', 'read');

			expect(global.fetch).toHaveBeenCalledWith(
				expect.stringContaining('/patients/p1/track'),
				expect.objectContaining({
					method: 'POST'
				})
			);
		});

		it('should handle tracking error gracefully', async () => {
			(global.fetch as any).mockResolvedValueOnce({
				ok: false
			});

			// Should not throw
			await expect(healthcareAPI.trackPHI('p1', 'Observation', 'read')).resolves.toBeUndefined();
		});
	});

	describe('getAuditTrail', () => {
		it('should retrieve audit trail with last 50 events', async () => {
			const mockAuditTrail = {
				events: [
					{
						id: 'e1',
						patientId: 'p1',
						userId: 'u1',
						userName: 'Dr. Sarah Connor',
						userRole: 'Cardiologist',
						action: 'read' as const,
						resourceType: 'Observation',
						timestamp: '2024-01-15T10:00:00Z',
						reason: 'Patient consultation',
						ipAddress: '192.168.1.1',
						success: true
					}
				],
				total: 1,
				page: 1,
				limit: 50
			};

			(global.fetch as any).mockResolvedValueOnce({
				ok: true,
				json: async () => mockAuditTrail
			});

			const result = await healthcareAPI.getAuditTrail('p1');
			expect(result.events).toHaveLength(1);
			expect(result.events[0].action).toBe('read');
			expect(result.events[0].success).toBe(true);
		});

		it('should handle pagination in audit trail', async () => {
			(global.fetch as any).mockResolvedValueOnce({
				ok: true,
				json: async () => ({ events: [], total: 0, page: 2, limit: 25 })
			});

			await healthcareAPI.getAuditTrail('p1', 2, 25);

			const fetchCall = (global.fetch as any).mock.calls[0][0];
			expect(fetchCall).toContain('page=2');
			expect(fetchCall).toContain('limit=25');
		});
	});

	describe('verifyConsent', () => {
		it('should verify consent status for resources', async () => {
			const mockConsent: ConsentStatus = {
				patientId: 'p1',
				resourceTypes: {
					Observation: { granted: true, grantedAt: '2024-01-01T00:00:00Z' },
					Medication: { granted: false },
					Condition: { granted: true, grantedAt: '2024-01-05T00:00:00Z' }
				},
				updatedAt: '2024-01-15T00:00:00Z'
			};

			(global.fetch as any).mockResolvedValueOnce({
				ok: true,
				json: async () => mockConsent
			});

			const result = await healthcareAPI.verifyConsent('p1');
			expect(result.resourceTypes.Observation.granted).toBe(true);
			expect(result.resourceTypes.Medication.granted).toBe(false);
			expect(result.resourceTypes.Condition.granted).toBe(true);
		});
	});

	describe('verifyHIPAA', () => {
		it('should return HIPAA compliance status', async () => {
			const mockCompliance: HIPAACompliance = {
				accessControl: { passed: true, details: 'Role-based access control implemented' },
				auditLogging: { passed: true, details: 'All PHI access logged with timestamps' },
				encryption: { passed: true, details: 'TLS 1.3 for transit, AES-256 at rest' },
				integrity: { passed: true, details: 'Digital signatures and checksums verified' },
				score: 100,
				lastChecked: '2024-01-15T15:30:00Z'
			};

			(global.fetch as any).mockResolvedValueOnce({
				ok: true,
				json: async () => mockCompliance
			});

			const result = await healthcareAPI.verifyHIPAA('p1');
			expect(result.score).toBe(100);
			expect(result.accessControl.passed).toBe(true);
			expect(result.auditLogging.passed).toBe(true);
			expect(result.encryption.passed).toBe(true);
			expect(result.integrity.passed).toBe(true);
		});

		it('should handle partial HIPAA compliance', async () => {
			const mockCompliance: HIPAACompliance = {
				accessControl: { passed: true, details: 'OK' },
				auditLogging: { passed: false, details: 'Missing timestamps on 5% of events' },
				encryption: { passed: true, details: 'OK' },
				integrity: { passed: true, details: 'OK' },
				score: 75,
				lastChecked: '2024-01-15T15:30:00Z'
			};

			(global.fetch as any).mockResolvedValueOnce({
				ok: true,
				json: async () => mockCompliance
			});

			const result = await healthcareAPI.verifyHIPAA('p1');
			expect(result.score).toBe(75);
			expect(result.auditLogging.passed).toBe(false);
		});
	});

	describe('deletePHI', () => {
		it('should delete PHI with GDPR confirmation', async () => {
			const mockResponse = {
				success: true,
				graceUntil: '2024-02-15T23:59:59Z'
			};

			(global.fetch as any).mockResolvedValueOnce({
				ok: true,
				json: async () => mockResponse
			});

			const result = await healthcareAPI.deletePHI('p1', 'Patient requested deletion');
			expect(result.success).toBe(true);
			expect(result.graceUntil).toBeDefined();
		});

		it('should enforce deletion grace period', async () => {
			const graceDate = new Date();
			graceDate.setDate(graceDate.getDate() + 30); // 30-day grace period

			(global.fetch as any).mockResolvedValueOnce({
				ok: true,
				json: async () => ({
					success: true,
					graceUntil: graceDate.toISOString()
				})
			});

			const result = await healthcareAPI.deletePHI('p1', 'Reason');
			expect(new Date(result.graceUntil).getTime()).toBeGreaterThan(Date.now());
		});

		it('should throw error on deletion failure', async () => {
			(global.fetch as any).mockResolvedValueOnce({
				ok: false,
				statusText: 'Forbidden'
			});

			await expect(healthcareAPI.deletePHI('p1', 'Reason')).rejects.toThrow('Failed to delete PHI');
		});
	});

	describe('Audit Trail Events', () => {
		it('should track different action types', async () => {
			const actions = ['read', 'write', 'delete', 'export'] as const;

			for (const action of actions) {
				(global.fetch as any).mockResolvedValueOnce({ ok: true });
				await healthcareAPI.trackPHI('p1', 'Observation', action);
			}

			expect(global.fetch).toHaveBeenCalledTimes(4);
		});

		it('should include reason in audit log', async () => {
			const mockAuditTrail = {
				events: [
					{
						id: 'e1',
						patientId: 'p1',
						userId: 'u1',
						userName: 'Dr. Smith',
						userRole: 'Cardiologist',
						action: 'read' as const,
						resourceType: 'Medication',
						timestamp: '2024-01-15T10:00:00Z',
						reason: 'Medication review for appointment',
						ipAddress: '192.168.1.1',
						success: true
					}
				],
				total: 1,
				page: 1,
				limit: 50
			};

			(global.fetch as any).mockResolvedValueOnce({
				ok: true,
				json: async () => mockAuditTrail
			});

			const result = await healthcareAPI.getAuditTrail('p1');
			expect(result.events[0].reason).toBe('Medication review for appointment');
		});
	});

	describe('Consent Workflow', () => {
		it('should track consent expiration', async () => {
			const expirationDate = new Date();
			expirationDate.setFullYear(expirationDate.getFullYear() + 1);

			const mockConsent: ConsentStatus = {
				patientId: 'p1',
				resourceTypes: {
					Observation: {
						granted: true,
						grantedAt: '2024-01-15T00:00:00Z',
						expiresAt: expirationDate.toISOString()
					}
				},
				updatedAt: '2024-01-15T00:00:00Z'
			};

			(global.fetch as any).mockResolvedValueOnce({
				ok: true,
				json: async () => mockConsent
			});

			const result = await healthcareAPI.verifyConsent('p1');
			expect(result.resourceTypes.Observation.expiresAt).toBeDefined();
		});
	});

	describe('Patient Search', () => {
		it('should search by patient ID', async () => {
			const mockResponse: PatientListResponse = {
				patients: [
					{
						id: 'p-12345',
						firstName: 'Bob',
						lastName: 'Wilson',
						dateOfBirth: '1995-05-10',
						mrn: 'MRN-12345',
						status: 'active',
						createdAt: '2024-01-01T00:00:00Z',
						updatedAt: '2024-01-15T00:00:00Z',
						consentStatus: 'granted'
					}
				],
				total: 1,
				page: 1,
				limit: 20
			};

			(global.fetch as any).mockResolvedValueOnce({
				ok: true,
				json: async () => mockResponse
			});

			const result = await healthcareAPI.listPatients(1, 20, 'MRN-12345');
			expect(result.patients[0].mrn).toBe('MRN-12345');
		});

		it('should search by patient name', async () => {
			const mockResponse: PatientListResponse = {
				patients: [
					{
						id: 'p2',
						firstName: 'Carol',
						lastName: 'Davis',
						dateOfBirth: '1980-03-25',
						mrn: 'MRN-003',
						status: 'active',
						createdAt: '2024-01-01T00:00:00Z',
						updatedAt: '2024-01-15T00:00:00Z',
						consentStatus: 'granted'
					}
				],
				total: 1,
				page: 1,
				limit: 20
			};

			(global.fetch as any).mockResolvedValueOnce({
				ok: true,
				json: async () => mockResponse
			});

			const result = await healthcareAPI.listPatients(1, 20, 'Carol Davis');
			expect(result.patients[0].firstName).toBe('Carol');
		});
	});
});
