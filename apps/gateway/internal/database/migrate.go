package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

//go:embed seed/phase1.sql
var seedFS embed.FS

func MigrateUp(ctx context.Context, dsn string) error {
	return withDatabase(ctx, dsn, func(db *sql.DB) error {
		_ = ctx
		goose.SetBaseFS(migrationsFS)
		if err := goose.SetDialect("postgres"); err != nil {
			return fmt.Errorf("set goose dialect: %w", err)
		}

		if err := goose.Up(db, "migrations"); err != nil {
			return fmt.Errorf("goose up: %w", err)
		}

		return nil
	})
}

func MigrateDownToZero(ctx context.Context, dsn string) error {
	return withDatabase(ctx, dsn, func(db *sql.DB) error {
		_ = ctx
		goose.SetBaseFS(migrationsFS)
		if err := goose.SetDialect("postgres"); err != nil {
			return fmt.Errorf("set goose dialect: %w", err)
		}

		if err := goose.Reset(db, "migrations"); err != nil {
			return fmt.Errorf("goose reset: %w", err)
		}

		return nil
	})
}

func SeedPhase1(ctx context.Context, dsn string) error {
	return withDatabase(ctx, dsn, func(db *sql.DB) error {
		seedSQL, err := seedFS.ReadFile("seed/phase1.sql")
		if err != nil {
			return fmt.Errorf("read phase1 seed SQL: %w", err)
		}

		if _, err := db.ExecContext(ctx, string(seedSQL)); err != nil {
			return fmt.Errorf("execute phase1 seed SQL: %w", err)
		}

		return nil
	})
}

func withDatabase(ctx context.Context, dsn string, fn func(db *sql.DB) error) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	if err := fn(db); err != nil {
		return err
	}

	return nil
}

func DefaultDatabaseURL() string {
	return strings.TrimSpace("postgres://bifrost:bifrost@127.0.0.1:5432/bifrost?sslmode=disable")
}
