package auth_test

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/kittors/bifrost/apps/gateway/internal/auth"
)

// 测试 helper 统一放置，业务测试文件只保留场景本身。

var authTestDatabaseCounter uint64

type roleSeed struct {
	id          string
	name        string
	displayName string
}

func insertUserWithRoles(t *testing.T, ctx context.Context, db *sql.DB, userID string, username string, displayName string, password string, roles []roleSeed) {
	t.Helper()

	hasher := auth.DefaultPasswordHasher()
	passwordHash, err := hasher.Hash(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	if _, err := db.ExecContext(
		ctx,
		`INSERT INTO users (id, username, display_name, email, password_hash, status)
		VALUES ($1, $2, $3, $4, $5, 'enabled')`,
		userID,
		username,
		displayName,
		username+"@example.com",
		passwordHash,
	); err != nil {
		t.Fatalf("insert user: %v", err)
	}

	for _, role := range roles {
		insertRole(t, ctx, db, role)

		if _, err := db.ExecContext(
			ctx,
			`INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)`,
			userID,
			role.id,
		); err != nil {
			t.Fatalf("insert user role %s: %v", role.id, err)
		}
	}
}

func insertRole(t *testing.T, ctx context.Context, db *sql.DB, role roleSeed) {
	t.Helper()

	if _, err := db.ExecContext(
		ctx,
		`INSERT INTO roles (id, name, display_name, description, is_system)
		VALUES ($1, $2, $3, '', true)
		ON CONFLICT (id) DO NOTHING`,
		role.id,
		role.name,
		role.displayName,
	); err != nil {
		t.Fatalf("insert role %s: %v", role.id, err)
	}
}

func insertDevice(t *testing.T, ctx context.Context, db *sql.DB, deviceID string, userID string, status string) {
	t.Helper()

	insertDeviceWithKey(t, ctx, db, deviceID, userID, status, "public-key", "fp_"+deviceID)
}

func insertDeviceWithKey(t *testing.T, ctx context.Context, db *sql.DB, deviceID string, userID string, status string, publicKey string, fingerprint string) {
	t.Helper()

	if _, err := db.ExecContext(
		ctx,
		`INSERT INTO devices (id, user_id, name, os, client_version, public_key, public_key_fingerprint, status)
		VALUES ($1, $2, 'Alice MacBook Pro', 'macOS', '1.0.0', 'public-key', $3, $4)`,
		deviceID,
		userID,
		fingerprint,
		status,
	); err != nil {
		t.Fatalf("insert device: %v", err)
	}

	if _, err := db.ExecContext(
		ctx,
		`UPDATE devices
		SET public_key = $2
		WHERE id = $1`,
		deviceID,
		publicKey,
	); err != nil {
		t.Fatalf("update device public key: %v", err)
	}
}

func insertService(t *testing.T, ctx context.Context, db *sql.DB, serviceID string, key string, name string, group string, publicPath string, status string) {
	t.Helper()

	if _, err := db.ExecContext(
		ctx,
		`INSERT INTO services (id, key, name, description, group_name, protocol, upstream_url, public_path, status)
		VALUES ($1, $2, $3, $4, $5, 'http', $6, $7, $8)`,
		serviceID,
		key,
		name,
		name+" service",
		group,
		"http://"+key+":8080",
		publicPath,
		status,
	); err != nil {
		t.Fatalf("insert service: %v", err)
	}
}

func insertRoleService(t *testing.T, ctx context.Context, db *sql.DB, roleID string, serviceID string) {
	t.Helper()

	if _, err := db.ExecContext(
		ctx,
		`INSERT INTO role_services (role_id, service_id)
		VALUES ($1, $2)`,
		roleID,
		serviceID,
	); err != nil {
		t.Fatalf("insert role service: %v", err)
	}
}

func insertUserServiceOverride(t *testing.T, ctx context.Context, db *sql.DB, userID string, serviceID string, effect string) {
	t.Helper()

	if _, err := db.ExecContext(
		ctx,
		`INSERT INTO user_service_overrides (user_id, service_id, effect, reason, created_by)
		VALUES ($1, $2, $3, 'test override', $1)`,
		userID,
		serviceID,
		effect,
	); err != nil {
		t.Fatalf("insert user service override: %v", err)
	}
}

func insertAuditEvent(t *testing.T, ctx context.Context, db *sql.DB, id string, eventType string, actorUserID string, targetType string, targetID string, serviceID string, result string) {
	t.Helper()

	var nullableServiceID any
	if serviceID == "" {
		nullableServiceID = nil
	} else {
		nullableServiceID = serviceID
	}

	if _, err := db.ExecContext(
		ctx,
		`INSERT INTO audit_events (id, request_id, type, actor_user_id, target_type, target_id, service_id, result, summary)
		VALUES ($1, 'req_test', $2, $3, $4, $5, $6, $7, 'test audit')`,
		id,
		eventType,
		actorUserID,
		targetType,
		targetID,
		nullableServiceID,
		result,
	); err != nil {
		t.Fatalf("insert audit event: %v", err)
	}
}

func assertRoleServices(t *testing.T, ctx context.Context, db *sql.DB, roleID string, expected []string) {
	t.Helper()

	rows, err := db.QueryContext(ctx, `SELECT service_id FROM role_services WHERE role_id = $1 ORDER BY service_id ASC`, roleID)
	if err != nil {
		t.Fatalf("query role services: %v", err)
	}
	defer rows.Close()

	var actual []string
	for rows.Next() {
		var serviceID string
		if err := rows.Scan(&serviceID); err != nil {
			t.Fatalf("scan role service: %v", err)
		}
		actual = append(actual, serviceID)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate role services: %v", err)
	}

	if strings.Join(actual, ",") != strings.Join(expected, ",") {
		t.Fatalf("expected role services %#v, got %#v", expected, actual)
	}
}

func assertUserServiceOverrides(t *testing.T, ctx context.Context, db *sql.DB, userID string, expected map[string]string) {
	t.Helper()

	rows, err := db.QueryContext(ctx, `SELECT service_id, effect FROM user_service_overrides WHERE user_id = $1`, userID)
	if err != nil {
		t.Fatalf("query user service overrides: %v", err)
	}
	defer rows.Close()

	actual := map[string]string{}
	for rows.Next() {
		var serviceID string
		var effect string
		if err := rows.Scan(&serviceID, &effect); err != nil {
			t.Fatalf("scan user service override: %v", err)
		}
		actual[serviceID] = effect
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate user service overrides: %v", err)
	}

	if len(actual) != len(expected) {
		t.Fatalf("expected overrides %#v, got %#v", expected, actual)
	}
	for serviceID, effect := range expected {
		if actual[serviceID] != effect {
			t.Fatalf("expected override %s=%s, got %s", serviceID, effect, actual[serviceID])
		}
	}
}

func assertAuditEventCountByRequest(t *testing.T, ctx context.Context, db *sql.DB, requestID string, expected int) {
	t.Helper()

	var count int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM audit_events WHERE request_id = $1`, requestID).Scan(&count); err != nil {
		t.Fatalf("count audit events by request: %v", err)
	}
	if count != expected {
		t.Fatalf("expected %d audit events for %s, got %d", expected, requestID, count)
	}
}

func generateEd25519Material(t *testing.T) (string, ed25519.PrivateKey, string) {
	t.Helper()

	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate ed25519 key: %v", err)
	}

	encodedPublicKey := base64.RawURLEncoding.EncodeToString(publicKey)
	fingerprint := "fp_" + encodedPublicKey[:16]
	return encodedPublicKey, privateKey, fingerprint
}

func issueAccessTokenForTest(t *testing.T, issuer auth.TokenIssuer, userID string, deviceID string, sessionID string) string {
	t.Helper()

	token, _, err := issuer.IssueAccessToken(auth.AccessTokenClaims{
		UserID:            userID,
		DeviceID:          deviceID,
		SessionID:         sessionID,
		PermissionVersion: 1,
	})
	if err != nil {
		t.Fatalf("issue access token for test: %v", err)
	}

	return token
}

func createTestDatabase(t *testing.T, ctx context.Context) string {
	t.Helper()

	adminDSN := os.Getenv("BIFROST_DATABASE_TEST_URL")
	if adminDSN == "" {
		adminDSN = "postgres://bifrost:bifrost@127.0.0.1:5432/postgres?sslmode=disable"
	}

	adminDB := openDB(t, adminDSN)

	databaseName := fmt.Sprintf("bifrost_auth_test_%d_%d", time.Now().UnixNano(), atomic.AddUint64(&authTestDatabaseCounter, 1))
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
