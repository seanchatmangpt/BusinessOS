package services

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// Benchmark test data generators
func generateTestText(size int) string {
	words := []string{
		"artificial", "intelligence", "machine", "learning", "neural", "network",
		"data", "science", "algorithm", "optimization", "research", "development",
		"technology", "innovation", "application", "system", "process", "method",
		"analysis", "computation", "model", "training", "evaluation", "performance",
	}

	var builder strings.Builder
	rand.Seed(time.Now().UnixNano())

	for builder.Len() < size {
		word := words[rand.Intn(len(words))]
		builder.WriteString(word)
		builder.WriteString(" ")
	}

	return builder.String()[:size]
}

func generateTestBlocks(count int, contentSize int) []Block {
	blocks := make([]Block, count)
	for i := 0; i < count; i++ {
		blocks[i] = Block{
			ID:      fmt.Sprintf("block-%d", i),
			Type:    "paragraph",
			Content: generateTestText(contentSize),
		}
	}
	return blocks
}

func generateTestQuery() string {
	queries := []string{
		"machine learning algorithm optimization",
		"neural network training methods",
		"data science best practices",
		"artificial intelligence applications",
		"performance evaluation techniques",
	}
	return queries[rand.Intn(len(queries))]
}

// ============================================================================
// Embedding Generation Benchmarks
// ============================================================================

// BenchmarkTextEmbedding benchmarks text embedding generation
func BenchmarkTextEmbedding(b *testing.B) {
	ctx := context.Background()
	pool := setupTestDB(b)
	defer pool.Close()

	service := NewEmbeddingService(pool, "http://localhost:11434")

	testCases := []struct {
		name string
		size int
	}{
		{"Small-50chars", 50},
		{"Medium-500chars", 500},
		{"Large-2000chars", 2000},
		{"XLarge-5000chars", 5000},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			text := generateTestText(tc.size)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_, err := service.GenerateEmbedding(ctx, text)
				if err != nil {
					b.Logf("Warning: embedding failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkTextEmbeddingParallel benchmarks parallel embedding generation
func BenchmarkTextEmbeddingParallel(b *testing.B) {
	ctx := context.Background()
	pool := setupTestDB(b)
	defer pool.Close()

	service := NewEmbeddingService(pool, "http://localhost:11434")
	text := generateTestText(500)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := service.GenerateEmbedding(ctx, text)
			if err != nil {
				b.Logf("Warning: embedding failed: %v", err)
			}
		}
	})
}

// BenchmarkImageEmbedding benchmarks image embedding generation
func BenchmarkImageEmbedding(b *testing.B) {
	ctx := context.Background()
	pool := setupTestDB(b)
	defer pool.Close()

	config := ImageEmbeddingConfig{
		Provider:     "local",
		LocalBaseURL: "http://localhost:8000",
		Dimensions:   512,
	}
	service := NewImageEmbeddingService(pool, config)

	// Generate a dummy image (1x1 pixel PNG)
	imageData := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := service.GenerateEmbedding(ctx, imageData)
		if err != nil {
			b.Logf("Warning: image embedding failed: %v", err)
		}
	}
}

// ============================================================================
// Vector Search Benchmarks
// ============================================================================

