package services

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// weaverMCPProcess runs `weaver registry mcp` and performs MCP JSON-RPC over stdio
// (newline-delimited messages per MCP spec).
type weaverMCPProcess struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Scanner
	mu     sync.Mutex
	nextID int64
}

func startWeaverMCPProcess(ctx context.Context, weaverBin, registryPath string) (*weaverMCPProcess, error) {
	// Long-lived child: do not tie to the caller's request context (cancel would kill the server).
	cmd := exec.Command(weaverBin, "registry", "mcp", "-r", registryPath)
	cmd.Env = os.Environ()
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		_ = stdin.Close()
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		_ = stdin.Close()
		return nil, err
	}
	go func() {
		_, _ = io.Copy(io.Discard, stderr)
	}()
	if err := cmd.Start(); err != nil {
		_ = stdin.Close()
		return nil, err
	}
	sc := bufio.NewScanner(stdoutPipe)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	p := &weaverMCPProcess{cmd: cmd, stdin: stdin, stdout: sc}
	if err := p.handshake(ctx); err != nil {
		_ = p.Close()
		return nil, err
	}
	return p, nil
}

func (p *weaverMCPProcess) Close() error {
	_ = p.stdin.Close()
	if p.cmd != nil && p.cmd.Process != nil {
		_ = p.cmd.Process.Kill()
	}
	if p.cmd != nil {
		_, _ = p.cmd.Process.Wait()
	}
	return nil
}

func (p *weaverMCPProcess) handshake(ctx context.Context) error {
	_, err := p.rpc(ctx, "initialize", map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]interface{}{},
		"clientInfo": map[string]interface{}{
			"name":    "businessos",
			"version": "1.0.0",
		},
	})
	if err != nil {
		return fmt.Errorf("weaver mcp initialize: %w", err)
	}
	if err := p.notify(ctx, "notifications/initialized", nil); err != nil {
		return fmt.Errorf("weaver mcp initialized notify: %w", err)
	}
	return nil
}

func (p *weaverMCPProcess) notify(ctx context.Context, method string, params interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	msg := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
	}
	if params != nil {
		msg["params"] = params
	}
	line, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	if _, err := p.stdin.Write(append(line, '\n')); err != nil {
		return err
	}
	return nil
}

func (p *weaverMCPProcess) rpc(ctx context.Context, method string, params interface{}) (json.RawMessage, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	id := atomic.AddInt64(&p.nextID, 1)
	msg := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"method":  method,
	}
	if params != nil {
		msg["params"] = params
	}
	line, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	if _, err := p.stdin.Write(append(line, '\n')); err != nil {
		return nil, err
	}
	return p.readUntilID(ctx, id)
}

func (p *weaverMCPProcess) readUntilID(ctx context.Context, wantID int64) (json.RawMessage, error) {
	deadline := time.Now().Add(60 * time.Second)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	for {
		if time.Now().After(deadline) {
			return nil, context.DeadlineExceeded
		}
		if !p.stdout.Scan() {
			if err := p.stdout.Err(); err != nil {
				return nil, err
			}
			return nil, errors.New("weaver mcp: stdout closed before response")
		}
		raw := p.stdout.Bytes()
		if len(bytes.TrimSpace(raw)) == 0 {
			continue
		}
		var envelope map[string]json.RawMessage
		if err := json.Unmarshal(raw, &envelope); err != nil {
			slog.Debug("weaver mcp: skip non-json line", "line", truncateForLog(string(raw), 200))
			continue
		}
		idRaw, hasID := envelope["id"]
		if !hasID {
			continue
		}
		var idFloat float64
		if err := json.Unmarshal(idRaw, &idFloat); err != nil {
			continue
		}
		if int64(idFloat) != wantID {
			continue
		}
		if errRaw, ok := envelope["error"]; ok {
			var rpcErr struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			}
			_ = json.Unmarshal(errRaw, &rpcErr)
			return nil, fmt.Errorf("weaver mcp rpc error %d: %s", rpcErr.Code, rpcErr.Message)
		}
		if res, ok := envelope["result"]; ok {
			return res, nil
		}
		return nil, errors.New("weaver mcp: response missing result")
	}
}

func truncateForLog(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}

func (p *weaverMCPProcess) listTools(ctx context.Context) ([]MCPClientTool, error) {
	raw, err := p.rpc(ctx, "tools/list", map[string]interface{}{})
	if err != nil {
		return nil, err
	}
	var out struct {
		Tools []MCPClientTool `json:"tools"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("tools/list decode: %w", err)
	}
	return out.Tools, nil
}

func (p *weaverMCPProcess) callTool(ctx context.Context, name string, arguments map[string]interface{}) (interface{}, error) {
	raw, err := p.rpc(ctx, "tools/call", map[string]interface{}{
		"name":      name,
		"arguments": arguments,
	})
	if err != nil {
		return nil, err
	}
	var wrapped struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		IsError bool `json:"isError"`
	}
	if err := json.Unmarshal(raw, &wrapped); err != nil {
		return nil, fmt.Errorf("tools/call decode: %w", err)
	}
	if wrapped.IsError {
		var parts []string
		for _, c := range wrapped.Content {
			parts = append(parts, c.Text)
		}
		return nil, fmt.Errorf("weaver tool error: %s", strings.Join(parts, "; "))
	}
	var texts []string
	for _, c := range wrapped.Content {
		if c.Type == "text" && c.Text != "" {
			texts = append(texts, c.Text)
		}
	}
	if len(texts) == 1 {
		var asJSON interface{}
		if json.Unmarshal([]byte(texts[0]), &asJSON) == nil {
			return asJSON, nil
		}
		return texts[0], nil
	}
	if len(texts) > 1 {
		return strings.Join(texts, "\n"), nil
	}
	var generic interface{}
	if err := json.Unmarshal(raw, &generic); err == nil {
		return generic, nil
	}
	return string(raw), nil
}
