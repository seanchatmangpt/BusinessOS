# HIPAA Compliance Rule Validator — Reference

**Location:** `desktop/backend-go/internal/compliance/hipaa.go`
**Test Suite:** `desktop/backend-go/internal/compliance/hipaa_test.go` (18 tests)
**Compliance Standard:** 45 CFR Part 164 (HIPAA Security Rule & Breach Notification)

---

## Overview

The HIPAA Compliance Rule Validator enforces Protected Health Information (PHI) security requirements across BusinessOS. It validates five compliance categories:

1. **Access Control** — Authorization roles for PHI access
2. **Encryption** — TLS 1.2+ in transit, AES-256 at rest
3. **Audit Logging** — All PHI accesses logged with timestamp/user/action
4. **Data Retention** — PHI deleted after 6-year retention period
5. **Breach Notification** — Automatic alert on unencrypted transmission

---

## Core Types

### HIPAARuleValidator

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

### PHIAccessEvent

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

### BreachNotification

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

### ComplianceMetrics

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

## Authorized Roles

Valid HIPAA authorization roles (case-sensitive):

| Role | Permissions | 45 CFR Reference |
|------|-------------|------------------|
| `hipaa_admin` | read, write, delete, access | 164.308(a)(4) |
| `hipaa_user` | read, write, access | 164.308(a)(4) |
| `hipaa_auditor` | read, access | 164.312(b) |
| `phi_viewer` | read, access | 164.502(b) |
| `phi_editor` | read, write, access | 164.502(b) |

---

## Encryption Requirements

### Transit (In-Flight)

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

### At Rest

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

## Data Retention (6 Years)

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

## Audit Logging

All PHI access must be logged with:
- **User ID** — Who accessed
- **Resource ID** — What was accessed (patient/record)
- **Action** — Operation type (read/write/delete)
- **Timestamp** — When accessed (UTC)
- **Outcome** — Success/denied/error

### Log Once

```go
v.LogPHIAccess(ctx, PHIAccessEvent{
    UserID:     "doctor1",
    Action:     "read",
    ResourceID: "patient123",
    Outcome:    "success",
    EncryptedTx: true,
})
```

### Verify Later

```go
found, err := v.ValidateAuditLogging(ctx, "doctor1", "patient123")
if found {
    // Access found in audit trail
}
```

### Retrieve Audit Log

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

## Breach Detection

Automatic detection of unencrypted PHI transmission.

### Detection

```go
// If TLS version < 1.2, breach is detected and callbacks invoked
if err := v.DetectBreach(ctx, tls.VersionTLS11, "patient_ssn_123-45-6789"); err != nil {
    // Breach detected — error returned, callbacks called
}
```

### Callbacks

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

## 45 CFR Mapping

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

## Metrics & Monitoring

### Get Current Metrics

```go
metrics := v.GetMetrics()
fmt.Printf("Total Violations: %d\n", metrics.TotalViolations)
fmt.Printf("  Access: %d\n", metrics.AccessViolations)
fmt.Printf("  Encryption: %d\n", metrics.EncryptionViolations)
fmt.Printf("  Audit: %d\n", metrics.AuditViolations)
fmt.Printf("  Retention: %d\n", metrics.RetentionViolations)
fmt.Printf("  Breaches: %d\n", metrics.BreachNotifications)
```

### Integrate with Prometheus

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

## Test Coverage

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
cd BusinessOS/desktop/backend-go
go test ./internal/compliance/hipaa* -v -count=1
```

---

## Integration Example

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

// Nurse updates record
if err := v.ValidateAccessControl(ctx, "nurse_jones", "write"); err != nil {
    return fmt.Errorf("write denied: %w", err)
}
v.LogPHIAccess(ctx, PHIAccessEvent{
    UserID:     "nurse_jones",
    Action:     "write",
    ResourceID: "patient_id_12345",
    Outcome:    "success",
})

// Auditor reviews
log, _ := v.GetAuditLog(ctx)
for _, event := range log {
    fmt.Printf("✓ %s: %s %s\n", event.Timestamp, event.UserID, event.Action)
}

// Verify retention
if err := v.ValidateRetention(ctx, "patient_id_12345"); err != nil {
    return fmt.Errorf("retention violation: %w", err)
}

// After 6 years, delete
v.DeletePHIData("patient_id_12345")

// Monitor metrics
metrics := v.GetMetrics()
if metrics.TotalViolations > 0 {
    log.Printf("WARNING: %d compliance violations", metrics.TotalViolations)
}
```

---

## Configuration

Default configuration (hardcoded, can be extended):

```go
Validator Configuration:
  Retention Days:    2190 (6 years)
  TLS Min Version:   0x0303 (TLS 1.2)
  Encryption Algo:   AES-256
  Audit Max Entries: 100,000 (oldest evicted if exceeded)
```

---

## Thread Safety

All operations are thread-safe via `sync.RWMutex`:

- **Read operations** (validation, querying) use RLock
- **Write operations** (registration, logging) use Lock
- **Atomic metrics** use `sync.atomic.Int64` (wait-free)

Safe for concurrent use:

```go
// Safe from multiple goroutines
go func() {
    for {
        v.LogPHIAccess(ctx, event)
        time.Sleep(100 * time.Millisecond)
    }
}()

go func() {
    for {
        _, _ = v.ValidateAuditLogging(ctx, "user", "resource")
        time.Sleep(100 * time.Millisecond)
    }
}()
```

---

## References

- **45 CFR Part 164** — HIPAA Security Rule & Breach Notification
- **NIST SP 800-88** — Guidelines for Media Sanitization
- **NIST SP 800-53** — Security and Privacy Controls
- **HHS OCR** — Breach Notification (45 CFR 164.400-414)
