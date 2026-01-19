package webhooks

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"
)

// =============================================================================
// SLACK SIGNATURE TESTS
// =============================================================================

func TestVerifySlackSignature(t *testing.T) {
	verifier := NewSignatureVerifier(map[string]string{
		"slack": "test-slack-secret",
	})

	body := []byte(`{"type":"url_verification","challenge":"test123"}`)
	timestamp := fmt.Sprintf("%d", time.Now().Unix())

	// Calculate valid signature
	baseString := fmt.Sprintf("v0:%s:%s", timestamp, string(body))
	mac := hmac.New(sha256.New, []byte("test-slack-secret"))
	mac.Write([]byte(baseString))
	validSignature := "v0=" + hex.EncodeToString(mac.Sum(nil))

	tests := []struct {
		name      string
		body      []byte
		timestamp string
		signature string
		want      bool
	}{
		{
			name:      "valid signature",
			body:      body,
			timestamp: timestamp,
			signature: validSignature,
			want:      true,
		},
		{
			name:      "invalid signature",
			body:      body,
			timestamp: timestamp,
			signature: "v0=invalid",
			want:      false,
		},
		{
			name:      "expired timestamp",
			body:      body,
			timestamp: fmt.Sprintf("%d", time.Now().Unix()-400), // 6+ minutes old
			signature: validSignature,
			want:      false,
		},
		{
			name:      "no secret configured",
			body:      body,
			timestamp: timestamp,
			signature: "anything",
			want:      true, // Should pass in development
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For "no secret" test, use empty verifier
			v := verifier
			if tt.name == "no secret configured" {
				v = NewSignatureVerifier(map[string]string{})
			}

			got := v.VerifySlackSignature(tt.body, tt.timestamp, tt.signature)
			if got != tt.want {
				t.Errorf("VerifySlackSignature() = %v, want %v", got, tt.want)
			}
		})
	}
}

// =============================================================================
// LINEAR SIGNATURE TESTS
// =============================================================================

func TestVerifyLinearSignature(t *testing.T) {
	verifier := NewSignatureVerifier(map[string]string{
		"linear": "test-linear-secret",
	})

	body := []byte(`{"action":"create","type":"Issue","data":{}}`)

	// Calculate valid signature
	mac := hmac.New(sha256.New, []byte("test-linear-secret"))
	mac.Write(body)
	validSignature := hex.EncodeToString(mac.Sum(nil))

	tests := []struct {
		name      string
		body      []byte
		signature string
		want      bool
	}{
		{
			name:      "valid signature",
			body:      body,
			signature: validSignature,
			want:      true,
		},
		{
			name:      "invalid signature",
			body:      body,
			signature: "invalid",
			want:      false,
		},
		{
			name:      "no secret configured",
			body:      body,
			signature: "anything",
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := verifier
			if tt.name == "no secret configured" {
				v = NewSignatureVerifier(map[string]string{})
			}

			got := v.VerifyLinearSignature(tt.body, tt.signature)
			if got != tt.want {
				t.Errorf("VerifyLinearSignature() = %v, want %v", got, tt.want)
			}
		})
	}
}

// =============================================================================
// HUBSPOT SIGNATURE TESTS
// =============================================================================

