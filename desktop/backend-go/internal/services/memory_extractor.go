package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MemoryExtractorService automatically extracts memories from conversations and content
type MemoryExtractorService struct {
	pool       *pgxpool.Pool
	logger     *slog.Logger
	llmService LLMService // Optional LLM service for enhanced extraction
}

// ExtractedMemory represents a memory extracted from content
type ExtractedMemory struct {
	ID          string                 `json:"id"`
	UserID      string                 `json:"user_id"`
	Type        MemoryType             `json:"type"`
	Content     string                 `json:"content"`
	Summary     string                 `json:"summary"`
	Source      MemorySource           `json:"source"`
	SourceID    string                 `json:"source_id,omitempty"`
	Confidence  float64                `json:"confidence"`
	Tags        []string               `json:"tags"`
	Entities    []string               `json:"entities"`
	RelatedTo   []string               `json:"related_to"`
	Context     string                 `json:"context,omitempty"`
	Importance  int                    `json:"importance"` // 1-10
	Metadata    map[string]interface{} `json:"metadata"`
	ExtractedAt time.Time              `json:"extracted_at"`
}

// MemoryType represents the type of extracted memory
type MemoryType string

const (
	MemoryTypeFact       MemoryType = "fact"
	MemoryTypePreference MemoryType = "preference"
	MemoryTypeDecision   MemoryType = "decision"
	MemoryTypeTask       MemoryType = "task"
	MemoryTypeReminder   MemoryType = "reminder"
	MemoryTypeInsight    MemoryType = "insight"
	MemoryTypeContact    MemoryType = "contact"
	MemoryTypeEvent      MemoryType = "event"
	MemoryTypeNote       MemoryType = "note"
	MemoryTypeCode       MemoryType = "code"
	MemoryTypeError      MemoryType = "error"
	MemoryTypeSolution   MemoryType = "solution"
)

// MemorySource represents the source of the memory
type MemorySource string

const (
	MemorySourceConversation MemorySource = "conversation"
	MemorySourceVoiceNote    MemorySource = "voice_note"
	MemorySourceDocument     MemorySource = "document"
	MemorySourceCode         MemorySource = "code"
	MemorySourceManual       MemorySource = "manual"
	MemorySourceImport       MemorySource = "import"
)

// ExtractionResult contains the results of memory extraction
type ExtractionResult struct {
	Memories        []ExtractedMemory `json:"memories"`
	TotalExtracted  int               `json:"total_extracted"`
	ByType          map[string]int    `json:"by_type"`
	ProcessingTime  string            `json:"processing_time"`
	SourceProcessed string            `json:"source_processed"`
}

// ExtractionOptions configures memory extraction behavior
type ExtractionOptions struct {
	ExtractFacts       bool    `json:"extract_facts"`
	ExtractPreferences bool    `json:"extract_preferences"`
	ExtractDecisions   bool    `json:"extract_decisions"`
	ExtractTasks       bool    `json:"extract_tasks"`
	ExtractInsights    bool    `json:"extract_insights"`
	ExtractContacts    bool    `json:"extract_contacts"`
	ExtractCode        bool    `json:"extract_code"`
	MinConfidence      float64 `json:"min_confidence"`
	MaxMemories        int     `json:"max_memories"`
}

// DefaultExtractionOptions returns default extraction options
func DefaultExtractionOptions() *ExtractionOptions {
	return &ExtractionOptions{
		ExtractFacts:       true,
		ExtractPreferences: true,
		ExtractDecisions:   true,
		ExtractTasks:       true,
		ExtractInsights:    true,
		ExtractContacts:    true,
		ExtractCode:        true,
		MinConfidence:      0.5,
		MaxMemories:        50,
	}
}

// NewMemoryExtractorService creates a new memory extractor service
func NewMemoryExtractorService(pool *pgxpool.Pool, embeddingService *EmbeddingService) *MemoryExtractorService {
	return &MemoryExtractorService{
		pool:   pool,
		logger: slog.Default(),
	}
}

// SetLLMService sets the LLM service for enhanced extraction
func (s *MemoryExtractorService) SetLLMService(llm LLMService) {
	s.llmService = llm
}

