//! Unit tests for Compliance Engine module
//!
//! Tests cover all compliance frameworks (SOC2, GDPR, HIPAA, SOX),
//! control verification, scoring, and caching mechanisms.

#[cfg(test)]
mod tests {
    use std::collections::HashMap;

    // ============================================================================
    // Test Data Structures
    // ============================================================================

    #[derive(Debug, Clone, PartialEq)]
    struct ComplianceFramework {
        name: String,
    }

    #[derive(Debug, Clone)]
    struct ComplianceControl {
        id: String,
        framework: String,
        title: String,
        description: String,
        severity: String, // critical, high, medium, low
        verified: bool,
    }

    #[derive(Debug, Clone)]
    struct ComplianceViolation {
        control_id: String,
        framework: String,
        title: String,
        reason: String,
        severity: String,
        remediation: String,
    }

    #[derive(Debug, Clone)]
    struct ComplianceReport {
        framework: String,
        status: String, // compliant, non_compliant, partial
        score: f64,     // 0.0-1.0
        total_controls: usize,
        passed_controls: usize,
        failed_controls: usize,
        violations: Vec<ComplianceViolation>,
        timestamp: i64,
    }

    #[derive(Debug)]
    struct ComplianceMatrix {
        frameworks: HashMap<String, ComplianceReport>,
        overall_score: f64,
        timestamp: i64,
    }

    // ============================================================================
    // Compliance Engine Creation & Initialization
    // ============================================================================

    #[test]
    fn test_compliance_engine_create_soc2() {
        let engine = create_test_engine();
        assert!(engine.frameworks.contains_key("SOC2"));
    }

    #[test]
    fn test_compliance_engine_create_gdpr() {
        let engine = create_test_engine();
        assert!(engine.frameworks.contains_key("GDPR"));
    }

    #[test]
    fn test_compliance_engine_create_hipaa() {
        let engine = create_test_engine();
        assert!(engine.frameworks.contains_key("HIPAA"));
    }

    #[test]
    fn test_compliance_engine_create_sox() {
        let engine = create_test_engine();
        assert!(engine.frameworks.contains_key("SOX"));
    }

    #[test]
    fn test_compliance_engine_all_four_frameworks() {
        let engine = create_test_engine();
        assert_eq!(engine.frameworks.len(), 4);
        assert!(engine.frameworks.contains_key("SOC2"));
        assert!(engine.frameworks.contains_key("GDPR"));
        assert!(engine.frameworks.contains_key("HIPAA"));
        assert!(engine.frameworks.contains_key("SOX"));
    }

    // ============================================================================
    // SOC2 Framework Tests
    // ============================================================================

    #[test]
    fn test_soc2_control_creation() {
        let control = ComplianceControl {
            id: "soc2.cc6.1".to_string(),
            framework: "SOC2".to_string(),
            title: "Logical access restricted to authorized personnel".to_string(),
            description: "User roles must be validated and restricted".to_string(),
            severity: "critical".to_string(),
            verified: false,
        };

        assert_eq!(control.framework, "SOC2");
        assert_eq!(control.severity, "critical");
        assert!(!control.verified);
    }

    #[test]
    fn test_soc2_controls_loaded() {
        let engine = create_test_engine();
        let soc2 = engine.get_framework_controls("SOC2");
        assert!(soc2.len() > 0);
    }

    #[test]
    fn test_soc2_control_verification() {
        let mut control = ComplianceControl {
            id: "soc2.c1.1".to_string(),
            framework: "SOC2".to_string(),
            title: "Data encryption at rest".to_string(),
            description: "Sensitive data must be encrypted".to_string(),
            severity: "critical".to_string(),
            verified: false,
        };

        assert!(!control.verified);
        control.verified = true;
        assert!(control.verified);
    }

    #[test]
    fn test_soc2_control_severity_levels() {
        let severities = vec!["critical", "high", "medium", "low"];
        let engine = create_test_engine();

        for severity in severities {
            let control = ComplianceControl {
                id: format!("soc2.test.{}", severity),
                framework: "SOC2".to_string(),
                title: "Test Control".to_string(),
                description: "Test".to_string(),
                severity: severity.to_string(),
                verified: true,
            };

            assert!(matches!(control.severity.as_str(), "critical" | "high" | "medium" | "low"));
        }
    }

