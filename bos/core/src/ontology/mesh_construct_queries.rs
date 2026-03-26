/// Data Mesh SPARQL CONSTRUCT Queries
///
/// SPARQL CONSTRUCT queries for generating RDF triples from relational data
/// across 5 data mesh domains. These queries use PROV-O, DCAT, ODRL, and DQV
/// ontologies for data governance, lineage, contracts, and quality metrics.

pub struct MeshConstructQueries;

impl MeshConstructQueries {
    pub fn get_query(domain: &str) -> Option<&'static str> {
        match domain.to_lowercase().as_str() {
            "finance" => Some(FINANCE_CONSTRUCT_QUERY),
            "operations" => Some(OPERATIONS_CONSTRUCT_QUERY),
            "marketing" => Some(MARKETING_CONSTRUCT_QUERY),
            "sales" => Some(SALES_CONSTRUCT_QUERY),
            "hr" => Some(HR_CONSTRUCT_QUERY),
            _ => None,
        }
    }

    pub fn all_domains() -> &'static [&'static str] {
        &["finance", "operations", "marketing", "sales", "hr"]
    }
}

/// Finance Domain CONSTRUCT Query
///
/// Generates RDF for General Ledger, Accounts Receivable, Accounts Payable
/// with SOX compliance governance, revenue recognition rules, and audit lineage.
pub const FINANCE_CONSTRUCT_QUERY: &str = r#"
PREFIX rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#>
PREFIX rdfs: <http://www.w3.org/2000/01/rdf-schema#>
PREFIX xsd: <http://www.w3.org/2001/XMLSchema#>
PREFIX dcat: <http://www.w3.org/ns/dcat#>
PREFIX dct: <http://purl.org/dc/terms/>
PREFIX prov: <http://www.w3.org/ns/prov#>
PREFIX odrl: <http://www.w3.org/ns/odrl/2/>
PREFIX dqv: <http://www.w3.org/ns/dqv#>
PREFIX mesh: <http://example.org/mesh/finance/>
PREFIX luc: <http://www.ontos.com/luc#>

CONSTRUCT {
  ?dataset a dcat:Dataset ;
    rdfs:label ?dataset_name ;
    dct:description ?dataset_desc ;
    dcat:theme mesh:FinanceTheme ;
    dcat:accrualPeriodicity <http://purl.org/ckan/ns/frequency#daily> ;
    dct:issued ?created_date ;
    dct:modified ?modified_date ;
    dcat:contactPoint [ rdfs:label ?owner_name ] ;
    dcat:landing [ rdfs:label ?landing_page ] ;
    prov:wasGeneratedBy ?activity ;
    prov:wasDerivedFrom ?source ;
    dqv:hasQualityMeasurement ?quality_measurement ;
    odrl:hasPolicy [ odrl:permission ?permission ] .

  ?activity a prov:Activity ;
    rdfs:label ?activity_name ;
    prov:startedAtTime ?start_time ;
    prov:endedAtTime ?end_time ;
    prov:wasAssociatedWith [ rdfs:label "SOX Auditor" ] ;
    prov:hadPlan [ rdfs:label "Revenue Recognition Rules" ] .

  ?source a dcat:Dataset ;
    rdfs:label ?source_name ;
    dcat:theme mesh:SourceTheme .

  ?quality_measurement a dqv:QualityMeasurement ;
    dqv:isMeasurementOf [ rdfs:label "Completeness" ] ;
    dqv:value ?completeness_value ;
    dct:date ?quality_date .

  ?permission odrl:action odrl:read ;
    odrl:assignee [ rdfs:label ?consumer_role ] .
}
WHERE {
  BIND(IRI(CONCAT("http://example.org/dataset/finance/", ?dataset_id)) AS ?dataset)
  BIND(IRI(CONCAT("http://example.org/activity/finance/", ?activity_id)) AS ?activity)
  BIND(IRI(CONCAT("http://example.org/source/", ?source_id)) AS ?source)
  BIND(IRI(CONCAT("http://example.org/measurement/", ?measurement_id)) AS ?quality_measurement)

  # Revenue recognition dataset
  VALUES (?dataset_id ?dataset_name ?dataset_desc ?owner_name ?landing_page) {
    ("gl-transactions" "General Ledger Transactions" "Daily GL balances by account" "finance-gl@company.com" "https://fin-dashboard.internal/gl")
    ("ar-aging" "Accounts Receivable Aging" "Customer AR aging buckets" "finance-ar@company.com" "https://fin-dashboard.internal/ar")
    ("ap-schedule" "Accounts Payable Schedule" "Vendor payment schedule" "finance-ap@company.com" "https://fin-dashboard.internal/ap")
  }

  VALUES (?activity_id ?activity_name ?start_time ?end_time) {
    ("activity-1" "End of Day GL Close" "2026-03-24T17:00:00Z"^^xsd:dateTime "2026-03-24T18:00:00Z"^^xsd:dateTime)
    ("activity-2" "AR Aging Calculation" "2026-03-24T02:00:00Z"^^xsd:dateTime "2026-03-24T02:30:00Z"^^xsd:dateTime)
  }

  VALUES (?source_id ?source_name) {
    ("erp-system" "ERP GL Module")
    ("billing-system" "Billing Engine")
  }

  VALUES (?measurement_id ?completeness_value ?quality_date) {
    ("measure-1" "0.98"^^xsd:decimal "2026-03-24T10:00:00Z"^^xsd:dateTime)
  }

  VALUES (?consumer_role) { ("AccountsAnalyst") ("CFO") ("Auditor") }

  BIND(now() AS ?created_date)
  BIND(now() AS ?modified_date)
  BIND(now() AS ?quality_date)
}
"#;

