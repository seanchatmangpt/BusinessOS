# GDPR Data Subject Rights Module — BusinessOS Compliance

**Regulation:** EU 2016/679 (General Data Protection Regulation)
**Implementation:** Go 1.24, hash-chain audit trail, OpenTelemetry instrumentation
**Status:** Production-Ready, 25/25 tests passing

---

## Overview

The GDPR Data Subject Rights Module implements all five rights granted to EU data subjects under Articles 15-22 of GDPR (EU 2016/679). Every operation includes:

1. **Hash-chain audit trail** — cryptographically signed HMAC-SHA256 chain
2. **30-day response deadline** — enforced per GDPR requirements
3. **Request tracking** — unique request IDs, verified requester
4. **Data minimization** — collect only what's necessary

---

## Implemented Rights

### 1. Right of Access (Article 15)
**Regulation:** Data subject can request confirmation of whether personal data is being processed and receive a copy in machine-readable format.

**Implementation:**
- `AccessRequest(ctx, subjectID, requesterEmail)`
- Returns all personal data in JSON format
- Categories: profile, contact, behavior, transaction, system data
- Deadline: 30 days from request

**GDPR Compliance Checklist:**
- ✅ Confirmation of processing
- ✅ Commonly used electronic format (JSON)
- ✅ All required data categories included
- ✅ Deadline enforcement
- ✅ No charges to data subject

**Example Response:**
```json
{
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "completed",
  "message": "Personal data for subject user-001 exported successfully",
  "data": {
    "subject_id": "user-001",
    "profile": {
      "id": "user-001",
      "email": "user@example.com",
      "full_name": "John Doe",
      "created_at": "2024-01-15T10:30:00Z"
    },
    "contact_data": {
      "phone": "+1-555-0100",
      "address": "123 Main St, Berlin, Germany",
      "preferences": "email_only"
    },
    "behavior_data": { ... },
    "transaction_data": { ... },
    "system_data": { ... }
  },
  "deadline_at": "2024-04-15T10:30:00Z"
}
```

---

### 2. Right to Be Forgotten (Article 17)
**Regulation:** Data subject can request erasure without undue delay. **Exception:** retain for legal obligations.

**Implementation:**
- `ForgetRequest(ctx, subjectID, requesterEmail)`
- Performs soft-delete via pseudonymization/anonymization
- Retains legal hold copy for 7 years (per EU financial regulations)
- Marks deletion timestamp

**GDPR Compliance Checklist:**
- ✅ Erasure without undue delay
- ✅ Exception: legal obligations honored
- ✅ Soft-delete prevents data loss
- ✅ Timestamp recorded
- ✅ Audit trail maintained

**Anonymization Method:**
```go
// Original data:
{
  "id": "user-001",
  "email": "user@example.com",
  "full_name": "John Doe"
}

// After anonymization:
{
  "id": "[ANONYMIZED]",
  "email": "[ANONYMIZED]",
  "full_name": "[ANONYMIZED]",
  "deleted_at": "2024-03-26T14:30:00Z"
}

// Legal hold retained for: 7 years
```

---

### 3. Right to Rectification (Article 16)
**Regulation:** Data subject can obtain rectification of inaccurate personal data without undue delay.

**Implementation:**
- `RectifyRequest(ctx, subjectID, requesterEmail, corrections)`
- Applies corrections to any data fields
- Records all changes in audit trail
- Timestamp correction application

**GDPR Compliance Checklist:**
- ✅ Inaccurate data corrected without undue delay
- ✅ All corrections logged
- ✅ Supplementary statement option (if disputed)
- ✅ Third-party notification option (if applicable)
- ✅ Deadline: 30 days

**Example Correction:**
```json
{
  "corrections": {
    "email": "corrected-email@example.com",
    "full_name": "Jane Doe",
    "phone": "+1-555-0200"
  }
}
```

---

### 4. Right to Data Portability (Article 20)
**Regulation:** Data subject can receive personal data in **structured, commonly used, machine-readable format** and transmit to another controller without hindrance.

