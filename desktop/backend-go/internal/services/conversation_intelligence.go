package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ConversationIntelligenceService provides intelligent analysis of conversations
type ConversationIntelligenceService struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

// ConversationAnalysis represents a detailed analysis of a conversation
type ConversationAnalysis struct {
	ID              string                 `json:"id"`
	ConversationID  string                 `json:"conversation_id"`
	UserID          string                 `json:"user_id"`
	Title           string                 `json:"title"`
	Summary         string                 `json:"summary"`
	KeyPoints       []string               `json:"key_points"`
	Topics          []ConversationTopic    `json:"topics"`
	Sentiment       SentimentAnalysis      `json:"sentiment"`
	Entities        []ConversationEntity   `json:"entities"`
	ActionItems     []ActionItem           `json:"action_items"`
	Questions       []Question             `json:"questions"`
	Decisions       []ConversationDecision `json:"decisions"`
	CodeMentions    []CodeMention          `json:"code_mentions"`
	MessageCount    int                    `json:"message_count"`
	TokenCount      int                    `json:"token_count"`
	Duration        string                 `json:"duration"`
	Metadata        map[string]interface{} `json:"metadata"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// ConversationTopic represents a topic discussed in the conversation
type ConversationTopic struct {
	Name       string   `json:"name"`
	Confidence float64  `json:"confidence"`
	Keywords   []string `json:"keywords"`
	FirstMention int    `json:"first_mention"` // Message index
	Frequency  int      `json:"frequency"`
}

// SentimentAnalysis represents sentiment analysis results
type SentimentAnalysis struct {
	Overall     string             `json:"overall"` // positive, negative, neutral, mixed
	Score       float64            `json:"score"`   // -1 to 1
	Progression []SentimentPoint   `json:"progression"`
	Highlights  []SentimentHighlight `json:"highlights"`
}

// SentimentPoint represents sentiment at a point in the conversation
type SentimentPoint struct {
	MessageIndex int     `json:"message_index"`
	Sentiment    string  `json:"sentiment"`
	Score        float64 `json:"score"`
}

// SentimentHighlight represents a significant sentiment moment
type SentimentHighlight struct {
	MessageIndex int    `json:"message_index"`
	Text         string `json:"text"`
	Sentiment    string `json:"sentiment"`
	Reason       string `json:"reason"`
}

// ConversationEntity represents an entity mentioned in the conversation
type ConversationEntity struct {
	Name       string   `json:"name"`
	Type       string   `json:"type"` // person, organization, technology, file, concept
	Mentions   int      `json:"mentions"`
	Context    []string `json:"context"`
	Related    []string `json:"related"`
}

// ActionItem represents a task or action mentioned
type ActionItem struct {
	ID           string   `json:"id"`
	Description  string   `json:"description"`
	Priority     string   `json:"priority"` // high, medium, low
	Status       string   `json:"status"`   // pending, completed, cancelled
	AssignedTo   string   `json:"assigned_to,omitempty"`
	DueDate      string   `json:"due_date,omitempty"`
	MessageIndex int      `json:"message_index"`
	Tags         []string `json:"tags"`
}

// Question represents a question in the conversation
type Question struct {
	Text         string `json:"text"`
	AskedBy      string `json:"asked_by"` // user, assistant
	MessageIndex int    `json:"message_index"`
	Answered     bool   `json:"answered"`
	Answer       string `json:"answer,omitempty"`
}

// ConversationDecision represents a decision made in the conversation
type ConversationDecision struct {
	Description  string   `json:"description"`
	Context      string   `json:"context"`
	Alternatives []string `json:"alternatives,omitempty"`
	Rationale    string   `json:"rationale,omitempty"`
	MessageIndex int      `json:"message_index"`
}

// CodeMention represents code discussed in the conversation
type CodeMention struct {
	FilePath     string `json:"file_path,omitempty"`
	Language     string `json:"language,omitempty"`
	Snippet      string `json:"snippet"`
	Context      string `json:"context"`
	MessageIndex int    `json:"message_index"`
}

// Message represents a conversation message for analysis
type Message struct {
	Role      string    `json:"role"` // user, assistant, system
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// wordCount is used for keyword frequency analysis
type wordCount struct {
	word  string
	count int
}

// NewConversationIntelligenceService creates a new conversation intelligence service
func NewConversationIntelligenceService(pool *pgxpool.Pool, embeddingService *EmbeddingService) *ConversationIntelligenceService {
	return &ConversationIntelligenceService{
		pool:   pool,
		logger: slog.Default().With("service", "conversation_intelligence"),
	}
}

// AnalyzeConversation performs full analysis on a conversation
func (s *ConversationIntelligenceService) AnalyzeConversation(ctx context.Context, conversationID, userID string, messages []Message) (*ConversationAnalysis, error) {
	analysis := &ConversationAnalysis{
		ID:             uuid.New().String(),
		ConversationID: conversationID,
		UserID:         userID,
		KeyPoints:      make([]string, 0),
		Topics:         make([]ConversationTopic, 0),
		Entities:       make([]ConversationEntity, 0),
		ActionItems:    make([]ActionItem, 0),
		Questions:      make([]Question, 0),
		Decisions:      make([]ConversationDecision, 0),
		CodeMentions:   make([]CodeMention, 0),
		Metadata:       make(map[string]interface{}),
		MessageCount:   len(messages),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if len(messages) == 0 {
		return analysis, nil
	}

	// Calculate duration
	if len(messages) > 1 {
		duration := messages[len(messages)-1].Timestamp.Sub(messages[0].Timestamp)
		analysis.Duration = duration.Round(time.Second).String()
	}

	// Extract topics
	analysis.Topics = s.extractTopics(messages)

	// Extract entities
	analysis.Entities = s.extractEntities(messages)

	// Extract questions
	analysis.Questions = s.extractQuestions(messages)

	// Extract action items
	analysis.ActionItems = s.extractActionItems(messages)

	// Extract decisions
	analysis.Decisions = s.extractDecisions(messages)

	// Extract code mentions
	analysis.CodeMentions = s.extractCodeMentions(messages)

	// Analyze sentiment
	analysis.Sentiment = s.analyzeSentiment(messages)

	// Generate title
	analysis.Title = s.generateTitle(messages, analysis.Topics)

	// Generate summary
	analysis.Summary = s.generateSummary(messages, analysis)

	// Extract key points
	analysis.KeyPoints = s.extractKeyPoints(messages, analysis)

	// Calculate token count (approximation)
	for _, msg := range messages {
		analysis.TokenCount += len(msg.Content) / 4
	}

	// Save to database
	if err := s.saveAnalysis(ctx, analysis); err != nil {
		s.logger.Warn("failed to save conversation analysis", "error", err)
	}

	return analysis, nil
}

// extractTopics extracts topics from messages
func (s *ConversationIntelligenceService) extractTopics(messages []Message) []ConversationTopic {
	// Keyword frequency analysis
	wordFreq := make(map[string]int)
	wordFirstMention := make(map[string]int)

	// Common stop words to filter
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "is": true, "are": true,
		"was": true, "were": true, "be": true, "been": true, "being": true,
		"have": true, "has": true, "had": true, "do": true, "does": true,
		"did": true, "will": true, "would": true, "could": true, "should": true,
		"may": true, "might": true, "must": true, "shall": true, "can": true,
		"this": true, "that": true, "these": true, "those": true, "i": true,
		"you": true, "he": true, "she": true, "it": true, "we": true, "they": true,
		"what": true, "which": true, "who": true, "whom": true, "whose": true,
		"where": true, "when": true, "why": true, "how": true, "all": true,
		"each": true, "every": true, "both": true, "few": true, "more": true,
		"most": true, "other": true, "some": true, "such": true, "no": true,
		"nor": true, "not": true, "only": true, "own": true, "same": true,
		"so": true, "than": true, "too": true, "very": true, "just": true,
		"and": true, "but": true, "if": true, "or": true, "because": true,
		"as": true, "until": true, "while": true, "of": true, "at": true,
		"by": true, "for": true, "with": true, "about": true, "against": true,
		"between": true, "into": true, "through": true, "during": true,
		"before": true, "after": true, "above": true, "below": true, "to": true,
		"from": true, "up": true, "down": true, "in": true, "out": true,
		"on": true, "off": true, "over": true, "under": true, "again": true,
		"further": true, "then": true, "once": true, "here": true, "there": true,
	}

	for idx, msg := range messages {
		words := regexp.MustCompile(`\b[a-zA-Z]{3,}\b`).FindAllString(strings.ToLower(msg.Content), -1)
		for _, word := range words {
			if stopWords[word] {
				continue
			}
			wordFreq[word]++
			if _, exists := wordFirstMention[word]; !exists {
				wordFirstMention[word] = idx
			}
		}
	}

	// Sort by frequency
	var counts []wordCount
	for word, count := range wordFreq {
		if count >= 2 { // Minimum frequency threshold
			counts = append(counts, wordCount{word, count})
		}
	}
	sort.Slice(counts, func(i, j int) bool {
		return counts[i].count > counts[j].count
	})

	// Create topics from top keywords
	topics := make([]ConversationTopic, 0)
	maxTopics := 5
	for i, wc := range counts {
		if i >= maxTopics {
			break
		}

		// Find related keywords
		relatedKeywords := s.findRelatedKeywords(wc.word, counts)

		topics = append(topics, ConversationTopic{
			Name:         wc.word,
			Confidence:   float64(wc.count) / float64(len(messages)),
			Keywords:     relatedKeywords,
			FirstMention: wordFirstMention[wc.word],
			Frequency:    wc.count,
		})
	}

	return topics
}

// findRelatedKeywords finds keywords that often appear with the given word
func (s *ConversationIntelligenceService) findRelatedKeywords(word string, counts []wordCount) []string {
	related := make([]string, 0)
	for _, wc := range counts {
		if wc.word != word && len(related) < 3 {
			related = append(related, wc.word)
		}
	}
	return related
}

// extractEntities extracts named entities from messages
func (s *ConversationIntelligenceService) extractEntities(messages []Message) []ConversationEntity {
	entities := make(map[string]*ConversationEntity)

	// Patterns for different entity types
	patterns := map[string]*regexp.Regexp{
		"file":       regexp.MustCompile(`(?i)[\w-]+\.(go|ts|js|svelte|py|sql|json|yaml|yml|md|txt|css|html|tsx|jsx)`),
		"path":       regexp.MustCompile(`(?:^|[^a-zA-Z0-9])(/[\w/.-]+|[\w]+/[\w/.-]+)`),
		"function":   regexp.MustCompile(`\b[a-z][a-zA-Z0-9]*\([^)]*\)`),
		"technology": regexp.MustCompile(`(?i)\b(react|svelte|vue|angular|node|go|golang|python|typescript|javascript|postgresql|redis|docker|kubernetes|aws|gcp|azure)\b`),
		"url":        regexp.MustCompile(`https?://[^\s]+`),
	}

	for idx, msg := range messages {
		for entityType, pattern := range patterns {
			matches := pattern.FindAllString(msg.Content, -1)
			for _, match := range matches {
				key := strings.ToLower(match)
				if existing, ok := entities[key]; ok {
					existing.Mentions++
					if len(existing.Context) < 3 {
						contextSnippet := s.extractContext(msg.Content, match, 50)
						existing.Context = append(existing.Context, contextSnippet)
					}
				} else {
					entities[key] = &ConversationEntity{
						Name:     match,
						Type:     entityType,
						Mentions: 1,
						Context:  []string{s.extractContext(msg.Content, match, 50)},
						Related:  make([]string, 0),
					}
					_ = idx // Can be used for position tracking
				}
			}
		}
	}

	// Convert to slice and sort by mentions
	result := make([]ConversationEntity, 0, len(entities))
	for _, e := range entities {
		result = append(result, *e)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Mentions > result[j].Mentions
	})

	// Limit to top entities
	if len(result) > 20 {
		result = result[:20]
	}

	return result
}

