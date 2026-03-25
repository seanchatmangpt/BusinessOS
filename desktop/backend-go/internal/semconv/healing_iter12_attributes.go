// Code generated from semconv/model/healing/registry.yaml. DO NOT EDIT.
// Wave 9 iteration 12

package semconv

import "go.opentelemetry.io/otel/attribute"

// Healing MTTR, escalation and repair strategy attributes (iter12).

const (
	// HealingEscalationLevelKey is the OTel attribute key for healing.escalation.level.
	// Escalation level reached during the healing operation.
	HealingEscalationLevelKey = attribute.Key("healing.escalation.level")
	// HealingRepairStrategyKey is the OTel attribute key for healing.repair.strategy.
	// The repair strategy selected for the current failure mode.
	HealingRepairStrategyKey = attribute.Key("healing.repair.strategy")
	// HealingAttemptKey is the OTel attribute key for healing.attempt.
	// Current healing attempt number (1-indexed).
	HealingAttemptKey = attribute.Key("healing.attempt")
)

// HealingEscalationLevel returns an attribute KeyValue for healing.escalation.level.
func HealingEscalationLevel(val string) attribute.KeyValue {
	return HealingEscalationLevelKey.String(val)
}

// HealingRepairStrategy returns an attribute KeyValue for healing.repair.strategy.
func HealingRepairStrategy(val string) attribute.KeyValue {
	return HealingRepairStrategyKey.String(val)
}

// HealingAttempt returns an attribute KeyValue for healing.attempt.
func HealingAttempt(val int64) attribute.KeyValue {
	return HealingAttemptKey.Int64(val)
}
