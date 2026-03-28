package db

import (
	"context"
	"fmt"
	"time"
)

// PoolConn is the minimal interface that an underlying database connection
// must satisfy so that PooledConnection can call Ping.
type PoolConn interface {
	// Ping verifies that the connection is still alive.
	Ping(ctx context.Context) error
	// Close releases the underlying connection back to its pool or closes it.
	Close()
}

// ConnFactory creates PoolConn instances on demand.
type ConnFactory interface {
	Acquire(ctx context.Context) (PoolConn, error)
}

// PooledConnection wraps a PoolConn and satisfies the Connection interface.
// IsValid() calls Ping with a 2-second timeout so that dead connections are
// surfaced rather than silently reused.
type PooledConnection struct {
	id      string
	conn    PoolConn
	factory ConnFactory
}

// NewPooledConnection acquires a connection from factory and wraps it.
func NewPooledConnection(id string, factory ConnFactory) (*PooledConnection, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	conn, err := factory.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("pool: acquire connection %q: %w", id, err)
	}

	return &PooledConnection{id: id, conn: conn, factory: factory}, nil
}

// NewPooledConnectionWithError is like NewPooledConnection but the factory is
// pre-configured to return a Ping error, used in tests to exercise the unhealthy
// code path.
func NewPooledConnectionWithError(id string, factory ConnFactory) (*PooledConnection, error) {
	return NewPooledConnection(id, factory)
}

// ID returns the connection identifier.
func (c *PooledConnection) ID() string { return c.id }

// Close releases the underlying connection.
func (c *PooledConnection) Close() error {
	if c.conn != nil {
		c.conn.Close()
	}
	return nil
}

// IsValid calls Ping with a 2-second timeout. Returns false if the connection
// is nil, the context times out, or Ping returns any error.
func (c *PooledConnection) IsValid() bool {
	if c.conn == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return c.conn.Ping(ctx) == nil
}

// ---------------------------------------------------------------------------
// MockConnFactory — test helper
// ---------------------------------------------------------------------------

// MockConnFactory produces mock PoolConns for unit tests.
// pingErr is the error that every mock connection's Ping will return.
type MockConnFactory struct {
	pingErr      error
	acquireCalls int
}

// NewMockConnFactory creates a MockConnFactory. Pass nil for a healthy factory.
func NewMockConnFactory(pingErr error) *MockConnFactory {
	return &MockConnFactory{pingErr: pingErr}
}

// Acquire creates a mock connection and increments the call counter.
func (f *MockConnFactory) Acquire(_ context.Context) (PoolConn, error) {
	f.acquireCalls++
	return &mockPoolConn{pingErr: f.pingErr}, nil
}

// AcquireCalls returns how many times Acquire has been called.
func (f *MockConnFactory) AcquireCalls() int { return f.acquireCalls }

// mockPoolConn is a test-only PoolConn.
type mockPoolConn struct {
	pingErr error
}

func (m *mockPoolConn) Ping(_ context.Context) error { return m.pingErr }
func (m *mockPoolConn) Close()                        {}
