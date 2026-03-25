package semconv

import "go.opentelemetry.io/otel/attribute"

// A2A Trust attributes (iter14)
const (
	A2ATrustScoreKey              = attribute.Key("a2a.trust.score")
	A2AReputationHistoryLengthKey = attribute.Key("a2a.reputation.history_length")
	A2ATrustDecayFactorKey        = attribute.Key("a2a.trust.decay_factor")
	A2ATrustUpdatedAtMsKey        = attribute.Key("a2a.trust.updated_at_ms")
)

func A2ATrustScore(val float64) attribute.KeyValue {
	return A2ATrustScoreKey.Float64(val)
}

func A2AReputationHistoryLength(val int) attribute.KeyValue {
	return A2AReputationHistoryLengthKey.Int(val)
}

func A2ATrustDecayFactor(val float64) attribute.KeyValue {
	return A2ATrustDecayFactorKey.Float64(val)
}

func A2ATrustUpdatedAtMs(val int) attribute.KeyValue {
	return A2ATrustUpdatedAtMsKey.Int(val)
}
