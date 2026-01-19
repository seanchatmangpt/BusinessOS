package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/rhl/businessos-backend/internal/services"
	"github.com/rhl/businessos-backend/internal/streaming"
)

// ResponseQualityMetrics contains all quality measurements for an agent response
type ResponseQualityMetrics struct {
	// Quality Scores (0-100)
	RelevanceScore    float64 `json:"relevance_score"`
	CompletenessScore float64 `json:"completeness_score"`
	BrevityScore      float64 `json:"brevity_score"`
	CoherenceScore    float64 `json:"coherence_score"`
	OverallScore      float64 `json:"overall_score"`

	// Voice-specific Scores
	VoiceFriendliness      float64 `json:"voice_friendliness"`
	VoiceAppropriateLength float64 `json:"voice_appropriate_length"`

	// Latency Metrics
	TimeToFirstToken  time.Duration `json:"time_to_first_token_ms"`
	TotalResponseTime time.Duration `json:"total_response_time_ms"`
	TokensPerSecond   float64       `json:"tokens_per_second"`

	// Response Characteristics
	ResponseLength    int     `json:"response_length_chars"`
	WordCount         int     `json:"word_count"`
	HasCodeBlocks     bool    `json:"has_code_blocks"`
	HasTables         bool    `json:"has_tables"`
	HasMarkdown       bool    `json:"has_markdown"`
	SentenceCount     int     `json:"sentence_count"`
	AvgSentenceLength float64 `json:"avg_sentence_length"`

	// Test Case Info
	TestCase      string `json:"test_case"`
	Input         string `json:"input"`
	Response      string `json:"response"`
	ExpectedType  string `json:"expected_type"`
	PassedTest    bool   `json:"passed_test"`
	FailureReason string `json:"failure_reason,omitempty"`

	// Metadata
	AgentType  string               `json:"agent_type"`
	Model      string               `json:"model"`
	Timestamp  time.Time            `json:"timestamp"`
	TokenUsage *services.TokenUsage `json:"token_usage,omitempty"`
}

// ValidationReport contains results for multiple test cases
type ValidationReport struct {
	TestSuite       string                   `json:"test_suite"`
	TotalTests      int                      `json:"total_tests"`
	PassedTests     int                      `json:"passed_tests"`
	FailedTests     int                      `json:"failed_tests"`
	AvgOverallScore float64                  `json:"avg_overall_score"`
	AvgLatency      time.Duration            `json:"avg_latency_ms"`
	TestResults     []ResponseQualityMetrics `json:"test_results"`
	Summary         string                   `json:"summary"`
	Timestamp       time.Time                `json:"timestamp"`
}

// ResponseValidator validates agent response quality
type ResponseValidator struct {
	registry             *AgentRegistryV2
	minRelevanceScore    float64
	minCompletenessScore float64
	minCoherenceScore    float64
	maxVoiceWords        int
	idealVoiceWords      int
}

// NewResponseValidator creates a new response validator
func NewResponseValidator(registry *AgentRegistryV2) *ResponseValidator {
	return &ResponseValidator{
		registry:             registry,
		minRelevanceScore:    70.0,
		minCompletenessScore: 75.0,
		minCoherenceScore:    80.0,
		maxVoiceWords:        200, // Max words for voice response
		idealVoiceWords:      100, // Ideal target for voice
	}
}

// TestCase represents a single test case
type TestCase struct {
	Name            string
	Input           string
	ExpectedType    string // "greeting", "question", "command", "context", "error"
	MinRelevance    float64
	MinCompleteness float64
	AgentType       AgentTypeV2
	Context         *services.TieredContext
}

