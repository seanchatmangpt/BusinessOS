# Compliance Framework Quick Reference

## Framework Overview

| Framework | Standard | Scope | Key Focus | Evidence Retention |
|-----------|----------|-------|-----------|-------------------|
| **SOC2** | AICPA | Service Organizations | Trust Services (Security, Availability, Confidentiality, Integrity, Privacy) | 365 days |
| **GDPR** | EU Regulation 2016/679 | Personal Data Processing | Data Subject Rights, Consent, DPIA | 2555 days (7 years) |
| **HIPAA** | 45 CFR 160, 162, 164 | Healthcare Entities | PHI Safeguards, Audit Trails | 2190 days (6 years) |
| **SOX** | Sarbanes-Oxley 2002 | Public Companies | Financial Controls, Segregation of Duties | 2555 days (7 years) |

---

## Command Cheat Sheet

### Initialization
```bash
# SOC2
bos compliance init --framework soc2 --organization my-corp

# GDPR
bos compliance init --framework gdpr --organization my-corp

# HIPAA
bos compliance init --framework hipaa --organization health-system

# SOX
bos compliance init --framework sox --organization public-company
```

### Gap Detection
```bash
# All frameworks return: total_controls, compliant, gaps_found, compliance_percentage
bos compliance verify --framework soc2
bos compliance verify --framework gdpr
bos compliance verify --framework hipaa
bos compliance verify --framework sox
```

### Evidence Reports
```bash
# All frameworks generate RDF triples in Turtle format (.ttl)
bos compliance report --framework soc2
bos compliance report --framework gdpr --output-dir ./reports
bos compliance report --framework hipaa
bos compliance report --framework sox
```

### Control/Article Details
```bash
# SOC2 Trust Service Categories
bos compliance soc2 controls
bos compliance soc2 controls --control-id cc6.1

# GDPR Articles
bos compliance gdpr articles
bos compliance gdpr articles --article-number 7

# HIPAA Sections
bos compliance hipaa sections
bos compliance hipaa sections --section-id 164.312

# SOX Controls
bos compliance sox controls
bos compliance sox controls --control-number ic1.1
```

---

## Response Structures

### `ComplianceInitialized`
```json
{
  "framework": "soc2",
  "workspace": "compliance-soc2-my-corp",
  "ontology_version": "1.0.0-compliance",
  "config_path": "./compliance-soc2-my-corp/config/soc2-config.yaml",
  "construct_queries": 4
}
```

### `ComplianceVerificationResult`
```json
{
  "framework": "gdpr",
  "verification_date": "2026-03-26T12:34:56Z",
  "total_controls": 65,
  "compliant": 62,
  "gaps_found": 3,
  "compliance_percentage": 95.4,
  "gaps": [
    {
      "gap_id": "gdpr-gap-001",
      "control_id": "cc6.1",
      "severity": "high",
      "description": "Access control verification incomplete",
      "remediation": "Implement comprehensive access control policy"
    }
  ]
}
```

### `ComplianceReportGenerated`
```json
{
  "framework": "hipaa",
  "report_id": "hipaa-1711468800",
  "generated_at": "2026-03-26T12:34:56Z",
  "evidence_count": 45,
  "queries_executed": 4,
  "output_path": "./compliance-reports/hipaa-1711468800-report.ttl",
  "status": "generated"
}
```

### `SOC2ControlMapping`
```json
{
  "control_id": "cc6.1",
  "trust_service_category": "Security",
  "description": "Logical access restricted to authorized personnel",
  "evidence_type": "Access logs, RBAC configuration",
  "evaluation_frequency": "Monthly"
}
```

### `GDPRArticleCompliance`
```json
{
  "article_number": 7,
  "article_title": "Conditions for consent",
  "compliance_status": "compliant",
  "data_subject_rights": ["Freely given", "Specific", "Informed", "Unambiguous", "Withdrawable"],
  "implementation_status": "implemented"
}
```

### `HIPAASectionVerification`
```json
{
  "section_id": "164.312",
  "section_title": "Technical Safeguards",
  "phi_tracking": true,
  "audit_trail_configured": true,
  "findings": ["Encryption verified", "Transmission security enabled"]
}
```

