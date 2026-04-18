package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/kittors/bifrost/apps/gateway/internal/auth"
	"github.com/kittors/bifrost/apps/gateway/internal/database"
)

// 后台管理测试覆盖用户、角色、服务、设备、审计和策略覆盖的配置闭环。

func TestServiceAdminUserListCreateAndUpdate(t *testing.T) {
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
			return "session_admin_users_01", nil
		},
		UserIDFactory: func() (string, error) {
			return "user_created_01", nil
		},
	}

	insertUserWithRoles(t, ctx, db, "user_admin", "admin", "Administrator", "correct horse battery staple", []roleSeed{{id: "role_admin", name: "admin", displayName: "Administrator"}})
	insertRole(t, ctx, db, roleSeed{id: "role_developer", name: "developer", displayName: "Developer"})

	loginResult, err := service.AdminLogin(ctx, auth.AdminLoginInput{
		Username: "admin",
		Password: "correct horse battery staple",
	})
	if err != nil {
		t.Fatalf("admin login: %v", err)
	}

	created, err := service.CreateAdminUser(ctx, auth.CreateAdminUserInput{
		AccessToken: loginResult.AccessToken,
		Username:    "charlie",
		DisplayName: "Charlie",
		Email:       "charlie@example.com",
		Password:    "ChangeMe123!",
		RoleIDs:     []string{"role_developer"},
	})
	if err != nil {
		t.Fatalf("create admin user: %v", err)
	}

	if created.ID != "user_created_01" {
		t.Fatalf("expected created user id user_created_01, got %q", created.ID)
	}

	var passwordHash string
	if err := db.QueryRowContext(ctx, `SELECT password_hash FROM users WHERE id = $1`, "user_created_01").Scan(&passwordHash); err != nil {
		t.Fatalf("query created user password hash: %v", err)
	}
	ok, err := auth.DefaultPasswordHasher().Verify(passwordHash, "ChangeMe123!")
	if err != nil {
		t.Fatalf("verify created user password: %v", err)
	}
	if !ok {
		t.Fatal("expected created user password to verify")
	}

	updated, err := service.UpdateAdminUser(ctx, auth.UpdateAdminUserInput{
		AccessToken: loginResult.AccessToken,
		UserID:      "user_created_01",
		DisplayName: "Charles",
		Email:       "charles@example.com",
		RoleIDs:     []string{"role_admin"},
	})
	if err != nil {
		t.Fatalf("update admin user: %v", err)
	}

	if updated.DisplayName != "Charles" {
		t.Fatalf("expected updated display name Charles, got %q", updated.DisplayName)
	}

	if len(updated.Roles) != 1 || updated.Roles[0] != "role_admin" {
		t.Fatalf("expected updated role_admin role, got %#v", updated.Roles)
	}

	list, err := service.ListAdminUsers(ctx, auth.ListAdminUsersInput{
		AccessToken: loginResult.AccessToken,
		Page:        1,
		PageSize:    20,
		Keyword:     "charles",
	})
	if err != nil {
		t.Fatalf("list admin users: %v", err)
	}

	if list.Pagination.Total != 1 {
		t.Fatalf("expected one listed user, got total %d", list.Pagination.Total)
	}

	if list.Items[0].ID != "user_created_01" {
		t.Fatalf("expected created user in list, got %q", list.Items[0].ID)
	}
}

