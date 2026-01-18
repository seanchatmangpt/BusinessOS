package services

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/rhl/businessos-backend/internal/config"
	"github.com/rhl/businessos-backend/internal/streaming"
)

// ============================================================================
// Tests for VoiceUserContext
// ============================================================================

func TestVoiceUserContext_Creation(t *testing.T) {
	// Arrange
	ctx := &VoiceUserContext{
		UserID:         "user123",
		Username:       "testuser",
		Email:          "test@example.com",
		DisplayName:    "Test User",
		WorkspaceID:    "ws123",
		WorkspaceName:  "Test Workspace",
		Role:           "admin",
		Title:          "Engineer",
		Timezone:       "UTC",
		OutputStyle:    "concise",
		ExpertiseAreas: []string{"Go", "Python", "AI"},
	}

	// Assert
	assert.Equal(t, "user123", ctx.UserID)
	assert.Equal(t, "testuser", ctx.Username)
	assert.Equal(t, "test@example.com", ctx.Email)
	assert.Equal(t, "Test User", ctx.DisplayName)
	assert.Equal(t, "ws123", ctx.WorkspaceID)
	assert.Equal(t, "Test Workspace", ctx.WorkspaceName)
	assert.Equal(t, "admin", ctx.Role)
	assert.Equal(t, "Engineer", ctx.Title)
	assert.Equal(t, "UTC", ctx.Timezone)
	assert.Equal(t, "concise", ctx.OutputStyle)
	assert.Len(t, ctx.ExpertiseAreas, 3)
	assert.Contains(t, ctx.ExpertiseAreas, "Go")
}

func TestVoiceUserContext_EmptyValues(t *testing.T) {
	// Arrange
	ctx := &VoiceUserContext{}

	// Assert
	assert.Empty(t, ctx.UserID)
	assert.Empty(t, ctx.Username)
	assert.Empty(t, ctx.DisplayName)
	assert.Empty(t, ctx.WorkspaceID)
	assert.Nil(t, ctx.ExpertiseAreas)
}

// ============================================================================
// Tests for VoiceSession
// ============================================================================

func TestVoiceSession_Creation(t *testing.T) {
	// Arrange
	now := time.Now()
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Act
	session := &VoiceSession{
		SessionID:   "session123",
		UserID:      "user123",
		WorkspaceID: "ws123",
		AgentRole:   "assistant",
		State:       0, // IDLE
		CreatedAt:   now,
		UpdatedAt:   now,
		audioBuffer: make([]byte, 0),
		Messages:    make([]Message, 0),
		cancel:      cancel,
	}

	// Assert
	assert.Equal(t, "session123", session.SessionID)
	assert.Equal(t, "user123", session.UserID)
	assert.Equal(t, "ws123", session.WorkspaceID)
	assert.Equal(t, "assistant", session.AgentRole)
	assert.NotNil(t, session.audioBuffer)
	assert.NotNil(t, session.Messages)
	assert.NotNil(t, session.cancel)
}

func TestVoiceSession_MessageHandling(t *testing.T) {
	// Arrange
	session := &VoiceSession{
		Messages:   make([]Message, 0),
		MessagesMu: sync.Mutex{},
	}

	// Act - Add user message
	userMsg := Message{
		Role:      "user",
		Content:   "Hello, agent!",
		Timestamp: time.Now(),
	}
	session.Messages = append(session.Messages, userMsg)

	// Act - Add agent message
	agentMsg := Message{
		Role:      "agent",
		Content:   "Hello! How can I help?",
		Timestamp: time.Now(),
	}
	session.Messages = append(session.Messages, agentMsg)

	// Assert
	assert.Len(t, session.Messages, 2)
	assert.Equal(t, "user", session.Messages[0].Role)
	assert.Equal(t, "Hello, agent!", session.Messages[0].Content)
	assert.Equal(t, "agent", session.Messages[1].Role)
	assert.Equal(t, "Hello! How can I help?", session.Messages[1].Content)
}

func TestVoiceSession_UserContextCaching(t *testing.T) {
	// Arrange
	session := &VoiceSession{
		UserContext: nil,
	}

	userCtx := &VoiceUserContext{
		UserID:      "user123",
		DisplayName: "Test User",
		Username:    "testuser",
	}

	// Act - Cache context
	session.UserContext = userCtx

	// Assert
	assert.NotNil(t, session.UserContext)
	assert.Equal(t, "Test User", session.UserContext.DisplayName)
	assert.Equal(t, "testuser", session.UserContext.Username)
}

