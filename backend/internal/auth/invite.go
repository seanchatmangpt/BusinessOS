package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	inviteTokenByteLen = 32
	inviteExpiry       = 7 * 24 * time.Hour
)

// InviteRecord is a row from the auth_invites table.
type InviteRecord struct {
	ID        string
	Email     string
	Role      string
	InvitedBy string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

// CreateInvite generates a new invite token for the given email address.
// Returns the raw (un-hashed) token that should be sent to the recipient.
// Only the SHA-256 hash is stored in the database.
func CreateInvite(ctx context.Context, pool *pgxpool.Pool, email, role, invitedByUserID string) (rawToken string, err error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Generate cryptographically random token.
	b := make([]byte, inviteTokenByteLen)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate invite token: %w", err)
	}
	rawToken = base64.RawURLEncoding.EncodeToString(b)
	tokenHash := hashInviteToken(rawToken)

	id, err := generateSecureToken(16)
	if err != nil {
		return "", err
	}

	expiresAt := time.Now().Add(inviteExpiry)

	_, err = pool.Exec(ctx, `
		INSERT INTO auth_invites (id, email, token_hash, role, invited_by, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
	`, id, email, tokenHash, role, invitedByUserID, expiresAt)
	if err != nil {
		return "", fmt.Errorf("insert invite: %w", err)
	}

	return rawToken, nil
}

// ValidateInvite looks up an invite by its raw token and returns the record
// if the invite is valid (not expired, not already used). The caller should
// call MarkInviteUsed after the user completes registration.
func ValidateInvite(ctx context.Context, pool *pgxpool.Pool, rawToken string) (*InviteRecord, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	tokenHash := hashInviteToken(rawToken)

	var inv InviteRecord
	err := pool.QueryRow(ctx, `
		SELECT id, email, role, invited_by, expires_at, used_at, created_at
		FROM auth_invites
		WHERE token_hash = $1
	`, tokenHash).Scan(
		&inv.ID,
		&inv.Email,
		&inv.Role,
		&inv.InvitedBy,
		&inv.ExpiresAt,
		&inv.UsedAt,
		&inv.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("invite not found")
	}

	if inv.UsedAt != nil {
		return nil, fmt.Errorf("invite already used")
	}
	if time.Now().After(inv.ExpiresAt) {
		return nil, fmt.Errorf("invite expired")
	}

	return &inv, nil
}

// MarkInviteUsed stamps the invite as consumed by the registering user.
func MarkInviteUsed(ctx context.Context, pool *pgxpool.Pool, rawToken string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	tokenHash := hashInviteToken(rawToken)

	tag, err := pool.Exec(ctx, `
		UPDATE auth_invites SET used_at = NOW()
		WHERE token_hash = $1 AND used_at IS NULL
	`, tokenHash)
	if err != nil {
		return fmt.Errorf("mark invite used: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("invite already consumed or not found")
	}
	return nil
}

// hashInviteToken returns the hex-encoded SHA-256 of the raw invite token.
// Only the hash is persisted; the raw token is sent to the user once and
// never stored.
func hashInviteToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}
