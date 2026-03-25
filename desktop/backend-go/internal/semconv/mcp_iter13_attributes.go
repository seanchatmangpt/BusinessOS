package semconv

import "go.opentelemetry.io/otel/attribute"

const McpToolVersionKey = attribute.Key("mcp.tool.version")
const McpToolSchemaHashKey = attribute.Key("mcp.tool.schema_hash")
const McpSessionIdKey = attribute.Key("mcp.session.id")
