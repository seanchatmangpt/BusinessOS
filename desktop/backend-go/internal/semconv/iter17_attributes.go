package semconv

import "go.opentelemetry.io/otel/attribute"

// Wave 9 Iteration 17: MCP Tool Versioning
const (
	MCPToolVersionKey           = attribute.Key("mcp.tool.version")
	MCPToolSchemaHashKey        = attribute.Key("mcp.tool.schema_hash")
	MCPToolDeprecatedKey        = attribute.Key("mcp.tool.deprecated")
	MCPToolDeprecationReasonKey = attribute.Key("mcp.tool.deprecation.reason")
)

func MCPToolVersion(val string) attribute.KeyValue      { return MCPToolVersionKey.String(val) }
func MCPToolSchemaHash(val string) attribute.KeyValue   { return MCPToolSchemaHashKey.String(val) }
func MCPToolDeprecationReason(val string) attribute.KeyValue {
	return MCPToolDeprecationReasonKey.String(val)
}

// Wave 9 Iteration 17: A2A Capability Negotiation
const (
	A2ACapNegotiationIDKey      = attribute.Key("a2a.capability.negotiation.id")
	A2ACapNegotiationOutcomeKey = attribute.Key("a2a.capability.negotiation.outcome")
	A2ACapNegotiationRoundsKey  = attribute.Key("a2a.capability.negotiation.rounds")
)

func A2ACapNegotiationID(val string) attribute.KeyValue { return A2ACapNegotiationIDKey.String(val) }
func A2ACapNegotiationOutcome(val string) attribute.KeyValue {
	return A2ACapNegotiationOutcomeKey.String(val)
}
func A2ACapNegotiationRounds(val int64) attribute.KeyValue {
	return A2ACapNegotiationRoundsKey.Int64(val)
}

// Wave 9 Iteration 17: Process Mining Root Cause
const (
	PMRootCauseIDKey         = attribute.Key("process.mining.root_cause.id")
	PMRootCauseTypeKey       = attribute.Key("process.mining.root_cause.type")
	PMRootCauseConfidenceKey = attribute.Key("process.mining.root_cause.confidence")
	PMAnomalyScoreKey        = attribute.Key("process.mining.anomaly.score")
)

func PMRootCauseID(val string) attribute.KeyValue  { return PMRootCauseIDKey.String(val) }
func PMRootCauseType(val string) attribute.KeyValue { return PMRootCauseTypeKey.String(val) }
func PMRootCauseConfidence(val float64) attribute.KeyValue {
	return PMRootCauseConfidenceKey.Float64(val)
}
func PMAnomalyScore(val float64) attribute.KeyValue { return PMAnomalyScoreKey.Float64(val) }

// Wave 9 Iteration 17: Consensus View Change
const (
	ConsensusViewChangeReasonKey     = attribute.Key("consensus.view_change.reason")
	ConsensusViewChangeDurationMsKey = attribute.Key("consensus.view_change.duration_ms")
	ConsensusViewChangeBackoffMsKey  = attribute.Key("consensus.view_change.backoff_ms")
)

func ConsensusViewChangeReason(val string) attribute.KeyValue {
	return ConsensusViewChangeReasonKey.String(val)
}
func ConsensusViewChangeDurationMs(val int64) attribute.KeyValue {
	return ConsensusViewChangeDurationMsKey.Int64(val)
}
func ConsensusViewChangeBackoffMs(val int64) attribute.KeyValue {
	return ConsensusViewChangeBackoffMsKey.Int64(val)
}

// Wave 9 Iteration 17: Healing Playbook
const (
	HealingPlaybookIDKey          = attribute.Key("healing.playbook.id")
	HealingPlaybookStepCountKey   = attribute.Key("healing.playbook.step_count")
	HealingPlaybookExecutionMsKey = attribute.Key("healing.playbook.execution_ms")
	HealingPlaybookStepCurrentKey = attribute.Key("healing.playbook.step_current")
)

func HealingPlaybookID(val string) attribute.KeyValue { return HealingPlaybookIDKey.String(val) }
func HealingPlaybookStepCount(val int64) attribute.KeyValue {
	return HealingPlaybookStepCountKey.Int64(val)
}
func HealingPlaybookExecutionMs(val int64) attribute.KeyValue {
	return HealingPlaybookExecutionMsKey.Int64(val)
}

// Wave 9 Iteration 17: LLM Context Management
const (
	LLMContextMaxTokensKey        = attribute.Key("llm.context.max_tokens")
	LLMContextOverflowStrategyKey = attribute.Key("llm.context.overflow_strategy")
	LLMContextOverflowCountKey    = attribute.Key("llm.context.overflow_count")
	LLMContextUtilizationKey      = attribute.Key("llm.context.utilization")
)

func LLMContextMaxTokens(val int64) attribute.KeyValue { return LLMContextMaxTokensKey.Int64(val) }
func LLMContextOverflowStrategy(val string) attribute.KeyValue {
	return LLMContextOverflowStrategyKey.String(val)
}
func LLMContextUtilization(val float64) attribute.KeyValue {
	return LLMContextUtilizationKey.Float64(val)
}

// Wave 9 Iteration 17: Agent Pipeline
const (
	AgentPipelineIDKey          = attribute.Key("agent.pipeline.id")
	AgentPipelineStageKey       = attribute.Key("agent.pipeline.stage")
	AgentPipelineStageCountKey  = attribute.Key("agent.pipeline.stage_count")
	AgentPipelineRetryPolicyKey = attribute.Key("agent.pipeline.retry_policy")
)

func AgentPipelineID(val string) attribute.KeyValue    { return AgentPipelineIDKey.String(val) }
func AgentPipelineStage(val string) attribute.KeyValue { return AgentPipelineStageKey.String(val) }

// Wave 9 Iteration 17: Workspace Activity
const (
	WorkspaceActivityTypeKey       = attribute.Key("workspace.activity.type")
	WorkspaceActivityCountKey      = attribute.Key("workspace.activity.count")
	WorkspaceActivityDurationMsKey = attribute.Key("workspace.activity.duration_ms")
)

func WorkspaceActivityType(val string) attribute.KeyValue { return WorkspaceActivityTypeKey.String(val) }
func WorkspaceActivityDurationMs(val int64) attribute.KeyValue {
	return WorkspaceActivityDurationMsKey.Int64(val)
}
