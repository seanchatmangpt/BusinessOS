package research

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/services"
)

// ResearchExecutorService executes research queries in parallel
type ResearchExecutorService struct {
	pool             *pgxpool.Pool
	webSearchService *services.WebSearchService
	hybridSearch     *services.HybridSearchService
	memoryService    *services.MemoryHierarchyService
	embeddingService *services.EmbeddingService
}

// NewResearchExecutorService creates a new research executor
func NewResearchExecutorService(
	pool *pgxpool.Pool,
	webSearch *services.WebSearchService,
	hybridSearch *services.HybridSearchService,
	memoryService *services.MemoryHierarchyService,
	embeddingService *services.EmbeddingService,
) *ResearchExecutorService {
	return &ResearchExecutorService{
		pool:             pool,
		webSearchService: webSearch,
		hybridSearch:     hybridSearch,
		memoryService:    memoryService,
		embeddingService: embeddingService,
	}
}

// SourceType indicates where a source came from
type SourceType string

const (
	SourceTypeWeb    SourceType = "web"
	SourceTypeRAG    SourceType = "rag"
	SourceTypeMemory SourceType = "memory"
)

// ResearchSource represents a single source found during research
type ResearchSource struct {
	ID             uuid.UUID  `json:"id"`
	TaskID         uuid.UUID  `json:"task_id"`
	QueryID        *uuid.UUID `json:"query_id,omitempty"`
	SourceType     SourceType `json:"source_type"`
	URL            *string    `json:"url,omitempty"`
	Domain         *string    `json:"domain,omitempty"`
	Title          string     `json:"title"`
	Content        string     `json:"content"`
	Snippet        string     `json:"snippet"`
	RelevanceScore float64    `json:"relevance_score"`
	Author         *string    `json:"author,omitempty"`
	PublishedAt    *time.Time `json:"published_at,omitempty"`
	ContentHash    string     `json:"content_hash"`
}

// ExecutionResult contains results from executing a sub-question
type ExecutionResult struct {
	QueryID  uuid.UUID        `json:"query_id"`
	Question string           `json:"question"`
	Sources  []ResearchSource `json:"sources"`
	Duration time.Duration    `json:"duration"`
	Error    error            `json:"error,omitempty"`
}

// Execute runs all sub-questions in parallel and collects sources
func (e *ResearchExecutorService) Execute(
	ctx context.Context,
	taskID uuid.UUID,
	queries []SubQuestion,
	userID string,
	workspaceID uuid.UUID,
) ([]ExecutionResult, error) {
	// Channel to collect results
	resultsChan := make(chan ExecutionResult, len(queries))

	// WaitGroup for parallel execution
	var wg sync.WaitGroup

	// Execute each query in parallel
	for _, query := range queries {
		wg.Add(1)
		go func(q SubQuestion) {
			defer wg.Done()

			start := time.Now()
			sources, err := e.executeQuery(ctx, taskID, q, userID, workspaceID)

			resultsChan <- ExecutionResult{
				QueryID:  q.ID,
				Question: q.Question,
				Sources:  sources,
				Duration: time.Since(start),
				Error:    err,
			}
		}(query)
	}

	// Wait for all queries to complete
	wg.Wait()
	close(resultsChan)

	// Collect results
	results := make([]ExecutionResult, 0, len(queries))
	for result := range resultsChan {
		results = append(results, result)
	}

	return results, nil
}

// executeQuery executes a single research query
func (e *ResearchExecutorService) executeQuery(
	ctx context.Context,
	taskID uuid.UUID,
	query SubQuestion,
	userID string,
	workspaceID uuid.UUID,
) ([]ResearchSource, error) {
	// Timeout for single query
	queryCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var allSources []ResearchSource

	// Execute based on search type
	switch query.SearchType {
	case SearchTypeWeb:
		sources, err := e.searchWeb(queryCtx, taskID, query)
		if err != nil {
			return nil, fmt.Errorf("web search failed: %w", err)
		}
		allSources = append(allSources, sources...)

	case SearchTypeRAG:
		sources, err := e.searchRAG(queryCtx, taskID, query, userID, workspaceID)
		if err != nil {
			return nil, fmt.Errorf("RAG search failed: %w", err)
		}
		allSources = append(allSources, sources...)

	case SearchTypeMemory:
		sources, err := e.searchMemory(queryCtx, taskID, query, userID, workspaceID)
		if err != nil {
			return nil, fmt.Errorf("memory search failed: %w", err)
		}
		allSources = append(allSources, sources...)

	case SearchTypeHybrid:
		// Execute all three in parallel
		var wg sync.WaitGroup
		sourceChan := make(chan []ResearchSource, 3)

		wg.Add(3)
		go func() {
			defer wg.Done()
			if sources, err := e.searchWeb(queryCtx, taskID, query); err == nil {
				sourceChan <- sources
			}
		}()
		go func() {
			defer wg.Done()
			if sources, err := e.searchRAG(queryCtx, taskID, query, userID, workspaceID); err == nil {
				sourceChan <- sources
			}
		}()
		go func() {
			defer wg.Done()
			if sources, err := e.searchMemory(queryCtx, taskID, query, userID, workspaceID); err == nil {
				sourceChan <- sources
			}
		}()

		wg.Wait()
		close(sourceChan)

		for sources := range sourceChan {
			allSources = append(allSources, sources...)
		}
	}

	return allSources, nil
}

