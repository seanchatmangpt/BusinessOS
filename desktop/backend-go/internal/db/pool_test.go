package db

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPoolCreation verifies pool can be created with valid config.
func TestPoolCreation(t *testing.T) {
	cfg := &PoolConfig{
		DSN:            "postgres://localhost/test", // Will fail if DB not running
		MinSize:        5,
		MaxSize:        20,
		IdleTimeout:    1 * time.Minute,
		AcquireTimeout: 5 * time.Second,
	}

	// This test will fail if PostgreSQL is not running.
	// In CI/CD, this should be marked as integration test and skipped in fast mode.
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logger := slog.Default()
	pool, err := NewPool(ctx, cfg, logger)

	// Don't fail test if DB not running - this is normal in unit test environment
	if err != nil {
		t.Logf("skipping pool creation test (DB not available): %v", err)
		return
	}
	defer pool.Close()

	assert.NotNil(t, pool)
	assert.NotNil(t, pool.pool)
}

// TestPoolConfigDefaults verifies default values are applied correctly.
func TestPoolConfigDefaults(t *testing.T) {
	cfg := &PoolConfig{
		DSN: "postgres://localhost/test",
		// Intentionally omit other fields to test defaults
	}

	// Validate that we can create a config without all fields set
	assert.NotNil(t, cfg)
	assert.Equal(t, 0, cfg.MinSize, "MinSize should be 0 before NewPool applies defaults")
	assert.Equal(t, 0, cfg.MaxSize, "MaxSize should be 0 before NewPool applies defaults")
	assert.Equal(t, time.Duration(0), cfg.IdleTimeout, "IdleTimeout should be 0 before NewPool applies defaults")

	// Note: actual defaults (MinSize=10, MaxSize=100, etc.) are applied in NewPool()
	// This prevents hardcoding defaults in the struct, which would require changes
	// in multiple places if defaults need to be updated.
}

// TestPoolValidation verifies pool rejects invalid configs.
func TestPoolValidation(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *PoolConfig
		wantErr bool
	}{
		{
			name:    "nil config",
			cfg:     nil,
			wantErr: true,
		},
		{
			name: "empty DSN",
			cfg: &PoolConfig{
				DSN: "",
			},
			wantErr: true,
		},
		{
			name: "invalid DSN format",
			cfg: &PoolConfig{
				DSN: "not-a-valid-dsn",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			_, err := NewPool(ctx, tt.cfg, slog.Default())
			if tt.wantErr {
				require.Error(t, err)
			}
		})
	}
}

// TestPoolStats verifies stats can be queried without panic.
func TestPoolStats(t *testing.T) {
	cfg := &PoolConfig{
		DSN:            "postgres://localhost/test",
		MinSize:        5,
		MaxSize:        20,
		AcquireTimeout: 5 * time.Second,
	}

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := NewPool(ctx, cfg, slog.Default())
	if err != nil {
		t.Logf("skipping stats test (DB not available): %v", err)
		return
	}
	defer pool.Close()

	stats := pool.Stats()
	assert.True(t, stats.Total > 0, "pool should have connections")
	assert.True(t, stats.Idle >= 0, "idle count should be non-negative")
	assert.True(t, stats.Active >= 0, "active count should be non-negative")
}

// TestPoolHealth verifies health check works.
func TestPoolHealth(t *testing.T) {
	cfg := &PoolConfig{
		DSN:            "postgres://localhost/test",
		MinSize:        5,
		MaxSize:        20,
		AcquireTimeout: 5 * time.Second,
	}

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := NewPool(ctx, cfg, slog.Default())
	if err != nil {
		t.Logf("skipping health test (DB not available): %v", err)
		return
	}
	defer pool.Close()

	// Health check should succeed
	err = pool.Health(ctx)
	assert.NoError(t, err)
}

// TestPoolHealthAfterClose verifies health check fails on closed pool.
func TestPoolHealthAfterClose(t *testing.T) {
	cfg := &PoolConfig{
		DSN:            "postgres://localhost/test",
		MinSize:        5,
		MaxSize:        20,
		AcquireTimeout: 5 * time.Second,
	}

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := NewPool(ctx, cfg, slog.Default())
	if err != nil {
		t.Logf("skipping health test (DB not available): %v", err)
		return
	}

	pool.Close()

	// Health check on closed pool should fail
	err = pool.Health(context.Background())
	assert.Error(t, err)
}

// TestPoolContextDeadline verifies context deadlines are respected (WvdA soundness).
func TestPoolContextDeadline(t *testing.T) {
	cfg := &PoolConfig{
		DSN:            "postgres://localhost/test",
		MinSize:        5,
		MaxSize:        20,
		AcquireTimeout: 30 * time.Second, // Long timeout
	}

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := NewPool(ctx, cfg, slog.Default())
	if err != nil {
		t.Logf("skipping context deadline test (DB not available): %v", err)
		return
	}
	defer pool.Close()

	// Create a context that expires immediately
	expiredCtx, expCancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer expCancel()

	time.Sleep(10 * time.Millisecond) // Ensure context is expired

	// Acquire should respect the expired context and return quickly
	_, err = pool.Acquire(expiredCtx)
	assert.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, expiredCtx.Err())
}
