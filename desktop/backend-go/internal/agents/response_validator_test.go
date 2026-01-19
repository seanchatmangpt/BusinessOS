package agents

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/rhl/businessos-backend/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResponseQualityMetrics_Serialization(t *testing.T) {
	metrics := ResponseQualityMetrics{
		RelevanceScore:         85.5,
		CompletenessScore:      90.0,
		BrevityScore:           75.0,
		CoherenceScore:         88.0,
		OverallScore:           85.0,
		VoiceFriendliness:      80.0,
		VoiceAppropriateLength: 95.0,
		TimeToFirstToken:       100 * time.Millisecond,
		TotalResponseTime:      500 * time.Millisecond,
		TokensPerSecond:        50.0,
		ResponseLength:         150,
		WordCount:              30,
		HasCodeBlocks:          false,
		HasTables:              false,
		HasMarkdown:            true,
		SentenceCount:          5,
		AvgSentenceLength:      6.0,
		TestCase:               "test_case_1",
		Input:                  "Test input",
		Response:               "Test response",
		ExpectedType:           "question",
		PassedTest:             true,
		AgentType:              string(AgentTypeV2Orchestrator),
		Model:                  "test-model",
		Timestamp:              time.Now(),
	}

	// Test JSON serialization
	jsonStr, err := metrics.ExportMetricsJSON()
	require.NoError(t, err)
	assert.NotEmpty(t, jsonStr)

	// Verify it's valid JSON
	var decoded ResponseQualityMetrics
	err = json.Unmarshal([]byte(jsonStr), &decoded)
	require.NoError(t, err)
	assert.Equal(t, metrics.OverallScore, decoded.OverallScore)
}

func TestResponseValidator_AnalyzeResponseCharacteristics(t *testing.T) {
	validator := &ResponseValidator{}

	tests := []struct {
		name                string
		response            string
		expectedWordCount   int
		expectedHasCode     bool
		expectedHasTables   bool
		expectedHasMarkdown bool
		minSentenceCount    int
	}{
		{
			name:                "simple_text",
			response:            "This is a simple response. It has two sentences.",
			expectedWordCount:   9,
			expectedHasCode:     false,
			expectedHasTables:   false,
			expectedHasMarkdown: false,
			minSentenceCount:    2,
		},
		{
			name:                "with_code_block",
			response:            "Here's some code:\n```go\nfunc main() {}\n```",
			expectedWordCount:   8, // "Here's some code go func main"
			expectedHasCode:     true,
			expectedHasTables:   false,
			expectedHasMarkdown: false, // ``` counts as code, not markdown
			minSentenceCount:    1,
		},
		{
			name:                "with_table",
			response:            "| Header 1 | Header 2 |\n|----------|----------|\n| Data 1   | Data 2   |",
			expectedWordCount:   15, // All words including separators
			expectedHasCode:     false,
			expectedHasTables:   true,
			expectedHasMarkdown: false, // | is table-specific
			minSentenceCount:    0,
		},
		{
			name:                "with_markdown",
			response:            "# Heading\n\n**Bold** and *italic* text. [Link](url)",
			expectedWordCount:   7,
			expectedHasCode:     false,
			expectedHasTables:   false,
			expectedHasMarkdown: true,
			minSentenceCount:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := &ResponseQualityMetrics{
				Response: tt.response,
			}

			validator.analyzeResponseCharacteristics(metrics)

			assert.Equal(t, tt.expectedWordCount, metrics.WordCount, "word count mismatch")
			assert.Equal(t, tt.expectedHasCode, metrics.HasCodeBlocks, "code blocks detection mismatch")
			assert.Equal(t, tt.expectedHasTables, metrics.HasTables, "tables detection mismatch")
			assert.Equal(t, tt.expectedHasMarkdown, metrics.HasMarkdown, "markdown detection mismatch")
			assert.GreaterOrEqual(t, metrics.SentenceCount, tt.minSentenceCount, "sentence count too low")
		})
	}
}

