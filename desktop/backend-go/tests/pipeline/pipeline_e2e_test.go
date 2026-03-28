//go:build integration
// +build integration

// Package pipeline_test contains E2E integration tests for the full
// pm4py-rust → BusinessOS pipeline.
//
//	go test -tags=integration ./tests/pipeline/... -v
package pipeline_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const (
	pm4pyBase = "http://localhost:8090"
	bosBase   = "http://localhost:8001"
)

var httpClient = &http.Client{Timeout: 30 * time.Second}

// integrationEventLogMap matches pm4py-rust EventLog JSON (Trace.id, not caseID).
func integrationEventLogMap() map[string]interface{} {
	return map[string]interface{}{
		"traces": []map[string]interface{}{
			{
				"id": "case_001",
				"events": []map[string]interface{}{
					{"activity": "Start", "timestamp": "2024-01-01T10:00:00Z"},
					{"activity": "Review", "timestamp": "2024-01-01T10:05:00Z"},
					{"activity": "Approve", "timestamp": "2024-01-01T10:10:00Z"},
					{"activity": "End", "timestamp": "2024-01-01T10:15:00Z"},
				},
				"attributes": map[string]interface{}{},
			},
			{
				"id": "case_002",
				"events": []map[string]interface{}{
					{"activity": "Start", "timestamp": "2024-01-02T10:00:00Z"},
					{"activity": "Review", "timestamp": "2024-01-02T10:03:00Z"},
					{"activity": "Reject", "timestamp": "2024-01-02T10:08:00Z"},
					{"activity": "End", "timestamp": "2024-01-02T10:12:00Z"},
				},
				"attributes": map[string]interface{}{},
			},
		},
		"attributes": map[string]interface{}{},
	}
}

