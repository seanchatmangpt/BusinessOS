package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// SmartChunkingService implements intelligent document chunking for RAG optimization
// Implements SORX 2.0 Smart Chunking SKILL
type SmartChunkingService struct {
	// Token estimator (approximation: ~4 chars per token for English)
	charsPerToken float64
}

// Chunk represents a document chunk with metadata
type Chunk struct {
	ID          string                 `json:"id"`
	Content     string                 `json:"content"`
	TokenCount  int                    `json:"token_count"`
	Position    int                    `json:"position"`
	Metadata    map[string]interface{} `json:"metadata"`
	ParentDocID string                 `json:"parent_doc_id"`
}

// ChunkOptions configures chunking behavior
type ChunkOptions struct {
	ChunkSize    int     // Target chunk size in tokens (default: 512)
	OverlapRatio float64 // Overlap ratio between chunks (default: 0.2 = 20%)
	PreserveCode bool    // Whether to preserve code block boundaries (default: true)
	MinChunkSize int     // Minimum chunk size in tokens (default: 128)
}

// DocumentType represents the type of document being chunked
type DocumentType string

const (
	DocTypeMarkdown  DocumentType = "markdown"
	DocTypeCode      DocumentType = "code"
	DocTypePlainText DocumentType = "plaintext"
	DocTypeJSON      DocumentType = "json"
	DocTypeXML       DocumentType = "xml"
)

// NewSmartChunkingService creates a new smart chunking service
func NewSmartChunkingService() *SmartChunkingService {
	return &SmartChunkingService{
		charsPerToken: 4.0, // Approximate average for English text
	}
}

// DefaultChunkOptions returns sensible defaults
func DefaultChunkOptions() ChunkOptions {
	return ChunkOptions{
		ChunkSize:    512,
		OverlapRatio: 0.2,
		PreserveCode: true,
		MinChunkSize: 128,
	}
}

// ChunkDocument intelligently chunks a document based on its type
func (s *SmartChunkingService) ChunkDocument(ctx context.Context, content string, docType DocumentType, parentDocID string, options ChunkOptions) ([]Chunk, error) {
	if content == "" {
		return nil, fmt.Errorf("empty content for chunking")
	}

	// Apply defaults
	if options.ChunkSize == 0 {
		options.ChunkSize = 512
	}
	if options.OverlapRatio == 0 {
		options.OverlapRatio = 0.2
	}
	if options.MinChunkSize == 0 {
		options.MinChunkSize = 128
	}

	// Validate options
	if options.ChunkSize < options.MinChunkSize {
		return nil, fmt.Errorf("chunk_size (%d) must be >= min_chunk_size (%d)", options.ChunkSize, options.MinChunkSize)
	}
	if options.OverlapRatio < 0 || options.OverlapRatio >= 1 {
		return nil, fmt.Errorf("overlap_ratio must be between 0 and 1")
	}

	// Route to appropriate chunking strategy
	switch docType {
	case DocTypeMarkdown:
		return s.ChunkMarkdown(ctx, content, parentDocID, options)
	case DocTypeCode:
		return s.ChunkCode(ctx, content, "", parentDocID, options) // Language detection happens inside
	case DocTypePlainText:
		return s.ChunkPlainText(ctx, content, parentDocID, options)
	case DocTypeJSON, DocTypeXML:
		return s.ChunkStructured(ctx, content, string(docType), parentDocID, options)
	default:
		// Default to plain text chunking
		return s.ChunkPlainText(ctx, content, parentDocID, options)
	}
}

// ChunkMarkdown chunks markdown content respecting structure
func (s *SmartChunkingService) ChunkMarkdown(ctx context.Context, content string, parentDocID string, options ChunkOptions) ([]Chunk, error) {
	var chunks []Chunk

	// Split by markdown sections (headers)
	sections := s.splitMarkdownSections(content)

	position := 0
	overlapSize := int(float64(options.ChunkSize) * options.OverlapRatio)

	for sectionIdx, section := range sections {
		// Check if section fits in one chunk
		sectionTokens := s.estimateTokens(section.Content)

		if sectionTokens <= options.ChunkSize {
			// Section fits in one chunk
			chunk := s.createChunk(section.Content, position, parentDocID, map[string]interface{}{
				"type":        "markdown_section",
				"heading":     section.Heading,
				"level":       section.Level,
				"section_idx": sectionIdx,
			})
			chunks = append(chunks, chunk)
			position++
		} else {
			// Section too large, split by paragraphs
			paragraphs := s.splitParagraphs(section.Content)
			subChunks := s.chunkParagraphs(paragraphs, position, parentDocID, options, map[string]interface{}{
				"type":        "markdown_section",
				"heading":     section.Heading,
				"level":       section.Level,
				"section_idx": sectionIdx,
			})
			chunks = append(chunks, subChunks...)
			position += len(subChunks)
		}
	}

	// Apply overlap between chunks
	if options.OverlapRatio > 0 && len(chunks) > 1 {
		chunks = s.applyOverlap(chunks, overlapSize)
	}

	return chunks, nil
}

