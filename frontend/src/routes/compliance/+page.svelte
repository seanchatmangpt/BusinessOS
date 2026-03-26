<script lang="ts">
	import { onMount } from 'svelte';
	import { fly, fade, slide } from 'svelte/transition';
	import { ChevronDown, TrendingUp, TrendingDown, AlertCircle, CheckCircle2, Clock } from 'lucide-svelte';
	import Button from '$lib/ui/button/Button.svelte';
	import Card from '$lib/ui/card/Card.svelte';
	import Badge from '$lib/ui/badge/Badge.svelte';
	import Tabs from '$lib/ui/tabs/Tabs.svelte';
	import TabsList from '$lib/ui/tabs/TabsList.svelte';
	import TabsTrigger from '$lib/ui/tabs/TabsTrigger.svelte';
	import TabsContent from '$lib/ui/tabs/TabsContent.svelte';
	import { complianceApi } from '$lib/api/compliance';
	import ComplianceScorecard from '$lib/components/ComplianceScorecard.svelte';
	import ControlsList from '$lib/components/ControlsList.svelte';

	// ── State ──────────────────────────────────────────────────────────────────────
	let selectedFramework = 'SOC2';
	let selectedSeverity = 'all';
	let isLoading = true;
	let expandedControls: Set<string> = new Set();
	let complianceStatus: any = null;
	let controls: any[] = [];
	let violations: any[] = [];
	let refreshTimer: NodeJS.Timeout | null = null;

	// ── Frameworks ─────────────────────────────────────────────────────────────────
	const frameworks = ['SOC2', 'GDPR', 'HIPAA', 'SOX'];
	const severityLevels = [
		{ value: 'all', label: 'All Severities' },
		{ value: 'critical', label: 'Critical' },
		{ value: 'high', label: 'High' },
		{ value: 'medium', label: 'Medium' },
		{ value: 'low', label: 'Low' }
	];

	// ── Computed ───────────────────────────────────────────────────────────────────
	$: filteredViolations = violations.filter((v) => {
		if (selectedSeverity === 'all') return true;
		return v.severity === selectedSeverity;
	});

	$: currentScore = complianceStatus?.[selectedFramework.toLowerCase()]?.score || 0;
	$: currentTrend = complianceStatus?.[selectedFramework.toLowerCase()]?.trend || 'stable';

	// ── Lifecycle ──────────────────────────────────────────────────────────────────
	onMount(async () => {
		await loadComplianceData();

		// Auto-refresh every 5 minutes
		refreshTimer = setInterval(loadComplianceData, 5 * 60 * 1000);

		return () => {
			if (refreshTimer) clearInterval(refreshTimer);
		};
	});

	// ── Functions ──────────────────────────────────────────────────────────────────
	async function loadComplianceData() {
		try {
			isLoading = true;

			// Load compliance status
			const statusResponse = await complianceApi.verifyCompliance();
			complianceStatus = statusResponse.data || {};

			// Load all controls
			const controlsResponse = await complianceApi.getControls();
			controls = controlsResponse.data || [];

			// Load violations
			const violationsResponse = await complianceApi.getViolations();
			violations = violationsResponse.data || [];

			isLoading = false;
		} catch (error) {
			console.error('Failed to load compliance data:', error);
			isLoading = false;
		}
	}

	function toggleControl(controlId: string) {
		if (expandedControls.has(controlId)) {
			expandedControls.delete(controlId);
		} else {
			expandedControls.add(controlId);
		}
		expandedControls = expandedControls;
	}

	async function exportReport() {
		try {
			const report = await complianceApi.getReport();
			const element = document.createElement('a');
			element.setAttribute('href', `data:text/json;charset=utf-8,${encodeURIComponent(JSON.stringify(report, null, 2))}`);
			element.setAttribute('download', `compliance-report-${Date.now()}.json`);
			element.style.display = 'none';
			document.body.appendChild(element);
			element.click();
			document.body.removeChild(element);
		} catch (error) {
			console.error('Failed to export report:', error);
		}
	}

	function getSeverityColor(severity: string): string {
		const colors: Record<string, string> = {
			critical: 'bg-red-100 text-red-800 border-red-300',
			high: 'bg-orange-100 text-orange-800 border-orange-300',
			medium: 'bg-yellow-100 text-yellow-800 border-yellow-300',
			low: 'bg-blue-100 text-blue-800 border-blue-300'
		};
		return colors[severity] || 'bg-gray-100 text-gray-800';
	}

	function getStatusIcon(status: string) {
		if (status === 'pass') return CheckCircle2;
		if (status === 'pending') return Clock;
		return AlertCircle;
	}

	function getControlsForFramework(framework: string): any[] {
		return controls.filter((c) => c.framework === framework.toLowerCase());
	}
</script>

