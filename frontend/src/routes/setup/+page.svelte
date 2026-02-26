<script lang="ts">
	import { goto } from '$app/navigation';
	import { fly, fade } from 'svelte/transition';
	import { onMount } from 'svelte';
	import { getBackendUrl } from '$lib/api/base';
	import { AuthLayout, FormInput, PasswordInput } from '$lib/components/auth';

	// ── State ────────────────────────────────────────────────────────────────
	type Screen = 'choose' | 'team-setup' | 'done';
	let screen = $state<Screen>('choose');
	let loading = $state(false);
	let error = $state('');

	// Team-setup form fields
	let adminName     = $state('');
	let adminEmail    = $state('');
	let adminPassword = $state('');

	// Auth mode returned by the server
	let authMode = $state<string | null>(null);
	let oauthProviders = $state({ google: false, github: false });

	onMount(async () => {
		// Fetch the current auth mode from the server.
		try {
			const resp = await fetch(`${getBackendUrl()}/api/auth/mode`, { credentials: 'include' });
			if (resp.ok) {
				const data = await resp.json();
				authMode = data.mode;
				oauthProviders = data.oauth_providers ?? { google: false, github: false };

				// If setup is already done, redirect to login.
				if (!data.needs_setup) {
					goto('/login');
					return;
				}
			}
		} catch {
			// Server may not be reachable yet — stay on page.
		}
	});

	// ── Handlers ─────────────────────────────────────────────────────────────

	async function chooseSingleUser() {
		// Single-user mode: post to the backend to confirm the mode switch,
		// then redirect to the dashboard (the owner session is auto-injected).
		loading = true;
		try {
			await fetch(`${getBackendUrl()}/api/auth/setup/single`, {
				method: 'POST',
				credentials: 'include',
			});
			goto('/dashboard');
		} catch {
			// Even if the request fails the backend defaults to single mode,
			// so navigating to the dashboard is safe.
			goto('/dashboard');
		} finally {
			loading = false;
		}
	}

	function chooseTeamMode() {
		screen = 'team-setup';
	}

	async function submitTeamSetup(e: Event) {
		e.preventDefault();
		error = '';
		loading = true;

		try {
			const resp = await fetch(`${getBackendUrl()}/api/auth/setup`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				credentials: 'include',
				body: JSON.stringify({
					name: adminName,
					email: adminEmail,
					password: adminPassword,
				}),
			});

			const data = await resp.json();

			if (!resp.ok) {
				error = data.error || 'Setup failed';
				loading = false;
				return;
			}

			screen = 'done';
			setTimeout(() => goto('/dashboard'), 1500);
		} catch (err) {
			error = (err as Error).message || 'Network error';
			loading = false;
		}
	}
</script>

