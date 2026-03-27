# FIBO Deal Integration for BusinessOS

**Status:** Complete (2026-03-25)
**Version:** 1.0.0
**Audience:** Financial Systems Engineers, Data Architects, Compliance Officers

## Overview

This document describes the integration of **Financial Industry Business Ontology (FIBO)** with BusinessOS deal, contract, and counterparty management systems. FIBO provides semantic models for financial concepts aligned with ISO standards and regulatory frameworks.

### Purpose

Map BusinessOS CRM deal structures to FIBO ontology to enable:
- Regulatory compliance validation (SOX, GDPR, FINRA)
- Machine-readable financial semantics
- Knowledge graph queries across deal portfolio
- Automated KYC/AML verification workflows
- Audit trail generation (PROV-O compatible)

---

## 1. FIBO Architecture for Deals

### 1.1 FIBO Namespace Mapping

BusinessOS uses three primary FIBO modules:

| FIBO Module | Purpose | BusinessOS Mapping |
|-------------|---------|-------------------|
| **FND** (Foundations) | Core business concepts | Agreements, Parties, MonetaryAmount |
| **FBC** (Financial Business and Commerce) | Financial instruments | FinancialInstrument, Transactions |
| **FEI** (Financial Economics and Industry) | Industry-specific concepts | Contracts, Guarantees, Underwriting |
| **BE** (Business Entities) | Legal entities and organizations | Party, LegalPerson, Organization |

**Namespace prefixes (Turtle/RDF):**

```turtle
@prefix fibo-fnd: <https://spec.edmcouncil.org/fibo/ontology/FND/> .
@prefix fibo-fbc: <https://spec.edmcouncil.org/fibo/ontology/FBC/> .
@prefix fibo-fei: <https://spec.edmcouncil.org/fibo/ontology/FEI/> .
@prefix fibo-be: <https://spec.edmcouncil.org/fibo/ontology/BE/> .
```

### 1.2 Core Class Hierarchy

```
fibo-fnd:Agreement
  ├── :Deal (BusinessOS deal entity)
  │   ├── fibo-fbc:FinancialInstrument
  │   └── :ContractualTerm (pricing, payment, liability)
  └── :Contract (signed agreement)

fibo-fnd:Party / fibo-be:LegalPerson
  ├── :BusinessOSParty
  │   ├── :InternalParty (buyer in our org)
  │   └── :Counterparty (external vendor/partner)
  │       ├── :BuyerRole
  │       ├── :SellerRole
  │       ├── :BrokerRole
  │       ├── :GuarantorRole (fibo-fei:Guarantor)
  │       └── :UnderwriterRole

fibo-fnd:MonetaryAmount
  └── :MonetaryAmount (amount + currency ISO 4217)
```

---

## 2. Domain Model Mapping

### 2.1 Deal Entity

BusinessOS **Deal** maps to FIBO **fibo-fnd:Agreement** + **fibo-fbc:FinancialInstrument**.

**RDF Representation:**

```turtle
<https://businessos.dev/id/deals/d-001>
  a :Deal ;
  a fibo-fnd:Agreement ;
  a fibo-fbc:FinancialInstrument ;
  :dealIdentifier "d-001" ;
  :dealName "Cloud Infrastructure Deal" ;
  :dealStatus :DealNegotiation ;
  :hasPrimaryBuyer <https://businessos.dev/id/parties/acme-corp> ;
  :hasPrimarySeller <https://businessos.dev/id/parties/cloudtech-inc> ;
  :dealAmount <https://businessos.dev/id/amounts/principal-001> ;
  :dealCurrency "USD" ;
  :expectedCloseDate "2026-06-30T00:00:00Z"^^xsd:dateTime ;
  :dealProbability "85"^^xsd:integer ;
  :currentStage :NegotiationStage ;
  schema:dateCreated "2026-01-15T10:30:00Z"^^xsd:dateTime .
```

**SQL to RDF Mapping:**

