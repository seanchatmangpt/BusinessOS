package semconv

import "go.opentelemetry.io/otel/attribute"

// PM simulation attributes (iter14)
const (
	ProcessMiningSimulationCasesKey      = attribute.Key("process.mining.simulation.cases")
	ProcessMiningSimulationNoiseRateKey  = attribute.Key("process.mining.simulation.noise_rate")
	ProcessMiningSimulationDurationMsKey = attribute.Key("process.mining.simulation.duration_ms")
	ProcessMiningReplayTokenCountKey     = attribute.Key("process.mining.replay.token_count")
)

func ProcessMiningSimulationCases(val int) attribute.KeyValue {
	return ProcessMiningSimulationCasesKey.Int(val)
}

func ProcessMiningSimulationNoiseRate(val float64) attribute.KeyValue {
	return ProcessMiningSimulationNoiseRateKey.Float64(val)
}

func ProcessMiningSimulationDurationMs(val int) attribute.KeyValue {
	return ProcessMiningSimulationDurationMsKey.Int(val)
}

func ProcessMiningReplayTokenCount(val int) attribute.KeyValue {
	return ProcessMiningReplayTokenCountKey.Int(val)
}
