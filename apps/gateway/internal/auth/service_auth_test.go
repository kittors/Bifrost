package auth_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/kittors/bifrost/apps/gateway/internal/auth"
	"github.com/kittors/bifrost/apps/gateway/internal/database"
)

// 认证会话测试只覆盖登录、刷新、退出和当前用户，避免和设备或策略测试混在一起。

func TestServiceAdminLoginCreatesSession(t *testing.T) {
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
		DB:               db,
		PasswordHasher:   auth.DefaultPasswordHasher(),
		TokenIssuer:      auth.TokenIssuer{Secret: []byte("0123456789abcdef0123456789abcdef"), TTL: 15 * time.Minute, Now: func() time.Time { return now }},
		Now:              func() time.Time { return now },
		RefreshTokenTTL:  7 * 24 * time.Hour,
		SessionIDFactory: func() (string, error) { return "session_admin_01", nil },
	}

	insertUserWithRoles(t, ctx, db, "user_admin", "admin", "Administrator", "correct horse battery staple", []roleSeed{{id: "role_admin", name: "admin", displayName: "Administrator"}})

	result, err := service.AdminLogin(ctx, auth.AdminLoginInput{
		Username: "admin",
		Password: "correct horse battery staple",
	})
	if err != nil {
		t.Fatalf("admin login: %v", err)
	}

	if result.AccessToken == "" {
		t.Fatal("expected access token")
	}

	if result.RefreshToken == "" {
		t.Fatal("expected refresh token")
	}

	if result.ExpiresIn != 900 {
		t.Fatalf("expected expiresIn 900, got %d", result.ExpiresIn)
	}

	if result.User.ID != "user_admin" {
		t.Fatalf("expected user id user_admin, got %q", result.User.ID)
	}

	if len(result.User.Roles) != 1 || result.User.Roles[0] != "role_admin" {
		t.Fatalf("expected admin role, got %#v", result.User.Roles)
	}

	var (
		sessionUserID  string
		sessionStatus  string
		refreshHash    string
		sessionExpires time.Time
		deviceID       sql.NullString
	)
	if err := db.QueryRowContext(
		ctx,
		`SELECT user_id, device_id, refresh_token_hash, status, expires_at
		FROM sessions
		WHERE id = $1`,
		"session_admin_01",
	).Scan(&sessionUserID, &deviceID, &refreshHash, &sessionStatus, &sessionExpires); err != nil {
		t.Fatalf("query created session: %v", err)
	}

	if sessionUserID != "user_admin" {
		t.Fatalf("expected session user user_admin, got %q", sessionUserID)
	}

	if deviceID.Valid {
		t.Fatalf("expected admin session without device binding, got %q", deviceID.String)
	}

	if refreshHash != auth.HashRefreshToken(result.RefreshToken) {
		t.Fatal("expected refresh token hash to be stored")
	}

	if sessionStatus != "active" {
		t.Fatalf("expected session status active, got %q", sessionStatus)
	}

	if want := now.Add(7 * 24 * time.Hour); !sessionExpires.Equal(want) {
		t.Fatalf("expected session expiry %s, got %s", want, sessionExpires)
	}
}

func TestServiceClientLoginRequiresTrustedDeviceAndCreatesBoundSession(t *testing.T) {
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
		DB:               db,
		PasswordHasher:   auth.DefaultPasswordHasher(),
		TokenIssuer:      auth.TokenIssuer{Secret: []byte("0123456789abcdef0123456789abcdef"), TTL: 15 * time.Minute, Now: func() time.Time { return now }},
		Now:              func() time.Time { return now },
		RefreshTokenTTL:  7 * 24 * time.Hour,
		SessionIDFactory: func() (string, error) { return "session_client_01", nil },
	}

	insertUserWithRoles(t, ctx, db, "user_alice", "alice", "Alice", "correct horse battery staple", []roleSeed{{id: "role_developer", name: "developer", displayName: "Developer"}})
	insertDevice(t, ctx, db, "device_alice_01", "user_alice", "trusted")

	result, err := service.ClientLogin(ctx, auth.ClientLoginInput{
		Username:      "alice",
		Password:      "correct horse battery staple",
		DeviceID:      "device_alice_01",
		ClientVersion: "1.0.0",
	})
	if err != nil {
		t.Fatalf("client login: %v", err)
	}

	if result.User.ID != "user_alice" {
		t.Fatalf("expected user id user_alice, got %q", result.User.ID)
	}

	var (
		sessionDeviceID string
		refreshHash     string
	)
	if err := db.QueryRowContext(
		ctx,
		`SELECT device_id, refresh_token_hash
		FROM sessions
		WHERE id = $1`,
		"session_client_01",
	).Scan(&sessionDeviceID, &refreshHash); err != nil {
		t.Fatalf("query created session: %v", err)
	}

	if sessionDeviceID != "device_alice_01" {
		t.Fatalf("expected session device device_alice_01, got %q", sessionDeviceID)
	}

	if refreshHash != auth.HashRefreshToken(result.RefreshToken) {
		t.Fatal("expected refresh token hash to be stored")
	}
}

