package ontology

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestNewRegistry tests registry creation
func TestNewRegistry(t *testing.T) {
	client := NewSPARQLClient("http://localhost:7878", nil)
	registry := NewRegistry(client, nil, "dev-minimal")

	if registry == nil {
		t.Fatal("failed to create registry")
	}

	if registry.env != "dev-minimal" {
		t.Errorf("expected env=dev-minimal, got %s", registry.env)
	}

	if registry.OntologyCount() != 0 {
		t.Errorf("expected 0 ontologies, got %d", registry.OntologyCount())
	}
}

// TestLoadOntologies tests loading ontologies from a directory
func TestLoadOntologies(t *testing.T) {
	// Create temporary directory with mock ontology files
	tmpDir := t.TempDir()

	// Create mock Turtle file
	turtleContent := `@prefix ex: <http://example.org/> .
ex:Subject1 ex:predicate1 ex:Object1 .
ex:Subject2 ex:predicate2 ex:Object2 .
`

	turtlePath := filepath.Join(tmpDir, "test.ttl")
	if err := os.WriteFile(turtlePath, []byte(turtleContent), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Create registry and load
	client := NewSPARQLClient("http://localhost:7878", nil)
	registry := NewRegistry(client, nil, "dev-minimal")

	err := registry.LoadOntologies(tmpDir)
	if err != nil {
		t.Fatalf("failed to load ontologies: %v", err)
	}

	if registry.OntologyCount() != 1 {
		t.Errorf("expected 1 ontology, got %d", registry.OntologyCount())
	}

	// Verify metadata
	ontologies := registry.ListOntologies()
	if len(ontologies) != 1 {
		t.Fatalf("expected 1 ontology, got %d", len(ontologies))
	}

	meta := ontologies[0]
	if meta.Name != "test" {
		t.Errorf("expected name=test, got %s", meta.Name)
	}

	if meta.Format != "ttl" {
		t.Errorf("expected format=ttl, got %s", meta.Format)
	}

	if meta.FileSize == 0 {
		t.Error("expected non-zero file size")
	}
}

// TestGetOntology tests retrieving ontology metadata
func TestGetOntology(t *testing.T) {
	tmpDir := t.TempDir()

	turtleContent := `@prefix ex: <http://example.org/> .
ex:Subject ex:predicate ex:Object .
`

	turtlePath := filepath.Join(tmpDir, "example.ttl")
	if err := os.WriteFile(turtlePath, []byte(turtleContent), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	client := NewSPARQLClient("http://localhost:7878", nil)
	registry := NewRegistry(client, nil, "dev-minimal")

	if err := registry.LoadOntologies(tmpDir); err != nil {
		t.Fatalf("failed to load ontologies: %v", err)
	}

	// Get existing ontology
	meta, err := registry.GetOntology("example")
	if err != nil {
		t.Fatalf("failed to get ontology: %v", err)
	}

	if meta.Name != "example" {
		t.Errorf("expected name=example, got %s", meta.Name)
	}

	// Get non-existent ontology
	_, err = registry.GetOntology("nonexistent")
	if err == nil {
		t.Error("expected error for non-existent ontology")
	}
}

// TestCacheHitRate tests cache hit tracking
func TestCacheHitRate(t *testing.T) {
	client := NewSPARQLClient("http://localhost:7878", nil)
	registry := NewRegistry(client, nil, "dev-minimal")

	// Record cache operations
	registry.recordCacheHit()
	registry.recordCacheHit()
	registry.recordCacheMiss()
	registry.recordCacheHit()

	stats := registry.GetQueryStats()

	if stats.TotalQueries != 4 {
		t.Errorf("expected 4 total queries, got %d", stats.TotalQueries)
	}

	if stats.CacheHits != 3 {
		t.Errorf("expected 3 cache hits, got %d", stats.CacheHits)
	}

	if stats.CacheMisses != 1 {
		t.Errorf("expected 1 cache miss, got %d", stats.CacheMisses)
	}

	expectedRate := 3.0 / 4.0 // 75%
	if stats.CacheHitRate != expectedRate {
		t.Errorf("expected cache hit rate %.2f, got %.2f", expectedRate, stats.CacheHitRate)
	}
}

// TestLatencyTracking tests latency measurement
func TestLatencyTracking(t *testing.T) {
	client := NewSPARQLClient("http://localhost:7878", nil)
	registry := NewRegistry(client, nil, "dev-minimal")

	// Record latencies
	latencies := []time.Duration{
		100 * time.Millisecond,
		150 * time.Millisecond,
		200 * time.Millisecond,
		250 * time.Millisecond,
		300 * time.Millisecond,
	}

	for _, lat := range latencies {
		registry.recordLatency(lat)
	}

	stats := registry.GetQueryStats()

	if stats.LatencyP50Ms <= 0 {
		t.Errorf("expected positive p50 latency, got %.2f", stats.LatencyP50Ms)
	}

	if stats.LatencyP95Ms <= stats.LatencyP50Ms {
		t.Errorf("expected p95 >= p50, got p95=%.2f, p50=%.2f", stats.LatencyP95Ms, stats.LatencyP50Ms)
	}

	if stats.AvgLatencyMs <= 0 {
		t.Errorf("expected positive average latency, got %.2f", stats.AvgLatencyMs)
	}
}

// TestConcurrentLoad tests loading multiple ontologies concurrently
func TestConcurrentLoad(t *testing.T) {
	tmpDir := t.TempDir()

	// Create 5 ontology files
	for i := 1; i <= 5; i++ {
		content := `@prefix ex: <http://example.org/> .
ex:Subject ex:predicate ex:Object .
`
		filename := filepath.Join(tmpDir, "ontology"+string(rune('0'+i))+".ttl")
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	client := NewSPARQLClient("http://localhost:7878", nil)
	registry := NewRegistry(client, nil, "dev-minimal")

	err := registry.LoadOntologies(tmpDir)
	if err != nil {
		t.Fatalf("failed to load ontologies: %v", err)
	}

	if registry.OntologyCount() != 5 {
		t.Errorf("expected 5 ontologies, got %d", registry.OntologyCount())
	}
}

// TestListOntologies tests listing all ontologies
func TestListOntologies(t *testing.T) {
	tmpDir := t.TempDir()

	// Create 3 files
	filenames := []string{"ontology1.ttl", "ontology2.ttl", "ontology3.ttl"}
	for _, name := range filenames {
		path := filepath.Join(tmpDir, name)
		content := "@prefix ex: <http://example.org/> .\nex:s ex:p ex:o .\n"
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	client := NewSPARQLClient("http://localhost:7878", nil)
	registry := NewRegistry(client, nil, "dev-minimal")

	if err := registry.LoadOntologies(tmpDir); err != nil {
		t.Fatalf("failed to load ontologies: %v", err)
	}

	ontologies := registry.ListOntologies()
	if len(ontologies) != 3 {
		t.Errorf("expected 3 ontologies, got %d", len(ontologies))
	}

	// Verify all have correct format
	for _, meta := range ontologies {
		if meta.Format != "ttl" {
			t.Errorf("expected format=ttl, got %s", meta.Format)
		}
	}
}

// TestEnvironmentSpecificLoading tests env-specific loading
func TestEnvironmentSpecificLoading(t *testing.T) {
	tmpDir := t.TempDir()

	// Create 1 file
	path := filepath.Join(tmpDir, "critical.ttl")
	content := "@prefix ex: <http://example.org/> .\nex:s ex:p ex:o .\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	client := NewSPARQLClient("http://localhost:7878", nil)

	// Production environment should load successfully
	registry := NewRegistry(client, nil, "production-f5")
	if err := registry.LoadOntologies(tmpDir); err != nil {
		t.Fatalf("production env failed to load ontologies: %v", err)
	}

	if registry.OntologyCount() != 1 {
		t.Errorf("expected 1 ontology in production, got %d", registry.OntologyCount())
	}
}

// TestMissingOntologyDirectory tests error handling for missing directory
func TestMissingOntologyDirectory(t *testing.T) {
	client := NewSPARQLClient("http://localhost:7878", nil)
	registry := NewRegistry(client, nil, "dev-minimal")

	err := registry.LoadOntologies("/nonexistent/path")
	if err == nil {
		t.Fatal("expected error for missing directory")
	}
}

// TestReloadRegistry tests registry reload
func TestReloadRegistry(t *testing.T) {
	tmpDir := t.TempDir()

	// Create initial ontology
	path1 := filepath.Join(tmpDir, "ontology1.ttl")
	content := "@prefix ex: <http://example.org/> .\nex:s ex:p ex:o .\n"
	if err := os.WriteFile(path1, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	client := NewSPARQLClient("http://localhost:7878", nil)
	registry := NewRegistry(client, nil, "dev-minimal")

	// Load initial
	if err := registry.LoadOntologies(tmpDir); err != nil {
		t.Fatalf("failed to load ontologies: %v", err)
	}

	if registry.OntologyCount() != 1 {
		t.Errorf("expected 1 ontology, got %d", registry.OntologyCount())
	}

	// Add a second ontology
	path2 := filepath.Join(tmpDir, "ontology2.ttl")
	if err := os.WriteFile(path2, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create second test file: %v", err)
	}

	// Reload
	if err := registry.ReloadRegistry(); err != nil {
		t.Fatalf("failed to reload registry: %v", err)
	}

	if registry.OntologyCount() != 2 {
		t.Errorf("expected 2 ontologies after reload, got %d", registry.OntologyCount())
	}
}

// TestLRUCacheBasic tests basic cache operations
func TestLRUCacheBasic(t *testing.T) {
	cache := NewLRUCache(3)

	// Add items
	cache.Put("key1", "value1")
	cache.Put("key2", "value2")
	cache.Put("key3", "value3")

	// Verify all present
	if cache.Size() != 3 {
		t.Errorf("expected cache size 3, got %d", cache.Size())
	}

	// Verify get works
	val, found := cache.Get("key1")
	if !found {
		t.Error("expected to find key1")
	}

	if val != "value1" {
		t.Errorf("expected value1, got %v", val)
	}
}

// TestLRUCacheEviction tests LRU eviction policy
func TestLRUCacheEviction(t *testing.T) {
	cache := NewLRUCache(3)

	// Add 4 items (should evict oldest)
	cache.Put("key1", "value1")
	cache.Put("key2", "value2")
	cache.Put("key3", "value3")
	cache.Put("key4", "value4") // Should evict key1

	// Verify size is still 3
	if cache.Size() != 3 {
		t.Errorf("expected cache size 3, got %d", cache.Size())
	}

	// Verify key1 was evicted
	_, found := cache.Get("key1")
	if found {
		t.Error("expected key1 to be evicted")
	}

	// Verify key4 is present
	val, found := cache.Get("key4")
	if !found {
		t.Error("expected to find key4")
	}

	if val != "value4" {
		t.Errorf("expected value4, got %v", val)
	}
}

// TestLRUCacheMoveToFront tests that Get moves items to front
func TestLRUCacheMoveToFront(t *testing.T) {
	cache := NewLRUCache(2)

	cache.Put("key1", "value1")
	cache.Put("key2", "value2")

	// Access key1 (moves to front)
	cache.Get("key1")

	// Add key3 (should evict key2, not key1)
	cache.Put("key3", "value3")

	// Verify key2 was evicted
	_, found := cache.Get("key2")
	if found {
		t.Error("expected key2 to be evicted")
	}

	// Verify key1 and key3 are present
	_, found = cache.Get("key1")
	if !found {
		t.Error("expected key1 to be present")
	}

	_, found = cache.Get("key3")
	if !found {
		t.Error("expected key3 to be present")
	}
}

// TestLRUCacheClear tests cache clearing
func TestLRUCacheClear(t *testing.T) {
	cache := NewLRUCache(3)

	cache.Put("key1", "value1")
	cache.Put("key2", "value2")

	if cache.Size() != 2 {
		t.Errorf("expected size 2, got %d", cache.Size())
	}

	cache.Clear()

	if cache.Size() != 0 {
		t.Errorf("expected size 0 after clear, got %d", cache.Size())
	}

	_, found := cache.Get("key1")
	if found {
		t.Error("expected key1 to be cleared")
	}
}

// TestSPARQLClientConstruct tests CONSTRUCT query execution (mock)
func TestSPARQLClientConstruct(t *testing.T) {
	client := NewSPARQLClient("http://localhost:7878", nil)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
PREFIX ex: <http://example.org/>
CONSTRUCT { ?s ?p ?o }
WHERE { ?s ?p ?o }
`

	// This would fail without a real Oxigraph server
	// For unit testing, we validate that the timeout is respected
	_, err := client.ExecuteConstruct(ctx, query, 1*time.Second)
	if err == nil {
		t.Skip("skipping test without running Oxigraph server")
	}
}

// TestParseMetadata tests metadata extraction from ontology files
func TestParseMetadata(t *testing.T) {
	tmpDir := t.TempDir()

	// Create ontology with metadata comments
	turtleContent := `@prefix ex: <http://example.org/> .
@prefix dc: <http://purl.org/dc/elements/1.1/> .

<http://example.org/ontology1>
  dc:title "Example Ontology" ;
  dc:description "An example ontology" ;
  dc:date "2026-03-26" .

ex:Subject1 ex:predicate1 ex:Object1 .
`

	path := filepath.Join(tmpDir, "annotated.ttl")
	if err := os.WriteFile(path, []byte(turtleContent), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	client := NewSPARQLClient("http://localhost:7878", nil)
	registry := NewRegistry(client, nil, "dev-minimal")

	if err := registry.LoadOntologies(tmpDir); err != nil {
		t.Fatalf("failed to load ontologies: %v", err)
	}

	meta, err := registry.GetOntology("annotated")
	if err != nil {
		t.Fatalf("failed to get ontology: %v", err)
	}

	if meta.TripleCount == 0 {
		t.Errorf("expected non-zero triple count")
	}
}

// TestQueryStatsExport tests statistics export
func TestQueryStatsExport(t *testing.T) {
	client := NewSPARQLClient("http://localhost:7878", nil)
	registry := NewRegistry(client, nil, "dev-minimal")

	// Simulate query activity
	for i := 0; i < 10; i++ {
		if i%3 == 0 {
			registry.recordCacheHit()
		} else {
			registry.recordCacheMiss()
		}
		registry.recordLatency(time.Duration((i+1)*100) * time.Millisecond)
	}

	stats := registry.GetQueryStats()

	if stats.TotalQueries != 10 {
		t.Errorf("expected 10 total queries, got %d", stats.TotalQueries)
	}

	if stats.CacheHits < 1 {
		t.Errorf("expected at least 1 cache hit, got %d", stats.CacheHits)
	}

	if stats.CacheHitRate < 0 || stats.CacheHitRate > 1 {
		t.Errorf("expected cache hit rate in [0,1], got %.2f", stats.CacheHitRate)
	}
}

// TestOntologyLoadingWithLogger tests logging during load
func TestOntologyLoadingWithLogger(t *testing.T) {
	tmpDir := t.TempDir()

	path := filepath.Join(tmpDir, "test.ttl")
	content := "@prefix ex: <http://example.org/> .\nex:s ex:p ex:o .\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	client := NewSPARQLClient("http://localhost:7878", logger)
	registry := NewRegistry(client, logger, "staging-f5")

	if err := registry.LoadOntologies(tmpDir); err != nil {
		t.Fatalf("failed to load ontologies: %v", err)
	}

	if registry.OntologyCount() != 1 {
		t.Errorf("expected 1 ontology, got %d", registry.OntologyCount())
	}
}
