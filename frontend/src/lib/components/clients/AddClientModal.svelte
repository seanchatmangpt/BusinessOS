<script lang="ts">
	import type { CreateClientData, ClientType, ClientStatus } from '$lib/api';
	import { fade, scale } from 'svelte/transition';

	interface Props {
		open: boolean;
		onCreate: (data: CreateClientData) => void;
	}

	let { open = $bindable(), onCreate }: Props = $props();

	// Form state
	let name = $state('');
	let type = $state<ClientType>('company');
	let email = $state('');
	let phone = $state('');
	let website = $state('');
	let industry = $state('');
	let companySize = $state('');
	let status = $state<ClientStatus>('lead');
	let source = $state('');
	let notes = $state('');
	let tagsInput = $state('');
	let submitting = $state(false);
	let errorMessage = $state<string | null>(null);

	// Address fields
	let address = $state('');
	let city = $state('');
	let stateRegion = $state('');
	let zipCode = $state('');
	let country = $state('');

	// Toggle for showing more fields
	let showMoreFields = $state(false);

	const companySizes = [
		{ value: '', label: 'Select size' },
		{ value: '1-10', label: '1-10 employees' },
		{ value: '11-50', label: '11-50 employees' },
		{ value: '51-200', label: '51-200 employees' },
		{ value: '201-500', label: '201-500 employees' },
		{ value: '501-1000', label: '501-1000 employees' },
		{ value: '1000+', label: '1000+ employees' }
	];

	const sources = [
		{ value: '', label: 'Select source' },
		{ value: 'website', label: 'Website' },
		{ value: 'referral', label: 'Referral' },
		{ value: 'cold-call', label: 'Cold Call' },
		{ value: 'linkedin', label: 'LinkedIn' },
		{ value: 'conference', label: 'Conference' },
		{ value: 'advertising', label: 'Advertising' },
		{ value: 'other', label: 'Other' }
	];

	function resetForm() {
		name = '';
		type = 'company';
		email = '';
		phone = '';
		website = '';
		industry = '';
		companySize = '';
		status = 'lead';
		source = '';
		notes = '';
		tagsInput = '';
		address = '';
		city = '';
		stateRegion = '';
		zipCode = '';
		country = '';
		showMoreFields = false;
		errorMessage = null;
	}

	function handleClose() {
		open = false;
		resetForm();
	}

	async function handleSubmit(e: Event) {
		e.preventDefault();
		if (!name.trim()) return;

		submitting = true;
		errorMessage = null;

		try {
			const tags = tagsInput
				.split(',')
				.map((t) => t.trim())
				.filter((t) => t);

			const data: CreateClientData = {
				name: name.trim(),
				type,
				email: email.trim() || undefined,
				phone: phone.trim() || undefined,
				website: website.trim() || undefined,
				industry: industry.trim() || undefined,
				company_size: companySize || undefined,
				status,
				source: source || undefined,
				notes: notes.trim() || undefined,
				tags: tags.length > 0 ? tags : undefined,
				address: address.trim() || undefined,
				city: city.trim() || undefined,
				state: stateRegion.trim() || undefined,
				zip_code: zipCode.trim() || undefined,
				country: country.trim() || undefined
			};

			await onCreate(data);
			handleClose();
		} catch (err) {
			errorMessage = err instanceof Error ? err.message : 'Failed to create client. Please try again.';
		} finally {
			submitting = false;
		}
	}
</script>

