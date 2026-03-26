# Healthcare Ontology — Quick Reference Guide

## Commands at a Glance

### 1. Initialize Healthcare Workspace
```bash
bos healthcare init --organization "Your Hospital" [--fhir-version r4|r5]
```
**Returns:** Workspace path, ontology version, compliance frameworks
**Frameworks:** HIPAA, GDPR, HITECH

### 2. Track PHI Lineage
```bash
bos healthcare phi track \
  --patient-id pat-001 \
  [--phi-elements name,ssn,diagnosis,medication] \
  [--storage-location postgres|s3|vault]
```
**Returns:** Tracking ID, element count, lineage depth, audit entries
**Standard Elements:** name, ssn, diagnosis, medication, lab_results, procedure

### 3. Check Patient Consent
```bash
bos healthcare consent check \
  --patient-id pat-001 \
  [--resource Observation|MedicationRequest|DiagnosticReport] \
  [--access-type read|write|delete]
```
**Returns:** Consent validity, constraints, expiration, ODRL status

**Constraint Matrix:**
| Access | Constraints |
|--------|-------------|
| read | purpose_limited, anonymization |
| write | audit_required, patient_notification |
| delete | legal_hold, encryption_required |

### 4. Verify HIPAA Compliance
```bash
bos healthcare hipaa verify \
  --organization "Your Hospital" \
  [--audit-depth 1000]
```
**Returns:** Compliance status, findings, 6-year retention, audit entry count

---

## SPARQL Queries

### Query 1: PHI Lineage (PROV-O)
**Purpose:** Show full data provenance chain
**File:** `queries/healthcare_sparql.rq` (lines 1-80)
**Example:** WHO accessed PATIENT_ID's SSN, WHEN, and HOW

### Query 2: Consent Enforcement (ODRL)
**Purpose:** Verify access against patient permissions
**File:** `queries/healthcare_sparql.rq` (lines 82-160)
**Example:** Does ACTOR have PERMISSION to READ OBSERVATION for PATIENT_ID?

### Query 3: Audit Trail (6-Year Retention)
**Purpose:** Query logs for compliance verification
**File:** `queries/healthcare_sparql.rq` (lines 162-230)
**Example:** List all access to PATIENT_ID's PHI in past 6 years with timestamps

### Query 4: Privacy Verification
**Purpose:** Check privacy controls per patient
**File:** `queries/healthcare_sparql.rq` (lines 232-310)
**Example:** Is data minimization enabled? Are unauthorized accesses detected?

---

## Data Model

### PHI Tracking (PROV-O Ontology)
```
phi:ProtectedHealthInformation
  ├── phi:patient_id
  ├── phi:element_type (name|ssn|diagnosis|medication|lab_results|procedure)
  ├── phi:created_at (ISO8601)
  ├── phi:data_location (postgres|s3|vault)
  ├── prov:wasGeneratedBy (Activity)
  ├── prov:wasDerivedFrom (source PHI)
  ├── prov:wasAttributedTo (Actor)
  └── prov:hadPrimarySource (authoritative source)

prov:Activity
  ├── prov:startedAtTime
  ├── prov:endedAtTime
  ├── prov:wasAssociatedWith (Actor)
  ├── phi:transformation_type
  └── phi:audit_timestamp
```

### Consent Model (ODRL Ontology)
```
odrl:Policy
  ├── odrl:uid (consent_id)
  ├── odrl:target (resource: Observation, MedicationRequest, etc.)
  ├── odrl:permission
  │   ├── odrl:action (read|write|delete)
  │   ├── odrl:assigner (patient)
  │   ├── odrl:assignee (healthcare provider)
  │   ├── odrl:constraint (purpose_limited, anonymization, etc.)
  │   └── odrl:duty (notification, audit)
  ├── odrl:prohibition (explicitly denied actions)
  ├── odrl:obligation (required actions)
  ├── dct:issued
  ├── dct:valid
  ├── dct:expires
  └── healthcare:consent_status (active|revoked|expired)
```

