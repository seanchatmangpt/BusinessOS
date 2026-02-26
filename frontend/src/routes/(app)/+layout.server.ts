import { redirect } from '@sveltejs/kit';
import type { LayoutServerLoad } from './$types';

// Backend URL based on environment
const isDev = process.env.NODE_ENV !== 'production';
const BACKEND_URL = isDev
	? 'http://localhost:8001'
	: 'https://businessos-api-460433387676.us-central1.run.app';

export const load: LayoutServerLoad = async ({ cookies, request, url }) => {
	// Check if this is an embedded iframe (e.g., chat in desktop window)
	const isEmbed = url.searchParams.get('embed') === 'true';

	// Get session cookie (set by backend after auth)
	// Backend uses 'better-auth.session_token' as cookie name
	const sessionCookie = cookies.get('better-auth.session_token');

	if (!sessionCookie) {
		// If embedded, return null user (parent window handles auth)
		if (isEmbed) {
			console.log('[Auth] Embedded context detected - skipping auth check');
			return {
				user: null,
				session: null,
				isEmbed: true
			};
		}
		// No session cookie - redirect to login
		throw redirect(303, '/login');
	}

	// Verify session with backend
	try {
		const response = await fetch(`${BACKEND_URL}/api/auth/session`, {
			method: 'GET',
			headers: {
				'Cookie': `better-auth.session_token=${sessionCookie}`,
				'User-Agent': request.headers.get('user-agent') || 'BusinessOS/1.0'
			},
			credentials: 'include'
		});

		if (!response.ok) {
			// Session invalid or expired
			console.warn(`[Auth] Session validation failed: ${response.status}`);
			if (isEmbed) {
				return { user: null, session: null, isEmbed: true };
			}
			throw redirect(303, '/login');
		}

		const sessionData = await response.json();

		if (!sessionData?.user) {
			// No user data in response
			console.warn('[Auth] No user data in session response');
			if (isEmbed) {
				return { user: null, session: null, isEmbed: true };
			}
			throw redirect(303, '/login');
		}

		// Return session data to all child routes
		return {
			user: sessionData.user,
			session: sessionData.session || { id: sessionCookie }
		};
	} catch (error) {
		// Network error or invalid response
		if (error instanceof Response) {
			// This is a redirect - let it propagate
			throw error;
		}

		console.error('[Auth] Session verification error:', error);
		if (isEmbed) {
			return { user: null, session: null, isEmbed: true };
		}
		throw redirect(303, '/login');
	}
};
