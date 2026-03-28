package utils

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// setupTestGin initializes gin in test mode and returns a context for testing.
func setupTestGin(method, path string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, path, nil)
	return c, w
}

// discardLogger returns a no-op logger for tests that do not inspect log output.
func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(&strings.Builder{}, &slog.HandlerOptions{Level: slog.LevelError}))
}

// ---------------------------------------------------------------------------
// ErrorBuilder constructors
// ---------------------------------------------------------------------------

func TestNewErrorBuilder_InitializesWithEmptyMessage(t *testing.T) {
	eb := NewErrorBuilder(discardLogger())

	if eb.statusCode != 0 {
		t.Errorf("expected statusCode 0, got %d", eb.statusCode)
	}
	if eb.code != "" {
		t.Errorf("expected empty code, got %q", eb.code)
	}
	if eb.message != "" {
		t.Errorf("expected empty message, got %q", eb.message)
	}
	if eb.details == nil {
		t.Error("expected details map to be initialized, got nil")
	}
}

func TestUnauthorized_DefaultMessage_WhenEmpty(t *testing.T) {
	eb := Unauthorized(discardLogger(), "")

	if eb.statusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", eb.statusCode)
	}
	if eb.code != ErrCodeUnauthorized {
		t.Errorf("expected %q, got %q", ErrCodeUnauthorized, eb.code)
	}
	if eb.message != "Not authenticated" {
		t.Errorf("expected default message, got %q", eb.message)
	}
}

func TestUnauthorized_UsesProvidedMessage(t *testing.T) {
	eb := Unauthorized(discardLogger(), "token expired")

	if eb.message != "token expired" {
		t.Errorf("expected %q, got %q", "token expired", eb.message)
	}
}

func TestBadRequest_Returns400(t *testing.T) {
	eb := BadRequest(discardLogger(), "bad input")

	if eb.statusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", eb.statusCode)
	}
	if eb.code != ErrCodeBadRequest {
		t.Errorf("expected %q, got %q", ErrCodeBadRequest, eb.code)
	}
}

func TestNotFound_FormatsResourceMessage(t *testing.T) {
	eb := NotFound(discardLogger(), "User")

	if eb.statusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", eb.statusCode)
	}
	if eb.message != "User not found" {
		t.Errorf("expected %q, got %q", "User not found", eb.message)
	}
}

func TestInternalError_Returns500(t *testing.T) {
	eb := InternalError(discardLogger(), "db connection failed")

	if eb.statusCode != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", eb.statusCode)
	}
	if eb.code != ErrCodeInternal {
		t.Errorf("expected %q, got %q", ErrCodeInternal, eb.code)
	}
}

func TestValidationError_SetsFieldDetail(t *testing.T) {
	eb := ValidationError(discardLogger(), "email")

	if eb.statusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", eb.statusCode)
	}
	if eb.code != ErrCodeValidation {
		t.Errorf("expected %q, got %q", ErrCodeValidation, eb.code)
	}
	if eb.message != "Validation failed" {
		t.Errorf("expected %q, got %q", "Validation failed", eb.message)
	}
	field, ok := eb.details["field"]
	if !ok || field != "email" {
		t.Errorf("expected field=%q in details, got %v", "email", eb.details["field"])
	}
}

func TestForbidden_DefaultMessage_WhenEmpty(t *testing.T) {
	eb := Forbidden(discardLogger(), "")

	if eb.statusCode != http.StatusForbidden {
		t.Errorf("expected 403, got %d", eb.statusCode)
	}
	if eb.message != "Access denied" {
		t.Errorf("expected default message, got %q", eb.message)
	}
}

func TestForbidden_UsesProvidedMessage(t *testing.T) {
	eb := Forbidden(discardLogger(), "insufficient scope")

	if eb.message != "insufficient scope" {
		t.Errorf("expected %q, got %q", "insufficient scope", eb.message)
	}
}

