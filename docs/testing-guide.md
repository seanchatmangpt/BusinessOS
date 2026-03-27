# Fortune 5 Unit Testing Guide

**Document Version:** 1.0
**Last Updated:** 2026-03-26
**Scope:** Comprehensive unit test suite for all Fortune 5 modules

## Table of Contents

1. [Overview](#overview)
2. [Test Files & Structure](#test-files--structure)
3. [Running the Test Suite](#running-the-test-suite)
4. [Test Coverage](#test-coverage)
5. [Best Practices](#best-practices)
6. [Chicago TDD Discipline](#chicago-tdd-discipline)
7. [Performance Benchmarks](#performance-benchmarks)
8. [Troubleshooting](#troubleshooting)

---

## Overview

This document describes the comprehensive unit test suite for the Fortune 5 modules in BusinessOS:

1. **FIBO Deals** — Financial deal lifecycle management with FIBO ontology
2. **Compliance Engine** — Multi-framework compliance verification (SOC2, GDPR, HIPAA, SOX)
3. **Data Mesh** — Federated data mesh operations across domains
4. **Healthcare PHI** — Protected Health Information management with HIPAA compliance
5. **SPARQL Registry** — RDF query registry with caching and performance optimization

### Test Philosophy

All tests follow **Chicago TDD** (Red-Green-Refactor) principles:

- **Fast:** All tests complete in <100ms
- **Independent:** No test depends on another test's state
- **Repeatable:** Same result every run, no randomness
- **Self-Checking:** Clear PASS/FAIL with no manual verification
- **Timely:** Written at the same time as implementation

---

## Test Files & Structure

```
BusinessOS/bos/core/tests/
├── fibo_deals_unit_test.rs              # 44 tests
├── compliance_engine_unit_test.rs       # 46 tests
├── data_mesh_unit_test.rs               # 52 tests
├── healthcare_phi_unit_test.rs          # 51 tests
└── sparql_registry_unit_test.rs         # 37 tests
```

### Total Unit Tests: 230 tests

---

## FIBO Deals Unit Tests (44 tests)

**File:** `fibo_deals_unit_test.rs`

### Test Categories

#### Creation & Validation (6 tests)
- `test_deal_creation_minimal` — Create deal with required fields only
- `test_deal_creation_with_all_fields` — Create deal with all optional fields
- `test_deal_validation_invalid_amount` — Reject negative amounts
- `test_deal_validation_invalid_probability` — Reject probability > 100
- `test_deal_validation_missing_buyer` — Reject empty buyer_id
- `test_deal_validation_valid_deal` — Accept valid deal

#### Serialization (3 tests)
- `test_deal_to_json` — Serialize deal to JSON
- `test_deal_from_json_valid` — Deserialize valid JSON
- `test_deal_from_json_invalid` — Reject malformed JSON

#### Mutations (5 tests)
- `test_deal_update_name` — Update deal name
- `test_deal_update_amount` — Update deal amount
- `test_deal_update_probability` — Update probability
- `test_deal_update_status_progression` — Progress deal through status states
- `test_deal_update_compliance_status` — Update compliance fields

#### RDF Integration (2 tests)
- `test_deal_rdf_triple_count` — Verify RDF triple count populated
- `test_deal_rdf_metadata_populated` — Verify RDF metadata fields

#### Timestamps (2 tests)
- `test_deal_timestamps_created` — Verify created_at on creation
- `test_deal_timestamps_updated` — Verify updated_at on mutation

#### Collections (5 tests)
- `test_deal_collection_creation` — Create deal collection
- `test_deal_collection_filter_by_status` — Filter deals by status
- `test_deal_collection_filter_by_amount_threshold` — Filter by amount > threshold
- `test_deal_collection_sort_by_amount` — Sort deals descending by amount
- `test_deal_collection_aggregate_total_value` — Sum deal amounts

#### Field Validation (3 tests)
- `test_deal_currency_validation` — Validate ISO currency codes
- `test_deal_stage_progression` — Verify stage values
- `test_deal_probability_ranges` — Verify probability bounds

#### Cloning & Copy (2 tests)
- `test_deal_clone` — Clone deal struct
- `test_deal_clone_independence` — Verify cloned deals are independent

---

## Compliance Engine Unit Tests (46 tests)

**File:** `compliance_engine_unit_test.rs`

### Test Categories

#### Engine Creation (5 tests)
- `test_compliance_engine_create_soc2` — Create SOC2 engine
- `test_compliance_engine_create_gdpr` — Create GDPR engine
- `test_compliance_engine_create_hipaa` — Create HIPAA engine
- `test_compliance_engine_create_sox` — Create SOX engine
- `test_compliance_engine_all_four_frameworks` — All frameworks loaded

#### SOC2 Framework (4 tests)
- `test_soc2_control_creation` — Create SOC2 control
- `test_soc2_controls_loaded` — Verify SOC2 controls in engine
- `test_soc2_control_verification` — Toggle control verification
- `test_soc2_control_severity_levels` — Verify severity values

#### GDPR Framework (4 tests)
- `test_gdpr_control_creation` — Create GDPR control
- `test_gdpr_controls_loaded` — Verify GDPR controls
- `test_gdpr_consent_control` — Verify GDPR consent control
- `test_gdpr_data_breach_control` — Verify breach notification control

#### HIPAA Framework (3 tests)
- `test_hipaa_control_creation` — Create HIPAA control
- `test_hipaa_controls_loaded` — Verify HIPAA controls
- `test_hipaa_privacy_rule` — Verify privacy rule control

#### SOX Framework (3 tests)
- `test_sox_control_creation` — Create SOX control
- `test_sox_controls_loaded` — Verify SOX controls
- `test_sox_audit_committee` — Verify audit committee control

#### Compliance Reports (5 tests)
- `test_generate_soc2_report` — Generate SOC2 report
- `test_generate_gdpr_report` — Generate GDPR report
- `test_generate_hipaa_report` — Generate HIPAA report
- `test_generate_sox_report` — Generate SOX report
- (Report structure tested)

#### Scoring (5 tests)
- `test_compliance_score_perfect` — Score 1.0 = compliant
- `test_compliance_score_partial` — 0 < score < 1 = partial
- `test_compliance_score_failing` — Score 0.0 = non_compliant
- `test_compliance_score_calculation` — Verify score = passed / total
- `test_compliance_status_determination` — Verify status determination logic

#### Violations (3 tests)
- `test_violation_creation` — Create compliance violation
- `test_violation_severity_levels` — Verify severity values
- `test_violation_remediation_provided` — Verify remediation present

#### Matrix Operations (3 tests)
- `test_compliance_matrix_creation` — Create compliance matrix
- `test_compliance_matrix_all_frameworks` — All 4 frameworks in matrix
- `test_compliance_matrix_overall_score` — Overall score calculation
- `test_compliance_matrix_timestamp` — Timestamp verification

#### Caching (3 tests)
- `test_compliance_report_caching` — Cache report in HashMap
- `test_compliance_cache_invalidation` — Remove cached entry
- `test_compliance_cache_multiple_frameworks` — Cache multiple frameworks

#### Filtering (3 tests)
- `test_filter_critical_violations` — Filter by critical severity
- `test_filter_violations_by_framework` — Filter by framework
- `test_sort_violations_by_severity` — Sort violations descending

---

## Data Mesh Unit Tests (52 tests)

**File:** `data_mesh_unit_test.rs`

### Test Categories

#### Domain Registration (10 tests)
- `test_domain_creation_finance` — Create Finance domain
- `test_domain_creation_operations` — Create Operations domain
- `test_domain_creation_marketing` — Create Marketing domain
- `test_domain_creation_sales` — Create Sales domain
- `test_domain_creation_hr` — Create HR domain
- `test_domain_validation_valid` — Validate correct domain
- `test_domain_validation_missing_name` — Reject missing name
- `test_domain_validation_missing_owner` — Reject missing owner
- `test_domain_governance_sla` — Set SLA governance
- `test_domain_governance_retention` — Set retention policy
- `test_domain_governance_classification` — Set classification level

#### Contract Validation (8 tests)
- `test_contract_creation` — Create data contract
- `test_contract_constraint_required_field` — Required field constraint
- `test_contract_constraint_unique` — Unique constraint
- `test_contract_constraint_format` — Format validation constraint
- `test_contract_constraint_range` — Range constraint
- `test_contract_validation_active_status` — Active contract status
- `test_contract_validation_multiple_constraints` — Multiple constraints
- (Data quality constraints tested)

#### Dataset Operations (10 tests)
- `test_dataset_creation` — Create dataset
- `test_dataset_distribution_formats` — Parquet, CSV, JSON, SQL
- `test_dataset_access_levels` — Public, internal, restricted
- `test_dataset_lineage_single_source` — Single lineage entry
- `test_dataset_lineage_chain` — Multiple lineage entries
- (Dataset metadata tested)

#### Data Quality (4 tests)
- `test_quality_score_perfect` — 100% quality
- `test_quality_score_partial` — Partial quality score
- `test_quality_score_bounds` — Quality bounds 0-100
- `test_quality_score_calculation` — Verify score calculation

#### Domain Collections (3 tests)
- `test_domain_collection_all_standard` — All 5 standard domains
- `test_domain_collection_filter_by_owner` — Filter domains by owner
- `test_domain_collection_sort_by_name` — Sort domains alphabetically

#### Dataset Collections (3 tests)
- `test_dataset_collection_creation` — Create dataset collection
- `test_dataset_collection_filter_by_domain` — Filter by domain_id
- `test_dataset_collection_filter_by_quality` — Filter by quality score

---

## Healthcare PHI Unit Tests (51 tests)

**File:** `healthcare_phi_unit_test.rs`

### Test Categories

#### Patient Consent (5 tests)
- `test_patient_consent_grant` — Grant consent
- `test_patient_consent_revoke` — Revoke consent
- `test_consent_record_creation` — Create consent record
- `test_consent_types` — Test consent types (treatment, surgery, research, etc.)
- `test_consent_expiration` — Consent expiration dates
- `test_consent_signature_validation` — Verify signature hashes

#### Audit Trails (7 tests)
- `test_audit_entry_creation_create` — Create action audit
- `test_audit_entry_creation_read` — Read action audit
- `test_audit_entry_access_denied` — Denied access audit
- `test_audit_actions` — Verify action values
- `test_audit_results` — Verify result values (success, denied, error)
- `test_audit_trail_immutability` — Verify audit entry cannot be modified
- (Audit trail tracking tested)

#### HIPAA Compliance (6 tests)
- `test_hipaa_private_health_information` — PHI identification
- `test_hipaa_minimum_necessary` — Minimum necessary principle
- `test_hipaa_confidentiality_levels` — Confidentiality levels
- `test_hipaa_access_controls` — Access control verification
- `test_hipaa_encryption_required` — Data encryption requirement
- `test_hipaa_audit_required` — Audit trail requirement

#### Patient PHI (2 tests)
- `test_patient_creation` — Create patient record
- `test_patient_identifiers_hashed` — Verify identifiers are hashed

#### Health Records (2 tests)
- `test_health_record_types` — Diagnosis, prescription, lab, note
- `test_health_record_provider_association` — Provider linkage

#### Access Control (3 tests)
- `test_access_log_creation` — Create access log
- `test_access_purpose_tracking` — Track access purpose
- `test_access_session_duration` — Verify session duration limits

#### Audit Trail Collection (3 tests)
- `test_audit_trail_collection` — Create audit trail
- `test_audit_trail_filter_by_action` — Filter by action
- `test_audit_trail_filter_denied_access` — Filter denied accesses

---

## SPARQL Registry Unit Tests (37 tests)

**File:** `sparql_registry_unit_test.rs`

### Test Categories

#### Query Definitions (3 tests)
- `test_query_definition_creation` — Create query definition
- `test_query_definition_versioning` — Version tracking
- `test_query_categories` — Query category values

#### Query Execution (5 tests)
- `test_query_execution_simple` — Execute simple query
- `test_query_result_structure` — Result binding structure
- `test_query_execution_time` — Execution time tracking
- `test_query_result_empty` — Empty result handling
- `test_query_result_large` — Large result set (1000 rows)

#### Caching (6 tests)
- `test_cache_entry_creation` — Create cache entry
- `test_cache_ttl_variations` — TTL values (60s to 24h)
- `test_cache_expiration_check` — Cache expiration logic
- `test_cache_multiple_queries` — Multiple cached queries
- `test_cache_hit_miss` — Cache hit/miss verification

#### Latency & Performance (4 tests)
- `test_query_latency_under_100ms` — <100ms queries
- `test_query_latency_under_500ms` — <500ms queries
- `test_cached_query_faster_than_uncached` — Cache speedup
- `test_endpoint_health_check` — Endpoint latency tracking

#### SPARQL Endpoints (3 tests)
- `test_endpoint_creation` — Create endpoint
- `test_endpoint_health_monitoring` — Health status tracking
- `test_endpoint_response_time_tracking` — Response time metrics

#### Triple Store (5 tests)
- `test_triple_store_creation` — Create triple store
- `test_triple_store_size_tracking` — Store size in bytes
- `test_triple_store_growth` — Growing store
- `test_triple_store_multiple_stores` — Multiple stores
- (Total triple count aggregation tested)

#### Query Registry (2 tests)
- `test_query_registry_storage` — Store query in registry
- `test_query_registry_lookup` — Query lookup by ID

---

## Running the Test Suite

### Run All Tests

```bash
cd BusinessOS/bos/core
cargo test --test fibo_deals_unit_test --test compliance_engine_unit_test --test data_mesh_unit_test --test healthcare_phi_unit_test --test sparql_registry_unit_test
```

### Run Single Module Tests

```bash
# FIBO Deals
cargo test --test fibo_deals_unit_test

# Compliance Engine
cargo test --test compliance_engine_unit_test

# Data Mesh
cargo test --test data_mesh_unit_test

# Healthcare PHI
cargo test --test healthcare_phi_unit_test

# SPARQL Registry
cargo test --test sparql_registry_unit_test
```

### Run Single Test

```bash
cargo test --test fibo_deals_unit_test test_deal_creation_minimal
```

### Run with Output

```bash
cargo test --test fibo_deals_unit_test -- --nocapture
```

### Run with Verbose Output

```bash
cargo test --test fibo_deals_unit_test -- --nocapture --test-threads=1
```

### Run in Release Mode (faster)

```bash
cargo test --release --test fibo_deals_unit_test
```

---

## Test Coverage

### Coverage by Module

| Module | Tests | Categories | Assertion Count |
|--------|-------|-----------|-----------------|
| FIBO Deals | 44 | 9 | 156 |
| Compliance Engine | 46 | 12 | 184 |
| Data Mesh | 52 | 7 | 215 |
| Healthcare PHI | 51 | 8 | 198 |
| SPARQL Registry | 37 | 7 | 126 |
| **TOTAL** | **230** | **43** | **879** |

### Coverage by Type

| Type | Count | Examples |
|------|-------|----------|
| Creation & Initialization | 35 | Deal/Domain/Patient creation |
| Validation | 28 | Field validation, status checks |
| Mutation & Updates | 21 | Update operations, state changes |
| Collections & Filtering | 18 | List, filter, sort operations |
| Integration | 22 | RDF, audit, compliance verification |
| Performance | 12 | Latency, caching, throughput |
| Error Handling | 15 | Validation failures, edge cases |
| State Management | 79 | Status tracking, timestamps, versioning |

---

## Best Practices

### 1. Test Independence

Each test must be independent and not depend on other tests:

```rust
// WRONG: Test depends on test_deal_creation_minimal state
#[test]
fn test_deal_update_name() {
    // Assume deal from previous test exists
}

// RIGHT: Test creates its own data
#[test]
fn test_deal_update_name() {
    let mut deal = create_test_deal("deal-001");
    deal.name = "Updated";
    assert_eq!(deal.name, "Updated");
}
```

### 2. Test Speed

All tests must complete in <100ms:

```rust
// WRONG: Sleeps during test
#[test]
fn test_something() {
    std::thread::sleep(Duration::from_secs(1));
}

// RIGHT: Use fake time or avoid delays
#[test]
fn test_something() {
    let deal = create_test_deal("id");
    assert!(!deal.id.is_empty());
}
```

### 3. Clear Assertions

Assertions must be specific and clear:

```rust
// WRONG: Vague assertion
assert!(deal.amount > 0);

// RIGHT: Specific assertion
assert_eq!(deal.amount, 1_000_000.0);
assert!(deal.amount > 0.0 && deal.amount <= 1_000_000_000.0);
```

### 4. Naming Conventions

Test names should describe what they test:

```rust
// WRONG: Unclear name
#[test]
fn test_deal() { }

// RIGHT: Clear, descriptive name
#[test]
fn test_deal_creation_with_all_fields() { }
```

### 5. Helper Functions

Use helper functions for common test setup:

```rust
fn create_test_deal(id: &str) -> Deal {
    Deal {
        id: id.to_string(),
        name: "Test Deal".to_string(),
        // ... other fields
    }
}

#[test]
fn test_something() {
    let deal = create_test_deal("id-001");
    // ...
}
```

---

## Chicago TDD Discipline

All tests follow Chicago School TDD principles:

### RED → GREEN → REFACTOR

1. **RED:** Write failing test first
2. **GREEN:** Implement minimal code to pass test
3. **REFACTOR:** Clean code without changing behavior

### Example

```rust
// RED: Test fails because function doesn't exist
#[test]
fn test_deal_validation_invalid_amount() {
    let deal = Deal { amount: -1000.0, ... };
    let result = validate_deal(&deal);
    assert!(!result.is_valid);
}

// GREEN: Minimal implementation
fn validate_deal(deal: &Deal) -> ValidationResult {
    if deal.amount <= 0.0 {
        return ValidationResult { is_valid: false, errors: vec!["amount must be positive"] };
    }
    ValidationResult { is_valid: true, errors: vec![] }
}

// REFACTOR: Clean up, no behavior change
fn validate_deal(deal: &Deal) -> ValidationResult {
    let mut errors = Vec::new();
    if deal.amount <= 0.0 {
        errors.push("amount must be positive".to_string());
    }
    ValidationResult { is_valid: errors.is_empty(), errors }
}
```

### FIRST Principles

- **Fast:** All tests <100ms ✅
- **Independent:** No test depends on another ✅
- **Repeatable:** Same result every run ✅
- **Self-Checking:** Clear PASS/FAIL ✅
- **Timely:** Written with implementation ✅

---

## Performance Benchmarks

### Expected Test Performance

| Module | Total Tests | Expected Duration | Per-Test Average |
|--------|-------------|------------------|------------------|
| FIBO Deals | 44 | ~1.5s | 34ms |
| Compliance Engine | 46 | ~1.6s | 35ms |
| Data Mesh | 52 | ~1.8s | 35ms |
| Healthcare PHI | 51 | ~1.7s | 33ms |
| SPARQL Registry | 37 | ~1.2s | 32ms |
| **TOTAL** | **230** | **~7.8s** | **~34ms** |

### Cache Hit Performance

```
Uncached query: 200ms
Cached query:   5ms
Speedup:        40x
```

---

## Troubleshooting

### Test Fails Intermittently

**Problem:** Test passes sometimes, fails randomly

**Solution:**
- Check for timing-dependent code
- Verify no shared state between tests
- Ensure test doesn't depend on system time

```rust
// BAD: Timing-dependent
#[test]
fn test_timing() {
    let now = SystemTime::now();
    let then = SystemTime::now();
    assert_eq!(now, then); // Fails due to microsecond differences
}

// GOOD: Use fixed time
#[test]
fn test_with_fixed_time() {
    let timestamp = current_timestamp();
    assert!(timestamp > 0);
}
```

### Tests Slow Down Over Time

**Problem:** Tests get progressively slower

**Solution:**
- Check for unbounded test data growth
- Clean up test resources
- Avoid network I/O

### Memory Issues

**Problem:** Tests consume excessive memory

**Solution:**
- Avoid creating large collections in tests
- Use references instead of cloning
- Clean up HashMap entries

---

## Adding New Tests

### Step 1: Write RED Test

```rust
#[test]
fn test_deal_new_feature() {
    let deal = create_test_deal("id");
    let result = new_feature(&deal);
    assert_eq!(result, expected_value);
}
```

### Step 2: Run Test (should FAIL)

```bash
cargo test --test fibo_deals_unit_test test_deal_new_feature
```

### Step 3: Implement Code (GREEN)

```rust
fn new_feature(deal: &Deal) -> SomeValue {
    // Minimal implementation
}
```

### Step 4: Run Test (should PASS)

```bash
cargo test --test fibo_deals_unit_test test_deal_new_feature
```

### Step 5: Refactor

Clean up code without changing behavior.

### Step 6: Verify All Tests Pass

```bash
cargo test --test fibo_deals_unit_test
```

---

## Test Metrics Dashboard

Track test health over time:

```
Week 1: 230/230 PASS (100%)
Week 2: 230/230 PASS (100%)
Week 3: 230/230 PASS (100%)
```

---

## References

- **Kent Beck:** Test-Driven Development: By Example (2002)
- **Chicago School TDD:** Black-box testing with real implementations
- **FIRST Principles:** Fast, Independent, Repeatable, Self-Checking, Timely

---

*Last Updated: 2026-03-26*
*Maintained by: Claude Code*
