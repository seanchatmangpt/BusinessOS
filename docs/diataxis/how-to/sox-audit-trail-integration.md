# How To: Integrate SOX Audit Trail into Financial Services

**Guide Type:** How-To (Task-Focused)
**Level:** Intermediate (assumes knowledge of Go services)
**Version:** 1.0.0
**Last Updated:** 2026-03-26

---

## Overview

This guide shows how to integrate the `SOXAuditValidator` into your BusinessOS financial services to ensure SOX 404(b) compliance.

**What You'll Learn:**
- Where to call `RecordFinancialMutation()`
- How to capture before/after values
- How to verify audit trail integrity
- How to handle concurrent mutations
- How to set up HMAC secret management

---

## Prerequisites

- Go 1.24+ installed
- Access to BusinessOS backend-go codebase
- Familiarity with Go services and context.Context
- Understanding of your business domain (what is a "transaction", "ledger entry", etc.)

---

## Step 1: Initialize the SOX Audit Validator

**Location:** In your service initialization code (e.g., `main.go` or `services/init.go`)

```go
package main

import (
    "log/slog"
    "os"
    "github.com/rhl/businessos-backend/internal/compliance"
)

// Initialize SOX audit trail validator (create once, reuse globally)
var soxValidator *compliance.SOXAuditValidator

func init() {
    hmacSecret := os.Getenv("SOX_HMAC_SECRET")
    if hmacSecret == "" {
        slog.Error("SOX_HMAC_SECRET not set - audit trail will not be available")
        return
    }

    if len(hmacSecret) < 32 {
        slog.Warn("SOX_HMAC_SECRET is too short - should be at least 32 bytes for security")
    }

    soxValidator = compliance.NewSOXAuditValidator(
        hmacSecret,
        slog.Default(),
    )
    slog.Info("SOX audit trail initialized")
}
```

**Environment Variable Setup** (in `.env`):

```bash
# Generate a secure secret:
# openssl rand -hex 32

SOX_HMAC_SECRET=a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2w3x4y5z6a7b8c9d0e
```

---

## Step 2: Create a Transaction Service with Audit Trail

**Example:** `internal/services/transaction_service.go`

```go
package services

import (
    "context"
    "encoding/json"
    "fmt"
    "log/slog"
    "time"

    "github.com/rhl/businessos-backend/internal/compliance"
    "github.com/rhl/businessos-backend/internal/models"
)

type TransactionService struct {
    repo      models.TransactionRepository
    audit     *compliance.SOXAuditValidator
    logger    *slog.Logger
}

func NewTransactionService(
    repo models.TransactionRepository,
    audit *compliance.SOXAuditValidator,
    logger *slog.Logger,
) *TransactionService {
    return &TransactionService{
        repo:   repo,
        audit:  audit,
        logger: logger,
    }
}

// CreateTransaction creates a new transaction and records audit entry
func (s *TransactionService) CreateTransaction(
    ctx context.Context,
    userID string,
    amount float64,
    accountID string,
    description string,
) (*models.Transaction, error) {
    // 1. Create the transaction in the database
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

    // 2. Record in SOX audit trail (required for compliance)
    afterValues, _ := json.Marshal(txn)

    _, err := s.audit.RecordFinancialMutation(
        ctx,
        userID,                              // Who made the change
        "human",                             // Actor type
        compliance.OperationCreate,          // What operation
        compliance.Transaction,              // What resource
        txn.ID,                              // Resource ID
        "new_transaction_created",           // Why (business justification)
        nil,                                 // No before state for CREATE
        json.RawMessage(afterValues),        // After state
    )

    if err != nil {
        // Log audit failure but don't fail the transaction creation
        s.logger.ErrorContext(ctx, "failed to record audit entry",
            "error", err,
            "transaction_id", txn.ID,
        )
        // NOTE: In production, you may want to escalate this to an alert
    }

    return txn, nil
}

// UpdateTransaction updates transaction and records audit entry
func (s *TransactionService) UpdateTransaction(
    ctx context.Context,
    userID string,
    transactionID string,
    newAmount *float64,
    newStatus *string,
    reason string,
) (*models.Transaction, error) {
    // 1. Get current state (before)
    before, err := s.repo.GetByID(ctx, transactionID)
    if err != nil {
        return nil, fmt.Errorf("transaction not found: %w", err)
    }

    beforeValues, _ := json.Marshal(before)

    // 2. Apply updates
    if newAmount != nil {
        before.Amount = *newAmount
    }
    if newStatus != nil {
        before.Status = *newStatus
    }
    before.UpdatedAt = time.Now()

    if err := s.repo.Update(ctx, before); err != nil {
        return nil, fmt.Errorf("failed to update transaction: %w", err)
    }

    // 3. Record in SOX audit trail (before and after states required for UPDATE)
    afterValues, _ := json.Marshal(before)

    _, err = s.audit.RecordFinancialMutation(
        ctx,
        userID,                              // Who made the change
        "human",                             // Actor type
        compliance.OperationUpdate,          // What operation
        compliance.Transaction,              // What resource
        transactionID,                       // Resource ID
        reason,                              // Why (user-provided reason)
        beforeValues,                        // Before state (required for UPDATE)
        json.RawMessage(afterValues),        // After state
    )

    if err != nil {
        s.logger.ErrorContext(ctx, "failed to record audit entry",
            "error", err,
            "transaction_id", transactionID,
        )
    }

    return before, nil
}

// ApproveTransaction approves a pending transaction
func (s *TransactionService) ApproveTransaction(
    ctx context.Context,
    userID string,
    transactionID string,
) (*models.Transaction, error) {
    return s.UpdateTransaction(
        ctx,
        userID,
        transactionID,
        nil,
        stringPtr("approved"),
        "manager_approval",  // Standard reason code for approvals
    )
}

// RejectTransaction rejects a transaction with reason
func (s *TransactionService) RejectTransaction(
    ctx context.Context,
    userID string,
    transactionID string,
    rejectionReason string,
) (*models.Transaction, error) {
    return s.UpdateTransaction(
        ctx,
        userID,
        transactionID,
        nil,
        stringPtr("rejected"),
        fmt.Sprintf("rejection: %s", rejectionReason),
    )
}

// Helper function
func stringPtr(s string) *string {
    return &s
}
```

