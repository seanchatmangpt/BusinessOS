package semconv

import "go.opentelemetry.io/otel/attribute"

// Canopy snapshot attributes (iter14)
const (
	CanopySnapshotIdKey        = attribute.Key("canopy.snapshot.id")
	CanopySnapshotSizeBytesKey = attribute.Key("canopy.snapshot.size_bytes")
)

func CanopySnapshotId(val string) attribute.KeyValue {
	return CanopySnapshotIdKey.String(val)
}

func CanopySnapshotSizeBytes(val int) attribute.KeyValue {
	return CanopySnapshotSizeBytesKey.Int(val)
}
