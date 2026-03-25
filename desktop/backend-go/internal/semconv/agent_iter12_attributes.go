// Code generated from semconv/model/agent/registry.yaml. DO NOT EDIT.
// Wave 9 iteration 12

package semconv

import "go.opentelemetry.io/otel/attribute"

// Agent topology, task status, coordination and messaging attributes (iter12).

const (
	// AgentTopologyTypeKey is the OTel attribute key for agent.topology.type.
	// The topology type of the agent network (e.g., star, mesh, hierarchical).
	AgentTopologyTypeKey = attribute.Key("agent.topology.type")
	// AgentTaskStatusKey is the OTel attribute key for agent.task.status.
	// Current execution status of the agent task.
	AgentTaskStatusKey = attribute.Key("agent.task.status")
	// AgentCoordinationLatencyMsKey is the OTel attribute key for agent.coordination.latency_ms.
	// Latency in milliseconds for agent-to-agent coordination round-trip.
	AgentCoordinationLatencyMsKey = attribute.Key("agent.coordination.latency_ms")
	// AgentMessageCountKey is the OTel attribute key for agent.message.count.
	// Total number of messages exchanged by the agent in this session.
	AgentMessageCountKey = attribute.Key("agent.message.count")
)

// AgentTopologyType returns an attribute KeyValue for agent.topology.type.
func AgentTopologyType(val string) attribute.KeyValue {
	return AgentTopologyTypeKey.String(val)
}

// AgentTaskStatus returns an attribute KeyValue for agent.task.status.
func AgentTaskStatus(val string) attribute.KeyValue {
	return AgentTaskStatusKey.String(val)
}

// AgentCoordinationLatencyMs returns an attribute KeyValue for agent.coordination.latency_ms.
func AgentCoordinationLatencyMs(val int64) attribute.KeyValue {
	return AgentCoordinationLatencyMsKey.Int64(val)
}

// AgentMessageCount returns an attribute KeyValue for agent.message.count.
func AgentMessageCount(val int64) attribute.KeyValue {
	return AgentMessageCountKey.Int64(val)
}
