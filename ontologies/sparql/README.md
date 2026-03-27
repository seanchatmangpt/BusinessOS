# Compliance SPARQL Queries — Testing & Execution Guide

**Version:** 1.0.0
**Last Updated:** 2026-03-26

---

## Directory Structure

```
ontologies/
├── sparql/
│   ├── README.md                           ← You are here
│   ├── compliance_violations.rq            ← Query 1: Violation discovery
│   ├── compliance_audit_trail.rq           ← Query 2: PROV-O audit chain
│   ├── compliance_metrics.rq               ← Query 3: Compliance scoring
│   ├── remediation_steps.rq                ← Query 4: Remediation workflow
│   └── test-queries.sh                     ← Validation and test script
├── docs/
│   └── compliance-queries.md               ← Full reference documentation
└── fibo-deal-sparql-queries.rq             ← Existing FIBO queries (reference)
```

---

## Quick Start

### Option 1: Execute with obsr CLI (Recommended)

```bash
# List available queries
ls -lh BusinessOS/ontologies/sparql/*.rq

# Run a query
cd BusinessOS/ontologies/sparql

# Violation discovery
obsr query -q compliance_violations.rq --param framework=SOC2 --param lookbackDays=30

# Compliance metrics
obsr query -q compliance_metrics.rq \
  --param framework=GDPR \
  --param period=month \
  --param currentPeriod=2026-03-26 \
  --param previousPeriod=2026-02-26

# Remediation steps
obsr query -q remediation_steps.rq --param ruleId=soc2.cc6.1

# Audit trail (CONSTRUCT output)
obsr query -q compliance_audit_trail.rq \
  --param framework=SOC2 \
  --param startDate=2026-01-01 \
  --param endDate=2026-03-26 \
  --format turtle > audit-trail.ttl
```

### Option 2: Execute via HTTP (Oxigraph Endpoint)

```bash
# Assuming Oxigraph SPARQL endpoint at http://localhost:7878/query

curl -X POST http://localhost:7878/query \
  -H "Content-Type: application/sparql-query" \
  -d @compliance_violations.rq

# With parameters (POST form-encoded)
curl -X POST http://localhost:7878/query \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "query=$(cat compliance_violations.rq | jq -sRr @uri)&framework=GDPR&lookbackDays=90"
```

### Option 3: Execute via BusinessOS API

```bash
# Query endpoint (if implemented)
curl -X POST http://localhost:8001/api/compliance/query \
  -H "Content-Type: application/json" \
  -d '{
    "query_name": "compliance_violations",
    "framework": "SOC2",
    "lookbackDays": 30
  }'
```

---

## Query Status & Testing

| Query | File | Type | Status | Tests | Notes |
|-------|------|------|--------|-------|-------|
| **Violations** | `compliance_violations.rq` | SELECT | ✅ Ready | See below | 65 lines, parameterized |
| **Audit Trail** | `compliance_audit_trail.rq` | CONSTRUCT | ✅ Ready | See below | 151 lines, PROV-O output |
| **Metrics** | `compliance_metrics.rq` | SELECT | ✅ Ready | See below | 153 lines, trend analysis |
| **Remediation** | `remediation_steps.rq` | SELECT | ✅ Ready | See below | 119 lines, ODRL-based |

---

## Testing the Queries

### Test 1: Syntax Validation

All queries follow SPARQL 1.1 standard. Validate with:

```bash
# Via Oxigraph (returns error if syntax invalid)
obsr query -q compliance_violations.rq < /dev/null

# Via rapper (from Raptor RDF tools)
rapper --check compliance_violations.rq 2>&1 | grep -i error

# Expected: No output means OK
```

### Test 2: Integration Test (Requires Oxigraph + Data)

```bash
# Step 1: Load test data into Oxigraph
obsr insert --file test-data/sample-audit-entries.ttl

# Step 2: Run query
obsr query -q compliance_violations.rq --param framework=SOC2

# Step 3: Verify output
# Expected output: CSV with violation records
```

### Test 3: Parameter Injection Test

Verify parameters are properly escaped (no SPARQL injection):

