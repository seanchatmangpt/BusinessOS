// Package config provides ontology registry loading and validation
package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

// OntologyEntry represents a single ontology in the registry
type OntologyEntry struct {
	Name         string   `yaml:"name"`
	IRI          string   `yaml:"iri"`
	Version      string   `yaml:"version"`
	Required     bool     `yaml:"required"`
	Enabled      bool     `yaml:"enabled"`
	Alias        *string  `yaml:"alias,omitempty"`
	Frameworks   []string `yaml:"frameworks,omitempty"`
	Dependencies []string `yaml:"dependencies,omitempty"`
	Validation   *struct {
		EntityCountMin       *int `yaml:"entity_count_min,omitempty"`
		EntityCountMax       *int `yaml:"entity_count_max,omitempty"`
		RelationshipCountMin *int `yaml:"relationship_count_min,omitempty"`
	} `yaml:"validation,omitempty"`
	Description *string `yaml:"description,omitempty"`
}

// Frameworks represents framework support configuration
type Frameworks struct {
	Compliance struct {
		SOC2  bool `yaml:"soc2"`
		HIPAA bool `yaml:"hipaa"`
		GDPR  bool `yaml:"gdpr"`
		SOX   bool `yaml:"sox"`
	} `yaml:"compliance"`
	Domains struct {
		Commerce   bool `yaml:"commerce"`
		Operations bool `yaml:"operations"`
		People     bool `yaml:"people"`
		Finance    bool `yaml:"finance"`
		Analytics  bool `yaml:"analytics"`
	} `yaml:"domains"`
	Extended struct {
		FIBO       bool `yaml:"fibo"`
		Healthcare bool `yaml:"healthcare"`
		Research   bool `yaml:"research"`
	} `yaml:"extended"`
}

// ValidationConfig represents validation settings
type ValidationConfig struct {
	Strict               string `yaml:"strict"`
	OnMissingRequired    string `yaml:"on_missing_required"`
	OnVersionMismatch    string `yaml:"on_version_mismatch"`
	OnCircularDependency string `yaml:"on_circular_dependency"`
}

// OntologyRegistry holds loaded ontology configuration
type OntologyRegistry struct {
	mu          sync.RWMutex
	ConfigName  string
	Environment string

	// Ontology lookup by name
	ontologies map[string]*OntologyEntry

	// Alias to name mapping
	aliasMap map[string]string

	// SPARQL namespace declarations
	namespaces map[string]string

	// Framework support
	frameworks Frameworks

	// Validation rules
	validation ValidationConfig
}

// OntologyRegistryConfig represents the full YAML structure
type OntologyRegistryConfig struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name        string `yaml:"name"`
		Environment string `yaml:"environment"`
		Version     string `yaml:"version,omitempty"`
		Description string `yaml:"description,omitempty"`
		CreatedAt   string `yaml:"created_at,omitempty"`
		UpdatedAt   string `yaml:"updated_at,omitempty"`
	} `yaml:"metadata"`
	Spec struct {
		Ontologies []OntologyEntry   `yaml:"ontologies"`
		Namespaces map[string]string `yaml:"namespaces,omitempty"`
		Frameworks Frameworks        `yaml:"frameworks,omitempty"`
		Validation ValidationConfig  `yaml:"validation,omitempty"`
	} `yaml:"spec"`
}

// LoadOntologyRegistry loads ontology config from YAML file
func LoadOntologyRegistry(filePath string) (*OntologyRegistry, error) {
	data, err := os.ReadFile(filepath.Clean(filePath))
	if err != nil {
		return nil, fmt.Errorf("failed to read ontology config file: %w", err)
	}

	return ParseOntologyRegistry(string(data))
}

// ParseOntologyRegistry parses ontology config from YAML string
func ParseOntologyRegistry(yaml_content string) (*OntologyRegistry, error) {
	var cfg OntologyRegistryConfig

	if err := yaml.Unmarshal([]byte(yaml_content), &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse ontology config: %w", err)
	}

	// Validate required fields
	if cfg.Metadata.Name == "" {
		return nil, fmt.Errorf("ontology config missing metadata.name")
	}
	if cfg.Metadata.Environment == "" {
		return nil, fmt.Errorf("ontology config missing metadata.environment")
	}

	registry := &OntologyRegistry{
		ConfigName:  cfg.Metadata.Name,
		Environment: cfg.Metadata.Environment,
		ontologies:  make(map[string]*OntologyEntry),
		aliasMap:    make(map[string]string),
		namespaces:  cfg.Spec.Namespaces,
		frameworks:  cfg.Spec.Frameworks,
		validation:  cfg.Spec.Validation,
	}

	// Build ontology registry and alias map
	for i, ontology := range cfg.Spec.Ontologies {
		// Check for duplicate aliases
		if ontology.Alias != nil {
			if existing, exists := registry.aliasMap[*ontology.Alias]; exists {
				return nil, fmt.Errorf("duplicate alias '%s': used by both '%s' and '%s'",
					*ontology.Alias, existing, ontology.Name)
			}
			registry.aliasMap[*ontology.Alias] = ontology.Name
		}
		registry.ontologies[ontology.Name] = &cfg.Spec.Ontologies[i]
	}

	// Validate
	if err := registry.Validate(); err != nil {
		return nil, err
	}

	slog.Info("Loaded ontology config",
		"config", registry.ConfigName,
		"environment", registry.Environment,
		"ontologies", len(registry.ontologies))

	return registry, nil
}

