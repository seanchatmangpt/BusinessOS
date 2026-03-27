<script lang="ts">
	import type { PetriNetJson } from '$lib/api/pm4py';

	// ── Props ─────────────────────────────────────────────────────────────────

	let {
		petriNet = null,
		activityFrequencies = {},
		bottleneckActivities = [],
		loading = false,
		error = null
	}: {
		petriNet: PetriNetJson | null;
		activityFrequencies?: Record<string, number>;
		bottleneckActivities?: string[];
		loading?: boolean;
		error?: string | null;
	} = $props();

	// ── Layout constants ──────────────────────────────────────────────────────

	const NODE_SPACING_X = 150;
	const NODE_SPACING_Y = 80;
	const PLACE_RADIUS = 20;
	const TRANSITION_W = 80;
	const TRANSITION_H = 40;

	// ── Node position type ────────────────────────────────────────────────────

	interface NodePos {
		x: number;
		y: number;
		type: 'place' | 'transition';
	}

	// ── Pan/zoom state ────────────────────────────────────────────────────────

	let pan = $state({ x: 40, y: 40, scale: 1 });
	let dragging = $state(false);
	let dragStart = $state({ x: 0, y: 0, panX: 0, panY: 0 });

	// ── Layout computation (pure $derived) ───────────────────────────────────

	const nodePositions = $derived(computeLayout(petriNet));

	function computeLayout(net: PetriNetJson | null): Map<string, NodePos> {
		const positions = new Map<string, NodePos>();
		if (!net) return positions;

		// Build set of all node ids
		const allNodes = new Set<string>();
		for (const p of net.places) allNodes.add(p.id);
		for (const t of net.transitions) allNodes.add(t.id);

		// Build adjacency list (directed: from → [to])
		const adj = new Map<string, string[]>();
		for (const id of allNodes) adj.set(id, []);
		for (const arc of net.arcs) {
			const targets = adj.get(arc.from);
			if (targets) targets.push(arc.to);
		}

		// BFS from initial_place to assign column depth
		const depth = new Map<string, number>();
		const queue: string[] = [net.initial_place];
		depth.set(net.initial_place, 0);

		while (queue.length > 0) {
			const current = queue.shift()!;
			const currentDepth = depth.get(current) ?? 0;
			const neighbors = adj.get(current) ?? [];
			for (const neighbor of neighbors) {
				if (!depth.has(neighbor)) {
					depth.set(neighbor, currentDepth + 1);
					queue.push(neighbor);
				}
			}
		}

		// Assign depth 0 to any disconnected nodes
		for (const id of allNodes) {
			if (!depth.has(id)) depth.set(id, 0);
		}

		// Group nodes by column
		const columns = new Map<number, string[]>();
		for (const [id, d] of depth.entries()) {
			const col = columns.get(d) ?? [];
			col.push(id);
			columns.set(d, col);
		}

		// Sort each column by incoming arc source column, then assign y positions
		for (const [col, nodeIds] of columns.entries()) {
			// Sort: nodes with earlier-column predecessors come first
			nodeIds.sort((a, b) => {
				const aMinPred = getMinPredecessorCol(a, net.arcs, depth);
				const bMinPred = getMinPredecessorCol(b, net.arcs, depth);
				return aMinPred - bMinPred;
			});

			const totalHeight = (nodeIds.length - 1) * NODE_SPACING_Y;
			const startY = totalHeight / 2;

			nodeIds.forEach((id, idx) => {
				const isPlace = net.places.some((p) => p.id === id);
				positions.set(id, {
					x: col * NODE_SPACING_X,
					y: -startY + idx * NODE_SPACING_Y,
					type: isPlace ? 'place' : 'transition'
				});
			});
		}

		return positions;
	}

	function getMinPredecessorCol(
		id: string,
		arcs: PetriNetJson['arcs'],
		depth: Map<string, number>
	): number {
		let min = Infinity;
		for (const arc of arcs) {
			if (arc.to === id) {
				const d = depth.get(arc.from) ?? 0;
				if (d < min) min = d;
			}
		}
		return min === Infinity ? 0 : min;
	}

	// ── SVG viewport dimensions ────────────────────────────────────────────────

	const svgBounds = $derived(computeBounds(nodePositions));

	function computeBounds(positions: Map<string, NodePos>): {
		minX: number;
		minY: number;
		width: number;
		height: number;
	} {
		if (positions.size === 0) return { minX: 0, minY: 0, width: 400, height: 200 };

		let minX = Infinity,
			minY = Infinity,
			maxX = -Infinity,
			maxY = -Infinity;

		for (const pos of positions.values()) {
			const hw = pos.type === 'place' ? PLACE_RADIUS : TRANSITION_W / 2;
			const hh = pos.type === 'place' ? PLACE_RADIUS : TRANSITION_H / 2;
			if (pos.x - hw < minX) minX = pos.x - hw;
			if (pos.y - hh < minY) minY = pos.y - hh;
			if (pos.x + hw > maxX) maxX = pos.x + hw;
			if (pos.y + hh > maxY) maxY = pos.y + hh;
		}

		return {
			minX: minX - 40,
			minY: minY - 40,
			width: maxX - minX + 80,
			height: maxY - minY + 80
		};
	}

	// ── Arc path computation ───────────────────────────────────────────────────

	function arcPath(
		from: string,
		to: string,
		positions: Map<string, NodePos>,
		net: PetriNetJson
	): string {
		const src = positions.get(from);
		const dst = positions.get(to);
		if (!src || !dst) return '';

		// Compute edge points on the boundary of each node
		const dx = dst.x - src.x;
		const dy = dst.y - src.y;
		const dist = Math.sqrt(dx * dx + dy * dy);
		if (dist === 0) return '';

		const ux = dx / dist;
		const uy = dy / dist;

		// Source offset: exit from boundary of source
		let sx: number, sy: number;
		if (src.type === 'place') {
			sx = src.x + ux * PLACE_RADIUS;
			sy = src.y + uy * PLACE_RADIUS;
		} else {
			// Transition: exit from rectangle edge
			const tw = TRANSITION_W / 2;
			const th = TRANSITION_H / 2;
			const scale = Math.min(Math.abs(tw / (ux || 0.001)), Math.abs(th / (uy || 0.001)));
			sx = src.x + ux * Math.min(scale, Math.sqrt(tw * tw + th * th));
			sy = src.y + uy * Math.min(scale, Math.sqrt(tw * tw + th * th));
		}

		// Destination offset: arrive at boundary of target (minus arrowhead length)
		let ex: number, ey: number;
		const arrowOffset = 10;
		if (dst.type === 'place') {
			ex = dst.x - ux * (PLACE_RADIUS + arrowOffset);
			ey = dst.y - uy * (PLACE_RADIUS + arrowOffset);
		} else {
			const tw = TRANSITION_W / 2;
			const th = TRANSITION_H / 2;
			const scale = Math.min(Math.abs(tw / (ux || 0.001)), Math.abs(th / (uy || 0.001)));
			const edgeDist = Math.min(scale, Math.sqrt(tw * tw + th * th));
			ex = dst.x - ux * (edgeDist + arrowOffset);
			ey = dst.y - uy * (edgeDist + arrowOffset);
		}

		// Use a simple quadratic curve with slight bend for parallel arcs
		const midX = (sx + ex) / 2;
		const midY = (sy + ey) / 2;

		void net; // net used in caller context only
		return `M ${sx} ${sy} Q ${midX} ${midY} ${ex} ${ey}`;
	}

	// ── Performance color (blue → red via HSL) ────────────────────────────────

	const maxFrequency = $derived(
		Object.values(activityFrequencies).reduce((a, b) => Math.max(a, b), 1)
	);

	function transitionFill(transition: PetriNetJson['transitions'][number]): string {
		const freq = activityFrequencies[transition.label] ?? activityFrequencies[transition.name] ?? 0;
		if (maxFrequency === 0 || freq === 0) return '#6366f1'; // indigo default
		const ratio = freq / maxFrequency;
		// HSL: 240 (blue) → 0 (red)
		const hue = Math.round(240 - ratio * 240);
		return `hsl(${hue}, 70%, 50%)`;
	}

	function isBottleneck(transition: PetriNetJson['transitions'][number]): boolean {
		return (
			bottleneckActivities.includes(transition.label) ||
			bottleneckActivities.includes(transition.name)
		);
	}

	function isInitialPlace(placeId: string): boolean {
		return petriNet?.initial_place === placeId;
	}

	function isFinalPlace(placeId: string): boolean {
		return petriNet?.final_place === placeId;
	}

	// ── Pan/zoom handlers ──────────────────────────────────────────────────────

	function onWheel(e: WheelEvent) {
		e.preventDefault();
		const factor = e.deltaY < 0 ? 1.1 : 0.9;
		pan.scale = Math.min(Math.max(pan.scale * factor, 0.2), 4);
	}

	function onPointerDown(e: PointerEvent) {
		dragging = true;
		dragStart = { x: e.clientX, y: e.clientY, panX: pan.x, panY: pan.y };
		(e.currentTarget as Element)?.setPointerCapture(e.pointerId);
	}

	function onPointerMove(e: PointerEvent) {
		if (!dragging) return;
		pan.x = dragStart.panX + (e.clientX - dragStart.x);
		pan.y = dragStart.panY + (e.clientY - dragStart.y);
	}

	function onPointerUp(e: PointerEvent) {
		dragging = false;
		(e.currentTarget as Element)?.releasePointerCapture(e.pointerId);
	}

	function resetView() {
		pan = { x: 40, y: 40, scale: 1 };
	}
