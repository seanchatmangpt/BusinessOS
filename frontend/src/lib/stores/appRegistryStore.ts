import { writable } from 'svelte/store';

// Global store for App Registry modal visibility
// This allows both MenuBar and Desktop3D to control the same modal
export const showAppRegistry = writable(false);

export function openAppRegistry() {
	console.log('[appRegistryStore] Opening App Registry');
	showAppRegistry.set(true);
}

export function closeAppRegistry() {
	console.log('[appRegistryStore] Closing App Registry');
	showAppRegistry.set(false);
}
