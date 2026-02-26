package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/rhl/businessos-backend/internal/services"
)

// ContextInjector handles injecting memories and documents into chat context
type ContextInjector struct {
	contextBuilder  *services.ContextBuilder
	documentService *services.DocumentProcessor
	memoryHandler   *MemoryHandler
}

// NewContextInjector creates a new context injector
func NewContextInjector(
	contextBuilder *services.ContextBuilder,
	documentService *services.DocumentProcessor,
	memoryHandler *MemoryHandler,
) *ContextInjector {
	return &ContextInjector{
		contextBuilder:  contextBuilder,
		documentService: documentService,
		memoryHandler:   memoryHandler,
	}
}

// InjectedContext represents context to inject into the system prompt
type InjectedContext struct {
	Memories       []MemoryContext   `json:"memories,omitempty"`
	Documents      []DocumentContext `json:"documents,omitempty"`
	RelevantBlocks []BlockContext    `json:"relevant_blocks,omitempty"`
	UserFacts      []string          `json:"user_facts,omitempty"`
	FormattedText  string            `json:"formatted_text"`
	TotalTokens    int               `json:"total_tokens"`
}

// MemoryContext represents a memory for context injection
type MemoryContext struct {
	ID      string  `json:"id"`
	Type    string  `json:"type"`
	Content string  `json:"content"`
	Score   float64 `json:"score,omitempty"`
}

// DocumentContext represents a document chunk for context injection
type DocumentContext struct {
	ID       string  `json:"id"`
	Filename string  `json:"filename"`
	Content  string  `json:"content"`
	Score    float64 `json:"score,omitempty"`
}

// BlockContext represents a relevant block for context injection
type BlockContext struct {
	ContextID  string  `json:"context_id"`
	DocName    string  `json:"doc_name"`
	Content    string  `json:"content"`
	BlockType  string  `json:"block_type"`
	Similarity float64 `json:"similarity"`
}

// ContextInjectionRequest represents a request for context injection
type ContextInjectionRequest struct {
	UserID      string
	Query       string
	ProjectID   *uuid.UUID
	NodeID      *uuid.UUID
	MaxMemories int
	MaxDocs     int
	MaxBlocks   int
}

// BuildInjectedContext builds context to inject based on the query
func (ci *ContextInjector) BuildInjectedContext(ctx context.Context, req ContextInjectionRequest) (*InjectedContext, error) {
	if req.MaxMemories <= 0 {
		req.MaxMemories = 5
	}
	if req.MaxDocs <= 0 {
		req.MaxDocs = 3
	}
	if req.MaxBlocks <= 0 {
		req.MaxBlocks = 5
	}

	result := &InjectedContext{
		Memories:       make([]MemoryContext, 0),
		Documents:      make([]DocumentContext, 0),
		RelevantBlocks: make([]BlockContext, 0),
		UserFacts:      make([]string, 0),
	}

	// Build hierarchical context using ContextBuilder
	if ci.contextBuilder != nil && req.Query != "" {
		hCtx, err := ci.contextBuilder.BuildContext(ctx, req.Query, req.UserID, req.MaxBlocks)
		if err == nil && hCtx != nil {
			for _, block := range hCtx.RelevantBlocks {
				result.RelevantBlocks = append(result.RelevantBlocks, BlockContext{
					ContextID:  block.ContextID.String(),
					DocName:    block.DocumentName,
					Content:    block.BlockContent,
					BlockType:  block.BlockType,
					Similarity: block.Similarity,
				})
			}
		}
	}

	// Get relevant document chunks
	if ci.documentService != nil && req.Query != "" {
		chunks, err := ci.documentService.GetRelevantChunks(ctx, req.UserID, req.Query, req.MaxDocs)
		if err == nil {
			for _, chunk := range chunks {
				result.Documents = append(result.Documents, DocumentContext{
					ID:       chunk.ID.String(),
					Filename: chunk.DocumentID.String(),
					Content:  chunk.Content,
					Score:    0, // DocumentChunk doesn't have Score field
				})
			}
		}
	}

	// Format all context into text
	result.FormattedText = ci.formatContext(result)
	result.TotalTokens = estimateTokens(result.FormattedText)

	return result, nil
}

