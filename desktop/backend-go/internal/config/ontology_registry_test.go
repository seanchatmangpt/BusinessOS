package config

import (
	"os"
	"path/filepath"
	"testing"
)

// ---------------------------------------------------------------------------
// Helper: minimal valid YAML for testing
// ---------------------------------------------------------------------------

const minimalValidYAML = `
apiVersion: v1
kind: OntologyRegistry
metadata:
  name: test-registry
  environment: development
spec:
  ontologies:
    - name: core
      iri: http://example.com/ontology/core
      version: "1.0.0"
      required: true
      enabled: true
`

// ---------------------------------------------------------------------------
// ParseOntologyRegistry() — happy path and validation
// ---------------------------------------------------------------------------

func TestParseOntologyRegistry_ValidYAML(t *testing.T) {
	registry, err := ParseOntologyRegistry(minimalValidYAML)
	if err != nil {
		t.Fatalf("ParseOntologyRegistry() error = %v", err)
	}
	if registry.ConfigName != "test-registry" {
		t.Errorf("ConfigName = %q, want %q", registry.ConfigName, "test-registry")
	}
	if registry.Environment != "development" {
		t.Errorf("Environment = %q, want %q", registry.Environment, "development")
	}
}

func TestParseOntologyRegistry_MissingMetadataName(t *testing.T) {
	yaml := `
apiVersion: v1
kind: OntologyRegistry
metadata:
  environment: development
spec:
  ontologies: []
`
	_, err := ParseOntologyRegistry(yaml)
	if err == nil {
		t.Fatal("ParseOntologyRegistry() = nil, want error for missing metadata.name")
	}
}

func TestParseOntologyRegistry_MissingMetadataEnvironment(t *testing.T) {
	yaml := `
apiVersion: v1
kind: OntologyRegistry
metadata:
  name: test
spec:
  ontologies: []
`
	_, err := ParseOntologyRegistry(yaml)
	if err == nil {
		t.Fatal("ParseOntologyRegistry() = nil, want error for missing metadata.environment")
	}
}

func TestParseOntologyRegistry_InvalidYAML(t *testing.T) {
	_, err := ParseOntologyRegistry("this is not yaml {{{")
	if err == nil {
		t.Fatal("ParseOntologyRegistry() = nil, want error for invalid YAML")
	}
}

func TestParseOntologyRegistry_EmptyYAML(t *testing.T) {
	_, err := ParseOntologyRegistry("")
	if err == nil {
		t.Fatal("ParseOntologyRegistry() = nil, want error for empty YAML")
	}
}

// ---------------------------------------------------------------------------
// LoadOntologyRegistry() — file-based loading
// ---------------------------------------------------------------------------

func TestLoadOntologyRegistry_ValidFile(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "ontology.yaml")
	if err := os.WriteFile(tmpFile, []byte(minimalValidYAML), 0644); err != nil {
		t.Fatal(err)
	}
	registry, err := LoadOntologyRegistry(tmpFile)
	if err != nil {
		t.Fatalf("LoadOntologyRegistry() error = %v", err)
	}
	if registry.ConfigName != "test-registry" {
		t.Errorf("ConfigName = %q, want %q", registry.ConfigName, "test-registry")
	}
}

func TestLoadOntologyRegistry_MissingFile(t *testing.T) {
	_, err := LoadOntologyRegistry("/nonexistent/ontology.yaml")
	if err == nil {
		t.Fatal("LoadOntologyRegistry() = nil, want error for missing file")
	}
}

// ---------------------------------------------------------------------------
// Get() — retrieve ontology by name
// ---------------------------------------------------------------------------

func TestGet_FoundReturnsOntology(t *testing.T) {
	registry, err := ParseOntologyRegistry(minimalValidYAML)
	if err != nil {
		t.Fatal(err)
	}
	entry, err := registry.Get("core")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if entry.Name != "core" {
		t.Errorf("Name = %q, want %q", entry.Name, "core")
	}
	if entry.IRI != "http://example.com/ontology/core" {
		t.Errorf("IRI = %q, want %q", entry.IRI, "http://example.com/ontology/core")
	}
}