```bash
# Test with special characters in parameter
obsr query -q compliance_violations.rq \
  --param framework='SOC2"; DROP GRAPH <test> #'

# Expected: Query fails gracefully (parameter not executed as SPARQL)
# NOT expected: Graph deletion occurs
```

### Test 4: Timeout Test

Verify query respects 10-second timeout:

```bash
# Run a query with deep recursion (intentionally slow)
time obsr query -q compliance_violations.rq --timeout 10000

# Expected: Exits cleanly after ~10 seconds
```

### Test 5: Output Format Test

```bash
# CSV output (default)
obsr query -q compliance_violations.rq --format csv > output.csv

# JSON output
obsr query -q compliance_violations.rq --format json > output.json

# Turtle output (for CONSTRUCT)
obsr query -q compliance_audit_trail.rq --format turtle > output.ttl

# N-Triples output (for CONSTRUCT)
obsr query -q compliance_audit_trail.rq --format ntriples > output.nt
```

---

## Expected Output Examples

### compliance_violations.rq Output

```csv
violationId,ruleId,severity,controlName,description,detectedAt,affectedEntity,remediationUrl
vio-entry-001-2026-03-26,soc2.cc6.1,critical,Logical access restricted,User alice accessed unauth resource,2026-03-26T14:30:00Z,user-alice,https://chatmangpt.com/api/compliance/remediation/soc2.cc6.1
vio-entry-002-2026-03-26,soc2.c1.1,high,Encryption at rest,Data in cache not encrypted,2026-03-25T09:15:00Z,cache-redis,https://chatmangpt.com/api/compliance/remediation/soc2.c1.1
```

### compliance_audit_trail.rq Output (Turtle)

```turtle
@prefix prov: <http://www.w3.org/ns/prov#> .
@prefix audit: <https://chatmangpt.com/ontology/audit/> .
@prefix dcterms: <http://purl.org/dc/terms/> .

<https://chatmangpt.com/audit/2026-03-26/activity-001> a prov:Activity ;
  audit:activityType "control_check" ;
  audit:framework "SOC2" ;
  prov:startedAtTime "2026-03-26T10:00:00Z"^^xsd:dateTime ;
  audit:result <https://chatmangpt.com/audit/2026-03-26/activity-001#result> ;
  prov:wasAssociatedWith <https://chatmangpt.com/agents/compliance-bot> ;
  audit:affectedResource <https://chatmangpt.com/resources/user-alice> .

<https://chatmangpt.com/audit/2026-03-26/activity-001#result> a audit:Result ;
  audit:resultStatus "FAILED" ;
  audit:resultMessage "User lacks MFA" ;
  audit:confidence 0.95 ;
  prov:wasGeneratedBy <https://chatmangpt.com/audit/2026-03-26/activity-001> .
```

### compliance_metrics.rq Output

```csv
framework,period,controlArea,totalControls,passedControls,failedControls,skippedControls,compliancePercent,previousPercent,trend,riskScore,actionItems
SOC2,month,Access Control,15,14,1,0,93.33,90.00,improving,6.67,2
SOC2,month,Encryption,10,10,0,0,100.00,100.00,stable,0.00,0
GDPR,month,Consent Management,5,5,0,0,100.00,100.00,stable,0.00,0
```

### remediation_steps.rq Output

```csv
stepNumber,stepTitle,stepDescription,estimatedHours,responsible,dueDate,relatedResources,prerequisiteSteps,successCriteria
1,Assess current state,Audit all user accounts for MFA status,4.0,Security Team,2026-03-28,https://wiki/mfa-audit,,Report: X users without MFA
2,Enable MFA,Log into IAM and enable TOTP,6.0,Security Team,2026-04-02,https://wiki/mfa-setup,1,All users have TOTP codes
3,User rollout,Distribute credentials and train users,12.0,Support Team,2026-04-10,https://wiki/training,2,>95% adoption rate
```

---

## Integration with BusinessOS

### A. Add to Compliance Dashboard Backend

```go
// BusinessOS: internal/handlers/compliance.go

func (h *ComplianceHandler) GetViolations(c *gin.Context) {
  framework := c.Query("framework")  // e.g., "SOC2"
  lookbackDays := c.DefaultQuery("lookbackDays", "30")

  // Execute SPARQL query via obsr
  results := h.sparqlService.Query(
    "compliance_violations.rq",
    map[string]interface{}{
      "framework": framework,
      "lookbackDays": lookbackDays,
    },
  )

  c.JSON(200, results)
}
```

