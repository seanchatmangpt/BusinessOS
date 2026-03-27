# Compliance SPARQL Queries Reference

**Version:** 1.0.0
**Last Updated:** 2026-03-26
**Status:** Stable
**Standard:** SPARQL 1.1
**Timeout:** 10 seconds maximum per query

---

## Overview

This document describes four SPARQL queries for compliance auditing in BusinessOS. These queries enable:

1. **Violation Detection** — Find compliance control failures
2. **Audit Trail Materialization** — Build complete PROV-O audit chains
3. **Compliance Metrics** — Compute compliance scores and trends
4. **Remediation Lookup** — Retrieve step-by-step remediation procedures

All queries are parameterized (accept input variables) and output deterministic results suitable for integration with compliance dashboards, audit reports, and regulatory filings.

---

## Compliance Framework Mappings

| Framework | Identifier | Scope | Controls | Status |
|-----------|-----------|-------|----------|--------|
| **SOC2** | `soc2.*` | System availability, security, confidentiality, integrity | CC (Common Criteria), A (Availability), C (Confidentiality), I (Integrity) | Fully supported |
| **GDPR** | `gdpr.*` | Data subject rights, consent, processing, residency | DS (Data Subject), CM (Consent), DPA (Agreement), DM (Minimization), DR (Residency) | Fully supported |
| **HIPAA** | `hipaa.*` | Protected health information safeguards | AC (Access), AU (Audit), IA (Integrity), AT (Awareness) | Fully supported |
| **SOX** | `sox.*` | Financial system controls, change management, audit | ITG (IT General), SA (System Availability), AL (Access Logging), CM (Change Management), FDI (Data Integrity) | Fully supported |

---

## Query 1: Compliance Violations

**File:** `compliance_violations.rq`
**Query Type:** SPARQL SELECT
**Purpose:** Discover control violations in a specific framework within a time window

### Input Parameters

```sparql
BIND("SOC2" AS ?framework)        # compliance framework
BIND(30 AS ?lookbackDays)         # days to look back from now
```

### Example Usage

```bash
# Via obsr CLI
obsr query -q compliance_violations.rq --param framework=SOC2 --param lookbackDays=30

# Via HTTP (if Oxigraph exposes SPARQL endpoint)
curl -X POST http://oxigraph:7878/query \
  -d @compliance_violations.rq \
  -H "Content-Type: application/sparql-query"
```

### Output Columns

| Column | Type | Description | Example |
|--------|------|-------------|---------|
| `violationId` | String | Unique violation identifier | `vio-audit-entry-2026-03-26` |
| `ruleId` | String | Compliance rule ID | `soc2.cc6.1` |
| `severity` | String | Severity level | `critical`, `high`, `medium`, `low` |
| `controlName` | String | Human-readable control name | `Logical access restricted to authorized personnel` |
| `description` | String | Violation description | `User "alice@example.com" accessed sensitive data without authorization` |
| `detectedAt` | DateTime | ISO8601 timestamp when violation was detected | `2026-03-26T14:30:00Z` |
| `affectedEntity` | String | Entity that failed (user ID, service ID, resource ID) | `user-123`, `service-api-prod` |
| `remediationUrl` | URL | Link to remediation procedure API endpoint | `https://chatmangpt.com/api/compliance/remediation/soc2.cc6.1` |

### Example Output

```csv
violationId,ruleId,severity,controlName,description,detectedAt,affectedEntity,remediationUrl
vio-audit-001-2026-03-26,soc2.cc6.1,critical,Logical access restricted,User alice accessed without permission,2026-03-26T14:30:00Z,user-alice,https://chatmangpt.com/api/compliance/remediation/soc2.cc6.1
vio-audit-002-2026-03-26,soc2.c1.1,high,Encryption at rest,Data not encrypted in cache,2026-03-25T09:15:00Z,cache-redis-prod,https://chatmangpt.com/api/compliance/remediation/soc2.c1.1
```

