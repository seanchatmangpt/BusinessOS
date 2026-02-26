<script lang="ts">
  import {
    VersionBadge,
    VersionDropdown,
    VersionTimelinePanel,
    RestoreConfirmDialog,
    SaveVersionModal,
    VersionPreviewModal,
    MOCK_VERSIONS,
    toVersionSummary
  } from '$lib/components/versioning';
  import type { Version, VersionSummary } from '$lib/types/versions';

  // State for demo
  let currentVersion = $state(5);
  let hasUnsavedChanges = $state(false);
  let isPreviewingOldVersion = $state(false);
  let previewingVersion = $state<number | undefined>(undefined);

  // Panel/modal states
  let showTimeline = $state(false);
  let showSaveModal = $state(false);
  let showRestoreDialog = $state(false);
  let showPreviewModal = $state(false);

  // Selected version for actions
  let selectedVersion = $state<Version>(MOCK_VERSIONS[0]);

  // Convert to summaries for dropdown
  const versionSummaries = $derived(MOCK_VERSIONS.map(toVersionSummary));

  // Handlers
  function handleVersionSelect(version: VersionSummary) {
    console.log('Selected version:', version);
    isPreviewingOldVersion = !version.isCurrent;
    previewingVersion = version.isCurrent ? undefined : version.versionNumber;
  }

  function handlePreview(version: Version) {
    selectedVersion = version;
    showPreviewModal = true;
  }

  function handleRestore(version: Version) {
    selectedVersion = version;
    showRestoreDialog = true;
  }

  function handleRestoreConfirm() {
    console.log('Restoring to version:', selectedVersion.versionNumber);
    showRestoreDialog = false;
    showPreviewModal = false;
    // In real app: call API, update state
  }

  function handleSave(label?: string) {
    console.log('Saving version with label:', label);
    showSaveModal = false;
    // In real app: call API, increment version
  }

  function exitPreview() {
    isPreviewingOldVersion = false;
    previewingVersion = undefined;
  }
</script>

<div class="demo-page">
  <header class="demo-header">
    <h1>Versioning Components Demo</h1>
    <p>Interactive preview of all versioning UI components</p>
  </header>

  <!-- Controls -->
  <section class="demo-section">
    <h2>Demo Controls</h2>
    <div class="controls">
      <label class="control">
        <input type="checkbox" bind:checked={hasUnsavedChanges} />
        <span>Has unsaved changes</span>
      </label>
      <button class="control-btn" onclick={exitPreview} disabled={!isPreviewingOldVersion}>
        Exit Preview Mode
      </button>
    </div>
  </section>

  <!-- VersionBadge -->
  <section class="demo-section">
    <h2>VersionBadge</h2>
    <p class="description">Shows current version status in toolbar. Click to open dropdown.</p>
    <div class="component-preview">
      <VersionBadge
        {currentVersion}
        {hasUnsavedChanges}
        {isPreviewingOldVersion}
        {previewingVersion}
        onclick={() => console.log('Badge clicked')}
      />
    </div>
    <div class="states">
      <span class="state-label">States:</span>
      <div class="state-demos">
        <div class="state-item">
          <span class="state-name">Normal:</span>
          <VersionBadge currentVersion={5} />
        </div>
        <div class="state-item">
          <span class="state-name">Unsaved:</span>
          <VersionBadge currentVersion={5} hasUnsavedChanges={true} />
        </div>
        <div class="state-item">
          <span class="state-name">Previewing:</span>
          <VersionBadge currentVersion={5} isPreviewingOldVersion={true} previewingVersion={3} />
        </div>
      </div>
    </div>
  </section>

  <!-- VersionDropdown -->
  <section class="demo-section">
    <h2>VersionDropdown</h2>
    <p class="description">Quick version selector with recent versions, view all, and save actions.</p>
    <div class="component-preview">
      <VersionDropdown
        {currentVersion}
        versions={versionSummaries}
        onVersionSelect={handleVersionSelect}
        onViewAll={() => (showTimeline = true)}
        onSaveVersion={() => (showSaveModal = true)}
      />
    </div>
  </section>

  <!-- Action Buttons for Panels/Modals -->
  <section class="demo-section">
    <h2>Panels & Modals</h2>
    <p class="description">Click to open each component.</p>
    <div class="action-buttons">
      <button class="action-btn" onclick={() => (showTimeline = true)}>
        Open Timeline Panel
      </button>
      <button class="action-btn" onclick={() => (showSaveModal = true)}>
        Open Save Modal
      </button>
      <button class="action-btn" onclick={() => { selectedVersion = MOCK_VERSIONS[2]; showRestoreDialog = true; }}>
        Open Restore Dialog (v3)
      </button>
      <button class="action-btn" onclick={() => { selectedVersion = MOCK_VERSIONS[1]; showPreviewModal = true; }}>
        Open Preview Modal (v4)
      </button>
    </div>
  </section>

  <!-- Integration Example -->
  <section class="demo-section">
    <h2>Toolbar Integration Example</h2>
    <p class="description">How it would look in an app toolbar.</p>
    <div class="toolbar-example">
      <div class="toolbar">
        <div class="toolbar-left">
          <span class="app-name">My CRM App</span>
        </div>
        <div class="toolbar-right">
          <VersionDropdown
            {currentVersion}
            versions={versionSummaries}
            onVersionSelect={handleVersionSelect}
            onViewAll={() => (showTimeline = true)}
            onSaveVersion={() => (showSaveModal = true)}
          />
          <button class="toolbar-btn">Settings</button>
          <button class="toolbar-btn primary">Publish</button>
        </div>
      </div>
    </div>
  </section>
