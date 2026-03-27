# FIBO Deal Management — Go Implementation Guide

**Status:** Complete (2026-03-26)
**Audience:** Go developers, API integrators, compliance engineers
**Purpose:** Implement financial deal lifecycle management with FIBO ontology integration in BusinessOS

---

## Overview

This guide covers the implementation of Agent 6: FIBO Deal Management in Go. The system provides:

1. **Service Layer** (`internal/services/fibo_deals.go`) — Core deal lifecycle operations
2. **HTTP Handlers** (`internal/handlers/fibo_deals.go`) — RESTful API endpoints
3. **Test Suite** (`*_test.go` files) — Comprehensive Chicago TDD tests

All deal data flows through **Oxigraph RDF triplestore** via SPARQL CONSTRUCT queries for semantic verification.

---

## Architecture

### Data Flow

```
HTTP Request
  ↓
Handler (validation, serialization)
  ↓
FIBODealsService (business logic)
  ↓
SPARQL CONSTRUCT query builder
  ↓
HTTP POST → Oxigraph (/query endpoint)
  ↓
RDF triples (response)
  ↓
Triple count + Deal response
```

### Key Design Decisions

1. **No ORM** — Direct SPARQL CONSTRUCT for semantic correctness
2. **5000ms Timeout** — All Oxigraph requests have explicit timeout + fallback
3. **Parameterized** — No SQL/SPARQL injection; queries built safely
4. **Idempotent IDs** — Auto-generated `d-{timestamp}` deal IDs
5. **Compliance-First** — KYC, AML, SOX checks built into service

---

## Service Layer API

### FIBODealsService

Located in: `internal/services/fibo_deals.go`

#### CreateDeal

```go
func (s *FIBODealsService) CreateDeal(ctx context.Context, deal *Deal) (*Deal, error)
```

**Purpose:** Create new financial deal and persist to RDF.

**Input:**
```go
deal := &Deal{
  Name:              "Cloud Infrastructure Deal",
  Amount:            250000.00,
  Currency:          "USD",
  BuyerID:           "acme-corp",
  SellerID:          "cloudtech-inc",
  ExpectedCloseDate: time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC),
  Probability:       85,
  Stage:             "negotiation",
}

result, err := service.CreateDeal(ctx, deal)
```

**Output:**
```json
{
  "id": "d-1740500123456",
  "name": "Cloud Infrastructure Deal",
  "amount": 250000.00,
  "currency": "USD",
  "status": "draft",
  "buyer_id": "acme-corp",
  "seller_id": "cloudtech-inc",
  "expected_close_date": "2026-06-30T00:00:00Z",
  "probability": 85,
  "stage": "negotiation",
  "created_at": "2026-03-26T14:30:00Z",
  "updated_at": "2026-03-26T14:30:00Z",
  "rdf_triple_count": 12,
  "compliance_status": "pending_verification"
}
```

**Validations:**
- `Name` required
- `Amount` > 0
- `BuyerID` required
- `SellerID` required
- `Probability` 0-100

**Error Handling:**
- Validation error: 400 (client error)
- Ontology timeout: "ontology query timeout (5000ms)"
- Ontology error: "ontology persistence failed: {detail}"

---

#### GetDeal

```go
func (s *FIBODealsService) GetDeal(ctx context.Context, dealID string) (*Deal, error)
```

**Purpose:** Retrieve single deal from RDF by ID.

**Input:**
```go
deal, err := service.GetDeal(ctx, "d-1740500123456")
```

**Output:** Deal struct with RDF metadata.

**Error Handling:**
- Empty dealID: "deal_id required"
- Not found: "deal retrieval failed: ..."

---

#### ListDeals

```go
func (s *FIBODealsService) ListDeals(ctx context.Context, limit int, offset int) ([]*Deal, error)
```

**Purpose:** Paginated list of all deals.

