package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/kittors/bifrost/apps/gateway/internal/auth"
	"github.com/kittors/bifrost/apps/gateway/internal/database"
)

// 审计测试单独保留，方便继续补登录失败、访问拒绝等审计场景。

func TestServiceWritesAuditEventsForKeyActions(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dsn := createTestDatabase(t, ctx)
	if err := database.MigrateUp(ctx, dsn); err != nil {
		t.Fatalf("migrate up: %v", err)
	}

	db := openDB(t, dsn)
	now := time.Date(2026, time.April, 17, 13, 30, 0, 0, time.UTC)
	service := auth.Service{
		DB:              db,
		PasswordHasher:  auth.DefaultPasswordHasher(),
		TokenIssuer:     auth.TokenIssuer{Secret: []byte("0123456789abcdef0123456789abcdef"), TTL: 15 * time.Minute, Now: func() time.Time { return now }},
		Now:             func() time.Time { return now },
		RefreshTokenTTL: 7 * 24 * time.Hour,
		SessionIDFactory: func() (string, error) {
			return "session_audit_01", nil
		},
		UserIDFactory: func() (string, error) {
			return "user_audit_created_01", nil
		},
	}

	insertUserWithRoles(t, ctx, db, "user_admin", "admin", "Administrator", "correct horse battery staple", []roleSeed{{id: "role_admin", name: "admin", displayName: "Administrator"}})
	insertRole(t, ctx, db, roleSeed{id: "role_developer", name: "developer", displayName: "Developer"})

	loginResult, err := service.AdminLogin(ctx, auth.AdminLoginInput{
		Username:  "admin",
		Password:  "correct horse battery staple",
		RequestID: "req_audit_login",
	})
	if err != nil {
		t.Fatalf("admin login: %v", err)
	}

	if _, err := service.CreateAdminUser(ctx, auth.CreateAdminUserInput{
		AccessToken: loginResult.AccessToken,
		RequestID:   "req_audit_user_create",
		Username:    "delta",
		DisplayName: "Delta",
		Email:       "delta@example.com",
		Password:    "ChangeMe123!",
		RoleIDs:     []string{"role_developer"},
	}); err != nil {
		t.Fatalf("create admin user: %v", err)
	}

	assertAuditEventCountByRequest(t, ctx, db, "req_audit_login", 1)
	assertAuditEventCountByRequest(t, ctx, db, "req_audit_user_create", 1)
}
