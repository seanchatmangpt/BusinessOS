<script lang="ts">
  import { onMount } from 'svelte';
  import { Cloud, CloudOff, CheckCircle, AlertCircle, Loader2, RefreshCw, ExternalLink } from 'lucide-svelte';
  import { currentWorkspace } from '$lib/stores/workspaces';
  import {
    getMIOSAStatus,
    pingMIOSACloud,
    syncToMIOSACloud,
    type MIOSAConnectionStatus,
    type SyncResult,
  } from '$lib/api/miosa';

  // ---------------------------------------------------------------------------
  // State
  // ---------------------------------------------------------------------------

  let status = $state<MIOSAConnectionStatus | null>(null);
  let isLoadingStatus = $state(true);
  let isSyncing = $state(false);
  let isPinging = $state(false);
  let lastSyncResult = $state<SyncResult | null>(null);
  let statusError = $state<string | null>(null);
  let syncError = $state<string | null>(null);

  // ---------------------------------------------------------------------------
  // Lifecycle
  // ---------------------------------------------------------------------------

  onMount(async () => {
    await loadStatus();
  });

  // ---------------------------------------------------------------------------
  // Actions
  // ---------------------------------------------------------------------------

  async function loadStatus() {
    isLoadingStatus = true;
    statusError = null;
    try {
      status = await getMIOSAStatus();
    } catch (err) {
      statusError = err instanceof Error ? err.message : 'Failed to load connection status';
    } finally {
      isLoadingStatus = false;
    }
  }

  async function handlePingCloud() {
    isPinging = true;
    syncError = null;
    try {
      const result = await pingMIOSACloud();
      if (!result.connected) {
        syncError = result.error ?? 'API key validation failed. Check your MIOSA_API_KEY.';
      }
      // Reload status to reflect updated connection state
      await loadStatus();
    } catch (err) {
      syncError = err instanceof Error ? err.message : 'Ping failed';
    } finally {
      isPinging = false;
    }
  }

  async function handleSync() {
    if (!$currentWorkspace?.id) {
      syncError = 'No workspace selected';
      return;
    }
    isSyncing = true;
    syncError = null;
    lastSyncResult = null;

    try {
      lastSyncResult = await syncToMIOSACloud($currentWorkspace.id);
      if (!lastSyncResult.success) {
        syncError = lastSyncResult.error ?? 'Sync failed';
      }
    } catch (err) {
      syncError = err instanceof Error ? err.message : 'Sync failed';
    } finally {
      isSyncing = false;
    }
  }

  // ---------------------------------------------------------------------------
  // Derived
  // ---------------------------------------------------------------------------

  const isConnected = $derived(status?.connected === true);
  const isCloud = $derived(status?.mode === 'cloud');
  const apiKeySet = $derived(status?.api_key_set === true);

  function formatDate(iso?: string): string {
    if (!iso) return 'Never';
    return new Date(iso).toLocaleString();
  }
</script>

