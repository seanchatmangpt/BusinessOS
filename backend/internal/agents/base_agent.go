package agents

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/config"
	"github.com/rhl/businessos-backend/internal/prompts"
	prompts_agents "github.com/rhl/businessos-backend/internal/prompts/agents"
	"github.com/rhl/businessos-backend/internal/services"
	"github.com/rhl/businessos-backend/internal/streaming"
	"github.com/rhl/businessos-backend/internal/tools"
)

// Sentinel errors for agent operations
var (
	// ErrToolRegistryNotAvailable indicates that the tool registry is not initialized
	ErrToolRegistryNotAvailable = errors.New("tool registry not available")
	// ErrToolNotEnabled indicates that a tool is not enabled for the agent
	ErrToolNotEnabled = errors.New("tool not enabled for this agent")
)

// BaseAgent provides common functionality for all agents
type BaseAgent struct {
	pool               *pgxpool.Pool
	cfg                *config.Config
	userID             string
	userName           string
	conversationID     *uuid.UUID
	model              string
	agentType          AgentType
	agentName          string
	description        string
	systemPrompt       string
	focusModePrompt    string // Focus mode specific prompt prefix
	outputStylePrompt  string // Output style specific instructions
	roleContextPrompt  string // Role-based permissions context (Feature 1)
	memoryContext      string // Workspace memory context (Feature: Memory Hierarchy)
	skillsPrompt       string // Available skills context (Agent Skills System)
	profileContext     string // User onboarding profile context (for personalization)
	contextReqs        ContextRequirements
	llmOptions         services.LLMOptions
	toolRegistry       *tools.AgentToolRegistry
	enabledTools       []string                     // Tool names this agent can use
	promptPersonalizer *services.PromptPersonalizer // For personalizing prompts with user data
	lastUserMessage    string                       // Last user message (for semantic personalization)
}

// BaseAgentConfig holds configuration for creating a BaseAgent
type BaseAgentConfig struct {
	Pool               *pgxpool.Pool
	Config             *config.Config
	UserID             string
	UserName           string
	ConversationID     *uuid.UUID
	Model              string
	AgentType          AgentType
	AgentName          string
	Description        string
	SystemPrompt       string
	ContextReqs        ContextRequirements
	EnabledTools       []string                     // Tool names this agent can use
	EmbeddingService   *services.EmbeddingService   // For context tools (tree_search, browse_tree, load_context)
	PromptPersonalizer *services.PromptPersonalizer // For personalizing prompts with user data
}

// NewBaseAgent creates a new base agent with the given configuration
func NewBaseAgent(cfg BaseAgentConfig) *BaseAgent {
	model := cfg.Model
	if model == "" && cfg.Config != nil {
		model = cfg.Config.GetActiveModel()
	}

	// Create tool registry if pool is available
	var toolRegistry *tools.AgentToolRegistry
	if cfg.Pool != nil && cfg.UserID != "" {
		if cfg.EmbeddingService != nil {
			// Use embedding-enabled registry for context tools (tree_search, browse_tree, load_context)
			toolRegistry = tools.NewAgentToolRegistryWithEmbedding(cfg.Pool, cfg.UserID, cfg.EmbeddingService)
		} else {
			toolRegistry = tools.NewAgentToolRegistry(cfg.Pool, cfg.UserID)
		}
	}

	return &BaseAgent{
		pool:               cfg.Pool,
		cfg:                cfg.Config,
		userID:             cfg.UserID,
		userName:           cfg.UserName,
		conversationID:     cfg.ConversationID,
		model:              model,
		agentType:          cfg.AgentType,
		agentName:          cfg.AgentName,
		description:        cfg.Description,
		systemPrompt:       cfg.SystemPrompt,
		contextReqs:        cfg.ContextReqs,
		llmOptions:         services.DefaultLLMOptions(),
		toolRegistry:       toolRegistry,
		enabledTools:       cfg.EnabledTools,
		promptPersonalizer: cfg.PromptPersonalizer,
	}
}

// Type returns the agent type
func (a *BaseAgent) Type() AgentType {
	return a.agentType
}

// Name returns the agent name
func (a *BaseAgent) Name() string {
	return a.agentName
}

// Description returns the agent description
func (a *BaseAgent) Description() string {
	return a.description
}

// GetSystemPrompt returns the system prompt
func (a *BaseAgent) GetSystemPrompt() string {
	return a.systemPrompt
}

// GetContextRequirements returns what context the agent needs
func (a *BaseAgent) GetContextRequirements() ContextRequirements {
	return a.contextReqs
}

