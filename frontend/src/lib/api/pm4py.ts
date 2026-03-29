import { request } from "./base";

// ── Petri Net Types ───────────────────────────────────────────────────────────

export interface PetriNetPlace {
  id: string;
  name: string;
  initial_marking: number;
}

export interface PetriNetTransition {
  id: string;
  name: string;
  /** label may be absent for invisible transitions */
  label: string;
}

export interface PetriNetArc {
  from: string;
  to: string;
  weight: number;
}

export interface PetriNetJson {
  places: PetriNetPlace[];
  transitions: PetriNetTransition[];
  arcs: PetriNetArc[];
  initial_place: string;
  final_place: string;
}

// ── Response Types ────────────────────────────────────────────────────────────

export interface DiscoveryResponse {
  petri_net: PetriNetJson;
  algorithm: string;
  execution_time_ms: number;
  event_count: number;
  trace_count: number;
}

export interface StatisticsResponse {
  activity_frequencies: Record<string, number>;
  bottleneck_activities: string[];
  variant_frequencies: Record<string, number>;
  trace_count: number;
  event_count: number;
  variant_count: number;
}

// ── Dashboard KPI Types ───────────────────────────────────────────────────────

export interface VariantEntry {
  label: string;
  count: number;
  percentage: number;
}

export interface BottleneckEntry {
  activity: string;
  frequency: number;
}

/**
 * Raw (snake_case) response from POST /api/pm4py/dashboard-kpi.
 * The Go handler serialises with json struct tags; all fields are snake_case.
 */
export interface DashboardKPIResponse {
  conformance_fitness: number;
  conformance_precision: number;
  is_conformant: boolean;
  variant_count: number;
  top_variants: VariantEntry[];
  bottleneck_activities: BottleneckEntry[];
  activity_frequencies: Record<string, number>;
  event_count: number;
  trace_count: number;
  fetched_at: string;
}

// ── Default empty event log ───────────────────────────────────────────────────

/**
 * A minimal well-formed event log.
 * Use this when no real data is loaded so the dashboard renders without errors.
 */
export const EMPTY_EVENT_LOG = { traces: [] as unknown[] };

// ── API Functions ─────────────────────────────────────────────────────────────

/**
 * Discover a process map (Petri net) from an event log.
 * POST /pm4py/discover
 *
 * Calls the PM4PyRustHandler which proxies to pm4py-rust /api/discovery/alpha.
 * Accepts { event_log, variant } and returns DiscoveryResponse with petri_net.
 */
export async function discoverProcessMap(
  eventLog: unknown,
  variant = "alpha",
): Promise<DiscoveryResponse> {
  return request<DiscoveryResponse>("/pm4py/discover", {
    method: "POST",
    body: { event_log: eventLog, variant },
  });
}

/**
 * Get statistics for an event log.
 * POST /pm4py/statistics
 *
 * Calls the PM4PyRustHandler which proxies to pm4py-rust /api/statistics.
 */
export async function getStatistics(
  eventLog: unknown,
  options?: Record<string, boolean>,
): Promise<StatisticsResponse> {
  return request<StatisticsResponse>("/pm4py/statistics", {
    method: "POST",
    body: { event_log: eventLog, ...(options ?? {}) },
  });
}

/**
 * Aggregate process mining KPIs for the dashboard widgets.
 * POST /pm4py/dashboard-kpi
 *
 * Calls PM4PyDashboardHandler which fans out to statistics + conformance
 * concurrently and returns a merged KPI payload.
 */
export async function getDashboardKPI(
  eventLog: unknown,
  petriNet?: PetriNetJson,
): Promise<DashboardKPIResponse> {
  const body: Record<string, unknown> = { event_log: eventLog };
  if (petriNet) {
    body.petri_net = petriNet;
  }
  return request<DashboardKPIResponse>("/pm4py/dashboard-kpi", {
    method: "POST",
    body,
  });
}