// Get retrieves an ontology by name
func (r *OntologyRegistry) Get(name string) (*OntologyEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ontology, exists := r.ontologies[name]
	if !exists {
		return nil, fmt.Errorf("ontology not found: %s", name)
	}
	return ontology, nil
}

// GetByAlias retrieves an ontology by alias
func (r *OntologyRegistry) GetByAlias(alias string) (*OntologyEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	name, exists := r.aliasMap[alias]
	if !exists {
		return nil, fmt.Errorf("alias not found: %s", alias)
	}

	ontology, _ := r.ontologies[name]
	return ontology, nil
}

// EnabledOntologies returns all enabled ontologies
func (r *OntologyRegistry) EnabledOntologies() []*OntologyEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*OntologyEntry
	for _, ontology := range r.ontologies {
		if ontology.Enabled {
			result = append(result, ontology)
		}
	}
	return result
}

// RequiredOntologies returns all required ontologies
func (r *OntologyRegistry) RequiredOntologies() []*OntologyEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*OntologyEntry
	for _, ontology := range r.ontologies {
		if ontology.Required {
			result = append(result, ontology)
		}
	}
	return result
}

// OntologiesByFramework returns ontologies for a specific framework
func (r *OntologyRegistry) OntologiesByFramework(framework string) []*OntologyEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*OntologyEntry
	for _, ontology := range r.ontologies {
		for _, fw := range ontology.Frameworks {
			if fw == framework {
				result = append(result, ontology)
				break
			}
		}
	}
	return result
}

// GetNamespace retrieves a namespace IRI by prefix
func (r *OntologyRegistry) GetNamespace(prefix string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	iri, exists := r.namespaces[prefix]
	return iri, exists
}

// SPARQLPrefixes generates SPARQL PREFIX lines
func (r *OntologyRegistry) SPARQLPrefixes() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var prefixes string
	for prefix, iri := range r.namespaces {
		prefixes += fmt.Sprintf("PREFIX %s: <%s>\n", prefix, iri)
	}
	return prefixes
}

// Validate performs validation checks on the registry
func (r *OntologyRegistry) Validate() error {
	// Check all required ontologies are enabled
	for _, ontology := range r.RequiredOntologies() {
		if !ontology.Enabled {
			return fmt.Errorf("required ontology '%s' is not enabled", ontology.Name)
		}
	}

	// Check for circular dependencies
	if err := r.validateNoCircularDeps(); err != nil {
		return err
	}

	// Check all dependencies exist
	for _, ontology := range r.ontologies {
		for _, dep := range ontology.Dependencies {
			if _, exists := r.ontologies[dep]; !exists {
				return fmt.Errorf("ontology '%s' depends on unknown ontology '%s'",
					ontology.Name, dep)
			}
		}
	}

	slog.Info("Ontology configuration validation passed")
	return nil
}

// validateNoCircularDeps checks for circular dependencies using DFS
func (r *OntologyRegistry) validateNoCircularDeps() error {
	visited := make(map[string]bool)
	for name := range r.ontologies {
		if err := r.checkCycle(name, visited, make(map[string]bool)); err != nil {
			return err
		}
	}
	return nil
}

// checkCycle recursively checks for cycles in dependency graph
func (r *OntologyRegistry) checkCycle(
	name string,
	visited map[string]bool,
	recStack map[string]bool,
) error {
	visited[name] = true
	recStack[name] = true

	ontology, exists := r.ontologies[name]
	if !exists {
		return nil
	}

	for _, dep := range ontology.Dependencies {
		if !visited[dep] {
			if err := r.checkCycle(dep, visited, recStack); err != nil {
				return err
			}
		} else if recStack[dep] {
			return fmt.Errorf("circular dependency detected in ontology '%s'", name)
		}
	}

	recStack[name] = false
	return nil
}

// Statistics returns registry statistics
type OntologyStats struct {
	TotalOntologies   int        `json:"total_ontologies"`
	EnabledCount      int        `json:"enabled_count"`
	RequiredCount     int        `json:"required_count"`
	FrameworksEnabled Frameworks `json:"frameworks_enabled"`
}

// Statistics returns registry statistics
func (r *OntologyRegistry) Statistics() OntologyStats {
	return OntologyStats{
		TotalOntologies:   len(r.ontologies),
		EnabledCount:      len(r.EnabledOntologies()),
		RequiredCount:     len(r.RequiredOntologies()),
		FrameworksEnabled: r.frameworks,
	}
}