| SQL Column | RDF Property | FIBO Equivalent | Range |
|-----------|--------------|-----------------|-------|
| `id` | `:dealIdentifier` | `fibo-fnd:hasIdentifier` | xsd:string |
| `name` | `:dealName` | `fibo-fnd:hasName` | xsd:string |
| `amount` | `:dealAmount` (ObjectProperty) | `fibo-fnd:hasMonetaryAmount` | :MonetaryAmount |
| `currency` | `:dealCurrency` | `fibo-fnd:hasCurrency` | xsd:string (ISO 4217) |
| `status` | `:dealStatus` | `fibo-fnd:hasStatus` | :DealStatus |
| `buyer_id` | `:hasPrimaryBuyer` | `fibo-fnd:hasParty` | :BusinessOSParty |
| `seller_id` | `:hasPrimarySeller` | `fibo-fnd:hasParty` | :BusinessOSParty |
| `expected_close_date` | `:expectedCloseDate` | `fibo-fnd:hasTerminationDate` | xsd:dateTime |
| `probability` | `:dealProbability` | (custom) | xsd:integer [0-100] |
| `stage` | `:currentStage` | (custom: BPMN extension) | :DealLifecycleStage |

### 2.2 Party Entity

BusinessOS **Party** (buyer/seller/broker/guarantor) maps to **fibo-be:LegalPerson** + **fibo-fnd:Party**.

**Party Roles** define function in deal:

```turtle
<https://businessos.dev/id/parties/acme-corp>
  a :BusinessOSParty ;
  a :InternalParty ;                    # Our organization
  a fibo-be:LegalPerson ;
  a fibo-fnd:Party ;
  :partyIdentifier "acme-corp" ;
  :legalName "Acme Corporation, Inc." ;
  :businessRegistrationNumber "12-3456789" ;
  :operatingJurisdiction "US" ;
  :partyRole :BuyerRole ;
  :hasKYCStatus :KYCVerified ;
  :kycVerificationDate "2025-06-01T00:00:00Z"^^xsd:dateTime ;
  :kycExpiryDate "2026-06-01T00:00:00Z"^^xsd:dateTime ;
  :amlScreeningResult "NO_MATCH" ;
  schema:name "Acme Corporation" ;
  schema:url "https://acmecorp.com" .
```

**Party Role Hierarchy:**

```
:DealRole (fibo-be:Role)
  ├── :BuyerRole (purchaser)
  ├── :SellerRole (provider)
  ├── :BrokerRole (intermediary)
  ├── :GuarantorRole (fibo-fei:Guarantor - assumes obligation)
  └── :UnderwriterRole (fibo-fei:Underwriter - assumes risk)
```

### 2.3 Monetary Amount

Precise financial amounts with currency:

```turtle
<https://businessos.dev/id/amounts/principal-001>
  a :MonetaryAmount ;
  a fibo-fnd:MonetaryAmount ;
  :amount "250000"^^xsd:decimal ;
  :currency "USD" ;
  schema:priceCurrency "USD" ;
  rdfs:comment "Annual recurring revenue (ARR)" .
```

**Properties:**

- `:amount` - Numeric value (xsd:decimal for precision)
- `:currency` - ISO 4217 code (USD, EUR, GBP, JPY, etc.)

### 2.4 Contract & Contractual Terms

**Contract** = formalized agreement with specific terms.

```turtle
<https://businessos.dev/id/contracts/c-001>
  a :Contract ;
  a fibo-fnd:Agreement ;
  a fibo-fei:ContractualElement ;
  :dealName "Cloud Services Master Service Agreement (MSA)" ;
  :hasContractualTerm <https://businessos.dev/id/terms/pricing-001> ;
  :hasContractualTerm <https://businessos.dev/id/terms/payment-001> ;
  :hasContractualTerm <https://businessos.dev/id/terms/liability-001> ;
  :hasContractualTerm <https://businessos.dev/id/terms/confidentiality-001> ;
  :hasContractualTerm <https://businessos.dev/id/terms/termination-001> .
```

**Contractual Term Classes** (extend :ContractualTerm / fibo-fnd:ContractualDefinition):

