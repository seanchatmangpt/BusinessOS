/**
 * Dashboard Data Store
 * Owns all API-fetched data for the dashboard: loading state, error state,
 * focus items, projects, tasks, activities, and energy level.
 * Provides mutation handlers that keep local state in sync after API calls.
 *
 * Singleton factory pattern — matches chatUIStore.svelte.ts conventions.
 */

import { api } from "$lib/api";
import type { PetriNetJson } from "$lib/api/pm4py";
import type {
  FocusItem,
  DashboardProjectRow,
  DashboardTaskRow,
  DashboardActivityRow,
  ProcessMiningKPIData,
} from "./types";

const SAMPLE_EVENT_LOG = {
  traces: [
    {
      case_id: "case_001",
      events: [
        { activity: "Submit Request", timestamp: "2024-01-01T09:00:00Z" },
        { activity: "Review Request", timestamp: "2024-01-01T10:00:00Z" },
        { activity: "Approve", timestamp: "2024-01-01T11:00:00Z" },
        { activity: "Close", timestamp: "2024-01-01T12:00:00Z" },
      ],
    },
    {
      case_id: "case_002",
      events: [
        { activity: "Submit Request", timestamp: "2024-01-02T09:00:00Z" },
        { activity: "Review Request", timestamp: "2024-01-02T10:00:00Z" },
        { activity: "Reject", timestamp: "2024-01-02T11:00:00Z" },
        { activity: "Close", timestamp: "2024-01-02T12:00:00Z" },
      ],
    },
    {
      case_id: "case_003",
      events: [
        { activity: "Submit Request", timestamp: "2024-01-03T09:00:00Z" },
        { activity: "Review Request", timestamp: "2024-01-03T10:30:00Z" },
        { activity: "Request Info", timestamp: "2024-01-03T11:00:00Z" },
        { activity: "Review Request", timestamp: "2024-01-03T14:00:00Z" },
        { activity: "Approve", timestamp: "2024-01-03T15:00:00Z" },
        { activity: "Close", timestamp: "2024-01-03T16:00:00Z" },
      ],
    },
  ],
};