// ValidateResponse tests a single response against quality criteria
func (v *ResponseValidator) ValidateResponse(
	ctx context.Context,
	testCase TestCase,
	userID string,
	userName string,
) (*ResponseQualityMetrics, error) {
	slog.Info("validating response", "test_case", testCase.Name)

	metrics := &ResponseQualityMetrics{
		TestCase:     testCase.Name,
		Input:        testCase.Input,
		ExpectedType: testCase.ExpectedType,
		AgentType:    string(testCase.AgentType),
		Timestamp:    time.Now(),
	}

	// Get agent
	agent := v.registry.GetAgent(
		testCase.AgentType,
		userID,
		userName,
		nil,
		testCase.Context,
	)

	// Create agent input
	input := AgentInput{
		Messages: []services.ChatMessage{
			{Role: "user", Content: testCase.Input},
		},
		Context:  testCase.Context,
		UserID:   userID,
		UserName: userName,
	}

	// Measure latency
	startTime := time.Now()
	var firstTokenTime time.Time
	var response strings.Builder
	tokenCount := 0

	// Execute agent
	eventChan, errChan := agent.Run(ctx, input)

	// Collect response
	for {
		select {
		case event, ok := <-eventChan:
			if !ok {
				eventChan = nil
				continue
			}

			// Record first token time
			if firstTokenTime.IsZero() && event.Type == streaming.EventTypeToken {
				firstTokenTime = time.Now()
				metrics.TimeToFirstToken = firstTokenTime.Sub(startTime)
			}

			// Collect content
			if event.Type == streaming.EventTypeToken && event.Content != "" {
				response.WriteString(event.Content)
				tokenCount++
			}

		case err, ok := <-errChan:
			if !ok {
				errChan = nil
				continue
			}
			if err != nil {
				return nil, fmt.Errorf("agent error: %w", err)
			}
		}

		if eventChan == nil && errChan == nil {
			break
		}
	}

	metrics.TotalResponseTime = time.Since(startTime)
	metrics.Response = response.String()

	// Calculate tokens per second
	if metrics.TotalResponseTime.Seconds() > 0 {
		metrics.TokensPerSecond = float64(tokenCount) / metrics.TotalResponseTime.Seconds()
	}

	// Analyze response characteristics
	v.analyzeResponseCharacteristics(metrics)

	// Calculate quality scores
	v.calculateRelevanceScore(metrics, testCase)
	v.calculateCompletenessScore(metrics, testCase)
	v.calculateBrevityScore(metrics)
	v.calculateCoherenceScore(metrics)
	v.calculateVoiceScores(metrics)

	// Calculate overall score (weighted average)
	metrics.OverallScore = v.calculateOverallScore(metrics)

	// Determine pass/fail
	metrics.PassedTest = v.evaluateTest(metrics, testCase)

	slog.Info("validation complete",
		"test_case", testCase.Name,
		"overall_score", metrics.OverallScore,
		"passed", metrics.PassedTest,
	)

	return metrics, nil
}

// analyzeResponseCharacteristics extracts basic characteristics from response
func (v *ResponseValidator) analyzeResponseCharacteristics(m *ResponseQualityMetrics) {
	m.ResponseLength = utf8.RuneCountInString(m.Response)

	// Word count
	words := strings.Fields(m.Response)
	m.WordCount = len(words)

	// Check for code blocks
	m.HasCodeBlocks = strings.Contains(m.Response, "```") ||
		strings.Contains(m.Response, "    ") // 4-space indent

	// Check for tables
	m.HasTables = strings.Contains(m.Response, "|") &&
		strings.Count(m.Response, "|") > 3

	// Check for markdown
	m.HasMarkdown = strings.Contains(m.Response, "#") ||
		strings.Contains(m.Response, "*") ||
		strings.Contains(m.Response, "_") ||
		strings.Contains(m.Response, "[")

	// Sentence analysis
	sentencePattern := regexp.MustCompile(`[.!?]+`)
	sentences := sentencePattern.Split(m.Response, -1)
	m.SentenceCount = 0
	totalSentenceLength := 0

	for _, sentence := range sentences {
		trimmed := strings.TrimSpace(sentence)
		if len(trimmed) > 0 {
			m.SentenceCount++
			totalSentenceLength += len(strings.Fields(trimmed))
		}
	}

	if m.SentenceCount > 0 {
		m.AvgSentenceLength = float64(totalSentenceLength) / float64(m.SentenceCount)
	}
}

