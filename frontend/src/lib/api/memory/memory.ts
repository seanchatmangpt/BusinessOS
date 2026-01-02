import { request } from '../base';
import type {
  Memory,
  MemoryListItem,
  CreateMemoryData,
  UpdateMemoryData,
  MemoryFilters,
  MemorySearchParams,
  MemorySearchResult,
  RelevantMemoryParams,
  MemoryStats,
  UserFact,
  UpdateUserFactData
} from './types';

// ============================================
// Memory CRUD Operations
// ============================================

export async function getMemories(filters?: MemoryFilters): Promise<MemoryListItem[]> {
  const params = new URLSearchParams();
  if (filters?.memory_type) params.set('memory_type', filters.memory_type);
  if (filters?.project_id) params.set('project_id', filters.project_id);
  if (filters?.node_id) params.set('node_id', filters.node_id);
  if (filters?.is_pinned !== undefined) params.set('is_pinned', String(filters.is_pinned));
  if (filters?.is_active !== undefined) params.set('is_active', String(filters.is_active));
  if (filters?.min_importance !== undefined) params.set('min_importance', String(filters.min_importance));
  if (filters?.tags?.length) params.set('tags', filters.tags.join(','));
  if (filters?.limit) params.set('limit', String(filters.limit));
  if (filters?.offset) params.set('offset', String(filters.offset));

  const query = params.toString();
  return request<MemoryListItem[]>(`/memories${query ? `?${query}` : ''}`);
}

export async function getMemory(id: string): Promise<Memory> {
  return request<Memory>(`/memories/${id}`);
}

export async function createMemory(data: CreateMemoryData): Promise<Memory> {
  return request<Memory>('/memories', { method: 'POST', body: data });
}

export async function updateMemory(id: string, data: UpdateMemoryData): Promise<Memory> {
  return request<Memory>(`/memories/${id}`, { method: 'PUT', body: data });
}

export async function deleteMemory(id: string): Promise<void> {
  return request(`/memories/${id}`, { method: 'DELETE' });
}

export async function pinMemory(id: string, pinned: boolean): Promise<Memory> {
  return request<Memory>(`/memories/${id}/pin`, { method: 'POST', body: { is_pinned: pinned } });
}

// ============================================
// Memory Search Operations
// ============================================

export async function searchMemories(params: MemorySearchParams): Promise<MemorySearchResult[]> {
  return request<MemorySearchResult[]>('/memories/search', { method: 'POST', body: params });
}

export async function getRelevantMemories(params: RelevantMemoryParams): Promise<MemorySearchResult[]> {
  return request<MemorySearchResult[]>('/memories/relevant', { method: 'POST', body: params });
}

// ============================================
// Memory Scoped Operations
// ============================================

export async function getProjectMemories(projectId: string, limit?: number): Promise<MemoryListItem[]> {
  const params = new URLSearchParams();
  if (limit) params.set('limit', String(limit));
  const query = params.toString();
  return request<MemoryListItem[]>(`/memories/project/${projectId}${query ? `?${query}` : ''}`);
}

export async function getNodeMemories(nodeId: string, limit?: number): Promise<MemoryListItem[]> {
  const params = new URLSearchParams();
  if (limit) params.set('limit', String(limit));
  const query = params.toString();
  return request<MemoryListItem[]>(`/memories/node/${nodeId}${query ? `?${query}` : ''}`);
}

// ============================================
// Memory Stats
// ============================================

export async function getMemoryStats(): Promise<MemoryStats> {
  return request<MemoryStats>('/memories/stats');
}

// ============================================
// User Facts Operations
// ============================================

export async function getUserFacts(): Promise<UserFact[]> {
  return request<UserFact[]>('/user-facts');
}

export async function updateUserFact(key: string, data: UpdateUserFactData): Promise<UserFact> {
  return request<UserFact>(`/user-facts/${encodeURIComponent(key)}`, { method: 'PUT', body: data });
}

export async function deleteUserFact(key: string): Promise<void> {
  return request(`/user-facts/${encodeURIComponent(key)}`, { method: 'DELETE' });
}
