package semconv

import "go.opentelemetry.io/otel/attribute"

// Iter32 — Agent workflow checkpoint, A2A contract amendment, PM replay comparison,
//           consensus epoch quorum snapshot, healing backpressure, LLM few-shot

// Agent Workflow Checkpoint
const AgentWorkflowCheckpointIDKey = attribute.Key("agent.workflow.checkpoint_id")
const AgentWorkflowCheckpointStepKey = attribute.Key("agent.workflow.checkpoint_step")
const AgentWorkflowResumeCountKey = attribute.Key("agent.workflow.resume_count")

func AgentWorkflowCheckpointID(val string) attribute.KeyValue   { return AgentWorkflowCheckpointIDKey.String(val) }
func AgentWorkflowCheckpointStep(val int64) attribute.KeyValue  { return AgentWorkflowCheckpointStepKey.Int64(val) }
func AgentWorkflowResumeCount(val int64) attribute.KeyValue     { return AgentWorkflowResumeCountKey.Int64(val) }

// A2A Contract Amendment
const A2AContractAmendmentIDKey = attribute.Key("a2a.contract.amendment.id")
const A2AContractAmendmentReasonKey = attribute.Key("a2a.contract.amendment.reason")
const A2AContractAmendmentVersionKey = attribute.Key("a2a.contract.amendment.version")

func A2AContractAmendmentID(val string) attribute.KeyValue      { return A2AContractAmendmentIDKey.String(val) }
func A2AContractAmendmentReason(val string) attribute.KeyValue  { return A2AContractAmendmentReasonKey.String(val) }
func A2AContractAmendmentVersion(val int64) attribute.KeyValue  { return A2AContractAmendmentVersionKey.Int64(val) }

// Process Mining Replay Comparison
const ProcessMiningReplayComparisonIDKey = attribute.Key("process.mining.replay.comparison_id")
const ProcessMiningReplayComparisonBaselineFitnessKey = attribute.Key("process.mining.replay.comparison.baseline_fitness")
const ProcessMiningReplayComparisonTargetFitnessKey = attribute.Key("process.mining.replay.comparison.target_fitness")
const ProcessMiningReplayComparisonDeltaKey = attribute.Key("process.mining.replay.comparison.delta")

func ProcessMiningReplayComparisonID(val string) attribute.KeyValue             { return ProcessMiningReplayComparisonIDKey.String(val) }
func ProcessMiningReplayComparisonBaselineFitness(val float64) attribute.KeyValue { return ProcessMiningReplayComparisonBaselineFitnessKey.Float64(val) }
func ProcessMiningReplayComparisonTargetFitness(val float64) attribute.KeyValue  { return ProcessMiningReplayComparisonTargetFitnessKey.Float64(val) }
func ProcessMiningReplayComparisonDelta(val float64) attribute.KeyValue          { return ProcessMiningReplayComparisonDeltaKey.Float64(val) }

// Consensus Epoch Quorum Snapshot
const ConsensusEpochQuorumSnapshotRoundKey = attribute.Key("consensus.epoch.quorum_snapshot_round")
const ConsensusEpochQuorumSnapshotSizeKey = attribute.Key("consensus.epoch.quorum_snapshot_size")
const ConsensusEpochQuorumSnapshotHashKey = attribute.Key("consensus.epoch.quorum_snapshot_hash")

func ConsensusEpochQuorumSnapshotRound(val int64) attribute.KeyValue  { return ConsensusEpochQuorumSnapshotRoundKey.Int64(val) }
func ConsensusEpochQuorumSnapshotSize(val int64) attribute.KeyValue   { return ConsensusEpochQuorumSnapshotSizeKey.Int64(val) }
func ConsensusEpochQuorumSnapshotHash(val string) attribute.KeyValue  { return ConsensusEpochQuorumSnapshotHashKey.String(val) }

// Healing Backpressure
const HealingBackpressureLevelKey = attribute.Key("healing.backpressure.level")
const HealingBackpressureQueueDepthKey = attribute.Key("healing.backpressure.queue_depth")
const HealingBackpressureDropRateKey = attribute.Key("healing.backpressure.drop_rate")

func HealingBackpressureLevel(val string) attribute.KeyValue      { return HealingBackpressureLevelKey.String(val) }
func HealingBackpressureQueueDepth(val int64) attribute.KeyValue  { return HealingBackpressureQueueDepthKey.Int64(val) }
func HealingBackpressureDropRate(val float64) attribute.KeyValue  { return HealingBackpressureDropRateKey.Float64(val) }

// LLM Few-Shot
const LLMFewShotExampleCountKey = attribute.Key("llm.few_shot.example_count")
const LLMFewShotSelectionStrategyKey = attribute.Key("llm.few_shot.selection_strategy")
const LLMFewShotRetrievalMsKey = attribute.Key("llm.few_shot.retrieval_ms")

func LLMFewShotExampleCount(val int64) attribute.KeyValue         { return LLMFewShotExampleCountKey.Int64(val) }
func LLMFewShotSelectionStrategy(val string) attribute.KeyValue   { return LLMFewShotSelectionStrategyKey.String(val) }
func LLMFewShotRetrievalMs(val int64) attribute.KeyValue          { return LLMFewShotRetrievalMsKey.Int64(val) }
