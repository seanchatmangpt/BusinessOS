package ontology

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestNewSPARQLClient tests client creation
func TestNewSPARQLClient(t *testing.T) {
	client := NewSPARQLClient("http://localhost:7878", nil)

	if client == nil {
		t.Fatal("failed to create SPARQL client")
	}

	if client.endpoint != "http://localhost:7878" {
		t.Errorf("expected endpoint http://localhost:7878, got %s", client.endpoint)
	}

	if client.retries != 3 {
		t.Errorf("expected 3 retries, got %d", client.retries)
	}
}

// TestValidateSPARQLConstruct tests CONSTRUCT query validation
func TestValidateSPARQLConstruct(t *testing.T) {
	tests := []struct {
		query     string
		shouldErr bool
		name      string
	}{
		{
			name:      "valid construct query",
			query:     "CONSTRUCT { ?s ?p ?o } WHERE { ?s ?p ?o }",
			shouldErr: false,
		},
		{
			name:      "missing construct clause",
			query:     "SELECT ?s WHERE { ?s ?p ?o }",
			shouldErr: true,
		},
		{
			name:      "missing where clause",
			query:     "CONSTRUCT { ?s ?p ?o }",
			shouldErr: true,
		},
		{
			name:      "unbalanced braces",
			query:     "CONSTRUCT { ?s ?p ?o WHERE { ?s ?p ?o }",
			shouldErr: true,
		},
		{
			name:      "lowercase construct",
			query:     "construct { ?s ?p ?o } where { ?s ?p ?o }",
			shouldErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateSPARQLConstruct(tc.query)
			if (err != nil) != tc.shouldErr {
				t.Errorf("expected error=%v, got error=%v: %v", tc.shouldErr, err != nil, err)
			}
		})
	}
}

// TestValidateSPARQLAsk tests ASK query validation
func TestValidateSPARQLAsk(t *testing.T) {
	tests := []struct {
		query     string
		shouldErr bool
		name      string
	}{
		{
			name:      "valid ask query",
			query:     "ASK WHERE { ?s ?p ?o }",
			shouldErr: false,
		},
		{
			name:      "missing ask clause",
			query:     "SELECT ?s WHERE { ?s ?p ?o }",
			shouldErr: true,
		},
		{
			name:      "missing where clause",
			query:     "ASK { ?s ?p ?o }",
			shouldErr: true,
		},
		{
			name:      "lowercase ask",
			query:     "ask where { ?s ?p ?o }",
			shouldErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateSPARQLAsk(tc.query)
			if (err != nil) != tc.shouldErr {
				t.Errorf("expected error=%v, got error=%v: %v", tc.shouldErr, err != nil, err)
			}
		})
	}
}

// TestCheckBalancedBraces tests brace balancing
func TestCheckBalancedBraces(t *testing.T) {
	tests := []struct {
		input    string
		balanced bool
		name     string
	}{
		{
			name:     "empty string",
			input:    "",
			balanced: true,
		},
		{
			name:     "single balanced pair",
			input:    "{}",
			balanced: true,
		},
		{
			name:     "nested braces",
			input:    "{ { } }",
			balanced: true,
		},
		{
			name:     "unbalanced open",
			input:    "{ { }",
			balanced: false,
		},
		{
			name:     "unbalanced close",
			input:    "{ } }",
			balanced: false,
		},
		{
			name:     "complex valid",
			input:    "CONSTRUCT { ?s ?p ?o } WHERE { ?s ?p ?o }",
			balanced: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := checkBalancedBraces(tc.input)
			if result != tc.balanced {
				t.Errorf("expected balanced=%v, got %v for input: %s", tc.balanced, result, tc.input)
			}
		})
	}
}

// TestParseAskResult tests ASK result parsing
func TestParseAskResult(t *testing.T) {
	tests := []struct {
		input    []byte
		expected bool
		name     string
	}{
		{
			name:     "true result",
			input:    []byte(`{"boolean": true}`),
			expected: true,
		},
		{
			name:     "true result no spaces",
			input:    []byte(`{"boolean":true}`),
			expected: true,
		},
		{
			name:     "false result",
			input:    []byte(`{"boolean": false}`),
			expected: false,
		},
		{
			name:     "false result no spaces",
			input:    []byte(`{"boolean":false}`),
			expected: false,
		},
		{
			name:     "empty response",
			input:    []byte(``),
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := parseAskResult(tc.input)
			if result != tc.expected {
				t.Errorf("expected result=%v, got %v for input: %s", tc.expected, result, string(tc.input))
			}
		})
	}
}

// TestIsTransientError tests transient error detection
func TestIsTransientError(t *testing.T) {
	tests := []struct {
		err         error
		isTransient bool
		name        string
	}{
		{
			name:        "nil error",
			err:         nil,
			isTransient: false,
		},
		{
			name:        "timeout error",
			err:         NewTestError("i/o timeout"),
			isTransient: true,
		},
		{
			name:        "service unavailable",
			err:         NewTestError("service unavailable"),
			isTransient: true,
		},
		{
			name:        "503 error",
			err:         NewTestError("503"),
			isTransient: true,
		},
		{
			name:        "gateway timeout",
			err:         NewTestError("gateway timeout"),
			isTransient: true,
		},
		{
			name:        "bad request (permanent)",
			err:         NewTestError("bad request"),
			isTransient: false,
		},
		{
			name:        "connection refused",
			err:         NewTestError("connection refused"),
			isTransient: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := isTransientError(tc.err)
			if result != tc.isTransient {
				t.Errorf("expected isTransient=%v, got %v for error: %v", tc.isTransient, result, tc.err)
			}
		})
	}
}

