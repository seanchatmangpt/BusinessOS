package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/agents"
	"github.com/rhl/businessos-backend/internal/config"
	"github.com/rhl/businessos-backend/internal/database"
	"github.com/rhl/businessos-backend/internal/services"
)

// CLI tool to validate agent response quality
// Usage:
//   go run ./cmd/validate_agents              # Run standard test suite
//   go run ./cmd/validate_agents -custom      # Run custom test cases
//   go run ./cmd/validate_agents -output report.json  # Write results to file

func main() {
	// Parse flags
	customTests := flag.Bool("custom", false, "Run custom test cases")
	outputFile := flag.String("output", "", "Output file for JSON report (default: stdout)")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	singleTest := flag.String("test", "", "Run a single test case by name")
	agentType := flag.String("agent", "orchestrator", "Agent type to test (orchestrator, document, project, task, client, analyst)")
	flag.Parse()

	// Setup logging
	logLevel := slog.LevelInfo
	if *verbose {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger)

	slog.Info("starting agent validation")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Connect to database (if needed for full context)
	var pool *pgxpool.Pool
	if cfg.DatabaseURL != "" {
		pool, err = database.Connect(cfg)
		if err != nil {
			slog.Warn("database connection failed, running without DB context", "error", err)
		} else {
			defer pool.Close()
			slog.Info("database connected")
		}
	}

	// Create services (simplified - not needed for basic validation)
	var embeddingService *services.EmbeddingService
	var promptPersonalizer *services.PromptPersonalizer
	// Note: Full services require learning/memory services
	// For basic validation, we can use nil

	// Create agent registry
	registry := agents.NewAgentRegistryV2(pool, cfg, embeddingService, promptPersonalizer)

	// Create validator
	validator := agents.NewResponseValidator(registry)

	ctx := context.Background()

	// Determine which tests to run
	var testCases []agents.TestCase
	var suiteName string

	if *singleTest != "" {
		// Run single test
		testCases = filterTestCaseByName(agents.GetStandardTestSuite(), *singleTest)
		if len(testCases) == 0 {
			log.Fatalf("test case not found: %s", *singleTest)
		}
		suiteName = fmt.Sprintf("single_test_%s", *singleTest)
	} else if *customTests {
		// Run custom tests
		testCases = getCustomTestSuite(*agentType)
		suiteName = "custom_test_suite"
	} else {
		// Run standard tests
		testCases = agents.GetStandardTestSuite()
		suiteName = "standard_test_suite"
	}

	slog.Info("running test suite", "suite", suiteName, "tests", len(testCases))

	// Run test suite
	startTime := time.Now()
	report, err := validator.RunTestSuite(
		ctx,
		suiteName,
		testCases,
		"validator-user",
		"Validator Bot",
	)
	if err != nil {
		log.Fatalf("test suite failed: %v", err)
	}
	duration := time.Since(startTime)

	slog.Info("test suite completed", "duration", duration)

	// Print summary to console
	fmt.Println("\n" + strings.Repeat("═", 80))
	fmt.Println("AGENT RESPONSE VALIDATION REPORT")
	fmt.Println(strings.Repeat("═", 80))
	fmt.Printf("Test Suite: %s\n", report.TestSuite)
	fmt.Printf("Total Tests: %d\n", report.TotalTests)
	fmt.Printf("Passed: %d\n", report.PassedTests)
	fmt.Printf("Failed: %d\n", report.FailedTests)
	fmt.Printf("Pass Rate: %.1f%%\n", float64(report.PassedTests)/float64(report.TotalTests)*100)
	fmt.Printf("Average Overall Score: %.2f\n", report.AvgOverallScore)
	fmt.Printf("Average Latency: %dms\n", report.AvgLatency.Milliseconds())
	fmt.Printf("Total Duration: %s\n", duration)
	fmt.Println(strings.Repeat("═", 80))

	// Print individual test results
	fmt.Println("\nDETAILED RESULTS:")
	fmt.Println(strings.Repeat("-", 80))

	for i, result := range report.TestResults {
		status := "✓ PASS"
		if !result.PassedTest {
			status = "✗ FAIL"
		}

		fmt.Printf("\n%d. %s %s\n", i+1, status, result.TestCase)
		fmt.Printf("   Input: %s\n", truncate(result.Input, 60))
		fmt.Printf("   Response: %s\n", truncate(result.Response, 60))
		fmt.Printf("   Scores: Overall=%.1f, Relevance=%.1f, Completeness=%.1f, Brevity=%.1f, Coherence=%.1f\n",
			result.OverallScore, result.RelevanceScore, result.CompletenessScore,
			result.BrevityScore, result.CoherenceScore)
		fmt.Printf("   Voice: Friendliness=%.1f, Length=%.1f\n",
			result.VoiceFriendliness, result.VoiceAppropriateLength)
		fmt.Printf("   Latency: First Token=%dms, Total=%dms, Tokens/sec=%.1f\n",
			result.TimeToFirstToken.Milliseconds(),
			result.TotalResponseTime.Milliseconds(),
			result.TokensPerSecond)

		if !result.PassedTest {
			fmt.Printf("   ⚠ Failure: %s\n", result.FailureReason)
		}
	}

	// Export JSON report
	jsonReport, err := report.ExportReportJSON()
	if err != nil {
		log.Fatalf("failed to export JSON: %v", err)
	}

	if *outputFile != "" {
		// Write to file
		err = os.WriteFile(*outputFile, []byte(jsonReport), 0644)
		if err != nil {
			log.Fatalf("failed to write output file: %v", err)
		}
		fmt.Printf("\n✓ Full report written to: %s\n", *outputFile)
	} else {
		// Print to stdout
		fmt.Println("\n" + strings.Repeat("═", 80))
		fmt.Println("JSON REPORT:")
		fmt.Println(strings.Repeat("═", 80))

		// Pretty print JSON
		var prettyJSON map[string]interface{}
		json.Unmarshal([]byte(jsonReport), &prettyJSON)
		prettyBytes, _ := json.MarshalIndent(prettyJSON, "", "  ")
		fmt.Println(string(prettyBytes))
	}

	// Exit with appropriate code
	if report.FailedTests > 0 {
		os.Exit(1)
	}
}

