import { request, raw } from '../base';
import type {
  GoogleAuthResponse,
  GoogleConnectionStatus,
  SlackAuthResponse,
  SlackConnectionStatus,
  SlackChannelsResponse,
  SlackNotificationsResponse,
  NotionAuthResponse,
  NotionConnectionStatus,
  NotionDatabasesResponse,
  NotionPagesResponse,
  NotionSyncResponse,
  HubSpotAuthResponse,
  HubSpotConnectionStatus,
  GoHighLevelAuthResponse,
  GoHighLevelConnectionStatus,
  LinearAuthResponse,
  LinearConnectionStatus,
  AsanaAuthResponse,
  AsanaConnectionStatus,
  GenericAuthResponse,
  GenericConnectionStatus,
  AllIntegrationsStatus,
  IntegrationSyncResponse,
  IntegrationProvider,
  FileImportResponse,
  ImportProgress
} from './types';

// ============================================
// All Integrations Status
// ============================================

/**
 * Get status of all integrations at once
 */
export async function getAllIntegrationsStatus() {
  return request<AllIntegrationsStatus>('/integrations/status');
}

/**
 * Get status of a specific integration
 */
export async function getIntegrationStatus(provider: IntegrationProvider) {
  return request<GenericConnectionStatus>(`/integrations/${provider}/status`);
}

/**
 * Generic initiate auth for any OAuth provider
 */
export async function initiateAuth(provider: IntegrationProvider) {
  return request<GenericAuthResponse>(`/integrations/${provider}/auth`);
}

/**
 * Generic disconnect for any provider
 */
export async function disconnectIntegration(provider: IntegrationProvider) {
  return request(`/integrations/${provider}`, { method: 'DELETE' });
}

/**
 * Trigger sync for a provider
 */
export async function syncIntegration(provider: IntegrationProvider) {
  return request<IntegrationSyncResponse>(`/integrations/${provider}/sync`, { method: 'POST' });
}

// ============================================
// Google OAuth Integration
// ============================================
export async function initiateGoogleAuth() {
  return request<GoogleAuthResponse>('/integrations/google/auth');
}

export async function getGoogleConnectionStatus() {
  return request<GoogleConnectionStatus>('/integrations/google/status');
}

export async function disconnectGoogle() {
  return request('/integrations/google', { method: 'DELETE' });
}

// ============================================
// Slack Integration
// ============================================

/**
 * Initiate Slack OAuth flow
 * Returns URL to redirect user to for Slack authorization
 */
export async function initiateSlackAuth() {
  return request<SlackAuthResponse>('/integrations/slack/auth');
}

/**
 * Get current Slack connection status
 */
export async function getSlackConnectionStatus() {
  return request<SlackConnectionStatus>('/integrations/slack/status');
}

/**
 * Disconnect Slack integration
 */
export async function disconnectSlack() {
  return request('/integrations/slack', { method: 'DELETE' });
}

/**
 * Get list of Slack channels the user has access to
 */
export async function getSlackChannels() {
  return request<SlackChannelsResponse>('/integrations/slack/channels');
}

/**
 * Get Slack notifications/messages
 * @param limit - Number of notifications to fetch (default 50)
 * @param cursor - Pagination cursor for fetching more results
 */
export async function getSlackNotifications(limit = 50, cursor?: string) {
  const params = new URLSearchParams({ limit: String(limit) });
  if (cursor) params.append('cursor', cursor);
  return request<SlackNotificationsResponse>(`/integrations/slack/notifications?${params}`);
}

// ============================================
// Notion Integration
// ============================================

/**
 * Initiate Notion OAuth flow
 * Returns URL to redirect user to for Notion authorization
 */
export async function initiateNotionAuth() {
  return request<NotionAuthResponse>('/integrations/notion/auth');
}

/**
 * Get current Notion connection status
 */
export async function getNotionConnectionStatus() {
  return request<NotionConnectionStatus>('/integrations/notion/status');
}

/**
 * Disconnect Notion integration
 */
export async function disconnectNotion() {
  return request('/integrations/notion', { method: 'DELETE' });
}

/**
 * Get list of Notion databases the integration has access to
 */
export async function getNotionDatabases() {
  return request<NotionDatabasesResponse>('/integrations/notion/databases');
}

/**
 * Get Notion pages
 * @param databaseId - Optional database ID to filter pages
 * @param cursor - Pagination cursor for fetching more results
 */
export async function getNotionPages(databaseId?: string, cursor?: string) {
  const params = new URLSearchParams();
  if (databaseId) params.append('database_id', databaseId);
  if (cursor) params.append('cursor', cursor);
  const query = params.toString();
  return request<NotionPagesResponse>(`/integrations/notion/pages${query ? `?${query}` : ''}`);
}