// formatContext formats the injected context into a system prompt section
func (ci *ContextInjector) formatContext(ctx *InjectedContext) string {
	var sb strings.Builder

	// Check if we have any context at all
	hasContext := len(ctx.UserFacts) > 0 || len(ctx.RelevantBlocks) > 0 ||
		len(ctx.Documents) > 0 || len(ctx.Memories) > 0

	if !hasContext {
		sb.WriteString("\n## Knowledge Base Status\n")
		sb.WriteString("The user currently has NO saved documents, memories, or context in their knowledge base.\n")
		sb.WriteString("Do NOT invent or assume any content about their projects or documents.\n")
		return sb.String()
	}

	// Add user facts
	if len(ctx.UserFacts) > 0 {
		sb.WriteString("\n## User Information\n")
		for _, fact := range ctx.UserFacts {
			sb.WriteString(fmt.Sprintf("- %s\n", fact))
		}
	}

	// Add relevant context blocks
	if len(ctx.RelevantBlocks) > 0 {
		sb.WriteString("\n## Relevant Context\n")
		for i, block := range ctx.RelevantBlocks {
			sb.WriteString(fmt.Sprintf("### Context %d: %s\n", i+1, block.DocName))
			sb.WriteString(block.Content)
			sb.WriteString("\n\n")
		}
	}

	// Add document chunks
	if len(ctx.Documents) > 0 {
		sb.WriteString("\n## Relevant Documents\n")
		for i, doc := range ctx.Documents {
			sb.WriteString(fmt.Sprintf("### Document %d: %s\n", i+1, doc.Filename))
			sb.WriteString(doc.Content)
			sb.WriteString("\n\n")
		}
	}

	// Add memories
	if len(ctx.Memories) > 0 {
		sb.WriteString("\n## Relevant Memories\n")
		for _, mem := range ctx.Memories {
			sb.WriteString(fmt.Sprintf("- [%s] %s\n", mem.Type, mem.Content))
		}
	}

	return sb.String()
}

// InjectIntoSystemPrompt injects context into a system prompt
func (ci *ContextInjector) InjectIntoSystemPrompt(systemPrompt string, injectedCtx *InjectedContext) string {
	if injectedCtx == nil {
		return systemPrompt
	}
	// Note: We still inject even if FormattedText is about empty knowledge base

	// Insert context before the main instructions
	contextSection := fmt.Sprintf(`
<injected_context>
The following context has been automatically retrieved based on the user's query:
%s
</injected_context>

`, injectedCtx.FormattedText)

	// If system prompt is empty, just return the context
	if systemPrompt == "" {
		return contextSection
	}

	// Insert context at the beginning of the system prompt
	return contextSection + systemPrompt
}

// estimateTokens provides a rough token estimate
func estimateTokens(text string) int {
	// Rough estimation: ~4 characters per token
	return len(text) / 4
}

// ShouldInjectContext determines if context injection should be performed
func ShouldInjectContext(query string, focusMode string) bool {
	// Always inject for certain focus modes
	if focusMode == "research" || focusMode == "deep-research" || focusMode == "writing" {
		return true
	}

	// Inject for longer queries (likely need more context)
	if len(query) > 100 {
		return true
	}

	// Check for context-requesting keywords
	contextKeywords := []string{
		"remember", "recall", "previous", "earlier", "before",
		"document", "file", "code", "project", "context",
		"you said", "we discussed", "last time",
	}

	queryLower := strings.ToLower(query)
	for _, keyword := range contextKeywords {
		if strings.Contains(queryLower, keyword) {
			return true
		}
	}

	return false
}