func TestConflict_Returns409(t *testing.T) {
	eb := Conflict(discardLogger(), "duplicate email")

	if eb.statusCode != http.StatusConflict {
		t.Errorf("expected 409, got %d", eb.statusCode)
	}
	if eb.code != ErrCodeConflict {
		t.Errorf("expected %q, got %q", ErrCodeConflict, eb.code)
	}
}

func TestServiceUnavailable_FormatsServiceMessage(t *testing.T) {
	eb := ServiceUnavailable(discardLogger(), "payment-gateway")

	if eb.statusCode != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", eb.statusCode)
	}
	if eb.message != "payment-gateway is temporarily unavailable" {
		t.Errorf("expected %q, got %q", "payment-gateway is temporarily unavailable", eb.message)
	}
}

func TestNotImplemented_FormatsFeatureMessage(t *testing.T) {
	eb := NotImplemented(discardLogger(), "AI suggestions")

	if eb.statusCode != http.StatusNotImplemented {
		t.Errorf("expected 501, got %d", eb.statusCode)
	}
	if eb.message != "AI suggestions is not implemented" {
		t.Errorf("expected %q, got %q", "AI suggestions is not implemented", eb.message)
	}
}

func TestTooManyRequests_FormatsResourceMessage(t *testing.T) {
	eb := TooManyRequests(discardLogger(), "/api/search")

	if eb.statusCode != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", eb.statusCode)
	}
	if eb.message != "Too many requests to /api/search" {
		t.Errorf("expected %q, got %q", "Too many requests to /api/search", eb.message)
	}
}

// ---------------------------------------------------------------------------
// ErrorBuilder fluent API
// ---------------------------------------------------------------------------

func TestWithDetails_AddsKeyValueToDetails(t *testing.T) {
	eb := BadRequest(discardLogger(), "bad").
		WithDetails("max_length", 255).
		WithDetails("field", "username")

	if eb.details["max_length"] != 255 {
		t.Errorf("expected max_length=255, got %v", eb.details["max_length"])
	}
	if eb.details["field"] != "username" {
		t.Errorf("expected field=username, got %v", eb.details["field"])
	}
}

func TestWithDetails_OverwritesExistingKey(t *testing.T) {
	eb := BadRequest(discardLogger(), "bad").
		WithDetails("retry_after", 10).
		WithDetails("retry_after", 60)

	if eb.details["retry_after"] != 60 {
		t.Errorf("expected retry_after=60 after overwrite, got %v", eb.details["retry_after"])
	}
}

func TestWithError_NilError_DoesNotAddDetail(t *testing.T) {
	eb := BadRequest(discardLogger(), "bad").
		WithError(nil)

	if _, exists := eb.details["error"]; exists {
		t.Error("expected no 'error' key in details when err is nil")
	}
}

func TestWithError_NonNilError_AddsErrorMessage(t *testing.T) {
	eb := InternalError(discardLogger(), "oops").
		WithError(errors.New("connection refused"))

	msg, ok := eb.details["error"].(string)
	if !ok {
		t.Fatalf("expected string in details[error], got %T", eb.details["error"])
	}
	if msg != "connection refused" {
		t.Errorf("expected %q, got %q", "connection refused", msg)
	}
}

func TestWithMessage_OverridesExistingMessage(t *testing.T) {
	eb := NotFound(discardLogger(), "Document").
		WithMessage("Document was deleted")

	if eb.message != "Document was deleted" {
		t.Errorf("expected overridden message, got %q", eb.message)
	}
}

// ---------------------------------------------------------------------------
// Respond — writes JSON to gin context
// ---------------------------------------------------------------------------

