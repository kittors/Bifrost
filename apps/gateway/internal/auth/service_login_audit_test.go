package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/kittors/bifrost/apps/gateway/internal/auth"
	"github.com/kittors/bifrost/apps/gateway/internal/contracts"
	"github.com/kittors/bifrost/apps/gateway/internal/database"
)

// 登录失败审计单独测试，避免混进成功登录或设备流程断言。
func TestServiceAdminLoginWritesFailedAuditEvent(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dsn := createTestDatabase(t, ctx)
	if err := database.MigrateUp(ctx, dsn); err != nil {
		t.Fatalf("migrate up: %v", err)
	}

	db := openDB(t, dsn)
	now := time.Date(2026, time.April, 18, 5, 10, 0, 0, time.UTC)
	service := auth.Service{
		DB:              db,
		PasswordHasher:  auth.DefaultPasswordHasher(),
		TokenIssuer:     auth.TokenIssuer{Secret: []byte("0123456789abcdef0123456789abcdef"), TTL: 15 * time.Minute, Now: func() time.Time { return now }},
		Now:             func() time.Time { return now },
		RefreshTokenTTL: 7 * 24 * time.Hour,
	}

	insertUserWithRoles(t, ctx, db, "user_admin", "admin", "Administrator", "ChangeMe123!", []roleSeed{{id: "role_admin", name: "admin", displayName: "Administrator"}})

	if _, err := service.AdminLogin(ctx, auth.AdminLoginInput{
		Password:  "WrongPassword!",
		RequestID: "req_audit_login_failed",
		Username:  "admin",
	}); err == nil {
		t.Fatal("expected invalid credentials error")
	}

	var (
		count    int
		targetID string
		auditTyp string
		result   string
	)
	if err := db.QueryRowContext(
		ctx,
		`SELECT COUNT(*), COALESCE(MAX(target_id), ''), COALESCE(MAX(type), ''), COALESCE(MAX(result), '')
		FROM audit_events
		WHERE request_id = $1`,
		"req_audit_login_failed",
	).Scan(&count, &targetID, &auditTyp, &result); err != nil {
		t.Fatalf("query login failed audit event: %v", err)
	}

	if count != 1 {
		t.Fatalf("expected 1 login failed audit event, got %d", count)
	}
	if targetID != "admin" {
		t.Fatalf("expected failed audit target admin, got %q", targetID)
	}
	if auditTyp != string(contracts.AuditEventTypeAuthLoginFailed) {
		t.Fatalf("expected audit type %q, got %q", contracts.AuditEventTypeAuthLoginFailed, auditTyp)
	}
	if result != "failure" {
		t.Fatalf("expected failed audit result, got %q", result)
	}
}