func TestVerifyHubSpotSignature(t *testing.T) {
	verifier := NewSignatureVerifier(map[string]string{
		"hubspot": "test-hubspot-secret",
	})

	body := []byte(`[{"objectId":123,"subscriptionType":"contact.creation"}]`)
	timestamp := fmt.Sprintf("%d", time.Now().UnixMilli())
	requestMethod := "POST"
	requestURI := "/webhooks/hubspot"

	// Calculate valid signature
	sourceString := requestMethod + requestURI + string(body) + timestamp
	mac := hmac.New(sha256.New, []byte("test-hubspot-secret"))
	mac.Write([]byte(sourceString))
	validSignature := hex.EncodeToString(mac.Sum(nil))

	tests := []struct {
		name          string
		body          []byte
		signature     string
		timestamp     string
		requestMethod string
		requestURI    string
		want          bool
	}{
		{
			name:          "valid signature",
			body:          body,
			signature:     validSignature,
			timestamp:     timestamp,
			requestMethod: requestMethod,
			requestURI:    requestURI,
			want:          true,
		},
		{
			name:          "invalid signature",
			body:          body,
			signature:     "invalid",
			timestamp:     timestamp,
			requestMethod: requestMethod,
			requestURI:    requestURI,
			want:          false,
		},
		{
			name:          "expired timestamp",
			body:          body,
			signature:     validSignature,
			timestamp:     fmt.Sprintf("%d", time.Now().UnixMilli()-400000), // 6+ minutes old
			requestMethod: requestMethod,
			requestURI:    requestURI,
			want:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := verifier.VerifyHubSpotSignature(tt.body, tt.signature, tt.timestamp, tt.requestMethod, tt.requestURI)
			if got != tt.want {
				t.Errorf("VerifyHubSpotSignature() = %v, want %v", got, tt.want)
			}
		})
	}
}

// =============================================================================
// NOTION SIGNATURE TESTS
// =============================================================================

func TestVerifyNotionSignature(t *testing.T) {
	verifier := NewSignatureVerifier(map[string]string{
		"notion": "test-notion-secret",
	})

	body := []byte(`{"event":"page.updated"}`)

	// Calculate valid signature
	mac := hmac.New(sha256.New, []byte("test-notion-secret"))
	mac.Write(body)
	validSignature := hex.EncodeToString(mac.Sum(nil))

	tests := []struct {
		name      string
		body      []byte
		signature string
		want      bool
	}{
		{
			name:      "valid signature",
			body:      body,
			signature: validSignature,
			want:      true,
		},
		{
			name:      "invalid signature",
			body:      body,
			signature: "invalid",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := verifier.VerifyNotionSignature(tt.body, tt.signature)
			if got != tt.want {
				t.Errorf("VerifyNotionSignature() = %v, want %v", got, tt.want)
			}
		})
	}
}

// =============================================================================
// AIRTABLE SIGNATURE TESTS
// =============================================================================

func TestVerifyAirtableSignature(t *testing.T) {
	verifier := NewSignatureVerifier(map[string]string{
		"airtable": "test-airtable-secret",
	})

	body := []byte(`{"records":[]}`)
	timestamp := fmt.Sprintf("%d", time.Now().Unix())

	// Calculate valid signature
	sourceString := timestamp + "." + string(body)
	mac := hmac.New(sha256.New, []byte("test-airtable-secret"))
	mac.Write([]byte(sourceString))
	validSignature := hex.EncodeToString(mac.Sum(nil))

	tests := []struct {
		name      string
		body      []byte
		signature string
		timestamp string
		want      bool
	}{
		{
			name:      "valid signature",
			body:      body,
			signature: validSignature,
			timestamp: timestamp,
			want:      true,
		},
		{
			name:      "invalid signature",
			body:      body,
			signature: "invalid",
			timestamp: timestamp,
			want:      false,
		},
		{
			name:      "expired timestamp",
			body:      body,
			signature: validSignature,
			timestamp: fmt.Sprintf("%d", time.Now().Unix()-400),
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := verifier.VerifyAirtableSignature(tt.body, tt.signature, tt.timestamp)
			if got != tt.want {
				t.Errorf("VerifyAirtableSignature() = %v, want %v", got, tt.want)
			}
		})
	}
}

// =============================================================================
// FATHOM SIGNATURE TESTS
// =============================================================================

func TestVerifyFathomSignature(t *testing.T) {
	verifier := NewSignatureVerifier(map[string]string{
		"fathom": "test-fathom-secret",
	})

	body := []byte(`{"type":"meeting.completed"}`)

	// Calculate valid signature
	mac := hmac.New(sha256.New, []byte("test-fathom-secret"))
	mac.Write(body)
	validSignature := hex.EncodeToString(mac.Sum(nil))

	tests := []struct {
		name      string
		body      []byte
		signature string
		want      bool
	}{
		{
			name:      "valid signature",
			body:      body,
			signature: validSignature,
			want:      true,
		},
		{
			name:      "invalid signature",
			body:      body,
			signature: "invalid",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := verifier.VerifyFathomSignature(tt.body, tt.signature)
			if got != tt.want {
				t.Errorf("VerifyFathomSignature() = %v, want %v", got, tt.want)
			}
		})
	}
}

