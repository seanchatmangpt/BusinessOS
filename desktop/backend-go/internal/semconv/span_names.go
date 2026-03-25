// Code generated from semconv/model/*/spans.yaml. DO NOT EDIT.
// Regenerate with: weaver registry generate (or update manually from spans.yaml)

package semconv

// Span names for ChatmanGPT operations.
// Use these constants instead of raw strings to enforce schema contracts.
//
// Example usage:
//   import "go.opentelemetry.io/otel"
//   tracer := otel.Tracer("businessos")
//   ctx, span := tracer.Start(ctx, SpanNameBosComplianceCheck)
//   defer span.End()
//   span.SetAttributes(BosComplianceFramework(BosComplianceFrameworkValues.Soc2))
const (
	// Healing domain
	SpanNameHealingDiagnosis = "healing.diagnosis"
	SpanNameHealingReflexArc = "healing.reflex_arc"

	// Agent domain
	SpanNameAgentDecision   = "agent.decision"
	SpanNameAgentLlmPredict = "agent.llm_predict"

	// Consensus domain (HotStuff BFT)
	SpanNameConsensusRound = "consensus.round"

	// MCP domain
	SpanNameMcpCall        = "mcp.call"
	SpanNameMcpToolExecute = "mcp.tool_execute"

	// A2A domain
	SpanNameA2ACall       = "a2a.call"
	SpanNameA2ACreateDeal = "a2a.create_deal"

	// Canopy domain
	SpanNameCanopyHeartbeat   = "canopy.heartbeat"
	SpanNameCanopyAdapterCall = "canopy.adapter_call"

	// Workflow domain (YAWL)
	SpanNameWorkflowExecute    = "workflow.execute"
	SpanNameWorkflowTransition = "workflow.transition"

	// Process Mining domain
	SpanNameProcessMiningDiscovery = "process.mining.discovery"
	SpanNameConformanceCheck       = "conformance.check"

	// BusinessOS domain
	SpanNameBosComplianceCheck    = "bos.compliance.check"
	SpanNameBosDecisionRecord     = "bos.decision.record"
	SpanNameBosWorkspaceOperation = "bos.workspace.operation"
	SpanNameBosAuditRecord        = "bos.audit.record"
	SpanNameBosGapDetect          = "bos.gap.detect"
)
