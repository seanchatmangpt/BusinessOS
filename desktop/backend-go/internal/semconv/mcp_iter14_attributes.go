package semconv

import "go.opentelemetry.io/otel/attribute"

// MCP resource attributes (iter14)
const (
	McpResourceUriKey       = attribute.Key("mcp.resource.uri")
	McpResourceMimeTypeKey  = attribute.Key("mcp.resource.mime_type")
	McpResourceSizeBytesKey = attribute.Key("mcp.resource.size_bytes")
)

func McpResourceUri(val string) attribute.KeyValue {
	return McpResourceUriKey.String(val)
}

func McpResourceMimeType(val string) attribute.KeyValue {
	return McpResourceMimeTypeKey.String(val)
}

func McpResourceSizeBytes(val int) attribute.KeyValue {
	return McpResourceSizeBytesKey.Int(val)
}