| Term Type | Purpose | Example Condition |
|-----------|---------|------------------|
| **PricingTerm** | Principal amount and valuation | "Annual fee: $250K" |
| **PaymentTerm** | Payment schedule and due dates | "Net 30 from invoice" |
| **LiabilityTerm** | Liability limits and indemnification | "Capped at 12 months fees" |
| **ConfidentialityTerm** | Confidential information protection | "5-year confidentiality obligation" |
| **TerminationTerm** | Exit and wind-down procedures | "90-day termination notice" |
| **ComplianceTerm** | Regulatory/compliance requirements | "SOC2 Type II certification required" |

---

## 3. Deal Lifecycle States (BPMN Integration)

Deals progress through lifecycle stages (BPMN-compatible):

```turtle
:DealLifecycleStage
  ├── :InitiationStage       # Deal created
  ├── :QualificationStage    # BANT evaluation
  ├── :NegotiationStage      # Terms discussion
  ├── :SignatureStage        # Contract signed
  ├── :ExecutionStage        # Obligations begin
  └── :ClosureStage          # Complete or terminated
```

**State Transitions** (with audit trail):

```turtle
?deal :currentStage :NegotiationStage ;
      :stageTransitionDate "2026-03-20T14:45:00Z"^^xsd:dateTime ;
      :stageDurationDays 35 .
```

---

## 4. KYC/AML Verification

### 4.1 KYC (Know Your Customer) Status

Party verification through identity, business registration, sanctions screening:

```turtle
?party :hasKYCStatus :KYCVerified ;
       :kycVerificationDate "2025-06-01T00:00:00Z"^^xsd:dateTime ;
       :kycExpiryDate "2026-06-01T00:00:00Z"^^xsd:dateTime ;
       :amlScreeningResult "NO_MATCH" ;
       :amlScreeningDate "2025-06-01T00:00:00Z"^^xsd:dateTime .
```

**KYC Status Values:**

- `:KYCPending` - Verification not completed
- `:KYCVerified` - Verification passed, current
- `:KYCRejected` - Verification failed, party not approved
- `:KYCExpired` - Verification expired, renewal required

### 4.2 AML Screening

Anti-Money Laundering screening against sanctions lists (OFAC, EU, UN):

```turtle
?party :amlScreeningResult "NO_MATCH"^^xsd:string ;
       :amlScreeningDate "2025-06-01T00:00:00Z"^^xsd:dateTime .
```

---

## 5. Compliance Framework Integration

Deal compliance validation against regulatory frameworks:

```turtle
<https://businessos.dev/id/deals/d-001>
  :hasComplianceRequirement <https://businessos.dev/id/requirements/sox-requirement> ;
  :hasComplianceRequirement <https://businessos.dev/id/requirements/gdpr-requirement> ;
  :hasComplianceRequirement <https://businessos.dev/id/requirements/finra-requirement> .

<https://businessos.dev/id/requirements/sox-requirement>
  a :ComplianceRequirement ;
  :complianceFramework :SOXCompliance ;
  :complianceStatus "compliant" ;
  :complianceValidationDate "2026-03-20T00:00:00Z"^^xsd:dateTime ;
  rdfs:comment "SOX IT General Controls verified" .
```

**Compliance Frameworks:**

- `:SOXCompliance` - Sarbanes-Oxley (financial reporting)
- `:GDPRCompliance` - General Data Protection Regulation (EU data)
- `:FINRACompliance` - Financial Industry Regulatory Authority (brokers)
- (extensible for HIPAA, CCPA, industry-specific)

---

## 6. SPARQL CONSTRUCT Queries

### 6.1 Deal Creation Query

**Purpose:** Transform SQL deal data → FIBO RDF

```sparql
CONSTRUCT {
  ?deal_uri a :Deal ;
    a fibo-fnd:Agreement ;
    a fibo-fbc:FinancialInstrument ;
    :dealIdentifier ?deal_id ;
    :dealName ?deal_name ;
    :hasPrimaryBuyer ?buyer_uri ;
    :hasPrimarySeller ?seller_uri ;
    :dealAmount ?amount_uri ;
    :dealCurrency ?currency ;
    :expectedCloseDate ?expected_close_date ;
    :currentStage ?stage .

  ?amount_uri a :MonetaryAmount ;
    :amount ?amount_value ;
    :currency ?currency .

  ?buyer_uri a :BusinessOSParty ;
    a fibo-fnd:Party ;
    :partyIdentifier ?buyer_id ;
    :legalName ?buyer_name ;
    :partyRole :BuyerRole .
}
WHERE {
  VALUES (?deal_id ?deal_name ?amount_value ?currency ...) {
    ("d-001" "Cloud Deal" "250000"^^xsd:decimal "USD" ...)
  }
  BIND(IRI(CONCAT(...)) AS ?deal_uri)
  BIND(IRI(CONCAT(...)) AS ?amount_uri)
  ...
}
```