func TestRespond_WritesCorrectStatusCode(t *testing.T) {
	c, w := setupTestGin(http.MethodGet, "/api/resource")

	NotFound(discardLogger(), "Resource").Respond(c)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestRespond_WritesJSONBody(t *testing.T) {
	c, w := setupTestGin(http.MethodGet, "/api/resource")

	BadRequest(discardLogger(), "missing field").Respond(c)

	var body ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if body.Error.Code != ErrCodeBadRequest {
		t.Errorf("expected code %q, got %q", ErrCodeBadRequest, body.Error.Code)
	}
	if body.Error.Message != "missing field" {
		t.Errorf("expected message %q, got %q", "missing field", body.Error.Message)
	}
}

func TestRespond_IncludesDetailsInJSON(t *testing.T) {
	c, w := setupTestGin(http.MethodGet, "/api/resource")

	ValidationError(discardLogger(), "email").
		WithDetails("reason", "invalid format").
		Respond(c)

	var body ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if body.Error.Details["field"] != "email" {
		t.Errorf("expected field=email in JSON details, got %v", body.Error.Details["field"])
	}
	if body.Error.Details["reason"] != "invalid format" {
		t.Errorf("expected reason in JSON details, got %v", body.Error.Details["reason"])
	}
}

func TestRespond_OmitsDetailsWhenEmpty(t *testing.T) {
	c, w := setupTestGin(http.MethodGet, "/api/resource")

	BadRequest(discardLogger(), "simple error").Respond(c)

	var body ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if body.Error.Details != nil {
		t.Errorf("expected nil details (omitted), got %v", body.Error.Details)
	}
}

func TestRespond_WithNilLogger_DoesNotPanic(t *testing.T) {
	c, w := setupTestGin(http.MethodGet, "/api/resource")

	eb := BadRequest(nil, "no logger").
		WithDetails("key", "val")
	eb.Respond(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 even with nil logger, got %d", w.Code)
	}
}

func TestRespond_IncludesUserID_WhenSetOnContext(t *testing.T) {
	c, w := setupTestGin(http.MethodGet, "/api/resource")
	c.Set("user_id", "user-42")

	Unauthorized(discardLogger(), "").Respond(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
	// The key assertion: no panic when user_id is present
}

// ---------------------------------------------------------------------------
// Convenience Respond* functions
// ---------------------------------------------------------------------------

func TestRespondUnauthorized_Sends401(t *testing.T) {
	c, w := setupTestGin(http.MethodGet, "/api/secret")

	RespondUnauthorized(c, discardLogger())

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}

	var body ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &body)
	if body.Error.Code != ErrCodeUnauthorized {
		t.Errorf("expected code %q, got %q", ErrCodeUnauthorized, body.Error.Code)
	}
	if body.Error.Message != "Not authenticated" {
		t.Errorf("expected %q, got %q", "Not authenticated", body.Error.Message)
	}
}

func TestRespondInvalidID_Sends400WithFieldMessage(t *testing.T) {
	c, w := setupTestGin(http.MethodPost, "/api/users")

	RespondInvalidID(c, discardLogger(), "user_id")

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	var body ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &body)
	if body.Error.Code != ErrCodeValidation {
		t.Errorf("expected %q, got %q", ErrCodeValidation, body.Error.Code)
	}
	if body.Error.Message != "Invalid user_id" {
		t.Errorf("expected %q, got %q", "Invalid user_id", body.Error.Message)
	}
}

func TestRespondNotFound_Sends404WithResource(t *testing.T) {
	c, w := setupTestGin(http.MethodGet, "/api/orders/99")

	RespondNotFound(c, discardLogger(), "Order")

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	var body ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &body)
	if body.Error.Message != "Order not found" {
		t.Errorf("expected %q, got %q", "Order not found", body.Error.Message)
	}
}

func TestRespondInternalError_IncludesErrorDetail(t *testing.T) {
	c, w := setupTestGin(http.MethodPost, "/api/data")

	RespondInternalError(c, discardLogger(), "save record", errors.New("disk full"))

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	var body ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &body)
	if body.Error.Code != ErrCodeInternal {
		t.Errorf("expected %q, got %q", ErrCodeInternal, body.Error.Code)
	}
	if body.Error.Message != "Failed to save record" {
		t.Errorf("expected %q, got %q", "Failed to save record", body.Error.Message)
	}
	if body.Error.Details["error"] != "disk full" {
		t.Errorf("expected error detail %q, got %v", "disk full", body.Error.Details["error"])
	}
}

