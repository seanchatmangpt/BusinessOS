package semconv

// Iter23: Agent spawn profiling attributes
const AgentSpawnParentId = "agent.spawn.parent_id"
const AgentSpawnStrategy = "agent.spawn.strategy"
const AgentSpawnLatencyMs = "agent.spawn.latency_ms"

// Iter23: A2A escrow mechanics attributes
const A2AEscrowId = "a2a.escrow.id"
const A2AEscrowAmount = "a2a.escrow.amount"
const A2AEscrowReleaseCondition = "a2a.escrow.release_condition"
const A2AEscrowStatus = "a2a.escrow.status"

// Iter23: PM bottleneck scoring attributes
const ProcessMiningBottleneckScore = "process.mining.bottleneck.score"
const ProcessMiningBottleneckRank = "process.mining.bottleneck.rank"
const ProcessMiningBottleneckImpactMs = "process.mining.bottleneck.impact_ms"

// Iter23: Consensus epoch key rotation attributes
const ConsensusEpochKeyRotationId = "consensus.epoch.key_rotation_id"
const ConsensusEpochKeyRotationReason = "consensus.epoch.key_rotation_reason"
const ConsensusEpochKeyRotationMs = "consensus.epoch.key_rotation_ms"

// Iter23: Healing quarantine attributes
const HealingQuarantineId = "healing.quarantine.id"
const HealingQuarantineReason = "healing.quarantine.reason"
const HealingQuarantineDurationMs = "healing.quarantine.duration_ms"
const HealingQuarantineActive = "healing.quarantine.active"

// Iter23: LLM function call routing attributes
const LLMFunctionCallName = "llm.function_call.name"
const LLMFunctionCallRoutingStrategy = "llm.function_call.routing_strategy"
const LLMFunctionCallLatencyMs = "llm.function_call.latency_ms"

// Iter23: ChatmanGPT namespace attributes
const ChatmangptWave = "chatmangpt.wave"
const ChatmangptVersion = "chatmangpt.version"
const ChatmangptDeployment = "chatmangpt.deployment"
