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

// Legacy agent type constants used by the simple orchestrator implementation.
// The primary agent type definitions are in agent_v2.go.
const (
	legacyAgentTypeOrchestrator AgentType = "orchestrator"
	legacyAgentTypeDocument     AgentType = "document"
	legacyAgentTypeAnalysis     AgentType = "analysis"
	legacyAgentTypePlanning     AgentType = "planning"
)

// LegacyAgent interface defines the simple agent contract used by the legacy orchestrator.
// The primary Agent interface is in agent_v2.go.
type LegacyAgent interface {
	Name() string
	Description() string
	SystemPrompt() string
	Run(ctx context.Context, messages []services.ChatMessage) (<-chan string, <-chan error)
	SetOptions(opts services.LLMOptions)
}

// LegacyBaseAgent provides common functionality for the simple legacy agents.
type LegacyBaseAgent struct {
	pool           *pgxpool.Pool
	cfg            *config.Config
	userID         string
	conversationID *uuid.UUID
	model          string
	agentType      AgentType
	systemPrompt   string
	options        services.LLMOptions
}

// newLegacyBaseAgent creates a new legacy base agent.
func newLegacyBaseAgent(
	pool *pgxpool.Pool,
	cfg *config.Config,
	userID string,
	conversationID *uuid.UUID,
	model string,
	agentType AgentType,
	systemPrompt string,
) *LegacyBaseAgent {
	if model == "" {
		model = cfg.DefaultModel
	}
	return &LegacyBaseAgent{
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

func (a *LegacyBaseAgent) Name() string {
	return string(a.agentType)
}

func (a *LegacyBaseAgent) Description() string {
	descriptions := map[AgentType]string{
		legacyAgentTypeOrchestrator: "Main coordinator that handles requests and delegates to sub-agents",
		legacyAgentTypeDocument:     "Creates professional business documents",
		legacyAgentTypeAnalysis:     "Analyzes data and provides insights",
		legacyAgentTypePlanning:     "Helps with planning and prioritization",
	}
	return descriptions[a.agentType]
}

func (a *LegacyBaseAgent) SystemPrompt() string {
	return a.systemPrompt
}

func (a *LegacyBaseAgent) SetOptions(opts services.LLMOptions) {
	a.options = opts
}

func (a *LegacyBaseAgent) Run(ctx context.Context, messages []services.ChatMessage) (<-chan string, <-chan error) {
	llm := services.NewLLMService(a.cfg, a.model)
	llm.SetOptions(a.options)
	return llm.StreamChat(ctx, messages, a.systemPrompt)
}

// LegacyDocumentAgent creates business documents using the simple agent pattern.
type LegacyDocumentAgent struct {
	*LegacyBaseAgent
}

func newLegacyDocumentAgent(pool *pgxpool.Pool, cfg *config.Config, userID string, conversationID *uuid.UUID, model string) *LegacyDocumentAgent {
	return &LegacyDocumentAgent{
		LegacyBaseAgent: newLegacyBaseAgent(
			pool, cfg, userID, conversationID, model,
			legacyAgentTypeDocument,
			prompts.GetPromptWithArtifactInstruction("document"),
		),
	}
}

// LegacyAnalysisAgent analyzes data and provides insights using the simple agent pattern.
type LegacyAnalysisAgent struct {
	*LegacyBaseAgent
}

func newLegacyAnalysisAgent(pool *pgxpool.Pool, cfg *config.Config, userID string, conversationID *uuid.UUID, model string) *LegacyAnalysisAgent {
	return &LegacyAnalysisAgent{
		LegacyBaseAgent: newLegacyBaseAgent(
			pool, cfg, userID, conversationID, model,
			legacyAgentTypeAnalysis,
			prompts.GetPromptWithArtifactInstruction("analyst"),
		),
	}
}

// LegacyPlanningAgent helps with planning and prioritization using the simple agent pattern.
type LegacyPlanningAgent struct {
	*LegacyBaseAgent
}

func newLegacyPlanningAgent(pool *pgxpool.Pool, cfg *config.Config, userID string, conversationID *uuid.UUID, model string) *LegacyPlanningAgent {
	return &LegacyPlanningAgent{
		LegacyBaseAgent: newLegacyBaseAgent(
			pool, cfg, userID, conversationID, model,
			legacyAgentTypePlanning,
			prompts.GetPromptWithArtifactInstruction("planner"),
		),
	}
}

// OrchestratorAgent coordinates all other agents using the simple delegation pattern.
type OrchestratorAgent struct {
	*LegacyBaseAgent
	subAgents map[string]LegacyAgent
}

// ProjectContext holds project information for the orchestrator.
type ProjectContext struct {
	Name        string
	Description string
}

// NewOrchestratorAgent creates a simple orchestrator agent.
func NewOrchestratorAgent(pool *pgxpool.Pool, cfg *config.Config, userID string, conversationID *uuid.UUID, model string) *OrchestratorAgent {
	return &OrchestratorAgent{
		LegacyBaseAgent: newLegacyBaseAgent(
			pool, cfg, userID, conversationID, model,
			legacyAgentTypeOrchestrator,
			prompts.GetPrompt("orchestrator"),
		),
		subAgents: make(map[string]LegacyAgent),
	}
}

// NewOrchestratorAgentWithContext creates an orchestrator with project and user context.
func NewOrchestratorAgentWithContext(pool *pgxpool.Pool, cfg *config.Config, userID string, conversationID *uuid.UUID, model string, userName string, project *ProjectContext) *OrchestratorAgent {
	var projectName, projectDesc string
	if project != nil {
		projectName = project.Name
		projectDesc = project.Description
	}

	systemPrompt := prompts.BuildOrchestratorPromptWithContext(userName, projectName, projectDesc)

	return &OrchestratorAgent{
		LegacyBaseAgent: newLegacyBaseAgent(
			pool, cfg, userID, conversationID, model,
			legacyAgentTypeOrchestrator,
			systemPrompt,
		),
		subAgents: make(map[string]LegacyAgent),
	}
}

// getSubAgent returns or creates a sub-agent by name.
func (o *OrchestratorAgent) getSubAgent(name string) LegacyAgent {
	name = strings.ToLower(name)

	if agent, ok := o.subAgents[name]; ok {
		return agent
	}

	var agent LegacyAgent
	switch name {
	case "documentagent", "document":
		agent = newLegacyDocumentAgent(o.pool, o.cfg, o.userID, o.conversationID, o.model)
	case "analysisagent", "analysis":
		agent = newLegacyAnalysisAgent(o.pool, o.cfg, o.userID, o.conversationID, o.model)
	case "planningagent", "planning":
		agent = newLegacyPlanningAgent(o.pool, o.cfg, o.userID, o.conversationID, o.model)
	default:
		return nil
	}

	o.subAgents[name] = agent
	return agent
}

// parseDelegation parses a delegation instruction from the response.
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

// Run executes the orchestrator, potentially delegating to sub-agents.
func (o *OrchestratorAgent) Run(ctx context.Context, messages []services.ChatMessage) (<-chan string, <-chan error) {
	chunks := make(chan string, 100)
	errs := make(chan error, 1)

	go func() {
		defer close(chunks)
		defer close(errs)

		llm := services.NewLLMService(o.cfg, o.model)
		llm.SetOptions(o.options)
		decision, err := llm.ChatComplete(ctx, messages, o.systemPrompt)
		if err != nil {
			errs <- err
			return
		}

		agentName := parseDelegation(decision)

		if agentName != "" {
			subAgent := o.getSubAgent(agentName)
			if subAgent != nil {
				subAgent.SetOptions(o.options)

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
				delegatePattern := regexp.MustCompile(`\[DELEGATE:\w+\]`)
				cleanedDecision := delegatePattern.ReplaceAllString(decision, "")
				cleanedDecision = strings.TrimSpace(cleanedDecision)
				if cleanedDecision != "" {
					chunks <- cleanedDecision
				} else {
					chunks <- "I can help you with that. What specifically would you like to work on?"
				}
			}
		} else {
			streamChunks, streamErrs := o.LegacyBaseAgent.Run(ctx, messages)

			delegatePattern := regexp.MustCompile(`\[DELEGATE:\w+\]`)

			for {
				select {
				case chunk, ok := <-streamChunks:
					if !ok {
						return
					}
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