func TestServiceRefreshSessionRotatesRefreshToken(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dsn := createTestDatabase(t, ctx)
	if err := database.MigrateUp(ctx, dsn); err != nil {
		t.Fatalf("migrate up: %v", err)
	}

	db := openDB(t, dsn)
	currentTime := time.Date(2026, time.April, 17, 13, 30, 0, 0, time.UTC)
	service := auth.Service{
		DB:               db,
		PasswordHasher:   auth.DefaultPasswordHasher(),
		TokenIssuer:      auth.TokenIssuer{Secret: []byte("0123456789abcdef0123456789abcdef"), TTL: 15 * time.Minute, Now: func() time.Time { return currentTime }},
		Now:              func() time.Time { return currentTime },
		RefreshTokenTTL:  7 * 24 * time.Hour,
		SessionIDFactory: func() (string, error) { return "session_refresh_01", nil },
	}

	insertUserWithRoles(t, ctx, db, "user_admin", "admin", "Administrator", "correct horse battery staple", []roleSeed{{id: "role_admin", name: "admin", displayName: "Administrator"}})

	loginResult, err := service.AdminLogin(ctx, auth.AdminLoginInput{
		Username: "admin",
		Password: "correct horse battery staple",
	})
	if err != nil {
		t.Fatalf("admin login: %v", err)
	}

	currentTime = currentTime.Add(time.Minute)

	refreshResult, err := service.RefreshSession(ctx, auth.RefreshInput{
		RefreshToken: loginResult.RefreshToken,
	})
	if err != nil {
		t.Fatalf("refresh session: %v", err)
	}

	if refreshResult.RefreshToken == loginResult.RefreshToken {
		t.Fatal("expected refresh token rotation")
	}

	var storedHash string
	if err := db.QueryRowContext(
		ctx,
		`SELECT refresh_token_hash
		FROM sessions
		WHERE id = $1`,
		"session_refresh_01",
	).Scan(&storedHash); err != nil {
		t.Fatalf("query refresh token hash: %v", err)
	}

	if storedHash != auth.HashRefreshToken(refreshResult.RefreshToken) {
		t.Fatal("expected rotated refresh token hash to be stored")
	}

	if storedHash == auth.HashRefreshToken(loginResult.RefreshToken) {
		t.Fatal("expected old refresh token hash to be replaced")
	}
}

func TestServiceLogoutRevokesSession(t *testing.T) {
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
		DB:               db,
		PasswordHasher:   auth.DefaultPasswordHasher(),
		TokenIssuer:      auth.TokenIssuer{Secret: []byte("0123456789abcdef0123456789abcdef"), TTL: 15 * time.Minute, Now: func() time.Time { return now }},
		Now:              func() time.Time { return now },
		RefreshTokenTTL:  7 * 24 * time.Hour,
		SessionIDFactory: func() (string, error) { return "session_logout_01", nil },
	}

	insertUserWithRoles(t, ctx, db, "user_admin", "admin", "Administrator", "correct horse battery staple", []roleSeed{{id: "role_admin", name: "admin", displayName: "Administrator"}})

	loginResult, err := service.AdminLogin(ctx, auth.AdminLoginInput{
		Username: "admin",
		Password: "correct horse battery staple",
	})
	if err != nil {
		t.Fatalf("admin login: %v", err)
	}

	if err := service.Logout(ctx, auth.LogoutInput{
		AccessToken: loginResult.AccessToken,
	}); err != nil {
		t.Fatalf("logout: %v", err)
	}

	var status string
	if err := db.QueryRowContext(
		ctx,
		`SELECT status
		FROM sessions
		WHERE id = $1`,
		"session_logout_01",
	).Scan(&status); err != nil {
		t.Fatalf("query session status: %v", err)
	}

	if status != "revoked" {
		t.Fatalf("expected revoked session, got %q", status)
	}
}

func TestServiceCurrentUserReturnsUserFromAccessToken(t *testing.T) {
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
		DB:               db,
		PasswordHasher:   auth.DefaultPasswordHasher(),
		TokenIssuer:      auth.TokenIssuer{Secret: []byte("0123456789abcdef0123456789abcdef"), TTL: 15 * time.Minute, Now: func() time.Time { return now }},
		Now:              func() time.Time { return now },
		RefreshTokenTTL:  7 * 24 * time.Hour,
		SessionIDFactory: func() (string, error) { return "session_me_01", nil },
	}

	insertUserWithRoles(t, ctx, db, "user_admin", "admin", "Administrator", "correct horse battery staple", []roleSeed{{id: "role_admin", name: "admin", displayName: "Administrator"}})

	loginResult, err := service.AdminLogin(ctx, auth.AdminLoginInput{
		Username: "admin",
		Password: "correct horse battery staple",
	})
	if err != nil {
		t.Fatalf("admin login: %v", err)
	}

	user, err := service.CurrentUser(ctx, auth.CurrentUserInput{
		AccessToken: loginResult.AccessToken,
	})
	if err != nil {
		t.Fatalf("current user: %v", err)
	}

	if user.ID != "user_admin" {
		t.Fatalf("expected user_admin, got %q", user.ID)
	}

	if len(user.Roles) != 1 || user.Roles[0] != "role_admin" {
		t.Fatalf("expected role_admin, got %#v", user.Roles)
	}
}
