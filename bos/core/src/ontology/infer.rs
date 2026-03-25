//! Auto-Ontology inference — generates ontology mappings from ODCS DataModel.
//!
//! Reads a workspace's `model.json` (data-modelling-sdk::DataModel) and
//! produces MappingConfig JSON compatible with existing ontology commands.
//!
//! **V1 Scope:** Enum detection uses `Column.enum_values` from the DataModel.
//! Value frequency analysis on string columns is deferred to v2 (requires data access).

use crate::ontology::mapping::{MappingConfig, PropertyMapping, Relationship, TableMapping};
use anyhow::{Context, Result};
use data_modelling_sdk::DataModel;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::path::Path;

// ---------------------------------------------------------------------------
// Public types
// ---------------------------------------------------------------------------

/// Configuration for the inference process.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct InferConfig {
    /// Minimum confidence threshold (0.0-1.0). Mappings below this are excluded.
    pub confidence_threshold: f64,
    /// Custom convention overrides (table name -> (ontology_namespace, class_name)).
    pub table_overrides: HashMap<String, (String, String)>,
    /// Custom column-to-predicate overrides (column name -> predicate).
    pub column_overrides: HashMap<String, String>,
}

impl Default for InferConfig {
    fn default() -> Self {
        Self {
            confidence_threshold: 0.0,
            table_overrides: HashMap::new(),
            column_overrides: HashMap::new(),
        }
    }
}

impl InferConfig {
    /// Config that only includes high-confidence mappings.
    pub fn high_confidence() -> Self {
        Self {
            confidence_threshold: 0.8,
            ..Default::default()
        }
    }
}

/// Confidence level for an inferred mapping.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum ConfidenceLevel {
    /// >= 0.8 — direct convention match
    High,
    /// 0.5-0.8 — partial match
    Medium,
    /// < 0.5 — no pattern match
    Low,
}

impl ConfidenceLevel {
    pub fn from_score(score: f64) -> Self {
        if score >= 0.8 {
            Self::High
        } else if score >= 0.5 {
            Self::Medium
        } else {
            Self::Low
        }
    }
}

/// Result of ontology inference.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct InferResult {
    pub tables_inferred: usize,
    pub properties_inferred: usize,
    pub relationships_inferred: usize,
    pub high_confidence: usize,
    pub medium_confidence: usize,
    pub low_confidence: usize,
    pub config: MappingConfig,
}

// ---------------------------------------------------------------------------
// Convention lookup tables
// ---------------------------------------------------------------------------

/// Built-in table name -> (ontology_namespace, class_name) conventions.
fn table_to_class(table: &str) -> Option<(String, String)> {
    let entry = match table {
        "projects" => Some(("schema", "Project")),
        "tasks" => Some(("bpmn", "Task")),
        "team_members" => Some(("org", "Member")),
        "members" => Some(("org", "Member")),
        "clients" => Some(("org", "Organization")),
        "organizations" => Some(("org", "Organization")),
        "contexts" => Some(("skos", "Concept")),
        "conversations" => Some(("schema", "Discussion")),
        "artifacts" => Some(("prov", "Entity")),
        "orders" => Some(("schema", "Order")),
        "invoices" => Some(("schema", "Invoice")),
        "employees" => Some(("org", "FormalOrganization")),
        "users" => Some(("foaf", "Person")),
        _ => None,
    };
    entry.map(|(ns, cls)| (ns.to_string(), cls.to_string()))
}

/// Built-in column name -> predicate conventions.
fn column_to_predicate(column: &str) -> Option<&'static str> {
    match column {
        "name" | "title" => Some("schema:name"),
        "description" | "summary" | "notes" => Some("schema:description"),
        "email" => Some("foaf:mbox"),
        "phone" => Some("schema:telephone"),
        "website" | "url" => Some("schema:url"),
        "status" => Some("schema:status"),
        "priority" => Some("schema:priority"),
        "role" => Some("org:role"),
        "type" => Some("rdf:type"),
        "content" | "body" => Some("schema:text"),
        "start_date" => Some("schema:startDate"),
        "end_date" | "due_date" => Some("schema:endDate"),
        "created_at" => Some("schema:dateCreated"),
        "updated_at" => Some("schema:dateModified"),
        "deleted_at" => Some("schema:dateDeleted"),
        "hourly_rate" => Some("schema:priceSpecification"),
        "price" | "amount" => Some("schema:price"),
        "industry" => Some("schema:industry"),
        "language" => Some("schema:programmingLanguage"),
        _ => None,
    }
}