</div>

<!-- Timeline Panel -->
<VersionTimelinePanel
  appId="app-123"
  versions={MOCK_VERSIONS}
  isOpen={showTimeline}
  onClose={() => (showTimeline = false)}
  onPreview={handlePreview}
  onRestore={handleRestore}
  onCompare={() => console.log('Compare clicked')}
/>

<!-- Save Version Modal -->
<SaveVersionModal
  appId="app-123"
  {currentVersion}
  isOpen={showSaveModal}
  onClose={() => (showSaveModal = false)}
  onSave={handleSave}
/>

<!-- Restore Confirm Dialog -->
<RestoreConfirmDialog
  version={selectedVersion}
  {currentVersion}
  isOpen={showRestoreDialog}
  onClose={() => (showRestoreDialog = false)}
  onConfirm={handleRestoreConfirm}
/>

<!-- Preview Modal -->
<VersionPreviewModal
  version={selectedVersion}
  isOpen={showPreviewModal}
  onClose={() => (showPreviewModal = false)}
  onRestore={handleRestore}
/>

<style>
  .demo-page {
    max-width: 900px;
    margin: 0 auto;
    padding: 40px 20px;
  }

  .demo-header {
    margin-bottom: 40px;
  }

  .demo-header h1 {
    font-size: 28px;
    font-weight: 700;
    color: #111827;
    margin: 0 0 8px 0;
  }

  :global(.dark) .demo-header h1 {
    color: #f3f4f6;
  }

  .demo-header p {
    font-size: 15px;
    color: #6b7280;
    margin: 0;
  }

  :global(.dark) .demo-header p {
    color: #9ca3af;
  }

  .demo-section {
    margin-bottom: 48px;
    padding-bottom: 48px;
    border-bottom: 1px solid #e5e7eb;
  }

  :global(.dark) .demo-section {
    border-bottom-color: #374151;
  }

  .demo-section:last-child {
    border-bottom: none;
  }

  .demo-section h2 {
    font-size: 18px;
    font-weight: 600;
    color: #111827;
    margin: 0 0 8px 0;
  }

  :global(.dark) .demo-section h2 {
    color: #f3f4f6;
  }

  .description {
    font-size: 14px;
    color: #6b7280;
    margin: 0 0 20px 0;
  }

  :global(.dark) .description {
    color: #9ca3af;
  }

  .component-preview {
    padding: 24px;
    background: #f9fafb;
    border: 1px solid #e5e7eb;
    border-radius: 8px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  :global(.dark) .component-preview {
    background: #1f2937;
    border-color: #374151;
  }

  .controls {
    display: flex;
    align-items: center;
    gap: 16px;
    flex-wrap: wrap;
  }

  .control {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 14px;
    color: #374151;
    cursor: pointer;
  }

  :global(.dark) .control {
    color: #d1d5db;
  }

  .control input {
    width: 16px;
    height: 16px;
  }

  .control-btn {
    padding: 8px 16px;
    font-size: 13px;
    font-weight: 500;
    color: #374151;
    background: #ffffff;
    border: 1px solid #e5e7eb;
    border-radius: 6px;
    cursor: pointer;
    font-family: inherit;
  }

  :global(.dark) .control-btn {
    color: #d1d5db;
    background: #374151;
    border-color: #4b5563;
  }

  .control-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .states {
    margin-top: 20px;
    padding-top: 20px;
    border-top: 1px dashed #e5e7eb;
  }

  :global(.dark) .states {
    border-top-color: #374151;
  }

  .state-label {
    font-size: 12px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    color: #9ca3af;
    display: block;
    margin-bottom: 12px;
  }

  .state-demos {
    display: flex;
    flex-wrap: wrap;
    gap: 24px;
  }

  .state-item {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .state-name {
    font-size: 13px;
    color: #6b7280;
  }

  :global(.dark) .state-name {
    color: #9ca3af;
  }

  .action-buttons {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
  }

  .action-btn {
    padding: 10px 20px;
    font-size: 14px;
    font-weight: 500;
    color: #ffffff;
    background: #6366f1;
    border: none;
    border-radius: 6px;
    cursor: pointer;
    font-family: inherit;
    transition: background 150ms ease;
  }

  .action-btn:hover {
    background: #4f46e5;
  }

  /* Toolbar Example */
  .toolbar-example {
    background: #f3f4f6;
    border-radius: 8px;
    padding: 16px;
  }

  :global(.dark) .toolbar-example {
    background: #111827;
  }

  .toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 16px;
    background: #ffffff;
    border: 1px solid #e5e7eb;
    border-radius: 8px;
  }

  :global(.dark) .toolbar {
    background: #1f2937;
    border-color: #374151;
  }

  .toolbar-left {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .app-name {
    font-size: 15px;
    font-weight: 600;
    color: #111827;
  }

  :global(.dark) .app-name {
    color: #f3f4f6;
  }

  .toolbar-right {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .toolbar-btn {
    padding: 6px 12px;
    font-size: 13px;
    font-weight: 500;
    color: #374151;
    background: transparent;
    border: 1px solid #e5e7eb;
    border-radius: 6px;
    cursor: pointer;
    font-family: inherit;
  }

  :global(.dark) .toolbar-btn {
    color: #d1d5db;
    border-color: #4b5563;
  }

  .toolbar-btn.primary {
    color: #ffffff;
    background: #6366f1;
    border-color: #6366f1;
  }
</style>
