# How to Set Up FIBO Deal Integration

**Step-by-step guide to integrate FIBO ontology into BusinessOS deal management**

---

## Prerequisites

- Oxigraph triplestore running (or compatible RDF database)
- BusinessOS PostgreSQL database configured
- SPARQL client (curl, Postman, or Oxigraph UI)
- Basic SPARQL/RDF knowledge

---

## Step 1: Load the Ontology

### 1.1 Using Oxigraph HTTP API

```bash
# Start Oxigraph (if not running)
docker run -d -p 8890:8890 oxigraph/oxigraph

# Load the FIBO ontology into default graph
curl -X POST http://localhost:8890/store \
  -H "Content-Type: text/turtle" \
  --data-binary @/path/to/fibo-deal-integration.ttl

# Verify load (query for ontology metadata)
curl -X POST http://localhost:8890/query \
  -H "Content-Type: application/sparql-query" \
  -d "SELECT ?ontology WHERE { ?ontology a owl:Ontology . }"
```

### 1.2 Using Named Graph

```bash
# Load into named graph (recommended for separation)
curl -X POST http://localhost:8890/store \
  -H "Content-Type: text/turtle" \
  -H "Graph-URI: https://ontology.businessos.dev/fibo-deals" \
  --data-binary @fibo-deal-integration.ttl
```

### 1.3 Verify Ontology Loaded

```sparql
PREFIX : <https://ontology.businessos.dev/fibo-deals/>
PREFIX owl: <http://www.w3.org/2002/07/owl#>

SELECT ?label ?version
WHERE {
  <https://ontology.businessos.dev/fibo-deals> a owl:Ontology ;
    rdfs:label ?label ;
    owl:versionInfo ?version .
}
```

**Expected result:**
```
label: "BusinessOS FIBO Deal Integration Ontology"
version: "1.0.0"
```

---

## Step 2: Insert Example Data

### 2.1 Load Example Deals (Turtle Format)

```bash
curl -X POST http://localhost:8890/store \
  -H "Content-Type: text/turtle" \
  --data-binary @/path/to/fibo-deal-examples.ttl
```

### 2.2 Verify Data Inserted

```sparql
SELECT (COUNT(?deal) AS ?deal_count)
WHERE {
  ?deal a <https://ontology.businessos.dev/fibo-deals/Deal> .
}
```

**Expected result:** `deal_count: 3` (from example file)

---

## Step 3: Set Up PostgreSQL to RDF Pipeline

### 3.1 Create R2RML Mapping File

Create file: `database/mappings/deals-to-rdf.ttl`

```turtle
@prefix rr: <http://www.w3.org/ns/r2rml#> .
@prefix : <https://ontology.businessos.dev/fibo-deals/> .
@prefix xsd: <http://www.w3.org/2001/XMLSchema#> .

<#DealsMapping>
  rr:logicalTable [ rr:tableName "deals" ] ;
  rr:subjectMap [
    rr:template "https://businessos.dev/id/deals/{id}" ;
    rr:class :Deal
  ] ;
  rr:predicateObjectMap [
    rr:predicate :dealIdentifier ;
    rr:objectMap [ rr:column "id" ]
  ] ;
  rr:predicateObjectMap [
    rr:predicate :dealName ;
    rr:objectMap [ rr:column "name" ]
  ] ;
  rr:predicateObjectMap [
    rr:predicate :dealAmount ;
    rr:objectMap [
      rr:parentTriplesMap <#AmountsMapping> ;
      rr:joinCondition [ rr:child "id" ; rr:parent "deal_id" ]
    ]
  ] .

<#AmountsMapping>
  rr:logicalTable [ rr:tableName "deal_amounts" ] ;
  rr:subjectMap [
    rr:template "https://businessos.dev/id/amounts/{deal_id}-principal" ;
    rr:class :MonetaryAmount
  ] ;
  rr:predicateObjectMap [
    rr:predicate :amount ;
    rr:objectMap [ rr:column "amount" ; rr:datatype xsd:decimal ]
  ] ;
  rr:predicateObjectMap [
    rr:predicate :currency ;
    rr:objectMap [ rr:column "currency_code" ]
  ] .
```

### 3.2 Execute R2RML Mapping

Using RMLMapper (Java):

```bash
java -jar rmlmapper.jar \
  -m database/mappings/deals-to-rdf.ttl \
  -o deals-rdf.ttl \
  -s turtle
```

