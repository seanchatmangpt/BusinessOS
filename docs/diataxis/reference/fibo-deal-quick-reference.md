# FIBO Deal Integration - Quick Reference

**Cheat Sheet for FIBO ontology usage in BusinessOS deals & contracts**

---

## Class Mapping Cheat Sheet

```
BUSINESSOS → FIBO MAPPING

Deal
  ├── class: :Deal
  ├── superclass: fibo-fnd:Agreement + fibo-fbc:FinancialInstrument
  └── properties:
      - :dealIdentifier (string UUID)
      - :dealName (string)
      - :dealStatus (:DealStatus)
      - :hasPrimaryBuyer (:BusinessOSParty)
      - :hasPrimarySeller (:BusinessOSParty)
      - :dealAmount (:MonetaryAmount)
      - :dealCurrency (string: ISO 4217)
      - :expectedCloseDate (xsd:dateTime)
      - :currentStage (:DealLifecycleStage)

Party
  ├── class: :BusinessOSParty
  ├── superclass: fibo-be:LegalPerson + fibo-fnd:Party
  ├── subclasses:
  │   ├── :InternalParty (our organization)
  │   └── :Counterparty (external party)
  └── properties:
      - :partyIdentifier (string)
      - :legalName (string)
      - :businessRegistrationNumber (string)
      - :operatingJurisdiction (string: ISO 3166-1)
      - :partyRole (:DealRole)
      - :hasKYCStatus (:KYCStatus)
      - :kycVerificationDate (xsd:dateTime)
      - :kycExpiryDate (xsd:dateTime)
      - :amlScreeningResult (string)

MonetaryAmount
  ├── class: :MonetaryAmount
  ├── superclass: fibo-fnd:MonetaryAmount
  └── properties:
      - :amount (xsd:decimal)
      - :currency (string: ISO 4217)

Contract
  ├── class: :Contract
  ├── superclass: fibo-fnd:Agreement + fibo-fei:ContractualElement
  └── properties:
      - :dealName (string)
      - :hasContractualTerm (:ContractualTerm)

ContractualTerm
  ├── subclasses:
  │   ├── :PricingTerm (fibo-fnd:ContractualDefinition)
  │   ├── :PaymentTerm (fibo-fnd:PaymentObligation)
  │   ├── :LiabilityTerm (fibo-fnd:ContractualDefinition)
  │   ├── :ConfidentialityTerm (fibo-fnd:ContractualDefinition)
  │   ├── :TerminationTerm (fibo-fnd:ContractualDefinition)
  │   └── :ComplianceTerm (fibo-fnd:ContractualDefinition)
  └── properties:
      - :termDescription (string)
      - :termCondition (string)
```

---

## Status & Enumeration Values

### DealStatus
```
:DealPipeline       → Deal in initial pipeline
:DealNegotiation    → Terms being negotiated
:DealSigned         → Agreement signed
:DealClosedWon      → Deal successfully completed
:DealClosedLost     → Deal terminated/lost
```

### DealLifecycleStage
```
:InitiationStage      → Deal created (day 0)
:QualificationStage   → BANT evaluation (week 1)
:NegotiationStage     → Terms discussion (week 2-4)
:SignatureStage       → Contract signatures collected (week 4-5)
:ExecutionStage       → Obligations begin (week 5+)
:ClosureStage         → Deal complete or terminated
```

### KYCStatus
```
:KYCPending     → Verification in progress
:KYCVerified    → Passed verification (current)
:KYCRejected    → Failed verification (blocked)
:KYCExpired     → Verification expired (renewal needed)
```

### DealRole (Party roles in deal)
```
:BuyerRole              → Purchaser/customer
:SellerRole             → Provider/vendor
:BrokerRole             → Intermediary
:GuarantorRole          → Guarantees performance (fibo-fei:Guarantor)
:UnderwriterRole        → Assumes risk (fibo-fei:Underwriter)
```

### ComplianceFramework
```
:SOXCompliance          → Sarbanes-Oxley (financial controls)
:GDPRCompliance         → General Data Protection Regulation (EU data)
:FINRACompliance        → Financial Industry Regulatory Authority
(extensible for HIPAA, CCPA, industry-specific)
```

---

## URI Patterns

