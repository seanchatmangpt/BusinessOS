package semconv

import "go.opentelemetry.io/otel/attribute"

const A2ACapabilityMatchScoreKey = attribute.Key("a2a.capability.match_score")
const A2ACapabilityRequiredKey = attribute.Key("a2a.capability.required")
const A2ACapabilityOfferedKey = attribute.Key("a2a.capability.offered")
const A2ARoutingStrategyKey = attribute.Key("a2a.routing.strategy")
const A2AQueueDepthKey = attribute.Key("a2a.queue.depth")
