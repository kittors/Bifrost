package auth_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/kittors/bifrost/apps/gateway/internal/auth"
	"github.com/kittors/bifrost/apps/gateway/internal/database"
)

// 客户端服务和代理授权测试放在一起，因为两者共享同一套访问策略判断。

func TestServiceListClientServicesAppliesRoleAndUserOverrides(t *testing.T) {
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
			return "session_services_01", nil
		},
	}

	insertUserWithRoles(t, ctx, db, "user_bob", "bob", "Bob", "correct horse battery staple", []roleSeed{{id: "role_ops", name: "ops", displayName: "Operations"}})
	insertDevice(t, ctx, db, "device_bob_01", "user_bob", "trusted")
	insertService(t, ctx, db, "service_jenkins", "jenkins", "Jenkins", "operations", "/s/jenkins", "enabled")
	insertService(t, ctx, db, "service_docs", "docs", "Docs", "shared", "/s/docs", "enabled")
	insertRoleService(t, ctx, db, "role_ops", "service_jenkins")
	insertRoleService(t, ctx, db, "role_ops", "service_docs")
	insertUserServiceOverride(t, ctx, db, "user_bob", "service_jenkins", "deny")

	loginResult, err := service.ClientLogin(ctx, auth.ClientLoginInput{
		Username:      "bob",
		Password:      "correct horse battery staple",
		DeviceID:      "device_bob_01",
		ClientVersion: "1.0.0",
	})
	if err != nil {
		t.Fatalf("client login: %v", err)
	}

	services, err := service.ListClientServices(ctx, auth.ListClientServicesInput{
		AccessToken: loginResult.AccessToken,
	})
	if err != nil {
		t.Fatalf("list client services: %v", err)
	}

	if len(services) != 1 {
		t.Fatalf("expected only 1 accessible service, got %d", len(services))
	}

	if services[0].ID != "service_docs" {
		t.Fatalf("expected docs service, got %q", services[0].ID)
	}

	if services[0].AccessSource != "role" {
		t.Fatalf("expected role access source, got %q", services[0].AccessSource)
	}
}

func TestServiceGetClientServiceAndCreateAccessURLRequireAuthorization(t *testing.T) {
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
			return "session_access_01", nil
		},
	}

	insertUserWithRoles(t, ctx, db, "user_alice", "alice", "Alice", "correct horse battery staple", []roleSeed{{id: "role_developer", name: "developer", displayName: "Developer"}})
	insertDevice(t, ctx, db, "device_alice_02", "user_alice", "trusted")
	insertService(t, ctx, db, "service_gitlab", "gitlab", "GitLab", "engineering", "/s/gitlab", "enabled")
	insertService(t, ctx, db, "service_jenkins", "jenkins", "Jenkins", "operations", "/s/jenkins", "enabled")
	insertRoleService(t, ctx, db, "role_developer", "service_gitlab")

	loginResult, err := service.ClientLogin(ctx, auth.ClientLoginInput{
		Username:      "alice",
		Password:      "correct horse battery staple",
		DeviceID:      "device_alice_02",
		ClientVersion: "1.0.0",
	})
	if err != nil {
		t.Fatalf("client login: %v", err)
	}

	detail, err := service.GetClientService(ctx, auth.GetClientServiceInput{
		AccessToken: loginResult.AccessToken,
		ServiceID:   "service_gitlab",
	})
	if err != nil {
		t.Fatalf("get client service: %v", err)
	}

	if detail.Key != "gitlab" {
		t.Fatalf("expected gitlab service key, got %q", detail.Key)
	}

	if detail.AccessSource != "role" {
		t.Fatalf("expected role access source, got %q", detail.AccessSource)
	}

	accessURL, err := service.CreateServiceAccessURL(ctx, auth.CreateServiceAccessURLInput{
		AccessToken: loginResult.AccessToken,
		ServiceID:   "service_gitlab",
	})
	if err != nil {
		t.Fatalf("create access url: %v", err)
	}

	if accessURL.PublicPath != "/s/gitlab" {
		t.Fatalf("expected /s/gitlab public path, got %q", accessURL.PublicPath)
	}

	if accessURL.ExpiresIn != 300 {
		t.Fatalf("expected expiresIn 300, got %d", accessURL.ExpiresIn)
	}

	if accessURL.AccessTicket == "" {
		t.Fatal("expected short-lived service access ticket")
	}

	resolvedFromTicket, err := service.ResolveProxyRequest(ctx, auth.ResolveProxyRequestInput{
		AccessTicket: accessURL.AccessTicket,
		ServiceKey:   "gitlab",
	})
	if err != nil {
		t.Fatalf("resolve proxy request from service access ticket: %v", err)
	}

	if resolvedFromTicket.ServiceID != "service_gitlab" {
		t.Fatalf("expected service_gitlab from ticket, got %q", resolvedFromTicket.ServiceID)
	}

	if _, err := service.ResolveProxyRequest(ctx, auth.ResolveProxyRequestInput{
		AccessTicket: accessURL.AccessTicket,
		ServiceKey:   "jenkins",
	}); err == nil {
		t.Fatal("expected service access ticket to be rejected for another service")
	}

	if _, err := service.GetClientService(ctx, auth.GetClientServiceInput{
		AccessToken: loginResult.AccessToken,
		ServiceID:   "service_jenkins",
	}); err == nil {
		t.Fatal("expected unauthorized service detail lookup to fail")
	}
}

