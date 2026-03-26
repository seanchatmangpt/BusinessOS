<script lang="ts">
	import { goto } from '$app/navigation';
	import { createDeal, type CreateDealRequest, type DealDomain } from '$lib/api/deals';

	let formData = $state<CreateDealRequest>({
		name: '',
		amount: 0,
		currency: 'USD',
		buyerId: '',
		sellerId: '',
		domain: 'Finance'
	});

	let errors = $state<Record<string, string>>({});
	let isSubmitting = $state(false);
	let globalError = $state<string | null>(null);

	function validateForm(): boolean {
		errors = {};

		if (!formData.name?.trim()) {
			errors.name = 'Deal name is required';
		}

		if (!formData.amount || formData.amount <= 0) {
			errors.amount = 'Amount must be greater than 0';
		}

		if (!formData.buyerId?.trim()) {
			errors.buyerId = 'Buyer ID is required';
		}

		if (!formData.sellerId?.trim()) {
			errors.sellerId = 'Seller ID is required';
		}

		return Object.keys(errors).length === 0;
	}

	async function handleSubmit(e: SubmitEvent) {
		e.preventDefault();
		globalError = null;

		if (!validateForm()) {
			return;
		}

		isSubmitting = true;

		try {
			const deal = await createDeal(formData);
			// Show success toast and navigate
			goto(`/deals/${deal.id}`);
		} catch (err) {
			globalError = err instanceof Error ? err.message : 'Failed to create deal';
		} finally {
			isSubmitting = false;
		}
	}

	function handleInputChange(field: string, value: any) {
		formData[field as keyof CreateDealRequest] = value;
		if (errors[field]) {
			delete errors[field];
		}
	}

	const domains: DealDomain[] = ['Finance', 'Other'];
</script>

