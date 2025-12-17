package services

import (
	"context"
	"sync"

	"github.com/rhl/businessos-backend/internal/config"
)

// LLMOptions holds configurable parameters for LLM calls
type LLMOptions struct {
	Temperature float64
	MaxTokens   int
	TopP        float64
}

// DefaultLLMOptions returns sensible defaults
func DefaultLLMOptions() LLMOptions {
	return LLMOptions{
		Temperature: 0.7,
		MaxTokens:   8192,
		TopP:        0.9,
	}
}

// TokenUsage tracks token consumption for LLM calls
type TokenUsage struct {
	InputTokens  int    `json:"input_tokens"`
	OutputTokens int    `json:"output_tokens"`
	TotalTokens  int    `json:"total_tokens"`
	Model        string `json:"model"`
	Provider     string `json:"provider"`
}

// AgentTrace tracks the flow of an agent request
type AgentTrace struct {
	AgentName     string       `json:"agent_name"`
	DelegatedTo   string       `json:"delegated_to,omitempty"`
	TokenUsage    *TokenUsage  `json:"token_usage,omitempty"`
	SubTraces     []AgentTrace `json:"sub_traces,omitempty"`
	StartTime     int64        `json:"start_time"`
	EndTime       int64        `json:"end_time,omitempty"`
	DurationMs    int64        `json:"duration_ms,omitempty"`
}

// StreamResult wraps streaming results with token usage
type StreamResult struct {
	Chunks     <-chan string
	Errors     <-chan error
	TokenUsage *TokenUsage
	mu         sync.Mutex
}

// SetTokenUsage safely sets token usage (called when stream completes)
func (sr *StreamResult) SetTokenUsage(usage *TokenUsage) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	sr.TokenUsage = usage
}

// GetTokenUsage safely gets token usage
func (sr *StreamResult) GetTokenUsage() *TokenUsage {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	return sr.TokenUsage
}

// LLMService interface for language model services
type LLMService interface {
	StreamChat(ctx context.Context, messages []ChatMessage, systemPrompt string) (<-chan string, <-chan error)
	StreamChatWithUsage(ctx context.Context, messages []ChatMessage, systemPrompt string) *StreamResult
	ChatComplete(ctx context.Context, messages []ChatMessage, systemPrompt string) (string, error)
	ChatCompleteWithUsage(ctx context.Context, messages []ChatMessage, systemPrompt string) (string, *TokenUsage, error)
	HealthCheck(ctx context.Context) bool
	GetModel() string
	GetProvider() string
	SetOptions(opts LLMOptions)
	GetOptions() LLMOptions
}

// NewLLMService creates the appropriate LLM service based on configuration
func NewLLMService(cfg *config.Config, model string) LLMService {
	switch cfg.GetActiveProvider() {
	case "ollama_cloud":
		return NewOllamaCloudService(cfg, model)
	case "anthropic":
		return NewAnthropicService(cfg, model)
	case "groq":
		return NewGroqService(cfg, model)
	default:
		// Default to local Ollama
		return NewOllamaService(cfg, model)
	}
}
