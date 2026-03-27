# FIBO REST API Documentation

**Version:** 1.0.0
**Base URL:** `http://localhost:8001/api`
**Status:** Production Ready
**Last Updated:** 2026-03-26

## Overview

The FIBO REST API provides endpoints for managing financial deals integrated with the Financial Industry Business Ontology (FIBO). All deal operations are persisted to a SPARQL-compatible ontology via the Oxigraph RDF triplestore.

**Key Features:**
- Full CRUD operations for financial deals
- FIBO ontology compliance (FND, FBC ontologies)
- KYC/AML/SOX compliance verification
- Automatic RDF triple generation via SPARQL CONSTRUCT
- Request/response validation with clear error messages
- Pagination support for list operations
- Metadata tracking (created_at, updated_at, RDF triple counts)

---

## Architecture

```
HTTP Request (Handler)
  ↓
Validation & Binding
  ↓
Service Layer (FIBODealsService)
  ↓
SPARQL CONSTRUCT Query Builder
  ↓
Oxigraph (HTTP POST /query)
  ↓
RDF Triplestore
  ↓
Response (N-Triples count + Deal object)
```

**Service Dependencies:**
- `services.FIBODealsService` — manages deal lifecycle with FIBO integration
- Oxigraph — RDF triplestore at `http://localhost:8890`
- PostgreSQL — stores deal metadata (optional, for indexing)

---

## OpenAPI 3.0 Specification

