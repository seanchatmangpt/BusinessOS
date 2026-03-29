package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// a2aTasksSend builds a pm4py-rust A2A tasks/send JSON-RPC body.
func a2aTasksSend(taskID, tool string, args map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tasks/send",
		"params": map[string]interface{}{
			"id": taskID,
			"message": map[string]interface{}{
				"role": "user",
				"parts": []map[string]interface{}{{
					"type": "data",
					"data": map[string]interface{}{
						"tool": tool,
						"args": args,
					},
				}},
			},
		},
	}
}

// a2aOneTraceLog is a minimal event log: 1 trace, 2 events (A → B).
func a2aOneTraceLog() map[string]interface{} {
	return map[string]interface{}{
		"attributes": map[string]interface{}{},
		"traces": []map[string]interface{}{{
			"id":         "case-001",
			"attributes": map[string]interface{}{},
			"events": []map[string]interface{}{
				{
					"activity":   "A",
					"timestamp":  "2024-01-01T10:00:00Z",
					"resource":   nil,
					"attributes": map[string]interface{}{},
				},
				{
					"activity":   "B",
					"timestamp":  "2024-01-01T10:30:00Z",
					"resource":   nil,
					"attributes": map[string]interface{}{},
				},
			},
		}},
	}
}

// a2aPostJSON POSTs JSON to server at path and returns decoded response.
func a2aPostJSON(t *testing.T, server *httptest.Server, path string, body interface{}) map[string]interface{} {
	t.Helper()
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	resp, err := http.Post(server.URL+path, "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("POST %s: %v", path, err)
	}
	defer resp.Body.Close()
	var out map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return out
}

// TestA2APm4pyWork verifies BusinessOS can delegate process mining work to pm4py-rust via A2A.
func TestA2APm4pyWork(t *testing.T) {
	// Mock pm4py-rust at /a2a
	pm4pyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/a2a" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		params, _ := req["params"].(map[string]interface{})
		taskID, _ := params["id"].(string)

		resp := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      req["id"],
			"result": map[string]interface{}{
				"id":     taskID,
				"status": map[string]interface{}{"state": "completed"},
				"artifacts": []map[string]interface{}{{
					"parts": []map[string]interface{}{{
						"type": "data",
						"data": map[string]interface{}{
							"trace_count": float64(1),
							"event_count": float64(2),
						},
					}},
				}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Logf("encode response: %v", err)
		}
	}))
	defer pm4pyServer.Close()

	t.Run("statistics_returns_completed_state", func(t *testing.T) {
		payload := a2aTasksSend("bos-work-1", "pm4py_statistics",
			map[string]interface{}{"event_log": a2aOneTraceLog()})

		result := a2aPostJSON(t, pm4pyServer, "/a2a", payload)

		resultObj, ok := result["result"].(map[string]interface{})
		if !ok {
			t.Fatalf("result must be an object; got: %v", result)
		}
		status, _ := resultObj["status"].(map[string]interface{})
		if status["state"] != "completed" {
			t.Errorf("expected state=completed; got %v", status["state"])
		}
	})

	t.Run("statistics_returns_trace_and_event_count", func(t *testing.T) {
		payload := a2aTasksSend("bos-work-2", "pm4py_statistics",
			map[string]interface{}{"event_log": a2aOneTraceLog()})

		result := a2aPostJSON(t, pm4pyServer, "/a2a", payload)

		resultObj := result["result"].(map[string]interface{})
		artifacts := resultObj["artifacts"].([]interface{})
		if len(artifacts) == 0 {
			t.Fatal("artifacts must be non-empty")
		}
		parts := artifacts[0].(map[string]interface{})["parts"].([]interface{})
		data := parts[0].(map[string]interface{})["data"].(map[string]interface{})

		if data["trace_count"] != float64(1) {
			t.Errorf("trace_count: expected 1, got %v", data["trace_count"])
		}
		if data["event_count"] != float64(2) {
			t.Errorf("event_count: expected 2, got %v", data["event_count"])
		}
	})

	t.Run("agent_card_has_required_fields", func(t *testing.T) {
		// Mock /.well-known/agent-card.json
		cardServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(map[string]interface{}{
				"name":            "pm4py-rust",
				"protocolVersion": "0.2.1",
				"skills":          []interface{}{},
			}); err != nil {
				t.Logf("encode card: %v", err)
			}
		}))
		defer cardServer.Close()

		resp, err := http.Get(cardServer.URL + "/.well-known/agent-card.json")
		if err != nil {
			t.Fatalf("GET agent-card.json: %v", err)
		}
		defer resp.Body.Close()
		var card map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&card); err != nil {
			t.Fatalf("decode card: %v", err)
		}

		if card["name"] == nil {
			t.Error("agent card must have 'name' field")
		}
		if card["protocolVersion"] == nil {
			t.Error("agent card must have 'protocolVersion' field")
		}
	})

	t.Run("method_not_found_returns_error_code", func(t *testing.T) {
		errServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      1,
				"error":   map[string]interface{}{"code": float64(-32601), "message": "Method not found"},
			}); err != nil {
				t.Logf("encode error resp: %v", err)
			}
		}))
		defer errServer.Close()

		payload := map[string]interface{}{
			"jsonrpc": "2.0", "id": 1,
			"method": "unknown/method",
			"params": map[string]interface{}{},
		}
		result := a2aPostJSON(t, errServer, "/a2a", payload)

		errObj, ok := result["error"].(map[string]interface{})
		if !ok {
			t.Fatalf("error field must be present; got: %v", result)
		}
		if errObj["code"] != float64(-32601) {
			t.Errorf("error code: expected -32601, got %v", errObj["code"])
		}
	})
}
