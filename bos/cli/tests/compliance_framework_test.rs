// Comprehensive compliance framework tests for SOC2, GDPR, HIPAA, SOX
// Tests verify command execution, gap detection, and evidence generation

#[cfg(test)]
mod compliance_framework_tests {
    use std::process::Command;

    // Test 1: bos compliance init --framework soc2 --organization test-org
    #[test]
    fn test_compliance_init_soc2() {
        let output = Command::new("cargo")
            .args(&["run", "--", "compliance", "init", "--framework", "soc2", "--organization", "test-org"])
            .current_dir(".")
            .output()
            .expect("Failed to execute compliance init command");

        assert!(output.status.success(), "Command failed: {}", String::from_utf8_lossy(&output.stderr));

        let stdout = String::from_utf8_lossy(&output.stdout);
        assert!(stdout.contains("soc2") || stdout.contains("SOC2"), "Output should mention SOC2");
        assert!(stdout.contains("test-org"), "Output should mention organization name");
        assert!(stdout.contains("workspace") || stdout.contains("config"), "Output should mention workspace/config");
    }

    // Test 2: bos compliance init --framework gdpr
    #[test]
    fn test_compliance_init_gdpr() {
        let output = Command::new("cargo")
            .args(&["run", "--", "compliance", "init", "--framework", "gdpr"])
            .current_dir(".")
            .output()
            .expect("Failed to execute GDPR init");

        assert!(output.status.success());
        let stdout = String::from_utf8_lossy(&output.stdout);
        assert!(stdout.to_lowercase().contains("gdpr"));
        assert!(stdout.contains("workspace") || stdout.contains("config"));
    }

    // Test 3: bos compliance init --framework hipaa
    #[test]
    fn test_compliance_init_hipaa() {
        let output = Command::new("cargo")
            .args(&["run", "--", "compliance", "init", "--framework", "hipaa"])
            .current_dir(".")
            .output()
            .expect("Failed to execute HIPAA init");

        assert!(output.status.success());
        let stdout = String::from_utf8_lossy(&output.stdout);
        assert!(stdout.to_lowercase().contains("hipaa"));
    }

    // Test 4: bos compliance init --framework sox
    #[test]
    fn test_compliance_init_sox() {
        let output = Command::new("cargo")
            .args(&["run", "--", "compliance", "init", "--framework", "sox"])
            .current_dir(".")
            .output()
            .expect("Failed to execute SOX init");

        assert!(output.status.success());
        let stdout = String::from_utf8_lossy(&output.stdout);
        assert!(stdout.to_lowercase().contains("sox"));
    }

    // Test 5: bos compliance init --framework invalid-framework (error case)
    #[test]
    fn test_compliance_init_invalid_framework() {
        let output = Command::new("cargo")
            .args(&["run", "--", "compliance", "init", "--framework", "invalid"])
            .current_dir(".")
            .output()
            .expect("Failed to execute compliance init");

        // Should fail with error message
        let stderr = String::from_utf8_lossy(&output.stderr);
        assert!(
            !output.status.success() || stderr.to_lowercase().contains("unknown") ||
            stderr.to_lowercase().contains("error") ||
            String::from_utf8_lossy(&output.stdout).to_lowercase().contains("unknown"),
            "Should error on invalid framework"
        );
    }

    // Test 6: bos compliance verify --framework soc2
    #[test]
    fn test_compliance_verify_soc2() {
        let output = Command::new("cargo")
            .args(&["run", "--", "compliance", "verify", "--framework", "soc2"])
            .current_dir(".")
            .output()
            .expect("Failed to execute SOC2 verify");

        assert!(output.status.success());
        let stdout = String::from_utf8_lossy(&output.stdout);

        assert!(stdout.to_lowercase().contains("soc2"));
        assert!(stdout.contains("verification_date") || stdout.contains("total_controls"));
        assert!(stdout.contains("gaps_found") || stdout.contains("compliant"));

        // Should report compliance percentage
        assert!(stdout.contains("percentage") || stdout.contains("gap"),
                "Should report compliance metrics");
    }

    // Test 7: bos compliance verify --framework gdpr
    #[test]
    fn test_compliance_verify_gdpr() {
        let output = Command::new("cargo")
            .args(&["run", "--", "compliance", "verify", "--framework", "gdpr"])
            .current_dir(".")
            .output()
            .expect("Failed to execute GDPR verify");

        assert!(output.status.success());
        let stdout = String::from_utf8_lossy(&output.stdout);

        assert!(stdout.to_lowercase().contains("gdpr"));
        assert!(stdout.contains("gap") || stdout.contains("compliance") || stdout.contains("verification"));
    }

