package semconv

import "go.opentelemetry.io/otel/attribute"

// Healing pattern attributes (iter14)
const (
	HealingPatternIdKey            = attribute.Key("healing.pattern.id")
	HealingPatternLibrarySizeKey   = attribute.Key("healing.pattern.library_size")
	HealingPatternMatchConfidenceKey = attribute.Key("healing.pattern.match_confidence")
)

func HealingPatternId(val string) attribute.KeyValue {
	return HealingPatternIdKey.String(val)
}

func HealingPatternLibrarySize(val int) attribute.KeyValue {
	return HealingPatternLibrarySizeKey.Int(val)
}

func HealingPatternMatchConfidence(val float64) attribute.KeyValue {
	return HealingPatternMatchConfidenceKey.Float64(val)
}
