package agents

import (
	"context"

	"github.com/google/uuid"
	"github.com/rhl/businessos-backend/internal/services"
	"github.com/rhl/businessos-backend/internal/streaming"
)

// VoiceAgentAdapter adapts AgentRegistryV2 to implement services.VoiceAgentProvider
// This allows the voice system to use Agent V2 without creating an import cycle
type VoiceAgentAdapter struct {
	registry *AgentRegistryV2
}

// NewVoiceAgentAdapter creates a new voice agent adapter
func NewVoiceAgentAdapter(registry *AgentRegistryV2) *VoiceAgentAdapter {
	return &VoiceAgentAdapter{
		registry: registry,
	}
}

// ExecuteVoiceAgent implements services.VoiceAgentProvider
// Executes an agent and returns streaming response for voice interactions
func (v *VoiceAgentAdapter) ExecuteVoiceAgent(
	ctx context.Context,
	userID string,
	userName string,
	conversationID *uuid.UUID,
	messages []services.ChatMessage,
	tieredContext *services.TieredContext,
	llmOptions services.LLMOptions,
) (<-chan streaming.StreamEvent, <-chan error) {
	// Get agent (use orchestrator for conversational responses)
	// Orchestrator is best for general conversation and can delegate to specialists if needed
	agent := v.registry.GetAgent(
		AgentTypeV2Orchestrator, // Orchestrator agent for general conversation
		userID,
		userName,
		conversationID,
		tieredContext,
	)

	// Create agent input
	input := AgentInput{
		Messages:      messages,
		Context:       tieredContext,
		UserID:        userID,
		UserName:      userName,
		Selections:    UserSelections{}, // Empty selections for voice
		FocusMode:     "",               // No focus mode for voice
		FocusModeOpts: nil,
	}

	// If conversationID provided, set it
	if conversationID != nil {
		input.ConversationID = *conversationID
	}

	// Set LLM options on the agent
	agent.SetOptions(llmOptions)

	// Execute agent and return channels
	// Use Run() for basic execution (RunWithTools() would enable tool use)
	return agent.Run(ctx, input)
}
