// Healthcare API client for HIPAA-compliant PHI management
// All endpoints respect PII masking, consent verification, and audit logging

export type PatientStatus = 'active' | 'inactive' | 'discharged';

export interface Patient {
	id: string;
	firstName: string;
	lastName: string;
	dateOfBirth: string;
	mrn: string; // Medical Record Number
	status: PatientStatus;
	createdAt: string;
	updatedAt: string;
	consentStatus: 'granted' | 'denied' | 'pending';
}

export interface AccessEvent {
	id: string;
	patientId: string;
	userId: string;
	userName: string;
	userRole: string; // e.g., "Cardiologist", "Nurse", "Administrator"
	action: 'read' | 'write' | 'delete' | 'export' | 'access_denied';
	resourceType: string; // e.g., "Observation", "Medication", "Condition"
	timestamp: string;
	reason?: string; // Why was this access made
	ipAddress?: string;
	success: boolean;
}

export interface ConsentStatus {
	patientId: string;
	resourceTypes: {
		[key: string]: {
			granted: boolean;
			grantedAt?: string;
			expiresAt?: string;
		};
	};
	updatedAt: string;
}

export interface HIPAACompliance {
	accessControl: {
		passed: boolean;
		details: string;
	};
	auditLogging: {
		passed: boolean;
		details: string;
	};
	encryption: {
		passed: boolean;
		details: string;
	};
	integrity: {
		passed: boolean;
		details: string;
	};
	score: number; // 0-100
	lastChecked: string;
}

export interface PatientListResponse {
	patients: Patient[];
	total: number;
	page: number;
	limit: number;
}

export interface AuditTrailResponse {
	events: AccessEvent[];
	total: number;
	page: number;
	limit: number;
}

class HealthcareAPIClient {
	private baseUrl: string;

	constructor() {
		this.baseUrl = '/api/healthcare';
	}

	/**
	 * List all patients with optional search
	 */
	async listPatients(
		page: number = 1,
		limit: number = 20,
		search?: string
	): Promise<PatientListResponse> {
		const params = new URLSearchParams({
			page: page.toString(),
			limit: limit.toString()
		});

		if (search) {
			params.append('search', search);
		}

		const response = await fetch(`${this.baseUrl}/patients?${params}`);
		if (!response.ok) {
			throw new Error(`Failed to list patients: ${response.statusText}`);
		}
		return response.json();
	}

	/**
	 * Get patient details by ID
	 * Returns masked PII for non-authorized users
	 */
	async getPatient(patientId: string): Promise<Patient> {
		const response = await fetch(`${this.baseUrl}/patients/${patientId}`);
		if (!response.ok) {
			throw new Error(`Failed to get patient: ${response.statusText}`);
		}
		return response.json();
	}

	/**
	 * Track PHI access (called automatically by components)
	 */
	async trackPHI(
		patientId: string,
		resourceType: string,
		action: 'read' | 'write' | 'delete' | 'export'
	): Promise<void> {
		const response = await fetch(`${this.baseUrl}/patients/${patientId}/track`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ resourceType, action })
		});

		if (!response.ok) {
			console.error('Failed to track PHI access');
		}
	}

	/**
	 * Get audit trail for a patient
	 * Last 50 events by default
	 */
	async getAuditTrail(
		patientId: string,
		page: number = 1,
		limit: number = 50
	): Promise<AuditTrailResponse> {
		const params = new URLSearchParams({
			page: page.toString(),
			limit: limit.toString()
		});

		const response = await fetch(`${this.baseUrl}/patients/${patientId}/audit?${params}`);
		if (!response.ok) {
			throw new Error(`Failed to get audit trail: ${response.statusText}`);
		}
		return response.json();
	}

	/**
	 * Verify consent status for specific resources
	 */
	async verifyConsent(patientId: string): Promise<ConsentStatus> {
		const response = await fetch(`${this.baseUrl}/patients/${patientId}/consent`);
		if (!response.ok) {
			throw new Error(`Failed to verify consent: ${response.statusText}`);
		}
		return response.json();
	}

	/**
	 * Check HIPAA compliance status
	 */
	async verifyHIPAA(patientId: string): Promise<HIPAACompliance> {
		const response = await fetch(`${this.baseUrl}/patients/${patientId}/compliance`);
		if (!response.ok) {
			throw new Error(`Failed to verify HIPAA compliance: ${response.statusText}`);
		}
		return response.json();
	}

	/**
	 * Delete patient PHI (GDPR "right to be forgotten")
	 * Requires confirmation and has a grace period
	 */
	async deletePHI(patientId: string, reason: string): Promise<{ success: boolean; graceUntil: string }> {
		const response = await fetch(`${this.baseUrl}/patients/${patientId}/delete`, {
			method: 'DELETE',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ reason })
		});

		if (!response.ok) {
			throw new Error(`Failed to delete PHI: ${response.statusText}`);
		}
		return response.json();
	}

	/**
	 * Mask PII for display (helper function)
	 * Used by components to mask names, MRN, DOB in audit logs
	 */
	maskPII(fullName: string, role?: string): string {
		const parts = fullName.split(' ');
		if (parts.length === 0) return '***';

		const lastName = parts[parts.length - 1];
		const firstInitial = parts[0].charAt(0);
		const masked = `${firstInitial}. ${lastName}`;

		return role ? `${masked} (${role})` : masked;
	}
}

export const healthcareAPI = new HealthcareAPIClient();
