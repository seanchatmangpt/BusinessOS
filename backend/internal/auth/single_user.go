package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	singleUserEmail = "owner@localhost"
	singleUserName  = "Owner"
	singleUserID    = "single-user-owner"

	// SingleUserSessionToken is a permanent, fixed token used in single-user
	// mode. Because there is no multi-user context, rotation is unnecessary.
	// The token is stored in the DB on first boot and reused on every request.
	singleUserSessionTokenKey = "single-user-session"
)

// SingleUserSession holds the session token that is auto-created on first
// boot and injected into every request in single-user mode.
type SingleUserSession struct {
	UserID string
	Token  string
}

// EnsureSingleUser creates the default owner user and a permanent session if
// they do not already exist. Safe to call on every startup; all operations are
// idempotent (INSERT … ON CONFLICT DO NOTHING).
func EnsureSingleUser(ctx context.Context, pool *pgxpool.Pool) (*SingleUserSession, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// Create the owner user if absent.
	_, err = tx.Exec(ctx, `
		INSERT INTO "user" (id, name, email, "emailVerified", "createdAt", "updatedAt")
		VALUES ($1, $2, $3, true, NOW(), NOW())
		ON CONFLICT (id) DO NOTHING
	`, singleUserID, singleUserName, singleUserEmail)
	if err != nil {
		return nil, fmt.Errorf("upsert single user: %w", err)
	}

	// Create the credential account entry (no password needed for single mode).
	accountID := singleUserID + "-account"
	_, err = tx.Exec(ctx, `
		INSERT INTO account (id, "userId", "accountId", "providerId", "createdAt", "updatedAt")
		VALUES ($1, $2, $3, 'single', NOW(), NOW())
		ON CONFLICT (id) DO NOTHING
	`, accountID, singleUserID, singleUserID)
	if err != nil {
		return nil, fmt.Errorf("upsert single account: %w", err)
	}

	// Look up any existing permanent session for this user.
	var existingToken string
	err = tx.QueryRow(ctx, `
		SELECT token FROM session
		WHERE "userId" = $1 AND "expiresAt" > NOW()
		ORDER BY "createdAt" DESC
		LIMIT 1
	`, singleUserID).Scan(&existingToken)

	if err == nil && existingToken != "" {
		// A valid session already exists — commit (no-op) and return it.
		if commitErr := tx.Commit(ctx); commitErr != nil {
			return nil, fmt.Errorf("commit: %w", commitErr)
		}
		return &SingleUserSession{UserID: singleUserID, Token: existingToken}, nil
	}

	// Generate a new permanent session token (100-year expiry).
	token, err := generateSecureToken(32)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	sessionID, err := generateSecureToken(16)
	if err != nil {
		return nil, fmt.Errorf("generate session id: %w", err)
	}

	expiresAt := time.Now().Add(100 * 365 * 24 * time.Hour) // ~100 years

	_, err = tx.Exec(ctx, `
		INSERT INTO session (id, "userId", token, "expiresAt", "createdAt", "updatedAt")
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		ON CONFLICT (token) DO NOTHING
	`, sessionID, singleUserID, token, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("insert session: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	slog.Info("single-user mode: owner user and session initialised",
		"user_id", singleUserID,
		"email", singleUserEmail,
	)

	return &SingleUserSession{UserID: singleUserID, Token: token}, nil
}

// generateSecureToken returns a URL-safe base64 token of the given byte length.
func generateSecureToken(byteLen int) (string, error) {
	b := make([]byte, byteLen)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("crypto/rand.Read: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
