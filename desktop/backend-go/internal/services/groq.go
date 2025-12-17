package services

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rhl/businessos-backend/internal/config"
)

// GroqService handles LLM inference via Groq API
type GroqService struct {
	apiKey  string
	model   string
	client  *http.Client
	options LLMOptions
}

// GroqMessage represents a message in the Groq format
type GroqMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// GroqRequest represents a request to the Groq API
type GroqRequest struct {
	Model       string        `json:"model"`
	Messages    []GroqMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Stream      bool          `json:"stream"`
}

// GroqResponse represents a non-streaming response from Groq
type GroqResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// GroqStreamResponse represents a streaming chunk from Groq
type GroqStreamResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

// NewGroqService creates a new Groq service instance
func NewGroqService(cfg *config.Config, model string) *GroqService {
	if model == "" {
		model = cfg.GroqModel
	}
	if model == "" {
		model = "llama-3.3-70b-versatile" // Default Groq model
	}

	return &GroqService{
		apiKey: cfg.GroqAPIKey,
		model:  model,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
		options: DefaultLLMOptions(),
	}
}

// SetOptions sets the LLM options for this service
func (s *GroqService) SetOptions(opts LLMOptions) {
	s.options = opts
}

// GetOptions returns the current LLM options
func (s *GroqService) GetOptions() LLMOptions {
	return s.options
}