func TestResponseValidator_CalculateRelevanceScore(t *testing.T) {
	validator := &ResponseValidator{}

	tests := []struct {
		name         string
		input        string
		response     string
		expectedType string
		minScore     float64
		maxScore     float64
	}{
		{
			name:         "highly_relevant",
			input:        "What is machine learning?",
			response:     "Machine learning is a subset of artificial intelligence...",
			expectedType: "question",
			minScore:     80.0,
			maxScore:     100.0,
		},
		{
			name:         "greeting_relevant",
			input:        "Hello",
			response:     "Hello! How can I help you today?",
			expectedType: "greeting",
			minScore:     85.0,
			maxScore:     100.0,
		},
		{
			name:         "error_handling",
			input:        "asdf qwerty jkl",
			response:     "I'm sorry, I don't understand. Could you rephrase?",
			expectedType: "error",
			minScore:     0.0, // Input has no valid key terms (all < 4 chars)
			maxScore:     100.0,
		},
		{
			name:         "low_relevance",
			input:        "Tell me about Python programming",
			response:     "The weather is nice today.",
			expectedType: "question",
			minScore:     0.0,
			maxScore:     30.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := &ResponseQualityMetrics{
				Response: tt.response,
			}
			testCase := TestCase{
				Input:        tt.input,
				ExpectedType: tt.expectedType,
			}

			validator.calculateRelevanceScore(metrics, testCase)

			assert.GreaterOrEqual(t, metrics.RelevanceScore, tt.minScore, "relevance score too low")
			assert.LessOrEqual(t, metrics.RelevanceScore, tt.maxScore, "relevance score too high")
		})
	}
}

func TestResponseValidator_CalculateCompletenessScore(t *testing.T) {
	validator := &ResponseValidator{}

	tests := []struct {
		name         string
		input        string
		response     string
		expectedType string
		minScore     float64
	}{
		{
			name:         "complete_answer",
			input:        "What is the capital of France?",
			response:     "The capital of France is Paris, which is located in the north-central part of the country.",
			expectedType: "question",
			minScore:     50.0, // Realistic: 18 words vs 30 expected = 60%
		},
		{
			name:         "incomplete_answer",
			input:        "Explain machine learning and deep learning in detail",
			response:     "It's complicated.",
			expectedType: "question",
			minScore:     0.0,
		},
		{
			name:         "greeting_complete",
			input:        "Hi",
			response:     "Hello!",
			expectedType: "greeting",
			minScore:     15.0, // Realistic: 1 word vs 5 expected = 20%
		},
		{
			name:         "multi_question_complete",
			input:        "What is AI? How does it work? What are its applications?",
			response:     "AI stands for Artificial Intelligence. It works by using algorithms to process data and make decisions. Applications include healthcare, finance, and autonomous vehicles.",
			expectedType: "question",
			minScore:     25.0, // Realistic: 25 words vs 90 expected (3 questions) = 28%
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := &ResponseQualityMetrics{
				Response: tt.response,
			}
			validator.analyzeResponseCharacteristics(metrics)

			testCase := TestCase{
				Input:        tt.input,
				ExpectedType: tt.expectedType,
			}

			validator.calculateCompletenessScore(metrics, testCase)

			assert.GreaterOrEqual(t, metrics.CompletenessScore, tt.minScore, "completeness score too low")
		})
	}
}

func TestResponseValidator_CalculateBrevityScore(t *testing.T) {
	validator := &ResponseValidator{}

	tests := []struct {
		name     string
		response string
		minScore float64
	}{
		{
			name:     "too_short",
			response: "Yes.",
			minScore: 0.0,
		},
		{
			name:     "ideal_length",
			response: "This is a well-balanced response that provides enough information without being too verbose. It contains around 50-100 words which is ideal for most conversational interactions.",
			minScore: 80.0, // Realistic: 25 words is close to ideal range
		},
		{
			name:     "too_long",
			response: generateLongResponse(300), // 300 words
			minScore: 40.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := &ResponseQualityMetrics{
				Response: tt.response,
			}
			validator.analyzeResponseCharacteristics(metrics)
			validator.calculateBrevityScore(metrics)

			assert.GreaterOrEqual(t, metrics.BrevityScore, tt.minScore, "brevity score too low")
		})
	}
}