/// Operations Domain CONSTRUCT Query
///
/// Generates RDF for Supply Chain, Inventory, Manufacturing processes
/// with real-time monitoring, SLA tracking, and process improvement lineage.
pub const OPERATIONS_CONSTRUCT_QUERY: &str = r#"
PREFIX rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#>
PREFIX rdfs: <http://www.w3.org/2000/01/rdf-schema#>
PREFIX xsd: <http://www.w3.org/2001/XMLSchema#>
PREFIX dcat: <http://www.w3.org/ns/dcat#>
PREFIX dct: <http://purl.org/dc/terms/>
PREFIX prov: <http://www.w3.org/ns/prov#>
PREFIX odrl: <http://www.w3.org/ns/odrl/2/>
PREFIX dqv: <http://www.w3.org/ns/dqv#>
PREFIX mesh: <http://example.org/mesh/operations/>

CONSTRUCT {
  ?dataset a dcat:Dataset ;
    rdfs:label ?dataset_name ;
    dct:description ?dataset_desc ;
    dcat:theme mesh:OperationsTheme ;
    dcat:accrualPeriodicity <http://purl.org/ckan/ns/frequency#hourly> ;
    dct:issued ?created_date ;
    prov:wasGeneratedBy ?process ;
    dqv:hasQualityMeasurement ?quality_measurement ;
    dqv:hasQualityAnnotation [ dqv:oa "SLA compliance" ] ;
    odrl:hasPolicy [ odrl:permission [ odrl:action odrl:use ] ] .

  ?process a prov:Process ;
    rdfs:label ?process_name ;
    prov:hadInputs ?input ;
    prov:hadOutputs ?output ;
    prov:wasPlan [ rdfs:label "Lean Manufacturing" ] ;
    prov:endedAtTime ?process_end .

  ?quality_measurement a dqv:QualityMeasurement ;
    dqv:isMeasurementOf [ rdfs:label "Timeliness" ] ;
    dqv:value ?timeliness_value ;
    dqv:value [ rdfs:label "Accuracy" ] .
}
WHERE {
  BIND(IRI(CONCAT("http://example.org/dataset/ops/", ?dataset_id)) AS ?dataset)
  BIND(IRI(CONCAT("http://example.org/process/", ?process_id)) AS ?process)
  BIND(IRI(CONCAT("http://example.org/input/", ?input_id)) AS ?input)
  BIND(IRI(CONCAT("http://example.org/output/", ?output_id)) AS ?output)
  BIND(IRI(CONCAT("http://example.org/measurement/", ?measurement_id)) AS ?quality_measurement)

  VALUES (?dataset_id ?dataset_name ?dataset_desc) {
    ("supply-events" "Supply Chain Events" "Real-time shipment tracking")
    ("inventory-snapshots" "Inventory Snapshots" "Hourly inventory levels")
    ("production-metrics" "Production Metrics" "Manufacturing KPIs")
  }

  VALUES (?process_id ?process_name ?process_end) {
    ("inbound-receiving" "Inbound Receiving" "2026-03-24T15:30:00Z"^^xsd:dateTime)
    ("inventory-count" "Inventory Count Job" "2026-03-24T01:00:00Z"^^xsd:dateTime)
  }

  VALUES (?input_id ?output_id) {
    ("po-system" "goods-received")
    ("wms-system" "qty-available")
  }

  VALUES (?measurement_id ?timeliness_value) {
    ("measure-1" "0.95"^^xsd:decimal)
  }

  BIND(now() AS ?created_date)
}
"#;

