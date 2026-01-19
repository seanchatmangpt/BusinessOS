<script lang="ts">
	import { fly } from 'svelte/transition';
	import { goto } from '$app/navigation';
	import type { DashboardProject } from '$lib/api';

	interface Props {
		projects?: DashboardProject[];
		onViewAll?: () => void;
	}

	let { projects = [], onViewAll }: Props = $props();

	type ProjectHealth = 'healthy' | 'at_risk' | 'critical';

	const healthColors: Record<ProjectHealth, string> = {
		healthy: 'bg-green-500',
		at_risk: 'bg-yellow-500',
		critical: 'bg-red-500'
	};

	const healthLabels: Record<ProjectHealth, string> = {
		healthy: 'On Track',
		at_risk: 'At Risk',
		critical: 'Critical'
	};

	function getDaysRemaining(due_date?: string | null): string {
		if (!due_date) return '';
		const due = new Date(due_date);
		const now = new Date();
		const days = Math.ceil((due.getTime() - now.getTime()) / (1000 * 60 * 60 * 24));
		if (days < 0) return `${Math.abs(days)}d overdue`;
		if (days === 0) return 'Due today';
		if (days === 1) return 'Due tomorrow';
		return `${days} days left`;
	}

	function handleProjectClick(projectId: string) {
		goto(`/projects/${projectId}`);
	}
</script>

<div class="bg-white dark:bg-[#1c1c1e] rounded-xl border border-gray-200 dark:border-white/10 p-6 shadow-sm hover:shadow-md transition-shadow duration-300">
	<div class="flex items-center justify-between mb-5">
		<div class="flex items-center gap-3">
			<div class="w-9 h-9 rounded-lg bg-gradient-to-br from-purple-500 to-purple-600 flex items-center justify-center shadow-md">
				<svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
				</svg>
			</div>
			<h2 class="text-lg font-bold text-gray-900 dark:text-white">Active Projects</h2>
		</div>
		{#if projects.length > 0}
			<button
				onclick={() => onViewAll?.()}
				class="btn-pill-ghost btn-pill-xs flex items-center gap-1.5"
			>
				View All
				<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
				</svg>
			</button>
		{/if}
	</div>

	{#if projects.length === 0}
		<div class="text-center py-10">
			<div class="w-16 h-16 bg-gradient-to-br from-purple-100 to-purple-50 dark:from-purple-500/20 dark:to-purple-500/10 rounded-2xl flex items-center justify-center mx-auto mb-4 shadow-sm">
				<svg class="w-8 h-8 text-purple-600 dark:text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"
					/>
				</svg>
			</div>
			<p class="text-base font-semibold text-gray-900 dark:text-white mb-1">No active projects</p>
			<p class="text-sm text-gray-500 dark:text-gray-400 mb-4">Create your first project to get started</p>
			<button
				onclick={() => goto('/projects')}
				class="btn-pill-outline btn-pill-sm inline-flex items-center gap-2"
			>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
				</svg>
				Create project
			</button>
		</div>
	{:else}
		<div class="flex gap-4 overflow-x-auto pb-2 -mx-1 px-1">
			{#each projects.slice(0, 5) as project, index (project.id)}
				<button
					onclick={() => handleProjectClick(project.id)}
					class="btn-pill-soft flex-shrink-0 w-52 text-left group"
					in:fly={{ y: 20, duration: 400, delay: index * 100 }}
				>
					<!-- Health indicator and name -->
					<div class="flex items-start gap-2 mb-3">
						<div class="w-2 h-2 rounded-full mt-1.5 {healthColors[project.health]}"></div>
						<div class="flex-1 min-w-0">
							<h3 class="text-sm font-medium text-gray-900 truncate group-hover:text-gray-700">
								{project.name}
							</h3>
						</div>
					</div>

					<!-- Client or Type -->
					<p class="text-xs text-gray-500 mb-1">
						{project.client_name ? `Client: ${project.client_name}` : project.project_type}
					</p>

					<!-- Due date -->
					{#if project.due_date}
						<p
							class="text-xs mb-3 {project.health === 'critical'
								? 'text-red-600'
								: 'text-gray-500'}"
						>
							{getDaysRemaining(project.due_date)}
						</p>
					{/if}

					<!-- Progress bar -->
					<div class="mb-2">
						<div class="h-1.5 bg-gray-100 rounded-full overflow-hidden">
							<div
								class="h-full rounded-full transition-all duration-500 {project.health ===
								'critical'
									? 'bg-red-500'
									: project.health === 'at_risk'
										? 'bg-yellow-500'
										: 'bg-green-500'}"
								style="width: {project.progress}%"
							></div>
						</div>
						<p class="text-xs text-gray-400 mt-1">{project.progress}% complete</p>
					</div>

					<!-- Team -->
					<div class="flex items-center gap-1 text-xs text-gray-500">
						<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197m13.5-9a2.5 2.5 0 11-5 0 2.5 2.5 0 015 0z"
							/>
						</svg>
						<span>{project.team_count} member{project.team_count !== 1 ? 's' : ''}</span>
					</div>
				</button>
			{/each}
		</div>
	{/if}
</div>
