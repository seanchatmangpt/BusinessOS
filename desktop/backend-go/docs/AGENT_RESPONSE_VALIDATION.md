# Agent Response Quality Validation

This document describes the Agent V2 response quality validation system.

## Overview

The response validator tests agent responses across multiple quality dimensions:

1. **Relevance** - Does the response address the user's input?
2. **Completeness** - Does it fully answer the question?
3. **Brevity** - Is it concise without being too short?
4. **Coherence** - Is it logically structured and readable?
5. **Voice-friendliness** - Is it appropriate for voice interaction?

## Architecture

### Core Components

#### `ResponseValidator` (`internal/agents/response_validator.go`)

The main validator that:
- Executes agent requests
- Measures latency (time to first token, total response time)
- Analyzes response characteristics
- Calculates quality scores
- Generates comprehensive reports

#### Quality Metrics

```go
type ResponseQualityMetrics struct {
    // Quality Scores (0-100)
    RelevanceScore    float64  // Semantic relevance to input
    CompletenessScore float64  // How fully it answers
    BrevityScore      float64  // Length appropriateness
    CoherenceScore    float64  // Logical flow and readability
    OverallScore      float64  // Weighted average

    // Voice-specific
    VoiceFriendliness      float64  // No code/tables/complex markdown
    VoiceAppropriateLength float64  // < 200 words ideal

    // Latency
    TimeToFirstToken  time.Duration  // Time until first response token
    TotalResponseTime time.Duration  // Total generation time
    TokensPerSecond   float64        // Throughput

    // Characteristics
    ResponseLength    int     // Character count
    WordCount         int     // Word count
    HasCodeBlocks     bool    // Contains code blocks
    HasTables         bool    // Contains tables
    HasMarkdown       bool    // Contains markdown
    SentenceCount     int     // Number of sentences
    AvgSentenceLength float64 // Average sentence length
}
```

### Quality Scoring

#### Relevance Score (0-100)

Measures semantic similarity between input and response using:
- **Keyword matching** - Key terms from input present in response
- **Expected pattern matching** - Response matches expected type (greeting, question, error, etc.)

Example:
```
Input: "What is machine learning?"
Response: "Machine learning is a subset of AI that uses algorithms to learn from data..."
Score: 95/100 (high keyword overlap + complete answer)
```

#### Completeness Score (0-100)

Evaluates if response fully addresses the input:
- **Question count** - Multi-part questions need comprehensive answers
- **Minimum word threshold** - Based on input complexity
- **Structure bonus** - Well-structured responses score higher

Example:
```
Input: "What is AI? How does it work? What are applications?"
Response: "AI is... It works by... Applications include..." (30+ words, addresses all 3 questions)
Score: 90/100
```

#### Brevity Score (0-100)

Penalizes responses that are too short or too long:
- **Ideal range**: 30-150 words
- **Too short**: < 30 words (incomplete)
- **Too long**: > 150 words (verbose, especially bad for voice)

Example:
```
Response: "Yes." (2 words)
Score: 6/100 (too short)

Response: 50 words, well-structured
Score: 100/100 (ideal)

Response: 300 words
Score: 65/100 (too long for conversational use)
```

#### Coherence Score (0-100)

Measures logical flow and readability:
- **Sentence length** - Not too long (hard to follow), not too short (choppy)
- **Connector words** - "however", "therefore", etc. improve flow
- **Repetition** - Penalizes excessive word repetition

Example:
```
Response: "First, understand basics. However, there are exceptions. Therefore, consider all factors."
Score: 95/100 (good connectors, balanced sentences)

Response: "Thing thing thing. The thing is thing thing thing."
Score: 60/100 (repetitive, poor coherence)
```

#### Voice Scores

**Voice Friendliness (0-100):**
- **Penalizes**: Code blocks (-40), tables (-30), markdown (-10)
- **Rewards**: Conversational tone (+10)

**Voice Appropriate Length (0-100):**
- **Ideal**: ≤ 100 words (100/100)
- **Acceptable**: ≤ 200 words (70-100/100)
- **Too long**: > 200 words (< 70/100)

#### Overall Score

Weighted average of all scores:
```
Overall = (Relevance × 0.30) +
          (Completeness × 0.25) +
          (Brevity × 0.15) +
          (Coherence × 0.20) +
          (Voice × 0.10)
```

### Test Cases

#### Standard Test Suite

Includes 12 comprehensive test cases:

1. **Greetings**
   - Simple greeting ("Hello")
   - Greeting with question ("Hi there! How are you?")

2. **Questions**
   - Simple factual ("What is the capital of France?")
   - Complex explanation ("Explain ML vs DL")
   - Multi-part ("What is PM? How to start? What tools?")