// calculateRelevanceScore measures semantic relevance to input
func (v *ResponseValidator) calculateRelevanceScore(m *ResponseQualityMetrics, tc TestCase) {
	// Simple keyword matching heuristic
	// In production, use embedding-based semantic similarity

	inputLower := strings.ToLower(tc.Input)
	responseLower := strings.ToLower(m.Response)

	// Extract key terms from input
	stopWords := map[string]bool{
		"the": true, "is": true, "at": true, "which": true, "on": true,
		"a": true, "an": true, "and": true, "or": true, "but": true,
		"in": true, "with": true, "to": true, "for": true, "of": true,
		"can": true, "you": true, "i": true, "me": true, "my": true,
		"what": true, "how": true, "when": true, "where": true, "why": true,
	}

	inputWords := strings.Fields(inputLower)
	keyTerms := []string{}
	for _, word := range inputWords {
		cleaned := strings.Trim(word, ".,!?;:")
		if len(cleaned) > 3 && !stopWords[cleaned] {
			keyTerms = append(keyTerms, cleaned)
		}
	}

	// Count how many key terms appear in response
	matches := 0
	for _, term := range keyTerms {
		if strings.Contains(responseLower, term) {
			matches++
		}
	}

	if len(keyTerms) > 0 {
		matchRatio := float64(matches) / float64(len(keyTerms))
		m.RelevanceScore = matchRatio * 100
	} else {
		// If no key terms, check for expected response patterns
		m.RelevanceScore = v.checkExpectedPattern(m, tc)
	}

	// Cap at 100
	if m.RelevanceScore > 100 {
		m.RelevanceScore = 100
	}
}

// checkExpectedPattern checks if response matches expected type
func (v *ResponseValidator) checkExpectedPattern(m *ResponseQualityMetrics, tc TestCase) float64 {
	responseLower := strings.ToLower(m.Response)

	switch tc.ExpectedType {
	case "greeting":
		// Should contain greeting words
		greetings := []string{"hello", "hi", "hey", "greetings", "welcome"}
		for _, g := range greetings {
			if strings.Contains(responseLower, g) {
				return 90.0
			}
		}
		return 50.0

	case "error":
		// Should acknowledge not knowing or apologize
		errorPhrases := []string{"don't know", "not sure", "can't", "unable", "sorry"}
		for _, phrase := range errorPhrases {
			if strings.Contains(responseLower, phrase) {
				return 85.0
			}
		}
		return 40.0

	case "question":
		// Should provide an answer
		if m.WordCount > 10 {
			return 75.0
		}
		return 50.0

	default:
		return 70.0
	}
}

// calculateCompletenessScore measures if response fully addresses input
func (v *ResponseValidator) calculateCompletenessScore(m *ResponseQualityMetrics, tc TestCase) {
	// Base score on response length relative to input complexity

	// Count questions in input
	questionCount := strings.Count(tc.Input, "?")
	if questionCount == 0 {
		// Check for question words
		questionWords := []string{"what", "how", "when", "where", "why", "who", "which"}
		inputLower := strings.ToLower(tc.Input)
		for _, qw := range questionWords {
			if strings.Contains(inputLower, qw) {
				questionCount++
				break
			}
		}
	}

	// Expected minimum word count based on input complexity
	minWords := 20
	if questionCount > 0 {
		minWords = 30 * questionCount
	}

	if tc.ExpectedType == "greeting" {
		minWords = 5
	} else if tc.ExpectedType == "error" {
		minWords = 10
	}

	// Score based on meeting minimum
	if m.WordCount >= minWords {
		m.CompletenessScore = 100.0
	} else {
		ratio := float64(m.WordCount) / float64(minWords)
		m.CompletenessScore = ratio * 100
	}

	// Bonus for structure (paragraphs, sentences)
	if m.SentenceCount > 2 && m.AvgSentenceLength > 5 {
		m.CompletenessScore = math.Min(100, m.CompletenessScore*1.1)
	}
}

