package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/SalehMWS/Muse/internal/shared/config"
)

func main() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return errors.New("usage: migrate <up|down|redo|status|up-by-one> [args...]")
	}

	command := args[0]
	commandArgs := args[1:]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	db, err := sql.Open("pgx", cfg.Postgres.DSN())
	if err != nil {
		return fmt.Errorf("migrate: open database: %w", err)
	}
	defer func() { _ = db.Close() }()

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("migrate: set dialect: %w", err)
	}

	if err := goose.RunContext(ctx, command, db, cfg.Migrations.Dir, commandArgs...); err != nil {
		return fmt.Errorf("migrate: %s: %w", command, err)
	}

	return nil
}
