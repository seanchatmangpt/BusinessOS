use clap_noun_verb_macros::{noun, verb};
use clap_noun_verb::Result;
use serde::{Deserialize, Serialize};

// ============================================================================
// Response Types
// ============================================================================

#[derive(Serialize)]
#[serde(rename_all = "snake_case")]
pub struct DomainCreated {
    pub domain_name: String,
    pub domain_id: String,
    pub template_path: String,
    pub governance_level: String,
}

#[derive(Serialize)]
#[serde(rename_all = "snake_case")]
pub struct ContractDefined {
    pub contract_id: String,
    pub dataset_name: String,
    pub domain: String,
    pub dcat_profile: String,
    pub odrl_permissions: Vec<String>,
    pub dqv_dimensions: Vec<String>,
}

#[derive(Serialize)]
#[serde(rename_all = "snake_case")]
pub struct DiscoverResult {
    pub query_id: String,
    pub datasets_found: usize,
    pub results: Vec<DatasetSummary>,
    pub domains_scanned: Vec<String>,
}

#[derive(Serialize)]
#[serde(rename_all = "snake_case")]
pub struct DatasetSummary {
    pub dataset_id: String,
    pub name: String,
    pub domain: String,
    pub owner: String,
    pub quality_score: f32,
    pub records_count: usize,
}

#[derive(Serialize)]
#[serde(rename_all = "snake_case")]
pub struct LineageResult {
    pub entity_id: String,
    pub entity_name: String,
    pub entity_type: String,
    pub upstream: Vec<LineageNode>,
    pub downstream: Vec<LineageNode>,
    pub provenance_triples: usize,
}

#[derive(Serialize)]
#[serde(rename_all = "snake_case")]
pub struct LineageNode {
    pub node_id: String,
    pub node_name: String,
    pub relationship_type: String,
    pub timestamp: String,
}

#[derive(Serialize)]
#[serde(rename_all = "snake_case")]
pub struct QualityMetrics {
    pub dataset_id: String,
    pub dataset_name: String,
    pub completeness: f32,
    pub accuracy: f32,
    pub consistency: f32,
    pub timeliness: f32,
    pub uniqueness: f32,
    pub overall_score: f32,
    pub issues_detected: usize,
    pub last_assessed: String,
}

// ============================================================================
// Input/Configuration Types
// ============================================================================

#[derive(Deserialize)]
pub struct ContractRequest {
    pub dataset_name: String,
    pub domain: String,
    pub owner: String,
    pub description: Option<String>,
}

#[derive(Deserialize)]
pub struct DiscoverQuery {
    pub domain_filter: Option<String>,
    pub quality_threshold: Option<f32>,
    pub owner_filter: Option<String>,
}

// ============================================================================
// Constants for Domain Definitions
// ============================================================================

const DOMAINS: &[&str] = &["Finance", "Operations", "Marketing", "Sales", "HR"];

fn get_domain_governance(domain: &str) -> String {
    match domain {
        "Finance" => "SOX-compliant, quarterly audits".to_string(),
        "Operations" => "Process-driven, real-time monitoring".to_string(),
        "Marketing" => "Campaign-focused, attribution tracking".to_string(),
        "Sales" => "Revenue-aligned, pipeline transparency".to_string(),
        "HR" => "Privacy-critical, confidential data".to_string(),
        _ => "Standard governance".to_string(),
    }
}

// ============================================================================
// NOUN: mesh
// ============================================================================

#[noun("mesh", "Data mesh domain governance commands")]

/// Create a new data mesh domain
///
/// # Arguments
/// * `name` - Domain name (Finance, Operations, Marketing, Sales, or HR)
/// * `description` - Domain description [hide]
/// * `owner` - Domain owner email [hide]
#[verb("domain", "create")]
fn domain_create(
    name: String,
    description: Option<String>,
    owner: Option<String>,
) -> Result<DomainCreated> {
    // Validate domain name
    let valid_domain = DOMAINS.iter().any(|d| {
        d.to_lowercase() == name.to_lowercase()
    });

    if !valid_domain {
        return Err(clap_noun_verb::NounVerbError::execution_error(
            format!(
                "Invalid domain '{}'. Valid domains are: {}",
                name,
                DOMAINS.join(", ")
            ),
        ));
    }

    let domain_id = format!(
        "{}-{}",
        name.to_lowercase(),
        chrono::Local::now().format("%Y%m%d%H%M%S")
    );

    let template_path = format!("./domains/{}", name.to_lowercase());
    let governance_level = get_domain_governance(&name);

    Ok(DomainCreated {
        domain_name: name,
        domain_id,
        template_path,
        governance_level,
    })
}