// ExtractFromConversation extracts memories from a conversation
func (s *MemoryExtractorService) ExtractFromConversation(ctx context.Context, userID string, messages []Message, opts *ExtractionOptions) (*ExtractionResult, error) {
	startTime := time.Now()

	if opts == nil {
		opts = DefaultExtractionOptions()
	}

	result := &ExtractionResult{
		Memories:        make([]ExtractedMemory, 0),
		ByType:          make(map[string]int),
		SourceProcessed: "conversation",
	}

	// Combine all message content for analysis
	var fullContent strings.Builder
	for _, msg := range messages {
		fullContent.WriteString(fmt.Sprintf("[%s]: %s\n\n", msg.Role, msg.Content))
	}
	content := fullContent.String()

	// Extract different types of memories
	if opts.ExtractFacts {
		facts := s.extractFacts(userID, content, messages)
		result.Memories = append(result.Memories, facts...)
	}

	if opts.ExtractPreferences {
		prefs := s.extractPreferences(userID, content, messages)
		result.Memories = append(result.Memories, prefs...)
	}

	if opts.ExtractDecisions {
		decisions := s.extractDecisionsFromContent(userID, content, messages)
		result.Memories = append(result.Memories, decisions...)
	}

	if opts.ExtractTasks {
		tasks := s.extractTasks(userID, content, messages)
		result.Memories = append(result.Memories, tasks...)
	}

	if opts.ExtractInsights {
		insights := s.extractInsights(userID, content, messages)
		result.Memories = append(result.Memories, insights...)
	}

	if opts.ExtractContacts {
		contacts := s.extractContacts(userID, content)
		result.Memories = append(result.Memories, contacts...)
	}

	if opts.ExtractCode {
		codeMemories := s.extractCodeMemories(userID, content, messages)
		result.Memories = append(result.Memories, codeMemories...)
	}

	// Filter by confidence
	filtered := make([]ExtractedMemory, 0)
	for _, m := range result.Memories {
		if m.Confidence >= opts.MinConfidence {
			filtered = append(filtered, m)
		}
	}
	result.Memories = filtered

	// Limit results
	if opts.MaxMemories > 0 && len(result.Memories) > opts.MaxMemories {
		result.Memories = result.Memories[:opts.MaxMemories]
	}

	// Calculate stats
	result.TotalExtracted = len(result.Memories)
	for _, m := range result.Memories {
		result.ByType[string(m.Type)]++
	}
	result.ProcessingTime = time.Since(startTime).String()

	// Save extracted memories
	for _, memory := range result.Memories {
		if err := s.saveMemory(ctx, &memory); err != nil {
			s.logger.Warn("failed to save extracted memory", "error", err)
		}
	}

	return result, nil
}

// ExtractFromVoiceNote extracts memories from transcribed voice note
func (s *MemoryExtractorService) ExtractFromVoiceNote(ctx context.Context, userID, transcript string, opts *ExtractionOptions) (*ExtractionResult, error) {
	startTime := time.Now()

	if opts == nil {
		opts = DefaultExtractionOptions()
	}

	result := &ExtractionResult{
		Memories:        make([]ExtractedMemory, 0),
		ByType:          make(map[string]int),
		SourceProcessed: "voice_note",
	}

	// Convert transcript to pseudo-messages for processing
	messages := []Message{{
		Role:      "user",
		Content:   transcript,
		Timestamp: time.Now(),
	}}

	// Extract memories (similar to conversation but with voice-specific patterns)
	if opts.ExtractFacts {
		facts := s.extractFacts(userID, transcript, messages)
		for i := range facts {
			facts[i].Source = MemorySourceVoiceNote
		}
		result.Memories = append(result.Memories, facts...)
	}

	if opts.ExtractTasks {
		tasks := s.extractTasksFromVoice(userID, transcript)
		result.Memories = append(result.Memories, tasks...)
	}

	if opts.ExtractInsights {
		insights := s.extractInsights(userID, transcript, messages)
		for i := range insights {
			insights[i].Source = MemorySourceVoiceNote
		}
		result.Memories = append(result.Memories, insights...)
	}

	// Extract reminders (voice notes often contain reminders)
	reminders := s.extractReminders(userID, transcript)
	result.Memories = append(result.Memories, reminders...)

	// Filter and limit
	filtered := make([]ExtractedMemory, 0)
	for _, m := range result.Memories {
		if m.Confidence >= opts.MinConfidence {
			filtered = append(filtered, m)
		}
	}
	result.Memories = filtered

	if opts.MaxMemories > 0 && len(result.Memories) > opts.MaxMemories {
		result.Memories = result.Memories[:opts.MaxMemories]
	}

	// Calculate stats
	result.TotalExtracted = len(result.Memories)
	for _, m := range result.Memories {
		result.ByType[string(m.Type)]++
	}
	result.ProcessingTime = time.Since(startTime).String()

	// Save memories
	for _, memory := range result.Memories {
		if err := s.saveMemory(ctx, &memory); err != nil {
			s.logger.Warn("failed to save extracted memory", "error", err)
		}
	}

	return result, nil
}

