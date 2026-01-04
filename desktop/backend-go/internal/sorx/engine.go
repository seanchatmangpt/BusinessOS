// Package sorx implements the Sorx skill execution engine.
// Sorx (System of Reasoning) executes skills using connected integrations.
package sorx

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Engine is the Sorx skill execution engine.
type Engine struct {
	pool   *pgxpool.Pool
	logger *slog.Logger

	// Execution tracking
	executions sync.Map // map[string]*Execution

	// Skill registry
	skills sync.Map // map[string]*SkillDefinition

	// Event bus for async communication
	events chan Event
	done   chan struct{}
}

// NewEngine creates a new Sorx engine.
func NewEngine(pool *pgxpool.Pool, logger *slog.Logger) *Engine {
	e := &Engine{
		pool:   pool,
		logger: logger,
		events: make(chan Event, 100),
		done:   make(chan struct{}),
	}

	// Register built-in skills
	e.registerBuiltinSkills()

	// Start event processor
	go e.processEvents()

	return e
}

// Close shuts down the engine gracefully.
func (e *Engine) Close() {
	close(e.done)
	close(e.events)
}

// ============================================================================
// Skill Execution
// ============================================================================

// ExecuteSkill starts a skill execution.
func (e *Engine) ExecuteSkill(ctx context.Context, req ExecuteRequest) (*Execution, error) {
	// Validate skill exists
	skillDef, ok := e.skills.Load(req.SkillID)
	if !ok {
		return nil, fmt.Errorf("skill not found: %s", req.SkillID)
	}
	skill := skillDef.(*SkillDefinition)

	// Check required integrations
	for _, provider := range skill.RequiredIntegrations {
		hasAccess, err := e.checkIntegrationAccess(ctx, req.UserID, provider)
		if err != nil {
			return nil, fmt.Errorf("failed to check integration %s: %w", provider, err)
		}
		if !hasAccess {
			return nil, fmt.Errorf("integration %s not connected", provider)
		}
	}

	// Create execution record
	exec := &Execution{
		ID:          uuid.New(),
		SkillID:     req.SkillID,
		UserID:      req.UserID,
		Status:      StatusPending,
		Params:      req.Params,
		Context:     make(map[string]interface{}),
		StepResults: make(map[string]interface{}),
		StartedAt:   time.Now().UTC(),
	}

	// Store execution
	e.executions.Store(exec.ID.String(), exec)

	// Persist to database
	if err := e.persistExecution(ctx, exec); err != nil {
		e.logger.Error("Failed to persist execution", "error", err, "execution_id", exec.ID)
	}

	// Start execution in background
	go e.runExecution(ctx, exec, skill)

	return exec, nil
}

// GetExecution retrieves an execution by ID.
func (e *Engine) GetExecution(id uuid.UUID) (*Execution, bool) {
	val, ok := e.executions.Load(id.String())
	if !ok {
		return nil, false
	}
	return val.(*Execution), true
}

// runExecution processes a skill execution.
func (e *Engine) runExecution(ctx context.Context, exec *Execution, skill *SkillDefinition) {
	exec.Status = StatusRunning
	e.updateExecution(ctx, exec)

	// Execute each step
	for i, step := range skill.Steps {
		exec.CurrentStep = i

		e.logger.Info("Executing step",
			"execution_id", exec.ID,
			"step_id", step.ID,
			"step_type", step.Type)

		result, err := e.executeStep(ctx, exec, &step)
		if err != nil {
			exec.Status = StatusFailed
			exec.Error = err.Error()
			e.updateExecution(ctx, exec)
			return
		}

		// Store step result
		exec.StepResults[step.ID] = result

		// Check if step requires human decision
		if step.RequiresDecision && result != nil {
			if decision, ok := result.(map[string]interface{}); ok {
				if decision["status"] == "pending" {
					exec.Status = StatusWaitingCallback
					e.updateExecution(ctx, exec)
					return // Will be resumed when decision is made
				}
			}
		}
	}

	// All steps complete
	exec.Status = StatusComplete
	exec.CompletedAt = timePtr(time.Now().UTC())
	e.updateExecution(ctx, exec)

	e.logger.Info("Execution completed", "execution_id", exec.ID)
}

// executeStep runs a single step in the skill.
func (e *Engine) executeStep(ctx context.Context, exec *Execution, step *Step) (interface{}, error) {
	switch step.Type {
	case StepTypeAction:
		return e.executeAction(ctx, exec, step)
	case StepTypeDecision:
		return e.requestDecision(ctx, exec, step)
	case StepTypeCondition:
		return e.evaluateCondition(ctx, exec, step)
	case StepTypeLoop:
		return e.executeLoop(ctx, exec, step)
	case StepTypeParallel:
		return e.executeParallel(ctx, exec, step)
	default:
		return nil, fmt.Errorf("unknown step type: %s", step.Type)
	}
}