// SetModel sets the model to use
func (a *BaseAgent) SetModel(model string) {
	a.model = model
}

// SetOptions sets the LLM options
func (a *BaseAgent) SetOptions(opts services.LLMOptions) {
	a.llmOptions = opts
}

// GetOptions returns the current LLM options
func (a *BaseAgent) GetOptions() services.LLMOptions {
	return a.llmOptions
}

// SetCustomSystemPrompt overrides the system prompt with a custom one (for custom agents)
func (a *BaseAgent) SetCustomSystemPrompt(prompt string) {
	slog.Debug("agent SetCustomSystemPrompt called", "prompt_len", len(prompt))
	if prompt != "" {
		a.systemPrompt = prompt
		slog.Debug("agent custom systemPrompt set", "len", len(a.systemPrompt))
	}
}

// SetFocusModePrompt sets a focus mode specific prompt prefix
func (a *BaseAgent) SetFocusModePrompt(prompt string) {
	a.focusModePrompt = prompt
}

// SetOutputStylePrompt sets an output style specific prompt section
func (a *BaseAgent) SetOutputStylePrompt(prompt string) {
	a.outputStylePrompt = prompt
}

// SetRoleContextPrompt sets role-based permission context (Feature 1)
func (a *BaseAgent) SetRoleContextPrompt(prompt string) {
	a.roleContextPrompt = prompt
	slog.Default().Debug("[Agent] SetRoleContextPrompt called", "chars", len(prompt))
}

// SetMemoryContext sets workspace memory context (Feature: Memory Hierarchy)
func (a *BaseAgent) SetMemoryContext(context string) {
	a.memoryContext = context
}

// SetSkillsPrompt sets the available skills context (Agent Skills System)
func (a *BaseAgent) SetSkillsPrompt(prompt string) {
	a.skillsPrompt = prompt
}

// SetProfileContext sets the user profile context for personalization
func (a *BaseAgent) SetProfileContext(context string) {
	a.profileContext = context
	slog.Default().Debug("[Agent] SetProfileContext called", "chars", len(context))
}

// SetLastUserMessage stores the last user message for personalization
func (a *BaseAgent) SetLastUserMessage(message string) {
	a.lastUserMessage = message
}

// GetEnabledTools returns the list of tools this agent can use
func (a *BaseAgent) GetEnabledTools() []string {
	return a.enabledTools
}

// GetToolDefinitions returns tool definitions for LLM function calling
func (a *BaseAgent) GetToolDefinitions() []map[string]interface{} {
	if a.toolRegistry == nil || len(a.enabledTools) == 0 {
		return nil
	}

	defs := make([]map[string]interface{}, 0)
	for _, toolName := range a.enabledTools {
		if tool, ok := a.toolRegistry.GetTool(toolName); ok {
			defs = append(defs, map[string]interface{}{
				"type": "function",
				"function": map[string]interface{}{
					"name":        tool.Name(),
					"description": tool.Description(),
					"parameters":  tool.InputSchema(),
				},
			})
		}
	}
	return defs
}

// ExecuteTool executes a tool by name with the given input
func (a *BaseAgent) ExecuteTool(ctx context.Context, toolName string, input json.RawMessage) (string, error) {
	if a.toolRegistry == nil {
		return "", ErrToolRegistryNotAvailable
	}

	// Check if tool is enabled for this agent
	enabled := false
	for _, t := range a.enabledTools {
		if t == toolName {
			enabled = true
			break
		}
	}
	if !enabled {
		return "", fmt.Errorf("%w: %s", ErrToolNotEnabled, toolName)
	}

	return a.toolRegistry.ExecuteTool(ctx, toolName, input)
}

