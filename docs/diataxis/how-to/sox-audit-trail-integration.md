# How To: Integrate SOX Audit Trail into Financial Services

**Guide Type:** How-To (Task-Focused)
**Level:** Intermediate (assumes knowledge of Go services)
**Version:** 1.1.0
**Last Updated:** 2026-03-27

---

## Overview

This guide shows how to integrate the BusinessOS audit trail components into your financial services to ensure SOX 404(b) compliance.

The audit trail pipeline has three cooperating layers:

| Layer | Package | Responsibility |
|-------|---------|---------------|
| **Request logging** | `internal/middleware/audit_log.go` | Log every HTTP request with user ID, path, status, latency |
| **A2A hash-chain logger** | `internal/middleware/audit.go` | Cryptographically chained entries for agent-to-agent calls |
| **Compliance service** | `internal/services/compliance_service.go` | Aggregate audit trail from OSA, verify chain integrity, gap analysis |

All log output passes through the **`SanitizedLogger`** (`internal/logging/sanitizer.go`) to strip PII and credentials before any entry is persisted or forwarded.

**What You'll Learn:**
- How to wire the `AuditLogger` and `AuditSensitiveAccess` Gin middleware
- How to create a `HashChainLogger` and call `LogA2ACall()`
- How to use `SafeLogFields`, `MaskSessionID`, and `MaskIP` before writing audit data
- How to verify hash-chain integrity with `VerifyChainIntegrity()`
- How to use `ComplianceService.GetAuditTrail()` and `VerifyAuditChain()`
- How audit events surface as OpenTelemetry spans via the global tracer

---

## Prerequisites

- Go 1.24+ installed
- Access to BusinessOS backend-go codebase
- Familiarity with Go services and `context.Context`
- Understanding of your business domain (what is a "transaction", "ledger entry", etc.)
- `SOX_HMAC_SECRET` environment variable set (see Step 1)

---

## Step 1: Set the HMAC Secret

**Location:** `.env` file (never commit this file)

```bash
# Generate a secure secret:
openssl rand -hex 32

SOX_HMAC_SECRET=a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2w3x4y5z6a7b8c9d0e
```

The HMAC secret is used by `HashChainLogger` to sign every audit entry: `HMAC-SHA256(previousHash + dataHash, secret)`. Entries signed with different secrets will fail `VerifyChainIntegrity()`, so do not rotate the secret without a migration plan.

**Validation** (in your service init):

```go
package main

import (
    "log/slog"
    "os"
)

func validateSOXConfig() {
    hmacSecret := os.Getenv("SOX_HMAC_SECRET")
    if hmacSecret == "" {
        slog.Error("SOX_HMAC_SECRET not set - audit trail will not be available")
        return
    }
    if len(hmacSecret) < 32 {
        slog.Warn("SOX_HMAC_SECRET is too short - should be at least 32 bytes for security")
    }
    slog.Info("SOX audit trail HMAC secret configured")
}
```

---

## Step 2: Mount the Audit Middleware on Your Router

**Location:** wherever you configure your Gin router (e.g., `cmd/server/bootstrap.go`)

Two middleware functions are available in `internal/middleware/audit_log.go`:

| Function | When to use |
|----------|------------|
| `AuditLogger()` | Apply globally — logs every request (user, method, path, status, latency) |
| `AuditSensitiveAccess(resourceType)` | Apply to route groups serving PII or confidential data |

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/rhl/businessos-backend/internal/middleware"
)

func setupRouter() *gin.Engine {
    r := gin.New()

    // Global audit log — applies to all routes
    r.Use(middleware.AuditLogger())

    // Sensitive data groups — extra detail for SOX-regulated resources
    financial := r.Group("/api/transactions")
    financial.Use(middleware.AuditSensitiveAccess("transaction"))
    {
        financial.POST("", transactionHandler.Create)
        financial.PUT("/:id/approve", transactionHandler.Approve)
        financial.PUT("/:id/reject", transactionHandler.Reject)
    }

    return r
}
```

`AuditLogger()` emits a structured `slog.Info("AUDIT", ...)` entry after each request containing `timestamp`, `user_id`, `method`, `path`, `status`, `ip`, `user_agent`, and `latency_ms`.

`AuditSensitiveAccess` emits a `slog.Warn("AUDIT_SENSITIVE_ACCESS", ...)` entry with `user_id`, `resource_type`, `resource_id`, `method`, `path`, `status`, `ip`, and `timestamp`.

---

## Step 3: Initialize the Hash-Chain Logger

**Location:** Service initialization (e.g., `cmd/server/bootstrap.go` or `internal/services/init.go`)

```go
package main