---

## Step 3: Capture Before/After Values Correctly

**Best Practices:**

### Full Object Snapshot

```go
// GOOD: Capture complete object state
type TransactionSnapshot struct {
    ID          string    `json:"id"`
    Amount      float64   `json:"amount"`
    AccountID   string    `json:"account_id"`
    Status      string    `json:"status"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

before, _ := json.Marshal(&TransactionSnapshot{
    ID:        txn.ID,
    Amount:    1000.00,
    AccountID: "acct-123",
    Status:    "pending",
    CreatedAt: time.Now(),
})

after, _ := json.Marshal(&TransactionSnapshot{
    ID:        txn.ID,
    Amount:    1500.00,  // Changed
    AccountID: "acct-123",
    Status:    "approved",  // Changed
    CreatedAt: time.Now(),
})
```

### Sparse Changes (Only Modified Fields)

```go
// ACCEPTABLE: Capture only fields that changed (space-efficient)
beforeChanges := json.RawMessage(`{"amount": 1000.00, "status": "pending"}`)
afterChanges := json.RawMessage(`{"amount": 1500.00, "status": "approved"}`)

s.audit.RecordFinancialMutation(ctx, userID, "human",
    compliance.OperationUpdate, compliance.Transaction, txnID,
    "amount_and_status_update", beforeChanges, afterChanges)
```

### What NOT to Do

```go
// WRONG: Don't capture nil for before/after
// - For CREATE: before MUST be nil or empty
// - For UPDATE: both before AND after are REQUIRED
// - For DELETE: before is REQUIRED, after can be empty

// WRONG: Don't use non-deterministic JSON (will fail signature verification)
// Use json.Marshal to ensure consistent ordering
json.RawMessage(`{status:"pending", amount:1000}`)  // ❌ Wrong order

// CORRECT: json.Marshal handles ordering
data, _ := json.Marshal(struct {
    Amount float64 `json:"amount"`
    Status string  `json:"status"`
}{1000, "pending"})  // ✅ Consistent order
```

---

## Step 4: Integrate with Handlers

**Example:** `internal/handlers/transaction_handler.go`

```go
package handlers

import (
    "log/slog"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/rhl/businessos-backend/internal/services"
)