// RunWithTools executes the agent with tool calling support (non-streaming)
func (a *BaseAgent) RunWithTools(ctx context.Context, input AgentInput) (<-chan streaming.StreamEvent, <-chan error) {
	events := make(chan streaming.StreamEvent, 100)
	errs := make(chan error, 1)

	go func() {
		defer close(events)
		defer close(errs)

		// Build messages with context
		messages := a.buildMessages(input)

		// Check if we have tools enabled
		if a.toolRegistry == nil || len(a.enabledTools) == 0 {
			// No tools, fall back to regular streaming
			a.runStreaming(ctx, input, events, errs)
			return
		}

		// Get tool definitions for enabled tools
		toolDefs := make([]services.ToolDefinition, 0)
		for _, toolName := range a.enabledTools {
			if tool, ok := a.toolRegistry.GetTool(toolName); ok {
				toolDefs = append(toolDefs, services.ToolDefinition{
					Name:        tool.Name(),
					Description: tool.Description(),
					Parameters:  tool.InputSchema(),
				})
			}
		}

		// Create Groq service for tool calling
		groqService := services.NewGroqService(a.cfg, a.model)
		groqService.SetOptions(a.llmOptions)

		// Build system prompt with thinking instructions if enabled
		systemPrompt := a.buildSystemPromptWithThinking()

		// First call with tools
		resp, err := groqService.ChatWithTools(ctx, messages, systemPrompt, toolDefs)
		if err != nil {
			errs <- err
			return
		}

		// Process tool calls if any
		if len(resp.ToolCalls) > 0 {
			// Send thinking event
			events <- streaming.StreamEvent{
				Type: streaming.EventTypeThinking,
				Data: "Executing tools...",
			}

			toolResults := make(map[string]string)
			for _, tc := range resp.ToolCalls {
				// Send tool execution event
				events <- streaming.StreamEvent{
					Type: streaming.EventTypeThinking,
					Data: fmt.Sprintf("Running tool: %s", tc.Name),
				}

				// Execute the tool
				result, err := a.toolRegistry.ExecuteTool(ctx, tc.Name, json.RawMessage(tc.Arguments))
				if err != nil {
					toolResults[tc.ID] = fmt.Sprintf("Error: %s", err.Error())
				} else {
					toolResults[tc.ID] = result
				}

				// Send tool result as event
				events <- streaming.StreamEvent{
					Type: streaming.EventTypeToken,
					Data: fmt.Sprintf("\n\n**Tool Result (%s):**\n%s\n\n", tc.Name, toolResults[tc.ID]),
				}
			}

			// Continue conversation with tool results
			finalResponse, err := groqService.ContinueWithToolResults(ctx, messages, systemPrompt, toolResults)
			if err != nil {
				errs <- err
				return
			}

			// Stream the final response
			for _, chunk := range splitIntoChunks(finalResponse, 50) {
				events <- streaming.StreamEvent{
					Type: streaming.EventTypeToken,
					Data: chunk,
				}
			}
		} else {
			// No tool calls, stream the response directly
			for _, chunk := range splitIntoChunks(resp.Content, 50) {
				events <- streaming.StreamEvent{
					Type: streaming.EventTypeToken,
					Data: chunk,
				}
			}
		}

		events <- streaming.StreamEvent{Type: streaming.EventTypeDone}
	}()

	return events, errs
}

// runStreaming handles regular streaming without tools
func (a *BaseAgent) runStreaming(ctx context.Context, input AgentInput, events chan<- streaming.StreamEvent, errs chan<- error) {
	messages := a.buildMessages(input)
	llm := services.NewLLMService(a.cfg, a.model)
	llm.SetOptions(a.llmOptions)
	detector := streaming.NewArtifactDetector()

	// Build system prompt with thinking instructions if enabled
	systemPrompt := a.buildSystemPromptWithThinking()
	chunks, llmErrs := llm.StreamChat(ctx, messages, systemPrompt)

	for {
		select {
		case chunk, ok := <-chunks:
			if !ok {
				for _, event := range detector.Flush() {
					events <- event
				}
				events <- streaming.StreamEvent{Type: streaming.EventTypeDone}
				return
			}
			for _, event := range detector.ProcessChunk(chunk) {
				events <- event
			}
		case err, ok := <-llmErrs:
			if ok && err != nil {
				errs <- err
				return
			}
			// Channel closed or nil error — stop selecting on it
			llmErrs = nil
		case <-ctx.Done():
			return
		}
	}
}

// splitIntoChunks splits a string into chunks for streaming
func splitIntoChunks(s string, chunkSize int) []string {
	var chunks []string
	for i := 0; i < len(s); i += chunkSize {
		end := i + chunkSize
		if end > len(s) {
			end = len(s)
		}
		chunks = append(chunks, s[i:end])
	}
	return chunks
}