### Semantics

The query:
1. Filters audit entries by framework and recent date
2. Identifies entries with `checkResult = "FAILED"`
3. Groups violations by rule and affected entity
4. Sorts by severity (descending) and date (descending)
5. Limits to 1000 results

**Guarantee:** Deterministic output (same input always produces same result set in same order).

---

## Query 2: Compliance Audit Trail

**File:** `compliance_audit_trail.rq`
**Query Type:** SPARQL CONSTRUCT
**Purpose:** Materialize complete PROV-O audit chain showing who did what when and with what result

### Input Parameters

```sparql
BIND("SOC2" AS ?framework)
BIND("2026-01-01"^^xsd:date AS ?startDate)
BIND("2026-03-26"^^xsd:date AS ?endDate)
```

### Example Usage

```bash
# CONSTRUCT outputs RDF triples (Turtle format by default)
obsr query -q compliance_audit_trail.rq \
  --param framework=SOC2 \
  --param startDate=2026-01-01 \
  --param endDate=2026-03-26 \
  --format turtle > audit-trail.ttl

# Or insert into triplestore
obsr insert --file audit-trail.ttl --graph "https://chatmangpt.com/audit/soc2/2026-q1"
```

### Output Structure (RDF Triples)

The query constructs the following RDF graph structure:

```
Activity (control check, remediation, approval)
├─ prov:wasAssociatedWith → Agent (person or service)
├─ prov:used → ComplianceRule (input requirement)
├─ audit:result → Result (outcome)
├─ audit:affectedResource → Resource (what was checked)
├─ prov:wasInformedBy → PreviousActivity (audit chain)
└─ audit:wasRecordedBy → RecordingService

Agent (foaf:Person or custom service)
├─ foaf:name
├─ foaf:mbox (email)
└─ audit:agentId

Result (outcome of the activity)
├─ audit:resultStatus ("PASSED" | "FAILED" | "PARTIAL" | "SKIPPED")
├─ audit:resultMessage
├─ audit:checkedDate
├─ audit:evidenceCount
└─ audit:confidence (0.0 - 1.0)

ComplianceRule (requirement that was checked)
├─ audit:ruleId
├─ audit:ruleName
├─ audit:severity
├─ audit:framework
└─ audit:controlArea

Resource (entity that was checked/remediated)
├─ audit:resourceType
├─ audit:resourceId
├─ audit:resourceName
└─ prov:wasDerivedFrom (previous state)
```

### Ontology Namespaces

| Namespace | Prefix | URI |
|-----------|--------|-----|
| RDF | `rdf:` | http://www.w3.org/1999/02/22-rdf-syntax-ns# |
| RDFS | `rdfs:` | http://www.w3.org/2000/01/rdf-schema# |
| Dublin Core | `dcterms:` | http://purl.org/dc/terms/ |
| PROV-O | `prov:` | http://www.w3.org/ns/prov# |
| FOAF | `foaf:` | http://xmlns.com/foaf/0.1/ |
| BusinessOS | `bos:` | https://chatmangpt.com/ontology/businessos/ |
| Audit | `audit:` | https://chatmangpt.com/ontology/audit/ |
| Compliance | `comp:` | https://chatmangpt.com/ontology/compliance/ |

### Example Output (Turtle)