// BenchmarkVectorSearch benchmarks vector similarity search
func BenchmarkVectorSearch(b *testing.B) {
	ctx := context.Background()
	pool := setupTestDB(b)
	defer pool.Close()

	service := NewEmbeddingService(pool, "http://localhost:11434")
	userID := "bench-user-" + uuid.New().String()

	// Setup: Index test documents
	contextID := uuid.New()
	setupTestContext(ctx, pool, contextID, userID)

	dataSizes := []struct {
		name       string
		blockCount int
	}{
		{"Small-10blocks", 10},
		{"Medium-100blocks", 100},
		{"Large-1000blocks", 1000},
	}

	for _, ds := range dataSizes {
		b.Run(ds.name, func(b *testing.B) {
			// Index blocks
			blocks := generateTestBlocks(ds.blockCount, 200)
			err := service.IndexDocument(ctx, contextID, blocks)
			if err != nil {
				b.Fatalf("Failed to index: %v", err)
			}

			query := generateTestQuery()

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_, err := service.SimilaritySearch(ctx, query, userID, 10)
				if err != nil {
					b.Fatalf("Search failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkVectorSearchLimits benchmarks different result limits
func BenchmarkVectorSearchLimits(b *testing.B) {
	ctx := context.Background()
	pool := setupTestDB(b)
	defer pool.Close()

	service := NewEmbeddingService(pool, "http://localhost:11434")
	userID := "bench-user-" + uuid.New().String()

	// Setup
	contextID := uuid.New()
	setupTestContext(ctx, pool, contextID, userID)
	blocks := generateTestBlocks(500, 200)
	service.IndexDocument(ctx, contextID, blocks)

	limits := []int{5, 10, 25, 50, 100}

	for _, limit := range limits {
		b.Run(fmt.Sprintf("Limit-%d", limit), func(b *testing.B) {
			query := generateTestQuery()

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_, err := service.SimilaritySearch(ctx, query, userID, limit)
				if err != nil {
					b.Fatalf("Search failed: %v", err)
				}
			}
		})
	}
}

// ============================================================================
// Hybrid Search Benchmarks
// ============================================================================

// BenchmarkHybridSearch benchmarks hybrid semantic + keyword search
func BenchmarkHybridSearch(b *testing.B) {
	ctx := context.Background()
	pool := setupTestDB(b)
	defer pool.Close()

	embeddingSvc := NewEmbeddingService(pool, "http://localhost:11434")
	hybridSvc := NewHybridSearchService(pool, embeddingSvc)
	userID := "bench-user-" + uuid.New().String()

	// Setup
	contextID := uuid.New()
	setupTestContext(ctx, pool, contextID, userID)
	blocks := generateTestBlocks(100, 200)
	embeddingSvc.IndexDocument(ctx, contextID, blocks)

	opts := DefaultHybridSearchOptions()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		query := generateTestQuery()
		_, err := hybridSvc.Search(ctx, query, userID, opts)
		if err != nil {
			b.Fatalf("Hybrid search failed: %v", err)
		}
	}
}

// BenchmarkHybridSearchWeights benchmarks different semantic/keyword weight combinations
func BenchmarkHybridSearchWeights(b *testing.B) {
	ctx := context.Background()
	pool := setupTestDB(b)
	defer pool.Close()

	embeddingSvc := NewEmbeddingService(pool, "http://localhost:11434")
	hybridSvc := NewHybridSearchService(pool, embeddingSvc)
	userID := "bench-user-" + uuid.New().String()

	// Setup
	contextID := uuid.New()
	setupTestContext(ctx, pool, contextID, userID)
	blocks := generateTestBlocks(100, 200)
	embeddingSvc.IndexDocument(ctx, contextID, blocks)

	weightCombos := []struct {
		name           string
		semanticWeight float64
		keywordWeight  float64
	}{
		{"SemanticOnly", 1.0, 0.0},
		{"KeywordOnly", 0.0, 1.0},
		{"Balanced", 0.5, 0.5},
		{"SemanticHeavy", 0.7, 0.3},
		{"KeywordHeavy", 0.3, 0.7},
	}

	for _, wc := range weightCombos {
		b.Run(wc.name, func(b *testing.B) {
			opts := HybridSearchOptions{
				SemanticWeight: wc.semanticWeight,
				KeywordWeight:  wc.keywordWeight,
				MaxResults:     10,
				RRFConstant:    60,
				MinSimilarity:  0.3,
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				query := generateTestQuery()
				_, err := hybridSvc.Search(ctx, query, userID, opts)
				if err != nil {
					b.Fatalf("Search failed: %v", err)
				}
			}
		})
	}
}

// ============================================================================
// Re-Ranking Benchmarks
// ============================================================================

// BenchmarkReRanking benchmarks result re-ranking
func BenchmarkReRanking(b *testing.B) {
	ctx := context.Background()
	pool := setupTestDB(b)
	defer pool.Close()

	embeddingSvc := NewEmbeddingService(pool, "http://localhost:11434")
	hybridSvc := NewHybridSearchService(pool, embeddingSvc)
	rerankerSvc := NewReRankerService(pool, embeddingSvc)
	userID := "bench-user-" + uuid.New().String()

	// Setup
	contextID := uuid.New()
	setupTestContext(ctx, pool, contextID, userID)
	blocks := generateTestBlocks(100, 200)
	embeddingSvc.IndexDocument(ctx, contextID, blocks)

	// Get initial results
	query := generateTestQuery()
	results, _ := hybridSvc.Search(ctx, query, userID, DefaultHybridSearchOptions())

	opts := DefaultReRankingOptions()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := rerankerSvc.ReRank(ctx, query, userID, results, opts)
		if err != nil {
			b.Fatalf("Re-ranking failed: %v", err)
		}
	}
}

// BenchmarkReRankingResultCounts benchmarks re-ranking with different result counts
func BenchmarkReRankingResultCounts(b *testing.B) {
	ctx := context.Background()
	pool := setupTestDB(b)
	defer pool.Close()

	embeddingSvc := NewEmbeddingService(pool, "http://localhost:11434")
	hybridSvc := NewHybridSearchService(pool, embeddingSvc)
	rerankerSvc := NewReRankerService(pool, embeddingSvc)
	userID := "bench-user-" + uuid.New().String()

	// Setup
	contextID := uuid.New()
	setupTestContext(ctx, pool, contextID, userID)
	blocks := generateTestBlocks(200, 200)
	embeddingSvc.IndexDocument(ctx, contextID, blocks)

	query := generateTestQuery()
	opts := DefaultReRankingOptions()

	resultCounts := []int{10, 25, 50, 100}

	for _, count := range resultCounts {
		b.Run(fmt.Sprintf("Results-%d", count), func(b *testing.B) {
			hybridOpts := DefaultHybridSearchOptions()
			hybridOpts.MaxResults = count
			results, _ := hybridSvc.Search(ctx, query, userID, hybridOpts)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_, err := rerankerSvc.ReRank(ctx, query, userID, results, opts)
				if err != nil {
					b.Fatalf("Re-ranking failed: %v", err)
				}
			}
		})
	}
}

// ============================================================================
// Chunking Benchmarks
// ============================================================================

// BenchmarkSmartChunking benchmarks document chunking strategies
func BenchmarkSmartChunking(b *testing.B) {
	processor := &DocumentProcessor{}

	documentSizes := []struct {
		name string
		size int
	}{
		{"Small-1KB", 1024},
		{"Medium-10KB", 10240},
		{"Large-100KB", 102400},
		{"XLarge-1MB", 1048576},
	}

	for _, ds := range documentSizes {
		b.Run(ds.name, func(b *testing.B) {
			text := generateTestText(ds.size)
			opts := DefaultChunkingOptions()

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_ = processor.chunkDocument(text, opts)
			}
		})
	}
}

// BenchmarkChunkingSizes benchmarks different chunk sizes
func BenchmarkChunkingSizes(b *testing.B) {
	processor := &DocumentProcessor{}
	text := generateTestText(50000) // 50KB document

	chunkSizes := []int{500, 1000, 1500, 2000, 3000}

	for _, size := range chunkSizes {
		b.Run(fmt.Sprintf("ChunkSize-%d", size), func(b *testing.B) {
			opts := ChunkingOptions{
				MaxChunkSize:    size,
				ChunkOverlap:    200,
				PreserveHeaders: true,
				SplitOnHeaders:  true,
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_ = processor.chunkDocument(text, opts)
			}
		})
	}
}

// BenchmarkChunkingStrategies benchmarks different chunking strategies
func BenchmarkChunkingStrategies(b *testing.B) {
	processor := &DocumentProcessor{}
	text := generateTestText(20000)

	strategies := []struct {
		name string
		opts ChunkingOptions
	}{
		{
			"FixedSizeNoOverlap",
			ChunkingOptions{MaxChunkSize: 1500, ChunkOverlap: 0, PreserveHeaders: false, SplitOnHeaders: false},
		},
		{
			"FixedSizeWithOverlap",
			ChunkingOptions{MaxChunkSize: 1500, ChunkOverlap: 200, PreserveHeaders: false, SplitOnHeaders: false},
		},
		{
			"HeaderBased",
			ChunkingOptions{MaxChunkSize: 1500, ChunkOverlap: 200, PreserveHeaders: true, SplitOnHeaders: true},
		},
	}

	for _, strategy := range strategies {
		b.Run(strategy.name, func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_ = processor.chunkDocument(text, strategy.opts)
			}
		})
	}
}