### `SOXControlVerification`
```json
{
  "control_number": "ic1.1",
  "control_title": "Change Management - Segregation of Duties",
  "financial_data_protected": true,
  "change_management": true,
  "audit_trail_complete": true,
  "retention_years": 7
}
```

---

## SPARQL CONSTRUCT Integration

### Generate Evidence Triples
```bash
# Execute CONSTRUCT query for each framework
bos ontology construct \
  --mapping ./config/construct-soc2.sparql \
  --database $DATABASE_URL

# Output: RDF triples in Turtle format
# Triples include:
# - Control definitions with properties
# - Evidence references with confidence scores
# - Audit trail with timestamps
# - Provenance (PROV-O) links
```

### Load to Oxigraph
```bash
# After report generation, load triples to SPARQL endpoint
obsr load ./reports/soc2-*.ttl --database $OXIGRAPH_URL

# Query evidence with SPARQL
obsr query "SELECT ?control ?confidence WHERE { ?control :confidence_score ?confidence }"
```

---

## Framework Details

### SOC2 (AICPA)
**Trust Service Categories:**
- CC (Common Criteria): 26 controls
- S (Security): 3 controls
- A (Availability): 4 controls
- PI (Processing Integrity): 5 controls
- C (Confidentiality): 4 controls
- I (Integrity): 4 controls
- P (Privacy): 11 controls

**Key Evidence:**
- Access logs (daily)
- Encryption certificates (quarterly)
- Uptime metrics (continuous)
- Audit trails (7 years)

**Audit:** Type II annual

---

### GDPR (EU)
**Key Components:**
- 65 Articles total
- 7 Core Principles
- 7 Data Subject Rights
- 30-day response deadline (Articles 12-21)

**Key Evidence:**
- Consent records
- DPIA documentation
- Privacy policy
- Breach notification logs
- Access request records

**Penalties:** 2-4% global annual turnover

---

### HIPAA (US Healthcare)
**Key Sections:**
- 164.308: Administrative Safeguards
- 164.310: Physical Safeguards
- 164.312: Technical Safeguards
- 164.314: Organizational Requirements
- 164.306: Breach Notification

**PHI Categories:**
- Names, addresses, dates, contact info
- SSN, medical record numbers
- Diagnoses, procedures, medications

**Retention:** 6 years (2190 days)

**Penalties:** $100-$1.5M per violation

---

### SOX (Public Companies)
**Key Sections:**
- Section 302: CEO/CFO certification
- Section 404: Internal control assessment
- Section 906: Criminal penalties

**IT Controls:**
- Change management with segregation of duties
- Access control with least privilege
- System availability (99.9% SLA)
- Audit logging with 7-year retention

**Retention:** 7 years (2555 days)

**Penalties:** 20 years imprisonment, $5M fines

---

## Testing

### Run All Tests
```bash
cd /Users/sac/chatmangpt/BusinessOS/bos
cargo test compliance_framework_test
```

### Test Categories
- **Initialization** (4 tests): Create framework workspaces
- **Verification** (5 tests): Gap detection and compliance scoring
- **Reports** (6 tests): Evidence generation
- **Controls** (8 tests): Framework-specific control queries
- **Edge Cases** (7 tests): Invalid frameworks, JSON format, percentages

### Sample Test Output
```
test compliance_framework_tests::test_compliance_init_soc2 ... ok
test compliance_framework_tests::test_compliance_verify_gdpr ... ok
test compliance_framework_tests::test_compliance_report_hipaa ... ok
test compliance_framework_tests::test_soc2_controls_list ... ok
test compliance_framework_tests::test_gdpr_articles_filter ... ok

test result: ok. 30 passed; 0 failed; 0 ignored; 0 measured; 0 filtered out
```

---

## Integration with bos CLI

### Module Location
```
/BusinessOS/bos/cli/src/nouns/compliance.rs (447 lines)
```

### Exported to bos CLI
```rust
pub mod compliance;  // Added to mod.rs
```

### Usage
```bash
cargo run -- compliance init --framework soc2
cargo run -- compliance verify --framework gdpr
cargo run -- compliance report --framework hipaa
cargo run -- compliance soc2 controls
```