/// Map column data_type to XSD datatype.
fn column_type_to_xsd(data_type: &str) -> Option<&'static str> {
    match data_type.to_lowercase().as_str() {
        "string" | "text" | "varchar" | "char" | "character varying"
        | "nvarchar" | "nvarchar2" | "varchar2" => Some("xsd:string"),
        "integer" | "int" | "bigint" | "smallint" | "serial" | "bigserial" => Some("xsd:integer"),
        "number" | "float" | "double" | "decimal" | "numeric" | "real" => Some("xsd:decimal"),
        "boolean" | "bool" => Some("xsd:boolean"),
        "timestamp" | "datetime" | "timestamptz" => Some("xsd:dateTime"),
        "date" => Some("xsd:date"),
        _ => None,
    }
}

// ---------------------------------------------------------------------------
// Confidence scoring
// ---------------------------------------------------------------------------

/// Compute confidence score for a table-to-class mapping.
fn table_confidence(table: &str) -> f64 {
    if table_to_class(table).is_some() {
        return 1.0;
    }
    0.3 // generic singularization fallback
}

/// Known prefix fragments for partial column matching.
const KNOWN_PREFIXES: &[&str] = &[
    "name", "description", "email", "phone", "url", "status", "priority",
    "role", "type", "content", "date", "created", "updated", "price",
    "amount", "industry", "language",
];