// extractFacts extracts factual information
func (s *MemoryExtractorService) extractFacts(userID, content string, messages []Message) []ExtractedMemory {
	memories := make([]ExtractedMemory, 0)

	// Patterns for facts
	factPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(?:i am|i'm|my name is|i work at|i work as|i live in|my job is)\s+([^.!?\n]+)`),
		regexp.MustCompile(`(?i)(?:the (?:answer|solution|result) is)\s+([^.!?\n]+)`),
		regexp.MustCompile(`(?i)(?:we use|our team uses|the project uses)\s+([^.!?\n]+)`),
		regexp.MustCompile(`(?i)(?:the (?:api|endpoint|url|port|host) is)\s+([^.!?\n]+)`),
		regexp.MustCompile(`(?i)(?:the password is|the key is|the secret is)\s+([^.!?\n]+)`),
	}

	for _, pattern := range factPatterns {
		matches := pattern.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) > 1 {
				fact := strings.TrimSpace(match[1])
				if len(fact) < 5 || len(fact) > 500 {
					continue
				}

				memories = append(memories, ExtractedMemory{
					ID:          uuid.New().String(),
					UserID:      userID,
					Type:        MemoryTypeFact,
					Content:     match[0],
					Summary:     fact,
					Source:      MemorySourceConversation,
					Confidence:  0.7,
					Tags:        s.extractTags(fact),
					Entities:    s.extractEntitiesFromText(fact),
					Importance:  5,
					Metadata:    make(map[string]interface{}),
					ExtractedAt: time.Now(),
				})
			}
		}
	}

	return s.deduplicateMemories(memories)
}

// extractPreferences extracts user preferences
func (s *MemoryExtractorService) extractPreferences(userID, content string, messages []Message) []ExtractedMemory {
	memories := make([]ExtractedMemory, 0)

	prefPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(?:i (?:prefer|like|love|enjoy|hate|dislike|want|need))\s+([^.!?\n]+)`),
		regexp.MustCompile(`(?i)(?:i (?:always|usually|typically|never))\s+([^.!?\n]+)`),
		regexp.MustCompile(`(?i)(?:my favorite|my preferred)\s+([^.!?\n]+)`),
		regexp.MustCompile(`(?i)(?:i'd rather|i would prefer)\s+([^.!?\n]+)`),
	}

	for _, pattern := range prefPatterns {
		matches := pattern.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) > 1 {
				pref := strings.TrimSpace(match[1])
				if len(pref) < 5 || len(pref) > 300 {
					continue
				}

				memories = append(memories, ExtractedMemory{
					ID:          uuid.New().String(),
					UserID:      userID,
					Type:        MemoryTypePreference,
					Content:     match[0],
					Summary:     pref,
					Source:      MemorySourceConversation,
					Confidence:  0.8,
					Tags:        s.extractTags(pref),
					Entities:    s.extractEntitiesFromText(pref),
					Importance:  6,
					Metadata:    make(map[string]interface{}),
					ExtractedAt: time.Now(),
				})
			}
		}
	}

	return s.deduplicateMemories(memories)
}

// extractDecisionsFromContent extracts decisions from content
func (s *MemoryExtractorService) extractDecisionsFromContent(userID, content string, messages []Message) []ExtractedMemory {
	memories := make([]ExtractedMemory, 0)

	decisionPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(?:decided to|will use|going with|chose|selected)\s+([^.!?\n]+)`),
		regexp.MustCompile(`(?i)(?:let's go with|we'll use)\s+([^.!?\n]+)`),
		regexp.MustCompile(`(?i)(?:the decision is|agreed to)\s+([^.!?\n]+)`),
	}

	for _, pattern := range decisionPatterns {
		matches := pattern.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) > 1 {
				decision := strings.TrimSpace(match[1])
				if len(decision) < 5 || len(decision) > 500 {
					continue
				}

				memories = append(memories, ExtractedMemory{
					ID:          uuid.New().String(),
					UserID:      userID,
					Type:        MemoryTypeDecision,
					Content:     match[0],
					Summary:     decision,
					Source:      MemorySourceConversation,
					Confidence:  0.75,
					Tags:        append(s.extractTags(decision), "decision"),
					Entities:    s.extractEntitiesFromText(decision),
					Importance:  7,
					Metadata:    make(map[string]interface{}),
					ExtractedAt: time.Now(),
				})
			}
		}
	}

	return s.deduplicateMemories(memories)
}

// extractTasks extracts tasks from content
func (s *MemoryExtractorService) extractTasks(userID, content string, messages []Message) []ExtractedMemory {
	memories := make([]ExtractedMemory, 0)

	taskPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(?:need to|should|must|have to|todo:|task:)\s+([^.!?\n]+)`),
		regexp.MustCompile(`(?i)- \[ \]\s+([^\n]+)`),
		regexp.MustCompile(`(?i)(?:implement|create|add|fix|update|remove)\s+([^.!?\n]+)`),
		regexp.MustCompile(`(?i)(?:don't forget to|remember to|make sure to)\s+([^.!?\n]+)`),
	}

	for _, pattern := range taskPatterns {
		matches := pattern.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) > 1 {
				task := strings.TrimSpace(match[1])
				if len(task) < 5 || len(task) > 300 {
					continue
				}

				memories = append(memories, ExtractedMemory{
					ID:          uuid.New().String(),
					UserID:      userID,
					Type:        MemoryTypeTask,
					Content:     match[0],
					Summary:     task,
					Source:      MemorySourceConversation,
					Confidence:  0.7,
					Tags:        append(s.extractTags(task), "task"),
					Entities:    s.extractEntitiesFromText(task),
					Importance:  s.inferTaskImportance(task),
					Metadata:    map[string]interface{}{"status": "pending"},
					ExtractedAt: time.Now(),
				})
			}
		}
	}

	return s.deduplicateMemories(memories)
}

