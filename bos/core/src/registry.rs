/// SPARQL CONSTRUCT Query Registry
///
/// Registry system for discovering, loading, parsing, and executing SPARQL CONSTRUCT queries.
/// Queries are organized in ontologies/sparql/constructs/ with metadata in comments.
///
/// Usage:
///   let registry = QueryRegistry::load("ontologies/sparql/constructs")?;
///   let query = registry.lookup("create_artifact")?;
///   let bound = query.bind(vec![
///     ("artifactId", "art-123"),
///     ("title", "API Spec"),
///   ])?;
///   let result = bound.construct_triples(&mut executor)?;

use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::fs;
use std::path::{Path, PathBuf};
use thiserror::Error;

#[derive(Error, Debug)]
pub enum RegistryError {
    #[error("Query not found: {0}")]
    QueryNotFound(String),

    #[error("IO error: {0}")]
    IoError(#[from] std::io::Error),

    #[error("Invalid metadata: {0}")]
    InvalidMetadata(String),

    #[error("Binding error: {0}")]
    BindingError(String),

    #[error("Query execution error: {0}")]
    ExecutionError(String),

    #[error("Category not found: {0}")]
    CategoryNotFound(String),
}

pub type Result<T> = std::result::Result<T, RegistryError>;

/// Query metadata extracted from .rq file comments
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
pub struct QueryMetadata {
    /// Query name for lookup
    pub name: String,

    /// Category: artifacts, projects, compliance, org, process, signal, agents
    pub category: String,

    /// Human-readable description
    pub description: String,

    /// Parameter names that must be bound before execution: ?name1, ?name2, ?name3
    pub params: Vec<String>,

    /// Description of returned RDF triples
    pub returns: String,

    /// Query version
    pub version: String,

    /// Stability: stable, beta, experimental
    pub stability: String,

    /// Whether this query requires parameters to be bound
    pub requires_parameters: bool,
}

/// A SPARQL CONSTRUCT query with metadata
#[derive(Debug, Clone)]
pub struct ConstructQuery {
    /// Query metadata
    pub metadata: QueryMetadata,

    /// Raw SPARQL query text
    pub sparql: String,

    /// File path where query was loaded from
    pub path: PathBuf,
}

impl ConstructQuery {
    /// Extract metadata from query comment block
    fn parse_metadata(sparql: &str, name: &str) -> Result<QueryMetadata> {
        let mut metadata = QueryMetadata {
            name: name.to_string(),
            category: String::new(),
            description: String::new(),
            params: Vec::new(),
            returns: String::new(),
            version: "1.0".to_string(),
            stability: "stable".to_string(),
            requires_parameters: false,
        };

        // Parse comment block at start of file
        for line in sparql.lines().take(50) {
            if !line.trim_start().starts_with('#') {
                break;
            }

            let trimmed = line.trim_start().trim_start_matches('#').trim();

            if trimmed.starts_with("@name:") {
                metadata.name = trimmed.strip_prefix("@name:").unwrap().trim().to_string();
            } else if trimmed.starts_with("@category:") {
                metadata.category = trimmed.strip_prefix("@category:").unwrap().trim().to_string();
            } else if trimmed.starts_with("@description:") {
                metadata.description = trimmed.strip_prefix("@description:").unwrap().trim().to_string();
            } else if trimmed.starts_with("@params:") {
                let params_str = trimmed.strip_prefix("@params:").unwrap().trim();
                metadata.params = params_str
                    .split(',')
                    .map(|p| p.trim().to_string())
                    .collect();
            } else if trimmed.starts_with("@returns:") {
                metadata.returns = trimmed.strip_prefix("@returns:").unwrap().trim().to_string();
            } else if trimmed.starts_with("@version:") {
                metadata.version = trimmed.strip_prefix("@version:").unwrap().trim().to_string();
            } else if trimmed.starts_with("@stability:") {
                metadata.stability = trimmed.strip_prefix("@stability:").unwrap().trim().to_string();
            } else if trimmed.starts_with("@requires_parameters:") {
                let val = trimmed
                    .strip_prefix("@requires_parameters:")
                    .unwrap()
                    .trim()
                    .to_lowercase();
                metadata.requires_parameters = val == "true";
            }
        }

        // Validate required fields
        if metadata.name.is_empty() {
            return Err(RegistryError::InvalidMetadata(
                format!("Missing @name in query: {}", name),
            ));
        }

        if metadata.category.is_empty() {
            return Err(RegistryError::InvalidMetadata(
                format!("Missing @category in query: {}", name),
            ));
        }

        Ok(metadata)
    }

