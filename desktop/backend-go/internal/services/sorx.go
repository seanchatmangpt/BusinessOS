// Package services provides business logic for BusinessOS.
package services

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/config"
	"github.com/rhl/businessos-backend/internal/utils"
)

// SorxService handles communication with the Sorx skill execution engine.
type SorxService struct {
	pool          *pgxpool.Pool
	cfg           *config.Config
	encryptionKey []byte
	httpClient    *http.Client
}

// NewSorxService creates a new Sorx service.
func NewSorxService(pool *pgxpool.Pool, cfg *config.Config) *SorxService {
	// Derive encryption key from SecretKey (should be 32 bytes for AES-256)
	key := []byte(cfg.SecretKey)
	if len(key) < 32 {
		// Pad key if too short (in production, use proper key derivation)
		padded := make([]byte, 32)
		copy(padded, key)
		key = padded
	}

	return &SorxService{
		pool:          pool,
		cfg:           cfg,
		encryptionKey: key[:32],
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ============================================================================
// Credential Management
// ============================================================================

// CredentialTicketRequest represents a request from Sorx for a credential.
type CredentialTicketRequest struct {
	Provider    string    `json:"provider"`
	Scope       string    `json:"scope"`
	SkillID     string    `json:"skill_id"`
	ExecutionID string    `json:"execution_id"`
	UserID      string    `json:"user_id"`
	EngineID    string    `json:"engine_id"`
	SessionID   string    `json:"session_id"`
	Timestamp   time.Time `json:"timestamp"`
	Signature   []byte    `json:"signature"`
}

// CredentialTicket is issued to Sorx after validation.
type CredentialTicket struct {
	ID        uuid.UUID `json:"id"`
	RequestID string    `json:"request_id"`
	Provider  string    `json:"provider"`
	Scope     string    `json:"scope"`
	SkillID   string    `json:"skill_id"`
	UserID    string    `json:"user_id"`
	EngineID  string    `json:"engine_id"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Nonce     string    `json:"nonce"`
	Signature []byte    `json:"signature"`
}

// CredentialResponse contains the encrypted credential.
type CredentialResponse struct {
	TicketID            uuid.UUID `json:"ticket_id"`
	EncryptedCredential []byte    `json:"encrypted_credential"`
	Nonce               []byte    `json:"nonce"`
	Provider            string    `json:"provider"`
	ExpiresAt           time.Time `json:"expires_at"`
}

// ValidateTicketRequest validates a credential ticket request from Sorx.
func (s *SorxService) ValidateTicketRequest(ctx context.Context, req CredentialTicketRequest) error {
	// Check timestamp freshness (within 30 seconds)
	if time.Since(req.Timestamp) > 30*time.Second {
		return fmt.Errorf("request timestamp too old")
	}

	// Verify signature (in production, use proper HMAC verification)
	// For now, we trust requests from known engine IDs

	// Check if user has the provider connected
	var exists bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM user_integrations
			WHERE user_id = $1 AND provider_id = $2 AND status = 'connected'
		)
	`, req.UserID, req.Provider).Scan(&exists)

	if err != nil {
		return fmt.Errorf("failed to check integration: %w", err)
	}
	if !exists {
		return fmt.Errorf("user does not have %s connected", req.Provider)
	}

	return nil
}

// IssueCredentialTicket creates a signed ticket for credential retrieval.
func (s *SorxService) IssueCredentialTicket(ctx context.Context, req CredentialTicketRequest) (*CredentialTicket, error) {
	// Generate nonce
	nonce, err := utils.GenerateNonce(16)
	if err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ticket := &CredentialTicket{
		ID:        uuid.New(),
		RequestID: req.ExecutionID,
		Provider:  req.Provider,
		Scope:     req.Scope,
		SkillID:   req.SkillID,
		UserID:    req.UserID,
		EngineID:  req.EngineID,
		IssuedAt:  time.Now().UTC(),
		ExpiresAt: time.Now().UTC().Add(60 * time.Second), // 60 second TTL
		Nonce:     fmt.Sprintf("%x", nonce),
	}

	// Sign the ticket (simplified - in production use Ed25519)
	// ticket.Signature = s.signTicket(ticket)

	return ticket, nil
}

// RedeemTicket exchanges a ticket for the encrypted credential.
func (s *SorxService) RedeemTicket(ctx context.Context, ticket *CredentialTicket) (*CredentialResponse, error) {
	// Verify ticket hasn't expired
	if time.Now().After(ticket.ExpiresAt) {
		return nil, fmt.Errorf("ticket expired")
	}

	// Get the user's integration
	var accessToken, refreshToken []byte
	var tokenExpires *time.Time
	err := s.pool.QueryRow(ctx, `
		SELECT access_token_encrypted, refresh_token_encrypted, token_expires_at
		FROM user_integrations
		WHERE user_id = $1 AND provider_id = $2 AND status = 'connected'
	`, ticket.UserID, ticket.Provider).Scan(&accessToken, &refreshToken, &tokenExpires)

	if err != nil {
		return nil, fmt.Errorf("failed to get integration: %w", err)
	}

	// Update last_used_at
	_, err = s.pool.Exec(ctx, `
		UPDATE user_integrations SET last_used_at = NOW()
		WHERE user_id = $1 AND provider_id = $2
	`, ticket.UserID, ticket.Provider)
	if err != nil {
		// Log but don't fail
	}

	// The token is already encrypted in DB, return it
	// In production, re-encrypt with a session key
	response := &CredentialResponse{
		TicketID:            ticket.ID,
		EncryptedCredential: accessToken,
		Provider:            ticket.Provider,
	}
	if tokenExpires != nil {
		response.ExpiresAt = *tokenExpires
	}

	return response, nil
}

// ============================================================================
// Callback Handling
// ============================================================================

// CallbackRequest represents a callback from Sorx.
type CallbackRequest struct {
	ID          uuid.UUID   `json:"id"`
	Type        string      `json:"type"` // request_skill, request_agent, request_decision, request_data, return_result
	ExecutionID string      `json:"execution_id"`
	SkillID     string      `json:"skill_id"`
	StepID      string      `json:"step_id"`
	UserID      string      `json:"user_id"`
	Timestamp   time.Time   `json:"timestamp"`
	Payload     interface{} `json:"payload"`
}

// CallbackResponse is sent back to Sorx.
type CallbackResponse struct {
	ID        uuid.UUID   `json:"id"`
	RequestID uuid.UUID   `json:"request_id"`
	Success   bool        `json:"success"`
	Error     string      `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	Result    interface{} `json:"result,omitempty"`
}

// HandleCallback processes a callback from Sorx.
func (s *SorxService) HandleCallback(ctx context.Context, req CallbackRequest) (*CallbackResponse, error) {
	response := &CallbackResponse{
		ID:        uuid.New(),
		RequestID: req.ID,
		Timestamp: time.Now().UTC(),
		Success:   true,
	}

	var err error
	switch req.Type {
	case "request_skill":
		response.Result, err = s.handleSkillCallback(ctx, req)
	case "request_agent":
		response.Result, err = s.handleAgentCallback(ctx, req)
	case "request_decision":
		response.Result, err = s.handleDecisionCallback(ctx, req)
	case "request_data":
		response.Result, err = s.handleDataCallback(ctx, req)
	case "return_result":
		err = s.handleResultCallback(ctx, req)
	case "update_progress":
		err = s.handleProgressCallback(ctx, req)
	case "log_event":
		err = s.handleLogCallback(ctx, req)
	default:
		err = fmt.Errorf("unknown callback type: %s", req.Type)
	}

	if err != nil {
		response.Success = false
		response.Error = err.Error()
	}

	return response, nil
}

func (s *SorxService) handleSkillCallback(ctx context.Context, req CallbackRequest) (interface{}, error) {
	// Parse payload
	payload, ok := req.Payload.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid skill callback payload")
	}

	skillID, _ := payload["skill_id"].(string)
	params, _ := payload["params"].(map[string]interface{})

	// For now, return a placeholder - would trigger actual skill execution
	return map[string]interface{}{
		"execution_id": uuid.New().String(),
		"skill_id":     skillID,
		"status":       "queued",
		"params":       params,
	}, nil
}

func (s *SorxService) handleAgentCallback(ctx context.Context, req CallbackRequest) (interface{}, error) {
	// This would call the AI agent for reasoning
	payload, ok := req.Payload.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid agent callback payload")
	}

	agentRole, _ := payload["agent_role"].(string)
	task, _ := payload["task"].(string)

	// Placeholder - would call AI
	return map[string]interface{}{
		"response":  fmt.Sprintf("Agent %s processed task: %s", agentRole, task),
		"reasoning": "Placeholder reasoning",
	}, nil
}

func (s *SorxService) handleDecisionCallback(ctx context.Context, req CallbackRequest) (interface{}, error) {
	payload, ok := req.Payload.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid decision callback payload")
	}

	question, _ := payload["question"].(string)
	options, _ := payload["options"].([]interface{})
	priority, _ := payload["priority"].(string)
	if priority == "" {
		priority = "medium"
	}

	// Convert options to []string
	var optionStrs []string
	for _, opt := range options {
		if s, ok := opt.(string); ok {
			optionStrs = append(optionStrs, s)
		}
	}

	// Create pending decision
	var decisionID uuid.UUID
	err := s.pool.QueryRow(ctx, `
		INSERT INTO pending_decisions (
			execution_id, skill_id, step_id, user_id,
			question, options, priority, context
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`, req.ExecutionID, req.SkillID, req.StepID, req.UserID,
		question, optionStrs, priority, payload).Scan(&decisionID)

	if err != nil {
		return nil, fmt.Errorf("failed to create pending decision: %w", err)
	}

	return map[string]interface{}{
		"id":           decisionID,
		"execution_id": req.ExecutionID,
		"skill_id":     req.SkillID,
		"status":       "pending",
	}, nil
}

func (s *SorxService) handleDataCallback(ctx context.Context, req CallbackRequest) (interface{}, error) {
	payload, ok := req.Payload.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data callback payload")
	}

	dataType, _ := payload["data_type"].(string)
	query, _ := payload["query"].(map[string]interface{})

	// Route to appropriate data handler based on data_type
	switch dataType {
	case "client":
		return s.handleClientDataRequest(ctx, req.UserID, query)
	case "project":
		return s.handleProjectDataRequest(ctx, req.UserID, query)
	case "task":
		return s.handleTaskDataRequest(ctx, req.UserID, query)
	case "context":
		return s.handleContextDataRequest(ctx, req.UserID, query)
	case "daily_log":
		return s.handleDailyLogDataRequest(ctx, req.UserID, query)
	default:
		return nil, fmt.Errorf("unknown data type: %s", dataType)
	}
}

func (s *SorxService) handleResultCallback(ctx context.Context, req CallbackRequest) error {
	payload, ok := req.Payload.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid result callback payload")
	}

	status, _ := payload["status"].(string)
	result, _ := payload["result"].(map[string]interface{})
	errorMsg, _ := payload["error"].(string)
	metrics, _ := payload["metrics"].(map[string]interface{})

	// Convert to JSON for storage
	resultJSON, _ := json.Marshal(result)
	metricsJSON, _ := json.Marshal(metrics)

	// Update skill execution
	_, err := s.pool.Exec(ctx, `
		UPDATE skill_executions SET
			status = $2,
			result = $3,
			error = $4,
			metrics = $5,
			completed_at = NOW()
		WHERE id = $1::uuid
	`, req.ExecutionID, status, resultJSON, errorMsg, metricsJSON)

	return err
}

func (s *SorxService) handleProgressCallback(ctx context.Context, req CallbackRequest) error {
	// Could store progress in Redis for real-time updates
	// For now, just log
	return nil
}

func (s *SorxService) handleLogCallback(ctx context.Context, req CallbackRequest) error {
	// Could store in audit log
	return nil
}

// ============================================================================
// Data Request Handlers
// ============================================================================

func (s *SorxService) handleClientDataRequest(ctx context.Context, userID string, query map[string]interface{}) (interface{}, error) {
	action, _ := query["action"].(string)
	data, _ := query["data"].(map[string]interface{})

	switch action {
	case "create":
		// Create client
		var clientID uuid.UUID
		err := s.pool.QueryRow(ctx, `
			INSERT INTO clients (user_id, name, email, company_name, status, metadata)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id
		`, userID, data["name"], data["email"], data["company_name"], data["status"], data).Scan(&clientID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"id": clientID}, nil

	case "find":
		filter, _ := query["filter"].(map[string]interface{})
		// Simplified find
		if email, ok := filter["email"].(string); ok {
			var client map[string]interface{}
			err := s.pool.QueryRow(ctx, `
				SELECT id, name, email, company_name FROM clients
				WHERE user_id = $1 AND email = $2
			`, userID, email).Scan(&client)
			if err != nil {
				return nil, nil // Not found
			}
			return client, nil
		}
		return nil, nil

	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

func (s *SorxService) handleProjectDataRequest(ctx context.Context, userID string, query map[string]interface{}) (interface{}, error) {
	action, _ := query["action"].(string)
	data, _ := query["data"].(map[string]interface{})

	switch action {
	case "create":
		var projectID uuid.UUID
		err := s.pool.QueryRow(ctx, `
			INSERT INTO projects (user_id, name, description, status, client_id)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id
		`, userID, data["name"], data["description"], data["status"], data["client_id"]).Scan(&projectID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"id": projectID}, nil

	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

func (s *SorxService) handleTaskDataRequest(ctx context.Context, userID string, query map[string]interface{}) (interface{}, error) {
	action, _ := query["action"].(string)
	data, _ := query["data"].(map[string]interface{})

	switch action {
	case "create":
		var taskID uuid.UUID
		err := s.pool.QueryRow(ctx, `
			INSERT INTO tasks (user_id, title, description, priority, status, source)
			VALUES ($1, $2, $3, $4, 'pending', $5)
			RETURNING id
		`, userID, data["title"], data["description"], data["priority"], data["source"]).Scan(&taskID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"id": taskID}, nil

	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

func (s *SorxService) handleContextDataRequest(ctx context.Context, userID string, query map[string]interface{}) (interface{}, error) {
	action, _ := query["action"].(string)
	data, _ := query["data"].(map[string]interface{})

	switch action {
	case "create":
		// Create knowledge base entry
		contentJSON, _ := json.Marshal(data["content"])
		sourceJSON, _ := json.Marshal(data["source"])

		var contextID uuid.UUID
		err := s.pool.QueryRow(ctx, `
			INSERT INTO contexts (user_id, title, type, content, source, tags)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id
		`, userID, data["title"], data["type"], contentJSON, sourceJSON, data["tags"]).Scan(&contextID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"id": contextID}, nil

	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

func (s *SorxService) handleDailyLogDataRequest(ctx context.Context, userID string, query map[string]interface{}) (interface{}, error) {
	// Placeholder for daily log operations
	return nil, fmt.Errorf("daily_log operations not yet implemented")
}

// ============================================================================
// Pending Decisions
// ============================================================================

// GetPendingDecisions returns all pending decisions for a user.
func (s *SorxService) GetPendingDecisions(ctx context.Context, userID string) ([]map[string]interface{}, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, execution_id, skill_id, step_id,
		       question, options, input_fields, context,
		       priority, status, created_at, expires_at
		FROM pending_decisions
		WHERE user_id = $1 AND status = 'pending'
		ORDER BY
			CASE priority
				WHEN 'urgent' THEN 1
				WHEN 'high' THEN 2
				WHEN 'medium' THEN 3
				WHEN 'low' THEN 4
			END,
			created_at ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var decisions []map[string]interface{}
	for rows.Next() {
		var d struct {
			ID          uuid.UUID
			ExecutionID string
			SkillID     string
			StepID      string
			Question    string
			Options     []string
			InputFields interface{}
			Context     interface{}
			Priority    string
			Status      string
			CreatedAt   time.Time
			ExpiresAt   *time.Time
		}
		if err := rows.Scan(&d.ID, &d.ExecutionID, &d.SkillID, &d.StepID,
			&d.Question, &d.Options, &d.InputFields, &d.Context,
			&d.Priority, &d.Status, &d.CreatedAt, &d.ExpiresAt); err != nil {
			return nil, err
		}

		decisions = append(decisions, map[string]interface{}{
			"id":           d.ID,
			"execution_id": d.ExecutionID,
			"skill_id":     d.SkillID,
			"step_id":      d.StepID,
			"question":     d.Question,
			"options":      d.Options,
			"input_fields": d.InputFields,
			"context":      d.Context,
			"priority":     d.Priority,
			"status":       d.Status,
			"created_at":   d.CreatedAt,
			"expires_at":   d.ExpiresAt,
		})
	}

	return decisions, nil
}

// GetPendingDecision returns a single pending decision.
func (s *SorxService) GetPendingDecision(ctx context.Context, decisionID uuid.UUID) (map[string]interface{}, error) {
	var d struct {
		ID          uuid.UUID
		ExecutionID string
		SkillID     string
		StepID      string
		UserID      string
		Question    string
		Options     []string
		InputFields interface{}
		Context     interface{}
		Priority    string
		Status      string
		Decision    *string
		Inputs      interface{}
		DecidedBy   *string
		DecidedAt   *time.Time
		CreatedAt   time.Time
		ExpiresAt   *time.Time
	}

	err := s.pool.QueryRow(ctx, `
		SELECT id, execution_id, skill_id, step_id, user_id,
		       question, options, input_fields, context,
		       priority, status, decision, decision_inputs,
		       decided_by, decided_at, created_at, expires_at
		FROM pending_decisions
		WHERE id = $1
	`, decisionID).Scan(&d.ID, &d.ExecutionID, &d.SkillID, &d.StepID, &d.UserID,
		&d.Question, &d.Options, &d.InputFields, &d.Context,
		&d.Priority, &d.Status, &d.Decision, &d.Inputs,
		&d.DecidedBy, &d.DecidedAt, &d.CreatedAt, &d.ExpiresAt)

	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"id":              d.ID,
		"execution_id":    d.ExecutionID,
		"skill_id":        d.SkillID,
		"step_id":         d.StepID,
		"user_id":         d.UserID,
		"question":        d.Question,
		"options":         d.Options,
		"input_fields":    d.InputFields,
		"context":         d.Context,
		"priority":        d.Priority,
		"status":          d.Status,
		"decision":        d.Decision,
		"decision_inputs": d.Inputs,
		"decided_by":      d.DecidedBy,
		"decided_at":      d.DecidedAt,
		"created_at":      d.CreatedAt,
		"expires_at":      d.ExpiresAt,
	}, nil
}

// RespondToDecision records a human's decision.
func (s *SorxService) RespondToDecision(ctx context.Context, decisionID uuid.UUID, userID string, decision string, inputs map[string]interface{}) error {
	inputsJSON, _ := json.Marshal(inputs)

	_, err := s.pool.Exec(ctx, `
		UPDATE pending_decisions SET
			status = 'decided',
			decision = $2,
			decision_inputs = $3,
			decided_by = $4,
			decided_at = NOW()
		WHERE id = $1 AND status = 'pending'
	`, decisionID, decision, inputsJSON, userID)

	return err
}

// ============================================================================
// Encryption Helpers
// ============================================================================

// Encrypt encrypts data using AES-GCM.
func (s *SorxService) Encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// Decrypt decrypts data using AES-GCM.
func (s *SorxService) Decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