// =============================================================================
// GOOGLE CHANNEL TOKEN TESTS
// =============================================================================

func TestVerifyGoogleChannelToken(t *testing.T) {
	verifier := NewSignatureVerifier(map[string]string{})

	tests := []struct {
		name  string
		token string
		want  bool
	}{
		{
			name:  "valid token",
			token: "user-123-abc",
			want:  true,
		},
		{
			name:  "empty token",
			token: "",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := verifier.VerifyGoogleChannelToken(tt.token)
			if got != tt.want {
				t.Errorf("VerifyGoogleChannelToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

// =============================================================================
// MICROSOFT CLIENT STATE TESTS
// =============================================================================

func TestVerifyMicrosoftClientState(t *testing.T) {
	verifier := NewSignatureVerifier(map[string]string{
		"microsoft": "expected-client-state",
	})

	tests := []struct {
		name        string
		clientState string
		want        bool
	}{
		{
			name:        "valid client state",
			clientState: "expected-client-state",
			want:        true,
		},
		{
			name:        "invalid client state",
			clientState: "wrong-state",
			want:        false,
		},
		{
			name:        "no secret configured",
			clientState: "anything",
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := verifier
			if tt.name == "no secret configured" {
				v = NewSignatureVerifier(map[string]string{})
			}

			got := v.VerifyMicrosoftClientState(tt.clientState)
			if got != tt.want {
				t.Errorf("VerifyMicrosoftClientState() = %v, want %v", got, tt.want)
			}
		})
	}
}

// =============================================================================
// GENERIC HMAC-SHA256 TESTS
// =============================================================================

func TestVerifyHMACSHA256(t *testing.T) {
	verifier := NewSignatureVerifier(map[string]string{
		"test-provider": "test-secret",
	})

	body := []byte(`{"test":"data"}`)

	// Calculate valid signature
	mac := hmac.New(sha256.New, []byte("test-secret"))
	mac.Write(body)
	validSignature := hex.EncodeToString(mac.Sum(nil))

	tests := []struct {
		name      string
		provider  string
		body      []byte
		signature string
		want      bool
	}{
		{
			name:      "valid signature",
			provider:  "test-provider",
			body:      body,
			signature: validSignature,
			want:      true,
		},
		{
			name:      "invalid signature",
			provider:  "test-provider",
			body:      body,
			signature: "invalid",
			want:      false,
		},
		{
			name:      "no secret configured",
			provider:  "unknown-provider",
			body:      body,
			signature: "anything",
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := verifier.VerifyHMACSHA256(tt.provider, tt.body, tt.signature)
			if got != tt.want {
				t.Errorf("VerifyHMACSHA256() = %v, want %v", got, tt.want)
			}
		})
	}
}

// =============================================================================
// COMPUTE HMAC-SHA256 TESTS (Helper Function)
// =============================================================================

func TestComputeHMACSHA256(t *testing.T) {
	secret := "test-secret"
	body := []byte(`{"test":"data"}`)

	// Calculate expected signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))

	got := ComputeHMACSHA256(secret, body)
	if got != expected {
		t.Errorf("ComputeHMACSHA256() = %v, want %v", got, expected)
	}
}

// =============================================================================
// SET/GET SECRET TESTS
// =============================================================================

func TestSignatureVerifier_SetGetSecret(t *testing.T) {
	verifier := NewSignatureVerifier(map[string]string{})

	// Initially empty
	if got := verifier.GetSecret("test"); got != "" {
		t.Errorf("GetSecret() = %v, want empty string", got)
	}

	// Set secret
	verifier.SetSecret("test", "secret123")

	// Verify it was set
	if got := verifier.GetSecret("test"); got != "secret123" {
		t.Errorf("GetSecret() = %v, want secret123", got)
	}
}