**Input:**
```go
deals, err := service.ListDeals(ctx, 50, 0)  // First 50 deals
```

**Output:** `[]*Deal` slice with pagination metadata.

**Pagination:**
- Default limit: 50
- Max limit: 1000
- Offset: 0-based

---

#### UpdateDeal

```go
func (s *FIBODealsService) UpdateDeal(ctx context.Context, dealID string, updates map[string]interface{}) (*Deal, error)
```

**Purpose:** Update deal fields in RDF.

**Input:**
```go
updates := map[string]interface{}{
  "dealStatus": "active",
  "dealProbability": 90,
}
deal, err := service.UpdateDeal(ctx, "d-1740500123456", updates)
```

**Output:** Updated deal struct.

**Error Handling:**
- Empty dealID: "deal_id required"
- No updates: "no updates provided"
- Deal not found: "deal not found: ..."

---

#### VerifyCompliance

```go
func (s *FIBODealsService) VerifyCompliance(ctx context.Context, dealID string) (map[string]interface{}, error)
```

**Purpose:** Check deal compliance across KYC, AML, SOX.

**Input:**
```go
result, err := service.VerifyCompliance(ctx, "d-1740500123456")
```

**Output:**
```json
{
  "kyc_verified": true,
  "kyc_triples": 5,
  "aml_screening": "passed",
  "aml_triples": 3,
  "sox_compliant": true,
  "sox_triples": 2,
  "compliance_status": "verified"
}
```

**Compliance Status Values:**
- `"verified"` — All checks passed (KYC ✓, AML ✓, SOX ✓)
- `"partial"` — KYC passed, AML/SOX pending
- `"failed"` — One or more checks failed

---

## HTTP API Endpoints

Located in: `internal/handlers/fibo_deals.go`

### POST /api/deals — Create Deal

**Request:**
```bash
curl -X POST http://localhost:8001/api/deals \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Cloud Infrastructure Deal",
    "amount": 250000.00,
    "currency": "USD",
    "buyer_id": "acme-corp",
    "seller_id": "cloudtech-inc",
    "expected_close_date": "2026-06-30T00:00:00Z",
    "probability": 85,
    "stage": "negotiation"
  }'
```

**Response (201 Created):**
```json
{
  "id": "d-1740500123456",
  "name": "Cloud Infrastructure Deal",
  "amount": 250000.00,
  "currency": "USD",
  "status": "draft",
  "buyer_id": "acme-corp",
  "seller_id": "cloudtech-inc",
  "expected_close_date": "2026-06-30T00:00:00Z",
  "probability": 85,
  "stage": "negotiation",
  "created_at": "2026-03-26T14:30:00Z",
  "updated_at": "2026-03-26T14:30:00Z",
  "rdf_triple_count": 12,
  "compliance_status": "pending_verification"
}
```

---

### GET /api/deals/:id — Get Deal

**Request:**
```bash
curl -X GET http://localhost:8001/api/deals/d-1740500123456
```

**Response (200 OK):**
```json
{
  "id": "d-1740500123456",
  "name": "Cloud Infrastructure Deal",
  "amount": 250000.00,
  "currency": "USD",
  "status": "draft",
  ...
}
```

**Error (404 Not Found):**
```json
{
  "error": "Deal not found: ..."
}
```

---

### GET /api/deals — List Deals

**Request:**
```bash
curl -X GET 'http://localhost:8001/api/deals?limit=50&offset=0'
```

**Response (200 OK):**
```json
{
  "deals": [
    {
      "id": "d-1",
      "name": "Deal 1",
      ...
    }
  ],
  "count": 25,
  "limit": 50,
  "offset": 0
}
```

**Pagination Query Parameters:**
- `limit` — Max results (default: 50, max: 1000)
- `offset` — Skip N results (default: 0)

---

### PATCH /api/deals/:id — Update Deal