/// Define a data contract (DCAT + ODRL + DQV)
///
/// # Arguments
/// * `dataset` - Dataset name
/// * `domain` - Domain name
/// * `owner` - Dataset owner
/// * `dcat_profile` - DCAT profile (full, lite) [default: lite]
/// * `dqv_dimensions` - DQV dimensions (csv) [default: accuracy,completeness,timeliness]
#[verb("contract", "define")]
fn contract_define(
    dataset: String,
    domain: String,
    owner: String,
    dcat_profile: Option<String>,
    dqv_dimensions: Option<String>,
) -> Result<ContractDefined> {
    let profile = dcat_profile.unwrap_or_else(|| "lite".to_string());
    let dims_str = dqv_dimensions.unwrap_or_else(|| {
        "accuracy,completeness,timeliness".to_string()
    });

    let dqv_dimensions: Vec<String> = dims_str
        .split(',')
        .map(|s| s.trim().to_string())
        .collect();

    let contract_id = format!(
        "contract-{}-{}",
        dataset.to_lowercase().replace(' ', "-"),
        chrono::Local::now().format("%Y%m%d%H%M%S")
    );

    let odrl_permissions = vec![
        "odrl:use".to_string(),
        "odrl:distribute".to_string(),
        "odrl:derive".to_string(),
    ];

    Ok(ContractDefined {
        contract_id,
        dataset_name: dataset,
        domain,
        dcat_profile: profile,
        odrl_permissions,
        dqv_dimensions,
    })
}

/// Discover datasets in data mesh
///
/// # Arguments
/// * `domain` - Optional domain filter
/// * `quality_threshold` - Minimum quality score (0.0-1.0) [hide]
/// * `owner` - Optional owner filter [hide]
#[verb("discover")]
fn discover(
    domain: Option<String>,
    quality_threshold: Option<String>,
    owner: Option<String>,
) -> Result<DiscoverResult> {
    let query_id = format!("discovery-{}", chrono::Local::now().format("%Y%m%d%H%M%S"));
    let threshold = quality_threshold
        .and_then(|t| t.parse::<f32>().ok())
        .unwrap_or(0.7);

    let results = perform_discovery(&domain, threshold, owner.as_deref());
    let domains_scanned = get_domains_to_scan(&domain);

    Ok(DiscoverResult {
        query_id,
        datasets_found: results.len(),
        results,
        domains_scanned,
    })
}

/// Trace data lineage (PROV-O provenance)
///
/// # Arguments
/// * `entity` - Entity ID or name to trace
/// * `depth` - Lineage depth (1-5) [default: 2]
#[verb("lineage")]
fn lineage(entity: String, depth: Option<String>) -> Result<LineageResult> {
    let _depth = depth.and_then(|d| d.parse::<usize>().ok()).unwrap_or(2);
    let entity_id = format!("entity-{}", entity.to_lowercase().replace(' ', "-"));
    let upstream = build_upstream_lineage();
    let downstream = build_downstream_lineage();

    Ok(LineageResult {
        entity_id,
        entity_name: entity,
        entity_type: "Dataset".to_string(),
        upstream,
        downstream,
        provenance_triples: 12,
    })
}

/// Assess dataset quality (DQV metrics)
///
/// # Arguments
/// * `dataset` - Dataset ID or name
/// * `dimensions` - Quality dimensions to assess (csv) [hide]
#[verb("quality")]
fn quality(dataset: String, dimensions: Option<String>) -> Result<QualityMetrics> {
    let _dims = dimensions;
    let metrics = calculate_quality_metrics();

    Ok(QualityMetrics {
        dataset_id: format!("dataset-{}", dataset.to_lowercase().replace(' ', "-")),
        dataset_name: dataset,
        completeness: metrics.0,
        accuracy: metrics.1,
        consistency: metrics.2,
        timeliness: metrics.3,
        uniqueness: metrics.4,
        overall_score: metrics.5,
        issues_detected: if metrics.5 < 0.85 { 5 } else { 0 },
        last_assessed: chrono::Local::now().to_rfc3339(),
    })
}

