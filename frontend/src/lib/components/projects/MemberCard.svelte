<script lang="ts">
	import type { ProjectMember, ProjectRole } from '$lib/api/projects/types';
	import RoleSelector from './RoleSelector.svelte';
	import { Shield, Users, Eye, Edit3, Trash2, MoreVertical } from 'lucide-svelte';
	import { DropdownMenu } from 'bits-ui';

	interface Props {
		member: ProjectMember;
		canEdit?: boolean;
		canRemove?: boolean;
		currentUserId?: string;
		onRoleChange?: (memberId: string, newRole: ProjectRole) => void;
		onRemove?: (memberId: string) => void;
	}

	let {
		member,
		canEdit = false,
		canRemove = false,
		currentUserId = '',
		onRoleChange,
		onRemove
	}: Props = $props();

	function getInitials(name: string): string {
		if (!name) return '?';
		return name
			.split(' ')
			.map((n) => n.charAt(0))
			.join('')
			.toUpperCase()
			.slice(0, 2);
	}

	function getRoleIcon(role: ProjectRole) {
		switch (role) {
			case 'lead':
				return Shield;
			case 'contributor':
				return Edit3;
			case 'reviewer':
				return Users;
			case 'viewer':
				return Eye;
		}
	}

	function getRoleColor(role: ProjectRole): string {
		switch (role) {
			case 'lead':
				return 'bg-purple-100 text-purple-700 border-purple-200';
			case 'contributor':
				return 'bg-blue-100 text-blue-700 border-blue-200';
			case 'reviewer':
				return 'bg-green-100 text-green-700 border-green-200';
			case 'viewer':
				return 'bg-gray-100 text-gray-700 border-gray-200';
		}
	}

	function getRoleLabel(role: ProjectRole): string {
		switch (role) {
			case 'lead':
				return 'Project Lead';
			case 'contributor':
				return 'Contributor';
			case 'reviewer':
				return 'Reviewer';
			case 'viewer':
				return 'Viewer';
		}
	}

	function getPermissionsList(member: ProjectMember): string[] {
		const permissions: string[] = [];
		if (member.can_edit) permissions.push('Edit');
		if (member.can_delete) permissions.push('Delete');
		if (member.can_invite) permissions.push('Invite');
		if (permissions.length === 0) permissions.push('Read-only');
		return permissions;
	}

	function handleRoleChange(newRole: ProjectRole) {
		onRoleChange?.(member.id, newRole);
	}

	function handleRemove() {
		if (confirm('Are you sure you want to remove this member from the project?')) {
			onRemove?.(member.id);
		}
	}

	const isCurrentUser = $derived(member.user_id === currentUserId);
	const RoleIcon = $derived(getRoleIcon(member.role));
	const permissions = $derived(getPermissionsList(member));
</script>

<div class="bg-white border border-gray-200 rounded-xl p-4 hover:shadow-md transition-all duration-200">
	<div class="flex items-start gap-4">
		<!-- Avatar -->
		<div class="flex-shrink-0">
			{#if member.user_avatar}
				<img
					src={member.user_avatar}
					alt={member.user_name || 'User'}
					class="w-12 h-12 rounded-full object-cover"
				/>
			{:else}
				<div
					class="w-12 h-12 rounded-full bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center"
				>
					<span class="text-white font-semibold text-sm">
						{getInitials(member.user_name || member.user_id)}
					</span>
				</div>
			{/if}
		</div>

		<!-- Member Info -->
		<div class="flex-1 min-w-0">
			<div class="flex items-start justify-between gap-2">
				<div class="flex-1 min-w-0">
					<div class="flex items-center gap-2">
						<h3 class="font-semibold text-gray-900 truncate">
							{member.user_name || member.user_id}
						</h3>
						{#if isCurrentUser}
							<span class="px-2 py-0.5 text-xs font-medium bg-blue-50 text-blue-700 rounded-full">
								You
							</span>
						{/if}
					</div>
					{#if member.user_email}
						<p class="text-sm text-gray-500 truncate">{member.user_email}</p>
					{/if}
				</div>

				<!-- Actions Menu -->
				{#if (canEdit || canRemove) && !isCurrentUser}
					<DropdownMenu.Root>
						<DropdownMenu.Trigger
							class="p-1.5 rounded-lg hover:bg-gray-100 transition-colors"
							aria-label="Member actions"
						>
							<MoreVertical class="w-4 h-4 text-gray-500" />
						</DropdownMenu.Trigger>
						<DropdownMenu.Portal>
							<DropdownMenu.Content
								class="z-50 min-w-[160px] bg-white border border-gray-200 rounded-lg shadow-lg p-1 animate-in fade-in-0 zoom-in-95"
								sideOffset={4}
							>
								{#if canRemove}
									<DropdownMenu.Item
										class="flex items-center gap-2 px-3 py-2 text-sm text-red-600 hover:bg-red-50 rounded-md cursor-pointer"
										onclick={handleRemove}
									>
										<Trash2 class="w-4 h-4" />
										<span>Remove member</span>
									</DropdownMenu.Item>
								{/if}
							</DropdownMenu.Content>
						</DropdownMenu.Portal>
					</DropdownMenu.Root>
				{/if}
			</div>

			<!-- Role Badge and Selector -->
			<div class="mt-3">
				{#if canEdit && !isCurrentUser}
					<RoleSelector value={member.role} onChange={handleRoleChange} />
				{:else}
					<div class="inline-flex items-center gap-2 px-3 py-1.5 rounded-lg border {getRoleColor(member.role)}">
						<svelte:component this={RoleIcon} class="w-3.5 h-3.5" />
						<span class="text-xs font-medium">{getRoleLabel(member.role)}</span>
					</div>
				{/if}
			</div>

			<!-- Permissions -->
			<div class="mt-3 flex flex-wrap gap-1.5">
				{#each permissions as permission}
					<span class="px-2 py-0.5 text-xs bg-gray-50 text-gray-600 rounded border border-gray-200">
						{permission}
					</span>
				{/each}
			</div>

			<!-- Member Since -->
			<div class="mt-2 text-xs text-gray-400">
				Member since {new Date(member.assigned_at).toLocaleDateString()}
			</div>
		</div>
	</div>
</div>
