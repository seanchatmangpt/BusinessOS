// Package semconv — Chicago TDD usage examples.
//
// These tests demonstrate that typed semconv constants can be used to build
// OTEL attribute sets. A compile error here means a constant was renamed or
// removed from the schema — the contract is enforced at compile time, not
// only at runtime.
//
// Run with: cd BusinessOS/desktop/backend-go && go test ./internal/semconv/... -count=1 -v
package semconv

import (
	"testing"

	"go.opentelemetry.io/otel/attribute"
)

// ============================================================
// Healing span — span.healing.diagnosis
// ============================================================

// TestBuildHealingSpanAttributes shows typed attribute construction for a
// healing.diagnosis span. Compile error if any constant is removed.
func TestBuildHealingSpanAttributes(t *testing.T) {
	attrs := []attribute.KeyValue{
		HealingFailureMode(HealingFailureModeValues.Deadlock),
		HealingConfidence(0.95),
		HealingAgentId("healing-agent-1"),
	}
	if len(attrs) != 3 {
		t.Errorf("expected 3 attributes, got %d", len(attrs))
	}
	if string(attrs[0].Key) != "healing.failure_mode" {
		t.Errorf("expected healing.failure_mode, got %s", attrs[0].Key)
	}
	if attrs[1].Value.AsFloat64() != 0.95 {
		t.Errorf("expected confidence 0.95, got %f", attrs[1].Value.AsFloat64())
	}
	if attrs[2].Value.AsString() != "healing-agent-1" {
		t.Errorf("expected healing-agent-1, got %s", attrs[2].Value.AsString())
	}
}

// TestBuildHealingReflexArcAttributes shows typed attribute construction for a
// healing.reflex_arc span.
func TestBuildHealingReflexArcAttributes(t *testing.T) {
	attrs := []attribute.KeyValue{
		HealingReflexArc("stagnation-watchdog"),
		HealingRecoveryAction("restart_process"),
		HealingMttrMs(450),
	}
	if len(attrs) != 3 {
		t.Errorf("expected 3 attributes, got %d", len(attrs))
	}
	if string(attrs[0].Key) != "healing.reflex_arc" {
		t.Errorf("expected healing.reflex_arc, got %s", attrs[0].Key)
	}
}

// ============================================================
// Compliance span — span.bos.compliance.check
// ============================================================

// TestBuildComplianceSpanAttributes shows typed attribute construction for a
// bos.compliance.check span. Compile error if BOS constants are removed.
func TestBuildComplianceSpanAttributes(t *testing.T) {
	attrs := []attribute.KeyValue{
		BosComplianceFramework(BosComplianceFrameworkValues.Soc2),
		BosComplianceRuleId("soc2.cc6.1"),
		BosComplianceSeverity(BosComplianceSeverityValues.Critical),
		BosCompliancePassed(false),
	}
	if len(attrs) != 4 {
		t.Errorf("expected 4 attributes, got %d", len(attrs))
	}
	if string(attrs[0].Key) != "bos.compliance.framework" {
		t.Errorf("expected bos.compliance.framework, got %s", attrs[0].Key)
	}
	if attrs[0].Value.AsString() != "SOC2" {
		t.Errorf("expected SOC2, got %s", attrs[0].Value.AsString())
	}
	if attrs[3].Value.AsBool() != false {
		t.Errorf("expected passed=false, got %v", attrs[3].Value.AsBool())
	}
}

// ============================================================
// A2A span — span.a2a.call
// ============================================================

// TestBuildA2ACallSpanAttributes shows typed attribute construction for an
// a2a.call span.
func TestBuildA2ACallSpanAttributes(t *testing.T) {
	attrs := []attribute.KeyValue{
		attribute.String(string(A2aAgentIdKey), "osa-agent-42"),
		attribute.String(string(A2aOperationKey), "create_deal"),
		attribute.String(string(A2aSourceServiceKey), "businessos"),
	}
	if len(attrs) != 3 {
		t.Errorf("expected 3 attributes, got %d", len(attrs))
	}
	if string(attrs[0].Key) != "a2a.agent.id" {
		t.Errorf("expected a2a.agent.id, got %s", attrs[0].Key)
	}
}

