<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { healthcareAPI, type Patient, type AuditTrailResponse, type ConsentStatus, type HIPAACompliance } from '$lib/api/healthcare';
	import PatientAccessLog from '$lib/components/PatientAccessLog.svelte';
	import HIPAAComplianceCard from '$lib/components/HIPAAComplianceCard.svelte';

	// Extract patient ID from route
	const patientId = $derived($page.params.patientId);

	// State
	let patient: Patient | null = $state(null);
	let auditTrail: AuditTrailResponse | null = $state(null);
	let consentStatus: ConsentStatus | null = $state(null);
	let hipaaCompliance: HIPAACompliance | null = $state(null);
	let loading = $state(true);
	let error = $state('');
	let activeTab = $state<'overview' | 'resources' | 'audit' | 'compliance'>('overview');
	let showDeleteConfirm = $state(false);
	let deleteReason = $state('');
	let deleting = $state(false);

	// Load patient data on mount
	onMount(async () => {
		try {
			const [p, audit, consent, compliance] = await Promise.all([
				healthcareAPI.getPatient(patientId),
				healthcareAPI.getAuditTrail(patientId),
				healthcareAPI.verifyConsent(patientId),
				healthcareAPI.verifyHIPAA(patientId)
			]);

			patient = p;
			auditTrail = audit;
			consentStatus = consent;
			hipaaCompliance = compliance;

			// Track that this patient record was accessed
			await healthcareAPI.trackPHI(patientId, 'Patient', 'read');
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load patient data';
		} finally {
			loading = false;
		}
	});

	async function handleDeletePHI() {
		if (!deleteReason.trim()) {
			error = 'Please provide a reason for deletion';
			return;
		}

		deleting = true;
		error = '';

		try {
			const result = await healthcareAPI.deletePHI(patientId, deleteReason);
			// Show confirmation
			error = `Deletion initiated. PHI will be permanently deleted by ${new Date(result.graceUntil).toLocaleString()}`;
			showDeleteConfirm = false;
			deleteReason = '';
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to delete PHI';
		} finally {
			deleting = false;
		}
	}

	function goBack() {
		goto('/healthcare');
	}

	function formatDate(dateString: string): string {
		return new Date(dateString).toLocaleDateString();
	}
</script>

