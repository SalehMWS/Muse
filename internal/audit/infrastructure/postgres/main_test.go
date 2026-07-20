package postgres_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"

	"github.com/SalehMWS/Muse/internal/shared/config"
)

var pool *pgxpool.Pool

func TestMain(m *testing.M) {
	os.Exit(run(m))
}

func run(m *testing.M) int {
	root := repoRoot()
	_ = godotenv.Load(filepath.Join(root, "configs", ".env"))

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "audit integration tests: load config:", err)
		return 1
	}

	ctx := context.Background()

	candidate, err := pgxpool.New(ctx, cfg.Postgres.DSN())
	if err != nil {
		fmt.Fprintln(os.Stderr, "audit integration tests: skipping, cannot create pool:", err)
		return m.Run()
	}
	if err := candidate.Ping(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "audit integration tests: skipping, cannot reach postgres:", err)
		candidate.Close()
		return m.Run()
	}

	if err := migrateUp(cfg, root); err != nil {
		fmt.Fprintln(os.Stderr, "audit integration tests: run migrations:", err)
		candidate.Close()
		return 1
	}

	pool = candidate
	defer pool.Close()

	return m.Run()
}

func migrateUp(cfg *config.Config, root string) error {
	sqlDB, err := sql.Open("pgx", cfg.Postgres.DSN())
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer func() { _ = sqlDB.Close() }()

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}

	return goose.Up(sqlDB, filepath.Join(root, "deployments", "migrations"))
}

func repoRoot() string {
	_, thisFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "..")
}