### 6.2 Compliance Checking Query

**Purpose:** Validate deal against compliance requirements

```sparql
CONSTRUCT {
  ?report_uri a :ComplianceValidationReport ;
    :dealUri ?deal_uri ;
    :overallStatus ?overall_status ;
    :hasValidation ?validation_uri ;
    :recommendedActions ?action_text .

  ?validation_uri a :ComplianceCheck ;
    :checkName "Required Contractual Terms" ;
    :checkResult ?terms_result ;
    :severityLevel "critical" .
}
WHERE {
  ?deal_uri a :Deal ;
    :hasPrimaryBuyer ?buyer_uri ;
    :hasPrimarySeller ?seller_uri ;
    :dealAmount ?amount_uri .

  ?contract_uri a :Contract ;
    :hasContractualTerm ?pricing_term_uri ;
    :hasContractualTerm ?payment_term_uri ;
    :hasContractualTerm ?liability_term_uri .

  ?buyer_uri :hasKYCStatus ?buyer_kyc_status .
  ?seller_uri :hasKYCStatus ?seller_kyc_status .

  BIND(IF(..., "PASS"@en, "FAIL"@en) AS ?terms_result)
  BIND(IF(..., "COMPLIANT"@en, "NON_COMPLIANT"@en) AS ?overall_status)
}
```

### 6.3 Party Verification (KYC/AML) Query

**Purpose:** Generate KYC/AML verification report

```sparql
CONSTRUCT {
  ?kyc_report_uri a :KYCVerificationReport ;
    :partyUri ?party_uri ;
    :partyName ?legal_name ;
    :verificationStatus ?verification_status ;
    :overallDecision ?kyc_decision ;
    :confidenceScore ?confidence_score ;
    :amlScreeningResult ?aml_match_status ;
    :expiryDate ?calculated_expiry_date .

  ?party_uri :hasKYCStatus ?new_kyc_status ;
    :kycVerificationDate ?report_date ;
    :kycExpiryDate ?calculated_expiry_date ;
    :amlScreeningDate ?report_date .
}
WHERE {
  ?party_uri a fibo-be:LegalPerson ;
    :legalName ?legal_name ;
    :businessRegistrationNumber ?registration_number ;
    :operatingJurisdiction ?jurisdiction .

  BIND(NOW() AS ?report_date)
  BIND(IF(BOUND(?registration_number), "PASS"@en, "FAIL"@en) AS ?id_check)
  BIND("NO_MATCH"@en AS ?aml_match_status)
  BIND(IF(..., :KYCVerified, :KYCPending) AS ?new_kyc_status)
}
```

---

## 7. Deal Lifecycle Use Cases

### 7.1 Deal Creation Workflow

1. **User creates deal in BusinessOS UI**
   - Input: name, amount, buyer, seller, expected close date
   - System: Insert into PostgreSQL deals table

2. **Insert RDF into triplestore**
   - SPARQL INSERT DATA with deal facts
   - OR run SPARQL CONSTRUCT query to generate RDF from SQL

3. **Run compliance checks**
   - Query 2 (Compliance Checking) validates:
     - Required contract terms present?
     - Buyer KYC verified?
     - Seller KYC verified?
     - Amount within authorized limits?
     - Jurisdiction compliant?

4. **Update deal status**
   - If compliant → Move to Negotiation stage
   - If non-compliant → Flag for review

### 7.2 Party Onboarding (KYC/AML)

1. **New counterparty added to deal**
   - User provides party name, registration number, jurisdiction

