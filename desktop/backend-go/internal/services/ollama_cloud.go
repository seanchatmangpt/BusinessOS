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

// OllamaCloudService handles LLM inference via Ollama Cloud API
type OllamaCloudService struct {
	apiKey  string
	model   string
	client  *http.Client
	options LLMOptions
}

const ollamaCloudBaseURL = "https://api.ollama.com/v1"

// OllamaCloudMessage represents a message in the OpenAI-compatible format
type OllamaCloudMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OllamaCloudRequest represents a request to Ollama Cloud
type OllamaCloudRequest struct {
	Model     string               `json:"model"`
	Messages  []OllamaCloudMessage `json:"messages"`
	MaxTokens int                  `json:"max_tokens,omitempty"`
	Stream    bool                 `json:"stream"`
}

// OllamaCloudResponse represents a non-streaming response
type OllamaCloudResponse struct {
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

// OllamaCloudStreamResponse represents a streaming chunk
type OllamaCloudStreamResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

// NewOllamaCloudService creates a new Ollama Cloud service instance
func NewOllamaCloudService(cfg *config.Config, model string) *OllamaCloudService {
	if model == "" {
		model = cfg.OllamaCloudModel
	}
	if model == "" {
		model = "llama3.2" // Default model
	}

	return &OllamaCloudService{
		apiKey: cfg.OllamaCloudAPIKey,
		model:  model,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
		options: DefaultLLMOptions(),
	}
}

// SetOptions sets the LLM options for this service
func (s *OllamaCloudService) SetOptions(opts LLMOptions) {
	s.options = opts
}

// GetOptions returns the current LLM options
func (s *OllamaCloudService) GetOptions() LLMOptions {
	return s.options
}

// StreamChat sends a chat request and streams the response
func (s *OllamaCloudService) StreamChat(ctx context.Context, messages []ChatMessage, systemPrompt string) (<-chan string, <-chan error) {
	chunks := make(chan string, 100)
	errs := make(chan error, 1)

	go func() {
		defer close(chunks)
		defer close(errs)

		// Convert messages to OpenAI format
		cloudMsgs := make([]OllamaCloudMessage, 0, len(messages)+1)

		for _, msg := range messages {
			if msg.Role == "system" {
				if systemPrompt != "" {
					systemPrompt = systemPrompt + "\n\n" + msg.Content
				} else {
					systemPrompt = msg.Content
				}
				continue
			}
			cloudMsgs = append(cloudMsgs, OllamaCloudMessage{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}

		// Add system message first if provided/combined
		if systemPrompt != "" {
			cloudMsgs = append([]OllamaCloudMessage{{
				Role:    "system",
				Content: systemPrompt,
			}}, cloudMsgs...)
		}

		reqBody := OllamaCloudRequest{
			Model:     s.model,
			Messages:  cloudMsgs,
			MaxTokens: 8192,
			Stream:    true,
		}

		body, err := json.Marshal(reqBody)
		if err != nil {
			errs <- fmt.Errorf("failed to marshal request: %w", err)
			return
		}

		req, err := http.NewRequestWithContext(ctx, "POST", ollamaCloudBaseURL+"/chat/completions", bytes.NewReader(body))
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
			errs <- fmt.Errorf("ollama cloud API error: %s - %s", resp.Status, string(body))
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

			var streamResp OllamaCloudStreamResponse
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
func (s *OllamaCloudService) ChatComplete(ctx context.Context, messages []ChatMessage, systemPrompt string) (string, error) {
	cloudMsgs := make([]OllamaCloudMessage, 0, len(messages)+1)

	for _, msg := range messages {
		if msg.Role == "system" {
			if systemPrompt != "" {
				systemPrompt = systemPrompt + "\n\n" + msg.Content
			} else {
				systemPrompt = msg.Content
			}
			continue
		}
		cloudMsgs = append(cloudMsgs, OllamaCloudMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	if systemPrompt != "" {
		cloudMsgs = append([]OllamaCloudMessage{{
			Role:    "system",
			Content: systemPrompt,
		}}, cloudMsgs...)
	}

	reqBody := OllamaCloudRequest{
		Model:     s.model,
		Messages:  cloudMsgs,
		MaxTokens: 8192,
		Stream:    false,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", ollamaCloudBaseURL+"/chat/completions", bytes.NewReader(body))
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
		return "", fmt.Errorf("ollama cloud API error: %s - %s", resp.Status, string(body))
	}

	var cloudResp OllamaCloudResponse
	if err := json.NewDecoder(resp.Body).Decode(&cloudResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(cloudResp.Choices) == 0 {
		return "", fmt.Errorf("no response from Ollama Cloud")
	}

	return cloudResp.Choices[0].Message.Content, nil
}

// HealthCheck checks if Ollama Cloud API is available
func (s *OllamaCloudService) HealthCheck(ctx context.Context) bool {
	return s.apiKey != ""
}

// GetModel returns the model name
func (s *OllamaCloudService) GetModel() string {
	return s.model
}

// GetProvider returns the provider name
func (s *OllamaCloudService) GetProvider() string {
	return "ollama_cloud"
}

// StreamChatWithUsage streams chat and tracks token usage
func (s *OllamaCloudService) StreamChatWithUsage(ctx context.Context, messages []ChatMessage, systemPrompt string) *StreamResult {
	chunks := make(chan string, 100)
	errs := make(chan error, 1)
	result := &StreamResult{
		Chunks: chunks,
		Errors: errs,
	}

	go func() {
		defer close(chunks)
		defer close(errs)

		cloudMsgs := make([]OllamaCloudMessage, 0, len(messages)+1)

		for _, msg := range messages {
			if msg.Role == "system" {
				if systemPrompt != "" {
					systemPrompt = systemPrompt + "\n\n" + msg.Content
				} else {
					systemPrompt = msg.Content
				}
				continue
			}
			cloudMsgs = append(cloudMsgs, OllamaCloudMessage{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}

		if systemPrompt != "" {
			cloudMsgs = append([]OllamaCloudMessage{{
				Role:    "system",
				Content: systemPrompt,
			}}, cloudMsgs...)
		}

		reqBody := OllamaCloudRequest{
			Model:     s.model,
			Messages:  cloudMsgs,
			MaxTokens: 8192,
			Stream:    true,
		}

		body, err := json.Marshal(reqBody)
		if err != nil {
			errs <- fmt.Errorf("failed to marshal request: %w", err)
			return
		}

		req, err := http.NewRequestWithContext(ctx, "POST", ollamaCloudBaseURL+"/chat/completions", bytes.NewReader(body))
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
			errs <- fmt.Errorf("ollama cloud API error: %s - %s", resp.Status, string(body))
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

			var streamResp OllamaCloudStreamResponse
			if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
				continue
			}

			if len(streamResp.Choices) > 0 && streamResp.Choices[0].Delta.Content != "" {
				content := streamResp.Choices[0].Delta.Content
				estimatedTokens += len(content) / 4 // Estimate tokens
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

		// Estimate token usage (streaming doesn't return actual counts)
		result.SetTokenUsage(&TokenUsage{
			InputTokens:  0, // Not available in streaming
			OutputTokens: estimatedTokens,
			TotalTokens:  estimatedTokens,
			Model:        s.model,
			Provider:     "ollama_cloud",
		})
	}()

	return result
}

// ChatCompleteWithUsage sends a non-streaming chat request and returns token usage
func (s *OllamaCloudService) ChatCompleteWithUsage(ctx context.Context, messages []ChatMessage, systemPrompt string) (string, *TokenUsage, error) {
	cloudMsgs := make([]OllamaCloudMessage, 0, len(messages)+1)

	for _, msg := range messages {
		if msg.Role == "system" {
			if systemPrompt != "" {
				systemPrompt = systemPrompt + "\n\n" + msg.Content
			} else {
				systemPrompt = msg.Content
			}
			continue
		}
		cloudMsgs = append(cloudMsgs, OllamaCloudMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	if systemPrompt != "" {
		cloudMsgs = append([]OllamaCloudMessage{{
			Role:    "system",
			Content: systemPrompt,
		}}, cloudMsgs...)
	}

	reqBody := OllamaCloudRequest{
		Model:     s.model,
		Messages:  cloudMsgs,
		MaxTokens: 8192,
		Stream:    false,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", ollamaCloudBaseURL+"/chat/completions", bytes.NewReader(body))
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
		return "", nil, fmt.Errorf("ollama cloud API error: %s - %s", resp.Status, string(body))
	}

	var cloudResp OllamaCloudResponse
	if err := json.NewDecoder(resp.Body).Decode(&cloudResp); err != nil {
		return "", nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(cloudResp.Choices) == 0 {
		return "", nil, fmt.Errorf("no response from Ollama Cloud")
	}

	usage := &TokenUsage{
		InputTokens:  cloudResp.Usage.PromptTokens,
		OutputTokens: cloudResp.Usage.CompletionTokens,
		TotalTokens:  cloudResp.Usage.TotalTokens,
		Model:        s.model,
		Provider:     "ollama_cloud",
	}

	return cloudResp.Choices[0].Message.Content, usage, nil
}
