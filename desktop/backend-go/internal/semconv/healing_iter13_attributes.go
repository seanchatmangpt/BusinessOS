package semconv

import "go.opentelemetry.io/otel/attribute"

const HealingCascadeDetectedKey = attribute.Key("healing.cascade.detected")
const HealingCascadeDepthKey = attribute.Key("healing.cascade.depth")
const HealingRootCauseIdKey = attribute.Key("healing.root_cause.id")