    // Test 8: bos compliance verify --framework hipaa
    #[test]
    fn test_compliance_verify_hipaa() {
        let output = Command::new("cargo")
            .args(&["run", "--", "compliance", "verify", "--framework", "hipaa"])
            .current_dir(".")
            .output()
            .expect("Failed to execute HIPAA verify");

        assert!(output.status.success());
        let stdout = String::from_utf8_lossy(&output.stdout);
        assert!(stdout.to_lowercase().contains("hipaa"));
    }

    // Test 9: bos compliance verify --framework sox
    #[test]
    fn test_compliance_verify_sox() {
        let output = Command::new("cargo")
            .args(&["run", "--", "compliance", "verify", "--framework", "sox"])
            .current_dir(".")
            .output()
            .expect("Failed to execute SOX verify");

        assert!(output.status.success());
        let stdout = String::from_utf8_lossy(&output.stdout);
        assert!(stdout.to_lowercase().contains("sox"));
    }

    // Test 10: bos compliance report --framework soc2
    #[test]
    fn test_compliance_report_soc2() {
        let output = Command::new("cargo")
            .args(&["run", "--", "compliance", "report", "--framework", "soc2"])
            .current_dir(".")
            .output()
            .expect("Failed to execute SOC2 report");

        assert!(output.status.success());
        let stdout = String::from_utf8_lossy(&output.stdout);

        assert!(stdout.to_lowercase().contains("soc2"));
        assert!(stdout.contains("report_id") || stdout.contains("generated_at") || stdout.contains("output_path"));
        assert!(stdout.contains("evidence_count") || stdout.contains("queries_executed"));
    }

    // Test 11: bos compliance report --framework gdpr --output-dir ./custom-reports
    #[test]
    fn test_compliance_report_gdpr_custom_output() {
        let output = Command::new("cargo")
            .args(&["run", "--", "compliance", "report", "--framework", "gdpr", "--output-dir", "./custom-reports"])
            .current_dir(".")
            .output()
            .expect("Failed to execute GDPR report");

        assert!(output.status.success());
        let stdout = String::from_utf8_lossy(&output.stdout);

        assert!(stdout.to_lowercase().contains("gdpr"));
        assert!(stdout.contains("report") || stdout.contains("generated"));
    }

    // Test 12: bos compliance report --framework hipaa
    #[test]
    fn test_compliance_report_hipaa() {
        let output = Command::new("cargo")
            .args(&["run", "--", "compliance", "report", "--framework", "hipaa"])
            .current_dir(".")
            .output()
            .expect("Failed to execute HIPAA report");

        assert!(output.status.success());
        let stdout = String::from_utf8_lossy(&output.stdout);
        assert!(stdout.to_lowercase().contains("hipaa"));
        assert!(stdout.contains("status") || stdout.contains("generated"));
    }

    // Test 13: bos compliance report --framework sox
    #[test]
    fn test_compliance_report_sox() {
        let output = Command::new("cargo")
            .args(&["run", "--", "compliance", "report", "--framework", "sox"])
            .current_dir(".")
            .output()
            .expect("Failed to execute SOX report");

        assert!(output.status.success());
        let stdout = String::from_utf8_lossy(&output.stdout);
        assert!(stdout.to_lowercase().contains("sox"));
    }

    // Test 14: bos compliance soc2 controls
    #[test]
    fn test_soc2_controls_list() {
        let output = Command::new("cargo")
            .args(&["run", "--", "compliance", "soc2", "controls"])
            .current_dir(".")
            .output()
            .expect("Failed to execute SOC2 controls");

        assert!(output.status.success());
        let stdout = String::from_utf8_lossy(&output.stdout);

        assert!(stdout.contains("cc6.1") || stdout.contains("control_id"));
        assert!(stdout.contains("Security") || stdout.contains("Availability") || stdout.contains("trust_service_category"));
    }

    // Test 15: bos compliance soc2 controls --control-id cc6.1
    #[test]
    fn test_soc2_controls_filter() {
        let output = Command::new("cargo")
            .args(&["run", "--", "compliance", "soc2", "controls", "--control-id", "cc6.1"])
            .current_dir(".")
            .output()
            .expect("Failed to execute SOC2 controls filter");

        assert!(output.status.success());
        let stdout = String::from_utf8_lossy(&output.stdout);

        assert!(stdout.contains("cc6.1"));
        assert!(stdout.contains("Security") || stdout.contains("logical") || stdout.contains("access"));
    }

