import { auth } from '$lib/server/auth';
import { svelteKitHandler } from 'better-auth/svelte-kit';
import { building } from '$app/environment';

// In development: Frontend calls Go backend directly via CORS (faster)
// In production with adapter-node: Can optionally proxy through SvelteKit
// In production with static: Use nginx/Caddy reverse proxy

export async function handle({ event, resolve }) {
	// Better Auth handles /api/auth/* routes
	return svelteKitHandler({ event, resolve, auth, building });
}
