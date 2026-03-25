// Code generated from semconv/model/llm/registry.yaml. DO NOT EDIT.
// Wave 9 iteration 12

package semconv

import "go.opentelemetry.io/otel/attribute"

// LLM safety, guardrail, context and retry attributes (iter12).

const (
	// LlmSafetyScoreKey is the OTel attribute key for llm.safety.score.
	// Safety score assigned to the LLM response in range [0.0, 1.0]; higher is safer.
	LlmSafetyScoreKey = attribute.Key("llm.safety.score")
	// LlmGuardrailTriggeredKey is the OTel attribute key for llm.guardrail.triggered.
	// Whether a guardrail was triggered during the LLM request or response.
	LlmGuardrailTriggeredKey = attribute.Key("llm.guardrail.triggered")
	// LlmGuardrailTypeKey is the OTel attribute key for llm.guardrail.type.
	// The type of guardrail that was triggered (e.g., content, pii, toxicity).
	LlmGuardrailTypeKey = attribute.Key("llm.guardrail.type")
	// LlmContextMessagesCountKey is the OTel attribute key for llm.context.messages_count.
	// Number of messages in the context window sent to the LLM.
	LlmContextMessagesCountKey = attribute.Key("llm.context.messages_count")
	// LlmRetryCountKey is the OTel attribute key for llm.retry.count.
	// Number of retries performed for the LLM request.
	LlmRetryCountKey = attribute.Key("llm.retry.count")
)

// LlmSafetyScore returns an attribute KeyValue for llm.safety.score.
func LlmSafetyScore(val float64) attribute.KeyValue {
	return LlmSafetyScoreKey.Float64(val)
}

// LlmGuardrailTriggered returns an attribute KeyValue for llm.guardrail.triggered.
func LlmGuardrailTriggered(val bool) attribute.KeyValue {
	return LlmGuardrailTriggeredKey.Bool(val)
}

// LlmGuardrailType returns an attribute KeyValue for llm.guardrail.type.
func LlmGuardrailType(val string) attribute.KeyValue {
	return LlmGuardrailTypeKey.String(val)
}

// LlmContextMessagesCount returns an attribute KeyValue for llm.context.messages_count.
func LlmContextMessagesCount(val int64) attribute.KeyValue {
	return LlmContextMessagesCountKey.Int64(val)
}

// LlmRetryCount returns an attribute KeyValue for llm.retry.count.
func LlmRetryCount(val int64) attribute.KeyValue {
	return LlmRetryCountKey.Int64(val)
}