    /// Create a query from file content
    fn from_file(path: PathBuf, sparql: String) -> Result<Self> {
        let name = path
            .file_stem()
            .and_then(|s| s.to_str())
            .ok_or_else(|| {
                RegistryError::InvalidMetadata(format!(
                    "Invalid file name: {}",
                    path.display()
                ))
            })?
            .to_string();

        let metadata = Self::parse_metadata(&sparql, &name)?;

        Ok(ConstructQuery {
            metadata,
            sparql,
            path,
        })
    }

    /// Bind parameter values to the query
    ///
    /// Parameters are ?name format. The bound query will have BIND statements injected.
    pub fn bind(&self, params: HashMap<String, String>) -> Result<BoundQuery> {
        // Validate that all required parameters are provided
        for param in &self.metadata.params {
            let param_name = param.trim_start_matches('?');
            if !params.contains_key(param_name) && self.metadata.requires_parameters {
                return Err(RegistryError::BindingError(format!(
                    "Missing required parameter: {}",
                    param_name
                )));
            }
        }

        Ok(BoundQuery {
            query: self.clone(),
            bindings: params,
        })
    }
}

/// A query with bound parameters, ready to execute
#[derive(Debug)]
pub struct BoundQuery {
    /// Original query
    query: ConstructQuery,

    /// Bound parameter values: key is without the ?
    bindings: HashMap<String, String>,
}

impl BoundQuery {
    /// Get the SPARQL query with BIND statements injected
    pub fn sparql_with_bindings(&self) -> String {
        let mut result = self.query.sparql.clone();

        // Inject BIND statements for each parameter at the start of WHERE clause
        let where_pos = result.find("WHERE {");
        if let Some(pos) = where_pos {
            let insert_pos = pos + 7; // "WHERE {".len()

            let mut bindings_str = String::new();
            for (key, value) in &self.bindings {
                // Escape string values
                let escaped = if value.starts_with('?') {
                    // It's a variable reference
                    value.clone()
                } else if value.starts_with("http://") || value.starts_with("https://") {
                    // It's an IRI
                    format!("<{}>", value)
                } else {
                    // It's a string literal
                    format!("\"{}\"", value.replace('"', "\\\""))
                };

                bindings_str.push_str(&format!("\n  BIND({} as ?{})", escaped, key));
            }

            result.insert_str(insert_pos, &bindings_str);
        }

        result
    }

    /// Get the original unbound query
    pub fn query(&self) -> &ConstructQuery {
        &self.query
    }

    /// Get the bindings
    pub fn bindings(&self) -> &HashMap<String, String> {
        &self.bindings
    }
}

/// Registry of all available CONSTRUCT queries
pub struct QueryRegistry {
    /// Queries indexed by name
    queries_by_name: HashMap<String, ConstructQuery>,

    /// Queries grouped by category
    queries_by_category: HashMap<String, Vec<ConstructQuery>>,

    /// Root directory where queries were loaded from
    root_dir: PathBuf,
}

impl QueryRegistry {
    /// Load all .rq files from a directory structure
    ///
    /// Expected structure:
    /// ```
    /// ontologies/sparql/constructs/
    /// ├── artifacts/
    /// │   ├── create_artifact.rq
    /// │   └── link_artifact_to_decision.rq
    /// ├── projects/
    /// ├── compliance/
    /// ├── org/
    /// ├── process/
    /// ├── signal/
    /// └── agents/
    /// ```
    pub fn load<P: AsRef<Path>>(root: P) -> Result<Self> {
        let root_path = root.as_ref().to_path_buf();

        if !root_path.exists() {
            return Err(RegistryError::IoError(std::io::Error::new(
                std::io::ErrorKind::NotFound,
                format!("Registry root not found: {}", root_path.display()),
            )));
        }

        let mut queries_by_name = HashMap::new();
        let mut queries_by_category = HashMap::new();

        // Discover all .rq files recursively
        Self::discover_queries(&root_path, &mut queries_by_name, &mut queries_by_category)?;

        Ok(QueryRegistry {
            queries_by_name,
            queries_by_category,
            root_dir: root_path,
        })
    }

