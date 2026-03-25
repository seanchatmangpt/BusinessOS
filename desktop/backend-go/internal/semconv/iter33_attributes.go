package semconv

import "go.opentelemetry.io/otel/attribute"

// Iter33 — MCP server metrics, A2A contract dispute, PM process hierarchy,
//           consensus epoch transition, healing surge, LLM RAG

// MCP Server Metrics
const MCPServerMetricsRequestCountKey = attribute.Key("mcp.server.metrics.request_count")
const MCPServerMetricsErrorRateKey = attribute.Key("mcp.server.metrics.error_rate")
const MCPServerMetricsP99LatencyMsKey = attribute.Key("mcp.server.metrics.p99_latency_ms")

func MCPServerMetricsRequestCount(val int64) attribute.KeyValue   { return MCPServerMetricsRequestCountKey.Int64(val) }
func MCPServerMetricsErrorRate(val float64) attribute.KeyValue    { return MCPServerMetricsErrorRateKey.Float64(val) }
func MCPServerMetricsP99LatencyMs(val float64) attribute.KeyValue { return MCPServerMetricsP99LatencyMsKey.Float64(val) }

// A2A Contract Dispute
const A2AContractDisputeIDKey = attribute.Key("a2a.contract.dispute.id")
const A2AContractDisputeReasonKey = attribute.Key("a2a.contract.dispute.reason")
const A2AContractDisputeStatusKey = attribute.Key("a2a.contract.dispute.status")

func A2AContractDisputeID(val string) attribute.KeyValue     { return A2AContractDisputeIDKey.String(val) }
func A2AContractDisputeReason(val string) attribute.KeyValue { return A2AContractDisputeReasonKey.String(val) }
func A2AContractDisputeStatus(val string) attribute.KeyValue { return A2AContractDisputeStatusKey.String(val) }

// Process Mining Hierarchy
const ProcessMiningHierarchyDepthKey = attribute.Key("process.mining.hierarchy.depth")
const ProcessMiningHierarchyParentProcessIDKey = attribute.Key("process.mining.hierarchy.parent_process_id")
const ProcessMiningHierarchyChildCountKey = attribute.Key("process.mining.hierarchy.child_count")

func ProcessMiningHierarchyDepth(val int64) attribute.KeyValue            { return ProcessMiningHierarchyDepthKey.Int64(val) }
func ProcessMiningHierarchyParentProcessID(val string) attribute.KeyValue { return ProcessMiningHierarchyParentProcessIDKey.String(val) }
func ProcessMiningHierarchyChildCount(val int64) attribute.KeyValue       { return ProcessMiningHierarchyChildCountKey.Int64(val) }

// Consensus Epoch Transition
const ConsensusEpochTransitionFromEpochKey = attribute.Key("consensus.epoch.transition.from_epoch")
const ConsensusEpochTransitionToEpochKey = attribute.Key("consensus.epoch.transition.to_epoch")
const ConsensusEpochTransitionTriggerKey = attribute.Key("consensus.epoch.transition.trigger")

func ConsensusEpochTransitionFromEpoch(val int64) attribute.KeyValue { return ConsensusEpochTransitionFromEpochKey.Int64(val) }
func ConsensusEpochTransitionToEpoch(val int64) attribute.KeyValue   { return ConsensusEpochTransitionToEpochKey.Int64(val) }
func ConsensusEpochTransitionTrigger(val string) attribute.KeyValue  { return ConsensusEpochTransitionTriggerKey.String(val) }

// Healing Surge
const HealingSurgeThresholdMultiplierKey = attribute.Key("healing.surge.threshold_multiplier")
const HealingSurgeDetectionWindowMsKey = attribute.Key("healing.surge.detection_window_ms")
const HealingSurgeMitigationStrategyKey = attribute.Key("healing.surge.mitigation_strategy")

func HealingSurgeThresholdMultiplier(val float64) attribute.KeyValue { return HealingSurgeThresholdMultiplierKey.Float64(val) }
func HealingSurgeDetectionWindowMs(val int64) attribute.KeyValue     { return HealingSurgeDetectionWindowMsKey.Int64(val) }
func HealingSurgeMitigationStrategy(val string) attribute.KeyValue   { return HealingSurgeMitigationStrategyKey.String(val) }

// LLM RAG
const LLMRAGRetrievalKKey = attribute.Key("llm.rag.retrieval_k")
const LLMRAGSimilarityThresholdKey = attribute.Key("llm.rag.similarity_threshold")
const LLMRAGContextWindowTokensKey = attribute.Key("llm.rag.context_window_tokens")

func LLMRAGRetrievalK(val int64) attribute.KeyValue            { return LLMRAGRetrievalKKey.Int64(val) }
func LLMRAGSimilarityThreshold(val float64) attribute.KeyValue { return LLMRAGSimilarityThresholdKey.Float64(val) }
func LLMRAGContextWindowTokens(val int64) attribute.KeyValue   { return LLMRAGContextWindowTokensKey.Int64(val) }