// extractContext extracts surrounding context for an entity mention
func (s *ConversationIntelligenceService) extractContext(text, entity string, chars int) string {
	idx := strings.Index(strings.ToLower(text), strings.ToLower(entity))
	if idx == -1 {
		return ""
	}

	start := idx - chars
	if start < 0 {
		start = 0
	}
	end := idx + len(entity) + chars
	if end > len(text) {
		end = len(text)
	}

	return strings.TrimSpace(text[start:end])
}

// extractQuestions extracts questions from the conversation
func (s *ConversationIntelligenceService) extractQuestions(messages []Message) []Question {
	questions := make([]Question, 0)

	questionPattern := regexp.MustCompile(`[^.!?]*\?`)

	for idx, msg := range messages {
		matches := questionPattern.FindAllString(msg.Content, -1)
		for _, match := range matches {
			match = strings.TrimSpace(match)
			if len(match) < 10 { // Skip very short questions
				continue
			}

			q := Question{
				Text:         match,
				AskedBy:      msg.Role,
				MessageIndex: idx,
				Answered:     false,
			}

			// Check if question was answered in subsequent messages
			if idx < len(messages)-1 {
				for i := idx + 1; i < len(messages) && i <= idx+3; i++ {
					if messages[i].Role != msg.Role && len(messages[i].Content) > 20 {
						q.Answered = true
						q.Answer = s.truncateText(messages[i].Content, 200)
						break
					}
				}
			}

			questions = append(questions, q)
		}
	}

	return questions
}