function createDashboardDataStore() {
  // ── Loading / error ──────────────────────────────────────────────────────────
  let isLoading = $state(true);
  let error = $state<string | null>(null);

  // ── API data ─────────────────────────────────────────────────────────────────
  let energyLevel = $state<number | null>(null);
  let focusItems = $state<FocusItem[]>([]);
  let projects = $state<DashboardProjectRow[]>([]);
  let tasks = $state<DashboardTaskRow[]>([]);
  let activities = $state<DashboardActivityRow[]>([]);

  // ── Process Mining KPI ───────────────────────────────────────────────────────
  let processMiningKPI = $state<ProcessMiningKPIData | null>(null);
  let isProcessMiningKPILoading = $state(false);
  let discoveredPetriNet = $state<PetriNetJson | null>(null);

  // ── Load ─────────────────────────────────────────────────────────────────────

  async function loadDashboard(): Promise<void> {
    try {
      isLoading = true;
      error = null;

      const summary = await api.getDashboardSummary();

      focusItems = summary.focus_items.map((item) => ({
        id: item.id,
        text: item.text,
        completed: item.completed,
      }));

      projects = summary.projects.map((p) => ({
        id: p.id,
        name: p.name,
        clientName: p.client_name ?? undefined,
        projectType: p.project_type,
        dueDate: p.due_date ?? undefined,
        progress: p.progress,
        health: p.health,
        teamCount: p.team_count,
      }));

      tasks = summary.tasks.map((t) => ({
        id: t.id,
        title: t.title,
        projectName: t.project_name ?? undefined,
        dueDate: t.due_date ?? undefined,
        priority: t.priority,
        completed: t.completed,
      }));

      activities = summary.activities.map((a) => ({
        id: a.id,
        type: a.type,
        description: a.description,
        actorName: a.actor_name ?? undefined,
        actorAvatar: a.actor_avatar ?? undefined,
        targetId: a.target_id ?? undefined,
        targetType: a.target_type ?? undefined,
        createdAt: a.created_at,
      }));

      energyLevel = summary.energy_level;
    } catch (err) {
      console.error("Failed to load dashboard:", err);
      error = err instanceof Error ? err.message : "Failed to load dashboard";
    } finally {
      isLoading = false;
    }
  }

  // ── Focus item handlers ───────────────────────────────────────────────────────

  async function handleFocusToggle(id: string): Promise<void> {
    const item = focusItems.find((i) => i.id === id);
    if (!item) return;
    try {
      await api.updateFocusItem(id, { completed: !item.completed });
      focusItems = focusItems.map((i) =>
        i.id === id ? { ...i, completed: !i.completed } : i,
      );
    } catch (err) {
      console.error("Failed to toggle focus item:", err);
    }
  }

  async function handleFocusAdd(text: string): Promise<void> {
    try {
      const newItem = await api.createFocusItem(text);
      focusItems = [
        ...focusItems,
        { id: newItem.id, text: newItem.text, completed: newItem.completed },
      ];
    } catch (err) {
      console.error("Failed to add focus item:", err);
    }
  }

  async function handleFocusRemove(id: string): Promise<void> {
    try {
      await api.deleteFocusItem(id);
      focusItems = focusItems.filter((item) => item.id !== id);
    } catch (err) {
      console.error("Failed to remove focus item:", err);
    }
  }

  function handleFocusEdit(): void {
    // TODO: Implement focus edit mode
    if (import.meta.env.DEV) console.log("Edit focus items");
  }

  // ── Task handlers ─────────────────────────────────────────────────────────────

  async function handleTaskToggle(id: string): Promise<void> {
    try {
      await api.toggleTask(id);
      tasks = tasks.map((task) =>
        task.id === id ? { ...task, completed: !task.completed } : task,
      );
    } catch (err) {
      console.error("Failed to toggle task:", err);
    }
  }

  // ── Process Mining KPI fetch ─────────────────────────────────────────────────

  async function loadProcessMiningKPI(eventLog?: unknown): Promise<void> {
    isProcessMiningKPILoading = true;
    try {
      const resp = await fetch("/api/pm4py/dashboard-kpi", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ event_log: eventLog ?? { traces: [] } }),
      });
      if (resp.ok) {
        // Map snake_case Go response to camelCase TypeScript types
        const raw = await resp.json();
        processMiningKPI = {
          conformanceFitness: raw.conformance_fitness ?? raw.conformanceFitness ?? 0,
          conformancePrecision: raw.conformance_precision ?? raw.conformancePrecision ?? 0,
          isConformant: raw.is_conformant ?? raw.isConformant ?? false,
          variantCount: raw.variant_count ?? raw.variantCount ?? 0,
          topVariants: raw.top_variants ?? raw.topVariants ?? [],
          bottleneckActivities: raw.bottleneck_activities ?? raw.bottleneckActivities ?? [],
          activityFrequencies: raw.activity_frequencies ?? raw.activityFrequencies ?? {},
          eventCount: raw.event_count ?? raw.eventCount ?? 0,
          traceCount: raw.trace_count ?? raw.traceCount ?? 0,
          fetchedAt: raw.fetched_at ?? raw.fetchedAt ?? new Date().toISOString(),
        } satisfies ProcessMiningKPIData;
      }
    } catch {
      // Silently fail — pm4py not running is expected in dev
    } finally {
      isProcessMiningKPILoading = false;
    }
  }

  async function discoverProcess(eventLog?: unknown): Promise<void> {
    const logToDiscover = eventLog ?? SAMPLE_EVENT_LOG;
    try {
      const resp = await fetch("/api/bos/discover", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ event_log: logToDiscover }),
      });
      if (resp.ok) {
        discoveredPetriNet = (await resp.json()) as PetriNetJson;
      }
    } catch {
      // Silently fail — pm4py not running is expected in dev
    }
  }

  // ── Energy ────────────────────────────────────────────────────────────────────

  function handleEnergySet(level: number): void {
    energyLevel = level;
    // TODO: Save to backend
  }

  // ── Public API ────────────────────────────────────────────────────────────────

  return {
    get isLoading() {
      return isLoading;
    },
    set isLoading(v: boolean) {
      isLoading = v;
    },

    get error() {
      return error;
    },
    set error(v: string | null) {
      error = v;
    },

    get energyLevel() {
      return energyLevel;
    },
    set energyLevel(v: number | null) {
      energyLevel = v;
    },

    get focusItems() {
      return focusItems;
    },
    set focusItems(v: FocusItem[]) {
      focusItems = v;
    },

    get projects() {
      return projects;
    },
    set projects(v: DashboardProjectRow[]) {
      projects = v;
    },

    get tasks() {
      return tasks;
    },
    set tasks(v: DashboardTaskRow[]) {
      tasks = v;
    },

    get activities() {
      return activities;
    },
    set activities(v: DashboardActivityRow[]) {
      activities = v;
    },

    get processMiningKPI() {
      return processMiningKPI;
    },
    set processMiningKPI(v: ProcessMiningKPIData | null) {
      processMiningKPI = v;
    },

    get isProcessMiningKPILoading() {
      return isProcessMiningKPILoading;
    },
    set isProcessMiningKPILoading(v: boolean) {
      isProcessMiningKPILoading = v;
    },

    get discoveredPetriNet() {
      return discoveredPetriNet;
    },
    set discoveredPetriNet(v: PetriNetJson | null) {
      discoveredPetriNet = v;
    },

    loadDashboard,
    handleFocusToggle,
    handleFocusAdd,
    handleFocusRemove,
    handleFocusEdit,
    handleTaskToggle,
    handleEnergySet,
    loadProcessMiningKPI,
    discoverProcess,
  };
}

export const dashboardDataStore = createDashboardDataStore();
