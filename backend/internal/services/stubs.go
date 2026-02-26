package services

// This file provides stub types that are referenced across the services package
// but do not have a dedicated implementation file.

import "github.com/google/uuid"

// ─── Delegation ──────────────────────────────────────────────────────────────

// DelegationTarget represents a custom agent that can receive delegated tasks.
type DelegationTarget struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	DisplayName   string    `json:"display_name"`
	Description   string    `json:"description"`
	Capabilities  []string  `json:"capabilities"`
	Category      string    `json:"category"`
	ModelOverride string    `json:"model_override,omitempty"`
	SystemPrompt  string    `json:"system_prompt,omitempty"`
	IsSystemAgent bool      `json:"is_system_agent"`
}

// DelegationService resolves and delegates tasks to custom agents.
type DelegationService struct {
	pool interface{} // pgxpool.Pool — kept as interface{} to avoid import cycle
}

// NewDelegationService returns a new DelegationService.
// The pool is typed as any to keep this stub free of direct dependencies;
// the real implementation should accept *pgxpool.Pool.
func NewDelegationService(pool interface{}) *DelegationService {
	return &DelegationService{pool: pool}
}

// ─── Agentic RAG ─────────────────────────────────────────────────────────────

// AgenticRAGRequest is the input for an agentic retrieval-augmented generation query.
type AgenticRAGRequest struct {
	Query              string     `json:"query"`
	UserID             string     `json:"user_id"`
	MaxResults         int        `json:"max_results,omitempty"`
	MinQualityScore    float64    `json:"min_quality_score,omitempty"`
	ProjectContext     *uuid.UUID `json:"project_id,omitempty"`
	TaskContext        *uuid.UUID `json:"task_id,omitempty"`
	UsePersonalization bool       `json:"use_personalization,omitempty"`
}

// AgenticRAGResponse is the output of an agentic RAG query.
type AgenticRAGResponse struct {
	Answer   string        `json:"answer"`
	Sources  []interface{} `json:"sources,omitempty"`
	Metadata interface{}   `json:"metadata,omitempty"`
}

// AgenticRAGService performs intelligent adaptive retrieval.
type AgenticRAGService struct{}

// Retrieve performs agentic RAG retrieval (stub — returns empty response).
func (s *AgenticRAGService) Retrieve(_ interface{}, _ AgenticRAGRequest) (*AgenticRAGResponse, error) {
	return &AgenticRAGResponse{}, nil
}

// SetCache wires a RAG cache into the service (no-op stub).
func (s *AgenticRAGService) SetCache(_ *RAGCacheService) {}

// SetQueryExpansion wires a query expansion service (no-op stub).
func (s *AgenticRAGService) SetQueryExpansion(_ *QueryExpansionService) {}

// NewAgenticRAGService creates a stub AgenticRAGService.
// Parameters are accepted for API compatibility but are not used.
func NewAgenticRAGService(_, _, _, _, _ interface{}) *AgenticRAGService {
	return &AgenticRAGService{}
}

// ─── Query Intent ─────────────────────────────────────────────────────────────

// QueryIntent classifies the intent behind a user query.
type QueryIntent string

const (
	IntentFactualLookup QueryIntent = "factual_lookup"
	IntentProcedural    QueryIntent = "procedural"
	IntentComparison    QueryIntent = "comparison"
	IntentExplanatory   QueryIntent = "explanatory"
	IntentRecent        QueryIntent = "recent"
	IntentExhaustive    QueryIntent = "exhaustive"
	IntentUnknown       QueryIntent = "unknown"
)

// ─── OSA Sync ────────────────────────────────────────────────────────────────

// OSASyncService is a no-op stub retained for signature compatibility with
// onboarding_service.go. The OSA integration has been removed from this
// open-source release; callers should always pass nil.
type OSASyncService struct{}

// SyncUser is a no-op stub.
func (s *OSASyncService) SyncUser(_ interface{}, _ uuid.UUID) error { return nil }

// SyncWorkspace is a no-op stub.
func (s *OSASyncService) SyncWorkspace(_ interface{}, _ uuid.UUID) error { return nil }

// ─── Learning ────────────────────────────────────────────────────────────────

// LearningConversationContext carries the context of a conversation turn for
// the auto-learning system.
type LearningConversationContext struct {
	UserID         string
	WorkspaceID    *uuid.UUID
	ConversationID uuid.UUID
	UserMessage    string
	AgentResponse  string
	AgentType      string
	FocusMode      string
	ProjectID      *uuid.UUID
	NodeID         *uuid.UUID
	ContextIDs     []uuid.UUID
	Timestamp      interface{}
}

// LearningService provides feedback-based personalisation and pattern detection.
// This is a stub for the open-source release.
type LearningService struct{}

// NewLearningService returns a stub LearningService.
func NewLearningService(_ interface{}) *LearningService { return &LearningService{} }

// BackfillRecentUsersBehaviorPatterns is a no-op stub.
func (s *LearningService) BackfillRecentUsersBehaviorPatterns(_ interface{}, _ int) (int, int, error) {
	return 0, 0, nil
}

// HealthCheck returns true (stub is always healthy).
func (s *LearningService) HealthCheck(_ interface{}) bool { return true }

// AutoLearningTriggers fires learning events after each conversation turn.
// This is a stub for the open-source release.
type AutoLearningTriggers struct{}

// NewAutoLearningTriggers returns a stub AutoLearningTriggers.
func NewAutoLearningTriggers(_, _, _ interface{}) *AutoLearningTriggers {
	return &AutoLearningTriggers{}
}

// ProcessConversationTurn is a no-op stub.
func (t *AutoLearningTriggers) ProcessConversationTurn(_ interface{}, _ LearningConversationContext) {
}

// ─── Skills Loader ──────────────────────────────────────────────────────────

// SkillsLoader loads agent skills from a YAML configuration file.
// This is a stub for the open-source release.
type SkillsLoader struct{}

// NewSkillsLoader returns a stub SkillsLoader.
func NewSkillsLoader(_ string) *SkillsLoader { return &SkillsLoader{} }

// LoadConfig is a no-op stub.
func (s *SkillsLoader) LoadConfig() error { return nil }

// IsLoaded always returns false in the stub (no skills file).
func (s *SkillsLoader) IsLoaded() bool { return false }

// GetEnabledSkills returns an empty slice.
func (s *SkillsLoader) GetEnabledSkills() []interface{} { return nil }

// GetSkillsPromptXML returns an empty string.
func (s *SkillsLoader) GetSkillsPromptXML() string { return "" }

// ─── Prompt Personalization ──────────────────────────────────────────────────

// PromptPersonalizer enriches agent prompts with user-specific context.
// This is a stub for the open-source release; the full implementation uses
// learning history, memory, and embedding services to personalize prompts.
type PromptPersonalizer struct{}

// NewPromptPersonalizer returns a stub PromptPersonalizer.
// Parameters are accepted for API compatibility but are not used.
func NewPromptPersonalizer(_, _, _, _ interface{}) *PromptPersonalizer {
	return &PromptPersonalizer{}
}

// BuildPersonalizedPrompt returns the base prompt unchanged in this stub.
func (p *PromptPersonalizer) BuildPersonalizedPrompt(_ interface{}, _ string, basePrompt string, _ string) (string, error) {
	return basePrompt, nil
}
