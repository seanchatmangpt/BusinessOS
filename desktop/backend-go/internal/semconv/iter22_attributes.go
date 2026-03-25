package semconv

// Iter22: Signal batch aggregation attributes
const SignalBatchSize = "signal.batch.size"
const SignalBatchWindowMs = "signal.batch.window_ms"
const SignalBatchDropCount = "signal.batch.drop_count"

// Iter22: Workspace memory compaction attributes
const WorkspaceMemoryCompactionRatio = "workspace.memory.compaction_ratio"
const WorkspaceMemoryCompactionMs = "workspace.memory.compaction_ms"
const WorkspaceMemoryItemsBefore = "workspace.memory.items_before"
const WorkspaceMemoryItemsAfter = "workspace.memory.items_after"

// Iter22: A2A bid evaluation attributes
const A2ABidStrategy = "a2a.bid.strategy"
const A2ABidScore = "a2a.bid.score"
const A2ABidWinnerId = "a2a.bid.winner_id"

// Iter22: PM alignment analysis attributes
const ProcessMiningAlignmentOptimalPathLength = "process.mining.alignment.optimal_path_length"
const ProcessMiningAlignmentMoveCount = "process.mining.alignment.move_count"
const ProcessMiningAlignmentFitnessDelta = "process.mining.alignment.fitness_delta"

// Iter22: Consensus partition recovery attributes
const ConsensusPartitionDetected = "consensus.partition.detected"
const ConsensusPartitionSize = "consensus.partition.size"
const ConsensusPartitionRecoveryMs = "consensus.partition.recovery_ms"
const ConsensusPartitionStrategy = "consensus.partition.strategy"

// Iter22: Healing rollback attributes
const HealingRollbackStrategy = "healing.rollback.strategy"
const HealingRollbackCheckpointId = "healing.rollback.checkpoint_id"
const HealingRollbackRecoveryMs = "healing.rollback.recovery_ms"
const HealingRollbackSuccess = "healing.rollback.success"

// Iter22: LLM structured output attributes
const LLMStructuredOutputSchemaId = "llm.structured_output.schema_id"
const LLMStructuredOutputValidationMs = "llm.structured_output.validation_ms"