// ============================================================================
// Tests for LLM Options
// ============================================================================

func TestGetVoiceLLMOptions_DefaultValues(t *testing.T) {
	// Act
	opts := getVoiceLLMOptions()

	// Assert
	assert.Equal(t, 0.7, opts.Temperature)
	assert.Equal(t, 500, opts.MaxTokens)
	assert.Equal(t, 0.9, opts.TopP)
	assert.False(t, opts.ThinkingEnabled)
	assert.Equal(t, 0, opts.MaxThinkingTokens)
}

func TestGetVoiceLLMOptions_ComparisonWithChat(t *testing.T) {
	// Arrange - Voice and chat options
	voiceOpts := getVoiceLLMOptions()

	// Create typical chat options
	chatOpts := LLMOptions{
		Temperature:       0.8,
		MaxTokens:         8192, // Much longer for chat
		TopP:              0.95,
		ThinkingEnabled:   true,
		MaxThinkingTokens: 8000,
	}

	// Assert - Voice should be more concise
	assert.Less(t, voiceOpts.MaxTokens, chatOpts.MaxTokens,
		"Voice should have shorter max tokens than chat")
	assert.False(t, voiceOpts.ThinkingEnabled, "Voice should not enable thinking")
	assert.True(t, chatOpts.ThinkingEnabled, "Chat can enable thinking")
}

// ============================================================================
// Tests for buildUserContext
// ============================================================================

func TestBuildUserContext_WithMockDB(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test that requires database setup")
	}

	// This test requires database setup - would normally use testdb
	// For now, we test the VoiceController interface

	// Arrange
	mockProvider := &mockVoiceAgentProvider{}
	vc := &VoiceController{
		agentProvider: mockProvider,
		sessions:      make(map[string]*VoiceSession),
	}

	// Assert
	assert.NotNil(t, vc.agentProvider)
	assert.NotNil(t, vc.sessions)
}

// ============================================================================
// Tests for VoiceController Methods
// ============================================================================

func TestNewVoiceController(t *testing.T) {
	// Arrange
	mockPool := (*pgxpool.Pool)(nil) // Mock pool
	cfg := &config.Config{}
	sttService := &WhisperService{}
	ttsService := &ElevenLabsService{}
	contextService := &TieredContextService{}
	agentProvider := &mockVoiceAgentProvider{}

	// Act
	vc := NewVoiceController(mockPool, cfg, sttService, ttsService, contextService, agentProvider)

	// Assert
	assert.NotNil(t, vc)
	assert.Equal(t, mockPool, vc.pool)
	assert.Equal(t, cfg, vc.cfg)
	assert.Equal(t, sttService, vc.STTService)
	assert.Equal(t, ttsService, vc.TTSService)
	assert.Equal(t, contextService, vc.contextService)
	assert.Equal(t, agentProvider, vc.agentProvider)
	assert.NotNil(t, vc.sessions)
	assert.Equal(t, 0, len(vc.sessions))
}

func TestGetOrCreateSession_NewSession(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test that requires database setup")
	}

	// Arrange
	vc := &VoiceController{
		sessions:   make(map[string]*VoiceSession),
		sessionsMu: sync.RWMutex{},
	}
	ctx := context.Background()
	sessionID := "session123"
	userID := "user123"

	// Act
	session, err := vc.GetOrCreateSession(ctx, sessionID, userID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, sessionID, session.SessionID)
	assert.Equal(t, userID, session.UserID)
	assert.NotNil(t, session.Messages)
	assert.NotNil(t, session.audioBuffer)
	assert.NotNil(t, session.cancel)
}

func TestGetOrCreateSession_ExistingSession(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test that requires synchronization")
	}

	// Arrange
	existingSession := &VoiceSession{
		SessionID: "session123",
		UserID:    "user123",
		Messages:  make([]Message, 0),
	}

	vc := &VoiceController{
		sessions:   make(map[string]*VoiceSession),
		sessionsMu: sync.RWMutex{},
	}
	vc.sessions["session123"] = existingSession

	// Act
	ctx := context.Background()
	session, err := vc.GetOrCreateSession(ctx, "session123", "user123")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, existingSession, session) // Should return same instance
	assert.Equal(t, "session123", session.SessionID)
}

// ============================================================================
// Tests for AccumulateStreamingResponse
// ============================================================================