/**
 * Sync a Notion database to BusinessOS
 * @param databaseId - The Notion database ID to sync
 */
export async function syncNotionDatabase(databaseId: string) {
  return request<NotionSyncResponse>('/integrations/notion/sync', {
    method: 'POST',
    body: { database_id: databaseId }
  });
}

// ============================================
// HubSpot Integration
// ============================================

export async function initiateHubSpotAuth() {
  return request<HubSpotAuthResponse>('/integrations/hubspot/auth');
}

export async function getHubSpotConnectionStatus() {
  return request<HubSpotConnectionStatus>('/integrations/hubspot/status');
}

export async function disconnectHubSpot() {
  return request('/integrations/hubspot', { method: 'DELETE' });
}

export async function syncHubSpot() {
  return request<IntegrationSyncResponse>('/integrations/hubspot/sync', { method: 'POST' });
}

// ============================================
// GoHighLevel Integration
// ============================================

export async function initiateGoHighLevelAuth() {
  return request<GoHighLevelAuthResponse>('/integrations/gohighlevel/auth');
}

export async function getGoHighLevelConnectionStatus() {
  return request<GoHighLevelConnectionStatus>('/integrations/gohighlevel/status');
}

export async function disconnectGoHighLevel() {
  return request('/integrations/gohighlevel', { method: 'DELETE' });
}

export async function syncGoHighLevel() {
  return request<IntegrationSyncResponse>('/integrations/gohighlevel/sync', { method: 'POST' });
}

// ============================================
// Linear Integration
// ============================================

export async function initiateLinearAuth() {
  return request<LinearAuthResponse>('/integrations/linear/auth');
}

export async function getLinearConnectionStatus() {
  return request<LinearConnectionStatus>('/integrations/linear/status');
}

export async function disconnectLinear() {
  return request('/integrations/linear', { method: 'DELETE' });
}

export async function syncLinear() {
  return request<IntegrationSyncResponse>('/integrations/linear/sync', { method: 'POST' });
}

// ============================================
// Asana Integration
// ============================================

export async function initiateAsanaAuth() {
  return request<AsanaAuthResponse>('/integrations/asana/auth');
}

export async function getAsanaConnectionStatus() {
  return request<AsanaConnectionStatus>('/integrations/asana/status');
}

export async function disconnectAsana() {
  return request('/integrations/asana', { method: 'DELETE' });
}

export async function syncAsana() {
  return request<IntegrationSyncResponse>('/integrations/asana/sync', { method: 'POST' });
}

// ============================================
// File Import (ChatGPT, Claude exports, etc.)
// ============================================

/**
 * Import data from a file export (ChatGPT conversations, Claude exports, etc.)
 * @param file - The file to import
 * @param source - The source of the export (chatgpt, claude, perplexity, etc.)
 */
export async function importFile(file: File, source: 'chatgpt' | 'claude' | 'perplexity' | 'gemini' | 'other') {
  const formData = new FormData();
  formData.append('file', file);
  formData.append('source', source);

  const response = await raw.postFormData('/integrations/import', formData);
  if (!response.ok) {
    const error = await response.json().catch(() => ({ detail: 'Import failed' }));
    throw new Error(error.detail || 'Import failed');
  }
  return response.json() as Promise<FileImportResponse>;
}

/**
 * Get progress of an ongoing import
 */
export async function getImportProgress(importId: string) {
  return request<ImportProgress>(`/integrations/import/${importId}/progress`);
}

// ============================================
// Custom MCP Connector
// ============================================

export interface MCPConnector {
  id: string;
  name: string;
  description?: string;
  server_url: string;
  auth_type: 'none' | 'api_key' | 'oauth';
  api_key?: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateMCPConnectorData {
  name: string;
  description?: string;
  server_url: string;
  auth_type: 'none' | 'api_key' | 'oauth';
  api_key?: string;
}

export async function getMCPConnectors() {
  return request<{ connectors: MCPConnector[] }>('/integrations/mcp/connectors');
}

export async function createMCPConnector(data: CreateMCPConnectorData) {
  return request<MCPConnector>('/integrations/mcp/connectors', {
    method: 'POST',
    body: data
  });
}

export async function updateMCPConnector(id: string, data: Partial<CreateMCPConnectorData>) {
  return request<MCPConnector>(`/integrations/mcp/connectors/${id}`, {
    method: 'PUT',
    body: data
  });
}

export async function deleteMCPConnector(id: string) {
  return request(`/integrations/mcp/connectors/${id}`, { method: 'DELETE' });
}

export async function testMCPConnector(id: string) {
  return request<{ success: boolean; message?: string }>(`/integrations/mcp/connectors/${id}/test`, {
    method: 'POST'
  });
}
