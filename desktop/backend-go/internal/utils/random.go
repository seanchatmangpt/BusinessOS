package utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

// GenerateRandomBytes generates cryptographically secure random bytes.
// Returns an error if the random number generator fails (extremely rare).
func GenerateRandomBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return bytes, nil
}

// GenerateRandomHex generates a random hex-encoded string.
// The output length will be 2 * byteLength (each byte = 2 hex chars).
func GenerateRandomHex(byteLength int) (string, error) {
	bytes, err := GenerateRandomBytes(byteLength)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateRandomBase64 generates a random base64-encoded string.
// Uses URL-safe encoding (base64.URLEncoding).
func GenerateRandomBase64(byteLength int) (string, error) {
	bytes, err := GenerateRandomBytes(byteLength)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// GenerateSessionToken generates a 32-byte base64-encoded session token.
// Output length: 44 characters (32 bytes base64 encoded).
func GenerateSessionToken() (string, error) {
	return GenerateRandomBase64(32)
}

// GenerateUserID generates a 16-byte base64-encoded user ID, truncated to 22 characters.
// This matches the existing auth system's user ID format.
func GenerateUserID() (string, error) {
	id, err := GenerateRandomBase64(16)
	if err != nil {
		return "", err
	}
	if len(id) < 22 {
		return "", fmt.Errorf("generated ID too short: %d < 22", len(id))
	}
	return id[:22], nil
}

// GenerateSessionID generates a 16-byte base64-encoded session ID, truncated to 22 characters.
// This matches the existing auth system's session ID format.
func GenerateSessionID() (string, error) {
	id, err := GenerateRandomBase64(16)
	if err != nil {
		return "", err
	}
	if len(id) < 22 {
		return "", fmt.Errorf("generated ID too short: %d < 22", len(id))
	}
	return id[:22], nil
}

// GenerateOAuthState generates a 32-byte base64-encoded state string for OAuth flows.
// Used for CSRF protection in OAuth authentication.
func GenerateOAuthState() (string, error) {
	return GenerateRandomBase64(32)
}

// GenerateShareID generates an 8-byte hex-encoded share ID.
// Output length: 16 characters (8 bytes = 16 hex chars).
func GenerateShareID() (string, error) {
	return GenerateRandomHex(8)
}

// GenerateShareToken generates a 16-byte hex-encoded share token.
// Output length: 32 characters (16 bytes = 32 hex chars).
func GenerateShareToken() (string, error) {
	return GenerateRandomHex(16)
}

// GenerateNonce generates a random nonce of the specified byte length.
// Commonly used for cryptographic operations requiring unique values.
func GenerateNonce(byteLength int) ([]byte, error) {
	return GenerateRandomBytes(byteLength)
}

// MustGenerateRandomHex is like GenerateRandomHex but panics on error.
// Only use when random generation failure is unrecoverable.
func MustGenerateRandomHex(byteLength int) string {
	result, err := GenerateRandomHex(byteLength)
	if err != nil {
		panic(fmt.Sprintf("failed to generate random hex: %v", err))
	}
	return result
}

// MustGenerateSessionToken is like GenerateSessionToken but panics on error.
// Only use when random generation failure is unrecoverable.
func MustGenerateSessionToken() string {
	result, err := GenerateSessionToken()
	if err != nil {
		panic(fmt.Sprintf("failed to generate session token: %v", err))
	}
	return result
}
