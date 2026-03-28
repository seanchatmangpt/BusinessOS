package compliance

import (
	"context"
	"encoding/json"
	"log/slog"
	"testing"
	"time"
)

// TestSOXAuditValidatorRecordFinancialMutation tests basic entry recording
func TestSOXAuditValidatorRecordFinancialMutation(t *testing.T) {
	validator := NewSOXAuditValidator("secret-key-at-least-32-bytes-long", slog.Default())

	before := json.RawMessage(`{"amount": 1000.00, "status": "pending"}`)
	after := json.RawMessage(`{"amount": 1500.00, "status": "approved"}`)

	entry, err := validator.RecordFinancialMutation(
		context.Background(),
		"user-123",
		"human",
		OperationUpdate,
		Transaction,
		"txn-456",
		"periodic_reconciliation",
		before,
		after,
	)

	if err != nil {
		t.Fatalf("RecordFinancialMutation failed: %v", err)
	}

	// Verify entry structure
	if entry.ID == "" {
		t.Fatal("Entry ID should not be empty")
	}
	if entry.SequenceNum != 1 {
		t.Fatalf("Expected sequence 1, got %d", entry.SequenceNum)
	}
	if entry.Actor != "user-123" {
		t.Fatalf("Expected actor user-123, got %s", entry.Actor)
	}
	if entry.Operation != OperationUpdate {
		t.Fatalf("Expected operation UPDATE, got %s", entry.Operation)
	}
	if entry.ResourceType != Transaction {
		t.Fatalf("Expected resource type Transaction, got %s", entry.ResourceType)
	}
	if entry.Status != "committed" {
		t.Fatalf("Expected status committed, got %s", entry.Status)
	}
	if entry.PreviousHash != "" {
		t.Fatal("First entry should have empty previous_hash")
	}
	if entry.DataHash == "" {
		t.Fatal("DataHash should not be empty")
	}
	if entry.Signature == "" {
		t.Fatal("Signature should not be empty")
	}
}

// TestSOXAuditValidatorMultipleEntries tests chain integrity with multiple entries
func TestSOXAuditValidatorMultipleEntries(t *testing.T) {
	validator := NewSOXAuditValidator("secret-key-at-least-32-bytes-long", slog.Default())
	ctx := context.Background()

	// Record 3 entries
	entry1, _ := validator.RecordFinancialMutation(
		ctx, "user-1", "human", OperationCreate, LedgerEntry,
		"ledger-1", "initial_entry",
		nil, json.RawMessage(`{"amount": 100.00}`),
	)

	entry2, _ := validator.RecordFinancialMutation(
		ctx, "user-2", "human", OperationUpdate, LedgerEntry,
		"ledger-1", "corrected_amount",
		json.RawMessage(`{"amount": 100.00}`),
		json.RawMessage(`{"amount": 150.00}`),
	)

	entry3, _ := validator.RecordFinancialMutation(
		ctx, "user-3", "human", OperationUpdate, LedgerEntry,
		"ledger-1", "final_reconciliation",
		json.RawMessage(`{"amount": 150.00}`),
		json.RawMessage(`{"amount": 150.00, "reconciled": true}`),
	)

	// Verify sequence numbers
	if entry1.SequenceNum != 1 {
		t.Fatalf("Expected entry1 sequence 1, got %d", entry1.SequenceNum)
	}
	if entry2.SequenceNum != 2 {
		t.Fatalf("Expected entry2 sequence 2, got %d", entry2.SequenceNum)
	}
	if entry3.SequenceNum != 3 {
		t.Fatalf("Expected entry3 sequence 3, got %d", entry3.SequenceNum)
	}

	// Verify chain links
	if entry2.PreviousHash != entry1.DataHash {
		t.Fatal("entry2 previous_hash should match entry1 data_hash")
	}
	if entry3.PreviousHash != entry2.DataHash {
		t.Fatal("entry3 previous_hash should match entry2 data_hash")
	}

	// Verify total count
	if validator.GetEntryCount() != 3 {
		t.Fatalf("Expected 3 entries, got %d", validator.GetEntryCount())
	}
}

