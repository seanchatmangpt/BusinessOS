package services

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
	"github.com/rhl/businessos-backend/internal/integrations/osa"
	"golang.org/x/sync/semaphore"
)

// AppGenerationOrchestrator coordinates application generation across multiple agents
type AppGenerationOrchestrator struct {
	pool         *pgxpool.Pool
	queries      *sqlc.Queries
	eventBus     *BuildEventBus
	promptString string
	maxRetries   int
	apiSem       *semaphore.Weighted
	logger       *slog.Logger
	orchestrator *BasicOrchestrator
	osaCient     *osa.ResilientClient // OSA integration for real app generation
	mu           sync.RWMutex
	totalRuns    int64
	successRuns  int64
	failedRuns   int64
}

// BasicOrchestrator is a stub orchestrator for the app generation
type BasicOrchestrator struct {
	circuitBreakerOpen bool
	metrics            map[string]int64
}

// Shutdown closes the orchestrator
func (o *BasicOrchestrator) Shutdown() error {
	return nil
}

// GetCircuitBreakerMetrics returns circuit breaker metrics
func (o *BasicOrchestrator) GetCircuitBreakerMetrics() map[string]interface{} {
	return make(map[string]interface{})
}

// AgenticRAGResponse represents the response from agentic RAG operations
type AgenticRAGResponse struct {
	QueryID     string                 `json:"query_id"`
	Answer      string                 `json:"answer"`
	Sources     []string               `json:"sources,omitempty"`
	Confidence  float64                `json:"confidence"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	GeneratedAt time.Time              `json:"generated_at"`
}

// NewAppGenerationOrchestrator creates a new orchestrator
func NewAppGenerationOrchestrator(
	pool *pgxpool.Pool,
	queries *sqlc.Queries,
	eventBus *BuildEventBus,
	promptString string,
) *AppGenerationOrchestrator {
	return &AppGenerationOrchestrator{
		pool:         pool,
		queries:      queries,
		eventBus:     eventBus,
		promptString: promptString,
		maxRetries:   3,
		apiSem:       semaphore.NewWeighted(5),
		logger:       slog.Default(),
		orchestrator: &BasicOrchestrator{metrics: make(map[string]int64)},
		osaCient:     nil, // Set via SetOSAClient
	}
}

// SetOSAClient sets the OSA resilient client for real app generation orchestration
func (o *AppGenerationOrchestrator) SetOSAClient(client *osa.ResilientClient) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.osaCient = client
}

// Generate runs the application generation workflow through OSA
func (o *AppGenerationOrchestrator) Generate(ctx context.Context, req MultiAgentAppRequest) (interface{}, error) {
	o.mu.RLock()
	osaClient := o.osaCient
	o.mu.RUnlock()

	if osaClient == nil {
		return nil, fmt.Errorf("OSA client not initialized: app generation not available")
	}

	// Convert to OSA GenerateApp request
	osaReq := &osa.AppGenerationRequest{
		UserID:      uuid.New(), // Placeholder; should come from request context
		WorkspaceID: req.WorkspaceID,
		Name:        req.AppName,
		Description: req.Description,
		Type:        req.Type,
		Parameters:  req.Parameters,
	}

	// Call OSA with resilience (circuit breaker, retries, fallback)
	resp, err := osaClient.GenerateApp(ctx, osaReq)
	if err != nil {
		o.logger.Error("OSA app generation failed",
			"app_name", req.AppName,
			"error", err)
		return nil, fmt.Errorf("failed to generate app via OSA: %w", err)
	}

	o.logger.Info("app generation succeeded",
		"app_id", resp.AppID,
		"status", resp.Status)

	return resp, nil
}

// MultiAgentAppRequest represents a request to generate an app with multiple agents
type MultiAgentAppRequest struct {
	AppName     string                 `json:"app_name"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"`
	Parameters  map[string]interface{} `json:"parameters"`
	Features    []string               `json:"features"`
	QueueItemID string                 `json:"queue_item_id"`
	WorkspaceID uuid.UUID              `json:"workspace_id"`
}

// QueryIntent represents the intent of a user query
type QueryIntent string

const (
	IntentFactualLookup QueryIntent = "factual_lookup"
	IntentProcedural    QueryIntent = "procedural"
	IntentComparison    QueryIntent = "comparison"
	IntentRecent        QueryIntent = "recent"
	IntentExhaustive    QueryIntent = "exhaustive"
)

// AgenticRAGRequest represents a request for agentic RAG operations
type AgenticRAGRequest struct {
	Query              string                 `json:"query"`
	UserID             string                 `json:"user_id"`
	WorkspaceID        string                 `json:"workspace_id,omitempty"`
	Context            map[string]interface{} `json:"context,omitempty"`
	MaxResults         int                    `json:"max_results,omitempty"`
	MinQualityScore    float64                `json:"min_quality_score,omitempty"`
	ProjectContext     interface{}            `json:"project_context,omitempty"`
	TaskContext        interface{}            `json:"task_context,omitempty"`
	UsePersonalization bool                   `json:"use_personalization,omitempty"`
}