3. **Commands**
   - Task extraction
   - Document creation

4. **Context**
   - Follow-up questions ("Tell me more about that")

5. **Error Handling**
   - Out of scope ("Weather on Mars?")
   - Unclear input ("asdf jkl qwerty")

6. **Voice-specific**
   - Concise answers
   - Conversational responses

#### Custom Test Suite

Voice-optimized and business-focused tests:
- Voice commands
- Business queries (projects, tasks, documents)
- Edge cases (long input, special characters)

## Usage

### Run Tests Programmatically

```go
package main

import (
    "context"
    "github.com/rhl/businessos-backend/internal/agents"
    "github.com/rhl/businessos-backend/internal/config"
)

func main() {
    cfg, _ := config.Load()
    registry := agents.NewAgentRegistryV2(nil, cfg, nil, nil)
    validator := agents.NewResponseValidator(registry)

    // Run standard test suite
    report, err := validator.RunTestSuite(
        context.Background(),
        "standard_tests",
        agents.GetStandardTestSuite(),
        "user-id",
        "User Name",
    )

    // Export results
    jsonReport, _ := report.ExportReportJSON()
    // ... use report
}
```

### Run Tests via CLI

```bash
# Run standard test suite
go run ./cmd/validate_agents

# Run standard tests with output file
go run ./cmd/validate_agents -output report.json

# Run custom test suite
go run ./cmd/validate_agents -custom

# Run single test case
go run ./cmd/validate_agents -test simple_greeting

# Test specific agent type
go run ./cmd/validate_agents -custom -agent document

# Verbose logging
go run ./cmd/validate_agents -verbose
```

### Run Unit Tests

```bash
# Run all validator tests
go test ./internal/agents/response_validator_test.go -v

# Run integration tests (requires API keys)
INTEGRATION_TEST=true go test ./internal/agents/response_validator_test.go -v

# Run specific test
go test ./internal/agents/response_validator_test.go -run TestResponseValidator_AnalyzeResponseCharacteristics -v
```

## Output Format

### Console Output

```
════════════════════════════════════════════════════════════════════════════════
AGENT RESPONSE VALIDATION REPORT
════════════════════════════════════════════════════════════════════════════════
Test Suite: standard_test_suite
Total Tests: 12
Passed: 10
Failed: 2
Pass Rate: 83.3%
Average Overall Score: 82.50
Average Latency: 450ms
Total Duration: 5.4s
════════════════════════════════════════════════════════════════════════════════

DETAILED RESULTS:
────────────────────────────────────────────────────────────────────────────────

1. ✓ PASS simple_greeting
   Input: Hello
   Response: Hello! How can I assist you today?
   Scores: Overall=88.5, Relevance=92.0, Completeness=85.0, Brevity=90.0, Coherence=87.0
   Voice: Friendliness=95.0, Length=100.0
   Latency: First Token=120ms, Total=380ms, Tokens/sec=21.1

2. ✗ FAIL complex_question
   Input: Can you explain the difference between machine learning...
   Response: ML is complicated.
   Scores: Overall=45.2, Relevance=60.0, Completeness=35.0, Brevity=40.0, Coherence=50.0
   Voice: Friendliness=80.0, Length=100.0
   Latency: First Token=110ms, Total=250ms, Tokens/sec=16.0
   ⚠ Failure: completeness too low: 35.0 < 75.0
```

### JSON Report

```json
{
  "test_suite": "standard_test_suite",
  "total_tests": 12,
  "passed_tests": 10,
  "failed_tests": 2,
  "avg_overall_score": 82.5,
  "avg_latency_ms": 450,
  "test_results": [
    {
      "relevance_score": 92.0,
      "completeness_score": 85.0,
      "brevity_score": 90.0,
      "coherence_score": 87.0,
      "overall_score": 88.5,
      "voice_friendliness": 95.0,
      "voice_appropriate_length": 100.0,
      "time_to_first_token_ms": 120,
      "total_response_time_ms": 380,
      "tokens_per_second": 21.1,
      "response_length_chars": 42,
      "word_count": 7,
      "has_code_blocks": false,
      "has_tables": false,
      "has_markdown": false,
      "sentence_count": 2,
      "avg_sentence_length": 3.5,
      "test_case": "simple_greeting",
      "input": "Hello",
      "response": "Hello! How can I assist you today?",
      "expected_type": "greeting",
      "passed_test": true,
      "agent_type": "orchestrator",
      "model": "claude-3-5-sonnet-20241022",
      "timestamp": "2026-01-19T10:30:00Z"
    }
  ],
  "summary": "Test Suite: standard_test_suite | Pass Rate: 83.3% (10/12) | Avg Score: 82.5 | Avg Latency: 450ms",
  "timestamp": "2026-01-19T10:30:00Z"
}
```