```yaml
openapi: 3.0.0
info:
  title: FIBO REST API
  version: 1.0.0
  description: Financial deal management with FIBO ontology integration
  contact:
    name: ChatmanGPT
    url: https://chatmangpt.com

servers:
  - url: http://localhost:8001/api
    description: Local development server

paths:
  /deals:
    post:
      summary: Create a new financial deal
      operationId: createDeal
      tags:
        - Deals
      security:
        - bearerAuth: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateDealRequest'
      responses:
        '201':
          description: Deal created successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/DealResponse'
        '400':
          description: Invalid request (missing required fields, validation error)
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '500':
          description: Internal server error (ontology persistence failed)
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'

    get:
      summary: List all deals with pagination
      operationId: listDeals
      tags:
        - Deals
      security:
        - bearerAuth: []
      parameters:
        - name: limit
          in: query
          schema:
            type: integer
            default: 50
            maximum: 1000
          description: Maximum number of deals to return
        - name: offset
          in: query
          schema:
            type: integer
            default: 0
          description: Number of deals to skip
      responses:
        '200':
          description: List of deals
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/DealListResponse'
        '500':
          description: Internal server error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'

  /deals/{id}:
    get:
      summary: Get a single deal by ID
      operationId: getDeal
      tags:
        - Deals
      security:
        - bearerAuth: []
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
          description: Deal ID (e.g., 'd-1234567890123')
      responses:
        '200':
          description: Deal retrieved successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/DealResponse'
        '404':
          description: Deal not found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '500':
          description: Internal server error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'

    patch:
      summary: Update an existing deal
      operationId: updateDeal
      tags:
        - Deals
      security:
        - bearerAuth: []
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
          description: Deal ID
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/UpdateDealRequest'
      responses:
        '200':
          description: Deal updated successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/DealResponse'
        '400':
          description: Invalid request or no updates provided
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '404':
          description: Deal not found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '500':
          description: Internal server error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'

  /deals/{id}/verify-compliance:
    post:
      summary: Verify deal compliance against KYC, AML, SOX
      operationId: verifyCompliance
      tags:
        - Compliance
      security:
        - bearerAuth: []
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
          description: Deal ID
      responses:
        '200':
          description: Compliance verification completed
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ComplianceVerificationResponse'
        '400':
          description: Invalid deal ID
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '500':
          description: Internal server error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'

components:
  schemas:
    CreateDealRequest:
      type: object
      required:
        - name
        - amount
        - buyer_id
        - seller_id
      properties:
        name:
          type: string
          minLength: 1
          example: "Acme Corp Acquisition"
        amount:
          type: number
          format: double
          minimum: 0.01
          example: 5000000.00
        currency:
          type: string
          enum: [USD, EUR, GBP, JPY]
          default: USD
          example: "USD"
        buyer_id:
          type: string
          example: "party-buyer-123"
        seller_id:
          type: string
          example: "party-seller-456"
        expected_close_date:
          type: string
          format: date-time
          example: "2026-06-30T00:00:00Z"
        probability:
          type: integer
          minimum: 0
          maximum: 100
          example: 75
        stage:
          type: string
          enum: [prospecting, negotiation, due_diligence, closing]
          default: "prospecting"
          example: "negotiation"

    UpdateDealRequest:
      type: object
      properties:
        name:
          type: string
          minLength: 1
          example: "Updated Deal Name"
        amount:
          type: number
          format: double
          minimum: 0.01
          example: 5500000.00
        currency:
          type: string
          enum: [USD, EUR, GBP, JPY]
          example: "USD"
        status:
          type: string
          enum: [draft, active, closed, abandoned]
          example: "active"
        expected_close_date:
          type: string
          format: date-time
          example: "2026-07-15T00:00:00Z"
        probability:
          type: integer
          minimum: 0
          maximum: 100
          example: 85
        stage:
          type: string
          enum: [prospecting, negotiation, due_diligence, closing]
          example: "due_diligence"

    DealResponse:
      type: object
      properties:
        id:
          type: string
          example: "d-1234567890123"
        name:
          type: string
          example: "Acme Corp Acquisition"
        amount:
          type: number
          format: double
          example: 5000000.00
        currency:
          type: string
          example: "USD"
        status:
          type: string
          example: "active"
        buyer_id:
          type: string
          example: "party-buyer-123"
        seller_id:
          type: string
          example: "party-seller-456"
        expected_close_date:
          type: string
          format: date-time
          example: "2026-06-30T00:00:00Z"
        probability:
          type: integer
          example: 75
        stage:
          type: string
          example: "negotiation"
        created_at:
          type: string
          format: date-time
          example: "2026-03-25T10:30:00Z"
        updated_at:
          type: string
          format: date-time
          example: "2026-03-26T14:15:00Z"
        rdf_triple_count:
          type: integer
          description: Number of RDF triples generated by CONSTRUCT
          example: 12
        compliance_status:
          type: string
          enum: [pending_verification, verified, partial, failed]
          example: "pending_verification"
        kyc_verified:
          type: boolean
          example: false
        aml_screening:
          type: string
          enum: [pending, passed, failed]
          example: "pending"

    DealListResponse:
      type: object
      properties:
        deals:
          type: array
          items:
            $ref: '#/components/schemas/DealResponse'
        count:
          type: integer
          example: 25
        limit:
          type: integer
          example: 50
        offset:
          type: integer
          example: 0

    ComplianceVerificationResponse:
      type: object
      properties:
        deal_id:
          type: string
          example: "d-1234567890123"
        compliance:
          type: object
          properties:
            kyc_verified:
              type: boolean
              example: true
            kyc_triples:
              type: integer
              description: Number of KYC verification triples
              example: 3
            aml_screening:
              type: string
              enum: [passed, failed, unknown]
              example: "passed"
            aml_triples:
              type: integer
              description: Number of AML screening triples
              example: 2
            sox_compliant:
              type: boolean
              example: true
            sox_triples:
              type: integer
              description: Number of SOX compliance triples
              example: 1
            compliance_status:
              type: string
              enum: [verified, partial, failed]
              example: "verified"

    ErrorResponse:
      type: object
      properties:
        error:
          type: string
          example: "Deal creation failed: validation error"

  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
```

---

## API Endpoints

### 1. Create Deal

**Endpoint:** `POST /api/deals`
**Authentication:** Required (Bearer token)
**Content-Type:** `application/json`

#### Request Body