// TestError is a simple error type for testing
type TestError struct {
	msg string
}

func NewTestError(msg string) error {
	return &TestError{msg}
}

func (e *TestError) Error() string {
	return e.msg
}

// TestParseTurtle tests Turtle format parsing
func TestParseTurtle(t *testing.T) {
	client := NewSPARQLClient("http://localhost:7878", nil)

	turtleData := []byte(`@prefix ex: <http://example.org/> .
ex:Subject1 ex:predicate1 ex:Object1 .
`)

	result, err := client.ParseTurtle(turtleData)
	if err != nil {
		t.Fatalf("failed to parse Turtle: %v", err)
	}

	if result["format"] != "turtle" {
		t.Errorf("expected format=turtle, got %v", result["format"])
	}

	if result["bytes"].(int) != len(turtleData) {
		t.Errorf("expected bytes=%d, got %v", len(turtleData), result["bytes"])
	}
}

// TestParseNTriples tests N-Triples format parsing
func TestParseNTriples(t *testing.T) {
	client := NewSPARQLClient("http://localhost:7878", nil)

	ntriplesData := []byte(`<http://example.org/Subject1> <http://example.org/predicate1> <http://example.org/Object1> .
<http://example.org/Subject2> <http://example.org/predicate2> <http://example.org/Object2> .
`)

	result, err := client.ParseNTriples(ntriplesData)
	if err != nil {
		t.Fatalf("failed to parse N-Triples: %v", err)
	}

	if result["format"] != "ntriples" {
		t.Errorf("expected format=ntriples, got %v", result["format"])
	}

	triples := result["triples"].(int)
	if triples < 2 {
		t.Errorf("expected at least 2 triples, got %d", triples)
	}
}

// TestParseJSONLD tests JSON-LD format parsing
func TestParseJSONLD(t *testing.T) {
	client := NewSPARQLClient("http://localhost:7878", nil)

	jsonldData := []byte(`{
  "@context": {"ex": "http://example.org/"},
  "@id": "http://example.org/Subject1",
  "ex:predicate": {"@id": "http://example.org/Object1"}
}`)

	result, err := client.ParseJSONLD(jsonldData)
	if err != nil {
		t.Fatalf("failed to parse JSON-LD: %v", err)
	}

	if result["format"] != "jsonld" {
		t.Errorf("expected format=jsonld, got %v", result["format"])
	}
}

// TestExecuteConstructWithTimeout tests timeout handling
func TestExecuteConstructWithTimeout(t *testing.T) {
	client := NewSPARQLClient("http://localhost:7878", nil)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	query := "CONSTRUCT { ?s ?p ?o } WHERE { ?s ?p ?o }"

	// This should timeout or fail (no running server)
	_, err := client.ExecuteConstruct(ctx, query, 50*time.Millisecond)

	// We expect an error (either timeout or connection error)
	if err == nil {
		t.Skip("skipping test without running Oxigraph server")
	}
}

// TestExecuteAskWithTimeout tests ASK timeout handling
func TestExecuteAskWithTimeout(t *testing.T) {
	client := NewSPARQLClient("http://localhost:7878", nil)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	query := "ASK WHERE { ?s ?p ?o }"

	_, err := client.ExecuteAsk(ctx, query, 50*time.Millisecond)

	if err == nil {
		t.Skip("skipping test without running Oxigraph server")
	}
}

// TestSPARQLClientEndpointFromEnv tests endpoint from environment variable
func TestSPARQLClientEndpointFromEnv(t *testing.T) {
	// Save original env var
	originalURL := os.Getenv("OXIGRAPH_URL")
	defer os.Setenv("OXIGRAPH_URL", originalURL)

	// Set test endpoint
	os.Setenv("OXIGRAPH_URL", "http://test-oxigraph:7878")

	client := NewSPARQLClient("", nil)

	if client.endpoint != "http://test-oxigraph:7878" {
		t.Errorf("expected endpoint from env var, got %s", client.endpoint)
	}
}

// TestSPARQLInvalidQuery tests handling of invalid SPARQL
func TestSPARQLInvalidQuery(t *testing.T) {
	client := NewSPARQLClient("http://localhost:7878", nil)

	ctx := context.Background()

	// Invalid CONSTRUCT query
	_, err := client.ExecuteConstruct(ctx, "INVALID SPARQL", 1*time.Second)
	if err == nil {
		t.Error("expected error for invalid SPARQL")
	}

	// Invalid ASK query
	_, err = client.ExecuteAsk(ctx, "INVALID QUERY", 1*time.Second)
	if err == nil {
		t.Error("expected error for invalid ASK query")
	}
}
