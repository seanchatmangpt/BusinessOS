// Package db provides PostgreSQL connection pooling and management.
// Uses pgx for type-safe, prepared-statement database access with proper
// timeout and error handling per Armstrong fault tolerance standards.
package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PoolConfig holds configuration for the connection pool.
type PoolConfig struct {
	DSN            string        // PostgreSQL connection string
	MinSize        int           // Minimum pool size
	MaxSize        int           // Maximum pool size
	IdleTimeout    time.Duration // Idle connection timeout
	AcquireTimeout time.Duration // Max time to acquire a connection
	MaxConnAge     time.Duration // Max lifetime per connection
}

// PoolStats holds pool statistics.
type PoolStats struct {
	Active    int
	Idle      int
	Total     int
	AvgWaitMs float64
}

// Pool wraps pgxpool with Armstrong-compliant error handling and timeouts.
type Pool struct {
	pool   *pgxpool.Pool
	config *PoolConfig
	logger *slog.Logger
}

// NewPool creates a new PostgreSQL connection pool.
// Returns error if connection cannot be established within AcquireTimeout.
func NewPool(ctx context.Context, config *PoolConfig, logger *slog.Logger) (*Pool, error) {
	if config == nil {
		return nil, fmt.Errorf("pool config is required")
	}

	if config.DSN == "" {
		return nil, fmt.Errorf("database DSN is required")
	}

	// Apply defaults
	if config.MinSize <= 0 {
		config.MinSize = 10
	}
	if config.MaxSize <= 0 {
		config.MaxSize = 100
	}
	if config.MaxSize < config.MinSize {
		config.MaxSize = config.MinSize * 2
	}
	if config.IdleTimeout == 0 {
		config.IdleTimeout = 5 * time.Minute
	}
	if config.AcquireTimeout == 0 {
		config.AcquireTimeout = 30 * time.Second
	}
	if config.MaxConnAge == 0 {
		config.MaxConnAge = 30 * time.Minute
	}

	// Build pgxpool config
	poolConfig, err := pgxpool.ParseConfig(config.DSN)
	if err != nil {
		return nil, fmt.Errorf("parse pool config: %w", err)
	}

	poolConfig.MinConns = int32(config.MinSize)
	poolConfig.MaxConns = int32(config.MaxSize)
	poolConfig.MaxConnIdleTime = config.IdleTimeout
	poolConfig.MaxConnLifetime = config.MaxConnAge

	// Create pool with timeout (WvdA soundness: all blocking ops have timeout)
	ctx, cancel := context.WithTimeout(ctx, config.AcquireTimeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	// Test connectivity (fail fast if database is down)
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	if logger == nil {
		logger = slog.Default()
	}

	logger.InfoContext(ctx, "connection pool initialized",
		slog.Int("min_size", config.MinSize),
		slog.Int("max_size", config.MaxSize),
		slog.Duration("idle_timeout", config.IdleTimeout),
		slog.Duration("max_age", config.MaxConnAge))

	return &Pool{
		pool:   pool,
		config: config,
		logger: logger,
	}, nil
}

// Acquire gets a connection from the pool with timeout enforcement.
// Returns error if context is cancelled or timeout exceeded.
// Armstrong compliance: respects context deadline, propagates timeout errors.
func (p *Pool) Acquire(ctx context.Context) (interface{}, error) {
	if p.pool == nil {
		return nil, fmt.Errorf("pool is closed")
	}

	// Create a child context with acquire timeout (in case parent timeout is longer)
	acqCtx, cancel := context.WithTimeout(ctx, p.config.AcquireTimeout)
	defer cancel()

	conn, err := p.pool.Acquire(acqCtx)
	if err != nil {
		// Propagate all errors: deadline exceeded, cancelled, or connection error
		return nil, fmt.Errorf("acquire connection: %w", err)
	}

	return conn, nil
}

// Close closes all connections in the pool.
// Blocks until all connections are returned and closed.
func (p *Pool) Close() error {
	if p.pool == nil {
		return nil
	}
	p.pool.Close()
	p.logger.Info("connection pool closed")
	return nil
}

// Stats returns current pool statistics.
func (p *Pool) Stats() PoolStats {
	if p.pool == nil {
		return PoolStats{}
	}

	stat := p.pool.Stat()
	return PoolStats{
		Active:    int(stat.AcquiredConns()),
		Idle:      int(stat.IdleConns()),
		Total:     int(stat.TotalConns()),
		AvgWaitMs: 0, // pgx doesn't expose per-operation wait time
	}
}

// Note: Query, QueryRow, and Exec methods are provided by the pgxpool.Pool
// directly, so they don't need to be wrapped here. Code using the pool should
// access them via pool.pool directly or use type assertions for error handling.

// Health returns health status of the pool.
// Returns error if pool is unhealthy or context deadline exceeded.
func (p *Pool) Health(ctx context.Context) error {
	if p.pool == nil {
		return fmt.Errorf("pool is closed")
	}

	// Create child context with short timeout for health check
	healthCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := p.pool.Ping(healthCtx); err != nil {
		return fmt.Errorf("pool health check failed: %w", err)
	}

	stat := p.pool.Stat()
	if stat.AcquiredConns() == 0 && stat.TotalConns() == 0 {
		return fmt.Errorf("no connections in pool")
	}

	return nil
}