// ChunkCode chunks code content respecting language syntax
func (s *SmartChunkingService) ChunkCode(ctx context.Context, content string, language string, parentDocID string, options ChunkOptions) ([]Chunk, error) {
	var chunks []Chunk

	// If language not specified, try to detect
	if language == "" {
		language = s.detectLanguage(content)
	}

	// Split by code blocks (functions, classes, etc.)
	codeBlocks := s.splitCodeBlocks(content, language)

	position := 0
	overlapSize := int(float64(options.ChunkSize) * options.OverlapRatio)

	for blockIdx, block := range codeBlocks {
		blockTokens := s.estimateTokens(block.Content)

		if blockTokens <= options.ChunkSize {
			// Block fits in one chunk
			chunk := s.createChunk(block.Content, position, parentDocID, map[string]interface{}{
				"type":      "code_block",
				"language":  language,
				"block_type": block.Type,
				"name":      block.Name,
				"block_idx": blockIdx,
			})
			chunks = append(chunks, chunk)
			position++
		} else {
			// Block too large, split by lines while preserving context
			subChunks := s.chunkCodeLines(block.Content, position, parentDocID, options, map[string]interface{}{
				"type":      "code_block",
				"language":  language,
				"block_type": block.Type,
				"name":      block.Name,
				"block_idx": blockIdx,
			})
			chunks = append(chunks, subChunks...)
			position += len(subChunks)
		}
	}

	// Apply overlap between chunks
	if options.OverlapRatio > 0 && len(chunks) > 1 {
		chunks = s.applyOverlap(chunks, overlapSize)
	}

	return chunks, nil
}

// ChunkPlainText chunks plain text respecting natural boundaries
func (s *SmartChunkingService) ChunkPlainText(ctx context.Context, content string, parentDocID string, options ChunkOptions) ([]Chunk, error) {
	// Split by paragraphs (double newline)
	paragraphs := s.splitParagraphs(content)

	position := 0
	chunks := s.chunkParagraphs(paragraphs, position, parentDocID, options, map[string]interface{}{
		"type": "plain_text",
	})

	// Apply overlap between chunks
	overlapSize := int(float64(options.ChunkSize) * options.OverlapRatio)
	if options.OverlapRatio > 0 && len(chunks) > 1 {
		chunks = s.applyOverlap(chunks, overlapSize)
	}

	return chunks, nil
}