// calculateBrevityScore penalizes overly long responses
func (v *ResponseValidator) calculateBrevityScore(m *ResponseQualityMetrics) {
	// Ideal: concise but complete
	// Penalize both too short and too long

	idealMin := 30
	idealMax := 150

	if m.WordCount < idealMin {
		// Too short
		ratio := float64(m.WordCount) / float64(idealMin)
		m.BrevityScore = ratio * 100
	} else if m.WordCount > idealMax {
		// Too long - penalty increases exponentially
		excess := m.WordCount - idealMax
		penalty := math.Min(50, float64(excess)/10)
		m.BrevityScore = 100 - penalty
	} else {
		// In ideal range
		m.BrevityScore = 100.0
	}
}

// calculateCoherenceScore measures logical flow and readability
func (v *ResponseValidator) calculateCoherenceScore(m *ResponseQualityMetrics) {
	score := 100.0

	// Penalize if average sentence is too long (hard to follow)
	if m.AvgSentenceLength > 30 {
		score -= 10
	}

	// Penalize if average sentence is too short (choppy)
	if m.AvgSentenceLength < 8 && m.SentenceCount > 1 {
		score -= 5
	}

	// Reward good sentence count (not too many, not too few)
	if m.SentenceCount >= 2 && m.SentenceCount <= 8 {
		score += 5
	}

	// Check for repeated words (indicates low quality)
	words := strings.Fields(strings.ToLower(m.Response))
	wordFreq := make(map[string]int)
	for _, word := range words {
		wordFreq[word]++
	}

	// Penalize if any word appears more than 10% of total
	for _, count := range wordFreq {
		if len(words) > 0 && float64(count)/float64(len(words)) > 0.1 {
			score -= 5
			break
		}
	}

	// Check for connector words (indicates flow)
	connectors := []string{"however", "therefore", "additionally", "furthermore",
		"moreover", "consequently", "thus", "hence"}
	hasConnectors := false
	responseLower := strings.ToLower(m.Response)
	for _, connector := range connectors {
		if strings.Contains(responseLower, connector) {
			hasConnectors = true
			break
		}
	}
	if hasConnectors && m.SentenceCount > 2 {
		score += 5
	}

	m.CoherenceScore = math.Max(0, math.Min(100, score))
}

// calculateVoiceScores evaluates voice-specific quality
func (v *ResponseValidator) calculateVoiceScores(m *ResponseQualityMetrics) {
	// Voice friendliness: penalize code, tables, markdown
	voiceScore := 100.0

	if m.HasCodeBlocks {
		voiceScore -= 40
	}
	if m.HasTables {
		voiceScore -= 30
	}
	if m.HasMarkdown {
		voiceScore -= 10
	}

	// Reward natural conversational tone
	conversationalWords := []string{"you", "your", "i", "we", "let's", "can", "will"}
	responseLower := strings.ToLower(m.Response)
	conversationalCount := 0
	for _, word := range conversationalWords {
		if strings.Contains(responseLower, " "+word+" ") {
			conversationalCount++
		}
	}
	if conversationalCount >= 2 {
		voiceScore += 10
	}

	m.VoiceFriendliness = math.Max(0, math.Min(100, voiceScore))

	// Voice appropriate length
	if m.WordCount <= v.idealVoiceWords {
		m.VoiceAppropriateLength = 100.0
	} else if m.WordCount <= v.maxVoiceWords {
		// In acceptable range but not ideal
		ratio := float64(v.maxVoiceWords-m.WordCount) / float64(v.maxVoiceWords-v.idealVoiceWords)
		m.VoiceAppropriateLength = 70 + (ratio * 30)
	} else {
		// Too long for voice
		excess := m.WordCount - v.maxVoiceWords
		penalty := math.Min(70, float64(excess)/5)
		m.VoiceAppropriateLength = 70 - penalty
	}
}

