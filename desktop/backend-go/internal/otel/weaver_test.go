package otel

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestIsLiveCheckEnabled(t *testing.T) {
	// Save and restore env
	orig := os.Getenv(EnvWeaverLiveCheck)
	defer os.Setenv(EnvWeaverLiveCheck, orig)

	tests := []struct {
		value string
		want  bool
	}{
		{"true", true},
		{"TRUE", false}, // case-sensitive
		{"", false},
		{"false", false},
		{"1", false},
	}

	for _, tc := range tests {
		os.Setenv(EnvWeaverLiveCheck, tc.value)
		got := IsLiveCheckEnabled()
		if got != tc.want {
			t.Errorf("IsLiveCheckEnabled(%q) = %v, want %v", tc.value, got, tc.want)
		}
	}
}

func TestSetupWeaverLiveCheckWithDefaultEndpoint(t *testing.T) {
	// Save and restore env
	origEndpoint := os.Getenv(EnvWeaverOTLPEndpoint)
	origLiveCheck := os.Getenv(EnvWeaverLiveCheck)
	defer os.Setenv(EnvWeaverOTLPEndpoint, origEndpoint)
	defer os.Setenv(EnvWeaverLiveCheck, origLiveCheck)

	os.Setenv(EnvWeaverOTLPEndpoint, "")
	os.Setenv(EnvWeaverLiveCheck, "true")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Note: this will fail to connect if no OTLP receiver is running,
	// but the TracerProvider is still created successfully.
	shutdown, err := SetupWeaverLiveCheck(ctx)
	if err != nil {
		// Setup itself should succeed even without a running receiver.
		t.Fatalf("SetupWeaverLiveCheck failed: %v", err)
	}
	if shutdown == nil {
		t.Fatal("SetupWeaverLiveCheck returned nil shutdown function")
	}

	// Verify shutdown does not panic
	err = shutdown(context.Background())
	// Shutdown may fail if no receiver was running, but should not panic.
	t.Logf("shutdown result (may fail if no receiver): %v", err)
}

func TestSetupWeaverLiveCheckWithCustomEndpoint(t *testing.T) {
	origEndpoint := os.Getenv(EnvWeaverOTLPEndpoint)
	defer os.Setenv(EnvWeaverOTLPEndpoint, origEndpoint)

	os.Setenv(EnvWeaverOTLPEndpoint, "custom-host:9999")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	shutdown, err := SetupWeaverLiveCheck(ctx)
	if err != nil {
		t.Fatalf("SetupWeaverLiveCheck with custom endpoint failed: %v", err)
	}
	if shutdown == nil {
		t.Fatal("expected non-nil shutdown function")
	}

	// Clean up
	_ = shutdown(context.Background())
}

func TestConstantsAreNonEmpty(t *testing.T) {
	if EnvWeaverLiveCheck == "" {
		t.Error("EnvWeaverLiveCheck constant is empty")
	}
	if EnvWeaverOTLPEndpoint == "" {
		t.Error("EnvWeaverOTLPEndpoint constant is empty")
	}
	if DefaultOTLPEndpoint == "" {
		t.Error("DefaultOTLPEndpoint constant is empty")
	}
	if defaultShutdownTimeout <= 0 {
		t.Errorf("defaultShutdownTimeout = %v, want positive duration", defaultShutdownTimeout)
	}
}