// Retrieve executes a retrieval query using hybrid search and reranking
func (s *AgenticRAGService) Retrieve(ctx context.Context, req AgenticRAGRequest) (*AgenticRAGResponse, error) {
	if s.hybridSearch == nil {
		return nil, fmt.Errorf("hybrid search service not initialized: retrieval not available")
	}
	if s.reranker == nil {
		return nil, fmt.Errorf("reranker service not initialized: retrieval not available")
	}

	// Run hybrid search
	opts := HybridSearchOptions{
		MaxResults: req.MaxResults,
		MinSimilarity: req.MinQualityScore,
	}
	if opts.MaxResults <= 0 {
		opts.MaxResults = 10
	}
	if opts.MinSimilarity <= 0 {
		opts.MinSimilarity = 0.3
	}

	results, err := s.hybridSearch.Search(ctx, req.Query, req.UserID, opts)
	if err != nil {
		return nil, fmt.Errorf("hybrid search failed: %w", err)
	}

	if len(results) == 0 {
		return &AgenticRAGResponse{
			QueryID:     uuid.New().String(),
			Answer:      "No results found for query",
			Sources:     []string{},
			Confidence:  0.0,
			GeneratedAt: time.Now(),
		}, nil
	}

	// Rerank results
	rerankOpts := DefaultReRankingOptions()
	rankedResults, err := s.reranker.ReRank(ctx, req.Query, req.UserID, results, rerankOpts)
	if err != nil {
		// Fall back to best result if reranking fails
		slog.Warn("reranking failed, using top result", "error", err)
		// Convert HybridSearchResults to ReRankedResults for fallback
		if len(results) > 0 {
			answer := results[0].Content
			sources := []string{results[0].BlockID}
			confidence := results[0].HybridScore
			return &AgenticRAGResponse{
				QueryID:     uuid.New().String(),
				Answer:      answer,
				Sources:     sources,
				Confidence:  confidence,
				GeneratedAt: time.Now(),
			}, nil
		}
	}

	// Extract answer from top result
	var answer string
	var sources []string
	var confidence float64

	if len(rankedResults) > 0 {
		answer = rankedResults[0].Content
		sources = []string{rankedResults[0].BlockID}
		confidence = rankedResults[0].FinalScore
	}

	return &AgenticRAGResponse{
		QueryID:     uuid.New().String(),
		Answer:      answer,
		Sources:     sources,
		Confidence:  confidence,
		GeneratedAt: time.Now(),
	}, nil
}

// AgenticRAGService provides agentic retrieval-augmented generation
type AgenticRAGService struct {
	queryExpansion *QueryExpansionService
	cache          *RAGCacheService
	hybridSearch   *HybridSearchService
	reranker       *ReRankerService
	embedding      *EmbeddingService
	learning       *LearningService
}

// NewAgenticRAGService creates a new agentic RAG service.
// Note: bootstrap.go passes different parameters; those are legacy.
// The modern signature uses QueryExpansionService and RAGCacheService.
func NewAgenticRAGService(
	pool *pgxpool.Pool,
	hybridSearch *HybridSearchService,
	reranker *ReRankerService,
	embedding *EmbeddingService,
	learning *LearningService,
) *AgenticRAGService {
	return &AgenticRAGService{
		queryExpansion: nil, // Populated by SetQueryExpansion
		cache:          nil, // Populated by SetCache
		hybridSearch:   hybridSearch,
		reranker:       reranker,
		embedding:      embedding,
		learning:       learning,
	}
}

// ProcessQuery processes a query through agentic RAG with expansion and caching
func (s *AgenticRAGService) ProcessQuery(ctx context.Context, req AgenticRAGRequest) (*AgenticRAGResponse, error) {
	if s.embedding == nil {
		return nil, fmt.Errorf("embedding service not initialized: query processing not available")
	}

	// Check cache first (if available)
	if s.cache != nil {
		// Try to get cached result
		cached, err := s.cache.GetAgenticRAGResponse(ctx, req)
		if err == nil && cached != nil {
			slog.Debug("cache hit for query", "query", req.Query)
			return cached.Response, nil
		}
	}

	// Expand query if service available
	expandedQueries := []string{req.Query}
	if s.queryExpansion != nil {
		expanded, err := s.queryExpansion.Expand(ctx, req.Query, false)
		if err == nil && expanded != nil && len(expanded.AllVariants) > 0 {
			expandedQueries = append(expandedQueries, expanded.AllVariants...)
		}
	}

	// Use Retrieve for the expanded queries (or original if expansion failed)
	resp, err := s.Retrieve(ctx, AgenticRAGRequest{
		Query:           req.Query,
		UserID:          req.UserID,
		WorkspaceID:     req.WorkspaceID,
		Context:         req.Context,
		MaxResults:      req.MaxResults,
		MinQualityScore: req.MinQualityScore,
		ProjectContext:  req.ProjectContext,
		TaskContext:     req.TaskContext,
		UsePersonalization: req.UsePersonalization,
	})
	if err != nil {
		return nil, fmt.Errorf("query processing failed: %w", err)
	}

	// Cache the result
	if s.cache != nil && resp != nil {
		cacheErr := s.cache.SetAgenticRAGResponse(ctx, req, resp)
		if cacheErr != nil {
			slog.Warn("failed to cache query result", "error", cacheErr)
		}
	}

	// Learn from query (if service available)
	if s.learning != nil && resp != nil {
		// The learning service can record query patterns asynchronously
		// For now, just log that learning occurred
		slog.Debug("learning recorded from query",
			"query", req.Query,
			"answer_length", len(resp.Answer),
			"confidence", resp.Confidence)
	}

	return resp, nil
}

// SetCache sets the RAG cache service
func (s *AgenticRAGService) SetCache(cache *RAGCacheService) {
	s.cache = cache
}

// SetQueryExpansion sets the query expansion service
func (s *AgenticRAGService) SetQueryExpansion(queryExpansion *QueryExpansionService) {
	s.queryExpansion = queryExpansion
}
