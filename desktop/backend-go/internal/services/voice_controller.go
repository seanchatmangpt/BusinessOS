package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/config"
	"github.com/rhl/businessos-backend/internal/streaming"
	voicev1 "github.com/rhl/businessos-backend/proto/voice/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// VoiceAgentProvider defines the interface for agent orchestration in voice
// This interface allows us to avoid importing the agents package (prevents import cycle)
type VoiceAgentProvider interface {
	// ExecuteVoiceAgent runs an agent with the given conversation history and returns streaming response
	ExecuteVoiceAgent(
		ctx context.Context,
		userID string,
		userName string,
		conversationID *uuid.UUID,
		messages []ChatMessage,
		tieredContext *TieredContext,
		llmOptions LLMOptions,
	) (<-chan streaming.StreamEvent, <-chan error)
}

// VoiceController orchestrates the complete voice pipeline:
// Audio In → STT (Whisper) → LLM (Agent V2) → TTS (ElevenLabs) → Audio Out
type VoiceController struct {
	voicev1.UnimplementedVoiceServiceServer
	pool           *pgxpool.Pool
	cfg            *config.Config
	STTService     *WhisperService    // Exported for Pure Go agent
	TTSService     *ElevenLabsService // Exported for Pure Go agent
	contextService *TieredContextService
	agentProvider  VoiceAgentProvider // Agent V2 system (interface to avoid import cycle)

	// Session management
	sessions   map[string]*VoiceSession
	sessionsMu sync.RWMutex
}

// VoiceSession represents an active voice conversation
type VoiceSession struct {
	SessionID   string
	UserID      string
	WorkspaceID string
	AgentRole   string
	State       voicev1.SessionState
	CreatedAt   time.Time
	UpdatedAt   time.Time

	// Audio buffering for STT
	audioBuffer []byte
	bufferMu    sync.Mutex

	// Conversation history (uses existing Message type from conversation_intelligence.go)
	Messages   []Message  // Exported for Pure Go agent
	MessagesMu sync.Mutex // Exported for Pure Go agent

	// User context (cached from DB)
	UserContext *VoiceUserContext
	contextMu   sync.Mutex

	// Cancel function for cleanup
	cancel context.CancelFunc
}

// VoiceUserContext represents loaded user context for voice sessions
type VoiceUserContext struct {
	UserID         string
	Username       string
	Email          string
	DisplayName    string
	WorkspaceID    string
	WorkspaceName  string
	Role           string
	Title          string
	Timezone       string
	OutputStyle    string
	ExpertiseAreas []string
}

// NewVoiceController creates a new voice controller
func NewVoiceController(
	pool *pgxpool.Pool,
	cfg *config.Config,
	sttService *WhisperService,
	ttsService *ElevenLabsService,
	contextService *TieredContextService,
	agentProvider VoiceAgentProvider, // Interface to avoid import cycle
) *VoiceController {
	return &VoiceController{
		pool:           pool,
		cfg:            cfg,
		STTService:     sttService,
		TTSService:     ttsService,
		contextService: contextService,
		agentProvider:  agentProvider,
		sessions:       make(map[string]*VoiceSession),
	}
}

