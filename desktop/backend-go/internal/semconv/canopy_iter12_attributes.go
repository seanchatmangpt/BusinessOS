// Code generated from semconv/model/canopy/registry.yaml. DO NOT EDIT.
// Wave 9 iteration 12

package semconv

import "go.opentelemetry.io/otel/attribute"

// Canopy protocol, sync strategy, conflict and peer count attributes (iter12).

const (
	// CanopyProtocolVersionKey is the OTel attribute key for canopy.protocol.version.
	// Version of the Canopy workspace protocol in use.
	CanopyProtocolVersionKey = attribute.Key("canopy.protocol.version")
	// CanopySyncStrategyKey is the OTel attribute key for canopy.sync.strategy.
	// Synchronization strategy applied for conflict resolution.
	CanopySyncStrategyKey = attribute.Key("canopy.sync.strategy")
	// CanopyConflictCountKey is the OTel attribute key for canopy.conflict.count.
	// Number of conflicts detected during the sync operation.
	CanopyConflictCountKey = attribute.Key("canopy.conflict.count")
	// CanopyPeerCountKey is the OTel attribute key for canopy.peer.count.
	// Number of active peers connected to this Canopy workspace node.
	CanopyPeerCountKey = attribute.Key("canopy.peer.count")
)

// CanopyProtocolVersion returns an attribute KeyValue for canopy.protocol.version.
func CanopyProtocolVersion(val string) attribute.KeyValue {
	return CanopyProtocolVersionKey.String(val)
}

// CanopySyncStrategy returns an attribute KeyValue for canopy.sync.strategy.
func CanopySyncStrategy(val string) attribute.KeyValue {
	return CanopySyncStrategyKey.String(val)
}

// CanopyConflictCount returns an attribute KeyValue for canopy.conflict.count.
func CanopyConflictCount(val int64) attribute.KeyValue {
	return CanopyConflictCountKey.Int64(val)
}

// CanopyPeerCount returns an attribute KeyValue for canopy.peer.count.
func CanopyPeerCount(val int64) attribute.KeyValue {
	return CanopyPeerCountKey.Int64(val)
}
