# Auto-Ontology Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement `bos ontology infer` — automatically generate ontology mappings from any ODCS workspace's `model.json` without manual configuration.

**Architecture:** The infer module reads `data-modelling-sdk::DataModel` (table/column/relationship definitions) from a workspace's `model.json`, applies convention-based mapping rules (table name → OWL class, column name → RDF predicate, column type → XSD datatype), and outputs a valid `MappingConfig` JSON compatible with existing `ontology construct`, `ontology export`, and `ontology execute` commands. Confidence scoring lets users filter low-quality inferences.

**V1 Scope Note:** Enum detection uses `Column.enum_values` from the DataModel (explicitly set by schema conversion). Value frequency analysis on string columns (scanning actual data) is deferred to v2 since it requires data access (CSV exports or DB connection), contradicting the spec's "no live PostgreSQL" constraint. Users can manually add `enum_values` to their `model.json` for v1.

**Tech Stack:** Rust (bos core crate), data-modelling-sdk v2.4 (DataModel type), clap-noun-verb (CLI), serde (JSON serialization), tempfile (tests)

**Spec:** `/Users/sac/chatmangpt/BusinessOS/docs/superpowers/specs/2026-03-23-reasoning-layer-design.md` (Innovation 1)

---

## File Structure

| File | Responsibility |
|------|---------------|
| `bos/core/src/ontology/infer.rs` | Core inference logic: convention tables, confidence scoring, DataModel → MappingConfig pipeline |
| `bos/core/src/ontology/mod.rs` | Add `pub mod infer;` |
| `bos/core/src/lib.rs` | Re-export `OntologyInferrer`, `InferConfig`, `InferResult` |
| `bos/cli/src/nouns/ontology.rs` | Add `infer` verb to existing ontology noun |
| Tests inline in `bos/core/src/ontology/infer.rs` | 10 tests covering convention matching, confidence scoring, FK detection, enum detection, overrides |

---

## Convention Tables (reference)

### Table-to-Class conventions

```
projects       → schema:Project
tasks          → bpmn:Task
team_members   → org:Member
members        → org:Member
clients        → org:Organization
organizations  → org:Organization
contexts       → skos:Concept
conversations  → schema:Discussion
artifacts      → prov:Entity
orders         → schema:Order
invoices       → schema:Invoice
employees      → org:FormalOrganization
users          → foaf:Person
```

### Column-to-Predicate conventions

```
id              → schema:identifier (if PK) or owl:sameAs
name            → schema:name
title           → schema:name
description     → schema:description
email           → foaf:mbox
phone           → schema:telephone
website         → schema:url
url             → schema:url
status          → schema:status
priority        → schema:priority
role            → org:role
type            → rdf:type
content         → schema:text
summary         → schema:description
notes           → schema:description
start_date      → schema:startDate
end_date        → schema:endDate
due_date        → schema:endDate
created_at      → schema:dateCreated
updated_at      → schema:dateModified
deleted_at      → schema:dateDeleted
hourly_rate     → schema:priceSpecification
price           → schema:price
amount          → schema:price
industry        → schema:industry
language        → schema:programmingLanguage
```

### Column-type-to-XSD conventions

```
string, text, varchar  → xsd:string
integer, int, bigint   → xsd:integer
number, float, double  → xsd:decimal
boolean, bool          → xsd:boolean
timestamp, datetime    → xsd:dateTime
date                   → xsd:date
```

---

## Task 1: Create `OntologyInferrer` struct with table-to-class convention lookup

**Files:**
- Create: `bos/core/src/ontology/infer.rs`
- Modify: `bos/core/src/ontology/mod.rs:1-4`

- [ ] **Step 1: Write the failing test for table-to-class convention lookup**