// TestSOXAuditValidatorVerifyImmutability tests immutability verification
func TestSOXAuditValidatorVerifyImmutability(t *testing.T) {
	validator := NewSOXAuditValidator("secret-key-at-least-32-bytes-long", slog.Default())
	ctx := context.Background()

	before := json.RawMessage(`{"amount": 1000.00}`)
	after := json.RawMessage(`{"amount": 2000.00}`)

	entry, _ := validator.RecordFinancialMutation(
		ctx, "user-123", "human", OperationUpdate, Transaction,
		"txn-789", "amount_correction", before, after,
	)

	// Verify entry immutability
	isValid, msg := validator.VerifyEntryImmutability(entry.ID)
	if !isValid {
		t.Fatalf("Entry immutability verification failed: %s", msg)
	}

	// Verify chain integrity
	valid, issues := validator.VerifyAuditTrailImmutability()
	if !valid {
		t.Fatalf("Chain integrity verification failed: %v", issues)
	}
	if len(issues) != 0 {
		t.Fatalf("Expected no issues, got %v", issues)
	}
}

// TestSOXAuditValidatorChainIntegrity tests hash chain integrity
func TestSOXAuditValidatorChainIntegrity(t *testing.T) {
	validator := NewSOXAuditValidator("secret-key-at-least-32-bytes-long", slog.Default())
	ctx := context.Background()

	// Record multiple entries
	for i := 0; i < 5; i++ {
		before := json.RawMessage(`{"iteration": ` + string(rune(i)) + `}`)
		after := json.RawMessage(`{"iteration": ` + string(rune(i+1)) + `}`)

		_, _ = validator.RecordFinancialMutation(
			ctx, "user-batch", "service_account", OperationUpdate, LedgerEntry,
			"ledger-batch", "batch_processing",
			before, after,
		)
	}

	// Verify chain is intact
	valid, issues := validator.VerifyAuditTrailImmutability()
	if !valid {
		t.Fatalf("Chain integrity failed: %v", issues)
	}

	// Get all entries and verify chain
	entries := validator.GetCompleteAuditTrail()
	if len(entries) != 5 {
		t.Fatalf("Expected 5 entries, got %d", len(entries))
	}

	for i := 1; i < len(entries); i++ {
		if entries[i].PreviousHash != entries[i-1].DataHash {
			t.Fatalf("Chain broken between entry %d and %d", i-1, i)
		}
	}
}

// TestSOXAuditValidatorBeforeAfterValues tests before/after value capture
func TestSOXAuditValidatorBeforeAfterValues(t *testing.T) {
	validator := NewSOXAuditValidator("secret-key-at-least-32-bytes-long", slog.Default())
	ctx := context.Background()

	beforeData := map[string]interface{}{
		"amount":   1000.00,
		"currency": "USD",
		"status":   "pending",
	}
	afterData := map[string]interface{}{
		"amount":   1500.00,
		"currency": "USD",
		"status":   "approved",
	}

	beforeJSON, _ := json.Marshal(beforeData)
	afterJSON, _ := json.Marshal(afterData)

	entry, _ := validator.RecordFinancialMutation(
		ctx, "auditor-1", "human", OperationUpdate, Transaction,
		"txn-audit", "external_audit_adjustment",
		json.RawMessage(beforeJSON),
		json.RawMessage(afterJSON),
	)

	// Verify before/after are captured exactly
	var before map[string]interface{}
	var after map[string]interface{}
	json.Unmarshal(entry.BeforeValues, &before)
	json.Unmarshal(entry.AfterValues, &after)

	if before["amount"] != 1000.00 {
		t.Fatalf("Expected before amount 1000.00, got %v", before["amount"])
	}
	if after["amount"] != 1500.00 {
		t.Fatalf("Expected after amount 1500.00, got %v", after["amount"])
	}
	if before["status"] != "pending" {
		t.Fatalf("Expected before status pending, got %s", before["status"])
	}
	if after["status"] != "approved" {
		t.Fatalf("Expected after status approved, got %s", after["status"])
	}
}

