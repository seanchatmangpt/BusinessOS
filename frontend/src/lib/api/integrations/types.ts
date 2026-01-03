// Integrations API Types

// Generic integration types
export type IntegrationProvider =
  | 'google' | 'slack' | 'notion' | 'hubspot' | 'gohighlevel'
  | 'linear' | 'asana' | 'monday' | 'trello' | 'jira' | 'clickup'
  | 'zoom' | 'loom' | 'fireflies' | 'fathom' | 'tldv' | 'granola'
  | 'dropbox' | 'discord' | 'teams' | 'salesforce' | 'pipedrive'
  | 'chatgpt' | 'claude' | 'perplexity' | 'gemini'
  | 'evernote' | 'obsidian' | 'roam' | 'apple_notes';

export interface GenericAuthResponse {
  auth_url: string;
}

export interface GenericConnectionStatus {
  connected: boolean;
  provider: IntegrationProvider;
  account_name?: string;
  account_id?: string;
  account_email?: string;
  connected_at?: string;
  last_synced_at?: string;
  sync_enabled?: boolean;
  error?: string;
}

export interface AllIntegrationsStatus {
  integrations: Record<IntegrationProvider, GenericConnectionStatus>;
}

export interface IntegrationSyncResponse {
  success: boolean;
  synced_count: number;
  message?: string;
}

// Google
export interface GoogleAuthResponse {
  auth_url: string;
}

export interface GoogleConnectionStatus {
  connected: boolean;
  email?: string;
  connected_at?: string;
}

// Slack
export interface SlackAuthResponse {
  auth_url: string;
}

export interface SlackConnectionStatus {
  connected: boolean;
  workspace_name?: string;
  workspace_id?: string;
  user_name?: string;
  user_id?: string;
  connected_at?: string;
}

export interface SlackChannel {
  id: string;
  name: string;
  is_private: boolean;
  is_member: boolean;
  num_members?: number;
}

export interface SlackNotification {
  id: string;
  channel_id: string;
  channel_name: string;
  text: string;
  user_id: string;
  user_name: string;
  timestamp: string;
  thread_ts?: string;
}

export interface SlackChannelsResponse {
  channels: SlackChannel[];
}

export interface SlackNotificationsResponse {
  notifications: SlackNotification[];
  has_more: boolean;
}

// Notion
export interface NotionAuthResponse {
  auth_url: string;
}

export interface NotionConnectionStatus {
  connected: boolean;
  workspace_name?: string;
  workspace_id?: string;
  workspace_icon?: string;
  connected_at?: string;
}

export interface NotionDatabase {
  id: string;
  title: string;
  description?: string;
  icon?: string;
  url: string;
  created_time: string;
  last_edited_time: string;
}

export interface NotionPage {
  id: string;
  title: string;
  icon?: string;
  url: string;
  parent_type: 'database' | 'page' | 'workspace';
  parent_id?: string;
  created_time: string;
  last_edited_time: string;
}

export interface NotionDatabasesResponse {
  databases: NotionDatabase[];
}

export interface NotionPagesResponse {
  pages: NotionPage[];
  has_more: boolean;
  next_cursor?: string;
}

export interface NotionSyncResponse {
  success: boolean;
  synced_count: number;
  message?: string;
}

// HubSpot
export interface HubSpotAuthResponse {
  auth_url: string;
}

export interface HubSpotConnectionStatus {
  connected: boolean;
  portal_id?: string;
  portal_name?: string;
  user_email?: string;
  connected_at?: string;
}

// GoHighLevel
export interface GoHighLevelAuthResponse {
  auth_url: string;
}

export interface GoHighLevelConnectionStatus {
  connected: boolean;
  location_id?: string;
  location_name?: string;
  user_email?: string;
  connected_at?: string;
}

// Linear
export interface LinearAuthResponse {
  auth_url: string;
}

export interface LinearConnectionStatus {
  connected: boolean;
  workspace_id?: string;
  workspace_name?: string;
  user_name?: string;
  connected_at?: string;
}

// Asana
export interface AsanaAuthResponse {
  auth_url: string;
}

export interface AsanaConnectionStatus {
  connected: boolean;
  workspace_id?: string;
  workspace_name?: string;
  user_name?: string;
  user_email?: string;
  connected_at?: string;
}

// File Import (for ChatGPT, Claude exports, etc.)
export interface FileImportResponse {
  success: boolean;
  imported_count: number;
  message?: string;
  import_id?: string;
}

export interface ImportProgress {
  import_id: string;
  status: 'pending' | 'processing' | 'completed' | 'failed';
  progress_percent: number;
  imported_count: number;
  total_count: number;
  error?: string;
}
