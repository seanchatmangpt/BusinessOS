package ontology

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// OntologyMetadata holds information about a loaded ontology
type OntologyMetadata struct {
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	Format      string    `json:"format"` // turtle, ntriples, jsonld
	TripleCount int       `json:"triple_count"`
	LoadedAt    time.Time `json:"loaded_at"`
	FileSize    int64     `json:"file_size"`
}

// QueryStats holds performance metrics for queries
type QueryStats struct {
	TotalQueries    int64           `json:"total_queries"`
	CacheHits       int64           `json:"cache_hits"`
	CacheMisses     int64           `json:"cache_misses"`
	LatencyP50Ms    float64         `json:"latency_p50_ms"`
	LatencyP95Ms    float64         `json:"latency_p95_ms"`
	LatencyP99Ms    float64         `json:"latency_p99_ms"`
	CacheHitRate    float64         `json:"cache_hit_rate"`
	AvgLatencyMs    float64         `json:"avg_latency_ms"`
	RecentLatencies []time.Duration `json:"-"` // Keep last 100 latencies for p-ile calculation
}

// Registry manages loaded ontologies and SPARQL query execution
type Registry struct {
	mu              sync.RWMutex
	ontologies      map[string]*OntologyMetadata
	queryCache      *LRUCache
	client          *SPARQLClient
	logger          *slog.Logger
	stats           QueryStats
	ontologyDir     string
	env             string // dev-minimal, staging-f5, production-f5
	latenciesMu     sync.Mutex
	recentLatencies []time.Duration
	maxLatencies    int
}

// NewRegistry creates a new ontology registry
func NewRegistry(client *SPARQLClient, logger *slog.Logger, env string) *Registry {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
	}
	if env == "" {
		env = os.Getenv("ONTOLOGY_ENV")
		if env == "" {
			env = "dev-minimal"
		}
	}

	return &Registry{
		ontologies:      make(map[string]*OntologyMetadata),
		queryCache:      NewLRUCache(1000), // 1000 item cache
		client:          client,
		logger:          logger,
		env:             env,
		recentLatencies: make([]time.Duration, 0, 100),
		maxLatencies:    100,
	}
}

// LoadOntologies loads all ontologies from the specified directory
// Expects structure: <dir>/*.ttl, *.rdf, *.jsonld
func (r *Registry) LoadOntologies(ontologyDir string) error {
	if _, err := os.Stat(ontologyDir); err != nil {
		return fmt.Errorf("ontology directory not found: %w", err)
	}

	r.ontologyDir = ontologyDir
	var loadedCount int
	var failedCount int
	var failedFiles []string

	err := filepath.WalkDir(ontologyDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		if ext != ".ttl" && ext != ".rdf" && ext != ".jsonld" && ext != ".rq" {
			return nil
		}

		if err := r.loadOntology(path); err != nil {
			r.logger.Warn("failed to load ontology", "path", path, "error", err)
			failedCount++
			failedFiles = append(failedFiles, path)
			return nil // Continue loading others
		}
		loadedCount++
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk directory: %w", err)
	}

	// For environment-specific loading, fail fast if critical ontologies are missing
	if r.env != "dev-minimal" && failedCount > 0 {
		return fmt.Errorf("failed to load %d ontologies (env=%s): %v", failedCount, r.env, failedFiles)
	}

	r.logger.Info("ontologies loaded",
		"loaded", loadedCount,
		"failed", failedCount,
		"total", len(r.ontologies),
		"env", r.env,
	)

	return nil
}

// loadOntology loads a single ontology file
func (r *Registry) loadOntology(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	filename := filepath.Base(path)
	ext := filepath.Ext(filename)

	metadata := &OntologyMetadata{
		Name:     filename[:len(filename)-len(ext)],
		Path:     path,
		Format:   ext[1:], // Remove leading dot
		LoadedAt: time.Now().UTC(),
		FileSize: info.Size(),
	}

	// Count triples by parsing (lightweight check)
	tripleCount, err := countTriples(path, metadata.Format)
	if err != nil {
		r.logger.Warn("failed to count triples", "path", path, "error", err)
		tripleCount = 0 // Continue with 0
	}
	metadata.TripleCount = tripleCount

	r.mu.Lock()
	r.ontologies[metadata.Name] = metadata
	r.mu.Unlock()

	r.logger.Debug("loaded ontology",
		"name", metadata.Name,
		"format", metadata.Format,
		"triples", metadata.TripleCount,
		"size_bytes", metadata.FileSize,
	)

	return nil
}

// countTriples provides a lightweight triple count (line count heuristic)
func countTriples(path string, format string) (int, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}

	// Simple heuristic: count lines with predicates
	// Real implementation would parse and count actual triples
	var count int
	for _, b := range content {
		if b == '\n' {
			count++
		}
	}
	return count / 3, nil // Rough estimate
}