// executeAction performs an integration action.
func (e *Engine) executeAction(ctx context.Context, exec *Execution, step *Step) (interface{}, error) {
	// Get action handler
	handler, ok := actionHandlers[step.Action]
	if !ok {
		return nil, fmt.Errorf("unknown action: %s", step.Action)
	}

	// Get integration credentials if needed
	var creds *Credentials
	if step.Integration != "" {
		var err error
		creds, err = e.getCredentials(ctx, exec.UserID, step.Integration)
		if err != nil {
			return nil, fmt.Errorf("failed to get credentials for %s: %w", step.Integration, err)
		}
	}

	// Execute action
	return handler(ctx, ActionContext{
		Execution:   exec,
		Step:        step,
		Credentials: creds,
		Params:      step.Params,
	})
}

// requestDecision creates a human-in-the-loop decision.
func (e *Engine) requestDecision(ctx context.Context, exec *Execution, step *Step) (interface{}, error) {
	decisionID := uuid.New()

	// Insert pending decision
	_, err := e.pool.Exec(ctx, `
		INSERT INTO pending_decisions (
			id, execution_id, skill_id, step_id, user_id,
			question, options, input_fields, context, priority, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 'pending')
	`, decisionID, exec.ID, exec.SkillID, step.ID, exec.UserID,
		step.DecisionQuestion, step.DecisionOptions, step.InputFields,
		exec.Context, step.Priority)

	if err != nil {
		return nil, fmt.Errorf("failed to create decision: %w", err)
	}

	e.logger.Info("Awaiting human decision",
		"execution_id", exec.ID,
		"decision_id", decisionID,
		"question", step.DecisionQuestion)

	return map[string]interface{}{
		"status":      "pending",
		"decision_id": decisionID,
	}, nil
}

// evaluateCondition checks a condition and returns the branch to take.
func (e *Engine) evaluateCondition(ctx context.Context, exec *Execution, step *Step) (interface{}, error) {
	// Simple condition evaluation
	// In production, this would use a proper expression evaluator
	condition := step.Condition
	if condition == "" {
		return map[string]interface{}{"branch": "default"}, nil
	}

	// For now, just return the condition result from params if present
	if result, ok := exec.Params[condition]; ok {
		return map[string]interface{}{"branch": result}, nil
	}

	return map[string]interface{}{"branch": "default"}, nil
}

// executeLoop runs steps in a loop.
func (e *Engine) executeLoop(ctx context.Context, exec *Execution, step *Step) (interface{}, error) {
	// Get items to iterate over
	items, ok := step.Params["items"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("loop step requires 'items' array in params")
	}

	var results []interface{}
	for i, item := range items {
		exec.Context["loop_index"] = i
		exec.Context["loop_item"] = item

		// Execute loop body (substeps)
		for _, substep := range step.Substeps {
			result, err := e.executeStep(ctx, exec, &substep)
			if err != nil {
				return nil, fmt.Errorf("loop iteration %d failed: %w", i, err)
			}
			results = append(results, result)
		}
	}

	return map[string]interface{}{"results": results}, nil
}

// executeParallel runs steps in parallel.
func (e *Engine) executeParallel(ctx context.Context, exec *Execution, step *Step) (interface{}, error) {
	var wg sync.WaitGroup
	results := make(map[string]interface{})
	var mu sync.Mutex
	var firstErr error

	for _, substep := range step.Substeps {
		wg.Add(1)
		go func(s Step) {
			defer wg.Done()
			result, err := e.executeStep(ctx, exec, &s)
			mu.Lock()
			defer mu.Unlock()
			if err != nil && firstErr == nil {
				firstErr = err
			}
			results[s.ID] = result
		}(substep)
	}

	wg.Wait()

	if firstErr != nil {
		return nil, firstErr
	}

	return results, nil
}

// ============================================================================
// Integration Helpers
// ============================================================================

