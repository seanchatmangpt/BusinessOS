/// Ontology Registry Loader
///
/// Loads, validates, and provides access to ontology configurations.
/// Supports hot-reload and environment-specific overrides.
///
/// Usage:
/// ```rust
/// let registry = OntologyRegistry::load("configs/production-f5.yaml")?;
/// let sales = registry.get("sales_ontology")?;
/// println!("Sales ontology: {}", sales.iri);
/// ```

use anyhow::{anyhow, Result};
use serde::{Deserialize, Serialize};
use std::collections::{HashMap, HashSet};
use std::fs;
use std::path::Path;
use tracing::{info, warn};

/// Ontology metadata and configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(crate = "serde")]
pub struct OntologyEntry {
    /// Unique identifier (e.g., "sales_ontology")
    pub name: String,

    /// Formal ontology IRI/namespace
    pub iri: String,

    /// Semantic version (MAJOR.MINOR.PATCH)
    pub version: String,

    /// Is this ontology required?
    pub required: bool,

    /// Is this ontology enabled?
    pub enabled: bool,

    /// Namespace alias for SPARQL queries
    pub alias: Option<String>,

    /// Associated compliance/domain frameworks
    #[serde(default)]
    pub frameworks: Vec<String>,

    /// Dependencies on other ontologies
    #[serde(default)]
    pub dependencies: Vec<String>,

    /// Validation constraints
    #[serde(default)]
    pub validation: Option<ValidationConstraints>,

    /// Human-readable description
    #[serde(default)]
    pub description: Option<String>,
}

/// Validation constraints for an ontology
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(crate = "serde")]
pub struct ValidationConstraints {
    pub entity_count_min: Option<usize>,
    pub entity_count_max: Option<usize>,
    pub relationship_count_min: Option<usize>,
}

/// Framework support configuration
#[derive(Debug, Clone, Serialize, Deserialize, Default)]
#[serde(crate = "serde")]
pub struct Frameworks {
    #[serde(default)]
    pub compliance: ComplianceFrameworks,

    #[serde(default)]
    pub domains: DomainFrameworks,

    #[serde(default)]
    pub extended: ExtendedFrameworks,
}

/// Compliance framework flags
#[derive(Debug, Clone, Serialize, Deserialize, Default)]
#[serde(crate = "serde")]
pub struct ComplianceFrameworks {
    pub soc2: bool,
    pub hipaa: bool,
    pub gdpr: bool,
    pub sox: bool,
}

/// Data mesh domain flags
#[derive(Debug, Clone, Serialize, Deserialize, Default)]
#[serde(crate = "serde")]
pub struct DomainFrameworks {
    pub commerce: bool,
    pub operations: bool,
    pub people: bool,
    pub finance: bool,
    pub analytics: bool,
}

/// Extended domain flags
#[derive(Debug, Clone, Serialize, Deserialize, Default)]
#[serde(crate = "serde")]
pub struct ExtendedFrameworks {
    pub fibo: bool,
    pub healthcare: bool,
    pub research: bool,
}

/// Validation configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(crate = "serde")]
pub struct ValidationConfig {
    pub strict: bool,
    pub on_missing_required: String,
    pub on_version_mismatch: String,
    pub on_circular_dependency: String,
}

impl Default for ValidationConfig {
    fn default() -> Self {
        Self {
            strict: true,
            on_missing_required: "error".to_string(),
            on_version_mismatch: "warn".to_string(),
            on_circular_dependency: "error".to_string(),
        }
    }
}

/// Ontology Registry
///
/// Holds loaded ontology configuration, provides lookup and validation.
#[derive(Debug, Clone)]
pub struct OntologyRegistry {
    /// Configuration metadata
    pub config_name: String,
    pub environment: String,

    /// Ontology lookup by name
    ontologies: HashMap<String, OntologyEntry>,

    /// Alias to name mapping (e.g., "sales" -> "sales_ontology")
    alias_map: HashMap<String, String>,

    /// SPARQL namespace declarations
    namespaces: HashMap<String, String>,

    /// Framework support
    frameworks: Frameworks,

    /// Validation rules
    validation: ValidationConfig,
}

impl OntologyRegistry {
    /// Load ontology configuration from YAML file
    pub fn load<P: AsRef<Path>>(path: P) -> Result<Self> {
        let path = path.as_ref();
        let contents = fs::read_to_string(path)?;
        Self::from_yaml(&contents)
    }