// getCustomTestSuite returns custom test cases for specific scenarios
func getCustomTestSuite(agentTypeStr string) []agents.TestCase {
	agentType := agents.AgentTypeV2FromString(agentTypeStr)

	return []agents.TestCase{
		// Voice-optimized tests
		{
			Name:            "voice_brief_greeting",
			Input:           "Hey",
			ExpectedType:    "greeting",
			MinRelevance:    80.0,
			MinCompleteness: 60.0,
			AgentType:       agentType,
		},
		{
			Name:            "voice_quick_question",
			Input:           "What's the status?",
			ExpectedType:    "question",
			MinRelevance:    70.0,
			MinCompleteness: 65.0,
			AgentType:       agentType,
		},
		{
			Name:            "voice_command",
			Input:           "Create a task",
			ExpectedType:    "command",
			MinRelevance:    75.0,
			MinCompleteness: 70.0,
			AgentType:       agentType,
		},
		{
			Name:            "voice_clarification",
			Input:           "Can you explain that?",
			ExpectedType:    "question",
			MinRelevance:    65.0,
			MinCompleteness: 60.0,
			AgentType:       agentType,
		},
		{
			Name:            "voice_error_graceful",
			Input:           "zxcvbnm",
			ExpectedType:    "error",
			MinRelevance:    50.0,
			MinCompleteness: 50.0,
			AgentType:       agentType,
		},

		// Business context tests
		{
			Name:            "business_project_query",
			Input:           "What's the timeline for the mobile app project?",
			ExpectedType:    "question",
			MinRelevance:    75.0,
			MinCompleteness: 70.0,
			AgentType:       agents.AgentTypeV2Project,
		},
		{
			Name:            "business_task_creation",
			Input:           "Add a task to follow up with the client about the proposal",
			ExpectedType:    "command",
			MinRelevance:    80.0,
			MinCompleteness: 75.0,
			AgentType:       agents.AgentTypeV2Task,
		},
		{
			Name:            "business_document_request",
			Input:           "Draft an email to the stakeholders",
			ExpectedType:    "command",
			MinRelevance:    80.0,
			MinCompleteness: 75.0,
			AgentType:       agents.AgentTypeV2Document,
		},

		// Edge cases
		{
			Name:            "edge_very_long_input",
			Input:           generateLongInput(500),
			ExpectedType:    "question",
			MinRelevance:    60.0,
			MinCompleteness: 60.0,
			AgentType:       agentType,
		},
		{
			Name:            "edge_special_characters",
			Input:           "What about @mentions, #hashtags, and $symbols?",
			ExpectedType:    "question",
			MinRelevance:    70.0,
			MinCompleteness: 65.0,
			AgentType:       agentType,
		},
	}
}

// filterTestCaseByName filters test cases by name
func filterTestCaseByName(testCases []agents.TestCase, name string) []agents.TestCase {
	var filtered []agents.TestCase
	for _, tc := range testCases {
		if tc.Name == name {
			filtered = append(filtered, tc)
		}
	}
	return filtered
}

// truncate truncates a string to maxLen characters
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// generateLongInput generates a long input string with specified word count
func generateLongInput(wordCount int) string {
	words := []string{
		"Can", "you", "help", "me", "understand", "the", "complex", "relationship",
		"between", "various", "business", "processes", "and", "how", "they", "impact",
		"our", "overall", "strategy", "for", "growth", "in", "the", "next", "quarter",
	}

	result := ""
	for i := 0; i < wordCount; i++ {
		result += words[i%len(words)] + " "
	}
	return result + "?"
}