// checkIntegrationAccess verifies a user has access to an integration.
func (e *Engine) checkIntegrationAccess(ctx context.Context, userID, provider string) (bool, error) {
	var exists bool
	err := e.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM user_integrations
			WHERE user_id = $1 AND provider_id = $2 AND status = 'connected'
		)
	`, userID, provider).Scan(&exists)
	return exists, err
}

// getCredentials retrieves encrypted credentials for an integration.
func (e *Engine) getCredentials(ctx context.Context, userID, provider string) (*Credentials, error) {
	var creds Credentials
	err := e.pool.QueryRow(ctx, `
		SELECT access_token_encrypted, refresh_token_encrypted, token_expires_at, scopes
		FROM user_integrations
		WHERE user_id = $1 AND provider_id = $2 AND status = 'connected'
	`, userID, provider).Scan(
		&creds.AccessTokenEncrypted,
		&creds.RefreshTokenEncrypted,
		&creds.ExpiresAt,
		&creds.Scopes,
	)
	if err != nil {
		return nil, err
	}
	creds.Provider = provider
	return &creds, nil
}

// ============================================================================
// Persistence
// ============================================================================

func (e *Engine) persistExecution(ctx context.Context, exec *Execution) error {
	_, err := e.pool.Exec(ctx, `
		INSERT INTO skill_executions (
			id, skill_id, user_id, status, params, context, started_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, exec.ID, exec.SkillID, exec.UserID, exec.Status, exec.Params, exec.Context, exec.StartedAt)
	return err
}

func (e *Engine) updateExecution(ctx context.Context, exec *Execution) error {
	_, err := e.pool.Exec(ctx, `
		UPDATE skill_executions SET
			status = $2,
			current_step = $3,
			step_results = $4,
			result = $5,
			error = $6,
			completed_at = $7
		WHERE id = $1
	`, exec.ID, exec.Status, exec.CurrentStep, exec.StepResults, exec.Result, exec.Error, exec.CompletedAt)
	return err
}

// ============================================================================
// Event Processing
// ============================================================================

func (e *Engine) processEvents() {
	for {
		select {
		case event := <-e.events:
			e.handleEvent(event)
		case <-e.done:
			return
		}
	}
}

func (e *Engine) handleEvent(event Event) {
	switch event.Type {
	case EventDecisionMade:
		e.resumeFromDecision(event.ExecutionID, event.Data)
	case EventIntegrationConnected:
		e.logger.Info("Integration connected", "data", event.Data)
	case EventIntegrationDisconnected:
		e.logger.Info("Integration disconnected", "data", event.Data)
	}
}

func (e *Engine) resumeFromDecision(executionID uuid.UUID, data interface{}) {
	exec, ok := e.GetExecution(executionID)
	if !ok {
		e.logger.Error("Execution not found for decision", "execution_id", executionID)
		return
	}

	if exec.Status != StatusWaitingCallback {
		e.logger.Warn("Execution not waiting for callback", "execution_id", executionID, "status", exec.Status)
		return
	}

	// Store decision result and continue
	exec.Context["decision_result"] = data
	exec.Status = StatusRunning

	// Resume execution from current step
	skill, ok := e.skills.Load(exec.SkillID)
	if !ok {
		e.logger.Error("Skill not found", "skill_id", exec.SkillID)
		return
	}

	go e.runExecution(context.Background(), exec, skill.(*SkillDefinition))
}

// ============================================================================
// Built-in Skills Registration
// ============================================================================

