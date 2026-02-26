<script lang="ts">
	import { Clock, Mail, Copy, Check, X } from 'lucide-svelte';
	import type { WorkspaceInvite } from '$lib/api/workspaces';

	interface Props {
		invite: WorkspaceInvite;
		onCopy?: (invite: WorkspaceInvite) => void;
		onRevoke?: (id: string) => void;
	}

	let { invite, onCopy, onRevoke }: Props = $props();

	let copied = $state(false);

	function formatExpiry(date: string) {
		const d = new Date(date);
		const now = new Date();
		const days = Math.ceil((d.getTime() - now.getTime()) / (1000 * 60 * 60 * 24));
		if (days <= 0) return 'Expired';
		if (days === 1) return '1 day';
		return `${days} days`;
	}

	function getRoleBadgeClasses(role: string) {
		switch (role) {
			case 'admin':
				return 'bg-purple-100 text-purple-700 dark:bg-purple-500/20 dark:text-purple-400';
			case 'manager':
				return 'bg-blue-100 text-blue-700 dark:bg-blue-500/20 dark:text-blue-400';
			default:
				return 'bg-gray-100 text-gray-700 dark:bg-gray-500/20 dark:text-gray-400';
		}
	}

	async function handleCopy() {
		onCopy?.(invite);
		copied = true;
		setTimeout(() => copied = false, 2000);
	}
</script>

<div
	class="group bg-white dark:bg-gray-800 border border-dashed border-gray-300 dark:border-gray-600 rounded-xl p-5 hover:border-gray-400 dark:hover:border-gray-500 transition-all duration-200 relative"
>
	<!-- Pending Badge -->
	<div class="absolute top-3 right-3">
		<span class="inline-flex items-center gap-1 px-2 py-0.5 text-xs font-medium rounded-full bg-amber-100 text-amber-700 dark:bg-amber-500/20 dark:text-amber-400">
			<Clock class="w-3 h-3" />
			Pending
		</span>
	</div>

	<!-- Avatar placeholder -->
	<div class="flex flex-col items-center text-center mb-4">
		<div class="w-16 h-16 rounded-full bg-gray-100 dark:bg-gray-700 flex items-center justify-center mb-3 ring-2 ring-gray-100 dark:ring-gray-700">
			<Mail class="w-7 h-7 text-gray-400 dark:text-gray-500" />
		</div>

		<h3 class="font-semibold text-gray-900 dark:text-white truncate max-w-full">{invite.email}</h3>
		<p class="text-sm text-gray-500 dark:text-gray-400">Invitation sent</p>
	</div>

	<!-- Divider -->
	<div class="border-t border-gray-100 dark:border-gray-700 my-4"></div>

	<!-- Role Badge -->
	<div class="flex justify-center mb-4">
		<span class="inline-flex items-center px-2.5 py-1 text-xs font-medium rounded-full capitalize {getRoleBadgeClasses(invite.role)}">
			{invite.role}
		</span>
	</div>

	<!-- Info -->
	<div class="space-y-2 text-sm">
		<div class="flex items-center justify-between text-gray-600 dark:text-gray-400">
			<span class="flex items-center gap-1.5">
				<Clock class="w-4 h-4 text-gray-400 dark:text-gray-500" />
				Expires in
			</span>
			<span class="font-medium text-gray-900 dark:text-white">{formatExpiry(invite.expires_at)}</span>
		</div>
	</div>

	<!-- Actions -->
	<div class="flex gap-2 mt-4">
		<button
			onclick={handleCopy}
			class="flex-1 flex items-center justify-center gap-1.5 px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-gray-50 dark:bg-gray-700 hover:bg-gray-100 dark:hover:bg-gray-600 rounded-lg transition-colors"
		>
			{#if copied}
				<Check class="w-4 h-4 text-green-500" />
				Copied
			{:else}
				<Copy class="w-4 h-4" />
				Copy Link
			{/if}
		</button>
		<button
			onclick={() => onRevoke?.(invite.id)}
			class="px-3 py-2 text-sm font-medium text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-900/20 hover:bg-red-100 dark:hover:bg-red-900/40 rounded-lg transition-colors"
			title="Revoke invitation"
		>
			<X class="w-4 h-4" />
		</button>
	</div>
</div>
