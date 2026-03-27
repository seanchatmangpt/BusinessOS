<script lang="ts">
	import type { HIPAACompliance } from '$lib/api/healthcare';

	interface Props {
		compliance: HIPAACompliance;
	}

	let { compliance }: Props = $props();

	// Get overall status color
	function getScoreColor(score: number): string {
		if (score >= 90) return 'text-green-600';
		if (score >= 70) return 'text-yellow-600';
		return 'text-red-600';
	}

	// Get check mark or X
	function getCheckIcon(passed: boolean): string {
		return passed ? '✓' : '✗';
	}

	// Get check color
	function getCheckColor(passed: boolean): string {
		return passed ? 'text-green-600' : 'text-red-600';
	}

	function formatLastChecked(timestamp: string): string {
		const date = new Date(timestamp);
		return date.toLocaleString();
	}
</script>

<div class="rounded-lg border border-gray-200 bg-white p-6 shadow-sm">
	<!-- Header with score -->
	<div class="mb-6 flex items-center justify-between">
		<h3 class="text-lg font-semibold text-gray-900">HIPAA Compliance Status</h3>
		<div class="text-center">
			<div class={`text-4xl font-bold ${getScoreColor(compliance.score)}`}>
				{compliance.score}%
			</div>
			<p class="text-xs text-gray-600">Compliance Score</p>
		</div>
	</div>

	<!-- Compliance checks -->
	<div class="space-y-4">
		<!-- Access Control -->
		<div class="flex items-start space-x-4 rounded-lg bg-gray-50 p-4">
			<div class={`mt-1 text-2xl font-bold ${getCheckColor(compliance.accessControl.passed)}`}>
				{getCheckIcon(compliance.accessControl.passed)}
			</div>
			<div class="flex-1">
				<h4 class="font-semibold text-gray-900">Access Control</h4>
				<p class="text-sm text-gray-600">{compliance.accessControl.details}</p>
			</div>
			<span class={`flex items-center text-xs font-semibold px-3 py-1 rounded-full ${
				compliance.accessControl.passed
					? 'bg-green-100 text-green-800'
					: 'bg-red-100 text-red-800'
			}`}>
				{compliance.accessControl.passed ? 'PASS' : 'FAIL'}
			</span>
		</div>

		<!-- Audit Logging -->
		<div class="flex items-start space-x-4 rounded-lg bg-gray-50 p-4">
			<div class={`mt-1 text-2xl font-bold ${getCheckColor(compliance.auditLogging.passed)}`}>
				{getCheckIcon(compliance.auditLogging.passed)}
			</div>
			<div class="flex-1">
				<h4 class="font-semibold text-gray-900">Audit Logging</h4>
				<p class="text-sm text-gray-600">{compliance.auditLogging.details}</p>
			</div>
			<span class={`flex items-center text-xs font-semibold px-3 py-1 rounded-full ${
				compliance.auditLogging.passed
					? 'bg-green-100 text-green-800'
					: 'bg-red-100 text-red-800'
			}`}>
				{compliance.auditLogging.passed ? 'PASS' : 'FAIL'}
			</span>
		</div>

		<!-- Encryption -->
		<div class="flex items-start space-x-4 rounded-lg bg-gray-50 p-4">
			<div class={`mt-1 text-2xl font-bold ${getCheckColor(compliance.encryption.passed)}`}>
				{getCheckIcon(compliance.encryption.passed)}
			</div>
			<div class="flex-1">
				<h4 class="font-semibold text-gray-900">Encryption at Rest & Transit</h4>
				<p class="text-sm text-gray-600">{compliance.encryption.details}</p>
			</div>
			<span class={`flex items-center text-xs font-semibold px-3 py-1 rounded-full ${
				compliance.encryption.passed
					? 'bg-green-100 text-green-800'
					: 'bg-red-100 text-red-800'
			}`}>
				{compliance.encryption.passed ? 'PASS' : 'FAIL'}
			</span>
		</div>

		<!-- Integrity -->
		<div class="flex items-start space-x-4 rounded-lg bg-gray-50 p-4">
			<div class={`mt-1 text-2xl font-bold ${getCheckColor(compliance.integrity.passed)}`}>
				{getCheckIcon(compliance.integrity.passed)}
			</div>
			<div class="flex-1">
				<h4 class="font-semibold text-gray-900">Data Integrity & Non-Repudiation</h4>
				<p class="text-sm text-gray-600">{compliance.integrity.details}</p>
			</div>
			<span class={`flex items-center text-xs font-semibold px-3 py-1 rounded-full ${
				compliance.integrity.passed
					? 'bg-green-100 text-green-800'
					: 'bg-red-100 text-red-800'
			}`}>
				{compliance.integrity.passed ? 'PASS' : 'FAIL'}
			</span>
		</div>
	</div>

	<!-- Footer with last check time -->
	<div class="mt-6 border-t border-gray-200 pt-4">
		<p class="text-xs text-gray-600">Last compliance check: {formatLastChecked(compliance.lastChecked)}</p>
	</div>
</div>

<style>
	/* Ensure accessibility with proper contrast */
	:global(.hipaa-card) {
		@apply border-l-4 border-l-blue-500;
	}
</style>