func (e *Engine) registerBuiltinSkills() {
	// ========================================================================
	// Register command-based skills (migrated from legacy commands)
	// ========================================================================
	RegisterCommandSkills(e)

	// ========================================================================
	// INTEGRATION SKILLS
	// ========================================================================

	// Email processing skill - Tier 3 (Reasoning AI)
	e.RegisterSkill(&SkillDefinition{
		ID:          "email.process_inbox",
		Name:        "Process Email Inbox",
		Description: "Scans inbox and extracts actionable items using AI analysis",
		Category:    "communication",
		Tier:        TierReasoningAI,
		RoleAffinity: []Role{RoleAny, RoleOperations},
		RequiredIntegrations: []string{"gmail"},
		DataPointsSatisfied: []string{"inbox.processed", "tasks.extracted"},
		RequiresApprovalAt: TemperatureWarm,
		Steps: []Step{
			{
				ID:          "fetch_emails",
				Type:        StepTypeAction,
				Action:      "gmail.list_messages",
				Integration: "gmail",
				Params:      map[string]interface{}{"max_results": 50, "label": "INBOX"},
			},
			{
				ID:     "analyze_with_agent",
				Type:   StepTypeAction,
				Action: "agent.analyst",
				Params: map[string]interface{}{
					"task": "Analyze these emails and extract: 1) Action items that should become tasks 2) Important dates/deadlines 3) Key information to remember. Format as structured JSON.",
					"from": "fetch_emails",
				},
			},
			{
				ID:          "create_tasks",
				Type:        StepTypeAction,
				Action:      "businessos.create_tasks",
				Params:      map[string]interface{}{"from": "analyze_with_agent"},
			},
		},
	})

	// CRM sync skill
	e.RegisterSkill(&SkillDefinition{
		ID:          "crm.sync_contacts",
		Name:        "Sync CRM Contacts",
		Description: "Syncs contacts from CRM to BusinessOS",
		Category:    "crm",
		RequiredIntegrations: []string{"hubspot"},
		Steps: []Step{
			{
				ID:          "fetch_contacts",
				Type:        StepTypeAction,
				Action:      "hubspot.list_contacts",
				Integration: "hubspot",
			},
			{
				ID:          "map_contacts",
				Type:        StepTypeAction,
				Action:      "transform.map_fields",
				Params:      map[string]interface{}{"mapping": "hubspot_to_client"},
			},
			{
				ID:          "upsert_clients",
				Type:        StepTypeAction,
				Action:      "businessos.upsert_clients",
			},
		},
	})

	// Task sync skill with decision
	e.RegisterSkill(&SkillDefinition{
		ID:          "tasks.import_with_review",
		Name:        "Import Tasks with Review",
		Description: "Imports tasks from external source with human review",
		Category:    "tasks",
		RequiredIntegrations: []string{},
		Steps: []Step{
			{
				ID:     "fetch_tasks",
				Type:   StepTypeAction,
				Action: "linear.list_issues",
			},
			{
				ID:               "review_tasks",
				Type:             StepTypeDecision,
				RequiresDecision: true,
				DecisionQuestion: "Which tasks should be imported?",
				DecisionOptions:  []string{"Import all", "Import assigned only", "Skip"},
				Priority:         "medium",
			},
			{
				ID:     "import_tasks",
				Type:   StepTypeAction,
				Action: "businessos.import_tasks",
			},
		},
	})

	// Calendar sync skill
	e.RegisterSkill(&SkillDefinition{
		ID:          "calendar.sync_events",
		Name:        "Sync Calendar Events",
		Description: "Syncs calendar events and creates daily log entries",
		Category:    "calendar",
		RequiredIntegrations: []string{"google_calendar"},
		Steps: []Step{
			{
				ID:          "fetch_events",
				Type:        StepTypeAction,
				Action:      "google_calendar.list_events",
				Integration: "google_calendar",
				Params:      map[string]interface{}{"days_ahead": 7},
			},
			{
				ID:     "create_log_entries",
				Type:   StepTypeAction,
				Action: "businessos.create_daily_log",
			},
		},
	})

	// Daily Brief skill - aggregates multiple sources with AI summarization
	e.RegisterSkill(&SkillDefinition{
		ID:          "daily.brief",
		Name:        "Generate Daily Brief",
		Description: "Creates a daily brief from calendar, tasks, and emails",
		Category:    "automation",
		RequiredIntegrations: []string{}, // Works with whatever is connected
		Steps: []Step{
			{
				ID:   "gather_calendar",
				Type: StepTypeAction,
				Action: "google_calendar.list_events",
				Params: map[string]interface{}{"days_ahead": 1},
				OnError: "continue", // Continue even if not connected
			},
			{
				ID:   "gather_tasks",
				Type: StepTypeAction,
				Action: "businessos.list_pending_tasks",
				OnError: "continue",
			},
			{
				ID:   "gather_emails",
				Type: StepTypeAction,
				Action: "gmail.list_messages",
				Params: map[string]interface{}{"max_results": 20, "label": "INBOX"},
				OnError: "continue",
			},
			{
				ID:   "synthesize_brief",
				Type: StepTypeAction,
				Action: "agent.orchestrator",
				Params: map[string]interface{}{
					"task": `Based on the gathered data, create a daily brief that includes:
1. **Today's Schedule** - Key meetings and events
2. **Priority Tasks** - Most important tasks to complete today
3. **Unread Emails** - Summary of important unread emails
4. **Key Reminders** - Any deadlines or important dates coming up

Format as a well-structured markdown document suitable for the user to review each morning.`,
				},
			},
			{
				ID:   "save_to_daily_log",
				Type: StepTypeAction,
				Action: "businessos.create_daily_log",
				Params: map[string]interface{}{"from": "synthesize_brief", "type": "daily_brief"},
			},
		},
	})

	// Knowledge extraction skill - uses document agent
	e.RegisterSkill(&SkillDefinition{
		ID:          "knowledge.extract_and_build",
		Name:        "Extract and Build Knowledge",
		Description: "Extracts key information from sources and creates knowledge nodes",
		Category:    "knowledge",
		RequiredIntegrations: []string{},
		Steps: []Step{
			{
				ID:   "analyze_source",
				Type: StepTypeAction,
				Action: "agent.analyst",
				Params: map[string]interface{}{
					"task": `Analyze the provided content and extract:
1. Key concepts and definitions
2. Important facts and figures
3. Relationships between entities
4. Actionable insights

Format each item as a structured knowledge node with: title, type, content, tags, and related_to fields.`,
				},
			},
			{
				ID:   "create_nodes",
				Type: StepTypeAction,
				Action: "businessos.create_nodes",
				Params: map[string]interface{}{"from": "analyze_source", "type": "knowledge"},
			},
		},
	})

	// Client analysis skill - uses analyst agent
	e.RegisterSkill(&SkillDefinition{
		ID:          "analysis.client_health",
		Name:        "Client Health Analysis",
		Description: "Analyzes client data and generates health report",
		Category:    "analysis",
		RequiredIntegrations: []string{},
		Steps: []Step{
			{
				ID:   "gather_client_data",
				Type: StepTypeAction,
				Action: "businessos.get_client_summary",
			},
			{
				ID:   "analyze_health",
				Type: StepTypeAction,
				Action: "agent.analyst",
				Params: map[string]interface{}{
					"task": `Analyze this client data and provide a health assessment including:
1. **Health Score** (1-10) with reasoning
2. **Strengths** - What's going well
3. **Risks** - Potential issues to address
4. **Recommendations** - Suggested actions
5. **Next Steps** - Specific tasks to improve the relationship`,
					"from": "gather_client_data",
				},
			},
		},
	})

	// Pipeline analysis skill
	e.RegisterSkill(&SkillDefinition{
		ID:          "analysis.pipeline",
		Name:        "Sales Pipeline Analysis",
		Description: "Analyzes sales pipeline and generates insights",
		Category:    "analysis",
		RequiredIntegrations: []string{},
		Steps: []Step{
			{
				ID:   "gather_pipeline",
				Type: StepTypeAction,
				Action: "businessos.get_pipeline_summary",
			},
			{
				ID:   "analyze_pipeline",
				Type: StepTypeAction,
				Action: "agent.analyst",
				Params: map[string]interface{}{
					"task": `Analyze this pipeline data and provide insights including:
1. **Pipeline Health** - Overall status and value
2. **Stage Distribution** - Where deals are concentrated
3. **Velocity Analysis** - How fast deals are moving
4. **At-Risk Deals** - Deals that need attention
5. **Forecast** - Projected outcomes based on current data
6. **Recommendations** - Strategic actions to improve pipeline`,
					"from": "gather_pipeline",
				},
			},
		},
	})

	// Meeting prep skill - uses orchestrator for comprehensive prep
	e.RegisterSkill(&SkillDefinition{
		ID:          "meeting.prepare",
		Name:        "Meeting Preparation",
		Description: "Prepares comprehensive brief for an upcoming meeting",
		Category:    "productivity",
		RequiredIntegrations: []string{},
		Steps: []Step{
			{
				ID:   "get_meeting_details",
				Type: StepTypeAction,
				Action: "google_calendar.get_event",
				OnError: "continue",
			},
			{
				ID:   "gather_context",
				Type: StepTypeAction,
				Action: "businessos.get_meeting_context",
				Params: map[string]interface{}{"from": "get_meeting_details"},
			},
			{
				ID:   "prepare_brief",
				Type: StepTypeAction,
				Action: "agent.orchestrator",
				Params: map[string]interface{}{
					"task": `Prepare a comprehensive meeting brief including:
1. **Meeting Overview** - Purpose, attendees, timing
2. **Attendee Profiles** - Relevant background on each attendee
3. **Historical Context** - Previous interactions and discussions
4. **Talking Points** - Key topics to cover
5. **Questions to Ask** - Strategic questions for the meeting
6. **Potential Objections** - Issues that might come up and how to address
7. **Desired Outcomes** - What success looks like for this meeting`,
					"from": "gather_context",
				},
			},
		},
	})
}

// RegisterSkill adds a skill to the registry.
func (e *Engine) RegisterSkill(skill *SkillDefinition) {
	e.skills.Store(skill.ID, skill)
	e.logger.Info("Registered skill", "skill_id", skill.ID, "name", skill.Name)
}

// ListSkills returns all registered skills.
func (e *Engine) ListSkills() []*SkillDefinition {
	var skills []*SkillDefinition
	e.skills.Range(func(key, value interface{}) bool {
		skills = append(skills, value.(*SkillDefinition))
		return true
	})
	return skills
}

// Helper
func timePtr(t time.Time) *time.Time {
	return &t
}
