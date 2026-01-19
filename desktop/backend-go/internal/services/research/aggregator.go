package research

import (
	"context"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/services"
)

// ResearchAggregatorService combines and ranks sources from multiple queries
type ResearchAggregatorService struct {
	pool             *pgxpool.Pool
	embeddingService *services.EmbeddingService
}

// NewResearchAggregatorService creates a new research aggregator
func NewResearchAggregatorService(
	pool *pgxpool.Pool,
	embeddingService *services.EmbeddingService,
) *ResearchAggregatorService {
	return &ResearchAggregatorService{
		pool:             pool,
		embeddingService: embeddingService,
	}
}

// AggregationConfig controls aggregation behavior
type AggregationConfig struct {
	MaxSources             int     // Maximum sources in final set
	RRFConstant            int     // Reciprocal Rank Fusion constant (typically 60)
	DeduplicationThreshold float64 // Cosine similarity threshold (0.9 = 90% similar)
	MinRelevanceScore      float64 // Minimum score to include
	DiversityEnabled       bool    // Ensure source diversity
	MaxSourcesPerDomain    int     // Limit sources from single domain
}

// DefaultAggregationConfig returns sensible defaults
func DefaultAggregationConfig() AggregationConfig {
	return AggregationConfig{
		MaxSources:             15,
		RRFConstant:            60,
		DeduplicationThreshold: 0.90,
		MinRelevanceScore:      0.3,
		DiversityEnabled:       true,
		MaxSourcesPerDomain:    3,
	}
}

// RankedSource represents a source after aggregation and ranking
type RankedSource struct {
	Source      ResearchSource `json:"source"`
	FinalScore  float64        `json:"final_score"`
	FinalRank   int            `json:"final_rank"`
	FromQueries []uuid.UUID    `json:"from_queries"` // Which queries found this
}

// AggregationResult contains the final ranked sources
type AggregationResult struct {
	Sources           []RankedSource `json:"sources"`
	TotalSources      int            `json:"total_sources"` // Before deduplication
	DuplicatesRemoved int            `json:"duplicates_removed"`
	UniqueDomains     int            `json:"unique_domains"`
	QualityScore      float64        `json:"quality_score"` // Overall quality 0-1
}

// Aggregate combines sources from multiple queries using RRF and deduplication
func (a *ResearchAggregatorService) Aggregate(
	ctx context.Context,
	results []ExecutionResult,
	config AggregationConfig,
) (*AggregationResult, error) {
	// Collect all sources
	allSources := make([]ResearchSource, 0)
	sourceToQueries := make(map[string][]uuid.UUID)

	for _, result := range results {
		for _, source := range result.Sources {
			allSources = append(allSources, source)

			// Track which queries found this source
			key := source.ContentHash
			sourceToQueries[key] = append(sourceToQueries[key], result.QueryID)
		}
	}

	totalSources := len(allSources)

	// Step 1: Deduplicate by content hash (exact matches)
	uniqueSources := a.deduplicateByHash(allSources, sourceToQueries)

	// Step 2: Deduplicate by semantic similarity
	uniqueSources, semanticDuplicates := a.deduplicateBySimilarity(ctx, uniqueSources, config.DeduplicationThreshold)

	duplicatesRemoved := totalSources - len(uniqueSources)

	// Step 3: Apply RRF ranking
	rankedSources := a.rankWithRRF(uniqueSources, sourceToQueries, config.RRFConstant)

	// Step 4: Filter by minimum relevance
	filteredSources := a.filterByRelevance(rankedSources, config.MinRelevanceScore)

	// Step 5: Apply diversity constraints
	if config.DiversityEnabled {
		filteredSources = a.ensureDiversity(filteredSources, config.MaxSourcesPerDomain)
	}

	// Step 6: Select top N sources
	if len(filteredSources) > config.MaxSources {
		filteredSources = filteredSources[:config.MaxSources]
	}

	// Step 7: Assign final ranks
	for i := range filteredSources {
		filteredSources[i].FinalRank = i + 1
	}

	// Calculate quality metrics
	uniqueDomains := a.countUniqueDomains(filteredSources)
	qualityScore := a.calculateQualityScore(filteredSources, uniqueDomains)

	return &AggregationResult{
		Sources:           filteredSources,
		TotalSources:      totalSources,
		DuplicatesRemoved: duplicatesRemoved + semanticDuplicates,
		UniqueDomains:     uniqueDomains,
		QualityScore:      qualityScore,
	}, nil
}