// extractActionItems extracts action items from messages
func (s *ConversationIntelligenceService) extractActionItems(messages []Message) []ActionItem {
	actionItems := make([]ActionItem, 0)

	// Patterns that indicate action items
	actionPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(?:need to|should|must|have to|going to|will|todo|task:|action:)\s+([^.!?\n]+)`),
		regexp.MustCompile(`(?i)(?:please|could you|can you)\s+([^.!?\n]+)`),
		regexp.MustCompile(`(?i)- \[ \]\s+([^\n]+)`), // Markdown task
		regexp.MustCompile(`(?i)(?:implement|create|add|fix|update|remove|delete|refactor)\s+([^.!?\n]+)`),
	}

	for idx, msg := range messages {
		for _, pattern := range actionPatterns {
			matches := pattern.FindAllStringSubmatch(msg.Content, -1)
			for _, match := range matches {
				if len(match) > 1 {
					description := strings.TrimSpace(match[1])
					if len(description) < 5 || len(description) > 200 {
						continue
					}

					actionItems = append(actionItems, ActionItem{
						ID:           uuid.New().String(),
						Description:  description,
						Priority:     s.inferPriority(description),
						Status:       "pending",
						MessageIndex: idx,
						Tags:         s.extractTags(description),
					})
				}
			}
		}
	}

	// Deduplicate
	seen := make(map[string]bool)
	unique := make([]ActionItem, 0)
	for _, item := range actionItems {
		key := strings.ToLower(item.Description)
		if !seen[key] {
			seen[key] = true
			unique = append(unique, item)
		}
	}

	return unique
}

// inferPriority infers priority from action item description
func (s *ConversationIntelligenceService) inferPriority(description string) string {
	lower := strings.ToLower(description)

	highPriorityWords := []string{"urgent", "critical", "asap", "immediately", "important", "must", "blocker"}
	for _, word := range highPriorityWords {
		if strings.Contains(lower, word) {
			return "high"
		}
	}

	lowPriorityWords := []string{"maybe", "consider", "could", "nice to have", "eventually", "later"}
	for _, word := range lowPriorityWords {
		if strings.Contains(lower, word) {
			return "low"
		}
	}

	return "medium"
}

// extractTags extracts tags from description
func (s *ConversationIntelligenceService) extractTags(description string) []string {
	tags := make([]string, 0)

	// Common categories
	categories := map[string][]string{
		"bug":         {"fix", "bug", "error", "issue", "broken"},
		"feature":     {"add", "implement", "create", "new", "feature"},
		"refactor":    {"refactor", "clean", "improve", "optimize"},
		"docs":        {"document", "readme", "comment", "docs"},
		"test":        {"test", "spec", "coverage"},
		"security":    {"security", "auth", "permission", "vulnerability"},
		"performance": {"performance", "speed", "optimize", "cache"},
	}

	lower := strings.ToLower(description)
	for tag, keywords := range categories {
		for _, keyword := range keywords {
			if strings.Contains(lower, keyword) {
				tags = append(tags, tag)
				break
			}
		}
	}

	return tags
}

// extractDecisions extracts decisions from the conversation
func (s *ConversationIntelligenceService) extractDecisions(messages []Message) []ConversationDecision {
	decisions := make([]ConversationDecision, 0)

	decisionPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(?:decided to|will use|going with|chose|selected|opted for)\s+([^.!?\n]+)`),
		regexp.MustCompile(`(?i)(?:let's|we should|we'll)\s+(?:go with|use|implement)\s+([^.!?\n]+)`),
		regexp.MustCompile(`(?i)(?:the (?:best|better) (?:approach|solution|option) is)\s+([^.!?\n]+)`),
	}

	for idx, msg := range messages {
		for _, pattern := range decisionPatterns {
			matches := pattern.FindAllStringSubmatch(msg.Content, -1)
			for _, match := range matches {
				if len(match) > 1 {
					decisions = append(decisions, ConversationDecision{
						Description:  strings.TrimSpace(match[1]),
						Context:      s.truncateText(msg.Content, 200),
						MessageIndex: idx,
					})
				}
			}
		}
	}

	return decisions
}