// TestSOXAuditValidatorGetAuditHistory tests resource-specific audit history retrieval
func TestSOXAuditValidatorGetAuditHistory(t *testing.T) {
	validator := NewSOXAuditValidator("secret-key-at-least-32-bytes-long", slog.Default())
	ctx := context.Background()

	// Record entries for 2 different resources
	_, _ = validator.RecordFinancialMutation(
		ctx, "user-1", "human", OperationCreate, Transaction,
		"txn-A", "create", nil, json.RawMessage(`{"id": "txn-A"}`),
	)
	_, _ = validator.RecordFinancialMutation(
		ctx, "user-2", "human", OperationCreate, LedgerEntry,
		"ledger-B", "create", nil, json.RawMessage(`{"id": "ledger-B"}`),
	)
	_, _ = validator.RecordFinancialMutation(
		ctx, "user-3", "human", OperationUpdate, Transaction,
		"txn-A", "update", json.RawMessage(`{}`), json.RawMessage(`{}`),
	)

	// Get history for txn-A (should have 2 entries)
	historyA := validator.GetAuditHistory(ctx, Transaction, "txn-A")
	if len(historyA) != 2 {
		t.Fatalf("Expected 2 entries for txn-A, got %d", len(historyA))
	}
	if historyA[0].Actor != "user-1" {
		t.Fatalf("Expected first entry by user-1, got %s", historyA[0].Actor)
	}
	if historyA[1].Actor != "user-3" {
		t.Fatalf("Expected second entry by user-3, got %s", historyA[1].Actor)
	}

	// Get history for ledger-B (should have 1 entry)
	historyB := validator.GetAuditHistory(ctx, LedgerEntry, "ledger-B")
	if len(historyB) != 1 {
		t.Fatalf("Expected 1 entry for ledger-B, got %d", len(historyB))
	}
}

// TestSOXAuditValidatorAllOperationTypes tests all operation types
func TestSOXAuditValidatorAllOperationTypes(t *testing.T) {
	validator := NewSOXAuditValidator("secret-key-at-least-32-bytes-long", slog.Default())
	ctx := context.Background()

	operations := []struct {
		opType    OperationType
		before    json.RawMessage
		after     json.RawMessage
		shouldErr bool
	}{
		{OperationCreate, nil, json.RawMessage(`{"new": "value"}`), false},
		{OperationRead, json.RawMessage(`{"data": "read"}`), json.RawMessage(`{"data": "read"}`), false},
		{OperationUpdate, json.RawMessage(`{"old": "value"}`), json.RawMessage(`{"new": "value"}`), false},
		{OperationDelete, json.RawMessage(`{"deleted": "value"}`), json.RawMessage(`{}`), false}, // DELETE with empty object is valid
	}

	for i, tc := range operations {
		entry, err := validator.RecordFinancialMutation(
			ctx, "user-ops", "human", tc.opType, Account,
			"acct-ops", "test_operation",
			tc.before, tc.after,
		)

		if tc.shouldErr && err == nil {
			t.Fatalf("Operation %d (%s): expected error, got none", i, tc.opType)
		}
		if !tc.shouldErr && err != nil {
			t.Fatalf("Operation %d (%s): unexpected error: %v", i, tc.opType, err)
		}
		if !tc.shouldErr && entry.Operation != tc.opType {
			t.Fatalf("Operation %d: expected %s, got %s", i, tc.opType, entry.Operation)
		}
	}
}

// TestSOXAuditValidatorValidationErrors tests required parameter validation
func TestSOXAuditValidatorValidationErrors(t *testing.T) {
	validator := NewSOXAuditValidator("secret-key-at-least-32-bytes-long", slog.Default())
	ctx := context.Background()

	tests := []struct {
		name           string
		actor          string
		resourceID     string
		reasonCode     string
		afterValues    json.RawMessage
		expectedSubstr string
	}{
		{"empty_actor", "", "res-1", "reason", json.RawMessage(`{}`), "actor cannot be empty"},
		{"empty_resource_id", "user-1", "", "reason", json.RawMessage(`{}`), "resourceID cannot be empty"},
		{"empty_reason_code", "user-1", "res-1", "", json.RawMessage(`{}`), "reasonCode cannot be empty"},
		{"empty_after_values", "user-1", "res-1", "reason", nil, "afterValues required"},
	}

	for _, tc := range tests {
		_, err := validator.RecordFinancialMutation(
			ctx, tc.actor, "human", OperationCreate, Account,
			tc.resourceID, tc.reasonCode, nil, tc.afterValues,
		)

		if err == nil {
			t.Fatalf("Test %s: expected error, got none", tc.name)
		}
		if !containsSubstring(err.Error(), tc.expectedSubstr) {
			t.Fatalf("Test %s: expected substring '%s', got '%s'", tc.name, tc.expectedSubstr, err.Error())
		}
	}
}

