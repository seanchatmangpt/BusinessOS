// Healthcare Ontology Integration Tests
// Tests for HIPAA-compliant healthcare commands with PHI tracking, consent enforcement,
// and HIPAA compliance verification.

use assert_cmd::Command;
use predicates::prelude::*;
use serde_json::json;
use std::fs;
use tempfile::TempDir;

/// Test: Healthcare initialization creates proper workspace with HIPAA compliance
#[test]
fn test_healthcare_init_creates_workspace() {
    let mut cmd = Command::cargo_bin("bos").expect("Failed to find bos binary");

    let output = cmd
        .arg("healthcare")
        .arg("init")
        .arg("--organization")
        .arg("TestHospital")
        .arg("--fhir-version")
        .arg("r4")
        .output()
        .expect("Failed to execute healthcare init");

    assert!(output.status.success(), "healthcare init should succeed");

    let stdout = String::from_utf8_lossy(&output.stdout);
    assert!(stdout.contains("healthcare-testhospital"), "Should create workspace with org name");
    assert!(stdout.contains("1.0.0-phi"), "Should include PHI version marking");
    assert!(stdout.contains("HIPAA"), "Should mention HIPAA framework");
}

/// Test: PHI tracking command creates lineage tracking ID with PROV-O structure
#[test]
fn test_phi_track_creates_lineage() {
    let mut cmd = Command::cargo_bin("bos").expect("Failed to find bos binary");

    let output = cmd
        .arg("healthcare")
        .arg("phi")
        .arg("track")
        .arg("--patient-id")
        .arg("pat-001")
        .arg("--phi-elements")
        .arg("name,ssn,diagnosis")
        .arg("--storage-location")
        .arg("postgres")
        .output()
        .expect("Failed to execute healthcare phi track");

    assert!(output.status.success(), "PHI tracking should succeed");

    let stdout = String::from_utf8_lossy(&output.stdout);
    assert!(stdout.contains("PHI-pat-001-"), "Should create tracking ID");
    assert!(stdout.contains("pat-001"), "Should include patient ID");
    assert!(stdout.contains("3"), "Should count 3 PHI elements");
    assert!(stdout.contains("postgres"), "Should identify storage location");
}

/// Test: PHI tracking with default phi-elements parameter
#[test]
fn test_phi_track_default_elements() {
    let mut cmd = Command::cargo_bin("bos").expect("Failed to find bos binary");

    let output = cmd
        .arg("healthcare")
        .arg("phi")
        .arg("track")
        .arg("--patient-id")
        .arg("pat-002")
        .output()
        .expect("Failed to execute healthcare phi track");

    assert!(output.status.success(), "PHI tracking with defaults should succeed");

    let stdout = String::from_utf8_lossy(&output.stdout);
    assert!(stdout.contains("pat-002"), "Should include patient ID");
    assert!(stdout.contains("2"), "Should default to 2 PHI elements (name,ssn)");
}

/// Test: Consent check enforces ODRL policies for resource access
#[test]
fn test_consent_check_read_access() {
    let mut cmd = Command::cargo_bin("bos").expect("Failed to find bos binary");

    let output = cmd
        .arg("healthcare")
        .arg("consent")
        .arg("check")
        .arg("--patient-id")
        .arg("pat-001")
        .arg("--resource")
        .arg("Observation")
        .arg("--access-type")
        .arg("read")
        .output()
        .expect("Failed to execute healthcare consent check");

    assert!(output.status.success(), "Consent check should succeed");

    let stdout = String::from_utf8_lossy(&output.stdout);
    assert!(stdout.contains("pat-001"), "Should include patient ID");
    assert!(stdout.contains("Observation"), "Should show resource type");
    assert!(stdout.contains("true"), "Should show consent valid");
    assert!(stdout.contains("purpose_limited"), "Read access should have purpose_limited constraint");
    assert!(stdout.contains("anonymization"), "Read access should have anonymization constraint");
    assert!(stdout.contains("ODRL"), "Should mention ODRL policy");
}

/// Test: Consent check enforces write-access constraints
#[test]
fn test_consent_check_write_access() {
    let mut cmd = Command::cargo_bin("bos").expect("Failed to find bos binary");

    let output = cmd
        .arg("healthcare")
        .arg("consent")
        .arg("check")
        .arg("--patient-id")
        .arg("pat-001")
        .arg("--resource")
        .arg("MedicationRequest")
        .arg("--access-type")
        .arg("write")
        .output()
        .expect("Failed to execute healthcare consent check");

    assert!(output.status.success(), "Write consent check should succeed");

    let stdout = String::from_utf8_lossy(&output.stdout);
    assert!(stdout.contains("audit_required"), "Write access should require audit");
    assert!(stdout.contains("patient_notification"), "Write should require patient notification");
}

/// Test: Consent check enforces delete-access constraints
#[test]
fn test_consent_check_delete_access() {
    let mut cmd = Command::cargo_bin("bos").expect("Failed to find bos binary");

    let output = cmd
        .arg("healthcare")
        .arg("consent")
        .arg("check")
        .arg("--patient-id")
        .arg("pat-001")
        .arg("--access-type")
        .arg("delete")
        .output()
        .expect("Failed to execute healthcare consent check");

    assert!(output.status.success(), "Delete consent check should succeed");

    let stdout = String::from_utf8_lossy(&output.stdout);
    assert!(stdout.contains("legal_hold"), "Delete access should require legal hold");
    assert!(stdout.contains("encryption_required"), "Delete should require encryption");
}

