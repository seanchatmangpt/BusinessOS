<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { healthcareAPI, type PatientListResponse, type Patient } from '$lib/api/healthcare';
	import HIPAAComplianceCard from '$lib/components/HIPAAComplianceCard.svelte';

	// State
	let patients: Patient[] = $state([]);
	let loading = $state(false);
	let error = $state('');
	let searchQuery = $state('');
	let currentPage = $state(1);
	let totalPatients = $state(0);
	const pageSize = 20;

	// Load patients on mount and search
	onMount(async () => {
		await loadPatients();
	});

	async function loadPatients() {
		loading = true;
		error = '';

		try {
			const result: PatientListResponse = await healthcareAPI.listPatients(
				currentPage,
				pageSize,
				searchQuery
			);
			patients = result.patients;
			totalPatients = result.total;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load patients';
		} finally {
			loading = false;
		}
	}

	async function handleSearch() {
		currentPage = 1;
		await loadPatients();
	}

	function goToPatientDetail(patientId: string) {
		goto(`/healthcare/${patientId}`);
	}

	function getConsentStatusColor(status: string): string {
		switch (status) {
			case 'granted':
				return 'bg-green-100 text-green-800';
			case 'denied':
				return 'bg-red-100 text-red-800';
			default:
				return 'bg-yellow-100 text-yellow-800';
		}
	}

	function getStatusColor(status: string): string {
		switch (status) {
			case 'active':
				return 'text-green-600';
			case 'discharged':
				return 'text-gray-600';
			default:
				return 'text-gray-400';
		}
	}

	const totalPages = $derived(Math.ceil(totalPatients / pageSize));
</script>