// extractTasksFromVoice extracts tasks from voice notes
func (s *MemoryExtractorService) extractTasksFromVoice(userID, transcript string) []ExtractedMemory {
	memories := make([]ExtractedMemory, 0)

	// Voice-specific patterns (more conversational)
	voiceTaskPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(?:i need to|i have to|i should|i must)\s+([^.!?\n]+)`),
		regexp.MustCompile(`(?i)(?:gotta|gonna|need to)\s+([^.!?\n]+)`),
		regexp.MustCompile(`(?i)(?:note to self|reminder|don't forget)\s*[:-]?\s*([^.!?\n]+)`),
	}

	for _, pattern := range voiceTaskPatterns {
		matches := pattern.FindAllStringSubmatch(transcript, -1)
		for _, match := range matches {
			if len(match) > 1 {
				task := strings.TrimSpace(match[1])
				if len(task) < 5 || len(task) > 300 {
					continue
				}

				memories = append(memories, ExtractedMemory{
					ID:          uuid.New().String(),
					UserID:      userID,
					Type:        MemoryTypeTask,
					Content:     match[0],
					Summary:     task,
					Source:      MemorySourceVoiceNote,
					Confidence:  0.65,
					Tags:        append(s.extractTags(task), "task", "voice"),
					Entities:    s.extractEntitiesFromText(task),
					Importance:  s.inferTaskImportance(task),
					Metadata:    map[string]interface{}{"status": "pending"},
					ExtractedAt: time.Now(),
				})
			}
		}
	}

	return s.deduplicateMemories(memories)
}

// extractReminders extracts reminders from voice notes
func (s *MemoryExtractorService) extractReminders(userID, transcript string) []ExtractedMemory {
	memories := make([]ExtractedMemory, 0)

	reminderPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(?:remind me to|reminder|don't let me forget)\s+([^.!?\n]+)`),
		regexp.MustCompile(`(?i)(?:at|on|by|before)\s+(\d{1,2}(?::\d{2})?\s*(?:am|pm)?|\w+day|\d{1,2}/\d{1,2})\s*(?:,|:)?\s*([^.!?\n]+)`),
		regexp.MustCompile(`(?i)(?:tomorrow|next week|later today)\s*(?:,|:)?\s*([^.!?\n]+)`),
	}

	for _, pattern := range reminderPatterns {
		matches := pattern.FindAllStringSubmatch(transcript, -1)
		for _, match := range matches {
			if len(match) > 1 {
				reminder := strings.TrimSpace(match[len(match)-1])
				if len(reminder) < 5 || len(reminder) > 200 {
					continue
				}

				memories = append(memories, ExtractedMemory{
					ID:          uuid.New().String(),
					UserID:      userID,
					Type:        MemoryTypeReminder,
					Content:     match[0],
					Summary:     reminder,
					Source:      MemorySourceVoiceNote,
					Confidence:  0.7,
					Tags:        []string{"reminder", "voice"},
					Entities:    s.extractEntitiesFromText(reminder),
					Importance:  6,
					Metadata:    make(map[string]interface{}),
					ExtractedAt: time.Now(),
				})
			}
		}
	}

	return memories
}