// deduplicateByHash removes exact duplicate content using hash
func (a *ResearchAggregatorService) deduplicateByHash(
	sources []ResearchSource,
	sourceToQueries map[string][]uuid.UUID,
) []ResearchSource {
	seen := make(map[string]bool)
	unique := make([]ResearchSource, 0)

	for _, source := range sources {
		if !seen[source.ContentHash] {
			seen[source.ContentHash] = true
			unique = append(unique, source)
		}
	}

	return unique
}

// deduplicateBySimilarity removes semantically similar sources
func (a *ResearchAggregatorService) deduplicateBySimilarity(
	ctx context.Context,
	sources []ResearchSource,
	threshold float64,
) ([]ResearchSource, int) {
	// TODO: Implement embedding-based similarity check
	// For now, return sources as-is
	// In production: generate embeddings, compute cosine similarity, remove duplicates
	return sources, 0
}

// rankWithRRF applies Reciprocal Rank Fusion to combine rankings
func (a *ResearchAggregatorService) rankWithRRF(
	sources []ResearchSource,
	sourceToQueries map[string][]uuid.UUID,
	k int,
) []RankedSource {
	// RRF formula: score = Σ(1 / (k + rank_i))
	// Where rank_i is the rank in query i (1-indexed)

	ranked := make([]RankedSource, len(sources))

	for i, source := range sources {
		// Base score from source's own relevance
		baseScore := source.RelevanceScore

		// RRF bonus for being found by multiple queries
		queries := sourceToQueries[source.ContentHash]
		rrfScore := 0.0
		for range queries {
			// Simplified: assume rank 1 for all (more sophisticated would track actual ranks)
			rrfScore += 1.0 / float64(k+1)
		}

		// Combined score
		finalScore := (baseScore * 0.7) + (rrfScore * 0.3)

		ranked[i] = RankedSource{
			Source:      source,
			FinalScore:  finalScore,
			FromQueries: queries,
		}
	}

	// Sort by final score (descending)
	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].FinalScore > ranked[j].FinalScore
	})

	return ranked
}

// filterByRelevance removes low-quality sources
func (a *ResearchAggregatorService) filterByRelevance(
	sources []RankedSource,
	minScore float64,
) []RankedSource {
	filtered := make([]RankedSource, 0)
	for _, source := range sources {
		if source.FinalScore >= minScore {
			filtered = append(filtered, source)
		}
	}
	return filtered
}

// ensureDiversity limits sources per domain
func (a *ResearchAggregatorService) ensureDiversity(
	sources []RankedSource,
	maxPerDomain int,
) []RankedSource {
	domainCount := make(map[string]int)
	diverse := make([]RankedSource, 0)

	for _, source := range sources {
		domain := ""
		if source.Source.Domain != nil {
			domain = *source.Source.Domain
		}

		if domain == "" || domainCount[domain] < maxPerDomain {
			diverse = append(diverse, source)
			domainCount[domain]++
		}
	}

	return diverse
}

// countUniqueDomains returns the number of unique domains
func (a *ResearchAggregatorService) countUniqueDomains(sources []RankedSource) int {
	domains := make(map[string]bool)
	for _, source := range sources {
		if source.Source.Domain != nil {
			domains[*source.Source.Domain] = true
		}
	}
	return len(domains)
}

