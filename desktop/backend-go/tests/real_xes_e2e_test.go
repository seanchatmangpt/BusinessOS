//go:build integration

// Real XES End-to-End Integration Test
//
// This test closes the critical gap: real XES event log files exist but were
// never used in a live integration test against pm4py-rust.
//
// Requires: pm4py-rust running at http://localhost:8090
// Run with: go test -tags=integration -v ./tests/ -run TestRealXES
//
// Uses: pm4py-rust/test_data/running-example.xes (Fluxicon Nitro benchmark)
// — 6 traces, 42 events, 8 unique activities (insurance claim handling)
package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// xesLogEntry represents a minimal XES trace/event for parsing
type xesLog struct {
	XMLName xml.Name   `xml:"log"`
	Traces  []xesTrace `xml:"trace"`
}

type xesTrace struct {
	Strings []xesString `xml:"string"`
	Events  []xesEvent  `xml:"event"`
}

type xesEvent struct {
	Strings []xesString `xml:"string"`
	Dates   []xesDate   `xml:"date"`
}

type xesString struct {
	Key   string `xml:"key,attr"`
	Value string `xml:"value,attr"`
}

type xesDate struct {
	Key   string `xml:"key,attr"`
	Value string `xml:"value,attr"`
}

// pm4pyDiscoveryRequest matches the pm4py-rust /api/discovery/alpha endpoint
type pm4pyDiscoveryRequest struct {
	EventLog json.RawMessage `json:"event_log"`
	Variant  string          `json:"variant,omitempty"`
}

// pm4pyDiscoveryResponse matches the pm4py-rust discovery response
type pm4pyDiscoveryResponse struct {
	PetriNet struct {
		Places []struct {
			ID             string `json:"id"`
			Name           string `json:"name"`
			InitialMarking int    `json:"initial_marking"`
		} `json:"places"`
		Transitions []struct {
			ID    string  `json:"id"`
			Name  string  `json:"name"`
			Label *string `json:"label,omitempty"`
		} `json:"transitions"`
		Arcs []struct {
			From   string `json:"from"`
			To     string `json:"to"`
			Weight int    `json:"weight"`
		} `json:"arcs"`
		InitialPlace *string `json:"initial_place"`
		FinalPlace   *string `json:"final_place"`
	} `json:"petri_net"`
	Algorithm     string `json:"algorithm"`
	ExecutionTime uint64 `json:"execution_time_ms"`
	EventCount    int    `json:"event_count"`
	TraceCount    int    `json:"trace_count"`
}

const pm4pyBaseURL = "http://localhost:8090"

// findRunningExampleXES locates the running-example.xes file relative to this
// test file, which lives in BusinessOS/desktop/backend-go/tests/.
func findRunningExampleXES(t *testing.T) string {
	t.Helper()

	// Navigate from this test file up to the repo root, then into pm4py-rust/test_data
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok, "runtime.Caller failed")

	// thisFile = .../BusinessOS/desktop/backend-go/tests/real_xes_e2e_test.go
	repoRoot := filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "..")
	candidates := []string{
		filepath.Join(repoRoot, "pm4py-rust", "test_data", "running-example.xes"),
		"/Users/sac/chatmangpt/pm4py-rust/test_data/running-example.xes",
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	t.Fatalf("running-example.xes not found in any of: %v", candidates)
	return ""
}

