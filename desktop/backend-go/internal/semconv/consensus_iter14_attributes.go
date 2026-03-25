package semconv

import "go.opentelemetry.io/otel/attribute"

// Consensus fault tolerance attributes (iter14)
const (
	ConsensusByzantineFaultsKey  = attribute.Key("consensus.byzantine_faults")
	ConsensusReplicaLagMsKey     = attribute.Key("consensus.replica.lag_ms")
)

func ConsensusByzantineFaults(val int) attribute.KeyValue {
	return ConsensusByzantineFaultsKey.Int(val)
}

func ConsensusReplicaLagMs(val int) attribute.KeyValue {
	return ConsensusReplicaLagMsKey.Int(val)
}
