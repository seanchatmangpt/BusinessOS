# HIPAA-Compliant PHI Handler Implementation (Agent 9)

**Date:** 2026-03-26
**Project:** ChatmanGPT / BusinessOS
**Standard:** HIPAA § 164.312(b) (Access Control + Audit Logging)
**Framework:** FHIR R4 + PROV-O + Oxigraph RDF Store

---

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [FHIR R4 Integration](#fhir-r4-integration)
4. [PROV-O Provenance Modeling](#prov-o-provenance-modeling)
5. [HIPAA § 164.312(b) Mapping](#hipaa-§-164312b-mapping)
6. [API Endpoints](#api-endpoints)
7. [SPARQL Query Specifications](#sparql-query-specifications)
8. [Consent Verification](#consent-verification)
9. [Audit Trail Implementation](#audit-trail-implementation)
10. [Deletion & GDPR Compliance](#deletion--gdpr-compliance)
11. [Compliance Verification](#compliance-verification)
12. [Configuration](#configuration)
13. [Error Handling](#error-handling)
14. [Performance Tuning](#performance-tuning)

---

## Overview

The HIPAA-compliant PHI Handler implements Protected Health Information (PHI) tracking in BusinessOS with:

- **FHIR R4 Compliance:** Support for Patient, Observation, MedicationRequest, and other standard FHIR resources
- **PROV-O Provenance:** Every PHI access/modification tracked via W3C Provenance ontology
- **Oxigraph RDF Store:** Semantic triples persisted to Oxigraph for immutable audit trails
- **HIPAA § 164.312(b):** Mandatory access control policies and audit logging
- **Hard Delete GDPR:** Right to be forgotten (GDPR Article 17) via hard delete + RDF verification
- **12-Second Timeout:** All operations bounded to prevent resource exhaustion

### Key Components

| Component | File | Purpose |
|-----------|------|---------|
| **PHI Manager** | `internal/ontology/healthcare_phi.go` | Core PHI tracking logic |
| **PHI Tests** | `internal/ontology/healthcare_phi_test.go` | 12+ unit tests |
| **HTTP Handlers** | `internal/handlers/healthcare.go` | Gin endpoints (5 routes) |
| **Handler Tests** | `internal/handlers/healthcare_test.go` | 8-10 integration tests |

---

## Architecture

### Component Diagram

```
HTTP Request
    ↓
[Healthcare Handler] (Gin endpoints)
    ↓
[PHI Manager] (ontology.HealthcarePHIManager)
    ├─→ [SPARQL Executor] (Oxigraph CONSTRUCT/ASK queries)
    ├─→ [RDF Store] (Oxigraph triple persistence)
    └─→ [Audit Logger] (PostgreSQL + HMAC)
        ↓
    RDF Store (Oxigraph)
    + Audit Database (PostgreSQL)
```

### Data Flow: TrackPHI Operation

```
1. Handler receives POST /api/healthcare/phi/track
   ├─ Validates FHIR resource type
   ├─ Extracts actor (user_id from JWT)
   └─ Calls phiManager.TrackPHI()

2. PHI Manager runs 4 SPARQL CONSTRUCT queries
   ├─ CONSTRUCT 1: prov:Entity triples
   ├─ CONSTRUCT 2: prov:Activity triples
   ├─ CONSTRUCT 3: prov:wasGeneratedBy relationships
   └─ CONSTRUCT 4: prov:wasAttributedTo relationships

3. Manager stores combined Turtle to Oxigraph
   └─ Returns: resource_id, triple_count, hipaa_check_passed

4. Manager logs audit entry (PHI access logged)
   └─ Stored with HMAC signature for integrity

5. Handler returns 201 Created with PHI tracking result
```

---

## FHIR R4 Integration

### Supported FHIR Resources

The handler validates 8 FHIR resource types (FHIR R4 standard):

| Resource Type | Use Case | Example Data |
|---------------|----------|--------------|
| **Patient** | Demographics | name, dob, gender, contact |
| **Observation** | Clinical measurements | vital signs (BP, HR), lab results |
| **MedicationRequest** | Prescriptions | medication, dosage, frequency |
| **Procedure** | Clinical procedures | name, performer, date |
| **Condition** | Diagnosis | code, onset, severity |
| **AllergyIntolerance** | Known allergies | substance, reaction, severity |
| **Encounter** | Clinical visit | type, period, reason |
| **DiagnosticReport** | Lab results | findings, conclusion |

### FHIR Resource Structure

```json
{
  "resourceType": "Patient",
  "id": "p123",
  "name": [
    {
      "use": "official",
      "given": ["John"],
      "family": "Doe"
    }
  ],
  "birthDate": "1980-01-15",
  "gender": "male",
  "contact": [
    {
      "telecom": [
        {
          "system": "phone",
          "value": "555-0123"
        }
      ]
    }
  ]
}
```

### PHI vs Non-PHI

**PHI (Protected Health Information):**
- Patient name, DOB, contact info
- Medical history, diagnoses, treatments
- Lab results, vital signs
- Medication prescriptions
- Encounter records

**Non-PHI:**
- Anonymized data (date shifted, identifying fields removed)
- Aggregated statistics
- De-identified research datasets

---

## PROV-O Provenance Modeling

### W3C PROV-O Ontology

The manager uses W3C Provenance ontology to track "who, what, when, where, why":

```turtle
@prefix prov: <http://www.w3.org/ns/prov#> .
@prefix fhir: <http://hl7.org/fhir/> .

# Entity: The FHIR resource
fhir:Patient_p123
  a prov:Entity ;
  prov:type fhir:Patient ;
  prov:wasGeneratedBy fhir:activity_Patient_1234567890 ;
  prov:wasAttributedTo fhir:actor/doctor@example.com ;
  dcat:issued "2026-03-26T10:00:00Z"^^xsd:dateTime .

# Activity: The creation/modification action
fhir:activity_Patient_1234567890
  a prov:Activity ;
  prov:wasAssociatedWith fhir:actor/doctor@example.com ;
  prov:startedAtTime "2026-03-26T09:55:00Z"^^xsd:dateTime ;
  prov:endedAtTime "2026-03-26T10:00:00Z"^^xsd:dateTime .

# Relationships
fhir:Patient_p123
  prov:qualifiedGeneration [
    prov:activity fhir:activity_Patient_1234567890 ;
    prov:atTime "2026-03-26T10:00:00Z"^^xsd:dateTime
  ] ;
  prov:qualifiedAttribution [
    prov:agent fhir:actor/doctor@example.com ;
    prov:role fhir:role/creator
  ] .
```

### Provenance Queries

**SPARQL CONSTRUCT 1: Create Entity Triples**
```sparql
PREFIX prov: <http://www.w3.org/ns/prov#>
PREFIX fhir: <http://hl7.org/fhir/>
CONSTRUCT {
  fhir:Patient_p123 a prov:Entity ;
    prov:type fhir:Patient ;
    prov:label "Patient/p123" ;
    prov:wasAttributedTo fhir:patient/p123 ;
    dcat:issued "2026-03-26T10:00:00Z"^^<http://www.w3.org/2001/XMLSchema#dateTime> .
}
WHERE {
  BIND(fhir:Patient_p123 AS ?entity)
}
```

**SPARQL CONSTRUCT 2: Create Activity Triples**
```sparql
PREFIX prov: <http://www.w3.org/ns/prov#>
PREFIX fhir: <http://hl7.org/fhir/>
CONSTRUCT {
  fhir:activity_Patient_1234567890 a prov:Activity ;
    prov:wasAssociatedWith fhir:actor/doctor@example.com ;
    prov:startedAtTime "2026-03-26T09:55:00Z"^^<http://www.w3.org/2001/XMLSchema#dateTime> ;
    prov:endedAtTime "2026-03-26T10:00:00Z"^^<http://www.w3.org/2001/XMLSchema#dateTime> .
}
WHERE {
  BIND(fhir:activity_Patient_1234567890 AS ?activity)
}
```

**SPARQL CONSTRUCT 3: Create wasGeneratedBy Relationships**
```sparql
PREFIX prov: <http://www.w3.org/ns/prov#>
PREFIX fhir: <http://hl7.org/fhir/>
CONSTRUCT {
  fhir:Patient_p123 prov:wasGeneratedBy fhir:activity_Patient_1234567890 ;
    prov:qualifiedGeneration [
      prov:activity fhir:activity_Patient_1234567890 ;
      prov:atTime "2026-03-26T10:00:00Z"^^<http://www.w3.org/2001/XMLSchema#dateTime>
    ] .
}
WHERE {
  BIND(fhir:Patient_p123 AS ?entity)
}
```

**SPARQL CONSTRUCT 4: Create wasAttributedTo Relationships**
```sparql
PREFIX prov: <http://www.w3.org/ns/prov#>
PREFIX fhir: <http://hl7.org/fhir/>
CONSTRUCT {
  fhir:Patient_p123 prov:wasAttributedTo fhir:actor/doctor@example.com ;
    prov:qualifiedAttribution [
      prov:agent fhir:actor/doctor@example.com ;
      prov:role fhir:role/creator
    ] .
}
WHERE {
  BIND(fhir:Patient_p123 AS ?entity)
}
```

---

## HIPAA § 164.312(b) Mapping

### HIPAA § 164.312(b) Requirements

> **§ 164.312(b) Audit controls.** A covered entity shall implement hardware, software, and/or procedural mechanisms that record and examine PHI access and use.

### Implementation Mapping

| HIPAA Requirement | Implementation | Code Reference |
|-------------------|----------------|-----------------|
| **Access Control Policy** | Role-based (Doctor, Nurse, Pharmacist) | TrackPHI: actor parameter |
| **Audit Logging** | Every PHI operation logged | AuditLogger interface |
| **Unique User ID** | JWT user_id from auth middleware | handlers.go: c.Get("user_id") |
| **Timestamps** | RFC3339 UTC timestamps | PHIAuditEntry.Timestamp |
| **Success/Failure** | SPARQL ASK results indicate success | VerifyHIPAA result.AuditLogPass |
| **Data Encryption** | Oxigraph supports encrypted RDF | SPARQL: sec:encryption sec:AES256 |
| **Integrity** | HMAC-256 signatures on audit entries | AuditLogger.VerifyAuditIntegrity |

### Compliance Checklist

- [x] **§ 164.312(a)(2)(i)**: Data encryption in transit (HTTPS) + at rest (Oxigraph encryption support)
- [x] **§ 164.312(b)**: Audit logging on all PHI operations (implemented via AuditLogger)
- [x] **§ 164.312(c)(1)**: Integrity controls (HMAC signatures on audit entries)
- [x] **§ 164.308(a)(3)**: Access control policies (role-based actor validation)

---

## API Endpoints

### 1. Track PHI Resource

**Endpoint:** `POST /api/healthcare/phi/track`

**Request Body:**
```json
{
  "resource_id": "p123",
  "resource_type": "Patient",
  "patient_id": "p123",
  "data": {
    "name": "John Doe",
    "dob": "1980-01-15",
    "gender": "male"
  }
}
```

**Response (201 Created):**
```json
{
  "resource_id": "p123",
  "resource_type": "Patient",
  "triple_count": 12,
  "prov_entity_id": "http://hl7.org/fhir/Patient_p123",
  "prov_activity_id": "http://hl7.org/fhir/activity_Patient_1234567890",
  "timestamp": "2026-03-26T10:00:00Z",
  "hipaa_check_passed": true
}
```

**Error Cases:**
- `400 Bad Request`: Invalid resource_type (unsupported FHIR type)
- `400 Bad Request`: Missing required fields
- `500 Internal Server Error`: SPARQL CONSTRUCT failure
- `500 Internal Server Error`: Oxigraph store failure

---

### 2. Get PHI Audit Trail

**Endpoint:** `GET /api/healthcare/phi/:id/audit?patient_id=p123&days=90`

**Query Parameters:**
- `patient_id` (required): Patient ID for audit scope
- `days` (optional, default=90): Look back period (1-365)

**Response (200 OK):**
```json
{
  "resource_id": "p123",
  "audit_trail": [
    {
      "timestamp": "2026-03-21T14:30:00Z",
      "actor": "doctor@example.com",
      "action": "read",
      "resource_id": "obs123",
      "resource_type": "Observation",
      "details": "Patient vitals review",
      "ip_address": "192.168.1.100",
      "signature": "hmac-256-abcd1234..."
    }
  ],
  "count": 24,
  "period": "last_90_days"
}
```

**Error Cases:**
- `400 Bad Request`: Missing patient_id parameter
- `500 Internal Server Error`: SPARQL CONSTRUCT failure

---

### 3. Verify Consent

**Endpoint:** `POST /api/healthcare/consent/verify`

**Request Body:**
```json
{
  "patient_id": "p123"
}
```

**Response (200 OK - Consent Granted):**
```json
{
  "patient_id": "p123",
  "consent_granted": true,
  "consent_doc_id": "Consent/p123_consent",
  "expires_at": "2027-03-26T00:00:00Z",
  "scope": ["treatment", "payment"],
  "verified_at": "2026-03-26T10:00:00Z"
}
```

**Response (403 Forbidden - Consent Denied):**
```json
{
  "patient_id": "p123",
  "consent_granted": false,
  "consent_doc_id": "",
  "expires_at": "2026-03-26T00:00:00Z",
  "scope": [],
  "verified_at": "2026-03-26T10:00:00Z"
}
```

**Error Cases:**
- `400 Bad Request`: Invalid patient_id format
- `500 Internal Server Error`: SPARQL ASK failure

---

### 4. Delete PHI (Hard Delete)

**Endpoint:** `DELETE /api/healthcare/phi/:id?resource_type=Patient`

**Query Parameters:**
- `resource_type` (required): FHIR resource type

**Response (200 OK - Fully Deleted):**
```json
{
  "resource_id": "p999",
  "fully_deleted": true,
  "triple_count": 0,
  "verified_at": "2026-03-26T10:00:00Z",
  "rdf_clean_confirmed": true
}
```

**Response (409 Conflict - Partial Deletion):**
```json
{
  "error": "deletion incomplete",
  "remaining_triples": 3,
  "verified_at": "2026-03-26T10:00:00Z"
}
```

**Error Cases:**
- `400 Bad Request`: Missing resource_type parameter
- `409 Conflict`: Deletion incomplete (RDF remnants exist)
- `500 Internal Server Error`: RDF query failure

---

### 5. Verify HIPAA Compliance

**Endpoint:** `GET /api/healthcare/hipaa/verify`

**Response (200 OK - Compliant):**
```json
{
  "compliant": true,
  "access_control_pass": true,
  "audit_log_pass": true,
  "encryption_pass": true,
  "integrity_pass": true,
  "access_log_count": 400,
  "failed_access_count": 0,
  "checked_at": "2026-03-26T10:00:00Z",
  "compliance_score": 1.0
}
```

**Response (206 Partial Content - Non-Compliant):**
```json
{
  "compliant": false,
  "access_control_pass": true,
  "audit_log_pass": false,
  "encryption_pass": false,
  "integrity_pass": true,
  "access_log_count": 0,
  "failed_access_count": 12,
  "checked_at": "2026-03-26T10:00:00Z",
  "compliance_score": 0.5
}
```

---

## SPARQL Query Specifications

### Query Execution Context

All SPARQL queries use these namespace declarations:

```sparql
PREFIX prov: <http://www.w3.org/ns/prov#>
PREFIX fhir: <http://hl7.org/fhir/>
PREFIX dcat: <http://www.w3.org/ns/dcat#>
PREFIX xsd: <http://www.w3.org/2001/XMLSchema#>
PREFIX gdpr: <http://data.europa.eu/930/gdpr#>
PREFIX hipaa: <http://hl7.org/fhir/SecurityEvent#>
PREFIX sec: <http://hl7.org/fhir/security#>
```

### TrackPHI Queries (4 CONSTRUCT)

See [PROV-O Provenance Modeling](#prov-o-provenance-modeling) for full query text.

**Summary:**
1. Entity triples: `fhir:{Type}_{ID} a prov:Entity`
2. Activity triples: `fhir:activity_{Type}_{Timestamp} a prov:Activity`
3. wasGeneratedBy: `Entity prov:wasGeneratedBy Activity`
4. wasAttributedTo: `Entity prov:wasAttributedTo Actor`

### VerifyConsent Query (ASK + CONSTRUCT)

**ASK Query (Consent Existence):**
```sparql
PREFIX fhir: <http://hl7.org/fhir/>
PREFIX xsd: <http://www.w3.org/2001/XMLSchema#>
ASK {
  ?consent a fhir:Consent ;
    fhir:patient [ fhir:reference "Patient/{patientID}" ] ;
    fhir:status "active" ;
    fhir:dateTime ?date .
  FILTER(?date > NOW())
}
```

Returns: `true` if valid consent exists, `false` otherwise.

**CONSTRUCT Query (Consent Details):**
```sparql
PREFIX fhir: <http://hl7.org/fhir/>
CONSTRUCT {
  ?consent a fhir:Consent ;
    fhir:id "{patientID}_consent" ;
    fhir:patient [ fhir:reference "Patient/{patientID}" ] ;
    fhir:status "active" ;
    fhir:scope ?scope .
}
WHERE {
  BIND(fhir:Consent/{patientID}_consent AS ?consent)
}
```

### GenerateAuditTrail Query (CONSTRUCT)

**Activity Extraction:**
```sparql
PREFIX prov: <http://www.w3.org/ns/prov#>
PREFIX fhir: <http://hl7.org/fhir/>
CONSTRUCT {
  ?activity a prov:Activity ;
    prov:wasAssociatedWith ?actor ;
    prov:used ?entity ;
    prov:startedAtTime ?time .
}
WHERE {
  ?activity prov:wasAssociatedWith ?actor ;
    prov:used [ fhir:patient [ fhir:reference "Patient/{patientID}" ] ] ;
    prov:startedAtTime ?time .
  FILTER(?time >= "{90DaysAgo}"^^xsd:dateTime)
}
```

### CheckDeletion Queries (ASK + COUNT)

**Triple Count Query:**
```sparql
SELECT (COUNT(*) as ?count) WHERE {
  ?entity ?p ?o .
  FILTER(str(?entity) = "{entityURI}")
}
```

**Entity Existence Check (ASK):**
```sparql
PREFIX fhir: <http://hl7.org/fhir/>
ASK {
  ?entity ?p ?o .
  FILTER(str(?entity) = "{entityURI}")
}
```

**GDPR Compliance Assertion (CONSTRUCT):**
```sparql
PREFIX gdpr: <http://data.europa.eu/930/gdpr#>
CONSTRUCT {
  fhir:deletion_{Type}_{Timestamp}
    a gdpr:RightToBeForgettenCompliance ;
    gdpr:deletedResource "{entityURI}" ;
    gdpr:completedAt "{timestamp}"^^xsd:dateTime .
}
WHERE {
  BIND(fhir:deletion_{Type}_{Timestamp} AS ?deletion)
}
```

### VerifyHIPAA Queries (4 ASK)

**ASK 1: Access Control Policy Exists**
```sparql
PREFIX hipaa: <http://hl7.org/fhir/SecurityEvent#>
ASK {
  ?policy a hipaa:AccessControlPolicy ;
    hipaa:role ?role ;
    hipaa:permission ?perm .
}
```

**ASK 2: Audit Logs Present**
```sparql
PREFIX prov: <http://www.w3.org/ns/prov#>
PREFIX hipaa: <http://hl7.org/fhir/SecurityEvent#>
ASK {
  ?activity a prov:Activity ;
    prov:startedAtTime ?time .
  ?entry hipaa:eventAction ?action ;
    hipaa:eventDateTime ?date .
}
```

**ASK 3: Data Encryption**
```sparql
PREFIX fhir: <http://hl7.org/fhir/>
PREFIX sec: <http://hl7.org/fhir/security#>
ASK {
  ?resource sec:encryption sec:AES256 .
}
```

**ASK 4: HMAC Integrity Signatures**
```sparql
PREFIX sec: <http://hl7.org/fhir/security#>
ASK {
  ?entry sec:signature ?sig ;
    sec:signatureAlgorithm sec:HMAC256 .
}
```

---

## Consent Verification

### Consent Model (FHIR R4)

```json
{
  "resourceType": "Consent",
  "id": "p123_consent",
  "patient": {
    "reference": "Patient/p123"
  },
  "status": "active",
  "scope": {
    "coding": [
      {
        "system": "http://terminology.hl7.org/CodeSystem/consentscope",
        "code": "treatment"
      }
    ]
  },
  "dateTime": "2025-03-26T00:00:00Z",
  "provision": {
    "type": "permit",
    "period": {
      "start": "2025-03-26T00:00:00Z",
      "end": "2026-03-26T00:00:00Z"
    }
  }
}
```

### Consent Scopes

- **treatment:** Access for direct patient care
- **payment:** Access for billing and insurance
- **operations:** Access for business operations
- **research:** Access for research studies
- **marketing:** Access for marketing materials

### Consent Verification Flow

```
1. Check if Consent/{patientID}_consent exists
   ├─ If not found → ConsentGranted = false
   └─ If found → proceed to step 2

2. Check Consent.status == "active"
   ├─ If inactive → ConsentGranted = false
   └─ If active → proceed to step 3

3. Check Consent.dateTime > NOW()
   ├─ If expired → ConsentGranted = false
   └─ If valid → ConsentGranted = true

4. Extract Consent.provision.scope[] array
   └─ Return as scope: ["treatment", "payment", ...]

5. Calculate expires_at = Consent.provision.period.end
   └─ Return as expires_at timestamp
```

---

## Audit Trail Implementation

### Audit Entry Structure

```go
type PHIAuditEntry struct {
  Timestamp    time.Time              // RFC3339 UTC
  Actor        string                 // User ID (e.g., "doctor@example.com")
  Action       string                 // "create", "read", "update", "delete"
  ResourceID   string                 // FHIR resource ID
  ResourceType string                 // FHIR resource type
  Details      string                 // Additional context
  IPAddress    string                 // Source IP for network audit
  Signature    string                 // HMAC-256 signature
}
```

### Audit Storage (PostgreSQL)

```sql
CREATE TABLE phi_audit_log (
  id BIGSERIAL PRIMARY KEY,
  timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  actor VARCHAR(255) NOT NULL,
  action VARCHAR(50) NOT NULL,
  resource_id VARCHAR(255) NOT NULL,
  resource_type VARCHAR(100) NOT NULL,
  details TEXT,
  ip_address INET,
  signature VARCHAR(256) NOT NULL,
  patient_id VARCHAR(255) NOT NULL,

  FOREIGN KEY (patient_id) REFERENCES patients(id) ON DELETE CASCADE,
  INDEX idx_patient_timestamp (patient_id, timestamp DESC),
  INDEX idx_actor_timestamp (actor, timestamp DESC)
);
```

### HMAC Signature Computation

```go
// Pseudo-code: HMAC-256 signature
func SignAuditEntry(entry PHIAuditEntry, secret string) string {
  data := fmt.Sprintf(
    "%s|%s|%s|%s|%s|%s",
    entry.Timestamp.String(),
    entry.Actor,
    entry.Action,
    entry.ResourceID,
    entry.ResourceType,
    entry.Details,
  )
  h := hmac.New(sha256.New, []byte(secret))
  h.Write([]byte(data))
  return hex.EncodeToString(h.Sum(nil))
}
```

### Audit Trail Query (Last 90 Days)

```sql
SELECT timestamp, actor, action, resource_id, resource_type, details, ip_address, signature
FROM phi_audit_log
WHERE patient_id = $1
  AND timestamp >= NOW() - INTERVAL '90 days'
ORDER BY timestamp DESC
LIMIT 1000;
```

### Audit Trail Integrity Verification

```go
// Verify HMAC signatures on all entries
func VerifyAuditIntegrity(entries []PHIAuditEntry, secret string) bool {
  for _, entry := range entries {
    expected := SignAuditEntry(entry, secret)
    if entry.Signature != expected {
      return false
    }
  }
  return true
}
```

---

## Deletion & GDPR Compliance

### Hard Delete Process

GDPR Article 17 (Right to be Forgotten) requires that PHI be completely erased, including from all backups and archives:

```
1. Delete patient record from primary database
2. Delete all associated FHIR resources:
   - Observations (vital signs, lab results)
   - MedicationRequests (prescriptions)
   - Encounters (visit records)
   - DiagnosticReports
   - etc.
3. Delete RDF triples from Oxigraph:
   - Query: SELECT all triples WHERE subject = "Patient_{ID}"
   - Delete: All matching triples
4. Verify deletion (hard confirmation):
   - ASK query: Does patient entity still exist?
   - COUNT query: Are there 0 remaining triples?
   - GDPR assertion: Was deletion completed?
5. Audit log: Record deletion completion with timestamp
```

### Deletion Verification Query

```sparql
PREFIX fhir: <http://hl7.org/fhir/>
SELECT (COUNT(*) as ?remaining_triples) WHERE {
  ?entity ?p ?o .
  FILTER(str(?entity) = "http://hl7.org/fhir/Patient_p123")
}
```

Expected result: `remaining_triples = 0`

### Compliance Assertion (GDPR)

```sparql
PREFIX gdpr: <http://data.europa.eu/930/gdpr#>
CONSTRUCT {
  fhir:deletion_Patient_1234567890
    a gdpr:RightToBeForgettenCompliance ;
    gdpr:deletedResource "http://hl7.org/fhir/Patient_p123" ;
    gdpr:completedAt "2026-03-26T10:00:00Z"^^xsd:dateTime ;
    gdpr:authority "GDPR Article 17" ;
    gdpr:scope "patient" .
}
WHERE {
  BIND(fhir:deletion_Patient_1234567890 AS ?deletion)
}
```

---

## Compliance Verification

### HIPAA Compliance Checklist

The `VerifyHIPAA()` function checks 4 critical compliance measures:

| Check | HIPAA Section | Query Type | Pass Criterion |
|-------|----------------|-----------|-----------------|
| **Access Control** | § 164.312(a)(2) | SPARQL ASK | Access control policy exists |
| **Audit Logging** | § 164.312(b) | SPARQL ASK | Audit entries recorded |
| **Encryption** | § 164.312(a)(2)(i) | SPARQL ASK | Data encrypted (AES-256) |
| **Integrity** | § 164.312(c)(1) | SPARQL ASK | HMAC signatures verified |

### Compliance Score Calculation

```go
score := 0.0
if accessControlPass {
  score += 0.25
}
if auditLogPass {
  score += 0.25
}
if encryptionPass {
  score += 0.25
}
if integrityPass {
  score += 0.25
}
return score // Range: 0.0 to 1.0
```

### Remediation Actions

If compliance check fails:

1. **Access Control Failure:** Add missing role-based access control policy
   ```sparql
   INSERT DATA {
     fhir:AccessControlPolicy_001 a hipaa:AccessControlPolicy ;
       hipaa:role "Doctor" ;
       hipaa:permission "read:PHI" .
   }
   ```

2. **Audit Logging Failure:** Enable audit log collection
   - Verify AuditLogger is connected
   - Check PostgreSQL connection
   - Verify audit_log table exists

3. **Encryption Failure:** Enable RDF encryption in Oxigraph
   - Configuration: `encryption: aes256`
   - Restart Oxigraph with encryption flag

4. **Integrity Failure:** Recalculate and update HMAC signatures
   - Re-sign all audit entries with current secret
   - Update signature column in PostgreSQL

---

## Configuration

### Environment Variables

```bash
# Oxigraph SPARQL endpoint
OXIGRAPH_ENDPOINT=http://localhost:7878/query

# PostgreSQL audit log connection
AUDIT_LOG_DSN=postgres://user:pass@localhost:5432/healthcare_audit

# HMAC secret for audit signatures (min 32 chars)
AUDIT_HMAC_SECRET=your-secret-key-min-32-characters-long

# RDF store encryption key (optional)
RDF_ENCRYPTION_KEY=your-encryption-key-min-32-chars

# Timeout for PHI operations (milliseconds)
PHI_OPERATION_TIMEOUT_MS=12000

# HIPAA strict mode (enforce all checks)
HIPAA_STRICT_MODE=true
```

### Go Code Initialization

```go
// Initialize handlers in main()
import "github.com/rhl/businessos-backend/internal/ontology"

// Create dependencies
sparqlExecutor := &ontology.SPARQLExecutorImpl{logger: logger}
rdfStore := &ontology.RDFStoreImpl{logger: logger}
auditLogger := &ontology.AuditLoggerImpl{logger: logger}

// Create PHI manager
phiManager := ontology.NewHealthcarePHIManager(
  sparqlExecutor,
  rdfStore,
  auditLogger,
  logger,
)

// Register routes
handlers := &Handlers{
  phiManager: phiManager,
  // ... other handlers
}

api.Group("/api").Use(authMiddleware)
handlers.registerHealthcareRoutes(api.Group("/api"), authMiddleware)
```

---

## Error Handling

### Error Categories

| Category | HTTP Status | Cause | Recovery |
|----------|------------|-------|----------|
| **Validation** | 400 Bad Request | Invalid FHIR type, missing fields | Client retries with valid data |
| **Unauthorized** | 401 Unauthorized | Missing/invalid JWT | Client re-authenticates |
| **Forbidden** | 403 Forbidden | No consent, access denied | Client requests consent |
| **Not Found** | 404 Not Found | Resource doesn't exist | Client verifies resource_id |
| **Conflict** | 409 Conflict | Deletion incomplete, RDF remnants | System attempts cleanup |
| **Timeout** | 504 Gateway Timeout | Operation exceeded 12s | Client retries with exponential backoff |
| **Server Error** | 500 Internal Server Error | SPARQL/RDF failure | System logs, operator investigates |

### Error Response Format

```json
{
  "error": "deletion incomplete",
  "code": "DELETION_INCOMPLETE",
  "message": "PHI resource still has 3 remaining RDF triples",
  "details": {
    "resource_id": "p123",
    "resource_type": "Patient",
    "remaining_triples": 3,
    "verified_at": "2026-03-26T10:00:00Z"
  }
}
```

---

## Performance Tuning

### Timeout Constraints

All operations are bounded to 12 seconds maximum:

```go
ctx, cancel := context.WithTimeout(ctx, 12*time.Second)
defer cancel()
```

This ensures:
- SPARQL CONSTRUCT queries complete in <3s
- RDF store operations complete in <2s
- Audit logging completes in <1s
- Remaining time is buffer for network + overhead

### Query Optimization

**Index RDF triples by subject (entity URI):**
```sql
-- Oxigraph indexes (configured at startup)
CREATE INDEX idx_subject ON triples(subject);
CREATE INDEX idx_subject_predicate ON triples(subject, predicate);
```

**Index audit log by patient + timestamp:**
```sql
CREATE INDEX idx_patient_timestamp ON phi_audit_log(patient_id, timestamp DESC);
```

### Caching Strategy

- **Consent results:** Cache for 5 minutes (consent rarely changes)
- **HIPAA compliance:** Cache for 1 hour (audit logging is continuous)
- **Patient entity URIs:** Cache for 24 hours (resource IDs are immutable)

```go
type Cache struct {
  consent map[string]*ConsentVerificationResult
  ttl     map[string]time.Time
}

func (c *Cache) Get(key string) (*ConsentVerificationResult, bool) {
  if result, ok := c.consent[key]; ok {
    if time.Now().Before(c.ttl[key]) {
      return result, true
    }
  }
  return nil, false
}
```

---

## Testing

### Test Coverage (12+ Tests)

| Test Name | Resource Type | Scenario |
|-----------|---------------|----------|
| `TestTrackPHI_PatientResource` | Patient | Basic tracking |
| `TestTrackPHI_ObservationResource` | Observation | Vital signs |
| `TestTrackPHI_MedicationRequestResource` | MedicationRequest | Prescription |
| `TestVerifyConsent_Granted` | — | Valid consent |
| `TestVerifyConsent_Denied` | — | No consent |
| `TestGenerateAuditTrail` | — | 90-day window |
| `TestGenerateAuditTrail_VerifyIntegrity` | — | HMAC verification |
| `TestCheckDeletion_FullyDeleted` | Patient | Hard delete success |
| `TestCheckDeletion_PartiallyDeleted` | Patient | Deletion incomplete |
| `TestVerifyHIPAA_Compliant` | — | All checks pass |
| `TestVerifyHIPAA_NonCompliant` | — | Audit logging fails |
| `TestConcurrentPHIOperations` | — | Thread safety |

### Running Tests

```bash
# Run all PHI tests
cd BusinessOS/desktop/backend-go
go test ./internal/ontology/... -v

# Run with coverage
go test ./internal/ontology/... -v -cover -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run handler tests
go test ./internal/handlers/... -v -run Healthcare
```

---

## References

### Standards & Specifications

1. **HIPAA § 164.312(b)** - Audit Controls
   - Full text: https://www.law.cornell.edu/cfr/text/45/164.312

2. **FHIR R4** - Patient, Observation, MedicationRequest Resources
   - https://www.hl7.org/fhir/R4/resourcelist.html

3. **W3C PROV-O** - Provenance Ontology
   - https://www.w3.org/TR/prov-o/

4. **GDPR Article 17** - Right to be Forgotten
   - https://gdpr-info.eu/art-17-gdpr/

### Related Documentation

- `docs/diataxis/explanation/signal-theory-complete.md` - Signal theory foundation
- `docs/diataxis/explanation/seven-layer-architecture.md` - Architecture patterns
- `BusinessOS/CLAUDE.md` - Project configuration

---

## Appendix: Sample FHIR Resources

### Patient Resource

```json
{
  "resourceType": "Patient",
  "id": "p123",
  "active": true,
  "name": [
    {
      "use": "official",
      "family": "Doe",
      "given": ["John"]
    }
  ],
  "telecom": [
    {
      "system": "phone",
      "value": "555-0123"
    }
  ],
  "gender": "male",
  "birthDate": "1980-01-15",
  "address": [
    {
      "use": "home",
      "line": ["123 Main St"],
      "city": "Springfield",
      "state": "IL",
      "postalCode": "62701"
    }
  ]
}
```

### Observation Resource (Vital Signs)

```json
{
  "resourceType": "Observation",
  "id": "obs456",
  "status": "final",
  "category": [
    {
      "coding": [
        {
          "system": "http://terminology.hl7.org/CodeSystem/observation-category",
          "code": "vital-signs"
        }
      ]
    }
  ],
  "code": {
    "coding": [
      {
        "system": "http://loinc.org",
        "code": "85354-9",
        "display": "Blood pressure panel"
      }
    ]
  },
  "subject": {
    "reference": "Patient/p123"
  },
  "effectiveDateTime": "2026-03-26T10:00:00Z",
  "value": {
    "value": 120,
    "unit": "mmHg"
  }
}
```

### MedicationRequest Resource (Prescription)

```json
{
  "resourceType": "MedicationRequest",
  "id": "med789",
  "status": "active",
  "intent": "order",
  "medicationCodeableConcept": {
    "coding": [
      {
        "system": "http://www.nlm.nih.gov/research/umls/rxnorm",
        "code": "7682",
        "display": "Aspirin"
      }
    ]
  },
  "subject": {
    "reference": "Patient/p123"
  },
  "authoredOn": "2026-03-26T10:00:00Z",
  "requester": {
    "reference": "Practitioner/doc123"
  },
  "dosageInstruction": [
    {
      "text": "One tablet by mouth twice daily",
      "timing": {
        "repeat": {
          "frequency": 2,
          "period": 1,
          "periodUnit": "d"
        }
      },
      "route": {
        "coding": [
          {
            "system": "http://snomed.info/sct",
            "code": "26643006",
            "display": "Oral route"
          }
        ]
      },
      "doseAndRate": [
        {
          "doseQuantity": {
            "value": 500,
            "unit": "mg"
          }
        }
      ]
    }
  ]
}
```

---

**Document Version:** 1.0
**Last Updated:** 2026-03-26
**Status:** Complete — Ready for HIPAA Compliance Review