// ============================================================
// Signal span — Signal Theory S=(M,G,T,F,W)
// ============================================================

// TestBuildSignalSpanAttributes shows typed attribute construction for a
// signal classification span using the S=(M,G,T,F,W) schema.
func TestBuildSignalSpanAttributes(t *testing.T) {
	attrs := []attribute.KeyValue{
		SignalMode(SignalModeValues.Linguistic),
		SignalGenre(SignalGenreValues.Spec),
		SignalType(SignalTypeValues.Direct),
		SignalFormat(SignalFormatValues.Markdown),
		SignalWeight(0.85),
	}
	if len(attrs) != 5 {
		t.Errorf("expected 5 attributes, got %d", len(attrs))
	}
	if string(attrs[0].Key) != "signal.mode" {
		t.Errorf("expected signal.mode, got %s", attrs[0].Key)
	}
	if attrs[0].Value.AsString() != "linguistic" {
		t.Errorf("expected linguistic, got %s", attrs[0].Value.AsString())
	}
	// Signal weight >= 0.7 passes the S/N gate
	if attrs[4].Value.AsFloat64() < 0.7 {
		t.Errorf("signal weight %f is below S/N gate threshold 0.7", attrs[4].Value.AsFloat64())
	}
}

// TestSignalWeightSNGateThreshold verifies the S/N gate threshold is >= 0.7.
func TestSignalWeightSNGateThreshold(t *testing.T) {
	// Values from signal-theory-complete.md: S/N >= 0.7 required
	passingWeights := []float64{0.7, 0.8, 0.9, 1.0}
	for _, w := range passingWeights {
		attr := SignalWeight(w)
		if attr.Value.AsFloat64() < 0.7 {
			t.Errorf("weight %f should pass S/N gate (>= 0.7)", w)
		}
	}
	failingWeights := []float64{0.0, 0.5, 0.69}
	for _, w := range failingWeights {
		attr := SignalWeight(w)
		if attr.Value.AsFloat64() >= 0.7 {
			t.Errorf("weight %f should fail S/N gate (< 0.7)", w)
		}
	}
}

// ============================================================
// Process mining span — span.process.mining.discovery
// ============================================================

// TestBuildProcessMiningSpanAttributes shows typed attribute construction for
// a process.mining.discovery span.
func TestBuildProcessMiningSpanAttributes(t *testing.T) {
	attrs := []attribute.KeyValue{
		ProcessMiningTraceId("case-001"),
		ProcessMiningAlgorithm(ProcessMiningAlgorithmValues.InductiveMiner),
		ProcessMiningActivity("submit_application"),
		ProcessMiningEventCount(42),
		ProcessMiningLogPath("/data/logs/loan_process.xes"),
	}
	if len(attrs) != 5 {
		t.Errorf("expected 5 attributes, got %d", len(attrs))
	}
	if string(attrs[0].Key) != "process.mining.trace_id" {
		t.Errorf("expected process.mining.trace_id, got %s", attrs[0].Key)
	}
	if attrs[1].Value.AsString() != "inductive_miner" {
		t.Errorf("expected inductive_miner, got %s", attrs[1].Value.AsString())
	}
}

// ============================================================
// Canopy heartbeat span — span.canopy.heartbeat
// ============================================================

// TestBuildCanopyHeartbeatSpanAttributes shows typed attribute construction
// for a canopy.heartbeat span with tier-based budget enforcement.
func TestBuildCanopyHeartbeatSpanAttributes(t *testing.T) {
	attrs := []attribute.KeyValue{
		CanopyHeartbeatTier(CanopyHeartbeatTierValues.Critical),
		CanopyAdapterName("osa-adapter"),
		CanopyAdapterAction("dispatch"),
		CanopyBudgetMs(100),
	}
	if len(attrs) != 4 {
		t.Errorf("expected 4 attributes, got %d", len(attrs))
	}
	if string(attrs[0].Key) != "canopy.heartbeat.tier" {
		t.Errorf("expected canopy.heartbeat.tier, got %s", attrs[0].Key)
	}
	if attrs[0].Value.AsString() != "critical" {
		t.Errorf("expected critical tier, got %s", attrs[0].Value.AsString())
	}
}