### 3.3 Load Mapped RDF

```bash
curl -X POST http://localhost:8890/store \
  -H "Content-Type: text/turtle" \
  --data-binary @deals-rdf.ttl
```

---

## Step 4: Execute SPARQL CONSTRUCT Queries

### 4.1 Query 1: Deal Creation

Create script: `scripts/construct-deals.sparql`

```sparql
# See fibo-deal-sparql-queries.rq line 29-113
PREFIX : <https://ontology.businessos.dev/fibo-deals/>
PREFIX fibo-fnd: <https://spec.edmcouncil.org/fibo/ontology/FND/>
PREFIX fibo-fbc: <https://spec.edmcouncil.org/fibo/ontology/FBC/>

CONSTRUCT {
  ?deal_uri a :Deal ;
    a fibo-fnd:Agreement ;
    :dealIdentifier ?deal_id ;
    :dealName ?deal_name ;
    :dealAmount ?amount_uri ;
    :dealCurrency ?currency .

  ?amount_uri a :MonetaryAmount ;
    :amount ?amount_value ;
    :currency ?currency .
}
WHERE {
  # ... (see full query in fibo-deal-sparql-queries.rq)
}
```

Execute:

```bash
curl -X POST http://localhost:8890/query \
  -H "Content-Type: application/sparql-query" \
  -d @scripts/construct-deals.sparql \
  --header "Accept: application/n-triples" \
  > result-deals.nt

# Insert results back
curl -X POST http://localhost:8890/store \
  -H "Content-Type: application/n-triples" \
  --data-binary @result-deals.nt
```

### 4.2 Query 2: Compliance Checking

```bash
curl -X POST http://localhost:8890/query \
  -H "Content-Type: application/sparql-query" \
  -d @scripts/construct-compliance.sparql \
  --header "Accept: text/turtle" \
  > result-compliance.ttl
```

### 4.3 Query 3: KYC Verification

```bash
curl -X POST http://localhost:8890/query \
  -H "Content-Type: application/sparql-query" \
  -d @scripts/construct-kyc.sparql \
  --header "Accept: text/turtle" \
  > result-kyc.ttl
```

---

## Step 5: Create REST API Endpoints

### 5.1 Get Deal in FIBO Format

**Endpoint:** `GET /api/deals/{deal_id}/fibo`

```go
// BusinessOS/desktop/backend-go/internal/handlers/deals_fibo.go

package handlers

import (
  "github.com/gin-gonic/gin"
  "net/http"
)

func (h *DealsHandler) GetDealFIBO(c *gin.Context) {
  dealID := c.Param("id")

  // Query triplestore for deal RDF
  query := fmt.Sprintf(`
    PREFIX : <https://ontology.businessos.dev/fibo-deals/>
    PREFIX fibo-fnd: <https://spec.edmcouncil.org/fibo/ontology/FND/>

    CONSTRUCT {
      ?deal ?p ?o .
    }
    WHERE {
      ?deal a :Deal ;
        :dealIdentifier "%s" ;
        ?p ?o .
    }
  `, dealID)

  // Execute SPARQL CONSTRUCT
  rdf, err := h.triplestore.Construct(c.Request.Context(), query)
  if err != nil {
    utils.RespondInternalError(c, slog.Default(), "query deal FIBO", err)
    return
  }

  // Return RDF in requested format
  contentType := c.Query("format")
  if contentType == "" {
    contentType = "text/turtle"
  }

  c.Data(http.StatusOK, contentType, rdf)
}
```

Register route:

```go
func (h *DealsHandler) RegisterRoutes(router *gin.Engine) {
  deals := router.Group("/api/deals")
  deals.GET("/:id/fibo", h.GetDealFIBO)
  deals.POST("/validate-compliance", h.ValidateCompliance)
  deals.POST("/parties/verify-kyc", h.VerifyKYC)
}
```

### 5.2 Validate Deal Compliance

**Endpoint:** `POST /api/deals/validate-compliance`

