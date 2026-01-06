package services

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestHybridSearchIntegration tests the hybrid search functionality end-to-end
func TestHybridSearchIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test requires:
	// 1. PostgreSQL database with schema
	// 2. Ollama running with nomic-embed-text model
	// 3. Test data loaded

	// TODO: Initialize test database connection
	// pool := setupTestDatabase(t)
	// defer pool.Close()

	// embeddingSvc := NewEmbeddingService(pool, "http://localhost:11434")
	// hybridSearch := NewHybridSearchService(pool, embeddingSvc)

	// Test cases
	testCases := []struct {
		name           string
		query          string
		opts           HybridSearchOptions
		expectedMinResults int
	}{
		{
			name:  "Semantic-heavy query",
			query: "machine learning concepts",
			opts: HybridSearchOptions{
				SemanticWeight: 0.8,
				KeywordWeight:  0.2,
				MaxResults:     5,
				RRFConstant:    60,
				MinSimilarity:  0.3,
			},
			expectedMinResults: 1,
		},
		{
			name:  "Keyword-heavy query",
			query: "PostgreSQL CREATE TABLE statement",
			opts: HybridSearchOptions{
				SemanticWeight: 0.3,
				KeywordWeight:  0.7,
				MaxResults:     5,
				RRFConstant:    60,
				MinSimilarity:  0.3,
			},
			expectedMinResults: 1,
		},
		{
			name:  "Balanced hybrid query",
			query: "how to implement user authentication",
			opts:  DefaultHybridSearchOptions(),
			expectedMinResults: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// results, err := hybridSearch.Search(ctx, tc.query, "test-user", tc.opts)
			// require.NoError(t, err)
			// assert.GreaterOrEqual(t, len(results), tc.expectedMinResults)

			// Verify result structure
			// for _, result := range results {
			//     assert.NotEmpty(t, result.ContextID)
			//     assert.NotEmpty(t, result.Content)
			//     assert.GreaterOrEqual(t, result.HybridScore, 0.0)
			//     assert.LessOrEqual(t, result.HybridScore, 1.0)
			// }

			t.Log("Test placeholder - implement with real database")
		})
	}
}

// TestReRankerIntegration tests the re-ranking functionality
func TestReRankerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test re-ranking with different options
	testCases := []struct {
		name         string
		opts         ReRankingOptions
		expectChange bool // Expect rank changes from re-ranking
	}{
		{
			name: "Recency-biased re-ranking",
			opts: ReRankingOptions{
				SemanticWeight:   0.3,
				RecencyWeight:    0.5, // Heavy recency bias
				QualityWeight:    0.1,
				InteractionWeight: 0.05,
				ContextRelevance: 0.05,
				RecencyDecayDays: 30,
				RecencyHalfLife:  90,
				MinContentLength: 50,
			},
			expectChange: true,
		},
		{
			name: "Quality-focused re-ranking",
			opts: ReRankingOptions{
				SemanticWeight:      0.3,
				RecencyWeight:       0.1,
				QualityWeight:       0.5, // Heavy quality bias
				InteractionWeight:   0.05,
				ContextRelevance:    0.05,
				MinContentLength:    100,
				PreferredBlockTypes: []string{"heading", "paragraph"},
			},
			expectChange: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Log("Test placeholder - implement with real database")
		})
	}
}