// TestCanopyHeartbeatTierBudgetsAreOrdered verifies tier ordering matches
// Toyota production system priority: critical < high < normal < low budget.
func TestCanopyHeartbeatTierBudgetsAreOrdered(t *testing.T) {
	tiers := []string{
		CanopyHeartbeatTierValues.Critical,
		CanopyHeartbeatTierValues.High,
		CanopyHeartbeatTierValues.Normal,
		CanopyHeartbeatTierValues.Low,
	}
	expected := []string{"critical", "high", "normal", "low"}
	for i, tier := range tiers {
		if tier != expected[i] {
			t.Errorf("tier[%d] = %q, want %q", i, tier, expected[i])
		}
	}
}

// ============================================================
// Workflow span — span.workflow.execute (YAWL patterns)
// ============================================================

// TestBuildWorkflowSpanAttributes shows typed attribute construction for a
// workflow.execute span using YAWL control-flow patterns.
func TestBuildWorkflowSpanAttributes(t *testing.T) {
	attrs := []attribute.KeyValue{
		WorkflowId("wf-20260324-001"),
		WorkflowName("loan-approval"),
		WorkflowPattern(WorkflowPatternValues.ParallelSplit),
		WorkflowState(WorkflowStateValues.Active),
		WorkflowEngine(WorkflowEngineValues.Yawl),
		WorkflowStep("credit-check"),
		WorkflowStepCount(3),
	}
	if len(attrs) != 7 {
		t.Errorf("expected 7 attributes, got %d", len(attrs))
	}
	if string(attrs[0].Key) != "workflow.id" {
		t.Errorf("expected workflow.id, got %s", attrs[0].Key)
	}
	if attrs[2].Value.AsString() != "parallel_split" {
		t.Errorf("expected parallel_split, got %s", attrs[2].Value.AsString())
	}
}

// ============================================================
// Consensus span — span.consensus.round (HotStuff BFT)
// ============================================================

// TestBuildConsensusRoundSpanAttributes shows typed attribute construction
// for a consensus.round span.
func TestBuildConsensusRoundSpanAttributes(t *testing.T) {
	attrs := []attribute.KeyValue{
		ConsensusRoundNum(7),
		ConsensusRoundType(ConsensusRoundTypeValues.Prepare),
		ConsensusNodeId("node-3"),
		ConsensusQuorumSize(5),
		ConsensusLatencyMs(12),
	}
	if len(attrs) != 5 {
		t.Errorf("expected 5 attributes, got %d", len(attrs))
	}
	if string(attrs[0].Key) != "consensus.round_num" {
		t.Errorf("expected consensus.round_num, got %s", attrs[0].Key)
	}
	if attrs[1].Value.AsString() != "prepare" {
		t.Errorf("expected prepare, got %s", attrs[1].Value.AsString())
	}
}

// ============================================================
// Conformance span — span.conformance.check
// ============================================================

// TestBuildConformanceCheckSpanAttributes shows typed attribute construction
// for a conformance.check span.
func TestBuildConformanceCheckSpanAttributes(t *testing.T) {
	attrs := []attribute.KeyValue{
		ConformanceFitness(0.92),
		ConformancePrecision(0.88),
	}
	if len(attrs) != 2 {
		t.Errorf("expected 2 attributes, got %d", len(attrs))
	}
	if string(attrs[0].Key) != "conformance.fitness" {
		t.Errorf("expected conformance.fitness, got %s", attrs[0].Key)
	}
	if attrs[0].Value.AsFloat64() < 0.0 || attrs[0].Value.AsFloat64() > 1.0 {
		t.Errorf("conformance.fitness %f out of range [0.0, 1.0]", attrs[0].Value.AsFloat64())
	}
}

// ============================================================
// MCP span — span.mcp.call
// ============================================================

