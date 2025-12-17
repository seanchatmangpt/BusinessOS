package agents

import (
	"context"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/config"
	"github.com/rhl/businessos-backend/internal/prompts"
	"github.com/rhl/businessos-backend/internal/services"
)

// AgentType represents different agent types
type AgentType string

const (
	AgentTypeOrchestrator AgentType = "orchestrator"
	AgentTypeDocument     AgentType = "document"
	AgentTypeAnalysis     AgentType = "analysis"
	AgentTypePlanning     AgentType = "planning"
)

// GetAgentForFocusMode maps a focus mode to the appropriate agent type
func GetAgentForFocusMode(focusMode string) AgentType {
	switch focusMode {
	case "research", "analyze":
		return AgentTypeAnalysis
	case "write":
		return AgentTypeDocument
	case "build":
		return AgentTypePlanning
	case "general":
		return AgentTypeOrchestrator
	default:
		return AgentTypeOrchestrator
	}
}

// Agent interface defines what all agents must implement
type Agent interface {
	Name() string
	Description() string
	SystemPrompt() string
	Run(ctx context.Context, messages []services.ChatMessage) (<-chan string, <-chan error)
	SetOptions(opts services.LLMOptions)
}

// BaseAgent provides common functionality for all agents
type BaseAgent struct {
	pool           *pgxpool.Pool
	cfg            *config.Config
	userID         string
	conversationID *uuid.UUID
	model          string
	agentType      AgentType
	systemPrompt   string
	options        services.LLMOptions
}

// NewBaseAgent creates a new base agent
func NewBaseAgent(
	pool *pgxpool.Pool,
	cfg *config.Config,
	userID string,
	conversationID *uuid.UUID,
	model string,
	agentType AgentType,
	systemPrompt string,
) *BaseAgent {
	if model == "" {
		model = cfg.DefaultModel
	}
	return &BaseAgent{
		pool:           pool,
		cfg:            cfg,
		userID:         userID,
		conversationID: conversationID,
		model:          model,
		agentType:      agentType,
		systemPrompt:   systemPrompt,
		options:        services.DefaultLLMOptions(),
	}
}

func (a *BaseAgent) Name() string {
	return string(a.agentType)
}

func (a *BaseAgent) Description() string {
	descriptions := map[AgentType]string{
		AgentTypeOrchestrator: "Main coordinator that handles requests and delegates to sub-agents",
		AgentTypeDocument:     "Creates professional business documents",
		AgentTypeAnalysis:     "Analyzes data and provides insights",
		AgentTypePlanning:     "Helps with planning and prioritization",
	}
	return descriptions[a.agentType]
}

func (a *BaseAgent) SystemPrompt() string {
	return a.systemPrompt
}

func (a *BaseAgent) SetOptions(opts services.LLMOptions) {
	a.options = opts
}

func (a *BaseAgent) Run(ctx context.Context, messages []services.ChatMessage) (<-chan string, <-chan error) {
	llm := services.NewLLMService(a.cfg, a.model)
	llm.SetOptions(a.options)
	return llm.StreamChat(ctx, messages, a.systemPrompt)
}

// DocumentAgent creates business documents
type DocumentAgent struct {
	*BaseAgent
}

func NewDocumentAgent(pool *pgxpool.Pool, cfg *config.Config, userID string, conversationID *uuid.UUID, model string) *DocumentAgent {
	return &DocumentAgent{
		BaseAgent: NewBaseAgent(
			pool, cfg, userID, conversationID, model,
			AgentTypeDocument,
			prompts.GetPromptWithArtifactInstruction("document"),
		),
	}
}

// AnalysisAgent analyzes data and provides insights
type AnalysisAgent struct {
	*BaseAgent
}

func NewAnalysisAgent(pool *pgxpool.Pool, cfg *config.Config, userID string, conversationID *uuid.UUID, model string) *AnalysisAgent {
	return &AnalysisAgent{
		BaseAgent: NewBaseAgent(
			pool, cfg, userID, conversationID, model,
			AgentTypeAnalysis,
			prompts.GetPromptWithArtifactInstruction("analyst"),
		),
	}
}

// PlanningAgent helps with planning and prioritization
type PlanningAgent struct {
	*BaseAgent
}

func NewPlanningAgent(pool *pgxpool.Pool, cfg *config.Config, userID string, conversationID *uuid.UUID, model string) *PlanningAgent {
	return &PlanningAgent{
		BaseAgent: NewBaseAgent(
			pool, cfg, userID, conversationID, model,
			AgentTypePlanning,
			prompts.GetPromptWithArtifactInstruction("planner"),
		),
	}
}

// OrchestratorAgent coordinates all other agents
type OrchestratorAgent struct {
	*BaseAgent
	subAgents map[string]Agent
}

// ProjectContext holds project information for the orchestrator
type ProjectContext struct {
	Name        string
	Description string
}

func NewOrchestratorAgent(pool *pgxpool.Pool, cfg *config.Config, userID string, conversationID *uuid.UUID, model string) *OrchestratorAgent {
	return &OrchestratorAgent{
		BaseAgent: NewBaseAgent(
			pool, cfg, userID, conversationID, model,
			AgentTypeOrchestrator,
			prompts.GetPrompt("orchestrator"),
		),
		subAgents: make(map[string]Agent),
	}
}

// NewOrchestratorAgentWithContext creates an orchestrator with project and user context
func NewOrchestratorAgentWithContext(pool *pgxpool.Pool, cfg *config.Config, userID string, conversationID *uuid.UUID, model string, userName string, project *ProjectContext) *OrchestratorAgent {
	var projectName, projectDesc string
	if project != nil {
		projectName = project.Name
		projectDesc = project.Description
	}

	systemPrompt := prompts.BuildOrchestratorPromptWithContext(userName, projectName, projectDesc)

	return &OrchestratorAgent{
		BaseAgent: NewBaseAgent(
			pool, cfg, userID, conversationID, model,
			AgentTypeOrchestrator,
			systemPrompt,
		),
		subAgents: make(map[string]Agent),
	}
}

