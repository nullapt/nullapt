package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool wraps a pgxpool connection pool.
type Pool struct {
	*pgxpool.Pool
}

// Connect opens a connection pool using a Postgres DSN.
func Connect(ctx context.Context, dsn string) (*Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parsing DSN: %w", err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connecting to postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("pinging postgres: %w", err)
	}
	return &Pool{pool}, nil
}
