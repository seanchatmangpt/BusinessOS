# Compliance SPARQL Queries — Complete Index

**Agent 37: Semantic Web Compliance Auditing**
**Delivered:** 2026-03-26
**Status:** ✅ Complete

---

## Quick Navigation

| Document | Purpose | Audience | Read Time |
|----------|---------|----------|-----------|
| **[README.md](README.md)** | Quick start guide (3 execution options, 5 tests, troubleshooting) | Operators, engineers | 15 min |
| **[MANIFEST.md](MANIFEST.md)** | Deliverables checklist (all 7 files, standards compliance, 80/20 coverage) | Project leads, QA | 10 min |
| **[../docs/compliance-queries.md](../docs/compliance-queries.md)** | Complete SPARQL reference (618 lines, all patterns, integration examples) | Developers, auditors | 45 min |

---

## The 4 SPARQL Queries

### Query 1: Violations Discovery
**File:** [compliance_violations.rq](compliance_violations.rq)
**Type:** SELECT (CSV output)
**Size:** 65 lines
**Purpose:** Find compliance control failures by framework and date

```sparql
-- Discover violations in the last 30 days
SELECT ?violationId ?ruleId ?severity ?description ?detectedAt
WHERE {
  ?auditEntry bos:checkFramework "SOC2" ;
    dcterms:issued ?detectedAt ;
    bos:checkResult "FAILED" .
}
ORDER BY DESC(?severity) DESC(?detectedAt)
```

**Quick Example:**
```bash
obsr query -q compliance_violations.rq --param framework=SOC2 --param lookbackDays=30
```

**Output:**
```csv
violationId,ruleId,severity,controlName,description,detectedAt,affectedEntity,remediationUrl
vio-001-2026-03-26,soc2.cc6.1,critical,Logical access,User alice unauth access,2026-03-26T14:30Z,user-alice,https://...
```

---

### Query 2: Audit Trail Materialization
**File:** [compliance_audit_trail.rq](compliance_audit_trail.rq)
**Type:** CONSTRUCT (RDF output)
**Size:** 151 lines
**Purpose:** Build PROV-O graph showing who did what when

```sparql
-- Materialize complete audit chain with PROV-O
CONSTRUCT {
  ?activity a prov:Activity ;
    prov:wasAssociatedWith ?agent ;
    audit:result ?resultNode ;
    audit:affectedResource ?resource .
  ?resultNode a audit:Result ;
    audit:resultStatus ?resultStatus .
}
WHERE {
  ?activity bos:auditType "control_check" ;
    dcterms:creator ?agentId ;
    bos:checkResult ?resultStatus .
}
```

**Quick Example:**
```bash
obsr query -q compliance_audit_trail.rq \
  --param framework=SOC2 \
  --param startDate=2026-01-01 \
  --param endDate=2026-03-26 \
  --format turtle > audit-trail.ttl
```

**Output (Turtle RDF):**
```turtle
<https://chatmangpt.com/audit/activity/check-001> a prov:Activity ;
  prov:wasAssociatedWith <https://chatmangpt.com/agents/compliance-bot> ;
  audit:result <https://chatmangpt.com/audit/activity/check-001#result> .

<https://chatmangpt.com/audit/activity/check-001#result> a audit:Result ;
  audit:resultStatus "FAILED" ;
  prov:wasGeneratedBy <https://chatmangpt.com/audit/activity/check-001> .
```

---

### Query 3: Compliance Metrics
**File:** [compliance_metrics.rq](compliance_metrics.rq)
**Type:** SELECT (CSV output)
**Size:** 153 lines
**Purpose:** Compute compliance scores and trends

```sparql
-- Calculate compliance %, trend, and risk score
SELECT ?controlArea ?compliancePercent ?trend ?riskScore ?actionItems
WHERE {
  # Count passed/failed controls by area
  OPTIONAL {
    ?auditEntry bos:checkFramework "SOC2" ;
      bos:controlArea ?controlArea ;
      bos:checkResult ?checkResult .
  }
}
GROUP BY ?controlArea
```

**Quick Example:**
```bash
obsr query -q compliance_metrics.rq \
  --param framework=GDPR \
  --param period=month \
  --param currentPeriod=2026-03-26 \
  --param previousPeriod=2026-02-26
```

**Output:**
```csv
framework,controlArea,compliancePercent,trend,riskScore,actionItems
GDPR,Consent Management,100.00,stable,0.00,0
GDPR,Data Residency,85.00,improving,15.00,2
```

---

### Query 4: Remediation Steps
**File:** [remediation_steps.rq](remediation_steps.rq)
**Type:** SELECT (CSV output)
**Size:** 119 lines
**Purpose:** Get step-by-step remediation workflow

