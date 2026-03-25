// Code generated from semconv/model/process_mining/registry.yaml. DO NOT EDIT.
// Wave 9 iteration 12

package semconv

import "go.opentelemetry.io/otel/attribute"

// Process mining streaming and drift detection attributes (iter12).

const (
	// ProcessMiningStreamingWindowSizeKey is the OTel attribute key for process_mining.streaming.window_size.
	// Number of events in the streaming window used for online process mining.
	ProcessMiningStreamingWindowSizeKey = attribute.Key("process_mining.streaming.window_size")
	// ProcessMiningStreamingLagMsKey is the OTel attribute key for process_mining.streaming.lag_ms.
	// Current lag in milliseconds between event production and processing.
	ProcessMiningStreamingLagMsKey = attribute.Key("process_mining.streaming.lag_ms")
	// ProcessMiningDriftDetectedKey is the OTel attribute key for process_mining.drift.detected.
	// Whether concept drift has been detected in the process stream.
	ProcessMiningDriftDetectedKey = attribute.Key("process_mining.drift.detected")
	// ProcessMiningDriftSeverityKey is the OTel attribute key for process_mining.drift.severity.
	// Severity level of detected concept drift (low, medium, high, critical).
	ProcessMiningDriftSeverityKey = attribute.Key("process_mining.drift.severity")
)

// ProcessMiningStreamingWindowSize returns an attribute KeyValue for process_mining.streaming.window_size.
func ProcessMiningStreamingWindowSize(val int64) attribute.KeyValue {
	return ProcessMiningStreamingWindowSizeKey.Int64(val)
}

// ProcessMiningStreamingLagMs returns an attribute KeyValue for process_mining.streaming.lag_ms.
func ProcessMiningStreamingLagMs(val int64) attribute.KeyValue {
	return ProcessMiningStreamingLagMsKey.Int64(val)
}

// ProcessMiningDriftDetected returns an attribute KeyValue for process_mining.drift.detected.
func ProcessMiningDriftDetected(val bool) attribute.KeyValue {
	return ProcessMiningDriftDetectedKey.Bool(val)
}

// ProcessMiningDriftSeverity returns an attribute KeyValue for process_mining.drift.severity.
func ProcessMiningDriftSeverity(val string) attribute.KeyValue {
	return ProcessMiningDriftSeverityKey.String(val)
}