import (
    "log/slog"
    "os"

    "github.com/rhl/businessos-backend/internal/middleware"
)

var auditLogger *middleware.HashChainLogger

func initAuditLogger() {
    hmacSecret := os.Getenv("SOX_HMAC_SECRET")
    if hmacSecret == "" {
        slog.Error("SOX_HMAC_SECRET not set - hash-chain audit logger unavailable")
        return
    }
    auditLogger = middleware.NewHashChainLogger(hmacSecret)
    slog.Info("SOX hash-chain audit logger initialized")
}
```

The `HashChainLogger` is thread-safe via in-memory append. It chains entries using:

- `DataHash = SHA256(agent + action + resourceType + resourceID + timestamp)`
- `Signature = HMAC-SHA256(previousHash + dataHash, secret)`

---

## Step 4: Sanitize Data Before Writing Audit Entries

**Package:** `internal/logging/sanitizer.go`

All data written to the audit trail must be sanitized to remove PII, credentials, and session tokens before logging. The `logging` package provides the following utilities:

| Function | Purpose |
|----------|---------|
| `logging.SafeLogFields(map[string]interface{})` | Redacts known-sensitive field names (`password`, `token`, `api_key`, `session_id`, `authorization`, `cookie`, etc.) |
| `logging.MaskSessionID(sessionID string)` | Shows first 8 chars of a session ID, masks the rest |
| `logging.MaskIP(ip string)` | Masks the last two octets of an IPv4 address (e.g., `192.168.xxx.xxx`) |
| `logging.SanitizeURL(rawURL string)` | Redacts all query parameters and path segments following `/session/` |

The `SanitizedLogger` automatically applies regex-based masking for:
- UUIDs (session IDs)
- Bearer tokens and JWT strings
- Email addresses
- IP addresses
- Secrets embedded in key=value patterns

**Example — sanitize before recording an audit entry:**

```go
import (
    "github.com/rhl/businessos-backend/internal/logging"
    "github.com/rhl/businessos-backend/internal/middleware"
)

func recordAuditedAction(
    auditLogger *middleware.HashChainLogger,
    agentID, action, resourceType, resourceID string,
    fields map[string]interface{},
) {
    // Sanitize any field map before it reaches the audit trail
    safeFields := logging.SafeLogFields(fields)

    // Log the safe fields using slog (picked up by AuditLogger middleware)
    slog.Info("financial_mutation",
        "agent", agentID,
        "action", action,
        "resource_type", resourceType,
        "resource_id", resourceID,
        "details", safeFields,
    )

    // Record in the cryptographic hash chain
    _, err := auditLogger.LogA2ACall(agentID, action, resourceType, resourceID, 0.85)
    if err != nil {
        slog.Error("failed to record audit entry",
            "error", err,
            "resource_id", resourceID,
        )
    }
}
```

**Example — mask session and IP in log output:**

```go
slog.Info("AUDIT",
    "session_id", logging.MaskSessionID(sessionID),   // "a3f9b2c1********"
    "client_ip",  logging.MaskIP(clientIP),            // "192.168.xxx.xxx"
    "path",       logging.SanitizeURL(requestURL),     // query params redacted
)
```

---

## Step 5: Record Financial Mutations via `LogA2ACall`

**Location:** In your service layer, wherever financial state changes

```go
package services

import (
    "context"
    "encoding/json"
    "fmt"
    "log/slog"
    "time"

    "github.com/rhl/businessos-backend/internal/logging"
    "github.com/rhl/businessos-backend/internal/middleware"
    "github.com/rhl/businessos-backend/internal/models"
)

type TransactionService struct {
    repo      models.TransactionRepository
    audit     *middleware.HashChainLogger
    logger    *slog.Logger
}