func TestAccumulateStreamingResponse_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping streaming test")
	}

	// Arrange
	ctx := context.Background()
	eventsChan := make(chan streaming.StreamEvent, 5)
	errsChan := make(chan error, 1)

	// Simulate streaming events
	go func() {
		eventsChan <- streaming.StreamEvent{
			Type:    streaming.EventTypeToken,
			Content: "Hello ",
		}
		eventsChan <- streaming.StreamEvent{
			Type:    streaming.EventTypeToken,
			Content: "world",
		}
		eventsChan <- streaming.StreamEvent{
			Type: streaming.EventTypeDone,
		}
		close(eventsChan)
		close(errsChan)
	}()

	// Act
	response, err := accumulateStreamingResponse(ctx, eventsChan, errsChan)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "Hello world", response)
}

func TestAccumulateStreamingResponse_WithThinkingEvents(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping streaming test")
	}

	// Arrange
	ctx := context.Background()
	eventsChan := make(chan streaming.StreamEvent, 5)
	errsChan := make(chan error, 1)

	// Simulate streaming with thinking (thinking should be ignored)
	go func() {
		eventsChan <- streaming.StreamEvent{
			Type:    streaming.EventTypeThinkingStart,
			Content: "[thinking...]",
		}
		eventsChan <- streaming.StreamEvent{
			Type:    streaming.EventTypeThinkingChunk,
			Content: "...internal thought...",
		}
		eventsChan <- streaming.StreamEvent{
			Type:    streaming.EventTypeThinkingEnd,
			Content: "[/thinking]",
		}
		eventsChan <- streaming.StreamEvent{
			Type:    streaming.EventTypeToken,
			Content: "Final answer",
		}
		eventsChan <- streaming.StreamEvent{
			Type: streaming.EventTypeDone,
		}
		close(eventsChan)
		close(errsChan)
	}()

	// Act
	response, err := accumulateStreamingResponse(ctx, eventsChan, errsChan)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "Final answer", response)
	assert.NotContains(t, response, "[thinking...]")
}

func TestAccumulateStreamingResponse_Error(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping streaming test")
	}

	// Arrange
	ctx := context.Background()
	eventsChan := make(chan streaming.StreamEvent, 1)
	errsChan := make(chan error, 1)

	// Simulate error
	go func() {
		errsChan <- assert.AnError
		close(eventsChan)
		close(errsChan)
	}()

	// Act
	response, err := accumulateStreamingResponse(ctx, eventsChan, errsChan)

	// Assert
	assert.Error(t, err)
	assert.Empty(t, response)
}

func TestGenerateFallbackResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{
			name:     "hello greeting",
			input:    "hello",
			contains: "How can I help",
		},
		{
			name:     "hi greeting",
			input:    "hi",
			contains: "What can I do",
		},
		{
			name:     "help request",
			input:    "help",
			contains: "I'm OSA",
		},
		{
			name:     "thanks",
			input:    "thanks",
			contains: "You're welcome",
		},
		{
			name:     "bye",
			input:    "bye",
			contains: "Goodbye",
		},
		{
			name:     "unknown input",
			input:    "what is the weather",
			contains: "I understand you said",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			response := generateFallbackResponse(tt.input)

			// Assert
			assert.NotEmpty(t, response)
			assert.Contains(t, response, tt.contains)
		})
	}
}

func TestGenerateFallbackResponse_CaseInsensitive(t *testing.T) {
	// Arrange
	inputs := []string{"HELLO", "Hello", "HeLLo"}

	for _, input := range inputs {
		// Act
		response := generateFallbackResponse(input)

		// Assert
		assert.NotEmpty(t, response)
		assert.Contains(t, response, "How can I help")
	}
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
	messages []ChatMessage,
	tieredContext *TieredContext,
	llmOptions LLMOptions,
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

func BenchmarkGenerateFallbackResponse(b *testing.B) {
	inputs := []string{
		"hello",
		"what is the weather",
		"can you help me",
		"thanks",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generateFallbackResponse(inputs[i%len(inputs)])
	}
}

func BenchmarkVoiceUserContext_Creation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = &VoiceUserContext{
			UserID:         "user123",
			Username:       "testuser",
			DisplayName:    "Test User",
			WorkspaceID:    "ws123",
			WorkspaceName:  "Test Workspace",
			Role:           "admin",
			ExpertiseAreas: []string{"Go", "Python", "AI"},
		}
	}
}