type TransactionHandler struct {
    txnService *services.TransactionService
    logger     *slog.Logger
}

func NewTransactionHandler(
    txnService *services.TransactionService,
    logger *slog.Logger,
) *TransactionHandler {
    return &TransactionHandler{
        txnService: txnService,
        logger:     logger,
    }
}

// POST /api/transactions
func (h *TransactionHandler) CreateTransaction(c *gin.Context) {
    type CreateRequest struct {
        Amount      float64 `json:"amount" binding:"required,gt=0"`
        AccountID   string  `json:"account_id" binding:"required"`
        Description string  `json:"description" binding:"required"`
    }

    var req CreateRequest
    if err := c.BindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Extract user ID from JWT token (already authenticated by middleware)
    userID, _ := c.Get("user_id")
    userIDStr := userID.(string)

    // Create transaction (automatically audited)
    txn, err := h.txnService.CreateTransaction(
        c.Request.Context(),
        userIDStr,
        req.Amount,
        req.AccountID,
        req.Description,
    )

    if err != nil {
        h.logger.ErrorContext(c.Request.Context(), "failed to create transaction",
            "error", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create transaction"})
        return
    }

    c.JSON(http.StatusCreated, txn)
}

// PUT /api/transactions/:id/approve
func (h *TransactionHandler) ApproveTransaction(c *gin.Context) {
    transactionID := c.Param("id")
    userID, _ := c.Get("user_id")
    userIDStr := userID.(string)

    // Approve transaction (automatically audited)
    txn, err := h.txnService.ApproveTransaction(
        c.Request.Context(),
        userIDStr,
        transactionID,
    )

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, txn)
}
```

---

## Step 5: Verify Audit Trail Integrity

**Periodic Verification** (e.g., run hourly or daily):

```go
package services

import (
    "log/slog"
    "time"

    "github.com/rhl/businessos-backend/internal/compliance"
)

type ComplianceService struct {
    audit  *compliance.SOXAuditValidator
    logger *slog.Logger
}

// VerifyAuditTrailIntegrity checks entire audit trail for tampering
func (s *ComplianceService) VerifyAuditTrailIntegrity() error {
    valid, issues := s.audit.VerifyAuditTrailImmutability()

    if !valid {
        s.logger.Error("SOX audit trail integrity violation",
            "issues_count", len(issues),
        )
        for i, issue := range issues {
            s.logger.Error("issue", "number", i, "detail", issue)
        }
        // ESCALATE: Alert security team
        return fmt.Errorf("audit trail integrity compromised: %d violations", len(issues))
    }

    s.logger.Info("SOX audit trail integrity verified",
        "total_entries", s.audit.GetEntryCount(),
    )
    return nil
}

// ComputeAuditFingerprint returns fingerprint for compliance reporting
func (s *ComplianceService) ComputeAuditFingerprint() string {
    return s.audit.ComputeAuditFingerprint()
}

// GetAuditHistory retrieves all changes to a resource
func (s *ComplianceService) GetAuditHistory(
    resourceType compliance.FinancialResourceType,
    resourceID string,
) []*compliance.SOXAuditEntry {
    return s.audit.GetAuditHistory(context.Background(), resourceType, resourceID)
}
```

**Register Periodic Verification** (in your scheduler):

```go
// In your main.go or scheduler initialization
scheduler.Every(1).Hours().Do(func() {
    if err := complianceService.VerifyAuditTrailIntegrity(); err != nil {
        alertSecurityTeam(err)
    }
})
```

---

## Step 6: Standard Reason Codes

Create a constants file for standardized reason codes:

**File:** `internal/compliance/reason_codes.go`

```go
package compliance

// StandardReasonCode represents a standardized business justification
type StandardReasonCode string