// extractCodeMentions extracts code snippets and references
func (s *ConversationIntelligenceService) extractCodeMentions(messages []Message) []CodeMention {
	mentions := make([]CodeMention, 0)

	codeBlockPattern := regexp.MustCompile("```(\\w*)\\n([\\s\\S]*?)```")
	filePathPattern := regexp.MustCompile(`(?:^|[^a-zA-Z0-9])([\w/-]+\.(go|ts|js|svelte|py|sql|json|yaml|yml|md|css|html|tsx|jsx))`)

	for idx, msg := range messages {
		// Extract code blocks
		codeMatches := codeBlockPattern.FindAllStringSubmatch(msg.Content, -1)
		for _, match := range codeMatches {
			if len(match) > 2 {
				mentions = append(mentions, CodeMention{
					Language:     match[1],
					Snippet:      s.truncateText(match[2], 500),
					Context:      "code block",
					MessageIndex: idx,
				})
			}
		}

		// Extract file references
		fileMatches := filePathPattern.FindAllStringSubmatch(msg.Content, -1)
		for _, match := range fileMatches {
			if len(match) > 1 {
				mentions = append(mentions, CodeMention{
					FilePath:     match[1],
					Context:      s.extractContext(msg.Content, match[1], 100),
					MessageIndex: idx,
				})
			}
		}
	}

	return mentions
}