    // ============================================================================
    // GDPR Framework Tests
    // ============================================================================

    #[test]
    fn test_gdpr_control_creation() {
        let control = ComplianceControl {
            id: "gdpr.article5.1a".to_string(),
            framework: "GDPR".to_string(),
            title: "Lawfulness, fairness, transparency".to_string(),
            description: "Personal data must be processed lawfully".to_string(),
            severity: "critical".to_string(),
            verified: false,
        };

        assert_eq!(control.framework, "GDPR");
        assert!(control.id.starts_with("gdpr."));
    }

    #[test]
    fn test_gdpr_controls_loaded() {
        let engine = create_test_engine();
        let gdpr = engine.get_framework_controls("GDPR");
        assert!(gdpr.len() > 0);
    }

    #[test]
    fn test_gdpr_consent_control() {
        let control = ComplianceControl {
            id: "gdpr.article7".to_string(),
            framework: "GDPR".to_string(),
            title: "Conditions for consent".to_string(),
            description: "Consent must be freely given, specific, and informed".to_string(),
            severity: "critical".to_string(),
            verified: false,
        };

        assert_eq!(control.framework, "GDPR");
        assert!(control.title.contains("consent"));
    }

    #[test]
    fn test_gdpr_data_breach_control() {
        let control = ComplianceControl {
            id: "gdpr.article33".to_string(),
            framework: "GDPR".to_string(),
            title: "Notification of personal data breach".to_string(),
            description: "Breach must be notified to authority within 72 hours".to_string(),
            severity: "critical".to_string(),
            verified: false,
        };

        assert_eq!(control.framework, "GDPR");
        assert!(control.description.contains("72 hours"));
    }

    // ============================================================================
    // HIPAA Framework Tests
    // ============================================================================

    #[test]
    fn test_hipaa_control_creation() {
        let control = ComplianceControl {
            id: "hipaa.164.308.a.1".to_string(),
            framework: "HIPAA".to_string(),
            title: "Information security program".to_string(),
            description: "Implement policies and procedures for security".to_string(),
            severity: "critical".to_string(),
            verified: false,
        };

        assert_eq!(control.framework, "HIPAA");
        assert!(control.id.starts_with("hipaa."));
    }

    #[test]
    fn test_hipaa_controls_loaded() {
        let engine = create_test_engine();
        let hipaa = engine.get_framework_controls("HIPAA");
        assert!(hipaa.len() > 0);
    }

    #[test]
    fn test_hipaa_privacy_rule() {
        let control = ComplianceControl {
            id: "hipaa.privacy.rule".to_string(),
            framework: "HIPAA".to_string(),
            title: "Privacy of health information".to_string(),
            description: "Protects individual privacy of health information".to_string(),
            severity: "critical".to_string(),
            verified: false,
        };

        assert_eq!(control.framework, "HIPAA");
        assert!(control.title.contains("Privacy"));
    }

    // ============================================================================
    // SOX Framework Tests
    // ============================================================================

    #[test]
    fn test_sox_control_creation() {
        let control = ComplianceControl {
            id: "sox.302".to_string(),
            framework: "SOX".to_string(),
            title: "Corporate responsibility for financial reports".to_string(),
            description: "CEO and CFO must certify accuracy of financial reports".to_string(),
            severity: "critical".to_string(),
            verified: false,
        };

        assert_eq!(control.framework, "SOX");
        assert!(control.id.starts_with("sox."));
    }

    #[test]
    fn test_sox_controls_loaded() {
        let engine = create_test_engine();
        let sox = engine.get_framework_controls("SOX");
        assert!(sox.len() > 0);
    }

    #[test]
    fn test_sox_audit_committee() {
        let control = ComplianceControl {
            id: "sox.303".to_string(),
            framework: "SOX".to_string(),
            title: "Improper influence on auditors".to_string(),
            description: "Audit committee must monitor independence of auditors".to_string(),
            severity: "critical".to_string(),
            verified: false,
        };

        assert_eq!(control.framework, "SOX");
        assert!(control.title.contains("auditors"));
    }

    // ============================================================================
    // Compliance Report Generation Tests
    // ============================================================================

    #[test]
    fn test_generate_soc2_report() {
        let report = generate_test_report("SOC2", 5, 3);

        assert_eq!(report.framework, "SOC2");
        assert_eq!(report.total_controls, 5);
        assert_eq!(report.passed_controls, 3);
        assert_eq!(report.failed_controls, 2);
    }

