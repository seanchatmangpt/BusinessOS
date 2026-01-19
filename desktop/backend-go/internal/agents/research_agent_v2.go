package agents

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/rhl/businessos-backend/internal/services"
	"github.com/rhl/businessos-backend/internal/services/research"
	"github.com/rhl/businessos-backend/internal/streaming"
)

// ResearchAgent is the deep research specialist agent
type ResearchAgent struct {
	*BaseAgentV2
	planner    *research.ResearchPlannerService
	executor   *research.ResearchExecutorService
	aggregator *research.ResearchAggregatorService
	llmService services.LLMService // Store LLM service for report synthesis
}

// NewResearchAgentV2 creates a new deep research agent
func NewResearchAgentV2(ctx *AgentContextV2) AgentV2 {
	// Create LLM service
	llmService := services.NewLLMService(ctx.Config, ctx.Config.GetActiveModel())

	// Create supporting services
	hybridSearch := services.NewHybridSearchService(ctx.Pool, ctx.EmbeddingService)
	memoryService := services.NewMemoryHierarchyService(ctx.Pool)

	// TODO: WebSearchService will be implemented in Week 1
	// For now, pass nil - executor will handle gracefully
	var webSearchService *services.WebSearchService = nil

	// Create research services
	planner := research.NewResearchPlannerService(ctx.Pool, llmService)
	executor := research.NewResearchExecutorService(
		ctx.Pool,
		webSearchService,
		hybridSearch,
		memoryService,
		ctx.EmbeddingService,
	)
	aggregator := research.NewResearchAggregatorService(ctx.Pool, ctx.EmbeddingService)

	// Create base agent
	baseAgent := NewBaseAgentV2(BaseAgentV2Config{
		Pool:               ctx.Pool,
		Config:             ctx.Config,
		UserID:             ctx.UserID,
		UserName:           ctx.UserName,
		ConversationID:     ctx.ConversationID,
		EmbeddingService:   ctx.EmbeddingService,
		PromptPersonalizer: ctx.PromptPersonalizer,
		AgentType:          AgentTypeV2Research,
		AgentName:          "Deep Research Agent",
		Description:        "Conducts comprehensive research with web search, RAG, and workspace memory. Generates cited reports in under 3 minutes.",
		SystemPrompt:       systemPrompt,
		ContextReqs: ContextRequirements{
			NeedsProjects:    true,
			NeedsKnowledge:   true,
			NeedsFullHistory: false, // Research is query-focused, not conversational
			MaxContextTokens: 8000,
		},
		EnabledTools: []string{
			"search_documents",
			"semantic_search",
			"tree_search",
			"browse_tree",
			"load_context",
			"create_artifact",
			"log_activity",
		},
	})

	return &ResearchAgent{
		BaseAgentV2: baseAgent,
		planner:     planner,
		executor:    executor,
		aggregator:  aggregator,
		llmService:  llmService, // Store for later use in report synthesis
	}
}