```json
{
  "name": "Acme Corp Acquisition",
  "amount": 5000000.00,
  "currency": "USD",
  "buyer_id": "party-buyer-123",
  "seller_id": "party-seller-456",
  "expected_close_date": "2026-06-30T00:00:00Z",
  "probability": 75,
  "stage": "negotiation"
}
```

#### Required Fields

| Field | Type | Validation |
|-------|------|-----------|
| `name` | string | Non-empty, required |
| `amount` | number | > 0, required |
| `buyer_id` | string | Non-empty, required |
| `seller_id` | string | Non-empty, required |

#### Optional Fields

| Field | Type | Default | Validation |
|-------|------|---------|-----------|
| `currency` | string | USD | One of: USD, EUR, GBP, JPY |
| `expected_close_date` | datetime | — | ISO 8601 or YYYY-MM-DD format |
| `probability` | integer | 0 | 0-100 inclusive |
| `stage` | string | prospecting | One of: prospecting, negotiation, due_diligence, closing |

#### Response

**Status:** 201 Created

```json
{
  "id": "d-1234567890123",
  "name": "Acme Corp Acquisition",
  "amount": 5000000.00,
  "currency": "USD",
  "status": "draft",
  "buyer_id": "party-buyer-123",
  "seller_id": "party-seller-456",
  "expected_close_date": "2026-06-30T00:00:00Z",
  "probability": 75,
  "stage": "negotiation",
  "created_at": "2026-03-25T10:30:00Z",
  "updated_at": "2026-03-25T10:30:00Z",
  "rdf_triple_count": 12,
  "compliance_status": "pending_verification",
  "kyc_verified": false,
  "aml_screening": "pending"
}
```

#### Errors

**400 Bad Request** — Missing required field or validation failure

```json
{
  "error": "deal name required"
}
```

**500 Internal Server Error** — Oxigraph CONSTRUCT failed

```json
{
  "error": "Deal creation failed: ontology persistence failed"
}
```

---

### 2. Get Deal

**Endpoint:** `GET /api/deals/{id}`
**Authentication:** Required

#### Path Parameters

| Parameter | Type | Required | Example |
|-----------|------|----------|---------|
| `id` | string | Yes | `d-1234567890123` |

#### Response

**Status:** 200 OK

```json
{
  "id": "d-1234567890123",
  "name": "Acme Corp Acquisition",
  "amount": 5000000.00,
  "currency": "USD",
  "status": "draft",
  "buyer_id": "party-buyer-123",
  "seller_id": "party-seller-456",
  "expected_close_date": "2026-06-30T00:00:00Z",
  "probability": 75,
  "stage": "negotiation",
  "created_at": "2026-03-25T10:30:00Z",
  "updated_at": "2026-03-25T10:30:00Z",
  "rdf_triple_count": 12,
  "compliance_status": "pending_verification",
  "kyc_verified": false,
  "aml_screening": "pending"
}
```

#### Errors

**404 Not Found**

```json
{
  "error": "Deal not found: deal retrieval failed"
}
```

---

### 3. List Deals

**Endpoint:** `GET /api/deals`
**Authentication:** Required

#### Query Parameters

| Parameter | Type | Default | Max | Description |
|-----------|------|---------|-----|-------------|
| `limit` | integer | 50 | 1000 | Results per page |
| `offset` | integer | 0 | — | Skip N results |

#### Example Request

