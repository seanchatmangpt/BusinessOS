<script lang="ts">
	import { X, Users, User, Check } from 'lucide-svelte';
	import type { SharePermission } from '$lib/types/modules';

	interface Props {
		moduleId: string;
		moduleName: string;
		isOpen: boolean;
		onClose: () => void;
		onShare: (data: { sharedWithUserId?: string; sharedWithWorkspaceId?: string; permissions: SharePermission[] }) => Promise<void>;
	}

	let { moduleId, moduleName, isOpen, onClose, onShare }: Props = $props();

	let shareType = $state<'user' | 'workspace'>('user');
	let userId = $state('');
	let permissions = $state<SharePermission[]>(['view', 'install']);
	let isSharing = $state(false);

	const allPermissions: Array<{ value: SharePermission; label: string; description: string }> = [
		{ value: 'view', label: 'View', description: 'Can view module details' },
		{ value: 'install', label: 'Install', description: 'Can install the module' },
		{ value: 'modify', label: 'Modify', description: 'Can edit the module' },
		{ value: 'reshare', label: 'Reshare', description: 'Can share with others' }
	];

	function togglePermission(permission: SharePermission) {
		if (permissions.includes(permission)) {
			permissions = permissions.filter(p => p !== permission);
		} else {
			permissions = [...permissions, permission];
		}
	}

	async function handleShare() {
		if (shareType === 'user' && !userId.trim()) {
			alert('Please enter a user ID or email');
			return;
		}

		if (permissions.length === 0) {
			alert('Please select at least one permission');
			return;
		}

		isSharing = true;
		try {
			await onShare({
				sharedWithUserId: shareType === 'user' ? userId : undefined,
				sharedWithWorkspaceId: shareType === 'workspace' ? 'current' : undefined,
				permissions
			});
			// Reset form
			userId = '';
			permissions = ['view', 'install'];
			onClose();
		} catch (error) {
			console.error('Failed to share module:', error);
		} finally {
			isSharing = false;
		}
	}
</script>

{#if isOpen}
	<div class="fixed inset-0 z-50 flex items-center justify-center">
		<!-- Backdrop -->
		<div
			class="absolute inset-0 bg-black/50"
			onclick={onClose}
		></div>

		<!-- Dialog -->
		<div class="relative bg-white rounded-2xl shadow-2xl w-full max-w-lg mx-4 overflow-hidden">
			<!-- Header -->
			<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200">
				<h2 class="text-xl font-semibold text-gray-900">Share Module</h2>
				<button
					onclick={onClose}
					class="p-2 hover:bg-gray-100 rounded-lg transition-colors"
				>
					<X class="w-5 h-5 text-gray-500" />
				</button>
			</div>

			<!-- Content -->
			<div class="px-6 py-5 space-y-5">
				<!-- Module Name -->
				<div>
					<p class="text-sm text-gray-600">
						Sharing: <span class="font-medium text-gray-900">{moduleName}</span>
					</p>
				</div>

				<!-- Share Type Tabs -->
				<div class="flex gap-2 p-1 bg-gray-100 rounded-lg">
					<button
						type="button"
						onclick={() => shareType = 'user'}
						class="flex-1 flex items-center justify-center gap-2 px-4 py-2 rounded-md transition-colors {shareType === 'user' ? 'bg-white text-gray-900 shadow-sm' : 'text-gray-600 hover:text-gray-900'}"
					>
						<User class="w-4 h-4" />
						<span class="text-sm font-medium">User</span>
					</button>
					<button
						type="button"
						onclick={() => shareType = 'workspace'}
						class="flex-1 flex items-center justify-center gap-2 px-4 py-2 rounded-md transition-colors {shareType === 'workspace' ? 'bg-white text-gray-900 shadow-sm' : 'text-gray-600 hover:text-gray-900'}"
					>
						<Users class="w-4 h-4" />
						<span class="text-sm font-medium">Workspace</span>
					</button>
				</div>

				<!-- User Input (if share type is user) -->
				{#if shareType === 'user'}
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-2">
							User ID or Email <span class="text-red-500">*</span>
						</label>
						<input
							type="text"
							bind:value={userId}
							placeholder="user@example.com"
							class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
						/>
					</div>
				{:else}
					<div class="bg-blue-50 border border-blue-200 rounded-lg p-4">
						<p class="text-sm text-blue-800">
							This module will be shared with your entire workspace.
						</p>
					</div>
				{/if}

				<!-- Permissions -->
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-3">
						Permissions <span class="text-red-500">*</span>
					</label>
					<div class="space-y-2">
						{#each allPermissions as perm}
							<button
								type="button"
								onclick={() => togglePermission(perm.value)}
								class="w-full flex items-start gap-3 p-3 border rounded-lg transition-all {permissions.includes(perm.value) ? 'border-blue-500 bg-blue-50' : 'border-gray-200 hover:border-gray-300'}"
							>
								<div class="flex-shrink-0 w-5 h-5 border-2 rounded flex items-center justify-center {permissions.includes(perm.value) ? 'border-blue-500 bg-blue-500' : 'border-gray-300'}">
									{#if permissions.includes(perm.value)}
										<Check class="w-3.5 h-3.5 text-white" />
									{/if}
								</div>
								<div class="flex-1 text-left">
									<p class="text-sm font-medium text-gray-900">{perm.label}</p>
									<p class="text-xs text-gray-600">{perm.description}</p>
								</div>
							</button>
						{/each}
					</div>
				</div>
			</div>

			<!-- Footer -->
			<div class="flex items-center justify-end gap-3 px-6 py-4 border-t border-gray-200 bg-gray-50">
				<button
					type="button"
					onclick={onClose}
					class="px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-100 rounded-lg transition-colors"
					disabled={isSharing}
				>
					Cancel
				</button>
				<button
					type="button"
					onclick={handleShare}
					disabled={isSharing}
					class="px-4 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
				>
					{isSharing ? 'Sharing...' : 'Share Module'}
				</button>
			</div>
		</div>
	</div>
{/if}
