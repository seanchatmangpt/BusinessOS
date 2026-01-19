<script lang="ts">
	import { fly, scale } from 'svelte/transition';
	import { goto } from '$app/navigation';
	import type { DashboardTask } from '$lib/api';

	interface Props {
		tasks?: DashboardTask[];
		onToggle?: (id: string) => void;
		onViewAll?: () => void;
	}

	let { tasks = [], onToggle, onViewAll }: Props = $props();

	type TaskPriority = 'critical' | 'high' | 'medium' | 'low';

	// Categorize tasks
	const categorizedTasks = $derived(() => {
		const today = new Date();
		today.setHours(0, 0, 0, 0);
		const nextWeek = new Date(today);
		nextWeek.setDate(nextWeek.getDate() + 7);

		const dueToday: DashboardTask[] = [];
		const upcoming: DashboardTask[] = [];
		const overdue: DashboardTask[] = [];

		for (const task of tasks.filter((t) => !t.completed)) {
			if (!task.due_date) {
				upcoming.push(task);
				continue;
			}
			const due = new Date(task.due_date);
			due.setHours(0, 0, 0, 0);

			if (due < today) {
				overdue.push(task);
			} else if (due.getTime() === today.getTime()) {
				dueToday.push(task);
			} else if (due <= nextWeek) {
				upcoming.push(task);
			}
		}

		return { dueToday, upcoming, overdue };
	});

	const priorityColors: Record<TaskPriority, string> = {
		critical: 'text-red-600 bg-red-50',
		high: 'text-orange-600 bg-orange-50',
		medium: 'text-yellow-600 bg-yellow-50',
		low: 'text-gray-600 bg-gray-50'
	};

	function formatDueDate(due_date?: string | null): string {
		if (!due_date) return '';
		const due = new Date(due_date);
		const today = new Date();
		today.setHours(0, 0, 0, 0);
		const diff = Math.ceil((due.getTime() - today.getTime()) / (1000 * 60 * 60 * 24));

		if (diff < 0) return `${Math.abs(diff)}d overdue`;
		if (diff === 0) return 'Today';
		if (diff === 1) return 'Tomorrow';
		return due.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
	}

	function handleToggle(id: string) {
		onToggle?.(id);
	}
</script>

