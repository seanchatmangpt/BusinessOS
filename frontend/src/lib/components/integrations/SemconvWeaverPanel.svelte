<script lang="ts">
	import { onMount } from 'svelte';
	import { slide } from 'svelte/transition';
	import * as aiApi from '$lib/api/ai';
	import type { Tool } from '$lib/api/ai/types';

	let loading = $state(true);
	let loadError = $state<string | null>(null);
	let tools = $state<Tool[]>([]);
	let expanded = $state<string | null>(null);
	let argsJson = $state<Record<string, string>>({});
	let resultText = $state<Record<string, string>>({});
	let running = $state<string | null>(null);

	onMount(async () => {
		try {
			const res = await aiApi.getTools();
			const list = res.tools || [];
			tools = list.filter(
				(t) => t.source === 'weaver-mcp' || t.name.startsWith('semconv.')
			);
			const next: Record<string, string> = {};
			for (const t of tools) {
				next[t.name] = argsJson[t.name] ?? '{}';
			}
			argsJson = next;
		} catch (e) {
			loadError = e instanceof Error ? e.message : 'Failed to load MCP tools';
		} finally {
			loading = false;
		}
	});

	function paramsFor(t: Tool): Record<string, unknown> | undefined {
		return t.parameters ?? t.input_schema;
	}

	async function runTool(name: string) {
		running = name;
		resultText = { ...resultText, [name]: '' };
		try {
			const raw = (argsJson[name] ?? '{}').trim();
			let args: Record<string, unknown> = {};
			if (raw) {
				args = JSON.parse(raw) as Record<string, unknown>;
			}
			const out = await aiApi.executeTool(name, args);
			if (out.success) {
				const r = out.result;
				resultText[name] =
					typeof r === 'string' ? r : JSON.stringify(r, null, 2);
			} else {
				resultText[name] = `Error: ${out.error}`;
			}
		} catch (e) {
			resultText[name] = e instanceof Error ? e.message : String(e);
		} finally {
			running = null;
		}
	}
</script>