func NewTransactionService(
    repo models.TransactionRepository,
    audit *middleware.HashChainLogger,
    logger *slog.Logger,
) *TransactionService {
    return &TransactionService{repo: repo, audit: audit, logger: logger}
}

// CreateTransaction creates a new transaction and records a hash-chain audit entry.
func (s *TransactionService) CreateTransaction(
    ctx context.Context,
    agentID string,
    amount float64,
    accountID string,
    description string,
) (*models.Transaction, error) {
    txn := &models.Transaction{
        ID:          uuid.New().String(),
        Amount:      amount,
        AccountID:   accountID,
        Description: description,
        Status:      "pending",
        CreatedAt:   time.Now(),
    }

    if err := s.repo.Create(ctx, txn); err != nil {
        return nil, fmt.Errorf("failed to create transaction: %w", err)
    }

    // Sanitize fields before audit logging
    auditFields := logging.SafeLogFields(map[string]interface{}{
        "amount":     amount,
        "account_id": accountID,
        "status":     "pending",
    })
    s.logger.InfoContext(ctx, "transaction_created", "fields", auditFields)

    // Record in hash chain (snScore 0.85 = high-confidence agent action)
    _, err := s.audit.LogA2ACall(
        agentID,             // who performed the action
        "create_transaction", // what was done
        "transaction",        // resource type
        txn.ID,               // resource ID
        0.85,                 // Signal/Noise score
    )
    if err != nil {
        s.logger.ErrorContext(ctx, "failed to record audit entry",
            "error", err,
            "transaction_id", txn.ID,
        )
        // NOTE: In production, escalate this to an alert — audit failure is a compliance event
    }

    return txn, nil
}

// ApproveTransaction approves a pending transaction and records a hash-chain audit entry.
func (s *TransactionService) ApproveTransaction(
    ctx context.Context,
    agentID string,
    transactionID string,
) (*models.Transaction, error) {
    txn, err := s.repo.GetByID(ctx, transactionID)
    if err != nil {
        return nil, fmt.Errorf("transaction not found: %w", err)
    }

    txn.Status = "approved"
    txn.UpdatedAt = time.Now()

    if err := s.repo.Update(ctx, txn); err != nil {
        return nil, fmt.Errorf("failed to update transaction: %w", err)
    }

    _, err = s.audit.LogA2ACall(
        agentID,
        "approve_transaction",
        "transaction",
        transactionID,
        0.90, // Higher confidence for approval actions
    )
    if err != nil {
        s.logger.ErrorContext(ctx, "failed to record approval audit entry",
            "error", err,
            "transaction_id", transactionID,
        )
    }

    return txn, nil
}
```

**`LogA2ACall` parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `agent` | `string` | Who performed the action (agent ID or user ID) |
| `action` | `string` | What was done (e.g., `"create_transaction"`, `"approve_transaction"`) |
| `resourceType` | `string` | Category of affected resource (e.g., `"transaction"`, `"ledger_entry"`) |
| `resourceID` | `string` | Identifier of the specific resource affected |
| `snScore` | `float64` | Signal/Noise score `[0.0–1.0]`; drives `GovernanceTier` assignment |

---

## Step 6: Mount the A2A Audit Middleware for Agent Routes

For A2A agent endpoints, use `A2AAuditMiddleware` to automatically log all non-GET requests:

```go
import "github.com/rhl/businessos-backend/internal/middleware"

auditChain := middleware.NewHashChainLogger(os.Getenv("SOX_HMAC_SECRET"))

a2aGroup := r.Group("/api/integrations/a2a")
a2aGroup.Use(middleware.A2AAuditMiddleware(auditChain))
{
    a2aGroup.POST("/crm/deals", crmHandler.CreateDeal)
    a2aGroup.PUT("/crm/deals/:id", crmHandler.UpdateDeal)
    // GET routes are excluded by A2AAuditMiddleware (read-only, no mutation)
}
```

The middleware extracts the `X-Agent-ID` header, derives `action` and `resourceType` from the HTTP method and URL path, and reads the optional `sn_score` context value set by upstream middleware.

---

## Step 7: Verify Hash-Chain Integrity

**Periodic Verification** (run hourly or daily via scheduler):

```go
package services