```rust
#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_table_convention_projects() {
        assert_eq!(table_to_class("projects"), Some(("schema".to_string(), "Project".to_string())));
    }

    #[test]
    fn test_table_convention_tasks() {
        assert_eq!(table_to_class("tasks"), Some(("bpmn".to_string(), "Task".to_string())));
    }

    #[test]
    fn test_table_convention_unknown() {
        let result = table_to_class("custom_table");
        assert!(result.is_none());
    }

    #[test]
    fn test_table_convention_singularization() {
        // Unknown table should fall through to generic singularization
        let result = infer_class_from_table("invoices");
        assert_eq!(result.0, "schema");
        assert_eq!(result.1, "Invoice");
    }
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/sac/chatmangpt/BusinessOS && cargo test -p bos-core --lib ontology::infer --no-run 2>&1 | head -20`
Expected: Compilation error — module doesn't exist

- [ ] **Step 3: Create infer.rs with convention tables and struct**

```rust
//! Auto-Ontology inference — generates ontology mappings from ODCS DataModel.
//!
//! Reads a workspace's `model.json` (data-modelling-sdk::DataModel) and
//! produces MappingConfig JSON compatible with existing ontology commands.

use crate::ontology::mapping::{MappingConfig, PropertyMapping, Relationship, TableMapping};
use anyhow::{Context, Result};
use data_modelling_sdk::DataModel;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::path::Path;

/// Configuration for the inference process.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct InferConfig {
    /// Minimum confidence threshold (0.0-1.0). Mappings below this are excluded.
    pub confidence_threshold: f64,
    /// Custom convention overrides (table name → ontology:class).
    pub table_overrides: HashMap<String, (String, String)>,
    /// Custom column-to-predicate overrides (column name → predicate).
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
    pub fn high_confidence() -> Self {
        Self {
            confidence_threshold: 0.8,
            ..Default::default()
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

/// Infers ontology mappings from an ODCS DataModel.
pub struct OntologyInferrer {
    config: InferConfig,
}

/// Built-in table name → (ontology_namespace, class_name) conventions.
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

/// Singularize a table name and assign default schema.org namespace.
/// Used as fallback when no convention matches.
fn infer_class_from_table(table: &str) -> (String, String) {
    let singular = singularize(table);
    let class = capitalize(&singular);
    ("schema".to_string(), class)
}

/// Basic English singularization.
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
```

- [ ] **Step 4: Add `pub mod infer;` to ontology/mod.rs**

```rust
pub mod bridge;
pub mod construct;
pub mod execute;
pub mod infer;
pub mod mapping;
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd /Users/sac/chatmangpt/BusinessOS && cargo test -p bos-core --lib ontology::infer`
Expected: 4 tests pass

- [ ] **Step 6: Commit**

```bash
cd /Users/sac/chatmangpt/BusinessOS
git add bos/core/src/ontology/infer.rs bos/core/src/ontology/mod.rs
git commit -m "feat(ontology): add OntologyInferrer with table-to-class convention lookup"
```

---

## Task 2: Add column-to-predicate and datatype conventions

**Files:**
- Modify: `bos/core/src/ontology/infer.rs`

- [ ] **Step 1: Write the failing tests**

```rust
    #[test]
    fn test_column_convention_name() {
        assert_eq!(column_to_predicate("name"), Some("schema:name".to_string()));
    }

    #[test]
    fn test_column_convention_created_at() {
        assert_eq!(column_to_predicate("created_at"), Some("schema:dateCreated".to_string()));
    }

    #[test]
    fn test_column_convention_unknown() {
        assert_eq!(column_to_predicate("xyz123"), None);
    }

    #[test]
    fn test_column_type_string() {
        assert_eq!(column_type_to_xsd("string"), Some("xsd:string".to_string()));
        assert_eq!(column_type_to_xsd("text"), Some("xsd:string".to_string()));
        assert_eq!(column_type_to_xsd("varchar"), Some("xsd:string".to_string()));
    }

    #[test]
    fn test_column_type_integer() {
        assert_eq!(column_type_to_xsd("integer"), Some("xsd:integer".to_string()));
        assert_eq!(column_type_to_xsd("bigint"), Some("xsd:integer".to_string()));
    }

    #[test]
    fn test_column_type_timestamp() {
        assert_eq!(column_type_to_xsd("timestamp"), Some("xsd:dateTime".to_string()));
        assert_eq!(column_type_to_xsd("date"), Some("xsd:date".to_string()));
    }

    #[test]
    fn test_column_type_unknown() {
        assert_eq!(column_type_to_xsd("custom_type"), None);
    }
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/sac/chatmangpt/BusinessOS && cargo test -p bos-core --lib ontology::infer::tests::test_column`
Expected: Compilation error — functions don't exist