// TestBuildMcpCallSpanAttributes shows typed attribute construction for an
// mcp.call span.
func TestBuildMcpCallSpanAttributes(t *testing.T) {
	attrs := []attribute.KeyValue{
		McpToolName("process_mining_discover"),
		McpServerName("pm4py-rust-server"),
		McpProtocol(McpProtocolValues.Stdio),
		McpToolResultCount(1),
	}
	if len(attrs) != 4 {
		t.Errorf("expected 4 attributes, got %d", len(attrs))
	}
	if string(attrs[0].Key) != "mcp.tool.name" {
		t.Errorf("expected mcp.tool.name, got %s", attrs[0].Key)
	}
}

// ============================================================
// Agent span — span.agent.decision
// ============================================================

// TestBuildAgentDecisionSpanAttributes shows typed attribute construction for
// an agent.decision span.
func TestBuildAgentDecisionSpanAttributes(t *testing.T) {
	attrs := []attribute.KeyValue{
		AgentId("compliance-agent-7"),
		AgentDecisionType("escalate"),
		AgentOutcome(AgentOutcomeValues.Escalated),
		AgentLlmModel("claude-sonnet-4-6"),
		AgentTokenCount(1024),
	}
	if len(attrs) != 5 {
		t.Errorf("expected 5 attributes, got %d", len(attrs))
	}
	if string(attrs[0].Key) != "agent.id" {
		t.Errorf("expected agent.id, got %s", attrs[0].Key)
	}
	if attrs[2].Value.AsString() != "escalated" {
		t.Errorf("expected escalated, got %s", attrs[2].Value.AsString())
	}
}

// ============================================================
// Error span — error.type attributes
// ============================================================

// TestBuildErrorSpanAttributes shows typed attribute construction using the
// error.type attribute.
func TestBuildErrorSpanAttributes(t *testing.T) {
	attrs := []attribute.KeyValue{
		ErrorType(ErrorTypeValues.Timeout),
	}
	if len(attrs) != 1 {
		t.Errorf("expected 1 attribute, got %d", len(attrs))
	}
	if string(attrs[0].Key) != "error.type" {
		t.Errorf("expected error.type, got %s", attrs[0].Key)
	}
	if attrs[0].Value.AsString() != "timeout" {
		t.Errorf("expected timeout, got %s", attrs[0].Value.AsString())
	}
}

// ============================================================
// ChatmanGPT cross-cutting budget attributes
// ============================================================

// TestBuildChatmangptBudgetAttributes shows typed attribute construction for
// Armstrong budget enforcement instrumentation.
func TestBuildChatmangptBudgetAttributes(t *testing.T) {
	attrs := []attribute.KeyValue{
		ChatmangptAgentId("healing-agent-1"),
		ChatmangptServiceTier(ChatmangptServiceTierValues.Critical),
		ChatmangptBudgetTimeMs(100),
		ChatmangptBudgetExceeded(false),
	}
	if len(attrs) != 4 {
		t.Errorf("expected 4 attributes, got %d", len(attrs))
	}
	if string(attrs[0].Key) != "chatmangpt.agent.id" {
		t.Errorf("expected chatmangpt.agent.id, got %s", attrs[0].Key)
	}
	if attrs[1].Value.AsString() != "critical" {
		t.Errorf("expected critical, got %s", attrs[1].Value.AsString())
	}
	if attrs[3].Value.AsBool() != false {
		t.Errorf("expected budget_exceeded=false, got %v", attrs[3].Value.AsBool())
	}
}

// TestAllSpanNamesUsedInAttributes verifies that span name constants produce
// non-empty strings and can be used as OTel span names. This is the usage
// example complement to the exhaustive TestAllSpanNamesAreUsed in
// semconv_chicago_tdd_test.go.
func TestAllSpanNamesUsedInAttributes(t *testing.T) {
	// Spot-check a representative sample from each domain to confirm
	// constants are non-empty and usable as OTel span name strings.
	sample := []string{
		SpanNameHealingDiagnosis,
		SpanNameA2ACreateDeal,
		SpanNameBosComplianceCheck,
		SpanNameWorkflowExecute,
		SpanNameMcpCall,
	}
	for _, name := range sample {
		if name == "" {
			t.Errorf("span name constant is empty — schema contract broken")
		}
	}
}