func TestServiceAdminUserDetailPasswordResetAndStatus(t *testing.T) {
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
			return "session_admin_user_detail_01", nil
		},
	}

	insertUserWithRoles(t, ctx, db, "user_admin", "admin", "Administrator", "correct horse battery staple", []roleSeed{{id: "role_admin", name: "admin", displayName: "Administrator"}})
	insertUserWithRoles(t, ctx, db, "user_alice", "alice", "Alice", "old password", []roleSeed{{id: "role_developer", name: "developer", displayName: "Developer"}})

	loginResult, err := service.AdminLogin(ctx, auth.AdminLoginInput{
		Username: "admin",
		Password: "correct horse battery staple",
	})
	if err != nil {
		t.Fatalf("admin login: %v", err)
	}

	detail, err := service.GetAdminUser(ctx, auth.GetAdminUserInput{
		AccessToken: loginResult.AccessToken,
		UserID:      "user_alice",
	})
	if err != nil {
		t.Fatalf("get admin user: %v", err)
	}
	if detail.Username != "alice" || detail.Roles[0] != "role_developer" {
		t.Fatalf("expected alice detail with developer role, got %#v", detail)
	}

	if err := service.ResetAdminUserPassword(ctx, auth.ResetAdminUserPasswordInput{
		AccessToken: loginResult.AccessToken,
		RequestID:   "req_reset_password",
		UserID:      "user_alice",
		Password:    "new password",
	}); err != nil {
		t.Fatalf("reset admin user password: %v", err)
	}

	var passwordHash string
	if err := db.QueryRowContext(ctx, `SELECT password_hash FROM users WHERE id = $1`, "user_alice").Scan(&passwordHash); err != nil {
		t.Fatalf("query reset password hash: %v", err)
	}
	ok, err := auth.DefaultPasswordHasher().Verify(passwordHash, "new password")
	if err != nil {
		t.Fatalf("verify reset password: %v", err)
	}
	if !ok {
		t.Fatal("expected reset password to verify")
	}

	disabled, err := service.SetAdminUserStatus(ctx, auth.SetAdminUserStatusInput{
		AccessToken: loginResult.AccessToken,
		RequestID:   "req_disable_user",
		UserID:      "user_alice",
		Status:      "disabled",
	})
	if err != nil {
		t.Fatalf("set admin user status: %v", err)
	}
	if disabled.Status != "disabled" {
		t.Fatalf("expected disabled user status, got %q", disabled.Status)
	}
}