// searchWeb searches external web sources
func (e *ResearchExecutorService) searchWeb(
	ctx context.Context,
	taskID uuid.UUID,
	query SubQuestion,
) ([]ResearchSource, error) {
	if e.webSearchService == nil {
		return nil, fmt.Errorf("web search service not available")
	}

	// Execute web search
	results, err := e.webSearchService.Search(ctx, query.Question, 10)
	if err != nil {
		return nil, err
	}

	// Convert to ResearchSources
	sources := make([]ResearchSource, 0, len(results.Results))
	for i, result := range results.Results {
		domain := extractDomain(result.URL)
		contentHash := computeHash(result.Snippet)

		sources = append(sources, ResearchSource{
			ID:             uuid.New(),
			TaskID:         taskID,
			QueryID:        &query.ID,
			SourceType:     SourceTypeWeb,
			URL:            &result.URL,
			Domain:         &domain,
			Title:          result.Title,
			Content:        result.Snippet, // Full content not available from search API
			Snippet:        result.Snippet,
			RelevanceScore: float64(len(results.Results)-i) / float64(len(results.Results)), // Rank-based score
			ContentHash:    contentHash,
		})
	}

	return sources, nil
}

// searchRAG searches internal documents using RAG
func (e *ResearchExecutorService) searchRAG(
	ctx context.Context,
	taskID uuid.UUID,
	query SubQuestion,
	userID string,
	workspaceID uuid.UUID,
) ([]ResearchSource, error) {
	if e.hybridSearch == nil {
		return nil, fmt.Errorf("hybrid search service not available")
	}

	// Execute hybrid search (semantic + keyword)
	results, err := e.hybridSearch.Search(ctx, query.Question, userID, services.HybridSearchOptions{
		SemanticWeight: 0.7,
		KeywordWeight:  0.3,
		MaxResults:     10,
		RRFConstant:    60,
	})
	if err != nil {
		return nil, err
	}

	// Convert to ResearchSources
	sources := make([]ResearchSource, 0, len(results))
	for _, result := range results {
		contentHash := computeHash(result.Content)

		sources = append(sources, ResearchSource{
			ID:             uuid.New(),
			TaskID:         taskID,
			QueryID:        &query.ID,
			SourceType:     SourceTypeRAG,
			Title:          result.ContextName, // Use context name as title
			Content:        result.Content,
			Snippet:        truncate(result.Content, 300),
			RelevanceScore: result.HybridScore, // Use hybrid RRF score
			ContentHash:    contentHash,
		})
	}

	return sources, nil
}

// searchMemory searches workspace memory
func (e *ResearchExecutorService) searchMemory(
	ctx context.Context,
	taskID uuid.UUID,
	query SubQuestion,
	userID string,
	workspaceID uuid.UUID,
) ([]ResearchSource, error) {
	if e.memoryService == nil {
		return nil, fmt.Errorf("memory service not available")
	}

	// Get accessible memories (all types)
	memories, err := e.memoryService.GetAccessibleMemories(ctx, workspaceID, userID, nil, 10)
	if err != nil {
		return nil, err
	}

	// Filter and score memories by relevance to query
	// TODO: Implement semantic scoring
	sources := make([]ResearchSource, 0, len(memories))
	for _, memory := range memories {
		contentHash := computeHash(memory.Content)

		sources = append(sources, ResearchSource{
			ID:             uuid.New(),
			TaskID:         taskID,
			QueryID:        &query.ID,
			SourceType:     SourceTypeMemory,
			Title:          memory.Title,
			Content:        memory.Content,
			Snippet:        truncate(memory.Content, 300),
			RelevanceScore: 0.5, // TODO: Compute semantic similarity
			ContentHash:    contentHash,
		})
	}

	return sources, nil
}

// Helper functions

func extractDomain(urlStr string) string {
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}
	return parsed.Host
}

func computeHash(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