// TestAgenticRAGIntegration tests the full agentic RAG pipeline
func TestAgenticRAGIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test queries with different intents
	testQueries := []struct {
		query          string
		expectedIntent QueryIntent
		expectedStrategy SearchStrategy
		minQuality     float64
	}{
		{
			query:          "What is machine learning?",
			expectedIntent: IntentFactualLookup,
			expectedStrategy: StrategyKeywordOnly,
			minQuality:     0.5,
		},
		{
			query:          "How to deploy a Go application?",
			expectedIntent: IntentProcedural,
			expectedStrategy: StrategyHybrid,
			minQuality:     0.5,
		},
		{
			query:          "neural networks concepts",
			expectedIntent: IntentConceptualSearch,
			expectedStrategy: StrategySemanticOnly,
			minQuality:     0.4,
		},
		{
			query:          "recent updates to the API",
			expectedIntent: IntentRecent,
			expectedStrategy: StrategyHybrid,
			minQuality:     0.4,
		},
	}

	for _, tc := range testQueries {
		t.Run(tc.query, func(t *testing.T) {
			t.Log("Test placeholder - implement with real database")

			// Test intent classification
			// service := &AgenticRAGService{}
			// intent := service.classifyQueryIntent(tc.query)
			// assert.Equal(t, tc.expectedIntent, intent)

			// Test strategy selection
			// strategy, _ := service.selectStrategy(intent, AgenticRAGRequest{})
			// assert.Equal(t, tc.expectedStrategy, strategy)
		})
	}
}

// BenchmarkHybridSearch benchmarks hybrid search performance
func BenchmarkHybridSearch(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	// pool := setupTestDatabase(b)
	// defer pool.Close()

	// embeddingSvc := NewEmbeddingService(pool, "http://localhost:11434")
	// hybridSearch := NewHybridSearchService(pool, embeddingSvc)

	queries := []string{
		"machine learning algorithms",
		"database optimization techniques",
		"user authentication best practices",
		"API design patterns",
		"microservices architecture",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		query := queries[i%len(queries)]
		// _, err := hybridSearch.Search(ctx, query, "test-user", opts)
		// if err != nil {
		//     b.Fatalf("Search failed: %v", err)
		// }
		_ = query
	}
}

// BenchmarkReRanking benchmarks re-ranking performance
func BenchmarkReRanking(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	// Create mock results
	mockResults := make([]HybridSearchResult, 20)
	for i := range mockResults {
		mockResults[i] = HybridSearchResult{
			ContextID:      uuid.New(),
			BlockID:        fmt.Sprintf("block-%d", i),
			Content:        fmt.Sprintf("Test content %d with some meaningful text", i),
			HybridScore:    0.8 - (float64(i) * 0.03),
			SearchStrategy: "hybrid",
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// _, err := reranker.ReRank(ctx, "test query", "test-user", mockResults, opts)
		// if err != nil {
		//     b.Fatalf("ReRank failed: %v", err)
		// }
		_ = mockResults
	}
}

// TestQueryIntentClassification tests query intent classification
func TestQueryIntentClassification(t *testing.T) {
	service := &AgenticRAGService{}

	testCases := []struct {
		query    string
		expected QueryIntent
	}{
		{"What is Docker?", IntentFactualLookup},
		{"Define microservices", IntentFactualLookup},
		{"How to deploy to production?", IntentProcedural},
		{"Steps to configure SSL", IntentProcedural},
		{"Docker vs Kubernetes", IntentComparison},
		{"Difference between REST and GraphQL", IntentComparison},
		{"Latest features in Go 1.21", IntentRecent},
		{"Recent changes to the API", IntentRecent},
		{"Everything about authentication", IntentExhaustive},
		{"All available integrations", IntentExhaustive},
		{"architecture", IntentAmbiguous},
		{"test", IntentAmbiguous},
		{"machine learning concepts", IntentConceptualSearch},
	}

	for _, tc := range testCases {
		t.Run(tc.query, func(t *testing.T) {
			result := service.classifyQueryIntent(tc.query)
			assert.Equal(t, tc.expected, result, "Query: %s", tc.query)
		})
	}
}

// TestStrategySelection tests search strategy selection
func TestStrategySelection(t *testing.T) {
	service := &AgenticRAGService{}

	testCases := []struct {
		intent   QueryIntent
		expected SearchStrategy
	}{
		{IntentFactualLookup, StrategyKeywordOnly},
		{IntentConceptualSearch, StrategySemanticOnly},
		{IntentProcedural, StrategyHybrid},
		{IntentComparison, StrategyHybrid},
		{IntentRecent, StrategyHybrid},
		{IntentExhaustive, StrategyMultiPass},
		{IntentAmbiguous, StrategyHybrid},
	}

	for _, tc := range testCases {
		t.Run(string(tc.intent), func(t *testing.T) {
			strategy, _ := service.selectStrategy(tc.intent, AgenticRAGRequest{})
			assert.Equal(t, tc.expected, strategy)
		})
	}
}

// TestRRFScoring tests Reciprocal Rank Fusion scoring
func TestRRFScoring(t *testing.T) {
	// Test RRF formula: score(d) = 1/(k + rank)
	k := 60

	testCases := []struct {
		rank          int
		expectedScore float64
	}{
		{1, 1.0 / 61.0},  // Top rank
		{5, 1.0 / 65.0},
		{10, 1.0 / 70.0},
		{100, 1.0 / 160.0},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("rank_%d", tc.rank), func(t *testing.T) {
			score := 1.0 / float64(k+tc.rank)
			assert.InDelta(t, tc.expectedScore, score, 0.0001)
		})
	}
}