2. **Run KYC verification (Query 3)**
   - Check identity (registration number exists)
   - Query external KYC provider (simulated in Query 3)
   - Screen against OFAC/EU/UN sanctions lists (AML)

3. **Generate KYC report**
   - :KYCVerified → Deal can proceed
   - :KYCPending → Escalate to compliance officer
   - :KYCRejected → Deal blocked

4. **Update party record**
   - `:hasKYCStatus` set to verification result
   - `:kycExpiryDate` set (typically 1 year)
   - `:amlScreeningResult` recorded

### 7.3 Contract Execution

1. **Draft contract with standard terms**
   - Create :Contract entity
   - Attach :ContractualTerm (pricing, payment, liability, confidentiality, termination)

2. **Validate compliance (Query 2)**
   - Verify all required term types present
   - Check regulatory requirements

3. **Execute signature workflow**
   - Both parties sign (collect e-signatures)
   - Update :SignatureStage

4. **Record audit trail**
   - Create :DealAuditEntry for signature
   - Timestamp, user, action, reason logged

### 7.4 Deal Closure

1. **Mark stage as Execution**
   - Performance obligations begin
   - Record :stageTransitionDate

2. **Monitor performance**
   - Payment terms (Net 30, Net 60) tracked
   - Delivery obligations monitored

3. **Final closure**
   - All terms satisfied → Mark :ClosureStage
   - Record actual close date
   - Generate deal performance report

---

## 8. Data Integration Points

### 8.1 PostgreSQL ↔ RDF Pipeline

**Forward (SQL → RDF):**

```
BusinessOS PostgreSQL
  ├── deals table
  ├── parties table
  ├── contracts table
  └── audit_log table
         ↓ (SPARQL CONSTRUCT)
    Oxigraph RDF Triplestore
      ├── Deal entities
      ├── Party entities
      └── Compliance reports
         ↓ (SPARQL SELECT/CONSTRUCT)
    Knowledge Graph Queries
      ├── "Find all deals with KYC-verified parties"
      ├── "Show compliance status by framework"
      └── "Generate deal portfolio report"
```

**Reverse (RDF → SQL):**

```
RDF Triplestore
  ├── Updated :dealStatus
  ├── Updated :currentStage
  └── New :DealAuditEntry
         ↓ (R2RML Mapping)
    BusinessOS PostgreSQL
      ├── deals (status updated)
      ├── deal_stages (new record)
      └── audit_log (new entry)
```

### 8.2 Ontology File Locations

| File | Purpose | Location |
|------|---------|----------|
| **fibo-deal-integration.ttl** | Core ontology (150+ statements) | `BusinessOS/ontologies/` |
| **fibo-deal-sparql-queries.rq** | 3 CONSTRUCT queries | `BusinessOS/ontologies/` |
| **fibo-deal-examples.ttl** | Example deals in RDF | `BusinessOS/ontologies/` |

### 8.3 RDF Data Locations

| Entity Type | URI Pattern | Example |
|------------|-------------|---------|
| Deal | `https://businessos.dev/id/deals/{deal_id}` | `...deals/d-acme-cloudtech-2026` |
| Party | `https://businessos.dev/id/parties/{party_id}` | `...parties/acme-corp` |
| Amount | `https://businessos.dev/id/amounts/{deal_id}-{type}` | `...amounts/d-001-principal` |
| Contract | `https://businessos.dev/id/contracts/{contract_id}` | `...contracts/c-001` |
| Term | `https://businessos.dev/id/terms/{deal_id}-{term_type}` | `...terms/d-001-pricing` |
| Audit | `https://businessos.dev/id/audits/{audit_id}` | `...audits/d-001-created` |

---

## 9. Implementation Checklist

### Phase 1: Ontology Setup
- [ ] Load `fibo-deal-integration.ttl` into Oxigraph triplestore
- [ ] Verify FIBO namespace imports resolve
- [ ] Run `weaver registry check` to validate schema conformance

### Phase 2: Data Integration
- [ ] Create R2RML mapping from PostgreSQL deals → RDF
- [ ] Build ETL pipeline (deals table → SPARQL INSERT)
- [ ] Test insertion of 3 example deals (Query 3: fibo-deal-examples.ttl)