## Interpreting Results

### Pass/Fail Criteria

A test passes if:
- **Relevance** ≥ threshold (default: 70)
- **Completeness** ≥ threshold (default: 75)
- **Coherence** ≥ threshold (default: 80)

### Quality Benchmarks

| Overall Score | Quality Level |
|---------------|---------------|
| 90-100        | Excellent     |
| 80-89         | Good          |
| 70-79         | Acceptable    |
| 60-69         | Needs Work    |
| < 60          | Poor          |

### Latency Benchmarks

| Metric            | Target     | Acceptable | Poor    |
|-------------------|------------|------------|---------|
| Time to First Token | < 200ms   | < 500ms    | > 500ms |
| Total Response    | < 1000ms   | < 2000ms   | > 2000ms |
| Tokens/sec        | > 30       | > 15       | < 15    |

### Voice Quality Benchmarks

For voice interactions:
- **VoiceFriendliness**: Should be > 85 (no code/tables)
- **VoiceAppropriateLength**: Should be > 85 (< 150 words)
- **WordCount**: Ideal 50-100 words

## Best Practices

### When to Run Validation

1. **During Development**
   - After implementing new agent features
   - Before deploying agent changes
   - When modifying prompts

2. **In CI/CD**
   - Run standard test suite in CI pipeline
   - Fail builds if pass rate < 80%
   - Track metrics over time

3. **For Quality Monitoring**
   - Weekly validation runs
   - Compare results across model versions
   - Identify regressions

### Writing Custom Test Cases

```go
customTest := agents.TestCase{
    Name:            "business_query",
    Input:           "What's the status of Project Alpha?",
    ExpectedType:    "question",
    MinRelevance:    75.0,
    MinCompleteness: 70.0,
    AgentType:       agents.AgentTypeV2Project,
    Context:         tieredContext, // Optional: provide business context
}
```

### Improving Agent Scores

If scores are low:

1. **Low Relevance**
   - Improve prompt instructions
   - Add examples to system prompt
   - Enhance context retrieval

2. **Low Completeness**
   - Instruct agent to provide comprehensive answers
   - Ensure sufficient context is provided
   - Check if agent has access to needed information

3. **Low Brevity**
   - Add "be concise" to prompt
   - Set max_tokens appropriately
   - For voice: emphasize brevity in prompt

4. **Low Coherence**
   - Improve prompt structure
   - Add examples of well-structured responses
   - Check for prompt confusion

5. **Low Voice Scores**
   - Instruct agent to avoid code/tables for voice
   - Emphasize conversational tone
   - Set target word count (< 100 words)

## Continuous Improvement

### Tracking Metrics Over Time

Store validation reports in a time-series database or log system:

```bash
# Run daily validation
go run ./cmd/validate_agents -output reports/$(date +%Y%m%d).json

# Compare results
diff reports/20260118.json reports/20260119.json
```

### A/B Testing Agent Changes

```go
// Test current agent
reportA, _ := validator.RunTestSuite(ctx, "agent_v1", testCases, ...)

// Test improved agent (with new prompt)
agentV2 := registry.GetAgent(...)
agentV2.SetCustomSystemPrompt(newPrompt)
reportB, _ := validator.RunTestSuite(ctx, "agent_v2", testCases, ...)

// Compare
fmt.Printf("V1 Score: %.2f\n", reportA.AvgOverallScore)
fmt.Printf("V2 Score: %.2f\n", reportB.AvgOverallScore)
```

## Limitations

### Current Limitations

1. **Relevance Scoring**: Uses simple keyword matching, not true semantic similarity
   - **Future**: Use embedding-based cosine similarity

2. **No Ground Truth**: Can't verify factual correctness
   - **Future**: Add knowledge base verification

3. **Single Turn**: Only tests single-turn interactions
   - **Future**: Add multi-turn conversation tests

4. **No User Feedback**: Doesn't incorporate actual user ratings
   - **Future**: Integrate with user feedback system

### Planned Enhancements

- [ ] Semantic similarity using embeddings
- [ ] Multi-turn conversation testing
- [ ] Factual accuracy verification
- [ ] User feedback integration
- [ ] Real-time monitoring dashboard
- [ ] Automated regression detection
- [ ] Performance benchmarking across models
- [ ] Voice-specific TTS simulation

## Examples

See `internal/agents/response_validator_test.go` for comprehensive test examples.

## Support

For questions or issues:
- Check test output for failure reasons
- Review agent prompts in `internal/prompts/agents/`
- Run with `-verbose` flag for detailed logging
- Check `ResponseQualityMetrics` for specific score breakdowns