func TestGet_NotFoundReturnsError(t *testing.T) {
	registry, err := ParseOntologyRegistry(minimalValidYAML)
	if err != nil {
		t.Fatal(err)
	}
	_, err = registry.Get("nonexistent")
	if err == nil {
		t.Fatal("Get() = nil, want error for nonexistent ontology")
	}
}

// ---------------------------------------------------------------------------
// GetByAlias() — retrieve ontology by alias
// ---------------------------------------------------------------------------

func TestGetByAlias_FoundReturnsOntology(t *testing.T) {
	yaml := minimalValidYAML + `
    - name: extended
      iri: http://example.com/ontology/extended
      version: "1.0.0"
      required: false
      enabled: true
      alias: ext
`
	registry, err := ParseOntologyRegistry(yaml)
	if err != nil {
		t.Fatal(err)
	}
	entry, err := registry.GetByAlias("ext")
	if err != nil {
		t.Fatalf("GetByAlias() error = %v", err)
	}
	if entry.Name != "extended" {
		t.Errorf("Name = %q, want %q", entry.Name, "extended")
	}
}

func TestGetByAlias_NotFoundReturnsError(t *testing.T) {
	registry, err := ParseOntologyRegistry(minimalValidYAML)
	if err != nil {
		t.Fatal(err)
	}
	_, err = registry.GetByAlias("no-such-alias")
	if err == nil {
		t.Fatal("GetByAlias() = nil, want error for nonexistent alias")
	}
}

func TestParseOntologyRegistry_DuplicateAliasFails(t *testing.T) {
	yaml := minimalValidYAML + `
    - name: ext1
      iri: http://example.com/ext1
      version: "1.0.0"
      required: false
      enabled: true
      alias: shared
    - name: ext2
      iri: http://example.com/ext2
      version: "1.0.0"
      required: false
      enabled: true
      alias: shared
`
	_, err := ParseOntologyRegistry(yaml)
	if err == nil {
		t.Fatal("ParseOntologyRegistry() = nil, want error for duplicate alias")
	}
}

// ---------------------------------------------------------------------------
// EnabledOntologies() / RequiredOntologies()
// ---------------------------------------------------------------------------

func TestEnabledOntologies_ReturnsOnlyEnabled(t *testing.T) {
	yaml := minimalValidYAML + `
    - name: disabled_one
      iri: http://example.com/disabled
      version: "1.0.0"
      required: false
      enabled: false
    - name: enabled_one
      iri: http://example.com/enabled
      version: "1.0.0"
      required: false
      enabled: true
`
	registry, err := ParseOntologyRegistry(yaml)
	if err != nil {
		t.Fatal(err)
	}
	enabled := registry.EnabledOntologies()
	if len(enabled) != 2 {
		t.Fatalf("EnabledOntologies() count = %d, want 2", len(enabled))
	}
	for _, e := range enabled {
		if !e.Enabled {
			t.Errorf("ontology %q is not enabled but returned by EnabledOntologies()", e.Name)
		}
	}
}

func TestRequiredOntologies_ReturnsOnlyRequired(t *testing.T) {
	yaml := minimalValidYAML + `
    - name: optional_one
      iri: http://example.com/optional
      version: "1.0.0"
      required: false
      enabled: true
`
	registry, err := ParseOntologyRegistry(yaml)
	if err != nil {
		t.Fatal(err)
	}
	required := registry.RequiredOntologies()
	if len(required) != 1 {
		t.Fatalf("RequiredOntologies() count = %d, want 1", len(required))
	}
	if required[0].Name != "core" {
		t.Errorf("RequiredOntologies()[0].Name = %q, want %q", required[0].Name, "core")
	}
}

// ---------------------------------------------------------------------------
// OntologiesByFramework()
// ---------------------------------------------------------------------------

