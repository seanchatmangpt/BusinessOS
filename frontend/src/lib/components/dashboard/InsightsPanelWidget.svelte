<script lang="ts">
	import { fade, fly } from 'svelte/transition';
	import type { DashboardTask, DashboardProject } from '$lib/api';

	interface Insight {
		id: string;
		type: 'success' | 'warning' | 'info' | 'tip';
		title: string;
		description: string;
		action?: {
			label: string;
			onClick: () => void;
		};
	}

	interface Props {
		tasks?: DashboardTask[];
		projects?: DashboardProject[];
		onAction?: (action: string, context?: any) => void;
	}

	let { tasks = [], projects = [], onAction }: Props = $props();

	// Generate insights based on data
	const insights = $derived<Insight[]>(() => {
		const generated: Insight[] = [];

		// Task insights
		const incompleteTasks = tasks.filter((t) => !t.completed);
		const overdueTasks = incompleteTasks.filter((t) => {
			if (!t.due_date) return false;
			return new Date(t.due_date) < new Date();
		});

		if (overdueTasks.length > 0) {
			generated.push({
				id: 'overdue-tasks',
				type: 'warning',
				title: `${overdueTasks.length} overdue ${overdueTasks.length === 1 ? 'task' : 'tasks'}`,
				description: `You have ${overdueTasks.length} tasks past their due date. Focus on these first to get back on track.`,
				action: {
					label: 'View overdue',
					onClick: () => onAction?.('view-overdue-tasks')
				}
			});
		}

		// Productivity insight
		const completedToday = tasks.filter((t) => {
			if (!t.completed) return false;
			// Simple check - in real app would check actual completion timestamp
			return true;
		}).length;

		if (completedToday > 3) {
			generated.push({
				id: 'productive-day',
				type: 'success',
				title: 'Productive day!',
				description: `You've completed ${completedToday} tasks. Keep up the great momentum!`,
			});
		}

		// Project health insights
		const atRiskProjects = projects.filter((p) => p.health === 'at_risk');
		if (atRiskProjects.length > 0) {
			generated.push({
				id: 'at-risk-projects',
				type: 'warning',
				title: `${atRiskProjects.length} ${atRiskProjects.length === 1 ? 'project' : 'projects'} at risk`,
				description: `${atRiskProjects.map((p) => p.name).join(', ')} ${atRiskProjects.length === 1 ? 'needs' : 'need'} attention to stay on track.`,
				action: {
					label: 'Review projects',
					onClick: () => onAction?.('view-projects')
				}
			});
		}

		// Due today insight
		const dueToday = incompleteTasks.filter((t) => {
			if (!t.due_date) return false;
			const due = new Date(t.due_date);
			const today = new Date();
			return due.toDateString() === today.toDateString();
		});

		if (dueToday.length > 0) {
			generated.push({
				id: 'due-today',
				type: 'info',
				title: `${dueToday.length} ${dueToday.length === 1 ? 'task' : 'tasks'} due today`,
				description: `Stay focused on completing these before end of day.`,
				action: {
					label: 'View tasks',
					onClick: () => onAction?.('view-tasks')
				}
			});
		}

		// Weekly planning tip
		const now = new Date();
		if (now.getDay() === 1 && now.getHours() < 12) {
			// Monday morning
			generated.push({
				id: 'weekly-planning',
				type: 'tip',
				title: 'Start your week strong',
				description: 'Take 10 minutes to review your priorities for the week ahead.',
			});
		}

		// Return top 3 insights
		return generated.slice(0, 3);
	});

	const typeStyles = {
		success: {
			border: 'border-green-200 dark:border-green-500/20',
			bg: 'bg-green-50 dark:bg-green-500/10',
			icon: 'text-green-600 dark:text-green-400',
			iconPath:
				'M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z'
		},
		warning: {
			border: 'border-orange-200 dark:border-orange-500/20',
			bg: 'bg-orange-50 dark:bg-orange-500/10',
			icon: 'text-orange-600 dark:text-orange-400',
			iconPath:
				'M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z'
		},
		info: {
			border: 'border-blue-200 dark:border-blue-500/20',
			bg: 'bg-blue-50 dark:bg-blue-500/10',
			icon: 'text-blue-600 dark:text-blue-400',
			iconPath:
				'M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z'
		},
		tip: {
			border: 'border-purple-200 dark:border-purple-500/20',
			bg: 'bg-purple-50 dark:bg-purple-500/10',
			icon: 'text-purple-600 dark:text-purple-400',
			iconPath:
				'M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z'
		}
	};
</script>

<div
	class="bg-white dark:bg-[#1c1c1e] rounded-xl border border-gray-200 dark:border-white/10 p-5 shadow-sm hover:shadow-md transition-shadow duration-300"
>
	<!-- Header -->
	<div class="flex items-center justify-between mb-4">
		<div class="flex items-center gap-2">
			<div
				class="w-8 h-8 rounded-lg bg-gradient-to-br from-purple-500 to-purple-600 flex items-center justify-center shadow-sm"
			>
				<svg class="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M13 10V3L4 14h7v7l9-11h-7z"
					/>
				</svg>
			</div>
			<h2 class="text-base font-semibold text-gray-900 dark:text-white/90">Insights</h2>
		</div>
	</div>

	<!-- Insights -->
	{#if insights.length === 0}
		<div class="text-center py-8" transition:fade={{ duration: 200 }}>
			<div
				class="w-14 h-14 bg-gradient-to-br from-purple-100 to-purple-50 dark:from-purple-500/20 dark:to-purple-500/10 rounded-xl flex items-center justify-center mx-auto mb-3 shadow-sm"
			>
				<svg
					class="w-7 h-7 text-purple-600 dark:text-purple-400"
					fill="none"
					stroke="currentColor"
					viewBox="0 0 24 24"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
					/>
				</svg>
			</div>
			<p class="text-sm font-medium text-gray-900 dark:text-white/90 mb-1">All good!</p>
			<p class="text-xs text-gray-500 dark:text-gray-400">No important insights right now</p>
		</div>
	{:else}
		<div class="space-y-3">
			{#each insights as insight, index (insight.id)}
				<div
					class="border {typeStyles[insight.type].border} {typeStyles[insight.type]
						.bg} rounded-lg p-3"
					in:fly={{ y: 10, duration: 300, delay: index * 100 }}
				>
					<div class="flex items-start gap-3">
						<!-- Icon -->
						<div class="flex-shrink-0 mt-0.5">
							<svg
								class="w-5 h-5 {typeStyles[insight.type].icon}"
								fill="none"
								stroke="currentColor"
								viewBox="0 0 24 24"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d={typeStyles[insight.type].iconPath}
								/>
							</svg>
						</div>

						<!-- Content -->
						<div class="flex-1 min-w-0">
							<p class="text-sm font-semibold text-gray-900 dark:text-white mb-1">
								{insight.title}
							</p>
							<p class="text-xs text-gray-600 dark:text-gray-400">
								{insight.description}
							</p>

							{#if insight.action}
								<button
									onclick={insight.action.onClick}
									class="btn-pill-link mt-2"
								>
									{insight.action.label} →
								</button>
							{/if}
						</div>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>
