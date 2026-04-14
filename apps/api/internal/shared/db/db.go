package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, err
	}

	configurePool(cfg)

	return pgxpool.NewWithConfig(ctx, cfg)
}

func configurePool(cfg *pgxpool.Config) {
	// Supabase transaction poolers can reuse backend sessions across requests.
	// Avoid implicit prepared statements from pgx to prevent "already exists" errors.
	cfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeExec
	cfg.MaxConns = 10
	cfg.MinConns = 2
	cfg.MaxConnIdleTime = 5 * time.Minute
}
