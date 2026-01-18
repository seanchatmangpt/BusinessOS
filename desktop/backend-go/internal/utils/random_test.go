package utils

import (
	"encoding/base64"
	"encoding/hex"
	"testing"
)

func TestGenerateRandomBytes(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{"8 bytes", 8},
		{"16 bytes", 16},
		{"32 bytes", 32},
		{"64 bytes", 64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bytes, err := GenerateRandomBytes(tt.length)
			if err != nil {
				t.Fatalf("GenerateRandomBytes failed: %v", err)
			}
			if len(bytes) != tt.length {
				t.Errorf("Expected %d bytes, got %d", tt.length, len(bytes))
			}

			// Verify randomness: generate another and ensure they're different
			bytes2, err := GenerateRandomBytes(tt.length)
			if err != nil {
				t.Fatalf("GenerateRandomBytes failed on second call: %v", err)
			}

			if string(bytes) == string(bytes2) {
				t.Error("Two consecutive calls produced identical output (highly unlikely)")
			}
		})
	}
}

func TestGenerateRandomHex(t *testing.T) {
	tests := []struct {
		name           string
		byteLength     int
		expectedLength int // hex output is 2x byte length
	}{
		{"8 bytes → 16 hex chars", 8, 16},
		{"16 bytes → 32 hex chars", 16, 32},
		{"32 bytes → 64 hex chars", 32, 64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hexStr, err := GenerateRandomHex(tt.byteLength)
			if err != nil {
				t.Fatalf("GenerateRandomHex failed: %v", err)
			}
			if len(hexStr) != tt.expectedLength {
				t.Errorf("Expected %d chars, got %d", tt.expectedLength, len(hexStr))
			}

			// Verify it's valid hex
			_, err = hex.DecodeString(hexStr)
			if err != nil {
				t.Errorf("Output is not valid hex: %v", err)
			}

			// Verify uniqueness
			hexStr2, err := GenerateRandomHex(tt.byteLength)
			if err != nil {
				t.Fatalf("GenerateRandomHex failed on second call: %v", err)
			}
			if hexStr == hexStr2 {
				t.Error("Two consecutive calls produced identical output")
			}
		})
	}
}

func TestGenerateRandomBase64(t *testing.T) {
	tests := []struct {
		name       string
		byteLength int
	}{
		{"16 bytes", 16},
		{"32 bytes", 32},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b64Str, err := GenerateRandomBase64(tt.byteLength)
			if err != nil {
				t.Fatalf("GenerateRandomBase64 failed: %v", err)
			}

			// Verify it's valid base64
			_, err = base64.URLEncoding.DecodeString(b64Str)
			if err != nil {
				t.Errorf("Output is not valid base64: %v", err)
			}

			// Verify uniqueness
			b64Str2, err := GenerateRandomBase64(tt.byteLength)
			if err != nil {
				t.Fatalf("GenerateRandomBase64 failed on second call: %v", err)
			}
			if b64Str == b64Str2 {
				t.Error("Two consecutive calls produced identical output")
			}
		})
	}
}

func TestGenerateSessionToken(t *testing.T) {
	token, err := GenerateSessionToken()
	if err != nil {
		t.Fatalf("GenerateSessionToken failed: %v", err)
	}

	// 32 bytes base64-encoded should be 44 characters
	if len(token) != 44 {
		t.Errorf("Expected 44 chars (32 bytes base64), got %d", len(token))
	}

	// Verify it's valid base64
	decoded, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		t.Errorf("Token is not valid base64: %v", err)
	}
	if len(decoded) != 32 {
		t.Errorf("Decoded token should be 32 bytes, got %d", len(decoded))
	}

	// Verify uniqueness
	token2, err := GenerateSessionToken()
	if err != nil {
		t.Fatalf("GenerateSessionToken failed on second call: %v", err)
	}
	if token == token2 {
		t.Error("Two consecutive session tokens are identical")
	}
}

func TestGenerateUserID(t *testing.T) {
	userID, err := GenerateUserID()
	if err != nil {
		t.Fatalf("GenerateUserID failed: %v", err)
	}

	// Should be exactly 22 characters (16 bytes base64 truncated)
	if len(userID) != 22 {
		t.Errorf("Expected 22 chars, got %d", len(userID))
	}

	// Verify uniqueness
	userID2, err := GenerateUserID()
	if err != nil {
		t.Fatalf("GenerateUserID failed on second call: %v", err)
	}
	if userID == userID2 {
		t.Error("Two consecutive user IDs are identical")
	}
}