// analyzeSentiment performs sentiment analysis on messages
func (s *ConversationIntelligenceService) analyzeSentiment(messages []Message) SentimentAnalysis {
	result := SentimentAnalysis{
		Progression: make([]SentimentPoint, 0),
		Highlights:  make([]SentimentHighlight, 0),
	}

	// Sentiment word lists
	positiveWords := map[string]float64{
		"great": 0.8, "good": 0.6, "excellent": 0.9, "perfect": 1.0,
		"thanks": 0.5, "thank": 0.5, "helpful": 0.7, "awesome": 0.9,
		"love": 0.8, "nice": 0.5, "wonderful": 0.8, "amazing": 0.9,
		"works": 0.6, "working": 0.5, "solved": 0.7, "fixed": 0.7,
		"yes": 0.3, "correct": 0.5, "right": 0.4, "exactly": 0.6,
	}

	negativeWords := map[string]float64{
		"bad": -0.6, "wrong": -0.5, "error": -0.4, "fail": -0.6,
		"failed": -0.6, "broken": -0.7, "issue": -0.3, "problem": -0.4,
		"bug": -0.3, "not working": -0.6, "doesn't work": -0.7,
		"confused": -0.4, "frustrating": -0.7, "annoying": -0.6,
		"hate": -0.8, "terrible": -0.9, "awful": -0.8, "worst": -0.9,
		"unfortunately": -0.3, "sadly": -0.4, "sorry": -0.2,
	}

	totalScore := 0.0
	messageScores := make([]float64, len(messages))

	for idx, msg := range messages {
		if msg.Role == "system" {
			continue
		}

		words := strings.Fields(strings.ToLower(msg.Content))
		score := 0.0
		wordCount := 0

		for _, word := range words {
			word = strings.Trim(word, ".,!?;:\"'")
			if val, ok := positiveWords[word]; ok {
				score += val
				wordCount++
			}
			if val, ok := negativeWords[word]; ok {
				score += val
				wordCount++
			}
		}

		if wordCount > 0 {
			score = score / float64(wordCount)
		}

		messageScores[idx] = score
		totalScore += score

		sentiment := "neutral"
		if score > 0.2 {
			sentiment = "positive"
		} else if score < -0.2 {
			sentiment = "negative"
		}

		result.Progression = append(result.Progression, SentimentPoint{
			MessageIndex: idx,
			Sentiment:    sentiment,
			Score:        score,
		})

		// Add highlights for strong sentiments
		if score > 0.5 || score < -0.5 {
			result.Highlights = append(result.Highlights, SentimentHighlight{
				MessageIndex: idx,
				Text:         s.truncateText(msg.Content, 100),
				Sentiment:    sentiment,
				Reason:       fmt.Sprintf("Strong %s sentiment detected", sentiment),
			})
		}
	}

	// Calculate overall sentiment
	avgScore := 0.0
	if len(messages) > 0 {
		avgScore = totalScore / float64(len(messages))
	}

	result.Score = avgScore
	if avgScore > 0.2 {
		result.Overall = "positive"
	} else if avgScore < -0.2 {
		result.Overall = "negative"
	} else if len(result.Highlights) > 2 {
		result.Overall = "mixed"
	} else {
		result.Overall = "neutral"
	}

	return result
}