func TestOntologiesByFramework_ReturnsMatching(t *testing.T) {
	yaml := minimalValidYAML + `
    - name: fibo
      iri: http://example.com/fibo
      version: "1.0.0"
      required: false
      enabled: true
      frameworks:
        - fibo
        - commerce
    - name: other
      iri: http://example.com/other
      version: "1.0.0"
      required: false
      enabled: true
      frameworks:
        - healthcare
`
	registry, err := ParseOntologyRegistry(yaml)
	if err != nil {
		t.Fatal(err)
	}
	results := registry.OntologiesByFramework("fibo")
	if len(results) != 1 {
		t.Fatalf("OntologiesByFramework('fibo') count = %d, want 1", len(results))
	}
	if results[0].Name != "fibo" {
		t.Errorf("OntologiesByFramework('fibo')[0].Name = %q, want %q", results[0].Name, "fibo")
	}
}

func TestOntologiesByFramework_NoMatchReturnsEmpty(t *testing.T) {
	registry, err := ParseOntologyRegistry(minimalValidYAML)
	if err != nil {
		t.Fatal(err)
	}
	results := registry.OntologiesByFramework("nonexistent_framework")
	if len(results) != 0 {
		t.Errorf("OntologiesByFramework('nonexistent_framework') count = %d, want 0", len(results))
	}
}

// ---------------------------------------------------------------------------
// GetNamespace() / SPARQLPrefixes()
// ---------------------------------------------------------------------------

func TestGetNamespace_FoundReturnsIRI(t *testing.T) {
	yaml := minimalValidYAML + `
  namespaces:
    rdf: http://www.w3.org/1999/02/22-rdf-syntax-ns#
    rdfs: http://www.w3.org/2000/01/rdf-schema#
`
	registry, err := ParseOntologyRegistry(yaml)
	if err != nil {
		t.Fatal(err)
	}
	iri, ok := registry.GetNamespace("rdf")
	if !ok {
		t.Fatal("GetNamespace('rdf') = false, want true")
	}
	if iri != "http://www.w3.org/1999/02/22-rdf-syntax-ns#" {
		t.Errorf("GetNamespace('rdf') = %q, want correct IRI", iri)
	}
}

func TestGetNamespace_NotFoundReturnsFalse(t *testing.T) {
	registry, err := ParseOntologyRegistry(minimalValidYAML)
	if err != nil {
		t.Fatal(err)
	}
	_, ok := registry.GetNamespace("nonexistent")
	if ok {
		t.Error("GetNamespace('nonexistent') = true, want false")
	}
}

func TestSPARQLPrefixes_GeneratesPrefixLines(t *testing.T) {
	yaml := minimalValidYAML + `
  namespaces:
    rdf: http://www.w3.org/1999/02/22-rdf-syntax-ns#
    rdfs: http://www.w3.org/2000/01/rdf-schema#
`
	registry, err := ParseOntologyRegistry(yaml)
	if err != nil {
		t.Fatal(err)
	}
	prefixes := registry.SPARQLPrefixes()
	if len(prefixes) == 0 {
		t.Fatal("SPARQLPrefixes() returned empty string")
	}
	// Check that both prefixes are present
	if !contains(prefixes, "PREFIX rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#>") {
		t.Error("SPARQLPrefixes() missing rdf prefix")
	}
	if !contains(prefixes, "PREFIX rdfs: <http://www.w3.org/2000/01/rdf-schema#>") {
		t.Error("SPARQLPrefixes() missing rdfs prefix")
	}
}

// ---------------------------------------------------------------------------
// Validate() — required ontology must be enabled
// ---------------------------------------------------------------------------

func TestValidate_RequiredOntologyDisabledFails(t *testing.T) {
	yaml := `
apiVersion: v1
kind: OntologyRegistry
metadata:
  name: test
  environment: development
spec:
  ontologies:
    - name: core
      iri: http://example.com/core
      version: "1.0.0"
      required: true
      enabled: false
`
	_, err := ParseOntologyRegistry(yaml)
	if err == nil {
		t.Fatal("ParseOntologyRegistry() = nil, want error for required but disabled ontology")
	}
}

// ---------------------------------------------------------------------------
// Validate() — circular dependency detection
// ---------------------------------------------------------------------------

