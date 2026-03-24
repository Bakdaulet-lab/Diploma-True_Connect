package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/trueconnect/backend/internal/repository"
)

type contextKey string

const txKey contextKey = "tx"

// UoW implements repository.UnitOfWork for PostgreSQL using pgxpool.
type UoW struct {
	pool *pgxpool.Pool
}

var _ repository.UnitOfWork = (*UoW)(nil)

// NewUoW creates a new UnitOfWork with the given connection pool.
func NewUoW(pool *pgxpool.Pool) *UoW {
	return &UoW{pool: pool}
}

// Do executes the given function within a transaction.
func (u *UoW) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	// If there's already a transaction, just use it.
	if _, ok := ctx.Value(txKey).(pgx.Tx); ok {
		return fn(ctx)
	}

	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}

	// Make sure we rollback if not committed.
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	// Inject the transaction into the context
	txCtx := context.WithValue(ctx, txKey, tx)

	if err := fn(txCtx); err != nil {
		return err // the defer will rollback
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	return nil
}

// dbRunner interface allows repos to use either pool or tx seamlessly.
type dbRunner interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, optionsAndArgs ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, optionsAndArgs ...interface{}) pgx.Row
}

// runner returns a dbRunner from the context if it exists, otherwise falls back to the pool.
func runner(ctx context.Context, pool *pgxpool.Pool) dbRunner {
	if tx, ok := ctx.Value(txKey).(pgx.Tx); ok {
		return tx
	}
	return pool
}