```
Deal:        https://businessos.dev/id/deals/{deal_id}
Party:       https://businessos.dev/id/parties/{party_id}
Amount:      https://businessos.dev/id/amounts/{deal_id}-{type}
Contract:    https://businessos.dev/id/contracts/{contract_id}
Term:        https://businessos.dev/id/terms/{deal_id}-{term_type}
Audit:       https://businessos.dev/id/audits/{event_id}
Compliance:  https://businessos.dev/id/requirements/{requirement_id}
KYC Report:  https://businessos.dev/id/reports/kyc-{uuid}
```

---

## Common SPARQL Query Patterns

### Find all verified counterparties
```sparql
SELECT ?party ?legalName ?kyc_status
WHERE {
  ?party a :Counterparty ;
    :legalName ?legalName ;
    :hasKYCStatus :KYCVerified .
}
```

### Find deals in negotiation
```sparql
SELECT ?deal ?name ?amount ?currency
WHERE {
  ?deal a :Deal ;
    :dealName ?name ;
    :dealStatus :DealNegotiation ;
    :dealAmount ?amt .
  ?amt :amount ?amount ;
       :currency ?currency .
}
ORDER BY DESC(?amount)
```

### Find deals by jurisdiction
```sparql
SELECT ?deal ?buyer ?seller
WHERE {
  ?deal a :Deal ;
    :hasPrimaryBuyer ?buyer ;
    :hasPrimarySeller ?seller .
  ?buyer :operatingJurisdiction "US" .
}
```

### Check compliance status
```sparql
SELECT ?deal ?framework ?status
WHERE {
  ?deal a :Deal ;
    :hasComplianceRequirement ?req .
  ?req :complianceFramework ?framework ;
       :complianceStatus ?status .
}
FILTER (?status != "compliant")
```

### Find non-KYC verified parties
```sparql
SELECT ?party ?name ?kyc_status ?expiry
WHERE {
  ?party a :BusinessOSParty ;
    :legalName ?name ;
    :hasKYCStatus ?kyc_status .
  OPTIONAL { ?party :kycExpiryDate ?expiry . }
  FILTER (?kyc_status != :KYCVerified)
}
```

---

## RDF Turtle Snippets

### Minimal deal
```turtle
<https://businessos.dev/id/deals/d-001>
  a :Deal ;
  :dealIdentifier "d-001" ;
  :dealName "Cloud Services" ;
  :dealStatus :DealNegotiation ;
  :dealAmount <https://businessos.dev/id/amounts/d-001-principal> ;
  :dealCurrency "USD" ;
  :currentStage :NegotiationStage .

<https://businessos.dev/id/amounts/d-001-principal>
  a :MonetaryAmount ;
  :amount "250000"^^xsd:decimal ;
  :currency "USD" .
```

### Buyer + Seller with roles
```turtle
<https://businessos.dev/id/parties/buyer-001>
  a :BusinessOSParty, :InternalParty ;
  :partyIdentifier "buyer-001" ;
  :legalName "Acme Corp" ;
  :partyRole :BuyerRole ;
  :hasKYCStatus :KYCVerified ;
  :kycExpiryDate "2026-06-01T00:00:00Z"^^xsd:dateTime .

<https://businessos.dev/id/parties/seller-001>
  a :BusinessOSParty, :Counterparty ;
  :partyIdentifier "seller-001" ;
  :legalName "TechVendor Inc" ;
  :partyRole :SellerRole ;
  :hasKYCStatus :KYCVerified ;
  :kycExpiryDate "2026-08-15T00:00:00Z"^^xsd:dateTime .
```

### Contract with terms
```turtle
<https://businessos.dev/id/contracts/c-001>
  a :Contract ;
  :dealName "MSA" ;
  :hasContractualTerm
    <https://businessos.dev/id/terms/c-001-pricing>,
    <https://businessos.dev/id/terms/c-001-payment>,
    <https://businessos.dev/id/terms/c-001-liability> .

<https://businessos.dev/id/terms/c-001-pricing>
  a :PricingTerm ;
  :termDescription "Monthly SaaS fee: $20,833" ;
  :termCondition "Annual prepayment or monthly billing" .

<https://businessos.dev/id/terms/c-001-payment>
  a :PaymentTerm ;
  :termDescription "Net 30 from invoice" ;
  :termCondition "Due within 30 days of receipt" .

<https://businessos.dev/id/terms/c-001-liability>
  a :LiabilityTerm ;
  :termDescription "Liability capped at 12 months of fees" ;
  :termCondition "Excludes data loss from third-party attacks" .
```

