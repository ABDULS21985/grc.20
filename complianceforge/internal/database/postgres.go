// Package database provides PostgreSQL connection management with pgxpool.
// It includes connection pooling, health checks, retry logic, and tenant context management.
package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/complianceforge/platform/internal/config"
)

// DB wraps pgxpool.Pool with additional functionality.
type DB struct {
	Pool *pgxpool.Pool
}

// New creates a new database connection pool with the given configuration.
func New(cfg config.DatabaseConfig) (*DB, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Configure connection pool
	poolConfig.MaxConns = int32(cfg.MaxOpenConns)
	poolConfig.MinConns = int32(cfg.MaxIdleConns)
	poolConfig.MaxConnLifetime = time.Duration(cfg.ConnMaxLifetime) * time.Second
	poolConfig.MaxConnIdleTime = time.Duration(cfg.ConnMaxIdleTime) * time.Second
	poolConfig.HealthCheckPeriod = 30 * time.Second

	// Configure connection-level settings: on each new connection, set the
	// search_path and ensure RLS context is clean.
	poolConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		_, err := conn.Exec(ctx, "SET search_path TO public")
		return err
	}

	// Connect with retry logic
	var pool *pgxpool.Pool
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		pool, err = pgxpool.NewWithConfig(ctx, poolConfig)
		cancel()
		if err == nil {
			break
		}
		log.Warn().
			Err(err).
			Int("attempt", i+1).
			Int("max_retries", maxRetries).
			Msg("Failed to connect to database, retrying...")
		time.Sleep(time.Duration(i+1) * 2 * time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database after %d retries: %w", maxRetries, err)
	}

	// Verify the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info().
		Str("host", cfg.Host).
		Int("port", cfg.Port).
		Str("database", cfg.Name).
		Int("max_conns", cfg.MaxOpenConns).
		Msg("Connected to PostgreSQL")

	return &DB{Pool: pool}, nil
}

// Close closes the database connection pool.
func (db *DB) Close() {
	db.Pool.Close()
	log.Info().Msg("Database connection pool closed")
}

// Health checks if the database is reachable.
func (db *DB) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	return db.Pool.Ping(ctx)
}

// Stats returns connection pool statistics.
func (db *DB) Stats() *pgxpool.Stat {
	return db.Pool.Stat()
}

// SetTenant sets the current tenant (organization_id) for RLS policies.
// This must be called at the beginning of each request to enforce tenant isolation.
func (db *DB) SetTenant(ctx context.Context, tx pgx.Tx, orgID string) error {
	_, err := tx.Exec(ctx, "SELECT set_config('app.current_tenant', $1, true)", orgID)
	return err
}

// BeginTx starts a new transaction with tenant context set.
func (db *DB) BeginTx(ctx context.Context, orgID string) (pgx.Tx, error) {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	if orgID != "" {
		if err := db.SetTenant(ctx, tx, orgID); err != nil {
			tx.Rollback(ctx)
			return nil, fmt.Errorf("failed to set tenant context: %w", err)
		}
	}

	return tx, nil
}

// ExecWithTenant executes a query within a tenant-scoped transaction.
func (db *DB) ExecWithTenant(ctx context.Context, orgID string, fn func(tx pgx.Tx) error) error {
	tx, err := db.BeginTx(ctx, orgID)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