/// Test: HIPAA verification confirms audit trail compliance
#[test]
fn test_hipaa_verify_compliance() {
    let mut cmd = Command::cargo_bin("bos").expect("Failed to find bos binary");

    let output = cmd
        .arg("healthcare")
        .arg("hipaa")
        .arg("verify")
        .arg("--organization")
        .arg("TestHospital")
        .arg("--audit-depth")
        .arg("1000")
        .output()
        .expect("Failed to execute healthcare hipaa verify");

    assert!(output.status.success(), "HIPAA verification should succeed");

    let stdout = String::from_utf8_lossy(&output.stdout);
    assert!(stdout.contains("TestHospital"), "Should include organization");
    assert!(stdout.contains("true"), "Should show compliant = true for adequate audit depth");
    assert!(stdout.contains("6"), "Should show 6-year retention requirement");
    assert!(stdout.contains("1000"), "Should show audit trail entries count");
    assert!(stdout.contains("microsecond"), "Should verify microsecond precision");
}

/// Test: HIPAA verification detects insufficient audit depth
#[test]
fn test_hipaa_verify_insufficient_audit() {
    let mut cmd = Command::cargo_bin("bos").expect("Failed to find bos binary");

    let output = cmd
        .arg("healthcare")
        .arg("hipaa")
        .arg("verify")
        .arg("--organization")
        .arg("SmallClinic")
        .arg("--audit-depth")
        .arg("50")
        .output()
        .expect("Failed to execute healthcare hipaa verify");

    assert!(output.status.success(), "HIPAA verification should complete");

    let stdout = String::from_utf8_lossy(&output.stdout);
    assert!(stdout.contains("insufficient audit trail"), "Should flag insufficient audit trail");
}

/// Test: PHI tracking includes audit entry count (lineage depth verification)
#[test]
fn test_phi_track_includes_audit_entries() {
    let mut cmd = Command::cargo_bin("bos").expect("Failed to find bos binary");

    let output = cmd
        .arg("healthcare")
        .arg("phi")
        .arg("track")
        .arg("--patient-id")
        .arg("pat-003")
        .arg("--phi-elements")
        .arg("name,ssn,diagnosis,medication,procedure")
        .output()
        .expect("Failed to execute healthcare phi track");

    assert!(output.status.success(), "PHI tracking should succeed");

    let stdout = String::from_utf8_lossy(&output.stdout);
    assert!(stdout.contains("audit"), "Should include audit entries");
    assert!(stdout.contains("lineage_depth"), "Should show lineage depth (PROV-O)");
}

/// Test: Integration - Healthcare init + PHI track workflow
#[test]
fn test_healthcare_workflow_init_then_track() {
    // Step 1: Initialize healthcare workspace
    let mut init_cmd = Command::cargo_bin("bos").expect("Failed to find bos binary");
    let init_output = init_cmd
        .arg("healthcare")
        .arg("init")
        .arg("--organization")
        .arg("ClinicA")
        .output()
        .expect("Failed to init healthcare");

    assert!(init_output.status.success(), "Initialization should succeed");

    // Step 2: Track PHI in the initialized workspace
    let mut track_cmd = Command::cargo_bin("bos").expect("Failed to find bos binary");
    let track_output = track_cmd
        .arg("healthcare")
        .arg("phi")
        .arg("track")
        .arg("--patient-id")
        .arg("pat-workflow-001")
        .output()
        .expect("Failed to track PHI");

    assert!(track_output.status.success(), "PHI tracking after init should succeed");

    let track_stdout = String::from_utf8_lossy(&track_output.stdout);
    assert!(track_stdout.contains("PHI-pat-workflow-001"), "Should have tracking ID from workflow");
}

#[test]
fn test_consent_check_default_resource() {
    let mut cmd = Command::cargo_bin("bos").expect("Failed to find bos binary");

    let output = cmd
        .arg("healthcare")
        .arg("consent")
        .arg("check")
        .arg("--patient-id")
        .arg("pat-001")
        .output()
        .expect("Failed to execute consent check");

    assert!(output.status.success(), "Consent check with defaults should succeed");

    let stdout = String::from_utf8_lossy(&output.stdout);
    assert!(stdout.contains("Observation"), "Should default to Observation resource");
}

#[test]
fn test_hipaa_verify_default_audit_depth() {
    let mut cmd = Command::cargo_bin("bos").expect("Failed to find bos binary");

    let output = cmd
        .arg("healthcare")
        .arg("hipaa")
        .arg("verify")
        .arg("--organization")
        .arg("ClinicB")
        .output()
        .expect("Failed to execute HIPAA verify");

    assert!(output.status.success(), "HIPAA verify with defaults should succeed");

    let stdout = String::from_utf8_lossy(&output.stdout);
    assert!(stdout.contains("1000"), "Should default to 1000 audit entries");
}