<div class="create-deal-page">
	<div class="page-header">
		<button class="btn-back" onclick={() => goto('/deals')} title="Back to deals">
			<svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
			</svg>
		</button>
		<div>
			<h1 class="page-title">Create New Deal</h1>
			<p class="page-subtitle">Fill in the details to create a new FIBO deal</p>
		</div>
	</div>

	<div class="form-container">
		{#if globalError}
			<div class="error-banner">
				<p class="error-text">{globalError}</p>
				<button
					class="error-dismiss"
					onclick={() => (globalError = null)}
					aria-label="Dismiss error"
				>
					×
				</button>
			</div>
		{/if}

		<form onsubmit={handleSubmit}>
			<!-- Deal Name -->
			<div class="form-group">
				<label for="name" class="form-label">Deal Name *</label>
				<input
					id="name"
					type="text"
					class="form-input"
					class:error={!!errors.name}
					placeholder="e.g., Acme Corp Acquisition"
					value={formData.name}
					onchange={(e) => handleInputChange('name', e.currentTarget.value)}
					disabled={isSubmitting}
				/>
				{#if errors.name}
					<p class="form-error">{errors.name}</p>
				{/if}
			</div>

			<!-- Amount and Currency Row -->
			<div class="form-row">
				<div class="form-group">
					<label for="amount" class="form-label">Amount *</label>
					<input
						id="amount"
						type="number"
						step="0.01"
						min="0"
						class="form-input"
						class:error={!!errors.amount}
						placeholder="0.00"
						value={formData.amount || ''}
						onchange={(e) => handleInputChange('amount', parseFloat(e.currentTarget.value))}
						disabled={isSubmitting}
					/>
					{#if errors.amount}
						<p class="form-error">{errors.amount}</p>
					{/if}
				</div>

				<div class="form-group">
					<label for="currency" class="form-label">Currency</label>
					<select
						id="currency"
						class="form-input"
						value={formData.currency}
						onchange={(e) => handleInputChange('currency', e.currentTarget.value)}
						disabled={isSubmitting}
					>
						<option value="USD">USD</option>
						<option value="EUR">EUR</option>
						<option value="GBP">GBP</option>
						<option value="JPY">JPY</option>
						<option value="CAD">CAD</option>
					</select>
				</div>
			</div>

			<!-- Buyer and Seller IDs Row -->
			<div class="form-row">
				<div class="form-group">
					<label for="buyerId" class="form-label">Buyer ID *</label>
					<input
						id="buyerId"
						type="text"
						class="form-input"
						class:error={!!errors.buyerId}
						placeholder="e.g., BUYER-001"
						value={formData.buyerId}
						onchange={(e) => handleInputChange('buyerId', e.currentTarget.value)}
						disabled={isSubmitting}
					/>
					{#if errors.buyerId}
						<p class="form-error">{errors.buyerId}</p>
					{/if}
				</div>

				<div class="form-group">
					<label for="sellerId" class="form-label">Seller ID *</label>
					<input
						id="sellerId"
						type="text"
						class="form-input"
						class:error={!!errors.sellerId}
						placeholder="e.g., SELLER-001"
						value={formData.sellerId}
						onchange={(e) => handleInputChange('sellerId', e.currentTarget.value)}
						disabled={isSubmitting}
					/>
					{#if errors.sellerId}
						<p class="form-error">{errors.sellerId}</p>
					{/if}
				</div>
			</div>

			<!-- Expected Close Date -->
			<div class="form-group">
				<label for="closeDate" class="form-label">Expected Close Date</label>
				<input
					id="closeDate"
					type="date"
					class="form-input"
					value={formData.expectedCloseDate}
					onchange={(e) => handleInputChange('expectedCloseDate', e.currentTarget.value)}
					disabled={isSubmitting}
				/>
			</div>

			<!-- Probability -->
			<div class="form-group">
				<label for="probability" class="form-label">Probability</label>
				<div class="probability-input">
					<input
						id="probability"
						type="range"
						min="0"
						max="100"
						step="1"
						value={formData.probability || 50}
						onchange={(e) => handleInputChange('probability', parseInt(e.currentTarget.value))}
						disabled={isSubmitting}
					/>
					<span class="probability-value">{formData.probability || 50}%</span>
				</div>
			</div>

			<!-- Domain -->
			<div class="form-group">
				<label for="domain" class="form-label">Domain</label>
				<select
					id="domain"
					class="form-input"
					value={formData.domain}
					onchange={(e) => handleInputChange('domain', e.currentTarget.value as DealDomain)}
					disabled={isSubmitting}
				>
					{#each domains as domain (domain)}
						<option value={domain}>{domain}</option>
					{/each}
				</select>
			</div>

			<!-- Stage -->
			<div class="form-group">
				<label for="stage" class="form-label">Stage</label>
				<select
					id="stage"
					class="form-input"
					value={formData.stage || ''}
					onchange={(e) => handleInputChange('stage', e.currentTarget.value)}
					disabled={isSubmitting}
				>
					<option value="">Select a stage</option>
					<option value="prospecting">Prospecting</option>
					<option value="qualification">Qualification</option>
					<option value="proposal">Proposal</option>
					<option value="negotiation">Negotiation</option>
					<option value="closing">Closing</option>
				</select>
			</div>

			<!-- Form Actions -->
			<div class="form-actions">
				<button
					type="button"
					class="btn-secondary"
					onclick={() => goto('/deals')}
					disabled={isSubmitting}
				>
					Cancel
				</button>
				<button type="submit" class="btn-primary" disabled={isSubmitting}>
					{#if isSubmitting}
						<svg class="spinner" viewBox="0 0 24 24" fill="none" stroke="currentColor">
							<circle cx="12" cy="12" r="10" stroke-width="2" />
							<path d="M12 2a10 10 0 0110 10" stroke-width="2" stroke-linecap="round" />
						</svg>
						Creating...
					{:else}
						<svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor">
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M12 4v16m8-8H4"
							/>
						</svg>
						Create Deal
					{/if}
				</button>
			</div>
		</form>

		<p class="form-help">
			<span class="required-indicator">*</span> Indicates required field
		</p>
	</div>
</div>

<style>
	.create-deal-page {
		display: flex;
		flex-direction: column;
		height: 100%;
		background: var(--dbg, #fff);
	}

	.page-header {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 20px 24px;
		border-bottom: 1px solid var(--dbd, #e0e0e0);
	}

	.btn-back {
		background: none;
		border: none;
		color: var(--dt, #111);
		cursor: pointer;
		padding: 4px;
		transition: opacity 0.15s ease;
	}

	.btn-back:hover {
		opacity: 0.6;
	}

	.icon {
		width: 16px;
		height: 16px;
	}

	.spinner {
		width: 16px;
		height: 16px;
		animation: spin 1s linear infinite;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}

	.page-title {
		font-size: 24px;
		font-weight: 700;
		color: var(--dt, #111);
		margin: 0;
	}

	.page-subtitle {
		font-size: 13px;
		color: var(--dt3, #888);
		margin: 4px 0 0 0;
	}

	.form-container {
		flex: 1;
		overflow-y: auto;
		padding: 24px;
		max-width: 600px;
	}

	.error-banner {
		margin-bottom: 20px;
		padding: 12px 16px;
		background: rgba(239, 68, 68, 0.06);
		border: 1px solid rgba(239, 68, 68, 0.2);
		border-radius: 8px;
		display: flex;
		align-items: center;
		justify-content: space-between;
	}

	.error-text {
		font-size: 13px;
		color: #ef4444;
		font-weight: 500;
		margin: 0;
	}

	.error-dismiss {
		background: none;
		border: none;
		color: #ef4444;
		cursor: pointer;
		font-size: 18px;
		padding: 0;
	}

	form {
		display: flex;
		flex-direction: column;
		gap: 20px;
	}

	.form-group {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}

	.form-row {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 16px;
	}

	.form-label {
		font-size: 13px;
		font-weight: 600;
		color: var(--dt, #111);
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.form-input,
	input[type='date'],
	input[type='range'],
	select {
		padding: 10px 12px;
		border: 1px solid var(--dbd, #e0e0e0);
		border-radius: 6px;
		font-size: 13px;
		color: var(--dt, #111);
		background: var(--dbg, #fff);
		transition: border-color 0.15s ease;
		font-family: inherit;
	}

	.form-input:focus,
	input[type='date']:focus,
	select:focus {
		outline: none;
		border-color: #6366f1;
		box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.1);
	}

	.form-input.error {
		border-color: #ef4444;
	}

	.form-input:disabled,
	input[type='date']:disabled,
	select:disabled {
		background: var(--dbg-secondary, #f5f5f5);
		opacity: 0.6;
		cursor: not-allowed;
	}

	.form-error {
		font-size: 12px;
		color: #ef4444;
		margin: 0;
		font-weight: 500;
	}

	.probability-input {
		display: flex;
		align-items: center;
		gap: 12px;
	}

	input[type='range'] {
		flex: 1;
		height: 6px;
		-webkit-appearance: none;
		appearance: none;
		background: linear-gradient(
			to right,
			#6366f1 0%,
			#6366f1 calc((var(--value, 50) * 1%) - 6px),
			var(--dbd, #e0e0e0) calc((var(--value, 50) * 1%) - 6px),
			var(--dbd, #e0e0e0) 100%
		);
		border-radius: 3px;
		padding: 0;
		border: none;
	}

	input[type='range']::-webkit-slider-thumb {
		-webkit-appearance: none;
		appearance: none;
		width: 18px;
		height: 18px;
		border-radius: 50%;
		background: #6366f1;
		cursor: pointer;
		border: 2px solid white;
		box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
	}

	input[type='range']::-moz-range-thumb {
		width: 18px;
		height: 18px;
		border-radius: 50%;
		background: #6366f1;
		cursor: pointer;
		border: 2px solid white;
		box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
	}

	.probability-value {
		font-size: 13px;
		font-weight: 600;
		color: #6366f1;
		min-width: 40px;
		text-align: right;
	}

	.form-actions {
		display: flex;
		gap: 12px;
		margin-top: 12px;
	}

	.btn-primary,
	.btn-secondary {
		flex: 1;
		padding: 12px 16px;
		border: none;
		border-radius: 6px;
		font-size: 13px;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.15s ease;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 6px;
	}

	.btn-primary {
		background: #6366f1;
		color: white;
	}

	.btn-primary:hover:not(:disabled) {
		background: #4f46e5;
		transform: translateY(-1px);
	}

	.btn-primary:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.btn-secondary {
		background: var(--dbg-secondary, #f5f5f5);
		color: var(--dt, #111);
		border: 1px solid var(--dbd, #e0e0e0);
	}

	.btn-secondary:hover:not(:disabled) {
		background: var(--dbg, #fff);
	}

	.btn-secondary:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.form-help {
		font-size: 12px;
		color: var(--dt3, #888);
		margin: 0;
		text-align: center;
	}

	.required-indicator {
		color: #ef4444;
		font-weight: 700;
	}

	@media (max-width: 640px) {
		.form-row {
			grid-template-columns: 1fr;
		}
	}
</style>