<div class="bg-white dark:bg-[#1c1c1e] rounded-xl border border-gray-200 dark:border-white/10 p-6 shadow-sm hover:shadow-md transition-shadow duration-300">
	<div class="flex items-center justify-between mb-5">
		<div class="flex items-center gap-3">
			<div class="w-9 h-9 rounded-lg bg-gradient-to-br from-green-500 to-green-600 flex items-center justify-center shadow-md">
				<svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
				</svg>
			</div>
			<h2 class="text-lg font-bold text-gray-900 dark:text-white">My Tasks</h2>
		</div>
		{#if tasks.length > 0}
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

	{#if tasks.length === 0}
		<div class="text-center py-10">
			<div class="w-16 h-16 bg-gradient-to-br from-green-100 to-green-50 dark:from-green-500/20 dark:to-green-500/10 rounded-2xl flex items-center justify-center mx-auto mb-4 shadow-sm">
				<svg class="w-8 h-8 text-green-600 dark:text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
				</svg>
			</div>
			<p class="text-base font-semibold text-gray-900 dark:text-white mb-1">All caught up!</p>
			<p class="text-sm text-gray-500 dark:text-gray-400 mb-4">No tasks due soon.</p>
			<button
				onclick={() => goto('/tasks')}
				class="btn-pill-outline btn-pill-sm inline-flex items-center gap-2"
			>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
				</svg>
				Add a task
			</button>
		</div>
	{:else}
		<div class="space-y-5 max-h-80 overflow-y-auto">
			<!-- Overdue -->
			{#if categorizedTasks().overdue.length > 0}
				<div>
					<h3 class="text-xs font-bold text-red-600 dark:text-red-400 uppercase tracking-wider mb-3 flex items-center gap-2">
						<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
						</svg>
						Overdue ({categorizedTasks().overdue.length})
					</h3>
					<div class="space-y-2">
						{#each categorizedTasks().overdue.slice(0, 3) as task (task.id)}
							<div
								class="flex items-center gap-3 p-3 rounded-lg bg-red-50 dark:bg-red-500/10 border border-red-100 dark:border-red-500/20 hover:border-red-200 dark:hover:border-red-500/30 transition-all"
								in:fly={{ x: -10, duration: 200 }}
							>
								<button
									onclick={() => handleToggle(task.id)}
									class="flex-shrink-0 w-5 h-5 rounded border-2 border-red-400 dark:border-red-400 hover:border-red-500 dark:hover:border-red-300 hover:bg-red-100 dark:hover:bg-red-500/20 transition-colors"
								></button>
								<div class="flex-1 min-w-0">
									<p class="text-sm font-medium text-gray-900 dark:text-white truncate">{task.title}</p>
									{#if task.project_name}
										<p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">{task.project_name}</p>
									{/if}
								</div>
								<span class="text-xs font-semibold text-red-600 dark:text-red-400 whitespace-nowrap flex items-center gap-1">
									<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
									</svg>
									{formatDueDate(task.due_date)}
								</span>
							</div>
						{/each}
					</div>
				</div>
			{/if}

			<!-- Due Today -->
			{#if categorizedTasks().dueToday.length > 0}
				<div>
					<h3 class="text-xs font-bold text-blue-600 dark:text-blue-400 uppercase tracking-wider mb-3 flex items-center gap-2">
						<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
						</svg>
						Due Today ({categorizedTasks().dueToday.length})
					</h3>
					<div class="space-y-2">
						{#each categorizedTasks().dueToday.slice(0, 3) as task (task.id)}
							<div
								class="flex items-center gap-3 p-3 rounded-lg border border-gray-200 dark:border-white/10 hover:bg-gray-50 dark:hover:bg-white/5 hover:border-gray-300 dark:hover:border-white/20 transition-all"
								in:fly={{ x: -10, duration: 200 }}
							>
								<button
									onclick={() => handleToggle(task.id)}
									class="flex-shrink-0 w-5 h-5 rounded border-2 border-gray-300 dark:border-gray-600 hover:border-blue-500 dark:hover:border-blue-400 hover:bg-blue-50 dark:hover:bg-blue-500/20 transition-colors"
								></button>
								<div class="flex-1 min-w-0">
									<p class="text-sm font-medium text-gray-900 dark:text-white truncate">{task.title}</p>
									{#if task.project_name}
										<p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">{task.project_name}</p>
									{/if}
								</div>
								<span
									class="text-xs font-semibold px-2.5 py-1 rounded-md {priorityColors[task.priority]} capitalize"
								>
									{task.priority}
								</span>
							</div>
						{/each}
					</div>
				</div>
			{/if}

			<!-- Upcoming -->
			{#if categorizedTasks().upcoming.length > 0}
				<div>
					<h3 class="text-xs font-bold text-gray-600 dark:text-gray-300 uppercase tracking-wider mb-3 flex items-center gap-2">
						<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
						</svg>
						Upcoming ({categorizedTasks().upcoming.length})
					</h3>
					<div class="space-y-2">
						{#each categorizedTasks().upcoming.slice(0, 3) as task (task.id)}
							<div
								class="flex items-center gap-3 p-3 rounded-lg border border-gray-200 dark:border-white/10 hover:bg-gray-50 dark:hover:bg-white/5 hover:border-gray-300 dark:hover:border-white/20 transition-all"
								in:fly={{ x: -10, duration: 200 }}
							>
								<button
									onclick={() => handleToggle(task.id)}
									class="flex-shrink-0 w-5 h-5 rounded border-2 border-gray-300 dark:border-gray-600 hover:border-gray-500 dark:hover:border-gray-400 hover:bg-gray-100 dark:hover:bg-white/10 transition-colors"
								></button>
								<div class="flex-1 min-w-0">
									<p class="text-sm font-medium text-gray-900 dark:text-white truncate">{task.title}</p>
									{#if task.project_name}
										<p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">{task.project_name}</p>
									{/if}
								</div>
								<span class="text-xs font-medium text-gray-500 dark:text-gray-400 whitespace-nowrap">
									{formatDueDate(task.due_date)}
								</span>
							</div>
						{/each}
					</div>
				</div>
			{/if}
		</div>
	{/if}
</div>