- [ ] **Step 3: Implement column convention functions**

Add these functions to `infer.rs`:

```rust
/// Built-in column name → predicate conventions.
fn column_to_predicate(column: &str) -> Option<String> {
    let entry = match column {
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
    };
    entry.map(String::from)
}

/// Map column data_type to XSD datatype.
fn column_type_to_xsd(data_type: &str) -> Option<String> {
    let lower = data_type.to_lowercase();
    match lower.as_str() {
        "string" | "text" | "varchar" | "char" | "character varying"
        | "nvarchar" | "nvarchar2" | "varchar2" => Some("xsd:string".to_string()),
        "integer" | "int" | "bigint" | "smallint" | "serial" | "bigserial" => Some("xsd:integer".to_string()),
        "number" | "float" | "double" | "decimal" | "numeric" | "real" => Some("xsd:decimal".to_string()),
        "boolean" | "bool" => Some("xsd:boolean".to_string()),
        "timestamp" | "datetime" | "timestamptz" => Some("xsd:dateTime".to_string()),
        "date" => Some("xsd:date".to_string()),
        _ => None,
    }
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/sac/chatmangpt/BusinessOS && cargo test -p bos-core --lib ontology::infer`
Expected: All tests pass (4 from Task 1 + 7 from Task 2 = 11)

- [ ] **Step 5: Commit**

```bash
cd /Users/sac/chatmangpt/BusinessOS
git add bos/core/src/ontology/infer.rs
git commit -m "feat(ontology): add column-to-predicate and datatype convention mappings"
```

---

## Task 3: Implement confidence scoring

**Files:**
- Modify: `bos/core/src/ontology/infer.rs`

- [ ] **Step 1: Write the failing tests**

```rust
    #[test]
    fn test_confidence_direct_convention_match() {
        // Direct convention match = high confidence
        assert!(table_confidence("projects").is_some());
        assert_eq!(table_confidence("projects").unwrap(), 1.0);
    }

    #[test]
    fn test_confidence_generic_singularization() {
        // Generic singularization = low confidence
        let conf = table_confidence("unknown_table_xyz").unwrap();
        assert!(conf < 0.5);
    }

    #[test]
    fn test_confidence_column_direct() {
        assert_eq!(column_confidence("name", "string"), 1.0);
        assert_eq!(column_confidence("created_at", "timestamp"), 1.0);
    }

    #[test]
    fn test_confidence_column_partial() {
        // Column name partially matches (e.g., "status_val" has "status" prefix)
        let conf = column_confidence("status_val", "string");
        assert!(conf >= 0.5 && conf < 0.8);
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
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/sac/chatmangpt/BusinessOS && cargo test -p bos-core --lib ontology::infer::tests::test_confidence`
Expected: Compilation error

- [ ] **Step 3: Implement confidence scoring**

Add to `infer.rs`:

```rust
/// Confidence level for an inferred mapping.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum ConfidenceLevel {
    High,   // >= 0.8 — direct convention match
    Medium, // 0.5-0.8 — partial match
    Low,    // < 0.5 — no pattern match
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

/// Compute confidence score for a table-to-class mapping.
fn table_confidence(table: &str) -> Option<f64> {
    if table_to_class(table).is_some() {
        return Some(1.0); // Direct convention match
    }
    // Generic singularization — low confidence
    Some(0.3)
}

/// Compute confidence score for a column-to-predicate mapping.
fn column_confidence(column: &str, data_type: &str) -> f64 {
    // Direct convention match
    if column_to_predicate(column).is_some() {
        return 1.0;
    }

    // Partial match: check if column contains a known convention word
    let known_prefixes = ["name", "description", "email", "phone", "url", "status",
        "priority", "role", "type", "content", "date", "created", "updated",
        "price", "amount", "industry", "language"];
    let lower = column.to_lowercase();

    for prefix in known_prefixes {
        if lower.contains(prefix) {
            return 0.6; // Partial match
        }
    }

    // Common suffix patterns
    if lower.ends_with("_id") {
        return 0.7; // FK column pattern
    }
    if lower.ends_with("_at") || lower.ends_with("_date") || lower.ends_with("_time") {
        return 0.7; // Temporal column pattern
    }

    // No pattern match at all
    0.2
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/sac/chatmangpt/BusinessOS && cargo test -p bos-core --lib ontology::infer`
Expected: All tests pass

- [ ] **Step 5: Commit**

```bash
cd /Users/sac/chatmangpt/BusinessOS
git add bos/core/src/ontology/infer.rs
git commit -m "feat(ontology): add confidence scoring for inferred mappings"
```

---

## Task 4: Implement the full inference pipeline (DataModel → MappingConfig)

**Files:**
- Modify: `bos/core/src/ontology/infer.rs`

This is the core logic: read DataModel, produce MappingConfig with confidence-filtered output.

- [ ] **Step 1: Write the failing integration test**