<AuthLayout>
	{#if screen === 'choose'}
		<div in:fly={{ y: 20, duration: 400 }}>
			<!-- Header -->
			<div class="mb-10">
				<h1 class="text-2xl font-bold text-gray-900 mb-2 font-mono tracking-tight">
					Welcome to BusinessOS
				</h1>
				<p class="text-gray-500 text-sm font-mono">
					How will you be using this?
				</p>
			</div>

			<!-- Mode selection cards -->
			<div class="space-y-4">
				<!-- Just me -->
				<button
					type="button"
					onclick={chooseSingleUser}
					disabled={loading}
					class="w-full text-left p-5 rounded-xl border-2 border-gray-100 hover:border-black hover:bg-gray-50 transition-all group"
				>
					<div class="flex items-start gap-4">
						<div class="w-10 h-10 bg-black text-white rounded-lg flex items-center justify-center flex-shrink-0 group-hover:scale-105 transition-transform">
							<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
							</svg>
						</div>
						<div class="flex-1">
							<div class="flex items-center justify-between mb-1">
								<h2 class="font-semibold text-gray-900 font-mono">Just me</h2>
								<span class="text-xs text-gray-400 font-mono bg-gray-100 px-2 py-0.5 rounded">Recommended</span>
							</div>
							<p class="text-sm text-gray-500 font-mono leading-relaxed">
								No login required. Go straight to your dashboard.
								Perfect for personal use or trying BusinessOS out.
							</p>
							<p class="text-xs text-gray-400 font-mono mt-2 flex items-center gap-1.5">
								<span class="w-1.5 h-1.5 bg-green-500 rounded-full"></span>
								Zero friction — always-on, no passwords
							</p>
						</div>
					</div>
				</button>

				<!-- My team -->
				<button
					type="button"
					onclick={chooseTeamMode}
					disabled={loading}
					class="w-full text-left p-5 rounded-xl border-2 border-gray-100 hover:border-black hover:bg-gray-50 transition-all group"
				>
					<div class="flex items-start gap-4">
						<div class="w-10 h-10 bg-gray-900 text-white rounded-lg flex items-center justify-center flex-shrink-0 group-hover:scale-105 transition-transform">
							<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
							</svg>
						</div>
						<div class="flex-1">
							<h2 class="font-semibold text-gray-900 font-mono mb-1">My team</h2>
							<p class="text-sm text-gray-500 font-mono leading-relaxed">
								Create accounts for multiple people.
								Invite teammates by email with role-based access.
							</p>
							<p class="text-xs text-gray-400 font-mono mt-2 flex items-center gap-1.5">
								<span class="w-1.5 h-1.5 bg-blue-500 rounded-full"></span>
								Works offline — no cloud services required
							</p>
						</div>
					</div>
				</button>
			</div>

			<!-- Advanced accordion (OAuth note) -->
			<details class="mt-6 group">
				<summary class="text-xs text-gray-400 hover:text-gray-600 font-mono cursor-pointer flex items-center gap-1.5 select-none list-none">
					<svg class="w-3.5 h-3.5 transition-transform group-open:rotate-90" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
					</svg>
					Advanced: OAuth configuration
				</summary>
				<div class="mt-3 pl-5 text-xs text-gray-400 font-mono space-y-1.5 leading-relaxed">
					<p>
						To enable Google or GitHub login, set <code class="bg-gray-100 px-1 rounded">AUTH_MODE=local+oauth</code>
						in your <code class="bg-gray-100 px-1 rounded">.env</code> file along with
						<code class="bg-gray-100 px-1 rounded">GOOGLE_CLIENT_ID</code> /
						<code class="bg-gray-100 px-1 rounded">GITHUB_CLIENT_ID</code> credentials,
						then restart the server.
					</p>
					<p class="text-gray-300">
						OAuth can also be configured later from Settings.
					</p>
				</div>
			</details>
		</div>

	{:else if screen === 'team-setup'}
		<div in:fly={{ y: 20, duration: 400 }}>
			<!-- Back -->
			<button
				type="button"
				onclick={() => { screen = 'choose'; error = ''; }}
				class="inline-flex items-center gap-1.5 text-xs text-gray-400 hover:text-gray-700 font-mono mb-8 transition-colors"
			>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
				</svg>
				Back
			</button>

			<div class="mb-8">
				<h1 class="text-2xl font-bold text-gray-900 mb-2 font-mono tracking-tight">Create your admin account</h1>
				<p class="text-gray-500 text-sm font-mono">
					This becomes the owner of your BusinessOS instance.
					You can invite teammates after setup.
				</p>
			</div>

			<form onsubmit={submitTeamSetup} class="space-y-5">
				{#if error}
					<div class="bg-red-50 border border-red-100 rounded-lg px-4 py-3 flex items-center gap-3" in:fly={{ y: -10, duration: 200 }}>
						<svg class="w-4 h-4 text-red-500 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
						<p class="text-sm text-red-600 font-mono">{error}</p>
					</div>
				{/if}

				<FormInput
					id="admin-name"
					label="Your name"
					type="text"
					bind:value={adminName}
					placeholder="Alex Johnson"
					autocomplete="name"
					required
				/>

				<FormInput
					id="admin-email"
					label="Email"
					type="email"
					bind:value={adminEmail}
					placeholder="you@company.com"
					autocomplete="email"
					required
				/>

				<PasswordInput
					id="admin-password"
					label="Password"
					bind:value={adminPassword}
					autocomplete="new-password"
					required
					showStrength
				/>

				<div class="bg-gray-50 rounded-lg px-4 py-3 text-xs text-gray-500 font-mono space-y-1">
					<p class="flex items-center gap-1.5">
						<span class="w-1.5 h-1.5 bg-green-500 rounded-full"></span>
						8+ characters, mixed case, number, special character
					</p>
				</div>

				<button
					type="submit"
					disabled={loading}
					class="w-full h-12 bg-black text-white rounded-lg text-sm font-medium hover:bg-gray-800 transition-all flex items-center justify-center gap-2 font-mono"
				>
					{#if loading}
						<svg class="animate-spin h-4 w-4" viewBox="0 0 24 24">
							<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" fill="none" />
							<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
						</svg>
						Initializing...
					{:else}
						Complete setup
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14 5l7 7m0 0l-7 7m7-7H3" />
						</svg>
					{/if}
				</button>
			</form>
		</div>

	{:else if screen === 'done'}
		<div in:fade={{ duration: 400 }} class="text-center py-12">
			<div class="w-16 h-16 bg-black rounded-full flex items-center justify-center mx-auto mb-6">
				<svg class="w-8 h-8 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
				</svg>
			</div>
			<h2 class="text-xl font-bold text-gray-900 font-mono mb-2">Setup complete</h2>
			<p class="text-sm text-gray-500 font-mono">Taking you to your dashboard...</p>
		</div>
	{/if}
</AuthLayout>