**Request:**
```bash
curl -X PATCH http://localhost:8001/api/deals/d-1740500123456 \
  -H "Content-Type: application/json" \
  -d '{
    "status": "active",
    "probability": 90,
    "stage": "closing"
  }'
```

**Response (200 OK):**
```json
{
  "id": "d-1740500123456",
  "name": "Cloud Infrastructure Deal",
  "status": "active",
  "probability": 90,
  "stage": "closing",
  "updated_at": "2026-03-26T15:45:00Z",
  ...
}
```

---

### POST /api/deals/:id/verify-compliance — Verify Compliance

**Request:**
```bash
curl -X POST http://localhost:8001/api/deals/d-1740500123456/verify-compliance
```

**Response (200 OK):**
```json
{
  "deal_id": "d-1740500123456",
  "compliance": {
    "kyc_verified": true,
    "kyc_triples": 5,
    "aml_screening": "passed",
    "aml_triples": 3,
    "sox_compliant": true,
    "sox_triples": 2,
    "compliance_status": "verified"
  }
}
```

---

## Error Handling

### Error Response Format

All errors follow standard JSON:

```json
{
  "error": "Description of what went wrong"
}
```

### Common HTTP Status Codes

| Code | Meaning | Example |
|------|---------|---------|
| 201 | Created | Deal successfully created |
| 400 | Bad Request | Missing required field, invalid amount |
| 404 | Not Found | Deal ID does not exist |
| 500 | Internal Error | Oxigraph connection failed, timeout exceeded |

### Detailed Error Messages

#### Validation Errors (400)
```json
{
  "error": "deal name required"
}
```

#### Ontology Timeout (500)
```json
{
  "error": "Deal creation failed: ontology query timeout (5000ms)"
}
```

#### Ontology Server Error (500)
```json
{
  "error": "Deal creation failed: ontology persistence failed: ontology query failed with status 500: Server error"
}
```

---

## Configuration

### Environment Variables

Set these in `.env` before deployment:

```bash
# Oxigraph endpoint (required)
OXIGRAPH_URL=http://oxigraph:8890

# Deal service (optional)
FIBO_DEALS_TIMEOUT_MS=5000         # Default: 5000
FIBO_DEALS_HTTP_POOL_SIZE=10       # Connection pool size
```

### Initialization

Register routes in your Gin router:

```go
import "github.com/rhl/businessos-backend/internal/handlers"
import "github.com/rhl/businessos-backend/internal/services"

// Create service
dealsService := services.NewFIBODealsService(os.Getenv("OXIGRAPH_URL"))

// Create handler
dealsHandler := handlers.NewFIBODealsHandler(dealsService)

// Register routes
api := router.Group("/api")
auth := yourAuthMiddleware
handlers.RegisterFIBODealsRoutes(api, dealsHandler, auth)
```

---

## Performance & Latency

### Expected Latencies

| Operation | P50 | P95 | P99 |
|-----------|-----|-----|-----|
| CreateDeal | 150ms | 300ms | 500ms |
| GetDeal | 100ms | 250ms | 400ms |
| ListDeals (50) | 200ms | 450ms | 600ms |
| UpdateDeal | 150ms | 350ms | 550ms |
| VerifyCompliance | 400ms | 800ms | 1200ms |

**Timeout:** 5000ms (5 seconds) for all operations.

### Optimization Tips

1. **Connection Pooling** — FIBODealsService pools HTTP connections (default: 10)
2. **Pagination** — Always use `limit=50` for ListDeals unless you need more
3. **Bulk Operations** — Create deals one at a time; Oxigraph doesn't support batch CONSTRUCT
4. **KYC Verification** — Run `VerifyCompliance` async after deal creation to avoid blocking

---

## Testing

### Run Tests

```bash
cd BusinessOS/desktop/backend-go

# All tests
go test ./internal/services -run Fibo -v
go test ./internal/handlers -run FIBODeals -v

# Specific test
go test ./internal/services -run TestCreateDealValid -v

# With coverage
go test ./internal/services -run Fibo -cover
```