### Audit Entry Model
```
healthcare:AuditEntry
  ├── healthcare:audit_id (unique ID)
  ├── healthcare:timestamp (microsecond precision)
  ├── healthcare:event_type (login|logout|access|modification|export|delete)
  ├── healthcare:actor (user/system)
  ├── healthcare:actor_ip
  ├── healthcare:patient_id
  ├── healthcare:resource_type (FHIR: Patient|Observation|MedicationRequest)
  ├── healthcare:action_taken (read|write|delete|view)
  ├── healthcare:access_granted (boolean)
  ├── healthcare:encryption_in_transit (boolean)
  ├── healthcare:encryption_at_rest (boolean)
  ├── healthcare:integrity_verified (hash/signature)
  ├── healthcare:retention_until (NOW + 6 years)
  └── healthcare:archival_status (active|archived|deleted)
```

---

## FHIR R4 Resources Included

### Patient
- **ID:** pat-001
- **Identifiers:** MRN-2026-001, SSN 123-45-6789
- **Name:** Alice Marie Chatman
- **Contact:** Email, phone, address
- **Demographics:** DOB 1985-03-15, female, married
- **Emergency Contact:** Robert Chatman

### Observation (Vital Signs)
- **ID:** obs-001
- **Type:** Systolic Blood Pressure (LOINC 8480-6)
- **Subject:** Alice Chatman (pat-001)
- **Performer:** Dr. Sarah Johnson
- **Value:** 130 mmHg (High)
- **Reference Range:** 90-120 mmHg
- **Effective:** 2026-03-25T14:30:00Z

### MedicationRequest
- **ID:** medrq-001
- **Medication:** Lisinopril 10 MG (RxNorm 617312)
- **Subject:** Alice Chatman (pat-001)
- **Indication:** Hypertension (SNOMED 38341003)
- **Dosage:** Once daily
- **Supply:** 30-day supply
- **Requester:** Dr. Sarah Johnson

**File:** `examples/fhir_patients.json`

---

## Compliance Checklist

### HIPAA (65 CFR Parts 160, 162, 164)
- [ ] Audit trail with microsecond-precision timestamps
- [ ] 6-year retention of audit logs
- [ ] User identification and authentication on all accesses
- [ ] Access request and response content captured
- [ ] Modification tracking (who, what, when)
- [ ] Encryption in transit (TLS 1.2+)
- [ ] Encryption at rest (AES-256)
- [ ] Integrity verification (HMAC/signature)
- [ ] Access control (role-based or attribute-based)
- [ ] Breach notification procedures

### GDPR (EU Regulation 2016/679)
- [ ] Consent obtained and verified (ODRL)
- [ ] Purpose limitation enforced
- [ ] Data minimization practiced
- [ ] Storage limitation (retention policy) enforced
- [ ] Integrity and confidentiality maintained
- [ ] Accountability documented (PROV-O)
- [ ] Data subject rights implemented (access, deletion, portability)

### HITECH Act (42 USC § 17921-17954)
- [ ] Genetic information protected separately
- [ ] Enforcement of HIPAA with civil penalties
- [ ] Breach notification compliance

---

## Testing

### Run All Tests
```bash
cd /Users/sac/chatmangpt/BusinessOS/bos
cargo test healthcare_ontology_test
```

### Run Single Test
```bash
cargo test test_healthcare_init_creates_workspace -- --nocapture
```

### Test Coverage (12 Functions)

| Test | Purpose |
|------|---------|
| test_healthcare_init_creates_workspace | Workspace initialization |
| test_phi_track_creates_lineage | PHI tracking with PROV-O |
| test_phi_track_default_elements | Default PHI elements |
| test_consent_check_read_access | Read access constraints |
| test_consent_check_write_access | Write access constraints |
| test_consent_check_delete_access | Delete access constraints |
| test_hipaa_verify_compliance | Adequate audit depth |
| test_hipaa_verify_insufficient_audit | Insufficient audit detection |
| test_phi_track_includes_audit_entries | Lineage depth verification |
| test_healthcare_workflow_init_then_track | Init → Track workflow |
| test_consent_check_default_resource | Default resource type |
| test_hipaa_verify_default_audit_depth | Default audit depth |