// generateTitle generates a title for the conversation
func (s *ConversationIntelligenceService) generateTitle(messages []Message, topics []ConversationTopic) string {
	// Use first user message as basis
	for _, msg := range messages {
		if msg.Role == "user" && len(msg.Content) > 10 {
			// Extract first sentence or line
			content := msg.Content
			if idx := strings.IndexAny(content, ".!?\n"); idx > 0 && idx < 100 {
				content = content[:idx]
			}
			if len(content) > 60 {
				content = content[:60] + "..."
			}
			return strings.TrimSpace(content)
		}
	}

	// Fall back to topics
	if len(topics) > 0 {
		return fmt.Sprintf("Discussion about %s", topics[0].Name)
	}

	return "Untitled Conversation"
}

// generateSummary generates a summary of the conversation
func (s *ConversationIntelligenceService) generateSummary(messages []Message, analysis *ConversationAnalysis) string {
	var sb strings.Builder

	// Opening
	sb.WriteString(fmt.Sprintf("A conversation with %d messages", len(messages)))
	if len(analysis.Topics) > 0 {
		topicNames := make([]string, 0)
		for _, t := range analysis.Topics {
			topicNames = append(topicNames, t.Name)
		}
		sb.WriteString(fmt.Sprintf(" discussing %s", strings.Join(topicNames, ", ")))
	}
	sb.WriteString(". ")

	// Key activities
	if len(analysis.CodeMentions) > 0 {
		sb.WriteString(fmt.Sprintf("Code was discussed in %d instances. ", len(analysis.CodeMentions)))
	}
	if len(analysis.ActionItems) > 0 {
		sb.WriteString(fmt.Sprintf("%d action items were identified. ", len(analysis.ActionItems)))
	}
	if len(analysis.Decisions) > 0 {
		sb.WriteString(fmt.Sprintf("%d decisions were made. ", len(analysis.Decisions)))
	}
	if len(analysis.Questions) > 0 {
		answered := 0
		for _, q := range analysis.Questions {
			if q.Answered {
				answered++
			}
		}
		sb.WriteString(fmt.Sprintf("%d of %d questions were answered. ", answered, len(analysis.Questions)))
	}

	// Sentiment
	sb.WriteString(fmt.Sprintf("Overall sentiment was %s.", analysis.Sentiment.Overall))

	return sb.String()
}