```rust
    #[test]
    fn test_infer_from_data_model() {
        use data_modelling_sdk::{DataModel, Table, Column};

        // Build a minimal DataModel matching BusinessOS schema
        let columns = vec![
            Column::new("id".to_string(), "string".to_string()),
            Column::new("name".to_string(), "string".to_string()),
            Column::new("status".to_string(), "string".to_string()),
            Column::new("created_at".to_string(), "timestamp".to_string()),
            Column::new("client_id".to_string(), "string".to_string()),
        ];
        // Set PK on first column
        let mut cols = columns;
        cols[0].primary_key = true;
        // Set FK on client_id
        cols[4].foreign_key = Some(data_modelling_sdk::models::column::ForeignKey {
            table_id: "target-uuid".to_string(),
            column_name: "id".to_string(),
        });

        let table = Table::new("projects".to_string(), cols);
        let mut model = DataModel::new(
            "test-workspace".to_string(),
            ".".to_string(),
            "control.yaml".to_string(),
        );
        model.tables.push(table);

        let inferrer = OntologyInferrer::new(InferConfig::default());
        let result = inferrer.infer(&model).unwrap();

        // Should infer 1 table
        assert_eq!(result.tables_inferred, 1);

        // Should produce valid MappingConfig
        let config = &result.config;
        assert_eq!(config.mappings.len(), 1);

        let mapping = &config.mappings[0];
        assert_eq!(mapping.table, "projects");
        assert_eq!(mapping.class, "Project");
        assert_eq!(mapping.ontology, "schema");

        // Should have property mappings for each column
        assert!(mapping.properties.len() >= 4); // id, name, status, created_at

        // id should be primary_key
        let id_prop = mapping.properties.iter().find(|p| p.column == "id").unwrap();
        assert!(id_prop.is_primary_key);

        // client_id should be object_type uri
        let client_prop = mapping.properties.iter().find(|p| p.column == "client_id");
        assert!(client_prop.is_some());
        let client_prop = client_prop.unwrap();
        assert_eq!(client_prop.object_type.as_deref(), Some("uri"));
    }

    #[test]
    fn test_infer_filters_by_confidence() {
        use data_modelling_sdk::{DataModel, Table, Column};

        let columns = vec![
            Column::new("id".to_string(), "string".to_string()),
            Column::new("name".to_string(), "string".to_string()),
            Column::new("xyz123".to_string(), "string".to_string()),
        ];
        let mut cols = columns;
        cols[0].primary_key = true;

        let table = Table::new("projects".to_string(), cols);
        let mut model = DataModel::new("test".to_string(), ".".to_string(), "control.yaml".to_string());
        model.tables.push(table);

        // High confidence only — should exclude "xyz123"
        let inferrer = OntologyInferrer::new(InferConfig::high_confidence());
        let result = inferrer.infer(&model).unwrap();

        let mapping = &result.config.mappings[0];
        let xyz_prop = mapping.properties.iter().find(|p| p.column == "xyz123");
        assert!(xyz_prop.is_none(), "low-confidence column should be filtered out");
    }

    #[test]
    fn test_infer_multiple_tables() {
        use data_modelling_sdk::{DataModel, Table, Column};

        let mut model = DataModel::new("test".to_string(), ".".to_string(), "control.yaml".to_string());

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

        // projects → schema:Project, tasks → bpmn:Task
        assert_eq!(result.config.mappings[0].class, "Project");
        assert_eq!(result.config.mappings[1].class, "Task");
    }

    #[test]
    fn test_infer_from_workspace_path() {
        use tempfile::TempDir;

        let tmp = TempDir::new().unwrap();
        let model_path = tmp.path().join("model.json");

        // Write a minimal model.json
        let json = serde_json::json!({
            "name": "test-ws",
            "tables": [{
                "name": "projects",
                "columns": [
                    {"name": "id", "data_type": "string", "primary_key": true},
                    {"name": "name", "data_type": "string"}
                ]
            }],
            "relationships": []
        });
        std::fs::write(&model_path, serde_json::to_string_pretty(&json).unwrap()).unwrap();

        let result = OntologyInferrer::infer_from_workspace(tmp.path(), InferConfig::default()).unwrap();
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
        use data_modelling_sdk::{DataModel, Table, Column};

        let mut config = InferConfig::default();
        config.table_overrides.insert("projects".to_string(), ("custom".to_string(), "Widget".to_string()));
        config.column_overrides.insert("name".to_string(), "custom:label".to_string());

        let mut cols = vec![
            Column::new("id".to_string(), "string".to_string()),
            Column::new("name".to_string(), "string".to_string()),
        ];
        cols[0].primary_key = true;

        let table = Table::new("projects".to_string(), cols);
        let mut model = DataModel::new("test".to_string(), ".".to_string(), "control.yaml".to_string());
        model.tables.push(table);

        let inferrer = OntologyInferrer::new(config);
        let result = inferrer.infer(&model).unwrap();

        let mapping = &result.config.mappings[0];
        // Override should replace built-in "schema:Project"
        assert_eq!(mapping.ontology, "custom");
        assert_eq!(mapping.class, "Widget");

        // Column override should replace built-in "schema:name"
        let name_prop = mapping.properties.iter().find(|p| p.column == "name").expect("name property should exist");
        assert_eq!(name_prop.predicate, "custom:label");
    }
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/sac/chatmangpt/BusinessOS && cargo test -p bos-core --lib ontology::infer::tests::test_infer`
Expected: Compilation error — methods don't exist

- [ ] **Step 3: Implement OntologyInferrer::new(), infer(), and infer_from_workspace()**

Add to `OntologyInferrer` impl block in `infer.rs`:

```rust
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

        // Build table-name-to-uuid lookup for relationship resolution
        let table_uuids: HashMap<String, uuid::Uuid> = model.tables.iter()
            .map(|t| (t.name.clone(), t.id))
            .collect();

        // Build uuid-to-table-name reverse lookup
        let uuid_to_name: HashMap<uuid::Uuid, String> = model.tables.iter()
            .map(|t| (t.id, t.name.clone()))
            .collect();

        for table in &model.tables {
            let (ontology_ns, class) = self.resolve_table_class(&table.name);
            let conf = table_confidence(&table.name).unwrap_or(0.3);

            match ConfidenceLevel::from_score(conf) {
                ConfidenceLevel::High => high += 1,
                ConfidenceLevel::Medium => medium += 1,
                ConfidenceLevel::Low => low += 1,
            }

            let pk_column = table.columns.iter()
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

                let predicate = self.resolve_column_predicate(&column.name, &table.name);

                let datatype = column_type_to_xsd(&column.data_type)
                    .unwrap_or_else(|| "xsd:string".to_string());

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
                    // Try to find target table name from UUID
                    if let Ok(fk_uuid) = uuid::Uuid::parse_str(&fk.table_id) {
                        if let Some(target_name) = uuid_to_name.get(&fk_uuid) {
                            prop.target_table = Some(target_name.clone());
                        }
                    }
                }

                // FK detection from _id suffix if no explicit FK
                if prop.object_type.is_none() && column.name.ends_with("_id") && !column.primary_key {
                    let target = column.name.trim_end_matches("_id");
                    // Only set if target table exists in model
                    if table_uuids.contains_key(target) {
                        prop.object_type = Some("uri".to_string());
                        prop.target_table = Some(target.to_string());
                    }
                }

                // Enum detection: use Column.enum_values if available
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
            let source_name = uuid_to_name.get(&rel.source_table_id)
                .cloned()
                .unwrap_or_default();
            let target_name = uuid_to_name.get(&rel.target_table_id)
                .cloned()
                .unwrap_or_default();

            if source_name.is_empty() || target_name.is_empty() {
                continue;
            }

            // Use label as property, or generate from table names
            let property = rel.label.clone()
                .unwrap_or_else(|| format!("schema:relatedTo"));

            relationships.push(Relationship {
                from_table: source_name,
                to_table: target_name,
                property,
                inverse: format!("schema:inverseOf"),
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

    /// Resolve table name to (ontology_namespace, class_name) using overrides then conventions.
    fn resolve_table_class(&self, table: &str) -> (String, String) {
        if let Some((ns, cls)) = self.config.table_overrides.get(table) {
            return (ns.clone(), cls.clone());
        }
        table_to_class(table)
            .unwrap_or_else(|| infer_class_from_table(table))
    }

    /// Resolve column name to predicate using overrides then conventions.
    fn resolve_column_predicate(&self, column: &str, _table: &str) -> String {
        if let Some(pred) = self.config.column_overrides.get(column) {
            return pred.clone();
        }
        column_to_predicate(column)
            .unwrap_or_else(|| {
                // Fallback: use schema: prefix + column name
                format!("schema:{}", column)
            })
    }
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/sac/chatmangpt/BusinessOS && cargo test -p bos-core --lib ontology::infer`
Expected: All tests pass (total should be ~15)

- [ ] **Step 5: Commit**

```bash
cd /Users/sac/chatmangpt/BusinessOS
git add bos/core/src/ontology/infer.rs
git commit -m "feat(ontology): implement DataModel-to-MappingConfig inference pipeline"
```

---

## Task 5: Re-export types from lib.rs

**Files:**
- Modify: `bos/core/src/lib.rs`

- [ ] **Step 1: Add re-exports**

Add to the `pub use` block in `lib.rs`:

```rust
pub use ontology::infer::{OntologyInferrer, InferConfig, InferResult, ConfidenceLevel};
```

- [ ] **Step 2: Verify compilation**

Run: `cd /Users/sac/chatmangpt/BusinessOS && cargo build -p bos-core 2>&1 | tail -5`
Expected: Compiles without errors

- [ ] **Step 3: Commit**

```bash
cd /Users/sac/chatmangpt/BusinessOS
git add bos/core/src/lib.rs
git commit -m "feat(ontology): re-export OntologyInferrer types from bos_core"
```

---

## Task 6: Add `infer` verb to CLI ontology noun