// ============================================================================
// Helper Functions
// ============================================================================

fn perform_discovery(
    domain_filter: &Option<String>,
    threshold: f32,
    owner_filter: Option<&str>,
) -> Vec<DatasetSummary> {
    let domains = get_domains_to_scan(domain_filter);
    let mut results = Vec::new();

    for d in domains {
        let datasets = simulate_datasets_for_domain(&d, threshold);
        results.extend(datasets);
    }

    if let Some(owner) = owner_filter {
        results.retain(|ds| ds.owner.to_lowercase().contains(&owner.to_lowercase()));
    }

    results
}

fn get_domains_to_scan(domain_filter: &Option<String>) -> Vec<String> {
    match domain_filter {
        Some(d) => vec![d.clone()],
        None => DOMAINS.iter().map(|s| s.to_string()).collect(),
    }
}

fn simulate_datasets_for_domain(domain: &str, threshold: f32) -> Vec<DatasetSummary> {
    let datasets = match domain {
        "Finance" => vec![
            ("GL Transactions", "finance-gl@company.com", 0.98, 500000),
            ("AR Aging", "finance-ar@company.com", 0.91, 100000),
        ],
        "Operations" => vec![
            ("Supply Chain Events", "ops-supply@company.com", 0.87, 1000000),
            ("Inventory Snapshots", "ops-inv@company.com", 0.95, 250000),
        ],
        "Marketing" => vec![
            ("Campaign Performance", "marketing-campaigns@company.com", 0.89, 50000),
            ("Customer Engagement", "marketing-cust@company.com", 0.84, 200000),
        ],
        "Sales" => vec![
            ("Pipeline Opportunities", "sales-pipeline@company.com", 0.93, 15000),
            ("Deal Velocity", "sales-deals@company.com", 0.90, 8000),
        ],
        "HR" => vec![
            ("Employee Roster", "hr-roster@company.com", 0.99, 5000),
            ("Payroll Records", "hr-payroll@company.com", 0.97, 50000),
        ],
        _ => vec![],
    };

    datasets
        .into_iter()
        .filter(|(_, _, score, _)| *score >= threshold)
        .map(|(name, owner, score, records)| DatasetSummary {
            dataset_id: format!("dataset-{}", name.to_lowercase().replace(' ', "-")),
            name: name.to_string(),
            domain: domain.to_string(),
            owner: owner.to_string(),
            quality_score: score,
            records_count: records,
        })
        .collect()
}

fn build_upstream_lineage() -> Vec<LineageNode> {
    vec![
        LineageNode {
            node_id: "source-erp".to_string(),
            node_name: "ERP System".to_string(),
            relationship_type: "prov:wasGeneratedBy".to_string(),
            timestamp: "2026-03-24T10:00:00Z".to_string(),
        },
        LineageNode {
            node_id: "transform-agg".to_string(),
            node_name: "Daily Aggregation".to_string(),
            relationship_type: "prov:wasDerivedFrom".to_string(),
            timestamp: "2026-03-24T12:00:00Z".to_string(),
        },
    ]
}

fn build_downstream_lineage() -> Vec<LineageNode> {
    vec![LineageNode {
        node_id: "report-finance".to_string(),
        node_name: "Finance Dashboard".to_string(),
        relationship_type: "prov:wasUsedBy".to_string(),
        timestamp: "2026-03-24T14:00:00Z".to_string(),
    }]
}

fn calculate_quality_metrics() -> (f32, f32, f32, f32, f32, f32) {
    let completeness = 0.96;
    let accuracy = 0.92;
    let consistency = 0.98;
    let timeliness = 0.88;
    let uniqueness = 0.99;
    let overall = (completeness + accuracy + consistency + timeliness + uniqueness) / 5.0;
    (completeness, accuracy, consistency, timeliness, uniqueness, overall)
}
