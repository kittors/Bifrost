package database_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/kittors/bifrost/apps/gateway/internal/auth"
	"github.com/kittors/bifrost/apps/gateway/internal/database"
)

var databaseTestDatabaseCounter uint64

func TestMigrateUpSeedAndDown(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dsn := createTestDatabase(t, ctx)

	if err := database.MigrateUp(ctx, dsn); err != nil {
		t.Fatalf("migrate up: %v", err)
	}

	assertTableExists(t, ctx, dsn, "users")
	assertTableExists(t, ctx, dsn, "services")
	assertTableExists(t, ctx, dsn, "audit_events")
	assertTableExists(t, ctx, dsn, "device_challenges")

	if err := database.SeedPhase1(ctx, dsn); err != nil {
		t.Fatalf("seed phase 1 first run: %v", err)
	}

	if err := database.SeedPhase1(ctx, dsn); err != nil {
		t.Fatalf("seed phase 1 second run: %v", err)
	}

	assertCount(t, ctx, dsn, "roles", 3)
	assertCount(t, ctx, dsn, "services", 3)
	assertCount(t, ctx, dsn, "users", 3)
	assertOverrideEffect(t, ctx, dsn, "user_bob", "service_jenkins", "deny")
	assertSeedPassword(t, ctx, dsn, "user_admin", "ChangeMe123!")

	if err := database.MigrateDownToZero(ctx, dsn); err != nil {
		t.Fatalf("migrate down to zero: %v", err)
	}

	assertTableMissing(t, ctx, dsn, "users")
}

func createTestDatabase(t *testing.T, ctx context.Context) string {
	t.Helper()

	adminDSN := os.Getenv("BIFROST_DATABASE_TEST_URL")
	if adminDSN == "" {
		adminDSN = "postgres://bifrost:bifrost@127.0.0.1:5432/postgres?sslmode=disable"
	}

	adminDB := openDB(t, adminDSN)

	databaseName := fmt.Sprintf("bifrost_test_%d_%d", time.Now().UnixNano(), atomic.AddUint64(&databaseTestDatabaseCounter, 1))
	if _, err := adminDB.ExecContext(ctx, "CREATE DATABASE "+databaseName); err != nil {
		t.Fatalf("create database %s: %v", databaseName, err)
	}

	t.Cleanup(func() {
		dropCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if _, err := adminDB.ExecContext(dropCtx, "DROP DATABASE "+databaseName+" WITH (FORCE)"); err != nil {
			t.Fatalf("drop database %s: %v", databaseName, err)
		}
	})

	return strings.Replace(adminDSN, "/postgres?", "/"+databaseName+"?", 1)
}

func openDB(t *testing.T, dsn string) *sql.DB {
	t.Helper()

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}

func assertTableExists(t *testing.T, ctx context.Context, dsn string, tableName string) {
	t.Helper()

	db := openDB(t, dsn)

	var exists bool
	if err := db.QueryRowContext(
		ctx,
		`SELECT EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = $1
		)`,
		tableName,
	).Scan(&exists); err != nil {
		t.Fatalf("check table %s exists: %v", tableName, err)
	}

	if !exists {
		t.Fatalf("expected table %s to exist", tableName)
	}
}

func assertTableMissing(t *testing.T, ctx context.Context, dsn string, tableName string) {
	t.Helper()

	db := openDB(t, dsn)

	var exists bool
	if err := db.QueryRowContext(
		ctx,
		`SELECT EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = $1
		)`,
		tableName,
	).Scan(&exists); err != nil {
		t.Fatalf("check table %s missing: %v", tableName, err)
	}

	if exists {
		t.Fatalf("expected table %s to be removed", tableName)
	}
}

func assertCount(t *testing.T, ctx context.Context, dsn string, tableName string, expected int) {
	t.Helper()

	db := openDB(t, dsn)

	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
	if err := db.QueryRowContext(ctx, query).Scan(&count); err != nil {
		t.Fatalf("count rows from %s: %v", tableName, err)
	}

	if count != expected {
		t.Fatalf("expected %d rows in %s, got %d", expected, tableName, count)
	}
}

func assertOverrideEffect(t *testing.T, ctx context.Context, dsn string, userID string, serviceID string, expectedEffect string) {
	t.Helper()

	db := openDB(t, dsn)

	var effect string
	if err := db.QueryRowContext(
		ctx,
		`SELECT effect
		FROM user_service_overrides
		WHERE user_id = $1 AND service_id = $2`,
		userID,
		serviceID,
	).Scan(&effect); err != nil {
		t.Fatalf("query override effect: %v", err)
	}

	if effect != expectedEffect {
		t.Fatalf("expected override effect %s, got %s", expectedEffect, effect)
	}
}

func assertSeedPassword(t *testing.T, ctx context.Context, dsn string, userID string, password string) {
	t.Helper()

	db := openDB(t, dsn)

	var passwordHash string
	if err := db.QueryRowContext(
		ctx,
		`SELECT password_hash
		FROM users
		WHERE id = $1`,
		userID,
	).Scan(&passwordHash); err != nil {
		t.Fatalf("query password hash: %v", err)
	}

	ok, err := auth.DefaultPasswordHasher().Verify(passwordHash, password)
	if err != nil {
		t.Fatalf("verify seed password: %v", err)
	}

	if !ok {
		t.Fatalf("expected password %q to match seeded hash for %s", password, userID)
	}
}