// calculateQualityScore computes overall quality metric
func (a *ResearchAggregatorService) calculateQualityScore(sources []RankedSource, uniqueDomains int) float64 {
	if len(sources) == 0 {
		return 0.0
	}

	// Components:
	// - Average relevance score (40%)
	// - Source diversity (30%)
	// - Source count (20%)
	// - Multi-query coverage (10%)

	// Average relevance
	totalRelevance := 0.0
	for _, source := range sources {
		totalRelevance += source.FinalScore
	}
	avgRelevance := totalRelevance / float64(len(sources))

	// Diversity score (0-1)
	diversityScore := float64(uniqueDomains) / float64(len(sources))
	if diversityScore > 1.0 {
		diversityScore = 1.0
	}

	// Source count score (0-1, plateau at 15 sources)
	sourceCountScore := float64(len(sources)) / 15.0
	if sourceCountScore > 1.0 {
		sourceCountScore = 1.0
	}

	// Multi-query coverage (sources found by multiple queries)
	multiQueryCount := 0
	for _, source := range sources {
		if len(source.FromQueries) > 1 {
			multiQueryCount++
		}
	}
	multiQueryScore := float64(multiQueryCount) / float64(len(sources))

	// Weighted combination
	quality := (avgRelevance * 0.4) +
		(diversityScore * 0.3) +
		(sourceCountScore * 0.2) +
		(multiQueryScore * 0.1)

	return quality
}

// SynthesizeReport generates the final research report
func (a *ResearchAggregatorService) SynthesizeReport(
	ctx context.Context,
	query string,
	sources []RankedSource,
	llmService services.LLMService,
) (string, error) {
	// Build synthesis prompt
	prompt := a.buildSynthesisPrompt(query, sources)

	// Set LLM options for synthesis
	llmService.SetOptions(services.LLMOptions{
		Model:       nil,
		Temperature: 0.7,
		MaxTokens:   3000,
	})

	// Generate report using LLM
	response, err := llmService.ChatComplete(ctx, []services.ChatMessage{
		{
			Role:    "user",
			Content: prompt,
		},
	}, synthesisSystemPrompt)

	if err != nil {
		return "", fmt.Errorf("failed to synthesize report: %w", err)
	}

	return response, nil
}

// buildSynthesisPrompt creates the LLM prompt for report generation
func (a *ResearchAggregatorService) buildSynthesisPrompt(query string, sources []RankedSource) string {
	prompt := fmt.Sprintf(`Create a comprehensive research report answering this query:

Query: "%s"

Based on the following sources:

`, query)

	for i, source := range sources {
		sourceNum := i + 1
		prompt += fmt.Sprintf(`[Source %d] %s
`, sourceNum, source.Source.Title)

		if source.Source.URL != nil {
			prompt += fmt.Sprintf(`URL: %s
`, *source.Source.URL)
		}

		prompt += fmt.Sprintf(`Content: %s

`, source.Source.Snippet)
	}

	prompt += `
Generate a well-structured report with:
1. Executive summary
2. Main findings (2-3 sections)
3. Conclusion
4. Sources section with inline citations [1], [2], etc.

Use markdown formatting. Cite sources inline using [1], [2] format.`

	return prompt
}

const synthesisSystemPrompt = `You are a research report writer. Create comprehensive, well-cited research reports.

Guidelines:
- Synthesize information from multiple sources into a coherent narrative
- Cite sources inline using [1], [2], [3] format
- Include a "Sources" section at the end listing all citations
- Use markdown formatting (headers, lists, bold, italics)
- Be objective and factual
- If sources contradict, acknowledge both perspectives
- Focus on answering the research question directly

Format:
## Executive Summary
[2-3 sentences]

## Main Findings
### [Topic 1]
[Content with citations [1], [2]]

### [Topic 2]
[Content with citations]

## Conclusion
[Summary]

## Sources
1. [Title] - [URL] (if available)
2. [Title] - [URL]
...`
