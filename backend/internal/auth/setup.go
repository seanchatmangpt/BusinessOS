package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SetupStatus describes the first-boot state of the application.
type SetupStatus struct {
	// NeedsSetup is true when no users exist yet and the mode requires setup.
	NeedsSetup bool
	// HasUsers is true when at least one user row exists in the database.
	HasUsers bool
	// Mode is the currently active auth mode.
	Mode AuthMode
}

// CheckSetupStatus queries the database to determine whether first-boot setup
// is still needed. In single-user mode setup is never required — the owner is
// auto-created instead.
func CheckSetupStatus(ctx context.Context, pool *pgxpool.Pool, mode AuthMode) (*SetupStatus, error) {
	// Single-user mode never needs a setup page.
	if mode == AuthModeSingle {
		return &SetupStatus{NeedsSetup: false, HasUsers: true, Mode: mode}, nil
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var count int
	err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM "user"`).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("count users: %w", err)
	}

	return &SetupStatus{
		NeedsSetup: count == 0,
		HasUsers:   count > 0,
		Mode:       mode,
	}, nil
}