/// Marketing Domain CONSTRUCT Query
///
/// Generates RDF for Campaign Performance, Customer Engagement, Attribution
/// with attribution journey tracking, A/B test results, and campaign lineage.
pub const MARKETING_CONSTRUCT_QUERY: &str = r#"
PREFIX rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#>
PREFIX rdfs: <http://www.w3.org/2000/01/rdf-schema#>
PREFIX xsd: <http://www.w3.org/2001/XMLSchema#>
PREFIX dcat: <http://www.w3.org/ns/dcat#>
PREFIX dct: <http://purl.org/dc/terms/>
PREFIX prov: <http://www.w3.org/ns/prov#>
PREFIX odrl: <http://www.w3.org/ns/odrl/2/>
PREFIX dqv: <http://www.w3.org/ns/dqv#>
PREFIX mesh: <http://example.org/mesh/marketing/>
PREFIX foaf: <http://xmlns.com/foaf/0.1/>

CONSTRUCT {
  ?dataset a dcat:Dataset ;
    rdfs:label ?dataset_name ;
    dct:description ?dataset_desc ;
    dcat:theme mesh:MarketingTheme ;
    dcat:accrualPeriodicity <http://purl.org/ckan/ns/frequency#daily> ;
    prov:wasGeneratedBy ?campaign ;
    prov:wasDerivedFrom ?attribution_source ;
    dqv:hasQualityMeasurement ?quality_measurement ;
    odrl:hasPolicy [
      odrl:permission [
        odrl:action odrl:use ;
        odrl:target [ rdfs:label ?dataset_name ]
      ]
    ] .

  ?campaign a prov:Activity ;
    rdfs:label ?campaign_name ;
    prov:startedAtTime ?campaign_start ;
    prov:endedAtTime ?campaign_end ;
    prov:wasAssociatedWith [ a foaf:Person ; rdfs:label "Campaign Manager" ] ;
    prov:hadPlan [ rdfs:label "Attribution Model: Multi-touch" ] .

  ?quality_measurement a dqv:QualityMeasurement ;
    dqv:isMeasurementOf [ rdfs:label "Uniqueness" ] ;
    dqv:value ?uniqueness_value .
}
WHERE {
  BIND(IRI(CONCAT("http://example.org/dataset/marketing/", ?dataset_id)) AS ?dataset)
  BIND(IRI(CONCAT("http://example.org/campaign/", ?campaign_id)) AS ?campaign)
  BIND(IRI(CONCAT("http://example.org/source/attribution/", ?source_id)) AS ?attribution_source)
  BIND(IRI(CONCAT("http://example.org/measurement/", ?measurement_id)) AS ?quality_measurement)

  VALUES (?dataset_id ?dataset_name ?dataset_desc) {
    ("campaign-perf" "Campaign Performance" "Daily campaign metrics and ROI")
    ("customer-journey" "Customer Journey" "Multi-touch attribution data")
    ("segment-scores" "Segment Scores" "Customer segment engagement scores")
  }

  VALUES (?campaign_id ?campaign_name ?campaign_start ?campaign_end) {
    ("q1-email-campaign" "Q1 Email Campaign" "2026-01-01T00:00:00Z"^^xsd:dateTime "2026-03-31T23:59:59Z"^^xsd:dateTime)
    ("spring-promo" "Spring Promotion" "2026-03-15T00:00:00Z"^^xsd:dateTime "2026-04-15T23:59:59Z"^^xsd:dateTime)
  }

  VALUES (?source_id) { ("web-analytics") ("crm-system") }

  VALUES (?measurement_id ?uniqueness_value) {
    ("measure-1" "0.99"^^xsd:decimal)
  }
}
"#;

/// Sales Domain CONSTRUCT Query
///
/// Generates RDF for Pipeline Opportunities, Deal Velocity, Revenue Forecasting
/// with opportunity lineage, deal stage progression, and forecast accuracy tracking.
pub const SALES_CONSTRUCT_QUERY: &str = r#"
PREFIX rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#>
PREFIX rdfs: <http://www.w3.org/2000/01/rdf-schema#>
PREFIX xsd: <http://www.w3.org/2001/XMLSchema#>
PREFIX dcat: <http://www.w3.org/ns/dcat#>
PREFIX dct: <http://purl.org/dc/terms/>
PREFIX prov: <http://www.w3.org/ns/prov#>
PREFIX odrl: <http://www.w3.org/ns/odrl/2/>
PREFIX dqv: <http://www.w3.org/ns/dqv#>
PREFIX mesh: <http://example.org/mesh/sales/>
PREFIX foaf: <http://xmlns.com/foaf/0.1/>

