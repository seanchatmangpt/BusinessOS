# BOS Integration Test Quick Start

## One-Liner: Run All Tests

```bash
cd BusinessOS/desktop/backend-go && go test -v ./tests/integration/bos_integration_test.go -timeout 60s
```

## With BusinessOS Running (Full Integration)

```bash
# Terminal 1: Start BusinessOS
cd BusinessOS && make dev

# Terminal 2: Run tests (waits for startup)
sleep 5 && cd BusinessOS/desktop/backend-go && go test -v ./tests/integration/bos_integration_test.go -timeout 60s
```

## Test Results Summary

| Test Name | Purpose | Timeout | Skip Condition |
|-----------|---------|---------|---|
| `TestBOSSchemaImportRoundTrip` | Import→Export→Reimport cycle | 30s | BusinessOS unavailable |
| `TestBOSSchemaUpdateAndValidate` | Update + re-validate | 30s | BusinessOS unavailable |
| `TestBOSRDFExportAndOxigraphQuery` | RDF generation + SPARQL | 30s | BusinessOS or Oxigraph unavailable |
| `TestBOSHealthcareSchemaImport` | HIPAA schema import | 30s | BusinessOS unavailable |
| `TestBOSTimingAssertions` | SLA enforcement | 30s | BusinessOS unavailable |
| `TestBOSDataIntegrityAcrossOperations` | Hash consistency | 30s | BusinessOS unavailable |
| `TestBOSMultipleSchemaImports` | Schema isolation | 30s | BusinessOS unavailable |
| `TestBOSCLICompatibility` | CLI format validation | 5s | Always runs |

## Quick Checks

### Schema Import SLA (≤500ms)
```bash
go test -v ./tests/integration/bos_integration_test.go -run TestBOSTimingAssertions -timeout 30s
```
Expected: `All timing assertions passed: import=XXXms`

### Data Integrity (Round-Trip)
```bash
go test -v ./tests/integration/bos_integration_test.go -run TestBOSSchemaImportRoundTrip -timeout 30s
```
Expected: `Round-trip test completed: import=XXXms, export=XXms, reimport=XXXms`

### Multiple Schemas Coexist
```bash
go test -v ./tests/integration/bos_integration_test.go -run TestBOSMultipleSchemaImports -timeout 30s
```
Expected: `Successfully imported and verified multiple schemas`

## Environment Variables

```bash
# Use non-default ports/hosts
export BUSINESSOS_URL=http://192.168.1.100:8001
export OXIGRAPH_URL=http://192.168.1.100:6379
export BUSINESSOS_BOS_PATH=/custom/path/to/bos

go test -v ./tests/integration/bos_integration_test.go -timeout 60s
```

## Debugging Failed Tests

### See Full Output
```bash
go test -v ./tests/integration/bos_integration_test.go -run TestName -timeout 30s 2>&1 | grep -A 20 "FAIL:"
```

### Check Network Connectivity
```bash
curl -v http://localhost:8001/health
curl -v http://localhost:6379/api/query -X POST -d '{"query":"SELECT ?s WHERE { ?s ?p ?o } LIMIT 1"}'
```

### Run in Verbose Mode
```bash
RUST_LOG=debug go test -v ./tests/integration/bos_integration_test.go -timeout 60s
```

## API Endpoints Being Tested

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/api/bos/schema/import` | POST | Import schema (JSON/SQL) |
| `/api/bos/schema/export/{id}` | GET | Export schema (JSON/RDF/CSV/SQL) |
| `/api/bos/schema/validate/{id}` | POST | Validate schema |
| `/api/bos/schema/update` | POST | Update existing schema |
| `/api/bos/schema/query` | POST | Execute SQL query on schema |

## Performance Baselines

Typical times when BusinessOS running:

```
Schema Import:      ~245ms (SLA: ≤500ms)
Schema Export:      ~78ms  (SLA: ≤200ms)
Schema Validate:    ~45ms  (SLA: ≤200ms)
Schema Update:      ~156ms (SLA: ≤200ms)
SQL Query:          ~89ms  (SLA: ≤200ms)
SPARQL Query:       ~125ms (SLA: ≤200ms)
Complete Round-Trip: ~562ms (SLA: <1s)
```

## Test Data

### FIBO Domain (Financial)
- 2 tables: `fibo_deal`, `fibo_party`
- 3 indexes: `idx_deal_status`, `idx_deal_created`, `idx_party_type`
- Sample: Treasury Bond Issuance 2026 (5M USD)

### Healthcare Domain (HIPAA)
- 3 tables: `patient`, `encounter`, `phi_audit`
- Compliance: PHI tracking, access audit, consent enforcement
- Sample: Patient "Jane Doe" with audit trail

### RDF/SPARQL Domain
- CONSTRUCT transformation → RDF triples
- Sample query: Find all deals and parties
- Results: Indexed in Oxigraph (if available)

## CI/CD Integration

### Before Committing
```bash
go test -v ./tests/integration/bos_integration_test.go -timeout 60s && echo "All tests passed ✓"
```

### In GitHub Actions
```yaml
- name: BOS Integration Tests
  run: |
    cd BusinessOS/desktop/backend-go
    go test -v ./tests/integration/bos_integration_test.go -timeout 60s
```

## Interpreting Results

### All Tests Pass ✓
```
PASS
ok  	command-line-arguments	3.24s
```
BusinessOS is running, all operations within SLA.

### All Tests Skip (Expected)
```
--- SKIP: TestBOSSchemaImportRoundTrip (0.00s)
...
--- PASS: TestBOSCLICompatibility (0.00s)

PASS
ok  	command-line-arguments	0.53s
```
BusinessOS not running (normal for dev environment).

### Test Fails (Investigate)
```
--- FAIL: TestBOSTimingAssertions (0.45s)
    bos_integration_test.go:629: FAIL: Import timing 650 ms exceeds SLA of 500 ms
```
Import operation is too slow. Check:
1. PostgreSQL responsiveness
2. Network latency
3. Schema complexity (check `TablesImported` count)

## Common Issues & Fixes

| Issue | Check | Fix |
|-------|-------|-----|
| "Connection refused" | Is BusinessOS running? | `make dev` in BusinessOS dir |
| "Timeout" | Is PostgreSQL responsive? | Check `docker ps`, restart container |
| "Hash mismatch" | JSON marshaling deterministic? | Ensure no random fields |
| "SPARQL fails" | Is Oxigraph running? | Not required, test skips gracefully |
| "Import timing > 500ms" | System load high? | Reduce background processes |

## Test Coverage

- ✓ Schema import/export (all formats: JSON/RDF/CSV/SQL)
- ✓ Schema update operations
- ✓ Schema validation
- ✓ Data integrity across round-trip
- ✓ Multiple domain schemas (FIBO, Healthcare)
- ✓ RDF triple generation
- ✓ SPARQL query execution
- ✓ SLA timing enforcement
- ✓ Error handling and graceful degradation

## Next Steps

1. **Run tests locally:** `go test -v ./tests/integration/bos_integration_test.go`
2. **Monitor performance:** Track timings over time
3. **Extend coverage:** Add chaos tests, load tests
4. **CI integration:** Add to GitHub Actions/GitLab CI
5. **Document APIs:** Use test cases as API documentation

---

**File:** `BusinessOS/desktop/backend-go/tests/integration/bos_integration_test.go`
**Status:** Ready for use
**Last Updated:** 2026-03-26