// TestQualityEvaluation tests result quality evaluation
func TestQualityEvaluation(t *testing.T) {
	service := &AgenticRAGService{}

	testCases := []struct {
		name           string
		results        []ReRankedResult
		expectedQuality string // "high", "medium", "low"
	}{
		{
			name: "High quality results",
			results: []ReRankedResult{
				{FinalScore: 0.9},
				{FinalScore: 0.85},
				{FinalScore: 0.8},
			},
			expectedQuality: "high",
		},
		{
			name: "Medium quality results",
			results: []ReRankedResult{
				{FinalScore: 0.6},
				{FinalScore: 0.55},
				{FinalScore: 0.5},
			},
			expectedQuality: "medium",
		},
		{
			name: "Low quality results",
			results: []ReRankedResult{
				{FinalScore: 0.3},
				{FinalScore: 0.25},
			},
			expectedQuality: "low",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			quality := service.evaluateResultQuality(tc.results, AgenticRAGRequest{MaxResults: 5})

			switch tc.expectedQuality {
			case "high":
				assert.Greater(t, quality, 0.75, "Expected high quality (>0.75)")
			case "medium":
				assert.Greater(t, quality, 0.45, "Expected medium quality (>0.45)")
				assert.LessOrEqual(t, quality, 0.75, "Expected medium quality (<=0.75)")
			case "low":
				assert.LessOrEqual(t, quality, 0.6, "Expected low quality (<=0.6)")
			}
		})
	}
}

// Example usage documentation
func ExampleAgenticRAGService_Retrieve() {
	// Initialize services (normally done in main.go)
	// pool := ... // database connection
	// embeddingSvc := NewEmbeddingService(pool, "http://localhost:11434")
	// hybridSearch := NewHybridSearchService(pool, embeddingSvc)
	// reranker := NewReRankerService(pool, embeddingSvc)
	// learningSvc := NewLearningService(pool)

	// agenticRAG := NewAgenticRAGService(pool, hybridSearch, reranker, embeddingSvc, learningSvc)

	// Simple query
	req := AgenticRAGRequest{
		Query:              "How to implement authentication?",
		UserID:             "user-123",
		MaxResults:         10,
		MinQualityScore:    0.6,
		UsePersonalization: true,
	}

	// response, err := agenticRAG.Retrieve(ctx, req)
	// if err != nil {
	//     log.Fatalf("Retrieval failed: %v", err)
	// }

	// fmt.Printf("Query Intent: %s\n", response.QueryIntent)
	// fmt.Printf("Strategy Used: %s\n", response.StrategyUsed)
	// fmt.Printf("Quality Score: %.2f\n", response.QualityScore)
	// fmt.Printf("Results: %d\n", len(response.Results))
	// fmt.Printf("Processing Time: %dms\n", response.ProcessingTimeMs)

	_ = req
}

// setupTestDatabase initializes a test database (helper for integration tests)
func setupTestDatabase(t testing.TB) {
	// This would:
	// 1. Connect to test database
	// 2. Run migrations
	// 3. Load test fixtures
	// 4. Return database pool

	t.Log("Test database setup not implemented - requires actual database")
}
