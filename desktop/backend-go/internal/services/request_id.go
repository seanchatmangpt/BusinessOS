package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// requestIDKey is the context key for request IDs
	requestIDKey contextKey = "request_id"
	// sessionIDKey is the context key for session IDs
	sessionIDKey contextKey = "session_id"
)

// GenerateRequestID generates a new request ID with "req_" prefix
// Format: req_{uuid}
func GenerateRequestID() string {
	return fmt.Sprintf("req_%s", uuid.New().String())
}

// GenerateSessionID generates a new session ID with "sess_" prefix
// Format: sess_{uuid}
func GenerateSessionID() string {
	return fmt.Sprintf("sess_%s", uuid.New().String())
}

// AddRequestIDToContext adds a request ID to the context
// If the request ID is empty, it generates a new one
func AddRequestIDToContext(ctx context.Context, requestID string) context.Context {
	if requestID == "" {
		requestID = GenerateRequestID()
	}
	return context.WithValue(ctx, requestIDKey, requestID)
}

// GetRequestIDFromContext extracts the request ID from context
// Returns empty string if not found
func GetRequestIDFromContext(ctx context.Context) string {
	if requestID, ok := ctx.Value(requestIDKey).(string); ok {
		return requestID
	}
	return ""
}

// AddSessionIDToContext adds a session ID to the context
// If the session ID is empty, it generates a new one
func AddSessionIDToContext(ctx context.Context, sessionID string) context.Context {
	if sessionID == "" {
		sessionID = GenerateSessionID()
	}
	return context.WithValue(ctx, sessionIDKey, sessionID)
}

// GetSessionIDFromContext extracts the session ID from context
// Returns empty string if not found
func GetSessionIDFromContext(ctx context.Context) string {
	if sessionID, ok := ctx.Value(sessionIDKey).(string); ok {
		return sessionID
	}
	return ""
}

// GetOrGenerateRequestID gets existing request ID from context or generates a new one
// This is useful for ensuring every operation has a request ID
func GetOrGenerateRequestID(ctx context.Context) (context.Context, string) {
	requestID := GetRequestIDFromContext(ctx)
	if requestID == "" {
		requestID = GenerateRequestID()
		ctx = AddRequestIDToContext(ctx, requestID)
	}
	return ctx, requestID
}

// GetOrGenerateSessionID gets existing session ID from context or generates a new one
// This is useful for ensuring every voice session has a session ID
func GetOrGenerateSessionID(ctx context.Context) (context.Context, string) {
	sessionID := GetSessionIDFromContext(ctx)
	if sessionID == "" {
		sessionID = GenerateSessionID()
		ctx = AddSessionIDToContext(ctx, sessionID)
	}
	return ctx, sessionID
}