```go
func (h *DealsHandler) ValidateCompliance(c *gin.Context) {
  var req struct {
    DealURI string `json:"deal_uri" binding:"required"`
  }

  if err := c.ShouldBindJSON(&req); err != nil {
    utils.RespondInvalidRequest(c, slog.Default(), err)
    return
  }

  // Execute Query 2 (Compliance Checking)
  complianceReport, err := h.triplestore.Construct(c.Request.Context(),
    complianceCheckingQuery(req.DealURI))

  if err != nil {
    utils.RespondInternalError(c, slog.Default(), "compliance check", err)
    return
  }

  // Parse and return compliance status
  c.JSON(http.StatusOK, gin.H{
    "deal_uri": req.DealURI,
    "report": complianceReport,
  })
}
```

### 5.3 Verify Party KYC

**Endpoint:** `POST /api/parties/verify-kyc`

```go
func (h *DealsHandler) VerifyKYC(c *gin.Context) {
  var req struct {
    PartyID string `json:"party_id" binding:"required"`
  }

  if err := c.ShouldBindJSON(&req); err != nil {
    utils.RespondInvalidRequest(c, slog.Default(), err)
    return
  }

  // Execute Query 3 (KYC Verification)
  kycReport, err := h.triplestore.Construct(c.Request.Context(),
    kycVerificationQuery(req.PartyID))

  if err != nil {
    utils.RespondInternalError(c, slog.Default(), "KYC verification", err)
    return
  }

  // Parse KYC decision from RDF
  decision := parseKYCDecision(kycReport)

  c.JSON(http.StatusOK, gin.H{
    "party_id": req.PartyID,
    "decision": decision,
    "report": kycReport,
  })
}
```

---

## Step 6: Set Up Automated Workflows

### 6.1 Deal Lifecycle Hook

**When deal status changes:**

```go
// In deals service
func (s *DealsService) UpdateDealStatus(ctx context.Context, dealID, newStatus string) error {
  // 1. Update SQL database
  err := s.repo.UpdateDealStatus(ctx, dealID, newStatus)
  if err != nil {
    return err
  }

  // 2. Update RDF triplestore
  updateQuery := fmt.Sprintf(`
    PREFIX : <https://ontology.businessos.dev/fibo-deals/>

    DELETE { ?deal :dealStatus ?oldStatus . }
    INSERT { ?deal :dealStatus :%s . }
    WHERE {
      ?deal :dealIdentifier "%s" ;
        :dealStatus ?oldStatus .
    }
  `, newStatus, dealID)

  err = s.triplestore.Update(ctx, updateQuery)
  if err != nil {
    slog.Error("failed to update deal RDF", "error", err)
    // Log but don't fail - SQL is source of truth
  }

  // 3. Create audit entry
  audit := &DealAuditEntry{
    DealID:    dealID,
    Action:    "status_change",
    Field:     "dealStatus",
    NewValue:  newStatus,
    Timestamp: time.Now(),
    UserID:    ctx.Value("user_id").(string),
  }

  return s.repo.CreateAuditEntry(ctx, audit)
}
```

### 6.2 Compliance Check on Deal Creation

```go
func (s *DealsService) CreateDeal(ctx context.Context, deal *Deal) error {
  // 1. Create deal in SQL
  createdDeal, err := s.repo.CreateDeal(ctx, deal)
  if err != nil {
    return err
  }

  // 2. Insert into RDF (asynchronous)
  go func() {
    s.insertDealRDF(context.Background(), createdDeal)
  }()

  // 3. Run compliance check (asynchronous)
  go func() {
    ctx := context.Background()
    report, err := s.checkCompliance(ctx, createdDeal.ID)
    if err != nil {
      slog.Error("compliance check failed", "deal_id", createdDeal.ID, "error", err)
      return
    }

    // If non-compliant, flag for review
    if report.Status == "NON_COMPLIANT" {
      s.FlagDealForReview(ctx, createdDeal.ID, report.Issues...)
    }
  }()

  return nil
}
```

---

## Step 7: Testing

### 7.1 Unit Tests for SPARQL

Create file: `internal/ontology/fibo_test.go`