// extractInsights extracts insights from content
func (s *MemoryExtractorService) extractInsights(userID, content string, messages []Message) []ExtractedMemory {
	memories := make([]ExtractedMemory, 0)

	insightPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(?:i realized|i learned|i discovered|i noticed|i found out)\s+([^.!?\n]+)`),
		regexp.MustCompile(`(?i)(?:the key (?:insight|takeaway|learning) is)\s+([^.!?\n]+)`),
		regexp.MustCompile(`(?i)(?:interesting(?:ly)?|importantly)\s*,?\s+([^.!?\n]+)`),
		regexp.MustCompile(`(?i)(?:turns out|it seems|apparently)\s+([^.!?\n]+)`),
	}

	for _, pattern := range insightPatterns {
		matches := pattern.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) > 1 {
				insight := strings.TrimSpace(match[1])
				if len(insight) < 10 || len(insight) > 500 {
					continue
				}

				memories = append(memories, ExtractedMemory{
					ID:          uuid.New().String(),
					UserID:      userID,
					Type:        MemoryTypeInsight,
					Content:     match[0],
					Summary:     insight,
					Source:      MemorySourceConversation,
					Confidence:  0.65,
					Tags:        append(s.extractTags(insight), "insight"),
					Entities:    s.extractEntitiesFromText(insight),
					Importance:  6,
					Metadata:    make(map[string]interface{}),
					ExtractedAt: time.Now(),
				})
			}
		}
	}

	return s.deduplicateMemories(memories)
}

// extractContacts extracts contact information
func (s *MemoryExtractorService) extractContacts(userID, content string) []ExtractedMemory {
	memories := make([]ExtractedMemory, 0)

	// Email pattern
	emailPattern := regexp.MustCompile(`[\w.-]+@[\w.-]+\.\w+`)
	emails := emailPattern.FindAllString(content, -1)
	for _, email := range emails {
		memories = append(memories, ExtractedMemory{
			ID:          uuid.New().String(),
			UserID:      userID,
			Type:        MemoryTypeContact,
			Content:     email,
			Summary:     fmt.Sprintf("Email: %s", email),
			Source:      MemorySourceConversation,
			Confidence:  0.9,
			Tags:        []string{"contact", "email"},
			Importance:  5,
			Metadata:    map[string]interface{}{"contact_type": "email"},
			ExtractedAt: time.Now(),
		})
	}

	// Phone pattern
	phonePattern := regexp.MustCompile(`\+?[\d\s()-]{10,}`)
	phones := phonePattern.FindAllString(content, -1)
	for _, phone := range phones {
		phone = strings.TrimSpace(phone)
		if len(phone) < 10 {
			continue
		}
		memories = append(memories, ExtractedMemory{
			ID:          uuid.New().String(),
			UserID:      userID,
			Type:        MemoryTypeContact,
			Content:     phone,
			Summary:     fmt.Sprintf("Phone: %s", phone),
			Source:      MemorySourceConversation,
			Confidence:  0.7,
			Tags:        []string{"contact", "phone"},
			Importance:  5,
			Metadata:    map[string]interface{}{"contact_type": "phone"},
			ExtractedAt: time.Now(),
		})
	}

	return memories
}

// extractCodeMemories extracts code-related memories
func (s *MemoryExtractorService) extractCodeMemories(userID, content string, messages []Message) []ExtractedMemory {
	memories := make([]ExtractedMemory, 0)

	// Error patterns
	errorPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)error:?\s*([^\n]+)`),
		regexp.MustCompile(`(?i)(?:got|received|seeing)\s+(?:error|exception)\s*:?\s*([^\n]+)`),
		regexp.MustCompile(`(?i)panic:?\s*([^\n]+)`),
	}

	for _, pattern := range errorPatterns {
		matches := pattern.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) > 1 {
				errorMsg := strings.TrimSpace(match[1])
				if len(errorMsg) < 10 || len(errorMsg) > 500 {
					continue
				}

				memories = append(memories, ExtractedMemory{
					ID:          uuid.New().String(),
					UserID:      userID,
					Type:        MemoryTypeError,
					Content:     match[0],
					Summary:     errorMsg,
					Source:      MemorySourceConversation,
					Confidence:  0.8,
					Tags:        []string{"error", "code"},
					Entities:    s.extractEntitiesFromText(errorMsg),
					Importance:  7,
					Metadata:    make(map[string]interface{}),
					ExtractedAt: time.Now(),
				})
			}
		}
	}

	// Solution patterns
	solutionPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(?:the (?:fix|solution) (?:is|was))\s+([^.!?\n]+)`),
		regexp.MustCompile(`(?i)(?:fixed (?:by|it by|this by))\s+([^.!?\n]+)`),
		regexp.MustCompile(`(?i)(?:solved (?:by|it by))\s+([^.!?\n]+)`),
	}

	for _, pattern := range solutionPatterns {
		matches := pattern.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) > 1 {
				solution := strings.TrimSpace(match[1])
				if len(solution) < 10 || len(solution) > 500 {
					continue
				}

				memories = append(memories, ExtractedMemory{
					ID:          uuid.New().String(),
					UserID:      userID,
					Type:        MemoryTypeSolution,
					Content:     match[0],
					Summary:     solution,
					Source:      MemorySourceConversation,
					Confidence:  0.75,
					Tags:        []string{"solution", "code"},
					Entities:    s.extractEntitiesFromText(solution),
					Importance:  8,
					Metadata:    make(map[string]interface{}),
					ExtractedAt: time.Now(),
				})
			}
		}
	}

	return s.deduplicateMemories(memories)
}

// extractTags extracts relevant tags from text
func (s *MemoryExtractorService) extractTags(text string) []string {
	tags := make([]string, 0)
	lower := strings.ToLower(text)

	// Technology tags
	techTags := map[string][]string{
		"go":         {"go", "golang"},
		"javascript": {"javascript", "js", "node"},
		"typescript": {"typescript", "ts"},
		"react":      {"react", "nextjs"},
		"svelte":     {"svelte", "sveltekit"},
		"python":     {"python", "py"},
		"database":   {"database", "sql", "postgresql", "postgres", "mysql", "redis"},
		"api":        {"api", "rest", "graphql", "endpoint"},
		"docker":     {"docker", "container", "kubernetes", "k8s"},
		"aws":        {"aws", "s3", "ec2", "lambda"},
		"git":        {"git", "github", "gitlab"},
	}

	for tag, keywords := range techTags {
		for _, keyword := range keywords {
			if strings.Contains(lower, keyword) {
				tags = append(tags, tag)
				break
			}
		}
	}

	return tags
}

// extractEntitiesFromText extracts entities from text
func (s *MemoryExtractorService) extractEntitiesFromText(text string) []string {
	entities := make([]string, 0)

	// File paths
	filePattern := regexp.MustCompile(`[\w/-]+\.(go|ts|js|svelte|py|sql|json|yaml|yml|md)`)
	files := filePattern.FindAllString(text, -1)
	entities = append(entities, files...)

	// URLs
	urlPattern := regexp.MustCompile(`https?://[^\s]+`)
	urls := urlPattern.FindAllString(text, -1)
	entities = append(entities, urls...)

	return entities
}