// buildUserContext loads user context from database for voice sessions
func (vc *VoiceController) buildUserContext(
	ctx context.Context,
	userID string,
) (*VoiceUserContext, error) {
	// Query user basic info
	var userCtx VoiceUserContext
	userCtx.UserID = userID

	// Query basic user info from user table
	err := vc.pool.QueryRow(ctx, `
		SELECT
			COALESCE(username, 'User') as username,
			COALESCE(email, '') as email,
			COALESCE(name, username, 'User') as display_name
		FROM "user"
		WHERE id = $1
	`, userID).Scan(&userCtx.Username, &userCtx.Email, &userCtx.DisplayName)

	if err != nil {
		slog.Warn("[VoiceController] User not found, using defaults",
			"user_id", userID, "error", err)
		userCtx.Username = "User"
		userCtx.DisplayName = "User"
		// Continue with defaults - voice should still work
	}

	// Query workspace membership (get primary workspace)
	err = vc.pool.QueryRow(ctx, `
		SELECT
			wm.workspace_id,
			w.name as workspace_name,
			wm.role,
			COALESCE(uwp.title, '') as title,
			COALESCE(uwp.timezone, 'UTC') as timezone,
			COALESCE(uwp.preferred_output_style, 'concise') as output_style,
			COALESCE(uwp.expertise_areas, '{}') as expertise_areas
		FROM workspace_members wm
		JOIN workspaces w ON w.id = wm.workspace_id
		LEFT JOIN user_workspace_profiles uwp
			ON uwp.workspace_id = wm.workspace_id AND uwp.user_id = wm.user_id
		WHERE wm.user_id = $1
		  AND wm.status = 'active'
		ORDER BY wm.created_at ASC
		LIMIT 1
	`, userID).Scan(
		&userCtx.WorkspaceID,
		&userCtx.WorkspaceName,
		&userCtx.Role,
		&userCtx.Title,
		&userCtx.Timezone,
		&userCtx.OutputStyle,
		&userCtx.ExpertiseAreas,
	)

	if err != nil {
		slog.Debug("[VoiceController] No workspace found for user",
			"user_id", userID, "error", err)
		// User may not have workspace yet - OK for voice
	}

	slog.Info("[VoiceController] User context loaded",
		"user_id", userID,
		"username", userCtx.Username,
		"workspace", userCtx.WorkspaceName,
		"role", userCtx.Role)

	return &userCtx, nil
}

// ProcessVoice handles bidirectional audio streaming
func (vc *VoiceController) ProcessVoice(stream voicev1.VoiceService_ProcessVoiceServer) error {
	ctx := stream.Context()
	var session *VoiceSession

	// First frame should establish the session
	firstFrame, err := stream.Recv()
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "failed to receive first frame: %v", err)
	}

	// Get or create session
	session, err = vc.GetOrCreateSession(ctx, firstFrame.SessionId, firstFrame.UserId)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to create session: %v", err)
	}

	slog.Info("[VoiceController] Session started",
		"session_id", session.SessionID,
		"user_id", session.UserID)

	// Update session state
	session.State = voicev1.SessionState_LISTENING

	// Send initial state update
	if err := stream.Send(&voicev1.AudioResponse{
		Type:  voicev1.ResponseType_STATE_UPDATE,
		State: voicev1.SessionState_LISTENING,
	}); err != nil {
		return err
	}

	// Process audio frames in a loop
	for {
		select {
		case <-ctx.Done():
			slog.Info("[VoiceController] Session ended",
				"session_id", session.SessionID,
				"reason", ctx.Err())
			vc.cleanupSession(session.SessionID)
			return nil

		default:
			frame, err := stream.Recv()
			if err == io.EOF {
				// Client closed stream
				slog.Info("[VoiceController] Client closed stream", "session_id", session.SessionID)
				vc.cleanupSession(session.SessionID)
				return nil
			}
			if err != nil {
				return status.Errorf(codes.Internal, "receive error: %v", err)
			}

			// Process the audio frame
			if err := vc.processAudioFrame(ctx, session, frame, stream); err != nil {
				slog.Error("[VoiceController] Error processing frame",
					"session_id", session.SessionID,
					"error", err)
				// Send error to client but continue
				stream.Send(&voicev1.AudioResponse{
					Type:  voicev1.ResponseType_ERROR,
					Error: err.Error(),
				})
			}
		}
	}
}

// processAudioFrame handles a single audio frame
func (vc *VoiceController) processAudioFrame(
	ctx context.Context,
	session *VoiceSession,
	frame *voicev1.AudioFrame,
	stream voicev1.VoiceService_ProcessVoiceServer,
) error {
	// Only process user audio (not agent audio echoed back)
	if frame.Direction != "user" {
		return nil
	}

	// Buffer audio data
	session.bufferMu.Lock()
	session.audioBuffer = append(session.audioBuffer, frame.AudioData...)
	session.bufferMu.Unlock()

	// If this is a final frame, process the complete utterance
	if frame.IsFinal {
		return vc.processCompleteUtterance(ctx, session, stream)
	}

	return nil
}

