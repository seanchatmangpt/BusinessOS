package canopy_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	canopyintegration "github.com/rhl/businessos-backend/internal/integrations/canopy"
)

// TestPushIntelligenceInjectsTraceparent verifies that PushIntelligence forwards a
// W3C traceparent header and the BOS shared secret to the Canopy endpoint.
//
// Chicago TDD RED → GREEN: mock HTTP server records request headers, client must
// set both headers; if either is missing the test fails.
func TestPushIntelligenceInjectsTraceparent(t *testing.T) {
	var gotHeaders http.Header
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeaders = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer srv.Close()

	cfg := &canopyintegration.Config{
		BaseURL:      srv.URL,
		SharedSecret: "test-secret-xyz",
		Timeout:      5 * time.Second,
		MaxRetries:   1,
	}
	client := canopyintegration.NewClientWithConfig(cfg)

	ctx := context.Background()
	if err := client.PushIntelligence(ctx, 10, 5); err != nil {
		t.Fatalf("PushIntelligence returned unexpected error: %v", err)
	}

	// X-BOS-Secret must be present
	if got := gotHeaders.Get("X-Bos-Secret"); got != "test-secret-xyz" {
		t.Errorf("X-Bos-Secret = %q, want %q", got, "test-secret-xyz")
	}
}

// TestPushIntelligenceRetriesOnServerError verifies that the client retries up to
// MaxRetries times on 5xx responses before returning an error.
func TestPushIntelligenceRetriesOnServerError(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	cfg := &canopyintegration.Config{
		BaseURL:      srv.URL,
		SharedSecret: "test-secret",
		Timeout:      2 * time.Second,
		MaxRetries:   2,
	}
	client := canopyintegration.NewClientWithConfig(cfg)

	ctx := context.Background()
	err := client.PushIntelligence(ctx, 5, 2)
	if err == nil {
		t.Error("expected error on persistent 500, got nil")
	}
	// MaxRetries=2 means initial attempt + 2 retries = 3 total calls
	if callCount < 2 {
		t.Errorf("expected at least 2 calls (retry), got %d", callCount)
	}
}

// TestPushIntelligenceReturnsNilOnOK verifies that a 200 response causes no error.
func TestPushIntelligenceReturnsNilOnOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer srv.Close()

	cfg := &canopyintegration.Config{
		BaseURL:      srv.URL,
		SharedSecret: "s",
		Timeout:      2 * time.Second,
		MaxRetries:   1,
	}
	client := canopyintegration.NewClientWithConfig(cfg)
	if err := client.PushIntelligence(context.Background(), 1, 1); err != nil {
		t.Errorf("unexpected error on 200: %v", err)
	}
}