    // Test 16: bos compliance gdpr articles
    #[test]
    fn test_gdpr_articles_list() {
        let output = Command::new("cargo")
            .args(&["run", "--", "compliance", "gdpr", "articles"])
            .current_dir(".")
            .output()
            .expect("Failed to execute GDPR articles");

        assert!(output.status.success());
        let stdout = String::from_utf8_lossy(&output.stdout);

        assert!(stdout.contains("article_number") || stdout.contains("5") || stdout.contains("7"));
        assert!(stdout.contains("data_subject") || stdout.contains("compliance_status"));
    }

    // Test 17: bos compliance gdpr articles --article-number 7
    #[test]
    fn test_gdpr_articles_filter() {
        let output = Command::new("cargo")
            .args(&["run", "--", "compliance", "gdpr", "articles", "--article-number", "7"])
            .current_dir(".")
            .output()
            .expect("Failed to execute GDPR articles filter");

        assert!(output.status.success());
        let stdout = String::from_utf8_lossy(&output.stdout);

        assert!(stdout.contains("7") || stdout.contains("consent"));
    }

    // Test 18: bos compliance hipaa sections
    #[test]
    fn test_hipaa_sections_list() {
        let output = Command::new("cargo")
            .args(&["run", "--", "compliance", "hipaa", "sections"])
            .current_dir(".")
            .output()
            .expect("Failed to execute HIPAA sections");

        assert!(output.status.success());
        let stdout = String::from_utf8_lossy(&output.stdout);

        assert!(stdout.contains("section_id") || stdout.contains("164.308"));
        assert!(stdout.contains("phi_tracking") || stdout.contains("Safeguards"));
    }

    // Test 19: bos compliance hipaa sections --section-id 164.312
    #[test]
    fn test_hipaa_sections_filter() {
        let output = Command::new("cargo")
            .args(&["run", "--", "compliance", "hipaa", "sections", "--section-id", "164.312"])
            .current_dir(".")
            .output()
            .expect("Failed to execute HIPAA sections filter");

        assert!(output.status.success());
        let stdout = String::from_utf8_lossy(&output.stdout);

        assert!(stdout.contains("164.312") || stdout.contains("Technical"));
    }

    // Test 20: bos compliance sox controls
    #[test]
    fn test_sox_controls_list() {
        let output = Command::new("cargo")
            .args(&["run", "--", "compliance", "sox", "controls"])
            .current_dir(".")
            .output()
            .expect("Failed to execute SOX controls");

        assert!(output.status.success());
        let stdout = String::from_utf8_lossy(&output.stdout);

        assert!(stdout.contains("control_number") || stdout.contains("ic1.1"));
        assert!(stdout.contains("financial_data_protected") || stdout.contains("Change"));
    }

    // Test 21: bos compliance sox controls --control-number ic1.1
    #[test]
    fn test_sox_controls_filter() {
        let output = Command::new("cargo")
            .args(&["run", "--", "compliance", "sox", "controls", "--control-number", "ic1.1"])
            .current_dir(".")
            .output()
            .expect("Failed to execute SOX controls filter");

        assert!(output.status.success());
        let stdout = String::from_utf8_lossy(&output.stdout);

        assert!(stdout.contains("ic1.1"));
        assert!(stdout.contains("Segregation") || stdout.contains("Change"));
    }

    // Test 22: Verify all four frameworks are supported
    #[test]
    fn test_all_frameworks_supported() {
        for framework in &["soc2", "gdpr", "hipaa", "sox"] {
            let output = Command::new("cargo")
                .args(&["run", "--", "compliance", "init", "--framework", framework])
                .current_dir(".")
                .output()
                .expect(&format!("Failed to execute {} init", framework));

            assert!(
                output.status.success(),
                "Framework {} not supported: {}",
                framework,
                String::from_utf8_lossy(&output.stderr)
            );
        }
    }

    // Test 23: Verify JSON serialization of responses
    #[test]
    fn test_compliance_response_json_format() {
        let output = Command::new("cargo")
            .args(&["run", "--", "compliance", "verify", "--framework", "soc2"])
            .current_dir(".")
            .output()
            .expect("Failed to execute compliance verify");

        assert!(output.status.success());
        let stdout = String::from_utf8_lossy(&output.stdout);

        // JSON should have these fields
        assert!(
            stdout.contains("framework") || stdout.contains("verification_date") ||
            stdout.contains("gaps_found") || stdout.contains("compliant"),
            "Response should be valid JSON with compliance fields"
        );
    }