```turtle
@prefix prov: <http://www.w3.org/ns/prov#> .
@prefix audit: <https://chatmangpt.com/ontology/audit/> .
@prefix dcterms: <http://purl.org/dc/terms/> .
@prefix foaf: <http://xmlns.com/foaf/0.1/> .

<https://chatmangpt.com/audit/activity/check-001> a prov:Activity ;
  a audit:ComplianceActivity ;
  audit:activityType "control_check" ;
  audit:framework "SOC2" ;
  prov:startedAtTime "2026-03-26T10:00:00Z"^^xsd:dateTime ;
  prov:endedAtTime "2026-03-26T10:00:30Z"^^xsd:dateTime ;
  dcterms:creator <https://chatmangpt.com/agents/compliance-bot> ;
  dcterms:subject "soc2.cc6.1" ;
  prov:wasAssociatedWith <https://chatmangpt.com/agents/compliance-bot> ;
  audit:affectedResource <https://chatmangpt.com/resources/user-alice> ;
  audit:result <https://chatmangpt.com/audit/activity/check-001#result> ;
  audit:status "completed" ;
  audit:durationSeconds 30 ;
  rdfs:comment "Checked user permissions" .

<https://chatmangpt.com/audit/activity/check-001#result> a audit:Result ;
  audit:resultStatus "FAILED" ;
  audit:resultMessage "User lacks required role" ;
  audit:confidence 0.95 ;
  prov:wasGeneratedBy <https://chatmangpt.com/audit/activity/check-001> .

<https://chatmangpt.com/agents/compliance-bot> a prov:Agent ;
  foaf:name "ComplianceCheckBot" ;
  audit:agentId "bot-compliance-001" .
```

### Semantics

The query:
1. Filters audit entries by framework and date range
2. Links each activity to its agent (who performed it)
3. Links each activity to compliance rule (what was checked)
4. Links each activity to resource (what was affected)
5. Chains activities temporally (`prov:wasInformedBy`)
6. Attaches result nodes with outcome and evidence
7. Adds recording service provenance

**Guarantee:** Complete audit trail with full traceability. Suitable for SOC2 Type II audit reports.

---

## Query 3: Compliance Metrics

**File:** `compliance_metrics.rq`
**Query Type:** SPARQL SELECT
**Purpose:** Compute compliance scores, trends, and risk scores across control areas

### Input Parameters

```sparql
BIND("SOC2" AS ?framework)
BIND("month" AS ?period)              # "week", "month", "quarter"
BIND("2026-03-26"^^xsd:date AS ?currentPeriod)
BIND("2026-02-26"^^xsd:date AS ?previousPeriod)
```

### Example Usage

```bash
# Retrieve compliance metrics
obsr query -q compliance_metrics.rq \
  --param framework=GDPR \
  --param period=month \
  --param currentPeriod=2026-03-26 \
  --param previousPeriod=2026-02-26

# Output as JSON for dashboard
obsr query -q compliance_metrics.rq \
  --param framework=SOC2 \
  --format json > metrics.json
```

### Output Columns

| Column | Type | Description | Example |
|--------|------|-------------|---------|
| `framework` | String | Compliance framework | `SOC2`, `GDPR`, `SOX`, `HIPAA` |
| `period` | String | Reporting period | `week`, `month`, `quarter` |
| `controlArea` | String | Category of controls | `Access Control`, `Encryption`, `Audit & Monitoring`, `Change Management` |
| `totalControls` | Integer | Count of applicable controls in area | `15` |
| `passedControls` | Integer | Controls with recent PASSED checks | `14` |
| `failedControls` | Integer | Controls with recent FAILED checks | `1` |
| `skippedControls` | Integer | Controls deferred (N/A) | `0` |
| `compliancePercent` | Decimal | (passed / (total - skipped)) × 100 | `93.33` |
| `previousPercent` | Decimal | Compliance % in previous period | `90.00` |
| `trend` | String | `improving`, `stable`, `declining` | `improving` |
| `riskScore` | Decimal | 0-100 weighted by failures and severity | `6.67` |
| `actionItems` | Integer | Count of open remediation items | `2` |

### Example Output