func TestServiceAdminRoleServiceDeviceAuditAndOverrideManagement(t *testing.T) {
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
			return "session_admin_cfg_01", nil
		},
		RoleIDFactory: func() (string, error) {
			return "role_created_01", nil
		},
		ServiceIDFactory: func() (string, error) {
			return "service_created_01", nil
		},
	}

	insertUserWithRoles(t, ctx, db, "user_admin", "admin", "Administrator", "correct horse battery staple", []roleSeed{{id: "role_admin", name: "admin", displayName: "Administrator"}})
	insertUserWithRoles(t, ctx, db, "user_alice", "alice", "Alice", "correct horse battery staple", []roleSeed{{id: "role_developer", name: "developer", displayName: "Developer"}})
	insertDevice(t, ctx, db, "device_alice_admin_01", "user_alice", "trusted")
	insertService(t, ctx, db, "service_docs", "docs", "Docs", "shared", "/s/docs", "enabled")
	insertService(t, ctx, db, "service_gitlab", "gitlab", "GitLab", "engineering", "/s/gitlab", "enabled")
	insertAuditEvent(t, ctx, db, "audit_01", "auth.login.succeeded", "user_admin", "user", "user_admin", "", "success")

	loginResult, err := service.AdminLogin(ctx, auth.AdminLoginInput{
		Username: "admin",
		Password: "correct horse battery staple",
	})
	if err != nil {
		t.Fatalf("admin login: %v", err)
	}

	roles, err := service.ListAdminRoles(ctx, auth.ListAdminRolesInput{
		AccessToken: loginResult.AccessToken,
		Page:        1,
		PageSize:    20,
		Keyword:     "admin",
	})
	if err != nil {
		t.Fatalf("list admin roles: %v", err)
	}
	if roles.Pagination.Total != 1 {
		t.Fatalf("expected 1 role in filtered list, got %d", roles.Pagination.Total)
	}

	createdRole, err := service.CreateAdminRole(ctx, auth.CreateAdminRoleInput{
		AccessToken: loginResult.AccessToken,
		Name:        "ops",
		DisplayName: "Operations",
		Description: "Ops team",
	})
	if err != nil {
		t.Fatalf("create admin role: %v", err)
	}
	if createdRole.ID != "role_created_01" {
		t.Fatalf("expected role_created_01, got %q", createdRole.ID)
	}

	services, err := service.ListAdminServices(ctx, auth.ListAdminServicesInput{
		AccessToken: loginResult.AccessToken,
		Page:        1,
		PageSize:    20,
		Group:       "shared",
	})
	if err != nil {
		t.Fatalf("list admin services: %v", err)
	}
	if services.Pagination.Total != 1 {
		t.Fatalf("expected 1 shared service, got %d", services.Pagination.Total)
	}

	createdService, err := service.CreateAdminService(ctx, auth.CreateAdminServiceInput{
		AccessToken: loginResult.AccessToken,
		Key:         "jenkins",
		Name:        "Jenkins",
		Description: "CI server",
		Group:       "operations",
		Protocol:    "http",
		UpstreamURL: "http://jenkins:8080",
		PublicPath:  "/s/jenkins",
		Enabled:     true,
	})
	if err != nil {
		t.Fatalf("create admin service: %v", err)
	}
	if createdService.ID != "service_created_01" {
		t.Fatalf("expected service_created_01, got %q", createdService.ID)
	}

	serviceDetail, err := service.GetAdminService(ctx, auth.GetAdminServiceInput{
		AccessToken: loginResult.AccessToken,
		ServiceID:   "service_docs",
	})
	if err != nil {
		t.Fatalf("get admin service: %v", err)
	}
	if serviceDetail.Key != "docs" {
		t.Fatalf("expected docs service detail, got %#v", serviceDetail)
	}

	updatedService, err := service.UpdateAdminService(ctx, auth.UpdateAdminServiceInput{
		AccessToken: loginResult.AccessToken,
		ServiceID:   "service_docs",
		Name:        "Docs Portal",
		Description: "Shared docs portal",
		Group:       "shared",
		Protocol:    "http",
		UpstreamURL: "http://docs:8080",
		PublicPath:  "/s/docs",
	})
	if err != nil {
		t.Fatalf("update admin service: %v", err)
	}
	if updatedService.Name != "Docs Portal" {
		t.Fatalf("expected updated service name, got %#v", updatedService)
	}

	disabledService, err := service.SetAdminServiceStatus(ctx, auth.SetAdminServiceStatusInput{
		AccessToken: loginResult.AccessToken,
		RequestID:   "req_disable_service",
		ServiceID:   "service_docs",
		Status:      "disabled",
	})
	if err != nil {
		t.Fatalf("disable admin service: %v", err)
	}
	if disabledService.Status != "disabled" {
		t.Fatalf("expected disabled service status, got %q", disabledService.Status)
	}

	devices, err := service.ListAdminDevices(ctx, auth.ListAdminDevicesInput{
		AccessToken: loginResult.AccessToken,
		Page:        1,
		PageSize:    20,
		UserID:      "user_alice",
	})
	if err != nil {
		t.Fatalf("list admin devices: %v", err)
	}
	if devices.Pagination.Total != 1 {
		t.Fatalf("expected 1 device for alice, got %d", devices.Pagination.Total)
	}

	deviceDetail, err := service.GetAdminDevice(ctx, auth.GetAdminDeviceInput{
		AccessToken: loginResult.AccessToken,
		DeviceID:    "device_alice_admin_01",
	})
	if err != nil {
		t.Fatalf("get admin device: %v", err)
	}
	if deviceDetail.UserUsername != "alice" {
		t.Fatalf("expected alice device detail, got %#v", deviceDetail)
	}

	disabledDevice, err := service.SetAdminDeviceStatus(ctx, auth.SetAdminDeviceStatusInput{
		AccessToken: loginResult.AccessToken,
		RequestID:   "req_disable_device",
		DeviceID:    "device_alice_admin_01",
		Status:      "disabled",
	})
	if err != nil {
		t.Fatalf("disable admin device: %v", err)
	}
	if disabledDevice.Status != "disabled" {
		t.Fatalf("expected disabled device status, got %q", disabledDevice.Status)
	}

	audits, err := service.ListAdminAuditEvents(ctx, auth.ListAdminAuditEventsInput{
		AccessToken: loginResult.AccessToken,
		Page:        1,
		PageSize:    20,
		Type:        "auth.login.succeeded",
	})
	if err != nil {
		t.Fatalf("list admin audit events: %v", err)
	}
	if audits.Pagination.Total != 2 {
		t.Fatalf("expected 2 audit events, got %d", audits.Pagination.Total)
	}

	if err := service.ReplaceRoleServices(ctx, auth.ReplaceRoleServicesInput{
		AccessToken: loginResult.AccessToken,
		RoleID:      "role_created_01",
		ServiceIDs:  []string{"service_docs", "service_gitlab"},
	}); err != nil {
		t.Fatalf("replace role services: %v", err)
	}
	assertRoleServices(t, ctx, db, "role_created_01", []string{"service_docs", "service_gitlab"})

	overrides, err := service.ReplaceUserServiceOverrides(ctx, auth.ReplaceUserServiceOverridesInput{
		AccessToken:     loginResult.AccessToken,
		UserID:          "user_alice",
		AllowServiceIDs: []string{"service_docs"},
		DenyServiceIDs:  []string{"service_gitlab"},
	})
	if err != nil {
		t.Fatalf("replace user service overrides: %v", err)
	}
	if len(overrides) != 2 {
		t.Fatalf("expected 2 overrides, got %d", len(overrides))
	}
	assertUserServiceOverrides(t, ctx, db, "user_alice", map[string]string{
		"service_docs":   "allow",
		"service_gitlab": "deny",
	})
}
