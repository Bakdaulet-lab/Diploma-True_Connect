package repository

import "context"

// UnitOfWork defines an abstraction for executing operations within a database transaction.
type UnitOfWork interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}