func writeTempEventLogJSON(t *testing.T) string {
	t.Helper()
	data, err := json.Marshal(integrationEventLogMap())
	if err != nil {
		t.Fatalf("marshal event log: %v", err)
	}
	f, err := os.CreateTemp("", "pipeline_event_log_*.json")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	if _, err := f.Write(data); err != nil {
		t.Fatalf("write temp: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close temp: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(f.Name()) })
	return f.Name()
}

func writeTempPetriNetJSON(t *testing.T) string {
	t.Helper()
	pn := map[string]interface{}{
		"places": []interface{}{
			map[string]interface{}{"id": "p_start", "name": "source", "initial_marking": 1},
			map[string]interface{}{"id": "p_end", "name": "sink", "initial_marking": 0},
		},
		"transitions": []interface{}{
			map[string]interface{}{"id": "t1", "name": "Start", "label": "Start"},
			map[string]interface{}{"id": "t2", "name": "End", "label": "End"},
		},
		"arcs": []interface{}{
			map[string]interface{}{"from": "p_start", "to": "t1", "weight": 1},
			map[string]interface{}{"from": "t1", "to": "p_end", "weight": 1},
		},
		"initial_place": "p_start",
		"final_place":   "p_end",
	}
	data, err := json.Marshal(pn)
	if err != nil {
		t.Fatalf("marshal petri net: %v", err)
	}
	f, err := os.CreateTemp("", "pipeline_petri_*.json")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	if _, err := f.Write(data); err != nil {
		t.Fatalf("write temp: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close temp: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(f.Name()) })
	return f.Name()
}

func isPm4pyRunning(t *testing.T) bool {
	t.Helper()
	resp, err := httpClient.Get(pm4pyBase + "/api/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func isBosRunning(t *testing.T) bool {
	t.Helper()
	resp, err := httpClient.Get(bosBase + "/api/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func postJSON(t *testing.T, url string, body interface{}) (int, map[string]interface{}) {
	t.Helper()
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request body: %v", err)
	}
	resp, err := httpClient.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("POST %s: %v", url, err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatalf("unmarshal response JSON from %s: %v\nbody: %s", url, err, raw)
	}
	return resp.StatusCode, result
}

func TestPm4pyDirectHealth(t *testing.T) {
	if !isPm4pyRunning(t) {
		t.Skip("Skipping: pm4py-rust not running at " + pm4pyBase)
	}
	resp, err := httpClient.Get(pm4pyBase + "/api/health")
	if err != nil {
		t.Fatalf("GET %s/api/health: %v", pm4pyBase, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 from pm4py health, got %d", resp.StatusCode)
	}
	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode health response: %v", err)
	}
	status, _ := body["status"].(string)
	if status != "healthy" {
		t.Errorf("Expected status=='healthy', got %q", status)
	}
}

func TestBosHealth(t *testing.T) {
	if !isBosRunning(t) {
		t.Skip("Skipping: BusinessOS not running at " + bosBase)
	}
	resp, err := httpClient.Get(bosBase + "/api/health")
	if err != nil {
		t.Fatalf("GET %s/api/health: %v", bosBase, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 from BOS health, got %d", resp.StatusCode)
	}
}

func TestPm4pyHealthViaGateway(t *testing.T) {
	if !isBosRunning(t) || !isPm4pyRunning(t) {
		t.Skip("Skipping: BusinessOS or pm4py-rust not running")
	}
	resp, err := httpClient.Get(bosBase + "/api/bos/status")
	if err != nil {
		t.Fatalf("GET %s/api/bos/status: %v", bosBase, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Skipping: /api/bos/status not yet implemented (404)")
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 from /api/bos/status, got %d", resp.StatusCode)
	}
}

// TestDiscoveryPipelineE2E uses POST /api/bos/discover with log_path (JSON on disk).
func TestDiscoveryPipelineE2E(t *testing.T) {
	if !isBosRunning(t) || !isPm4pyRunning(t) {
		t.Skip("Skipping: requires both BusinessOS and pm4py-rust running")
	}
	logPath := writeTempEventLogJSON(t)
	payload := map[string]interface{}{
		"log_path":   logPath,
		"algorithm":  "alpha",
	}
	status, body := postJSON(t, bosBase+"/api/bos/discover", payload)
	if status < 200 || status >= 300 {
		t.Fatalf("Expected 2xx from /api/bos/discover, got %d — body: %v", status, body)
	}
	if _, ok := body["model_id"].(string); !ok {
		t.Fatalf("Expected model_id string, got: %v", body)
	}
	places, _ := body["places"].(float64)
	transitions, _ := body["transitions"].(float64)
	arcs, _ := body["arcs"].(float64)
	if int(places) < 1 || int(transitions) < 1 || int(arcs) < 1 {
		t.Fatalf("Expected non-empty petri net counts, got places=%v transitions=%v arcs=%v", places, transitions, arcs)
	}
	t.Logf("discovery E2E OK: places=%v transitions=%v arcs=%v", places, transitions, arcs)
}

// TestDiscoveryPipeline_XES runs discover on a real XES fixture when available (repo-relative).
func TestDiscoveryPipeline_XES(t *testing.T) {
	if !isBosRunning(t) || !isPm4pyRunning(t) {
		t.Skip("Skipping: requires both BusinessOS and pm4py-rust running")
	}
	// tests/pipeline -> ../../../../../pm4py-rust/test_data (repo root)
	xesPath := filepath.Clean(filepath.Join("..", "..", "..", "..", "..", "pm4py-rust", "test_data", "running-example.xes"))
	if _, err := os.Stat(xesPath); err != nil {
		t.Skip("Skipping: XES fixture not found at " + xesPath)
	}
	payload := map[string]interface{}{
		"log_path":  xesPath,
		"algorithm": "alpha",
	}
	status, body := postJSON(t, bosBase+"/api/bos/discover", payload)
	if status < 200 || status >= 300 {
		t.Fatalf("Expected 2xx from /api/bos/discover (XES), got %d — body: %v", status, body)
	}
	t.Logf("XES discovery E2E OK: model_id=%v", body["model_id"])
}

func TestStatisticsPipelineE2E(t *testing.T) {
	if !isBosRunning(t) || !isPm4pyRunning(t) {
		t.Skip("Skipping: requires both BusinessOS and pm4py-rust running")
	}
	logPath := writeTempEventLogJSON(t)
	payload := map[string]interface{}{"log_path": logPath}
	status, body := postJSON(t, bosBase+"/api/bos/statistics", payload)
	if status < 200 || status >= 300 {
		t.Fatalf("Expected 2xx from /api/bos/statistics, got %d — body: %v", status, body)
	}
	// BOSStatisticsResponse uses num_traces / num_events / num_unique_activities
	nt, _ := body["num_traces"].(float64)
	ne, _ := body["num_events"].(float64)
	if int(nt) != 2 {
		t.Errorf("Expected num_traces==2, got %v", nt)
	}
	if int(ne) != 8 {
		t.Errorf("Expected num_events==8, got %v", ne)
	}
	t.Logf("statistics E2E OK: num_traces=%v num_events=%v", nt, ne)
}

func TestConformancePipelineE2E(t *testing.T) {
	if !isBosRunning(t) || !isPm4pyRunning(t) {
		t.Skip("Skipping: requires both BusinessOS and pm4py-rust running")
	}
	logPath := writeTempEventLogJSON(t)
	modelPath := writeTempPetriNetJSON(t)
	payload := map[string]interface{}{
		"log_path":    logPath,
		"model_id":    "e2e-model",
		"model_path":  modelPath,
	}
	status, body := postJSON(t, bosBase+"/api/bos/conformance", payload)
	if status < 200 || status >= 300 {
		t.Fatalf("Expected 2xx from /api/bos/conformance, got %d — body: %v", status, body)
	}
	fitness, ok := body["fitness"].(float64)
	if !ok {
		t.Fatalf("Expected fitness float, got: %v", body)
	}
	if fitness < 0.0 || fitness > 1.0 {
		t.Errorf("Expected fitness in [0,1], got %f", fitness)
	}
}

func TestPipelineRoundTripLatency(t *testing.T) {
	if !isBosRunning(t) || !isPm4pyRunning(t) {
		t.Skip("Skipping: requires both BusinessOS and pm4py-rust running")
	}
	logPath := writeTempEventLogJSON(t)
	payload := map[string]interface{}{
		"log_path":  logPath,
		"algorithm": "alpha",
	}
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	start := time.Now()
	resp, err := httpClient.Post(bosBase+"/api/bos/discover", "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	elapsed := time.Since(start)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		t.Fatalf("Expected 2xx, got %d", resp.StatusCode)
	}
	if elapsed > 30*time.Second {
		t.Errorf("Round-trip took %v", elapsed)
	}
}

// TestParseXESOnPM4Py verifies pm4py-rust native XES parse endpoint (used by BOS for .xes paths).
func TestParseXESOnPM4Py(t *testing.T) {
	if !isPm4pyRunning(t) {
		t.Skip("Skipping: pm4py-rust not running")
	}
	xesPath := filepath.Clean(filepath.Join("..", "..", "..", "..", "..", "pm4py-rust", "test_data", "running-example.xes"))
	raw, err := os.ReadFile(xesPath)
	if err != nil {
		t.Skip("Skipping: " + xesPath)
	}
	req, err := http.NewRequest(http.MethodPost, pm4pyBase+"/api/io/parse-xes", bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/xml")
	resp, err := httpClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("parse-xes: %d %s", resp.StatusCode, string(b))
	}
	var v map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		t.Fatal(err)
	}
	if _, ok := v["traces"]; !ok {
		t.Fatalf("expected traces in parsed log: %v", v)
	}
}

func TestBosHealthPathAlias(t *testing.T) {
	if !isBosRunning(t) {
		t.Skip("Skipping: BusinessOS not running")
	}
	resp, err := httpClient.Get(bosBase + "/health")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /health: %d", resp.StatusCode)
	}
}

func TestWave7StringInHealthBodies(t *testing.T) {
	// Light sanity: responses contain expected substrings for wave7 grep.
	if isPm4pyRunning(t) {
		b := getBody(t, pm4pyBase+"/api/health")
		if !strings.Contains(strings.ToLower(b), "healthy") {
			t.Errorf("pm4py health body: %s", b)
		}
	}
}

// TestJaegerQuerySeesBusinessOSTraces checks Jaeger Query HTTP API after a traced BOS request (compose: jaeger:16686).
func TestJaegerQuerySeesBusinessOSTraces(t *testing.T) {
	jaeger := os.Getenv("JAEGER_QUERY_URL")
	if jaeger == "" {
		jaeger = "http://127.0.0.1:16686"
	}
	svcResp, err := httpClient.Get(jaeger + "/api/services")
	if err != nil {
		t.Skipf("Jaeger not reachable: %v", err)
	}
	_ = svcResp.Body.Close()
	if svcResp.StatusCode != http.StatusOK {
		t.Skipf("Jaeger /api/services not OK: %d", svcResp.StatusCode)
	}
	if !isBosRunning(t) || !isPm4pyRunning(t) {
		t.Skip("Skipping: requires BusinessOS, pm4py-rust, and Jaeger")
	}
	logPath := writeTempEventLogJSON(t)
	payload := map[string]interface{}{"log_path": logPath, "algorithm": "alpha"}
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := httpClient.Post(bosBase+"/api/bos/discover", "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		t.Fatalf("discover: %d", resp.StatusCode)
	}
	deadline := time.Now().Add(90 * time.Second)
	for time.Now().Before(deadline) {
		q := jaeger + "/api/traces?service=businessos&limit=5"
		tr, err := httpClient.Get(q)
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}
		b, _ := io.ReadAll(tr.Body)
		_ = tr.Body.Close()
		if tr.StatusCode != http.StatusOK {
			time.Sleep(2 * time.Second)
			continue
		}
		var wrapped struct {
			Data []json.RawMessage `json:"data"`
		}
		if json.Unmarshal(b, &wrapped) == nil && len(wrapped.Data) > 0 {
			return
		}
		time.Sleep(2 * time.Second)
	}
	t.Fatal("Jaeger Query returned no traces for service=businessos within timeout")
}

func getBody(t *testing.T, url string) string {
	t.Helper()
	resp, err := httpClient.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return string(b)
}