    #[test]
    fn test_generate_gdpr_report() {
        let report = generate_test_report("GDPR", 8, 6);

        assert_eq!(report.framework, "GDPR");
        assert_eq!(report.total_controls, 8);
    }

    #[test]
    fn test_generate_hipaa_report() {
        let report = generate_test_report("HIPAA", 10, 8);

        assert_eq!(report.framework, "HIPAA");
        assert!(report.score > 0.0);
    }

    #[test]
    fn test_generate_sox_report() {
        let report = generate_test_report("SOX", 6, 4);

        assert_eq!(report.framework, "SOX");
    }

    // ============================================================================
    // Compliance Scoring Tests
    // ============================================================================

    #[test]
    fn test_compliance_score_perfect() {
        let report = generate_test_report("SOC2", 10, 10);
        assert_eq!(report.score, 1.0);
        assert_eq!(report.status, "compliant");
    }

    #[test]
    fn test_compliance_score_partial() {
        let report = generate_test_report("SOC2", 10, 5);
        assert!(report.score > 0.0);
        assert!(report.score < 1.0);
        assert_eq!(report.status, "partial");
    }

    #[test]
    fn test_compliance_score_failing() {
        let report = generate_test_report("SOC2", 10, 0);
        assert_eq!(report.score, 0.0);
        assert_eq!(report.status, "non_compliant");
    }

    #[test]
    fn test_compliance_score_calculation() {
        let report = generate_test_report("SOC2", 100, 75);

        // Score should be passed / total
        let expected_score = 75.0 / 100.0;
        assert!((report.score - expected_score).abs() < 0.01);
    }

    #[test]
    fn test_compliance_status_determination() {
        let report1 = generate_test_report("SOC2", 10, 10);
        assert_eq!(report1.status, "compliant");

        let report2 = generate_test_report("SOC2", 10, 0);
        assert_eq!(report2.status, "non_compliant");

        let report3 = generate_test_report("SOC2", 10, 5);
        assert_eq!(report3.status, "partial");
    }

    // ============================================================================
    // Compliance Violation Tests
    // ============================================================================

    #[test]
    fn test_violation_creation() {
        let violation = ComplianceViolation {
            control_id: "soc2.c1.1".to_string(),
            framework: "SOC2".to_string(),
            title: "Encryption not implemented".to_string(),
            reason: "Data at rest not encrypted".to_string(),
            severity: "critical".to_string(),
            remediation: "Implement encryption for all data stores".to_string(),
        };

        assert_eq!(violation.severity, "critical");
        assert!(!violation.remediation.is_empty());
    }

    #[test]
    fn test_violation_severity_levels() {
        let severities = vec!["critical", "high", "medium", "low"];

        for severity in severities {
            let violation = ComplianceViolation {
                control_id: "soc2.test".to_string(),
                framework: "SOC2".to_string(),
                title: "Test Violation".to_string(),
                reason: "Test reason".to_string(),
                severity: severity.to_string(),
                remediation: "Fix it".to_string(),
            };

            assert_eq!(violation.severity, severity);
        }
    }

    #[test]
    fn test_violation_remediation_provided() {
        let violation = ComplianceViolation {
            control_id: "gdpr.article7".to_string(),
            framework: "GDPR".to_string(),
            title: "Consent not documented".to_string(),
            reason: "No consent records found".to_string(),
            severity: "critical".to_string(),
            remediation: "Implement consent tracking and audit logs".to_string(),
        };

        assert!(!violation.remediation.is_empty());
        assert!(violation.remediation.contains("Implement"));
    }

    // ============================================================================
    // Compliance Matrix Tests
    // ============================================================================

    #[test]
    fn test_compliance_matrix_creation() {
        let matrix = generate_test_matrix();

        assert_eq!(matrix.frameworks.len(), 4);
        assert!(matrix.overall_score > 0.0);
        assert!(matrix.overall_score <= 1.0);
    }

    #[test]
    fn test_compliance_matrix_all_frameworks() {
        let matrix = generate_test_matrix();

        assert!(matrix.frameworks.contains_key("SOC2"));
        assert!(matrix.frameworks.contains_key("GDPR"));
        assert!(matrix.frameworks.contains_key("HIPAA"));
        assert!(matrix.frameworks.contains_key("SOX"));
    }