```sparql
-- Retrieve remediation steps for a violation
SELECT ?stepNumber ?stepTitle ?description ?estimatedHours ?dueDate
WHERE {
  ?rule bos:ruleId "soc2.cc6.1" ;
    odrl:hasPolicy ?policy .
  ?policy comp:remediationPlan ?plan .
  ?plan comp:hasStep ?step ;
    comp:stepNumber ?stepNumber ;
    dcterms:title ?stepTitle .
}
ORDER BY ?stepNumber
```

**Quick Example:**
```bash
obsr query -q remediation_steps.rq --param ruleId=soc2.cc6.1
```

**Output:**
```csv
stepNumber,stepTitle,description,estimatedHours,responsible,dueDate,successCriteria
1,Assess MFA status,Audit all users,4.0,Security Team,2026-03-28,Report ready
2,Enable TOTP,Configure in IAM,6.0,Security Team,2026-04-02,All users enabled
3,User training,Distribute credentials,12.0,Support Team,2026-04-10,>95% adoption
```

---

## Documentation Files

### README.md
**Quick-start guide for executing and testing queries**

Includes:
- 3 ways to run queries (obsr CLI, HTTP, API)
- 5 test scenarios with expected outputs
- Integration examples (Go, Elixir)
- Troubleshooting Q&A
- Performance tuning tips

**Read:** 15 minutes
**For:** Operators, integration engineers

---

### MANIFEST.md
**Deliverables checklist and acceptance criteria**

Includes:
- File manifest (7 files, 1,106 lines total)
- Standards compliance (SPARQL 1.1, PROV-O, ODRL 2.0)
- Testing status (6 categories ✅)
- 80/20 coverage summary
- Production readiness assessment
- Next steps

**Read:** 10 minutes
**For:** Project leads, QA engineers

---

### ../docs/compliance-queries.md
**Complete SPARQL reference documentation**

Includes:
- Full documentation for each query
- Compliance framework mappings (SOC2, GDPR, HIPAA, SOX)
- Input parameters with examples
- Output columns with descriptions
- Example usage (CLI, HTTP, API)
- Example outputs (CSV, Turtle, JSON)
- Integration patterns (3 scenarios)
- Performance & execution
- Parameterization guide
- Troubleshooting guide
- References & standards

**Read:** 45 minutes
**For:** Developers, compliance auditors, architects

---

## File Sizes

| File | Lines | Size | Type |
|------|-------|------|------|
| compliance_violations.rq | 65 | 2.6K | SPARQL SELECT |
| compliance_audit_trail.rq | 151 | 5.2K | SPARQL CONSTRUCT |
| compliance_metrics.rq | 153 | 5.3K | SPARQL SELECT |
| remediation_steps.rq | 119 | 4.0K | SPARQL SELECT |
| compliance-queries.md | 618 | 20K | Documentation |
| README.md | 416 | 11K | Documentation |
| MANIFEST.md | 317 | 9.8K | Documentation |
| **TOTAL** | **1,839** | **58K** | |

---

## Standards Supported

### SPARQL 1.1 Features Used
- SELECT (violations, metrics, remediation)
- CONSTRUCT (audit trail)
- BIND for parameterization
- OPTIONAL for left outer join
- FILTER for conditions
- GROUP BY / COUNT aggregations
- ROUND for decimals
- ORDER BY / LIMIT for sorting
- VALUES for multi-value selection

### Semantic Web Standards
- **PROV-O** — W3C Provenance Ontology (activities, agents, entities)
- **ODRL 2.0** — Open Digital Rights Language (policies, permissions, remediation)
- **Dublin Core** — dcterms (issued, creator, title, subject)
- **FOAF** — Friend of a Friend (name, mbox, givenName)
- **RDF/RDFS/XSD** — W3C foundational namespaces

### Compliance Frameworks
- **SOC2** — Trust Service Criteria (CC, A, C, I)
- **GDPR** — General Data Protection Regulation (DS, CM, DPA, DM, DR)
- **HIPAA** — Health Insurance Portability (AC, AU, IA, AT)
- **SOX** — Sarbanes-Oxley (ITG, SA, AL, CM, FDI)

---

## Execution Methods

### Method 1: obsr CLI (Recommended)
```bash
cd BusinessOS/ontologies/sparql
obsr query -q compliance_violations.rq --param framework=SOC2
obsr query -q compliance_audit_trail.rq --format turtle > audit.ttl
```

### Method 2: HTTP (Oxigraph Endpoint)
```bash
curl -X POST http://localhost:7878/query \
  -H "Content-Type: application/sparql-query" \
  -d @compliance_violations.rq
```

