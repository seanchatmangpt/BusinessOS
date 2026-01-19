package research

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/services"
)

// ResearchPlannerService breaks down user queries into focused sub-questions
type ResearchPlannerService struct {
	pool       *pgxpool.Pool
	llmService services.LLMService
}

// NewResearchPlannerService creates a new research planner
func NewResearchPlannerService(pool *pgxpool.Pool, llmService services.LLMService) *ResearchPlannerService {
	return &ResearchPlannerService{
		pool:       pool,
		llmService: llmService,
	}
}

// SearchStrategy determines how to approach the research
type SearchStrategy string

const (
	StrategyWebHeavy    SearchStrategy = "web_heavy"    // Emphasize external sources
	StrategyLocalHeavy  SearchStrategy = "local_heavy"  // Emphasize internal knowledge
	StrategyHybrid      SearchStrategy = "hybrid"       // Balanced approach
	StrategyMemoryFirst SearchStrategy = "memory_first" // Start with workspace knowledge
)

// SearchType specifies where to search for a sub-question
type SearchType string

const (
	SearchTypeWeb    SearchType = "web"
	SearchTypeRAG    SearchType = "rag"
	SearchTypeMemory SearchType = "memory"
	SearchTypeHybrid SearchType = "hybrid"
)

// SubQuestion represents a focused research question
type SubQuestion struct {
	ID           uuid.UUID   `json:"id"`
	Question     string      `json:"question"`
	SearchType   SearchType  `json:"search_type"`
	Weight       float64     `json:"weight"`       // Importance 0.0-1.0
	OrderNum     int         `json:"order_num"`    // Execution order
	Dependencies []uuid.UUID `json:"dependencies"` // Which questions must complete first
	Reasoning    string      `json:"reasoning"`    // Why this question is important
}

// ResearchPlan contains the decomposed research strategy
type ResearchPlan struct {
	OriginalQuery     string         `json:"original_query"`
	SubQuestions      []SubQuestion  `json:"sub_questions"`
	Strategy          SearchStrategy `json:"strategy"`
	EstimatedTime     time.Duration  `json:"estimated_time"`
	SearchPriority    int            `json:"search_priority"` // 1-10, how important is external search
	RequiresWebSearch bool           `json:"requires_web_search"`
	Reasoning         string         `json:"reasoning"` // Overall plan reasoning
}

// Plan generates a research plan from a user query
func (p *ResearchPlannerService) Plan(ctx context.Context, query string, userID string, workspaceID uuid.UUID) (*ResearchPlan, error) {
	// Build planning prompt
	prompt := p.buildPlanningPrompt(query)

	// Set LLM options for planning
	p.llmService.SetOptions(services.LLMOptions{
		Model:       nil, // Use default
		Temperature: 0.3, // Low temperature for consistent decomposition
		MaxTokens:   1500,
	})

	// Generate plan using LLM
	response, err := p.llmService.ChatComplete(ctx, []services.ChatMessage{
		{
			Role:    "user",
			Content: prompt,
		},
	}, systemPrompt)

	if err != nil {
		return nil, fmt.Errorf("failed to generate research plan: %w", err)
	}

	// Parse LLM response into structured plan
	plan, err := p.parsePlanResponse(response, query)
	if err != nil {
		// Fallback to template-based plan if parsing fails
		return p.generateTemplatePlan(query), nil
	}

	// Validate plan
	if err := p.validatePlan(plan); err != nil {
		return nil, fmt.Errorf("invalid research plan: %w", err)
	}

	return plan, nil
}

// buildPlanningPrompt creates the LLM prompt for research planning
func (p *ResearchPlannerService) buildPlanningPrompt(query string) string {
	return fmt.Sprintf(`Analyze this research query and create a comprehensive research plan:

Query: "%s"

Generate 3-5 focused sub-questions that will comprehensively answer this query. For each sub-question:
1. Make it specific and answerable
2. Determine if it needs web search, internal documents (RAG), workspace memory, or hybrid
3. Assign importance weight (0.0-1.0)
4. Specify if it depends on other questions

Return your response as JSON with this structure:
{
  "sub_questions": [
    {
      "question": "...",
      "search_type": "web|rag|memory|hybrid",
      "weight": 0.8,
      "order_num": 1,
      "dependencies": [],
      "reasoning": "why this question is important"
    }
  ],
  "strategy": "web_heavy|local_heavy|hybrid|memory_first",
  "requires_web_search": true,
  "reasoning": "overall plan reasoning"
}

Focus on quality over quantity. Each sub-question should add unique value.`, query)
}