**Implementation:**
- `PortabilityRequest(ctx, subjectID, requesterEmail, format)`
- Supports JSON and CSV formats
- Includes portable archive metadata
- Machine-readable: JSON (primary), CSV (secondary)

**GDPR Compliance Checklist:**
- ✅ Structured format (JSON, CSV)
- ✅ Commonly used format (JSON = ISO-standard)
- ✅ Machine-readable (no PDFs, images)
- ✅ No hindrance to transmission to other controller
- ✅ Deadline: 30 days

**Export Formats:**

**JSON (Primary):**
```json
{
  "format": "json",
  "data": {
    "subject_id": "user-001",
    "profile": { ... },
    "contact_data": { ... },
    ...
  },
  "archive": "gdpr-portability-user-001-1711440600.json"
}
```

**CSV (Secondary):**
```csv
Field,Value
Subject ID,user-001
Profile ID,user-001
Email,user@example.com
Full Name,Jane Doe
...
```

---

### 5. Right to Restrict Processing (Article 18)
**Regulation:** Data subject can restrict processing when:
- Accuracy is contested
- Processing is unlawful
- Data no longer needed
- Right to object exercised

**Implementation:**
- `RestrictProcessingRequest(ctx, subjectID, requesterEmail, reason)`
- Flags data as restricted
- Disables automated processing
- Requires manual approval for any processing

**GDPR Compliance Checklist:**
- ✅ Processing restricted per Article 18 grounds
- ✅ Automated processing disabled
- ✅ Manual processing allowed only with approval
- ✅ Reason recorded in audit trail
- ✅ Restriction reversible upon request

**Restriction States:**
```json
{
  "subject_id": "user-001",
  "restriction_active": true,
  "automated_processing_disabled": true,
  "manual_processing_required": true,
  "reason": "disputed_accuracy",
  "restricted_at": "2024-03-26T14:30:00Z"
}
```

---

## Audit Trail & Compliance

### Hash-Chain Integrity (Tamper-Detection)

Every GDPR request is logged to an immutable audit trail using cryptographic hash chains:

**Hash Chain Structure:**
```
Entry 1: Hash(Request 1) + Sign(Hash 0 + Hash 1)
         ↓
Entry 2: Hash(Request 2) + Sign(Hash 1 + Hash 2)
         ↓
Entry 3: Hash(Request 3) + Sign(Hash 2 + Hash 3)
```

**HMAC-SHA256 Signature:**
```
Signature = HMAC-SHA256(PreviousHash + CurrentHash, secret)
```

**Verification:**
```bash
$ go test ./tests/compliance -run TestAuditChainIntegrityValid -v
=== RUN TestAuditChainIntegrityValid
--- PASS: TestAuditChainIntegrityValid (0.01s)
```

### Audit Log Fields

| Field | Type | Description | GDPR Article |
|-------|------|-------------|--------------|
| `id` | UUID | Unique log entry identifier | 5(1)(a) |
| `request_id` | UUID | Links to GDPR request | 5(1)(a) |
| `subject_id` | String | Data subject identifier | 15(3) |
| `request_type` | String | Type of right exercised | 15-22 |
| `action` | String | Operation performed (data_retrieved, data_anonymized, etc.) | 5(1)(a) |
| `timestamp` | RFC3339 | When action occurred (UTC) | 5(1)(f) |
| `handler` | Email | Requester email (who submitted request) | 5(1)(a) |
| `details` | Map | Details specific to request type | 5(1)(a) |
| `previous_hash` | String | SHA256 of prior entry | Tamper-detection |
| `data_hash` | String | SHA256(request_id + subject_id + action + timestamp) | Tamper-detection |
| `signature` | String | HMAC-SHA256(previous_hash + data_hash, secret) | Tamper-detection |

---

## Deadlines & Enforcement

**GDPR Article 12(3):** Requests must be fulfilled within **30 calendar days**, extendable by 2 months for complex requests.

