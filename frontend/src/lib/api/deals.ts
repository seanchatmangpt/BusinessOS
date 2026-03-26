import { writable } from 'svelte/store';

/**
 * FIBO Deal types matching Go backend
 */

export type DealStatus = 'draft' | 'pending' | 'active' | 'closed';
export type ComplianceStatus = 'pass' | 'fail' | 'pending';
export type DealDomain = 'Finance' | 'Other';

export interface Deal {
  id: string;
  name: string;
  amount: number;
  currency: string;
  status: DealStatus;
  buyerId: string;
  sellerId: string;
  expectedCloseDate: string;
  probability: number;
  stage: string;
  createdAt: string;
  updatedAt: string;
  rdfTripleCount: number;
  complianceStatus: ComplianceStatus;
  kycVerified: boolean;
  amlScreening: string;
  domain?: DealDomain;
}

export interface CreateDealRequest {
  name: string;
  amount: number;
  currency: string;
  buyerId: string;
  sellerId: string;
  expectedCloseDate?: string;
  probability?: number;
  stage?: string;
  domain?: DealDomain;
}

export interface UpdateDealRequest {
  name?: string;
  amount?: number;
  currency?: string;
  status?: DealStatus;
  expectedCloseDate?: string;
  probability?: number;
  stage?: string;
  domain?: DealDomain;
}

export interface DealResponse {
  data?: Deal;
  error?: string;
  message?: string;
}

export interface DealListResponse {
  data?: Deal[];
  pagination?: {
    total: number;
    limit: number;
    offset: number;
  };
  error?: string;
}

const API_TIMEOUT_MS = 10000;

/**
 * Make an API request with timeout
 */
async function apiRequest<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> {
  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), API_TIMEOUT_MS);

  try {
    const response = await fetch(`/api${endpoint}`, {
      ...options,
      signal: controller.signal,
      headers: {
        'Content-Type': 'application/json',
        ...options.headers
      }
    });

    clearTimeout(timeoutId);

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || `HTTP ${response.status}`);
    }

    return response.json();
  } catch (error) {
    clearTimeout(timeoutId);
    if (error instanceof Error) {
      if (error.name === 'AbortError') {
        throw new Error(`Request timeout after ${API_TIMEOUT_MS}ms`);
      }
      throw error;
    }
    throw new Error('Unknown error');
  }
}

/**
 * List all deals with pagination and filtering
 */
export async function listDeals(
  limit: number = 20,
  offset: number = 0,
  statusFilter?: DealStatus,
  domainFilter?: DealDomain
): Promise<Deal[]> {
  const params = new URLSearchParams();
  params.set('limit', limit.toString());
  params.set('offset', offset.toString());

  if (statusFilter) {
    params.set('status', statusFilter);
  }

  if (domainFilter) {
    params.set('domain', domainFilter);
  }

  const response = await apiRequest<DealListResponse>(
    `/deals?${params.toString()}`
  );
  return response.data || [];
}

/**
 * Get a single deal by ID
 */
export async function getDeal(dealId: string): Promise<Deal> {
  const response = await apiRequest<DealResponse>(`/deals/${dealId}`);
  if (!response.data) {
    throw new Error('Deal not found');
  }
  return response.data;
}

/**
 * Create a new deal
 */
export async function createDeal(request: CreateDealRequest): Promise<Deal> {
  const response = await apiRequest<DealResponse>('/deals', {
    method: 'POST',
    body: JSON.stringify(request)
  });
  if (!response.data) {
    throw new Error(response.error || 'Failed to create deal');
  }
  return response.data;
}

/**
 * Update an existing deal
 */
export async function updateDeal(
  dealId: string,
  request: UpdateDealRequest
): Promise<Deal> {
  const response = await apiRequest<DealResponse>(`/deals/${dealId}`, {
    method: 'PATCH',
    body: JSON.stringify(request)
  });
  if (!response.data) {
    throw new Error(response.error || 'Failed to update deal');
  }
  return response.data;
}

/**
 * Delete a deal
 */
export async function deleteDeal(dealId: string): Promise<void> {
  await apiRequest(`/deals/${dealId}`, {
    method: 'DELETE'
  });
}

/**
 * Verify deal compliance
 */
export async function verifyCompliance(dealId: string): Promise<Deal> {
  const response = await apiRequest<DealResponse>(
    `/deals/${dealId}/verify-compliance`,
    {
      method: 'POST'
    }
  );
  if (!response.data) {
    throw new Error(response.error || 'Compliance verification failed');
  }
  return response.data;
}

/**
 * Svelte store for deals list
 */
export const dealsStore = writable<{
  deals: Deal[];
  loading: boolean;
  error: string | null;
  total: number;
  limit: number;
  offset: number;
}>({
  deals: [],
  loading: false,
  error: null,
  total: 0,
  limit: 20,
  offset: 0
});

/**
 * Load deals into store
 */
export async function loadDeals(
  limit: number = 20,
  offset: number = 0,
  statusFilter?: DealStatus,
  domainFilter?: DealDomain
) {
  dealsStore.update((s) => ({ ...s, loading: true, error: null }));

  try {
    const deals = await listDeals(limit, offset, statusFilter, domainFilter);
    dealsStore.update((s) => ({
      ...s,
      deals,
      loading: false,
      limit,
      offset,
      total: deals.length
    }));
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Unknown error';
    dealsStore.update((s) => ({ ...s, loading: false, error: message }));
    throw error;
  }
}