import (
    "fmt"
    "log/slog"

    "github.com/rhl/businessos-backend/internal/middleware"
)

// VerifyAuditChainIntegrity checks the entire in-memory hash chain for tampering.
func VerifyAuditChainIntegrity(auditLogger *middleware.HashChainLogger) error {
    valid, issues := auditLogger.VerifyChainIntegrity()

    if !valid {
        slog.Error("SOX audit chain integrity violation",
            "issues_count", len(issues),
        )
        for i, issue := range issues {
            slog.Error("chain issue", "number", i, "detail", issue)
        }
        // ESCALATE: Alert security team
        return fmt.Errorf("audit chain integrity compromised: %d violations", len(issues))
    }

    slog.Info("SOX audit chain integrity verified")
    return nil
}
```

`VerifyChainIntegrity()` checks three things for every entry:

1. `DataHash` recomputed from fields matches stored `DataHash`
2. `Signature` recomputed from `HMAC(previousHash + dataHash)` matches stored `Signature`
3. `PreviousHash` matches the `DataHash` of the immediately preceding entry

**Register Periodic Verification** (in your scheduler):

```go
scheduler.Every(1).Hours().Do(func() {
    if err := VerifyAuditChainIntegrity(auditLogger); err != nil {
        alertSecurityTeam(err)
    }
})
```

---

## Step 8: Use the Compliance Service for Cross-System Audit Trail

The `ComplianceService` (`internal/services/compliance_service.go`) aggregates audit data from OSA and BusinessOS, verifies the hash chain, and surfaces gaps.

**Initialize:**

```go
import "github.com/rhl/businessos-backend/internal/services"

complianceSvc := services.NewComplianceService(
    os.Getenv("OSA_BASE_URL"),  // e.g., "http://localhost:8089"
    slog.Default(),
)
```

**Retrieve audit trail for a session:**

```go
trail, err := complianceSvc.GetAuditTrail(ctx, services.AuditTrailParams{
    SessionID: sessionID,
    From:      time.Now().Add(-24 * time.Hour),
    To:        time.Now(),
    Limit:     100,
    Offset:    0,
})
if err != nil {
    return fmt.Errorf("get audit trail: %w", err)
}
// trail.Entries is []AuditEntry{ID, SessionID, Timestamp, Action, Actor, ToolName, Details, Hash, PrevHash}
```

**Verify chain integrity via the compliance service:**

```go
result, err := complianceSvc.VerifyAuditChain(ctx, sessionID)
if err != nil {
    return fmt.Errorf("verify audit chain: %w", err)
}
// result.Verified bool
// result.Entries  int
// result.Issues   []string
// result.MerkleRoot string
```

**Evaluate a specific audit entry against compliance rules:**

```go
entry := services.AuditEntry{
    ID:        "entry-uuid",
    SessionID: sessionID,
    Timestamp: time.Now(),
    Action:    "approve_transaction",
    Actor:     "manager-001",
}
if err := complianceSvc.EvaluateAuditEvent(ctx, entry, "manager"); err != nil {
    slog.Error("compliance rule triggered", "error", err)
}
```

**Run SOX gap analysis:**

```go
resp, err := complianceSvc.VerifyCompliance(ctx, services.ComplianceVerifyRequest{
    WorkspaceID: workspaceID,
    Framework:   "SOX",
})
// resp.Status              — "compliant" | "partial" | "non_compliant"
// resp.FindingsCount       — number of open gaps
// resp.RemediationProgress — fraction of gaps resolved
// resp.Gaps                — []ComplianceGap (ID, Framework, Control, Description, Severity, Status)
```

---

## Step 9: OpenTelemetry Integration

BusinessOS initializes a global OTEL tracer at startup (`internal/observability/tracer.go`) with service name `"businessos"`. Audit-related operations can be wrapped in spans so they appear in Jaeger.

**Initialize the tracer** (already done in `cmd/server/bootstrap.go`):

```go
import "github.com/rhl/businessos-backend/internal/observability"