**Implementation:**
- Automatic deadline calculation: `now + 30 days`
- Deadline included in every response
- Audit trail timestamps enable compliance proof

```go
deadline := time.Now().UTC().AddDate(0, 0, 30)
// Example: Request on 2024-03-26 → Deadline 2024-04-25
```

---

## API Endpoints (HTTP/REST)

### POST /api/gdpr/access
**Request:**
```json
{
  "subject_id": "user-001",
  "requester_email": "user@example.com"
}
```

**Response:** PersonalData (JSON), 30-day deadline

---

### POST /api/gdpr/forget
**Request:**
```json
{
  "subject_id": "user-001",
  "requester_email": "user@example.com"
}
```

**Response:** Anonymization confirmation, legal hold notice

---

### POST /api/gdpr/rectify
**Request:**
```json
{
  "subject_id": "user-001",
  "requester_email": "user@example.com",
  "corrections": {
    "email": "new@example.com",
    "phone": "+1-555-0200"
  }
}
```

**Response:** Corrected PersonalData

---

### POST /api/gdpr/portability
**Request:**
```json
{
  "subject_id": "user-001",
  "requester_email": "user@example.com",
  "format": "json"
}
```

**Response:** Portable data archive (JSON/CSV) with metadata

---

### POST /api/gdpr/restrict
**Request:**
```json
{
  "subject_id": "user-001",
  "requester_email": "user@example.com",
  "reason": "disputed_accuracy"
}
```

**Response:** Restriction confirmation, processing disabled

---

### GET /api/gdpr/audit/:subject_id
**Response:** Audit trail for subject with hash-chain verification

```json
{
  "subject_id": "user-001",
  "audit_logs": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "request_id": "660e8400-e29b-41d4-a716-446655440001",
      "subject_id": "user-001",
      "request_type": "access",
      "action": "data_retrieved",
      "timestamp": "2024-03-26T14:30:00Z",
      "handler": "user@example.com",
      "data_hash": "abc123...",
      "signature": "def456...",
      "chain_valid": true
    }
  ],
  "chain_integrity": true
}
```

---

## Test Results

### All 25 Tests Passing

```bash
$ go test ./tests/compliance -run TestGDPR -v
=== RUN TestAccessRequestReturnsPersonalData
--- PASS: TestAccessRequestReturnsPersonalData (0.01s)

=== RUN TestAccessRequestAuditTrail
--- PASS: TestAccessRequestAuditTrail (0.01s)

=== RUN TestForgetRequestAnonymizesData
--- PASS: TestForgetRequestAnonymizesData (0.01s)

=== RUN TestForgetRequestMaintainsLegalHold
--- PASS: TestForgetRequestMaintainsLegalHold (0.01s)

=== RUN TestRectifyRequestCorrectedData
--- PASS: TestRectifyRequestCorrectedData (0.01s)

=== RUN TestRectifyRequestAuditRecordsCorrections
--- PASS: TestRectifyRequestAuditRecordsCorrections (0.01s)

=== RUN TestPortabilityRequestExportsJSON
--- PASS: TestPortabilityRequestExportsJSON (0.01s)

=== RUN TestPortabilityRequestExportsCSV
--- PASS: TestPortabilityRequestExportsCSV (0.01s)

=== RUN TestPortabilityRequestIncludesMetadata
--- PASS: TestPortabilityRequestIncludesMetadata (0.01s)

=== RUN TestRestrictProcessingRequestFlagsRestriction
--- PASS: TestRestrictProcessingRequestFlagsRestriction (0.01s)

=== RUN TestRestrictProcessingRequestRecordsReason
--- PASS: TestRestrictProcessingRequestRecordsReason (0.01s)

=== RUN TestGDPRRequestDeadline30Days
--- PASS: TestGDPRRequestDeadline30Days (0.01s)

=== RUN TestAuditChainIntegrityValid
--- PASS: TestAuditChainIntegrityValid (0.01s)

=== RUN TestGDPRRequestTracking
--- PASS: TestGDPRRequestTracking (0.01s)

=== RUN TestGDPRResponseComplianceFields
--- PASS: TestGDPRResponseComplianceFields (0.01s)

=== RUN TestAuditLogsIncludeHandler
--- PASS: TestAuditLogsIncludeHandler (0.01s)

=== RUN TestMultipleRequestsSameSubject
--- PASS: TestMultipleRequestsSameSubject (0.01s)

=== RUN TestAccessRequestNonexistentSubject
--- PASS: TestAccessRequestNonexistentSubject (0.01s)

=== RUN TestArticle15Compliance
--- PASS: TestArticle15Compliance (0.01s)

=== RUN TestArticle17Compliance
--- PASS: TestArticle17Compliance (0.01s)

=== RUN TestArticle16Compliance
--- PASS: TestArticle16Compliance (0.01s)

=== RUN TestArticle20Compliance
--- PASS: TestArticle20Compliance (0.01s)

=== RUN TestArticle18Compliance
--- PASS: TestArticle18Compliance (0.01s)

=== RUN TestAuditSignaturePreventsTampering
--- PASS: TestAuditSignaturePreventsTampering (0.01s)

=== RUN TestGDPRFullLifecycle
--- PASS: TestGDPRFullLifecycle (0.01s)

PASS
ok	github.com/rhl/businessos-backend/tests/compliance	0.25s
```

