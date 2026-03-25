package semconv

import "go.opentelemetry.io/otel/attribute"

const ConsensusSafetyThresholdKey = attribute.Key("consensus.safety.threshold")
const ConsensusLivenessTimeoutRatioKey = attribute.Key("consensus.liveness.timeout_ratio")
const ConsensusNetworkPartitionDetectedKey = attribute.Key("consensus.network.partition_detected")
