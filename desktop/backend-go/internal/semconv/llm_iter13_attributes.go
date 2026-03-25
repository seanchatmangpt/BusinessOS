package semconv

import "go.opentelemetry.io/otel/attribute"

const LlmChainOfThoughtStepsKey = attribute.Key("llm.chain_of_thought.steps")
const LlmChainOfThoughtEnabledKey = attribute.Key("llm.chain_of_thought.enabled")
const LlmToolCallCountKey = attribute.Key("llm.tool.call_count")
const LlmCacheHitKey = attribute.Key("llm.cache.hit")