tp, err := observability.InitTracer(ctx, os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
if err != nil {
    slog.Error("tracer init failed", "error", err)
}
defer observability.ShutdownTracer(ctx, tp)
```

**Wrap audit-critical operations in spans:**

```go
import "go.opentelemetry.io/otel"

func (s *TransactionService) CreateTransactionWithSpan(
    ctx context.Context,
    agentID string,
    amount float64,
    accountID string,
) (*models.Transaction, error) {
    tracer := otel.Tracer("businessos")
    ctx, span := tracer.Start(ctx, "sox.audit.create_transaction")
    defer span.End()

    span.SetAttributes(
        attribute.String("sox.agent_id", agentID),
        attribute.String("sox.resource_type", "transaction"),
        attribute.Float64("sox.sn_score", 0.85),
    )

    txn, err := s.CreateTransaction(ctx, agentID, amount, accountID, "")
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return nil, err
    }

    span.SetAttributes(attribute.String("sox.resource_id", txn.ID))
    span.SetStatus(codes.Ok, "")
    return txn, nil
}
```

**Expected Jaeger span** (visible at `http://localhost:16686`, service `businessos`):

```json
{
  "service": "businessos",
  "span_name": "sox.audit.create_transaction",
  "status": "ok",
  "attributes": {
    "sox.agent_id": "agent-001",
    "sox.resource_type": "transaction",
    "sox.resource_id": "txn-uuid",
    "sox.sn_score": 0.85
  }
}
```

---

## Step 10: Standard Action Codes

Create a constants file to standardize the `action` argument passed to `LogA2ACall`:

**File:** `internal/compliance/reason_codes.go`

```go
package compliance

// StandardActionCode represents a standardized business action for SOX audit entries.
type StandardActionCode string

const (
    // Regular Operations
    ActionCreateTransaction StandardActionCode = "create_transaction"
    ActionUpdateTransaction StandardActionCode = "update_transaction"
    ActionDeleteTransaction StandardActionCode = "delete_transaction"

    // Approval Process
    ActionManagerApproval   StandardActionCode = "manager_approval"
    ActionExecutiveApproval StandardActionCode = "executive_approval"
    ActionRejection         StandardActionCode = "rejection"

    // Reconciliation
    ActionPeriodEndReconciliation StandardActionCode = "period_end_reconciliation"
    ActionBankReconciliation      StandardActionCode = "bank_reconciliation"
    ActionInternalAudit           StandardActionCode = "internal_audit"
    ActionExternalAuditAdjustment StandardActionCode = "external_audit_adjustment"

    // Error Correction
    ActionDataEntryError   StandardActionCode = "data_entry_error_correction"
    ActionSystemError      StandardActionCode = "system_error_correction"
    ActionDuplicateRemoval StandardActionCode = "duplicate_removal"

    // Policy/System Changes
    ActionPolicyChange        StandardActionCode = "policy_change"
    ActionSystemMigration     StandardActionCode = "system_migration"
    ActionConfigurationUpdate StandardActionCode = "configuration_update"

    // Investigation
    ActionFraudInvestigation      StandardActionCode = "fraud_investigation"
    ActionComplianceInvestigation StandardActionCode = "compliance_investigation"
)
```

**Usage:**

```go
_, err = auditLogger.LogA2ACall(
    agentID,
    string(compliance.ActionManagerApproval),
    "transaction",
    txnID,
    0.90,
)
```

---

## Step 11: Test Your Integration

**Unit Test Example:**

```go
func TestTransactionServiceAuditTrail(t *testing.T) {
    // Setup
    auditLogger := middleware.NewHashChainLogger("secret-key-at-least-32-bytes-long")
    repo := &mockTransactionRepo{}
    service := services.NewTransactionService(repo, auditLogger, slog.Default())

    // Create transaction
    txn, err := service.CreateTransaction(
        context.Background(),
        "agent-001",
        1000.00,
        "acct-456",
        "Test transaction",
    )
    assert.NoError(t, err)
    assert.NotEmpty(t, txn.ID)

    // Verify audit entry was created in hash chain
    entries := auditLogger.QueryAuditTrail("transaction", txn.ID)
    assert.Equal(t, 1, len(entries))
    assert.Equal(t, "agent-001", entries[0].Agent)
    assert.Equal(t, "create_transaction", entries[0].Action)

    // Verify chain integrity
    valid, issues := auditLogger.VerifyChainIntegrity()
    assert.True(t, valid)
    assert.Empty(t, issues)
}

func TestSanitizedFieldsAreNotLogged(t *testing.T) {
    fields := map[string]interface{}{
        "amount":   1000.00,
        "password": "super-secret",
        "token":    "eyJhbGciOiJIUzI1NiJ9...",
        "account":  "acct-123",
    }

    safe := logging.SafeLogFields(fields)

    assert.Equal(t, "[REDACTED]", safe["password"])
    assert.Equal(t, "[REDACTED]", safe["token"])
    assert.Equal(t, 1000.00, safe["amount"])    // Not a sensitive field
    assert.Equal(t, "acct-123", safe["account"]) // Not a sensitive field
}
```