// ChunkStructured chunks structured data (JSON, XML)
func (s *SmartChunkingService) ChunkStructured(ctx context.Context, content string, format string, parentDocID string, options ChunkOptions) ([]Chunk, error) {
	// For structured data, split by top-level elements/objects
	// This is a simplified implementation - production might use proper parsers

	var chunks []Chunk
	position := 0

	// Simple line-based splitting for now
	lines := strings.Split(content, "\n")
	currentChunk := []string{}
	currentTokens := 0

	for _, line := range lines {
		lineTokens := s.estimateTokens(line)

		if currentTokens+lineTokens > options.ChunkSize && len(currentChunk) > 0 {
			// Create chunk from accumulated lines
			chunkContent := strings.Join(currentChunk, "\n")
			chunk := s.createChunk(chunkContent, position, parentDocID, map[string]interface{}{
				"type":   "structured",
				"format": format,
			})
			chunks = append(chunks, chunk)
			position++

			// Start new chunk with overlap
			overlapLines := int(float64(len(currentChunk)) * options.OverlapRatio)
			if overlapLines > 0 {
				currentChunk = currentChunk[len(currentChunk)-overlapLines:]
				currentTokens = s.estimateTokens(strings.Join(currentChunk, "\n"))
			} else {
				currentChunk = []string{}
				currentTokens = 0
			}
		}

		currentChunk = append(currentChunk, line)
		currentTokens += lineTokens
	}

	// Add remaining content as final chunk
	if len(currentChunk) > 0 {
		chunkContent := strings.Join(currentChunk, "\n")
		chunk := s.createChunk(chunkContent, position, parentDocID, map[string]interface{}{
			"type":   "structured",
			"format": format,
		})
		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

// === Helper Functions ===

// MarkdownSection represents a markdown section
type MarkdownSection struct {
	Heading string
	Level   int
	Content string
}

// splitMarkdownSections splits markdown by headers
func (s *SmartChunkingService) splitMarkdownSections(content string) []MarkdownSection {
	var sections []MarkdownSection
	lines := strings.Split(content, "\n")

	var currentSection *MarkdownSection
	headerRegex := regexp.MustCompile(`^(#{1,6})\s+(.+)$`)

	for _, line := range lines {
		matches := headerRegex.FindStringSubmatch(line)
		if matches != nil {
			// Save previous section
			if currentSection != nil {
				currentSection.Content = strings.TrimSpace(currentSection.Content)
				sections = append(sections, *currentSection)
			}

			// Start new section
			level := len(matches[1])
			heading := matches[2]
			currentSection = &MarkdownSection{
				Heading: heading,
				Level:   level,
				Content: line + "\n",
			}
		} else if currentSection != nil {
			currentSection.Content += line + "\n"
		} else {
			// Content before first header
			if len(sections) == 0 {
				currentSection = &MarkdownSection{
					Heading: "",
					Level:   0,
					Content: line + "\n",
				}
			}
		}
	}

	// Add final section
	if currentSection != nil {
		currentSection.Content = strings.TrimSpace(currentSection.Content)
		sections = append(sections, *currentSection)
	}

	return sections
}

// CodeBlock represents a code block
type CodeBlock struct {
	Type    string // function, class, method, etc.
	Name    string
	Content string
}

// splitCodeBlocks splits code into logical blocks
func (s *SmartChunkingService) splitCodeBlocks(content string, language string) []CodeBlock {
	// Simplified implementation - production would use proper parsers
	var blocks []CodeBlock

	switch strings.ToLower(language) {
	case "go", "golang":
		blocks = s.splitGoCode(content)
	case "python", "py":
		blocks = s.splitPythonCode(content)
	case "javascript", "js", "typescript", "ts":
		blocks = s.splitJSCode(content)
	default:
		// Fallback: split by blank lines
		blocks = s.splitByBlankLines(content)
	}

	return blocks
}

// splitGoCode splits Go code by functions and types
func (s *SmartChunkingService) splitGoCode(content string) []CodeBlock {
	var blocks []CodeBlock
	lines := strings.Split(content, "\n")

	funcRegex := regexp.MustCompile(`^func\s+(\w+)`)
	typeRegex := regexp.MustCompile(`^type\s+(\w+)`)

	var currentBlock *CodeBlock
	var currentLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check for function declaration
		if matches := funcRegex.FindStringSubmatch(trimmed); matches != nil {
			if currentBlock != nil {
				currentBlock.Content = strings.Join(currentLines, "\n")
				blocks = append(blocks, *currentBlock)
			}
			currentBlock = &CodeBlock{
				Type: "function",
				Name: matches[1],
			}
			currentLines = []string{line}
		} else if matches := typeRegex.FindStringSubmatch(trimmed); matches != nil {
			if currentBlock != nil {
				currentBlock.Content = strings.Join(currentLines, "\n")
				blocks = append(blocks, *currentBlock)
			}
			currentBlock = &CodeBlock{
				Type: "type",
				Name: matches[1],
			}
			currentLines = []string{line}
		} else if currentBlock != nil {
			currentLines = append(currentLines, line)
		} else {
			// Code before first function/type
			if len(blocks) == 0 {
				currentBlock = &CodeBlock{
					Type: "header",
					Name: "package",
				}
				currentLines = []string{line}
			}
		}
	}

	// Add final block
	if currentBlock != nil {
		currentBlock.Content = strings.Join(currentLines, "\n")
		blocks = append(blocks, *currentBlock)
	}

	return blocks
}

// splitPythonCode splits Python code by functions and classes
func (s *SmartChunkingService) splitPythonCode(content string) []CodeBlock {
	var blocks []CodeBlock
	lines := strings.Split(content, "\n")

	funcRegex := regexp.MustCompile(`^def\s+(\w+)`)
	classRegex := regexp.MustCompile(`^class\s+(\w+)`)

	var currentBlock *CodeBlock
	var currentLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if matches := funcRegex.FindStringSubmatch(trimmed); matches != nil {
			if currentBlock != nil {
				currentBlock.Content = strings.Join(currentLines, "\n")
				blocks = append(blocks, *currentBlock)
			}
			currentBlock = &CodeBlock{
				Type: "function",
				Name: matches[1],
			}
			currentLines = []string{line}
		} else if matches := classRegex.FindStringSubmatch(trimmed); matches != nil {
			if currentBlock != nil {
				currentBlock.Content = strings.Join(currentLines, "\n")
				blocks = append(blocks, *currentBlock)
			}
			currentBlock = &CodeBlock{
				Type: "class",
				Name: matches[1],
			}
			currentLines = []string{line}
		} else if currentBlock != nil {
			currentLines = append(currentLines, line)
		} else {
			if len(blocks) == 0 {
				currentBlock = &CodeBlock{
					Type: "header",
					Name: "imports",
				}
				currentLines = []string{line}
			}
		}
	}

	if currentBlock != nil {
		currentBlock.Content = strings.Join(currentLines, "\n")
		blocks = append(blocks, *currentBlock)
	}

	return blocks
}