/// Compute confidence score for a column-to-predicate mapping.
fn column_confidence(column: &str, _data_type: &str) -> f64 {
    if column_to_predicate(column).is_some() {
        return 1.0;
    }

    let lower = column.to_lowercase();
    for prefix in KNOWN_PREFIXES {
        if lower.contains(prefix) {
            return 0.6; // partial match
        }
    }
    if lower.ends_with("_id") {
        return 0.7; // FK column pattern
    }
    if lower.ends_with("_at") || lower.ends_with("_date") || lower.ends_with("_time") {
        return 0.7; // temporal column pattern
    }

    0.2 // no pattern match
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/// Basic English singularization (handles -ies, -ses, -s).
fn singularize(word: &str) -> String {
    if word.ends_with("ies") && word.len() > 3 {
        format!("{}y", &word[..word.len() - 3])
    } else if word.ends_with("ses") && word.len() > 3 {
        format!("{}s", &word[..word.len() - 2])
    } else if word.ends_with("s") && word.len() > 1 {
        word[..word.len() - 1].to_string()
    } else {
        word.to_string()
    }
}

/// Capitalize first letter.
fn capitalize(s: &str) -> String {
    let mut chars = s.chars();
    match chars.next() {
        Some(c) => c.to_uppercase().collect::<String>() + chars.as_str(),
        None => String::new(),
    }
}

/// Fallback: singularize table name and assign default schema.org namespace.
fn infer_class_from_table(table: &str) -> (String, String) {
    let singular = singularize(table);
    let class = capitalize(&singular);
    ("schema".to_string(), class)
}

// ---------------------------------------------------------------------------
// OntologyInferrer
// ---------------------------------------------------------------------------

/// Infers ontology mappings from an ODCS DataModel.
pub struct OntologyInferrer {
    config: InferConfig,
}

impl OntologyInferrer {
    pub fn new(config: InferConfig) -> Self {
        Self { config }
    }

    /// Infer ontology mappings from a DataModel.
    pub fn infer(&self, model: &DataModel) -> Result<InferResult> {
        let mut mappings = Vec::new();
        let mut relationships = Vec::new();
        let mut high = 0usize;
        let mut medium = 0usize;
        let mut low = 0usize;
        let mut total_props = 0usize;

        // Build lookups for relationship resolution
        let uuid_to_name: HashMap<uuid::Uuid, String> = model
            .tables
            .iter()
            .map(|t| (t.id, t.name.clone()))
            .collect();
        let table_names: std::collections::HashSet<&str> =
            model.tables.iter().map(|t| t.name.as_str()).collect();

        for table in &model.tables {
            let (ontology_ns, class) = self.resolve_table_class(&table.name);
            let conf = table_confidence(&table.name);

            match ConfidenceLevel::from_score(conf) {
                ConfidenceLevel::High => high += 1,
                ConfidenceLevel::Medium => medium += 1,
                ConfidenceLevel::Low => low += 1,
            }

            let pk_column = table
                .columns
                .iter()
                .find(|c| c.primary_key)
                .map(|c| c.name.clone())
                .unwrap_or_else(|| "id".to_string());

            let uri_template = format!(
                "http://businessos.dev/id/{}/{{{}}}",
                table.name, pk_column
            );

            let mut properties = Vec::new();

            for column in &table.columns {
                let prop_conf = column_confidence(&column.name, &column.data_type);

                if prop_conf < self.config.confidence_threshold {
                    continue;
                }

                let predicate = self.resolve_column_predicate(&column.name);

                let datatype = column_type_to_xsd(&column.data_type)
                    .unwrap_or("xsd:string")
                    .to_string();

                let mut prop = PropertyMapping {
                    column: column.name.clone(),
                    predicate,
                    datatype,
                    is_primary_key: column.primary_key,
                    object_type: None,
                    target_table: None,
                    value_map: HashMap::new(),
                };

                // FK detection from column.foreign_key
                if let Some(ref fk) = column.foreign_key {
                    prop.object_type = Some("uri".to_string());
                    if let Ok(fk_uuid) = uuid::Uuid::parse_str(&fk.table_id) {
                        if let Some(target_name) = uuid_to_name.get(&fk_uuid) {
                            prop.target_table = Some(target_name.clone());
                        }
                    }
                }

                // FK detection from _id suffix (only if no explicit FK and not PK)
                if prop.object_type.is_none()
                    && column.name.ends_with("_id")
                    && !column.primary_key
                {
                    let target = column.name.trim_end_matches("_id");
                    if table_names.contains(target) {
                        prop.object_type = Some("uri".to_string());
                        prop.target_table = Some(target.to_string());
                    }
                }

                // Enum detection from Column.enum_values
                if !column.enum_values.is_empty() && column.enum_values.len() <= 10 {
                    for val in &column.enum_values {
                        let uri = format!("http://businessos.dev/enum/{}/{}", table.name, val);
                        prop.value_map.insert(val.clone(), uri);
                    }
                }

                properties.push(prop);
                total_props += 1;
            }

            mappings.push(TableMapping {
                table: table.name.clone(),
                ontology: ontology_ns,
                class,
                uri_template,
                properties,
            });
        }

        // Infer relationships from DataModel.relationships
        for rel in &model.relationships {
            let source_name = uuid_to_name
                .get(&rel.source_table_id)
                .cloned()
                .unwrap_or_default();
            let target_name = uuid_to_name
                .get(&rel.target_table_id)
                .cloned()
                .unwrap_or_default();

            if source_name.is_empty() || target_name.is_empty() {
                continue;
            }

            let property = rel
                .label
                .clone()
                .unwrap_or_else(|| "schema:relatedTo".to_string());

            relationships.push(Relationship {
                from_table: source_name,
                to_table: target_name,
                property,
                inverse: "schema:inverseOf".to_string(),
            });
        }

        // Sort mappings by table name for deterministic output
        mappings.sort_by(|a, b| a.table.cmp(&b.table));

        // Build standard prefixes
        let mut prefixes = HashMap::new();
        prefixes.insert("rdf".to_string(), "http://www.w3.org/1999/02/22-rdf-syntax-ns#".to_string());
        prefixes.insert("rdfs".to_string(), "http://www.w3.org/2000/01/rdf-schema#".to_string());
        prefixes.insert("xsd".to_string(), "http://www.w3.org/2001/XMLSchema#".to_string());
        prefixes.insert("owl".to_string(), "http://www.w3.org/2002/07/owl#".to_string());
        prefixes.insert("schema".to_string(), "https://schema.org/".to_string());
        prefixes.insert("bpmn".to_string(), "http://www.omg.org/spec/BPMN/20100524/MODEL#".to_string());
        prefixes.insert("org".to_string(), "http://www.w3.org/ns/org#".to_string());
        prefixes.insert("prov".to_string(), "http://www.w3.org/ns/prov#".to_string());
        prefixes.insert("skos".to_string(), "http://www.w3.org/2004/02/skos/core#".to_string());
        prefixes.insert("foaf".to_string(), "http://xmlns.com/foaf/0.1/".to_string());

        Ok(InferResult {
            tables_inferred: mappings.len(),
            properties_inferred: total_props,
            relationships_inferred: relationships.len(),
            high_confidence: high,
            medium_confidence: medium,
            low_confidence: low,
            config: MappingConfig {
                prefixes,
                mappings,
                relationships,
            },
        })
    }

    /// Infer ontology mappings from a workspace directory (reads model.json).
    pub fn infer_from_workspace(workspace_path: &Path, config: InferConfig) -> Result<InferResult> {
        let model_path = workspace_path.join("model.json");
        let content = std::fs::read_to_string(&model_path)
            .with_context(|| format!("Failed to read model.json: {}", model_path.display()))?;
        let model: DataModel = serde_json::from_str(&content)
            .with_context(|| "Failed to parse model.json as DataModel")?;
        let inferrer = Self::new(config);
        inferrer.infer(&model)
    }

    /// Resolve table name to (ontology_namespace, class_name).
    fn resolve_table_class(&self, table: &str) -> (String, String) {
        if let Some((ns, cls)) = self.config.table_overrides.get(table) {
            return (ns.clone(), cls.clone());
        }
        table_to_class(table).unwrap_or_else(|| infer_class_from_table(table))
    }

    /// Resolve column name to predicate.
    fn resolve_column_predicate(&self, column: &str) -> String {
        if let Some(pred) = self.config.column_overrides.get(column) {
            return pred.clone();
        }
        column_to_predicate(column)
            .map(String::from)
            .unwrap_or_else(|| format!("schema:{}", column))
    }
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

#[cfg(test)]
mod tests {
    use super::*;

    // -- Task 1: table-to-class conventions --

    #[test]
    fn test_table_convention_projects() {
        assert_eq!(
            table_to_class("projects"),
            Some(("schema".to_string(), "Project".to_string()))
        );
    }

    #[test]
    fn test_table_convention_tasks() {
        assert_eq!(
            table_to_class("tasks"),
            Some(("bpmn".to_string(), "Task".to_string()))
        );
    }

    #[test]
    fn test_table_convention_unknown() {
        assert!(table_to_class("custom_table").is_none());
    }

    #[test]
    fn test_table_convention_singularization() {
        let result = infer_class_from_table("invoices");
        assert_eq!(result.0, "schema");
        assert_eq!(result.1, "Invoice");
    }

    // -- Task 2: column-to-predicate and datatype conventions --

    #[test]
    fn test_column_convention_name() {
        assert_eq!(column_to_predicate("name"), Some("schema:name"));
    }

    #[test]
    fn test_column_convention_created_at() {
        assert_eq!(column_to_predicate("created_at"), Some("schema:dateCreated"));
    }

    #[test]
    fn test_column_convention_unknown() {
        assert_eq!(column_to_predicate("xyz123"), None);
    }

    #[test]
    fn test_column_type_string() {
        assert_eq!(column_type_to_xsd("string"), Some("xsd:string"));
        assert_eq!(column_type_to_xsd("text"), Some("xsd:string"));
        assert_eq!(column_type_to_xsd("varchar"), Some("xsd:string"));
    }

    #[test]
    fn test_column_type_integer() {
        assert_eq!(column_type_to_xsd("integer"), Some("xsd:integer"));
        assert_eq!(column_type_to_xsd("bigint"), Some("xsd:integer"));
    }

    #[test]
    fn test_column_type_timestamp() {
        assert_eq!(column_type_to_xsd("timestamp"), Some("xsd:dateTime"));
        assert_eq!(column_type_to_xsd("date"), Some("xsd:date"));
    }

    #[test]
    fn test_column_type_unknown() {
        assert_eq!(column_type_to_xsd("custom_type"), None);
    }

    // -- Task 3: confidence scoring --

    #[test]
    fn test_confidence_direct_convention_match() {
        assert_eq!(table_confidence("projects"), 1.0);
    }

    #[test]
    fn test_confidence_generic_singularization() {
        let conf = table_confidence("unknown_table_xyz");
        assert!(conf < 0.5);
    }

    #[test]
    fn test_confidence_column_direct() {
        assert_eq!(column_confidence("name", "string"), 1.0);
        assert_eq!(column_confidence("created_at", "timestamp"), 1.0);
    }

    #[test]
    fn test_confidence_column_partial() {
        let conf = column_confidence("status_val", "string");
        assert!((0.5..0.8).contains(&conf));
    }

    #[test]
    fn test_confidence_column_unknown() {
        let conf = column_confidence("xyz", "string");
        assert!(conf < 0.5);
    }

    #[test]
    fn test_confidence_level_enum() {
        assert_eq!(ConfidenceLevel::from_score(0.9), ConfidenceLevel::High);
        assert_eq!(ConfidenceLevel::from_score(0.7), ConfidenceLevel::Medium);
        assert_eq!(ConfidenceLevel::from_score(0.3), ConfidenceLevel::Low);
    }

    // -- Task 4: full inference pipeline --

    #[test]
    fn test_infer_from_data_model() {
        use data_modelling_sdk::{Column, DataModel, Table};

        let mut cols = vec![
            Column::new("id".to_string(), "string".to_string()),
            Column::new("name".to_string(), "string".to_string()),
            Column::new("status".to_string(), "string".to_string()),
            Column::new("created_at".to_string(), "timestamp".to_string()),
            Column::new("client_id".to_string(), "string".to_string()),
        ];
        cols[0].primary_key = true;
        cols[4].foreign_key = Some(data_modelling_sdk::models::column::ForeignKey {
            table_id: "00000000-0000-0000-0000-000000000001".to_string(),
            column_name: "id".to_string(),
        });

        let table = Table::new("projects".to_string(), cols);
        let mut model =
            DataModel::new("test-ws".to_string(), ".".to_string(), "control.yaml".to_string());
        model.tables.push(table);

        let inferrer = OntologyInferrer::new(InferConfig::default());
        let result = inferrer.infer(&model).unwrap();

        assert_eq!(result.tables_inferred, 1);
        assert_eq!(result.config.mappings.len(), 1);

        let mapping = &result.config.mappings[0];
        assert_eq!(mapping.table, "projects");
        assert_eq!(mapping.class, "Project");
        assert_eq!(mapping.ontology, "schema");
        assert!(mapping.properties.len() >= 4);

        let id_prop = mapping
            .properties
            .iter()
            .find(|p| p.column == "id")
            .expect("id property should exist");
        assert!(id_prop.is_primary_key);

        let client_prop = mapping
            .properties
            .iter()
            .find(|p| p.column == "client_id");
        assert!(client_prop.is_some());
        assert_eq!(client_prop.unwrap().object_type.as_deref(), Some("uri"));
    }

    #[test]
    fn test_infer_filters_by_confidence() {
        use data_modelling_sdk::{Column, DataModel, Table};

        let mut cols = vec![
            Column::new("id".to_string(), "string".to_string()),
            Column::new("name".to_string(), "string".to_string()),
            Column::new("xyz123".to_string(), "string".to_string()),
        ];
        cols[0].primary_key = true;

        let table = Table::new("projects".to_string(), cols);
        let mut model =
            DataModel::new("test".to_string(), ".".to_string(), "control.yaml".to_string());
        model.tables.push(table);

        let inferrer = OntologyInferrer::new(InferConfig::high_confidence());
        let result = inferrer.infer(&model).unwrap();

        let mapping = &result.config.mappings[0];
        assert!(
            mapping.properties.iter().find(|p| p.column == "xyz123").is_none(),
            "low-confidence column should be filtered out"
        );
    }

    #[test]
    fn test_infer_multiple_tables() {
        use data_modelling_sdk::{Column, DataModel, Table};

        let mut model =
            DataModel::new("test".to_string(), ".".to_string(), "control.yaml".to_string());

        let mut proj_cols = vec![
            Column::new("id".to_string(), "string".to_string()),
            Column::new("name".to_string(), "string".to_string()),
        ];
        proj_cols[0].primary_key = true;
        model.tables.push(Table::new("projects".to_string(), proj_cols));

        let mut task_cols = vec![
            Column::new("id".to_string(), "string".to_string()),
            Column::new("title".to_string(), "string".to_string()),
            Column::new("status".to_string(), "string".to_string()),
        ];
        task_cols[0].primary_key = true;
        model.tables.push(Table::new("tasks".to_string(), task_cols));

        let inferrer = OntologyInferrer::new(InferConfig::default());
        let result = inferrer.infer(&model).unwrap();

        assert_eq!(result.tables_inferred, 2);
        assert_eq!(result.config.mappings.len(), 2);
        assert_eq!(result.config.mappings[0].class, "Project");
        assert_eq!(result.config.mappings[1].class, "Task");
    }

    #[test]
    fn test_infer_from_workspace_path() {
        use data_modelling_sdk::{Column, DataModel, Table};
        use tempfile::TempDir;

        let tmp = TempDir::new().unwrap();
        let model_path = tmp.path().join("model.json");

        // Build a real DataModel and serialize it (SDK uses camelCase)
        let mut cols = vec![
            Column::new("id".to_string(), "string".to_string()),
            Column::new("name".to_string(), "string".to_string()),
        ];
        cols[0].primary_key = true;
        let table = Table::new("projects".to_string(), cols);
        let mut model =
            DataModel::new("test-ws".to_string(), tmp.path().to_string_lossy().to_string(), "control.yaml".to_string());
        model.tables.push(table);

        std::fs::write(&model_path, serde_json::to_string_pretty(&model).unwrap()).unwrap();

        let result =
            OntologyInferrer::infer_from_workspace(tmp.path(), InferConfig::default()).unwrap();
        assert_eq!(result.tables_inferred, 1);
    }

    #[test]
    fn test_infer_from_workspace_missing_model() {
        use tempfile::TempDir;

        let tmp = TempDir::new().unwrap();
        let result = OntologyInferrer::infer_from_workspace(tmp.path(), InferConfig::default());
        assert!(result.is_err());
    }

    #[test]
    fn test_infer_with_custom_overrides() {
        use data_modelling_sdk::{Column, DataModel, Table};

        let mut config = InferConfig::default();
        config
            .table_overrides
            .insert("projects".to_string(), ("custom".to_string(), "Widget".to_string()));
        config
            .column_overrides
            .insert("name".to_string(), "custom:label".to_string());

        let mut cols = vec![
            Column::new("id".to_string(), "string".to_string()),
            Column::new("name".to_string(), "string".to_string()),
        ];
        cols[0].primary_key = true;

        let table = Table::new("projects".to_string(), cols);
        let mut model =
            DataModel::new("test".to_string(), ".".to_string(), "control.yaml".to_string());
        model.tables.push(table);

        let inferrer = OntologyInferrer::new(config);
        let result = inferrer.infer(&model).unwrap();

        let mapping = &result.config.mappings[0];
        assert_eq!(mapping.ontology, "custom");
        assert_eq!(mapping.class, "Widget");

        let name_prop = mapping
            .properties
            .iter()
            .find(|p| p.column == "name")
            .expect("name property should exist");
        assert_eq!(name_prop.predicate, "custom:label");
    }
}