</script>

<div class="dw-pmv-widget">
	<!-- Header -->
	<div class="flex items-center justify-between mb-4">
		<div class="flex items-center gap-2">
			<div class="dw-pmv-icon w-8 h-8 rounded-lg flex items-center justify-center shadow-sm">
				<svg class="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 3H5a2 2 0 00-2 2v4m6-6h10a2 2 0 012 2v4M9 3v18m0 0h10a2 2 0 002-2V9M9 21H5a2 2 0 01-2-2V9m0 0h18" />
				</svg>
			</div>
			<h2 class="dw-pmv-title text-base font-semibold">Process Map</h2>
		</div>

		{#if petriNet && !loading}
			<button
				onclick={resetView}
				class="dw-pmv-meta text-xs px-2 py-1 rounded hover:opacity-80 transition-opacity"
				title="Reset pan/zoom"
			>
				Reset View
			</button>
		{/if}
	</div>

	{#if error}
		<!-- Error state -->
		<div class="dw-pmv-error flex flex-col items-center justify-center gap-3 py-8">
			<svg class="w-8 h-8 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
			</svg>
			<p class="text-xs text-center">{error}</p>
		</div>
	{:else if loading}
		<!-- Skeleton -->
		<div class="space-y-3 animate-pulse">
			<div class="dw-pmv-skeleton h-4 rounded w-1/2"></div>
			<div class="dw-pmv-skeleton rounded-lg" style="height: 220px;"></div>
			<div class="flex gap-2">
				<div class="dw-pmv-skeleton h-3 rounded flex-1"></div>
				<div class="dw-pmv-skeleton h-3 rounded w-1/3"></div>
			</div>
		</div>
	{:else if !petriNet}
		<!-- Empty state -->
		<div class="flex flex-col items-center justify-center gap-3 py-8">
			<svg class="w-10 h-10 dw-pmv-meta opacity-40" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 17V7m0 10a2 2 0 01-2 2H5a2 2 0 01-2-2V7a2 2 0 012-2h2a2 2 0 012 2m0 10a2 2 0 002 2h2a2 2 0 002-2M9 7a2 2 0 012-2h2a2 2 0 012 2m0 10V7m0 10a2 2 0 002 2h2a2 2 0 002-2V7a2 2 0 00-2-2h-2a2 2 0 00-2 2" />
			</svg>
			<p class="dw-pmv-meta text-xs text-center">
				No process map loaded.<br />Run process discovery to visualize a Petri net.
			</p>
		</div>
	{:else}
		<!-- SVG Canvas -->
		<div
			class="dw-pmv-canvas rounded-lg overflow-hidden cursor-grab"
			class:cursor-grabbing={dragging}
			style="height: 260px; position: relative;"
			role="img"
			aria-label="Process map Petri net visualization"
			onwheel={onWheel}
			onpointerdown={onPointerDown}
			onpointermove={onPointerMove}
			onpointerup={onPointerUp}
		>
			<svg
				width="100%"
				height="100%"
				style="display: block;"
			>
				<defs>
					<marker
						id="arrowhead"
						markerWidth="10"
						markerHeight="14"
						refX="9"
						refY="7"
						orient="auto"
					>
						<path d="M 0 0 L 10 7 L 0 14 z" fill="var(--dw-pmv-arc, #6366f1)" />
					</marker>
					<marker
						id="arrowhead-bottleneck"
						markerWidth="10"
						markerHeight="14"
						refX="9"
						refY="7"
						orient="auto"
					>
						<path d="M 0 0 L 10 7 L 0 14 z" fill="#f59e0b" />
					</marker>
				</defs>

				<g transform="translate({pan.x},{pan.y}) scale({pan.scale})">
					<!-- Arcs (drawn first so nodes are on top) -->
					{#each petriNet.arcs as arc}
						{@const path = arcPath(arc.from, arc.to, nodePositions, petriNet)}
						{#if path}
							<path
								d={path}
								fill="none"
								stroke="var(--dw-pmv-arc, #6366f1)"
								stroke-width="1.5"
								stroke-opacity="0.7"
								marker-end="url(#arrowhead)"
							/>
						{/if}
					{/each}

					<!-- Places -->
					{#each petriNet.places as place}
						{@const pos = nodePositions.get(place.id)}
						{#if pos}
							<g transform="translate({pos.x},{pos.y})">
								<circle
									r={PLACE_RADIUS}
									fill={isInitialPlace(place.id) ? 'var(--dw-pmv-place-initial, #22c55e)' : isFinalPlace(place.id) ? 'var(--dw-pmv-place-final, #ef4444)' : 'var(--dw-pmv-place-bg, #fff)'}
									stroke="var(--dw-pmv-place-stroke, #6366f1)"
									stroke-width="2"
								/>
								{#if place.initial_marking > 0 && !isInitialPlace(place.id)}
									<!-- Token dot -->
									<circle r="5" fill="var(--dw-pmv-token, #6366f1)" />
								{/if}
								<title>{place.name}</title>
							</g>
						{/if}
					{/each}

					<!-- Transitions -->
					{#each petriNet.transitions as transition}
						{@const pos = nodePositions.get(transition.id)}
						{@const bneck = isBottleneck(transition)}
						{#if pos}
							<g transform="translate({pos.x},{pos.y})" class={bneck ? 'pmv-bottleneck' : ''}>
								<rect
									x={-TRANSITION_W / 2}
									y={-TRANSITION_H / 2}
									width={TRANSITION_W}
									height={TRANSITION_H}
									rx="4"
									fill={transitionFill(transition)}
									stroke={bneck ? '#f59e0b' : 'none'}
									stroke-width={bneck ? 2.5 : 0}
								/>
								<text
									text-anchor="middle"
									dominant-baseline="middle"
									fill="#fff"
									font-size="9"
									font-family="system-ui, sans-serif"
									pointer-events="none"
								>
									{(transition.label || transition.name).slice(0, 12)}
								</text>
								<title>{transition.label || transition.name}</title>
							</g>
						{/if}
					{/each}
				</g>
			</svg>
		</div>

		<!-- Legend -->
		<div class="flex items-center gap-4 mt-3 flex-wrap">
			<div class="flex items-center gap-1.5">
				<circle class="dw-pmv-legend-circle" style="background: var(--dw-pmv-place-initial, #22c55e);" />
				<span class="dw-pmv-meta text-[10px]">Start</span>
			</div>
			<div class="flex items-center gap-1.5">
				<circle class="dw-pmv-legend-circle" style="background: var(--dw-pmv-place-final, #ef4444);" />
				<span class="dw-pmv-meta text-[10px]">End</span>
			</div>
			<div class="flex items-center gap-1.5">
				<div class="dw-pmv-legend-rect" style="background: #6366f1;"></div>
				<span class="dw-pmv-meta text-[10px]">Activity</span>
			</div>
			{#if bottleneckActivities.length > 0}
				<div class="flex items-center gap-1.5">
					<div class="dw-pmv-legend-rect" style="background: #f59e0b;"></div>
					<span class="dw-pmv-meta text-[10px]">Bottleneck</span>
				</div>
			{/if}
			<div class="flex items-center gap-3 ml-auto">
				<span class="dw-pmv-meta text-[10px]">
					{petriNet.places.length} places · {petriNet.transitions.length} transitions
				</span>
			</div>
		</div>
	{/if}
</div>

<style>
	.dw-pmv-widget {
		background: var(--dbg, var(--bos-card, #fff));
		border: 1px solid var(--dbd, var(--bos-border, #e0e0e0));
		border-radius: 0.75rem;
		padding: 1.25rem;
		box-shadow: var(--bos-shadow-1, 0 1px 2px rgba(0, 0, 0, 0.05));
		transition: box-shadow 0.3s;
	}
	.dw-pmv-widget:hover {
		box-shadow: var(--bos-shadow-2, 0 4px 6px rgba(0, 0, 0, 0.07));
	}

	.dw-pmv-icon {
		background: linear-gradient(135deg, #6366f1, #4f46e5);
	}

	.dw-pmv-title {
		color: var(--dt, var(--bos-text-primary, #111));
	}

	.dw-pmv-meta {
		color: var(--dt3, var(--bos-text-tertiary, #888));
	}

	.dw-pmv-error {
		color: var(--dt2, var(--bos-text-secondary, #555));
	}

	.dw-pmv-skeleton {
		background: var(--dbg3, var(--bos-hover, #eee));
	}

	.dw-pmv-canvas {
		background: var(--dbg2, var(--bos-hover, #f5f5f5));
		border: 1px solid var(--dbd, var(--bos-border, #e0e0e0));
		--dw-pmv-arc: #6366f1;
		--dw-pmv-place-bg: #fff;
		--dw-pmv-place-stroke: #6366f1;
		--dw-pmv-place-initial: #22c55e;
		--dw-pmv-place-final: #ef4444;
		--dw-pmv-token: #6366f1;
	}

	.cursor-grabbing {
		cursor: grabbing !important;
	}

	/* Legend items */
	.dw-pmv-legend-circle {
		display: inline-block;
		width: 10px;
		height: 10px;
		border-radius: 50%;
		flex-shrink: 0;
	}

	.dw-pmv-legend-rect {
		display: inline-block;
		width: 12px;
		height: 8px;
		border-radius: 2px;
		flex-shrink: 0;
	}

	/* Bottleneck pulse animation */
	@keyframes pulse-stroke {
		0%, 100% { stroke-opacity: 1; }
		50% { stroke-opacity: 0.4; }
	}

	:global(.pmv-bottleneck rect) {
		animation: pulse-stroke 1.5s ease-in-out infinite;
	}
</style>