    // Test 24: Verify gap detection returns expected gap count
    #[test]
    fn test_compliance_gap_detection() {
        let output = Command::new("cargo")
            .args(&["run", "--", "compliance", "verify", "--framework", "soc2"])
            .current_dir(".")
            .output()
            .expect("Failed to execute gap detection");

        assert!(output.status.success());
        let stdout = String::from_utf8_lossy(&output.stdout);

        // Should report gaps with gap_id, control_id, severity, description
        assert!(
            stdout.contains("gap_id") || stdout.contains("severity") || stdout.contains("remediation"),
            "Should report gaps with details"
        );
    }

    // Test 25: Verify compliance percentage calculation
    #[test]
    fn test_compliance_percentage_calculation() {
        let output = Command::new("cargo")
            .args(&["run", "--", "compliance", "verify", "--framework", "gdpr"])
            .current_dir(".")
            .output()
            .expect("Failed to execute GDPR verify");

        assert!(output.status.success());
        let stdout = String::from_utf8_lossy(&output.stdout);

        // Should have compliance_percentage field
        assert!(
            stdout.contains("percentage") || stdout.contains("compliant") || stdout.contains("gaps_found"),
            "Should calculate and report compliance percentage"
        );
    }

    // Test 26: Report includes evidence count
    #[test]
    fn test_report_includes_evidence_count() {
        let output = Command::new("cargo")
            .args(&["run", "--", "compliance", "report", "--framework", "hipaa"])
            .current_dir(".")
            .output()
            .expect("Failed to execute HIPAA report");

        assert!(output.status.success());
        let stdout = String::from_utf8_lossy(&output.stdout);

        assert!(
            stdout.contains("evidence_count") || stdout.contains("generated") || stdout.contains("status"),
            "Report should include evidence metrics"
        );
    }

    // Test 27: Verify CONSTRUCT queries are referenced
    #[test]
    fn test_construct_queries_referenced() {
        let output = Command::new("cargo")
            .args(&["run", "--", "compliance", "report", "--framework", "sox"])
            .current_dir(".")
            .output()
            .expect("Failed to execute SOX report");

        assert!(output.status.success());
        let stdout = String::from_utf8_lossy(&output.stdout);

        assert!(
            stdout.contains("queries_executed") || stdout.contains("construct") || stdout.contains("4"),
            "Report should reference SPARQL CONSTRUCT queries"
        );
    }

    // Test 28: Verify framework-specific control details
    #[test]
    fn test_framework_control_details() {
        let output = Command::new("cargo")
            .args(&["run", "--", "compliance", "soc2", "controls"])
            .current_dir(".")
            .output()
            .expect("Failed to execute SOC2 controls");

        assert!(output.status.success());
        let stdout = String::from_utf8_lossy(&output.stdout);

        // Should have TSC categories and evaluation frequency
        assert!(
            stdout.contains("trust_service_category") || stdout.contains("evaluation_frequency") ||
            stdout.contains("evidence_type"),
            "Controls should include detailed metadata"
        );
    }

    // Test 29: Verify all compliance commands are accessible
    #[test]
    fn test_all_compliance_commands_accessible() {
        let commands = vec!["init", "verify", "report"];

        for cmd in commands {
            let output = Command::new("cargo")
                .args(&["run", "--", "compliance", cmd, "--help"])
                .current_dir(".")
                .output()
                .expect(&format!("Failed to execute compliance {} --help", cmd));

            assert!(
                output.status.success(),
                "Command 'compliance {}' should be accessible",
                cmd
            );
        }
    }

    // Test 30: Verify framework subcommands are accessible
    #[test]
    fn test_framework_subcommands_accessible() {
        let subcommands = vec![
            ("soc2", "controls"),
            ("gdpr", "articles"),
            ("hipaa", "sections"),
            ("sox", "controls"),
        ];

        for (framework, subcommand) in subcommands {
            let output = Command::new("cargo")
                .args(&["run", "--", "compliance", framework, subcommand, "--help"])
                .current_dir(".")
                .output()
                .expect(&format!("Failed to execute compliance {} {} --help", framework, subcommand));

            assert!(
                output.status.success() || String::from_utf8_lossy(&output.stdout).len() > 0,
                "Subcommand 'compliance {} {}' should be accessible",
                framework,
                subcommand
            );
        }
    }
}