// extractKeyPoints extracts key points from the conversation
func (s *ConversationIntelligenceService) extractKeyPoints(messages []Message, analysis *ConversationAnalysis) []string {
	points := make([]string, 0)

	// Add decisions as key points
	for _, d := range analysis.Decisions {
		if len(d.Description) > 10 {
			points = append(points, "Decision: "+d.Description)
		}
	}

	// Add high-priority action items
	for _, a := range analysis.ActionItems {
		if a.Priority == "high" {
			points = append(points, "Action: "+a.Description)
		}
	}

	// Limit to 5 key points
	if len(points) > 5 {
		points = points[:5]
	}

	return points
}

// truncateText truncates text to maxLen characters
func (s *ConversationIntelligenceService) truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}

// saveAnalysis saves the conversation analysis to the database
func (s *ConversationIntelligenceService) saveAnalysis(ctx context.Context, analysis *ConversationAnalysis) error {
	keyPointsJSON, _ := json.Marshal(analysis.KeyPoints)
	topicsJSON, _ := json.Marshal(analysis.Topics)
	sentimentJSON, _ := json.Marshal(analysis.Sentiment)
	entitiesJSON, _ := json.Marshal(analysis.Entities)
	actionItemsJSON, _ := json.Marshal(analysis.ActionItems)
	questionsJSON, _ := json.Marshal(analysis.Questions)
	decisionsJSON, _ := json.Marshal(analysis.Decisions)
	codeMentionsJSON, _ := json.Marshal(analysis.CodeMentions)
	metadataJSON, _ := json.Marshal(analysis.Metadata)

	_, err := s.pool.Exec(ctx,
		`INSERT INTO conversation_summaries
		 (id, conversation_id, user_id, title, summary, key_points, topics, sentiment,
		  entities, action_items, questions, decisions, code_mentions, message_count,
		  token_count, duration, metadata, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
		 ON CONFLICT (conversation_id) DO UPDATE SET
		    title = EXCLUDED.title,
		    summary = EXCLUDED.summary,
		    key_points = EXCLUDED.key_points,
		    topics = EXCLUDED.topics,
		    sentiment = EXCLUDED.sentiment,
		    entities = EXCLUDED.entities,
		    action_items = EXCLUDED.action_items,
		    questions = EXCLUDED.questions,
		    decisions = EXCLUDED.decisions,
		    code_mentions = EXCLUDED.code_mentions,
		    message_count = EXCLUDED.message_count,
		    token_count = EXCLUDED.token_count,
		    duration = EXCLUDED.duration,
		    metadata = EXCLUDED.metadata,
		    updated_at = EXCLUDED.updated_at`,
		analysis.ID, analysis.ConversationID, analysis.UserID, analysis.Title, analysis.Summary,
		keyPointsJSON, topicsJSON, sentimentJSON, entitiesJSON, actionItemsJSON,
		questionsJSON, decisionsJSON, codeMentionsJSON, analysis.MessageCount,
		analysis.TokenCount, analysis.Duration, metadataJSON, analysis.CreatedAt, analysis.UpdatedAt)

	return err
}

