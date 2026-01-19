package services

import (
	"context"
	"strings"
	"testing"
)

func TestGenerateRequestID(t *testing.T) {
	// Generate multiple request IDs
	requestID1 := GenerateRequestID()
	requestID2 := GenerateRequestID()

	// Verify format
	if !strings.HasPrefix(requestID1, "req_") {
		t.Errorf("Request ID should have 'req_' prefix, got: %s", requestID1)
	}

	// Verify uniqueness
	if requestID1 == requestID2 {
		t.Error("Request IDs should be unique")
	}

	// Verify length (req_ + UUID format)
	if len(requestID1) != 40 { // "req_" (4) + UUID (36)
		t.Errorf("Request ID should be 40 characters, got: %d", len(requestID1))
	}
}

func TestGenerateSessionID(t *testing.T) {
	// Generate multiple session IDs
	sessionID1 := GenerateSessionID()
	sessionID2 := GenerateSessionID()

	// Verify format
	if !strings.HasPrefix(sessionID1, "sess_") {
		t.Errorf("Session ID should have 'sess_' prefix, got: %s", sessionID1)
	}

	// Verify uniqueness
	if sessionID1 == sessionID2 {
		t.Error("Session IDs should be unique")
	}

	// Verify length (sess_ + UUID format)
	if len(sessionID1) != 41 { // "sess_" (5) + UUID (36)
		t.Errorf("Session ID should be 41 characters, got: %d", len(sessionID1))
	}
}

func TestAddRequestIDToContext(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name       string
		requestID  string
		expectAuto bool
	}{
		{
			name:       "with explicit request ID",
			requestID:  "req_test-123",
			expectAuto: false,
		},
		{
			name:       "with empty request ID (auto-generate)",
			requestID:  "",
			expectAuto: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			newCtx := AddRequestIDToContext(ctx, tc.requestID)

			// Extract request ID
			extractedID := GetRequestIDFromContext(newCtx)

			if tc.expectAuto {
				// Should auto-generate
				if extractedID == "" {
					t.Error("Request ID should be auto-generated when empty")
				}
				if !strings.HasPrefix(extractedID, "req_") {
					t.Errorf("Auto-generated request ID should have 'req_' prefix, got: %s", extractedID)
				}
			} else {
				// Should use provided ID
				if extractedID != tc.requestID {
					t.Errorf("Expected request ID %s, got %s", tc.requestID, extractedID)
				}
			}
		})
	}
}

func TestGetRequestIDFromContext(t *testing.T) {
	testCases := []struct {
		name       string
		setupCtx   func() context.Context
		expectedID string
	}{
		{
			name: "context with request ID",
			setupCtx: func() context.Context {
				ctx := context.Background()
				return AddRequestIDToContext(ctx, "req_test-456")
			},
			expectedID: "req_test-456",
		},
		{
			name: "context without request ID",
			setupCtx: func() context.Context {
				return context.Background()
			},
			expectedID: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := tc.setupCtx()
			requestID := GetRequestIDFromContext(ctx)

			if requestID != tc.expectedID {
				t.Errorf("Expected request ID '%s', got '%s'", tc.expectedID, requestID)
			}
		})
	}
}

func TestAddSessionIDToContext(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name       string
		sessionID  string
		expectAuto bool
	}{
		{
			name:       "with explicit session ID",
			sessionID:  "sess_test-session-123",
			expectAuto: false,
		},
		{
			name:       "with empty session ID (auto-generate)",
			sessionID:  "",
			expectAuto: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			newCtx := AddSessionIDToContext(ctx, tc.sessionID)

			// Extract session ID
			extractedID := GetSessionIDFromContext(newCtx)

			if tc.expectAuto {
				// Should auto-generate
				if extractedID == "" {
					t.Error("Session ID should be auto-generated when empty")
				}
				if !strings.HasPrefix(extractedID, "sess_") {
					t.Errorf("Auto-generated session ID should have 'sess_' prefix, got: %s", extractedID)
				}
			} else {
				// Should use provided ID
				if extractedID != tc.sessionID {
					t.Errorf("Expected session ID %s, got %s", tc.sessionID, extractedID)
				}
			}
		})
	}
}

