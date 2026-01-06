<script lang="ts">
	import { DropdownMenu } from 'bits-ui';
	import type { ProjectRole } from '$lib/api/projects/types';
	import { Shield, Users, Eye, Edit3 } from 'lucide-svelte';

	interface Props {
		value: ProjectRole;
		disabled?: boolean;
		onChange?: (role: ProjectRole) => void;
	}

	let { value = $bindable('viewer'), disabled = false, onChange }: Props = $props();

	const roles: Array<{
		value: ProjectRole;
		label: string;
		description: string;
		icon: typeof Shield;
		color: string;
	}> = [
		{
			value: 'lead',
			label: 'Project Lead',
			description: 'Full control - can manage members and settings',
			icon: Shield,
			color: 'text-purple-600'
		},
		{
			value: 'contributor',
			label: 'Contributor',
			description: 'Can edit and contribute to project',
			icon: Edit3,
			color: 'text-blue-600'
		},
		{
			value: 'reviewer',
			label: 'Reviewer',
			description: 'Can review and comment',
			icon: Users,
			color: 'text-green-600'
		},
		{
			value: 'viewer',
			label: 'Viewer',
			description: 'Read-only access',
			icon: Eye,
			color: 'text-gray-600'
		}
	];

	const selectedRole = $derived(roles.find((r) => r.value === value) || roles[3]);

	function handleRoleSelect(role: ProjectRole) {
		value = role;
		onChange?.(role);
	}
</script>

<DropdownMenu.Root>
	<DropdownMenu.Trigger
		class="flex items-center justify-between gap-2 px-3 py-2 text-sm border border-gray-200 rounded-lg hover:bg-gray-50 transition-colors disabled:opacity-50 disabled:cursor-not-allowed min-w-[180px]"
		{disabled}
	>
		<div class="flex items-center gap-2">
			<svelte:component this={selectedRole.icon} class="w-4 h-4 {selectedRole.color}" />
			<span class="font-medium text-gray-900">{selectedRole.label}</span>
		</div>
		<svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
		</svg>
	</DropdownMenu.Trigger>

	<DropdownMenu.Portal>
		<DropdownMenu.Content
			class="z-50 min-w-[280px] bg-white border border-gray-200 rounded-xl shadow-lg p-1 animate-in fade-in-0 zoom-in-95"
			sideOffset={4}
		>
			{#each roles as role}
				<DropdownMenu.Item
					class="px-3 py-2.5 hover:bg-gray-50 rounded-lg cursor-pointer transition-colors"
					onclick={() => handleRoleSelect(role.value)}
				>
					<div class="flex items-start gap-3">
						<svelte:component this={role.icon} class="w-4 h-4 {role.color} mt-0.5" />
						<div class="flex-1 min-w-0">
							<div class="flex items-center gap-2">
								<span class="text-sm font-medium text-gray-900">{role.label}</span>
								{#if role.value === value}
									<svg class="w-4 h-4 text-blue-600" fill="currentColor" viewBox="0 0 20 20">
										<path
											fill-rule="evenodd"
											d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
											clip-rule="evenodd"
										/>
									</svg>
								{/if}
							</div>
							<p class="text-xs text-gray-500 mt-0.5">{role.description}</p>
						</div>
					</div>
				</DropdownMenu.Item>
			{/each}
		</DropdownMenu.Content>
	</DropdownMenu.Portal>
</DropdownMenu.Root>