// Run executes the deep research workflow
func (a *ResearchAgent) Run(ctx context.Context, input AgentInput) (<-chan streaming.StreamEvent, <-chan error) {
	events := make(chan streaming.StreamEvent, 100)
	errs := make(chan error, 1)

	go func() {
		defer close(events)
		defer close(errs)

		// Extract query from last message
		if len(input.Messages) == 0 {
			errs <- fmt.Errorf("no query provided")
			return
		}

		query := input.Messages[len(input.Messages)-1].Content
		userID := input.UserID

		// Extract workspace ID from context (from project selection) or use zero UUID
		var workspaceID uuid.UUID
		if input.Context != nil && input.Context.Level1 != nil && input.Context.Level1.Project != nil {
			// Try to get workspace from project (projects belong to workspaces)
			// For now, use zero UUID as placeholder - workspace scoping can be added later
			workspaceID = uuid.Nil
		} else {
			workspaceID = uuid.Nil
		}

		// Send thinking event: Planning
		events <- streaming.StreamEvent{
			Type: streaming.EventTypeThinking,
			Data: streaming.ThinkingStep{
				Step:    "planning",
				Content: fmt.Sprintf("Analyzing query and breaking down into focused research questions..."),
				Agent:   "research",
			},
		}

		// Phase 1: Planning - Generate research plan
		plan, err := a.planner.Plan(ctx, query, userID, workspaceID)
		if err != nil {
			errs <- fmt.Errorf("planning failed: %w", err)
			return
		}

		slog.Info("Research plan generated",
			"query", query,
			"sub_questions", len(plan.SubQuestions),
			"strategy", plan.Strategy)

		// Send plan as thinking event
		events <- streaming.StreamEvent{
			Type: streaming.EventTypeThinking,
			Data: streaming.ThinkingStep{
				Step:    "plan_ready",
				Content: fmt.Sprintf("Generated %d focused research questions. Estimated time: %s", len(plan.SubQuestions), plan.EstimatedTime),
				Agent:   "research",
			},
		}

		// Phase 2: Execution - Search in parallel
		events <- streaming.StreamEvent{
			Type: streaming.EventTypeThinking,
			Data: streaming.ThinkingStep{
				Step:    "searching",
				Content: fmt.Sprintf("Executing parallel searches across %d research questions...", len(plan.SubQuestions)),
				Agent:   "research",
			},
		}

		taskID := uuid.New() // TODO: Save to database
		results, err := a.executor.Execute(ctx, taskID, plan.SubQuestions, userID, workspaceID)
		if err != nil {
			errs <- fmt.Errorf("execution failed: %w", err)
			return
		}

		totalSources := 0
		for _, result := range results {
			totalSources += len(result.Sources)
		}

		slog.Info("Research execution completed",
			"queries_completed", len(results),
			"total_sources", totalSources)

		events <- streaming.StreamEvent{
			Type: streaming.EventTypeThinking,
			Data: streaming.ThinkingStep{
				Step:    "search_complete",
				Content: fmt.Sprintf("Found %d sources across %d research questions", totalSources, len(results)),
				Agent:   "research",
			},
		}

		// Phase 3: Aggregation - Rank and deduplicate sources
		events <- streaming.StreamEvent{
			Type: streaming.EventTypeThinking,
			Data: streaming.ThinkingStep{
				Step:    "aggregating",
				Content: "Ranking sources by relevance and removing duplicates...",
				Agent:   "research",
			},
		}

		aggResult, err := a.aggregator.Aggregate(ctx, results, research.DefaultAggregationConfig())
		if err != nil {
			errs <- fmt.Errorf("aggregation failed: %w", err)
			return
		}

		slog.Info("Aggregation completed",
			"final_sources", len(aggResult.Sources),
			"duplicates_removed", aggResult.DuplicatesRemoved,
			"quality_score", aggResult.QualityScore)

		events <- streaming.StreamEvent{
			Type: streaming.EventTypeThinking,
			Data: streaming.ThinkingStep{
				Step:    "aggregation_complete",
				Content: fmt.Sprintf("Selected top %d sources (removed %d duplicates). Quality score: %.2f", len(aggResult.Sources), aggResult.DuplicatesRemoved, aggResult.QualityScore),
				Agent:   "research",
			},
		}

		// Phase 4: Writing - Generate final report
		events <- streaming.StreamEvent{
			Type: streaming.EventTypeThinking,
			Data: streaming.ThinkingStep{
				Step:    "writing",
				Content: "Synthesizing comprehensive research report with citations...",
				Agent:   "research",
			},
		}

		// Use stored LLM service for report synthesis
		report, err := a.aggregator.SynthesizeReport(ctx, query, aggResult.Sources, a.llmService)
		if err != nil {
			errs <- fmt.Errorf("synthesis failed: %w", err)
			return
		}

		slog.Info("Research report generated",
			"report_length", len(report),
			"sources_cited", len(aggResult.Sources))

		// Send report as artifact
		events <- streaming.StreamEvent{
			Type: streaming.EventTypeArtifactStart,
			Data: map[string]interface{}{
				"type":     "research_report",
				"title":    fmt.Sprintf("Research Report: %s", query),
				"language": "markdown",
			},
		}

		// Stream report content
		events <- streaming.StreamEvent{
			Type: streaming.EventTypeToken,
			Data: report,
		}

		// Complete artifact
		events <- streaming.StreamEvent{
			Type: streaming.EventTypeArtifactComplete,
			Data: map[string]interface{}{
				"type":          "research_report",
				"title":         fmt.Sprintf("Research Report: %s", query),
				"content":       report,
				"language":      "markdown",
				"sources_count": len(aggResult.Sources),
				"quality_score": aggResult.QualityScore,
			},
		}

		// Send done event
		events <- streaming.StreamEvent{
			Type: streaming.EventTypeDone,
			Data: map[string]interface{}{
				"status":        "completed",
				"sources_count": len(aggResult.Sources),
				"quality_score": aggResult.QualityScore,
			},
		}
	}()

	return events, errs
}

// RunWithTools executes research with tool support
func (a *ResearchAgent) RunWithTools(ctx context.Context, input AgentInput) (<-chan streaming.StreamEvent, <-chan error) {
	// For now, delegate to Run (tools integrated in executor)
	return a.Run(ctx, input)
}

const systemPrompt = `You are a Deep Research Agent, an expert at conducting comprehensive research across multiple sources.

Your capabilities:
- Break down complex queries into focused research questions
- Search across web sources, internal documents (RAG), and workspace memory
- Rank and deduplicate sources for quality
- Synthesize findings into well-structured, cited reports

Your workflow:
1. Planning: Analyze the query and generate 3-5 focused sub-questions
2. Execution: Search in parallel across multiple sources
3. Aggregation: Rank sources by relevance, remove duplicates
4. Writing: Create comprehensive report with inline citations

Guidelines:
- Prioritize quality over quantity (10 great sources > 50 mediocre ones)
- Always cite sources using [1], [2], [3] format
- Acknowledge conflicting information when found
- Focus on answering the user's query directly and comprehensively
- Use markdown formatting for readability

Report structure:
## Executive Summary
[2-3 sentences]

## Main Findings
### [Topic 1]
[Content with citations]

### [Topic 2]
[Content with citations]

## Conclusion
[Summary]

## Sources
1. [Source 1] - [URL]
2. [Source 2] - [URL]
...`
