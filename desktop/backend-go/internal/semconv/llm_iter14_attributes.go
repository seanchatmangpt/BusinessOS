package semconv

import "go.opentelemetry.io/otel/attribute"

// LLM token budget attributes (iter14)
const (
	LlmTokenPromptCountKey     = attribute.Key("llm.token.prompt_count")
	LlmTokenCompletionCountKey = attribute.Key("llm.token.completion_count")
	LlmTokenBudgetRemainingKey = attribute.Key("llm.token.budget_remaining")
	LlmModelVersionKey         = attribute.Key("llm.model.version")
)

func LlmTokenPromptCount(val int) attribute.KeyValue {
	return LlmTokenPromptCountKey.Int(val)
}

func LlmTokenCompletionCount(val int) attribute.KeyValue {
	return LlmTokenCompletionCountKey.Int(val)
}

func LlmTokenBudgetRemaining(val int) attribute.KeyValue {
	return LlmTokenBudgetRemainingKey.Int(val)
}

func LlmModelVersion(val string) attribute.KeyValue {
	return LlmModelVersionKey.String(val)
}