// Run executes the agent with streaming output
func (a *BaseAgent) Run(ctx context.Context, input AgentInput) (<-chan streaming.StreamEvent, <-chan error) {
	events := make(chan streaming.StreamEvent, 100)
	errs := make(chan error, 1)

	go func() {
		defer close(events)
		defer close(errs)

		// Build messages with context
		messages := a.buildMessages(input)

		// Create LLM service
		llm := services.NewLLMService(a.cfg, a.model)
		llm.SetOptions(a.llmOptions)

		// Build system prompt with thinking instructions if enabled
		systemPrompt := a.buildSystemPromptWithThinking()

		// Create artifact detector for streaming
		detector := streaming.NewArtifactDetector()

		// Stream response
		chunks, llmErrs := llm.StreamChat(ctx, messages, systemPrompt)

		// Process chunks through artifact detector
		for {
			select {
			case chunk, ok := <-chunks:
				if !ok {
					// Stream ended - flush detector
					for _, event := range detector.Flush() {
						events <- event
					}
					events <- streaming.StreamEvent{Type: streaming.EventTypeDone}
					return
				}
				// Process chunk through artifact detector
				for _, event := range detector.ProcessChunk(chunk) {
					events <- event
				}

			case err, ok := <-llmErrs:
				if ok && err != nil {
					errs <- err
					return
				}
				// Channel closed or nil error — stop selecting on it
				llmErrs = nil

			case <-ctx.Done():
				return
			}
		}
	}()

	return events, errs
}

// buildMessages prepares messages for the LLM, including context
func (a *BaseAgent) buildMessages(input AgentInput) []services.ChatMessage {
	messages := make([]services.ChatMessage, 0, len(input.Messages)+1)

	// Capture last user message for personalization
	for i := len(input.Messages) - 1; i >= 0; i-- {
		if input.Messages[i].Role == "user" {
			a.lastUserMessage = input.Messages[i].Content
			break
		}
	}

	// Prepend context as system message if available
	if input.Context != nil {
		contextContent := ""
		if a.contextReqs.MaxContextTokens > 0 {
			contextContent = input.Context.FormatForAIWithTokenBudget(a.contextReqs.MaxContextTokens)
		} else {
			contextContent = input.Context.FormatForAI()
		}
		if contextContent != "" {
			contextMsg := services.ChatMessage{
				Role:    "system",
				Content: contextContent,
			}
			messages = append(messages, contextMsg)
		}
	}

	// Add conversation messages
	messages = append(messages, input.Messages...)

	return messages
}

// buildSystemPromptWithThinking returns the system prompt with thinking instructions if enabled
func (a *BaseAgent) buildSystemPromptWithThinking() string {
	slog.Debug("buildSystemPromptWithThinking called", "systemPrompt_len", len(a.systemPrompt))

	// CRITICAL: Start with profile context at the VERY BEGINNING
	// Profile context helps personalize responses based on user's business context
	var result string
	if a.profileContext != "" {
		result = a.profileContext
		slog.Default().Debug("[Agent] ✓ PROFILE CONTEXT placed at START of prompt",
			"chars", len(a.profileContext))
	}

	// Then add role context (for permissions)
	if a.roleContextPrompt != "" {
		if result != "" {
			result += "\n\n"
		}
		result += a.roleContextPrompt
		slog.Default().Debug("[Agent] ✓ ROLE CONTEXT added",
			"chars", len(a.roleContextPrompt))
	}

	// Then add focus mode prompt if set
	if a.focusModePrompt != "" {
		if result != "" {
			result += "\n\n"
		}
		result += a.focusModePrompt
		slog.Debug("applied focus mode prompt", "len", len(a.focusModePrompt))
	}

	// Then add output style prompt if set
	if a.outputStylePrompt != "" {
		if result != "" {
			result += "\n\n"
		}
		result += a.outputStylePrompt
		slog.Debug("applied output style prompt", "len", len(a.outputStylePrompt))
	}

	// Then add workspace memory context if set (Feature: Memory Hierarchy)
	if a.memoryContext != "" {
		if result != "" {
			result += "\n\n"
		}
		result += a.memoryContext
		slog.Default().Debug("[Agent] Applied memory context", "chars", len(a.memoryContext))
	}

	// Then add skills context if set (Agent Skills System)
	if a.skillsPrompt != "" {
		if result != "" {
			result += "\n\n"
		}
		result += a.skillsPrompt
		slog.Default().Debug("[Agent] Applied skills context", "chars", len(a.skillsPrompt))
	}

	// Now add the base system prompt
	if result != "" {
		result += "\n\n"
	}
	result += a.systemPrompt

	// Apply personalization AFTER base prompt but before thinking
	// This allows personalization to enhance but not override role context
	if a.promptPersonalizer != nil && a.userID != "" {
		ctx := context.Background()
		personalizedPrompt, err := a.promptPersonalizer.BuildPersonalizedPrompt(ctx, a.userID, result, a.lastUserMessage)
		if err == nil && personalizedPrompt != "" {
			result = personalizedPrompt
			slog.Default().Debug("[Agent] Applied prompt personalization",
				"total_chars", len(result))
		}
	}

	// Finally, add thinking instructions if enabled
	if a.llmOptions.ThinkingEnabled {
		// Use custom thinking instruction from template if provided, otherwise use default
		thinkingInstruction := prompts.ThinkingInstruction
		if a.llmOptions.ThinkingInstruction != "" {
			thinkingInstruction = a.llmOptions.ThinkingInstruction
		}
		slog.Debug("thinking enabled", "instruction_len", len(thinkingInstruction))
		return result + "\n\n" + thinkingInstruction
	}
	return result
}