func TestGenerateSessionID(t *testing.T) {
	sessionID, err := GenerateSessionID()
	if err != nil {
		t.Fatalf("GenerateSessionID failed: %v", err)
	}

	// Should be exactly 22 characters (16 bytes base64 truncated)
	if len(sessionID) != 22 {
		t.Errorf("Expected 22 chars, got %d", len(sessionID))
	}

	// Verify uniqueness
	sessionID2, err := GenerateSessionID()
	if err != nil {
		t.Fatalf("GenerateSessionID failed on second call: %v", err)
	}
	if sessionID == sessionID2 {
		t.Error("Two consecutive session IDs are identical")
	}
}

func TestGenerateOAuthState(t *testing.T) {
	state, err := GenerateOAuthState()
	if err != nil {
		t.Fatalf("GenerateOAuthState failed: %v", err)
	}

	// 32 bytes base64-encoded should be 44 characters
	if len(state) != 44 {
		t.Errorf("Expected 44 chars (32 bytes base64), got %d", len(state))
	}

	// Verify it's valid base64
	_, err = base64.URLEncoding.DecodeString(state)
	if err != nil {
		t.Errorf("State is not valid base64: %v", err)
	}

	// Verify uniqueness
	state2, err := GenerateOAuthState()
	if err != nil {
		t.Fatalf("GenerateOAuthState failed on second call: %v", err)
	}
	if state == state2 {
		t.Error("Two consecutive OAuth states are identical")
	}
}

func TestGenerateShareID(t *testing.T) {
	shareID, err := GenerateShareID()
	if err != nil {
		t.Fatalf("GenerateShareID failed: %v", err)
	}

	// 8 bytes hex = 16 characters
	if len(shareID) != 16 {
		t.Errorf("Expected 16 chars (8 bytes hex), got %d", len(shareID))
	}

	// Verify it's valid hex
	_, err = hex.DecodeString(shareID)
	if err != nil {
		t.Errorf("ShareID is not valid hex: %v", err)
	}

	// Verify uniqueness
	shareID2, err := GenerateShareID()
	if err != nil {
		t.Fatalf("GenerateShareID failed on second call: %v", err)
	}
	if shareID == shareID2 {
		t.Error("Two consecutive share IDs are identical")
	}
}

func TestGenerateShareToken(t *testing.T) {
	token, err := GenerateShareToken()
	if err != nil {
		t.Fatalf("GenerateShareToken failed: %v", err)
	}

	// 16 bytes hex = 32 characters
	if len(token) != 32 {
		t.Errorf("Expected 32 chars (16 bytes hex), got %d", len(token))
	}

	// Verify it's valid hex
	_, err = hex.DecodeString(token)
	if err != nil {
		t.Errorf("ShareToken is not valid hex: %v", err)
	}

	// Verify uniqueness
	token2, err := GenerateShareToken()
	if err != nil {
		t.Fatalf("GenerateShareToken failed on second call: %v", err)
	}
	if token == token2 {
		t.Error("Two consecutive share tokens are identical")
	}
}

func TestGenerateNonce(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{"12 bytes (common for GCM)", 12},
		{"16 bytes", 16},
		{"24 bytes", 24},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nonce, err := GenerateNonce(tt.length)
			if err != nil {
				t.Fatalf("GenerateNonce failed: %v", err)
			}
			if len(nonce) != tt.length {
				t.Errorf("Expected %d bytes, got %d", tt.length, len(nonce))
			}

			// Verify uniqueness
			nonce2, err := GenerateNonce(tt.length)
			if err != nil {
				t.Fatalf("GenerateNonce failed on second call: %v", err)
			}
			if string(nonce) == string(nonce2) {
				t.Error("Two consecutive nonces are identical")
			}
		})
	}
}

func TestMustGenerateRandomHex(t *testing.T) {
	// Should not panic under normal circumstances
	hexStr := MustGenerateRandomHex(16)
	if len(hexStr) != 32 {
		t.Errorf("Expected 32 chars, got %d", len(hexStr))
	}

	// Verify it's valid hex
	_, err := hex.DecodeString(hexStr)
	if err != nil {
		t.Errorf("Output is not valid hex: %v", err)
	}
}

func TestMustGenerateSessionToken(t *testing.T) {
	// Should not panic under normal circumstances
	token := MustGenerateSessionToken()
	if len(token) != 44 {
		t.Errorf("Expected 44 chars, got %d", len(token))
	}

	// Verify it's valid base64
	_, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		t.Errorf("Token is not valid base64: %v", err)
	}
}

// Benchmark tests
func BenchmarkGenerateRandomBytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = GenerateRandomBytes(32)
	}
}

func BenchmarkGenerateSessionToken(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = GenerateSessionToken()
	}
}

func BenchmarkGenerateUserID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = GenerateUserID()
	}
}

func BenchmarkGenerateOAuthState(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = GenerateOAuthState()
	}
}