**Files:**
- Modify: `bos/cli/src/nouns/ontology.rs`

- [ ] **Step 1: Add the infer verb**

Add to `ontology.rs` after the existing imports and before the `#[noun]` macro:

```rust
#[derive(Serialize)]
#[serde(rename_all = "snake_case")]
pub struct OntologyInferred {
    pub tables_inferred: usize,
    pub properties_inferred: usize,
    pub relationships_inferred: usize,
    pub high_confidence: usize,
    pub medium_confidence: usize,
    pub low_confidence: usize,
    pub output_path: String,
}
```

Add the verb inside the `#[noun("ontology", ...)]` block (after the `execute` verb):

```rust
/// Infer ontology mappings from workspace schema
///
/// # Arguments
/// * `workspace` - Workspace directory containing model.json [default: .]
/// * `output` - Output mapping file path
/// * `confidence` - Minimum confidence threshold (high, medium, low) [default: low]
/// * `conventions` - Custom conventions override file [hide]
/// * `dry_run` - Print mappings without writing file
#[verb("infer")]
fn infer(
    workspace: Option<String>,
    output: String,
    confidence: Option<String>,
    conventions: Option<String>,
    dry_run: bool,
) -> Result<OntologyInferred> {
    let ws = workspace.unwrap_or_else(|| ".".to_string());
    let path = std::path::Path::new(&ws);

    let mut config = match confidence.as_deref() {
        Some("high") => bos_core::InferConfig::high_confidence(),
        Some("medium") => bos_core::InferConfig {
            confidence_threshold: 0.5,
            ..Default::default()
        },
        _ => bos_core::InferConfig::default(),
    };

    // Load custom conventions override if provided
    if let Some(conv_path) = conventions {
        let content = std::fs::read_to_string(&conv_path)
            .map_err(|e| clap_noun_verb::NounVerbError::execution_error(
                format!("Failed to read conventions file: {e}")
            ))?;
        let overrides: serde_json::Value = serde_json::from_str(&content)
            .map_err(|e| clap_noun_verb::NounVerbError::execution_error(
                format!("Failed to parse conventions JSON: {e}")
            ))?;

        if let Some(tables) = overrides.get("tables").and_then(|v| v.as_object()) {
            for (key, val) in tables {
                if let Some(arr) = val.as_array() {
                    if arr.len() == 2 {
                        let ns = arr[0].as_str()
                            .ok_or_else(|| clap_noun_verb::NounVerbError::execution_error(
                                format!("Invalid table override for '{}': first element must be a string", key)
                            ))?;
                        let cls = arr[1].as_str()
                            .ok_or_else(|| clap_noun_verb::NounVerbError::execution_error(
                                format!("Invalid table override for '{}': second element must be a string", key)
                            ))?;
                        config.table_overrides.insert(key.clone(), (ns.to_string(), cls.to_string()));
                    }
                }
            }
        }
        if let Some(columns) = overrides.get("columns").and_then(|v| v.as_object()) {
            for (key, val) in columns {
                if let Some(s) = val.as_str() {
                    config.column_overrides.insert(key.clone(), s.to_string());
                }
            }
        }
    }

    let result = bos_core::OntologyInferrer::infer_from_workspace(path, config)
        .map_err(|e| clap_noun_verb::NounVerbError::execution_error(e.to_string()))?;

    if dry_run {
        let json = serde_json::to_string_pretty(&result.config)
            .map_err(|e| clap_noun_verb::NounVerbError::execution_error(
                format!("Failed to serialize: {e}")
            ))?;
        eprintln!("{}", json);
    } else {
        let json = serde_json::to_string_pretty(&result.config)
            .map_err(|e| clap_noun_verb::NounVerbError::execution_error(
                format!("Failed to serialize: {e}")
            ))?;
        std::fs::write(&output, &json)
            .map_err(|e| clap_noun_verb::NounVerbError::execution_error(
                format!("Failed to write output: {e}")
            ))?;
    }

    Ok(OntologyInferred {
        tables_inferred: result.tables_inferred,
        properties_inferred: result.properties_inferred,
        relationships_inferred: result.relationships_inferred,
        high_confidence: result.high_confidence,
        medium_confidence: result.medium_confidence,
        low_confidence: result.low_confidence,
        output_path: output,
    })
}
```