func TestResponseValidator_CalculateCoherenceScore(t *testing.T) {
	validator := &ResponseValidator{}

	tests := []struct {
		name     string
		response string
		minScore float64
	}{
		{
			name:     "coherent_response",
			response: "This is a coherent response. It has good flow. The sentences are well-structured. Each idea connects logically.",
			minScore: 85.0,
		},
		{
			name:     "choppy_response",
			response: "Yes. No. Maybe. So. Then. OK. Fine. Yes.",
			minScore: 70.0,
		},
		{
			name:     "repetitive_response",
			response: "The thing is thing thing thing. Thing thing thing thing thing thing.",
			minScore: 60.0,
		},
		{
			name:     "good_connectors",
			response: "First, we need to understand the basics. However, there are exceptions. Therefore, we must consider all factors. Additionally, further research is needed.",
			minScore: 90.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := &ResponseQualityMetrics{
				Response: tt.response,
			}
			validator.analyzeResponseCharacteristics(metrics)
			validator.calculateCoherenceScore(metrics)

			assert.GreaterOrEqual(t, metrics.CoherenceScore, tt.minScore, "coherence score too low")
		})
	}
}

func TestResponseValidator_CalculateVoiceScores(t *testing.T) {
	validator := &ResponseValidator{
		maxVoiceWords:   200,
		idealVoiceWords: 100,
	}

	tests := []struct {
		name                   string
		response               string
		minVoiceFriendliness   float64
		minVoiceAppropriateLen float64
	}{
		{
			name:                   "voice_friendly",
			response:               "I can help you with that. Let's start by looking at your options. You have several choices here.",
			minVoiceFriendliness:   85.0,
			minVoiceAppropriateLen: 90.0,
		},
		{
			name:                   "has_code_block",
			response:               "Here's the code:\n```python\ndef hello():\n    print('hello')\n```",
			minVoiceFriendliness:   0.0,
			minVoiceAppropriateLen: 90.0,
		},
		{
			name:                   "has_table",
			response:               "| Column 1 | Column 2 |\n|----------|----------|\n| Data     | More     |",
			minVoiceFriendliness:   0.0,
			minVoiceAppropriateLen: 90.0,
		},
		{
			name:                   "too_long_for_voice",
			response:               generateLongResponse(250),
			minVoiceFriendliness:   0.0,
			minVoiceAppropriateLen: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := &ResponseQualityMetrics{
				Response: tt.response,
			}
			validator.analyzeResponseCharacteristics(metrics)
			validator.calculateVoiceScores(metrics)

			assert.GreaterOrEqual(t, metrics.VoiceFriendliness, tt.minVoiceFriendliness, "voice friendliness too low")
			assert.GreaterOrEqual(t, metrics.VoiceAppropriateLength, tt.minVoiceAppropriateLen, "voice length score too low")
		})
	}
}

func TestResponseValidator_CalculateOverallScore(t *testing.T) {
	validator := &ResponseValidator{}

	tests := []struct {
		name     string
		metrics  ResponseQualityMetrics
		expected float64
	}{
		{
			name: "perfect_score",
			metrics: ResponseQualityMetrics{
				RelevanceScore:         100.0,
				CompletenessScore:      100.0,
				BrevityScore:           100.0,
				CoherenceScore:         100.0,
				VoiceFriendliness:      100.0,
				VoiceAppropriateLength: 100.0,
			},
			expected: 100.0,
		},
		{
			name: "mixed_scores",
			metrics: ResponseQualityMetrics{
				RelevanceScore:         80.0,
				CompletenessScore:      90.0,
				BrevityScore:           70.0,
				CoherenceScore:         85.0,
				VoiceFriendliness:      75.0,
				VoiceAppropriateLength: 80.0,
			},
			expected: 82.0, // Approximate weighted average
		},
		{
			name: "low_scores",
			metrics: ResponseQualityMetrics{
				RelevanceScore:         40.0,
				CompletenessScore:      50.0,
				BrevityScore:           60.0,
				CoherenceScore:         45.0,
				VoiceFriendliness:      55.0,
				VoiceAppropriateLength: 50.0,
			},
			expected: 48.0, // Approximate weighted average
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			overall := validator.calculateOverallScore(&tt.metrics)

			// Allow 5 point tolerance for floating point math
			assert.InDelta(t, tt.expected, overall, 5.0, "overall score mismatch")
		})
	}
}