```go
package ontology

import (
  "testing"
  "github.com/stretchr/testify/assert"
)

func TestConstructDeal(t *testing.T) {
  client := setupTestTriplestore(t)
  defer client.Close()

  // Insert test deal
  testDeal := `
    <https://businessos.dev/id/deals/test-001>
      a <https://ontology.businessos.dev/fibo-deals/Deal> ;
      <https://ontology.businessos.dev/fibo-deals/dealIdentifier> "test-001" ;
      <https://ontology.businessos.dev/fibo-deals/dealName> "Test Deal" .
  `

  err := client.Insert(testDeal)
  assert.NoError(t, err)

  // Execute CONSTRUCT query
  result, err := client.Construct(ConstructDealQuery)
  assert.NoError(t, err)
  assert.NotEmpty(t, result)

  // Verify result contains FIBO types
  assert.Contains(t, result, "fibo-fnd:Agreement")
}

func TestComplianceCheck(t *testing.T) {
  client := setupTestTriplestore(t)
  defer client.Close()

  // Insert test data (deal + contract + parties)
  insertTestData(t, client)

  // Run compliance check
  report, err := client.Construct(ComplianceCheckingQuery)
  assert.NoError(t, err)

  // Verify report contains compliance validation
  assert.Contains(t, report, "ComplianceValidationReport")
  assert.Contains(t, report, "COMPLIANT") // Or NON_COMPLIANT
}

func TestKYCVerification(t *testing.T) {
  client := setupTestTriplestore(t)
  defer client.Close()

  // Insert test party
  testParty := `
    <https://businessos.dev/id/parties/test-party>
      a <https://ontology.businessos.dev/fibo-deals/BusinessOSParty> ;
      <https://ontology.businessos.dev/fibo-deals/legalName> "Test Corp" ;
      <https://ontology.businessos.dev/fibo-deals/businessRegistrationNumber> "12-3456789" ;
      <https://ontology.businessos.dev/fibo-deals/operatingJurisdiction> "US" .
  `

  err := client.Insert(testParty)
  assert.NoError(t, err)

  // Run KYC verification
  report, err := client.Construct(KYCVerificationQuery)
  assert.NoError(t, err)

  // Verify decision
  assert.Contains(t, report, "KYCVerified") // Or KYCPending/KYCRejected
}
```

Run tests:

```bash
cd BusinessOS/desktop/backend-go
go test ./internal/ontology/... -v
```

### 7.2 Integration Tests

```bash
# Test end-to-end: SQL → RDF → SPARQL → API
make test-integration TEST=fibo_deal_integration
```

---

## Step 8: Monitoring & Maintenance

### 8.1 Health Check

```bash
# Verify ontology is loaded and queryable
curl -X POST http://localhost:8890/query \
  -H "Content-Type: application/sparql-query" \
  -d "SELECT (COUNT(*) AS ?count) WHERE { ?s a ?o . }"
```

### 8.2 Backup Ontology

```bash
# Export RDF for backup
curl -X POST http://localhost:8890/query \
  -H "Content-Type: application/sparql-query" \
  -d "CONSTRUCT { ?s ?p ?o } WHERE { ?s ?p ?o }" \
  --header "Accept: text/turtle" \
  > backup-ontology-$(date +%Y%m%d).ttl
```

### 8.3 Monitor RDF Sizes

```sparql
# Check graph statistics
SELECT ?graph (COUNT(*) AS ?triple_count)
WHERE {
  GRAPH ?graph { ?s ?p ?o . }
}
GROUP BY ?graph
ORDER BY DESC(?triple_count)
```

---

## Troubleshooting

### Ontology Not Loading

```bash
# Validate Turtle syntax
rapper -i turtle -c fibo-deal-integration.ttl
```

### SPARQL Query Timeout

```bash
# Reduce result set with LIMIT
SELECT * WHERE { ... } LIMIT 100

# Add FILTER to narrow search
FILTER (?dealStatus = :DealNegotiation)
```

### RDF Not Visible After Insert

```bash
# Verify insert worked
curl -X POST http://localhost:8890/query \
  -H "Content-Type: application/sparql-query" \
  -d "SELECT * WHERE { ?s ?p ?o } LIMIT 10"

# Check if data is in expected graph
curl -X POST http://localhost:8890/query \
  -H "Content-Type: application/sparql-query" \
  -d "SELECT ?g WHERE { GRAPH ?g { ?s ?p ?o } } LIMIT 1"
```

---

## Next Steps

1. **Deploy to staging:** Run full integration test suite
2. **Load production data:** R2RML mapping of all historical deals
3. **Set up dashboards:** SPARQL queries for compliance/KYC metrics
4. **Automate workflows:** Compliance checks on deal creation
5. **Enable knowledge graph:** Allow cross-deal analytics queries

---

**Version:** 1.0.0
**Last Updated:** 2026-03-25