// containsSubstring checks if str contains substr
func containsSubstring(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestSOXAuditValidatorSignatureImmutability tests signature verification
func TestSOXAuditValidatorSignatureImmutability(t *testing.T) {
	validator := NewSOXAuditValidator("secret-key-at-least-32-bytes-long", slog.Default())
	ctx := context.Background()

	entry, _ := validator.RecordFinancialMutation(
		ctx, "user-sig", "human", OperationUpdate, Transaction,
		"txn-sig", "signature_test",
		json.RawMessage(`{"amount": 1000}`),
		json.RawMessage(`{"amount": 2000}`),
	)

	// Verify original signature is valid
	isValid, msg := validator.VerifyEntryImmutability(entry.ID)
	if !isValid {
		t.Fatalf("Original entry verification failed: %s", msg)
	}

	// Attempt to tamper with entry (modifying the internal entry)
	// This simulates an attacker trying to modify an entry
	entry.AfterValues = json.RawMessage(`{"amount": 3000}`) // Tamper!

	// Verification should now fail (signature no longer matches)
	isValid, msg = validator.VerifyEntryImmutability(entry.ID)
	if isValid {
		t.Fatal("Tampered entry should fail verification, but passed")
	}
	if msg != "data hash mismatch - entry was modified" {
		t.Fatalf("Expected 'data hash mismatch', got '%s'", msg)
	}
}

// TestSOXAuditValidatorComputeAuditFingerprint tests fingerprint computation
func TestSOXAuditValidatorComputeAuditFingerprint(t *testing.T) {
	validator := NewSOXAuditValidator("secret-key-at-least-32-bytes-long", slog.Default())
	ctx := context.Background()

	// Compute fingerprint on empty audit trail
	fingerprint1 := validator.ComputeAuditFingerprint()
	if fingerprint1 != "" {
		t.Fatal("Empty audit trail should have empty fingerprint")
	}

	// Add first entry
	validator.RecordFinancialMutation(
		ctx, "user-1", "human", OperationCreate, Transaction,
		"txn-1", "create", nil, json.RawMessage(`{"id": "1"}`),
	)
	fingerprint2 := validator.ComputeAuditFingerprint()

	// Add second entry
	validator.RecordFinancialMutation(
		ctx, "user-2", "human", OperationUpdate, Transaction,
		"txn-1", "update", json.RawMessage(`{}`), json.RawMessage(`{}`),
	)
	fingerprint3 := validator.ComputeAuditFingerprint()

	// Fingerprints should be different (entries changed)
	if fingerprint2 == fingerprint3 {
		t.Fatal("Fingerprints should differ after adding entry")
	}

	// Both should be non-empty and have correct length (SHA256 = 64 hex chars)
	if len(fingerprint2) != 64 {
		t.Fatalf("Expected 64-char fingerprint, got %d", len(fingerprint2))
	}
	if len(fingerprint3) != 64 {
		t.Fatalf("Expected 64-char fingerprint, got %d", len(fingerprint3))
	}
}

// TestSOXAuditValidatorTimestampAccuracy tests timestamp recording
func TestSOXAuditValidatorTimestampAccuracy(t *testing.T) {
	validator := NewSOXAuditValidator("secret-key-at-least-32-bytes-long", slog.Default())
	ctx := context.Background()

	before := time.Now().UTC()
	entry, _ := validator.RecordFinancialMutation(
		ctx, "user-time", "human", OperationCreate, LedgerEntry,
		"ledger-time", "timestamp_test", nil, json.RawMessage(`{}`),
	)
	after := time.Now().UTC()

	// Timestamp should be within 1 second
	if entry.Timestamp.Before(before) || entry.Timestamp.After(after.Add(1*time.Second)) {
		t.Fatalf("Timestamp out of range: entry=%v, before=%v, after=%v",
			entry.Timestamp, before, after)
	}
}

// TestSOXAuditValidatorResourceTypes tests all financial resource types
func TestSOXAuditValidatorResourceTypes(t *testing.T) {
	validator := NewSOXAuditValidator("secret-key-at-least-32-bytes-long", slog.Default())
	ctx := context.Background()

	resourceTypes := []FinancialResourceType{
		LedgerEntry, Transaction, Account, JournalEntry,
		PaymentRecord, InvoiceRecord, ExpenseRecord, BudgetAlloc,
	}

	for _, rt := range resourceTypes {
		entry, err := validator.RecordFinancialMutation(
			ctx, "user-resources", "human", OperationCreate, rt,
			"res-"+string(rt), "resource_type_test", nil, json.RawMessage(`{}`),
		)

		if err != nil {
			t.Fatalf("Resource type %s: unexpected error: %v", rt, err)
		}
		if entry.ResourceType != rt {
			t.Fatalf("Resource type mismatch: expected %s, got %s", rt, entry.ResourceType)
		}
	}
}

// TestSOXAuditValidatorConcurrentWrites tests concurrent entry recording
func TestSOXAuditValidatorConcurrentWrites(t *testing.T) {
	validator := NewSOXAuditValidator("secret-key-at-least-32-bytes-long", slog.Default())
	ctx := context.Background()

	// Record 10 entries concurrently
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(index int) {
			resourceID := string(rune(48 + index)) // "0" to "9"
			_, _ = validator.RecordFinancialMutation(
				ctx, "user-concurrent", "human", OperationCreate, Transaction,
				resourceID, "concurrent_write", nil, json.RawMessage(`{}`),
			)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all entries were recorded
	if validator.GetEntryCount() != 10 {
		t.Fatalf("Expected 10 entries, got %d", validator.GetEntryCount())
	}

	// Verify chain integrity even with concurrent writes
	valid, issues := validator.VerifyAuditTrailImmutability()
	if !valid {
		t.Fatalf("Chain integrity failed with concurrent writes: %v", issues)
	}
}

// TestSOXAuditValidatorMustRecordPanic tests panic wrapper
func TestSOXAuditValidatorMustRecordPanic(t *testing.T) {
	validator := NewSOXAuditValidator("secret-key-at-least-32-bytes-long", slog.Default())
	ctx := context.Background()

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Expected panic for invalid actor, but did not panic")
		}
	}()

	// This should panic due to empty actor
	validator.MustRecordFinancialMutation(
		ctx, "", "human", OperationCreate, Transaction,
		"res-1", "reason", nil, json.RawMessage(`{}`),
	)
}