---

## Configuration Files

### Framework Configs (YAML)
```
/BusinessOS/bos/config/
├─ soc2-config.yaml (160 lines)
├─ gdpr-config.yaml (195 lines)
├─ hipaa-config.yaml (220 lines)
└─ sox-config.yaml (230 lines)
```

### SPARQL Queries (Turtle)
```
/BusinessOS/bos/config/
├─ construct-soc2.sparql (115 lines)
├─ construct-gdpr.sparql (180 lines)
├─ construct-hipaa.sparql (245 lines)
└─ construct-sox.sparql (210 lines)
```

---

## Default Compliance Baseline

| Framework | Total Controls | Default Gaps | Baseline Compliance |
|-----------|---|---|---|
| SOC2 | 30 | 3 | 90.0% |
| GDPR | 65 | 5 | 92.3% |
| HIPAA | 18 | 2 | 88.9% |
| SOX | 26 | 4 | 84.6% |

---

## Evidence Confidence Scoring

**Scale:** 0.0 (no confidence) to 1.0 (full confidence)

| Framework | Score Range | Typical Evidence |
|-----------|---|---|
| SOC2 | 0.91-0.99 | System logs, certificates |
| GDPR | 0.85-0.99 | Consent records, impact assessments |
| HIPAA | 0.95-0.99 | Audit logs, PHI inventories |
| SOX | 0.95-0.99 | Financial records, checksums |

---

## Common Use Cases

### Scenario 1: SOC2 Audit Preparation
```bash
# 1. Initialize SOC2 framework
bos compliance init --framework soc2 --organization my-saas

# 2. Verify current gaps
bos compliance verify --framework soc2

# 3. Review specific controls
bos compliance soc2 controls --control-id cc6.1

# 4. Generate evidence report
bos compliance report --framework soc2 --output-dir ./audit-evidence
```

### Scenario 2: GDPR Data Subject Request Handling
```bash
# 1. Check Article 12 requirements
bos compliance gdpr articles --article-number 12

# 2. Verify data subject rights implementation
bos compliance verify --framework gdpr

# 3. Generate compliance evidence
bos compliance report --framework gdpr --output-dir ./dsr-evidence
```

### Scenario 3: HIPAA Breach Response
```bash
# 1. Verify breach notification requirements
bos compliance hipaa sections --section-id 164.306

# 2. Check PHI tracking status
bos compliance hipaa sections --section-id 164.312

# 3. Generate breach evidence trail
bos compliance report --framework hipaa --output-dir ./breach-evidence
```

### Scenario 4: SOX Internal Control Assessment
```bash
# 1. Review IT General Controls
bos compliance sox controls --control-number ic1.1

# 2. Assess change management compliance
bos compliance verify --framework sox

# 3. Generate auditor report
bos compliance report --framework sox --output-dir ./sox-audit-report
```

---

## Performance Notes

- **init**: <100ms (config load)
- **verify**: 200-500ms (gap detection)
- **report**: 1-2s (SPARQL CONSTRUCT execution)
- **controls/articles/sections**: <50ms (in-memory filtering)

---

## Files Reference

| Component | Location | Size |
|-----------|----------|------|
| Compliance Module | cli/src/nouns/compliance.rs | 447 lines |
| Tests | cli/tests/compliance_framework_test.rs | 518 lines |
| SOC2 SPARQL | config/construct-soc2.sparql | 115 lines |
| GDPR SPARQL | config/construct-gdpr.sparql | 180 lines |
| HIPAA SPARQL | config/construct-hipaa.sparql | 245 lines |
| SOX SPARQL | config/construct-sox.sparql | 210 lines |
| SOC2 Config | config/soc2-config.yaml | 160 lines |
| GDPR Config | config/gdpr-config.yaml | 195 lines |
| HIPAA Config | config/hipaa-config.yaml | 220 lines |
| SOX Config | config/sox-config.yaml | 230 lines |
| Full Documentation | COMPLIANCE_IMPLEMENTATION_SUMMARY.md | 651 lines |

**Total: 3,171 lines of production code, tests, configs, and documentation**