<!-- Panel header -->
<div class="space-y-6">
  <div>
    <h3 class="text-lg font-semibold text-gray-900 dark:text-white">MIOSA Cloud</h3>
    <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
      Optionally sync your workspace configuration (settings, agents, app definitions,
      templates) to MIOSA Cloud. Raw business data stays local.
    </p>
  </div>

  <!-- Status banner -->
  {#if isLoadingStatus}
    <div class="flex items-center gap-2 rounded-lg border border-gray-200 bg-gray-50 p-4 dark:border-gray-700 dark:bg-gray-800">
      <Loader2 class="h-4 w-4 animate-spin text-gray-400" aria-hidden="true" />
      <span class="text-sm text-gray-500">Loading connection status...</span>
    </div>
  {:else if statusError}
    <div class="flex items-center gap-2 rounded-lg border border-red-200 bg-red-50 p-4 dark:border-red-800 dark:bg-red-900/20">
      <AlertCircle class="h-4 w-4 flex-shrink-0 text-red-500" aria-hidden="true" />
      <span class="text-sm text-red-700 dark:text-red-400">{statusError}</span>
    </div>
  {:else if status}
    <!-- Connected state -->
    {#if isConnected}
      <div class="flex items-start gap-3 rounded-lg border border-green-200 bg-green-50 p-4 dark:border-green-800 dark:bg-green-900/20">
        <CheckCircle class="mt-0.5 h-4 w-4 flex-shrink-0 text-green-600 dark:text-green-400" aria-hidden="true" />
        <div class="flex-1">
          <p class="text-sm font-medium text-green-800 dark:text-green-300">Connected to MIOSA Cloud</p>
          <p class="mt-0.5 text-xs text-green-600 dark:text-green-500">
            Last synced: {formatDate(status.last_sync)}
          </p>
        </div>
        <Cloud class="h-4 w-4 text-green-500" aria-hidden="true" />
      </div>

    <!-- API key set but not yet validated / not in cloud mode -->
    {:else if apiKeySet && !isCloud}
      <div class="flex items-start gap-3 rounded-lg border border-amber-200 bg-amber-50 p-4 dark:border-amber-800 dark:bg-amber-900/20">
        <AlertCircle class="mt-0.5 h-4 w-4 flex-shrink-0 text-amber-600" aria-hidden="true" />
        <div class="flex-1">
          <p class="text-sm font-medium text-amber-800 dark:text-amber-300">API key found but cloud mode is not active</p>
          <p class="mt-0.5 text-xs text-amber-600 dark:text-amber-500">
            Set <code class="rounded bg-amber-100 px-1 dark:bg-amber-900">OSA_MODE=cloud</code> in your .env to enable cloud sync.
          </p>
        </div>
      </div>

    <!-- No connection -->
    {:else}
      <div class="flex items-start gap-3 rounded-lg border border-gray-200 bg-gray-50 p-4 dark:border-gray-700 dark:bg-gray-800">
        <CloudOff class="mt-0.5 h-4 w-4 flex-shrink-0 text-gray-400" aria-hidden="true" />
        <div>
          <p class="text-sm font-medium text-gray-700 dark:text-gray-300">Not connected to MIOSA Cloud</p>
          <p class="mt-0.5 text-xs text-gray-500">
            Running in local mode. Your data stays on your machine.
          </p>
        </div>
      </div>
    {/if}
  {/if}

  <!-- Configuration section -->
  <div class="space-y-4 rounded-lg border border-gray-200 p-4 dark:border-gray-700">
    <h4 class="text-sm font-medium text-gray-900 dark:text-white">Configuration</h4>

    <div class="space-y-1">
      <label for="miosa-api-key" class="block text-xs font-medium text-gray-700 dark:text-gray-300">
        MIOSA API Key
      </label>
      <p class="text-xs text-gray-500 dark:text-gray-400">
        Set via <code class="rounded bg-gray-100 px-1 dark:bg-gray-700">MIOSA_API_KEY</code> in your .env file,
        or in the Electron app via the system keychain. The key is never sent to the browser.
      </p>
      <div class="mt-2 flex items-center gap-2">
        <div
          class="flex-1 rounded-md border border-gray-200 bg-gray-50 px-3 py-2 text-sm text-gray-500 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-400"
          aria-label="API key status"
        >
          {#if apiKeySet}
            <span class="flex items-center gap-1.5">
              <CheckCircle class="h-3.5 w-3.5 text-green-500" aria-hidden="true" />
              API key configured in backend
            </span>
          {:else}
            Not configured — add MIOSA_API_KEY to your .env
          {/if}
        </div>
        <a
          href="https://app.miosa.ai/settings/api-keys"
          target="_blank"
          rel="noopener noreferrer"
          class="flex items-center gap-1 text-xs text-blue-600 hover:text-blue-700 dark:text-blue-400"
          aria-label="Get your MIOSA API key (opens in new tab)"
        >
          Get key
          <ExternalLink class="h-3 w-3" aria-hidden="true" />
        </a>
      </div>
    </div>

    <div class="space-y-1">
      <label class="block text-xs font-medium text-gray-700 dark:text-gray-300">
        OSA Mode
      </label>
      <p class="text-xs text-gray-500 dark:text-gray-400">
        Set <code class="rounded bg-gray-100 px-1 dark:bg-gray-700">OSA_MODE=cloud</code> to route OSA agent traffic
        through MIOSA Cloud. Default is <code class="rounded bg-gray-100 px-1 dark:bg-gray-700">local</code>
        (localhost:8089, SQLite, no cloud).
      </p>
      <div
        class="mt-1 inline-flex items-center rounded-md px-2.5 py-1 text-xs font-medium
          {isCloud
            ? 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-300'
            : 'bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300'}"
        aria-label="Current OSA mode"
      >
        Current mode: {status?.mode ?? 'local'}
      </div>
    </div>
  </div>

  <!-- Actions -->
  <div class="space-y-3">
    <!-- Validate API Key -->
    <div class="flex items-center justify-between">
      <div>
        <p class="text-sm font-medium text-gray-900 dark:text-white">Validate connection</p>
        <p class="text-xs text-gray-500 dark:text-gray-400">
          Confirm your API key is valid by pinging MIOSA Cloud.
        </p>
      </div>
      <button
        onclick={handlePingCloud}
        disabled={isPinging || !apiKeySet}
        class="flex items-center gap-2 rounded-md bg-white px-3 py-2 text-sm font-medium text-gray-700
          shadow-sm ring-1 ring-gray-300 hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-50
          dark:bg-gray-700 dark:text-gray-200 dark:ring-gray-600 dark:hover:bg-gray-600"
        aria-label="Validate MIOSA API key"
      >
        {#if isPinging}
          <Loader2 class="h-4 w-4 animate-spin" aria-hidden="true" />
          Validating...
        {:else}
          <RefreshCw class="h-4 w-4" aria-hidden="true" />
          Validate key
        {/if}
      </button>
    </div>

    <!-- Sync to MIOSA Cloud -->
    <div class="flex items-center justify-between">
      <div>
        <p class="text-sm font-medium text-gray-900 dark:text-white">Publish to MIOSA Cloud</p>
        <p class="text-xs text-gray-500 dark:text-gray-400">
          Syncs workspace config, agents, app definitions, and templates.
          No business data is included.
        </p>
      </div>
      <button
        onclick={handleSync}
        disabled={isSyncing || !isConnected}
        class="flex items-center gap-2 rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white
          hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-50
          focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
        aria-label="Publish workspace configuration to MIOSA Cloud"
      >
        {#if isSyncing}
          <Loader2 class="h-4 w-4 animate-spin" aria-hidden="true" />
          Syncing...
        {:else}
          <Cloud class="h-4 w-4" aria-hidden="true" />
          Publish to MIOSA Cloud
        {/if}
      </button>
    </div>
  </div>

  <!-- Feedback messages -->
  {#if syncError}
    <div
      class="flex items-center gap-2 rounded-lg border border-red-200 bg-red-50 p-3 dark:border-red-800 dark:bg-red-900/20"
      role="alert"
    >
      <AlertCircle class="h-4 w-4 flex-shrink-0 text-red-500" aria-hidden="true" />
      <span class="text-sm text-red-700 dark:text-red-400">{syncError}</span>
    </div>
  {/if}

  {#if lastSyncResult?.success}
    <div
      class="flex items-center gap-2 rounded-lg border border-green-200 bg-green-50 p-3 dark:border-green-800 dark:bg-green-900/20"
      role="status"
    >
      <CheckCircle class="h-4 w-4 flex-shrink-0 text-green-500" aria-hidden="true" />
      <span class="text-sm text-green-700 dark:text-green-400">
        Synced successfully at {formatDate(lastSyncResult.synced_at)}
        {#if lastSyncResult.manifest_id}
          &nbsp;(manifest {lastSyncResult.manifest_id.slice(0, 8)})
        {/if}
      </span>
    </div>
  {/if}

  <!-- Informational footer -->
  <div class="rounded-lg bg-gray-50 p-3 dark:bg-gray-800">
    <p class="text-xs text-gray-500 dark:text-gray-400">
      <strong class="text-gray-700 dark:text-gray-300">What syncs:</strong>
      workspace settings, agent configurations, custom app definitions, templates.
    </p>
    <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
      <strong class="text-gray-700 dark:text-gray-300">What stays local:</strong>
      tasks, projects, contacts, deals, conversations, emails, files, embeddings.
    </p>
  </div>
</div>