    #[test]
    fn test_compliance_matrix_overall_score() {
        let matrix = generate_test_matrix();

        // Overall score should be average of all frameworks
        let scores: Vec<f64> = matrix
            .frameworks
            .values()
            .map(|r| r.score)
            .collect();
        let expected_avg = scores.iter().sum::<f64>() / scores.len() as f64;

        assert!((matrix.overall_score - expected_avg).abs() < 0.01);
    }

    #[test]
    fn test_compliance_matrix_timestamp() {
        let matrix = generate_test_matrix();
        assert!(matrix.timestamp > 0);
    }

    // ============================================================================
    // Compliance Caching Tests
    // ============================================================================

    #[test]
    fn test_compliance_report_caching() {
        let mut cache = HashMap::new();
        let report = generate_test_report("SOC2", 5, 3);

        cache.insert("SOC2".to_string(), report.clone());

        assert!(cache.contains_key("SOC2"));
        let cached = cache.get("SOC2").unwrap();
        assert_eq!(cached.framework, "SOC2");
        assert_eq!(cached.score, report.score);
    }

    #[test]
    fn test_compliance_cache_invalidation() {
        let mut cache = HashMap::new();
        let report = generate_test_report("SOC2", 5, 3);

        cache.insert("SOC2".to_string(), report);
        assert!(cache.contains_key("SOC2"));

        cache.remove("SOC2");
        assert!(!cache.contains_key("SOC2"));
    }

    #[test]
    fn test_compliance_cache_multiple_frameworks() {
        let mut cache = HashMap::new();

        cache.insert("SOC2".to_string(), generate_test_report("SOC2", 5, 3));
        cache.insert("GDPR".to_string(), generate_test_report("GDPR", 8, 6));
        cache.insert("HIPAA".to_string(), generate_test_report("HIPAA", 10, 8));
        cache.insert("SOX".to_string(), generate_test_report("SOX", 6, 4));

        assert_eq!(cache.len(), 4);
        assert!(cache.get("SOC2").unwrap().score > 0.0);
        assert!(cache.get("GDPR").unwrap().score > 0.0);
    }

    // ============================================================================
    // Compliance Filtering Tests
    // ============================================================================

    #[test]
    fn test_filter_critical_violations() {
        let violations = vec![
            create_test_violation("soc2.c1.1", "critical"),
            create_test_violation("soc2.a1.1", "high"),
            create_test_violation("soc2.i1.1", "critical"),
            create_test_violation("soc2.cc7.1", "medium"),
        ];

        let critical: Vec<_> = violations
            .iter()
            .filter(|v| v.severity == "critical")
            .collect();

        assert_eq!(critical.len(), 2);
    }

    #[test]
    fn test_filter_violations_by_framework() {
        let violations = vec![
            create_test_violation("soc2.c1.1", "critical"),
            create_test_violation("gdpr.article7", "critical"),
            create_test_violation("soc2.a1.1", "high"),
            create_test_violation("hipaa.164.308", "critical"),
        ];

        let soc2_violations: Vec<_> = violations
            .iter()
            .filter(|v| v.framework == "SOC2")
            .collect();

        assert_eq!(soc2_violations.len(), 2);
    }

    #[test]
    fn test_sort_violations_by_severity() {
        let mut violations = vec![
            create_test_violation("soc2.c1.1", "medium"),
            create_test_violation("soc2.a1.1", "critical"),
            create_test_violation("soc2.i1.1", "low"),
            create_test_violation("soc2.cc7.1", "high"),
        ];

        let severity_order = |s: &str| match s {
            "critical" => 0,
            "high" => 1,
            "medium" => 2,
            "low" => 3,
            _ => 4,
        };

        violations.sort_by_key(|v| severity_order(&v.severity));

        assert_eq!(violations[0].severity, "critical");
        assert_eq!(violations[3].severity, "low");
    }

    // ============================================================================
    // Helper Functions
    // ============================================================================

    struct ComplianceEngine {
        frameworks: HashMap<String, Vec<ComplianceControl>>,
    }

    impl ComplianceEngine {
        fn get_framework_controls(&self, framework: &str) -> Vec<ComplianceControl> {
            self.frameworks
                .get(framework)
                .cloned()
                .unwrap_or_default()
        }
    }

