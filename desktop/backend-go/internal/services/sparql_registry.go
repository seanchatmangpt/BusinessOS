package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// QueryMetadata holds parsed metadata from SPARQL query comments
type QueryMetadata struct {
	Name               string   `json:"name"`
	Category           string   `json:"category"`
	Description        string   `json:"description"`
	Params             []string `json:"params"`
	Returns            string   `json:"returns"`
	Version            string   `json:"version"`
	Stability          string   `json:"stability"`
	RequiresParameters bool     `json:"requires_parameters"`
}

// ConstructQuery represents a SPARQL CONSTRUCT query with metadata
type ConstructQuery struct {
	Metadata QueryMetadata `json:"metadata"`
	SPARQL   string        `json:"sparql"`
	Path     string        `json:"path"`
}

// BoundQuery is a query with parameters bound and ready to execute
type BoundQuery struct {
	query    *ConstructQuery
	bindings map[string]string
}

// RegistryStats holds statistics about the registry
type RegistryStats struct {
	TotalQueries       int            `json:"total_queries"`
	Categories         int            `json:"categories"`
	QueriesPerCategory map[string]int `json:"queries_per_category"`
}

// SparqlRegistry manages SPARQL CONSTRUCT queries
type SparqlRegistry struct {
	queriesByName     map[string]*ConstructQuery
	queriesByCategory map[string][]*ConstructQuery
	rootDir           string
	logger            *slog.Logger
}

// NewSparqlRegistry creates a new registry loader
func NewSparqlRegistry(logger *slog.Logger) *SparqlRegistry {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
	}

	return &SparqlRegistry{
		queriesByName:     make(map[string]*ConstructQuery),
		queriesByCategory: make(map[string][]*ConstructQuery),
		logger:            logger,
	}
}

// Load discovers and loads all .rq files from a directory
//
// Expected structure:
//
//	ontologies/sparql/constructs/
//	├── artifacts/
//	├── projects/
//	├── compliance/
//	├── org/
//	├── process/
//	├── signal/
//	└── agents/
func (r *SparqlRegistry) Load(rootDir string) error {
	if _, err := os.Stat(rootDir); err != nil {
		return fmt.Errorf("registry root not found: %w", err)
	}

	r.rootDir = rootDir

	// Walk directory tree discovering .rq files
	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && filepath.Ext(path) == ".rq" {
			if err := r.loadQuery(path); err != nil {
				r.logger.Warn("failed to load query", "path", path, "error", err)
				// Continue loading other queries
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk directory: %w", err)
	}

	r.logger.Info("registry loaded",
		"total_queries", len(r.queriesByName),
		"categories", len(r.queriesByCategory),
	)

	return nil
}

// loadQuery loads a single .rq file
func (r *SparqlRegistry) loadQuery(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	sparql := string(content)

	// Parse metadata from comments
	metadata, err := parseMetadata(sparql, path)
	if err != nil {
		return err
	}

	query := &ConstructQuery{
		Metadata: metadata,
		SPARQL:   sparql,
		Path:     path,
	}

	// Index by name
	r.queriesByName[metadata.Name] = query

	// Index by category
	r.queriesByCategory[metadata.Category] = append(
		r.queriesByCategory[metadata.Category],
		query,
	)

	return nil
}

// parseMetadata extracts metadata from .rq file comments
func parseMetadata(sparql string, filename string) (QueryMetadata, error) {
	metadata := QueryMetadata{
		Version:   "1.0",
		Stability: "stable",
	}

	// Extract filename without extension as fallback name
	baseName := filepath.Base(filename)
	baseName = strings.TrimSuffix(baseName, filepath.Ext(baseName))

	// Parse comment lines at the start
	lines := strings.Split(sparql, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "#") {
			break
		}

		trimmed = strings.TrimPrefix(trimmed, "#")
		trimmed = strings.TrimSpace(trimmed)

		// Parse metadata fields
		if strings.HasPrefix(trimmed, "@name:") {
			metadata.Name = strings.TrimSpace(strings.TrimPrefix(trimmed, "@name:"))
		} else if strings.HasPrefix(trimmed, "@category:") {
			metadata.Category = strings.TrimSpace(strings.TrimPrefix(trimmed, "@category:"))
		} else if strings.HasPrefix(trimmed, "@description:") {
			metadata.Description = strings.TrimSpace(strings.TrimPrefix(trimmed, "@description:"))
		} else if strings.HasPrefix(trimmed, "@params:") {
			paramsStr := strings.TrimSpace(strings.TrimPrefix(trimmed, "@params:"))
			for _, p := range strings.Split(paramsStr, ",") {
				metadata.Params = append(metadata.Params, strings.TrimSpace(p))
			}
		} else if strings.HasPrefix(trimmed, "@returns:") {
			metadata.Returns = strings.TrimSpace(strings.TrimPrefix(trimmed, "@returns:"))
		} else if strings.HasPrefix(trimmed, "@version:") {
			metadata.Version = strings.TrimSpace(strings.TrimPrefix(trimmed, "@version:"))
		} else if strings.HasPrefix(trimmed, "@stability:") {
			metadata.Stability = strings.TrimSpace(strings.TrimPrefix(trimmed, "@stability:"))
		} else if strings.HasPrefix(trimmed, "@requires_parameters:") {
			val := strings.TrimSpace(strings.TrimPrefix(trimmed, "@requires_parameters:"))
			metadata.RequiresParameters = strings.ToLower(val) == "true"
		}
	}

	// Validate required fields
	if metadata.Name == "" {
		metadata.Name = baseName // Use filename as fallback
	}

	if metadata.Category == "" {
		return QueryMetadata{}, errors.New("missing @category in query")
	}

	return metadata, nil
}