// parsePlanResponse extracts structured plan from LLM response
func (p *ResearchPlannerService) parsePlanResponse(response string, originalQuery string) (*ResearchPlan, error) {
	var planData struct {
		SubQuestions []struct {
			Question     string   `json:"question"`
			SearchType   string   `json:"search_type"`
			Weight       float64  `json:"weight"`
			OrderNum     int      `json:"order_num"`
			Dependencies []string `json:"dependencies"`
			Reasoning    string   `json:"reasoning"`
		} `json:"sub_questions"`
		Strategy          string `json:"strategy"`
		RequiresWebSearch bool   `json:"requires_web_search"`
		Reasoning         string `json:"reasoning"`
	}

	if err := json.Unmarshal([]byte(response), &planData); err != nil {
		return nil, fmt.Errorf("failed to parse plan JSON: %w", err)
	}

	// Convert to ResearchPlan
	plan := &ResearchPlan{
		OriginalQuery:     originalQuery,
		Strategy:          SearchStrategy(planData.Strategy),
		RequiresWebSearch: planData.RequiresWebSearch,
		Reasoning:         planData.Reasoning,
		SubQuestions:      make([]SubQuestion, len(planData.SubQuestions)),
	}

	for i, sq := range planData.SubQuestions {
		plan.SubQuestions[i] = SubQuestion{
			ID:           uuid.New(),
			Question:     sq.Question,
			SearchType:   SearchType(sq.SearchType),
			Weight:       sq.Weight,
			OrderNum:     sq.OrderNum,
			Dependencies: []uuid.UUID{}, // TODO: Parse dependencies
			Reasoning:    sq.Reasoning,
		}
	}

	// Estimate time (30s per sub-question + 30s overhead)
	plan.EstimatedTime = time.Duration(len(plan.SubQuestions)*30+30) * time.Second

	return plan, nil
}

// generateTemplatePlan creates a fallback plan if LLM fails
func (p *ResearchPlannerService) generateTemplatePlan(query string) *ResearchPlan {
	return &ResearchPlan{
		OriginalQuery:     query,
		Strategy:          StrategyHybrid,
		RequiresWebSearch: true,
		Reasoning:         "Fallback template plan generated due to parsing error",
		SubQuestions: []SubQuestion{
			{
				ID:         uuid.New(),
				Question:   query, // Use original query
				SearchType: SearchTypeHybrid,
				Weight:     1.0,
				OrderNum:   1,
				Reasoning:  "Primary research question",
			},
			{
				ID:         uuid.New(),
				Question:   query + " latest developments",
				SearchType: SearchTypeWeb,
				Weight:     0.7,
				OrderNum:   2,
				Reasoning:  "Recent updates and news",
			},
			{
				ID:         uuid.New(),
				Question:   query + " best practices",
				SearchType: SearchTypeRAG,
				Weight:     0.6,
				OrderNum:   3,
				Reasoning:  "Internal knowledge and documentation",
			},
		},
		EstimatedTime: 120 * time.Second,
	}
}

// validatePlan ensures the plan is well-formed
func (p *ResearchPlannerService) validatePlan(plan *ResearchPlan) error {
	if len(plan.SubQuestions) == 0 {
		return fmt.Errorf("plan must have at least one sub-question")
	}

	if len(plan.SubQuestions) > 10 {
		return fmt.Errorf("plan has too many sub-questions (max 10): %d", len(plan.SubQuestions))
	}

	// Validate each sub-question
	for i, sq := range plan.SubQuestions {
		if sq.Question == "" {
			return fmt.Errorf("sub-question %d has empty question", i)
		}
		if sq.Weight < 0 || sq.Weight > 1 {
			return fmt.Errorf("sub-question %d has invalid weight: %f", i, sq.Weight)
		}
	}

	return nil
}

// System prompt for research planning
const systemPrompt = `You are an expert research planner. Your role is to break down complex research queries into focused, answerable sub-questions.

Guidelines:
- Generate 3-5 high-quality sub-questions (quality over quantity)
- Each sub-question should be specific and actionable
- Consider: What information is needed? Where is it likely found?
- Balance web search (external), RAG (internal docs), and memory (workspace knowledge)
- Assign weights based on importance to answering the original query
- Specify dependencies only when truly necessary (most questions can be parallel)

Search types:
- "web": External web search required (news, current events, public knowledge)
- "rag": Internal document search (company docs, uploaded files)
- "memory": Workspace knowledge (past conversations, saved insights)
- "hybrid": Combination of sources

Return ONLY valid JSON, no markdown formatting.`