// processCompleteUtterance handles a complete user speech segment
func (vc *VoiceController) processCompleteUtterance(
	ctx context.Context,
	session *VoiceSession,
	stream voicev1.VoiceService_ProcessVoiceServer,
) error {
	// Get buffered audio
	session.bufferMu.Lock()
	audioData := make([]byte, len(session.audioBuffer))
	copy(audioData, session.audioBuffer)
	session.audioBuffer = session.audioBuffer[:0] // Clear buffer
	session.bufferMu.Unlock()

	slog.Info("[VoiceController] Processing complete utterance",
		"session_id", session.SessionID,
		"audio_bytes", len(audioData))

	// 1. STT: Convert audio to text
	session.State = voicev1.SessionState_THINKING
	stream.Send(&voicev1.AudioResponse{
		Type:  voicev1.ResponseType_STATE_UPDATE,
		State: voicev1.SessionState_THINKING,
	})

	// Convert audio bytes to io.Reader for existing Whisper service
	audioReader := bytes.NewReader(audioData)
	transcriptionResult, err := vc.STTService.Transcribe(ctx, audioReader, "wav")
	if err != nil {
		return fmt.Errorf("STT failed: %w", err)
	}

	transcript := transcriptionResult.Text

	slog.Info("[VoiceController] User transcript",
		"session_id", session.SessionID,
		"text", transcript)

	// Send user transcript to client
	stream.Send(&voicev1.AudioResponse{
		Type: voicev1.ResponseType_TRANSCRIPT_USER,
		Text: transcript,
	})

	// Add to session history
	session.MessagesMu.Lock()
	session.Messages = append(session.Messages, Message{
		Role:      "user",
		Content:   transcript,
		Timestamp: time.Now(),
	})
	session.MessagesMu.Unlock()

	// 2. LLM: Get agent response using Agent V2 system
	agentResponse, err := vc.GetAgentResponse(ctx, session, transcript)
	if err != nil {
		return fmt.Errorf("LLM failed: %w", err)
	}

	slog.Info("[VoiceController] Agent response",
		"session_id", session.SessionID,
		"text", agentResponse)

	// Send agent transcript to client
	stream.Send(&voicev1.AudioResponse{
		Type: voicev1.ResponseType_TRANSCRIPT_AGENT,
		Text: agentResponse,
	})

	// Add to session history
	session.MessagesMu.Lock()
	session.Messages = append(session.Messages, Message{
		Role:      "agent",
		Content:   agentResponse,
		Timestamp: time.Now(),
	})
	session.MessagesMu.Unlock()

	// 3. TTS: Convert text to audio
	session.State = voicev1.SessionState_SPEAKING
	stream.Send(&voicev1.AudioResponse{
		Type:  voicev1.ResponseType_STATE_UPDATE,
		State: voicev1.SessionState_SPEAKING,
	})

	audioBytes, err := vc.TTSService.TextToSpeech(ctx, agentResponse)
	if err != nil {
		return fmt.Errorf("TTS failed: %w", err)
	}

	// Stream audio back in chunks (for low latency)
	chunkSize := 4096
	for i := 0; i < len(audioBytes); i += chunkSize {
		end := i + chunkSize
		if end > len(audioBytes) {
			end = len(audioBytes)
		}

		if err := stream.Send(&voicev1.AudioResponse{
			Type:      voicev1.ResponseType_AUDIO,
			AudioData: audioBytes[i:end],
			Sequence:  uint64(i / chunkSize),
		}); err != nil {
			return fmt.Errorf("failed to send audio chunk: %w", err)
		}
	}

	// Send DONE signal
	stream.Send(&voicev1.AudioResponse{
		Type: voicev1.ResponseType_DONE,
	})

	// Back to listening
	session.State = voicev1.SessionState_LISTENING
	stream.Send(&voicev1.AudioResponse{
		Type:  voicev1.ResponseType_STATE_UPDATE,
		State: voicev1.SessionState_LISTENING,
	})

	return nil
}