```csv
framework,period,controlArea,totalControls,passedControls,failedControls,skippedControls,compliancePercent,previousPercent,trend,riskScore,actionItems
SOC2,month,Access Control,15,14,1,0,93.33,90.00,improving,6.67,2
SOC2,month,Encryption,10,10,0,0,100.00,100.00,stable,0.00,0
SOC2,month,Audit & Monitoring,12,11,1,0,91.67,83.33,improving,8.33,3
SOC2,month,Change Management,8,7,1,0,87.50,75.00,improving,12.50,1
GDPR,month,Consent Management,5,5,0,0,100.00,100.00,stable,0.00,0
GDPR,month,Data Subject Rights,4,3,1,0,75.00,50.00,improving,25.00,1
```

### Trend Calculation

```
Current Compliance 93.33%
Previous Compliance 90.00%
Delta = +3.33%

If Delta > +2.0% → "improving"
If -2.0% ≤ Delta ≤ +2.0% → "stable"
If Delta < -2.0% → "declining"
```

### Risk Score Calculation

```
Risk Score = 100 - Compliance Percent (with severity weighting)

Example:
  Compliance 93.33% → Risk Score 6.67
  Compliance 75.00% → Risk Score 25.00
  Compliance 100.00% → Risk Score 0.00

Weighting by severity:
  - Critical failures: +3x
  - High failures: +2x
  - Medium failures: +1x
  - Low failures: +0.5x
```

### Semantics

The query:
1. Aggregates control check results by control area
2. Calculates pass rate and compliance percentage
3. Compares against previous period
4. Calculates trend (improving, stable, declining)
5. Computes risk score (inverse of compliance)
6. Counts open remediation items

**Guarantee:** Deterministic metrics. Suitable for executive dashboards and compliance reports.

---

## Query 4: Remediation Steps

**File:** `remediation_steps.rq`
**Query Type:** SPARQL SELECT
**Purpose:** Retrieve step-by-step remediation procedures for a specific violation or rule

### Input Parameters

```sparql
BIND("vio-12345-2026-03-26" AS ?violationId)  # OR provide ruleId directly
BIND("soc2.cc6.1" AS ?ruleId)
```

### Example Usage

```bash
# Get remediation for a specific violation
obsr query -q remediation_steps.rq \
  --param violationId=vio-12345-2026-03-26

# Or by rule ID (query returns steps for all instances)
obsr query -q remediation_steps.rq \
  --param ruleId=gdpr.cm.1

# Output as JSON for workflow
obsr query -q remediation_steps.rq \
  --param ruleId=sox.itg.1 \
  --format json > remediation-workflow.json
```

### Output Columns

| Column | Type | Description | Example |
|--------|------|-------------|---------|
| `stepNumber` | Integer | Sequential order (1, 2, 3, ...) | `1` |
| `stepTitle` | String | Brief action title | `Enable MFA for all user accounts` |
| `stepDescription` | String | Detailed instructions | `Log into IAM console, navigate to MFA settings, enable TOTP for all active users` |
| `estimatedHours` | Decimal | Effort estimate in hours | `8.5` |
| `responsible` | String | Role or team | `Security Team`, `Database Admin`, `Finance Officer` |
| `dueDate` | Date | When remediation should complete | `2026-04-26` |
| `relatedResources` | String | Links to docs, scripts, tools | `https://internal.wiki/mfa-setup, https://github.com/tools/mfa-script` |
| `prerequisiteSteps` | String | Step numbers that must come first | `1, 2` |
| `successCriteria` | String | How to verify completion | `All users report MFA enabled in audit log` |

### Example Output

```csv
stepNumber,stepTitle,stepDescription,estimatedHours,responsible,dueDate,relatedResources,prerequisiteSteps,successCriteria
1,Assess current state,Audit all user accounts and identify non-MFA users,4.0,Security Team,2026-03-28,"https://wiki/audit-guide",,"Report: X users without MFA"
2,Enable MFA in IAM,Log into IAM, enable TOTP for MFA-capable services,6.0,Security Team,2026-04-02,"https://wiki/mfa-setup, https://github.com/mfa-script","1","All users receive TOTP QR codes"
3,User rollout,Distribute MFA credentials and train users,12.0,Support Team,2026-04-10,"https://wiki/user-training","2","Survey: >95% user adoption"
4,Verification and sign-off,Verify all users MFA-enabled in audit trail,2.0,Compliance Officer,2026-04-26,"https://wiki/verification","3","Signed audit report"
```