func TestValidate_CircularDependencyFails(t *testing.T) {
	yaml := `
apiVersion: v1
kind: OntologyRegistry
metadata:
  name: test
  environment: development
spec:
  ontologies:
    - name: a
      iri: http://example.com/a
      version: "1.0.0"
      required: false
      enabled: true
      dependencies:
        - b
    - name: b
      iri: http://example.com/b
      version: "1.0.0"
      required: false
      enabled: true
      dependencies:
        - a
`
	_, err := ParseOntologyRegistry(yaml)
	if err == nil {
		t.Fatal("ParseOntologyRegistry() = nil, want error for circular dependency")
	}
}

func TestValidate_SelfDependencyFails(t *testing.T) {
	yaml := `
apiVersion: v1
kind: OntologyRegistry
metadata:
  name: test
  environment: development
spec:
  ontologies:
    - name: self_ref
      iri: http://example.com/self
      version: "1.0.0"
      required: false
      enabled: true
      dependencies:
        - self_ref
`
	_, err := ParseOntologyRegistry(yaml)
	if err == nil {
		t.Fatal("ParseOntologyRegistry() = nil, want error for self-dependency")
	}
}

func TestValidate_ThreeWayCircularDependencyFails(t *testing.T) {
	yaml := `
apiVersion: v1
kind: OntologyRegistry
metadata:
  name: test
  environment: development
spec:
  ontologies:
    - name: x
      iri: http://example.com/x
      version: "1.0.0"
      required: false
      enabled: true
      dependencies:
        - y
    - name: y
      iri: http://example.com/y
      version: "1.0.0"
      required: false
      enabled: true
      dependencies:
        - z
    - name: z
      iri: http://example.com/z
      version: "1.0.0"
      required: false
      enabled: true
      dependencies:
        - x
`
	_, err := ParseOntologyRegistry(yaml)
	if err == nil {
		t.Fatal("ParseOntologyRegistry() = nil, want error for three-way circular dependency")
	}
}

// ---------------------------------------------------------------------------
// Validate() — missing dependency detection
// ---------------------------------------------------------------------------

func TestValidate_MissingDependencyFails(t *testing.T) {
	yaml := `
apiVersion: v1
kind: OntologyRegistry
metadata:
  name: test
  environment: development
spec:
  ontologies:
    - name: dependent
      iri: http://example.com/dependent
      version: "1.0.0"
      required: false
      enabled: true
      dependencies:
        - nonexistent_ontology
`
	_, err := ParseOntologyRegistry(yaml)
	if err == nil {
		t.Fatal("ParseOntologyRegistry() = nil, want error for missing dependency")
	}
}

// ---------------------------------------------------------------------------
// Validate() — valid dependency chain passes
// ---------------------------------------------------------------------------

func TestValidate_ValidDependencyChainPasses(t *testing.T) {
	yaml := `
apiVersion: v1
kind: OntologyRegistry
metadata:
  name: test
  environment: development
spec:
  ontologies:
    - name: base
      iri: http://example.com/base
      version: "1.0.0"
      required: false
      enabled: true
    - name: derived
      iri: http://example.com/derived
      version: "1.0.0"
      required: false
      enabled: true
      dependencies:
        - base
    - name: leaf
      iri: http://example.com/leaf
      version: "1.0.0"
      required: false
      enabled: true
      dependencies:
        - derived
`
	registry, err := ParseOntologyRegistry(yaml)
	if err != nil {
		t.Fatalf("ParseOntologyRegistry() error = %v, want nil for valid dependency chain", err)
	}
	if len(registry.ontologies) != 3 {
		t.Errorf("ontology count = %d, want 3", len(registry.ontologies))
	}
}

// ---------------------------------------------------------------------------
// Statistics()
// ---------------------------------------------------------------------------