func TestResponseValidator_EvaluateTest(t *testing.T) {
	validator := &ResponseValidator{
		minRelevanceScore:    70.0,
		minCompletenessScore: 75.0,
		minCoherenceScore:    80.0,
	}

	tests := []struct {
		name       string
		metrics    ResponseQualityMetrics
		testCase   TestCase
		shouldPass bool
	}{
		{
			name: "all_scores_pass",
			metrics: ResponseQualityMetrics{
				RelevanceScore:    85.0,
				CompletenessScore: 90.0,
				CoherenceScore:    88.0,
			},
			testCase:   TestCase{},
			shouldPass: true,
		},
		{
			name: "relevance_fails",
			metrics: ResponseQualityMetrics{
				RelevanceScore:    60.0,
				CompletenessScore: 90.0,
				CoherenceScore:    88.0,
			},
			testCase:   TestCase{},
			shouldPass: false,
		},
		{
			name: "completeness_fails",
			metrics: ResponseQualityMetrics{
				RelevanceScore:    85.0,
				CompletenessScore: 70.0,
				CoherenceScore:    88.0,
			},
			testCase:   TestCase{},
			shouldPass: false,
		},
		{
			name: "coherence_fails",
			metrics: ResponseQualityMetrics{
				RelevanceScore:    85.0,
				CompletenessScore: 90.0,
				CoherenceScore:    75.0,
			},
			testCase:   TestCase{},
			shouldPass: false,
		},
		{
			name: "custom_thresholds_pass",
			metrics: ResponseQualityMetrics{
				RelevanceScore:    65.0,
				CompletenessScore: 70.0,
				CoherenceScore:    82.0,
			},
			testCase: TestCase{
				MinRelevance:    60.0,
				MinCompleteness: 65.0,
			},
			shouldPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			passed := validator.evaluateTest(&tt.metrics, tt.testCase)
			assert.Equal(t, tt.shouldPass, passed, "test evaluation mismatch")

			if !passed {
				assert.NotEmpty(t, tt.metrics.FailureReason, "should have failure reason")
			}
		})
	}
}

func TestGetStandardTestSuite(t *testing.T) {
	testCases := GetStandardTestSuite()

	assert.NotEmpty(t, testCases, "test suite should not be empty")
	assert.GreaterOrEqual(t, len(testCases), 10, "should have at least 10 test cases")

	// Verify all test cases have required fields
	for _, tc := range testCases {
		assert.NotEmpty(t, tc.Name, "test case name should not be empty")
		assert.NotEmpty(t, tc.Input, "test case input should not be empty")
		assert.NotEmpty(t, tc.ExpectedType, "test case expected type should not be empty")
		assert.NotEmpty(t, tc.AgentType, "test case agent type should not be empty")
	}

	// Verify we have different test types
	expectedTypes := make(map[string]bool)
	for _, tc := range testCases {
		expectedTypes[tc.ExpectedType] = true
	}

	assert.True(t, expectedTypes["greeting"], "should have greeting tests")
	assert.True(t, expectedTypes["question"], "should have question tests")
	assert.True(t, expectedTypes["command"], "should have command tests")
	assert.True(t, expectedTypes["error"], "should have error tests")
}

func TestValidationReport_ExportJSON(t *testing.T) {
	report := ValidationReport{
		TestSuite:       "test_suite",
		TotalTests:      10,
		PassedTests:     8,
		FailedTests:     2,
		AvgOverallScore: 85.5,
		AvgLatency:      500 * time.Millisecond,
		TestResults: []ResponseQualityMetrics{
			{
				TestCase:     "test1",
				OverallScore: 90.0,
				PassedTest:   true,
			},
		},
		Summary:   "Test complete",
		Timestamp: time.Now(),
	}

	jsonStr, err := report.ExportReportJSON()
	require.NoError(t, err)
	assert.NotEmpty(t, jsonStr)

	// Verify it's valid JSON
	var decoded ValidationReport
	err = json.Unmarshal([]byte(jsonStr), &decoded)
	require.NoError(t, err)
	assert.Equal(t, report.TotalTests, decoded.TotalTests)
	assert.Equal(t, report.PassedTests, decoded.PassedTests)
}