```bash
curl -X GET "http://localhost:8001/api/deals?limit=100&offset=0" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

#### Response

**Status:** 200 OK

```json
{
  "deals": [
    {
      "id": "d-1234567890123",
      "name": "Acme Corp Acquisition",
      "amount": 5000000.00,
      "currency": "USD",
      "status": "draft",
      "buyer_id": "party-buyer-123",
      "seller_id": "party-seller-456",
      "expected_close_date": "2026-06-30T00:00:00Z",
      "probability": 75,
      "stage": "negotiation",
      "created_at": "2026-03-25T10:30:00Z",
      "updated_at": "2026-03-25T10:30:00Z",
      "rdf_triple_count": 12,
      "compliance_status": "pending_verification",
      "kyc_verified": false,
      "aml_screening": "pending"
    }
  ],
  "count": 1,
  "limit": 50,
  "offset": 0
}
```

---

### 4. Update Deal

**Endpoint:** `PATCH /api/deals/{id}`
**Authentication:** Required

#### Path Parameters

| Parameter | Type | Required |
|-----------|------|----------|
| `id` | string | Yes |

#### Request Body (all fields optional)

```json
{
  "name": "Updated Deal Name",
  "amount": 5500000.00,
  "currency": "EUR",
  "status": "active",
  "expected_close_date": "2026-07-15T00:00:00Z",
  "probability": 85,
  "stage": "due_diligence"
}
```

#### Response

**Status:** 200 OK

```json
{
  "id": "d-1234567890123",
  "name": "Updated Deal Name",
  "amount": 5500000.00,
  "currency": "EUR",
  "status": "active",
  "buyer_id": "party-buyer-123",
  "seller_id": "party-seller-456",
  "expected_close_date": "2026-07-15T00:00:00Z",
  "probability": 85,
  "stage": "due_diligence",
  "created_at": "2026-03-25T10:30:00Z",
  "updated_at": "2026-03-26T14:15:00Z",
  "rdf_triple_count": 15,
  "compliance_status": "pending_verification",
  "kyc_verified": false,
  "aml_screening": "pending"
}
```

#### Errors

**400 Bad Request** — No updates provided or validation failure

```json
{
  "error": "no updates provided"
}
```

**404 Not Found**

```json
{
  "error": "deal not found: deal retrieval failed"
}
```

---

### 5. Verify Compliance

**Endpoint:** `POST /api/deals/{id}/verify-compliance`
**Authentication:** Required
**Compliance Checks:** KYC, AML, SOX

#### Path Parameters

| Parameter | Type | Required |
|-----------|------|----------|
| `id` | string | Yes |

#### Response

**Status:** 200 OK

```json
{
  "deal_id": "d-1234567890123",
  "compliance": {
    "kyc_verified": true,
    "kyc_triples": 3,
    "aml_screening": "passed",
    "aml_triples": 2,
    "sox_compliant": true,
    "sox_triples": 1,
    "compliance_status": "verified"
  }
}
```

#### Compliance Status Values

| Value | Meaning |
|-------|---------|
| `verified` | All KYC, AML, SOX checks passed |
| `partial` | Some checks passed (e.g., KYC only) |
| `failed` | One or more checks failed |

#### Errors

**400 Bad Request** — Missing deal ID

```json
{
  "error": "deal_id parameter required"
}
```

**500 Internal Server Error**

```json
{
  "error": "Compliance verification failed: ..."
}
```

---

## Error Handling

All errors return a JSON object with an `error` field:

```json
{
  "error": "Description of what went wrong"
}
```

### HTTP Status Codes

| Code | Meaning | When |
|------|---------|------|
| 201 | Created | Deal successfully created |
| 200 | OK | Successful retrieval or update |
| 400 | Bad Request | Missing required fields, validation failure, invalid JSON |
| 404 | Not Found | Deal doesn't exist |
| 409 | Conflict | (Reserved for future use) |
| 500 | Internal Server Error | Oxigraph failure, database error, unhandled exception |

### Validation Errors

Request validation errors include detailed information:

```json
{
  "error": "deal amount must be positive"
}
```

Common validation failures:
- `deal name required` — name field is empty
- `deal amount must be positive` — amount ≤ 0
- `buyer_id required` — missing buyer_id
- `seller_id required` — missing seller_id
- `probability must be 0-100` — probability outside range
- `invalid datetime format` — expected_close_date doesn't parse

---

## Performance SLA

| Operation | Target P95 | P99 | Notes |
|-----------|-----------|-----|-------|
| Create Deal | 300ms | 500ms | Includes SPARQL CONSTRUCT execution |
| Get Deal | 150ms | 250ms | Read from RDF triplestore |
| List Deals (limit=50) | 200ms | 400ms | SPARQL query with LIMIT/OFFSET |
| Update Deal | 250ms | 450ms | Verification + CONSTRUCT |
| Verify Compliance | 500ms | 1000ms | Parallel KYC/AML/SOX checks |

**Assumptions:**
- Oxigraph responds within 100-300ms
- Network latency < 50ms
- PostgreSQL queries < 50ms (optional indexing)

**Bottlenecks:**
- Oxigraph timeout: 5000ms (configured in service)
- Slow SPARQL query: > 1 second on 100k+ triples

---

## SPARQL Integration

All deal operations execute SPARQL CONSTRUCT queries against Oxigraph.

### Create Deal CONSTRUCT Example

```sparql
PREFIX : <https://businessos.dev/id/>
PREFIX fibo-fnd: <https://spec.edmcouncil.org/fibo/ontology/FND/>
PREFIX fibo-fbc: <https://spec.edmcouncil.org/fibo/ontology/FBC/>
PREFIX xsd: <http://www.w3.org/2001/XMLSchema#>
PREFIX schema: <https://schema.org/>

