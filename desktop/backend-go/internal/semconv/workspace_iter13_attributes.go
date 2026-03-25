package semconv

import "go.opentelemetry.io/otel/attribute"

const WorkspaceOrchestrationPatternKey = attribute.Key("workspace.orchestration.pattern")
const WorkspaceTaskQueueDepthKey = attribute.Key("workspace.task.queue.depth")
const WorkspaceIterationCountKey = attribute.Key("workspace.iteration.count")