// StreamChat sends a chat request and streams the response
func (s *GroqService) StreamChat(ctx context.Context, messages []ChatMessage, systemPrompt string) (<-chan string, <-chan error) {
	chunks := make(chan string, 100)
	errs := make(chan error, 1)

	go func() {
		defer close(chunks)
		defer close(errs)

		// Convert messages to Groq format
		groqMsgs := make([]GroqMessage, 0, len(messages)+1)

		// Add system message first if provided
		if systemPrompt != "" {
			groqMsgs = append(groqMsgs, GroqMessage{
				Role:    "system",
				Content: systemPrompt,
			})
		}

		for _, msg := range messages {
			if msg.Role == "system" {
				// Combine with existing system prompt
				continue
			}
			groqMsgs = append(groqMsgs, GroqMessage{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}

		reqBody := GroqRequest{
			Model:     s.model,
			Messages:  groqMsgs,
			MaxTokens: 8192,
			Stream:    true,
		}

		body, err := json.Marshal(reqBody)
		if err != nil {
			errs <- fmt.Errorf("failed to marshal request: %w", err)
			return
		}

		req, err := http.NewRequestWithContext(ctx, "POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewReader(body))
		if err != nil {
			errs <- fmt.Errorf("failed to create request: %w", err)
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+s.apiKey)

		resp, err := s.client.Do(req)
		if err != nil {
			errs <- fmt.Errorf("failed to send request: %w", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			errs <- fmt.Errorf("groq API error: %s - %s", resp.Status, string(body))
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" || !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				return
			}

			var streamResp GroqStreamResponse
			if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
				continue // Skip malformed lines
			}

			if len(streamResp.Choices) > 0 && streamResp.Choices[0].Delta.Content != "" {
				select {
				case chunks <- streamResp.Choices[0].Delta.Content:
				case <-ctx.Done():
					return
				}
			}
		}

		if err := scanner.Err(); err != nil {
			errs <- fmt.Errorf("error reading response: %w", err)
		}
	}()

	return chunks, errs
}

// ChatComplete sends a non-streaming chat request
func (s *GroqService) ChatComplete(ctx context.Context, messages []ChatMessage, systemPrompt string) (string, error) {
	// Convert messages to Groq format
	groqMsgs := make([]GroqMessage, 0, len(messages)+1)

	if systemPrompt != "" {
		groqMsgs = append(groqMsgs, GroqMessage{
			Role:    "system",
			Content: systemPrompt,
		})
	}

	for _, msg := range messages {
		if msg.Role == "system" {
			continue
		}
		groqMsgs = append(groqMsgs, GroqMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	reqBody := GroqRequest{
		Model:     s.model,
		Messages:  groqMsgs,
		MaxTokens: 8192,
		Stream:    false,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("groq API error: %s - %s", resp.Status, string(body))
	}

	var groqResp GroqResponse
	if err := json.NewDecoder(resp.Body).Decode(&groqResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(groqResp.Choices) == 0 {
		return "", fmt.Errorf("no response from Groq")
	}

	return groqResp.Choices[0].Message.Content, nil
}

// HealthCheck checks if Groq API is available
func (s *GroqService) HealthCheck(ctx context.Context) bool {
	return s.apiKey != ""
}

// GetModel returns the model name
func (s *GroqService) GetModel() string {
	return s.model
}

// GetProvider returns the provider name
func (s *GroqService) GetProvider() string {
	return "groq"
}

// StreamChatWithUsage streams chat and tracks token usage
func (s *GroqService) StreamChatWithUsage(ctx context.Context, messages []ChatMessage, systemPrompt string) *StreamResult {
	chunks := make(chan string, 100)
	errs := make(chan error, 1)
	result := &StreamResult{
		Chunks: chunks,
		Errors: errs,
	}

	go func() {
		defer close(chunks)
		defer close(errs)

		groqMsgs := make([]GroqMessage, 0, len(messages)+1)
		if systemPrompt != "" {
			groqMsgs = append(groqMsgs, GroqMessage{
				Role:    "system",
				Content: systemPrompt,
			})
		}

		for _, msg := range messages {
			if msg.Role == "system" {
				continue
			}
			groqMsgs = append(groqMsgs, GroqMessage{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}

		reqBody := GroqRequest{
			Model:     s.model,
			Messages:  groqMsgs,
			MaxTokens: 8192,
			Stream:    true,
		}

		body, err := json.Marshal(reqBody)
		if err != nil {
			errs <- fmt.Errorf("failed to marshal request: %w", err)
			return
		}

		req, err := http.NewRequestWithContext(ctx, "POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewReader(body))
		if err != nil {
			errs <- fmt.Errorf("failed to create request: %w", err)
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+s.apiKey)

		resp, err := s.client.Do(req)
		if err != nil {
			errs <- fmt.Errorf("failed to send request: %w", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			errs <- fmt.Errorf("groq API error: %s - %s", resp.Status, string(body))
			return
		}

		var estimatedTokens int
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" || !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				break
			}

			var streamResp GroqStreamResponse
			if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
				continue
			}

			if len(streamResp.Choices) > 0 && streamResp.Choices[0].Delta.Content != "" {
				content := streamResp.Choices[0].Delta.Content
				estimatedTokens += len(content) / 4
				select {
				case chunks <- content:
				case <-ctx.Done():
					return
				}
			}
		}

		if err := scanner.Err(); err != nil {
			errs <- fmt.Errorf("error reading response: %w", err)
		}

		result.SetTokenUsage(&TokenUsage{
			InputTokens:  0,
			OutputTokens: estimatedTokens,
			TotalTokens:  estimatedTokens,
			Model:        s.model,
			Provider:     "groq",
		})
	}()

	return result
}

// ChatCompleteWithUsage sends a non-streaming chat request and returns token usage
func (s *GroqService) ChatCompleteWithUsage(ctx context.Context, messages []ChatMessage, systemPrompt string) (string, *TokenUsage, error) {
	groqMsgs := make([]GroqMessage, 0, len(messages)+1)

	if systemPrompt != "" {
		groqMsgs = append(groqMsgs, GroqMessage{
			Role:    "system",
			Content: systemPrompt,
		})
	}

	for _, msg := range messages {
		if msg.Role == "system" {
			continue
		}
		groqMsgs = append(groqMsgs, GroqMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	reqBody := GroqRequest{
		Model:     s.model,
		Messages:  groqMsgs,
		MaxTokens: 8192,
		Stream:    false,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", nil, fmt.Errorf("groq API error: %s - %s", resp.Status, string(body))
	}

	var groqResp GroqResponse
	if err := json.NewDecoder(resp.Body).Decode(&groqResp); err != nil {
		return "", nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(groqResp.Choices) == 0 {
		return "", nil, fmt.Errorf("no response from Groq")
	}

	usage := &TokenUsage{
		InputTokens:  groqResp.Usage.PromptTokens,
		OutputTokens: groqResp.Usage.CompletionTokens,
		TotalTokens:  groqResp.Usage.TotalTokens,
		Model:        s.model,
		Provider:     "groq",
	}

	return groqResp.Choices[0].Message.Content, usage, nil
}
