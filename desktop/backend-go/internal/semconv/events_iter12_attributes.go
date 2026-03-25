// Code generated from semconv/model/event/registry.yaml. DO NOT EDIT.
// Wave 9 iteration 12

package semconv

import "go.opentelemetry.io/otel/attribute"

// Event delivery, handler and schema version attributes (iter12).

const (
	// EventDeliveryStatusKey is the OTel attribute key for event.delivery.status.
	// Delivery status of the event (e.g., delivered, failed, pending, retrying).
	EventDeliveryStatusKey = attribute.Key("event.delivery.status")
	// EventHandlerCountKey is the OTel attribute key for event.handler.count.
	// Number of handlers that processed this event.
	EventHandlerCountKey = attribute.Key("event.handler.count")
	// EventSchemaVersionKey is the OTel attribute key for event.schema.version.
	// Schema version of the event payload used for deserialization.
	EventSchemaVersionKey = attribute.Key("event.schema.version")
)

// EventDeliveryStatus returns an attribute KeyValue for event.delivery.status.
func EventDeliveryStatus(val string) attribute.KeyValue {
	return EventDeliveryStatusKey.String(val)
}

// EventHandlerCount returns an attribute KeyValue for event.handler.count.
func EventHandlerCount(val int64) attribute.KeyValue {
	return EventHandlerCountKey.Int64(val)
}

// EventSchemaVersion returns an attribute KeyValue for event.schema.version.
func EventSchemaVersion(val string) attribute.KeyValue {
	return EventSchemaVersionKey.String(val)
}