    /// Parse ontology configuration from YAML string
    pub fn from_yaml(yaml: &str) -> Result<Self> {
        #[derive(Deserialize)]
        struct Config {
            #[serde(default)]
            metadata: ConfigMetadata,

            #[serde(default)]
            spec: ConfigSpec,
        }

        #[derive(Deserialize, Default)]
        struct ConfigMetadata {
            name: Option<String>,
            environment: Option<String>,
        }

        #[derive(Deserialize, Default)]
        struct ConfigSpec {
            ontologies: Vec<OntologyEntry>,

            #[serde(default)]
            namespaces: HashMap<String, String>,

            #[serde(default)]
            frameworks: Frameworks,

            #[serde(default)]
            validation: ValidationConfig,
        }

        let config: Config = serde_yaml::from_str(yaml)?;

        let config_name = config
            .metadata
            .name
            .ok_or_else(|| anyhow!("Missing metadata.name"))?;
        let environment = config
            .metadata
            .environment
            .ok_or_else(|| anyhow!("Missing metadata.environment"))?;

        let mut registry = Self {
            config_name: config_name.clone(),
            environment: environment.clone(),
            ontologies: HashMap::new(),
            alias_map: HashMap::new(),
            namespaces: config.spec.namespaces,
            frameworks: config.spec.frameworks,
            validation: config.spec.validation,
        };

        // Build ontology registry and alias map
        for ontology in config.spec.ontologies {
            if let Some(alias) = &ontology.alias {
                // Check for duplicate aliases
                if let Some(existing) = registry.alias_map.get(alias) {
                    return Err(anyhow!(
                        "Duplicate alias '{}': used by both '{}' and '{}'",
                        alias,
                        existing,
                        ontology.name
                    ));
                }
                registry.alias_map.insert(alias.clone(), ontology.name.clone());
            }
            registry.ontologies.insert(ontology.name.clone(), ontology);
        }

        // Validate
        registry.validate()?;

        info!(
            "Loaded ontology config '{}' for environment '{}' with {} ontologies",
            config_name,
            environment,
            registry.ontologies.len()
        );

        Ok(registry)
    }

    /// Get ontology by name
    pub fn get(&self, name: &str) -> Result<&OntologyEntry> {
        self.ontologies
            .get(name)
            .ok_or_else(|| anyhow!("Ontology not found: {}", name))
    }

    /// Get ontology by alias
    pub fn get_by_alias(&self, alias: &str) -> Result<&OntologyEntry> {
        let name = self
            .alias_map
            .get(alias)
            .ok_or_else(|| anyhow!("Alias not found: {}", alias))?;
        self.get(name)
    }

    /// Get all enabled ontologies
    pub fn enabled_ontologies(&self) -> Vec<&OntologyEntry> {
        self.ontologies
            .values()
            .filter(|o| o.enabled)
            .collect::<Vec<_>>()
    }

    /// Get all required ontologies
    pub fn required_ontologies(&self) -> Vec<&OntologyEntry> {
        self.ontologies
            .values()
            .filter(|o| o.required)
            .collect::<Vec<_>>()
    }

    /// Get ontologies for a specific framework
    pub fn ontologies_by_framework(&self, framework: &str) -> Vec<&OntologyEntry> {
        self.ontologies
            .values()
            .filter(|o| o.frameworks.contains(&framework.to_string()))
            .collect()
    }

    /// Get SPARQL namespace declarations
    pub fn get_namespace(&self, prefix: &str) -> Option<&str> {
        self.namespaces.get(prefix).map(|s| s.as_str())
    }

    /// Get all namespace declarations as SPARQL PREFIX lines
    pub fn sparql_prefixes(&self) -> String {
        let mut prefixes = Vec::new();
        for (prefix, iri) in &self.namespaces {
            prefixes.push(format!("PREFIX {}: <{}>", prefix, iri));
        }
        prefixes.join("\n")
    }

    /// Validate ontology configuration
    fn validate(&self) -> Result<()> {
        // Check all required ontologies are enabled
        for ontology in self.required_ontologies() {
            if !ontology.enabled {
                return Err(anyhow!(
                    "Required ontology '{}' is not enabled",
                    ontology.name
                ));
            }
        }

        // Check for circular dependencies
        self.validate_no_circular_deps()?;

        // Check all dependencies exist
        for ontology in self.ontologies.values() {
            for dep in &ontology.dependencies {
                if !self.ontologies.contains_key(dep) {
                    return Err(anyhow!(
                        "Ontology '{}' depends on unknown ontology '{}'",
                        ontology.name,
                        dep
                    ));
                }
            }
        }

        info!("Ontology configuration validation passed");
        Ok(())
    }

    /// Verify no circular dependencies exist
    fn validate_no_circular_deps(&self) -> Result<()> {
        for ontology in self.ontologies.values() {
            let mut visited = HashSet::new();
            self.check_cycle(&ontology.name, &mut visited)?;
        }
        Ok(())
    }