// GetAgentResponse gets LLM response using Agent V2 system
// Exported for Pure Go voice agent
func (vc *VoiceController) GetAgentResponse(
	ctx context.Context,
	session *VoiceSession,
	userMessage string,
) (string, error) {
	// If no agent provider, use fallback response
	if vc.agentProvider == nil {
		slog.Warn("[VoiceController] No agent provider configured, using fallback")
		return generateFallbackResponse(userMessage), nil
	}

	// Create timeout context for agent execution (30s max for voice)
	agentCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Convert session messages to ChatMessage format
	session.MessagesMu.Lock()
	chatMessages := make([]ChatMessage, len(session.Messages))
	for i, msg := range session.Messages {
		chatMessages[i] = ChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}
	session.MessagesMu.Unlock()

	// Build tiered context for the agent
	var tieredCtx *TieredContext
	if vc.contextService != nil {
		tieredReq := TieredContextRequest{
			UserID: session.UserID,
			// Add workspace/project context if available
			// For now, basic user context
		}
		tieredCtx, _ = vc.contextService.BuildTieredContext(agentCtx, tieredReq)
	}

	// Get voice-optimized LLM options
	llmOpts := getVoiceLLMOptions()

	// Load user context (cached in session)
	session.contextMu.Lock()
	if session.UserContext == nil {
		session.UserContext, _ = vc.buildUserContext(agentCtx, session.UserID)
		if session.UserContext == nil {
			// Fallback if loading failed
			session.UserContext = &VoiceUserContext{
				UserID:      session.UserID,
				Username:    "User",
				DisplayName: "User",
			}
		}
	}
	userCtx := session.UserContext
	session.contextMu.Unlock()

	// Use loaded user name
	userName := userCtx.DisplayName
	if userName == "" {
		userName = userCtx.Username
	}

	// Execute agent via provider interface (handles Agent V2 internally)
	// Note: conversationID is nil for voice sessions (they have their own session ID)
	events, errs := vc.agentProvider.ExecuteVoiceAgent(
		agentCtx,
		session.UserID,
		userName,
		nil, // No conversation ID for voice
		chatMessages,
		tieredCtx,
		llmOpts,
	)

	// Accumulate response from streaming events
	response, err := accumulateStreamingResponse(agentCtx, events, errs)
	if err != nil {
		slog.Error("[VoiceController] Agent execution failed",
			"error", err,
			"session_id", session.SessionID)
		// Return fallback response on error
		return generateFallbackResponse(userMessage), nil
	}

	slog.Info("[VoiceController] Agent response generated",
		"session_id", session.SessionID,
		"response_len", len(response))

	return response, nil
}

// getVoiceLLMOptions returns LLM options optimized for voice conversations
// Voice needs shorter, more concise responses than text chat
func getVoiceLLMOptions() LLMOptions {
	return LLMOptions{
		Temperature:       0.7, // Slightly creative but consistent
		MaxTokens:         500, // Short responses for voice (vs 8192 for chat)
		TopP:              0.9,
		ThinkingEnabled:   false, // No thinking tags in voice responses
		MaxThinkingTokens: 0,
	}
}

// accumulateStreamingResponse accumulates tokens from streaming events into final response
func accumulateStreamingResponse(
	ctx context.Context,
	events <-chan streaming.StreamEvent,
	errs <-chan error,
) (string, error) {
	var fullResponse strings.Builder

	for {
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("context cancelled: %w", ctx.Err())

		case err, ok := <-errs:
			if ok && err != nil {
				return "", fmt.Errorf("agent error: %w", err)
			}

		case event, ok := <-events:
			if !ok {
				// Stream ended successfully
				return fullResponse.String(), nil
			}

			// Process event types
			switch event.Type {
			case streaming.EventTypeToken:
				// Accumulate text tokens from Content field
				fullResponse.WriteString(event.Content)

			case streaming.EventTypeThinking,
				streaming.EventTypeThinkingStart,
				streaming.EventTypeThinkingChunk,
				streaming.EventTypeThinkingEnd:
				// Ignore thinking events for voice (voice responses should be direct)
				slog.Debug("[VoiceController] Thinking event (ignored for voice)")

			case streaming.EventTypeDone:
				// Stream completed
				return fullResponse.String(), nil

			case streaming.EventTypeError:
				// Error event - message in Content field
				if event.Content != "" {
					return "", fmt.Errorf("agent error event: %s", event.Content)
				}
				return "", fmt.Errorf("unknown error event")

			default:
				// Ignore other event types (tool calls, artifacts, etc.)
				slog.Debug("[VoiceController] Ignoring event type for voice",
					"type", event.Type)
			}
		}
	}
}