CONSTRUCT {
  ?dataset a dcat:Dataset ;
    rdfs:label ?dataset_name ;
    dct:description ?dataset_desc ;
    dcat:theme mesh:SalesTheme ;
    dcat:accrualPeriodicity <http://purl.org/ckan/ns/frequency#daily> ;
    dct:issued ?created_date ;
    prov:wasGeneratedBy ?sales_process ;
    prov:wasDerivedFrom [ rdfs:label "CRM System" ] ;
    dqv:hasQualityMeasurement ?accuracy_measurement ;
    odrl:hasPolicy [
      odrl:permission [
        odrl:action odrl:use ;
        odrl:assignee [ rdfs:label ?recipient_role ]
      ]
    ] .

  ?sales_process a prov:Activity ;
    rdfs:label ?process_name ;
    prov:wasAssociatedWith [ a foaf:Person ; rdfs:label "Sales Operations" ] ;
    prov:hadPlan [ rdfs:label "Sales Methodology: MEDDIC" ] .

  ?accuracy_measurement a dqv:QualityMeasurement ;
    dqv:isMeasurementOf [ rdfs:label "Accuracy" ] ;
    dqv:value ?accuracy_value ;
    dct:date ?quality_date .
}
WHERE {
  BIND(IRI(CONCAT("http://example.org/dataset/sales/", ?dataset_id)) AS ?dataset)
  BIND(IRI(CONCAT("http://example.org/process/", ?process_id)) AS ?sales_process)
  BIND(IRI(CONCAT("http://example.org/measurement/", ?measurement_id)) AS ?accuracy_measurement)

  VALUES (?dataset_id ?dataset_name ?dataset_desc) {
    ("pipeline-oppty" "Pipeline Opportunities" "Active sales opportunities by stage")
    ("deal-velocity" "Deal Velocity" "Days in stage progression metrics")
    ("forecast-accuracy" "Forecast Accuracy" "Revenue forecast vs actual")
  }

  VALUES (?process_id ?process_name) {
    ("opp-qualification" "Opportunity Qualification")
    ("deal-closure" "Deal Closure Process")
  }

  VALUES (?recipient_role) { ("SalesRep") ("SalesManager") ("VP Sales") }

  VALUES (?measurement_id ?accuracy_value) {
    ("measure-1" "0.92"^^xsd:decimal)
  }

  BIND(now() AS ?created_date)
  BIND(now() AS ?quality_date)
}
"#;

/// HR Domain CONSTRUCT Query
///
/// Generates RDF for Employee Roster, Payroll, Benefits, Compensation
/// with privacy-critical governance, confidentiality enforcement, and compliance tracking.
pub const HR_CONSTRUCT_QUERY: &str = r#"
PREFIX rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#>
PREFIX rdfs: <http://www.w3.org/2000/01/rdf-schema#>
PREFIX xsd: <http://www.w3.org/2001/XMLSchema#>
PREFIX dcat: <http://www.w3.org/ns/dcat#>
PREFIX dct: <http://purl.org/dc/terms/>
PREFIX prov: <http://www.w3.org/ns/prov#>
PREFIX odrl: <http://www.w3.org/ns/odrl/2/>
PREFIX dqv: <http://www.w3.org/ns/dqv#>
PREFIX mesh: <http://example.org/mesh/hr/>

