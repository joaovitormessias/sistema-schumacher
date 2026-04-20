package db

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestConfigurePoolSetsExecMode(t *testing.T) {
	cfg, err := pgxpool.ParseConfig("postgres://user:pass@localhost:5432/schumacher?sslmode=disable")
	if err != nil {
		t.Fatalf("ParseConfig() error = %v", err)
	}

	cfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheStatement

	configurePool(cfg)

	if cfg.ConnConfig.DefaultQueryExecMode != pgx.QueryExecModeExec {
		t.Fatalf("DefaultQueryExecMode = %v, want %v", cfg.ConnConfig.DefaultQueryExecMode, pgx.QueryExecModeExec)
	}
	if cfg.MaxConns != 10 {
		t.Fatalf("MaxConns = %d, want 10", cfg.MaxConns)
	}
	if cfg.MinConns != 2 {
		t.Fatalf("MinConns = %d, want 2", cfg.MinConns)
	}
	if cfg.MaxConnIdleTime != 5*time.Minute {
		t.Fatalf("MaxConnIdleTime = %v, want %v", cfg.MaxConnIdleTime, 5*time.Minute)
	}
}
