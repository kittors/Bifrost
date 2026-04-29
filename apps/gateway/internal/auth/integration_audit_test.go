package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/kittors/bifrost/apps/gateway/internal/auth"
	"github.com/kittors/bifrost/apps/gateway/internal/contracts"
)

func TestIntegrationAuditListReturnsNewestEventsFirst(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	service, db := newSeededIntegrationService(t, ctx)
	adminLogin := bootstrapSeedAdmin(t, ctx, service)

	if _, err := service.CreateAdminUser(ctx, auth.CreateAdminUserInput{
		AccessToken: adminLogin.AccessToken,
		RequestID:   "req_integration_audit_create",
		Username:    "echo",
		DisplayName: "Echo",
		Email:       "echo@example.com",
		Password:    "ChangeMe123!",
		RoleIDs:     []string{"role_developer"},
	}); err != nil {
		t.Fatalf("create admin user: %v", err)
	}

	// 人工拉开时间戳，确保排序断言稳定，不依赖数据库当前时间分辨率。
	if _, err := db.ExecContext(
		ctx,
		`UPDATE audit_events
		SET occurred_at = CASE request_id
			WHEN 'req_integration_admin_login' THEN '2026-04-18T12:20:01Z'
			WHEN 'req_integration_audit_create' THEN '2026-04-18T12:20:02Z'
			ELSE occurred_at
		END
		WHERE request_id IN ('req_integration_admin_login', 'req_integration_audit_create')`,
	); err != nil {
		t.Fatalf("normalize audit timestamps: %v", err)
	}

	audits, err := service.ListAdminAuditEvents(ctx, auth.ListAdminAuditEventsInput{
		AccessToken: adminLogin.AccessToken,
		Page:        1,
		PageSize:    10,
	})
	if err != nil {
		t.Fatalf("list admin audit events: %v", err)
	}

	if len(audits.Items) < 2 {
		t.Fatalf("expected at least 2 audit events, got %d", len(audits.Items))
	}

	if audits.Items[0].Type != string(contracts.AuditEventTypeAdminUserCreated) {
		t.Fatalf("expected newest audit to be admin user created, got %s", audits.Items[0].Type)
	}

	if audits.Items[1].Type != string(contracts.AuditEventTypeAuthLoginSucceeded) {
		t.Fatalf("expected second audit to be admin login succeeded, got %s", audits.Items[1].Type)
	}
}