### ODRL Policy Mapping

Remediation steps are extracted from ODRL (Open Digital Rights Language) policies:

```
Compliance Rule
  ├─ odrl:hasPolicy → ODRL Policy
  │   ├─ odrl:hasPermission (action allowances)
  │   ├─ comp:remediationPlan → Remediation Plan
  │   │   ├─ comp:targetDate
  │   │   ├─ comp:priority (high, medium, low)
  │   │   └─ comp:hasStep → Step (1..N)
  │   │       ├─ comp:stepNumber
  │   │       ├─ dcterms:title
  │   │       ├─ rdfs:comment (description)
  │   │       ├─ comp:estimatedEffortHours
  │   │       ├─ comp:assignedTo (responsible)
  │   │       ├─ comp:dueDate
  │   │       ├─ comp:relatesTo → Resources
  │   │       ├─ comp:requires → Prerequisite Steps
  │   │       └─ comp:successCriteria
```

### Due Date Calculation by Severity

The query auto-calculates due dates based on rule severity:

| Severity | Auto Due Date | Override |
|----------|---|---|
| `critical` | NOW + 1 day | Query honors `comp:dueDate` if explicit |
| `high` | NOW + 3 days | Query honors `comp:dueDate` if explicit |
| `medium` | NOW + 7 days | Query honors `comp:dueDate` if explicit |
| `low` | NOW + 14 days | Query honors `comp:dueDate` if explicit |

### Semantics

The query:
1. Finds the violation or rule
2. Retrieves ODRL policy and remediation plan
3. Extracts steps in order
4. Collects related resources (docs, scripts, tools)
5. Identifies prerequisite dependencies
6. Calculates due dates by severity
7. Orders by step number

**Guarantee:** Complete remediation workflow. Suitable for tracking through issue management systems.

---

## Integration Patterns

### Pattern 1: Violation Detection + Remediation

```sparql
# 1. Find violations
SELECT ?violationId ?ruleId FROM compliance_violations

# 2. For each violation, get remediation steps
FOR EACH ?violation:
  SELECT ?stepNumber ?stepTitle FROM remediation_steps
  WHERE ?ruleId = ?violation.ruleId

# 3. Create remediation ticket
POST /api/remediation/create
  {
    "violation_id": ?violationId,
    "steps": [?stepNumber, ?stepTitle, ...]
  }
```

### Pattern 2: Monthly Compliance Report

```sparql
# 1. Get metrics for current month
SELECT ?framework ?controlArea ?compliancePercent ?trend
FROM compliance_metrics

# 2. Construct audit trail for report
CONSTRUCT {
  ?activity prov:wasAssociatedWith ?agent
  ?activity audit:result ?result
}
FROM compliance_audit_trail

# 3. Render dashboard/PDF with trends and audit trail
RENDER compliance_dashboard.html
  {framework, metrics, audit_trail}
```

### Pattern 3: Trend Analysis (Week-over-Week)

```sparql
# Run compliance_metrics three times:
# - Week N (current)
# - Week N-1 (previous)
# - Week N-2 (two weeks ago)

# Plot as time series
BIND(?compliancePercent AS Y)
BIND(?period AS X)

# Dashboard shows:
# - Line chart: compliance % over time
# - Table: trend per control area (improving/stable/declining)
# - Risk heatmap: high-risk areas highlighted
```

---

## Execution and Performance

### Query Timeout

All queries have a **10-second timeout**. If a query exceeds 10 seconds:
1. Oxigraph cancels execution
2. Partial results returned (if streaming) or empty result set
3. Increase timeout in config: `query_timeout_ms: 15000`

### Index Requirements

