package database

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// MigrationRecord represents a migration tracked in the database
type MigrationRecord struct {
	Version   string
	AppliedAt time.Time
	Checksum  string
}

// MigrationFile represents a migration file on disk
type MigrationFile struct {
	Version      string
	Name         string
	Path         string
	RollbackPath string
	Checksum     string
	IsRollback   bool
}

// MigrationRunner manages database schema evolution
type MigrationRunner struct {
	pool           *pgxpool.Pool
	migrationsDir  string
	schemaTable    string
	logger         *slog.Logger
}

// NewMigrationRunner creates a new migration runner
func NewMigrationRunner(pool *pgxpool.Pool, migrationsDir string, logger *slog.Logger) *MigrationRunner {
	if logger == nil {
		logger = slog.Default()
	}
	return &MigrationRunner{
		pool:          pool,
		migrationsDir: migrationsDir,
		schemaTable:   "schema_migrations",
		logger:        logger,
	}
}

// InitSchema creates the schema_migrations table if it doesn't exist
func (m *MigrationRunner) InitSchema(ctx context.Context) error {
	createTableSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			checksum VARCHAR(64) NOT NULL,
			duration_ms INTEGER,
			status VARCHAR(50) DEFAULT 'success'
		);

		CREATE INDEX IF NOT EXISTS idx_%s_applied_at ON %s(applied_at DESC);
	`, m.schemaTable, m.schemaTable, m.schemaTable)

	_, err := m.pool.Exec(ctx, createTableSQL)
	if err != nil {
		m.logger.Error("Failed to initialize schema migrations table", slog.String("error", err.Error()))
		return fmt.Errorf("initialize schema migrations table: %w", err)
	}

	m.logger.Info("Schema migrations table initialized", slog.String("table", m.schemaTable))
	return nil
}

// MigrateUp applies all pending migrations in order
func (m *MigrationRunner) MigrateUp(ctx context.Context) error {
	if err := m.InitSchema(ctx); err != nil {
		return err
	}

	// Load all migration files
	migrations, err := m.loadMigrations()
	if err != nil {
		return fmt.Errorf("load migrations: %w", err)
	}

	if len(migrations) == 0 {
		m.logger.Info("No migrations found")
		return nil
	}

	// Get applied migrations
	applied, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("get applied migrations: %w", err)
	}

	appliedMap := make(map[string]*MigrationRecord)
	for _, a := range applied {
		appliedMap[a.Version] = a
	}

	// Apply pending migrations
	totalApplied := 0
	for _, mig := range migrations {
		if mig.IsRollback {
			continue // Skip rollback files
		}

		if record, exists := appliedMap[mig.Version]; exists {
			// Migration already applied, verify checksum
			if record.Checksum != mig.Checksum {
				m.logger.Warn("Migration file modified after application",
					slog.String("version", mig.Version),
					slog.String("expected_checksum", record.Checksum),
					slog.String("actual_checksum", mig.Checksum))
				return fmt.Errorf("migration %s: checksum mismatch (file modified after application)", mig.Version)
			}
			m.logger.Info("Migration already applied", slog.String("version", mig.Version))
			continue
		}

		// Apply migration
		m.logger.Info("Applying migration", slog.String("version", mig.Version), slog.String("name", mig.Name))

		start := time.Now()
		migCopy := mig
		if err := m.applyMigration(ctx, &migCopy); err != nil {
			m.logger.Error("Failed to apply migration",
				slog.String("version", mig.Version),
				slog.String("error", err.Error()))
			return fmt.Errorf("apply migration %s: %w", mig.Version, err)
		}

		duration := time.Since(start)
		totalApplied++

		m.logger.Info("Migration applied successfully",
			slog.String("version", mig.Version),
			slog.Duration("duration", duration))
	}

	if totalApplied == 0 {
		m.logger.Info("All migrations already applied")
	} else {
		m.logger.Info("Migrations completed", slog.Int("count", totalApplied))
	}

	return nil
}

// MigrateDown rolls back the last N migrations in reverse order
func (m *MigrationRunner) MigrateDown(ctx context.Context, count int) error {
	if count <= 0 {
		return fmt.Errorf("count must be > 0")
	}

	// Get applied migrations in reverse order
	applied, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("get applied migrations: %w", err)
	}

	if len(applied) == 0 {
		m.logger.Info("No migrations to rollback")
		return nil
	}

	// Sort in reverse order for rollback
	sort.Slice(applied, func(i, j int) bool {
		return applied[i].AppliedAt.After(applied[j].AppliedAt)
	})

	// Rollback count migrations
	if count > len(applied) {
		count = len(applied)
	}

	for i := 0; i < count; i++ {
		rec := applied[i]

		// Load migration to find rollback file
		migrations, err := m.loadMigrations()
		if err != nil {
			return fmt.Errorf("load migrations: %w", err)
		}

		var migFile *MigrationFile
		for _, mig := range migrations {
			if mig.Version == rec.Version {
				migFile = &mig
				break
			}
		}

		if migFile == nil {
			return fmt.Errorf("migration %s not found on disk", rec.Version)
		}

		if migFile.RollbackPath == "" {
			m.logger.Warn("No rollback file found for migration",
				slog.String("version", rec.Version),
				slog.String("name", rec.Version))
			return fmt.Errorf("migration %s: no rollback file (rollback_*.sql not found)", rec.Version)
		}

		// Execute rollback
		m.logger.Info("Rolling back migration", slog.String("version", rec.Version))

		start := time.Now()
		if err := m.rollbackMigration(ctx, migFile); err != nil {
			m.logger.Error("Failed to rollback migration",
				slog.String("version", rec.Version),
				slog.String("error", err.Error()))
			return fmt.Errorf("rollback migration %s: %w", rec.Version, err)
		}

		duration := time.Since(start)

		// Remove from schema_migrations
		if err := m.recordRollback(ctx, rec.Version); err != nil {
			return fmt.Errorf("record rollback %s: %w", rec.Version, err)
		}

		m.logger.Info("Migration rolled back successfully",
			slog.String("version", rec.Version),
			slog.Duration("duration", duration))
	}

	m.logger.Info("Rollback completed", slog.Int("count", count))
	return nil
}

// GetVersion returns the current applied schema version
func (m *MigrationRunner) GetVersion(ctx context.Context) (string, error) {
	applied, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return "", err
	}

	if len(applied) == 0 {
		return "0", nil
	}

	// Return highest version
	highest := applied[0]
	for _, a := range applied {
		if a.Version > highest.Version {
			highest = a
		}
	}

	return highest.Version, nil
}

// VerifyChecksum verifies that a migration file hasn't been modified
func (m *MigrationRunner) VerifyChecksum(ctx context.Context, version string, filePath string) (bool, error) {
	// Get recorded checksum
	applied, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return false, err
	}

	var recorded string
	for _, a := range applied {
		if a.Version == version {
			recorded = a.Checksum
			break
		}
	}

	if recorded == "" {
		return false, fmt.Errorf("migration %s not found in schema_migrations", version)
	}

	// Compute current file checksum
	current, err := computeChecksum(filePath)
	if err != nil {
		return false, err
	}

	return recorded == current, nil
}

// ListMigrations returns all migrations on disk and their status
func (m *MigrationRunner) ListMigrations(ctx context.Context) (map[string]interface{}, error) {
	migrations, err := m.loadMigrations()
	if err != nil {
		return nil, err
	}

	applied, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return nil, err
	}

	appliedMap := make(map[string]*MigrationRecord)
	for _, a := range applied {
		appliedMap[a.Version] = a
	}

	status := make(map[string]interface{})
	for _, mig := range migrations {
		if mig.IsRollback {
			continue
		}

		rec, isApplied := appliedMap[mig.Version]
		info := map[string]interface{}{
			"name":     mig.Name,
			"path":     mig.Path,
			"applied":  isApplied,
			"checksum": mig.Checksum,
		}

		if isApplied {
			info["applied_at"] = rec.AppliedAt
			info["checksum_verified"] = rec.Checksum == mig.Checksum
		}

		status[mig.Version] = info
	}

	return status, nil
}

// ============================================================================
// PRIVATE HELPER METHODS
// ============================================================================

// applyMigration executes a migration in a transaction
func (m *MigrationRunner) applyMigration(ctx context.Context, mig *MigrationFile) error {
	// Read migration SQL
	sql, err := os.ReadFile(mig.Path)
	if err != nil {
		return fmt.Errorf("read migration file: %w", err)
	}

	// Execute in transaction
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Execute migration
	if _, err := tx.Exec(ctx, string(sql)); err != nil {
		return fmt.Errorf("execute migration: %w", err)
	}

	// Record migration in schema_migrations
	recordSQL := fmt.Sprintf(`
		INSERT INTO %s (version, checksum, status)
		VALUES ($1, $2, $3)
		ON CONFLICT (version) DO NOTHING
	`, m.schemaTable)

	if _, err := tx.Exec(ctx, recordSQL, mig.Version, mig.Checksum, "success"); err != nil {
		return fmt.Errorf("record migration: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// rollbackMigration executes a rollback in a transaction
func (m *MigrationRunner) rollbackMigration(ctx context.Context, mig *MigrationFile) error {
	// Read rollback SQL
	sql, err := os.ReadFile(mig.RollbackPath)
	if err != nil {
		return fmt.Errorf("read rollback file: %w", err)
	}

	// Execute in transaction
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Execute rollback
	if _, err := tx.Exec(ctx, string(sql)); err != nil {
		return fmt.Errorf("execute rollback: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// recordRollback removes a migration version from schema_migrations
func (m *MigrationRunner) recordRollback(ctx context.Context, version string) error {
	sql := fmt.Sprintf("DELETE FROM %s WHERE version = $1", m.schemaTable)
	_, err := m.pool.Exec(ctx, sql, version)
	return err
}

// loadMigrations loads all migration files from disk
func (m *MigrationRunner) loadMigrations() ([]MigrationFile, error) {
	entries, err := os.ReadDir(m.migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("read migrations directory: %w", err)
	}

	var migrations []MigrationFile
	rollbackMap := make(map[string]string) // version -> rollback path

	// First pass: identify rollback files
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.HasPrefix(name, "rollback_") && strings.HasSuffix(name, ".sql") {
			// Extract version from rollback_NNN.sql
			version := strings.TrimPrefix(name, "rollback_")
			version = strings.TrimSuffix(version, ".sql")
			rollbackPath := filepath.Join(m.migrationsDir, name)
			rollbackMap[version] = rollbackPath
		}
	}

	// Second pass: create migration objects
	for _, entry := range entries {
		if entry.IsDir() || strings.HasPrefix(entry.Name(), "rollback_") {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}

		path := filepath.Join(m.migrationsDir, name)

		// Extract version from NNN_name.sql
		version := strings.Split(name, "_")[0]

		// Compute checksum
		checksum, err := computeChecksum(path)
		if err != nil {
			return nil, fmt.Errorf("compute checksum for %s: %w", name, err)
		}

		mig := MigrationFile{
			Version:      version,
			Name:         name,
			Path:         path,
			RollbackPath: rollbackMap[version],
			Checksum:     checksum,
			IsRollback:   false,
		}

		migrations = append(migrations, mig)
	}

	// Sort by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// getAppliedMigrations retrieves all applied migrations from the database
func (m *MigrationRunner) getAppliedMigrations(ctx context.Context) ([]*MigrationRecord, error) {
	query := fmt.Sprintf(`
		SELECT version, applied_at, checksum
		FROM %s
		ORDER BY version ASC
	`, m.schemaTable)

	rows, err := m.pool.Query(ctx, query)
	if err != nil {
		// Table doesn't exist yet
		if strings.Contains(err.Error(), "does not exist") {
			return nil, nil
		}
		return nil, fmt.Errorf("query schema_migrations: %w", err)
	}
	defer rows.Close()

	var records []*MigrationRecord
	for rows.Next() {
		var rec MigrationRecord
		if err := rows.Scan(&rec.Version, &rec.AppliedAt, &rec.Checksum); err != nil {
			return nil, fmt.Errorf("scan migration record: %w", err)
		}
		records = append(records, &rec)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return records, nil
}

// computeChecksum computes SHA256 of a file
func computeChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("compute hash: %w", err)
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