// inferTaskImportance infers task importance from description
func (s *MemoryExtractorService) inferTaskImportance(description string) int {
	lower := strings.ToLower(description)

	if strings.Contains(lower, "urgent") || strings.Contains(lower, "critical") ||
		strings.Contains(lower, "asap") || strings.Contains(lower, "blocker") {
		return 9
	}

	if strings.Contains(lower, "important") || strings.Contains(lower, "must") ||
		strings.Contains(lower, "required") {
		return 7
	}

	if strings.Contains(lower, "nice to have") || strings.Contains(lower, "maybe") ||
		strings.Contains(lower, "could") {
		return 3
	}

	return 5
}

// deduplicateMemories removes duplicate memories
func (s *MemoryExtractorService) deduplicateMemories(memories []ExtractedMemory) []ExtractedMemory {
	seen := make(map[string]bool)
	unique := make([]ExtractedMemory, 0)

	for _, m := range memories {
		key := strings.ToLower(m.Summary)
		if !seen[key] {
			seen[key] = true
			unique = append(unique, m)
		}
	}

	return unique
}

// saveMemory saves an extracted memory to the database
func (s *MemoryExtractorService) saveMemory(ctx context.Context, memory *ExtractedMemory) error {
	tagsJSON, _ := json.Marshal(memory.Tags)
	entitiesJSON, _ := json.Marshal(memory.Entities)
	relatedJSON, _ := json.Marshal(memory.RelatedTo)
	metadataJSON, _ := json.Marshal(memory.Metadata)

	// Add entities and related to metadata
	if memory.Metadata == nil {
		memory.Metadata = make(map[string]interface{})
	}
	memory.Metadata["entities"] = memory.Entities
	memory.Metadata["related_to"] = memory.RelatedTo
	metadataJSON, _ = json.Marshal(memory.Metadata)

	_, err := s.pool.Exec(ctx,
		`INSERT INTO memories (id, user_id, memory_type, content, summary, tags, importance_score, source_type, source_id, metadata, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		 ON CONFLICT (id) DO UPDATE SET
		    content = EXCLUDED.content,
		    summary = EXCLUDED.summary,
		    tags = EXCLUDED.tags,
		    metadata = EXCLUDED.metadata`,
		memory.ID, memory.UserID, string(memory.Type), memory.Content, memory.Summary,
		tagsJSON, memory.Importance, string(memory.Source), memory.SourceID, metadataJSON, memory.ExtractedAt)

	// Store additional details in a separate table if needed
	_ = entitiesJSON
	_ = relatedJSON

	return err
}