---

## Integration Points

### With Oxigraph (RDF Store)
```
SPARQL CONSTRUCT queries → Oxigraph at http://localhost:8089
├── PROV-O triples (PHI lineage)
├── ODRL triples (consent policies)
├── FHIR triples (patient/observation/medication)
└── Healthcare custom triples
```

### With PostgreSQL (Audit Trail)
```
healthcare.audit_entries table (6-year retention)
├── audit_id (UUID)
├── timestamp (TIMESTAMP WITH TIMEZONE, microsecond)
├── event_type (VARCHAR)
├── actor_id (VARCHAR)
├── patient_id (VARCHAR, encrypted)
├── resource_type (VARCHAR)
├── action_taken (VARCHAR)
├── access_granted (BOOLEAN)
└── retention_until (TIMESTAMP)
```

### With Jaeger (OpenTelemetry)
```
Span: service=businessos, span_name=healthcare.{init,phi_track,consent_check,hipaa_verify}
├── Attributes: patient_id, resource_type, organization
├── Status: ok|error
└── Duration: latency_ms
```

---

## Error Handling

### Insufficient Audit Depth
```
Error: Insufficient audit trail depth for compliance
Threshold: 100 entries
Remedy: Ensure audit logging is enabled and audit_depth ≥ 100
```

### Consent Expired
```
Error: Consent expired on 2026-12-30
Remedy: Patient must grant new consent before access
```

### PHI Encryption Required
```
Error: PHI stored without encryption at rest
Remedy: Enable AES-256 encryption for storage location
```

---

## Performance Notes

- **PHI Tracking:** O(n) where n = PHI elements
- **Consent Check:** O(1) policy lookup + O(m) constraint evaluation (m = constraints)
- **HIPAA Verify:** O(k) where k = audit entries (linear scan, optimizable with index on retention_until)
- **Privacy Verification:** O(p) where p = privacy settings (small constant)

**Recommended Indexes:**
```sql
CREATE INDEX idx_audit_patient_timestamp ON healthcare.audit_entries(patient_id, timestamp DESC);
CREATE INDEX idx_audit_retention ON healthcare.audit_entries(retention_until);
CREATE INDEX idx_phi_tracking_created ON phi_tracking(created_at DESC);
CREATE INDEX idx_consent_patient_resource ON consent_policies(patient_id, resource_type, expiration);
```

---

## Deployment Checklist

- [ ] Add `healthcare` module to `cli/src/nouns/mod.rs`
- [ ] Ensure `chrono` crate is in Cargo.toml dependencies
- [ ] Load SPARQL queries into Oxigraph triplestore
- [ ] Initialize FHIR patient records in database
- [ ] Configure 6-year audit log retention in PostgreSQL
- [ ] Set up encryption for storage locations (Vault, S3)
- [ ] Enable TLS 1.2+ for all API endpoints
- [ ] Configure Jaeger for OTEL span collection
- [ ] Run healthcare test suite (`cargo test healthcare_ontology_test`)
- [ ] Verify `cargo check --lib` passes with no errors

---

## References

- **FHIR R4:** https://www.hl7.org/fhir/r4/
- **PROV-O:** https://www.w3.org/TR/prov-o/
- **ODRL:** https://www.w3.org/ns/odrl/2/
- **HIPAA:** 45 CFR Parts 160, 162, 164
- **GDPR:** EU Regulation 2016/679
- **HITECH:** 42 USC § 17921-17954

---

**Version:** 1.0.0-phi | **Last Updated:** 2026-03-25 | **Status:** Production-Ready