---

## File Locations

| File | Purpose |
|------|---------|
| `ontologies/fibo-deal-integration.ttl` | Core ontology definition (150+ statements) |
| `ontologies/fibo-deal-sparql-queries.rq` | 3 CONSTRUCT queries |
| `ontologies/fibo-deal-examples.ttl` | Example deals in RDF format |
| `docs/diataxis/explanation/fibo-deal-integration-complete.md` | Full documentation |
| `docs/diataxis/reference/fibo-deal-quick-reference.md` | This file |

---

## SPARQL CONSTRUCT Queries Summary

### Query 1: Deal Creation
**Input:** SQL deal data
**Output:** FIBO-compliant deal RDF with parties, amounts, terms
**Location:** `fibo-deal-sparql-queries.rq` (lines 29-113)

### Query 2: Compliance Checking
**Input:** Deal + Contract URIs
**Output:** Compliance validation report (pass/fail on 4 checks)
**Location:** `fibo-deal-sparql-queries.rq` (lines 131-215)
**Checks:**
- Required contractual terms present?
- Party KYC verified?
- Deal amount authorized?
- Jurisdiction compliant?

### Query 3: KYC/AML Verification
**Input:** Party URI
**Output:** KYC verification report (status, AML screening, confidence score)
**Location:** `fibo-deal-sparql-queries.rq` (lines 226-311)
**Validates:**
- Identity verification (registration number)
- AML screening (OFAC/EU/UN sanctions)
- Beneficial owner restrictions
- Jurisdiction risks

---

## Integration Points

### With PostgreSQL (deals table)
```
SQL: INSERT INTO deals (id, name, amount, currency, buyer_id, seller_id, status)
  ↓ R2RML mapping
RDF: INSERT DATA { <https://businessos.dev/id/deals/d-001> a :Deal; ... }
  ↓ SPARQL CONSTRUCT Query 1
RDF: Full deal with parties, amounts, lifecycle stage
```

### With compliance checks
```
API: POST /api/deals/compliance { deal_uri: "..." }
  ↓ Execute Query 2
RDF: :ComplianceValidationReport with status (COMPLIANT/NON_COMPLIANT)
  ↓ Update deal status
DB: UPDATE deals SET status = 'NEGOTIATION' WHERE id = 'd-001'
```

### With KYC/AML verification
```
API: POST /api/parties/verify-kyc { party_id: "..." }
  ↓ Execute Query 3
RDF: :KYCVerificationReport with decision (APPROVED/PENDING_REVIEW)
  ↓ Update party
DB: UPDATE parties SET kyc_status = 'verified' WHERE id = 'p-001'
```

---

## Troubleshooting 1-Liners

| Problem | Fix |
|---------|-----|
| Namespace errors | Check @prefix in .ttl file matches FIBO specs |
| Query returns 0 results | FILTER might be wrong; try without FILTER first |
| xsd:dateTime parsing fails | Use format: "2026-03-25T14:30:00Z"^^xsd:dateTime |
| Party KYC won't verify | Check businessRegistrationNumber is not empty |
| Deal amount comparison fails | Use xsd:decimal not integer for currency |
| Audit trail missing | Create :DealAuditEntry after every deal update |
| Compliance check too slow | Add FILTER to reduce result set before JOIN |

---

## One-Minute Setup

1. **Load ontology**
   ```bash
   curl -X POST http://localhost:8890/sparql \
     -H "Content-Type: text/turtle" \
     --data-binary @fibo-deal-integration.ttl
   ```

2. **Insert example data**
   ```bash
   curl -X POST http://localhost:8890/sparql \
     -H "Content-Type: text/turtle" \
     --data-binary @fibo-deal-examples.ttl
   ```

3. **Test Query 1 (Deal)**
   ```sparql
   SELECT ?deal ?name ?amount
   WHERE {
     ?deal a :Deal ;
       :dealName ?name ;
       :dealAmount ?amt .
     ?amt :amount ?amount .
   }
   LIMIT 5
   ```

4. **Test Query 2 (Compliance)**
   - See fibo-deal-sparql-queries.rq line 131+

5. **Test Query 3 (KYC)**
   - See fibo-deal-sparql-queries.rq line 226+

---

**Last Updated:** 2026-03-25
**Ontology Version:** 1.0.0