- [ ] **Step 2: Verify compilation**

Run: `cd /Users/sac/chatmangpt/BusinessOS && cargo build -p bos-cli 2>&1 | tail -10`
Expected: Compiles without errors

- [ ] **Step 3: Test the CLI command against the existing workspace**

```bash
cd /Users/sac/chatmangpt/BusinessOS
./bos/target/release/bos ontology infer --workspace . --output /tmp/test-inferred.json --dry-run 2>&1 | head -50
```

Expected: Prints a valid MappingConfig JSON with 7 tables (if model.json has 7 tables) or fewer

- [ ] **Step 4: Commit**

```bash
cd /Users/sac/chatmangpt/BusinessOS
git add bos/cli/src/nouns/ontology.rs
git commit -m "feat(cli): add 'bos ontology infer' verb with confidence filtering and dry-run"
```

---

## Task 7: Verify end-to-end with existing ontology pipeline

**Files:** None new — verification only

- [ ] **Step 1: Build release binary**

```bash
cd /Users/sac/chatmangpt/BusinessOS && cargo build --release -p bos-cli
```

- [ ] **Step 2: Generate inferred mappings**

```bash
./bos/target/release/bos ontology infer --workspace . --output /tmp/inferred-mappings.json
```

- [ ] **Step 3: Verify the output is valid MappingConfig JSON**

```bash
# Check it's valid JSON
python3 -c "import json; d=json.load(open('/tmp/inferred-mappings.json')); print(f'Tables: {len(d[\"mappings\"])}, Prefixes: {len(d[\"prefixes\"])}')"
```

- [ ] **Step 4: Use inferred mappings with existing construct command**

```bash
./bos/target/release/bos ontology construct --mapping /tmp/inferred-mappings.json --output /tmp/inferred-queries/
ls /tmp/inferred-queries/
```

Expected: .rq files generated for each inferred table

- [ ] **Step 5: Run full test suite**

```bash
cd /Users/sac/chatmangpt/BusinessOS && cargo test -p bos-core -p bos-cli
```

Expected: All tests pass (41 existing + ~15 new infer tests = ~56 total)

- [ ] **Step 6: Commit any fixes if needed**

```bash
git add -A
git commit -m "fix(ontology): address e2e verification issues" || echo "No fixes needed"
```

---

## Task 8: Rebuild release binary and update memory

- [ ] **Step 1: Build final release**

```bash
cd /Users/sac/chatmangpt/BusinessOS && cargo build --release -p bos-cli
```

- [ ] **Step 2: Run final test suite**

```bash
cargo test -p bos-core -p bos-cli 2>&1 | tail -20
```

- [ ] **Step 3: Update auto-memory with new test count and feature status**

Write to memory file the new test baseline and feature completion status.

---

## Summary

| Task | Description | New Tests | Files Modified |
|------|------------|-----------|----------------|
| 1 | Table-to-class conventions + struct | 4 | infer.rs (create), mod.rs |
| 2 | Column-to-predicate + datatype conventions | 7 | infer.rs |
| 3 | Confidence scoring | 6 | infer.rs |
| 4 | Full inference pipeline (DataModel → MappingConfig) | 6 | infer.rs |
| 5 | Re-export types from lib.rs | 0 | lib.rs |
| 6 | CLI `infer` verb | 0 | ontology.rs |
| 7 | E2E verification | 0 | none |
| 8 | Final build + memory update | 0 | memory |
| **Total** | | **~23** | **4 files** |

**New total test count: ~64** (41 existing + 23 new)