func TestGetSessionIDFromContext(t *testing.T) {
	testCases := []struct {
		name       string
		setupCtx   func() context.Context
		expectedID string
	}{
		{
			name: "context with session ID",
			setupCtx: func() context.Context {
				ctx := context.Background()
				return AddSessionIDToContext(ctx, "sess_test-session-789")
			},
			expectedID: "sess_test-session-789",
		},
		{
			name: "context without session ID",
			setupCtx: func() context.Context {
				return context.Background()
			},
			expectedID: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := tc.setupCtx()
			sessionID := GetSessionIDFromContext(ctx)

			if sessionID != tc.expectedID {
				t.Errorf("Expected session ID '%s', got '%s'", tc.expectedID, sessionID)
			}
		})
	}
}

func TestGetOrGenerateRequestID(t *testing.T) {
	testCases := []struct {
		name       string
		setupCtx   func() context.Context
		expectAuto bool
	}{
		{
			name: "context without request ID",
			setupCtx: func() context.Context {
				return context.Background()
			},
			expectAuto: true,
		},
		{
			name: "context with existing request ID",
			setupCtx: func() context.Context {
				ctx := context.Background()
				return AddRequestIDToContext(ctx, "req_existing-123")
			},
			expectAuto: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := tc.setupCtx()
			newCtx, requestID := GetOrGenerateRequestID(ctx)

			if requestID == "" {
				t.Error("Request ID should never be empty")
			}

			// Verify context contains the request ID
			extractedID := GetRequestIDFromContext(newCtx)
			if extractedID != requestID {
				t.Errorf("Context should contain request ID %s, got %s", requestID, extractedID)
			}

			if tc.expectAuto {
				if !strings.HasPrefix(requestID, "req_") {
					t.Errorf("Auto-generated request ID should have 'req_' prefix, got: %s", requestID)
				}
			} else {
				if requestID != "req_existing-123" {
					t.Errorf("Expected existing request ID 'req_existing-123', got %s", requestID)
				}
			}
		})
	}
}

func TestGetOrGenerateSessionID(t *testing.T) {
	testCases := []struct {
		name       string
		setupCtx   func() context.Context
		expectAuto bool
	}{
		{
			name: "context without session ID",
			setupCtx: func() context.Context {
				return context.Background()
			},
			expectAuto: true,
		},
		{
			name: "context with existing session ID",
			setupCtx: func() context.Context {
				ctx := context.Background()
				return AddSessionIDToContext(ctx, "sess_existing-session-456")
			},
			expectAuto: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := tc.setupCtx()
			newCtx, sessionID := GetOrGenerateSessionID(ctx)

			if sessionID == "" {
				t.Error("Session ID should never be empty")
			}

			// Verify context contains the session ID
			extractedID := GetSessionIDFromContext(newCtx)
			if extractedID != sessionID {
				t.Errorf("Context should contain session ID %s, got %s", sessionID, extractedID)
			}

			if tc.expectAuto {
				if !strings.HasPrefix(sessionID, "sess_") {
					t.Errorf("Auto-generated session ID should have 'sess_' prefix, got: %s", sessionID)
				}
			} else {
				if sessionID != "sess_existing-session-456" {
					t.Errorf("Expected existing session ID 'sess_existing-session-456', got %s", sessionID)
				}
			}
		})
	}
}

func TestContextPropagation(t *testing.T) {
	// Test that IDs propagate through nested contexts
	ctx := context.Background()

	// Add both request and session IDs
	ctx = AddRequestIDToContext(ctx, "req_test-propagate")
	ctx = AddSessionIDToContext(ctx, "sess_test-propagate")

	// Create a derived context (simulating passing context through functions)
	derivedCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Verify IDs are preserved in derived context
	requestID := GetRequestIDFromContext(derivedCtx)
	sessionID := GetSessionIDFromContext(derivedCtx)

	if requestID != "req_test-propagate" {
		t.Errorf("Request ID should propagate through derived context, got: %s", requestID)
	}

	if sessionID != "sess_test-propagate" {
		t.Errorf("Session ID should propagate through derived context, got: %s", sessionID)
	}
}

func TestThreadSafety(t *testing.T) {
	// Test concurrent request ID generation (should be safe)
	const numGoroutines = 100
	results := make(chan string, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			results <- GenerateRequestID()
		}()
	}

	// Collect results and check uniqueness
	seen := make(map[string]bool)
	for i := 0; i < numGoroutines; i++ {
		id := <-results
		if seen[id] {
			t.Errorf("Duplicate request ID generated: %s", id)
		}
		seen[id] = true
	}
}