<div class="sw-panel ih-card">
	<div class="sw-panel__head">
		<div>
			<h3 class="sw-panel__title">Semantic conventions (Weaver)</h3>
			<p class="sw-panel__sub">
				Built-in <code class="sw-code">weaver registry mcp</code> over your ChatmanGPT
				<code class="sw-code">semconv/model</code>. Tools are prefixed <code class="sw-code">semconv.</code> and are
				available to the AI stack via <code class="sw-code">GET /api/mcp/tools</code>.
			</p>
		</div>
		<span class="sw-badge">stdio MCP</span>
	</div>

	{#if loading}
		<p class="sw-muted">Loading Weaver tools…</p>
	{:else if loadError}
		<div class="ih-alert ih-alert--error ih-alert--sm"><p>{loadError}</p></div>
	{:else if tools.length === 0}
		<div class="sw-empty">
			<p class="sw-muted">
				No Weaver semconv tools found. Enable the backend with
				<code class="sw-code">WEAVER_SEMCONV_ENABLED=1</code>, mount
				<code class="sw-code">semconv/model</code>, and ensure <code class="sw-code">weaver</code> is on
				<code class="sw-code">PATH</code> (Docker image includes it).
			</p>
		</div>
	{:else}
		<p class="sw-count">{tools.length} tool{tools.length === 1 ? '' : 's'} from registry</p>
		<ul class="sw-list">
			{#each tools as t (t.name)}
				<li class="sw-item">
					<button
						type="button"
						class="sw-item__toggle"
						onclick={() => (expanded = expanded === t.name ? null : t.name)}
					>
						<span class="sw-item__name">{t.name}</span>
						<svg
							class="sw-chevron {expanded === t.name ? 'sw-chevron--open' : ''}"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
							><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" /></svg
						>
					</button>
					{#if t.description}
						<p class="sw-item__desc">{t.description}</p>
					{/if}
					{#if expanded === t.name}
						<div class="sw-detail" transition:slide={{ duration: 150 }}>
							{#if paramsFor(t)}
								<details class="sw-schema">
									<summary>Input schema (JSON Schema)</summary>
									<pre class="sw-pre">{JSON.stringify(paramsFor(t), null, 2)}</pre>
								</details>
							{/if}
							<label class="sw-label">
								Arguments (JSON object)
								<textarea
									class="sw-textarea"
									rows="4"
									value={argsJson[t.name] ?? '{}'}
									spellcheck="false"
									oninput={(e) => {
										argsJson = { ...argsJson, [t.name]: e.currentTarget.value };
									}}
								></textarea>
							</label>
							<button
								type="button"
								class="btn-pill btn-pill-primary btn-pill-sm sw-run"
								disabled={running === t.name}
								onclick={() => runTool(t.name)}
							>
								{running === t.name ? 'Running…' : 'Run tool'}
							</button>
							{#if resultText[t.name]}
								<div class="sw-result">
									<span class="sw-result__label">Result</span>
									<pre class="sw-pre sw-pre--result">{resultText[t.name]}</pre>
								</div>
							{/if}
						</div>
					{/if}
				</li>
			{/each}
		</ul>
	{/if}
</div>

<style>
	.sw-panel {
		margin-bottom: 1.5rem;
		padding: 1.25rem 1.5rem;
		border-radius: 12px;
		border: 1px solid rgba(0, 0, 0, 0.08);
	}
	:global(.dark) .sw-panel {
		border-color: rgba(255, 255, 255, 0.1);
	}
	.sw-panel__head {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 1rem;
		margin-bottom: 1rem;
	}
	.sw-panel__title {
		font-size: 1.05rem;
		font-weight: 600;
		margin: 0 0 0.35rem;
	}
	.sw-panel__sub {
		margin: 0;
		font-size: 0.85rem;
		line-height: 1.5;
		color: var(--color-text-muted, #64748b);
		max-width: 52rem;
	}
	.sw-code {
		font-size: 0.8em;
		padding: 0.1em 0.35em;
		border-radius: 4px;
		background: rgba(0, 0, 0, 0.06);
	}
	:global(.dark) .sw-code {
		background: rgba(255, 255, 255, 0.08);
	}
	.sw-badge {
		flex-shrink: 0;
		font-size: 0.7rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.04em;
		padding: 0.35rem 0.6rem;
		border-radius: 999px;
		background: rgba(59, 130, 246, 0.12);
		color: rgb(37, 99, 235);
	}
	:global(.dark) .sw-badge {
		background: rgba(59, 130, 246, 0.2);
		color: rgb(147, 197, 253);
	}
	.sw-muted {
		margin: 0;
		font-size: 0.9rem;
		color: var(--color-text-muted, #64748b);
	}
	.sw-count {
		margin: 0 0 0.75rem;
		font-size: 0.8rem;
		font-weight: 500;
		color: var(--color-text-muted, #64748b);
	}
	.sw-list {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}
	.sw-item {
		border: 1px solid rgba(0, 0, 0, 0.08);
		border-radius: 10px;
		padding: 0.65rem 0.85rem;
	}
	:global(.dark) .sw-item {
		border-color: rgba(255, 255, 255, 0.1);
	}
	.sw-item__toggle {
		display: flex;
		width: 100%;
		align-items: center;
		justify-content: space-between;
		gap: 0.5rem;
		background: none;
		border: none;
		padding: 0;
		cursor: pointer;
		text-align: left;
		font: inherit;
		color: inherit;
	}
	.sw-item__name {
		font-family: ui-monospace, monospace;
		font-size: 0.85rem;
		font-weight: 600;
	}
	.sw-item__desc {
		margin: 0.35rem 0 0;
		font-size: 0.8rem;
		color: var(--color-text-muted, #64748b);
		line-height: 1.4;
	}
	.sw-chevron {
		width: 1.1rem;
		height: 1.1rem;
		flex-shrink: 0;
		transition: transform 0.15s ease;
	}
	.sw-chevron--open {
		transform: rotate(180deg);
	}
	.sw-detail {
		margin-top: 0.85rem;
		padding-top: 0.85rem;
		border-top: 1px solid rgba(0, 0, 0, 0.06);
	}
	:global(.dark) .sw-detail {
		border-top-color: rgba(255, 255, 255, 0.08);
	}
	.sw-schema {
		margin-bottom: 0.75rem;
		font-size: 0.8rem;
	}
	.sw-label {
		display: flex;
		flex-direction: column;
		gap: 0.35rem;
		font-size: 0.8rem;
		font-weight: 500;
		margin-bottom: 0.65rem;
	}
	.sw-textarea {
		font-family: ui-monospace, monospace;
		font-size: 0.8rem;
		padding: 0.5rem 0.65rem;
		border-radius: 8px;
		border: 1px solid rgba(0, 0, 0, 0.12);
		background: transparent;
		resize: vertical;
		min-height: 5rem;
	}
	:global(.dark) .sw-textarea {
		border-color: rgba(255, 255, 255, 0.15);
	}
	.sw-run {
		margin-bottom: 0.75rem;
	}
	.sw-result__label {
		display: block;
		font-size: 0.7rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.04em;
		color: var(--color-text-muted, #64748b);
		margin-bottom: 0.35rem;
	}
	.sw-pre {
		margin: 0;
		padding: 0.65rem 0.75rem;
		border-radius: 8px;
		font-size: 0.75rem;
		line-height: 1.45;
		overflow: auto;
		max-height: 16rem;
		background: rgba(0, 0, 0, 0.04);
	}
	:global(.dark) .sw-pre {
		background: rgba(255, 255, 255, 0.06);
	}
	.sw-pre--result {
		max-height: 20rem;
	}
	.sw-empty {
		padding: 0.25rem 0;
	}
</style>