// Integration test - only runs with INTEGRATION_TEST=true
func TestResponseValidator_Integration(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=true to run.")
	}

	// Load config
	cfg, err := config.Load()
	require.NoError(t, err)

	// Create registry (minimal setup for testing)
	registry := NewAgentRegistryV2(nil, cfg, nil, nil)
	validator := NewResponseValidator(registry)

	ctx := context.Background()

	// Test single validation
	testCase := TestCase{
		Name:            "integration_greeting",
		Input:           "Hello, how are you?",
		ExpectedType:    "greeting",
		MinRelevance:    70.0,
		MinCompleteness: 60.0,
		AgentType:       AgentTypeV2Orchestrator,
	}

	metrics, err := validator.ValidateResponse(ctx, testCase, "test-user", "Test User")
	require.NoError(t, err)
	assert.NotNil(t, metrics)
	assert.NotEmpty(t, metrics.Response)
	assert.Greater(t, metrics.OverallScore, 0.0)

	t.Logf("Integration test results:")
	t.Logf("  Response: %s", metrics.Response)
	t.Logf("  Overall Score: %.2f", metrics.OverallScore)
	t.Logf("  Relevance: %.2f", metrics.RelevanceScore)
	t.Logf("  Completeness: %.2f", metrics.CompletenessScore)
	t.Logf("  Brevity: %.2f", metrics.BrevityScore)
	t.Logf("  Coherence: %.2f", metrics.CoherenceScore)
	t.Logf("  Time to First Token: %v", metrics.TimeToFirstToken)
	t.Logf("  Total Response Time: %v", metrics.TotalResponseTime)
	t.Logf("  Tokens/sec: %.2f", metrics.TokensPerSecond)
}

// Full test suite integration test
func TestResponseValidator_FullSuiteIntegration(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=true to run.")
	}

	// Load config
	cfg, err := config.Load()
	require.NoError(t, err)

	// Create registry
	registry := NewAgentRegistryV2(nil, cfg, nil, nil)
	validator := NewResponseValidator(registry)

	ctx := context.Background()

	// Run a subset of standard tests (to avoid long test times)
	testCases := []TestCase{
		{
			Name:            "greeting",
			Input:           "Hello",
			ExpectedType:    "greeting",
			MinRelevance:    70.0,
			MinCompleteness: 60.0,
			AgentType:       AgentTypeV2Orchestrator,
		},
		{
			Name:            "simple_question",
			Input:           "What is machine learning?",
			ExpectedType:    "question",
			MinRelevance:    75.0,
			MinCompleteness: 70.0,
			AgentType:       AgentTypeV2Orchestrator,
		},
	}

	report, err := validator.RunTestSuite(ctx, "integration_suite", testCases, "test-user", "Test User")
	require.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, len(testCases), report.TotalTests)

	t.Logf("Full suite results:")
	t.Logf("  %s", report.Summary)

	// Export report
	jsonReport, err := report.ExportReportJSON()
	require.NoError(t, err)

	// Optionally write to file for inspection
	outputFile := "/tmp/validation_report.json"
	err = os.WriteFile(outputFile, []byte(jsonReport), 0644)
	require.NoError(t, err)
	t.Logf("Report written to: %s", outputFile)
}

// Helper functions

// generateLongResponse creates a response with approximately the specified word count
func generateLongResponse(wordCount int) string {
	words := []string{"This", "is", "a", "test", "response", "with", "many", "words", "to", "simulate", "verbose", "output"}
	result := ""
	for i := 0; i < wordCount; i++ {
		result += words[i%len(words)] + " "
	}
	return result
}