// calculateOverallScore computes weighted average of all scores
func (v *ResponseValidator) calculateOverallScore(m *ResponseQualityMetrics) float64 {
	// Weights
	weights := map[string]float64{
		"relevance":    0.30,
		"completeness": 0.25,
		"brevity":      0.15,
		"coherence":    0.20,
		"voice":        0.10,
	}

	overall := (m.RelevanceScore * weights["relevance"]) +
		(m.CompletenessScore * weights["completeness"]) +
		(m.BrevityScore * weights["brevity"]) +
		(m.CoherenceScore * weights["coherence"]) +
		((m.VoiceFriendliness + m.VoiceAppropriateLength) / 2 * weights["voice"])

	return math.Round(overall*100) / 100
}

// evaluateTest determines if test passes based on criteria
func (v *ResponseValidator) evaluateTest(m *ResponseQualityMetrics, tc TestCase) bool {
	minRelevance := tc.MinRelevance
	if minRelevance == 0 {
		minRelevance = v.minRelevanceScore
	}

	minCompleteness := tc.MinCompleteness
	if minCompleteness == 0 {
		minCompleteness = v.minCompletenessScore
	}

	if m.RelevanceScore < minRelevance {
		m.FailureReason = fmt.Sprintf("relevance too low: %.1f < %.1f", m.RelevanceScore, minRelevance)
		return false
	}

	if m.CompletenessScore < minCompleteness {
		m.FailureReason = fmt.Sprintf("completeness too low: %.1f < %.1f", m.CompletenessScore, minCompleteness)
		return false
	}

	if m.CoherenceScore < v.minCoherenceScore {
		m.FailureReason = fmt.Sprintf("coherence too low: %.1f < %.1f", m.CoherenceScore, v.minCoherenceScore)
		return false
	}

	return true
}

// RunTestSuite executes a full suite of test cases
func (v *ResponseValidator) RunTestSuite(
	ctx context.Context,
	suiteName string,
	testCases []TestCase,
	userID string,
	userName string,
) (*ValidationReport, error) {
	slog.Info("running test suite", "suite", suiteName, "tests", len(testCases))

	report := &ValidationReport{
		TestSuite:   suiteName,
		TotalTests:  len(testCases),
		PassedTests: 0,
		FailedTests: 0,
		TestResults: make([]ResponseQualityMetrics, 0, len(testCases)),
		Timestamp:   time.Now(),
	}

	totalLatency := time.Duration(0)
	totalScore := 0.0

	for _, tc := range testCases {
		metrics, err := v.ValidateResponse(ctx, tc, userID, userName)
		if err != nil {
			slog.Error("test case failed", "case", tc.Name, "error", err)
			// Create failure metrics
			metrics = &ResponseQualityMetrics{
				TestCase:      tc.Name,
				Input:         tc.Input,
				ExpectedType:  tc.ExpectedType,
				PassedTest:    false,
				FailureReason: err.Error(),
			}
		}

		report.TestResults = append(report.TestResults, *metrics)

		if metrics.PassedTest {
			report.PassedTests++
		} else {
			report.FailedTests++
		}

		totalLatency += metrics.TotalResponseTime
		totalScore += metrics.OverallScore
	}

	if len(testCases) > 0 {
		report.AvgLatency = totalLatency / time.Duration(len(testCases))
		report.AvgOverallScore = totalScore / float64(len(testCases))
	}

	// Generate summary
	passRate := float64(report.PassedTests) / float64(report.TotalTests) * 100
	report.Summary = fmt.Sprintf(
		"Test Suite: %s | Pass Rate: %.1f%% (%d/%d) | Avg Score: %.1f | Avg Latency: %dms",
		suiteName,
		passRate,
		report.PassedTests,
		report.TotalTests,
		report.AvgOverallScore,
		report.AvgLatency.Milliseconds(),
	)

	slog.Info("test suite complete", "summary", report.Summary)

	return report, nil
}

