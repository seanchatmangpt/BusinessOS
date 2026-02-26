package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/integrations/google"
)

type OnboardingService struct {
	pool           *pgxpool.Pool
	aiService      *OnboardingAIService
	validator      *OnboardingValidator
	emailAnalyzer  *EmailAnalyzerService
	gmailService   *google.GmailService
	osaSyncService *OSASyncService
}

// OnboardingSession represents an onboarding session
type OnboardingSession struct {
	ID                 uuid.UUID              `json:"id"`
	UserID             string                 `json:"user_id"`
	Status             string                 `json:"status"`
	CurrentStep        string                 `json:"current_step"`
	StepsCompleted     []string               `json:"steps_completed"`
	ExtractedData      map[string]interface{} `json:"extracted_data"`
	LowConfidenceCount int                    `json:"low_confidence_count"`
	FallbackTriggered  bool                   `json:"fallback_triggered"`
	WorkspaceID        *uuid.UUID             `json:"workspace_id,omitempty"`
	StartedAt          time.Time              `json:"started_at"`
	CompletedAt        *time.Time             `json:"completed_at,omitempty"`
	ExpiresAt          time.Time              `json:"expires_at"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
}

// ConversationMessage represents a message in the conversation
type ConversationMessage struct {
	ID              uuid.UUID              `json:"id"`
	SessionID       uuid.UUID              `json:"session_id"`
	Role            string                 `json:"role"` // "user", "agent", "system"
	Content         string                 `json:"content"`
	ConfidenceScore *float64               `json:"confidence_score,omitempty"`
	ExtractedFields map[string]interface{} `json:"extracted_fields,omitempty"`
	QuestionType    *string                `json:"question_type,omitempty"`
	SequenceNumber  int                    `json:"sequence_number"`
	CreatedAt       time.Time              `json:"created_at"`
}

// ExtractedOnboardingData represents the data extracted from conversation
type ExtractedOnboardingData struct {
	WorkspaceName string   `json:"workspace_name,omitempty"`
	BusinessType  string   `json:"business_type,omitempty"` // Raw user input
	TeamSize      string   `json:"team_size,omitempty"`     // Raw user input
	Role          string   `json:"role,omitempty"`          // Raw user input
	Challenge     string   `json:"challenge,omitempty"`
	Integrations  []string `json:"integrations,omitempty"`
	// Normalized values for internal use (not stored in DB)
	NormalizedBusinessType string `json:"-"` // Used for logic only
	NormalizedTeamSize     string `json:"-"` // Used for logic only
	NormalizedRole         string `json:"-"` // Used for logic only
}

// OnboardingStatus check result
type OnboardingStatus struct {
	NeedsOnboarding bool               `json:"needs_onboarding"`
	HasSession      bool               `json:"has_session"`
	Session         *OnboardingSession `json:"session,omitempty"`
	WorkspaceCount  int                `json:"workspace_count"`
}

// SendMessageRequest for sending a message
type SendMessageRequest struct {
	Content string `json:"content" binding:"required"`
}

// SendMessageResponse from AI
type SendMessageResponse struct {
	Message                 ConversationMessage     `json:"message"`
	NextStep                string                  `json:"next_step"`
	IsComplete              bool                    `json:"is_complete"`
	ShouldShowFallback      bool                    `json:"should_show_fallback"`
	ExtractedData           ExtractedOnboardingData `json:"extracted_data"`
	RecommendedIntegrations []string                `json:"recommended_integrations,omitempty"`
}

// FallbackFormData for manual form submission
type FallbackFormData struct {
	// From Quick Info
	WorkspaceName string `json:"workspace_name" binding:"required"`
	BusinessType  string `json:"business_type" binding:"required"`
	TeamSize      string `json:"team_size"`
	Role          string `json:"role"`

	// From Fallback Form (5 questions)
	ToolsUsed     []string `json:"tools_used"`      // Q1: What tools do you use?
	MainFocus     string   `json:"main_focus"`      // Q2: Main work focus
	Challenge     string   `json:"challenge"`       // Q3: Biggest challenge
	WorkStyle     string   `json:"work_style"`      // Q4: How you work
	WhatWouldHelp []string `json:"what_would_help"` // Q5: What would help (max 3)

	// Optional integrations
	Integrations []string `json:"integrations"`
}

// CompleteOnboardingResponse after finishing
type CompleteOnboardingResponse struct {
	WorkspaceID   uuid.UUID `json:"workspace_id"`
	WorkspaceName string    `json:"workspace_name"`
	WorkspaceSlug string    `json:"workspace_slug"`
	RedirectURL   string    `json:"redirect_url"`
}

func NewOnboardingService(pool *pgxpool.Pool, aiService *OnboardingAIService, gmailService *google.GmailService, osaSyncService *OSASyncService) *OnboardingService {
	var emailAnalyzer *EmailAnalyzerService
	if gmailService != nil {
		emailAnalyzer = NewEmailAnalyzerService(pool, gmailService)
	}

	return &OnboardingService{
		pool:           pool,
		aiService:      aiService,
		validator:      NewOnboardingValidator(),
		emailAnalyzer:  emailAnalyzer,
		gmailService:   gmailService,
		osaSyncService: osaSyncService,
	}
}

// CheckOnboardingStatus checks if user needs onboarding
func (s *OnboardingService) CheckOnboardingStatus(ctx context.Context, userID string) (*OnboardingStatus, error) {
	status := &OnboardingStatus{}

	// Check workspace membership count
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM workspace_members 
		WHERE user_id = $1 AND status = 'active'
	`, userID).Scan(&status.WorkspaceCount)
	if err != nil {
		return nil, fmt.Errorf("check workspace count: %w", err)
	}

	// If user has workspaces, no onboarding needed
	if status.WorkspaceCount > 0 {
		status.NeedsOnboarding = false
		return status, nil
	}

	// Check for existing in-progress session
	session, err := s.GetResumeableSession(ctx, userID)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("check session: %w", err)
	}

	status.NeedsOnboarding = true
	if session != nil {
		status.HasSession = true
		status.Session = session
	}

	return status, nil
}