CONSTRUCT {
  :deals/d-1234567890123 a :Deal ;
    a fibo-fnd:Agreement ;
    a fibo-fbc:FinancialInstrument ;
    :dealIdentifier "d-1234567890123" ;
    :dealName "Acme Corp Acquisition" ;
    :dealAmount 5000000.0 ;
    :dealCurrency "USD" ;
    :dealStatus "draft" ;
    :hasPrimaryBuyer :parties/party-buyer-123 ;
    :hasPrimarySeller :parties/party-seller-456 ;
    :expectedCloseDate "2026-06-30T00:00:00Z"^^xsd:dateTime ;
    :dealProbability "75"^^xsd:integer ;
    :currentStage "negotiation" ;
    schema:dateCreated "2026-03-25T10:30:00Z"^^xsd:dateTime .
}
WHERE { BIND(TRUE as ?dummy) }
```

### Compliance Verification CONSTRUCT Example

```sparql
PREFIX : <https://businessos.dev/id/>
PREFIX fibo-be: <https://spec.edmcouncil.org/fibo/ontology/BE/>

CONSTRUCT {
  ?party :hasKYCStatus :KYCVerified ;
    :kycVerificationDate ?kycDate ;
    :kycExpiryDate ?expiryDate .
}
WHERE {
  BIND(:deals/d-1234567890123 as ?deal)
  ?deal :hasPrimaryBuyer ?party ;
    :hasPrimarySeller ?sellerParty .
  ?party :hasKYCStatus :KYCVerified ;
    :kycExpiryDate ?expiryDate .
  FILTER(?expiryDate > NOW())
}
```

---

## Code Examples

### Go Client (using standard library)

```go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func createDeal(token string) error {
	client := &http.Client{}

	payload := map[string]interface{}{
		"name":        "Acme Corp Acquisition",
		"amount":      5000000.00,
		"currency":    "USD",
		"buyer_id":    "party-buyer-123",
		"seller_id":   "party-seller-456",
		"probability": 75,
		"stage":       "negotiation",
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "http://localhost:8001/api/deals", bytes.NewReader(body))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("status %d", resp.StatusCode)
	}

	return nil
}
```

### cURL

```bash
# Create deal
curl -X POST http://localhost:8001/api/deals \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Acme Corp Acquisition",
    "amount": 5000000.00,
    "currency": "USD",
    "buyer_id": "party-buyer-123",
    "seller_id": "party-seller-456",
    "probability": 75,
    "stage": "negotiation"
  }'

# Get deal
curl -X GET http://localhost:8001/api/deals/d-1234567890123 \
  -H "Authorization: Bearer YOUR_TOKEN"

# List deals with pagination
curl -X GET "http://localhost:8001/api/deals?limit=100&offset=0" \
  -H "Authorization: Bearer YOUR_TOKEN"

# Update deal
curl -X PATCH http://localhost:8001/api/deals/d-1234567890123 \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated Name",
    "probability": 85,
    "stage": "due_diligence"
  }'