// GetStandardTestSuite returns a comprehensive set of standard test cases
func GetStandardTestSuite() []TestCase {
	return []TestCase{
		// Greetings
		{
			Name:            "simple_greeting",
			Input:           "Hello",
			ExpectedType:    "greeting",
			MinRelevance:    80.0,
			MinCompleteness: 60.0,
			AgentType:       AgentTypeV2Orchestrator,
		},
		{
			Name:            "greeting_with_question",
			Input:           "Hi there! How are you?",
			ExpectedType:    "greeting",
			MinRelevance:    75.0,
			MinCompleteness: 65.0,
			AgentType:       AgentTypeV2Orchestrator,
		},

		// Question answering
		{
			Name:            "simple_question",
			Input:           "What is the capital of France?",
			ExpectedType:    "question",
			MinRelevance:    85.0,
			MinCompleteness: 80.0,
			AgentType:       AgentTypeV2Orchestrator,
		},
		{
			Name:            "complex_question",
			Input:           "Can you explain the difference between machine learning and deep learning?",
			ExpectedType:    "question",
			MinRelevance:    80.0,
			MinCompleteness: 75.0,
			AgentType:       AgentTypeV2Orchestrator,
		},
		{
			Name:            "multi_part_question",
			Input:           "What is project management? How do I get started? What tools should I use?",
			ExpectedType:    "question",
			MinRelevance:    75.0,
			MinCompleteness: 70.0,
			AgentType:       AgentTypeV2Project,
		},

		// Commands
		{
			Name:            "task_extraction_command",
			Input:           "Extract tasks from: Finish report by Friday, call client next week, review budget",
			ExpectedType:    "command",
			MinRelevance:    80.0,
			MinCompleteness: 75.0,
			AgentType:       AgentTypeV2Task,
		},
		{
			Name:            "document_creation_command",
			Input:           "Write a business proposal for a new mobile app",
			ExpectedType:    "command",
			MinRelevance:    80.0,
			MinCompleteness: 75.0,
			AgentType:       AgentTypeV2Document,
		},

		// Context maintenance
		{
			Name:            "follow_up_question",
			Input:           "Tell me more about that",
			ExpectedType:    "context",
			MinRelevance:    60.0,
			MinCompleteness: 60.0,
			AgentType:       AgentTypeV2Orchestrator,
		},

		// Error handling
		{
			Name:            "out_of_scope",
			Input:           "What's the weather like on Mars?",
			ExpectedType:    "error",
			MinRelevance:    70.0,
			MinCompleteness: 65.0,
			AgentType:       AgentTypeV2Orchestrator,
		},
		{
			Name:            "unclear_request",
			Input:           "asdf jkl qwerty",
			ExpectedType:    "error",
			MinRelevance:    50.0,
			MinCompleteness: 50.0,
			AgentType:       AgentTypeV2Orchestrator,
		},

		// Voice-specific tests
		{
			Name:            "voice_concise_answer",
			Input:           "What time is it?",
			ExpectedType:    "question",
			MinRelevance:    80.0,
			MinCompleteness: 70.0,
			AgentType:       AgentTypeV2Orchestrator,
		},
		{
			Name:            "voice_conversational",
			Input:           "Tell me about your capabilities",
			ExpectedType:    "question",
			MinRelevance:    75.0,
			MinCompleteness: 70.0,
			AgentType:       AgentTypeV2Orchestrator,
		},
	}
}

// ExportReportJSON exports validation report as JSON string
func (r *ValidationReport) ExportReportJSON() (string, error) {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal report: %w", err)
	}
	return string(data), nil
}

// ExportMetricsJSON exports single metrics as JSON string
func (m *ResponseQualityMetrics) ExportMetricsJSON() (string, error) {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal metrics: %w", err)
	}
	return string(data), nil
}