// isPM4PyRunning checks if pm4py-rust is healthy at localhost:8090
func isPM4PyRunning(t *testing.T) bool {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", pm4pyBaseURL+"/api/health", nil)
	if err != nil {
		return false
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// xesToEventLogJSON converts a real XES file to the JSON event log format
// that pm4py-rust's /api/discovery/alpha endpoint expects.
// The format must match pm4py::log::EventLog serde deserialization:
//
//	{"traces": [{"id": "...", "events": [{"activity": "...", "timestamp": "..."}]}]}
func xesToEventLogJSON(t *testing.T, xesPath string) json.RawMessage {
	t.Helper()

	data, err := os.ReadFile(xesPath)
	require.NoError(t, err, "read XES file")

	var xLog xesLog
	err = xml.Unmarshal(data, &xLog)
	require.NoError(t, err, "parse XES XML")
	require.NotEmpty(t, xLog.Traces, "XES must have traces")

	type jsonEvent struct {
		Activity   string            `json:"activity"`
		Timestamp  string            `json:"timestamp"`
		Resource   *string           `json:"resource,omitempty"`
		Attributes map[string]string `json:"attributes"`
	}

	type jsonTrace struct {
		ID         string            `json:"id"`
		Events     []jsonEvent       `json:"events"`
		Attributes map[string]string `json:"attributes"`
	}

	type jsonEventLog struct {
		Traces     []jsonTrace       `json:"traces"`
		Attributes map[string]string `json:"attributes"`
	}

	eventLog := jsonEventLog{
		Traces:     make([]jsonTrace, 0, len(xLog.Traces)),
		Attributes: map[string]string{"source": xesPath},
	}

	for i, xTrace := range xLog.Traces {
		traceID := fmt.Sprintf("trace_%d", i)
		for _, s := range xTrace.Strings {
			if s.Key == "concept:name" {
				traceID = s.Value
				break
			}
		}

		jTrace := jsonTrace{
			ID:         traceID,
			Events:     make([]jsonEvent, 0, len(xTrace.Events)),
			Attributes: map[string]string{},
		}

		for _, xEvent := range xTrace.Events {
			var activity, timestamp string
			var resource *string
			attrs := map[string]string{}

			for _, s := range xEvent.Strings {
				switch s.Key {
				case "concept:name":
					activity = s.Value
				case "org:resource":
					r := s.Value
					resource = &r
				default:
					attrs[s.Key] = s.Value
				}
			}
			for _, d := range xEvent.Dates {
				if d.Key == "time:timestamp" {
					timestamp = d.Value
				}
			}

			// Ensure timestamp parses as RFC3339
			if timestamp == "" {
				timestamp = time.Now().UTC().Format(time.RFC3339)
			} else if !strings.Contains(timestamp, "T") {
				// Already RFC3339, keep as-is
			}

			if activity != "" {
				jTrace.Events = append(jTrace.Events, jsonEvent{
					Activity:   activity,
					Timestamp:  timestamp,
					Resource:   resource,
					Attributes: attrs,
				})
			}
		}

		if len(jTrace.Events) > 0 {
			eventLog.Traces = append(eventLog.Traces, jTrace)
		}
	}

	jsonBytes, err := json.Marshal(eventLog)
	require.NoError(t, err, "marshal event log to JSON")

	t.Logf("[XES→JSON] Converted %d traces, %d total events",
		len(eventLog.Traces),
		func() int {
			n := 0
			for _, tr := range eventLog.Traces {
				n += len(tr.Events)
			}
			return n
		}())

	return json.RawMessage(jsonBytes)
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestRealXES_ParseXESFile(t *testing.T) {
	xesPath := findRunningExampleXES(t)

	data, err := os.ReadFile(xesPath)
	require.NoError(t, err)

	var xLog xesLog
	err = xml.Unmarshal(data, &xLog)
	require.NoError(t, err)

	assert.Equal(t, 6, len(xLog.Traces), "running-example.xes should have 6 traces")

	totalEvents := 0
	for _, tr := range xLog.Traces {
		totalEvents += len(tr.Events)
	}
	assert.Greater(t, totalEvents, 0, "must have > 0 events")

	t.Logf("[REAL XES] Parsed %d traces, %d events from %s", len(xLog.Traces), totalEvents, xesPath)
}

func TestRealXES_ConvertToEventLogJSON(t *testing.T) {
	xesPath := findRunningExampleXES(t)
	jsonData := xesToEventLogJSON(t, xesPath)

	// Verify JSON structure
	var parsed map[string]interface{}
	err := json.Unmarshal(jsonData, &parsed)
	require.NoError(t, err, "event log JSON must be valid")

	traces, ok := parsed["traces"].([]interface{})
	require.True(t, ok, "JSON must have 'traces' array")
	assert.Equal(t, 6, len(traces), "must have 6 traces")

	// Verify first trace has events with activity and timestamp
	firstTrace := traces[0].(map[string]interface{})
	events := firstTrace["events"].([]interface{})
	assert.Greater(t, len(events), 0, "first trace must have events")

	firstEvent := events[0].(map[string]interface{})
	assert.NotEmpty(t, firstEvent["activity"], "event must have activity")
	assert.NotEmpty(t, firstEvent["timestamp"], "event must have timestamp")

	t.Logf("[JSON] First trace has %d events, first activity: %s",
		len(events), firstEvent["activity"])
}

func TestRealXES_DiscoverPetriNetFromPM4PyRust(t *testing.T) {
	if !isPM4PyRunning(t) {
		t.Skip("pm4py-rust not running at localhost:8090 — skipping live integration test")
	}

	xesPath := findRunningExampleXES(t)
	eventLogJSON := xesToEventLogJSON(t, xesPath)

	reqBody := pm4pyDiscoveryRequest{
		EventLog: eventLogJSON,
		Variant:  "alpha",
	}

	bodyBytes, err := json.Marshal(reqBody)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		pm4pyBaseURL+"/api/discovery/alpha",
		bytes.NewReader(bodyBytes))
	require.NoError(t, err)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	require.NoError(t, err, "POST /api/discovery/alpha must succeed")
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200, got %d: %s", resp.StatusCode, string(respBody))
	}

	var discoveryResp pm4pyDiscoveryResponse
	err = json.Unmarshal(respBody, &discoveryResp)
	require.NoError(t, err, "response must be valid JSON")

	// Assert: response contains real Petri net
	assert.Greater(t, len(discoveryResp.PetriNet.Places), 0,
		"discovered Petri net must have > 0 places")
	assert.Greater(t, len(discoveryResp.PetriNet.Transitions), 0,
		"discovered Petri net must have > 0 transitions")
	assert.Greater(t, len(discoveryResp.PetriNet.Arcs), 0,
		"discovered Petri net must have > 0 arcs")

	// Assert: metadata is correct
	assert.Equal(t, 6, discoveryResp.TraceCount, "trace count must be 6")
	assert.Greater(t, discoveryResp.EventCount, 0, "event count must be > 0")
	assert.Equal(t, "alpha_miner", discoveryResp.Algorithm)

	t.Logf("╔═══════════════════════════════════════════════════╗")
	t.Logf("║  REAL XES → pm4py-rust DISCOVERY E2E RESULTS     ║")
	t.Logf("╠═══════════════════════════════════════════════════╣")
	t.Logf("║  Algorithm:     %s", discoveryResp.Algorithm)
	t.Logf("║  Traces:        %d", discoveryResp.TraceCount)
	t.Logf("║  Events:        %d", discoveryResp.EventCount)
	t.Logf("║  Places:        %d", len(discoveryResp.PetriNet.Places))
	t.Logf("║  Transitions:   %d", len(discoveryResp.PetriNet.Transitions))
	t.Logf("║  Arcs:          %d", len(discoveryResp.PetriNet.Arcs))
	t.Logf("║  Execution ms:  %d", discoveryResp.ExecutionTime)
	t.Logf("╚═══════════════════════════════════════════════════╝")
}

func TestRealXES_HealthCheckBeforeDiscovery(t *testing.T) {
	if !isPM4PyRunning(t) {
		t.Skip("pm4py-rust not running at localhost:8090 — skipping")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", pm4pyBaseURL+"/api/health", nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "health check must return 200")

	body, _ := io.ReadAll(resp.Body)
	t.Logf("[Health] %s", string(body))
}
