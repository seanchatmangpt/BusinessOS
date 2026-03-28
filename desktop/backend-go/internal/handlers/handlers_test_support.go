package handlers

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/config"
)

// NewHandlersForTest creates a minimal Handlers instance suitable for unit
// tests. Only the fields accessed by the handler under test need to be set;
// all others are left at their zero values.
func NewHandlersForTest(pool *pgxpool.Pool, cfg *config.Config) *Handlers {
	return &Handlers{
		pool: pool,
		cfg:  cfg,
	}
}