// CreateSession creates a new onboarding session
func (s *OnboardingService) CreateSession(ctx context.Context, userID string) (*OnboardingSession, error) {
	// First, expire/abandon any existing sessions for this user
	_, err := s.pool.Exec(ctx, `
		UPDATE onboarding_sessions 
		SET status = 'abandoned', updated_at = NOW()
		WHERE user_id = $1 AND status = 'in_progress'
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("abandon old sessions: %w", err)
	}

	// Create new session
	var session OnboardingSession
	err = s.pool.QueryRow(ctx, `
		INSERT INTO onboarding_sessions (user_id)
		VALUES ($1)
		RETURNING id, user_id, status, current_step, steps_completed, extracted_data, 
		          low_confidence_count, fallback_triggered, workspace_id, started_at, 
		          completed_at, expires_at, created_at, updated_at
	`, userID).Scan(
		&session.ID, &session.UserID, &session.Status, &session.CurrentStep,
		&session.StepsCompleted, &session.ExtractedData, &session.LowConfidenceCount,
		&session.FallbackTriggered, &session.WorkspaceID, &session.StartedAt,
		&session.CompletedAt, &session.ExpiresAt, &session.CreatedAt, &session.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	// Add initial agent message
	initialMessage := "Hi! I'm here to help set up your workspace. What's your company called?"
	_, err = s.AddMessage(ctx, session.ID, "agent", initialMessage, nil, onboardingStrPtr("company_name"))
	if err != nil {
		return nil, fmt.Errorf("add initial message: %w", err)
	}

	return &session, nil
}

// GetSession retrieves a session by ID
func (s *OnboardingService) GetSession(ctx context.Context, sessionID uuid.UUID) (*OnboardingSession, error) {
	var session OnboardingSession
	err := s.pool.QueryRow(ctx, `
		SELECT id, user_id, status, current_step, steps_completed, extracted_data, 
		       low_confidence_count, fallback_triggered, workspace_id, started_at, 
		       completed_at, expires_at, created_at, updated_at
		FROM onboarding_sessions
		WHERE id = $1
	`, sessionID).Scan(
		&session.ID, &session.UserID, &session.Status, &session.CurrentStep,
		&session.StepsCompleted, &session.ExtractedData, &session.LowConfidenceCount,
		&session.FallbackTriggered, &session.WorkspaceID, &session.StartedAt,
		&session.CompletedAt, &session.ExpiresAt, &session.CreatedAt, &session.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("session not found")
	}
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}
	return &session, nil
}

// GetResumeableSession gets an existing in-progress session for a user
func (s *OnboardingService) GetResumeableSession(ctx context.Context, userID string) (*OnboardingSession, error) {
	var session OnboardingSession
	err := s.pool.QueryRow(ctx, `
		SELECT id, user_id, status, current_step, steps_completed, extracted_data, 
		       low_confidence_count, fallback_triggered, workspace_id, started_at, 
		       completed_at, expires_at, created_at, updated_at
		FROM onboarding_sessions
		WHERE user_id = $1 AND status = 'in_progress' AND expires_at > NOW()
		ORDER BY created_at DESC
		LIMIT 1
	`, userID).Scan(
		&session.ID, &session.UserID, &session.Status, &session.CurrentStep,
		&session.StepsCompleted, &session.ExtractedData, &session.LowConfidenceCount,
		&session.FallbackTriggered, &session.WorkspaceID, &session.StartedAt,
		&session.CompletedAt, &session.ExpiresAt, &session.CreatedAt, &session.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get resumeable session: %w", err)
	}
	return &session, nil
}

// GetSessionWithHistory retrieves a session with its conversation history
func (s *OnboardingService) GetSessionWithHistory(ctx context.Context, sessionID uuid.UUID) (*OnboardingSession, []ConversationMessage, error) {
	session, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return nil, nil, err
	}

	messages, err := s.GetConversationHistory(ctx, sessionID, 0) // 0 = all messages
	if err != nil {
		return nil, nil, err
	}

	return session, messages, nil
}

// AbandonSession marks a session as abandoned
func (s *OnboardingService) AbandonSession(ctx context.Context, sessionID uuid.UUID, userID string) error {
	result, err := s.pool.Exec(ctx, `
		UPDATE onboarding_sessions
		SET status = 'abandoned', updated_at = NOW()
		WHERE id = $1 AND user_id = $2 AND status = 'in_progress'
	`, sessionID, userID)
	if err != nil {
		return fmt.Errorf("abandon session: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("session not found or already completed")
	}
	return nil
}

// AddMessage adds a message to the conversation
func (s *OnboardingService) AddMessage(ctx context.Context, sessionID uuid.UUID, role, content string, confidenceScore *float64, questionType *string) (*ConversationMessage, error) {
	// Get next sequence number
	var seqNum int
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(MAX(sequence_number), 0) + 1
		FROM onboarding_conversation_history
		WHERE session_id = $1
	`, sessionID).Scan(&seqNum)
	if err != nil {
		return nil, fmt.Errorf("get sequence number: %w", err)
	}

	var msg ConversationMessage
	err = s.pool.QueryRow(ctx, `
		INSERT INTO onboarding_conversation_history (session_id, role, content, confidence_score, question_type, sequence_number)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, session_id, role, content, confidence_score, extracted_fields, question_type, sequence_number, created_at
	`, sessionID, role, content, confidenceScore, questionType, seqNum).Scan(
		&msg.ID, &msg.SessionID, &msg.Role, &msg.Content, &msg.ConfidenceScore,
		&msg.ExtractedFields, &msg.QuestionType, &msg.SequenceNumber, &msg.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("add message: %w", err)
	}

	return &msg, nil
}

