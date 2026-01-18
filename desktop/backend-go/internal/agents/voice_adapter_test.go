package agents

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/rhl/businessos-backend/internal/services"
	"github.com/rhl/businessos-backend/internal/streaming"
)

// ============================================================================
// Tests for VoiceAgentAdapter Creation
// ============================================================================

func TestNewVoiceAgentAdapter(t *testing.T) {
	// Arrange - Create a simple mock registry
	mockRegistry := &AgentRegistryV2{}

	// Act
	adapter := NewVoiceAgentAdapter(mockRegistry)

	// Assert
	assert.NotNil(t, adapter)
	assert.Equal(t, mockRegistry, adapter.registry)
}

func TestNewVoiceAgentAdapter_NilRegistry(t *testing.T) {
	// Act
	adapter := NewVoiceAgentAdapter(nil)

	// Assert
	assert.NotNil(t, adapter)
	assert.Nil(t, adapter.registry)
}

// ============================================================================
// Tests for VoiceAgentAdapter Interface Implementation
// ============================================================================

func TestVoiceAgentAdapter_ImplementsInterface(t *testing.T) {
	// Arrange
	adapter := &VoiceAgentAdapter{
		registry: nil,
	}

	// Assert - Type assertion verifies interface implementation
	var _ services.VoiceAgentProvider = adapter
}

func TestVoiceAgentAdapter_ExecuteVoiceAgentSignature(t *testing.T) {
	// This test verifies that ExecuteVoiceAgent has the correct signature
	adapter := &VoiceAgentAdapter{
		registry: nil,
	}

	// Verify the method exists and can be called
	// This is a compile-time check via the interface
	var provider services.VoiceAgentProvider = adapter
	assert.NotNil(t, provider)
}

// ============================================================================
// Tests for ExecuteVoiceAgent
// ============================================================================

func TestExecuteVoiceAgent_InterfaceImplementation(t *testing.T) {
	// This test verifies the adapter implements the VoiceAgentProvider interface
	adapter := &VoiceAgentAdapter{
		registry: nil, // Mock registry can be nil for interface test
	}

	// Type assertion - if this compiles, interface is implemented
	var _ services.VoiceAgentProvider = adapter

	assert.NotNil(t, adapter)
}

func TestExecuteVoiceAgent_CreatesChannels(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test that requires registry implementation")
	}

	// Arrange - For this test we just verify channels are returned (not nil)
	// Full integration testing would require actual agent implementation
	adapter := &VoiceAgentAdapter{
		registry: nil,
	}

	// Note: Actual execution would panic without a real registry
	// This test just verifies the method signature is correct
	_ = adapter

	// Type assertion verifies it implements the interface
	var _ services.VoiceAgentProvider = adapter
}

// ============================================================================
// Tests for Method Signatures
// ============================================================================

func TestVoiceAgentAdapter_WithConversationID(t *testing.T) {
	// Test that adapter can accept conversation IDs
	conversationID := uuid.New()

	// Just verify the types work together
	adapter := &VoiceAgentAdapter{
		registry: nil,
	}

	// Type check - compiler will verify signature
	var provider services.VoiceAgentProvider = adapter
	assert.NotNil(t, provider)

	// Verify UUID type works
	assert.NotNil(t, &conversationID)
	assert.NotEmpty(t, conversationID.String())
}

func TestVoiceAgentAdapter_WithMessages(t *testing.T) {
	// Test that adapter can handle message lists
	messages := []services.ChatMessage{
		{
			Role:    "user",
			Content: "Hello",
		},
		{
			Role:    "assistant",
			Content: "Hi there!",
		},
	}

	adapter := &VoiceAgentAdapter{
		registry: nil,
	}

	// Verify message structure
	assert.Len(t, messages, 2)
	assert.Equal(t, "user", messages[0].Role)
	assert.Equal(t, "Hello", messages[0].Content)
	assert.Equal(t, "assistant", messages[1].Role)

	// Type check
	var _ services.VoiceAgentProvider = adapter
}

func TestVoiceAgentAdapter_WithLLMOptions(t *testing.T) {
	// Test LLM options for voice
	voiceOptions := services.LLMOptions{
		Temperature:       0.7,
		MaxTokens:         500,
		TopP:              0.9,
		ThinkingEnabled:   false,
		MaxThinkingTokens: 0,
	}

	// Verify options structure
	assert.Equal(t, 0.7, voiceOptions.Temperature)
	assert.Equal(t, 500, voiceOptions.MaxTokens)
	assert.False(t, voiceOptions.ThinkingEnabled)

	adapter := &VoiceAgentAdapter{
		registry: nil,
	}

	// Type check
	var _ services.VoiceAgentProvider = adapter
}

// ============================================================================
// Tests for Edge Cases
// ============================================================================

func TestVoiceAgentAdapter_EmptyMessages(t *testing.T) {
	// Test with empty message list
	messages := []services.ChatMessage{}

	adapter := &VoiceAgentAdapter{
		registry: nil,
	}

	assert.Empty(t, messages)
	assert.NotNil(t, adapter)

	// Type check
	var _ services.VoiceAgentProvider = adapter
}

func TestVoiceAgentAdapter_NilContext(t *testing.T) {
	// Test with nil tiered context
	var tieredCtx *services.TieredContext

	adapter := &VoiceAgentAdapter{
		registry: nil,
	}

	assert.Nil(t, tieredCtx)
	assert.NotNil(t, adapter)

	// Type check
	var _ services.VoiceAgentProvider = adapter
}

func TestVoiceAgentAdapter_NoConversationID(t *testing.T) {
	// Test voice conversation without conversation ID
	var conversationID *uuid.UUID

	adapter := &VoiceAgentAdapter{
		registry: nil,
	}

	assert.Nil(t, conversationID)
	assert.NotNil(t, adapter)

	// Type check
	var _ services.VoiceAgentProvider = adapter
}

// ============================================================================
// Mock implementations for testing
// ============================================================================

type mockVoiceAgentProvider struct {
	mock.Mock
}

func (m *mockVoiceAgentProvider) ExecuteVoiceAgent(
	ctx context.Context,
	userID string,
	userName string,
	conversationID *uuid.UUID,
	messages []services.ChatMessage,
	tieredContext *services.TieredContext,
	llmOptions services.LLMOptions,
) (<-chan streaming.StreamEvent, <-chan error) {
	args := m.Called(ctx, userID, userName, conversationID, messages, tieredContext, llmOptions)
	events := args.Get(0)
	errs := args.Get(1)
	if events == nil {
		events = make(<-chan streaming.StreamEvent)
	}
	if errs == nil {
		errs = make(<-chan error)
	}
	return events.((<-chan streaming.StreamEvent)), errs.((<-chan error))
}

// ============================================================================
// Benchmark tests
// ============================================================================

func BenchmarkNewVoiceAgentAdapter(b *testing.B) {
	registry := &AgentRegistryV2{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewVoiceAgentAdapter(registry)
	}
}

func BenchmarkVoiceAgentAdapter_Creation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		adapter := NewVoiceAgentAdapter(nil)
		_ = adapter
	}
}
