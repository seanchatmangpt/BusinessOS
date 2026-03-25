package semconv

import "go.opentelemetry.io/otel/attribute"

// Wave 9 Iteration 16: ChatmanGPT Session
const (
	ChatmangptSessionIDKey            = attribute.Key("chatmangpt.session.id")
	ChatmangptSessionTokenCountKey    = attribute.Key("chatmangpt.session.token_count")
	ChatmangptSessionModelSwitchesKey = attribute.Key("chatmangpt.session.model_switches")
	ChatmangptSessionTurnCountKey     = attribute.Key("chatmangpt.session.turn_count")
)

func ChatmangptSessionID(val string) attribute.KeyValue {
	return ChatmangptSessionIDKey.String(val)
}
func ChatmangptSessionTokenCount(val int64) attribute.KeyValue {
	return ChatmangptSessionTokenCountKey.Int64(val)
}
func ChatmangptSessionModelSwitches(val int64) attribute.KeyValue {
	return ChatmangptSessionModelSwitchesKey.Int64(val)
}
func ChatmangptSessionTurnCount(val int64) attribute.KeyValue {
	return ChatmangptSessionTurnCountKey.Int64(val)
}

// Wave 9 Iteration 16: A2A Message Routing
const (
	A2AMessagePriorityKey  = attribute.Key("a2a.message.priority")
	A2AMessageSizeBytesKey = attribute.Key("a2a.message.size_bytes")
	A2AMessageEncodingKey  = attribute.Key("a2a.message.encoding")
	A2AMessageTTLMsKey     = attribute.Key("a2a.message.ttl_ms")
)

func A2AMessagePriority(val string) attribute.KeyValue { return A2AMessagePriorityKey.String(val) }
func A2AMessageSizeBytes(val int64) attribute.KeyValue { return A2AMessageSizeBytesKey.Int64(val) }
func A2AMessageEncoding(val string) attribute.KeyValue { return A2AMessageEncodingKey.String(val) }
func A2AMessageTTLMs(val int64) attribute.KeyValue     { return A2AMessageTTLMsKey.Int64(val) }

// Wave 9 Iteration 16: Process Mining Decision Mining
const (
	PMDecisionPointIDKey    = attribute.Key("process.mining.decision.point_id")
	PMDecisionOutcomeKey    = attribute.Key("process.mining.decision.outcome")
	PMDecisionConfidenceKey = attribute.Key("process.mining.decision.confidence")
	PMDecisionRuleCountKey  = attribute.Key("process.mining.decision.rule_count")
	PMActivityFrequencyKey  = attribute.Key("process.mining.activity.frequency")
)

func PMDecisionPointID(val string) attribute.KeyValue { return PMDecisionPointIDKey.String(val) }
func PMDecisionOutcome(val string) attribute.KeyValue { return PMDecisionOutcomeKey.String(val) }
func PMDecisionConfidence(val float64) attribute.KeyValue {
	return PMDecisionConfidenceKey.Float64(val)
}
func PMDecisionRuleCount(val int64) attribute.KeyValue  { return PMDecisionRuleCountKey.Int64(val) }
func PMActivityFrequency(val float64) attribute.KeyValue { return PMActivityFrequencyKey.Float64(val) }

// Wave 9 Iteration 16: Consensus Leader Rotation
const (
	ConsensusLeaderRotationCountKey = attribute.Key("consensus.leader.rotation_count")
	ConsensusLeaderTenureMsKey      = attribute.Key("consensus.leader.tenure_ms")
	ConsensusLeaderScoreKey         = attribute.Key("consensus.leader.score")
)

func ConsensusLeaderRotationCount(val int64) attribute.KeyValue {
	return ConsensusLeaderRotationCountKey.Int64(val)
}
func ConsensusLeaderTenureMs(val int64) attribute.KeyValue {
	return ConsensusLeaderTenureMsKey.Int64(val)
}
func ConsensusLeaderScore(val float64) attribute.KeyValue {
	return ConsensusLeaderScoreKey.Float64(val)
}

// Wave 9 Iteration 16: Healing Prediction
const (
	HealingPredictionHorizonMsKey  = attribute.Key("healing.prediction.horizon_ms")
	HealingPredictionConfidenceKey = attribute.Key("healing.prediction.confidence")
	HealingPredictionModelKey      = attribute.Key("healing.prediction.model")
)

func HealingPredictionHorizonMs(val int64) attribute.KeyValue {
	return HealingPredictionHorizonMsKey.Int64(val)
}
func HealingPredictionConfidence(val float64) attribute.KeyValue {
	return HealingPredictionConfidenceKey.Float64(val)
}
func HealingPredictionModel(val string) attribute.KeyValue {
	return HealingPredictionModelKey.String(val)
}

// Wave 9 Iteration 16: LLM Streaming
const (
	LLMStreamingChunkCountKey      = attribute.Key("llm.streaming.chunk_count")
	LLMStreamingFirstTokenMsKey    = attribute.Key("llm.streaming.first_token_ms")
	LLMStreamingTokensPerSecondKey = attribute.Key("llm.streaming.tokens_per_second")
	LLMStreamingCompleteKey        = attribute.Key("llm.streaming.complete")
)

func LLMStreamingChunkCount(val int64) attribute.KeyValue {
	return LLMStreamingChunkCountKey.Int64(val)
}
func LLMStreamingFirstTokenMs(val int64) attribute.KeyValue {
	return LLMStreamingFirstTokenMsKey.Int64(val)
}
func LLMStreamingTokensPerSecond(val float64) attribute.KeyValue {
	return LLMStreamingTokensPerSecondKey.Float64(val)
}
func LLMStreamingComplete(val bool) attribute.KeyValue {
	return LLMStreamingCompleteKey.Bool(val)
}

// Wave 9 Iteration 16: Workspace Context Snapshot
const (
	WorkspaceContextSnapshotIDKey       = attribute.Key("workspace.context.snapshot_id")
	WorkspaceContextCompressionRatioKey = attribute.Key("workspace.context.compression_ratio")
	WorkspaceContextSizeTokensKey       = attribute.Key("workspace.context.size_tokens")
)

func WorkspaceContextSnapshotID(val string) attribute.KeyValue {
	return WorkspaceContextSnapshotIDKey.String(val)
}
func WorkspaceContextCompressionRatio(val float64) attribute.KeyValue {
	return WorkspaceContextCompressionRatioKey.Float64(val)
}
func WorkspaceContextSizeTokens(val int64) attribute.KeyValue {
	return WorkspaceContextSizeTokensKey.Int64(val)
}