// generateFallbackResponse generates a simple fallback response when Agent V2 unavailable
func generateFallbackResponse(userMessage string) string {
	// Simple pattern matching for common queries
	responses := map[string]string{
		"hello":  "Hello! How can I help you today?",
		"hi":     "Hi there! What can I do for you?",
		"help":   "I'm OSA, your AI assistant. I can help you with various tasks. What would you like to know?",
		"thanks": "You're welcome! Is there anything else I can help with?",
		"bye":    "Goodbye! Feel free to ask me anything anytime.",
	}

	// Check for simple matches
	lowerMsg := strings.ToLower(userMessage)
	for key, response := range responses {
		if strings.Contains(lowerMsg, key) {
			return response
		}
	}

	// Default response
	return fmt.Sprintf("I understand you said: %s. I'm here to help! What would you like to know more about?", userMessage)
}

// GetSessionContext retrieves user context for voice session
func (vc *VoiceController) GetSessionContext(
	ctx context.Context,
	req *voicev1.SessionRequest,
) (*voicev1.SessionContext, error) {
	session, err := vc.GetOrCreateSession(ctx, req.SessionId, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get session: %v", err)
	}

	// Load user context (use cached if available)
	session.contextMu.Lock()
	if session.UserContext == nil {
		session.UserContext, _ = vc.buildUserContext(ctx, session.UserID)
	}
	userCtx := session.UserContext
	session.contextMu.Unlock()

	// Default fallback
	userName := "User"
	workspaceID := session.WorkspaceID

	// Use loaded context if available
	if userCtx != nil {
		userName = userCtx.DisplayName
		if userName == "" {
			userName = userCtx.Username
		}
		if userCtx.WorkspaceID != "" {
			workspaceID = userCtx.WorkspaceID
		}
	}

	return &voicev1.SessionContext{
		SessionId:   session.SessionID,
		UserId:      session.UserID,
		UserName:    userName,
		WorkspaceId: workspaceID,
		AgentRole:   session.AgentRole,
	}, nil
}

// UpdateSessionState updates session state
func (vc *VoiceController) UpdateSessionState(
	ctx context.Context,
	req *voicev1.SessionStateUpdate,
) (*voicev1.SessionStateResponse, error) {
	vc.sessionsMu.Lock()
	defer vc.sessionsMu.Unlock()

	session, exists := vc.sessions[req.SessionId]
	if !exists {
		return &voicev1.SessionStateResponse{
			Success: false,
			Message: "session not found",
		}, nil
	}

	session.State = req.State
	session.UpdatedAt = time.Now()

	slog.Info("[VoiceController] Session state updated",
		"session_id", req.SessionId,
		"new_state", req.State.String())

	return &voicev1.SessionStateResponse{
		Success: true,
		Message: "state updated",
	}, nil
}

// GetOrCreateSession gets an existing session or creates a new one
// Exported for Pure Go voice agent
func (vc *VoiceController) GetOrCreateSession(
	ctx context.Context,
	sessionID string,
	userID string,
) (*VoiceSession, error) {
	vc.sessionsMu.Lock()
	defer vc.sessionsMu.Unlock()

	if session, exists := vc.sessions[sessionID]; exists {
		return session, nil
	}

	// Create new session
	sessionCtx, cancel := context.WithCancel(ctx)
	session := &VoiceSession{
		SessionID:   sessionID,
		UserID:      userID,
		WorkspaceID: "", // TODO: fetch from user
		AgentRole:   "assistant",
		State:       voicev1.SessionState_IDLE,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		audioBuffer: make([]byte, 0, 1024*1024), // 1MB initial capacity
		Messages:    make([]Message, 0, 100),
		cancel:      cancel,
	}

	vc.sessions[sessionID] = session

	// Cleanup after 1 hour of inactivity
	go vc.sessionTimeout(sessionCtx, sessionID, 1*time.Hour)

	return session, nil
}

// sessionTimeout cleans up session after timeout
func (vc *VoiceController) sessionTimeout(ctx context.Context, sessionID string, timeout time.Duration) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return
	case <-timer.C:
		slog.Info("[VoiceController] Session timeout", "session_id", sessionID)
		vc.cleanupSession(sessionID)
	}
}

// cleanupSession removes session and frees resources
func (vc *VoiceController) cleanupSession(sessionID string) {
	vc.sessionsMu.Lock()
	defer vc.sessionsMu.Unlock()

	if session, exists := vc.sessions[sessionID]; exists {
		if session.cancel != nil {
			session.cancel()
		}
		delete(vc.sessions, sessionID)
		slog.Info("[VoiceController] Session cleaned up", "session_id", sessionID)
	}
}