# Verify compliance
curl -X POST http://localhost:8001/api/deals/d-1234567890123/verify-compliance \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

## Configuration

### Environment Variables

| Variable | Default | Purpose |
|----------|---------|---------|
| `OXIGRAPH_URL` | `http://localhost:8890` | Oxigraph RDF store endpoint |
| `DEALS_TIMEOUT_MS` | `5000` | Max wait time for ontology query |

### Service Initialization

```go
dealsService := services.NewFIBODealsService(os.Getenv("OXIGRAPH_URL"))
handler := handlers.NewFIBODealsHandler(dealsService)
api := router.Group("/api")
handlers.RegisterFIBODealsRoutes(api, handler, authMiddleware)
```

---

## Testing

### Unit Tests (13 tests, all passing)

```bash
cd BusinessOS/desktop/backend-go
go test ./internal/handlers/... -run "TestHandler" -v
```

**Test Coverage:**
- ✅ Create deal (valid input)
- ✅ Create deal (missing required fields)
- ✅ Create deal (invalid amount)
- ✅ Get deal (success)
- ✅ List deals (success)
- ✅ Update deal (success)
- ✅ Update deal (missing ID)
- ✅ Verify compliance (success)
- ✅ Invalid JSON request
- ✅ List deals with pagination
- ✅ Create deal with expected_close_date
- ✅ Validate probability bounds
- ✅ Update deal with multiple fields

### Test Database Setup

```bash
# Start Oxigraph test server
docker run -p 8890:7878 oxigraph/oxigraph:latest serve --bind 0.0.0.0

# Run tests
go test ./internal/... -count=1 -v
```

---

## Compliance Frameworks

### SOC2 (System and Organization Controls)

**Scope:** CC6 (Logical Access Control), CC8 (Audit Trails)
**Verification:** Business rules enforce user identity + activity logging

### GDPR (General Data Protection Regulation)

**Scope:** Article 32 (security), Article 30 (records of processing)
**Verification:** Audit trail in RDF, encrypted in transit (TLS)

### HIPAA (Health Insurance Portability)

**Scope:** §164.312 (Technical safeguards), §164.308 (Administrative)
**Verification:** Access control + audit logging (if PHI involved)

### SOX (Sarbanes-Oxley)

**Scope:** §302 (Management certification), §302 (Internal controls)
**Verification:** Deal transaction audit trail, immutable records

---

## Troubleshooting

### Issue: 500 "ontology persistence failed"

**Cause:** Oxigraph is unreachable or CONSTRUCT query fails
**Solution:**
1. Check Oxigraph is running: `curl http://localhost:8890/query`
2. Verify SPARQL syntax in service logs
3. Check deal data for invalid characters (quotes, newlines)

### Issue: 404 "Deal not found"

**Cause:** Deal ID doesn't exist in RDF store
**Solution:**
1. Verify deal was created (check creation response)
2. Check deal ID format (should be `d-` prefix)
3. Query Oxigraph directly: `SPARQL SELECT * WHERE { ?deal a :Deal . }`

### Issue: Timeout after 5s

**Cause:** Oxigraph query is slow (too many triples or complex query)
**Solution:**
1. Add LIMIT to queries
2. Index frequently queried predicates in Oxigraph
3. Increase timeout (edit `services/fibo_deals.go`)

---

## Roadmap

**Phase 2 (Q2 2026):**
- [ ] Batch create/update operations
- [ ] Deal search by name, amount, stage
- [ ] Deal workflow state machine (prospecting → closed)
- [ ] Deal metrics dashboard (total value, close rate)

**Phase 3 (Q3 2026):**
- [ ] Deal templates (SaaS, Loan, Defense contract)
- [ ] Multi-currency conversion
- [ ] Deal comparison API
- [ ] Historical deal analysis

---

## Support

For issues or questions:
- GitHub Issues: https://github.com/seanchatmangpt/chatmangpt/issues
- Email: info@chatmangpt.com
- Slack: #businessos-development

---

**Last Updated:** 2026-03-26
**API Version:** 1.0.0
**Status:** Production