// GetAnalysis retrieves a conversation analysis
func (s *ConversationIntelligenceService) GetAnalysis(ctx context.Context, conversationID string) (*ConversationAnalysis, error) {
	var analysis ConversationAnalysis
	var keyPointsJSON, topicsJSON, sentimentJSON, entitiesJSON []byte
	var actionItemsJSON, questionsJSON, decisionsJSON, codeMentionsJSON, metadataJSON []byte

	err := s.pool.QueryRow(ctx,
		`SELECT id, conversation_id, user_id, title, summary, key_points, topics, sentiment,
		        entities, action_items, questions, decisions, code_mentions, message_count,
		        token_count, duration, metadata, created_at, updated_at
		 FROM conversation_summaries WHERE conversation_id = $1`,
		conversationID).Scan(
		&analysis.ID, &analysis.ConversationID, &analysis.UserID, &analysis.Title, &analysis.Summary,
		&keyPointsJSON, &topicsJSON, &sentimentJSON, &entitiesJSON, &actionItemsJSON,
		&questionsJSON, &decisionsJSON, &codeMentionsJSON, &analysis.MessageCount,
		&analysis.TokenCount, &analysis.Duration, &metadataJSON, &analysis.CreatedAt, &analysis.UpdatedAt)

	if err != nil {
		return nil, err
	}

	json.Unmarshal(keyPointsJSON, &analysis.KeyPoints)
	json.Unmarshal(topicsJSON, &analysis.Topics)
	json.Unmarshal(sentimentJSON, &analysis.Sentiment)
	json.Unmarshal(entitiesJSON, &analysis.Entities)
	json.Unmarshal(actionItemsJSON, &analysis.ActionItems)
	json.Unmarshal(questionsJSON, &analysis.Questions)
	json.Unmarshal(decisionsJSON, &analysis.Decisions)
	json.Unmarshal(codeMentionsJSON, &analysis.CodeMentions)
	json.Unmarshal(metadataJSON, &analysis.Metadata)

	return &analysis, nil
}

// SearchConversations searches conversation analyses
func (s *ConversationIntelligenceService) SearchConversations(ctx context.Context, userID, query string, limit int) ([]ConversationAnalysis, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, conversation_id, user_id, title, summary, key_points, topics, sentiment,
		        entities, action_items, message_count, token_count, duration, created_at
		 FROM conversation_summaries
		 WHERE user_id = $1 AND (
		    title ILIKE $2 OR summary ILIKE $2
		 )
		 ORDER BY created_at DESC
		 LIMIT $3`,
		userID, "%"+query+"%", limit)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	analyses := make([]ConversationAnalysis, 0)
	for rows.Next() {
		var a ConversationAnalysis
		var keyPointsJSON, topicsJSON, sentimentJSON, entitiesJSON, actionItemsJSON []byte

		err := rows.Scan(&a.ID, &a.ConversationID, &a.UserID, &a.Title, &a.Summary,
			&keyPointsJSON, &topicsJSON, &sentimentJSON, &entitiesJSON, &actionItemsJSON,
			&a.MessageCount, &a.TokenCount, &a.Duration, &a.CreatedAt)
		if err != nil {
			continue
		}

		json.Unmarshal(keyPointsJSON, &a.KeyPoints)
		json.Unmarshal(topicsJSON, &a.Topics)
		json.Unmarshal(sentimentJSON, &a.Sentiment)
		json.Unmarshal(entitiesJSON, &a.Entities)
		json.Unmarshal(actionItemsJSON, &a.ActionItems)

		analyses = append(analyses, a)
	}

	return analyses, nil
}
