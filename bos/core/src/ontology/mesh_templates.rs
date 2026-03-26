/// Data Mesh Domain Templates
///
/// YAML-based templates for initializing data mesh domains with
/// governance policies, quality rules, and access control definitions.

use serde::{Deserialize, Serialize};
use std::collections::HashMap;

#[derive(Debug, Serialize, Deserialize)]
pub struct DomainTemplate {
    pub name: String,
    pub description: String,
    pub governance_level: String,
    pub steward: String,
    pub data_classification: String,
    pub sla: SLA,
    pub quality_rules: Vec<QualityRule>,
    pub data_contracts: Vec<DataContract>,
    pub access_policies: Vec<AccessPolicy>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct SLA {
    pub availability_percentage: f32,
    pub max_latency_minutes: u32,
    pub rto_hours: u32,
    pub rpo_minutes: u32,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct QualityRule {
    pub name: String,
    pub metric: String,
    pub threshold: f32,
    pub dimension: String, // completeness, accuracy, consistency, timeliness, uniqueness
}

#[derive(Debug, Serialize, Deserialize)]
pub struct DataContract {
    pub name: String,
    pub datasets: Vec<String>,
    pub owner: String,
    pub retention_days: u32,
    pub refresh_frequency: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct AccessPolicy {
    pub role: String,
    pub datasets: Vec<String>,
    pub operations: Vec<String>, // read, write, delete
    pub conditions: String,
}

impl DomainTemplate {
    pub fn finance() -> Self {
        DomainTemplate {
            name: "Finance".to_string(),
            description: "Financial data domain - GL, AR, AP, Revenue Recognition".to_string(),
            governance_level: "SOX-compliant, quarterly audits".to_string(),
            steward: "Chief Financial Officer".to_string(),
            data_classification: "Confidential".to_string(),
            sla: SLA {
                availability_percentage: 99.9,
                max_latency_minutes: 5,
                rto_hours: 1,
                rpo_minutes: 15,
            },
            quality_rules: vec![
                QualityRule {
                    name: "GL Completeness".to_string(),
                    metric: "row_count".to_string(),
                    threshold: 0.95,
                    dimension: "completeness".to_string(),
                },
                QualityRule {
                    name: "GL Accuracy".to_string(),
                    metric: "reconciliation_variance".to_string(),
                    threshold: 0.001,
                    dimension: "accuracy".to_string(),
                },
            ],
            data_contracts: vec![
                DataContract {
                    name: "GL Transaction Feed".to_string(),
                    datasets: vec!["gl_transactions".to_string()],
                    owner: "finance-gl@company.com".to_string(),
                    retention_days: 2555, // 7 years for audit
                    refresh_frequency: "hourly".to_string(),
                },
                DataContract {
                    name: "AR Aging Report".to_string(),
                    datasets: vec!["ar_aging".to_string()],
                    owner: "finance-ar@company.com".to_string(),
                    retention_days: 1825, // 5 years
                    refresh_frequency: "daily".to_string(),
                },
            ],
            access_policies: vec![
                AccessPolicy {
                    role: "Accountant".to_string(),
                    datasets: vec!["gl_transactions".to_string(), "ar_aging".to_string()],
                    operations: vec!["read".to_string()],
                    conditions: "within_business_hours AND requires_mfa".to_string(),
                },
                AccessPolicy {
                    role: "CFO".to_string(),
                    datasets: vec!["gl_transactions".to_string(), "ar_aging".to_string(), "ap_schedule".to_string()],
                    operations: vec!["read".to_string(), "write".to_string()],
                    conditions: "unrestricted".to_string(),
                },
                AccessPolicy {
                    role: "Auditor".to_string(),
                    datasets: vec!["gl_transactions".to_string()],
                    operations: vec!["read".to_string()],
                    conditions: "read_only AND audit_logging".to_string(),
                },
            ],
        }
    }

    pub fn operations() -> Self {
        DomainTemplate {
            name: "Operations".to_string(),
            description: "Operational data domain - Supply Chain, Inventory, Manufacturing".to_string(),
            governance_level: "Process-driven, real-time monitoring".to_string(),
            steward: "Chief Operations Officer".to_string(),
            data_classification: "Internal".to_string(),
            sla: SLA {
                availability_percentage: 99.5,
                max_latency_minutes: 1,
                rto_hours: 4,
                rpo_minutes: 30,
            },
            quality_rules: vec![
                QualityRule {
                    name: "Inventory Timeliness".to_string(),
                    metric: "update_frequency_minutes".to_string(),
                    threshold: 60.0,
                    dimension: "timeliness".to_string(),
                },
                QualityRule {
                    name: "Supply Chain Accuracy".to_string(),
                    metric: "shipment_confirmation_rate".to_string(),
                    threshold: 0.98,
                    dimension: "accuracy".to_string(),
                },
            ],
            data_contracts: vec![
                DataContract {
                    name: "Supply Chain Events Feed".to_string(),
                    datasets: vec!["supply_events".to_string()],
                    owner: "ops-supply@company.com".to_string(),
                    retention_days: 730, // 2 years
                    refresh_frequency: "real-time".to_string(),
                },
                DataContract {
                    name: "Inventory Snapshots".to_string(),
                    datasets: vec!["inventory_snapshots".to_string()],
                    owner: "ops-inv@company.com".to_string(),
                    retention_days: 365, // 1 year
                    refresh_frequency: "hourly".to_string(),
                },
            ],
            access_policies: vec![
                AccessPolicy {
                    role: "OperationsTeam".to_string(),
                    datasets: vec!["supply_events".to_string(), "inventory_snapshots".to_string()],
                    operations: vec!["read".to_string(), "write".to_string()],
                    conditions: "real_time_dashboard".to_string(),
                },
            ],
        }
    }

    pub fn marketing() -> Self {
        DomainTemplate {
            name: "Marketing".to_string(),
            description: "Marketing data domain - Campaigns, Customer Engagement, Attribution".to_string(),
            governance_level: "Campaign-focused, attribution tracking".to_string(),
            steward: "Chief Marketing Officer".to_string(),
            data_classification: "Internal".to_string(),
            sla: SLA {
                availability_percentage: 99.0,
                max_latency_minutes: 30,
                rto_hours: 8,
                rpo_minutes: 120,
            },
            quality_rules: vec![
                QualityRule {
                    name: "Campaign Data Uniqueness".to_string(),
                    metric: "duplicate_customer_records".to_string(),
                    threshold: 0.01,
                    dimension: "uniqueness".to_string(),
                },
                QualityRule {
                    name: "Attribution Consistency".to_string(),
                    metric: "model_agreement_percentage".to_string(),
                    threshold: 0.85,
                    dimension: "consistency".to_string(),
                },
            ],
            data_contracts: vec![
                DataContract {
                    name: "Campaign Performance Data".to_string(),
                    datasets: vec!["campaign_performance".to_string()],
                    owner: "marketing-campaigns@company.com".to_string(),
                    retention_days: 1095, // 3 years
                    refresh_frequency: "daily".to_string(),
                },
            ],
            access_policies: vec![
                AccessPolicy {
                    role: "MarketingAnalyst".to_string(),
                    datasets: vec!["campaign_performance".to_string()],
                    operations: vec!["read".to_string()],
                    conditions: "data_minimization AND pii_masking".to_string(),
                },
            ],
        }
    }

    pub fn sales() -> Self {
        DomainTemplate {
            name: "Sales".to_string(),
            description: "Sales data domain - Pipeline, Opportunities, Forecasting".to_string(),
            governance_level: "Revenue-aligned, pipeline transparency".to_string(),
            steward: "Chief Revenue Officer".to_string(),
            data_classification: "Confidential".to_string(),
            sla: SLA {
                availability_percentage: 99.9,
                max_latency_minutes: 5,
                rto_hours: 2,
                rpo_minutes: 60,
            },
            quality_rules: vec![
                QualityRule {
                    name: "Opportunity Accuracy".to_string(),
                    metric: "lost_deal_prediction_accuracy".to_string(),
                    threshold: 0.85,
                    dimension: "accuracy".to_string(),
                },
                QualityRule {
                    name: "Forecast Completeness".to_string(),
                    metric: "reps_with_forecast_submission".to_string(),
                    threshold: 0.95,
                    dimension: "completeness".to_string(),
                },
            ],
            data_contracts: vec![
                DataContract {
                    name: "Pipeline Opportunities".to_string(),
                    datasets: vec!["pipeline_opportunities".to_string()],
                    owner: "sales-pipeline@company.com".to_string(),
                    retention_days: 365, // 1 year
                    refresh_frequency: "daily".to_string(),
                },
                DataContract {
                    name: "Deal Velocity Metrics".to_string(),
                    datasets: vec!["deal_velocity".to_string()],
                    owner: "sales-deals@company.com".to_string(),
                    retention_days: 730, // 2 years
                    refresh_frequency: "daily".to_string(),
                },
            ],
            access_policies: vec![
                AccessPolicy {
                    role: "SalesRepresentative".to_string(),
                    datasets: vec!["pipeline_opportunities".to_string()],
                    operations: vec!["read".to_string(), "write".to_string()],
                    conditions: "own_opportunities_only".to_string(),
                },
                AccessPolicy {
                    role: "SalesManager".to_string(),
                    datasets: vec!["pipeline_opportunities".to_string(), "deal_velocity".to_string()],
                    operations: vec!["read".to_string()],
                    conditions: "team_opportunities".to_string(),
                },
            ],
        }
    }

    pub fn hr() -> Self {
        DomainTemplate {
            name: "HR".to_string(),
            description: "HR data domain - Employees, Payroll, Benefits, Compensation".to_string(),
            governance_level: "Privacy-critical, confidential data".to_string(),
            steward: "Chief Human Resources Officer".to_string(),
            data_classification: "Highly Confidential".to_string(),
            sla: SLA {
                availability_percentage: 99.99,
                max_latency_minutes: 60,
                rto_hours: 1,
                rpo_minutes: 5,
            },
            quality_rules: vec![
                QualityRule {
                    name: "Payroll Accuracy".to_string(),
                    metric: "audit_discrepancy_rate".to_string(),
                    threshold: 0.0,
                    dimension: "accuracy".to_string(),
                },
                QualityRule {
                    name: "Employee Roster Completeness".to_string(),
                    metric: "roster_vs_systems_match".to_string(),
                    threshold: 1.0,
                    dimension: "completeness".to_string(),
                },
            ],
            data_contracts: vec![
                DataContract {
                    name: "Employee Roster".to_string(),
                    datasets: vec!["employee_roster".to_string()],
                    owner: "hr-roster@company.com".to_string(),
                    retention_days: 2555, // 7 years for compliance
                    refresh_frequency: "daily".to_string(),
                },
                DataContract {
                    name: "Payroll Records".to_string(),
                    datasets: vec!["payroll_records".to_string()],
                    owner: "hr-payroll@company.com".to_string(),
                    retention_days: 2555, // 7 years for tax compliance
                    refresh_frequency: "bi-weekly".to_string(),
                },
            ],
            access_policies: vec![
                AccessPolicy {
                    role: "HRManager".to_string(),
                    datasets: vec!["employee_roster".to_string()],
                    operations: vec!["read".to_string()],
                    conditions: "mfa_required AND encrypted_channel AND audit_logging".to_string(),
                },
                AccessPolicy {
                    role: "PayrollAdministrator".to_string(),
                    datasets: vec!["payroll_records".to_string()],
                    operations: vec!["read".to_string(), "write".to_string()],
                    conditions: "dual_approval AND end_to_end_encryption AND full_audit_trail".to_string(),
                },
                AccessPolicy {
                    role: "CEO".to_string(),
                    datasets: vec!["employee_roster".to_string(), "payroll_records".to_string()],
                    operations: vec!["read".to_string()],
                    conditions: "highly_restricted AND executive_approval AND no_export".to_string(),
                },
            ],
        }
    }

    pub fn by_name(name: &str) -> Option<Self> {
        match name.to_lowercase().as_str() {
            "finance" => Some(Self::finance()),
            "operations" => Some(Self::operations()),
            "marketing" => Some(Self::marketing()),
            "sales" => Some(Self::sales()),
            "hr" => Some(Self::hr()),
            _ => None,
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_finance_template() {
        let template = DomainTemplate::finance();
        assert_eq!(template.name, "Finance");
        assert!(template.sla.availability_percentage >= 99.0);
        assert!(!template.quality_rules.is_empty());
        assert!(!template.access_policies.is_empty());
    }

    #[test]
    fn test_operations_template() {
        let template = DomainTemplate::operations();
        assert_eq!(template.name, "Operations");
        assert_eq!(template.sla.max_latency_minutes, 1);
    }

    #[test]
    fn test_marketing_template() {
        let template = DomainTemplate::marketing();
        assert_eq!(template.name, "Marketing");
        assert!(template.description.contains("Attribution"));
    }

    #[test]
    fn test_sales_template() {
        let template = DomainTemplate::sales();
        assert_eq!(template.name, "Sales");
        assert!(!template.data_contracts.is_empty());
    }

    #[test]
    fn test_hr_template() {
        let template = DomainTemplate::hr();
        assert_eq!(template.name, "HR");
        assert_eq!(template.data_classification, "Highly Confidential");
        assert!(template.sla.availability_percentage > 99.9);
    }

    #[test]
    fn test_by_name_lookup() {
        assert!(DomainTemplate::by_name("finance").is_some());
        assert!(DomainTemplate::by_name("Finance").is_some());
        assert!(DomainTemplate::by_name("operations").is_some());
        assert!(DomainTemplate::by_name("marketing").is_some());
        assert!(DomainTemplate::by_name("sales").is_some());
        assert!(DomainTemplate::by_name("hr").is_some());
        assert!(DomainTemplate::by_name("invalid").is_none());
    }

    #[test]
    fn test_all_templates_have_slas() {
        for domain in &["finance", "operations", "marketing", "sales", "hr"] {
            let template = DomainTemplate::by_name(domain).unwrap();
            assert!(template.sla.availability_percentage > 0.0);
            assert!(template.sla.availability_percentage <= 100.0);
            assert!(template.sla.rto_hours > 0);
            assert!(template.sla.rpo_minutes > 0);
        }
    }

    #[test]
    fn test_access_policies_have_conditions() {
        let finance = DomainTemplate::finance();
        for policy in finance.access_policies {
            assert!(!policy.conditions.is_empty());
            assert!(!policy.operations.is_empty());
        }
    }

    #[test]
    fn test_data_contracts_retention() {
        let templates = vec![
            DomainTemplate::finance(),
            DomainTemplate::operations(),
            DomainTemplate::marketing(),
            DomainTemplate::sales(),
            DomainTemplate::hr(),
        ];

        for template in templates {
            for contract in template.data_contracts {
                assert!(contract.retention_days > 0);
                assert!(!contract.refresh_frequency.is_empty());
            }
        }
    }
}