### Phase 3: SPARQL Queries
- [ ] Test CONSTRUCT Query 1 (Deal creation)
- [ ] Test CONSTRUCT Query 2 (Compliance checking)
- [ ] Test CONSTRUCT Query 3 (KYC/AML verification)
- [ ] Verify all queries execute on test data without errors

### Phase 4: API Integration
- [ ] Create `/api/deals/fibo/{deal_id}` endpoint
  - Returns deal in RDF (Turtle, JSON-LD, RDF/XML)
- [ ] Create `/api/deals/compliance` endpoint
  - POST deal → runs Query 2 → returns compliance report
- [ ] Create `/api/parties/verify-kyc` endpoint
  - POST party → runs Query 3 → returns KYC report

### Phase 5: Testing & Validation
- [ ] Unit tests for SPARQL query execution
- [ ] Integration tests for SQL → RDF pipeline
- [ ] Compliance test: Deal can't move to Execution without KYC
- [ ] Audit test: All deal updates logged to audit_log + :DealAuditEntry

---

## 10. Advanced Topics

### 10.1 Deal Portfolio Analytics

Query across all deals in portfolio:

```sparql
SELECT ?deal_name ?total_value ?avg_stage_duration
WHERE {
  ?deal a :Deal ;
    :dealName ?deal_name ;
    :dealAmount ?amount .
  ?amount :amount ?amount_value .
}
GROUP BY ?deal
ORDER BY DESC(?amount_value)
```

### 10.2 Compliance Dashboard

Real-time compliance status by framework:

```sparql
SELECT ?framework ?compliant_count ?non_compliant_count
WHERE {
  ?deal :hasComplianceRequirement ?req .
  ?req :complianceFramework ?framework ;
       :complianceStatus ?status .
}
GROUP BY ?framework ?status
```

### 10.3 Deal Lifecycle Metrics

Average duration in each stage:

```sparql
SELECT ?stage (AVG(?duration) AS ?avg_days)
WHERE {
  ?deal :currentStage ?stage ;
        :stageDurationDays ?duration .
}
GROUP BY ?stage
ORDER BY DESC(?avg_days)
```

---

## 11. References

- **FIBO Specification:** https://spec.edmcouncil.org/fibo/
- **FIBO FND (Foundations):** https://spec.edmcouncil.org/fibo/ontology/FND/
- **FIBO FBC (Financial Business & Commerce):** https://spec.edmcouncil.org/fibo/ontology/FBC/
- **FIBO BE (Business Entities):** https://spec.edmcouncil.org/fibo/ontology/BE/
- **FIBO FEI (Financial Economics & Industry):** https://spec.edmcouncil.org/fibo/ontology/FEI/
- **W3C OWL 2:** https://www.w3.org/TR/owl2-overview/
- **W3C SPARQL 1.1:** https://www.w3.org/TR/sparql11-overview/
- **Oxigraph Triplestore:** https://oxigraph.org/
- **R2RML (Relational to RDF Mapping):** https://www.w3.org/TR/r2rml/

---

## 12. Troubleshooting

### Issue: Namespace Resolution Error

**Symptom:** SPARQL query fails with "undefined namespace fibo-fnd"

**Solution:**
1. Verify `fibo-deal-integration.ttl` loaded into triplestore
2. Check @prefix declarations match query
3. Ensure FIBO ontology imports resolve (may require network access)

### Issue: Deal Won't Insert into RDF

**Symptom:** SPARQL INSERT DATA returns error

**Solution:**
1. Verify all required properties have values (no null/empty)
2. Check xsd:datatype annotations match expected range
3. Use SPARQL CONSTRUCT with VALUES instead of INSERT DATA

### Issue: KYC Status Not Updating

**Symptom:** Query 3 returns :KYCPending but should return :KYCVerified

**Solution:**
1. Verify party has :businessRegistrationNumber (required)
2. Check :operatingJurisdiction is valid (US, EU, GB, CA, AU)
3. Run Query 3 manually to debug WHERE clause

---

**Version History:**
- v1.0.0 (2026-03-25) - Initial release with 3 CONSTRUCT queries and examples