### Method 3: BusinessOS API
```bash
curl -X POST http://localhost:8001/api/compliance/query \
  -H "Content-Type: application/json" \
  -d '{"query_name": "compliance_violations", "framework": "SOC2"}'
```

---

## Integration Examples

### Go (BusinessOS)
```go
// Fetch violations
violations := sparqlService.Query(
  "compliance_violations.rq",
  map[string]interface{}{"framework": "SOC2"},
)
```

### Elixir (Canopy)
```elixir
# Generate compliance report
metrics = SparqlService.query(
  "compliance_metrics.rq",
  framework: "GDPR",
  period: "month"
)
```

### JavaScript (Frontend)
```javascript
// Load metrics for dashboard
const metrics = await fetch('/api/compliance/metrics?framework=SOC2')
  .then(r => r.json())
```

---

## Common Use Cases

### Use Case 1: Violation Remediation
```
1. Run compliance_violations.rq → Find failures
2. For each violation:
3.   Run remediation_steps.rq → Get instructions
4.   Create Jira tickets
5.   Assign to teams
6.   Track completion
```

### Use Case 2: Monthly Compliance Report
```
1. Run compliance_metrics.rq → Get scores by control area
2. Run compliance_audit_trail.rq → Get audit trail
3. Generate PDF report
4. Send to board/auditors
```

### Use Case 3: Trend Analysis
```
1. Run compliance_metrics.rq for current month
2. Run compliance_metrics.rq for previous month
3. Compare compliance %
4. Plot trend (improving/stable/declining)
5. Update compliance dashboard
```

---

## Testing the Queries

### Test 1: Syntax Validation
```bash
obsr query -q compliance_violations.rq < /dev/null
# Expected: No errors
```

### Test 2: Sample Data Load
```bash
obsr insert --file test-data/sample-audit-entries.ttl
obsr query -q compliance_violations.rq --param framework=SOC2
# Expected: CSV with violations from test data
```

### Test 3: Parameter Injection
```bash
obsr query -q compliance_violations.rq \
  --param framework='SOC2"; DROP GRAPH <test> #'
# Expected: Safe (no code execution)
```

### Test 4: Timeout
```bash
time obsr query -q compliance_violations.rq --timeout 10000
# Expected: Completes within ~10 seconds
```

### Test 5: Output Formats
```bash
obsr query -q compliance_violations.rq --format csv > violations.csv
obsr query -q compliance_violations.rq --format json > violations.json
obsr query -q compliance_audit_trail.rq --format turtle > audit.ttl
```

---

## Troubleshooting

### Q: Empty results
**A:** Check that audit data exists in Oxigraph and framework names match.
```bash
SELECT (COUNT(*) AS ?count) WHERE { ?s ?p ?o }
SELECT DISTINCT ?framework WHERE { ?s bos:checkFramework ?framework }
```

### Q: Query timeout
**A:** Increase timeout or add indices on framework + date columns.
See compliance-queries.md "Index Requirements" section.

### Q: Syntax error
**A:** Validate with `obsr query -q file.rq < /dev/null` to see exact error.

---

## Production Checklist

Before deploying to production:

- [ ] Oxigraph triplestore deployed and running
- [ ] Compliance data loaded into RDF store
- [ ] Indices created (see compliance-queries.md)
- [ ] Sample queries executed and verified
- [ ] Integration code in BusinessOS/Canopy/OSA deployed
- [ ] Scheduled jobs for compliance reporting set up
- [ ] Dashboard UI displays metrics
- [ ] Alert rules configured for critical violations

---

## References

**Standards:**
- SPARQL 1.1: https://www.w3.org/TR/sparql11-query/
- PROV-O: https://www.w3.org/TR/prov-o/
- ODRL 2.0: https://www.w3.org/ns/odrl/2/

**Tools:**
- Oxigraph: https://github.com/oxigraph/oxigraph
- obsr CLI: https://github.com/oxigraph/oxigraph#cli

**Compliance Frameworks:**
- SOC2: https://www.aicpa.org/soc2
- GDPR: https://gdpr-info.eu/
- HIPAA: https://www.hhs.gov/hipaa/
- SOX: https://www.congress.gov/107/plaws/publ204/PLAW-107publ204.pdf

---

## Support

**Issues:** GitHub → https://github.com/seanchatmangpt/chatmangpt/issues
**Email:** compliance@chatmangpt.com
**Slack:** #compliance-audit-trail

---

**Last Updated:** 2026-03-26
**Agent:** Agent 37 (Semantic Web)
**Status:** ✅ Complete
