package ontology

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"
)

// SPARQLClient provides high-performance SPARQL query execution
type SPARQLClient struct {
	endpoint   string
	httpClient *http.Client
	logger     *slog.Logger
	retries    int
	backoffMs  int
}

// NewSPARQLClient creates a new SPARQL client
// endpoint: base URL of Oxigraph (e.g., http://localhost:7878)
func NewSPARQLClient(endpoint string, logger *slog.Logger) *SPARQLClient {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
	}

	if endpoint == "" {
		endpoint = os.Getenv("OXIGRAPH_URL")
		if endpoint == "" {
			endpoint = "http://localhost:7878"
		}
	}

	return &SPARQLClient{
		endpoint: endpoint,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxConnsPerHost:     100,
				MaxIdleConns:        50,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		logger:    logger,
		retries:   3,
		backoffMs: 100,
	}
}

// ExecuteConstruct executes a SPARQL CONSTRUCT query
// Returns RDF data in the requested format
func (c *SPARQLClient) ExecuteConstruct(ctx context.Context, query string, timeout time.Duration) ([]byte, error) {
	if timeout == 0 {
		timeout = 5 * time.Second // Default CONSTRUCT timeout
	}

	// Apply timeout to context
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Validate SPARQL syntax
	if err := validateSPARQLConstruct(query); err != nil {
		return nil, fmt.Errorf("invalid SPARQL: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt < c.retries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			backoff := time.Duration(c.backoffMs*(1<<uint(attempt-1))) * time.Millisecond
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		result, err := c.executeRequest(ctx, query, "CONSTRUCT", "text/turtle")
		if err == nil {
			return result, nil
		}

		lastErr = err
		c.logger.Warn("CONSTRUCT query attempt failed",
			"attempt", attempt+1,
			"error", err,
		)

		// Distinguish transient vs permanent errors
		if !isTransientError(err) {
			return nil, err // Don't retry non-transient errors
		}
	}

	return nil, fmt.Errorf("CONSTRUCT query failed after %d retries: %w", c.retries, lastErr)
}

// ExecuteSelect executes a SPARQL SELECT query
// Returns JSON results
func (c *SPARQLClient) ExecuteSelect(ctx context.Context, query string, timeout time.Duration) ([]byte, error) {
	if timeout == 0 {
		timeout = 5 * time.Second // Default SELECT timeout
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Validate SPARQL syntax
	if err := validateSPARQLSelect(query); err != nil {
		return nil, fmt.Errorf("invalid SPARQL: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt < c.retries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(c.backoffMs*(1<<uint(attempt-1))) * time.Millisecond
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		result, err := c.executeRequest(ctx, query, "SELECT", "application/sparql-results+json")
		if err == nil {
			return result, nil
		}

		lastErr = err
		c.logger.Warn("SELECT query attempt failed",
			"attempt", attempt+1,
			"error", err,
		)

		if !isTransientError(err) {
			return nil, err
		}
	}

	return nil, fmt.Errorf("SELECT query failed after %d retries: %w", c.retries, lastErr)
}

// ExecuteAsk executes a SPARQL ASK query
// Returns a boolean result
func (c *SPARQLClient) ExecuteAsk(ctx context.Context, query string, timeout time.Duration) (bool, error) {
	if timeout == 0 {
		timeout = 3 * time.Second // Default ASK timeout (compliance checks are faster)
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Validate SPARQL syntax
	if err := validateSPARQLAsk(query); err != nil {
		return false, fmt.Errorf("invalid SPARQL: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt < c.retries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(c.backoffMs*(1<<uint(attempt-1))) * time.Millisecond
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return false, ctx.Err()
			}
		}

		result, err := c.executeRequest(ctx, query, "ASK", "application/sparql-results+json")
		if err == nil {
			// Parse boolean result from SPARQL JSON response
			return parseAskResult(result), nil
		}

		lastErr = err
		c.logger.Warn("ASK query attempt failed",
			"attempt", attempt+1,
			"error", err,
		)

		if !isTransientError(err) {
			return false, err
		}
	}

	return false, fmt.Errorf("ASK query failed after %d retries: %w", c.retries, lastErr)
}

// executeRequest performs the HTTP request to the SPARQL endpoint
func (c *SPARQLClient) executeRequest(ctx context.Context, query string, queryType string, acceptHeader string) ([]byte, error) {
	body := bytes.NewBufferString(fmt.Sprintf("query=%s", query))

	req, err := http.NewRequestWithContext(ctx, "POST", c.endpoint+"/query", body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", acceptHeader)
	req.Header.Set("User-Agent", "BusinessOS-OntologyRegistry/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	result, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Handle SPARQL error responses
	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case http.StatusBadRequest:
			return nil, fmt.Errorf("SPARQL 400 bad query: %s", string(result))
		case http.StatusServiceUnavailable:
			return nil, fmt.Errorf("SPARQL 503 service unavailable")
		case http.StatusGatewayTimeout, http.StatusRequestTimeout:
			return nil, fmt.Errorf("SPARQL timeout")
		default:
			return nil, fmt.Errorf("SPARQL %d error: %s", resp.StatusCode, string(result))
		}
	}

	return result, nil
}

// ParseTurtle parses Turtle RDF format
func (c *SPARQLClient) ParseTurtle(data []byte) (map[string]interface{}, error) {
	// Lightweight validation and parsing
	// Full implementation would use a real RDF parser
	if !bytes.Contains(data, []byte("@prefix")) && !bytes.Contains(data, []byte("<http")) {
		return nil, fmt.Errorf("invalid Turtle format")
	}

	return map[string]interface{}{
		"format": "turtle",
		"bytes":  len(data),
	}, nil
}

// ParseNTriples parses N-Triples RDF format
func (c *SPARQLClient) ParseNTriples(data []byte) (map[string]interface{}, error) {
	// Count triples (each line is a triple)
	lineCount := bytes.Count(data, []byte("\n"))

	return map[string]interface{}{
		"format":  "ntriples",
		"triples": lineCount,
		"bytes":   len(data),
	}, nil
}

// ParseJSONLD parses JSON-LD format
func (c *SPARQLClient) ParseJSONLD(data []byte) (map[string]interface{}, error) {
	// Lightweight validation
	if !bytes.Contains(data, []byte("@context")) && !bytes.Contains(data, []byte("{")) {
		return nil, fmt.Errorf("invalid JSON-LD format")
	}

	return map[string]interface{}{
		"format": "jsonld",
		"bytes":  len(data),
	}, nil
}

// Close closes the HTTP client
func (c *SPARQLClient) Close() error {
	c.httpClient.CloseIdleConnections()
	return nil
}

// Helper functions

// validateSPARQLConstruct validates CONSTRUCT query syntax
func validateSPARQLConstruct(query string) error {
	upper := strings.ToUpper(strings.TrimSpace(query))

	if !strings.Contains(upper, "CONSTRUCT") {
		return fmt.Errorf("missing CONSTRUCT clause")
	}

	if !strings.Contains(upper, "WHERE") {
		return fmt.Errorf("missing WHERE clause")
	}

	// Check balanced braces
	if !checkBalancedBraces(query) {
		return fmt.Errorf("unbalanced braces")
	}

	return nil
}

// validateSPARQLSelect validates SELECT query syntax
func validateSPARQLSelect(query string) error {
	upper := strings.ToUpper(strings.TrimSpace(query))

	if !strings.Contains(upper, "SELECT") {
		return fmt.Errorf("missing SELECT clause")
	}

	if !strings.Contains(upper, "WHERE") {
		return fmt.Errorf("missing WHERE clause")
	}

	// Check balanced braces
	if !checkBalancedBraces(query) {
		return fmt.Errorf("unbalanced braces")
	}

	return nil
}

// validateSPARQLAsk validates ASK query syntax
func validateSPARQLAsk(query string) error {
	upper := strings.ToUpper(strings.TrimSpace(query))

	if !strings.Contains(upper, "ASK") {
		return fmt.Errorf("missing ASK clause")
	}

	if !strings.Contains(upper, "WHERE") {
		return fmt.Errorf("missing WHERE clause")
	}

	if !checkBalancedBraces(query) {
		return fmt.Errorf("unbalanced braces")
	}

	return nil
}

// checkBalancedBraces checks if braces are balanced
func checkBalancedBraces(query string) bool {
	var depth int
	for _, ch := range query {
		switch ch {
		case '{':
			depth++
		case '}':
			depth--
			if depth < 0 {
				return false
			}
		}
	}
	return depth == 0
}

// parseAskResult parses the boolean result from ASK query
func parseAskResult(data []byte) bool {
	// Look for "true" or "false" in the response
	str := string(data)
	return strings.Contains(str, "\"boolean\": true") || strings.Contains(str, "\"boolean\":true")
}

// isTransientError determines if an error is transient (retry-able)
func isTransientError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Transient errors (retry)
	transient := []string{
		"timeout",
		"service unavailable",
		"503",
		"gateway timeout",
		"504",
		"connection refused",
		"connection reset",
		"i/o timeout",
	}

	for _, pattern := range transient {
		if strings.Contains(strings.ToLower(errStr), pattern) {
			return true
		}
	}

	return false
}