// ============================================================================
// Cache Benchmarks
// ============================================================================

// BenchmarkCacheHit benchmarks cache hit performance
func BenchmarkCacheHit(b *testing.B) {
	ctx := context.Background()
	redisClient := setupTestRedis(b)
	defer redisClient.Close()

	config := DefaultRAGCacheConfig()
	cache := NewRAGCacheService(redisClient, config)

	embeddingSvc := &EmbeddingService{cache: cache}
	text := generateTestText(500)

	// Warm up cache
	embedding := make([]float32, 768)
	for i := range embedding {
		embedding[i] = rand.Float32()
	}
	cache.SetEmbedding(ctx, text, embedding)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := cache.GetEmbedding(ctx, text)
		if err != nil {
			b.Fatalf("Cache get failed: %v", err)
		}
	}
}

// BenchmarkCacheMiss benchmarks cache miss performance
func BenchmarkCacheMiss(b *testing.B) {
	ctx := context.Background()
	redisClient := setupTestRedis(b)
	defer redisClient.Close()

	config := DefaultRAGCacheConfig()
	cache := NewRAGCacheService(redisClient, config)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		text := generateTestText(500) + fmt.Sprintf("-%d", i) // Unique text each time
		_, _ = cache.GetEmbedding(ctx, text)
	}
}