{#if open}
	<div
		class="cr-modal-overlay"
		onclick={handleClose}
		role="dialog"
		aria-modal="true"
		aria-label="Add new client"
		transition:fade={{ duration: 150 }}
	>
		<div
			class="cr-modal"
			onclick={(e) => e.stopPropagation()}
			role="document"
			transition:scale={{ duration: 200, start: 0.95 }}
		>
			<!-- Header -->
			<div class="cr-modal__header">
				<div>
					<h2 class="cr-modal__title">Add New Client</h2>
					<p class="cr-modal__subtitle">Create a new client record</p>
				</div>
				<button
					class="cr-modal__close"
					onclick={handleClose}
					aria-label="Close modal"
				>
					<svg width="18" height="18" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			</div>

			<!-- Form -->
			<form onsubmit={handleSubmit} class="cr-modal__body">
				{#if errorMessage}
					<div class="cr-modal__error">
						<p class="cr-modal__error-text">{errorMessage}</p>
					</div>
				{/if}

				<div class="cr-modal__fields">
					<!-- Type Toggle -->
					<div class="cr-modal__type-toggle">
						<button
							type="button"
							onclick={() => (type = 'company')}
							class="cr-modal__type-btn {type === 'company' ? 'cr-modal__type-btn--active' : ''}"
							aria-pressed={type === 'company'}
						>
							Company
						</button>
						<button
							type="button"
							onclick={() => (type = 'individual')}
							class="cr-modal__type-btn {type === 'individual' ? 'cr-modal__type-btn--active' : ''}"
							aria-pressed={type === 'individual'}
						>
							Individual
						</button>
					</div>

					<!-- Name -->
					<div class="cr-modal__field">
						<label for="client-name" class="cr-modal__label">
							{type === 'company' ? 'Company Name' : 'Full Name'} *
						</label>
						<input
							id="client-name"
							type="text"
							bind:value={name}
							required
							disabled={submitting}
							class="cr-modal__input"
							placeholder={type === 'company' ? 'Acme Inc.' : 'John Doe'}
						/>
					</div>

					<!-- Email & Phone -->
					<div class="cr-modal__row">
						<div class="cr-modal__field">
							<label for="client-email" class="cr-modal__label">Email</label>
							<input
								id="client-email"
								type="email"
								bind:value={email}
								disabled={submitting}
								class="cr-modal__input"
								placeholder="email@example.com"
							/>
						</div>
						<div class="cr-modal__field">
							<label for="client-phone" class="cr-modal__label">Phone</label>
							<input
								id="client-phone"
								type="tel"
								bind:value={phone}
								disabled={submitting}
								class="cr-modal__input"
								placeholder="+1 (555) 123-4567"
							/>
						</div>
					</div>

					<!-- Status & Source -->
					<div class="cr-modal__row">
						<div class="cr-modal__field">
							<label for="client-status" class="cr-modal__label">Status</label>
							<select
								id="client-status"
								bind:value={status}
								disabled={submitting}
								class="cr-modal__input"
							>
								<option value="lead">Lead</option>
								<option value="prospect">Prospect</option>
								<option value="active">Active</option>
								<option value="inactive">Inactive</option>
							</select>
						</div>
						<div class="cr-modal__field">
							<label for="client-source" class="cr-modal__label">Source</label>
							<select
								id="client-source"
								bind:value={source}
								disabled={submitting}
								class="cr-modal__input"
							>
								{#each sources as s}
									<option value={s.value}>{s.label}</option>
								{/each}
							</select>
						</div>
					</div>

					<!-- Company-specific fields -->
					{#if type === 'company'}
						<div class="cr-modal__row">
							<div class="cr-modal__field">
								<label for="client-industry" class="cr-modal__label">Industry</label>
								<input
									id="client-industry"
									type="text"
									bind:value={industry}
									disabled={submitting}
									class="cr-modal__input"
									placeholder="Technology"
								/>
							</div>
							<div class="cr-modal__field">
								<label for="client-size" class="cr-modal__label">Company Size</label>
								<select
									id="client-size"
									bind:value={companySize}
									disabled={submitting}
									class="cr-modal__input"
								>
									{#each companySizes as size}
										<option value={size.value}>{size.label}</option>
									{/each}
								</select>
							</div>
						</div>

						<div class="cr-modal__field">
							<label for="client-website" class="cr-modal__label">Website</label>
							<input
								id="client-website"
								type="url"
								bind:value={website}
								disabled={submitting}
								class="cr-modal__input"
								placeholder="https://example.com"
							/>
						</div>
					{/if}

					<!-- Tags -->
					<div class="cr-modal__field">
						<label for="client-tags" class="cr-modal__label">Tags</label>
						<input
							id="client-tags"
							type="text"
							bind:value={tagsInput}
							disabled={submitting}
							class="cr-modal__input"
							placeholder="vip, enterprise, tech (comma separated)"
						/>
					</div>

					<!-- Show More Toggle -->
					<button
						type="button"
						onclick={() => (showMoreFields = !showMoreFields)}
						class="cr-modal__toggle"
						aria-expanded={showMoreFields}
					>
						<svg
							class="cr-modal__toggle-icon {showMoreFields ? 'cr-modal__toggle-icon--open' : ''}"
							width="14"
							height="14"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
							aria-hidden="true"
						>
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
						</svg>
						{showMoreFields ? 'Show less' : 'Show more fields'}
					</button>

					{#if showMoreFields}
						<!-- Address -->
						<div class="cr-modal__field">
							<label for="client-address" class="cr-modal__label">Address</label>
							<input
								id="client-address"
								type="text"
								bind:value={address}
								disabled={submitting}
								class="cr-modal__input"
								placeholder="123 Main St"
							/>
						</div>

						<div class="cr-modal__row">
							<div class="cr-modal__field">
								<label for="client-city" class="cr-modal__label">City</label>
								<input
									id="client-city"
									type="text"
									bind:value={city}
									disabled={submitting}
									class="cr-modal__input"
									placeholder="San Francisco"
								/>
							</div>
							<div class="cr-modal__field">
								<label for="client-state" class="cr-modal__label">State/Region</label>
								<input
									id="client-state"
									type="text"
									bind:value={stateRegion}
									disabled={submitting}
									class="cr-modal__input"
									placeholder="CA"
								/>
							</div>
						</div>

						<div class="cr-modal__row">
							<div class="cr-modal__field">
								<label for="client-zip" class="cr-modal__label">ZIP/Postal Code</label>
								<input
									id="client-zip"
									type="text"
									bind:value={zipCode}
									disabled={submitting}
									class="cr-modal__input"
									placeholder="94102"
								/>
							</div>
							<div class="cr-modal__field">
								<label for="client-country" class="cr-modal__label">Country</label>
								<input
									id="client-country"
									type="text"
									bind:value={country}
									disabled={submitting}
									class="cr-modal__input"
									placeholder="United States"
								/>
							</div>
						</div>

						<!-- Notes -->
						<div class="cr-modal__field">
							<label for="client-notes" class="cr-modal__label">Notes</label>
							<textarea
								id="client-notes"
								bind:value={notes}
								rows="3"
								disabled={submitting}
								class="cr-modal__input cr-modal__textarea"
								placeholder="Additional notes about this client..."
							></textarea>
						</div>
					{/if}
				</div>

				<!-- Footer -->
				<div class="cr-modal__actions">
					<button
						type="button"
						onclick={handleClose}
						disabled={submitting}
						class="btn-rounded btn-rounded-ghost"
					>
						Cancel
					</button>
					<button
						type="submit"
						disabled={submitting || !name.trim()}
						class="btn-pill btn-pill-primary btn-pill-sm"
					>
						{submitting ? 'Creating...' : 'Create Client'}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}

<style>
	/* ─── Modal Overlay (Foundation CRM pattern, cr- prefix) ─── */
	.cr-modal-overlay {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.5);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 50;
	}
	.cr-modal {
		background: var(--dbg, #fff);
		border-radius: var(--radius-md, 12px);
		box-shadow: 0 16px 48px rgba(0, 0, 0, 0.15);
		width: 100%;
		max-width: 32rem;
		max-height: 90vh;
		overflow: hidden;
		display: flex;
		flex-direction: column;
		margin: 0 1rem;
	}
	:global(.dark) .cr-modal {
		box-shadow: 0 16px 48px rgba(0, 0, 0, 0.4);
	}

	/* Header */
	.cr-modal__header {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		padding: var(--space-4, 1rem) var(--space-6, 1.5rem);
		border-bottom: 1px solid var(--dbd, #e0e0e0);
	}
	.cr-modal__title {
		font-size: var(--text-lg, 1.125rem);
		font-weight: var(--font-semibold, 600);
		color: var(--dt, #111);
		margin: 0;
	}
	.cr-modal__subtitle {
		font-size: var(--text-sm, 0.875rem);
		color: var(--dt3, #888);
		margin: 2px 0 0;
	}
	.cr-modal__close {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 28px;
		height: 28px;
		border-radius: var(--radius-xs, 4px);
		border: none;
		background: transparent;
		color: var(--dt3, #888);
		cursor: pointer;
		transition: background 0.15s, color 0.15s;
	}
	.cr-modal__close:hover {
		background: var(--dbg2, #f5f5f5);
		color: var(--dt, #111);
	}

	/* Body */
	.cr-modal__body {
		flex: 1;
		overflow-y: auto;
		display: flex;
		flex-direction: column;
	}

	/* Error */
	.cr-modal__error {
		margin: var(--space-4, 1rem) var(--space-6, 1.5rem) 0;
		padding: var(--space-2, 0.5rem) var(--space-3, 0.75rem);
		border-radius: var(--radius-sm, 8px);
		background: color-mix(in srgb, #ef4444 10%, var(--dbg, #fff));
		border: 1px solid color-mix(in srgb, #ef4444 25%, var(--dbd, #e0e0e0));
	}
	.cr-modal__error-text {
		font-size: var(--text-sm, 0.875rem);
		color: #ef4444;
		margin: 0;
		font-weight: 500;
	}

	/* Fields */
	.cr-modal__fields {
		display: flex;
		flex-direction: column;
		gap: var(--space-4, 1rem);
		padding: var(--space-4, 1rem) var(--space-6, 1.5rem);
	}
	.cr-modal__field {
		display: flex;
		flex-direction: column;
	}
	.cr-modal__row {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: var(--space-4, 1rem);
	}
	.cr-modal__label {
		font-size: var(--text-sm, 0.875rem);
		font-weight: var(--font-medium, 500);
		color: var(--dt2, #555);
		margin-bottom: var(--space-1, 0.25rem);
	}
	.cr-modal__input {
		width: 100%;
		padding: var(--space-2, 0.5rem) var(--space-3, 0.75rem);
		border: 1px solid var(--dbd, #e0e0e0);
		border-radius: var(--radius-sm, 8px);
		font-size: var(--text-sm, 0.875rem);
		color: var(--dt, #111);
		background: var(--dbg, #fff);
		outline: none;
		transition: border-color 0.15s, box-shadow 0.15s;
		box-sizing: border-box;
	}
	.cr-modal__input:focus {
		border-color: var(--dt3, #888);
		box-shadow: 0 0 0 2px color-mix(in srgb, var(--dt3, #888) 25%, transparent);
	}
	.cr-modal__input::placeholder {
		color: var(--dt4, #bbb);
	}
	.cr-modal__input:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}
	.cr-modal__textarea {
		resize: none;
		min-height: 5rem;
	}

	/* Type Toggle */
	.cr-modal__type-toggle {
		display: flex;
		gap: var(--space-2, 0.5rem);
	}
	.cr-modal__type-btn {
		flex: 1;
		padding: var(--space-2, 0.5rem) var(--space-4, 1rem);
		font-size: var(--text-sm, 0.875rem);
		font-weight: var(--font-medium, 500);
		border-radius: var(--radius-sm, 8px);
		border: 1px solid var(--dbd, #e0e0e0);
		cursor: pointer;
		transition: background 0.15s, color 0.15s, border-color 0.15s;
		background: var(--dbg2, #f5f5f5);
		color: var(--dt2, #555);
	}
	.cr-modal__type-btn:hover {
		background: var(--dbg3, #eee);
		color: var(--dt, #111);
	}
	.cr-modal__type-btn--active {
		background: var(--dt, #111);
		color: var(--dbg, #fff);
		border-color: var(--dt, #111);
	}
	.cr-modal__type-btn--active:hover {
		background: var(--dt, #111);
		color: var(--dbg, #fff);
	}

	/* Toggle more fields */
	.cr-modal__toggle {
		display: flex;
		align-items: center;
		gap: var(--space-1, 0.25rem);
		font-size: var(--text-sm, 0.875rem);
		color: var(--dt3, #888);
		background: none;
		border: none;
		cursor: pointer;
		padding: 0;
		transition: color 0.15s;
	}
	.cr-modal__toggle:hover {
		color: var(--dt, #111);
	}
	.cr-modal__toggle-icon {
		transition: transform 0.2s;
	}
	.cr-modal__toggle-icon--open {
		transform: rotate(180deg);
	}

	/* Actions */
	.cr-modal__actions {
		display: flex;
		justify-content: flex-end;
		gap: var(--space-3, 0.75rem);
		padding: var(--space-4, 1rem) var(--space-6, 1.5rem);
		border-top: 1px solid var(--dbd, #e0e0e0);
	}
</style>