### Test Fixtures

Three example deals are tested:

1. **SaaS Deal** — $250K annual subscription
   - Buyer: acme-corp | Seller: saastech-inc | Probability: 90%

2. **Loan Deal** — $5M business expansion
   - Buyer: acme-corp | Seller: finance-bank | Probability: 70%

3. **Defense Contract** — $12.5M aerospace
   - Buyer: us-defense-dept | Seller: aerospace-corp | Probability: 65%

---

## Troubleshooting

### "ontology query timeout (5000ms)"

**Problem:** Oxigraph is slow or unreachable.

**Solutions:**
1. Check Oxigraph is running: `curl http://oxigraph:8890/health`
2. Increase timeout in code (not recommended for production)
3. Check network connectivity from Go container
4. Profile Oxigraph query performance: `weaver registry check`

### "Deal not found"

**Problem:** Deal ID doesn't exist in RDF.

**Solutions:**
1. Verify deal was created: `ListDeals` to see all deals
2. Check dealID format (should start with `d-`)
3. Ensure Oxigraph has the deal data (use `GET /api/ontology/data/deals`)

### "no updates provided"

**Problem:** PATCH request has empty updates object.

**Solutions:**
```bash
# Wrong
curl -X PATCH http://localhost:8001/api/deals/d-123 -d '{}'

# Correct
curl -X PATCH http://localhost:8001/api/deals/d-123 -d '{"status": "active"}'
```

### Compliance Verification Returns "failed"

**Problem:** KYC, AML, or SOX checks failing.

**Solutions:**
1. Check party KYC status: Verify buyer/seller have `hasKYCStatus: KYCVerified`
2. Check AML screening: Ensure parties have clean AML results
3. Check deal amount: SOX compliance checks deal < $1B

---

## Standards & Compliance

### Code Standards

- **Logging:** Use `slog` for all logs (never `fmt.Printf`)
- **HTTP Client:** All requests have explicit timeout
- **Error Messages:** Include operation, input, and root cause
- **No Secrets:** All sensitive config via environment variables

### Go Vet

```bash
cd BusinessOS/desktop/backend-go
go vet ./internal/services/fibo_deals.go
go vet ./internal/handlers/fibo_deals.go
# Expected: no warnings
```

### Test Coverage

```bash
go test ./internal/services -run Fibo -cover
# Expected: >80% coverage
```

---

## Integration Examples

### Create Deal + Verify Compliance

```bash
# 1. Create deal
DEAL=$(curl -s -X POST http://localhost:8001/api/deals \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Deal",
    "amount": 100000,
    "buyer_id": "buyer-1",
    "seller_id": "seller-1",
    "probability": 75
  }')

DEAL_ID=$(echo $DEAL | jq -r '.id')

# 2. Wait for async processing (optional)
sleep 1

# 3. Verify compliance
curl -X POST http://localhost:8001/api/deals/$DEAL_ID/verify-compliance
```

### Batch Create with Progress Tracking

```bash
for i in {1..10}; do
  curl -X POST http://localhost:8001/api/deals \
    -H "Content-Type: application/json" \
    -d "{
      \"name\": \"Deal $i\",
      \"amount\": $((100000 * i)),
      \"buyer_id\": \"buyer-1\",
      \"seller_id\": \"seller-$i\",
      \"probability\": 75
    }" &
done
wait
echo "Created 10 deals"
```

---

## References

- **FIBO Ontology:** https://spec.edmcouncil.org/fibo/
- **Oxigraph:** https://oxigraph.org/
- **SPARQL Standard:** https://www.w3.org/TR/sparql11-query/
- **Go HTTP Client:** https://pkg.go.dev/net/http
- **Gin Framework:** https://gin-gonic.com/

---

**Version:** 1.0.0 | **Updated:** 2026-03-26