    /// Recursively check for cycles
    fn check_cycle(&self, name: &str, visited: &mut HashSet<String>) -> Result<()> {
        if visited.contains(name) {
            return Err(anyhow!("Circular dependency detected in ontology '{}'", name));
        }

        visited.insert(name.to_string());

        if let Some(ontology) = self.ontologies.get(name) {
            for dep in &ontology.dependencies {
                self.check_cycle(dep, visited)?;
            }
        }

        visited.remove(name);
        Ok(())
    }

    /// Get ontology statistics
    pub fn statistics(&self) -> OntologyStats {
        OntologyStats {
            total_ontologies: self.ontologies.len(),
            enabled_count: self.enabled_ontologies().len(),
            required_count: self.required_ontologies().len(),
            frameworks_enabled: self.frameworks.clone(),
        }
    }
}

/// Ontology configuration statistics
#[derive(Debug, Serialize)]
pub struct OntologyStats {
    pub total_ontologies: usize,
    pub enabled_count: usize,
    pub required_count: usize,
    pub frameworks_enabled: Frameworks,
}

#[cfg(test)]
mod tests {
    use super::*;

    const MINIMAL_CONFIG: &str = r#"
apiVersion: "ontology.chatmangpt.com/v1"
kind: "OntologyConfig"
metadata:
  name: "test-minimal"
  environment: "development"
  version: "1.0.0"
spec:
  ontologies:
    - name: "test_ontology"
      iri: "https://ontology.test/v1"
      version: "1.0.0"
      required: true
      enabled: true
      alias: "test"
      frameworks: [test]
      description: "Test ontology"
  namespaces:
    test: "https://ontology.test/v1"
  validation:
    strict: false
    on_missing_required: "error"
    on_version_mismatch: "warn"
    on_circular_dependency: "error"
"#;

    #[test]
    fn test_load_minimal_config() {
        let registry = OntologyRegistry::from_yaml(MINIMAL_CONFIG).unwrap();
        assert_eq!(registry.config_name, "test-minimal");
        assert_eq!(registry.environment, "development");
        assert_eq!(registry.ontologies.len(), 1);
    }

    #[test]
    fn test_get_ontology() {
        let registry = OntologyRegistry::from_yaml(MINIMAL_CONFIG).unwrap();
        let ontology = registry.get("test_ontology").unwrap();
        assert_eq!(ontology.name, "test_ontology");
        assert_eq!(ontology.iri, "https://ontology.test/v1");
    }

    #[test]
    fn test_get_by_alias() {
        let registry = OntologyRegistry::from_yaml(MINIMAL_CONFIG).unwrap();
        let ontology = registry.get_by_alias("test").unwrap();
        assert_eq!(ontology.name, "test_ontology");
    }

    #[test]
    fn test_enabled_ontologies() {
        let registry = OntologyRegistry::from_yaml(MINIMAL_CONFIG).unwrap();
        assert_eq!(registry.enabled_ontologies().len(), 1);
    }

    #[test]
    fn test_required_ontologies() {
        let registry = OntologyRegistry::from_yaml(MINIMAL_CONFIG).unwrap();
        assert_eq!(registry.required_ontologies().len(), 1);
    }

    #[test]
    fn test_sparql_prefixes() {
        let registry = OntologyRegistry::from_yaml(MINIMAL_CONFIG).unwrap();
        let prefixes = registry.sparql_prefixes();
        assert!(prefixes.contains("PREFIX test:"));
        assert!(prefixes.contains("https://ontology.test/v1"));
    }

    #[test]
    fn test_circular_dependency_detection() {
        let circular_config = r#"
apiVersion: "ontology.chatmangpt.com/v1"
kind: "OntologyConfig"
metadata:
  name: "test-circular"
  environment: "development"
spec:
  ontologies:
    - name: "onto_a"
      iri: "https://test.com/a"
      version: "1.0.0"
      required: true
      enabled: true
      dependencies: ["onto_b"]
    - name: "onto_b"
      iri: "https://test.com/b"
      version: "1.0.0"
      required: true
      enabled: true
      dependencies: ["onto_a"]
"#;
        let result = OntologyRegistry::from_yaml(circular_config);
        assert!(result.is_err());
        assert!(result.unwrap_err().to_string().contains("Circular"));
    }

    #[test]
    fn test_statistics() {
        let registry = OntologyRegistry::from_yaml(MINIMAL_CONFIG).unwrap();
        let stats = registry.statistics();
        assert_eq!(stats.total_ontologies, 1);
        assert_eq!(stats.enabled_count, 1);
        assert_eq!(stats.required_count, 1);
    }
}
