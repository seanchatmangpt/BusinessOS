// Server hooks for SvelteKit
// Auth is now handled by the Go backend, so we just pass through requests

export async function handle({ event, resolve }) {
	return resolve(event);
}