// GetConversationHistory retrieves conversation messages
func (s *OnboardingService) GetConversationHistory(ctx context.Context, sessionID uuid.UUID, limit int) ([]ConversationMessage, error) {
	query := `
		SELECT id, session_id, role, content, confidence_score, extracted_fields, question_type, sequence_number, created_at
		FROM onboarding_conversation_history
		WHERE session_id = $1
		ORDER BY sequence_number ASC
	`
	if limit > 0 {
		query = fmt.Sprintf(`
			SELECT * FROM (
				SELECT id, session_id, role, content, confidence_score, extracted_fields, question_type, sequence_number, created_at
				FROM onboarding_conversation_history
				WHERE session_id = $1
				ORDER BY sequence_number DESC
				LIMIT %d
			) sub ORDER BY sequence_number ASC
		`, limit)
	}

	rows, err := s.pool.Query(ctx, query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("get conversation history: %w", err)
	}
	defer rows.Close()

	var messages []ConversationMessage
	for rows.Next() {
		var msg ConversationMessage
		err := rows.Scan(
			&msg.ID, &msg.SessionID, &msg.Role, &msg.Content, &msg.ConfidenceScore,
			&msg.ExtractedFields, &msg.QuestionType, &msg.SequenceNumber, &msg.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// ProcessUserMessage processes a user message and returns AI response
func (s *OnboardingService) ProcessUserMessage(ctx context.Context, sessionID uuid.UUID, userID, content string) (*SendMessageResponse, error) {
	// Get session
	session, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if session.UserID != userID {
		return nil, fmt.Errorf("session does not belong to user")
	}

	// Check if session is still valid
	if session.Status != "in_progress" {
		return nil, fmt.Errorf("session is not in progress")
	}

	if time.Now().After(session.ExpiresAt) {
		return nil, fmt.Errorf("session has expired")
	}

	// Add user message
	_, err = s.AddMessage(ctx, sessionID, "user", content, nil, nil)
	if err != nil {
		return nil, err
	}

	// Process based on current step (chip selections vs chat)
	response, err := s.processStepResponse(ctx, session, content)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// processStepResponse handles the response based on current step
func (s *OnboardingService) processStepResponse(ctx context.Context, session *OnboardingSession, content string) (*SendMessageResponse, error) {
	response := &SendMessageResponse{}
	extractedData := ExtractedOnboardingData{}

	// Parse existing extracted data
	if session.ExtractedData != nil {
		dataBytes, _ := json.Marshal(session.ExtractedData)
		json.Unmarshal(dataBytes, &extractedData)
	}

	// Get conversation history for AI context
	history, err := s.GetConversationHistory(ctx, session.ID, 10)
	if err != nil {
		history = []ConversationMessage{} // Continue without history if error
	}

	// Convert to AI message format
	var aiHistory []OnboardingChatMessage
	for _, msg := range history {
		role := msg.Role
		if role == "agent" {
			role = "assistant"
		}
		aiHistory = append(aiHistory, OnboardingChatMessage{
			Role:    role,
			Content: msg.Content,
		})
	}

	// Call AI to process the message and get response
	aiResponse, aiErr := s.aiService.ProcessMessage(
		ctx,
		content,
		session.CurrentStep,
		structToMap(extractedData),
		aiHistory,
	)

	// Variables for step processing
	nextStep := session.CurrentStep
	agentMessage := ""
	var validationError *ValidationError

	// Process based on current step with validation
	switch session.CurrentStep {
	case "company_name":
		// Validate company name
		if err := s.validator.ValidateCompanyName(content); err != nil {
			validationError = err
			if aiErr == nil && aiResponse != nil && aiResponse.AgentMessage != "" {
				agentMessage = aiResponse.AgentMessage
			} else {
				agentMessage = fmt.Sprintf("Hmm, %s. Could you try a different name?", err.Message)
			}
		} else {
			extractedData.WorkspaceName = s.validator.SanitizeInput(content)
			nextStep = "business_type"
			if aiErr == nil && aiResponse != nil && aiResponse.AgentMessage != "" {
				agentMessage = aiResponse.AgentMessage
			} else {
				agentMessage = "Great name! What kind of work does " + extractedData.WorkspaceName + " do?"
			}
		}

	case "business_type":
		// Normalize for validation but store raw input
		normalized := s.validator.NormalizeBusinessType(content)
		if err := s.validator.ValidateBusinessType(normalized); err != nil {
			validationError = err
			if aiErr == nil && aiResponse != nil && aiResponse.AgentMessage != "" {
				agentMessage = aiResponse.AgentMessage
			} else {
				agentMessage = "I didn't quite catch that. Are you an agency, startup, freelancer, consultant, or something else?"
			}
		} else {
			extractedData.BusinessType = s.validator.SanitizeInput(content) // Store sanitized input
			extractedData.NormalizedBusinessType = normalized               // Keep normalized for logic
			if normalized == "freelance" {
				extractedData.TeamSize = "solo (freelancer)"
				nextStep = "role"
				if aiErr == nil && aiResponse != nil && aiResponse.AgentMessage != "" {
					agentMessage = aiResponse.AgentMessage
				} else {
					agentMessage = "Solo power! What's your role - founder, consultant, or something else?"
				}
			} else {
				nextStep = "team_size"
				if aiErr == nil && aiResponse != nil && aiResponse.AgentMessage != "" {
					agentMessage = aiResponse.AgentMessage
				} else {
					agentMessage = "Nice! How big is your team?"
				}
			}
		}

	case "team_size":
		// Normalize for validation but store raw input
		normalized := s.validator.NormalizeTeamSize(content)
		if err := s.validator.ValidateTeamSize(normalized); err != nil {
			validationError = err
			if aiErr == nil && aiResponse != nil && aiResponse.AgentMessage != "" {
				agentMessage = aiResponse.AgentMessage
			} else {
				agentMessage = "Could you tell me roughly how many people are on your team? Just you, 2-5, 6-10, 11-50, or more?"
			}
		} else {
			extractedData.TeamSize = s.validator.SanitizeInput(content) // Store sanitized input
			extractedData.NormalizedTeamSize = normalized               // Keep normalized for logic
			nextStep = "role"
			if aiErr == nil && aiResponse != nil && aiResponse.AgentMessage != "" {
				agentMessage = aiResponse.AgentMessage
			} else {
				agentMessage = "Got it! And what's your role in the team?"
			}
		}

	case "role":
		// Store sanitized input and normalize for internal use
		extractedData.Role = s.validator.SanitizeInput(content)           // Store sanitized input
		extractedData.NormalizedRole = s.validator.NormalizeRole(content) // Keep normalized for logic
		nextStep = "challenge"
		if aiErr == nil && aiResponse != nil && aiResponse.AgentMessage != "" {
			agentMessage = aiResponse.AgentMessage
		} else {
			agentMessage = "Awesome! What's the biggest challenge you're hoping to solve with Business OS?"
		}

	case "challenge":
		// Validate challenge
		if err := s.validator.ValidateChallenge(content); err != nil {
			validationError = err
			if aiErr == nil && aiResponse != nil && aiResponse.AgentMessage != "" {
				agentMessage = aiResponse.AgentMessage
			} else {
				agentMessage = "Could you tell me a bit more about the challenges you're facing? A sentence or two would help me understand."
			}
		} else {
			extractedData.Challenge = s.validator.SanitizeInput(content)
			nextStep = "integrations"
			response.RecommendedIntegrations = ComputeRecommendations(extractedData.Challenge, extractedData.BusinessType)
			if aiErr == nil && aiResponse != nil && aiResponse.AgentMessage != "" {
				agentMessage = aiResponse.AgentMessage
			} else {
				agentMessage = "I hear you! Based on what you've shared, I've got some tool recommendations. Let's connect your favorites!"
			}
			response.IsComplete = false
		}

	case "integrations":
		nextStep = "complete"
		response.IsComplete = true
		agentMessage = "Perfect! Your workspace is ready. Let's get started!"
	}

	// If validation failed, don't advance the step
	if validationError != nil {
		response.Message = ConversationMessage{
			Role:    "agent",
			Content: agentMessage,
		}
		response.NextStep = session.CurrentStep // Stay on current step
		response.ExtractedData = extractedData

		// Add agent message for the retry prompt
		_, _ = s.AddMessage(ctx, session.ID, "agent", agentMessage, nil, nil)

		return response, nil
	}

	// Update session
	stepsCompleted := append(session.StepsCompleted, session.CurrentStep)
	extractedDataMap := structToMap(extractedData)

	_, err = s.pool.Exec(ctx, `
		UPDATE onboarding_sessions
		SET current_step = $1, steps_completed = $2, extracted_data = $3, updated_at = NOW()
		WHERE id = $4
	`, nextStep, stepsCompleted, extractedDataMap, session.ID)
	if err != nil {
		return nil, fmt.Errorf("update session: %w", err)
	}

	// Add agent message
	if agentMessage != "" {
		msg, err := s.AddMessage(ctx, session.ID, "agent", agentMessage, nil, &nextStep)
		if err != nil {
			return nil, err
		}
		response.Message = *msg
	}

	response.NextStep = nextStep
	response.ExtractedData = extractedData

	return response, nil
}

// CompleteOnboarding completes the onboarding and creates the workspace
func (s *OnboardingService) CompleteOnboarding(ctx context.Context, sessionID uuid.UUID, userID string, integrations []string) (*CompleteOnboardingResponse, error) {
	session, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if session.UserID != userID {
		return nil, fmt.Errorf("session does not belong to user")
	}

	// Parse extracted data
	var extractedData ExtractedOnboardingData
	if session.ExtractedData != nil {
		dataBytes, _ := json.Marshal(session.ExtractedData)
		json.Unmarshal(dataBytes, &extractedData)
	}

	// Add integrations
	extractedData.Integrations = integrations

	// Apply defaults for missing required fields (allows "skip" flow)
	if extractedData.WorkspaceName == "" {
		extractedData.WorkspaceName = "My Workspace"
	}
	if extractedData.BusinessType == "" {
		extractedData.BusinessType = "other"
	}
	if extractedData.TeamSize == "" {
		extractedData.TeamSize = "solo"
	}

	// Validate integrations if provided
	if len(integrations) > 0 {
		if err := s.validator.ValidateIntegrations(integrations); err != nil {
			return nil, fmt.Errorf("invalid integrations: %s", err.Message)
		}
	}

	// Start transaction
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Create workspace
	workspaceName := extractedData.WorkspaceName
	if workspaceName == "" {
		workspaceName = "My Workspace"
	}
	slug := generateSlugFromName(workspaceName)

	var workspace struct {
		ID   uuid.UUID
		Name string
		Slug string
	}
	err = tx.QueryRow(ctx, `
		INSERT INTO workspaces (name, slug, owner_id, onboarding_completed_at, onboarding_data)
		VALUES ($1, $2, $3, NOW(), $4)
		RETURNING id, name, slug
	`, workspaceName, slug, userID, structToMap(extractedData)).Scan(&workspace.ID, &workspace.Name, &workspace.Slug)
	if err != nil {
		return nil, fmt.Errorf("create workspace: %w", err)
	}

	// Seed default roles
	_, err = tx.Exec(ctx, "SELECT seed_default_workspace_roles($1)", workspace.ID)
	if err != nil {
		// Try without the function if it doesn't exist
		_, err = tx.Exec(ctx, `
			INSERT INTO workspace_roles (workspace_id, name, display_name, is_system, hierarchy_level, permissions)
			VALUES 
				($1, 'owner', 'Owner', true, 100, '{"all": true}'::jsonb),
				($1, 'admin', 'Admin', true, 80, '{"manage_members": true, "manage_settings": true}'::jsonb),
				($1, 'member', 'Member', true, 50, '{"read": true, "write": true}'::jsonb)
			ON CONFLICT DO NOTHING
		`, workspace.ID)
		if err != nil {
			return nil, fmt.Errorf("seed roles: %w", err)
		}
	}

	// Add owner as first member
	_, err = tx.Exec(ctx, `
		INSERT INTO workspace_members (workspace_id, user_id, role_name, status, joined_at)
		VALUES ($1, $2, 'owner', 'active', NOW())
	`, workspace.ID, userID)
	if err != nil {
		return nil, fmt.Errorf("add owner: %w", err)
	}

	// Create onboarding profile
	recommendations := ComputeRecommendations(extractedData.Challenge, extractedData.BusinessType)
	_, err = tx.Exec(ctx, `
		INSERT INTO workspace_onboarding_profiles 
			(workspace_id, business_type, team_size, owner_role, main_challenge, recommended_integrations, onboarding_session_id, onboarding_method)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'conversational')
	`, workspace.ID, extractedData.BusinessType, extractedData.TeamSize, extractedData.Role, extractedData.Challenge, recommendations, session.ID)
	if err != nil {
		return nil, fmt.Errorf("create onboarding profile: %w", err)
	}

	// Commit the transaction first so workspace exists
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	// Sync user and workspace to OSA if OSA sync service is available
	if s.osaSyncService != nil {
		// Convert userID string to UUID
		userUUID, err := uuid.Parse(userID)
		if err != nil {
			slog.Warn("Failed to parse user ID for OSA sync",
				"user_id", userID,
				"error", err,
			)
		} else {
			// Sync user to OSA (async)
			go func() {
				bgCtx := context.Background()
				if err := s.osaSyncService.SyncUser(bgCtx, userUUID); err != nil {
					slog.Warn("Failed to sync user to OSA",
						"user_id", userID,
						"error", err,
					)
				} else {
					slog.Info("User synced to OSA successfully",
						"user_id", userID,
					)
				}
			}()

			// Sync workspace to OSA (async)
			go func() {
				bgCtx := context.Background()
				if err := s.osaSyncService.SyncWorkspace(bgCtx, workspace.ID); err != nil {
					slog.Warn("Failed to sync workspace to OSA",
						"workspace_id", workspace.ID,
						"error", err,
					)
				} else {
					slog.Info("Workspace synced to OSA successfully",
						"workspace_id", workspace.ID,
						"name", workspace.Name,
					)
				}
			}()
		}
	} else {
		slog.Debug("OSA sync service not available, skipping OSA sync",
			"workspace_id", workspace.ID,
		)
	}

	// Trigger post-onboarding app generation (fire-and-forget)
	go func() {
		postOnboardingService := NewPostOnboardingService(s.pool, slog.Default())
		if err := postOnboardingService.QueueAppsForWorkspace(context.Background(), workspace.ID); err != nil {
			slog.Warn("Failed to queue post-onboarding apps",
				"workspace_id", workspace.ID,
				"error", err,
			)
		} else {
			slog.Info("Successfully queued post-onboarding apps",
				"workspace_id", workspace.ID,
			)
		}
	}()

	// Transform AI analysis to workspace profile if analysis exists
	if err := s.transformAIAnalysisToWorkspaceProfile(ctx, workspace.ID, userID); err != nil {
		// Log warning but don't fail - workspace was created successfully
		slog.Default().Warn("Failed to transform AI analysis to workspace profile",
			"workspace_id", workspace.ID,
			"error", err,
		)
	}

	// Analyze and save email metadata if Gmail is connected
	if s.emailAnalyzer != nil && s.gmailService != nil {
		slog.Info("Analyzing user emails for onboarding insights",
			"user_id", userID,
			"session_id", session.ID,
		)

		// Analyze and save recent emails (async to not block workspace creation)
		go func() {
			bgCtx := context.Background()
			metadata, err := s.emailAnalyzer.AnalyzeAndSaveRecentEmails(bgCtx, userID, session.ID, 50)
			if err != nil {
				slog.Warn("Failed to analyze emails during onboarding",
					"user_id", userID,
					"session_id", session.ID,
					"error", err,
				)
			} else {
				slog.Info("Email metadata saved successfully",
					"user_id", userID,
					"session_id", session.ID,
					"emails_analyzed", metadata.TotalEmails,
					"tools_detected", len(metadata.DetectedTools),
				)
			}
		}()
	}

	// Update session as completed
	_, err = s.pool.Exec(ctx, `
		UPDATE onboarding_sessions
		SET status = 'completed', workspace_id = $1, completed_at = NOW(),
		    extracted_data = $2, current_step = 'complete', updated_at = NOW()
		WHERE id = $3
	`, workspace.ID, structToMap(extractedData), session.ID)
	if err != nil {
		// Log warning but don't fail - workspace was created successfully
		slog.Default().Warn("Failed to update session status", "error", err)
	}

	return &CompleteOnboardingResponse{
		WorkspaceID:   workspace.ID,
		WorkspaceName: workspace.Name,
		WorkspaceSlug: workspace.Slug,
		RedirectURL:   "/window",
	}, nil
}

// SubmitFallbackForm handles fallback form submission
func (s *OnboardingService) SubmitFallbackForm(ctx context.Context, sessionID uuid.UUID, userID string, data *FallbackFormData) (*CompleteOnboardingResponse, error) {
	// Validate the fallback form data
	if validationErrors := s.validator.ValidateFallbackForm(data); validationErrors.HasErrors() {
		return nil, fmt.Errorf("validation failed: %s", validationErrors.Error())
	}

	session, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if session.UserID != userID {
		return nil, fmt.Errorf("session does not belong to user")
	}

	// Build extracted_data from fallback form
	extractedData := map[string]interface{}{
		"workspace_name": data.WorkspaceName,
		"business_type":  data.BusinessType,
		"team_size":      data.TeamSize,
		"role":           data.Role,
		"challenge":      data.Challenge,
		"integrations":   data.Integrations,
		// NEW: Fallback form fields
		"tools_used":      data.ToolsUsed,
		"main_focus":      data.MainFocus,
		"work_style":      data.WorkStyle,
		"what_would_help": data.WhatWouldHelp,
	}

	// Update session with fallback flag and extracted data
	_, err = s.pool.Exec(ctx, `
		UPDATE onboarding_sessions
		SET fallback_triggered = true,
		    extracted_data = $1,
		    updated_at = NOW()
		WHERE id = $2
	`, extractedData, sessionID)
	if err != nil {
		return nil, fmt.Errorf("update session with fallback data: %w", err)
	}

	// Complete onboarding with form data
	return s.CompleteOnboarding(ctx, sessionID, userID, data.Integrations)
}

func onboardingStrPtr(s string) *string {
	return &s
}

func structToMap(v interface{}) map[string]interface{} {
	data, _ := json.Marshal(v)
	var result map[string]interface{}
	json.Unmarshal(data, &result)
	return result
}

// GetRecommendations returns integration recommendations based on session data
func (s *OnboardingService) GetRecommendations(ctx context.Context, sessionID uuid.UUID, userID string) ([]string, error) {
	session, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if session.UserID != userID {
		return nil, fmt.Errorf("session does not belong to user")
	}

	// Parse extracted data
	var extractedData ExtractedOnboardingData
	if session.ExtractedData != nil {
		dataBytes, _ := json.Marshal(session.ExtractedData)
		json.Unmarshal(dataBytes, &extractedData)
	}

	return ComputeRecommendations(extractedData.Challenge, extractedData.BusinessType), nil
}

func generateSlugFromName(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	// Remove special characters
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	slug = result.String()
	// Add random suffix for uniqueness
	slug = fmt.Sprintf("%s-%s", slug, uuid.New().String()[:8])
	return slug
}

// ComputeRecommendations returns integration recommendations based on challenge and business type.
// This is the single source of truth for recommendation logic.
func ComputeRecommendations(challenge, businessType string) []string {
	challengeLower := strings.ToLower(challenge)

	if strings.Contains(challengeLower, "organiz") || strings.Contains(challengeLower, "chaos") || strings.Contains(challengeLower, "mess") {
		return []string{"notion", "google", "linear"}
	}
	if strings.Contains(challengeLower, "scale") || strings.Contains(challengeLower, "grow") || strings.Contains(challengeLower, "automat") {
		return []string{"linear", "slack", "airtable"}
	}
	if strings.Contains(challengeLower, "client") || strings.Contains(challengeLower, "customer") || strings.Contains(challengeLower, "crm") {
		return []string{"hubspot", "slack", "google"}
	}
	if strings.Contains(challengeLower, "team") || strings.Contains(challengeLower, "collaborat") || strings.Contains(challengeLower, "communic") {
		return []string{"slack", "notion", "linear"}
	}
	if strings.Contains(challengeLower, "time") || strings.Contains(challengeLower, "busy") || strings.Contains(challengeLower, "meeting") {
		return []string{"google", "fathom", "slack"}
	}

	// Default by business type
	switch businessType {
	case "agency", "consulting":
		return []string{"hubspot", "slack", "notion"}
	case "startup":
		return []string{"linear", "slack", "notion"}
	case "freelance":
		return []string{"google", "notion", "fathom"}
	default:
		return []string{"google", "slack", "notion"}
	}
}

// transformAIAnalysisToWorkspaceProfile extracts structured data from AI analysis
// and populates/updates the workspace_onboarding_profiles table
func (s *OnboardingService) transformAIAnalysisToWorkspaceProfile(
	ctx context.Context,
	workspaceID uuid.UUID,
	userID string,
) error {
	// Fetch onboarding analysis for this user/workspace
	var analysis struct {
		ID             uuid.UUID
		ProfileSummary string
		Insights       []byte
		ToolsUsed      []byte
	}

	err := s.pool.QueryRow(ctx, `
		SELECT id, profile_summary, insights, tools_used
		FROM onboarding_user_analysis
		WHERE user_id = $1 AND workspace_id = $2 AND status = 'completed'
		ORDER BY created_at DESC
		LIMIT 1
	`, userID, workspaceID).Scan(
		&analysis.ID,
		&analysis.ProfileSummary,
		&analysis.Insights,
		&analysis.ToolsUsed,
	)

	if err == pgx.ErrNoRows {
		// No AI analysis found, skip transformation
		slog.Default().Info("No AI analysis found for workspace, skipping transformation",
			"workspace_id", workspaceID,
			"user_id", userID,
		)
		return nil
	}
	if err != nil {
		return fmt.Errorf("fetch analysis: %w", err)
	}

	// Extract structured fields from AI-generated profile summary
	businessType := extractBusinessTypeFromSummary(analysis.ProfileSummary)
	teamSize := extractTeamSizeFromSummary(analysis.ProfileSummary)
	ownerRole := extractOwnerRoleFromSummary(analysis.ProfileSummary)
	mainChallenge := extractMainChallengeFromInsights(analysis.Insights)

	// Extract recommended integrations from tools used
	var recommendedIntegrations []string
	if err := json.Unmarshal(analysis.ToolsUsed, &recommendedIntegrations); err != nil {
		slog.Default().Warn("failed to unmarshal tools_used", "error", err)
		recommendedIntegrations = []string{}
	}

	// Convert to JSONB for database
	integrationsJSON, err := json.Marshal(recommendedIntegrations)
	if err != nil {
		integrationsJSON = nil
	}

	// Check if profile already exists
	var existingID uuid.UUID
	err = s.pool.QueryRow(ctx, `
		SELECT id FROM workspace_onboarding_profiles WHERE workspace_id = $1
	`, workspaceID).Scan(&existingID)

	if err == nil {
		// Update existing profile
		_, err = s.pool.Exec(ctx, `
			UPDATE workspace_onboarding_profiles
			SET business_type = $1,
			    team_size = $2,
			    owner_role = $3,
			    main_challenge = $4,
			    recommended_integrations = $5,
			    updated_at = NOW()
			WHERE workspace_id = $6
		`, businessType, teamSize, ownerRole, mainChallenge, integrationsJSON, workspaceID)

		if err != nil {
			return fmt.Errorf("update profile: %w", err)
		}

		slog.Default().Info("workspace profile updated from AI analysis",
			"workspace_id", workspaceID,
			"business_type", businessType,
			"team_size", teamSize,
		)
		return nil
	}

	// Profile doesn't exist, but it should have been created in CompleteOnboarding
	// Log warning and skip
	if err == pgx.ErrNoRows {
		slog.Default().Warn("workspace_onboarding_profiles entry not found, but should exist",
			"workspace_id", workspaceID,
		)
		return nil
	}

	return fmt.Errorf("check existing profile: %w", err)
}

// Helper functions for extraction from AI analysis

func extractBusinessTypeFromSummary(summary string) string {
	// Simple keyword matching (can be enhanced with NLP)
	summaryLower := strings.ToLower(summary)

	if strings.Contains(summaryLower, "agency") || strings.Contains(summaryLower, "consulting") {
		return "agency"
	}
	if strings.Contains(summaryLower, "startup") || strings.Contains(summaryLower, "tech") {
		return "startup"
	}
	if strings.Contains(summaryLower, "enterprise") || strings.Contains(summaryLower, "corporation") {
		return "enterprise"
	}
	if strings.Contains(summaryLower, "freelance") || strings.Contains(summaryLower, "contractor") {
		return "freelancer"
	}
	if strings.Contains(summaryLower, "ecommerce") || strings.Contains(summaryLower, "retail") {
		return "ecommerce"
	}

	return "other" // default
}

func extractTeamSizeFromSummary(summary string) string {
	summaryLower := strings.ToLower(summary)

	if strings.Contains(summaryLower, "solo") || strings.Contains(summaryLower, "individual") {
		return "1"
	}
	if strings.Contains(summaryLower, "small team") || strings.Contains(summaryLower, "2-10") {
		return "2-10"
	}
	if strings.Contains(summaryLower, "medium") || strings.Contains(summaryLower, "11-50") {
		return "11-50"
	}
	if strings.Contains(summaryLower, "large") || strings.Contains(summaryLower, "50+") {
		return "51+"
	}

	return "2-10" // default to small team
}

func extractOwnerRoleFromSummary(summary string) string {
	summaryLower := strings.ToLower(summary)

	if strings.Contains(summaryLower, "founder") || strings.Contains(summaryLower, "ceo") {
		return "founder"
	}
	if strings.Contains(summaryLower, "manager") || strings.Contains(summaryLower, "director") {
		return "manager"
	}
	if strings.Contains(summaryLower, "developer") || strings.Contains(summaryLower, "engineer") {
		return "developer"
	}
	if strings.Contains(summaryLower, "designer") {
		return "designer"
	}

	return "other"
}

func extractMainChallengeFromInsights(insightsJSON []byte) string {
	var insights []string
	if err := json.Unmarshal(insightsJSON, &insights); err != nil || len(insights) == 0 {
		return "productivity"
	}

	// Take first insight as main challenge
	firstInsight := strings.ToLower(insights[0])

	if strings.Contains(firstInsight, "time") || strings.Contains(firstInsight, "productivity") {
		return "productivity"
	}
	if strings.Contains(firstInsight, "team") || strings.Contains(firstInsight, "collaboration") {
		return "collaboration"
	}
	if strings.Contains(firstInsight, "communication") {
		return "communication"
	}
	if strings.Contains(firstInsight, "organization") || strings.Contains(firstInsight, "management") {
		return "organization"
	}

	return "productivity" // default
}

// OnboardingProfileData represents the user's onboarding profile from workspace_onboarding_profiles
type OnboardingProfileData struct {
	BusinessType            string   `json:"business_type"`
	TeamSize                string   `json:"team_size"`
	OwnerRole               string   `json:"owner_role"`
	MainChallenge           string   `json:"main_challenge"`
	RecommendedIntegrations []string `json:"recommended_integrations"`
	// Optional AI analysis data from onboarding_user_analysis table
	ProfileSummary string   `json:"profile_summary,omitempty"`
	Insights       []string `json:"insights,omitempty"`
	ToolsUsed      []string `json:"tools_used,omitempty"`
}

// GetUserProfile retrieves the user's most recent onboarding profile from workspace_onboarding_profiles
// This is used to inject personalized context into agent prompts
func (s *OnboardingService) GetUserProfile(ctx context.Context, userID string) (*OnboardingProfileData, error) {
	// Query the most recent profile for this user
	// Join with workspaces to get the workspace_id
	var profile struct {
		BusinessType            string
		TeamSize                string
		OwnerRole               string
		MainChallenge           string
		RecommendedIntegrations []byte
		ProfileSummary          *string
		Insights                []byte
		ToolsUsed               []byte
	}

	// First, check if there's AI analysis data available
	analysisQuery := `
		SELECT
			profile_summary,
			insights,
			tools_used
		FROM onboarding_user_analysis
		WHERE user_id = $1 AND status = 'completed'
		ORDER BY created_at DESC
		LIMIT 1
	`

	var analysisSummary *string
	var analysisInsights []byte
	var analysisTools []byte

	err := s.pool.QueryRow(ctx, analysisQuery, userID).Scan(
		&analysisSummary,
		&analysisInsights,
		&analysisTools,
	)

	// AI analysis is optional, so we continue even if not found
	if err != nil && err != pgx.ErrNoRows {
		slog.Default().Warn("Failed to fetch AI analysis for user", "user_id", userID, "error", err)
	}

	// Now fetch the profile from workspace_onboarding_profiles
	// Join with workspace_members to find the user's workspace
	query := `
		SELECT
			p.business_type,
			p.team_size,
			p.owner_role,
			p.main_challenge,
			p.recommended_integrations
		FROM workspace_onboarding_profiles p
		INNER JOIN workspace_members wm ON wm.workspace_id = p.workspace_id
		WHERE wm.user_id = $1 AND wm.status = 'active'
		ORDER BY p.created_at DESC
		LIMIT 1
	`

	err = s.pool.QueryRow(ctx, query, userID).Scan(
		&profile.BusinessType,
		&profile.TeamSize,
		&profile.OwnerRole,
		&profile.MainChallenge,
		&profile.RecommendedIntegrations,
	)

	if err == pgx.ErrNoRows {
		// No profile found - user hasn't completed onboarding
		slog.Default().Debug("No onboarding profile found for user", "user_id", userID)
		return nil, fmt.Errorf("no onboarding profile found")
	}
	if err != nil {
		return nil, fmt.Errorf("fetch profile: %w", err)
	}

	// Build result
	result := &OnboardingProfileData{
		BusinessType:  profile.BusinessType,
		TeamSize:      profile.TeamSize,
		OwnerRole:     profile.OwnerRole,
		MainChallenge: profile.MainChallenge,
	}

	// Parse recommended integrations
	if len(profile.RecommendedIntegrations) > 0 {
		if err := json.Unmarshal(profile.RecommendedIntegrations, &result.RecommendedIntegrations); err != nil {
			slog.Default().Warn("Failed to unmarshal recommended integrations", "error", err)
			result.RecommendedIntegrations = []string{}
		}
	}

	// Add AI analysis data if available
	if analysisSummary != nil {
		result.ProfileSummary = *analysisSummary
	}

	if len(analysisInsights) > 0 {
		if err := json.Unmarshal(analysisInsights, &result.Insights); err != nil {
			slog.Default().Warn("Failed to unmarshal insights", "error", err)
		}
	}

	if len(analysisTools) > 0 {
		if err := json.Unmarshal(analysisTools, &result.ToolsUsed); err != nil {
			slog.Default().Warn("Failed to unmarshal tools_used", "error", err)
		}
	}

	return result, nil
}

// BuildProfilePrefix constructs a prompt prefix from the user's profile
// This is injected at the start of agent system prompts for personalization
func BuildProfilePrefix(profile *OnboardingProfileData) string {
	if profile == nil {
		return ""
	}

	var builder strings.Builder
	builder.WriteString("## USER PROFILE CONTEXT\n\n")
	builder.WriteString("You are assisting a user with the following profile:\n\n")

	// Business context
	builder.WriteString(fmt.Sprintf("**Business Type:** %s\n", profile.BusinessType))
	builder.WriteString(fmt.Sprintf("**Team Size:** %s\n", profile.TeamSize))
	builder.WriteString(fmt.Sprintf("**Role:** %s\n", profile.OwnerRole))
	builder.WriteString(fmt.Sprintf("**Main Challenge:** %s\n", profile.MainChallenge))

	// Tools/Integrations context
	if len(profile.RecommendedIntegrations) > 0 {
		builder.WriteString(fmt.Sprintf("\n**Preferred Tools:** %s\n", strings.Join(profile.RecommendedIntegrations, ", ")))
	}

	// AI insights (if available)
	if len(profile.Insights) > 0 {
		builder.WriteString("\n**Key Insights:**\n")
		for _, insight := range profile.Insights {
			builder.WriteString(fmt.Sprintf("- %s\n", insight))
		}
	}

	// Profile summary (if available)
	if profile.ProfileSummary != "" {
		builder.WriteString("\n**Profile Summary:**\n")
		builder.WriteString(profile.ProfileSummary)
		builder.WriteString("\n")
	}

	builder.WriteString("\n**IMPORTANT:** Use this profile information to personalize your responses and recommendations. Tailor your language, examples, and suggestions to match the user's business context, team size, and challenges.\n")
	builder.WriteString("\n---\n\n")

	return builder.String()
}