const (
    // Regular Operations
    CreateNew              StandardReasonCode = "new_transaction_created"
    UpdateAmount           StandardReasonCode = "amount_correction"
    UpdateStatus           StandardReasonCode = "status_change"

    // Approval Process
    ManagerApproval        StandardReasonCode = "manager_approval"
    ExecutiveApproval      StandardReasonCode = "executive_approval"
    RejectionByManager     StandardReasonCode = "rejection_by_manager"

    // Reconciliation
    PeriodEndReconciliation StandardReasonCode = "periodic_reconciliation"
    BankReconciliation      StandardReasonCode = "bank_reconciliation"
    InternalAudit           StandardReasonCode = "internal_audit"
    ExternalAuditAdjustment StandardReasonCode = "external_audit_adjustment"

    // Error Correction
    DataEntryError          StandardReasonCode = "data_entry_error"
    SystemError             StandardReasonCode = "system_error_correction"
    DuplicateRemoval        StandardReasonCode = "duplicate_removal"

    // Policy/System Changes
    PolicyChange            StandardReasonCode = "policy_change"
    SystemMigration         StandardReasonCode = "system_migration"
    ConfigurationUpdate     StandardReasonCode = "configuration_update"

    // Investigation/Fraud
    FraudInvestigation      StandardReasonCode = "fraud_investigation"
    ComplianceInvestigation StandardReasonCode = "compliance_investigation"
)
```

**Usage:**

```go
// Type-safe reason codes
s.audit.RecordFinancialMutation(
    ctx, userID, "human", compliance.OperationUpdate,
    compliance.Transaction, txnID,
    string(compliance.ManagerApproval),  // Using constant
    beforeValues, afterValues,
)
```

---

## Step 7: Handle Concurrent Writes Safely

The `SOXAuditValidator` is thread-safe (uses `sync.RWMutex`), so concurrent mutations are safe:

```go
// SAFE: Multiple goroutines can call RecordFinancialMutation concurrently
go func() {
    s.audit.RecordFinancialMutation(ctx, "user-1", "human", ...)
}()

go func() {
    s.audit.RecordFinancialMutation(ctx, "user-2", "human", ...)
}()

// Both entries will be sequentially numbered and chained correctly
// No race conditions or data corruption
```

---

## Step 8: Test Your Integration

**Unit Test Example:**

```go
func TestTransactionServiceAuditTrail(t *testing.T) {
    // Setup
    audit := compliance.NewSOXAuditValidator("secret-key-at-least-32-bytes-long", slog.Default())
    repo := &mockTransactionRepo{}
    service := services.NewTransactionService(repo, audit, slog.Default())

    // Create transaction
    txn, err := service.CreateTransaction(
        context.Background(),
        "user-123",
        1000.00,
        "acct-456",
        "Test transaction",
    )
    assert.NoError(t, err)

    // Verify audit entry was created
    history := audit.GetAuditHistory(context.Background(),
        compliance.Transaction, txn.ID)
    assert.Equal(t, 1, len(history))
    assert.Equal(t, "user-123", history[0].Actor)

    // Verify chain integrity
    valid, issues := audit.VerifyAuditTrailImmutability()
    assert.True(t, valid)
    assert.Empty(t, issues)
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
# Generate and set in .env
openssl rand -hex 32 > /tmp/secret.txt
echo "SOX_HMAC_SECRET=$(cat /tmp/secret.txt)" >> .env
```

### Issue: Audit Entry Signature Invalid

**Cause:** HMAC secret changed between entries

**Solution:**
- Rotate secret carefully with timestamp-based backup
- All entries must use same secret for chain verification

### Issue: Audit Trail Integrity Violation

**Cause:** Entry was modified after creation

**Action:**
1. Stop processing
2. Alert security team
3. Investigate who had access
4. Consider re-rotating HMAC secret

---

## Security Checklist

Before deploying to production:

- [ ] HMAC secret is >32 bytes
- [ ] HMAC secret stored in environment variable (never hardcoded)
- [ ] HMAC secret rotated at least annually
- [ ] All financial mutations call `RecordFinancialMutation()`
- [ ] Before/after values captured exactly (not modified)
- [ ] Reason codes standardized and documented
- [ ] Weekly integrity verification scheduled
- [ ] Audit trail access restricted to compliance/audit roles
- [ ] PostgreSQL retention policy in place (7-year archival)
- [ ] Backup/recovery procedures tested

---

## Next Steps

1. Identify all financial mutation points in your codebase
2. Update services to call `RecordFinancialMutation()`
3. Add audit trail verification to your health checks
4. Set up compliance reporting endpoint (exports audit history)
5. Test with SOX auditor

---

*Version 1.0.0 | 2026-03-26 | SOX 404(b) Integration Guide*