---

## Troubleshooting

### Issue: HMAC Secret Not Set

**Error:**
```
SOX_HMAC_SECRET not set - audit trail will not be available
```

**Solution:**
```bash
openssl rand -hex 32 > /tmp/secret.txt
echo "SOX_HMAC_SECRET=$(cat /tmp/secret.txt)" >> .env
```

### Issue: Chain Integrity Violation

**Error reported by `VerifyChainIntegrity()`:**
```
entry <uuid> signature invalid
entry <uuid> chain link broken
```

**Causes:**
- HMAC secret changed after entries were written
- Entries were modified after creation (tamper detected)

**Action:**
1. Stop processing immediately
2. Alert security team
3. Identify which entries are compromised via `issues` slice
4. Do not rotate the HMAC secret until after investigation

### Issue: OSA Audit Trail Unavailable

**Error from `ComplianceService.GetAuditTrail()`:**
```
fetch audit trail from OSA: OSA unavailable: ...
```

**Cause:** OSA service at `OSA_BASE_URL` is not reachable.

**Action:**
- `ComplianceService` caches the last successful response per session
- If no cache exists, the call returns an error — do not swallow it
- Check `OSA_BASE_URL` env var and OSA service health at `/health`

### Issue: Sensitive Data Appears in Logs

**Symptom:** Audit log entries contain raw session IDs, tokens, or email addresses.

**Fix:**
- Wrap all field maps with `logging.SafeLogFields(fields)` before passing to `slog`
- Use `logging.MaskSessionID(id)` when logging session identifiers
- Use `logging.MaskIP(ip)` when logging client IP addresses
- Use `logging.SanitizeURL(url)` when logging request URLs with query parameters

---

## Security Checklist

Before deploying to production:

- [ ] `SOX_HMAC_SECRET` is at least 32 bytes
- [ ] `SOX_HMAC_SECRET` is stored in environment variable (never hardcoded or committed)
- [ ] `SOX_HMAC_SECRET` rotated at least annually (with migration plan for existing entries)
- [ ] `AuditLogger()` middleware mounted globally on Gin router
- [ ] `AuditSensitiveAccess(resourceType)` applied to all financial route groups
- [ ] `A2AAuditMiddleware` applied to all agent-facing route groups
- [ ] All field maps sanitized with `logging.SafeLogFields()` before logging
- [ ] `VerifyChainIntegrity()` scheduled to run at least hourly
- [ ] `ComplianceService.VerifyAuditChain()` called nightly for cross-system verification
- [ ] Audit trail access restricted to compliance/audit roles
- [ ] PostgreSQL retention policy in place (7-year archival for SOX 404)
- [ ] Backup/recovery procedures tested
- [ ] OTEL spans emitted for all high-value mutation operations (visible in Jaeger)

---

## Next Steps

1. Identify all financial mutation points in your codebase
2. Wire `HashChainLogger` into each service and call `LogA2ACall()` on every mutation
3. Mount `AuditLogger()` and `AuditSensitiveAccess()` on all applicable route groups
4. Instrument high-value operations with OTEL spans (see Step 9)
5. Add `VerifyAuditChain()` to your nightly compliance reporting endpoint
6. Test with SOX auditor using the evidence collected by `ComplianceService.CollectEvidence()`

---

*Version 1.1.0 | 2026-03-27 | SOX 404(b) Integration Guide*
