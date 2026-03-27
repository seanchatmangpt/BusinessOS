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

// ── API Functions ─────────────────────────────────────────────────────────────

/**
 * Discover a process map (Petri net) from an event log.
 * POST /bos/discover
 *
 * Proxied through BusinessOS gateway to pm4py-rust on port 8090.
 */
export async function discoverProcessMap(
  eventLog: unknown,
): Promise<DiscoveryResponse> {
  return request<DiscoveryResponse>("/bos/discover", {
    method: "POST",
    body: eventLog,
  });
}

/**
 * Get statistics for an event log.
 * POST /bos/statistics
 *
 * Proxied through BusinessOS gateway to pm4py-rust on port 8090.
 */
export async function getStatistics(
  eventLog: unknown,
  options?: Record<string, boolean>,
): Promise<StatisticsResponse> {
  return request<StatisticsResponse>("/bos/statistics", {
    method: "POST",
    body: { event_log: eventLog, ...(options ?? {}) },
  });
}