func TestServiceResolveProxyRequestByServiceKeyEnforcesPolicy(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dsn := createTestDatabase(t, ctx)
	if err := database.MigrateUp(ctx, dsn); err != nil {
		t.Fatalf("migrate up: %v", err)
	}

	db := openDB(t, dsn)
	now := time.Date(2026, time.April, 17, 13, 30, 0, 0, time.UTC)
	sessionCounter := 0
	service := auth.Service{
		DB:              db,
		PasswordHasher:  auth.DefaultPasswordHasher(),
		TokenIssuer:     auth.TokenIssuer{Secret: []byte("0123456789abcdef0123456789abcdef"), TTL: 15 * time.Minute, Now: func() time.Time { return now }},
		Now:             func() time.Time { return now },
		RefreshTokenTTL: 7 * 24 * time.Hour,
		SessionIDFactory: func() (string, error) {
			sessionCounter++
			return fmt.Sprintf("session_proxy_%02d", sessionCounter), nil
		},
	}

	insertUserWithRoles(t, ctx, db, "user_alice", "alice", "Alice", "correct horse battery staple", []roleSeed{{id: "role_developer", name: "developer", displayName: "Developer"}})
	insertUserWithRoles(t, ctx, db, "user_bob", "bob", "Bob", "correct horse battery staple", []roleSeed{{id: "role_ops", name: "ops", displayName: "Operations"}})
	insertService(t, ctx, db, "service_gitlab", "gitlab", "GitLab", "engineering", "/s/gitlab", "enabled")
	insertService(t, ctx, db, "service_jenkins", "jenkins", "Jenkins", "operations", "/s/jenkins", "enabled")
	insertRoleService(t, ctx, db, "role_developer", "service_gitlab")
	insertRoleService(t, ctx, db, "role_ops", "service_jenkins")
	insertUserServiceOverride(t, ctx, db, "user_bob", "service_jenkins", "deny")
	insertDevice(t, ctx, db, "device_alice_01", "user_alice", "trusted")
	insertDevice(t, ctx, db, "device_bob_01", "user_bob", "trusted")

	aliceLogin, err := service.ClientLogin(ctx, auth.ClientLoginInput{
		Username:      "alice",
		Password:      "correct horse battery staple",
		DeviceID:      "device_alice_01",
		ClientVersion: "1.0.0",
	})
	if err != nil {
		t.Fatalf("alice login: %v", err)
	}

	resolved, err := service.ResolveProxyRequest(ctx, auth.ResolveProxyRequestInput{
		AccessToken: aliceLogin.AccessToken,
		ServiceKey:  "gitlab",
	})
	if err != nil {
		t.Fatalf("resolve proxy request: %v", err)
	}

	if resolved.ServiceID != "service_gitlab" {
		t.Fatalf("expected service_gitlab, got %q", resolved.ServiceID)
	}
	if resolved.UpstreamURL != "http://gitlab:8080" {
		t.Fatalf("expected upstream url http://gitlab:8080, got %q", resolved.UpstreamURL)
	}
	if resolved.AccessSource != "role" {
		t.Fatalf("expected role access source, got %q", resolved.AccessSource)
	}

	bobLogin, err := service.ClientLogin(ctx, auth.ClientLoginInput{
		Username:      "bob",
		Password:      "correct horse battery staple",
		DeviceID:      "device_bob_01",
		ClientVersion: "1.0.0",
	})
	if err != nil {
		t.Fatalf("bob login: %v", err)
	}

	if _, err := service.ResolveProxyRequest(ctx, auth.ResolveProxyRequestInput{
		AccessToken: bobLogin.AccessToken,
		ServiceKey:  "jenkins",
	}); err == nil {
		t.Fatal("expected deny override to block proxy request")
	}
}

