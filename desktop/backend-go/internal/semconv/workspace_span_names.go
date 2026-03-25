package semconv

const (
	// workspace_activity_track is the span name for "workspace.activity.track".
	//
	// Tracking a workspace activity — records type, duration, and context of user/agent actions.
	// Kind: internal
	// Stability: development
	WorkspaceActivityTrack = "workspace.activity.track"
	// workspace_checkpoint_save is the span name for "workspace.checkpoint.save".
	//
	// Saving a workspace checkpoint — persisting agent state and task queue for recovery.
	// Kind: internal
	// Stability: development
	WorkspaceCheckpointSave = "workspace.checkpoint.save"
	// workspace_context_checkpoint is the span name for "workspace.context.checkpoint".
	//
	// Creating a context checkpoint — snapshot of current workspace state for potential rollback.
	// Kind: internal
	// Stability: development
	WorkspaceContextCheckpoint = "workspace.context.checkpoint"
	// workspace_context_snapshot is the span name for "workspace.context.snapshot".
	//
	// Creating a compressed snapshot of workspace context for persistence or recovery.
	// Kind: internal
	// Stability: development
	WorkspaceContextSnapshot = "workspace.context.snapshot"
	// workspace_context_update is the span name for "workspace.context.update".
	//
	// Context window update — tokens added or pruned from the workspace context.
	// Kind: internal
	// Stability: development
	WorkspaceContextUpdate = "workspace.context.update"
	// workspace_memory_compact is the span name for "workspace.memory.compact".
	//
	// Workspace memory compaction — reducing memory footprint by consolidating and pruning stored context items.
	// Kind: internal
	// Stability: development
	WorkspaceMemoryCompact = "workspace.memory.compact"
	// workspace_orchestrate is the span name for "workspace.orchestrate".
	//
	// Orchestrating work distribution across agents in the workspace.
	// Kind: internal
	// Stability: development
	WorkspaceOrchestrate = "workspace.orchestrate"
	// workspace_session_end is the span name for "workspace.session.end".
	//
	// Ending a workspace session — recording final metrics and persisting session state.
	// Kind: internal
	// Stability: development
	WorkspaceSessionEnd = "workspace.session.end"
	// workspace_session_start is the span name for "workspace.session.start".
	//
	// Workspace session initialization — agent begins processing in a new session context.
	// Kind: internal
	// Stability: development
	WorkspaceSessionStart = "workspace.session.start"
	// workspace_share is the span name for "workspace.share".
	//
	// Sharing a workspace with other agents — granting access with defined permissions and scope.
	// Kind: internal
	// Stability: development
	WorkspaceShare = "workspace.share"
	// workspace_tool_invoke is the span name for "workspace.tool.invoke".
	//
	// Tool invocation within a workspace session.
	// Kind: internal
	// Stability: development
	WorkspaceToolInvoke = "workspace.tool.invoke"
)