func TestRespondInternalError_WithNilError_DoesNotPanic(t *testing.T) {
	c, w := setupTestGin(http.MethodPost, "/api/data")

	RespondInternalError(c, discardLogger(), "save record", nil)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestRespondBadRequest_Sends400(t *testing.T) {
	c, w := setupTestGin(http.MethodPost, "/api/items")

	RespondBadRequest(c, discardLogger(), "invalid JSON")

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	var body ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &body)
	if body.Error.Message != "invalid JSON" {
		t.Errorf("expected %q, got %q", "invalid JSON", body.Error.Message)
	}
}

func TestRespondForbidden_Sends403(t *testing.T) {
	c, w := setupTestGin(http.MethodDelete, "/api/admin/users")

	RespondForbidden(c, discardLogger(), "admin required")

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}

	var body ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &body)
	if body.Error.Code != ErrCodeForbidden {
		t.Errorf("expected %q, got %q", ErrCodeForbidden, body.Error.Code)
	}
}

func TestRespondConflict_Sends409(t *testing.T) {
	c, w := setupTestGin(http.MethodPost, "/api/users")

	RespondConflict(c, discardLogger(), "email already exists")

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", w.Code)
	}
}

func TestRespondInvalidRequest_IncludesWrappedError(t *testing.T) {
	c, w := setupTestGin(http.MethodPost, "/api/data")

	RespondInvalidRequest(c, discardLogger(), errors.New("unexpected EOF"))

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	var body ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &body)
	if body.Error.Message != "Invalid request body" {
		t.Errorf("expected %q, got %q", "Invalid request body", body.Error.Message)
	}
	if body.Error.Details["error"] != "unexpected EOF" {
		t.Errorf("expected error detail, got %v", body.Error.Details["error"])
	}
}

func TestRespondServiceUnavailable_Sends503(t *testing.T) {
	c, w := setupTestGin(http.MethodGet, "/api/ai/generate")

	RespondServiceUnavailable(c, discardLogger(), "LLM service")

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", w.Code)
	}

	var body ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &body)
	if body.Error.Message != "LLM service is temporarily unavailable" {
		t.Errorf("expected %q, got %q", "LLM service is temporarily unavailable", body.Error.Message)
	}
}

func TestRespondNotImplemented_Sends501(t *testing.T) {
	c, w := setupTestGin(http.MethodPost, "/api/batch")

	RespondNotImplemented(c, discardLogger(), "batch processing")

	if w.Code != http.StatusNotImplemented {
		t.Errorf("expected 501, got %d", w.Code)
	}

	var body ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &body)
	if body.Error.Message != "batch processing is not implemented" {
		t.Errorf("expected %q, got %q", "batch processing is not implemented", body.Error.Message)
	}
}

func TestRespondTooManyRequests_Sends429(t *testing.T) {
	c, w := setupTestGin(http.MethodGet, "/api/search")

	RespondTooManyRequests(c, discardLogger(), "/api/search")

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", w.Code)
	}

	var body ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &body)
	if body.Error.Message != "Too many requests to /api/search" {
		t.Errorf("expected %q, got %q", "Too many requests to /api/search", body.Error.Message)
	}
}

// ---------------------------------------------------------------------------
// Error code constants
// ---------------------------------------------------------------------------

func TestErrorCodeConstants_AreNonEmpty(t *testing.T) {
	codes := []ErrorCode{
		ErrCodeValidation, ErrCodeUnauthorized, ErrCodeForbidden, ErrCodeNotFound,
		ErrCodeConflict, ErrCodeRateLimited, ErrCodeInternal, ErrCodeBadRequest,
		ErrCodeUnprocessable, ErrCodeServiceUnavailable, ErrCodeNotImplemented,
		ErrCodeTooManyRequests,
	}
	for _, code := range codes {
		if string(code) == "" {
			t.Error("error code constant must not be empty")
		}
	}
}