// ExecuteSPARQLConstruct executes a CONSTRUCT query with caching
func (r *Registry) ExecuteSPARQLConstruct(ctx context.Context, query string, timeout time.Duration) ([]byte, error) {
	// Check cache first (O(1) lookup)
	if cached, found := r.queryCache.Get(query); found {
		r.recordCacheHit()
		r.logger.Debug("cache hit", "query_len", len(query))
		return cached.([]byte), nil
	}

	r.recordCacheMiss()

	// Execute query with timeout
	start := time.Now()
	result, err := r.client.ExecuteConstruct(ctx, query, timeout)
	latency := time.Since(start)
	r.recordLatency(latency)

	if err != nil {
		r.logger.Error("CONSTRUCT query failed",
			"query_len", len(query),
			"latency_ms", latency.Milliseconds(),
			"error", err,
		)
		return nil, err
	}

	// Cache the result (max 1000 entries, LRU eviction)
	r.queryCache.Put(query, result)

	r.logger.Debug("CONSTRUCT query executed",
		"result_size", len(result),
		"latency_ms", latency.Milliseconds(),
	)

	return result, nil
}

// ExecuteSPARQLAsk executes an ASK query (boolean result)
func (r *Registry) ExecuteSPARQLAsk(ctx context.Context, query string, timeout time.Duration) (bool, error) {
	start := time.Now()
	result, err := r.client.ExecuteAsk(ctx, query, timeout)
	latency := time.Since(start)
	r.recordLatency(latency)

	if err != nil {
		r.logger.Error("ASK query failed",
			"query_len", len(query),
			"latency_ms", latency.Milliseconds(),
			"error", err,
		)
		return false, err
	}

	r.logger.Debug("ASK query executed",
		"result", result,
		"latency_ms", latency.Milliseconds(),
	)

	return result, nil
}

// GetOntology returns metadata for a loaded ontology
func (r *Registry) GetOntology(name string) (*OntologyMetadata, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	meta, exists := r.ontologies[name]
	if !exists {
		return nil, fmt.Errorf("ontology not found: %s", name)
	}
	return meta, nil
}

// ListOntologies returns all loaded ontologies
func (r *Registry) ListOntologies() []*OntologyMetadata {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*OntologyMetadata, 0, len(r.ontologies))
	for _, meta := range r.ontologies {
		result = append(result, meta)
	}
	return result
}

// OntologyCount returns the number of loaded ontologies
func (r *Registry) OntologyCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.ontologies)
}

// ReloadRegistry clears cache and reloads all ontologies from disk
func (r *Registry) ReloadRegistry() error {
	r.mu.Lock()
	r.ontologies = make(map[string]*OntologyMetadata)
	r.mu.Unlock()

	r.queryCache.Clear()

	// Reload from ontologyDir
	if r.ontologyDir == "" {
		return fmt.Errorf("no ontology directory configured")
	}

	return r.LoadOntologies(r.ontologyDir)
}

// GetQueryStats returns current query performance statistics
func (r *Registry) GetQueryStats() QueryStats {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := r.stats
	stats.RecentLatencies = nil // Don't expose internal slice

	// Recalculate percentiles from recent latencies
	if len(r.recentLatencies) > 0 {
		stats.LatencyP50Ms = calculatePercentile(r.recentLatencies, 50)
		stats.LatencyP95Ms = calculatePercentile(r.recentLatencies, 95)
		stats.LatencyP99Ms = calculatePercentile(r.recentLatencies, 99)

		var totalMs float64
		for _, lat := range r.recentLatencies {
			totalMs += lat.Seconds() * 1000
		}
		stats.AvgLatencyMs = totalMs / float64(len(r.recentLatencies))
	}

	if stats.TotalQueries > 0 {
		stats.CacheHitRate = float64(stats.CacheHits) / float64(stats.TotalQueries)
	}

	return stats
}

// recordCacheHit increments cache hit counter
func (r *Registry) recordCacheHit() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stats.CacheHits++
	r.stats.TotalQueries++
}

// recordCacheMiss increments cache miss counter
func (r *Registry) recordCacheMiss() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stats.CacheMisses++
	r.stats.TotalQueries++
}

// recordLatency adds a latency measurement
func (r *Registry) recordLatency(latency time.Duration) {
	r.latenciesMu.Lock()
	defer r.latenciesMu.Unlock()

	r.recentLatencies = append(r.recentLatencies, latency)
	if len(r.recentLatencies) > r.maxLatencies {
		r.recentLatencies = r.recentLatencies[1:] // Keep sliding window
	}
}

// calculatePercentile calculates the pth percentile of latencies
func calculatePercentile(latencies []time.Duration, p int) float64 {
	if len(latencies) == 0 {
		return 0
	}

	// Simple percentile (not interpolated)
	index := (len(latencies) * p) / 100
	if index >= len(latencies) {
		index = len(latencies) - 1
	}

	return float64(latencies[index].Milliseconds())
}

// Close closes the SPARQL client connection
func (r *Registry) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}