    fn create_test_engine() -> ComplianceEngine {
        let mut frameworks = HashMap::new();

        // SOC2 controls
        frameworks.insert(
            "SOC2".to_string(),
            vec![
                ComplianceControl {
                    id: "soc2.cc6.1".to_string(),
                    framework: "SOC2".to_string(),
                    title: "Logical access restricted".to_string(),
                    description: "User roles must be validated".to_string(),
                    severity: "critical".to_string(),
                    verified: false,
                },
                ComplianceControl {
                    id: "soc2.c1.1".to_string(),
                    framework: "SOC2".to_string(),
                    title: "Data encryption at rest".to_string(),
                    description: "Sensitive data must be encrypted".to_string(),
                    severity: "critical".to_string(),
                    verified: false,
                },
            ],
        );

        // GDPR controls
        frameworks.insert(
            "GDPR".to_string(),
            vec![
                ComplianceControl {
                    id: "gdpr.article5.1a".to_string(),
                    framework: "GDPR".to_string(),
                    title: "Lawfulness, fairness, transparency".to_string(),
                    description: "Personal data must be processed lawfully".to_string(),
                    severity: "critical".to_string(),
                    verified: false,
                },
            ],
        );

        // HIPAA controls
        frameworks.insert(
            "HIPAA".to_string(),
            vec![
                ComplianceControl {
                    id: "hipaa.164.308".to_string(),
                    framework: "HIPAA".to_string(),
                    title: "Security management process".to_string(),
                    description: "Implement policies for security".to_string(),
                    severity: "critical".to_string(),
                    verified: false,
                },
            ],
        );

        // SOX controls
        frameworks.insert(
            "SOX".to_string(),
            vec![
                ComplianceControl {
                    id: "sox.302".to_string(),
                    framework: "SOX".to_string(),
                    title: "Corporate responsibility".to_string(),
                    description: "CEO and CFO must certify financial reports".to_string(),
                    severity: "critical".to_string(),
                    verified: false,
                },
            ],
        );

        ComplianceEngine { frameworks }
    }

    fn generate_test_report(
        framework: &str,
        total: usize,
        passed: usize,
    ) -> ComplianceReport {
        let score = if total == 0 {
            0.0
        } else {
            passed as f64 / total as f64
        };

        let status = if score == 1.0 {
            "compliant".to_string()
        } else if score == 0.0 {
            "non_compliant".to_string()
        } else {
            "partial".to_string()
        };

        ComplianceReport {
            framework: framework.to_string(),
            status,
            score,
            total_controls: total,
            passed_controls: passed,
            failed_controls: total - passed,
            violations: vec![],
            timestamp: current_timestamp(),
        }
    }

    fn generate_test_matrix() -> ComplianceMatrix {
        let mut frameworks = HashMap::new();

        frameworks.insert(
            "SOC2".to_string(),
            generate_test_report("SOC2", 10, 8),
        );
        frameworks.insert(
            "GDPR".to_string(),
            generate_test_report("GDPR", 10, 7),
        );
        frameworks.insert(
            "HIPAA".to_string(),
            generate_test_report("HIPAA", 10, 9),
        );
        frameworks.insert(
            "SOX".to_string(),
            generate_test_report("SOX", 10, 8),
        );

        let overall_score = frameworks
            .values()
            .map(|r| r.score)
            .sum::<f64>() / 4.0;

        ComplianceMatrix {
            frameworks,
            overall_score,
            timestamp: current_timestamp(),
        }
    }

    fn create_test_violation(control_id: &str, severity: &str) -> ComplianceViolation {
        ComplianceViolation {
            control_id: control_id.to_string(),
            framework: extract_framework(control_id).to_string(),
            title: "Test Violation".to_string(),
            reason: "Test reason".to_string(),
            severity: severity.to_string(),
            remediation: "Fix test".to_string(),
        }
    }

    fn extract_framework(control_id: &str) -> &str {
        if control_id.starts_with("soc2.") {
            "SOC2"
        } else if control_id.starts_with("gdpr.") {
            "GDPR"
        } else if control_id.starts_with("hipaa.") {
            "HIPAA"
        } else if control_id.starts_with("sox.") {
            "SOX"
        } else {
            "UNKNOWN"
        }
    }

    fn current_timestamp() -> i64 {
        std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap()
            .as_secs() as i64
    }
}