// ExtractWithLLM performs LLM-enhanced memory extraction
// This captures nuanced information that pattern-based extraction might miss
func (s *MemoryExtractorService) ExtractWithLLM(ctx context.Context, userID string, messages []Message, opts *ExtractionOptions) (*ExtractionResult, error) {
	startTime := time.Now()

	if s.llmService == nil {
		// Fall back to pattern-based extraction if no LLM service
		return s.ExtractFromConversation(ctx, userID, messages, opts)
	}

	if opts == nil {
		opts = DefaultExtractionOptions()
	}

	result := &ExtractionResult{
		Memories:        make([]ExtractedMemory, 0),
		ByType:          make(map[string]int),
		SourceProcessed: "conversation_llm",
	}

	// Build conversation text for LLM
	var conversationText strings.Builder
	for _, msg := range messages {
		conversationText.WriteString(fmt.Sprintf("[%s]: %s\n\n", msg.Role, msg.Content))
	}

	// Create extraction prompt
	systemPrompt := `You are a memory extraction assistant. Analyze the conversation and extract important information that should be remembered for future reference.

Extract the following types of memories:
1. **Facts**: Concrete information, names, dates, technical details, configurations
2. **Preferences**: User likes, dislikes, working style, communication preferences
3. **Decisions**: Choices made, selected approaches, agreed-upon solutions
4. **Tasks**: Action items, TODOs, things to implement or fix
5. **Insights**: Learnings, discoveries, important realizations
6. **Errors/Solutions**: Problems encountered and how they were solved

For each extracted memory, provide:
- type: one of [fact, preference, decision, task, insight, error, solution]
- summary: brief one-line summary (max 100 chars)
- content: full context (max 300 chars)
- importance: 1-10 scale (10 = critical)
- confidence: 0.0-1.0 how confident you are this should be remembered

IMPORTANT: Focus on information that would be valuable in future conversations.
Skip generic statements and focus on specific, actionable, or unique information.

Respond ONLY with a JSON array of extracted memories. Example:
[
  {"type": "fact", "summary": "Project uses PostgreSQL with pgvector", "content": "The database is PostgreSQL with pgvector extension for embeddings", "importance": 7, "confidence": 0.9},
  {"type": "decision", "summary": "Chose Svelte over React", "content": "Decided to use SvelteKit for the frontend due to better performance", "importance": 8, "confidence": 0.85}
]

If no meaningful memories can be extracted, return an empty array: []`

	chatMessages := []ChatMessage{
		{Role: "user", Content: conversationText.String()},
	}

	// Call LLM for extraction
	response, err := s.llmService.ChatComplete(ctx, chatMessages, systemPrompt)
	if err != nil {
		s.logger.Warn("LLM extraction failed, falling back to pattern-based", "error", err)
		return s.ExtractFromConversation(ctx, userID, messages, opts)
	}

	// Parse LLM response
	llmMemories := s.parseLLMResponse(response, userID)
	result.Memories = append(result.Memories, llmMemories...)

	// Also run pattern-based extraction to catch things LLM might miss
	patternResult, _ := s.ExtractFromConversation(ctx, userID, messages, opts)
	if patternResult != nil {
		// Merge pattern-based results, avoiding duplicates
		for _, pm := range patternResult.Memories {
			if !s.isDuplicate(pm, result.Memories) {
				result.Memories = append(result.Memories, pm)
			}
		}
	}

	// Filter by confidence
	filtered := make([]ExtractedMemory, 0)
	for _, m := range result.Memories {
		if m.Confidence >= opts.MinConfidence {
			filtered = append(filtered, m)
		}
	}
	result.Memories = filtered

	// Limit results
	if opts.MaxMemories > 0 && len(result.Memories) > opts.MaxMemories {
		result.Memories = result.Memories[:opts.MaxMemories]
	}

	// Calculate stats
	result.TotalExtracted = len(result.Memories)
	for _, m := range result.Memories {
		result.ByType[string(m.Type)]++
	}
	result.ProcessingTime = time.Since(startTime).String()

	// Save extracted memories
	for _, memory := range result.Memories {
		if err := s.saveMemory(ctx, &memory); err != nil {
			s.logger.Warn("failed to save extracted memory", "error", err)
		}
	}

	s.logger.Info("LLM-enhanced extraction completed",
		"total", result.TotalExtracted,
		"duration", result.ProcessingTime)

	return result, nil
}