// TestSOXAuditValidatorActorTypes tests different actor types
func TestSOXAuditValidatorActorTypes(t *testing.T) {
	validator := NewSOXAuditValidator("secret-key-at-least-32-bytes-long", slog.Default())
	ctx := context.Background()

	actorTypes := []string{"human", "service_account", "agent", "system"}

	for i, actorType := range actorTypes {
		entry, err := validator.RecordFinancialMutation(
			ctx, "actor-"+actorType, actorType, OperationCreate, Account,
			"acct-"+actorType, "actor_type_test", nil, json.RawMessage(`{}`),
		)

		if err != nil {
			t.Fatalf("Actor type %s: unexpected error: %v", actorType, err)
		}
		if entry.ActorType != actorType {
			t.Fatalf("Entry %d: expected actor type %s, got %s", i, actorType, entry.ActorType)
		}
	}
}

// BenchmarkSOXAuditValidatorRecordEntry benchmarks entry recording
func BenchmarkSOXAuditValidatorRecordEntry(b *testing.B) {
	validator := NewSOXAuditValidator("secret-key-at-least-32-bytes-long", slog.Default())
	ctx := context.Background()

	before := json.RawMessage(`{"amount": 1000.00}`)
	after := json.RawMessage(`{"amount": 2000.00}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.RecordFinancialMutation(
			ctx, "user-bench", "human", OperationUpdate, Transaction,
			"txn-bench", "benchmark", before, after,
		)
	}
}

// BenchmarkSOXAuditValidatorVerifyImmutability benchmarks verification
func BenchmarkSOXAuditValidatorVerifyImmutability(b *testing.B) {
	validator := NewSOXAuditValidator("secret-key-at-least-32-bytes-long", slog.Default())
	ctx := context.Background()

	// Pre-populate with 1000 entries
	for i := 0; i < 1000; i++ {
		validator.RecordFinancialMutation(
			ctx, "user-setup", "human", OperationCreate, Transaction,
			"txn-"+string(rune(i)), "setup", nil, json.RawMessage(`{}`),
		)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.VerifyAuditTrailImmutability()
	}
}