For optimal performance, create these indices:

```sparql
# Index 1: Audit entries by framework and date
CREATE INDEX idx_audit_framework_date
ON bos:AuditEntry (bos:checkFramework, dcterms:issued DESC)

# Index 2: Violations by rule
CREATE INDEX idx_violation_ruleId
ON bos:Violation (bos:ruleId)

# Index 3: Resources by type
CREATE INDEX idx_resource_type
ON audit:Resource (audit:resourceType)

# Index 4: Remediation items by status
CREATE INDEX idx_remediation_status
ON bos:RemediationItem (bos:status)
```

### Estimated Cardinality

For a mid-size enterprise:

| Query | Data Size | Typical Runtime |
|-------|-----------|-----------------|
| `compliance_violations` | 10K audit entries, 50 rules | 200ms |
| `compliance_audit_trail` | 10K entries × 10 quads = 100K triples | 800ms |
| `compliance_metrics` | 100 control areas × 50 frameworks | 400ms |
| `remediation_steps` | 1 rule × 10-20 steps | 50ms |

---

## Parameterization

All queries use SPARQL parameterization via `BIND()`:

```sparql
# Safe: No injection risk
BIND("SOC2" AS ?framework)
BIND("2026-03-26"^^xsd:date AS ?startDate)

# NOT SAFE: String interpolation (do NOT use)
BIND("" + ?frameworkInput + "" AS ?framework)  # ❌ Wrong
```

### Parameter Types

| Type | Example | SPARQL |
|------|---------|--------|
| String | `"SOC2"` | `BIND("SOC2" AS ?x)` |
| Date | `2026-03-26` | `BIND("2026-03-26"^^xsd:date AS ?x)` |
| DateTime | `2026-03-26T10:00:00Z` | `BIND("2026-03-26T10:00:00Z"^^xsd:dateTime AS ?x)` |
| Integer | `30` | `BIND(30 AS ?x)` |
| Decimal | `0.95` | `BIND(0.95 AS ?x)` |
| Boolean | `true` | `BIND(true AS ?x)` |

---

## Troubleshooting

### Q1: Query times out

**A:** Increase timeout in Oxigraph config or add indices. See "Index Requirements" above.

### Q2: Empty results

**A:** Check:
1. Data exists in triplestore: `SELECT (COUNT(*) AS ?count) WHERE { ?s ?p ?o }`
2. Framework matches: verify `bos:checkFramework` values
3. Date range: ensure audit entries fall within `?startDate` to `?endDate`

### Q3: Query returns duplicates

**A:** SPARQL CONSTRUCT may duplicate triples if multiple patterns match. Use `DISTINCT` in SELECT or deduplicate on output:

```bash
# Via RDF tools
rapper audit-trail.ttl --quiet | sort | uniq > audit-trail-unique.ttl

# Or in SPARQL 1.1
SELECT DISTINCT ?s ?p ?o WHERE { ?s ?p ?o }
```

### Q4: How to filter by multiple frameworks?

**A:** Use `VALUES`:

```sparql
VALUES ?framework { "SOC2" "GDPR" "HIPAA" }
?audit bos:checkFramework ?framework .
```

---

## References

- **SPARQL 1.1 W3C Spec:** https://www.w3.org/TR/sparql11-query/
- **PROV-O (W3C):** https://www.w3.org/TR/prov-o/
- **ODRL 2.0:** https://www.w3.org/ns/odrl/2/
- **Oxigraph SPARQL Endpoint:** https://github.com/oxigraph/oxigraph
- **BusinessOS Compliance Rules:** `BusinessOS/config/compliance-rules.yaml`

---

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2026-03-26 | Initial release: 4 queries, full framework support |

---

## Support

For issues or questions:
- **GitHub:** https://github.com/seanchatmangpt/chatmangpt/issues
- **Email:** compliance@chatmangpt.com
- **Slack:** #compliance-audit-trail