### B. Add to Compliance Reporting Job

```elixir
# Canopy: lib/canopy/jobs/compliance_reporter.ex

def run_monthly_compliance_report(framework) do
  # 1. Get metrics
  metrics = SparqlService.query(
    "compliance_metrics.rq",
    framework: framework,
    period: "month",
    currentPeriod: Date.today(),
    previousPeriod: Date.add(Date.today(), -30)
  )

  # 2. Get audit trail
  audit_trail = SparqlService.construct(
    "compliance_audit_trail.rq",
    framework: framework,
    startDate: Date.add(Date.today(), -30),
    endDate: Date.today()
  )

  # 3. Generate PDF report
  ComplianceReportGenerator.generate(metrics, audit_trail)
end
```

### C. Add to Remediation Workflow

```go
// BusinessOS: internal/workflows/remediation.go

func CreateRemediationWorkflow(violationId string, ruleId string) {
  // 1. Get remediation steps
  steps := sparqlService.Query(
    "remediation_steps.rq",
    map[string]interface{}{
      "violationId": violationId,
      "ruleId": ruleId,
    },
  )

  // 2. Create Jira tickets or issue tasks
  for _, step := range steps {
    CreateTask(Task{
      Title: step.StepTitle,
      Description: step.StepDescription,
      AssignedTo: step.Responsible,
      DueDate: step.DueDate,
      Effort: step.EstimatedHours,
      Links: step.RelatedResources,
    })
  }
}
```

---

## Troubleshooting

### Issue: "No results returned"

**Causes:**
1. Oxigraph has no data loaded
2. Date range doesn't match audit entries
3. Framework name doesn't exist in data

**Solution:**
```bash
# Check if data exists
obsr query -q - <<EOF
SELECT (COUNT(*) AS ?count) WHERE { ?s ?p ?o }
EOF

# Should return > 0

# List all frameworks in data
obsr query -q - <<EOF
SELECT DISTINCT ?framework WHERE {
  ?s bos:checkFramework ?framework .
}
EOF
```

### Issue: "Query timeout (>10 seconds)"

**Causes:**
1. Large dataset (>1M triples)
2. Complex recursive patterns
3. Missing indices

**Solution:**
```bash
# Increase timeout
obsr query -q compliance_violations.rq --timeout 30000

# Or add indices (if using triple database)
# See compliance-queries.md "Index Requirements" section
```

### Issue: "SPARQL syntax error"

**Causes:**
1. Malformed query
2. Invalid namespace prefix
3. Type mismatch in FILTER

**Solution:**
```bash
# Validate syntax with Oxigraph
obsr query -q compliance_violations.rq < /dev/null

# Check error message for specific line/column
# Common: Missing ; at end of pattern, undefined PREFIX
```

---

## Performance Tuning

### For 1M+ triples

```sparql
# Add LIMIT to SELECT queries
SELECT ... LIMIT 10000

# Use FILTER to reduce matches early
FILTER (?framework = "SOC2")  # Before JOIN

# For CONSTRUCT, materialize incrementally
# Instead of: CONSTRUCT { many patterns }
# Use: Multiple smaller CONSTRUCT calls
```

### Memory Usage

```bash
# Check Oxigraph memory
ps aux | grep oxigraph | grep -o Res=\S+

# If > 4GB, increase JVM heap:
# export JAVA_OPTS="-Xmx8g"
# Then restart Oxigraph
```

---

## References

- **Full Documentation:** `ontologies/docs/compliance-queries.md`
- **SPARQL Spec:** https://www.w3.org/TR/sparql11-query/
- **Oxigraph CLI:** https://github.com/oxigraph/oxigraph#cli
- **PROV-O:** https://www.w3.org/TR/prov-o/
- **Compliance Rules:** `BusinessOS/config/compliance-rules.yaml`

---

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2026-03-26 | Initial release with 4 queries |

---

## License

These SPARQL queries are part of ChatmanGPT and licensed under the same terms as the main project.

For questions or contributions, see: https://github.com/seanchatmangpt/chatmangpt