// splitJSCode splits JavaScript/TypeScript code
func (s *SmartChunkingService) splitJSCode(content string) []CodeBlock {
	var blocks []CodeBlock
	lines := strings.Split(content, "\n")

	funcRegex := regexp.MustCompile(`^(function|const|let|var)\s+(\w+)\s*[=\(]`)
	classRegex := regexp.MustCompile(`^class\s+(\w+)`)

	var currentBlock *CodeBlock
	var currentLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if matches := classRegex.FindStringSubmatch(trimmed); matches != nil {
			if currentBlock != nil {
				currentBlock.Content = strings.Join(currentLines, "\n")
				blocks = append(blocks, *currentBlock)
			}
			currentBlock = &CodeBlock{
				Type: "class",
				Name: matches[1],
			}
			currentLines = []string{line}
		} else if matches := funcRegex.FindStringSubmatch(trimmed); matches != nil {
			if currentBlock != nil {
				currentBlock.Content = strings.Join(currentLines, "\n")
				blocks = append(blocks, *currentBlock)
			}
			currentBlock = &CodeBlock{
				Type: "function",
				Name: matches[2],
			}
			currentLines = []string{line}
		} else if currentBlock != nil {
			currentLines = append(currentLines, line)
		} else {
			if len(blocks) == 0 {
				currentBlock = &CodeBlock{
					Type: "header",
					Name: "imports",
				}
				currentLines = []string{line}
			}
		}
	}

	if currentBlock != nil {
		currentBlock.Content = strings.Join(currentLines, "\n")
		blocks = append(blocks, *currentBlock)
	}

	return blocks
}

// splitByBlankLines splits content by blank lines
func (s *SmartChunkingService) splitByBlankLines(content string) []CodeBlock {
	var blocks []CodeBlock
	lines := strings.Split(content, "\n")

	var currentLines []string
	blockIdx := 0

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			if len(currentLines) > 0 {
				blocks = append(blocks, CodeBlock{
					Type:    "block",
					Name:    fmt.Sprintf("block_%d", blockIdx),
					Content: strings.Join(currentLines, "\n"),
				})
				blockIdx++
				currentLines = []string{}
			}
		} else {
			currentLines = append(currentLines, line)
		}
	}

	if len(currentLines) > 0 {
		blocks = append(blocks, CodeBlock{
			Type:    "block",
			Name:    fmt.Sprintf("block_%d", blockIdx),
			Content: strings.Join(currentLines, "\n"),
		})
	}

	return blocks
}