<div class="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100 p-8">
	<div class="max-w-7xl mx-auto">
		<!-- Header -->
		<div class="mb-8" in:fly={{ y: -20, duration: 300 }}>
			<h1 class="text-4xl font-bold text-gray-900 mb-2">Compliance Dashboard</h1>
			<p class="text-gray-600">Monitor and verify compliance across major frameworks</p>
		</div>

		<!-- Scorecards Row -->
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-8" in:fly={{ y: 20, duration: 400, delay: 50 }}>
			{#each frameworks as framework (framework)}
				<button
					on:click={() => (selectedFramework = framework)}
					class="transition-all duration-300 transform hover:scale-105"
				>
					<ComplianceScorecard
						{framework}
						score={complianceStatus?.[framework.toLowerCase()]?.score || 0}
						trend={complianceStatus?.[framework.toLowerCase()]?.trend || 'stable'}
						lastUpdated={complianceStatus?.[framework.toLowerCase()]?.lastUpdated || new Date().toISOString()}
						isSelected={selectedFramework === framework}
					/>
				</button>
			{/each}
		</div>

		<!-- Main Content -->
		<div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
			<!-- Controls Section -->
			<div class="lg:col-span-2">
				<Card class="p-6">
					<div class="mb-6">
						<h2 class="text-2xl font-bold text-gray-900 mb-4">
							{selectedFramework} Controls
						</h2>
						<div class="flex items-center gap-2">
							<div class="flex-1 text-sm text-gray-600">
								{getControlsForFramework(selectedFramework).length} controls
							</div>
						</div>
					</div>

					{#if isLoading}
						<div class="flex justify-center py-12">
							<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
						</div>
					{:else}
						<Tabs defaultValue="controls" class="w-full">
							<TabsList class="grid w-full grid-cols-2">
								<TabsTrigger value="controls">All Controls</TabsTrigger>
								<TabsTrigger value="violations">Violations</TabsTrigger>
							</TabsList>

							<TabsContent value="controls" class="mt-6">
								{#if getControlsForFramework(selectedFramework).length === 0}
									<div class="text-center py-8 text-gray-500">No controls found for this framework</div>
								{:else}
									<ControlsList
										controls={getControlsForFramework(selectedFramework)}
										{expandedControls}
										on:toggle={(e) => toggleControl(e.detail)}
									/>
								{/if}
							</TabsContent>

							<TabsContent value="violations" class="mt-6">
								<div class="space-y-3">
									{#if filteredViolations.length === 0}
										<div class="text-center py-8 text-gray-500">
											No violations found {selectedSeverity !== 'all' ? `with severity ${selectedSeverity}` : ''}
										</div>
									{:else}
										{#each filteredViolations as violation (violation.id)}
											<div
												class="border border-gray-200 rounded-lg p-4 hover:bg-gray-50 transition-colors"
												in:slide={{ duration: 300 }}
											>
												<div class="flex items-start justify-between mb-2">
													<div class="flex items-center gap-3 flex-1">
														<Badge class={getSeverityColor(violation.severity)}>
															{violation.severity.toUpperCase()}
														</Badge>
														<span class="font-medium text-gray-900">{violation.controlId}</span>
													</div>
													<span class="text-sm text-gray-500">{violation.framework?.toUpperCase()}</span>
												</div>
												<p class="text-sm text-gray-600 mb-3">{violation.description}</p>
												{#if violation.remediation}
													<div class="bg-blue-50 border-l-4 border-blue-300 p-3 rounded">
														<p class="text-sm font-medium text-blue-900">Remediation:</p>
														<p class="text-sm text-blue-800 mt-1">{violation.remediation}</p>
													</div>
												{/if}
											</div>
										{/each}
									{/if}
								</div>
							</TabsContent>
						</Tabs>
					{/if}
				</Card>
			</div>

			<!-- Sidebar -->
			<div class="space-y-6">
				<!-- Severity Filter -->
				<Card class="p-6">
					<h3 class="text-lg font-bold text-gray-900 mb-4">Filter by Severity</h3>
					<div class="space-y-2">
						{#each severityLevels as level (level.value)}
							<button
								on:click={() => (selectedSeverity = level.value)}
								class={`w-full text-left px-4 py-2 rounded-lg transition-colors ${
									selectedSeverity === level.value
										? 'bg-blue-600 text-white'
										: 'bg-gray-100 text-gray-700 hover:bg-gray-200'
								}`}
							>
								{level.label}
								<span class="float-right text-sm opacity-75">
									{level.value === 'all'
										? violations.length
										: violations.filter((v) => v.severity === level.value).length}
								</span>
							</button>
						{/each}
					</div>
				</Card>

				<!-- Quick Stats -->
				<Card class="p-6 bg-gradient-to-br from-green-50 to-emerald-50 border-green-200">
					<h3 class="text-lg font-bold text-gray-900 mb-4">Overall Status</h3>
					<div class="space-y-3">
						<div>
							<p class="text-sm text-gray-600">Total Controls</p>
							<p class="text-3xl font-bold text-green-700">{controls.length}</p>
						</div>
						<div>
							<p class="text-sm text-gray-600">Violations</p>
							<p class="text-3xl font-bold text-red-600">{violations.length}</p>
						</div>
						<div>
							<p class="text-sm text-gray-600">Passing</p>
							<p class="text-3xl font-bold text-blue-600">
								{controls.filter((c) => c.status === 'pass').length}
							</p>
						</div>
					</div>
				</Card>

				<!-- Actions -->
				<Card class="p-6">
					<h3 class="text-lg font-bold text-gray-900 mb-4">Actions</h3>
					<div class="space-y-2">
						<Button on:click={loadComplianceData} class="w-full" variant="outline">Refresh Data</Button>
						<Button on:click={exportReport} class="w-full">Export Report</Button>
					</div>
				</Card>
			</div>
		</div>
	</div>
</div>

<style>
	:global(body) {
		@apply bg-gradient-to-br from-slate-50 to-slate-100;
	}
</style>