<div class="min-h-screen bg-gray-50 p-6">
	<div class="mx-auto max-w-6xl">
		<!-- Header with back button -->
		<div class="mb-6 flex items-center gap-4">
			<button
				onclick={goBack}
				class="rounded-lg border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-100"
			>
				← Back to Patients
			</button>
		</div>

		{#if loading}
			<div class="flex items-center justify-center p-12">
				<div class="text-center">
					<div class="mb-4 h-12 w-12 animate-spin rounded-full border-4 border-gray-300 border-t-blue-600 mx-auto"></div>
					<p class="text-gray-600">Loading patient details...</p>
				</div>
			</div>
		{:else if error && !patient}
			<div class="rounded-lg border border-red-200 bg-red-50 p-4">
				<p class="text-red-800">{error}</p>
			</div>
		{:else if patient}
			<!-- Patient Header -->
			<div class="mb-8 rounded-lg border border-gray-200 bg-white p-6 shadow-sm">
				<div class="flex items-start justify-between">
					<div>
						<h1 class="text-2xl font-bold text-gray-900">{patient.firstName} {patient.lastName}</h1>
						<p class="mt-1 text-sm text-gray-600">MRN: <span class="font-mono font-semibold">{patient.mrn}</span></p>
						<p class="mt-1 text-sm text-gray-600">DOB: {formatDate(patient.dateOfBirth)}</p>
					</div>
					<div class="text-right">
						<span class="inline-flex items-center rounded-full px-3 py-1 text-sm font-medium {
							patient.status === 'active'
								? 'bg-green-100 text-green-800'
								: 'bg-gray-100 text-gray-800'
						}">
							{patient.status.charAt(0).toUpperCase() + patient.status.slice(1)}
						</span>
					</div>
				</div>
			</div>

			<!-- Tab Navigation -->
			<div class="mb-6 flex gap-2 border-b border-gray-200">
				{#each [
					{ id: 'overview', label: 'Overview' },
					{ id: 'resources', label: 'Resources' },
					{ id: 'audit', label: 'Audit Trail' },
					{ id: 'compliance', label: 'Compliance' }
				] as tab}
					<button
						onclick={() => (activeTab = tab.id)}
						class={`px-4 py-2 font-medium text-sm border-b-2 transition-colors ${
							activeTab === tab.id
								? 'border-blue-600 text-blue-600'
								: 'border-transparent text-gray-600 hover:text-gray-900'
						}`}
					>
						{tab.label}
					</button>
				{/each}
			</div>

			<!-- Tab Content -->
			<div class="space-y-6">
				{#if activeTab === 'overview'}
					<!-- Overview Tab -->
					<div class="rounded-lg border border-gray-200 bg-white p-6 shadow-sm">
						<h2 class="mb-4 text-lg font-semibold text-gray-900">Patient Information</h2>
						<div class="grid grid-cols-2 gap-6">
							<div>
								<label class="text-sm font-semibold text-gray-700">First Name</label>
								<p class="mt-1 text-gray-900">{patient.firstName}</p>
							</div>
							<div>
								<label class="text-sm font-semibold text-gray-700">Last Name</label>
								<p class="mt-1 text-gray-900">{patient.lastName}</p>
							</div>
							<div>
								<label class="text-sm font-semibold text-gray-700">Date of Birth</label>
								<p class="mt-1 text-gray-900">{formatDate(patient.dateOfBirth)}</p>
							</div>
							<div>
								<label class="text-sm font-semibold text-gray-700">Status</label>
								<p class="mt-1 text-gray-900">{patient.status}</p>
							</div>
							<div>
								<label class="text-sm font-semibold text-gray-700">Record Created</label>
								<p class="mt-1 text-gray-900">{new Date(patient.createdAt).toLocaleString()}</p>
							</div>
							<div>
								<label class="text-sm font-semibold text-gray-700">Last Updated</label>
								<p class="mt-1 text-gray-900">{new Date(patient.updatedAt).toLocaleString()}</p>
							</div>
						</div>
					</div>

					<!-- Consent Status -->
					{#if consentStatus}
						<div class="rounded-lg border border-gray-200 bg-white p-6 shadow-sm">
							<h2 class="mb-4 text-lg font-semibold text-gray-900">Consent Status</h2>
							<div class="space-y-3">
								{#each Object.entries(consentStatus.resourceTypes) as [resourceType, consent]}
									<div class="flex items-center justify-between rounded-lg bg-gray-50 p-4">
										<span class="font-medium text-gray-900">{resourceType}</span>
										<span class={`inline-flex items-center rounded-full px-3 py-1 text-xs font-medium ${
											consent.granted
												? 'bg-green-100 text-green-800'
												: 'bg-red-100 text-red-800'
										}`}>
											{consent.granted ? 'Granted' : 'Denied'}
										</span>
									</div>
								{/each}
							</div>
						</div>
					{/if}
				{/if}

				{#if activeTab === 'resources'}
					<!-- Resources Tab -->
					<div class="rounded-lg border border-gray-200 bg-white p-6 shadow-sm">
						<h2 class="mb-4 text-lg font-semibold text-gray-900">FHIR Resources</h2>
						<p class="mb-4 text-sm text-gray-600">Patient clinical resources with consent verification:</p>
						<div class="space-y-3">
							{#each ['Observation', 'Medication', 'Condition', 'Procedure', 'Appointment', 'CarePlan'] as resourceType}
								<div class="flex items-center justify-between rounded-lg bg-gray-50 p-4">
									<div>
										<p class="font-medium text-gray-900">{resourceType}</p>
										<p class="text-xs text-gray-600">FHIR resource type</p>
									</div>
									{#if consentStatus?.resourceTypes[resourceType]?.granted}
										<span class="inline-flex items-center rounded-full bg-green-100 px-3 py-1 text-xs font-medium text-green-800">
											✓ Accessible
										</span>
									{:else}
										<span class="inline-flex items-center rounded-full bg-gray-100 px-3 py-1 text-xs font-medium text-gray-800">
											✗ No Consent
										</span>
									{/if}
								</div>
							{/each}
						</div>
					</div>
				{/if}

				{#if activeTab === 'audit' && auditTrail}
					<!-- Audit Trail Tab -->
					<div class="rounded-lg border border-gray-200 bg-white p-6 shadow-sm">
						<PatientAccessLog events={auditTrail.events} />
					</div>
				{/if}

				{#if activeTab === 'compliance' && hipaaCompliance}
					<!-- Compliance Tab -->
					<div class="space-y-6">
						<HIPAAComplianceCard compliance={hipaaCompliance} />

						<!-- GDPR Delete Section -->
						<div class="rounded-lg border border-red-200 bg-red-50 p-6">
							<h2 class="mb-2 text-lg font-semibold text-red-900">Right to be Forgotten (GDPR)</h2>
							<p class="mb-4 text-sm text-red-800">
								You can request permanent deletion of this patient's PHI. This action cannot be undone.
							</p>

							{#if !showDeleteConfirm}
								<button
									onclick={() => (showDeleteConfirm = true)}
									class="rounded-lg bg-red-600 px-4 py-2 font-medium text-white hover:bg-red-700"
								>
									Delete PHI
								</button>
							{:else}
								<div class="space-y-4 rounded-lg bg-white p-4">
									<div>
										<label class="block text-sm font-semibold text-gray-700">Reason for Deletion</label>
										<textarea
											bind:value={deleteReason}
											placeholder="Explain why you're requesting deletion..."
											class="mt-2 w-full rounded-lg border border-gray-300 px-4 py-2 text-gray-900 focus:border-red-500 focus:outline-none"
											rows="4"
										></textarea>
									</div>

									{#if error}
										<div class="rounded-lg bg-yellow-50 p-3">
											<p class="text-sm text-yellow-800">{error}</p>
										</div>
									{/if}

									<div class="flex gap-3">
										<button
											onclick={handleDeletePHI}
											disabled={deleting || !deleteReason.trim()}
											class="rounded-lg bg-red-600 px-4 py-2 font-medium text-white hover:bg-red-700 disabled:bg-gray-400"
										>
											{deleting ? 'Processing...' : 'Confirm Deletion'}
										</button>
										<button
											onclick={() => {
												showDeleteConfirm = false;
												deleteReason = '';
											}}
											class="rounded-lg border border-gray-300 px-4 py-2 font-medium text-gray-700 hover:bg-gray-50"
										>
											Cancel
										</button>
									</div>
								</div>
							{/if}
						</div>
					</div>
				{/if}
			</div>
		{/if}
	</div>
</div>