// Pool returns the database pool
func (a *BaseAgent) Pool() *pgxpool.Pool {
	return a.pool
}

// Config returns the configuration
func (a *BaseAgent) Config() *config.Config {
	return a.cfg
}

// UserID returns the user ID
func (a *BaseAgent) UserID() string {
	return a.userID
}

// UserName returns the user name
func (a *BaseAgent) UserName() string {
	return a.userName
}

// ConversationID returns the conversation ID
func (a *BaseAgent) ConversationID() *uuid.UUID {
	return a.conversationID
}

// Model returns the current model
func (a *BaseAgent) Model() string {
	return a.model
}

// Agent constructors

// NewOrchestrator creates a new orchestrator agent
func NewOrchestrator(ctx *AgentContext) Agent {
	systemPrompt := prompts.ComposeWithUserContext(
		prompts_agents.OrchestratorAgentPrompt+prompts.ArtifactInstruction,
		ctx.UserName, "", "",
	)
	return NewBaseAgent(BaseAgentConfig{
		Pool:               ctx.Pool,
		Config:             ctx.Config,
		UserID:             ctx.UserID,
		UserName:           ctx.UserName,
		ConversationID:     ctx.ConversationID,
		EmbeddingService:   ctx.EmbeddingService,
		PromptPersonalizer: ctx.PromptPersonalizer,
		AgentType:          AgentTypeOrchestrator,
		AgentName:          "OSA Orchestrator",
		Description:        "Primary interface that handles general requests and routes to specialists",
		SystemPrompt:       systemPrompt,
		ContextReqs: ContextRequirements{
			NeedsProjects:    true,
			NeedsTasks:       true,
			NeedsClients:     true,
			NeedsKnowledge:   true,
			MaxContextTokens: 10000,
		},
		EnabledTools: []string{
			"search_documents", "get_project", "get_task", "get_client",
			"create_task", "create_project", "create_client",
			"create_artifact", "log_activity",
			"tree_search", "browse_tree", "load_context", // Context tools for knowledge base
		},
	})
}

// NewDocumentAgent creates a new document agent
func NewDocumentAgent(ctx *AgentContext) Agent {
	systemPrompt := prompts.DefaultComposer.ComposeForDocument(prompts_agents.DocumentAgentPrompt)
	return NewBaseAgent(BaseAgentConfig{
		Pool:               ctx.Pool,
		Config:             ctx.Config,
		UserID:             ctx.UserID,
		UserName:           ctx.UserName,
		ConversationID:     ctx.ConversationID,
		EmbeddingService:   ctx.EmbeddingService,
		PromptPersonalizer: ctx.PromptPersonalizer,
		AgentType:          AgentTypeDocument,
		AgentName:          "Document Specialist",
		Description:        "Creates formal business documents: proposals, SOPs, reports, frameworks",
		SystemPrompt:       systemPrompt,
		ContextReqs: ContextRequirements{
			NeedsProjects:    true,
			NeedsKnowledge:   true,
			NeedsClients:     true,
			MaxContextTokens: 10000,
		},
		EnabledTools: []string{
			"create_artifact", "search_documents", "get_project", "get_client",
			"log_activity",
			"tree_search", "browse_tree", "load_context", // Context tools for knowledge base
		},
	})
}

