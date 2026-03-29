---
title: "Reference: Compliance Rules Validator"
relates_to: [circuit-breaker-configuration, sox-audit-trail-integration]
---

# Compliance Rules Validator — Reference

**Frameworks Covered:** HIPAA · GDPR · SOX
**Package:** `desktop/backend-go/internal/compliance/`
**Config:** `config/compliance-rules.yaml`, `config/hipaa-rules.yaml`

---

## Framework Navigation

| Framework | Standard | Implementation File | Test File | Tests |
|-----------|----------|---------------------|-----------|-------|
| [HIPAA](#hipaa) | 45 CFR Part 164 | `internal/compliance/hipaa.go` | `internal/compliance/hipaa_test.go` | 18 |
| [GDPR](#gdpr) | EU 2016/679 | `internal/compliance/gdpr.go` | `internal/compliance/gdpr_test.go` | 32 |
| [SOX](#sox) | Sarbanes-Oxley Act (15 U.S.C.) | `internal/compliance/sox.go` | `internal/compliance/sox_test.go` | 14 |

---

## HIPAA

**Location:** `desktop/backend-go/internal/compliance/hipaa.go`
**Test Suite:** `desktop/backend-go/internal/compliance/hipaa_test.go` (18 tests)
**Compliance Standard:** 45 CFR Part 164 (HIPAA Security Rule & Breach Notification)

### Overview

The HIPAA Compliance Rule Validator enforces Protected Health Information (PHI) security requirements across BusinessOS. It validates five compliance categories:

1. **Access Control** — Authorization roles for PHI access
2. **Encryption** — TLS 1.2+ in transit, AES-256 at rest
3. **Audit Logging** — All PHI accesses logged with timestamp/user/action
4. **Data Retention** — PHI deleted after 6-year retention period
5. **Breach Notification** — Automatic alert on unencrypted transmission

### Core Types

#### HIPAARuleValidator

Primary validator struct. Thread-safe via `sync.RWMutex`.

```go
v := NewHIPAARuleValidator(logger)

// Register authorized users with roles
v.RegisterAuthorizedUser("doctor1", []string{"hipaa_user"})

// Validate access control
if err := v.ValidateAccessControl(ctx, "doctor1", "read"); err != nil {
    // User not authorized
}

// Check encryption (TLS 1.2+, AES-256)
if err := v.ValidateEncryption(ctx, tls.VersionTLS12, "AES-256"); err != nil {
    // Encryption requirement violated
}

// Log PHI access
v.LogPHIAccess(ctx, PHIAccessEvent{
    UserID:     "doctor1",
    Action:     "read",
    ResourceID: "patient123",
    Outcome:    "success",
})

// Check retention (6 years)
v.RegisterPHIData("patient123")
if err := v.ValidateRetention(ctx, "patient123"); err != nil {
    // PHI exceeds retention limit
}

// Detect breaches
if err := v.DetectBreach(ctx, tlsVersion, unencryptedData); err != nil {
    // Breach detected
}
```

#### PHIAccessEvent

Audit trail entry for single PHI access.

```go
type PHIAccessEvent struct {
    ID          string            // Auto-generated if empty
    Timestamp   time.Time         // Auto-set if zero
    UserID      string            // Required: who accessed
    Action      string            // Required: read|write|delete|access
    ResourceID  string            // Required: patient ID, record ID
    Outcome     string            // success|denied|error
    Details     map[string]string // Optional context
    EncryptedTx bool              // Transport encryption flag
}
```

#### BreachNotification

Detected unencrypted transmission event.

```go
type BreachNotification struct {
    ID              string    // Unique breach ID
    Timestamp       time.Time // When detected
    ResourceID      string    // What was transmitted unencrypted
    UnencryptedData string    // Data content (may be redacted)
    DetectedBy      string    // Which rule detected it
    Severity        string    // "critical" | "high"
}
```

#### ComplianceMetrics

Current violation counters.

```go
type ComplianceMetrics struct {
    AccessViolations     int64  // Unauthorized access attempts
    EncryptionViolations int64  // Bad TLS version or algo
    AuditViolations      int64  // Unlogged accesses
    RetentionViolations  int64  // PHI exceeding 6-year limit
    BreachNotifications  int64  // Unencrypted transmissions detected
    TotalViolations      int64  // Sum of all above
}
```

---

### Authorized Roles

Valid HIPAA authorization roles (case-sensitive):

| Role | Permissions | 45 CFR Reference |
|------|-------------|------------------|
| `hipaa_admin` | read, write, delete, access | 164.308(a)(4) |
| `hipaa_user` | read, write, access | 164.308(a)(4) |
| `hipaa_auditor` | read, access | 164.312(b) |
| `phi_viewer` | read, access | 164.502(b) |
| `phi_editor` | read, write, access | 164.502(b) |

---

### Encryption Requirements

#### Transit (In-Flight)

- **TLS Version:** ≥ TLS 1.2 (0x0303)
- **Supported:** TLS 1.2, TLS 1.3
- **Rejected:** SSL 3.0, TLS 1.0, TLS 1.1
- **45 CFR Reference:** 164.312(a)(2)(i)

```go
// Valid: TLS 1.2+
v.ValidateEncryption(ctx, tls.VersionTLS12, "AES-256")  // ✓

// Invalid: TLS 1.1
v.ValidateEncryption(ctx, tls.VersionTLS11, "AES-256")  // ✗
```

#### At Rest

- **Algorithm:** AES-256 only
- **Key Management:** Customer-controlled
- **45 CFR Reference:** 164.312(a)(2)(i)

```go
// Valid
v.ValidateEncryption(ctx, tls.VersionTLS12, "AES-256")  // ✓

// Invalid
v.ValidateEncryption(ctx, tls.VersionTLS12, "AES-128")  // ✗
v.ValidateEncryption(ctx, tls.VersionTLS12, "DES")      // ✗
```

---

### Data Retention (6 Years)

HIPAA requires minimum 6-year retention of PHI records (45 CFR 164.404).

```go
// Register when PHI is created
v.RegisterPHIData("patient_id_001")

// Later, validate it hasn't exceeded limit
if err := v.ValidateRetention(ctx, "patient_id_001"); err == nil {
    // PHI within 6-year limit — OK
}

// After proper deletion, unregister
v.DeletePHIData("patient_id_001")
```

**Timeline:**
- Day 0: Data created, `RegisterPHIData()` called
- Year 6: Data may be deleted
- Year 6.001+: `ValidateRetention()` returns error

---

### Audit Logging

All PHI access must be logged with:
- **User ID** — Who accessed
- **Resource ID** — What was accessed (patient/record)
- **Action** — Operation type (read/write/delete)
- **Timestamp** — When accessed (UTC)
- **Outcome** — Success/denied/error

#### Log Once

```go
v.LogPHIAccess(ctx, PHIAccessEvent{
    UserID:     "doctor1",
    Action:     "read",
    ResourceID: "patient123",
    Outcome:    "success",
    EncryptedTx: true,
})
```

#### Verify Later

```go
found, err := v.ValidateAuditLogging(ctx, "doctor1", "patient123")
if found {
    // Access found in audit trail
}
```

#### Retrieve Audit Log

```go
log, _ := v.GetAuditLog(ctx)
for _, event := range log {
    fmt.Printf("%s: %s %s %s → %s\n",
        event.Timestamp,
        event.UserID,
        event.Action,
        event.ResourceID,
        event.Outcome,
    )
}
```

---

### Breach Detection

Automatic detection of unencrypted PHI transmission.

#### Detection

```go
// If TLS version < 1.2, breach is detected and callbacks invoked
if err := v.DetectBreach(ctx, tls.VersionTLS11, "patient_ssn_123-45-6789"); err != nil {
    // Breach detected — error returned, callbacks called
}
```

#### Callbacks

Register callbacks to handle breaches (e.g., send alerts, create tickets):

```go
v.RegisterBreachCallback(func(ctx context.Context, breach *BreachNotification) error {
    log.Printf("BREACH: %s detected at %s", breach.ID, breach.Timestamp)
    // Send alert email, create incident, etc.
    return nil
})

// Later, when breach detected:
v.DetectBreach(ctx, tls.VersionTLS10, "patient_data")
// → Callback is invoked with breach details
```

---

### 45 CFR Mapping

Complete mapping to HIPAA Security Rule sections:

| Rule Category | CFR Section | Requirement |
|---|---|---|
| **Access Control** | 45 CFR 164.308(a)(4) | Verify user authorization before PHI access |
| **Access Management** | 45 CFR 164.308(a)(4)(ii) | Document access control policies |
| **Encryption (Transit)** | 45 CFR 164.312(a)(2)(i) | TLS 1.2+ for all PHI transmission |
| **Encryption (Rest)** | 45 CFR 164.312(a)(2)(i) | AES-256 or equivalent for stored PHI |
| **Audit Controls** | 45 CFR 164.312(b) | Log all PHI accesses with user/time/action |
| **Integrity Controls** | 45 CFR 164.308(a)(2)(ii) | Verify PHI integrity (hashing, checksums) |
| **Data Retention** | 45 CFR 164.404 | Retain PHI records ≥ 6 years |
| **Breach Notification** | 45 CFR 164.404 | Notify affected individuals within 60 days |
| **Minimum Necessary** | 45 CFR 164.502(b) | Only access PHI needed for treatment/ops |
| **Business Associates** | 45 CFR 164.308(b) | Require BA agreements with vendors |

**Compliance Function Reference:**

```go
mapping := CFRMapping()
// Returns map[string]string with all CFR citations
```

---

### HIPAA Rule IDs (hipaa-rules.yaml)

| Rule ID | Title | Condition | Action | Severity |
|---------|-------|-----------|--------|----------|
| `hipaa.ac.1` | Only authorized users can access PHI | `user.role != admin AND data.classification == phi` | escalate | critical |
| `hipaa.er.1` | All PHI data must be encrypted at rest | `data.encrypted == false AND data.contains_phi == true` | create_gap | critical |
| `hipaa.et.1` | HTTPS/TLS required for PHI transmission | `transmission.protocol != https AND message.contains_phi == true` | escalate | critical |
| `hipaa.al.1` | All PHI access must be logged with timestamp and user | `audit_log.missing_phi_access_entries == true` | create_gap | high |
| `hipaa.dr.1` | PHI retained only as long as required (max 7 years) | `data.retention_days > 2555 AND data.contains_phi == true` | notify | medium |

---

### HIPAA Metrics & Monitoring

#### Get Current Metrics

```go
metrics := v.GetMetrics()
fmt.Printf("Total Violations: %d\n", metrics.TotalViolations)
fmt.Printf("  Access: %d\n", metrics.AccessViolations)
fmt.Printf("  Encryption: %d\n", metrics.EncryptionViolations)
fmt.Printf("  Audit: %d\n", metrics.AuditViolations)
fmt.Printf("  Retention: %d\n", metrics.RetentionViolations)
fmt.Printf("  Breaches: %d\n", metrics.BreachNotifications)
```

#### Integrate with Prometheus

```go
import "github.com/prometheus/client_golang/prometheus"

// Expose metrics endpoint
hipaa_violations := prometheus.NewGaugeVec(...)

go func() {
    ticker := time.NewTicker(30 * time.Second)
    for range ticker.C {
        metrics := v.GetMetrics()
        hipaa_violations.WithLabelValues("access").Set(float64(metrics.AccessViolations))
        hipaa_violations.WithLabelValues("encryption").Set(float64(metrics.EncryptionViolations))
        // ...
    }
}()
```

---

### HIPAA Test Coverage

18 tests covering all compliance categories:

| Test | Coverage |
|------|----------|
| `TestNewHIPAARuleValidator_DefaultConfiguration` | Initialization |
| `TestRegisterAuthorizedUser_ValidRoles` | User registration with role validation |
| `TestValidateAccessControl_AuthorizedAccess` | Successful PHI access |
| `TestValidateAccessControl_RoleBasedPermissions` | Role-specific permissions |
| `TestValidateEncryption_TLSVersion` | TLS version validation (1.2+) |
| `TestValidateEncryption_EncryptionAlgorithm` | Encryption algorithm (AES-256) |
| `TestLogPHIAccess_AuditTrail` | Audit log recording |
| `TestValidateAuditLogging_AccessFound` | Audit log verification |
| `TestRegisterPHIData_TrackingCreation` | PHI data registration |
| `TestValidateRetention_WithinLimit` | Retention validation (within 6 years) |
| `TestValidateRetention_ExceedsLimit` | Retention validation (exceeds limit) |
| `TestDeletePHIData_RemovalFromTracking` | PHI data deletion |
| `TestDetectBreach_UnencryptedTransmission` | Breach detection on TLS < 1.2 |
| `TestRegisterBreachCallback_CallbackInvoked` | Breach notification callbacks |
| `TestGetMetrics_ViolationTracking` | Metrics collection |
| `TestCFRMapping_ComplianceReferences` | CFR mapping completeness |
| `TestGetAuditLog_CopyReturned` | Audit log retrieval |
| `TestComplexScenario_MultiUserMultiAction` | Multi-user workflow (realistic scenario) |

**Run Tests:**
```bash
cd desktop/backend-go
go test ./internal/compliance/hipaa* -v -count=1
```

---

### HIPAA Integration Example

Real-world workflow:

```go
// Initialize
v := NewHIPAARuleValidator(logger)

// Register staff
v.RegisterAuthorizedUser("dr_smith", []string{"hipaa_user"})
v.RegisterAuthorizedUser("nurse_jones", []string{"phi_editor"})
v.RegisterAuthorizedUser("auditor_lee", []string{"hipaa_auditor"})

// Register patient PHI
v.RegisterPHIData("patient_id_12345")

// Doctor reads record (over TLS 1.2, AES-256)
if err := v.ValidateAccessControl(ctx, "dr_smith", "read"); err != nil {
    return fmt.Errorf("access denied: %w", err)
}
if err := v.ValidateEncryption(ctx, tls.VersionTLS12, "AES-256"); err != nil {
    return fmt.Errorf("encryption failed: %w", err)
}
v.LogPHIAccess(ctx, PHIAccessEvent{
    UserID:     "dr_smith",
    Action:     "read",
    ResourceID: "patient_id_12345",
    Outcome:    "success",
})

// After 6 years, delete
v.DeletePHIData("patient_id_12345")

// Monitor metrics
metrics := v.GetMetrics()
if metrics.TotalViolations > 0 {
    log.Printf("WARNING: %d compliance violations", metrics.TotalViolations)
}
```

---

### HIPAA Configuration

Default configuration (hardcoded, can be extended):

```
Validator Configuration:
  Retention Days:    2190 (6 years)
  TLS Min Version:   0x0303 (TLS 1.2)
  Encryption Algo:   AES-256
  Audit Max Entries: 100,000 (oldest evicted if exceeded)
```

---

### HIPAA Thread Safety

All operations are thread-safe via `sync.RWMutex`:

- **Read operations** (validation, querying) use RLock
- **Write operations** (registration, logging) use Lock
- **Atomic metrics** use `sync.atomic.Int64` (wait-free)

---

## GDPR

**Location:** `desktop/backend-go/internal/compliance/gdpr.go`
**Test Suite:** `desktop/backend-go/internal/compliance/gdpr_test.go` (32 tests)
**Compliance Standard:** EU Regulation 2016/679 (General Data Protection Regulation)

### Overview

The GDPR Service enforces data subject rights and personal data handling requirements for EU residents and organizations subject to GDPR jurisdiction. It implements the five data subject rights with hash-chain audit trails.

Enforced compliance categories:

1. **Data Subject Access** — Article 15: fulfill access requests within 30 days
2. **Right to Erasure** — Article 17: anonymize personal data on request (soft-delete)
3. **Right to Rectification** — Article 16: apply and log corrections to personal data
4. **Data Portability** — Article 20: export personal data in JSON or CSV format
5. **Restriction of Processing** — Article 18: flag data as restricted, disable automated processing
6. **Tamper-Proof Audit Trail** — HMAC-SHA256 hash chain over all rights operations

### Core Types

#### GDPRService

Primary service struct. Not thread-safe by default (use external mutex for concurrent access).

```go
svc := NewGDPRService("hmac-audit-secret", logger)

// Insert sample data for testing
svc.InsertSampleData("user-123")

// Fulfill access request (Article 15)
resp, err := svc.AccessRequest(ctx, "user-123", "dpo@example.com")

// Right to be forgotten (Article 17)
resp, err := svc.ForgetRequest(ctx, "user-123", "dpo@example.com")

// Rectification (Article 16)
corrections := map[string]interface{}{"email": "new@example.com"}
resp, err := svc.RectifyRequest(ctx, "user-123", "dpo@example.com", corrections)

// Portability (Article 20) — formats: "json", "csv"
resp, err := svc.PortabilityRequest(ctx, "user-123", "dpo@example.com", "json")

// Restrict processing (Article 18)
resp, err := svc.RestrictProcessingRequest(ctx, "user-123", "dpo@example.com", "legal hold")

// Retrieve request by ID
req := svc.QueryGDPRRequest(resp.RequestID)

// Audit trail by subject
trail := svc.GetAuditTrail("user-123")

// Verify audit chain integrity
valid, issues := svc.VerifyAuditChainIntegrity()
```

#### GDPRRequest

Record of a single data subject rights request.

```go
type GDPRRequest struct {
    ID             string      // UUID
    SubjectID      string      // Data subject identifier
    RequestType    string      // access|be_forgotten|rectification|portability|restrict_processing
    Timestamp      time.Time   // When request was received (UTC)
    Status         string      // pending|approved|completed|denied
    ResponseData   interface{} // Optional: corrections applied, export produced
    Reason         string      // Optional: restriction reason
    RequesterEmail string      // DPO or requester contact
    Verified       bool        // Identity verified flag
    DeadlineAt     time.Time   // 30-day response deadline
}
```

#### GDPRResponse

Returned from all five rights operations.

```go
type GDPRResponse struct {
    RequestID  string      // UUID of GDPRRequest created
    Status     string      // "completed" on success
    Message    string      // Human-readable outcome description
    Data       interface{} // Operation-specific payload
    Timestamp  time.Time   // When completed (UTC)
    DeadlineAt time.Time   // 30 days from Timestamp
}
```

#### PersonalData

Aggregated personal data structure for a single data subject.

```go
type PersonalData struct {
    SubjectID       string                 // Data subject identifier
    Profile         *DataSubject           // Core profile (name, email)
    ContactData     map[string]interface{} // Phone, address, preferences
    BehaviorData    map[string]interface{} // Login history, theme preferences
    TransactionData map[string]interface{} // Purchase history, currency
    SystemData      map[string]interface{} // User agent, IP geolocation
    ExportedAt      time.Time              // When this snapshot was generated
}
```

#### GDPRAuditLog

Hash-chain audit trail entry. Each entry signs the previous entry's hash.

```go
type GDPRAuditLog struct {
    ID           string                 // UUID
    RequestID    string                 // Parent GDPRRequest ID
    SubjectID    string                 // Data subject
    RequestType  string                 // GDPR right exercised
    Action       string                 // data_retrieved|data_anonymized|data_corrected|data_exported|processing_restricted
    Timestamp    time.Time              // UTC
    Handler      string                 // Requester email / DPO identity
    Details      map[string]interface{} // Operation-specific context
    PreviousHash string                 // SHA256 of prior entry's data (empty for first)
    DataHash     string                 // SHA256(RequestID+SubjectID+RequestType+Action+Timestamp)
    Signature    string                 // HMAC-SHA256(PreviousHash+DataHash, auditSecret)
}
```

---

### GDPR Rights Constants

| Constant | Value | GDPR Article |
|----------|-------|--------------|
| `RightOfAccess` | `"access"` | Article 15 |
| `RightToBeForotten` | `"be_forgotten"` | Article 17 |
| `RightOfRectification` | `"rectification"` | Article 16 |
| `RightOfPortability` | `"portability"` | Article 20 |
| `RightToRestrictProcessing` | `"restrict_processing"` | Article 18 |

---

### GDPR PII Fields (PersonalData Categories)

GDPR PII is equivalent to HIPAA PHI in scope. The following data categories are subject to all five rights:

| Category | Fields | GDPR Article |
|----------|--------|--------------|
| **Profile** | id, email, full_name, created_at, restricted_at, deleted_at | Art. 4(1), Art. 15 |
| **Contact** | phone, address, postal_code, city, preferences | Art. 5(1)(c), Art. 15 |
| **Behavior** | last_login, login_count, preference_theme | Art. 5(1)(b), Art. 15 |
| **Transaction** | total_purchases, currency, purchase_history | Art. 5(1)(b), Art. 15 |
| **System** | user_agent, ip_geolocation | Art. 5(1)(e), Art. 17(3)(a) |

---

### GDPR Enforcement Actions

What happens when each right is exercised:

| Operation | Method | Enforcement | 30-Day Deadline |
|-----------|--------|-------------|-----------------|
| Access Request | `AccessRequest()` | Collect all PII across all categories; return as JSON | Yes |
| Erasure Request | `ForgetRequest()` | Soft-delete: anonymize Profile (ID, email, name → `[ANONYMIZED]`); flag `DeletedAt` | Yes |
| Rectification | `RectifyRequest()` | Apply field-level corrections to `ContactData`; log corrections map | Yes |
| Portability | `PortabilityRequest()` | Export full `PersonalData` as JSON or CSV archive | Yes |
| Restrict Processing | `RestrictProcessingRequest()` | Set `Profile.RestrictedAt`; disable automated processing flag | Yes |

All operations:
- Create a `GDPRRequest` record with status `"approved"` or `"completed"`
- Append a `GDPRAuditLog` entry signed into the HMAC hash chain
- Return a `GDPRResponse` with `Status: "completed"` and `DeadlineAt: now + 30 days`

---

### GDPR Audit Chain Integrity

The GDPR service uses a tamper-evident HMAC-SHA256 hash chain over all audit log entries:

```
Entry N:
  DataHash  = SHA256(RequestID + SubjectID + RequestType + Action + Timestamp)
  Signature = HMAC-SHA256(PreviousHash + DataHash, auditSecret)

Entry N+1:
  PreviousHash = Entry N's DataHash   ← chain link
  DataHash     = SHA256(...)
  Signature    = HMAC-SHA256(Entry N's DataHash + new DataHash, auditSecret)
```

To verify the chain is unmodified:

```go
valid, issues := svc.VerifyAuditChainIntegrity()
if !valid {
    for _, issue := range issues {
        log.Printf("GDPR audit chain violation: %s", issue)
    }
}
```

Verification checks for each entry:
1. Recomputes `DataHash` from fields — mismatch means data was modified
2. Recomputes `Signature` — mismatch means tampered after creation
3. Confirms `PreviousHash == prior entry's DataHash` — mismatch means chain was spliced

---

### GDPR Rule IDs (compliance-rules.yaml)

| Rule ID | Title | Condition | Action | Severity |
|---------|-------|-----------|--------|----------|
| `gdpr.ds.1` | Data subject access requests must be fulfilled | `data_subject_request.pending == true AND days_elapsed > 30` | escalate | critical |
| `gdpr.cm.1` | Explicit consent required for personal data processing | `data_processing.requires_consent == true AND user.consent_given != true` | create_gap | critical |
| `gdpr.dpa.1` | Data Processing Agreement with all sub-processors | `processor.dpa_signed != true AND processor.handles_data == true` | escalate | critical |
| `gdpr.dm.1` | Only collect personal data actually needed | `data_collected.field_count > data_needed.field_count` | create_gap | medium |
| `gdpr.dr.1` | EU personal data must reside in EU | `data.contains_pii == true AND data.location != eu AND org.region == eu` | escalate | critical |

---

### GDPR Article Mapping

| GDPR Article | Subject | Implemented By |
|---|---|---|
| Art. 5(1)(c) | Data minimization | `gdpr.dm.1` rule |
| Art. 7 | Consent | `gdpr.cm.1` rule |
| Art. 12-21 | Data subject rights (30-day deadline) | `gdpr.ds.1` rule; all five request methods |
| Art. 15 | Right of access | `AccessRequest()` |
| Art. 16 | Right to rectification | `RectifyRequest()` |
| Art. 17 | Right to be forgotten | `ForgetRequest()` — soft-delete + anonymization |
| Art. 18 | Right to restrict processing | `RestrictProcessingRequest()` |
| Art. 20 | Right to data portability | `PortabilityRequest()` (JSON/CSV) |
| Art. 28 | Data Processing Agreements | `gdpr.dpa.1` rule |
| Art. 44-49 | International data transfers | `gdpr.dr.1` rule |

---

### GDPR Test Coverage

32 tests + 2 benchmarks:

| Test | Coverage |
|------|----------|
| `TestNewGDPRService_Initialization` | Constructor, internal state |
| `TestNewGDPRService_NilLogger` | Nil logger fallback |
| `TestGDPRConstants_Values` | All 5 right constants |
| `TestAccessRequest_Success` | Art. 15 — basic access |
| `TestAccessRequest_WithSampleData` | Art. 15 — with real PII data |
| `TestAccessRequest_CreatesAuditLog` | Audit trail entry on access |
| `TestForgetRequest_Success` | Art. 17 — basic erasure |
| `TestForgetRequest_AnonymizesData` | Soft-delete + anonymization |
| `TestForgetRequest_CreatesAuditLog` | Audit trail entry on erasure |
| `TestRectifyRequest_Success` | Art. 16 — basic rectification |
| `TestRectifyRequest_AppliesCorrections` | Field-level correction application |
| `TestRectifyRequest_CreatesAuditLog` | Audit trail entry on correction |
| `TestPortabilityRequest_JSONFormat` | Art. 20 — JSON export |
| `TestPortabilityRequest_CSVFormat` | Art. 20 — CSV export |
| `TestPortabilityRequest_DefaultFormat` | Art. 20 — format default |
| `TestPortabilityRequest_CreatesAuditLog` | Audit trail entry on export |
| `TestRestrictProcessingRequest_Success` | Art. 18 — basic restriction |
| `TestRestrictProcessingRequest_FlagsProfile` | RestrictedAt timestamp set |
| `TestRestrictProcessingRequest_CreatesAuditLog` | Audit trail entry on restriction |
| `TestQueryGDPRRequest_Found` | Request lookup by ID |
| `TestQueryGDPRRequest_NotFound` | Nil return for unknown ID |
| `TestGetAuditTrail_FiltersBySubject` | Per-subject trail filtering |
| `TestGetAuditTrail_Empty` | Empty result for unknown subject |
| `TestVerifyAuditChainIntegrity_EmptyChain` | Empty chain valid |
| `TestVerifyAuditChainIntegrity_ValidChain` | Multi-entry valid chain |
| `TestVerifyAuditChainIntegrity_ChainLinks` | PreviousHash linking |
| `TestInsertSampleData_CreatesProfile` | Sample data structure |
| `TestGDPRRequestDeadline_Is30Days` | 30-day deadline for all 5 rights |
| `TestComplexScenario_FullDataSubjectLifecycle` | Full lifecycle: insert → access → rectify → port → restrict → forget |
| `BenchmarkGDPRAccessRequest` | Access request throughput |
| `BenchmarkGDPRVerifyAuditChain` | Chain verification throughput (10-entry chain) |

**Run Tests:**
```bash
cd desktop/backend-go
go test ./internal/compliance/gdpr* -v -count=1
```

---

### GDPR Integration Example

Full data subject lifecycle (access → rectify → export → restrict → forget):

```go
svc := NewGDPRService("your-hmac-secret-here", logger)

// Register subject data
svc.InsertSampleData("user-eu-456")

// 1. Subject requests their data (Art. 15)
accessResp, err := svc.AccessRequest(ctx, "user-eu-456", "dpo@company.com")
// accessResp.Data contains all PII categories

// 2. Subject corrects inaccurate field (Art. 16)
_, err = svc.RectifyRequest(ctx, "user-eu-456", "dpo@company.com",
    map[string]interface{}{"email": "corrected@example.com", "city": "Berlin"})

// 3. Subject requests portable export (Art. 20)
portResp, err := svc.PortabilityRequest(ctx, "user-eu-456", "dpo@company.com", "json")
// portResp.Data["archive"] = "gdpr-portability-user-eu-456-<timestamp>.json"

// 4. Subject restricts processing during dispute (Art. 18)
_, err = svc.RestrictProcessingRequest(ctx, "user-eu-456", "dpo@company.com", "data accuracy dispute")

// 5. Subject exercises erasure (Art. 17)
_, err = svc.ForgetRequest(ctx, "user-eu-456", "dpo@company.com")
// Profile fields replaced with [ANONYMIZED], legal hold retained for 7 years

// 6. DPO verifies complete audit trail
trail := svc.GetAuditTrail("user-eu-456")
// len(trail) == 5 (one entry per operation above)

// 7. Verify chain has not been tampered
valid, issues := svc.VerifyAuditChainIntegrity()
if !valid {
    log.Printf("GDPR audit integrity failure: %v", issues)
}
```

---

## SOX

**Location:** `desktop/backend-go/internal/compliance/sox.go`
**Test Suite:** `desktop/backend-go/internal/compliance/sox_test.go` (14 tests + 2 benchmarks)
**Compliance Standard:** Sarbanes-Oxley Act, Section 404(b) (IT General Controls for Financial Systems)

### Overview

The SOX Audit Validator enforces immutable audit trail requirements for all financial data mutations. It implements SOX 404(b) requirements:

1. **Immutable Audit Trail** — Append-only log; no UPDATE or DELETE on entries
2. **Who / What / When / Why** — Actor, operation, timestamp, and business reason code captured on every mutation
3. **Before/After Values** — JSON snapshots of resource state pre- and post-mutation
4. **Hash-Chain Integrity** — SHA-256 data hashes + HMAC-SHA256 signatures linking all entries
5. **7-Year Retention** — All audit entries retained for 7 years (2,555 days) per SOX record rules
6. **Tamper Detection** — `VerifyAuditTrailImmutability()` detects any modification after commit

### Core Types

#### SOXAuditValidator

Primary validator struct. Thread-safe via `sync.RWMutex`.

```go
v := NewSOXAuditValidator("secret-key-at-least-32-bytes-long", logger)

// Record a financial mutation
entry, err := v.RecordFinancialMutation(
    ctx,
    "user-123",               // actor (required)
    "human",                  // actorType
    OperationUpdate,          // operation
    Transaction,              // resourceType
    "txn-456",                // resourceID (required)
    "periodic_reconciliation",// reasonCode (required)
    before,                   // beforeValues JSON (nil for CREATE)
    after,                    // afterValues JSON (required for non-DELETE)
)

// Verify complete trail immutability
valid, issues := v.VerifyAuditTrailImmutability()

// Get history for one resource
history := v.GetAuditHistory(ctx, Transaction, "txn-456")

// Verify a specific entry has not been modified
isValid, msg := v.VerifyEntryImmutability(entry.ID)

// Snapshot for SOX compliance reporting
all := v.GetCompleteAuditTrail()

// Fingerprint of all entries (detect collection-level tampering)
fingerprint := v.ComputeAuditFingerprint()
```

#### SOXAuditEntry

Immutable audit trail entry for a single financial mutation.

```go
type SOXAuditEntry struct {
    ID                string                // UUID
    SequenceNum       int64                 // Monotonically increasing (1, 2, 3, ...)
    Timestamp         time.Time             // UTC when action occurred
    Actor             string                // user ID, service account, agent ID
    ActorType         string                // human|service_account|agent|system
    Operation         OperationType         // CREATE|UPDATE|DELETE|READ
    ResourceType      FinancialResourceType // ledger_entry|transaction|account|...
    ResourceID        string                // Unique ID of affected resource
    ReasonCode        string                // Business justification (required)
    BeforeValues      json.RawMessage       // Resource state snapshot before change
    AfterValues       json.RawMessage       // Resource state snapshot after change
    ChangeSummary     string                // Human-readable diff description
    PreviousHash      string                // SHA-256 of prior entry (empty for first)
    DataHash          string                // SHA-256(actor+operation+resourceID+ts+before+after)
    Signature         string                // HMAC-SHA256(PreviousHash+DataHash, secret)
    ChainValid        bool                  // Chain link verified
    IntegrityVerified bool                  // Data hash verified
    Status            string                // "committed" (immutable)
}
```

---

### SOX Operation Types

| Constant | Value | When to Use |
|----------|-------|-------------|
| `OperationCreate` | `"CREATE"` | New financial record created |
| `OperationUpdate` | `"UPDATE"` | Existing financial record modified |
| `OperationDelete` | `"DELETE"` | Financial record deleted (soft or hard) |
| `OperationRead` | `"READ"` | Sensitive financial record read by auditor |

---

### SOX Financial Resource Types

All resource types tracked for SOX compliance:

| Constant | Value | Description |
|----------|-------|-------------|
| `LedgerEntry` | `"ledger_entry"` | General ledger entries |
| `Transaction` | `"transaction"` | Financial transactions |
| `Account` | `"account"` | Chart of accounts entries |
| `JournalEntry` | `"journal_entry"` | Accounting journal entries |
| `PaymentRecord` | `"payment_record"` | Payment processing records |
| `InvoiceRecord` | `"invoice_record"` | Invoices issued or received |
| `ExpenseRecord` | `"expense_record"` | Expense reports and approvals |
| `BudgetAlloc` | `"budget_allocation"` | Budget allocation records |

---

### SOX Actor Types

| ActorType | When to Use |
|-----------|-------------|
| `"human"` | User-initiated changes via UI |
| `"service_account"` | Automated processes, cron jobs |
| `"agent"` | AI agent operations |
| `"system"` | System-initiated (batch reconciliation, integrity checks) |

---

### SOX Audit Trail Integrity

Each entry is cryptographically linked to the prior entry:

```
Entry 1 (SequenceNum=1):
  PreviousHash = ""   ← genesis entry
  DataHash     = SHA256("actor:op:resourceID:unix_ts:before:after")
  Signature    = HMAC-SHA256("" + DataHash, secret)

Entry 2 (SequenceNum=2):
  PreviousHash = Entry 1's DataHash       ← chain link
  DataHash     = SHA256(...)
  Signature    = HMAC-SHA256(Entry1.DataHash + DataHash, secret)

Entry N (SequenceNum=N):
  PreviousHash = Entry N-1's DataHash     ← chain link
  ...
```

`VerifyAuditTrailImmutability()` checks three invariants for every entry:
1. Recomputed `DataHash` matches stored `DataHash` — proves data not modified
2. Recomputed `Signature` matches stored `Signature` — proves tamper detection key is intact
3. `PreviousHash` matches prior entry's `DataHash` — proves chain not spliced

`ComputeAuditFingerprint()` returns `SHA256(concat all DataHashes)` — a single value for collection-level integrity.

---

### SOX Required Parameters

`RecordFinancialMutation` enforces these constraints or returns an error:

| Parameter | Required | Constraint |
|-----------|----------|------------|
| `actor` | Yes | Non-empty string; cannot be anonymous |
| `resourceID` | Yes | Non-empty string; uniquely identifies affected resource |
| `reasonCode` | Yes | Non-empty; business justification for the change |
| `afterValues` | Yes (non-DELETE) | JSON snapshot of resource after change |
| `beforeValues` | No (nil for CREATE) | JSON snapshot of resource before change |

Validation errors returned:

| Error Text | Cause |
|------------|-------|
| `"SOX audit: actor cannot be empty"` | Empty actor |
| `"SOX audit: resourceID cannot be empty"` | Empty resourceID |
| `"SOX audit: reasonCode cannot be empty (business justification required)"` | Empty reasonCode |
| `"SOX audit: afterValues required for CREATE operation"` | afterValues nil for non-DELETE |

---

### SOX Rule IDs (compliance-rules.yaml)

| Rule ID | Title | Condition | Action | Severity |
|---------|-------|-----------|--------|----------|
| `sox.itg.1` | Change management enforces segregation of duties | `change.requires_approval == true AND change.approved_by == change.made_by` | escalate | critical |
| `sox.sa.1` | Financial systems must maintain 99.9% uptime | `system.measured_uptime < 99.9` | escalate | critical |
| `sox.al.1` | All access to financial data must be logged for 7 years | `financial_data.access_logged == false OR audit_log.retention_days < 2555` | create_gap | high |
| `sox.cm.1` | All production changes must be documented and approved | `production_change.documented == false` | escalate | high |
| `sox.fdi.1` | Financial records must have verified checksums/hashes | `financial_record.has_checksum == false OR checksum.verified == false` | escalate | critical |

---

### SOX Segregation of Duties

The `sox.itg.1` rule enforces that the same person cannot both make and approve a production change. This is validated at the rule-engine level against the `change.approved_by` and `change.made_by` fields. The `SOXAuditValidator` records who made every mutation; segregation of duties policy must be enforced at the application layer before calling `RecordFinancialMutation`.

---

### SOX Audit Trail Requirements Summary

| Requirement | Implementation |
|---|---|
| Append-only (no UPDATE/DELETE on entries) | `entries` slice is write-only; `GetCompleteAuditTrail()` returns a copy |
| Who | `Actor` field required; `ActorType` categorizes human vs. automated |
| What | `Operation` + `ResourceType` + `ResourceID` + `BeforeValues` + `AfterValues` |
| When | `Timestamp` (UTC) recorded at write time |
| Why | `ReasonCode` required (e.g., `"periodic_reconciliation"`, `"error_correction"`) |
| Change summary | `ChangeSummary` auto-generated human-readable diff |
| 7-year retention | Tracked via `sox.al.1` rule; retention_days ≥ 2555 |
| Hash-chain | SHA-256 + HMAC-SHA256 on every entry |
| Fingerprint | `ComputeAuditFingerprint()` = SHA256 of all DataHashes |

---

### SOX Test Coverage

14 tests + 2 benchmarks:

| Test | Coverage |
|------|----------|
| `TestSOXAuditValidatorRecordFinancialMutation` | Single entry — fields, sequence, status, hashes |
| `TestSOXAuditValidatorMultipleEntries` | Chain links between 3 entries |
| `TestSOXAuditValidatorVerifyImmutability` | Entry + chain verification |
| `TestSOXAuditValidatorChainIntegrity` | 5-entry batch chain integrity |
| `TestSOXAuditValidatorBeforeAfterValues` | Before/after JSON snapshot accuracy |
| `TestSOXAuditValidatorGetAuditHistory` | Resource-specific history filtering |
| `TestSOXAuditValidatorAllOperationTypes` | All 4 operation types (CREATE/READ/UPDATE/DELETE) |
| `TestSOXAuditValidatorValidationErrors` | Required field validation errors |
| `TestSOXAuditValidatorSignatureImmutability` | Tamper detection (modified AfterValues fails) |
| `TestSOXAuditValidatorComputeAuditFingerprint` | Fingerprint changes after new entries |
| `TestSOXAuditValidatorTimestampAccuracy` | Timestamp within 1 second of recording |
| `TestSOXAuditValidatorResourceTypes` | All 8 financial resource types |
| `TestSOXAuditValidatorConcurrentWrites` | 10 concurrent goroutines, chain integrity verified |
| `TestSOXAuditValidatorMustRecordPanic` | Panic wrapper for invalid input |
| `BenchmarkSOXAuditValidatorRecordEntry` | Entry recording throughput |
| `BenchmarkSOXAuditValidatorVerifyImmutability` | 1,000-entry chain verification throughput |

**Run Tests:**
```bash
cd desktop/backend-go
go test ./internal/compliance/sox* -v -count=1
```

---

### SOX Integration Example

Recording a financial reconciliation with full before/after capture:

```go
v := NewSOXAuditValidator("your-hmac-secret-32-bytes-min", logger)

// Reconciliation: amount corrected from 1000 to 1500
before := json.RawMessage(`{"amount": 1000.00, "currency": "USD", "status": "pending"}`)
after  := json.RawMessage(`{"amount": 1500.00, "currency": "USD", "status": "approved"}`)

entry, err := v.RecordFinancialMutation(
    ctx,
    "user-finance-123",      // actor: who made the change
    "human",                  // actorType
    OperationUpdate,          // operation
    Transaction,              // resourceType
    "txn-Q1-2026-0047",       // resourceID
    "periodic_reconciliation",// reasonCode (required for SOX)
    before,                   // state before change
    after,                    // state after change
)
if err != nil {
    return fmt.Errorf("SOX recording failed: %w", err)
}

// Verify integrity of the entire trail
valid, issues := v.VerifyAuditTrailImmutability()
if !valid {
    for _, issue := range issues {
        log.Printf("SOX VIOLATION: %s", issue)
    }
}

// SOX compliance report: fingerprint of complete audit trail
fingerprint := v.ComputeAuditFingerprint()
log.Printf("SOX audit trail fingerprint: %s (total entries: %d)",
    fingerprint, v.GetEntryCount())

// Retrieve complete history for one resource (for external auditors)
history := v.GetAuditHistory(ctx, Transaction, "txn-Q1-2026-0047")
for _, e := range history {
    log.Printf("[%d] %s by %s (%s): %s",
        e.SequenceNum, e.Operation, e.Actor, e.ReasonCode, e.ChangeSummary)
}
```

---

## Compliance Rule YAML Structure

BusinessOS enforces compliance rules from two YAML files:

| File | Frameworks | Format |
|------|------------|--------|
| `config/compliance-rules.yaml` | SOC2, GDPR, SOX | Multi-document YAML |
| `config/hipaa-rules.yaml` | HIPAA | Single-document YAML array |

### compliance-rules.yaml Schema

The `compliance-rules.yaml` file uses a flat rule array at the top level:

```yaml
rules:
  - id: "framework.category.n"   # Dot-separated: framework.category.sequence
    title: "Human-readable title"
    description: |               # Optional: detailed explanation
      Multi-line description of what this rule enforces.
    condition: "field.path operator value"  # DSL condition expression
    action: "escalate|create_gap|notify|audit"
    enabled: true                # Set false to disable without removing
    severity: "critical|high|medium|low"
    framework: "SOC2|GDPR|SOX|HIPAA|CUSTOM"
```

### Rule ID Conventions

| Pattern | Example | Used By |
|---------|---------|---------|
| `hipaa.<category>.<n>` | `hipaa.ac.1` | HIPAA rules |
| `gdpr.<category>.<n>` | `gdpr.ds.1` | GDPR rules |
| `sox.<category>.<n>` | `sox.itg.1` | SOX rules |
| `soc2.<control>.<n>` | `soc2.cc6.1` | SOC2 rules |

### Condition DSL Field Paths

Fields available in rule conditions:

| Field Path | Type | Description |
|---|---|---|
| `user.role` | string | Authenticated user role |
| `user.consent_given` | bool | GDPR consent flag |
| `data.encrypted` | bool | Encryption-at-rest status |
| `data.contains_phi` | bool | HIPAA PHI flag |
| `data.contains_pii` | bool | GDPR PII flag |
| `data.retention_days` | int | Days data has been retained |
| `data.location` | string | Storage region (e.g., `"eu"`) |
| `audit_entry.signature_valid` | bool | Audit chain signature status |
| `audit_log.retention_days` | int | Audit log retention duration |
| `audit_log.missing_phi_access_entries` | bool | Missing PHI access log flag |
| `service.uptime` | float | Service uptime percentage |
| `system.measured_uptime` | float | System-measured uptime percentage |
| `financial_data.access_logged` | bool | Financial data access logged flag |
| `financial_record.has_checksum` | bool | Financial record checksum present |
| `checksum.verified` | bool | Checksum verification status |
| `change.approved_by` | string | Who approved the change |
| `change.made_by` | string | Who made the change |
| `change.requires_approval` | bool | Change approval gate flag |
| `production_change.documented` | bool | Change documentation status |
| `processor.dpa_signed` | bool | DPA signed with data processor |
| `processor.handles_data` | bool | Processor handles personal data |
| `data_processing.requires_consent` | bool | Processing requires consent |
| `data_subject_request.pending` | bool | Unresolved data subject request |
| `days_elapsed` | int | Days since data subject request received |
| `data_collected.field_count` | int | Fields actually collected |
| `data_needed.field_count` | int | Fields required for processing purpose |

### Enforcement Actions

| Action | Effect |
|--------|--------|
| `escalate` | Escalate to CISO or compliance officer with severity and context |
| `create_gap` | Open a compliance gap record for tracking and remediation |
| `notify` | Send notification to configured channels (security-team, legal-team, etc.) |
| `audit` | Write an audit log entry without blocking the operation |

### Hot Reload (Without Restart)

```bash
curl -X POST http://localhost:8001/api/compliance/reload-rules
# Response: {"status": "reloaded"}
```

Validate before reload:
```bash
yamllint config/compliance-rules.yaml
python3 scripts/validate-compliance-rules.py
bash scripts/verify-compliance-rules.sh
```

---

## All Frameworks: Run All Tests

```bash
cd desktop/backend-go
go test ./internal/compliance/... -v -count=1
```

---

## See Also

- `desktop/backend-go/internal/compliance/gdpr.go` — GDPR service implementation
- `desktop/backend-go/internal/compliance/gdpr_test.go` — GDPR test suite (32 tests)
- `desktop/backend-go/internal/compliance/sox.go` — SOX audit validator implementation
- `desktop/backend-go/internal/compliance/sox_test.go` — SOX test suite (14 tests)
- `desktop/backend-go/internal/compliance/hipaa.go` — HIPAA validator implementation
- `config/compliance-rules.yaml` — SOC2, GDPR, SOX rule definitions
- `config/hipaa-rules.yaml` — HIPAA rule definitions
- `docs/diataxis/explanation/soc2-compliance-rule-engine.md` — Rule engine internals
- **References:**
  - 45 CFR Part 164 — HIPAA Security Rule & Breach Notification
  - EU 2016/679 — General Data Protection Regulation
  - Sarbanes-Oxley Act, Section 404(b) — IT General Controls
  - NIST SP 800-88 — Guidelines for Media Sanitization
  - NIST SP 800-53 — Security and Privacy Controls
  - HHS OCR — Breach Notification (45 CFR 164.400-414)