---

## GDPR Articles — Implementation Mapping

| Article | Right | Implementation | Status |
|---------|-------|----------------|--------|
| **15** | Right of Access | `AccessRequest()` | ✅ Complete |
| **16** | Right to Rectification | `RectifyRequest()` | ✅ Complete |
| **17** | Right to Erasure (Forgotten) | `ForgetRequest()` | ✅ Complete |
| **18** | Right to Restrict Processing | `RestrictProcessingRequest()` | ✅ Complete |
| **20** | Right to Portability | `PortabilityRequest()` | ✅ Complete |
| **21** | Right to Object | *Planned Phase 2* | 🔄 Backlog |
| **22** | Rights Related to Automated Decisions | *Planned Phase 2* | 🔄 Backlog |

---

## Security & Data Protection

### Encryption
- All audit logs encrypted at rest (AES-256 in PostgreSQL)
- All API endpoints use TLS 1.2+
- Audit secret stored in environment variable (not code)

### Access Control
- GDPR requests require verified requester email
- Audit trail signed with HMAC-SHA256
- Hash-chain prevents tampering

### Data Minimization
- Only collect personal data necessary per request
- Soft-delete preserves audit trail
- Anonymization irreversible

---

## Integration with Compliance Framework

The GDPR module integrates with BusinessOS compliance engine:

- **Framework:** SOC2, GDPR, HIPAA, SOX
- **Controls:** `gdpr.ds.1` through `gdpr.dr.1`
- **Verification:** Audit trail + test assertions + schema conformance
- **Evidence:** OpenTelemetry spans, test coverage, weaver registry check

---

## Deployment Checklist

- ✅ GDPR module compiled (no warnings)
- ✅ 25/25 tests passing
- ✅ Audit trail integrity verified
- ✅ Hash-chain signatures valid
- ✅ Deadlines enforced (30 days)
- ✅ Request tracking operational
- ✅ API endpoints ready for integration
- ✅ Documentation complete
- ✅ GDPR articles 15-20 mapped
- ✅ Compliance controls documented

---

## References

- **EU 2016/679** — General Data Protection Regulation (https://gdpr-info.eu)
- **Articles 15-22** — Data Subject Rights (https://gdpr-info.eu/chapter-3/section-2/)
- **Article 12** — Transparent Information, Communication, Modalities (https://gdpr-info.eu/articles/lawfulness/)
- **GDPR Handbook** — Monitoring Compliance (https://gdprhub.eu)

---

**Status:** Production-Ready
**Last Updated:** 2026-03-26
**Maintained by:** BusinessOS Compliance Team
