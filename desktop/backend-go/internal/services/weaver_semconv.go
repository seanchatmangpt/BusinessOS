package services

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const weaverSemconvPrefix = "semconv."

var (
	weaverProcMu     sync.Mutex
	weaverProc       *weaverMCPProcess
	weaverToolsMu    sync.Mutex
	weaverToolsCache []MCPTool
	weaverToolsReady bool
)

// WeaverSemconvEnabled returns true when BusinessOS should attach the local weaver MCP
// (stdio) for the ChatmanGPT semconv registry.
func WeaverSemconvEnabled() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("WEAVER_SEMCONV_ENABLED")))
	return v == "1" || v == "true" || v == "yes"
}

// WeaverSemconvRegistryPath resolves the registry directory passed to `weaver registry mcp -r`.
func WeaverSemconvRegistryPath() string {
	if p := strings.TrimSpace(os.Getenv("WEAVER_SEMCONV_REGISTRY")); p != "" {
		return filepath.Clean(p)
	}
	return "/semconv/model"
}

// WeaverBinary returns the weaver executable path (default: look up "weaver" on PATH).
func WeaverBinary() string {
	if p := strings.TrimSpace(os.Getenv("WEAVER_BIN")); p != "" {
		return p
	}
	return "weaver"
}

func weaverSemconvConfigured() bool {
	if !WeaverSemconvEnabled() {
		return false
	}
	reg := WeaverSemconvRegistryPath()
	st, err := os.Stat(reg)
	if err != nil || !st.IsDir() {
		return false
	}
	return true
}

func appendWeaverSemconvTools(ctx context.Context, dest []MCPTool) []MCPTool {
	if !weaverSemconvConfigured() {
		return dest
	}
	tools, err := loadWeaverSemconvTools(ctx)
	if err != nil {
		slog.Debug("weaver semconv MCP tools not loaded", "error", err)
		return dest
	}
	return append(dest, tools...)
}

func loadWeaverSemconvTools(ctx context.Context) ([]MCPTool, error) {
	weaverToolsMu.Lock()
	defer weaverToolsMu.Unlock()
	if weaverToolsReady {
		return weaverToolsCache, nil
	}
	p, err := ensureWeaverMCP(ctx)
	if err != nil {
		return nil, err
	}
	rawTools, err := p.listTools(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]MCPTool, 0, len(rawTools))
	for _, t := range rawTools {
		desc := t.Description
		if desc == "" {
			desc = "OpenTelemetry semantic convention tool (weaver MCP)"
		}
		out = append(out, MCPTool{
			Name:        weaverSemconvPrefix + t.Name,
			Description: desc,
			Parameters:  t.InputSchema,
			Source:      "weaver-mcp",
		})
	}
	weaverToolsCache = out
	weaverToolsReady = true
	return out, nil
}

func ensureWeaverMCP(ctx context.Context) (*weaverMCPProcess, error) {
	weaverProcMu.Lock()
	defer weaverProcMu.Unlock()
	if weaverProc != nil {
		return weaverProc, nil
	}
	bin := WeaverBinary()
	reg := WeaverSemconvRegistryPath()
	p, err := startWeaverMCPProcess(ctx, bin, reg)
	if err != nil {
		return nil, fmt.Errorf("start weaver mcp: %w", err)
	}
	weaverProc = p
	slog.Info("Weaver semconv MCP connected", "registry", reg, "weaver", bin)
	return weaverProc, nil
}

// ExecuteWeaverSemconvTool runs a tool exposed by weaver registry mcp (names must use semconv. prefix).
func ExecuteWeaverSemconvTool(ctx context.Context, namespacedTool string, arguments map[string]interface{}) (interface{}, error) {
	if !strings.HasPrefix(namespacedTool, weaverSemconvPrefix) {
		return nil, fmt.Errorf("not a weaver semconv tool: %s", namespacedTool)
	}
	underlying := strings.TrimPrefix(namespacedTool, weaverSemconvPrefix)
	if underlying == "" {
		return nil, fmt.Errorf("empty weaver tool name")
	}
	p, err := ensureWeaverMCP(ctx)
	if err != nil {
		return nil, err
	}
	return p.callTool(ctx, underlying, arguments)
}

// ResetWeaverSemconvForTests closes the MCP child process and clears caches (tests only).
func ResetWeaverSemconvForTests() {
	weaverProcMu.Lock()
	if weaverProc != nil {
		_ = weaverProc.Close()
		weaverProc = nil
	}
	weaverProcMu.Unlock()
	weaverToolsMu.Lock()
	weaverToolsCache = nil
	weaverToolsReady = false
	weaverToolsMu.Unlock()
}