// splitParagraphs splits text into paragraphs
func (s *SmartChunkingService) splitParagraphs(content string) []string {
	// Split by double newline or more
	paragraphRegex := regexp.MustCompile(`\n\s*\n`)
	paragraphs := paragraphRegex.Split(content, -1)

	var result []string
	for _, p := range paragraphs {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

// chunkParagraphs combines paragraphs into chunks
func (s *SmartChunkingService) chunkParagraphs(paragraphs []string, startPosition int, parentDocID string, options ChunkOptions, baseMetadata map[string]interface{}) []Chunk {
	var chunks []Chunk
	position := startPosition

	var currentChunk []string
	currentTokens := 0

	for _, para := range paragraphs {
		paraTokens := s.estimateTokens(para)

		// If single paragraph exceeds chunk size, split by sentences
		if paraTokens > options.ChunkSize {
			// Flush current chunk
			if len(currentChunk) > 0 {
				chunkContent := strings.Join(currentChunk, "\n\n")
				metadata := s.copyMetadata(baseMetadata)
				metadata["paragraph_count"] = len(currentChunk)
				chunk := s.createChunk(chunkContent, position, parentDocID, metadata)
				chunks = append(chunks, chunk)
				position++
				currentChunk = []string{}
				currentTokens = 0
			}

			// Split large paragraph by sentences
			sentences := s.splitSentences(para)
			sentenceChunks := s.chunkSentences(sentences, position, parentDocID, options, baseMetadata)
			chunks = append(chunks, sentenceChunks...)
			position += len(sentenceChunks)
			continue
		}

		// Would adding this paragraph exceed chunk size?
		if currentTokens+paraTokens > options.ChunkSize && len(currentChunk) > 0 {
			// Create chunk from accumulated paragraphs
			chunkContent := strings.Join(currentChunk, "\n\n")
			metadata := s.copyMetadata(baseMetadata)
			metadata["paragraph_count"] = len(currentChunk)
			chunk := s.createChunk(chunkContent, position, parentDocID, metadata)
			chunks = append(chunks, chunk)
			position++

			// Start new chunk
			currentChunk = []string{}
			currentTokens = 0
		}

		currentChunk = append(currentChunk, para)
		currentTokens += paraTokens
	}

	// Add remaining content as final chunk
	if len(currentChunk) > 0 {
		chunkContent := strings.Join(currentChunk, "\n\n")
		metadata := s.copyMetadata(baseMetadata)
		metadata["paragraph_count"] = len(currentChunk)
		chunk := s.createChunk(chunkContent, position, parentDocID, metadata)
		chunks = append(chunks, chunk)
	}

	return chunks
}

// chunkSentences combines sentences into chunks
func (s *SmartChunkingService) chunkSentences(sentences []string, startPosition int, parentDocID string, options ChunkOptions, baseMetadata map[string]interface{}) []Chunk {
	var chunks []Chunk
	position := startPosition

	var currentChunk []string
	currentTokens := 0

	for _, sentence := range sentences {
		sentenceTokens := s.estimateTokens(sentence)

		if currentTokens+sentenceTokens > options.ChunkSize && len(currentChunk) > 0 {
			chunkContent := strings.Join(currentChunk, " ")
			metadata := s.copyMetadata(baseMetadata)
			metadata["sentence_count"] = len(currentChunk)
			chunk := s.createChunk(chunkContent, position, parentDocID, metadata)
			chunks = append(chunks, chunk)
			position++

			currentChunk = []string{}
			currentTokens = 0
		}

		currentChunk = append(currentChunk, sentence)
		currentTokens += sentenceTokens
	}

	if len(currentChunk) > 0 {
		chunkContent := strings.Join(currentChunk, " ")
		metadata := s.copyMetadata(baseMetadata)
		metadata["sentence_count"] = len(currentChunk)
		chunk := s.createChunk(chunkContent, position, parentDocID, metadata)
		chunks = append(chunks, chunk)
	}

	return chunks
}

// chunkCodeLines splits code into chunks by lines
func (s *SmartChunkingService) chunkCodeLines(content string, startPosition int, parentDocID string, options ChunkOptions, baseMetadata map[string]interface{}) []Chunk {
	var chunks []Chunk
	position := startPosition

	lines := strings.Split(content, "\n")
	var currentChunk []string
	currentTokens := 0

	for _, line := range lines {
		lineTokens := s.estimateTokens(line)

		if currentTokens+lineTokens > options.ChunkSize && len(currentChunk) > 0 {
			chunkContent := strings.Join(currentChunk, "\n")
			metadata := s.copyMetadata(baseMetadata)
			metadata["line_count"] = len(currentChunk)
			chunk := s.createChunk(chunkContent, position, parentDocID, metadata)
			chunks = append(chunks, chunk)
			position++

			currentChunk = []string{}
			currentTokens = 0
		}

		currentChunk = append(currentChunk, line)
		currentTokens += lineTokens
	}

	if len(currentChunk) > 0 {
		chunkContent := strings.Join(currentChunk, "\n")
		metadata := s.copyMetadata(baseMetadata)
		metadata["line_count"] = len(currentChunk)
		chunk := s.createChunk(chunkContent, position, parentDocID, metadata)
		chunks = append(chunks, chunk)
	}

	return chunks
}

// splitSentences splits text into sentences
func (s *SmartChunkingService) splitSentences(text string) []string {
	// Simple sentence splitting - production might use NLP library
	sentenceRegex := regexp.MustCompile(`[.!?]+\s+`)
	sentences := sentenceRegex.Split(text, -1)

	var result []string
	for _, s := range sentences {
		trimmed := strings.TrimSpace(s)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

// applyOverlap applies overlap between consecutive chunks
func (s *SmartChunkingService) applyOverlap(chunks []Chunk, overlapTokens int) []Chunk {
	if len(chunks) <= 1 || overlapTokens == 0 {
		return chunks
	}

	for i := 1; i < len(chunks); i++ {
		prevChunk := chunks[i-1]
		currentChunk := chunks[i]

		// Get overlap from previous chunk
		overlapContent := s.extractOverlap(prevChunk.Content, overlapTokens)
		if overlapContent != "" {
			// Prepend overlap to current chunk
			currentChunk.Content = overlapContent + "\n...\n" + currentChunk.Content
			currentChunk.TokenCount = s.estimateTokens(currentChunk.Content)
			currentChunk.Metadata["has_overlap"] = true
			currentChunk.Metadata["overlap_tokens"] = s.estimateTokens(overlapContent)
			chunks[i] = currentChunk
		}
	}

	return chunks
}

// extractOverlap extracts the last N tokens from content
func (s *SmartChunkingService) extractOverlap(content string, targetTokens int) string {
	// Split into sentences and take last few
	sentences := s.splitSentences(content)
	if len(sentences) == 0 {
		return ""
	}

	var overlap []string
	tokens := 0

	// Work backwards from end
	for i := len(sentences) - 1; i >= 0; i-- {
		sentenceTokens := s.estimateTokens(sentences[i])
		if tokens+sentenceTokens > targetTokens && len(overlap) > 0 {
			break
		}
		overlap = append([]string{sentences[i]}, overlap...)
		tokens += sentenceTokens
	}

	return strings.Join(overlap, " ")
}

// createChunk creates a chunk with ID and metadata
func (s *SmartChunkingService) createChunk(content string, position int, parentDocID string, metadata map[string]interface{}) Chunk {
	content = strings.TrimSpace(content)
	tokenCount := s.estimateTokens(content)

	// Generate unique chunk ID
	chunkID := s.generateChunkID(parentDocID, position, content)

	// Add system metadata
	metadata["position"] = position
	metadata["token_count"] = tokenCount
	metadata["char_count"] = len(content)

	return Chunk{
		ID:          chunkID,
		Content:     content,
		TokenCount:  tokenCount,
		Position:    position,
		Metadata:    metadata,
		ParentDocID: parentDocID,
	}
}

// generateChunkID generates a unique ID for a chunk
func (s *SmartChunkingService) generateChunkID(parentDocID string, position int, content string) string {
	// Hash based on parent, position, and content
	hasher := sha256.New()
	hasher.Write([]byte(fmt.Sprintf("%s:%d:%s", parentDocID, position, content)))
	hash := hex.EncodeToString(hasher.Sum(nil))
	return hash[:16] // Use first 16 chars
}

// estimateTokens estimates token count for text
func (s *SmartChunkingService) estimateTokens(text string) int {
	// Simple estimation: count characters and divide by chars per token
	// More accurate would be to use a tokenizer like tiktoken
	return int(float64(len(text)) / s.charsPerToken)
}

// detectLanguage attempts to detect programming language
func (s *SmartChunkingService) detectLanguage(content string) string {
	// Simple heuristic-based detection
	content = strings.TrimSpace(content)

	// Go
	if strings.Contains(content, "package ") && strings.Contains(content, "func ") {
		return "go"
	}

	// Python
	if strings.Contains(content, "def ") && strings.Contains(content, ":") {
		return "python"
	}

	// JavaScript/TypeScript
	if strings.Contains(content, "function ") || strings.Contains(content, "const ") || strings.Contains(content, "=>") {
		return "javascript"
	}

	// Java
	if strings.Contains(content, "public class ") || strings.Contains(content, "private ") {
		return "java"
	}

	// C/C++
	if strings.Contains(content, "#include") || strings.Contains(content, "int main(") {
		return "c"
	}

	// Fallback
	return "unknown"
}

// copyMetadata creates a copy of metadata map
func (s *SmartChunkingService) copyMetadata(metadata map[string]interface{}) map[string]interface{} {
	copy := make(map[string]interface{})
	for k, v := range metadata {
		copy[k] = v
	}
	return copy
}

// GetChunkStatistics returns statistics about chunks
func (s *SmartChunkingService) GetChunkStatistics(chunks []Chunk) map[string]interface{} {
	if len(chunks) == 0 {
		return map[string]interface{}{
			"count":       0,
			"total_tokens": 0,
			"avg_tokens":  0,
			"min_tokens":  0,
			"max_tokens":  0,
		}
	}

	totalTokens := 0
	minTokens := chunks[0].TokenCount
	maxTokens := chunks[0].TokenCount

	for _, chunk := range chunks {
		totalTokens += chunk.TokenCount
		if chunk.TokenCount < minTokens {
			minTokens = chunk.TokenCount
		}
		if chunk.TokenCount > maxTokens {
			maxTokens = chunk.TokenCount
		}
	}

	return map[string]interface{}{
		"count":        len(chunks),
		"total_tokens": totalTokens,
		"avg_tokens":   totalTokens / len(chunks),
		"min_tokens":   minTokens,
		"max_tokens":   maxTokens,
	}
}

// OptimizeChunks post-processes chunks to improve quality
func (s *SmartChunkingService) OptimizeChunks(chunks []Chunk, options ChunkOptions) []Chunk {
	var optimized []Chunk

	for i, chunk := range chunks {
		// Skip very small chunks unless it's the last one
		if chunk.TokenCount < options.MinChunkSize && i < len(chunks)-1 {
			// Merge with next chunk
			if i+1 < len(chunks) {
				nextChunk := chunks[i+1]
				mergedContent := chunk.Content + "\n\n" + nextChunk.Content
				mergedTokens := s.estimateTokens(mergedContent)

				if mergedTokens <= options.ChunkSize {
					// Create merged chunk
					merged := s.createChunk(mergedContent, chunk.Position, chunk.ParentDocID, chunk.Metadata)
					optimized = append(optimized, merged)
					// Skip next chunk since we merged it
					i++
					continue
				}
			}
		}

		// Remove excessive whitespace
		chunk.Content = s.normalizeWhitespace(chunk.Content)
		chunk.TokenCount = s.estimateTokens(chunk.Content)

		optimized = append(optimized, chunk)
	}

	return optimized
}

// normalizeWhitespace removes excessive whitespace
func (s *SmartChunkingService) normalizeWhitespace(text string) string {
	// Replace multiple spaces with single space
	text = regexp.MustCompile(`[ \t]+`).ReplaceAllString(text, " ")

	// Replace more than 2 newlines with 2 newlines
	text = regexp.MustCompile(`\n{3,}`).ReplaceAllString(text, "\n\n")

	// Trim leading/trailing whitespace
	text = strings.TrimSpace(text)

	return text
}

// ValidateChunk validates a chunk meets quality criteria
func (s *SmartChunkingService) ValidateChunk(chunk Chunk, options ChunkOptions) (bool, string) {
	// Check minimum size
	if chunk.TokenCount < options.MinChunkSize {
		return false, fmt.Sprintf("chunk too small: %d tokens (min: %d)", chunk.TokenCount, options.MinChunkSize)
	}

	// Check maximum size
	if chunk.TokenCount > options.ChunkSize*2 {
		return false, fmt.Sprintf("chunk too large: %d tokens (max: %d)", chunk.TokenCount, options.ChunkSize*2)
	}

	// Check content is not empty
	if strings.TrimSpace(chunk.Content) == "" {
		return false, "chunk content is empty"
	}

	// Check content is not just whitespace or punctuation
	hasAlphanumeric := false
	for _, r := range chunk.Content {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			hasAlphanumeric = true
			break
		}
	}
	if !hasAlphanumeric {
		return false, "chunk contains no alphanumeric characters"
	}

	return true, ""
}
