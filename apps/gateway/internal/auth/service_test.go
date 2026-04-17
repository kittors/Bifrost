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
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/kittors/bifrost/apps/gateway/internal/auth"
	"github.com/kittors/bifrost/apps/gateway/internal/database"
)

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

func TestServiceRegisterDeviceCreatesTrustedDeviceForCurrentUser(t *testing.T) {
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
		SessionIDFactory: func() (string, error) { return "session_register_01", nil },
		DeviceIDFactory:  func() (string, error) { return "device_registered_01", nil },
	}

	insertUserWithRoles(t, ctx, db, "user_alice", "alice", "Alice", "correct horse battery staple", []roleSeed{{id: "role_developer", name: "developer", displayName: "Developer"}})

	accessToken := issueAccessTokenForTest(t, service.TokenIssuer, "user_alice", "", "session_register_01")

	publicKey, _, fingerprint := generateEd25519Material(t)

	device, err := service.RegisterDevice(ctx, auth.RegisterDeviceInput{
		AccessToken:          accessToken,
		Name:                 "Alice MacBook Pro",
		OS:                   "macOS",
		ClientVersion:        "1.0.0",
		PublicKey:            publicKey,
		PublicKeyFingerprint: fingerprint,
	})
	if err != nil {
		t.Fatalf("register device: %v", err)
	}

	if device.ID != "device_registered_01" {
		t.Fatalf("expected device id device_registered_01, got %q", device.ID)
	}

	if device.Status != "trusted" {
		t.Fatalf("expected trusted status, got %q", device.Status)
	}

	var storedUserID string
	var storedFingerprint string
	if err := db.QueryRowContext(
		ctx,
		`SELECT user_id, public_key_fingerprint
		FROM devices
		WHERE id = $1`,
		"device_registered_01",
	).Scan(&storedUserID, &storedFingerprint); err != nil {
		t.Fatalf("query registered device: %v", err)
	}

	if storedUserID != "user_alice" {
		t.Fatalf("expected device bound to user_alice, got %q", storedUserID)
	}

	if storedFingerprint != fingerprint {
		t.Fatalf("expected fingerprint %q, got %q", fingerprint, storedFingerprint)
	}
}

func TestServiceCreateAndVerifyDeviceChallenge(t *testing.T) {
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
		DB:                 db,
		PasswordHasher:     auth.DefaultPasswordHasher(),
		TokenIssuer:        auth.TokenIssuer{Secret: []byte("0123456789abcdef0123456789abcdef"), TTL: 15 * time.Minute, Now: func() time.Time { return now }},
		Now:                func() time.Time { return now },
		RefreshTokenTTL:    7 * 24 * time.Hour,
		SessionIDFactory:   func() (string, error) { return "session_challenge_01", nil },
		ChallengeIDFactory: func() (string, error) { return "challenge_01", nil },
		ChallengeTTL:       2 * time.Minute,
	}

	insertUserWithRoles(t, ctx, db, "user_alice", "alice", "Alice", "correct horse battery staple", []roleSeed{{id: "role_developer", name: "developer", displayName: "Developer"}})
	publicKey, privateKey, fingerprint := generateEd25519Material(t)
	insertDeviceWithKey(t, ctx, db, "device_alice_sign", "user_alice", "trusted", publicKey, fingerprint)

	accessToken := issueAccessTokenForTest(t, service.TokenIssuer, "user_alice", "", "session_challenge_01")

	challenge, err := service.CreateDeviceChallenge(ctx, auth.CreateDeviceChallengeInput{
		AccessToken: accessToken,
		DeviceID:    "device_alice_sign",
	})
	if err != nil {
		t.Fatalf("create device challenge: %v", err)
	}

	if challenge.ID != "challenge_01" {
		t.Fatalf("expected challenge id challenge_01, got %q", challenge.ID)
	}

	if challenge.ExpiresIn != 120 {
		t.Fatalf("expected expiresIn 120, got %d", challenge.ExpiresIn)
	}

	rawChallenge, err := base64.RawURLEncoding.DecodeString(challenge.Challenge)
	if err != nil {
		t.Fatalf("decode challenge: %v", err)
	}

	signature := ed25519.Sign(privateKey, rawChallenge)
	verified, err := service.VerifyDeviceChallenge(ctx, auth.VerifyDeviceChallengeInput{
		AccessToken: accessToken,
		ChallengeID: "challenge_01",
		Signature:   base64.RawURLEncoding.EncodeToString(signature),
	})
	if err != nil {
		t.Fatalf("verify device challenge: %v", err)
	}

	if !verified.Verified {
		t.Fatal("expected verified challenge")
	}

	var verifiedAt sql.NullTime
	if err := db.QueryRowContext(
		ctx,
		`SELECT verified_at
		FROM device_challenges
		WHERE id = $1`,
		"challenge_01",
	).Scan(&verifiedAt); err != nil {
		t.Fatalf("query verified challenge: %v", err)
	}

	if !verifiedAt.Valid {
		t.Fatal("expected challenge verified_at to be set")
	}
}

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

	if _, err := service.GetClientService(ctx, auth.GetClientServiceInput{
		AccessToken: loginResult.AccessToken,
		ServiceID:   "service_jenkins",
	}); err == nil {
		t.Fatal("expected unauthorized service detail lookup to fail")
	}
}

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

	databaseName := fmt.Sprintf("bifrost_auth_test_%d", time.Now().UnixNano())
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
