<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { team } from '$lib/stores/team';
	import { goto } from '$app/navigation';
	import type { TeamMemberDetailResponse } from '$lib/api';
	import StatusBadge from '$lib/components/team/StatusBadge.svelte';
	import CapacityBar from '$lib/components/team/CapacityBar.svelte';
	import { ArrowLeft, Mail, Calendar, Briefcase, CheckSquare, Clock } from 'lucide-svelte';

	let member = $state<TeamMemberDetailResponse | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);

	const memberId = $derived($page.params.memberId);

	onMount(async () => {
		await loadMember();
	});

	async function loadMember() {
		loading = true;
		error = null;
		try {
			const data = await team.loadMember(memberId);
			if (data) {
				member = data;
			} else {
				error = 'Member not found';
			}
		} catch (err) {
			console.error('Failed to load member:', err);
			error = 'Failed to load member details';
		} finally {
			loading = false;
		}
	}

	function getInitials(name: string) {
		return name
			.split(' ')
			.map((n) => n.charAt(0))
			.join('')
			.toUpperCase()
			.slice(0, 2);
	}

	function formatDate(dateStr: string | undefined) {
		if (!dateStr) return 'N/A';
		return new Date(dateStr).toLocaleDateString('en-US', {
			month: 'long',
			day: 'numeric',
			year: 'numeric'
		});
	}
</script>