    /// Recursively discover .rq files
    fn discover_queries(
        dir: &Path,
        by_name: &mut HashMap<String, ConstructQuery>,
        by_category: &mut HashMap<String, Vec<ConstructQuery>>,
    ) -> Result<()> {
        for entry in fs::read_dir(dir)? {
            let entry = entry?;
            let path = entry.path();

            if path.is_dir() {
                Self::discover_queries(&path, by_name, by_category)?;
            } else if path.extension().map_or(false, |ext| ext == "rq") {
                let content = fs::read_to_string(&path)?;
                let query = ConstructQuery::from_file(path, content)?;

                // Index by name
                by_name.insert(query.metadata.name.clone(), query.clone());

                // Index by category
                by_category
                    .entry(query.metadata.category.clone())
                    .or_insert_with(Vec::new)
                    .push(query);
            }
        }

        Ok(())
    }

    /// Look up a query by name
    pub fn lookup(&self, name: &str) -> Result<ConstructQuery> {
        self.queries_by_name
            .get(name)
            .cloned()
            .ok_or_else(|| RegistryError::QueryNotFound(name.to_string()))
    }

    /// Get all queries in a category
    pub fn queries_in_category(&self, category: &str) -> Result<Vec<ConstructQuery>> {
        self.queries_by_category
            .get(category)
            .map(|q| q.clone())
            .ok_or_else(|| RegistryError::CategoryNotFound(category.to_string()))
    }

    /// List all available categories
    pub fn categories(&self) -> Vec<String> {
        let mut cats: Vec<_> = self.queries_by_category.keys().cloned().collect();
        cats.sort();
        cats
    }

    /// List all available query names
    pub fn query_names(&self) -> Vec<String> {
        let mut names: Vec<_> = self.queries_by_name.keys().cloned().collect();
        names.sort();
        names
    }

    /// Get registry statistics
    pub fn stats(&self) -> RegistryStats {
        RegistryStats {
            total_queries: self.queries_by_name.len(),
            categories: self.queries_by_category.len(),
            queries_per_category: self
                .queries_by_category
                .iter()
                .map(|(cat, queries)| (cat.clone(), queries.len()))
                .collect(),
        }
    }
}

/// Registry statistics
#[derive(Debug, Serialize)]
pub struct RegistryStats {
    pub total_queries: usize,
    pub categories: usize,
    pub queries_per_category: HashMap<String, usize>,
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_metadata_parsing() {
        let sparql = r#"# @name: test_query
# @category: artifacts
# @description: Test query
# @params: ?id, ?name
# @returns: triples
# @version: 1.0
# @stability: stable
# @requires_parameters: true

PREFIX test: <http://test.dev/>

CONSTRUCT { ?s ?p ?o } WHERE { ?s ?p ?o }"#;

        let metadata = ConstructQuery::parse_metadata(sparql, "test").unwrap();
        assert_eq!(metadata.name, "test_query");
        assert_eq!(metadata.category, "artifacts");
        assert_eq!(metadata.params.len(), 2);
        assert!(metadata.requires_parameters);
    }

    #[test]
    fn test_binding() {
        let query = ConstructQuery {
            metadata: QueryMetadata {
                name: "test".to_string(),
                category: "test".to_string(),
                description: "Test".to_string(),
                params: vec!["?id".to_string(), "?name".to_string()],
                returns: "triples".to_string(),
                version: "1.0".to_string(),
                stability: "stable".to_string(),
                requires_parameters: true,
            },
            sparql: "CONSTRUCT { ?s ?p ?o } WHERE { ?s ?p ?o }".to_string(),
            path: PathBuf::from("test.rq"),
        };

        let mut params = HashMap::new();
        params.insert("id".to_string(), "123".to_string());
        params.insert("name".to_string(), "Test Name".to_string());

        let bound = query.bind(params).unwrap();
        let sparql = bound.sparql_with_bindings();

        assert!(sparql.contains("BIND("));
        assert!(sparql.contains("\"123\""));
        assert!(sparql.contains("\"Test Name\""));
    }
}