// llmExtractedMemory represents the JSON structure from LLM response
type llmExtractedMemory struct {
	Type       string  `json:"type"`
	Summary    string  `json:"summary"`
	Content    string  `json:"content"`
	Importance int     `json:"importance"`
	Confidence float64 `json:"confidence"`
}

// parseLLMResponse parses the LLM JSON response into ExtractedMemory structs
func (s *MemoryExtractorService) parseLLMResponse(response, userID string) []ExtractedMemory {
	memories := make([]ExtractedMemory, 0)

	// Clean response - remove markdown code blocks if present
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	// Try to find JSON array in response
	startIdx := strings.Index(response, "[")
	endIdx := strings.LastIndex(response, "]")
	if startIdx == -1 || endIdx == -1 || endIdx <= startIdx {
		s.logger.Warn("could not find JSON array in LLM response")
		return memories
	}
	response = response[startIdx : endIdx+1]

	var llmMemories []llmExtractedMemory
	if err := json.Unmarshal([]byte(response), &llmMemories); err != nil {
		s.logger.Warn("failed to parse LLM response", "error", err)
		return memories
	}

	// Convert to ExtractedMemory format
	for _, lm := range llmMemories {
		memType := s.mapLLMType(lm.Type)
		if memType == "" {
			continue
		}

		// Validate
		if len(lm.Summary) < 5 || lm.Confidence < 0.3 {
			continue
		}

		memories = append(memories, ExtractedMemory{
			ID:          uuid.New().String(),
			UserID:      userID,
			Type:        memType,
			Content:     lm.Content,
			Summary:     lm.Summary,
			Source:      MemorySourceConversation,
			Confidence:  lm.Confidence,
			Tags:        s.extractTags(lm.Summary + " " + lm.Content),
			Entities:    s.extractEntitiesFromText(lm.Content),
			Importance:  lm.Importance,
			Metadata:    map[string]interface{}{"extraction_method": "llm"},
			ExtractedAt: time.Now(),
		})
	}

	return memories
}

// mapLLMType maps LLM type strings to MemoryType
func (s *MemoryExtractorService) mapLLMType(llmType string) MemoryType {
	switch strings.ToLower(llmType) {
	case "fact":
		return MemoryTypeFact
	case "preference":
		return MemoryTypePreference
	case "decision":
		return MemoryTypeDecision
	case "task":
		return MemoryTypeTask
	case "insight":
		return MemoryTypeInsight
	case "error":
		return MemoryTypeError
	case "solution":
		return MemoryTypeSolution
	case "reminder":
		return MemoryTypeReminder
	case "contact":
		return MemoryTypeContact
	case "code":
		return MemoryTypeCode
	default:
		return MemoryTypeFact // Default to fact
	}
}

// isDuplicate checks if a memory is a duplicate of existing memories
func (s *MemoryExtractorService) isDuplicate(m ExtractedMemory, existing []ExtractedMemory) bool {
	mLower := strings.ToLower(m.Summary)
	for _, e := range existing {
		eLower := strings.ToLower(e.Summary)
		// Check for exact match or high similarity
		if mLower == eLower {
			return true
		}
		// Check for substring match (one contains the other)
		if len(mLower) > 10 && len(eLower) > 10 {
			if strings.Contains(mLower, eLower) || strings.Contains(eLower, mLower) {
				return true
			}
		}
	}
	return false
}

// GetExtractedMemories retrieves extracted memories for a user
func (s *MemoryExtractorService) GetExtractedMemories(ctx context.Context, userID string, memoryType string, limit int) ([]ExtractedMemory, error) {
	query := `SELECT id, user_id, memory_type, content, summary, tags, importance_score, source_type, source_id, metadata, created_at
	          FROM memories WHERE user_id = $1`
	args := []interface{}{userID}

	if memoryType != "" {
		query += " AND memory_type = $2"
		args = append(args, memoryType)
	}

	query += " ORDER BY created_at DESC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	memories := make([]ExtractedMemory, 0)
	for rows.Next() {
		var m ExtractedMemory
		var tagsJSON, metadataJSON []byte
		var source, sourceID, summary *string

		err := rows.Scan(&m.ID, &m.UserID, &m.Type, &m.Content, &summary, &tagsJSON,
			&m.Importance, &source, &sourceID, &metadataJSON, &m.ExtractedAt)
		if err != nil {
			continue
		}

		json.Unmarshal(tagsJSON, &m.Tags)
		json.Unmarshal(metadataJSON, &m.Metadata)
		if source != nil {
			m.Source = MemorySource(*source)
		}
		if sourceID != nil {
			m.SourceID = *sourceID
		}
		if summary != nil {
			m.Summary = *summary
		}

		memories = append(memories, m)
	}

	return memories, nil
}