// getSubAgent returns or creates a sub-agent by name
func (o *OrchestratorAgent) getSubAgent(name string) Agent {
	name = strings.ToLower(name)

	if agent, ok := o.subAgents[name]; ok {
		return agent
	}

	var agent Agent
	switch name {
	case "documentagent", "document":
		agent = NewDocumentAgent(o.pool, o.cfg, o.userID, o.conversationID, o.model)
	case "analysisagent", "analysis":
		agent = NewAnalysisAgent(o.pool, o.cfg, o.userID, o.conversationID, o.model)
	case "planningagent", "planning":
		agent = NewPlanningAgent(o.pool, o.cfg, o.userID, o.conversationID, o.model)
	default:
		return nil
	}

	o.subAgents[name] = agent
	return agent
}

// parseDelegation parses a delegation instruction from the response
func parseDelegation(response string) (agentName string) {
	if !strings.Contains(response, "[DELEGATE:") {
		return ""
	}

	pattern := regexp.MustCompile(`\[DELEGATE:(\w+)\]`)
	matches := pattern.FindStringSubmatch(response)
	if len(matches) < 2 {
		return ""
	}

	return matches[1]
}

// Run executes the orchestrator, potentially delegating to sub-agents
func (o *OrchestratorAgent) Run(ctx context.Context, messages []services.ChatMessage) (<-chan string, <-chan error) {
	chunks := make(chan string, 100)
	errs := make(chan error, 1)

	go func() {
		defer close(chunks)
		defer close(errs)

		// First, get orchestrator's decision (non-streaming)
		llm := services.NewLLMService(o.cfg, o.model)
		llm.SetOptions(o.options)
		decision, err := llm.ChatComplete(ctx, messages, o.systemPrompt)
		if err != nil {
			errs <- err
			return
		}

		// Check if we should delegate
		agentName := parseDelegation(decision)

		if agentName != "" {
			// Delegate to sub-agent
			subAgent := o.getSubAgent(agentName)
			if subAgent != nil {
				// Pass LLM options to sub-agent
				subAgent.SetOptions(o.options)

				// Sub-agent receives the original messages (full conversation context)
				// No need to add extra context - the sub-agent has the full conversation

				// Stream sub-agent response
				subChunks, subErrs := subAgent.Run(ctx, messages)

				for {
					select {
					case chunk, ok := <-subChunks:
						if !ok {
							return
						}
						chunks <- chunk
					case err := <-subErrs:
						if err != nil {
							errs <- err
						}
						return
					case <-ctx.Done():
						return
					}
				}
			} else {
				// Unknown agent, strip the delegate tag and return cleaned response
				delegatePattern := regexp.MustCompile(`\[DELEGATE:\w+\]`)
				cleanedDecision := delegatePattern.ReplaceAllString(decision, "")
				cleanedDecision = strings.TrimSpace(cleanedDecision)
				if cleanedDecision != "" {
					chunks <- cleanedDecision
				} else {
					// If nothing left after stripping, give a helpful response
					chunks <- "I can help you with that. What specifically would you like to work on?"
				}
			}
		} else {
			// No delegation, stream orchestrator's response
			streamChunks, streamErrs := o.BaseAgent.Run(ctx, messages)

			// Clean any leaked delegation tags from streamed output
			delegatePattern := regexp.MustCompile(`\[DELEGATE:\w+\]`)

			for {
				select {
				case chunk, ok := <-streamChunks:
					if !ok {
						return
					}
					// Filter out any leaked [DELEGATE:X] tags
					cleanChunk := delegatePattern.ReplaceAllString(chunk, "")
					if cleanChunk != "" {
						chunks <- cleanChunk
					}
				case err := <-streamErrs:
					if err != nil {
						errs <- err
					}
					return
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return chunks, errs
}

// GetAgent returns an agent by type
func GetAgent(
	agentType AgentType,
	pool *pgxpool.Pool,
	cfg *config.Config,
	userID string,
	conversationID *uuid.UUID,
	model string,
) Agent {
	switch agentType {
	case AgentTypeOrchestrator:
		return NewOrchestratorAgent(pool, cfg, userID, conversationID, model)
	case AgentTypeDocument:
		return NewDocumentAgent(pool, cfg, userID, conversationID, model)
	case AgentTypeAnalysis:
		return NewAnalysisAgent(pool, cfg, userID, conversationID, model)
	case AgentTypePlanning:
		return NewPlanningAgent(pool, cfg, userID, conversationID, model)
	default:
		return NewOrchestratorAgent(pool, cfg, userID, conversationID, model)
	}
}

// GetAgentWithContext creates an agent with user and project context
func GetAgentWithContext(
	agentType AgentType,
	pool *pgxpool.Pool,
	cfg *config.Config,
	userID string,
	conversationID *uuid.UUID,
	model string,
	userName string,
	project *ProjectContext,
) Agent {
	switch agentType {
	case AgentTypeOrchestrator:
		return NewOrchestratorAgentWithContext(pool, cfg, userID, conversationID, model, userName, project)
	case AgentTypeDocument:
		return NewDocumentAgent(pool, cfg, userID, conversationID, model)
	case AgentTypeAnalysis:
		return NewAnalysisAgent(pool, cfg, userID, conversationID, model)
	case AgentTypePlanning:
		return NewPlanningAgent(pool, cfg, userID, conversationID, model)
	default:
		return NewOrchestratorAgentWithContext(pool, cfg, userID, conversationID, model, userName, project)
	}
}