// BenchmarkCacheSet benchmarks cache write performance
func BenchmarkCacheSet(b *testing.B) {
	ctx := context.Background()
	redisClient := setupTestRedis(b)
	defer redisClient.Close()

	config := DefaultRAGCacheConfig()
	cache := NewRAGCacheService(redisClient, config)

	embedding := make([]float32, 768)
	for i := range embedding {
		embedding[i] = rand.Float32()
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		text := generateTestText(500) + fmt.Sprintf("-%d", i)
		err := cache.SetEmbedding(ctx, text, embedding)
		if err != nil {
			b.Fatalf("Cache set failed: %v", err)
		}
	}
}

// BenchmarkHybridSearchWithCache benchmarks hybrid search with caching
func BenchmarkHybridSearchWithCache(b *testing.B) {
	ctx := context.Background()
	pool := setupTestDB(b)
	defer pool.Close()

	redisClient := setupTestRedis(b)
	defer redisClient.Close()

	cache := NewRAGCacheService(redisClient, DefaultRAGCacheConfig())
	embeddingSvc := NewEmbeddingService(pool, "http://localhost:11434")
	embeddingSvc.SetCache(cache)

	hybridSvc := NewHybridSearchService(pool, embeddingSvc)
	userID := "bench-user-" + uuid.New().String()

	// Setup
	contextID := uuid.New()
	setupTestContext(ctx, pool, contextID, userID)
	blocks := generateTestBlocks(100, 200)
	embeddingSvc.IndexDocument(ctx, contextID, blocks)

	query := generateTestQuery()
	opts := DefaultHybridSearchOptions()

	// Warm up cache
	hybridSvc.Search(ctx, query, userID, opts)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := hybridSvc.Search(ctx, query, userID, opts)
		if err != nil {
			b.Fatalf("Search failed: %v", err)
		}
	}
}

// ============================================================================
// End-to-End Pipeline Benchmarks
// ============================================================================

// BenchmarkFullRAGPipeline benchmarks the complete RAG pipeline
func BenchmarkFullRAGPipeline(b *testing.B) {
	ctx := context.Background()
	pool := setupTestDB(b)
	defer pool.Close()

	redisClient := setupTestRedis(b)
	defer redisClient.Close()

	cache := NewRAGCacheService(redisClient, DefaultRAGCacheConfig())
	embeddingSvc := NewEmbeddingService(pool, "http://localhost:11434")
	embeddingSvc.SetCache(cache)

	hybridSvc := NewHybridSearchService(pool, embeddingSvc)
	rerankerSvc := NewReRankerService(pool, embeddingSvc)
	userID := "bench-user-" + uuid.New().String()

	// Setup
	contextID := uuid.New()
	setupTestContext(ctx, pool, contextID, userID)
	blocks := generateTestBlocks(100, 200)
	embeddingSvc.IndexDocument(ctx, contextID, blocks)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		query := generateTestQuery()

		// 1. Hybrid search
		results, err := hybridSvc.Search(ctx, query, userID, DefaultHybridSearchOptions())
		if err != nil {
			b.Fatalf("Search failed: %v", err)
		}

		// 2. Re-rank
		_, err = rerankerSvc.ReRank(ctx, query, userID, results, DefaultReRankingOptions())
		if err != nil {
			b.Fatalf("Re-ranking failed: %v", err)
		}
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

func setupTestDB(b *testing.B) *pgxpool.Pool {
	// In a real test, connect to test database
	// For benchmark, we'll skip if no connection available
	b.Skip("Skipping: requires database connection")
	return nil
}

func setupTestRedis(b *testing.B) *redis.Client {
	// In a real test, connect to Redis
	b.Skip("Skipping: requires Redis connection")
	return nil
}

func setupTestContext(ctx context.Context, pool *pgxpool.Pool, contextID uuid.UUID, userID string) {
	// Create test context in database
	query := `
		INSERT INTO contexts (id, user_id, name, type, is_archived)
		VALUES ($1, $2, 'Benchmark Context', 'document', false)
		ON CONFLICT (id) DO NOTHING
	`
	pool.Exec(ctx, query, contextID, userID)
}