// Lookup retrieves a query by name
func (r *SparqlRegistry) Lookup(name string) (*ConstructQuery, error) {
	q, exists := r.queriesByName[name]
	if !exists {
		return nil, fmt.Errorf("query not found: %s", name)
	}
	return q, nil
}

// QueriesInCategory returns all queries in a category
func (r *SparqlRegistry) QueriesInCategory(category string) ([]*ConstructQuery, error) {
	queries, exists := r.queriesByCategory[category]
	if !exists {
		return nil, fmt.Errorf("category not found: %s", category)
	}
	return queries, nil
}

// Categories returns all available categories
func (r *SparqlRegistry) Categories() []string {
	cats := make([]string, 0, len(r.queriesByCategory))
	for cat := range r.queriesByCategory {
		cats = append(cats, cat)
	}
	sort.Strings(cats)
	return cats
}

// QueryNames returns all available query names
func (r *SparqlRegistry) QueryNames() []string {
	names := make([]string, 0, len(r.queriesByName))
	for name := range r.queriesByName {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Stats returns registry statistics
func (r *SparqlRegistry) Stats() RegistryStats {
	stats := RegistryStats{
		TotalQueries:       len(r.queriesByName),
		Categories:         len(r.queriesByCategory),
		QueriesPerCategory: make(map[string]int),
	}

	for cat, queries := range r.queriesByCategory {
		stats.QueriesPerCategory[cat] = len(queries)
	}

	return stats
}

// Bind creates a bound query with parameter values
//
// Usage:
//
//	bound, err := query.Bind(map[string]string{
//	  "artifactId": "art-123",
//	  "title": "API Spec",
//	})
//	sparql := bound.SPARQLWithBindings()
func (q *ConstructQuery) Bind(params map[string]string) (*BoundQuery, error) {
	// Validate required parameters
	if q.Metadata.RequiresParameters {
		for _, param := range q.Metadata.Params {
			paramName := strings.TrimPrefix(param, "?")
			if _, exists := params[paramName]; !exists {
				return nil, fmt.Errorf("missing required parameter: %s", paramName)
			}
		}
	}

	return &BoundQuery{
		query:    q,
		bindings: params,
	}, nil
}

// SPARQLWithBindings returns the SPARQL query with BIND statements injected
func (b *BoundQuery) SPARQLWithBindings() string {
	result := b.query.SPARQL

	// Find WHERE clause
	wherePos := strings.Index(result, "WHERE {")
	if wherePos == -1 {
		return result
	}

	insertPos := wherePos + len("WHERE {")

	// Build BIND statements
	var bindings strings.Builder
	for key, value := range b.bindings {
		var escaped string

		switch {
		case strings.HasPrefix(value, "?"):
			// Variable reference
			escaped = value
		case strings.HasPrefix(value, "http://"), strings.HasPrefix(value, "https://"):
			// IRI
			escaped = fmt.Sprintf("<%s>", value)
		default:
			// String literal - escape quotes
			escaped = fmt.Sprintf("\"%s\"", strings.ReplaceAll(value, "\"", "\\\""))
		}

		fmt.Fprintf(&bindings, "\n  BIND(%s as ?%s)", escaped, key)
	}

	// Insert bindings
	return result[:insertPos] + bindings.String() + result[insertPos:]
}

// Query returns the original unbound query
func (b *BoundQuery) Query() *ConstructQuery {
	return b.query
}

// Bindings returns the parameter bindings
func (b *BoundQuery) Bindings() map[string]string {
	return b.bindings
}

// ValidateSPARQL performs basic validation of SPARQL syntax
func ValidateSPARQL(sparql string) error {
	// Check for required clauses
	if !strings.Contains(strings.ToUpper(sparql), "CONSTRUCT") {
		return errors.New("missing CONSTRUCT clause")
	}

	if !strings.Contains(strings.ToUpper(sparql), "WHERE") {
		return errors.New("missing WHERE clause")
	}

	// Check for balanced braces
	constructCount := strings.Count(strings.ToUpper(sparql), "CONSTRUCT {")
	whereCount := strings.Count(strings.ToUpper(sparql), "WHERE {")
	closeBraceCount := strings.Count(sparql, "}")

	expectedBraces := constructCount + whereCount
	if closeBraceCount < expectedBraces {
		return errors.New("unbalanced braces in SPARQL")
	}

	return nil
}

// ExportAsJSON exports registry metadata as JSON
func (r *SparqlRegistry) ExportAsJSON() ([]byte, error) {
	queries := make([]*ConstructQuery, 0, len(r.queriesByName))
	for _, q := range r.queriesByName {
		queries = append(queries, q)
	}

	return json.MarshalIndent(queries, "", "  ")
}

// ReloadCategory reloads queries in a specific category
func (r *SparqlRegistry) ReloadCategory(category string) error {
	catPath := filepath.Join(r.rootDir, category)

	if _, err := os.Stat(catPath); err != nil {
		return fmt.Errorf("category directory not found: %s", category)
	}

	// Clear category
	delete(r.queriesByCategory, category)

	// Reload
	err := filepath.WalkDir(catPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && filepath.Ext(path) == ".rq" {
			if err := r.loadQuery(path); err != nil {
				r.logger.Warn("failed to reload query", "path", path, "error", err)
			}
		}

		return nil
	})

	return err
}
