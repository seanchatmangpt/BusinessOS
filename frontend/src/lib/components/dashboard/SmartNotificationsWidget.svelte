<script lang="ts">
	import { fade, fly } from 'svelte/transition';
	import type { DashboardTask, DashboardProject } from '$lib/api';

	interface SmartNotification {
		id: string;
		type: 'alert' | 'reminder' | 'update' | 'celebration';
		title: string;
		message: string;
		time: string;
		priority: 'high' | 'medium' | 'low';
		actionable?: boolean;
		onAction?: () => void;
	}

	interface Props {
		tasks?: DashboardTask[];
		projects?: DashboardProject[];
		onAction?: (action: string, context?: any) => void;
		onViewAll?: () => void;
	}

	let { tasks = [], projects = [], onAction, onViewAll }: Props = $props();

	// Generate smart notifications based on data
	const notifications = $derived<SmartNotification[]>(() => {
		const generated: SmartNotification[] = [];
		const now = new Date();

		// Overdue tasks alert
		const overdue = tasks.filter((t) => {
			if (t.completed || !t.due_date) return false;
			return new Date(t.due_date) < now;
		});

		if (overdue.length > 0) {
			generated.push({
				id: 'overdue',
				type: 'alert',
				title: 'Overdue tasks',
				message: `${overdue.length} ${overdue.length === 1 ? 'task is' : 'tasks are'} past ${overdue.length === 1 ? 'its' : 'their'} deadline`,
				time: 'now',
				priority: 'high',
				actionable: true,
				onAction: () => onAction?.('view-overdue')
			});
		}

		// Due today reminder
		const dueToday = tasks.filter((t) => {
			if (t.completed || !t.due_date) return false;
			const due = new Date(t.due_date);
			return due.toDateString() === now.toDateString();
		});

		if (dueToday.length > 0) {
			generated.push({
				id: 'due-today',
				type: 'reminder',
				title: 'Tasks due today',
				message: `${dueToday.length} ${dueToday.length === 1 ? 'task needs' : 'tasks need'} your attention before end of day`,
				time: formatTime(now),
				priority: 'medium',
				actionable: true,
				onAction: () => onAction?.('view-today')
			});
		}

		// Project updates
		const criticalProjects = projects.filter((p) => p.health === 'critical');
		if (criticalProjects.length > 0) {
			generated.push({
				id: 'critical-projects',
				type: 'alert',
				title: 'Critical project alert',
				message: `${criticalProjects.map((p) => p.name).join(', ')} ${criticalProjects.length === 1 ? 'requires' : 'require'} immediate attention`,
				time: formatTime(now),
				priority: 'high',
				actionable: true,
				onAction: () => onAction?.('view-projects')
			});
		}

		// Completed tasks celebration
		const completedToday = tasks.filter((t) => {
			if (!t.completed) return false;
			// In real app, check completion timestamp
			return true;
		});

		if (completedToday.length >= 5) {
			generated.push({
				id: 'celebration',
				type: 'celebration',
				title: 'Great progress!',
				message: `You've completed ${completedToday.length} tasks today. Keep it up!`,
				time: formatTime(now),
				priority: 'low'
			});
		}

		// Return top 4
		return generated.slice(0, 4);
	});

	function formatTime(date: Date): string {
		const hours = date.getHours();
		const minutes = date.getMinutes();
		const ampm = hours >= 12 ? 'PM' : 'AM';
		const displayHours = hours % 12 || 12;
		const displayMinutes = minutes.toString().padStart(2, '0');
		return `${displayHours}:${displayMinutes} ${ampm}`;
	}

	const typeConfig = {
		alert: {
			icon: 'M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z',
			color: 'text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-500/10'
		},
		reminder: {
			icon: 'M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z',
			color: 'text-blue-600 dark:text-blue-400 bg-blue-50 dark:bg-blue-500/10'
		},
		update: {
			icon: 'M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z',
			color: 'text-gray-600 dark:text-gray-400 bg-gray-50 dark:bg-gray-500/10'
		},
		celebration: {
			icon: 'M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z',
			color: 'text-green-600 dark:text-green-400 bg-green-50 dark:bg-green-500/10'
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
				class="w-8 h-8 rounded-lg bg-gradient-to-br from-red-500 to-red-600 flex items-center justify-center shadow-sm"
			>
				<svg class="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9"
					/>
				</svg>
			</div>
			<h2 class="text-base font-semibold text-gray-900 dark:text-white/90">Alerts</h2>
			{#if notifications.length > 0}
				<span
					class="px-2 py-0.5 text-xs font-semibold rounded-full bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-400"
				>
					{notifications.length}
				</span>
			{/if}
		</div>
		{#if notifications.length > 0}
			<button
				onclick={onViewAll}
				class="btn-pill-ghost btn-pill-xs"
			>
				View all
			</button>
		{/if}
	</div>

	<!-- Notifications -->
	{#if notifications.length === 0}
		<div class="text-center py-8" transition:fade={{ duration: 200 }}>
			<div
				class="w-14 h-14 bg-gradient-to-br from-gray-100 to-gray-50 dark:from-gray-500/20 dark:to-gray-500/10 rounded-xl flex items-center justify-center mx-auto mb-3 shadow-sm"
			>
				<svg
					class="w-7 h-7 text-gray-400 dark:text-gray-500"
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
			<p class="text-sm font-medium text-gray-900 dark:text-white/90 mb-1">No alerts</p>
			<p class="text-xs text-gray-500 dark:text-gray-400">You're all caught up!</p>
		</div>
	{:else}
		<div class="space-y-3">
			{#each notifications as notif, index (notif.id)}
				<div
					class="flex items-start gap-3 p-3 rounded-lg border border-gray-200 dark:border-white/10 hover:bg-gray-50 dark:hover:bg-white/5 transition-colors"
					in:fly={{ x: -10, duration: 300, delay: index * 75 }}
				>
					<!-- Icon -->
					<div class="flex-shrink-0 w-8 h-8 rounded-lg {typeConfig[notif.type].color} flex items-center justify-center">
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d={typeConfig[notif.type].icon}
							/>
						</svg>
					</div>

					<!-- Content -->
					<div class="flex-1 min-w-0">
						<div class="flex items-start justify-between gap-2 mb-1">
							<p class="text-sm font-semibold text-gray-900 dark:text-white">
								{notif.title}
							</p>
							<span class="text-xs text-gray-500 dark:text-gray-400 whitespace-nowrap">
								{notif.time}
							</span>
						</div>
						<p class="text-xs text-gray-600 dark:text-gray-400 mb-2">
							{notif.message}
						</p>

						{#if notif.actionable && notif.onAction}
							<button
								onclick={notif.onAction}
								class="btn-pill-link"
							>
								Take action →
							</button>
						{/if}
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>
