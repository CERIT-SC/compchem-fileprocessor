package repository_common

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func CommitTx(ctx context.Context, tx pgx.Tx, logger *zap.Logger) error {
	err := tx.Commit(ctx)
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrTxClosed) {
		logger.Error("Transaction already closed")
		return err
	}

	if errors.Is(err, pgx.ErrTxCommitRollback) {
		logger.Error("Pg aborted this transaction and will rollback instead of commit")
		return err
	}

	return err
}

func QueryOneTx[T any](
	ctx context.Context,
	tx pgx.Tx,
	query string,
	args ...any,
) (*T, error) {
	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[T])
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func QueryOne[T any](
	ctx context.Context,
	pool *pgxpool.Pool,
	query string,
	args ...any,
) (*T, error) {
	rows, err := pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[T])
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func QueryManyTx[T any](
	ctx context.Context,
	tx pgx.Tx,
	query string,
	args ...any,
) ([]T, error) {
	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results, err := pgx.CollectRows(rows, pgx.RowToStructByName[T])
	if err != nil {
		return nil, err
	}

	return results, nil
}

func QueryMany[T any](
	ctx context.Context,
	pool *pgxpool.Pool,
	query string,
	args ...any,
) ([]T, error) {
	rows, err := pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results, err := pgx.CollectRows(rows, pgx.RowToStructByName[T])
	if err != nil {
		return nil, err
	}

	return results, nil
}