<div class="min-h-screen bg-gray-50 p-6">
	<div class="mx-auto max-w-6xl">
		<!-- Header -->
		<div class="mb-8">
			<h1 class="text-3xl font-bold text-gray-900">Healthcare PHI Dashboard</h1>
			<p class="mt-2 text-gray-600">HIPAA-compliant patient information management with full audit trail</p>
		</div>

		<!-- Search Section -->
		<div class="mb-8 rounded-lg border border-gray-200 bg-white p-6 shadow-sm">
			<h2 class="mb-4 text-lg font-semibold text-gray-900">Search Patients</h2>

			<div class="flex gap-4">
				<input
					type="text"
					placeholder="Search by Patient ID or name..."
					bind:value={searchQuery}
					on:keydown={(e) => e.key === 'Enter' && handleSearch()}
					class="flex-1 rounded-lg border border-gray-300 px-4 py-2 text-gray-900 placeholder-gray-500 focus:border-blue-500 focus:outline-none"
				/>
				<button
					onclick={handleSearch}
					disabled={loading}
					class="rounded-lg bg-blue-600 px-6 py-2 font-medium text-white hover:bg-blue-700 disabled:bg-gray-400"
				>
					{loading ? 'Searching...' : 'Search'}
				</button>
			</div>
		</div>

		<!-- Error Message -->
		{#if error}
			<div class="mb-6 rounded-lg border border-red-200 bg-red-50 p-4">
				<p class="text-red-800">{error}</p>
			</div>
		{/if}

		<!-- Patients Table -->
		<div class="rounded-lg border border-gray-200 bg-white shadow-sm">
			{#if loading}
				<div class="flex items-center justify-center p-12">
					<div class="text-center">
						<div class="mb-4 h-12 w-12 animate-spin rounded-full border-4 border-gray-300 border-t-blue-600 mx-auto"></div>
						<p class="text-gray-600">Loading patients...</p>
					</div>
				</div>
			{:else if patients.length === 0}
				<div class="flex items-center justify-center p-12">
					<div class="text-center">
						<p class="text-gray-600">No patients found</p>
					</div>
				</div>
			{:else}
				<div class="overflow-x-auto">
					<table class="w-full divide-y divide-gray-200">
						<thead class="bg-gray-50">
							<tr>
								<th class="px-6 py-3 text-left text-sm font-semibold text-gray-700">Patient Name</th>
								<th class="px-6 py-3 text-left text-sm font-semibold text-gray-700">MRN</th>
								<th class="px-6 py-3 text-left text-sm font-semibold text-gray-700">Status</th>
								<th class="px-6 py-3 text-left text-sm font-semibold text-gray-700">Consent</th>
								<th class="px-6 py-3 text-left text-sm font-semibold text-gray-700">Added</th>
								<th class="px-6 py-3 text-left text-sm font-semibold text-gray-700">Actions</th>
							</tr>
						</thead>
						<tbody class="divide-y divide-gray-200">
							{#each patients as patient (patient.id)}
								<tr class="hover:bg-gray-50">
									<td class="px-6 py-4 text-sm text-gray-900">
										{patient.firstName} {patient.lastName}
									</td>
									<td class="px-6 py-4 text-sm font-mono text-gray-600">
										{patient.mrn}
									</td>
									<td class="px-6 py-4 text-sm">
										<span class={`font-medium ${getStatusColor(patient.status)}`}>
											{patient.status.charAt(0).toUpperCase() + patient.status.slice(1)}
										</span>
									</td>
									<td class="px-6 py-4 text-sm">
										<span
											class={`inline-flex items-center rounded-full px-3 py-1 text-xs font-medium ${getConsentStatusColor(
												patient.consentStatus
											)}`}
										>
											{patient.consentStatus.charAt(0).toUpperCase() + patient.consentStatus.slice(1)}
										</span>
									</td>
									<td class="px-6 py-4 text-sm text-gray-600">
										{new Date(patient.createdAt).toLocaleDateString()}
									</td>
									<td class="px-6 py-4 text-sm">
										<button
											onclick={() => goToPatientDetail(patient.id)}
											class="rounded-lg bg-blue-100 px-3 py-1 text-blue-700 hover:bg-blue-200"
										>
											View Details
										</button>
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>

				<!-- Pagination -->
				{#if totalPages > 1}
					<div class="flex items-center justify-between border-t border-gray-200 px-6 py-4">
						<div class="text-sm text-gray-600">
							Showing {(currentPage - 1) * pageSize + 1} to {Math.min(currentPage * pageSize, totalPatients)} of {totalPatients} patients
						</div>

						<div class="flex gap-2">
							<button
								onclick={() => {
									if (currentPage > 1) {
										currentPage--;
										loadPatients();
									}
								}}
								disabled={currentPage === 1}
								class="rounded-lg border border-gray-300 px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:bg-gray-50 disabled:text-gray-400"
							>
								Previous
							</button>

							{#each Array.from({ length: Math.min(5, totalPages) }, (_, i) => i + 1) as page}
								<button
									onclick={() => {
										currentPage = page;
										loadPatients();
									}}
									class={`rounded-lg px-3 py-2 text-sm font-medium ${
										currentPage === page
											? 'bg-blue-600 text-white'
											: 'border border-gray-300 text-gray-700 hover:bg-gray-50'
									}`}
								>
									{page}
								</button>
							{/each}

							<button
								onclick={() => {
									if (currentPage < totalPages) {
										currentPage++;
										loadPatients();
									}
								}}
								disabled={currentPage === totalPages}
								class="rounded-lg border border-gray-300 px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:bg-gray-50 disabled:text-gray-400"
							>
								Next
							</button>
						</div>
					</div>
				{/if}
			{/if}
		</div>

		<!-- HIPAA Information Card -->
		<div class="mt-8 rounded-lg border border-blue-200 bg-blue-50 p-6">
			<h3 class="mb-2 font-semibold text-blue-900">HIPAA Compliance Notice</h3>
			<p class="text-sm text-blue-800">
				All patient information is protected under HIPAA regulations. Every access to PHI is logged and audited.
				Unauthorized access is prohibited and will be reported to compliance authorities.
			</p>
		</div>
	</div>
</div>

<style>
	:global(body) {
		@apply bg-gray-50;
	}
</style>