func TestStatistics_ReturnsCorrectCounts(t *testing.T) {
	yaml := minimalValidYAML + `
    - name: optional
      iri: http://example.com/optional
      version: "1.0.0"
      required: false
      enabled: false
`
	registry, err := ParseOntologyRegistry(yaml)
	if err != nil {
		t.Fatal(err)
	}
	stats := registry.Statistics()
	if stats.TotalOntologies != 2 {
		t.Errorf("TotalOntologies = %d, want 2", stats.TotalOntologies)
	}
	if stats.EnabledCount != 1 {
		t.Errorf("EnabledCount = %d, want 1", stats.EnabledCount)
	}
	if stats.RequiredCount != 1 {
		t.Errorf("RequiredCount = %d, want 1", stats.RequiredCount)
	}
}

// ---------------------------------------------------------------------------
// Frameworks configuration
// ---------------------------------------------------------------------------

func TestFrameworks_ParsedFromYAML(t *testing.T) {
	yaml := minimalValidYAML + `
  frameworks:
    compliance:
      soc2: true
      hipaa: false
      gdpr: true
      sox: false
    domains:
      commerce: true
      operations: false
    extended:
      fibo: true
`
	registry, err := ParseOntologyRegistry(yaml)
	if err != nil {
		t.Fatal(err)
	}
	stats := registry.Statistics()
	if !stats.FrameworksEnabled.Compliance.SOC2 {
		t.Error("frameworks.compliance.soc2 = false, want true")
	}
	if stats.FrameworksEnabled.Compliance.HIPAA {
		t.Error("frameworks.compliance.hipaa = true, want false")
	}
	if !stats.FrameworksEnabled.Compliance.GDPR {
		t.Error("frameworks.compliance.gdpr = false, want true")
	}
	if !stats.FrameworksEnabled.Domains.Commerce {
		t.Error("frameworks.domains.commerce = false, want true")
	}
	if stats.FrameworksEnabled.Domains.Operations {
		t.Error("frameworks.domains.operations = true, want false")
	}
	if !stats.FrameworksEnabled.Extended.FIBO {
		t.Error("frameworks.extended.fibo = false, want true")
	}
}

// ---------------------------------------------------------------------------
// OntologyEntry optional fields
// ---------------------------------------------------------------------------

func TestOntologyEntry_OptionalFields(t *testing.T) {
	yaml := `
apiVersion: v1
kind: OntologyRegistry
metadata:
  name: test
  environment: development
spec:
  ontologies:
    - name: core
      iri: http://example.com/core
      version: "1.0.0"
      required: false
      enabled: true
    - name: full_entry
      iri: http://example.com/full
      version: "2.0.0"
      required: false
      enabled: true
      alias: full
      frameworks:
        - compliance
        - analytics
      dependencies:
        - core
      description: "A full ontology entry"
      validation:
        entity_count_min: 10
        entity_count_max: 100
        relationship_count_min: 5
`
	registry, err := ParseOntologyRegistry(yaml)
	if err != nil {
		t.Fatal(err)
	}
	entry, err := registry.Get("full_entry")
	if err != nil {
		t.Fatal(err)
	}
	if entry.Version != "2.0.0" {
		t.Errorf("Version = %q, want %q", entry.Version, "2.0.0")
	}
	if entry.Alias == nil || *entry.Alias != "full" {
		t.Error("Alias not set or incorrect")
	}
	if len(entry.Frameworks) != 2 {
		t.Errorf("Frameworks count = %d, want 2", len(entry.Frameworks))
	}
	if len(entry.Dependencies) != 1 || entry.Dependencies[0] != "core" {
		t.Errorf("Dependencies = %v, want [core]", entry.Dependencies)
	}
	if entry.Description == nil || *entry.Description != "A full ontology entry" {
		t.Error("Description not set or incorrect")
	}
	if entry.Validation == nil {
		t.Fatal("Validation is nil")
	}
	if entry.Validation.EntityCountMin == nil || *entry.Validation.EntityCountMin != 10 {
		t.Error("EntityCountMin not set or incorrect")
	}
	if entry.Validation.EntityCountMax == nil || *entry.Validation.EntityCountMax != 100 {
		t.Error("EntityCountMax not set or incorrect")
	}
}

// ---------------------------------------------------------------------------
// Helper
// ---------------------------------------------------------------------------

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