func TestServiceResolveProxyRequestRejectsDisabledServiceUserAndDevice(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dsn := createTestDatabase(t, ctx)
	if err := database.MigrateUp(ctx, dsn); err != nil {
		t.Fatalf("migrate up: %v", err)
	}

	db := openDB(t, dsn)
	now := time.Date(2026, time.April, 17, 13, 30, 0, 0, time.UTC)
	sessionCounter := 0
	service := auth.Service{
		DB:              db,
		PasswordHasher:  auth.DefaultPasswordHasher(),
		TokenIssuer:     auth.TokenIssuer{Secret: []byte("0123456789abcdef0123456789abcdef"), TTL: 15 * time.Minute, Now: func() time.Time { return now }},
		Now:             func() time.Time { return now },
		RefreshTokenTTL: 7 * 24 * time.Hour,
		SessionIDFactory: func() (string, error) {
			sessionCounter++
			return fmt.Sprintf("session_proxy_status_%02d", sessionCounter), nil
		},
	}

	insertUserWithRoles(t, ctx, db, "user_alice", "alice", "Alice", "correct horse battery staple", []roleSeed{{id: "role_developer", name: "developer", displayName: "Developer"}})
	insertService(t, ctx, db, "service_gitlab", "gitlab", "GitLab", "engineering", "/s/gitlab", "enabled")
	insertRoleService(t, ctx, db, "role_developer", "service_gitlab")
	insertDevice(t, ctx, db, "device_alice_01", "user_alice", "trusted")

	loginResult, err := service.ClientLogin(ctx, auth.ClientLoginInput{
		Username:      "alice",
		Password:      "correct horse battery staple",
		DeviceID:      "device_alice_01",
		ClientVersion: "1.0.0",
	})
	if err != nil {
		t.Fatalf("alice login: %v", err)
	}

	if _, err := db.ExecContext(ctx, `UPDATE services SET status = 'disabled' WHERE id = 'service_gitlab'`); err != nil {
		t.Fatalf("disable service: %v", err)
	}
	if _, err := service.ResolveProxyRequest(ctx, auth.ResolveProxyRequestInput{
		AccessToken: loginResult.AccessToken,
		RequestID:   "req_service_disabled",
		ServiceKey:  "gitlab",
	}); err == nil {
		t.Fatal("expected disabled service to be rejected")
	}
	assertAuditEventCountByRequest(t, ctx, db, "req_service_disabled", 1)

	if _, err := db.ExecContext(ctx, `UPDATE services SET status = 'enabled' WHERE id = 'service_gitlab'`); err != nil {
		t.Fatalf("enable service: %v", err)
	}
	if _, err := db.ExecContext(ctx, `UPDATE users SET status = 'disabled' WHERE id = 'user_alice'`); err != nil {
		t.Fatalf("disable user: %v", err)
	}
	if _, err := service.ResolveProxyRequest(ctx, auth.ResolveProxyRequestInput{
		AccessToken: loginResult.AccessToken,
		RequestID:   "req_user_disabled",
		ServiceKey:  "gitlab",
	}); err == nil {
		t.Fatal("expected disabled user to be rejected")
	}

	if _, err := db.ExecContext(ctx, `UPDATE users SET status = 'enabled' WHERE id = 'user_alice'`); err != nil {
		t.Fatalf("enable user: %v", err)
	}
	if _, err := db.ExecContext(ctx, `UPDATE devices SET status = 'disabled' WHERE id = 'device_alice_01'`); err != nil {
		t.Fatalf("disable device: %v", err)
	}
	if _, err := service.ResolveProxyRequest(ctx, auth.ResolveProxyRequestInput{
		AccessToken: loginResult.AccessToken,
		RequestID:   "req_device_disabled",
		ServiceKey:  "gitlab",
	}); err == nil {
		t.Fatal("expected disabled device to be rejected")
	}
}