<div class="min-h-full bg-gray-50 dark:bg-gray-900">
	<!-- Header -->
	<div class="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">
		<div class="max-w-5xl mx-auto px-6 py-4">
			<button
				onclick={() => goto('/team')}
				class="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white transition-colors mb-4"
			>
				<ArrowLeft class="w-4 h-4" />
				Back to Team
			</button>

			{#if loading}
				<div class="animate-pulse">
					<div class="flex items-center gap-6">
						<div class="w-24 h-24 rounded-full bg-gray-200 dark:bg-gray-700"></div>
						<div class="space-y-3">
							<div class="h-8 w-48 bg-gray-200 dark:bg-gray-700 rounded"></div>
							<div class="h-4 w-32 bg-gray-200 dark:bg-gray-700 rounded"></div>
						</div>
					</div>
				</div>
			{:else if error}
				<div class="text-center py-12">
					<p class="text-red-600 dark:text-red-400 mb-4">{error}</p>
					<button
						onclick={loadMember}
						class="px-4 py-2 bg-gray-900 dark:bg-gray-100 text-white dark:text-gray-900 rounded-lg text-sm font-medium hover:bg-gray-800 dark:hover:bg-gray-200 transition-colors"
					>
						Try Again
					</button>
				</div>
			{:else if member}
				<div class="flex items-start gap-6">
					<!-- Avatar -->
					{#if member.avatar_url}
						<img
							src={member.avatar_url}
							alt={member.name}
							class="w-24 h-24 rounded-full object-cover border-4 border-white dark:border-gray-700 shadow-lg"
						/>
					{:else}
						<div class="w-24 h-24 rounded-full bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center text-white text-2xl font-bold border-4 border-white dark:border-gray-700 shadow-lg">
							{getInitials(member.name)}
						</div>
					{/if}

					<!-- Info -->
					<div class="flex-1">
						<div class="flex items-center gap-3 mb-1">
							<h1 class="text-2xl font-bold text-gray-900 dark:text-white">{member.name}</h1>
							<StatusBadge status={member.status as 'available' | 'busy' | 'overloaded' | 'ooo'} />
						</div>
						<p class="text-lg text-gray-600 dark:text-gray-400">{member.role}</p>
						<div class="flex items-center gap-4 mt-3 text-sm text-gray-500 dark:text-gray-400">
							<span class="flex items-center gap-1.5">
								<Mail class="w-4 h-4" />
								{member.email}
							</span>
							<span class="flex items-center gap-1.5">
								<Calendar class="w-4 h-4" />
								Joined {formatDate(member.joined_at)}
							</span>
						</div>
					</div>
				</div>
			{/if}
		</div>
	</div>

	{#if member && !loading}
		<div class="max-w-5xl mx-auto px-6 py-8">
			<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
				<!-- Stats Cards -->
				<div class="lg:col-span-2 grid grid-cols-2 gap-4">
					<div class="bg-white dark:bg-gray-800 rounded-xl p-5 border border-gray-200 dark:border-gray-700">
						<div class="flex items-center gap-3 mb-3">
							<div class="w-10 h-10 rounded-lg bg-blue-100 dark:bg-blue-900/30 flex items-center justify-center">
								<Briefcase class="w-5 h-5 text-blue-600 dark:text-blue-400" />
							</div>
							<span class="text-sm font-medium text-gray-600 dark:text-gray-400">Active Projects</span>
						</div>
						<p class="text-3xl font-bold text-gray-900 dark:text-white">{member.active_projects}</p>
					</div>

					<div class="bg-white dark:bg-gray-800 rounded-xl p-5 border border-gray-200 dark:border-gray-700">
						<div class="flex items-center gap-3 mb-3">
							<div class="w-10 h-10 rounded-lg bg-green-100 dark:bg-green-900/30 flex items-center justify-center">
								<CheckSquare class="w-5 h-5 text-green-600 dark:text-green-400" />
							</div>
							<span class="text-sm font-medium text-gray-600 dark:text-gray-400">Open Tasks</span>
						</div>
						<p class="text-3xl font-bold text-gray-900 dark:text-white">{member.open_tasks}</p>
					</div>

					<!-- Capacity -->
					<div class="col-span-2 bg-white dark:bg-gray-800 rounded-xl p-5 border border-gray-200 dark:border-gray-700">
						<div class="flex items-center justify-between mb-4">
							<span class="text-sm font-medium text-gray-600 dark:text-gray-400">Current Capacity</span>
							<span class="text-sm font-semibold text-gray-900 dark:text-white">{member.capacity}%</span>
						</div>
						<CapacityBar capacity={member.capacity} size="lg" />
					</div>
				</div>

				<!-- Skills -->
				<div class="bg-white dark:bg-gray-800 rounded-xl p-5 border border-gray-200 dark:border-gray-700">
					<h3 class="text-sm font-semibold text-gray-900 dark:text-white mb-4">Skills</h3>
					{#if member.skills && member.skills.length > 0}
						<div class="flex flex-wrap gap-2">
							{#each member.skills as skill}
								<span class="px-3 py-1.5 bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 text-sm rounded-full">
									{skill}
								</span>
							{/each}
						</div>
					{:else}
						<p class="text-sm text-gray-500 dark:text-gray-400">No skills listed</p>
					{/if}
				</div>

				<!-- Recent Activity -->
				<div class="lg:col-span-3 bg-white dark:bg-gray-800 rounded-xl p-5 border border-gray-200 dark:border-gray-700">
					<h3 class="text-sm font-semibold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
						<Clock class="w-4 h-4" />
						Recent Activity
					</h3>
					{#if member.activities && member.activities.length > 0}
						<div class="space-y-3">
							{#each member.activities.slice(0, 5) as activity}
								<div class="flex items-start gap-3 p-3 rounded-lg bg-gray-50 dark:bg-gray-700/50">
									<div class="w-2 h-2 rounded-full bg-blue-500 mt-2 flex-shrink-0"></div>
									<div class="flex-1 min-w-0">
										<p class="text-sm text-gray-900 dark:text-white">{activity.description}</p>
										<p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
											{formatDate(activity.created_at)}
										</p>
									</div>
								</div>
							{/each}
						</div>
					{:else}
						<p class="text-sm text-gray-500 dark:text-gray-400">No recent activity</p>
					{/if}
				</div>
			</div>
		</div>
	{/if}
</div>
