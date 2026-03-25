// Package semconv provides OTel attribute helpers generated from the ChatmanGPT registry.
//
// Usage example showing Chicago TDD-compliant span instrumentation for compliance checking:
//
//	import (
//	  "context"
//	  "go.opentelemetry.io/otel"
//	  "go.opentelemetry.io/otel/codes"
//	  "github.com/rhl/businessos-backend/internal/semconv"
//	)
//
//	func checkComplianceRule(ctx context.Context, ruleID, framework, severity string) error {
//	  tracer := otel.Tracer("businessos")
//	  ctx, span := tracer.Start(ctx, "bos.compliance.check")
//	  defer span.End()
//
//	  span.SetAttributes(
//	    semconv.BosComplianceRuleId(ruleID),
//	    semconv.BosComplianceFramework(framework),
//	    semconv.BosComplianceSeverity(severity),
//	  )
//
//	  // ... check logic ...
//
//	  passed := true // result of evaluation
//	  span.SetAttributes(semconv.BosCompliancePassed(passed))
//
//	  if !passed {
//	    span.SetStatus(codes.Error, "compliance rule violated")
//	  }
//
//	  return nil
//	}
//
// Framework enum values are provided by BosComplianceFrameworkValues:
//
//	semconv.BosComplianceFramework(semconv.BosComplianceFrameworkValues.Soc2)   // "SOC2"
//	semconv.BosComplianceFramework(semconv.BosComplianceFrameworkValues.Hipaa)  // "HIPAA"
//	semconv.BosComplianceFramework(semconv.BosComplianceFrameworkValues.Gdpr)   // "GDPR"
//	semconv.BosComplianceFramework(semconv.BosComplianceFrameworkValues.Sox)    // "SOX"
//
// Severity enum values are provided by BosComplianceSeverityValues:
//
//	semconv.BosComplianceSeverity(semconv.BosComplianceSeverityValues.Critical) // "critical"
//	semconv.BosComplianceSeverity(semconv.BosComplianceSeverityValues.High)     // "high"
//	semconv.BosComplianceSeverity(semconv.BosComplianceSeverityValues.Medium)   // "medium"
//	semconv.BosComplianceSeverity(semconv.BosComplianceSeverityValues.Low)      // "low"
//
// The OTEL infrastructure in BusinessOS:
//
//   - Tracer provider: internal/observability/tracer.go — InitTracer() registers a global
//     OTLP HTTP exporter sending to the configured OTEL collector endpoint.
//     Service name is set to "businessos" via semconv.ServiceName.
//
//   - HTTP middleware: internal/middleware/tracing.go — W3C Trace Context propagation
//     (traceparent header) for inbound HTTP requests. Note: this is a lightweight
//     custom implementation; the global otel.Tracer("businessos") in service code
//     integrates with the SDK-backed provider from InitTracer().
//
//   - Span naming convention: "bos.<domain>.<operation>"
//     Examples: "bos.compliance.check", "bos.compliance.reload_rules",
//               "bos.compliance.get_status", "bos.compliance.audit_chain"
//
// Traces are viewable in Jaeger UI at http://localhost:16686 (service: "businessos").
package semconv