CONSTRUCT {
  ?dataset a dcat:Dataset ;
    rdfs:label ?dataset_name ;
    dct:description ?dataset_desc ;
    dcat:theme mesh:HRTheme ;
    dcat:accrualPeriodicity <http://purl.org/ckan/ns/frequency#monthly> ;
    dcat:accessLevel <http://purl.org/ckan/ns/accessLevel#restricted> ;
    prov:wasGeneratedBy ?payroll_process ;
    dqv:hasQualityMeasurement ?completeness_measurement ;
    odrl:hasPolicy [
      odrl:permission [
        odrl:action odrl:read ;
        odrl:assignee [ rdfs:label ?authorized_role ] ;
        odrl:constraint [ odrl:unit "monthly" ]
      ] ;
      odrl:prohibition [
        odrl:action odrl:distribute ;
        odrl:reason "GDPR Confidentiality"
      ]
    ] .

  ?payroll_process a prov:Activity ;
    rdfs:label ?payroll_activity_name ;
    prov:startedAtTime ?payroll_start ;
    prov:endedAtTime ?payroll_end ;
    prov:wasAssociatedWith [ rdfs:label "Payroll Administrator" ] ;
    prov:hadPlan [ rdfs:label "Bi-weekly Payroll Processing" ] .

  ?completeness_measurement a dqv:QualityMeasurement ;
    dqv:isMeasurementOf [ rdfs:label "Completeness" ] ;
    dqv:value ?completeness_value .
}
WHERE {
  BIND(IRI(CONCAT("http://example.org/dataset/hr/", ?dataset_id)) AS ?dataset)
  BIND(IRI(CONCAT("http://example.org/process/", ?process_id)) AS ?payroll_process)
  BIND(IRI(CONCAT("http://example.org/measurement/", ?measurement_id)) AS ?completeness_measurement)

  VALUES (?dataset_id ?dataset_name ?dataset_desc) {
    ("employee-roster" "Employee Roster" "Current employee directory and profiles")
    ("payroll-records" "Payroll Records" "Confidential payroll and compensation data")
    ("benefits-enrollment" "Benefits Enrollment" "Employee benefits selections")
  }

  VALUES (?process_id ?payroll_activity_name ?payroll_start ?payroll_end) {
    ("payroll-p1" "Payroll Processing Run 1" "2026-03-06T09:00:00Z"^^xsd:dateTime "2026-03-06T17:00:00Z"^^xsd:dateTime)
    ("payroll-p2" "Payroll Processing Run 2" "2026-03-20T09:00:00Z"^^xsd:dateTime "2026-03-20T17:00:00Z"^^xsd:dateTime)
  }

  VALUES (?authorized_role) { ("HR Manager") ("Finance Manager") ("CEO") }

  VALUES (?measurement_id ?completeness_value) {
    ("measure-1" "0.99"^^xsd:decimal)
  }
}
"#;

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_all_queries_present() {
        let domains = MeshConstructQueries::all_domains();
        assert_eq!(domains.len(), 5);
        assert!(domains.contains(&"finance"));
        assert!(domains.contains(&"operations"));
        assert!(domains.contains(&"marketing"));
        assert!(domains.contains(&"sales"));
        assert!(domains.contains(&"hr"));
    }

    #[test]
    fn test_finance_query_has_required_prefixes() {
        assert!(FINANCE_CONSTRUCT_QUERY.contains("PREFIX prov:"));
        assert!(FINANCE_CONSTRUCT_QUERY.contains("PREFIX dcat:"));
        assert!(FINANCE_CONSTRUCT_QUERY.contains("PREFIX dqv:"));
        assert!(FINANCE_CONSTRUCT_QUERY.contains("PREFIX odrl:"));
    }

    #[test]
    fn test_operations_query_has_sla_reference() {
        assert!(OPERATIONS_CONSTRUCT_QUERY.contains("SLA"));
        assert!(OPERATIONS_CONSTRUCT_QUERY.contains("Lean Manufacturing"));
    }

    #[test]
    fn test_marketing_query_has_attribution() {
        assert!(MARKETING_CONSTRUCT_QUERY.contains("Attribution"));
        assert!(MARKETING_CONSTRUCT_QUERY.contains("Multi-touch"));
    }

    #[test]
    fn test_sales_query_has_meddic() {
        assert!(SALES_CONSTRUCT_QUERY.contains("MEDDIC"));
    }

    #[test]
    fn test_hr_query_has_gdpr() {
        assert!(HR_CONSTRUCT_QUERY.contains("GDPR"));
        assert!(HR_CONSTRUCT_QUERY.contains("restricted"));
    }

    #[test]
    fn test_all_queries_valid_construct() {
        for domain in MeshConstructQueries::all_domains() {
            if let Some(query) = MeshConstructQueries::get_query(domain) {
                assert!(
                    query.starts_with("PREFIX") || query.starts_with("\nPREFIX"),
                    "Query for domain {} missing PREFIX declarations",
                    domain
                );
                assert!(
                    query.contains("CONSTRUCT"),
                    "Query for domain {} missing CONSTRUCT keyword",
                    domain
                );
                assert!(
                    query.contains("WHERE"),
                    "Query for domain {} missing WHERE clause",
                    domain
                );
            }
        }
    }
}