// NewProjectAgent creates a new project/planning agent
func NewProjectAgent(ctx *AgentContext) Agent {
	systemPrompt := prompts.Compose(prompts_agents.ProjectAgentPrompt)
	return NewBaseAgent(BaseAgentConfig{
		Pool:               ctx.Pool,
		Config:             ctx.Config,
		UserID:             ctx.UserID,
		UserName:           ctx.UserName,
		ConversationID:     ctx.ConversationID,
		EmbeddingService:   ctx.EmbeddingService,
		PromptPersonalizer: ctx.PromptPersonalizer,
		AgentType:          AgentTypeProject,
		AgentName:          "Project Specialist",
		Description:        "Project management and planning specialist",
		SystemPrompt:       systemPrompt,
		ContextReqs: ContextRequirements{
			NeedsProjects:    true,
			NeedsTasks:       true,
			NeedsTeam:        true,
			NeedsClients:     true,
			MaxContextTokens: 8000,
		},
		EnabledTools: []string{
			"create_project", "update_project", "get_project", "list_projects",
			"create_task", "bulk_create_tasks", "assign_task",
			"get_team_capacity", "search_documents",
			"create_artifact", "log_activity",
			"tree_search", "browse_tree", "load_context", // Context tools for knowledge base
		},
	})
}

// NewClientAgent creates a new client agent
func NewClientAgent(ctx *AgentContext) Agent {
	systemPrompt := prompts.Compose(prompts_agents.ClientAgentPrompt)
	return NewBaseAgent(BaseAgentConfig{
		Pool:               ctx.Pool,
		Config:             ctx.Config,
		UserID:             ctx.UserID,
		UserName:           ctx.UserName,
		ConversationID:     ctx.ConversationID,
		EmbeddingService:   ctx.EmbeddingService,
		PromptPersonalizer: ctx.PromptPersonalizer,
		AgentType:          AgentTypeClient,
		AgentName:          "Client Specialist",
		Description:        "Client relationship and pipeline specialist",
		SystemPrompt:       systemPrompt,
		ContextReqs: ContextRequirements{
			NeedsClients:     true,
			NeedsProjects:    true,
			NeedsKnowledge:   true,
			MaxContextTokens: 6000,
		},
		EnabledTools: []string{
			"create_client", "update_client", "get_client",
			"log_client_interaction", "update_client_pipeline",
			"search_documents", "get_project",
			"create_artifact", "log_activity",
			"tree_search", "browse_tree", "load_context", // Context tools for knowledge base
		},
	})
}

// NewAnalystAgent creates a new analyst agent
func NewAnalystAgent(ctx *AgentContext) Agent {
	systemPrompt := prompts.DefaultComposer.ComposeForAnalysis(prompts_agents.AnalystAgentPrompt)
	return NewBaseAgent(BaseAgentConfig{
		Pool:               ctx.Pool,
		Config:             ctx.Config,
		UserID:             ctx.UserID,
		UserName:           ctx.UserName,
		ConversationID:     ctx.ConversationID,
		EmbeddingService:   ctx.EmbeddingService,
		PromptPersonalizer: ctx.PromptPersonalizer,
		AgentType:          AgentTypeAnalyst,
		AgentName:          "Analyst Specialist",
		Description:        "Data analysis and insights specialist",
		SystemPrompt:       systemPrompt,
		ContextReqs: ContextRequirements{
			NeedsProjects:    true,
			NeedsTasks:       true,
			NeedsClients:     true,
			NeedsTeam:        true,
			MaxContextTokens: 8000,
		},
		EnabledTools: []string{
			"query_metrics", "get_team_capacity",
			"list_projects", "list_tasks", "get_project",
			"search_documents", "create_artifact",
			"log_activity",
			"tree_search", "browse_tree", "load_context", // Context tools for knowledge base
		},
	})
}

// NewTaskAgent creates a new task management agent
func NewTaskAgent(ctx *AgentContext) Agent {
	systemPrompt := prompts.Compose(prompts_agents.TaskAgentPrompt)
	return NewBaseAgent(BaseAgentConfig{
		Pool:               ctx.Pool,
		Config:             ctx.Config,
		UserID:             ctx.UserID,
		UserName:           ctx.UserName,
		ConversationID:     ctx.ConversationID,
		EmbeddingService:   ctx.EmbeddingService,
		PromptPersonalizer: ctx.PromptPersonalizer,
		AgentType:          AgentTypeTask,
		AgentName:          "Task Specialist",
		Description:        "Task management, prioritization, scheduling, and dependencies",
		SystemPrompt:       systemPrompt,
		ContextReqs: ContextRequirements{
			NeedsProjects:    true,
			NeedsTasks:       true,
			NeedsTeam:        true,
			MaxContextTokens: 8000,
		},
		EnabledTools: []string{
			"create_task", "update_task", "get_task", "list_tasks",
			"bulk_create_tasks", "move_task", "assign_task",
			"get_team_capacity", "get_project",
			"log_activity",
			"tree_search", "browse_tree", "load_context", // Context tools for knowledge base
		},
	})
}
